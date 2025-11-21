package shell

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/ivannovak/glide/internal/config"
	"github.com/ivannovak/glide/internal/context"
)

// TestExecutor handles test command execution with Pest
type TestExecutor struct {
	executor *Executor
	docker   *DockerExecutor
	ctx      *context.ProjectContext
	config   *config.CommandConfig
}

// NewTestExecutor creates a new test command executor
func NewTestExecutor(ctx *context.ProjectContext, cfg *config.CommandConfig) *TestExecutor {
	return &TestExecutor{
		executor: NewExecutor(Options{
			Verbose: false,
		}),
		docker: NewDockerExecutor(ctx),
		ctx:    ctx,
		config: cfg,
	}
}

// Run executes Pest tests with full argument passthrough
func (te *TestExecutor) Run(args []string) error {
	// Check if Docker is running
	if !te.docker.IsRunning() {
		return fmt.Errorf("docker is not running. Please start Docker Desktop and try again")
	}

	// Check if containers are running
	status, err := te.docker.GetContainerStatus()
	if err != nil || len(status) == 0 {
		if te.config != nil && te.config.Docker.AutoStart {
			color.Yellow("Starting Docker containers...")
			if err := te.docker.Up(true); err != nil {
				return fmt.Errorf("failed to start containers: %w", err)
			}
		} else {
			return fmt.Errorf("docker containers are not running. Run 'glideup' first")
		}
	}

	// Build the test command
	pestArgs := te.buildPestCommand(args)

	// Execute tests in PHP container
	if len(pestArgs) > 0 {
		return te.docker.RunInContainer("php", pestArgs[0], pestArgs[1:]...)
	}
	return fmt.Errorf("no test command to execute")
}

// buildPestCommand builds the Pest command with appropriate defaults
func (te *TestExecutor) buildPestCommand(userArgs []string) []string {
	// Start with vendor/bin/pest
	args := []string{"vendor/bin/pest"}

	// If no arguments provided, use defaults from config
	if len(userArgs) == 0 && te.config != nil {
		if te.config.Test.Parallel {
			args = append(args, "--parallel")
			if te.config.Test.Processes > 0 {
				args = append(args, fmt.Sprintf("--processes=%d", te.config.Test.Processes))
			}
		}
		if te.config.Test.Coverage {
			args = append(args, "--coverage")
		}
		if te.config.Test.Verbose {
			args = append(args, "-v")
		}
	} else {
		// Pass through all user arguments without interpretation
		args = append(args, userArgs...)
	}

	return args
}

// PassthroughToPest passes all arguments directly to Pest
func (te *TestExecutor) PassthroughToPest(args []string) error {
	// Check Docker status first
	if !te.docker.IsRunning() {
		return fmt.Errorf("docker is not running. Please start Docker Desktop and try again")
	}

	// Build full command for docker exec
	execArgs := []string{"exec", "php", "vendor/bin/pest"}
	execArgs = append(execArgs, args...)

	// Use docker-compose with passthrough
	return te.docker.Compose(execArgs...)
}

// RunWithDefaults runs tests with configuration defaults
func (te *TestExecutor) RunWithDefaults() error {
	var args []string

	if te.config != nil {
		if te.config.Test.Parallel {
			args = append(args, "--parallel")
			if te.config.Test.Processes > 0 {
				args = append(args, fmt.Sprintf("--processes=%d", te.config.Test.Processes))
			}
		}
		if te.config.Test.Coverage {
			args = append(args, "--coverage")
		}
		if te.config.Test.Verbose {
			args = append(args, "-v")
		}
	} else {
		// Fallback defaults if no config
		args = []string{"--parallel", "--processes=3"}
	}

	return te.Run(args)
}

// CheckDependencies verifies test dependencies are installed
func (te *TestExecutor) CheckDependencies() error {
	// Check if Pest is installed
	output, err := te.docker.ComposeCapture("exec", "-T", "php", "test", "-f", "vendor/bin/pest")
	if err != nil || strings.Contains(output, "cannot find") {
		color.Yellow("Pest not found. Installing test dependencies...")

		// Run composer install
		if err := te.docker.RunInContainer("php", "composer", "install", "--dev"); err != nil {
			return fmt.Errorf("failed to install dependencies: %w", err)
		}
	}

	return nil
}

// PrepareTestDatabase sets up the test database
func (te *TestExecutor) PrepareTestDatabase() error {
	// Run migrations for test database
	color.Cyan("Preparing test database...")

	// Set test environment
	testEnv := []string{"APP_ENV=testing"}
	cmd := NewCommand("docker", "compose")

	// Add compose files
	for _, file := range te.ctx.ComposeFiles {
		cmd.Args = append(cmd.Args, "-f", file)
	}

	// Add exec command
	cmd.Args = append(cmd.Args, "exec", "-T", "php", "php", "artisan", "migrate:fresh", "--env=testing", "--force")
	cmd.Environment = testEnv

	result, err := te.executor.Execute(cmd)
	if err != nil {
		return fmt.Errorf("failed to prepare test database: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("migration failed with exit code %d", result.ExitCode)
	}

	return nil
}

// RunSingleTest runs a specific test file or method
func (te *TestExecutor) RunSingleTest(testPath string, method string) error {
	args := []string{testPath}

	if method != "" {
		args = append(args, "--filter", method)
	}

	// Add any configured defaults
	if te.config != nil && te.config.Test.Verbose {
		args = append(args, "-v")
	}

	return te.Run(args)
}

// ListTests lists all available tests
func (te *TestExecutor) ListTests() error {
	return te.Run([]string{"--list-tests"})
}

// ShowCoverage runs tests with coverage report
func (te *TestExecutor) ShowCoverage(format string) error {
	args := []string{"--coverage"}

	if format != "" {
		args = append(args, "--coverage-"+format)
	}

	return te.Run(args)
}
