package app

import (
	"io"
	"os"

	"github.com/ivannovak/glide/internal/config"
	"github.com/ivannovak/glide/internal/context"
	"github.com/ivannovak/glide/internal/shell"
	"github.com/ivannovak/glide/pkg/output"
)

// Application is the main dependency container for the CLI
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
}

// Option is a functional option for configuring the Application
type Option func(*Application)

// NewApplication creates a new application with dependencies
func NewApplication(opts ...Option) *Application {
	app := &Application{
		Writer: os.Stdout,
	}

	// Apply options
	for _, opt := range opts {
		opt(app)
	}

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

	return app
}

// WithOutputManager sets a custom output manager
func WithOutputManager(manager *output.Manager) Option {
	return func(app *Application) {
		app.OutputManager = manager
	}
}

// WithOutputFormat configures the output format
func WithOutputFormat(format output.Format, quiet, noColor bool) Option {
	return func(app *Application) {
		app.OutputManager = output.NewManager(format, quiet, noColor, app.Writer)
	}
}

// WithProjectContext sets the project context
func WithProjectContext(ctx *context.ProjectContext) Option {
	return func(app *Application) {
		app.ProjectContext = ctx
	}
}

// WithConfig sets the configuration
func WithConfig(cfg *config.Config) Option {
	return func(app *Application) {
		app.Config = cfg
	}
}

// WithShellExecutor sets a custom shell executor
func WithShellExecutor(executor *shell.Executor) Option {
	return func(app *Application) {
		app.ShellExecutor = executor
	}
}

// WithConfigLoader sets a custom config loader
func WithConfigLoader(loader *config.Loader) Option {
	return func(app *Application) {
		app.ConfigLoader = loader
	}
}

// WithWriter sets a custom output writer
func WithWriter(writer io.Writer) Option {
	return func(app *Application) {
		app.Writer = writer
		// Update output manager if it exists
		if app.OutputManager != nil {
			app.OutputManager.SetWriter(writer)
		}
	}
}

// GetShellExecutor returns the shell executor
func (app *Application) GetShellExecutor() *shell.Executor {
	return app.ShellExecutor
}

// GetConfigLoader returns a config loader, creating one if needed
func (app *Application) GetConfigLoader() *config.Loader {
	if app.ConfigLoader == nil {
		app.ConfigLoader = config.NewLoader()
	}
	return app.ConfigLoader
}
