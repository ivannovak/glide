# Plugin SDK v2 Migration Guide

This guide helps you migrate plugins from SDK v1 to SDK v2. The v2 SDK provides significant improvements in type safety, developer experience, and maintainability while maintaining full backward compatibility with v1 plugins.

## Table of Contents

- [Why Migrate to v2?](#why-migrate-to-v2)
- [Key Differences](#key-differences)
- [Migration Steps](#migration-steps)
- [v1 to v2 Comparison](#v1-to-v2-comparison)
- [Example Migration](#example-migration)
- [Backward Compatibility](#backward-compatibility)
- [FAQ](#faq)

## Why Migrate to v2?

The v2 SDK addresses several pain points from v1:

### Type Safety
- **v1**: Configuration uses `map[string]string` requiring manual type conversion
- **v2**: Type-safe configuration using Go generics - no manual conversion needed

### Simplified Plugin Development
- **v1**: Separate interfaces for in-process vs gRPC plugins, manual command registration
- **v2**: Unified Plugin interface, declarative command definitions

### Better Developer Experience
- **v1**: BasePlugin requires manual handler registration and boilerplate
- **v2**: Sensible defaults via embedding BasePlugin with minimal boilerplate

### Improved Lifecycle Management
- **v1**: Lifecycle exists in SDK but not in gRPC protocol
- **v2**: Unified lifecycle across all plugin types with proper state management

### Cleaner API
- **v1**: Multiple interfaces to implement (PluginIdentifier, PluginRegistrar, PluginConfigurable)
- **v2**: Single Plugin[C] interface with type parameter for config

## Key Differences

| Feature | v1 | v2 |
|---------|----|----|
| **Configuration** | `map[string]string` | Type-safe generic `C` |
| **Command Registration** | Manual Cobra registration | Declarative `[]Command` |
| **Lifecycle** | Separate interface | Built into Plugin |
| **Metadata** | Multiple methods | Single `Metadata()` |
| **Base Implementation** | Complex setup | Embed `BasePlugin[C]` |
| **Schema Validation** | Manual | Automatic via `ConfigSchema()` |

## Migration Steps

### Step 1: Update Imports

```go
// Old (v1)
import (
    "github.com/ivannovak/glide/v2/pkg/plugin/sdk/v1"
)

// New (v2)
import (
    v2 "github.com/ivannovak/glide/v2/pkg/plugin/sdk/v2"
)
```

### Step 2: Define Your Config Type

```go
// v2 uses type-safe config structs
type MyPluginConfig struct {
    APIKey    string `json:"apiKey"`
    Timeout   int    `json:"timeout"`
    EnableSSL bool   `json:"enableSSL"`
}
```

### Step 3: Convert Plugin Struct

```go
// Old (v1)
type MyPlugin struct {
    v1.BasePlugin
    apiKey string
}

// New (v2)
type MyPlugin struct {
    v2.BasePlugin[MyPluginConfig]
    apiClient *APIClient
}
```

### Step 4: Implement Plugin Interface

```go
// Metadata - simpler than v1's Name(), Version(), Description()
func (p *MyPlugin) Metadata() v2.Metadata {
    return v2.Metadata{
        Name:        "my-plugin",
        Version:     "2.0.0",
        Description: "My awesome plugin migrated to v2",
        Author:      "Your Name",
        License:     "MIT",
    }
}

// ConfigSchema - optional, provides validation
func (p *MyPlugin) ConfigSchema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "apiKey": map[string]interface{}{
                "type": "string",
                "description": "API key for authentication",
            },
            "timeout": map[string]interface{}{
                "type": "integer",
                "default": 30,
            },
        },
        "required": []string{"apiKey"},
    }
}

// Configure - type-safe!
func (p *MyPlugin) Configure(ctx context.Context, config MyPluginConfig) error {
    // No manual type conversion needed!
    p.apiClient = NewAPIClient(config.APIKey, config.Timeout)
    return nil
}

// Commands - declarative definition
func (p *MyPlugin) Commands() []v2.Command {
    return []v2.Command{
        {
            Name:        "status",
            Description: "Check plugin status",
            Handler: v2.SimpleCommandHandler(func(ctx context.Context, req *v2.ExecuteRequest) (*v2.ExecuteResponse, error) {
                status := p.apiClient.GetStatus()
                return &v2.ExecuteResponse{
                    ExitCode: 0,
                    Output:   fmt.Sprintf("Status: %s", status),
                }, nil
            }),
        },
    }
}
```

### Step 5: Lifecycle (Optional)

If your plugin needs initialization or cleanup, implement lifecycle methods:

```go
func (p *MyPlugin) Init(ctx context.Context) error {
    // One-time initialization
    return nil
}

func (p *MyPlugin) Start(ctx context.Context) error {
    // Start background workers, open connections, etc.
    return p.apiClient.Connect()
}

func (p *MyPlugin) Stop(ctx context.Context) error {
    // Cleanup
    return p.apiClient.Disconnect()
}

func (p *MyPlugin) HealthCheck(ctx context.Context) error {
    return p.apiClient.Ping()
}
```

If you don't need these, the BasePlugin provides no-op defaults.

## v1 to v2 Comparison

### Configuration

**v1 gRPC Plugin:**
```go
func (p *MyPlugin) Configure(ctx context.Context, req *v1.ConfigureRequest) (*v1.ConfigureResponse, error) {
    // Manual type conversion
    apiKey, ok := req.Config["apiKey"]
    if !ok {
        return &v1.ConfigureResponse{Success: false, Message: "apiKey required"}, nil
    }

    timeout := 30
    if timeoutStr, ok := req.Config["timeout"]; ok {
        var err error
        timeout, err = strconv.Atoi(timeoutStr)
        if err != nil {
            return &v1.ConfigureResponse{Success: false, Message: "invalid timeout"}, nil
        }
    }

    p.apiKey = apiKey
    p.timeout = timeout
    return &v1.ConfigureResponse{Success: true}, nil
}
```

**v2 Plugin:**
```go
type MyPluginConfig struct {
    APIKey  string `json:"apiKey"`
    Timeout int    `json:"timeout"`
}

func (p *MyPlugin) Configure(ctx context.Context, config MyPluginConfig) error {
    // Type-safe, no conversion needed!
    p.apiClient = NewAPIClient(config.APIKey, config.Timeout)
    return nil
}
```

### Command Registration

**v1 In-Process Plugin:**
```go
func (p *MyPlugin) Register(root *cobra.Command) error {
    statusCmd := &cobra.Command{
        Use:   "status",
        Short: "Check status",
        RunE: func(cmd *cobra.Command, args []string) error {
            status := p.getStatus()
            fmt.Println(status)
            return nil
        },
    }
    root.AddCommand(statusCmd)
    return nil
}
```

**v1 gRPC Plugin:**
```go
func (p *MyPlugin) setupCommands() {
    p.RegisterCommand("status", &StatusCommand{plugin: p})
}

type StatusCommand struct {
    plugin *MyPlugin
}

func (c *StatusCommand) Info() *v1.CommandInfo {
    return &v1.CommandInfo{
        Name:        "status",
        Description: "Check status",
    }
}

func (c *StatusCommand) Execute(ctx context.Context, req *v1.ExecuteRequest) (*v1.ExecuteResponse, error) {
    status := c.plugin.getStatus()
    return &v1.ExecuteResponse{
        Success:  true,
        ExitCode: 0,
        Stdout:   []byte(status),
    }, nil
}
```

**v2 Plugin (Works for Both):**
```go
func (p *MyPlugin) Commands() []v2.Command {
    return []v2.Command{
        {
            Name:        "status",
            Description: "Check status",
            Category:    "core",
            Handler: v2.SimpleCommandHandler(func(ctx context.Context, req *v2.ExecuteRequest) (*v2.ExecuteResponse, error) {
                status := p.getStatus()
                return &v2.ExecuteResponse{
                    ExitCode: 0,
                    Output:   status,
                }, nil
            }),
        },
    }
}
```

### Metadata

**v1 In-Process:**
```go
func (p *MyPlugin) Name() string { return "my-plugin" }
func (p *MyPlugin) Version() string { return "1.0.0" }
func (p *MyPlugin) Description() string { return "My plugin" }
```

**v1 gRPC:**
```protobuf
message PluginMetadata {
  string name = 1;
  string version = 2;
  string description = 3;
  // ... 10+ other fields
}
```

**v2 (Unified):**
```go
func (p *MyPlugin) Metadata() v2.Metadata {
    return v2.Metadata{
        Name:         "my-plugin",
        Version:      "2.0.0",
        Description:  "My plugin",
        Author:       "Your Name",
        License:      "MIT",
        Dependencies: []v2.Dependency{
            {Name: "docker", Version: "^1.0.0"},
        },
    }
}
```

## Example Migration

Let's migrate a complete v1 plugin to v2.

### Before (v1 In-Process Plugin)

```go
package myplugin

import (
    "fmt"
    "github.com/spf13/cobra"
    "github.com/ivannovak/glide/v2/pkg/config"
)

type Config struct {
    APIKey string `yaml:"apiKey"`
}

type Plugin struct {
    config *Config
}

func init() {
    config.Register("my-plugin", &Config{})
}

func (p *Plugin) Name() string        { return "my-plugin" }
func (p *Plugin) Version() string     { return "1.0.0" }
func (p *Plugin) Description() string { return "My plugin" }

func (p *Plugin) Configure() error {
    cfg, err := config.Get[*Config]("my-plugin")
    if err != nil {
        return err
    }
    p.config = cfg
    return nil
}

func (p *Plugin) Register(root *cobra.Command) error {
    cmd := &cobra.Command{
        Use:   "greet",
        Short: "Greet the user",
        RunE: func(cmd *cobra.Command, args []string) error {
            fmt.Printf("Hello from %s!\n", p.Name())
            return nil
        },
    }
    root.AddCommand(cmd)
    return nil
}
```

### After (v2 Plugin)

```go
package myplugin

import (
    "context"
    "fmt"
    v2 "github.com/ivannovak/glide/v2/pkg/plugin/sdk/v2"
)

type Config struct {
    APIKey string `json:"apiKey"`
}

type Plugin struct {
    v2.BasePlugin[Config]
}

func New() *Plugin {
    p := &Plugin{}
    p.SetMetadata(v2.Metadata{
        Name:        "my-plugin",
        Version:     "2.0.0",
        Description: "My plugin migrated to v2",
    })

    p.SetCommands([]v2.Command{
        {
            Name:        "greet",
            Description: "Greet the user",
            Category:    "core",
            Handler: v2.SimpleCommandHandler(p.greet),
        },
    })

    return p
}

func (p *Plugin) Configure(ctx context.Context, config Config) error {
    // Config is already validated and type-safe!
    return p.BasePlugin.Configure(ctx, config)
}

func (p *Plugin) greet(ctx context.Context, req *v2.ExecuteRequest) (*v2.ExecuteResponse, error) {
    return &v2.ExecuteResponse{
        ExitCode: 0,
        Output:   fmt.Sprintf("Hello from %s!", p.Metadata().Name),
    }, nil
}
```

**What Changed:**
1. ✅ Embed `v2.BasePlugin[Config]` instead of custom struct
2. ✅ Single `Metadata()` instead of Name/Version/Description
3. ✅ Declarative `Commands()` instead of manual Cobra registration
4. ✅ Type-safe `Configure(ctx, Config)` instead of `Configure() + config.Get`
5. ✅ Removed `init()` and global config registration
6. ✅ Constructor pattern with `New()` for cleaner initialization

## Backward Compatibility

### Running v1 Plugins in v2 Environment

The v2 SDK includes adapters that allow v1 plugins to work seamlessly:

```go
// Wrap v1 gRPC plugin for v2
v1GRPCPlugin := myV1Plugin // v1.GlidePluginClient
v2Plugin := v2.AdaptV1GRPCPlugin(v1GRPCPlugin)

// Wrap v1 in-process plugin for v2
v1InProcessPlugin := &myOldPlugin{}
v2Plugin := v2.AdaptV1InProcessPlugin(v1InProcessPlugin)
```

### Running v2 Plugins in v1 Environment

If you need to run a v2 plugin in a v1 context (during gradual migration):

```go
v2Plugin := &MyV2Plugin{}
v1Adapter := v2.AdaptV2ToV1(v2Plugin)

// v1Adapter now implements v1 interfaces
root := &cobra.Command{}
v1Adapter.Register(root)
```

### Coexistence

You can have both v1 and v2 plugins loaded simultaneously. The CLI automatically detects plugin version and uses appropriate adapters.

## FAQ

### Q: Do I need to migrate immediately?

No. v1 plugins continue to work indefinitely through the adapter layer. Migrate when convenient or when building new plugins.

### Q: Can I mix v1 and v2 plugins?

Yes! The CLI handles both versions transparently.

### Q: What about existing v1 plugins I don't control?

They continue to work without changes. The v2 adapter handles them automatically.

### Q: Is v2 slower than v1 due to adapters?

No. Adapters are zero-overhead wrappers. Pure v2 plugins have slightly better performance due to reduced type conversions.

### Q: How do I validate configuration in v2?

Implement `ConfigSchema()` to return a JSON Schema. The CLI validates config against this schema before calling `Configure()`.

```go
func (p *MyPlugin) ConfigSchema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "apiKey": map[string]interface{}{
                "type": "string",
                "minLength": 1,
            },
        },
        "required": []string{"apiKey"},
    }
}
```

### Q: What about interactive commands?

v2 has improved interactive command support via `InteractiveCommandHandler`:

```go
{
    Name:        "shell",
    Interactive: true,
    InteractiveHandler: &MyInteractiveHandler{},
}

type MyInteractiveHandler struct{}

func (h *MyInteractiveHandler) ExecuteInteractive(ctx context.Context, session *v2.InteractiveSession) error {
    for {
        line, err := session.ReadLine()
        if err != nil {
            return err
        }
        session.WriteLine("> " + line)
    }
}
```

### Q: How do dependencies work in v2?

Declare dependencies in Metadata:

```go
func (p *MyPlugin) Metadata() v2.Metadata {
    return v2.Metadata{
        Name:    "my-plugin",
        Version: "2.0.0",
        Dependencies: []v2.Dependency{
            {Name: "docker", Version: "^1.0.0"},
            {Name: "database", Version: ">=2.0.0 <3.0.0"},
        },
    }
}
```

The CLI resolves dependencies and loads plugins in correct order.

### Q: Can I access other plugins from my plugin?

Not directly in v2.0. This feature is planned for v2.1 via plugin context injection.

## Next Steps

1. Read the [Plugin Development Guide](PLUGIN-DEVELOPMENT.md) for v2 best practices
2. Check out example plugins in `examples/plugins/v2/`
3. Join the discussion in GitHub Discussions for migration questions

## Resources

- [Plugin SDK v2 API Reference](../../pkg/plugin/sdk/v2/)
- [Plugin Development Guide](PLUGIN-DEVELOPMENT.md)
- [v2 Example Plugins](../../examples/plugins/v2/)
- [Migration Support](https://github.com/ivannovak/glide/discussions)

---

**Version:** 1.0
**Last Updated:** 2025-01-28
**Status:** Active
