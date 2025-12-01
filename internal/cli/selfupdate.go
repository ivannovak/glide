package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/ivannovak/glide/v3/internal/config"
	internalContext "github.com/ivannovak/glide/v3/internal/context"
	"github.com/ivannovak/glide/v3/pkg/output"
	"github.com/ivannovak/glide/v3/pkg/update"
	"github.com/ivannovak/glide/v3/pkg/version"
	"github.com/spf13/cobra"
)

// SelfUpdateCommand handles the self-update command
type SelfUpdateCommand struct {
	ctx *internalContext.ProjectContext
	cfg *config.Config
}

// NewSelfUpdateCommand creates a new self-update command
func NewSelfUpdateCommand(ctx *internalContext.ProjectContext, cfg *config.Config) *cobra.Command {
	suc := &SelfUpdateCommand{
		ctx: ctx,
		cfg: cfg,
	}

	var force bool

	cmd := &cobra.Command{
		Use:   "self-update [flags]",
		Short: "Update Glide CLI to the latest version",
		Long: `Update Glide CLI to the latest version available on GitHub.

This command will:
1. Check for the latest available version
2. Download the appropriate binary for your platform
3. Verify the download with SHA256 checksum (if available)
4. Replace the current binary with the new version
5. Create a backup of the current binary

The update process is atomic and will rollback on failure.

Examples:
  glide self-update              # Check and install updates
  glide self-update --force      # Force update even if already on latest`,
		Aliases:       []string{"update", "upgrade"},
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return suc.execute(cmd, args, force)
		},
	}

	// Add flags
	cmd.Flags().BoolVar(&force, "force", false, "Force update even if already on latest version")

	return cmd
}

// execute runs the self-update command
func (suc *SelfUpdateCommand) execute(cmd *cobra.Command, args []string, force bool) error {
	buildInfo := version.GetBuildInfo()
	currentVersion := buildInfo.Version

	// Don't allow self-update for development builds
	if currentVersion == "dev" {
		output.Error("Cannot self-update development builds")
		output.Info("Please use the install script or download a release binary")
		return fmt.Errorf("self-update not available for development builds")
	}

	output.Info(fmt.Sprintf("Current version: %s", currentVersion))
	output.Info("Checking for updates...")

	// Check for updates first
	checker := update.NewChecker(currentVersion)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	updateInfo, err := checker.CheckForUpdate(ctx)
	if err != nil {
		output.Error(fmt.Sprintf("Failed to check for updates: %v", err))
		return err
	}

	if !updateInfo.Available && !force {
		output.Success(fmt.Sprintf("You are already running the latest version (%s)", currentVersion))
		return nil
	}

	if force && !updateInfo.Available {
		output.Warning("Forcing reinstall of current version")
	} else {
		output.Info(fmt.Sprintf("New version available: %s", updateInfo.LatestVersion))
		output.Info(fmt.Sprintf("Release date: %s", updateInfo.PublishedAt.Format("2006-01-02")))
	}

	// Ask for confirmation
	output.Raw("\n")
	output.Warning("This will replace your current Glide binary.")
	output.Raw("Do you want to continue? (y/N): ")

	var response string
	fmt.Scanln(&response)
	if response != "y" && response != "Y" {
		output.Info("Update cancelled")
		return nil
	}

	// Perform the update
	output.Info("Downloading update...")

	updater := update.NewUpdater(currentVersion)
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel2()

	if err := updater.SelfUpdate(ctx2); err != nil {
		output.Error(fmt.Sprintf("Update failed: %v", err))
		output.Info("Your current binary has not been modified")
		return err
	}

	output.Success(fmt.Sprintf("Successfully updated to version %s", updateInfo.LatestVersion))
	output.Info("Please run 'glide version' to verify the update")

	return nil
}
