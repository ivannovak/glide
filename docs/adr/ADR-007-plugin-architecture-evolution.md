# ADR-007: Plugin Architecture Evolution

## Status
Accepted

## Date
2025-09-08

## Context
The Glide CLI initially mixed framework capabilities (worktree management, plugin system) with project-specific implementations (Docker commands, Laravel testing). This created a false assumption that all projects using Glide would have the same structure and needs.

Key issues identified:
1. Docker Management, Testing, and Development Tools commands depend on specific Docker setup not universally applicable
2. Core innovation of Glide (multimodal development workflow) was obscured by project-specific commands
3. Plugin commands were strictly namespaced, limiting customizability
4. No support for project-local plugins
5. Category system was rigid with hardcoded priorities

## Decision

### 1. Separate Framework from Implementation
- **Core Glide** provides only universal commands:
  - `setup` - Configure Glide for a project
  - `plugins` - Manage plugins
  - `version` - Version information
  - `help` - Help system
  - `project` - Multi-worktree management (when in that mode)
  - `self-update` - Update Glide itself

- **Project-specific commands** move to plugins:
  - Docker management commands → plugin
  - Testing commands → plugin
  - Development tools → plugin

### 2. Category Priority System
Use magnitude-based spacing (10, 20, 30...) instead of sequential (1, 2, 3...):
- Core Commands: 10
- Global Commands: 20
- Setup & Configuration: 30
- [Available space: 40-70 for project/plugin categories]
- Plugin Commands: 80
- Help & Documentation: 90

This allows plugins to insert categories without decimal notation.

### 3. Plugin Discovery Locations
Support multiple plugin discovery locations:
```
~/.glide/plugins/          # Global plugins (always loaded)
./.glide/plugins/          # Project-local plugins (context-aware)
```

### 4. Flexible Command Registration
Plugins can choose registration strategy:
```go
type PluginMetadata struct {
    Name        string
    Namespaced  bool // If false, commands register directly to root
    Categories  []CategoryDefinition // Plugin can define new categories
}
```

Examples:
- **Namespaced**: `glideacme mysql` (safe, no conflicts)
- **Global**: `glide mysql` (convenient, requires conflict resolution)

### 5. Categories as First-Class Objects
Categories become richer objects:
```go
type CategoryDefinition struct {
    Slug        string  // "core", "docker", etc.
    Name        string  // "Core Commands"
    Description string  // "Essential commands for..."
    Priority    int     // 10, 20, 30...
    Source      string  // "builtin" or "plugin:name"
}
```

## Consequences

### Positive
- **Clear separation of concerns**: Framework vs implementation
- **Project flexibility**: Each project can compose its perfect command set
- **Better discoverability**: Core Glide innovation (worktree management) is more visible
- **Extensibility**: Plugins can add categories and commands without modifying core
- **Context-aware**: Project-local plugins load only when relevant

### Negative
- **Potential conflicts**: Global command registration may cause naming conflicts
- **Complexity**: Multiple plugin locations increase discovery complexity

### Neutral
- **Learning curve**: Users need to understand plugin system for full functionality
- **Distribution**: Need mechanism for plugin installation/management

## Future Considerations

- Plugin dependency management
- Plugin versioning and compatibility
- Central plugin registry/marketplace
- Conflict resolution strategies for project commands
- Plugin configuration inheritance (global → project → local)

## Implementation

Phase 1: Update category priorities to use magnitude-based spacing (COMPLETED)
Phase 2: Implement project-local plugin discovery (COMPLETED)
Phase 3: Add namespaced vs global registration support (COMPLETED)
Phase 4: Move project-specific commands to plugins (COMPLETED)

## Alternatives Considered

1. **Keep all commands in core**: Rejected due to lack of flexibility
2. **Force all plugins to be namespaced**: Rejected as too restrictive
3. **Single plugin directory**: Rejected as not supporting project-specific needs

## References

- Original discussion: Help command enhancement and plugin architecture review
- Related: ADR-002 on initial plugin system design
