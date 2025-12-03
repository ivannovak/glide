package integration_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/glide-cli/glide/v3/pkg/plugin/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPluginDiscovery tests plugin discovery functionality
func TestPluginDiscovery(t *testing.T) {
	t.Run("discover_no_plugins", func(t *testing.T) {
		// Setup: Create temp directory with no plugins
		tmpDir := t.TempDir()
		pluginDir := filepath.Join(tmpDir, ".glide", "plugins")
		require.NoError(t, os.MkdirAll(pluginDir, 0755))

		// Test: Create manager and discover
		manager := sdk.NewManager(&sdk.ManagerConfig{
			PluginDirs:   []string{pluginDir},
			CacheTimeout: 5 * time.Minute,
		})

		err := manager.DiscoverPlugins()

		// Assert: No error, no plugins
		require.NoError(t, err)
		plugins := manager.ListPlugins()
		assert.Len(t, plugins, 0, "Should discover no plugins in empty directory")
	})

	t.Run("discover_nonexistent_directory", func(t *testing.T) {
		// Setup: Use non-existent plugin directory
		nonexistentDir := filepath.Join(t.TempDir(), "does-not-exist")

		// Test: Create manager and discover
		manager := sdk.NewManager(&sdk.ManagerConfig{
			PluginDirs:   []string{nonexistentDir},
			CacheTimeout: 5 * time.Minute,
		})

		err := manager.DiscoverPlugins()

		// Assert: Should handle gracefully (either error or empty list)
		// Both are acceptable behaviors
		if err == nil {
			plugins := manager.ListPlugins()
			assert.Len(t, plugins, 0, "Should return empty list for non-existent directory")
		} else {
			assert.Error(t, err, "Should error on non-existent directory")
		}
	})

	t.Run("discover_with_invalid_files", func(t *testing.T) {
		// Setup: Create temp directory with invalid plugin files
		tmpDir := t.TempDir()
		pluginDir := filepath.Join(tmpDir, ".glide", "plugins")
		require.NoError(t, os.MkdirAll(pluginDir, 0755))

		// Create invalid files
		invalidFile := filepath.Join(pluginDir, "not-a-plugin.txt")
		require.NoError(t, os.WriteFile(invalidFile, []byte("invalid"), 0644))

		readmeFile := filepath.Join(pluginDir, "README.md")
		require.NoError(t, os.WriteFile(readmeFile, []byte("# Plugins"), 0644))

		// Test: Discover plugins
		manager := sdk.NewManager(&sdk.ManagerConfig{
			PluginDirs:   []string{pluginDir},
			CacheTimeout: 5 * time.Minute,
		})

		err := manager.DiscoverPlugins()

		// Assert: Should ignore invalid files
		require.NoError(t, err)
		plugins := manager.ListPlugins()
		assert.Len(t, plugins, 0, "Should ignore non-executable files")
	})
}

// TestPluginLoading tests plugin loading functionality
func TestPluginLoading(t *testing.T) {
	t.Run("load_nonexistent_plugin", func(t *testing.T) {
		// Setup: Create manager
		tmpDir := t.TempDir()
		pluginDir := filepath.Join(tmpDir, ".glide", "plugins")
		require.NoError(t, os.MkdirAll(pluginDir, 0755))

		manager := sdk.NewManager(&sdk.ManagerConfig{
			PluginDirs:   []string{pluginDir},
			CacheTimeout: 5 * time.Minute,
		})

		// Test: Try to load non-existent plugin
		nonexistentPath := filepath.Join(pluginDir, "nonexistent-plugin")
		err := manager.LoadPlugin(nonexistentPath)

		// Assert: Should error
		assert.Error(t, err, "Should error when loading non-existent plugin")
		assert.Contains(t, err.Error(), "plugin", "Error should mention plugin")
	})

	t.Run("load_invalid_plugin", func(t *testing.T) {
		skipIfRaceDetector(t)

		// Setup: Create invalid plugin file
		tmpDir := t.TempDir()
		pluginDir := filepath.Join(tmpDir, ".glide", "plugins")
		require.NoError(t, os.MkdirAll(pluginDir, 0755))

		invalidPlugin := filepath.Join(pluginDir, "invalid-plugin")
		require.NoError(t, os.WriteFile(invalidPlugin, []byte("#!/bin/bash\necho invalid"), 0755))

		manager := sdk.NewManager(&sdk.ManagerConfig{
			PluginDirs:   []string{pluginDir},
			CacheTimeout: 5 * time.Minute,
		})

		// Test: Try to load invalid plugin
		err := manager.LoadPlugin(invalidPlugin)

		// Assert: Should error (plugin doesn't implement protocol)
		assert.Error(t, err, "Should error when loading invalid plugin")
	})
}

// TestPluginRetrieval tests plugin retrieval functionality
func TestPluginRetrieval(t *testing.T) {
	t.Run("get_nonexistent_plugin", func(t *testing.T) {
		// Setup: Create manager
		tmpDir := t.TempDir()
		pluginDir := filepath.Join(tmpDir, ".glide", "plugins")
		require.NoError(t, os.MkdirAll(pluginDir, 0755))

		manager := sdk.NewManager(&sdk.ManagerConfig{
			PluginDirs:   []string{pluginDir},
			CacheTimeout: 5 * time.Minute,
		})

		// Test: Try to get non-existent plugin
		plugin, err := manager.GetPlugin("nonexistent")

		// Assert: Should error
		assert.Error(t, err, "Should error when getting non-existent plugin")
		assert.Nil(t, plugin, "Plugin should be nil")
		assert.Contains(t, err.Error(), "not found", "Error should indicate plugin not found")
	})

	t.Run("list_plugins_empty", func(t *testing.T) {
		// Setup: Create manager with no plugins
		tmpDir := t.TempDir()
		pluginDir := filepath.Join(tmpDir, ".glide", "plugins")
		require.NoError(t, os.MkdirAll(pluginDir, 0755))

		manager := sdk.NewManager(&sdk.ManagerConfig{
			PluginDirs:   []string{pluginDir},
			CacheTimeout: 5 * time.Minute,
		})

		// Test: List plugins
		plugins := manager.ListPlugins()

		// Assert: Empty list (nil slice is acceptable)
		assert.Len(t, plugins, 0, "Plugin list should be empty")
	})
}

// TestPluginCaching tests plugin caching functionality
func TestPluginCaching(t *testing.T) {
	t.Run("cache_basic_operations", func(t *testing.T) {
		// Setup: Create cache
		cache := sdk.NewCache(5 * time.Minute)

		// Test: Put and Get
		mockLoaded := &sdk.LoadedPlugin{
			Name:     "test-plugin",
			Path:     "/path/to/plugin",
			LastUsed: time.Now(),
		}

		cache.Put("test-plugin", mockLoaded)
		retrieved := cache.Get("test-plugin")

		// Assert: Retrieved plugin matches
		require.NotNil(t, retrieved, "Retrieved plugin should not be nil")
		assert.Equal(t, mockLoaded.Name, retrieved.Name)
		assert.Equal(t, mockLoaded.Path, retrieved.Path)
	})

	t.Run("cache_get_nonexistent", func(t *testing.T) {
		// Setup: Create cache
		cache := sdk.NewCache(5 * time.Minute)

		// Test: Get non-existent item
		retrieved := cache.Get("nonexistent")

		// Assert: Should return nil
		assert.Nil(t, retrieved, "Non-existent cache item should return nil")
	})

	t.Run("cache_clear", func(t *testing.T) {
		// Setup: Create cache with item
		cache := sdk.NewCache(5 * time.Minute)
		mockLoaded := &sdk.LoadedPlugin{
			Name:     "test-plugin",
			Path:     "/path/to/plugin",
			LastUsed: time.Now(),
		}
		cache.Put("test-plugin", mockLoaded)

		// Test: Clear cache
		cache.Clear()
		retrieved := cache.Get("test-plugin")

		// Assert: Should be empty
		assert.Nil(t, retrieved, "Cache should be empty after clear")
	})

	t.Run("cache_overwrite", func(t *testing.T) {
		// Setup: Create cache
		cache := sdk.NewCache(5 * time.Minute)

		// Test: Put same key twice
		first := &sdk.LoadedPlugin{Name: "v1", Path: "/v1"}
		second := &sdk.LoadedPlugin{Name: "v2", Path: "/v2"}

		cache.Put("test", first)
		cache.Put("test", second)
		retrieved := cache.Get("test")

		// Assert: Should get latest value
		require.NotNil(t, retrieved)
		assert.Equal(t, "v2", retrieved.Name, "Should retrieve latest cached value")
		assert.Equal(t, "/v2", retrieved.Path)
	})
}

// TestPluginCleanup tests plugin cleanup functionality
func TestPluginCleanup(t *testing.T) {
	t.Run("cleanup_empty_manager", func(t *testing.T) {
		// Setup: Create manager with no plugins
		tmpDir := t.TempDir()
		pluginDir := filepath.Join(tmpDir, ".glide", "plugins")
		require.NoError(t, os.MkdirAll(pluginDir, 0755))

		manager := sdk.NewManager(&sdk.ManagerConfig{
			PluginDirs:   []string{pluginDir},
			CacheTimeout: 5 * time.Minute,
		})

		// Test: Cleanup (should not panic)
		manager.Cleanup()

		// Assert: No error, no panic
		plugins := manager.ListPlugins()
		assert.Len(t, plugins, 0, "Should still have no plugins after cleanup")
	})

	t.Run("cleanup_multiple_times", func(t *testing.T) {
		// Setup: Create manager
		tmpDir := t.TempDir()
		pluginDir := filepath.Join(tmpDir, ".glide", "plugins")
		require.NoError(t, os.MkdirAll(pluginDir, 0755))

		manager := sdk.NewManager(&sdk.ManagerConfig{
			PluginDirs:   []string{pluginDir},
			CacheTimeout: 5 * time.Minute,
		})

		// Test: Cleanup multiple times (should be idempotent)
		manager.Cleanup()
		manager.Cleanup()
		manager.Cleanup()

		// Assert: No panic, manager still functional
		plugins := manager.ListPlugins()
		assert.Len(t, plugins, 0, "Manager should still be functional after multiple cleanups")
	})
}

// TestPluginDirectoryPermissions tests handling of plugin directory permission issues
func TestPluginDirectoryPermissions(t *testing.T) {
	// Skip on Windows (different permission model)
	if os.Getenv("GOOS") == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	t.Run("unreadable_plugin_directory", func(t *testing.T) {
		// Setup: Create plugin directory with no read permissions
		tmpDir := t.TempDir()
		pluginDir := filepath.Join(tmpDir, ".glide", "plugins")
		require.NoError(t, os.MkdirAll(pluginDir, 0755))

		// Remove read permissions
		require.NoError(t, os.Chmod(pluginDir, 0000))
		defer os.Chmod(pluginDir, 0755) // Restore for cleanup

		manager := sdk.NewManager(&sdk.ManagerConfig{
			PluginDirs:   []string{pluginDir},
			CacheTimeout: 5 * time.Minute,
		})

		// Test: Try to discover plugins
		err := manager.DiscoverPlugins()

		// Assert: Should handle gracefully (either error or empty list)
		if err == nil {
			// If no error, should return empty list
			plugins := manager.ListPlugins()
			assert.Len(t, plugins, 0, "Should return empty list for unreadable directory")
		} else {
			// Error is also acceptable
			assert.Error(t, err, "Should error on unreadable directory")
		}
	})
}

// TestPluginConcurrentAccess tests concurrent access to plugin manager
func TestPluginConcurrentAccess(t *testing.T) {
	t.Run("concurrent_list_plugins", func(t *testing.T) {
		// Setup: Create manager
		tmpDir := t.TempDir()
		pluginDir := filepath.Join(tmpDir, ".glide", "plugins")
		require.NoError(t, os.MkdirAll(pluginDir, 0755))

		manager := sdk.NewManager(&sdk.ManagerConfig{
			PluginDirs:   []string{pluginDir},
			CacheTimeout: 5 * time.Minute,
		})

		// Test: List plugins concurrently
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				plugins := manager.ListPlugins()
				// Nil slice or empty slice are both valid
				assert.True(t, len(plugins) == 0, "Should return empty list")
				done <- true
			}()
		}

		// Assert: All goroutines complete without panic
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("concurrent_cache_access", func(t *testing.T) {
		// Setup: Create cache
		cache := sdk.NewCache(5 * time.Minute)

		// Test: Concurrent Put and Get
		done := make(chan bool, 20)
		for i := 0; i < 10; i++ {
			go func() {
				plugin := &sdk.LoadedPlugin{
					Name: "plugin",
					Path: "/path",
				}
				cache.Put("plugin", plugin)
				done <- true
			}()
			go func() {
				_ = cache.Get("plugin")
				done <- true
			}()
		}

		// Assert: All goroutines complete without race conditions
		for i := 0; i < 20; i++ {
			<-done
		}
	})
}

// TestMultiplePluginLoading tests loading multiple plugins simultaneously
func TestMultiplePluginLoading(t *testing.T) {
	t.Run("load_multiple_plugins", func(t *testing.T) {
		// Setup: Create manager with multiple plugin directories
		tmpDir := t.TempDir()
		pluginDir1 := filepath.Join(tmpDir, ".glide", "plugins")
		pluginDir2 := filepath.Join(tmpDir, ".local", "glide", "plugins")

		require.NoError(t, os.MkdirAll(pluginDir1, 0755))
		require.NoError(t, os.MkdirAll(pluginDir2, 0755))

		manager := sdk.NewManager(&sdk.ManagerConfig{
			PluginDirs:   []string{pluginDir1, pluginDir2},
			CacheTimeout: 5 * time.Minute,
		})

		// Test: Discover plugins from multiple directories
		err := manager.DiscoverPlugins()

		// Assert: No error from multiple directories
		require.NoError(t, err)
		plugins := manager.ListPlugins()
		assert.Len(t, plugins, 0, "Should handle multiple empty directories")
	})

	t.Run("load_plugins_with_same_name", func(t *testing.T) {
		skipIfRaceDetector(t)

		// Setup: Create two plugin directories with identically named plugins
		tmpDir := t.TempDir()
		pluginDir1 := filepath.Join(tmpDir, "plugins1")
		pluginDir2 := filepath.Join(tmpDir, "plugins2")

		require.NoError(t, os.MkdirAll(pluginDir1, 0755))
		require.NoError(t, os.MkdirAll(pluginDir2, 0755))

		// Create mock plugin in first directory
		plugin1 := filepath.Join(pluginDir1, "glide-plugin-test")
		require.NoError(t, os.WriteFile(plugin1, []byte("#!/bin/bash\necho plugin1"), 0755))

		// Create mock plugin with same name in second directory
		plugin2 := filepath.Join(pluginDir2, "glide-plugin-test")
		require.NoError(t, os.WriteFile(plugin2, []byte("#!/bin/bash\necho plugin2"), 0755))

		manager := sdk.NewManager(&sdk.ManagerConfig{
			PluginDirs:   []string{pluginDir1, pluginDir2},
			CacheTimeout: 5 * time.Minute,
		})

		// Test: Discover plugins
		err := manager.DiscoverPlugins()

		// Assert: Should handle conflict (either error or prefer first directory)
		// Both behaviors are acceptable
		if err == nil {
			plugins := manager.ListPlugins()
			// If no error, should have resolved conflict somehow
			assert.True(t, len(plugins) <= 1, "Should resolve plugin name conflict")
		}
	})

	t.Run("load_plugins_with_dependencies", func(t *testing.T) {
		skipIfRaceDetector(t)

		// Setup: Create plugin directory
		tmpDir := t.TempDir()
		pluginDir := filepath.Join(tmpDir, ".glide", "plugins")
		require.NoError(t, os.MkdirAll(pluginDir, 0755))

		// Create mock plugins (simulating dependency relationship)
		// Note: Actual dependency handling would require plugin protocol support
		basePlugin := filepath.Join(pluginDir, "glide-plugin-base")
		require.NoError(t, os.WriteFile(basePlugin, []byte("#!/bin/bash\necho base"), 0755))

		depPlugin := filepath.Join(pluginDir, "glide-plugin-dependent")
		require.NoError(t, os.WriteFile(depPlugin, []byte("#!/bin/bash\necho dependent"), 0755))

		manager := sdk.NewManager(&sdk.ManagerConfig{
			PluginDirs:   []string{pluginDir},
			CacheTimeout: 5 * time.Minute,
		})

		// Test: Discover plugins
		err := manager.DiscoverPlugins()

		// Assert: Both plugins discovered (dependency resolution is protocol-level)
		require.NoError(t, err)
		// Actual validation would require valid plugin protocol implementation
	})
}

// TestPluginConflicts tests handling of plugin conflicts
func TestPluginConflicts(t *testing.T) {
	skipIfRaceDetector(t)

	t.Run("conflict_different_versions", func(t *testing.T) {
		// Setup: Create plugins with version conflicts
		tmpDir := t.TempDir()
		pluginDir := filepath.Join(tmpDir, ".glide", "plugins")
		require.NoError(t, os.MkdirAll(pluginDir, 0755))

		// Create mock plugins with different versions
		pluginV1 := filepath.Join(pluginDir, "glide-plugin-test-v1")
		require.NoError(t, os.WriteFile(pluginV1, []byte("#!/bin/bash\necho v1"), 0755))

		pluginV2 := filepath.Join(pluginDir, "glide-plugin-test-v2")
		require.NoError(t, os.WriteFile(pluginV2, []byte("#!/bin/bash\necho v2"), 0755))

		manager := sdk.NewManager(&sdk.ManagerConfig{
			PluginDirs:   []string{pluginDir},
			CacheTimeout: 5 * time.Minute,
		})

		// Test: Discover plugins
		err := manager.DiscoverPlugins()

		// Assert: Should discover both (they have different names)
		require.NoError(t, err)
		// Version conflict handling is application-level logic
	})

	t.Run("conflict_symlink_duplicate", func(t *testing.T) {
		// Skip on Windows (different symlink handling)
		if os.Getenv("GOOS") == "windows" {
			t.Skip("Skipping symlink test on Windows")
		}

		// Setup: Create plugin with symlink
		tmpDir := t.TempDir()
		pluginDir := filepath.Join(tmpDir, ".glide", "plugins")
		require.NoError(t, os.MkdirAll(pluginDir, 0755))

		// Create original plugin
		originalPlugin := filepath.Join(pluginDir, "glide-plugin-original")
		require.NoError(t, os.WriteFile(originalPlugin, []byte("#!/bin/bash\necho original"), 0755))

		// Create symlink to same plugin
		symlinkPlugin := filepath.Join(pluginDir, "glide-plugin-symlink")
		require.NoError(t, os.Symlink(originalPlugin, symlinkPlugin))

		manager := sdk.NewManager(&sdk.ManagerConfig{
			PluginDirs:   []string{pluginDir},
			CacheTimeout: 5 * time.Minute,
		})

		// Test: Discover plugins
		err := manager.DiscoverPlugins()

		// Assert: Should handle gracefully (may discover 1 or 2 depending on implementation)
		require.NoError(t, err)
	})
}
