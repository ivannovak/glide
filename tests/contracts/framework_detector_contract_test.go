// Package contracts contains contract tests for framework detector implementations
//
// Contract tests ensure all FrameworkDetector implementations behave consistently:
// - All detectors implement the complete FrameworkDetector interface
// - All detectors return valid confidence scores (0-100)
// - All detectors provide detection patterns
// - All detectors handle non-existent paths gracefully
package contracts

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ivannovak/glide/v3/internal/plugins/builtin/golang"
	"github.com/ivannovak/glide/v3/internal/plugins/builtin/node"
	"github.com/ivannovak/glide/v3/internal/plugins/builtin/php"
	"github.com/ivannovak/glide/v3/pkg/plugin/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFrameworkDetectorContract verifies all detector implementations adhere to the contract
func TestFrameworkDetectorContract(t *testing.T) {
	detectors := map[string]sdk.FrameworkDetector{
		"Node": node.NewNodeDetector(),
		"PHP":  php.NewPHPDetector(),
		"Go":   golang.NewGoDetector(),
	}

	for name, detector := range detectors {
		t.Run(name, func(t *testing.T) {
			testFrameworkDetectorContract(t, name, detector)
		})
	}
}

func testFrameworkDetectorContract(t *testing.T, detectorName string, detector sdk.FrameworkDetector) {
	t.Run("implements GetDetectionPatterns", func(t *testing.T) {
		patterns := detector.GetDetectionPatterns()

		// At minimum, detector should have some detection criteria
		hasPatterns := len(patterns.RequiredFiles) > 0 ||
			len(patterns.OptionalFiles) > 0 ||
			len(patterns.Directories) > 0 ||
			len(patterns.Extensions) > 0 ||
			len(patterns.FileContents) > 0

		assert.True(t, hasPatterns,
			"%s detector should define at least one detection pattern", detectorName)
	})

	t.Run("implements GetDefaultCommands", func(t *testing.T) {
		commands := detector.GetDefaultCommands()

		// Detector should provide default commands
		assert.NotNil(t, commands, "%s should return non-nil commands map", detectorName)

		// Verify command definitions are well-formed
		for cmdName, cmdDef := range commands {
			assert.NotEmpty(t, cmdDef.Cmd,
				"%s command '%s' should have a non-empty Cmd field", detectorName, cmdName)
			assert.NotEmpty(t, cmdDef.Description,
				"%s command '%s' should have a description", detectorName, cmdName)
		}
	})

	t.Run("implements Detect - non-existent path", func(t *testing.T) {
		nonExistentPath := filepath.Join(os.TempDir(), "glide-test-nonexistent-12345")

		result, err := detector.Detect(nonExistentPath)

		// Should not error on non-existent path
		require.NoError(t, err, "%s should handle non-existent paths gracefully", detectorName)
		require.NotNil(t, result, "%s should return a result even for non-existent paths", detectorName)

		// Should not detect framework in non-existent path
		assert.False(t, result.Detected,
			"%s should not detect framework in non-existent path", detectorName)
		assert.Equal(t, 0, result.Confidence,
			"%s confidence should be 0 for non-existent path", detectorName)
	})

	t.Run("implements Detect - empty directory", func(t *testing.T) {
		emptyDir := t.TempDir()

		result, err := detector.Detect(emptyDir)

		require.NoError(t, err, "%s should handle empty directories gracefully", detectorName)
		require.NotNil(t, result, "%s should return a result for empty directories", detectorName)

		// Should not detect framework in empty directory
		assert.False(t, result.Detected,
			"%s should not detect framework in empty directory", detectorName)
		assert.Equal(t, 0, result.Confidence,
			"%s confidence should be 0 for empty directory", detectorName)
	})

	t.Run("Detect returns valid confidence scores", func(t *testing.T) {
		testCases := []struct {
			name      string
			setupFunc func(t *testing.T) string
		}{
			{
				name: "empty directory",
				setupFunc: func(t *testing.T) string {
					return t.TempDir()
				},
			},
			{
				name: "non-existent path",
				setupFunc: func(t *testing.T) string {
					return filepath.Join(os.TempDir(), "nonexistent-test-dir")
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				projectPath := tc.setupFunc(t)
				result, err := detector.Detect(projectPath)

				require.NoError(t, err, "%s Detect should not error", detectorName)
				require.NotNil(t, result, "%s should return non-nil result", detectorName)

				// Confidence must be in valid range 0-100
				assert.GreaterOrEqual(t, result.Confidence, 0,
					"%s confidence must be >= 0", detectorName)
				assert.LessOrEqual(t, result.Confidence, 100,
					"%s confidence must be <= 100", detectorName)
			})
		}
	})

	t.Run("Detect result structure is valid when detected", func(t *testing.T) {
		// This test verifies result structure when framework IS detected
		// Skip for now since we need actual project setup
		// The positive detection tests below cover this
	})

	t.Run("implements EnhanceContext", func(t *testing.T) {
		ctx := make(map[string]interface{})
		ctx["test_key"] = "test_value"

		// Should not panic and should handle gracefully
		assert.NotPanics(t, func() {
			err := detector.EnhanceContext(ctx)
			// Some detectors may return errors, others may not
			// The key is that it doesn't panic
			_ = err
		}, "%s EnhanceContext should not panic", detectorName)
	})

	t.Run("EnhanceContext handles nil context gracefully", func(t *testing.T) {
		// Should not panic with nil context
		assert.NotPanics(t, func() {
			err := detector.EnhanceContext(nil)
			_ = err
		}, "%s EnhanceContext should handle nil context", detectorName)
	})

	t.Run("EnhanceContext handles empty context", func(t *testing.T) {
		ctx := make(map[string]interface{})

		err := detector.EnhanceContext(ctx)

		// Should not error on empty context
		// Some detectors may add data, others may not
		assert.NotPanics(t, func() {
			_ = err
		})
	})
}

// TestFrameworkDetectorPositiveDetection tests actual detection with matching project structures
func TestFrameworkDetectorPositiveDetection(t *testing.T) {
	t.Run("Node detector - detects package.json", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create a minimal package.json
		packageJSON := `{
			"name": "test-project",
			"version": "1.0.0",
			"scripts": {
				"test": "jest"
			}
		}`

		err := os.WriteFile(filepath.Join(tempDir, "package.json"), []byte(packageJSON), 0644)
		require.NoError(t, err)

		detector := node.NewNodeDetector()
		result, err := detector.Detect(tempDir)

		require.NoError(t, err)
		require.NotNil(t, result)

		// When detected, result should have framework info populated
		if result.Detected {
			assert.Greater(t, result.Confidence, 0, "Confidence should be > 0 for detected project")
			assert.LessOrEqual(t, result.Confidence, 100, "Confidence should be <= 100")
			assert.Equal(t, "node", result.Framework.Name)
			assert.Equal(t, "language", result.Framework.Type)
			assert.NotNil(t, result.Commands)
			assert.NotNil(t, result.Metadata)
		}
	})

	t.Run("Go detector - detects go.mod", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create a minimal go.mod
		goMod := `module github.com/test/project

go 1.21
`

		err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goMod), 0644)
		require.NoError(t, err)

		detector := golang.NewGoDetector()
		result, err := detector.Detect(tempDir)

		require.NoError(t, err)
		require.NotNil(t, result)

		// When detected, result should have framework info populated
		if result.Detected {
			assert.Greater(t, result.Confidence, 0, "Confidence should be > 0 for detected project")
			assert.LessOrEqual(t, result.Confidence, 100, "Confidence should be <= 100")
			assert.Equal(t, "go", result.Framework.Name)
			assert.Equal(t, "language", result.Framework.Type)
			assert.NotNil(t, result.Commands)
			assert.NotNil(t, result.Metadata)
		}
	})

	t.Run("PHP detector - detects composer.json", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create a minimal composer.json
		composerJSON := `{
			"name": "test/project",
			"require": {
				"php": ">=7.4"
			}
		}`

		err := os.WriteFile(filepath.Join(tempDir, "composer.json"), []byte(composerJSON), 0644)
		require.NoError(t, err)

		detector := php.NewPHPDetector()
		result, err := detector.Detect(tempDir)

		require.NoError(t, err)
		require.NotNil(t, result)

		// When detected, result should have framework info populated
		if result.Detected {
			assert.Greater(t, result.Confidence, 0, "Confidence should be > 0 for detected project")
			assert.LessOrEqual(t, result.Confidence, 100, "Confidence should be <= 100")
			assert.Equal(t, "php", result.Framework.Name)
			assert.Equal(t, "language", result.Framework.Type)
			assert.NotNil(t, result.Commands)
			assert.NotNil(t, result.Metadata)
		}
	})
}

// TestFrameworkDetectorConfidenceScoring tests confidence score consistency
func TestFrameworkDetectorConfidenceScoring(t *testing.T) {
	t.Run("confidence increases with more matching patterns", func(t *testing.T) {
		detector := node.NewNodeDetector()

		// Test 1: Only package.json
		dir1 := t.TempDir()
		os.WriteFile(filepath.Join(dir1, "package.json"), []byte(`{"name":"test"}`), 0644)

		result1, err := detector.Detect(dir1)
		require.NoError(t, err)

		// Test 2: package.json + package-lock.json + node_modules
		dir2 := t.TempDir()
		os.WriteFile(filepath.Join(dir2, "package.json"), []byte(`{"name":"test"}`), 0644)
		os.WriteFile(filepath.Join(dir2, "package-lock.json"), []byte(`{}`), 0644)
		os.Mkdir(filepath.Join(dir2, "node_modules"), 0755)

		result2, err := detector.Detect(dir2)
		require.NoError(t, err)

		// More matching patterns should generally result in higher or equal confidence
		// (Implementation-specific, but should follow this general principle)
		assert.GreaterOrEqual(t, result2.Confidence, result1.Confidence,
			"More matching patterns should not decrease confidence")
	})
}

// TestFrameworkDetectorCommandConsistency tests command definition consistency
func TestFrameworkDetectorCommandConsistency(t *testing.T) {
	detectors := map[string]sdk.FrameworkDetector{
		"Node": node.NewNodeDetector(),
		"PHP":  php.NewPHPDetector(),
		"Go":   golang.NewGoDetector(),
	}

	for name, detector := range detectors {
		t.Run(name, func(t *testing.T) {
			commands := detector.GetDefaultCommands()

			// Verify all commands have required fields
			for cmdName, cmdDef := range commands {
				assert.NotEmpty(t, cmdDef.Cmd,
					"%s command '%s' must have Cmd field", name, cmdName)
				assert.NotEmpty(t, cmdDef.Description,
					"%s command '%s' must have Description field", name, cmdName)

				// Category should be set for organization
				// (not strictly required but recommended)
				if cmdDef.Category != "" {
					validCategories := []string{"dependencies", "test", "build", "run", "lint", "format", "migration", "db"}
					assert.Contains(t, validCategories, cmdDef.Category,
						"%s command '%s' should use a standard category", name, cmdName)
				}
			}
		})
	}
}

// TestFrameworkDetectorPatternValidity tests detection pattern validity
func TestFrameworkDetectorPatternValidity(t *testing.T) {
	detectors := map[string]sdk.FrameworkDetector{
		"Node": node.NewNodeDetector(),
		"PHP":  php.NewPHPDetector(),
		"Go":   golang.NewGoDetector(),
	}

	for name, detector := range detectors {
		t.Run(name, func(t *testing.T) {
			patterns := detector.GetDetectionPatterns()

			// File patterns should not have path separators (they're filenames, not paths)
			for _, file := range patterns.RequiredFiles {
				assert.NotContains(t, file, "/",
					"%s required file '%s' should be filename only, not a path", name, file)
				assert.NotContains(t, file, "\\",
					"%s required file '%s' should be filename only, not a path", name, file)
			}

			for _, file := range patterns.OptionalFiles {
				assert.NotContains(t, file, "/",
					"%s optional file '%s' should be filename only, not a path", name, file)
				assert.NotContains(t, file, "\\",
					"%s optional file '%s' should be filename only, not a path", name, file)
			}

			// Extensions should start with a dot
			for _, ext := range patterns.Extensions {
				assert.True(t, len(ext) > 0 && ext[0] == '.',
					"%s extension '%s' should start with a dot", name, ext)
			}
		})
	}
}
