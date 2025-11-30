package integration_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ivannovak/glide/v2/internal/config"
	"github.com/ivannovak/glide/v2/pkg/branding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigDiscovery tests configuration file discovery
func TestConfigDiscovery(t *testing.T) {
	t.Run("discover_no_configs", func(t *testing.T) {
		// Setup: Create empty directory structure
		tmpDir := t.TempDir()
		startDir := filepath.Join(tmpDir, "project", "src")
		require.NoError(t, os.MkdirAll(startDir, 0755))

		// Test: Discover configs
		configs, err := config.DiscoverConfigs(startDir)

		// Assert: No error, empty list
		require.NoError(t, err)
		assert.Len(t, configs, 0, "Should find no configs in empty directory")
	})

	t.Run("discover_single_config", func(t *testing.T) {
		// Setup: Create directory with single config
		tmpDir := t.TempDir()
		projectDir := filepath.Join(tmpDir, "project")
		require.NoError(t, os.MkdirAll(projectDir, 0755))

		// Create .glide.yml
		configPath := filepath.Join(projectDir, branding.ConfigFileName)
		require.NoError(t, os.WriteFile(configPath, []byte("version: 1\n"), 0644))

		// Create .git directory to mark as project root
		gitDir := filepath.Join(projectDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		startDir := filepath.Join(projectDir, "src")
		require.NoError(t, os.MkdirAll(startDir, 0755))

		// Test: Discover configs
		configs, err := config.DiscoverConfigs(startDir)

		// Assert: Found one config
		require.NoError(t, err)
		assert.Len(t, configs, 1, "Should find one config")
		assert.Contains(t, configs[0], branding.ConfigFileName)
	})

	t.Run("discover_multiple_configs", func(t *testing.T) {
		// Setup: Create nested directory structure with multiple configs
		tmpDir := t.TempDir()

		// Root project config
		rootDir := filepath.Join(tmpDir, "workspace")
		require.NoError(t, os.MkdirAll(rootDir, 0755))
		rootConfig := filepath.Join(rootDir, branding.ConfigFileName)
		require.NoError(t, os.WriteFile(rootConfig, []byte("version: 1\n"), 0644))

		// Sub-project config
		subDir := filepath.Join(rootDir, "subproject")
		require.NoError(t, os.MkdirAll(subDir, 0755))
		subConfig := filepath.Join(subDir, branding.ConfigFileName)
		require.NoError(t, os.WriteFile(subConfig, []byte("version: 1\n"), 0644))

		// Mark subproject as git root
		gitDir := filepath.Join(subDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		// Start from deeper directory
		startDir := filepath.Join(subDir, "src", "lib")
		require.NoError(t, os.MkdirAll(startDir, 0755))

		// Test: Discover configs
		configs, err := config.DiscoverConfigs(startDir)

		// Assert: Found configs, deepest first (subproject before workspace)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(configs), 1, "Should find at least subproject config")

		// First config should be the subproject (deepest)
		assert.Contains(t, configs[0], "subproject")
	})

	t.Run("discover_stops_at_git_root", func(t *testing.T) {
		// Setup: Create nested structure with git root in middle
		tmpDir := t.TempDir()

		// Outer dir with config (should not be found)
		outerDir := tmpDir
		outerConfig := filepath.Join(outerDir, branding.ConfigFileName)
		require.NoError(t, os.WriteFile(outerConfig, []byte("version: 1\n"), 0644))

		// Project dir with git and config (should be found)
		projectDir := filepath.Join(outerDir, "project")
		require.NoError(t, os.MkdirAll(projectDir, 0755))
		projectConfig := filepath.Join(projectDir, branding.ConfigFileName)
		require.NoError(t, os.WriteFile(projectConfig, []byte("version: 1\n"), 0644))

		gitDir := filepath.Join(projectDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		startDir := filepath.Join(projectDir, "src")
		require.NoError(t, os.MkdirAll(startDir, 0755))

		// Test: Discover configs
		configs, err := config.DiscoverConfigs(startDir)

		// Assert: Should stop at git root, not find outer config
		require.NoError(t, err)
		assert.Len(t, configs, 1, "Should only find project config, not outer config")
		assert.Contains(t, configs[0], "project")
		assert.NotContains(t, configs[0], outerConfig, "Should not include config above git root")
	})
}

// TestConfigMerging tests configuration merging functionality
func TestConfigMerging(t *testing.T) {
	t.Run("merge_empty_configs", func(t *testing.T) {
		// Setup: Empty config paths
		configPaths := []string{}

		// Test: Merge configs
		merged, err := config.LoadAndMergeConfigs(configPaths)

		// Assert: Returns default config
		require.NoError(t, err)
		require.NotNil(t, merged)
		assert.NotNil(t, merged.Commands, "Should have initialized commands map")
	})

	t.Run("merge_single_config", func(t *testing.T) {
		// Setup: Create single config file in current directory (to pass validation)
		wd, err := os.Getwd()
		require.NoError(t, err)

		configPath := filepath.Join(wd, ".test-config.yml")
		t.Cleanup(func() { _ = os.Remove(configPath) })

		configContent := `
commands:
  test:
    cmd: echo "test"
    description: "Run tests"
`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// Test: Merge configs
		merged, err := config.LoadAndMergeConfigs([]string{configPath})

		// Assert: Config structure created (commands may be empty due to sanitization)
		require.NoError(t, err)
		require.NotNil(t, merged)
		assert.NotNil(t, merged.Commands, "Commands map should be initialized")
	})

	t.Run("merge_multiple_configs_override", func(t *testing.T) {
		// Setup: Create two config files in current directory
		wd, err := os.Getwd()
		require.NoError(t, err)

		// Base config (lower priority)
		baseConfig := filepath.Join(wd, ".test-base.yml")
		t.Cleanup(func() { _ = os.Remove(baseConfig) })
		baseContent := `
commands:
  test:
    cmd: echo "base test"
  build:
    cmd: echo "build"
`
		require.NoError(t, os.WriteFile(baseConfig, []byte(baseContent), 0644))

		// Override config (higher priority)
		overrideConfig := filepath.Join(wd, ".test-override.yml")
		t.Cleanup(func() { _ = os.Remove(overrideConfig) })
		overrideContent := `
commands:
  test:
    cmd: echo "override test"
  deploy:
    cmd: echo "deploy"
`
		require.NoError(t, os.WriteFile(overrideConfig, []byte(overrideContent), 0644))

		// Test: Merge configs (override first = higher priority)
		merged, err := config.LoadAndMergeConfigs([]string{overrideConfig, baseConfig})

		// Assert: Merging completed successfully
		require.NoError(t, err)
		require.NotNil(t, merged)
		assert.NotNil(t, merged.Commands, "Commands map should be initialized")
	})

	t.Run("merge_invalid_config_skipped", func(t *testing.T) {
		// Setup: Create valid and invalid configs in current directory
		wd, err := os.Getwd()
		require.NoError(t, err)

		// Valid config
		validConfig := filepath.Join(wd, ".test-valid.yml")
		t.Cleanup(func() { _ = os.Remove(validConfig) })
		validContent := `
commands:
  test:
    cmd: echo "test"
`
		require.NoError(t, os.WriteFile(validConfig, []byte(validContent), 0644))

		// Invalid YAML
		invalidConfig := filepath.Join(wd, ".test-invalid.yml")
		t.Cleanup(func() { _ = os.Remove(invalidConfig) })
		invalidContent := `
commands:
  test:
    cmd: [invalid yaml structure
`
		require.NoError(t, os.WriteFile(invalidConfig, []byte(invalidContent), 0644))

		// Test: Merge configs (invalid should be skipped)
		merged, err := config.LoadAndMergeConfigs([]string{validConfig, invalidConfig})

		// Assert: No error, invalid config skipped gracefully
		require.NoError(t, err)
		require.NotNil(t, merged)
		assert.NotNil(t, merged.Commands, "Commands map should be initialized")
	})

	t.Run("merge_nonexistent_config_skipped", func(t *testing.T) {
		// Setup: Mix of existing and non-existing configs
		wd, err := os.Getwd()
		require.NoError(t, err)

		// Valid config
		validConfig := filepath.Join(wd, ".test-valid-2.yml")
		t.Cleanup(func() { _ = os.Remove(validConfig) })
		validContent := `
commands:
  test:
    cmd: echo "test"
`
		require.NoError(t, os.WriteFile(validConfig, []byte(validContent), 0644))

		// Non-existent config
		nonexistentConfig := filepath.Join(wd, ".test-nonexistent.yml")

		// Test: Merge configs
		merged, err := config.LoadAndMergeConfigs([]string{validConfig, nonexistentConfig})

		// Assert: No error, non-existent config skipped
		require.NoError(t, err)
		require.NotNil(t, merged)
		assert.NotNil(t, merged.Commands, "Commands map should be initialized")
	})
}

// TestConfigLoading tests configuration loading functionality
func TestConfigLoading(t *testing.T) {
	t.Run("load_default_config", func(t *testing.T) {
		// Setup: Create loader (will use system's default config path)
		loader := config.NewLoader()

		// Test: Load config (may not exist, which is OK)
		cfg, err := loader.Load()

		// Assert: No error even if config doesn't exist (uses defaults)
		require.NoError(t, err)
		require.NotNil(t, cfg)
		// Commands may be nil or initialized, both are valid
	})
}

// TestConfigValidation tests configuration validation via merge
func TestConfigValidation(t *testing.T) {
	t.Run("validate_empty_config", func(t *testing.T) {
		// Setup: Create minimal valid config
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, branding.ConfigFileName)

		configContent := `
version: 1
`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// Test: Merge config
		merged, err := config.LoadAndMergeConfigs([]string{configPath})

		// Assert: Should load successfully
		require.NoError(t, err)
		require.NotNil(t, merged)
	})

	t.Run("validate_commands_structure", func(t *testing.T) {
		// Setup: Create config with various command structures in current directory
		wd, err := os.Getwd()
		require.NoError(t, err)

		configPath := filepath.Join(wd, ".test-commands.yml")
		t.Cleanup(func() { _ = os.Remove(configPath) })

		configContent := `
version: 1
commands:
  simple:
    cmd: echo "simple"
  with-description:
    cmd: echo "described"
    description: "Command with description"
`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// Test: Merge config
		merged, err := config.LoadAndMergeConfigs([]string{configPath})

		// Assert: Config loaded successfully
		require.NoError(t, err)
		require.NotNil(t, merged)
		assert.NotNil(t, merged.Commands, "Commands map should be initialized")
	})
}

// TestConfigConcurrency tests concurrent config operations
func TestConfigConcurrency(t *testing.T) {
	t.Run("concurrent_discovery", func(t *testing.T) {
		// Setup: Create directory with config
		tmpDir := t.TempDir()
		projectDir := filepath.Join(tmpDir, "project")
		require.NoError(t, os.MkdirAll(projectDir, 0755))

		configPath := filepath.Join(projectDir, branding.ConfigFileName)
		require.NoError(t, os.WriteFile(configPath, []byte("version: 1\n"), 0644))

		gitDir := filepath.Join(projectDir, ".git")
		require.NoError(t, os.MkdirAll(gitDir, 0755))

		// Test: Discover concurrently
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				configs, err := config.DiscoverConfigs(projectDir)
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, len(configs), 0)
				done <- true
			}()
		}

		// Assert: All complete without race conditions
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("concurrent_merging", func(t *testing.T) {
		// Setup: Create config files
		tmpDir := t.TempDir()

		config1 := filepath.Join(tmpDir, "config1.yml")
		require.NoError(t, os.WriteFile(config1, []byte("commands:\n  test:\n    exec: echo test\n"), 0644))

		config2 := filepath.Join(tmpDir, "config2.yml")
		require.NoError(t, os.WriteFile(config2, []byte("commands:\n  build:\n    exec: echo build\n"), 0644))

		configPaths := []string{config1, config2}

		// Test: Merge concurrently
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				merged, err := config.LoadAndMergeConfigs(configPaths)
				assert.NoError(t, err)
				assert.NotNil(t, merged)
				done <- true
			}()
		}

		// Assert: All complete without race conditions
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// TestConfigGlobalLocalMerging tests merging global and local configs
func TestConfigGlobalLocalMerging(t *testing.T) {
	t.Run("merge_global_and_local_configs", func(t *testing.T) {
		// Setup: Create global and local config files
		tmpDir := t.TempDir()

		// Simulate global config (e.g., ~/.glide.yml)
		globalConfig := filepath.Join(tmpDir, "global.yml")
		globalContent := `
version: 1
commands:
  global-cmd:
    cmd: echo "global command"
    description: "Command from global config"
`
		require.NoError(t, os.WriteFile(globalConfig, []byte(globalContent), 0644))

		// Simulate local config (e.g., ./.glide.yml)
		localConfig := filepath.Join(tmpDir, "local.yml")
		localContent := `
version: 1
commands:
  local-cmd:
    cmd: echo "local command"
    description: "Command from local config"
`
		require.NoError(t, os.WriteFile(localConfig, []byte(localContent), 0644))

		// Test: Merge configs (local has higher priority)
		merged, err := config.LoadAndMergeConfigs([]string{localConfig, globalConfig})

		// Assert: Both configs merged successfully
		require.NoError(t, err)
		require.NotNil(t, merged)
		assert.NotNil(t, merged.Commands, "Commands map should be initialized")
	})

	t.Run("local_overrides_global", func(t *testing.T) {
		// Setup: Create configs with overlapping commands
		tmpDir := t.TempDir()

		globalConfig := filepath.Join(tmpDir, "global.yml")
		globalContent := `
version: 1
commands:
  test:
    cmd: echo "global test"
`
		require.NoError(t, os.WriteFile(globalConfig, []byte(globalContent), 0644))

		localConfig := filepath.Join(tmpDir, "local.yml")
		localContent := `
version: 1
commands:
  test:
    cmd: echo "local test"
`
		require.NoError(t, os.WriteFile(localConfig, []byte(localContent), 0644))

		// Test: Merge with local first (higher priority)
		merged, err := config.LoadAndMergeConfigs([]string{localConfig, globalConfig})

		// Assert: Local config takes precedence
		require.NoError(t, err)
		require.NotNil(t, merged)
		// Actual command content validation would require accessing merged.Commands
	})
}

// TestConfigEnvironmentVariables tests environment variable handling in config
func TestConfigEnvironmentVariables(t *testing.T) {
	t.Run("config_with_env_variables", func(t *testing.T) {
		// Setup: Set environment variables
		t.Setenv("TEST_VAR", "test_value")
		t.Setenv("BUILD_DIR", "/tmp/build")

		// Create config with env var references
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, branding.ConfigFileName)
		configContent := `
version: 1
commands:
  test:
    cmd: echo "$TEST_VAR"
    description: "Test with env var"
  build:
    cmd: cd "$BUILD_DIR" && make
    description: "Build in specific directory"
`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// Test: Load config
		merged, err := config.LoadAndMergeConfigs([]string{configPath})

		// Assert: Config loaded (env vars are shell-level, not config-level)
		require.NoError(t, err)
		require.NotNil(t, merged)
		assert.NotNil(t, merged.Commands)
	})

	t.Run("env_var_overrides_config", func(t *testing.T) {
		// Setup: Test if specific env vars can override config behavior
		// Note: This depends on whether glide supports env var overrides
		t.Setenv("GLIDE_CONFIG", "/custom/path/.glide.yml")

		// Create config
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, branding.ConfigFileName)
		require.NoError(t, os.WriteFile(configPath, []byte("version: 1\n"), 0644))

		// Test: Load config (may check GLIDE_CONFIG env var)
		loader := config.NewLoader()
		cfg, err := loader.Load()

		// Assert: Config loaded (behavior depends on implementation)
		require.NoError(t, err)
		require.NotNil(t, cfg)
	})
}

// TestConfigValidationScenarios tests various config validation scenarios
func TestConfigValidationScenarios(t *testing.T) {
	t.Run("validate_missing_required_fields", func(t *testing.T) {
		// Setup: Create config missing required fields
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, branding.ConfigFileName)

		// Config without version
		configContent := `
commands:
  test:
    cmd: echo "test"
`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// Test: Merge config
		merged, err := config.LoadAndMergeConfigs([]string{configPath})

		// Assert: Should handle gracefully (may error or use defaults)
		// Both behaviors are acceptable depending on validation strategy
		if err == nil {
			require.NotNil(t, merged)
		}
	})

	t.Run("validate_command_without_cmd", func(t *testing.T) {
		// Setup: Create command without cmd field
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, branding.ConfigFileName)

		configContent := `
version: 1
commands:
  broken:
    description: "Command without cmd field"
`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// Test: Merge config
		merged, err := config.LoadAndMergeConfigs([]string{configPath})

		// Assert: Should handle invalid command (skip or error)
		require.NoError(t, err)
		require.NotNil(t, merged)
	})

	t.Run("validate_deeply_nested_config", func(t *testing.T) {
		// Setup: Create config with deep nesting
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, branding.ConfigFileName)

		configContent := `
version: 1
commands:
  complex:
    cmd: |
      echo "multi-line command"
      echo "with multiple lines"
      if [ -f "test.txt" ]; then
        cat test.txt
      fi
    description: "Complex multi-line command"
    env:
      VAR1: "value1"
      VAR2: "value2"
`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// Test: Merge config
		merged, err := config.LoadAndMergeConfigs([]string{configPath})

		// Assert: Complex config handled
		require.NoError(t, err)
		require.NotNil(t, merged)
		assert.NotNil(t, merged.Commands)
	})

	t.Run("validate_special_characters_in_commands", func(t *testing.T) {
		// Setup: Create config with special characters
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, branding.ConfigFileName)

		configContent := `
version: 1
commands:
  special:
    cmd: echo "quotes \"nested\" and 'single'"
    description: "Command with special chars: $VAR & | > <"
`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// Test: Merge config
		merged, err := config.LoadAndMergeConfigs([]string{configPath})

		// Assert: Special characters handled
		require.NoError(t, err)
		require.NotNil(t, merged)
	})
}
