# ADR-004: Error Handling Approach

## Status
Accepted

## Date
2025-09-03

## Context
Glide needs consistent error handling that:
- Provides clear feedback to users
- Helps with debugging
- Maintains security (no sensitive data in errors)
- Enables error recovery
- Supports different error severities

Errors can originate from:
- User input validation
- System operations
- Docker operations
- Plugin execution
- Network requests

## Decision
We will implement a structured error handling approach:

1. **Error Wrapping**: Use Go 1.13+ error wrapping with context
2. **Error Types**: Distinguish between user and system errors
3. **Error Context**: Include relevant context without sensitive data
4. **Suggestions**: Provide actionable suggestions for user errors
5. **Graceful Degradation**: Fall back when possible

Error Structure:
```go
// User-facing errors
type UserError struct {
    Message    string
    Suggestion string
    Details    map[string]string
}

// System errors (logged, generic message shown)
type SystemError struct {
    Cause   error
    Context string
}
```

Error Propagation:
```go
if err != nil {
    return fmt.Errorf("failed to execute command %s: %w", cmd, err)
}
```

## Consequences

### Positive
- Consistent error messages
- Better debugging with context
- Clear user guidance
- Security by default
- Error recovery possible

### Negative
- More verbose error handling
- Potential error message duplication
- Translation complexity
- Testing overhead

## Implementation
Error handling patterns throughout codebase:
- Wrap errors with context
- User vs system error distinction
- Suggestions in error messages
- Graceful degradation where possible
- No sensitive data in errors

## Best Practices
1. Always wrap errors with context
2. Provide suggestions for user errors
3. Log system errors, show generic message
4. Never expose sensitive data
5. Test error conditions