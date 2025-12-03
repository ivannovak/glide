package plugintest

import (
	"github.com/glide-cli/glide/v3/pkg/plugin"
	"github.com/spf13/cobra"
)

// MockPlugin implements the Plugin interface for testing
type MockPlugin struct {
	NameValue     string
	VersionValue  string
	Configured    bool
	Registered    bool
	ConfigError   error
	RegisterError error
	RegisterFunc  func(*cobra.Command) error // Allow overriding for tests
	ConfigureFunc func() error               // Allow overriding for tests
	MetadataValue plugin.PluginMetadata
}

// NewMockPlugin creates a new mock plugin with sensible defaults
func NewMockPlugin(name string) *MockPlugin {
	return &MockPlugin{
		NameValue:    name,
		VersionValue: "1.0.0",
		MetadataValue: plugin.PluginMetadata{
			Name:        name,
			Version:     "1.0.0",
			Author:      "Test Author",
			Description: "Test plugin for " + name,
			Commands: []plugin.CommandInfo{
				{
					Name:        "test-" + name,
					Category:    "test",
					Description: "Test command",
				},
			},
			BuildTags:  []string{"test"},
			ConfigKeys: []string{"test_key"},
		},
	}
}

// Name returns the plugin identifier
func (m *MockPlugin) Name() string {
	return m.NameValue
}

// Version returns the plugin version
func (m *MockPlugin) Version() string {
	return m.VersionValue
}

// Configure allows plugin-specific configuration
func (m *MockPlugin) Configure() error {
	m.Configured = true

	if m.ConfigureFunc != nil {
		return m.ConfigureFunc()
	}

	return m.ConfigError
}

// Register adds plugin commands to the command tree
func (m *MockPlugin) Register(root *cobra.Command) error {
	// Allow custom Register function for tests
	if m.RegisterFunc != nil {
		return m.RegisterFunc(root)
	}

	// Default implementation
	m.Registered = true
	if m.RegisterError != nil {
		return m.RegisterError
	}

	// Add a test command
	root.AddCommand(&cobra.Command{
		Use:   "test-" + m.NameValue,
		Short: "Test command from " + m.NameValue,
	})
	return nil
}

// Metadata returns plugin information
func (m *MockPlugin) Metadata() plugin.PluginMetadata {
	return m.MetadataValue
}

// WithError configures the mock to return an error
func (m *MockPlugin) WithError(err error) *MockPlugin {
	m.RegisterError = err
	return m
}

// WithConfigError configures the mock to return a configuration error
func (m *MockPlugin) WithConfigError(err error) *MockPlugin {
	m.ConfigError = err
	return m
}

// WithMetadata sets custom metadata
func (m *MockPlugin) WithMetadata(meta plugin.PluginMetadata) *MockPlugin {
	m.MetadataValue = meta
	return m
}

// Reset resets the mock state
func (m *MockPlugin) Reset() {
	m.Configured = false
	m.Registered = false
}
