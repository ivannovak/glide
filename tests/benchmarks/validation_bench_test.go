package benchmarks_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glide-cli/glide/v3/pkg/validation"
)

// BenchmarkValidatePathSimple benchmarks validating a simple relative path
func BenchmarkValidatePathSimple(b *testing.B) {
	tmpDir := b.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		b.Fatal(err)
	}

	opts := validation.PathValidationOptions{
		BaseDir:       tmpDir,
		AllowAbsolute: false,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = validation.ValidatePath("test.txt", opts)
	}
}

// BenchmarkValidatePathAbsolute benchmarks validating absolute paths
func BenchmarkValidatePathAbsolute(b *testing.B) {
	tmpDir := b.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		b.Fatal(err)
	}

	opts := validation.PathValidationOptions{
		BaseDir:       tmpDir,
		AllowAbsolute: true,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = validation.ValidatePath(testFile, opts)
	}
}

// BenchmarkValidatePathNested benchmarks validating nested paths
func BenchmarkValidatePathNested(b *testing.B) {
	tmpDir := b.TempDir()

	// Create nested directory structure
	nestedDir := filepath.Join(tmpDir, "a", "b", "c", "d")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		b.Fatal(err)
	}

	testFile := filepath.Join(nestedDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		b.Fatal(err)
	}

	opts := validation.PathValidationOptions{
		BaseDir:       tmpDir,
		AllowAbsolute: false,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = validation.ValidatePath("a/b/c/d/test.txt", opts)
	}
}

// BenchmarkValidatePathWithSymlinks benchmarks validating paths with symlink resolution
func BenchmarkValidatePathWithSymlinks(b *testing.B) {
	tmpDir := b.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "real.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		b.Fatal(err)
	}

	// Create a symlink
	linkPath := filepath.Join(tmpDir, "link.txt")
	if err := os.Symlink(testFile, linkPath); err != nil {
		b.Skip("Symlink creation not supported")
	}

	opts := validation.PathValidationOptions{
		BaseDir:        tmpDir,
		AllowAbsolute:  false,
		FollowSymlinks: true,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = validation.ValidatePath("link.txt", opts)
	}
}

// BenchmarkValidatePathWithExistsCheck benchmarks validating paths with existence check
func BenchmarkValidatePathWithExistsCheck(b *testing.B) {
	tmpDir := b.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		b.Fatal(err)
	}

	opts := validation.PathValidationOptions{
		BaseDir:       tmpDir,
		AllowAbsolute: false,
		RequireExists: true,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = validation.ValidatePath("test.txt", opts)
	}
}

// BenchmarkValidatePathTraversalAttempt benchmarks detecting path traversal
func BenchmarkValidatePathTraversalAttempt(b *testing.B) {
	tmpDir := b.TempDir()

	opts := validation.PathValidationOptions{
		BaseDir:       tmpDir,
		AllowAbsolute: false,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// This should fail with path traversal error
		_, _ = validation.ValidatePath("../../../etc/passwd", opts)
	}
}

// BenchmarkValidatePathMalicious benchmarks detecting various malicious paths
func BenchmarkValidatePathMalicious(b *testing.B) {
	tmpDir := b.TempDir()

	opts := validation.PathValidationOptions{
		BaseDir:       tmpDir,
		AllowAbsolute: false,
	}

	maliciousPaths := []string{
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32",
		"test/../../../etc/passwd",
		"/etc/passwd",
		"test/./../../etc/passwd",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, path := range maliciousPaths {
			_, _ = validation.ValidatePath(path, opts)
		}
	}
}

// BenchmarkValidatePathAllOptions benchmarks validation with all options enabled
func BenchmarkValidatePathAllOptions(b *testing.B) {
	tmpDir := b.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		b.Fatal(err)
	}

	opts := validation.PathValidationOptions{
		BaseDir:        tmpDir,
		AllowAbsolute:  true,
		FollowSymlinks: true,
		RequireExists:  true,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = validation.ValidatePath(testFile, opts)
	}
}

// BenchmarkValidatePathAllocation measures allocations for path validation
func BenchmarkValidatePathAllocation(b *testing.B) {
	tmpDir := b.TempDir()

	opts := validation.PathValidationOptions{
		BaseDir:       tmpDir,
		AllowAbsolute: false,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = validation.ValidatePath("test/file.txt", opts)
	}
}

// BenchmarkValidatePathConcurrent benchmarks concurrent path validation
func BenchmarkValidatePathConcurrent(b *testing.B) {
	tmpDir := b.TempDir()

	// Create test files
	for j := 0; j < 10; j++ {
		testFile := filepath.Join(tmpDir, "test"+string(rune('a'+j))+".txt")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			b.Fatal(err)
		}
	}

	opts := validation.PathValidationOptions{
		BaseDir:       tmpDir,
		AllowAbsolute: false,
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			path := "test" + string(rune('a'+(i%10))) + ".txt"
			_, _ = validation.ValidatePath(path, opts)
			i++
		}
	})
}
