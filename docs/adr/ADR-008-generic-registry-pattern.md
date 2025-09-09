# ADR-008: Generic Registry Pattern

**Status**: Accepted  
**Date**: 2025-09-09  
**Decision Makers**: Architecture Team  

## Context

The Glide CLI system had four nearly identical registry implementations:
- `pkg/plugin/registry.go` - Plugin registry
- `internal/cli/registry.go` - Command registry  
- `pkg/output/registry.go` - Formatter registry
- `pkg/interfaces/registry.go` - Generic interface

This duplication (~400 lines) created maintenance burden, increased bug surface area, and violated DRY principles.

## Decision

We will implement a single generic registry using Go 1.18+ generics that can be specialized for different types while maintaining type safety.

## Implementation

### Generic Registry Design

```go
// pkg/registry/registry.go
type Registry[T any] struct {
    mu      sync.RWMutex
    items   map[string]T
    aliases map[string]string
}

// Specialized registries
type PluginRegistry = Registry[Plugin]
type CommandRegistry struct {
    *Registry[Factory]
    metadata map[string]Metadata  // Additional fields via composition
}
```

### Key Features

1. **Thread Safety**: All operations protected by RWMutex
2. **Alias Support**: Built-in alias resolution and management
3. **Consistent Ordering**: Sorted output for deterministic iteration
4. **Type Safety**: Compile-time type checking via generics
5. **Extensibility**: Composition allows specialized registries to add fields

## Consequences

### Positive
- **Eliminated 400+ lines** of duplicate code
- **Single source of truth** for registry logic
- **Reduced maintenance** - bugs fixed once, features added once
- **Type-safe** specialization without runtime assertions
- **Consistent behavior** across all registry uses

### Negative
- **Requires Go 1.18+** for generics support
- **Slight learning curve** for developers unfamiliar with generics
- **Potential performance overhead** from interface boxing (negligible in practice)

### Neutral
- Migration required careful testing to ensure backward compatibility
- Some specialized registries needed adaptation to work with composition

## Validation

- All existing tests pass
- Race detector finds no issues
- 100% backward compatibility maintained
- Performance benchmarks show no regression

## References

- [ARCHITECTURAL_REVIEW_REPORT.md](../../ARCHITECTURAL_REVIEW_REPORT.md)
- [REGISTRY_CONSOLIDATION_SUMMARY.md](../../REGISTRY_CONSOLIDATION_SUMMARY.md)
- [Generic Registry Implementation](../../pkg/registry/registry.go)