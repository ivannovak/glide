package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ivannovak/glide/v2/internal/config"
	"github.com/ivannovak/glide/v2/internal/context"
	"github.com/ivannovak/glide/v2/internal/docker"
	glideErrors "github.com/ivannovak/glide/v2/pkg/errors"
	"github.com/ivannovak/glide/v2/pkg/output"
	"github.com/ivannovak/glide/v2/pkg/progress"
	"github.com/spf13/cobra"
)

// ProjectDownCommand handles the project down command
type ProjectDownCommand struct {
	ctx *context.ProjectContext
	cfg *config.Config
}

// ExecuteProjectDown is called from global.go
func ExecuteProjectDown(ctx *context.ProjectContext, cfg *config.Config, cmd *cobra.Command, args []string) error {
	gdc := &ProjectDownCommand{
		ctx: ctx,
		cfg: cfg,
	}
	return gdc.Execute(cmd, args)
}

// Execute runs the project down command
func (c *ProjectDownCommand) Execute(cmd *cobra.Command, args []string) error {
	// Validate we're in multi-worktree mode
	if err := ValidateMultiWorktreeMode(c.ctx, "down"); err != nil {
		return err
	}

	// Get flags
	removeOrphans, _ := cmd.Flags().GetBool("remove-orphans")
	removeVolumes, _ := cmd.Flags().GetBool("volumes")

	// Confirm if removing volumes
	if removeVolumes {
		output.Warning("⚠️  Warning: --volumes will delete all Docker volumes (data loss!)")
		output.Printf("Are you sure you want to continue? [y/N]: ")

		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			output.Info("Aborted.")
			return nil
		}
	}

	// Display header
	output.Info("🛑 Stopping Docker Containers Across All Worktrees")
	output.Println(strings.Repeat("=", 50))
	output.Println()

	// Track results
	successCount := 0
	failureCount := 0
	errors := []string{}

	// Stop containers in main repository (vcs/)
	vcsDir := filepath.Join(c.ctx.ProjectRoot, "vcs")
	if _, err := os.Stat(vcsDir); err == nil {
		output.Printf("📍 Main Repository (vcs/): ")
		if err := c.stopContainers(vcsDir, removeOrphans, removeVolumes); err != nil {
			output.Error("❌ Failed")
			errors = append(errors, fmt.Sprintf("vcs: %v", err))
			failureCount++
		} else {
			output.Success("✅ Stopped")
			successCount++
		}
	}

	// Stop containers in all worktrees
	worktreesDir := filepath.Join(c.ctx.ProjectRoot, "worktrees")
	if entries, err := os.ReadDir(worktreesDir); err == nil && len(entries) > 0 {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			worktreePath := filepath.Join(worktreesDir, entry.Name())

			// Check if it's a valid worktree
			gitFile := filepath.Join(worktreePath, ".git")
			if _, err := os.Stat(gitFile); err != nil {
				continue
			}

			output.Printf("📍 Worktree %s: ", entry.Name())

			if err := c.stopContainers(worktreePath, removeOrphans, removeVolumes); err != nil {
				// Check if it's just "no containers" error
				if strings.Contains(err.Error(), "no containers") || strings.Contains(err.Error(), "no such service") {
					output.Warning("⚠️  No containers")
				} else {
					output.Error("❌ Failed")
					errors = append(errors, fmt.Sprintf("%s: %v", entry.Name(), err))
					failureCount++
				}
			} else {
				output.Success("✅ Stopped")
				successCount++
			}
		}
	}

	// Summary
	output.Println()
	output.Println(strings.Repeat("-", 50))

	if failureCount == 0 {
		output.Success("✅ Successfully stopped containers in %d location(s)", successCount)
		if removeVolumes {
			output.Warning("⚠️  Volumes have been removed")
		}
	} else {
		output.Warning("⚠️  Stopped containers in %d location(s), failed in %d", successCount, failureCount)
		if len(errors) > 0 {
			output.Println("\nErrors encountered:")
			for _, err := range errors {
				output.Printf("  - %s\n", err)
			}
		}
	}

	return nil
}

// stopContainers stops Docker containers in a directory
func (c *ProjectDownCommand) stopContainers(dir string, removeOrphans bool, removeVolumes bool) error {
	// Create a context for this directory
	dirCtx := &context.ProjectContext{
		WorkingDir:      dir,
		ProjectRoot:     c.ctx.ProjectRoot,
		DevelopmentMode: c.ctx.DevelopmentMode,
		DockerRunning:   c.ctx.DockerRunning,
	}

	// Resolve Docker compose files
	resolver := docker.NewResolver(dirCtx)
	if err := resolver.Resolve(); err != nil {
		// No docker-compose.yml found, skip silently
		return nil
	}

	// Build docker-compose down command
	downArgs := []string{"down"}
	if removeOrphans {
		downArgs = append(downArgs, "--remove-orphans")
	}
	if removeVolumes {
		downArgs = append(downArgs, "--volumes")
	}

	// Get compose command
	args := resolver.GetComposeCommand(downArgs...)

	// Execute docker compose down
	execCmd := exec.Command("docker", args...)
	execCmd.Dir = dir

	// Run with progress indicator
	spinner := progress.NewSpinner("Stopping containers")
	spinner.Start()

	output, err := execCmd.CombinedOutput()

	if err != nil {
		spinner.Error("Failed")
		// Check if it's just "no containers" error
		outputStr := string(output)
		if strings.Contains(outputStr, "no containers") || strings.Contains(outputStr, "No such service") {
			return glideErrors.NewDockerError("no containers running in this worktree",
				glideErrors.WithExitCode(0), // Not really an error
			)
		}
		return glideErrors.Wrap(err, "docker-compose down failed",
			glideErrors.WithSuggestions(
				"Check Docker daemon status: docker ps",
				"Try stopping manually: docker compose down",
				"Check for permission issues",
			),
		)
	}

	spinner.Success("Stopped")
	return nil
}
