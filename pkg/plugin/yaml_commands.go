package plugin

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ivannovak/glide/v3/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// LoadPluginYAMLCommands loads YAML commands from a plugin's commands.yml file
func LoadPluginYAMLCommands(pluginPath string) (map[string]*config.Command, error) {
	// Try multiple locations for commands.yml
	var commandsPath string

	// Check if pluginPath is a file or directory
	info, err := os.Stat(pluginPath)
	if err != nil {
		return nil, nil // No commands file
	}

	if info.IsDir() {
		// If it's a directory, look for commands.yml in it
		commandsPath = filepath.Join(pluginPath, "commands.yml")
	} else {
		// If it's a file (the plugin binary), look for commands.yml alongside it
		dir := filepath.Dir(pluginPath)
		// Try commands.yml with the plugin name prefix
		pluginName := strings.TrimSuffix(filepath.Base(pluginPath), filepath.Ext(pluginPath))
		commandsPath = filepath.Join(dir, pluginName+".commands.yml")

		// If that doesn't exist, try plain commands.yml in the same directory
		if _, err := os.Stat(commandsPath); os.IsNotExist(err) {
			commandsPath = filepath.Join(dir, "commands.yml")
		}
	}

	// Check if commands.yml exists
	if _, err := os.Stat(commandsPath); os.IsNotExist(err) {
		return nil, nil // No YAML commands
	}

	// Read the commands file
	data, err := os.ReadFile(commandsPath)
	if err != nil {
		return nil, err
	}

	// Parse the YAML
	var rawCommands struct {
		Commands config.CommandMap `yaml:"commands"`
	}

	if err := yaml.Unmarshal(data, &rawCommands); err != nil {
		return nil, err
	}

	// Parse and return the commands
	return config.ParseCommands(rawCommands.Commands)
}

// AddPluginYAMLCommands adds YAML commands from plugins to the root command
func AddPluginYAMLCommands(rootCmd *cobra.Command, registry interface {
	AddYAMLCommand(string, *config.Command) error
}) error {
	// Get plugin directories
	pluginDirs := getPluginDirectories()

	for _, dir := range pluginDirs {
		// Skip if directory doesn't exist
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		// Read all entries in the plugin directory
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			// Skip directories
			if entry.IsDir() {
				continue
			}

			pluginPath := filepath.Join(dir, entry.Name())

			// Load YAML commands for this plugin
			commands, err := LoadPluginYAMLCommands(pluginPath)
			if err != nil || commands == nil {
				continue // Skip if no commands or error
			}

			// Add commands to registry
			for name, cmd := range commands {
				// Plugin commands have lower priority than local commands
				// Safe to ignore: Plugin YAML command registration errors are logged by registry
				// Conflicts are expected and handled by priority system
				// so they should already be overridden if there's a conflict
				_ = registry.AddYAMLCommand(name, cmd)
			}
		}
	}

	return nil
}

// getPluginDirectories returns the list of plugin directories to search
func getPluginDirectories() []string {
	var dirs []string

	// Global plugin directory
	home, _ := os.UserHomeDir()
	if home != "" {
		dirs = append(dirs, filepath.Join(home, ".glide", "plugins"))
	}

	// Current directory plugins
	if cwd, err := os.Getwd(); err == nil {
		dirs = append(dirs, filepath.Join(cwd, ".glide", "plugins"))

		// Walk up to find parent plugin directories
		current := cwd
		for {
			parent := filepath.Dir(current)
			if parent == current || parent == "/" || parent == home {
				break
			}

			pluginDir := filepath.Join(parent, ".glide", "plugins")
			if info, err := os.Stat(pluginDir); err == nil && info.IsDir() {
				dirs = append(dirs, pluginDir)
			}

			// Check for .git to stop at project root
			if _, err := os.Stat(filepath.Join(parent, ".git")); err == nil {
				break
			}

			current = parent
		}
	}

	return dirs
}
