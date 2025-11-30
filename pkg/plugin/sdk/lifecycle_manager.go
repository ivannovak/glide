package sdk

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// LifecycleManager orchestrates the lifecycle of multiple plugins
type LifecycleManager struct {
	mu                sync.RWMutex
	plugins           map[string]*ManagedPlugin
	config            *LifecycleConfig
	healthCheckTicker *time.Ticker
	shutdownChan      chan struct{}
	wg                sync.WaitGroup
}

// ManagedPlugin wraps a plugin with lifecycle management
type ManagedPlugin struct {
	Name            string
	Plugin          Lifecycle
	State           *StateTracker
	LastHealthCheck time.Time
	HealthCheckErr  error
}

// LifecycleConfig configures the lifecycle manager
type LifecycleConfig struct {
	// InitTimeout is the maximum time allowed for Init operations
	InitTimeout time.Duration

	// StartTimeout is the maximum time allowed for Start operations
	StartTimeout time.Duration

	// StopTimeout is the maximum time allowed for Stop operations
	StopTimeout time.Duration

	// HealthCheckTimeout is the maximum time allowed for HealthCheck operations
	HealthCheckTimeout time.Duration

	// HealthCheckInterval is how often to run health checks (0 disables periodic checks)
	HealthCheckInterval time.Duration

	// UnhealthyThreshold is the number of consecutive failed health checks before marking unhealthy
	UnhealthyThreshold int
}

// DefaultLifecycleConfig returns sensible default configuration
func DefaultLifecycleConfig() *LifecycleConfig {
	return &LifecycleConfig{
		InitTimeout:         30 * time.Second,
		StartTimeout:        30 * time.Second,
		StopTimeout:         10 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		HealthCheckInterval: 30 * time.Second,
		UnhealthyThreshold:  3,
	}
}

// NewLifecycleManager creates a new lifecycle manager
func NewLifecycleManager(config *LifecycleConfig) *LifecycleManager {
	if config == nil {
		config = DefaultLifecycleConfig()
	}

	return &LifecycleManager{
		plugins:      make(map[string]*ManagedPlugin),
		config:       config,
		shutdownChan: make(chan struct{}),
	}
}

// Register adds a plugin to lifecycle management
func (lm *LifecycleManager) Register(name string, plugin Lifecycle) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if _, exists := lm.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}

	lm.plugins[name] = &ManagedPlugin{
		Name:   name,
		Plugin: plugin,
		State:  NewStateTracker(name),
	}

	return nil
}

// Unregister removes a plugin from lifecycle management
// Plugin must be stopped before unregistering
func (lm *LifecycleManager) Unregister(name string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	managed, exists := lm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not registered", name)
	}

	if managed.State.Get() != StateStopped {
		return fmt.Errorf("plugin %s must be stopped before unregistering (current state: %s)",
			name, managed.State.Get())
	}

	delete(lm.plugins, name)
	return nil
}

// InitPlugin initializes a specific plugin
func (lm *LifecycleManager) InitPlugin(ctx context.Context, name string) error {
	lm.mu.RLock()
	managed, exists := lm.plugins[name]
	lm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("plugin %s not registered", name)
	}

	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, lm.config.InitTimeout)
	defer cancel()

	// Call Init
	if err := managed.Plugin.Init(ctx); err != nil {
		// Transition to errored state
		_ = managed.State.Set(StateErrored)
		return NewLifecycleError("Init", name, "initialization failed", err)
	}

	// Transition to initialized state
	if err := managed.State.Set(StateInitialized); err != nil {
		return err
	}

	return nil
}

// StartPlugin starts a specific plugin
func (lm *LifecycleManager) StartPlugin(ctx context.Context, name string) error {
	lm.mu.RLock()
	managed, exists := lm.plugins[name]
	lm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("plugin %s not registered", name)
	}

	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, lm.config.StartTimeout)
	defer cancel()

	// Call Start
	if err := managed.Plugin.Start(ctx); err != nil {
		// Transition to errored state
		_ = managed.State.Set(StateErrored)
		return NewLifecycleError("Start", name, "startup failed", err)
	}

	// Transition to started state
	if err := managed.State.Set(StateStarted); err != nil {
		return err
	}

	return nil
}

// StopPlugin stops a specific plugin
func (lm *LifecycleManager) StopPlugin(ctx context.Context, name string) error {
	lm.mu.RLock()
	managed, exists := lm.plugins[name]
	lm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("plugin %s not registered", name)
	}

	// Check if can shutdown
	if !managed.State.CanShutdown() {
		return nil // Already stopped
	}

	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, lm.config.StopTimeout)
	defer cancel()

	// Call Stop
	if err := managed.Plugin.Stop(ctx); err != nil {
		// Still transition to stopped state even on error
		_ = managed.State.Set(StateStopped)
		return NewLifecycleError("Stop", name, "shutdown failed", err)
	}

	// Transition to stopped state
	if err := managed.State.Set(StateStopped); err != nil {
		return err
	}

	return nil
}

// InitAll initializes all registered plugins
func (lm *LifecycleManager) InitAll(ctx context.Context) error {
	lm.mu.RLock()
	names := make([]string, 0, len(lm.plugins))
	for name := range lm.plugins {
		names = append(names, name)
	}
	lm.mu.RUnlock()

	// Initialize plugins sequentially
	for _, name := range names {
		if err := lm.InitPlugin(ctx, name); err != nil {
			return fmt.Errorf("failed to initialize plugins: %w", err)
		}
	}

	return nil
}

// StartAll starts all initialized plugins
func (lm *LifecycleManager) StartAll(ctx context.Context) error {
	lm.mu.RLock()
	names := make([]string, 0, len(lm.plugins))
	for name := range lm.plugins {
		names = append(names, name)
	}
	lm.mu.RUnlock()

	// Start plugins sequentially
	for _, name := range names {
		if err := lm.StartPlugin(ctx, name); err != nil {
			return fmt.Errorf("failed to start plugins: %w", err)
		}
	}

	// Start health checking if configured
	if lm.config.HealthCheckInterval > 0 {
		lm.startHealthChecking()
	}

	return nil
}

// StopAll stops all plugins in reverse order
func (lm *LifecycleManager) StopAll(ctx context.Context) error {
	// Stop health checking
	lm.stopHealthChecking()

	lm.mu.RLock()
	names := make([]string, 0, len(lm.plugins))
	for name := range lm.plugins {
		names = append(names, name)
	}
	lm.mu.RUnlock()

	// Stop plugins in reverse order (LIFO)
	var lastErr error
	for i := len(names) - 1; i >= 0; i-- {
		if err := lm.StopPlugin(ctx, names[i]); err != nil {
			// Continue stopping other plugins, but remember the error
			lastErr = err
		}
	}

	return lastErr
}

// HealthCheckPlugin performs a health check on a specific plugin
func (lm *LifecycleManager) HealthCheckPlugin(name string) error {
	lm.mu.RLock()
	managed, exists := lm.plugins[name]
	lm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("plugin %s not registered", name)
	}

	// Only health check operational plugins
	if !managed.State.IsOperational() {
		return nil
	}

	// Create timeout context
	ctx, cancel := context.WithTimeout(context.Background(), lm.config.HealthCheckTimeout)
	defer cancel()

	// Run health check in goroutine to enforce timeout
	errChan := make(chan error, 1)
	go func() {
		errChan <- managed.Plugin.HealthCheck()
	}()

	select {
	case err := <-errChan:
		lm.mu.Lock()
		managed.LastHealthCheck = time.Now()
		managed.HealthCheckErr = err
		lm.mu.Unlock()

		if err != nil {
			return NewLifecycleError("HealthCheck", name, "health check failed", err)
		}
		return nil

	case <-ctx.Done():
		lm.mu.Lock()
		managed.LastHealthCheck = time.Now()
		managed.HealthCheckErr = ctx.Err()
		lm.mu.Unlock()

		return NewLifecycleError("HealthCheck", name, "health check timeout", ctx.Err())
	}
}

// startHealthChecking begins periodic health checks
func (lm *LifecycleManager) startHealthChecking() {
	lm.healthCheckTicker = time.NewTicker(lm.config.HealthCheckInterval)

	lm.wg.Add(1)
	go func() {
		defer lm.wg.Done()

		for {
			select {
			case <-lm.healthCheckTicker.C:
				lm.runHealthChecks()

			case <-lm.shutdownChan:
				return
			}
		}
	}()
}

// stopHealthChecking stops periodic health checks
func (lm *LifecycleManager) stopHealthChecking() {
	if lm.healthCheckTicker != nil {
		lm.healthCheckTicker.Stop()
		close(lm.shutdownChan)
		lm.wg.Wait()
	}
}

// runHealthChecks runs health checks on all plugins
func (lm *LifecycleManager) runHealthChecks() {
	lm.mu.RLock()
	names := make([]string, 0, len(lm.plugins))
	for name := range lm.plugins {
		names = append(names, name)
	}
	lm.mu.RUnlock()

	for _, name := range names {
		_ = lm.HealthCheckPlugin(name)
	}
}

// GetPluginState returns the current state of a plugin
func (lm *LifecycleManager) GetPluginState(name string) (PluginState, error) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	managed, exists := lm.plugins[name]
	if !exists {
		return StateUninitialized, fmt.Errorf("plugin %s not registered", name)
	}

	return managed.State.Get(), nil
}

// GetPluginHealth returns the last health check result for a plugin
func (lm *LifecycleManager) GetPluginHealth(name string) (lastCheck time.Time, err error) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	managed, exists := lm.plugins[name]
	if !exists {
		return time.Time{}, fmt.Errorf("plugin %s not registered", name)
	}

	return managed.LastHealthCheck, managed.HealthCheckErr
}

// ListPlugins returns all registered plugin names
func (lm *LifecycleManager) ListPlugins() []string {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	names := make([]string, 0, len(lm.plugins))
	for name := range lm.plugins {
		names = append(names, name)
	}

	return names
}
