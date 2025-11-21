// Package docker provides deprecated Docker functionality.
// DEPRECATED: This package exists only for backward compatibility.
// All Docker functionality has moved to plugins/docker.
// Use: glide docker [command]
package docker

import "github.com/ivannovak/glide/internal/context"

// DEPRECATED: Use plugins/docker instead
type Resolver struct{}

func NewResolver(ctx *context.ProjectContext) *Resolver       { return &Resolver{} }
func (r *Resolver) Resolve() error                            { return nil }
func (r *Resolver) GetComposeCommand(args ...string) []string { return nil }
func (r *Resolver) GetRelativeComposeFiles() []string         { return nil }
func (r *Resolver) GetComposeProjectName() string             { return "" }
func (r *Resolver) GetDockerNetwork() string                  { return "" }
func (r *Resolver) GetComposeFiles() []string                 { return nil }
func (r *Resolver) GetOverrideFile() string                   { return "" }
func (r *Resolver) ValidateSetup() error                      { return nil }

// Container represents a Docker container (deprecated)
type Container struct {
	Name    string
	Service string
	State   string
	Status  string
}

// ServiceHealth represents service health status (deprecated)
type ServiceHealth struct {
	Service string
	Healthy bool
	Summary string
}

type ContainerManager struct{}

func NewContainerManager(ctx *context.ProjectContext) *ContainerManager  { return &ContainerManager{} }
func (cm *ContainerManager) GetStatus() ([]Container, error)             { return nil, nil }
func (cm *ContainerManager) GetComposeServices() ([]string, error)       { return nil, nil }
func (cm *ContainerManager) GetOrphanedContainers() ([]Container, error) { return nil, nil }
func (cm *ContainerManager) IsRunning(service string) bool               { return false }

type HealthMonitor struct{}

func NewHealthMonitor(ctx *context.ProjectContext) *HealthMonitor { return &HealthMonitor{} }
func (hm *HealthMonitor) CheckHealth() ([]ServiceHealth, error)   { return nil, nil }

type ErrorHandler struct{}

func NewErrorHandler(verbose bool) *ErrorHandler       { return &ErrorHandler{} }
func (eh *ErrorHandler) Handle(err error) string       { return err.Error() }
func (eh *ErrorHandler) SuggestFix(err error) []string { return nil }

func ParseDockerError(op string, output string, err error) error { return err }
func IsRetryable(err error) bool                                 { return false }
