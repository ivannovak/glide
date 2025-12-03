package branding

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultValues(t *testing.T) {
	// Test default values
	assert.Equal(t, "glide", CommandName)
	assert.Equal(t, ".glide.yml", ConfigFileName)
	assert.Equal(t, "Glide", ProjectName)
	assert.Equal(t, "context-aware development CLI", Description)
	assert.Contains(t, LongDescription, "modern")
	assert.Equal(t, "glide", CompletionDir)
	assert.Equal(t, "https://github.com/glide-cli/glide", RepositoryURL)
}

func TestGetConfigPath(t *testing.T) {
	// Save original values
	originalConfigFileName := ConfigFileName
	defer func() {
		ConfigFileName = originalConfigFileName
	}()

	// Test default config path
	homeDir, _ := os.UserHomeDir()
	expectedPath := filepath.Join(homeDir, ".glide.yml")
	assert.Equal(t, expectedPath, GetConfigPath())

	// Test with custom config file name
	ConfigFileName = ".mycli.yml"
	expectedPath = filepath.Join(homeDir, ".mycli.yml")
	assert.Equal(t, expectedPath, GetConfigPath())
}

func TestGetShortDescription(t *testing.T) {
	// Save original values
	originalProjectName := ProjectName
	originalDescription := Description
	defer func() {
		ProjectName = originalProjectName
		Description = originalDescription
	}()

	// Test default short description
	assert.Equal(t, "Glide context-aware development CLI", GetShortDescription())

	// Test with custom values
	ProjectName = "MyProject"
	Description = "awesome tool"
	assert.Equal(t, "MyProject awesome tool", GetShortDescription())
}

func TestGetFullDescription(t *testing.T) {
	// Save original values
	originalCommandName := CommandName
	originalProjectName := ProjectName
	defer func() {
		CommandName = originalCommandName
		ProjectName = originalProjectName
	}()

	// Test default full description
	desc := GetFullDescription()
	assert.Contains(t, desc, "Glid")
	assert.Contains(t, desc, "modern development CLI")
	assert.Contains(t, desc, "glides through complex workflows")

	// Test with custom values
	CommandName = "mycli"
	ProjectName = "MyProject"
	desc = GetFullDescription()
	assert.Contains(t, desc, "Mycli")
	assert.Contains(t, desc, "modern development CLI")
}

func TestGetCompletionPath(t *testing.T) {
	// Save original value
	originalCompletionDir := CompletionDir
	defer func() {
		CompletionDir = originalCompletionDir
	}()

	tests := []struct {
		name     string
		shell    string
		expected string
	}{
		{
			name:     "bash completion path",
			shell:    "bash",
			expected: "/usr/local/etc/bash_completion.d/glide",
		},
		{
			name:     "zsh completion path",
			shell:    "zsh",
			expected: "/usr/local/share/zsh/site-functions/glide",
		},
		{
			name:     "fish completion path",
			shell:    "fish",
			expected: filepath.Join(os.Getenv("HOME"), ".config", "fish", "completions", "glide"),
		},
		{
			name:     "unknown shell",
			shell:    "unknown",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetCompletionPath(tt.shell)
			assert.Equal(t, tt.expected, result)
		})
	}

	// Test with custom completion dir
	CompletionDir = "mycli"
	path := GetCompletionPath("bash")
	assert.Equal(t, "/usr/local/etc/bash_completion.d/mycli", path)
}

func TestCapitalize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"glid", "Glid"},
		{"mycli", "Mycli"},
		{"a", "A"},
		{"", ""},
		{"A", "!"}, // ASCII math: 'A' - 32 = '!'
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := capitalize(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBrandingCustomization(t *testing.T) {
	// Save original values
	originalCommandName := CommandName
	originalConfigFileName := ConfigFileName
	originalProjectName := ProjectName
	originalDescription := Description
	originalCompletionDir := CompletionDir
	originalRepositoryURL := RepositoryURL

	defer func() {
		CommandName = originalCommandName
		ConfigFileName = originalConfigFileName
		ProjectName = originalProjectName
		Description = originalDescription
		CompletionDir = originalCompletionDir
		RepositoryURL = originalRepositoryURL
	}()

	// Simulate build-time ldflags customization
	CommandName = "acme"
	ConfigFileName = ".acme.yml"
	ProjectName = "ACME Corp"
	Description = "deployment tool"
	CompletionDir = "acme"
	RepositoryURL = "https://github.com/acme/acme-cli"

	// Verify all values are customized
	assert.Equal(t, "acme", CommandName)
	assert.Equal(t, ".acme.yml", ConfigFileName)
	assert.Equal(t, "ACME Corp", ProjectName)
	assert.Equal(t, "deployment tool", Description)
	assert.Equal(t, "acme", CompletionDir)
	assert.Equal(t, "https://github.com/acme/acme-cli", RepositoryURL)

	// Verify functions use the custom values
	assert.Equal(t, "ACME Corp deployment tool", GetShortDescription())
	assert.Contains(t, GetFullDescription(), "Acme")
	assert.Contains(t, GetFullDescription(), "modern development CLI")

	homeDir, _ := os.UserHomeDir()
	assert.Equal(t, filepath.Join(homeDir, ".acme.yml"), GetConfigPath())

	assert.True(t, strings.HasSuffix(GetCompletionPath("bash"), "/acme"))
}

func TestLongDescription(t *testing.T) {
	// Test that long description contains expected content
	assert.Contains(t, LongDescription, "modern")
	assert.Contains(t, LongDescription, "context-aware")
	assert.Contains(t, LongDescription, "glides through complex workflows")
	assert.Contains(t, LongDescription, "multi-worktree")
}
