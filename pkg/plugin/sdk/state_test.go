package sdk

import (
	"sync"
	"testing"
)

func TestPluginState_String(t *testing.T) {
	tests := []struct {
		state    PluginState
		expected string
	}{
		{StateUninitialized, "Uninitialized"},
		{StateInitialized, "Initialized"},
		{StateStarted, "Started"},
		{StateStopped, "Stopped"},
		{StateErrored, "Errored"},
		{PluginState(99), "Unknown(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.state.String(); got != tt.expected {
				t.Errorf("PluginState.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsValidTransition(t *testing.T) {
	tests := []struct {
		name  string
		from  PluginState
		to    PluginState
		valid bool
	}{
		// Valid transitions
		{"Uninitialized to Initialized", StateUninitialized, StateInitialized, true},
		{"Uninitialized to Errored", StateUninitialized, StateErrored, true},
		{"Uninitialized to Stopped", StateUninitialized, StateStopped, true},
		{"Initialized to Started", StateInitialized, StateStarted, true},
		{"Initialized to Errored", StateInitialized, StateErrored, true},
		{"Initialized to Stopped", StateInitialized, StateStopped, true},
		{"Started to Stopped", StateStarted, StateStopped, true},
		{"Started to Errored", StateStarted, StateErrored, true},
		{"Errored to Stopped", StateErrored, StateStopped, true},

		// Invalid transitions
		{"Uninitialized to Started", StateUninitialized, StateStarted, false},
		{"Initialized to Uninitialized", StateInitialized, StateUninitialized, false},
		{"Started to Initialized", StateStarted, StateInitialized, false},
		{"Started to Uninitialized", StateStarted, StateUninitialized, false},
		{"Stopped to any state", StateStopped, StateStarted, false},
		{"Stopped to Initialized", StateStopped, StateInitialized, false},
		{"Errored to Started", StateErrored, StateStarted, false},
		{"Errored to Initialized", StateErrored, StateInitialized, false},
		{"Invalid from state", PluginState(99), StateStarted, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidTransition(tt.from, tt.to); got != tt.valid {
				t.Errorf("IsValidTransition(%v, %v) = %v, want %v",
					tt.from, tt.to, got, tt.valid)
			}
		})
	}
}

func TestStateTracker_Get(t *testing.T) {
	st := NewStateTracker("test-plugin")
	if got := st.Get(); got != StateUninitialized {
		t.Errorf("NewStateTracker initial state = %v, want %v", got, StateUninitialized)
	}
}

func TestStateTracker_Set_ValidTransitions(t *testing.T) {
	tests := []struct {
		name        string
		transitions []PluginState
		wantErr     bool
	}{
		{
			name:        "successful lifecycle",
			transitions: []PluginState{StateInitialized, StateStarted, StateStopped},
			wantErr:     false,
		},
		{
			name:        "error during init",
			transitions: []PluginState{StateErrored, StateStopped},
			wantErr:     false,
		},
		{
			name:        "error after start",
			transitions: []PluginState{StateInitialized, StateStarted, StateErrored, StateStopped},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := NewStateTracker("test-plugin")

			for i, targetState := range tt.transitions {
				err := st.Set(targetState)
				if (err != nil) != tt.wantErr {
					t.Errorf("Set() step %d error = %v, wantErr %v", i, err, tt.wantErr)
				}

				if err == nil && st.Get() != targetState {
					t.Errorf("After Set(%v), state = %v, want %v", targetState, st.Get(), targetState)
				}
			}
		})
	}
}

func TestStateTracker_Set_InvalidTransitions(t *testing.T) {
	tests := []struct {
		name         string
		initialState PluginState
		targetState  PluginState
		expectError  bool
	}{
		{"skip init", StateUninitialized, StateStarted, true},
		{"go back from started", StateStarted, StateInitialized, true},
		{"restart after stop", StateStopped, StateStarted, true},
		{"recover from error without stop", StateErrored, StateStarted, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := NewStateTracker("test-plugin")

			// Set initial state using ForceSet to bypass validation
			if tt.initialState != StateUninitialized {
				st.ForceSet(tt.initialState)
			}

			err := st.Set(tt.targetState)
			if (err != nil) != tt.expectError {
				t.Errorf("Set() error = %v, expectError %v", err, tt.expectError)
			}

			if tt.expectError {
				// State should not have changed
				if st.Get() != tt.initialState {
					t.Errorf("After failed Set(), state = %v, want %v", st.Get(), tt.initialState)
				}

				// Check error type
				var stErr *StateTransitionError
				if err == nil {
					t.Error("Expected StateTransitionError, got nil")
				} else {
					var ok bool
					stErr, ok = err.(*StateTransitionError)
					if !ok {
						t.Errorf("Expected StateTransitionError, got %T", err)
					} else {
						if stErr.Plugin != "test-plugin" {
							t.Errorf("StateTransitionError.Plugin = %v, want test-plugin", stErr.Plugin)
						}
						if stErr.CurrentState != tt.initialState {
							t.Errorf("StateTransitionError.CurrentState = %v, want %v", stErr.CurrentState, tt.initialState)
						}
						if stErr.TargetState != tt.targetState {
							t.Errorf("StateTransitionError.TargetState = %v, want %v", stErr.TargetState, tt.targetState)
						}
					}
				}
			}
		})
	}
}

func TestStateTracker_ForceSet(t *testing.T) {
	st := NewStateTracker("test-plugin")

	// Force set to an invalid transition
	st.ForceSet(StateStarted)

	if st.Get() != StateStarted {
		t.Errorf("After ForceSet(StateStarted), state = %v, want %v", st.Get(), StateStarted)
	}

	// Can force set from stopped (normally terminal)
	st.ForceSet(StateStopped)
	st.ForceSet(StateStarted)

	if st.Get() != StateStarted {
		t.Errorf("After ForceSet(StateStarted) from Stopped, state = %v, want %v", st.Get(), StateStarted)
	}
}

func TestStateTracker_IsOperational(t *testing.T) {
	tests := []struct {
		state       PluginState
		operational bool
	}{
		{StateUninitialized, false},
		{StateInitialized, false},
		{StateStarted, true},
		{StateStopped, false},
		{StateErrored, false},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			st := NewStateTracker("test-plugin")
			st.ForceSet(tt.state)

			if got := st.IsOperational(); got != tt.operational {
				t.Errorf("IsOperational() = %v, want %v for state %v", got, tt.operational, tt.state)
			}
		})
	}
}

func TestStateTracker_CanShutdown(t *testing.T) {
	tests := []struct {
		state       PluginState
		canShutdown bool
	}{
		{StateUninitialized, true},
		{StateInitialized, true},
		{StateStarted, true},
		{StateStopped, false},
		{StateErrored, true},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			st := NewStateTracker("test-plugin")
			st.ForceSet(tt.state)

			if got := st.CanShutdown(); got != tt.canShutdown {
				t.Errorf("CanShutdown() = %v, want %v for state %v", got, tt.canShutdown, tt.state)
			}
		})
	}
}

func TestStateTracker_Concurrency(t *testing.T) {
	st := NewStateTracker("test-plugin")
	var wg sync.WaitGroup

	// Try concurrent state transitions
	wg.Add(3)

	go func() {
		defer wg.Done()
		_ = st.Set(StateInitialized)
	}()

	go func() {
		defer wg.Done()
		_ = st.Get()
	}()

	go func() {
		defer wg.Done()
		_ = st.IsOperational()
	}()

	wg.Wait()

	// After concurrent access, state should be consistent
	finalState := st.Get()
	if finalState != StateUninitialized && finalState != StateInitialized {
		t.Errorf("After concurrent access, unexpected state: %v", finalState)
	}
}

func TestStateTransitionError_Error(t *testing.T) {
	err := &StateTransitionError{
		Plugin:       "test-plugin",
		CurrentState: StateUninitialized,
		TargetState:  StateStarted,
		Operation:    "Set",
	}

	expected := "plugin test-plugin: invalid state transition from Uninitialized to Started during Set"
	if got := err.Error(); got != expected {
		t.Errorf("StateTransitionError.Error() = %v, want %v", got, expected)
	}
}
