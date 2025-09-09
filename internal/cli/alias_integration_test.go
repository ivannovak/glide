package cli

import (
	"bytes"
	"testing"

	"github.com/ivannovak/glide/pkg/app"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestCLI_AliasIntegration(t *testing.T) {
	// Create a test application
	application := app.NewApplication()

	// Create CLI and builder
	cli := New(application)

	// Build root command
	rootCmd := &cobra.Command{
		Use:   "glid",
		Short: "Test CLI",
	}

	// Add local commands which should include aliases
	cli.AddLocalCommands(rootCmd)

	// Test that core commands exist with their aliases
	// Note: Docker and dev commands have been moved to plugins
	versionCmd := findCommand(rootCmd, "version")
	assert.NotNil(t, versionCmd)

	setupCmd := findCommand(rootCmd, "setup")
	assert.NotNil(t, setupCmd)

	pluginsCmd := findCommand(rootCmd, "plugins")
	assert.NotNil(t, pluginsCmd)
}

func TestCLI_AliasExecution(t *testing.T) {
	// Create a test application
	application := app.NewApplication()

	// Create a builder with a test command
	builder := NewBuilder(application)

	// Register a test command with alias
	testExecuted := false
	builder.registry.Register("hello", func() *cobra.Command {
		return &cobra.Command{
			Use:   "hello",
			Short: "Test command",
			Run: func(cmd *cobra.Command, args []string) {
				testExecuted = true
			},
		}
	}, Metadata{
		Name:    "hello",
		Aliases: []string{"h"},
	})

	// Build the root command
	rootCmd := &cobra.Command{
		Use:   "test",
		Short: "Test app",
	}

	// Add commands from registry
	for _, cmd := range builder.registry.CreateAll() {
		rootCmd.AddCommand(cmd)
	}

	// Test executing with the alias
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"h"})

	err := rootCmd.Execute()
	assert.NoError(t, err)
	assert.True(t, testExecuted, "Command should have been executed via alias")
}

func TestCLI_AliasHelp(t *testing.T) {
	// Create a builder with test commands
	application := app.NewApplication()
	builder := NewBuilder(application)

	// Register a command with multiple aliases
	builder.registry.Register("migrate", func() *cobra.Command {
		return &cobra.Command{
			Use:   "migrate",
			Short: "Run database migrations",
		}
	}, Metadata{
		Name:    "migrate",
		Aliases: []string{"m", "mig"},
	})

	// Build root command
	rootCmd := &cobra.Command{
		Use:   "test",
		Short: "Test app",
	}

	for _, cmd := range builder.registry.CreateAll() {
		rootCmd.AddCommand(cmd)
	}

	// Capture help output
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"migrate", "--help"})

	err := rootCmd.Execute()
	assert.NoError(t, err)

	// Check that help output is generated
	output := buf.String()
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "Run database migrations")
}

func TestBuilder_RegisteredAliases(t *testing.T) {
	// Create a real builder and verify it has the expected aliases
	application := app.NewApplication()
	builder := NewBuilder(application)

	// Check that self-update has aliases 'update' and 'upgrade'
	meta, exists := builder.registry.GetMetadata("self-update")
	assert.True(t, exists)
	assert.Contains(t, meta.Aliases, "update")
	assert.Contains(t, meta.Aliases, "upgrade")

	// Verify self-update aliases resolve correctly
	factory, exists := builder.registry.Get("update")
	assert.True(t, exists)
	assert.NotNil(t, factory)

	factory, exists = builder.registry.Get("upgrade")
	assert.True(t, exists)
	assert.NotNil(t, factory)
}

func TestCLI_AliasConflictPrevention(t *testing.T) {
	// Test that we cannot register conflicting aliases
	registry := NewRegistry()

	// Register first command with alias
	err := registry.Register("first", func() *cobra.Command {
		return &cobra.Command{Use: "first"}
	}, Metadata{
		Name:    "first",
		Aliases: []string{"f"},
	})
	assert.NoError(t, err)

	// Try to register second command with same alias
	err = registry.Register("second", func() *cobra.Command {
		return &cobra.Command{Use: "second"}
	}, Metadata{
		Name:    "second",
		Aliases: []string{"f"}, // Conflict!
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "alias f already registered")

	// Try to register a command named 'f' (conflicts with alias)
	err = registry.Register("f", func() *cobra.Command {
		return &cobra.Command{Use: "f"}
	}, Metadata{
		Name: "f",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "item name f conflicts with existing alias")
}

// Helper function to find a command by name
func findCommand(root *cobra.Command, name string) *cobra.Command {
	for _, cmd := range root.Commands() {
		if cmd.Name() == name {
			return cmd
		}
	}
	return nil
}

// Test that aliases work correctly in practice
func TestCLI_RealWorldAliasUsage(t *testing.T) {
	// Create a mock root command that simulates the real CLI
	rootCmd := &cobra.Command{
		Use:           "glid",
		Short:         "Glide CLI",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	// Track which command was executed
	var executedCommand string

	// Add artisan command with alias
	artisanCmd := &cobra.Command{
		Use:     "artisan",
		Aliases: []string{"a"},
		Short:   "Run Artisan commands",
		Run: func(cmd *cobra.Command, args []string) {
			executedCommand = "artisan"
		},
	}
	rootCmd.AddCommand(artisanCmd)

	// Add composer command with alias
	composerCmd := &cobra.Command{
		Use:     "composer",
		Aliases: []string{"c"},
		Short:   "Run Composer commands",
		Run: func(cmd *cobra.Command, args []string) {
			executedCommand = "composer"
		},
	}
	rootCmd.AddCommand(composerCmd)

	// Test executing 'a' runs artisan
	executedCommand = ""
	rootCmd.SetArgs([]string{"a"})
	err := rootCmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, "artisan", executedCommand)

	// Test executing 'c' runs composer
	executedCommand = ""
	rootCmd.SetArgs([]string{"c"})
	err = rootCmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, "composer", executedCommand)

	// Test that full names still work
	executedCommand = ""
	rootCmd.SetArgs([]string{"artisan"})
	err = rootCmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, "artisan", executedCommand)
}

// Test that aliases show up in help and completion
func TestCLI_AliasVisibility(t *testing.T) {
	rootCmd := &cobra.Command{
		Use:   "test",
		Short: "Test CLI",
	}

	// Add a command with multiple aliases
	cmd := &cobra.Command{
		Use:     "database",
		Aliases: []string{"db", "d"},
		Short:   "Database operations",
	}
	rootCmd.AddCommand(cmd)

	// Get help output
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"--help"})
	rootCmd.Execute()

	helpText := buf.String()

	// The main command list should show the primary name
	assert.Contains(t, helpText, "database")

	// Test that we can get help for the command using an alias
	buf.Reset()
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"db", "--help"})
	err := rootCmd.Execute()
	assert.NoError(t, err)

	aliasHelpText := buf.String()
	// The help should show something about the command
	assert.NotEmpty(t, aliasHelpText)
	assert.Contains(t, aliasHelpText, "Database operations")
}
