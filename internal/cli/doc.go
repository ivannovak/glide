// Package cli provides the command-line interface implementation for Glide.
//
// This package contains the Cobra command tree, command handlers, and
// CLI-specific logic. It integrates with the container for dependency
// injection and uses output formatters for consistent output.
//
// # Command Structure
//
// Commands are organized hierarchically:
//
//	glide                    # Root command
//	├── version              # Show version information
//	├── config               # Configuration management
//	│   ├── show             # Display current config
//	│   └── set              # Set config values
//	├── context              # Context information
//	├── plugins              # Plugin management
//	│   ├── list             # List installed plugins
//	│   └── install          # Install a plugin
//	└── [plugin commands]    # Commands from plugins
//
// # Root Command
//
// Build the root command:
//
//	root := cli.NewRootCommand()
//	if err := root.Execute(); err != nil {
//	    os.Exit(1)
//	}
//
// # Command Options
//
// Commands support common options:
//
//	--format    Output format (table, json, yaml)
//	--quiet     Suppress non-essential output
//	--no-color  Disable color output
//	--debug     Enable debug output
//
// # Integration with Container
//
// Commands receive dependencies through the container:
//
//	func runVersion(cmd *cobra.Command, args []string) error {
//	    return container.Run(cmd.Context(), func(
//	        cfg *config.Config,
//	        out *output.Manager,
//	    ) error {
//	        out.Print(version.GetBuildInfo())
//	        return nil
//	    })
//	}
//
// # Plugin Commands
//
// Plugin-provided commands are automatically registered:
//
//	// Plugins register commands during startup
//	for _, plugin := range plugins {
//	    for _, cmd := range plugin.Commands() {
//	        root.AddCommand(cmd)
//	    }
//	}
package cli
