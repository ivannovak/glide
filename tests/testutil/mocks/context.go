package mocks

//lint:file-ignore SA1019 interfaces.ProjectContext is deprecated but still in use until v3.0.0

import (
	"github.com/ivannovak/glide/v3/pkg/interfaces"
	"github.com/stretchr/testify/mock"
)

// MockContextDetector is a mock implementation of the ContextDetector interface
type MockContextDetector struct {
	mock.Mock
}

// Detect mocks the Detect method
func (m *MockContextDetector) Detect(workingDir string) (interfaces.ProjectContext, error) {
	args := m.Called(workingDir)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(interfaces.ProjectContext), args.Error(1)
}

// DetectWithRoot mocks the DetectWithRoot method
func (m *MockContextDetector) DetectWithRoot(workingDir, projectRoot string) (interfaces.ProjectContext, error) {
	args := m.Called(workingDir, projectRoot)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(interfaces.ProjectContext), args.Error(1)
}

// MockProjectContext is a mock implementation of the ProjectContext interface
type MockProjectContext struct {
	mock.Mock
}

// GetWorkingDir mocks the GetWorkingDir method
func (m *MockProjectContext) GetWorkingDir() string {
	args := m.Called()
	return args.String(0)
}

// GetProjectRoot mocks the GetProjectRoot method
func (m *MockProjectContext) GetProjectRoot() string {
	args := m.Called()
	return args.String(0)
}

// GetDevelopmentMode mocks the GetDevelopmentMode method
func (m *MockProjectContext) GetDevelopmentMode() string {
	args := m.Called()
	return args.String(0)
}

// GetLocation mocks the GetLocation method
func (m *MockProjectContext) GetLocation() string {
	args := m.Called()
	return args.String(0)
}

// IsDockerRunning mocks the IsDockerRunning method
func (m *MockProjectContext) IsDockerRunning() bool {
	args := m.Called()
	return args.Bool(0)
}

// GetComposeFiles mocks the GetComposeFiles method
func (m *MockProjectContext) GetComposeFiles() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]string)
}

// IsWorktree mocks the IsWorktree method
func (m *MockProjectContext) IsWorktree() bool {
	args := m.Called()
	return args.Bool(0)
}

// GetWorktreeName mocks the GetWorktreeName method
func (m *MockProjectContext) GetWorktreeName() string {
	args := m.Called()
	return args.String(0)
}

// ExpectContextDetection is a helper to set up expected context detection
func ExpectContextDetection(m *MockContextDetector, workingDir string, ctx interfaces.ProjectContext, err error) *mock.Call {
	return m.On("Detect", workingDir).Return(ctx, err)
}

// ExpectContextDetectionWithRoot is a helper to set up expected context detection with root
func ExpectContextDetectionWithRoot(m *MockContextDetector, workingDir, projectRoot string, ctx interfaces.ProjectContext, err error) *mock.Call {
	return m.On("DetectWithRoot", workingDir, projectRoot).Return(ctx, err)
}
