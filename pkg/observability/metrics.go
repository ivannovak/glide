// Package observability provides performance metrics collection and monitoring.
package observability

import (
	"sync"
	"sync/atomic"
	"time"
)

// MetricType defines the type of metric
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeTiming    MetricType = "timing"
)

// Metric represents a single metric value
type Metric struct {
	Name      string                 `json:"name"`
	Type      MetricType             `json:"type"`
	Value     float64                `json:"value"`
	Labels    map[string]string      `json:"labels,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// HistogramBucket defines a bucket for histogram metrics
type HistogramBucket struct {
	UpperBound float64 `json:"upper_bound"`
	Count      int64   `json:"count"`
}

// HistogramMetric provides histogram functionality
type HistogramMetric struct {
	mu      sync.RWMutex
	name    string
	labels  map[string]string
	buckets []float64 // Upper bounds
	counts  []int64   // Counts per bucket
	sum     float64
	count   int64
}

// NewHistogramMetric creates a new histogram with the given buckets
func NewHistogramMetric(name string, buckets []float64, labels map[string]string) *HistogramMetric {
	return &HistogramMetric{
		name:    name,
		labels:  labels,
		buckets: buckets,
		counts:  make([]int64, len(buckets)+1), // +1 for +Inf bucket
	}
}

// Observe records a value in the histogram
func (h *HistogramMetric) Observe(value float64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.sum += value
	h.count++

	// Find the right bucket
	for i, bound := range h.buckets {
		if value <= bound {
			h.counts[i]++
			return
		}
	}
	// Value is greater than all bounds, goes in +Inf bucket
	h.counts[len(h.buckets)]++
}

// Summary returns the histogram summary
func (h *HistogramMetric) Summary() (sum float64, count int64, buckets []HistogramBucket) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	buckets = make([]HistogramBucket, len(h.buckets)+1)
	var cumCount int64
	for i, bound := range h.buckets {
		cumCount += h.counts[i]
		buckets[i] = HistogramBucket{
			UpperBound: bound,
			Count:      cumCount,
		}
	}
	// +Inf bucket
	cumCount += h.counts[len(h.buckets)]
	buckets[len(h.buckets)] = HistogramBucket{
		UpperBound: float64(^uint64(0) >> 1), // MaxFloat64 approximation
		Count:      cumCount,
	}

	return h.sum, h.count, buckets
}

// MetricsCollector collects and aggregates metrics
type MetricsCollector struct {
	mu         sync.RWMutex
	counters   map[string]*int64
	gauges     map[string]*float64
	histograms map[string]*HistogramMetric
	timings    map[string][]time.Duration
	enabled    bool
	maxSamples int
}

// DefaultMetricsCollector is the global metrics collector
var DefaultMetricsCollector = NewMetricsCollector()

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		counters:   make(map[string]*int64),
		gauges:     make(map[string]*float64),
		histograms: make(map[string]*HistogramMetric),
		timings:    make(map[string][]time.Duration),
		enabled:    true,
		maxSamples: 1000, // Keep last 1000 timing samples
	}
}

// Enable enables metrics collection
func (mc *MetricsCollector) Enable() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.enabled = true
}

// Disable disables metrics collection
func (mc *MetricsCollector) Disable() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.enabled = false
}

// IsEnabled returns whether metrics collection is enabled
func (mc *MetricsCollector) IsEnabled() bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.enabled
}

// IncrementCounter increments a counter metric
func (mc *MetricsCollector) IncrementCounter(name string) {
	mc.IncrementCounterBy(name, 1)
}

// IncrementCounterBy increments a counter metric by a specific value
func (mc *MetricsCollector) IncrementCounterBy(name string, delta int64) {
	if !mc.IsEnabled() {
		return
	}

	mc.mu.Lock()
	if mc.counters[name] == nil {
		var v int64
		mc.counters[name] = &v
	}
	mc.mu.Unlock()

	atomic.AddInt64(mc.counters[name], delta)
}

// GetCounter returns the current value of a counter
func (mc *MetricsCollector) GetCounter(name string) int64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if mc.counters[name] == nil {
		return 0
	}
	return atomic.LoadInt64(mc.counters[name])
}

// SetGauge sets a gauge metric value
func (mc *MetricsCollector) SetGauge(name string, value float64) {
	if !mc.IsEnabled() {
		return
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.gauges[name] == nil {
		mc.gauges[name] = new(float64)
	}
	*mc.gauges[name] = value
}

// GetGauge returns the current value of a gauge
func (mc *MetricsCollector) GetGauge(name string) float64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if mc.gauges[name] == nil {
		return 0
	}
	return *mc.gauges[name]
}

// RecordTiming records a timing measurement
func (mc *MetricsCollector) RecordTiming(name string, duration time.Duration) {
	if !mc.IsEnabled() {
		return
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	timings := mc.timings[name]
	if len(timings) >= mc.maxSamples {
		// Remove oldest sample (FIFO)
		timings = timings[1:]
	}
	mc.timings[name] = append(timings, duration)

	// Also update histogram if it exists
	if h, ok := mc.histograms[name]; ok {
		h.Observe(float64(duration.Nanoseconds()))
	}
}

// GetTimingStats returns statistics for a timing metric
func (mc *MetricsCollector) GetTimingStats(name string) TimingStats {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	timings := mc.timings[name]
	if len(timings) == 0 {
		return TimingStats{}
	}

	var sum time.Duration
	minTiming := timings[0]
	maxTiming := timings[0]

	for _, d := range timings {
		sum += d
		if d < minTiming {
			minTiming = d
		}
		if d > maxTiming {
			maxTiming = d
		}
	}

	return TimingStats{
		Count:   len(timings),
		Min:     minTiming,
		Max:     maxTiming,
		Avg:     time.Duration(int64(sum) / int64(len(timings))),
		Total:   sum,
		Samples: len(timings),
	}
}

// TimingStats contains statistics for timing measurements
type TimingStats struct {
	Count   int           `json:"count"`
	Min     time.Duration `json:"min"`
	Max     time.Duration `json:"max"`
	Avg     time.Duration `json:"avg"`
	Total   time.Duration `json:"total"`
	Samples int           `json:"samples"`
}

// CreateHistogram creates a histogram metric
func (mc *MetricsCollector) CreateHistogram(name string, buckets []float64) *HistogramMetric {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	h := NewHistogramMetric(name, buckets, nil)
	mc.histograms[name] = h
	return h
}

// GetHistogram returns a histogram by name
func (mc *MetricsCollector) GetHistogram(name string) *HistogramMetric {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.histograms[name]
}

// Snapshot returns a snapshot of all metrics
func (mc *MetricsCollector) Snapshot() MetricsSnapshot {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	snapshot := MetricsSnapshot{
		Timestamp: time.Now(),
		Counters:  make(map[string]int64),
		Gauges:    make(map[string]float64),
		Timings:   make(map[string]TimingStats),
	}

	for name, counter := range mc.counters {
		snapshot.Counters[name] = atomic.LoadInt64(counter)
	}

	for name, gauge := range mc.gauges {
		snapshot.Gauges[name] = *gauge
	}

	for name := range mc.timings {
		snapshot.Timings[name] = mc.GetTimingStats(name)
	}

	return snapshot
}

// Reset resets all metrics
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.counters = make(map[string]*int64)
	mc.gauges = make(map[string]*float64)
	mc.histograms = make(map[string]*HistogramMetric)
	mc.timings = make(map[string][]time.Duration)
}

// MetricsSnapshot contains a point-in-time snapshot of all metrics
type MetricsSnapshot struct {
	Timestamp time.Time              `json:"timestamp"`
	Counters  map[string]int64       `json:"counters"`
	Gauges    map[string]float64     `json:"gauges"`
	Timings   map[string]TimingStats `json:"timings"`
}

// Timer provides a convenient way to measure operation duration
type Timer struct {
	name      string
	start     time.Time
	collector *MetricsCollector
}

// StartTimer starts a new timer
func StartTimer(name string) *Timer {
	return StartTimerWithCollector(name, DefaultMetricsCollector)
}

// StartTimerWithCollector starts a timer with a specific collector
func StartTimerWithCollector(name string, collector *MetricsCollector) *Timer {
	return &Timer{
		name:      name,
		start:     time.Now(),
		collector: collector,
	}
}

// Stop stops the timer and records the duration
func (t *Timer) Stop() time.Duration {
	duration := time.Since(t.start)
	t.collector.RecordTiming(t.name, duration)
	return duration
}

// Duration returns the elapsed time without stopping
func (t *Timer) Duration() time.Duration {
	return time.Since(t.start)
}

// Convenience functions using DefaultMetricsCollector

// IncrementCounter increments a counter using the default collector
func IncrementCounter(name string) {
	DefaultMetricsCollector.IncrementCounter(name)
}

// SetGauge sets a gauge using the default collector
func SetGauge(name string, value float64) {
	DefaultMetricsCollector.SetGauge(name, value)
}

// RecordTiming records timing using the default collector
func RecordTiming(name string, duration time.Duration) {
	DefaultMetricsCollector.RecordTiming(name, duration)
}

// GetSnapshot returns a snapshot from the default collector
func GetSnapshot() MetricsSnapshot {
	return DefaultMetricsCollector.Snapshot()
}
