package config

import (
	"fmt"
	"reflect"
	"sync"
)

// Registry manages typed configurations for plugins and core systems.
// It provides thread-safe registration and retrieval of configurations.
//
// The registry is a singleton that should be accessed via the package-level
// Register and Get functions for most use cases.
type Registry struct {
	mu      sync.RWMutex
	configs map[string]interface{} // Stores TypedConfig[T] instances (type erased)
	types   map[string]reflect.Type
}

var (
	// Global registry instance
	globalRegistry = NewRegistry()
)

// NewRegistry creates a new configuration registry.
func NewRegistry() *Registry {
	return &Registry{
		configs: make(map[string]interface{}),
		types:   make(map[string]reflect.Type),
	}
}

// Register registers a typed configuration with default values.
// This is a package-level function that uses the global registry.
//
// Type parameters:
//   - T: The configuration struct type
//
// Parameters:
//   - name: Unique identifier for this configuration (e.g., plugin name)
//   - defaults: Default values for the configuration
//
// Returns an error if a configuration with this name is already registered.
//
// Example:
//
//	type PluginConfig struct {
//	    APIKey  string `json:"api_key"`
//	    Timeout int    `json:"timeout"`
//	}
//
//	err := config.Register("my-plugin", PluginConfig{
//	    Timeout: 30,
//	})
func Register[T any](name string, defaults T) error {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	// Check if already registered
	if _, exists := globalRegistry.configs[name]; exists {
		return fmt.Errorf("configuration %q is already registered", name)
	}

	// Create typed config
	tc := NewTypedConfig(name, defaults)

	// Generate schema from type
	schema := NewJSONSchemaFromValue(defaults)
	tc.Schema = schema

	// Store config and type information
	globalRegistry.configs[name] = tc
	globalRegistry.types[name] = reflect.TypeOf(defaults)

	return nil
}

// Get retrieves a typed configuration by name.
// This is a package-level function that uses the global registry.
//
// Type parameters:
//   - T: The configuration struct type (must match the type used in Register)
//
// Parameters:
//   - name: Unique identifier for this configuration
//
// Returns the typed configuration or an error if not found or type mismatch.
//
// Example:
//
//	cfg, err := config.Get[PluginConfig]("my-plugin")
//	if err != nil {
//	    return err
//	}
//	// cfg is fully typed!
//	fmt.Println(cfg.Value.APIKey)
func Get[T any](name string) (*TypedConfig[T], error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	// Get config from registry
	raw, exists := globalRegistry.configs[name]
	if !exists {
		return nil, fmt.Errorf("configuration %q not found (did you forget to register it?)", name)
	}

	// Type assert to TypedConfig[T]
	tc, ok := raw.(*TypedConfig[T])
	if !ok {
		// Get the actual type for better error message
		actualType := globalRegistry.types[name]
		var zero T
		expectedType := reflect.TypeOf(zero)
		return nil, fmt.Errorf("configuration %q type mismatch: registered as %v, requested as %v",
			name, actualType, expectedType)
	}

	return tc, nil
}

// GetValue retrieves the configuration value directly (without the TypedConfig wrapper).
// This is a convenience function for cases where you just need the value.
//
// Type parameters:
//   - T: The configuration struct type
//
// Parameters:
//   - name: Unique identifier for this configuration
//
// Returns the configuration value or an error if not found or type mismatch.
//
// Example:
//
//	cfg, err := config.GetValue[PluginConfig]("my-plugin")
//	if err != nil {
//	    return err
//	}
//	fmt.Println(cfg.APIKey)  // Direct access to value
func GetValue[T any](name string) (T, error) {
	tc, err := Get[T](name)
	if err != nil {
		var zero T
		return zero, err
	}
	return tc.Value, nil
}

// Update updates a registered configuration with new values.
// This is a package-level convenience function that uses the global registry.
//
// Parameters:
//   - name: Unique identifier for this configuration
//   - rawConfig: Raw configuration data (typically from map[string]interface{} or YAML)
//
// Returns an error if the configuration is not registered or the update fails.
//
// Example:
//
//	rawConfig := map[string]interface{}{
//	    "api_key": "secret123",
//	    "timeout": 60,
//	}
//	err := config.Update("my-plugin", rawConfig)
func Update(name string, rawConfig interface{}) error {
	return globalRegistry.Update(name, rawConfig)
}

// Update updates a registered configuration with new values.
//
// Parameters:
//   - name: Unique identifier for this configuration
//   - rawConfig: Raw configuration data (typically from map[string]interface{} or YAML)
//
// Returns an error if the configuration is not registered or the update fails.
func (r *Registry) Update(name string, rawConfig interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Get config from registry
	raw, exists := r.configs[name]
	if !exists {
		return fmt.Errorf("configuration %q not found", name)
	}

	// We need to call Merge on the TypedConfig, but it's type-erased
	// Use reflection to call the Merge method
	mergeMethod := reflect.ValueOf(raw).MethodByName("Merge")
	if !mergeMethod.IsValid() {
		return fmt.Errorf("configuration %q does not support Merge operation", name)
	}

	// Call Merge method
	results := mergeMethod.Call([]reflect.Value{reflect.ValueOf(rawConfig)})
	if len(results) != 1 {
		return fmt.Errorf("unexpected Merge result for configuration %q", name)
	}

	// Check if Merge returned an error
	if !results[0].IsNil() {
		return results[0].Interface().(error)
	}

	return nil
}

// Unregister removes a configuration from the registry.
// This is useful for testing or dynamic plugin unloading.
//
// Parameters:
//   - name: Unique identifier for this configuration
//
// Returns an error if the configuration is not registered.
func Unregister(name string) error {
	return globalRegistry.Unregister(name)
}

// Unregister removes a configuration from the registry.
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.configs[name]; !exists {
		return fmt.Errorf("configuration %q not found", name)
	}

	delete(r.configs, name)
	delete(r.types, name)
	return nil
}

// List returns all registered configuration names.
// This is useful for debugging and introspection.
func List() []string {
	return globalRegistry.List()
}

// List returns all registered configuration names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.configs))
	for name := range r.configs {
		names = append(names, name)
	}
	return names
}

// Exists checks if a configuration is registered.
//
// Parameters:
//   - name: Unique identifier for this configuration
//
// Returns true if the configuration exists, false otherwise.
func Exists(name string) bool {
	return globalRegistry.Exists(name)
}

// Exists checks if a configuration is registered.
func (r *Registry) Exists(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.configs[name]
	return exists
}

// Reset clears all registered configurations.
// This is primarily useful for testing.
func Reset() {
	globalRegistry.Reset()
}

// Reset clears all registered configurations.
func (r *Registry) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.configs = make(map[string]interface{})
	r.types = make(map[string]reflect.Type)
}

// GetSchema retrieves the JSON Schema for a registered configuration.
//
// Parameters:
//   - name: Unique identifier for this configuration
//
// Returns the schema or an error if not found.
func GetSchema(name string) (map[string]interface{}, error) {
	return globalRegistry.GetSchema(name)
}

// GetSchema retrieves the JSON Schema for a registered configuration.
func (r *Registry) GetSchema(name string) (map[string]interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	raw, exists := r.configs[name]
	if !exists {
		return nil, fmt.Errorf("configuration %q not found", name)
	}

	// Use reflection to access the Schema field
	v := reflect.ValueOf(raw)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	schemaField := v.FieldByName("Schema")
	if !schemaField.IsValid() {
		return nil, fmt.Errorf("configuration %q has no Schema field", name)
	}

	if schemaField.IsNil() {
		return nil, fmt.Errorf("configuration %q has no schema defined", name)
	}

	// Call GenerateSchema method
	schema := schemaField.Interface().(ConfigSchema)
	return schema.GenerateSchema()
}

// Validate validates a registered configuration against its schema.
//
// Parameters:
//   - name: Unique identifier for this configuration
//
// Returns an error if validation fails or the configuration is not found.
func Validate(name string) error {
	return globalRegistry.Validate(name)
}

// Validate validates a registered configuration against its schema.
func (r *Registry) Validate(name string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	raw, exists := r.configs[name]
	if !exists {
		return fmt.Errorf("configuration %q not found", name)
	}

	// Use reflection to call the Validate method
	validateMethod := reflect.ValueOf(raw).MethodByName("Validate")
	if !validateMethod.IsValid() {
		return fmt.Errorf("configuration %q does not support validation", name)
	}

	// Call Validate method
	results := validateMethod.Call(nil)
	if len(results) != 1 {
		return fmt.Errorf("unexpected Validate result for configuration %q", name)
	}

	// Check if Validate returned an error
	if !results[0].IsNil() {
		return results[0].Interface().(error)
	}

	return nil
}
