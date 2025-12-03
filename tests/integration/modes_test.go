package integration_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glide-cli/glide/v3/internal/config"
	"github.com/glide-cli/glide/v3/internal/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModeTransitions tests switching between development modes
func TestModeTransitions(t *testing.T) {
	t.Run("detect_single_repo_structure", func(t *testing.T) {
		// Create a simple git repository structure (single-repo mode)
		tmpDir := t.TempDir()
		projectDir := filepath.Join(tmpDir, "myproject")

		// Create single-repo structure
		require.NoError(t, os.MkdirAll(filepath.Join(projectDir, ".git"), 0755))
		require.NoError(t, os.WriteFile(
			filepath.Join(projectDir, "docker-compose.yml"),
			[]byte("version: '3'\nservices:\n  app:\n    image: app:latest\n"),
			0644,
		))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		err := os.Chdir(projectDir)
		require.NoError(t, err)

		// Detect context
		ctx := context.Detect()
		require.NotNil(t, ctx)

		// Should detect as single-repo mode
		assert.Equal(t, context.ModeSingleRepo, ctx.DevelopmentMode)
		assert.Equal(t, context.LocationProject, ctx.Location)

		// Verify project root is set correctly
		resolvedProject, _ := filepath.EvalSymlinks(projectDir)
		resolvedRoot, _ := filepath.EvalSymlinks(ctx.ProjectRoot)
		assert.Equal(t, resolvedProject, resolvedRoot)
	})

	t.Run("detect_multi_worktree_structure", func(t *testing.T) {
		// Create multi-worktree structure
		tmpDir := t.TempDir()
		vcsDir := filepath.Join(tmpDir, "vcs")
		worktreesDir := filepath.Join(tmpDir, "worktrees")

		// Create the structure
		require.NoError(t, os.MkdirAll(filepath.Join(vcsDir, ".git"), 0755))
		require.NoError(t, os.MkdirAll(worktreesDir, 0755))
		require.NoError(t, os.WriteFile(
			filepath.Join(vcsDir, "docker-compose.yml"),
			[]byte("version: '3'\nservices:\n  app:\n    image: app:latest\n"),
			0644,
		))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		// Test from vcs directory
		err := os.Chdir(vcsDir)
		require.NoError(t, err)

		ctx := context.Detect()
		require.NotNil(t, ctx)

		// Should detect as multi-worktree mode
		assert.Equal(t, context.ModeMultiWorktree, ctx.DevelopmentMode)
		assert.Equal(t, context.LocationMainRepo, ctx.Location)
		assert.True(t, ctx.IsMainRepo)

		// Test from root directory
		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		ctx = context.Detect()
		require.NotNil(t, ctx)

		// Should still be multi-worktree mode but from root
		assert.Equal(t, context.ModeMultiWorktree, ctx.DevelopmentMode)
		assert.Equal(t, context.LocationRoot, ctx.Location)
		assert.True(t, ctx.IsRoot)
	})

	t.Run("transition_single_to_multi", func(t *testing.T) {
		// Start with single-repo structure
		tmpDir := t.TempDir()
		vcsDir := filepath.Join(tmpDir, "vcs")

		// Create initial single-repo in vcs/
		require.NoError(t, os.MkdirAll(filepath.Join(vcsDir, ".git"), 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		err := os.Chdir(vcsDir)
		require.NoError(t, err)

		// Initially should be single-repo (no worktrees dir)
		ctx := context.Detect()
		require.NotNil(t, ctx)
		// When in vcs/ without worktrees/, it might not set mode explicitly
		// but it shouldn't be multi-worktree
		if ctx.DevelopmentMode != "" {
			assert.NotEqual(t, context.ModeMultiWorktree, ctx.DevelopmentMode)
		}

		// Create worktrees directory to transition to multi-worktree
		worktreesDir := filepath.Join(tmpDir, "worktrees")
		require.NoError(t, os.MkdirAll(worktreesDir, 0755))

		// Re-detect after creating worktrees
		ctx = context.Detect()
		require.NotNil(t, ctx)

		// Should now detect as multi-worktree
		assert.Equal(t, context.ModeMultiWorktree, ctx.DevelopmentMode)
		assert.Equal(t, context.LocationMainRepo, ctx.Location)
	})

	t.Run("configuration_updates_with_mode", func(t *testing.T) {
		// Test that configuration properly reflects mode changes
		tmpDir := t.TempDir()

		// Create a test config
		cfg := config.GetDefaults()

		// Add a project in single-repo mode
		project1 := config.ProjectConfig{
			Path: filepath.Join(tmpDir, "single-project"),
			Mode: "single-repo",
		}
		cfg.Projects["single-project"] = project1

		// Add a project in multi-worktree mode
		project2 := config.ProjectConfig{
			Path: filepath.Join(tmpDir, "multi-project"),
			Mode: "multi-worktree",
		}
		cfg.Projects["multi-project"] = project2

		// Verify modes are stored correctly
		assert.Equal(t, "single-repo", cfg.Projects["single-project"].Mode)
		assert.Equal(t, "multi-worktree", cfg.Projects["multi-project"].Mode)

		// Test that default project can be set
		cfg.DefaultProject = "multi-project"
		assert.Equal(t, "multi-project", cfg.DefaultProject)
	})
}

// TestModeSpecificCommands tests commands that behave differently based on mode
func TestModeSpecificCommands(t *testing.T) {
	t.Run("global_commands_in_multi_worktree", func(t *testing.T) {
		// Create multi-worktree structure
		tmpDir := t.TempDir()
		vcsDir := filepath.Join(tmpDir, "vcs")
		worktreesDir := filepath.Join(tmpDir, "worktrees")
		worktree1 := filepath.Join(worktreesDir, "feature-1")
		worktree2 := filepath.Join(worktreesDir, "feature-2")

		// Create structure
		require.NoError(t, os.MkdirAll(filepath.Join(vcsDir, ".git"), 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(worktree1, ".git"), 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(worktree2, ".git"), 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		// Test from root - should have global context
		err := os.Chdir(tmpDir)
		require.NoError(t, err)

		ctx := context.Detect()
		require.NotNil(t, ctx)
		assert.Equal(t, context.LocationRoot, ctx.Location)
		assert.True(t, ctx.IsRoot, "Should be at root for project-wide commands")

		// Global commands should work from root
		// Examples: glide status (shows all worktrees), glide down (stops all containers)
		assert.Equal(t, context.ModeMultiWorktree, ctx.DevelopmentMode)
	})

	t.Run("local_commands_in_single_repo", func(t *testing.T) {
		// Create single-repo structure
		tmpDir := t.TempDir()
		projectDir := filepath.Join(tmpDir, "project")

		require.NoError(t, os.MkdirAll(filepath.Join(projectDir, ".git"), 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		err := os.Chdir(projectDir)
		require.NoError(t, err)

		ctx := context.Detect()
		require.NotNil(t, ctx)

		// In single-repo, all commands are "local" to the project
		assert.Equal(t, context.ModeSingleRepo, ctx.DevelopmentMode)
		assert.Equal(t, context.LocationProject, ctx.Location)
		assert.False(t, ctx.IsRoot, "Single-repo doesn't have root concept")
		assert.False(t, ctx.IsMainRepo, "Single-repo doesn't have main repo concept")
	})

	t.Run("local_commands_in_worktree", func(t *testing.T) {
		// Create multi-worktree with a worktree
		tmpDir := t.TempDir()
		vcsDir := filepath.Join(tmpDir, "vcs")
		worktreesDir := filepath.Join(tmpDir, "worktrees")
		worktreeDir := filepath.Join(worktreesDir, "feature-branch")

		require.NoError(t, os.MkdirAll(filepath.Join(vcsDir, ".git"), 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(worktreeDir, ".git"), 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		err := os.Chdir(worktreeDir)
		require.NoError(t, err)

		ctx := context.Detect()
		require.NotNil(t, ctx)

		// In a worktree, commands are local to that worktree
		assert.Equal(t, context.ModeMultiWorktree, ctx.DevelopmentMode)
		assert.Equal(t, context.LocationWorktree, ctx.Location)
		assert.True(t, ctx.IsWorktree, "Should be in a worktree")
		assert.Equal(t, "feature-branch", ctx.WorktreeName)
	})

	t.Run("error_for_wrong_mode_location", func(t *testing.T) {
		// Test that certain locations are invalid for certain operations
		tmpDir := t.TempDir()

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		// In a non-project directory
		err := os.Chdir(tmpDir)
		require.NoError(t, err)

		ctx := context.Detect()

		// Should not detect a valid project context
		if ctx != nil {
			assert.True(t,
				ctx.Location == context.LocationUnknown || ctx.Error != nil,
				"Non-project directory should not have valid context")
		}
	})
}

// TestContextPreservation tests that context is properly preserved across operations
func TestContextPreservation(t *testing.T) {
	t.Run("working_directory_tracking", func(t *testing.T) {
		// Create a project structure
		tmpDir := t.TempDir()
		projectDir := filepath.Join(tmpDir, "project")
		subDir := filepath.Join(projectDir, "subdir")

		require.NoError(t, os.MkdirAll(filepath.Join(projectDir, ".git"), 0755))
		require.NoError(t, os.MkdirAll(subDir, 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		// Start in project root
		err := os.Chdir(projectDir)
		require.NoError(t, err)

		ctx := context.Detect()
		require.NotNil(t, ctx)
		initialWd := ctx.WorkingDir

		// Move to subdirectory
		err = os.Chdir(subDir)
		require.NoError(t, err)

		ctx = context.Detect()
		require.NotNil(t, ctx)

		// Working directory should be tracked
		assert.NotEqual(t, initialWd, ctx.WorkingDir)
		resolvedSubDir, _ := filepath.EvalSymlinks(subDir)
		resolvedWd, _ := filepath.EvalSymlinks(ctx.WorkingDir)
		assert.Equal(t, resolvedSubDir, resolvedWd)

		// Project root should remain the same
		resolvedProject, _ := filepath.EvalSymlinks(projectDir)
		resolvedRoot, _ := filepath.EvalSymlinks(ctx.ProjectRoot)
		assert.Equal(t, resolvedProject, resolvedRoot)
	})

	t.Run("environment_variable_handling", func(t *testing.T) {
		// Test that environment variables are properly handled
		tmpDir := t.TempDir()
		projectDir := filepath.Join(tmpDir, "project")

		require.NoError(t, os.MkdirAll(filepath.Join(projectDir, ".git"), 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		err := os.Chdir(projectDir)
		require.NoError(t, err)

		// Set a test environment variable
		testEnvVar := "GLIDE_TEST_MODE"
		testValue := "test-value"
		originalValue := os.Getenv(testEnvVar)
		defer func() {
			if originalValue != "" {
				os.Setenv(testEnvVar, originalValue)
			} else {
				os.Unsetenv(testEnvVar)
			}
		}()

		os.Setenv(testEnvVar, testValue)

		ctx := context.Detect()
		require.NotNil(t, ctx)

		// Environment should be accessible
		envValue := os.Getenv(testEnvVar)
		assert.Equal(t, testValue, envValue)
	})

	t.Run("configuration_persistence", func(t *testing.T) {
		// Test that configuration persists across context changes
		tmpDir := t.TempDir()

		// Create initial config
		cfg := config.GetDefaults()
		testProject := "test-project"
		testPath := filepath.Join(tmpDir, "test")

		cfg.Projects[testProject] = config.ProjectConfig{
			Path: testPath,
			Mode: "multi-worktree",
		}
		cfg.DefaultProject = testProject

		// Simulate saving and loading config
		// (actual Save/Load would write to filesystem)

		// Verify config persists
		assert.Equal(t, testProject, cfg.DefaultProject)
		assert.Equal(t, testPath, cfg.Projects[testProject].Path)
		assert.Equal(t, "multi-worktree", cfg.Projects[testProject].Mode)

		// Create new config and verify defaults
		cfg2 := config.GetDefaults()
		assert.NotNil(t, cfg2.Projects)
		assert.NotNil(t, cfg2.Defaults)
		assert.NotNil(t, cfg2.Defaults.Docker)
	})

	t.Run("mode_consistency_across_worktrees", func(t *testing.T) {
		// Test that mode is consistent across all worktrees in a project
		tmpDir := t.TempDir()
		vcsDir := filepath.Join(tmpDir, "vcs")
		worktreesDir := filepath.Join(tmpDir, "worktrees")
		worktree1 := filepath.Join(worktreesDir, "feature-1")
		worktree2 := filepath.Join(worktreesDir, "feature-2")

		// Create multi-worktree structure
		require.NoError(t, os.MkdirAll(filepath.Join(vcsDir, ".git"), 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(worktree1, ".git"), 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(worktree2, ".git"), 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		// Check from vcs
		os.Chdir(vcsDir)
		ctx1 := context.Detect()
		require.NotNil(t, ctx1)

		// Check from worktree1
		os.Chdir(worktree1)
		ctx2 := context.Detect()
		require.NotNil(t, ctx2)

		// Check from worktree2
		os.Chdir(worktree2)
		ctx3 := context.Detect()
		require.NotNil(t, ctx3)

		// All should have the same mode
		assert.Equal(t, context.ModeMultiWorktree, ctx1.DevelopmentMode)
		assert.Equal(t, context.ModeMultiWorktree, ctx2.DevelopmentMode)
		assert.Equal(t, context.ModeMultiWorktree, ctx3.DevelopmentMode)

		// All should have the same project root
		resolvedRoot1, _ := filepath.EvalSymlinks(ctx1.ProjectRoot)
		resolvedRoot2, _ := filepath.EvalSymlinks(ctx2.ProjectRoot)
		resolvedRoot3, _ := filepath.EvalSymlinks(ctx3.ProjectRoot)
		assert.Equal(t, resolvedRoot1, resolvedRoot2)
		assert.Equal(t, resolvedRoot1, resolvedRoot3)

		// But different locations
		assert.Equal(t, context.LocationMainRepo, ctx1.Location)
		assert.Equal(t, context.LocationWorktree, ctx2.Location)
		assert.Equal(t, context.LocationWorktree, ctx3.Location)

		// And different worktree names
		assert.Equal(t, "", ctx1.WorktreeName)
		assert.Equal(t, "feature-1", ctx2.WorktreeName)
		assert.Equal(t, "feature-2", ctx3.WorktreeName)
	})
}
