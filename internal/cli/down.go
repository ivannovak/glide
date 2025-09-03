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
	"github.com/ivannovak/glide/pkg/prompt"
	"github.com/spf13/cobra"
)

// DownCommand handles the down command
type DownCommand struct {
	ctx *context.ProjectContext
	cfg *config.Config
}

// NewDownCommand creates a new down command
func NewDownCommand(ctx *context.ProjectContext, cfg *config.Config) *cobra.Command {
	dc := &DownCommand{
		ctx: ctx,
		cfg: cfg,
	}

	cmd := &cobra.Command{
		Use:   "down [flags]",
		Short: "Stop Docker containers",
		Long: `Stop and remove Docker containers for the current project.

This command stops all running containers and removes them along with
their networks. By default, volumes are preserved to maintain data.

Options:
  -v, --volumes           Remove named volumes declared in the compose file
  --remove-orphans        Remove containers for services not in compose file
  -t, --timeout int       Timeout in seconds before killing containers (default 10)
  --rmi string           Remove images (all/local)

Examples:
  glid down                      # Stop and remove containers
  glid down -v                   # Also remove volumes (WARNING: data loss)
  glid down --remove-orphans     # Clean up orphaned containers
  glid down --rmi local          # Remove locally built images
  glid down --rmi all            # Remove all images used by services
  glid down -t 30                # Give containers 30 seconds to stop

Data Persistence:
  By default, volumes are preserved when stopping containers. This means:
  - Database data is retained
  - File uploads are preserved
  - Redis cache persists
  
  Use -v flag only when you want to completely reset your environment.

Graceful Shutdown:
  Containers are given time to shut down gracefully before being killed.
  Use --timeout to adjust this period for slow-stopping services.`,
		RunE:          dc.Execute,
		SilenceUsage:  true, // Don't show usage on error
		SilenceErrors: true, // Let our error handler handle errors
	}

	// Add flags
	cmd.Flags().BoolP("volumes", "v", false, "Remove named volumes")
	cmd.Flags().Bool("remove-orphans", false, "Remove containers for services not in compose file")
	cmd.Flags().IntP("timeout", "t", 10, "Timeout in seconds before killing containers")
	cmd.Flags().String("rmi", "", "Remove images (all/local)")

	return cmd
}

// Execute runs the down command
func (c *DownCommand) Execute(cmd *cobra.Command, args []string) error {
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
		output.Warning("No docker-compose files found in this project")
		return nil
	}

	// Get flags
	volumes, _ := cmd.Flags().GetBool("volumes")
	removeOrphans, _ := cmd.Flags().GetBool("remove-orphans")
	timeout, _ := cmd.Flags().GetInt("timeout")
	rmi, _ := cmd.Flags().GetString("rmi")

	// Use config default for remove-orphans if not explicitly set
	if !cmd.Flags().Changed("remove-orphans") && c.cfg != nil {
		removeOrphans = c.cfg.Defaults.Docker.RemoveOrphans
	}

	// Warn about data loss if using -v
	if volumes {
		if !c.confirmVolumeDeletion() {
			output.Warning("Operation cancelled")
			return nil
		}
	}

	// Show what we're doing
	c.showShutdownMessage(resolver, volumes)

	// Build the docker-compose down command
	dockerArgs := c.buildDownCommand(resolver, volumes, removeOrphans, timeout, rmi)

	// Execute the down command
	if err := c.executeDownCommand(dockerArgs); err != nil {
		return err
	}

	// Show success message
	c.showSuccessMessage(volumes)

	return nil
}

// buildDownCommand constructs the docker-compose down command
func (c *DownCommand) buildDownCommand(resolver *docker.Resolver, volumes, removeOrphans bool, timeout int, rmi string) []string {
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

	// Add down command
	dockerArgs = append(dockerArgs, "down")

	// Add timeout
	dockerArgs = append(dockerArgs, "-t", fmt.Sprintf("%d", timeout))

	// Add flags
	if volumes {
		dockerArgs = append(dockerArgs, "-v")
	}

	if removeOrphans {
		dockerArgs = append(dockerArgs, "--remove-orphans")
	}

	if rmi != "" {
		dockerArgs = append(dockerArgs, "--rmi", rmi)
	}

	return dockerArgs
}

// executeDownCommand runs the docker-compose down command
func (c *DownCommand) executeDownCommand(dockerArgs []string) error {
	spinner := progress.NewSpinner("Stopping Docker containers")
	spinner.Start()

	executor := shell.NewExecutor(shell.Options{
		Verbose: false,
	})

	// Set a longer timeout for the down command itself
	// This should be longer than the container stop timeout
	shellCmd := shell.NewCommand("docker", dockerArgs...)
	shellCmd.Timeout = 2 * time.Minute

	result, err := executor.Execute(shellCmd)

	if err != nil {
		spinner.Error("Failed to stop containers")
		return c.handleDownError(err)
	}

	if result.ExitCode != 0 {
		// Check if it's just "no containers to stop"
		if result.ExitCode == 1 {
			spinner.Success("No containers were running")
			return nil
		}
		spinner.Error(fmt.Sprintf("Docker down failed with exit code %d", result.ExitCode))
		return glideErrors.NewDockerError("failed to stop containers",
			glideErrors.WithExitCode(result.ExitCode),
			glideErrors.WithSuggestions(
				"Check if containers are still running: glid docker ps",
				"Try stopping with longer timeout: glid down -t 30",
				"Force stop all containers: docker kill $(docker ps -q)",
			),
		)
	}

	spinner.Success("Containers stopped and removed")

	// Update context to reflect Docker is no longer running
	c.ctx.DockerRunning = false

	return nil
}

// confirmVolumeDeletion asks for confirmation before deleting volumes
func (c *DownCommand) confirmVolumeDeletion() bool {
	output.Warning("\n⚠ WARNING: The -v flag will delete all data in Docker volumes!")
	output.Warning("This includes:")
	output.Warning("  - Database data")
	output.Warning("  - Uploaded files")
	output.Warning("  - Cache data")
	output.Warning("  - Any other data stored in volumes")

	output.Println()

	// Use the destructive confirmation for volume deletion
	confirmed, _ := prompt.ConfirmDestructive("delete all Docker volumes and their data")
	return confirmed
}

// showShutdownMessage shows what we're about to do
func (c *DownCommand) showShutdownMessage(resolver *docker.Resolver, volumes bool) {
	output.Info("Stopping Docker containers...")

	if c.ctx.IsWorktree {
		output.Info("Project: %s", resolver.GetComposeProjectName())
	}

	if volumes {
		output.Warning("Volumes will be removed (data will be lost)")
	}
}

// showSuccessMessage shows success information
func (c *DownCommand) showSuccessMessage(volumes bool) {
	output.Success("\n✓ Docker containers stopped and removed")

	if volumes {
		output.Warning("Volumes have been removed")
		output.Info("\nTo recreate with fresh data:")
		output.Info("  glid up")
		output.Info("  glid artisan migrate:fresh --seed")
	} else {
		output.Info("\nData in volumes has been preserved")
		output.Info("To restart containers: glid up")
	}
}

// handleDownError provides helpful error messages
func (c *DownCommand) handleDownError(err error) error {
	errStr := err.Error()

	// Docker daemon not running
	if strings.Contains(strings.ToLower(errStr), "cannot connect to the docker daemon") {
		// This is actually OK for down command - nothing to stop
		output.Success("✓ Docker is not running (nothing to stop)")
		return nil
	}

	// Permission errors
	if strings.Contains(strings.ToLower(errStr), "permission denied") {
		return glideErrors.NewPermissionError("/var/run/docker.sock", "permission denied stopping containers",
			glideErrors.WithError(err),
			glideErrors.WithSuggestions(
				"Add your user to docker group: sudo usermod -aG docker $USER",
				"Log out and back in for group changes to take effect",
				"Or run with appropriate permissions",
			),
		)
	}

	// Timeout errors
	if strings.Contains(strings.ToLower(errStr), "timeout") {
		return glideErrors.NewTimeoutError("containers took too long to stop",
			glideErrors.WithError(err),
			glideErrors.WithSuggestions(
				"Wait and try again",
				"Use a longer timeout: glid down -t 60",
				"Force stop all containers: docker kill $(docker ps -q)",
				"Check for hanging processes in containers",
			),
		)
	}

	return glideErrors.AnalyzeError(err)
}
