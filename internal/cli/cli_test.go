package cli

import (
	"bytes"
	"os"
	"testing"

	"github.com/ivannovak/glide/v2/internal/config"
	"github.com/ivannovak/glide/v2/internal/context"
	"github.com/ivannovak/glide/v2/pkg/output"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCLI(t *testing.T) {
	t.Run("creates CLI with dependencies", func(t *testing.T) {
		outputMgr := output.NewManager(output.FormatPlain, false, false, os.Stdout)
		ctx := &context.ProjectContext{
			ProjectRoot: "/test",
		}
		cfg := &config.Config{
			DefaultProject: "test",
		}

		cli := New(outputMgr, ctx, cfg)

		assert.NotNil(t, cli)
		assert.NotNil(t, cli.outputManager)
		assert.NotNil(t, cli.projectContext)
		assert.NotNil(t, cli.config)
		assert.Equal(t, outputMgr, cli.outputManager)
		assert.Equal(t, ctx, cli.projectContext)
		assert.Equal(t, cfg, cli.config)
	})
}

func TestCLICommandCreation(t *testing.T) {
	outputMgr, ctx, cfg := createTestDependencies()
	cli := New(outputMgr, ctx, cfg)

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
		assert.Equal(t, "project", cmd.Use)
	})
}

func TestCLIAddLocalCommands(t *testing.T) {
	outputMgr, ctx, cfg := createTestDependencies()
	cli := New(outputMgr, ctx, cfg)

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

		outputMgr := output.NewManager(output.FormatPlain, false, false, buf)
		cfg := &config.Config{}
		cli := New(outputMgr, ctx, cfg)

		cmd := &cobra.Command{}
		cmd.SetOut(buf)

		cli.showContext(cmd)

		outputStr := buf.String()
		assert.Contains(t, outputStr, "Project Context")
		assert.Contains(t, outputStr, "/test/working")
		assert.Contains(t, outputStr, "/test/project")
		assert.Contains(t, outputStr, "single-repo")
		assert.Contains(t, outputStr, "Docker Running: true")
		assert.Contains(t, outputStr, "docker-compose.yml")
	})

	t.Run("auto-detects context when not provided", func(t *testing.T) {
		buf := &bytes.Buffer{}
		outputMgr := output.NewManager(output.FormatPlain, false, false, buf)
		// Create minimal context for testing
		ctx := &context.ProjectContext{
			WorkingDir: "/test",
		}
		cfg := &config.Config{}
		cli := New(outputMgr, ctx, cfg)

		cmd := &cobra.Command{}
		cmd.SetOut(buf)

		cli.showContext(cmd)

		outputStr := buf.String()
		// With DI container integration, context is now auto-detected
		assert.Contains(t, outputStr, "Project Context")
		assert.Contains(t, outputStr, "Working Directory:")
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

		outputMgr := output.NewManager(output.FormatPlain, false, false, buf)
		cfg := &config.Config{}
		cli := New(outputMgr, ctx, cfg)

		cmd := &cobra.Command{}
		cmd.SetOut(buf)

		cli.showContext(cmd)

		outputStr := buf.String()
		assert.Contains(t, outputStr, "multi-worktree")
		assert.Contains(t, outputStr, "Is Worktree: true")
		assert.Contains(t, outputStr, "Worktree Name: feature-branch")
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

		outputMgr := output.NewManager(output.FormatPlain, false, false, buf)
		ctx := &context.ProjectContext{
			ProjectRoot: "/test",
		}
		cli := New(outputMgr, ctx, cfg)

		cmd := &cobra.Command{}
		cmd.SetOut(buf)

		cli.showConfig(cmd)

		outputStr := buf.String()
		assert.Contains(t, outputStr, "Configuration")
		assert.Contains(t, outputStr, "myproject")
		assert.Contains(t, outputStr, "/path/to/project")
		assert.Contains(t, outputStr, "Parallel: true")
		assert.Contains(t, outputStr, "Processes: 4")
		assert.Contains(t, outputStr, "Coverage: true")
	})

	t.Run("auto-loads config when not provided", func(t *testing.T) {
		buf := &bytes.Buffer{}
		outputMgr := output.NewManager(output.FormatPlain, false, false, buf)
		ctx := &context.ProjectContext{
			ProjectRoot: "/test",
		}
		cfg := &config.Config{}
		cli := New(outputMgr, ctx, cfg)

		cmd := &cobra.Command{}
		cmd.SetOut(buf)

		cli.showConfig(cmd)

		outputStr := buf.String()
		// With DI container integration, config is now auto-loaded
		assert.Contains(t, outputStr, "Configuration")
	})
}

func TestCLIDependencyInjection(t *testing.T) {
	t.Run("commands use injected output manager", func(t *testing.T) {
		buf := &bytes.Buffer{}
		outputMgr := output.NewManager(output.FormatJSON, false, false, buf)
		ctx := &context.ProjectContext{
			ProjectRoot: "/test",
		}
		cfg := &config.Config{}
		cli := New(outputMgr, ctx, cfg)

		// Test that the CLI uses the injected output manager
		assert.Equal(t, output.FormatJSON, cli.outputManager.GetFormat())

		// Output should go to our buffer
		_ = cli.outputManager.Info("test message")
		assert.Contains(t, buf.String(), "test message")
	})

	t.Run("commands share dependencies", func(t *testing.T) {
		ctx := &context.ProjectContext{
			ProjectRoot: "/shared/project",
		}
		cfg := &config.Config{
			DefaultProject: "shared",
		}

		outputMgr := output.NewManager(output.FormatPlain, false, false, os.Stdout)
		cli := New(outputMgr, ctx, cfg)

		// All commands should see the same context and config
		assert.Same(t, ctx, cli.projectContext)
		assert.Same(t, cfg, cli.config)
	})
}

func TestCLITestShell(t *testing.T) {
	t.Run("executes shell tests", func(t *testing.T) {
		buf := &bytes.Buffer{}
		outputMgr := output.NewManager(output.FormatPlain, false, false, buf)
		ctx := &context.ProjectContext{
			ProjectRoot: "/test",
		}
		cfg := &config.Config{}
		cli := New(outputMgr, ctx, cfg)

		cmd := &cobra.Command{}
		cmd.SetOut(buf)

		// This should run without errors
		cli.testShell(cmd, []string{})

		outputStr := buf.String()
		assert.Contains(t, outputStr, "Shell Execution Test")
		assert.Contains(t, outputStr, "Test 1: Capture output")
		assert.Contains(t, outputStr, "Test 2: Command with timeout")
		assert.Contains(t, outputStr, "Test 3: Progress indicator")
	})
}

// Helper function to create test dependencies
func createTestDependencies() (*output.Manager, *context.ProjectContext, *config.Config) {
	outputMgr := output.NewManager(output.FormatTable, false, false, os.Stdout)
	ctx := &context.ProjectContext{
		ProjectRoot:     "/test/project",
		WorkingDir:      "/test/project",
		DevelopmentMode: context.ModeSingleRepo,
	}
	cfg := &config.Config{
		DefaultProject: "test",
	}
	return outputMgr, ctx, cfg
}

func TestCLIIntegration(t *testing.T) {
	t.Run("full CLI workflow", func(t *testing.T) {
		// Create dependencies
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

		outputMgr := output.NewManager(output.FormatTable, false, false, buf)
		cli := New(outputMgr, ctx, cfg)
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
		err = contextCmd.RunE(contextCmd, []string{})
		require.NoError(t, err)

		outputStr := buf.String()
		assert.Contains(t, outputStr, "Project Context")
		assert.Contains(t, outputStr, "/test/project")
	})
}
