# Plugin Development Guide

## Table of Contents
- [Overview](#overview)
- [Quick Start](#quick-start)
- [Plugin Architecture](#plugin-architecture)
  - [Plugin Interface](#plugin-interface)
  - [Plugin Metadata](#plugin-metadata)
  - [Plugin Dependencies](#plugin-dependencies)
- [Creating a Plugin](#creating-a-plugin)
- [Type-Safe Configuration](#type-safe-configuration)
- [Lifecycle Management](#lifecycle-management)
- [Command Registration](#command-registration)
- [Using Command Aliases](#using-command-aliases)
- [Best Practices](#best-practices)
- [Examples](#examples)

## Overview

Glide supports a powerful plugin system that allows developers to extend the CLI with custom commands. Plugins are standalone binaries that communicate with Glide via gRPC using SDK v2.

SDK v2 provides:

- **Type-safe configuration** using Go generics
- **Unified lifecycle management** (Init/Start/Stop/HealthCheck)
- **Declarative command definition**
- **Simplified development** with `BasePlugin[C]`

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/ivannovak/glide/v2/pkg/plugin/sdk/v2"
)

// Define your type-safe configuration
type MyConfig struct {
    APIKey  string `json:"apiKey" validate:"required"`
    Timeout int    `json:"timeout" validate:"min=1"`
}

// Create your plugin by embedding BasePlugin
type MyPlugin struct {
    v2.BasePlugin[MyConfig]
}

// Metadata returns plugin information
func (p *MyPlugin) Metadata() v2.Metadata {
    return v2.Metadata{
        Name:        "my-plugin",
        Version:     "1.0.0",
        Description: "My awesome plugin",
        Author:      "Your Name",
        License:     "MIT",
    }
}

// Configure is called with type-safe config
func (p *MyPlugin) Configure(ctx context.Context, config MyConfig) error {
    return p.BasePlugin.Configure(ctx, config)
}

// Commands returns the list of commands
func (p *MyPlugin) Commands() []v2.Command {
    return []v2.Command{
        {
            Name:        "greet",
            Description: "Greet the user",
            Category:    "custom",
            Handler: v2.SimpleCommandHandler(func(ctx context.Context, req *v2.ExecuteRequest) (*v2.ExecuteResponse, error) {
                return &v2.ExecuteResponse{
                    ExitCode: 0,
                    Output:   fmt.Sprintf("Hello from %s!", p.Metadata().Name),
                }, nil
            }),
        },
    }
}

func main() {
    plugin := &MyPlugin{}
    if err := v2.Serve(plugin); err != nil {
        os.Exit(1)
    }
}
```

## Plugin Architecture

### Plugin Interface

All plugins implement the `Plugin[C]` interface where `C` is your configuration type:

```go
type Plugin[C any] interface {
    // Metadata returns plugin information
    Metadata() Metadata

    // ConfigSchema returns JSON Schema for config validation (optional)
    ConfigSchema() map[string]interface{}

    // Configure initializes the plugin with config
    Configure(ctx context.Context, config C) error

    // Commands returns available commands
    Commands() []Command

    // Lifecycle methods
    Init(ctx context.Context) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    HealthCheck(ctx context.Context) error
}
```

Most plugins should embed `BasePlugin[C]` which provides sensible defaults for all methods.

### Plugin Metadata

The `Metadata` structure provides information about your plugin:

```go
type Metadata struct {
    Name         string
    Version      string
    Author       string
    Description  string
    License      string
    Homepage     string
    Dependencies []Dependency
}

func (p *MyPlugin) Metadata() v2.Metadata {
    return v2.Metadata{
        Name:        "my-plugin",
        Version:     "1.0.0",
        Author:      "Your Team",
        Description: "Database management plugin",
        License:     "MIT",
        Dependencies: []v2.Dependency{
            {Name: "docker", Version: "^1.0.0"},
        },
    }
}
```

### Plugin Dependencies

Plugins can declare dependencies on other plugins:

```go
func (p *MyPlugin) Metadata() v2.Metadata {
    return v2.Metadata{
        Name:    "my-plugin",
        Version: "1.0.0",
        Dependencies: []v2.Dependency{
            {Name: "docker", Version: "^1.0.0", Optional: false},
            {Name: "kubernetes", Version: ">=2.0.0", Optional: true},
        },
    }
}
```

#### Version Constraints

- **Exact version**: `"1.2.3"` - Requires exactly version 1.2.3
- **Caret range**: `"^1.2.3"` - Compatible with version 1.x.x (>=1.2.3 <2.0.0)
- **Tilde range**: `"~1.2.3"` - Compatible with version 1.2.x (>=1.2.3 <1.3.0)
- **Comparison**: `">=1.0.0"`, `">1.0.0"`, `"<=2.0.0"`, `"<2.0.0"`
- **Range**: `">=1.0.0 <2.0.0"` - Between versions

## Creating a Plugin

### Project Structure

```
glide-plugin-myname/
├── go.mod
├── go.sum
├── main.go
├── plugin/
│   ├── plugin.go      # Main plugin implementation
│   ├── config.go      # Configuration types
│   └── commands.go    # Command implementations
└── README.md
```

### Initialize Project

```bash
mkdir glide-plugin-myname
cd glide-plugin-myname
go mod init github.com/yourusername/glide-plugin-myname
go get github.com/ivannovak/glide/v2@latest
```

## Type-Safe Configuration

Define configuration as a Go struct with validation tags:

```go
type Config struct {
    // Required field
    APIKey string `json:"apiKey" yaml:"apiKey" validate:"required"`

    // Numeric constraints
    Timeout int `json:"timeout" yaml:"timeout" validate:"min=1,max=300"`

    // String constraints
    Environment string `json:"environment" validate:"oneof=dev staging prod"`

    // Optional with default (set in DefaultConfig)
    MaxRetries int `json:"maxRetries" yaml:"maxRetries"`

    // Boolean
    EnableSSL bool `json:"enableSSL" yaml:"enableSSL"`
}

// DefaultConfig returns sensible defaults
func DefaultConfig() Config {
    return Config{
        Timeout:    30,
        MaxRetries: 3,
        EnableSSL:  true,
    }
}
```

### Config Schema (Optional)

Provide a JSON Schema for enhanced validation:

```go
func (p *MyPlugin) ConfigSchema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "apiKey": map[string]interface{}{
                "type":        "string",
                "description": "API key for authentication",
                "minLength":   1,
            },
            "timeout": map[string]interface{}{
                "type":    "integer",
                "default": 30,
                "minimum": 1,
                "maximum": 300,
            },
        },
        "required": []string{"apiKey"},
    }
}
```

### User Configuration

Users configure plugins in `.glide.yml`:

```yaml
plugins:
  my-plugin:
    apiKey: "your-key-here"
    timeout: 60
    enableSSL: true
```

## Lifecycle Management

SDK v2 provides unified lifecycle management:

```go
// Init is called once after plugin load
func (p *MyPlugin) Init(ctx context.Context) error {
    // One-time initialization
    p.logger = logging.New("my-plugin")
    return nil
}

// Start is called when plugin should begin operation
func (p *MyPlugin) Start(ctx context.Context) error {
    // Connect to services, start workers
    var err error
    p.client, err = NewAPIClient(p.GetConfig().APIKey)
    return err
}

// Stop is called for graceful shutdown
func (p *MyPlugin) Stop(ctx context.Context) error {
    // Cleanup resources
    if p.client != nil {
        return p.client.Close()
    }
    return nil
}

// HealthCheck returns nil if healthy
func (p *MyPlugin) HealthCheck(ctx context.Context) error {
    return p.client.Ping(ctx)
}
```

### State Machine

Plugins follow a defined state machine:

```
Discovered → Loading → Initializing → Ready → Running → Stopping → Stopped
                ↓           ↓           ↓
              Error       Error       Error
```

## Command Registration

### Declarative Commands

```go
func (p *MyPlugin) Commands() []v2.Command {
    return []v2.Command{
        {
            Name:        "status",
            Description: "Check plugin status",
            Category:    "core",
            Aliases:     []string{"st"},
            Handler: v2.SimpleCommandHandler(p.statusCommand),
        },
        {
            Name:        "deploy",
            Description: "Deploy the application",
            Category:    "deployment",
            Flags: []v2.Flag{
                {Name: "environment", Short: "e", Description: "Target environment", Default: "staging"},
                {Name: "force", Short: "f", Description: "Force deployment", Type: v2.FlagBool},
            },
            Handler: v2.SimpleCommandHandler(p.deployCommand),
        },
    }
}

func (p *MyPlugin) statusCommand(ctx context.Context, req *v2.ExecuteRequest) (*v2.ExecuteResponse, error) {
    status := p.GetStatus()
    return &v2.ExecuteResponse{
        ExitCode: 0,
        Output:   fmt.Sprintf("Status: %s", status),
    }, nil
}

func (p *MyPlugin) deployCommand(ctx context.Context, req *v2.ExecuteRequest) (*v2.ExecuteResponse, error) {
    env := req.Flags["environment"]
    force := req.Flags["force"] == "true"

    if err := p.Deploy(env, force); err != nil {
        return &v2.ExecuteResponse{
            ExitCode: 1,
            Error:    err.Error(),
        }, nil
    }

    return &v2.ExecuteResponse{
        ExitCode: 0,
        Output:   fmt.Sprintf("Deployed to %s", env),
    }, nil
}
```

### Command Categories

| Category | ID | Priority | Description |
|----------|-----|----------|-------------|
| Core Commands | `core` | 10 | Essential commands |
| Project Commands | `project` | 20 | Multi-worktree management |
| Setup & Config | `setup` | 30 | Setup and configuration |
| Docker | `docker` | 40 | Container management |
| Testing | `testing` | 50 | Test execution |
| Development | `developer` | 60 | Code quality tools |
| Database | `database` | 70 | Database management |
| Help | `help` | 90 | Documentation |

### Custom Categories

```go
func (p *MyPlugin) CustomCategories() []v2.Category {
    return []v2.Category{
        {
            ID:          "infrastructure",
            Name:        "Infrastructure Management",
            Description: "AWS, Terraform, and cloud resources",
            Priority:    110, // 100-199 for plugins
        },
    }
}
```

## Using Command Aliases

### Plugin-Level Aliases

```go
func (p *MyPlugin) Metadata() v2.Metadata {
    return v2.Metadata{
        Name:    "database",
        Aliases: []string{"db", "d"},
        // ...
    }
}
```

Users can then use:
- `glide database status`
- `glide db status`
- `glide d status`

### Command-Level Aliases

```go
{
    Name:        "migrate",
    Aliases:     []string{"m"},
    Description: "Run database migrations",
    // ...
}
```

Users can combine aliases:
- `glide database migrate`
- `glide db m`
- `glide d m`

## Best Practices

### 1. Error Handling

```go
import "github.com/ivannovak/glide/v2/pkg/errors"

func (p *MyPlugin) executeCommand(ctx context.Context, req *v2.ExecuteRequest) (*v2.ExecuteResponse, error) {
    if err := p.doSomething(); err != nil {
        return &v2.ExecuteResponse{
            ExitCode: 1,
            Error: errors.Wrap(err, "failed to execute",
                errors.WithSuggestions(
                    "Check your configuration",
                    "Verify network connectivity",
                ),
            ).Error(),
        }, nil
    }
    return &v2.ExecuteResponse{ExitCode: 0, Output: "Success"}, nil
}
```

### 2. Output Management

```go
import "github.com/ivannovak/glide/v2/pkg/output"

func (p *MyPlugin) executeCommand(ctx context.Context, req *v2.ExecuteRequest) (*v2.ExecuteResponse, error) {
    // For structured output
    result := map[string]interface{}{
        "status": "healthy",
        "uptime": "24h",
    }

    // Format based on user preference
    format := req.Flags["format"]
    var out string
    switch format {
    case "json":
        data, _ := json.Marshal(result)
        out = string(data)
    default:
        out = fmt.Sprintf("Status: %s\nUptime: %s", result["status"], result["uptime"])
    }

    return &v2.ExecuteResponse{ExitCode: 0, Output: out}, nil
}
```

### 3. Configuration Validation

```go
func (p *MyPlugin) Configure(ctx context.Context, config Config) error {
    // Store config
    if err := p.BasePlugin.Configure(ctx, config); err != nil {
        return err
    }

    // Additional validation
    if config.APIKey != "" && len(config.APIKey) < 10 {
        return fmt.Errorf("API key must be at least 10 characters")
    }

    return nil
}
```

### 4. Resource Management

```go
type MyPlugin struct {
    v2.BasePlugin[Config]
    client    *APIClient
    closeOnce sync.Once
}

func (p *MyPlugin) Stop(ctx context.Context) error {
    var err error
    p.closeOnce.Do(func() {
        if p.client != nil {
            err = p.client.Close()
        }
    })
    return err
}
```

## Examples

### Complete Database Plugin

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/ivannovak/glide/v2/pkg/plugin/sdk/v2"
)

type Config struct {
    Connection string `json:"connection" yaml:"connection" validate:"required"`
    Migrations string `json:"migrations" yaml:"migrations"`
}

type DatabasePlugin struct {
    v2.BasePlugin[Config]
    db *sql.DB
}

func (p *DatabasePlugin) Metadata() v2.Metadata {
    return v2.Metadata{
        Name:        "database",
        Version:     "1.0.0",
        Author:      "Your Team",
        Description: "Database management plugin",
        Aliases:     []string{"db", "d"},
    }
}

func (p *DatabasePlugin) Configure(ctx context.Context, config Config) error {
    return p.BasePlugin.Configure(ctx, config)
}

func (p *DatabasePlugin) Start(ctx context.Context) error {
    var err error
    p.db, err = sql.Open("postgres", p.GetConfig().Connection)
    return err
}

func (p *DatabasePlugin) Stop(ctx context.Context) error {
    if p.db != nil {
        return p.db.Close()
    }
    return nil
}

func (p *DatabasePlugin) HealthCheck(ctx context.Context) error {
    return p.db.PingContext(ctx)
}

func (p *DatabasePlugin) Commands() []v2.Command {
    return []v2.Command{
        {
            Name:        "migrate",
            Aliases:     []string{"m"},
            Description: "Run database migrations",
            Category:    "database",
            Handler:     v2.SimpleCommandHandler(p.migrate),
        },
        {
            Name:        "status",
            Aliases:     []string{"st"},
            Description: "Check database status",
            Category:    "database",
            Handler:     v2.SimpleCommandHandler(p.status),
        },
    }
}

func (p *DatabasePlugin) migrate(ctx context.Context, req *v2.ExecuteRequest) (*v2.ExecuteResponse, error) {
    // Migration logic
    return &v2.ExecuteResponse{
        ExitCode: 0,
        Output:   "Migrations completed successfully",
    }, nil
}

func (p *DatabasePlugin) status(ctx context.Context, req *v2.ExecuteRequest) (*v2.ExecuteResponse, error) {
    if err := p.db.PingContext(ctx); err != nil {
        return &v2.ExecuteResponse{
            ExitCode: 1,
            Error:    fmt.Sprintf("Database unreachable: %v", err),
        }, nil
    }
    return &v2.ExecuteResponse{
        ExitCode: 0,
        Output:   "Database connection: OK",
    }, nil
}

func main() {
    if err := v2.Serve(&DatabasePlugin{}); err != nil {
        os.Exit(1)
    }
}
```

### Testing Your Plugin

```go
package main

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestPluginMetadata(t *testing.T) {
    plugin := &DatabasePlugin{}
    meta := plugin.Metadata()

    assert.Equal(t, "database", meta.Name)
    assert.Equal(t, "1.0.0", meta.Version)
    assert.Contains(t, meta.Aliases, "db")
}

func TestPluginConfigure(t *testing.T) {
    plugin := &DatabasePlugin{}
    ctx := context.Background()

    err := plugin.Configure(ctx, Config{
        Connection: "postgres://localhost/test",
    })
    require.NoError(t, err)
}

func TestPluginCommands(t *testing.T) {
    plugin := &DatabasePlugin{}
    commands := plugin.Commands()

    assert.Len(t, commands, 2)
    assert.Equal(t, "migrate", commands[0].Name)
    assert.Equal(t, "status", commands[1].Name)
}
```

## Plugin Installation

### Build and Install

```bash
# Build
go build -o glide-plugin-database

# Install globally
cp glide-plugin-database ~/.glide/plugins/

# Or install for project
mkdir -p .glide/plugins
cp glide-plugin-database .glide/plugins/

# Verify
glide plugins list
```

### Plugin Discovery Locations

1. `~/.glide/plugins/` - User plugins
2. `.glide/plugins/` in ancestor directories
3. `./.glide/plugins/` - Current directory
4. `/usr/local/lib/glide/plugins/` - System plugins

## Further Reading

- [Tutorial: Creating Your First Plugin](tutorials/02-first-plugin.md)
- [SDK v2 API Reference](../pkg/plugin/sdk/v2/)
- [Error Handling Guide](guides/error-handling.md)
- [Architecture Overview](architecture/)
