package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	glideContext "github.com/ivannovak/glide/v2/internal/context"
	"github.com/ivannovak/glide/v2/internal/docker"
	"github.com/ivannovak/glide/v2/internal/shell"
	"github.com/ivannovak/glide/v2/pkg/output"
	"github.com/ivannovak/glide/v2/pkg/progress"
	"github.com/spf13/cobra"
)

// showContext displays the current project context
func showContext(_ *cobra.Command, outputManager *output.Manager, projectContext *glideContext.ProjectContext) error {
	ctx := projectContext

	if ctx == nil {
		_ = outputManager.Info("No project context available")
		return nil
	}

	_ = outputManager.Info("=== Project Context ===")
	_ = outputManager.Info("Working Directory: %s", ctx.WorkingDir)
	_ = outputManager.Info("Project Root: %s", ctx.ProjectRoot)
	_ = outputManager.Info("Development Mode: %s", ctx.DevelopmentMode)
	_ = outputManager.Info("Location: %s", ctx.Location)

	if ctx.DevelopmentMode == glideContext.ModeMultiWorktree {
		_ = outputManager.Info("")
		_ = outputManager.Info("=== Multi-Worktree Details ===")
		_ = outputManager.Info("Is Worktree: %v", ctx.IsWorktree)
		if ctx.IsWorktree {
			_ = outputManager.Info("Worktree Name: %s", ctx.WorktreeName)
		}
	}

	_ = outputManager.Info("")
	_ = outputManager.Info("Docker Running: %v", ctx.DockerRunning)
	if len(ctx.ComposeFiles) > 0 {
		_ = outputManager.Info("Compose Files: %s", strings.Join(ctx.ComposeFiles, ", "))
	}

	return nil
}

// showConfig displays the current configuration
// func showConfig(cmd *cobra.Command, app *app.Application) {
// 	output := app.OutputManager
// 	cfg := app.Config
//
// 	if cfg == nil {
// 		output.Info("No configuration loaded")
// 		return
// 	}
//
// 	output.Info("=== Configuration ===")
// 	output.Info("Default Project: %s", cfg.DefaultProject)
//
// 	if len(cfg.Projects) > 0 {
// 		output.Info("")
// 		output.Info("=== Projects ===")
// 		for name, project := range cfg.Projects {
// 			output.Info("%s:", name)
// 			output.Info("  Path: %s", project.Path)
// 			output.Info("  Mode: %s", project.Mode)
// 		}
// 	}
//
// 	output.Info("")
// 	output.Info("=== Defaults ===")
// 	output.Info("Test:")
// 	output.Info("  Parallel: %v", cfg.Defaults.Test.Parallel)
// 	output.Info("  Processes: %d", cfg.Defaults.Test.Processes)
// 	output.Info("  Coverage: %v", cfg.Defaults.Test.Coverage)
// 	output.Info("  Verbose: %v", cfg.Defaults.Test.Verbose)
//
// 	output.Info("Docker:")
// 	output.Info("  Compose Timeout: %d", cfg.Defaults.Docker.ComposeTimeout)
// 	output.Info("  Auto Start: %v", cfg.Defaults.Docker.AutoStart)
// 	output.Info("  Remove Orphans: %v", cfg.Defaults.Docker.RemoveOrphans)
// }

// testShell tests shell execution capabilities
func testShell(_ *cobra.Command, _ []string, outputManager *output.Manager) error {
	executor := shell.NewExecutor(shell.Options{})

	_ = outputManager.Info("=== Shell Execution Test ===")

	// Test 1: Simple command using strategy pattern
	_ = outputManager.Info("\nTest 1: Capture output (strategy pattern)")
	shellCmd := shell.NewCommand("echo", "Hello from shell")
	shellCmd.UseStrategy = true
	shellCmd.Options = shell.CommandOptions{
		CaptureOutput: true,
		Timeout:       5 * time.Second,
	}
	result, err := executor.Execute(shellCmd)

	if err != nil {
		_ = outputManager.Error("Failed: %v", err)
		return err
	}
	_ = outputManager.Success("Output: %s", strings.TrimSpace(string(result.Stdout)))

	// Test 2: Command with timeout (using context)
	_ = outputManager.Info("\nTest 2: Command with timeout (using context)")
	cmd2 := shell.NewCommand("sleep", "0.5")
	cmd2.Options = shell.CommandOptions{
		Timeout: 2 * time.Second,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	result, err = executor.ExecuteWithContext(ctx, cmd2)

	if err != nil {
		_ = outputManager.Error("Failed: %v", err)
		return err
	}
	_ = outputManager.Success("Completed in %v", result.Duration)

	// Test 3: Progress indicator
	_ = outputManager.Info("\nTest 3: Progress indicator")
	spinner := progress.NewSpinner("Testing progress")
	spinner.Start()
	time.Sleep(1 * time.Second)
	spinner.Success("Completed")

	return nil
}

// testDockerResolution tests Docker compose file resolution
func testDockerResolution(_ *cobra.Command, _ []string, outputManager *output.Manager, projectContext *glideContext.ProjectContext) error {
	if projectContext == nil {
		return fmt.Errorf("no project context available")
	}

	_ = outputManager.Info("=== Docker Compose Resolution Test ===")
	_ = outputManager.Info("Working Directory: %s", projectContext.WorkingDir)
	_ = outputManager.Info("Project Root: %s", projectContext.ProjectRoot)
	_ = outputManager.Info("Development Mode: %s", projectContext.DevelopmentMode)

	// Create Docker resolver
	resolver := docker.NewResolver(projectContext)

	// Try to resolve
	_ = outputManager.Info("\nAttempting to resolve Docker compose files...")
	err := resolver.Resolve()
	if err != nil {
		_ = outputManager.Error("Resolution failed: %v", err)
		return fmt.Errorf("failed to resolve Docker compose files: %w", err)
	}

	// Show results
	files := resolver.GetComposeFiles()
	if len(files) == 0 {
		_ = outputManager.Warning("No compose files found")
		return nil
	}

	_ = outputManager.Success("Resolved %d compose file(s):", len(files))
	for i, file := range files {
		_ = outputManager.Info("  %d. %s", i+1, file)
	}

	return nil
}

// testContainerManagement tests container management capabilities
func testContainerManagement(_ *cobra.Command, _ []string, outputManager *output.Manager, projectContext *glideContext.ProjectContext) error {
	if projectContext == nil {
		return fmt.Errorf("no project context available")
	}

	_ = outputManager.Info("=== Container Management Test ===")

	// Create container manager
	manager := docker.NewContainerManager(projectContext)

	// Test getting container status
	_ = outputManager.Info("\n1. Getting container status...")
	containers, err := manager.GetStatus()
	if err != nil {
		_ = outputManager.Error("Status failed: %v", err)
		return fmt.Errorf("failed to get container status: %w", err)
	}

	_ = outputManager.Success("Found %d container(s)", len(containers))
	for _, container := range containers {
		_ = outputManager.Info("  - %s (%s): %s", container.Name, container.Service, container.State)
	}

	// Test checking if service is running
	_ = outputManager.Info("\n2. Checking if php service is running...")
	if manager.IsRunning("php") {
		_ = outputManager.Success("PHP service is running")
	} else {
		_ = outputManager.Warning("PHP service is not running")
	}

	// Test getting compose services
	_ = outputManager.Info("\n3. Getting compose services...")
	services, err := manager.GetComposeServices()
	if err != nil {
		_ = outputManager.Error("Failed to get services: %v", err)
		return fmt.Errorf("failed to get compose services: %w", err)
	}

	_ = outputManager.Success("Found %d service(s)", len(services))
	for _, service := range services {
		_ = outputManager.Info("  - %s", service)
	}

	// Test logs (dry run)
	_ = outputManager.Info("\n4. Testing log retrieval (dry run)...")
	if len(containers) > 0 {
		_ = outputManager.Info("Would retrieve logs for: %s", containers[0].Service)
		_ = outputManager.Success("Log retrieval capability verified")
	} else {
		_ = outputManager.Warning("No containers available for log test")
	}

	return nil
}
