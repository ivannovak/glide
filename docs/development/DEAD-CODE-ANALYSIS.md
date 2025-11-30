# Dead Code Analysis

**Date:** 2025-11-29
**Phase:** Phase 6 Technical Debt Cleanup

## Summary

Dead code analysis was performed using `golang.org/x/tools/cmd/deadcode`.

### Code Removed

| File | Functions Removed | Reason |
|------|-------------------|--------|
| `internal/cli/framework_commands.go` | Entire file | Never integrated into application |
| `internal/shell/plugin_aware.go` | Entire file | Unused executor wrapper |
| `internal/shell/registry.go` | Entire file | Only used by removed code |
| `internal/cli/worktree.go` | `ExecuteWorktreeCommand` | Unused convenience wrapper |
| `internal/config/config.go` | `Save` | Never called |

### Intentionally Retained Dead Code

The following categories of "dead code" are intentionally retained:

#### 1. Plugin SDK Infrastructure (~400 functions)

Located in:
- `pkg/plugin/sdk/` - Plugin development SDK
- `pkg/plugin/sdk/v1/` - v1 plugin interfaces
- `pkg/plugin/sdk/v2/` - v2 plugin interfaces

**Rationale:** These are public API surfaces for plugin developers. They may not be called by the core Glide application but are essential for third-party plugin development.

#### 2. Test Utilities (~200 functions)

Located in:
- `internal/mocks/` - Mock implementations for testing
- `pkg/plugin/plugintest/` - Plugin testing harness
- `tests/testutil/` - Test helper functions
- `tests/contracts/` - Contract test framework

**Rationale:** Test utilities are used in tests but appear "dead" to static analysis that only considers production code paths.

#### 3. Progress/UI Components (~100 functions)

Located in:
- `pkg/progress/` - Progress bars, spinners, multi-progress
- `pkg/output/` - Output formatters, colors

**Rationale:** UI components for future features. Some are used conditionally or only in specific code paths.

#### 4. Observability Infrastructure (~50 functions)

Located in:
- `pkg/observability/health.go` - Health check framework
- `pkg/observability/logging.go` - Performance logging
- `pkg/observability/metrics.go` - Metrics collection

**Rationale:** Infrastructure for monitoring that may be enabled via configuration or used in production deployments.

## Recommendations

1. **Plugin SDK**: Keep all SDK functions as they form the public API
2. **Test Utilities**: Keep for comprehensive testing support
3. **Progress/UI**: Keep for feature completeness
4. **Observability**: Keep for production deployment support

## Future Work

- Add `//go:embed` comments to intentionally unused but public API functions
- Consider moving truly dead internal code to `internal/deprecated/` if needed for reference
- Regular dead code audits as part of release process
