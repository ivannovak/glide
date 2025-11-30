package plugin

import (
	"github.com/ivannovak/glide/v2/pkg/plugin/sdk"
	"github.com/spf13/cobra"
)

// PluginIdentifier provides plugin identification and metadata.
//
// This interface follows the Interface Segregation Principle by separating
// identification concerns from lifecycle and configuration concerns.
//
// Implementations should return consistent values across multiple calls.
//
// Example:
//
//	type MyPlugin struct {
//	    name    string
//	    version string
//	}
//
//	func (p *MyPlugin) Name() string { return p.name }
//	func (p *MyPlugin) Version() string { return p.version }
//	func (p *MyPlugin) Metadata() PluginMetadata {
//	    return PluginMetadata{Name: p.name, Version: p.version}
//	}
type PluginIdentifier interface {
	// Name returns the plugin identifier (e.g., "docker", "node")
	Name() string

	// Version returns the plugin version (e.g., "1.0.0")
	Version() string

	// Metadata returns comprehensive plugin information
	Metadata() PluginMetadata
}

// PluginRegistrar handles plugin command registration.
//
// This interface allows plugins to add their commands to the CLI command tree.
//
// Example:
//
//	func (p *MyPlugin) Register(root *cobra.Command) error {
//	    cmd := &cobra.Command{
//	        Use: "my-command",
//	        RunE: p.handleCommand,
//	    }
//	    root.AddCommand(cmd)
//	    return nil
//	}
type PluginRegistrar interface {
	// Register adds plugin commands to the command tree
	Register(root *cobra.Command) error
}

// PluginConfigurable handles plugin-specific configuration.
//
// Plugins implement this interface to perform initialization and configuration
// using the pkg/config type-safe configuration system.
//
// Type-safe configuration approach:
//  1. Define a typed configuration struct
//  2. Register it using pkg/config.Register() in init()
//  3. Access typed config in Configure() via pkg/config.Get[T]() or pkg/config.GetValue[T]()
//
// Example:
//
//	type MyPluginConfig struct {
//	    Endpoint string `json:"endpoint" yaml:"endpoint" validate:"required,url"`
//	    Timeout  int    `json:"timeout" yaml:"timeout" validate:"min=1,max=300"`
//	}
//
//	func init() {
//	    config.Register("my-plugin", MyPluginConfig{Timeout: 30})
//	}
//
//	func (p *MyPlugin) Configure() error {
//	    // Get typed config from registry (populated by config loader from YAML)
//	    cfg, err := config.GetValue[MyPluginConfig]("my-plugin")
//	    if err != nil {
//	        return fmt.Errorf("failed to get plugin config: %w", err)
//	    }
//	    p.endpoint = cfg.Endpoint
//	    p.timeout = cfg.Timeout
//	    return nil
//	}
//
// See pkg/config/MIGRATION.md for complete details.
type PluginConfigurable interface {
	// Configure initializes the plugin.
	// Plugins should use pkg/config.Get[T]() to access their typed configuration.
	Configure() error
}

// Plugin defines the complete interface for Glide extensions.
//
// This is a composite interface that combines all plugin sub-interfaces for
// convenience. Most plugins should implement all three sub-interfaces.
//
// For testing or specialized use cases, you can implement only the
// sub-interfaces you need (e.g., just PluginIdentifier for read-only plugins).
//
// Thread Safety: Plugin methods may be called concurrently. Implementations
// must be safe for concurrent access if they maintain internal state.
//
// Deprecated: While not deprecated for use, prefer depending on specific
// sub-interfaces (PluginIdentifier, PluginRegistrar, PluginConfigurable)
// where possible to follow the Interface Segregation Principle.
type Plugin interface {
	PluginIdentifier
	PluginRegistrar
	PluginConfigurable
}

// PluginMetadata describes a plugin
type PluginMetadata struct {
	Name         string
	Version      string
	Author       string
	Description  string
	Aliases      []string // Plugin-level aliases
	Commands     []CommandInfo
	BuildTags    []string               // Required build tags
	ConfigKeys   []string               // Configuration keys used
	Dependencies []sdk.PluginDependency // Plugin dependencies
}

// CommandInfo describes a plugin command
type CommandInfo struct {
	Name        string
	Category    string
	Description string
	Aliases     []string
}
