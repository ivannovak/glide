package mocks

import (
	"github.com/stretchr/testify/mock"
)

// MockRegistry is a mock implementation of the Registry interface
type MockRegistry struct {
	mock.Mock
}

// Register mocks the Register method
func (m *MockRegistry) Register(key string, value interface{}) error {
	args := m.Called(key, value)
	return args.Error(0)
}

// Get mocks the Get method
func (m *MockRegistry) Get(key string) (interface{}, bool) {
	args := m.Called(key)
	return args.Get(0), args.Bool(1)
}

// List mocks the List method
func (m *MockRegistry) List() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]string)
}

// Remove mocks the Remove method
func (m *MockRegistry) Remove(key string) bool {
	args := m.Called(key)
	return args.Bool(0)
}

// ExpectPluginLoad is a helper to set up expected plugin loading
func ExpectPluginLoad(m *MockRegistry, name string, plugin interface{}) *mock.Call {
	return m.On("Get", name).Return(plugin, true)
}

// ExpectPluginNotFound is a helper to set up expected plugin not found
func ExpectPluginNotFound(m *MockRegistry, name string) *mock.Call {
	return m.On("Get", name).Return(nil, false)
}

// ExpectPluginRegister is a helper to set up expected plugin registration
func ExpectPluginRegister(m *MockRegistry, name string, plugin interface{}, err error) *mock.Call {
	return m.On("Register", name, plugin).Return(err)
}

// ExpectPluginList is a helper to set up expected plugin list
func ExpectPluginList(m *MockRegistry, plugins []string) *mock.Call {
	return m.On("List").Return(plugins)
}
