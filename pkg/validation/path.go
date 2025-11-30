// Package validation provides security validation functions for user input
package validation

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	// ErrPathTraversal is returned when a path attempts directory traversal
	ErrPathTraversal = errors.New("path traversal detected")

	// ErrAbsolutePath is returned when an absolute path is provided where relative is expected
	ErrAbsolutePath = errors.New("absolute paths are not allowed")

	// ErrSymlinkTraversal is returned when a symlink points outside the base directory
	ErrSymlinkTraversal = errors.New("symlink traversal detected")

	// ErrInvalidPath is returned for malformed or invalid paths
	ErrInvalidPath = errors.New("invalid path")
)

// PathValidationOptions configures path validation behavior
type PathValidationOptions struct {
	// BaseDir is the root directory that paths must stay within
	BaseDir string

	// AllowAbsolute allows absolute paths if they're within BaseDir
	AllowAbsolute bool

	// FollowSymlinks determines if symlinks should be resolved and validated
	FollowSymlinks bool

	// RequireExists requires the path to exist
	RequireExists bool
}

// ValidatePath validates that a file path is safe and doesn't attempt directory traversal
// It checks for:
// - Path traversal attempts (../)
// - Absolute paths (when not allowed)
// - Symlink attacks (when following symlinks)
// - Paths outside the base directory
//
// Returns the cleaned, absolute path if valid, or an error if unsafe.
func ValidatePath(inputPath string, opts PathValidationOptions) (string, error) {
	if inputPath == "" {
		return "", fmt.Errorf("%w: empty path", ErrInvalidPath)
	}

	// Get absolute base directory
	baseDir, err := filepath.Abs(opts.BaseDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve base directory: %w", err)
	}

	// Check for null bytes (security issue in some contexts)
	if strings.Contains(inputPath, "\x00") {
		return "", fmt.Errorf("%w: null byte in path", ErrInvalidPath)
	}

	// Handle absolute paths
	if filepath.IsAbs(inputPath) {
		if !opts.AllowAbsolute {
			return "", fmt.Errorf("%w: %s", ErrAbsolutePath, inputPath)
		}
		// For absolute paths, just verify they're within baseDir
		return validateAbsolutePath(inputPath, baseDir, opts)
	}

	// For relative paths, join with base directory
	fullPath := filepath.Join(baseDir, inputPath)

	return validateAbsolutePath(fullPath, baseDir, opts)
}

// validateAbsolutePath validates an absolute path against a base directory
func validateAbsolutePath(path, baseDir string, opts PathValidationOptions) (string, error) {
	// Clean the path to resolve . and .. components
	cleanPath := filepath.Clean(path)

	// If following symlinks, resolve them
	if opts.FollowSymlinks {
		resolvedPath, err := resolveSymlinks(cleanPath, baseDir)
		if err != nil {
			return "", err
		}
		cleanPath = resolvedPath
	}

	// Check if path exists (if required)
	if opts.RequireExists {
		if _, err := os.Stat(cleanPath); err != nil {
			if os.IsNotExist(err) {
				return "", fmt.Errorf("%w: path does not exist: %s", ErrInvalidPath, cleanPath)
			}
			return "", fmt.Errorf("failed to stat path: %w", err)
		}
	}

	// Verify the path is within baseDir
	if !isWithinBase(cleanPath, baseDir) {
		return "", fmt.Errorf("%w: path %s is outside base directory %s", ErrPathTraversal, cleanPath, baseDir)
	}

	return cleanPath, nil
}

// resolveSymlinks resolves all symlinks in a path and validates they don't escape baseDir
func resolveSymlinks(path, baseDir string) (string, error) {
	// EvalSymlinks follows all symlinks and returns the final path
	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		// If the file doesn't exist, that's ok - we might be validating a path for creation
		if os.IsNotExist(err) {
			return path, nil
		}
		return "", fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	// Verify the resolved path is still within baseDir
	if !isWithinBase(resolvedPath, baseDir) {
		return "", fmt.Errorf("%w: symlink resolves to %s outside base directory %s",
			ErrSymlinkTraversal, resolvedPath, baseDir)
	}

	return resolvedPath, nil
}

// isWithinBase checks if path is within or equal to baseDir
func isWithinBase(path, baseDir string) bool {
	// Clean both paths
	path = filepath.Clean(path)
	baseDir = filepath.Clean(baseDir)

	// Convert to absolute paths
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return false
	}

	// Try to resolve symlinks in baseDir (should usually exist)
	// This handles /var vs /private/var on macOS
	evalBase, err := filepath.EvalSymlinks(absBase)
	if err == nil {
		absBase = evalBase
		// If we resolved the base, we need to normalize the path too
		// by resolving symlinks in the prefix
		// This handles cases where /var -> /private/var
		// We need to ensure both paths use the same symlink resolution

		// Get the directory that would contain the path
		pathDir := filepath.Dir(absPath)

		// Try to resolve symlinks in the directory part
		// (the file itself might not exist yet)
		evalDir, err := filepath.EvalSymlinks(pathDir)
		if err == nil {
			// Reconstruct the path with the resolved directory
			absPath = filepath.Join(evalDir, filepath.Base(absPath))
		} else {
			// If the directory doesn't exist either, try resolving what we can
			// by going up until we find a directory that exists
			for d := pathDir; d != "/" && d != "."; d = filepath.Dir(d) {
				if evalDir, err := filepath.EvalSymlinks(d); err == nil {
					// Found a directory we can resolve, reconstruct the path
					remaining, relErr := filepath.Rel(d, absPath)
					if relErr == nil {
						absPath = filepath.Join(evalDir, remaining)
					}
					break
				}
			}
		}
	}

	// Use Rel to check if path is within base
	// If Rel succeeds without "..", the path is within base
	rel, err := filepath.Rel(absBase, absPath)
	if err != nil {
		return false
	}

	// Check if the relative path tries to go up
	return !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != ".."
}

// MustValidatePath is like ValidatePath but panics on error
// Useful for paths that are known to be safe (e.g., constants)
func MustValidatePath(inputPath string, opts PathValidationOptions) string {
	result, err := ValidatePath(inputPath, opts)
	if err != nil {
		panic(fmt.Sprintf("path validation failed: %v", err))
	}
	return result
}

// ValidatePathSimple is a convenience function for simple path validation
// It uses the current working directory as base and doesn't follow symlinks
func ValidatePathSimple(inputPath string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	return ValidatePath(inputPath, PathValidationOptions{
		BaseDir:        cwd,
		AllowAbsolute:  false,
		FollowSymlinks: false,
		RequireExists:  false,
	})
}
