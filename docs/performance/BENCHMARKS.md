# Performance Benchmarks

This document contains performance benchmarks for critical Glide operations and guidelines for running performance regression tests.

## Overview

Performance benchmarks are located in `tests/benchmarks/` and cover three critical areas:
1. **Plugin Operations** - Plugin discovery, loading, and caching
2. **Configuration Management** - Config loading, discovery, and merging
3. **Context Detection** - Project context detection across different scenarios

## Running Benchmarks

### Run All Benchmarks

```bash
go test -bench=. -benchmem ./tests/benchmarks/...
```

### Run Specific Benchmark Suite

```bash
# Plugin benchmarks only
go test -bench=Plugin -benchmem ./tests/benchmarks/plugin_bench_test.go

# Config benchmarks only
go test -bench=Config -benchmem ./tests/benchmarks/config_bench_test.go

# Context benchmarks only
go test -bench=Context -benchmem ./tests/benchmarks/context_bench_test.go
```

### Save Baseline for Comparison

```bash
# Save current benchmarks as baseline
go test -bench=. -benchmem ./tests/benchmarks/... > baseline.txt

# Run benchmarks and compare
go test -bench=. -benchmem ./tests/benchmarks/... > current.txt
benchcmp baseline.txt current.txt
```

## Performance Targets

### Plugin Operations

| Operation | Target Time | Target Allocations |
|-----------|-------------|-------------------|
| Plugin Discovery (10 plugins) | < 10ms | < 5000 allocs |
| Plugin Discovery (100 plugins) | < 100ms | < 50000 allocs |
| Plugin Cache Get | < 1μs | 0 allocs |
| Plugin Cache Put | < 1μs | 0 allocs |
| Plugin List | < 1ms | < 100 allocs |

### Configuration Management

| Operation | Target Time | Target Allocations |
|-----------|-------------|-------------------|
| Config Load (single) | < 30ms | < 200 allocs |
| Config Discovery (5 levels) | < 30ms | < 100 allocs |
| Config Merging (5 files) | < 150ms | < 1000 allocs |
| Config Validation | < 30ms | < 200 allocs |

### Context Detection

| Operation | Target Time | Target Allocations |
|-----------|-------------|-------------------|
| Context Detection (basic) | < 100ms | < 200 allocs |
| Context Detection (multi-framework) | < 200ms | < 300 allocs |
| Context Detection (nested 5 levels) | < 150ms | < 250 allocs |
| Context Validation | < 1μs | 0 allocs |

## Baseline Results

Benchmarks were run on Apple M3 Max (darwin/arm64).

### Plugin Benchmarks

```
BenchmarkPluginDiscovery-16                       	       1	1192052709 ns/op	 1434440 B/op	    5653 allocs/op
BenchmarkPluginDiscoveryEmpty-16                  	   69687	     15848 ns/op	    2112 B/op	      15 allocs/op
BenchmarkPluginDiscoveryLarge-16                  	       1	1152903417 ns/op	 1426824 B/op	    5675 allocs/op
BenchmarkPluginListPlugins-16                     	37372878	        31.58 ns/op	       0 B/op	       0 allocs/op
BenchmarkPluginCachePut-16                        	21283456	        56.44 ns/op	       0 B/op	       0 allocs/op
BenchmarkPluginCacheGet-16                        	29394141	        40.86 ns/op	       0 B/op	       0 allocs/op
BenchmarkPluginCachePutGet-16                     	14001006	        88.35 ns/op	       0 B/op	       0 allocs/op
BenchmarkPluginCleanup-16                         	 6613257	       183.2 ns/op	       0 B/op	       0 allocs/op
```

### Configuration Benchmarks

```
BenchmarkConfigLoad-16                            	   46386	     25967 ns/op	    9665 B/op	     111 allocs/op
BenchmarkConfigDiscoverySingleLevel-16            	  244566	      5213 ns/op	    1376 B/op	      10 allocs/op
BenchmarkConfigDiscoveryNested-16                 	   55406	     21780 ns/op	    6144 B/op	      50 allocs/op
BenchmarkConfigMergingEmpty-16                    	  726447	      1739 ns/op	     722 B/op	       8 allocs/op
BenchmarkConfigMergingSingle-16                   	   45961	     25822 ns/op	    9761 B/op	     111 allocs/op
BenchmarkConfigMergingMultiple-16                 	   10000	    119511 ns/op	   46484 B/op	     523 allocs/op
BenchmarkConfigMergingLarge-16                    	   47678	     26189 ns/op	    9745 B/op	     111 allocs/op
BenchmarkConfigValidation-16                      	   42861	     26140 ns/op	    9713 B/op	     111 allocs/op
```

### Context Detection Benchmarks

```
BenchmarkContextDetection-16                      	      16	  70372143 ns/op	   25315 B/op	     142 allocs/op
BenchmarkContextDetectionWithConfig-16            	      16	  75019898 ns/op	   24434 B/op	     134 allocs/op
BenchmarkContextDetectionMultiFramework-16        	      15	 119722175 ns/op	   25555 B/op	     142 allocs/op
BenchmarkContextDetectionSingleRepo-16            	      18	  74608845 ns/op	   25394 B/op	     142 allocs/op
BenchmarkContextDetectionMultiWorktree-16         	      16	  74153099 ns/op	   25875 B/op	     144 allocs/op
BenchmarkContextValidation-16                     	1000000000	         0.2790 ns/op	       0 B/op	       0 allocs/op
```

## Performance Regression Detection

### CI Integration

Add to `.github/workflows/ci.yml`:

```yaml
- name: Run Performance Benchmarks
  run: |
    go test -bench=. -benchmem ./tests/benchmarks/... > current-bench.txt

- name: Check for Performance Regressions
  run: |
    # Compare with baseline (stored in repository or artifacts)
    # Fail if performance degrades > 20%
```

### Regression Thresholds

Performance regressions are flagged when:
- **Time**: > 20% slower than baseline
- **Memory**: > 20% more allocations than baseline
- **Bytes**: > 20% more bytes allocated than baseline

### Manual Regression Testing

```bash
# 1. Establish baseline before changes
git checkout main
go test -bench=. -benchmem ./tests/benchmarks/... > baseline.txt

# 2. Test with changes
git checkout feature-branch
go test -bench=. -benchmem ./tests/benchmarks/... > feature.txt

# 3. Compare results
benchcmp baseline.txt feature.txt

# 4. Review differences
# Look for:
# - Large time increases (> 20%)
# - Allocation count increases
# - Memory usage increases
```

## Optimization Guidelines

### Plugin Operations

1. **Cache aggressively** - Use the plugin cache for all lookups
2. **Lazy loading** - Only load plugin metadata when needed
3. **Parallel discovery** - Discover plugins from multiple directories in parallel

### Configuration Management

1. **Minimize file I/O** - Cache discovered config paths
2. **Efficient merging** - Merge configs in a single pass
3. **Schema validation** - Validate early to avoid processing invalid configs

### Context Detection

1. **Cache results** - Context rarely changes during execution
2. **Short-circuit checks** - Check for .git before walking filesystem
3. **Limit search depth** - Stop at reasonable project root boundaries

## Profiling

### CPU Profiling

```bash
go test -bench=BenchmarkPluginDiscovery -benchmem -cpuprofile=cpu.prof ./tests/benchmarks/
go tool pprof cpu.prof
```

### Memory Profiling

```bash
go test -bench=BenchmarkPluginDiscovery -benchmem -memprofile=mem.prof ./tests/benchmarks/
go tool pprof mem.prof
```

### Trace Analysis

```bash
go test -bench=BenchmarkPluginDiscovery -trace=trace.out ./tests/benchmarks/
go tool trace trace.out
```

## Known Performance Considerations

### Plugin Discovery

- **Issue**: First-time discovery is slow due to filesystem operations
- **Mitigation**: Cache plugin list; invalidate only when plugin directory changes
- **Target**: < 100ms for 100 plugins

### Context Detection

- **Issue**: Git operations can be slow in large repositories
- **Mitigation**: Cache context; avoid re-detection within same process
- **Target**: < 100ms for typical projects

### Config Merging

- **Issue**: YAML parsing adds overhead for large configs
- **Mitigation**: Lazy load command definitions; parse only when needed
- **Target**: < 150ms for 5 config files

## Future Optimizations

1. **Concurrent Plugin Discovery**: Parallelize filesystem operations
2. **Config Lazy Loading**: Load command definitions on-demand
3. **Context Caching**: Persistent cache across invocations
4. **Binary Config Format**: Optional binary format for faster loading

## References

- [Go Benchmarking Best Practices](https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go)
- [Avoiding Memory Allocations](https://www.instana.com/blog/golang-memory-allocation/)
- [Go Performance Optimization](https://github.com/dgryski/go-perfbook)

---

**Last Updated**: 2025-11-28
**Baseline Platform**: Apple M3 Max, darwin/arm64
**Go Version**: 1.21+
