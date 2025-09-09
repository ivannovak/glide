package plugin

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// Mock plugin with aliases
type MockAliasPlugin struct {
	name    string
	aliases []string
}

func (p *MockAliasPlugin) Name() string {
	return p.name
}

func (p *MockAliasPlugin) Version() string {
	return "1.0.0"
}

func (p *MockAliasPlugin) Register(root *cobra.Command) error {
	cmd := &cobra.Command{
		Use:   p.name,
		Short: "Mock plugin command",
	}
	root.AddCommand(cmd)
	return nil
}

func (p *MockAliasPlugin) Configure(config map[string]interface{}) error {
	return nil
}

func (p *MockAliasPlugin) Metadata() PluginMetadata {
	return PluginMetadata{
		Name:        p.name,
		Version:     "1.0.0",
		Description: "Mock plugin with aliases",
		Aliases:     p.aliases,
	}
}

func TestRegistry_RegisterWithAliases(t *testing.T) {
	registry := NewRegistry()

	// Create plugin with aliases
	plugin := &MockAliasPlugin{
		name:    "database",
		aliases: []string{"db", "d"},
	}

	// Register plugin
	err := registry.RegisterPlugin(plugin)
	assert.NoError(t, err)

	// Verify plugin can be retrieved by name
	p, exists := registry.Get("database")
	assert.True(t, exists)
	assert.NotNil(t, p)
	assert.Equal(t, "database", p.Name())

	// Verify plugin can be retrieved by aliases
	p, exists = registry.Get("db")
	assert.True(t, exists)
	assert.NotNil(t, p)
	assert.Equal(t, "database", p.Name())

	p, exists = registry.Get("d")
	assert.True(t, exists)
	assert.NotNil(t, p)
	assert.Equal(t, "database", p.Name())
}

func TestRegistry_AliasConflicts(t *testing.T) {
	registry := NewRegistry()

	// Register first plugin with alias
	plugin1 := &MockAliasPlugin{
		name:    "database",
		aliases: []string{"db"},
	}
	err := registry.RegisterPlugin(plugin1)
	assert.NoError(t, err)

	// Try to register second plugin with same alias
	plugin2 := &MockAliasPlugin{
		name:    "debug",
		aliases: []string{"db"}, // Conflict!
	}
	err = registry.RegisterPlugin(plugin2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "alias db already registered")

	// Try to register plugin with name that conflicts with existing alias
	plugin3 := &MockAliasPlugin{
		name:    "db", // Conflicts with alias
		aliases: []string{},
	}
	err = registry.RegisterPlugin(plugin3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "item name db conflicts with existing alias")
}

func TestRegistry_ResolveAlias(t *testing.T) {
	registry := NewRegistry()

	plugin := &MockAliasPlugin{
		name:    "composer",
		aliases: []string{"c", "comp"},
	}
	registry.RegisterPlugin(plugin)

	// Test resolving aliases
	canonical, ok := registry.ResolveAlias("c")
	assert.True(t, ok)
	assert.Equal(t, "composer", canonical)

	canonical, ok = registry.ResolveAlias("comp")
	assert.True(t, ok)
	assert.Equal(t, "composer", canonical)

	// Test resolving non-existent alias
	canonical, ok = registry.ResolveAlias("nonexistent")
	assert.False(t, ok)
	assert.Empty(t, canonical)

	// Test that plugin name is not resolved as alias
	canonical, ok = registry.ResolveAlias("composer")
	assert.False(t, ok)
	assert.Empty(t, canonical)
}

func TestRegistry_GetAliases(t *testing.T) {
	registry := NewRegistry()

	plugin := &MockAliasPlugin{
		name:    "terraform",
		aliases: []string{"tf", "terra"},
	}
	registry.RegisterPlugin(plugin)

	// Get aliases for existing plugin
	aliases := registry.GetAliases("terraform")
	assert.Equal(t, []string{"terra", "tf"}, aliases) // Aliases are sorted alphabetically

	// Get aliases for non-existent plugin
	aliases = registry.GetAliases("nonexistent")
	assert.Nil(t, aliases)
}

func TestRegistry_IsAlias(t *testing.T) {
	registry := NewRegistry()

	plugin := &MockAliasPlugin{
		name:    "kubernetes",
		aliases: []string{"k8s", "k"},
	}
	registry.RegisterPlugin(plugin)

	// Test checking aliases
	assert.True(t, registry.IsAlias("k8s"))
	assert.True(t, registry.IsAlias("k"))
	assert.False(t, registry.IsAlias("kubernetes")) // Plugin name, not alias
	assert.False(t, registry.IsAlias("nonexistent"))
}

func TestRegistry_MultiplePluginsWithAliases(t *testing.T) {
	registry := NewRegistry()

	// Register multiple plugins with aliases
	plugins := []*MockAliasPlugin{
		{name: "database", aliases: []string{"db", "d"}},
		{name: "kubernetes", aliases: []string{"k8s", "k"}},
		{name: "terraform", aliases: []string{"tf"}},
	}

	for _, p := range plugins {
		err := registry.RegisterPlugin(p)
		assert.NoError(t, err)
	}

	// Verify all plugins and aliases work
	testCases := []struct {
		query        string
		expectedName string
		shouldExist  bool
	}{
		// Direct names
		{"database", "database", true},
		{"kubernetes", "kubernetes", true},
		{"terraform", "terraform", true},
		// Aliases
		{"db", "database", true},
		{"d", "database", true},
		{"k8s", "kubernetes", true},
		{"k", "kubernetes", true},
		{"tf", "terraform", true},
		// Non-existent
		{"mysql", "", false},
		{"docker", "", false},
	}

	for _, tc := range testCases {
		plugin, exists := registry.Get(tc.query)
		assert.Equal(t, tc.shouldExist, exists, "Query: %s", tc.query)
		if tc.shouldExist {
			assert.Equal(t, tc.expectedName, plugin.Name(), "Query: %s", tc.query)
		}
	}
}

func TestRegistry_LoadAllWithAliases(t *testing.T) {
	registry := NewRegistry()

	// Register plugin with aliases
	plugin := &MockAliasPlugin{
		name:    "example",
		aliases: []string{"ex", "e"},
	}
	err := registry.RegisterPlugin(plugin)
	assert.NoError(t, err)

	// Create root command
	root := &cobra.Command{Use: "glid"}

	// Load all plugins
	err = registry.LoadAll(root)
	assert.NoError(t, err)

	// Verify command was added
	cmd, _, err := root.Find([]string{"example"})
	assert.NoError(t, err)
	assert.NotNil(t, cmd)

	// Note: Plugin aliases don't automatically create command aliases
	// That would be handled by the plugin's Register method if desired
}

// Test global registry functions
func TestGlobalRegistry_WithAliases(t *testing.T) {
	// Clear global registry for test
	globalRegistry = NewRegistry()

	// Register plugin with aliases using global function
	plugin := &MockAliasPlugin{
		name:    "testing",
		aliases: []string{"t", "test"},
	}
	err := Register(plugin)
	assert.NoError(t, err)

	// Test Get with alias
	p, exists := Get("t")
	assert.True(t, exists)
	assert.Equal(t, "testing", p.Name())

	p, exists = Get("test")
	assert.True(t, exists)
	assert.Equal(t, "testing", p.Name())

	// Verify through global registry
	globalReg := GetGlobalRegistry()
	assert.True(t, globalReg.IsAlias("t"))
	assert.True(t, globalReg.IsAlias("test"))

	canonical, ok := globalReg.ResolveAlias("t")
	assert.True(t, ok)
	assert.Equal(t, "testing", canonical)
}
