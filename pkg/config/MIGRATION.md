# Migration Guide: From `map[string]interface{}` to Type-Safe Configuration

This document provides a comprehensive guide for migrating from untyped `map[string]interface{}` configuration to the new type-safe configuration system.

## Table of Contents

1. [Overview](#overview)
2. [Why Migrate?](#why-migrate)
3. [Migration Path](#migration-path)
4. [Step-by-Step Examples](#step-by-step-examples)
5. [Common Patterns](#common-patterns)
6. [Troubleshooting](#troubleshooting)

---

## Overview

The new type-safe configuration system eliminates runtime type assertions and provides:

- **Compile-time type safety**: Catch configuration errors at compile time
- **Autocompletion**: IDE support for configuration fields
- **Schema validation**: Automatic JSON Schema generation and validation
- **Better error messages**: Clear type mismatch errors with type information
- **Zero runtime overhead**: Generics compile down to efficient code

---

## Why Migrate?

### Before (Unsafe ❌)

```go
// Old way: Untyped config with runtime type assertions
pluginConfig := map[string]interface{}{
    "api_key": "secret123",
    "timeout": 30,
    "enabled": true,
}

// Runtime type assertion - can panic or fail
apiKey, ok := pluginConfig["api_key"].(string)
if !ok {
    return fmt.Errorf("api_key must be a string")
}

timeout, ok := pluginConfig["timeout"].(int)
if !ok {
    return fmt.Errorf("timeout must be an int")
}

// Easy to make typos - no compile-time checking
value := pluginConfig["api_ky"]  // Typo! Returns nil at runtime
```

### After (Safe ✅)

```go
// New way: Type-safe configuration
type PluginConfig struct {
    APIKey  string `json:"api_key" yaml:"api_key" validate:"required"`
    Timeout int    `json:"timeout" yaml:"timeout" validate:"min=1,max=3600"`
    Enabled bool   `json:"enabled" yaml:"enabled"`
}

// Register with defaults
config.Register("my-plugin", PluginConfig{
    Timeout: 30,
    Enabled: false,
})

// Retrieve typed config - no type assertions!
cfg, err := config.GetValue[PluginConfig]("my-plugin")
if err != nil {
    return err
}

// Fully typed - IDE autocomplete works!
client := NewClient(cfg.APIKey, cfg.Timeout)
// cfg.APIKy  // Compile error - field doesn't exist!
```

---

## Migration Path

### Phase 1: Define Configuration Structs

For each plugin or component using `map[string]interface{}`, create a configuration struct:

```go
// Before
type Plugin struct {
    config map[string]interface{}
}

// After - Step 1: Define the config struct
type PluginConfig struct {
    // Use json/yaml tags for field mapping
    APIKey    string   `json:"api_key" yaml:"api_key" validate:"required"`
    Timeout   int      `json:"timeout" yaml:"timeout" validate:"min=1,max=3600"`
    Enabled   bool     `json:"enabled" yaml:"enabled"`
    Endpoints []string `json:"endpoints" yaml:"endpoints"`
}

type Plugin struct {
    config *config.TypedConfig[PluginConfig]  // Or just PluginConfig
}
```

### Phase 2: Register Configuration

In your plugin's initialization code, register the configuration:

```go
func init() {
    // Register with sensible defaults
    err := config.Register("my-plugin", PluginConfig{
        Timeout: 30,
        Enabled: false,
        Endpoints: []string{},
    })
    if err != nil {
        log.Fatalf("Failed to register plugin config: %v", err)
    }
}
```

### Phase 3: Update Configuration Loading

Replace `map[string]interface{}` operations with typed config access:

```go
// Before
func (p *Plugin) LoadConfig(rawConfig map[string]interface{}) error {
    p.config = rawConfig

    // Lots of type assertions...
    if apiKey, ok := rawConfig["api_key"].(string); ok {
        p.apiKey = apiKey
    }
    // ... more assertions ...

    return nil
}

// After
func (p *Plugin) LoadConfig(rawConfig interface{}) error {
    // Update the registered config
    if err := config.Update("my-plugin", rawConfig); err != nil {
        return fmt.Errorf("invalid config: %w", err)
    }

    // Validate against schema
    if err := config.Validate("my-plugin"); err != nil {
        return fmt.Errorf("config validation failed: %w", err)
    }

    // Get typed config
    cfg, err := config.Get[PluginConfig]("my-plugin")
    if err != nil {
        return err
    }

    p.config = cfg
    return nil
}
```

### Phase 4: Update Configuration Access

Replace runtime type assertions with direct field access:

```go
// Before
func (p *Plugin) GetAPIKey() (string, error) {
    apiKey, ok := p.config["api_key"].(string)
    if !ok {
        return "", fmt.Errorf("api_key not found or invalid type")
    }
    return apiKey, nil
}

func (p *Plugin) GetTimeout() int {
    if timeout, ok := p.config["timeout"].(int); ok {
        return timeout
    }
    return 30 // default
}

// After
func (p *Plugin) GetAPIKey() string {
    return p.config.Value.APIKey  // Typed access - no error handling needed!
}

func (p *Plugin) GetTimeout() int {
    return p.config.Value.Timeout  // Defaults handled during registration
}
```

---

## Step-by-Step Examples

### Example 1: Simple Plugin Configuration

**Before:**
```go
type MyPlugin struct {
    name   string
    config map[string]interface{}
}

func (p *MyPlugin) Initialize(config map[string]interface{}) error {
    p.config = config

    // Type assertions everywhere
    if name, ok := config["name"].(string); ok {
        p.name = name
    } else {
        return fmt.Errorf("name must be a string")
    }

    return nil
}

func (p *MyPlugin) GetSetting(key string) interface{} {
    return p.config[key]  // Untyped return
}
```

**After:**
```go
// Step 1: Define config struct
type MyPluginConfig struct {
    Name        string `json:"name" yaml:"name" validate:"required"`
    Description string `json:"description" yaml:"description"`
    Enabled     bool   `json:"enabled" yaml:"enabled"`
}

type MyPlugin struct {
    name   string
    config *config.TypedConfig[MyPluginConfig]
}

// Step 2: Register in init()
func init() {
    config.Register("my-plugin", MyPluginConfig{
        Enabled: true,  // Sensible default
    })
}

// Step 3: Update initialization
func (p *MyPlugin) Initialize(rawConfig interface{}) error {
    // Update registered config
    if err := config.Update("my-plugin", rawConfig); err != nil {
        return err
    }

    // Get typed config
    cfg, err := config.Get[MyPluginConfig]("my-plugin")
    if err != nil {
        return err
    }

    p.config = cfg
    p.name = cfg.Value.Name  // Type-safe access!
    return nil
}

// Step 4: Type-safe getters
func (p *MyPlugin) GetDescription() string {
    return p.config.Value.Description  // No type assertion needed!
}

func (p *MyPlugin) IsEnabled() bool {
    return p.config.Value.Enabled
}
```

### Example 2: Complex Nested Configuration

**Before:**
```go
config := map[string]interface{}{
    "database": map[string]interface{}{
        "host": "localhost",
        "port": 5432,
        "credentials": map[string]interface{}{
            "username": "admin",
            "password": "secret",
        },
    },
}

// Deeply nested type assertions
db := config["database"].(map[string]interface{})
creds := db["credentials"].(map[string]interface{})
username := creds["username"].(string)  // Can panic!
```

**After:**
```go
// Step 1: Define nested structs
type DatabaseCredentials struct {
    Username string `json:"username" yaml:"username" validate:"required"`
    Password string `json:"password" yaml:"password" validate:"required"`
}

type DatabaseConfig struct {
    Host        string               `json:"host" yaml:"host"`
    Port        int                  `json:"port" yaml:"port"`
    Credentials DatabaseCredentials `json:"credentials" yaml:"credentials"`
}

type AppConfig struct {
    Database DatabaseConfig `json:"database" yaml:"database"`
}

// Step 2: Register
config.Register("app", AppConfig{
    Database: DatabaseConfig{
        Host: "localhost",
        Port: 5432,
    },
})

// Step 3: Type-safe access
cfg, _ := config.GetValue[AppConfig]("app")
username := cfg.Database.Credentials.Username  // Clean and safe!
```

### Example 3: Migration for Plugin SDK Manager

This is the main migration target for Glide.

**Before:**
```go
// internal/config/types.go
type Config struct {
    Plugins map[string]interface{} `yaml:"plugins"`  // ❌ Untyped
}

// pkg/plugin/sdk/manager.go
type ManagerConfig struct {
    // No plugin-specific config type
}

func (m *Manager) LoadPlugin(info *PluginInfo, rawConfig map[string]interface{}) error {
    // Pass untyped config to plugin
    // ...
}
```

**After:**
```go
// Step 1: Define plugin config interface
// pkg/config/plugin.go
type PluginConfig interface {
    Validate() error
}

// Step 2: Update internal config
// internal/config/types.go
type Config struct {
    Plugins map[string]*config.TypedConfig[config.PluginConfig] `yaml:"plugins"`
}

// Step 3: Each plugin defines its own config
// plugins/my-plugin/config.go
type MyPluginConfig struct {
    APIEndpoint string `json:"api_endpoint" validate:"required,url"`
    RetryCount  int    `json:"retry_count" validate:"min=0,max=10"`
}

func init() {
    config.Register("my-plugin", MyPluginConfig{
        RetryCount: 3,
    })
}

// Step 4: Manager uses typed configs
// pkg/plugin/sdk/manager.go
func (m *Manager) LoadPlugin(info *PluginInfo, rawConfig interface{}) error {
    // Plugin configs are registered, just update them
    if err := config.Update(info.Name, rawConfig); err != nil {
        return fmt.Errorf("invalid plugin config: %w", err)
    }

    // Validate against schema
    if err := config.Validate(info.Name); err != nil {
        return fmt.Errorf("plugin config validation failed: %w", err)
    }

    // Plugin can now retrieve its typed config
    // ...
}
```

---

## Common Patterns

### Pattern 1: Optional Fields with Defaults

```go
type Config struct {
    RequiredField string `json:"required" validate:"required"`
    OptionalField string `json:"optional,omitempty"`  // omitempty = optional
    DefaultField  int    `json:"default"`
}

// Set defaults during registration
config.Register("app", Config{
    DefaultField: 42,  // This value used when not specified in YAML
})
```

### Pattern 2: Validation Rules

```go
type Config struct {
    Email    string `json:"email" validate:"required,email"`
    Age      int    `json:"age" validate:"min=0,max=120"`
    Role     string `json:"role" validate:"enum=admin|user|guest"`
    URL      string `json:"url" validate:"required,url"`
    Password string `json:"password" validate:"min=8"`
}
```

### Pattern 3: Arrays and Maps

```go
type Config struct {
    // Arrays
    Endpoints []string          `json:"endpoints"`
    Ports     []int             `json:"ports"`

    // Maps
    Headers   map[string]string `json:"headers"`
    Metadata  map[string]int    `json:"metadata"`
}
```

### Pattern 4: Updating Configuration at Runtime

```go
// Get current config
cfg, err := config.Get[MyConfig]("app")
if err != nil {
    return err
}

// Modify value
cfg.Value.Timeout = 60

// Validate new value
if err := cfg.Validate(); err != nil {
    return err
}

// Changes are automatically reflected
```

---

## Troubleshooting

### Error: "configuration type mismatch"

**Cause:** Requesting a config with the wrong type parameter.

```go
// Registered as PluginConfig
config.Register("my-plugin", PluginConfig{})

// Error: Requested as different type
cfg, err := config.Get[OtherConfig]("my-plugin")
// Error: configuration "my-plugin" type mismatch: registered as PluginConfig, requested as OtherConfig
```

**Solution:** Use the same type as registration.

### Error: "required field missing"

**Cause:** Configuration doesn't include a required field.

```go
type Config struct {
    APIKey string `json:"api_key" validate:"required"`
}

// Error: Missing required field
config.Update("app", map[string]interface{}{
    "timeout": 30,  // api_key is missing!
})
```

**Solution:** Either provide the field or remove `validate:"required"` tag.

### Error: "failed to unmarshal config"

**Cause:** YAML/JSON field names don't match struct tags.

```go
type Config struct {
    APIKey string `json:"apiKey"`  // Note: camelCase
}

// YAML uses snake_case
config.Update("app", map[string]interface{}{
    "api_key": "secret",  // Mismatch!
})
```

**Solution:** Ensure YAML/JSON tags match your config file format:
```go
type Config struct {
    APIKey string `json:"api_key" yaml:"api_key"`  // snake_case for YAML
}
```

---

## Backward Compatibility

During migration, you may need to support both old and new config formats:

```go
type Plugin struct {
    legacyConfig map[string]interface{}  // Old format
    typedConfig  *config.TypedConfig[PluginConfig]  // New format
}

func (p *Plugin) LoadConfig(rawConfig interface{}) error {
    // Try new format first
    if err := config.Update("my-plugin", rawConfig); err == nil {
        cfg, err := config.Get[PluginConfig]("my-plugin")
        if err == nil {
            p.typedConfig = cfg
            return nil
        }
    }

    // Fall back to legacy format
    if legacy, ok := rawConfig.(map[string]interface{}); ok {
        p.legacyConfig = legacy
        return nil
    }

    return fmt.Errorf("unsupported config format")
}

func (p *Plugin) GetAPIKey() string {
    // Prefer typed config
    if p.typedConfig != nil {
        return p.typedConfig.Value.APIKey
    }

    // Fall back to legacy
    if apiKey, ok := p.legacyConfig["api_key"].(string); ok {
        return apiKey
    }

    return ""
}
```

---

## Summary

| Old Way | New Way |
|---------|---------|
| `map[string]interface{}` | `TypedConfig[T]` |
| Runtime type assertions | Compile-time type safety |
| No validation | JSON Schema validation |
| Manual default handling | Automatic defaults |
| No IDE support | Full autocomplete |
| Panic on type mismatch | Clear error messages |

**Migration Checklist:**

- [x] Define configuration struct with proper tags
- [x] Register configuration in `init()`
- [x] Update configuration loading to use `config.Update()`
- [x] Replace type assertions with typed field access
- [x] Add validation rules using `validate` tags
- [x] Test with invalid configurations
- [x] Update documentation
- [x] Remove old `map[string]interface{}` code

---

For more information, see:
- `pkg/config/typed.go` - TypedConfig implementation
- `pkg/config/registry.go` - Registration and retrieval
- `pkg/config/schema.go` - Schema generation and validation
