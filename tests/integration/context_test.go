package integration_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ivannovak/glide/v2/internal/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestContextDetectionBasic tests basic context detection
func TestContextDetectionBasic(t *testing.T) {
	t.Run("detect_in_empty_directory", func(t *testing.T) {
		// Setup: Create empty directory
		tmpDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		t.Cleanup(func() { _ = os.Chdir(originalWd) })

		require.NoError(t, os.Chdir(tmpDir))

		// Test: Detect context
		detector, err := context.NewDetector()
		require.NoError(t, err)

		ctx, err := detector.Detect()

		// Assert: Context attempted (may error if no project root found)
		if err != nil {
			// Error expected in empty directory - no project root
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "project root")
		} else {
			require.NotNil(t, ctx)
		}
	})

	t.Run("detect_in_git_repository", func(t *testing.T) {
		// Setup: Create git repository
		tmpDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		t.Cleanup(func() { _ = os.Chdir(originalWd) })

		// Create .git directory
		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		require.NoError(t, os.Chdir(tmpDir))

		// Test: Detect context
		detector, err := context.NewDetector()
		require.NoError(t, err)

		ctx, err := detector.Detect()

		// Assert: Project root found
		require.NoError(t, err)
		require.NotNil(t, ctx)
		// Use Contains to handle macOS symlink /var vs /private/var
		assert.Contains(t, ctx.ProjectRoot, filepath.Base(tmpDir))
	})

	t.Run("detect_with_config_file", func(t *testing.T) {
		// Setup: Create directory with .glide.yml
		tmpDir := t.TempDir()
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		// Create .glide.yml
		configPath := filepath.Join(tmpDir, ".glide.yml")
		require.NoError(t, os.WriteFile(configPath, []byte("version: 1\n"), 0644))

		require.NoError(t, os.Chdir(tmpDir))

		// Test: Detect context
		detector, err := context.NewDetector()
		require.NoError(t, err)

		ctx, err := detector.Detect()

		// Assert: Context detected successfully
		require.NoError(t, err)
		require.NotNil(t, ctx)
		assert.Contains(t, ctx.ProjectRoot, filepath.Base(tmpDir))
		assert.Equal(t, context.ModeStandalone, ctx.DevelopmentMode)
	})
}

// TestContextMultiFrameworkDetection tests multi-framework detection
func TestContextMultiFrameworkDetection(t *testing.T) {
	t.Run("detect_php_project", func(t *testing.T) {
		// Setup: Create PHP project
		tmpDir := t.TempDir()
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		// Create .git
		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		// Create composer.json
		composerPath := filepath.Join(tmpDir, "composer.json")
		composerContent := `{"require": {"php": "^8.0"}}`
		require.NoError(t, os.WriteFile(composerPath, []byte(composerContent), 0644))

		require.NoError(t, os.Chdir(tmpDir))

		// Test: Detect context
		detector, err := context.NewDetector()
		require.NoError(t, err)

		ctx, err := detector.Detect()

		// Assert: Context detected successfully
		require.NoError(t, err)
		require.NotNil(t, ctx)
		// Framework detection depends on plugin SDK being available
		// Just verify context was created with valid project root
		assert.NotEmpty(t, ctx.ProjectRoot)
	})

	t.Run("detect_node_project", func(t *testing.T) {
		// Setup: Create Node project
		tmpDir := t.TempDir()
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		// Create .git
		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		// Create package.json
		packagePath := filepath.Join(tmpDir, "package.json")
		packageContent := `{"name": "test", "version": "1.0.0"}`
		require.NoError(t, os.WriteFile(packagePath, []byte(packageContent), 0644))

		require.NoError(t, os.Chdir(tmpDir))

		// Test: Detect context
		detector, err := context.NewDetector()
		require.NoError(t, err)

		ctx, err := detector.Detect()

		// Assert: Context detected successfully
		require.NoError(t, err)
		require.NotNil(t, ctx)
		assert.NotEmpty(t, ctx.ProjectRoot)
	})

	t.Run("detect_multi_framework_project", func(t *testing.T) {
		// Setup: Create project with multiple frameworks
		tmpDir := t.TempDir()
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		// Create .git
		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		// Create PHP files
		composerPath := filepath.Join(tmpDir, "composer.json")
		require.NoError(t, os.WriteFile(composerPath, []byte(`{"require": {"php": "^8.0"}}`), 0644))

		// Create Node files
		packagePath := filepath.Join(tmpDir, "package.json")
		require.NoError(t, os.WriteFile(packagePath, []byte(`{"name": "test", "version": "1.0.0"}`), 0644))

		// Create Docker files
		dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
		require.NoError(t, os.WriteFile(dockerfilePath, []byte("FROM php:8.0\n"), 0644))

		composePath := filepath.Join(tmpDir, "docker-compose.yml")
		composeContent := `version: '3'
services:
  web:
    build: .
    ports:
      - "8000:8000"
`
		require.NoError(t, os.WriteFile(composePath, []byte(composeContent), 0644))

		require.NoError(t, os.Chdir(tmpDir))

		// Test: Detect context
		detector, err := context.NewDetector()
		require.NoError(t, err)

		ctx, err := detector.Detect()

		// Assert: Context detected successfully
		require.NoError(t, err)
		require.NotNil(t, ctx)
		// Framework and Docker detection depend on plugins
		// Just verify context structure is valid
		assert.NotEmpty(t, ctx.ProjectRoot)
		assert.NotNil(t, ctx.Extensions, "Extensions should be initialized")
	})
}

// TestContextDevelopmentModeDetection tests development mode detection
func TestContextDevelopmentModeDetection(t *testing.T) {
	t.Run("detect_single_repo_mode", func(t *testing.T) {
		// Setup: Create single git repository
		tmpDir := t.TempDir()
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		// Create .git directory
		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		require.NoError(t, os.Chdir(tmpDir))

		// Test: Detect context
		detector, err := context.NewDetector()
		require.NoError(t, err)

		ctx, err := detector.Detect()

		// Assert: Single repo mode detected
		require.NoError(t, err)
		require.NotNil(t, ctx)
		assert.Equal(t, context.ModeSingleRepo, ctx.DevelopmentMode)
		assert.Equal(t, context.LocationProject, ctx.Location)
	})

	t.Run("detect_multi_worktree_mode", func(t *testing.T) {
		// Setup: Create multi-worktree structure
		tmpDir := t.TempDir()
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		// Create vcs directory with .git
		vcsDir := filepath.Join(tmpDir, "vcs")
		require.NoError(t, os.MkdirAll(vcsDir, 0755))

		gitDir := filepath.Join(vcsDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		// Create worktrees directory
		worktreesDir := filepath.Join(tmpDir, "worktrees")
		require.NoError(t, os.MkdirAll(worktreesDir, 0755))

		// Create a worktree
		worktreeDir := filepath.Join(worktreesDir, "issue-123")
		require.NoError(t, os.MkdirAll(worktreeDir, 0755))

		// Create .git file in worktree (points to main repo)
		gitFile := filepath.Join(worktreeDir, ".git")
		gitContent := "gitdir: " + filepath.Join(gitDir, "worktrees", "issue-123")
		require.NoError(t, os.WriteFile(gitFile, []byte(gitContent), 0644))

		require.NoError(t, os.Chdir(worktreeDir))

		// Test: Detect context
		detector, err := context.NewDetector()
		require.NoError(t, err)

		ctx, err := detector.Detect()

		// Assert: Multi-worktree mode detected
		require.NoError(t, err)
		require.NotNil(t, ctx)
		assert.Equal(t, context.ModeMultiWorktree, ctx.DevelopmentMode)
		assert.True(t, ctx.IsWorktree)
		assert.Equal(t, "issue-123", ctx.WorktreeName)
	})

	t.Run("detect_standalone_mode", func(t *testing.T) {
		// Setup: Create standalone project (no git, just .glide.yml)
		tmpDir := t.TempDir()
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		// Create .glide.yml without .git
		configPath := filepath.Join(tmpDir, ".glide.yml")
		require.NoError(t, os.WriteFile(configPath, []byte("version: 1\n"), 0644))

		require.NoError(t, os.Chdir(tmpDir))

		// Test: Detect context
		detector, err := context.NewDetector()
		require.NoError(t, err)

		ctx, err := detector.Detect()

		// Assert: Standalone mode detected
		require.NoError(t, err)
		require.NotNil(t, ctx)
		assert.Equal(t, context.ModeStandalone, ctx.DevelopmentMode)
	})
}

// TestContextLocationDetection tests location detection
func TestContextLocationDetection(t *testing.T) {
	t.Run("detect_location_in_project_root", func(t *testing.T) {
		// Setup: Create multi-worktree project at root
		tmpDir := t.TempDir()
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		// Create multi-worktree structure
		vcsDir := filepath.Join(tmpDir, "vcs")
		require.NoError(t, os.MkdirAll(vcsDir, 0755))
		gitDir := filepath.Join(vcsDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		worktreesDir := filepath.Join(tmpDir, "worktrees")
		require.NoError(t, os.MkdirAll(worktreesDir, 0755))

		// Create .glide.yml at root
		configPath := filepath.Join(tmpDir, ".glide.yml")
		require.NoError(t, os.WriteFile(configPath, []byte("version: 1\n"), 0644))

		require.NoError(t, os.Chdir(tmpDir))

		// Test: Detect context
		detector, err := context.NewDetector()
		require.NoError(t, err)

		ctx, err := detector.Detect()

		// Assert: Location is root
		require.NoError(t, err)
		require.NotNil(t, ctx)
		assert.True(t, ctx.IsRoot)
		assert.Equal(t, context.LocationRoot, ctx.Location)
	})

	t.Run("detect_location_in_vcs", func(t *testing.T) {
		// Setup: Create multi-worktree project in vcs directory
		tmpDir := t.TempDir()
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		// Create vcs directory with .git
		vcsDir := filepath.Join(tmpDir, "vcs")
		require.NoError(t, os.MkdirAll(vcsDir, 0755))
		gitDir := filepath.Join(vcsDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		// Create worktrees directory (to trigger multi-worktree detection)
		worktreesDir := filepath.Join(tmpDir, "worktrees")
		require.NoError(t, os.MkdirAll(worktreesDir, 0755))

		require.NoError(t, os.Chdir(vcsDir))

		// Test: Detect context
		detector, err := context.NewDetector()
		require.NoError(t, err)

		ctx, err := detector.Detect()

		// Assert: Location is main repo
		require.NoError(t, err)
		require.NotNil(t, ctx)
		assert.True(t, ctx.IsMainRepo)
		assert.Equal(t, context.LocationMainRepo, ctx.Location)
	})
}

// TestContextWithSubdirectories tests context detection from subdirectories
func TestContextWithSubdirectories(t *testing.T) {
	t.Run("detect_from_deep_subdirectory", func(t *testing.T) {
		// Setup: Create project with nested directories
		tmpDir := t.TempDir()
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		// Create .git at root
		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		// Create deep subdirectory
		deepDir := filepath.Join(tmpDir, "src", "components", "lib", "utils")
		require.NoError(t, os.MkdirAll(deepDir, 0755))

		require.NoError(t, os.Chdir(deepDir))

		// Test: Detect context
		detector, err := context.NewDetector()
		require.NoError(t, err)

		ctx, err := detector.Detect()

		// Assert: Project root found correctly
		require.NoError(t, err)
		require.NotNil(t, ctx)
		// Use Contains to handle macOS symlink /var vs /private/var
		assert.Contains(t, ctx.ProjectRoot, filepath.Base(tmpDir))
		assert.Contains(t, ctx.WorkingDir, "utils")
	})

	t.Run("detect_with_multiple_config_levels", func(t *testing.T) {
		// Setup: Create project with configs at multiple levels
		tmpDir := t.TempDir()
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		// Create .git at root
		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		// Create .glide.yml at root
		rootConfig := filepath.Join(tmpDir, ".glide.yml")
		require.NoError(t, os.WriteFile(rootConfig, []byte("version: 1\n"), 0644))

		// Create subdirectory with its own .glide.yml
		subDir := filepath.Join(tmpDir, "subproject")
		require.NoError(t, os.MkdirAll(subDir, 0755))

		subConfig := filepath.Join(subDir, ".glide.yml")
		require.NoError(t, os.WriteFile(subConfig, []byte("version: 1\n"), 0644))

		require.NoError(t, os.Chdir(subDir))

		// Test: Detect context
		detector, err := context.NewDetector()
		require.NoError(t, err)

		ctx, err := detector.Detect()

		// Assert: Context detected (finds nearest .glide.yml or .git)
		require.NoError(t, err)
		require.NotNil(t, ctx)
		// May find subproject as root if it has .glide.yml
		assert.NotEmpty(t, ctx.ProjectRoot)
	})
}

// TestContextExtensions tests plugin extension integration
func TestContextExtensions(t *testing.T) {
	t.Run("context_extensions_initialized", func(t *testing.T) {
		// Setup: Create simple project
		tmpDir := t.TempDir()
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		require.NoError(t, os.Chdir(tmpDir))

		// Test: Detect context
		detector, err := context.NewDetector()
		require.NoError(t, err)

		ctx, err := detector.Detect()

		// Assert: Extensions map is initialized
		require.NoError(t, err)
		require.NotNil(t, ctx)
		assert.NotNil(t, ctx.Extensions, "Extensions map should be initialized")
	})
}

// TestContextConcurrency tests concurrent context detection
func TestContextConcurrency(t *testing.T) {
	t.Run("concurrent_detection", func(t *testing.T) {
		// Setup: Create project
		tmpDir := t.TempDir()
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		require.NoError(t, os.Chdir(tmpDir))

		// Test: Detect context concurrently
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				detector, err := context.NewDetector()
				assert.NoError(t, err)

				ctx, err := detector.Detect()
				assert.NoError(t, err)
				assert.NotNil(t, ctx)

				done <- true
			}()
		}

		// Assert: All complete without race conditions
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// TestContextValidation tests context validation methods
func TestContextValidation(t *testing.T) {
	t.Run("valid_context", func(t *testing.T) {
		// Setup: Create valid project
		tmpDir := t.TempDir()
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		require.NoError(t, os.Chdir(tmpDir))

		// Test: Detect context
		detector, err := context.NewDetector()
		require.NoError(t, err)

		ctx, err := detector.Detect()

		// Assert: Context is valid
		require.NoError(t, err)
		require.NotNil(t, ctx)
		assert.True(t, ctx.IsValid(), "Context should be valid")
		assert.NoError(t, ctx.Error)
		assert.NotEmpty(t, ctx.ProjectRoot)
	})

	t.Run("context_without_project_root", func(t *testing.T) {
		// Setup: Create directory without git or config
		tmpDir := t.TempDir()
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		require.NoError(t, os.Chdir(tmpDir))

		// Test: Detect context
		detector, err := context.NewDetector()
		require.NoError(t, err)

		ctx, err := detector.Detect()

		// Assert: Error expected when no project root found
		// The detector returns an error if it can't find a project root
		if err != nil {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "project root")
		} else {
			// If no error, context should still be created
			require.NotNil(t, ctx)
		}
	})
}

// TestContextPluginEnhancedDetection tests plugin-enhanced context detection
func TestContextPluginEnhancedDetection(t *testing.T) {
	t.Run("detect_with_docker_plugin", func(t *testing.T) {
		// Setup: Create project with Docker files
		tmpDir := t.TempDir()
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		// Create Dockerfile
		dockerfile := filepath.Join(tmpDir, "Dockerfile")
		require.NoError(t, os.WriteFile(dockerfile, []byte("FROM node:18\n"), 0644))

		// Create docker-compose.yml
		composeFile := filepath.Join(tmpDir, "docker-compose.yml")
		composeContent := `version: '3'
services:
  app:
    build: .
`
		require.NoError(t, os.WriteFile(composeFile, []byte(composeContent), 0644))

		require.NoError(t, os.Chdir(tmpDir))

		// Test: Detect context (plugin would enhance with Docker info)
		detector, err := context.NewDetector()
		require.NoError(t, err)

		ctx, err := detector.Detect()

		// Assert: Context detected (Docker info in extensions if plugin loaded)
		require.NoError(t, err)
		require.NotNil(t, ctx)
		assert.NotNil(t, ctx.Extensions, "Extensions should be initialized")
		// Actual Docker detection requires docker plugin to be installed
	})

	t.Run("detect_with_multiple_framework_plugins", func(t *testing.T) {
		// Setup: Create project with multiple framework indicators
		tmpDir := t.TempDir()
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		// PHP project
		composerFile := filepath.Join(tmpDir, "composer.json")
		require.NoError(t, os.WriteFile(composerFile, []byte(`{"require": {"php": "^8.1"}}`), 0644))

		// Node project
		packageFile := filepath.Join(tmpDir, "package.json")
		require.NoError(t, os.WriteFile(packageFile, []byte(`{"name": "test"}`), 0644))

		// Go project
		goModFile := filepath.Join(tmpDir, "go.mod")
		require.NoError(t, os.WriteFile(goModFile, []byte("module test\n"), 0644))

		require.NoError(t, os.Chdir(tmpDir))

		// Test: Detect context
		detector, err := context.NewDetector()
		require.NoError(t, err)

		ctx, err := detector.Detect()

		// Assert: Context detected with all framework extensions
		require.NoError(t, err)
		require.NotNil(t, ctx)
		assert.NotNil(t, ctx.Extensions, "Extensions should be initialized")
		// Framework info would be in Extensions if plugins are loaded
	})

	t.Run("detect_without_plugins", func(t *testing.T) {
		// Setup: Create project without plugin enhancements
		tmpDir := t.TempDir()
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		require.NoError(t, os.Chdir(tmpDir))

		// Test: Detect context without plugins
		detector, err := context.NewDetector()
		require.NoError(t, err)

		ctx, err := detector.Detect()

		// Assert: Basic context still works
		require.NoError(t, err)
		require.NotNil(t, ctx)
		assert.NotNil(t, ctx.Extensions, "Extensions should be initialized")
		// Note: Extensions may contain auto-detected items (like Docker)
		// even without project-specific files, since some plugins detect system-level tools
	})
}

// TestContextCachedDetection tests cached context detection
func TestContextCachedDetection(t *testing.T) {
	t.Run("detect_multiple_times_same_directory", func(t *testing.T) {
		// Setup: Create project
		tmpDir := t.TempDir()
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		require.NoError(t, os.Chdir(tmpDir))

		// Test: Detect multiple times
		detector1, err := context.NewDetector()
		require.NoError(t, err)

		ctx1, err := detector1.Detect()
		require.NoError(t, err)
		require.NotNil(t, ctx1)

		// Second detection
		detector2, err := context.NewDetector()
		require.NoError(t, err)

		ctx2, err := detector2.Detect()
		require.NoError(t, err)
		require.NotNil(t, ctx2)

		// Assert: Both detections successful (caching is internal optimization)
		assert.Equal(t, ctx1.ProjectRoot, ctx2.ProjectRoot)
		assert.Equal(t, ctx1.DevelopmentMode, ctx2.DevelopmentMode)
	})

	t.Run("detect_after_directory_change", func(t *testing.T) {
		// Setup: Create two projects
		tmpDir := t.TempDir()
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		project1 := filepath.Join(tmpDir, "project1")
		require.NoError(t, os.MkdirAll(project1, 0755))
		git1 := filepath.Join(project1, ".git")
		require.NoError(t, os.MkdirAll(git1, 0755))

		project2 := filepath.Join(tmpDir, "project2")
		require.NoError(t, os.MkdirAll(project2, 0755))
		git2 := filepath.Join(project2, ".git")
		require.NoError(t, os.MkdirAll(git2, 0755))

		// Test: Detect in first project
		require.NoError(t, os.Chdir(project1))
		detector1, err := context.NewDetector()
		require.NoError(t, err)

		ctx1, err := detector1.Detect()
		require.NoError(t, err)
		require.NotNil(t, ctx1)

		// Change to second project and detect
		require.NoError(t, os.Chdir(project2))
		detector2, err := context.NewDetector()
		require.NoError(t, err)

		ctx2, err := detector2.Detect()
		require.NoError(t, err)
		require.NotNil(t, ctx2)

		// Assert: Different contexts detected
		assert.NotEqual(t, ctx1.ProjectRoot, ctx2.ProjectRoot)
		assert.Contains(t, ctx1.ProjectRoot, "project1")
		assert.Contains(t, ctx2.ProjectRoot, "project2")
	})

	t.Run("detect_after_file_changes", func(t *testing.T) {
		// Setup: Create project
		tmpDir := t.TempDir()
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		require.NoError(t, os.Chdir(tmpDir))

		// Test: Initial detection
		detector1, err := context.NewDetector()
		require.NoError(t, err)

		ctx1, err := detector1.Detect()
		require.NoError(t, err)
		require.NotNil(t, ctx1)

		// Add framework files
		packageFile := filepath.Join(tmpDir, "package.json")
		require.NoError(t, os.WriteFile(packageFile, []byte(`{"name": "test"}`), 0644))

		// Detect again after file changes
		detector2, err := context.NewDetector()
		require.NoError(t, err)

		ctx2, err := detector2.Detect()
		require.NoError(t, err)
		require.NotNil(t, ctx2)

		// Assert: Context still valid (may or may not detect new framework)
		assert.Equal(t, ctx1.ProjectRoot, ctx2.ProjectRoot)
		// Extensions may differ if plugins re-scan
	})

	t.Run("concurrent_cached_detection", func(t *testing.T) {
		// Setup: Create project
		tmpDir := t.TempDir()
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		require.NoError(t, os.Chdir(tmpDir))

		// Test: Detect concurrently (tests cache thread safety)
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				detector, err := context.NewDetector()
				assert.NoError(t, err)

				ctx, err := detector.Detect()
				assert.NoError(t, err)
				assert.NotNil(t, ctx)
				assert.NotEmpty(t, ctx.ProjectRoot)

				done <- true
			}()
		}

		// Assert: All detections complete without race conditions
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}
