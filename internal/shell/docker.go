package shell

import (
	"os"
	"strings"

	"github.com/ivannovak/glide/internal/context"
)

// DockerExecutor specializes in Docker and docker-compose commands
type DockerExecutor struct {
	executor *Executor
	ctx      *context.ProjectContext
}

// NewDockerExecutor creates a new Docker command executor
func NewDockerExecutor(ctx *context.ProjectContext) *DockerExecutor {
	return &DockerExecutor{
		executor: NewExecutor(Options{
			Verbose: false,
		}),
		ctx: ctx,
	}
}

// Compose runs a docker-compose command with proper file resolution
func (de *DockerExecutor) Compose(args ...string) error {
	// Build docker-compose command with resolved files
	composeArgs := []string{"compose"}

	// Add compose files
	for _, file := range de.ctx.ComposeFiles {
		composeArgs = append(composeArgs, "-f", file)
	}

	// Add user arguments
	composeArgs = append(composeArgs, args...)

	// Execute with passthrough
	return de.executor.Run("docker", composeArgs...)
}

// ComposeCapture runs docker-compose and captures output
func (de *DockerExecutor) ComposeCapture(args ...string) (string, error) {
	// Build docker-compose command with resolved files
	composeArgs := []string{"compose"}

	// Add compose files
	for _, file := range de.ctx.ComposeFiles {
		composeArgs = append(composeArgs, "-f", file)
	}

	// Add user arguments
	composeArgs = append(composeArgs, args...)

	// Execute and capture
	return de.executor.RunCapture("docker", composeArgs...)
}

// Up starts Docker containers
func (de *DockerExecutor) Up(detach bool, services ...string) error {
	args := []string{"up"}
	if detach {
		args = append(args, "-d")
	}
	args = append(args, services...)

	return de.Compose(args...)
}

// Down stops Docker containers
func (de *DockerExecutor) Down(removeVolumes bool, removeOrphans bool) error {
	args := []string{"down"}
	if removeVolumes {
		args = append(args, "-v")
	}
	if removeOrphans {
		args = append(args, "--remove-orphans")
	}

	return de.Compose(args...)
}

// Exec executes a command in a container
func (de *DockerExecutor) Exec(service string, command []string, interactive bool) error {
	args := []string{"exec"}
	if interactive {
		args = append(args, "-it")
	}
	args = append(args, service)
	args = append(args, command...)

	// Use interactive mode if -it flag is present
	if interactive {
		cmd := NewInteractiveCommand("docker", append([]string{"compose"}, args...)...)
		for _, file := range de.ctx.ComposeFiles {
			cmd.Args = append([]string{"-f", file}, cmd.Args...)
		}
		_, err := de.executor.Execute(cmd)
		return err
	}

	return de.Compose(args...)
}

// Logs shows container logs
func (de *DockerExecutor) Logs(follow bool, tail string, services ...string) error {
	args := []string{"logs"}
	if follow {
		args = append(args, "-f")
	}
	if tail != "" {
		args = append(args, "--tail", tail)
	}
	args = append(args, services...)

	return de.Compose(args...)
}

// PS shows container status
func (de *DockerExecutor) PS(all bool) (string, error) {
	args := []string{"ps"}
	if all {
		args = append(args, "-a")
	}

	return de.ComposeCapture(args...)
}

// IsRunning checks if Docker daemon is running
func (de *DockerExecutor) IsRunning() bool {
	cmd := NewCommand("docker", "info")
	cmd.Mode = ModeCapture
	result, _ := de.executor.Execute(cmd)
	return result != nil && result.ExitCode == 0
}

// GetContainerStatus returns the status of containers
func (de *DockerExecutor) GetContainerStatus() (map[string]string, error) {
	output, err := de.ComposeCapture("ps", "--format", "table {{.Name}}\t{{.Status}}")
	if err != nil {
		return nil, err
	}

	status := make(map[string]string)
	lines := strings.Split(output, "\n")

	// Skip header
	if len(lines) > 1 {
		for _, line := range lines[1:] {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				name := parts[0]
				statusStr := strings.Join(parts[1:], " ")
				status[name] = statusStr
			}
		}
	}

	return status, nil
}

// RunInContainer runs a command inside a specific container
func (de *DockerExecutor) RunInContainer(service string, command string, args ...string) error {
	execArgs := []string{service, command}
	execArgs = append(execArgs, args...)
	return de.Exec(service, execArgs, false)
}

// Shell opens an interactive shell in a container
func (de *DockerExecutor) Shell(service string) error {
	// Try bash first, fall back to sh
	err := de.Exec(service, []string{"bash"}, true)
	if err != nil {
		// Try sh as fallback
		return de.Exec(service, []string{"sh"}, true)
	}
	return nil
}

// PassthroughToCompose passes all arguments directly to docker-compose
func (de *DockerExecutor) PassthroughToCompose(args []string) error {
	// This is for the 'glid docker' command - complete passthrough
	composeArgs := []string{"compose"}

	// Add compose files
	for _, file := range de.ctx.ComposeFiles {
		composeArgs = append(composeArgs, "-f", file)
	}

	// Add all user arguments without interpretation
	composeArgs = append(composeArgs, args...)

	// Create passthrough command
	cmd := NewPassthroughCommand("docker", composeArgs...)
	cmd.SignalForward = true
	cmd.InheritEnv = true

	// Check if this looks like an interactive command
	for _, arg := range args {
		if arg == "exec" {
			// Look for -it flags
			for _, a := range args {
				if a == "-it" || a == "-i" || a == "-t" {
					cmd.Mode = ModeInteractive
					cmd.AllocateTTY = true
					break
				}
			}
			break
		}
	}

	result, err := de.executor.Execute(cmd)
	if err != nil {
		return err
	}

	// Exit with same code as docker-compose
	if result.ExitCode != 0 {
		os.Exit(result.ExitCode)
	}

	return nil
}
