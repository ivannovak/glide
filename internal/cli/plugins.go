package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/ivannovak/glide/pkg/plugin/sdk"
	v1 "github.com/ivannovak/glide/pkg/plugin/sdk/v1"
	"github.com/spf13/cobra"
)

// NewPluginsCommand creates the plugins management command
func NewPluginsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugins",
		Short: "Manage Glide runtime plugins",
		Long:  `Manage Glide runtime plugins including listing, installing, and removing plugins.`,
	}

	cmd.AddCommand(
		newPluginListCommand(),
		newPluginInfoCommand(),
		newPluginInstallCommand(),
		newPluginRemoveCommand(),
		newPluginReloadCommand(),
	)

	return cmd
}

// newPluginListCommand lists all available plugins
func newPluginListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all available plugins",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := sdk.NewManager(nil)

			// Discover plugins
			if err := manager.DiscoverPlugins(); err != nil {
				return fmt.Errorf("failed to discover plugins: %w", err)
			}

			// List plugins
			plugins := manager.ListPlugins()
			if len(plugins) == 0 {
				fmt.Println("No plugins found.")
				fmt.Println("\nTo install plugins, place them in:")
				fmt.Println("  ~/.glide/plugins/")
				fmt.Println("  /usr/local/lib/glide/plugins/")
				return nil
			}

			// Display plugins in table format
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "NAME\tVERSION\tDESCRIPTION\tSTATUS")
			_, _ = fmt.Fprintln(w, "----\t-------\t-----------\t------")

			for _, p := range plugins {
				status := "Loaded"
				// Check if client has exited
				if p.Client.Exited() {
					status = "Stopped"
				}

				// Use metadata directly
				metadata := p.Metadata
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					metadata.Name,
					metadata.Version,
					metadata.Description,
					status,
				)
			}
			_ = w.Flush()

			return nil
		},
	}
}

// newPluginInfoCommand shows detailed information about a plugin
func newPluginInfoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "info <plugin-name>",
		Short: "Show detailed information about a plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := sdk.NewManager(nil)

			// Discover plugins
			if err := manager.DiscoverPlugins(); err != nil {
				return fmt.Errorf("failed to discover plugins: %w", err)
			}

			// Get specific plugin
			loadedPlugin, err := manager.GetPlugin(args[0])
			if err != nil {
				return err
			}

			// Use metadata directly
			metadata := loadedPlugin.Metadata

			// Display plugin information
			fmt.Printf("Plugin: %s\n", metadata.Name)
			fmt.Printf("Version: %s\n", metadata.Version)
			fmt.Printf("Author: %s\n", metadata.Author)
			fmt.Printf("Description: %s\n", metadata.Description)
			fmt.Printf("Path: %s\n", loadedPlugin.Path)

			if metadata.Homepage != "" {
				fmt.Printf("Homepage: %s\n", metadata.Homepage)
			}

			if metadata.License != "" {
				fmt.Printf("License: %s\n", metadata.License)
			}

			// List commands - need to get from plugin
			glidePlugin := loadedPlugin.Plugin

			commandList, err := glidePlugin.ListCommands(cmd.Context(), &v1.Empty{})
			if err == nil && commandList != nil && len(commandList.Commands) > 0 {
				fmt.Println("\nCommands:")
				for _, cmd := range commandList.Commands {
					interactive := ""
					if cmd.Interactive {
						interactive = " (interactive)"
					}
					fmt.Printf("  %s - %s%s\n", cmd.Name, cmd.Description, interactive)
				}
			}

			// List capabilities
			capabilities, err := glidePlugin.GetCapabilities(cmd.Context(), &v1.Empty{})
			if err == nil && capabilities != nil {
				fmt.Println("\nCapabilities Required:")
				if capabilities.RequiresDocker {
					fmt.Println("  - Docker")
				}
				if capabilities.RequiresNetwork {
					fmt.Println("  - Network")
				}
				if capabilities.RequiresFilesystem {
					fmt.Println("  - Filesystem")
				}
				if capabilities.RequiresInteractive {
					fmt.Println("  - Interactive/TTY")
				}

				if len(capabilities.RequiredCommands) > 0 {
					fmt.Println("\nRequired Commands:")
					for _, cmd := range capabilities.RequiredCommands {
						fmt.Printf("  - %s\n", cmd)
					}
				}

				if len(capabilities.RequiredEnvVars) > 0 {
					fmt.Println("\nRequired Environment Variables:")
					for _, env := range capabilities.RequiredEnvVars {
						fmt.Printf("  - %s\n", env)
					}
				}
			}

			return nil
		},
	}
}

// newPluginInstallCommand installs a new plugin
func newPluginInstallCommand() *cobra.Command {
	var source string

	cmd := &cobra.Command{
		Use:   "install <plugin-path>",
		Short: "Install a plugin from a local file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pluginPath := args[0]

			// Verify plugin exists
			if _, err := os.Stat(pluginPath); err != nil {
				return fmt.Errorf("plugin file not found: %w", err)
			}

			// Get plugin name from path
			pluginName := filepath.Base(pluginPath)
			if !strings.HasPrefix(pluginName, "glide-plugin-") {
				return fmt.Errorf("plugin name must start with 'glide-plugin-'")
			}

			// Determine installation directory
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}

			installDir := filepath.Join(home, ".glide", "plugins")
			if err := os.MkdirAll(installDir, 0755); err != nil {
				return fmt.Errorf("failed to create plugins directory: %w", err)
			}

			// Copy plugin to installation directory
			destPath := filepath.Join(installDir, pluginName)

			// Copy file
			src, err := os.Open(pluginPath)
			if err != nil {
				return fmt.Errorf("failed to open source file: %w", err)
			}
			defer src.Close()

			dst, err := os.Create(destPath)
			if err != nil {
				return fmt.Errorf("failed to create destination file: %w", err)
			}
			defer dst.Close()

			if _, err := io.Copy(dst, src); err != nil {
				return fmt.Errorf("failed to copy plugin: %w", err)
			}

			// Make plugin executable
			if err := os.Chmod(destPath, 0755); err != nil {
				return fmt.Errorf("failed to make plugin executable: %w", err)
			}

			// Load and validate plugin
			manager := sdk.NewManager(nil)
			if err := manager.LoadPlugin(destPath); err != nil {
				// Remove plugin if validation fails
				os.Remove(destPath)
				return fmt.Errorf("plugin validation failed: %w", err)
			}

			fmt.Printf("Plugin '%s' installed successfully to %s\n", pluginName, destPath)
			fmt.Println("Run 'glide plugins list' to see all available plugins")

			return nil
		},
	}

	cmd.Flags().StringVar(&source, "source", "", "Source URL or path for the plugin")

	return cmd
}

// newPluginRemoveCommand removes an installed plugin
func newPluginRemoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <plugin-name>",
		Short: "Remove an installed plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pluginName := args[0]

			// Find plugin path
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}

			// Check multiple locations
			locations := []string{
				filepath.Join(home, ".glide", "plugins", "glide-plugin-"+pluginName),
				filepath.Join(home, ".glide", "plugins", pluginName),
			}

			var pluginPath string
			for _, path := range locations {
				if _, err := os.Stat(path); err == nil {
					pluginPath = path
					break
				}
			}

			if pluginPath == "" {
				return fmt.Errorf("plugin '%s' not found", pluginName)
			}

			// Remove plugin
			if err := os.Remove(pluginPath); err != nil {
				return fmt.Errorf("failed to remove plugin: %w", err)
			}

			fmt.Printf("Plugin '%s' removed successfully\n", pluginName)

			return nil
		},
	}
}

// newPluginReloadCommand reloads all plugins
func newPluginReloadCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "reload",
		Short: "Reload all plugins",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := sdk.NewManager(nil)

			// Clean up existing plugins
			manager.Cleanup()

			// Rediscover plugins
			if err := manager.DiscoverPlugins(); err != nil {
				return fmt.Errorf("failed to discover plugins: %w", err)
			}

			plugins := manager.ListPlugins()
			fmt.Printf("Reloaded %d plugin(s)\n", len(plugins))

			return nil
		},
	}
}
