package plugin

import (
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
// Plugins that require configuration should implement this interface.
// The config parameter is a map of configuration keys to values.
//
// Example:
//
//	func (p *MyPlugin) Configure(config map[string]interface{}) error {
//	    if endpoint, ok := config["endpoint"].(string); ok {
//	        p.endpoint = endpoint
//	    }
//	    return p.validateConfig()
//	}
type PluginConfigurable interface {
	// Configure allows plugin-specific configuration
	Configure(config map[string]interface{}) error
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
	Name        string
	Version     string
	Author      string
	Description string
	Aliases     []string // Plugin-level aliases
	Commands    []CommandInfo
	BuildTags   []string // Required build tags
	ConfigKeys  []string // Configuration keys used
}

// CommandInfo describes a plugin command
type CommandInfo struct {
	Name        string
	Category    string
	Description string
	Aliases     []string
}
