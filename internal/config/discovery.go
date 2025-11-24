package config

import (
	"os"
	"path/filepath"

	"github.com/ivannovak/glide/v2/pkg/branding"
	"gopkg.in/yaml.v3"
)

// DiscoverConfigs finds all configuration files up the directory tree
func DiscoverConfigs(startDir string) ([]string, error) {
	var configs []string

	// Get home directory to stop searching there
	home, _ := os.UserHomeDir()

	// Walk up the directory tree
	current := startDir
	for {
		// Check if we've reached root or home directory
		if current == "/" || current == home || current == filepath.Dir(current) {
			break
		}

		// Check for configuration file in this directory
		// Use the branded config filename from branding package
		configPath := filepath.Join(current, branding.ConfigFileName)
		if _, err := os.Stat(configPath); err == nil {
			configs = append(configs, configPath)
		}

		// Check if we've reached project root (has .git)
		gitPath := filepath.Join(current, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			// Add this config if it exists and isn't already added
			configPath := filepath.Join(current, branding.ConfigFileName)
			if _, err := os.Stat(configPath); err == nil {
				// Check if not already added (might be same as current)
				if len(configs) == 0 || configs[len(configs)-1] != configPath {
					configs = append(configs, configPath)
				}
			}
			// Stop here - we've found project root
			break
		}

		// Move up to parent directory
		current = filepath.Dir(current)
	}

	// Reverse the order so deepest configs come first (highest priority)
	for i, j := 0, len(configs)-1; i < j; i, j = i+1, j-1 {
		configs[i], configs[j] = configs[j], configs[i]
	}

	return configs, nil
}

// LoadAndMergeConfigs loads multiple config files and merges them
func LoadAndMergeConfigs(configPaths []string) (*Config, error) {
	merged := &Config{
		Commands: make(CommandMap),
		Projects: make(map[string]ProjectConfig),
		Plugins:  make(map[string]interface{}),
	}

	// Load configs in reverse order (lowest priority first)
	// so that higher priority configs override
	for i := len(configPaths) - 1; i >= 0; i-- {
		data, err := os.ReadFile(configPaths[i])
		if err != nil {
			continue // Skip configs that can't be read
		}

		var cfg Config
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			continue // Skip invalid configs
		}

		// Merge commands (later configs override earlier ones)
		if cfg.Commands != nil {
			for name, cmd := range cfg.Commands {
				merged.Commands[name] = cmd
			}
		}

		// Merge projects
		if cfg.Projects != nil {
			for name, proj := range cfg.Projects {
				merged.Projects[name] = proj
			}
		}

		// Merge plugins
		if cfg.Plugins != nil {
			for name, plugin := range cfg.Plugins {
				merged.Plugins[name] = plugin
			}
		}

		// Take the first non-empty default project
		if merged.DefaultProject == "" && cfg.DefaultProject != "" {
			merged.DefaultProject = cfg.DefaultProject
		}

		// Merge defaults (take first non-zero values)
		mergeDefaults(&merged.Defaults, &cfg.Defaults)
	}

	return merged, nil
}

// mergeDefaults merges default configurations, preferring non-zero values
func mergeDefaults(target, source *DefaultsConfig) {
	// Test defaults
	if target.Test.Processes == 0 && source.Test.Processes != 0 {
		target.Test.Processes = source.Test.Processes
	}
	if !target.Test.Parallel && source.Test.Parallel {
		target.Test.Parallel = source.Test.Parallel
	}
	if !target.Test.Coverage && source.Test.Coverage {
		target.Test.Coverage = source.Test.Coverage
	}
	if !target.Test.Verbose && source.Test.Verbose {
		target.Test.Verbose = source.Test.Verbose
	}

	// Docker defaults
	if target.Docker.ComposeTimeout == 0 && source.Docker.ComposeTimeout != 0 {
		target.Docker.ComposeTimeout = source.Docker.ComposeTimeout
	}
	if !target.Docker.AutoStart && source.Docker.AutoStart {
		target.Docker.AutoStart = source.Docker.AutoStart
	}
	if !target.Docker.RemoveOrphans && source.Docker.RemoveOrphans {
		target.Docker.RemoveOrphans = source.Docker.RemoveOrphans
	}

	// Color defaults
	if target.Colors.Enabled == "" && source.Colors.Enabled != "" {
		target.Colors.Enabled = source.Colors.Enabled
	}

	// Worktree defaults
	if !target.Worktree.AutoSetup && source.Worktree.AutoSetup {
		target.Worktree.AutoSetup = source.Worktree.AutoSetup
	}
	if !target.Worktree.CopyEnv && source.Worktree.CopyEnv {
		target.Worktree.CopyEnv = source.Worktree.CopyEnv
	}
	if !target.Worktree.RunMigrations && source.Worktree.RunMigrations {
		target.Worktree.RunMigrations = source.Worktree.RunMigrations
	}
}
