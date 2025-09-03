package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Defaults(t *testing.T) {
	cfg := GetDefaults()

	assert.NotNil(t, cfg.Projects)
	assert.Equal(t, "auto", cfg.Defaults.Colors.Enabled)
	assert.True(t, cfg.Defaults.Test.Parallel)
	assert.Equal(t, 3, cfg.Defaults.Test.Processes)
	assert.False(t, cfg.Defaults.Test.Coverage)
	assert.True(t, cfg.Defaults.Docker.AutoStart)
	assert.Equal(t, 30, cfg.Defaults.Docker.ComposeTimeout)
}

func TestProjectConfig(t *testing.T) {
	cfg := ProjectConfig{
		Path: "/home/user/project",
		Mode: "multi-worktree",
	}

	assert.Equal(t, "/home/user/project", cfg.Path)
	assert.Equal(t, "multi-worktree", cfg.Mode)
}

func TestCommandConfig(t *testing.T) {
	cfg := CommandConfig{
		Test: TestConfig{
			Parallel:  true,
			Processes: 4,
			Coverage:  true,
			Verbose:   true,
			Args:      []string{"--race"},
		},
		Docker: DockerConfig{
			ComposeTimeout: 60,
			AutoStart:      false,
			RemoveOrphans:  true,
			ComposeFiles:   []string{"docker-compose.yml"},
		},
		Colors: ColorConfig{
			Enabled: true,
		},
		Worktree: WorktreeConfig{
			AutoSetup:     true,
			CopyEnv:       false,
			RunMigrations: true,
		},
		ActiveProject: &ProjectConfig{
			Path: "/test/project",
			Mode: "single-repo",
		},
	}

	assert.True(t, cfg.Test.Parallel)
	assert.Equal(t, 4, cfg.Test.Processes)
	assert.True(t, cfg.Test.Coverage)
	assert.Contains(t, cfg.Test.Args, "--race")
	
	assert.Equal(t, 60, cfg.Docker.ComposeTimeout)
	assert.False(t, cfg.Docker.AutoStart)
	assert.True(t, cfg.Docker.RemoveOrphans)
	assert.Contains(t, cfg.Docker.ComposeFiles, "docker-compose.yml")
	
	assert.True(t, cfg.Colors.Enabled)
	
	assert.True(t, cfg.Worktree.AutoSetup)
	assert.False(t, cfg.Worktree.CopyEnv)
	assert.True(t, cfg.Worktree.RunMigrations)
	
	assert.NotNil(t, cfg.ActiveProject)
	assert.Equal(t, "/test/project", cfg.ActiveProject.Path)
	assert.Equal(t, "single-repo", cfg.ActiveProject.Mode)
}

func TestLoader_Load(t *testing.T) {
	// Set HOME to temp directory for testing
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Create .glide.yml in HOME directory
	configPath := filepath.Join(tempDir, ".glide.yml")
	yamlContent := `
projects:
  myproject:
    path: /home/user/myproject
    mode: multi-worktree
default_project: myproject
defaults:
  test:
    parallel: false
    processes: 2
  docker:
    compose_timeout: 60
  colors:
    enabled: always
`
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	loader := NewLoader()
	cfg, err := loader.Load()
	require.NoError(t, err)

	assert.Equal(t, "myproject", cfg.DefaultProject)
	assert.Equal(t, "/home/user/myproject", cfg.Projects["myproject"].Path)
	assert.Equal(t, "multi-worktree", cfg.Projects["myproject"].Mode)
	assert.False(t, cfg.Defaults.Test.Parallel)
	assert.Equal(t, 2, cfg.Defaults.Test.Processes)
	assert.Equal(t, 60, cfg.Defaults.Docker.ComposeTimeout)
	assert.Equal(t, "always", cfg.Defaults.Colors.Enabled)
}

func TestLoader_LoadNonExistentFile(t *testing.T) {
	// Set HOME to temp directory for testing
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	loader := NewLoader()
	cfg, err := loader.Load()

	// Should return default config when file doesn't exist
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.Projects)
}

func TestLoader_LoadInvalidYAML(t *testing.T) {
	// Set HOME to temp directory for testing
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	configPath := filepath.Join(tempDir, ".glide.yml")
	invalidYAML := `
projects:
  myproject: [this is invalid
`
	err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
	require.NoError(t, err)

	loader := NewLoader()
	_, err = loader.Load()
	assert.Error(t, err)
}

func TestLoader_Save(t *testing.T) {
	// Set HOME to temp directory for testing
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	cfg := &Config{
		Projects: map[string]ProjectConfig{
			"test": {
				Path: "/test/path",
				Mode: "single-repo",
			},
		},
		DefaultProject: "test",
		Defaults: DefaultsConfig{
			Test: TestDefaults{
				Parallel:  true,
				Processes: 4,
			},
		},
	}

	loader := NewLoader()
	err := loader.Save(cfg)
	require.NoError(t, err)

	// Read back and verify
	loadedCfg, err := loader.Load()
	require.NoError(t, err)

	assert.Equal(t, cfg.DefaultProject, loadedCfg.DefaultProject)
	assert.Equal(t, cfg.Projects["test"].Path, loadedCfg.Projects["test"].Path)
	assert.Equal(t, cfg.Defaults.Test.Parallel, loadedCfg.Defaults.Test.Parallel)
}