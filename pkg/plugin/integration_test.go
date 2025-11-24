package plugin_test

import (
	"testing"

	"github.com/ivannovak/glide/v2/pkg/plugin"
	"github.com/ivannovak/glide/v2/pkg/plugin/plugintest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPlugin simulates a real plugin
type TestPlugin struct {
	commands []string
}

func (t *TestPlugin) Name() string {
	return "integration-test"
}

func (t *TestPlugin) Version() string {
	return "1.0.0-test"
}

func (t *TestPlugin) Configure(config map[string]interface{}) error {
	// Simulate reading configuration
	if config != nil {
		if testConfig, ok := config["integration-test"]; ok {
			// Process configuration
			_ = testConfig
		}
	}
	return nil
}

func (t *TestPlugin) Register(root *cobra.Command) error {
	// Add test commands
	t.commands = []string{"test-cmd1", "test-cmd2", "test-cmd3"}

	for _, cmdName := range t.commands {
		cmd := &cobra.Command{
			Use:   cmdName,
			Short: "Test command " + cmdName,
			RunE: func(cmd *cobra.Command, args []string) error {
				// Command implementation
				return nil
			},
		}
		root.AddCommand(cmd)
	}

	return nil
}

func (t *TestPlugin) Metadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name:        t.Name(),
		Version:     t.Version(),
		Author:      "Integration Test",
		Description: "Plugin for integration testing",
		Commands: []plugin.CommandInfo{
			{Name: "test-cmd1", Category: "test", Description: "Test command 1"},
			{Name: "test-cmd2", Category: "test", Description: "Test command 2"},
			{Name: "test-cmd3", Category: "test", Description: "Test command 3"},
		},
		BuildTags:  []string{"test"},
		ConfigKeys: []string{"enabled"},
	}
}

func TestPluginIntegration(t *testing.T) {
	t.Run("full plugin lifecycle", func(t *testing.T) {
		harness := plugintest.NewTestHarness(t)

		// Create and register a test plugin
		testPlugin := &TestPlugin{}
		err := harness.RegisterPlugin(testPlugin)
		require.NoError(t, err)

		// Verify plugin is registered
		harness.AssertPluginRegistered("integration-test")

		// Verify commands were added
		harness.AssertCommandExists("test-cmd1")
		harness.AssertCommandExists("test-cmd2")
		harness.AssertCommandExists("test-cmd3")

		// Execute a command
		output, err := harness.ExecuteCommand("test-cmd1")
		require.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("plugin with configuration", func(t *testing.T) {
		harness := plugintest.NewTestHarness(t)

		// Set up configuration
		config := plugintest.NewConfigBuilder().
			WithPlugin("integration-test", map[string]interface{}{
				"enabled": true,
				"debug":   false,
			}).
			Build()

		harness.WithConfig(config)

		// Register plugin
		testPlugin := &TestPlugin{}
		err := harness.RegisterPlugin(testPlugin)
		require.NoError(t, err)

		// Plugin should be configured
		harness.AssertPluginRegistered("integration-test")
	})

	t.Run("multiple plugins", func(t *testing.T) {
		harness := plugintest.NewTestHarness(t)

		// Register multiple plugins
		plugin1 := &TestPlugin{commands: []string{"p1-cmd1", "p1-cmd2"}}
		plugin2 := plugintest.NewMockPlugin("plugin2")
		plugin2.VersionValue = "2.0.0"
		plugin3 := plugintest.NewMockPlugin("plugin3")
		plugin3.VersionValue = "3.0.0"

		require.NoError(t, harness.RegisterPlugin(plugin1))
		require.NoError(t, harness.RegisterPlugin(plugin2))
		require.NoError(t, harness.RegisterPlugin(plugin3))

		// Verify all plugins are registered
		plugins := harness.ListPlugins()
		assert.Len(t, plugins, 3)

		// Verify each plugin
		harness.AssertPluginRegistered("integration-test")
		harness.AssertPluginRegistered("plugin2")
		harness.AssertPluginRegistered("plugin3")

		// Verify mock plugins were configured
		assert.True(t, plugin2.Configured)
		assert.True(t, plugin3.Configured)
	})

	t.Run("plugin command execution", func(t *testing.T) {
		harness := plugintest.NewTestHarness(t)

		// Create a mock plugin with custom execute function
		executed := false
		p := plugintest.NewMockPlugin("exec-test")
		p.RegisterFunc = func(root *cobra.Command) error {
			cmd := &cobra.Command{
				Use:   "exec-test-cmd",
				Short: "Executable test command",
				RunE: func(cmd *cobra.Command, args []string) error {
					executed = true
					return nil
				},
			}
			root.AddCommand(cmd)
			p.Registered = true
			return nil
		}

		require.NoError(t, harness.RegisterPlugin(p))

		// Execute the command
		_, err := harness.ExecuteCommand("exec-test-cmd")
		require.NoError(t, err)
		assert.True(t, executed, "Command should have been executed")
	})
}

func TestPluginDiscovery(t *testing.T) {
	t.Run("discover plugin metadata using fixtures", func(t *testing.T) {
		harness := plugintest.NewTestHarness(t)
		fixtures := plugintest.NewFixtures()

		// Register plugins with different metadata
		plugins := []plugin.Plugin{
			fixtures.SimplePlugin("database-plugin"),
			fixtures.ComplexPlugin("docker-plugin"),
			&TestPlugin{},
		}

		for _, p := range plugins {
			require.NoError(t, harness.RegisterPlugin(p))
		}

		// Discover all plugins
		allPlugins := harness.ListPlugins()
		assert.Len(t, allPlugins, 3)

		// Check metadata for each
		for _, p := range allPlugins {
			meta := p.Metadata()
			assert.NotEmpty(t, meta.Name)
			assert.NotEmpty(t, meta.Version)
			assert.NotEmpty(t, meta.Description)

			// TestPlugin has different structure
			if meta.Name == "integration-test" {
				assert.Len(t, meta.Commands, 3)
			}
		}
	})

	t.Run("plugin filtering", func(t *testing.T) {
		harness := plugintest.NewTestHarness(t)
		factory := plugintest.NewPluginFactory()

		// Create plugins with different tags
		dbPlugin := factory.Create(
			plugintest.WithName("db-plugin"),
			plugintest.WithBuildTags("database", "production"),
		)

		devPlugin := factory.Create(
			plugintest.WithName("dev-plugin"),
			plugintest.WithBuildTags("development", "test"),
		)

		require.NoError(t, harness.RegisterPlugin(dbPlugin))
		require.NoError(t, harness.RegisterPlugin(devPlugin))

		// Get all plugins
		plugins := harness.ListPlugins()
		assert.Len(t, plugins, 2)

		// Verify build tags
		for _, p := range plugins {
			meta := p.Metadata()
			if meta.Name == "db-plugin" {
				assert.Contains(t, meta.BuildTags, "database")
				assert.Contains(t, meta.BuildTags, "production")
			} else if meta.Name == "dev-plugin" {
				assert.Contains(t, meta.BuildTags, "development")
				assert.Contains(t, meta.BuildTags, "test")
			}
		}
	})
}

func TestPluginErrorHandling(t *testing.T) {
	t.Run("configuration errors", func(t *testing.T) {
		harness := plugintest.NewTestHarness(t)
		fixtures := plugintest.NewFixtures()

		// Create a plugin that returns config error
		p := fixtures.ErrorPlugin("error-test")

		// Try to register - should fail during configuration
		err := harness.RegisterPlugin(p)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "configuration error")
	})

	t.Run("registration errors", func(t *testing.T) {
		harness := plugintest.NewTestHarness(t)

		// Create a plugin that returns registration error
		p := plugintest.NewMockPlugin("reg-error")
		p.ConfigError = nil              // No config error
		p.RegisterError = assert.AnError // But registration error

		// Registration should fail
		err := harness.RegisterPlugin(p)
		assert.Error(t, err)
	})
}

func TestPluginWithTestScenarios(t *testing.T) {
	t.Run("single plugin scenario", func(t *testing.T) {
		harness := plugintest.NewTestHarness(t)

		// Use pre-configured scenario
		plugin, config := plugintest.TestScenarios.SinglePlugin()

		harness.WithConfig(config)
		err := harness.RegisterPlugin(plugin)
		require.NoError(t, err)

		harness.AssertPluginRegistered("test-plugin")
		assert.True(t, plugin.Configured)
		assert.True(t, plugin.Registered)
	})

	t.Run("multiple plugin scenario", func(t *testing.T) {
		harness := plugintest.NewTestHarness(t)

		// Use pre-configured scenario
		plugins, config := plugintest.TestScenarios.MultiplePlugin()

		harness.WithConfig(config)

		for _, p := range plugins {
			err := harness.RegisterPlugin(p)
			require.NoError(t, err)
		}

		// Verify all plugins
		assert.Len(t, harness.ListPlugins(), 3)
		harness.AssertPluginRegistered("plugin-a")
		harness.AssertPluginRegistered("plugin-b")
		harness.AssertPluginRegistered("plugin-c")
	})

	t.Run("error scenario", func(t *testing.T) {
		harness := plugintest.NewTestHarness(t)

		// Use error scenario
		plugin, config := plugintest.TestScenarios.ErrorScenario()

		harness.WithConfig(config)
		err := harness.RegisterPlugin(plugin)

		// Should fail with config error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "configuration error")
	})
}
