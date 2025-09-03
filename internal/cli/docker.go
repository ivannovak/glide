package cli

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ivannovak/glide/internal/config"
	"github.com/ivannovak/glide/internal/context"
	"github.com/ivannovak/glide/internal/docker"
	"github.com/ivannovak/glide/internal/shell"
	glideErrors "github.com/ivannovak/glide/pkg/errors"
	"github.com/ivannovak/glide/pkg/output"
	"github.com/spf13/cobra"
)

// DockerCommand handles the docker command
type DockerCommand struct {
	ctx *context.ProjectContext
	cfg *config.Config
}

// NewDockerCommand creates a new docker command
func NewDockerCommand(ctx *context.ProjectContext, cfg *config.Config) *cobra.Command {
	dc := &DockerCommand{
		ctx: ctx,
		cfg: cfg,
	}

	cmd := &cobra.Command{
		Use:   "docker [docker-compose arguments]",
		Short: "Pass-through to docker-compose with automatic file resolution",
		Long: `Execute docker-compose commands with automatic compose file resolution.

This command is a complete pass-through to docker-compose, automatically
resolving the correct compose files based on your location and development mode.

All arguments are passed directly to docker-compose without modification.

Examples:
  glid docker up -d                  # Start containers in background
  glid docker down                    # Stop containers
  glid docker ps                      # List containers
  glid docker exec php bash           # Enter PHP container
  glid docker exec -it php bash       # Interactive shell
  glid docker logs -f nginx           # Follow nginx logs
  glid docker restart php             # Restart PHP container
  glid docker build --no-cache        # Rebuild containers

Compose File Resolution:
  The command automatically determines the correct compose files:
  - From vcs/: Uses docker-compose.yml and ../docker-compose.override.yml
  - From worktrees/*/: Uses docker-compose.yml and ../../docker-compose.override.yml
  - Handles missing override files gracefully

Interactive Commands:
  TTY allocation is automatically handled for interactive commands like:
  - exec with -it flags
  - run commands
  - shell access

Signal Handling:
  Ctrl+C and other signals are properly forwarded to docker-compose.`,
		DisableFlagParsing: true, // Pass all flags through to docker-compose
		RunE:               dc.Execute,
		SilenceUsage:       true, // Don't show usage on error
		SilenceErrors:      true, // Let our error handler handle errors
	}

	return cmd
}

// Execute runs the docker command
func (c *DockerCommand) Execute(cmd *cobra.Command, args []string) error {
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

	// Resolve Docker compose files
	resolver := docker.NewResolver(c.ctx)
	if err := resolver.Resolve(); err != nil {
		return glideErrors.Wrap(err, "failed to resolve Docker configuration",
			glideErrors.WithSuggestions(
				"Check if docker-compose.yml exists",
				"Verify Docker is installed: docker --version",
				"Ensure you're in the project root or worktree",
			),
		)
	}

	// Check if compose files exist
	if len(c.ctx.ComposeFiles) == 0 {
		return glideErrors.NewFileNotFoundError("docker-compose.yml",
			glideErrors.WithSuggestions(
				"Check if docker-compose.yml exists in the project root",
				"Ensure you're in the correct directory",
				"For worktrees, check if files were copied from vcs/",
			),
		)
	}

	// Build the docker-compose command
	dockerArgs := c.buildDockerCommand(resolver, args)

	// Show what we're running (only in verbose mode)
	if c.cfg != nil && c.cfg.Defaults.Test.Verbose {
		output.Info("Running: docker %s", shell.JoinArgs(dockerArgs))
	}

	// Check if this is an interactive command
	isInteractive := c.isInteractiveCommand(args)

	// Execute the command with proper signal handling
	return c.executeDockerCommand(dockerArgs, isInteractive)
}

// buildDockerCommand constructs the full docker-compose command
func (c *DockerCommand) buildDockerCommand(resolver *docker.Resolver, args []string) []string {
	// Start with compose subcommand
	dockerArgs := []string{"compose"}

	// Add compose file flags
	for _, file := range resolver.GetRelativeComposeFiles() {
		dockerArgs = append(dockerArgs, "-f", file)
	}

	// Add project name if in worktree
	if c.ctx.IsWorktree && c.ctx.WorktreeName != "" {
		projectName := resolver.GetComposeProjectName()
		dockerArgs = append(dockerArgs, "-p", projectName)
	}

	// Add all user-provided arguments
	dockerArgs = append(dockerArgs, args...)

	return dockerArgs
}

// isInteractiveCommand checks if the command requires TTY allocation
func (c *DockerCommand) isInteractiveCommand(args []string) bool {
	// Check for explicit -it or -i flags
	for _, arg := range args {
		if arg == "-it" || arg == "-i" || arg == "--interactive" {
			return true
		}
	}

	// Check for commands that are typically interactive
	if len(args) > 0 {
		switch args[0] {
		case "exec":
			// exec is interactive if it has -it flag or ends with shell command
			if len(args) > 2 {
				lastArg := args[len(args)-1]
				// Common shell commands
				if lastArg == "bash" || lastArg == "sh" || lastArg == "/bin/bash" || lastArg == "/bin/sh" {
					return true
				}
			}
		case "run":
			// run is often interactive
			return true
		}
	}

	return false
}

// executeDockerCommand executes the docker-compose command with proper handling
func (c *DockerCommand) executeDockerCommand(dockerArgs []string, isInteractive bool) error {
	// Create the command
	var shellCmd *shell.Command
	if isInteractive {
		// Use passthrough for interactive commands
		shellCmd = shell.NewPassthroughCommand("docker", dockerArgs...)
	} else {
		// Use regular command for non-interactive
		shellCmd = shell.NewCommand("docker", dockerArgs...)
	}

	// Set up signal handling to forward signals to docker-compose
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Create executor
	executor := shell.NewExecutor(shell.Options{
		Verbose: false,
	})

	// Execute in a goroutine so we can handle signals
	type execResult struct {
		result *shell.Result
		err    error
	}

	resultChan := make(chan execResult, 1)
	go func() {
		result, err := executor.Execute(shellCmd)
		resultChan <- execResult{result: result, err: err}
	}()

	// Wait for either command completion or signal
	select {
	case result := <-resultChan:
		// Command completed normally
		if result.err != nil {
			// Check for common Docker errors and provide helpful messages
			if err := c.handleDockerError(result.err, dockerArgs); err != nil {
				return err
			}
			return result.err
		}

		if result.result.ExitCode != 0 {
			// Don't print error message for expected non-zero exits (like docker ps when no containers)
			if !c.isExpectedNonZeroExit(dockerArgs, result.result.ExitCode) {
				return glideErrors.NewDockerError(fmt.Sprintf("docker-compose exited with code %d", result.result.ExitCode),
					glideErrors.WithExitCode(result.result.ExitCode),
					glideErrors.WithSuggestions(
						"Check the docker-compose output above for errors",
						"Verify Docker is running: docker ps",
						"Check Docker logs: glid docker logs",
					),
				)
			}
		}

		return nil

	case sig := <-sigChan:
		// Signal received, forward it to docker-compose
		// The shell executor should handle this, but we'll ensure cleanup
		output.Warning("\nReceived signal %v, stopping docker-compose...", sig)

		// Wait a bit for graceful shutdown
		select {
		case result := <-resultChan:
			if result.err != nil && result.err.Error() != "signal: interrupt" {
				return result.err
			}
		case <-sigChan:
			// Force kill if needed
			return glideErrors.NewCommandError("docker", 130) // Standard SIGINT exit code
		}

		return nil
	}
}

// handleDockerError provides helpful error messages for common Docker issues
func (c *DockerCommand) handleDockerError(err error, args []string) error {
	errStr := err.Error()

	// Docker daemon not running
	if strings.Contains(strings.ToLower(errStr), "cannot connect to the docker daemon") ||
		strings.Contains(strings.ToLower(errStr), "docker daemon is not running") {
		return glideErrors.NewDockerError("Docker daemon is not running",
			glideErrors.WithError(err),
			glideErrors.WithSuggestions(
				"Start Docker Desktop application",
				"On Mac: open -a Docker",
				"On Linux: sudo systemctl start docker",
				"Check Docker status: docker ps",
			),
		)
	}

	// Permission denied
	if strings.Contains(strings.ToLower(errStr), "permission denied") {
		return glideErrors.NewPermissionError("/var/run/docker.sock", "Docker permission denied",
			glideErrors.WithError(err),
			glideErrors.WithSuggestions(
				"Add your user to docker group: sudo usermod -aG docker $USER",
				"Log out and back in for group changes to take effect",
				"On Linux, you may need to run with sudo",
				"Check Docker socket permissions: ls -la /var/run/docker.sock",
			),
		)
	}

	// Compose file not found
	if strings.Contains(strings.ToLower(errStr), "no such file") ||
		strings.Contains(strings.ToLower(errStr), "not found") {
		suggestions := []string{
			"Make sure you're in a project directory",
			"Check if docker-compose.yml exists",
		}
		for _, file := range c.ctx.ComposeFiles {
			suggestions = append(suggestions, fmt.Sprintf("Expected file: %s", file))
		}
		return glideErrors.NewFileNotFoundError("docker-compose files",
			glideErrors.WithError(err),
			glideErrors.WithSuggestions(suggestions...),
		)
	}

	// Port already in use
	if strings.Contains(strings.ToLower(errStr), "port is already allocated") ||
		strings.Contains(strings.ToLower(errStr), "address already in use") {
		return glideErrors.NewNetworkError("port conflict - address already in use",
			glideErrors.WithError(err),
			glideErrors.WithSuggestions(
				"Stop conflicting containers: glid docker down",
				"Or from root: make down-all",
				"Check what's using the port: lsof -i :PORT",
				"Change port in docker-compose.yml",
			),
		)
	}

	// Return analyzed error if no specific pattern matched
	return glideErrors.AnalyzeError(err)
}

// isExpectedNonZeroExit checks if a non-zero exit code is expected for the command
func (c *DockerCommand) isExpectedNonZeroExit(args []string, exitCode int) bool {
	// Some docker-compose commands return non-zero when there's nothing wrong
	// For example, 'ps' returns 1 when no containers are running
	if len(args) > 1 && args[1] == "ps" && exitCode == 1 {
		return true
	}

	return false
}
