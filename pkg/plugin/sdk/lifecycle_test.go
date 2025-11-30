package sdk

import (
	"context"
	"errors"
	"testing"
)

func TestLifecycleError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *LifecycleError
		expected string
	}{
		{
			name: "with message",
			err: &LifecycleError{
				Phase:   "Init",
				Plugin:  "test-plugin",
				Message: "initialization failed",
				Cause:   errors.New("connection timeout"),
			},
			expected: "plugin test-plugin lifecycle error in Init: initialization failed: connection timeout",
		},
		{
			name: "without message",
			err: &LifecycleError{
				Phase:  "Start",
				Plugin: "test-plugin",
				Cause:  errors.New("port already in use"),
			},
			expected: "plugin test-plugin lifecycle error in Start: port already in use",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("LifecycleError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestLifecycleError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &LifecycleError{
		Phase:  "Init",
		Plugin: "test-plugin",
		Cause:  cause,
	}

	if unwrapped := err.Unwrap(); unwrapped != cause {
		t.Errorf("LifecycleError.Unwrap() = %v, want %v", unwrapped, cause)
	}
}

func TestNewLifecycleError(t *testing.T) {
	cause := errors.New("test error")
	err := NewLifecycleError("Stop", "my-plugin", "shutdown timeout", cause)

	if err.Phase != "Stop" {
		t.Errorf("Phase = %v, want Stop", err.Phase)
	}
	if err.Plugin != "my-plugin" {
		t.Errorf("Plugin = %v, want my-plugin", err.Plugin)
	}
	if err.Message != "shutdown timeout" {
		t.Errorf("Message = %v, want shutdown timeout", err.Message)
	}
	if err.Cause != cause {
		t.Errorf("Cause = %v, want %v", err.Cause, cause)
	}
}

// mockLifecycle is a test implementation of the Lifecycle interface
type mockLifecycle struct {
	initErr   error
	startErr  error
	stopErr   error
	healthErr error

	initCalled   bool
	startCalled  bool
	stopCalled   bool
	healthCalled bool
}

func (m *mockLifecycle) Init(ctx context.Context) error {
	m.initCalled = true
	return m.initErr
}

func (m *mockLifecycle) Start(ctx context.Context) error {
	m.startCalled = true
	return m.startErr
}

func (m *mockLifecycle) Stop(ctx context.Context) error {
	m.stopCalled = true
	return m.stopErr
}

func (m *mockLifecycle) HealthCheck() error {
	m.healthCalled = true
	return m.healthErr
}

func TestMockLifecycle_SuccessfulLifecycle(t *testing.T) {
	mock := &mockLifecycle{}
	ctx := context.Background()

	// Init
	if err := mock.Init(ctx); err != nil {
		t.Errorf("Init() error = %v", err)
	}
	if !mock.initCalled {
		t.Error("Init() was not called")
	}

	// Start
	if err := mock.Start(ctx); err != nil {
		t.Errorf("Start() error = %v", err)
	}
	if !mock.startCalled {
		t.Error("Start() was not called")
	}

	// HealthCheck
	if err := mock.HealthCheck(); err != nil {
		t.Errorf("HealthCheck() error = %v", err)
	}
	if !mock.healthCalled {
		t.Error("HealthCheck() was not called")
	}

	// Stop
	if err := mock.Stop(ctx); err != nil {
		t.Errorf("Stop() error = %v", err)
	}
	if !mock.stopCalled {
		t.Error("Stop() was not called")
	}
}

func TestMockLifecycle_InitError(t *testing.T) {
	expectedErr := errors.New("init failed")
	mock := &mockLifecycle{initErr: expectedErr}
	ctx := context.Background()

	err := mock.Init(ctx)
	if err != expectedErr {
		t.Errorf("Init() error = %v, want %v", err, expectedErr)
	}
	if !mock.initCalled {
		t.Error("Init() was not called")
	}
}

func TestMockLifecycle_StartError(t *testing.T) {
	expectedErr := errors.New("start failed")
	mock := &mockLifecycle{startErr: expectedErr}
	ctx := context.Background()

	err := mock.Start(ctx)
	if err != expectedErr {
		t.Errorf("Start() error = %v, want %v", err, expectedErr)
	}
	if !mock.startCalled {
		t.Error("Start() was not called")
	}
}

func TestMockLifecycle_StopError(t *testing.T) {
	expectedErr := errors.New("stop failed")
	mock := &mockLifecycle{stopErr: expectedErr}
	ctx := context.Background()

	err := mock.Stop(ctx)
	if err != expectedErr {
		t.Errorf("Stop() error = %v, want %v", err, expectedErr)
	}
	if !mock.stopCalled {
		t.Error("Stop() was not called")
	}
}

func TestMockLifecycle_HealthCheckError(t *testing.T) {
	expectedErr := errors.New("unhealthy")
	mock := &mockLifecycle{healthErr: expectedErr}

	err := mock.HealthCheck()
	if err != expectedErr {
		t.Errorf("HealthCheck() error = %v, want %v", err, expectedErr)
	}
	if !mock.healthCalled {
		t.Error("HealthCheck() was not called")
	}
}

func TestMockLifecycle_ContextCancellation(t *testing.T) {
	mock := &mockLifecycle{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Lifecycle methods should still be callable with cancelled context
	// Implementation should check ctx.Err() and handle appropriately
	_ = mock.Init(ctx)
	_ = mock.Start(ctx)
	_ = mock.Stop(ctx)

	if !mock.initCalled || !mock.startCalled || !mock.stopCalled {
		t.Error("Lifecycle methods should be called even with cancelled context")
	}
}

func TestNewLifecycleError_Formatting(t *testing.T) {
	// Test NewLifecycleError and its Error() method
	cause := errors.New("test error")
	err := NewLifecycleError("Stop", "my-plugin", "shutdown timeout", cause)

	expectedMsg := "plugin my-plugin lifecycle error in Stop: shutdown timeout: test error"
	if got := err.Error(); got != expectedMsg {
		t.Errorf("Error() = %v, want %v", got, expectedMsg)
	}
}

func TestStateTransitionError_Formatting(t *testing.T) {
	err := &StateTransitionError{
		Plugin:       "test-plugin",
		CurrentState: StateInitialized,
		TargetState:  StateUninitialized,
		Operation:    "Set",
	}

	expectedMsg := "plugin test-plugin: invalid state transition from Initialized to Uninitialized during Set"
	if got := err.Error(); got != expectedMsg {
		t.Errorf("Error() = %v, want %v", got, expectedMsg)
	}
}
