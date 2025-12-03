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
	"strings"
	"text/tabwriter"

	"github.com/glide-cli/glide/v3/pkg/branding"
	"github.com/glide-cli/glide/v3/pkg/plugin/sdk"
	v1 "github.com/glide-cli/glide/v3/pkg/plugin/sdk/v1"
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
		newPluginUpdateCommand(),
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
			// Safe to ignore: Table header formatting (informational display only)
			_, _ = fmt.Fprintln(w, "NAME\tVERSION\tDESCRIPTION\tSTATUS")
			_, _ = fmt.Fprintln(w, "----\t-------\t-----------\t------")

			for _, p := range plugins {
				status := "Loaded"
				// Check if client has exited
				if p.Client.Exited() {
					status = "Stopped"
				}

				// Use metadata directly
				// Safe to ignore: Plugin list row formatting (informational display only)
				metadata := p.Metadata
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					metadata.Name,
					metadata.Version,
					metadata.Description,
					status,
				)
			}
			// Safe to ignore: Table flush (informational display, operation continues if fails)
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
  glide plugins install github.com/glide-cli/glide-plugin-go

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

	// Install from temporary file with proper plugin name
	pluginName := filepath.Base(repo) // e.g., "glide-plugin-go"
	return installFromFileWithName(tempFile, pluginName)
}

// installFromFile installs a plugin from a local file
// It derives the plugin name from the file path
func installFromFile(pluginPath string) error {
	// Get plugin name from path (remove platform suffix if present)
	pluginName := filepath.Base(pluginPath)
	// Remove -darwin-arm64, -linux-amd64, etc. suffixes
	for _, osName := range []string{"darwin", "linux", "windows"} {
		for _, arch := range []string{"amd64", "arm64", "386"} {
			suffix := "-" + osName + "-" + arch
			if len(pluginName) > len(suffix) && pluginName[len(pluginName)-len(suffix):] == suffix {
				pluginName = pluginName[:len(pluginName)-len(suffix)]
			}
			suffixExe := suffix + ".exe"
			if len(pluginName) > len(suffixExe) && pluginName[len(pluginName)-len(suffixExe):] == suffixExe {
				pluginName = pluginName[:len(pluginName)-len(suffixExe)]
			}
		}
	}

	return installFromFileWithName(pluginPath, pluginName)
}

// installFromFileWithName installs a plugin from a local file with an explicit name
func installFromFileWithName(pluginPath, pluginName string) error {
	// Verify plugin exists
	if _, err := os.Stat(pluginPath); err != nil {
		return fmt.Errorf("plugin file not found: %w", err)
	}

	// Determine installation directory
	installDir := branding.GetGlobalPluginDir()
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugins directory: %w", err)
	}

	// Copy plugin to installation directory
	destPath := filepath.Join(installDir, pluginName)

	// Check if plugin already exists and remove it first (allows updates/reinstalls)
	if _, err := os.Stat(destPath); err == nil {
		if err := os.Remove(destPath); err != nil {
			return fmt.Errorf("failed to remove existing plugin: %w", err)
		}
	}

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

// newPluginUpdateCommand updates installed plugins
func newPluginUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [plugin-name]",
		Short: "Update installed plugins to the latest version",
		Long: `Update one or all installed plugins to their latest versions from GitHub.

Examples:
  # Update all plugins
  glide plugins update

  # Update a specific plugin
  glide plugins update go`,
		Aliases: []string{"upgrade"},
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := sdk.NewManager(nil)

			// Discover plugins
			if err := manager.DiscoverPlugins(); err != nil {
				return fmt.Errorf("failed to discover plugins: %w", err)
			}

			plugins := manager.ListPlugins()
			if len(plugins) == 0 {
				fmt.Println("No plugins installed.")
				return nil
			}

			// Determine which plugins to update
			var pluginsToUpdate []*sdk.LoadedPlugin
			if len(args) > 0 {
				// Update specific plugin
				pluginName := args[0]
				plugin, err := manager.GetPlugin(pluginName)
				if err != nil {
					return fmt.Errorf("plugin '%s' not found", pluginName)
				}
				pluginsToUpdate = append(pluginsToUpdate, plugin)
			} else {
				// Update all plugins
				pluginsToUpdate = plugins
			}

			// Update each plugin
			updatedCount := 0
			for _, plugin := range pluginsToUpdate {
				metadata := plugin.Metadata

				// Check if plugin has Homepage (GitHub URL)
				if metadata.Homepage == "" {
					fmt.Printf("⚠️  %s: No homepage specified, skipping\n", metadata.Name)
					continue
				}

				// Parse GitHub repo from homepage
				repo := extractGitHubRepo(metadata.Homepage)
				if repo == "" {
					fmt.Printf("⚠️  %s: Homepage is not a GitHub URL, skipping\n", metadata.Name)
					continue
				}

				fmt.Printf("Checking %s...\n", metadata.Name)

				// Get latest release
				release, err := getLatestRelease(repo)
				if err != nil {
					fmt.Printf("❌ %s: Failed to check for updates: %v\n", metadata.Name, err)
					continue
				}

				// Compare versions
				if release.TagName == metadata.Version || release.TagName == "v"+metadata.Version {
					fmt.Printf("✓ %s is already up to date (%s)\n", metadata.Name, metadata.Version)
					continue
				}

				fmt.Printf("📦 %s: %s → %s\n", metadata.Name, metadata.Version, release.TagName)

				// Download and install
				if err := installPluginFromRelease(plugin.Path, repo, release); err != nil {
					fmt.Printf("❌ %s: Update failed: %v\n", metadata.Name, err)
					continue
				}

				fmt.Printf("✅ %s updated successfully\n", metadata.Name)
				updatedCount++
			}

			if updatedCount == 0 {
				fmt.Println("\nNo plugins were updated.")
			} else {
				fmt.Printf("\n✅ Updated %d plugin(s). Run 'glide plugins reload' to reload plugins.\n", updatedCount)
			}

			return nil
		},
	}

	return cmd
}

// extractGitHubRepo extracts owner/repo from a GitHub URL
func extractGitHubRepo(homepage string) string {
	// Remove protocol
	homepage = strings.TrimPrefix(homepage, "https://")
	homepage = strings.TrimPrefix(homepage, "http://")

	// Check if it's a GitHub URL
	if !strings.HasPrefix(homepage, "github.com/") {
		return ""
	}

	// Extract owner/repo (remove github.com/)
	path := homepage[11:] // len("github.com/") = 11

	// Split by / and take first two parts
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return ""
	}

	return parts[0] + "/" + parts[1]
}

// installPluginFromRelease installs a plugin from a GitHub release
func installPluginFromRelease(existingPath, repo string, release *GitHubRelease) error {
	// Determine platform-specific binary name
	pluginName := filepath.Base(repo)
	binaryName := pluginName + "-" + runtime.GOOS + "-" + runtime.GOARCH
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
		return fmt.Errorf("no binary found for %s-%s", runtime.GOOS, runtime.GOARCH)
	}

	// Download to temporary file
	tmpFile, err := downloadFile(downloadURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer os.Remove(tmpFile)

	// Make executable
	if err := os.Chmod(tmpFile, 0755); err != nil {
		return fmt.Errorf("failed to make executable: %w", err)
	}

	// Replace existing plugin
	if err := os.Remove(existingPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove old plugin: %w", err)
	}

	if err := os.Rename(tmpFile, existingPath); err != nil {
		return fmt.Errorf("failed to install plugin: %w", err)
	}

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
