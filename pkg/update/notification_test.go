package update

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultNotificationConfig(t *testing.T) {
	config := DefaultNotificationConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, DefaultCheckInterval, config.CheckInterval)
	assert.Empty(t, config.LastNotifiedVersion)
}

func TestNotificationManager_ShouldCheck(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		config         *NotificationConfig
		lastCheckTime  time.Time
		want           bool
	}{
		{
			name:           "should check when never checked before",
			currentVersion: "1.0.0",
			config:         DefaultNotificationConfig(),
			lastCheckTime:  time.Time{}, // zero value
			want:           true,
		},
		{
			name:           "should check when interval has passed",
			currentVersion: "1.0.0",
			config:         DefaultNotificationConfig(),
			lastCheckTime:  time.Now().Add(-25 * time.Hour),
			want:           true,
		},
		{
			name:           "should not check when checked recently",
			currentVersion: "1.0.0",
			config:         DefaultNotificationConfig(),
			lastCheckTime:  time.Now().Add(-1 * time.Hour),
			want:           false,
		},
		{
			name:           "should not check for dev builds",
			currentVersion: "dev",
			config:         DefaultNotificationConfig(),
			lastCheckTime:  time.Time{},
			want:           false,
		},
		{
			name:           "should not check when disabled",
			currentVersion: "1.0.0",
			config: &NotificationConfig{
				Enabled:       false,
				CheckInterval: DefaultCheckInterval,
			},
			lastCheckTime: time.Time{},
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nm := NewNotificationManager(tt.currentVersion, tt.config)
			nm.state.LastCheckTime = tt.lastCheckTime

			got := nm.ShouldCheck()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNotificationManager_StatePeristence(t *testing.T) {
	// Create temp directory for state file
	tmpDir := t.TempDir()

	// Create manager with custom state dir
	nm := &NotificationManager{
		config:         DefaultNotificationConfig(),
		currentVersion: "1.0.0",
		stateDir:       tmpDir,
		state:          &UpdateState{},
	}

	// Set some state - use UTC for consistent JSON round-trip
	checkTime := time.Now().UTC().Truncate(time.Second)
	nm.state.LastCheckTime = checkTime
	nm.state.LatestVersion = "2.0.0"
	nm.state.LatestVersionInfo = &UpdateInfo{
		Available:      true,
		CurrentVersion: "1.0.0",
		LatestVersion:  "2.0.0",
	}

	// Save state
	err := nm.saveState()
	require.NoError(t, err)

	// Verify file exists
	statePath := filepath.Join(tmpDir, stateFileName)
	assert.FileExists(t, statePath)

	// Create new manager and load state
	nm2 := &NotificationManager{
		config:         DefaultNotificationConfig(),
		currentVersion: "1.0.0",
		stateDir:       tmpDir,
		state:          &UpdateState{},
	}
	nm2.loadState()

	// Verify state was loaded
	assert.Equal(t, checkTime, nm2.state.LastCheckTime)
	assert.Equal(t, "2.0.0", nm2.state.LatestVersion)
	assert.NotNil(t, nm2.state.LatestVersionInfo)
	assert.True(t, nm2.state.LatestVersionInfo.Available)
}

func TestNotificationManager_GetCachedUpdateInfo(t *testing.T) {
	// Use temp directory to avoid loading real state from ~/.glide
	tmpDir := t.TempDir()

	nm := &NotificationManager{
		config:         DefaultNotificationConfig(),
		currentVersion: "1.0.0",
		stateDir:       tmpDir,
		state:          &UpdateState{},
	}

	// No cached info initially
	assert.Nil(t, nm.GetCachedUpdateInfo())

	// Set cached info
	nm.state.LatestVersionInfo = &UpdateInfo{
		Available:      true,
		CurrentVersion: "1.0.0",
		LatestVersion:  "2.0.0",
	}

	// Should return cached info
	info := nm.GetCachedUpdateInfo()
	require.NotNil(t, info)
	assert.Equal(t, "2.0.0", info.LatestVersion)

	// Mark as notified
	nm.MarkNotified("2.0.0")

	// Give goroutine time to save
	time.Sleep(50 * time.Millisecond)

	// Should not return info after notification
	assert.Nil(t, nm.GetCachedUpdateInfo())
}

func TestNotificationManager_MarkNotified(t *testing.T) {
	tmpDir := t.TempDir()

	nm := &NotificationManager{
		config:         DefaultNotificationConfig(),
		currentVersion: "1.0.0",
		stateDir:       tmpDir,
		state:          &UpdateState{},
	}

	nm.MarkNotified("2.0.0")

	// Give goroutine time to save
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, "2.0.0", nm.state.LastNotifiedVersion)

	// Verify persisted
	data, err := os.ReadFile(filepath.Join(tmpDir, stateFileName))
	require.NoError(t, err)

	var state UpdateState
	require.NoError(t, json.Unmarshal(data, &state))
	assert.Equal(t, "2.0.0", state.LastNotifiedVersion)
}

func TestFormatNotification(t *testing.T) {
	tests := []struct {
		name string
		info *UpdateInfo
		want string
	}{
		{
			name: "nil info returns empty",
			info: nil,
			want: "",
		},
		{
			name: "not available returns empty",
			info: &UpdateInfo{
				Available: false,
			},
			want: "",
		},
		{
			name: "available shows notification",
			info: &UpdateInfo{
				Available:      true,
				CurrentVersion: "1.0.0",
				LatestVersion:  "2.0.0",
			},
			want: "\n💡 Update available: 1.0.0 → 2.0.0\n   Run 'glide self-update' to upgrade\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatNotification(tt.info)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatNotificationCompact(t *testing.T) {
	info := &UpdateInfo{
		Available:      true,
		CurrentVersion: "1.0.0",
		LatestVersion:  "2.0.0",
	}

	got := FormatNotificationCompact(info)
	assert.Contains(t, got, "1.0.0 → 2.0.0")
	assert.Contains(t, got, "self-update")
}

func TestNotificationManager_CheckForUpdateAsync(t *testing.T) {
	// This test uses a short timeout to avoid actually calling GitHub
	nm := NewNotificationManager("dev", DefaultNotificationConfig())

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	resultChan := nm.CheckForUpdateAsync(ctx)

	// Should complete without hanging
	select {
	case result := <-resultChan:
		// For dev version, should return nil (no update check performed)
		assert.Nil(t, result)
	case <-time.After(1 * time.Second):
		t.Fatal("CheckForUpdateAsync timed out")
	}
}

func TestUpdateState_JSON(t *testing.T) {
	state := &UpdateState{
		LastCheckTime:       time.Now().UTC().Truncate(time.Second),
		LatestVersion:       "2.0.0",
		LastNotifiedVersion: "1.5.0",
		LatestVersionInfo: &UpdateInfo{
			Available:      true,
			CurrentVersion: "1.0.0",
			LatestVersion:  "2.0.0",
			ReleaseURL:     "https://github.com/test/releases/2.0.0",
		},
	}

	// Marshal
	data, err := json.Marshal(state)
	require.NoError(t, err)

	// Unmarshal
	var decoded UpdateState
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, state.LastCheckTime, decoded.LastCheckTime)
	assert.Equal(t, state.LatestVersion, decoded.LatestVersion)
	assert.Equal(t, state.LastNotifiedVersion, decoded.LastNotifiedVersion)
	assert.NotNil(t, decoded.LatestVersionInfo)
	assert.Equal(t, state.LatestVersionInfo.LatestVersion, decoded.LatestVersionInfo.LatestVersion)
}
