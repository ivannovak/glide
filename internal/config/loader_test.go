package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glide-cli/glide/v3/internal/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader_Load_WithDefaults(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	loader := NewLoader()
	cfg, err := loader.Load()
	require.NoError(t, err)

	// Should return defaults when no config exists
	assert.NotNil(t, cfg)
	assert.Equal(t, "auto", cfg.Defaults.Colors.Enabled)
	assert.True(t, cfg.Defaults.Test.Parallel)
	assert.Equal(t, 3, cfg.Defaults.Test.Processes)
}

func TestLoader_Load_AppliesDefaults(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Create config with partial settings
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

	loader := NewLoader()
	cfg, err := loader.Load()
	require.NoError(t, err)

	// Should have custom value
	assert.Equal(t, 5, cfg.Defaults.Test.Processes)
	// Should have default values for missing fields
	assert.Equal(t, "auto", cfg.Defaults.Colors.Enabled)
	assert.Equal(t, 30, cfg.Defaults.Docker.ComposeTimeout)
}

func TestLoader_Validate_InvalidProjectMode(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	configPath := filepath.Join(tempDir, ".glide.yml")
	yamlContent := `
projects:
  test:
    path: /test
    mode: invalid-mode
`
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	loader := NewLoader()
	_, err = loader.Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid mode")
}

func TestLoader_Validate_MissingProjectPath(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	configPath := filepath.Join(tempDir, ".glide.yml")
	yamlContent := `
projects:
  test:
    mode: multi-worktree
`
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	loader := NewLoader()
	_, err = loader.Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "has no path")
}

func TestLoader_Validate_InvalidTestProcesses(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	configPath := filepath.Join(tempDir, ".glide.yml")
	yamlContent := `
defaults:
  test:
    processes: 200
`
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	loader := NewLoader()
	_, err = loader.Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid test processes")
}

func TestLoader_Validate_InvalidColorSetting(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	configPath := filepath.Join(tempDir, ".glide.yml")
	yamlContent := `
defaults:
  colors:
    enabled: invalid
`
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	loader := NewLoader()
	_, err = loader.Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid color setting")
}

func TestLoader_Validate_NonExistentDefaultProject(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	configPath := filepath.Join(tempDir, ".glide.yml")
	yamlContent := `
projects:
  test:
    path: /test
default_project: nonexistent
`
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	loader := NewLoader()
	_, err = loader.Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "default project")
	assert.Contains(t, err.Error(), "does not exist")
}

func TestLoader_AddProject(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	loader := NewLoader()

	// Create test directory
	projectPath := filepath.Join(tempDir, "myproject")
	err := os.MkdirAll(projectPath, 0755)
	require.NoError(t, err)

	// Add project
	err = loader.AddProject("myproj", projectPath, "multi-worktree")
	require.NoError(t, err)

	// Verify it was saved
	cfg, err := loader.Load()
	require.NoError(t, err)

	assert.Contains(t, cfg.Projects, "myproj")
	assert.Contains(t, cfg.Projects["myproj"].Path, "myproject")
	assert.Equal(t, "multi-worktree", cfg.Projects["myproj"].Mode)
	assert.Equal(t, "myproj", cfg.DefaultProject, "First project should be set as default")
}

func TestLoader_LoadWithContext_MatchingProject(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Create config with a project
	configPath := filepath.Join(tempDir, ".glide.yml")
	yamlContent := `
projects:
  myproject:
    path: /home/user/myproject
    mode: multi-worktree
`
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	loader := NewLoader()
	ctx := &context.ProjectContext{
		ProjectRoot: "/home/user/myproject",
	}

	cfg, activeProject, err := loader.LoadWithContext(ctx)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.NotNil(t, activeProject)

	assert.Equal(t, "/home/user/myproject", activeProject.Path)
	assert.Equal(t, "multi-worktree", activeProject.Mode)
}

func TestLoader_LoadWithContext_NestedPath(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	configPath := filepath.Join(tempDir, ".glide.yml")
	yamlContent := `
projects:
  myproject:
    path: /home/user/myproject
`
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	loader := NewLoader()
	// Context is inside the project
	ctx := &context.ProjectContext{
		ProjectRoot: "/home/user/myproject/subdir",
	}

	cfg, activeProject, err := loader.LoadWithContext(ctx)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.NotNil(t, activeProject)

	assert.Equal(t, "/home/user/myproject", activeProject.Path)
}

func TestLoader_LoadWithContext_DefaultProject(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	configPath := filepath.Join(tempDir, ".glide.yml")
	yamlContent := `
projects:
  default:
    path: /home/default
  other:
    path: /home/other
default_project: default
`
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	loader := NewLoader()
	// Context doesn't match any project
	ctx := &context.ProjectContext{
		ProjectRoot: "/home/unmatched",
	}

	cfg, activeProject, err := loader.LoadWithContext(ctx)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.NotNil(t, activeProject)

	// Should fall back to default project
	assert.Equal(t, "/home/default", activeProject.Path)
}

func TestLoader_LoadWithContext_NoContext(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	configPath := filepath.Join(tempDir, ".glide.yml")
	yamlContent := `
projects:
  test:
    path: /test
`
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	loader := NewLoader()
	cfg, activeProject, err := loader.LoadWithContext(nil)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Should return nil for active project when no context
	assert.Nil(t, activeProject)
}

func TestLoader_ConfigExists(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	loader := NewLoader()

	// Should not exist initially
	assert.False(t, loader.ConfigExists())

	// Create config
	configPath := filepath.Join(tempDir, ".glide.yml")
	err := os.WriteFile(configPath, []byte("{}"), 0644)
	require.NoError(t, err)

	// Should exist now
	assert.True(t, loader.ConfigExists())
}

func TestLoader_GetConfigPath(t *testing.T) {
	loader := NewLoader()
	path := loader.GetConfigPath()

	assert.NotEmpty(t, path)
	assert.Contains(t, path, ".glide")
}

func TestLoader_Save_CreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Remove the temp dir to test directory creation
	configDir := filepath.Join(tempDir, ".glide")
	err := os.RemoveAll(configDir)
	require.NoError(t, err)

	loader := NewLoader()
	cfg := &Config{
		Projects: make(map[string]ProjectConfig),
	}

	err = loader.Save(cfg)
	require.NoError(t, err)

	// Verify directory was created
	info, err := os.Stat(filepath.Dir(loader.GetConfigPath()))
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestLoader_Validate_ValidModes(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	validModes := []string{"multi-worktree", "single-repo", ""}

	for _, mode := range validModes {
		t.Run("mode_"+mode, func(t *testing.T) {
			configPath := filepath.Join(tempDir, ".glide.yml")
			yamlContent := `
projects:
  test:
    path: /test
    mode: ` + mode + `
`
			err := os.WriteFile(configPath, []byte(yamlContent), 0644)
			require.NoError(t, err)

			loader := NewLoader()
			_, err = loader.Load()
			assert.NoError(t, err)
		})
	}
}

func TestLoader_Validate_ValidColorSettings(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	validColors := []string{"auto", "always", "never", ""}

	for _, color := range validColors {
		t.Run("color_"+color, func(t *testing.T) {
			configPath := filepath.Join(tempDir, ".glide.yml")
			yamlContent := `
defaults:
  colors:
    enabled: ` + color + `
`
			err := os.WriteFile(configPath, []byte(yamlContent), 0644)
			require.NoError(t, err)

			loader := NewLoader()
			_, err = loader.Load()
			assert.NoError(t, err)
		})
	}
}

func TestLoader_Validate_TestProcessesBoundaries(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	tests := []struct {
		name      string
		yaml      string
		shouldErr bool
	}{
		{"zero", "0", false}, // Zero will get default value applied
		{"one", "1", false},
		{"hundred", "100", false},
		{"over_hundred", "101", true},
		{"negative", "-1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := filepath.Join(tempDir, ".glide.yml")
			yamlContent := `
defaults:
  test:
    processes: ` + tt.yaml + `
`
			err := os.WriteFile(configPath, []byte(yamlContent), 0644)
			require.NoError(t, err)

			loader := NewLoader()
			_, err = loader.Load()

			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
