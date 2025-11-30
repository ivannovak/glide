package container

import (
	"context"

	"github.com/ivannovak/glide/v3/pkg/logging"
	"go.uber.org/fx"
)

// LifecycleParams groups all components that need lifecycle management.
type LifecycleParams struct {
	fx.In

	Lifecycle fx.Lifecycle
	Logger    *logging.Logger
}

// registerLifecycleHooks registers startup and shutdown hooks for the application.
//
// This is called automatically by uber-fx when the container is created.
//
// Lifecycle hooks execute in dependency order:
//   - OnStart: from least dependent to most dependent
//   - OnStop: from most dependent to least dependent (reverse order)
//
// Currently, we only log startup and shutdown messages.
// Additional components can register their own hooks by taking fx.Lifecycle
// as a dependency in their provider functions.
func registerLifecycleHooks(params LifecycleParams) {
	params.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			params.Logger.Info("Starting glide application")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			params.Logger.Info("Shutting down glide application")
			return nil
		},
	})
}
