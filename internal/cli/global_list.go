package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ivannovak/glide/internal/config"
	"github.com/ivannovak/glide/internal/context"
	"github.com/ivannovak/glide/pkg/output"
	"github.com/spf13/cobra"
)

// GlobalListCommand handles the global list command
type GlobalListCommand struct {
	ctx *context.ProjectContext
	cfg *config.Config
}

// WorktreeInfo contains information about a worktree
type WorktreeInfo struct {
	Name         string
	Path         string
	Branch       string
	LastCommit   string
	CommitDate   time.Time
	IsClean      bool
	HasContainers bool
}

// ExecuteGlobalList is called from global.go
func ExecuteGlobalList(ctx *context.ProjectContext, cfg *config.Config, cmd *cobra.Command, args []string) error {
	glc := &GlobalListCommand{
		ctx: ctx,
		cfg: cfg,
	}
	return glc.Execute(cmd, args)
}

// Execute runs the global list command
func (c *GlobalListCommand) Execute(cmd *cobra.Command, args []string) error {
	// Validate we're in multi-worktree mode
	if err := ValidateMultiWorktreeMode(c.ctx, "list"); err != nil {
		return err
	}

	// Get format flag
	format, _ := cmd.Flags().GetString("format")

	// Display header
	if format != "json" {
		output.Info("📂 Active Git Worktrees")
		output.Println(strings.Repeat("=", 70))
		output.Println()
	}

	// Collect worktree information
	worktrees := []WorktreeInfo{}

	// Check main repository (vcs/)
	vcsDir := filepath.Join(c.ctx.ProjectRoot, "vcs")
	if _, err := os.Stat(vcsDir); err == nil {
		if info := c.getWorktreeInfo(vcsDir, "vcs"); info != nil {
			worktrees = append(worktrees, *info)
		}
	}

	// Check all worktrees
	worktreesDir := filepath.Join(c.ctx.ProjectRoot, "worktrees")
	if entries, err := os.ReadDir(worktreesDir); err == nil {
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

			if info := c.getWorktreeInfo(worktreePath, entry.Name()); info != nil {
				worktrees = append(worktrees, *info)
			}
		}
	}

	// Display results based on format
	if format == "json" {
		c.displayJSON(worktrees)
	} else {
		c.displayTable(worktrees)
	}

	// Summary
	if format != "json" {
		output.Println()
		output.Println(strings.Repeat("-", 70))
		output.Info("Total worktrees: %d", len(worktrees))
		
		// Count active containers
		activeCount := 0
		for _, w := range worktrees {
			if w.HasContainers {
				activeCount++
			}
		}
		if activeCount > 0 {
			output.Success("Active Docker environments: %d", activeCount)
		}
	}

	return nil
}

// getWorktreeInfo collects information about a worktree
func (c *GlobalListCommand) getWorktreeInfo(path string, name string) *WorktreeInfo {
	info := &WorktreeInfo{
		Name: name,
		Path: path,
	}

	// Get branch name
	branchCmd := exec.Command("git", "branch", "--show-current")
	branchCmd.Dir = path
	if output, err := branchCmd.Output(); err == nil {
		info.Branch = strings.TrimSpace(string(output))
	}

	// Get last commit info
	logCmd := exec.Command("git", "log", "-1", "--format=%h|%s|%ai")
	logCmd.Dir = path
	if output, err := logCmd.Output(); err == nil {
		parts := strings.Split(strings.TrimSpace(string(output)), "|")
		if len(parts) >= 3 {
			info.LastCommit = fmt.Sprintf("%s %s", parts[0], parts[1])
			if t, err := time.Parse("2006-01-02 15:04:05 -0700", parts[2]); err == nil {
				info.CommitDate = t
			}
		}
	}

	// Check if working directory is clean
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = path
	if output, err := statusCmd.Output(); err == nil {
		info.IsClean = len(strings.TrimSpace(string(output))) == 0
	}

	// Check for Docker containers
	composeFile := filepath.Join(path, "docker-compose.yml")
	if _, err := os.Stat(composeFile); err == nil {
		// Try to check container status
		psCmd := exec.Command("docker", "compose", "-f", composeFile, "ps", "-q")
		psCmd.Dir = path
		if output, err := psCmd.Output(); err == nil {
			info.HasContainers = len(strings.TrimSpace(string(output))) > 0
		}
	}

	return info
}

// displayTable displays worktrees in table format
func (c *GlobalListCommand) displayTable(worktrees []WorktreeInfo) {
	if len(worktrees) == 0 {
		output.Warning("No worktrees found")
		return
	}

	// Display each worktree
	for i, w := range worktrees {
		// Name and type
		if w.Name == "vcs" {
			output.Info("📍 Main Repository (vcs/)")
		} else {
			output.Info("📍 %s", w.Name)
		}

		// Branch
		if w.Branch != "" {
			output.Printf("   Branch: ")
			output.Success("%s", w.Branch)
			if !w.IsClean {
				output.Warning(" [modified]")
			}
			output.Println()
		}

		// Last commit
		if w.LastCommit != "" {
			output.Printf("   Commit: %s", w.LastCommit)
			if !w.CommitDate.IsZero() {
				output.Printf(" (%s)", c.formatRelativeTime(w.CommitDate))
			}
			output.Println()
		}

		// Docker status
		output.Printf("   Docker: ")
		if w.HasContainers {
			output.Success("🟢 Running")
		} else {
			output.Warning("⚪ Stopped")
		}
		output.Println()

		// Path
		output.Printf("   Path: %s\n", w.Path)

		if i < len(worktrees)-1 {
			output.Println()
		}
	}
}

// displayJSON displays worktrees in JSON format
func (c *GlobalListCommand) displayJSON(worktrees []WorktreeInfo) {
	output.Println("[")
	for i, w := range worktrees {
		output.Printf("  {\n")
		output.Printf("    \"name\": \"%s\",\n", w.Name)
		output.Printf("    \"path\": \"%s\",\n", w.Path)
		output.Printf("    \"branch\": \"%s\",\n", w.Branch)
		output.Printf("    \"last_commit\": \"%s\",\n", w.LastCommit)
		output.Printf("    \"is_clean\": %t,\n", w.IsClean)
		output.Printf("    \"has_containers\": %t\n", w.HasContainers)
		output.Printf("  }")
		if i < len(worktrees)-1 {
			output.Printf(",")
		}
		output.Println()
	}
	output.Println("]")
}

// formatRelativeTime formats a time as relative to now
func (c *GlobalListCommand) formatRelativeTime(t time.Time) string {
	duration := time.Since(t)
	
	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if duration < 30*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else if duration < 365*24*time.Hour {
		months := int(duration.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	}
	
	years := int(duration.Hours() / 24 / 365)
	if years == 1 {
		return "1 year ago"
	}
	return fmt.Sprintf("%d years ago", years)
}