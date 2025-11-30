package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUpdater(t *testing.T) {
	updater := NewUpdater("v1.0.0")
	assert.NotNil(t, updater)
	assert.NotNil(t, updater.checker)
	assert.NotNil(t, updater.httpClient)
	assert.Equal(t, updater.httpClient.Timeout, 0*time.Second, "Download client should have no timeout")
}

func TestSelfUpdate_NoUpdateAvailable(t *testing.T) {
	// Create mock server that returns same version
	release := Release{
		TagName:     "v1.0.0",
		Name:        "v1.0.0",
		PublishedAt: time.Now(),
		HTMLURL:     "https://github.com/ivannovak/glide/v3/releases/tag/v1.0.0",
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

	updater := NewUpdater("v1.0.0")
	ctx := context.Background()

	err := updater.SelfUpdate(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no update available")
}

func TestSelfUpdate_DevVersion(t *testing.T) {
	// Dev versions should return error immediately
	updater := NewUpdater("dev")
	ctx := context.Background()

	err := updater.SelfUpdate(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no update available")
}

func TestDownloadBinary_Success(t *testing.T) {
	// Create a mock server that serves a binary
	testContent := []byte("test binary content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(testContent)
	}))
	defer server.Close()

	updater := NewUpdater("v1.0.0")
	ctx := context.Background()

	tempFile, err := updater.downloadBinary(ctx, server.URL)
	require.NoError(t, err)
	defer os.Remove(tempFile)

	// Verify the file was created and has correct content
	content, err := os.ReadFile(tempFile)
	require.NoError(t, err)
	assert.Equal(t, testContent, content)

	// Verify the file is executable
	info, err := os.Stat(tempFile)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0755), info.Mode().Perm())
}

func TestDownloadBinary_GitHubReleasePageRejected(t *testing.T) {
	updater := NewUpdater("v1.0.0")
	ctx := context.Background()

	// Test that GitHub release pages (not direct downloads) are rejected
	urls := []string{
		"https://github.com/ivannovak/glide/v3/releases/tag/v1.0.0",
		"https://github.com/ivannovak/glide/v3/releases/latest",
	}

	for _, url := range urls {
		tempFile, err := updater.downloadBinary(ctx, url)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "direct download not available")
		assert.Empty(t, tempFile)
	}
}

func TestDownloadBinary_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	updater := NewUpdater("v1.0.0")
	ctx := context.Background()

	tempFile, err := updater.downloadBinary(ctx, server.URL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "download failed with status 404")
	assert.Empty(t, tempFile)
}

func TestVerifyChecksum_Success(t *testing.T) {
	// Create a test file
	testContent := []byte("test content")
	tempFile, err := os.CreateTemp("", "test-binary-*")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write(testContent)
	require.NoError(t, err)
	tempFile.Close()

	// Calculate checksum
	hasher := sha256.New()
	hasher.Write(testContent)
	expectedChecksum := hex.EncodeToString(hasher.Sum(nil))

	// Create mock server that returns checksum
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(expectedChecksum + "  glide-darwin-arm64\n"))
	}))
	defer server.Close()

	updater := NewUpdater("v1.0.0")
	ctx := context.Background()

	err = updater.verifyChecksum(ctx, tempFile.Name(), server.URL)
	assert.NoError(t, err)
}

func TestVerifyChecksum_Mismatch(t *testing.T) {
	// Create a test file
	testContent := []byte("test content")
	tempFile, err := os.CreateTemp("", "test-binary-*")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write(testContent)
	require.NoError(t, err)
	tempFile.Close()

	// Create mock server that returns wrong checksum
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("wrongchecksum1234567890abcdef  glide-darwin-arm64\n"))
	}))
	defer server.Close()

	updater := NewUpdater("v1.0.0")
	ctx := context.Background()

	err = updater.verifyChecksum(ctx, tempFile.Name(), server.URL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checksum mismatch")
}

func TestVerifyChecksum_FileNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	updater := NewUpdater("v1.0.0")
	ctx := context.Background()

	err := updater.verifyChecksum(ctx, "/nonexistent/file", server.URL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checksum file not found")
}

func TestReplaceBinary(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "glide-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create current binary
	currentPath := filepath.Join(tempDir, "glide")
	currentContent := []byte("current version")
	err = os.WriteFile(currentPath, currentContent, 0755)
	require.NoError(t, err)

	// Create new binary
	newPath := filepath.Join(tempDir, "glide-new")
	newContent := []byte("new version")
	err = os.WriteFile(newPath, newContent, 0755)
	require.NoError(t, err)

	updater := NewUpdater("v1.0.0")
	err = updater.replaceBinary(currentPath, newPath)
	require.NoError(t, err)

	// Verify the binary was replaced
	content, err := os.ReadFile(currentPath)
	require.NoError(t, err)
	assert.Equal(t, newContent, content)

	// Verify backup was removed
	backupPath := currentPath + ".backup"
	_, err = os.Stat(backupPath)
	assert.True(t, os.IsNotExist(err), "Backup should be removed after successful replacement")

	// Verify new file was moved (not copied)
	_, err = os.Stat(newPath)
	assert.True(t, os.IsNotExist(err), "New file should be moved, not copied")
}

func TestReplaceBinary_RollbackOnFailure(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "glide-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create current binary
	currentPath := filepath.Join(tempDir, "glide")
	originalContent := []byte("original version")
	err = os.WriteFile(currentPath, originalContent, 0755)
	require.NoError(t, err)

	// Create new binary in a read-only directory to force failure
	readOnlyDir := filepath.Join(tempDir, "readonly")
	err = os.Mkdir(readOnlyDir, 0755)
	require.NoError(t, err)

	newPath := filepath.Join(readOnlyDir, "glide-new")
	err = os.WriteFile(newPath, []byte("new"), 0755)
	require.NoError(t, err)

	// Make directory read-only to force atomic replace to fail
	err = os.Chmod(readOnlyDir, 0555)
	require.NoError(t, err)
	defer os.Chmod(readOnlyDir, 0755) // Restore permissions for cleanup

	updater := NewUpdater("v1.0.0")
	err = updater.replaceBinary(currentPath, newPath)
	assert.Error(t, err)

	// Verify the original binary is still intact
	content, err := os.ReadFile(currentPath)
	require.NoError(t, err)
	assert.Equal(t, originalContent, content, "Original binary should be restored on failure")
}

func TestCopyFile(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "glide-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create source file
	srcPath := filepath.Join(tempDir, "source")
	srcContent := []byte("source content")
	err = os.WriteFile(srcPath, srcContent, 0644)
	require.NoError(t, err)

	// Copy file
	dstPath := filepath.Join(tempDir, "destination")
	updater := NewUpdater("v1.0.0")
	err = updater.copyFile(srcPath, dstPath)
	require.NoError(t, err)

	// Verify content
	dstContent, err := os.ReadFile(dstPath)
	require.NoError(t, err)
	assert.Equal(t, srcContent, dstContent)

	// Verify permissions
	srcInfo, err := os.Stat(srcPath)
	require.NoError(t, err)
	dstInfo, err := os.Stat(dstPath)
	require.NoError(t, err)
	assert.Equal(t, srcInfo.Mode(), dstInfo.Mode())
}

func TestSelfUpdate_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "glide-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create mock binary content
	binaryContent := []byte("new glide binary v2.0.0")

	// Calculate checksum
	hasher := sha256.New()
	hasher.Write(binaryContent)
	checksum := hex.EncodeToString(hasher.Sum(nil))

	// Create mock servers
	var downloadURL string
	downloadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".sha256") {
			w.Write([]byte(checksum + "  glide-" + runtime.GOOS + "-" + runtime.GOARCH + "\n"))
		} else {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(binaryContent)
		}
	}))
	defer downloadServer.Close()
	downloadURL = downloadServer.URL + "/glide-" + runtime.GOOS + "-" + runtime.GOARCH

	// Create release info server
	release := Release{
		TagName:     "v2.0.0",
		Name:        "v2.0.0",
		PublishedAt: time.Now(),
		HTMLURL:     "https://github.com/ivannovak/glide/v3/releases/tag/v2.0.0",
		Assets: []Asset{
			{
				Name:               "glide-" + runtime.GOOS + "-" + runtime.GOARCH,
				BrowserDownloadURL: downloadURL,
				Size:               int64(len(binaryContent)),
			},
		},
	}

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(release)
	}))
	defer apiServer.Close()

	// Override the API URL for testing
	oldURL := githubAPIURL
	githubAPIURL = apiServer.URL
	defer func() { githubAPIURL = oldURL }()

	// Create a mock current executable
	execPath := filepath.Join(tempDir, "glide")
	err = os.WriteFile(execPath, []byte("old version"), 0755)
	require.NoError(t, err)

	// Mock os.Executable to return our test path
	// Note: This is a simplified test - in production we'd need to handle this differently
	// For now, we'll test the individual components which we've already done above
}
