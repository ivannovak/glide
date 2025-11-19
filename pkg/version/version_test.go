package version

import (
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetBuildInfo(t *testing.T) {
	// Save original values
	originalVersion := Version
	originalBuildDate := BuildDate
	originalGitCommit := GitCommit

	// Restore original values after test
	defer func() {
		Version = originalVersion
		BuildDate = originalBuildDate
		GitCommit = originalGitCommit
	}()

	// Test setting build info
	SetBuildInfo("v1.2.3", "2025-01-01", "abc123def")

	assert.Equal(t, "v1.2.3", Version)
	assert.Equal(t, "2025-01-01", BuildDate)
	assert.Equal(t, "abc123def", GitCommit)
}

func TestSet(t *testing.T) {
	// Save original value
	originalVersion := Version
	defer func() {
		Version = originalVersion
	}()

	// Test legacy Set function
	Set("v2.0.0")
	assert.Equal(t, "v2.0.0", Version)
}

func TestGet(t *testing.T) {
	// Save original value
	originalVersion := Version
	defer func() {
		Version = originalVersion
	}()

	// Test getting version
	Version = "test-version"
	assert.Equal(t, "test-version", Get())
}

func TestGetBuildInfo(t *testing.T) {
	// Save original values
	originalVersion := Version
	originalBuildDate := BuildDate
	originalGitCommit := GitCommit

	// Restore original values after test
	defer func() {
		Version = originalVersion
		BuildDate = originalBuildDate
		GitCommit = originalGitCommit
	}()

	// Set test values
	Version = "v1.0.0"
	BuildDate = "2025-01-01"
	GitCommit = "testcommit"

	info := GetBuildInfo()

	assert.Equal(t, "v1.0.0", info.Version)
	assert.Equal(t, "2025-01-01", info.BuildDate)
	assert.Equal(t, "testcommit", info.GitCommit)
	assert.Equal(t, runtime.Version(), info.GoVersion)
	assert.Equal(t, runtime.GOOS, info.OS)
	assert.Equal(t, runtime.GOARCH, info.Architecture)
	assert.Equal(t, runtime.Compiler, info.Compiler)
}

func TestGetVersionString(t *testing.T) {
	// Save original value
	originalVersion := Version
	defer func() {
		Version = originalVersion
	}()

	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "development version",
			version:  "dev",
			expected: "glide version dev (development build)",
		},
		{
			name:     "release version",
			version:  "v1.2.3",
			expected: "glide version v1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version = tt.version
			assert.Equal(t, tt.expected, GetVersionString())
		})
	}
}

func TestGetSystemInfo(t *testing.T) {
	info := GetSystemInfo()

	// Should contain OS, Architecture, and Go version
	assert.Contains(t, info, runtime.GOOS)
	assert.Contains(t, info, runtime.GOARCH)
	assert.Contains(t, info, "Go:")
	assert.Contains(t, info, "OS:")
	assert.Contains(t, info, "Architecture:")

	// Should be properly formatted
	parts := strings.Split(info, ", ")
	assert.Len(t, parts, 3) // OS, Architecture, Go
}

func TestBuildInfoStruct(t *testing.T) {
	info := BuildInfo{
		Version:      "v1.0.0",
		BuildDate:    "2025-01-01",
		GitCommit:    "abc123",
		GoVersion:    "go1.21",
		OS:           "linux",
		Architecture: "amd64",
		Compiler:     "gc",
	}

	assert.Equal(t, "v1.0.0", info.Version)
	assert.Equal(t, "2025-01-01", info.BuildDate)
	assert.Equal(t, "abc123", info.GitCommit)
	assert.Equal(t, "go1.21", info.GoVersion)
	assert.Equal(t, "linux", info.OS)
	assert.Equal(t, "amd64", info.Architecture)
	assert.Equal(t, "gc", info.Compiler)
}
