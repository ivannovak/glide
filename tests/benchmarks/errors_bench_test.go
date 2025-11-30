package benchmarks_test

import (
	"errors"
	"fmt"
	"testing"

	glideerrors "github.com/ivannovak/glide/v3/pkg/errors"
)

// BenchmarkErrorNew benchmarks creating a basic GlideError
func BenchmarkErrorNew(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = glideerrors.New(glideerrors.TypeConfig, "test error message")
	}
}

// BenchmarkErrorNewWithOptions benchmarks creating GlideError with options
func BenchmarkErrorNewWithOptions(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = glideerrors.New(
			glideerrors.TypeConfig,
			"test error message",
			glideerrors.WithExitCode(1),
			glideerrors.WithContext("key1", "value1"),
			glideerrors.WithContext("key2", "value2"),
		)
	}
}

// BenchmarkErrorNewWithSuggestions benchmarks creating GlideError with suggestions
func BenchmarkErrorNewWithSuggestions(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = glideerrors.New(
			glideerrors.TypeConfig,
			"test error message",
			glideerrors.WithSuggestions(
				"Try running glide setup",
				"Check your configuration file",
				"Ensure you have the correct permissions",
			),
		)
	}
}

// BenchmarkErrorNewFileNotFound benchmarks creating file not found errors
func BenchmarkErrorNewFileNotFound(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = glideerrors.NewFileNotFoundError("/path/to/missing/file.txt")
	}
}

// BenchmarkErrorNewPermission benchmarks creating permission errors
func BenchmarkErrorNewPermission(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = glideerrors.NewPermissionError("/path/to/protected/file", "access denied")
	}
}

// BenchmarkErrorNewConfig benchmarks creating config errors
func BenchmarkErrorNewConfig(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = glideerrors.NewConfigError("invalid configuration format")
	}
}

// BenchmarkErrorWrap benchmarks wrapping errors
func BenchmarkErrorWrap(b *testing.B) {
	cause := errors.New("underlying cause")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = glideerrors.Wrap(cause, "wrapped error")
	}
}

// BenchmarkErrorWrapDeep benchmarks wrapping errors multiple levels deep
func BenchmarkErrorWrapDeep(b *testing.B) {
	cause := errors.New("root cause")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := glideerrors.Wrap(cause, "level 1")
		err = glideerrors.Wrap(err, "level 2")
		err = glideerrors.Wrap(err, "level 3")
		_ = err
	}
}

// BenchmarkErrorString benchmarks error string formatting
func BenchmarkErrorString(b *testing.B) {
	err := glideerrors.New(
		glideerrors.TypeConfig,
		"test error",
		glideerrors.WithContext("key", "value"),
		glideerrors.WithSuggestions("Try this", "Or that"),
	)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}

// BenchmarkErrorFormat benchmarks error Format method
func BenchmarkErrorFormat(b *testing.B) {
	err := glideerrors.New(
		glideerrors.TypeConfig,
		"test error",
		glideerrors.WithContext("key", "value"),
		glideerrors.WithSuggestions("Try this", "Or that"),
	)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("%+v", err)
	}
}

// BenchmarkErrorIs benchmarks errors.Is compatibility
func BenchmarkErrorIs(b *testing.B) {
	target := glideerrors.New(glideerrors.TypeConfig, "target")
	err := glideerrors.Wrap(target, "wrapped")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = errors.Is(err, target)
	}
}

// BenchmarkErrorTypeCheck benchmarks error type checking
func BenchmarkErrorTypeCheck(b *testing.B) {
	err := glideerrors.NewConfigError("test config error")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = glideerrors.Is(err, glideerrors.TypeConfig)
	}
}

// BenchmarkErrorContextAccess benchmarks accessing error context
func BenchmarkErrorContextAccess(b *testing.B) {
	err := glideerrors.New(
		glideerrors.TypeConfig,
		"test error",
		glideerrors.WithContext("key1", "value1"),
		glideerrors.WithContext("key2", "value2"),
		glideerrors.WithContext("key3", "value3"),
	)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = err.Context
	}
}

// BenchmarkErrorAllocation measures allocations for error creation
func BenchmarkErrorAllocation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := glideerrors.New(
			glideerrors.TypeConfig,
			"test error",
			glideerrors.WithExitCode(1),
			glideerrors.WithContext("path", "/some/path"),
			glideerrors.WithSuggestions("suggestion 1", "suggestion 2"),
		)
		_ = err
	}
}
