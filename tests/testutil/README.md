# Test Utilities Package

The `tests/testutil` package provides reusable test helpers, fixtures, and assertions to make testing easier and more consistent across the Glide codebase.

## Overview

This package includes:
- **Fixture Factories**: Create test instances of core types with sensible defaults
- **Assertion Helpers**: Simple, chainable assertion functions
- **Test Utilities**: Helper functions for common test operations

## Fixture Factories

### Creating Test Contexts

```go
import "github.com/ivannovak/glide/v2/tests/testutil"

// Basic context with defaults
ctx := testutil.NewTestContext()

// Customize with options
ctx := testutil.NewTestContext(
    testutil.WithProjectRoot("/my/project"),
    testutil.WithProjectName("my-app"),
    testutil.WithDevelopmentMode(context.ModeMultiWorktree),
)

// Pre-configured multi-worktree scenarios
rootCtx := testutil.NewTestContext(testutil.WithMultiWorktreeRoot())
mainCtx := testutil.NewTestContext(testutil.WithMultiWorktreeMainRepo())
wtCtx := testutil.NewTestContext(testutil.WithWorktree("feature-123"))

// Docker configuration
ctx := testutil.NewTestContext(
    testutil.WithDockerRunning(true),
    testutil.WithComposeFiles("docker-compose.yml", "docker-compose.override.yml"),
)
```

#### Available Context Options

- `WithWorkingDir(dir string)` - Set working directory
- `WithProjectRoot(root string)` - Set project root
- `WithProjectName(name string)` - Set project name
- `WithDevelopmentMode(mode)` - Set development mode
- `WithLocation(location)` - Set location type
- `WithMultiWorktreeRoot()` - Configure as multi-worktree root
- `WithMultiWorktreeMainRepo()` - Configure as main repo
- `WithWorktree(name)` - Configure as worktree
- `WithDockerRunning(bool)` - Set Docker status
- `WithComposeFiles(...files)` - Set compose files
- `WithExtension(name, data)` - Add plugin extension
- `WithCommandScope(scope)` - Set command scope
- `WithContextError(err)` - Set context error

### Creating Test Configs

```go
// Basic config with defaults
cfg := testutil.NewTestConfig()

// Customize with options
cfg := testutil.NewTestConfig(
    testutil.WithDefaultProject("my-project"),
    testutil.WithProject("my-project", "/path/to/project", "single-repo"),
    testutil.WithTestDefaults(true, true), // parallel, coverage
    testutil.WithCommand("deploy", "kubectl apply -f k8s/"),
)
```

#### Available Config Options

- `WithDefaultProject(name)` - Set default project
- `WithProject(name, path, mode)` - Add a project
- `WithTestDefaults(parallel, coverage)` - Set test defaults
- `WithCommand(name, cmd)` - Add a command

### Creating Test Applications

```go
// Basic application with defaults
app := testutil.NewTestApplication()

// Customize with options
buf := testutil.NewTestWriter()
ctx := testutil.NewTestContext(testutil.WithProjectRoot("/test"))
cfg := testutil.NewTestConfig()

app := testutil.NewTestApplication(
    testutil.WithTestWriter(buf),
    testutil.WithTestContext(ctx),
    testutil.WithTestConfig(cfg),
    testutil.WithTestOutputFormat(output.FormatJSON),
)

// Capture output
app.OutputManager.Info("Hello")
output := buf.String()
```

#### Available App Options

- `WithTestWriter(w)` - Set custom writer (defaults to bytes.Buffer)
- `WithTestContext(ctx)` - Set project context
- `WithTestConfig(cfg)` - Set configuration
- `WithTestOutputFormat(format)` - Set output format
- `WithTestShellExecutor(executor)` - Set shell executor

## Assertion Helpers

Simple assertion functions that fail tests with helpful messages:

```go
func TestMyFeature(t *testing.T) {
    result, err := MyFunction()

    // Error assertions
    testutil.AssertNoError(t, err, "MyFunction failed")
    testutil.AssertError(t, err, "expected error")
    testutil.AssertErrorContains(t, err, "not found")

    // Equality assertions
    testutil.AssertEqual(t, "expected", result, "wrong result")
    testutil.AssertNotEqual(t, "bad", result, "should be different")

    // Boolean assertions
    testutil.AssertTrue(t, result.IsValid(), "should be valid")
    testutil.AssertFalse(t, result.IsEmpty(), "should not be empty")

    // Nil assertions
    testutil.AssertNil(t, result.Error, "should have no error")
    testutil.AssertNotNil(t, result.Data, "should have data")

    // String assertions
    testutil.AssertContains(t, output, "success", "should contain success")
    testutil.AssertNotContains(t, output, "error", "should not contain error")
    testutil.AssertEmpty(t, result.Warning, "should have no warning")
    testutil.AssertNotEmpty(t, result.Name, "should have name")

    // Collection assertions
    testutil.AssertLen(t, result.Items, 3, "should have 3 items")

    // Struct comparison with detailed diff
    expected := &MyStruct{Field1: "a", Field2: "b"}
    actual := &MyStruct{Field1: "a", Field2: "c"}
    testutil.AssertStructEqual(t, expected, actual) // Shows which fields differ

    // Panic assertions
    testutil.AssertPanics(t, func() {
        PanickyFunction()
    }, "should panic")

    testutil.AssertNoPanic(t, func() {
        SafeFunction()
    }, "should not panic")
}
```

### Require vs Assert

- **Assert**: Test continues after failure (use for multiple checks)
- **Require**: Test stops immediately after failure (use for prerequisites)

```go
func TestWithRequire(t *testing.T) {
    result, err := Setup()
    testutil.RequireNoError(t, err, "setup must succeed")
    testutil.RequireNotNil(t, result, "must have result")

    // Now safe to use result
    testutil.AssertEqual(t, "expected", result.Value, "check value")
}
```

## Test Utilities

### Temporary Directories

```go
func TestWithTempDir(t *testing.T) {
    dir, cleanup := testutil.TempDir(t)
    defer cleanup()

    // Use dir for test operations
    filePath := filepath.Join(dir, "test.txt")
    os.WriteFile(filePath, []byte("test"), 0644)

    // cleanup() automatically removes dir
}
```

### Output Capture

```go
func TestOutputCapture(t *testing.T) {
    buf := testutil.NewTestWriter()
    app := testutil.NewTestApplication(
        testutil.WithTestWriter(buf),
    )

    app.OutputManager.Info("Hello, World!")

    output := buf.String()
    testutil.AssertContains(t, output, "Hello, World!", "should have message")
}
```

## Best Practices

### 1. Use Sensible Defaults

The factory functions provide working defaults, customize only what you need:

```go
// Good: Only specify what matters for this test
ctx := testutil.NewTestContext(
    testutil.WithProjectName("my-app"),
)

// Avoid: Over-specifying everything
ctx := testutil.NewTestContext(
    testutil.WithWorkingDir("/test/project"),  // default is fine
    testutil.WithProjectRoot("/test/project"), // default is fine
    testutil.WithProjectName("my-app"),        // only this matters
    testutil.WithDevelopmentMode(context.ModeSingleRepo), // default is fine
)
```

### 2. Use Pre-configured Helpers

For common scenarios, use the pre-configured helpers:

```go
// Good: Clear intent
ctx := testutil.NewTestContext(testutil.WithMultiWorktreeRoot())

// Avoid: Manual configuration
ctx := testutil.NewTestContext(
    testutil.WithDevelopmentMode(context.ModeMultiWorktree),
    testutil.WithLocation(context.LocationRoot),
    // ... plus setting 3 other boolean flags
)
```

### 3. Prefer Assert Over testify Require

Use our assertion helpers for consistency:

```go
// Good: Consistent with codebase
testutil.AssertNoError(t, err, "operation failed")
testutil.AssertEqual(t, expected, actual, "wrong value")

// Avoid: Mixed styles
require.NoError(t, err)
assert.Equal(t, expected, actual)
```

### 4. Use Descriptive Messages

Assertion messages should explain what was expected:

```go
// Good: Clear context
testutil.AssertTrue(t, ctx.IsValid(), "context should be valid after detection")

// Avoid: Generic messages
testutil.AssertTrue(t, ctx.IsValid(), "check failed")
```

### 5. Structure Tests for Readability

```go
func TestFeature(t *testing.T) {
    // Arrange: Set up test fixtures
    ctx := testutil.NewTestContext(testutil.WithProjectRoot("/test"))
    app := testutil.NewTestApplication(testutil.WithTestContext(ctx))

    // Act: Perform the operation
    result, err := app.DoSomething()

    // Assert: Verify results
    testutil.AssertNoError(t, err, "operation should succeed")
    testutil.AssertEqual(t, "expected", result, "wrong result")
}
```

## Integration with Existing Tests

The testutil package works alongside `testify/assert` and `testify/require`. You can mix them as needed:

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/ivannovak/glide/v2/tests/testutil"
)

func TestMixed(t *testing.T) {
    // Use testutil for fixtures
    ctx := testutil.NewTestContext()

    // Use either assertion style
    testutil.AssertNotNil(t, ctx, "context should exist")
    assert.NotNil(t, ctx) // Also fine

    // testutil assertions have descriptive messages built-in
    testutil.AssertEqual(t, "expected", ctx.ProjectName, "wrong project name")
}
```

## Examples

### Testing Context Detection

```go
func TestContextDetection(t *testing.T) {
    tests := []struct {
        name     string
        setup    func() *context.ProjectContext
        expected context.DevelopmentMode
    }{
        {
            name: "single repo mode",
            setup: func() *context.ProjectContext {
                return testutil.NewTestContext(
                    testutil.WithDevelopmentMode(context.ModeSingleRepo),
                )
            },
            expected: context.ModeSingleRepo,
        },
        {
            name: "multi-worktree root",
            setup: func() *context.ProjectContext {
                return testutil.NewTestContext(
                    testutil.WithMultiWorktreeRoot(),
                )
            },
            expected: context.ModeMultiWorktree,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := tt.setup()
            testutil.AssertEqual(t, tt.expected, ctx.DevelopmentMode, "wrong mode")
        })
    }
}
```

### Testing with Output Capture

```go
func TestCommandOutput(t *testing.T) {
    buf := testutil.NewTestWriter()
    app := testutil.NewTestApplication(
        testutil.WithTestWriter(buf),
    )

    err := app.OutputManager.Success("Operation completed")
    testutil.AssertNoError(t, err, "output failed")

    output := buf.String()
    testutil.AssertContains(t, output, "Operation completed", "missing message")
    testutil.AssertContains(t, output, "✓", "missing checkmark")
}
```

### Testing with Temporary Files

```go
func TestConfigLoading(t *testing.T) {
    dir, cleanup := testutil.TempDir(t)
    defer cleanup()

    // Create test config file
    configPath := filepath.Join(dir, ".glide.yml")
    err := os.WriteFile(configPath, []byte("version: 2"), 0644)
    testutil.AssertNoError(t, err, "failed to create config")

    // Test loading
    ctx := testutil.NewTestContext(testutil.WithProjectRoot(dir))
    cfg, err := LoadConfig(ctx)
    testutil.AssertNoError(t, err, "failed to load config")
    testutil.AssertNotNil(t, cfg, "config should be loaded")
}
```

## Package Organization

```
tests/testutil/
├── README.md           # This file
├── fixtures.go         # Factory functions for test data
├── assertions.go       # Assertion helpers
└── examples_test.go    # Usage examples
```

## Contributing

When adding new test helpers:

1. Add factory functions to `fixtures.go`
2. Add assertions to `assertions.go`
3. Document usage in this README
4. Add examples to `examples_test.go`
5. Keep helpers simple and focused
6. Follow existing naming conventions

## FAQ

**Q: Should I use testutil assertions or testify?**
A: Use testutil for consistency, but testify is fine too. Don't mix within a single test.

**Q: When should I create a new fixture factory?**
A: When you find yourself repeatedly creating the same test setup in multiple tests.

**Q: Can I combine multiple options?**
A: Yes! Options are composable: `NewTestContext(WithProjectRoot("/test"), WithDockerRunning(true))`

**Q: What about mocks?**
A: For now, use `testify/mock` directly. A future subtask will add mock helpers to this package.

## Related Documentation

- [Testing Best Practices](../../docs/testing/best-practices.md)
- [Table-Driven Tests](./TABLE_TESTS.md) (coming soon)
- [Mock Implementations](./mocks/README.md) (coming soon)
