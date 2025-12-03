package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glide-cli/glide/v3/internal/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWorktreeManagement tests git worktree operations
func TestWorktreeManagement(t *testing.T) {
	// Skip in short mode (pre-commit hooks)
	if testing.Short() {
		t.Skip("Skipping worktree tests in short mode")
	}

	// Check if git is available
	if err := exec.Command("git", "--version").Run(); err != nil {
		t.Skip("Git is not available")
		return
	}

	t.Run("worktree_creation", func(t *testing.T) {
		// Create a temporary git repository
		tmpDir := t.TempDir()
		vcsDir := filepath.Join(tmpDir, "vcs")
		worktreesDir := filepath.Join(tmpDir, "worktrees")

		// Initialize git repo
		require.NoError(t, os.MkdirAll(vcsDir, 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		err := os.Chdir(vcsDir)
		require.NoError(t, err)

		// Initialize repo
		cmd := exec.Command("git", "init")
		err = cmd.Run()
		require.NoError(t, err)

		// Configure git
		exec.Command("git", "config", "user.email", "test@example.com").Run()
		exec.Command("git", "config", "user.name", "Test User").Run()

		// Create initial commit
		require.NoError(t, os.WriteFile("README.md", []byte("# Test Project"), 0644))
		exec.Command("git", "add", ".").Run()
		exec.Command("git", "commit", "-m", "Initial commit").Run()

		// Create worktrees directory
		require.NoError(t, os.MkdirAll(worktreesDir, 0755))

		// Create a worktree
		worktreePath := filepath.Join(worktreesDir, "feature-test")
		cmd = exec.Command("git", "worktree", "add", worktreePath, "-b", "feature-test")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "Failed to create worktree: %s", string(output))

		// Verify worktree was created
		assert.DirExists(t, worktreePath)
		assert.FileExists(t, filepath.Join(worktreePath, "README.md"))

		// List worktrees
		cmd = exec.Command("git", "worktree", "list")
		output, err = cmd.CombinedOutput()
		require.NoError(t, err)
		assert.Contains(t, string(output), "feature-test")
	})

	t.Run("worktree_detection", func(t *testing.T) {
		tmpDir := t.TempDir()
		vcsDir := filepath.Join(tmpDir, "vcs")
		worktreesDir := filepath.Join(tmpDir, "worktrees")
		worktreeDir := filepath.Join(worktreesDir, "feature-branch")

		// Create structure
		require.NoError(t, os.MkdirAll(vcsDir, 0755))
		require.NoError(t, os.MkdirAll(worktreeDir, 0755))

		// Initialize main repo
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		os.Chdir(vcsDir)
		exec.Command("git", "init").Run()
		exec.Command("git", "config", "user.email", "test@example.com").Run()
		exec.Command("git", "config", "user.name", "Test User").Run()
		os.WriteFile("test.txt", []byte("test"), 0644)
		exec.Command("git", "add", ".").Run()
		exec.Command("git", "commit", "-m", "Initial").Run()

		// Add worktree
		exec.Command("git", "worktree", "add", worktreeDir, "-b", "feature-branch").Run()

		// Change to worktree and detect context
		os.Chdir(worktreeDir)

		ctx := context.Detect()
		assert.NotNil(t, ctx)
		assert.Equal(t, context.ModeMultiWorktree, ctx.DevelopmentMode)
		assert.Equal(t, context.LocationWorktree, ctx.Location)
		assert.Equal(t, "feature-branch", ctx.WorktreeName)
	})

	t.Run("worktree_isolation", func(t *testing.T) {
		tmpDir := t.TempDir()
		vcsDir := filepath.Join(tmpDir, "vcs")
		worktreesDir := filepath.Join(tmpDir, "worktrees")
		worktree1 := filepath.Join(worktreesDir, "feature-1")
		worktree2 := filepath.Join(worktreesDir, "feature-2")

		// Setup main repo
		require.NoError(t, os.MkdirAll(vcsDir, 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		os.Chdir(vcsDir)
		exec.Command("git", "init").Run()
		exec.Command("git", "config", "user.email", "test@example.com").Run()
		exec.Command("git", "config", "user.name", "Test User").Run()
		os.WriteFile("README.md", []byte("# Test"), 0644)
		exec.Command("git", "add", ".").Run()
		exec.Command("git", "commit", "-m", "Initial").Run()

		// Create worktrees directory
		require.NoError(t, os.MkdirAll(worktreesDir, 0755))

		// Create two worktrees
		exec.Command("git", "worktree", "add", worktree1, "-b", "feature-1").Run()
		exec.Command("git", "worktree", "add", worktree2, "-b", "feature-2").Run()

		// Make changes in worktree1
		os.Chdir(worktree1)
		os.WriteFile("feature1.txt", []byte("Feature 1"), 0644)
		exec.Command("git", "add", ".").Run()
		exec.Command("git", "commit", "-m", "Feature 1").Run()

		// Make different changes in worktree2
		os.Chdir(worktree2)
		os.WriteFile("feature2.txt", []byte("Feature 2"), 0644)
		exec.Command("git", "add", ".").Run()
		exec.Command("git", "commit", "-m", "Feature 2").Run()

		// Verify isolation - feature1.txt should not exist in worktree2
		assert.NoFileExists(t, filepath.Join(worktree2, "feature1.txt"))
		assert.FileExists(t, filepath.Join(worktree2, "feature2.txt"))

		// And vice versa
		assert.FileExists(t, filepath.Join(worktree1, "feature1.txt"))
		assert.NoFileExists(t, filepath.Join(worktree1, "feature2.txt"))
	})
}

// TestWorktreeOperations tests worktree-specific operations
func TestWorktreeOperations(t *testing.T) {
	// Skip in short mode (pre-commit hooks)
	if testing.Short() {
		t.Skip("Skipping worktree tests in short mode")
	}

	// Check if git supports worktrees
	cmd := exec.Command("git", "worktree", "list")
	if err := cmd.Run(); err != nil {
		t.Skip("Git worktree support not available")
		return
	}

	t.Run("worktree_branch_management", func(t *testing.T) {
		tmpDir := t.TempDir()
		vcsDir := filepath.Join(tmpDir, "vcs")
		worktreesDir := filepath.Join(tmpDir, "worktrees")

		// Setup
		require.NoError(t, os.MkdirAll(vcsDir, 0755))
		require.NoError(t, os.MkdirAll(worktreesDir, 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		os.Chdir(vcsDir)
		exec.Command("git", "init").Run()
		exec.Command("git", "config", "user.email", "test@example.com").Run()
		exec.Command("git", "config", "user.name", "Test User").Run()
		os.WriteFile("main.txt", []byte("main"), 0644)
		exec.Command("git", "add", ".").Run()
		exec.Command("git", "commit", "-m", "Initial").Run()

		// Create worktree with new branch
		worktreePath := filepath.Join(worktreesDir, "new-feature")
		cmd := exec.Command("git", "worktree", "add", worktreePath, "-b", "new-feature")
		err := cmd.Run()
		require.NoError(t, err)

		// Switch to worktree and verify branch
		os.Chdir(worktreePath)
		cmd = exec.Command("git", "branch", "--show-current")
		output, err := cmd.Output()
		require.NoError(t, err)
		assert.Equal(t, "new-feature", strings.TrimSpace(string(output)))
	})

	t.Run("worktree_removal", func(t *testing.T) {
		tmpDir := t.TempDir()
		vcsDir := filepath.Join(tmpDir, "vcs")
		worktreesDir := filepath.Join(tmpDir, "worktrees")
		worktreePath := filepath.Join(worktreesDir, "temp-feature")

		// Setup
		require.NoError(t, os.MkdirAll(vcsDir, 0755))
		require.NoError(t, os.MkdirAll(worktreesDir, 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		os.Chdir(vcsDir)
		exec.Command("git", "init").Run()
		exec.Command("git", "config", "user.email", "test@example.com").Run()
		exec.Command("git", "config", "user.name", "Test User").Run()
		os.WriteFile("test.txt", []byte("test"), 0644)
		exec.Command("git", "add", ".").Run()
		exec.Command("git", "commit", "-m", "Initial").Run()

		// Create worktree
		exec.Command("git", "worktree", "add", worktreePath, "-b", "temp-feature").Run()
		assert.DirExists(t, worktreePath)

		// Remove worktree
		cmd := exec.Command("git", "worktree", "remove", worktreePath)
		err := cmd.Run()
		require.NoError(t, err)

		// Verify removal
		assert.NoDirExists(t, worktreePath)

		// Verify it's not in worktree list
		cmd = exec.Command("git", "worktree", "list")
		output, err := cmd.Output()
		require.NoError(t, err)
		assert.NotContains(t, string(output), "temp-feature")
	})

	t.Run("worktree_listing", func(t *testing.T) {
		tmpDir := t.TempDir()
		vcsDir := filepath.Join(tmpDir, "vcs")
		worktreesDir := filepath.Join(tmpDir, "worktrees")

		// Setup
		require.NoError(t, os.MkdirAll(vcsDir, 0755))
		require.NoError(t, os.MkdirAll(worktreesDir, 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		os.Chdir(vcsDir)
		exec.Command("git", "init").Run()
		exec.Command("git", "config", "user.email", "test@example.com").Run()
		exec.Command("git", "config", "user.name", "Test User").Run()
		os.WriteFile("test.txt", []byte("test"), 0644)
		exec.Command("git", "add", ".").Run()
		exec.Command("git", "commit", "-m", "Initial").Run()

		// Create multiple worktrees
		worktree1 := filepath.Join(worktreesDir, "feature-a")
		worktree2 := filepath.Join(worktreesDir, "feature-b")
		worktree3 := filepath.Join(worktreesDir, "bugfix-123")

		exec.Command("git", "worktree", "add", worktree1, "-b", "feature-a").Run()
		exec.Command("git", "worktree", "add", worktree2, "-b", "feature-b").Run()
		exec.Command("git", "worktree", "add", worktree3, "-b", "bugfix-123").Run()

		// List worktrees
		cmd := exec.Command("git", "worktree", "list")
		output, err := cmd.Output()
		require.NoError(t, err)

		outputStr := string(output)
		assert.Contains(t, outputStr, "feature-a")
		assert.Contains(t, outputStr, "feature-b")
		assert.Contains(t, outputStr, "bugfix-123")

		// Count worktrees (including main)
		lines := strings.Split(strings.TrimSpace(outputStr), "\n")
		assert.Equal(t, 4, len(lines), "Should have 4 worktrees (main + 3 features)")
	})
}

// TestWorktreeWithDocker tests worktree operations with Docker context
func TestWorktreeWithDocker(t *testing.T) {
	// Skip in short mode (pre-commit hooks)
	if testing.Short() {
		t.Skip("Skipping worktree tests in short mode")
	}

	// Skip if Docker is not available
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skip("Docker is not available")
		return
	}

	t.Run("worktree_docker_isolation", func(t *testing.T) {
		tmpDir := t.TempDir()
		vcsDir := filepath.Join(tmpDir, "vcs")
		worktreesDir := filepath.Join(tmpDir, "worktrees")
		worktree1 := filepath.Join(worktreesDir, "feature-1")

		// Setup
		require.NoError(t, os.MkdirAll(vcsDir, 0755))
		require.NoError(t, os.MkdirAll(worktreesDir, 0755))

		// Create docker-compose.yml in vcs
		composeContent := `version: '3'
services:
  test:
    image: alpine:latest
    command: echo "test"
`
		require.NoError(t, os.WriteFile(
			filepath.Join(vcsDir, "docker-compose.yml"),
			[]byte(composeContent),
			0644,
		))

		// Create override in root
		overrideContent := `version: '3'
services:
  test:
    environment:
      - TEST_ENV=override
`
		require.NoError(t, os.WriteFile(
			filepath.Join(tmpDir, "docker-compose.override.yml"),
			[]byte(overrideContent),
			0644,
		))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		os.Chdir(vcsDir)
		exec.Command("git", "init").Run()
		exec.Command("git", "config", "user.email", "test@example.com").Run()
		exec.Command("git", "config", "user.name", "Test User").Run()
		exec.Command("git", "add", ".").Run()
		exec.Command("git", "commit", "-m", "Initial").Run()

		// Create worktree
		exec.Command("git", "worktree", "add", worktree1, "-b", "feature-1").Run()

		// Verify docker-compose.yml exists in worktree
		assert.FileExists(t, filepath.Join(worktree1, "docker-compose.yml"))

		// Verify override is accessible from parent path
		assert.FileExists(t, filepath.Join(tmpDir, "docker-compose.override.yml"))
	})
}
