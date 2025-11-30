# Error Handling Guidelines

**Status:** Active
**Last Updated:** 2025-11-27
**Related:** [ERROR_HANDLING_AUDIT.md](../technical-debt/ERROR_HANDLING_AUDIT.md)

## Overview

This document establishes error handling guidelines for the Glide CLI codebase. Following these patterns ensures consistent, helpful error messages and proper error propagation.

## Core Principles

### 1. Errors Should Be Returned, Not Logged and Swallowed

**❌ BAD:**
```go
if err := doWork(); err != nil {
    log.Printf("error doing work: %v", err)
    return // Silently continue
}
```

**✅ GOOD:**
```go
if err := doWork(); err != nil {
    return fmt.Errorf("failed to do work: %w", err)
}
```

### 2. Use Structured Errors with Suggestions

The `pkg/errors` package provides structured error types with helpful suggestions:

```go
if _, err := os.Open(path); err != nil {
    return errors.NewFileNotFoundError(path,
        errors.WithSuggestions(
            "Check if the file exists",
            "Verify you're in the correct directory",
        ),
    )
}
```

### 3. Wrap Errors with Context

Always wrap errors with context about what operation failed:

```go
if err := loadConfig(path); err != nil {
    return fmt.Errorf("failed to load config from %s: %w", path, err)
}
```

Or use structured wrapping:

```go
if err := loadConfig(path); err != nil {
    return errors.Wrap(err, "failed to load config",
        errors.WithContext("path", path),
    )
}
```

---

## When to Return vs Log Errors

### Return Errors When:

- Operation failed and caller needs to know
- Error affects program correctness
- User action is required
- Operation cannot continue

### Log (Don't Return) When:

- Best-effort operation (nice-to-have)
- Error doesn't affect main functionality
- Already in cleanup/defer path
- Alternative behavior is acceptable

---

## When It's Safe to Ignore Errors

Some errors are genuinely safe to ignore, but **MUST be documented** with comments.

### Pattern: Safe-to-Ignore Comment

```go
// Safe to ignore: [why it's safe and what happens if it fails]
_, _ = fmt.Fprintf(w, "output")
```

### Examples of Safe-to-Ignore Scenarios

#### 1. Cosmetic Output (Progress Bars, Spinners)

```go
// Safe to ignore: Progress bar formatting (cosmetic only, doesn't affect operation)
_, _ = fmt.Fprintf(w, "\r%s", progressBar)
```

**Rationale:** If formatting fails, the worst case is missing visual feedback. Program functionality is unaffected.

#### 2. Cleanup in Defer

```go
defer func() {
    // Safe to ignore: Container cleanup in defer, errors would be logged by Stop()
    // We're already in cleanup path, nothing useful to do with error here
    _ = container.Stop(ctx)
}()
```

**Rationale:** In defer/cleanup, we're already handling an error or exiting. Propagating another error adds no value.

#### 3. Logging Itself

```go
// Safe to ignore: slog.Handler.Handle rarely fails, and if it does, we can't log the error
// (infinite recursion). Handler implementations are expected to not error on normal use.
_ = handler.Handle(ctx, record)
```

**Rationale:** Can't log errors about logging. Handler contracts expect no errors in normal use.

#### 4. Best-Effort Operations

```go
// Safe to ignore: Plugin YAML command registration errors are logged by registry
// Conflicts are expected and handled by priority system
_ = registry.AddYAMLCommand(name, cmd)
```

**Rationale:** Operation is best-effort. Registry handles conflicts internally. CLI continues with available commands.

---

## Error Types and Constructors

### Common Error Types

Use the structured error types from `pkg/errors`:

| Error Type | Constructor | Use Case |
|------------|-------------|----------|
| User Error | `NewUserError()` | Invalid user input/config |
| System Error | `NewSystemError()` | Internal/infrastructure failures |
| Plugin Error | `NewPluginError()` | Plugin operation failures |
| Config Error | `NewConfigError()` | Configuration problems |
| Permission Error | `NewPermissionError()` | File/directory access denied |
| File Not Found | `NewFileNotFoundError()` | Missing files |
| Network Error | `NewNetworkError()` | Network/connectivity issues |
| Timeout Error | `NewTimeoutError()` | Operation exceeded time limit |

### Using Error Constructors

#### User Errors (Input Validation)

```go
if pluginName == "" {
    return errors.NewUserError(
        "plugin name is required",
        "Add a 'name' field to your plugin configuration",
    )
}
```

#### System Errors (Infrastructure)

```go
if err := initializeDatabase(); err != nil {
    return errors.NewSystemError("failed to initialize database", err)
}
```

#### Plugin Errors

```go
if err := plugin.Execute(cmd); err != nil {
    return errors.NewPluginError(plugin.Name(), "command execution failed", err)
}
```

---

## Cobra Command Error Handling

### Use RunE, Not Run

**❌ BAD:**
```go
Run: func(cmd *cobra.Command, args []string) {
    if err := doWork(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

**✅ GOOD:**
```go
RunE: func(cmd *cobra.Command, args []string) error {
    if err := doWork(); err != nil {
        return fmt.Errorf("failed to do work: %w", err)
    }
    return nil
}
```

**Rationale:**
- Cobra handles `RunE` errors properly with exit codes
- Error handler can add suggestions
- Tests can verify error behavior
- No need for manual os.Exit()

---

## Testing Error Paths

### Always Test Error Cases

```go
func TestLoadConfig_FileNotFound(t *testing.T) {
    _, err := LoadConfig("/nonexistent/path")

    // Verify error occurred
    require.Error(t, err)

    // Verify error type (if using structured errors)
    assert.True(t, errors.Is(err, errors.TypeFileNotFound))

    // Verify error message contains context
    assert.Contains(t, err.Error(), "/nonexistent/path")
}
```

## References

- [Error Handling Audit](../technical-debt/ERROR_HANDLING_AUDIT.md) - Complete audit of current error handling
- [pkg/errors Documentation](../../pkg/errors/) - Error types and constructors
- [Implementation Checklist](../specs/gold-standard-remediation/implementation-checklist.md) - Task 1.5
