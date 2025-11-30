# Table-Driven Test Framework

This document describes the table-driven test framework provided by the `testutil` package.

## Overview

Table-driven tests are a testing pattern where you define multiple test cases as data structures and iterate over them with a single test function. This approach:

- Reduces code duplication
- Makes it easy to add new test cases
- Improves test readability and maintainability
- Provides consistent test structure across the codebase

The `testutil` package provides a flexible, type-safe framework for writing table-driven tests with support for setup/teardown, parallel execution, and error handling.

## Core Types

### TestCase[T]

The `TestCase` struct represents a single test case in a table-driven test:

```go
type TestCase[T any] struct {
    Name          string      // Descriptive name for the test case
    Input         T           // Input data for the test
    Expected      any         // Expected result
    Setup         func(*testing.T)  // Optional per-case setup
    Teardown      func(*testing.T)  // Optional per-case teardown
    Skip          bool        // Skip this test case
    SkipReason    string      // Reason for skipping
    Parallel      bool        // Run this case in parallel
    ExpectError   bool        // Expect an error result
    ErrorContains string      // Expected error message substring
}
```

### TableTestOptions

Options for configuring the test runner:

```go
type TableTestOptions struct {
    Parallel        bool              // Run all cases in parallel
    GlobalSetup     func(*testing.T)  // Run once before all cases
    GlobalTeardown  func(*testing.T)  // Run once after all cases
}
```

## Functions

### RunTableTests

The main function for running table-driven tests:

```go
func RunTableTests[T any, R any](
    t *testing.T,
    cases []TestCase[T],
    opts *TableTestOptions,
    testFn TestFunc[T, R],
)
```

**Parameters:**
- `t`: Testing instance
- `cases`: Array of test cases to run
- `opts`: Optional configuration (can be nil)
- `testFn`: Function that implements the test logic

**Example:**

```go
cases := []testutil.TestCase[string]{
    {Name: "uppercase", Input: "hello", Expected: "HELLO"},
    {Name: "empty", Input: "", Expected: ""},
}

testutil.RunTableTests(t, cases, nil, func(t *testing.T, input string) (string, error) {
    return strings.ToUpper(input), nil
})
```

### RunSimpleTableTests

Simplified API for tests that don't return errors:

```go
func RunSimpleTableTests[T any, R any](
    t *testing.T,
    cases []TestCase[T],
    opts *TableTestOptions,
    testFn func(*testing.T, T) R,
)
```

**Example:**

```go
cases := []testutil.TestCase[int]{
    {Name: "double 5", Input: 5, Expected: 10},
    {Name: "double 0", Input: 0, Expected: 0},
}

testutil.RunSimpleTableTests(t, cases, nil, func(t *testing.T, input int) int {
    return input * 2
})
```

### RunTableTestsWithContext

Variant that provides cleanup function registration:

```go
func RunTableTestsWithContext[T any, R any](
    t *testing.T,
    cases []TestCase[T],
    opts *TableTestOptions,
    testFn func(*testing.T, func(func()), T) (R, error),
)
```

**Example:**

```go
testutil.RunTableTestsWithContext(t, cases, nil,
    func(t *testing.T, cleanup func(func()), input string) (string, error) {
        file := createTempFile()
        cleanup(func() { os.Remove(file) })
        return processFile(file), nil
    })
```

## Usage Patterns

### Basic Table Test

The simplest form of table-driven test:

```go
func TestBasic(t *testing.T) {
    cases := []testutil.TestCase[string]{
        {
            Name:     "convert to uppercase",
            Input:    "hello",
            Expected: "HELLO",
        },
        {
            Name:     "already uppercase",
            Input:    "WORLD",
            Expected: "WORLD",
        },
    }

    testutil.RunTableTests(t, cases, nil, func(t *testing.T, input string) (string, error) {
        return strings.ToUpper(input), nil
    })
}
```

### Testing Error Cases

Use `ExpectError` and `ErrorContains` to test error conditions:

```go
func TestWithErrors(t *testing.T) {
    cases := []testutil.TestCase[int]{
        {
            Name:     "valid input",
            Input:    5,
            Expected: 25,
        },
        {
            Name:        "negative should error",
            Input:       -1,
            ExpectError: true,
        },
        {
            Name:          "zero with specific error",
            Input:         0,
            ExpectError:   true,
            ErrorContains: "cannot be zero",
        },
    }

    testutil.RunTableTests(t, cases, nil, func(t *testing.T, input int) (int, error) {
        if input <= 0 {
            return 0, fmt.Errorf("input cannot be zero or negative")
        }
        return input * input, nil
    })
}
```

### Setup and Teardown

Per-case setup and teardown functions:

```go
func TestWithSetupTeardown(t *testing.T) {
    cases := []testutil.TestCase[string]{
        {
            Name:     "test with resources",
            Input:    "data",
            Expected: "processed",
            Setup: func(t *testing.T) {
                // Prepare test fixtures
                createTestFile()
            },
            Teardown: func(t *testing.T) {
                // Clean up resources
                removeTestFile()
            },
        },
    }

    testutil.RunTableTests(t, cases, nil, yourTestFunc)
}
```

### Global Setup and Teardown

Run setup/teardown once for all test cases:

```go
func TestWithGlobalSetup(t *testing.T) {
    cases := []testutil.TestCase[string]{
        {Name: "test 1", Input: "a", Expected: "A"},
        {Name: "test 2", Input: "b", Expected: "B"},
    }

    opts := &testutil.TableTestOptions{
        GlobalSetup: func(t *testing.T) {
            // Initialize shared state
            initDatabase()
        },
        GlobalTeardown: func(t *testing.T) {
            // Clean up shared state
            closeDatabase()
        },
    }

    testutil.RunTableTests(t, cases, opts, yourTestFunc)
}
```

### Parallel Execution

Run specific test cases in parallel:

```go
func TestParallel(t *testing.T) {
    cases := []testutil.TestCase[int]{
        {
            Name:     "parallel test 1",
            Input:    1,
            Expected: 2,
            Parallel: true,  // This case runs in parallel
        },
        {
            Name:     "serial test",
            Input:    2,
            Expected: 4,
            Parallel: false, // This case runs serially
        },
    }

    testutil.RunTableTests(t, cases, nil, yourTestFunc)
}
```

Or run all cases in parallel:

```go
func TestAllParallel(t *testing.T) {
    cases := []testutil.TestCase[int]{
        {Name: "test 1", Input: 1, Expected: 2},
        {Name: "test 2", Input: 2, Expected: 4},
    }

    opts := &testutil.TableTestOptions{
        Parallel: true, // All cases run in parallel
    }

    testutil.RunTableTests(t, cases, opts, yourTestFunc)
}
```

### Skipping Test Cases

Skip specific test cases during development:

```go
func TestWithSkip(t *testing.T) {
    cases := []testutil.TestCase[string]{
        {
            Name:     "implemented feature",
            Input:    "test",
            Expected: "TEST",
        },
        {
            Name:       "future feature",
            Input:      "future",
            Skip:       true,
            SkipReason: "Feature not implemented yet",
        },
    }

    testutil.RunTableTests(t, cases, nil, yourTestFunc)
}
```

### Complex Input/Output Types

Use structs for complex test scenarios:

```go
func TestComplexTypes(t *testing.T) {
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
    }

    testutil.RunTableTests(t, cases, nil, func(t *testing.T, input Input) (Output, error) {
        return Output{
            Message: fmt.Sprintf("Hello %s", input.Name),
            Double:  input.Value * 2,
        }, nil
    })
}
```

## When to Use Table Tests

### ✅ Good Use Cases

Table-driven tests work well when:

1. **Testing multiple inputs/outputs**: You have many similar test cases with different inputs and expected outputs
2. **Testing edge cases**: You want to systematically test boundary conditions
3. **Regression testing**: Adding new test cases for each bug fix
4. **Parametric testing**: Testing the same logic with different parameters
5. **Consistent behavior**: Verifying consistent behavior across multiple scenarios

**Examples:**
- String parsing/formatting functions
- Validation logic
- Math operations
- Configuration parsing
- Data transformations

### ❌ When NOT to Use

Avoid table-driven tests when:

1. **Complex test logic**: Each test case requires significantly different setup or assertions
2. **Single test case**: Only one or two test cases (use regular test functions)
3. **Shared mutable state**: Test cases interfere with each other
4. **Complex assertions**: Verification logic varies significantly between cases
5. **Integration tests**: Tests with complex dependencies or side effects

## Best Practices

### 1. Descriptive Test Names

Use clear, descriptive names for each test case:

```go
// ✅ Good
{Name: "empty string returns empty string", Input: "", Expected: ""}
{Name: "nil input returns error", Input: nil, ExpectError: true}

// ❌ Bad
{Name: "test1", Input: "", Expected: ""}
{Name: "error case", Input: nil, ExpectError: true}
```

### 2. One Behavior Per Test

Each table should test one specific behavior:

```go
// ✅ Good - Tests string uppercasing
func TestToUpper(t *testing.T) {
    cases := []testutil.TestCase[string]{
        {Name: "lowercase", Input: "hello", Expected: "HELLO"},
        {Name: "uppercase", Input: "HELLO", Expected: "HELLO"},
    }
    testutil.RunTableTests(t, cases, nil, toUpperFunc)
}

// ❌ Bad - Mixed behaviors
func TestStringFunctions(t *testing.T) {
    cases := []testutil.TestCase[string]{
        {Name: "uppercase", Input: "hello", Expected: "HELLO"},
        {Name: "reverse", Input: "hello", Expected: "olleh"},  // Different behavior
    }
    testutil.RunTableTests(t, cases, nil, ???)  // What function to test?
}
```

### 3. Group Related Cases

Organize test cases logically:

```go
func TestValidation(t *testing.T) {
    cases := []testutil.TestCase[string]{
        // Valid inputs
        {Name: "valid email", Input: "user@example.com", Expected: true},
        {Name: "valid with subdomain", Input: "user@mail.example.com", Expected: true},

        // Invalid inputs
        {Name: "missing @", Input: "userexample.com", ExpectError: true},
        {Name: "missing domain", Input: "user@", ExpectError: true},
        {Name: "empty string", Input: "", ExpectError: true},
    }

    testutil.RunTableTests(t, cases, nil, validateEmailFunc)
}
```

### 4. Use Setup/Teardown Wisely

Only use setup/teardown when necessary:

```go
// ✅ Good - Cleanup required
{
    Name:  "create temp file",
    Input: "data",
    Setup: func(t *testing.T) {
        createTempFile()
    },
    Teardown: func(t *testing.T) {
        removeTempFile()  // Necessary cleanup
    },
}

// ❌ Bad - Unnecessary setup
{
    Name:  "simple test",
    Input: "data",
    Setup: func(t *testing.T) {
        // Nothing to set up
    },
}
```

### 5. Parallel Test Considerations

Be careful with parallel execution and shared state:

```go
// ✅ Good - No shared state
func TestParallelSafe(t *testing.T) {
    cases := []testutil.TestCase[int]{
        {Name: "test 1", Input: 1, Expected: 2, Parallel: true},
        {Name: "test 2", Input: 2, Expected: 4, Parallel: true},
    }
    opts := &testutil.TableTestOptions{Parallel: true}
    testutil.RunTableTests(t, cases, opts, pureFunction)
}

// ❌ Bad - Shared mutable state
var counter int // UNSAFE with parallel tests

func TestParallelUnsafe(t *testing.T) {
    cases := []testutil.TestCase[int]{
        {Name: "test 1", Input: 1, Parallel: true},  // Race condition!
        {Name: "test 2", Input: 2, Parallel: true},
    }
    testutil.RunTableTests(t, cases, nil, func(t *testing.T, input int) (int, error) {
        counter++  // UNSAFE!
        return input, nil
    })
}
```

### 6. Keep Test Functions Simple

The test function should be simple and focused:

```go
// ✅ Good - Simple, focused test function
testutil.RunTableTests(t, cases, nil, func(t *testing.T, input string) (string, error) {
    return myPackage.Process(input)
})

// ❌ Bad - Complex logic in test function
testutil.RunTableTests(t, cases, nil, func(t *testing.T, input string) (string, error) {
    // Too much logic here
    if input == "special" {
        setupSpecialCase()
    }
    result := myPackage.Process(input)
    if result != "" {
        validateResult(result)
    }
    return result, nil
})
```

### 7. Error Testing Best Practices

Be specific about expected errors:

```go
// ✅ Good - Specific error checking
{
    Name:          "invalid format",
    Input:         "bad-format",
    ExpectError:   true,
    ErrorContains: "invalid format", // Verify specific error
}

// ❌ Less ideal - Just checking any error
{
    Name:        "invalid format",
    Input:       "bad-format",
    ExpectError: true, // Any error accepted
}
```

## Comparison with Traditional Tests

### Traditional Approach

```go
func TestToUpper(t *testing.T) {
    // Test 1
    result := strings.ToUpper("hello")
    if result != "HELLO" {
        t.Errorf("Expected HELLO, got %s", result)
    }

    // Test 2
    result = strings.ToUpper("world")
    if result != "WORLD" {
        t.Errorf("Expected WORLD, got %s", result)
    }

    // Test 3
    result = strings.ToUpper("")
    if result != "" {
        t.Errorf("Expected empty string, got %s", result)
    }
}
```

### Table-Driven Approach

```go
func TestToUpper(t *testing.T) {
    cases := []testutil.TestCase[string]{
        {Name: "lowercase", Input: "hello", Expected: "HELLO"},
        {Name: "lowercase", Input: "world", Expected: "WORLD"},
        {Name: "empty", Input: "", Expected: ""},
    }

    testutil.RunTableTests(t, cases, nil, func(t *testing.T, input string) (string, error) {
        return strings.ToUpper(input), nil
    })
}
```

**Benefits:**
- Less code duplication
- Easier to add new cases
- Consistent error reporting
- Better test organization
- Each case runs as a subtest (better failure reporting)

## Migration Guide

To convert existing tests to table-driven tests:

1. **Identify the pattern**: Look for repeated test code with different inputs
2. **Extract the test logic**: Move the core logic into a test function
3. **Define test cases**: Create a slice of `TestCase` structs
4. **Use RunTableTests**: Call the appropriate `Run*` function
5. **Verify behavior**: Ensure all tests still pass

Example migration:

```go
// Before
func TestOld(t *testing.T) {
    if result := Double(5); result != 10 {
        t.Error("failed")
    }
    if result := Double(0); result != 0 {
        t.Error("failed")
    }
}

// After
func TestNew(t *testing.T) {
    cases := []testutil.TestCase[int]{
        {Name: "double 5", Input: 5, Expected: 10},
        {Name: "double 0", Input: 0, Expected: 0},
    }
    testutil.RunSimpleTableTests(t, cases, nil, func(t *testing.T, input int) int {
        return Double(input)
    })
}
```

## Examples in the Codebase

See `tests/testutil/table_test.go` for comprehensive examples including:

- Simple table tests
- Error handling
- Setup/teardown (per-case and global)
- Parallel execution
- Skipped tests
- Complex input/output types
- Cleanup with context

## See Also

- [Go Testing Documentation](https://pkg.go.dev/testing)
- [Table-Driven Tests in Go](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [testutil Package README](./README.md)
