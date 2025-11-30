package mocks

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMockOutputManager_Display(t *testing.T) {
	mockOutput := new(MockOutputManager)
	data := map[string]string{"key": "value"}

	// Set up expectation
	mockOutput.On("Display", data).Return(nil)

	// Execute
	err := mockOutput.Display(data)

	// Verify
	assert.NoError(t, err)
	mockOutput.AssertExpectations(t)
}

func TestMockOutputManager_Info(t *testing.T) {
	mockOutput := new(MockOutputManager)
	format := "Info message: %s"
	args := []interface{}{"test"}

	// Set up expectation
	mockOutput.On("Info", format, args).Return(nil)

	// Execute
	err := mockOutput.Info(format, args...)

	// Verify
	assert.NoError(t, err)
	mockOutput.AssertExpectations(t)
}

func TestMockOutputManager_Success(t *testing.T) {
	mockOutput := new(MockOutputManager)
	format := "Success: %s"
	args := []interface{}{"completed"}

	// Set up expectation
	mockOutput.On("Success", format, args).Return(nil)

	// Execute
	err := mockOutput.Success(format, args...)

	// Verify
	assert.NoError(t, err)
	mockOutput.AssertExpectations(t)
}

func TestMockOutputManager_Error(t *testing.T) {
	mockOutput := new(MockOutputManager)
	format := "Error: %s"
	args := []interface{}{"failed"}

	// Set up expectation
	mockOutput.On("Error", format, args).Return(nil)

	// Execute
	err := mockOutput.Error(format, args...)

	// Verify
	assert.NoError(t, err)
	mockOutput.AssertExpectations(t)
}

func TestMockOutputManager_Warning(t *testing.T) {
	mockOutput := new(MockOutputManager)
	format := "Warning: %s"
	args := []interface{}{"caution"}

	// Set up expectation
	mockOutput.On("Warning", format, args).Return(nil)

	// Execute
	err := mockOutput.Warning(format, args...)

	// Verify
	assert.NoError(t, err)
	mockOutput.AssertExpectations(t)
}

func TestMockOutputManager_Raw(t *testing.T) {
	mockOutput := new(MockOutputManager)
	text := "raw output text"

	// Set up expectation
	mockOutput.On("Raw", text).Return(nil)

	// Execute
	err := mockOutput.Raw(text)

	// Verify
	assert.NoError(t, err)
	mockOutput.AssertExpectations(t)
}

func TestMockOutputManager_Printf(t *testing.T) {
	mockOutput := new(MockOutputManager)
	format := "Printf: %d"
	args := []interface{}{42}

	// Set up expectation
	mockOutput.On("Printf", format, args).Return(nil)

	// Execute
	err := mockOutput.Printf(format, args...)

	// Verify
	assert.NoError(t, err)
	mockOutput.AssertExpectations(t)
}

func TestMockOutputManager_Println(t *testing.T) {
	mockOutput := new(MockOutputManager)
	args := []interface{}{"line1", "line2"}

	// Set up expectation
	mockOutput.On("Println", args).Return(nil)

	// Execute
	err := mockOutput.Println(args...)

	// Verify
	assert.NoError(t, err)
	mockOutput.AssertExpectations(t)
}

func TestExpectOutput_Info(t *testing.T) {
	mockOutput := new(MockOutputManager)
	message := "test message"

	// Use helper
	ExpectOutput(mockOutput, "info", message, nil)

	// Execute
	err := mockOutput.Info(message, mock.Anything)

	// Verify
	assert.NoError(t, err)
	mockOutput.AssertExpectations(t)
}

func TestExpectOutput_Success(t *testing.T) {
	mockOutput := new(MockOutputManager)
	message := "success message"

	// Use helper
	ExpectOutput(mockOutput, "success", message, nil)

	// Execute
	err := mockOutput.Success(message, mock.Anything)

	// Verify
	assert.NoError(t, err)
	mockOutput.AssertExpectations(t)
}

func TestExpectOutput_Error(t *testing.T) {
	mockOutput := new(MockOutputManager)
	message := "error message"
	expectedError := errors.New("output error")

	// Use helper
	ExpectOutput(mockOutput, "error", message, expectedError)

	// Execute
	err := mockOutput.Error(message, mock.Anything)

	// Verify
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockOutput.AssertExpectations(t)
}

func TestExpectRawOutput(t *testing.T) {
	mockOutput := new(MockOutputManager)
	text := "raw text"

	// Use helper
	ExpectRawOutput(mockOutput, text, nil)

	// Execute
	err := mockOutput.Raw(text)

	// Verify
	assert.NoError(t, err)
	mockOutput.AssertExpectations(t)
}

func TestExpectDisplayOutput(t *testing.T) {
	mockOutput := new(MockOutputManager)
	data := map[string]interface{}{"key": "value"}

	// Use helper
	ExpectDisplayOutput(mockOutput, data, nil)

	// Execute
	err := mockOutput.Display(data)

	// Verify
	assert.NoError(t, err)
	mockOutput.AssertExpectations(t)
}
