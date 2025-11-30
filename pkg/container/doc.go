// Package container provides dependency injection using uber-fx.
//
// The container manages application dependencies and their lifecycle,
// replacing the deprecated pkg/app.Application god object pattern with
// proper dependency injection.
//
// # Basic Usage
//
//	c, err := container.New()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	err = c.Run(ctx, func(cfg *config.Config, out *output.Manager) error {
//	    // Dependencies are automatically injected
//	    out.Success("Configuration loaded from: %s", cfg.Path)
//	    return nil
//	})
//
// # Default Providers
//
// The container automatically provides these dependencies:
//   - *logging.Logger - Structured logging
//   - *config.Loader and *config.Config - Configuration management
//   - *context.Detector and *context.ProjectContext - Project context detection
//   - *output.Manager - Output formatting and display
//   - *shell.Executor - Shell command execution
//   - *plugin.Registry - Plugin management
//
// # Custom Providers
//
// Override default providers for testing or customization:
//
//	c, err := container.New(
//	    fx.Replace(fx.Annotate(
//	        mockConfig,
//	        fx.As(new(*config.Config)),
//	    )),
//	)
//
// # Lifecycle Management
//
// The container manages startup and shutdown of all registered components:
//
//	c.Run(ctx, func() error {
//	    // All dependencies are started
//	    <-ctx.Done()
//	    // Graceful shutdown on context cancellation
//	    return nil
//	})
//
// See docs/adr/ADR-013-dependency-injection.md for design rationale.
package container
