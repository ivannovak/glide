// Package sdk provides the plugin manager for runtime plugin loading
package sdk

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-plugin"
	v1 "github.com/ivannovak/glide/pkg/plugin/sdk/v1"
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
	home, _ := os.UserHomeDir()
	return &ManagerConfig{
		PluginDirs: []string{
			filepath.Join(home, ".glide", "plugins"),
			"/usr/local/lib/glide/plugins",
			"./plugins",
		},
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

	return &Manager{
		plugins:    make(map[string]*LoadedPlugin),
		discoverer: NewDiscoverer(config.PluginDirs),
		validator:  NewValidator(config.SecurityStrict),
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

	// Create plugin client
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  v1.HandshakeConfig,
		Plugins:          v1.PluginMap,
		Cmd:              exec.Command(info.Path),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Managed:          true,
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
		return m.executeInteractive(plugin, command, args)
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

// executeInteractive handles interactive commands
func (m *Manager) executeInteractive(plugin *LoadedPlugin, command string, args []string) error {
	// This is a simplified example - real implementation would:
	// 1. Create PTY for terminal interaction
	// 2. Set up bidirectional streaming
	// 3. Handle signals and resize events
	// 4. Stream stdin/stdout/stderr

	fmt.Printf("Starting interactive session for %s %s\n", plugin.Name, command)

	// In a real implementation, we would use StartInteractive with streaming
	// For now, return a placeholder message
	return fmt.Errorf("interactive commands not fully implemented in this example")
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

	for _, dir := range d.dirs {
		// Expand home directory
		if strings.HasPrefix(dir, "~") {
			home, _ := os.UserHomeDir()
			dir = filepath.Join(home, dir[2:])
		}

		// Check if directory exists
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		// Find plugin executables
		matches, err := filepath.Glob(filepath.Join(dir, "glide-plugin-*"))
		if err != nil {
			continue
		}

		for _, path := range matches {
			info, err := os.Stat(path)
			if err != nil || info.IsDir() {
				continue
			}

			// Check if executable
			if info.Mode()&0111 == 0 {
				continue
			}

			name := strings.TrimPrefix(filepath.Base(path), "glide-plugin-")
			plugins = append(plugins, &PluginInfo{
				Name: name,
				Path: path,
			})
		}
	}

	return plugins, nil
}
