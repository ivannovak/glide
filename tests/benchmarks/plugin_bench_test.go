package benchmarks_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ivannovak/glide/v3/pkg/plugin/sdk"
)

// BenchmarkPluginDiscoveryLazy benchmarks lazy plugin discovery performance
// This measures only the filesystem scanning without plugin loading
func BenchmarkPluginDiscoveryLazy(b *testing.B) {
	// Setup: Create plugin directory with mock plugins
	tmpDir := b.TempDir()
	pluginDir := filepath.Join(tmpDir, ".glide", "plugins")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		b.Fatal(err)
	}

	// Create 10 mock plugin files
	for i := 0; i < 10; i++ {
		pluginPath := filepath.Join(pluginDir, fmt.Sprintf("glide-plugin-test%02d", i))
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
		// Use lazy discovery (filesystem scan only, no plugin loading)
		_ = manager.DiscoverPluginsLazy()
	}
}

// BenchmarkPluginDiscovery benchmarks plugin discovery with loading
// Note: This will be slower due to plugin load attempts
func BenchmarkPluginDiscovery(b *testing.B) {
	// Setup: Create plugin directory with mock plugins
	tmpDir := b.TempDir()
	pluginDir := filepath.Join(tmpDir, ".glide", "plugins")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		b.Fatal(err)
	}

	// Create 10 mock plugin files
	for i := 0; i < 10; i++ {
		pluginPath := filepath.Join(pluginDir, fmt.Sprintf("glide-plugin-test%02d", i))
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

// BenchmarkPluginDiscoveryEmpty benchmarks lazy discovery in empty directory
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
		_ = manager.DiscoverPluginsLazy()
	}
}

// BenchmarkPluginDiscoveryLargeLazy benchmarks lazy discovery with many plugins
// This is the more realistic benchmark - discovery without loading
func BenchmarkPluginDiscoveryLargeLazy(b *testing.B) {
	tmpDir := b.TempDir()
	pluginDir := filepath.Join(tmpDir, ".glide", "plugins")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		b.Fatal(err)
	}

	// Create 100 mock plugin files
	for i := 0; i < 100; i++ {
		pluginPath := filepath.Join(pluginDir, fmt.Sprintf("glide-plugin-test%03d", i))
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
		_ = manager.DiscoverPluginsLazy()
	}
}

// BenchmarkPluginDiscoveryMultipleDirs benchmarks lazy discovery across multiple directories
func BenchmarkPluginDiscoveryMultipleDirs(b *testing.B) {
	tmpDir := b.TempDir()

	// Create 5 plugin directories with 20 plugins each
	var pluginDirs []string
	for d := 0; d < 5; d++ {
		pluginDir := filepath.Join(tmpDir, fmt.Sprintf("dir%d", d), ".glide", "plugins")
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			b.Fatal(err)
		}
		pluginDirs = append(pluginDirs, pluginDir)

		for i := 0; i < 20; i++ {
			pluginPath := filepath.Join(pluginDir, fmt.Sprintf("glide-plugin-d%d-test%02d", d, i))
			if err := os.WriteFile(pluginPath, []byte("#!/bin/bash\necho test"), 0755); err != nil {
				b.Fatal(err)
			}
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager := sdk.NewManager(&sdk.ManagerConfig{
			PluginDirs:   pluginDirs,
			CacheTimeout: 5 * time.Minute,
		})
		_ = manager.DiscoverPluginsLazy()
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
