package testutil_test

import (
	"testing"

	"github.com/ivannovak/glide/v2/internal/context"
	"github.com/ivannovak/glide/v2/tests/testutil"
)

// ExampleNewTestContext demonstrates creating a basic test context
func ExampleNewTestContext() {
	ctx := testutil.NewTestContext()

	// Context has sensible defaults
	_ = ctx.ProjectRoot     // "/test/project"
	_ = ctx.DevelopmentMode // ModeSingleRepo
	_ = ctx.Location        // LocationProject
	_ = ctx.DockerRunning   // false
}

// ExampleNewTestContext_customized demonstrates customizing a test context
func ExampleNewTestContext_customized() {
	ctx := testutil.NewTestContext(
		testutil.WithProjectRoot("/my/project"),
		testutil.WithProjectName("my-app"),
		testutil.WithDockerRunning(true),
		testutil.WithComposeFiles("docker-compose.yml"),
	)

	_ = ctx.ProjectRoot   // "/my/project"
	_ = ctx.ProjectName   // "my-app"
	_ = ctx.DockerRunning // true
}

// ExampleNewTestContext_multiWorktree demonstrates multi-worktree configurations
func ExampleNewTestContext_multiWorktree() {
	// Root context
	rootCtx := testutil.NewTestContext(
		testutil.WithMultiWorktreeRoot(),
	)
	_ = rootCtx.IsRoot     // true
	_ = rootCtx.IsMainRepo // false
	_ = rootCtx.IsWorktree // false

	// Main repo context
	mainCtx := testutil.NewTestContext(
		testutil.WithMultiWorktreeMainRepo(),
	)
	_ = mainCtx.IsRoot     // false
	_ = mainCtx.IsMainRepo // true
	_ = mainCtx.IsWorktree // false

	// Worktree context
	wtCtx := testutil.NewTestContext(
		testutil.WithWorktree("feature-123"),
	)
	_ = wtCtx.IsRoot       // false
	_ = wtCtx.IsMainRepo   // false
	_ = wtCtx.IsWorktree   // true
	_ = wtCtx.WorktreeName // "feature-123"
}

// ExampleNewTestConfig demonstrates creating a test configuration
func ExampleNewTestConfig() {
	cfg := testutil.NewTestConfig()

	// Config has sensible defaults
	_ = cfg.DefaultProject                 // "test-project"
	_ = cfg.Defaults.Test.Parallel         // false
	_ = cfg.Defaults.Docker.ComposeTimeout // 30
}

// ExampleNewTestConfig_customized demonstrates customizing a test configuration
func ExampleNewTestConfig_customized() {
	cfg := testutil.NewTestConfig(
		testutil.WithDefaultProject("my-app"),
		testutil.WithProject("my-app", "/path/to/app", "multi-worktree"),
		testutil.WithTestDefaults(true, true), // parallel, coverage
		testutil.WithCommand("deploy", "kubectl apply -f k8s/"),
	)

	_ = cfg.DefaultProject          // "my-app"
	_ = cfg.Projects["my-app"].Path // "/path/to/app"
	_ = cfg.Defaults.Test.Parallel  // true
	_ = cfg.Commands["deploy"]      // "kubectl apply -f k8s/"
}

// Note: Application factories are not provided to avoid import cycles.
// Application is deprecated - use direct dependency injection instead.
// Tests should pass outputManager, projectContext, and config directly to constructors
// using testutil fixtures (NewTestContext, NewTestConfig, output.NewManager).

// TestAssertions demonstrates the assertion helpers
func TestAssertions(t *testing.T) {
	t.Run("error assertions", func(t *testing.T) {
		// AssertNoError
		var err error
		testutil.AssertNoError(t, err, "should have no error")

		// AssertError
		err = context.ErrProjectRootNotFound
		testutil.AssertError(t, err, "should have error")

		// AssertErrorContains
		testutil.AssertErrorContains(t, err, "project root")
	})

	t.Run("equality assertions", func(t *testing.T) {
		testutil.AssertEqual(t, "expected", "expected", "should be equal")
		testutil.AssertNotEqual(t, "a", "b", "should be different")
	})

	t.Run("boolean assertions", func(t *testing.T) {
		testutil.AssertTrue(t, true, "should be true")
		testutil.AssertFalse(t, false, "should be false")
	})

	t.Run("nil assertions", func(t *testing.T) {
		var ptr *string
		testutil.AssertNil(t, ptr, "should be nil")

		str := "test"
		testutil.AssertNotNil(t, &str, "should not be nil")
	})

	t.Run("string assertions", func(t *testing.T) {
		output := "Hello, World!"
		testutil.AssertContains(t, output, "World", "should contain")
		testutil.AssertNotContains(t, output, "Goodbye", "should not contain")
		testutil.AssertEmpty(t, "", "should be empty")
		testutil.AssertNotEmpty(t, output, "should not be empty")
	})

	t.Run("length assertions", func(t *testing.T) {
		slice := []string{"a", "b", "c"}
		testutil.AssertLen(t, slice, 3, "should have length 3")

		m := map[string]int{"a": 1, "b": 2}
		testutil.AssertLen(t, m, 2, "should have length 2")

		testutil.AssertLen(t, "hello", 5, "should have length 5")
	})

	t.Run("struct comparison", func(t *testing.T) {
		type TestStruct struct {
			Name  string
			Value int
		}

		expected := TestStruct{Name: "test", Value: 42}
		actual := TestStruct{Name: "test", Value: 42}

		testutil.AssertStructEqual(t, expected, actual)
	})
}

// TestFixtureComposition demonstrates composing fixtures
func TestFixtureComposition(t *testing.T) {
	// Create a complete test setup
	ctx := testutil.NewTestContext(
		testutil.WithProjectRoot("/test/myapp"),
		testutil.WithProjectName("myapp"),
		testutil.WithMultiWorktreeRoot(),
	)

	cfg := testutil.NewTestConfig(
		testutil.WithDefaultProject("myapp"),
		testutil.WithTestDefaults(true, true),
	)

	// Verify the fixtures
	testutil.AssertNotNil(t, ctx, "context should be created")
	testutil.AssertEqual(t, "/test/myapp", ctx.ProjectRoot, "wrong project root")
	testutil.AssertEqual(t, "myapp", ctx.ProjectName, "wrong project name")
	testutil.AssertTrue(t, ctx.IsRoot, "should be root")

	testutil.AssertNotNil(t, cfg, "config should be created")
	testutil.AssertEqual(t, "myapp", cfg.DefaultProject, "wrong default project")
	testutil.AssertTrue(t, cfg.Defaults.Test.Parallel, "parallel should be enabled")
	testutil.AssertTrue(t, cfg.Defaults.Test.Coverage, "coverage should be enabled")
}

// TestTempDir demonstrates using temporary directories
func TestTempDir(t *testing.T) {
	dir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// dir is a real temporary directory
	testutil.AssertNotEmpty(t, dir, "should have directory path")

	// cleanup() removes the directory when test completes
}

// TestTableDriven demonstrates table-driven tests with testutil
func TestTableDriven(t *testing.T) {
	tests := []struct {
		name     string
		mode     context.DevelopmentMode
		expected bool
	}{
		{
			name:     "multi-worktree can use project commands",
			mode:     context.ModeMultiWorktree,
			expected: true,
		},
		{
			name:     "single-repo cannot use project commands",
			mode:     context.ModeSingleRepo,
			expected: false,
		},
		{
			name:     "standalone cannot use project commands",
			mode:     context.ModeStandalone,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testutil.NewTestContext(
				testutil.WithDevelopmentMode(tt.mode),
			)

			result := ctx.CanUseProjectCommands()
			testutil.AssertEqual(t, tt.expected, result, "wrong result")
		})
	}
}

// TestBufferUsage demonstrates using NewTestWriter for output capture
func TestBufferUsage(t *testing.T) {
	buf := testutil.NewTestWriter()

	// Write to buffer
	buf.WriteString("Processing...")
	buf.WriteString("Done!")

	// Verify content
	output := buf.String()
	testutil.AssertContains(t, output, "Processing...", "should have processing message")
	testutil.AssertContains(t, output, "Done!", "should have done message")
}

// TestContextValidation demonstrates validating context state
func TestContextValidation(t *testing.T) {
	t.Run("valid context", func(t *testing.T) {
		ctx := testutil.NewTestContext(
			testutil.WithProjectRoot("/valid/project"),
		)

		testutil.AssertTrue(t, ctx.IsValid(), "should be valid")
		testutil.AssertNil(t, ctx.Error, "should have no error")
	})

	t.Run("invalid context with error", func(t *testing.T) {
		ctx := testutil.NewTestContext(
			testutil.WithContextError(context.ErrProjectRootNotFound),
		)

		testutil.AssertFalse(t, ctx.IsValid(), "should be invalid")
		testutil.AssertNotNil(t, ctx.Error, "should have error")
		testutil.AssertErrorContains(t, ctx.Error, "project root")
	})
}

// TestMultiWorktreeScenarios demonstrates testing multi-worktree scenarios
func TestMultiWorktreeScenarios(t *testing.T) {
	scenarios := []struct {
		name       string
		setup      func() *context.ProjectContext
		isRoot     bool
		isMainRepo bool
		isWorktree bool
	}{
		{
			name: "root location",
			setup: func() *context.ProjectContext {
				return testutil.NewTestContext(testutil.WithMultiWorktreeRoot())
			},
			isRoot:     true,
			isMainRepo: false,
			isWorktree: false,
		},
		{
			name: "main repo location",
			setup: func() *context.ProjectContext {
				return testutil.NewTestContext(testutil.WithMultiWorktreeMainRepo())
			},
			isRoot:     false,
			isMainRepo: true,
			isWorktree: false,
		},
		{
			name: "worktree location",
			setup: func() *context.ProjectContext {
				return testutil.NewTestContext(testutil.WithWorktree("feature-x"))
			},
			isRoot:     false,
			isMainRepo: false,
			isWorktree: true,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			ctx := scenario.setup()

			testutil.AssertEqual(t, scenario.isRoot, ctx.IsRoot, "IsRoot mismatch")
			testutil.AssertEqual(t, scenario.isMainRepo, ctx.IsMainRepo, "IsMainRepo mismatch")
			testutil.AssertEqual(t, scenario.isWorktree, ctx.IsWorktree, "IsWorktree mismatch")
			testutil.AssertEqual(t, context.ModeMultiWorktree, ctx.DevelopmentMode, "should be multi-worktree mode")
		})
	}
}

// TestConfigMerging demonstrates testing configuration merging
func TestConfigMerging(t *testing.T) {
	cfg := testutil.NewTestConfig(
		testutil.WithDefaultProject("project-a"),
		testutil.WithProject("project-a", "/path/a", "single-repo"),
		testutil.WithProject("project-b", "/path/b", "multi-worktree"),
		testutil.WithTestDefaults(true, true),
	)

	// Verify projects
	testutil.AssertLen(t, cfg.Projects, 3, "should have 3 projects") // includes default test-project
	testutil.AssertEqual(t, "project-a", cfg.DefaultProject, "wrong default project")

	// Verify project-a config
	projectA, exists := cfg.Projects["project-a"]
	testutil.AssertTrue(t, exists, "project-a should exist")
	testutil.AssertEqual(t, "/path/a", projectA.Path, "wrong path")
	testutil.AssertEqual(t, "single-repo", projectA.Mode, "wrong mode")

	// Verify defaults
	testutil.AssertTrue(t, cfg.Defaults.Test.Parallel, "parallel should be enabled")
	testutil.AssertTrue(t, cfg.Defaults.Test.Coverage, "coverage should be enabled")
}
