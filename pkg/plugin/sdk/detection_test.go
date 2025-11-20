package sdk

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseFrameworkDetector(t *testing.T) {
	t.Run("basic detection with required files", func(t *testing.T) {
		// Create test directory
		tmpDir := t.TempDir()

		// Create test files
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.sum"), []byte(""), 0644))

		// Create detector
		detector := NewBaseFrameworkDetector(FrameworkInfo{
			Name:    "test",
			Version: "1.0.0",
			Type:    "language",
		})

		detector.SetPatterns(DetectionPatterns{
			RequiredFiles: []string{"go.mod"},
			OptionalFiles: []string{"go.sum"},
		})

		// Test detection
		result, err := detector.Detect(tmpDir)
		require.NoError(t, err)
		assert.True(t, result.Detected)
		assert.Equal(t, "test", result.Framework.Name)
		assert.Equal(t, "1.0.0", result.Framework.Version)
		assert.Equal(t, 100, result.Confidence) // 20 for required + 10 for optional
	})

	t.Run("detection fails without required files", func(t *testing.T) {
		tmpDir := t.TempDir()

		detector := NewBaseFrameworkDetector(FrameworkInfo{
			Name: "test",
			Type: "language",
		})

		detector.SetPatterns(DetectionPatterns{
			RequiredFiles: []string{"required.txt"},
		})

		result, err := detector.Detect(tmpDir)
		require.NoError(t, err)
		assert.False(t, result.Detected)
	})

	t.Run("detection with directories", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create required file
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte("{}"), 0644))

		// Create directory
		require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "node_modules"), 0755))

		detector := NewBaseFrameworkDetector(FrameworkInfo{
			Name: "node",
			Type: "language",
		})

		detector.SetPatterns(DetectionPatterns{
			RequiredFiles: []string{"package.json"},
			Directories:   []string{"node_modules"},
		})

		result, err := detector.Detect(tmpDir)
		require.NoError(t, err)
		assert.True(t, result.Detected)
		assert.GreaterOrEqual(t, result.Confidence, 50) // Should have sufficient confidence
	})

	t.Run("detection with file content patterns", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create config file with content
		configContent := `{
			"name": "test-app",
			"dependencies": {
				"react": "^18.0.0"
			}
		}`
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(configContent), 0644))

		detector := NewBaseFrameworkDetector(FrameworkInfo{
			Name: "react",
			Type: "framework",
		})

		detector.SetPatterns(DetectionPatterns{
			RequiredFiles: []string{"package.json"},
			FileContents: []ContentPattern{
				{
					Filepath: "package.json",
					Contains: []string{"react"},
				},
			},
		})

		result, err := detector.Detect(tmpDir)
		require.NoError(t, err)
		assert.True(t, result.Detected)
		assert.True(t, result.Confidence >= 50)
	})

	t.Run("detection with file extensions", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create required file
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test"), 0644))

		// Create files with specific extensions
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "test.go"), []byte("package main"), 0644))

		detector := NewBaseFrameworkDetector(FrameworkInfo{
			Name: "go",
			Type: "language",
		})

		detector.SetPatterns(DetectionPatterns{
			RequiredFiles: []string{"go.mod"},
			Extensions:    []string{".go"},
		})

		result, err := detector.Detect(tmpDir)
		require.NoError(t, err)
		assert.True(t, result.Detected)
		assert.GreaterOrEqual(t, result.Confidence, 50) // Should have sufficient confidence
	})

	t.Run("get default commands", func(t *testing.T) {
		detector := NewBaseFrameworkDetector(FrameworkInfo{
			Name: "test",
		})

		commands := map[string]CommandDefinition{
			"build": {
				Cmd:         "make build",
				Description: "Build the project",
			},
			"test": {
				Cmd:         "make test",
				Description: "Run tests",
			},
		}

		detector.SetCommands(commands)

		result := detector.GetDefaultCommands()
		assert.Len(t, result, 2)
		assert.Equal(t, "make build", result["build"].Cmd)
		assert.Equal(t, "make test", result["test"].Cmd)
	})

	t.Run("enhance context", func(t *testing.T) {
		detector := NewBaseFrameworkDetector(FrameworkInfo{
			Name:    "test",
			Version: "1.0.0",
		})

		ctx := make(map[string]interface{})
		err := detector.EnhanceContext(ctx)
		require.NoError(t, err)

		frameworks, ok := ctx["frameworks"].([]string)
		assert.True(t, ok)
		assert.Contains(t, frameworks, "test")

		versions, ok := ctx["framework_versions"].(map[string]string)
		assert.True(t, ok)
		assert.Equal(t, "1.0.0", versions["test"])
	})
}

func TestContentPattern(t *testing.T) {
	t.Run("regex pattern matching", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create file with regex-matchable content
		content := `
			version = "1.2.3"
			name = "test-app"
		`
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "config.toml"), []byte(content), 0644))

		detector := NewBaseFrameworkDetector(FrameworkInfo{
			Name: "test",
		})

		detector.SetPatterns(DetectionPatterns{
			RequiredFiles: []string{"config.toml"},
			FileContents: []ContentPattern{
				{
					Filepath: "config.toml",
					Regex:    `version\s*=\s*"[\d.]+"`,
				},
			},
		})

		result, err := detector.Detect(tmpDir)
		require.NoError(t, err)
		assert.True(t, result.Detected)
	})

	t.Run("multiple contains patterns", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create file with multiple patterns
		content := `{
			"scripts": {
				"test": "jest",
				"build": "webpack"
			}
		}`
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(content), 0644))

		detector := NewBaseFrameworkDetector(FrameworkInfo{
			Name: "js-tools",
		})

		detector.SetPatterns(DetectionPatterns{
			RequiredFiles: []string{"package.json"},
			FileContents: []ContentPattern{
				{
					Filepath: "package.json",
					Contains: []string{"jest", "mocha", "vitest"}, // Any of these
				},
			},
		})

		result, err := detector.Detect(tmpDir)
		require.NoError(t, err)
		assert.True(t, result.Detected)
		assert.True(t, result.Confidence >= 50)
	})
}

func TestConfidenceCalculation(t *testing.T) {
	t.Run("confidence increases with more matches", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create all types of detectable items
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "required.txt"), []byte(""), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "optional.txt"), []byte(""), 0644))
		require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "expected-dir"), 0755))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "config.json"), []byte(`{"key": "value"}`), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "file.ext"), []byte(""), 0644))

		detector := NewBaseFrameworkDetector(FrameworkInfo{
			Name: "full-match",
		})

		detector.SetPatterns(DetectionPatterns{
			RequiredFiles: []string{"required.txt"},
			OptionalFiles: []string{"optional.txt"},
			Directories:   []string{"expected-dir"},
			FileContents: []ContentPattern{
				{
					Filepath: "config.json",
					Contains: []string{"key"},
				},
			},
			Extensions: []string{".ext"},
		})

		result, err := detector.Detect(tmpDir)
		require.NoError(t, err)
		assert.True(t, result.Detected)
		assert.Equal(t, 100, result.Confidence) // All patterns matched
	})

	t.Run("low confidence prevents detection", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create only required file
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "maybe.txt"), []byte(""), 0644))

		detector := NewBaseFrameworkDetector(FrameworkInfo{
			Name: "low-confidence",
		})

		detector.SetPatterns(DetectionPatterns{
			RequiredFiles: []string{"maybe.txt"},
			OptionalFiles: []string{"missing1.txt", "missing2.txt", "missing3.txt"},
			Directories:   []string{"missing-dir1", "missing-dir2"},
		})

		result, err := detector.Detect(tmpDir)
		require.NoError(t, err)
		assert.False(t, result.Detected) // Confidence < 50%
	})
}
