package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ivannovak/glide/v2/internal/config"
	"github.com/ivannovak/glide/v2/internal/context"
	"github.com/ivannovak/glide/v2/pkg/output"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCommandTreeBuilding tests the command tree construction
func TestCommandTreeBuilding(t *testing.T) {
	t.Run("root command", func(t *testing.T) {
		projectContext := &context.ProjectContext{}
		cfg := &config.Config{}
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)

		builder := NewBuilder(projectContext, cfg, outputManager)
		rootCmd := builder.Build()

		assert.NotNil(t, rootCmd)
		assert.Equal(t, "glide", rootCmd.Use)
		assert.NotEmpty(t, rootCmd.Short)
		assert.NotEmpty(t, rootCmd.Long)
		assert.True(t, rootCmd.SilenceErrors)
		assert.True(t, rootCmd.SilenceUsage)
	})

	t.Run("subcommands", func(t *testing.T) {
		projectContext := &context.ProjectContext{}
		cfg := &config.Config{}
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)

		builder := NewBuilder(projectContext, cfg, outputManager)
		rootCmd := builder.Build()

		// Verify expected subcommands are present
		expectedCommands := []string{
			"setup",
			"plugins",
			"config",
			"completion",
			"project",
			"version",
			"help",
			"self-update",
		}

		for _, cmdName := range expectedCommands {
			cmd := findCommand(rootCmd, cmdName)
			assert.NotNil(t, cmd, "expected command %s to be registered", cmdName)
		}
	})

	t.Run("nested commands", func(t *testing.T) {
		projectContext := &context.ProjectContext{}
		cfg := &config.Config{}
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)

		builder := NewBuilder(projectContext, cfg, outputManager)
		rootCmd := builder.Build()

		// Check for nested commands under "project"
		projectCmd := findCommand(rootCmd, "project")
		require.NotNil(t, projectCmd)

		// Project command has subcommands: status, down, list, clean
		expectedSubcommands := []string{"status", "down", "list", "clean"}
		for _, subCmd := range expectedSubcommands {
			cmd := findCommand(projectCmd, subCmd)
			assert.NotNil(t, cmd, "expected subcommand %s under project", subCmd)
		}
	})

	t.Run("plugin commands", func(t *testing.T) {
		// Plugin commands are added dynamically, so we test the mechanism
		projectContext := &context.ProjectContext{}
		cfg := &config.Config{}
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)

		builder := NewBuilder(projectContext, cfg, outputManager)
		registry := builder.GetRegistry()

		// Add a mock plugin command
		pluginFactory := func() *cobra.Command {
			return &cobra.Command{
				Use:   "plugin-test",
				Short: "Test plugin command",
			}
		}

		err := registry.Register("plugin-test", pluginFactory, Metadata{
			Name:        "plugin-test",
			Category:    CategoryPlugin,
			Description: "Test plugin command",
		})
		require.NoError(t, err)

		// Verify it's registered
		_, exists := registry.Get("plugin-test")
		assert.True(t, exists)

		// Verify it appears in plugin category
		pluginCommands := registry.GetByCategory(CategoryPlugin)
		assert.Contains(t, pluginCommands, "plugin-test")
	})
}

// TestAliasRegistration tests command alias functionality
func TestAliasRegistration(t *testing.T) {
	t.Run("global aliases", func(t *testing.T) {
		registry := NewRegistry()

		factory := func() *cobra.Command {
			return &cobra.Command{Use: "self-update"}
		}

		metadata := Metadata{
			Name:        "self-update",
			Category:    CategoryCore,
			Description: "Update Glide",
			Aliases:     []string{"update", "upgrade"},
		}

		err := registry.Register("self-update", factory, metadata)
		require.NoError(t, err)

		// Verify command can be retrieved by all aliases
		for _, alias := range []string{"update", "upgrade"} {
			_, exists := registry.Get(alias)
			assert.True(t, exists, "alias %s should resolve to command", alias)
		}

		// Verify metadata includes aliases
		meta, exists := registry.GetMetadata("self-update")
		require.True(t, exists)
		assert.Equal(t, []string{"update", "upgrade"}, meta.Aliases)
	})

	t.Run("command aliases", func(t *testing.T) {
		registry := NewRegistry()

		factory := func() *cobra.Command {
			return &cobra.Command{Use: "plugins"}
		}

		metadata := Metadata{
			Name:        "plugins",
			Category:    CategoryCore,
			Description: "Manage plugins",
			Aliases:     []string{"plugin"},
		}

		err := registry.Register("plugins", factory, metadata)
		require.NoError(t, err)

		// Verify alias resolution
		canonical, isAlias := registry.ResolveAlias("plugin")
		assert.True(t, isAlias)
		assert.Equal(t, "plugins", canonical)
	})

	t.Run("alias conflicts", func(t *testing.T) {
		registry := NewRegistry()

		factory1 := func() *cobra.Command {
			return &cobra.Command{Use: "command1"}
		}

		factory2 := func() *cobra.Command {
			return &cobra.Command{Use: "command2"}
		}

		// Register first command with alias
		err := registry.Register("command1", factory1, Metadata{
			Name:    "command1",
			Aliases: []string{"c"},
		})
		require.NoError(t, err)

		// Try to register second command with same alias - should fail
		err = registry.Register("command2", factory2, Metadata{
			Name:    "command2",
			Aliases: []string{"c"},
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "alias c already registered")

		// Try to register command with name that conflicts with existing alias
		err = registry.Register("c", factory2, Metadata{
			Name: "c",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "conflicts with existing alias")
	})
}

// TestFlagRegistration tests flag handling
func TestFlagRegistration(t *testing.T) {
	t.Run("global flags", func(t *testing.T) {
		projectContext := &context.ProjectContext{}
		cfg := &config.Config{}
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)

		builder := NewBuilder(projectContext, cfg, outputManager)
		rootCmd := builder.Build()

		// Note: Global flags are typically added at a higher level (in main.go)
		// This test verifies that the root command can have flags added
		rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose output")
		rootCmd.PersistentFlags().String("config", "", "Config file path")

		assert.NotNil(t, rootCmd.PersistentFlags().Lookup("verbose"))
		assert.NotNil(t, rootCmd.PersistentFlags().Lookup("config"))
	})

	t.Run("command flags", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:   "test",
			Short: "Test command",
		}

		cmd.Flags().String("output", "table", "Output format")
		cmd.Flags().Bool("force", false, "Force operation")

		assert.NotNil(t, cmd.Flags().Lookup("output"))
		assert.NotNil(t, cmd.Flags().Lookup("force"))
	})

	t.Run("flag inheritance", func(t *testing.T) {
		rootCmd := &cobra.Command{
			Use: "root",
		}

		rootCmd.PersistentFlags().Bool("debug", false, "Debug mode")

		subCmd := &cobra.Command{
			Use:   "sub",
			Short: "Subcommand",
		}

		rootCmd.AddCommand(subCmd)

		// Persistent flags should be inherited by subcommands
		assert.NotNil(t, subCmd.InheritedFlags().Lookup("debug"))
	})

	t.Run("flag conflicts", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "test",
		}

		cmd.Flags().String("output", "table", "Output format")

		// Cobra panics when you try to add a duplicate flag
		// This is expected behavior - we should avoid duplicates
		flag := cmd.Flags().Lookup("output")
		assert.NotNil(t, flag)
		assert.Equal(t, "Output format", flag.Usage)

		// Verify duplicate flag causes panic
		assert.Panics(t, func() {
			cmd.Flags().String("output", "json", "Different description")
		}, "Adding duplicate flag should panic")
	})
}

// TestYAMLCommandRegistration tests YAML-defined command registration
func TestYAMLCommandRegistration(t *testing.T) {
	t.Run("simple YAML command", func(t *testing.T) {
		registry := NewRegistry()

		yamlCmd := &config.Command{
			Cmd:         "echo hello",
			Description: "Say hello",
			Help:        "Prints hello to stdout",
		}

		err := registry.AddYAMLCommand("hello", yamlCmd)
		require.NoError(t, err)

		// Verify command is registered
		factory, exists := registry.Get("hello")
		assert.True(t, exists)
		assert.NotNil(t, factory)

		// Verify metadata
		meta, exists := registry.GetMetadata("hello")
		require.True(t, exists)
		assert.Equal(t, "hello", meta.Name)
		assert.Equal(t, CategoryYAML, meta.Category)
	})

	t.Run("YAML command with alias", func(t *testing.T) {
		registry := NewRegistry()

		yamlCmd := &config.Command{
			Cmd:         "docker-compose up",
			Description: "Start containers",
			Alias:       "up",
		}

		err := registry.AddYAMLCommand("docker-up", yamlCmd)
		require.NoError(t, err)

		// Verify both name and alias work
		_, exists := registry.Get("docker-up")
		assert.True(t, exists)

		_, exists = registry.Get("up")
		assert.True(t, exists)
	})

	t.Run("YAML command with category", func(t *testing.T) {
		registry := NewRegistry()

		yamlCmd := &config.Command{
			Cmd:         "docker ps",
			Description: "List containers",
			Category:    "docker",
		}

		err := registry.AddYAMLCommand("ps", yamlCmd)
		require.NoError(t, err)

		// Verify category mapping
		meta, exists := registry.GetMetadata("ps")
		require.True(t, exists)
		assert.Equal(t, CategoryDocker, meta.Category)
	})

	t.Run("YAML command annotations", func(t *testing.T) {
		registry := NewRegistry()

		yamlCmd := &config.Command{
			Cmd:         "npm test",
			Description: "Run tests",
		}

		err := registry.AddYAMLCommand("test", yamlCmd)
		require.NoError(t, err)

		// Create the command and check annotations
		factory, _ := registry.Get("test")
		cmd := factory()

		// YAML commands should have the yaml_command annotation
		assert.Equal(t, "true", cmd.Annotations["yaml_command"])

		// YAML commands should disable flag parsing for pass-through
		assert.True(t, cmd.DisableFlagParsing)
	})
}

// TestProtectedCommands tests that core commands cannot be overridden
func TestProtectedCommands(t *testing.T) {
	protectedCommands := []string{
		"help", "setup", "plugins", "plugin", "self-update",
		"update", "upgrade", "version", "completion", "global",
		"config", "context",
	}

	for _, cmdName := range protectedCommands {
		t.Run(cmdName, func(t *testing.T) {
			assert.True(t, isProtectedCommand(cmdName),
				"command %s should be protected", cmdName)
		})
	}

	t.Run("non-protected command", func(t *testing.T) {
		assert.False(t, isProtectedCommand("my-custom-command"))
	})
}

// TestCommandMetadata tests metadata storage and retrieval
func TestCommandMetadata(t *testing.T) {
	t.Run("metadata storage", func(t *testing.T) {
		registry := NewRegistry()

		factory := func() *cobra.Command {
			return &cobra.Command{Use: "test"}
		}

		metadata := Metadata{
			Name:        "test",
			Category:    CategoryCore,
			Description: "Test command",
			Aliases:     []string{"t"},
			Hidden:      true,
		}

		err := registry.Register("test", factory, metadata)
		require.NoError(t, err)

		// Retrieve and verify metadata
		meta, exists := registry.GetMetadata("test")
		require.True(t, exists)
		assert.Equal(t, "test", meta.Name)
		assert.Equal(t, CategoryCore, meta.Category)
		assert.Equal(t, "Test command", meta.Description)
		assert.Equal(t, []string{"t"}, meta.Aliases)
		assert.True(t, meta.Hidden)
	})

	t.Run("metadata by alias", func(t *testing.T) {
		registry := NewRegistry()

		factory := func() *cobra.Command {
			return &cobra.Command{Use: "test"}
		}

		metadata := Metadata{
			Name:    "test",
			Aliases: []string{"t"},
		}

		err := registry.Register("test", factory, metadata)
		require.NoError(t, err)

		// Retrieve metadata by alias
		meta, exists := registry.GetMetadata("t")
		require.True(t, exists)
		assert.Equal(t, "test", meta.Name)
	})

	t.Run("category-based retrieval", func(t *testing.T) {
		registry := NewRegistry()

		factory := func() *cobra.Command {
			return &cobra.Command{Use: "test"}
		}

		// Register commands in different categories
		registry.Register("core1", factory, Metadata{
			Name:     "core1",
			Category: CategoryCore,
		})
		registry.Register("core2", factory, Metadata{
			Name:     "core2",
			Category: CategoryCore,
		})
		registry.Register("docker1", factory, Metadata{
			Name:     "docker1",
			Category: CategoryDocker,
		})

		// Retrieve by category
		coreCommands := registry.GetByCategory(CategoryCore)
		assert.Len(t, coreCommands, 2)
		assert.Contains(t, coreCommands, "core1")
		assert.Contains(t, coreCommands, "core2")

		dockerCommands := registry.GetByCategory(CategoryDocker)
		assert.Len(t, dockerCommands, 1)
		assert.Contains(t, dockerCommands, "docker1")
	})
}

// TestDebugCommandsRegistration tests debug command registration
func TestDebugCommandsRegistration(t *testing.T) {
	t.Run("debug commands present with context", func(t *testing.T) {
		projectContext := &context.ProjectContext{}
		cfg := &config.Config{}
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)

		builder := NewBuilder(projectContext, cfg, outputManager)
		rootCmd := builder.Build()

		// Debug commands should be present
		debugCommands := []string{"context", "shell-test", "docker-test", "container-test"}
		for _, cmdName := range debugCommands {
			cmd := findCommand(rootCmd, cmdName)
			assert.NotNil(t, cmd, "debug command %s should be present", cmdName)
		}
	})

	t.Run("debug commands absent without context", func(t *testing.T) {
		cfg := &config.Config{}
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)

		builder := NewBuilder(nil, cfg, outputManager)
		rootCmd := builder.Build()

		// Debug commands should NOT be present
		debugCommands := []string{"context", "shell-test", "docker-test", "container-test"}
		for _, cmdName := range debugCommands {
			cmd := findCommand(rootCmd, cmdName)
			assert.Nil(t, cmd, "debug command %s should not be present without context", cmdName)
		}
	})
}

// TestHiddenCommands tests that hidden commands are properly marked
func TestHiddenCommands(t *testing.T) {
	t.Run("hidden metadata propagates to cobra command", func(t *testing.T) {
		registry := NewRegistry()

		factory := func() *cobra.Command {
			return &cobra.Command{Use: "secret"}
		}

		metadata := Metadata{
			Name:     "secret",
			Category: CategoryDebug,
			Hidden:   true,
		}

		err := registry.Register("secret", factory, metadata)
		require.NoError(t, err)

		// Create command and verify it's hidden
		commands := registry.CreateAll()
		require.Len(t, commands, 1)
		assert.True(t, commands[0].Hidden)
	})

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

// TestYAMLCommandLoading tests the YAML command discovery and loading
func TestYAMLCommandLoading(t *testing.T) {
	t.Run("load YAML commands from .glide.yml", func(t *testing.T) {
		// Create a temporary directory with .glide.yml
		tmpDir := t.TempDir()

		glideYAML := `
commands:
  hello:
    cmd: echo hello
    description: Say hello
  world:
    cmd: echo world
    description: Say world
    alias: w
`
		err := os.WriteFile(filepath.Join(tmpDir, ".glide.yml"), []byte(glideYAML), 0644)
		require.NoError(t, err)

		// Change to temp directory
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)
		os.Chdir(tmpDir)

		projectContext := &context.ProjectContext{}
		cfg := &config.Config{}
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)

		builder := NewBuilder(projectContext, cfg, outputManager)
		builder.loadYAMLCommands()

		registry := builder.GetRegistry()

		// Verify YAML commands are loaded
		_, exists := registry.Get("hello")
		assert.True(t, exists, "hello command should be loaded from .glide.yml")

		_, exists = registry.Get("world")
		assert.True(t, exists, "world command should be loaded from .glide.yml")

		// Verify alias works
		_, exists = registry.Get("w")
		assert.True(t, exists, "alias 'w' should resolve to 'world'")
	})

	t.Run("protected commands cannot be overridden by YAML", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Try to override a protected command
		glideYAML := `
commands:
  help:
    cmd: echo "my custom help"
    description: Custom help
`
		err := os.WriteFile(filepath.Join(tmpDir, ".glide.yml"), []byte(glideYAML), 0644)
		require.NoError(t, err)

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)
		os.Chdir(tmpDir)

		projectContext := &context.ProjectContext{}
		cfg := &config.Config{}
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)

		builder := NewBuilder(projectContext, cfg, outputManager)
		builder.loadYAMLCommands()

		registry := builder.GetRegistry()

		// Verify the help command is still the original, not the YAML version
		factory, exists := registry.Get("help")
		require.True(t, exists)

		cmd := factory()
		// The original help command won't have DisableFlagParsing
		// but YAML commands do
		assert.False(t, cmd.DisableFlagParsing,
			"help command should not be overridden by YAML")
	})
}
