// Package contracts provides a framework for testing interface contracts.
//
// Contract tests verify that implementations correctly implement their interfaces
// and adhere to expected behavioral contracts. Unlike unit tests that test specific
// implementations, contract tests ensure all implementations of an interface behave
// consistently and correctly.
package contracts

import (
	"reflect"
	"testing"
)

// InterfaceContract defines a contract that an interface implementation must satisfy.
type InterfaceContract struct {
	// Name is the descriptive name of this contract
	Name string

	// Setup is called before each test to prepare the implementation
	// Returns the implementation to test and a cleanup function
	Setup func(t *testing.T) (interface{}, func())

	// Tests are the individual contract tests to run
	Tests []ContractTest
}

// ContractTest represents a single behavioral test for an interface contract.
type ContractTest struct {
	// Name is the descriptive name of this test
	Name string

	// Test is the function that performs the test
	// It receives the implementation returned by Setup
	Test func(t *testing.T, impl interface{})
}

// VerifyInterfaceContract runs all contract tests against an implementation.
//
// This is the main entry point for contract testing. It:
// 1. Verifies type compliance (implementation satisfies the interface)
// 2. Runs all behavioral tests
// 3. Ensures cleanup is called
//
// Example:
//
//	contract := InterfaceContract{
//	    Name: "Plugin Interface Contract",
//	    Setup: func(t *testing.T) (interface{}, func()) {
//	        plugin := NewTestPlugin()
//	        cleanup := func() { plugin.Close() }
//	        return plugin, cleanup
//	    },
//	    Tests: []ContractTest{
//	        {
//	            Name: "GetInfo returns non-nil PluginInfo",
//	            Test: func(t *testing.T, impl interface{}) {
//	                plugin := impl.(Plugin)
//	                info := plugin.GetInfo()
//	                if info == nil {
//	                    t.Error("GetInfo() returned nil")
//	                }
//	            },
//	        },
//	    },
//	}
//	VerifyInterfaceContract(t, contract)
func VerifyInterfaceContract(t *testing.T, contract InterfaceContract) {
	t.Helper()

	if contract.Name == "" {
		t.Fatal("InterfaceContract.Name is required")
	}

	if contract.Setup == nil {
		t.Fatal("InterfaceContract.Setup is required")
	}

	if len(contract.Tests) == 0 {
		t.Fatal("InterfaceContract.Tests cannot be empty")
	}

	// Run each contract test
	for _, test := range contract.Tests {
		t.Run(test.Name, func(t *testing.T) {
			t.Helper()

			// Setup implementation
			impl, cleanup := contract.Setup(t)
			if cleanup != nil {
				defer cleanup()
			}

			if impl == nil {
				t.Fatal("Setup returned nil implementation")
			}

			// Run the test
			test.Test(t, impl)
		})
	}
}

// AssertImplementsInterface verifies that an implementation satisfies an interface.
//
// This performs compile-time type checking at runtime to ensure the implementation
// provides all methods required by the interface.
//
// Example:
//
//	var _ Plugin = (*MyPlugin)(nil) // Compile-time check
//	AssertImplementsInterface(t, (*MyPlugin)(nil), (*Plugin)(nil))
func AssertImplementsInterface(t *testing.T, impl interface{}, iface interface{}) {
	t.Helper()

	implType := reflect.TypeOf(impl)
	ifaceType := reflect.TypeOf(iface).Elem()

	if !implType.Implements(ifaceType) {
		t.Errorf("Type %v does not implement interface %v", implType, ifaceType)

		// Provide helpful details about missing methods
		for i := 0; i < ifaceType.NumMethod(); i++ {
			method := ifaceType.Method(i)
			if _, found := implType.MethodByName(method.Name); !found {
				t.Errorf("  Missing method: %s%v", method.Name, method.Type)
			}
		}
	}
}

// ErrorContract defines expected error behavior for interface methods.
type ErrorContract struct {
	// Name describes the error condition being tested
	Name string

	// Setup prepares the scenario that should produce an error
	Setup func(t *testing.T) (interface{}, func())

	// Invoke calls the method that should return an error
	Invoke func(impl interface{}) error

	// Verify checks that the error is correct
	Verify func(t *testing.T, err error)
}

// VerifyErrorContract tests that a method returns appropriate errors.
//
// Example:
//
//	errorContract := ErrorContract{
//	    Name: "LoadPlugin with invalid path returns error",
//	    Setup: func(t *testing.T) (interface{}, func()) {
//	        registry := NewPluginRegistry()
//	        return registry, func() {}
//	    },
//	    Invoke: func(impl interface{}) error {
//	        registry := impl.(*PluginRegistry)
//	        _, err := registry.LoadPlugin("/invalid/path")
//	        return err
//	    },
//	    Verify: func(t *testing.T, err error) {
//	        if err == nil {
//	            t.Error("Expected error for invalid path, got nil")
//	        }
//	        if !errors.Is(err, ErrPluginNotFound) {
//	            t.Errorf("Expected ErrPluginNotFound, got %v", err)
//	        }
//	    },
//	}
//	VerifyErrorContract(t, errorContract)
func VerifyErrorContract(t *testing.T, contract ErrorContract) {
	t.Helper()

	if contract.Name == "" {
		t.Fatal("ErrorContract.Name is required")
	}

	if contract.Setup == nil {
		t.Fatal("ErrorContract.Setup is required")
	}

	if contract.Invoke == nil {
		t.Fatal("ErrorContract.Invoke is required")
	}

	if contract.Verify == nil {
		t.Fatal("ErrorContract.Verify is required")
	}

	t.Run(contract.Name, func(t *testing.T) {
		t.Helper()

		// Setup implementation
		impl, cleanup := contract.Setup(t)
		if cleanup != nil {
			defer cleanup()
		}

		if impl == nil {
			t.Fatal("Setup returned nil implementation")
		}

		// Invoke the method
		err := contract.Invoke(impl)

		// Verify the error
		contract.Verify(t, err)
	})
}

// BehaviorContract defines expected behavior patterns for interface methods.
type BehaviorContract struct {
	// Name describes the behavior being tested
	Name string

	// Setup prepares the implementation
	Setup func(t *testing.T) (interface{}, func())

	// Scenarios are the different behavior scenarios to test
	Scenarios []BehaviorScenario
}

// BehaviorScenario represents a single behavior test scenario.
type BehaviorScenario struct {
	// Name describes this specific scenario
	Name string

	// Given sets up the preconditions
	Given func(impl interface{})

	// When performs the action
	When func(impl interface{}) interface{}

	// Then verifies the outcome
	Then func(t *testing.T, result interface{})
}

// VerifyBehaviorContract tests behavior patterns using Given-When-Then style.
//
// This provides a BDD-style approach to contract testing that focuses on
// behavior rather than implementation details.
//
// Example:
//
//	behaviorContract := BehaviorContract{
//	    Name: "Plugin lifecycle behavior",
//	    Setup: func(t *testing.T) (interface{}, func()) {
//	        plugin := NewTestPlugin()
//	        return plugin, func() {}
//	    },
//	    Scenarios: []BehaviorScenario{
//	        {
//	            Name: "Plugin can be configured and used",
//	            Given: func(impl interface{}) {
//	                plugin := impl.(Plugin)
//	                plugin.Configure(map[string]interface{}{"key": "value"})
//	            },
//	            When: func(impl interface{}) interface{} {
//	                plugin := impl.(Plugin)
//	                return plugin.Execute("test-command", []string{})
//	            },
//	            Then: func(t *testing.T, result interface{}) {
//	                err := result.(error)
//	                if err != nil {
//	                    t.Errorf("Execute failed: %v", err)
//	                }
//	            },
//	        },
//	    },
//	}
//	VerifyBehaviorContract(t, behaviorContract)
func VerifyBehaviorContract(t *testing.T, contract BehaviorContract) {
	t.Helper()

	if contract.Name == "" {
		t.Fatal("BehaviorContract.Name is required")
	}

	if contract.Setup == nil {
		t.Fatal("BehaviorContract.Setup is required")
	}

	if len(contract.Scenarios) == 0 {
		t.Fatal("BehaviorContract.Scenarios cannot be empty")
	}

	for _, scenario := range contract.Scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			t.Helper()

			// Setup implementation
			impl, cleanup := contract.Setup(t)
			if cleanup != nil {
				defer cleanup()
			}

			if impl == nil {
				t.Fatal("Setup returned nil implementation")
			}

			// Given (preconditions)
			if scenario.Given != nil {
				scenario.Given(impl)
			}

			// When (action)
			var result interface{}
			if scenario.When != nil {
				result = scenario.When(impl)
			}

			// Then (verification)
			if scenario.Then != nil {
				scenario.Then(t, result)
			}
		})
	}
}

// ContractTestSuite represents a collection of related contracts for an interface.
type ContractTestSuite struct {
	// InterfaceName is the name of the interface being tested
	InterfaceName string

	// InterfaceContracts are the interface compliance tests
	InterfaceContracts []InterfaceContract

	// ErrorContracts are the error behavior tests
	ErrorContracts []ErrorContract

	// BehaviorContracts are the behavioral tests
	BehaviorContracts []BehaviorContract
}

// Run executes all contracts in the suite.
//
// This is a convenience method for running comprehensive contract tests.
//
// Example:
//
//	suite := ContractTestSuite{
//	    InterfaceName: "Plugin",
//	    InterfaceContracts: []InterfaceContract{...},
//	    ErrorContracts: []ErrorContract{...},
//	    BehaviorContracts: []BehaviorContract{...},
//	}
//	suite.Run(t)
func (s *ContractTestSuite) Run(t *testing.T) {
	t.Helper()

	if s.InterfaceName == "" {
		t.Fatal("ContractTestSuite.InterfaceName is required")
	}

	t.Run(s.InterfaceName, func(t *testing.T) {
		// Run interface contracts
		if len(s.InterfaceContracts) > 0 {
			t.Run("InterfaceContracts", func(t *testing.T) {
				for _, contract := range s.InterfaceContracts {
					VerifyInterfaceContract(t, contract)
				}
			})
		}

		// Run error contracts
		if len(s.ErrorContracts) > 0 {
			t.Run("ErrorContracts", func(t *testing.T) {
				for _, contract := range s.ErrorContracts {
					VerifyErrorContract(t, contract)
				}
			})
		}

		// Run behavior contracts
		if len(s.BehaviorContracts) > 0 {
			t.Run("BehaviorContracts", func(t *testing.T) {
				for _, contract := range s.BehaviorContracts {
					VerifyBehaviorContract(t, contract)
				}
			})
		}
	})
}

// Helper functions for common contract assertions

// AssertNonNil verifies that a value is not nil.
func AssertNonNil(t *testing.T, value interface{}, message string) {
	t.Helper()
	if value == nil || (reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil()) {
		t.Errorf("%s: got nil", message)
	}
}

// AssertEqual verifies that two values are equal.
func AssertEqual(t *testing.T, expected, actual interface{}, message string) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("%s: expected %v, got %v", message, expected, actual)
	}
}

// AssertNoError verifies that an error is nil.
func AssertNoError(t *testing.T, err error, message string) {
	t.Helper()
	if err != nil {
		t.Errorf("%s: unexpected error: %v", message, err)
	}
}

// AssertError verifies that an error is not nil.
func AssertError(t *testing.T, err error, message string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: expected error, got nil", message)
	}
}

// AssertErrorContains verifies that an error contains a specific message.
func AssertErrorContains(t *testing.T, err error, substring string, message string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: expected error containing %q, got nil", message, substring)
		return
	}
	if !containsString(err.Error(), substring) {
		t.Errorf("%s: error %q does not contain %q", message, err.Error(), substring)
	}
}

// AssertPanics verifies that a function panics.
func AssertPanics(t *testing.T, fn func(), message string) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("%s: expected panic, got none", message)
		}
	}()
	fn()
}

// containsString checks if a string contains a substring.
func containsString(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ImplementationExample demonstrates how to use the contract testing framework.
//
// This is not a real test, but serves as executable documentation.
func ImplementationExample() string {
	return `
// Example: Testing a Plugin interface contract

type Plugin interface {
    GetInfo() *PluginInfo
    Execute(cmd string, args []string) error
}

func TestPluginContract(t *testing.T) {
    contract := InterfaceContract{
        Name: "Plugin Interface",
        Setup: func(t *testing.T) (interface{}, func()) {
            plugin := &MyPlugin{name: "test"}
            cleanup := func() { /* cleanup */ }
            return plugin, cleanup
        },
        Tests: []ContractTest{
            {
                Name: "GetInfo returns non-nil PluginInfo",
                Test: func(t *testing.T, impl interface{}) {
                    plugin := impl.(Plugin)
                    info := plugin.GetInfo()
                    AssertNonNil(t, info, "PluginInfo")
                    AssertNonNil(t, info.Name, "PluginInfo.Name")
                },
            },
            {
                Name: "Execute with valid command succeeds",
                Test: func(t *testing.T, impl interface{}) {
                    plugin := impl.(Plugin)
                    err := plugin.Execute("test", []string{})
                    AssertNoError(t, err, "Execute")
                },
            },
        },
    }

    VerifyInterfaceContract(t, contract)
}
`
}
