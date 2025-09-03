package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ivannovak/glide/internal/config"
	"github.com/ivannovak/glide/internal/context"
	"github.com/ivannovak/glide/internal/docker"
	glideErrors "github.com/ivannovak/glide/pkg/errors"
	"github.com/ivannovak/glide/pkg/output"
	"github.com/spf13/cobra"
)

// ShellCommand handles the shell command
type ShellCommand struct {
	ctx *context.ProjectContext
	cfg *config.Config
}

// NewShellCommand creates a new shell command
func NewShellCommand(ctx *context.ProjectContext, cfg *config.Config) *cobra.Command {
	sc := &ShellCommand{
		ctx: ctx,
		cfg: cfg,
	}

	cmd := &cobra.Command{
		Use:   "shell [service] [command]",
		Short: "Attach to a container shell",
		Long: `Open an interactive shell session in a Docker container.

By default, this command attaches to the PHP container with bash.
You can specify a different service or shell command if needed.

Examples:
  glid shell                    # Enter PHP container with bash
  glid shell php                # Explicitly specify PHP container
  glid shell nginx              # Enter nginx container
  glid shell mysql              # Enter MySQL container
  glid shell php sh             # Use sh instead of bash
  glid shell php "ls -la"       # Run a command and exit

Available Services:
  The available services depend on your docker-compose.yml file.
  Common services include:
  - php (default)
  - nginx
  - mysql
  - redis
  - node

Shell Selection:
  The command tries these shells in order:
  1. bash (if available)
  2. sh (fallback)
  
User Context:
  By default, you enter as the container's default user.
  Use 'docker exec -u' directly if you need a specific user.

Tips:
  - Use Ctrl+D or 'exit' to leave the shell
  - The working directory is the project root
  - All project files are mounted in the container`,
		Args:          cobra.MaximumNArgs(2),
		RunE:          sc.Execute,
		SilenceUsage:  true,  // Don't show usage on error
		SilenceErrors: true,  // Let our error handler handle errors
	}

	return cmd
}

// Execute runs the shell command
func (c *ShellCommand) Execute(cmd *cobra.Command, args []string) error {
	// Check if we're in a valid project
	if c.ctx.ProjectRoot == "" {
		return glideErrors.NewConfigError("not in a project directory",
			glideErrors.WithSuggestions(
				"Navigate to a project directory",
				"Run: glid setup to initialize a new project",
				"Check if you're in the correct directory",
			),
		)
	}

	// Check if Docker is running
	if !c.ctx.DockerRunning {
		return glideErrors.NewDockerError("Docker containers are not running",
			glideErrors.WithSuggestions(
				"Start containers: glid up",
				"Check Docker daemon: docker ps",
				"Verify Docker Desktop is running",
			),
		)
	}

	// Resolve Docker compose files
	resolver := docker.NewResolver(c.ctx)
	if err := resolver.Resolve(); err != nil {
		return glideErrors.Wrap(err, "failed to resolve Docker configuration",
			glideErrors.WithSuggestions(
				"Check if docker-compose.yml exists",
				"Verify you're in the correct project directory",
				"Ensure Docker is installed: docker --version",
			),
		)
	}

	// Determine service and command
	service := "php" // default service
	shellCmd := ""   // will be determined automatically

	if len(args) > 0 {
		service = args[0]
	}
	if len(args) > 1 {
		shellCmd = args[1]
	}

	// Check if the service is running
	if err := c.checkServiceRunning(service); err != nil {
		return err
	}

	// If no shell command specified, detect the best shell
	if shellCmd == "" {
		shellCmd = c.detectShell(service)
	}

	// Build and execute the docker exec command
	return c.executeShell(resolver, service, shellCmd)
}

// checkServiceRunning verifies that the service container is running
func (c *ShellCommand) checkServiceRunning(service string) error {
	manager := docker.NewContainerManager(c.ctx)
	containers, err := manager.GetStatus()
	if err != nil {
		return glideErrors.Wrap(err, "failed to check container status",
			glideErrors.WithSuggestions(
				"Check if Docker is running: docker ps",
				"Verify Docker permissions",
				"Try restarting Docker Desktop",
			),
		)
	}

	// Look for the service
	found := false
	running := false
	for _, container := range containers {
		if container.Service == service {
			found = true
			if container.State == "running" {
				running = true
				break
			}
		}
	}

	if !found {
		// Build list of available services
		services := make(map[string]bool)
		var serviceList []string
		for _, container := range containers {
			if !services[container.Service] {
				services[container.Service] = true
				serviceList = append(serviceList, container.Service)
			}
		}
		
		suggestions := []string{
			fmt.Sprintf("Service '%s' not found in docker-compose.yml", service),
			"Available services: " + strings.Join(serviceList, ", "),
			"Use default: glid shell",
		}
		
		return glideErrors.NewConfigError(fmt.Sprintf("service '%s' not found", service),
			glideErrors.WithSuggestions(suggestions...),
		)
	}

	if !running {
		return glideErrors.NewDockerError(fmt.Sprintf("service '%s' is not running", service),
			glideErrors.WithSuggestions(
				"Start all containers: glid up",
				fmt.Sprintf("Start specific service: glid docker up %s", service),
				"Check container status: glid docker ps",
			),
		)
	}

	return nil
}

// detectShell determines the best shell to use for a service
func (c *ShellCommand) detectShell(service string) string {
	// For most services, try bash first, then sh
	// MySQL is special - it should use mysql client
	if service == "mysql" {
		return "mysql -u root -p$MYSQL_ROOT_PASSWORD"
	}

	// Default to bash for most containers
	// The actual detection happens in the container
	return "/bin/bash || /bin/sh"
}

// executeShell runs the docker exec command to attach to the shell
func (c *ShellCommand) executeShell(resolver *docker.Resolver, service, shellCmd string) error {
	// Build docker-compose exec command
	dockerArgs := resolver.GetComposeCommand("exec")

	// Add service and shell command
	dockerArgs = append(dockerArgs, service)

	// Parse shell command - it might be a complex command
	if strings.Contains(shellCmd, "||") {
		// Try bash first, fallback to sh
		dockerArgs = append(dockerArgs, "sh", "-c", fmt.Sprintf("if [ -x /bin/bash ]; then exec /bin/bash; else exec /bin/sh; fi"))
	} else if strings.Contains(shellCmd, " ") {
		// Complex command, wrap in sh -c
		dockerArgs = append(dockerArgs, "sh", "-c", shellCmd)
	} else {
		// Simple command
		dockerArgs = append(dockerArgs, shellCmd)
	}

	// Use os/exec directly for proper TTY handling
	dockerCmd := exec.Command("docker", dockerArgs...)
	
	// Connect stdin, stdout, stderr
	dockerCmd.Stdin = os.Stdin
	dockerCmd.Stdout = os.Stdout
	dockerCmd.Stderr = os.Stderr

	// Show what we're doing
	output.Info("Attaching to %s container...", service)
	if service == "php" {
		output.Info("Type 'exit' or press Ctrl+D to leave the shell")
	}

	// Run the command
	if err := dockerCmd.Run(); err != nil {
		// Check if it's just an exit code (normal for shells)
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Exit code from the shell is not really an error
			if exitErr.ExitCode() > 0 && exitErr.ExitCode() < 128 {
				// Normal exit
				return nil
			}
		}
		return glideErrors.Wrap(err, "failed to attach to container",
			glideErrors.WithSuggestions(
				"Check if the container is still running: glid docker ps",
				"Verify the shell exists in the container",
				"Try a different shell: glid shell " + service + " sh",
			),
		)
	}

	return nil
}