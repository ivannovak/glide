package container

import (
	"io"

	"github.com/ivannovak/glide/v2/internal/config"
	"github.com/ivannovak/glide/v2/internal/context"
	"github.com/ivannovak/glide/v2/pkg/logging"
	"go.uber.org/fx"
)

// Option is a functional option for configuring the container.
//
// Options are typically used in tests to override default providers.
type Option = fx.Option

// WithLogger overrides the logger provider.
//
// Useful in tests to capture log output or disable logging.
//
// Example:
//
//	testLogger := logging.New(&logging.Config{Level: slog.LevelDebug})
//	c, _ := container.New(container.WithLogger(testLogger))
func WithLogger(logger *logging.Logger) Option {
	return fx.Replace(func() *logging.Logger {
		return logger
	})
}

// WithWriter overrides the output writer.
//
// Useful in tests to capture output to a buffer.
//
// Example:
//
//	buf := &bytes.Buffer{}
//	c, _ := container.New(container.WithWriter(buf))
func WithWriter(w io.Writer) Option {
	return fx.Replace(func() io.Writer {
		return w
	})
}

// WithConfig overrides the config provider.
//
// Useful in tests to provide a specific configuration.
//
// Example:
//
//	testCfg := &config.Config{}
//	c, _ := container.New(container.WithConfig(testCfg))
func WithConfig(cfg *config.Config) Option {
	return fx.Replace(func(params ConfigParams) (*config.Config, error) {
		return cfg, nil
	})
}

// WithProjectContext overrides the project context provider.
//
// Useful in tests to provide a specific project context.
//
// Example:
//
//	testCtx := &context.ProjectContext{}
//	c, _ := container.New(container.WithProjectContext(testCtx))
func WithProjectContext(ctx *context.ProjectContext) Option {
	return fx.Replace(func(params ProjectContextParams) (*context.ProjectContext, error) {
		return ctx, nil
	})
}

// WithoutLifecycle disables lifecycle hooks for faster tests.
//
// This prevents OnStart and OnStop hooks from executing,
// which can speed up tests that don't need full initialization.
//
// Example:
//
//	c, _ := container.New(container.WithoutLifecycle())
func WithoutLifecycle() Option {
	return fx.Options(
		// Skip lifecycle invocations
		fx.Invoke(func() {
			// No-op instead of registerLifecycleHooks
		}),
	)
}
