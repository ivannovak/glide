package validation

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestValidatePath(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		inputPath   string
		opts        PathValidationOptions
		expectError bool
		errorType   error
		setup       func() string // Returns path to test against
		cleanup     func()
	}{
		{
			name:      "valid relative path",
			inputPath: "safe/path.txt",
			opts: PathValidationOptions{
				BaseDir:        tmpDir,
				AllowAbsolute:  false,
				FollowSymlinks: false,
				RequireExists:  false,
			},
			expectError: false,
		},
		{
			name:      "path traversal with ../",
			inputPath: "../../../etc/passwd",
			opts: PathValidationOptions{
				BaseDir:        tmpDir,
				AllowAbsolute:  false,
				FollowSymlinks: false,
				RequireExists:  false,
			},
			expectError: true,
			errorType:   ErrPathTraversal,
		},
		{
			name:      "path traversal with mixed separators",
			inputPath: "safe/../../../etc/passwd",
			opts: PathValidationOptions{
				BaseDir:        tmpDir,
				AllowAbsolute:  false,
				FollowSymlinks: false,
				RequireExists:  false,
			},
			expectError: true,
			errorType:   ErrPathTraversal,
		},
		{
			name:      "absolute path when not allowed",
			inputPath: "/etc/passwd",
			opts: PathValidationOptions{
				BaseDir:        tmpDir,
				AllowAbsolute:  false,
				FollowSymlinks: false,
				RequireExists:  false,
			},
			expectError: true,
			errorType:   ErrAbsolutePath,
		},
		{
			name:      "absolute path when allowed and within base",
			inputPath: filepath.Join(tmpDir, "safe.txt"),
			opts: PathValidationOptions{
				BaseDir:        tmpDir,
				AllowAbsolute:  true,
				FollowSymlinks: false,
				RequireExists:  false,
			},
			expectError: false,
		},
		{
			name:      "absolute path outside base when allowed",
			inputPath: "/etc/passwd",
			opts: PathValidationOptions{
				BaseDir:        tmpDir,
				AllowAbsolute:  true,
				FollowSymlinks: false,
				RequireExists:  false,
			},
			expectError: true,
			errorType:   ErrPathTraversal,
		},
		{
			name:      "empty path",
			inputPath: "",
			opts: PathValidationOptions{
				BaseDir:        tmpDir,
				AllowAbsolute:  false,
				FollowSymlinks: false,
				RequireExists:  false,
			},
			expectError: true,
			errorType:   ErrInvalidPath,
		},
		{
			name:      "null byte injection",
			inputPath: "safe\x00../../../etc/passwd",
			opts: PathValidationOptions{
				BaseDir:        tmpDir,
				AllowAbsolute:  false,
				FollowSymlinks: false,
				RequireExists:  false,
			},
			expectError: true,
			errorType:   ErrInvalidPath,
		},
		{
			name:      "path with . components",
			inputPath: "./safe/./path/./file.txt",
			opts: PathValidationOptions{
				BaseDir:        tmpDir,
				AllowAbsolute:  false,
				FollowSymlinks: false,
				RequireExists:  false,
			},
			expectError: false,
		},
		{
			name:      "require exists - file exists",
			inputPath: "existing.txt",
			opts: PathValidationOptions{
				BaseDir:        tmpDir,
				AllowAbsolute:  false,
				FollowSymlinks: false,
				RequireExists:  true,
			},
			setup: func() string {
				path := filepath.Join(tmpDir, "existing.txt")
				if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
				return "existing.txt"
			},
			expectError: false,
		},
		{
			name:      "require exists - file does not exist",
			inputPath: "nonexistent.txt",
			opts: PathValidationOptions{
				BaseDir:        tmpDir,
				AllowAbsolute:  false,
				FollowSymlinks: false,
				RequireExists:  true,
			},
			expectError: true,
			errorType:   ErrInvalidPath,
		},
		{
			name:      "symlink within base",
			inputPath: "symlink.txt",
			opts: PathValidationOptions{
				BaseDir:        tmpDir,
				AllowAbsolute:  false,
				FollowSymlinks: true,
				RequireExists:  false,
			},
			setup: func() string {
				targetPath := filepath.Join(tmpDir, "target.txt")
				symlinkPath := filepath.Join(tmpDir, "symlink.txt")

				if err := os.WriteFile(targetPath, []byte("test"), 0644); err != nil {
					t.Fatalf("setup failed: %v", err)
				}

				if err := os.Symlink(targetPath, symlinkPath); err != nil {
					t.Fatalf("setup failed: %v", err)
				}

				return "symlink.txt"
			},
			cleanup: func() {
				os.Remove(filepath.Join(tmpDir, "symlink.txt"))
				os.Remove(filepath.Join(tmpDir, "target.txt"))
			},
			expectError: false,
		},
		{
			name:      "symlink traversal attack",
			inputPath: "evil_symlink.txt",
			opts: PathValidationOptions{
				BaseDir:        tmpDir,
				AllowAbsolute:  false,
				FollowSymlinks: true,
				RequireExists:  false,
			},
			setup: func() string {
				symlinkPath := filepath.Join(tmpDir, "evil_symlink.txt")

				// Create symlink pointing outside base dir
				if err := os.Symlink("/etc/passwd", symlinkPath); err != nil {
					t.Fatalf("setup failed: %v", err)
				}

				return "evil_symlink.txt"
			},
			cleanup: func() {
				os.Remove(filepath.Join(tmpDir, "evil_symlink.txt"))
			},
			expectError: true,
			errorType:   ErrSymlinkTraversal,
		},
	}

	// Add Windows-specific tests
	if runtime.GOOS == "windows" {
		tests = append(tests, []struct {
			name        string
			inputPath   string
			opts        PathValidationOptions
			expectError bool
			errorType   error
			setup       func() string
			cleanup     func()
		}{
			{
				name:      "windows path traversal with backslashes",
				inputPath: "..\\..\\..\\Windows\\System32\\config\\sam",
				opts: PathValidationOptions{
					BaseDir:        tmpDir,
					AllowAbsolute:  false,
					FollowSymlinks: false,
					RequireExists:  false,
				},
				expectError: true,
				errorType:   ErrPathTraversal,
			},
			{
				name:      "windows absolute path",
				inputPath: "C:\\Windows\\System32\\config\\sam",
				opts: PathValidationOptions{
					BaseDir:        tmpDir,
					AllowAbsolute:  false,
					FollowSymlinks: false,
					RequireExists:  false,
				},
				expectError: true,
				errorType:   ErrAbsolutePath,
			},
		}...)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run setup if provided
			inputPath := tt.inputPath
			if tt.setup != nil {
				inputPath = tt.setup()
			}

			// Run cleanup if provided
			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			// Test the validation
			result, err := ValidatePath(inputPath, tt.opts)

			// Check error expectation
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none, result: %s", result)
					return
				}

				// Check error type if specified
				if tt.errorType != nil && !containsError(err, tt.errorType) {
					t.Errorf("expected error type %v, got %v", tt.errorType, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				// Verify result is within base directory
				absBase, _ := filepath.Abs(tt.opts.BaseDir)
				if !isWithinBase(result, absBase) {
					t.Errorf("result %s is outside base directory %s", result, absBase)
				}
			}
		})
	}
}

func TestValidatePathSimple(t *testing.T) {
	tests := []struct {
		name        string
		inputPath   string
		expectError bool
	}{
		{
			name:        "valid relative path",
			inputPath:   "safe/path.txt",
			expectError: false,
		},
		{
			name:        "path traversal",
			inputPath:   "../../../etc/passwd",
			expectError: true,
		},
		{
			name:        "absolute path",
			inputPath:   "/etc/passwd",
			expectError: true,
		},
		{
			name:        "empty path",
			inputPath:   "",
			expectError: true,
		},
		{
			name:        "null byte",
			inputPath:   "test\x00file",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidatePathSimple(tt.inputPath)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestMustValidatePath(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("panics on invalid path", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic but got none")
			}
		}()

		MustValidatePath("../../../etc/passwd", PathValidationOptions{
			BaseDir:        tmpDir,
			AllowAbsolute:  false,
			FollowSymlinks: false,
			RequireExists:  false,
		})
	})

	t.Run("does not panic on valid path", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("unexpected panic: %v", r)
			}
		}()

		result := MustValidatePath("safe/path.txt", PathValidationOptions{
			BaseDir:        tmpDir,
			AllowAbsolute:  false,
			FollowSymlinks: false,
			RequireExists:  false,
		})

		if result == "" {
			t.Error("expected non-empty result")
		}
	})
}

func TestIsWithinBase(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		path     string
		baseDir  string
		expected bool
	}{
		{
			name:     "path within base",
			path:     filepath.Join(tmpDir, "safe", "path.txt"),
			baseDir:  tmpDir,
			expected: true,
		},
		{
			name:     "path equal to base",
			path:     tmpDir,
			baseDir:  tmpDir,
			expected: true,
		},
		{
			name:     "path outside base",
			path:     "/etc/passwd",
			baseDir:  tmpDir,
			expected: false,
		},
		{
			name:     "path with .. outside base",
			path:     filepath.Join(tmpDir, "..", "..", "etc", "passwd"),
			baseDir:  tmpDir,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isWithinBase(tt.path, tt.baseDir)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestResolveSymlinks(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("resolves symlink within base", func(t *testing.T) {
		targetPath := filepath.Join(tmpDir, "target.txt")
		symlinkPath := filepath.Join(tmpDir, "symlink.txt")

		if err := os.WriteFile(targetPath, []byte("test"), 0644); err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		defer os.Remove(targetPath)

		if err := os.Symlink(targetPath, symlinkPath); err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		defer os.Remove(symlinkPath)

		result, err := resolveSymlinks(symlinkPath, tmpDir)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// The result should be the fully resolved target path
		// On macOS, this might be /private/var/... instead of /var/...
		// So we check that the symlink resolved correctly by comparing
		// with what filepath.EvalSymlinks returns
		expectedPath, _ := filepath.EvalSymlinks(targetPath)
		if result != expectedPath {
			t.Errorf("expected %s, got %s", expectedPath, result)
		}
	})

	t.Run("detects symlink traversal", func(t *testing.T) {
		symlinkPath := filepath.Join(tmpDir, "evil_symlink.txt")

		if err := os.Symlink("/etc/passwd", symlinkPath); err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		defer os.Remove(symlinkPath)

		_, err := resolveSymlinks(symlinkPath, tmpDir)
		if err == nil {
			t.Error("expected error but got none")
		}

		if !containsError(err, ErrSymlinkTraversal) {
			t.Errorf("expected ErrSymlinkTraversal, got %v", err)
		}
	})

	t.Run("handles non-existent symlink", func(t *testing.T) {
		nonExistentPath := filepath.Join(tmpDir, "nonexistent.txt")

		result, err := resolveSymlinks(nonExistentPath, tmpDir)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Should return the original path since it doesn't exist
		if result != nonExistentPath {
			t.Errorf("expected %s, got %s", nonExistentPath, result)
		}
	})
}

// TestEdgeCases tests various edge cases for improved coverage
func TestEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("stat error on existing file check", func(t *testing.T) {
		// This tests the error path in resolveSymlinks when file doesn't exist
		result, err := ValidatePath("nonexistent.txt", PathValidationOptions{
			BaseDir:        tmpDir,
			AllowAbsolute:  false,
			FollowSymlinks: true,  // This triggers resolveSymlinks
			RequireExists:  false, // Don't require it to exist
		})
		if err != nil {
			t.Errorf("unexpected error for nonexistent file with RequireExists=false: %v", err)
		}
		if result == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("abs error handling", func(t *testing.T) {
		// Test with an invalid base directory would be difficult to trigger
		// as Abs is very robust, but we can test the path itself
		result, err := ValidatePath("valid/path", PathValidationOptions{
			BaseDir:        tmpDir,
			AllowAbsolute:  false,
			FollowSymlinks: false,
			RequireExists:  false,
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("evalSymlinks in isWithinBase with deeply nested path", func(t *testing.T) {
		// Create a deeply nested relative path to test the isWithinBase logic
		result, err := ValidatePath("a/b/c/d/e/file.txt", PathValidationOptions{
			BaseDir:        tmpDir,
			AllowAbsolute:  false,
			FollowSymlinks: false,
			RequireExists:  false,
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == "" {
			t.Error("expected non-empty result")
		}
	})
}

// BenchmarkValidatePath benchmarks the path validation function
func BenchmarkValidatePath(b *testing.B) {
	tmpDir := b.TempDir()
	opts := PathValidationOptions{
		BaseDir:        tmpDir,
		AllowAbsolute:  false,
		FollowSymlinks: false,
		RequireExists:  false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ValidatePath("safe/path.txt", opts)
	}
}

func BenchmarkValidatePathWithSymlinks(b *testing.B) {
	tmpDir := b.TempDir()

	// Create a symlink for benchmarking
	targetPath := filepath.Join(tmpDir, "target.txt")
	symlinkPath := filepath.Join(tmpDir, "symlink.txt")

	if err := os.WriteFile(targetPath, []byte("test"), 0644); err != nil {
		b.Fatalf("setup failed: %v", err)
	}

	if err := os.Symlink(targetPath, symlinkPath); err != nil {
		b.Fatalf("setup failed: %v", err)
	}

	opts := PathValidationOptions{
		BaseDir:        tmpDir,
		AllowAbsolute:  false,
		FollowSymlinks: true,
		RequireExists:  false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ValidatePath("symlink.txt", opts)
	}
}

// containsError checks if an error chain contains a specific error
func containsError(err, target error) bool {
	if err == nil {
		return target == nil
	}
	if target == nil {
		return false
	}

	// Simple string matching since we don't have errors.Is from Go 1.13+
	// but our error messages wrap the sentinel errors
	return err == target || (err != nil && target != nil && err.Error() != "" &&
		len(err.Error()) >= len(target.Error()) &&
		err.Error()[:len(target.Error())] == target.Error())
}
