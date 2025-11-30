package testutil

import (
	"bytes"
	"os"

	"github.com/ivannovak/glide/v2/internal/config"
	"github.com/ivannovak/glide/v2/internal/context"
)

// ContextOption is a functional option for configuring test ProjectContext
type ContextOption func(*context.ProjectContext)

// ConfigOption is a functional option for configuring test Config
type ConfigOption func(*config.Config)

// NewTestContext creates a test ProjectContext with sensible defaults
// and allows customization via functional options
func NewTestContext(opts ...ContextOption) *context.ProjectContext {
	ctx := &context.ProjectContext{
		WorkingDir:         "/test/project",
		ProjectRoot:        "/test/project",
		ProjectName:        "test-project",
		DevelopmentMode:    context.ModeSingleRepo,
		Location:           context.LocationProject,
		IsRoot:             false,
		IsMainRepo:         false,
		IsWorktree:         false,
		WorktreeName:       "",
		Extensions:         make(map[string]interface{}),
		ComposeFiles:       []string{},
		DockerRunning:      false,
		ContainersStatus:   make(map[string]context.ContainerStatus),
		DetectedFrameworks: []string{},
		FrameworkVersions:  make(map[string]string),
		FrameworkCommands:  make(map[string]string),
		FrameworkMetadata:  make(map[string]map[string]string),
		CommandScope:       "local",
		Error:              nil,
	}

	for _, opt := range opts {
		opt(ctx)
	}

	return ctx
}

// WithWorkingDir sets the working directory
func WithWorkingDir(dir string) ContextOption {
	return func(ctx *context.ProjectContext) {
		ctx.WorkingDir = dir
	}
}

// WithProjectRoot sets the project root
func WithProjectRoot(root string) ContextOption {
	return func(ctx *context.ProjectContext) {
		ctx.ProjectRoot = root
	}
}

// WithProjectName sets the project name
func WithProjectName(name string) ContextOption {
	return func(ctx *context.ProjectContext) {
		ctx.ProjectName = name
	}
}

// WithDevelopmentMode sets the development mode
func WithDevelopmentMode(mode context.DevelopmentMode) ContextOption {
	return func(ctx *context.ProjectContext) {
		ctx.DevelopmentMode = mode
	}
}

// WithLocation sets the location type
func WithLocation(location context.LocationType) ContextOption {
	return func(ctx *context.ProjectContext) {
		ctx.Location = location
	}
}

// WithMultiWorktreeRoot configures context as multi-worktree root
func WithMultiWorktreeRoot() ContextOption {
	return func(ctx *context.ProjectContext) {
		ctx.DevelopmentMode = context.ModeMultiWorktree
		ctx.Location = context.LocationRoot
		ctx.IsRoot = true
		ctx.IsMainRepo = false
		ctx.IsWorktree = false
	}
}

// WithMultiWorktreeMainRepo configures context as multi-worktree main repo
func WithMultiWorktreeMainRepo() ContextOption {
	return func(ctx *context.ProjectContext) {
		ctx.DevelopmentMode = context.ModeMultiWorktree
		ctx.Location = context.LocationMainRepo
		ctx.IsRoot = false
		ctx.IsMainRepo = true
		ctx.IsWorktree = false
	}
}

// WithWorktree configures context as a worktree
func WithWorktree(name string) ContextOption {
	return func(ctx *context.ProjectContext) {
		ctx.DevelopmentMode = context.ModeMultiWorktree
		ctx.Location = context.LocationWorktree
		ctx.IsRoot = false
		ctx.IsMainRepo = false
		ctx.IsWorktree = true
		ctx.WorktreeName = name
	}
}

// WithDockerRunning sets Docker as running
func WithDockerRunning(running bool) ContextOption {
	return func(ctx *context.ProjectContext) {
		ctx.DockerRunning = running
	}
}

// WithComposeFiles sets compose files
func WithComposeFiles(files ...string) ContextOption {
	return func(ctx *context.ProjectContext) {
		ctx.ComposeFiles = files
	}
}

// WithExtension adds a plugin extension
func WithExtension(name string, data interface{}) ContextOption {
	return func(ctx *context.ProjectContext) {
		if ctx.Extensions == nil {
			ctx.Extensions = make(map[string]interface{})
		}
		ctx.Extensions[name] = data
	}
}

// WithCommandScope sets the command scope
func WithCommandScope(scope string) ContextOption {
	return func(ctx *context.ProjectContext) {
		ctx.CommandScope = scope
	}
}

// WithContextError sets an error on the context
func WithContextError(err error) ContextOption {
	return func(ctx *context.ProjectContext) {
		ctx.Error = err
	}
}

// NewTestConfig creates a test Config with sensible defaults
func NewTestConfig(opts ...ConfigOption) *config.Config {
	cfg := &config.Config{
		DefaultProject: "test-project",
		Projects: map[string]config.ProjectConfig{
			"test-project": {
				Path:     "/test/project",
				Mode:     "single-repo",
				Commands: make(config.CommandMap),
			},
		},
		Defaults: config.DefaultsConfig{
			Test: config.TestDefaults{
				Parallel:  false,
				Processes: 1,
				Coverage:  false,
				Verbose:   false,
			},
			Docker: config.DockerDefaults{
				ComposeTimeout: 30,
				AutoStart:      false,
				RemoveOrphans:  true,
			},
			Colors: config.ColorDefaults{
				Enabled: "auto",
			},
			Worktree: config.WorktreeDefaults{
				AutoSetup:     true,
				CopyEnv:       true,
				RunMigrations: false,
			},
		},
		// NOTE: Plugin configuration migrated to pkg/config type-safe system
		Commands: make(config.CommandMap),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

// WithDefaultProject sets the default project name
func WithDefaultProject(name string) ConfigOption {
	return func(cfg *config.Config) {
		cfg.DefaultProject = name
	}
}

// WithProject adds a project configuration
func WithProject(name, path, mode string) ConfigOption {
	return func(cfg *config.Config) {
		if cfg.Projects == nil {
			cfg.Projects = make(map[string]config.ProjectConfig)
		}
		cfg.Projects[name] = config.ProjectConfig{
			Path:     path,
			Mode:     mode,
			Commands: make(config.CommandMap),
		}
	}
}

// WithTestDefaults sets test defaults
func WithTestDefaults(parallel, coverage bool) ConfigOption {
	return func(cfg *config.Config) {
		cfg.Defaults.Test.Parallel = parallel
		cfg.Defaults.Test.Coverage = coverage
	}
}

// WithCommand adds a command to the config
func WithCommand(name, cmd string) ConfigOption {
	return func(cfg *config.Config) {
		if cfg.Commands == nil {
			cfg.Commands = make(config.CommandMap)
		}
		cfg.Commands[name] = cmd
	}
}

// NewTestWriter creates a bytes.Buffer for capturing test output
func NewTestWriter() *bytes.Buffer {
	return &bytes.Buffer{}
}

// TempDir creates a temporary directory for testing
// Returns the directory path and a cleanup function
func TempDir(t TestingT) (string, func()) {
	dir, err := os.MkdirTemp("", "glide-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	cleanup := func() {
		_ = os.RemoveAll(dir)
	}

	return dir, cleanup
}

// TestingT is a minimal interface for testing.T to avoid direct dependency
type TestingT interface {
	Fatalf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Helper()
}
