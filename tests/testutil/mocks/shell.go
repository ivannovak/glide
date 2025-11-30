package mocks

import (
	"context"
	"time"

	"github.com/ivannovak/glide/v2/pkg/interfaces"
	"github.com/stretchr/testify/mock"
)

// MockShellExecutor is a mock implementation of the ShellExecutor interface
type MockShellExecutor struct {
	mock.Mock
}

// Execute mocks the Execute method
func (m *MockShellExecutor) Execute(ctx context.Context, cmd interfaces.ShellCommand) (*interfaces.ShellResult, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*interfaces.ShellResult), args.Error(1)
}

// ExecuteWithTimeout mocks the ExecuteWithTimeout method
func (m *MockShellExecutor) ExecuteWithTimeout(cmd interfaces.ShellCommand, timeout time.Duration) (*interfaces.ShellResult, error) {
	args := m.Called(cmd, timeout)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*interfaces.ShellResult), args.Error(1)
}

// ExecuteWithProgress mocks the ExecuteWithProgress method
func (m *MockShellExecutor) ExecuteWithProgress(cmd interfaces.ShellCommand, message string) error {
	args := m.Called(cmd, message)
	return args.Error(0)
}

// MockShellCommand is a mock implementation of the ShellCommand interface
type MockShellCommand struct {
	mock.Mock
}

// GetCommand mocks the GetCommand method
func (m *MockShellCommand) GetCommand() string {
	args := m.Called()
	return args.String(0)
}

// GetArgs mocks the GetArgs method
func (m *MockShellCommand) GetArgs() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]string)
}

// GetOptions mocks the GetOptions method
func (m *MockShellCommand) GetOptions() interfaces.ShellCommandOptions {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(interfaces.ShellCommandOptions)
}

// GetWorkingDir mocks the GetWorkingDir method
func (m *MockShellCommand) GetWorkingDir() string {
	args := m.Called()
	return args.String(0)
}

// GetEnvironment mocks the GetEnvironment method
func (m *MockShellCommand) GetEnvironment() map[string]string {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(map[string]string)
}

// ExpectCommandExecution is a helper to set up expected command execution
func ExpectCommandExecution(m *MockShellExecutor, cmd interfaces.ShellCommand, result *interfaces.ShellResult, err error) *mock.Call {
	return m.On("Execute", mock.Anything, cmd).Return(result, err)
}

// ExpectCommandExecutionWithTimeout is a helper to set up expected command execution with timeout
func ExpectCommandExecutionWithTimeout(m *MockShellExecutor, cmd interfaces.ShellCommand, timeout time.Duration, result *interfaces.ShellResult, err error) *mock.Call {
	return m.On("ExecuteWithTimeout", cmd, timeout).Return(result, err)
}

// ExpectCommandExecutionWithProgress is a helper to set up expected command execution with progress
func ExpectCommandExecutionWithProgress(m *MockShellExecutor, cmd interfaces.ShellCommand, message string, err error) *mock.Call {
	return m.On("ExecuteWithProgress", cmd, message).Return(err)
}
