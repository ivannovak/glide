package cli

import (
	"testing"

	"github.com/ivannovak/glide/v2/internal/config"
	"github.com/ivannovak/glide/v2/internal/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateMultiWorktreeMode(t *testing.T) {
	tests := []struct {
		name        string
		mode        context.DevelopmentMode
		commandName string
		wantErr     bool
	}{
		{
			name:        "valid - multi-worktree mode",
			mode:        context.ModeMultiWorktree,
			commandName: "status",
			wantErr:     false,
		},
		{
			name:        "invalid - single-repo mode",
			mode:        context.ModeSingleRepo,
			commandName: "status",
			wantErr:     true,
		},
		{
			name:        "invalid - standalone mode",
			mode:        context.ModeStandalone,
			commandName: "worktree",
			wantErr:     true,
		},
		{
			name:        "invalid - empty mode",
			mode:        "",
			commandName: "down",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &context.ProjectContext{
				DevelopmentMode: tt.mode,
			}

			err := ValidateMultiWorktreeMode(ctx, tt.commandName)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateSingleRepoMode(t *testing.T) {
	tests := []struct {
		name        string
		mode        context.DevelopmentMode
		commandName string
		wantErr     bool
	}{
		{
			name:        "valid - single-repo mode",
			mode:        context.ModeSingleRepo,
			commandName: "up",
			wantErr:     false,
		},
		{
			name:        "invalid - multi-worktree mode",
			mode:        context.ModeMultiWorktree,
			commandName: "up",
			wantErr:     true,
		},
		{
			name:        "invalid - standalone mode",
			mode:        context.ModeStandalone,
			commandName: "test",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &context.ProjectContext{
				DevelopmentMode: tt.mode,
			}

			err := ValidateSingleRepoMode(ctx, tt.commandName)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestShowModeError(t *testing.T) {
	tests := []struct {
		name         string
		currentMode  context.DevelopmentMode
		requiredMode context.DevelopmentMode
		commandName  string
	}{
		{
			name:         "multi-worktree required",
			currentMode:  context.ModeSingleRepo,
			requiredMode: context.ModeMultiWorktree,
			commandName:  "worktree",
		},
		{
			name:         "single-repo required",
			currentMode:  context.ModeMultiWorktree,
			requiredMode: context.ModeSingleRepo,
			commandName:  "up",
		},
		{
			name:         "standalone to multi-worktree",
			currentMode:  context.ModeStandalone,
			requiredMode: context.ModeMultiWorktree,
			commandName:  "status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ShowModeError(tt.currentMode, tt.requiredMode, tt.commandName)

			// Should always return an error
			require.Error(t, err)

			// Error message should mention the command
			assert.Contains(t, err.Error(), tt.commandName)
		})
	}
}

func TestShowAvailableCommands(t *testing.T) {
	tests := []struct {
		name string
		mode context.DevelopmentMode
	}{
		{
			name: "multi-worktree mode",
			mode: context.ModeMultiWorktree,
		},
		{
			name: "single-repo mode",
			mode: context.ModeSingleRepo,
		},
		{
			name: "standalone mode",
			mode: context.ModeStandalone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify it doesn't panic
			assert.NotPanics(t, func() {
				ShowAvailableCommands(tt.mode)
			})
		})
	}
}

func TestShowCommandSuggestion(t *testing.T) {
	tests := []struct {
		name             string
		attemptedCommand string
		suggestions      []string
		mode             context.DevelopmentMode
	}{
		{
			name:             "single suggestion",
			attemptedCommand: "satus",
			suggestions:      []string{"status"},
			mode:             context.ModeMultiWorktree,
		},
		{
			name:             "multiple suggestions",
			attemptedCommand: "dow",
			suggestions:      []string{"down", "down-all"},
			mode:             context.ModeMultiWorktree,
		},
		{
			name:             "no suggestions",
			attemptedCommand: "xyz",
			suggestions:      []string{},
			mode:             context.ModeSingleRepo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &context.ProjectContext{
				DevelopmentMode: tt.mode,
			}

			err := ShowCommandSuggestion(tt.attemptedCommand, tt.suggestions, ctx)

			// Should always return an error
			require.Error(t, err)

			// Error message should indicate unknown command
			assert.Contains(t, err.Error(), "unknown command")
		})
	}
}

func TestShowUnknownCommandError(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		mode     context.DevelopmentMode
		location context.LocationType
	}{
		{
			name:     "multi-worktree at root",
			command:  "unknown",
			mode:     context.ModeMultiWorktree,
			location: context.LocationRoot,
		},
		{
			name:     "multi-worktree in worktree",
			command:  "badcmd",
			mode:     context.ModeMultiWorktree,
			location: context.LocationWorktree,
		},
		{
			name:     "single-repo",
			command:  "notfound",
			mode:     context.ModeSingleRepo,
			location: context.LocationProject,
		},
		{
			name:     "standalone",
			command:  "test",
			mode:     context.ModeStandalone,
			location: context.LocationUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &context.ProjectContext{
				DevelopmentMode: tt.mode,
				Location:        tt.location,
			}
			cfg := &config.Config{}

			err := ShowUnknownCommandError(tt.command, ctx, cfg)

			// Should always return an error
			require.Error(t, err)

			// Error message should mention the command
			assert.Contains(t, err.Error(), tt.command)
		})
	}
}

func TestShowContextAwareHelp(t *testing.T) {
	tests := []struct {
		name string
		mode context.DevelopmentMode
	}{
		{
			name: "multi-worktree mode",
			mode: context.ModeMultiWorktree,
		},
		{
			name: "single-repo mode",
			mode: context.ModeSingleRepo,
		},
		{
			name: "standalone mode",
			mode: context.ModeStandalone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &context.ProjectContext{
				DevelopmentMode: tt.mode,
			}
			cfg := &config.Config{}

			err := ShowContextAwareHelp(ctx, cfg)

			// Should not error
			assert.NoError(t, err)
		})
	}
}
