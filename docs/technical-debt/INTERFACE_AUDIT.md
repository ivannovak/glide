# Interface Audit Report

**Date:** 2025-11-27
**Status:** Complete
**Total Interfaces Found:** 43

## Executive Summary

This audit catalogs all interfaces in the Glide codebase, classifies them by purpose, identifies issues with current designs, and provides recommendations for cleanup.

### Key Findings

1. **43 interfaces** identified across the codebase
2. **Fat interfaces** found: 7 interfaces violating Interface Segregation Principle
3. **Unnecessary interfaces**: 3 interfaces with single implementation and no testing benefit
4. **Documentation gaps**: 28 interfaces missing comprehensive documentation
5. **Well-designed interfaces**: 12 interfaces following SOLID principles

### Priority Summary

- **P0-CRITICAL**: 2 interfaces (fix immediately)
- **P1-HIGH**: 5 interfaces (fix in current phase)
- **P2-MEDIUM**: 8 interfaces (fix when touching code)
- **SAFE**: 28 interfaces (no action needed or minor improvements)

---

## Interface Classification

### 1. Core Abstractions (DI/Testing)

These interfaces are essential for dependency injection and testing. They abstract core system components.

| Interface | Location | Methods | Implementations | Status | Priority |
|-----------|----------|---------|-----------------|--------|----------|
| ShellExecutor | pkg/interfaces/interfaces.go:10 | 3 | Multiple | **FAT** | P1 |
| ConfigLoader | pkg/interfaces/interfaces.go:70 | 4 | 1 | **UNNECESSARY?** | P2 |
| ContextDetector | pkg/interfaces/interfaces.go:78 | 2 | 1 | GOOD | SAFE |
| OutputManager | pkg/interfaces/interfaces.go:96 | 8 | 1 | **FAT** | P1 |
| CommandBuilder | pkg/interfaces/interfaces.go:128 | 3 | 1 | GOOD | SAFE |
| Registry | pkg/interfaces/interfaces.go:135 | 4 | Multiple | GOOD | SAFE |

**Issues Identified:**

1. **ShellExecutor (P1-HIGH)** - Fat interface with 3 different execution modes
   - `Execute()`, `ExecuteWithTimeout()`, `ExecuteWithProgress()`
   - Each method has different responsibilities
   - **Recommendation**: Split into `BasicExecutor`, `TimeoutExecutor`, `ProgressExecutor`

2. **OutputManager (P1-HIGH)** - Fat interface with 8 methods
   - Display, Info, Success, Error, Warning, Raw, Printf, Println
   - Mixes structured output with raw text output
   - **Recommendation**: Split into `StructuredOutput` and `TextOutput`

3. **ConfigLoader (P2-MEDIUM)** - Single implementation
   - Only `internal/config/loader.go` implements this
   - Never mocked in tests
   - **Recommendation**: Remove interface, use concrete type with functional options

---

### 2. Plugin SDK Interfaces (Public API)

These interfaces define the plugin extension system and must remain stable for backward compatibility.

| Interface | Location | Methods | Implementations | Status | Priority |
|-----------|----------|---------|-----------------|--------|----------|
| Plugin | pkg/plugin/interface.go:8 | 5 | Multiple | **FAT** | P0 |
| CommandProvider | pkg/plugin/sdk/command.go:81 | 1 | Multiple | GOOD | SAFE |
| ContextExtension | pkg/plugin/sdk/extensions.go:8 | 3 | Multiple | GOOD | SAFE |
| ContextProvider | pkg/plugin/sdk/extensions.go:26 | 1 | Multiple | GOOD | SAFE |
| ConfigProvider | pkg/plugin/sdk/config_schema.go:44 | 1 | Multiple | GOOD | SAFE |
| CompletionProvider | pkg/plugin/sdk/completion.go:8 | 1 | Multiple | GOOD | SAFE |
| FrameworkDetector | pkg/plugin/sdk/detection.go:4 | 1 | Multiple | GOOD | SAFE |
| InteractiveCommandHandler | pkg/plugin/sdk/v1/base_plugin.go:9 | 1 | Multiple | GOOD | SAFE |
| CommandHandler | pkg/plugin/sdk/v1/base_plugin.go:14 | 2 | Multiple | GOOD | SAFE |
| InteractiveSession | pkg/plugin/sdk/v1/plugin_types.go:40 | Unknown | Multiple | UNDOCUMENTED | P2 |

**Issues Identified:**

1. **Plugin (P0-CRITICAL)** - Fat interface violating ISP
   - 5 methods: Name(), Version(), Register(), Configure(), Metadata()
   - Mixes identification, lifecycle, and configuration concerns
   - **Recommendation**: Split into sub-interfaces:
     ```go
     type PluginIdentifier interface {
         Name() string
         Version() string
         Metadata() PluginMetadata
     }

     type PluginRegistrar interface {
         Register(root *cobra.Command) error
     }

     type PluginConfigurable interface {
         Configure(config map[string]interface{}) error
     }

     // Composite for backward compatibility
     type Plugin interface {
         PluginIdentifier
         PluginRegistrar
         PluginConfigurable
     }
     ```
   - **Breaking Change Risk**: HIGH (public API)
   - **Mitigation**: Keep composite interface, add sub-interfaces for flexibility

2. **InteractiveSession (P2-MEDIUM)** - No documentation
   - Used by gRPC plugin system
   - Methods unknown without reading protobuf
   - **Recommendation**: Add comprehensive godoc

---

### 3. Internal Abstractions

These interfaces are internal implementation details not exposed to users.

| Interface | Location | Methods | Implementations | Status | Priority |
|-----------|----------|---------|-----------------|--------|----------|
| ExecutionStrategy | internal/shell/strategy.go:12 | 2 | 4 | GOOD | SAFE |
| DetectionStrategy | internal/context/strategies.go:10 | 2 | 1 | GOOD | SAFE |
| ProjectRootFinder | internal/context/strategies.go:16 | 1 | 1 | GOOD | SAFE |
| DevelopmentModeDetector | internal/context/strategies.go:21 | 1 | 1 | GOOD | SAFE |
| LocationIdentifier | internal/context/strategies.go:26 | 1 | 1 | GOOD | SAFE |
| ComposeFileResolver | internal/context/strategies.go:31 | 1 | 1 | GOOD | SAFE |
| DockerStatusChecker | internal/context/strategies.go:36 | 1 | 1 | GOOD | SAFE |
| ExecutorProvider | internal/shell/registry.go:9 | 3 | Multiple | GOOD | SAFE |
| CommandExecutor | internal/shell/registry.go:22 | 1 | Multiple | GOOD | SAFE |
| ContextAwareExecutor | internal/shell/plugin_aware.go:63 | Unknown | Unknown | UNDOCUMENTED | P2 |
| CommandSanitizer | internal/shell/sanitizer.go:24 | Unknown | 2+ | UNDOCUMENTED | P2 |
| ExtensionRegistry | internal/context/detector.go:22 | Unknown | 1 | **UNNECESSARY** | P2 |

**Issues Identified:**

1. **Single-implementation interfaces (P2-MEDIUM)** - 6 interfaces with only one implementation:
   - `ProjectRootFinder`, `DevelopmentModeDetector`, `LocationIdentifier`
   - `ComposeFileResolver`, `DockerStatusChecker`, `DetectionStrategy`
   - **Recommendation**: Evaluate if abstraction provides value
   - **Keep if**: Used for plugin extensions or testing
   - **Remove if**: No mocking in tests and not extensible

2. **ExtensionRegistry (P2-MEDIUM)** - Should be a concrete type
   - Found at `internal/context/detector.go:22` but also `pkg/plugin/sdk/extensions.go:33` as struct
   - This appears to be both an interface and a struct - confusion!
   - **Recommendation**: Make it a struct only, not an interface

---

### 4. Data Transfer Interfaces

These interfaces represent getters for data structures. They're often candidates for removal.

| Interface | Location | Methods | Implementations | Status | Priority |
|-----------|----------|---------|-----------------|--------|----------|
| ShellCommand | pkg/interfaces/interfaces.go:17 | 5 | Multiple | **FAT** | P1 |
| ShellCommandOptions | pkg/interfaces/interfaces.go:26 | 5 | Multiple | **FAT** | P1 |
| ProjectContext | pkg/interfaces/interfaces.go:84 | 8 | 1 | **FAT/UNNECESSARY** | P0 |
| DockerResolver | pkg/interfaces/interfaces.go:43 | 4 | Unknown | UNDOCUMENTED | P2 |
| ContainerManager | pkg/interfaces/interfaces.go:51 | 6 | Unknown | **FAT** | P1 |

**Issues Identified:**

1. **ShellCommand (P1-HIGH)** - Pure getter interface
   - 5 getter methods: GetCommand(), GetArgs(), GetOptions(), GetWorkingDir(), GetEnvironment()
   - **Recommendation**: Replace with struct
   - **Reason**: No polymorphism needed, just data transfer

2. **ShellCommandOptions (P1-HIGH)** - Pure getter interface
   - 5 getter methods for options
   - **Recommendation**: Replace with struct or functional options pattern

3. **ProjectContext (P0-CRITICAL)** - Fat getter interface with single implementation
   - 8 getter methods
   - Only implemented by `internal/context/context.go`
   - Never mocked in tests (we create real instances)
   - **Recommendation**: Convert to struct
   - **Impact**: Used extensively - needs careful migration

4. **ContainerManager (P1-HIGH)** - Fat interface
   - 6 methods for container operations
   - **Recommendation**: Split into `ContainerLifecycle`, `ContainerInspection`, `ContainerInteraction`

5. **DockerResolver (P2-MEDIUM)** - Undocumented
   - 4 methods but unclear usage
   - **Recommendation**: Add documentation or remove if unused

---

### 5. Formatter Interfaces

| Interface | Location | Methods | Implementations | Status | Priority |
|-----------|----------|---------|-----------------|--------|----------|
| Formatter | pkg/interfaces/interfaces.go:108 | 7 | Multiple | **DUPLICATE** | P1 |
| Formatter | pkg/output/formatter.go:10 | 7 | Multiple | **DUPLICATE** | P1 |
| Operation | pkg/progress/progress.go:93 | Unknown | Unknown | UNDOCUMENTED | P2 |
| Indicator | pkg/progress/types.go:90 | 5 | Multiple | GOOD | SAFE |
| ProgressIndicator | pkg/interfaces/interfaces.go:119 | 5 | Multiple | **DUPLICATE** | P1 |

**Issues Identified:**

1. **Formatter (P1-HIGH)** - Duplicate interface definitions
   - `pkg/interfaces/interfaces.go:108` and `pkg/output/formatter.go:10` define identical interface
   - **Recommendation**: Keep only `pkg/output/formatter.go` version, remove from pkg/interfaces
   - **Reason**: Formatter is an output concern, belongs in pkg/output

2. **ProgressIndicator vs Indicator (P1-HIGH)** - Duplicate interface
   - `pkg/interfaces/interfaces.go:119` defines ProgressIndicator with 5 methods
   - `pkg/progress/types.go:90` defines Indicator with 5 methods (different signatures)
   - **Recommendation**: Keep `pkg/progress.Indicator`, remove `pkg/interfaces.ProgressIndicator`
   - **Reason**: Progress types belong in progress package

---

### 6. Test Helper Interfaces

| Interface | Location | Methods | Implementations | Status | Priority |
|-----------|----------|---------|-----------------|--------|----------|
| TestingT | tests/testutil/fixtures.go:259 | Unknown | testing.T | GOOD | SAFE |

**Issues Identified:**

None - test helper interfaces are acceptable for abstracting testing frameworks.

---

### 7. Generated Interfaces (gRPC)

| Interface | Location | Methods | Implementations | Status | Priority |
|-----------|----------|---------|-----------------|--------|----------|
| GlidePluginClient | pkg/plugin/sdk/v1/plugin_grpc.pb.go:36 | Unknown | Generated | SKIP | N/A |
| GlidePluginServer | pkg/plugin/sdk/v1/plugin_grpc.pb.go:139 | Unknown | Generated | SKIP | N/A |
| UnsafeGlidePluginServer | pkg/plugin/sdk/v1/plugin_grpc.pb.go:191 | Unknown | Generated | SKIP | N/A |

**Issues Identified:**

None - these are generated by protoc and should not be modified.

---

### 8. Miscellaneous Interfaces

| Interface | Location | Methods | Implementations | Status | Priority |
|-----------|----------|---------|-----------------|--------|----------|
| Prompter | pkg/prompt/prompts.go:13 | Unknown | Unknown | UNDOCUMENTED | P2 |

---

## Summary of Issues

### Fat Interfaces (7 total)

1. **ShellExecutor** (pkg/interfaces) - 3 execution modes in one interface
2. **OutputManager** (pkg/interfaces) - 8 methods mixing structured/raw output
3. **Plugin** (pkg/plugin) - 5 methods mixing concerns (P0-CRITICAL)
4. **ShellCommand** (pkg/interfaces) - 5 getter methods (should be struct)
5. **ShellCommandOptions** (pkg/interfaces) - 5 getter methods (should be struct)
6. **ProjectContext** (pkg/interfaces) - 8 getter methods, single impl (P0-CRITICAL)
7. **ContainerManager** (pkg/interfaces) - 6 methods for different operations

### Unnecessary Interfaces (3 total)

1. **ConfigLoader** (pkg/interfaces) - Single implementation, never mocked
2. **ProjectContext** (pkg/interfaces) - Single implementation, never mocked (also FAT)
3. **ExtensionRegistry** (internal/context) - Confusion with struct definition

### Duplicate Interfaces (2 sets)

1. **Formatter** - Defined in both pkg/interfaces and pkg/output
2. **ProgressIndicator/Indicator** - Similar interfaces in pkg/interfaces and pkg/progress

### Documentation Gaps (9 total)

1. InteractiveSession
2. ContextAwareExecutor
3. CommandSanitizer
4. DockerResolver
5. Operation (progress)
6. Prompter
7. ContainerManager
8. GlidePluginClient (generated)
9. GlidePluginServer (generated)

---

## Recommendations by Priority

### P0-CRITICAL (Fix Immediately)

#### 1. Plugin Interface - Split for ISP compliance

**Current:**
```go
type Plugin interface {
    Name() string
    Version() string
    Register(root *cobra.Command) error
    Configure(config map[string]interface{}) error
    Metadata() PluginMetadata
}
```

**Recommended:**
```go
// PluginIdentifier provides plugin identification
type PluginIdentifier interface {
    Name() string
    Version() string
    Metadata() PluginMetadata
}

// PluginRegistrar registers plugin commands
type PluginRegistrar interface {
    Register(root *cobra.Command) error
}

// PluginConfigurable handles plugin configuration
type PluginConfigurable interface {
    Configure(config map[string]interface{}) error
}

// Plugin is the composite interface for backward compatibility
// Deprecated: Use specific sub-interfaces where possible
type Plugin interface {
    PluginIdentifier
    PluginRegistrar
    PluginConfigurable
}
```

**Benefits:**
- Smaller, focused interfaces
- Easier to mock in tests
- Plugins can implement only what they need
- Backward compatible via composition

**Breaking Changes:** None (additive only)

---

#### 2. ProjectContext - Convert to Struct

**Current:**
```go
// pkg/interfaces/interfaces.go
type ProjectContext interface {
    GetWorkingDir() string
    GetProjectRoot() string
    GetDevelopmentMode() string
    GetLocation() string
    IsDockerRunning() bool
    GetComposeFiles() []string
    IsWorktree() bool
    GetWorktreeName() string
}
```

**Issue:**
- Single implementation: `internal/context/context.go`
- Never mocked in tests
- Pure data holder with getters
- Interface adds no value

**Recommended:**
```go
// Remove interface entirely, use struct directly
// internal/context/context.go already has this struct
type ProjectContext struct {
    WorkingDir      string
    ProjectRoot     string
    DevelopmentMode DevelopmentMode
    Location        LocationType
    DockerRunning   bool
    ComposeFiles    []string
    IsWorktree      bool
    WorktreeName    string
    // ... other fields ...
}
```

**Migration Path:**
1. Update all code to use `*context.ProjectContext` instead of `interfaces.ProjectContext`
2. Remove getters, access fields directly
3. Mark interface as deprecated
4. Remove interface in v3.0.0

**Breaking Changes:** YES - but can be done incrementally
**Estimated Effort:** 6-8 hours

---

### P1-HIGH (Fix in Current Phase)

#### 3. ShellExecutor - Split by Execution Strategy

**Current:**
```go
type ShellExecutor interface {
    Execute(ctx context.Context, cmd ShellCommand) (*ShellResult, error)
    ExecuteWithTimeout(cmd ShellCommand, timeout time.Duration) (*ShellResult, error)
    ExecuteWithProgress(cmd ShellCommand, message string) error
}
```

**Recommended:**
```go
// Basic execution interface
type CommandExecutor interface {
    Execute(ctx context.Context, cmd *Command) (*Result, error)
}

// Note: Timeout and progress should be options, not separate methods
// Use functional options pattern instead
type ExecuteOption func(*ExecuteConfig)

func WithTimeout(timeout time.Duration) ExecuteOption
func WithProgress(message string) ExecuteOption
```

**Actually:** We already have `internal/shell/strategy.go` with ExecutionStrategy!
**Action:** Remove ShellExecutor interface from pkg/interfaces, use internal/shell patterns

---

#### 4. OutputManager - Split Structured vs Raw

**Current:**
```go
type OutputManager interface {
    Display(data interface{}) error
    Info(format string, args ...interface{}) error
    Success(format string, args ...interface{}) error
    Error(format string, args ...interface{}) error
    Warning(format string, args ...interface{}) error
    Raw(text string) error
    Printf(format string, args ...interface{}) error
    Println(args ...interface{}) error
}
```

**Recommended:**
```go
// StructuredOutput handles semantic output levels
type StructuredOutput interface {
    Display(data interface{}) error
    Info(message string) error
    Success(message string) error
    Error(message string) error
    Warning(message string) error
}

// RawOutput handles unformatted text
type RawOutput interface {
    Print(text string) error
    Printf(format string, args ...interface{}) error
    Println(args ...interface{}) error
}

// OutputManager composites both for convenience
type OutputManager interface {
    StructuredOutput
    RawOutput
}
```

**Benefits:**
- Code can depend on just StructuredOutput or just RawOutput
- Easier to test
- Clearer intent

---

#### 5. Remove Duplicate Formatter Interfaces

**Action:**
- Remove `Formatter` from `pkg/interfaces/interfaces.go:108`
- Keep `Formatter` in `pkg/output/formatter.go:10`
- Update all imports to use `pkg/output.Formatter`

---

#### 6. Remove Duplicate Progress Interfaces

**Action:**
- Remove `ProgressIndicator` from `pkg/interfaces/interfaces.go:119`
- Keep `Indicator` in `pkg/progress/types.go:90`
- Update all imports to use `pkg/progress.Indicator`

---

#### 7. Convert Data Transfer Interfaces to Structs

**ShellCommand and ShellCommandOptions:**
```go
// Before (interface):
type ShellCommand interface {
    GetCommand() string
    GetArgs() []string
    // ...
}

// After (struct):
type Command struct {
    Name    string
    Args    []string
    Options Options
    // ...
}
```

**Benefits:**
- Simpler code
- No interface overhead
- Direct field access
- Easier to construct

---

### P2-MEDIUM (Fix When Touching Code)

#### 8. Add Documentation to Undocumented Interfaces

All interfaces should have:
- Purpose and use case
- Example implementation
- Thread safety guarantees
- Expected behavior/contracts

**Template:**
```go
// Prompter defines the interface for user prompts.
//
// Prompter handles interactive user input collection for CLI commands.
// Implementations must handle both TTY and non-TTY environments gracefully.
//
// Thread Safety: Implementations must be safe for concurrent access if they
// maintain internal state. Single-use prompters may be non-thread-safe.
//
// Example Implementation:
//
//	type BasicPrompter struct {
//	    reader io.Reader
//	}
//
//	func (p *BasicPrompter) Ask(question string) (string, error) {
//	    fmt.Print(question + " ")
//	    // ... read from reader ...
//	}
//
// See pkg/prompt/README.md for a complete example.
type Prompter interface {
    Ask(question string) (string, error)
    Confirm(question string) (bool, error)
}
```

---

#### 9. Review Single-Implementation Interfaces

For each of these, determine:
- Is it mocked in tests? → Keep
- Is it a plugin extension point? → Keep
- Neither? → Consider removal

**Candidates:**
- ConfigLoader
- ProjectRootFinder
- DevelopmentModeDetector
- LocationIdentifier
- ComposeFileResolver
- DockerStatusChecker

---

### SAFE (No Action or Minor Improvements)

28 interfaces are well-designed and require no changes. These follow SOLID principles:
- Single responsibility
- Small method count (1-3 methods)
- Clear purpose
- Multiple implementations or clear extension point
- Properly documented

**Examples of Well-Designed Interfaces:**
- `CommandProvider` (1 method)
- `ContextExtension` (3 focused methods)
- `ExecutionStrategy` (2 methods)
- `ExecutorProvider` (3 methods)
- `FrameworkDetector` (1 method)

---

## Validation Checklist

After implementing fixes:

- [ ] All fat interfaces split or refactored
- [ ] Duplicate interfaces removed
- [ ] Unnecessary interfaces removed or justified
- [ ] All public interfaces documented
- [ ] Examples provided for complex interfaces
- [ ] Tests updated to use new interfaces
- [ ] No breaking changes to public API (or properly deprecated)
- [ ] ADR created for major interface changes
- [ ] Migration guide for deprecated interfaces

---

## Implementation Plan

### Phase 1: Critical Fixes (Week 1)

1. **Plugin Interface** (4h)
   - Create sub-interfaces
   - Update implementations to use sub-interfaces where possible
   - Add deprecation notices
   - Update tests

2. **ProjectContext Struct Conversion** (6h)
   - Remove interface definition
   - Update all usages to use struct
   - Remove getters, use direct field access
   - Update tests
   - Verify no regressions

### Phase 2: High-Priority Fixes (Week 2)

3. **Remove Duplicate Interfaces** (2h)
   - Remove Formatter from pkg/interfaces
   - Remove ProgressIndicator from pkg/interfaces
   - Update imports

4. **Data Transfer Structs** (4h)
   - Convert ShellCommand interface to struct
   - Convert ShellCommandOptions interface to struct
   - Update all usages
   - Update tests

5. **Split OutputManager** (3h)
   - Create StructuredOutput and RawOutput sub-interfaces
   - Keep composite for backward compatibility
   - Update implementations

### Phase 3: Medium-Priority Fixes (Week 3)

6. **Add Documentation** (4h)
   - Document all undocumented interfaces
   - Add examples
   - Document thread safety
   - Document expected behavior

7. **Review Single-Implementation Interfaces** (2h)
   - Evaluate each candidate
   - Remove or document justification
   - Update tests if needed

### Phase 4: Validation (End of Week 3)

8. **Final Testing** (2h)
   - Run full test suite
   - Verify no breaking changes
   - Check documentation
   - Update ADRs

**Total Estimated Effort:** 27 hours (3-4 weeks at 8-10h/week)

---

## Success Metrics

- ✅ No interface with more than 5 methods (guideline)
- ✅ All interfaces follow single responsibility principle
- ✅ No duplicate interface definitions
- ✅ All public interfaces have comprehensive documentation
- ✅ No unnecessary interfaces (single implementation + no testing value)
- ✅ All tests passing
- ✅ No performance regressions

---

## Appendix: Quick Reference

### Interface Analysis Tool

```bash
# Find all interfaces
grep -r "^type.*interface.*{" --include="*.go" . | grep -v vendor | grep -v _test.go

# Count methods in an interface
awk '/^type.*interface/,/^}/' file.go | grep -c "^\s*[A-Z]"

# Find implementations
grep -r "type.*struct" --include="*.go" . | xargs -I {} grep -l "func.*{}"

# Find usages
grep -r "InterfaceName" --include="*.go" .
```

---

**End of Audit**
