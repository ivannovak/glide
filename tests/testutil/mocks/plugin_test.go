package mocks

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockRegistry_Register(t *testing.T) {
	mockRegistry := new(MockRegistry)
	key := "test-plugin"
	value := "plugin-instance"

	// Set up expectation
	mockRegistry.On("Register", key, value).Return(nil)

	// Execute
	err := mockRegistry.Register(key, value)

	// Verify
	assert.NoError(t, err)
	mockRegistry.AssertExpectations(t)
}

func TestMockRegistry_Get(t *testing.T) {
	mockRegistry := new(MockRegistry)
	key := "test-plugin"
	expectedValue := "plugin-instance"

	// Set up expectation
	mockRegistry.On("Get", key).Return(expectedValue, true)

	// Execute
	value, found := mockRegistry.Get(key)

	// Verify
	assert.True(t, found)
	assert.Equal(t, expectedValue, value)
	mockRegistry.AssertExpectations(t)
}

func TestMockRegistry_GetNotFound(t *testing.T) {
	mockRegistry := new(MockRegistry)
	key := "nonexistent"

	// Set up expectation
	mockRegistry.On("Get", key).Return(nil, false)

	// Execute
	value, found := mockRegistry.Get(key)

	// Verify
	assert.False(t, found)
	assert.Nil(t, value)
	mockRegistry.AssertExpectations(t)
}

func TestMockRegistry_List(t *testing.T) {
	mockRegistry := new(MockRegistry)
	expectedList := []string{"plugin1", "plugin2", "plugin3"}

	// Set up expectation
	mockRegistry.On("List").Return(expectedList)

	// Execute
	list := mockRegistry.List()

	// Verify
	assert.Equal(t, expectedList, list)
	mockRegistry.AssertExpectations(t)
}

func TestMockRegistry_Remove(t *testing.T) {
	mockRegistry := new(MockRegistry)
	key := "test-plugin"

	// Set up expectation
	mockRegistry.On("Remove", key).Return(true)

	// Execute
	removed := mockRegistry.Remove(key)

	// Verify
	assert.True(t, removed)
	mockRegistry.AssertExpectations(t)
}

func TestExpectPluginLoad(t *testing.T) {
	mockRegistry := new(MockRegistry)
	name := "test-plugin"
	plugin := "plugin-instance"

	// Use helper
	ExpectPluginLoad(mockRegistry, name, plugin)

	// Execute
	result, found := mockRegistry.Get(name)

	// Verify
	assert.True(t, found)
	assert.Equal(t, plugin, result)
	mockRegistry.AssertExpectations(t)
}

func TestExpectPluginNotFound(t *testing.T) {
	mockRegistry := new(MockRegistry)
	name := "nonexistent"

	// Use helper
	ExpectPluginNotFound(mockRegistry, name)

	// Execute
	result, found := mockRegistry.Get(name)

	// Verify
	assert.False(t, found)
	assert.Nil(t, result)
	mockRegistry.AssertExpectations(t)
}

func TestExpectPluginRegister(t *testing.T) {
	mockRegistry := new(MockRegistry)
	name := "test-plugin"
	plugin := "plugin-instance"
	expectedError := errors.New("registration failed")

	// Use helper
	ExpectPluginRegister(mockRegistry, name, plugin, expectedError)

	// Execute
	err := mockRegistry.Register(name, plugin)

	// Verify
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockRegistry.AssertExpectations(t)
}

func TestExpectPluginList(t *testing.T) {
	mockRegistry := new(MockRegistry)
	plugins := []string{"plugin1", "plugin2"}

	// Use helper
	ExpectPluginList(mockRegistry, plugins)

	// Execute
	list := mockRegistry.List()

	// Verify
	assert.Equal(t, plugins, list)
	mockRegistry.AssertExpectations(t)
}
