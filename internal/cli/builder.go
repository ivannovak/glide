package cli

import (
	"github.com/ivannovak/glide/pkg/app"
	"github.com/ivannovak/glide/pkg/branding"
	"github.com/spf13/cobra"
)

// Builder handles the construction of CLI commands
type Builder struct {
	app      *app.Application
	registry *Registry
}

// NewBuilder creates a new command builder
func NewBuilder(application *app.Application) *Builder {
	builder := &Builder{
		app:      application,
		registry: NewRegistry(),
	}

	// Register all commands
	builder.registerCommands()

	return builder
}

// registerCommands registers all available commands
func (b *Builder) registerCommands() {
	// Core commands
	b.registry.Register("setup", func() *cobra.Command {
		return NewSetupCommand(b.app.ProjectContext, b.app.Config)
	}, Metadata{
		Name:        "setup",
		Category:    CategorySetup,
		Description: "Initial setup and configuration",
	})

	// Plugin management commands
	b.registry.Register("plugins", func() *cobra.Command {
		return NewPluginsCommand()
	}, Metadata{
		Name:        "plugins",
		Category:    CategoryCore,
		Description: "Manage runtime plugins",
	})

	b.registry.Register("config", func() *cobra.Command {
		return NewConfigCommand(b.app.Config)
	}, Metadata{
		Name:        "config",
		Category:    CategoryDebug,
		Description: "Show configuration",
		Hidden:      true,
	})

	b.registry.Register("completion", func() *cobra.Command {
		return NewCompletionCommand(b.app.ProjectContext, b.app.Config)
	}, Metadata{
		Name:        "completion",
		Category:    CategorySetup,
		Description: "Generate shell completion scripts",
	})

	b.registry.Register("global", func() *cobra.Command {
		return NewGlobalCommand(b.app.ProjectContext, b.app.Config)
	}, Metadata{
		Name:        "global",
		Category:    CategoryGlobal,
		Description: "Global configuration management",
	})

	b.registry.Register("version", func() *cobra.Command {
		return NewVersionCommand(b.app.ProjectContext, b.app.Config)
	}, Metadata{
		Name:        "version",
		Category:    CategoryCore,
		Description: "Display version information",
	})

	b.registry.Register("help", func() *cobra.Command {
		return NewHelpCommand(b.app.ProjectContext, b.app.Config)
	}, Metadata{
		Name:        "help",
		Category:    CategoryHelp,
		Description: "Context-aware help and guidance",
	})

	// Project-specific commands have been moved to glide-plugin-chirocat
	// Docker commands: up, down, status, logs, shell
	// Developer commands: test, artisan, composer, lint
	// These are now provided via the runtime plugin system

	b.registry.Register("self-update", func() *cobra.Command {
		return NewSelfUpdateCommand(b.app.ProjectContext, b.app.Config)
	}, Metadata{
		Name:        "self-update",
		Category:    CategoryCore,
		Description: "Update Glide CLI to the latest version",
	})
}

// Build creates the root command with all subcommands
func (b *Builder) Build() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           branding.CommandName,
		Short:         branding.GetShortDescription(),
		Long:          branding.GetFullDescription(),
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	// Add all registered commands
	for _, cmd := range b.registry.CreateAll() {
		rootCmd.AddCommand(cmd)
	}

	// Add debug commands if context is available
	if b.app.ProjectContext != nil {
		b.addDebugCommands(rootCmd)
	}

	// Register completions
	b.registerCompletions(rootCmd)

	return rootCmd
}

// addDebugCommands adds debug commands to the root command
func (b *Builder) addDebugCommands(rootCmd *cobra.Command) {
	// Context debug command
	rootCmd.AddCommand(&cobra.Command{
		Use:          "context",
		Short:        "Show detected project context (debug)",
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			showContext(cmd, b.app)
		},
	})

	// Shell test command
	rootCmd.AddCommand(&cobra.Command{
		Use:          "shell-test",
		Short:        "Test shell execution (debug)",
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			testShell(cmd, args, b.app)
		},
	})

	// Docker test command
	rootCmd.AddCommand(&cobra.Command{
		Use:          "docker-test",
		Short:        "Test Docker compose resolution (debug)",
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			testDockerResolution(cmd, args, b.app)
		},
	})

	// Container test command
	rootCmd.AddCommand(&cobra.Command{
		Use:          "container-test",
		Short:        "Test Docker container management (debug)",
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			testContainerManagement(cmd, args, b.app)
		},
	})
}

// registerCompletions registers shell completions for the root command
func (b *Builder) registerCompletions(rootCmd *cobra.Command) {
	// Register completions for the root command
	// This is a simplified implementation - completions are handled by individual commands
}

// GetRegistry returns the command registry
func (b *Builder) GetRegistry() *Registry {
	return b.registry
}
