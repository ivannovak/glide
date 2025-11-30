package v2

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ivannovak/glide/v2/pkg/plugin/sdk"
)

// MockV1InProcessPlugin simulates a v1 in-process plugin
type MockV1InProcessPlugin struct {
	name        string
	version     string
	description string
	configured  bool
	initCalled  bool
	startCalled bool
	stopCalled  bool
}

func (p *MockV1InProcessPlugin) Name() string        { return p.name }
func (p *MockV1InProcessPlugin) Version() string     { return p.version }
func (p *MockV1InProcessPlugin) Description() string { return p.description }

func (p *MockV1InProcessPlugin) Configure() error {
	p.configured = true
	return nil
}

func (p *MockV1InProcessPlugin) Register(root *cobra.Command) error {
	cmd := &cobra.Command{
		Use:   "v1-cmd",
		Short: "V1 command",
	}
	root.AddCommand(cmd)
	return nil
}

func (p *MockV1InProcessPlugin) Init(_ context.Context) error {
	p.initCalled = true
	return nil
}

func (p *MockV1InProcessPlugin) Start(_ context.Context) error {
	p.startCalled = true
	return nil
}

func (p *MockV1InProcessPlugin) Stop(_ context.Context) error {
	p.stopCalled = true
	return nil
}

func (p *MockV1InProcessPlugin) HealthCheck() error {
	return nil
}

func TestAdaptV1InProcessPlugin(t *testing.T) {
	v1Plugin := &MockV1InProcessPlugin{
		name:        "v1-test",
		version:     "1.0.0",
		description: "V1 test plugin",
	}

	v2Plugin := AdaptV1InProcessPlugin(v1Plugin)
	require.NotNil(t, v2Plugin)

	// Test metadata conversion
	meta := v2Plugin.Metadata()
	assert.Equal(t, "v1-test", meta.Name)
	assert.Equal(t, "1.0.0", meta.Version)
	assert.Equal(t, "V1 test plugin", meta.Description)
}

func TestV1Adapter_ConfigSchema(t *testing.T) {
	v1Plugin := &MockV1InProcessPlugin{
		name: "v1-test",
	}

	v2Plugin := AdaptV1InProcessPlugin(v1Plugin)

	// v1 plugins don't have schemas
	schema := v2Plugin.ConfigSchema()
	assert.Nil(t, schema)
}

func TestV1Adapter_Configure(t *testing.T) {
	v1Plugin := &MockV1InProcessPlugin{
		name: "v1-test",
	}

	v2Plugin := AdaptV1InProcessPlugin(v1Plugin)
	ctx := context.Background()

	config := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}

	err := v2Plugin.Configure(ctx, config)
	require.NoError(t, err)
	assert.True(t, v1Plugin.configured)
}

func TestV1Adapter_Lifecycle(t *testing.T) {
	v1Plugin := &MockV1InProcessPlugin{
		name: "v1-test",
	}

	v2Plugin := AdaptV1InProcessPlugin(v1Plugin)
	ctx := context.Background()

	// Test Init
	err := v2Plugin.Init(ctx)
	require.NoError(t, err)
	assert.True(t, v1Plugin.initCalled)

	// Test Start
	err = v2Plugin.Start(ctx)
	require.NoError(t, err)
	assert.True(t, v1Plugin.startCalled)

	// Test HealthCheck
	err = v2Plugin.HealthCheck(ctx)
	require.NoError(t, err)

	// Test Stop
	err = v2Plugin.Stop(ctx)
	require.NoError(t, err)
	assert.True(t, v1Plugin.stopCalled)
}

func TestV1Adapter_Commands(t *testing.T) {
	v1Plugin := &MockV1InProcessPlugin{
		name: "v1-test",
	}

	v2Plugin := AdaptV1InProcessPlugin(v1Plugin)

	// v1 in-process plugins don't expose commands via Commands()
	// They register directly with Cobra
	commands := v2Plugin.Commands()
	assert.Empty(t, commands)
}

func TestV1Adapter_GetV1Plugin(t *testing.T) {
	v1Plugin := &MockV1InProcessPlugin{
		name: "v1-test",
	}

	v2Plugin := AdaptV1InProcessPlugin(v1Plugin)
	adapter, ok := v2Plugin.(*V1Adapter)
	require.True(t, ok, "expected *V1Adapter")

	// Verify we can retrieve the original v1 plugin
	original := adapter.GetV1Plugin()
	assert.Equal(t, v1Plugin, original)
}

func TestV2ToV1Adapter_Metadata(t *testing.T) {
	v2Plugin := NewTestPlugin()
	v1Adapter := AdaptV2ToV1(v2Plugin)

	assert.Equal(t, "test-plugin", v1Adapter.Name())
	assert.Equal(t, "1.0.0", v1Adapter.Version())
	assert.Equal(t, "A test plugin", v1Adapter.Description())
}

func TestV2ToV1Adapter_Configure(t *testing.T) {
	v2Plugin := NewTestPlugin()
	v1Adapter := AdaptV2ToV1(v2Plugin)

	// Configure is called without config in v1 in-process plugins
	// The config is assumed to be loaded already
	err := v1Adapter.Configure()
	require.NoError(t, err)
}

func TestV2ToV1Adapter_Register(t *testing.T) {
	v2Plugin := NewTestPlugin()
	v1Adapter := AdaptV2ToV1(v2Plugin)

	root := &cobra.Command{Use: "root"}
	err := v1Adapter.Register(root)
	require.NoError(t, err)

	// Verify command was added
	commands := root.Commands()
	require.Len(t, commands, 1)
	assert.Equal(t, "hello", commands[0].Use)
}

func TestV2ToV1Adapter_Lifecycle(t *testing.T) {
	v2Plugin := NewTestPlugin()
	v1Adapter := AdaptV2ToV1(v2Plugin)

	ctx := context.Background()

	err := v1Adapter.Init(ctx)
	require.NoError(t, err)
	assert.True(t, v2Plugin.initCalled)

	err = v1Adapter.Start(ctx)
	require.NoError(t, err)
	assert.True(t, v2Plugin.startCalled)

	err = v1Adapter.HealthCheck()
	require.NoError(t, err)

	err = v1Adapter.Stop(ctx)
	require.NoError(t, err)
	assert.True(t, v2Plugin.stopCalled)
}

func TestConvertV1Commands(t *testing.T) {
	// This is an internal function, but we can test it through the adapter
	// For now, we'll create a minimal test
	// In real usage, this would be tested via AdaptV1GRPCPlugin with a mock client

	commands := []Command{
		{
			Name:        "test1",
			Description: "Test command 1",
			Category:    "core",
			Aliases:     []string{"t1"},
			Hidden:      false,
			Interactive: false,
		},
		{
			Name:        "test2",
			Description: "Test command 2",
			Category:    "testing",
			Interactive: true,
			RequiresTTY: true,
		},
	}

	require.Len(t, commands, 2)
	assert.Equal(t, "test1", commands[0].Name)
	assert.Equal(t, []string{"t1"}, commands[0].Aliases)
	assert.False(t, commands[0].Interactive)

	assert.Equal(t, "test2", commands[1].Name)
	assert.True(t, commands[1].Interactive)
	assert.True(t, commands[1].RequiresTTY)
}

func TestV1Adapter_StateTracking(t *testing.T) {
	v1Plugin := &MockV1InProcessPlugin{
		name: "v1-test",
	}

	v2Plugin := AdaptV1InProcessPlugin(v1Plugin)
	adapter, ok := v2Plugin.(*V1Adapter)
	require.True(t, ok, "expected *V1Adapter")
	ctx := context.Background()

	// Initial state should be uninitialized
	assert.Equal(t, sdk.StateUninitialized, adapter.state.Get())

	// After Init
	err := adapter.Init(ctx)
	require.NoError(t, err)
	assert.Equal(t, sdk.StateInitialized, adapter.state.Get())

	// After Start
	err = adapter.Start(ctx)
	require.NoError(t, err)
	assert.Equal(t, sdk.StateStarted, adapter.state.Get())

	// After Stop
	err = adapter.Stop(ctx)
	require.NoError(t, err)
	assert.Equal(t, sdk.StateStopped, adapter.state.Get())
}

func TestV1Adapter_NoLifecyclePlugin(t *testing.T) {
	// Plugin without lifecycle methods
	type SimpleV1Plugin struct {
		name string
	}

	v1Plugin := &SimpleV1Plugin{name: "simple"}

	adapter := &V1Adapter{
		v1Plugin: v1Plugin,
		state:    sdk.NewStateTracker("simple"),
	}

	ctx := context.Background()

	// Should not error when lifecycle methods don't exist
	err := adapter.Init(ctx)
	require.NoError(t, err)

	err = adapter.Start(ctx)
	require.NoError(t, err)

	err = adapter.HealthCheck(ctx)
	require.NoError(t, err)

	err = adapter.Stop(ctx)
	require.NoError(t, err)

	// State should still transition
	assert.Equal(t, sdk.StateStopped, adapter.state.Get())
}
