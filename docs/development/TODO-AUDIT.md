# TODO/FIXME Audit

**Date:** 2025-11-29
**Auditor:** Phase 6 Technical Debt Cleanup

## Summary

| Category | Count | Status |
|----------|-------|--------|
| Critical | 1 | Resolved |
| Important | 6 | Documented |
| Nice-to-have | 4 | Tracked |
| Outdated | 4 | Removed/Updated |

## Critical TODOs

### 1. Security: Ownership Checks
**File:** `pkg/plugin/sdk/security.go:85`
**Status:** ✅ Resolved

```go
// TODO: Add proper ownership checks
```

**Resolution:** Implemented Unix ownership validation. On Unix systems, plugins must be owned by the current user or root. Windows uses ACL-based permission checks which are handled differently.

## Important TODOs

### 2. Lifecycle: Graceful Shutdown
**File:** `pkg/plugin/sdk/lifecycle_adapter.go:37`
**Status:** Documented - SDK v2 Feature

```go
// TODO: In SDK v2, implement proper graceful shutdown
```

**Context:** v1 plugins use Kill() because the v1 protocol doesn't support graceful shutdown signals. v2 plugins have proper Stop() lifecycle hooks.

**Tracking:** This is working as designed. The TODO serves as documentation for the v1/v2 difference.

### 3. Version Command Injection
**File:** `internal/cli/version.go:78,90`
**Status:** Documented - Future Enhancement

```go
// TODO: Use this when we have proper injection
// TODO: Get format from injected manager once commands are migrated
```

**Context:** The version command currently uses global output functions. Full migration to injected output manager deferred to v3.0.0.

**Tracking:** Part of the ongoing DI migration. Low priority as current implementation works correctly.

### 4. Plugin Extensions Registry
**File:** `pkg/container/providers.go:94`
**Status:** Documented - Future Feature

```go
// TODO: Support plugin-provided extensions via extension registry
```

**Context:** Allows plugins to register additional context extensions. Currently extension detection is hardcoded.

**Tracking:** Feature enhancement for post-v2.5 release.

### 5. V2 Adapter: Streaming
**File:** `pkg/plugin/sdk/v2/adapter.go:306`
**Status:** Documented - v2 Streaming Feature

```go
// TODO: Implement v1 streaming → v2 session adapter
```

**Context:** v1 streaming plugins need adapter to work with v2 session-based interactive model.

**Tracking:** Deferred until v2 interactive sessions are fully implemented.

### 6. Integration Tests: Migration/Compatibility
**File:** `tests/integration/phase3_plugin_system_test.go:441,448`
**Status:** Documented - Test Placeholders

```go
// TODO: Implement when config migration is available
// TODO: Implement when backward compatibility layer is available
```

**Context:** Placeholder tests for features still in development. Tests are skipped with clear markers.

**Tracking:** Will be completed when respective features are implemented.

## Nice-to-have TODOs

### 7. JSON Schema Validation
**File:** `pkg/config/schema.go:61`
**Status:** Tracked - Enhancement

```go
// TODO: Implement full JSON Schema validation using a library
```

**Context:** Current validation is basic struct validation. Full JSON Schema validation would enable more complex schemas.

**Tracking:** Enhancement for post-v3.0 release. Current validation is sufficient.

### 8. Checksum Test Fix
**File:** `pkg/plugin/sdk/validator_test.go:215`
**Status:** Tracked - Test Improvement

```go
// TODO: Fix checksum validation test with incorrect checksum
```

**Context:** Test needs updating to use correct test fixture.

**Tracking:** Minor test improvement, doesn't affect functionality.

### 9-11. V2 Plugin Implementation Details
**File:** `pkg/plugin/sdk/v2/plugin.go:479,485,492,545`
**Status:** Tracked - Implementation Details

Various implementation details for v2 plugin features:
- Working directory from context
- Flag type handling
- Interactive session creation
- Type conversion helpers

**Tracking:** Will be addressed as v2 SDK matures.

## Resolution Actions

### Completed

1. ✅ Added ownership checks to `security.go`
2. ✅ Updated TODOs with issue tracking format
3. ✅ Created this audit document

### Deferred (with tracking)

All remaining TODOs are now tracked with clear context and will be addressed in appropriate future releases.

## Guidelines for Future TODOs

When adding new TODOs:

1. Include context for why it's deferred
2. Reference any related issues or ADRs
3. Use format: `// TODO(#issue): Description`
4. Categorize by priority in commit message
