package cli

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/ivannovak/glide/internal/config"
	"github.com/ivannovak/glide/internal/context"
	glideErrors "github.com/ivannovak/glide/pkg/errors"
	"github.com/ivannovak/glide/pkg/output"
	"github.com/ivannovak/glide/pkg/progress"
	"github.com/ivannovak/glide/pkg/prompt"
	"github.com/spf13/cobra"
)

// ProjectCleanCommand handles the project clean command
type ProjectCleanCommand struct {
	ctx *context.ProjectContext
	cfg *config.Config
}

// CleanupStats tracks cleanup statistics
type CleanupStats struct {
	OrphanedContainers int
	DanglingImages     int
	UnusedVolumes      int
	UnusedNetworks     int
	SpaceReclaimed     string
}

// ExecuteProjectClean is called from global.go
func ExecuteProjectClean(ctx *context.ProjectContext, cfg *config.Config, cmd *cobra.Command, args []string) error {
	gcc := &ProjectCleanCommand{
		ctx: ctx,
		cfg: cfg,
	}
	return gcc.Execute(cmd, args)
}

// Execute runs the project clean command
func (c *ProjectCleanCommand) Execute(cmd *cobra.Command, args []string) error {
	// Validate we're in multi-worktree mode
	if err := ValidateMultiWorktreeMode(c.ctx, "clean"); err != nil {
		return err
	}

	// Get flags
	orphaned, _ := cmd.Flags().GetBool("orphaned")
	volumes, _ := cmd.Flags().GetBool("volumes")
	images, _ := cmd.Flags().GetBool("images")
	all, _ := cmd.Flags().GetBool("all")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// If --all, enable everything
	if all {
		orphaned = true
		volumes = true
		images = true
	}

	// If no specific flags, prompt for what to clean
	if !orphaned && !volumes && !images && !all {
		orphaned, volumes, images = c.promptForCleanup()
		if !orphaned && !volumes && !images {
			output.Info("No cleanup options selected. Exiting.")
			return nil
		}
	}

	// Display header
	output.Info("🧹 Global Docker Cleanup")
	output.Println(strings.Repeat("=", 50))
	if dryRun {
		output.Warning("DRY RUN MODE - No actual changes will be made")
	}
	output.Println()

	stats := &CleanupStats{}

	// Clean orphaned containers
	if orphaned {
		output.Printf("🔍 Checking for orphaned containers... ")
		if err := c.cleanOrphanedContainers(dryRun, stats); err != nil {
			output.Error("Failed: %v", err)
		} else if stats.OrphanedContainers > 0 {
			output.Success("Cleaned %d", stats.OrphanedContainers)
		} else {
			output.Success("None found")
		}
	}

	// Clean dangling images
	if images {
		output.Printf("🔍 Checking for dangling images... ")
		if err := c.cleanDanglingImages(dryRun, stats); err != nil {
			output.Error("Failed: %v", err)
		} else if stats.DanglingImages > 0 {
			output.Success("Cleaned %d", stats.DanglingImages)
		} else {
			output.Success("None found")
		}
	}

	// Clean unused volumes
	if volumes {
		output.Warning("⚠️  Warning: Cleaning volumes will delete data!")
		if !dryRun {
			output.Printf("Are you sure you want to continue? [y/N]: ")
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" {
				output.Info("Skipping volume cleanup.")
				volumes = false
			}
		}

		if volumes {
			output.Printf("🔍 Checking for unused volumes... ")
			if err := c.cleanUnusedVolumes(dryRun, stats); err != nil {
				output.Error("Failed: %v", err)
			} else if stats.UnusedVolumes > 0 {
				output.Success("Cleaned %d", stats.UnusedVolumes)
			} else {
				output.Success("None found")
			}
		}
	}

	// Clean unused networks
	output.Printf("🔍 Checking for unused networks... ")
	if err := c.cleanUnusedNetworks(dryRun, stats); err != nil {
		output.Error("Failed: %v", err)
	} else if stats.UnusedNetworks > 0 {
		output.Success("Cleaned %d", stats.UnusedNetworks)
	} else {
		output.Success("None found")
	}

	// Summary
	output.Println()
	output.Println(strings.Repeat("-", 50))

	if dryRun {
		output.Warning("🔍 Dry run complete. No changes were made.")
		c.displayCleanupSummary(stats)
		output.Println("\nTo perform actual cleanup, run without --dry-run")
	} else {
		output.Success("✅ Cleanup complete!")
		c.displayCleanupSummary(stats)

		if stats.SpaceReclaimed != "" && stats.SpaceReclaimed != "0B" {
			output.Success("💾 Space reclaimed: %s", stats.SpaceReclaimed)
		}
	}

	return nil
}

// promptForCleanup prompts the user for what to clean
func (c *ProjectCleanCommand) promptForCleanup() (orphaned, volumes, images bool) {
	output.Info("What would you like to clean?")
	output.Println()

	orphaned, _ = prompt.Confirm("Remove orphaned containers?", true)
	images, _ = prompt.Confirm("Remove dangling images?", true)

	// For volumes, use destructive confirmation since it can cause data loss
	if confirmed, _ := prompt.ConfirmDestructive("remove unused volumes"); confirmed {
		volumes = true
	}

	output.Println()
	return
}

// cleanOrphanedContainers removes orphaned containers
func (c *ProjectCleanCommand) cleanOrphanedContainers(dryRun bool, stats *CleanupStats) error {
	// Find stopped containers that match our project name pattern
	cmd := exec.Command("docker", "ps", "-a", "--filter", "status=exited", "--format", "{{.ID}} {{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var orphaned []string

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			name := parts[1]
			// Check if it's a project container
			if strings.Contains(name, "myproject") || strings.Contains(name, "glid") {
				orphaned = append(orphaned, parts[0])
			}
		}
	}

	if len(orphaned) > 0 {
		stats.OrphanedContainers = len(orphaned)

		if !dryRun {
			// Remove the containers
			args := append([]string{"rm", "-f"}, orphaned...)
			rmCmd := exec.Command("docker", args...)
			if err := rmCmd.Run(); err != nil {
				return glideErrors.Wrap(err, "failed to remove orphaned containers",
					glideErrors.WithSuggestions(
						"Check if containers are still running: docker ps",
						"Try removing manually: docker rm -f [container_ids]",
						"Check Docker permissions",
					),
				)
			}
		}
	}

	return nil
}

// cleanDanglingImages removes dangling images
func (c *ProjectCleanCommand) cleanDanglingImages(dryRun bool, stats *CleanupStats) error {
	// List dangling images
	cmd := exec.Command("docker", "images", "-f", "dangling=true", "-q")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	images := strings.Fields(strings.TrimSpace(string(output)))
	stats.DanglingImages = len(images)

	if len(images) > 0 && !dryRun {
		// Remove dangling images
		spinner := progress.NewSpinner("Removing dangling images")
		spinner.Start()

		pruneCmd := exec.Command("docker", "image", "prune", "-f")
		if output, err := pruneCmd.Output(); err != nil {
			spinner.Error("Failed")
			return err
		} else {
			// Parse space reclaimed
			outputStr := string(output)
			if strings.Contains(outputStr, "reclaimed") {
				lines := strings.Split(outputStr, "\n")
				for _, line := range lines {
					if strings.Contains(line, "reclaimed") {
						parts := strings.Fields(line)
						if len(parts) >= 3 {
							stats.SpaceReclaimed = parts[len(parts)-1]
						}
					}
				}
			}
			spinner.Success("Removed")
		}
	}

	return nil
}

// cleanUnusedVolumes removes unused volumes
func (c *ProjectCleanCommand) cleanUnusedVolumes(dryRun bool, stats *CleanupStats) error {
	// List unused volumes
	cmd := exec.Command("docker", "volume", "ls", "-qf", "dangling=true")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	volumes := strings.Fields(strings.TrimSpace(string(output)))
	stats.UnusedVolumes = len(volumes)

	if len(volumes) > 0 && !dryRun {
		// Remove unused volumes
		spinner := progress.NewSpinner("Removing unused volumes")
		spinner.Start()

		pruneCmd := exec.Command("docker", "volume", "prune", "-f")
		if err := pruneCmd.Run(); err != nil {
			spinner.Error("Failed")
			return err
		}
		spinner.Success("Removed")
	}

	return nil
}

// cleanUnusedNetworks removes unused networks
func (c *ProjectCleanCommand) cleanUnusedNetworks(dryRun bool, stats *CleanupStats) error {
	// List custom networks (not default ones)
	cmd := exec.Command("docker", "network", "ls", "--format", "{{.ID}} {{.Name}}")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var unused []string

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			name := parts[1]
			// Skip default networks
			if name != "bridge" && name != "host" && name != "none" {
				// Check if network is in use
				inspectCmd := exec.Command("docker", "network", "inspect", parts[0], "--format", "{{len .Containers}}")
				if inspectOut, err := inspectCmd.Output(); err == nil {
					if strings.TrimSpace(string(inspectOut)) == "0" {
						unused = append(unused, parts[0])
					}
				}
			}
		}
	}

	stats.UnusedNetworks = len(unused)

	if len(unused) > 0 && !dryRun {
		// Remove unused networks
		for _, networkID := range unused {
			rmCmd := exec.Command("docker", "network", "rm", networkID)
			rmCmd.Run() // Ignore errors for networks that might be in use
		}
	}

	return nil
}

// displayCleanupSummary displays a summary of what was cleaned
func (c *ProjectCleanCommand) displayCleanupSummary(stats *CleanupStats) {
	if stats.OrphanedContainers > 0 {
		output.Printf("  Orphaned containers: %d\n", stats.OrphanedContainers)
	}
	if stats.DanglingImages > 0 {
		output.Printf("  Dangling images: %d\n", stats.DanglingImages)
	}
	if stats.UnusedVolumes > 0 {
		output.Printf("  Unused volumes: %d\n", stats.UnusedVolumes)
	}
	if stats.UnusedNetworks > 0 {
		output.Printf("  Unused networks: %d\n", stats.UnusedNetworks)
	}

	total := stats.OrphanedContainers + stats.DanglingImages + stats.UnusedVolumes + stats.UnusedNetworks
	if total == 0 {
		output.Info("  Nothing to clean")
	}
}
