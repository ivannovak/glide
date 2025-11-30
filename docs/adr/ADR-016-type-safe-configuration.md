# ADR-016: Type-Safe Configuration

## Status
Accepted

## Date
2025-11-28

## Context

The original configuration system used `map[string]interface{}` for all plugin and custom configuration. This approach had several problems:

1. **Runtime Type Assertions**: Every access required type assertion
2. **No Compile-Time Safety**: Type errors discovered at runtime
3. **Poor IDE Support**: No autocomplete or type checking
4. **Documentation Gap**: Config structure not self-documenting
5. **Validation Difficulty**: Custom validation for each field

Example of the problematic pattern:

```go
// Old pattern - runtime type assertions
func (p *MyPlugin) Configure(raw interface{}) error {
    config, ok := raw.(map[string]interface{})
    if !ok {
        return errors.New("invalid config type")
    }

    apiKey, ok := config["api_key"].(string)
    if !ok {
        return errors.New("api_key must be a string")
    }

    timeout, ok := config["timeout"].(int)
    if !ok {
        timeout = 30 // default
    }

    // ... more type assertions
}
```

## Decision

We implemented type-safe configuration using Go generics:

### Core Design

```go
// Type-safe configuration wrapper
type TypedConfig[T any] struct {
    Value    T           // Actual configuration data
    Name     string      // Configuration identifier
    Version  int         // Schema version for migrations
    Defaults T           // Default values
    Schema   ConfigSchema // JSON Schema for validation
}

// Plugin interface with type parameter
type Plugin[C any] interface {
    ConfigSchema() map[string]interface{}
    Configure(ctx context.Context, config C) error
    // ...
}
```

### Benefits

1. **Compile-Time Type Safety**: Errors caught before runtime
2. **IDE Support**: Full autocomplete and type checking
3. **Self-Documenting**: Config struct defines the schema
4. **Easier Validation**: Field-level validation via tags
5. **Better Defaults**: Type-safe default values

## Consequences

### Positive

1. **Fewer Runtime Errors**: Type mismatches caught at compile time
2. **Better Developer Experience**: IDE autocomplete works
3. **Self-Documenting Config**: Struct tags describe schema
4. **Validation Integration**: Works with JSON Schema
5. **Migration Support**: Version-aware config updates

### Negative

1. **Go Version Requirement**: Requires Go 1.18+ for generics
2. **Migration Effort**: Existing plugins need updates
3. **Complexity**: Generic types can be harder to understand
4. **Serialization**: Requires careful handling of JSON/YAML

## Implementation

### Defining Configuration

```go
// plugins/myplugin/config.go
type Config struct {
    // Required field with validation
    APIKey string `json:"api_key" yaml:"api_key" validate:"required"`

    // Optional field with default
    Timeout int `json:"timeout" yaml:"timeout" validate:"min=1"`

    // Nested configuration
    Retry RetryConfig `json:"retry" yaml:"retry"`
}

type RetryConfig struct {
    MaxAttempts int           `json:"max_attempts" yaml:"max_attempts"`
    Delay       time.Duration `json:"delay" yaml:"delay"`
}
```

### Creating Typed Config

```go
// Register with defaults
cfg := config.NewTypedConfig("my-plugin", Config{
    Timeout: 30,
    Retry: RetryConfig{
        MaxAttempts: 3,
        Delay:       time.Second,
    },
})

// Set validation schema
cfg.SetSchema(config.ConfigSchema{
    Type: "object",
    Properties: map[string]config.PropertySchema{
        "api_key": {Type: "string", MinLength: ptrTo(10)},
        "timeout": {Type: "integer", Minimum: ptrTo(1.0)},
    },
    Required: []string{"api_key"},
})
```

### Plugin Configuration

```go
type MyPlugin struct {
    v2.BasePlugin[Config]
}

func (p *MyPlugin) Configure(ctx context.Context, cfg Config) error {
    // cfg is strongly typed - no assertions needed
    if cfg.APIKey == "" {
        return errors.New("API key required")
    }

    // Store config
    return p.Init(cfg)
}

// Access config anywhere
func (p *MyPlugin) Execute(ctx context.Context, args []string) error {
    cfg := p.GetConfig() // Returns Config, not interface{}
    client := NewClient(cfg.APIKey, cfg.Timeout)
    // ...
}
```

### Migration Support

```go
// Define migrations
migrations := []config.Migration{
    {
        FromVersion: 1,
        ToVersion:   2,
        Transform: func(data map[string]interface{}) map[string]interface{} {
            // Rename field
            if old, ok := data["apiKey"]; ok {
                data["api_key"] = old
                delete(data, "apiKey")
            }
            return data
        },
    },
}

// Apply migrations automatically
cfg.SetMigrations(migrations)
cfg.Merge(rawConfig)
```

## Alternatives Considered

### 1. mapstructure with Tags

Use mapstructure to decode maps into structs.

```go
var cfg Config
mapstructure.Decode(rawConfig, &cfg)
```

**Rejected because**:
- Still runtime errors
- No compile-time validation
- Less type safety than generics

### 2. Code Generation

Generate config types from schema.

**Rejected because**:
- Adds build complexity
- Less flexible
- Schema-first is harder to maintain

### 3. Interface-Based Configuration

Define interfaces for config types.

**Rejected because**:
- More boilerplate
- Less type safety
- Interfaces can't have fields

### 4. Keep map[string]interface{}

Continue with dynamic configuration.

**Rejected because**:
- All the problems we're solving remain
- Poor developer experience
- Runtime errors

## Global Registry

For application-wide configuration:

```go
// Register configuration
config.GlobalRegistry.Register("my-plugin", cfg)

// Retrieve typed config anywhere
cfg, err := config.GlobalRegistry.Get[Config]("my-plugin")
if err != nil {
    return err
}
// cfg is fully typed
```

## Validation

JSON Schema validation is integrated:

```go
schema := cfg.ConfigSchema{
    Type: "object",
    Properties: map[string]cfg.PropertySchema{
        "port": {
            Type:    "integer",
            Minimum: ptrTo(1.0),
            Maximum: ptrTo(65535.0),
        },
        "host": {
            Type:      "string",
            Pattern:   "^[a-zA-Z0-9.-]+$",
            MaxLength: ptrTo(255),
        },
    },
    Required: []string{"host"},
}

tc := cfg.NewTypedConfig("server", defaults)
tc.SetSchema(schema)

err := tc.Validate()
if err != nil {
    // Detailed validation errors
}
```

## References

- [Go Generics Tutorial](https://go.dev/doc/tutorial/generics)
- [JSON Schema](https://json-schema.org/)
- [pkg/config Documentation](../../pkg/config/doc.go)
- [Plugin Development Guide](../guides/plugin-development.md)
- [SDK v2 Migration Guide](../guides/PLUGIN-SDK-V2-MIGRATION.md)
