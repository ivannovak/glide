package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ivannovak/glide/internal/config"
	"github.com/ivannovak/glide/internal/context"
	"github.com/ivannovak/glide/internal/docker"
	"github.com/ivannovak/glide/internal/shell"
	glideErrors "github.com/ivannovak/glide/pkg/errors"
	"github.com/ivannovak/glide/pkg/output"
	"github.com/ivannovak/glide/pkg/progress"
	"github.com/spf13/cobra"
)

// TestCommand handles the test command
type TestCommand struct {
	ctx *context.ProjectContext
	cfg *config.Config
}

// NewTestCommand creates a new test command
func NewTestCommand(ctx *context.ProjectContext, cfg *config.Config) *cobra.Command {
	tc := &TestCommand{
		ctx: ctx,
		cfg: cfg,
	}

	cmd := &cobra.Command{
		Use:   "test [flags] [-- pest arguments]",
		Short: "Run Pest tests with full argument pass-through",
		Long: `Run Pest tests in the Docker environment.

This command provides complete pass-through to Pest, meaning all arguments
after -- are passed directly to Pest without modification.

Examples:
  glid test                           # Run all tests with default settings
  glid test -- --filter UserTest      # Run specific test class
  glid test -- --group api            # Run tests in a specific group
  glid test -- --parallel             # Run tests in parallel
  glid test -- --coverage             # Run with code coverage
  glid test -- tests/Unit/UserTest.php # Run specific test file

Default Configuration:
  The test command uses your configuration defaults from ~/.glide.yml:
  - parallel: Run tests in parallel mode
  - processes: Number of parallel processes
  - coverage: Generate code coverage report
  - verbose: Show detailed output

To override defaults, pass arguments directly to Pest.`,
		DisableFlagParsing: true, // Pass all flags through to Pest
		RunE:               tc.Execute,
		SilenceUsage:       true,  // Don't show usage on error
		SilenceErrors:      true,  // Let our error handler handle errors
	}

	return cmd
}

// Execute runs the test command
func (c *TestCommand) Execute(cmd *cobra.Command, args []string) error {
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

	// Check if Docker is running
	if !c.ctx.DockerRunning {
		output.Warning("Docker is not running. Starting Docker containers...")
		if err := c.startDocker(); err != nil {
			return glideErrors.Wrap(err, "failed to start Docker",
				glideErrors.WithSuggestions(
					"Check Docker Desktop is installed",
					"Start Docker Desktop manually",
					"Run: glid up to start containers",
				),
			)
		}
	}

	// Check dependencies
	if err := c.checkDependencies(); err != nil {
		return err
	}

	// Setup test database if needed
	if err := c.setupTestDatabase(); err != nil {
		return glideErrors.Wrap(err, "failed to setup test database",
			glideErrors.WithSuggestions(
				"Check database connection in .env.testing",
				"Ensure MySQL container is running: glid status",
				"Run migrations manually: glid artisan migrate --env=testing",
			),
		)
	}

	// Build Pest command with arguments
	pestArgs := c.buildPestCommand(args)

	// Show what we're running
	output.Info("Running: pest %s", strings.Join(pestArgs, " "))

	// Execute Pest with progress indication
	return c.runPestWithProgress(pestArgs)
}

// checkDependencies verifies that required dependencies are available
func (c *TestCommand) checkDependencies() error {
	// Check if vendor directory exists
	vendorPath := filepath.Join(c.ctx.ProjectRoot, "vendor")
	if _, err := os.Stat(vendorPath); os.IsNotExist(err) {
		output.Warning("Dependencies not installed. Running composer install...")
		
		spinner := progress.NewSpinner("Installing dependencies")
		spinner.Start()
		
		// Run composer install
		resolver := docker.NewResolver(c.ctx)
		if err := resolver.Resolve(); err != nil {
			spinner.Error("Failed to resolve Docker configuration")
			return err
		}
		
		composeCmd := resolver.GetComposeCommand("exec", "php", "composer", "install")
		
		executor := shell.NewExecutor(shell.Options{
			Verbose: false,
		})
		
		shellCmd := shell.NewCommand("docker", composeCmd...)
		if _, err := executor.Execute(shellCmd); err != nil {
			spinner.Error("Failed to install dependencies")
			return glideErrors.NewDependencyError("composer", "composer install failed",
				glideErrors.WithError(err),
				glideErrors.WithSuggestions(
					"Check composer.json exists",
					"Ensure PHP container is running: glid status",
					"Try clearing composer cache: glid composer clear-cache",
					"Check network connectivity in Docker",
				),
			)
		}
		
		spinner.Success("Dependencies installed")
	}

	// Check if Pest is installed
	pestPath := filepath.Join(c.ctx.ProjectRoot, "vendor", "bin", "pest")
	if _, err := os.Stat(pestPath); os.IsNotExist(err) {
		return glideErrors.NewDependencyError("pestphp/pest", "Pest is not installed",
			glideErrors.WithSuggestions(
				"Install Pest: glid composer require pestphp/pest --dev",
				"Run: glid composer install",
				"Check composer.json includes pestphp/pest in require-dev",
			),
		)
	}

	return nil
}

// setupTestDatabase ensures the test database is ready
func (c *TestCommand) setupTestDatabase() error {
	// Check if we need to setup test database
	// This is determined by checking if APP_ENV is set to testing
	envPath := filepath.Join(c.ctx.ProjectRoot, ".env.testing")
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		// No .env.testing file, skip test database setup
		return nil
	}

	output.Info("Preparing test database...")
	
	resolver := docker.NewResolver(c.ctx)
	if err := resolver.Resolve(); err != nil {
		return err
	}

	executor := shell.NewExecutor(shell.Options{
		Verbose: false,
	})

	// Run migrations on test database
	composeCmd := resolver.GetComposeCommand(
		"exec",
		"-e", "APP_ENV=testing",
		"php",
		"php", "artisan", "migrate", "--force",
	)

	if output, err := executor.RunCapture("docker", composeCmd...); err != nil {
		// Check if it's just "nothing to migrate"
		if !strings.Contains(output, "Nothing to migrate") {
			return glideErrors.NewDatabaseError("failed to run test migrations",
				glideErrors.WithError(err),
				glideErrors.WithSuggestions(
					"Check .env.testing database configuration",
					"Ensure test database exists",
					"Run: glid artisan migrate:fresh --env=testing",
					"Check migration files for errors",
				),
			)
		}
	}

	return nil
}

// buildPestCommand constructs the Pest command with arguments
func (c *TestCommand) buildPestCommand(args []string) []string {
	pestArgs := []string{}

	// Check if user provided arguments
	if len(args) > 0 {
		// Pass through all user arguments directly
		pestArgs = args
	} else {
		// Use default configuration
		if c.cfg != nil && c.cfg.Defaults.Test.Parallel {
			pestArgs = append(pestArgs, "--parallel")
			if c.cfg.Defaults.Test.Processes > 0 {
				pestArgs = append(pestArgs, fmt.Sprintf("--processes=%d", c.cfg.Defaults.Test.Processes))
			}
		}

		if c.cfg != nil && c.cfg.Defaults.Test.Coverage {
			pestArgs = append(pestArgs, "--coverage")
		}

		if c.cfg != nil && c.cfg.Defaults.Test.Verbose {
			pestArgs = append(pestArgs, "-v")
		}
	}

	return pestArgs
}

// runPestWithProgress executes Pest with progress indication
func (c *TestCommand) runPestWithProgress(pestArgs []string) error {
	resolver := docker.NewResolver(c.ctx)
	if err := resolver.Resolve(); err != nil {
		return err
	}

	// Build the complete Docker command
	dockerArgs := []string{}
	dockerArgs = append(dockerArgs, resolver.GetComposeCommand("exec")...)
	
	// Add TTY allocation for interactive output
	dockerArgs = append(dockerArgs, "-T")
	
	// Add test environment if .env.testing exists
	envPath := filepath.Join(c.ctx.ProjectRoot, ".env.testing")
	if _, err := os.Stat(envPath); err == nil {
		dockerArgs = append(dockerArgs, "-e", "APP_ENV=testing")
	}
	
	dockerArgs = append(dockerArgs, "php", "./vendor/bin/pest")
	dockerArgs = append(dockerArgs, pestArgs...)

	// Create a pass-through command for real-time output
	shellCmd := shell.NewPassthroughCommand("docker", dockerArgs[1:]...)
	
	executor := shell.NewExecutor(shell.Options{
		Verbose: false,
	})

	// Execute with pass-through for real-time test output
	result, err := executor.Execute(shellCmd)
	if err != nil {
		return glideErrors.Wrap(err, "test execution failed",
			glideErrors.WithSuggestions(
				"Check test syntax for errors",
				"Ensure all dependencies are installed: glid composer install",
				"Verify Docker containers are healthy: glid status",
				"Run tests with verbose output: glid test -- -v",
			),
		)
	}

	// Check test results
	if result.ExitCode != 0 {
		output.Error("\n✗ Tests failed with exit code %d", result.ExitCode)
		return glideErrors.NewCommandError("pest", result.ExitCode,
			glideErrors.WithSuggestions(
				"Review the test output above for failure details",
				"Run specific test: glid test -- --filter TestName",
				"Run with verbose output: glid test -- -v",
				"Check test database state: glid artisan migrate:status --env=testing",
			),
		)
	}

	output.Success("\n✓ All tests passed!")
	return nil
}

// startDocker starts the Docker containers
func (c *TestCommand) startDocker() error {
	spinner := progress.NewSpinner("Starting Docker containers")
	spinner.Start()

	resolver := docker.NewResolver(c.ctx)
	if err := resolver.Resolve(); err != nil {
		spinner.Error("Failed to resolve Docker configuration")
		return err
	}

	executor := shell.NewExecutor(shell.Options{
		Verbose: false,
	})

	composeCmd := resolver.GetComposeCommand("up", "-d")
	shellCmd := shell.NewCommand("docker", composeCmd...)
	
	if _, err := executor.Execute(shellCmd); err != nil {
		spinner.Error("Failed to start Docker")
		return err
	}

	// Wait for containers to be healthy
	health := docker.NewHealthMonitor(c.ctx)
	if err := health.WaitForHealthy(30*time.Second); err != nil {
		spinner.Error("Docker containers failed health check")
		return err
	}

	spinner.Success("Docker containers started")
	
	// Update context to reflect Docker is now running
	c.ctx.DockerRunning = true
	
	return nil
}

// NewTestCommandCLI creates the test command for the CLI
func (c *CLI) NewTestCommand() *cobra.Command {
	return NewTestCommand(c.app.ProjectContext, c.app.Config)
}