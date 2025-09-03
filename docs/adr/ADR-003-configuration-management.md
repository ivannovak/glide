# ADR-003: Configuration Management Strategy

## Status
Accepted

## Date
2025-09-03

## Context
Glide needs flexible configuration management that supports:
- User preferences (global)
- Project-specific settings
- Team-shared configurations
- Environment-specific overrides
- Plugin configurations

Configuration should be:
- Easy to understand
- Version control friendly
- Shareable across teams
- Secure (no secrets)

## Decision
We will implement a hierarchical configuration system with the following precedence (highest to lowest):

1. Command-line flags
2. Environment variables (`GLIDE_*`)
3. Project configuration (`.glide.yml`)
4. Global configuration (`~/.glide.yml`)
5. Default values (compiled-in)

Configuration format: YAML for human readability and comment support.

Structure:
```yaml
mode: multi
projects:
  projectname:
    path: /path/to/project
    docker: true
plugins:
  pluginname:
    setting: value
```

## Consequences

### Positive
- Clear precedence rules
- Team configuration sharing
- Environment-specific overrides
- Human-readable format
- Comment support
- Git-friendly

### Negative
- YAML parsing overhead
- Potential formatting issues
- Manual merging complexity
- No type safety in files

## Implementation
Configuration management in `internal/config/`:
- `Config` struct for configuration data
- `Loader` for loading from multiple sources
- `Manager` for runtime management
- Validation for configuration integrity

## Alternatives Considered
1. **JSON**: Rejected due to no comments
2. **TOML**: Rejected due to less familiarity
3. **HCL**: Rejected due to complexity
4. **Environment only**: Rejected as too limiting