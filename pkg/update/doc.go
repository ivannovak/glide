// Package update provides self-update functionality for the Glide CLI.
//
// This package checks for new releases on GitHub and can download and
// install updates automatically. It uses semantic versioning for
// version comparison.
//
// # Checking for Updates
//
// Check if a new version is available:
//
//	info, err := update.Check(currentVersion)
//	if err != nil {
//	    log.Error("Failed to check for updates", "error", err)
//	    return
//	}
//
//	if info.Available {
//	    fmt.Printf("Update available: %s -> %s\n",
//	        info.CurrentVersion, info.LatestVersion)
//	    fmt.Printf("Release notes: %s\n", info.ReleaseNotes)
//	}
//
// # Performing Updates
//
// Download and install the latest version:
//
//	updater := update.NewUpdater()
//	err := updater.Update(info)
//	if err != nil {
//	    log.Error("Update failed", "error", err)
//	    return
//	}
//	fmt.Println("Update complete! Please restart.")
//
// # Update Information
//
// The UpdateInfo struct contains details about available updates:
//
//	type UpdateInfo struct {
//	    Available      bool
//	    CurrentVersion string
//	    LatestVersion  string
//	    ReleaseURL     string
//	    ReleaseNotes   string
//	    DownloadURL    string
//	}
//
// # Platform Detection
//
// Downloads are automatically selected for the current platform:
//   - darwin/amd64, darwin/arm64 (macOS)
//   - linux/amd64, linux/arm64 (Linux)
//   - windows/amd64 (Windows)
//
// # Security
//
// Updates are verified by:
//   - HTTPS-only downloads
//   - SHA256 checksum verification (when available)
//   - Binary signature verification (when available)
package update
