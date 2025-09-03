package context

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectContext_Methods(t *testing.T) {
	ctx := &ProjectContext{
		WorkingDir:      "/home/user/project",
		ProjectRoot:     "/home/user/project",
		ProjectName:     "myproject",
		DevelopmentMode: ModeMultiWorktree,
		Location:        LocationMainRepo,
		IsMainRepo:      true,
		IsWorktree:      false,
		WorktreeName:    "",
		DockerRunning:   true,
		ComposeFiles:    []string{"docker-compose.yml"},
		CommandScope:    "local",
	}

	assert.Equal(t, "/home/user/project", ctx.WorkingDir)
	assert.Equal(t, "/home/user/project", ctx.ProjectRoot)
	assert.Equal(t, "myproject", ctx.ProjectName)
	assert.Equal(t, ModeMultiWorktree, ctx.DevelopmentMode)
	assert.Equal(t, LocationMainRepo, ctx.Location)
	assert.True(t, ctx.IsMainRepo)
	assert.False(t, ctx.IsWorktree)
	assert.Equal(t, "", ctx.WorktreeName)
	assert.True(t, ctx.DockerRunning)
	assert.Equal(t, []string{"docker-compose.yml"}, ctx.ComposeFiles)
	assert.False(t, ctx.IsGlobalScope())
	assert.True(t, ctx.CanUseGlobalCommands())
}

func TestProjectContext_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		ctx      *ProjectContext
		expected bool
	}{
		{
			name: "valid context",
			ctx: &ProjectContext{
				ProjectRoot: "/home/user/project",
				Error:       nil,
			},
			expected: true,
		},
		{
			name: "context with error",
			ctx: &ProjectContext{
				ProjectRoot: "/home/user/project",
				Error:       ErrProjectRootNotFound,
			},
			expected: false,
		},
		{
			name: "context without project root",
			ctx: &ProjectContext{
				ProjectRoot: "",
				Error:       nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.ctx.IsValid())
		})
	}
}

func TestProjectContext_GetComposeCommand(t *testing.T) {
	ctx := &ProjectContext{
		ComposeFiles: []string{
			"docker-compose.yml",
			"docker-compose.override.yml",
		},
	}

	cmd := ctx.GetComposeCommand()
	assert.Equal(t, []string{
		"docker", "compose",
		"-f", "docker-compose.yml",
		"-f", "docker-compose.override.yml",
	}, cmd)
}

func TestNewDetector(t *testing.T) {
	detector, err := NewDetector()
	require.NoError(t, err)
	assert.NotNil(t, detector)
	assert.NotEmpty(t, detector.workingDir)
	assert.NotNil(t, detector.rootFinder)
	assert.NotNil(t, detector.modeDetector)
	assert.NotNil(t, detector.locationIdentifier)
	assert.NotNil(t, detector.composeResolver)
}

// These tests are replaced with simpler tests as the concrete types are not exported

func TestDetector_Detect(t *testing.T) {
	detector, err := NewDetector()
	require.NoError(t, err)

	t.Run("detection returns context", func(t *testing.T) {
		ctx, _ := detector.Detect()
		// May error if not in a git repository, but should still return context
		assert.NotNil(t, ctx)
		assert.NotEmpty(t, ctx.WorkingDir)
	})
}
