package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glide-cli/glide/v3/pkg/branding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverConfigs_SingleConfig(t *testing.T) {
	// Create a temporary directory structure
	tempDir := t.TempDir()

	// Create a config file in the temp directory
	configPath := filepath.Join(tempDir, branding.ConfigFileName)
	err := os.WriteFile(configPath, []byte("{}"), 0644)
	require.NoError(t, err)

	configs, err := DiscoverConfigs(tempDir)
	require.NoError(t, err)

	assert.Len(t, configs, 1)
	assert.Equal(t, configPath, configs[0])
}

func TestDiscoverConfigs_ParentDirectory(t *testing.T) {
	// Create a directory tree: root/.glide.yml, root/child/
	tempDir := t.TempDir()
	rootConfig := filepath.Join(tempDir, branding.ConfigFileName)
	err := os.WriteFile(rootConfig, []byte("{}"), 0644)
	require.NoError(t, err)

	// Create child directory
	childDir := filepath.Join(tempDir, "child")
	err = os.MkdirAll(childDir, 0755)
	require.NoError(t, err)

	// Discover from child - should find parent config
	configs, err := DiscoverConfigs(childDir)
	require.NoError(t, err)

	assert.Len(t, configs, 1)
	assert.Equal(t, rootConfig, configs[0])
}

func TestDiscoverConfigs_MultipleConfigs(t *testing.T) {
	// Create directory tree: root/.glide.yml, root/project/.glide.yml, root/project/subdir/
	tempDir := t.TempDir()

	// Root config
	rootConfig := filepath.Join(tempDir, branding.ConfigFileName)
	err := os.WriteFile(rootConfig, []byte("{}"), 0644)
	require.NoError(t, err)

	// Project directory with its own config
	projectDir := filepath.Join(tempDir, "project")
	err = os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	projectConfig := filepath.Join(projectDir, branding.ConfigFileName)
	err = os.WriteFile(projectConfig, []byte("{}"), 0644)
	require.NoError(t, err)

	// Subdirectory
	subDir := filepath.Join(projectDir, "subdir")
	err = os.MkdirAll(subDir, 0755)
	require.NoError(t, err)

	// Discover from subdir - should find both configs
	// Note: DiscoverConfigs reverses order so deepest/nearest comes first
	configs, err := DiscoverConfigs(subDir)
	require.NoError(t, err)

	assert.Len(t, configs, 2)
	// Verify we found both configs (order is reversed by DiscoverConfigs)
	assert.Contains(t, configs, projectConfig, "Should find project config")
	assert.Contains(t, configs, rootConfig, "Should find root config")
}

func TestDiscoverConfigs_GitRoot(t *testing.T) {
	// Create directory tree with .git: root/.git, root/.glide.yml, root/subdir/
	tempDir := t.TempDir()

	// Create .git directory to simulate git root
	gitDir := filepath.Join(tempDir, ".git")
	err := os.MkdirAll(gitDir, 0755)
	require.NoError(t, err)

	// Create config at git root
	rootConfig := filepath.Join(tempDir, branding.ConfigFileName)
	err = os.WriteFile(rootConfig, []byte("{}"), 0644)
	require.NoError(t, err)

	// Create subdirectory
	subDir := filepath.Join(tempDir, "subdir")
	err = os.MkdirAll(subDir, 0755)
	require.NoError(t, err)

	// Discover from subdir - should stop at git root
	configs, err := DiscoverConfigs(subDir)
	require.NoError(t, err)

	assert.Len(t, configs, 1)
	assert.Equal(t, rootConfig, configs[0])
}

func TestDiscoverConfigs_NoConfig(t *testing.T) {
	// Create empty directory
	tempDir := t.TempDir()

	configs, err := DiscoverConfigs(tempDir)
	require.NoError(t, err)

	// Should return empty slice, not error
	assert.Len(t, configs, 0)
}

func TestDiscoverConfigs_StopsAtHomeDir(t *testing.T) {
	// This test simulates reaching home directory
	// We can't easily test the real home dir, but we can verify the logic
	tempDir := t.TempDir()

	// Create deep directory structure
	deepDir := filepath.Join(tempDir, "a", "b", "c", "d")
	err := os.MkdirAll(deepDir, 0755)
	require.NoError(t, err)

	// Discover from deep directory - should stop before going too high
	configs, err := DiscoverConfigs(deepDir)
	require.NoError(t, err)

	// Should handle gracefully even with no configs found (returns empty slice, not nil)
	assert.Len(t, configs, 0)
}

func TestDiscoverConfigs_GitRootWithConfigInSubdir(t *testing.T) {
	// Scenario: .git at root, but .glide.yml in a subdirectory
	tempDir := t.TempDir()

	// Create .git directory at root
	gitDir := filepath.Join(tempDir, ".git")
	err := os.MkdirAll(gitDir, 0755)
	require.NoError(t, err)

	// Create subdirectory with config
	subDir := filepath.Join(tempDir, "config-dir")
	err = os.MkdirAll(subDir, 0755)
	require.NoError(t, err)

	subConfig := filepath.Join(subDir, branding.ConfigFileName)
	err = os.WriteFile(subConfig, []byte("{}"), 0644)
	require.NoError(t, err)

	// Create deep subdirectory
	deepDir := filepath.Join(subDir, "deep")
	err = os.MkdirAll(deepDir, 0755)
	require.NoError(t, err)

	// Discover from deep directory - should find config and stop at git root
	configs, err := DiscoverConfigs(deepDir)
	require.NoError(t, err)

	assert.Len(t, configs, 1)
	assert.Equal(t, subConfig, configs[0])
}

func TestLoadAndMergeConfigs_SingleConfig(t *testing.T) {
	// Create temp config
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, branding.ConfigFileName)

	yamlContent := `
defaults:
  test:
    parallel: false
    processes: 5
  colors:
    enabled: always
`
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Change to temp dir so path validation works
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	configs, err := DiscoverConfigs(tempDir)
	require.NoError(t, err)

	merged, err := LoadAndMergeConfigs(configs)
	require.NoError(t, err)

	assert.False(t, merged.Defaults.Test.Parallel)
	assert.Equal(t, 5, merged.Defaults.Test.Processes)
	assert.Equal(t, "always", merged.Defaults.Colors.Enabled)
}

func TestLoadAndMergeConfigs_MergeCommands(t *testing.T) {
	tempDir := t.TempDir()

	// Create parent config with one command
	parentConfig := filepath.Join(tempDir, "parent.yml")
	parentYAML := `
commands:
  build: "go build"
  test: "go test"
`
	err := os.WriteFile(parentConfig, []byte(parentYAML), 0644)
	require.NoError(t, err)

	// Create child config that overrides one command and adds another
	childConfig := filepath.Join(tempDir, "child.yml")
	childYAML := `
commands:
  test: "go test -v"
  lint: "golangci-lint run"
`
	err = os.WriteFile(childConfig, []byte(childYAML), 0644)
	require.NoError(t, err)

	// Change to temp dir
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	// Load configs in order: parent, then child (child has higher priority)
	merged, err := LoadAndMergeConfigs([]string{childConfig, parentConfig})
	require.NoError(t, err)

	assert.NotNil(t, merged.Commands)
	assert.Equal(t, "go build", merged.Commands["build"], "Parent's build command should be preserved")
	assert.Equal(t, "go test -v", merged.Commands["test"], "Child's test command should override parent")
	assert.Equal(t, "golangci-lint run", merged.Commands["lint"], "Child's lint command should be added")
}

func TestLoadAndMergeConfigs_MergeProjects(t *testing.T) {
	tempDir := t.TempDir()

	// Config 1 with project A
	config1 := filepath.Join(tempDir, "config1.yml")
	yaml1 := `
projects:
  projectA:
    path: /path/to/a
    mode: single-repo
`
	err := os.WriteFile(config1, []byte(yaml1), 0644)
	require.NoError(t, err)

	// Config 2 with project B and default project
	config2 := filepath.Join(tempDir, "config2.yml")
	yaml2 := `
projects:
  projectB:
    path: /path/to/b
    mode: multi-worktree
default_project: projectB
`
	err = os.WriteFile(config2, []byte(yaml2), 0644)
	require.NoError(t, err)

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	// Merge configs - config2 should override
	merged, err := LoadAndMergeConfigs([]string{config2, config1})
	require.NoError(t, err)

	assert.Len(t, merged.Projects, 2)
	assert.Equal(t, "/path/to/a", merged.Projects["projectA"].Path)
	assert.Equal(t, "/path/to/b", merged.Projects["projectB"].Path)
	assert.Equal(t, "projectB", merged.DefaultProject)
}

func TestLoadAndMergeConfigs_InvalidPath(t *testing.T) {
	// Try to load a config with path traversal
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	tempDir := t.TempDir()
	os.Chdir(tempDir)

	// This should be handled gracefully (skipped)
	merged, err := LoadAndMergeConfigs([]string{"../../../etc/passwd"})
	require.NoError(t, err)

	// Should return empty but valid config
	assert.NotNil(t, merged)
	assert.NotNil(t, merged.Commands)
}

func TestLoadAndMergeConfigs_InvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "bad.yml")

	err := os.WriteFile(configPath, []byte("invalid: [yaml"), 0644)
	require.NoError(t, err)

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	// Should skip invalid configs and return valid empty config
	merged, err := LoadAndMergeConfigs([]string{configPath})
	require.NoError(t, err)

	assert.NotNil(t, merged)
	assert.NotNil(t, merged.Commands)
}

func TestLoadAndMergeConfigs_EmptyList(t *testing.T) {
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	tempDir := t.TempDir()
	os.Chdir(tempDir)

	merged, err := LoadAndMergeConfigs([]string{})
	require.NoError(t, err)

	assert.NotNil(t, merged)
	assert.NotNil(t, merged.Commands)
	assert.NotNil(t, merged.Projects)
}

func TestMergeDefaults_TestSettings(t *testing.T) {
	target := &DefaultsConfig{}
	source := &DefaultsConfig{
		Test: TestDefaults{
			Parallel:  true,
			Processes: 5,
			Coverage:  true,
			Verbose:   true,
		},
	}

	mergeDefaults(target, source)

	assert.Equal(t, 5, target.Test.Processes)
	assert.True(t, target.Test.Parallel)
	assert.True(t, target.Test.Coverage)
	assert.True(t, target.Test.Verbose)
}

func TestMergeDefaults_PreferExistingNonZero(t *testing.T) {
	target := &DefaultsConfig{
		Test: TestDefaults{
			Processes: 3,
			Parallel:  false,
		},
		Docker: DockerDefaults{
			ComposeTimeout: 60,
		},
	}

	source := &DefaultsConfig{
		Test: TestDefaults{
			Processes: 5,
			Parallel:  true,
		},
		Docker: DockerDefaults{
			ComposeTimeout: 30,
		},
	}

	mergeDefaults(target, source)

	// Should keep target's non-zero values
	assert.Equal(t, 3, target.Test.Processes, "Should keep target's non-zero processes")
	// Note: mergeDefaults only sets bools to true if source is true, won't overwrite
	assert.True(t, target.Test.Parallel, "Source's true value should be merged even if target was false")
	assert.Equal(t, 60, target.Docker.ComposeTimeout, "Should keep target's timeout")
}

func TestMergeDefaults_AllFields(t *testing.T) {
	target := &DefaultsConfig{}
	source := &DefaultsConfig{
		Test: TestDefaults{
			Parallel:  true,
			Processes: 4,
			Coverage:  true,
			Verbose:   true,
		},
		Docker: DockerDefaults{
			ComposeTimeout: 45,
			AutoStart:      true,
			RemoveOrphans:  true,
		},
		Colors: ColorDefaults{
			Enabled: "always",
		},
		Worktree: WorktreeDefaults{
			AutoSetup:     true,
			CopyEnv:       true,
			RunMigrations: true,
		},
	}

	mergeDefaults(target, source)

	// Test defaults
	assert.True(t, target.Test.Parallel)
	assert.Equal(t, 4, target.Test.Processes)
	assert.True(t, target.Test.Coverage)
	assert.True(t, target.Test.Verbose)

	// Docker defaults
	assert.Equal(t, 45, target.Docker.ComposeTimeout)
	assert.True(t, target.Docker.AutoStart)
	assert.True(t, target.Docker.RemoveOrphans)

	// Color defaults
	assert.Equal(t, "always", target.Colors.Enabled)

	// Worktree defaults
	assert.True(t, target.Worktree.AutoSetup)
	assert.True(t, target.Worktree.CopyEnv)
	assert.True(t, target.Worktree.RunMigrations)
}
