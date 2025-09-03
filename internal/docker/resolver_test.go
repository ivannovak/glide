package docker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ivannovak/glide/internal/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewResolver(t *testing.T) {
	ctx := &context.ProjectContext{
		WorkingDir:      "/home/user/project/vcs",
		ProjectRoot:     "/home/user/project",
		DevelopmentMode: "standard",
		Location:        "main",
		ComposeFiles:    []string{"docker-compose.yml"},
	}

	resolver := NewResolver(ctx)
	assert.NotNil(t, resolver)
	assert.Equal(t, ctx, resolver.ctx)
}

func TestResolver_GetComposeFiles(t *testing.T) {
	tests := []struct {
		name     string
		ctx      *context.ProjectContext
		expected []string
	}{
		{
			name: "single compose file",
			ctx: &context.ProjectContext{
				ComposeFiles: []string{"docker-compose.yml"},
			},
			expected: []string{"docker-compose.yml"},
		},
		{
			name: "multiple compose files",
			ctx: &context.ProjectContext{
				ComposeFiles: []string{
					"docker-compose.yml",
					"docker-compose.override.yml",
				},
			},
			expected: []string{
				"docker-compose.yml",
				"docker-compose.override.yml",
			},
		},
		{
			name: "no compose files",
			ctx: &context.ProjectContext{
				ComposeFiles: []string{},
			},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewResolver(tt.ctx)
			files := resolver.GetComposeFiles()
			assert.Equal(t, tt.expected, files)
		})
	}
}

// These test cases are replaced with simpler tests as the methods don't exist

func TestResolver_Resolve(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()
	vcsDir := filepath.Join(tempDir, "vcs")
	err := os.MkdirAll(vcsDir, 0755)
	require.NoError(t, err)

	// Create docker-compose.yml
	composeFile := filepath.Join(vcsDir, "docker-compose.yml")
	err = os.WriteFile(composeFile, []byte("version: '3'"), 0644)
	require.NoError(t, err)

	// Create override file
	overrideFile := filepath.Join(tempDir, "docker-compose.override.yml")
	err = os.WriteFile(overrideFile, []byte("version: '3'"), 0644)
	require.NoError(t, err)

	// Create worktrees directory for the second test
	worktreesDir := filepath.Join(tempDir, "worktrees", "feature")
	err = os.MkdirAll(worktreesDir, 0755)
	require.NoError(t, err)

	// Create docker-compose.yml in worktree
	worktreeComposeFile := filepath.Join(worktreesDir, "docker-compose.yml")
	err = os.WriteFile(worktreeComposeFile, []byte("version: '3'"), 0644)
	require.NoError(t, err)

	tests := []struct {
		name        string
		ctx         *context.ProjectContext
		expectError bool
		expectFiles int
	}{
		{
			name: "multi-worktree main repo mode",
			ctx: &context.ProjectContext{
				WorkingDir:      vcsDir,
				ProjectRoot:     tempDir,
				DevelopmentMode: context.ModeMultiWorktree,
				Location:        context.LocationMainRepo,
				ComposeFiles:    []string{},
			},
			expectError: false,
			expectFiles: 2, // main file + override file
		},
		{
			name: "multi-worktree worktree mode",
			ctx: &context.ProjectContext{
				WorkingDir:      worktreesDir,
				ProjectRoot:     tempDir,
				DevelopmentMode: context.ModeMultiWorktree,
				Location:        context.LocationWorktree,
				WorktreeName:    "feature",
				ComposeFiles:    []string{},
			},
			expectError: false,
			expectFiles: 2, // main file + override file
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewResolver(tt.ctx)
			err := resolver.Resolve()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, resolver.ctx.ComposeFiles, tt.expectFiles)
			}
		})
	}
}

func TestResolver_ValidateSetup(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()
	vcsDir := filepath.Join(tempDir, "vcs")
	err := os.MkdirAll(vcsDir, 0755)
	require.NoError(t, err)

	// Create docker-compose.yml
	composeFile := filepath.Join(vcsDir, "docker-compose.yml")
	err = os.WriteFile(composeFile, []byte("version: '3'"), 0644)
	require.NoError(t, err)

	tests := []struct {
		name           string
		ctx            *context.ProjectContext
		expectError    bool
		skipDockerTest bool
	}{
		{
			name: "valid setup with compose files (skip Docker daemon check)",
			ctx: &context.ProjectContext{
				WorkingDir:    vcsDir,
				ComposeFiles:  []string{composeFile},
				DockerRunning: true, // Simulate Docker running for this test
			},
			expectError:    false,
			skipDockerTest: true, // We'll mock the Docker check
		},
		{
			name: "no compose files",
			ctx: &context.ProjectContext{
				WorkingDir:   vcsDir,
				ComposeFiles: []string{},
			},
			expectError:    true,
			skipDockerTest: false,
		},
		{
			name: "non-existent compose file",
			ctx: &context.ProjectContext{
				WorkingDir:   vcsDir,
				ComposeFiles: []string{"/nonexistent/docker-compose.yml"},
			},
			expectError:    true,
			skipDockerTest: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewResolver(tt.ctx)
			err := resolver.ValidateSetup()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestContainerManager_New(t *testing.T) {
	ctx := &context.ProjectContext{
		WorkingDir:   "/home/user/project/vcs",
		ComposeFiles: []string{"docker-compose.yml"},
	}

	manager := NewContainerManager(ctx)
	assert.NotNil(t, manager)
	assert.Equal(t, ctx, manager.ctx)
}

// Mock implementations for testing
type mockShellExecutor struct{}

func (m *mockShellExecutor) Execute(cmd interface{}) (interface{}, error) {
	return nil, nil
}

func (m *mockShellExecutor) ExecuteWithTimeout(cmd interface{}, timeout interface{}) (interface{}, error) {
	return nil, nil
}

func (m *mockShellExecutor) ExecuteWithProgress(cmd interface{}, message string) error {
	return nil
}

type mockOutputManager struct{}

func (m *mockOutputManager) Display(data interface{}) error {
	return nil
}

func (m *mockOutputManager) Info(format string, args ...interface{}) error {
	return nil
}

func (m *mockOutputManager) Success(format string, args ...interface{}) error {
	return nil
}

func (m *mockOutputManager) Error(format string, args ...interface{}) error {
	return nil
}

func (m *mockOutputManager) Warning(format string, args ...interface{}) error {
	return nil
}

func (m *mockOutputManager) Raw(text string) error {
	return nil
}

func (m *mockOutputManager) Printf(format string, args ...interface{}) error {
	return nil
}

func (m *mockOutputManager) Println(args ...interface{}) error {
	return nil
}
