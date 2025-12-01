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

// VersionCommand handles the version command
type VersionCommand struct {
	ctx *internalContext.ProjectContext
	cfg *config.Config
}

// VersionData represents version information for structured output
type VersionData struct {
	Version      string `json:"version" yaml:"version"`
	GitCommit    string `json:"git_commit" yaml:"git_commit"`
	GoVersion    string `json:"go_version" yaml:"go_version"`
	OS           string `json:"os" yaml:"os"`
	Architecture string `json:"architecture" yaml:"architecture"`
	BuildTime    string `json:"build_time" yaml:"build_time"`
	Compiler     string `json:"compiler" yaml:"compiler"`
}

// NewVersionCommand creates a new version command
func NewVersionCommand(ctx *internalContext.ProjectContext, cfg *config.Config) *cobra.Command {
	vc := &VersionCommand{
		ctx: ctx,
		cfg: cfg,
	}

	var checkUpdate bool

	cmd := &cobra.Command{
		Use:   "version [flags]",
		Short: "Display version information",
		Long: `Display version information for Glide CLI including build details and system information.

This command shows the current version of Glide along with:
- Go version used to build the binary
- Operating system and architecture
- Build time and compiler information
- Optional update availability check

The output format can be controlled using the global --format flag.

Examples:
  glide version                    # Show version information
  glide version --check-update     # Check for available updates
  glide version --format json      # Output as JSON
  glide version --format yaml      # Output as YAML`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return vc.execute(cmd, args, checkUpdate)
		},
	}

	// Add flags
	cmd.Flags().BoolVar(&checkUpdate, "check-update", false, "Check for available updates")

	return cmd
}

// execute runs the version command
func (vc *VersionCommand) execute(cmd *cobra.Command, args []string, checkUpdate bool) error {
	buildInfo := version.GetBuildInfo()

	// Create structured data for output
	// TODO: Use this when we have proper injection
	// versionData := VersionData{
	// 	Version:      buildInfo.Version,
	// 	GitCommit:    buildInfo.GitCommit,
	// 	GoVersion:    buildInfo.GoVersion,
	// 	OS:           buildInfo.OS,
	// 	Architecture: buildInfo.Architecture,
	// 	BuildTime:    buildInfo.BuildDate,
	// 	Compiler:     buildInfo.Compiler,
	// }

	// For now, just show text output until we have proper injection
	// TODO: Get format from injected manager once commands are migrated

	// For table/plain output, display formatted text
	output.Info(version.GetVersionString())
	output.Raw("\n")
	output.Raw("Build Information:\n")
	output.Raw(fmt.Sprintf("  Git Commit:    %s\n", buildInfo.GitCommit))
	output.Raw(fmt.Sprintf("  Build Time:    %s\n", buildInfo.BuildDate))
	output.Raw(fmt.Sprintf("  Go Version:    %s\n", buildInfo.GoVersion))
	output.Raw(fmt.Sprintf("  OS:            %s\n", buildInfo.OS))
	output.Raw(fmt.Sprintf("  Architecture:  %s\n", buildInfo.Architecture))
	output.Raw(fmt.Sprintf("  Compiler:      %s\n", buildInfo.Compiler))

	// Check for updates if requested
	if checkUpdate {
		output.Raw("\n")
		output.Info("Checking for updates...")

		checker := update.NewChecker(buildInfo.Version)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		updateInfo, err := checker.CheckForUpdate(ctx)
		if err != nil {
			output.Warning(fmt.Sprintf("Failed to check for updates: %v", err))
		} else {
			output.Raw("\n")
			output.Raw(update.FormatUpdateMessage(updateInfo))
		}
	}

	return nil
}
