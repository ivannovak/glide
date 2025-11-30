package cli

import (
	"bytes"
	stderrors "errors"
	"fmt"
	"strings"
	"testing"

	"github.com/ivannovak/glide/v2/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestErrorFormatting tests error display formatting
func TestErrorFormatting(t *testing.T) {
	t.Run("user error format", func(t *testing.T) {
		err := errors.New(errors.TypeCommand, "command not found")

		handler := errors.DefaultHandler()
		handler.NoColor = true
		buf := &bytes.Buffer{}
		handler.Writer = buf

		code := handler.Handle(err)
		assert.Equal(t, 1, code)

		output := buf.String()
		assert.Contains(t, output, "Command Error")
		assert.Contains(t, output, "command not found")
	})

	t.Run("system error format", func(t *testing.T) {
		err := errors.New(errors.TypeDocker, "docker daemon not running",
			errors.WithExitCode(2))

		handler := errors.DefaultHandler()
		handler.NoColor = true
		buf := &bytes.Buffer{}
		handler.Writer = buf

		code := handler.Handle(err)
		assert.Equal(t, 2, code)

		output := buf.String()
		assert.Contains(t, output, "Docker Error")
		assert.Contains(t, output, "docker daemon not running")
	})

	t.Run("plugin error format", func(t *testing.T) {
		err := errors.New(errors.TypeDependency, "plugin dependency missing")

		handler := errors.DefaultHandler()
		handler.NoColor = true
		buf := &bytes.Buffer{}
		handler.Writer = buf

		code := handler.Handle(err)
		assert.Equal(t, 1, code)

		output := buf.String()
		assert.Contains(t, output, "Dependency Error")
		assert.Contains(t, output, "plugin dependency missing")
	})

	t.Run("error with wrapped error", func(t *testing.T) {
		baseErr := fmt.Errorf("original error")
		err := errors.New(errors.TypeConfig, "configuration failed",
			errors.WithError(baseErr))

		handler := errors.DefaultHandler()
		handler.NoColor = true
		handler.Verbose = true
		buf := &bytes.Buffer{}
		handler.Writer = buf

		code := handler.Handle(err)
		assert.Equal(t, 1, code)

		output := buf.String()
		assert.Contains(t, output, "Configuration Error")
		assert.Contains(t, output, "configuration failed")
		assert.Contains(t, output, "Underlying error")
		assert.Contains(t, output, "original error")
	})

	t.Run("generic error format", func(t *testing.T) {
		err := fmt.Errorf("generic error message")

		handler := errors.DefaultHandler()
		handler.NoColor = true
		buf := &bytes.Buffer{}
		handler.Writer = buf

		code := handler.Handle(err)
		assert.Equal(t, 1, code)

		output := buf.String()
		assert.Contains(t, output, "Error")
		assert.Contains(t, output, "generic error message")
	})
}

// TestErrorSuggestions tests error suggestion generation
func TestErrorSuggestions(t *testing.T) {
	t.Run("command not found suggestions", func(t *testing.T) {
		err := errors.New(errors.TypeCommand, "command not found: artisa",
			errors.WithSuggestions("Did you mean 'artisan'?"))

		handler := errors.DefaultHandler()
		handler.NoColor = true
		buf := &bytes.Buffer{}
		handler.Writer = buf

		handler.Handle(err)

		output := buf.String()
		assert.Contains(t, output, "Possible solutions:")
		assert.Contains(t, output, "Did you mean 'artisan'?")
	})

	t.Run("Docker daemon suggestions", func(t *testing.T) {
		// Use AnalyzeError to get smart suggestions
		baseErr := fmt.Errorf("cannot connect to the docker daemon")
		err := errors.AnalyzeError(baseErr)

		handler := errors.DefaultHandler()
		handler.NoColor = true
		buf := &bytes.Buffer{}
		handler.Writer = buf

		handler.Handle(err)

		output := buf.String()
		assert.Contains(t, output, "Possible solutions:")
		// Should suggest starting Docker
		assert.True(t,
			strings.Contains(output, "Docker") ||
				strings.Contains(output, "docker"),
			"should contain Docker-related suggestion")
	})

	t.Run("database connection suggestions", func(t *testing.T) {
		baseErr := fmt.Errorf("SQLSTATE[HY000] [2002] Connection refused")
		err := errors.AnalyzeError(baseErr)

		handler := errors.DefaultHandler()
		handler.NoColor = true
		buf := &bytes.Buffer{}
		handler.Writer = buf

		handler.Handle(err)

		output := buf.String()
		assert.Contains(t, output, "Possible solutions:")
	})

	t.Run("permission denied suggestions", func(t *testing.T) {
		baseErr := fmt.Errorf("permission denied: /var/log/app.log")
		err := errors.AnalyzeError(baseErr)

		handler := errors.DefaultHandler()
		handler.NoColor = true
		buf := &bytes.Buffer{}
		handler.Writer = buf

		handler.Handle(err)

		output := buf.String()
		assert.Contains(t, output, "Possible solutions:")
	})

	t.Run("file not found suggestions", func(t *testing.T) {
		baseErr := fmt.Errorf("no such file or directory: .env")
		err := errors.AnalyzeError(baseErr)

		handler := errors.DefaultHandler()
		handler.NoColor = true
		buf := &bytes.Buffer{}
		handler.Writer = buf

		handler.Handle(err)

		output := buf.String()
		assert.Contains(t, output, "Possible solutions:")
	})

	t.Run("multiple suggestions", func(t *testing.T) {
		err := errors.New(errors.TypeDocker, "Docker error",
			errors.WithSuggestions(
				"Start Docker Desktop",
				"Check: docker ps",
				"Restart containers: glide up",
			))

		handler := errors.DefaultHandler()
		handler.NoColor = true
		buf := &bytes.Buffer{}
		handler.Writer = buf

		handler.Handle(err)

		output := buf.String()
		assert.Contains(t, output, "Possible solutions:")
		assert.Contains(t, output, "Start Docker Desktop")
		assert.Contains(t, output, "Check: docker ps")
		assert.Contains(t, output, "Restart containers: glide up")
	})
}

// TestErrorExitCodes tests exit code handling
func TestErrorExitCodes(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
	}{
		{
			name:     "success (no error)",
			err:      nil,
			wantCode: 0,
		},
		{
			name:     "user error (default)",
			err:      errors.New(errors.TypeCommand, "user error"),
			wantCode: 1,
		},
		{
			name:     "system error",
			err:      errors.New(errors.TypeDocker, "system error", errors.WithExitCode(2)),
			wantCode: 2,
		},
		{
			name:     "panic/crash",
			err:      errors.New(errors.TypeRuntime, "panic occurred", errors.WithExitCode(3)),
			wantCode: 3,
		},
		{
			name:     "custom exit code",
			err:      errors.New(errors.TypeCommand, "custom error", errors.WithExitCode(42)),
			wantCode: 42,
		},
		{
			name:     "generic error defaults to 1",
			err:      fmt.Errorf("generic error"),
			wantCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := errors.DefaultHandler()
			buf := &bytes.Buffer{}
			handler.Writer = buf

			code := handler.Handle(tt.err)
			assert.Equal(t, tt.wantCode, code)
		})
	}
}

// TestErrorTypes tests different error type handling
func TestErrorTypes(t *testing.T) {
	errorTypes := []struct {
		errType      errors.ErrorType
		expectedText string
	}{
		{errors.TypeDocker, "Docker Error"},
		{errors.TypeContainer, "Container Error"},
		{errors.TypePermission, "Permission Error"},
		{errors.TypeFileNotFound, "File Not Found"},
		{errors.TypeDependency, "Dependency Error"},
		{errors.TypeMissing, "Missing Resource"},
		{errors.TypeConfig, "Configuration Error"},
		{errors.TypeNetwork, "Network Error"},
		{errors.TypeConnection, "Connection Error"},
		{errors.TypeDatabase, "Database Error"},
		{errors.TypeMode, "Mode Error"},
		{errors.TypeWrongMode, "Wrong Mode"},
		{errors.TypeTimeout, "Timeout"},
		{errors.TypeCommand, "Command Error"},
	}

	for _, tc := range errorTypes {
		t.Run(string(tc.errType), func(t *testing.T) {
			err := errors.New(tc.errType, "test error")

			handler := errors.DefaultHandler()
			handler.NoColor = true
			buf := &bytes.Buffer{}
			handler.Writer = buf

			handler.Handle(err)

			output := buf.String()
			assert.Contains(t, output, tc.expectedText)
		})
	}
}

// TestErrorContext tests error context display
func TestErrorContext(t *testing.T) {
	t.Run("context shown in verbose mode", func(t *testing.T) {
		err := errors.New(errors.TypeDocker, "docker error",
			errors.WithContext("container", "php"),
			errors.WithContext("status", "exited"))

		handler := errors.DefaultHandler()
		handler.NoColor = true
		handler.Verbose = true
		buf := &bytes.Buffer{}
		handler.Writer = buf

		handler.Handle(err)

		output := buf.String()
		assert.Contains(t, output, "Context:")
		assert.Contains(t, output, "container")
		assert.Contains(t, output, "php")
		assert.Contains(t, output, "status")
		assert.Contains(t, output, "exited")
	})

	t.Run("context hidden in non-verbose mode", func(t *testing.T) {
		err := errors.New(errors.TypeDocker, "docker error",
			errors.WithContext("container", "php"))

		handler := errors.DefaultHandler()
		handler.NoColor = true
		handler.Verbose = false
		buf := &bytes.Buffer{}
		handler.Writer = buf

		handler.Handle(err)

		output := buf.String()
		assert.NotContains(t, output, "Context:")
	})

	t.Run("context-based suggestions", func(t *testing.T) {
		baseErr := fmt.Errorf("connection refused")
		context := map[string]string{
			"container": "php",
		}

		err := errors.EnhanceError(baseErr, context)
		require.NotNil(t, err)

		// Should have context-specific suggestions for PHP container
		assert.True(t, err.HasSuggestions())

		handler := errors.DefaultHandler()
		handler.NoColor = true
		handler.Verbose = true
		buf := &bytes.Buffer{}
		handler.Writer = buf

		handler.Handle(err)

		output := buf.String()
		assert.Contains(t, output, "Possible solutions:")
	})
}

// TestErrorWrapping tests error wrapping and unwrapping
func TestErrorWrapping(t *testing.T) {
	t.Run("wrap standard error", func(t *testing.T) {
		baseErr := fmt.Errorf("base error")
		err := errors.New(errors.TypeCommand, "wrapped error",
			errors.WithError(baseErr))

		// Test Unwrap
		unwrapped := stderrors.Unwrap(err)
		assert.Equal(t, baseErr, unwrapped)
	})

	t.Run("error string includes wrapped error", func(t *testing.T) {
		baseErr := fmt.Errorf("base error")
		err := errors.New(errors.TypeCommand, "wrapped error",
			errors.WithError(baseErr))

		errStr := err.Error()
		assert.Contains(t, errStr, "wrapped error")
		assert.Contains(t, errStr, "base error")
	})

	t.Run("chained error wrapping", func(t *testing.T) {
		err1 := fmt.Errorf("level 1")
		err2 := errors.New(errors.TypeCommand, "level 2", errors.WithError(err1))
		err3 := errors.New(errors.TypeDocker, "level 3", errors.WithError(err2))

		// Should be able to unwrap all the way down
		unwrap1 := stderrors.Unwrap(err3)
		assert.Equal(t, err2, unwrap1)

		unwrap2 := stderrors.Unwrap(unwrap1)
		assert.Equal(t, err1, unwrap2)
	})
}

// TestErrorBuilders tests error creation with various options
func TestErrorBuilders(t *testing.T) {
	t.Run("create with all options", func(t *testing.T) {
		baseErr := fmt.Errorf("base error")
		glideErr := errors.New(errors.TypeCommand, "test error",
			errors.WithError(baseErr),
			errors.WithSuggestions("suggestion 1", "suggestion 2"),
			errors.WithContext("key", "value"),
			errors.WithExitCode(42))

		assert.Equal(t, errors.TypeCommand, glideErr.Type)
		assert.Equal(t, "test error", glideErr.Message)
		assert.Equal(t, baseErr, glideErr.Err)
		assert.Len(t, glideErr.Suggestions, 2)
		assert.Equal(t, "suggestion 1", glideErr.Suggestions[0])
		assert.Equal(t, "suggestion 2", glideErr.Suggestions[1])
		assert.Equal(t, 42, glideErr.Code)

		val, ok := glideErr.GetContext("key")
		assert.True(t, ok)
		assert.Equal(t, "value", val)
	})

	t.Run("fluent error building", func(t *testing.T) {
		glideErr := errors.New(errors.TypeCommand, "test error")

		glideErr.
			AddSuggestion("suggestion 1").
			AddSuggestion("suggestion 2").
			AddContext("key1", "value1").
			AddContext("key2", "value2").
			WithCode(10)

		assert.Len(t, glideErr.Suggestions, 2)
		assert.Len(t, glideErr.Context, 2)
		assert.Equal(t, 10, glideErr.Code)
	})
}

// TestSuggestionEngine tests the suggestion engine
func TestSuggestionEngine(t *testing.T) {
	t.Run("analyze docker daemon error", func(t *testing.T) {
		err := fmt.Errorf("cannot connect to the docker daemon at unix:///var/run/docker.sock")
		analyzed := errors.AnalyzeError(err)

		require.NotNil(t, analyzed)
		assert.Equal(t, errors.TypeDocker, analyzed.Type)
		assert.True(t, analyzed.HasSuggestions())
	})

	t.Run("analyze database connection error", func(t *testing.T) {
		err := fmt.Errorf("SQLSTATE[HY000] [2002] Connection refused")
		analyzed := errors.AnalyzeError(err)

		require.NotNil(t, analyzed)
		assert.Equal(t, errors.TypeDatabase, analyzed.Type)
		assert.True(t, analyzed.HasSuggestions())
	})

	t.Run("analyze permission error", func(t *testing.T) {
		err := fmt.Errorf("permission denied: /var/log/app.log")
		analyzed := errors.AnalyzeError(err)

		require.NotNil(t, analyzed)
		assert.Equal(t, errors.TypePermission, analyzed.Type)
		assert.True(t, analyzed.HasSuggestions())
	})

	t.Run("analyze file not found error", func(t *testing.T) {
		err := fmt.Errorf("no such file or directory: .env")
		analyzed := errors.AnalyzeError(err)

		require.NotNil(t, analyzed)
		assert.Equal(t, errors.TypeFileNotFound, analyzed.Type)
		assert.True(t, analyzed.HasSuggestions())
	})

	t.Run("analyze already GlideError", func(t *testing.T) {
		original := errors.New(errors.TypeCommand, "test",
			errors.WithSuggestions("original suggestion"))

		analyzed := errors.AnalyzeError(original)

		require.NotNil(t, analyzed)
		assert.True(t, analyzed.HasSuggestions())
		assert.Contains(t, analyzed.Suggestions, "original suggestion")
	})
}

// TestErrorColorOutput tests color handling
func TestErrorColorOutput(t *testing.T) {
	t.Run("color output enabled", func(t *testing.T) {
		err := errors.New(errors.TypeCommand, "test error")

		handler := errors.DefaultHandler()
		handler.NoColor = false // Enable color
		buf := &bytes.Buffer{}
		handler.Writer = buf

		handler.Handle(err)

		output := buf.String()
		// Output should contain the message
		assert.Contains(t, output, "test error")
	})

	t.Run("color output disabled", func(t *testing.T) {
		err := errors.New(errors.TypeCommand, "test error")

		handler := errors.DefaultHandler()
		handler.NoColor = true // Disable color
		buf := &bytes.Buffer{}
		handler.Writer = buf

		handler.Handle(err)

		output := buf.String()
		assert.Contains(t, output, "Command Error")
		assert.Contains(t, output, "test error")
	})
}

// TestErrorIcons tests error icon display
func TestErrorIcons(t *testing.T) {
	errorIcons := []struct {
		errType      errors.ErrorType
		expectedIcon string
	}{
		{errors.TypeDocker, "🐳"},
		{errors.TypePermission, "🔒"},
		{errors.TypeFileNotFound, "📁"},
		{errors.TypeDependency, "📦"},
		{errors.TypeConfig, "⚙️"},
		{errors.TypeNetwork, "🌐"},
		{errors.TypeDatabase, "🗄️"},
		{errors.TypeMode, "🔄"},
		{errors.TypeTimeout, "⏱️"},
		{errors.TypeCommand, "💻"},
	}

	for _, tc := range errorIcons {
		t.Run(string(tc.errType), func(t *testing.T) {
			err := errors.New(tc.errType, "test error")

			handler := errors.DefaultHandler()
			handler.NoColor = true
			buf := &bytes.Buffer{}
			handler.Writer = buf

			handler.Handle(err)

			output := buf.String()
			assert.Contains(t, output, tc.expectedIcon)
		})
	}
}
