package mocks

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ivannovak/glide/v2/pkg/interfaces"
	"github.com/stretchr/testify/assert"
)

func TestMockShellExecutor_Execute(t *testing.T) {
	mockExec := new(MockShellExecutor)
	mockCmd := new(MockShellCommand)
	ctx := context.Background()

	expectedResult := &interfaces.ShellResult{
		Stdout:   "output",
		Stderr:   "",
		ExitCode: 0,
		Duration: 100 * time.Millisecond,
	}

	// Set up expectation
	mockExec.On("Execute", ctx, mockCmd).Return(expectedResult, nil)

	// Execute
	result, err := mockExec.Execute(ctx, mockCmd)

	// Verify
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	mockExec.AssertExpectations(t)
}

func TestMockShellExecutor_ExecuteWithTimeout(t *testing.T) {
	mockExec := new(MockShellExecutor)
	mockCmd := new(MockShellCommand)
	timeout := 5 * time.Second

	expectedResult := &interfaces.ShellResult{
		Stdout:   "output",
		Stderr:   "",
		ExitCode: 0,
		Duration: 2 * time.Second,
	}

	// Set up expectation
	mockExec.On("ExecuteWithTimeout", mockCmd, timeout).Return(expectedResult, nil)

	// Execute
	result, err := mockExec.ExecuteWithTimeout(mockCmd, timeout)

	// Verify
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	mockExec.AssertExpectations(t)
}

func TestMockShellExecutor_ExecuteWithProgress(t *testing.T) {
	mockExec := new(MockShellExecutor)
	mockCmd := new(MockShellCommand)
	message := "Running command..."

	// Set up expectation
	mockExec.On("ExecuteWithProgress", mockCmd, message).Return(nil)

	// Execute
	err := mockExec.ExecuteWithProgress(mockCmd, message)

	// Verify
	assert.NoError(t, err)
	mockExec.AssertExpectations(t)
}

func TestMockShellExecutor_ExecuteError(t *testing.T) {
	mockExec := new(MockShellExecutor)
	mockCmd := new(MockShellCommand)
	ctx := context.Background()
	expectedError := errors.New("execution failed")

	// Set up expectation
	mockExec.On("Execute", ctx, mockCmd).Return(nil, expectedError)

	// Execute
	result, err := mockExec.Execute(ctx, mockCmd)

	// Verify
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, result)
	mockExec.AssertExpectations(t)
}

func TestExpectCommandExecution(t *testing.T) {
	mockExec := new(MockShellExecutor)
	mockCmd := new(MockShellCommand)

	expectedResult := &interfaces.ShellResult{
		Stdout:   "test output",
		Stderr:   "",
		ExitCode: 0,
		Duration: 50 * time.Millisecond,
	}

	// Use helper
	ExpectCommandExecution(mockExec, mockCmd, expectedResult, nil)

	// Execute
	result, err := mockExec.Execute(context.Background(), mockCmd)

	// Verify
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	mockExec.AssertExpectations(t)
}

func TestExpectCommandExecutionWithTimeout(t *testing.T) {
	mockExec := new(MockShellExecutor)
	mockCmd := new(MockShellCommand)
	timeout := 3 * time.Second

	expectedResult := &interfaces.ShellResult{
		Stdout:   "test output",
		Stderr:   "",
		ExitCode: 0,
		Duration: 1 * time.Second,
	}

	// Use helper
	ExpectCommandExecutionWithTimeout(mockExec, mockCmd, timeout, expectedResult, nil)

	// Execute
	result, err := mockExec.ExecuteWithTimeout(mockCmd, timeout)

	// Verify
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	mockExec.AssertExpectations(t)
}

func TestExpectCommandExecutionWithProgress(t *testing.T) {
	mockExec := new(MockShellExecutor)
	mockCmd := new(MockShellCommand)
	message := "Processing..."

	// Use helper
	ExpectCommandExecutionWithProgress(mockExec, mockCmd, message, nil)

	// Execute
	err := mockExec.ExecuteWithProgress(mockCmd, message)

	// Verify
	assert.NoError(t, err)
	mockExec.AssertExpectations(t)
}

func TestMockShellCommand(t *testing.T) {
	mockCmd := new(MockShellCommand)

	// Set up expectations
	mockCmd.On("GetCommand").Return("test-command")
	mockCmd.On("GetArgs").Return([]string{"arg1", "arg2"})
	mockCmd.On("GetWorkingDir").Return("/tmp/test")
	mockCmd.On("GetEnvironment").Return(map[string]string{"KEY": "value"})

	// Execute and verify
	assert.Equal(t, "test-command", mockCmd.GetCommand())
	assert.Equal(t, []string{"arg1", "arg2"}, mockCmd.GetArgs())
	assert.Equal(t, "/tmp/test", mockCmd.GetWorkingDir())
	assert.Equal(t, map[string]string{"KEY": "value"}, mockCmd.GetEnvironment())
	mockCmd.AssertExpectations(t)
}
