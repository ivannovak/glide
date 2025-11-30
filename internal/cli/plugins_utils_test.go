package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsGitHubURL(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected bool
	}{
		{
			name:     "github.com short format",
			source:   "github.com/user/repo",
			expected: true,
		},
		{
			name:     "https github.com format",
			source:   "https://github.com/user/repo",
			expected: true,
		},
		{
			name:     "https github.com with git extension",
			source:   "https://github.com/user/repo.git",
			expected: true,
		},
		{
			name:     "not a github URL - gitlab",
			source:   "gitlab.com/user/repo",
			expected: false,
		},
		{
			name:     "not a github URL - bitbucket",
			source:   "https://bitbucket.org/user/repo",
			expected: false,
		},
		{
			name:     "empty string",
			source:   "",
			expected: false,
		},
		{
			name:     "too short",
			source:   "github.com",
			expected: false,
		},
		{
			name:     "partial match",
			source:   "mygithub.com/user/repo",
			expected: false,
		},
		{
			name:     "http github (not https)",
			source:   "http://github.com/user/repo",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isGitHubURL(tt.source)
			assert.Equal(t, tt.expected, result, "isGitHubURL(%q) = %v, want %v", tt.source, result, tt.expected)
		})
	}
}

func TestExtractGitHubRepo(t *testing.T) {
	tests := []struct {
		name     string
		homepage string
		expected string
	}{
		{
			name:     "https github URL",
			homepage: "https://github.com/user/repo",
			expected: "user/repo",
		},
		{
			name:     "http github URL",
			homepage: "http://github.com/user/repo",
			expected: "user/repo",
		},
		{
			name:     "github URL without protocol",
			homepage: "github.com/user/repo",
			expected: "user/repo",
		},
		{
			name:     "github URL with path",
			homepage: "https://github.com/user/repo/issues/123",
			expected: "user/repo",
		},
		{
			name:     "github URL with .git extension",
			homepage: "https://github.com/user/repo.git",
			expected: "user/repo.git",
		},
		{
			name:     "non-github URL",
			homepage: "https://gitlab.com/user/repo",
			expected: "",
		},
		{
			name:     "empty string",
			homepage: "",
			expected: "",
		},
		{
			name:     "github.com only",
			homepage: "github.com/",
			expected: "",
		},
		{
			name:     "github.com with only user",
			homepage: "github.com/user",
			expected: "",
		},
		{
			name:     "complex path with multiple segments",
			homepage: "https://github.com/org/repo/tree/main/pkg/test",
			expected: "org/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractGitHubRepo(tt.homepage)
			assert.Equal(t, tt.expected, result, "extractGitHubRepo(%q) = %q, want %q", tt.homepage, result, tt.expected)
		})
	}
}

func TestIsValidGitHubDownloadURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "valid github.com download URL",
			url:      "https://github.com/user/repo/releases/download/v1.0.0/binary",
			expected: true,
		},
		{
			name:     "valid api.github.com URL",
			url:      "https://api.github.com/repos/user/repo/releases/assets/123",
			expected: true,
		},
		{
			name:     "github.com base URL (exactly 19 chars - invalid)",
			url:      "https://github.com/",
			expected: false,
		},
		{
			name:     "api.github.com base URL",
			url:      "https://api.github.com/",
			expected: true,
		},
		{
			name:     "invalid - http instead of https",
			url:      "http://github.com/user/repo",
			expected: false,
		},
		{
			name:     "invalid - different domain",
			url:      "https://gitlab.com/user/repo",
			expected: false,
		},
		{
			name:     "invalid - empty string",
			url:      "",
			expected: false,
		},
		{
			name:     "invalid - too short",
			url:      "https://github",
			expected: false,
		},
		{
			name:     "invalid - subdomain of github",
			url:      "https://mygithub.com/user/repo",
			expected: false,
		},
		{
			name:     "invalid - github in path but not domain",
			url:      "https://example.com/github.com/user/repo",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidGitHubDownloadURL(tt.url)
			assert.Equal(t, tt.expected, result, "isValidGitHubDownloadURL(%q) = %v, want %v", tt.url, result, tt.expected)
		})
	}
}
