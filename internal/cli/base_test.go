package cli

import (
	"bytes"
	"os"
	"testing"

	"github.com/glide-cli/glide/v3/internal/config"
	"github.com/glide-cli/glide/v3/internal/context"
	"github.com/glide-cli/glide/v3/pkg/output"
	"github.com/stretchr/testify/assert"
)

func TestNewBaseCommand(t *testing.T) {
	t.Run("creates base command with dependencies", func(t *testing.T) {
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)
		projectContext := &context.ProjectContext{}
		cfg := &config.Config{}

		base := NewBaseCommand(outputManager, projectContext, cfg)

		assert.NotNil(t, base)
		assert.NotNil(t, base.Output())
		assert.NotNil(t, base.Context())
		assert.NotNil(t, base.Config())
	})
}

func TestBaseCommandOutput(t *testing.T) {
	t.Run("returns output manager", func(t *testing.T) {
		buf := &bytes.Buffer{}
		outputManager := output.NewManager(output.FormatJSON, false, false, buf)
		projectContext := &context.ProjectContext{}
		cfg := &config.Config{}

		base := NewBaseCommand(outputManager, projectContext, cfg)

		manager := base.Output()
		assert.NotNil(t, manager)
		assert.Equal(t, output.FormatJSON, manager.GetFormat())
	})

	t.Run("output manager writes to provided writer", func(t *testing.T) {
		buf := &bytes.Buffer{}
		outputManager := output.NewManager(output.FormatTable, false, false, buf)
		projectContext := &context.ProjectContext{}
		cfg := &config.Config{}

		base := NewBaseCommand(outputManager, projectContext, cfg)

		base.Output().Raw("test output")
		assert.Equal(t, "test output", buf.String())
	})
}

func TestBaseCommandContext(t *testing.T) {
	t.Run("returns project context", func(t *testing.T) {
		ctx := &context.ProjectContext{
			ProjectRoot:     "/test/project",
			WorkingDir:      "/test/working",
			DevelopmentMode: context.ModeMultiWorktree,
		}
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)
		cfg := &config.Config{}

		base := NewBaseCommand(outputManager, ctx, cfg)

		projectCtx := base.Context()
		assert.NotNil(t, projectCtx)
		assert.Equal(t, ctx, projectCtx)
		assert.Equal(t, "/test/project", projectCtx.ProjectRoot)
	})

	t.Run("handles nil context gracefully", func(t *testing.T) {
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)
		cfg := &config.Config{}

		base := NewBaseCommand(outputManager, nil, cfg)

		projectCtx := base.Context()
		assert.Nil(t, projectCtx)
	})
}

func TestBaseCommandConfig(t *testing.T) {
	t.Run("returns config", func(t *testing.T) {
		cfg := &config.Config{
			DefaultProject: "test-project",
			Projects: map[string]config.ProjectConfig{
				"test-project": {
					Path: "/path/to/test",
					Mode: "single-repo",
				},
			},
		}
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)
		projectContext := &context.ProjectContext{}

		base := NewBaseCommand(outputManager, projectContext, cfg)

		config := base.Config()
		assert.NotNil(t, config)
		assert.Equal(t, cfg, config)
		assert.Equal(t, "test-project", config.DefaultProject)
	})

	t.Run("handles nil config gracefully", func(t *testing.T) {
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)
		projectContext := &context.ProjectContext{}

		base := NewBaseCommand(outputManager, projectContext, nil)

		config := base.Config()
		assert.Nil(t, config)
	})
}

func TestBaseCommandIntegration(t *testing.T) {
	t.Run("base command provides access to all dependencies", func(t *testing.T) {
		// Set up complete dependencies
		buf := &bytes.Buffer{}
		ctx := &context.ProjectContext{
			ProjectRoot:     "/integration/test",
			WorkingDir:      "/integration/test",
			DevelopmentMode: context.ModeSingleRepo,
			DockerRunning:   true,
		}
		cfg := &config.Config{
			DefaultProject: "integration",
			Defaults: config.DefaultsConfig{
				Test: config.TestDefaults{
					Parallel: true,
					Coverage: true,
				},
			},
		}
		outputManager := output.NewManager(output.FormatTable, false, false, buf)

		base := NewBaseCommand(outputManager, ctx, cfg)

		// Verify all accessors work
		assert.NotNil(t, base.Output())
		assert.NotNil(t, base.Context())
		assert.NotNil(t, base.Config())

		// Verify they return the correct instances
		assert.Equal(t, ctx, base.Context())
		assert.Equal(t, cfg, base.Config())

		// Test that output works through the base command
		base.Output().Info("Integration test message")
		assert.Contains(t, buf.String(), "Integration test message")
	})
}

func TestBaseCommandEmbedding(t *testing.T) {
	// Test that BaseCommand can be embedded in other command structs
	type TestCommand struct {
		BaseCommand
		customField string
	}

	t.Run("embedded base command works correctly", func(t *testing.T) {
		outputManager := output.NewManager(output.FormatTable, false, false, os.Stdout)
		projectContext := &context.ProjectContext{
			ProjectRoot: "/embed/test",
		}
		cfg := &config.Config{}

		testCmd := &TestCommand{
			BaseCommand: NewBaseCommand(outputManager, projectContext, cfg),
			customField: "custom value",
		}

		// Can access base command methods
		assert.NotNil(t, testCmd.Output())
		assert.NotNil(t, testCmd.Context())
		assert.Equal(t, "/embed/test", testCmd.Context().ProjectRoot)

		// Can also access custom fields
		assert.Equal(t, "custom value", testCmd.customField)
	})
}

func TestBaseCommandUsageExample(t *testing.T) {
	// This test demonstrates how a command would use BaseCommand
	t.Run("example command implementation", func(t *testing.T) {
		// Create a mock command that uses BaseCommand
		type ExampleCommand struct {
			BaseCommand
		}

		// Implement a method that uses the base command
		executeExample := func(cmd *ExampleCommand) error {
			// Use the output manager
			cmd.Output().Info("Starting example command")

			// Check context
			if cmd.Context() == nil {
				return cmd.Output().Error("No project context available")
			}

			// Use config
			if cmd.Config() != nil {
				cmd.Output().Info("Default project: %s", cmd.Config().DefaultProject)
			}

			return nil
		}

		// Set up the command
		buf := &bytes.Buffer{}
		outputManager := output.NewManager(output.FormatTable, false, false, buf)
		projectContext := &context.ProjectContext{
			ProjectRoot: "/example",
		}
		cfg := &config.Config{
			DefaultProject: "example-project",
		}

		exampleCmd := &ExampleCommand{
			BaseCommand: NewBaseCommand(outputManager, projectContext, cfg),
		}

		// Execute the command
		err := executeExample(exampleCmd)
		assert.NoError(t, err)

		// Verify output
		output := buf.String()
		assert.Contains(t, output, "Starting example command")
		assert.Contains(t, output, "Default project: example-project")
	})
}
