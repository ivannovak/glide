package errors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	err := New(TypeDocker, "test message")
	
	assert.Equal(t, TypeDocker, err.Type)
	assert.Equal(t, "test message", err.Message)
	assert.Equal(t, 1, err.Code) // Default exit code
	assert.Nil(t, err.Err)
	assert.Empty(t, err.Suggestions)
	assert.Nil(t, err.Context)
}

func TestNewWithOptions(t *testing.T) {
	underlying := fmt.Errorf("underlying error")
	
	err := New(TypeNetwork, "test message",
		WithError(underlying),
		WithExitCode(99),
		WithSuggestions("suggestion 1", "suggestion 2"),
		WithContext("key", "value"),
	)
	
	assert.Equal(t, TypeNetwork, err.Type)
	assert.Equal(t, "test message", err.Message)
	assert.Equal(t, 99, err.Code)
	assert.Equal(t, underlying, err.Err)
	assert.Equal(t, []string{"suggestion 1", "suggestion 2"}, err.Suggestions)
	assert.Equal(t, "value", err.Context["key"])
}

func TestNewDockerError(t *testing.T) {
	err := NewDockerError("docker daemon not running")
	
	assert.Equal(t, TypeDocker, err.Type)
	assert.Equal(t, "docker daemon not running", err.Message)
	assert.Equal(t, 125, err.Code) // Docker's standard exit code
}

func TestNewContainerError(t *testing.T) {
	err := NewContainerError("mycontainer", "container failed to start")
	
	assert.Equal(t, TypeContainer, err.Type)
	assert.Equal(t, "container failed to start", err.Message)
	assert.Equal(t, 125, err.Code)
	assert.Equal(t, "mycontainer", err.Context["container"])
}

func TestNewPermissionError(t *testing.T) {
	err := NewPermissionError("/tmp/test", "access denied")
	
	assert.Equal(t, TypePermission, err.Type)
	assert.Equal(t, "access denied", err.Message)
	assert.Equal(t, 126, err.Code) // Standard permission denied exit code
	assert.Equal(t, "/tmp/test", err.Context["path"])
	assert.True(t, len(err.Suggestions) > 0)
	assert.Contains(t, err.Suggestions[0], "ls -la /tmp/test")
}

func TestNewFileNotFoundError(t *testing.T) {
	err := NewFileNotFoundError("/missing/file.txt")
	
	assert.Equal(t, TypeFileNotFound, err.Type)
	assert.Equal(t, "file not found: /missing/file.txt", err.Message)
	assert.Equal(t, 127, err.Code)
	assert.Equal(t, "/missing/file.txt", err.Context["path"])
	assert.True(t, len(err.Suggestions) > 0)
	assert.Contains(t, err.Suggestions[0], "Check if the file exists")
}

func TestNewDependencyError(t *testing.T) {
	err := NewDependencyError("docker", "docker not installed")
	
	assert.Equal(t, TypeDependency, err.Type)
	assert.Equal(t, "docker not installed", err.Message)
	assert.Equal(t, 127, err.Code)
	assert.Equal(t, "docker", err.Context["dependency"])
}

func TestNewConfigError(t *testing.T) {
	err := NewConfigError("invalid configuration")
	
	assert.Equal(t, TypeConfig, err.Type)
	assert.Equal(t, "invalid configuration", err.Message)
	assert.Equal(t, 78, err.Code) // EX_CONFIG from sysexits.h
	assert.True(t, len(err.Suggestions) > 0)
	assert.Contains(t, err.Suggestions[0], "~/.glide.yml")
}

func TestNewNetworkError(t *testing.T) {
	err := NewNetworkError("connection failed")
	
	assert.Equal(t, TypeNetwork, err.Type)
	assert.Equal(t, "connection failed", err.Message)
	assert.Equal(t, 69, err.Code) // EX_UNAVAILABLE from sysexits.h
}

func TestNewDatabaseError(t *testing.T) {
	err := NewDatabaseError("cannot connect to database")
	
	assert.Equal(t, TypeDatabase, err.Type)
	assert.Equal(t, "cannot connect to database", err.Message)
	assert.Equal(t, 69, err.Code)
	assert.True(t, len(err.Suggestions) > 0)
	assert.Contains(t, err.Suggestions[0], "MySQL container")
}

func TestNewModeError(t *testing.T) {
	err := NewModeError("standard", "multi-worktree", "glid commit")
	
	assert.Equal(t, TypeMode, err.Type)
	assert.Contains(t, err.Message, "glid commit")
	assert.Contains(t, err.Message, "standard mode")
	assert.Equal(t, 65, err.Code) // EX_DATAERR from sysexits.h
	assert.Equal(t, "standard", err.Context["current_mode"])
	assert.Equal(t, "multi-worktree", err.Context["required_mode"])
	assert.Equal(t, "glid commit", err.Context["command"])
	assert.True(t, len(err.Suggestions) > 0)
}

func TestNewCommandError(t *testing.T) {
	err := NewCommandError("npm install", 1)
	
	assert.Equal(t, TypeCommand, err.Type)
	assert.Equal(t, "command failed: npm install", err.Message)
	assert.Equal(t, 1, err.Code)
	assert.Equal(t, "npm install", err.Context["command"])
}

func TestNewTimeoutError(t *testing.T) {
	err := NewTimeoutError("database migration")
	
	assert.Equal(t, TypeTimeout, err.Type)
	assert.Equal(t, "operation timed out: database migration", err.Message)
	assert.Equal(t, 124, err.Code) // Standard timeout exit code
	assert.Equal(t, "database migration", err.Context["operation"])
	assert.True(t, len(err.Suggestions) > 0)
}

func TestNewRuntimeError(t *testing.T) {
	err := NewRuntimeError("out of memory")
	
	assert.Equal(t, TypeRuntime, err.Type)
	assert.Equal(t, "out of memory", err.Message)
	assert.Equal(t, 71, err.Code) // EX_OSERR from sysexits.h
}

func TestWrapNilError(t *testing.T) {
	result := Wrap(nil, "wrapping nil")
	assert.Nil(t, result)
}

func TestWrapStandardError(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	wrapped := Wrap(originalErr, "wrapped message")
	
	require.NotNil(t, wrapped)
	assert.Equal(t, TypeUnknown, wrapped.Type)
	assert.Equal(t, "wrapped message", wrapped.Message)
	assert.Equal(t, originalErr, wrapped.Err)
}

func TestWrapGlideError(t *testing.T) {
	original := NewDockerError("docker failed")
	original.AddSuggestion("restart docker")
	original.AddContext("service", "mysql")
	
	wrapped := Wrap(original, "deployment failed")
	
	require.NotNil(t, wrapped)
	assert.Equal(t, TypeDocker, wrapped.Type) // Preserves type
	assert.Equal(t, "deployment failed", wrapped.Message)
	assert.Equal(t, original, wrapped.Err)
	assert.Equal(t, original.Suggestions, wrapped.Suggestions) // Preserves suggestions
	assert.Equal(t, original.Context, wrapped.Context) // Preserves context
	assert.Equal(t, original.Code, wrapped.Code) // Preserves exit code
}

func TestIsFunction(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		errType  ErrorType
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			errType:  TypeDocker,
			expected: false,
		},
		{
			name:     "standard error",
			err:      fmt.Errorf("standard error"),
			errType:  TypeDocker,
			expected: false,
		},
		{
			name:     "matching GlideError",
			err:      NewDockerError("docker error"),
			errType:  TypeDocker,
			expected: true,
		},
		{
			name:     "non-matching GlideError",
			err:      NewDockerError("docker error"),
			errType:  TypeNetwork,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Is(tt.err, tt.errType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGlideErrorError(t *testing.T) {
	tests := []struct {
		name     string
		err      *GlideError
		expected string
	}{
		{
			name: "error without underlying error",
			err: &GlideError{
				Message: "test message",
			},
			expected: "test message",
		},
		{
			name: "error with underlying error",
			err: &GlideError{
				Message: "wrapper message",
				Err:     fmt.Errorf("underlying error"),
			},
			expected: "wrapper message: underlying error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGlideErrorUnwrap(t *testing.T) {
	underlying := fmt.Errorf("underlying error")
	err := &GlideError{
		Message: "wrapper",
		Err:     underlying,
	}
	
	assert.Equal(t, underlying, err.Unwrap())
}

func TestGlideErrorIs(t *testing.T) {
	dockerErr1 := NewDockerError("error 1")
	dockerErr2 := NewDockerError("error 2")
	networkErr := NewNetworkError("network error")
	standardErr := fmt.Errorf("standard error")

	tests := []struct {
		name     string
		err      *GlideError
		target   error
		expected bool
	}{
		{
			name:     "same type GlideError",
			err:      dockerErr1,
			target:   dockerErr2,
			expected: true,
		},
		{
			name:     "different type GlideError",
			err:      dockerErr1,
			target:   networkErr,
			expected: false,
		},
		{
			name:     "standard error target",
			err:      dockerErr1,
			target:   standardErr,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Is(tt.target)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGlideErrorHasSuggestions(t *testing.T) {
	errWithSuggestions := NewPermissionError("/tmp/test", "access denied") // Has default suggestions
	errWithoutSuggestions := &GlideError{Message: "no suggestions"}
	
	assert.True(t, errWithSuggestions.HasSuggestions())
	assert.False(t, errWithoutSuggestions.HasSuggestions())
}

func TestGlideErrorGetContext(t *testing.T) {
	err := &GlideError{
		Context: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}
	
	value, ok := err.GetContext("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", value)
	
	_, ok = err.GetContext("nonexistent")
	assert.False(t, ok)
	
	// Test nil context
	errNoContext := &GlideError{}
	_, ok = errNoContext.GetContext("key")
	assert.False(t, ok)
}

func TestGlideErrorAddSuggestion(t *testing.T) {
	err := &GlideError{Message: "test"}
	
	result := err.AddSuggestion("suggestion 1")
	assert.Equal(t, err, result) // Should return same instance
	assert.Equal(t, []string{"suggestion 1"}, err.Suggestions)
	
	err.AddSuggestion("suggestion 2")
	assert.Equal(t, []string{"suggestion 1", "suggestion 2"}, err.Suggestions)
}

func TestGlideErrorAddContext(t *testing.T) {
	err := &GlideError{Message: "test"}
	
	result := err.AddContext("key1", "value1")
	assert.Equal(t, err, result) // Should return same instance
	require.NotNil(t, err.Context)
	assert.Equal(t, "value1", err.Context["key1"])
	
	err.AddContext("key2", "value2")
	assert.Equal(t, "value1", err.Context["key1"])
	assert.Equal(t, "value2", err.Context["key2"])
}

func TestGlideErrorWithCode(t *testing.T) {
	err := &GlideError{Message: "test", Code: 1}
	
	result := err.WithCode(99)
	assert.Equal(t, err, result) // Should return same instance
	assert.Equal(t, 99, err.Code)
}

func TestErrorOptions(t *testing.T) {
	underlying := fmt.Errorf("underlying")
	
	err := New(TypeNetwork, "test message",
		WithError(underlying),
		WithExitCode(42),
		WithSuggestions("suggestion 1", "suggestion 2"),
		WithContext("key1", "value1"),
		WithContext("key2", "value2"),
	)
	
	assert.Equal(t, underlying, err.Err)
	assert.Equal(t, 42, err.Code)
	assert.Equal(t, []string{"suggestion 1", "suggestion 2"}, err.Suggestions)
	assert.Equal(t, "value1", err.Context["key1"])
	assert.Equal(t, "value2", err.Context["key2"])
}

func TestCommonErrorMatches(t *testing.T) {
	commonErr := &CommonError{
		Pattern: "permission denied",
		Type:    TypePermission,
	}
	
	tests := []struct {
		name     string
		message  string
		expected bool
	}{
		{
			name:     "exact match",
			message:  "permission denied",
			expected: true,
		},
		{
			name:     "case insensitive match",
			message:  "PERMISSION DENIED",
			expected: true,
		},
		{
			name:     "contains pattern",
			message:  "error: permission denied for user",
			expected: true,
		},
		{
			name:     "no match",
			message:  "file not found",
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := commonErr.Matches(tt.message)
			assert.Equal(t, tt.expected, result)
		})
	}
}