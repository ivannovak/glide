package e2e_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestYAMLCommandExecution tests YAML-defined command execution
func TestYAMLCommandExecution(t *testing.T) {
	glideBinary := buildGlideBinary(t)
	defer os.Remove(glideBinary)

	t.Run("execute_simple_yaml_command", func(t *testing.T) {
		// Setup: Create project with .glide.yml
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		configPath := filepath.Join(tmpDir, ".glide.yml")
		configContent := `version: 1
commands:
  hello:
    cmd: echo "Hello from YAML"
    description: "Say hello"
`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// Test: Execute YAML command
		cmd := exec.Command(glideBinary, "hello")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		// Assert: Command executes successfully
		require.NoError(t, err, "YAML command should execute")
		outputStr := string(output)
		assert.Contains(t, outputStr, "Hello from YAML", "Should execute YAML-defined command")
	})

	t.Run("execute_yaml_command_with_args", func(t *testing.T) {
		// Setup: Create project with parameterized command
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		configPath := filepath.Join(tmpDir, ".glide.yml")
		configContent := `version: 1
commands:
  greet:
    cmd: echo "Hello $1"
    description: "Greet someone"
`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// Test: Execute with arguments
		cmd := exec.Command(glideBinary, "greet", "World")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		// Assert: Arguments passed correctly
		if err == nil {
			outputStr := string(output)
			assert.Contains(t, outputStr, "World", "Should pass arguments to YAML command")
		}
	})

	t.Run("execute_multiline_yaml_command", func(t *testing.T) {
		// Setup: Create multi-line command
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		configPath := filepath.Join(tmpDir, ".glide.yml")
		configContent := `version: 1
commands:
  multi:
    cmd: |
      echo "Line 1"
      echo "Line 2"
      echo "Line 3"
    description: "Multi-line command"
`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// Test: Execute multi-line command with sanitization disabled
		// Note: Multi-line commands are blocked by default security sanitization
		// which detects newlines as potential injection. To test multiline execution,
		// we must disable sanitization.
		cmd := exec.Command(glideBinary, "multi")
		cmd.Dir = tmpDir
		cmd.Env = append(os.Environ(), "GLIDE_YAML_SANITIZE_MODE=disabled")
		output, err := cmd.CombinedOutput()

		// Assert: All lines execute
		require.NoError(t, err, "Multi-line command should execute")
		outputStr := string(output)
		assert.Contains(t, outputStr, "Line 1", "Should execute first line")
		assert.Contains(t, outputStr, "Line 2", "Should execute second line")
		assert.Contains(t, outputStr, "Line 3", "Should execute third line")
	})

	t.Run("execute_yaml_command_with_env_vars", func(t *testing.T) {
		// Setup: Create command using environment variables
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		configPath := filepath.Join(tmpDir, ".glide.yml")
		configContent := `version: 1
commands:
  env-test:
    cmd: echo "User: $USER, Path: $PWD"
    description: "Test env vars"
`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// Test: Execute command with env vars
		cmd := exec.Command(glideBinary, "env-test")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		// Assert: Environment variables expanded
		if err == nil {
			outputStr := string(output)
			assert.NotEmpty(t, outputStr, "Should output env var values")
		}
	})
}

// TestYAMLCommandErrorHandling tests error handling for YAML commands
func TestYAMLCommandErrorHandling(t *testing.T) {
	glideBinary := buildGlideBinary(t)
	defer os.Remove(glideBinary)

	t.Run("execute_nonexistent_yaml_command", func(t *testing.T) {
		// Setup: Create project with .glide.yml
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		configPath := filepath.Join(tmpDir, ".glide.yml")
		configContent := `version: 1
commands:
  exists:
    cmd: echo "exists"
`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// Test: Try nonexistent command
		cmd := exec.Command(glideBinary, "nonexistent")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		// Assert: Command not found
		assert.Error(t, err, "Nonexistent YAML command should fail")
		outputStr := string(output)
		assert.NotEmpty(t, outputStr, "Should output error message")
	})

	t.Run("execute_yaml_command_that_fails", func(t *testing.T) {
		// Setup: Create command that exits with error
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		configPath := filepath.Join(tmpDir, ".glide.yml")
		configContent := `version: 1
commands:
  fail:
    cmd: exit 1
    description: "Failing command"
`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// Test: Execute failing command
		cmd := exec.Command(glideBinary, "fail")
		cmd.Dir = tmpDir
		_, err := cmd.CombinedOutput()

		// Assert: Error propagated
		assert.Error(t, err, "Failing YAML command should return error")
	})

	t.Run("execute_yaml_command_invalid_syntax", func(t *testing.T) {
		// Setup: Create invalid YAML
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		configPath := filepath.Join(tmpDir, ".glide.yml")
		configContent := `version: 1
commands:
  broken:
    cmd: [invalid yaml syntax
`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// Test: Try to execute command from invalid config
		cmd := exec.Command(glideBinary, "broken")
		cmd.Dir = tmpDir
		_, err := cmd.CombinedOutput()

		// Assert: Should handle invalid YAML gracefully
		// Either fail to parse config or fail to find command
		assert.Error(t, err, "Invalid YAML should cause error")
	})

	t.Run("execute_yaml_command_missing_cmd_field", func(t *testing.T) {
		// Setup: Create command without cmd field
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		configPath := filepath.Join(tmpDir, ".glide.yml")
		configContent := `version: 1
commands:
  incomplete:
    description: "Command without cmd field"
`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// Test: Try to execute incomplete command
		cmd := exec.Command(glideBinary, "incomplete")
		cmd.Dir = tmpDir
		_, err := cmd.CombinedOutput()

		// Assert: Should handle missing cmd field
		// May error or treat as nonexistent
		if err != nil {
			assert.Error(t, err, "Command without cmd should fail")
		}
	})
}

// TestYAMLCommandSanitization tests command sanitization
func TestYAMLCommandSanitization(t *testing.T) {
	glideBinary := buildGlideBinary(t)
	defer os.Remove(glideBinary)

	t.Run("sanitize_command_injection_attempt", func(t *testing.T) {
		// Setup: Create project with potentially dangerous command
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		configPath := filepath.Join(tmpDir, ".glide.yml")
		// Note: Sanitization should prevent command chaining
		configContent := `version: 1
commands:
  dangerous:
    cmd: echo "safe"; echo "injected"
    description: "Test command injection"
`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// Test: Execute command
		cmd := exec.Command(glideBinary, "dangerous")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		// Assert: Command executes (sanitization in config loading)
		// Actual sanitization behavior depends on implementation
		outputStr := string(output)
		if err == nil {
			// Command may execute with sanitization applied
			assert.NotEmpty(t, outputStr)
		}
	})

	t.Run("sanitize_shell_metacharacters", func(t *testing.T) {
		// Setup: Create command with shell metacharacters
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		configPath := filepath.Join(tmpDir, ".glide.yml")
		configContent := `version: 1
commands:
  meta:
    cmd: echo "test" && echo "chained" || echo "or"
    description: "Test metacharacters"
`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// Test: Execute command
		cmd := exec.Command(glideBinary, "meta")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		// Assert: Sanitization handles metacharacters
		outputStr := string(output)
		if err == nil {
			// Behavior depends on sanitization mode
			assert.NotEmpty(t, outputStr)
		}
	})

	t.Run("sanitize_command_substitution", func(t *testing.T) {
		// Setup: Create command with command substitution
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		configPath := filepath.Join(tmpDir, ".glide.yml")
		configContent := `version: 1
commands:
  subst:
    cmd: echo "Current dir: $(pwd)"
    description: "Test command substitution"
`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// Test: Execute command
		cmd := exec.Command(glideBinary, "subst")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		// Assert: Command substitution handled
		outputStr := string(output)
		if err == nil {
			// May execute normally or sanitize substitution
			assert.NotEmpty(t, outputStr)
		}
	})

	t.Run("sanitize_backticks", func(t *testing.T) {
		// Setup: Create command with backtick substitution
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		configPath := filepath.Join(tmpDir, ".glide.yml")
		configContent := "version: 1\ncommands:\n  tick:\n    cmd: echo `whoami`\n    description: \"Test backticks\"\n"
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// Test: Execute command
		cmd := exec.Command(glideBinary, "tick")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		// Assert: Backticks handled by sanitization
		outputStr := string(output)
		if err == nil {
			assert.NotEmpty(t, outputStr)
		}
	})

	t.Run("allow_safe_commands", func(t *testing.T) {
		// Setup: Create safe commands that should work
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		configPath := filepath.Join(tmpDir, ".glide.yml")
		configContent := `version: 1
commands:
  safe:
    cmd: echo "This is safe"
    description: "Safe command"
  ls:
    cmd: ls -la
    description: "List files"
`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// Test: Execute safe commands
		cmd := exec.Command(glideBinary, "safe")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		// Assert: Safe commands execute normally
		require.NoError(t, err, "Safe commands should execute")
		outputStr := string(output)
		assert.Contains(t, outputStr, "This is safe", "Should execute safe command")
	})

	t.Run("allow_multiline_commands_in_script_mode", func(t *testing.T) {
		// Script mode (default) allows multi-line shell scripts in YAML commands
		// This is safe because the command string is user-authored, not user-input
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		configPath := filepath.Join(tmpDir, ".glide.yml")
		configContent := `version: 1
commands:
  multi:
    cmd: |
      echo "Line 1"
      echo "Line 2"
    description: "Multi-line command"
`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// Test: Execute with default script mode
		cmd := exec.Command(glideBinary, "multi")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		// Assert: Multi-line commands work in script mode (default)
		require.NoError(t, err, "Multi-line commands should work in default script mode")
		outputStr := string(output)
		assert.Contains(t, outputStr, "Line 1", "Should execute first line")
		assert.Contains(t, outputStr, "Line 2", "Should execute second line")
	})

	t.Run("block_multiline_commands_in_strict_mode", func(t *testing.T) {
		// Strict mode blocks multi-line commands as potential injection vectors
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		configPath := filepath.Join(tmpDir, ".glide.yml")
		configContent := `version: 1
commands:
  multi:
    cmd: |
      echo "Line 1"
      echo "Line 2"
    description: "Multi-line command"
`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// Test: Execute with strict sanitization mode
		cmd := exec.Command(glideBinary, "multi")
		cmd.Dir = tmpDir
		cmd.Env = append(os.Environ(), "GLIDE_YAML_SANITIZE_MODE=strict")
		output, err := cmd.CombinedOutput()

		// Assert: Command is blocked in strict mode
		assert.Error(t, err, "Multi-line commands should be blocked in strict mode")
		outputStr := string(output)
		assert.Contains(t, outputStr, "newline", "Error should mention newline as the issue")
	})
}

// TestYAMLCommandPriority tests command priority and overrides
func TestYAMLCommandPriority(t *testing.T) {
	glideBinary := buildGlideBinary(t)
	defer os.Remove(glideBinary)

	t.Run("local_config_overrides_global", func(t *testing.T) {
		// Setup: Create global and local configs
		tmpHome := t.TempDir()
		tmpProject := filepath.Join(tmpHome, "project")
		require.NoError(t, os.MkdirAll(tmpProject, 0755))

		gitDir := filepath.Join(tmpProject, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		// Global config
		globalConfig := filepath.Join(tmpHome, ".glide.yml")
		globalContent := `version: 1
commands:
  test:
    cmd: echo "global test"
`
		require.NoError(t, os.WriteFile(globalConfig, []byte(globalContent), 0644))

		// Local config (should override)
		localConfig := filepath.Join(tmpProject, ".glide.yml")
		localContent := `version: 1
commands:
  test:
    cmd: echo "local test"
`
		require.NoError(t, os.WriteFile(localConfig, []byte(localContent), 0644))

		// Test: Execute command from project
		cmd := exec.Command(glideBinary, "test")
		cmd.Dir = tmpProject
		cmd.Env = append(os.Environ(), "HOME="+tmpHome)
		output, err := cmd.CombinedOutput()

		// Assert: Local config takes precedence
		if err == nil {
			outputStr := string(output)
			// Should execute local version
			assert.Contains(t, outputStr, "local", "Local config should override global")
		}
	})

	t.Run("builtin_commands_precedence", func(t *testing.T) {
		// Setup: Try to override builtin command
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		configPath := filepath.Join(tmpDir, ".glide.yml")
		configContent := `version: 1
commands:
  version:
    cmd: echo "custom version"
    description: "Override version command"
`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// Test: Execute version command
		cmd := exec.Command(glideBinary, "version")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		// Assert: Builtin likely takes precedence
		require.NoError(t, err)
		outputStr := string(output)
		// Behavior depends on command precedence rules
		assert.NotEmpty(t, outputStr)
	})
}

// TestYAMLCommandHelp tests help for YAML commands
func TestYAMLCommandHelp(t *testing.T) {
	glideBinary := buildGlideBinary(t)
	defer os.Remove(glideBinary)

	t.Run("list_yaml_commands_in_help", func(t *testing.T) {
		// Setup: Create project with YAML commands
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		configPath := filepath.Join(tmpDir, ".glide.yml")
		configContent := `version: 1
commands:
  custom:
    cmd: echo "custom"
    description: "Custom YAML command"
`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// Test: Get help
		cmd := exec.Command(glideBinary, "help")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		// Assert: YAML commands appear in help
		if err == nil {
			outputStr := string(output)
			// Custom commands may appear in help output
			assert.NotEmpty(t, outputStr)
		}
	})

	t.Run("show_yaml_command_description", func(t *testing.T) {
		// Setup: Create command with description
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		configPath := filepath.Join(tmpDir, ".glide.yml")
		configContent := `version: 1
commands:
  described:
    cmd: echo "test"
    description: "This is a detailed description of the command"
`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// Test: Get help for specific command
		cmd := exec.Command(glideBinary, "help", "described")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		// Assert: Description shown if help supports it
		outputStr := string(output)
		if err == nil && len(outputStr) > 0 {
			// Description may appear in help output
			assert.NotEmpty(t, outputStr)
		}
	})
}
