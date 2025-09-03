package context

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStandardProjectRootFinder(t *testing.T) {
	finder := NewStandardProjectRootFinder()
	assert.NotNil(t, finder)
}

func TestStandardProjectRootFinder_FindRoot(t *testing.T) {
	finder := NewStandardProjectRootFinder()

	// Create a temporary directory structure for testing
	tempDir := t.TempDir()
	gitDir := filepath.Join(tempDir, ".git")
	err := os.Mkdir(gitDir, 0755)
	require.NoError(t, err)

	subDir := filepath.Join(tempDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	require.NoError(t, err)

	tests := []struct {
		name        string
		workingDir  string
		expectRoot  string
		expectError bool
	}{
		{
			name:        "find root from root dir",
			workingDir:  tempDir,
			expectRoot:  tempDir,
			expectError: false,
		},
		{
			name:        "find root from subdirectory",
			workingDir:  subDir,
			expectRoot:  tempDir,
			expectError: false,
		},
		{
			name:        "no git root found",
			workingDir:  "/tmp",
			expectRoot:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := finder.FindRoot(tt.workingDir)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectRoot, root)
			}
		})
	}
}

func TestNewStandardDevelopmentModeDetector(t *testing.T) {
	detector := NewStandardDevelopmentModeDetector()
	assert.NotNil(t, detector)
}

func TestStandardDevelopmentModeDetector_DetectMode(t *testing.T) {
	detector := NewStandardDevelopmentModeDetector()

	tests := []struct {
		name        string
		projectRoot string
		expected    DevelopmentMode
	}{
		{
			name:        "single repo mode - project root with git",
			projectRoot: "/home/user/project",
			expected:    ModeSingleRepo,
		},
		{
			name:        "multi-worktree mode - project root with vcs",
			projectRoot: "/home/user/project",
			expected:    ModeMultiWorktree,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create necessary directory structure for tests
			tempDir := t.TempDir()

			if tt.expected == ModeMultiWorktree {
				// Create temporary vcs directory to simulate multi-worktree structure
				vcsDir := filepath.Join(tempDir, "vcs")
				err := os.Mkdir(vcsDir, 0755)
				require.NoError(t, err)

				worktreesDir := filepath.Join(tempDir, "worktrees")
				err = os.Mkdir(worktreesDir, 0755)
				require.NoError(t, err)

				mode := detector.DetectMode(tempDir)
				assert.Equal(t, ModeMultiWorktree, mode)
			} else {
				// Create git directory for single repo
				gitDir := filepath.Join(tempDir, ".git")
				err := os.Mkdir(gitDir, 0755)
				require.NoError(t, err)

				mode := detector.DetectMode(tempDir)
				assert.Equal(t, ModeSingleRepo, mode)
			}
		})
	}
}

func TestNewStandardLocationIdentifier(t *testing.T) {
	identifier := NewStandardLocationIdentifier()
	assert.NotNil(t, identifier)
}

func TestStandardLocationIdentifier_IdentifyLocation(t *testing.T) {
	identifier := NewStandardLocationIdentifier()

	tests := []struct {
		name         string
		workingDir   string
		projectRoot  string
		mode         DevelopmentMode
		expectedType LocationType
	}{
		{
			name:         "single repo project",
			workingDir:   "/home/user/project",
			projectRoot:  "/home/user/project",
			mode:         ModeSingleRepo,
			expectedType: LocationProject,
		},
		{
			name:         "multi-worktree root",
			workingDir:   "/home/user/project",
			projectRoot:  "/home/user/project",
			mode:         ModeMultiWorktree,
			expectedType: LocationRoot,
		},
		{
			name:         "multi-worktree main repo",
			workingDir:   "/home/user/project/vcs",
			projectRoot:  "/home/user/project",
			mode:         ModeMultiWorktree,
			expectedType: LocationMainRepo,
		},
		{
			name:         "multi-worktree worktree",
			workingDir:   "/home/user/project/worktrees/feature-branch",
			projectRoot:  "/home/user/project",
			mode:         ModeMultiWorktree,
			expectedType: LocationWorktree,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ProjectContext{
				ProjectRoot:     tt.projectRoot,
				DevelopmentMode: tt.mode,
			}

			locType := identifier.IdentifyLocation(ctx, tt.workingDir)
			assert.Equal(t, tt.expectedType, locType)

			// Verify context is updated correctly
			switch tt.expectedType {
			case LocationRoot:
				assert.True(t, ctx.IsRoot)
				assert.False(t, ctx.IsMainRepo)
				assert.False(t, ctx.IsWorktree)
			case LocationMainRepo:
				assert.False(t, ctx.IsRoot)
				assert.True(t, ctx.IsMainRepo)
				assert.False(t, ctx.IsWorktree)
			case LocationWorktree:
				assert.False(t, ctx.IsRoot)
				assert.False(t, ctx.IsMainRepo)
				assert.True(t, ctx.IsWorktree)
				if tt.workingDir == "/home/user/project/worktrees/feature-branch" {
					assert.Equal(t, "feature-branch", ctx.WorktreeName)
				}
			}
		})
	}
}

func TestNewStandardComposeFileResolver(t *testing.T) {
	resolver := NewStandardComposeFileResolver()
	assert.NotNil(t, resolver)
}

func TestStandardComposeFileResolver_ResolveFiles(t *testing.T) {
	resolver := NewStandardComposeFileResolver()

	tests := []struct {
		name        string
		location    LocationType
		expectFiles int
	}{
		{
			name:        "single repo project",
			location:    LocationProject,
			expectFiles: 0, // No files created in test
		},
		{
			name:        "multi-worktree main repo",
			location:    LocationMainRepo,
			expectFiles: 0, // No files created in test
		},
		{
			name:        "worktree directory",
			location:    LocationWorktree,
			expectFiles: 0, // No files created in test
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory structure
			tempDir := t.TempDir()

			ctx := &ProjectContext{
				ProjectRoot:  tempDir,
				Location:     tt.location,
				WorktreeName: "feature", // For worktree tests
			}

			files := resolver.ResolveFiles(ctx)
			assert.Len(t, files, tt.expectFiles)
		})
	}
}

func TestProjectContext_LocationHelpers(t *testing.T) {
	tests := []struct {
		name string
		ctx  *ProjectContext
		test func(*testing.T, *ProjectContext)
	}{
		{
			name: "single repo context",
			ctx: &ProjectContext{
				DevelopmentMode: ModeSingleRepo,
				Location:        LocationProject,
				IsRoot:          false,
				IsMainRepo:      false,
				IsWorktree:      false,
			},
			test: func(t *testing.T, ctx *ProjectContext) {
				assert.False(t, ctx.IsRoot)
				assert.False(t, ctx.IsMainRepo)
				assert.False(t, ctx.IsWorktree)
				assert.False(t, ctx.CanUseGlobalCommands())
			},
		},
		{
			name: "multi-worktree root",
			ctx: &ProjectContext{
				DevelopmentMode: ModeMultiWorktree,
				Location:        LocationRoot,
				IsRoot:          true,
				IsMainRepo:      false,
				IsWorktree:      false,
			},
			test: func(t *testing.T, ctx *ProjectContext) {
				assert.True(t, ctx.IsRoot)
				assert.False(t, ctx.IsMainRepo)
				assert.False(t, ctx.IsWorktree)
				assert.True(t, ctx.CanUseGlobalCommands())
			},
		},
		{
			name: "multi-worktree main repo",
			ctx: &ProjectContext{
				DevelopmentMode: ModeMultiWorktree,
				Location:        LocationMainRepo,
				IsRoot:          false,
				IsMainRepo:      true,
				IsWorktree:      false,
			},
			test: func(t *testing.T, ctx *ProjectContext) {
				assert.False(t, ctx.IsRoot)
				assert.True(t, ctx.IsMainRepo)
				assert.False(t, ctx.IsWorktree)
				assert.True(t, ctx.CanUseGlobalCommands())
			},
		},
		{
			name: "worktree context",
			ctx: &ProjectContext{
				DevelopmentMode: ModeMultiWorktree,
				Location:        LocationWorktree,
				IsRoot:          false,
				IsMainRepo:      false,
				IsWorktree:      true,
				WorktreeName:    "feature-branch",
			},
			test: func(t *testing.T, ctx *ProjectContext) {
				assert.False(t, ctx.IsRoot)
				assert.False(t, ctx.IsMainRepo)
				assert.True(t, ctx.IsWorktree)
				assert.Equal(t, "feature-branch", ctx.WorktreeName)
				assert.True(t, ctx.CanUseGlobalCommands())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t, tt.ctx)
		})
	}
}
