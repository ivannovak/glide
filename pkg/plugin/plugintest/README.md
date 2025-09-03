# Plugin Test Harness

The `plugintest` package provides comprehensive testing utilities for Glide plugin developers. It includes test harnesses, mock implementations, assertions, and fixtures to make plugin testing easy and consistent.

## Quick Start

```go
package myplugin_test

import (
    "testing"
    "github.com/ivannovak/glide/pkg/plugin/plugintest"
    "github.com/stretchr/testify/assert"
)

func TestMyPlugin(t *testing.T) {
    // Create test harness
    harness := plugintest.NewTestHarness(t)
    
    // Register your plugin
    plugin := NewMyPlugin()
    err := harness.RegisterPlugin(plugin)
    assert.NoError(t, err)
    
    // Test command execution
    output, err := harness.ExecuteCommand("mycommand", "--flag=value")
    assert.NoError(t, err)
    assert.Contains(t, output, "expected output")
}
```

## Components

### 1. Test Harness (`harness.go`)

The main testing environment for plugins.

```go
harness := plugintest.NewTestHarness(t)

// Register plugins
harness.RegisterPlugin(myPlugin)

// Set configuration
harness.WithConfig(map[string]interface{}{
    "myplugin": map[string]interface{}{
        "key": "value",
    },
})

// Execute commands
output, err := harness.ExecuteCommand("command", "arg1", "arg2")

// Assertions
harness.AssertPluginRegistered("myplugin")
harness.AssertCommandExists("mycommand")
```

### 2. Mock Plugin (`mock.go`)

A configurable mock implementation of the Plugin interface.

```go
// Create a simple mock
mock := plugintest.NewMockPlugin("test")

// Configure behavior
mock.WithError(errors.New("test error"))
mock.WithConfigError(errors.New("config error"))
mock.WithMetadata(customMetadata)

// Use in tests
err := registry.Register(mock)
assert.True(t, mock.Configured)
assert.True(t, mock.Registered)
```

### 3. Command Helpers (`command.go`)

Utilities for testing Cobra commands.

```go
// Command helper
helper := plugintest.NewCommandHelper()
output, err := helper.ExecuteCommand(cmd, "arg1", "arg2")

// Command builder
cmd := plugintest.NewCommandBuilder("test").
    WithShort("Test command").
    WithFlag("verbose", "v", false, "Verbose output").
    WithSubCommand(subCmd).
    Build()

// Command tree verification
exists := plugintest.CommandExists(root, "subcommand", "action")
```

### 4. Configuration Testing (`config.go`)

Tools for testing plugin configuration.

```go
// Config builder
config := plugintest.NewConfigBuilder().
    WithPlugin("myplugin", map[string]interface{}{
        "key": "value",
    }).
    WithValue("global", true).
    Build()

// Test config
testConfig := plugintest.NewTestConfig().
    SetPlugin("myplugin", pluginConfig).
    Set("debug", true)

// Validation
validator := plugintest.NewConfigValidator().
    RequireKeys("myplugin", "api_key", "endpoint")
err := validator.Validate(config)
```

### 5. Assertions (`assertions.go`)

Plugin-specific test assertions.

```go
assert := plugintest.NewAssertions(t)

// Plugin assertions
assert.AssertPluginMetadata(plugin, expectedMetadata)
assert.AssertPluginRegistered(registry, "myplugin")
assert.AssertPluginConfigured(mockPlugin)

// Command assertions
assert.AssertCommandStructure(cmd, "use", "short description")
assert.AssertCommandHasSubcommand(parent, "child")
assert.AssertCommandTree(root, []string{"cmd1"}, []string{"cmd2", "sub"})

// Output assertions
assert.AssertOutputContains(output, "expected")
assert.AssertErrorContains(err, "error message")

// Fluent assertions
plugintest.NewQuickAssert(t).
    Plugin(myPlugin).
    HasName("myplugin").
    HasVersion("1.0.0").
    HasCommands(5)
```

### 6. Fixtures (`fixtures.go`)

Test data generators and scenarios.

```go
fixtures := plugintest.NewFixtures()

// Create test plugins
simple := fixtures.SimplePlugin("test")
complex := fixtures.ComplexPlugin("advanced")
errorPlugin := fixtures.ErrorPlugin("error")

// Generate configurations
config := fixtures.SampleConfig("myplugin")
multiConfig := fixtures.MultiPluginConfig("p1", "p2", "p3")

// Use factory pattern
factory := plugintest.NewPluginFactory()
plugin := factory.Create(
    plugintest.WithName("custom"),
    plugintest.WithVersion("2.0.0"),
    plugintest.WithCommands(cmd1, cmd2),
)

// Pre-configured scenarios
plugin, config := plugintest.TestScenarios.SinglePlugin()
plugins, config := plugintest.TestScenarios.MultiplePlugin()
```

## Testing Patterns

### Basic Plugin Test

```go
func TestPlugin_Basic(t *testing.T) {
    harness := plugintest.NewTestHarness(t)
    plugin := NewMyPlugin()
    
    // Test registration
    err := harness.RegisterPlugin(plugin)
    assert.NoError(t, err)
    
    // Verify metadata
    meta := plugin.Metadata()
    assert.Equal(t, "myplugin", meta.Name)
    assert.Equal(t, "1.0.0", meta.Version)
    
    // Test command execution
    output, err := harness.ExecuteCommand("mycommand")
    assert.NoError(t, err)
    assert.Contains(t, output, "Success")
}
```

### Configuration Test

```go
func TestPlugin_Configuration(t *testing.T) {
    harness := plugintest.NewTestHarness(t)
    
    // Set up configuration
    config := plugintest.NewConfigBuilder().
        WithPlugin("myplugin", map[string]interface{}{
            "api_key": "test-key",
            "timeout": 30,
        }).
        Build()
    
    harness.WithConfig(config)
    
    // Register plugin
    plugin := NewMyPlugin()
    err := harness.RegisterPlugin(plugin)
    assert.NoError(t, err)
    
    // Verify configuration was applied
    // (Plugin should use the configuration internally)
    output, err := harness.ExecuteCommand("mycommand")
    assert.NoError(t, err)
}
```

### Command Structure Test

```go
func TestPlugin_Commands(t *testing.T) {
    harness := plugintest.NewTestHarness(t)
    plugin := NewMyPlugin()
    
    err := harness.RegisterPlugin(plugin)
    assert.NoError(t, err)
    
    // Verify command tree
    harness.AssertCommandExists("myplugin")
    harness.AssertCommandExists("myplugin", "list")
    harness.AssertCommandExists("myplugin", "get")
    
    // Test subcommand execution
    output, err := harness.ExecuteCommand("myplugin", "list")
    assert.NoError(t, err)
    assert.Contains(t, output, "Items:")
}
```

### Error Handling Test

```go
func TestPlugin_ErrorHandling(t *testing.T) {
    harness := plugintest.NewTestHarness(t)
    assertions := plugintest.NewAssertions(t)
    
    // Test with invalid configuration
    plugin := NewMyPlugin()
    harness.WithConfig(map[string]interface{}{
        "myplugin": map[string]interface{}{
            // Missing required "api_key"
        },
    })
    
    err := harness.RegisterPlugin(plugin)
    assertions.AssertErrorContains(err, "api_key is required")
}
```

### Integration Test

```go
func TestPlugin_Integration(t *testing.T) {
    // Use fixtures for consistent test data
    fixtures := plugintest.NewFixtures()
    harness := plugintest.NewTestHarness(t)
    
    // Set up multiple plugins
    plugins := fixtures.PluginSet("test", 3)
    config := fixtures.MultiPluginConfig("test-1", "test-2", "test-3")
    
    harness.WithConfig(config)
    
    // Register all plugins
    for _, p := range plugins {
        err := harness.RegisterPlugin(p)
        assert.NoError(t, err)
    }
    
    // Verify all plugins are registered
    registered := harness.ListPlugins()
    assert.Len(t, registered, 3)
}
```

## Best Practices

1. **Use the Test Harness**: Always use `TestHarness` for consistency
2. **Leverage Fixtures**: Use fixtures for common test data
3. **Test Configuration**: Always test with various configurations
4. **Verify Commands**: Test the command structure and execution
5. **Check Error Cases**: Test error conditions and edge cases
6. **Use Assertions**: Use the provided assertions for clarity
7. **Clean State**: Reset harness between test cases if needed

## Example Plugin Test Suite

```go
package myplugin_test

import (
    "testing"
    "github.com/ivannovak/glide/pkg/plugin/plugintest"
    "github.com/stretchr/testify/suite"
)

type PluginTestSuite struct {
    suite.Suite
    harness *plugintest.TestHarness
}

func (s *PluginTestSuite) SetupTest() {
    s.harness = plugintest.NewTestHarness(s.T())
}

func (s *PluginTestSuite) TearDownTest() {
    s.harness.Reset()
}

func (s *PluginTestSuite) TestRegistration() {
    plugin := NewMyPlugin()
    err := s.harness.RegisterPlugin(plugin)
    s.NoError(err)
    s.harness.AssertPluginRegistered("myplugin")
}

func (s *PluginTestSuite) TestCommands() {
    plugin := NewMyPlugin()
    s.NoError(s.harness.RegisterPlugin(plugin))
    
    output, err := s.harness.ExecuteCommand("mycommand")
    s.NoError(err)
    s.Contains(output, "Success")
}

func TestPluginSuite(t *testing.T) {
    suite.Run(t, new(PluginTestSuite))
}
```

## Advanced Usage

### Custom Mock Behavior

```go
mock := plugintest.NewMockPlugin("test")
mock.RegisterFunc = func(root *cobra.Command) error {
    // Custom registration logic
    cmd := &cobra.Command{
        Use: "custom",
        RunE: func(cmd *cobra.Command, args []string) error {
            return nil
        },
    }
    root.AddCommand(cmd)
    return nil
}
```

### Parallel Testing

```go
func TestPlugin_Parallel(t *testing.T) {
    t.Parallel()
    
    // Each parallel test gets its own harness
    harness := plugintest.NewTestHarness(t)
    // ... test code ...
}
```

### Benchmarking

```go
func BenchmarkPluginRegistration(b *testing.B) {
    harness := plugintest.NewTestHarness(&testing.T{})
    plugin := NewMyPlugin()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        harness.Reset()
        _ = harness.RegisterPlugin(plugin)
    }
}
```