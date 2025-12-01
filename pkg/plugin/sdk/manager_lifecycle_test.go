package sdk

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewManager tests manager creation with various configurations
func TestNewManager(t *testing.T) {
	tests := []struct {
		name   string
		config *ManagerConfig
	}{
		{
			name:   "nil config uses defaults",
			config: nil,
		},
		{
			name: "custom config",
			config: &ManagerConfig{
				PluginDirs:     []string{"/custom/path"},
				CacheTimeout:   10 * time.Minute,
				MaxPlugins:     20,
				EnableDebug:    true,
				SecurityStrict: false,
			},
		},
		{
			name: "minimal config",
			config: &ManagerConfig{
				PluginDirs: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.config)
			require.NotNil(t, m)
			assert.NotNil(t, m.plugins)
			assert.NotNil(t, m.discoverer)
			assert.NotNil(t, m.validator)
			assert.NotNil(t, m.cache)
			assert.NotNil(t, m.config)

			if tt.config == nil {
				// Should use default config
				assert.NotNil(t, m.config.PluginDirs)
				assert.Greater(t, len(m.config.PluginDirs), 0)
			} else {
				// Should use provided config
				assert.Equal(t, tt.config.PluginDirs, m.config.PluginDirs)
			}
		})
	}
}

// TestDefaultConfig tests default configuration generation
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	require.NotNil(t, config)
	assert.NotEmpty(t, config.PluginDirs, "Should have at least one plugin directory")
	assert.Greater(t, config.CacheTimeout, time.Duration(0), "Cache timeout should be positive")
	assert.Greater(t, config.MaxPlugins, 0, "Max plugins should be positive")
	assert.True(t, config.SecurityStrict, "Security should be strict by default")
}

// TestCache tests the plugin cache functionality
func TestCache(t *testing.T) {
	cache := NewCache(5 * time.Minute)
	require.NotNil(t, cache)

	t.Run("empty cache", func(t *testing.T) {
		result := cache.Get("/nonexistent/plugin")
		assert.Nil(t, result)
	})

	t.Run("put and get", func(t *testing.T) {
		testPlugin := &LoadedPlugin{
			Name: "test-plugin",
			Path: "/test/path",
		}

		cache.Put("/test/path", testPlugin)

		retrieved := cache.Get("/test/path")
		require.NotNil(t, retrieved)
		assert.Equal(t, "test-plugin", retrieved.Name)
		assert.Equal(t, "/test/path", retrieved.Path)
	})

	t.Run("overwrite existing", func(t *testing.T) {
		plugin1 := &LoadedPlugin{Name: "plugin1", Path: "/path"}
		plugin2 := &LoadedPlugin{Name: "plugin2", Path: "/path"}

		cache.Put("/path", plugin1)
		cache.Put("/path", plugin2)

		retrieved := cache.Get("/path")
		require.NotNil(t, retrieved)
		assert.Equal(t, "plugin2", retrieved.Name)
	})

	t.Run("clear cache", func(t *testing.T) {
		cache.Put("/path1", &LoadedPlugin{Name: "p1"})
		cache.Put("/path2", &LoadedPlugin{Name: "p2"})

		cache.Clear()

		assert.Nil(t, cache.Get("/path1"))
		assert.Nil(t, cache.Get("/path2"))
	})

	t.Run("concurrent access", func(t *testing.T) {
		// Test concurrent Put/Get operations
		done := make(chan bool)

		go func() {
			for i := 0; i < 100; i++ {
				path := fmt.Sprintf("/path/%d", i)
				cache.Put(path, &LoadedPlugin{Name: fmt.Sprintf("plugin%d", i)})
			}
			done <- true
		}()

		go func() {
			for i := 0; i < 100; i++ {
				path := fmt.Sprintf("/path/%d", i)
				cache.Get(path)
			}
			done <- true
		}()

		<-done
		<-done
		// If we got here without panicking, concurrent access works
	})
}

// buildTestPlugin creates a minimal test plugin binary
func buildTestPlugin(t *testing.T, dir, name string, shouldSucceed bool) string {
	t.Helper()

	pluginSrc := `package main

import (
	"context"
	"github.com/hashicorp/go-plugin"
	sdk "github.com/ivannovak/glide/v3/pkg/plugin/sdk/v1"
)

func main() {
	basePlugin := sdk.NewBasePlugin(&sdk.PluginMetadata{
		Name:        "%s",
		Version:     "1.0.0",
		Author:      "Test",
		Description: "Test plugin",
		MinSdk:      "v1.0.0",
	})

	basePlugin.RegisterCommand("test", sdk.NewSimpleCommand(
		&sdk.CommandInfo{
			Name:        "test",
			Description: "Test command",
			Category:    sdk.CategoryDeveloper,
		},
		func(ctx context.Context, req *sdk.ExecuteRequest) (*sdk.ExecuteResponse, error) {
			return &sdk.ExecuteResponse{
				Success: true,
				Stdout:  []byte("test output\n"),
			}, nil
		},
	))

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: sdk.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"glide": &sdk.GlidePluginImpl{Impl: basePlugin},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
`

	// Get absolute path to glide module root (3 levels up from pkg/plugin/sdk)
	cwd, err := os.Getwd()
	require.NoError(t, err)
	glideRoot := filepath.Join(cwd, "../../..")
	glideRoot, err = filepath.Abs(glideRoot)
	require.NoError(t, err)

	goModContent := fmt.Sprintf(`module testplugin

go 1.23

replace github.com/ivannovak/glide/v3 => %s

require github.com/ivannovak/glide/v3 v3.0.0
`, glideRoot)

	srcPath := filepath.Join(dir, "main.go")
	modPath := filepath.Join(dir, "go.mod")
	binPath := filepath.Join(dir, name)

	// Write go.mod
	err = os.WriteFile(modPath, []byte(goModContent), 0644)
	require.NoError(t, err)

	// Write source
	err = os.WriteFile(srcPath, []byte(fmt.Sprintf(pluginSrc, name)), 0644)
	require.NoError(t, err)

	// Run go mod tidy to get dependencies
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = dir
	tidyOutput, err := tidyCmd.CombinedOutput()
	if err != nil {
		t.Logf("go mod tidy output: %s", string(tidyOutput))
	}
	require.NoError(t, err, "Failed to run go mod tidy: %s", string(tidyOutput))

	// Build plugin
	cmd := exec.Command("go", "build", "-o", binPath, srcPath)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()

	if shouldSucceed {
		require.NoError(t, err, "Failed to build test plugin: %s", string(output))

		// Make executable
		err = os.Chmod(binPath, 0755)
		require.NoError(t, err)

		return binPath
	}

	return binPath
}

// TestLoadPlugin tests loading individual plugins
func TestLoadPlugin(t *testing.T) {
	t.Run("load valid plugin", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping plugin build test in short mode")
		}

		tmpDir := t.TempDir()
		pluginPath := buildTestPlugin(t, tmpDir, "test-plugin", true)

		config := &ManagerConfig{
			PluginDirs:     []string{tmpDir},
			SecurityStrict: false, // Disable strict mode for testing
		}
		m := NewManager(config)

		err := m.LoadPlugin(pluginPath)
		assert.NoError(t, err)

		// Verify plugin was loaded
		plugin, err := m.GetPlugin("test-plugin")
		require.NoError(t, err)
		assert.Equal(t, "test-plugin", plugin.Name)
		assert.Equal(t, "1.0.0", plugin.Metadata.Version)
	})

	t.Run("load nonexistent plugin", func(t *testing.T) {
		m := NewManager(nil)

		err := m.LoadPlugin("/nonexistent/plugin")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("load plugin from untrusted path in strict mode", func(t *testing.T) {
		trustedDir := t.TempDir()
		untrustedDir := t.TempDir() // Separate temp directory, not inside trustedDir

		// Create a simple executable in the untrusted directory
		pluginPath := filepath.Join(untrustedDir, "plugin")
		err := os.WriteFile(pluginPath, []byte("#!/bin/sh\necho test"), 0755)
		require.NoError(t, err)

		config := &ManagerConfig{
			PluginDirs:     []string{trustedDir}, // Only trustedDir is trusted
			SecurityStrict: true,
		}
		m := NewManager(config)

		err = m.LoadPlugin(pluginPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("load cached plugin", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping plugin build test in short mode")
		}

		tmpDir := t.TempDir()
		pluginPath := buildTestPlugin(t, tmpDir, "cached-plugin", true)

		config := &ManagerConfig{
			PluginDirs:     []string{tmpDir},
			SecurityStrict: false,
		}
		m := NewManager(config)

		// Load first time
		err := m.LoadPlugin(pluginPath)
		require.NoError(t, err)

		plugin1, err := m.GetPlugin("cached-plugin")
		require.NoError(t, err)

		// Load again (should use cache)
		err = m.LoadPlugin(pluginPath)
		require.NoError(t, err)

		plugin2, err := m.GetPlugin("cached-plugin")
		require.NoError(t, err)

		// Should be the same instance from cache
		assert.Equal(t, plugin1.LastUsed.Unix(), plugin2.LastUsed.Unix())
	})
}

// TestGetPlugin tests retrieving loaded plugins
func TestGetPlugin(t *testing.T) {
	t.Run("get nonexistent plugin", func(t *testing.T) {
		m := NewManager(nil)

		plugin, err := m.GetPlugin("nonexistent")
		assert.Error(t, err)
		assert.Nil(t, plugin)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("get loaded plugin updates last used", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping plugin build test in short mode")
		}

		tmpDir := t.TempDir()
		pluginPath := buildTestPlugin(t, tmpDir, "time-test", true)

		config := &ManagerConfig{
			PluginDirs:     []string{tmpDir},
			SecurityStrict: false,
		}
		m := NewManager(config)

		err := m.LoadPlugin(pluginPath)
		require.NoError(t, err)

		// Get plugin first time
		plugin1, err := m.GetPlugin("time-test")
		require.NoError(t, err)
		time1 := plugin1.LastUsed

		// Wait a bit
		time.Sleep(10 * time.Millisecond)

		// Get plugin again
		plugin2, err := m.GetPlugin("time-test")
		require.NoError(t, err)
		time2 := plugin2.LastUsed

		// LastUsed should be updated
		assert.True(t, time2.After(time1), "LastUsed should be updated on Get")
	})
}

// TestExecuteCommand tests command execution through the manager
func TestExecuteCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping plugin build test in short mode")
	}

	tmpDir := t.TempDir()
	pluginPath := buildTestPlugin(t, tmpDir, "exec-test", true)

	config := &ManagerConfig{
		PluginDirs:     []string{tmpDir},
		SecurityStrict: false,
	}
	m := NewManager(config)

	err := m.LoadPlugin(pluginPath)
	require.NoError(t, err)

	t.Run("execute valid command", func(t *testing.T) {
		// ExecuteCommand prints to stdout, so we test it doesn't error
		err := m.ExecuteCommand("exec-test", "test", []string{})
		assert.NoError(t, err)
	})

	t.Run("execute unknown command", func(t *testing.T) {
		err := m.ExecuteCommand("exec-test", "unknown", []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("execute on nonexistent plugin", func(t *testing.T) {
		err := m.ExecuteCommand("nonexistent", "test", []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

// TestListPlugins tests listing all loaded plugins
func TestListPlugins(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping plugin build test in short mode")
	}

	tmpDir := t.TempDir()

	// Build multiple plugins
	plugin1Path := buildTestPlugin(t, tmpDir, "plugin1", true)
	plugin2Path := buildTestPlugin(t, tmpDir, "plugin2", true)

	config := &ManagerConfig{
		PluginDirs:     []string{tmpDir},
		SecurityStrict: false,
	}
	m := NewManager(config)

	// Load plugins
	err := m.LoadPlugin(plugin1Path)
	require.NoError(t, err)
	err = m.LoadPlugin(plugin2Path)
	require.NoError(t, err)

	// List plugins
	plugins := m.ListPlugins()

	assert.Len(t, plugins, 2)

	names := make(map[string]bool)
	for _, p := range plugins {
		names[p.Name] = true
	}

	assert.True(t, names["plugin1"])
	assert.True(t, names["plugin2"])
}

// TestCleanup tests plugin cleanup and resource management
func TestCleanup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping plugin build test in short mode")
	}

	tmpDir := t.TempDir()
	pluginPath := buildTestPlugin(t, tmpDir, "cleanup-test", true)

	config := &ManagerConfig{
		PluginDirs:     []string{tmpDir},
		SecurityStrict: false,
	}
	m := NewManager(config)

	err := m.LoadPlugin(pluginPath)
	require.NoError(t, err)

	// Verify plugin is running
	plugin, err := m.GetPlugin("cleanup-test")
	require.NoError(t, err)
	assert.False(t, plugin.Client.Exited())

	// Cleanup
	m.Cleanup()

	// Plugin should be killed
	assert.True(t, plugin.Client.Exited(), "Plugin should be killed after cleanup")

	// Manager should have no plugins
	plugins := m.ListPlugins()
	assert.Len(t, plugins, 0, "Manager should have no plugins after cleanup")
}

// TestDiscoverPlugins tests automatic plugin discovery
func TestDiscoverPlugins(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping plugin build test in short mode")
	}

	tmpDir := t.TempDir()
	pluginDir := filepath.Join(tmpDir, "plugins")
	require.NoError(t, os.MkdirAll(pluginDir, 0755))

	// Build test plugins
	buildTestPlugin(t, pluginDir, "discovered1", true)
	buildTestPlugin(t, pluginDir, "discovered2", true)

	config := &ManagerConfig{
		PluginDirs:     []string{pluginDir},
		SecurityStrict: false,
	}
	m := NewManager(config)

	err := m.DiscoverPlugins()
	require.NoError(t, err)

	// Should have discovered both plugins
	plugins := m.ListPlugins()
	assert.Len(t, plugins, 2)

	names := make(map[string]bool)
	for _, p := range plugins {
		names[p.Name] = true
	}

	assert.True(t, names["discovered1"])
	assert.True(t, names["discovered2"])
}

// TestPluginRegistrationDuplicates tests handling duplicate plugin names
func TestPluginRegistrationDuplicates(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping plugin build test in short mode")
	}

	tmpDir := t.TempDir()
	dir1 := filepath.Join(tmpDir, "dir1")
	dir2 := filepath.Join(tmpDir, "dir2")
	require.NoError(t, os.MkdirAll(dir1, 0755))
	require.NoError(t, os.MkdirAll(dir2, 0755))

	// Build plugins with same name in different directories
	plugin1Path := buildTestPlugin(t, dir1, "duplicate", true)
	buildTestPlugin(t, dir2, "duplicate", true)

	config := &ManagerConfig{
		PluginDirs:     []string{dir1, dir2},
		SecurityStrict: false,
		EnableDebug:    false, // Disable debug to avoid log spam
	}
	m := NewManager(config)

	// Load first plugin
	err := m.LoadPlugin(plugin1Path)
	require.NoError(t, err)

	// Discovery should skip already-loaded plugins
	err = m.DiscoverPlugins()
	require.NoError(t, err)

	// Should only have one plugin (first one loaded)
	plugins := m.ListPlugins()
	assert.Len(t, plugins, 1)
	assert.Equal(t, "duplicate", plugins[0].Name)
}
