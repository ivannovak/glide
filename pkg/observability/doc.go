// Package observability provides metrics collection, logging, and health monitoring.
//
// This package implements observability infrastructure for performance monitoring,
// structured logging, and health checks. It supports counters, gauges, histograms,
// timing measurements, and health endpoints.
//
// # Metrics Collection
//
// Collect various metric types:
//
//	mc := observability.NewMetricsCollector()
//
//	// Counters - monotonically increasing values
//	mc.IncrementCounter("requests_total")
//	mc.IncrementCounterBy("bytes_processed", 1024)
//
//	// Gauges - values that can go up or down
//	mc.SetGauge("connections_active", 42)
//	mc.SetGauge("memory_usage_mb", 256.5)
//
//	// Timings - operation duration tracking
//	mc.RecordTiming("api_latency", 150*time.Millisecond)
//	stats := mc.GetTimingStats("api_latency") // Min, Max, Avg, P95
//
// # Timer Utility
//
// Measure operation duration:
//
//	timer := observability.StartTimer("database_query")
//	result, err := db.Query(sql)
//	duration := timer.Stop() // Automatically recorded
//
//	// Check duration without stopping
//	elapsed := timer.Duration()
//
// # Histograms
//
// Track value distributions:
//
//	h := mc.CreateHistogram("request_size",
//	    []float64{100, 500, 1000, 5000, 10000})
//	h.Observe(256)
//	h.Observe(1024)
//	h.Observe(8192)
//
//	sum, count, buckets := h.Summary()
//
// # Health Monitoring
//
// Implement health checks for system components:
//
//	monitor := observability.NewHealthMonitor("1.0.0")
//
//	// Register component health checkers
//	monitor.RegisterChecker(observability.NewPluginHealthChecker("plugins", pluginMgr))
//	monitor.RegisterChecker(observability.NewConfigHealthChecker("config", cfgPath))
//
//	// Check overall health
//	report := monitor.Check(ctx)
//	if report.Status != observability.HealthStatusHealthy {
//	    log.Warn("System degraded", "status", report.Status)
//	}
//
// # Performance Logging
//
// Structured performance logging:
//
//	logger := observability.NewPerformanceLogger()
//	tracker := logger.LogOperationStart("process_batch", map[string]string{
//	    "batch_id": batchID,
//	})
//	tracker.AddMetadata("items", itemCount)
//	duration := tracker.Finish(err)
//
// # Default Collectors
//
// Use global default instances for convenience:
//
//	observability.IncrementCounter("global_counter")
//	observability.RecordTiming("global_timing", duration)
//	snapshot := observability.GetSnapshot()
//
// # Thread Safety
//
// All collectors and monitors are safe for concurrent use.
// Atomic operations are used for counters, and RWMutex for complex types.
package observability
