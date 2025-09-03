package plugintest

import (
	"fmt"

	"github.com/ivannovak/glide/pkg/plugin"
	"github.com/spf13/cobra"
)

// Fixtures provides test data generators for plugin testing
type Fixtures struct{}

// NewFixtures creates a new fixtures helper
func NewFixtures() *Fixtures {
	return &Fixtures{}
}

// SimplePlugin creates a simple test plugin
func (f *Fixtures) SimplePlugin(name string) *MockPlugin {
	return NewMockPlugin(name)
}

// ComplexPlugin creates a plugin with multiple commands
func (f *Fixtures) ComplexPlugin(name string) *MockPlugin {
	p := NewMockPlugin(name)
	p.MetadataValue = plugin.PluginMetadata{
		Name:        name,
		Version:     "2.0.0",
		Author:      "Test Team",
		Description: fmt.Sprintf("Complex test plugin %s", name),
		Commands: []plugin.CommandInfo{
			{Name: "list", Category: "data", Description: "List items"},
			{Name: "get", Category: "data", Description: "Get item"},
			{Name: "create", Category: "data", Description: "Create item"},
			{Name: "update", Category: "data", Description: "Update item"},
			{Name: "delete", Category: "data", Description: "Delete item"},
		},
		BuildTags:  []string{name, "test"},
		ConfigKeys: []string{"api_key", "endpoint", "timeout", "retries"},
	}
	
	// Custom register function that adds actual commands
	p.RegisterFunc = func(root *cobra.Command) error {
		p.Registered = true
		
		// Create a parent command
		parentCmd := &cobra.Command{
			Use:   name,
			Short: fmt.Sprintf("%s plugin commands", name),
		}
		
		// Add subcommands
		parentCmd.AddCommand(
			&cobra.Command{Use: "list", Short: "List items"},
			&cobra.Command{Use: "get", Short: "Get item"},
			&cobra.Command{Use: "create", Short: "Create item"},
			&cobra.Command{Use: "update", Short: "Update item"},
			&cobra.Command{Use: "delete", Short: "Delete item"},
		)
		
		root.AddCommand(parentCmd)
		return nil
	}
	
	return p
}

// ErrorPlugin creates a plugin that returns errors
func (f *Fixtures) ErrorPlugin(name string) *MockPlugin {
	p := NewMockPlugin(name)
	p.ConfigError = fmt.Errorf("configuration error for %s", name)
	p.RegisterError = fmt.Errorf("registration error for %s", name)
	return p
}

// SampleConfig generates a sample configuration
func (f *Fixtures) SampleConfig(pluginName string) map[string]interface{} {
	return map[string]interface{}{
		pluginName: map[string]interface{}{
			"enabled":  true,
			"api_key":  "test-api-key-123",
			"endpoint": "https://api.example.com",
			"timeout":  30,
			"retries":  3,
			"features": map[string]bool{
				"caching":     true,
				"compression": false,
				"logging":     true,
			},
		},
	}
}

// MultiPluginConfig generates configuration for multiple plugins
func (f *Fixtures) MultiPluginConfig(pluginNames ...string) map[string]interface{} {
	config := make(map[string]interface{})
	for i, name := range pluginNames {
		config[name] = map[string]interface{}{
			"enabled": true,
			"port":    8080 + i,
			"host":    fmt.Sprintf("%s.example.com", name),
		}
	}
	return config
}

// SampleMetadata generates sample plugin metadata
func (f *Fixtures) SampleMetadata(name string) plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name:        name,
		Version:     "1.0.0",
		Author:      "Test Author",
		Description: fmt.Sprintf("Sample plugin %s for testing", name),
		Commands: []plugin.CommandInfo{
			{
				Name:        fmt.Sprintf("%s-cmd", name),
				Category:    "test",
				Description: fmt.Sprintf("Test command for %s", name),
				Aliases:     []string{name[:1]},
			},
		},
		BuildTags:  []string{name},
		ConfigKeys: []string{fmt.Sprintf("%s_key", name), fmt.Sprintf("%s_secret", name)},
	}
}

// CommandInfo generates sample command info
func (f *Fixtures) CommandInfo(name, category string) plugin.CommandInfo {
	return plugin.CommandInfo{
		Name:        name,
		Category:    category,
		Description: fmt.Sprintf("Test %s command in %s category", name, category),
		Aliases:     []string{name[:1]},
	}
}

// PluginSet creates a set of related plugins
func (f *Fixtures) PluginSet(prefix string, count int) []*MockPlugin {
	plugins := make([]*MockPlugin, count)
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("%s-%d", prefix, i+1)
		plugins[i] = f.SimplePlugin(name)
	}
	return plugins
}

// TestScenarios provides pre-configured test scenarios
var TestScenarios = struct {
	SinglePlugin   func() (*MockPlugin, map[string]interface{})
	MultiplePlugin func() ([]*MockPlugin, map[string]interface{})
	ErrorScenario  func() (*MockPlugin, map[string]interface{})
}{
	SinglePlugin: func() (*MockPlugin, map[string]interface{}) {
		f := NewFixtures()
		p := f.SimplePlugin("test-plugin")
		config := f.SampleConfig("test-plugin")
		return p, config
	},
	
	MultiplePlugin: func() ([]*MockPlugin, map[string]interface{}) {
		f := NewFixtures()
		plugins := []*MockPlugin{
			f.SimplePlugin("plugin-a"),
			f.ComplexPlugin("plugin-b"),
			f.SimplePlugin("plugin-c"),
		}
		config := f.MultiPluginConfig("plugin-a", "plugin-b", "plugin-c")
		return plugins, config
	},
	
	ErrorScenario: func() (*MockPlugin, map[string]interface{}) {
		f := NewFixtures()
		p := f.ErrorPlugin("error-plugin")
		config := f.SampleConfig("error-plugin")
		return p, config
	},
}

// PluginFactory provides factory methods for creating test plugins
type PluginFactory struct {
	fixtures *Fixtures
}

// NewPluginFactory creates a new plugin factory
func NewPluginFactory() *PluginFactory {
	return &PluginFactory{
		fixtures: NewFixtures(),
	}
}

// Create creates a plugin with the given options
func (f *PluginFactory) Create(opts ...PluginOption) *MockPlugin {
	p := NewMockPlugin("factory-plugin")
	
	for _, opt := range opts {
		opt(p)
	}
	
	return p
}

// PluginOption configures a mock plugin
type PluginOption func(*MockPlugin)

// WithName sets the plugin name
func WithName(name string) PluginOption {
	return func(p *MockPlugin) {
		p.NameValue = name
		p.MetadataValue.Name = name
	}
}

// WithVersion sets the plugin version
func WithVersion(version string) PluginOption {
	return func(p *MockPlugin) {
		p.VersionValue = version
		p.MetadataValue.Version = version
	}
}

// WithCommands sets the plugin commands
func WithCommands(commands ...plugin.CommandInfo) PluginOption {
	return func(p *MockPlugin) {
		p.MetadataValue.Commands = commands
	}
}

// WithConfigError sets a configuration error
func WithConfigError(err error) PluginOption {
	return func(p *MockPlugin) {
		p.ConfigError = err
	}
}

// WithRegisterError sets a registration error
func WithRegisterError(err error) PluginOption {
	return func(p *MockPlugin) {
		p.RegisterError = err
	}
}

// WithBuildTags sets build tags
func WithBuildTags(tags ...string) PluginOption {
	return func(p *MockPlugin) {
		p.MetadataValue.BuildTags = tags
	}
}

// WithConfigKeys sets config keys
func WithConfigKeys(keys ...string) PluginOption {
	return func(p *MockPlugin) {
		p.MetadataValue.ConfigKeys = keys
	}
}