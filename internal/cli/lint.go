package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ivannovak/glide/internal/config"
	"github.com/ivannovak/glide/internal/context"
	"github.com/ivannovak/glide/internal/docker"
	"github.com/ivannovak/glide/internal/shell"
	glideErrors "github.com/ivannovak/glide/pkg/errors"
	"github.com/ivannovak/glide/pkg/output"
	"github.com/ivannovak/glide/pkg/progress"
	"github.com/spf13/cobra"
)

// LintCommand handles the lint command
type LintCommand struct {
	ctx *context.ProjectContext
	cfg *config.Config
}

// NewLintCommand creates a new lint command
func NewLintCommand(ctx *context.ProjectContext, cfg *config.Config) *cobra.Command {
	lc := &LintCommand{
		ctx: ctx,
		cfg: cfg,
	}

	cmd := &cobra.Command{
		Use:   "lint [paths...] [flags]",
		Short: "Run PHP CS Fixer",
		Long: `Run PHP CS Fixer to check and fix code style issues.

This command runs PHP CS Fixer inside the Docker container to ensure
consistent code formatting across the project.

Examples:
  glid lint                           # Check all PHP files
  glid lint --fix                     # Fix all style issues
  glid lint app/                      # Check specific directory
  glid lint app/Models/User.php       # Check specific file
  glid lint --dry-run                 # Show what would be fixed
  glid lint --diff                    # Show detailed diff of changes
  glid lint --config=.php-cs-fixer.php # Use custom config file

Options:
  --fix           Apply fixes automatically
  --dry-run       Show what would be fixed without applying
  --diff          Show diff of changes
  --diff-format   Format for diff output (udiff, sbd)
  --format        Output format (txt, json, xml, junit, checkstyle)
  --config        Path to config file (default: .php-cs-fixer.php)
  --cache-file    Path to cache file
  --using-cache   Use cache (yes/no)
  --verbose       Show detailed output
  --stop-on-violation  Stop on first violation

Common Rules:
  The linter checks for:
  - PSR-12 coding standard compliance
  - Consistent indentation (spaces vs tabs)
  - Line endings (LF vs CRLF)
  - Trailing whitespace
  - Blank lines at end of file
  - Proper use of namespaces
  - Import sorting
  - Method and property visibility
  - And many more based on your configuration

Configuration:
  PHP CS Fixer configuration is typically in:
  - .php-cs-fixer.php (modern)
  - .php-cs-fixer.dist.php (distributable)
  - .php_cs (legacy)
  - .php_cs.dist (legacy distributable)

Tips:
  - Always run lint before committing code
  - Use --dry-run first to preview changes
  - Add to pre-commit hooks for automation
  - Custom rules can be added to config file`,
		RunE:          lc.Execute,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Add flags
	cmd.Flags().Bool("fix", false, "Apply fixes automatically")
	cmd.Flags().Bool("dry-run", false, "Show what would be fixed without applying")
	cmd.Flags().Bool("diff", false, "Show diff of changes")
	cmd.Flags().String("diff-format", "udiff", "Format for diff output (udiff, sbd)")
	cmd.Flags().String("format", "txt", "Output format")
	cmd.Flags().String("config", "", "Path to config file")
	cmd.Flags().String("cache-file", "", "Path to cache file")
	cmd.Flags().String("using-cache", "yes", "Use cache (yes/no)")
	cmd.Flags().BoolP("verbose", "v", false, "Show detailed output")
	cmd.Flags().Bool("stop-on-violation", false, "Stop on first violation")

	return cmd
}

// Execute runs the lint command
func (c *LintCommand) Execute(cmd *cobra.Command, args []string) error {
	// Check if we're in a valid project
	if c.ctx.ProjectRoot == "" {
		return glideErrors.New(glideErrors.TypeInvalid, "not in a project directory",
			glideErrors.WithSuggestions(
				"Navigate to a project root directory",
				"Ensure you're in the correct repository",
				"Check if the project was properly initialized",
			))
	}

	// Check if Docker is running
	if !c.ctx.DockerRunning {
		output.Warning("Docker is not running. Starting containers...")
		if err := c.startDocker(); err != nil {
			return glideErrors.NewDockerError("failed to start Docker",
				glideErrors.WithError(err),
				glideErrors.WithSuggestions(
					"Ensure Docker Desktop is installed and running",
					"Check Docker daemon status with 'docker info'",
					"Restart Docker Desktop if necessary",
				))
		}
	}

	// Resolve Docker compose files
	resolver := docker.NewResolver(c.ctx)
	if err := resolver.Resolve(); err != nil {
		return glideErrors.NewConfigError("failed to resolve Docker configuration",
			glideErrors.WithError(err),
			glideErrors.WithSuggestions(
				"Check if docker-compose.yml exists in the project",
				"Verify Docker Compose configuration is valid",
				"Ensure all required Docker Compose files are present",
			))
	}

	// Get flags
	fix, _ := cmd.Flags().GetBool("fix")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	diff, _ := cmd.Flags().GetBool("diff")
	diffFormat, _ := cmd.Flags().GetString("diff-format")
	format, _ := cmd.Flags().GetString("format")
	configFile, _ := cmd.Flags().GetString("config")
	cacheFile, _ := cmd.Flags().GetString("cache-file")
	usingCache, _ := cmd.Flags().GetString("using-cache")
	verbose, _ := cmd.Flags().GetBool("verbose")
	stopOnViolation, _ := cmd.Flags().GetBool("stop-on-violation")

	// Build and execute PHP CS Fixer command
	return c.runPhpCsFixer(resolver, args, fix, dryRun, diff, diffFormat, format, configFile, cacheFile, usingCache, verbose, stopOnViolation)
}

// runPhpCsFixer executes PHP CS Fixer in the container
func (c *LintCommand) runPhpCsFixer(resolver *docker.Resolver, paths []string, fix, dryRun, diff bool, diffFormat, format, configFile, cacheFile, usingCache string, verbose, stopOnViolation bool) error {
	// Check if PHP CS Fixer is available
	if err := c.checkPhpCsFixer(resolver); err != nil {
		return err
	}

	// Build PHP CS Fixer command
	fixerArgs := c.buildFixerCommand(paths, fix, dryRun, diff, diffFormat, format, configFile, cacheFile, usingCache, verbose, stopOnViolation)

	// Build Docker command
	dockerArgs := resolver.GetComposeCommand("exec", "-T")
	dockerArgs = append(dockerArgs, "php")
	dockerArgs = append(dockerArgs, fixerArgs...)

	// Show what we're doing
	action := "Checking"
	if fix && !dryRun {
		action = "Fixing"
	} else if dryRun {
		action = "Analyzing"
	}

	spinner := progress.NewSpinner(fmt.Sprintf("%s code style issues", action))
	spinner.Start()

	// Execute command
	executor := shell.NewExecutor(shell.Options{
		Verbose: verbose,
	})

	// For diff output or verbose mode, use passthrough
	var shellCmd *shell.Command
	if diff || verbose || format != "txt" {
		spinner.Stop()
		shellCmd = shell.NewPassthroughCommand("docker", dockerArgs...)
	} else {
		shellCmd = shell.NewCommand("docker", dockerArgs...)
	}

	result, err := executor.Execute(shellCmd)

	if !diff && !verbose {
		if err != nil || result.ExitCode != 0 {
			spinner.Error("Style issues found")

			// PHP CS Fixer returns different exit codes
			// 0 = no issues or issues fixed
			// 1 = general error
			// 4 = issues found (not fixed)
			// 8 = issues fixed
			// 16 = configuration error
			// 32 = fixing error
			// 64 = exception

			if result.ExitCode == 4 {
				output.Warning("\nCode style issues detected!")
				if !fix {
					output.Println("\nTo fix these issues, run:")
					output.Info("  glid lint --fix")
				}
				return nil // Don't treat style issues as error
			} else if result.ExitCode == 8 {
				spinner.Success("Code style issues fixed")
				return nil
			}

			return glideErrors.NewCommandError("PHP CS Fixer", result.ExitCode,
				glideErrors.WithSuggestions(
					"Check PHP CS Fixer configuration file",
					"Run with --verbose flag to see detailed output",
					"Verify PHP CS Fixer is properly installed",
				))
		}

		if fix && !dryRun {
			spinner.Success("Code style fixed")
		} else {
			spinner.Success("No style issues found")
		}
	}

	return nil
}

// buildFixerCommand constructs the PHP CS Fixer command
func (c *LintCommand) buildFixerCommand(paths []string, fix, dryRun, diff bool, diffFormat, format, configFile, cacheFile, usingCache string, verbose, stopOnViolation bool) []string {
	args := []string{}

	// Check for vendor binary vs global
	args = append(args, "./vendor/bin/php-cs-fixer")

	// Command (fix is the only command now)
	args = append(args, "fix")

	// Paths to check (default to current directory)
	if len(paths) > 0 {
		for _, path := range paths {
			// Ensure paths are relative to project root
			if !filepath.IsAbs(path) {
				args = append(args, path)
			} else {
				rel, _ := filepath.Rel(c.ctx.ProjectRoot, path)
				args = append(args, rel)
			}
		}
	} else {
		// Default to checking all PHP files
		args = append(args, ".")
	}

	// Flags
	if !fix || dryRun {
		args = append(args, "--dry-run")
	}

	if diff {
		args = append(args, "--diff")
		if diffFormat != "" && diffFormat != "udiff" {
			args = append(args, "--diff-format="+diffFormat)
		}
	}

	if format != "" && format != "txt" {
		args = append(args, "--format="+format)
	}

	if configFile != "" {
		args = append(args, "--config="+configFile)
	}

	if cacheFile != "" {
		args = append(args, "--cache-file="+cacheFile)
	}

	if usingCache != "yes" {
		args = append(args, "--using-cache="+usingCache)
	}

	if verbose {
		args = append(args, "-vvv")
	}

	if stopOnViolation {
		args = append(args, "--stop-on-violation")
	}

	return args
}

// checkPhpCsFixer verifies PHP CS Fixer is installed
func (c *LintCommand) checkPhpCsFixer(resolver *docker.Resolver) error {
	// Check if PHP CS Fixer exists in vendor
	dockerArgs := resolver.GetComposeCommand("exec", "-T")
	dockerArgs = append(dockerArgs, "php", "test", "-f", "./vendor/bin/php-cs-fixer")

	executor := shell.NewExecutor(shell.Options{
		Verbose: false,
	})

	shellCmd := shell.NewCommand("docker", dockerArgs...)
	result, err := executor.Execute(shellCmd)

	if err != nil || result.ExitCode != 0 {
		output.Warning("PHP CS Fixer not found in vendor/bin/")
		output.Println("\nTo install PHP CS Fixer, run:")
		output.Info("  glid composer require --dev friendsofphp/php-cs-fixer")
		return glideErrors.NewDependencyError("friendsofphp/php-cs-fixer", "PHP CS Fixer not installed",
			glideErrors.WithSuggestions(
				"Run 'glid composer require --dev friendsofphp/php-cs-fixer'",
				"Install PHP CS Fixer globally with 'composer global require friendsofphp/php-cs-fixer'",
				"Check if vendor directory exists and is properly set up",
			))
	}

	return nil
}

// startDocker starts the Docker containers
func (c *LintCommand) startDocker() error {
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

	spinner.Success("Docker containers started")

	// Update context
	c.ctx.DockerRunning = true

	return nil
}

// getConfigFile finds the PHP CS Fixer config file
func (c *LintCommand) getConfigFile() string {
	// Check for config files in order of preference
	configFiles := []string{
		".php-cs-fixer.php",
		".php-cs-fixer.dist.php",
		".php_cs",
		".php_cs.dist",
	}

	for _, file := range configFiles {
		path := filepath.Join(c.ctx.ProjectRoot, file)
		if fileExists(path) {
			return file
		}
	}

	return ""
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	// This would normally use os.Stat
	// For now, return false
	return false
}

// showFixSuggestion shows how to fix issues
func (c *LintCommand) showFixSuggestion(paths []string) {
	output.Println("\nTo automatically fix these issues, run:")

	cmd := "glid lint --fix"
	if len(paths) > 0 {
		cmd += " " + strings.Join(paths, " ")
	}

	output.Info("  %s", cmd)
}
