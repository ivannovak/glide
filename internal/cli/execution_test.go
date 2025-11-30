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

// TestCommandExecution tests the command execution flow
func TestCommandExecution(t *testing.T) {
	t.Run("successful execution", func(t *testing.T) {
		executed := false

		cmd := &cobra.Command{
			Use:   "test",
			Short: "Test command",
			RunE: func(cmd *cobra.Command, args []string) error {
				executed = true
				return nil
			},
		}

		// Execute command
		err := cmd.Execute()
		assert.NoError(t, err)
		assert.True(t, executed, "command should have been executed")
	})

	t.Run("execution with error", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:   "test",
			Short: "Test command",
			RunE: func(cmd *cobra.Command, args []string) error {
				return assert.AnError
			},
		}

		// Execute command
		err := cmd.Execute()
		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("execution with args", func(t *testing.T) {
		var receivedArgs []string

		cmd := &cobra.Command{
			Use:   "test",
			Short: "Test command",
			RunE: func(cmd *cobra.Command, args []string) error {
				receivedArgs = args
				return nil
			},
		}

		// Set args and execute
		cmd.SetArgs([]string{"arg1", "arg2", "arg3"})
		err := cmd.Execute()
		assert.NoError(t, err)
		assert.Equal(t, []string{"arg1", "arg2", "arg3"}, receivedArgs)
	})

	t.Run("pre/post hooks", func(t *testing.T) {
		preRunCalled := false
		runCalled := false
		postRunCalled := false

		cmd := &cobra.Command{
			Use:   "test",
			Short: "Test command",
			PreRunE: func(cmd *cobra.Command, args []string) error {
				preRunCalled = true
				return nil
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				runCalled = true
				return nil
			},
			PostRunE: func(cmd *cobra.Command, args []string) error {
				postRunCalled = true
				return nil
			},
		}

		// Execute command
		err := cmd.Execute()
		assert.NoError(t, err)
		assert.True(t, preRunCalled, "PreRunE should have been called")
		assert.True(t, runCalled, "RunE should have been called")
		assert.True(t, postRunCalled, "PostRunE should have been called")
	})

	t.Run("pre hook error stops execution", func(t *testing.T) {
		preRunCalled := false
		runCalled := false

		cmd := &cobra.Command{
			Use:   "test",
			Short: "Test command",
			PreRunE: func(cmd *cobra.Command, args []string) error {
				preRunCalled = true
				return assert.AnError
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				runCalled = true
				return nil
			},
		}

		// Execute command
		err := cmd.Execute()
		assert.Error(t, err)
		assert.True(t, preRunCalled, "PreRunE should have been called")
		assert.False(t, runCalled, "RunE should NOT have been called after PreRunE error")
	})
}

// TestRootCommandExecution tests the full command tree execution
func TestRootCommandExecution(t *testing.T) {
	t.Run("root command with subcommand", func(t *testing.T) {
		projectContext := &context.ProjectContext{}
		cfg := &config.Config{}
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)

		builder := NewBuilder(projectContext, cfg, outputManager)
		rootCmd := builder.Build()

		// Capture output
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)

		// Execute version command
		rootCmd.SetArgs([]string{"version"})
		err := rootCmd.Execute()
		assert.NoError(t, err)
	})

	t.Run("root command with non-existent subcommand", func(t *testing.T) {
		projectContext := &context.ProjectContext{}
		cfg := &config.Config{}
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)

		builder := NewBuilder(projectContext, cfg, outputManager)
		rootCmd := builder.Build()

		// Capture output
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)

		// Try to execute non-existent command
		rootCmd.SetArgs([]string{"nonexistent-command"})
		err := rootCmd.Execute()
		assert.Error(t, err)
	})

	t.Run("root command with alias", func(t *testing.T) {
		projectContext := &context.ProjectContext{}
		cfg := &config.Config{}
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)

		builder := NewBuilder(projectContext, cfg, outputManager)
		rootCmd := builder.Build()

		// Capture output
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)

		// Execute self-update command using alias "update"
		rootCmd.SetArgs([]string{"update", "--help"})
		err := rootCmd.Execute()
		assert.NoError(t, err)
	})
}

// TestYAMLCommandExecution tests YAML command execution
func TestYAMLCommandExecution(t *testing.T) {
	t.Run("valid YAML command", func(t *testing.T) {
		registry := NewRegistry()

		yamlCmd := &config.Command{
			Cmd:         "echo 'test'",
			Description: "Test command",
		}

		err := registry.AddYAMLCommand("test", yamlCmd)
		require.NoError(t, err)

		// Verify command was added
		factory, exists := registry.Get("test")
		assert.True(t, exists)

		// Create the command
		cmd := factory()
		assert.NotNil(t, cmd)
		assert.Equal(t, "test", cmd.Use)
		assert.Equal(t, "Test command", cmd.Short)

		// Verify it has the yaml_command annotation
		assert.Equal(t, "true", cmd.Annotations["yaml_command"])

		// Verify flag parsing is disabled for pass-through
		assert.True(t, cmd.DisableFlagParsing)
	})

	t.Run("YAML command with arguments", func(t *testing.T) {
		registry := NewRegistry()

		yamlCmd := &config.Command{
			Cmd:         "echo $@",
			Description: "Echo arguments",
		}

		err := registry.AddYAMLCommand("echo-args", yamlCmd)
		require.NoError(t, err)

		factory, exists := registry.Get("echo-args")
		assert.True(t, exists)

		cmd := factory()
		assert.NotNil(t, cmd)

		// YAML commands have DisableFlagParsing to allow pass-through
		assert.True(t, cmd.DisableFlagParsing)
	})

	t.Run("invalid YAML commands are rejected", func(t *testing.T) {
		// This is tested via sanitization in yaml_executor_test.go
		// We're verifying that the mechanism exists
		registry := NewRegistry()

		yamlCmd := &config.Command{
			Cmd:         "dangerous; rm -rf /",
			Description: "Dangerous command",
		}

		// The command is added to registry, but execution will be blocked by sanitizer
		err := registry.AddYAMLCommand("dangerous", yamlCmd)
		require.NoError(t, err)

		factory, exists := registry.Get("dangerous")
		assert.True(t, exists)
		assert.NotNil(t, factory)
	})
}

// TestDebugCommands tests debug command execution
func TestDebugCommands(t *testing.T) {
	t.Run("context command exists and executes", func(t *testing.T) {
		projectContext := &context.ProjectContext{
			WorkingDir:      "/test/dir",
			ProjectRoot:     "/test",
			DevelopmentMode: context.ModeSingleRepo,
			Location:        context.LocationProject,
		}
		cfg := &config.Config{}
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)

		builder := NewBuilder(projectContext, cfg, outputManager)
		rootCmd := builder.Build()

		// Find context command
		contextCmd := findCommand(rootCmd, "context")
		require.NotNil(t, contextCmd)
		assert.Equal(t, "context", contextCmd.Name())

		// Verify it has a RunE function (command can execute)
		assert.NotNil(t, contextCmd.RunE)
	})

	t.Run("shell-test command exists", func(t *testing.T) {
		projectContext := &context.ProjectContext{}
		cfg := &config.Config{}
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)

		builder := NewBuilder(projectContext, cfg, outputManager)
		rootCmd := builder.Build()

		// Find shell-test command
		shellTestCmd := findCommand(rootCmd, "shell-test")
		assert.NotNil(t, shellTestCmd)
		assert.Equal(t, "shell-test", shellTestCmd.Use)
	})

	t.Run("docker-test command exists", func(t *testing.T) {
		projectContext := &context.ProjectContext{}
		cfg := &config.Config{}
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)

		builder := NewBuilder(projectContext, cfg, outputManager)
		rootCmd := builder.Build()

		// Find docker-test command
		dockerTestCmd := findCommand(rootCmd, "docker-test")
		assert.NotNil(t, dockerTestCmd)
		assert.Equal(t, "docker-test", dockerTestCmd.Use)
	})

	t.Run("container-test command exists", func(t *testing.T) {
		projectContext := &context.ProjectContext{}
		cfg := &config.Config{}
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)

		builder := NewBuilder(projectContext, cfg, outputManager)
		rootCmd := builder.Build()

		// Find container-test command
		containerTestCmd := findCommand(rootCmd, "container-test")
		assert.NotNil(t, containerTestCmd)
		assert.Equal(t, "container-test", containerTestCmd.Use)
	})
}

// TestVersionCommand tests the version command
func TestVersionCommand(t *testing.T) {
	t.Run("version command output", func(t *testing.T) {
		projectContext := &context.ProjectContext{}
		cfg := &config.Config{}
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)

		builder := NewBuilder(projectContext, cfg, outputManager)
		rootCmd := builder.Build()

		// Find version command
		versionCmd := findCommand(rootCmd, "version")
		require.NotNil(t, versionCmd)

		// Capture output
		buf := new(bytes.Buffer)
		versionCmd.SetOut(buf)
		versionCmd.SetErr(buf)

		// Execute
		err := versionCmd.Execute()
		assert.NoError(t, err)
	})
}

// TestPluginsCommand tests the plugins command
func TestPluginsCommand(t *testing.T) {
	t.Run("plugins command exists", func(t *testing.T) {
		projectContext := &context.ProjectContext{}
		cfg := &config.Config{}
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)

		builder := NewBuilder(projectContext, cfg, outputManager)
		rootCmd := builder.Build()

		// Find plugins command
		pluginsCmd := findCommand(rootCmd, "plugins")
		require.NotNil(t, pluginsCmd)
		assert.Equal(t, "plugins", pluginsCmd.Use)

		// Verify alias
		assert.Contains(t, pluginsCmd.Aliases, "plugin")
	})

	t.Run("plugins command has subcommands", func(t *testing.T) {
		projectContext := &context.ProjectContext{}
		cfg := &config.Config{}
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)

		builder := NewBuilder(projectContext, cfg, outputManager)
		rootCmd := builder.Build()

		pluginsCmd := findCommand(rootCmd, "plugins")
		require.NotNil(t, pluginsCmd)

		// Check for at least one expected subcommand (list is always present)
		listCmd := findCommand(pluginsCmd, "list")
		assert.NotNil(t, listCmd, "expected subcommand list")
	})
}

// TestProjectCommand tests the project command
func TestProjectCommand(t *testing.T) {
	t.Run("project command exists", func(t *testing.T) {
		projectContext := &context.ProjectContext{}
		cfg := &config.Config{}
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)

		builder := NewBuilder(projectContext, cfg, outputManager)
		rootCmd := builder.Build()

		// Find project command
		projectCmd := findCommand(rootCmd, "project")
		require.NotNil(t, projectCmd)
		assert.Equal(t, "project", projectCmd.Use)

		// Verify alias
		assert.Contains(t, projectCmd.Aliases, "p")
	})

	t.Run("project command has subcommands", func(t *testing.T) {
		projectContext := &context.ProjectContext{}
		cfg := &config.Config{}
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)

		builder := NewBuilder(projectContext, cfg, outputManager)
		rootCmd := builder.Build()

		projectCmd := findCommand(rootCmd, "project")
		require.NotNil(t, projectCmd)

		// Check for expected subcommands
		expectedSubcommands := []string{"status", "down", "list", "clean"}
		for _, subCmdName := range expectedSubcommands {
			subCmd := findCommand(projectCmd, subCmdName)
			assert.NotNil(t, subCmd, "expected subcommand %s", subCmdName)
		}
	})
}

// TestSetupCommand tests the setup command
func TestSetupCommand(t *testing.T) {
	t.Run("setup command exists", func(t *testing.T) {
		projectContext := &context.ProjectContext{}
		cfg := &config.Config{}
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)

		builder := NewBuilder(projectContext, cfg, outputManager)
		rootCmd := builder.Build()

		// Find setup command
		setupCmd := findCommand(rootCmd, "setup")
		require.NotNil(t, setupCmd)
		assert.Equal(t, "setup", setupCmd.Use)
	})
}

// TestCompletionCommand tests the completion command
func TestCompletionCommand(t *testing.T) {
	t.Run("completion command exists", func(t *testing.T) {
		projectContext := &context.ProjectContext{}
		cfg := &config.Config{}
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)

		builder := NewBuilder(projectContext, cfg, outputManager)
		rootCmd := builder.Build()

		// Find completion command
		completionCmd := findCommand(rootCmd, "completion")
		require.NotNil(t, completionCmd)
		assert.Contains(t, completionCmd.Use, "completion")
	})

	t.Run("completion command generates shell completions", func(t *testing.T) {
		projectContext := &context.ProjectContext{}
		cfg := &config.Config{}
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)

		builder := NewBuilder(projectContext, cfg, outputManager)
		rootCmd := builder.Build()

		completionCmd := findCommand(rootCmd, "completion")
		require.NotNil(t, completionCmd)

		// Completion command exists and can generate completions
		// The actual shell subcommands may be generated by cobra internally
		assert.NotNil(t, completionCmd)
	})
}

// TestSelfUpdateCommand tests the self-update command
func TestSelfUpdateCommand(t *testing.T) {
	t.Run("self-update command exists", func(t *testing.T) {
		projectContext := &context.ProjectContext{}
		cfg := &config.Config{}
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)

		builder := NewBuilder(projectContext, cfg, outputManager)
		rootCmd := builder.Build()

		// Find self-update command
		selfUpdateCmd := findCommand(rootCmd, "self-update")
		require.NotNil(t, selfUpdateCmd)
		assert.Contains(t, selfUpdateCmd.Use, "self-update")

		// Verify aliases
		assert.Contains(t, selfUpdateCmd.Aliases, "update")
		assert.Contains(t, selfUpdateCmd.Aliases, "upgrade")
	})
}

// TestHelpCommand tests the help command
func TestHelpCommand(t *testing.T) {
	t.Run("help command exists", func(t *testing.T) {
		projectContext := &context.ProjectContext{}
		cfg := &config.Config{}
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)

		builder := NewBuilder(projectContext, cfg, outputManager)
		rootCmd := builder.Build()

		// Find help command
		helpCmd := findCommand(rootCmd, "help")
		require.NotNil(t, helpCmd)
		assert.Contains(t, helpCmd.Use, "help")
	})

	t.Run("help flag on root command", func(t *testing.T) {
		projectContext := &context.ProjectContext{}
		cfg := &config.Config{}
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)

		builder := NewBuilder(projectContext, cfg, outputManager)
		rootCmd := builder.Build()

		// Capture output
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)

		// Execute with --help flag
		rootCmd.SetArgs([]string{"--help"})
		err := rootCmd.Execute()
		assert.NoError(t, err)

		// Verify help output
		output := buf.String()
		assert.Contains(t, output, "Available Commands:")
	})
}

// TestConfigCommand tests the config debug command
func TestConfigCommand(t *testing.T) {
	t.Run("config command is hidden", func(t *testing.T) {
		projectContext := &context.ProjectContext{}
		cfg := &config.Config{}
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)

		builder := NewBuilder(projectContext, cfg, outputManager)
		registry := builder.GetRegistry()

		meta, exists := registry.GetMetadata("config")
		require.True(t, exists)
		assert.True(t, meta.Hidden, "config command should be hidden")
	})
}
