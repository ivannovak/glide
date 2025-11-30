package sdk

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDefaultLifecycleConfig(t *testing.T) {
	config := DefaultLifecycleConfig()

	if config.InitTimeout != 30*time.Second {
		t.Errorf("InitTimeout = %v, want 30s", config.InitTimeout)
	}
	if config.StartTimeout != 30*time.Second {
		t.Errorf("StartTimeout = %v, want 30s", config.StartTimeout)
	}
	if config.StopTimeout != 10*time.Second {
		t.Errorf("StopTimeout = %v, want 10s", config.StopTimeout)
	}
	if config.HealthCheckTimeout != 5*time.Second {
		t.Errorf("HealthCheckTimeout = %v, want 5s", config.HealthCheckTimeout)
	}
	if config.HealthCheckInterval != 30*time.Second {
		t.Errorf("HealthCheckInterval = %v, want 30s", config.HealthCheckInterval)
	}
	if config.UnhealthyThreshold != 3 {
		t.Errorf("UnhealthyThreshold = %v, want 3", config.UnhealthyThreshold)
	}
}

func TestNewLifecycleManager(t *testing.T) {
	// With config
	config := &LifecycleConfig{InitTimeout: 10 * time.Second}
	lm := NewLifecycleManager(config)

	if lm.config.InitTimeout != 10*time.Second {
		t.Errorf("InitTimeout = %v, want 10s", lm.config.InitTimeout)
	}

	// Without config (should use defaults)
	lm2 := NewLifecycleManager(nil)
	if lm2.config.InitTimeout != 30*time.Second {
		t.Errorf("Default InitTimeout = %v, want 30s", lm2.config.InitTimeout)
	}
}

func TestLifecycleManager_Register(t *testing.T) {
	lm := NewLifecycleManager(nil)
	mock := &mockLifecycle{}

	// Register plugin
	err := lm.Register("test-plugin", mock)
	if err != nil {
		t.Errorf("Register() error = %v", err)
	}

	// Check plugin is registered
	state, err := lm.GetPluginState("test-plugin")
	if err != nil {
		t.Errorf("GetPluginState() error = %v", err)
	}
	if state != StateUninitialized {
		t.Errorf("Initial state = %v, want Uninitialized", state)
	}

	// Try to register same plugin again
	err = lm.Register("test-plugin", mock)
	if err == nil {
		t.Error("Register() duplicate should return error")
	}
}

func TestLifecycleManager_Unregister(t *testing.T) {
	lm := NewLifecycleManager(nil)
	mock := &mockLifecycle{}

	// Unregister non-existent plugin
	err := lm.Unregister("non-existent")
	if err == nil {
		t.Error("Unregister() non-existent should return error")
	}

	// Register and init plugin
	_ = lm.Register("test-plugin", mock)

	// Try to unregister without stopping
	err = lm.Unregister("test-plugin")
	if err == nil {
		t.Error("Unregister() non-stopped plugin should return error")
	}

	// Stop and unregister
	lm.plugins["test-plugin"].State.ForceSet(StateStopped)
	err = lm.Unregister("test-plugin")
	if err != nil {
		t.Errorf("Unregister() stopped plugin error = %v", err)
	}

	// Verify plugin is unregistered
	_, err = lm.GetPluginState("test-plugin")
	if err == nil {
		t.Error("GetPluginState() should fail for unregistered plugin")
	}
}

func TestLifecycleManager_InitPlugin_Success(t *testing.T) {
	lm := NewLifecycleManager(nil)
	mock := &mockLifecycle{}

	_ = lm.Register("test-plugin", mock)

	ctx := context.Background()
	err := lm.InitPlugin(ctx, "test-plugin")
	if err != nil {
		t.Errorf("InitPlugin() error = %v", err)
	}

	if !mock.initCalled {
		t.Error("Init() was not called")
	}

	state, _ := lm.GetPluginState("test-plugin")
	if state != StateInitialized {
		t.Errorf("State = %v, want Initialized", state)
	}
}

func TestLifecycleManager_InitPlugin_Failure(t *testing.T) {
	lm := NewLifecycleManager(nil)
	expectedErr := errors.New("init failed")
	mock := &mockLifecycle{initErr: expectedErr}

	_ = lm.Register("test-plugin", mock)

	ctx := context.Background()
	err := lm.InitPlugin(ctx, "test-plugin")
	if err == nil {
		t.Error("InitPlugin() should return error on Init failure")
	}

	state, _ := lm.GetPluginState("test-plugin")
	if state != StateErrored {
		t.Errorf("State = %v, want Errored", state)
	}
}

func TestLifecycleManager_InitPlugin_Timeout(t *testing.T) {
	config := &LifecycleConfig{
		InitTimeout: 100 * time.Millisecond,
	}
	lm := NewLifecycleManager(config)

	// Mock that blocks in Init
	mock := &slowLifecycle{initDelay: 1 * time.Second}

	_ = lm.Register("test-plugin", mock)

	ctx := context.Background()
	err := lm.InitPlugin(ctx, "test-plugin")
	if err == nil {
		t.Error("InitPlugin() should timeout")
	}

	state, _ := lm.GetPluginState("test-plugin")
	if state != StateErrored {
		t.Errorf("State = %v, want Errored", state)
	}
}

func TestLifecycleManager_StartPlugin_Success(t *testing.T) {
	lm := NewLifecycleManager(nil)
	mock := &mockLifecycle{}

	_ = lm.Register("test-plugin", mock)
	_ = lm.InitPlugin(context.Background(), "test-plugin")

	ctx := context.Background()
	err := lm.StartPlugin(ctx, "test-plugin")
	if err != nil {
		t.Errorf("StartPlugin() error = %v", err)
	}

	if !mock.startCalled {
		t.Error("Start() was not called")
	}

	state, _ := lm.GetPluginState("test-plugin")
	if state != StateStarted {
		t.Errorf("State = %v, want Started", state)
	}
}

func TestLifecycleManager_StartPlugin_Failure(t *testing.T) {
	lm := NewLifecycleManager(nil)
	expectedErr := errors.New("start failed")
	mock := &mockLifecycle{startErr: expectedErr}

	_ = lm.Register("test-plugin", mock)
	_ = lm.InitPlugin(context.Background(), "test-plugin")

	ctx := context.Background()
	err := lm.StartPlugin(ctx, "test-plugin")
	if err == nil {
		t.Error("StartPlugin() should return error on Start failure")
	}

	state, _ := lm.GetPluginState("test-plugin")
	if state != StateErrored {
		t.Errorf("State = %v, want Errored", state)
	}
}

func TestLifecycleManager_StopPlugin_Success(t *testing.T) {
	lm := NewLifecycleManager(nil)
	mock := &mockLifecycle{}

	_ = lm.Register("test-plugin", mock)
	_ = lm.InitPlugin(context.Background(), "test-plugin")
	_ = lm.StartPlugin(context.Background(), "test-plugin")

	ctx := context.Background()
	err := lm.StopPlugin(ctx, "test-plugin")
	if err != nil {
		t.Errorf("StopPlugin() error = %v", err)
	}

	if !mock.stopCalled {
		t.Error("Stop() was not called")
	}

	state, _ := lm.GetPluginState("test-plugin")
	if state != StateStopped {
		t.Errorf("State = %v, want Stopped", state)
	}
}

func TestLifecycleManager_StopPlugin_AlreadyStopped(t *testing.T) {
	lm := NewLifecycleManager(nil)
	mock := &mockLifecycle{}

	_ = lm.Register("test-plugin", mock)
	lm.plugins["test-plugin"].State.ForceSet(StateStopped)

	ctx := context.Background()
	err := lm.StopPlugin(ctx, "test-plugin")
	if err != nil {
		t.Errorf("StopPlugin() already stopped should not error, got %v", err)
	}

	if mock.stopCalled {
		t.Error("Stop() should not be called on already stopped plugin")
	}
}

func TestLifecycleManager_StopPlugin_Failure(t *testing.T) {
	lm := NewLifecycleManager(nil)
	expectedErr := errors.New("stop failed")
	mock := &mockLifecycle{stopErr: expectedErr}

	_ = lm.Register("test-plugin", mock)
	lm.plugins["test-plugin"].State.ForceSet(StateStarted)

	ctx := context.Background()
	err := lm.StopPlugin(ctx, "test-plugin")
	if err == nil {
		t.Error("StopPlugin() should return error on Stop failure")
	}

	// Even on error, state should transition to Stopped
	state, _ := lm.GetPluginState("test-plugin")
	if state != StateStopped {
		t.Errorf("State = %v, want Stopped (even on error)", state)
	}
}

func TestLifecycleManager_InitAll(t *testing.T) {
	lm := NewLifecycleManager(nil)

	mock1 := &mockLifecycle{}
	mock2 := &mockLifecycle{}

	_ = lm.Register("plugin1", mock1)
	_ = lm.Register("plugin2", mock2)

	ctx := context.Background()
	err := lm.InitAll(ctx)
	if err != nil {
		t.Errorf("InitAll() error = %v", err)
	}

	if !mock1.initCalled || !mock2.initCalled {
		t.Error("Init() should be called on all plugins")
	}
}

func TestLifecycleManager_InitAll_Failure(t *testing.T) {
	lm := NewLifecycleManager(nil)

	mock1 := &mockLifecycle{}
	mock2 := &mockLifecycle{initErr: errors.New("init failed")}

	_ = lm.Register("plugin1", mock1)
	_ = lm.Register("plugin2", mock2)

	ctx := context.Background()
	err := lm.InitAll(ctx)
	if err == nil {
		t.Error("InitAll() should return error if any plugin fails")
	}
}

func TestLifecycleManager_StartAll(t *testing.T) {
	config := &LifecycleConfig{
		HealthCheckInterval: 0, // Disable periodic health checks for test
	}
	lm := NewLifecycleManager(config)

	mock1 := &mockLifecycle{}
	mock2 := &mockLifecycle{}

	_ = lm.Register("plugin1", mock1)
	_ = lm.Register("plugin2", mock2)

	_ = lm.InitAll(context.Background())

	ctx := context.Background()
	err := lm.StartAll(ctx)
	if err != nil {
		t.Errorf("StartAll() error = %v", err)
	}

	if !mock1.startCalled || !mock2.startCalled {
		t.Error("Start() should be called on all plugins")
	}
}

func TestLifecycleManager_StopAll(t *testing.T) {
	lm := NewLifecycleManager(nil)

	mock1 := &mockLifecycle{}
	mock2 := &mockLifecycle{}

	_ = lm.Register("plugin1", mock1)
	_ = lm.Register("plugin2", mock2)

	_ = lm.InitAll(context.Background())
	_ = lm.StartAll(context.Background())

	ctx := context.Background()
	err := lm.StopAll(ctx)
	if err != nil {
		t.Errorf("StopAll() error = %v", err)
	}

	if !mock1.stopCalled || !mock2.stopCalled {
		t.Error("Stop() should be called on all plugins")
	}
}

func TestLifecycleManager_HealthCheckPlugin_Success(t *testing.T) {
	lm := NewLifecycleManager(nil)
	mock := &mockLifecycle{}

	_ = lm.Register("test-plugin", mock)
	_ = lm.InitPlugin(context.Background(), "test-plugin")
	_ = lm.StartPlugin(context.Background(), "test-plugin")

	err := lm.HealthCheckPlugin("test-plugin")
	if err != nil {
		t.Errorf("HealthCheckPlugin() error = %v", err)
	}

	if !mock.healthCalled {
		t.Error("HealthCheck() was not called")
	}

	lastCheck, healthErr := lm.GetPluginHealth("test-plugin")
	if lastCheck.IsZero() {
		t.Error("LastHealthCheck should be set")
	}
	if healthErr != nil {
		t.Errorf("HealthCheckErr = %v, want nil", healthErr)
	}
}

func TestLifecycleManager_HealthCheckPlugin_Failure(t *testing.T) {
	lm := NewLifecycleManager(nil)
	expectedErr := errors.New("unhealthy")
	mock := &mockLifecycle{healthErr: expectedErr}

	_ = lm.Register("test-plugin", mock)
	_ = lm.InitPlugin(context.Background(), "test-plugin")
	_ = lm.StartPlugin(context.Background(), "test-plugin")

	err := lm.HealthCheckPlugin("test-plugin")
	if err == nil {
		t.Error("HealthCheckPlugin() should return error on unhealthy plugin")
	}

	_, healthErr := lm.GetPluginHealth("test-plugin")
	if healthErr == nil {
		t.Error("GetPluginHealth() should return health check error")
	}
}

func TestLifecycleManager_HealthCheckPlugin_Timeout(t *testing.T) {
	config := &LifecycleConfig{
		HealthCheckTimeout: 100 * time.Millisecond,
	}
	lm := NewLifecycleManager(config)

	// Mock that blocks in HealthCheck
	mock := &slowLifecycle{healthDelay: 1 * time.Second}

	_ = lm.Register("test-plugin", mock)
	lm.plugins["test-plugin"].State.ForceSet(StateStarted)

	err := lm.HealthCheckPlugin("test-plugin")
	if err == nil {
		t.Error("HealthCheckPlugin() should timeout")
	}
}

func TestLifecycleManager_HealthCheckPlugin_NotOperational(t *testing.T) {
	lm := NewLifecycleManager(nil)
	mock := &mockLifecycle{}

	_ = lm.Register("test-plugin", mock)
	// Don't init/start - keep in Uninitialized state

	err := lm.HealthCheckPlugin("test-plugin")
	if err != nil {
		t.Errorf("HealthCheckPlugin() non-operational should not error, got %v", err)
	}

	if mock.healthCalled {
		t.Error("HealthCheck() should not be called on non-operational plugin")
	}
}

func TestLifecycleManager_ListPlugins(t *testing.T) {
	lm := NewLifecycleManager(nil)

	_ = lm.Register("plugin1", &mockLifecycle{})
	_ = lm.Register("plugin2", &mockLifecycle{})

	plugins := lm.ListPlugins()
	if len(plugins) != 2 {
		t.Errorf("ListPlugins() returned %d plugins, want 2", len(plugins))
	}

	// Check both plugins are in the list
	found := make(map[string]bool)
	for _, name := range plugins {
		found[name] = true
	}

	if !found["plugin1"] || !found["plugin2"] {
		t.Error("ListPlugins() missing expected plugins")
	}
}

// slowLifecycle is a mock that can simulate delays
type slowLifecycle struct {
	initDelay   time.Duration
	startDelay  time.Duration
	stopDelay   time.Duration
	healthDelay time.Duration
}

func (s *slowLifecycle) Init(ctx context.Context) error {
	if s.initDelay > 0 {
		select {
		case <-time.After(s.initDelay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (s *slowLifecycle) Start(ctx context.Context) error {
	if s.startDelay > 0 {
		select {
		case <-time.After(s.startDelay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (s *slowLifecycle) Stop(ctx context.Context) error {
	if s.stopDelay > 0 {
		select {
		case <-time.After(s.stopDelay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (s *slowLifecycle) HealthCheck() error {
	if s.healthDelay > 0 {
		time.Sleep(s.healthDelay)
	}
	return nil
}

func TestLifecycleManager_PeriodicHealthChecks(t *testing.T) {
	config := &LifecycleConfig{
		HealthCheckInterval: 100 * time.Millisecond,
		HealthCheckTimeout:  50 * time.Millisecond,
	}
	lm := NewLifecycleManager(config)

	mock := &mockLifecycle{}

	_ = lm.Register("test-plugin", mock)
	_ = lm.InitAll(context.Background())

	// StartAll starts the periodic health checking
	_ = lm.StartAll(context.Background())

	// Wait for at least two health check cycles to ensure it runs
	time.Sleep(250 * time.Millisecond)

	// Stop health checking
	err := lm.StopAll(context.Background())
	if err != nil {
		t.Errorf("StopAll() error = %v", err)
	}

	if !mock.healthCalled {
		t.Error("Periodic health check should have been called")
	}

	// Verify last health check was recorded
	lastCheck, _ := lm.GetPluginHealth("test-plugin")
	if lastCheck.IsZero() {
		t.Error("LastHealthCheck should be set by periodic checks")
	}
}

func TestLifecycleManager_InitPlugin_NotRegistered(t *testing.T) {
	lm := NewLifecycleManager(nil)

	err := lm.InitPlugin(context.Background(), "non-existent")
	if err == nil {
		t.Error("InitPlugin() should error for non-registered plugin")
	}
}

func TestLifecycleManager_StartPlugin_NotRegistered(t *testing.T) {
	lm := NewLifecycleManager(nil)

	err := lm.StartPlugin(context.Background(), "non-existent")
	if err == nil {
		t.Error("StartPlugin() should error for non-registered plugin")
	}
}

func TestLifecycleManager_StopPlugin_NotRegistered(t *testing.T) {
	lm := NewLifecycleManager(nil)

	err := lm.StopPlugin(context.Background(), "non-existent")
	if err == nil {
		t.Error("StopPlugin() should error for non-registered plugin")
	}
}

func TestLifecycleManager_HealthCheckPlugin_NotRegistered(t *testing.T) {
	lm := NewLifecycleManager(nil)

	err := lm.HealthCheckPlugin("non-existent")
	if err == nil {
		t.Error("HealthCheckPlugin() should error for non-registered plugin")
	}
}

func TestLifecycleManager_GetPluginHealth_NotRegistered(t *testing.T) {
	lm := NewLifecycleManager(nil)

	_, err := lm.GetPluginHealth("non-existent")
	if err == nil {
		t.Error("GetPluginHealth() should error for non-registered plugin")
	}
}
