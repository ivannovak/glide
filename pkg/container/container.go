// Package container provides dependency injection using uber-fx.
//
// The container manages application dependencies and their lifecycle.
// It replaces the old pkg/app.Application God Object pattern with
// proper dependency injection.
//
// Example usage:
//
//	ctx := context.Background()
//	c, err := container.New()
//	if err != nil {
//	    return err
//	}
//
//	return c.Run(ctx, func() error {
//	    // Application logic
//	    return nil
//	})
package container

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/fx"
)

// Container wraps uber-fx and provides lifecycle management for the application.
type Container struct {
	app *fx.App
}

// New creates a new dependency injection container with the given options.
//
// The container automatically provides:
//   - Logger (from pkg/logging)
//   - ConfigLoader and Config (from internal/config)
//   - ContextDetector and ProjectContext (from internal/context)
//   - OutputManager (from pkg/output)
//   - ShellExecutor (from internal/shell)
//   - PluginRegistry (from pkg/plugin)
//
// Options can be used to override default providers for testing.
//
// Example:
//
//	c, err := container.New()
//	if err != nil {
//	    return err
//	}
func New(opts ...fx.Option) (*Container, error) {
	// Merge default options with user options
	allOpts := append(
		[]fx.Option{
			// Core providers - defined in providers.go
			fx.Provide(
				// Logging (no dependencies)
				provideLogger,

				// Writer (no dependencies)
				provideWriter,

				// Config (depends on logger)
				provideConfigLoader,
				provideConfig,

				// Context (depends on config, logger)
				provideContextDetector,
				provideProjectContext,

				// Output (depends on writer, logger)
				provideOutputManager,

				// Shell (depends on logger)
				provideShellExecutor,

				// Plugin registry (depends on logger)
				providePluginRegistry,
			),

			// Lifecycle hooks - defined in lifecycle.go
			fx.Invoke(registerLifecycleHooks),

			// Use NopLogger to suppress fx debug output by default
			fx.NopLogger,
		},
		opts...,
	)

	app := fx.New(allOpts...)
	if app.Err() != nil {
		return nil, fmt.Errorf("failed to create container: %w", app.Err())
	}

	return &Container{app: app}, nil
}

// Start starts the container and all managed components.
//
// This executes all OnStart hooks registered in the lifecycle.
// Components are started in dependency order.
//
// The context can be used to set a startup timeout.
func (c *Container) Start(ctx context.Context) error {
	if err := c.app.Start(ctx); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}
	return nil
}

// Stop gracefully shuts down the container.
//
// This executes all OnStop hooks registered in the lifecycle.
// Components are stopped in reverse dependency order.
//
// The context can be used to set a shutdown timeout.
func (c *Container) Stop(ctx context.Context) error {
	if err := c.app.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}
	return nil
}

// Run executes the application function with proper lifecycle management.
//
// It starts all components, runs the provided function, and ensures
// graceful shutdown even if the function panics or returns an error.
//
// The shutdown timeout is 10 seconds.
//
// Example:
//
//	c.Run(ctx, func() error {
//	    return rootCmd.Execute()
//	})
func (c *Container) Run(ctx context.Context, fn func() error) error {
	// Start all components
	if err := c.Start(ctx); err != nil {
		return err
	}

	// Ensure cleanup on exit
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = c.Stop(stopCtx)
	}()

	// Run the application function
	return fn()
}

// Invoke runs a function with dependencies injected from the container.
//
// This is useful for extracting dependencies after container creation.
// The container must be started first.
//
// Example:
//
//	c.Start(ctx)
//	defer c.Stop(ctx)
//	err := c.Invoke(func(l *logging.Logger) {
//	    // use logger
//	})
func (c *Container) Invoke(fn interface{}) error {
	// fx.App doesn't expose Invoke directly, so we need to extract during Start
	// This is a limitation - users should get dependencies via Run() or provider extraction
	return fmt.Errorf("Invoke not supported - extract dependencies in Run() function instead")
}
