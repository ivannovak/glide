package plugin_test

import (
	"testing"

	"github.com/ivannovak/glide/pkg/plugin"
	"github.com/ivannovak/glide/pkg/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPluginInterface(t *testing.T) {
	t.Run("plugin implements interface", func(t *testing.T) {
		p := plugintest.NewMockPlugin("test-plugin")

		// Verify it implements the interface
		var _ plugin.Plugin = p

		assert.Equal(t, "test-plugin", p.Name())
		assert.Equal(t, "1.0.0", p.Version())
	})

	t.Run("plugin metadata using harness", func(t *testing.T) {
		harness := plugintest.NewTestHarness(t)
		p := plugintest.NewMockPlugin("test-plugin")

		err := harness.RegisterPlugin(p)
		require.NoError(t, err)

		meta := p.Metadata()
		assert.Equal(t, "test-plugin", meta.Name)
		assert.Equal(t, "1.0.0", meta.Version)
		assert.Equal(t, "Test Author", meta.Author)
		assert.Len(t, meta.Commands, 1)
		assert.Equal(t, "test-test-plugin", meta.Commands[0].Name)

		harness.AssertPluginRegistered("test-plugin")
	})

	t.Run("plugin configure with harness", func(t *testing.T) {
		harness := plugintest.NewTestHarness(t)
		p := plugintest.NewMockPlugin("test-plugin")

		config := plugintest.NewConfigBuilder().
			WithValue("test_key", "test_value").
			Build()

		harness.WithConfig(config)
		err := harness.RegisterPlugin(p)
		require.NoError(t, err)

		assert.True(t, p.Configured)
		assert.Equal(t, config, p.ReceivedConfig)
	})

	t.Run("plugin register with command helper", func(t *testing.T) {
		harness := plugintest.NewTestHarness(t)
		p := plugintest.NewMockPlugin("test-plugin")

		err := harness.RegisterPlugin(p)
		require.NoError(t, err)
		assert.True(t, p.Registered)

		// Verify command was added using harness assertions
		harness.AssertCommandExists("test-test-plugin")
	})
}

func TestCommandInfo(t *testing.T) {
	t.Run("command info structure using fixtures", func(t *testing.T) {
		fixtures := plugintest.NewFixtures()
		info := fixtures.CommandInfo("test-command", "testing")

		assert.Equal(t, "test-command", info.Name)
		assert.Equal(t, "testing", info.Category)
		assert.Contains(t, info.Description, "test-command")
		assert.Contains(t, info.Description, "testing")
		assert.Len(t, info.Aliases, 1)
		assert.Equal(t, "t", info.Aliases[0])
	})
}

func TestPluginMetadata(t *testing.T) {
	t.Run("metadata structure using fixtures", func(t *testing.T) {
		fixtures := plugintest.NewFixtures()
		meta := fixtures.SampleMetadata("awesome-plugin")

		assert.Equal(t, "awesome-plugin", meta.Name)
		assert.Equal(t, "1.0.0", meta.Version)
		assert.Equal(t, "Test Author", meta.Author)
		assert.Contains(t, meta.Description, "awesome-plugin")
		assert.Len(t, meta.Commands, 1)
		assert.Contains(t, meta.BuildTags, "awesome-plugin")
		assert.Contains(t, meta.ConfigKeys, "awesome-plugin_key")
	})

	t.Run("metadata with assertions helper", func(t *testing.T) {
		assertions := plugintest.NewAssertions(t)
		p := plugintest.NewMockPlugin("test")

		expected := plugin.PluginMetadata{
			Name:        "test",
			Version:     "1.0.0",
			Author:      "Test Author",
			Description: "Test plugin for test",
			Commands: []plugin.CommandInfo{
				{
					Name:        "test-test",
					Category:    "test",
					Description: "Test command",
				},
			},
			BuildTags:  []string{"test"},
			ConfigKeys: []string{"test_key"},
		}

		assertions.AssertPluginMetadata(p, expected)
	})
}
