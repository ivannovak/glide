package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/glide-cli/glide/v3/internal/config"
	"github.com/glide-cli/glide/v3/internal/context"
	glideErrors "github.com/glide-cli/glide/v3/pkg/errors"
	"github.com/glide-cli/glide/v3/pkg/output"
	"github.com/spf13/cobra"
)

// WorktreeCommand handles the worktree management command
type WorktreeCommand struct {
	ctx *context.ProjectContext
	cfg *config.Config
}

// NewWorktreeCommand creates a new worktree command
func NewWorktreeCommand(ctx *context.ProjectContext, cfg *config.Config) *cobra.Command {
	wc := &WorktreeCommand{
		ctx: ctx,
		cfg: cfg,
	}

	// This is the implementation that will be called from global.go
	return wc.createCommand()
}

// createCommand creates the worktree command
func (c *WorktreeCommand) createCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "worktree [branch-name]",
		Short: "Create and manage worktrees",
		Long: `Create and manage Git worktrees for parallel development.

This command creates a new worktree for feature development.
Each worktree is an independent checkout of your repository
with its own branch and working directory.

Arguments:
  branch-name   Name of the branch (e.g., feature/user-auth)

Options:
  --from        Base branch or commit (default: main)
  --no-env      Don't copy .env file from vcs/

Examples:
  glide g worktree feature/api                    # Create from main
  glide g worktree fix/bug-123 --from develop     # Create from develop
  glide g worktree feature/ui --no-env            # Create without copying .env

Workflow:
  1. Creates worktree in worktrees/[branch-name]
  2. Copies .env from vcs/ (unless --no-env)`,
		RunE:          c.Execute,
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Add flags
	cmd.Flags().String("from", "main", "Base branch or commit")
	cmd.Flags().Bool("no-env", false, "Don't copy .env file")

	return cmd
}

// Execute runs the worktree command
func (c *WorktreeCommand) Execute(cmd *cobra.Command, args []string) error {
	// Validate we're in multi-worktree mode
	if err := ValidateMultiWorktreeMode(c.ctx, "worktree"); err != nil {
		return err
	}

	// Get branch name
	branchName := args[0]

	// Get flags
	fromBranch, _ := cmd.Flags().GetString("from")
	noEnv, _ := cmd.Flags().GetBool("no-env")

	// Display header
	output.Info("🌳 Creating Worktree: %s", branchName)
	output.Println(strings.Repeat("=", 40))
	output.Println()

	// Determine paths
	vcsDir := filepath.Join(c.ctx.ProjectRoot, "vcs")
	worktreesDir := filepath.Join(c.ctx.ProjectRoot, "worktrees")
	worktreeName := c.sanitizeName(branchName)
	worktreePath := filepath.Join(worktreesDir, worktreeName)

	// Check if worktree already exists
	if _, err := os.Stat(worktreePath); err == nil {
		return glideErrors.NewConfigError(fmt.Sprintf("worktree already exists at %s", worktreePath),
			glideErrors.WithSuggestions(
				"Remove the existing worktree: git worktree remove "+worktreePath,
				"Choose a different branch name",
				"Use 'git worktree list' to see all worktrees",
			),
		)
	}

	// Create worktrees directory if it doesn't exist
	if err := os.MkdirAll(worktreesDir, 0755); err != nil {
		return glideErrors.NewPermissionError(worktreesDir, "failed to create worktrees directory",
			glideErrors.WithError(err),
			glideErrors.WithSuggestions(
				"Check directory permissions: ls -la "+filepath.Dir(worktreesDir),
				"Ensure parent directory exists and is writable",
				"Run with appropriate permissions",
			),
		)
	}

	// Fetch latest changes
	if err := c.fetchLatest(vcsDir); err != nil {
		return err
	}

	// Check if this is a remote branch
	remoteBranch := ""
	if fromBranch != "" && fromBranch != "main" && fromBranch != "master" {
		// Check if it's a remote branch reference
		if strings.HasPrefix(fromBranch, "origin/") {
			remoteBranch = fromBranch
		}
	}

	// Create the worktree
	if err := c.createWorktree(vcsDir, worktreePath, branchName, fromBranch, remoteBranch); err != nil {
		return err
	}

	output.Success("✅ Worktree created successfully!")
	output.Println()

	// Copy .env file unless --no-env
	if !noEnv {
		if err := c.copyEnvFile(vcsDir, worktreePath); err != nil {
			output.Warning("⚠️  Warning: %v", err)
		}
	}

	// Show summary
	c.showSummary(worktreePath, branchName, remoteBranch)

	return nil
}

// sanitizeName converts branch name to directory-safe name
func (c *WorktreeCommand) sanitizeName(name string) string {
	// Replace non-alphanumeric characters with hyphens
	result := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return '-'
	}, name)

	// Remove leading/trailing hyphens
	result = strings.Trim(result, "-")

	// Collapse multiple hyphens
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}

	return result
}

// fetchLatest fetches the latest changes from origin
func (c *WorktreeCommand) fetchLatest(vcsDir string) error {
	output.Printf("📥 Fetching latest changes... ")

	cmd := exec.Command("git", "fetch", "origin")
	cmd.Dir = vcsDir

	if cmdOutput, err := cmd.CombinedOutput(); err != nil {
		output.Println()
		return glideErrors.NewNetworkError("failed to fetch latest changes from origin",
			glideErrors.WithError(err),
			glideErrors.WithContext("output", string(cmdOutput)),
			glideErrors.WithSuggestions(
				"Check network connectivity: ping github.com",
				"Verify git remote configuration: git remote -v",
				"Ensure you have proper authentication",
				"Try running: git fetch origin manually",
			),
		)
	}

	output.Success("✓")
	return nil
}

// createWorktree creates the actual Git worktree
func (c *WorktreeCommand) createWorktree(vcsDir, worktreePath, branchName, fromBranch, remoteBranch string) error {
	var cmd *exec.Cmd

	if remoteBranch != "" {
		// Checkout existing remote branch
		output.Warning("🔗 Creating worktree from remote branch: %s", remoteBranch)

		// Check if remote branch exists
		checkCmd := exec.Command("git", "ls-remote", "--heads", "origin", strings.TrimPrefix(remoteBranch, "origin/"))
		checkCmd.Dir = vcsDir
		if err := checkCmd.Run(); err != nil {
			return glideErrors.NewConfigError(fmt.Sprintf("remote branch %s does not exist", remoteBranch),
				glideErrors.WithError(err),
				glideErrors.WithSuggestions(
					"Check available remote branches: git branch -r",
					"Fetch latest changes: git fetch origin",
					"Verify branch name spelling",
					"Create the branch first on remote if needed",
				),
			)
		}

		cmd = exec.Command("git", "worktree", "add", worktreePath, "-b", branchName, remoteBranch)
	} else {
		// Check if local branch already exists
		checkCmd := exec.Command("git", "show-ref", "--verify", "--quiet", fmt.Sprintf("refs/heads/%s", branchName))
		checkCmd.Dir = vcsDir

		if checkCmd.Run() == nil {
			// Branch exists locally
			output.Warning("📌 Using existing local branch: %s", branchName)
			cmd = exec.Command("git", "worktree", "add", worktreePath, branchName)
		} else {
			// Create new branch
			output.Warning("✨ Creating new branch: %s from %s", branchName, fromBranch)
			cmd = exec.Command("git", "worktree", "add", worktreePath, "-b", branchName, fromBranch)
		}
	}

	cmd.Dir = vcsDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return glideErrors.NewCommandError("git worktree add", 1,
			glideErrors.WithError(err),
			glideErrors.WithContext("output", string(output)),
			glideErrors.WithSuggestions(
				"Check if the branch name is valid",
				"Ensure the worktree path doesn't already exist",
				"Verify git repository state: git status",
				"Try cleaning up: git worktree prune",
			),
		)
	}

	return nil
}

// copyEnvFile copies the .env file from vcs to the worktree
func (c *WorktreeCommand) copyEnvFile(vcsDir, worktreePath string) error {
	envSource := filepath.Join(vcsDir, ".env")
	envDest := filepath.Join(worktreePath, ".env")

	// Check if source exists
	if _, err := os.Stat(envSource); os.IsNotExist(err) {
		return glideErrors.NewFileNotFoundError(envSource,
			glideErrors.WithSuggestions(
				"Create a .env file in the vcs directory",
				"Copy .env.example to .env if available",
				"Use --no-env flag to skip copying .env file",
			),
		)
	}

	output.Printf("📋 Copying .env file... ")

	// Read source file
	data, err := os.ReadFile(envSource)
	if err != nil {
		output.Println()
		return glideErrors.NewPermissionError(envSource, "failed to read .env file",
			glideErrors.WithError(err),
			glideErrors.WithSuggestions(
				"Check file permissions: ls -la "+envSource,
				"Ensure you have read access to the file",
				"Verify the file is not corrupted",
			),
		)
	}

	// Write to destination
	if err := os.WriteFile(envDest, data, 0644); err != nil {
		output.Println()
		return glideErrors.NewPermissionError(envDest, "failed to write .env file",
			glideErrors.WithError(err),
			glideErrors.WithSuggestions(
				"Check destination directory permissions: ls -la "+filepath.Dir(envDest),
				"Ensure the directory exists and is writable",
				"Check available disk space",
			),
		)
	}

	output.Success("✓")
	return nil
}

// showSummary displays the completion summary
func (c *WorktreeCommand) showSummary(worktreePath, branchName, remoteBranch string) {
	output.Println()
	output.Success("🎉 Worktree Creation Complete!")
	output.Println()
	output.Printf("📍 Location: %s\n", worktreePath)
	output.Printf("🌿 Branch: %s\n", branchName)
	if remoteBranch != "" {
		output.Printf("🔗 Tracking: %s\n", remoteBranch)
	}
	output.Println()

	output.Info("📝 Next steps:")
	output.Printf("  cd %s\n", worktreePath)
	output.Println("  glide up                    # Start Docker containers")
	output.Println("  glide artisan migrate       # Run migrations")
	output.Println()
	output.Info("Happy coding! 🚀")
}
