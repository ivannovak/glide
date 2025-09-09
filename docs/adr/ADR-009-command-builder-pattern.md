# ADR-009: Shell Command Builder Pattern

**Status**: Accepted  
**Date**: 2025-09-09  
**Decision Makers**: Architecture Team  

## Context

Shell execution strategies contained significant duplication:
- Each of 4 strategies had ~30-50 lines of identical command setup code
- Total duplication: ~200 lines across strategies
- Mixed concerns between command building and execution logic
- Difficult to maintain consistency across strategies

## Decision

Extract command building logic into a centralized `CommandBuilder` that handles all common setup, allowing strategies to focus solely on their unique execution patterns.

## Implementation

### Builder Pattern Architecture

```go
// internal/shell/builder.go
type CommandBuilder struct {
    cmd *Command
    ctx context.Context
}

// Fluent API for configuration
func (b *CommandBuilder) WithContext(ctx context.Context) *CommandBuilder
func (b *CommandBuilder) Build() *exec.Cmd
func (b *CommandBuilder) BuildWithCapture() (*exec.Cmd, *bytes.Buffer, *bytes.Buffer)
func (b *CommandBuilder) BuildWithStreaming(out, err io.Writer) *exec.Cmd
```

### Memory Safety Innovation

```go
// Prevents memory exhaustion from unbounded output
type LimitedBuffer struct {
    buffer bytes.Buffer  // Non-embedded to prevent type assertion bypass
    limit  int
    closed bool
}
```

### Strategy Simplification

**Before**: 64 lines
```go
func (s *BasicStrategy) Execute(ctx context.Context, cmd *Command) (*Result, error) {
    // 30+ lines of command setup...
    // Actual execution logic
}
```

**After**: 5 lines
```go
func (s *BasicStrategy) Execute(ctx context.Context, cmd *Command) (*Result, error) {
    builder := NewCommandBuilder(cmd).WithContext(ctx)
    execCmd, stdout, stderr := builder.BuildWithMixedOutput()
    result := builder.ExecuteAndCollectResult(execCmd, stdout, stderr)
    return result, nil
}
```

## Consequences

### Positive
- **67% code reduction** in strategies (300→100 lines)
- **Single source of truth** for command configuration
- **Memory protection** via 10MB buffer limits
- **Improved testability** - builder can be tested independently
- **Clear separation of concerns** - building vs execution

### Negative
- **Additional abstraction layer** - slight cognitive overhead
- **Memory overhead** from defensive copying (minimal)

### Neutral
- Requires understanding of builder pattern
- All strategies now depend on CommandBuilder

## Security Improvements

1. **Thread Safety**: Defensive copying prevents race conditions
2. **Memory Bounds**: LimitedBuffer prevents OOM attacks
3. **Context Respect**: Proper context cancellation handling

## Validation

- Comprehensive test suite (1000+ lines)
- Race detector passes
- Memory usage bounded in stress tests
- All existing strategy tests pass

## References

- [SHELL_BUILDER_EXTRACTION_SUMMARY.md](../../SHELL_BUILDER_EXTRACTION_SUMMARY.md)
- [SHELL_BUILDER_FINAL_VALIDATION.md](../../SHELL_BUILDER_FINAL_VALIDATION.md)
- [Command Builder Implementation](../../internal/shell/builder.go)