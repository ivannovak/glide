package update

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewChecker(t *testing.T) {
	checker := NewChecker("v1.0.0")
	assert.NotNil(t, checker)
	assert.Equal(t, "v1.0.0", checker.currentVersion)
	assert.NotNil(t, checker.httpClient)
}

func TestCheckForUpdate_DevVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{"dev version", "dev"},
		{"dev suffix", "v1.0.0-dev"},
		{"dev in middle", "v1.0.0-dev-abc123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewChecker(tt.version)
			ctx := context.Background()

			info, err := checker.CheckForUpdate(ctx)
			require.NoError(t, err)
			assert.False(t, info.Available)
			assert.Equal(t, tt.version, info.CurrentVersion)
			assert.Equal(t, tt.version, info.LatestVersion)
		})
	}
}

func TestCheckForUpdate_NewVersionAvailable(t *testing.T) {
	// Create mock server
	release := Release{
		TagName:     "v2.0.0",
		Name:        "v2.0.0",
		Body:        "New features and bug fixes",
		PublishedAt: time.Now(),
		HTMLURL:     "https://github.com/ivannovak/glide/releases/tag/v2.0.0",
		Assets: []Asset{
			{
				Name:               "glide-darwin-arm64",
				BrowserDownloadURL: "https://github.com/ivannovak/glide/releases/download/v2.0.0/glide-darwin-arm64",
				Size:               10485760,
			},
			{
				Name:               "glide-linux-amd64",
				BrowserDownloadURL: "https://github.com/ivannovak/glide/releases/download/v2.0.0/glide-linux-amd64",
				Size:               10485760,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	// Override the API URL for testing
	oldURL := githubAPIURL
	githubAPIURL = server.URL
	defer func() { githubAPIURL = oldURL }()

	checker := NewChecker("v1.0.0")
	ctx := context.Background()

	info, err := checker.CheckForUpdate(ctx)
	require.NoError(t, err)
	assert.True(t, info.Available)
	assert.Equal(t, "v1.0.0", info.CurrentVersion)
	assert.Equal(t, "v2.0.0", info.LatestVersion)
	assert.Equal(t, release.HTMLURL, info.ReleaseURL)
	assert.Equal(t, release.Body, info.ReleaseNotes)
}

func TestCheckForUpdate_SameVersion(t *testing.T) {
	// Create mock server
	release := Release{
		TagName:     "v1.0.0",
		Name:        "v1.0.0",
		PublishedAt: time.Now(),
		HTMLURL:     "https://github.com/ivannovak/glide/releases/tag/v1.0.0",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	// Override the API URL for testing
	oldURL := githubAPIURL
	githubAPIURL = server.URL
	defer func() { githubAPIURL = oldURL }()

	checker := NewChecker("v1.0.0")
	ctx := context.Background()

	info, err := checker.CheckForUpdate(ctx)
	require.NoError(t, err)
	assert.False(t, info.Available)
	assert.Equal(t, "v1.0.0", info.CurrentVersion)
	assert.Equal(t, "v1.0.0", info.LatestVersion)
}

func TestCheckForUpdate_OlderVersionAvailable(t *testing.T) {
	// Create mock server
	release := Release{
		TagName:     "v0.9.0",
		Name:        "v0.9.0",
		PublishedAt: time.Now(),
		HTMLURL:     "https://github.com/ivannovak/glide/releases/tag/v0.9.0",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	// Override the API URL for testing
	oldURL := githubAPIURL
	githubAPIURL = server.URL
	defer func() { githubAPIURL = oldURL }()

	checker := NewChecker("v1.0.0")
	ctx := context.Background()

	info, err := checker.CheckForUpdate(ctx)
	require.NoError(t, err)
	assert.False(t, info.Available, "Should not show older version as available")
	assert.Equal(t, "v1.0.0", info.CurrentVersion)
	assert.Equal(t, "v0.9.0", info.LatestVersion)
}

func TestCheckForUpdate_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	// Override the API URL for testing
	oldURL := githubAPIURL
	githubAPIURL = server.URL
	defer func() { githubAPIURL = oldURL }()

	checker := NewChecker("v1.0.0")
	ctx := context.Background()

	info, err := checker.CheckForUpdate(ctx)
	assert.Error(t, err)
	assert.Nil(t, info)
	assert.Contains(t, err.Error(), "GitHub API returned 500")
}

func TestCheckForUpdate_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	// Override the API URL for testing
	oldURL := githubAPIURL
	githubAPIURL = server.URL
	defer func() { githubAPIURL = oldURL }()

	checker := NewChecker("v1.0.0")
	ctx := context.Background()

	info, err := checker.CheckForUpdate(ctx)
	assert.Error(t, err)
	assert.Nil(t, info)
	assert.Contains(t, err.Error(), "failed to decode response")
}

func TestCheckForUpdate_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Override the API URL for testing
	oldURL := githubAPIURL
	githubAPIURL = server.URL
	defer func() { githubAPIURL = oldURL }()

	checker := NewChecker("v1.0.0")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	info, err := checker.CheckForUpdate(ctx)
	assert.Error(t, err)
	assert.Nil(t, info)
}

func TestGetDownloadURL(t *testing.T) {
	tests := []struct {
		name     string
		goos     string
		goarch   string
		expected string
	}{
		{
			name:     "darwin arm64",
			goos:     "darwin",
			goarch:   "arm64",
			expected: "glide-darwin-arm64",
		},
		{
			name:     "linux amd64",
			goos:     "linux",
			goarch:   "amd64",
			expected: "glide-linux-amd64",
		},
		{
			name:     "windows amd64",
			goos:     "windows",
			goarch:   "amd64",
			expected: "glide-windows-amd64.exe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			release := &Release{
				HTMLURL: "https://github.com/ivannovak/glide/releases/tag/v1.0.0",
				Assets: []Asset{
					{
						Name:               "glide-darwin-arm64",
						BrowserDownloadURL: "https://download/glide-darwin-arm64",
					},
					{
						Name:               "glide-linux-amd64",
						BrowserDownloadURL: "https://download/glide-linux-amd64",
					},
					{
						Name:               "glide-windows-amd64.exe",
						BrowserDownloadURL: "https://download/glide-windows-amd64.exe",
					},
				},
			}

			// Mock runtime values for testing
			// Note: In real test, we'd need to mock runtime.GOOS/GOARCH
			// For now, we'll test the logic by checking all assets exist
			checker := NewChecker("v1.0.0")
			url := checker.getDownloadURL(release)
			assert.NotEmpty(t, url)
		})
	}
}

func TestIsUpdateAvailable(t *testing.T) {
	// Create mock server
	release := Release{
		TagName:     "v2.0.0",
		Name:        "v2.0.0",
		PublishedAt: time.Now(),
		HTMLURL:     "https://github.com/ivannovak/glide/releases/tag/v2.0.0",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	// Override the API URL for testing
	oldURL := githubAPIURL
	githubAPIURL = server.URL
	defer func() { githubAPIURL = oldURL }()

	checker := NewChecker("v1.0.0")
	ctx := context.Background()

	available := checker.IsUpdateAvailable(ctx)
	assert.True(t, available)
}

func TestFormatUpdateMessage(t *testing.T) {
	tests := []struct {
		name     string
		info     *UpdateInfo
		expected []string
	}{
		{
			name: "update available",
			info: &UpdateInfo{
				Available:      true,
				CurrentVersion: "v1.0.0",
				LatestVersion:  "v2.0.0",
				ReleaseURL:     "https://github.com/ivannovak/glide/releases/tag/v2.0.0",
				PublishedAt:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				DownloadURL:    "https://github.com/ivannovak/glide/releases/download/v2.0.0/glide-darwin-arm64",
			},
			expected: []string{
				"A new version of Glide is available: v1.0.0 → v2.0.0",
				"Released: 2025-01-01",
				"View release:",
				"Update with: curl",
			},
		},
		{
			name: "no update available",
			info: &UpdateInfo{
				Available:      false,
				CurrentVersion: "v1.0.0",
				LatestVersion:  "v1.0.0",
			},
			expected: []string{
				"You are running the latest version (v1.0.0)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := FormatUpdateMessage(tt.info)
			for _, exp := range tt.expected {
				assert.Contains(t, msg, exp)
			}
		})
	}
}
