# Plugin Development Guide

## Overview

Glide supports runtime plugins that extend the CLI with custom commands. Plugins are standalone binaries that communicate with Glide via gRPC using **SDK v2**.

For comprehensive documentation, see [docs/plugin-development.md](docs/plugin-development.md).

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
    }
}

// Configure is called with type-safe config
func (p *MyPlugin) Configure(ctx context.Context, config MyConfig) error {
    return p.BasePlugin.Configure(ctx, config)
}

// Commands returns the list of commands this plugin provides
func (p *MyPlugin) Commands() []v2.Command {
    return []v2.Command{
        {
            Name:        "greet",
            Description: "Greet the user",
            Category:    "custom",
            Handler: v2.SimpleCommandHandler(func(ctx context.Context, req *v2.ExecuteRequest) (*v2.ExecuteResponse, error) {
                name := "World"
                if len(req.Args) > 0 {
                    name = req.Args[0]
                }
                return &v2.ExecuteResponse{
                    ExitCode: 0,
                    Output:   fmt.Sprintf("Hello, %s!", name),
                }, nil
            }),
        },
    }
}

// Optional: Lifecycle hooks
func (p *MyPlugin) Init(ctx context.Context) error {
    // One-time initialization
    return nil
}

func (p *MyPlugin) Start(ctx context.Context) error {
    // Called when plugin starts
    return nil
}

func (p *MyPlugin) Stop(ctx context.Context) error {
    // Graceful shutdown
    return nil
}

func (p *MyPlugin) HealthCheck(ctx context.Context) error {
    // Return nil if healthy
    return nil
}

func main() {
    plugin := &MyPlugin{}
    if err := v2.Serve(plugin); err != nil {
        os.Exit(1)
    }
}
```

## Key Features

### Type-Safe Configuration

Define your configuration as a Go struct with validation tags:

```go
type Config struct {
    APIKey    string `json:"apiKey" validate:"required"`
    Timeout   int    `json:"timeout" validate:"min=1,max=300"`
    EnableSSL bool   `json:"enableSSL"`
}
```

Users configure your plugin in `.glide.yml`:

```yaml
plugins:
  my-plugin:
    apiKey: "your-key-here"
    timeout: 60
    enableSSL: true
```

### Lifecycle Management

SDK v2 provides unified lifecycle hooks:

| Hook | Purpose |
|------|---------|
| `Init(ctx)` | One-time initialization after load |
| `Start(ctx)` | Begin operation (connect to services) |
| `Stop(ctx)` | Graceful shutdown (cleanup resources) |
| `HealthCheck(ctx)` | Verify plugin health |

### Command Categories

Organize your commands into categories:

| Category | ID | Description |
|----------|-----|-------------|
| Core Commands | `core` | Essential development commands |
| Project Commands | `project` | Multi-worktree management |
| Setup & Config | `setup` | Project setup and configuration |
| Docker | `docker` | Container management |
| Testing | `testing` | Test execution |
| Development | `developer` | Code quality tools |
| Database | `database` | Database management |
| Help | `help` | Documentation |

Custom categories (priority 100-199):

```go
func (p *MyPlugin) CustomCategories() []v2.Category {
    return []v2.Category{
        {
            ID:          "infrastructure",
            Name:        "Infrastructure",
            Description: "Cloud and infrastructure commands",
            Priority:    110,
        },
    }
}
```

## Plugin Installation

### Plugin Discovery Locations

1. **User plugins**: `~/.glide/plugins/`
2. **Ancestor directories**: `.glide/plugins/` in parent directories
3. **Current directory**: `./.glide/plugins/`
4. **System plugins**: `/usr/local/lib/glide/plugins/`

### Naming Convention

Plugin binaries must follow: `glide-plugin-{name}`

Example: `glide-plugin-docker`, `glide-plugin-aws`

### Building and Installing

```bash
# Build
go build -o glide-plugin-myname

# Install globally
cp glide-plugin-myname ~/.glide/plugins/

# Verify
glide plugins list
```

## Best Practices

1. **Use type-safe configuration** - Define config structs with validation
2. **Implement lifecycle hooks** - Properly initialize and cleanup resources
3. **Choose appropriate categories** - Don't use `core` unless essential
4. **Handle errors gracefully** - Return meaningful error messages
5. **Follow naming conventions** - lowercase, hyphens for commands

## Documentation

- **[Full Plugin Development Guide](docs/plugin-development.md)** - Comprehensive documentation
- **[Tutorials](docs/tutorials/02-first-plugin.md)** - Step-by-step plugin creation
- **[SDK v2 API Reference](pkg/plugin/sdk/v2/)** - API documentation

## Example Plugins

- **[Docker Plugin](https://github.com/ivannovak/glide-plugin-docker)** - Production-ready reference
- **[Plugin Boilerplate](examples/plugin-boilerplate/)** - Starter template
