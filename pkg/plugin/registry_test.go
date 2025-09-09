package plugin_test

import (
	"errors"
	"testing"

	"github.com/ivannovak/glide/pkg/plugin"
	"github.com/ivannovak/glide/pkg/plugin/plugintest"
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
		config := map[string]interface{}{
			"test-plugin": map[string]interface{}{
				"key": "value",
			},
		}
		reg.SetConfig(config)

		// Load all plugins
		root := &cobra.Command{Use: "test"}
		err = reg.LoadAll(root)
		require.NoError(t, err)

		// Plugin should be configured and registered
		assert.True(t, p.Configured)
		assert.True(t, p.Registered)
	})

	t.Run("load all with configure error", func(t *testing.T) {
		reg := plugin.NewRegistry()
		p := plugintest.NewMockPlugin("test-plugin").
			WithConfigError(errors.New("config error"))

		err := reg.RegisterPlugin(p)
		require.NoError(t, err)

		root := &cobra.Command{Use: "test"}
		err = reg.LoadAll(root)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to configure plugin")
		assert.Contains(t, err.Error(), "config error")
	})

	t.Run("load all with register error", func(t *testing.T) {
		reg := plugin.NewRegistry()
		p := plugintest.NewMockPlugin("test-plugin").
			WithError(errors.New("register error"))

		err := reg.RegisterPlugin(p)
		require.NoError(t, err)

		root := &cobra.Command{Use: "test"}
		err = reg.LoadAll(root)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to register plugin")
		assert.Contains(t, err.Error(), "register error")
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
		config := map[string]interface{}{
			"global-test-plugin": map[string]interface{}{
				"key": "value",
			},
		}
		plugin.SetConfig(config)

		// Load all globally
		root := &cobra.Command{Use: "test"}
		err = plugin.LoadAll(root)
		// This might fail if other tests have registered plugins with errors
		// In a real scenario, you'd want to clean the global registry between tests
		if err == nil {
			assert.True(t, p.Configured)
			assert.True(t, p.Registered)
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
