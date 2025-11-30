# Plugin SDK v2 Migration Guide

> **Important:** SDK v1 is deprecated and no longer supported. All plugins must use SDK v2.

This guide helps you migrate existing plugins to SDK v2. The v2 SDK provides significant improvements in type safety, developer experience, and maintainability.

## Table of Contents

- [Why SDK v2?](#why-sdk-v2)
- [Key Differences](#key-differences)
- [Migration Steps](#migration-steps)
- [Code Migration Examples](#code-migration-examples)
- [FAQ](#faq)

## Why SDK v2?

SDK v2 addresses several limitations:

| Feature | v1 (Deprecated) | v2 |
|---------|-----------------|-----|
| **Configuration** | `map[string]string` requiring manual parsing | Type-safe generics `C` |
| **Command Registration** | Manual Cobra setup | Declarative `[]Command` |
| **Lifecycle** | Inconsistent across plugin types | Unified Init/Start/Stop/HealthCheck |
| **Metadata** | Multiple methods | Single `Metadata()` struct |
| **Base Implementation** | Complex boilerplate | Simple `BasePlugin[C]` embed |
| **Validation** | Manual | Automatic via ConfigSchema |

## Key Differences

### Configuration

**Before (v1):**
```go
func (p *MyPlugin) Configure(ctx context.Context, req *v1.ConfigureRequest) (*v1.ConfigureResponse, error) {
    // Manual type conversion, error prone
    apiKey, ok := req.Config["apiKey"]
    if !ok {
        return &v1.ConfigureResponse{Success: false}, nil
    }
    timeout, _ := strconv.Atoi(req.Config["timeout"])
    // ...
}
```

**After (v2):**
```go
type Config struct {
    APIKey  string `json:"apiKey" validate:"required"`
    Timeout int    `json:"timeout" validate:"min=1"`
}

func (p *MyPlugin) Configure(ctx context.Context, config Config) error {
    // Type-safe, validated automatically
    return p.BasePlugin.Configure(ctx, config)
}
```

### Plugin Structure

**Before (v1):**
```go
type MyPlugin struct {
    v1.BasePlugin
    apiKey string
    timeout int
}

func (p *MyPlugin) Name() string { return "my-plugin" }
func (p *MyPlugin) Version() string { return "1.0.0" }
func (p *MyPlugin) Description() string { return "..." }
```

**After (v2):**
```go
type MyPlugin struct {
    v2.BasePlugin[Config]
}

func (p *MyPlugin) Metadata() v2.Metadata {
    return v2.Metadata{
        Name:        "my-plugin",
        Version:     "1.0.0",
        Description: "...",
    }
}
```

### Commands

**Before (v1):**
```go
func (p *MyPlugin) Register(root *cobra.Command) error {
    cmd := &cobra.Command{
        Use:   "status",
        Short: "Check status",
        RunE: func(cmd *cobra.Command, args []string) error {
            // implementation
        },
    }
    root.AddCommand(cmd)
    return nil
}
```

**After (v2):**
```go
func (p *MyPlugin) Commands() []v2.Command {
    return []v2.Command{
        {
            Name:        "status",
            Description: "Check status",
            Handler: v2.SimpleCommandHandler(func(ctx context.Context, req *v2.ExecuteRequest) (*v2.ExecuteResponse, error) {
                return &v2.ExecuteResponse{ExitCode: 0, Output: "OK"}, nil
            }),
        },
    }
}
```

## Migration Steps

### Step 1: Update Imports

```go
// Remove
import "github.com/ivannovak/glide/v2/pkg/plugin/sdk/v1"

// Add
import "github.com/ivannovak/glide/v2/pkg/plugin/sdk/v2"
```

### Step 2: Define Config Type

```go
type Config struct {
    APIKey    string `json:"apiKey" validate:"required"`
    Timeout   int    `json:"timeout" validate:"min=1,max=300"`
    EnableSSL bool   `json:"enableSSL"`
}
```

### Step 3: Update Plugin Struct

```go
// Before
type MyPlugin struct {
    v1.BasePlugin
    config map[string]string
}

// After
type MyPlugin struct {
    v2.BasePlugin[Config]
}
```

### Step 4: Implement Metadata

```go
func (p *MyPlugin) Metadata() v2.Metadata {
    return v2.Metadata{
        Name:        "my-plugin",
        Version:     "2.0.0",
        Description: "My plugin",
        Author:      "Your Name",
    }
}
```

### Step 5: Update Configure

```go
func (p *MyPlugin) Configure(ctx context.Context, config Config) error {
    // Type-safe config available immediately
    return p.BasePlugin.Configure(ctx, config)
}
```

### Step 6: Update Commands

```go
func (p *MyPlugin) Commands() []v2.Command {
    return []v2.Command{
        {
            Name:        "mycommand",
            Description: "Does something",
            Category:    "custom",
            Handler:     v2.SimpleCommandHandler(p.myCommandHandler),
        },
    }
}

func (p *MyPlugin) myCommandHandler(ctx context.Context, req *v2.ExecuteRequest) (*v2.ExecuteResponse, error) {
    // Access type-safe config
    config := p.GetConfig()

    return &v2.ExecuteResponse{
        ExitCode: 0,
        Output:   fmt.Sprintf("Using API key: %s", config.APIKey),
    }, nil
}
```

### Step 7: Add Lifecycle (Optional)

```go
func (p *MyPlugin) Init(ctx context.Context) error {
    // One-time initialization
    return nil
}

func (p *MyPlugin) Start(ctx context.Context) error {
    // Start operation
    return nil
}

func (p *MyPlugin) Stop(ctx context.Context) error {
    // Cleanup
    return nil
}

func (p *MyPlugin) HealthCheck(ctx context.Context) error {
    // Return nil if healthy
    return nil
}
```

### Step 8: Update Main

```go
func main() {
    plugin := &MyPlugin{}
    if err := v2.Serve(plugin); err != nil {
        os.Exit(1)
    }
}
```

## Code Migration Examples

### Complete Before/After

**Before (v1):**
```go
package main

import (
    "context"
    "strconv"

    "github.com/spf13/cobra"
    "github.com/ivannovak/glide/v2/pkg/config"
    v1 "github.com/ivannovak/glide/v2/pkg/plugin/sdk/v1"
)

type Plugin struct {
    apiKey  string
    timeout int
}

func init() {
    config.Register("my-plugin", map[string]string{})
}

func (p *Plugin) Name() string        { return "my-plugin" }
func (p *Plugin) Version() string     { return "1.0.0" }
func (p *Plugin) Description() string { return "My plugin" }

func (p *Plugin) Configure() error {
    cfg, _ := config.Get[map[string]string]("my-plugin")
    p.apiKey = cfg["apiKey"]
    p.timeout, _ = strconv.Atoi(cfg["timeout"])
    return nil
}

func (p *Plugin) Register(root *cobra.Command) error {
    cmd := &cobra.Command{
        Use:   "greet",
        Short: "Greet the user",
        RunE: func(cmd *cobra.Command, args []string) error {
            fmt.Println("Hello!")
            return nil
        },
    }
    root.AddCommand(cmd)
    return nil
}
```

**After (v2):**
```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/ivannovak/glide/v2/pkg/plugin/sdk/v2"
)

type Config struct {
    APIKey  string `json:"apiKey" validate:"required"`
    Timeout int    `json:"timeout" validate:"min=1"`
}

type Plugin struct {
    v2.BasePlugin[Config]
}

func (p *Plugin) Metadata() v2.Metadata {
    return v2.Metadata{
        Name:        "my-plugin",
        Version:     "2.0.0",
        Description: "My plugin",
    }
}

func (p *Plugin) Configure(ctx context.Context, config Config) error {
    return p.BasePlugin.Configure(ctx, config)
}

func (p *Plugin) Commands() []v2.Command {
    return []v2.Command{
        {
            Name:        "greet",
            Description: "Greet the user",
            Handler: v2.SimpleCommandHandler(func(ctx context.Context, req *v2.ExecuteRequest) (*v2.ExecuteResponse, error) {
                return &v2.ExecuteResponse{ExitCode: 0, Output: "Hello!"}, nil
            }),
        },
    }
}

func main() {
    if err := v2.Serve(&Plugin{}); err != nil {
        os.Exit(1)
    }
}
```

## FAQ

### Q: What happened to v1 support?

SDK v1 is deprecated and no longer supported. All plugins must migrate to v2.

### Q: How do I validate configuration?

Use struct tags for validation:

```go
type Config struct {
    APIKey string `json:"apiKey" validate:"required,min=10"`
    Port   int    `json:"port" validate:"min=1,max=65535"`
}
```

Or implement `ConfigSchema()` for JSON Schema validation.

### Q: How do I handle interactive commands?

Use `InteractiveCommandHandler`:

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

### Q: How do dependencies work?

Declare in Metadata:

```go
func (p *MyPlugin) Metadata() v2.Metadata {
    return v2.Metadata{
        Name:    "my-plugin",
        Version: "2.0.0",
        Dependencies: []v2.Dependency{
            {Name: "docker", Version: "^1.0.0"},
        },
    }
}
```

## Resources

- [Plugin Development Guide](../plugin-development.md)
- [SDK v2 API Reference](../../pkg/plugin/sdk/v2/)
- [Plugin Tutorial](../tutorials/02-first-plugin.md)
