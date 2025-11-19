package cli

import (
	"bytes"
	"testing"

	"github.com/ivannovak/glide/internal/config"
	"github.com/ivannovak/glide/internal/context"
	"github.com/ivannovak/glide/pkg/app"
	"github.com/ivannovak/glide/pkg/output"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCLI(t *testing.T) {
	t.Run("creates CLI with application", func(t *testing.T) {
		application := app.NewApplication(
			app.WithProjectContext(&context.ProjectContext{
				ProjectRoot: "/test",
			}),
			app.WithConfig(&config.Config{
				DefaultProject: "test",
			}),
		)

		cli := New(application)

		assert.NotNil(t, cli)
		assert.NotNil(t, cli.app)
		assert.Equal(t, application, cli.app)
	})
}

func TestCLICommandCreation(t *testing.T) {
	application := createTestApplication()
	cli := New(application)

	t.Run("NewSetupCommand", func(t *testing.T) {
		cmd := cli.NewSetupCommand()
		assert.NotNil(t, cmd)
		assert.Equal(t, "setup", cmd.Use)
	})

	t.Run("NewConfigCommand", func(t *testing.T) {
		cmd := cli.NewConfigCommand()
		assert.NotNil(t, cmd)
		assert.Equal(t, "config", cmd.Use)
	})

	t.Run("NewCompletionCommand", func(t *testing.T) {
		cmd := cli.NewCompletionCommand()
		assert.NotNil(t, cmd)
		assert.Equal(t, "completion [shell]", cmd.Use)
	})

	t.Run("NewProjectCommand", func(t *testing.T) {
		cmd := cli.NewProjectCommand()
		assert.NotNil(t, cmd)
		assert.Equal(t, "global", cmd.Use)
	})
}

func TestCLIAddLocalCommands(t *testing.T) {
	application := createTestApplication()
	cli := New(application)

	rootCmd := &cobra.Command{
		Use: "test",
	}

	cli.AddLocalCommands(rootCmd)

	// Check that commands were added
	commandNames := make([]string, 0)
	for _, cmd := range rootCmd.Commands() {
		commandNames = append(commandNames, cmd.Name())
	}

	// Verify some expected commands
	assert.Contains(t, commandNames, "context")
	assert.Contains(t, commandNames, "config")
	assert.Contains(t, commandNames, "version")
	assert.Contains(t, commandNames, "setup")
	assert.Contains(t, commandNames, "plugins")
	// Docker and dev commands have been moved to plugins
}

func TestCLIShowContext(t *testing.T) {
	t.Run("displays project context", func(t *testing.T) {
		buf := &bytes.Buffer{}
		ctx := &context.ProjectContext{
			WorkingDir:      "/test/working",
			ProjectRoot:     "/test/project",
			DevelopmentMode: context.ModeSingleRepo,
			Location:        context.LocationProject,
			DockerRunning:   true,
			ComposeFiles:    []string{"docker-compose.yml"},
		}

		application := app.NewApplication(
			app.WithProjectContext(ctx),
			app.WithWriter(buf),
		)
		cli := New(application)

		cmd := &cobra.Command{}
		cmd.SetOut(buf)

		cli.showContext(cmd)

		output := buf.String()
		assert.Contains(t, output, "Project Context")
		assert.Contains(t, output, "/test/working")
		assert.Contains(t, output, "/test/project")
		assert.Contains(t, output, "single-repo")
		assert.Contains(t, output, "Docker Running: true")
		assert.Contains(t, output, "docker-compose.yml")
	})

	t.Run("handles nil context", func(t *testing.T) {
		buf := &bytes.Buffer{}
		application := app.NewApplication(
			app.WithWriter(buf),
		)
		cli := New(application)

		cmd := &cobra.Command{}
		cmd.SetOut(buf)

		cli.showContext(cmd)

		output := buf.String()
		assert.Contains(t, output, "No project context available")
	})

	t.Run("shows multi-worktree details", func(t *testing.T) {
		buf := &bytes.Buffer{}
		ctx := &context.ProjectContext{
			WorkingDir:      "/test/worktrees/feature",
			ProjectRoot:     "/test",
			DevelopmentMode: context.ModeMultiWorktree,
			IsWorktree:      true,
			WorktreeName:    "feature-branch",
		}

		application := app.NewApplication(
			app.WithProjectContext(ctx),
			app.WithWriter(buf),
		)
		cli := New(application)

		cmd := &cobra.Command{}
		cmd.SetOut(buf)

		cli.showContext(cmd)

		output := buf.String()
		assert.Contains(t, output, "multi-worktree")
		assert.Contains(t, output, "Is Worktree: true")
		assert.Contains(t, output, "Worktree Name: feature-branch")
	})
}

func TestCLIShowConfig(t *testing.T) {
	t.Run("displays configuration", func(t *testing.T) {
		buf := &bytes.Buffer{}
		cfg := &config.Config{
			DefaultProject: "myproject",
			Projects: map[string]config.ProjectConfig{
				"myproject": {
					Path: "/path/to/project",
					Mode: "single-repo",
				},
			},
			Defaults: config.DefaultsConfig{
				Test: config.TestDefaults{
					Parallel:  true,
					Processes: 4,
					Coverage:  true,
					Verbose:   false,
				},
				Docker: config.DockerDefaults{
					ComposeTimeout: 30,
					AutoStart:      true,
					RemoveOrphans:  true,
				},
			},
		}

		application := app.NewApplication(
			app.WithConfig(cfg),
			app.WithWriter(buf),
		)
		cli := New(application)

		cmd := &cobra.Command{}
		cmd.SetOut(buf)

		cli.showConfig(cmd)

		output := buf.String()
		assert.Contains(t, output, "Configuration")
		assert.Contains(t, output, "myproject")
		assert.Contains(t, output, "/path/to/project")
		assert.Contains(t, output, "Parallel: true")
		assert.Contains(t, output, "Processes: 4")
		assert.Contains(t, output, "Coverage: true")
	})

	t.Run("handles nil config", func(t *testing.T) {
		buf := &bytes.Buffer{}
		application := app.NewApplication(
			app.WithWriter(buf),
		)
		cli := New(application)

		cmd := &cobra.Command{}
		cmd.SetOut(buf)

		cli.showConfig(cmd)

		output := buf.String()
		assert.Contains(t, output, "No configuration loaded")
	})
}

func TestCLIDependencyInjection(t *testing.T) {
	t.Run("commands use injected output manager", func(t *testing.T) {
		buf := &bytes.Buffer{}
		application := app.NewApplication(
			app.WithWriter(buf),
			app.WithOutputFormat(output.FormatJSON, false, false),
		)
		cli := New(application)

		// Test that the CLI uses the injected output manager
		assert.Equal(t, output.FormatJSON, cli.app.OutputManager.GetFormat())

		// Output should go to our buffer
		cli.app.OutputManager.Info("test message")
		assert.Contains(t, buf.String(), "test message")
	})

	t.Run("commands share application state", func(t *testing.T) {
		ctx := &context.ProjectContext{
			ProjectRoot: "/shared/project",
		}
		cfg := &config.Config{
			DefaultProject: "shared",
		}

		application := app.NewApplication(
			app.WithProjectContext(ctx),
			app.WithConfig(cfg),
		)
		cli := New(application)

		// All commands should see the same context and config
		assert.Same(t, ctx, cli.app.ProjectContext)
		assert.Same(t, cfg, cli.app.Config)
	})
}

func TestCLITestShell(t *testing.T) {
	t.Run("executes shell tests", func(t *testing.T) {
		buf := &bytes.Buffer{}
		application := app.NewApplication(
			app.WithWriter(buf),
			app.WithProjectContext(&context.ProjectContext{
				ProjectRoot: "/test",
			}),
		)
		cli := New(application)

		cmd := &cobra.Command{}
		cmd.SetOut(buf)

		// This should run without errors
		cli.testShell(cmd, []string{})

		output := buf.String()
		assert.Contains(t, output, "Shell Execution Test")
		assert.Contains(t, output, "Test 1: Capture output")
		assert.Contains(t, output, "Test 2: Command with timeout")
		assert.Contains(t, output, "Test 3: Progress indicator")
	})
}

// Helper function to create a test application
func createTestApplication() *app.Application {
	return app.NewApplication(
		app.WithProjectContext(&context.ProjectContext{
			ProjectRoot:     "/test/project",
			WorkingDir:      "/test/project",
			DevelopmentMode: context.ModeSingleRepo,
		}),
		app.WithConfig(&config.Config{
			DefaultProject: "test",
		}),
		app.WithOutputFormat(output.FormatTable, false, false),
	)
}

func TestCLIIntegration(t *testing.T) {
	t.Run("full CLI workflow", func(t *testing.T) {
		// Create a complete application
		buf := &bytes.Buffer{}
		ctx := &context.ProjectContext{
			WorkingDir:      "/test/project",
			ProjectRoot:     "/test/project",
			DevelopmentMode: context.ModeSingleRepo,
			Location:        context.LocationProject,
			DockerRunning:   false,
		}
		cfg := &config.Config{
			DefaultProject: "integration-test",
			Projects: map[string]config.ProjectConfig{
				"integration-test": {
					Path: "/test/project",
					Mode: "single-repo",
				},
			},
		}

		application := app.NewApplication(
			app.WithProjectContext(ctx),
			app.WithConfig(cfg),
			app.WithWriter(buf),
			app.WithOutputFormat(output.FormatTable, false, false),
		)

		cli := New(application)
		require.NotNil(t, cli)

		// Create root command and add commands
		rootCmd := &cobra.Command{
			Use:   "glide",
			Short: "Test CLI",
		}
		rootCmd.SetOut(buf)

		cli.AddLocalCommands(rootCmd)

		// Verify commands were added
		assert.True(t, rootCmd.HasSubCommands())

		// Test that we can execute a command
		contextCmd, _, err := rootCmd.Find([]string{"context"})
		require.NoError(t, err)
		require.NotNil(t, contextCmd)

		// Execute the context command with proper setup
		buf.Reset()
		contextCmd.SetOut(buf)
		contextCmd.SetErr(buf)
		contextCmd.Run(contextCmd, []string{})

		output := buf.String()
		assert.Contains(t, output, "Project Context")
		assert.Contains(t, output, "/test/project")
	})
}
