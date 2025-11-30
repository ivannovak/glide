package mocks

import (
	"github.com/stretchr/testify/mock"
)

// MockConfigLoader is a mock implementation of the ConfigLoader interface
type MockConfigLoader struct {
	mock.Mock
}

// Load mocks the Load method
func (m *MockConfigLoader) Load(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

// LoadDefault mocks the LoadDefault method
func (m *MockConfigLoader) LoadDefault() error {
	args := m.Called()
	return args.Error(0)
}

// GetConfig mocks the GetConfig method
func (m *MockConfigLoader) GetConfig() interface{} {
	args := m.Called()
	return args.Get(0)
}

// Save mocks the Save method
func (m *MockConfigLoader) Save(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

// ExpectConfigLoad is a helper to set up expected config loading
func ExpectConfigLoad(m *MockConfigLoader, path string, err error) *mock.Call {
	return m.On("Load", path).Return(err)
}

// ExpectConfigLoadDefault is a helper to set up expected default config loading
func ExpectConfigLoadDefault(m *MockConfigLoader, err error) *mock.Call {
	return m.On("LoadDefault").Return(err)
}

// ExpectConfigGet is a helper to set up expected config retrieval
func ExpectConfigGet(m *MockConfigLoader, config interface{}) *mock.Call {
	return m.On("GetConfig").Return(config)
}

// ExpectConfigSave is a helper to set up expected config save
func ExpectConfigSave(m *MockConfigLoader, path string, err error) *mock.Call {
	return m.On("Save", path).Return(err)
}
