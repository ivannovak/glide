package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
)

var (
	// GitHub API endpoint for latest release
	githubAPIURL = "https://api.github.com/repos/ivannovak/glide/releases/latest"
)

const (
	// Timeout for API requests
	requestTimeout = 10 * time.Second
)

// Release represents a GitHub release
type Release struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
	Assets      []Asset   `json:"assets"`
}

// Asset represents a release asset
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// UpdateInfo contains information about available updates
type UpdateInfo struct {
	Available      bool
	CurrentVersion string
	LatestVersion  string
	ReleaseURL     string
	ReleaseNotes   string
	PublishedAt    time.Time
	DownloadURL    string
	AssetSize      int64
}

// Checker handles version update checking
type Checker struct {
	currentVersion string
	httpClient     *http.Client
}

// NewChecker creates a new update checker
func NewChecker(currentVersion string) *Checker {
	return &Checker{
		currentVersion: currentVersion,
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

// CheckForUpdate checks if a newer version is available
func (c *Checker) CheckForUpdate(ctx context.Context) (*UpdateInfo, error) {
	// Skip update check for development builds
	if c.currentVersion == "dev" || strings.Contains(c.currentVersion, "dev") {
		return &UpdateInfo{
			Available:      false,
			CurrentVersion: c.currentVersion,
			LatestVersion:  c.currentVersion,
		}, nil
	}

	// Fetch latest release from GitHub
	release, err := c.fetchLatestRelease(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}

	// Parse versions
	current, err := semver.NewVersion(c.currentVersion)
	if err != nil {
		// If current version can't be parsed, assume it's older
		return &UpdateInfo{
			Available:      true,
			CurrentVersion: c.currentVersion,
			LatestVersion:  release.TagName,
			ReleaseURL:     release.HTMLURL,
			ReleaseNotes:   release.Body,
			PublishedAt:    release.PublishedAt,
			DownloadURL:    c.getDownloadURL(release),
		}, nil
	}

	latest, err := semver.NewVersion(release.TagName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse latest version %s: %w", release.TagName, err)
	}

	// Compare versions
	updateAvailable := latest.GreaterThan(current)

	// Find appropriate download URL
	downloadURL := c.getDownloadURL(release)
	var assetSize int64
	for _, asset := range release.Assets {
		if asset.BrowserDownloadURL == downloadURL {
			assetSize = asset.Size
			break
		}
	}

	return &UpdateInfo{
		Available:      updateAvailable,
		CurrentVersion: c.currentVersion,
		LatestVersion:  release.TagName,
		ReleaseURL:     release.HTMLURL,
		ReleaseNotes:   release.Body,
		PublishedAt:    release.PublishedAt,
		DownloadURL:    downloadURL,
		AssetSize:      assetSize,
	}, nil
}

// fetchLatestRelease fetches the latest release information from GitHub
func (c *Checker) fetchLatestRelease(ctx context.Context) (*Release, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubAPIURL, nil)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "glide-cli-updater")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned %d: %s", resp.StatusCode, string(body))
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &release, nil
}

// getDownloadURL returns the appropriate download URL for the current platform
func (c *Checker) getDownloadURL(release *Release) string {
	// Determine platform binary name
	platform := fmt.Sprintf("glide-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		platform += ".exe"
	}

	// Find matching asset
	for _, asset := range release.Assets {
		if asset.Name == platform {
			return asset.BrowserDownloadURL
		}
	}

	// Fallback to release page if no direct download found
	return release.HTMLURL
}

// IsUpdateAvailable is a convenience method for quick checking
func (c *Checker) IsUpdateAvailable(ctx context.Context) bool {
	info, err := c.CheckForUpdate(ctx)
	if err != nil {
		return false
	}
	return info.Available
}

// FormatUpdateMessage formats a user-friendly update notification
func FormatUpdateMessage(info *UpdateInfo) string {
	if !info.Available {
		return fmt.Sprintf("You are running the latest version (%s)", info.CurrentVersion)
	}

	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("A new version of Glide is available: %s → %s\n",
		info.CurrentVersion, info.LatestVersion))
	msg.WriteString(fmt.Sprintf("Released: %s\n", info.PublishedAt.Format("2006-01-02")))

	if info.DownloadURL != "" && !strings.Contains(info.DownloadURL, "github.com/ivannovak/glide/releases") {
		msg.WriteString(fmt.Sprintf("\nDownload: %s\n", info.DownloadURL))
	} else {
		msg.WriteString(fmt.Sprintf("\nView release: %s\n", info.ReleaseURL))
	}

	msg.WriteString("\nUpdate with: curl -fsSL https://raw.githubusercontent.com/ivannovak/glide/main/install.sh | bash")

	return msg.String()
}
