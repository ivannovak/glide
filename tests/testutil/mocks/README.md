# Mock Implementations for Testing

This package provides mock implementations of core Glide interfaces using [testify/mock](https://github.com/stretchr/testify#mock-package). These mocks simplify unit testing by allowing you to control the behavior of dependencies and verify interactions.

## Table of Contents

- [Available Mocks](#available-mocks)
- [Quick Start](#quick-start)
- [Mock Implementations](#mock-implementations)
  - [ShellExecutor](#shellexecutor)
  - [PluginRegistry](#pluginregistry)
  - [OutputManager](#outputmanager)
  - [ContextDetector & ProjectContext](#contextdetector--projectcontext)
  - [ConfigLoader](#configloader)
- [Helper Functions](#helper-functions)
- [Best Practices](#best-practices)
- [Common Patterns](#common-patterns)
- [Testing the Mocks](#testing-the-mocks)

## Available Mocks

This package provides mocks for the following interfaces:

| Mock | Interface | Purpose |
|------|-----------|---------|
| `MockShellExecutor` | `ShellExecutor` | Mock command execution |
| `MockShellCommand` | `ShellCommand` | Mock shell command objects |
| `MockRegistry` | `Registry` | Mock plugin registry |
| `MockOutputManager` | `OutputManager` | Mock output/logging |
| `MockContextDetector` | `ContextDetector` | Mock context detection |
| `MockProjectContext` | `ProjectContext` | Mock project context |
| `MockConfigLoader` | `ConfigLoader` | Mock configuration loading |

## Quick Start

```go
package mypackage

import (
    "testing"
    "github.com/ivannovak/glide/v2/tests/testutil/mocks"
    "github.com/stretchr/testify/assert"
)

func TestMyFunction(t *testing.T) {
    // Create a mock
    mockExec := new(mocks.MockShellExecutor)
    mockCmd := new(mocks.MockShellCommand)

    // Set up expectations
    mockExec.On("Execute", mock.Anything, mockCmd).Return(&interfaces.ShellResult{
        Stdout:   "success",
        ExitCode: 0,
    }, nil)

    // Use the mock in your test
    result, err := mockExec.Execute(context.Background(), mockCmd)

    // Verify behavior
    assert.NoError(t, err)
    assert.Equal(t, "success", result.Stdout)

    // Verify expectations were met
    mockExec.AssertExpectations(t)
}
```

## Mock Implementations

### ShellExecutor

Mock for executing shell commands.

**Available Methods:**
- `Execute(ctx context.Context, cmd ShellCommand) (*ShellResult, error)`
- `ExecuteWithTimeout(cmd ShellCommand, timeout time.Duration) (*ShellResult, error)`
- `ExecuteWithProgress(cmd ShellCommand, message string) error`

**Example:**

```go
func TestShellExecution(t *testing.T) {
    mockExec := new(mocks.MockShellExecutor)
    mockCmd := new(mocks.MockShellCommand)

    // Set up success case
    expectedResult := &interfaces.ShellResult{
        Stdout:   "command output",
        Stderr:   "",
        ExitCode: 0,
        Duration: 100 * time.Millisecond,
    }

    mockExec.On("Execute", mock.Anything, mockCmd).Return(expectedResult, nil)

    // Execute
    result, err := mockExec.Execute(context.Background(), mockCmd)

    // Verify
    assert.NoError(t, err)
    assert.Equal(t, expectedResult, result)
    mockExec.AssertExpectations(t)
}
```

**Helper Functions:**
- `ExpectCommandExecution(m, cmd, result, err)` - Set up command execution expectation
- `ExpectCommandExecutionWithTimeout(m, cmd, timeout, result, err)` - Set up timed execution
- `ExpectCommandExecutionWithProgress(m, cmd, message, err)` - Set up progress execution

**Example with Helper:**

```go
func TestShellExecutionWithHelper(t *testing.T) {
    mockExec := new(mocks.MockShellExecutor)
    mockCmd := new(mocks.MockShellCommand)

    result := &interfaces.ShellResult{Stdout: "output", ExitCode: 0}

    // Use helper to set up expectation
    mocks.ExpectCommandExecution(mockExec, mockCmd, result, nil)

    // Execute and verify
    actual, err := mockExec.Execute(context.Background(), mockCmd)
    assert.NoError(t, err)
    assert.Equal(t, result, actual)
    mockExec.AssertExpectations(t)
}
```

### PluginRegistry

Mock for plugin registry operations.

**Available Methods:**
- `Register(key string, value interface{}) error`
- `Get(key string) (interface{}, bool)`
- `List() []string`
- `Remove(key string) bool`

**Example:**

```go
func TestPluginLoading(t *testing.T) {
    mockRegistry := new(mocks.MockRegistry)

    // Set up plugin loading
    plugin := &MyPlugin{}
    mockRegistry.On("Get", "my-plugin").Return(plugin, true)

    // Test
    result, found := mockRegistry.Get("my-plugin")

    // Verify
    assert.True(t, found)
    assert.Equal(t, plugin, result)
    mockRegistry.AssertExpectations(t)
}
```

**Helper Functions:**
- `ExpectPluginLoad(m, name, plugin)` - Set up successful plugin load
- `ExpectPluginNotFound(m, name)` - Set up plugin not found
- `ExpectPluginRegister(m, name, plugin, err)` - Set up plugin registration
- `ExpectPluginList(m, plugins)` - Set up plugin list

**Example with Helpers:**

```go
func TestPluginRegistryHelpers(t *testing.T) {
    mockRegistry := new(mocks.MockRegistry)

    // Use helpers
    mocks.ExpectPluginLoad(mockRegistry, "plugin1", "instance1")
    mocks.ExpectPluginNotFound(mockRegistry, "plugin2")
    mocks.ExpectPluginList(mockRegistry, []string{"plugin1", "plugin3"})

    // Verify expectations work
    p1, found1 := mockRegistry.Get("plugin1")
    assert.True(t, found1)
    assert.Equal(t, "instance1", p1)

    _, found2 := mockRegistry.Get("plugin2")
    assert.False(t, found2)

    list := mockRegistry.List()
    assert.Equal(t, []string{"plugin1", "plugin3"}, list)

    mockRegistry.AssertExpectations(t)
}
```

### OutputManager

Mock for output and logging operations.

**Available Methods:**
- `Display(data interface{}) error`
- `Info(format string, args ...interface{}) error`
- `Success(format string, args ...interface{}) error`
- `Error(format string, args ...interface{}) error`
- `Warning(format string, args ...interface{}) error`
- `Raw(text string) error`
- `Printf(format string, args ...interface{}) error`
- `Println(args ...interface{}) error`

**Example:**

```go
func TestOutput(t *testing.T) {
    mockOutput := new(mocks.MockOutputManager)

    // Set up expectations
    mockOutput.On("Info", "Processing: %s", mock.Anything).Return(nil)
    mockOutput.On("Success", "Done!", mock.Anything).Return(nil)

    // Test
    err := mockOutput.Info("Processing: %s", "item1")
    assert.NoError(t, err)

    err = mockOutput.Success("Done!")
    assert.NoError(t, err)

    mockOutput.AssertExpectations(t)
}
```

**Helper Functions:**
- `ExpectOutput(m, level, message, err)` - Set up output at specific level
- `ExpectRawOutput(m, text, err)` - Set up raw output
- `ExpectDisplayOutput(m, data, err)` - Set up display output

**Example with Helpers:**

```go
func TestOutputHelpers(t *testing.T) {
    mockOutput := new(mocks.MockOutputManager)

    // Use helpers for different output levels
    mocks.ExpectOutput(mockOutput, "info", "Starting...", nil)
    mocks.ExpectOutput(mockOutput, "success", "Complete!", nil)
    mocks.ExpectOutput(mockOutput, "error", "Failed!", errors.New("error"))
    mocks.ExpectRawOutput(mockOutput, "raw data", nil)

    // Test
    mockOutput.Info("Starting...", mock.Anything)
    mockOutput.Success("Complete!", mock.Anything)
    mockOutput.Error("Failed!", mock.Anything)
    mockOutput.Raw("raw data")

    mockOutput.AssertExpectations(t)
}
```

### ContextDetector & ProjectContext

Mocks for project context detection and representation.

**ContextDetector Methods:**
- `Detect(workingDir string) (ProjectContext, error)`
- `DetectWithRoot(workingDir, projectRoot string) (ProjectContext, error)`

**ProjectContext Methods:**
- `GetWorkingDir() string`
- `GetProjectRoot() string`
- `GetDevelopmentMode() string`
- `GetLocation() string`
- `IsDockerRunning() bool`
- `GetComposeFiles() []string`
- `IsWorktree() bool`
- `GetWorktreeName() string`

**Example:**

```go
func TestContextDetection(t *testing.T) {
    mockDetector := new(mocks.MockContextDetector)
    mockContext := new(mocks.MockProjectContext)

    // Set up mock context
    mockContext.On("GetWorkingDir").Return("/test/project")
    mockContext.On("GetProjectRoot").Return("/test")
    mockContext.On("GetDevelopmentMode").Return("docker")
    mockContext.On("IsDockerRunning").Return(true)

    // Set up detection
    mockDetector.On("Detect", "/test/project").Return(mockContext, nil)

    // Test
    ctx, err := mockDetector.Detect("/test/project")
    assert.NoError(t, err)
    assert.Equal(t, "/test/project", ctx.GetWorkingDir())
    assert.Equal(t, "docker", ctx.GetDevelopmentMode())
    assert.True(t, ctx.IsDockerRunning())

    mockDetector.AssertExpectations(t)
    mockContext.AssertExpectations(t)
}
```

**Helper Functions:**
- `ExpectContextDetection(m, workingDir, ctx, err)` - Set up context detection
- `ExpectContextDetectionWithRoot(m, workingDir, projectRoot, ctx, err)` - Set up detection with root

**Example with Helpers:**

```go
func TestContextDetectionHelpers(t *testing.T) {
    mockDetector := new(mocks.MockContextDetector)
    mockContext := new(mocks.MockProjectContext)

    // Use helper
    mocks.ExpectContextDetection(mockDetector, "/test", mockContext, nil)

    // Test
    ctx, err := mockDetector.Detect("/test")
    assert.NoError(t, err)
    assert.Equal(t, mockContext, ctx)

    mockDetector.AssertExpectations(t)
}
```

### ConfigLoader

Mock for configuration loading operations.

**Available Methods:**
- `Load(path string) error`
- `LoadDefault() error`
- `GetConfig() interface{}`
- `Save(path string) error`

**Example:**

```go
func TestConfigLoading(t *testing.T) {
    mockLoader := new(mocks.MockConfigLoader)
    config := map[string]interface{}{"key": "value"}

    // Set up expectations
    mockLoader.On("Load", "/test/config.yaml").Return(nil)
    mockLoader.On("GetConfig").Return(config)

    // Test
    err := mockLoader.Load("/test/config.yaml")
    assert.NoError(t, err)

    loadedConfig := mockLoader.GetConfig()
    assert.Equal(t, config, loadedConfig)

    mockLoader.AssertExpectations(t)
}
```

**Helper Functions:**
- `ExpectConfigLoad(m, path, err)` - Set up config load
- `ExpectConfigLoadDefault(m, err)` - Set up default config load
- `ExpectConfigGet(m, config)` - Set up config retrieval
- `ExpectConfigSave(m, path, err)` - Set up config save

**Example with Helpers:**

```go
func TestConfigLoadingHelpers(t *testing.T) {
    mockLoader := new(mocks.MockConfigLoader)
    config := map[string]string{"key": "value"}

    // Use helpers
    mocks.ExpectConfigLoad(mockLoader, "/test/config.yaml", nil)
    mocks.ExpectConfigGet(mockLoader, config)
    mocks.ExpectConfigSave(mockLoader, "/test/config.yaml", nil)

    // Test workflow
    mockLoader.Load("/test/config.yaml")
    result := mockLoader.GetConfig()
    assert.Equal(t, config, result)
    mockLoader.Save("/test/config.yaml")

    mockLoader.AssertExpectations(t)
}
```

## Helper Functions

Helper functions provide a convenient way to set up common expectations. They:

1. **Simplify test setup** - Less boilerplate code
2. **Improve readability** - Intent is clear
3. **Reduce errors** - Consistent setup patterns
4. **Encapsulate complexity** - Hide mock internals

**When to Use Helpers:**

✅ **Use helpers when:**
- Setting up common scenarios
- You want concise, readable tests
- The expectation is straightforward

❌ **Use direct mock setup when:**
- You need fine-grained control
- You're testing edge cases
- You need to use advanced testify/mock features (e.g., `mock.MatchedBy`)

## Best Practices

### 1. Always Assert Expectations

```go
func TestExample(t *testing.T) {
    mockExec := new(mocks.MockShellExecutor)

    // Set up expectations
    mockExec.On("Execute", mock.Anything, mock.Anything).Return(result, nil)

    // Your test code...

    // ALWAYS verify expectations were met
    mockExec.AssertExpectations(t)
}
```

### 2. Use `mock.Anything` for Unimportant Arguments

```go
// When you don't care about the context
mockExec.On("Execute", mock.Anything, specificCmd).Return(result, nil)

// When you don't care about format args
mockOutput.On("Info", "Starting...", mock.Anything).Return(nil)
```

### 3. Use Specific Matchers When Needed

```go
// Match specific values
mockExec.On("Execute", context.Background(), mockCmd).Return(result, nil)

// Use MatchedBy for complex matching
mockExec.On("Execute", mock.Anything, mock.MatchedBy(func(cmd interfaces.ShellCommand) bool {
    return cmd.GetCommand() == "npm"
})).Return(result, nil)
```

### 4. Set Up Multiple Expectations for Multiple Calls

```go
func TestMultipleCalls(t *testing.T) {
    mockOutput := new(mocks.MockOutputManager)

    // Set up multiple calls
    mockOutput.On("Info", "Step 1", mock.Anything).Return(nil).Once()
    mockOutput.On("Info", "Step 2", mock.Anything).Return(nil).Once()
    mockOutput.On("Success", "Done!", mock.Anything).Return(nil).Once()

    // Your code that makes multiple calls...

    mockOutput.AssertExpectations(t)
}
```

### 5. Test Error Cases

```go
func TestErrorHandling(t *testing.T) {
    mockExec := new(mocks.MockShellExecutor)
    mockCmd := new(mocks.MockShellCommand)

    // Set up error case
    expectedError := errors.New("command failed")
    mockExec.On("Execute", mock.Anything, mockCmd).Return(nil, expectedError)

    // Test error handling
    _, err := mockExec.Execute(context.Background(), mockCmd)
    assert.Error(t, err)
    assert.Equal(t, expectedError, err)

    mockExec.AssertExpectations(t)
}
```

### 6. Use Table-Driven Tests with Mocks

```go
func TestWithTableDriven(t *testing.T) {
    tests := []struct {
        name      string
        setupMock func(*mocks.MockShellExecutor)
        wantErr   bool
    }{
        {
            name: "success case",
            setupMock: func(m *mocks.MockShellExecutor) {
                m.On("Execute", mock.Anything, mock.Anything).Return(&interfaces.ShellResult{ExitCode: 0}, nil)
            },
            wantErr: false,
        },
        {
            name: "error case",
            setupMock: func(m *mocks.MockShellExecutor) {
                m.On("Execute", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockExec := new(mocks.MockShellExecutor)
            tt.setupMock(mockExec)

            // Test code...

            mockExec.AssertExpectations(t)
        })
    }
}
```

## Common Patterns

### Pattern 1: Testing Command Execution

```go
func TestCommandExecution(t *testing.T) {
    mockExec := new(mocks.MockShellExecutor)
    mockCmd := new(mocks.MockShellCommand)

    // Set up command details
    mockCmd.On("GetCommand").Return("npm")
    mockCmd.On("GetArgs").Return([]string{"install"})

    // Set up execution result
    mocks.ExpectCommandExecution(mockExec, mockCmd, &interfaces.ShellResult{
        Stdout:   "installed packages",
        ExitCode: 0,
    }, nil)

    // Test your code...

    mockExec.AssertExpectations(t)
    mockCmd.AssertExpectations(t)
}
```

### Pattern 2: Testing with Multiple Dependencies

```go
func TestWithMultipleDependencies(t *testing.T) {
    mockExec := new(mocks.MockShellExecutor)
    mockOutput := new(mocks.MockOutputManager)
    mockContext := new(mocks.MockProjectContext)

    // Set up all mocks
    mockContext.On("GetWorkingDir").Return("/test")
    mocks.ExpectOutput(mockOutput, "info", "Starting...", nil)
    mocks.ExpectCommandExecution(mockExec, mock.Anything, &interfaces.ShellResult{ExitCode: 0}, nil)
    mocks.ExpectOutput(mockOutput, "success", "Done!", nil)

    // Test code that uses all dependencies...

    mockExec.AssertExpectations(t)
    mockOutput.AssertExpectations(t)
    mockContext.AssertExpectations(t)
}
```

### Pattern 3: Testing Context Detection Flow

```go
func TestContextDetectionFlow(t *testing.T) {
    mockDetector := new(mocks.MockContextDetector)
    mockContext := new(mocks.MockProjectContext)

    // Set up context properties
    mockContext.On("GetWorkingDir").Return("/test/project")
    mockContext.On("IsDockerRunning").Return(true)
    mockContext.On("GetComposeFiles").Return([]string{"docker-compose.yml"})

    // Set up detection
    mocks.ExpectContextDetection(mockDetector, "/test/project", mockContext, nil)

    // Test detection flow...

    mockDetector.AssertExpectations(t)
    mockContext.AssertExpectations(t)
}
```

### Pattern 4: Testing Plugin Loading with Registry

```go
func TestPluginLoadingFlow(t *testing.T) {
    mockRegistry := new(mocks.MockRegistry)

    // Set up plugin registry
    mocks.ExpectPluginList(mockRegistry, []string{"docker", "node"})
    mocks.ExpectPluginLoad(mockRegistry, "docker", &DockerPlugin{})
    mocks.ExpectPluginLoad(mockRegistry, "node", &NodePlugin{})

    // Test plugin loading logic...

    mockRegistry.AssertExpectations(t)
}
```

## Testing the Mocks

The mocks themselves are thoroughly tested to ensure reliability. Run mock tests with:

```bash
go test ./tests/testutil/mocks/... -v
```

Current test coverage: **100%** (52 tests passing)

## Examples in Real Tests

See these tests for real-world usage examples:

- `pkg/app/application_test.go` - Uses testutil fixtures
- `tests/testutil/examples_test.go` - Example usage patterns
- `tests/testutil/table_test.go` - Table-driven test examples

## Additional Resources

- [testify/mock documentation](https://pkg.go.dev/github.com/stretchr/testify/mock)
- [testutil fixtures](../fixtures.go) - Fixture factory functions
- [testutil assertions](../assertions.go) - Custom assertions
- [testutil table tests](../TABLE_TESTS.md) - Table-driven test framework

## Contributing

When adding new interfaces to Glide:

1. Add the interface to `pkg/interfaces/interfaces.go`
2. Create a mock implementation in this package
3. Add helper functions for common use cases
4. Write comprehensive tests for the mock
5. Document usage patterns in this README
6. Add examples to demonstrate common scenarios

## Questions?

If you have questions about using these mocks or need additional mock implementations, please:

1. Check the test files (`*_test.go`) for examples
2. Review the testify/mock documentation
3. Ask in the development channel
