package cli

import (
	"fmt"
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

// ArtisanCommand handles the artisan command
type ArtisanCommand struct {
	ctx *context.ProjectContext
	cfg *config.Config
}

// NewArtisanCommand creates a new artisan command
func NewArtisanCommand(ctx *context.ProjectContext, cfg *config.Config) *cobra.Command {
	ac := &ArtisanCommand{
		ctx: ctx,
		cfg: cfg,
	}

	cmd := &cobra.Command{
		Use:   "artisan [artisan arguments]",
		Short: "Run Artisan commands via Docker",
		Long: `Execute Laravel Artisan commands within the Docker PHP container.

This command passes all arguments directly to Artisan running inside
the PHP container. It handles interactive commands, migrations, and
all other Artisan functionality.

Examples:
  glid artisan migrate                # Run database migrations
  glid artisan migrate:fresh --seed   # Fresh migration with seeding
  glid artisan make:controller UserController  # Create a controller
  glid artisan make:model User -m     # Create model with migration
  glid artisan tinker                 # Start interactive REPL
  glid artisan queue:work              # Start queue worker
  glid artisan cache:clear             # Clear application cache
  glid artisan route:list              # List all routes
  glid artisan db:seed                 # Run database seeders
  glid artisan horizon                # Start Horizon dashboard

Interactive Commands:
  Commands like 'tinker' and 'db' are automatically detected and run
  with proper TTY allocation for interactive use.

Migration Commands:
  glid artisan migrate                # Run pending migrations
  glid artisan migrate:rollback       # Rollback last migration batch
  glid artisan migrate:fresh          # Drop all tables and re-migrate
  glid artisan migrate:status         # Check migration status
  glid artisan migrate:reset          # Rollback all migrations

Cache Commands:
  glid artisan cache:clear             # Clear application cache
  glid artisan config:clear            # Clear configuration cache
  glid artisan route:clear             # Clear route cache
  glid artisan view:clear              # Clear compiled views
  glid artisan optimize:clear          # Clear all cached data

Queue Commands:
  glid artisan queue:work              # Process queue jobs
  glid artisan queue:listen            # Listen for new jobs
  glid artisan queue:failed            # List failed jobs
  glid artisan queue:retry all         # Retry all failed jobs

Environment:
  Commands are executed with the application environment from .env file.
  Use APP_ENV to control the environment (local, staging, production).`,
		DisableFlagParsing: true, // Pass all flags through to artisan
		RunE:               ac.Execute,
		SilenceUsage:       true,  // Don't show usage on error
		SilenceErrors:      true,  // Let our error handler handle errors
	}

	return cmd
}

// Execute runs the artisan command
func (c *ArtisanCommand) Execute(cmd *cobra.Command, args []string) error {
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

	// Resolve Docker compose files
	resolver := docker.NewResolver(c.ctx)
	if err := resolver.Resolve(); err != nil {
		return glideErrors.Wrap(err, "failed to resolve Docker configuration",
			glideErrors.WithSuggestions(
				"Check if docker-compose.yml exists",
				"Verify Docker is installed: docker --version",
			),
		)
	}

	// Build the artisan command
	artisanArgs := c.buildArtisanCommand(resolver, args)

	// Show what we're running for certain commands
	if c.shouldShowCommand(args) {
		output.Info("Running: php artisan %s", strings.Join(args, " "))
	}

	// Check if this needs special handling
	isInteractive := c.isInteractiveCommand(args)
	needsProgress := c.needsProgressIndicator(args)

	// Execute the command
	if needsProgress && !isInteractive {
		return c.executeWithProgress(artisanArgs, args)
	}
	return c.executeArtisanCommand(artisanArgs, isInteractive)
}

// buildArtisanCommand constructs the docker-compose exec command for artisan
func (c *ArtisanCommand) buildArtisanCommand(resolver *docker.Resolver, args []string) []string {
	// Start with docker-compose exec command
	dockerArgs := resolver.GetComposeCommand("exec")

	// For interactive commands, don't use -T flag
	if !c.isInteractiveCommand(args) {
		dockerArgs = append(dockerArgs, "-T")
	}

	// Add PHP container, php binary, and artisan
	dockerArgs = append(dockerArgs, "php", "php", "artisan")

	// Add all artisan arguments
	dockerArgs = append(dockerArgs, args...)

	return dockerArgs
}

// isInteractiveCommand checks if artisan command needs TTY
func (c *ArtisanCommand) isInteractiveCommand(args []string) bool {
	if len(args) == 0 {
		// No arguments means list commands, which is not interactive
		return false
	}

	// Interactive commands that need TTY
	interactiveCommands := []string{
		"tinker",      // Laravel REPL
		"db",          // Database CLI
		"shell",       // Shell access
	}

	// Check if the first argument is an interactive command
	for _, cmd := range interactiveCommands {
		if args[0] == cmd {
			return true
		}
	}

	// Check for interactive flags
	for _, arg := range args {
		if arg == "--interactive" || arg == "-i" {
			return true
		}
	}

	return false
}

// shouldShowCommand determines if we should display the command being run
func (c *ArtisanCommand) shouldShowCommand(args []string) bool {
	// Always show for migrations and important operations
	if len(args) > 0 {
		importantCommands := []string{
			"migrate", "migrate:fresh", "migrate:rollback", "migrate:reset",
			"db:seed", "queue:work", "queue:listen", "horizon",
			"cache:clear", "config:clear", "route:clear", "view:clear",
			"optimize", "optimize:clear",
		}

		for _, cmd := range importantCommands {
			if strings.HasPrefix(args[0], cmd) {
				return true
			}
		}
	}

	// Otherwise check verbose setting
	return c.cfg != nil && c.cfg.Defaults.Test.Verbose
}

// needsProgressIndicator checks if the command would benefit from progress indication
func (c *ArtisanCommand) needsProgressIndicator(args []string) bool {
	if len(args) == 0 {
		return false
	}

	// Commands that typically take time
	progressCommands := []string{
		"migrate", "migrate:fresh", "migrate:rollback", "migrate:reset",
		"db:seed",
		"queue:work", "queue:listen",
		"horizon",
		"ide-helper:generate", "ide-helper:models",
	}

	for _, cmd := range progressCommands {
		if strings.HasPrefix(args[0], cmd) {
			return true
		}
	}

	return false
}

// executeWithProgress runs artisan with a progress indicator
func (c *ArtisanCommand) executeWithProgress(dockerArgs []string, artisanArgs []string) error {
	// Determine the operation for the progress message
	operation := "Running artisan command"
	if len(artisanArgs) > 0 {
		operation = fmt.Sprintf("Running artisan %s", artisanArgs[0])
	}

	spinner := progress.NewSpinner(operation)
	spinner.Start()
	defer spinner.Stop()

	// Execute the command
	executor := shell.NewExecutor(shell.Options{
		Verbose: false,
	})

	// For migrations and seeds, we want to see the output
	shellCmd := shell.NewPassthroughCommand("docker", dockerArgs...)
	result, err := executor.Execute(shellCmd)

	if err != nil {
		spinner.Error(fmt.Sprintf("Artisan %s failed", artisanArgs[0]))
		return err
	}

	if result.ExitCode != 0 {
		// Check if this is an expected non-zero exit
		if !c.isExpectedNonZeroExit(artisanArgs, result.ExitCode) {
			spinner.Error(fmt.Sprintf("Artisan %s failed with exit code %d", artisanArgs[0], result.ExitCode))
			return glideErrors.NewCommandError("artisan command failed", result.ExitCode,
				glideErrors.WithSuggestions(
					"Check the command syntax",
					"Run with --verbose for more details",
					"Check Laravel logs: glid logs php",
				),
			)
		}
	}

	spinner.Success(fmt.Sprintf("Artisan %s completed", artisanArgs[0]))
	return nil
}

// executeArtisanCommand runs an artisan command
func (c *ArtisanCommand) executeArtisanCommand(dockerArgs []string, isInteractive bool) error {
	executor := shell.NewExecutor(shell.Options{
		Verbose: false,
	})

	// Use appropriate command type based on interactivity
	var shellCmd *shell.Command
	if isInteractive {
		shellCmd = shell.NewInteractiveCommand("docker", dockerArgs...)
	} else {
		shellCmd = shell.NewPassthroughCommand("docker", dockerArgs...)
	}

	result, err := executor.Execute(shellCmd)

	if err != nil {
		return c.handleArtisanError(err)
	}

	if result.ExitCode != 0 {
		// Don't error for commands that might legitimately return non-zero
		if !c.isExpectedNonZeroExit(dockerArgs, result.ExitCode) {
			return glideErrors.NewCommandError(fmt.Sprintf("artisan command exited with code %d", result.ExitCode), result.ExitCode,
				glideErrors.WithSuggestions(
					"Check the command syntax and arguments",
					"View Laravel logs for details: glid logs php",
					"Try running with --verbose flag",
				),
			)
		}
	}

	return nil
}

// handleArtisanError provides helpful error messages for common issues
func (c *ArtisanCommand) handleArtisanError(err error) error {
	errStr := err.Error()

	// PHP container not running
	if strings.Contains(errStr, "service \"php\" is not running") {
		output.Error("PHP container is not running")
		output.Warning("\nTo start the containers:")
		output.Warning("  glid up")
		output.Warning("  or")
		output.Warning("  glid docker up -d")
		return glideErrors.NewDockerError("PHP container is not running",
			glideErrors.WithSuggestions(
				"Start the containers: glid up",
				"Or use: glid docker up -d",
				"Check container status: glid docker ps",
			),
		)
	}

	// Database connection errors
	if strings.Contains(errStr, "SQLSTATE") || strings.Contains(errStr, "could not find driver") {
		output.Error("Database connection error")
		output.Warning("\nPossible solutions:")
		output.Warning("  - Check if MySQL container is running: glid docker ps")
		output.Warning("  - Verify database credentials in .env file")
		output.Warning("  - Ensure DB_HOST is set to 'mysql' for Docker")
		output.Warning("  - Run: glid docker up -d mysql")
		return glideErrors.NewDatabaseError("database connection failed",
			glideErrors.WithSuggestions(
				"Check if MySQL container is running: glid docker ps",
				"Verify database credentials in .env file",
				"Ensure DB_HOST is set to 'mysql' for Docker",
				"Start MySQL: glid docker up -d mysql",
			),
		)
	}

	// Migration errors
	if strings.Contains(errStr, "Nothing to migrate") {
		// This is not really an error
		output.Success("✓ No pending migrations")
		return nil
	}

	if strings.Contains(errStr, "Migration table not found") {
		output.Warning("Migration table doesn't exist. Run: glid artisan migrate:install")
		return glideErrors.NewDatabaseError("migration table not found",
			glideErrors.WithSuggestions(
				"Create migration table: glid artisan migrate:install",
				"Or run fresh migration: glid artisan migrate:fresh",
				"Check database connection",
			),
		)
	}

	// Permission errors
	if strings.Contains(errStr, "Permission denied") || strings.Contains(errStr, "permission denied") {
		output.Error("Permission denied")
		output.Warning("\nPossible solutions:")
		output.Warning("  - Check file permissions in the project")
		output.Warning("  - Ensure storage and bootstrap/cache are writable")
		output.Warning("  - Run: glid artisan storage:link")
		return glideErrors.NewPermissionError("storage", "permission denied in Laravel application",
			glideErrors.WithSuggestions(
				"Check file permissions in the project",
				"Ensure storage and bootstrap/cache are writable",
				"Run: glid artisan storage:link",
				"Fix permissions: chmod -R 775 storage bootstrap/cache",
			),
		)
	}

	// Class not found errors
	if strings.Contains(errStr, "Class") && strings.Contains(errStr, "not found") {
		output.Error("Class not found error")
		output.Warning("\nPossible solutions:")
		output.Warning("  - Run: glid composer dump-autoload")
		output.Warning("  - Check for typos in class names")
		output.Warning("  - Ensure the class file exists")
		return glideErrors.NewDependencyError("Laravel class", "class not found in Laravel application",
			glideErrors.WithSuggestions(
				"Run: glid composer dump-autoload",
				"Check for typos in class names",
				"Ensure the class file exists",
				"Run: glid composer install if dependencies are missing",
			),
		)
	}

	return err
}

// isExpectedNonZeroExit checks if a non-zero exit is expected
func (c *ArtisanCommand) isExpectedNonZeroExit(args []string, exitCode int) bool {
	if len(args) == 0 {
		return false
	}

	// Some artisan commands return non-zero for non-error conditions
	// For example:
	// - migrate:status returns 1 if there are pending migrations
	// - queue:work returns 1 if stopped gracefully
	// - Some make commands return 1 if file already exists

	// Check for specific commands
	if strings.HasPrefix(args[0], "migrate:status") && exitCode == 1 {
		// Pending migrations is not an error
		return true
	}

	if strings.HasPrefix(args[0], "make:") && exitCode == 1 {
		// File already exists is not an unexpected error
		output.Warning("File already exists or command cancelled")
		return true
	}

	return false
}

// startDocker starts the Docker containers
func (c *ArtisanCommand) startDocker() error {
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

	// Wait for PHP and MySQL containers
	health := docker.NewHealthMonitor(c.ctx)
	if err := health.WaitForHealthy(30*time.Second, "php", "mysql"); err != nil {
		spinner.Error("Containers failed health check")
		return err
	}

	spinner.Success("Docker containers started")
	
	// Update context to reflect Docker is now running
	c.ctx.DockerRunning = true
	
	return nil
}