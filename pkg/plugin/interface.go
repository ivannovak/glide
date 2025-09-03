package plugin

import (
	"github.com/spf13/cobra"
)

// Plugin defines the interface for Glide extensions
type Plugin interface {
	// Name returns the plugin identifier
	Name() string

	// Version returns the plugin version
	Version() string

	// Register adds plugin commands to the command tree
	Register(root *cobra.Command) error

	// Configure allows plugin-specific configuration
	Configure(config map[string]interface{}) error

	// Metadata returns plugin information
	Metadata() PluginMetadata
}

// PluginMetadata describes a plugin
type PluginMetadata struct {
	Name        string
	Version     string
	Author      string
	Description string
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
