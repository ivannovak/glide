package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ivannovak/glide/v2/internal/context"
	"github.com/ivannovak/glide/v2/pkg/branding"
	"gopkg.in/yaml.v3"
)

// Loader handles configuration loading and merging
type Loader struct {
	configPath string
	config     *Config
}

// NewLoader creates a new configuration loader
func NewLoader() *Loader {
	return &Loader{
		configPath: branding.GetConfigPath(),
	}
}

// Load loads the configuration from the config file
func (l *Loader) Load() (*Config, error) {
	// Start with defaults
	config := GetDefaults()

	// Check if config file exists
	if _, err := os.Stat(l.configPath); os.IsNotExist(err) {
		// No config file is not an error, just use defaults
		l.config = &config
		return l.config, nil
	}

	// Read config file
	data, err := os.ReadFile(l.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply defaults for any missing values
	l.applyDefaults(&config)

	// Validate configuration
	if err := l.validate(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	l.config = &config
	return l.config, nil
}

// LoadWithContext loads configuration and detects the active project
func (l *Loader) LoadWithContext(ctx *context.ProjectContext) (*Config, *ProjectConfig, error) {
	config, err := l.Load()
	if err != nil {
		return nil, nil, err
	}

	// Try to find matching project based on context
	activeProject := l.detectActiveProject(config, ctx)

	return config, activeProject, nil
}

// Save saves the configuration to ~/.glide.yml
func (l *Loader) Save(config *Config) error {
	// Ensure directory exists
	dir := filepath.Dir(l.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write file
	if err := os.WriteFile(l.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// AddProject adds a new project to the configuration
func (l *Loader) AddProject(name, path, mode string) error {
	config, err := l.Load()
	if err != nil {
		return err
	}

	if config.Projects == nil {
		config.Projects = make(map[string]ProjectConfig)
	}

	// Resolve absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	config.Projects[name] = ProjectConfig{
		Path: absPath,
		Mode: mode,
	}

	// Set as default if it's the first project
	if config.DefaultProject == "" {
		config.DefaultProject = name
	}

	return l.Save(config)
}

// detectActiveProject finds the project matching the current context
func (l *Loader) detectActiveProject(config *Config, ctx *context.ProjectContext) *ProjectConfig {
	if ctx == nil || ctx.ProjectRoot == "" {
		return nil
	}

	// Check each project to see if it matches our context
	for _, project := range config.Projects {
		// Resolve project path
		projectPath, err := filepath.Abs(project.Path)
		if err != nil {
			continue
		}

		// Check if context root matches project path
		if projectPath == ctx.ProjectRoot {
			proj := project // Create a copy
			return &proj
		}

		// Check if we're inside the project
		if strings.HasPrefix(ctx.ProjectRoot, projectPath) {
			proj := project // Create a copy
			return &proj
		}
	}

	// If no match found but we have a default project, check that
	if config.DefaultProject != "" {
		if proj, ok := config.Projects[config.DefaultProject]; ok {
			return &proj
		}
	}

	return nil
}

// applyDefaults fills in any missing configuration values with defaults
func (l *Loader) applyDefaults(config *Config) {
	defaults := GetDefaults()

	// Test defaults
	if config.Defaults.Test.Processes == 0 {
		config.Defaults.Test.Processes = defaults.Defaults.Test.Processes
	}

	// Docker defaults
	if config.Defaults.Docker.ComposeTimeout == 0 {
		config.Defaults.Docker.ComposeTimeout = defaults.Defaults.Docker.ComposeTimeout
	}

	// Color defaults
	if config.Defaults.Colors.Enabled == "" {
		config.Defaults.Colors.Enabled = defaults.Defaults.Colors.Enabled
	}

	// Initialize maps if needed
	if config.Projects == nil {
		config.Projects = make(map[string]ProjectConfig)
	}
}

// validate checks if the configuration is valid
func (l *Loader) validate(config *Config) error {
	// Validate projects
	for name, project := range config.Projects {
		if project.Path == "" {
			return fmt.Errorf("project %s has no path", name)
		}

		// Validate mode
		if project.Mode != "" && project.Mode != "multi-worktree" && project.Mode != "single-repo" {
			return fmt.Errorf("project %s has invalid mode: %s", name, project.Mode)
		}
	}

	// Validate test settings
	if config.Defaults.Test.Processes < 1 || config.Defaults.Test.Processes > 100 {
		return fmt.Errorf("invalid test processes value: %d (must be 1-100)", config.Defaults.Test.Processes)
	}

	// Validate color settings
	validColors := []string{"auto", "always", "never", ""}
	valid := false
	for _, v := range validColors {
		if config.Defaults.Colors.Enabled == v {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid color setting: %s (must be auto/always/never)", config.Defaults.Colors.Enabled)
	}

	// Validate default project exists if specified
	if config.DefaultProject != "" {
		if _, ok := config.Projects[config.DefaultProject]; !ok {
			return fmt.Errorf("default project %s does not exist", config.DefaultProject)
		}
	}

	return nil
}

// GetConfigPath returns the path to the config file
func (l *Loader) GetConfigPath() string {
	return l.configPath
}

// ConfigExists checks if a config file exists
func (l *Loader) ConfigExists() bool {
	_, err := os.Stat(l.configPath)
	return err == nil
}
