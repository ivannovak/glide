package container

//lint:file-ignore SA1019 plugin.Plugin is deprecated but still valid for use

import (
	"fmt"
	"io"
	"os"

	"github.com/ivannovak/glide/v2/internal/config"
	"github.com/ivannovak/glide/v2/internal/context"
	"github.com/ivannovak/glide/v2/internal/shell"
	"github.com/ivannovak/glide/v2/pkg/logging"
	"github.com/ivannovak/glide/v2/pkg/output"
	"github.com/ivannovak/glide/v2/pkg/plugin"
	"go.uber.org/fx"
)

// Provider functions create and configure application dependencies.
// These are called by uber-fx in dependency order.

// provideLogger creates the application logger.
//
// The logger is configured from environment variables:
//   - GLIDE_LOG_LEVEL: debug, info, warn, error
//   - GLIDE_LOG_FORMAT: text, json
//   - GLIDE_DEBUG: enables debug logging
func provideLogger() *logging.Logger {
	return logging.New(logging.FromEnv())
}

// provideWriter provides the output writer.
//
// Defaults to os.Stdout. Can be overridden in tests using WithWriter().
func provideWriter() io.Writer {
	return os.Stdout
}

// provideConfigLoader creates the configuration loader.
func provideConfigLoader(logger *logging.Logger) *config.Loader {
	logger.Debug("Creating config loader")
	return config.NewLoader()
}

// ConfigParams groups dependencies for config provider.
type ConfigParams struct {
	fx.In

	Loader *config.Loader
	Logger *logging.Logger
}

// provideConfig loads the application configuration.
//
// If the config file doesn't exist, returns a default config.
// Only returns an error for actual loading failures.
func provideConfig(params ConfigParams) (*config.Config, error) {
	params.Logger.Debug("Loading configuration")

	cfg, err := params.Loader.Load()
	if err != nil {
		// If file doesn't exist, use default config
		if os.IsNotExist(err) {
			params.Logger.Debug("Config file not found, using defaults")
			return &config.Config{}, nil
		}
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	params.Logger.Debug("Configuration loaded successfully")
	return cfg, nil
}

// provideContextDetector creates the project context detector.
func provideContextDetector(logger *logging.Logger) (*context.Detector, error) {
	logger.Debug("Creating context detector")
	return context.NewDetector()
}

// ProjectContextParams groups dependencies for project context provider.
type ProjectContextParams struct {
	fx.In

	Detector *context.Detector
	Plugins  []*plugin.Plugin `optional:"true"` // Get all registered plugins
	Logger   *logging.Logger
}

// provideProjectContext detects the project context.
//
// Uses the detector to identify frameworks and tools in the project.
// Plugins can contribute additional detection logic via the extension registry.
func provideProjectContext(params ProjectContextParams) (*context.ProjectContext, error) {
	params.Logger.Debug("Detecting project context")

	// TODO: Support plugin-provided extensions via extension registry
	// For now, just detect without extensions
	ctx, err := params.Detector.Detect()
	if err != nil {
		params.Logger.Error("Failed to detect project context", "error", err)
		return nil, fmt.Errorf("failed to detect project context: %w", err)
	}

	params.Logger.Debug("Project context detected")
	return ctx, nil
}

// OutputManagerParams groups dependencies for output manager provider.
type OutputManagerParams struct {
	fx.In

	Writer io.Writer
	Logger *logging.Logger
}

// provideOutputManager creates the output manager.
//
// The output manager handles formatted output to the user.
// Uses table format by default. Can be configured via CLI flags.
func provideOutputManager(params OutputManagerParams) *output.Manager {
	params.Logger.Debug("Creating output manager")
	return output.NewManager(
		output.FormatTable, // Default format, can be overridden
		false,              // quiet
		false,              // noColor
		params.Writer,
	)
}

// provideShellExecutor creates the shell command executor.
func provideShellExecutor(logger *logging.Logger) *shell.Executor {
	logger.Debug("Creating shell executor")
	return shell.NewExecutor(shell.Options{})
}

// providePluginRegistry creates the plugin registry.
func providePluginRegistry(logger *logging.Logger) *plugin.Registry {
	logger.Debug("Creating plugin registry")
	return plugin.NewRegistry()
}
