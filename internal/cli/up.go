package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/ivannovak/glide/internal/config"
	"github.com/ivannovak/glide/internal/context"
	"github.com/ivannovak/glide/internal/docker"
	"github.com/ivannovak/glide/internal/shell"
	glideErrors "github.com/ivannovak/glide/pkg/errors"
	"github.com/ivannovak/glide/pkg/output"
	"github.com/ivannovak/glide/pkg/progress"
	"github.com/spf13/cobra"
)

// UpCommand handles the up command
type UpCommand struct {
	ctx *context.ProjectContext
	cfg *config.Config
}

// NewUpCommand creates a new up command
func NewUpCommand(ctx *context.ProjectContext, cfg *config.Config) *cobra.Command {
	uc := &UpCommand{
		ctx: ctx,
		cfg: cfg,
	}

	cmd := &cobra.Command{
		Use:   "up [flags]",
		Short: "Start Docker containers",
		Long: `Start all Docker containers for the current project.

This command starts all services defined in your docker-compose.yml file.
It automatically detects the correct compose files based on your location
and development mode.

Options:
  -d, --detach        Run containers in the background (default)
  -b, --build         Build images before starting containers
  --force-recreate    Recreate containers even if configuration unchanged
  --no-deps          Don't start linked services
  --remove-orphans   Remove containers for services not in compose file
  --wait             Wait for containers to be healthy before returning
  --timeout int      Timeout in seconds for health checks (default 60)

Examples:
  glid up                    # Start all containers in background
  glid up -b                 # Rebuild and start containers
  glid up --force-recreate   # Force recreate all containers
  glid up --wait             # Wait for containers to be healthy
  glid up --remove-orphans   # Clean up orphaned containers

Container Health:
  After starting, the command performs health checks on all containers
  to ensure they're running properly. Use --wait to block until healthy.

Project Isolation:
  In multi-worktree mode, containers are isolated by project name to
  allow multiple worktrees to run simultaneously.`,
		RunE:          uc.Execute,
		SilenceUsage:  true, // Don't show usage on error
		SilenceErrors: true, // Let our error handler handle errors
	}

	// Add flags
	cmd.Flags().BoolP("detach", "d", true, "Run containers in the background")
	cmd.Flags().BoolP("build", "b", false, "Build images before starting")
	cmd.Flags().Bool("force-recreate", false, "Recreate containers even if configuration unchanged")
	cmd.Flags().Bool("no-deps", false, "Don't start linked services")
	cmd.Flags().Bool("remove-orphans", false, "Remove containers for services not in compose file")
	cmd.Flags().Bool("wait", false, "Wait for containers to be healthy")
	cmd.Flags().Int("timeout", 60, "Timeout in seconds for health checks")

	return cmd
}

// Execute runs the up command
func (c *UpCommand) Execute(cmd *cobra.Command, args []string) error {
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

	// Get flags
	detach, _ := cmd.Flags().GetBool("detach")
	build, _ := cmd.Flags().GetBool("build")
	forceRecreate, _ := cmd.Flags().GetBool("force-recreate")
	noDeps, _ := cmd.Flags().GetBool("no-deps")
	removeOrphans, _ := cmd.Flags().GetBool("remove-orphans")
	wait, _ := cmd.Flags().GetBool("wait")
	timeout, _ := cmd.Flags().GetInt("timeout")

	// Use config default for remove-orphans if not explicitly set
	if !cmd.Flags().Changed("remove-orphans") && c.cfg != nil {
		removeOrphans = c.cfg.Defaults.Docker.RemoveOrphans
	}

	// Show what we're doing
	c.showStartupMessage(resolver)

	// Build the docker-compose up command
	dockerArgs := c.buildUpCommand(resolver, detach, build, forceRecreate, noDeps, removeOrphans)

	// Execute the up command
	if err := c.executeUpCommand(dockerArgs); err != nil {
		return err
	}

	// Perform health checks
	if wait || c.shouldAutoWait() {
		if err := c.waitForHealthy(timeout); err != nil {
			output.Warning("Warning: Some containers may not be fully ready: %v", err)
			// Don't fail the command, just warn
		}
	} else {
		// Quick health check without waiting
		go c.performQuickHealthCheck()
	}

	// Show success message
	c.showSuccessMessage(detach)

	return nil
}

// buildUpCommand constructs the docker-compose up command
func (c *UpCommand) buildUpCommand(resolver *docker.Resolver, detach, build, forceRecreate, noDeps, removeOrphans bool) []string {
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

	// Add up command
	dockerArgs = append(dockerArgs, "up")

	// Add flags
	if detach {
		dockerArgs = append(dockerArgs, "-d")
	}

	if build {
		dockerArgs = append(dockerArgs, "--build")
	}

	if forceRecreate {
		dockerArgs = append(dockerArgs, "--force-recreate")
	}

	if noDeps {
		dockerArgs = append(dockerArgs, "--no-deps")
	}

	if removeOrphans {
		dockerArgs = append(dockerArgs, "--remove-orphans")
	}

	return dockerArgs
}

// executeUpCommand runs the docker-compose up command
func (c *UpCommand) executeUpCommand(dockerArgs []string) error {
	spinner := progress.NewSpinner("Starting Docker containers")
	spinner.Start()

	executor := shell.NewExecutor(shell.Options{
		Verbose: false,
	})

	// Check if we're in detached mode
	isDetached := false
	for _, arg := range dockerArgs {
		if arg == "-d" {
			isDetached = true
			break
		}
	}

	// Use appropriate command based on detached mode
	var shellCmd *shell.Command
	if isDetached {
		// For detached mode, we can capture output
		shellCmd = shell.NewCommand("docker", dockerArgs...)
	} else {
		// For attached mode, use passthrough
		spinner.Stop() // Stop spinner for attached mode
		shellCmd = shell.NewPassthroughCommand("docker", dockerArgs...)
	}

	result, err := executor.Execute(shellCmd)

	if isDetached {
		if err != nil {
			spinner.Error("Failed to start containers")
			return c.handleUpError(err)
		}

		if result.ExitCode != 0 {
			spinner.Error(fmt.Sprintf("Docker up failed with exit code %d", result.ExitCode))
			return glideErrors.NewDockerError("failed to start containers",
				glideErrors.WithSuggestions(
					"Check Docker logs: glid docker logs",
					"Verify Docker daemon is running",
					"Check for port conflicts: glid status",
					"Try rebuilding: glid up --build",
				),
				glideErrors.WithExitCode(result.ExitCode),
			)
		}

		spinner.Success("Containers started")
	} else {
		// For attached mode, just return any error
		if err != nil {
			return c.handleUpError(err)
		}
	}

	// Update context to reflect Docker is now running
	c.ctx.DockerRunning = true

	return nil
}

// waitForHealthy waits for containers to become healthy
func (c *UpCommand) waitForHealthy(timeoutSeconds int) error {
	spinner := progress.NewSpinner("Waiting for containers to be healthy")
	spinner.Start()
	defer spinner.Stop()

	health := docker.NewHealthMonitor(c.ctx)
	timeout := time.Duration(timeoutSeconds) * time.Second

	if err := health.WaitForHealthy(timeout); err != nil {
		spinner.Error("Some containers failed health checks")
		return err
	}

	spinner.Success("All containers are healthy")
	return nil
}

// performQuickHealthCheck does a quick health check without blocking
func (c *UpCommand) performQuickHealthCheck() {
	// Wait a moment for containers to start
	time.Sleep(3 * time.Second)

	health := docker.NewHealthMonitor(c.ctx)
	healthStatus, err := health.CheckHealth()
	if err != nil {
		return
	}

	// Check for unhealthy services
	unhealthy := []string{}
	for _, service := range healthStatus {
		if !service.Healthy {
			unhealthy = append(unhealthy, service.Service)
		}
	}

	if len(unhealthy) > 0 {
		output.Warning("\n⚠ Warning: Some services may not be ready: %s", strings.Join(unhealthy, ", "))
		output.Warning("Run 'glid docker ps' to check status")
	}
}

// shouldAutoWait determines if we should automatically wait for health
func (c *UpCommand) shouldAutoWait() bool {
	// Could be configured in config file
	return false
}

// showStartupMessage shows what we're about to do
func (c *UpCommand) showStartupMessage(resolver *docker.Resolver) {
	output.Info("Starting Docker containers...")

	if c.ctx.IsWorktree {
		output.Info("Project: %s", resolver.GetComposeProjectName())
	}

	// Show compose files being used (in verbose mode)
	if c.cfg != nil && c.cfg.Defaults.Test.Verbose {
		output.Info("Using compose files:")
		for _, file := range resolver.GetRelativeComposeFiles() {
			output.Printf("  - %s\n", file)
		}
	}
}

// showSuccessMessage shows success information
func (c *UpCommand) showSuccessMessage(detached bool) {
	output.Success("\n✓ Docker containers started successfully")

	if detached {
		output.Info("\nUseful commands:")
		output.Info("  glid docker ps          # Check container status")
		output.Info("  glid docker logs -f     # View logs")
		output.Info("  glid shell              # Enter PHP container")
		output.Info("  glid down               # Stop containers")
	}
}

// handleUpError provides helpful error messages
func (c *UpCommand) handleUpError(err error) error {
	errStr := err.Error()

	// Docker daemon not running
	if strings.Contains(errStr, "Cannot connect to the Docker daemon") ||
		strings.Contains(errStr, "docker daemon is not running") {
		return glideErrors.NewDockerError("Docker daemon is not running",
			glideErrors.WithSuggestions(
				"Start Docker Desktop application",
				"On Mac: open -a Docker",
				"On Linux: sudo systemctl start docker",
				"Check Docker status: docker ps",
			),
		)
	}

	// Port conflicts
	if strings.Contains(errStr, "port is already allocated") || strings.Contains(errStr, "address already in use") {
		return glideErrors.NewNetworkError("port conflict - address already in use",
			glideErrors.WithSuggestions(
				"Stop conflicting containers: glid down-all",
				"Or from root: make down-all",
				"Check what's using the port: lsof -i :PORT",
				"Change port in docker-compose.yml",
			),
		)
	}

	// Image build errors
	if strings.Contains(errStr, "error building") || strings.Contains(errStr, "failed to build") {
		return glideErrors.NewDockerError("Docker build failed",
			glideErrors.WithSuggestions(
				"Check Dockerfile syntax",
				"Ensure base images are accessible",
				"Run: glid docker build --no-cache",
				"Run: glid ecr-login if using ECR images",
			),
		)
	}

	// Network errors
	if strings.Contains(errStr, "network") && strings.Contains(errStr, "not found") {
		return glideErrors.NewNetworkError("Docker network error",
			glideErrors.WithSuggestions(
				"Reset Docker networks:",
				"Run: glid docker down",
				"Run: glid up --force-recreate",
				"Check Docker network: docker network ls",
			),
		)
	}

	return glideErrors.AnalyzeError(err)
}
