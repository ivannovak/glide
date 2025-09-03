package cli

import (
	"github.com/ivannovak/glide/internal/config"
	"github.com/ivannovak/glide/internal/context"
	"github.com/spf13/cobra"
)

// GlobalCommand handles global commands for multi-worktree mode
type GlobalCommand struct {
	ctx *context.ProjectContext
	cfg *config.Config
}

// NewGlobalCommand creates the global command group
func NewGlobalCommand(ctx *context.ProjectContext, cfg *config.Config) *cobra.Command {
	gc := &GlobalCommand{
		ctx: ctx,
		cfg: cfg,
	}

	cmd := &cobra.Command{
		Use:     "global",
		Aliases: []string{"g"},
		Short:   "Global commands for managing all worktrees",
		Long: `Global commands for multi-worktree development mode.

These commands operate across all worktrees in your project.
They are only available when using multi-worktree development mode.

Available Commands:
  status         Show Docker status for all worktrees
  down           Stop all Docker containers across all worktrees
  worktree       Create and manage worktrees
  list           List all active worktrees
  clean          Clean up orphaned containers and resources

Examples:
  glid g status                    # Show status of all worktrees
  glid g down                      # Stop all containers
  glid g worktree feature/new      # Create new worktree
  glid g list                      # List all worktrees
  glid g clean --orphaned          # Clean orphaned containers

Note:
  These commands are only available in multi-worktree mode.
  Use 'glid setup' to configure your development mode.`,
		PersistentPreRunE: gc.validateMode,
	}

	// Add subcommands
	cmd.AddCommand(gc.newStatusCommand())
	cmd.AddCommand(gc.newDownCommand())
	cmd.AddCommand(gc.newWorktreeCommand())
	cmd.AddCommand(gc.newListCommand())
	cmd.AddCommand(gc.newCleanCommand())

	return cmd
}

// validateMode ensures we're in multi-worktree mode
func (gc *GlobalCommand) validateMode(cmd *cobra.Command, args []string) error {
	return ValidateMultiWorktreeMode(gc.ctx, "global")
}

// newStatusCommand creates the global status command
func (gc *GlobalCommand) newStatusCommand() *cobra.Command {
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

Examples:
  glid g status                # Show all worktree statuses
  glid g status --verbose      # Include detailed container info`,
		RunE: gc.executeStatus,
	}
	
	// Add flags
	cmd.Flags().Bool("verbose", false, "Show detailed container information")
	
	return cmd
}

// newDownCommand creates the global down command
func (gc *GlobalCommand) newDownCommand() *cobra.Command {
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
  glid g down                    # Stop all containers
  glid g down --remove-orphans   # Also remove orphaned containers
  glid g down --volumes          # Also remove volumes (data loss!)`,
		RunE: gc.executeDown,
	}
	
	// Add flags
	cmd.Flags().Bool("remove-orphans", false, "Remove orphaned containers")
	cmd.Flags().Bool("volumes", false, "Remove volumes (WARNING: deletes data)")
	
	return cmd
}

// newWorktreeCommand creates the worktree management command
func (gc *GlobalCommand) newWorktreeCommand() *cobra.Command {
	// Use the actual worktree implementation
	return NewWorktreeCommand(gc.ctx, gc.cfg)
}

// newListCommand creates the list command
func (gc *GlobalCommand) newListCommand() *cobra.Command {
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
  glid g list                      # List all worktrees
  glid g ls                        # Short alias
  glid g list --format json        # JSON output`,
		RunE: gc.executeList,
	}
	
	// Add flags
	cmd.Flags().String("format", "table", "Output format (table or json)")
	
	return cmd
}

// newCleanCommand creates the clean command
func (gc *GlobalCommand) newCleanCommand() *cobra.Command {
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
  glid g clean                    # Interactive cleanup
  glid g clean --orphaned         # Remove orphaned containers
  glid g clean --all              # Full cleanup
  glid g clean --dry-run          # Preview cleanup`,
		RunE: gc.executeClean,
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

func (gc *GlobalCommand) executeStatus(cmd *cobra.Command, args []string) error {
	return ExecuteGlobalStatus(gc.ctx, gc.cfg, cmd, args)
}

func (gc *GlobalCommand) executeDown(cmd *cobra.Command, args []string) error {
	return ExecuteGlobalDown(gc.ctx, gc.cfg, cmd, args)
}

func (gc *GlobalCommand) executeList(cmd *cobra.Command, args []string) error {
	return ExecuteGlobalList(gc.ctx, gc.cfg, cmd, args)
}

func (gc *GlobalCommand) executeClean(cmd *cobra.Command, args []string) error {
	return ExecuteGlobalClean(gc.ctx, gc.cfg, cmd, args)
}