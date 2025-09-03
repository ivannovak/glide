package cli

import (
	"testing"

	"github.com/ivannovak/glide/internal/config"
	internalContext "github.com/ivannovak/glide/internal/context"
	"github.com/ivannovak/glide/pkg/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionCommand_Basic(t *testing.T) {
	// Save original values
	oldVersion := version.Version
	oldBuildDate := version.BuildDate
	oldGitCommit := version.GitCommit
	defer func() {
		version.Version = oldVersion
		version.BuildDate = oldBuildDate
		version.GitCommit = oldGitCommit
	}()

	// Set test values
	version.Version = "v1.0.0"
	version.BuildDate = "2025-01-01T00:00:00Z"
	version.GitCommit = "abc123"

	// Create test context
	ctx := &internalContext.ProjectContext{}
	cfg := &config.Config{}

	// Create command
	cmd := NewVersionCommand(ctx, cfg)
	require.NotNil(t, cmd)

	// Execute command - output goes to stdout not the buffer
	err := cmd.Execute()
	require.NoError(t, err)

	// The command prints to stdout via output package, not to the buffer
	// We can verify the command structure itself
	assert.Equal(t, "version [flags]", cmd.Use)
	assert.Contains(t, cmd.Short, "Display version information")
}

func TestVersionCommand_Flags(t *testing.T) {
	// Create test context
	ctx := &internalContext.ProjectContext{}
	cfg := &config.Config{}

	// Create command
	cmd := NewVersionCommand(ctx, cfg)
	require.NotNil(t, cmd)

	// Check that check-update flag exists
	checkUpdateFlag := cmd.Flag("check-update")
	assert.NotNil(t, checkUpdateFlag)
	assert.Equal(t, "check-update", checkUpdateFlag.Name)
	assert.Equal(t, "false", checkUpdateFlag.DefValue)
}

func TestVersionCommand_Structure(t *testing.T) {
	// Create test context
	ctx := &internalContext.ProjectContext{}
	cfg := &config.Config{}

	// Create command
	cmd := NewVersionCommand(ctx, cfg)
	require.NotNil(t, cmd)

	// Verify command metadata
	assert.Equal(t, "version [flags]", cmd.Use)
	assert.Equal(t, "Display version information", cmd.Short)
	assert.Contains(t, cmd.Long, "Display version information for Glide CLI")
	assert.Contains(t, cmd.Long, "Go version")
	assert.Contains(t, cmd.Long, "Operating system")
	assert.Contains(t, cmd.Long, "check-update")

	// Verify silence settings
	assert.True(t, cmd.SilenceUsage)
	assert.True(t, cmd.SilenceErrors)
}

func TestVersionCommand_DevVersion(t *testing.T) {
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
	cmd := NewVersionCommand(ctx, cfg)
	require.NotNil(t, cmd)

	// Execute with check-update flag
	cmd.SetArgs([]string{"--check-update"})

	// Should execute without error even with dev version
	err := cmd.Execute()
	require.NoError(t, err)
}
