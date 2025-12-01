package config

// CommandMap handles both simple string and structured Command formats
type CommandMap map[string]interface{}

// Command represents a user-defined command
type Command struct {
	// The actual command(s) to execute
	Cmd string `yaml:"cmd"`

	// Optional fields for structured format
	Alias       string `yaml:"alias,omitempty"`
	Description string `yaml:"description,omitempty"`
	Help        string `yaml:"help,omitempty"`
	Category    string `yaml:"category,omitempty"`
}

// Config represents the global Glide configuration
type Config struct {
	Projects       map[string]ProjectConfig `yaml:"projects"`
	DefaultProject string                   `yaml:"default_project"`
	Defaults       DefaultsConfig           `yaml:"defaults"`
	Commands       CommandMap               `yaml:"commands,omitempty"`

	// NOTE: Plugin configuration has been migrated to the type-safe pkg/config system.
	// Plugins register their typed configs using config.Register() in their init() functions,
	// and the config loader automatically updates them from YAML via the raw plugin configs map.
	// See pkg/config/MIGRATION.md for details.
}

// ProjectConfig represents a single project configuration
type ProjectConfig struct {
	Path     string     `yaml:"path"`
	Mode     string     `yaml:"mode"` // multi-worktree or single-repo
	Commands CommandMap `yaml:"commands,omitempty"`
}

// DefaultsConfig contains default settings
type DefaultsConfig struct {
	Test     TestDefaults     `yaml:"test"`
	Docker   DockerDefaults   `yaml:"docker"`
	Colors   ColorDefaults    `yaml:"colors"`
	Worktree WorktreeDefaults `yaml:"worktree"`
	Update   UpdateDefaults   `yaml:"update"`
}

// UpdateDefaults contains update notification settings
type UpdateDefaults struct {
	// CheckEnabled controls whether automatic update checks are performed
	CheckEnabled bool `yaml:"check_enabled"`
	// CheckIntervalHours is the number of hours between update checks (default: 24)
	CheckIntervalHours int `yaml:"check_interval_hours"`
	// NotifyEnabled controls whether update notifications are shown
	NotifyEnabled bool `yaml:"notify_enabled"`
}

// TestDefaults contains default test settings
type TestDefaults struct {
	Parallel  bool `yaml:"parallel"`
	Processes int  `yaml:"processes"`
	Coverage  bool `yaml:"coverage"`
	Verbose   bool `yaml:"verbose"`
}

// DockerDefaults contains default Docker settings
type DockerDefaults struct {
	ComposeTimeout int  `yaml:"compose_timeout"`
	AutoStart      bool `yaml:"auto_start"`
	RemoveOrphans  bool `yaml:"remove_orphans"`
}

// ColorDefaults contains color output settings
type ColorDefaults struct {
	Enabled string `yaml:"enabled"` // auto, always, never
}

// WorktreeDefaults contains worktree-related defaults
type WorktreeDefaults struct {
	AutoSetup     bool `yaml:"auto_setup"`
	CopyEnv       bool `yaml:"copy_env"`
	RunMigrations bool `yaml:"run_migrations"`
}

// CommandConfig represents runtime configuration with precedence applied
type CommandConfig struct {
	// Merged configuration from all sources
	Test     TestConfig
	Docker   DockerConfig
	Colors   ColorConfig
	Worktree WorktreeConfig

	// Currently active project
	ActiveProject *ProjectConfig
}

// TestConfig represents runtime test configuration
type TestConfig struct {
	Parallel  bool
	Processes int
	Coverage  bool
	Verbose   bool
	Args      []string // Additional arguments passed through
}

// DockerConfig represents runtime Docker configuration
type DockerConfig struct {
	ComposeTimeout int
	AutoStart      bool
	RemoveOrphans  bool
	ComposeFiles   []string // Resolved compose files
}

// ColorConfig represents runtime color configuration
type ColorConfig struct {
	Enabled bool
}

// WorktreeConfig represents runtime worktree configuration
type WorktreeConfig struct {
	AutoSetup     bool
	CopyEnv       bool
	RunMigrations bool
}

// GetDefaults returns a Config with all default values
func GetDefaults() Config {
	return Config{
		Projects: make(map[string]ProjectConfig),
		Defaults: DefaultsConfig{
			Test: TestDefaults{
				Parallel:  true,
				Processes: 3,
				Coverage:  false,
				Verbose:   false,
			},
			Docker: DockerDefaults{
				ComposeTimeout: 30,
				AutoStart:      true,
				RemoveOrphans:  false,
			},
			Colors: ColorDefaults{
				Enabled: "auto",
			},
			Worktree: WorktreeDefaults{
				AutoSetup:     false,
				CopyEnv:       true,
				RunMigrations: false,
			},
			Update: UpdateDefaults{
				CheckEnabled:       true,
				CheckIntervalHours: 24,
				NotifyEnabled:      true,
			},
		},
	}
}
