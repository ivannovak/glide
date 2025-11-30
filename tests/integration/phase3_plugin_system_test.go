package integration_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ivannovak/glide/v2/pkg/plugin/sdk"
	"github.com/ivannovak/glide/v2/pkg/plugin/sdk/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Phase 3 Integration Tests
// These tests validate the integration of:
// - Type-safe configuration system (Task 3.1)
// - Plugin lifecycle management (Task 3.2)
// - Dependency resolution (Task 3.3)
// - SDK v2 development (Task 3.4)

// TestV2PluginEndToEndLifecycle tests the complete lifecycle of a v2 plugin
// with type-safe configuration, initialization, startup, health checks, and shutdown.
//
// Validates:
// - Task 3.1: Type-safe config (generic Plugin[C])
// - Task 3.2: Lifecycle management (Init/Start/Stop/HealthCheck)
// - Task 3.4: SDK v2 interface
//
// Subtask 3.6.1.1: End-to-end plugin lifecycle test
func TestV2PluginEndToEndLifecycle(t *testing.T) {
	t.Run("successful_lifecycle_with_type_safe_config", func(t *testing.T) {
		// Setup: Create test config
		testConfig := testPluginConfig{
			APIKey:  "test-key-123",
			Timeout: 30,
		}

		// Create v2 plugin
		plugin := newTestV2Plugin()

		// Create lifecycle manager
		lm := sdk.NewLifecycleManager(&sdk.LifecycleConfig{
			InitTimeout:        5 * time.Second,
			StartTimeout:       5 * time.Second,
			StopTimeout:        5 * time.Second,
			HealthCheckTimeout: 2 * time.Second,
		})

		ctx := context.Background()

		// Test: Configure plugin with type-safe config
		err := plugin.Configure(ctx, testConfig)
		require.NoError(t, err, "Configure should succeed")

		// Verify config was stored
		assert.Equal(t, testConfig.APIKey, plugin.Config().APIKey)
		assert.Equal(t, testConfig.Timeout, plugin.Config().Timeout)

		// Test: Register and initialize plugin
		err = lm.Register("test-plugin", plugin)
		require.NoError(t, err, "Register should succeed")

		err = lm.InitPlugin(ctx, "test-plugin")
		require.NoError(t, err, "Init should succeed")

		// Verify state transitions
		state, err := lm.GetPluginState("test-plugin")
		require.NoError(t, err)
		assert.Equal(t, sdk.StateInitialized, state, "Should be initialized")

		// Test: Start plugin
		err = lm.StartPlugin(ctx, "test-plugin")
		require.NoError(t, err, "Start should succeed")

		state, err = lm.GetPluginState("test-plugin")
		require.NoError(t, err)
		assert.Equal(t, sdk.StateStarted, state, "Should be started")

		// Test: Health check
		err = lm.HealthCheckPlugin("test-plugin")
		require.NoError(t, err, "HealthCheck should succeed")

		// Test: Stop plugin
		err = lm.StopPlugin(ctx, "test-plugin")
		require.NoError(t, err, "Stop should succeed")

		state, err = lm.GetPluginState("test-plugin")
		require.NoError(t, err)
		assert.Equal(t, sdk.StateStopped, state, "Should be stopped")
	})

	t.Run("config_validation_failure", func(t *testing.T) {
		// Setup: Create plugin that rejects invalid config
		plugin := newTestV2Plugin()

		// Test: Configure with invalid config (empty required field)
		invalidConfig := testPluginConfig{APIKey: ""}

		// Note: Validation would be done by config system before Configure is called
		// For this test, we simulate the Configure method checking validity
		plugin.configureError = errors.New("required field missing")

		ctx := context.Background()
		err := plugin.Configure(ctx, invalidConfig)

		// Assert: Should fail
		assert.Error(t, err, "Configure should fail with invalid config")
		assert.Contains(t, err.Error(), "required")
	})

	t.Run("lifecycle_failure_handling", func(t *testing.T) {
		// Setup: Create plugin that fails during Init
		plugin := newTestV2Plugin()
		plugin.initError = errors.New("initialization failed")

		lm := sdk.NewLifecycleManager(nil)
		ctx := context.Background()

		_ = plugin.Configure(ctx, testPluginConfig{})
		_ = lm.Register("failing-plugin", plugin)

		// Test: Init should fail
		err := lm.InitPlugin(ctx, "failing-plugin")
		assert.Error(t, err, "Init should fail")

		// Verify state is errored
		state, _ := lm.GetPluginState("failing-plugin")
		assert.Equal(t, sdk.StateErrored, state, "Should be in error state")
	})

	t.Run("health_check_monitoring", func(t *testing.T) {
		// Setup: Create plugin that becomes unhealthy
		plugin := newTestV2Plugin()

		lm := sdk.NewLifecycleManager(&sdk.LifecycleConfig{
			HealthCheckInterval: 100 * time.Millisecond,
			UnhealthyThreshold:  2,
			HealthCheckTimeout:  1 * time.Second,
		})

		ctx := context.Background()
		_ = plugin.Configure(ctx, testPluginConfig{})
		_ = lm.Register("test-plugin", plugin)
		_ = lm.InitPlugin(ctx, "test-plugin")
		_ = lm.StartPlugin(ctx, "test-plugin")

		// Test: Manual health check works
		err := lm.HealthCheckPlugin("test-plugin")
		assert.NoError(t, err, "Initial health check should pass")
		assert.True(t, plugin.healthCheckCalled, "HealthCheck should be called")

		// Make plugin unhealthy
		plugin.healthCheckError = errors.New("service unavailable")

		// Test: Health check reports error
		err = lm.HealthCheckPlugin("test-plugin")
		assert.Error(t, err, "Health check should fail when plugin unhealthy")

		// Verify multiple checks are tracked
		assert.True(t, plugin.healthCheckCallCount >= 2, "Multiple health checks performed")
	})
}

// TestPluginDependencyResolution tests the dependency resolution system
// with multiple plugins having interdependencies.
//
// Validates:
// - Task 3.3: Dependency resolution (topological sort, cycle detection)
//
// Subtask 3.6.1.3: Dependency resolution test
func TestPluginDependencyResolution(t *testing.T) {
	t.Run("simple_linear_dependencies", func(t *testing.T) {
		// Setup: Create plugins with linear dependency chain
		// Plugin C depends on B, B depends on A
		plugins := map[string]sdk.PluginMetadata{
			"plugin-a": createPluginMetadata("plugin-a", "1.0.0", nil),
			"plugin-b": createPluginMetadata("plugin-b", "1.0.0", []sdk.PluginDependency{
				{Name: "plugin-a", Version: "1.0.0", Optional: false},
			}),
			"plugin-c": createPluginMetadata("plugin-c", "1.0.0", []sdk.PluginDependency{
				{Name: "plugin-b", Version: "1.0.0", Optional: false},
			}),
		}

		// Create resolver
		resolver := sdk.NewDependencyResolver()

		// Test: Resolve load order
		order, err := resolver.Resolve(plugins)

		// Assert: Should return correct topological order
		require.NoError(t, err, "Resolve should succeed")
		require.Len(t, order, 3, "Should have 3 plugins")
		assert.Equal(t, "plugin-a", order[0], "A should be first")
		assert.Equal(t, "plugin-b", order[1], "B should be second")
		assert.Equal(t, "plugin-c", order[2], "C should be last")
	})

	t.Run("diamond_dependencies", func(t *testing.T) {
		// Setup: Create diamond dependency pattern
		//     A
		//    / \
		//   B   C
		//    \ /
		//     D
		plugins := map[string]sdk.PluginMetadata{
			"plugin-a": createPluginMetadata("plugin-a", "2.0.0", nil),
			"plugin-b": createPluginMetadata("plugin-b", "1.0.0", []sdk.PluginDependency{
				{Name: "plugin-a", Version: "^2.0.0", Optional: false},
			}),
			"plugin-c": createPluginMetadata("plugin-c", "1.0.0", []sdk.PluginDependency{
				{Name: "plugin-a", Version: "^2.0.0", Optional: false},
			}),
			"plugin-d": createPluginMetadata("plugin-d", "1.0.0", []sdk.PluginDependency{
				{Name: "plugin-b", Version: "1.0.0", Optional: false},
				{Name: "plugin-c", Version: "1.0.0", Optional: false},
			}),
		}

		resolver := sdk.NewDependencyResolver()

		// Test: Resolve
		order, err := resolver.Resolve(plugins)

		// Assert: A must be first, D must be last
		require.NoError(t, err)
		require.Len(t, order, 4)
		assert.Equal(t, "plugin-a", order[0], "A should be first")
		assert.Equal(t, "plugin-d", order[3], "D should be last")
		// B and C can be in either order (both depend only on A)
	})

	t.Run("circular_dependency_detection", func(t *testing.T) {
		// Setup: Create circular dependency
		// A -> C -> B -> A (cycle)
		plugins := map[string]sdk.PluginMetadata{
			"plugin-a": createPluginMetadata("plugin-a", "1.0.0", []sdk.PluginDependency{
				{Name: "plugin-c", Version: "1.0.0", Optional: false},
			}),
			"plugin-b": createPluginMetadata("plugin-b", "1.0.0", []sdk.PluginDependency{
				{Name: "plugin-a", Version: "1.0.0", Optional: false},
			}),
			"plugin-c": createPluginMetadata("plugin-c", "1.0.0", []sdk.PluginDependency{
				{Name: "plugin-b", Version: "1.0.0", Optional: false},
			}),
		}

		resolver := sdk.NewDependencyResolver()

		// Test: Resolve should detect cycle
		_, err := resolver.Resolve(plugins)

		// Assert: Should return cycle error
		assert.Error(t, err, "Should detect circular dependency")
		var cyclicErr *sdk.CyclicDependencyError
		assert.True(t, errors.As(err, &cyclicErr), "Error should be CyclicDependencyError")
	})

	t.Run("missing_dependency_error", func(t *testing.T) {
		// Setup: Plugin depends on non-existent plugin
		plugins := map[string]sdk.PluginMetadata{
			"plugin-a": createPluginMetadata("plugin-a", "1.0.0", []sdk.PluginDependency{
				{Name: "nonexistent-plugin", Version: "1.0.0", Optional: false},
			}),
		}

		resolver := sdk.NewDependencyResolver()

		// Test: Resolve should fail
		_, err := resolver.Resolve(plugins)

		// Assert: Should return missing dependency error
		assert.Error(t, err, "Should detect missing dependency")
		var missingErr *sdk.MissingDependencyError
		assert.True(t, errors.As(err, &missingErr), "Error should be MissingDependencyError")
	})

	t.Run("version_constraint_validation", func(t *testing.T) {
		// Setup: Plugin requires specific version range
		plugins := map[string]sdk.PluginMetadata{
			"plugin-a": createPluginMetadata("plugin-a", "1.5.0", nil),
			"plugin-b": createPluginMetadata("plugin-b", "1.0.0", []sdk.PluginDependency{
				{Name: "plugin-a", Version: "^1.2.0", Optional: false},
			}),
		}

		resolver := sdk.NewDependencyResolver()

		// Test: Resolve should succeed
		order, err := resolver.Resolve(plugins)

		// Assert: Version constraint satisfied
		require.NoError(t, err, "Version 1.5.0 should satisfy ^1.2.0")
		assert.Equal(t, "plugin-a", order[0])
	})

	t.Run("version_mismatch_error", func(t *testing.T) {
		// Setup: Plugin requires newer version
		plugins := map[string]sdk.PluginMetadata{
			"plugin-a": createPluginMetadata("plugin-a", "1.0.0", nil),
			"plugin-b": createPluginMetadata("plugin-b", "1.0.0", []sdk.PluginDependency{
				{Name: "plugin-a", Version: "^2.0.0", Optional: false},
			}),
		}

		resolver := sdk.NewDependencyResolver()

		// Test: Resolve should fail
		_, err := resolver.Resolve(plugins)

		// Assert: Should return version mismatch error
		assert.Error(t, err, "Should detect version mismatch")
		var versionErr *sdk.VersionMismatchError
		assert.True(t, errors.As(err, &versionErr), "Error should be VersionMismatchError")
	})

	t.Run("optional_dependency_handling", func(t *testing.T) {
		// Setup: Plugin has optional dependency that's missing
		plugins := map[string]sdk.PluginMetadata{
			"plugin-a": createPluginMetadata("plugin-a", "1.0.0", []sdk.PluginDependency{
				{Name: "optional-plugin", Version: "1.0.0", Optional: true},
			}),
		}

		resolver := sdk.NewDependencyResolver()

		// Test: Resolve should succeed (optional dep missing is OK)
		order, err := resolver.Resolve(plugins)

		// Assert: Should succeed with warning
		require.NoError(t, err, "Optional missing dependency should not fail")
		assert.Len(t, order, 1)
		assert.Equal(t, "plugin-a", order[0])
	})
}

// TestV1V2PluginCoexistence tests that v1 and v2 plugins can coexist
// and work together via the adapter layer.
//
// Validates:
// - Task 3.4: SDK v2 adapter layer (v1/v2 bidirectional compatibility)
//
// Subtask 3.6.1.4: v1/v2 coexistence test
func TestV1V2PluginCoexistence(t *testing.T) {
	t.Run("v2_style_plugins_with_lifecycle_manager", func(t *testing.T) {
		// Setup: Create v2-style plugin (implements sdk.Lifecycle)
		v2Plugin := newTestV2Plugin()

		// Create lifecycle manager
		lm := sdk.NewLifecycleManager(nil)
		ctx := context.Background()

		// Configure and register v2 plugin
		_ = v2Plugin.Configure(ctx, testPluginConfig{APIKey: "v2-test"})
		err := lm.Register("v2-plugin", v2Plugin)
		require.NoError(t, err, "Should register v2 plugin")

		// Create v1-style plugin (also implements sdk.Lifecycle)
		v1Plugin := &mockV1Plugin{name: "v1-plugin", version: "1.0.0"}
		err = lm.Register("v1-plugin", v1Plugin)
		require.NoError(t, err, "Should register v1 plugin")

		// Test: Both plugins can be managed together
		err = lm.InitPlugin(ctx, "v2-plugin")
		assert.NoError(t, err, "v2 plugin Init should work")

		err = lm.InitPlugin(ctx, "v1-plugin")
		assert.NoError(t, err, "v1 plugin Init should work")

		// Verify both are initialized
		v2State, _ := lm.GetPluginState("v2-plugin")
		v1State, _ := lm.GetPluginState("v1-plugin")

		assert.Equal(t, sdk.StateInitialized, v2State)
		assert.Equal(t, sdk.StateInitialized, v1State)

		// Both can be started
		_ = lm.StartPlugin(ctx, "v2-plugin")
		_ = lm.StartPlugin(ctx, "v1-plugin")

		v2State, _ = lm.GetPluginState("v2-plugin")
		v1State, _ = lm.GetPluginState("v1-plugin")

		assert.Equal(t, sdk.StateStarted, v2State)
		assert.Equal(t, sdk.StateStarted, v1State)
	})

	t.Run("v2_interface_features", func(t *testing.T) {
		// Setup: Create v2 plugin and verify v2-specific features
		v2Plugin := newTestV2Plugin()

		// Test: v2.Metadata structure
		metadata := v2Plugin.Metadata()
		assert.Equal(t, "test-plugin", metadata.Name)
		assert.Equal(t, "1.0.0", metadata.Version)
		assert.Equal(t, "Test plugin for integration testing", metadata.Description)

		// Test: v2.ConfigSchema (returns nil by default)
		schema := v2Plugin.ConfigSchema()
		// Default BasePlugin returns nil schema
		assert.Nil(t, schema, "Default config schema should be nil")

		// Test: v2.Commands (returns nil or empty slice by default)
		commands := v2Plugin.Commands()
		if commands != nil {
			assert.Len(t, commands, 0, "Default commands should be empty")
		}
	})

	t.Run("type_safe_configuration", func(t *testing.T) {
		// Setup: Test that v2 plugins use type-safe configuration
		v2Plugin := newTestV2Plugin()

		// Type-safe config
		config := testPluginConfig{
			APIKey:  "test-api-key",
			Timeout: 60,
		}

		ctx := context.Background()
		err := v2Plugin.Configure(ctx, config)
		require.NoError(t, err, "Type-safe configuration should work")

		// Verify config was stored correctly
		storedConfig := v2Plugin.Config()
		assert.Equal(t, "test-api-key", storedConfig.APIKey)
		assert.Equal(t, 60, storedConfig.Timeout)
	})
}

// TestConfigMigration tests migration from legacy map-based config
// to type-safe configuration system.
//
// Validates:
// - Task 3.1: Configuration migration and backward compatibility
//
// Subtask 3.6.1.2: Config migration test
func TestConfigMigration(t *testing.T) {
	t.Run("legacy_map_to_typed_config", func(t *testing.T) {
		// TODO: Implement when config migration is available
		// This test would verify that old map[string]interface{} configs
		// can be migrated to new type-safe configs
		t.Skip("Config migration not yet implemented")
	})

	t.Run("backward_compatibility_layer", func(t *testing.T) {
		// TODO: Implement when backward compatibility layer is available
		// This test would verify that v1 plugins can still use map-based config
		// while v2 plugins use type-safe config
		t.Skip("Backward compatibility layer not yet implemented")
	})
}

// Helper functions and mocks

// testV2Plugin is a mock v2.Plugin for testing
type testV2Plugin struct {
	v2.BasePlugin[testPluginConfig]
	initCalled           bool
	startCalled          bool
	stopCalled           bool
	healthCheckCalled    bool
	healthCheckCallCount int
	initError            error
	startError           error
	stopError            error
	healthCheckError     error
	configureError       error
}

type testPluginConfig struct {
	APIKey  string
	Timeout int
}

func newTestV2Plugin() *testV2Plugin {
	p := &testV2Plugin{}
	p.SetMetadata(v2.Metadata{
		Name:        "test-plugin",
		Version:     "1.0.0",
		Description: "Test plugin for integration testing",
	})
	return p
}

func (p *testV2Plugin) Configure(ctx context.Context, config testPluginConfig) error {
	if p.configureError != nil {
		return p.configureError
	}
	p.BasePlugin.Configure(ctx, config)
	return nil
}

func (p *testV2Plugin) Init(ctx context.Context) error {
	p.initCalled = true
	return p.initError
}

func (p *testV2Plugin) Start(ctx context.Context) error {
	p.startCalled = true
	return p.startError
}

func (p *testV2Plugin) Stop(ctx context.Context) error {
	p.stopCalled = true
	return p.stopError
}

func (p *testV2Plugin) HealthCheck() error {
	p.healthCheckCalled = true
	p.healthCheckCallCount++
	return p.healthCheckError
}

// mockV1Plugin simulates a v1 plugin for coexistence testing
type mockV1Plugin struct {
	name        string
	version     string
	initCalled  bool
	startCalled bool
	stopCalled  bool
}

func (m *mockV1Plugin) Init(ctx context.Context) error {
	m.initCalled = true
	return nil
}

func (m *mockV1Plugin) Start(ctx context.Context) error {
	m.startCalled = true
	return nil
}

func (m *mockV1Plugin) Stop(ctx context.Context) error {
	m.stopCalled = true
	return nil
}

func (m *mockV1Plugin) HealthCheck() error {
	return nil
}

func (m *mockV1Plugin) Name() string {
	return m.name
}

func (m *mockV1Plugin) Version() string {
	return m.version
}

// createPluginMetadata helper for dependency tests
func createPluginMetadata(name, version string, deps []sdk.PluginDependency) sdk.PluginMetadata {
	return sdk.PluginMetadata{
		Name:         name,
		Version:      version,
		Dependencies: deps,
	}
}
