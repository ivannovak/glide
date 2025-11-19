package branding

import (
	"fmt"
	"os"
	"path/filepath"
)

// These variables can be overridden at build time using ldflags
// Example: go build -ldflags "-X github.com/ivannovak/glide/pkg/branding.CommandName=mycli"
var (
	// CommandName is the name of the CLI command (e.g., "glide", "mycli")
	CommandName = "glide"

	// ConfigFileName is the name of the configuration file (e.g., ".glide.yml")
	ConfigFileName = ".glide.yml"

	// ProjectName is the name of the project (e.g., "Glide", "MyProject")
	ProjectName = "Glide"

	// Description is a short description of the CLI tool
	Description = "context-aware development CLI"

	// LongDescription provides more detailed information about the tool
	LongDescription = `A modern, context-aware development CLI that glides through complex workflows.
It provides intelligent project detection, transparent argument passing, and supports
both single-repository and multi-worktree development modes.`

	// CompletionDir is the directory name for shell completions
	CompletionDir = "glide"

	// RepositoryURL is the URL of the source repository (for updates, documentation, etc.)
	RepositoryURL = "https://github.com/ivannovak/glide"
)

// GetConfigPath returns the full path to the configuration file
func GetConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ConfigFileName)
}

// GetShortDescription returns a formatted short description
func GetShortDescription() string {
	return fmt.Sprintf("%s %s", ProjectName, Description)
}

// GetFullDescription returns the full formatted description for the CLI
func GetFullDescription() string {
	return fmt.Sprintf(`%s is a modern development CLI that glides through complex workflows.
It provides intelligent context awareness, transparent argument passing, and supports
both single-repository and multi-worktree development modes.`,
		capitalize(CommandName))
}

// GetPluginDirName returns the base directory name for plugins (e.g., ".glide")
// This is derived from the ConfigFileName by removing the extension
func GetPluginDirName() string {
	// Extract base name from ConfigFileName
	// e.g., ".glide.yml" -> ".glide", ".mycli.yml" -> ".mycli"
	base := ConfigFileName
	ext := filepath.Ext(base)
	if ext != "" && ext != base {
		// Only remove extension if there's a name before it
		base = base[:len(base)-len(ext)]
	}
	return base
}

// GetGlobalPluginDir returns the path to the global plugins directory
func GetGlobalPluginDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, GetPluginDirName(), "plugins")
}

// GetLocalPluginDir returns the path to a local plugins directory
func GetLocalPluginDir(baseDir string) string {
	return filepath.Join(baseDir, GetPluginDirName(), "plugins")
}

// GetCompletionPath returns the path for shell completion files
func GetCompletionPath(shell string) string {
	var dir string
	switch shell {
	case "bash":
		dir = "/usr/local/etc/bash_completion.d"
	case "zsh":
		dir = "/usr/local/share/zsh/site-functions"
	case "fish":
		homeDir, _ := os.UserHomeDir()
		dir = filepath.Join(homeDir, ".config", "fish", "completions")
	default:
		return ""
	}
	return filepath.Join(dir, CompletionDir)
}

// capitalize returns a string with the first letter capitalized
func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]-32) + s[1:]
}
