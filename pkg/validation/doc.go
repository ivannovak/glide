// Package validation provides security validation functions for user input.
//
// This package protects against common security vulnerabilities including
// path traversal attacks, symlink attacks, and command injection. Use these
// functions to validate all user-provided paths and inputs before use.
//
// # Path Validation
//
// Validate file paths to prevent directory traversal:
//
//	opts := validation.PathValidationOptions{
//	    BaseDir:        "/app/data",
//	    AllowAbsolute:  false,
//	    FollowSymlinks: true,
//	    RequireExists:  true,
//	}
//
//	safePath, err := validation.ValidatePath(userInput, opts)
//	if err != nil {
//	    // Handle validation failure
//	    return err
//	}
//	// safePath is guaranteed to be within BaseDir
//
// # Security Checks
//
// The package detects and prevents:
//   - Path traversal attempts (../)
//   - Absolute paths when relative expected
//   - Symlink attacks pointing outside base directory
//   - Null bytes in paths (security bypass attempts)
//   - Paths outside the allowed base directory
//
// # Error Types
//
// Specific error types for different validation failures:
//
//	if errors.Is(err, validation.ErrPathTraversal) {
//	    // Path contained ../ sequences
//	}
//	if errors.Is(err, validation.ErrSymlinkTraversal) {
//	    // Symlink points outside base directory
//	}
//	if errors.Is(err, validation.ErrAbsolutePath) {
//	    // Absolute path when relative required
//	}
//
// # Best Practices
//
// Always validate paths before:
//   - Reading user-specified files
//   - Writing to user-specified locations
//   - Executing commands with user-provided paths
//   - Including files in responses
//
// Example secure file read:
//
//	func ReadUserFile(userPath, baseDir string) ([]byte, error) {
//	    safePath, err := validation.ValidatePath(userPath, validation.PathValidationOptions{
//	        BaseDir:        baseDir,
//	        FollowSymlinks: true,
//	        RequireExists:  true,
//	    })
//	    if err != nil {
//	        return nil, fmt.Errorf("invalid path: %w", err)
//	    }
//	    return os.ReadFile(safePath)
//	}
package validation
