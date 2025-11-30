# Performance Guide

This document describes Glide's performance requirements, how to measure performance, and guidelines for maintaining performance standards.

## Performance Budgets

Glide defines explicit performance budgets for critical operations. These budgets are codified in `pkg/performance/budgets.go` and enforced through benchmarks and CI.

### Critical Operations (P0)

| Operation | Target | Description |
|-----------|--------|-------------|
| Context detection | <100ms | Detect project context (git root, frameworks, worktree mode) |
| Plugin discovery | <500ms | Discover and enumerate all available plugins |
| Startup total | <300ms | Total time from start to ready state (excluding plugins) |

### High Priority Operations (P1)

| Operation | Target | Description |
|-----------|--------|-------------|
| Config load | <50ms | Load a single configuration file |
| Config merge (single) | <30ms | Merge a single configuration file |
| Config merge (multiple) | <100ms | Merge multiple (5+) configuration files |
| Plugin load | <200ms | Load and initialize a single plugin |
| Command lookup | <1ms | Look up a command by name |
| Path validation | <50μs | Validate a file path for security |

### Standard Operations (P2)

| Operation | Target | Description |
|-----------|--------|-------------|
| Plugin cache get | <10μs | Retrieve a plugin from cache |
| Error creation | <1μs | Create a structured error |
| Error wrap | <500ns | Wrap an existing error |
| Registry get | <100ns | Retrieve an item from registry |
| Registry list | <10μs | List all items in registry (100 items) |

## Running Benchmarks

### Basic Benchmark Run

```bash
# Run all benchmarks
go test -bench=. -benchmem ./tests/benchmarks/...

# Run with more iterations for accuracy
go test -bench=. -benchmem -benchtime=3s -count=5 ./tests/benchmarks/...

# Run specific benchmark
go test -bench=BenchmarkContextDetection -benchmem ./tests/benchmarks/...
```

### Comparing Against Baseline

```bash
# Using the comparison script
./scripts/benchmark-compare.sh

# Or manually with benchstat
go test -bench=. -benchmem -benchtime=1s -count=5 ./tests/benchmarks/... 2>&1 | \
  grep -E "^Bench.*ns/op" > new.txt
benchstat benchmarks.txt new.txt
```

### Updating the Baseline

After intentional performance changes, update the baseline:

```bash
go test -bench=. -benchmem -benchtime=1s -count=5 ./tests/benchmarks/... 2>&1 | \
  grep -E "^Bench.*ns/op" > benchmarks.txt
```

## Profiling

### CPU Profiling

```bash
# Generate CPU profile
go test -bench=BenchmarkContextDetection -cpuprofile=cpu.prof ./tests/benchmarks/...

# Analyze profile
go tool pprof cpu.prof

# Interactive web interface
go tool pprof -http=:8080 cpu.prof
```

### Memory Profiling

```bash
# Generate memory profile
go test -bench=BenchmarkContextDetection -memprofile=mem.prof ./tests/benchmarks/...

# Analyze allocations
go tool pprof -alloc_space mem.prof

# Analyze bytes
go tool pprof -alloc_objects mem.prof
```

### Trace Profiling

```bash
# Generate trace
go test -bench=BenchmarkContextDetection -trace=trace.out ./tests/benchmarks/...

# Analyze trace
go tool trace trace.out
```

## Using Performance Budgets in Code

```go
import (
    "time"
    "github.com/ivannovak/glide/v2/pkg/performance"
)

func someOperation() {
    start := time.Now()

    // ... do work ...

    duration := time.Since(start)

    // Check against budget
    result := performance.Measure("context_detection", duration, allocations, bytes)
    if !result.Passes {
        log.Warn("Operation exceeded performance budget",
            "operation", result.Operation,
            "duration", result.Duration,
            "passesDuration", result.PassesDuration,
        )
    }
}
```

## CI Integration

Benchmarks run in CI on:
- Manual trigger (workflow_dispatch)
- PRs labeled with "benchmark"
- Nightly scheduled runs

To trigger benchmarks on a PR, add the `benchmark` label.

### Regression Detection

CI will fail if any benchmark shows >15% regression compared to the baseline. Minor regressions (<15%) are noted but won't fail the build.

## Common Performance Issues

### High Allocation Count

**Symptom:** High `allocs/op` in benchmark output

**Causes:**
- Creating slices without pre-allocation
- String concatenation in loops
- Interface boxing

**Solutions:**
```go
// Bad: Grows slice repeatedly
var result []string
for _, item := range items {
    result = append(result, item.Name)
}

// Good: Pre-allocate
result := make([]string, 0, len(items))
for _, item := range items {
    result = append(result, item.Name)
}

// Bad: String concatenation
s := ""
for _, part := range parts {
    s += part
}

// Good: Use strings.Builder
var b strings.Builder
for _, part := range parts {
    b.WriteString(part)
}
s := b.String()
```

### Slow File Operations

**Symptom:** High latency in file-related benchmarks

**Solutions:**
- Use `filepath.WalkDir` instead of `filepath.Walk`
- Cache file system results where appropriate
- Use `os.Stat` instead of `os.Lstat` when symlinks don't matter
- Consider concurrent file operations with worker pools

### Memory Leaks

**Symptom:** Growing memory over time

**Detection:**
```bash
# Run with memory sanitizer
GODEBUG=gctrace=1 go test -bench=. ./tests/benchmarks/...

# Use pprof heap profile
go tool pprof -http=:8080 mem.prof
```

**Common causes:**
- Goroutine leaks
- Unclosed resources (files, connections)
- Circular references preventing GC

## Writing Performant Code

### Guidelines

1. **Measure first**: Always benchmark before optimizing
2. **Profile to find hotspots**: Don't guess, use pprof
3. **Pre-allocate slices**: When size is known
4. **Avoid reflection**: Use type assertions or generics
5. **Minimize allocations**: Reuse buffers, use sync.Pool
6. **Consider concurrency**: But measure to ensure benefit

### Anti-patterns to Avoid

```go
// Anti-pattern: Reflection in hot path
reflect.ValueOf(x).FieldByName("Name")

// Anti-pattern: Unnecessary interface{}
func process(data interface{}) // Use generics instead

// Anti-pattern: Creating closures in loops
for _, item := range items {
    go func() {
        process(item) // item captured by reference
    }()
}
```

## Maintaining Performance

1. **Run benchmarks before PRs**: Use `./scripts/benchmark-compare.sh`
2. **Review benchmark results in CI**: Check the benchmark report on PRs
3. **Update baselines intentionally**: When performance improves
4. **Document regressions**: If a regression is accepted, explain why
5. **Monitor production**: If metrics are enabled, watch for degradation
