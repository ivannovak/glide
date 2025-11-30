package config_test

import (
	"fmt"

	"github.com/ivannovak/glide/v2/pkg/config"
)

// ExampleRegister demonstrates basic type-safe configuration usage.
func ExampleRegister() {
	// Define a configuration struct
	type PluginConfig struct {
		APIKey  string `json:"api_key" yaml:"api_key" validate:"required"`
		Timeout int    `json:"timeout" yaml:"timeout" validate:"min=1,max=3600"`
		Enabled bool   `json:"enabled" yaml:"enabled"`
	}

	// Register configuration with defaults
	err := config.Register("example-plugin", PluginConfig{
		Timeout: 30,
		Enabled: false,
	})
	if err != nil {
		panic(err)
	}

	// Retrieve typed configuration
	cfg, err := config.GetValue[PluginConfig]("example-plugin")
	if err != nil {
		panic(err)
	}

	// Access fields with full type safety - no type assertions!
	fmt.Printf("Timeout: %d\n", cfg.Timeout)
	fmt.Printf("Enabled: %v\n", cfg.Enabled)

	// Output:
	// Timeout: 30
	// Enabled: false
}

// ExampleGet demonstrates retrieving TypedConfig with schema access.
func ExampleGet() {
	type MyConfig struct {
		Name    string `json:"name" validate:"required"`
		Count   int    `json:"count" validate:"min=0"`
		Enabled bool   `json:"enabled"`
	}

	// Register (error ignored for example brevity)
	if err := config.Register("my-config", MyConfig{
		Count:   5,
		Enabled: true,
	}); err != nil {
		panic(err)
	}

	// Get TypedConfig wrapper
	tc, err := config.Get[MyConfig]("my-config")
	if err != nil {
		panic(err)
	}

	// Access value
	fmt.Printf("Count: %d\n", tc.Value.Count)
	fmt.Printf("Enabled: %v\n", tc.Value.Enabled)

	// Access metadata
	fmt.Printf("Config Name: %s\n", tc.Name)

	// Output:
	// Count: 5
	// Enabled: true
	// Config Name: my-config
}

// ExampleUpdate demonstrates updating configuration at runtime.
func ExampleUpdate() {
	type Config struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}

	// Register with defaults
	if err := config.Register("server", Config{
		Host: "localhost",
		Port: 8080,
	}); err != nil {
		panic(err)
	}

	// Update with new values
	err := config.Update("server", map[string]interface{}{
		"host": "example.com",
		"port": 9000,
	})
	if err != nil {
		panic(err)
	}

	// Retrieve updated config
	cfg, err := config.GetValue[Config]("server")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Host: %s, Port: %d\n", cfg.Host, cfg.Port)

	// Output:
	// Host: example.com, Port: 9000
}

// ExampleGenerateSchemaFromValue demonstrates JSON Schema generation.
func ExampleGenerateSchemaFromValue() {
	type APIConfig struct {
		Endpoint string `json:"endpoint" validate:"required"`
		Timeout  int    `json:"timeout" validate:"min=1,max=300"`
		Retries  int    `json:"retries" validate:"min=0,max=10"`
	}

	schema, err := config.GenerateSchemaFromValue(APIConfig{})
	if err != nil {
		panic(err)
	}

	// Schema contains validation rules
	fmt.Printf("Schema type: %v\n", schema["type"])
	fmt.Printf("Has properties: %v\n", schema["properties"] != nil)
	fmt.Printf("Has required fields: %v\n", schema["required"] != nil)

	// Output:
	// Schema type: object
	// Has properties: true
	// Has required fields: true
}
