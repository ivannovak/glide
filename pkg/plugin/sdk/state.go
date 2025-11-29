package sdk

import (
	"fmt"
	"sync"
)

// PluginState represents the current lifecycle state of a plugin
type PluginState int

const (
	// StateUninitialized means the plugin process has been started but Init has not been called
	StateUninitialized PluginState = iota

	// StateInitialized means Init has completed successfully
	StateInitialized

	// StateStarted means Start has completed successfully and the plugin is operational
	StateStarted

	// StateStopped means Stop has been called and the plugin is no longer operational
	StateStopped

	// StateErrored means the plugin encountered an error during lifecycle operations
	StateErrored
)

// String returns the string representation of a PluginState
func (s PluginState) String() string {
	switch s {
	case StateUninitialized:
		return "Uninitialized"
	case StateInitialized:
		return "Initialized"
	case StateStarted:
		return "Started"
	case StateStopped:
		return "Stopped"
	case StateErrored:
		return "Errored"
	default:
		return fmt.Sprintf("Unknown(%d)", s)
	}
}

// stateTransitions defines the valid state transitions
// Maps current state to allowed next states
var stateTransitions = map[PluginState][]PluginState{
	StateUninitialized: {StateInitialized, StateErrored, StateStopped},
	StateInitialized:   {StateStarted, StateErrored, StateStopped},
	StateStarted:       {StateStopped, StateErrored},
	StateStopped:       {},             // Terminal state (can't transition out)
	StateErrored:       {StateStopped}, // Can attempt graceful shutdown from error state
}

// IsValidTransition checks if a state transition is allowed
func IsValidTransition(from, to PluginState) bool {
	allowedStates, ok := stateTransitions[from]
	if !ok {
		return false
	}

	for _, allowed := range allowedStates {
		if allowed == to {
			return true
		}
	}
	return false
}

// StateTracker tracks the lifecycle state of a plugin with thread safety
type StateTracker struct {
	mu    sync.RWMutex
	state PluginState
	name  string
}

// NewStateTracker creates a new state tracker for a plugin
func NewStateTracker(name string) *StateTracker {
	return &StateTracker{
		state: StateUninitialized,
		name:  name,
	}
}

// Get returns the current state
func (st *StateTracker) Get() PluginState {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return st.state
}

// Set attempts to set a new state
// Returns an error if the transition is invalid
func (st *StateTracker) Set(newState PluginState) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	if !IsValidTransition(st.state, newState) {
		return &StateTransitionError{
			Plugin:       st.name,
			CurrentState: st.state,
			TargetState:  newState,
			Operation:    "Set",
		}
	}

	st.state = newState
	return nil
}

// ForceSet sets the state without validation
// Use only for error recovery scenarios
func (st *StateTracker) ForceSet(newState PluginState) {
	st.mu.Lock()
	defer st.mu.Unlock()
	st.state = newState
}

// IsOperational returns true if the plugin is in a state where it can handle commands
func (st *StateTracker) IsOperational() bool {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return st.state == StateStarted
}

// CanShutdown returns true if the plugin is in a state where it can be shut down
func (st *StateTracker) CanShutdown() bool {
	st.mu.RLock()
	defer st.mu.RUnlock()
	// Can attempt shutdown from any state except already stopped
	return st.state != StateStopped
}
