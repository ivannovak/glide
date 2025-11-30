package errors

import (
	"fmt"
)

// New creates a new GlideError with the given type and message
func New(errType ErrorType, message string, opts ...ErrorOption) *GlideError {
	e := &GlideError{
		Type:    errType,
		Message: message,
		Code:    1, // Default exit code
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// NewDockerError creates a Docker-related error
func NewDockerError(message string, opts ...ErrorOption) *GlideError {
	defaultOpts := []ErrorOption{
		WithExitCode(125), // Docker's standard exit code for daemon errors
	}
	opts = append(defaultOpts, opts...)
	return New(TypeDocker, message, opts...)
}

// NewContainerError creates a container-specific error
func NewContainerError(container, message string, opts ...ErrorOption) *GlideError {
	defaultOpts := []ErrorOption{
		WithContext("container", container),
		WithExitCode(125),
	}
	opts = append(defaultOpts, opts...)
	return New(TypeContainer, message, opts...)
}

// NewPermissionError creates a permission-related error
func NewPermissionError(path, message string, opts ...ErrorOption) *GlideError {
	defaultOpts := []ErrorOption{
		WithContext("path", path),
		WithExitCode(126), // Standard permission denied exit code
		WithSuggestions(
			fmt.Sprintf("Check permissions: ls -la %s", path),
			fmt.Sprintf("Fix permissions: chmod 755 %s", path),
			"Run with sudo if necessary",
		),
	}
	opts = append(defaultOpts, opts...)
	return New(TypePermission, message, opts...)
}

// NewFileNotFoundError creates a file not found error
func NewFileNotFoundError(path string, opts ...ErrorOption) *GlideError {
	defaultOpts := []ErrorOption{
		WithContext("path", path),
		WithExitCode(127),
		WithSuggestions(
			"Check if the file exists",
			"Verify the path is correct",
			"Ensure you're in the right directory",
		),
	}
	opts = append(defaultOpts, opts...)
	return New(TypeFileNotFound, fmt.Sprintf("file not found: %s", path), opts...)
}

// NewDependencyError creates a missing dependency error
func NewDependencyError(dependency, message string, opts ...ErrorOption) *GlideError {
	defaultOpts := []ErrorOption{
		WithContext("dependency", dependency),
		WithExitCode(127),
	}
	opts = append(defaultOpts, opts...)
	return New(TypeDependency, message, opts...)
}

// NewConfigError creates a configuration error
func NewConfigError(message string, opts ...ErrorOption) *GlideError {
	defaultOpts := []ErrorOption{
		WithExitCode(78), // EX_CONFIG from sysexits.h
		WithSuggestions(
			"Check ~/.glide.yml configuration",
			"Run: glideconfig list",
			"Run: glidesetup to reconfigure",
		),
	}
	opts = append(defaultOpts, opts...)
	return New(TypeConfig, message, opts...)
}

// NewNetworkError creates a network-related error
func NewNetworkError(message string, opts ...ErrorOption) *GlideError {
	defaultOpts := []ErrorOption{
		WithExitCode(69), // EX_UNAVAILABLE from sysexits.h
	}
	opts = append(defaultOpts, opts...)
	return New(TypeNetwork, message, opts...)
}

// NewDatabaseError creates a database connection error
func NewDatabaseError(message string, opts ...ErrorOption) *GlideError {
	defaultOpts := []ErrorOption{
		WithExitCode(69),
		WithSuggestions(
			"Check if MySQL container is running: glidestatus",
			"Verify database credentials in .env file",
			"Ensure DB_HOST is set to 'mysql' for Docker",
			"Run: glidedocker up -d mysql",
		),
	}
	opts = append(defaultOpts, opts...)
	return New(TypeDatabase, message, opts...)
}

// NewModeError creates a development mode error
func NewModeError(currentMode, requiredMode, command string, opts ...ErrorOption) *GlideError {
	defaultOpts := []ErrorOption{
		WithContext("current_mode", currentMode),
		WithContext("required_mode", requiredMode),
		WithContext("command", command),
		WithExitCode(65), // EX_DATAERR from sysexits.h
		WithSuggestions(
			fmt.Sprintf("This command requires %s mode", requiredMode),
			"Run: glidesetup to change development mode",
			fmt.Sprintf("Current mode: %s", currentMode),
		),
	}
	opts = append(defaultOpts, opts...)
	return New(TypeMode, fmt.Sprintf("command '%s' is not available in %s mode", command, currentMode), opts...)
}

// NewCommandError creates a generic command error
func NewCommandError(command string, exitCode int, opts ...ErrorOption) *GlideError {
	defaultOpts := []ErrorOption{
		WithContext("command", command),
		WithExitCode(exitCode),
	}
	opts = append(defaultOpts, opts...)
	return New(TypeCommand, fmt.Sprintf("command failed: %s", command), opts...)
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(operation string, opts ...ErrorOption) *GlideError {
	defaultOpts := []ErrorOption{
		WithContext("operation", operation),
		WithExitCode(124), // Standard timeout exit code
		WithSuggestions(
			"Try running the command again",
			"Check if the system is under heavy load",
			"Increase timeout if possible",
		),
	}
	opts = append(defaultOpts, opts...)
	return New(TypeTimeout, fmt.Sprintf("operation timed out: %s", operation), opts...)
}

// NewRuntimeError creates a runtime error (memory, resource, etc)
func NewRuntimeError(message string, opts ...ErrorOption) *GlideError {
	defaultOpts := []ErrorOption{
		WithExitCode(71), // EX_OSERR from sysexits.h
	}
	opts = append(defaultOpts, opts...)
	return New(TypeRuntime, message, opts...)
}

// Wrap wraps an existing error with additional context
func Wrap(err error, message string, opts ...ErrorOption) *GlideError {
	if err == nil {
		return nil
	}

	// If it's already a GlideError, preserve its properties
	if glideErr, ok := err.(*GlideError); ok {
		// Create a new error with the wrapped message
		wrapped := &GlideError{
			Type:        glideErr.Type,
			Message:     message,
			Err:         glideErr,
			Suggestions: glideErr.Suggestions,
			Context:     glideErr.Context,
			Code:        glideErr.Code,
		}

		// Apply any new options
		for _, opt := range opts {
			opt(wrapped)
		}

		return wrapped
	}

	// Create a new GlideError wrapping the original
	return New(TypeUnknown, message, append(opts, WithError(err))...)
}

// Is checks if an error is of a specific type
func Is(err error, errType ErrorType) bool {
	if err == nil {
		return false
	}

	glideErr, ok := err.(*GlideError)
	if !ok {
		return false
	}

	return glideErr.Type == errType
}

// NewUserError creates an error caused by user input or configuration.
// This is a convenience wrapper for common user-facing errors.
//
// Example:
//
//	return errors.NewUserError(
//	    "plugin name is required",
//	    "Add a 'name' field to your plugin configuration",
//	)
func NewUserError(message, suggestion string) *GlideError {
	return New(TypeInvalid, message,
		WithSuggestions(suggestion),
		WithExitCode(64), // EX_USAGE from sysexits.h
	)
}

// NewSystemError creates an error for internal/infrastructure failures
// beyond user control.
//
// Example:
//
//	if err := initializeDatabase(); err != nil {
//	    return errors.NewSystemError("failed to initialize database", err)
//	}
func NewSystemError(message string, cause error) *GlideError {
	return New(TypeRuntime, message,
		WithError(cause),
		WithExitCode(71), // EX_OSERR from sysexits.h
	)
}

// NewPluginError creates an error for plugin-specific operations.
//
// Example:
//
//	if err := plugin.Execute(cmd); err != nil {
//	    return errors.NewPluginError(plugin.Name(), "command execution failed", err)
//	}
func NewPluginError(pluginName, message string, cause error) *GlideError {
	opts := []ErrorOption{
		WithContext("plugin", pluginName),
		WithExitCode(1),
	}
	if cause != nil {
		opts = append(opts, WithError(cause))
	}
	return New(TypeCommand, fmt.Sprintf("plugin '%s': %s", pluginName, message), opts...)
}

// WithSuggestion is a convenience function to add a suggestion to any error.
//
// Example:
//
//	if err := connectToServer(addr); err != nil {
//	    return errors.WithSuggestion(
//	        fmt.Errorf("failed to connect to %s: %w", addr, err),
//	        "Check that the server is running and the address is correct",
//	    )
//	}
func WithSuggestion(err error, suggestion string) *GlideError {
	if err == nil {
		return nil
	}

	// If it's already a GlideError, add suggestion
	if glideErr, ok := err.(*GlideError); ok {
		return glideErr.AddSuggestion(suggestion)
	}

	// Otherwise wrap it
	return Wrap(err, err.Error(), WithSuggestions(suggestion))
}
