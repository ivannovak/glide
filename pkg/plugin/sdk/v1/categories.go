package v1

// Command category constants for organizing commands in help output.
// Categories are displayed in priority order (lower numbers appear first).
const (
	// CategoryCore represents essential development commands
	// Priority: 10
	// Use for: Version info, plugin management, self-update
	CategoryCore = "core"

	// CategoryGlobal represents multi-worktree management commands
	// Priority: 20
	// Use for: Commands that operate across all worktrees
	CategoryGlobal = "global"

	// CategorySetup represents project setup and configuration
	// Priority: 30
	// Use for: Initial setup, configuration, shell completions
	CategorySetup = "setup"

	// CategoryDocker represents container and service control
	// Priority: 40
	// Use for: Docker compose operations, container management
	CategoryDocker = "docker"

	// CategoryTesting represents test execution and coverage
	// Priority: 50
	// Use for: Test runners, coverage tools, test utilities
	CategoryTesting = "testing"

	// CategoryDeveloper represents code quality and development utilities
	// Priority: 60
	// Use for: Linters, formatters, build tools, code generators
	CategoryDeveloper = "developer"

	// CategoryDatabase represents database management and access
	// Priority: 70
	// Use for: Migrations, database CLI, backups, schema tools
	CategoryDatabase = "database"

	// CategoryPlugin represents commands from plugins (default if not specified)
	// Priority: 80
	// Use for: Plugin commands without explicit category
	CategoryPlugin = "plugin"

	// CategoryHelp represents help topics and documentation
	// Priority: 90
	// Use for: Documentation commands, help topics
	CategoryHelp = "help"

	// CategoryDebug represents debug and diagnostic commands
	// Priority: 100
	// Use for: Debug utilities, hidden by default in help
	CategoryDebug = "debug"
)

// CategoryInfo provides metadata about command categories
type CategoryInfo struct {
	ID          string // Category constant (e.g., CategoryDocker)
	Name        string // Display name (e.g., "Docker Management")
	Description string // Category description
	Priority    int    // Display priority (lower = higher priority)
}

// Categories provides metadata for all available categories
var Categories = map[string]CategoryInfo{
	CategoryCore: {
		ID:          CategoryCore,
		Name:        "Core Commands",
		Description: "Essential development commands",
		Priority:    10,
	},
	CategoryGlobal: {
		ID:          CategoryGlobal,
		Name:        "Global Commands",
		Description: "Multi-worktree management",
		Priority:    20,
	},
	CategorySetup: {
		ID:          CategorySetup,
		Name:        "Setup & Configuration",
		Description: "Project setup and configuration",
		Priority:    30,
	},
	CategoryDocker: {
		ID:          CategoryDocker,
		Name:        "Docker Management",
		Description: "Container and service control",
		Priority:    40,
	},
	CategoryTesting: {
		ID:          CategoryTesting,
		Name:        "Testing",
		Description: "Test execution and coverage",
		Priority:    50,
	},
	CategoryDeveloper: {
		ID:          CategoryDeveloper,
		Name:        "Development Tools",
		Description: "Code quality and utilities",
		Priority:    60,
	},
	CategoryDatabase: {
		ID:          CategoryDatabase,
		Name:        "Database",
		Description: "Database management and access",
		Priority:    70,
	},
	CategoryPlugin: {
		ID:          CategoryPlugin,
		Name:        "Plugin Commands",
		Description: "Commands from installed plugins",
		Priority:    80,
	},
	CategoryHelp: {
		ID:          CategoryHelp,
		Name:        "Help & Documentation",
		Description: "Help topics and guides",
		Priority:    90,
	},
	CategoryDebug: {
		ID:          CategoryDebug,
		Name:        "Debug",
		Description: "Debug and diagnostic utilities",
		Priority:    100,
	},
}

// Visibility constants for command context visibility
const (
	VisibilityAlways       = "always"        // Show in all contexts (default)
	VisibilityProjectOnly  = "project-only"  // Show only when in a project
	VisibilityWorktreeOnly = "worktree-only" // Show only in worktrees (not at multi-worktree root)
	VisibilityRootOnly     = "root-only"     // Show only at multi-worktree root
	VisibilityNonRoot      = "non-root"      // Show everywhere except multi-worktree root
)
