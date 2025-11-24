package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"text/tabwriter"

	"github.com/ivannovak/glide/v2/pkg/branding"
	"github.com/ivannovak/glide/v2/pkg/plugin/sdk"
	v1 "github.com/ivannovak/glide/v2/pkg/plugin/sdk/v1"
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
				fmt.Printf("  %s\n", branding.GetGlobalPluginDir())
				fmt.Printf("  /usr/local/lib/%s/plugins/\n", branding.CommandName)
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
	cmd := &cobra.Command{
		Use:   "install <plugin-path-or-url>",
		Short: "Install a plugin from a local file or GitHub release",
		Long: `Install a plugin from a local file or GitHub repository.

Examples:
  # Install from GitHub (downloads latest release)
  glide plugins install github.com/ivannovak/glide-plugin-go

  # Install from local file
  glide plugins install ./glide-plugin-go

Supported formats:
  - github.com/owner/repo (downloads latest release binary)
  - /path/to/plugin-binary (installs local file)`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			source := args[0]

			// Check if source is a GitHub URL
			if isGitHubURL(source) {
				return installFromGitHub(cmd.Context(), source)
			}

			// Install from local file
			return installFromFile(source)
		},
	}

	return cmd
}

// isGitHubURL checks if the source looks like a GitHub repository
func isGitHubURL(source string) bool {
	return len(source) > 11 && (source[:11] == "github.com/" || source[:19] == "https://github.com/")
}

// installFromGitHub downloads and installs a plugin from GitHub releases
func installFromGitHub(ctx context.Context, repo string) error {
	// Parse repository (remove https:// prefix if present)
	repo = filepath.Base(filepath.Dir(repo)) + "/" + filepath.Base(repo)
	if repo[:11] == "github.com/" {
		repo = repo[11:]
	}

	fmt.Printf("Installing plugin from github.com/%s...\n", repo)

	// Get latest release
	release, err := getLatestRelease(repo)
	if err != nil {
		return fmt.Errorf("failed to get latest release: %w", err)
	}

	// Determine platform-specific binary name
	binaryName := filepath.Base(repo) + "-" + runtime.GOOS + "-" + runtime.GOARCH
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	// Find matching asset
	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == binaryName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no binary found for %s-%s in release %s", runtime.GOOS, runtime.GOARCH, release.TagName)
	}

	// Download binary
	fmt.Printf("Downloading %s...\n", binaryName)
	tempFile, err := downloadFile(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download plugin: %w", err)
	}
	defer os.Remove(tempFile)

	// Install from temporary file
	return installFromFile(tempFile)
}

// installFromFile installs a plugin from a local file
func installFromFile(pluginPath string) error {
	// Verify plugin exists
	if _, err := os.Stat(pluginPath); err != nil {
		return fmt.Errorf("plugin file not found: %w", err)
	}

	// Get plugin name from path (remove platform suffix if present)
	pluginName := filepath.Base(pluginPath)
	// Remove -darwin-arm64, -linux-amd64, etc. suffixes
	for _, os := range []string{"darwin", "linux", "windows"} {
		for _, arch := range []string{"amd64", "arm64", "386"} {
			suffix := "-" + os + "-" + arch
			if len(pluginName) > len(suffix) && pluginName[len(pluginName)-len(suffix):] == suffix {
				pluginName = pluginName[:len(pluginName)-len(suffix)]
			}
			suffixExe := suffix + ".exe"
			if len(pluginName) > len(suffixExe) && pluginName[len(pluginName)-len(suffixExe):] == suffixExe {
				pluginName = pluginName[:len(pluginName)-len(suffixExe)]
			}
		}
	}

	// Determine installation directory
	installDir := branding.GetGlobalPluginDir()
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
}

// newPluginRemoveCommand removes an installed plugin
func newPluginRemoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <plugin-name>",
		Short: "Remove an installed plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pluginName := args[0]

			// Check if plugin exists
			pluginDir := branding.GetGlobalPluginDir()
			pluginPath := filepath.Join(pluginDir, pluginName)

			if _, err := os.Stat(pluginPath); err != nil {
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

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// getLatestRelease fetches the latest release from GitHub
func getLatestRelease(repo string) (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Set User-Agent header (required by GitHub API)
	req.Header.Set("User-Agent", "glide-cli")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// downloadFile downloads a file from a URL to a temporary file
func downloadFile(url string) (string, error) {
	// Validate URL to ensure it's from GitHub (security: G107)
	if !isValidGitHubDownloadURL(url) {
		return "", fmt.Errorf("invalid download URL: must be from github.com")
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "glide-plugin-*")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	// Download file #nosec G107 - URL is validated to be from github.com
	resp, err := http.Get(url)
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Copy to temp file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	return tmpFile.Name(), nil
}

// isValidGitHubDownloadURL validates that a URL is from GitHub releases
func isValidGitHubDownloadURL(url string) bool {
	// Must start with https://github.com/ or https://api.github.com/
	return len(url) > 19 && (url[:19] == "https://github.com/" || url[:23] == "https://api.github.com/")
}
