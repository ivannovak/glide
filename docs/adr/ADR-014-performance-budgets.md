# ADR-014: Performance Budgets

## Status
Accepted

## Date
2025-11-29

## Context

As Glide grew in complexity with plugins, configuration merging, and context detection, startup time and command responsiveness degraded. Without explicit performance targets, regressions went unnoticed until they significantly impacted user experience.

Key observations:
- Plugin discovery could take over 1 second with many plugins
- Context detection with Docker checks added ~70ms overhead
- Configuration loading added variable latency
- No automated regression detection in CI

Users expect CLI tools to be fast and responsive. Industry benchmarks suggest:
- Startup time under 200ms for interactive tools
- Command overhead under 100ms
- Total execution time competitive with native tools

## Decision

We adopted a formal performance budget system with:

1. **Explicit Performance Targets**: Named budgets for critical operations
2. **Automated Monitoring**: Budget checking integrated with metrics
3. **CI Enforcement**: Benchmarks run on every PR
4. **Documentation**: Clear targets for contributors

### Performance Budgets

| Operation | Budget | Priority | Rationale |
|-----------|--------|----------|-----------|
| `startup_total` | 200ms | Critical | User-perceptible delay threshold |
| `context_detection` | 100ms | Critical | Runs on every command |
| `config_load` | 50ms | Critical | Runs on every command |
| `plugin_discovery` | 500ms | Critical | Can be cached/lazy |
| `plugin_load` | 1000ms | High | Single plugin, can warm cache |
| `command_overhead` | 100ms | High | Excluding command execution |

### Implementation

```go
// pkg/performance/budgets.go
type Budget struct {
    Name        string
    MaxDuration time.Duration
    Priority    string        // "critical", "high", "medium"
    Description string
}

var budgets = []Budget{
    {
        Name:        "startup_total",
        MaxDuration: 200 * time.Millisecond,
        Priority:    "critical",
        Description: "Total CLI startup time including all initialization",
    },
    // ... other budgets
}
```

## Consequences

### Positive

1. **Clear Targets**: Contributors know performance expectations
2. **Regression Detection**: CI catches performance regressions
3. **Optimization Focus**: Budget violations highlight problem areas
4. **Documentation**: Budgets serve as performance documentation
5. **Accountability**: Each budget has an owner/priority

### Negative

1. **Maintenance Overhead**: Budgets need updating as features change
2. **False Positives**: CI environment variations may cause flaky failures
3. **Optimization Pressure**: May lead to premature optimization

## Implementation

### Budget Definition

```go
// pkg/performance/budgets.go
func ListBudgets() []Budget {
    return budgets
}

func GetBudget(name string) (Budget, bool) {
    for _, b := range budgets {
        if b.Name == name {
            return b, true
        }
    }
    return Budget{}, false
}
```

### Integration with Observability

```go
// Health checks verify budgets
func (hm *HealthMonitor) checkPerformance() *PerformanceReport {
    for _, budget := range performance.ListBudgets() {
        stats := hm.metrics.GetTimingStats(budget.Name)
        if stats.Count > 0 && stats.Avg > budget.MaxDuration {
            // Budget exceeded - report degraded health
        }
    }
}
```

### CI Benchmarks

```yaml
# .github/workflows/benchmarks.yml
- name: Run benchmarks
  run: |
    go test -bench=. -benchmem ./... > benchmark.txt

- name: Compare with baseline
  run: |
    benchstat baseline.txt benchmark.txt
    # Fail if critical budgets exceeded
```

## Alternatives Considered

### 1. Ad-hoc Performance Testing

Just run benchmarks occasionally without formal budgets.

**Rejected because**: Regressions go unnoticed, no clear targets for optimization.

### 2. External APM Tools

Use Datadog, New Relic, or similar for performance monitoring.

**Rejected because**: Overkill for CLI tool, adds external dependency, not available in CI.

### 3. Compile-time Constraints

Use go:generate or build tags to enforce budgets.

**Rejected because**: Too complex, performance varies by environment.

## Validation

### Benchmark Examples

```go
func BenchmarkStartupTotal(b *testing.B) {
    budget := performance.MustGetBudget("startup_total")

    for i := 0; i < b.N; i++ {
        start := time.Now()
        // Simulate full startup
        elapsed := time.Since(start)

        if elapsed > budget.MaxDuration {
            b.Errorf("Budget exceeded: %v > %v",
                elapsed, budget.MaxDuration)
        }
    }
}
```

### Current Performance

| Operation | Before | After | Budget | Status |
|-----------|--------|-------|--------|--------|
| Context Detection | ~72ms | ~19μs | 100ms | ✅ |
| Plugin Discovery | ~1.35s | ~47μs | 500ms | ✅ |
| Config Load | ~126μs | ~124μs | 50ms | ✅ |
| Startup Total | ~200ms | ~50ms | 200ms | ✅ |

## References

- [Web Vitals - Google](https://web.dev/vitals/)
- [CLI Performance Best Practices](https://clig.dev/#performance)
- [Performance Guide](../guides/performance.md)
- [pkg/performance](../../pkg/performance/doc.go)
