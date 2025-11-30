package plugintest

//lint:file-ignore SA1019 plugin.Plugin is deprecated but still valid for use

import (
	"strings"
	"testing"

	"github.com/ivannovak/glide/v2/pkg/plugin"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Assertions provides plugin-specific test assertions
type Assertions struct {
	t *testing.T
}

// NewAssertions creates a new assertions helper
func NewAssertions(t *testing.T) *Assertions {
	return &Assertions{t: t}
}

// AssertPluginMetadata verifies plugin metadata
func (a *Assertions) AssertPluginMetadata(p plugin.Plugin, expected plugin.PluginMetadata) {
	actual := p.Metadata()

	assert.Equal(a.t, expected.Name, actual.Name, "Plugin name mismatch")
	assert.Equal(a.t, expected.Version, actual.Version, "Plugin version mismatch")
	assert.Equal(a.t, expected.Author, actual.Author, "Plugin author mismatch")
	assert.Equal(a.t, expected.Description, actual.Description, "Plugin description mismatch")

	// Check commands
	assert.Len(a.t, actual.Commands, len(expected.Commands), "Command count mismatch")
	for i, cmd := range expected.Commands {
		if i < len(actual.Commands) {
			assert.Equal(a.t, cmd.Name, actual.Commands[i].Name, "Command name mismatch at index %d", i)
			assert.Equal(a.t, cmd.Category, actual.Commands[i].Category, "Command category mismatch at index %d", i)
			assert.Equal(a.t, cmd.Description, actual.Commands[i].Description, "Command description mismatch at index %d", i)
		}
	}

	// Check build tags
	assert.ElementsMatch(a.t, expected.BuildTags, actual.BuildTags, "Build tags mismatch")

	// Check config keys
	assert.ElementsMatch(a.t, expected.ConfigKeys, actual.ConfigKeys, "Config keys mismatch")
}

// AssertPluginRegistered verifies a plugin is properly registered
func (a *Assertions) AssertPluginRegistered(registry *plugin.Registry, name string) {
	p, exists := registry.Get(name)
	assert.True(a.t, exists, "Plugin %s should be registered", name)
	assert.NotNil(a.t, p, "Plugin %s should not be nil", name)
	assert.Equal(a.t, name, p.Name(), "Plugin name should match")
}

// AssertPluginNotRegistered verifies a plugin is not registered
func (a *Assertions) AssertPluginNotRegistered(registry *plugin.Registry, name string) {
	_, exists := registry.Get(name)
	assert.False(a.t, exists, "Plugin %s should not be registered", name)
}

// AssertCommandStructure verifies command structure
func (a *Assertions) AssertCommandStructure(cmd *cobra.Command, expectedUse, expectedShort string) {
	assert.Equal(a.t, expectedUse, cmd.Use, "Command Use field mismatch")
	assert.Equal(a.t, expectedShort, cmd.Short, "Command Short description mismatch")
}

// AssertCommandHasSubcommand verifies a command has a specific subcommand
func (a *Assertions) AssertCommandHasSubcommand(parent *cobra.Command, subcommandName string) {
	found := false
	for _, cmd := range parent.Commands() {
		if cmd.Name() == subcommandName {
			found = true
			break
		}
	}
	assert.True(a.t, found, "Command %s should have subcommand %s", parent.Name(), subcommandName)
}

// AssertCommandTree verifies the entire command tree structure
func (a *Assertions) AssertCommandTree(root *cobra.Command, expectedPaths ...[]string) {
	for _, path := range expectedPaths {
		cmd, _, err := root.Find(path)
		assert.NoError(a.t, err, "Should find command path %v", path)
		assert.NotNil(a.t, cmd, "Command at path %v should not be nil", path)
	}
}

// AssertErrorContains verifies an error contains expected text
func (a *Assertions) AssertErrorContains(err error, expected string) {
	require.Error(a.t, err, "Expected an error")
	assert.Contains(a.t, err.Error(), expected, "Error message should contain expected text")
}

// AssertNoError verifies no error occurred
func (a *Assertions) AssertNoError(err error) {
	assert.NoError(a.t, err, "Expected no error")
}

// AssertOutputContains verifies output contains expected text
func (a *Assertions) AssertOutputContains(output, expected string) {
	assert.Contains(a.t, output, expected, "Output should contain expected text")
}

// AssertOutputNotContains verifies output does not contain text
func (a *Assertions) AssertOutputNotContains(output, unexpected string) {
	assert.NotContains(a.t, output, unexpected, "Output should not contain unexpected text")
}

// AssertOutputLines verifies output has expected number of lines
func (a *Assertions) AssertOutputLines(output string, expectedLines int) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Len(a.t, lines, expectedLines, "Output should have %d lines", expectedLines)
}

// AssertPluginConfigured verifies a plugin was configured
func (a *Assertions) AssertPluginConfigured(p interface{}) {
	if mock, ok := p.(*MockPlugin); ok {
		assert.True(a.t, mock.Configured, "Plugin should be configured")
	} else {
		a.t.Errorf("AssertPluginConfigured requires a MockPlugin, got %T", p)
	}
}

// AssertPluginRegisteredCommands verifies a plugin registered its commands
func (a *Assertions) AssertPluginRegisteredCommands(p interface{}) {
	if mock, ok := p.(*MockPlugin); ok {
		assert.True(a.t, mock.Registered, "Plugin should have registered commands")
	} else {
		a.t.Errorf("AssertPluginRegisteredCommands requires a MockPlugin, got %T", p)
	}
}

// AssertConfigApplied verifies configuration was applied to a plugin.
// Note: With the type-safe configuration system, this only checks that Configure() was called.
// Plugin configuration is now handled via pkg/config.Register/Get, not via Configure parameter.
func (a *Assertions) AssertConfigApplied(p interface{}, _ map[string]interface{}) {
	if mock, ok := p.(*MockPlugin); ok {
		assert.True(a.t, mock.Configured, "Plugin should be configured")
		// Configuration is now type-safe via pkg/config, not passed to Configure()
	} else {
		a.t.Errorf("AssertConfigApplied requires a MockPlugin, got %T", p)
	}
}

// AssertFlagValue verifies a command flag has expected value
func (a *Assertions) AssertFlagValue(cmd *cobra.Command, flagName, expectedValue string) {
	flag := cmd.Flags().Lookup(flagName)
	require.NotNil(a.t, flag, "Flag %s should exist", flagName)
	assert.Equal(a.t, expectedValue, flag.Value.String(), "Flag %s should have value %s", flagName, expectedValue)
}

// AssertFlagExists verifies a command has a specific flag
func (a *Assertions) AssertFlagExists(cmd *cobra.Command, flagName string) {
	flag := cmd.Flags().Lookup(flagName)
	assert.NotNil(a.t, flag, "Flag %s should exist", flagName)
}

// AssertFlagNotExists verifies a command does not have a specific flag
func (a *Assertions) AssertFlagNotExists(cmd *cobra.Command, flagName string) {
	flag := cmd.Flags().Lookup(flagName)
	assert.Nil(a.t, flag, "Flag %s should not exist", flagName)
}

// QuickAssert provides a fluent interface for multiple assertions
type QuickAssert struct {
	t *testing.T
}

// NewQuickAssert creates a new quick assert helper
func NewQuickAssert(t *testing.T) *QuickAssert {
	return &QuickAssert{t: t}
}

// Plugin asserts on a plugin
func (q *QuickAssert) Plugin(p plugin.Plugin) *PluginAssert {
	return &PluginAssert{t: q.t, plugin: p}
}

// Command asserts on a command
func (q *QuickAssert) Command(cmd *cobra.Command) *CommandAssert {
	return &CommandAssert{t: q.t, cmd: cmd}
}

// PluginAssert provides fluent assertions for plugins
type PluginAssert struct {
	t      *testing.T
	plugin plugin.Plugin
}

// HasName asserts plugin has expected name
func (p *PluginAssert) HasName(name string) *PluginAssert {
	assert.Equal(p.t, name, p.plugin.Name())
	return p
}

// HasVersion asserts plugin has expected version
func (p *PluginAssert) HasVersion(version string) *PluginAssert {
	assert.Equal(p.t, version, p.plugin.Version())
	return p
}

// HasCommands asserts plugin has expected number of commands
func (p *PluginAssert) HasCommands(count int) *PluginAssert {
	meta := p.plugin.Metadata()
	assert.Len(p.t, meta.Commands, count)
	return p
}

// CommandAssert provides fluent assertions for commands
type CommandAssert struct {
	t   *testing.T
	cmd *cobra.Command
}

// HasUse asserts command has expected use
func (c *CommandAssert) HasUse(use string) *CommandAssert {
	assert.Equal(c.t, use, c.cmd.Use)
	return c
}

// HasShort asserts command has expected short description
func (c *CommandAssert) HasShort(short string) *CommandAssert {
	assert.Equal(c.t, short, c.cmd.Short)
	return c
}

// HasSubcommands asserts command has expected number of subcommands
func (c *CommandAssert) HasSubcommands(count int) *CommandAssert {
	assert.Len(c.t, c.cmd.Commands(), count)
	return c
}
