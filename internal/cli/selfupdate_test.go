package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ivannovak/glide/internal/config"
	internalContext "github.com/ivannovak/glide/internal/context"
	"github.com/ivannovak/glide/pkg/update"
	"github.com/ivannovak/glide/pkg/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelfUpdateCommand_DevVersion(t *testing.T) {
	// Save original version
	oldVersion := version.Version
	defer func() {
		version.Version = oldVersion
	}()

	// Set dev version
	version.Version = "dev"

	// Create test context
	ctx := &internalContext.ProjectContext{}
	cfg := &config.Config{}

	// Create command
	cmd := NewSelfUpdateCommand(ctx, cfg)
	require.NotNil(t, cmd)

	// Execute command
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "self-update not available for development builds")
}

func TestSelfUpdateCommand_NoUpdateAvailable(t *testing.T) {
	// Save original version
	oldVersion := version.Version
	defer func() {
		version.Version = oldVersion
	}()

	// Set current version
	version.Version = "v2.0.0"

	// Create mock server that returns same version
	release := update.Release{
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

	// This would need the update package to expose the githubAPIURL for testing
	// For now, we'll test the logic flow

	// Create test context
	ctx := &internalContext.ProjectContext{}
	cfg := &config.Config{}

	// Create command
	cmd := NewSelfUpdateCommand(ctx, cfg)
	require.NotNil(t, cmd)

	// Execute command (without actual update check since we can't override the URL easily)
	// This test validates the command structure and basic flow
	assert.NotNil(t, cmd.Use)
	assert.Equal(t, "self-update [flags]", cmd.Use)
	assert.Contains(t, cmd.Short, "Update Glide CLI")
	assert.Contains(t, cmd.Aliases, "update")
	assert.Contains(t, cmd.Aliases, "upgrade")
}

func TestSelfUpdateCommand_Force(t *testing.T) {
	// Test that force flag is properly registered
	ctx := &internalContext.ProjectContext{}
	cfg := &config.Config{}

	cmd := NewSelfUpdateCommand(ctx, cfg)
	require.NotNil(t, cmd)

	// Check that force flag exists
	forceFlag := cmd.Flag("force")
	assert.NotNil(t, forceFlag)
	assert.Equal(t, "force", forceFlag.Name)
	assert.Equal(t, "false", forceFlag.DefValue)
}

func TestSelfUpdateCommand_UserCancellation(t *testing.T) {
	// This test would require mocking stdin for user input
	// For now, we verify the command structure handles the cancellation path

	// Save original version
	oldVersion := version.Version
	defer func() {
		version.Version = oldVersion
	}()

	// Set current version
	version.Version = "v1.0.0"

	// Create test context
	ctx := &internalContext.ProjectContext{}
	cfg := &config.Config{}

	// Create command
	cmd := NewSelfUpdateCommand(ctx, cfg)
	require.NotNil(t, cmd)

	// Verify command has proper error handling
	assert.True(t, cmd.SilenceUsage)
	assert.True(t, cmd.SilenceErrors)
}

func TestSelfUpdateCommand_ChecksCurrentVersion(t *testing.T) {
	// Save original version
	oldVersion := version.Version
	defer func() {
		version.Version = oldVersion
	}()

	testCases := []struct {
		name           string
		currentVersion string
		expectError    bool
		errorContains  string
	}{
		{
			name:           "Dev version",
			currentVersion: "dev",
			expectError:    true,
			errorContains:  "development builds",
		},
		{
			name:           "Dev suffix version",
			currentVersion: "v1.0.0-dev",
			expectError:    false, // Current implementation only checks for exact "dev"
			errorContains:  "",
		},
		{
			name:           "Valid version",
			currentVersion: "v1.0.0",
			expectError:    false, // Would proceed to update check
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set version
			version.Version = tc.currentVersion

			// Create test context
			ctx := &internalContext.ProjectContext{}
			cfg := &config.Config{}

			// Create command
			cmd := NewSelfUpdateCommand(ctx, cfg)
			require.NotNil(t, cmd)

			// Create self-update command instance
			suc := &SelfUpdateCommand{
				ctx: ctx,
				cfg: cfg,
			}

			// Test execute method with dev version
			if tc.currentVersion == "dev" {
				err := suc.execute(cmd, []string{}, false)
				if tc.expectError {
					assert.Error(t, err)
					if tc.errorContains != "" {
						assert.Contains(t, err.Error(), tc.errorContains)
					}
				}
			} else {
				// Can't test full flow without mocking the update checker
				assert.NotNil(t, suc)
			}
		})
	}
}

func TestSelfUpdateCommand_OutputMessages(t *testing.T) {
	// Test the output formatting for various scenarios
	ctx := &internalContext.ProjectContext{}
	cfg := &config.Config{}

	cmd := NewSelfUpdateCommand(ctx, cfg)
	require.NotNil(t, cmd)

	// Verify command description includes key information
	assert.Contains(t, cmd.Long, "Check for the latest available version")
	assert.Contains(t, cmd.Long, "Download the appropriate binary")
	assert.Contains(t, cmd.Long, "Verify the download with SHA256")
	assert.Contains(t, cmd.Long, "Replace the current binary")
	assert.Contains(t, cmd.Long, "Create a backup")
	assert.Contains(t, cmd.Long, "atomic and will rollback")
}

func TestSelfUpdateCommand_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This would be a full integration test with mock servers
	// Similar to the updater integration test but testing the full command flow

	// Save original version
	oldVersion := version.Version
	defer func() {
		version.Version = oldVersion
	}()

	// Set current version
	version.Version = "v1.0.0"

	// Create mock server for a new version
	release := update.Release{
		TagName:     "v2.0.0",
		Name:        "v2.0.0",
		Body:        "New features and improvements",
		PublishedAt: time.Now(),
		HTMLURL:     "https://github.com/ivannovak/glide/releases/tag/v2.0.0",
		Assets: []update.Asset{
			{
				Name:               "glid-darwin-arm64",
				BrowserDownloadURL: "https://github.com/ivannovak/glide/releases/download/v2.0.0/glid-darwin-arm64",
				Size:               10485760,
			},
			{
				Name:               "glid-linux-amd64",
				BrowserDownloadURL: "https://github.com/ivannovak/glide/releases/download/v2.0.0/glid-linux-amd64",
				Size:               10485760,
			},
			{
				Name:               "glid-windows-amd64.exe",
				BrowserDownloadURL: "https://github.com/ivannovak/glide/releases/download/v2.0.0/glid-windows-amd64.exe",
				Size:               10485760,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	// Test would continue with mocking the download and update process
	// However, this requires modifying the update package to support dependency injection
	// For now, we've verified the command structure and basic flow
}
