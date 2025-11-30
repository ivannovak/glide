// Package sdk provides lifecycle management for plugins
package sdk

import (
	"context"
	"fmt"
)

// Lifecycle defines the lifecycle interface for plugins
// Plugins implementing this interface can have their initialization,
// startup, shutdown, and health monitored by the lifecycle manager.
type Lifecycle interface {
	// Init prepares the plugin for operation
	// Called once after plugin process starts
	Init(ctx context.Context) error

	// Start begins plugin operation
	// Called after successful Init
	Start(ctx context.Context) error

	// Stop gracefully shuts down the plugin
	// Called during application shutdown or plugin unload
	Stop(ctx context.Context) error

	// HealthCheck reports the health status of the plugin
	// Returns nil if healthy, error describing the issue otherwise
	HealthCheck() error
}

// LifecycleError represents an error during lifecycle operations
type LifecycleError struct {
	Phase   string // Init, Start, Stop, HealthCheck
	Plugin  string // Plugin name
	Cause   error  // Underlying error
	Message string // Additional context
}

func (e *LifecycleError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("plugin %s lifecycle error in %s: %s: %v",
			e.Plugin, e.Phase, e.Message, e.Cause)
	}
	return fmt.Sprintf("plugin %s lifecycle error in %s: %v",
		e.Plugin, e.Phase, e.Cause)
}

func (e *LifecycleError) Unwrap() error {
	return e.Cause
}

// NewLifecycleError creates a new lifecycle error
func NewLifecycleError(phase, plugin, message string, cause error) *LifecycleError {
	return &LifecycleError{
		Phase:   phase,
		Plugin:  plugin,
		Message: message,
		Cause:   cause,
	}
}

// StateTransitionError indicates an invalid state transition was attempted
type StateTransitionError struct {
	Plugin       string
	CurrentState PluginState
	TargetState  PluginState
	Operation    string
}

func (e *StateTransitionError) Error() string {
	return fmt.Sprintf("plugin %s: invalid state transition from %s to %s during %s",
		e.Plugin, e.CurrentState, e.TargetState, e.Operation)
}
