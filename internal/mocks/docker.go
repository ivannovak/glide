package mocks

import (
	"github.com/glide-cli/glide/v3/pkg/interfaces"
	"github.com/stretchr/testify/mock"
)

// DockerResolver is a mock implementation of interfaces.DockerResolver
type DockerResolver struct {
	mock.Mock
}

// Resolve mocks the Resolve method
func (m *DockerResolver) Resolve() error {
	args := m.Called()
	return args.Error(0)
}

// GetComposeFiles mocks the GetComposeFiles method
func (m *DockerResolver) GetComposeFiles() []string {
	args := m.Called()
	if result := args.Get(0); result != nil {
		return result.([]string)
	}
	return nil
}

// BuildDockerCommand mocks the BuildDockerCommand method
func (m *DockerResolver) BuildDockerCommand(cmdArgs ...string) string {
	args := m.Called(cmdArgs)
	return args.String(0)
}

// GetBaseArgs mocks the GetBaseArgs method
func (m *DockerResolver) GetBaseArgs() []string {
	args := m.Called()
	if result := args.Get(0); result != nil {
		return result.([]string)
	}
	return nil
}

// ContainerManager is a mock implementation of interfaces.ContainerManager
type ContainerManager struct {
	mock.Mock
}

// Up mocks the Up method
func (m *ContainerManager) Up() error {
	args := m.Called()
	return args.Error(0)
}

// Down mocks the Down method
func (m *ContainerManager) Down() error {
	args := m.Called()
	return args.Error(0)
}

// Status mocks the Status method
func (m *ContainerManager) Status() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

// Logs mocks the Logs method
func (m *ContainerManager) Logs(service string, tail int) (string, error) {
	args := m.Called(service, tail)
	return args.String(0), args.Error(1)
}

// Shell mocks the Shell method
func (m *ContainerManager) Shell(service string) error {
	args := m.Called(service)
	return args.Error(0)
}

// ListContainers mocks the ListContainers method
func (m *ContainerManager) ListContainers() ([]interfaces.ContainerInfo, error) {
	args := m.Called()
	if result := args.Get(0); result != nil {
		return result.([]interfaces.ContainerInfo), args.Error(1)
	}
	return nil, args.Error(1)
}
