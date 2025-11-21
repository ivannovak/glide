package app

import (
	"bytes"
	"testing"

	"github.com/ivannovak/glide/internal/config"
	"github.com/ivannovak/glide/internal/context"
	"github.com/ivannovak/glide/pkg/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewApplication(t *testing.T) {
	t.Run("creates application with defaults", func(t *testing.T) {
		app := NewApplication()

		assert.NotNil(t, app)
		assert.NotNil(t, app.OutputManager)
		assert.NotNil(t, app.ShellExecutor)
		assert.NotNil(t, app.Writer)
	})

	t.Run("creates application with options", func(t *testing.T) {
		buf := &bytes.Buffer{}
		ctx := &context.ProjectContext{
			ProjectRoot: "/test/project",
		}
		cfg := &config.Config{
			DefaultProject: "test",
		}

		app := NewApplication(
			WithWriter(buf),
			WithProjectContext(ctx),
			WithConfig(cfg),
			WithOutputFormat(output.FormatJSON, true, true),
		)

		assert.NotNil(t, app)
		assert.Equal(t, buf, app.Writer)
		assert.Equal(t, ctx, app.ProjectContext)
		assert.Equal(t, cfg, app.Config)
		assert.NotNil(t, app.OutputManager)
		assert.True(t, app.OutputManager.IsQuiet())
	})
}

func TestApplicationOptions(t *testing.T) {
	t.Run("WithOutputManager", func(t *testing.T) {
		manager := output.NewManager(output.FormatYAML, false, false, nil)
		app := NewApplication(WithOutputManager(manager))

		assert.Equal(t, manager, app.OutputManager)
		assert.Equal(t, output.FormatYAML, app.OutputManager.GetFormat())
	})

	t.Run("WithOutputFormat", func(t *testing.T) {
		app := NewApplication(WithOutputFormat(output.FormatJSON, true, false))

		assert.NotNil(t, app.OutputManager)
		assert.Equal(t, output.FormatJSON, app.OutputManager.GetFormat())
		assert.True(t, app.OutputManager.IsQuiet())
	})

	t.Run("WithProjectContext", func(t *testing.T) {
		ctx := &context.ProjectContext{
			WorkingDir:      "/working",
			ProjectRoot:     "/project",
			DevelopmentMode: context.ModeMultiWorktree,
		}
		app := NewApplication(WithProjectContext(ctx))

		assert.Equal(t, ctx, app.ProjectContext)
	})

	t.Run("WithConfig", func(t *testing.T) {
		cfg := &config.Config{
			DefaultProject: "myproject",
			Projects: map[string]config.ProjectConfig{
				"myproject": {
					Path: "/path/to/project",
					Mode: "multi-worktree",
				},
			},
		}
		app := NewApplication(WithConfig(cfg))

		assert.Equal(t, cfg, app.Config)
	})

	t.Run("WithWriter updates OutputManager", func(t *testing.T) {
		buf := &bytes.Buffer{}
		app := NewApplication(
			WithOutputFormat(output.FormatTable, false, false),
			WithWriter(buf),
		)

		assert.Equal(t, buf, app.Writer)
		// Verify the writer was propagated to OutputManager
		app.OutputManager.Raw("test")
		assert.Equal(t, "test", buf.String())
	})
}

func TestGetShellExecutor(t *testing.T) {
	t.Run("returns default executor", func(t *testing.T) {
		app := NewApplication()
		executor := app.GetShellExecutor()

		assert.NotNil(t, executor)
	})
}

func TestGetConfigLoader(t *testing.T) {
	t.Run("creates loader when needed", func(t *testing.T) {
		app := NewApplication()
		loader := app.GetConfigLoader()

		assert.NotNil(t, loader)
		assert.Equal(t, loader, app.ConfigLoader) // Should cache it
	})

	t.Run("returns existing loader", func(t *testing.T) {
		app := NewApplication()

		loader1 := app.GetConfigLoader()
		loader2 := app.GetConfigLoader()

		assert.Same(t, loader1, loader2)
	})
}

func TestApplicationIntegration(t *testing.T) {
	t.Run("full application setup", func(t *testing.T) {
		// Create a complete application setup
		buf := &bytes.Buffer{}
		ctx := &context.ProjectContext{
			ProjectRoot:     "/test/project",
			WorkingDir:      "/test/project",
			DevelopmentMode: context.ModeSingleRepo,
			Location:        context.LocationProject,
		}
		cfg := &config.Config{
			DefaultProject: "test",
			Defaults: config.DefaultsConfig{
				Test: config.TestDefaults{
					Parallel: true,
					Coverage: true,
				},
			},
		}

		app := NewApplication(
			WithProjectContext(ctx),
			WithConfig(cfg),
			WithWriter(buf),
			WithOutputFormat(output.FormatTable, false, false),
		)

		// Verify all components are properly initialized
		require.NotNil(t, app)
		require.NotNil(t, app.OutputManager)
		require.NotNil(t, app.ProjectContext)
		require.NotNil(t, app.Config)
		require.NotNil(t, app.ShellExecutor)

		// Test output functionality
		err := app.OutputManager.Info("Test message")
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Test message")

		// Test accessor methods
		assert.NotNil(t, app.GetShellExecutor())
		assert.NotNil(t, app.GetConfigLoader())
	})
}
