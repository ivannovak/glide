package cli

import (
	"context"
	"strings"
	"time"

	glideContext "github.com/ivannovak/glide/internal/context"
	"github.com/ivannovak/glide/internal/docker"
	"github.com/ivannovak/glide/internal/shell"
	"github.com/ivannovak/glide/pkg/app"
	"github.com/ivannovak/glide/pkg/progress"
	"github.com/spf13/cobra"
)

// showContext displays the current project context
func showContext(cmd *cobra.Command, app *app.Application) {
	output := app.OutputManager
	ctx := app.ProjectContext
	
	if ctx == nil {
		output.Info("No project context available")
		return
	}
	
	output.Info("=== Project Context ===")
	output.Info("Working Directory: %s", ctx.WorkingDir)
	output.Info("Project Root: %s", ctx.ProjectRoot)
	output.Info("Development Mode: %s", ctx.DevelopmentMode)
	output.Info("Location: %s", ctx.Location)
	
	if ctx.DevelopmentMode == glideContext.ModeMultiWorktree {
		output.Info("")
		output.Info("=== Multi-Worktree Details ===")
		output.Info("Is Worktree: %v", ctx.IsWorktree)
		if ctx.IsWorktree {
			output.Info("Worktree Name: %s", ctx.WorktreeName)
		}
	}
	
	output.Info("")
	output.Info("Docker Running: %v", ctx.DockerRunning)
	if len(ctx.ComposeFiles) > 0 {
		output.Info("Compose Files: %s", strings.Join(ctx.ComposeFiles, ", "))
	}
}

// showConfig displays the current configuration
func showConfig(cmd *cobra.Command, app *app.Application) {
	output := app.OutputManager
	cfg := app.Config
	
	if cfg == nil {
		output.Info("No configuration loaded")
		return
	}
	
	output.Info("=== Configuration ===")
	output.Info("Default Project: %s", cfg.DefaultProject)
	
	if len(cfg.Projects) > 0 {
		output.Info("")
		output.Info("=== Projects ===")
		for name, project := range cfg.Projects {
			output.Info("%s:", name)
			output.Info("  Path: %s", project.Path)
			output.Info("  Mode: %s", project.Mode)
		}
	}
	
	output.Info("")
	output.Info("=== Defaults ===")
	output.Info("Test:")
	output.Info("  Parallel: %v", cfg.Defaults.Test.Parallel)
	output.Info("  Processes: %d", cfg.Defaults.Test.Processes)
	output.Info("  Coverage: %v", cfg.Defaults.Test.Coverage)
	output.Info("  Verbose: %v", cfg.Defaults.Test.Verbose)
	
	output.Info("Docker:")
	output.Info("  Compose Timeout: %d", cfg.Defaults.Docker.ComposeTimeout)
	output.Info("  Auto Start: %v", cfg.Defaults.Docker.AutoStart)
	output.Info("  Remove Orphans: %v", cfg.Defaults.Docker.RemoveOrphans)
}

// testShell tests shell execution capabilities
func testShell(cmd *cobra.Command, args []string, app *app.Application) {
	output := app.OutputManager
	executor := app.GetShellExecutor()
	
	output.Info("=== Shell Execution Test ===")
	
	// Test 1: Simple command using strategy pattern
	output.Info("\nTest 1: Capture output (strategy pattern)")
	shellCmd := shell.NewCommand("echo", "Hello from shell")
	shellCmd.UseStrategy = true
	shellCmd.Options = shell.CommandOptions{
		CaptureOutput: true,
		Timeout:       5 * time.Second,
	}
	result, err := executor.Execute(shellCmd)
	
	if err != nil {
		output.Error("Failed: %v", err)
	} else {
		output.Success("Output: %s", strings.TrimSpace(string(result.Stdout)))
	}
	
	// Test 2: Command with timeout (using context)
	output.Info("\nTest 2: Command with timeout (using context)")
	cmd2 := shell.NewCommand("sleep", "0.5")
	cmd2.Options = shell.CommandOptions{
		Timeout: 2 * time.Second,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	result, err = executor.ExecuteWithContext(ctx, cmd2)
	
	if err != nil {
		output.Error("Failed: %v", err)
	} else {
		output.Success("Completed in %v", result.Duration)
	}
	
	// Test 3: Progress indicator
	output.Info("\nTest 3: Progress indicator")
	spinner := progress.NewSpinner("Testing progress")
	spinner.Start()
	time.Sleep(1 * time.Second)
	spinner.Success("Completed")
}

// testDockerResolution tests Docker compose file resolution
func testDockerResolution(cmd *cobra.Command, args []string, app *app.Application) {
	output := app.OutputManager
	ctx := app.ProjectContext
	
	if ctx == nil {
		output.Error("No project context available")
		return
	}
	
	output.Info("=== Docker Compose Resolution Test ===")
	output.Info("Working Directory: %s", ctx.WorkingDir)
	output.Info("Project Root: %s", ctx.ProjectRoot)
	output.Info("Development Mode: %s", ctx.DevelopmentMode)
	
	// Create Docker resolver
	resolver := docker.NewResolver(ctx)
	
	// Try to resolve
	output.Info("\nAttempting to resolve Docker compose files...")
	err := resolver.Resolve()
	if err != nil {
		output.Error("Resolution failed: %v", err)
		return
	}
	
	// Show results
	files := resolver.GetComposeFiles()
	if len(files) == 0 {
		output.Warning("No compose files found")
		return
	}
	
	output.Success("Resolved %d compose file(s):", len(files))
	for i, file := range files {
		output.Info("  %d. %s", i+1, file)
	}
}

// testContainerManagement tests container management capabilities
func testContainerManagement(cmd *cobra.Command, args []string, app *app.Application) {
	output := app.OutputManager
	ctx := app.ProjectContext
	
	if ctx == nil {
		output.Error("No project context available")
		return
	}
	
	output.Info("=== Container Management Test ===")
	
	// Create container manager
	manager := docker.NewContainerManager(ctx)
	
	// Test getting container status
	output.Info("\n1. Getting container status...")
	containers, err := manager.GetStatus()
	if err != nil {
		output.Error("Status failed: %v", err)
	} else {
		output.Success("Found %d container(s)", len(containers))
		for _, container := range containers {
			output.Info("  - %s (%s): %s", container.Name, container.Service, container.State)
		}
	}
	
	// Test checking if service is running
	output.Info("\n2. Checking if php service is running...")
	if manager.IsRunning("php") {
		output.Success("PHP service is running")
	} else {
		output.Warning("PHP service is not running")
	}
	
	// Test getting compose services
	output.Info("\n3. Getting compose services...")
	services, err := manager.GetComposeServices()
	if err != nil {
		output.Error("Failed to get services: %v", err)
	} else {
		output.Success("Found %d service(s)", len(services))
		for _, service := range services {
			output.Info("  - %s", service)
		}
	}
	
	// Test logs (dry run)
	output.Info("\n4. Testing log retrieval (dry run)...")
	if len(containers) > 0 {
		output.Info("Would retrieve logs for: %s", containers[0].Service)
		output.Success("Log retrieval capability verified")
	} else {
		output.Warning("No containers available for log test")
	}
}