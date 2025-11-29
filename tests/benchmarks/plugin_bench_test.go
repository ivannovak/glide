package benchmarks_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ivannovak/glide/v2/pkg/plugin/sdk"
)

// BenchmarkPluginDiscovery benchmarks plugin discovery performance
func BenchmarkPluginDiscovery(b *testing.B) {
	// Setup: Create plugin directory with mock plugins
	tmpDir := b.TempDir()
	pluginDir := filepath.Join(tmpDir, ".glide", "plugins")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		b.Fatal(err)
	}

	// Create 10 mock plugin files
	for i := 0; i < 10; i++ {
		pluginPath := filepath.Join(pluginDir, "glide-plugin-test"+string(rune('a'+i)))
		if err := os.WriteFile(pluginPath, []byte("#!/bin/bash\necho test"), 0755); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager := sdk.NewManager(&sdk.ManagerConfig{
			PluginDirs:   []string{pluginDir},
			CacheTimeout: 5 * time.Minute,
		})
		_ = manager.DiscoverPlugins()
	}
}

// BenchmarkPluginDiscoveryEmpty benchmarks discovery in empty directory
func BenchmarkPluginDiscoveryEmpty(b *testing.B) {
	tmpDir := b.TempDir()
	pluginDir := filepath.Join(tmpDir, ".glide", "plugins")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager := sdk.NewManager(&sdk.ManagerConfig{
			PluginDirs:   []string{pluginDir},
			CacheTimeout: 5 * time.Minute,
		})
		_ = manager.DiscoverPlugins()
	}
}

// BenchmarkPluginDiscoveryLarge benchmarks discovery with many plugins
func BenchmarkPluginDiscoveryLarge(b *testing.B) {
	tmpDir := b.TempDir()
	pluginDir := filepath.Join(tmpDir, ".glide", "plugins")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		b.Fatal(err)
	}

	// Create 100 mock plugin files
	for i := 0; i < 100; i++ {
		pluginPath := filepath.Join(pluginDir, "glide-plugin-test", string(rune(i)))
		if err := os.WriteFile(pluginPath, []byte("#!/bin/bash\necho test"), 0755); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager := sdk.NewManager(&sdk.ManagerConfig{
			PluginDirs:   []string{pluginDir},
			CacheTimeout: 5 * time.Minute,
		})
		_ = manager.DiscoverPlugins()
	}
}

// BenchmarkPluginListPlugins benchmarks listing plugins
func BenchmarkPluginListPlugins(b *testing.B) {
	tmpDir := b.TempDir()
	pluginDir := filepath.Join(tmpDir, ".glide", "plugins")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		b.Fatal(err)
	}

	manager := sdk.NewManager(&sdk.ManagerConfig{
		PluginDirs:   []string{pluginDir},
		CacheTimeout: 5 * time.Minute,
	})

	_ = manager.DiscoverPlugins()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = manager.ListPlugins()
	}
}

// BenchmarkPluginCachePut benchmarks cache put operations
func BenchmarkPluginCachePut(b *testing.B) {
	cache := sdk.NewCache(5 * time.Minute)
	plugin := &sdk.LoadedPlugin{
		Name:     "test-plugin",
		Path:     "/path/to/plugin",
		LastUsed: time.Now(),
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache.Put("test-plugin", plugin)
	}
}

// BenchmarkPluginCacheGet benchmarks cache get operations
func BenchmarkPluginCacheGet(b *testing.B) {
	cache := sdk.NewCache(5 * time.Minute)
	plugin := &sdk.LoadedPlugin{
		Name:     "test-plugin",
		Path:     "/path/to/plugin",
		LastUsed: time.Now(),
	}
	cache.Put("test-plugin", plugin)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = cache.Get("test-plugin")
	}
}

// BenchmarkPluginCachePutGet benchmarks combined put and get
func BenchmarkPluginCachePutGet(b *testing.B) {
	cache := sdk.NewCache(5 * time.Minute)
	plugin := &sdk.LoadedPlugin{
		Name:     "test-plugin",
		Path:     "/path/to/plugin",
		LastUsed: time.Now(),
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache.Put("test-plugin", plugin)
		_ = cache.Get("test-plugin")
	}
}

// BenchmarkPluginCleanup benchmarks cleanup operations
func BenchmarkPluginCleanup(b *testing.B) {
	tmpDir := b.TempDir()
	pluginDir := filepath.Join(tmpDir, ".glide", "plugins")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		b.Fatal(err)
	}

	manager := sdk.NewManager(&sdk.ManagerConfig{
		PluginDirs:   []string{pluginDir},
		CacheTimeout: 5 * time.Minute,
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager.Cleanup()
	}
}
