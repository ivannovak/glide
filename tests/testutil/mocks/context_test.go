package mocks

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockContextDetector_Detect(t *testing.T) {
	mockDetector := new(MockContextDetector)
	mockContext := new(MockProjectContext)
	workingDir := "/test/dir"

	// Set up expectation
	mockDetector.On("Detect", workingDir).Return(mockContext, nil)

	// Execute
	ctx, err := mockDetector.Detect(workingDir)

	// Verify
	assert.NoError(t, err)
	assert.Equal(t, mockContext, ctx)
	mockDetector.AssertExpectations(t)
}

func TestMockContextDetector_DetectWithRoot(t *testing.T) {
	mockDetector := new(MockContextDetector)
	mockContext := new(MockProjectContext)
	workingDir := "/test/dir"
	projectRoot := "/test"

	// Set up expectation
	mockDetector.On("DetectWithRoot", workingDir, projectRoot).Return(mockContext, nil)

	// Execute
	ctx, err := mockDetector.DetectWithRoot(workingDir, projectRoot)

	// Verify
	assert.NoError(t, err)
	assert.Equal(t, mockContext, ctx)
	mockDetector.AssertExpectations(t)
}

func TestMockContextDetector_DetectError(t *testing.T) {
	mockDetector := new(MockContextDetector)
	workingDir := "/test/dir"
	expectedError := errors.New("detection failed")

	// Set up expectation
	mockDetector.On("Detect", workingDir).Return(nil, expectedError)

	// Execute
	ctx, err := mockDetector.Detect(workingDir)

	// Verify
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, ctx)
	mockDetector.AssertExpectations(t)
}

func TestMockProjectContext(t *testing.T) {
	mockContext := new(MockProjectContext)

	// Set up expectations
	mockContext.On("GetWorkingDir").Return("/test/working")
	mockContext.On("GetProjectRoot").Return("/test")
	mockContext.On("GetDevelopmentMode").Return("docker")
	mockContext.On("GetLocation").Return("container")
	mockContext.On("IsDockerRunning").Return(true)
	mockContext.On("GetComposeFiles").Return([]string{"docker-compose.yml"})
	mockContext.On("IsWorktree").Return(false)
	mockContext.On("GetWorktreeName").Return("")

	// Execute and verify
	assert.Equal(t, "/test/working", mockContext.GetWorkingDir())
	assert.Equal(t, "/test", mockContext.GetProjectRoot())
	assert.Equal(t, "docker", mockContext.GetDevelopmentMode())
	assert.Equal(t, "container", mockContext.GetLocation())
	assert.True(t, mockContext.IsDockerRunning())
	assert.Equal(t, []string{"docker-compose.yml"}, mockContext.GetComposeFiles())
	assert.False(t, mockContext.IsWorktree())
	assert.Equal(t, "", mockContext.GetWorktreeName())
	mockContext.AssertExpectations(t)
}

func TestMockProjectContext_Worktree(t *testing.T) {
	mockContext := new(MockProjectContext)

	// Set up expectations for worktree scenario
	mockContext.On("GetWorkingDir").Return("/test/worktrees/feature-branch")
	mockContext.On("GetProjectRoot").Return("/test")
	mockContext.On("IsWorktree").Return(true)
	mockContext.On("GetWorktreeName").Return("feature-branch")

	// Execute and verify
	assert.Equal(t, "/test/worktrees/feature-branch", mockContext.GetWorkingDir())
	assert.Equal(t, "/test", mockContext.GetProjectRoot())
	assert.True(t, mockContext.IsWorktree())
	assert.Equal(t, "feature-branch", mockContext.GetWorktreeName())
	mockContext.AssertExpectations(t)
}

func TestExpectContextDetection(t *testing.T) {
	mockDetector := new(MockContextDetector)
	mockContext := new(MockProjectContext)
	workingDir := "/test/dir"

	// Use helper
	ExpectContextDetection(mockDetector, workingDir, mockContext, nil)

	// Execute
	ctx, err := mockDetector.Detect(workingDir)

	// Verify
	assert.NoError(t, err)
	assert.Equal(t, mockContext, ctx)
	mockDetector.AssertExpectations(t)
}

func TestExpectContextDetectionWithRoot(t *testing.T) {
	mockDetector := new(MockContextDetector)
	mockContext := new(MockProjectContext)
	workingDir := "/test/dir"
	projectRoot := "/test"

	// Use helper
	ExpectContextDetectionWithRoot(mockDetector, workingDir, projectRoot, mockContext, nil)

	// Execute
	ctx, err := mockDetector.DetectWithRoot(workingDir, projectRoot)

	// Verify
	assert.NoError(t, err)
	assert.Equal(t, mockContext, ctx)
	mockDetector.AssertExpectations(t)
}

func TestExpectContextDetectionError(t *testing.T) {
	mockDetector := new(MockContextDetector)
	workingDir := "/test/dir"
	expectedError := errors.New("detection failed")

	// Use helper
	ExpectContextDetection(mockDetector, workingDir, nil, expectedError)

	// Execute
	ctx, err := mockDetector.Detect(workingDir)

	// Verify
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, ctx)
	mockDetector.AssertExpectations(t)
}
