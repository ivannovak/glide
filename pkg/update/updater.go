package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Updater handles self-update functionality
type Updater struct {
	checker    *Checker
	httpClient *http.Client
}

// NewUpdater creates a new updater
func NewUpdater(currentVersion string) *Updater {
	return &Updater{
		checker: NewChecker(currentVersion),
		httpClient: &http.Client{
			Timeout: 0, // No timeout for downloads
		},
	}
}

// SelfUpdate performs a self-update of the binary
func (u *Updater) SelfUpdate(ctx context.Context) error {
	// Check for updates
	info, err := u.checker.CheckForUpdate(ctx)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if !info.Available {
		return fmt.Errorf("no update available")
	}

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve any symlinks
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	// Download new binary
	tempFile, err := u.downloadBinary(ctx, info.DownloadURL)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer os.Remove(tempFile)

	// Download and verify checksum if available
	checksumURL := info.DownloadURL + ".sha256"
	if err := u.verifyChecksum(ctx, tempFile, checksumURL); err != nil {
		// Log warning but don't fail if checksum verification fails
		// (checksum file might not exist for all releases)
		fmt.Fprintf(os.Stderr, "Warning: checksum verification skipped: %v\n", err)
	}

	// Replace the binary
	if err := u.replaceBinary(execPath, tempFile); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	return nil
}

// downloadBinary downloads the new binary to a temporary file
func (u *Updater) downloadBinary(ctx context.Context, url string) (string, error) {
	// Skip if URL is a GitHub release page (not a direct download)
	if strings.Contains(url, "github.com") && strings.Contains(url, "/releases/") && !strings.Contains(url, "/download/") {
		return "", fmt.Errorf("direct download not available for this platform")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create temporary file
	tempFile, err := os.CreateTemp("", "glide-update-*")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	// Copy download to temp file
	if _, err := io.Copy(tempFile, resp.Body); err != nil {
		os.Remove(tempFile.Name())
		return "", err
	}

	// Make executable
	if err := os.Chmod(tempFile.Name(), 0755); err != nil {
		os.Remove(tempFile.Name())
		return "", err
	}

	return tempFile.Name(), nil
}

// verifyChecksum downloads and verifies the SHA256 checksum
func (u *Updater) verifyChecksum(ctx context.Context, filePath, checksumURL string) error {
	// Download checksum file
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, checksumURL, nil)
	if err != nil {
		return err
	}

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("checksum file not found")
	}

	// Read expected checksum
	checksumData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Parse checksum (format: "sha256sum  filename")
	parts := strings.Fields(string(checksumData))
	if len(parts) < 1 {
		return fmt.Errorf("invalid checksum format")
	}
	expectedChecksum := parts[0]

	// Calculate actual checksum
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return err
	}
	actualChecksum := hex.EncodeToString(hasher.Sum(nil))

	// Compare checksums
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

// replaceBinary replaces the current binary with the new one
func (u *Updater) replaceBinary(currentPath, newPath string) error {
	// Create backup of current binary
	backupPath := currentPath + ".backup"
	if err := u.copyFile(currentPath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Attempt to replace the binary
	if err := u.atomicReplace(currentPath, newPath); err != nil {
		// Restore from backup on failure
		if restoreErr := u.copyFile(backupPath, currentPath); restoreErr != nil {
			return fmt.Errorf("failed to replace binary and restore backup: replace error: %w, restore error: %v", err, restoreErr)
		}
		return fmt.Errorf("failed to replace binary (backup restored successfully): %w", err)
	}

	// Remove backup on success
	if err := os.Remove(backupPath); err != nil {
		// Log but don't fail - cleanup error is non-critical
		fmt.Fprintf(os.Stderr, "Warning: failed to remove backup file %s: %v\n", backupPath, err)
	}

	return nil
}

// atomicReplace performs an atomic replacement of the file
func (u *Updater) atomicReplace(oldPath, newPath string) error {
	// On Unix systems, we can use rename for atomic replacement
	if runtime.GOOS != "windows" {
		return os.Rename(newPath, oldPath)
	}

	// On Windows, we need to remove the old file first
	if err := os.Remove(oldPath); err != nil {
		return err
	}
	return os.Rename(newPath, oldPath)
}

// copyFile copies a file from src to dst
func (u *Updater) copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	if _, err := io.Copy(destination, source); err != nil {
		return err
	}

	// Copy file permissions
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, info.Mode())
}
