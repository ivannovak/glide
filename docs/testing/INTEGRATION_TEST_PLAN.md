# Integration Test Plan

**Date:** 2025-11-28
**Phase:** Phase 2: Testing Infrastructure
**Purpose:** Identify gaps in integration test coverage and plan new integration tests

## Table of Contents

- [Executive Summary](#executive-summary)
- [Existing Integration Tests](#existing-integration-tests)
- [Gap Analysis](#gap-analysis)
- [New Integration Tests](#new-integration-tests)
- [Test Templates](#test-templates)
- [Implementation Plan](#implementation-plan)

---

## Executive Summary

### Current State

- **Integration Tests:** 35 tests across 6 test files
- **Coverage Areas:** Dependencies, Docker, Modes, Pass-through commands, Setup, Worktrees
- **Test Structure:** Using testify/assert, t.TempDir() for isolation, real components

### Gaps Identified

**CRITICAL GAPS:**
1. **No plugin loading integration tests** - Plugin discovery, loading, unloading
2. **No config merging tests** - Global + local config interaction
3. **No multi-framework context detection tests** - Context detection across different frameworks
4. **No command execution pipeline tests** - End-to-end command flow

**HIGH PRIORITY GAPS:**
5. Config validation integration
6. Plugin command execution integration
7. Error propagation across layers
8. State management across operations

### Recommendation

Add **12 new integration test files** covering the critical gaps, focusing on:
- Plugin workflows (loading, execution, lifecycle)
- Config operations (discovery, merging, validation)
- Context detection (multi-framework, caching)
- Command execution (pipelines, error handling)

---

## Existing Integration Tests

### 1. Dependencies Tests (`dependencies_test.go`)

**Coverage:** 7 tests, ~10KB

| Test | Purpose |
|------|---------|
| TestDockerDependency | Verify Docker availability |
| TestAWSCLIDependency | Verify AWS CLI availability |
| TestPHPToolsDependency | Verify PHP tools availability |
| TestGitDependency | Verify Git availability |
| TestMakeDependency | Verify Make availability |
| TestSystemRequirements | Verify all system requirements |
| TestDependencyErrorMessages | Verify error messages for missing dependencies |

**What's Tested:**
- System dependency checks
- Error messages for missing tools
- Requirement validation

**What's Not Tested:**
- Dependency version requirements
- Optional vs required dependencies

---

### 2. Docker Tests (`docker_test.go`)

**Coverage:** 8 tests, ~13.5KB

| Test | Purpose |
|------|---------|
| TestDockerOperations | Basic Docker operations |
| TestDockerContainerLifecycle | Container create/start/stop/remove |
| TestDockerNetworking | Container networking |
| TestDockerVolumeOperations | Volume create/mount/remove |
| TestDockerComposeIntegration | Docker Compose integration |
| TestDockerHealthChecks | Container health checking |
| TestDockerResourceCleanup | Resource cleanup after tests |
| TestDockerErrorHandling | Docker error scenarios |

**What's Tested:**
- Full Docker lifecycle
- Networking and volumes
- Compose integration
- Error handling

**What's Not Tested:**
- Docker registry interaction
- Multi-container dependencies
- Port conflict handling

---

### 3. Modes Tests (`modes_test.go`)

**Coverage:** 3 tests, ~13KB

| Test | Purpose |
|------|---------|
| TestModeTransitions | Switching between modes |
| TestModeSpecificCommands | Commands available in each mode |
| TestContextPreservation | Context preserved across mode changes |

**What's Tested:**
- Mode switching
- Mode-specific command availability
- Context persistence

**What's Not Tested:**
- Invalid mode transitions
- Mode validation
- Mode configuration

---

### 4. Pass-through Tests (`passthrough_test.go`)

**Coverage:** 8 tests, ~11.5KB

| Test | Purpose |
|------|---------|
| TestPassThroughCommands | Generic pass-through |
| TestComposerPassThrough | Composer command forwarding |
| TestGitPassThrough | Git command forwarding |
| TestNPMPassThrough | NPM command forwarding |
| TestInteractiveCommands | Interactive command handling |
| TestComplexArgumentHandling | Complex argument parsing |
| TestSignalPropagation | Signal forwarding to child processes |
| TestOutputStreaming | Real-time output streaming |
| TestPassThroughMode | Pass-through mode behavior |

**What's Tested:**
- Command forwarding to tools
- Interactive commands
- Signal and output handling
- Argument parsing

**What's Not Tested:**
- Pass-through errors
- Multiple simultaneous pass-throughs
- Pass-through with plugins

---

### 5. Setup Tests (`setup_test.go`)

**Coverage:** 5 tests, ~12KB

| Test | Purpose |
|------|---------|
| TestSetupCommand | Setup command execution |
| TestConfigurationCreation | Config file creation |
| TestProjectDetection | Project type detection |
| TestModeSelection | Mode selection during setup |
| TestEnvironmentValidation | Environment validation |

**What's Tested:**
- Setup workflow
- Config creation
- Project detection
- Environment validation

**What's Not Tested:**
- Setup in existing projects
- Setup rollback on error
- Setup with custom config

---

### 6. Worktree Tests (`worktree_test.go`)

**Coverage:** 3 tests, ~11KB

| Test | Purpose |
|------|---------|
| TestWorktreeManagement | Worktree discovery and management |
| TestWorktreeOperations | Operations in worktree context |
| TestWorktreeWithDocker | Docker in worktree mode |

**What's Tested:**
- Worktree detection
- Multi-worktree operations
- Docker in worktree mode

**What's Not Tested:**
- Worktree creation/deletion
- Cross-worktree operations
- Worktree conflicts

---

## Gap Analysis

### Critical Gaps (P0)

#### Gap 1: Plugin Loading Integration ⚠️ CRITICAL

**Missing Tests:**
- Plugin discovery from multiple sources (local, remote, registry)
- Plugin validation (security, compatibility)
- Plugin loading order and dependencies
- Plugin conflict resolution
- Plugin unloading and cleanup

**Impact:** HIGH - Plugin system is core functionality with low coverage (8.6%)

**Priority:** P0

**Files Affected:**
- `pkg/plugin/`
- `pkg/plugin/sdk/v1/`
- `internal/cli/` (plugin command integration)

**Test File:** `tests/integration/plugin_test.go` (NEW)

---

#### Gap 2: Config Merging Integration ⚠️ CRITICAL

**Missing Tests:**
- Global config + local config merging
- Environment variable overrides
- Config precedence rules
- Config validation across layers
- Config modification and persistence

**Impact:** HIGH - Config system critical, only 26.7% coverage

**Priority:** P0

**Files Affected:**
- `internal/config/`

**Test File:** `tests/integration/config_test.go` (NEW)

---

#### Gap 3: Context Detection Integration ⚠️ CRITICAL

**Missing Tests:**
- Multi-framework detection (PHP + Node + Docker)
- Context caching and invalidation
- Plugin-enhanced context detection
- Context detection in edge cases (monorepos, nested projects)

**Impact:** MEDIUM-HIGH - Context affects command behavior

**Priority:** P0

**Files Affected:**
- `internal/context/`
- `internal/detection/`

**Test File:** `tests/integration/context_test.go` (NEW)

---

#### Gap 4: Command Execution Pipeline ⚠️ CRITICAL

**Missing Tests:**
- Full command execution flow (parse → validate → execute → output)
- Error propagation through layers
- Pre/post hooks execution
- Command aliasing
- YAML command execution with sanitization

**Impact:** HIGH - CLI is user-facing, only 12.1% coverage

**Priority:** P0

**Files Affected:**
- `internal/cli/`
- `pkg/plugin/`

**Test File:** `tests/integration/command_pipeline_test.go` (NEW)

---

### High Priority Gaps (P1)

#### Gap 5: Config Validation Integration

**Missing Tests:**
- Invalid config handling
- Required field validation
- Type validation
- Cross-field validation

**Test File:** `tests/integration/config_test.go` (expand)

---

#### Gap 6: Plugin Command Execution

**Missing Tests:**
- Plugin command discovery
- Plugin command execution
- Plugin command errors
- Plugin command output

**Test File:** `tests/integration/plugin_test.go` (expand)

---

#### Gap 7: Error Handling Integration

**Missing Tests:**
- Error propagation across components
- Error formatting and suggestions
- Error recovery
- Error logging

**Test File:** `tests/integration/error_handling_test.go` (NEW)

---

#### Gap 8: Output Manager Integration

**Missing Tests:**
- Output formatting (JSON, YAML, table)
- Output level filtering
- Progress indicators
- Multi-output writers

**Test File:** `tests/integration/output_test.go` (NEW)

---

## New Integration Tests

### Test File 1: `plugin_test.go` (NEW)

**Purpose:** Test plugin loading, validation, and execution workflows

**Tests to Add:**

```go
// Plugin Discovery & Loading
TestPluginDiscovery                  // Discover plugins from multiple sources
TestPluginLoading                    // Load plugins in correct order
TestPluginValidation                 // Validate plugin security and compatibility
TestPluginConflictResolution         // Handle plugin conflicts
TestPluginUnloading                  // Unload plugins and cleanup

// Plugin Execution
TestPluginCommandDiscovery           // Discover commands from loaded plugins
TestPluginCommandExecution           // Execute plugin commands
TestPluginCommandErrors              // Handle plugin command errors
TestPluginCommandOutput              // Capture plugin command output

// Plugin Dependencies
TestPluginDependencies               // Load plugins with dependencies
TestPluginDependencyOrder            // Correct dependency resolution order
TestPluginCircularDependencies       // Detect circular dependencies

// Plugin Lifecycle
TestPluginConfiguration              // Configure plugins
TestPluginInitialization             // Initialize plugins
TestPluginShutdown                   // Graceful plugin shutdown
```

**Estimated Effort:** 12 hours (24 tests)

---

### Test File 2: `config_test.go` (NEW)

**Purpose:** Test config discovery, merging, and validation

**Tests to Add:**

```go
// Config Discovery
TestConfigDiscovery                  // Discover .glide.yml files
TestConfigParentSearch               // Search parent directories
TestConfigHomeDirectory              // Home directory fallback
TestConfigNotFound                   // Handle no config found

// Config Merging
TestConfigMerging                    // Merge global + local configs
TestConfigPrecedence                 // Config precedence rules
TestEnvVarOverrides                  // Environment variable overrides
TestConfigInheritance                // Config value inheritance

// Config Validation
TestInvalidConfig                    // Handle invalid YAML
TestMissingRequiredFields            // Validate required fields
TestTypeValidation                   // Validate field types
TestCrossFieldValidation             // Cross-field validation

// Config Modification
TestConfigSet                        // Set config values
TestConfigUnset                      // Unset config values
TestConfigSave                       // Save config changes
TestConfigRollback                   // Rollback on error
```

**Estimated Effort:** 10 hours (20 tests)

---

### Test File 3: `context_test.go` (NEW - currently only TestContextPreservation exists)

**Purpose:** Test context detection across frameworks and scenarios

**Tests to Add:**

```go
// Multi-Framework Detection
TestPHPProjectDetection              // Detect PHP projects (composer.json)
TestNodeProjectDetection             // Detect Node projects (package.json)
TestDockerProjectDetection           // Detect Docker projects (docker-compose.yml)
TestMultiFrameworkDetection          // Detect multiple frameworks

// Context Caching
TestContextCaching                   // Cache context detection results
TestContextInvalidation              // Invalidate cache on changes
TestContextRefresh                   // Refresh context manually

// Plugin-Enhanced Detection
TestPluginContextEnhancement         // Plugins enhance context
TestPluginContextDetectors           // Custom context detectors from plugins

// Edge Cases
TestMonorepoDetection                // Detect monorepo structure
TestNestedProjectDetection           // Detect nested projects
TestSymlinkProjectDetection          // Handle symlinks correctly
TestEmptyProjectDetection            // Handle empty directories
```

**Estimated Effort:** 8 hours (16 tests)

---

### Test File 4: `command_pipeline_test.go` (NEW)

**Purpose:** Test end-to-end command execution pipeline

**Tests to Add:**

```go
// Command Parsing
TestCommandParsing                   // Parse commands correctly
TestFlagParsing                      // Parse flags correctly
TestArgumentParsing                  // Parse arguments correctly
TestAliasResolution                  // Resolve command aliases

// Command Validation
TestCommandValidation                // Validate commands before execution
TestUnknownCommand                   // Handle unknown commands
TestInvalidArguments                 // Handle invalid arguments
TestMissingRequiredArgs              // Handle missing required arguments

// Command Execution
TestCommandExecution                 // Execute commands end-to-end
TestYAMLCommandExecution             // Execute YAML-defined commands
TestYAMLCommandSanitization          // Verify YAML sanitization
TestPreHooks                         // Execute pre-hooks
TestPostHooks                        // Execute post-hooks

// Error Handling
TestCommandExecutionErrors           // Handle execution errors
TestTimeoutHandling                  // Handle command timeouts
TestSignalInterruption               // Handle signal interruptions
```

**Estimated Effort:** 10 hours (18 tests)

---

### Test File 5: `error_handling_test.go` (NEW)

**Purpose:** Test error propagation and handling across layers

**Tests to Add:**

```go
// Error Propagation
TestErrorPropagation                 // Errors propagate through layers
TestErrorWrapping                    // Errors wrapped with context
TestErrorUnwrapping                  // Errors unwrapped correctly

// Error Formatting
TestUserErrorFormatting              // User-facing error messages
TestSystemErrorFormatting            // System error messages
TestPluginErrorFormatting            // Plugin error messages
TestErrorSuggestions                 // Error suggestions displayed

// Error Recovery
TestGracefulDegradation              // Graceful degradation on errors
TestErrorRecovery                    // Recover from errors when possible
TestRollbackOnError                  // Rollback changes on error

// Error Logging
TestErrorLogging                     // Errors logged correctly
TestDebugErrorLogging                // Debug mode error details
TestErrorMetrics                     // Error metrics collected
```

**Estimated Effort:** 8 hours (15 tests)

---

### Test File 6: `output_test.go` (NEW)

**Purpose:** Test output formatting and management

**Tests to Add:**

```go
// Output Formatting
TestPlainOutput                      // Plain text output
TestJSONOutput                       // JSON output formatting
TestYAMLOutput                       // YAML output formatting
TestTableOutput                      // Table output formatting

// Output Levels
TestDebugOutput                      // Debug level output
TestInfoOutput                       // Info level output
TestWarningOutput                    // Warning level output
TestErrorOutput                      // Error level output

// Progress Indicators
TestSpinnerProgress                  // Spinner progress indicator
TestProgressBar                      // Progress bar indicator
TestMultiProgress                    // Multiple progress indicators

// Output Writers
TestStdoutOutput                     // Output to stdout
TestStderrOutput                     // Output to stderr
TestFileOutput                       // Output to file
TestBufferedOutput                   // Buffered output capture
```

**Estimated Effort:** 8 hours (16 tests)

---

## Test Templates

### Template 1: Plugin Integration Test

```go
package integration_test

import (
    "path/filepath"
    "testing"

    "github.com/glide-cli/glide/v3/pkg/plugin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestPluginDiscovery(t *testing.T) {
    // Setup: Create temp directory with mock plugins
    tmpDir := t.TempDir()
    pluginDir := filepath.Join(tmpDir, ".glide", "plugins")
    require.NoError(t, os.MkdirAll(pluginDir, 0755))

    // Create mock plugin binary
    pluginPath := filepath.Join(pluginDir, "test-plugin")
    createMockPlugin(t, pluginPath)

    // Test: Discover plugins
    registry := plugin.NewRegistry(pluginDir)
    plugins, err := registry.Discover()

    // Assert
    require.NoError(t, err)
    assert.Len(t, plugins, 1)
    assert.Equal(t, "test-plugin", plugins[0].Name)
}

func createMockPlugin(t *testing.T, path string) {
    // Create a minimal plugin binary that implements the protocol
    // ...
}
```

### Template 2: Config Integration Test

```go
func TestConfigMerging(t *testing.T) {
    // Setup: Create global and local configs
    tmpDir := t.TempDir()
    homeDir := filepath.Join(tmpDir, "home")
    projectDir := filepath.Join(tmpDir, "project")
    require.NoError(t, os.MkdirAll(homeDir, 0755))
    require.NoError(t, os.MkdirAll(projectDir, 0755))

    // Create global config
    globalConfig := `
plugins:
  - name: global-plugin
settings:
  theme: dark
`
    require.NoError(t, os.WriteFile(
        filepath.Join(homeDir, ".glide.yml"),
        []byte(globalConfig),
        0644,
    ))

    // Create local config
    localConfig := `
settings:
  theme: light  # Override global
  language: en
`
    require.NoError(t, os.WriteFile(
        filepath.Join(projectDir, ".glide.yml"),
        []byte(localConfig),
        0644,
    ))

    // Test: Load merged config
    os.Setenv("HOME", homeDir)
    defer os.Unsetenv("HOME")

    cfg, err := config.Load(projectDir)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, "light", cfg.Settings.Theme) // Local overrides global
    assert.Equal(t, "en", cfg.Settings.Language)
    assert.Len(t, cfg.Plugins, 1) // Global plugin included
}
```

### Template 3: Context Detection Test

```go
func TestMultiFrameworkDetection(t *testing.T) {
    // Setup: Create project with multiple frameworks
    tmpDir := t.TempDir()
    require.NoError(t, os.MkdirAll(tmpDir, 0755))

    // Create PHP project files
    require.NoError(t, os.WriteFile(
        filepath.Join(tmpDir, "composer.json"),
        []byte(`{"require": {"php": "^8.0"}}`),
        0644,
    ))

    // Create Node project files
    require.NoError(t, os.WriteFile(
        filepath.Join(tmpDir, "package.json"),
        []byte(`{"name": "test", "version": "1.0.0"}`),
        0644,
    ))

    // Create Docker files
    require.NoError(t, os.WriteFile(
        filepath.Join(tmpDir, "docker-compose.yml"),
        []byte(`version: '3'\nservices:\n  web:\n    image: nginx`),
        0644,
    ))

    // Test: Detect context
    originalWd, _ := os.Getwd()
    defer os.Chdir(originalWd)
    require.NoError(t, os.Chdir(tmpDir))

    ctx := context.Detect()

    // Assert
    require.NotNil(t, ctx)
    assert.True(t, ctx.HasPHP)
    assert.True(t, ctx.HasNode)
    assert.True(t, ctx.HasDocker)
    assert.Contains(t, ctx.Frameworks, "PHP")
    assert.Contains(t, ctx.Frameworks, "Node")
}
```

### Template 4: Command Pipeline Test

```go
func TestYAMLCommandExecution(t *testing.T) {
    // Setup: Create temp directory with .glide.yml
    tmpDir := t.TempDir()
    configPath := filepath.Join(tmpDir, ".glide.yml")

    config := `
commands:
  hello:
    exec: echo "Hello, World!"
  ls:
    exec: ls -la
`
    require.NoError(t, os.WriteFile(configPath, []byte(config), 0644))

    originalWd, _ := os.Getwd()
    defer os.Chdir(originalWd)
    require.NoError(t, os.Chdir(tmpDir))

    // Test: Execute YAML command
    rootCmd := setupTestCLI(t)
    var buf bytes.Buffer
    rootCmd.SetOut(&buf)
    rootCmd.SetArgs([]string{"hello"})

    err := rootCmd.Execute()

    // Assert
    require.NoError(t, err)
    assert.Contains(t, buf.String(), "Hello, World!")
}
```

---

## Implementation Plan

### Phase 2.6: Integration Tests & E2E (20 hours)

#### Subtask 2.6.1: Expand Integration Tests (8h)

**Week 1: Critical Gap Tests (Days 1-2)**
- [ ] Create `plugin_test.go` (4h)
  - Plugin discovery tests
  - Plugin loading tests
  - Plugin execution tests

- [ ] Create `config_test.go` (4h)
  - Config discovery tests
  - Config merging tests
  - Config validation tests

**Week 1: Additional Gap Tests (Days 3-4)**
- [ ] Expand `context_test.go` (2h)
  - Multi-framework detection
  - Context caching
  - Plugin-enhanced detection

- [ ] Create `command_pipeline_test.go` (2h)
  - Command parsing tests
  - Command validation tests
  - YAML command execution tests

**Week 2: Error & Output Tests (Days 1-2)**
- [ ] Create `error_handling_test.go` (2h)
  - Error propagation tests
  - Error formatting tests
  - Error recovery tests

- [ ] Create `output_test.go` (2h)
  - Output formatting tests
  - Output level tests
  - Progress indicator tests

**Total:** 8 hours

---

#### Verification

After each test file:
```bash
# Run new integration tests
go test ./tests/integration/plugin_test.go -v
go test ./tests/integration/config_test.go -v
go test ./tests/integration/context_test.go -v
go test ./tests/integration/command_pipeline_test.go -v
go test ./tests/integration/error_handling_test.go -v
go test ./tests/integration/output_test.go -v

# Run all integration tests
go test ./tests/integration/... -v

# All should pass
```

---

## Success Metrics

### Coverage Metrics

- **New Tests:** +100 integration tests (35 → 135)
- **New Test Files:** +6 integration test files (6 → 12)
- **Critical Gaps Closed:** 4/4 (100%)
- **High Priority Gaps Closed:** 4/4 (100%)

### Quality Metrics

- All integration tests passing
- No flaky tests (<1% flake rate)
- Integration tests run in <2 minutes
- Clear test names and assertions

### Documentation

- All new tests documented in this plan
- Test templates provided
- Integration with unit tests clear

---

## Appendix

### Test File Organization

```
tests/integration/
├── dependencies_test.go      # System dependencies ✅ EXISTS
├── docker_test.go            # Docker operations ✅ EXISTS
├── modes_test.go             # Mode transitions ✅ EXISTS
├── passthrough_test.go       # Pass-through commands ✅ EXISTS
├── setup_test.go             # Setup workflow ✅ EXISTS
├── worktree_test.go          # Worktree operations ✅ EXISTS
├── plugin_test.go            # ⚠️ NEW - Plugin workflows
├── config_test.go            # ⚠️ NEW - Config operations
├── context_test.go           # ⚠️ EXPAND - Context detection
├── command_pipeline_test.go  # ⚠️ NEW - Command execution
├── error_handling_test.go    # ⚠️ NEW - Error propagation
└── output_test.go            # ⚠️ NEW - Output formatting
```

### References

- **Coverage Analysis:** `docs/testing/COVERAGE_ANALYSIS.md`
- **Phase 2 Tasks:** `docs/specs/gold-standard-remediation/implementation-checklist.md` (lines 2703-2803)
- **Existing Tests:** `tests/integration/*.go`
- **Test Helpers:** `tests/testutil/README.md`
