// Package app provides the application container for the Glide CLI.
//
// Deprecated: This package is deprecated in favor of pkg/container.
// The Application type was a service locator anti-pattern that has been
// replaced with proper dependency injection using uber-fx.
//
// All new code should use container.New() for dependency injection.
// See docs/adr/ADR-013-dependency-injection.md for migration guidance.
//
// This package is maintained only for backward compatibility and will
// be removed in v3.0.0.
//
// # Migration
//
// Before (using Application):
//
//	app, err := app.NewApplication()
//	// app.Config, app.OutputManager, etc.
//
// After (using Container):
//
//	c, err := container.New()
//	c.Run(ctx, func(cfg *config.Config, out *output.Manager) error {
//	    // Dependencies are injected
//	    return nil
//	})
package app
