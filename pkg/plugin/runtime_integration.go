package plugin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ivannovak/glide/pkg/branding"
	"github.com/ivannovak/glide/pkg/plugin/sdk"
	v1 "github.com/ivannovak/glide/pkg/plugin/sdk/v1"
	"github.com/spf13/cobra"
)

// RuntimePluginIntegration handles runtime plugin loading and command execution
type RuntimePluginIntegration struct {
	manager          *sdk.Manager
	customCategories []*v1.CustomCategory
}

// globalPluginCategories stores custom categories from all loaded plugins
var globalPluginCategories []*v1.CustomCategory

// NewRuntimePluginIntegration creates a new runtime plugin integration
func NewRuntimePluginIntegration() *RuntimePluginIntegration {
	return &RuntimePluginIntegration{
		manager:          sdk.NewManager(nil),
		customCategories: make([]*v1.CustomCategory, 0),
	}
}

// LoadRuntimePlugins discovers and loads all runtime plugins
func (r *RuntimePluginIntegration) LoadRuntimePlugins(rootCmd *cobra.Command) error {
	// Discover plugins
	if err := r.manager.DiscoverPlugins(); err != nil {
		// Don't fail if no plugins found
		return nil
	}

	// Get all loaded plugins
	plugins := r.manager.ListPlugins()

	// Add commands from each plugin
	for _, plugin := range plugins {
		if err := r.addPluginCommands(rootCmd, plugin); err != nil {
			// Log error but continue loading other plugins
			fmt.Fprintf(os.Stderr, "Warning: failed to add commands from plugin %s: %v\n", plugin.Name, err)
		}
	}

	return nil
}

// addPluginCommands adds commands from a plugin to the root command
func (r *RuntimePluginIntegration) addPluginCommands(rootCmd *cobra.Command, plugin *sdk.LoadedPlugin) error {
	// Use plugin directly as it's already the correct type
	glidePlugin := plugin.Plugin

	// Get command list from plugin
	ctx := context.Background()
	commandList, err := glidePlugin.ListCommands(ctx, &v1.Empty{})
	if err != nil {
		return fmt.Errorf("failed to get command list: %w", err)
	}

	// Get metadata
	metadata := plugin.Metadata

	// Get custom categories if the plugin defines them
	customCategories, _ := glidePlugin.GetCustomCategories(ctx, &v1.Empty{})
	if customCategories != nil && len(customCategories.Categories) > 0 {
		// Register custom categories with the help system
		r.registerCustomCategories(customCategories.Categories)
	}

	// Check if plugin wants global registration (not namespaced)
	// Default to namespaced (true) if not specified for backward compatibility
	namespaced := true
	if !metadata.GetNamespaced() {
		namespaced = false
	}

	// If plugin requests global registration (not namespaced)
	if !namespaced {
		// Add commands directly to root
		for _, cmd := range commandList.Commands {
			pluginCommand := r.createPluginCommand(plugin, glidePlugin, cmd)
			// Mark as coming from a plugin for help display
			if pluginCommand.Annotations == nil {
				pluginCommand.Annotations = make(map[string]string)
			}
			pluginCommand.Annotations["plugin"] = plugin.Name
			pluginCommand.Annotations["global_plugin"] = "true"

			// Check for conflicts
			conflicted := false
			for _, existing := range rootCmd.Commands() {
				if existing.Name() == pluginCommand.Name() {
					fmt.Fprintf(os.Stderr, "Warning: plugin %s command '%s' conflicts with existing command, skipping\n",
						plugin.Name, pluginCommand.Name())
					conflicted = true
					break
				}
			}

			if conflicted {
				continue
			}

			rootCmd.AddCommand(pluginCommand)
		}
		return nil
	}

	// Default namespaced behavior (existing code)
	// Create a group command for the plugin if it has multiple commands
	if len(commandList.Commands) > 1 {
		// Create group command
		pluginCmd := &cobra.Command{
			Use:   metadata.Name,
			Short: metadata.Description,
			Long:  fmt.Sprintf("%s\n\nVersion: %s\nAuthor: %s", metadata.Description, metadata.Version, metadata.Author),
			Annotations: map[string]string{
				"category": "plugin",
				"plugin":   plugin.Name,
			},
		}

		// Add plugin-level aliases
		if len(metadata.Aliases) > 0 {
			pluginCmd.Aliases = metadata.Aliases
		}

		// Add individual commands to group
		for _, cmd := range commandList.Commands {
			subCmd := r.createPluginCommand(plugin, glidePlugin, cmd)
			pluginCmd.AddCommand(subCmd)
		}

		rootCmd.AddCommand(pluginCmd)
	} else if len(commandList.Commands) == 1 {
		// Single command - check if we need a group for plugin aliases
		cmd := commandList.Commands[0]

		// If the plugin has aliases, create a group command to support them
		if len(metadata.Aliases) > 0 {
			// Create group command with plugin aliases
			pluginCmd := &cobra.Command{
				Use:     metadata.Name,
				Aliases: metadata.Aliases,
				Short:   metadata.Description,
				Long:    fmt.Sprintf("%s\n\nVersion: %s\nAuthor: %s", metadata.Description, metadata.Version, metadata.Author),
				Annotations: map[string]string{
					"category": "plugin",
					"plugin":   plugin.Name,
				},
			}

			// Add the single command to the group
			subCmd := r.createPluginCommand(plugin, glidePlugin, cmd)
			pluginCmd.AddCommand(subCmd)

			rootCmd.AddCommand(pluginCmd)
		} else {
			// No plugin aliases - add command directly to root (but still namespaced)
			pluginCommand := r.createPluginCommand(plugin, glidePlugin, cmd)
			rootCmd.AddCommand(pluginCommand)
		}
	}

	return nil
}

// createPluginCommand creates a cobra command for a plugin command
func (r *RuntimePluginIntegration) createPluginCommand(plugin *sdk.LoadedPlugin, glidePlugin v1.GlidePluginClient, cmdInfo *v1.CommandInfo) *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdInfo.Name,
		Short: cmdInfo.Description,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Check if command is interactive
			if cmdInfo.Interactive {
				// Handle interactive command
				return r.executeInteractiveCommand(ctx, plugin, glidePlugin, cmdInfo.Name, args)
			} else {
				// Execute non-interactive command
				req := &v1.ExecuteRequest{
					Command: cmdInfo.Name,
					Args:    args,
				}

				resp, err := glidePlugin.ExecuteCommand(ctx, req)
				if err != nil {
					return fmt.Errorf("command execution failed: %w", err)
				}

				if resp.RequiresInteractive {
					// Command requested interactive mode
					return r.executeInteractiveCommand(ctx, plugin, glidePlugin, cmdInfo.Name, args)
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

				return nil
			}
		},
	}

	// Add aliases if any
	if len(cmdInfo.Aliases) > 0 {
		cmd.Aliases = cmdInfo.Aliases
	}

	// Mark as hidden if needed
	if cmdInfo.Hidden {
		cmd.Hidden = true
	}

	// Add annotations
	cmd.Annotations = make(map[string]string)

	// Mark as a plugin command
	cmd.Annotations["plugin"] = plugin.Name

	// Add category - default to "plugin" if not specified
	if cmdInfo.Category != "" {
		cmd.Annotations["category"] = cmdInfo.Category
	} else {
		cmd.Annotations["category"] = "plugin"
	}

	// Add visibility annotation - default to "always" if not specified
	if cmdInfo.Visibility != "" {
		cmd.Annotations["visibility"] = cmdInfo.Visibility
	} else {
		cmd.Annotations["visibility"] = v1.VisibilityAlways
	}

	return cmd
}

// executeInteractiveCommand handles interactive command execution
func (r *RuntimePluginIntegration) executeInteractiveCommand(ctx context.Context, plugin *sdk.LoadedPlugin, glidePlugin v1.GlidePluginClient, command string, args []string) error {
	// Use the manager's executeInteractive implementation which handles all the streaming
	return r.manager.ExecuteInteractive(plugin, command, args)
}

// LoadAllRuntimePlugins is the main entry point for loading runtime plugins
func LoadAllRuntimePlugins(rootCmd *cobra.Command) error {
	integration := NewRuntimePluginIntegration()
	return integration.LoadRuntimePlugins(rootCmd)
}

// ExecuteRuntimePlugin executes a specific runtime plugin command
func ExecuteRuntimePlugin(pluginName, commandName string, args []string) error {
	integration := NewRuntimePluginIntegration()

	// Discover plugins
	if err := integration.manager.DiscoverPlugins(); err != nil {
		return fmt.Errorf("failed to discover plugins: %w", err)
	}

	// Execute the command
	return integration.manager.ExecuteCommand(pluginName, commandName, args)
}

// GetRuntimePluginPath returns the path to runtime plugins directory
func GetRuntimePluginPath() string {
	return branding.GetGlobalPluginDir()
}

// IsRuntimePluginInstalled checks if a runtime plugin is installed
func IsRuntimePluginInstalled(name string) bool {
	pluginPath := GetRuntimePluginPath()
	pluginFile := filepath.Join(pluginPath, "glide-plugin-"+name)

	if _, err := os.Stat(pluginFile); err == nil {
		return true
	}

	// Also check without prefix
	pluginFile = filepath.Join(pluginPath, name)
	_, err := os.Stat(pluginFile)
	return err == nil
}

// registerCustomCategories stores custom categories from a plugin
func (r *RuntimePluginIntegration) registerCustomCategories(categories []*v1.CustomCategory) {
	r.customCategories = append(r.customCategories, categories...)
	// Also update global variable
	globalPluginCategories = append(globalPluginCategories, categories...)
}

// GetCustomCategories returns all custom categories defined by plugins
func (r *RuntimePluginIntegration) GetCustomCategories() []*v1.CustomCategory {
	return r.customCategories
}

// GetGlobalPluginCategories returns all custom categories from loaded plugins
func GetGlobalPluginCategories() []*v1.CustomCategory {
	return globalPluginCategories
}
