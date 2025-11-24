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

	"github.com/hashicorp/go-hclog"
	plugin "github.com/hashicorp/go-plugin"
	"github.com/ivannovak/glide/v2/pkg/branding"
	v1 "github.com/ivannovak/glide/v2/pkg/plugin/sdk/v1"
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
	mu         sync.RWMutex
	plugins    map[string]*LoadedPlugin
	discoverer *Discoverer
	validator  *Validator
	cache      *Cache
	config     *ManagerConfig
}

// LoadedPlugin represents a loaded and running plugin
type LoadedPlugin struct {
	Name     string
	Path     string
	Client   *plugin.Client
	Plugin   v1.GlidePluginClient
	Metadata *v1.PluginMetadata
	LastUsed time.Time
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

	return &Manager{
		plugins:    make(map[string]*LoadedPlugin),
		discoverer: NewDiscoverer(config.PluginDirs),
		validator:  validator,
		cache:      NewCache(config.CacheTimeout),
		config:     config,
	}
}

// DiscoverPlugins finds all available plugins
func (m *Manager) DiscoverPlugins() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugins, err := m.discoverer.Scan()
	if err != nil {
		return fmt.Errorf("plugin discovery failed: %w", err)
	}

	for _, p := range plugins {
		if m.config.EnableDebug {
			log.Printf("Discovered plugin: %s at %s", p.Name, p.Path)
		}

		// Don't reload if already loaded
		if _, exists := m.plugins[p.Name]; exists {
			continue
		}

		// Load the plugin
		if err := m.loadPlugin(p); err != nil {
			log.Printf("Failed to load plugin %s: %v", p.Name, err)
			continue
		}
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

	return m.loadPlugin(info)
}

// loadPlugin internal method to load a plugin
func (m *Manager) loadPlugin(info *PluginInfo) error {
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
		// Full debug output
		logger = hclog.New(&hclog.LoggerOptions{
			Name:   "plugin",
			Level:  hclog.Debug,
			Output: os.Stderr,
		})
	case os.Getenv("GLIDE_PLUGIN_TRACE") == "true" || os.Getenv("PLUGIN_TRACE") == "true":
		// Trace level (most verbose)
		logger = hclog.New(&hclog.LoggerOptions{
			Name:   "plugin",
			Level:  hclog.Trace,
			Output: os.Stderr,
		})
	default:
		// Suppress all plugin debug output by default
		logger = hclog.NewNullLogger()
	}

	// Create plugin client
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  v1.HandshakeConfig,
		Plugins:          v1.PluginMap,
		Cmd:              exec.Command(info.Path),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
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

	// Create loaded plugin
	loaded := &LoadedPlugin{
		Name:     metadata.Name,
		Path:     info.Path,
		Client:   client,
		Plugin:   glidePlugin,
		Metadata: metadata,
		LastUsed: time.Now(),
	}

	// Store in manager and cache
	m.plugins[metadata.Name] = loaded
	m.cache.Put(info.Path, loaded)

	if m.config.EnableDebug {
		log.Printf("Loaded plugin: %s v%s", metadata.Name, metadata.Version)
	}

	return nil
}

// GetPlugin returns a loaded plugin by name
func (m *Manager) GetPlugin(name string) (*LoadedPlugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugin, exists := m.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	// Update last used time
	plugin.LastUsed = time.Now()

	// Check if client is still alive
	if plugin.Client.Exited() {
		return nil, fmt.Errorf("plugin %s has exited", name)
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

// Cleanup shuts down all plugins
func (m *Manager) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, plugin := range m.plugins {
		if m.config.EnableDebug {
			log.Printf("Shutting down plugin: %s", name)
		}
		plugin.Client.Kill()
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

// Scan searches for plugins in configured directories
func (d *Discoverer) Scan() ([]*PluginInfo, error) {
	var plugins []*PluginInfo
	seen := make(map[string]bool) // Track seen plugins to avoid duplicates

	for _, dir := range d.dirs {
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

		// Check if directory exists
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		// Find all files in the plugin directory
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			path := filepath.Join(dir, entry.Name())
			info, err := os.Stat(path)
			if err != nil {
				continue
			}

			// Check if executable
			if info.Mode()&0111 == 0 {
				continue
			}

			// Use the filename as the plugin name
			name := entry.Name()

			// Skip if we've already seen this plugin (project-local takes precedence)
			if seen[name] {
				continue
			}
			seen[name] = true

			plugins = append(plugins, &PluginInfo{
				Name: name,
				Path: path,
			})
		}
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
