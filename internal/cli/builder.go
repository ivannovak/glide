package cli

import (
	"os"

	"github.com/ivannovak/glide/internal/config"
	"github.com/ivannovak/glide/pkg/app"
	"github.com/ivannovak/glide/pkg/branding"
	"github.com/ivannovak/glide/pkg/plugin"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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

	// YAML commands are loaded later in AddLocalCommands
	// after the working directory is established

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
		Aliases:     []string{"plugin"},
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

	b.registry.Register("project", func() *cobra.Command {
		return NewProjectCommand(b.app.ProjectContext, b.app.Config)
	}, Metadata{
		Name:        "project",
		Category:    CategoryProject,
		Description: "Project-wide commands for multi-worktree mode",
		Aliases:     []string{"p"},
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
		Aliases:     []string{"update", "upgrade"},
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

// loadYAMLCommands discovers and loads YAML-defined commands with proper priority ordering
func (b *Builder) loadYAMLCommands() {
	// 1. Core commands are already registered (highest priority)

	// 2. Discover and load all .glide.yml files up the tree
	cwd, _ := os.Getwd()
	configPaths, err := config.DiscoverConfigs(cwd)
	if err == nil && len(configPaths) > 0 {
		localConfigs, err := config.LoadAndMergeConfigs(configPaths)
		if err == nil && localConfigs.Commands != nil {
			commands, err := config.ParseCommands(localConfigs.Commands)
			if err == nil {
				for name, cmd := range commands {
					// Check for conflicts with core commands
					if !isProtectedCommand(name) {
						_ = b.registry.AddYAMLCommand(name, cmd)
					}
				}
			}
		}
	}

	// 3. Plugin-bundled YAML commands
	// Load YAML commands from plugin directories
	_ = plugin.AddPluginYAMLCommands(nil, b.registry)

	// 4. Load global commands (~/.glide/config.yml) - lowest priority
	globalConfigPath := branding.GetConfigPath()
	if _, err := os.Stat(globalConfigPath); err == nil {
		data, err := os.ReadFile(globalConfigPath)
		if err == nil {
			var globalConfig config.Config
			if err := yaml.Unmarshal(data, &globalConfig); err == nil {
				if globalConfig.Commands != nil {
					commands, err := config.ParseCommands(globalConfig.Commands)
					if err == nil {
						for name, cmd := range commands {
							// Only add if not already defined
							if _, exists := b.registry.Get(name); !exists {
								_ = b.registry.AddYAMLCommand(name, cmd)
							}
						}
					}
				}
			}
		}
	}
}

// isProtectedCommand checks if a command name is protected (core command)
func isProtectedCommand(name string) bool {
	protected := []string{
		"help", "setup", "plugins", "plugin", "self-update",
		"update", "upgrade", "version", "completion", "global",
		"config", "context", "shell-test", "docker-test", "container-test",
	}
	for _, p := range protected {
		if name == p {
			return true
		}
	}
	return false
}
