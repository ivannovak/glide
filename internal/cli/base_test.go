package cli

import (
	"bytes"
	"testing"

	"github.com/ivannovak/glide/v2/internal/config"
	"github.com/ivannovak/glide/v2/internal/context"
	"github.com/ivannovak/glide/v2/pkg/app"
	"github.com/ivannovak/glide/v2/pkg/output"
	"github.com/stretchr/testify/assert"
)

func TestNewBaseCommand(t *testing.T) {
	t.Run("creates base command with application", func(t *testing.T) {
		application := app.NewApplication()
		base := NewBaseCommand(application)

		assert.NotNil(t, base)
		assert.NotNil(t, base.app)
		assert.Equal(t, application, base.app)
	})
}

func TestBaseCommandOutput(t *testing.T) {
	t.Run("returns output manager from application", func(t *testing.T) {
		buf := &bytes.Buffer{}
		application := app.NewApplication(
			app.WithWriter(buf),
			app.WithOutputFormat(output.FormatJSON, false, false),
		)
		base := NewBaseCommand(application)

		manager := base.Output()
		assert.NotNil(t, manager)
		assert.Equal(t, application.OutputManager, manager)
		assert.Equal(t, output.FormatJSON, manager.GetFormat())
	})

	t.Run("output manager writes to application writer", func(t *testing.T) {
		buf := &bytes.Buffer{}
		application := app.NewApplication(
			app.WithWriter(buf),
		)
		base := NewBaseCommand(application)

		base.Output().Raw("test output")
		assert.Equal(t, "test output", buf.String())
	})
}

func TestBaseCommandContext(t *testing.T) {
	t.Run("returns project context from application", func(t *testing.T) {
		ctx := &context.ProjectContext{
			ProjectRoot:     "/test/project",
			WorkingDir:      "/test/working",
			DevelopmentMode: context.ModeMultiWorktree,
		}
		application := app.NewApplication(
			app.WithProjectContext(ctx),
		)
		base := NewBaseCommand(application)

		projectCtx := base.Context()
		assert.NotNil(t, projectCtx)
		assert.Equal(t, ctx, projectCtx)
		assert.Equal(t, "/test/project", projectCtx.ProjectRoot)
	})

	t.Run("returns nil when no context", func(t *testing.T) {
		application := app.NewApplication()
		base := NewBaseCommand(application)

		projectCtx := base.Context()
		assert.Nil(t, projectCtx)
	})
}

func TestBaseCommandConfig(t *testing.T) {
	t.Run("returns config from application", func(t *testing.T) {
		cfg := &config.Config{
			DefaultProject: "test-project",
			Projects: map[string]config.ProjectConfig{
				"test-project": {
					Path: "/path/to/test",
					Mode: "single-repo",
				},
			},
		}
		application := app.NewApplication(
			app.WithConfig(cfg),
		)
		base := NewBaseCommand(application)

		config := base.Config()
		assert.NotNil(t, config)
		assert.Equal(t, cfg, config)
		assert.Equal(t, "test-project", config.DefaultProject)
	})

	t.Run("returns nil when no config", func(t *testing.T) {
		application := app.NewApplication()
		base := NewBaseCommand(application)

		config := base.Config()
		assert.Nil(t, config)
	})
}

func TestBaseCommandApplication(t *testing.T) {
	t.Run("returns the full application", func(t *testing.T) {
		application := app.NewApplication()
		base := NewBaseCommand(application)

		app := base.Application()
		assert.NotNil(t, app)
		assert.Equal(t, application, app)
	})
}

func TestBaseCommandIntegration(t *testing.T) {
	t.Run("base command provides access to all dependencies", func(t *testing.T) {
		// Set up a complete application
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

		application := app.NewApplication(
			app.WithProjectContext(ctx),
			app.WithConfig(cfg),
			app.WithWriter(buf),
			app.WithOutputFormat(output.FormatTable, false, false),
		)

		base := NewBaseCommand(application)

		// Verify all accessors work
		assert.NotNil(t, base.Output())
		assert.NotNil(t, base.Context())
		assert.NotNil(t, base.Config())
		assert.NotNil(t, base.Application())

		// Verify they return the correct instances
		assert.Equal(t, application.OutputManager, base.Output())
		assert.Equal(t, ctx, base.Context())
		assert.Equal(t, cfg, base.Config())
		assert.Equal(t, application, base.Application())

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
		application := app.NewApplication(
			app.WithProjectContext(&context.ProjectContext{
				ProjectRoot: "/embed/test",
			}),
		)

		testCmd := &TestCommand{
			BaseCommand: NewBaseCommand(application),
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

			// Access the full application if needed
			app := cmd.Application()
			if app.GetShellExecutor() != nil {
				cmd.Output().Success("Shell executor available")
			}

			return nil
		}

		// Set up the command
		buf := &bytes.Buffer{}
		application := app.NewApplication(
			app.WithProjectContext(&context.ProjectContext{
				ProjectRoot: "/example",
			}),
			app.WithConfig(&config.Config{
				DefaultProject: "example-project",
			}),
			app.WithWriter(buf),
		)

		exampleCmd := &ExampleCommand{
			BaseCommand: NewBaseCommand(application),
		}

		// Execute the command
		err := executeExample(exampleCmd)
		assert.NoError(t, err)

		// Verify output
		output := buf.String()
		assert.Contains(t, output, "Starting example command")
		assert.Contains(t, output, "Default project: example-project")
		assert.Contains(t, output, "Shell executor available")
	})
}
