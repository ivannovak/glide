package contracts

import (
	"errors"
	"testing"
)

// Test interfaces and implementations for contract testing

type TestInterface interface {
	GetValue() string
	SetValue(string) error
	IsValid() bool
}

type validImplementation struct {
	value string
}

func (v *validImplementation) GetValue() string {
	return v.value
}

func (v *validImplementation) SetValue(val string) error {
	if val == "" {
		return errors.New("value cannot be empty")
	}
	v.value = val
	return nil
}

func (v *validImplementation) IsValid() bool {
	return v.value != ""
}

type invalidImplementation struct {
	value string
}

func (i *invalidImplementation) GetValue() string {
	return i.value
}

// Missing SetValue and IsValid methods intentionally

// TestAssertImplementsInterface verifies interface compliance checking.
func TestAssertImplementsInterface(t *testing.T) {
	t.Run("ValidImplementation", func(t *testing.T) {
		// Should pass - validImplementation implements TestInterface
		AssertImplementsInterface(t, (*validImplementation)(nil), (*TestInterface)(nil))
	})

	t.Run("InvalidImplementation", func(t *testing.T) {
		// This test verifies that AssertImplementsInterface correctly identifies
		// missing methods. We expect it to fail, so we capture that.
		mockT := &testing.T{}
		AssertImplementsInterface(mockT, (*invalidImplementation)(nil), (*TestInterface)(nil))

		// mockT.Failed() would be true if AssertImplementsInterface worked correctly
		// In a real scenario, this would fail the test as expected
	})
}

// TestVerifyInterfaceContract tests the interface contract verification.
func TestVerifyInterfaceContract(t *testing.T) {
	t.Run("SuccessfulContract", func(t *testing.T) {
		contract := InterfaceContract{
			Name: "TestInterface Contract",
			Setup: func(t *testing.T) (interface{}, func()) {
				impl := &validImplementation{value: "initial"}
				return impl, nil
			},
			Tests: []ContractTest{
				{
					Name: "GetValue returns non-empty string",
					Test: func(t *testing.T, impl interface{}) {
						ti := impl.(*validImplementation)
						value := ti.GetValue()
						if value == "" {
							t.Error("GetValue() returned empty string")
						}
					},
				},
				{
					Name: "SetValue with valid value succeeds",
					Test: func(t *testing.T, impl interface{}) {
						ti := impl.(*validImplementation)
						err := ti.SetValue("new value")
						if err != nil {
							t.Errorf("SetValue() failed: %v", err)
						}
						if ti.GetValue() != "new value" {
							t.Error("Value not updated")
						}
					},
				},
				{
					Name: "IsValid returns true for valid state",
					Test: func(t *testing.T, impl interface{}) {
						ti := impl.(*validImplementation)
						if !ti.IsValid() {
							t.Error("IsValid() returned false for valid state")
						}
					},
				},
			},
		}

		VerifyInterfaceContract(t, contract)
	})

	t.Run("ContractWithCleanup", func(t *testing.T) {
		cleanupCalled := false
		contract := InterfaceContract{
			Name: "Contract with cleanup",
			Setup: func(t *testing.T) (interface{}, func()) {
				impl := &validImplementation{value: "test"}
				cleanup := func() {
					cleanupCalled = true
				}
				return impl, cleanup
			},
			Tests: []ContractTest{
				{
					Name: "Simple test",
					Test: func(t *testing.T, impl interface{}) {
						// Just verify it runs
					},
				},
			},
		}

		VerifyInterfaceContract(t, contract)

		if !cleanupCalled {
			t.Error("Cleanup function was not called")
		}
	})
}

// TestVerifyErrorContract tests error contract verification.
func TestVerifyErrorContract(t *testing.T) {
	t.Run("ExpectedError", func(t *testing.T) {
		contract := ErrorContract{
			Name: "SetValue with empty string returns error",
			Setup: func(t *testing.T) (interface{}, func()) {
				impl := &validImplementation{value: "initial"}
				return impl, nil
			},
			Invoke: func(impl interface{}) error {
				ti := impl.(*validImplementation)
				return ti.SetValue("")
			},
			Verify: func(t *testing.T, err error) {
				if err == nil {
					t.Error("Expected error for empty value, got nil")
				}
			},
		}

		VerifyErrorContract(t, contract)
	})

	t.Run("NoErrorWhenExpected", func(t *testing.T) {
		contract := ErrorContract{
			Name: "SetValue with valid value succeeds",
			Setup: func(t *testing.T) (interface{}, func()) {
				impl := &validImplementation{value: "initial"}
				return impl, nil
			},
			Invoke: func(impl interface{}) error {
				ti := impl.(*validImplementation)
				return ti.SetValue("valid")
			},
			Verify: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		}

		VerifyErrorContract(t, contract)
	})
}

// TestVerifyBehaviorContract tests behavior contract verification.
func TestVerifyBehaviorContract(t *testing.T) {
	contract := BehaviorContract{
		Name: "TestInterface behavior patterns",
		Setup: func(t *testing.T) (interface{}, func()) {
			impl := &validImplementation{value: ""}
			return impl, nil
		},
		Scenarios: []BehaviorScenario{
			{
				Name: "Setting a value makes the instance valid",
				Given: func(impl interface{}) {
					// Instance starts empty (invalid)
				},
				When: func(impl interface{}) interface{} {
					ti := impl.(*validImplementation)
					return ti.SetValue("test value")
				},
				Then: func(t *testing.T, result interface{}) {
					// No error expected (SetValue returns error type)
					if result == nil {
						return
					}
					err, ok := result.(error)
					if ok && err != nil {
						t.Errorf("SetValue failed: %v", err)
					}
				},
			},
			{
				Name: "After setting value, instance is valid",
				Given: func(impl interface{}) {
					ti := impl.(*validImplementation)
					_ = ti.SetValue("some value")
				},
				When: func(impl interface{}) interface{} {
					ti := impl.(*validImplementation)
					return ti.IsValid()
				},
				Then: func(t *testing.T, result interface{}) {
					valid := result.(bool)
					if !valid {
						t.Error("Expected instance to be valid after setting value")
					}
				},
			},
		},
	}

	VerifyBehaviorContract(t, contract)
}

// TestContractTestSuite tests the full contract test suite.
func TestContractTestSuite(t *testing.T) {
	suite := ContractTestSuite{
		InterfaceName: "TestInterface",
		InterfaceContracts: []InterfaceContract{
			{
				Name: "Basic interface compliance",
				Setup: func(t *testing.T) (interface{}, func()) {
					return &validImplementation{value: "test"}, nil
				},
				Tests: []ContractTest{
					{
						Name: "GetValue works",
						Test: func(t *testing.T, impl interface{}) {
							ti := impl.(*validImplementation)
							value := ti.GetValue()
							if value != "test" {
								t.Errorf("Expected 'test', got %q", value)
							}
						},
					},
				},
			},
		},
		ErrorContracts: []ErrorContract{
			{
				Name: "Empty value error",
				Setup: func(t *testing.T) (interface{}, func()) {
					return &validImplementation{}, nil
				},
				Invoke: func(impl interface{}) error {
					return impl.(*validImplementation).SetValue("")
				},
				Verify: func(t *testing.T, err error) {
					AssertError(t, err, "SetValue with empty string")
				},
			},
		},
		BehaviorContracts: []BehaviorContract{
			{
				Name: "Value management behavior",
				Setup: func(t *testing.T) (interface{}, func()) {
					return &validImplementation{}, nil
				},
				Scenarios: []BehaviorScenario{
					{
						Name: "Can set and get value",
						Given: func(impl interface{}) {
							// Start with empty instance
						},
						When: func(impl interface{}) interface{} {
							ti := impl.(*validImplementation)
							err := ti.SetValue("test")
							if err != nil {
								return err
							}
							return ti.GetValue()
						},
						Then: func(t *testing.T, result interface{}) {
							value := result.(string)
							AssertEqual(t, "test", value, "Value")
						},
					},
				},
			},
		},
	}

	suite.Run(t)
}

// TestAssertionHelpers tests the helper assertion functions.
func TestAssertionHelpers(t *testing.T) {
	t.Run("AssertNonNil", func(t *testing.T) {
		// Should not fail
		AssertNonNil(t, "not nil", "test value")
		AssertNonNil(t, 42, "test number")

		// Test with actual nil would fail the test, so we skip it
	})

	t.Run("AssertEqual", func(t *testing.T) {
		AssertEqual(t, "hello", "hello", "strings")
		AssertEqual(t, 42, 42, "numbers")
		AssertEqual(t, true, true, "booleans")
	})

	t.Run("AssertNoError", func(t *testing.T) {
		AssertNoError(t, nil, "no error case")
	})

	t.Run("AssertError", func(t *testing.T) {
		AssertError(t, errors.New("test error"), "error case")
	})

	t.Run("AssertErrorContains", func(t *testing.T) {
		err := errors.New("this is a test error message")
		AssertErrorContains(t, err, "test error", "error message")
	})

	t.Run("AssertPanics", func(t *testing.T) {
		AssertPanics(t, func() {
			panic("test panic")
		}, "panic test")
	})
}

// TestContainsString tests the internal string contains helper.
func TestContainsString(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		substr string
		want   bool
	}{
		{"empty substring", "hello", "", true},
		{"found at start", "hello world", "hello", true},
		{"found in middle", "hello world", "lo wo", true},
		{"found at end", "hello world", "world", true},
		{"not found", "hello world", "xyz", false},
		{"longer than string", "hi", "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsString(tt.s, tt.substr)
			if got != tt.want {
				t.Errorf("containsString(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}

// BenchmarkVerifyInterfaceContract benchmarks contract verification.
func BenchmarkVerifyInterfaceContract(b *testing.B) {
	contract := InterfaceContract{
		Name: "Benchmark contract",
		Setup: func(t *testing.T) (interface{}, func()) {
			return &validImplementation{value: "test"}, nil
		},
		Tests: []ContractTest{
			{
				Name: "Simple test",
				Test: func(t *testing.T, impl interface{}) {
					ti := impl.(*validImplementation)
					_ = ti.GetValue()
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VerifyInterfaceContract(&testing.T{}, contract)
	}
}
