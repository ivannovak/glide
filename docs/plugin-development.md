# Plugin Development Guide

## Table of Contents
- [Overview](#overview)
- [Plugin Architecture](#plugin-architecture)
- [Creating a Plugin](#creating-a-plugin)
- [Command Registration](#command-registration)
- [Using Command Aliases](#using-command-aliases)
- [Best Practices](#best-practices)
- [Examples](#examples)

## Overview

Glide supports a powerful plugin system that allows developers to extend the CLI with custom commands. Plugins can be built into the binary at compile time or loaded dynamically at runtime.

## Plugin Architecture

### Plugin Interface

All plugins must implement the `Plugin` interface:

```go
type Plugin interface {
    // Name returns the plugin identifier
    Name() string

    // Version returns the plugin version
    Version() string

    // Register adds plugin commands to the command tree
    Register(root *cobra.Command) error

    // Configure initializes the plugin
    // Plugins use pkg/config for type-safe configuration
    Configure() error

    // Metadata returns plugin information
    Metadata() PluginMetadata
}
```

### Plugin Metadata

The `PluginMetadata` structure provides information about your plugin:

```go
type PluginMetadata struct {
    Name        string
    Version     string
    Author      string
    Description string
    Commands    []CommandInfo
    BuildTags   []string // Required build tags
    ConfigKeys  []string // Configuration keys used
}

type CommandInfo struct {
    Name        string
    Category    string
    Description string
    Aliases     []string // Command aliases
}
```

## Creating a Plugin

### Basic Plugin Structure

```go
package myplugin

import (
    "fmt"
    "github.com/ivannovak/glide/pkg/plugin"
    "github.com/ivannovak/glide/pkg/config"
    "github.com/spf13/cobra"
)

// MyPluginConfig defines typed configuration for this plugin
type MyPluginConfig struct {
    Endpoint string `json:"endpoint" yaml:"endpoint" validate:"required,url"`
    Timeout  int    `json:"timeout" yaml:"timeout" validate:"min=1,max=300"`
}

type MyPlugin struct {
    endpoint string
    timeout  int
}

func init() {
    // Register typed configuration with defaults
    config.Register("myplugin", MyPluginConfig{
        Timeout: 30, // Default timeout
    })
}

func New() plugin.Plugin {
    return &MyPlugin{}
}

func (p *MyPlugin) Name() string {
    return "myplugin"
}

func (p *MyPlugin) Version() string {
    return "1.0.0"
}

func (p *MyPlugin) Configure() error {
    // Get typed config from registry (populated by config loader from YAML)
    cfg, err := config.GetValue[MyPluginConfig]("myplugin")
    if err != nil {
        return fmt.Errorf("failed to get plugin config: %w", err)
    }

    p.endpoint = cfg.Endpoint
    p.timeout = cfg.Timeout
    return nil
}

func (p *MyPlugin) Metadata() plugin.PluginMetadata {
    return plugin.PluginMetadata{
        Name:        p.Name(),
        Version:     p.Version(),
        Author:      "Your Name",
        Description: "My custom plugin",
        Commands: []plugin.CommandInfo{
            {
                Name:        "mycommand",
                Category:    "custom",
                Description: "Execute my custom command",
                Aliases:     []string{"mc"}, // Add aliases here
            },
        },
        ConfigKeys: []string{"endpoint", "timeout"},
    }
}
```

## Command Registration

### Registering Commands with the CLI

When your plugin is loaded, its `Register` method is called with the root command:

```go
func (p *MyPlugin) Register(root *cobra.Command) error {
    // Create your command
    cmd := &cobra.Command{
        Use:     "mycommand",
        Aliases: []string{"mc"}, // Define aliases
        Short:   "Execute my custom command",
        Long:    `A longer description of what your command does.`,
        RunE:    p.executeCommand,
    }

    // Add flags
    cmd.Flags().StringP("output", "o", "text", "Output format (text, json)")

    // Add to root
    root.AddCommand(cmd)

    return nil
}

func (p *MyPlugin) executeCommand(cmd *cobra.Command, args []string) error {
    // Your command implementation
    return nil
}
```

## Using Aliases

### What are Aliases?

Aliases provide short, memorable alternatives to both plugin names and command names. They allow users to type less while maintaining clarity.

### Plugin-Level Aliases

Plugin-level aliases allow users to access your entire plugin with a shorter name. For example, if your plugin is named `database`, you can provide aliases like `db` or `d`.

#### Adding Plugin Aliases

Define plugin aliases in your `PluginMetadata`:

```go
func (p *MyPlugin) Metadata() plugin.PluginMetadata {
    return plugin.PluginMetadata{
        Name:        "database",
        Version:     "1.0.0",
        Author:      "Your Name",
        Description: "Database management plugin",
        Aliases:     []string{"db", "d"}, // Plugin-level aliases
        Commands: []plugin.CommandInfo{
            // Command definitions...
        },
    }
}
```

With plugin aliases, users can use:
- `glide database status` (full name)
- `glide db status` (using alias)
- `glide d status` (using shorter alias)

### Command-Level Aliases

Command aliases provide shortcuts for individual commands within your plugin.

#### Adding Command Aliases

There are two places to define command aliases:

1. **In the CommandInfo** (for documentation):
```go
Commands: []plugin.CommandInfo{
    {
        Name:        "migrate",
        Aliases:     []string{"m"},
        Description: "Run database migrations",
    },
}
```

2. **In the Cobra Command** (for functionality):
```go
cmd := &cobra.Command{
    Use:     "migrate",
    Aliases: []string{"m"},
    Short:   "Run database migrations",
}
```

### Alias Best Practices

1. **Keep aliases short**: 1-3 characters is ideal
2. **Make them memorable**: Use logical abbreviations
3. **Avoid conflicts**: Check existing commands and aliases
4. **Be consistent**: Use similar patterns across your plugin
5. **Document them**: List aliases in help text

### Example: Multi-Command Plugin with Aliases

```go
func (p *MyPlugin) Register(root *cobra.Command) error {
    // Main plugin command group
    pluginCmd := &cobra.Command{
        Use:     "myplugin",
        Aliases: []string{"mp"},
        Short:   "My plugin commands",
    }

    // Subcommand with aliases
    dbCmd := &cobra.Command{
        Use:     "database",
        Aliases: []string{"db", "d"},
        Short:   "Database operations",
    }

    // Database subcommands
    dbMigrateCmd := &cobra.Command{
        Use:     "migrate",
        Aliases: []string{"m"},
        Short:   "Run database migrations",
        RunE:    p.migrate,
    }

    dbSeedCmd := &cobra.Command{
        Use:     "seed",
        Aliases: []string{"s"},
        Short:   "Seed the database",
        RunE:    p.seed,
    }

    // Build command tree
    dbCmd.AddCommand(dbMigrateCmd)
    dbCmd.AddCommand(dbSeedCmd)
    pluginCmd.AddCommand(dbCmd)
    root.AddCommand(pluginCmd)

    return nil
}
```

With this setup, users can use any of these commands:
- `glide myplugin database migrate`
- `glide mp db m` (using all aliases)
- `glide mp database migrate` (mixing full names and aliases)

### Combining Plugin and Command Aliases

When both plugin and command aliases are defined, they can be combined for maximum brevity:

```go
// Example with plugin "database" (aliases: "db", "d")
// and command "migrate" (alias: "m")

// All these are equivalent:
glide database migrate --fresh
glide db migrate --fresh      // Plugin alias
glide d migrate --fresh       // Shorter plugin alias
glide database m --fresh      // Command alias
glide db m --fresh           // Both aliases
glide d m --fresh           // Shortest form
```

## Best Practices

### 1. Command Naming
- Use clear, descriptive names
- Follow existing CLI conventions
- Group related commands under a parent command

### 2. Error Handling
```go
func (p *MyPlugin) executeCommand(cmd *cobra.Command, args []string) error {
    if len(args) < 1 {
        return fmt.Errorf("required argument missing")
    }

    // Use Glide's error types for consistency
    if err := p.doSomething(); err != nil {
        return glideErrors.Wrap(err, "failed to execute command",
            glideErrors.WithSuggestions(
                "Check your configuration",
                "Try running with --verbose",
            ),
        )
    }

    return nil
}
```

### 3. Configuration
```go
// Define typed configuration struct
type MyPluginConfig struct {
    APIKey  string `json:"api_key" yaml:"api_key" validate:"required"`
    Timeout int    `json:"timeout" yaml:"timeout" validate:"min=1,max=300"`
}

func init() {
    // Register with defaults in plugin's init()
    config.Register("myplugin", MyPluginConfig{
        Timeout: 30, // default
    })
}

func (p *MyPlugin) Configure() error {
    // Get typed config (populated by config loader from YAML)
    cfg, err := config.GetValue[MyPluginConfig]("myplugin")
    if err != nil {
        return fmt.Errorf("failed to get plugin config: %w", err)
    }

    // Use the validated, type-safe configuration
    p.apiKey = cfg.APIKey
    p.timeout = cfg.Timeout

    return nil
}
```

Users configure plugins in `.glide.yml`:
```yaml
plugins:
  myplugin:
    api_key: "your-key-here"
    timeout: 60
```

### 4. Output Management
```go
import "github.com/ivannovak/glide/pkg/output"

func (p *MyPlugin) executeCommand(cmd *cobra.Command, args []string) error {
    // Use Glide's output manager
    output.Info("Starting operation...")

    // Show progress
    spinner := progress.NewSpinner("Processing")
    spinner.Start()
    defer spinner.Stop()

    // Do work...

    spinner.Success("Operation completed")
    return nil
}
```

## Examples

### Complete Example: Database Plugin

```go
package database

import (
    "fmt"
    "github.com/ivannovak/glide/pkg/plugin"
    "github.com/ivannovak/glide/pkg/config"
    "github.com/ivannovak/glide/pkg/output"
    "github.com/spf13/cobra"
)

// DatabasePluginConfig defines typed configuration
type DatabasePluginConfig struct {
    Connection string `json:"connection" yaml:"connection" validate:"required"`
    Migrations string `json:"migrations" yaml:"migrations"`
}

type DatabasePlugin struct {
    connString string
    migrations string
}

func init() {
    // Register typed configuration with defaults
    config.Register("database", DatabasePluginConfig{
        Migrations: "./migrations",
    })
}

func New() plugin.Plugin {
    return &DatabasePlugin{}
}

func (p *DatabasePlugin) Name() string {
    return "database"
}

func (p *DatabasePlugin) Version() string {
    return "1.0.0"
}

func (p *DatabasePlugin) Configure() error {
    cfg, err := config.GetValue[DatabasePluginConfig]("database")
    if err != nil {
        return fmt.Errorf("failed to get database config: %w", err)
    }

    p.connString = cfg.Connection
    p.migrations = cfg.Migrations
    return nil
}

func (p *DatabasePlugin) Register(root *cobra.Command) error {
    // Main database command
    dbCmd := &cobra.Command{
        Use:     "database",
        Aliases: []string{"db", "d"},
        Short:   "Database management commands",
    }

    // Migrate command
    migrateCmd := &cobra.Command{
        Use:     "migrate",
        Aliases: []string{"m"},
        Short:   "Run database migrations",
        RunE: func(cmd *cobra.Command, args []string) error {
            output.Info("Running migrations...")
            // Migration logic here
            output.Success("Migrations completed")
            return nil
        },
    }

    // Seed command
    seedCmd := &cobra.Command{
        Use:     "seed",
        Aliases: []string{"s"},
        Short:   "Seed the database",
        RunE: func(cmd *cobra.Command, args []string) error {
            output.Info("Seeding database...")
            // Seeding logic here
            output.Success("Database seeded")
            return nil
        },
    }

    // Status command
    statusCmd := &cobra.Command{
        Use:     "status",
        Aliases: []string{"st"},
        Short:   "Check database status",
        RunE: func(cmd *cobra.Command, args []string) error {
            output.Info("Database Status:")
            output.Info("  Connection: %s", p.connString)
            // Check status logic
            return nil
        },
    }

    // Build command tree
    dbCmd.AddCommand(migrateCmd, seedCmd, statusCmd)
    root.AddCommand(dbCmd)

    return nil
}

func (p *DatabasePlugin) Metadata() plugin.PluginMetadata {
    return plugin.PluginMetadata{
        Name:        p.Name(),
        Version:     p.Version(),
        Author:      "Your Team",
        Description: "Database management plugin",
        Aliases:     []string{"db", "d"}, // Plugin-level aliases
        Commands: []plugin.CommandInfo{
            {
                Name:        "migrate",
                Aliases:     []string{"m"},
                Description: "Run migrations",
            },
            {
                Name:        "seed",
                Aliases:     []string{"s"},
                Description: "Seed database",
            },
            {
                Name:        "status",
                Aliases:     []string{"st"},
                Description: "Check status",
            },
        },
        ConfigKeys: []string{"connection", "timeout"},
    }
}
```

### Testing Your Plugin

```go
package database

import (
    "testing"
    "github.com/spf13/cobra"
    "github.com/stretchr/testify/assert"
)

func TestPluginRegistration(t *testing.T) {
    plugin := New()
    root := &cobra.Command{Use: "test"}

    err := plugin.Register(root)
    assert.NoError(t, err)

    // Check command was added
    dbCmd, _, err := root.Find([]string{"database"})
    assert.NoError(t, err)
    assert.NotNil(t, dbCmd)

    // Check aliases work
    dbCmd, _, err = root.Find([]string{"db"})
    assert.NoError(t, err)
    assert.NotNil(t, dbCmd)

    // Check subcommand
    _, _, err = root.Find([]string{"db", "migrate"})
    assert.NoError(t, err)

    // Check subcommand alias
    _, _, err = root.Find([]string{"db", "m"})
    assert.NoError(t, err)
}
```

## Loading Your Plugin

### Build-Time Plugin

Add your plugin to the build tags and register it:

```go
// cmd/glid/plugins_custom.go
// +build custom

package main

import (
    "github.com/ivannovak/glide/pkg/plugin"
    "your-module/plugins/database"
)

func init() {
    plugin.Register(database.New())
}
```

Build with tags:
```bash
go build -tags custom -o glide cmd/glide/main.go
```

### Runtime Plugin (Future)

Place your compiled plugin in `~/.glide/plugins/` and it will be loaded automatically.

## Troubleshooting

### Common Issues

1. **Alias conflicts**: Check existing commands with `glide --help`
2. **Registration errors**: Ensure unique command names
3. **Configuration issues**: Validate required config keys
4. **Missing dependencies**: Check your go.mod file

### Debug Output

Enable debug mode to see plugin loading:
```bash
GLIDE_DEBUG=true glide your-command
```

## Framework Detection

Plugins can contribute framework detectors to automatically detect programming languages, frameworks, and tools in projects. See the [Framework Detection Guide](./framework-detection.md) for details on:

- Creating custom framework detectors
- Using the BaseFrameworkDetector for pattern-based detection
- Registering detectors with your plugin
- Command injection based on detected frameworks

## Further Reading

- [Cobra Documentation](https://github.com/spf13/cobra)
- [Glide Architecture](./architecture.md)
- [Framework Detection](./framework-detection.md)
- [Error Handling Guide](./error-handling.md)
- [Testing Plugins](./testing.md)
