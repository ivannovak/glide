package testutil_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/ivannovak/glide/v2/tests/testutil"
)

// TestRunTableTests_Simple demonstrates a basic table-driven test
func TestRunTableTests_Simple(t *testing.T) {
	cases := []testutil.TestCase[string]{
		{
			Name:     "uppercase single word",
			Input:    "hello",
			Expected: "HELLO",
		},
		{
			Name:     "uppercase multiple words",
			Input:    "hello world",
			Expected: "HELLO WORLD",
		},
		{
			Name:     "already uppercase",
			Input:    "HELLO",
			Expected: "HELLO",
		},
		{
			Name:     "empty string",
			Input:    "",
			Expected: "",
		},
	}

	testutil.RunTableTests(t, cases, nil, func(t *testing.T, input string) (string, error) {
		return strings.ToUpper(input), nil
	})
}

// TestRunTableTests_WithErrors demonstrates error handling in table tests
func TestRunTableTests_WithErrors(t *testing.T) {
	cases := []testutil.TestCase[int]{
		{
			Name:     "valid positive number",
			Input:    5,
			Expected: 25,
		},
		{
			Name:        "negative number should error",
			Input:       -1,
			ExpectError: true,
		},
		{
			Name:          "zero should error with specific message",
			Input:         0,
			ExpectError:   true,
			ErrorContains: "cannot be zero",
		},
	}

	testutil.RunTableTests(t, cases, nil, func(t *testing.T, input int) (int, error) {
		if input < 0 {
			return 0, errors.New("input cannot be negative")
		}
		if input == 0 {
			return 0, errors.New("input cannot be zero")
		}
		return input * input, nil
	})
}

// TestRunTableTests_WithSetupTeardown demonstrates setup/teardown per test case
func TestRunTableTests_WithSetupTeardown(t *testing.T) {
	var setupCount, teardownCount int

	cases := []testutil.TestCase[string]{
		{
			Name:     "test with setup and teardown",
			Input:    "data",
			Expected: "DATA-processed",
			Setup: func(t *testing.T) {
				setupCount++
				t.Logf("Setup called (count: %d)", setupCount)
			},
			Teardown: func(t *testing.T) {
				teardownCount++
				t.Logf("Teardown called (count: %d)", teardownCount)
			},
		},
		{
			Name:     "another test with setup and teardown",
			Input:    "more",
			Expected: "MORE-processed",
			Setup: func(t *testing.T) {
				setupCount++
				t.Logf("Setup called (count: %d)", setupCount)
			},
			Teardown: func(t *testing.T) {
				teardownCount++
				t.Logf("Teardown called (count: %d)", teardownCount)
			},
		},
	}

	testutil.RunTableTests(t, cases, nil, func(t *testing.T, input string) (string, error) {
		return strings.ToUpper(input) + "-processed", nil
	})

	// Verify setup and teardown were called for each case
	if setupCount != 2 {
		t.Errorf("Expected 2 setup calls, got %d", setupCount)
	}
	if teardownCount != 2 {
		t.Errorf("Expected 2 teardown calls, got %d", teardownCount)
	}
}

// TestRunTableTests_WithGlobalSetupTeardown demonstrates global setup/teardown
func TestRunTableTests_WithGlobalSetupTeardown(t *testing.T) {
	var globalState string

	cases := []testutil.TestCase[string]{
		{
			Name:     "first test",
			Input:    "a",
			Expected: "global-a",
		},
		{
			Name:     "second test",
			Input:    "b",
			Expected: "global-b",
		},
	}

	opts := &testutil.TableTestOptions{
		GlobalSetup: func(t *testing.T) {
			globalState = "global"
			t.Logf("Global setup called")
		},
		GlobalTeardown: func(t *testing.T) {
			globalState = ""
			t.Logf("Global teardown called")
		},
	}

	testutil.RunTableTests(t, cases, opts, func(t *testing.T, input string) (string, error) {
		return globalState + "-" + input, nil
	})

	// Verify global teardown was called
	if globalState != "" {
		t.Errorf("Expected global state to be cleaned up, got %q", globalState)
	}
}

// TestRunTableTests_Parallel demonstrates parallel test execution
func TestRunTableTests_Parallel(t *testing.T) {
	cases := []testutil.TestCase[int]{
		{
			Name:     "test 1",
			Input:    1,
			Expected: 2,
			Parallel: true,
		},
		{
			Name:     "test 2",
			Input:    2,
			Expected: 4,
			Parallel: true,
		},
		{
			Name:     "test 3",
			Input:    3,
			Expected: 6,
			Parallel: true,
		},
	}

	testutil.RunTableTests(t, cases, nil, func(t *testing.T, input int) (int, error) {
		return input * 2, nil
	})
}

// TestRunTableTests_AllParallel demonstrates running all tests in parallel
func TestRunTableTests_AllParallel(t *testing.T) {
	cases := []testutil.TestCase[int]{
		{
			Name:     "test 1",
			Input:    1,
			Expected: 2,
		},
		{
			Name:     "test 2",
			Input:    2,
			Expected: 4,
		},
		{
			Name:     "test 3",
			Input:    3,
			Expected: 6,
		},
	}

	opts := &testutil.TableTestOptions{
		Parallel: true,
	}

	testutil.RunTableTests(t, cases, opts, func(t *testing.T, input int) (int, error) {
		return input * 2, nil
	})
}

// TestRunTableTests_Skip demonstrates skipping test cases
func TestRunTableTests_Skip(t *testing.T) {
	cases := []testutil.TestCase[string]{
		{
			Name:     "this test runs",
			Input:    "run",
			Expected: "RUN",
		},
		{
			Name:       "this test is skipped",
			Input:      "skip",
			Skip:       true,
			SkipReason: "Not implemented yet",
		},
		{
			Name:     "this test also runs",
			Input:    "run",
			Expected: "RUN",
		},
	}

	testutil.RunTableTests(t, cases, nil, func(t *testing.T, input string) (string, error) {
		return strings.ToUpper(input), nil
	})
}

// TestRunSimpleTableTests demonstrates the simplified API for non-error tests
func TestRunSimpleTableTests(t *testing.T) {
	cases := []testutil.TestCase[int]{
		{
			Name:     "double positive",
			Input:    5,
			Expected: 10,
		},
		{
			Name:     "double negative",
			Input:    -3,
			Expected: -6,
		},
		{
			Name:     "double zero",
			Input:    0,
			Expected: 0,
		},
	}

	testutil.RunSimpleTableTests(t, cases, nil, func(t *testing.T, input int) int {
		return input * 2
	})
}

// TestRunTableTestsWithContext demonstrates using cleanup functions
func TestRunTableTestsWithContext(t *testing.T) {
	cases := []testutil.TestCase[string]{
		{
			Name:     "test with cleanup",
			Input:    "data",
			Expected: "DATA-cleaned",
		},
	}

	var cleaned bool

	testutil.RunTableTestsWithContext(t, cases, nil, func(t *testing.T, cleanup func(func()), input string) (string, error) {
		// Register cleanup function
		cleanup(func() {
			cleaned = true
		})

		return strings.ToUpper(input) + "-cleaned", nil
	})

	// Cleanup should have been called
	if !cleaned {
		t.Error("Expected cleanup to be called")
	}
}

// TestRunTableTests_ComplexStruct demonstrates testing with complex types
func TestRunTableTests_ComplexStruct(t *testing.T) {
	type Input struct {
		Name  string
		Value int
	}

	type Output struct {
		Message string
		Double  int
	}

	cases := []testutil.TestCase[Input]{
		{
			Name: "valid input",
			Input: Input{
				Name:  "test",
				Value: 5,
			},
			Expected: Output{
				Message: "Hello test",
				Double:  10,
			},
		},
		{
			Name: "empty name",
			Input: Input{
				Name:  "",
				Value: 3,
			},
			Expected: Output{
				Message: "Hello ",
				Double:  6,
			},
		},
	}

	testutil.RunTableTests(t, cases, nil, func(t *testing.T, input Input) (Output, error) {
		return Output{
			Message: fmt.Sprintf("Hello %s", input.Name),
			Double:  input.Value * 2,
		}, nil
	})
}

// TestRunTableTests_MixedParallel demonstrates mixing parallel and serial tests
func TestRunTableTests_MixedParallel(t *testing.T) {
	var sharedState int
	_ = sharedState // Acknowledge the variable is intentionally unused in test function

	cases := []testutil.TestCase[int]{
		{
			Name:     "serial test 1",
			Input:    1,
			Expected: 2, // input * 2
			Parallel: false,
			Setup: func(t *testing.T) {
				sharedState = 1
			},
		},
		{
			Name:     "parallel test 1",
			Input:    2,
			Expected: 4, // input * 2
			Parallel: true,
		},
		{
			Name:     "parallel test 2",
			Input:    3,
			Expected: 6, // input * 2
			Parallel: true,
		},
		{
			Name:     "serial test 2",
			Input:    4,
			Expected: 8, // input * 2
			Parallel: false,
			Setup: func(t *testing.T) {
				sharedState = 4
			},
		},
	}

	testutil.RunTableTests(t, cases, nil, func(t *testing.T, input int) (int, error) {
		// Parallel tests don't depend on sharedState
		return input * 2, nil
	})
}
