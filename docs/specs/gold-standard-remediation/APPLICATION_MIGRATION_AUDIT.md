# Application God Object Migration Audit

**Date:** 2025-11-26
**Task:** 1.3.1 - Audit Application Usage Patterns
**Goal:** Document all usages of `pkg/app/Application` to plan migration to container DI

## Executive Summary

The Application God Object is used in **5 production files** and **5 test files**:

**Production Files:**
- `cmd/glide/main.go` - Entry point, creates Application and CLI
- `internal/cli/cli.go` - CLI struct stores Application reference
- `internal/cli/builder.go` - Builder struct stores Application reference
- `internal/cli/base.go` - BaseCommand struct stores Application reference
- `internal/cli/debug.go` - Debug helper functions accept Application parameter

**Test Files:**
- `internal/cli/cli_test.go` - 11 usages
- `internal/cli/base_test.go` - 11 usages
- `internal/cli/alias_integration_test.go` - 4 usages
- `tests/testutil/examples_test.go` - 1 comment reference

**Documentation Files:** (not migrated)
- `docs/adr/ADR-013-dependency-injection.md`
- `docs/specs/gold-standard-remediation/DI-ARCHITECTURE-DESIGN.md`
- `docs/specs/gold-standard-remediation/implementation-checklist.md`

---

## Detailed Usage Analysis

### 1. `cmd/glide/main.go` (Production)

**Lines:** 12 (import), 74-78 (creation), 138 (pass to CLI)

**Usage Pattern:** Application Creation and Bootstrap
```go
// Line 12
import "github.com/ivannovak/glide/v2/pkg/app"

// Lines 74-78
application := app.NewApplication(
    app.WithProjectContext(ctx),
    app.WithConfig(cfg),
    app.WithOutputFormat(output.FormatTable, false, false),
)

// Line 81
output.SetGlobalManager(application.OutputManager)

// Line 138
cli := cliPkg.New(application)
```

**Dependencies Used:**
- `application.OutputManager` - for global manager
- Passed to `cliPkg.New()`

**Migration Strategy:**
```go
// Replace with:
var (
    outputManager  *output.Manager
    projectContext *context.ProjectContext
    config         *config.Config
)

c, err := container.New(
    container.WithProjectContext(ctx),
    container.WithConfig(cfg),
    fx.Populate(&outputManager, &projectContext, &config),
)
if err != nil {
    return fmt.Errorf("failed to initialize dependencies: %w", err)
}

output.SetGlobalManager(outputManager)
cli := cliPkg.New(outputManager, projectContext, config)
```

**Impact:** CRITICAL - Entry point for entire application

---

### 2. `internal/cli/cli.go` (Production)

**Lines:** 11 (import), 19 (field), 24-28 (constructor), 38-276 (field access)

**Usage Pattern:** Service Locator Pattern
```go
// Line 11
import "github.com/ivannovak/glide/v2/pkg/app"

// Line 19
type CLI struct {
    app     *app.Application
    builder *Builder
}

// Lines 24-28
func New(application *app.Application) *CLI {
    return &CLI{
        app:     application,
        builder: NewBuilder(application),
    }
}
```

**Field Access Patterns:**
- `c.app.ProjectContext` - Lines 38, 48, 53, 59, 69, 137, 244, 276, 320, 398
- `c.app.Config` - Lines 38, 43, 48, 53, 59, 69
- `c.app.OutputManager` - Lines 254, 257, 266, 281, 310, 332, 338, 369, 412, 427, 447, 450, 462, 464, 474, 492

**Dependencies Needed:**
1. `outputManager *output.Manager` - Most frequently used (16 times)
2. `projectContext *context.ProjectContext` - Used 10 times
3. `config *config.Config` - Used 6 times

**Migration Strategy:**
```go
type CLI struct {
    outputManager  *output.Manager
    projectContext *context.ProjectContext
    config         *config.Config
    builder        *Builder
}

func New(
    outputManager *output.Manager,
    projectContext *context.ProjectContext,
    config *config.Config,
) *CLI {
    builder := NewBuilder(outputManager, projectContext, config)
    return &CLI{
        outputManager:  outputManager,
        projectContext: projectContext,
        config:         config,
        builder:        builder,
    }
}
```

Then replace all:
- `c.app.OutputManager` → `c.outputManager`
- `c.app.ProjectContext` → `c.projectContext`
- `c.app.Config` → `c.config`

**Impact:** HIGH - Core CLI structure, affects all commands

---

### 3. `internal/cli/builder.go` (Production)

**Lines:** 7 (import), 16 (field), 21-33 (constructor), 40+ (field access)

**Usage Pattern:** Service Locator Pattern (Builder)
```go
// Line 7
import "github.com/ivannovak/glide/v2/pkg/app"

// Line 16
type Builder struct {
    app      *app.Application
    registry *Registry
}

// Lines 21-33
func NewBuilder(application *app.Application) *Builder {
    builder := &Builder{
        app:      application,
        registry: NewRegistry(),
    }
    builder.registerCommands()
    return builder
}
```

**Field Access Patterns:**
- `b.app.ProjectContext` - Lines 40, 48, 58, 67, 75, 84, 92, 105
- `b.app.Config` - Lines 40, 43, 48, 58, 67, 75, 84, 92, 105
- `b.app` - Passed to functions at lines 148, 158, 168, 178

**Dependencies Needed:**
1. `projectContext *context.ProjectContext` - 8 direct + 4 passed = 12 usages
2. `config *config.Config` - 9 direct + 4 passed = 13 usages

**Migration Strategy:**
```go
type Builder struct {
    projectContext *context.ProjectContext
    config         *config.Config
    registry       *Registry
}

func NewBuilder(
    projectContext *context.ProjectContext,
    config *config.Config,
) *Builder {
    builder := &Builder{
        projectContext: projectContext,
        config:         config,
        registry:       NewRegistry(),
    }
    builder.registerCommands()
    return builder
}
```

Then replace all:
- `b.app.ProjectContext` → `b.projectContext`
- `b.app.Config` → `b.config`
- `b.app` passed to debug functions → pass individual dependencies

**Impact:** MEDIUM - Affects command registration

---

### 4. `internal/cli/base.go` (Production)

**Lines:** 4 (import), 12 (field), 16-20 (constructor), 22+ (methods)

**Usage Pattern:** Base Command Pattern
```go
// Line 4
import "github.com/ivannovak/glide/v2/pkg/app"

// Line 12
type BaseCommand struct {
    app *app.Application
}

// Lines 16-20
func NewBaseCommand(application *app.Application) BaseCommand {
    return BaseCommand{
        app: application,
    }
}

// Lines 22-40 - Getter methods
func (b *BaseCommand) OutputManager() *output.Manager {
    return b.app.OutputManager
}

func (b *BaseCommand) ProjectContext() *context.ProjectContext {
    return b.app.ProjectContext
}

func (b *BaseCommand) Config() *config.Config {
    return b.app.Config
}

func (b *BaseCommand) Application() *app.Application {
    return b.app
}
```

**Migration Strategy:**
```go
type BaseCommand struct {
    outputManager  *output.Manager
    projectContext *context.ProjectContext
    config         *config.Config
}

func NewBaseCommand(
    outputManager *output.Manager,
    projectContext *context.ProjectContext,
    config *config.Config,
) BaseCommand {
    return BaseCommand{
        outputManager:  outputManager,
        projectContext: projectContext,
        config:         config,
    }
}

// Getter methods stay the same but return direct fields
func (b *BaseCommand) OutputManager() *output.Manager {
    return b.outputManager
}

func (b *BaseCommand) ProjectContext() *context.ProjectContext {
    return b.projectContext
}

func (b *BaseCommand) Config() *config.Config {
    return b.config
}

// Remove Application() method entirely
```

**Impact:** MEDIUM - Affects all commands that use BaseCommand

---

### 5. `internal/cli/debug.go` (Production)

**Lines:** 6 (import), 18, 89, 138, 178 (function parameters)

**Usage Pattern:** Helper Functions with Application Parameter
```go
// Line 6
import "github.com/ivannovak/glide/v2/pkg/app"

// Line 18
func showContext(cmd *cobra.Command, app *app.Application) error {
    ctx := app.ProjectContext
    // ... uses ctx
}

// Line 89
func testShell(cmd *cobra.Command, args []string, app *app.Application) error {
    // Uses:
    // - app.OutputManager (lines 101, 104, 113, 121, 125, 158)
    // - app.ProjectContext (line 129)
}

// Line 138
func testDockerResolution(cmd *cobra.Command, args []string, app *app.Application) error {
    ctx := app.ProjectContext
    // Uses:
    // - app.OutputManager (lines 149, 154, 179)
}

// Line 178
func testContainerManagement(cmd *cobra.Command, args []string, app *app.Application) error {
    ctx := app.ProjectContext
    // Uses:
    // - app.OutputManager (lines 194, 209, 215, 222, 225, 233, 240, 242, 246)
}
```

**Migration Strategy:**
```go
// Replace all functions to accept individual dependencies:

func showContext(
    cmd *cobra.Command,
    projectContext *context.ProjectContext,
) error {
    // Use projectContext directly
}

func testShell(
    cmd *cobra.Command,
    args []string,
    outputManager *output.Manager,
    projectContext *context.ProjectContext,
) error {
    // Use dependencies directly
}

func testDockerResolution(
    cmd *cobra.Command,
    args []string,
    outputManager *output.Manager,
    projectContext *context.ProjectContext,
) error {
    // Use dependencies directly
}

func testContainerManagement(
    cmd *cobra.Command,
    args []string,
    outputManager *output.Manager,
    projectContext *context.ProjectContext,
) error {
    // Use dependencies directly
}
```

**Impact:** LOW - Only affects debug commands

---

## Test File Analysis

### 6. `internal/cli/cli_test.go` (Test)

**Usage Count:** 11 instances of `app.NewApplication()`

**Lines:** 18, 101, 123, 149, 193, 215, 234, 256, 271, 294-295, 329

**Pattern:** Test fixture creation
```go
application := app.NewApplication(
    app.WithOutputFormat(output.FormatPlain, true, true),
)
cli := New(application)
```

**Migration Strategy:**
```go
// Option 1: Direct dependency injection
outputMgr := output.NewManager(output.FormatPlain, true, true, os.Stdout)
ctx := &context.ProjectContext{WorkingDir: t.TempDir()}
cfg := &config.Config{}
cli := New(outputMgr, ctx, cfg)

// Option 2: Use container for integration tests
var outputManager *output.Manager
var projectContext *context.ProjectContext
var config *config.Config

c, err := container.New(
    container.WithWriter(io.Discard),
    fx.Populate(&outputManager, &projectContext, &config),
)
require.NoError(t, err)
defer c.Stop(context.Background())

cli := New(outputManager, projectContext, config)
```

**Impact:** HIGH - All CLI tests need updating

---

### 7. `internal/cli/base_test.go` (Test)

**Usage Count:** 11 instances of `app.NewApplication()`

**Lines:** 16, 28, 42, 59, 71, 92, 104, 115, 144, 179, 234

**Pattern:** BaseCommand testing
```go
application := app.NewApplication()
cmd := NewBaseCommand(application)
```

**Migration Strategy:**
```go
outputMgr := output.NewManager(output.FormatTable, false, false, os.Stdout)
ctx := &context.ProjectContext{WorkingDir: t.TempDir()}
cfg := &config.Config{}

cmd := NewBaseCommand(outputMgr, ctx, cfg)
```

**Impact:** MEDIUM - BaseCommand tests

---

### 8. `internal/cli/alias_integration_test.go` (Test)

**Usage Count:** 4 instances of `app.NewApplication()`

**Lines:** 14, 42, 86, 127

**Pattern:** Integration testing with minimal setup
```go
application := app.NewApplication()
cli := New(application)
```

**Migration Strategy:**
```go
// Minimal setup for integration tests
outputMgr := output.NewManager(output.FormatTable, false, false, os.Stdout)
ctx := &context.ProjectContext{WorkingDir: t.TempDir()}
cfg := &config.Config{}

cli := New(outputMgr, ctx, cfg)
```

**Impact:** MEDIUM - Integration tests

---

### 9. `tests/testutil/examples_test.go` (Test/Documentation)

**Usage Count:** 1 comment reference

**Line:** 89 (comment only)

**Content:**
```go
// Tests that need Application can create them directly using app.NewApplication()
```

**Migration Strategy:**
Update comment to:
```go
// Tests that need dependencies can create them directly or use container.New()
// For simple tests, create dependencies directly:
//   outputMgr := output.NewManager(...)
// For complex tests, use container with fx.Populate:
//   var deps DependencyStruct
//   container.New(fx.Populate(&deps))
```

**Impact:** TRIVIAL - Documentation only

---

## Migration Dependency Graph

```
┌─────────────────────┐
│ cmd/glide/main.go   │  (1) Start here - update to use container
│ Creates Application │
└──────────┬──────────┘
           │ passes app to
           ↓
    ┌──────────────┐
    │  cli.New()   │  (2) Change signature to accept dependencies
    └──────┬───────┘
           │ stores app.Application
           ↓
    ┌──────────────────┐
    │  internal/cli/   │  (3) Refactor package internals
    │  - CLI struct    │      - Remove app field
    │  - Builder       │      - Add individual dependency fields
    │  - BaseCommand   │      - Update constructors
    │  - debug helpers │      - Update all references
    └──────┬───────────┘
           │
           ↓
    ┌──────────────────┐
    │  Test Files      │  (4) Update after production code works
    │  - cli_test      │
    │  - base_test     │
    │  - alias_test    │
    └──────────────────┘
```

**Migration Order:**
1. **Phase 1:** Refactor `internal/cli` package to support dependency injection (Subtask 1.3.2)
   - Update struct definitions
   - Update constructors to accept dependencies
   - Keep Application as optional for backward compatibility during transition

2. **Phase 2:** Update `cmd/glide/main.go` to use container (Subtask 1.3.3)
   - Replace `app.NewApplication()` with `container.New()`
   - Extract dependencies with `fx.Populate`
   - Pass dependencies to CLI constructor

3. **Phase 3:** Update test files (Subtask 1.3.4)
   - Replace Application creation with dependency injection
   - Use mocks from `tests/testutil/mocks` where appropriate

4. **Phase 4:** Mark Application for deprecation (Subtask 1.3.5)
   - Add removal timeline (v3.0.0)
   - Update documentation

---

## Dependencies Per Component

### CLI Struct Needs:
1. `*output.Manager` - 16 usages
2. `*context.ProjectContext` - 10 usages
3. `*config.Config` - 6 usages

### Builder Struct Needs:
1. `*context.ProjectContext` - 12 usages
2. `*config.Config` - 13 usages

### BaseCommand Struct Needs:
1. `*output.Manager` - via getter
2. `*context.ProjectContext` - via getter
3. `*config.Config` - via getter

### Debug Helper Functions Need:
- `showContext`: ProjectContext only
- `testShell`: OutputManager, ProjectContext
- `testDockerResolution`: OutputManager, ProjectContext
- `testContainerManagement`: OutputManager, ProjectContext

---

## Potential Issues & Risks

### 1. Circular Import Risk
**Risk:** LOW
**Reason:** `internal/cli` already imports from `internal/` and `pkg/`, not vice versa.
**Mitigation:** No changes needed - dependency graph is clean.

### 2. Test Complexity
**Risk:** MEDIUM
**Reason:** 26 test instances need updating.
**Mitigation:** Create helper function for common test setup patterns.

### 3. Backward Compatibility
**Risk:** LOW (internal package)
**Reason:** `internal/cli` is not exported; only `cmd/glide/main.go` uses it.
**Mitigation:** No external consumers to break.

### 4. Builder Pattern Complexity
**Risk:** LOW
**Reason:** Builder currently passes Application to debug functions.
**Mitigation:** Update debug functions to accept individual dependencies.

---

## Test Migration Helpers

Create helper function in test files:

```go
// testutil helper for CLI tests
type testDependencies struct {
    OutputManager  *output.Manager
    ProjectContext *context.ProjectContext
    Config         *config.Config
}

func newTestDependencies(t *testing.T) testDependencies {
    return testDependencies{
        OutputManager:  output.NewManager(output.FormatPlain, true, true, io.Discard),
        ProjectContext: &context.ProjectContext{WorkingDir: t.TempDir()},
        Config:         &config.Config{},
    }
}

// Usage:
deps := newTestDependencies(t)
cli := New(deps.OutputManager, deps.ProjectContext, deps.Config)
```

---

## Success Metrics

- [ ] Zero `import "github.com/ivannovak/glide/v2/pkg/app"` in `cmd/` (except tests)
- [ ] Zero `import "github.com/ivannovak/glide/v2/pkg/app"` in `internal/cli/` production code
- [ ] All 26 test instances migrated
- [ ] All tests passing
- [ ] No regressions in functionality
- [ ] Build succeeds: `go build cmd/glide/main.go`
- [ ] Smoke tests pass:
  - `./glide version`
  - `./glide help`
  - `./glide context`

---

## Estimated Effort Breakdown

Based on usage analysis:

| Subtask | Effort | Reason |
|---------|--------|--------|
| 1.3.1 Audit | 2h | ✅ COMPLETE |
| 1.3.2 Refactor CLI | 4h | 5 files, ~100 LOC changes |
| 1.3.3 Update main.go | 3h | 1 file, critical path, testing needed |
| 1.3.4 Update tests | 4h | 26 instances, create helpers |
| 1.3.5 Deprecation | 1h | Update comments, ADR |
| 1.3.6 Integration test | 2h | Full test suite, smoke tests |
| **Total** | **16h** | |

---

## Next Steps

1. ✅ Complete this audit document
2. ⬜ Begin Subtask 1.3.2: Refactor CLI package
3. ⬜ Continue with remaining subtasks in order

---

**Audit Completed:** 2025-11-26
**Auditor:** Claude Code (AI Assistant)
**Status:** READY FOR MIGRATION
