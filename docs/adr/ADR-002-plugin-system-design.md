# ADR-002: Plugin System Design

## Status
Accepted

## Date
2025-09-03

## Context
Glide needs to be extensible to support organization-specific workflows and tools. Different teams have different requirements:
- Custom commands for proprietary tools
- Organization-specific workflows
- Integration with internal services
- Specialized development environments

We need a plugin system that is:
- Secure (no direct system access)
- Stable (plugin crashes don't affect host)
- Cross-language compatible
- Easy to develop and distribute

## Decision
We will implement a plugin system using Hashicorp's go-plugin library with gRPC for communication:

1. **Process Isolation**: Each plugin runs in a separate process
2. **gRPC Communication**: Language-agnostic protocol
3. **Protocol Buffers**: Strongly-typed interfaces
4. **Capability Model**: Plugins declare required permissions
5. **Hot Reload**: Plugins can be added without restart

Architecture:
```
Glide (Host) <--gRPC--> Plugin (Process)
```

## Consequences

### Positive
- Complete process isolation for security
- Plugin crashes don't affect Glide
- Cross-language plugin support
- Clear API boundaries
- Version compatibility management
- Hot-reload capability

### Negative
- IPC overhead (~10ms per call)
- Larger binary size (gRPC + protobuf)
- More complex debugging
- Plugin distribution needed
- Memory overhead per plugin

## Implementation
Plugin system implemented in `pkg/plugin/` with:
- SDK in `pkg/plugin/sdk/v1/`
- Protocol definitions in `.proto` files
- Plugin manager for lifecycle
- Registry for command integration
- Test harness in `plugintest/`

## Alternatives Considered
1. **Lua Scripting**: Rejected due to limited ecosystem
2. **Shared Libraries**: Rejected due to stability concerns
3. **REST API**: Rejected due to complexity and overhead
4. **WebAssembly**: Rejected due to immaturity