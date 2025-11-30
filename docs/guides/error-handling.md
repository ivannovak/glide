# Error Handling Guide

This guide covers best practices for error handling in Glide and plugins.

## Table of Contents
- [Overview](#overview)
- [Error Types](#error-types)
- [Creating Errors](#creating-errors)
- [Error Context](#error-context)
- [User-Friendly Suggestions](#user-friendly-suggestions)
- [Error Handling in Plugins](#error-handling-in-plugins)
- [Best Practices](#best-practices)

## Overview

Glide uses structured error handling through the `pkg/errors` package. This provides:
- **Typed errors** with exit codes
- **Contextual information** for debugging
- **User-friendly suggestions** for resolution
- **Consistent formatting** across all commands

## Error Types

### Built-in Error Types

```go
import "github.com/ivannovak/glide/v3/pkg/errors"

// Available error types
errors.TypeConfig      // Configuration errors
errors.TypeDocker      // Docker daemon errors
errors.TypeContainer   // Container-specific errors
errors.TypePermission  // Permission errors
errors.TypeFileNotFound // Missing file errors
errors.TypeDependency  // Missing dependency errors
errors.TypePlugin      // Plugin-related errors
errors.TypeValidation  // Input validation errors
errors.TypeNetwork     // Network-related errors
errors.TypeTimeout     // Timeout errors
```

### Exit Codes

Each error type has an associated exit code:

| Type | Exit Code | Description |
|------|-----------|-------------|
| General | 1 | Default error code |
| Docker | 125 | Docker daemon errors |
| Permission | 126 | Permission denied |
| FileNotFound | 127 | File or dependency not found |
| Validation | 2 | Invalid input |

## Creating Errors

### Simple Errors

```go
import "github.com/ivannovak/glide/v3/pkg/errors"

// Basic error
err := errors.New(errors.TypeConfig, "invalid configuration value")

// With typed constructors
err := errors.NewConfigError("invalid port number")
err := errors.NewDockerError("Docker daemon not running")
err := errors.NewFileNotFoundError("/path/to/config.yml")
err := errors.NewPermissionError("/etc/glide", "cannot read directory")
err := errors.NewValidationError("username", "must be at least 3 characters")
```

### Errors with Options

```go
err := errors.New(errors.TypeValidation, "invalid input",
    errors.WithExitCode(2),
    errors.WithContext("field", "username"),
    errors.WithContext("value", username),
    errors.WithCause(originalErr),
    errors.WithSuggestions(
        "Username must be 3-20 characters",
        "Only alphanumeric characters allowed",
    ),
)
```

## Error Context

### Adding Context

Context helps users and developers understand what went wrong:

```go
err := errors.NewPluginError("failed to load plugin",
    errors.WithContext("plugin", "docker-compose"),
    errors.WithContext("path", "/home/user/.glide/plugins/docker-compose"),
    errors.WithContext("reason", "binary not executable"),
)
```

### Wrapping Errors

Preserve the original error while adding context:

```go
file, err := os.Open(configPath)
if err != nil {
    return errors.New(errors.TypeFileNotFound, "failed to open config",
        errors.WithCause(err),
        errors.WithContext("path", configPath),
    )
}
```

## User-Friendly Suggestions

### Adding Suggestions

Help users resolve issues:

```go
err := errors.NewDockerError("Docker daemon is not running",
    errors.WithSuggestions(
        "Start Docker Desktop",
        "Run: sudo systemctl start docker",
        "Check Docker status: docker ps",
    ),
)
```

### Dynamic Suggestions

Generate suggestions based on context:

```go
func createFileNotFoundError(path string) *errors.GlideError {
    suggestions := []string{
        fmt.Sprintf("Check if the file exists: ls -la %s", filepath.Dir(path)),
        "Verify the path is correct",
    }

    // Add context-specific suggestions
    if strings.HasSuffix(path, ".yml") {
        suggestions = append(suggestions, "Create a new config: glide init")
    }

    return errors.NewFileNotFoundError(path,
        errors.WithSuggestions(suggestions...),
    )
}
```

## Error Handling in Plugins

### Plugin Error Pattern

```go
func (p *MyPlugin) Execute(ctx context.Context, args []string) error {
    // Validate input
    if len(args) == 0 {
        return errors.NewValidationError("args", "at least one argument required",
            errors.WithSuggestions(
                "Usage: glide my-command <arg>",
                "See: glide my-command --help",
            ),
        )
    }

    // Handle external errors
    result, err := p.callExternalService(args[0])
    if err != nil {
        return errors.NewPluginError("service call failed",
            errors.WithCause(err),
            errors.WithContext("plugin", p.Metadata().Name),
            errors.WithContext("argument", args[0]),
        )
    }

    return nil
}
```

### Error Propagation

Let errors bubble up with added context:

```go
func (p *MyPlugin) loadConfig() error {
    data, err := os.ReadFile(p.configPath)
    if err != nil {
        if os.IsNotExist(err) {
            return errors.NewFileNotFoundError(p.configPath,
                errors.WithContext("plugin", p.Metadata().Name),
            )
        }
        if os.IsPermission(err) {
            return errors.NewPermissionError(p.configPath, "cannot read config",
                errors.WithContext("plugin", p.Metadata().Name),
            )
        }
        return errors.New(errors.TypeConfig, "failed to read config",
            errors.WithCause(err),
        )
    }
    return nil
}
```

## Best Practices

### 1. Be Specific

```go
// Bad: vague error
return errors.New(errors.TypeConfig, "error loading config")

// Good: specific error with context
return errors.New(errors.TypeConfig,
    fmt.Sprintf("invalid port %d: must be between 1 and 65535", port),
    errors.WithContext("field", "port"),
    errors.WithContext("value", port),
)
```

### 2. Include Actionable Suggestions

```go
// Bad: no guidance
return errors.NewPermissionError(path, "access denied")

// Good: actionable suggestions
return errors.NewPermissionError(path, "access denied",
    errors.WithSuggestions(
        fmt.Sprintf("Check permissions: ls -la %s", path),
        fmt.Sprintf("Fix permissions: chmod 755 %s", path),
        "Run with elevated privileges if necessary",
    ),
)
```

### 3. Preserve Error Chain

```go
// Bad: losing original error
if err != nil {
    return errors.New(errors.TypeNetwork, "connection failed")
}

// Good: preserving the chain
if err != nil {
    return errors.New(errors.TypeNetwork, "connection failed",
        errors.WithCause(err),
        errors.WithContext("host", host),
        errors.WithContext("port", port),
    )
}
```

### 4. Use Appropriate Error Types

```go
// Match error type to the actual problem
switch {
case os.IsNotExist(err):
    return errors.NewFileNotFoundError(path)
case os.IsPermission(err):
    return errors.NewPermissionError(path, "cannot access")
case isTimeout(err):
    return errors.New(errors.TypeTimeout, "operation timed out",
        errors.WithCause(err),
    )
default:
    return errors.New(errors.TypePlugin, "unexpected error",
        errors.WithCause(err),
    )
}
```

### 5. Don't Swallow Errors

```go
// Bad: swallowing errors
result, _ := doSomething()

// Bad: logging but not returning
if err != nil {
    log.Error("operation failed", "error", err)
    // Missing: return err
}

// Good: handle or return
if err != nil {
    return errors.New(errors.TypePlugin, "operation failed",
        errors.WithCause(err),
    )
}
```

### 6. Error Messages

Follow these guidelines for error messages:
- Start with lowercase (they may be wrapped)
- Be concise but descriptive
- Avoid technical jargon in user-facing messages
- Include relevant values

```go
// Good error messages
"invalid configuration: port must be between 1 and 65535"
"Docker daemon not responding after 30 seconds"
"plugin 'my-plugin' failed to initialize: missing API key"
```

## Error Display

The error handler formats errors consistently:

```go
handler := errors.NewHandler(outputManager)
exitCode := handler.Handle(err)
os.Exit(exitCode)
```

Example output:
```
Error: Docker daemon not responding

Details:
  Type: docker
  Host: localhost
  Timeout: 30s

Suggestions:
  • Start Docker Desktop
  • Run: sudo systemctl start docker
  • Check status: docker info
```

## Testing Errors

### Testing Error Types

```go
func TestValidation(t *testing.T) {
    err := validate(invalidInput)

    assert.Error(t, err)

    var glideErr *errors.GlideError
    require.True(t, errors.As(err, &glideErr))
    assert.Equal(t, errors.TypeValidation, glideErr.Type)
    assert.Equal(t, 2, glideErr.Code)
}
```

### Testing Suggestions

```go
func TestErrorSuggestions(t *testing.T) {
    err := errors.NewDockerError("daemon not running",
        errors.WithSuggestions("Start Docker", "Check status"),
    )

    assert.Contains(t, err.Suggestions, "Start Docker")
    assert.Contains(t, err.Suggestions, "Check status")
}
```

## See Also

- [ADR-004: Error Handling Approach](../adr/ADR-004-error-handling-approach.md)
- [pkg/errors Documentation](../../pkg/errors/doc.go)
- [Development Guidelines](../development/ERROR_HANDLING.md)
