# Glide Architecture

This document provides an overview of Glide's architecture, design principles, and component interactions.

## Table of Contents
- [Overview](#overview)
- [Core Principles](#core-principles)
- [System Architecture](#system-architecture)
- [Package Structure](#package-structure)
- [Component Interactions](#component-interactions)
- [Plugin System](#plugin-system)
- [Configuration System](#configuration-system)
- [Dependency Injection](#dependency-injection)

## Overview

Glide is a context-aware development CLI that provides:
- **Project Detection**: Automatic detection of project type, structure, and environment
- **Plugin System**: Extensible command system via plugins
- **Configuration Merging**: Hierarchical configuration from multiple sources
- **Multi-Worktree Support**: First-class support for git worktrees

## Core Principles

### 1. Context Awareness
Glide detects and adapts to the current project context automatically:
- Project root detection
- Development mode (single-repo vs multi-worktree)
- Docker/container environment
- Framework and tooling detection

### 2. Plugin-First Architecture
Core functionality is minimal; features are provided by plugins:
- Clean separation of concerns
- Easy extensibility
- Independent versioning and updates

### 3. Type Safety
Extensive use of Go generics for compile-time safety:
- Type-safe configuration
- Generic registry pattern
- Strongly-typed plugin interfaces

### 4. Performance
Optimized for fast startup and responsive execution:
- Lazy initialization
- On-demand plugin loading
- Parallel operations where possible

## System Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                              CLI Layer                               │
│  ┌───────────────┐  ┌────────────────┐  ┌─────────────────────────┐ │
│  │ Root Command  │──│ Plugin Commands│──│ Builtin Commands        │ │
│  └───────┬───────┘  └────────┬───────┘  └───────────┬─────────────┘ │
└──────────┼───────────────────┼──────────────────────┼───────────────┘
           │                   │                      │
           ▼                   ▼                      ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         Container Layer                              │
│  ┌─────────────────────────────────────────────────────────────────┐│
│  │                    Dependency Injection (uber-fx)                ││
│  │  ┌─────────┐ ┌──────────┐ ┌─────────┐ ┌─────────┐ ┌──────────┐ ││
│  │  │ Logger  │ │  Config  │ │ Context │ │ Output  │ │ Plugins  │ ││
│  │  └─────────┘ └──────────┘ └─────────┘ └─────────┘ └──────────┘ ││
│  └─────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────┘
           │                   │                      │
           ▼                   ▼                      ▼
┌─────────────────────────────────────────────────────────────────────┐
│                          Core Services                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────────┐  │
│  │   Context    │  │   Config     │  │      Plugin Manager      │  │
│  │  Detection   │  │   Loading    │  │  ┌────────────────────┐  │  │
│  │              │  │              │  │  │    Discovery       │  │  │
│  │ • Root       │  │ • Defaults   │  │  │    Loading         │  │  │
│  │ • Mode       │  │ • Global     │  │  │    Lifecycle       │  │  │
│  │ • Docker     │  │ • Project    │  │  │    Dependencies    │  │  │
│  └──────────────┘  │ • Merging    │  │  └────────────────────┘  │  │
                     └──────────────┘  └──────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
           │                   │                      │
           ▼                   ▼                      ▼
┌─────────────────────────────────────────────────────────────────────┐
│                           Plugins                                    │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────────┐  │
│  │   Docker     │  │    Custom    │  │        SDK v2            │  │
│  │   Plugin     │  │   Plugins    │  │  ┌────────────────────┐  │  │
│  │              │  │              │  │  │ Type-Safe Config   │  │  │
│  │ • compose    │  │ • my-plugin  │  │  │ Lifecycle Hooks    │  │  │
│  │ • exec       │  │ • ...        │  │  │ Command Builder    │  │  │
│  └──────────────┘  └──────────────┘  │  └────────────────────┘  │  │
└─────────────────────────────────────────────────────────────────────┘
```

## Package Structure

### Public Packages (`pkg/`)

| Package | Purpose |
|---------|---------|
| `pkg/container` | Dependency injection container |
| `pkg/config` | Type-safe configuration |
| `pkg/errors` | Structured error handling |
| `pkg/logging` | Structured logging |
| `pkg/output` | Output formatting |
| `pkg/plugin/sdk` | Plugin SDK |
| `pkg/registry` | Generic registry pattern |
| `pkg/validation` | Input validation |
| `pkg/observability` | Metrics and health checks |
| `pkg/performance` | Performance budgets |

### Internal Packages (`internal/`)

| Package | Purpose |
|---------|---------|
| `internal/cli` | CLI command implementation |
| `internal/config` | Configuration loading |
| `internal/context` | Project context detection |
| `internal/docker` | Docker integration |
| `internal/shell` | Shell command execution |
| `internal/plugins` | Builtin plugins |

## Component Interactions

### Startup Sequence

```
1. CLI Entry Point
   │
   ├── 2. Create Container
   │       │
   │       ├── Provide Logger
   │       ├── Provide Config Loader
   │       ├── Provide Context Detector (lazy Docker)
   │       ├── Provide Output Manager
   │       └── Provide Plugin Registry
   │
   ├── 3. Discover Plugins (lazy)
   │       │
   │       ├── Scan plugin directories
   │       └── Build plugin metadata index
   │
   ├── 4. Register Commands
   │       │
   │       ├── Root command
   │       ├── Builtin commands
   │       └── Plugin commands (on-demand)
   │
   └── 5. Execute Command
           │
           ├── Detect context
           ├── Load config
           ├── Load required plugin
           └── Execute command handler
```

### Command Execution Flow

```go
// 1. User runs: glide docker-compose up
func main() {
    root := cli.NewRootCommand()
    root.Execute()
}

// 2. Container provides dependencies
container.Run(ctx, func(
    cfg *config.Config,
    out *output.Manager,
    plugins *sdk.Manager,
) error {
    // 3. Find matching command
    cmd := plugins.GetCommand("docker-compose")

    // 4. Execute with context
    return cmd.Execute(ctx, []string{"up"})
})
```

## Plugin System

### Plugin Architecture

```
┌─────────────────────────────────────────────────────┐
│                    Plugin Host                       │
│  ┌───────────────────────────────────────────────┐  │
│  │              Plugin Manager                    │  │
│  │  • Discovery (parallel directory scan)        │  │
│  │  • Validation (security, checksums)           │  │
│  │  • Loading (on-demand, cached)                │  │
│  │  • Lifecycle (init, start, stop)              │  │
│  └─────────────────────┬─────────────────────────┘  │
└────────────────────────┼────────────────────────────┘
                         │ gRPC
                         ▼
┌─────────────────────────────────────────────────────┐
│                   Plugin Process                     │
│  ┌───────────────────────────────────────────────┐  │
│  │                   SDK v2                       │  │
│  │  • BasePlugin[Config]                         │  │
│  │  • Type-safe configuration                    │  │
│  │  • Lifecycle hooks                            │  │
│  │  • Command handlers                           │  │
│  └───────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

### Plugin Lifecycle

```
Discovered → Loading → Initializing → Ready → Running → Stopping → Stopped
     │          │           │           │         │          │
     │          │           │           │         │          └── Cleanup
     │          │           │           │         └── Normal operation
     │          │           │           └── Startup complete
     │          │           └── OnStart()
     │          └── Init(), Configure()
     └── Found in plugin directory
```

## Configuration System

### Configuration Hierarchy

```
Priority (lowest to highest):
1. Built-in Defaults
   │
2. Global Config (~/.glide/config.yml)
   │
3. Project Config (.glide.yml)
   │
4. Environment Variables (GLIDE_*)
   │
5. Command-line Flags
```

### Type-Safe Configuration

```go
// Define typed config
type MyPluginConfig struct {
    APIKey  string `json:"api_key"`
    Timeout int    `json:"timeout"`
}

// Register with defaults
config.Register("my-plugin", MyPluginConfig{
    Timeout: 30,
})

// Retrieve typed config
cfg, err := config.Get[MyPluginConfig]("my-plugin")
// cfg.APIKey, cfg.Timeout are strongly typed
```

## Dependency Injection

### Container Pattern

```go
// Container setup
container, err := container.New(
    // Custom providers
    fx.Provide(myProvider),
    // Replacements for testing
    fx.Replace(mockConfig),
)

// Run with injected dependencies
container.Run(ctx, func(
    log *logging.Logger,
    cfg *config.Config,
    out *output.Manager,
) error {
    log.Info("Starting operation")
    return nil
})
```

### Provider Chain

```
Logger (no deps)
    ↓
ConfigLoader (Logger)
    ↓
Config (ConfigLoader)
    ↓
ContextDetector (Config, Logger)
    ↓
ProjectContext (ContextDetector)
    ↓
PluginManager (Logger, Config)
    ↓
OutputManager (Logger)
```

## Design Decisions

Key architectural decisions are documented in ADRs:

| ADR | Decision |
|-----|----------|
| [ADR-001](../adr/ADR-001-context-aware-architecture.md) | Context-aware architecture |
| [ADR-002](../adr/ADR-002-plugin-system-design.md) | Plugin system design |
| [ADR-003](../adr/ADR-003-configuration-management.md) | Configuration management |
| [ADR-004](../adr/ADR-004-error-handling-approach.md) | Error handling approach |
| [ADR-008](../adr/ADR-008-generic-registry-pattern.md) | Generic registry pattern |
| [ADR-013](../adr/ADR-013-dependency-injection.md) | Dependency injection |

## See Also

- [Plugin Development Guide](../guides/plugin-development.md)
- [Error Handling Guide](../guides/error-handling.md)
- [Performance Guide](../guides/performance.md)
- [ADR Index](../adr/README.md)
