# Context Usage Guidelines

**Purpose:** Define proper usage of `context.Context` in the Glide codebase
**Status:** ACTIVE
**Last Updated:** 2025-11-27
**Related:** ADR-013 (Dependency Injection), Task 1.6 (Remove WithValue)

---

## Overview

Go's `context.Context` is a powerful tool for managing request-scoped values, cancellation signals, and deadlines. However, it's frequently misused as a "grab bag" for passing dependencies, configuration, or domain data. This document establishes clear guidelines for proper context usage in Glide.

---

## Core Principles

### The Golden Rule

> **Context is for request-scoped metadata, NOT for dependencies or domain data.**

If you're tempted to use `context.WithValue` to pass application data, **stop** and use one of these instead:
- Explicit function parameters
- Struct fields
- Dependency injection (via `pkg/container`)

---

## ✅ Valid Uses of Context

### 1. Cancellation

**Purpose:** Signal that an operation should stop.

```go
func ProcessData(ctx context.Context, data []byte) error {
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()

    // Start background worker
    errCh := make(chan error, 1)
    go func() {
        errCh <- worker(ctx, data)
    }()

    select {
    case <-ctx.Done():
        return ctx.Err() // Cancelled
    case err := <-errCh:
        return err
    }
}
```

**When to use:**
- Long-running operations that should be interruptible
- Coordinating multiple goroutines
- Propagating cancellation through call chains

---

### 2. Deadlines and Timeouts

**Purpose:** Enforce time limits on operations.

```go
func FetchWithTimeout(ctx context.Context, url string) ([]byte, error) {
    // 5-second timeout for this operation
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    return io.ReadAll(resp.Body)
}
```

**When to use:**
- Network requests
- Database queries
- Any operation that might hang indefinitely

---

### 3. Request-Scoped Values (Rare and Specific)

**Purpose:** Pass cross-cutting concerns through middleware layers.

**⚠️ WARNING:** This is the most misused feature. Use ONLY for:
- **Trace IDs** (distributed tracing)
- **Request IDs** (logging correlation)
- **Authentication tokens** (already on the request)
- **Deadline/timeout metadata** (already on context)

```go
// ACCEPTABLE: Infrastructure-level metadata
type requestIDKey struct{}

func WithRequestID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, requestIDKey{}, id)
}

func GetRequestID(ctx context.Context) string {
    if id, ok := ctx.Value(requestIDKey{}).(string); ok {
        return id
    }
    return ""
}

// Used in HTTP middleware
func RequestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := WithRequestID(r.Context(), generateID())
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

**Strict criteria for context.WithValue:**
1. ✅ Infrastructure concern (logging, tracing, auth)
2. ✅ Already on the request (not created by your code)
3. ✅ Optional for functionality (code works without it)
4. ✅ Used across layer boundaries (middleware → handler)
5. ✅ NOT business logic data
6. ✅ NOT configuration
7. ✅ NOT dependencies (services, repositories, etc.)

If you can't check ALL boxes, **don't use `context.WithValue`**.

---

## ❌ Invalid Uses (Anti-patterns)

### ❌ 1. Passing Domain Data

**DON'T:**
```go
// WRONG: User is domain data, not request metadata
ctx = context.WithValue(ctx, "user", currentUser)

func ProcessOrder(ctx context.Context, orderID string) error {
    user := ctx.Value("user").(*User) // Type assertion can panic!
    // ...
}
```

**DO:**
```go
// CORRECT: Explicit parameter
func ProcessOrder(ctx context.Context, user *User, orderID string) error {
    // user is type-safe and visible in signature
}
```

---

### ❌ 2. Passing Configuration

**DON'T:**
```go
// WRONG: Config is application-wide, not request-scoped
ctx = context.WithValue(ctx, "config", appConfig)

func HandleRequest(ctx context.Context) error {
    cfg := ctx.Value("config").(*Config)
    // ...
}
```

**DO:**
```go
// CORRECT: Struct field via dependency injection
type Handler struct {
    config *Config
}

func (h *Handler) Handle(ctx context.Context) error {
    // Use h.config directly
}

// Or: Explicit parameter
func HandleRequest(ctx context.Context, cfg *Config) error {
    // ...
}
```

---

### ❌ 3. Passing Dependencies

**DON'T:**
```go
// WRONG: Database is a dependency, not request metadata
ctx = context.WithValue(ctx, "db", database)

func SaveUser(ctx context.Context, user *User) error {
    db := ctx.Value("db").(*sql.DB)
    // ...
}
```

**DO:**
```go
// CORRECT: Dependency injection
type UserService struct {
    db *sql.DB
}

func (s *UserService) SaveUser(ctx context.Context, user *User) error {
    // Use s.db, pass ctx for cancellation
    return s.db.QueryRowContext(ctx, "INSERT INTO users...").Scan(...)
}
```

---

### ❌ 4. Optional Parameters

**DON'T:**
```go
// WRONG: Optional behavior via context
ctx = context.WithValue(ctx, "verbose", true)

func Process(ctx context.Context) {
    if verbose, ok := ctx.Value("verbose").(bool); ok && verbose {
        log.Println("Processing...")
    }
}
```

**DO:**
```go
// CORRECT: Explicit parameter
func Process(ctx context.Context, verbose bool) {
    if verbose {
        log.Println("Processing...")
    }
}
```

---

### ❌ 5. Avoiding Import Cycles

**DON'T:**
```go
// WRONG: Using context to pass data and avoid import cycles
// (This was the rationale for the dead code we removed)
ctx = context.WithValue(ctx, "project_context", projectCtx)
```

**DO:**
```go
// CORRECT: Fix the import cycle with proper architecture
// - Use interfaces
// - Split packages
// - Use dependency injection
```

---

## Proper Alternatives

### Alternative 1: Explicit Function Parameters

**Best for:** Single function calls, simple data flow

```go
// Clear, type-safe, no magic
func ProcessUser(ctx context.Context, user *User, config *Config) error {
    // All dependencies visible in signature
}
```

**Benefits:**
- ✅ Type-safe
- ✅ Self-documenting
- ✅ No hidden dependencies
- ✅ Easy to test

---

### Alternative 2: Struct Fields

**Best for:** Methods with multiple shared dependencies

```go
type UserService struct {
    db     *sql.DB
    config *Config
    logger *slog.Logger
}

func NewUserService(db *sql.DB, cfg *Config, logger *slog.Logger) *UserService {
    return &UserService{
        db:     db,
        config: cfg,
        logger: logger,
    }
}

func (s *UserService) CreateUser(ctx context.Context, user *User) error {
    // Access s.db, s.config, s.logger directly
    // Pass ctx for cancellation/timeout only
    return s.db.QueryRowContext(ctx, "INSERT INTO users...").Scan(...)
}
```

**Benefits:**
- ✅ Dependencies explicit at construction time
- ✅ Reusable across multiple methods
- ✅ Clear ownership
- ✅ Easy to mock for testing

---

### Alternative 3: Dependency Injection Container

**Best for:** Application-wide dependencies, complex dependency graphs

```go
// Use pkg/container for automatic wiring
import "github.com/ivannovak/glide/v2/pkg/container"

func main() {
    c, err := container.New()
    if err != nil {
        log.Fatal(err)
    }

    // Dependencies automatically injected
    var service *UserService
    if err := c.Run(context.Background(), func(s *UserService) error {
        service = s
        return service.Start(context.Background())
    }); err != nil {
        log.Fatal(err)
    }
}
```

**Benefits:**
- ✅ Centralized dependency management
- ✅ Automatic wiring
- ✅ Lifecycle management
- ✅ Easy to swap implementations (testing, environments)

See `docs/adr/ADR-013-dependency-injection.md` for details.

---

## Context Best Practices

### 1. Context as First Parameter

```go
// CORRECT
func DoWork(ctx context.Context, data string) error

// WRONG
func DoWork(data string, ctx context.Context) error
```

---

### 2. Never Store Context in Structs

```go
// WRONG
type Worker struct {
    ctx context.Context // Don't do this!
}

// CORRECT
type Worker struct {
    config *Config
}

func (w *Worker) DoWork(ctx context.Context) error {
    // Pass ctx as parameter
}
```

**Rationale:** Context lifecycle is tied to requests/operations, not objects.

---

### 3. Don't Pass Nil Context

```go
// WRONG
DoWork(nil, data)

// CORRECT
DoWork(context.Background(), data)
DoWork(context.TODO(), data) // If context will be added later
```

---

### 4. Propagate Context Through Call Chain

```go
func HandleRequest(ctx context.Context) error {
    // Pass ctx down, don't create new root contexts
    return processData(ctx, data)
}

func processData(ctx context.Context, data []byte) error {
    // Continue passing ctx
    return saveData(ctx, data)
}

func saveData(ctx context.Context, data []byte) error {
    // Use ctx for cancellation
    return db.ExecContext(ctx, "INSERT...")
}
```

---

### 5. Check for Cancellation in Loops

```go
func ProcessItems(ctx context.Context, items []Item) error {
    for _, item := range items {
        // Check for cancellation
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }

        if err := processItem(ctx, item); err != nil {
            return err
        }
    }
    return nil
}
```

---

## Linting Enforcement

We use `forbidigo` to prevent `context.WithValue` misuse:

```yaml
# .golangci.yml
linters-settings:
  forbidigo:
    forbid:
      - p: 'context\.WithValue'
        msg: 'context.WithValue is an anti-pattern. Use explicit parameters, struct fields, or dependency injection. See docs/development/CONTEXT_GUIDELINES.md'
      - p: 'ctx\.Value\('
        msg: 'ctx.Value is an anti-pattern. Use explicit parameters, struct fields, or dependency injection. See docs/development/CONTEXT_GUIDELINES.md'
```

**Exceptions:** If you have a legitimate use case (see "Valid Uses" section), add a `//nolint:forbidigo` comment with justification:

```go
//nolint:forbidigo // Legitimate use: trace ID for distributed tracing
ctx = context.WithValue(ctx, traceIDKey{}, generateTraceID())
```

---

## Testing Patterns

### Testing Functions with Context

```go
func TestProcessUser(t *testing.T) {
    ctx := context.Background()

    // Test normal operation
    err := ProcessUser(ctx, &User{Name: "Alice"})
    assert.NoError(t, err)
}

func TestProcessUser_Cancellation(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    cancel() // Cancel immediately

    // Should respect cancellation
    err := ProcessUser(ctx, &User{Name: "Bob"})
    assert.ErrorIs(t, err, context.Canceled)
}

func TestProcessUser_Timeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
    defer cancel()

    time.Sleep(10 * time.Millisecond) // Ensure timeout

    err := ProcessUser(ctx, &User{Name: "Charlie"})
    assert.ErrorIs(t, err, context.DeadlineExceeded)
}
```

---

## Decision Tree

```
Need to pass data to a function?
│
├─ Is it cancellation signal? ──────────> Use context.WithCancel()
├─ Is it a deadline/timeout? ───────────> Use context.WithTimeout()
├─ Is it trace/request ID metadata? ────> MAYBE context.WithValue() (see criteria)
│
└─ Is it domain data/config/dependency?
   │
   ├─ Single function call? ────────────> Explicit parameter
   ├─ Multiple related methods? ────────> Struct field
   └─ Application-wide dependency? ─────> Dependency injection (pkg/container)
```

---

## Common Mistakes

### Mistake 1: "Context makes it optional"

❌ **Wrong thinking:** "If I use context.WithValue, the function can work without it"

✅ **Correct:** Use an explicit optional parameter or functional options:

```go
// Option 1: Explicit optional
func Process(ctx context.Context, required string, optional *OptionalData) error

// Option 2: Functional options
func Process(ctx context.Context, required string, opts ...Option) error
```

---

### Mistake 2: "It avoids changing function signatures"

❌ **Wrong thinking:** "Adding context.WithValue means I don't have to update signatures"

✅ **Correct:** Changing signatures is GOOD - it makes dependencies explicit:

```go
// Before (hidden dependency)
func Process(ctx context.Context) error {
    db := ctx.Value("db").(*sql.DB) // Hidden!
}

// After (explicit dependency)
func Process(ctx context.Context, db *sql.DB) error {
    // Clear what's needed
}
```

---

### Mistake 3: "Import cycles force me to use context"

❌ **Wrong thinking:** "I have an import cycle, so I'll use context to pass data"

✅ **Correct:** Fix the import cycle with proper architecture:

```go
// Problem: Package A imports B, B imports A

// Solution 1: Extract interface to third package
// pkg/interfaces/repository.go
type Repository interface {
    Save(data Data) error
}

// Solution 2: Use dependency injection
// Container handles wiring without import cycles

// Solution 3: Restructure packages
// Split packages by layer (domain, infrastructure, etc.)
```

---

## Migration Guide

If you find existing code using `context.WithValue` for domain data:

### Step 1: Identify what's being passed

```go
// Find the WithValue call
ctx = context.WithValue(ctx, "user", user)

// Find the retrieval
user := ctx.Value("user").(*User)
```

### Step 2: Determine the proper alternative

- **Single function?** → Add parameter
- **Multiple methods?** → Add struct field
- **Application-wide?** → Use DI container

### Step 3: Refactor

```go
// Before
func ProcessOrder(ctx context.Context, orderID string) error {
    user := ctx.Value("user").(*User)
    return saveOrder(ctx, orderID, user)
}

// After
func ProcessOrder(ctx context.Context, user *User, orderID string) error {
    return saveOrder(ctx, user, orderID)
}
```

### Step 4: Update all callers

```go
// Before
ProcessOrder(ctx, orderID)

// After
ProcessOrder(ctx, currentUser, orderID)
```

---

## References

### Go Blog Posts
- [Go Concurrency Patterns: Context](https://go.dev/blog/context)
- [Contexts and structs](https://go.dev/blog/context-and-structs)

### Community Guidelines
- [Context should go away for Go 2](https://faiface.github.io/post/context-should-go-away-go2/) (critique)
- [Effective Go: Context](https://go.dev/doc/effective_go#context)

### Internal Documentation
- `docs/adr/ADR-013-dependency-injection.md` - DI architecture
- `docs/technical-debt/CONTEXT_WITHVALUE_AUDIT.md` - Task 1.6 audit
- `pkg/container/` - Dependency injection container

---

## Summary

### ✅ DO
- Use context for cancellation and timeouts
- Pass context as first parameter
- Propagate context through call chains
- Use explicit parameters for domain data
- Use struct fields for method dependencies
- Use DI container for application-wide dependencies

### ❌ DON'T
- Use context.WithValue for domain data
- Use context.WithValue for configuration
- Use context.WithValue for dependencies
- Store context in structs
- Pass nil context (use context.Background())
- Use context to avoid changing signatures

---

**Remember:** If you're unsure whether to use context.WithValue, **the answer is probably no**. Use explicit parameters or dependency injection instead.

**Questions?** See `docs/adr/ADR-013-dependency-injection.md` or ask in code review.
