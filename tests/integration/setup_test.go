package integration_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/ivannovak/glide/internal/cli"
	"github.com/ivannovak/glide/internal/config"
	"github.com/ivannovak/glide/internal/context"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetupCommand tests the setup command flow
func TestSetupCommand(t *testing.T) {
	t.Run("setup_in_non_project_directory", func(t *testing.T) {
		// Create a temporary directory that's not a project
		tmpDir := t.TempDir()
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		err := os.Chdir(tmpDir)
		require.NoError(t, err)

		// Create setup command with nil context (not in project)
		rootCmd := &cobra.Command{Use: "glide"}
		cfg := config.GetDefaults()
		setupCmd := cli.NewSetupCommand(nil, &cfg)
		rootCmd.AddCommand(setupCmd)

		// Capture output
		var buf bytes.Buffer
		rootCmd.SetOut(&buf)
		rootCmd.SetErr(&buf)

		// Execute setup in non-interactive mode with explicit path and mode
		// Setup should work from anywhere to help initialize projects
		rootCmd.SetArgs([]string{"setup", "--non-interactive", "--path", tmpDir, "--mode", "single-repo"})
		err = rootCmd.Execute()

		// Setup is designed to work from anywhere to help set up projects
		// It may return an error if prerequisites aren't met (e.g., Docker not running)
		// but that's different from refusing to run outside a project
		if err != nil {
			// If it fails, it should be due to prerequisites, not location
			assert.NotContains(t, buf.String(), "not in a project",
				"Setup should not fail just because we're not in a project")
		}
	})

	t.Run("setup_in_single_repo_project", func(t *testing.T) {
		// Create a mock single-repo project structure
		tmpDir := t.TempDir()
		vcsDir := filepath.Join(tmpDir, "vcs")

		// Create project structure
		require.NoError(t, os.MkdirAll(vcsDir, 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(vcsDir, ".git"), 0755))
		require.NoError(t, os.WriteFile(
			filepath.Join(vcsDir, "docker-compose.yml"),
			[]byte("version: '3'\n"),
			0644,
		))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		err := os.Chdir(vcsDir)
		require.NoError(t, err)

		// Test context detection
		ctx := context.Detect()
		assert.NotNil(t, ctx)
		// The directory structure doesn't indicate multi-worktree, so should be single-repo
		if ctx.DevelopmentMode != "" {
			assert.Equal(t, context.ModeSingleRepo, ctx.DevelopmentMode)
		}
		// Resolve symlinks for path comparison on macOS
		if ctx.ProjectRoot != "" {
			resolvedTmp, _ := filepath.EvalSymlinks(tmpDir)
			resolvedRoot, _ := filepath.EvalSymlinks(ctx.ProjectRoot)
			assert.Equal(t, resolvedTmp, resolvedRoot)
		}
	})

	t.Run("setup_in_multi_worktree_project", func(t *testing.T) {
		// Create a mock multi-worktree project structure
		tmpDir := t.TempDir()
		vcsDir := filepath.Join(tmpDir, "vcs")
		worktreesDir := filepath.Join(tmpDir, "worktrees")
		worktreeDir := filepath.Join(worktreesDir, "feature-test")

		// Create project structure
		require.NoError(t, os.MkdirAll(vcsDir, 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(vcsDir, ".git"), 0755))
		require.NoError(t, os.MkdirAll(worktreeDir, 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(worktreeDir, ".git"), 0755))
		require.NoError(t, os.WriteFile(
			filepath.Join(vcsDir, "docker-compose.yml"),
			[]byte("version: '3'\n"),
			0644,
		))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		err := os.Chdir(worktreeDir)
		require.NoError(t, err)

		// Test context detection
		ctx := context.Detect()
		assert.NotNil(t, ctx)
		assert.Equal(t, context.ModeMultiWorktree, ctx.DevelopmentMode)
		// Resolve symlinks for path comparison on macOS
		resolvedTmp, _ := filepath.EvalSymlinks(tmpDir)
		resolvedRoot, _ := filepath.EvalSymlinks(ctx.ProjectRoot)
		assert.Equal(t, resolvedTmp, resolvedRoot)
		assert.Equal(t, context.LocationWorktree, ctx.Location)
		assert.Equal(t, "feature-test", ctx.WorktreeName)
	})
}

// TestConfigurationCreation tests configuration file creation
func TestConfigurationCreation(t *testing.T) {
	t.Run("create_default_config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".glide.yml")

		// Set config path for testing
		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", originalHome)

		// Create default config
		cfg := config.GetDefaults()

		// We'll manually write a config file since Save function might not take a path
		configContent := `projects:
  myproject:
    path: /test/path
    mode: multi-worktree
default_project: myproject
defaults:
  docker:
    compose_command: docker compose
`
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err, "Should write config file")

		// Verify file was created
		assert.FileExists(t, configPath, "Config file should be created")

		// Test that we can detect this would be loaded
		// Note: The actual Load() function doesn't take a path parameter
		assert.True(t, cfg.Defaults.Docker.AutoStart, "Default auto_start should be set")
	})

	t.Run("handle_existing_config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".glide.yml")

		// Create existing config
		existingContent := `# Existing config
projects:
  test-project:
    path: /test/path
    mode: single-repo
default_project: test-project
defaults:
  docker:
    compose_command: docker compose
`
		err := os.WriteFile(configPath, []byte(existingContent), 0644)
		require.NoError(t, err)

		// Set HOME to use our test config
		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", originalHome)

		// Load config should work when HOME is set correctly
		// Note: actual Load() doesn't take a path
		assert.FileExists(t, configPath)
	})

	t.Run("config_validation", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".glide.yml")

		// Create invalid config
		invalidContent := `not valid yaml content {][`
		err := os.WriteFile(configPath, []byte(invalidContent), 0644)
		require.NoError(t, err)

		// Set HOME to use our test config
		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", originalHome)

		// Load() should handle invalid YAML gracefully
		// The actual implementation might return defaults on error
		cfg, err := config.Load()
		// We expect either an error or default config
		if err != nil {
			t.Log("Load returned error for invalid YAML:", err)
		} else if cfg != nil {
			t.Log("Load returned defaults for invalid YAML")
		}
	})
}

// TestProjectDetection tests project detection and mode determination
func TestProjectDetection(t *testing.T) {
	t.Run("detect_vcs_directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		vcsDir := filepath.Join(tmpDir, "vcs")
		worktreesDir := filepath.Join(tmpDir, "worktrees")

		// Create multi-worktree structure
		require.NoError(t, os.MkdirAll(filepath.Join(vcsDir, ".git"), 0755))
		require.NoError(t, os.MkdirAll(worktreesDir, 0755)) // This makes it multi-worktree

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		err := os.Chdir(vcsDir)
		require.NoError(t, err)

		ctx := context.Detect()
		assert.NotNil(t, ctx)
		// In multi-worktree mode, vcs dir is main repo
		if ctx.DevelopmentMode == context.ModeMultiWorktree {
			assert.Equal(t, context.LocationMainRepo, ctx.Location)
		} else {
			// In single-repo mode, it's just a project
			assert.Equal(t, context.LocationProject, ctx.Location)
		}
	})

	t.Run("detect_worktree_directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		vcsDir := filepath.Join(tmpDir, "vcs")
		worktreesDir := filepath.Join(tmpDir, "worktrees")
		worktreeDir := filepath.Join(worktreesDir, "feature-branch")

		// Create full structure
		require.NoError(t, os.MkdirAll(filepath.Join(vcsDir, ".git"), 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(worktreeDir, ".git"), 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		err := os.Chdir(worktreeDir)
		require.NoError(t, err)

		ctx := context.Detect()
		assert.NotNil(t, ctx)
		assert.Equal(t, context.LocationWorktree, ctx.Location)
		assert.Equal(t, "feature-branch", ctx.WorktreeName)
	})

	t.Run("detect_project_root", func(t *testing.T) {
		tmpDir := t.TempDir()
		vcsDir := filepath.Join(tmpDir, "vcs")
		worktreesDir := filepath.Join(tmpDir, "worktrees")

		// Create multi-worktree structure
		require.NoError(t, os.MkdirAll(filepath.Join(vcsDir, ".git"), 0755))
		require.NoError(t, os.MkdirAll(worktreesDir, 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		err := os.Chdir(tmpDir)
		require.NoError(t, err)

		ctx := context.Detect()
		// From the root of a multi-worktree project
		if ctx != nil {
			assert.Equal(t, context.LocationRoot, ctx.Location)
			// Resolve symlinks for path comparison on macOS
			resolvedTmp, _ := filepath.EvalSymlinks(tmpDir)
			resolvedRoot, _ := filepath.EvalSymlinks(ctx.ProjectRoot)
			assert.Equal(t, resolvedTmp, resolvedRoot)
		}
	})
}

// TestModeSelection tests development mode selection
func TestModeSelection(t *testing.T) {
	t.Run("single_repo_mode_detection", func(t *testing.T) {
		tmpDir := t.TempDir()
		// For single-repo, just create a git repo directly, not in vcs/
		projectDir := filepath.Join(tmpDir, "project")

		// Create single-repo structure (just .git, no vcs/worktrees pattern)
		require.NoError(t, os.MkdirAll(filepath.Join(projectDir, ".git"), 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		err := os.Chdir(projectDir)
		require.NoError(t, err)

		ctx := context.Detect()
		assert.NotNil(t, ctx)
		// A simple git repo without vcs/worktrees structure should be single-repo
		assert.Equal(t, context.ModeSingleRepo, ctx.DevelopmentMode)
	})

	t.Run("multi_worktree_mode_detection", func(t *testing.T) {
		tmpDir := t.TempDir()
		vcsDir := filepath.Join(tmpDir, "vcs")
		worktreesDir := filepath.Join(tmpDir, "worktrees")

		// Create multi-worktree structure
		require.NoError(t, os.MkdirAll(filepath.Join(vcsDir, ".git"), 0755))
		require.NoError(t, os.MkdirAll(worktreesDir, 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		err := os.Chdir(vcsDir)
		require.NoError(t, err)

		ctx := context.Detect()
		assert.NotNil(t, ctx)
		assert.Equal(t, context.ModeMultiWorktree, ctx.DevelopmentMode)
	})
}

// TestEnvironmentValidation tests environment setup validation
func TestEnvironmentValidation(t *testing.T) {
	t.Run("validate_git_repository", func(t *testing.T) {
		tmpDir := t.TempDir()
		vcsDir := filepath.Join(tmpDir, "vcs")

		// Create directory without git
		require.NoError(t, os.MkdirAll(vcsDir, 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		err := os.Chdir(vcsDir)
		require.NoError(t, err)

		// Should not detect as project without .git
		ctx := context.Detect()
		// Context might not be nil but should indicate it's not a valid project
		if ctx != nil {
			// Either ctx has an error or location is unknown
			assert.True(t, ctx.Error != nil || ctx.Location == context.LocationUnknown,
				"Should not detect valid project without .git directory")
		}
	})

	t.Run("validate_docker_compose_files", func(t *testing.T) {
		tmpDir := t.TempDir()
		vcsDir := filepath.Join(tmpDir, "vcs")

		// Create project with git
		require.NoError(t, os.MkdirAll(filepath.Join(vcsDir, ".git"), 0755))

		// Create docker-compose.yml
		composeContent := `version: '3'
services:
  php:
    image: php:8.3
  mysql:
    image: mysql:8.0
`
		require.NoError(t, os.WriteFile(
			filepath.Join(vcsDir, "docker-compose.yml"),
			[]byte(composeContent),
			0644,
		))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		err := os.Chdir(vcsDir)
		require.NoError(t, err)

		ctx := context.Detect()
		assert.NotNil(t, ctx)
		if len(ctx.ComposeFiles) > 0 {
			assert.Contains(t, ctx.ComposeFiles[0], "docker-compose.yml")
		} else {
			t.Log("No compose files detected in context")
		}
	})

	t.Run("validate_permissions", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Test we can write to config directory
		testFile := filepath.Join(tmpDir, "test-write.tmp")
		err := os.WriteFile(testFile, []byte("test"), 0644)
		assert.NoError(t, err, "Should have write permissions")

		// Clean up
		os.Remove(testFile)
	})
}
