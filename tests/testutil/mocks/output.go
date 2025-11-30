package mocks

import (
	"github.com/stretchr/testify/mock"
)

// MockOutputManager is a mock implementation of the OutputManager interface
type MockOutputManager struct {
	mock.Mock
}

// Display mocks the Display method
func (m *MockOutputManager) Display(data interface{}) error {
	args := m.Called(data)
	return args.Error(0)
}

// Info mocks the Info method
func (m *MockOutputManager) Info(format string, args ...interface{}) error {
	callArgs := m.Called(format, args)
	return callArgs.Error(0)
}

// Success mocks the Success method
func (m *MockOutputManager) Success(format string, args ...interface{}) error {
	callArgs := m.Called(format, args)
	return callArgs.Error(0)
}

// Error mocks the Error method
func (m *MockOutputManager) Error(format string, args ...interface{}) error {
	callArgs := m.Called(format, args)
	return callArgs.Error(0)
}

// Warning mocks the Warning method
func (m *MockOutputManager) Warning(format string, args ...interface{}) error {
	callArgs := m.Called(format, args)
	return callArgs.Error(0)
}

// Raw mocks the Raw method
func (m *MockOutputManager) Raw(text string) error {
	args := m.Called(text)
	return args.Error(0)
}

// Printf mocks the Printf method
func (m *MockOutputManager) Printf(format string, args ...interface{}) error {
	callArgs := m.Called(format, args)
	return callArgs.Error(0)
}

// Println mocks the Println method
func (m *MockOutputManager) Println(args ...interface{}) error {
	callArgs := m.Called(args)
	return callArgs.Error(0)
}

// ExpectOutput is a helper to set up expected output at a specific level
func ExpectOutput(m *MockOutputManager, level string, message string, err error) *mock.Call {
	switch level {
	case "info":
		return m.On("Info", message, mock.Anything).Return(err)
	case "success":
		return m.On("Success", message, mock.Anything).Return(err)
	case "error":
		return m.On("Error", message, mock.Anything).Return(err)
	case "warning":
		return m.On("Warning", message, mock.Anything).Return(err)
	default:
		return m.On("Display", message).Return(err)
	}
}

// ExpectRawOutput is a helper to set up expected raw output
func ExpectRawOutput(m *MockOutputManager, text string, err error) *mock.Call {
	return m.On("Raw", text).Return(err)
}

// ExpectDisplayOutput is a helper to set up expected display output
func ExpectDisplayOutput(m *MockOutputManager, data interface{}, err error) *mock.Call {
	return m.On("Display", data).Return(err)
}
