package app

import (
	"testing"

	"github.com/ivannovak/glide/v3/internal/context"
	"github.com/ivannovak/glide/v3/pkg/output"
	"github.com/ivannovak/glide/v3/tests/testutil"
	"github.com/stretchr/testify/assert"
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
		buf := testutil.NewTestWriter()
		ctx := testutil.NewTestContext(
			testutil.WithProjectRoot("/test/project"),
		)
		cfg := testutil.NewTestConfig(
			testutil.WithDefaultProject("test"),
		)

		app := NewApplication(
			WithWriter(buf),
			WithProjectContext(ctx),
			WithConfig(cfg),
			WithOutputFormat(output.FormatJSON, true, true),
		)

		testutil.AssertNotNil(t, app, "app should be created")
		assert.Equal(t, buf, app.Writer)
		assert.Equal(t, ctx, app.ProjectContext)
		assert.Equal(t, cfg, app.Config)
		testutil.AssertNotNil(t, app.OutputManager, "output manager should be created")
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
		ctx := testutil.NewTestContext(
			testutil.WithWorkingDir("/working"),
			testutil.WithProjectRoot("/project"),
			testutil.WithDevelopmentMode(context.ModeMultiWorktree),
		)
		app := NewApplication(WithProjectContext(ctx))

		assert.Equal(t, ctx, app.ProjectContext)
	})

	t.Run("WithConfig", func(t *testing.T) {
		cfg := testutil.NewTestConfig(
			testutil.WithDefaultProject("myproject"),
			testutil.WithProject("myproject", "/path/to/project", "multi-worktree"),
		)
		app := NewApplication(WithConfig(cfg))

		assert.Equal(t, cfg, app.Config)
	})

	t.Run("WithWriter updates OutputManager", func(t *testing.T) {
		buf := testutil.NewTestWriter()
		app := NewApplication(
			WithOutputFormat(output.FormatTable, false, false),
			WithWriter(buf),
		)

		assert.Equal(t, buf, app.Writer)
		// Verify the writer was propagated to OutputManager
		app.OutputManager.Raw("test")
		testutil.AssertEqual(t, "test", buf.String(), "output should match")
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
		// Create a complete application setup using testutil
		buf := testutil.NewTestWriter()
		ctx := testutil.NewTestContext(
			testutil.WithProjectRoot("/test/project"),
			testutil.WithWorkingDir("/test/project"),
			testutil.WithDevelopmentMode(context.ModeSingleRepo),
			testutil.WithLocation(context.LocationProject),
		)
		cfg := testutil.NewTestConfig(
			testutil.WithDefaultProject("test"),
			testutil.WithTestDefaults(true, true),
		)

		app := NewApplication(
			WithProjectContext(ctx),
			WithConfig(cfg),
			WithWriter(buf),
			WithOutputFormat(output.FormatTable, false, false),
		)

		// Verify all components are properly initialized
		testutil.RequireNotNil(t, app, "app must be created")
		testutil.RequireNotNil(t, app.OutputManager, "output manager must exist")
		testutil.RequireNotNil(t, app.ProjectContext, "project context must exist")
		testutil.RequireNotNil(t, app.Config, "config must exist")
		testutil.RequireNotNil(t, app.ShellExecutor, "shell executor must exist")

		// Test output functionality
		err := app.OutputManager.Info("Test message")
		testutil.AssertNoError(t, err, "output should succeed")
		testutil.AssertContains(t, buf.String(), "Test message", "should contain message")

		// Test accessor methods
		testutil.AssertNotNil(t, app.GetShellExecutor(), "should have executor")
		testutil.AssertNotNil(t, app.GetConfigLoader(), "should have config loader")
	})
}
