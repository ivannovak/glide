package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/ivannovak/glide/v2/internal/config"
	"github.com/ivannovak/glide/v2/internal/context"
	"github.com/ivannovak/glide/v2/internal/docker"
	"github.com/ivannovak/glide/v2/internal/shell"
	glideErrors "github.com/ivannovak/glide/v2/pkg/errors"
	"github.com/ivannovak/glide/v2/pkg/output"
	"github.com/ivannovak/glide/v2/pkg/progress"
	"github.com/spf13/cobra"
)

// CLI represents the main CLI application
type CLI struct {
	outputManager  *output.Manager
	projectContext *context.ProjectContext
	config         *config.Config
	builder        *Builder
}

// New creates a new CLI instance with dependencies
func New(
	outputManager *output.Manager,
	projectContext *context.ProjectContext,
	cfg *config.Config,
) *CLI {
	return &CLI{
		outputManager:  outputManager,
		projectContext: projectContext,
		config:         cfg,
		builder:        NewBuilder(projectContext, cfg, outputManager),
	}
}

// BuildRootCommand creates the root command with all subcommands
func (c *CLI) BuildRootCommand() *cobra.Command {
	return c.builder.Build()
}

// NewSetupCommand creates the setup command
func (c *CLI) NewSetupCommand() *cobra.Command {
	return NewSetupCommand(c.projectContext, c.config)
}

// NewConfigCommand creates the config command
func (c *CLI) NewConfigCommand() *cobra.Command {
	return NewConfigCommand(c.config)
}

// NewCompletionCommand creates the completion command
func (c *CLI) NewCompletionCommand() *cobra.Command {
	return NewCompletionCommand(c.projectContext, c.config)
}

// NewProjectCommand creates the global command
func (c *CLI) NewProjectCommand() *cobra.Command {
	return NewProjectCommand(c.projectContext, c.config)
}

// AddProjectCommands adds global commands to the provided command
func (c *CLI) AddProjectCommands(cmd *cobra.Command) {
	// Add all project subcommands to the parent project command
	globalCmd := NewProjectCommand(c.projectContext, c.config)

	// Add each subcommand directly to the provided command
	for _, subCmd := range globalCmd.Commands() {
		cmd.AddCommand(subCmd)
	}
}

// RegisterCompletions registers completion functions for all commands
func (c *CLI) RegisterCompletions(rootCmd *cobra.Command) {
	completionManager := NewCompletionManager(c.projectContext, c.config)
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.showContext(cmd)
		},
	})

	// Add shell test command (debug)
	cmd.AddCommand(&cobra.Command{
		Use:          "shell-test",
		Short:        "Test shell execution (debug)",
		SilenceUsage: true,
		Hidden:       true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.testShell(cmd, args)
		},
	})

	// Add docker resolution test command (debug)
	cmd.AddCommand(&cobra.Command{
		Use:          "docker-test",
		Short:        "Test Docker compose resolution (debug)",
		SilenceUsage: true,
		Hidden:       true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.testDockerResolution(cmd, args)
		},
	})

	// Add container management test command (debug)
	cmd.AddCommand(&cobra.Command{
		Use:          "container-test",
		Short:        "Test Docker container management (debug)",
		SilenceUsage: true,
		Hidden:       true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.testContainerManagement(cmd, args)
		},
	})
}

// showContext displays the detected project context
func (c *CLI) showContext(cmd *cobra.Command) error {
	if c.projectContext == nil {
		cmd.Println("No project context available")
		return nil
	}
	ctx := c.projectContext

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

	return nil
}

// showConfig displays the loaded configuration
func (c *CLI) showConfig(cmd *cobra.Command) {
	cmd.Println("=== Configuration ===")

	if c.config == nil {
		cmd.Println("No configuration loaded")
		return
	}

	cfg := c.config

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
func (c *CLI) testShell(cmd *cobra.Command, args []string) error {
	cmd.Println("=== Shell Execution Test ===")

	executor := shell.NewExecutor(shell.Options{
		Verbose: true,
	})

	// Test 1: Simple command capture
	cmd.Println("Test 1: Capture output")
	capturedOutput, err := executor.RunCapture("echo", "Hello from Glide!")
	if err != nil {
		_ = c.outputManager.Error("Failed: %v", err)
		return fmt.Errorf("test 1 failed: %w", err)
	}
	_ = c.outputManager.Success("Success: %s", capturedOutput)

	// Test 2: Command with timeout
	cmd.Println("\nTest 2: Command with timeout")
	err = executor.RunWithTimeout(2*time.Second, "sleep", "1")
	if err != nil {
		_ = c.outputManager.Error("Failed: %v", err)
		return fmt.Errorf("test 2 failed: %w", err)
	}
	_ = c.outputManager.Success("Success: Command completed within timeout")

	// Test 3: Progress indicator
	cmd.Println("\nTest 3: Progress indicator")
	spinner := progress.NewSpinner("Running test command")
	spinner.Start()
	time.Sleep(2 * time.Second)
	spinner.Success("Test completed")

	// Test 4: Docker check
	if c.projectContext != nil && len(c.projectContext.ComposeFiles) > 0 {
		cmd.Println("\nTest 4: Docker integration")
		docker := shell.NewDockerExecutor(c.projectContext)
		if docker.IsRunning() {
			_ = c.outputManager.Success("Docker is running")

			status, err := docker.GetContainerStatus()
			if err == nil && len(status) > 0 {
				cmd.Println("Container status:")
				for name, state := range status {
					cmd.Printf("  %s: %s\n", name, state)
				}
			}
		} else {
			_ = c.outputManager.Warning("Docker is not running")
		}
	}

	// Test 5: Pass-through (if args provided)
	if len(args) > 0 {
		cmd.Println("\nTest 5: Pass-through execution")
		cmd.Printf("Running: %v\n", args)

		shellCmd := shell.NewPassthroughCommand(args[0], args[1:]...)
		result, err := executor.Execute(shellCmd)
		if err != nil {
			_ = c.outputManager.Error("Failed: %v", err)
			return fmt.Errorf("test 5 failed: %w", err)
		}
		if result.ExitCode != 0 {
			_ = c.outputManager.Error("Command exited with code %d", result.ExitCode)
			return fmt.Errorf("test 5 command exited with code %d", result.ExitCode)
		}
		_ = c.outputManager.Success("Command succeeded")
	}

	return nil
}

// testDockerResolution tests the Docker compose file resolution
func (c *CLI) testDockerResolution(cmd *cobra.Command, args []string) error {
	cmd.Println("=== Docker Compose Resolution Test ===")

	if c.projectContext == nil {
		return fmt.Errorf("no project context available")
	}

	// Create Docker resolver
	resolver := docker.NewResolver(c.projectContext)

	// Perform resolution
	err := resolver.Resolve()
	if err != nil {
		_ = c.outputManager.Error("Resolution failed: %v", err)
		return fmt.Errorf("failed to resolve Docker compose files: %w", err)
	}

	// Display resolved compose files
	cmd.Println("Resolved Compose Files:")
	if len(c.projectContext.ComposeFiles) == 0 {
		_ = c.outputManager.Warning("  No compose files found")
	} else {
		for i, file := range c.projectContext.ComposeFiles {
			_ = c.outputManager.Success("  [%d] %s", i+1, file)
		}
	}

	// Show relative paths
	cmd.Println("\nRelative Paths:")
	relFiles := resolver.GetRelativeComposeFiles()
	for i, file := range relFiles {
		cmd.Printf("  [%d] %s\n", i+1, file)
	}

	// Show Docker status
	cmd.Printf("\nDocker Daemon Running: %v\n", c.projectContext.DockerRunning)

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
		_ = c.outputManager.Warning("  Validation warning: %v", err)
	} else {
		_ = c.outputManager.Success("  Docker setup validated successfully")
	}

	// Show override file if present
	if override := resolver.GetOverrideFile(); override != "" {
		cmd.Printf("\nOverride File: %s\n", override)
	}

	// Test different modes
	if c.projectContext.DevelopmentMode == context.ModeMultiWorktree {
		cmd.Println("\nTesting Single-Repo Mode Resolution:")
		tempResolver := docker.NewResolver(&context.ProjectContext{
			ProjectRoot:     c.projectContext.ProjectRoot,
			WorkingDir:      c.projectContext.WorkingDir,
			DevelopmentMode: context.ModeSingleRepo,
			Location:        context.LocationProject,
		})
		if err := tempResolver.Resolve(); err == nil {
			cmd.Printf("  Would use files: %v\n", tempResolver.GetRelativeComposeFiles())
		}
	}

	return nil
}

// testContainerManagement tests Docker container management functionality
func (c *CLI) testContainerManagement(cmd *cobra.Command, args []string) error {
	cmd.Println("=== Docker Container Management Test ===")

	if c.projectContext == nil || len(c.projectContext.ComposeFiles) == 0 {
		return fmt.Errorf("no Docker compose files found in this project")
	}

	// Create container manager and health monitor
	manager := docker.NewContainerManager(c.projectContext)
	health := docker.NewHealthMonitor(c.projectContext)
	errorHandler := docker.NewErrorHandler(true)

	// Test 1: Get container status
	cmd.Println("Test 1: Container Status")
	containers, err := manager.GetStatus()
	if err != nil {
		_ = c.outputManager.Error("  Error: %s", errorHandler.Handle(err))

		// Show suggestions
		suggestions := errorHandler.SuggestFix(err)
		if len(suggestions) > 0 {
			cmd.Println("\n  Suggestions:")
			for _, s := range suggestions {
				cmd.Printf("    - %s\n", s)
			}
		}
		return fmt.Errorf("failed to get container status: %w", err)
	}

	if len(containers) == 0 {
		_ = c.outputManager.Warning("  No containers found (containers may not be running)")
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
		_ = c.outputManager.Error("  Error checking health: %v", err)
		return fmt.Errorf("failed to check container health: %w", err)
	}

	for _, service := range healthStatus {
		if service.Healthy {
			_ = c.outputManager.Success("  ✓ %s: %s", service.Service, service.Summary)
		} else {
			_ = c.outputManager.Warning("  ⚠ %s: %s", service.Service, service.Summary)
		}
	}

	// Test 3: Check for orphaned containers
	cmd.Println("\nTest 3: Orphaned Containers")
	orphaned, err := manager.GetOrphanedContainers()
	if err != nil {
		_ = c.outputManager.Error("  Error checking orphaned containers: %v", err)
		return fmt.Errorf("failed to check orphaned containers: %w", err)
	}

	if len(orphaned) == 0 {
		_ = c.outputManager.Success("  No orphaned containers found")
	} else {
		_ = c.outputManager.Warning("  Found %d orphaned container(s):", len(orphaned))
		for _, container := range orphaned {
			cmd.Printf("    - %s (%s)\n", container.Name, container.Service)
		}
	}

	// Test 4: Get compose services
	cmd.Println("\nTest 4: Compose Services")
	services, err := manager.GetComposeServices()
	if err != nil {
		_ = c.outputManager.Error("  Error getting services: %v", err)
		return fmt.Errorf("failed to get compose services: %w", err)
	}

	cmd.Printf("  Defined services: %s\n", strings.Join(services, ", "))

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
		_ = c.outputManager.Success("  This error is retryable")
	} else {
		_ = c.outputManager.Error("  This error is not retryable")
	}

	return nil
}
