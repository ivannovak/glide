// DEPRECATED: Docker functionality moved to plugins/docker
package shell

import "github.com/ivannovak/glide/internal/context"

// DEPRECATED: Use plugins/docker instead
type DockerExecutor struct{}

func NewDockerExecutor(ctx *context.ProjectContext) *DockerExecutor {
	return &DockerExecutor{}
}

func (d *DockerExecutor) IsRunning() bool                                { return false }
func (d *DockerExecutor) GetContainerStatus() (map[string]string, error) { return nil, nil }
