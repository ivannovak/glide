package app

import (
	"io"
	"os"

	"github.com/glide-cli/glide/v3/internal/config"
	"github.com/glide-cli/glide/v3/internal/context"
	"github.com/glide-cli/glide/v3/internal/shell"
	"github.com/glide-cli/glide/v3/pkg/container"
	"github.com/glide-cli/glide/v3/pkg/logging"
	"github.com/glide-cli/glide/v3/pkg/output"
	"go.uber.org/fx"
)

// Application is the main dependency container for the CLI.
//
// Deprecated: This type is deprecated and will be removed in v3.0.0.
// Use pkg/container.Container instead for dependency injection.
//
// The Application pattern was a service locator anti-pattern that has been
// replaced with proper dependency injection using uber-fx. All code should
// migrate to using container.New() and explicit dependency passing.
//
// Migration Guide: See docs/adr/ADR-013-dependency-injection.md
//
// This type is kept for backward compatibility only. The Application now uses
// the DI container internally but maintains the same API for existing code.
type Application struct {
	// Core dependencies
	OutputManager  *output.Manager
	ProjectContext *context.ProjectContext
	Config         *config.Config

	// Service dependencies
	// REMOVED: DockerResolver   *docker.Resolver      // Moved to plugin
	// REMOVED: ContainerManager *docker.ContainerManager // Moved to plugin
	ShellExecutor *shell.Executor
	ConfigLoader  *config.Loader

	// Configuration options
	Writer io.Writer

	// Internal container (not exported, used for lifecycle management)
	container *container.Container
}

// Option is a functional option for configuring the Application.
//
// Deprecated: This type will be removed in v3.0.0.
// Use container options (fx.Option) instead for dependency injection.
//
// Migration: Replace app.Option with fx.Option and use container.WithXXX functions.
type Option func(*Application)

// NewApplication creates a new application with dependencies.
//
// Deprecated: This function will be removed in v3.0.0.
// Use container.New() instead for dependency injection.
//
// Migration:
//
//	// Old:
//	app := app.NewApplication(
//	    app.WithOutputFormat(output.FormatJSON, false, false),
//	)
//	cli := New(app)
//
//	// New:
//	outputMgr := output.NewManager(output.FormatJSON, false, false, os.Stdout)
//	ctx, _ := context.DetectWithExtensions()
//	cfg, _ := config.Load()
//	cli := New(outputMgr, ctx, cfg)
//
// The function now uses the DI container internally, extracting dependencies
// and populating the Application struct fields for backward compatibility.
func NewApplication(opts ...Option) *Application {
	app := &Application{
		Writer: os.Stdout,
	}

	// Apply old-style options to the app struct first
	// This captures user-provided overrides before container creation
	for _, opt := range opts {
		opt(app)
	}

	// If user provided any dependencies directly that cannot be easily overridden
	// via container options, fall back to old behavior entirely.
	// This ensures backward compatibility with all existing option patterns.
	if app.OutputManager != nil || app.ShellExecutor != nil || app.ConfigLoader != nil ||
		app.Config != nil || app.ProjectContext != nil {
		initializeFallback(app)
		return app
	}

	// At this point, user only provided Writer (or nothing).
	// We can safely use the container.

	// Convert old-style options to container options
	containerOpts := convertToContainerOptions(app)

	// Create variables to populate from the container
	var (
		outputManager  *output.Manager
		projectContext *context.ProjectContext
		cfg            *config.Config
		shellExecutor  *shell.Executor
		configLoader   *config.Loader
		logger         *logging.Logger
	)

	// Add fx.Populate to extract dependencies
	containerOpts = append(containerOpts, fx.Populate(
		&outputManager,
		&projectContext,
		&cfg,
		&shellExecutor,
		&configLoader,
		&logger,
	))

	// Create the container with dependency extraction
	c, err := container.New(containerOpts...)
	if err != nil {
		// Fallback to old behavior if container creation fails
		initializeFallback(app)
		return app
	}

	// Populate app fields from extracted dependencies
	app.OutputManager = outputManager
	app.ProjectContext = projectContext
	app.Config = cfg
	app.ShellExecutor = shellExecutor
	app.ConfigLoader = configLoader
	app.container = c

	// Update the output manager's writer if user provided one
	if app.Writer != nil && app.Writer != os.Stdout {
		if app.OutputManager != nil {
			app.OutputManager.SetWriter(app.Writer)
		}
	}

	return app
}

// convertToContainerOptions converts old Application options to container fx.Options.
func convertToContainerOptions(app *Application) []fx.Option {
	var opts []fx.Option

	// Convert Writer option
	if app.Writer != nil && app.Writer != os.Stdout {
		opts = append(opts, container.WithWriter(app.Writer))
	}

	// Convert Config option
	if app.Config != nil {
		opts = append(opts, container.WithConfig(app.Config))
	}

	// Convert ProjectContext option
	if app.ProjectContext != nil {
		opts = append(opts, container.WithProjectContext(app.ProjectContext))
	}

	return opts
}

// initializeFallback initializes the application using the old pattern.
// This is used when container initialization fails or when user provides
// dependencies directly via options that can't be overridden in the container.
func initializeFallback(app *Application) {
	// Initialize output manager if not provided
	if app.OutputManager == nil {
		app.OutputManager = output.NewManager(
			output.FormatTable,
			false,
			false,
			app.Writer,
		)
	}

	// Initialize shell executor if not provided
	if app.ShellExecutor == nil {
		app.ShellExecutor = shell.NewExecutor(shell.Options{})
	}

	// ConfigLoader will be created lazily via GetConfigLoader
}

// WithOutputManager sets a custom output manager.
//
// Deprecated: This option will be removed in v3.0.0.
// This triggers fallback mode and bypasses the DI container.
// For new code, create dependencies via container.New() with custom providers.
func WithOutputManager(manager *output.Manager) Option {
	return func(app *Application) {
		app.OutputManager = manager
	}
}

// WithOutputFormat configures the output format.
//
// Deprecated: This option will be removed in v3.0.0.
// This triggers fallback mode and bypasses the DI container.
// For new code, configure output via the OutputManager from the container.
func WithOutputFormat(format output.Format, quiet, noColor bool) Option {
	return func(app *Application) {
		app.OutputManager = output.NewManager(format, quiet, noColor, app.Writer)
	}
}

// WithProjectContext sets the project context.
//
// Deprecated: This option will be removed in v3.0.0.
// Use container.WithProjectContext() instead.
func WithProjectContext(ctx *context.ProjectContext) Option {
	return func(app *Application) {
		app.ProjectContext = ctx
	}
}

// WithConfig sets the configuration.
//
// Deprecated: This option will be removed in v3.0.0.
// Use container.WithConfig() instead.
func WithConfig(cfg *config.Config) Option {
	return func(app *Application) {
		app.Config = cfg
	}
}

// WithShellExecutor sets a custom shell executor.
//
// Deprecated: This option will be removed in v3.0.0.
// This triggers fallback mode and bypasses the DI container.
// For new code, use the ShellExecutor from the container.
func WithShellExecutor(executor *shell.Executor) Option {
	return func(app *Application) {
		app.ShellExecutor = executor
	}
}

// WithConfigLoader sets a custom config loader.
//
// Deprecated: This option will be removed in v3.0.0.
// This triggers fallback mode and bypasses the DI container.
// For new code, use the ConfigLoader from the container.
func WithConfigLoader(loader *config.Loader) Option {
	return func(app *Application) {
		app.ConfigLoader = loader
	}
}

// WithWriter sets a custom output writer.
//
// Deprecated: This option will be removed in v3.0.0.
// Use container.WithWriter() instead.
func WithWriter(writer io.Writer) Option {
	return func(app *Application) {
		app.Writer = writer
		// Update output manager if it exists
		if app.OutputManager != nil {
			app.OutputManager.SetWriter(writer)
		}
	}
}

// GetShellExecutor returns the shell executor.
//
// Deprecated: This method will be removed in v3.0.0.
// Access dependencies directly from the container instead.
func (app *Application) GetShellExecutor() *shell.Executor {
	return app.ShellExecutor
}

// GetConfigLoader returns a config loader, creating one if needed.
//
// Deprecated: This method will be removed in v3.0.0.
// Access dependencies directly from the container instead.
func (app *Application) GetConfigLoader() *config.Loader {
	if app.ConfigLoader == nil {
		app.ConfigLoader = config.NewLoader()
	}
	return app.ConfigLoader
}
