package mocks

import (
	"io"

	"github.com/stretchr/testify/mock"
)

// OutputManager is a mock implementation of interfaces.OutputManager
type OutputManager struct {
	mock.Mock
}

// Display mocks the Display method
func (m *OutputManager) Display(data interface{}) error {
	args := m.Called(data)
	return args.Error(0)
}

// Info mocks the Info method
func (m *OutputManager) Info(format string, args ...interface{}) error {
	mockArgs := m.Called(format, args)
	return mockArgs.Error(0)
}

// Success mocks the Success method
func (m *OutputManager) Success(format string, args ...interface{}) error {
	mockArgs := m.Called(format, args)
	return mockArgs.Error(0)
}

// Error mocks the Error method
func (m *OutputManager) Error(format string, args ...interface{}) error {
	mockArgs := m.Called(format, args)
	return mockArgs.Error(0)
}

// Warning mocks the Warning method
func (m *OutputManager) Warning(format string, args ...interface{}) error {
	mockArgs := m.Called(format, args)
	return mockArgs.Error(0)
}

// Raw mocks the Raw method
func (m *OutputManager) Raw(text string) error {
	args := m.Called(text)
	return args.Error(0)
}

// Printf mocks the Printf method
func (m *OutputManager) Printf(format string, args ...interface{}) error {
	mockArgs := m.Called(format, args)
	return mockArgs.Error(0)
}

// Println mocks the Println method
func (m *OutputManager) Println(args ...interface{}) error {
	mockArgs := m.Called(args)
	return mockArgs.Error(0)
}

// Formatter is a mock implementation of interfaces.Formatter
type Formatter struct {
	mock.Mock
}

// Display mocks the Display method
func (m *Formatter) Display(data interface{}) error {
	args := m.Called(data)
	return args.Error(0)
}

// Info mocks the Info method
func (m *Formatter) Info(format string, args ...interface{}) error {
	mockArgs := m.Called(format, args)
	return mockArgs.Error(0)
}

// Success mocks the Success method
func (m *Formatter) Success(format string, args ...interface{}) error {
	mockArgs := m.Called(format, args)
	return mockArgs.Error(0)
}

// Error mocks the Error method
func (m *Formatter) Error(format string, args ...interface{}) error {
	mockArgs := m.Called(format, args)
	return mockArgs.Error(0)
}

// Warning mocks the Warning method
func (m *Formatter) Warning(format string, args ...interface{}) error {
	mockArgs := m.Called(format, args)
	return mockArgs.Error(0)
}

// Raw mocks the Raw method
func (m *Formatter) Raw(text string) error {
	args := m.Called(text)
	return args.Error(0)
}

// SetWriter mocks the SetWriter method
func (m *Formatter) SetWriter(w io.Writer) {
	m.Called(w)
}
