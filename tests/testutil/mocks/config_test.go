package mocks

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockConfigLoader_Load(t *testing.T) {
	mockLoader := new(MockConfigLoader)
	path := "/test/config.yaml"

	// Set up expectation
	mockLoader.On("Load", path).Return(nil)

	// Execute
	err := mockLoader.Load(path)

	// Verify
	assert.NoError(t, err)
	mockLoader.AssertExpectations(t)
}

func TestMockConfigLoader_LoadError(t *testing.T) {
	mockLoader := new(MockConfigLoader)
	path := "/test/invalid.yaml"
	expectedError := errors.New("failed to load config")

	// Set up expectation
	mockLoader.On("Load", path).Return(expectedError)

	// Execute
	err := mockLoader.Load(path)

	// Verify
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockLoader.AssertExpectations(t)
}

func TestMockConfigLoader_LoadDefault(t *testing.T) {
	mockLoader := new(MockConfigLoader)

	// Set up expectation
	mockLoader.On("LoadDefault").Return(nil)

	// Execute
	err := mockLoader.LoadDefault()

	// Verify
	assert.NoError(t, err)
	mockLoader.AssertExpectations(t)
}

func TestMockConfigLoader_GetConfig(t *testing.T) {
	mockLoader := new(MockConfigLoader)
	expectedConfig := map[string]interface{}{
		"setting1": "value1",
		"setting2": 42,
	}

	// Set up expectation
	mockLoader.On("GetConfig").Return(expectedConfig)

	// Execute
	config := mockLoader.GetConfig()

	// Verify
	assert.Equal(t, expectedConfig, config)
	mockLoader.AssertExpectations(t)
}

func TestMockConfigLoader_Save(t *testing.T) {
	mockLoader := new(MockConfigLoader)
	path := "/test/config.yaml"

	// Set up expectation
	mockLoader.On("Save", path).Return(nil)

	// Execute
	err := mockLoader.Save(path)

	// Verify
	assert.NoError(t, err)
	mockLoader.AssertExpectations(t)
}

func TestMockConfigLoader_SaveError(t *testing.T) {
	mockLoader := new(MockConfigLoader)
	path := "/test/readonly/config.yaml"
	expectedError := errors.New("failed to save config")

	// Set up expectation
	mockLoader.On("Save", path).Return(expectedError)

	// Execute
	err := mockLoader.Save(path)

	// Verify
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockLoader.AssertExpectations(t)
}

func TestExpectConfigLoad(t *testing.T) {
	mockLoader := new(MockConfigLoader)
	path := "/test/config.yaml"

	// Use helper
	ExpectConfigLoad(mockLoader, path, nil)

	// Execute
	err := mockLoader.Load(path)

	// Verify
	assert.NoError(t, err)
	mockLoader.AssertExpectations(t)
}

func TestExpectConfigLoadDefault(t *testing.T) {
	mockLoader := new(MockConfigLoader)

	// Use helper
	ExpectConfigLoadDefault(mockLoader, nil)

	// Execute
	err := mockLoader.LoadDefault()

	// Verify
	assert.NoError(t, err)
	mockLoader.AssertExpectations(t)
}

func TestExpectConfigGet(t *testing.T) {
	mockLoader := new(MockConfigLoader)
	config := map[string]string{"key": "value"}

	// Use helper
	ExpectConfigGet(mockLoader, config)

	// Execute
	result := mockLoader.GetConfig()

	// Verify
	assert.Equal(t, config, result)
	mockLoader.AssertExpectations(t)
}

func TestExpectConfigSave(t *testing.T) {
	mockLoader := new(MockConfigLoader)
	path := "/test/config.yaml"
	expectedError := errors.New("save failed")

	// Use helper
	ExpectConfigSave(mockLoader, path, expectedError)

	// Execute
	err := mockLoader.Save(path)

	// Verify
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockLoader.AssertExpectations(t)
}

func TestMockConfigLoader_CompleteWorkflow(t *testing.T) {
	mockLoader := new(MockConfigLoader)
	path := "/test/config.yaml"
	config := map[string]interface{}{"key": "value"}

	// Set up expectations for a complete workflow
	mockLoader.On("Load", path).Return(nil)
	mockLoader.On("GetConfig").Return(config)
	mockLoader.On("Save", path).Return(nil)

	// Execute workflow
	err := mockLoader.Load(path)
	assert.NoError(t, err)

	loadedConfig := mockLoader.GetConfig()
	assert.Equal(t, config, loadedConfig)

	err = mockLoader.Save(path)
	assert.NoError(t, err)

	mockLoader.AssertExpectations(t)
}
