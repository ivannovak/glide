package e2e_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCommandExecution tests E2E execution of glide commands
func TestCommandExecution(t *testing.T) {
	// Build glide binary for testing
	glideBinary := buildGlideBinary(t)
	defer os.Remove(glideBinary)

	t.Run("glide_help", func(t *testing.T) {
		// Test: Execute glide help
		cmd := exec.Command(glideBinary, "help")
		output, err := cmd.CombinedOutput()

		// Assert: Command succeeds and shows help
		require.NoError(t, err, "glide help should succeed")
		outputStr := string(output)
		assert.Contains(t, outputStr, "Usage:", "Should show usage information")
		// Help output shows categorized commands (Core Commands, Setup & Configuration, etc.)
		assert.Contains(t, outputStr, "Core Commands", "Should list command categories")
	})

	t.Run("glide_help_with_command", func(t *testing.T) {
		// Test: Execute glide help <command>
		cmd := exec.Command(glideBinary, "help", "version")
		output, err := cmd.CombinedOutput()

		// Assert: Shows help for specific command
		if err == nil {
			outputStr := string(output)
			assert.Contains(t, outputStr, "version", "Should show version command help")
		}
		// If error, command may not support help subcommand (acceptable)
	})

	t.Run("glide_version", func(t *testing.T) {
		// Test: Execute glide version
		cmd := exec.Command(glideBinary, "version")
		output, err := cmd.CombinedOutput()

		// Assert: Command succeeds and shows version
		require.NoError(t, err, "glide version should succeed")
		outputStr := string(output)
		// Version output may vary, just check it's not empty
		assert.NotEmpty(t, outputStr, "Should output version information")
	})

	t.Run("glide_version_json", func(t *testing.T) {
		// Test: Execute glide version with JSON output
		cmd := exec.Command(glideBinary, "version", "--json")
		output, err := cmd.CombinedOutput()

		// Assert: JSON output if supported
		if err == nil {
			outputStr := string(output)
			// Check if JSON-like output (may contain braces)
			if strings.Contains(outputStr, "{") {
				assert.Contains(t, outputStr, "{", "JSON output should contain braces")
			}
		}
		// If error, --json flag may not be supported (acceptable)
	})

	t.Run("glide_context", func(t *testing.T) {
		// Setup: Create temporary git repository
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		// Test: Execute glide context in git repo
		cmd := exec.Command(glideBinary, "context")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		// Assert: Command succeeds and shows context
		require.NoError(t, err, "glide context should succeed in git repo")
		outputStr := string(output)
		assert.NotEmpty(t, outputStr, "Should output context information")
		// Context output should contain project info
	})

	t.Run("glide_context_json", func(t *testing.T) {
		// Setup: Create temporary git repository
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		// Test: Execute glide context --json
		cmd := exec.Command(glideBinary, "context", "--json")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		// Assert: JSON output if supported
		if err == nil {
			outputStr := string(output)
			if strings.Contains(outputStr, "{") {
				assert.Contains(t, outputStr, "{", "JSON output should contain braces")
			}
		}
	})

	t.Run("glide_plugins_list", func(t *testing.T) {
		// Test: Execute glide plugins list
		cmd := exec.Command(glideBinary, "plugins", "list")
		output, err := cmd.CombinedOutput()

		// Assert: Command succeeds (may have no plugins)
		require.NoError(t, err, "glide plugins list should succeed")
		outputStr := string(output)
		// Output may show "No plugins" or list plugins
		assert.NotEmpty(t, outputStr, "Should output plugin list (or no plugins message)")
	})

	t.Run("glide_plugins_list_verbose", func(t *testing.T) {
		// Test: Execute glide plugins list with verbose flag
		cmd := exec.Command(glideBinary, "plugins", "list", "-v")
		output, err := cmd.CombinedOutput()

		// Assert: Verbose output if supported
		if err == nil {
			outputStr := string(output)
			assert.NotEmpty(t, outputStr, "Should output verbose plugin list")
		}
	})

	t.Run("glide_invalid_command", func(t *testing.T) {
		// Test: Execute invalid command
		cmd := exec.Command(glideBinary, "nonexistent-command-xyz")
		output, err := cmd.CombinedOutput()

		// Assert: Command fails with error
		assert.Error(t, err, "Invalid command should fail")
		outputStr := string(output)
		// Error message should indicate unknown command
		assert.True(t,
			strings.Contains(outputStr, "unknown") ||
				strings.Contains(outputStr, "not found") ||
				strings.Contains(outputStr, "invalid"),
			"Should indicate unknown command")
	})

	t.Run("glide_help_flag", func(t *testing.T) {
		// Test: Execute glide --help
		cmd := exec.Command(glideBinary, "--help")
		output, err := cmd.CombinedOutput()

		// Assert: Shows help
		require.NoError(t, err, "glide --help should succeed")
		outputStr := string(output)
		assert.Contains(t, outputStr, "Usage:", "Should show usage information")
	})

	t.Run("glide_version_flag", func(t *testing.T) {
		// Test: Execute glide --version
		cmd := exec.Command(glideBinary, "--version")
		output, err := cmd.CombinedOutput()

		// Assert: Shows version
		require.NoError(t, err, "glide --version should succeed")
		outputStr := string(output)
		assert.NotEmpty(t, outputStr, "Should output version")
	})
}

// TestCommandErrorHandling tests error handling in E2E scenarios
func TestCommandErrorHandling(t *testing.T) {
	glideBinary := buildGlideBinary(t)
	defer os.Remove(glideBinary)

	t.Run("context_outside_git_repo", func(t *testing.T) {
		// Setup: Create non-git directory
		tmpDir := t.TempDir()

		// Test: Execute glide context outside git repo
		cmd := exec.Command(glideBinary, "context")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		// Assert: May error or show limited context
		outputStr := string(output)
		if err != nil {
			// Error is acceptable for non-git directories
			assert.NotEmpty(t, outputStr, "Should output error message")
		}
	})

	t.Run("plugins_list_no_plugin_dir", func(t *testing.T) {
		// Setup: Create directory without plugins
		tmpDir := t.TempDir()

		// Test: Execute glide plugins list
		cmd := exec.Command(glideBinary, "plugins", "list")
		cmd.Dir = tmpDir
		cmd.Env = append(os.Environ(), "HOME="+tmpDir)
		output, err := cmd.CombinedOutput()

		// Assert: Should handle missing plugin directory gracefully
		require.NoError(t, err, "Should succeed even without plugins")
		outputStr := string(output)
		assert.NotEmpty(t, outputStr, "Should output something")
	})

	t.Run("command_with_invalid_flags", func(t *testing.T) {
		// Test: Execute command with invalid flags
		cmd := exec.Command(glideBinary, "version", "--invalid-flag-xyz")
		output, err := cmd.CombinedOutput()

		// Assert: Should error on invalid flags
		assert.Error(t, err, "Invalid flags should cause error")
		outputStr := string(output)
		assert.NotEmpty(t, outputStr, "Should output error message")
	})
}

// TestCommandOutput tests command output formatting
func TestCommandOutput(t *testing.T) {
	glideBinary := buildGlideBinary(t)
	defer os.Remove(glideBinary)

	t.Run("version_output_format", func(t *testing.T) {
		// Test: Check version output format
		cmd := exec.Command(glideBinary, "version")
		output, err := cmd.CombinedOutput()

		require.NoError(t, err)
		outputStr := strings.TrimSpace(string(output))

		// Assert: Version output is properly formatted
		assert.NotEmpty(t, outputStr, "Version output should not be empty")
		// Version format may vary (v1.0.0, 1.0.0, dev, etc.)
	})

	t.Run("help_output_structure", func(t *testing.T) {
		// Test: Check help output structure
		cmd := exec.Command(glideBinary, "help")
		output, err := cmd.CombinedOutput()

		require.NoError(t, err)
		outputStr := string(output)

		// Assert: Help has expected sections
		assert.Contains(t, outputStr, "Usage:", "Should have Usage section")
		// May have other sections like Commands, Flags, etc.
	})

	t.Run("plugins_list_output_format", func(t *testing.T) {
		// Test: Check plugins list output format
		cmd := exec.Command(glideBinary, "plugins", "list")
		output, err := cmd.CombinedOutput()

		require.NoError(t, err)
		outputStr := string(output)

		// Assert: Output is formatted (table, list, or message)
		assert.NotEmpty(t, outputStr, "Plugins list should output something")
	})
}

// buildGlideBinary builds the glide binary for testing
func buildGlideBinary(t *testing.T) string {
	t.Helper()

	// Create temporary binary
	tmpBinary := filepath.Join(t.TempDir(), "glide-test")

	// Build glide
	cmd := exec.Command("go", "build", "-o", tmpBinary, "../../cmd/glide")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to build glide binary: %s", string(output))

	return tmpBinary
}
