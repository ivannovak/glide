package mocks

import (
	"github.com/stretchr/testify/mock"
)

// ConfigLoader is a mock implementation of interfaces.ConfigLoader
type ConfigLoader struct {
	mock.Mock
}

// Load mocks the Load method
func (m *ConfigLoader) Load(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

// LoadDefault mocks the LoadDefault method
func (m *ConfigLoader) LoadDefault() error {
	args := m.Called()
	return args.Error(0)
}

// GetConfig mocks the GetConfig method
func (m *ConfigLoader) GetConfig() interface{} {
	args := m.Called()
	return args.Get(0)
}

// Save mocks the Save method
func (m *ConfigLoader) Save(path string) error {
	args := m.Called(path)
	return args.Error(0)
}