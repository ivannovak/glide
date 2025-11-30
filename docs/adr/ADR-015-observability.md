# ADR-015: Observability Infrastructure

## Status
Accepted

## Date
2025-11-29

## Context

As Glide matured, we needed better visibility into:

1. **Performance**: Which operations are slow? Are we meeting budgets?
2. **Errors**: Where do errors occur? What's the error rate?
3. **Health**: Is the system healthy? Are all components operational?
4. **Debugging**: What happened during a problematic run?

Without proper observability, debugging issues required extensive logging and reproduction. Performance regressions went undetected until users reported them.

### Requirements

- Lightweight (CLI tools must be fast)
- No external dependencies for basic operation
- Structured data for analysis
- Integration with existing logging
- Health endpoint for plugin system

## Decision

We implemented a comprehensive observability package (`pkg/observability`) with:

1. **Metrics Collection**: Counters, gauges, histograms, timings
2. **Performance Logging**: Structured operation logs
3. **Health Monitoring**: Component health checks and status

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Observability Layer                       │
├─────────────────┬──────────────────┬───────────────────────┤
│ MetricsCollector│ PerformanceLogger│    HealthMonitor      │
├─────────────────┼──────────────────┼───────────────────────┤
│ • Counters      │ • JSON logging   │ • Component checks    │
│ • Gauges        │ • Operation tracking │ • Budget validation │
│ • Histograms    │ • Caller info    │ • Runtime stats       │
│ • Timings       │ • Metadata       │ • Plugin health       │
└─────────────────┴──────────────────┴───────────────────────┘
```

### Components

#### MetricsCollector

```go
type MetricsCollector struct {
    counters   map[string]*int64
    gauges     map[string]*float64
    histograms map[string]*HistogramMetric
    timings    map[string][]time.Duration
}
```

#### PerformanceLogger

```go
type PerformanceLog struct {
    Timestamp   time.Time
    Level       LogLevel
    Operation   string
    Duration    time.Duration
    Success     bool
    Error       string
    Labels      map[string]string
    Metadata    map[string]interface{}
}
```

#### HealthMonitor

```go
type HealthMonitor struct {
    startTime  time.Time
    version    string
    checkers   []HealthChecker
    metrics    *MetricsCollector
    budgets    []performance.Budget
}
```

## Consequences

### Positive

1. **Visibility**: Clear view into system performance
2. **Debugging**: Structured logs aid troubleshooting
3. **Regression Detection**: Budget violations caught early
4. **Plugin Monitoring**: Health checks for plugin system
5. **Lightweight**: Minimal overhead in normal operation

### Negative

1. **Memory Overhead**: Metrics stored in memory
2. **Complexity**: Additional package to maintain
3. **Learning Curve**: Developers need to instrument code

## Implementation

### Metrics Collection

```go
// Record a timing
timer := observability.StartTimer("plugin_load")
plugin, err := loader.Load(name)
timer.Stop() // Automatically records duration

// Increment counter
observability.IncrementCounter("commands_executed")

// Set gauge
observability.SetGauge("active_plugins", float64(count))

// Get statistics
stats := mc.GetTimingStats("plugin_load")
fmt.Printf("Avg: %v, P95: %v\n", stats.Avg, stats.P95)
```

### Performance Logging

```go
logger := observability.NewPerformanceLogger()

// Track an operation
tracker := logger.LogOperationStart("api_call", map[string]string{
    "endpoint": "/users",
})
tracker.AddMetadata("user_id", userID)

result, err := api.Call()
tracker.Finish(err) // Logs with duration, success, error
```

### Health Monitoring

```go
monitor := observability.NewHealthMonitor("1.0.0")

// Register checkers
monitor.RegisterChecker(NewPluginHealthChecker("plugins", mgr))
monitor.RegisterChecker(NewConfigHealthChecker("config", path))

// Check health
report := monitor.Check(ctx)
if report.Status != HealthStatusHealthy {
    log.Warn("System degraded")
}

// Access runtime stats
fmt.Printf("Goroutines: %d\n", report.Runtime.NumGoroutine)
fmt.Printf("Heap: %d MB\n", report.Runtime.HeapAlloc/1024/1024)
```

### JSON Output

Performance logs output structured JSON:

```json
{
  "timestamp": "2025-11-29T10:30:45Z",
  "level": "info",
  "operation": "plugin_load",
  "duration_ns": 45000000,
  "duration_ms": 45.0,
  "success": true,
  "labels": {"plugin": "docker"},
  "caller": "pkg/plugin/sdk/manager.go:123"
}
```

## Alternatives Considered

### 1. OpenTelemetry

Full observability stack with tracing, metrics, logging.

**Rejected because**:
- Heavyweight for CLI tool
- Requires external collector for value
- Adds significant dependencies

### 2. Prometheus Client

Standard metrics library for Go.

**Rejected because**:
- Designed for long-running services
- Pull-based metrics don't fit CLI model
- Requires external Prometheus server

### 3. Structured Logging Only

Use structured logging (slog) for all observability.

**Rejected because**:
- No aggregation or statistics
- Hard to compute P95, averages
- No histogram support

### 4. External APM

Use Datadog, New Relic, Honeycomb, etc.

**Rejected because**:
- External dependency
- Network overhead
- Privacy concerns
- Cost

## Thread Safety

All collectors are thread-safe:

```go
type MetricsCollector struct {
    mu sync.RWMutex
    // ...
}

// Atomic counter increments
atomic.AddInt64(counter, 1)

// Mutex for complex operations
func (mc *MetricsCollector) RecordTiming(name string, d time.Duration) {
    mc.mu.Lock()
    defer mc.mu.Unlock()
    mc.timings[name] = append(mc.timings[name], d)
}
```

## Performance Overhead

Measured overhead:
- Counter increment: ~10ns
- Timing record: ~100ns
- Histogram observe: ~50ns
- Snapshot creation: ~1μs

Total overhead for typical command: <1ms

## Future Enhancements

1. **Persistent Metrics**: Store metrics across runs
2. **Export Formats**: Prometheus, StatsD support
3. **Alerting**: Threshold-based alerts
4. **Dashboard**: Built-in metrics visualization

## References

- [OpenTelemetry](https://opentelemetry.io/)
- [Prometheus Go Client](https://github.com/prometheus/client_golang)
- [Structured Logging with slog](https://pkg.go.dev/log/slog)
- [Performance Guide](../guides/performance.md)
- [pkg/observability](../../pkg/observability/doc.go)
