# Tutorial 2: Creating Your First Plugin

In this tutorial, you'll build a complete Glide plugin from scratch using SDK v2. By the end, you'll have a working plugin that adds custom commands to Glide.

## What You'll Learn

- Plugin project structure
- Implementing the Plugin interface
- Type-safe configuration
- Building and installing plugins

## Prerequisites

- Completed [Tutorial 1: Getting Started](./01-getting-started.md)
- Go 1.21 or later
- Basic Go programming knowledge

## Step 1: Create Plugin Project

### Initialize the Project

```bash
mkdir glide-plugin-hello
cd glide-plugin-hello
go mod init github.com/yourusername/glide-plugin-hello
```

### Add Dependencies

```bash
go get github.com/ivannovak/glide/v2@latest
```

### Project Structure

Create this structure:
```
glide-plugin-hello/
├── go.mod
├── go.sum
├── main.go
└── plugin/
    └── plugin.go
```

## Step 2: Define Configuration

Create `plugin/plugin.go`:

```go
package plugin

import (
    "context"
    "fmt"

    "github.com/ivannovak/glide/v2/pkg/plugin/sdk/v2"
)

// Config defines the plugin's configuration
type Config struct {
    // Greeting prefix (e.g., "Hello", "Hi", "Hey")
    Greeting string `json:"greeting" yaml:"greeting"`

    // Include timestamp in output
    ShowTime bool `json:"show_time" yaml:"show_time"`

    // Custom suffix (e.g., "!", ".")
    Suffix string `json:"suffix" yaml:"suffix"`
}

// DefaultConfig returns sensible defaults
func DefaultConfig() Config {
    return Config{
        Greeting: "Hello",
        ShowTime: false,
        Suffix:   "!",
    }
}
```

## Step 3: Implement the Plugin

Add to `plugin/plugin.go`:

```go
// HelloPlugin is our main plugin type
type HelloPlugin struct {
    v2.BasePlugin[Config]
}

// New creates a new HelloPlugin instance
func New() *HelloPlugin {
    return &HelloPlugin{}
}

// Metadata returns plugin information
func (p *HelloPlugin) Metadata() v2.Metadata {
    return v2.Metadata{
        Name:        "hello",
        Version:     "1.0.0",
        Description: "A friendly greeting plugin for Glide",
        Author:      "Your Name",
        Homepage:    "https://github.com/yourusername/glide-plugin-hello",
    }
}

// ConfigSchema returns JSON Schema for configuration validation
func (p *HelloPlugin) ConfigSchema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "greeting": map[string]interface{}{
                "type":        "string",
                "description": "Greeting prefix",
                "default":     "Hello",
            },
            "show_time": map[string]interface{}{
                "type":        "boolean",
                "description": "Include timestamp",
                "default":     false,
            },
            "suffix": map[string]interface{}{
                "type":        "string",
                "description": "Message suffix",
                "default":     "!",
            },
        },
    }
}

// Configure is called with the parsed configuration
func (p *HelloPlugin) Configure(ctx context.Context, config Config) error {
    // Store config in BasePlugin
    if err := p.Init(config); err != nil {
        return err
    }

    // Validate configuration
    if config.Greeting == "" {
        config.Greeting = "Hello"
    }

    return nil
}
```

## Step 4: Add Commands

Add commands to `plugin/plugin.go`:

```go
import (
    "context"
    "fmt"
    "time"

    "github.com/ivannovak/glide/v2/pkg/plugin/sdk/v2"
)

// Commands returns the commands this plugin provides
func (p *HelloPlugin) Commands() []v2.Command {
    return []v2.Command{
        {
            Name:        "hello",
            Description: "Say hello to someone",
            Usage:       "hello [name]",
            Run:         p.runHello,
        },
        {
            Name:        "goodbye",
            Description: "Say goodbye to someone",
            Usage:       "goodbye [name]",
            Run:         p.runGoodbye,
        },
    }
}

func (p *HelloPlugin) runHello(ctx context.Context, args []string) error {
    cfg := p.GetConfig()
    name := "World"
    if len(args) > 0 {
        name = args[0]
    }

    message := fmt.Sprintf("%s, %s%s", cfg.Greeting, name, cfg.Suffix)

    if cfg.ShowTime {
        message = fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), message)
    }

    fmt.Println(message)
    return nil
}

func (p *HelloPlugin) runGoodbye(ctx context.Context, args []string) error {
    cfg := p.GetConfig()
    name := "World"
    if len(args) > 0 {
        name = args[0]
    }

    message := fmt.Sprintf("Goodbye, %s%s", name, cfg.Suffix)

    if cfg.ShowTime {
        message = fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), message)
    }

    fmt.Println(message)
    return nil
}
```

## Step 5: Create Entry Point

Create `main.go`:

```go
package main

import (
    "os"

    "github.com/ivannovak/glide/v2/pkg/plugin/sdk/v2"
    "github.com/yourusername/glide-plugin-hello/plugin"
)

func main() {
    p := plugin.New()

    if err := v2.Serve(p); err != nil {
        os.Exit(1)
    }
}
```

## Step 6: Build the Plugin

```bash
# Build the plugin
go build -o glide-plugin-hello .

# Verify it runs
./glide-plugin-hello --help
```

## Step 7: Install the Plugin

### Install Globally

```bash
# Copy to global plugin directory
mkdir -p ~/.glide/plugins
cp glide-plugin-hello ~/.glide/plugins/
```

### Install Locally (Project-specific)

```bash
# Copy to project plugin directory
mkdir -p /path/to/project/.glide/plugins
cp glide-plugin-hello /path/to/project/.glide/plugins/
```

## Step 8: Configure the Plugin

In your project's `.glide.yml`:

```yaml
# Plugin configuration
plugins:
  hello:
    greeting: "Hey"
    show_time: true
    suffix: "!"
```

## Step 9: Use Your Plugin

```bash
# Navigate to a Glide project
cd /path/to/project

# Use the hello command
glide hello
# Output: [14:30:45] Hey, World!

glide hello Alice
# Output: [14:30:47] Hey, Alice!

glide goodbye Bob
# Output: [14:30:50] Goodbye, Bob!
```

## Step 10: Add Lifecycle Hooks (Optional)

Enhance your plugin with lifecycle management:

```go
// OnStart is called when the plugin starts
func (p *HelloPlugin) OnStart(ctx context.Context) error {
    fmt.Println("Hello plugin started!")
    return nil
}

// OnStop is called during shutdown
func (p *HelloPlugin) OnStop(ctx context.Context) error {
    fmt.Println("Hello plugin stopping...")
    return nil
}

// HealthCheck reports plugin health
func (p *HelloPlugin) HealthCheck(ctx context.Context) error {
    // Check any dependencies here
    return nil
}
```

## Complete Code

Here's the complete `plugin/plugin.go`:

```go
package plugin

import (
    "context"
    "fmt"
    "time"

    "github.com/ivannovak/glide/v2/pkg/plugin/sdk/v2"
)

// Config defines the plugin's configuration
type Config struct {
    Greeting string `json:"greeting" yaml:"greeting"`
    ShowTime bool   `json:"show_time" yaml:"show_time"`
    Suffix   string `json:"suffix" yaml:"suffix"`
}

// HelloPlugin is our main plugin type
type HelloPlugin struct {
    v2.BasePlugin[Config]
}

// New creates a new HelloPlugin instance
func New() *HelloPlugin {
    return &HelloPlugin{}
}

// Metadata returns plugin information
func (p *HelloPlugin) Metadata() v2.Metadata {
    return v2.Metadata{
        Name:        "hello",
        Version:     "1.0.0",
        Description: "A friendly greeting plugin for Glide",
        Author:      "Your Name",
    }
}

// ConfigSchema returns JSON Schema for configuration validation
func (p *HelloPlugin) ConfigSchema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "greeting":  map[string]interface{}{"type": "string"},
            "show_time": map[string]interface{}{"type": "boolean"},
            "suffix":    map[string]interface{}{"type": "string"},
        },
    }
}

// Configure is called with the parsed configuration
func (p *HelloPlugin) Configure(ctx context.Context, config Config) error {
    if config.Greeting == "" {
        config.Greeting = "Hello"
    }
    if config.Suffix == "" {
        config.Suffix = "!"
    }
    return p.Init(config)
}

// Commands returns the commands this plugin provides
func (p *HelloPlugin) Commands() []v2.Command {
    return []v2.Command{
        {
            Name:        "hello",
            Description: "Say hello to someone",
            Run:         p.runHello,
        },
        {
            Name:        "goodbye",
            Description: "Say goodbye to someone",
            Run:         p.runGoodbye,
        },
    }
}

func (p *HelloPlugin) runHello(ctx context.Context, args []string) error {
    cfg := p.GetConfig()
    name := "World"
    if len(args) > 0 {
        name = args[0]
    }

    message := fmt.Sprintf("%s, %s%s", cfg.Greeting, name, cfg.Suffix)
    if cfg.ShowTime {
        message = fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), message)
    }

    fmt.Println(message)
    return nil
}

func (p *HelloPlugin) runGoodbye(ctx context.Context, args []string) error {
    cfg := p.GetConfig()
    name := "World"
    if len(args) > 0 {
        name = args[0]
    }

    message := fmt.Sprintf("Goodbye, %s%s", name, cfg.Suffix)
    if cfg.ShowTime {
        message = fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), message)
    }

    fmt.Println(message)
    return nil
}
```

## What's Next?

You've built your first plugin! Continue learning:

1. **[Tutorial 3: Advanced Configuration](./03-advanced-configuration.md)** - Multi-project setups
2. **[Plugin Development Guide](../plugin-development.md)** - Complete reference
3. **[SDK v2 Migration Guide](../guides/PLUGIN-SDK-V2-MIGRATION.md)** - Advanced features

## Troubleshooting

### Plugin Not Found

Ensure the plugin is in the correct directory:
```bash
ls -la ~/.glide/plugins/
```

### Permission Denied

Make the plugin executable:
```bash
chmod +x ~/.glide/plugins/glide-plugin-hello
```

### Configuration Not Applied

Check your `.glide.yml` has the correct structure:
```yaml
plugins:
  hello:  # Must match plugin name from Metadata()
    greeting: "Hey"
```

## Summary

In this tutorial, you learned how to:
- Create a plugin project structure
- Define type-safe configuration
- Implement the Plugin interface
- Add commands with handlers
- Build and install the plugin
- Configure the plugin in `.glide.yml`

Plugins are powerful - they can add any functionality you need to Glide!
