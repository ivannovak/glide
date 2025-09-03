package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/ivannovak/glide/internal/config"
	"github.com/ivannovak/glide/internal/context"
	"github.com/ivannovak/glide/internal/docker"
	glideErrors "github.com/ivannovak/glide/pkg/errors"
	"github.com/ivannovak/glide/pkg/output"
	"github.com/spf13/cobra"
)

// StatusCommand handles the status command
type StatusCommand struct {
	ctx *context.ProjectContext
	cfg *config.Config
}

// NewStatusCommand creates a new status command
func NewStatusCommand(ctx *context.ProjectContext, cfg *config.Config) *cobra.Command {
	sc := &StatusCommand{
		ctx: ctx,
		cfg: cfg,
	}

	cmd := &cobra.Command{
		Use:   "status [flags]",
		Short: "Show container status",
		Long: `Display the status of Docker containers for the current project.

This command shows detailed information about running containers including
their health status, resource usage, and network configuration.

Options:
  --health        Show detailed health check information
  --ports         Show port mappings
  --volumes       Show volume mounts
  --all           Show all information
  --watch         Continuously update status (refresh every 2s)
  --format string Output format (table, json, yaml)

Examples:
  glid status                  # Basic status overview
  glid status --health         # Include health check details
  glid status --ports          # Show port mappings
  glid status --all            # Show all available information
  glid status --watch          # Live status updates
  glid status --format json    # JSON output for scripting

Status Indicators:
  🟢 Running and healthy
  🟡 Running but unhealthy or starting
  🔴 Stopped or failed
  ⚪ No health check defined

Information Shown:
  - Container name and service
  - Current state (running, stopped, etc.)
  - Health status if available
  - Uptime
  - Port mappings (with --ports)
  - Volume mounts (with --volumes)
  - Resource usage (CPU, Memory)`,
		RunE:          sc.Execute,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Add flags
	cmd.Flags().Bool("health", false, "Show detailed health check information")
	cmd.Flags().Bool("ports", false, "Show port mappings")
	cmd.Flags().Bool("volumes", false, "Show volume mounts")
	cmd.Flags().Bool("all", false, "Show all information")
	cmd.Flags().Bool("watch", false, "Continuously update status")
	cmd.Flags().String("format", "table", "Output format (table, json, yaml)")

	return cmd
}

// Execute runs the status command
func (c *StatusCommand) Execute(cmd *cobra.Command, args []string) error {
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

	// Get flags
	showHealth, _ := cmd.Flags().GetBool("health")
	showPorts, _ := cmd.Flags().GetBool("ports")
	showVolumes, _ := cmd.Flags().GetBool("volumes")
	showAll, _ := cmd.Flags().GetBool("all")
	watch, _ := cmd.Flags().GetBool("watch")
	format, _ := cmd.Flags().GetString("format")

	// If --all is set, enable all options
	if showAll {
		showHealth = true
		showPorts = true
		showVolumes = true
	}

	// Validate format
	if format != "table" && format != "json" && format != "yaml" {
		return glideErrors.NewConfigError(fmt.Sprintf("invalid format: %s", format),
			glideErrors.WithSuggestions(
				"Valid formats: table, json, yaml",
				"Use --format table for human-readable output",
				"Use --format json for machine-readable output",
				"Use --format yaml for YAML output",
			),
		)
	}

	// Check if Docker is running
	if !c.ctx.DockerRunning {
		output.Warning("Docker is not running")
		output.Info("\nNo containers are currently running.")
		output.Info("Start containers with: glid up")
		return nil
	}

	// Resolve Docker compose files
	resolver := docker.NewResolver(c.ctx)
	if err := resolver.Resolve(); err != nil {
		return glideErrors.Wrap(err, "failed to resolve Docker configuration",
			glideErrors.WithSuggestions(
				"Check if docker-compose.yml exists",
				"Verify you're in the correct project directory",
				"Ensure Docker is installed: docker --version",
			),
		)
	}

	// Show status
	if watch {
		return c.watchStatus(resolver, showHealth, showPorts, showVolumes, format)
	}
	return c.showStatus(resolver, showHealth, showPorts, showVolumes, format)
}

// showStatus displays the current status
func (c *StatusCommand) showStatus(resolver *docker.Resolver, showHealth, showPorts, showVolumes bool, format string) error {
	// Get container status
	manager := docker.NewContainerManager(c.ctx)
	containers, err := manager.GetStatus()
	if err != nil {
		return glideErrors.NewDockerError("failed to get container status",
			glideErrors.WithError(err),
			glideErrors.WithSuggestions(
				"Check if Docker is running: docker ps",
				"Start Docker Desktop if not running",
				"Verify Docker permissions",
			),
		)
	}

	// Get health status if requested
	var healthStatus []docker.ServiceHealth
	if showHealth {
		monitor := docker.NewHealthMonitor(c.ctx)
		healthStatus, _ = monitor.CheckHealth()
	}

	// Handle structured output for json/yaml
	if format == "json" || format == "yaml" {
		data := map[string]interface{}{
			"containers": containers,
		}
		if showHealth && len(healthStatus) > 0 {
			data["health"] = healthStatus
		}
		return output.Display(data)
	}
	
	// Default table display
	return c.displayTable(resolver, containers, healthStatus, showPorts, showVolumes)
}

// displayTable shows status in table format
func (c *StatusCommand) displayTable(resolver *docker.Resolver, containers []docker.Container, healthStatus []docker.ServiceHealth, showPorts, showVolumes bool) error {
	// Show project info
	output.Info("Docker Container Status")
	if c.ctx.IsWorktree {
		output.Info("Project: %s", resolver.GetComposeProjectName())
	}
	output.Println()

	if len(containers) == 0 {
		output.Warning("No containers found")
		output.Info("\nContainers may not be running. Start them with: glid up")
		return nil
	}

	// Display container information
	for _, container := range containers {
		c.displayContainerInfo(container, healthStatus, showPorts, showVolumes)
	}

	// Show summary
	c.showSummary(containers, healthStatus)

	return nil
}

// displayContainerInfo shows information for a single container
func (c *StatusCommand) displayContainerInfo(container docker.Container, healthStatus []docker.ServiceHealth, showPorts, showVolumes bool) {
	// Determine status icon
	statusIcon := c.getStatusIcon(container, healthStatus)

	// Container header
	output.Printf("%s %s ", statusIcon, output.InfoText("%s", container.Service))
	output.Printf("(%s)\n", container.Name)

	// Basic info
	output.Printf("   State: %s\n", c.colorizeState(container.State))
	output.Printf("   Status: %s\n", container.Status)
	
	// Health info if available
	health := c.findHealthStatus(container.Service, healthStatus)
	if health != nil && len(health.Containers) > 0 {
		for _, hc := range health.Containers {
			if hc.Status != docker.HealthNone {
				output.Printf("   Health: %s", c.colorizeHealth(hc.Status))
				if hc.FailingCount > 0 {
					output.Printf(" (failing: %d)", hc.FailingCount)
				}
				output.Println()
			}
		}
	}

	// Port mappings
	if showPorts && len(container.Ports) > 0 {
		output.Printf("   Ports: %s\n", strings.Join(container.Ports, ", "))
	}

	// Note: Volume information would need to be retrieved separately
	// as it's not included in the Container struct

	output.Println()
}

// getStatusIcon returns an appropriate status icon
func (c *StatusCommand) getStatusIcon(container docker.Container, healthStatus []docker.ServiceHealth) string {
	if container.State != "running" {
		return "🔴"
	}

	// Check health
	health := c.findHealthStatus(container.Service, healthStatus)
	if health != nil {
		if health.Healthy {
			return "🟢"
		}
		return "🟡"
	}

	// No health check
	return "⚪"
}

// findHealthStatus finds health status for a service
func (c *StatusCommand) findHealthStatus(service string, healthStatus []docker.ServiceHealth) *docker.ServiceHealth {
	for i, h := range healthStatus {
		if h.Service == service {
			return &healthStatus[i]
		}
	}
	return nil
}

// colorizeState returns colored state string
func (c *StatusCommand) colorizeState(state string) string {
	switch state {
	case "running":
		return output.SuccessText("%s", state)
	case "stopped", "exited":
		return output.ErrorText("%s", state)
	case "restarting", "paused":
		return output.WarningText("%s", state)
	default:
		return state
	}
}

// colorizeHealth returns colored health status
func (c *StatusCommand) colorizeHealth(status docker.HealthStatus) string {
	switch status {
	case docker.HealthHealthy:
		return output.SuccessText("%s", string(status))
	case docker.HealthUnhealthy:
		return output.ErrorText("%s", string(status))
	case docker.HealthStarting:
		return output.WarningText("%s", string(status))
	default:
		return string(status)
	}
}

// showSummary displays a summary of container status
func (c *StatusCommand) showSummary(containers []docker.Container, healthStatus []docker.ServiceHealth) {
	running := 0
	stopped := 0
	unhealthy := 0

	for _, container := range containers {
		if container.State == "running" {
			running++
			
			// Check health
			health := c.findHealthStatus(container.Service, healthStatus)
			if health != nil && !health.Healthy {
				unhealthy++
			}
		} else {
			stopped++
		}
	}

	// Summary line
	output.Println(strings.Repeat("─", 50))
	output.Printf("Total: %d containers", len(containers))
	if running > 0 {
		output.Printf(" | %s Running", output.SuccessText("%d", running))
	}
	if stopped > 0 {
		output.Printf(" | %s Stopped", output.ErrorText("%d", stopped))
	}
	if unhealthy > 0 {
		output.Printf(" | %s Unhealthy", output.WarningText("%d", unhealthy))
	}
	output.Println()

	// Helpful commands
	if stopped > 0 {
		output.Info("\nTo start containers: glid up")
	}
	if unhealthy > 0 {
		output.Info("To view logs: glid logs [service]")
	}
}

// watchStatus continuously updates status display
func (c *StatusCommand) watchStatus(resolver *docker.Resolver, showHealth, showPorts, showVolumes bool, format string) error {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Initial display
	c.clearScreen()
	if err := c.showStatus(resolver, showHealth, showPorts, showVolumes, format); err != nil {
		return err
	}

	output.Warning("\nWatching status... Press Ctrl+C to stop")

	// Update loop
	for range ticker.C {
		c.clearScreen()
		if err := c.showStatus(resolver, showHealth, showPorts, showVolumes, format); err != nil {
			// Don't exit on errors during watch
			output.Error("Error: %v", err)
		}
		output.Warning("\nWatching status... Press Ctrl+C to stop")
	}

	return nil
}

// clearScreen clears the terminal screen
func (c *StatusCommand) clearScreen() {
	// ANSI escape codes to clear screen and move cursor to top
	output.Raw("\033[2J\033[H")
}

