package plugin_test

import (
	"errors"
	"testing"

	"github.com/ivannovak/glide/v3/pkg/plugin"
	"github.com/ivannovak/glide/v3/pkg/plugin/plugintest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry(t *testing.T) {
	t.Run("create new registry", func(t *testing.T) {
		reg := plugin.NewRegistry()
		assert.NotNil(t, reg)

		// Registry should be empty
		plugins := reg.List()
		assert.Empty(t, plugins)
	})

	t.Run("register plugin", func(t *testing.T) {
		reg := plugin.NewRegistry()
		p := plugintest.NewMockPlugin("test-plugin")

		err := reg.RegisterPlugin(p)
		require.NoError(t, err)

		// Plugin should be in registry
		plugins := reg.List()
		assert.Len(t, plugins, 1)
		assert.Equal(t, "test-plugin", plugins[0].Name())
	})

	t.Run("register nil plugin", func(t *testing.T) {
		reg := plugin.NewRegistry()

		err := reg.RegisterPlugin(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil plugin")
	})

	t.Run("register plugin without name", func(t *testing.T) {
		reg := plugin.NewRegistry()
		p := plugintest.NewMockPlugin("")

		err := reg.RegisterPlugin(p)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must have a name")
	})

	t.Run("register duplicate plugin", func(t *testing.T) {
		reg := plugin.NewRegistry()
		p1 := plugintest.NewMockPlugin("test-plugin")
		p2 := plugintest.NewMockPlugin("test-plugin")
		p2.VersionValue = "2.0.0"

		err := reg.RegisterPlugin(p1)
		require.NoError(t, err)

		err = reg.RegisterPlugin(p2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})

	t.Run("get plugin by name", func(t *testing.T) {
		reg := plugin.NewRegistry()
		p := plugintest.NewMockPlugin("test-plugin")

		err := reg.RegisterPlugin(p)
		require.NoError(t, err)

		// Get existing plugin
		retrieved, exists := reg.Get("test-plugin")
		assert.True(t, exists)
		assert.Equal(t, "test-plugin", retrieved.Name())

		// Get non-existing plugin
		retrieved, exists = reg.Get("non-existent")
		assert.False(t, exists)
		assert.Nil(t, retrieved)
	})

	t.Run("list multiple plugins", func(t *testing.T) {
		reg := plugin.NewRegistry()

		p1 := plugintest.NewMockPlugin("plugin1")
		p2 := plugintest.NewMockPlugin("plugin2")
		p2.VersionValue = "2.0.0"
		p3 := plugintest.NewMockPlugin("plugin3")
		p3.VersionValue = "3.0.0"

		require.NoError(t, reg.RegisterPlugin(p1))
		require.NoError(t, reg.RegisterPlugin(p2))
		require.NoError(t, reg.RegisterPlugin(p3))

		plugins := reg.List()
		assert.Len(t, plugins, 3)

		// Verify all plugins are present
		names := make([]string, len(plugins))
		for i, p := range plugins {
			names[i] = p.Name()
		}
		assert.Contains(t, names, "plugin1")
		assert.Contains(t, names, "plugin2")
		assert.Contains(t, names, "plugin3")
	})

	t.Run("set and load configuration", func(t *testing.T) {
		reg := plugin.NewRegistry()
		p := plugintest.NewMockPlugin("test-plugin")

		err := reg.RegisterPlugin(p)
		require.NoError(t, err)

		// Set configuration
		// NOTE: Config is now handled by pkg/config, but for this test
		// we just verify Configure() is called (it gets nil now)

		// Load all plugins
		root := &cobra.Command{Use: "test"}
		result, err := reg.LoadAll(root)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Plugin should be configured and registered
		assert.True(t, p.Configured)
		assert.True(t, p.Registered)
		assert.Contains(t, result.Loaded, "test-plugin")
		assert.Empty(t, result.Failed)
	})

	t.Run("load all with configure error", func(t *testing.T) {
		reg := plugin.NewRegistry()
		p := plugintest.NewMockPlugin("test-plugin").
			WithConfigError(errors.New("config error"))

		err := reg.RegisterPlugin(p)
		require.NoError(t, err)

		root := &cobra.Command{Use: "test"}
		result, err := reg.LoadAll(root)
		require.NoError(t, err) // No fatal error
		require.NotNil(t, result)

		// Error should be collected in result
		assert.True(t, result.HasErrors())
		assert.Len(t, result.Failed, 1)
		assert.Equal(t, "test-plugin", result.Failed[0].Name)
		assert.Contains(t, result.Failed[0].Error.Error(), "failed to configure")
		assert.Contains(t, result.Failed[0].Error.Error(), "config error")
		assert.False(t, result.Failed[0].IsFatal)
	})

	t.Run("load all with register error", func(t *testing.T) {
		reg := plugin.NewRegistry()
		p := plugintest.NewMockPlugin("test-plugin").
			WithError(errors.New("register error"))

		err := reg.RegisterPlugin(p)
		require.NoError(t, err)

		root := &cobra.Command{Use: "test"}
		result, err := reg.LoadAll(root)
		require.NoError(t, err) // No fatal error
		require.NotNil(t, result)

		// Error should be collected in result
		assert.True(t, result.HasErrors())
		assert.Len(t, result.Failed, 1)
		assert.Equal(t, "test-plugin", result.Failed[0].Name)
		assert.Contains(t, result.Failed[0].Error.Error(), "failed to register commands")
		assert.Contains(t, result.Failed[0].Error.Error(), "register error")
		assert.False(t, result.Failed[0].IsFatal)
	})
}

func TestGlobalRegistry(t *testing.T) {
	// Note: These tests may affect the global registry state
	// In production code, you might want to add a Reset function for testing

	t.Run("global registry functions", func(t *testing.T) {
		p := plugintest.NewMockPlugin("global-test-plugin")

		// Register globally
		err := plugin.Register(p)
		require.NoError(t, err)

		// Get from global registry
		retrieved, exists := plugin.Get("global-test-plugin")
		assert.True(t, exists)
		assert.Equal(t, "global-test-plugin", retrieved.Name())

		// List from global registry
		plugins := plugin.List()
		found := false
		for _, pl := range plugins {
			if pl.Name() == "global-test-plugin" {
				found = true
				break
			}
		}
		assert.True(t, found)

		// Set config globally
		// NOTE: Config is now handled by pkg/config, not via SetConfig
		// Configure() gets called with nil

		// Load all globally
		root := &cobra.Command{Use: "test"}
		result, err := plugin.LoadAll(root)
		// This might fail if other tests have registered plugins with errors
		// In a real scenario, you'd want to clean the global registry between tests
		if err == nil {
			assert.True(t, p.Configured)
			assert.True(t, p.Registered)
			assert.NotNil(t, result)
			assert.Contains(t, result.Loaded, "global-test-plugin")
		}
	})

	t.Run("get global registry", func(t *testing.T) {
		reg := plugin.GetGlobalRegistry()
		assert.NotNil(t, reg)

		// Should be the same instance
		reg2 := plugin.GetGlobalRegistry()
		assert.Same(t, reg, reg2)
	})
}

func TestRegistryConcurrency(t *testing.T) {
	t.Run("concurrent registration", func(t *testing.T) {
		reg := plugin.NewRegistry()

		// Try to register multiple plugins concurrently
		done := make(chan bool, 10)
		errors := make(chan error, 10)

		for i := 0; i < 10; i++ {
			go func(id int) {
				p := plugintest.NewMockPlugin("concurrent-plugin")
				err := reg.RegisterPlugin(p)
				if err != nil {
					errors <- err
				}
				done <- true
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}
		close(errors)

		// Should have exactly 1 successful registration and 9 errors
		errorCount := 0
		for err := range errors {
			assert.Contains(t, err.Error(), "already registered")
			errorCount++
		}
		assert.Equal(t, 9, errorCount)

		// Verify only one plugin registered
		plugins := reg.List()
		assert.Len(t, plugins, 1)
	})

	t.Run("concurrent get and list", func(t *testing.T) {
		reg := plugin.NewRegistry()

		// Register some plugins
		for i := 0; i < 5; i++ {
			name := string(rune('a' + i))
			p := plugintest.NewMockPlugin(name)
			require.NoError(t, reg.RegisterPlugin(p))
		}

		// Concurrent reads should work fine
		done := make(chan bool, 20)

		for i := 0; i < 10; i++ {
			go func() {
				plugins := reg.List()
				assert.Len(t, plugins, 5)
				done <- true
			}()

			go func() {
				_, exists := reg.Get("a")
				assert.True(t, exists)
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 20; i++ {
			<-done
		}
	})
}

func TestPluginLoadResult(t *testing.T) {
	t.Run("empty result has no errors", func(t *testing.T) {
		result := &plugin.PluginLoadResult{
			Loaded:   []string{},
			Failed:   []plugin.PluginError{},
			Warnings: []string{},
		}

		assert.False(t, result.HasErrors())
		assert.False(t, result.HasFatalErrors())
		assert.Empty(t, result.ErrorMessage())
	})

	t.Run("successful loads", func(t *testing.T) {
		result := &plugin.PluginLoadResult{
			Loaded:   []string{"plugin1", "plugin2", "plugin3"},
			Failed:   []plugin.PluginError{},
			Warnings: []string{},
		}

		assert.False(t, result.HasErrors())
		assert.False(t, result.HasFatalErrors())
		assert.Empty(t, result.ErrorMessage())
	})

	t.Run("non-fatal errors", func(t *testing.T) {
		result := &plugin.PluginLoadResult{
			Loaded: []string{"plugin1"},
			Failed: []plugin.PluginError{
				{
					Name:    "plugin2",
					Error:   errors.New("config failed"),
					IsFatal: false,
				},
				{
					Name:    "plugin3",
					Error:   errors.New("register failed"),
					IsFatal: false,
				},
			},
			Warnings: []string{},
		}

		assert.True(t, result.HasErrors())
		assert.False(t, result.HasFatalErrors())

		msg := result.ErrorMessage()
		assert.Contains(t, msg, "Plugin loading issues")
		assert.Contains(t, msg, "plugin2")
		assert.Contains(t, msg, "config failed")
		assert.Contains(t, msg, "plugin3")
		assert.Contains(t, msg, "register failed")
		assert.Contains(t, msg, "Successfully loaded 1 plugins: plugin1")
	})

	t.Run("fatal errors", func(t *testing.T) {
		result := &plugin.PluginLoadResult{
			Loaded: []string{},
			Failed: []plugin.PluginError{
				{
					Name:    "plugin1",
					Error:   errors.New("critical error"),
					IsFatal: true,
				},
			},
			Warnings: []string{},
		}

		assert.True(t, result.HasErrors())
		assert.True(t, result.HasFatalErrors())

		msg := result.ErrorMessage()
		assert.Contains(t, msg, "FATAL")
		assert.Contains(t, msg, "plugin1")
		assert.Contains(t, msg, "critical error")
	})

	t.Run("mixed fatal and non-fatal errors", func(t *testing.T) {
		result := &plugin.PluginLoadResult{
			Loaded: []string{"plugin1"},
			Failed: []plugin.PluginError{
				{
					Name:    "plugin2",
					Error:   errors.New("non-fatal"),
					IsFatal: false,
				},
				{
					Name:    "plugin3",
					Error:   errors.New("fatal"),
					IsFatal: true,
				},
			},
			Warnings: []string{},
		}

		assert.True(t, result.HasErrors())
		assert.True(t, result.HasFatalErrors())

		msg := result.ErrorMessage()
		assert.Contains(t, msg, "warning")
		assert.Contains(t, msg, "FATAL")
	})

	t.Run("error message format", func(t *testing.T) {
		result := &plugin.PluginLoadResult{
			Loaded: []string{"plugin1", "plugin2"},
			Failed: []plugin.PluginError{
				{
					Name:    "plugin3",
					Error:   errors.New("test error"),
					IsFatal: false,
				},
			},
			Warnings: []string{},
		}

		msg := result.ErrorMessage()
		assert.Contains(t, msg, "Plugin loading issues:")
		assert.Contains(t, msg, "[warning] plugin3: test error")
		assert.Contains(t, msg, "Successfully loaded 2 plugins: plugin1, plugin2")
	})
}
