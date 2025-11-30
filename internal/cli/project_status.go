package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ivannovak/glide/v3/internal/config"
	"github.com/ivannovak/glide/v3/internal/context"
	"github.com/ivannovak/glide/v3/internal/docker"
	"github.com/ivannovak/glide/v3/pkg/output"
	"github.com/spf13/cobra"
)

// ProjectStatusCommand handles the project status command
type ProjectStatusCommand struct {
	ctx *context.ProjectContext
	cfg *config.Config
}

// ExecuteProjectStatus is called from project.go
func ExecuteProjectStatus(ctx *context.ProjectContext, cfg *config.Config, cmd *cobra.Command, args []string) error {
	psc := &ProjectStatusCommand{
		ctx: ctx,
		cfg: cfg,
	}
	return psc.Execute(cmd, args)
}

// Execute runs the project status command
func (c *ProjectStatusCommand) Execute(cmd *cobra.Command, args []string) error {
	// Validate we're in multi-worktree mode
	if err := ValidateMultiWorktreeMode(c.ctx, "status"); err != nil {
		return err
	}

	// Display header
	output.Info("🐳 Docker Status Across All Worktrees")
	output.Println(strings.Repeat("=", 50))
	output.Println()

	// Track if any containers are running
	hasRunningContainers := false

	// Check main repository (vcs/)
	vcsDir := filepath.Join(c.ctx.ProjectRoot, "vcs")
	if _, err := os.Stat(vcsDir); err == nil {
		output.Println("📍 Main Repository (vcs/):")
		status, running := c.getDockerStatus(vcsDir, "vcs")
		output.Printf("%s", status)
		if running {
			hasRunningContainers = true
		}
		output.Println()
	}

	// Check all worktrees
	worktreesDir := filepath.Join(c.ctx.ProjectRoot, "worktrees")
	if entries, err := os.ReadDir(worktreesDir); err == nil && len(entries) > 0 {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			worktreePath := filepath.Join(worktreesDir, entry.Name())

			// Check if it's a valid worktree (has .git file)
			gitFile := filepath.Join(worktreePath, ".git")
			if _, err := os.Stat(gitFile); err != nil {
				continue
			}

			// Get branch name
			branchName := c.getBranchName(worktreePath)

			output.Printf("📍 Worktree: %s", entry.Name())
			if branchName != "" {
				output.Info(" (%s)", branchName)
			}
			output.Println()

			status, running := c.getDockerStatus(worktreePath, entry.Name())
			output.Printf("%s", status)
			if running {
				hasRunningContainers = true
			}
			output.Println()
		}
	} else {
		output.Warning("  No worktrees found")
		output.Println()
	}

	// Summary
	output.Println(strings.Repeat("-", 50))
	if hasRunningContainers {
		output.Success("✅ Docker containers are running")
		output.Println("\nTo stop all containers, run:")
		output.Info("  glidep down")
	} else {
		output.Warning("⚠️  No Docker containers are currently running")
		output.Println("\nTo start containers in a worktree:")
		output.Println("  1. cd to the worktree directory")
		output.Println("  2. Run: glideup")
	}

	return nil
}

// getDockerStatus checks Docker status for a directory
func (c *ProjectStatusCommand) getDockerStatus(dir string, name string) (string, bool) {
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
		return output.WarningText("  ⚠️  No docker-compose.yml found\n"), false
	}

	// Get compose command
	args := resolver.GetComposeCommand("ps", "--format", "table")

	// Execute docker compose ps
	cmd := exec.Command("docker", args...)
	cmd.Dir = dir

	cmdOutput, err := cmd.Output()
	if err != nil {
		return output.ErrorText("  ❌ Error checking status: %v\n", err), false
	}

	outputStr := string(cmdOutput)
	lines := strings.Split(strings.TrimSpace(outputStr), "\n")

	// Check if any containers are running
	hasRunning := false
	var result strings.Builder

	if len(lines) <= 1 || (len(lines) == 2 && strings.Contains(lines[0], "NAME")) {
		result.WriteString(output.WarningText("  ⚠️  No containers\n"))
	} else {
		// Parse container status
		runningCount := 0
		stoppedCount := 0

		for i, line := range lines {
			if i == 0 || line == "" {
				continue // Skip header
			}

			if strings.Contains(line, "Up") || strings.Contains(line, "running") {
				runningCount++
				hasRunning = true
			} else if strings.Contains(line, "Exited") || strings.Contains(line, "stopped") {
				stoppedCount++
			}
		}

		if runningCount > 0 {
			result.WriteString(output.SuccessText("  🟢 %d running", runningCount))
		}
		if stoppedCount > 0 {
			if runningCount > 0 {
				result.WriteString(", ")
			} else {
				result.WriteString("  ")
			}
			result.WriteString(output.WarningText("🟡 %d stopped", stoppedCount))
		}
		result.WriteString("\n")

		// Show container details if verbose flag is set
		// Note: cmd here is the exec.Cmd, not cobra.Command
		// We'd need to pass the verbose flag from Execute method
	}

	return result.String(), hasRunning
}

// getBranchName gets the current branch name for a worktree
func (c *ProjectStatusCommand) getBranchName(worktreePath string) string {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = worktreePath

	if cmdOutput, err := cmd.Output(); err == nil {
		return strings.TrimSpace(string(cmdOutput))
	}

	return ""
}
