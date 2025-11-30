// Package config provides type-safe configuration management for Glide plugins and core systems.
//
// This package enables type-safe configuration handling through Go generics, eliminating the need
// for runtime type assertions and map[string]interface{} conversions.
//
// Example usage:
//
//	type MyPluginConfig struct {
//	    APIKey    string `json:"api_key" yaml:"api_key"`
//	    Timeout   int    `json:"timeout" yaml:"timeout"`
//	    EnableLog bool   `json:"enable_log" yaml:"enable_log"`
//	}
//
//	// Register config with defaults
//	config.Register("my-plugin", MyPluginConfig{
//	    Timeout:   30,
//	    EnableLog: false,
//	})
//
//	// Retrieve typed config
//	cfg, err := config.Get[MyPluginConfig]("my-plugin")
//	if err != nil {
//	    return err
//	}
//	// cfg is fully typed - no type assertions needed!
//	client := NewClient(cfg.APIKey, cfg.Timeout)
package config

import (
	"encoding/json"
	"fmt"
	"reflect"

	"gopkg.in/yaml.v3"
)

// TypedConfig wraps a configuration value with type information and metadata.
// This provides compile-time type safety while maintaining runtime flexibility
// for serialization and validation.
type TypedConfig[T any] struct {
	// Value holds the actual configuration data
	Value T

	// Name identifies this configuration (e.g., plugin name)
	Name string

	// Version allows for config schema versioning and migrations
	Version int

	// Defaults holds the default values for this configuration
	Defaults T

	// Schema holds the JSON Schema definition for validation (optional)
	Schema ConfigSchema
}

// NewTypedConfig creates a new typed configuration with the given defaults.
//
// Parameters:
//   - name: Unique identifier for this configuration
//   - defaults: Default values to use when config is missing or incomplete
//
// Returns a TypedConfig with Value initialized to defaults.
func NewTypedConfig[T any](name string, defaults T) *TypedConfig[T] {
	return &TypedConfig[T]{
		Name:     name,
		Value:    defaults,
		Defaults: defaults,
		Version:  1, // Start at version 1
	}
}

// Merge merges raw configuration data into this typed config.
// This is used when loading config from YAML/JSON files.
//
// The merge process:
//  1. If rawConfig is nil, keeps current Value (which is Defaults on first load)
//  2. If rawConfig is already type T, uses it directly
//  3. Otherwise, attempts to unmarshal via JSON round-trip
//
// Parameters:
//   - rawConfig: Raw configuration data (typically from map[string]interface{})
//
// Returns an error if the merge fails (e.g., type mismatch, invalid data).
func (tc *TypedConfig[T]) Merge(rawConfig interface{}) error {
	if rawConfig == nil {
		// No config provided, keep defaults
		return nil
	}

	// Try direct type assertion first
	if typed, ok := rawConfig.(T); ok {
		tc.Value = typed
		return nil
	}

	// Fall back to JSON round-trip for type conversion
	// This handles map[string]interface{} -> struct conversions
	jsonData, err := json.Marshal(rawConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal raw config for %s: %w", tc.Name, err)
	}

	var value T
	if err := json.Unmarshal(jsonData, &value); err != nil {
		return fmt.Errorf("failed to unmarshal config for %s: %w", tc.Name, err)
	}

	tc.Value = value
	return nil
}

// MergeYAML merges YAML configuration data into this typed config.
//
// Parameters:
//   - yamlData: Raw YAML bytes
//
// Returns an error if the YAML is invalid or doesn't match the expected type.
func (tc *TypedConfig[T]) MergeYAML(yamlData []byte) error {
	var value T
	if err := yaml.Unmarshal(yamlData, &value); err != nil {
		return fmt.Errorf("failed to unmarshal YAML config for %s: %w", tc.Name, err)
	}

	tc.Value = value
	return nil
}

// MergeJSON merges JSON configuration data into this typed config.
//
// Parameters:
//   - jsonData: Raw JSON bytes
//
// Returns an error if the JSON is invalid or doesn't match the expected type.
func (tc *TypedConfig[T]) MergeJSON(jsonData []byte) error {
	var value T
	if err := json.Unmarshal(jsonData, &value); err != nil {
		return fmt.Errorf("failed to unmarshal JSON config for %s: %w", tc.Name, err)
	}

	tc.Value = value
	return nil
}

// Reset resets the configuration value to its defaults.
func (tc *TypedConfig[T]) Reset() {
	tc.Value = tc.Defaults
}

// Clone creates a deep copy of this TypedConfig.
// This is useful when you need to modify a config without affecting the original.
func (tc *TypedConfig[T]) Clone() (*TypedConfig[T], error) {
	// Use JSON round-trip for deep copy
	jsonData, err := json.Marshal(tc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config for cloning: %w", err)
	}

	var clone TypedConfig[T]
	if err := json.Unmarshal(jsonData, &clone); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config clone: %w", err)
	}

	return &clone, nil
}

// TypeName returns the Go type name of the configuration value.
// This is useful for debugging and error messages.
func (tc *TypedConfig[T]) TypeName() string {
	var zero T
	t := reflect.TypeOf(zero)
	if t == nil {
		return "unknown"
	}
	return t.String()
}

// Validate validates the configuration value against its schema.
// Returns nil if no schema is defined.
//
// Returns an error if validation fails.
func (tc *TypedConfig[T]) Validate() error {
	if tc.Schema == nil {
		// No schema defined, skip validation
		return nil
	}

	return tc.Schema.Validate(tc.Value)
}

// MarshalYAML implements yaml.Marshaler to customize YAML output.
// This allows TypedConfig to be serialized directly to YAML files.
func (tc *TypedConfig[T]) MarshalYAML() (interface{}, error) {
	return tc.Value, nil
}

// UnmarshalYAML implements yaml.Unmarshaler to customize YAML input.
// This allows TypedConfig to be deserialized directly from YAML files.
func (tc *TypedConfig[T]) UnmarshalYAML(value *yaml.Node) error {
	var v T
	if err := value.Decode(&v); err != nil {
		return fmt.Errorf("failed to decode YAML for %s: %w", tc.Name, err)
	}
	tc.Value = v
	return nil
}

// MarshalJSON implements json.Marshaler to customize JSON output.
// This allows TypedConfig to be serialized directly to JSON.
func (tc *TypedConfig[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(tc.Value)
}

// UnmarshalJSON implements json.Unmarshaler to customize JSON input.
// This allows TypedConfig to be deserialized directly from JSON.
func (tc *TypedConfig[T]) UnmarshalJSON(data []byte) error {
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("failed to decode JSON for %s: %w", tc.Name, err)
	}
	tc.Value = v
	return nil
}
