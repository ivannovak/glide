# Contract Testing Framework

This package provides a comprehensive framework for testing interface contracts in the Glide codebase.

## Table of Contents

- [What are Contract Tests?](#what-are-contract-tests)
- [Why Contract Testing?](#why-contract-testing)
- [Quick Start](#quick-start)
- [Framework Components](#framework-components)
- [Usage Guide](#usage-guide)
- [Best Practices](#best-practices)
- [Examples](#examples)
- [FAQ](#faq)

---

## What are Contract Tests?

**Contract tests** verify that implementations correctly implement their interfaces and adhere to expected behavioral contracts. Unlike unit tests that test specific implementations, contract tests ensure all implementations of an interface behave consistently and correctly.

### Contract Test vs Unit Test

| Aspect | Unit Test | Contract Test |
|--------|-----------|---------------|
| **Focus** | Specific implementation details | Interface compliance and behavior |
| **Scope** | Single implementation | All implementations of an interface |
| **Goal** | Verify implementation works | Verify contract adherence |
| **Reuse** | Implementation-specific | Shared across implementations |

### Example

```go
// Unit Test (implementation-specific)
func TestMyPlugin_Execute(t *testing.T) {
    plugin := &MyPlugin{}
    err := plugin.Execute("cmd", []string{})
    assert.NoError(t, err)
}

// Contract Test (interface-level)
func TestPlugin_Contract(t *testing.T) {
    contract := InterfaceContract{
        Name: "Plugin.Execute always returns error for invalid commands",
        Tests: []ContractTest{
            // Tests that apply to ALL Plugin implementations
        },
    }
    VerifyInterfaceContract(t, contract)
}
```

---

## Why Contract Testing?

### Benefits

1. **Interface Compliance**: Ensures all implementations satisfy the interface contract
2. **Behavioral Consistency**: All implementations behave the same way for the same inputs
3. **Better Documentation**: Contracts serve as executable specifications
4. **Regression Prevention**: Breaking contract changes are caught immediately
5. **Confidence in Refactoring**: Safe to swap implementations knowing they follow the same contract

### When to Use Contract Tests

✅ **Use contract tests when:**
- An interface has multiple implementations
- The interface is a plugin point (Plugin SDK)
- Behavioral consistency is critical (config loaders, output formatters)
- You want to test error handling across implementations

❌ **Don't use contract tests when:**
- There's only one implementation and no plan for more
- Implementation details are more important than interface compliance
- The interface is internal and unlikely to have multiple implementations

---

## Quick Start

### 1. Define Your Interface Contract

```go
import "github.com/ivannovak/glide/v3/tests/contracts"

func TestPluginContract(t *testing.T) {
    contract := contracts.InterfaceContract{
        Name: "Plugin Interface",
        Setup: func(t *testing.T) (interface{}, func()) {
            // Create implementation to test
            plugin := &MyPlugin{name: "test"}

            // Cleanup function (optional)
            cleanup := func() {
                plugin.Close()
            }

            return plugin, cleanup
        },
        Tests: []contracts.ContractTest{
            {
                Name: "GetInfo returns non-nil PluginInfo",
                Test: func(t *testing.T, impl interface{}) {
                    plugin := impl.(Plugin)
                    info := plugin.GetInfo()
                    contracts.AssertNonNil(t, info, "PluginInfo")
                },
            },
        },
    }

    contracts.VerifyInterfaceContract(t, contract)
}
```

### 2. Run Your Tests

```bash
go test ./tests/contracts/... -v
```

---

## Framework Components

### 1. InterfaceContract

Tests that an implementation satisfies interface requirements.

```go
type InterfaceContract struct {
    Name  string                                      // Descriptive name
    Setup func(t *testing.T) (interface{}, func())  // Create implementation
    Tests []ContractTest                             // Individual tests
}
```

**Use for:**
- Verifying all interface methods work correctly
- Testing method interactions
- Ensuring implementations follow interface semantics

### 2. ErrorContract

Tests that methods return appropriate errors.

```go
type ErrorContract struct {
    Name   string                                      // Error scenario name
    Setup  func(t *testing.T) (interface{}, func())  // Create implementation
    Invoke func(impl interface{}) error               // Call method that should error
    Verify func(t *testing.T, err error)              // Verify error correctness
}
```

**Use for:**
- Testing error conditions
- Verifying error types
- Ensuring error messages are helpful

### 3. BehaviorContract

Tests behavioral patterns using Given-When-Then (BDD style).

```go
type BehaviorContract struct {
    Name      string                                   // Behavior description
    Setup     func(t *testing.T) (interface{}, func()) // Create implementation
    Scenarios []BehaviorScenario                       // Test scenarios
}

type BehaviorScenario struct {
    Name  string                              // Scenario name
    Given func(impl interface{})              // Preconditions
    When  func(impl interface{}) interface{}  // Action
    Then  func(t *testing.T, result interface{}) // Verification
}
```

**Use for:**
- Testing complex workflows
- Verifying state transitions
- Testing interactions between methods

### 4. ContractTestSuite

Combines all contract types for comprehensive testing.

```go
type ContractTestSuite struct {
    InterfaceName      string
    InterfaceContracts []InterfaceContract
    ErrorContracts     []ErrorContract
    BehaviorContracts  []BehaviorContract
}
```

**Use for:**
- Comprehensive interface testing
- Testing multiple implementations with the same suite
- Organizing related contracts

---

## Usage Guide

### Testing Interface Compliance

```go
func TestPluginInterfaceCompliance(t *testing.T) {
    implementations := []struct {
        name string
        impl Plugin
    }{
        {"MyPlugin", &MyPlugin{}},
        {"AnotherPlugin", &AnotherPlugin{}},
    }

    for _, tt := range implementations {
        t.Run(tt.name, func(t *testing.T) {
            contract := contracts.InterfaceContract{
                Name: "Plugin basic compliance",
                Setup: func(t *testing.T) (interface{}, func()) {
                    return tt.impl, nil
                },
                Tests: []contracts.ContractTest{
                    {
                        Name: "GetInfo returns valid PluginInfo",
                        Test: func(t *testing.T, impl interface{}) {
                            plugin := impl.(Plugin)
                            info := plugin.GetInfo()

                            contracts.AssertNonNil(t, info, "PluginInfo")
                            contracts.AssertNonNil(t, info.Name, "Name")
                            contracts.AssertNonNil(t, info.Version, "Version")
                        },
                    },
                },
            }

            contracts.VerifyInterfaceContract(t, contract)
        })
    }
}
```

### Testing Error Handling

```go
func TestPluginErrorContract(t *testing.T) {
    contract := contracts.ErrorContract{
        Name: "Execute with invalid command returns error",
        Setup: func(t *testing.T) (interface{}, func()) {
            return &MyPlugin{}, nil
        },
        Invoke: func(impl interface{}) error {
            plugin := impl.(Plugin)
            return plugin.Execute("invalid-command", []string{})
        },
        Verify: func(t *testing.T, err error) {
            contracts.AssertError(t, err, "Execute with invalid command")
            contracts.AssertErrorContains(t, err, "unknown command", "error message")
        },
    }

    contracts.VerifyErrorContract(t, contract)
}
```

### Testing Behavior Patterns

```go
func TestPluginLifecycleBehavior(t *testing.T) {
    contract := contracts.BehaviorContract{
        Name: "Plugin lifecycle",
        Setup: func(t *testing.T) (interface{}, func()) {
            return &MyPlugin{}, nil
        },
        Scenarios: []contracts.BehaviorScenario{
            {
                Name: "Plugin can be configured before use",
                Given: func(impl interface{}) {
                    // Plugin starts unconfigured
                },
                When: func(impl interface{}) interface{} {
                    plugin := impl.(Plugin)
                    config := map[string]interface{}{"key": "value"}
                    return plugin.Configure(config)
                },
                Then: func(t *testing.T, result interface{}) {
                    err := result.(error)
                    contracts.AssertNoError(t, err, "Configure")
                },
            },
            {
                Name: "Execute fails before configuration",
                Given: func(impl interface{}) {
                    // Plugin not configured
                },
                When: func(impl interface{}) interface{} {
                    plugin := impl.(Plugin)
                    return plugin.Execute("cmd", []string{})
                },
                Then: func(t *testing.T, result interface{}) {
                    err := result.(error)
                    contracts.AssertError(t, err, "Execute before Configure")
                },
            },
        },
    }

    contracts.VerifyBehaviorContract(t, contract)
}
```

### Using ContractTestSuite

```go
func TestCompletePluginContract(t *testing.T) {
    suite := contracts.ContractTestSuite{
        InterfaceName: "Plugin",

        InterfaceContracts: []contracts.InterfaceContract{
            {
                Name: "Basic plugin operations",
                Setup: func(t *testing.T) (interface{}, func()) {
                    return &MyPlugin{}, nil
                },
                Tests: []contracts.ContractTest{
                    // Interface compliance tests
                },
            },
        },

        ErrorContracts: []contracts.ErrorContract{
            {
                Name: "Error handling",
                // Error tests
            },
        },

        BehaviorContracts: []contracts.BehaviorContract{
            {
                Name: "Lifecycle behavior",
                // Behavior tests
            },
        },
    }

    suite.Run(t)
}
```

---

## Best Practices

### 1. One Contract Per Interface

Create a separate contract test file for each interface:

```
tests/contracts/
├── plugin_contract_test.go      # Plugin interface
├── output_contract_test.go      # OutputManager interface
├── config_contract_test.go      # ConfigLoader interface
└── ...
```

### 2. Test All Implementations

Use table-driven tests to verify all implementations:

```go
func TestAllPluginImplementations(t *testing.T) {
    implementations := []struct{
        name string
        factory func() Plugin
    }{
        {"BasePlugin", func() Plugin { return &BasePlugin{} }},
        {"CustomPlugin", func() Plugin { return &CustomPlugin{} }},
    }

    for _, impl := range implementations {
        t.Run(impl.name, func(t *testing.T) {
            // Run contract tests
        })
    }
}
```

### 3. Focus on Contract, Not Implementation

✅ **Good:**
```go
{
    Name: "GetInfo returns non-nil PluginInfo",
    Test: func(t *testing.T, impl interface{}) {
        info := impl.(Plugin).GetInfo()
        contracts.AssertNonNil(t, info, "PluginInfo")
    },
}
```

❌ **Bad:**
```go
{
    Name: "GetInfo uses correct internal field",
    Test: func(t *testing.T, impl interface{}) {
        // Don't test implementation details!
        plugin := impl.(*MyPlugin)
        if plugin.internalField != "expected" {
            t.Error("wrong internal field")
        }
    },
}
```

### 4. Test Error Contracts Thoroughly

Every error condition should have a contract:

```go
errorContracts := []contracts.ErrorContract{
    {Name: "Nil config returns error", /* ... */},
    {Name: "Invalid config returns error", /* ... */},
    {Name: "Missing required field returns error", /* ... */},
}
```

### 5. Use Descriptive Names

Names should describe **what** is being tested, not **how**:

✅ **Good:**
- "GetInfo returns non-nil PluginInfo"
- "Execute fails for unconfigured plugin"
- "Configure accepts valid config"

❌ **Bad:**
- "Test 1"
- "GetInfo works"
- "Check plugin"

### 6. Keep Setup Simple

Setup should create a minimal valid instance:

```go
Setup: func(t *testing.T) (interface{}, func()) {
    // Minimal setup
    plugin := &MyPlugin{
        name: "test",
        // Only required fields
    }
    return plugin, nil
}
```

### 7. Always Provide Cleanup

Even if it's a no-op, provide a cleanup function for consistency:

```go
Setup: func(t *testing.T) (interface{}, func()) {
    impl := &MyImpl{}
    cleanup := func() {
        // Cleanup resources
        impl.Close()
    }
    return impl, cleanup
}
```

---

## Examples

### Example 1: Plugin SDK Contract

```go
// tests/contracts/plugin_contract_test.go

package contracts

import (
    "testing"
    "github.com/ivannovak/glide/v3/pkg/plugin/sdk/v1"
)

func TestPluginSDKContract(t *testing.T) {
    suite := ContractTestSuite{
        InterfaceName: "Plugin",

        InterfaceContracts: []InterfaceContract{
            {
                Name: "Plugin metadata contract",
                Setup: func(t *testing.T) (interface{}, func()) {
                    plugin := v1.NewBasePlugin("test", "1.0.0")
                    return plugin, nil
                },
                Tests: []ContractTest{
                    {
                        Name: "GetMetadata returns valid metadata",
                        Test: func(t *testing.T, impl interface{}) {
                            plugin := impl.(v1.Plugin)
                            meta := plugin.GetMetadata()

                            AssertNonNil(t, meta, "metadata")
                            AssertNonNil(t, meta.Name, "metadata.Name")
                            AssertNonNil(t, meta.Version, "metadata.Version")
                        },
                    },
                },
            },
        },

        ErrorContracts: []ErrorContract{
            {
                Name: "ExecuteCommand with unknown command returns error",
                Setup: func(t *testing.T) (interface{}, func()) {
                    plugin := v1.NewBasePlugin("test", "1.0.0")
                    return plugin, nil
                },
                Invoke: func(impl interface{}) error {
                    plugin := impl.(v1.Plugin)
                    _, err := plugin.ExecuteCommand("unknown", []string{})
                    return err
                },
                Verify: func(t *testing.T, err error) {
                    AssertError(t, err, "ExecuteCommand with unknown command")
                },
            },
        },
    }

    suite.Run(t)
}
```

### Example 2: OutputManager Contract

```go
// tests/contracts/output_contract_test.go

package contracts

import (
    "testing"
    "github.com/ivannovak/glide/v3/pkg/output"
)

func TestOutputManagerContract(t *testing.T) {
    contract := InterfaceContract{
        Name: "OutputManager formatting contract",
        Setup: func(t *testing.T) (interface{}, func()) {
            mgr := output.NewManager(output.FormatPlain, false, false, nil)
            return mgr, nil
        },
        Tests: []ContractTest{
            {
                Name: "Info writes to output",
                Test: func(t *testing.T, impl interface{}) {
                    mgr := impl.(*output.Manager)
                    // Test that Info doesn't panic
                    mgr.Info("test message")
                },
            },
            {
                Name: "Error writes error message",
                Test: func(t *testing.T, impl interface{}) {
                    mgr := impl.(*output.Manager)
                    // Test that Error doesn't panic
                    mgr.Error(errors.New("test error"))
                },
            },
        },
    }

    VerifyInterfaceContract(t, contract)
}
```

---

## FAQ

### Q: When should I write contract tests vs unit tests?

**A:** Write both! Contract tests ensure interface compliance and behavioral consistency across implementations. Unit tests verify specific implementation details and edge cases. Contract tests answer "does this behave like a Plugin?", unit tests answer "does this specific logic work correctly?"

### Q: Can I use contract tests for internal interfaces?

**A:** Yes, but only if the interface has or will have multiple implementations. For single-implementation interfaces, unit tests are usually sufficient.

### Q: How do I test platform-specific implementations?

**A:** Use build tags and separate contract test files:

```go
// +build unix

package contracts

func TestUnixSpecificContract(t *testing.T) {
    // Unix-specific contract tests
}
```

### Q: Should I test generated code (protobuf, etc.)?

**A:** No. Generated code is tested by the generator. Focus on hand-written code. For protobuf, test the protocol behavior, not the generated getters/setters.

### Q: How do I test async behavior?

**A:** Use channels or sync primitives in your contract tests:

```go
{
    Name: "Async operation completes",
    Test: func(t *testing.T, impl interface{}) {
        done := make(chan bool)
        go func() {
            impl.(AsyncInterface).DoAsync()
            done <- true
        }()

        select {
        case <-done:
            // Success
        case <-time.After(5 * time.Second):
            t.Error("Async operation timed out")
        }
    },
}
```

### Q: Can I use table-driven tests with contracts?

**A:** Yes! Combine them:

```go
scenarios := []struct{
    name string
    input string
    expectError bool
}{
    {"valid input", "hello", false},
    {"empty input", "", true},
}

for _, scenario := range scenarios {
    contract := ErrorContract{
        Name: scenario.name,
        // Use scenario values in contract
    }
    VerifyErrorContract(t, contract)
}
```

---

## Integration with Glide Testing

### Relationship to Other Test Types

| Test Type | Purpose | Location | Contract Tests Help |
|-----------|---------|----------|---------------------|
| **Unit Tests** | Test implementation details | `*_test.go` next to code | Define expected behavior |
| **Contract Tests** | Test interface compliance | `tests/contracts/` | Ensure consistency |
| **Integration Tests** | Test component interaction | `tests/integration/` | Verify real implementations |
| **E2E Tests** | Test full workflows | `tests/e2e/` | Ensure interface stability |

### Coverage Strategy

Contract tests contribute to coverage but shouldn't be the only tests:

- **Contract Tests:** 20-30% of total coverage (interface compliance)
- **Unit Tests:** 50-60% of total coverage (implementation details)
- **Integration Tests:** 10-20% of total coverage (interactions)
- **E2E Tests:** 5-10% of total coverage (user workflows)

---

## Next Steps

1. **Read the coverage analysis:** `docs/testing/COVERAGE_ANALYSIS.md`
2. **Check existing contract tests:** Look in `tests/contracts/*_test.go`
3. **Write your first contract:** Start with a simple interface
4. **Run the tests:** `go test ./tests/contracts/... -v`
5. **Review the examples:** See `framework_test.go` for working examples

---

## References

- **Contract Testing Pattern:** https://martinfowler.com/bliki/ContractTest.html
- **Interface Compliance:** Go Programming Language Spec
- **BDD (Behavior-Driven Development):** Given-When-Then pattern
- **Glide Testing Strategy:** `docs/testing/COVERAGE_ANALYSIS.md`
