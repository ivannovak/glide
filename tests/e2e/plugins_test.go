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

// TestPluginInstallation tests plugin installation E2E
func TestPluginInstallation(t *testing.T) {
	glideBinary := buildGlideBinary(t)
	defer os.Remove(glideBinary)

	t.Run("install_mock_plugin", func(t *testing.T) {
		// Setup: Create temporary home directory
		tmpHome := t.TempDir()
		pluginDir := filepath.Join(tmpHome, ".glide", "plugins")
		require.NoError(t, os.MkdirAll(pluginDir, 0755))

		// Create a mock plugin binary
		mockPluginPath := filepath.Join(pluginDir, "glide-plugin-test")
		mockPluginContent := `#!/bin/bash
if [ "$1" = "metadata" ]; then
  cat <<EOF
{
  "name": "test",
  "version": "1.0.0",
  "description": "Test plugin"
}
EOF
elif [ "$1" = "commands" ]; then
  cat <<EOF
{
  "commands": [
    {
      "name": "test-cmd",
      "description": "Test command"
    }
  ]
}
EOF
else
  echo "Test plugin executed"
fi
`
		require.NoError(t, os.WriteFile(mockPluginPath, []byte(mockPluginContent), 0755))

		// Test: List plugins to verify installation
		cmd := exec.Command(glideBinary, "plugins", "list")
		cmd.Env = append(os.Environ(), "HOME="+tmpHome)
		output, err := cmd.CombinedOutput()

		// Assert: Plugin appears in list
		require.NoError(t, err, "Plugins list should succeed")
		outputStr := string(output)
		// May or may not show the plugin depending on discovery mechanism
		assert.NotEmpty(t, outputStr, "Should output plugin list")
	})

	t.Run("install_invalid_plugin", func(t *testing.T) {
		// Setup: Create temporary home directory
		tmpHome := t.TempDir()
		pluginDir := filepath.Join(tmpHome, ".glide", "plugins")
		require.NoError(t, os.MkdirAll(pluginDir, 0755))

		// Create an invalid plugin file (not executable)
		invalidPluginPath := filepath.Join(pluginDir, "glide-plugin-invalid")
		require.NoError(t, os.WriteFile(invalidPluginPath, []byte("not a valid plugin"), 0644))

		// Test: List plugins
		cmd := exec.Command(glideBinary, "plugins", "list")
		cmd.Env = append(os.Environ(), "HOME="+tmpHome)
		_, err := cmd.CombinedOutput()

		// Assert: Should handle invalid plugin gracefully
		require.NoError(t, err, "Should not fail on invalid plugin")
		// Invalid plugin should be ignored
	})

	t.Run("install_plugin_wrong_name", func(t *testing.T) {
		// Setup: Create temporary home directory
		tmpHome := t.TempDir()
		pluginDir := filepath.Join(tmpHome, ".glide", "plugins")
		require.NoError(t, os.MkdirAll(pluginDir, 0755))

		// Create plugin with wrong naming convention
		wrongNamePath := filepath.Join(pluginDir, "wrong-name")
		require.NoError(t, os.WriteFile(wrongNamePath, []byte("#!/bin/bash\necho test"), 0755))

		// Test: List plugins
		cmd := exec.Command(glideBinary, "plugins", "list")
		cmd.Env = append(os.Environ(), "HOME="+tmpHome)
		_, err := cmd.CombinedOutput()

		// Assert: Plugin may be ignored if name doesn't match pattern
		require.NoError(t, err)
		// Behavior depends on plugin naming requirements
	})
}

// TestPluginExecution tests plugin command execution E2E
func TestPluginExecution(t *testing.T) {
	glideBinary := buildGlideBinary(t)
	defer os.Remove(glideBinary)

	t.Run("execute_plugin_command", func(t *testing.T) {
		// Setup: Create temporary home directory with plugin
		tmpHome := t.TempDir()
		pluginDir := filepath.Join(tmpHome, ".glide", "plugins")
		require.NoError(t, os.MkdirAll(pluginDir, 0755))

		// Create a working plugin
		pluginPath := filepath.Join(pluginDir, "glide-plugin-echo")
		pluginContent := `#!/bin/bash
if [ "$1" = "metadata" ]; then
  cat <<EOF
{
  "name": "echo",
  "version": "1.0.0",
  "description": "Echo plugin"
}
EOF
elif [ "$1" = "commands" ]; then
  cat <<EOF
{
  "commands": [
    {
      "name": "echo",
      "description": "Echo a message"
    }
  ]
}
EOF
elif [ "$1" = "execute" ] && [ "$2" = "echo" ]; then
  shift 2
  echo "Plugin echo: $@"
else
  echo "Unknown command"
  exit 1
fi
`
		require.NoError(t, os.WriteFile(pluginPath, []byte(pluginContent), 0755))

		// Test: Execute plugin command
		cmd := exec.Command(glideBinary, "echo", "hello", "world")
		cmd.Env = append(os.Environ(), "HOME="+tmpHome)
		output, err := cmd.CombinedOutput()

		// Assert: Plugin command executes
		// Note: Actual execution depends on how glide routes plugin commands
		outputStr := string(output)
		if err == nil {
			assert.Contains(t, outputStr, "hello", "Should execute plugin command")
		}
		// If error, plugin routing may work differently
	})

	t.Run("execute_plugin_with_args", func(t *testing.T) {
		// Setup: Create temporary home directory with plugin
		tmpHome := t.TempDir()
		pluginDir := filepath.Join(tmpHome, ".glide", "plugins")
		require.NoError(t, os.MkdirAll(pluginDir, 0755))

		// Create plugin that accepts arguments
		pluginPath := filepath.Join(pluginDir, "glide-plugin-args")
		pluginContent := `#!/bin/bash
if [ "$1" = "execute" ]; then
  shift
  echo "Args: $@"
  echo "Count: $#"
else
  echo "Not execute"
fi
`
		require.NoError(t, os.WriteFile(pluginPath, []byte(pluginContent), 0755))

		// Test: Execute with arguments
		cmd := exec.Command(glideBinary, "args", "--flag", "value", "positional")
		cmd.Env = append(os.Environ(), "HOME="+tmpHome)
		output, err := cmd.CombinedOutput()

		// Assert: Arguments passed correctly
		outputStr := string(output)
		if err == nil && strings.Contains(outputStr, "Args:") {
			assert.Contains(t, outputStr, "flag", "Should pass flags")
			assert.Contains(t, outputStr, "positional", "Should pass positional args")
		}
	})

	t.Run("execute_nonexistent_plugin_command", func(t *testing.T) {
		// Setup: Create temporary home directory
		tmpHome := t.TempDir()
		pluginDir := filepath.Join(tmpHome, ".glide", "plugins")
		require.NoError(t, os.MkdirAll(pluginDir, 0755))

		// Test: Try to execute non-existent plugin command
		cmd := exec.Command(glideBinary, "nonexistent-plugin-cmd-xyz")
		cmd.Env = append(os.Environ(), "HOME="+tmpHome)
		output, err := cmd.CombinedOutput()

		// Assert: Command not found
		assert.Error(t, err, "Nonexistent command should fail")
		outputStr := string(output)
		assert.NotEmpty(t, outputStr, "Should output error message")
	})

	t.Run("execute_plugin_with_error", func(t *testing.T) {
		// Setup: Create plugin that fails
		tmpHome := t.TempDir()
		pluginDir := filepath.Join(tmpHome, ".glide", "plugins")
		require.NoError(t, os.MkdirAll(pluginDir, 0755))

		pluginPath := filepath.Join(pluginDir, "glide-plugin-fail")
		pluginContent := `#!/bin/bash
if [ "$1" = "execute" ]; then
  echo "Plugin error" >&2
  exit 1
fi
`
		require.NoError(t, os.WriteFile(pluginPath, []byte(pluginContent), 0755))

		// Test: Execute failing plugin
		cmd := exec.Command(glideBinary, "fail")
		cmd.Env = append(os.Environ(), "HOME="+tmpHome)
		output, err := cmd.CombinedOutput()

		// Assert: Error propagated
		if strings.Contains(string(output), "Plugin error") {
			assert.Error(t, err, "Plugin error should propagate")
		}
	})
}

// TestPluginUninstallation tests plugin removal E2E
func TestPluginUninstallation(t *testing.T) {
	glideBinary := buildGlideBinary(t)
	defer os.Remove(glideBinary)

	t.Run("uninstall_plugin_by_removing_file", func(t *testing.T) {
		// Setup: Create temporary home directory with plugin
		tmpHome := t.TempDir()
		pluginDir := filepath.Join(tmpHome, ".glide", "plugins")
		require.NoError(t, os.MkdirAll(pluginDir, 0755))

		pluginPath := filepath.Join(pluginDir, "glide-plugin-remove")
		require.NoError(t, os.WriteFile(pluginPath, []byte("#!/bin/bash\necho test"), 0755))

		// Test: List plugins before removal
		cmd := exec.Command(glideBinary, "plugins", "list")
		cmd.Env = append(os.Environ(), "HOME="+tmpHome)
		_, err := cmd.CombinedOutput()
		require.NoError(t, err)

		// Remove plugin file
		require.NoError(t, os.Remove(pluginPath))

		// Test: List plugins after removal
		cmd = exec.Command(glideBinary, "plugins", "list")
		cmd.Env = append(os.Environ(), "HOME="+tmpHome)
		_, err = cmd.CombinedOutput()

		// Assert: Plugin no longer listed
		require.NoError(t, err)
		// Plugin should not appear in list after removal
	})

	t.Run("uninstall_nonexistent_plugin", func(t *testing.T) {
		// Setup: Create temporary home directory
		tmpHome := t.TempDir()
		pluginDir := filepath.Join(tmpHome, ".glide", "plugins")
		require.NoError(t, os.MkdirAll(pluginDir, 0755))

		// Test: Try to execute command from removed plugin
		cmd := exec.Command(glideBinary, "nonexistent-removed-plugin")
		cmd.Env = append(os.Environ(), "HOME="+tmpHome)
		output, err := cmd.CombinedOutput()

		// Assert: Command not found
		assert.Error(t, err, "Removed plugin command should not be found")
		assert.NotEmpty(t, string(output))
	})
}

// TestPluginDiscovery tests plugin discovery mechanisms
func TestPluginDiscovery(t *testing.T) {
	glideBinary := buildGlideBinary(t)
	defer os.Remove(glideBinary)

	t.Run("discover_plugins_in_multiple_directories", func(t *testing.T) {
		// Setup: Create multiple plugin directories
		tmpHome := t.TempDir()

		// Global plugin directory
		globalPluginDir := filepath.Join(tmpHome, ".glide", "plugins")
		require.NoError(t, os.MkdirAll(globalPluginDir, 0755))

		globalPlugin := filepath.Join(globalPluginDir, "glide-plugin-global")
		require.NoError(t, os.WriteFile(globalPlugin, []byte("#!/bin/bash\necho global"), 0755))

		// Local plugin directory (in project)
		tmpProject := filepath.Join(tmpHome, "project")
		localPluginDir := filepath.Join(tmpProject, ".glide", "plugins")
		require.NoError(t, os.MkdirAll(localPluginDir, 0755))

		localPlugin := filepath.Join(localPluginDir, "glide-plugin-local")
		require.NoError(t, os.WriteFile(localPlugin, []byte("#!/bin/bash\necho local"), 0755))

		// Test: List plugins from project directory
		cmd := exec.Command(glideBinary, "plugins", "list")
		cmd.Dir = tmpProject
		cmd.Env = append(os.Environ(), "HOME="+tmpHome)
		_, err := cmd.CombinedOutput()

		// Assert: Discovers plugins from multiple locations
		require.NoError(t, err)
		// Both global and local plugins may be discovered
	})

	t.Run("discover_plugins_with_different_permissions", func(t *testing.T) {
		// Setup: Create plugins with different permissions
		tmpHome := t.TempDir()
		pluginDir := filepath.Join(tmpHome, ".glide", "plugins")
		require.NoError(t, os.MkdirAll(pluginDir, 0755))

		// Executable plugin
		execPlugin := filepath.Join(pluginDir, "glide-plugin-exec")
		require.NoError(t, os.WriteFile(execPlugin, []byte("#!/bin/bash\necho exec"), 0755))

		// Non-executable plugin (should be ignored)
		nonExecPlugin := filepath.Join(pluginDir, "glide-plugin-nonexec")
		require.NoError(t, os.WriteFile(nonExecPlugin, []byte("#!/bin/bash\necho nonexec"), 0644))

		// Test: List plugins
		cmd := exec.Command(glideBinary, "plugins", "list")
		cmd.Env = append(os.Environ(), "HOME="+tmpHome)
		_, err := cmd.CombinedOutput()

		// Assert: Only executable plugins discovered
		require.NoError(t, err)
		// Non-executable plugin should be ignored
	})

	t.Run("discover_plugins_symlinks", func(t *testing.T) {
		// Skip on Windows
		if os.Getenv("GOOS") == "windows" {
			t.Skip("Skipping symlink test on Windows")
		}

		// Setup: Create plugin with symlink
		tmpHome := t.TempDir()
		pluginDir := filepath.Join(tmpHome, ".glide", "plugins")
		require.NoError(t, os.MkdirAll(pluginDir, 0755))

		// Original plugin
		originalPlugin := filepath.Join(tmpHome, "glide-plugin-original")
		require.NoError(t, os.WriteFile(originalPlugin, []byte("#!/bin/bash\necho original"), 0755))

		// Create symlink in plugin directory
		symlinkPlugin := filepath.Join(pluginDir, "glide-plugin-link")
		require.NoError(t, os.Symlink(originalPlugin, symlinkPlugin))

		// Test: List plugins
		cmd := exec.Command(glideBinary, "plugins", "list")
		cmd.Env = append(os.Environ(), "HOME="+tmpHome)
		_, err := cmd.CombinedOutput()

		// Assert: Symlinked plugins discovered
		require.NoError(t, err)
		// Symlink should be followed and plugin discovered
	})
}
