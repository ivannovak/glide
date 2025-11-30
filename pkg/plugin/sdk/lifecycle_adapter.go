package sdk

import (
	"context"
)

// lifecycleAdapter adapts a LoadedPlugin to the Lifecycle interface
// This allows the LifecycleManager to manage plugin processes
type lifecycleAdapter struct {
	loaded *LoadedPlugin
}

// newLifecycleAdapter creates a new lifecycle adapter for a loaded plugin
func newLifecycleAdapter(loaded *LoadedPlugin) Lifecycle {
	return &lifecycleAdapter{
		loaded: loaded,
	}
}

// Init prepares the plugin (currently a no-op since plugin is already loaded)
func (a *lifecycleAdapter) Init(ctx context.Context) error {
	// Plugin process is already started and metadata fetched during Load
	// This could be extended in the future for SDK v2 plugins
	return nil
}

// Start makes the plugin operational (currently a no-op)
func (a *lifecycleAdapter) Start(ctx context.Context) error {
	// Plugin is operational immediately after loading in SDK v1
	// This could be extended in the future for SDK v2 plugins with actual Start methods
	return nil
}

// Stop gracefully shuts down the plugin
func (a *lifecycleAdapter) Stop(ctx context.Context) error {
	// For now, use Kill() since v1 plugins don't have a graceful shutdown protocol
	// TODO: In SDK v2, implement proper graceful shutdown
	if a.loaded.Client != nil {
		a.loaded.Client.Kill()
	}
	return nil
}

// HealthCheck verifies the plugin is responsive
func (a *lifecycleAdapter) HealthCheck() error {
	// Check if the client is still alive by pinging it
	// If the plugin process has died, this will fail
	if a.loaded.Client == nil {
		return NewLifecycleError("HealthCheck", a.loaded.Name, "plugin client is nil", nil)
	}

	if a.loaded.Client.Exited() {
		return NewLifecycleError("HealthCheck", a.loaded.Name, "plugin process has exited", nil)
	}

	// Plugin is alive - could be extended with actual RPC health check in v2
	return nil
}
