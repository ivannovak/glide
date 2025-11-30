// Package config provides type-safe configuration management for Glide.
//
// This package enables type-safe configuration handling through Go generics,
// eliminating runtime type assertions and map[string]interface{} conversions.
// It supports schema validation, configuration migration, and hierarchical
// configuration merging.
//
// # Type-Safe Configuration
//
// Define configuration structs with JSON/YAML tags:
//
//	type MyPluginConfig struct {
//	    APIKey    string `json:"api_key" yaml:"api_key"`
//	    Timeout   int    `json:"timeout" yaml:"timeout"`
//	    EnableLog bool   `json:"enable_log" yaml:"enable_log"`
//	}
//
//	// Register with defaults
//	config.Register("my-plugin", MyPluginConfig{
//	    Timeout:   30,
//	    EnableLog: false,
//	})
//
//	// Retrieve typed config - no type assertions needed
//	cfg, err := config.Get[MyPluginConfig]("my-plugin")
//	client := NewClient(cfg.APIKey, cfg.Timeout)
//
// # Schema Validation
//
// Define JSON Schema for configuration validation:
//
//	schema := config.ConfigSchema{
//	    Type: "object",
//	    Properties: map[string]config.PropertySchema{
//	        "api_key": {Type: "string", MinLength: ptrTo(10)},
//	        "timeout": {Type: "integer", Minimum: ptrTo(1.0)},
//	    },
//	    Required: []string{"api_key"},
//	}
//
//	tc := config.NewTypedConfig("my-plugin", defaults)
//	tc.SetSchema(schema)
//	err := tc.Validate()
//
// # Configuration Migration
//
// Handle breaking changes with versioned migrations:
//
//	migrations := []config.Migration{
//	    {
//	        FromVersion: 1,
//	        ToVersion:   2,
//	        Transform: func(data map[string]interface{}) map[string]interface{} {
//	            // Rename "old_field" to "new_field"
//	            data["new_field"] = data["old_field"]
//	            delete(data, "old_field")
//	            return data
//	        },
//	    },
//	}
//
// # Global Registry
//
// Use the global registry for application-wide configuration:
//
//	config.GlobalRegistry.Register("my-plugin", cfg)
//	retrieved, err := config.GlobalRegistry.Get[MyPluginConfig]("my-plugin")
//
// See docs/adr/ADR-003-configuration-management.md for design rationale.
package config
