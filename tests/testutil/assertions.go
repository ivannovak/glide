package testutil

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// AssertNoError fails the test if err is not nil
func AssertNoError(t TestingT, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", msg, err)
	}
}

// AssertError fails the test if err is nil
func AssertError(t TestingT, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected an error but got nil", msg)
	}
}

// AssertErrorContains fails the test if err is nil or doesn't contain the substring
func AssertErrorContains(t TestingT, err error, substring string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q but got nil", substring)
	}
	if !strings.Contains(err.Error(), substring) {
		t.Fatalf("expected error to contain %q but got: %v", substring, err)
	}
}

// AssertEqual fails the test if expected and actual are not equal
func AssertEqual(t TestingT, expected, actual interface{}, msg string) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("%s: expected %v, got %v", msg, expected, actual)
	}
}

// AssertNotEqual fails the test if expected and actual are equal
func AssertNotEqual(t TestingT, expected, actual interface{}, msg string) {
	t.Helper()
	if reflect.DeepEqual(expected, actual) {
		t.Errorf("%s: expected values to be different, but both were %v", msg, expected)
	}
}

// AssertTrue fails the test if condition is false
func AssertTrue(t TestingT, condition bool, msg string) {
	t.Helper()
	if !condition {
		t.Errorf("%s: expected true but got false", msg)
	}
}

// AssertFalse fails the test if condition is true
func AssertFalse(t TestingT, condition bool, msg string) {
	t.Helper()
	if condition {
		t.Errorf("%s: expected false but got true", msg)
	}
}

// AssertNil fails the test if value is not nil
func AssertNil(t TestingT, value interface{}, msg string) {
	t.Helper()
	if value != nil && !reflect.ValueOf(value).IsNil() {
		t.Errorf("%s: expected nil but got %v", msg, value)
	}
}

// AssertNotNil fails the test if value is nil
func AssertNotNil(t TestingT, value interface{}, msg string) {
	t.Helper()
	if value == nil || reflect.ValueOf(value).IsNil() {
		t.Errorf("%s: expected non-nil value", msg)
	}
}

// AssertContains fails the test if the string doesn't contain the substring
func AssertContains(t TestingT, haystack, needle, msg string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("%s: expected string to contain %q, got: %s", msg, needle, haystack)
	}
}

// AssertNotContains fails the test if the string contains the substring
func AssertNotContains(t TestingT, haystack, needle, msg string) {
	t.Helper()
	if strings.Contains(haystack, needle) {
		t.Errorf("%s: expected string not to contain %q, got: %s", msg, needle, haystack)
	}
}

// AssertEmpty fails the test if the string is not empty
func AssertEmpty(t TestingT, value, msg string) {
	t.Helper()
	if value != "" {
		t.Errorf("%s: expected empty string but got: %s", msg, value)
	}
}

// AssertNotEmpty fails the test if the string is empty
func AssertNotEmpty(t TestingT, value, msg string) {
	t.Helper()
	if value == "" {
		t.Errorf("%s: expected non-empty string", msg)
	}
}

// AssertLen fails the test if the slice/map/string doesn't have the expected length
func AssertLen(t TestingT, value interface{}, expectedLen int, msg string) {
	t.Helper()
	v := reflect.ValueOf(value)
	actualLen := 0

	switch v.Kind() {
	case reflect.Slice, reflect.Map, reflect.String, reflect.Array:
		actualLen = v.Len()
	default:
		t.Fatalf("%s: cannot get length of type %T", msg, value)
	}

	if actualLen != expectedLen {
		t.Errorf("%s: expected length %d but got %d", msg, expectedLen, actualLen)
	}
}

// AssertStructEqual performs a deep comparison of two structs and reports differences
func AssertStructEqual(t TestingT, expected, actual interface{}) {
	t.Helper()

	if !reflect.DeepEqual(expected, actual) {
		expectedVal := reflect.ValueOf(expected)
		actualVal := reflect.ValueOf(actual)

		if expectedVal.Type() != actualVal.Type() {
			t.Fatalf("type mismatch: expected %T but got %T", expected, actual)
		}

		// For structs, show detailed field differences
		if expectedVal.Kind() == reflect.Struct {
			t.Errorf("structs are not equal:")
			for i := 0; i < expectedVal.NumField(); i++ {
				expectedField := expectedVal.Field(i)
				actualField := actualVal.Field(i)
				fieldName := expectedVal.Type().Field(i).Name

				if !reflect.DeepEqual(expectedField.Interface(), actualField.Interface()) {
					t.Errorf("  field %s: expected %v, got %v",
						fieldName,
						expectedField.Interface(),
						actualField.Interface())
				}
			}
		} else {
			t.Errorf("expected %v, got %v", expected, actual)
		}
	}
}

// AssertPanics fails the test if the function doesn't panic
func AssertPanics(t TestingT, fn func(), msg string) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("%s: expected panic but function completed normally", msg)
		}
	}()
	fn()
}

// AssertNoPanic fails the test if the function panics
func AssertNoPanic(t TestingT, fn func(), msg string) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("%s: unexpected panic: %v", msg, r)
		}
	}()
	fn()
}

// RequireNoError fails the test immediately if err is not nil
func RequireNoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", msg, err)
	}
}

// RequireNotNil fails the test immediately if value is nil
func RequireNotNil(t *testing.T, value interface{}, msg string) {
	t.Helper()
	if value == nil || reflect.ValueOf(value).IsNil() {
		t.Fatalf("%s: expected non-nil value", msg)
	}
}

// FailNow fails the test immediately with a message
func FailNow(t *testing.T, format string, args ...interface{}) {
	t.Helper()
	t.Fatalf(format, args...)
}

// AssertFileExists checks if a file exists at the given path
func AssertFileExists(t TestingT, _ string, msg string) {
	t.Helper()
	// This would require os package access, but keeping interface simple
	// Implementation can be added by tests that need it
	t.Errorf("%s: file existence check not implemented - use standard library", msg)
}

// FormatValue formats a value for display in error messages
func FormatValue(v interface{}) string {
	if v == nil {
		return "nil"
	}
	return fmt.Sprintf("%v (type: %T)", v, v)
}
