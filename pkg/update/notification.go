package update

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/glide-cli/glide/v3/pkg/logging"
)

const (
	// Default check interval is 24 hours
	DefaultCheckInterval = 24 * time.Hour

	// State file name in config directory
	stateFileName = "update-state.json"

	// Quick check timeout to avoid blocking CLI startup
	quickCheckTimeout = 3 * time.Second
)

// NotificationConfig configures update notification behavior
type NotificationConfig struct {
	// Enabled controls whether update checks are performed
	Enabled bool `json:"enabled" yaml:"enabled"`

	// CheckInterval is the duration between update checks
	CheckInterval time.Duration `json:"check_interval" yaml:"check_interval"`

	// LastNotifiedVersion is the version we last notified about (to avoid repeat notifications)
	LastNotifiedVersion string `json:"last_notified_version,omitempty" yaml:"last_notified_version,omitempty"`
}

// DefaultNotificationConfig returns the default notification configuration
func DefaultNotificationConfig() *NotificationConfig {
	return &NotificationConfig{
		Enabled:       true,
		CheckInterval: DefaultCheckInterval,
	}
}

// UpdateState persists update check state between CLI invocations
type UpdateState struct {
	// LastCheckTime is when we last checked for updates
	LastCheckTime time.Time `json:"last_check_time"`

	// LatestVersion is the latest version found during last check
	LatestVersion string `json:"latest_version,omitempty"`

	// LatestVersionInfo contains full update info if available
	LatestVersionInfo *UpdateInfo `json:"latest_version_info,omitempty"`

	// LastNotifiedVersion is the version we last showed a notification for
	LastNotifiedVersion string `json:"last_notified_version,omitempty"`
}

// NotificationManager handles background update checking and notifications
type NotificationManager struct {
	config         *NotificationConfig
	currentVersion string
	stateDir       string
	state          *UpdateState
	mu             sync.RWMutex
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(currentVersion string, config *NotificationConfig) *NotificationManager {
	if config == nil {
		config = DefaultNotificationConfig()
	}

	// Get state directory (typically ~/.glide/)
	stateDir := getStateDir()

	nm := &NotificationManager{
		config:         config,
		currentVersion: currentVersion,
		stateDir:       stateDir,
		state:          &UpdateState{},
	}

	// Load existing state
	nm.loadState()

	return nm
}

// getStateDir returns the directory for storing update state
func getStateDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".glide")
}

// statePath returns the full path to the state file
func (nm *NotificationManager) statePath() string {
	if nm.stateDir == "" {
		return ""
	}
	return filepath.Join(nm.stateDir, stateFileName)
}

// loadState loads persisted state from disk
func (nm *NotificationManager) loadState() {
	path := nm.statePath()
	if path == "" {
		return
	}

	data, err := os.ReadFile(path)
	if err != nil {
		// File doesn't exist or can't be read - start fresh
		logging.Debug("No update state file found, starting fresh", "path", path)
		return
	}

	var state UpdateState
	if err := json.Unmarshal(data, &state); err != nil {
		logging.Debug("Failed to parse update state, starting fresh", "error", err)
		return
	}

	nm.mu.Lock()
	nm.state = &state
	nm.mu.Unlock()

	logging.Debug("Loaded update state",
		"last_check", state.LastCheckTime,
		"latest_version", state.LatestVersion)
}

// saveState persists state to disk
func (nm *NotificationManager) saveState() error {
	path := nm.statePath()
	if path == "" {
		return fmt.Errorf("no state directory available")
	}

	nm.mu.RLock()
	data, err := json.MarshalIndent(nm.state, "", "  ")
	nm.mu.RUnlock()

	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(nm.stateDir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// ShouldCheck returns true if enough time has passed since the last check
func (nm *NotificationManager) ShouldCheck() bool {
	if !nm.config.Enabled {
		return false
	}

	// Skip check for dev builds
	if nm.currentVersion == "dev" {
		return false
	}

	nm.mu.RLock()
	lastCheck := nm.state.LastCheckTime
	nm.mu.RUnlock()

	return time.Since(lastCheck) > nm.config.CheckInterval
}

// CheckForUpdateAsync performs a non-blocking update check
// It runs in a goroutine and updates state if a new version is found
// Returns a channel that receives the result (or nil on timeout/error)
func (nm *NotificationManager) CheckForUpdateAsync(ctx context.Context) <-chan *UpdateInfo {
	resultChan := make(chan *UpdateInfo, 1)

	go func() {
		defer close(resultChan)

		// Use quick timeout to avoid blocking CLI
		checkCtx, cancel := context.WithTimeout(ctx, quickCheckTimeout)
		defer cancel()

		checker := NewChecker(nm.currentVersion)
		info, err := checker.CheckForUpdate(checkCtx)

		if err != nil {
			logging.Debug("Background update check failed", "error", err)
			return
		}

		// Update state
		nm.mu.Lock()
		nm.state.LastCheckTime = time.Now()
		nm.state.LatestVersion = info.LatestVersion
		if info.Available {
			nm.state.LatestVersionInfo = info
		}
		nm.mu.Unlock()

		// Save state synchronously within this goroutine
		// (must complete before process exits for state to persist)
		if err := nm.saveState(); err != nil {
			logging.Debug("Failed to save update state", "error", err)
		}

		if info.Available {
			resultChan <- info
		}
	}()

	return resultChan
}

// GetCachedUpdateInfo returns cached update info if available and not yet notified
func (nm *NotificationManager) GetCachedUpdateInfo() *UpdateInfo {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	info := nm.state.LatestVersionInfo
	if info == nil || !info.Available {
		return nil
	}

	// Don't show notification if we already notified about this version
	if nm.state.LastNotifiedVersion == info.LatestVersion {
		return nil
	}

	return info
}

// MarkNotified marks the current latest version as notified
func (nm *NotificationManager) MarkNotified(version string) {
	nm.mu.Lock()
	nm.state.LastNotifiedVersion = version
	nm.mu.Unlock()

	// Save state synchronously to ensure it persists before process exits
	if err := nm.saveState(); err != nil {
		logging.Debug("Failed to save notification state", "error", err)
	}
}

// FormatNotification creates a concise update notification message
func FormatNotification(info *UpdateInfo) string {
	if info == nil || !info.Available {
		return ""
	}

	return fmt.Sprintf(
		"\n💡 Update available: %s → %s\n   Run 'glide self-update' to upgrade\n",
		info.CurrentVersion,
		info.LatestVersion,
	)
}

// FormatNotificationCompact creates a single-line update notification
func FormatNotificationCompact(info *UpdateInfo) string {
	if info == nil || !info.Available {
		return ""
	}

	return fmt.Sprintf(
		"💡 Update available: %s → %s (run 'glide self-update')",
		info.CurrentVersion,
		info.LatestVersion,
	)
}
