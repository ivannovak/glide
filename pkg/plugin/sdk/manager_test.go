package sdk

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindAncestorPluginDirs(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Create nested directories with .glide/plugins at different levels
	projectRoot := filepath.Join(tmpDir, "project")
	subDir1 := filepath.Join(projectRoot, "sub1")
	subDir2 := filepath.Join(subDir1, "sub2")

	// Create .glide/plugins directories at different levels
	require.NoError(t, os.MkdirAll(filepath.Join(projectRoot, ".glide", "plugins"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(subDir1, ".glide", "plugins"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(subDir2, "other"), 0755)) // subDir2 has no .glide/plugins

	// Change to the deepest directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)

	require.NoError(t, os.Chdir(subDir2))

	// Find ancestor plugin directories
	dirs := findAncestorPluginDirs()

	// Should find plugins in sub1 and project root, in that order (most specific first)
	assert.Len(t, dirs, 2)
	// Resolve symlinks for comparison on macOS
	expectedSub1, _ := filepath.EvalSymlinks(filepath.Join(subDir1, ".glide", "plugins"))
	expectedRoot, _ := filepath.EvalSymlinks(filepath.Join(projectRoot, ".glide", "plugins"))
	actualDir0, _ := filepath.EvalSymlinks(dirs[0])
	actualDir1, _ := filepath.EvalSymlinks(dirs[1])
	assert.Equal(t, expectedSub1, actualDir0)
	assert.Equal(t, expectedRoot, actualDir1)
}

func TestFindAncestorPluginDirsStopsAtHome(t *testing.T) {
	// This test ensures we don't look above the home directory
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	// Create a temp dir in home
	tmpDir, err := os.MkdirTemp(home, "glide-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create .glide/plugins in the temp dir
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".glide", "plugins"), 0755))

	// Change to temp dir
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)

	require.NoError(t, os.Chdir(tmpDir))

	// Find ancestor plugin directories
	dirs := findAncestorPluginDirs()

	// Should only find the one in tmpDir, not go above home
	assert.Len(t, dirs, 1)
	expectedPath, _ := filepath.EvalSymlinks(filepath.Join(tmpDir, ".glide", "plugins"))
	actualPath, _ := filepath.EvalSymlinks(dirs[0])
	assert.Equal(t, expectedPath, actualPath)
}

func TestDefaultConfigIncludesAncestorDirs(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "project")
	subDir := filepath.Join(projectRoot, "sub")

	// Create .glide/plugins directory at project root
	require.NoError(t, os.MkdirAll(filepath.Join(projectRoot, ".glide", "plugins"), 0755))
	require.NoError(t, os.MkdirAll(subDir, 0755))

	// Change to subdirectory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)

	require.NoError(t, os.Chdir(subDir))

	// Get default config
	config := DefaultConfig()

	// Should include the ancestor .glide/plugins directory
	found := false
	expectedPath, _ := filepath.EvalSymlinks(filepath.Join(projectRoot, ".glide", "plugins"))
	for _, dir := range config.PluginDirs {
		resolvedDir, _ := filepath.EvalSymlinks(dir)
		if resolvedDir == expectedPath {
			found = true
			break
		}
	}
	assert.True(t, found, "Should include ancestor .glide/plugins directory")
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		value    string
		expected bool
	}{
		{
			name:     "contains value",
			slice:    []string{"a", "b", "c"},
			value:    "b",
			expected: true,
		},
		{
			name:     "does not contain value",
			slice:    []string{"a", "b", "c"},
			value:    "d",
			expected: false,
		},
		{
			name:     "empty slice",
			slice:    []string{},
			value:    "a",
			expected: false,
		},
		{
			name:     "nil slice",
			slice:    nil,
			value:    "a",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDiscovererWithAncestorDirs(t *testing.T) {
	// Create a temporary directory structure with plugins at different levels
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "project")
	subDir := filepath.Join(projectRoot, "sub")

	// Create plugin directories
	projectPluginDir := filepath.Join(projectRoot, ".glide", "plugins")
	subPluginDir := filepath.Join(subDir, ".glide", "plugins")

	require.NoError(t, os.MkdirAll(projectPluginDir, 0755))
	require.NoError(t, os.MkdirAll(subPluginDir, 0755))

	// Create plugin executables
	projectPlugin := filepath.Join(projectPluginDir, "project-plugin")
	subPlugin := filepath.Join(subPluginDir, "sub-plugin")

	require.NoError(t, os.WriteFile(projectPlugin, []byte("#!/bin/sh\n"), 0755))
	require.NoError(t, os.WriteFile(subPlugin, []byte("#!/bin/sh\n"), 0755))

	// Change to subdirectory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)

	require.NoError(t, os.Chdir(subDir))

	// Create discoverer with ancestor dirs
	dirs := findAncestorPluginDirs()
	discoverer := NewDiscoverer(dirs)

	// Scan for plugins
	plugins, err := discoverer.Scan()
	require.NoError(t, err)

	// Should find both plugins, with sub plugin taking precedence
	assert.Len(t, plugins, 2)

	// Find the plugins by name
	var foundProject, foundSub bool
	for _, p := range plugins {
		if p.Name == "project" {
			foundProject = true
		}
		if p.Name == "sub" {
			foundSub = true
		}
	}

	assert.True(t, foundProject, "Should find project-level plugin")
	assert.True(t, foundSub, "Should find sub-directory plugin")
}
