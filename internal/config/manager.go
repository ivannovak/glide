package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/glide-cli/glide/v3/internal/context"
	"github.com/spf13/cobra"
)

// Manager handles configuration with precedence
type Manager struct {
	loader        *Loader
	config        *Config
	activeProject *ProjectConfig
	commandConfig *CommandConfig
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	return &Manager{
		loader: NewLoader(),
	}
}

// Initialize loads configuration and applies context
func (m *Manager) Initialize(ctx *context.ProjectContext) error {
	config, project, err := m.loader.LoadWithContext(ctx)
	if err != nil {
		return err
	}

	m.config = config
	m.activeProject = project
	m.commandConfig = m.buildCommandConfig()

	return nil
}

// ApplyFlags applies command-line flags to override configuration
func (m *Manager) ApplyFlags(cmd *cobra.Command) {
	if m.commandConfig == nil {
		m.commandConfig = m.buildCommandConfig()
	}

	// Test flags
	if cmd.Flags().Changed("parallel") {
		val, _ := cmd.Flags().GetBool("parallel")
		m.commandConfig.Test.Parallel = val
	}
	if cmd.Flags().Changed("processes") {
		val, _ := cmd.Flags().GetInt("processes")
		m.commandConfig.Test.Processes = val
	}
	if cmd.Flags().Changed("coverage") {
		val, _ := cmd.Flags().GetBool("coverage")
		m.commandConfig.Test.Coverage = val
	}
	if cmd.Flags().Changed("verbose") {
		val, _ := cmd.Flags().GetBool("verbose")
		m.commandConfig.Test.Verbose = val
	}

	// Docker flags
	if cmd.Flags().Changed("timeout") {
		val, _ := cmd.Flags().GetInt("timeout")
		m.commandConfig.Docker.ComposeTimeout = val
	}
	if cmd.Flags().Changed("no-auto-start") {
		m.commandConfig.Docker.AutoStart = false
	}
	if cmd.Flags().Changed("remove-orphans") {
		m.commandConfig.Docker.RemoveOrphans = true
	}

	// Color flags
	if cmd.Flags().Changed("color") {
		val, _ := cmd.Flags().GetString("color")
		m.commandConfig.Colors.Enabled = (val == "always" || val == "true")
	}
	if cmd.Flags().Changed("no-color") {
		m.commandConfig.Colors.Enabled = false
	}

	// Worktree flags
	if cmd.Flags().Changed("auto-setup") {
		m.commandConfig.Worktree.AutoSetup = true
	}
	if cmd.Flags().Changed("no-copy-env") {
		m.commandConfig.Worktree.CopyEnv = false
	}
	if cmd.Flags().Changed("run-migrations") {
		m.commandConfig.Worktree.RunMigrations = true
	}
}

// buildCommandConfig creates runtime configuration with all precedence applied
func (m *Manager) buildCommandConfig() *CommandConfig {
	if m.config == nil {
		defaults := GetDefaults()
		m.config = &defaults
	}

	cc := &CommandConfig{
		ActiveProject: m.activeProject,
	}

	// Apply configuration precedence: defaults < config < environment

	// Test configuration
	cc.Test.Parallel = m.config.Defaults.Test.Parallel
	cc.Test.Processes = m.config.Defaults.Test.Processes
	cc.Test.Coverage = m.config.Defaults.Test.Coverage
	cc.Test.Verbose = m.config.Defaults.Test.Verbose

	// Check environment variables
	if val := os.Getenv("GLIDE_TEST_PARALLEL"); val != "" {
		cc.Test.Parallel = val == "true" || val == "1"
	}
	if val := os.Getenv("GLIDE_TEST_PROCESSES"); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			cc.Test.Processes = n
		}
	}
	if val := os.Getenv("GLIDE_TEST_COVERAGE"); val != "" {
		cc.Test.Coverage = val == "true" || val == "1"
	}

	// Docker configuration
	cc.Docker.ComposeTimeout = m.config.Defaults.Docker.ComposeTimeout
	cc.Docker.AutoStart = m.config.Defaults.Docker.AutoStart
	cc.Docker.RemoveOrphans = m.config.Defaults.Docker.RemoveOrphans

	if val := os.Getenv("GLIDE_DOCKER_TIMEOUT"); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			cc.Docker.ComposeTimeout = n
		}
	}
	if val := os.Getenv("GLIDE_DOCKER_AUTO_START"); val != "" {
		cc.Docker.AutoStart = val == "true" || val == "1"
	}

	// Color configuration
	cc.Colors.Enabled = m.determineColorEnabled(m.config.Defaults.Colors.Enabled)

	// Worktree configuration
	cc.Worktree.AutoSetup = m.config.Defaults.Worktree.AutoSetup
	cc.Worktree.CopyEnv = m.config.Defaults.Worktree.CopyEnv
	cc.Worktree.RunMigrations = m.config.Defaults.Worktree.RunMigrations

	if val := os.Getenv("GLIDE_WORKTREE_AUTO_SETUP"); val != "" {
		cc.Worktree.AutoSetup = val == "true" || val == "1"
	}

	return cc
}

// determineColorEnabled resolves color setting based on mode and environment
func (m *Manager) determineColorEnabled(mode string) bool {
	// Check environment variable first
	if val := os.Getenv("GLIDE_COLORS"); val != "" {
		mode = val
	}
	if val := os.Getenv("NO_COLOR"); val != "" {
		return false
	}

	switch strings.ToLower(mode) {
	case "always", "true", "1":
		return true
	case "never", "false", "0":
		return false
	case "auto":
		// Check if output is a terminal
		if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
			return true
		}
		return false
	default:
		return true // Default to enabled
	}
}

// GetConfig returns the loaded configuration
func (m *Manager) GetConfig() *Config {
	if m.config == nil {
		defaults := GetDefaults()
		return &defaults
	}
	return m.config
}

// GetCommandConfig returns the runtime command configuration
func (m *Manager) GetCommandConfig() *CommandConfig {
	if m.commandConfig == nil {
		m.commandConfig = m.buildCommandConfig()
	}
	return m.commandConfig
}

// GetActiveProject returns the currently active project
func (m *Manager) GetActiveProject() *ProjectConfig {
	return m.activeProject
}

// GetLoader returns the configuration loader
func (m *Manager) GetLoader() *Loader {
	return m.loader
}

// GetProjectByName returns a project configuration by name
func (m *Manager) GetProjectByName(name string) (*ProjectConfig, bool) {
	if m.config == nil {
		return nil, false
	}

	project, ok := m.config.Projects[name]
	if !ok {
		return nil, false
	}

	return &project, true
}

// SetDefaultProject sets the default project
func (m *Manager) SetDefaultProject(name string) error {
	if m.config == nil {
		return fmt.Errorf("no configuration loaded")
	}

	if _, ok := m.config.Projects[name]; !ok {
		return fmt.Errorf("project %s does not exist", name)
	}

	m.config.DefaultProject = name
	return m.loader.Save(m.config)
}
