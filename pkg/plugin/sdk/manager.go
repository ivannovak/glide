// Package sdk provides the plugin manager for runtime plugin loading
package sdk

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/glide-cli/glide/v3/pkg/branding"
	v1 "github.com/glide-cli/glide/v3/pkg/plugin/sdk/v1"
	"github.com/hashicorp/go-hclog"
	goplugin "github.com/hashicorp/go-plugin"
)

// Cache is a simple plugin cache
type Cache struct {
	mu    sync.RWMutex
	items map[string]*LoadedPlugin
}

// NewCache creates a new cache
func NewCache(timeout time.Duration) *Cache {
	return &Cache{
		items: make(map[string]*LoadedPlugin),
	}
}

// Get retrieves a plugin from cache
func (c *Cache) Get(path string) *LoadedPlugin {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.items[path]
}

// Put adds a plugin to cache
func (c *Cache) Put(path string, plugin *LoadedPlugin) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[path] = plugin
}

// Clear clears the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*LoadedPlugin)
}

// Manager handles plugin discovery, loading, and lifecycle
type Manager struct {
	mu               sync.RWMutex
	plugins          map[string]*LoadedPlugin
	discovered       map[string]*PluginInfo // Discovered but not yet loaded
	discoverer       *Discoverer
	validator        *Validator
	cache            *Cache
	config           *ManagerConfig
	lifecycleManager *LifecycleManager
	resolver         *DependencyResolver
}

// LoadedPlugin represents a loaded and running plugin
type LoadedPlugin struct {
	Name     string
	Path     string
	Client   *goplugin.Client
	Plugin   v1.GlidePluginClient
	Metadata *v1.PluginMetadata
	LastUsed time.Time
	State    *StateTracker // Lifecycle state tracking
}

// ManagerConfig configures the plugin manager
type ManagerConfig struct {
	PluginDirs     []string
	CacheTimeout   time.Duration
	MaxPlugins     int
	EnableDebug    bool
	SecurityStrict bool
}

// DefaultConfig returns default manager configuration
func DefaultConfig() *ManagerConfig {
	// Build plugin directories list
	pluginDirs := []string{
		branding.GetGlobalPluginDir(), // Global plugins
	}

	// Add all ancestor plugin directories up to root or home
	ancestorDirs := findAncestorPluginDirs()
	pluginDirs = append(pluginDirs, ancestorDirs...)

	// Add current directory plugins if not already included
	localPluginDir := branding.GetLocalPluginDir(".")
	if !contains(pluginDirs, localPluginDir) {
		pluginDirs = append(pluginDirs, localPluginDir)
	}

	// Add system-wide plugin directory if it exists
	// Use the branded command name for system directory
	systemPluginDir := fmt.Sprintf("/usr/local/lib/%s/plugins", branding.CommandName)
	if _, err := os.Stat(systemPluginDir); err == nil {
		pluginDirs = append(pluginDirs, systemPluginDir)
	}

	return &ManagerConfig{
		PluginDirs:     pluginDirs,
		CacheTimeout:   5 * time.Minute,
		MaxPlugins:     10,
		EnableDebug:    os.Getenv("GLIDE_PLUGIN_DEBUG") == "1",
		SecurityStrict: true,
	}
}

// NewManager creates a new plugin manager
func NewManager(config *ManagerConfig) *Manager {
	if config == nil {
		config = DefaultConfig()
	}

	validator := NewValidator(config.SecurityStrict)
	// Add all configured plugin directories as trusted paths
	for _, dir := range config.PluginDirs {
		validator.AddTrustedPath(dir)
	}

	// Create lifecycle manager with default config
	lifecycleConfig := DefaultLifecycleConfig()
	lifecycleManager := NewLifecycleManager(lifecycleConfig)

	// Create dependency resolver
	resolver := NewDependencyResolver()

	return &Manager{
		plugins:          make(map[string]*LoadedPlugin),
		discovered:       make(map[string]*PluginInfo),
		discoverer:       NewDiscoverer(config.PluginDirs),
		validator:        validator,
		cache:            NewCache(config.CacheTimeout),
		config:           config,
		lifecycleManager: lifecycleManager,
		resolver:         resolver,
	}
}

// DiscoverPlugins finds all available plugins and loads them
// For lazy loading, use DiscoverPluginsLazy() instead
func (m *Manager) DiscoverPlugins() error {
	return m.DiscoverPluginsWithOptions(false)
}

// DiscoverPluginsLazy discovers plugins without loading them
// Plugins will be loaded on-demand when GetPlugin is called
func (m *Manager) DiscoverPluginsLazy() error {
	return m.DiscoverPluginsWithOptions(true)
}

// DiscoverPluginsWithOptions discovers plugins with configurable loading behavior
func (m *Manager) DiscoverPluginsWithOptions(lazy bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugins, err := m.discoverer.Scan()
	if err != nil {
		return fmt.Errorf("plugin discovery failed: %w", err)
	}

	if lazy {
		// Just store discovered plugins without loading
		for _, p := range plugins {
			if m.config.EnableDebug {
				log.Printf("Discovered plugin (lazy): %s at %s", p.Name, p.Path)
			}

			// Skip if already loaded or discovered
			if _, exists := m.plugins[p.Name]; exists {
				continue
			}
			if _, exists := m.discovered[p.Name]; exists {
				continue
			}

			m.discovered[p.Name] = p
		}
		return nil
	}

	// Sequential plugin loading for non-lazy mode
	return m.loadPluginsSequential(plugins)
}

// loadPluginsSequential loads multiple plugins one at a time.
// Note: Parallel loading was removed due to a data race in hashicorp/go-plugin v1.7.0
// (race between goroutines in Client.Start). Sequential loading is sufficient for
// typical plugin counts (1-5 plugins) and avoids the race condition.
func (m *Manager) loadPluginsSequential(plugins []*PluginInfo) error {
	for _, p := range plugins {
		// Skip if already loaded
		if _, exists := m.plugins[p.Name]; exists {
			continue
		}

		if m.config.EnableDebug {
			log.Printf("Loading plugin: %s at %s", p.Name, p.Path)
		}

		if err := m.loadPluginUnlocked(p); err != nil {
			log.Printf("Failed to load plugin %s: %v", p.Name, err)
			// Continue loading other plugins even if one fails
		}
	}

	return nil
}

// loadPluginUnlocked loads a plugin without holding the lock (for parallel loading)
// Note: Caller must hold m.mu.Lock()
func (m *Manager) loadPluginUnlocked(info *PluginInfo) error {
	// Validate plugin
	if err := m.validator.Validate(info.Path); err != nil {
		return fmt.Errorf("plugin validation failed: %w", err)
	}

	// Check cache
	if cached := m.cache.Get(info.Path); cached != nil {
		m.plugins[info.Name] = cached
		return nil
	}

	// Configure plugin logger based on environment
	var logger hclog.Logger
	switch {
	case os.Getenv("GLIDE_PLUGIN_DEBUG") == "true" || os.Getenv("PLUGIN_DEBUG") == "true":
		logger = hclog.New(&hclog.LoggerOptions{
			Name:   "plugin",
			Level:  hclog.Debug,
			Output: os.Stderr,
		})
	case os.Getenv("GLIDE_PLUGIN_TRACE") == "true" || os.Getenv("PLUGIN_TRACE") == "true":
		logger = hclog.New(&hclog.LoggerOptions{
			Name:   "plugin",
			Level:  hclog.Trace,
			Output: os.Stderr,
		})
	default:
		logger = hclog.NewNullLogger()
	}

	// Create plugin client
	client := goplugin.NewClient(&goplugin.ClientConfig{
		HandshakeConfig:  v1.HandshakeConfig,
		Plugins:          v1.PluginMap,
		Cmd:              exec.Command(info.Path),
		AllowedProtocols: []goplugin.Protocol{goplugin.ProtocolGRPC},
		Managed:          true,
		Logger:           logger,
	})

	// Connect to plugin
	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return fmt.Errorf("failed to connect to plugin: %w", err)
	}

	// Dispense the plugin
	raw, err := rpcClient.Dispense("glide")
	if err != nil {
		client.Kill()
		return fmt.Errorf("failed to dispense plugin: %w", err)
	}

	glidePlugin, ok := raw.(v1.GlidePluginClient)
	if !ok {
		client.Kill()
		return fmt.Errorf("plugin does not implement GlidePlugin interface")
	}

	// Get metadata
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	metadata, err := glidePlugin.GetMetadata(ctx, &v1.Empty{})
	if err != nil {
		client.Kill()
		return fmt.Errorf("failed to get plugin metadata: %w", err)
	}

	// Create loaded plugin with state tracker
	loaded := &LoadedPlugin{
		Name:     metadata.Name,
		Path:     info.Path,
		Client:   client,
		Plugin:   glidePlugin,
		Metadata: metadata,
		LastUsed: time.Now(),
		State:    NewStateTracker(metadata.Name),
	}

	// Store in manager and cache
	m.plugins[metadata.Name] = loaded
	m.cache.Put(info.Path, loaded)

	// Register with lifecycle manager
	adapter := newLifecycleAdapter(loaded)
	if err := m.lifecycleManager.Register(metadata.Name, adapter); err != nil {
		client.Kill()
		delete(m.plugins, metadata.Name)
		return fmt.Errorf("failed to register plugin with lifecycle manager: %w", err)
	}

	// Initialize and start the plugin through lifecycle
	lifecycleCtx := context.Background()
	if err := m.lifecycleManager.InitPlugin(lifecycleCtx, metadata.Name); err != nil {
		client.Kill()
		delete(m.plugins, metadata.Name)
		_ = m.lifecycleManager.Unregister(metadata.Name)
		return fmt.Errorf("failed to initialize plugin: %w", err)
	}

	if err := m.lifecycleManager.StartPlugin(lifecycleCtx, metadata.Name); err != nil {
		client.Kill()
		delete(m.plugins, metadata.Name)
		_ = m.lifecycleManager.Unregister(metadata.Name)
		return fmt.Errorf("failed to start plugin: %w", err)
	}

	if m.config.EnableDebug {
		log.Printf("Loaded plugin: %s v%s", metadata.Name, metadata.Version)
	}

	return nil
}

// LoadPlugin loads a specific plugin by path
func (m *Manager) LoadPlugin(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	info := &PluginInfo{
		Name: filepath.Base(path),
		Path: path,
	}

	return m.loadPluginUnlocked(info)
}

// GetPlugin returns a loaded plugin by name
// If the plugin was discovered but not loaded (lazy loading), it will be loaded on-demand
func (m *Manager) GetPlugin(name string) (*LoadedPlugin, error) {
	// First check if already loaded (read lock)
	m.mu.RLock()
	plugin, exists := m.plugins[name]
	if exists {
		m.mu.RUnlock()
		// Update last used time
		plugin.LastUsed = time.Now()

		// Check if client is still alive
		if plugin.Client.Exited() {
			return nil, fmt.Errorf("plugin %s has exited", name)
		}

		return plugin, nil
	}

	// Check if discovered but not loaded
	info, discovered := m.discovered[name]
	m.mu.RUnlock()

	if !discovered {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	// Load the plugin on-demand (needs write lock)
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if plugin, exists := m.plugins[name]; exists {
		plugin.LastUsed = time.Now()
		return plugin, nil
	}

	// Load the plugin
	if err := m.loadPluginUnlocked(info); err != nil {
		return nil, fmt.Errorf("failed to load plugin %s: %w", name, err)
	}

	// Remove from discovered since it's now loaded
	delete(m.discovered, name)

	plugin = m.plugins[name]
	if plugin == nil {
		return nil, fmt.Errorf("plugin %s failed to load", name)
	}

	return plugin, nil
}

// ExecuteCommand runs a plugin command
func (m *Manager) ExecuteCommand(pluginName, command string, args []string) error {
	plugin, err := m.GetPlugin(pluginName)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Check if command is interactive
	commands, err := plugin.Plugin.ListCommands(ctx, &v1.Empty{})
	if err != nil {
		return fmt.Errorf("failed to list commands: %w", err)
	}

	var cmdInfo *v1.CommandInfo
	for _, cmd := range commands.Commands {
		if cmd.Name == command {
			cmdInfo = cmd
			break
		}
	}

	if cmdInfo == nil {
		return fmt.Errorf("command %s not found in plugin %s", command, pluginName)
	}

	// Execute command
	if cmdInfo.Interactive {
		// Handle interactive command
		return m.ExecuteInteractive(plugin, command, args)
	} else {
		// Execute non-interactive command
		req := &v1.ExecuteRequest{
			Command: command,
			Args:    args,
		}

		resp, err := plugin.Plugin.ExecuteCommand(ctx, req)
		if err != nil {
			return fmt.Errorf("command execution failed: %w", err)
		}

		if !resp.Success {
			return fmt.Errorf("command failed: %s", resp.Error)
		}

		// Output results
		if len(resp.Stdout) > 0 {
			fmt.Print(string(resp.Stdout))
		}
		if len(resp.Stderr) > 0 {
			fmt.Fprint(os.Stderr, string(resp.Stderr))
		}
	}

	return nil
}

// ExecuteInteractive handles interactive commands with bidirectional streaming
func (m *Manager) ExecuteInteractive(plugin *LoadedPlugin, command string, args []string) error {
	// Create context for the interactive session
	ctx := context.Background()

	// Start the interactive stream with the plugin
	stream, err := plugin.Plugin.StartInteractive(ctx)
	if err != nil {
		return fmt.Errorf("failed to start interactive session: %w", err)
	}

	// Send the command name as the first message so the plugin knows which command to execute
	// Use a special message type or format to indicate this is the command routing message
	if err := stream.Send(&v1.StreamMessage{
		Type: v1.StreamMessage_STDIN,
		Data: []byte(command),
	}); err != nil {
		return fmt.Errorf("failed to send command name: %w", err)
	}

	// Create channels for communication
	errCh := make(chan error, 3)

	// Handle stdin forwarding to the plugin
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				if err != io.EOF {
					errCh <- fmt.Errorf("stdin read error: %w", err)
				}
				return
			}

			if err := stream.Send(&v1.StreamMessage{
				Type: v1.StreamMessage_STDIN,
				Data: buf[:n],
			}); err != nil {
				errCh <- fmt.Errorf("failed to send stdin: %w", err)
				return
			}
		}
	}()

	// Handle output from the plugin
	go func() {
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				errCh <- nil
				return
			}
			if err != nil {
				errCh <- fmt.Errorf("stream recv error: %w", err)
				return
			}

			switch msg.Type {
			case v1.StreamMessage_STDOUT:
				os.Stdout.Write(msg.Data)
			case v1.StreamMessage_STDERR:
				os.Stderr.Write(msg.Data)
			case v1.StreamMessage_EXIT:
				if msg.ExitCode != 0 {
					errCh <- fmt.Errorf("command exited with code %d", msg.ExitCode)
				} else {
					errCh <- nil
				}
				return
			case v1.StreamMessage_ERROR:
				errCh <- fmt.Errorf("plugin error: %s", msg.Error)
				return
			}
		}
	}()

	// Handle signals (Ctrl+C, etc.)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	go func() {
		for sig := range sigCh {
			var signalStr string
			switch sig {
			case syscall.SIGINT:
				signalStr = "SIGINT"
			case syscall.SIGTERM:
				signalStr = "SIGTERM"
			default:
				continue
			}

			if err := stream.Send(&v1.StreamMessage{
				Type:   v1.StreamMessage_SIGNAL,
				Signal: signalStr,
			}); err != nil {
				// Log error but don't fail the session
				continue
			}

			if sig == syscall.SIGTERM {
				errCh <- nil
				return
			}
		}
	}()

	// Wait for completion
	return <-errCh
}

// ListPlugins returns all loaded plugins
func (m *Manager) ListPlugins() []*LoadedPlugin {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var plugins []*LoadedPlugin
	for _, p := range m.plugins {
		plugins = append(plugins, p)
	}
	return plugins
}

// ListDiscoveredPlugins returns all discovered plugin names (both loaded and unloaded)
func (m *Manager) ListDiscoveredPlugins() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.plugins)+len(m.discovered))

	// Add loaded plugins
	for name := range m.plugins {
		names = append(names, name)
	}

	// Add discovered but unloaded plugins
	for name := range m.discovered {
		names = append(names, name)
	}

	return names
}

// IsPluginLoaded returns true if the plugin is currently loaded
func (m *Manager) IsPluginLoaded(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.plugins[name]
	return exists
}

// IsPluginDiscovered returns true if the plugin has been discovered (loaded or not)
func (m *Manager) IsPluginDiscovered(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, exists := m.plugins[name]; exists {
		return true
	}
	if _, exists := m.discovered[name]; exists {
		return true
	}
	return false
}

// Cleanup shuts down all plugins
// Cleanup gracefully shuts down all plugins
func (m *Manager) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Use lifecycle manager for graceful shutdown
	ctx := context.Background()
	if err := m.lifecycleManager.StopAll(ctx); err != nil {
		if m.config.EnableDebug {
			log.Printf("Error during graceful shutdown: %v", err)
		}
	}

	// Unregister all plugins from lifecycle manager
	for name := range m.plugins {
		_ = m.lifecycleManager.Unregister(name)
	}

	m.plugins = make(map[string]*LoadedPlugin)
	m.cache.Clear()
}

// Discoverer finds plugins in configured directories
type Discoverer struct {
	dirs []string
}

// PluginInfo contains basic plugin information
type PluginInfo struct {
	Name string
	Path string
}

// NewDiscoverer creates a plugin discoverer
func NewDiscoverer(dirs []string) *Discoverer {
	return &Discoverer{dirs: dirs}
}

// scanResult holds results from scanning a single directory
type scanResult struct {
	plugins []*PluginInfo
	err     error
}

// Scan searches for plugins in configured directories
func (d *Discoverer) Scan() ([]*PluginInfo, error) {
	// Parallel scan: launch goroutines for each directory
	results := make(chan scanResult, len(d.dirs))

	for _, dir := range d.dirs {
		go func(dir string) {
			plugins, err := d.scanDirectory(dir)
			results <- scanResult{plugins: plugins, err: err}
		}(dir)
	}

	// Collect results
	var allPlugins []*PluginInfo
	seen := make(map[string]bool)

	for i := 0; i < len(d.dirs); i++ {
		result := <-results
		if result.err != nil {
			continue // Skip directories with errors
		}

		for _, p := range result.plugins {
			// Skip if we've already seen this plugin (first found takes precedence)
			if seen[p.Name] {
				continue
			}
			seen[p.Name] = true
			allPlugins = append(allPlugins, p)
		}
	}

	return allPlugins, nil
}

// scanDirectory scans a single directory for plugins
func (d *Discoverer) scanDirectory(dir string) ([]*PluginInfo, error) {
	// Expand home directory
	if strings.HasPrefix(dir, "~") {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, dir[2:])
	}

	// Handle relative paths (like ./.glide/plugins)
	if !filepath.IsAbs(dir) {
		cwd, err := os.Getwd()
		if err == nil {
			dir = filepath.Join(cwd, dir)
		}
	}

	// Check if directory exists (fast path for non-existent)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil
	}

	// Find all files in the plugin directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var plugins []*PluginInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Use DirEntry.Type() to avoid extra stat call when possible
		fileInfo, err := entry.Info()
		if err != nil {
			continue
		}

		// Check if executable
		if fileInfo.Mode()&0111 == 0 {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		plugins = append(plugins, &PluginInfo{
			Name: entry.Name(),
			Path: path,
		})
	}

	return plugins, nil
}

// findAncestorPluginDirs walks up the directory tree looking for plugin directories
func findAncestorPluginDirs() []string {
	var dirs []string

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return dirs
	}

	// Get home directory to stop searching there
	home, _ := os.UserHomeDir()

	// Walk up the directory tree
	current := cwd
	for {
		// Check if we've reached root or home directory
		if current == "/" || current == home || current == filepath.Dir(current) {
			break
		}

		// Check if branded plugin directory exists in this directory
		pluginDir := branding.GetLocalPluginDir(current)
		if info, err := os.Stat(pluginDir); err == nil && info.IsDir() {
			dirs = append(dirs, pluginDir)
		}

		// Move up to parent directory
		current = filepath.Dir(current)
	}

	// Don't reverse - the order is already correct with most specific (deepest) first
	// as we walk up from the current directory

	return dirs
}

// contains checks if a string slice contains a value
func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// convertToPluginMetadata converts v1.PluginMetadata to PluginMetadata
// for dependency resolution
func convertToPluginMetadata(v1Meta *v1.PluginMetadata) PluginMetadata {
	deps := make([]PluginDependency, len(v1Meta.Dependencies))
	for i, d := range v1Meta.Dependencies {
		deps[i] = PluginDependency{
			Name:     d.Name,
			Version:  d.Version,
			Optional: d.Optional,
		}
	}

	return PluginMetadata{
		Name:         v1Meta.Name,
		Version:      v1Meta.Version,
		Author:       v1Meta.Author,
		Description:  v1Meta.Description,
		Dependencies: deps,
	}
}

// ResolveLoadOrder resolves the correct plugin load order based on dependencies.
//
// This method can be used before calling DiscoverPlugins() to determine the
// optimal load order when multiple plugins have dependencies on each other.
//
// Returns:
//   - A slice of plugin names in dependency order (dependencies before dependents)
//   - An error if there are circular dependencies, missing required dependencies,
//     or version mismatches
//
// Example:
//
//	loadOrder, err := manager.ResolveLoadOrder()
//	if err != nil {
//	    return fmt.Errorf("dependency resolution failed: %w", err)
//	}
//	// Load plugins in the determined order
func (m *Manager) ResolveLoadOrder() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Build plugin metadata map
	pluginMeta := make(map[string]PluginMetadata)
	for name, loaded := range m.plugins {
		pluginMeta[name] = convertToPluginMetadata(loaded.Metadata)
	}

	// Resolve dependencies
	return m.resolver.Resolve(pluginMeta)
}
