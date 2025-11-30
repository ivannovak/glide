package benchmarks_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ivannovak/glide/v3/internal/context"
)

// BenchmarkContextDetection benchmarks basic context detection
// Note: This tests with lazy Docker checking (default mode)
func BenchmarkContextDetection(b *testing.B) {
	// Setup: Create git repository
	tmpDir := b.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		b.Fatal(err)
	}

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	if err := os.Chdir(tmpDir); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		detector, err := context.NewDetector()
		if err != nil {
			b.Fatal(err)
		}
		_, _ = detector.Detect()
	}
}

// BenchmarkContextDetectionFast benchmarks fast context detection
// This skips Docker daemon checks entirely for maximum startup speed
func BenchmarkContextDetectionFast(b *testing.B) {
	// Setup: Create git repository
	tmpDir := b.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		b.Fatal(err)
	}

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	if err := os.Chdir(tmpDir); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		detector, err := context.NewDetectorFast()
		if err != nil {
			b.Fatal(err)
		}
		_, _ = detector.Detect()
	}
}

// BenchmarkContextDetectionWithConfig benchmarks detection with config file
func BenchmarkContextDetectionWithConfig(b *testing.B) {
	tmpDir := b.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		b.Fatal(err)
	}

	configPath := filepath.Join(tmpDir, ".glide.yml")
	if err := os.WriteFile(configPath, []byte("version: 1\n"), 0644); err != nil {
		b.Fatal(err)
	}

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	if err := os.Chdir(tmpDir); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		detector, err := context.NewDetector()
		if err != nil {
			b.Fatal(err)
		}
		_, _ = detector.Detect()
	}
}

// BenchmarkContextDetectionNested benchmarks detection from deep directory
func BenchmarkContextDetectionNested(b *testing.B) {
	tmpDir := b.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		b.Fatal(err)
	}

	deepDir := filepath.Join(tmpDir, "a", "b", "c", "d", "e")
	if err := os.MkdirAll(deepDir, 0755); err != nil {
		b.Fatal(err)
	}

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	if err := os.Chdir(deepDir); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		detector, err := context.NewDetector()
		if err != nil {
			b.Fatal(err)
		}
		_, _ = detector.Detect()
	}
}

// BenchmarkContextDetectionMultiFramework benchmarks detection with multiple frameworks
func BenchmarkContextDetectionMultiFramework(b *testing.B) {
	tmpDir := b.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		b.Fatal(err)
	}

	// PHP project
	composerPath := filepath.Join(tmpDir, "composer.json")
	if err := os.WriteFile(composerPath, []byte(`{"require": {"php": "^8.0"}}`), 0644); err != nil {
		b.Fatal(err)
	}

	// Node project
	packagePath := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(packagePath, []byte(`{"name": "test", "version": "1.0.0"}`), 0644); err != nil {
		b.Fatal(err)
	}

	// Go project
	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module test\n"), 0644); err != nil {
		b.Fatal(err)
	}

	// Docker
	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte("FROM node:18\n"), 0644); err != nil {
		b.Fatal(err)
	}

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	if err := os.Chdir(tmpDir); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		detector, err := context.NewDetector()
		if err != nil {
			b.Fatal(err)
		}
		_, _ = detector.Detect()
	}
}

// BenchmarkContextDetectionSingleRepo benchmarks single repo mode detection
func BenchmarkContextDetectionSingleRepo(b *testing.B) {
	tmpDir := b.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		b.Fatal(err)
	}

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	if err := os.Chdir(tmpDir); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		detector, err := context.NewDetector()
		if err != nil {
			b.Fatal(err)
		}
		_, _ = detector.Detect()
	}
}

// BenchmarkContextDetectionMultiWorktree benchmarks multi-worktree mode detection
func BenchmarkContextDetectionMultiWorktree(b *testing.B) {
	tmpDir := b.TempDir()

	// Create vcs directory with .git
	vcsDir := filepath.Join(tmpDir, "vcs")
	if err := os.MkdirAll(vcsDir, 0755); err != nil {
		b.Fatal(err)
	}

	gitDir := filepath.Join(vcsDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		b.Fatal(err)
	}

	// Create worktrees directory
	worktreesDir := filepath.Join(tmpDir, "worktrees")
	if err := os.MkdirAll(worktreesDir, 0755); err != nil {
		b.Fatal(err)
	}

	// Create a worktree
	worktreeDir := filepath.Join(worktreesDir, "issue-123")
	if err := os.MkdirAll(worktreeDir, 0755); err != nil {
		b.Fatal(err)
	}

	// Create .git file in worktree
	gitFile := filepath.Join(worktreeDir, ".git")
	gitContent := "gitdir: " + filepath.Join(gitDir, "worktrees", "issue-123")
	if err := os.WriteFile(gitFile, []byte(gitContent), 0644); err != nil {
		b.Fatal(err)
	}

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	if err := os.Chdir(worktreeDir); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		detector, err := context.NewDetector()
		if err != nil {
			b.Fatal(err)
		}
		_, _ = detector.Detect()
	}
}

// BenchmarkContextValidation benchmarks context validation
func BenchmarkContextValidation(b *testing.B) {
	tmpDir := b.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		b.Fatal(err)
	}

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	if err := os.Chdir(tmpDir); err != nil {
		b.Fatal(err)
	}

	// Pre-create context for benchmarking validation only
	detector, err := context.NewDetector()
	if err != nil {
		b.Fatal(err)
	}

	ctx, err := detector.Detect()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = ctx.IsValid()
	}
}
