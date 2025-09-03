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

// ComposerCommand handles the composer command
type ComposerCommand struct {
	ctx *context.ProjectContext
	cfg *config.Config
}

// NewComposerCommand creates a new composer command
func NewComposerCommand(ctx *context.ProjectContext, cfg *config.Config) *cobra.Command {
	cc := &ComposerCommand{
		ctx: ctx,
		cfg: cfg,
	}

	cmd := &cobra.Command{
		Use:   "composer [composer arguments]",
		Short: "Run Composer commands via Docker",
		Long: `Execute Composer commands within the Docker PHP container.

This command passes all arguments directly to Composer running inside
the PHP container. It handles working directory mapping and ensures
proper dependency caching.

Examples:
  glid composer install              # Install dependencies
  glid composer update               # Update dependencies
  glid composer require laravel/ui   # Add a new package
  glid composer remove package/name  # Remove a package
  glid composer dump-autoload        # Regenerate autoloader
  glid composer show                 # List installed packages
  glid composer validate             # Validate composer.json

Advanced Usage:
  glid composer install --no-dev     # Install without dev dependencies
  glid composer update --dry-run     # Preview updates without applying
  glid composer require --dev phpunit/phpunit  # Add dev dependency
  glid composer exec phpunit         # Execute vendor binary

Working Directory:
  Commands are executed in the project root directory within the container.
  The vendor directory is cached in the container for performance.

Performance Notes:
  - Initial install may be slower due to Docker volume mounting
  - Subsequent operations use cached dependencies
  - Consider running 'composer install' during container build for production`,
		DisableFlagParsing: true, // Pass all flags through to composer
		RunE:               cc.Execute,
		SilenceUsage:       true,  // Don't show usage on error
		SilenceErrors:      true,  // Let our error handler handle errors
	}

	return cmd
}

// Execute runs the composer command
func (c *ComposerCommand) Execute(cmd *cobra.Command, args []string) error {
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
					"Start Docker Desktop manually",
					"Check Docker daemon status: docker ps",
					"Review Docker logs for errors",
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
				"Verify you're in the correct project directory",
				"Ensure Docker is installed: docker --version",
			),
		)
	}

	// Build the composer command
	composerArgs := c.buildComposerCommand(resolver, args)

	// Show what we're running (in verbose mode)
	if c.shouldShowCommand(args) {
		output.Info("Running: composer %s", strings.Join(args, " "))
	}

	// Check if this needs special handling
	needsProgress := c.needsProgressIndicator(args)

	// Execute the command
	if needsProgress {
		return c.executeWithProgress(composerArgs, args)
	}
	return c.executeComposerCommand(composerArgs)
}

// buildComposerCommand constructs the docker-compose exec command for composer
func (c *ComposerCommand) buildComposerCommand(resolver *docker.Resolver, args []string) []string {
	// Start with docker-compose exec command
	dockerArgs := resolver.GetComposeCommand("exec")

	// Add -T flag to disable TTY allocation for non-interactive commands
	// This prevents "the input device is not a TTY" errors in CI/scripts
	if !c.isInteractiveCommand(args) {
		dockerArgs = append(dockerArgs, "-T")
	}

	// Set working directory to project root in container
	dockerArgs = append(dockerArgs, "-w", "/var/www/html")

	// Add PHP container and composer command
	dockerArgs = append(dockerArgs, "php", "composer")

	// Add all composer arguments
	dockerArgs = append(dockerArgs, args...)

	return dockerArgs
}

// isInteractiveCommand checks if composer command needs TTY
func (c *ComposerCommand) isInteractiveCommand(args []string) bool {
	// Composer is generally non-interactive except for init
	if len(args) > 0 && args[0] == "init" {
		return true
	}
	
	// Check for explicit interaction flags
	for _, arg := range args {
		if arg == "--interactive" || arg == "-i" {
			return true
		}
	}
	
	return false
}

// shouldShowCommand determines if we should display the command being run
func (c *ComposerCommand) shouldShowCommand(args []string) bool {
	// Always show for install/update/require/remove
	if len(args) > 0 {
		switch args[0] {
		case "install", "update", "require", "remove":
			return true
		}
	}
	
	// Otherwise check verbose setting
	return c.cfg != nil && c.cfg.Defaults.Test.Verbose
}

// needsProgressIndicator checks if the command would benefit from progress indication
func (c *ComposerCommand) needsProgressIndicator(args []string) bool {
	if len(args) == 0 {
		return false
	}

	// Commands that typically take time
	switch args[0] {
	case "install", "update", "require", "remove":
		return true
	case "create-project":
		return true
	default:
		return false
	}
}

// executeWithProgress runs composer with a progress indicator
func (c *ComposerCommand) executeWithProgress(dockerArgs []string, composerArgs []string) error {
	// Determine the operation for the progress message
	operation := "Running composer"
	if len(composerArgs) > 0 {
		operation = fmt.Sprintf("Running composer %s", composerArgs[0])
	}

	spinner := progress.NewSpinner(operation)
	spinner.Start()
	defer spinner.Stop()

	// Execute the command
	executor := shell.NewExecutor(shell.Options{
		Verbose: false,
	})

	// For install/update, we want to see the output
	shellCmd := shell.NewPassthroughCommand("docker", dockerArgs...)
	result, err := executor.Execute(shellCmd)

	if err != nil {
		spinner.Error(fmt.Sprintf("Composer %s failed", composerArgs[0]))
		return err
	}

	if result.ExitCode != 0 {
		spinner.Error(fmt.Sprintf("Composer %s failed with exit code %d", composerArgs[0], result.ExitCode))
		return glideErrors.NewCommandError("composer", result.ExitCode,
			glideErrors.WithSuggestions(
				"Check the composer output above for specific errors",
				"Verify composer.json is valid: glid composer validate",
				"Clear composer cache: glid composer clear-cache",
				"Try with verbose mode: glid composer " + composerArgs[0] + " -vvv",
			),
		)
	}

	spinner.Success(fmt.Sprintf("Composer %s completed", composerArgs[0]))
	return nil
}

// executeComposerCommand runs a composer command without progress indication
func (c *ComposerCommand) executeComposerCommand(dockerArgs []string) error {
	executor := shell.NewExecutor(shell.Options{
		Verbose: false,
	})

	// Use passthrough for all composer commands to see output
	shellCmd := shell.NewPassthroughCommand("docker", dockerArgs...)
	result, err := executor.Execute(shellCmd)

	if err != nil {
		return c.handleComposerError(err)
	}

	if result.ExitCode != 0 {
		// Don't error for commands that might legitimately return non-zero
		// For example, 'composer outdated' returns 1 if packages are outdated
		if !c.isExpectedNonZeroExit(dockerArgs, result.ExitCode) {
			return glideErrors.NewCommandError("composer", result.ExitCode,
				glideErrors.WithSuggestions(
					"Check the composer output above for errors",
					"Verify composer.json syntax: glid composer validate",
					"Check PHP container logs: glid docker logs php",
				),
			)
		}
	}

	return nil
}

// handleComposerError provides helpful error messages for common issues
func (c *ComposerCommand) handleComposerError(err error) error {
	errStr := err.Error()

	// PHP container not running
	if strings.Contains(errStr, "service \"php\" is not running") {
		return glideErrors.NewDockerError("PHP container is not running",
			glideErrors.WithError(err),
			glideErrors.WithSuggestions(
				"Start containers: glid up",
				"Or manually: glid docker up -d",
				"Check container status: glid docker ps",
			),
		)
	}

	// Memory limit errors
	if strings.Contains(errStr, "Allowed memory size") || strings.Contains(errStr, "memory limit") {
		return glideErrors.NewRuntimeError("composer ran out of memory",
			glideErrors.WithError(err),
			glideErrors.WithSuggestions(
				"Increase PHP memory limit in Docker container",
				"Run with unlimited memory: COMPOSER_MEMORY_LIMIT=-1 glid composer install",
				"Clear composer cache: glid composer clear-cache",
				"Check available system memory",
			),
		)
	}

	// Network/connectivity issues
	if strings.Contains(errStr, "could not be fully loaded") || strings.Contains(errStr, "packagist.org") {
		return glideErrors.NewNetworkError("composer package repository unreachable",
			glideErrors.WithError(err),
			glideErrors.WithSuggestions(
				"Check your internet connection",
				"Try again in a few moments",
				"Clear composer cache: glid composer clear-cache",
				"Use a mirror: glid composer config repos.packagist composer https://packagist.org",
				"Check proxy settings if behind corporate firewall",
			),
		)
	}

	return glideErrors.AnalyzeError(err)
}

// isExpectedNonZeroExit checks if a non-zero exit is expected
func (c *ComposerCommand) isExpectedNonZeroExit(args []string, exitCode int) bool {
	// Find the composer command in the args
	composerCmdIndex := -1
	for i, arg := range args {
		if arg == "composer" {
			composerCmdIndex = i
			break
		}
	}

	if composerCmdIndex >= 0 && composerCmdIndex+1 < len(args) {
		composerCmd := args[composerCmdIndex+1]
		
		// 'outdated' returns 1 if packages are outdated (expected)
		if composerCmd == "outdated" && exitCode == 1 {
			return true
		}
		
		// 'validate' returns non-zero if composer.json is invalid (expected)
		if composerCmd == "validate" && exitCode != 0 {
			// Still show the validation errors, but don't treat as unexpected error
			output.Warning("Composer validation found issues")
			return true
		}
	}

	return false
}

// startDocker starts the Docker containers
func (c *ComposerCommand) startDocker() error {
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

	// Wait for PHP container specifically
	health := docker.NewHealthMonitor(c.ctx)
	if err := health.WaitForHealthy(30*time.Second, "php"); err != nil {
		spinner.Error("PHP container failed health check")
		return err
	}

	spinner.Success("Docker containers started")
	
	// Update context to reflect Docker is now running
	c.ctx.DockerRunning = true
	
	return nil
}

// checkComposerCache checks if composer cache is properly configured
func (c *ComposerCommand) checkComposerCache() {
	// Check if composer cache volume is mounted
	cacheDir := filepath.Join(c.ctx.ProjectRoot, ".composer-cache")
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		// Cache directory doesn't exist, composer will use container's cache
		// This is fine but might be slower for repeated operations
		if c.cfg != nil && c.cfg.Defaults.Test.Verbose {
			output.Warning("Tip: Consider adding a composer cache volume for better performance")
		}
	}
}