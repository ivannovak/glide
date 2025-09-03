# ADR-001: Context-Aware Architecture

## Status
Accepted

## Date
2025-09-03

## Context
Glide needs to support multiple development workflows and project structures. Developers work in different contexts:
- Single repository development
- Multi-worktree development  
- Different locations within a project
- With or without Docker

The CLI should adapt its behavior based on the current context without requiring manual configuration or mode switching.

## Decision
We will implement a context-aware architecture where Glide automatically detects and adapts to:

1. **Project Structure**: Single-repo vs multi-worktree
2. **Current Location**: Root, main repo, worktree, or outside project
3. **Environment State**: Docker availability, running services
4. **Configuration**: Global, project, and environment settings

The context detection happens automatically on every command execution, with caching for performance.

## Consequences

### Positive
- Zero configuration for users
- Seamless workflow transitions
- Intelligent command behavior
- Reduced cognitive load
- Better error messages

### Negative
- Slightly increased startup time (mitigated by caching)
- More complex command routing
- Potential edge cases in detection
- Testing complexity increases

## Implementation
Context detection is implemented in `internal/context/` with:
- Git repository detection for project root
- Directory structure analysis for mode detection
- Docker API checks for container status
- Caching layer for performance