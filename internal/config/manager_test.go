package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glide-cli/glide/v3/internal/context"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_Initialize(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Create config
	configPath := filepath.Join(tempDir, ".glide.yml")
	yamlContent := `
projects:
  test:
    path: /test
defaults:
  test:
    processes: 5
`
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	manager := NewManager()
	ctx := &context.ProjectContext{
		ProjectRoot: "/test",
	}

	err = manager.Initialize(ctx)
	require.NoError(t, err)

	cfg := manager.GetConfig()
	assert.NotNil(t, cfg)
	assert.Equal(t, 5, cfg.Defaults.Test.Processes)

	activeProject := manager.GetActiveProject()
	assert.NotNil(t, activeProject)
	assert.Equal(t, "/test", activeProject.Path)
}

func TestManager_GetCommandConfig_Defaults(t *testing.T) {
	manager := NewManager()

	cc := manager.GetCommandConfig()
	assert.NotNil(t, cc)

	// Should have default values
	assert.True(t, cc.Test.Parallel)
	assert.Equal(t, 3, cc.Test.Processes)
	assert.Equal(t, 30, cc.Docker.ComposeTimeout)
}

func TestManager_EnvironmentVariables_Override(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Create config
	configPath := filepath.Join(tempDir, ".glide.yml")
	yamlContent := `
defaults:
  test:
    parallel: false
    processes: 5
  docker:
    compose_timeout: 60
`
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Set environment variables
	os.Setenv("GLIDE_TEST_PARALLEL", "true")
	defer os.Unsetenv("GLIDE_TEST_PARALLEL")

	os.Setenv("GLIDE_TEST_PROCESSES", "10")
	defer os.Unsetenv("GLIDE_TEST_PROCESSES")

	os.Setenv("GLIDE_DOCKER_TIMEOUT", "120")
	defer os.Unsetenv("GLIDE_DOCKER_TIMEOUT")

	manager := NewManager()
	err = manager.Initialize(nil)
	require.NoError(t, err)

	cc := manager.GetCommandConfig()

	// Environment should override config
	assert.True(t, cc.Test.Parallel, "Environment variable should override config")
	assert.Equal(t, 10, cc.Test.Processes, "Environment variable should override config")
	assert.Equal(t, 120, cc.Docker.ComposeTimeout, "Environment variable should override config")
}

func TestManager_ApplyFlags_OverrideAll(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Create config
	configPath := filepath.Join(tempDir, ".glide.yml")
	yamlContent := `
defaults:
  test:
    parallel: false
    processes: 5
`
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Set environment variable
	os.Setenv("GLIDE_TEST_PROCESSES", "10")
	defer os.Unsetenv("GLIDE_TEST_PROCESSES")

	manager := NewManager()
	err = manager.Initialize(nil)
	require.NoError(t, err)

	// Create command with flags
	cmd := &cobra.Command{}
	cmd.Flags().Bool("parallel", false, "")
	cmd.Flags().Int("processes", 0, "")
	cmd.Flags().Bool("coverage", false, "")

	// Set flag values
	cmd.Flags().Set("parallel", "true")
	cmd.Flags().Set("processes", "20")
	cmd.Flags().Set("coverage", "true")

	manager.ApplyFlags(cmd)
	cc := manager.GetCommandConfig()

	// Flags should override both environment and config
	assert.True(t, cc.Test.Parallel)
	assert.Equal(t, 20, cc.Test.Processes, "Flags should override environment and config")
	assert.True(t, cc.Test.Coverage)
}

func TestManager_ColorEnabled_Auto(t *testing.T) {
	manager := NewManager()
	err := manager.Initialize(nil)
	require.NoError(t, err)

	cc := manager.GetCommandConfig()

	// Default is "auto" which depends on TTY
	// We can't reliably test TTY detection, but we can verify it doesn't crash
	assert.NotNil(t, cc)
}

func TestManager_ColorEnabled_Always(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	configPath := filepath.Join(tempDir, ".glide.yml")
	yamlContent := `
defaults:
  colors:
    enabled: always
`
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	manager := NewManager()
	err = manager.Initialize(nil)
	require.NoError(t, err)

	cc := manager.GetCommandConfig()
	assert.True(t, cc.Colors.Enabled)
}

func TestManager_ColorEnabled_Never(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	configPath := filepath.Join(tempDir, ".glide.yml")
	yamlContent := `
defaults:
  colors:
    enabled: never
`
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	manager := NewManager()
	err = manager.Initialize(nil)
	require.NoError(t, err)

	cc := manager.GetCommandConfig()
	assert.False(t, cc.Colors.Enabled)
}

func TestManager_ColorEnabled_NO_COLOR_Env(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Config says always
	configPath := filepath.Join(tempDir, ".glide.yml")
	yamlContent := `
defaults:
  colors:
    enabled: always
`
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// But NO_COLOR env should override
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	manager := NewManager()
	err = manager.Initialize(nil)
	require.NoError(t, err)

	cc := manager.GetCommandConfig()
	assert.False(t, cc.Colors.Enabled, "NO_COLOR should override config")
}

func TestManager_GetProjectByName(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	configPath := filepath.Join(tempDir, ".glide.yml")
	yamlContent := `
projects:
  proj1:
    path: /path1
  proj2:
    path: /path2
`
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	manager := NewManager()
	err = manager.Initialize(nil)
	require.NoError(t, err)

	proj, ok := manager.GetProjectByName("proj1")
	assert.True(t, ok)
	assert.NotNil(t, proj)
	assert.Equal(t, "/path1", proj.Path)

	proj, ok = manager.GetProjectByName("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, proj)
}

func TestManager_SetDefaultProject(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	configPath := filepath.Join(tempDir, ".glide.yml")
	yamlContent := `
projects:
  proj1:
    path: /path1
  proj2:
    path: /path2
`
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	manager := NewManager()
	err = manager.Initialize(nil)
	require.NoError(t, err)

	err = manager.SetDefaultProject("proj2")
	require.NoError(t, err)

	// Reload to verify it was saved
	cfg := manager.GetConfig()
	assert.Equal(t, "proj2", cfg.DefaultProject)
}

func TestManager_SetDefaultProject_NonExistent(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	configPath := filepath.Join(tempDir, ".glide.yml")
	yamlContent := `
projects:
  proj1:
    path: /path1
`
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	manager := NewManager()
	err = manager.Initialize(nil)
	require.NoError(t, err)

	err = manager.SetDefaultProject("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestManager_ApplyFlags_DockerFlags(t *testing.T) {
	manager := NewManager()
	err := manager.Initialize(nil)
	require.NoError(t, err)

	cmd := &cobra.Command{}
	cmd.Flags().Int("timeout", 0, "")
	cmd.Flags().Bool("no-auto-start", false, "")
	cmd.Flags().Bool("remove-orphans", false, "")

	cmd.Flags().Set("timeout", "90")
	cmd.Flags().Set("no-auto-start", "true")
	cmd.Flags().Set("remove-orphans", "true")

	manager.ApplyFlags(cmd)
	cc := manager.GetCommandConfig()

	assert.Equal(t, 90, cc.Docker.ComposeTimeout)
	assert.False(t, cc.Docker.AutoStart, "no-auto-start flag should disable auto start")
	assert.True(t, cc.Docker.RemoveOrphans)
}

func TestManager_ApplyFlags_ColorFlags(t *testing.T) {
	manager := NewManager()
	err := manager.Initialize(nil)
	require.NoError(t, err)

	cmd := &cobra.Command{}
	cmd.Flags().String("color", "", "")
	cmd.Flags().Bool("no-color", false, "")

	// Test --color=always
	cmd.Flags().Set("color", "always")
	manager.ApplyFlags(cmd)
	cc := manager.GetCommandConfig()
	assert.True(t, cc.Colors.Enabled)

	// Test --no-color
	cmd2 := &cobra.Command{}
	cmd2.Flags().String("color", "", "")
	cmd2.Flags().Bool("no-color", false, "")
	cmd2.Flags().Set("no-color", "true")

	manager2 := NewManager()
	err = manager2.Initialize(nil)
	require.NoError(t, err)

	manager2.ApplyFlags(cmd2)
	cc2 := manager2.GetCommandConfig()
	assert.False(t, cc2.Colors.Enabled)
}

func TestManager_ApplyFlags_WorktreeFlags(t *testing.T) {
	manager := NewManager()
	err := manager.Initialize(nil)
	require.NoError(t, err)

	cmd := &cobra.Command{}
	cmd.Flags().Bool("auto-setup", false, "")
	cmd.Flags().Bool("no-copy-env", false, "")
	cmd.Flags().Bool("run-migrations", false, "")

	cmd.Flags().Set("auto-setup", "true")
	cmd.Flags().Set("no-copy-env", "true")
	cmd.Flags().Set("run-migrations", "true")

	manager.ApplyFlags(cmd)
	cc := manager.GetCommandConfig()

	assert.True(t, cc.Worktree.AutoSetup)
	assert.False(t, cc.Worktree.CopyEnv, "no-copy-env should disable copy")
	assert.True(t, cc.Worktree.RunMigrations)
}

func TestManager_PrecedenceOrder(t *testing.T) {
	// This test verifies the complete precedence chain:
	// defaults < config < environment < flags

	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// 1. Config sets processes to 5
	configPath := filepath.Join(tempDir, ".glide.yml")
	yamlContent := `
defaults:
  test:
    processes: 5
`
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// 2. Environment sets to 10
	os.Setenv("GLIDE_TEST_PROCESSES", "10")
	defer os.Unsetenv("GLIDE_TEST_PROCESSES")

	manager := NewManager()
	err = manager.Initialize(nil)
	require.NoError(t, err)

	// Before flags: env (10) should override config (5)
	cc := manager.GetCommandConfig()
	assert.Equal(t, 10, cc.Test.Processes)

	// 3. Flags set to 20
	cmd := &cobra.Command{}
	cmd.Flags().Int("processes", 0, "")
	cmd.Flags().Set("processes", "20")

	manager.ApplyFlags(cmd)
	cc = manager.GetCommandConfig()

	// Flags (20) should override env (10) and config (5)
	assert.Equal(t, 20, cc.Test.Processes)
}

func TestManager_GetLoader(t *testing.T) {
	manager := NewManager()
	loader := manager.GetLoader()

	assert.NotNil(t, loader)
}

func TestManager_BuildCommandConfig_NilConfig(t *testing.T) {
	manager := NewManager()

	// Don't initialize, so config is nil
	cc := manager.GetCommandConfig()

	// Should return defaults when config is nil
	assert.NotNil(t, cc)
	assert.Equal(t, 3, cc.Test.Processes)
}

func TestManager_EnvironmentVariables_Coverage(t *testing.T) {
	manager := NewManager()

	os.Setenv("GLIDE_TEST_COVERAGE", "true")
	defer os.Unsetenv("GLIDE_TEST_COVERAGE")

	err := manager.Initialize(nil)
	require.NoError(t, err)

	cc := manager.GetCommandConfig()
	assert.True(t, cc.Test.Coverage)
}

func TestManager_EnvironmentVariables_DockerAutoStart(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	manager := NewManager()

	os.Setenv("GLIDE_DOCKER_AUTO_START", "false")
	defer os.Unsetenv("GLIDE_DOCKER_AUTO_START")

	err := manager.Initialize(nil)
	require.NoError(t, err)

	cc := manager.GetCommandConfig()
	assert.False(t, cc.Docker.AutoStart)
}

func TestManager_EnvironmentVariables_WorktreeAutoSetup(t *testing.T) {
	manager := NewManager()

	os.Setenv("GLIDE_WORKTREE_AUTO_SETUP", "true")
	defer os.Unsetenv("GLIDE_WORKTREE_AUTO_SETUP")

	err := manager.Initialize(nil)
	require.NoError(t, err)

	cc := manager.GetCommandConfig()
	assert.True(t, cc.Worktree.AutoSetup)
}

func TestManager_ColorEnabled_GLIDECOLORSEnv(t *testing.T) {
	manager := NewManager()

	os.Setenv("GLIDE_COLORS", "always")
	defer os.Unsetenv("GLIDE_COLORS")

	err := manager.Initialize(nil)
	require.NoError(t, err)

	cc := manager.GetCommandConfig()
	assert.True(t, cc.Colors.Enabled)
}
