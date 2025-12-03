# Performance Guide

This guide covers performance optimization and monitoring in Glide.

## Table of Contents
- [Overview](#overview)
- [Performance Budgets](#performance-budgets)
- [Critical Path Optimization](#critical-path-optimization)
- [Metrics Collection](#metrics-collection)
- [Benchmarking](#benchmarking)
- [Plugin Performance](#plugin-performance)
- [Best Practices](#best-practices)

## Overview

Glide is designed for fast startup and responsive execution. Key performance targets:

| Operation | Target | Priority |
|-----------|--------|----------|
| Startup (cold) | < 200ms | Critical |
| Context Detection | < 100ms | Critical |
| Config Load | < 50ms | Critical |
| Plugin Discovery | < 500ms | Critical |
| Command Execution | < 100ms overhead | High |

## Performance Budgets

### Using Performance Budgets

```go
import "github.com/glide-cli/glide/v3/pkg/performance"

// List all budgets
budgets := performance.ListBudgets()
for _, budget := range budgets {
    fmt.Printf("%s: max %v (%s)\n",
        budget.Name, budget.MaxDuration, budget.Priority)
}

// Get specific budget
budget, ok := performance.GetBudget("context_detection")
if ok && elapsed > budget.MaxDuration {
    log.Warn("Budget exceeded",
        "operation", budget.Name,
        "actual", elapsed,
        "budget", budget.MaxDuration,
    )
}
```

### Standard Budgets

| Budget Name | Max Duration | Priority | Description |
|-------------|--------------|----------|-------------|
| `startup_total` | 200ms | Critical | Total CLI startup time |
| `context_detection` | 100ms | Critical | Project context detection |
| `config_load` | 50ms | Critical | Configuration loading |
| `plugin_discovery` | 500ms | Critical | Plugin discovery (lazy) |
| `plugin_load` | 1s | High | Single plugin load |
| `command_overhead` | 100ms | High | Command dispatch overhead |

### Custom Budgets

Register application-specific budgets:

```go
performance.RegisterBudget(performance.Budget{
    Name:        "api_call",
    MaxDuration: 2 * time.Second,
    Priority:    "high",
    Description: "Maximum API response time",
})
```

## Critical Path Optimization

### Startup Optimization

Glide uses lazy initialization for fast startup:

```go
// Fast detector skips expensive checks
detector, err := context.NewDetectorFast()
// Docker status checked lazily on first use

// Lazy plugin discovery
manager := sdk.NewManager(config)
manager.DiscoverPluginsLazy() // ~47μs vs ~1.35s eager loading
```

### Plugin Loading

Plugins are loaded on-demand:

```go
// Discovery is fast - no process spawning
manager.DiscoverPluginsLazy()

// Loading happens when plugin is needed
plugin, err := manager.GetPlugin("docker-compose")
// First access loads the plugin (~100ms)
// Subsequent access is cached (~1μs)
```

### Context Detection

Optimized context detection:

```go
// Standard detector (full checks)
detector, _ := context.NewDetector()
ctx, _ := detector.Detect() // ~72ms

// Fast detector (lazy Docker check)
detector, _ := context.NewDetectorFast()
ctx, _ := detector.Detect() // ~19μs

// Check Docker when actually needed
detector.EnsureDockerStatus(ctx)
```

## Metrics Collection

### Recording Metrics

```go
import "github.com/glide-cli/glide/v3/pkg/observability"

// Counters
observability.IncrementCounter("commands_executed")
observability.IncrementCounterBy("bytes_processed", 1024)

// Gauges
observability.SetGauge("active_plugins", 5)

// Timings
timer := observability.StartTimer("api_call")
result := callAPI()
timer.Stop() // Automatically records duration
```

### Timer Usage

```go
// Simple timing
timer := observability.StartTimer("operation")
doWork()
duration := timer.Stop()

// Manual recording
start := time.Now()
doWork()
observability.RecordTiming("operation", time.Since(start))

// With metrics collector
mc := observability.NewMetricsCollector()
timer := observability.StartTimerWithCollector("operation", mc)
doWork()
timer.Stop()

stats := mc.GetTimingStats("operation")
fmt.Printf("Avg: %v, P95: %v\n", stats.Avg, stats.P95)
```

### Histograms

```go
mc := observability.NewMetricsCollector()
hist := mc.CreateHistogram("request_size",
    []float64{100, 500, 1000, 5000, 10000})

// Record values
hist.Observe(256)
hist.Observe(1024)
hist.Observe(8192)

// Get summary
sum, count, buckets := hist.Summary()
```

## Benchmarking

### Writing Benchmarks

```go
func BenchmarkContextDetection(b *testing.B) {
    detector, _ := context.NewDetectorFast()
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        detector.Detect()
    }
}

func BenchmarkPluginDiscovery(b *testing.B) {
    config := sdk.DefaultConfig()
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        manager := sdk.NewManager(config)
        manager.DiscoverPluginsLazy()
    }
}
```

### Running Benchmarks

```bash
# Run all benchmarks
go test -bench=. ./...

# Run specific benchmark
go test -bench=BenchmarkContextDetection ./internal/context/...

# With memory stats
go test -bench=. -benchmem ./...

# Save baseline
go test -bench=. ./... > baseline.txt

# Compare with baseline
go test -bench=. ./... | benchstat baseline.txt -
```

### Benchmark Assertions

```go
func BenchmarkStartup(b *testing.B) {
    budget := performance.MustGetBudget("startup_total")

    var totalDuration time.Duration
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        start := time.Now()
        app := NewApplication()
        totalDuration += time.Since(start)
    }

    avgDuration := totalDuration / time.Duration(b.N)
    if avgDuration > budget.MaxDuration {
        b.Errorf("Startup budget exceeded: %v > %v",
            avgDuration, budget.MaxDuration)
    }
}
```

## Plugin Performance

### Efficient Plugin Design

```go
type MyPlugin struct {
    v2.BasePlugin[MyConfig]
    // Cache expensive resources
    cachedClient *http.Client
    cacheOnce    sync.Once
}

func (p *MyPlugin) getClient() *http.Client {
    p.cacheOnce.Do(func() {
        p.cachedClient = &http.Client{
            Timeout: 30 * time.Second,
        }
    })
    return p.cachedClient
}

// Lazy initialization
func (p *MyPlugin) Init(ctx context.Context) error {
    // Don't do expensive work here
    // Defer to first command execution
    return nil
}
```

### Avoiding Performance Pitfalls

```go
// Bad: Expensive initialization on every command
func (p *MyPlugin) Execute(ctx context.Context, args []string) error {
    client := initializeExpensiveClient() // 500ms each time
    return client.DoWork()
}

// Good: Initialize once, reuse
func (p *MyPlugin) Execute(ctx context.Context, args []string) error {
    client := p.getClient() // ~1μs after first call
    return client.DoWork()
}
```

### Plugin Load Time

Keep plugin load time under 100ms:

```go
func (p *MyPlugin) Init(ctx context.Context) error {
    // Good: Quick checks
    if p.Config().APIKey == "" {
        return errors.NewConfigError("API key required")
    }

    // Bad: Expensive network calls
    // Don't do this in Init:
    // resp, _ := http.Get("https://api.example.com/validate")

    return nil
}

func (p *MyPlugin) Start(ctx context.Context) error {
    // OK: Expensive operations here (called when needed)
    return p.connectToService()
}
```

## Best Practices

### 1. Measure Before Optimizing

```go
// Profile first
timer := observability.StartTimer("suspected_slow_operation")
result := doOperation()
elapsed := timer.Stop()

log.Info("Operation timing", "duration", elapsed)
// Only optimize if actually slow
```

### 2. Use Lazy Initialization

```go
type Service struct {
    client     *Client
    clientOnce sync.Once
}

func (s *Service) getClient() *Client {
    s.clientOnce.Do(func() {
        s.client = NewExpensiveClient()
    })
    return s.client
}
```

### 3. Cache Expensive Results

```go
type Detector struct {
    mu          sync.RWMutex
    cachedRoot  string
    cacheValid  bool
}

func (d *Detector) ProjectRoot() string {
    d.mu.RLock()
    if d.cacheValid {
        defer d.mu.RUnlock()
        return d.cachedRoot
    }
    d.mu.RUnlock()

    d.mu.Lock()
    defer d.mu.Unlock()
    // Double-check after acquiring write lock
    if d.cacheValid {
        return d.cachedRoot
    }

    d.cachedRoot = d.findProjectRoot()
    d.cacheValid = true
    return d.cachedRoot
}
```

### 4. Parallel Operations

```go
func (m *Manager) loadPlugins(plugins []string) error {
    var wg sync.WaitGroup
    errChan := make(chan error, len(plugins))

    for _, name := range plugins {
        wg.Add(1)
        go func(n string) {
            defer wg.Done()
            if err := m.loadPlugin(n); err != nil {
                errChan <- err
            }
        }(name)
    }

    wg.Wait()
    close(errChan)

    // Collect errors
    var errs []error
    for err := range errChan {
        errs = append(errs, err)
    }
    return errors.Join(errs...)
}
```

### 5. Avoid Unnecessary Work

```go
// Bad: Always check Docker
func (d *Detector) Detect() (*Context, error) {
    ctx := &Context{}
    ctx.DockerAvailable = d.checkDocker() // 50ms
    return ctx, nil
}

// Good: Check Docker only when needed
func (d *Detector) Detect() (*Context, error) {
    ctx := &Context{}
    // Defer Docker check
    ctx.Extensions["_dockerCheckDeferred"] = true
    return ctx, nil
}

func (d *Detector) EnsureDockerStatus(ctx *Context) {
    if _, ok := ctx.Extensions["_dockerCheckDeferred"]; ok {
        ctx.DockerAvailable = d.checkDocker()
        delete(ctx.Extensions, "_dockerCheckDeferred")
    }
}
```

## Performance Monitoring

### Health Checks

```go
import "github.com/glide-cli/glide/v3/pkg/observability"

monitor := observability.NewHealthMonitor("1.0.0")
report := monitor.Check(ctx)

fmt.Printf("Status: %s\n", report.Status)
fmt.Printf("Uptime: %v\n", report.Uptime)

for name, budget := range report.Performance.BudgetStatus {
    if !budget.Passes {
        fmt.Printf("WARNING: %s exceeded budget: %v > %v\n",
            name, budget.Actual, budget.Target)
    }
}
```

### Runtime Statistics

```go
report := monitor.Check(ctx)
runtime := report.Runtime

fmt.Printf("Go Version: %s\n", runtime.GoVersion)
fmt.Printf("Goroutines: %d\n", runtime.NumGoroutine)
fmt.Printf("Heap Alloc: %d MB\n", runtime.HeapAlloc/1024/1024)
fmt.Printf("GC Pauses: %d\n", runtime.NumGC)
```

## See Also

- [Performance Documentation](../development/PERFORMANCE.md)
- [Benchmarks](../performance/BENCHMARKS.md)
- [pkg/observability](../../pkg/observability/doc.go)
- [pkg/performance](../../pkg/performance/doc.go)
