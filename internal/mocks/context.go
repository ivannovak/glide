package mocks

import (
	"github.com/ivannovak/glide/pkg/interfaces"
	"github.com/stretchr/testify/mock"
)

// ContextDetector is a mock implementation of interfaces.ContextDetector
type ContextDetector struct {
	mock.Mock
}

// Detect mocks the Detect method
func (m *ContextDetector) Detect(workingDir string) (interfaces.ProjectContext, error) {
	args := m.Called(workingDir)
	if result := args.Get(0); result != nil {
		return result.(interfaces.ProjectContext), args.Error(1)
	}
	return nil, args.Error(1)
}

// DetectWithRoot mocks the DetectWithRoot method
func (m *ContextDetector) DetectWithRoot(workingDir, projectRoot string) (interfaces.ProjectContext, error) {
	args := m.Called(workingDir, projectRoot)
	if result := args.Get(0); result != nil {
		return result.(interfaces.ProjectContext), args.Error(1)
	}
	return nil, args.Error(1)
}

// ProjectContext is a mock implementation of interfaces.ProjectContext
type ProjectContext struct {
	mock.Mock
}

// GetWorkingDir returns the mocked working directory
func (m *ProjectContext) GetWorkingDir() string {
	args := m.Called()
	return args.String(0)
}

// GetProjectRoot returns the mocked project root
func (m *ProjectContext) GetProjectRoot() string {
	args := m.Called()
	return args.String(0)
}

// GetDevelopmentMode returns the mocked development mode
func (m *ProjectContext) GetDevelopmentMode() string {
	args := m.Called()
	return args.String(0)
}

// GetLocation returns the mocked location
func (m *ProjectContext) GetLocation() string {
	args := m.Called()
	return args.String(0)
}

// IsDockerRunning returns whether Docker is running
func (m *ProjectContext) IsDockerRunning() bool {
	args := m.Called()
	return args.Bool(0)
}

// GetComposeFiles returns the mocked compose files
func (m *ProjectContext) GetComposeFiles() []string {
	args := m.Called()
	if result := args.Get(0); result != nil {
		return result.([]string)
	}
	return nil
}

// IsWorktree returns whether this is a worktree
func (m *ProjectContext) IsWorktree() bool {
	args := m.Called()
	return args.Bool(0)
}

// GetWorktreeName returns the worktree name
func (m *ProjectContext) GetWorktreeName() string {
	args := m.Called()
	return args.String(0)
}
