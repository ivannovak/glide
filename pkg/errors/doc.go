// Package errors provides structured error handling for the Glide CLI.
//
// This package defines error types, constructors, and utilities for creating
// user-friendly, actionable error messages. All errors include context,
// exit codes, and optional suggestions for resolution.
//
// # Error Types
//
// Errors are categorized by type for consistent handling:
//   - TypeConfig: Configuration errors
//   - TypeDocker: Docker daemon and container errors
//   - TypeContainer: Container-specific errors
//   - TypePermission: File and directory permission errors
//   - TypeFileNotFound: Missing file errors
//   - TypeDependency: Missing dependency errors
//   - TypePlugin: Plugin-related errors
//   - TypeValidation: Input validation errors
//
// # Creating Errors
//
// Use typed constructors for common error scenarios:
//
//	// Docker daemon not running
//	err := errors.NewDockerError("Docker daemon is not running",
//	    errors.WithSuggestions("Start Docker Desktop", "Run: docker ps"))
//
//	// File not found with context
//	err := errors.NewFileNotFoundError("/path/to/config.yml")
//
//	// Permission error with path context
//	err := errors.NewPermissionError("/etc/config", "cannot read file")
//
// # Error Options
//
// Customize errors with functional options:
//
//	err := errors.New(errors.TypeValidation, "invalid input",
//	    errors.WithExitCode(2),
//	    errors.WithContext("field", "username"),
//	    errors.WithCause(originalErr),
//	    errors.WithSuggestions(
//	        "Check the input format",
//	        "See documentation for valid values",
//	    ))
//
// # Error Handling
//
// Use the Handler for consistent error display:
//
//	handler := errors.NewHandler(output.NewManager(...))
//	exitCode := handler.Handle(err)
//	os.Exit(exitCode)
//
// # Exit Codes
//
// Standard exit codes are used for different error types:
//   - 1: General errors
//   - 125: Docker errors
//   - 126: Permission errors
//   - 127: File not found / dependency errors
package errors
