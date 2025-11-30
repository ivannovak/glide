// Package testutil provides testing utilities for table-driven tests.
package testutil

import (
	"testing"
)

// TestCase represents a single test case in a table-driven test.
// Type parameter T represents the test input/output type.
type TestCase[T any] struct {
	// Name is the descriptive name for this test case
	Name string

	// Input is the input data for the test case
	Input T

	// Expected is the expected result for this test case
	Expected any

	// Setup is an optional function to run before the test case executes.
	// Use this to prepare test fixtures, mock expectations, or test state.
	Setup func(t *testing.T)

	// Teardown is an optional function to run after the test case completes.
	// Use this to clean up resources, verify mock expectations, or reset state.
	Teardown func(t *testing.T)

	// Skip, if true, will skip this test case
	Skip bool

	// SkipReason provides an explanation for why the test case is skipped
	SkipReason string

	// Parallel, if true, will run this test case in parallel with other parallel cases
	Parallel bool

	// ExpectError, if true, indicates this test case should produce an error
	ExpectError bool

	// ErrorContains, if set, will verify the error message contains this string
	ErrorContains string
}

// TableTestOptions configures behavior for RunTableTests.
type TableTestOptions struct {
	// Parallel, if true, runs all test cases in parallel
	Parallel bool

	// GlobalSetup runs once before all test cases
	GlobalSetup func(t *testing.T)

	// GlobalTeardown runs once after all test cases
	GlobalTeardown func(t *testing.T)
}

// TestFunc is the function signature for table-driven test functions.
// It takes a test case and returns a result and error.
type TestFunc[T any, R any] func(t *testing.T, input T) (R, error)

// RunTableTests executes a table-driven test with the provided test cases and test function.
//
// Example:
//
//	cases := []testutil.TestCase[string]{
//	    {Name: "valid input", Input: "test", Expected: "TEST"},
//	    {Name: "empty input", Input: "", ExpectError: true},
//	}
//	testutil.RunTableTests(t, cases, nil, func(t *testing.T, input string) (string, error) {
//	    return strings.ToUpper(input), nil
//	})
func RunTableTests[T any, R any](
	t *testing.T,
	cases []TestCase[T],
	opts *TableTestOptions,
	testFn TestFunc[T, R],
) {
	t.Helper()

	// Apply default options
	if opts == nil {
		opts = &TableTestOptions{}
	}

	// Run global setup if provided
	if opts.GlobalSetup != nil {
		opts.GlobalSetup(t)
	}

	// Run global teardown if provided
	if opts.GlobalTeardown != nil {
		defer opts.GlobalTeardown(t)
	}

	// Execute each test case
	for _, tc := range cases {
		tc := tc // Capture range variable

		t.Run(tc.Name, func(t *testing.T) {
			t.Helper()

			// Skip if requested
			if tc.Skip {
				if tc.SkipReason != "" {
					t.Skip(tc.SkipReason)
				} else {
					t.Skip()
				}
			}

			// Run in parallel if requested
			if tc.Parallel || opts.Parallel {
				t.Parallel()
			}

			// Run setup if provided
			if tc.Setup != nil {
				tc.Setup(t)
			}

			// Run teardown if provided
			if tc.Teardown != nil {
				defer tc.Teardown(t)
			}

			// Execute the test function
			result, err := testFn(t, tc.Input)

			// Handle error expectations
			if tc.ExpectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tc.ErrorContains != "" {
					AssertErrorContains(t, err, tc.ErrorContains)
				}
				return
			}

			// Verify no unexpected error
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify result matches expected
			if tc.Expected != nil {
				AssertStructEqual(t, tc.Expected, result)
			}
		})
	}
}

// RunSimpleTableTests is a convenience wrapper for tests that only need to verify
// the result without error handling.
//
// Example:
//
//	cases := []testutil.TestCase[int]{
//	    {Name: "double 5", Input: 5, Expected: 10},
//	    {Name: "double 0", Input: 0, Expected: 0},
//	}
//	testutil.RunSimpleTableTests(t, cases, nil, func(t *testing.T, input int) int {
//	    return input * 2
//	})
func RunSimpleTableTests[T any, R any](
	t *testing.T,
	cases []TestCase[T],
	opts *TableTestOptions,
	testFn func(t *testing.T, input T) R,
) {
	t.Helper()

	// Wrap the simple function in a TestFunc
	wrappedFn := func(t *testing.T, input T) (R, error) {
		return testFn(t, input), nil
	}

	RunTableTests(t, cases, opts, wrappedFn)
}

// RunTableTestsWithContext is a variant that provides a test-specific cleanup
// function to each test case.
//
// Example:
//
//	cases := []testutil.TestCase[string]{
//	    {Name: "test with cleanup", Input: "data"},
//	}
//	testutil.RunTableTestsWithContext(t, cases, nil, func(t *testing.T, cleanup func(), input string) (string, error) {
//	    file := createTempFile()
//	    cleanup(func() { os.Remove(file) })
//	    return processFile(file), nil
//	})
func RunTableTestsWithContext[T any, R any](
	t *testing.T,
	cases []TestCase[T],
	opts *TableTestOptions,
	testFn func(t *testing.T, cleanup func(func()), input T) (R, error),
) {
	t.Helper()

	// Wrap the test function to provide cleanup
	wrappedFn := func(t *testing.T, input T) (R, error) {
		cleanup := func(fn func()) {
			t.Cleanup(fn)
		}
		return testFn(t, cleanup, input)
	}

	RunTableTests(t, cases, opts, wrappedFn)
}
