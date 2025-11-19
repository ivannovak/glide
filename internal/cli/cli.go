package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/ivannovak/glide/internal/context"
	"github.com/ivannovak/glide/internal/docker"
	"github.com/ivannovak/glide/internal/shell"
	"github.com/ivannovak/glide/pkg/app"
	glideErrors "github.com/ivannovak/glide/pkg/errors"
	"github.com/ivannovak/glide/pkg/progress"
	"github.com/spf13/cobra"
)

// CLI represents the main CLI application
type CLI struct {
	app     *app.Application
	builder *Builder
}

// New creates a new CLI instance with an Application
func New(application *app.Application) *CLI {
	return &CLI{
		app:     application,
		builder: NewBuilder(application),
	}
}

// BuildRootCommand creates the root command with all subcommands
func (c *CLI) BuildRootCommand() *cobra.Command {
	return c.builder.Build()
}

// NewSetupCommand creates the setup command
func (c *CLI) NewSetupCommand() *cobra.Command {
	return NewSetupCommand(c.app.ProjectContext, c.app.Config)
}

// NewConfigCommand creates the config command
func (c *CLI) NewConfigCommand() *cobra.Command {
	return NewConfigCommand(c.app.Config)
}

// NewCompletionCommand creates the completion command
func (c *CLI) NewCompletionCommand() *cobra.Command {
	return NewCompletionCommand(c.app.ProjectContext, c.app.Config)
}

// NewProjectCommand creates the global command
func (c *CLI) NewProjectCommand() *cobra.Command {
	return NewProjectCommand(c.app.ProjectContext, c.app.Config)
}

// AddProjectCommands adds global commands to the provided command
func (c *CLI) AddProjectCommands(cmd *cobra.Command) {
	// Add all project subcommands to the parent project command
	globalCmd := NewProjectCommand(c.app.ProjectContext, c.app.Config)

	// Add each subcommand directly to the provided command
	for _, subCmd := range globalCmd.Commands() {
		cmd.AddCommand(subCmd)
	}
}

// RegisterCompletions registers completion functions for all commands
func (c *CLI) RegisterCompletions(rootCmd *cobra.Command) {
	completionManager := NewCompletionManager(c.app.ProjectContext, c.app.Config)
	completionManager.RegisterCommandCompletions(rootCmd)
}

// AddLocalCommands adds local commands to the provided command
func (c *CLI) AddLocalCommands(cmd *cobra.Command) {
	// Add debug commands
	c.addDebugCommands(cmd)

	// Load YAML commands from current directory
	c.builder.loadYAMLCommands()

	// Add all registered commands from the builder's registry
	// This ensures aliases are properly set
	for _, subCmd := range c.builder.registry.CreateAll() {
		cmd.AddCommand(subCmd)
	}
}

// addDebugCommands adds debug-only commands
func (c *CLI) addDebugCommands(cmd *cobra.Command) {
	// Add context debug command
	cmd.AddCommand(&cobra.Command{
		Use:          "context",
		Short:        "Show detected project context (debug)",
		SilenceUsage: true,
		Hidden:       true, // Hide debug commands
		Run: func(cmd *cobra.Command, args []string) {
			c.showContext(cmd)
		},
	})

	// Add shell test command (debug)
	cmd.AddCommand(&cobra.Command{
		Use:          "shell-test",
		Short:        "Test shell execution (debug)",
		SilenceUsage: true,
		Hidden:       true,
		Run: func(cmd *cobra.Command, args []string) {
			c.testShell(cmd, args)
		},
	})

	// Add docker resolution test command (debug)
	cmd.AddCommand(&cobra.Command{
		Use:          "docker-test",
		Short:        "Test Docker compose resolution (debug)",
		SilenceUsage: true,
		Hidden:       true,
		Run: func(cmd *cobra.Command, args []string) {
			c.testDockerResolution(cmd, args)
		},
	})

	// Add container management test command (debug)
	cmd.AddCommand(&cobra.Command{
		Use:          "container-test",
		Short:        "Test Docker container management (debug)",
		SilenceUsage: true,
		Hidden:       true,
		Run: func(cmd *cobra.Command, args []string) {
			c.testContainerManagement(cmd, args)
		},
	})
}

// showContext displays the detected project context
func (c *CLI) showContext(cmd *cobra.Command) {
	ctx := c.app.ProjectContext
	if ctx == nil {
		cmd.Println("No project context available")
		return
	}

	cmd.Println("=== Project Context ===")
	cmd.Printf("Working Directory: %s\n", ctx.WorkingDir)
	cmd.Printf("Project Root: %s\n", ctx.ProjectRoot)
	cmd.Printf("Development Mode: %s\n", ctx.DevelopmentMode)
	cmd.Printf("Location: %s\n", ctx.Location)

	if ctx.DevelopmentMode == context.ModeMultiWorktree {
		cmd.Printf("Is Root: %v\n", ctx.IsRoot)
		cmd.Printf("Is Main Repo: %v\n", ctx.IsMainRepo)
		cmd.Printf("Is Worktree: %v\n", ctx.IsWorktree)
		if ctx.WorktreeName != "" {
			cmd.Printf("Worktree Name: %s\n", ctx.WorktreeName)
		}
	}

	cmd.Printf("\nDocker Running: %v\n", ctx.DockerRunning)

	if len(ctx.ComposeFiles) > 0 {
		cmd.Println("\nCompose Files:")
		for _, file := range ctx.ComposeFiles {
			cmd.Printf("  - %s\n", file)
		}
	}

	if ctx.ComposeOverride != "" {
		cmd.Printf("Override File: %s\n", ctx.ComposeOverride)
	}

	// Display detected frameworks
	if len(ctx.DetectedFrameworks) > 0 {
		cmd.Println("\nDetected Frameworks:")
		for _, fw := range ctx.DetectedFrameworks {
			version := ""
			if v, exists := ctx.FrameworkVersions[fw]; exists && v != "" {
				version = fmt.Sprintf(" (v%s)", v)
			}
			cmd.Printf("  - %s%s\n", fw, version)
		}
	}

	// Display framework commands
	if len(ctx.FrameworkCommands) > 0 {
		cmd.Println("\nAvailable Framework Commands:")
		// Group commands by category if metadata available
		for name := range ctx.FrameworkCommands {
			cmd.Printf("  - %s\n", name)
		}
	}

	if ctx.Error != nil {
		cmd.Printf("\nContext Error: %v\n", ctx.Error)
	}
}

// showConfig displays the loaded configuration
func (c *CLI) showConfig(cmd *cobra.Command) {
	cmd.Println("=== Configuration ===")

	if c.app.Config == nil {
		cmd.Println("No configuration loaded")
		return
	}

	cfg := c.app.Config

	cmd.Println("\nProjects:")
	for name, project := range cfg.Projects {
		cmd.Printf("  %s:\n", name)
		cmd.Printf("    Path: %s\n", project.Path)
		cmd.Printf("    Mode: %s\n", project.Mode)
	}

	if cfg.DefaultProject != "" {
		cmd.Printf("\nDefault Project: %s\n", cfg.DefaultProject)
	}

	cmd.Println("\nDefaults:")
	cmd.Println("  Test:")
	cmd.Printf("    Parallel: %v\n", cfg.Defaults.Test.Parallel)
	cmd.Printf("    Processes: %d\n", cfg.Defaults.Test.Processes)
	cmd.Printf("    Coverage: %v\n", cfg.Defaults.Test.Coverage)
	cmd.Printf("    Verbose: %v\n", cfg.Defaults.Test.Verbose)

	cmd.Println("  Docker:")
	cmd.Printf("    Compose Timeout: %d\n", cfg.Defaults.Docker.ComposeTimeout)
	cmd.Printf("    Auto Start: %v\n", cfg.Defaults.Docker.AutoStart)
	cmd.Printf("    Remove Orphans: %v\n", cfg.Defaults.Docker.RemoveOrphans)

	cmd.Println("  Colors:")
	cmd.Printf("    Enabled: %s\n", cfg.Defaults.Colors.Enabled)

	cmd.Println("  Worktree:")
	cmd.Printf("    Auto Setup: %v\n", cfg.Defaults.Worktree.AutoSetup)
	cmd.Printf("    Copy Env: %v\n", cfg.Defaults.Worktree.CopyEnv)
	cmd.Printf("    Run Migrations: %v\n", cfg.Defaults.Worktree.RunMigrations)
}

// testShell tests the shell execution framework
func (c *CLI) testShell(cmd *cobra.Command, args []string) {
	cmd.Println("=== Shell Execution Test ===")

	executor := shell.NewExecutor(shell.Options{
		Verbose: true,
	})

	// Test 1: Simple command capture
	cmd.Println("Test 1: Capture output")
	capturedOutput, err := executor.RunCapture("echo", "Hello from Glide!")
	if err != nil {
		c.app.OutputManager.Error("Failed: %v", err)
	} else {
		c.app.OutputManager.Success("Success: %s", capturedOutput)
	}

	// Test 2: Command with timeout
	cmd.Println("\nTest 2: Command with timeout")
	err = executor.RunWithTimeout(2*time.Second, "sleep", "1")
	if err != nil {
		c.app.OutputManager.Error("Failed: %v", err)
	} else {
		c.app.OutputManager.Success("Success: Command completed within timeout")
	}

	// Test 3: Progress indicator
	cmd.Println("\nTest 3: Progress indicator")
	spinner := progress.NewSpinner("Running test command")
	spinner.Start()
	time.Sleep(2 * time.Second)
	spinner.Success("Test completed")

	// Test 4: Docker check
	ctx := c.app.ProjectContext
	if ctx != nil && len(ctx.ComposeFiles) > 0 {
		cmd.Println("\nTest 4: Docker integration")
		docker := shell.NewDockerExecutor(ctx)
		if docker.IsRunning() {
			c.app.OutputManager.Success("Docker is running")

			status, err := docker.GetContainerStatus()
			if err == nil && len(status) > 0 {
				cmd.Println("Container status:")
				for name, state := range status {
					cmd.Printf("  %s: %s\n", name, state)
				}
			}
		} else {
			c.app.OutputManager.Warning("Docker is not running")
		}
	}

	// Test 5: Pass-through (if args provided)
	if len(args) > 0 {
		cmd.Println("\nTest 5: Pass-through execution")
		cmd.Printf("Running: %v\n", args)

		shellCmd := shell.NewPassthroughCommand(args[0], args[1:]...)
		result, err := executor.Execute(shellCmd)
		if err != nil {
			c.app.OutputManager.Error("Failed: %v", err)
		} else if result.ExitCode != 0 {
			c.app.OutputManager.Error("Command exited with code %d", result.ExitCode)
		} else {
			c.app.OutputManager.Success("Command succeeded")
		}
	}
}

// testDockerResolution tests the Docker compose file resolution
func (c *CLI) testDockerResolution(cmd *cobra.Command, args []string) {
	cmd.Println("=== Docker Compose Resolution Test ===")

	ctx := c.app.ProjectContext
	if ctx == nil {
		c.app.OutputManager.Warning("No project context available")
		return
	}

	// Create Docker resolver
	resolver := docker.NewResolver(ctx)

	// Perform resolution
	err := resolver.Resolve()
	if err != nil {
		c.app.OutputManager.Error("Resolution failed: %v", err)
		return
	}

	// Display resolved compose files
	cmd.Println("Resolved Compose Files:")
	if len(ctx.ComposeFiles) == 0 {
		c.app.OutputManager.Warning("  No compose files found")
	} else {
		for i, file := range ctx.ComposeFiles {
			c.app.OutputManager.Success("  [%d] %s", i+1, file)
		}
	}

	// Show relative paths
	cmd.Println("\nRelative Paths:")
	relFiles := resolver.GetRelativeComposeFiles()
	for i, file := range relFiles {
		cmd.Printf("  [%d] %s\n", i+1, file)
	}

	// Show Docker status
	cmd.Printf("\nDocker Daemon Running: %v\n", ctx.DockerRunning)

	// Show project name
	cmd.Printf("Compose Project Name: %s\n", resolver.GetComposeProjectName())
	cmd.Printf("Docker Network: %s\n", resolver.GetDockerNetwork())

	// Test building a compose command
	cmd.Println("\nExample Docker Compose Command:")
	composeCmd := resolver.GetComposeCommand("ps")
	cmd.Printf("  docker %s\n", fmt.Sprintf("%v", composeCmd))

	// Validate setup
	cmd.Println("\nValidating Docker Setup:")
	if err := resolver.ValidateSetup(); err != nil {
		c.app.OutputManager.Warning("  Validation warning: %v", err)
	} else {
		c.app.OutputManager.Success("  Docker setup validated successfully")
	}

	// Show override file if present
	if override := resolver.GetOverrideFile(); override != "" {
		cmd.Printf("\nOverride File: %s\n", override)
	}

	// Test different modes
	if ctx.DevelopmentMode == context.ModeMultiWorktree {
		cmd.Println("\nTesting Single-Repo Mode Resolution:")
		tempResolver := docker.NewResolver(&context.ProjectContext{
			ProjectRoot:     ctx.ProjectRoot,
			WorkingDir:      ctx.WorkingDir,
			DevelopmentMode: context.ModeSingleRepo,
			Location:        context.LocationProject,
		})
		if err := tempResolver.Resolve(); err == nil {
			cmd.Printf("  Would use files: %v\n", tempResolver.GetRelativeComposeFiles())
		}
	}
}

// testContainerManagement tests Docker container management functionality
func (c *CLI) testContainerManagement(cmd *cobra.Command, args []string) {
	cmd.Println("=== Docker Container Management Test ===")

	ctx := c.app.ProjectContext
	if ctx == nil || len(ctx.ComposeFiles) == 0 {
		c.app.OutputManager.Warning("No Docker compose files found in this project")
		return
	}

	// Create container manager and health monitor
	manager := docker.NewContainerManager(ctx)
	health := docker.NewHealthMonitor(ctx)
	errorHandler := docker.NewErrorHandler(true)

	// Test 1: Get container status
	cmd.Println("Test 1: Container Status")
	containers, err := manager.GetStatus()
	if err != nil {
		c.app.OutputManager.Error("  Error: %s", errorHandler.Handle(err))

		// Show suggestions
		suggestions := errorHandler.SuggestFix(err)
		if len(suggestions) > 0 {
			cmd.Println("\n  Suggestions:")
			for _, s := range suggestions {
				cmd.Printf("    - %s\n", s)
			}
		}
		return
	}

	if len(containers) == 0 {
		c.app.OutputManager.Warning("  No containers found (containers may not be running)")
	} else {
		for _, container := range containers {
			status := "🔴"
			if container.State == "running" {
				status = "🟢"
			}
			cmd.Printf("  %s %s (%s): %s\n", status, container.Service, container.Name, container.Status)
		}
	}

	// Test 2: Health checks
	cmd.Println("\nTest 2: Health Checks")
	healthStatus, err := health.CheckHealth()
	if err != nil {
		c.app.OutputManager.Error("  Error checking health: %v", err)
	} else {
		for _, service := range healthStatus {
			if service.Healthy {
				c.app.OutputManager.Success("  ✓ %s: %s", service.Service, service.Summary)
			} else {
				c.app.OutputManager.Warning("  ⚠ %s: %s", service.Service, service.Summary)
			}
		}
	}

	// Test 3: Check for orphaned containers
	cmd.Println("\nTest 3: Orphaned Containers")
	orphaned, err := manager.GetOrphanedContainers()
	if err != nil {
		c.app.OutputManager.Error("  Error checking orphaned containers: %v", err)
	} else if len(orphaned) == 0 {
		c.app.OutputManager.Success("  No orphaned containers found")
	} else {
		c.app.OutputManager.Warning("  Found %d orphaned container(s):", len(orphaned))
		for _, container := range orphaned {
			cmd.Printf("    - %s (%s)\n", container.Name, container.Service)
		}
	}

	// Test 4: Get compose services
	cmd.Println("\nTest 4: Compose Services")
	services, err := manager.GetComposeServices()
	if err != nil {
		c.app.OutputManager.Error("  Error getting services: %v", err)
	} else {
		cmd.Printf("  Defined services: %s\n", strings.Join(services, ", "))
	}

	// Test 5: Error handling
	cmd.Println("\nTest 5: Error Handling")
	testErr := docker.ParseDockerError("test", "Cannot connect to the Docker daemon",
		glideErrors.NewDockerError("connection refused",
			glideErrors.WithSuggestions(
				"Start Docker Desktop application",
				"Check Docker daemon status",
			),
		))
	cmd.Printf("  Parsed error: %s\n", errorHandler.Handle(testErr))

	if docker.IsRetryable(testErr) {
		c.app.OutputManager.Success("  This error is retryable")
	} else {
		c.app.OutputManager.Error("  This error is not retryable")
	}
}
