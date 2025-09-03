package cli

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ivannovak/glide/internal/config"
	"github.com/ivannovak/glide/internal/context"
	"github.com/ivannovak/glide/internal/docker"
	glideErrors "github.com/ivannovak/glide/pkg/errors"
	"github.com/ivannovak/glide/pkg/output"
	"github.com/spf13/cobra"
)

// LogsCommand handles the logs command
type LogsCommand struct {
	ctx *context.ProjectContext
	cfg *config.Config
}

// NewLogsCommand creates a new logs command
func NewLogsCommand(ctx *context.ProjectContext, cfg *config.Config) *cobra.Command {
	lc := &LogsCommand{
		ctx: ctx,
		cfg: cfg,
	}

	cmd := &cobra.Command{
		Use:   "logs [service] [flags]",
		Short: "View container logs",
		Long: `Display logs from Docker containers.

This command shows logs from one or all services. By default, it shows
logs from all services. Use -f to follow log output in real-time.

Examples:
  glid logs                    # Show all logs
  glid logs php                # Show PHP container logs
  glid logs -f                 # Follow all logs (Ctrl+C to stop)
  glid logs php -f             # Follow PHP logs
  glid logs --tail 50          # Show last 50 lines from all services
  glid logs php --tail 100     # Show last 100 lines from PHP
  glid logs --since 1h         # Show logs from last hour
  glid logs --since 2023-01-01 # Show logs since date
  glid logs -t                 # Show timestamps
  glid logs --no-color         # Disable colored output

Services:
  Specify a service name to see logs from that container only.
  Common services:
  - php
  - nginx
  - mysql
  - redis

Filtering:
  --tail N        Show last N lines (default shows all)
  --since TIME    Show logs since timestamp (e.g., 2023-01-01) or relative time (e.g., 1h, 30m)
  --until TIME    Show logs before timestamp
  --grep PATTERN  Filter logs by pattern (simple string match)

Output Options:
  -f, --follow    Follow log output (real-time streaming)
  -t, --timestamps  Show timestamps
  --no-color      Disable colored output
  --no-prefix     Don't print service name prefix

Tips:
  - Use Ctrl+C to stop following logs
  - Logs are color-coded by service for easier reading
  - Pipe to grep for advanced filtering: glid logs | grep ERROR
  - Save logs to file: glid logs > logs.txt`,
		Args:          cobra.MaximumNArgs(1),
		RunE:          lc.Execute,
		SilenceUsage:  true, // Don't show usage on error
		SilenceErrors: true, // Let our error handler handle errors
	}

	// Add flags
	cmd.Flags().BoolP("follow", "f", false, "Follow log output")
	cmd.Flags().BoolP("timestamps", "t", false, "Show timestamps")
	cmd.Flags().Int("tail", 0, "Number of lines to show from the end of the logs")
	cmd.Flags().String("since", "", "Show logs since timestamp (e.g., 2023-01-01) or relative time (e.g., 1h)")
	cmd.Flags().String("until", "", "Show logs before timestamp")
	cmd.Flags().Bool("no-color", false, "Disable colored output")
	cmd.Flags().Bool("no-prefix", false, "Don't print service name prefix")
	cmd.Flags().String("grep", "", "Filter logs by pattern")

	return cmd
}

// Execute runs the logs command
func (c *LogsCommand) Execute(cmd *cobra.Command, args []string) error {
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
		output.Warning("Docker is not running. No containers to show logs from.")
		output.Warning("Start containers with: glid up")
		return nil
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

	// Get service name if specified
	service := ""
	if len(args) > 0 {
		service = args[0]
		// Verify service exists
		if err := c.verifyService(service); err != nil {
			return err
		}
	}

	// Get flags
	follow, _ := cmd.Flags().GetBool("follow")
	timestamps, _ := cmd.Flags().GetBool("timestamps")
	tail, _ := cmd.Flags().GetInt("tail")
	since, _ := cmd.Flags().GetString("since")
	until, _ := cmd.Flags().GetString("until")
	noColor, _ := cmd.Flags().GetBool("no-color")
	noPrefix, _ := cmd.Flags().GetBool("no-prefix")
	grepPattern, _ := cmd.Flags().GetString("grep")

	// Build and execute logs command
	return c.executeLogs(resolver, service, follow, timestamps, tail, since, until, noColor, noPrefix, grepPattern)
}

// verifyService checks if the specified service exists
func (c *LogsCommand) verifyService(service string) error {
	manager := docker.NewContainerManager(c.ctx)
	services, err := manager.GetComposeServices()
	if err != nil {
		// Don't fail, just warn
		return nil
	}

	found := false
	for _, svc := range services {
		if svc == service {
			found = true
			break
		}
	}

	if !found {
		output.Warning("Warning: Service '%s' not found in compose file", service)
		output.Warning("Available services: %s", strings.Join(services, ", "))
		// Continue anyway - docker-compose will handle the error
	}

	return nil
}

// executeLogs runs the docker-compose logs command
func (c *LogsCommand) executeLogs(resolver *docker.Resolver, service string, follow, timestamps bool, tail int, since, until string, noColor, noPrefix bool, grepPattern string) error {
	// Build docker-compose logs command
	dockerArgs := resolver.GetComposeCommand("logs")

	// Add flags
	if follow {
		dockerArgs = append(dockerArgs, "-f")
	}

	if timestamps {
		dockerArgs = append(dockerArgs, "-t")
	}

	if tail > 0 {
		dockerArgs = append(dockerArgs, "--tail", fmt.Sprintf("%d", tail))
	}

	if since != "" {
		dockerArgs = append(dockerArgs, "--since", since)
	}

	if until != "" {
		dockerArgs = append(dockerArgs, "--until", until)
	}

	if noColor {
		dockerArgs = append(dockerArgs, "--no-color")
	}

	if noPrefix {
		dockerArgs = append(dockerArgs, "--no-log-prefix")
	}

	// Add service if specified
	if service != "" {
		dockerArgs = append(dockerArgs, service)
	}

	// Handle grep pattern with pipe if specified
	if grepPattern != "" {
		return c.executeLogsWithGrep(dockerArgs, grepPattern, follow)
	}

	// Use os/exec directly for proper signal handling
	dockerCmd := exec.Command("docker", dockerArgs...)

	// Connect stdout and stderr
	dockerCmd.Stdout = os.Stdout
	dockerCmd.Stderr = os.Stderr

	// Set up signal handling for graceful shutdown
	if follow {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			<-sigChan
			// Kill the docker logs process
			if dockerCmd.Process != nil {
				dockerCmd.Process.Kill()
			}
		}()

		// Show instructions
		output.Info("Following logs... Press Ctrl+C to stop")
		if service != "" {
			output.Info("Showing logs for: %s", service)
		} else {
			output.Info("Showing logs for all services")
		}
		output.Println()
	}

	// Run the command
	if err := dockerCmd.Run(); err != nil {
		// Check if it's just an interrupt (Ctrl+C)
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == -1 || exitErr.ExitCode() == 130 {
				// Interrupted by signal, this is expected
				output.Println() // New line after Ctrl+C
				return nil
			}
		}

		// Check for common errors
		if strings.Contains(err.Error(), "no such service") {
			return glideErrors.NewConfigError(fmt.Sprintf("service '%s' not found", service),
				glideErrors.WithError(err),
				glideErrors.WithSuggestions(
					"Check available services: glid docker ps",
					"Show all logs: glid logs",
					"List services: glid status",
				),
			)
		}

		return glideErrors.Wrap(err, "failed to get logs",
			glideErrors.WithSuggestions(
				"Check if containers are running: glid docker ps",
				"Start containers: glid up",
				"View Docker daemon logs for errors",
			),
		)
	}

	return nil
}

// executeLogsWithGrep runs logs with grep filtering
func (c *LogsCommand) executeLogsWithGrep(dockerArgs []string, pattern string, follow bool) error {
	// Create docker command
	dockerCmd := exec.Command("docker", dockerArgs...)

	// Create grep command
	grepArgs := []string{"--line-buffered", pattern}
	if !follow {
		// Add color to grep output when not following
		grepArgs = append([]string{"--color=auto"}, grepArgs...)
	}
	grepCmd := exec.Command("grep", grepArgs...)

	// Create pipe
	pipe, err := dockerCmd.StdoutPipe()
	if err != nil {
		return glideErrors.Wrap(err, "failed to create pipe for log filtering",
			glideErrors.WithSuggestions(
				"Try without grep filter: glid logs",
				"Use Docker directly: docker compose logs | grep '"+pattern+"'",
			),
		)
	}

	// Connect grep input to docker output
	grepCmd.Stdin = pipe
	grepCmd.Stdout = os.Stdout
	grepCmd.Stderr = os.Stderr

	// Also capture docker stderr
	dockerCmd.Stderr = os.Stderr

	// Start both commands
	if err := dockerCmd.Start(); err != nil {
		return glideErrors.Wrap(err, "failed to start docker logs",
			glideErrors.WithSuggestions(
				"Check if Docker is running: docker ps",
				"Verify container exists: glid docker ps",
			),
		)
	}

	if err := grepCmd.Start(); err != nil {
		dockerCmd.Process.Kill()
		return glideErrors.Wrap(err, "failed to start grep filter",
			glideErrors.WithSuggestions(
				"Check if grep is available on your system",
				"Try without filtering: glid logs",
			),
		)
	}

	// Set up signal handling
	if follow {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			<-sigChan
			// Kill both processes
			if dockerCmd.Process != nil {
				dockerCmd.Process.Kill()
			}
			if grepCmd.Process != nil {
				grepCmd.Process.Kill()
			}
		}()

		output.Info("Following logs with filter '%s'... Press Ctrl+C to stop", pattern)
		output.Println()
	}

	// Wait for both commands
	dockerErr := dockerCmd.Wait()
	grepErr := grepCmd.Wait()

	// Check for errors (ignore grep exit code 1 = no matches)
	if dockerErr != nil {
		if exitErr, ok := dockerErr.(*exec.ExitError); ok {
			if exitErr.ExitCode() == -1 || exitErr.ExitCode() == 130 {
				// Interrupted by signal
				return nil
			}
		}
		return glideErrors.Wrap(dockerErr, "docker logs command failed",
			glideErrors.WithSuggestions(
				"Check if containers are still running: glid docker ps",
				"Try with fewer options: glid logs",
			),
		)
	}

	if grepErr != nil {
		if exitErr, ok := grepErr.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				// No matches found
				output.Warning("No logs matching pattern '%s'", pattern)
				return nil
			}
		}
	}

	return nil
}
