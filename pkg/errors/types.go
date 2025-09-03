package errors

import (
	"fmt"
	"strings"
)

// ErrorType represents the category of error
type ErrorType string

const (
	// Docker related errors
	TypeDocker    ErrorType = "docker"
	TypeContainer ErrorType = "container"

	// File system errors
	TypePermission   ErrorType = "permission"
	TypeFileNotFound ErrorType = "file_not_found"

	// Dependency errors
	TypeDependency ErrorType = "dependency"
	TypeMissing    ErrorType = "missing"

	// Configuration errors
	TypeConfig  ErrorType = "configuration"
	TypeInvalid ErrorType = "invalid"

	// Network errors
	TypeNetwork    ErrorType = "network"
	TypeConnection ErrorType = "connection"

	// Mode errors
	TypeMode      ErrorType = "mode"
	TypeWrongMode ErrorType = "wrong_mode"

	// Database errors
	TypeDatabase ErrorType = "database"

	// Generic errors
	TypeCommand ErrorType = "command"
	TypeTimeout ErrorType = "timeout"
	TypeRuntime ErrorType = "runtime"
	TypeUnknown ErrorType = "unknown"
)

// GlideError is the standard error type for the Glide CLI
type GlideError struct {
	Type        ErrorType
	Message     string
	Err         error             // Underlying error
	Suggestions []string          // Helpful suggestions
	Context     map[string]string // Additional context
	Code        int               // Exit code
}

// Error implements the error interface
func (e *GlideError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *GlideError) Unwrap() error {
	return e.Err
}

// Is checks if the error is of a specific type
func (e *GlideError) Is(target error) bool {
	t, ok := target.(*GlideError)
	if !ok {
		return false
	}
	return e.Type == t.Type
}

// HasSuggestions returns true if the error has suggestions
func (e *GlideError) HasSuggestions() bool {
	return len(e.Suggestions) > 0
}

// GetContext returns a context value
func (e *GlideError) GetContext(key string) (string, bool) {
	if e.Context == nil {
		return "", false
	}
	val, ok := e.Context[key]
	return val, ok
}

// AddSuggestion adds a suggestion to the error
func (e *GlideError) AddSuggestion(suggestion string) *GlideError {
	e.Suggestions = append(e.Suggestions, suggestion)
	return e
}

// AddContext adds context to the error
func (e *GlideError) AddContext(key, value string) *GlideError {
	if e.Context == nil {
		e.Context = make(map[string]string)
	}
	e.Context[key] = value
	return e
}

// WithCode sets the exit code for the error
func (e *GlideError) WithCode(code int) *GlideError {
	e.Code = code
	return e
}

// ErrorOption is a functional option for creating errors
type ErrorOption func(*GlideError)

// WithError wraps an underlying error
func WithError(err error) ErrorOption {
	return func(e *GlideError) {
		e.Err = err
	}
}

// WithSuggestions adds suggestions to the error
func WithSuggestions(suggestions ...string) ErrorOption {
	return func(e *GlideError) {
		e.Suggestions = append(e.Suggestions, suggestions...)
	}
}

// WithContext adds context to the error
func WithContext(key, value string) ErrorOption {
	return func(e *GlideError) {
		if e.Context == nil {
			e.Context = make(map[string]string)
		}
		e.Context[key] = value
	}
}

// WithExitCode sets the exit code
func WithExitCode(code int) ErrorOption {
	return func(e *GlideError) {
		e.Code = code
	}
}

// CommonError represents a common error pattern
type CommonError struct {
	Pattern     string    // Error message pattern to match
	Type        ErrorType // Error type to assign
	Suggestions []string  // Default suggestions
}

// Matches checks if an error message matches this pattern
func (ce *CommonError) Matches(message string) bool {
	return strings.Contains(strings.ToLower(message), strings.ToLower(ce.Pattern))
}
