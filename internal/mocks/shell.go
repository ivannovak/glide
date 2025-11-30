package mocks

import (
	"context"
	"time"

	"github.com/ivannovak/glide/v3/pkg/interfaces"
	"github.com/stretchr/testify/mock"
)

// ShellExecutor is a mock implementation of interfaces.ShellExecutor
type ShellExecutor struct {
	mock.Mock
}

// Execute mocks the Execute method
func (m *ShellExecutor) Execute(ctx context.Context, cmd interfaces.ShellCommand) (*interfaces.ShellResult, error) {
	args := m.Called(ctx, cmd)
	if result := args.Get(0); result != nil {
		return result.(*interfaces.ShellResult), args.Error(1)
	}
	return nil, args.Error(1)
}

// ExecuteWithTimeout mocks the ExecuteWithTimeout method
func (m *ShellExecutor) ExecuteWithTimeout(cmd interfaces.ShellCommand, timeout time.Duration) (*interfaces.ShellResult, error) {
	args := m.Called(cmd, timeout)
	if result := args.Get(0); result != nil {
		return result.(*interfaces.ShellResult), args.Error(1)
	}
	return nil, args.Error(1)
}

// ExecuteWithProgress mocks the ExecuteWithProgress method
func (m *ShellExecutor) ExecuteWithProgress(cmd interfaces.ShellCommand, message string) error {
	args := m.Called(cmd, message)
	return args.Error(0)
}

// ShellCommand is a mock implementation of interfaces.ShellCommand
type ShellCommand struct {
	mock.Mock
}

// GetCommand returns the mocked command
func (m *ShellCommand) GetCommand() string {
	args := m.Called()
	return args.String(0)
}

// GetArgs returns the mocked arguments
func (m *ShellCommand) GetArgs() []string {
	args := m.Called()
	if result := args.Get(0); result != nil {
		return result.([]string)
	}
	return nil
}

// GetOptions returns the mocked options
func (m *ShellCommand) GetOptions() interfaces.ShellCommandOptions {
	args := m.Called()
	if result := args.Get(0); result != nil {
		return result.(interfaces.ShellCommandOptions)
	}
	return nil
}

// GetWorkingDir returns the mocked working directory
func (m *ShellCommand) GetWorkingDir() string {
	args := m.Called()
	return args.String(0)
}

// GetEnvironment returns the mocked environment variables
func (m *ShellCommand) GetEnvironment() map[string]string {
	args := m.Called()
	if result := args.Get(0); result != nil {
		return result.(map[string]string)
	}
	return nil
}

// ShellCommandOptions is a mock implementation of interfaces.ShellCommandOptions
type ShellCommandOptions struct {
	mock.Mock
}

// IsCaptureOutput returns whether output should be captured
func (m *ShellCommandOptions) IsCaptureOutput() bool {
	args := m.Called()
	return args.Bool(0)
}

// GetTimeout returns the timeout duration
func (m *ShellCommandOptions) GetTimeout() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

// IsStreamOutput returns whether output should be streamed
func (m *ShellCommandOptions) IsStreamOutput() bool {
	args := m.Called()
	return args.Bool(0)
}

// GetOutputWriter returns the output writer
func (m *ShellCommandOptions) GetOutputWriter() interface{} {
	args := m.Called()
	return args.Get(0)
}

// GetErrorWriter returns the error writer
func (m *ShellCommandOptions) GetErrorWriter() interface{} {
	args := m.Called()
	return args.Get(0)
}
