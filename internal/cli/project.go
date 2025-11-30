package cli

import (
	"github.com/ivannovak/glide/v3/internal/config"
	"github.com/ivannovak/glide/v3/internal/context"
	"github.com/spf13/cobra"
)

// ProjectCommand handles project-wide commands for multi-worktree mode
type ProjectCommand struct {
	ctx *context.ProjectContext
	cfg *config.Config
}

// NewProjectCommand creates the project command group
func NewProjectCommand(ctx *context.ProjectContext, cfg *config.Config) *cobra.Command {
	pc := &ProjectCommand{
		ctx: ctx,
		cfg: cfg,
	}

	cmd := &cobra.Command{
		Use:     "project",
		Aliases: []string{"p"},
		Short:   "Project-wide commands for managing all worktrees",
		Long: `Project-wide commands for multi-worktree development mode.

These commands operate across all worktrees in your project.
They are only available when using multi-worktree development mode.

Available Commands:
  status         Show Docker status for all worktrees
  down           Stop all Docker containers across all worktrees
  worktree       Create and manage worktrees
  list           List all active worktrees
  clean          Clean up orphaned containers and resources

Examples:
  glide p status                    # Show status of all worktrees
  glide p down                      # Stop all containers
  glide p worktree feature/new      # Create new worktree
  glide p list                      # List all worktrees
  glide p clean --orphaned          # Clean orphaned containers

Note:
  These commands are only available in multi-worktree mode.
  Use 'glide setup' to configure your development mode.`,
		PersistentPreRunE: pc.validateMode,
	}

	// Add subcommands
	cmd.AddCommand(pc.newStatusCommand())
	cmd.AddCommand(pc.newDownCommand())
	cmd.AddCommand(pc.newWorktreeCommand())
	cmd.AddCommand(pc.newListCommand())
	cmd.AddCommand(pc.newCleanCommand())

	return cmd
}

// validateMode ensures we're in multi-worktree mode
func (pc *ProjectCommand) validateMode(cmd *cobra.Command, args []string) error {
	return ValidateMultiWorktreeMode(pc.ctx, "project")
}

// newStatusCommand creates the global status command
func (pc *ProjectCommand) newStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show Docker status for all worktrees",
		Long: `Display Docker container status across all worktrees.

This command shows the status of Docker containers in:
  - Main repository (vcs/)
  - All active worktrees (worktrees/*)

Status includes:
  - Container state (running/stopped)
  - Health status
  - Resource usage
  - Port mappings

Note: vcs/ should typically stay on the main branch as a reference.

Examples:
  glide p status                # Show all worktree statuses
  glide p status --verbose      # Include detailed container info`,
		RunE: pc.executeStatus,
	}

	// Add flags
	cmd.Flags().Bool("verbose", false, "Show detailed container information")

	return cmd
}

// newDownCommand creates the global down command
func (pc *ProjectCommand) newDownCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down",
		Short: "Stop all Docker containers across all worktrees",
		Long: `Stop Docker containers in all worktrees.

This command stops containers in:
  - Main repository (vcs/)
  - All active worktrees (worktrees/*)

Options:
  --remove-orphans  Remove orphaned containers
  --volumes         Remove volumes (WARNING: deletes data)

Examples:
  glide p down                    # Stop all containers
  glide p down --remove-orphans   # Also remove orphaned containers
  glide p down --volumes          # Also remove volumes (data loss!)`,
		RunE: pc.executeDown,
	}

	// Add flags
	cmd.Flags().Bool("remove-orphans", false, "Remove orphaned containers")
	cmd.Flags().Bool("volumes", false, "Remove volumes (WARNING: deletes data)")

	return cmd
}

// newWorktreeCommand creates the worktree management command
func (pc *ProjectCommand) newWorktreeCommand() *cobra.Command {
	// Use the actual worktree implementation
	return NewWorktreeCommand(pc.ctx, pc.cfg)
}

// newListCommand creates the list command
func (pc *ProjectCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all active worktrees",
		Long: `List all active Git worktrees in the project.

Shows information about each worktree:
  - Branch name
  - Directory path
  - Docker status
  - Last commit

Examples:
  glide p list                      # List all worktrees
  glide p ls                        # Short alias
  glide p list --format json        # JSON output`,
		RunE: pc.executeList,
	}

	// Add flags
	cmd.Flags().String("format", "table", "Output format (table or json)")

	return cmd
}

// newCleanCommand creates the clean command
func (pc *ProjectCommand) newCleanCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean up orphaned containers and resources",
		Long: `Clean up orphaned Docker containers and unused resources.

This command helps maintain a clean development environment by:
  - Removing orphaned Docker containers
  - Cleaning up unused Docker volumes
  - Removing dangling images
  - Cleaning unused networks

Options:
  --orphaned     Remove orphaned containers only
  --volumes      Also remove unused volumes
  --images       Also remove dangling images
  --all          Clean everything (containers, volumes, images)
  --dry-run      Show what would be cleaned without doing it

Examples:
  glide p clean                    # Interactive cleanup
  glide p clean --orphaned         # Remove orphaned containers
  glide p clean --all              # Full cleanup
  glide p clean --dry-run          # Preview cleanup`,
		RunE: pc.executeClean,
	}

	// Add flags
	cmd.Flags().Bool("orphaned", false, "Remove orphaned containers only")
	cmd.Flags().Bool("volumes", false, "Also remove unused volumes")
	cmd.Flags().Bool("images", false, "Also remove dangling images")
	cmd.Flags().Bool("all", false, "Clean everything")
	cmd.Flags().Bool("dry-run", false, "Show what would be cleaned")

	return cmd
}

// Command implementations - delegated to separate files

func (pc *ProjectCommand) executeStatus(cmd *cobra.Command, args []string) error {
	return ExecuteProjectStatus(pc.ctx, pc.cfg, cmd, args)
}

func (pc *ProjectCommand) executeDown(cmd *cobra.Command, args []string) error {
	return ExecuteProjectDown(pc.ctx, pc.cfg, cmd, args)
}

func (pc *ProjectCommand) executeList(cmd *cobra.Command, args []string) error {
	return ExecuteProjectList(pc.ctx, pc.cfg, cmd, args)
}

func (pc *ProjectCommand) executeClean(cmd *cobra.Command, args []string) error {
	return ExecuteProjectClean(pc.ctx, pc.cfg, cmd, args)
}
