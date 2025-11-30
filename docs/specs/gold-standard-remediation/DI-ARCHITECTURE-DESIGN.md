# Dependency Injection Architecture Design

**Status:** Draft
**Created:** 2025-11-26
**Author:** Claude (Gold Standard Remediation - Task 1.1)
**Related:** Phase 1 - Core Architecture Refactoring

## Executive Summary

This document specifies the design of the Dependency Injection (DI) system that will replace the current `Application` God Object pattern. The new architecture uses **uber-fx** for compile-time dependency injection, lifecycle management, and explicit dependency graphs.

## Current State Analysis

### Problems with Current Approach

The current `pkg/app/application.go` implements a **Service Locator anti-pattern**:

```go
type Application struct {
    OutputManager  *output.Manager
    ProjectContext *context.ProjectContext
    Config         *config.Config
    ShellExecutor  *shell.Executor
    ConfigLoader   *config.Loader
    Writer         io.Writer
}
```

**Critical Issues:**

1. **No Lifecycle Management**
   - Dependencies created in arbitrary order
   - No graceful shutdown
   - Resource leaks possible

2. **Hidden Dependencies**
   - Circular dependencies possible (not detected at compile time)
   - Hard to understand what depends on what
   - Makes refactoring dangerous

3. **Testing Complexity**
   - Must bootstrap entire Application for unit tests
   - Hard to mock individual dependencies
   - Leads to integration tests instead of unit tests

4. **No Compile-Time Safety**
   - Dependencies can be nil
   - Lazy initialization with `GetConfigLoader()` hides issues
   - Runtime panics instead of compile errors

### Dependency Graph (Current)

```
main.go
  ├─> config.Load() -> *config.Config
  ├─> context.DetectWithExtensions(plugins) -> *context.ProjectContext
  ├─> app.NewApplication(opts) -> *app.Application
  │     ├─> output.NewManager() -> *output.Manager
  │     ├─> shell.NewExecutor() -> *shell.Executor
  │     └─> (lazy) config.NewLoader() -> *config.Loader
  ├─> plugin.LoadAll(cmd) -> registry operations
  └─> cobra.Command.Execute()
```

**Problems:**
- Dependencies created imperatively in main()
- No explicit dependency declaration
- Order matters (config before context, etc.)
- Easy to break by reordering

## Proposed Architecture

### Technology Choice: uber-fx

**Rationale:**
- ✅ Type-safe dependency injection
- ✅ Lifecycle management (startup/shutdown hooks)
- ✅ Explicit dependency graphs (compile-time validation)
- ✅ Testing-friendly (can override any dependency)
- ✅ Excellent error messages
- ✅ Production-proven (used by Uber, many others)

**Alternatives Considered:**
- ❌ **google/wire**: Requires codegen, less flexible
- ❌ **Manual DI**: Too much boilerplate, error-prone
- ❌ **Service Locator**: Current anti-pattern

### Container Design

#### Package Structure

```
pkg/
├── container/
│   ├── container.go        # Container type and New()
│   ├── providers.go        # Constructor providers
│   ├── lifecycle.go        # Lifecycle hooks
│   ├── options.go          # Container options
│   └── container_test.go   # Container tests
```

#### Core Container Interface

```go
// pkg/container/container.go
package container

import (
    "context"
    "go.uber.org/fx"
)

// Container wraps uber-fx and provides lifecycle management
type Container struct {
    app *fx.App
}

// New creates a new dependency injection container
func New(opts ...fx.Option) (*Container, error) {
    // Merge default options with user options
    allOpts := append(
        []fx.Option{
            // Core providers
            fx.Provide(
                // Logging (no dependencies)
                provideLogger,

                // Config (depends on logger)
                provideConfigLoader,
                provideConfig,

                // Context (depends on config, logger)
                provideContextDetector,
                provideProjectContext,

                // Output (depends on logger)
                provideOutputManager,

                // Shell (depends on logger)
                provideShellExecutor,

                // Plugin registry (depends on logger)
                providePluginRegistry,

                // CLI builder (depends on all above)
                provideCLIBuilder,

                // Root command (depends on CLI builder)
                provideRootCommand,
            ),

            // Lifecycle hooks
            fx.Invoke(registerLifecycleHooks),
        },
        opts...,
    )

    app := fx.New(allOpts...)
    if app.Err() != nil {
        return nil, app.Err()
    }

    return &Container{app: app}, nil
}

// Start starts the container and all managed components
func (c *Container) Start(ctx context.Context) error {
    return c.app.Start(ctx)
}

// Stop gracefully shuts down the container
func (c *Container) Stop(ctx context.Context) error {
    return c.app.Stop(ctx)
}

// Run executes the application with proper lifecycle management
func (c *Container) Run(ctx context.Context, fn func() error) error {
    // Start all components
    if err := c.Start(ctx); err != nil {
        return fmt.Errorf("failed to start container: %w", err)
    }

    // Ensure cleanup on exit
    defer func() {
        stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        _ = c.Stop(stopCtx)
    }()

    // Run the application function
    return fn()
}
```

#### Provider Functions

```go
// pkg/container/providers.go
package container

import (
    "io"
    "os"

    "github.com/ivannovak/glide/v2/internal/config"
    "github.com/ivannovak/glide/v2/internal/context"
    "github.com/ivannovak/glide/v2/internal/shell"
    "github.com/ivannovak/glide/v2/pkg/logging"
    "github.com/ivannovak/glide/v2/pkg/output"
    "github.com/ivannovak/glide/v2/pkg/plugin"
)

// ProviderParams groups common dependencies
type ProviderParams struct {
    fx.In

    Logger *logging.Logger
}

// Logger provider
func provideLogger() *logging.Logger {
    return logging.New(logging.FromEnv())
}

// Writer provider (can be overridden in tests)
func provideWriter() io.Writer {
    return os.Stdout
}

// Config loader provider
func provideConfigLoader(logger *logging.Logger) *config.Loader {
    return config.NewLoader()
}

// Config provider
func provideConfig(loader *config.Loader, logger *logging.Logger) (*config.Config, error) {
    cfg, err := loader.Load()
    if err != nil && !os.IsNotExist(err) {
        return nil, fmt.Errorf("failed to load config: %w", err)
    }
    return cfg, nil
}

// Context detector provider
func provideContextDetector(logger *logging.Logger) *context.Detector {
    return context.NewDetector()
}

// ProjectContext provider
type ProjectContextParams struct {
    fx.In

    Detector *context.Detector
    Plugins  []*plugin.Plugin `optional:"true"` // Get all registered plugins
    Logger   *logging.Logger
}

func provideProjectContext(params ProjectContextParams) *context.ProjectContext {
    // Convert plugins to extension providers
    extensionProviders := make([]interface{}, len(params.Plugins))
    for i, p := range params.Plugins {
        extensionProviders[i] = p
    }

    return params.Detector.DetectWithExtensions(extensionProviders)
}

// OutputManager provider
type OutputManagerParams struct {
    fx.In

    Writer io.Writer
    Logger *logging.Logger
}

func provideOutputManager(params OutputManagerParams) *output.Manager {
    return output.NewManager(
        output.FormatTable, // Default format, can be overridden
        false,              // quiet
        false,              // noColor
        params.Writer,
    )
}

// ShellExecutor provider
func provideShellExecutor(logger *logging.Logger) *shell.Executor {
    return shell.NewExecutor(shell.Options{})
}

// PluginRegistry provider
func providePluginRegistry(logger *logging.Logger) *plugin.Registry {
    return plugin.NewRegistry()
}
```

#### Lifecycle Management

```go
// pkg/container/lifecycle.go
package container

import (
    "context"

    "go.uber.org/fx"
)

// LifecycleParams groups all components that need lifecycle management
type LifecycleParams struct {
    fx.In

    Lifecycle fx.Lifecycle
    Logger    *logging.Logger
    // Add other components that need lifecycle hooks
}

// registerLifecycleHooks registers startup and shutdown hooks
func registerLifecycleHooks(params LifecycleParams) {
    params.Lifecycle.Append(fx.Hook{
        OnStart: func(ctx context.Context) error {
            params.Logger.Info("Starting glide application")
            return nil
        },
        OnStop: func(ctx context.Context) error {
            params.Logger.Info("Shutting down glide application")
            return nil
        },
    })
}
```

#### Testing Support

```go
// pkg/container/options.go
package container

import (
    "io"

    "go.uber.org/fx"
)

// Option is a functional option for configuring the container
type Option = fx.Option

// WithLogger overrides the logger for testing
func WithLogger(logger *logging.Logger) Option {
    return fx.Provide(func() *logging.Logger {
        return logger
    })
}

// WithWriter overrides the output writer for testing
func WithWriter(w io.Writer) Option {
    return fx.Provide(func() io.Writer {
        return w
    })
}

// WithConfig overrides the config for testing
func WithConfig(cfg *config.Config) Option {
    return fx.Provide(func() (*config.Config, error) {
        return cfg, nil
    })
}

// WithProjectContext overrides the project context for testing
func WithProjectContext(ctx *context.ProjectContext) Option {
    return fx.Provide(func() *context.ProjectContext {
        return ctx
    })
}

// WithoutLifecycle disables lifecycle hooks for faster tests
func WithoutLifecycle() Option {
    return fx.Options(
        fx.Provide(fx.Annotated{
            Target: func() fx.Lifecycle {
                return fx.NewLifecycle(fx.NopLogger)
            },
        }),
    )
}
```

### Migration Strategy

#### Phase 1: Coexistence (Backward Compatible)

Keep the old `Application` but implement it using the new container:

```go
// pkg/app/application.go (updated)
package app

import (
    "github.com/ivannovak/glide/v2/pkg/container"
)

// Deprecated: Use pkg/container.Container instead. Will be removed in v3.0.0.
type Application struct {
    container *container.Container

    // Keep fields for backward compatibility
    OutputManager  *output.Manager
    ProjectContext *context.ProjectContext
    Config         *config.Config
    ShellExecutor  *shell.Executor
    ConfigLoader   *config.Loader
    Writer         io.Writer
}

// Deprecated: Use container.New() instead
func NewApplication(opts ...Option) *Application {
    // Convert old options to fx.Option
    var fxOpts []fx.Option

    // ... conversion logic ...

    c, err := container.New(fxOpts...)
    if err != nil {
        panic(err) // Maintain old API contract (no error return)
    }

    app := &Application{
        container: c,
    }

    // Extract dependencies from container for backward compatibility
    // ... populate fields ...

    return app
}
```

#### Phase 2: Migration (Gradual)

Update `cmd/glide/main.go` to use the new container:

```go
// cmd/glide/main.go (new version)
package main

import (
    "context"

    "github.com/ivannovak/glide/v2/pkg/container"
    "github.com/spf13/cobra"
)

func Execute() error {
    ctx := context.Background()

    // Create DI container
    c, err := container.New()
    if err != nil {
        return fmt.Errorf("failed to create container: %w", err)
    }

    // Run application with lifecycle management
    return c.Run(ctx, func() error {
        // Extract root command from container
        var rootCmd *cobra.Command
        if err := c.Invoke(func(cmd *cobra.Command) {
            rootCmd = cmd
        }); err != nil {
            return err
        }

        // Execute command
        return rootCmd.Execute()
    })
}
```

#### Phase 3: Cleanup (Breaking)

Remove deprecated `Application` type in v3.0.0.

### Dependency Graph (New)

```
container.New()
  ├─> provideLogger() -> *logging.Logger
  │
  ├─> provideWriter() -> io.Writer
  │
  ├─> provideConfigLoader(logger) -> *config.Loader
  ├─> provideConfig(loader, logger) -> *config.Config
  │
  ├─> provideContextDetector(logger) -> *context.Detector
  ├─> provideProjectContext(detector, plugins, logger) -> *context.ProjectContext
  │
  ├─> provideOutputManager(writer, logger) -> *output.Manager
  │
  ├─> provideShellExecutor(logger) -> *shell.Executor
  │
  ├─> providePluginRegistry(logger) -> *plugin.Registry
  │
  ├─> provideCLIBuilder(all above) -> *cli.Builder
  │
  └─> provideRootCommand(builder) -> *cobra.Command
```

**Benefits:**
- ✅ Explicit dependency graph
- ✅ Compile-time validation
- ✅ Proper initialization order
- ✅ Lifecycle management
- ✅ Easy testing (override any provider)

## Implementation Plan

### Task 1.1 Subtasks (20 hours total)

#### Subtask 1.1.1: Create Container Package Structure (2h)
- [ ] Create `pkg/container/` directory
- [ ] Create `container.go` with Container type
- [ ] Create `providers.go` for provider functions
- [ ] Create `lifecycle.go` for lifecycle hooks
- [ ] Create `options.go` for testing options
- [ ] Add uber-fx dependency to go.mod

**Files to Create:**
- `pkg/container/container.go`
- `pkg/container/providers.go`
- `pkg/container/lifecycle.go`
- `pkg/container/options.go`
- `pkg/container/container_test.go`

**Validation:**
```bash
go mod tidy
go build ./pkg/container/...
```

#### Subtask 1.1.2: Implement Core Providers (4h)
- [ ] Implement `provideLogger()`
- [ ] Implement `provideWriter()`
- [ ] Implement `provideConfigLoader()`
- [ ] Implement `provideConfig()`
- [ ] Implement `provideContextDetector()`
- [ ] Implement `provideProjectContext()`
- [ ] Implement `provideOutputManager()`
- [ ] Implement `provideShellExecutor()`
- [ ] Add tests for each provider

**Files to Modify:**
- `pkg/container/providers.go`
- `pkg/container/container_test.go`

**Validation:**
```bash
go test ./pkg/container/... -v
```

#### Subtask 1.1.3: Implement Container Lifecycle (2h)
- [ ] Implement `Container.New()`
- [ ] Implement `Container.Start()`
- [ ] Implement `Container.Stop()`
- [ ] Implement `Container.Run()`
- [ ] Add lifecycle hooks registration
- [ ] Add tests for lifecycle management

**Files to Modify:**
- `pkg/container/container.go`
- `pkg/container/lifecycle.go`
- `pkg/container/container_test.go`

**Validation:**
```bash
go test ./pkg/container/... -v -run TestLifecycle
```

#### Subtask 1.1.4: Add Testing Support (2h)
- [ ] Implement `WithLogger()`
- [ ] Implement `WithWriter()`
- [ ] Implement `WithConfig()`
- [ ] Implement `WithProjectContext()`
- [ ] Implement `WithoutLifecycle()`
- [ ] Add integration tests using overrides

**Files to Modify:**
- `pkg/container/options.go`
- `pkg/container/container_test.go`

**Validation:**
```bash
go test ./pkg/container/... -v -run TestOptions
```

#### Subtask 1.1.5: Create Backward Compatibility Shim (4h)
- [ ] Update `pkg/app/application.go` to use container internally
- [ ] Convert old Options to fx.Option
- [ ] Extract dependencies from container for field access
- [ ] Add deprecation comments
- [ ] Update tests to ensure backward compatibility

**Files to Modify:**
- `pkg/app/application.go`
- `pkg/app/application_test.go`

**Validation:**
```bash
go test ./pkg/app/... -v
# All existing tests should pass without modification
```

#### Subtask 1.1.6: Create ADR Document (2h)
- [ ] Document DI architecture decision
- [ ] Explain uber-fx choice
- [ ] Document migration strategy
- [ ] Add examples and best practices

**Files to Create:**
- `docs/adr/ADR-013-dependency-injection.md`

#### Subtask 1.1.7: Update Implementation Checklist (2h)
- [ ] Fill in detailed subtasks for Task 1.2-1.5
- [ ] Add acceptance criteria
- [ ] Add validation steps
- [ ] Update effort estimates

**Files to Modify:**
- `docs/specs/gold-standard-remediation/implementation-checklist.md`

#### Subtask 1.1.8: Integration Testing (2h)
- [ ] Test container initialization
- [ ] Test dependency resolution
- [ ] Test lifecycle management
- [ ] Test backward compatibility
- [ ] Smoke test with existing CLI

**Validation:**
```bash
# Build with new container (via shim)
go build ./cmd/glide
./glide version
./glide help
./glide context

# Run full test suite
go test ./... -v

# Check coverage
go test -coverprofile=coverage.out ./pkg/container/...
go tool cover -func=coverage.out
# Should be >90%
```

### Acceptance Criteria

- [ ] Container package compiles without errors
- [ ] All providers implement proper dependency injection
- [ ] Lifecycle hooks work (startup/shutdown)
- [ ] Testing options allow easy mocking
- [ ] Backward compatibility shim works (all existing tests pass)
- [ ] ADR document complete and reviewed
- [ ] Implementation checklist updated
- [ ] Integration tests passing
- [ ] Coverage >90% on new code
- [ ] No regressions in existing functionality

### Migration Timeline

| Phase | Duration | Tasks |
|-------|----------|-------|
| Design (Task 1.1) | 20h | This document + implementation |
| Implementation (Task 1.2) | 24h | Update main.go to use container |
| Migration (Task 1.3) | 16h | Remove old Application usage |
| Cleanup (Task 1.4) | 8h | Remove deprecated code |
| **Total** | **68h** | **≈9 days** |

## Risk Assessment

### High Risks

1. **Breaking Changes**
   - **Mitigation:** Backward compatibility shim for 6 months
   - **Test:** Run full test suite at each step

2. **Plugin System Impact**
   - **Mitigation:** Plugins loaded via DI providers
   - **Test:** Smoke test all plugins

3. **Performance Regression**
   - **Mitigation:** Benchmark container initialization
   - **Test:** Compare startup time before/after

### Medium Risks

1. **Learning Curve**
   - **Mitigation:** Comprehensive ADR and examples
   - **Test:** Pair programming sessions

2. **Testing Complexity**
   - **Mitigation:** Testing helpers in container/options.go
   - **Test:** Migrate existing tests gradually

## Success Metrics

1. **Functional**
   - [ ] All existing tests pass
   - [ ] No behavioral changes
   - [ ] CLI works identically

2. **Quality**
   - [ ] >90% test coverage on new code
   - [ ] Zero linter warnings
   - [ ] All race conditions detected

3. **Maintainability**
   - [ ] Clear dependency graph
   - [ ] Easy to add new dependencies
   - [ ] Easy to test components

## References

- [uber-fx Documentation](https://uber-go.github.io/fx/)
- [Dependency Injection in Go](https://blog.drewolson.org/dependency-injection-in-go)
- [Gold Standard Remediation - Tech Spec](./tech.md)
- [ADR-002: Plugin System Design](../../adr/ADR-002-plugin-system-design.md)

## Appendix A: Example Usage

### Creating a Container (Production)

```go
package main

import (
    "context"
    "github.com/ivannovak/glide/v2/pkg/container"
)

func main() {
    ctx := context.Background()

    c, err := container.New()
    if err != nil {
        log.Fatal(err)
    }

    if err := c.Run(ctx, func() error {
        // Application logic here
        return nil
    }); err != nil {
        log.Fatal(err)
    }
}
```

### Testing with Container

```go
package mypackage_test

import (
    "bytes"
    "testing"

    "github.com/ivannovak/glide/v2/pkg/container"
    "github.com/ivannovak/glide/v2/tests/testutil"
)

func TestMyFeature(t *testing.T) {
    buf := bytes.NewBuffer(nil)
    cfg := testutil.NewTestConfig()

    c, err := container.New(
        container.WithWriter(buf),
        container.WithConfig(cfg),
        container.WithoutLifecycle(), // Faster tests
    )
    require.NoError(t, err)

    // Test with mocked dependencies
    // ...
}
```

## Appendix B: Provider Dependency Matrix

| Provider | Dependencies | Produces |
|----------|-------------|----------|
| provideLogger | - | *logging.Logger |
| provideWriter | - | io.Writer |
| provideConfigLoader | Logger | *config.Loader |
| provideConfig | Loader, Logger | *config.Config |
| provideContextDetector | Logger | *context.Detector |
| provideProjectContext | Detector, Plugins, Logger | *context.ProjectContext |
| provideOutputManager | Writer, Logger | *output.Manager |
| provideShellExecutor | Logger | *shell.Executor |
| providePluginRegistry | Logger | *plugin.Registry |

---

**Document Version:** 1.0
**Last Updated:** 2025-11-26
**Status:** Ready for Implementation
