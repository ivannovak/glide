package benchmarks_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glide-cli/glide/v3/internal/config"
	"github.com/glide-cli/glide/v3/pkg/branding"
)

// BenchmarkConfigLoad benchmarks configuration loading
func BenchmarkConfigLoad(b *testing.B) {
	loader := config.NewLoader()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = loader.Load()
	}
}

// BenchmarkConfigDiscoverySingleLevel benchmarks config discovery in flat structure
func BenchmarkConfigDiscoverySingleLevel(b *testing.B) {
	// Setup: Create directory with single config
	tmpDir := b.TempDir()
	projectDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		b.Fatal(err)
	}

	configPath := filepath.Join(projectDir, branding.ConfigFileName)
	if err := os.WriteFile(configPath, []byte("version: 1\n"), 0644); err != nil {
		b.Fatal(err)
	}

	gitDir := filepath.Join(projectDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = config.DiscoverConfigs(projectDir)
	}
}

// BenchmarkConfigDiscoveryNested benchmarks config discovery in nested structure
func BenchmarkConfigDiscoveryNested(b *testing.B) {
	// Setup: Create nested directory structure
	tmpDir := b.TempDir()
	deepDir := filepath.Join(tmpDir, "a", "b", "c", "d", "e")
	if err := os.MkdirAll(deepDir, 0755); err != nil {
		b.Fatal(err)
	}

	// Add configs at multiple levels
	for _, dir := range []string{tmpDir, filepath.Join(tmpDir, "a", "b"), deepDir} {
		configPath := filepath.Join(dir, branding.ConfigFileName)
		if err := os.WriteFile(configPath, []byte("version: 1\n"), 0644); err != nil {
			b.Fatal(err)
		}
	}

	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = config.DiscoverConfigs(deepDir)
	}
}

// BenchmarkConfigMergingEmpty benchmarks merging empty configs
func BenchmarkConfigMergingEmpty(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = config.LoadAndMergeConfigs([]string{})
	}
}

// BenchmarkConfigMergingSingle benchmarks merging single config
func BenchmarkConfigMergingSingle(b *testing.B) {
	tmpDir := b.TempDir()
	configPath := filepath.Join(tmpDir, branding.ConfigFileName)

	configContent := `version: 1
commands:
  test:
    cmd: echo "test"
    description: "Run tests"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = config.LoadAndMergeConfigs([]string{configPath})
	}
}

// BenchmarkConfigMergingMultiple benchmarks merging multiple configs
func BenchmarkConfigMergingMultiple(b *testing.B) {
	tmpDir := b.TempDir()

	// Create multiple config files
	var configs []string
	for j := 0; j < 5; j++ {
		configPath := filepath.Join(tmpDir, "config"+string(rune('a'+j))+".yml")
		configContent := `version: 1
commands:
  test:
    cmd: echo "test"
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			b.Fatal(err)
		}
		configs = append(configs, configPath)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = config.LoadAndMergeConfigs(configs)
	}
}

// BenchmarkConfigMergingLarge benchmarks merging large configs
func BenchmarkConfigMergingLarge(b *testing.B) {
	tmpDir := b.TempDir()
	configPath := filepath.Join(tmpDir, branding.ConfigFileName)

	// Create config with many commands
	configContent := "version: 1\ncommands:\n"
	for j := 0; j < 100; j++ {
		configContent += "  cmd" + string(rune(j)) + ":\n"
		configContent += "    cmd: echo test\n"
		configContent += "    description: Test command\n"
	}

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = config.LoadAndMergeConfigs([]string{configPath})
	}
}

// BenchmarkConfigValidation benchmarks config validation
func BenchmarkConfigValidation(b *testing.B) {
	tmpDir := b.TempDir()
	configPath := filepath.Join(tmpDir, branding.ConfigFileName)

	configContent := `version: 1
commands:
  test:
    cmd: echo "test"
    description: "Run tests"
  build:
    cmd: make build
    description: "Build project"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = config.LoadAndMergeConfigs([]string{configPath})
	}
}
