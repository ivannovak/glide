package observability

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsCollector_Counter(t *testing.T) {
	mc := NewMetricsCollector()

	// Test increment
	mc.IncrementCounter("test_counter")
	assert.Equal(t, int64(1), mc.GetCounter("test_counter"))

	// Test increment multiple times
	mc.IncrementCounter("test_counter")
	mc.IncrementCounter("test_counter")
	assert.Equal(t, int64(3), mc.GetCounter("test_counter"))

	// Test increment by value
	mc.IncrementCounterBy("test_counter", 10)
	assert.Equal(t, int64(13), mc.GetCounter("test_counter"))

	// Test non-existent counter
	assert.Equal(t, int64(0), mc.GetCounter("nonexistent"))
}

func TestMetricsCollector_Gauge(t *testing.T) {
	mc := NewMetricsCollector()

	// Test set gauge
	mc.SetGauge("test_gauge", 42.5)
	assert.Equal(t, 42.5, mc.GetGauge("test_gauge"))

	// Test update gauge
	mc.SetGauge("test_gauge", 100.0)
	assert.Equal(t, 100.0, mc.GetGauge("test_gauge"))

	// Test non-existent gauge
	assert.Equal(t, 0.0, mc.GetGauge("nonexistent"))
}

func TestMetricsCollector_Timing(t *testing.T) {
	mc := NewMetricsCollector()

	// Record some timings
	mc.RecordTiming("test_timing", 100*time.Millisecond)
	mc.RecordTiming("test_timing", 200*time.Millisecond)
	mc.RecordTiming("test_timing", 300*time.Millisecond)

	stats := mc.GetTimingStats("test_timing")
	assert.Equal(t, 3, stats.Count)
	assert.Equal(t, 100*time.Millisecond, stats.Min)
	assert.Equal(t, 300*time.Millisecond, stats.Max)
	assert.Equal(t, 200*time.Millisecond, stats.Avg)
	assert.Equal(t, 600*time.Millisecond, stats.Total)
}

func TestMetricsCollector_TimingMaxSamples(t *testing.T) {
	mc := NewMetricsCollector()
	mc.maxSamples = 3 // Small sample size for testing

	// Record more than max samples
	for i := 0; i < 5; i++ {
		mc.RecordTiming("test_timing", time.Duration(i)*time.Millisecond)
	}

	stats := mc.GetTimingStats("test_timing")
	assert.Equal(t, 3, stats.Samples) // Should only have last 3 samples
}

func TestMetricsCollector_Histogram(t *testing.T) {
	mc := NewMetricsCollector()

	// Create histogram with buckets
	buckets := []float64{10, 50, 100, 500, 1000}
	h := mc.CreateHistogram("test_histogram", buckets)

	// Observe some values
	h.Observe(5)    // bucket 10
	h.Observe(25)   // bucket 50
	h.Observe(75)   // bucket 100
	h.Observe(200)  // bucket 500
	h.Observe(2000) // bucket +Inf

	sum, count, histBuckets := h.Summary()
	assert.Equal(t, int64(5), count)
	assert.Equal(t, float64(5+25+75+200+2000), sum)
	assert.Equal(t, 6, len(histBuckets)) // 5 buckets + 1 for +Inf

	// Verify cumulative counts
	assert.Equal(t, int64(1), histBuckets[0].Count) // <= 10
	assert.Equal(t, int64(2), histBuckets[1].Count) // <= 50
	assert.Equal(t, int64(3), histBuckets[2].Count) // <= 100
	assert.Equal(t, int64(4), histBuckets[3].Count) // <= 500
	assert.Equal(t, int64(4), histBuckets[4].Count) // <= 1000
	assert.Equal(t, int64(5), histBuckets[5].Count) // +Inf
}

func TestMetricsCollector_EnableDisable(t *testing.T) {
	mc := NewMetricsCollector()

	// Should be enabled by default
	assert.True(t, mc.IsEnabled())

	// Disable and verify counters don't increment
	mc.Disable()
	assert.False(t, mc.IsEnabled())

	mc.IncrementCounter("disabled_counter")
	assert.Equal(t, int64(0), mc.GetCounter("disabled_counter"))

	// Enable and verify counters work
	mc.Enable()
	mc.IncrementCounter("enabled_counter")
	assert.Equal(t, int64(1), mc.GetCounter("enabled_counter"))
}

func TestMetricsCollector_Snapshot(t *testing.T) {
	mc := NewMetricsCollector()

	mc.IncrementCounter("counter1")
	mc.IncrementCounter("counter2")
	mc.IncrementCounter("counter2")
	mc.SetGauge("gauge1", 123.45)
	mc.RecordTiming("timing1", 100*time.Millisecond)

	snapshot := mc.Snapshot()

	assert.Equal(t, int64(1), snapshot.Counters["counter1"])
	assert.Equal(t, int64(2), snapshot.Counters["counter2"])
	assert.Equal(t, 123.45, snapshot.Gauges["gauge1"])
	assert.Equal(t, 1, snapshot.Timings["timing1"].Count)
	assert.NotZero(t, snapshot.Timestamp)
}

func TestMetricsCollector_Reset(t *testing.T) {
	mc := NewMetricsCollector()

	mc.IncrementCounter("counter")
	mc.SetGauge("gauge", 42)
	mc.RecordTiming("timing", 100*time.Millisecond)

	mc.Reset()

	assert.Equal(t, int64(0), mc.GetCounter("counter"))
	assert.Equal(t, 0.0, mc.GetGauge("gauge"))
	stats := mc.GetTimingStats("timing")
	assert.Equal(t, 0, stats.Count)
}

func TestTimer(t *testing.T) {
	mc := NewMetricsCollector()

	timer := StartTimerWithCollector("test_operation", mc)
	time.Sleep(10 * time.Millisecond)
	duration := timer.Stop()

	assert.GreaterOrEqual(t, duration, 10*time.Millisecond)

	stats := mc.GetTimingStats("test_operation")
	assert.Equal(t, 1, stats.Count)
	assert.GreaterOrEqual(t, stats.Min, 10*time.Millisecond)
}

func TestTimer_Duration(t *testing.T) {
	timer := StartTimerWithCollector("test", NewMetricsCollector())
	time.Sleep(5 * time.Millisecond)

	d := timer.Duration()
	assert.GreaterOrEqual(t, d, 5*time.Millisecond)
}

func TestMetricsCollector_Concurrent(t *testing.T) {
	mc := NewMetricsCollector()
	var wg sync.WaitGroup

	// Concurrent counter increments
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mc.IncrementCounter("concurrent_counter")
		}()
	}

	// Concurrent timing records
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			mc.RecordTiming("concurrent_timing", time.Duration(i)*time.Microsecond)
		}(i)
	}

	wg.Wait()

	assert.Equal(t, int64(100), mc.GetCounter("concurrent_counter"))
	stats := mc.GetTimingStats("concurrent_timing")
	assert.Equal(t, 100, stats.Count)
}

func TestHistogramMetric_Concurrent(t *testing.T) {
	h := NewHistogramMetric("test", []float64{10, 100, 1000}, nil)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(val float64) {
			defer wg.Done()
			h.Observe(val)
		}(float64(i))
	}

	wg.Wait()

	sum, count, _ := h.Summary()
	assert.Equal(t, int64(100), count)
	// Sum should be 0+1+2+...+99 = 4950
	assert.Equal(t, float64(4950), sum)
}

func TestDefaultMetricsCollector(t *testing.T) {
	// Reset default collector for testing
	DefaultMetricsCollector.Reset()

	IncrementCounter("default_counter")
	SetGauge("default_gauge", 99.9)
	RecordTiming("default_timing", 50*time.Millisecond)

	snapshot := GetSnapshot()

	assert.Equal(t, int64(1), snapshot.Counters["default_counter"])
	assert.Equal(t, 99.9, snapshot.Gauges["default_gauge"])
	assert.Equal(t, 1, snapshot.Timings["default_timing"].Count)
}

func BenchmarkCounterIncrement(b *testing.B) {
	mc := NewMetricsCollector()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mc.IncrementCounter("benchmark_counter")
	}
}

func BenchmarkTimingRecord(b *testing.B) {
	mc := NewMetricsCollector()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mc.RecordTiming("benchmark_timing", 100*time.Millisecond)
	}
}

func BenchmarkHistogramObserve(b *testing.B) {
	h := NewHistogramMetric("benchmark", []float64{10, 50, 100, 500, 1000}, nil)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		h.Observe(float64(i % 1000))
	}
}

func BenchmarkTimer(b *testing.B) {
	mc := NewMetricsCollector()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		timer := StartTimerWithCollector("benchmark", mc)
		timer.Stop()
	}
}

func TestEmptyTimingStats(t *testing.T) {
	mc := NewMetricsCollector()
	stats := mc.GetTimingStats("nonexistent")
	require.Equal(t, 0, stats.Count)
	require.Equal(t, time.Duration(0), stats.Min)
	require.Equal(t, time.Duration(0), stats.Max)
}
