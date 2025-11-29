package plugintest

import (
	"fmt"
	"reflect"
)

// Package plugintest provides testing utilities for plugin development.
//
// MIGRATION NOTE: The configuration helpers in this package use the deprecated
// map[string]interface{} configuration approach. For new tests, consider using
// the pkg/config type-safe configuration system directly:
//
//   type MyPluginConfig struct {
//       Endpoint string `json:"endpoint"`
//   }
//
//   func TestMyPlugin(t *testing.T) {
//       // Register typed config
//       config.Register("my-plugin", MyPluginConfig{})
//
//       // Update with test data
//       config.Update("my-plugin", map[string]interface{}{
//           "endpoint": "http://test",
//       })
//
//       // Plugin accesses typed config
//       cfg, _ := config.GetValue[MyPluginConfig]("my-plugin")
//   }
//
// See pkg/config/MIGRATION.md for complete migration guide.

// ConfigBuilder provides a fluent interface for building test configurations
type ConfigBuilder struct {
	config map[string]interface{}
}

// NewConfigBuilder creates a new configuration builder
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: make(map[string]interface{}),
	}
}

// WithPlugin adds plugin-specific configuration
func (b *ConfigBuilder) WithPlugin(name string, config map[string]interface{}) *ConfigBuilder {
	b.config[name] = config
	return b
}

// WithValue adds a top-level configuration value
func (b *ConfigBuilder) WithValue(key string, value interface{}) *ConfigBuilder {
	b.config[key] = value
	return b
}

// Build returns the built configuration
func (b *ConfigBuilder) Build() map[string]interface{} {
	return b.config
}

// TestConfig represents a test configuration
type TestConfig map[string]interface{}

// NewTestConfig creates a new test configuration
func NewTestConfig() TestConfig {
	return make(TestConfig)
}

// SetPlugin sets plugin-specific configuration
func (c TestConfig) SetPlugin(name string, config map[string]interface{}) TestConfig {
	c[name] = config
	return c
}

// GetPlugin gets plugin-specific configuration
func (c TestConfig) GetPlugin(name string) (map[string]interface{}, bool) {
	val, exists := c[name]
	if !exists {
		return nil, false
	}

	config, ok := val.(map[string]interface{})
	return config, ok
}

// Set sets a configuration value
func (c TestConfig) Set(key string, value interface{}) TestConfig {
	c[key] = value
	return c
}

// Get gets a configuration value
func (c TestConfig) Get(key string) (interface{}, bool) {
	val, exists := c[key]
	return val, exists
}

// Merge merges another configuration into this one
func (c TestConfig) Merge(other map[string]interface{}) TestConfig {
	for k, v := range other {
		c[k] = v
	}
	return c
}

// Clone creates a deep copy of the configuration
func (c TestConfig) Clone() TestConfig {
	clone := make(TestConfig)
	for k, v := range c {
		clone[k] = deepCopyValue(v)
	}
	return clone
}

// deepCopyValue creates a deep copy of a value
func deepCopyValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case map[string]interface{}:
		copy := make(map[string]interface{})
		for k, v := range val {
			copy[k] = deepCopyValue(v)
		}
		return copy
	case []interface{}:
		copy := make([]interface{}, len(val))
		for i, v := range val {
			copy[i] = deepCopyValue(v)
		}
		return copy
	default:
		// For basic types, return as-is
		return val
	}
}

// ConfigValidator validates plugin configurations
type ConfigValidator struct {
	requiredKeys map[string][]string // plugin name -> required keys
}

// NewConfigValidator creates a new configuration validator
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{
		requiredKeys: make(map[string][]string),
	}
}

// RequireKeys sets required keys for a plugin
func (v *ConfigValidator) RequireKeys(pluginName string, keys ...string) *ConfigValidator {
	v.requiredKeys[pluginName] = keys
	return v
}

// Validate validates a configuration
func (v *ConfigValidator) Validate(config map[string]interface{}) error {
	for pluginName, requiredKeys := range v.requiredKeys {
		pluginConfig, exists := config[pluginName]
		if !exists {
			return fmt.Errorf("missing configuration for plugin: %s", pluginName)
		}

		pluginMap, ok := pluginConfig.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid configuration type for plugin %s: expected map, got %T", pluginName, pluginConfig)
		}

		for _, key := range requiredKeys {
			if _, exists := pluginMap[key]; !exists {
				return fmt.Errorf("missing required key '%s' in %s configuration", key, pluginName)
			}
		}
	}

	return nil
}

// ValidateKeys checks if all required keys are present
func ValidateKeys(config map[string]interface{}, requiredKeys ...string) error {
	for _, key := range requiredKeys {
		if _, exists := config[key]; !exists {
			return fmt.Errorf("missing required key: %s", key)
		}
	}
	return nil
}

// ValidateType checks if a value is of the expected type
func ValidateType(value interface{}, expectedType reflect.Type) error {
	actualType := reflect.TypeOf(value)
	if actualType != expectedType {
		return fmt.Errorf("type mismatch: expected %v, got %v", expectedType, actualType)
	}
	return nil
}

// DefaultConfigs provides common test configurations
var DefaultConfigs = struct {
	Empty    TestConfig
	Basic    TestConfig
	Complete TestConfig
}{
	Empty: NewTestConfig(),

	Basic: NewTestConfig().
		Set("debug", false).
		Set("verbose", false),

	Complete: NewTestConfig().
		Set("debug", true).
		Set("verbose", true).
		SetPlugin("test", map[string]interface{}{
			"enabled": true,
			"timeout": 30,
			"retries": 3,
		}),
}

// MockConfigSource provides a mock configuration source for testing
type MockConfigSource struct {
	configs map[string]TestConfig
	current string
}

// NewMockConfigSource creates a new mock configuration source
func NewMockConfigSource() *MockConfigSource {
	return &MockConfigSource{
		configs: make(map[string]TestConfig),
		current: "default",
	}
}

// AddConfig adds a named configuration
func (m *MockConfigSource) AddConfig(name string, config TestConfig) *MockConfigSource {
	m.configs[name] = config
	return m
}

// UseConfig switches to a named configuration
func (m *MockConfigSource) UseConfig(name string) error {
	if _, exists := m.configs[name]; !exists {
		return fmt.Errorf("configuration '%s' not found", name)
	}
	m.current = name
	return nil
}

// GetConfig returns the current configuration
func (m *MockConfigSource) GetConfig() TestConfig {
	if config, exists := m.configs[m.current]; exists {
		return config
	}
	return NewTestConfig()
}
