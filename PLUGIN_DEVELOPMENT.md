# Plugin Development Guide

## Overview
Glide supports runtime plugins that extend the CLI with custom commands. Plugins are standalone binaries that communicate with Glide via gRPC.

## Plugin Structure

### Basic Plugin Template
```go
package main

import (
    "context"
    sdk "github.com/ivannovak/glide/pkg/plugin/sdk/v1"
    "github.com/hashicorp/go-plugin"
)

type MyPlugin struct {
    sdk.UnimplementedGlidePluginServer
}

func (p *MyPlugin) GetMetadata(ctx context.Context, _ *sdk.Empty) (*sdk.PluginMetadata, error) {
    return &sdk.PluginMetadata{
        Name:        "myplugin",
        Version:     "1.0.0",
        Author:      "Your Name",
        Description: "Plugin description",
        MinSdk:      "v1.0.0",
        Namespaced:  false, // Set to true to prefix commands with plugin name
    }, nil
}

func main() {
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: sdk.HandshakeConfig,
        Plugins: map[string]plugin.Plugin{
            "glide": &sdk.GlidePluginImpl{
                Impl: &MyPlugin{},
            },
        },
        GRPCServer: plugin.DefaultGRPCServer,
    })
}
```

## Command Categories

Commands should be assigned to appropriate categories for proper organization in help output. Categories are displayed in priority order with consistent styling.

### Available Categories

| Category | ID | Priority | Description | Use For |
|----------|-----|----------|-------------|---------|
| **Core Commands** | `core` | 10 | Essential development commands | Version, plugins, self-update |
| **Global Commands** | `project` | 20 | Multi-worktree management | Commands that operate across all worktrees |
| **Setup & Configuration** | `setup` | 30 | Project setup and configuration | Initial setup, configuration, completions |
| **Docker Management** | `docker` | 40 | Container and service control | Docker compose operations, container management |
| **Testing** | `testing` | 50 | Test execution and coverage | Test runners, coverage tools |
| **Development Tools** | `developer` | 60 | Code quality and utilities | Linters, formatters, build tools |
| **Database** | `database` | 70 | Database management and access | Migrations, database CLI, backups |
| **Help & Documentation** | `help` | 90 | Help topics and guides | Documentation commands |

### Setting Command Category

```go
func (p *MyPlugin) ListCommands(ctx context.Context, _ *sdk.Empty) (*sdk.CommandList, error) {
    return &sdk.CommandList{
        Commands: []*sdk.CommandInfo{
            {
                Name:        "mycommand",
                Description: "Description of my command",
                Category:    sdk.CategoryDocker, // Use SDK constants
                Aliases:     []string{"mc"},
            },
        },
    }, nil
}
```

### Defining Custom Categories

Plugins can define their own categories with custom display properties:

```go
func (p *MyPlugin) GetCustomCategories(ctx context.Context, _ *sdk.Empty) (*sdk.CategoryList, error) {
    return &sdk.CategoryList{
        Categories: []*sdk.CustomCategory{
            {
                Id:          "infrastructure",
                Name:        "Infrastructure Management",
                Description: "AWS, Terraform, and cloud resources",
                Priority:    110, // 100-199 recommended for plugins
            },
            {
                Id:          "monitoring",
                Name:        "Monitoring & Observability",
                Description: "Metrics, logs, and alerting",
                Priority:    120,
            },
        },
    }, nil
}

// Then use your custom category in commands:
func (p *MyPlugin) ListCommands(ctx context.Context, _ *sdk.Empty) (*sdk.CommandList, error) {
    return &sdk.CommandList{
        Commands: []*sdk.CommandInfo{
            {
                Name:        "terraform-plan",
                Description: "Run terraform plan",
                Category:    "infrastructure", // Use custom category ID
            },
        },
    }, nil
}
```

#### Priority Ranges
- **0-99**: Reserved for core system categories
- **100-199**: Recommended for plugin-defined categories
- **200+**: Low priority/miscellaneous categories

## Command Registration

### Global vs Namespaced Commands

Plugins can register commands either globally or under a namespace:

```go
func (p *MyPlugin) GetMetadata(ctx context.Context, _ *sdk.Empty) (*sdk.PluginMetadata, error) {
    return &sdk.PluginMetadata{
        Name:        "myplugin",
        Namespaced:  false, // false = global, true = namespaced
    }, nil
}
```

- **Global** (`Namespaced: false`): Commands are registered at root level
  - Command: `glidemycommand`
  - Use for: Core functionality, common operations

- **Namespaced** (`Namespaced: true`): Commands are prefixed with plugin name
  - Command: `glidemyplugin:mycommand`
  - Use for: Plugin-specific utilities, specialized tools

## Command Implementation

### Basic Command
```go
func (p *MyPlugin) ExecuteCommand(ctx context.Context, req *sdk.ExecuteRequest) (*sdk.ExecuteResponse, error) {
    switch req.Command {
    case "mycommand":
        // Command implementation
        return &sdk.ExecuteResponse{
            Success: true,
            Stdout:  []byte("Command output\n"),
        }, nil
    default:
        return &sdk.ExecuteResponse{
            Success: false,
            Error:   fmt.Sprintf("unknown command: %s", req.Command),
        }, nil
    }
}
```

### Interactive Commands
For commands that need terminal interaction:

```go
func (p *MyPlugin) StartInteractive(stream sdk.GlidePlugin_StartInteractiveServer) error {
    // Handle interactive session
    // Use stream.Send() and stream.Recv() for bidirectional communication
    return nil
}
```

## Plugin Installation

### Plugin Discovery Locations

Glide searches for plugins in the following locations (in order of precedence):

1. **User plugins**: `~/.glide/plugins/`
   - Global plugins available across all projects
   
2. **Ancestor directories**: `.glide/plugins/` in parent directories
   - Walks up from current directory to home or root
   - Most specific (deepest) directories take precedence
   - Enables project-wide plugins shared across subdirectories
   
3. **Current directory**: `./.glide/plugins/`
   - Directory-specific plugins
   
4. **System plugins**: `/usr/local/lib/glide/plugins/`
   - System-wide plugins (if directory exists)

### Ancestor Directory Discovery

When working in a subdirectory of a project, Glide automatically discovers plugins from parent directories:

```
project/
├── .glide/plugins/          # Available from anywhere in project/
│   └── glide-plugin-project
├── frontend/
│   ├── .glide/plugins/      # Available from frontend/ and subdirs
│   │   └── glide-plugin-ui
│   └── components/          # Can use both project and ui plugins
└── backend/
    └── api/                 # Can use project plugin
```

In the above structure:
- Working from `frontend/components/` finds plugins from both `frontend/.glide/plugins/` and `project/.glide/plugins/`
- Working from `backend/api/` finds plugins from `project/.glide/plugins/`
- More specific plugins (deeper in tree) override less specific ones with the same name

### Naming Convention
Plugin binaries should follow the naming pattern: `glide-plugin-{name}`

Example: `glide-plugin-docker`, `glide-plugin-aws`

## Best Practices

### 1. Category Selection
- Choose the most appropriate category for your commands
- Don't use `core` category unless truly essential
- Consider creating focused plugins for specific domains

### 2. Command Naming
- Use clear, descriptive command names
- Provide meaningful aliases for common operations
- Follow Unix naming conventions (lowercase, hyphens)

### 3. Error Handling
```go
return &sdk.ExecuteResponse{
    Success:  false,
    ExitCode: 1,
    Error:    "Descriptive error message",
    Stderr:   []byte("Detailed error output\n"),
}, nil
```

### 4. Output Format
- Use stdout for normal output
- Use stderr for errors and warnings
- Respect quiet mode flags when possible

### 5. Configuration
```go
func (p *MyPlugin) Configure(ctx context.Context, req *sdk.ConfigureRequest) (*sdk.ConfigureResponse, error) {
    // Handle plugin configuration
    // Config comes from ~/.glide.yml plugins section
    return &sdk.ConfigureResponse{
        Success: true,
    }, nil
}
```

## Testing Your Plugin

1. Build your plugin:
```bash
go build -o ~/.glide/plugins/glide-plugin-myname
```

2. Verify it's loaded:
```bash
glideplugins list
```

3. Test your commands:
```bash
glidemycommand
glidehelp  # Should show your command in the appropriate category
```

## SDK Reference

The Plugin SDK is located at: `github.com/ivannovak/glide/pkg/plugin/sdk/v1`

Key interfaces:
- `GlidePluginServer` - Main plugin interface
- `ExecuteRequest` - Command execution request
- `ExecuteResponse` - Command execution response
- `CommandInfo` - Command metadata
- `PluginMetadata` - Plugin information

## Example Plugins

### Official Docker Plugin

**Repository**: https://github.com/ivannovak/glide-plugin-docker

A complete, production-ready plugin demonstrating:
- Comprehensive Docker and Docker Compose integration
- 20+ commands across multiple categories
- Context-aware configuration and auto-detection
- Multi-service orchestration
- Health monitoring and resource cleanup
- Best practices for plugin architecture

This is the recommended reference for building full-featured plugins.

### Other Examples

See `glide-plugin-chirocat` for examples of:
- Multiple command categories
- Both standard and interactive commands
- Configuration handling
- Error management

## Version Compatibility

Always specify the minimum SDK version your plugin requires:
```go
MinSdk: "v1.0.0"
```

This ensures compatibility with the Glide CLI version.