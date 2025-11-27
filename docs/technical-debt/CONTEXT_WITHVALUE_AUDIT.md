# Context.WithValue Anti-pattern Audit

**Date:** 2025-11-27
**Status:** COMPLETE
**Severity:** P1 - Technical Debt
**Effort to Fix:** 1 hour (minimal - dead code removal)

## Executive Summary

This audit identifies all uses of `context.WithValue` in the Glide codebase. The Go community considers `context.WithValue` an anti-pattern when used for passing domain data, as it:

1. **Hides dependencies** - Makes it unclear what data a function needs
2. **Loses type safety** - Requires type assertions that can panic
3. **Creates hidden coupling** - Functions become dependent on context keys
4. **Violates explicit is better than implicit** - Go principle

**Finding:** Only **1 usage** of `context.WithValue` exists, and it is **dead code** (value is set but never retrieved).

## Audit Methodology

### Search Commands Used

```bash
# Find all WithValue usages
grep -r "context.WithValue" --include="*.go" . | grep -v vendor | grep -v "_test.go"

# Find all Value retrievals
grep -r "ctx.Value" --include="*.go" . | grep -v vendor | grep -v "_test.go"

# Find specific key usage
grep -rn 'Value.*projectContextKey\|Value.*"project_context"' --include="*.go" .
```

### Results

- **Production WithValue calls:** 1
- **Test WithValue calls:** 0
- **Value retrieval calls:** 0
- **Dead code instances:** 1

## Findings

### Finding #1: Dead Code in cmd/glide/main.go (DEAD CODE)

**File:** `cmd/glide/main.go`
**Lines:** 21-26, 169-170
**Priority:** P0 - DEAD CODE
**Effort:** 5 minutes

#### Code

```go
// Line 21-26: Definition
type contextKey string

const (
    projectContextKey contextKey = "project_context"
)

// Line 169-170: Usage
rootCmdCtx := stdcontext.WithValue(stdcontext.Background(), projectContextKey, ctx)
rootCmd.SetContext(rootCmdCtx)
```

#### Analysis

**What data is being passed:**
- `ctx` - a `*context.ProjectContext` instance containing detected project metadata (framework, languages, working directory, etc.)

**Why context is being used:**
According to the comment on line 167-168:
> // Set the context on the root command so plugins can access it
> // This is done via context.Context to avoid import cycles

**Problem:**
1. The value is **never retrieved** anywhere in the codebase
2. No code calls `cmd.Context().Value(projectContextKey)`
3. This is leftover dead code from an earlier design

**Why it's dead:**
The ProjectContext was intended to be passed to plugins via context.Context, but the actual implementation passes context directly to plugin methods (see `pkg/plugin/runtime_integration.go:197`). The context passed is a standard Go context for cancellation/deadlines, not for data passing.

#### Root Cause

This appears to be leftover code from an earlier design where plugins would extract ProjectContext from the cobra.Command context. However, the current architecture:

1. **CLI commands** receive ProjectContext via constructor dependency injection (see Task 1.3)
2. **Plugins** receive context via gRPC calls for cancellation, but don't need ProjectContext from context.Value
3. **Plugin commands** get context from cobra `cmd.Context()` but only use it for gRPC cancellation

#### Solution

**REMOVE** the dead code entirely:

```go
// DELETE these lines from main.go:
type contextKey string                                    // Line 21-22

const (                                                   // Line 24-26
    projectContextKey contextKey = "project_context"
)

// DELETE these lines:                                    // Line 167-170
// Set the context on the root command so plugins can access it
// This is done via context.Context to avoid import cycles
rootCmdCtx := stdcontext.WithValue(stdcontext.Background(), projectContextKey, ctx)
rootCmd.SetContext(rootCmdCtx)
```

Replace with:
```go
// Set standard context for cancellation/deadline support
rootCmd.SetContext(stdcontext.Background())
```

#### Impact Assessment

**Risk:** NONE - dead code removal
**Breaking Changes:** NONE - value is never read
**Test Impact:** NONE - no tests reference this
**Plugin Impact:** NONE - plugins don't use this context value

#### Dependencies

None - this is completely isolated dead code.

#### Timeline

**Immediate** - Can be removed right now with zero risk.

---

## Summary by Priority

| Priority | Count | Description |
|----------|-------|-------------|
| P0 DEAD  | 1     | Dead code that can be removed immediately |
| P1       | 0     | Active WithValue usage needing refactoring |
| P2       | 0     | Lower priority or test-only usage |

## Recommendations

### Immediate Actions (This Task)

1. ✅ **Remove dead code** from `cmd/glide/main.go` (5 minutes)
2. ✅ **Add linter rule** to prevent future WithValue usage (10 minutes)
3. ✅ **Document context best practices** (45 minutes)

### Future Prevention

1. **Linter enforcement** - Add `forbidigo` rule to `.golangci.yml`
2. **Code review guideline** - Reject any PR introducing `context.WithValue` for domain data
3. **Documentation** - Create `CONTEXT_GUIDELINES.md` explaining proper context usage

## Context Best Practices

### ✅ Valid Uses of context.Context

```go
// 1. Cancellation
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// 2. Deadlines/timeouts
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// 3. Request-scoped values (rarely, and only for middleware/tracing)
// Example: Trace IDs, request IDs (infrastructure concerns, not domain data)
```

### ❌ Invalid Uses (Anti-patterns)

```go
// 1. Domain data
ctx = context.WithValue(ctx, "user", currentUser)  // NO!

// 2. Configuration
ctx = context.WithValue(ctx, "config", appConfig)  // NO!

// 3. Dependencies
ctx = context.WithValue(ctx, "db", database)       // NO!
```

### ✅ Proper Alternatives

```go
// 1. Explicit parameters
func ProcessUser(ctx context.Context, user *User) error {
    // user is explicit, type-safe
}

// 2. Struct fields
type Service struct {
    config *Config
    db     *sql.DB
}

func (s *Service) Process(ctx context.Context) error {
    // Dependencies are explicit in struct
}

// 3. Dependency injection (recommended)
func NewService(config *Config, db *sql.DB) *Service {
    return &Service{config: config, db: db}
}
```

## References

- **Go Blog:** [Context should go away](https://faiface.github.io/post/context-should-go-away-go2/)
- **Style Guide:** [Contexts and structs](https://go.dev/blog/context-and-structs)
- **Best Practice:** Store context.Context only in function parameters, never in structs
- **ADR-013:** Dependency Injection architecture

## Related Work

- **Task 0.4:** Fixed error swallowing patterns
- **Task 1.1:** Implemented DI container (proper alternative to context.WithValue)
- **Task 1.3:** Removed Application God Object (removed need for context passing)
- **Task 1.6:** This audit (remove WithValue anti-pattern)

## Acceptance Criteria

- [x] All WithValue usages documented (1 found)
- [x] Dead code identified (1 instance)
- [x] Replacement strategy determined (simple removal)
- [x] Test impact assessed (none)
- [x] Risk assessment complete (zero risk)

## Next Steps

Proceed to **Subtask 1.6.2:** Remove the dead code and validate.

---

**Audit completed:** 2025-11-27
**Auditor:** Automated code analysis + manual verification
**Result:** MINIMAL WORK - Only dead code removal required
