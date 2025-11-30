package observability

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"
)

// LogLevel defines the severity of a log entry
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// PerformanceLog represents a structured performance log entry
type PerformanceLog struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       LogLevel               `json:"level"`
	Operation   string                 `json:"operation"`
	Duration    time.Duration          `json:"duration_ns"`
	DurationMS  float64                `json:"duration_ms"`
	Success     bool                   `json:"success"`
	Error       string                 `json:"error,omitempty"`
	Labels      map[string]string      `json:"labels,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Caller      string                 `json:"caller,omitempty"`
	GoRoutines  int                    `json:"goroutines,omitempty"`
	HeapAlloc   uint64                 `json:"heap_alloc_bytes,omitempty"`
	StackAlloc  uint64                 `json:"stack_alloc_bytes,omitempty"`
	Allocations uint64                 `json:"allocations,omitempty"`
}

// PerformanceLogger provides structured performance logging
type PerformanceLogger struct {
	output           io.Writer
	minLevel         LogLevel
	enabled          bool
	includeRuntime   bool
	includeCaller    bool
	operationCounter *MetricsCollector
}

// DefaultPerformanceLogger is the global performance logger
var DefaultPerformanceLogger = NewPerformanceLogger()

// NewPerformanceLogger creates a new performance logger
func NewPerformanceLogger() *PerformanceLogger {
	return &PerformanceLogger{
		output:           os.Stderr,
		minLevel:         LogLevelInfo,
		enabled:          true,
		includeRuntime:   false,
		includeCaller:    true,
		operationCounter: DefaultMetricsCollector,
	}
}

// SetOutput sets the output writer
func (pl *PerformanceLogger) SetOutput(w io.Writer) {
	pl.output = w
}

// SetMinLevel sets the minimum log level
func (pl *PerformanceLogger) SetMinLevel(level LogLevel) {
	pl.minLevel = level
}

// Enable enables performance logging
func (pl *PerformanceLogger) Enable() {
	pl.enabled = true
}

// Disable disables performance logging
func (pl *PerformanceLogger) Disable() {
	pl.enabled = false
}

// SetIncludeRuntime enables/disables runtime stats in logs
func (pl *PerformanceLogger) SetIncludeRuntime(include bool) {
	pl.includeRuntime = include
}

// SetIncludeCaller enables/disables caller info in logs
func (pl *PerformanceLogger) SetIncludeCaller(include bool) {
	pl.includeCaller = include
}

// shouldLog returns true if the given level should be logged
func (pl *PerformanceLogger) shouldLog(level LogLevel) bool {
	if !pl.enabled {
		return false
	}

	levels := map[LogLevel]int{
		LogLevelDebug: 0,
		LogLevelInfo:  1,
		LogLevelWarn:  2,
		LogLevelError: 3,
	}

	return levels[level] >= levels[pl.minLevel]
}

// LogOperation logs a completed operation with timing
func (pl *PerformanceLogger) LogOperation(operation string, duration time.Duration, err error, labels map[string]string, metadata map[string]interface{}) {
	level := LogLevelInfo
	if err != nil {
		level = LogLevelError
	}

	if !pl.shouldLog(level) {
		return
	}

	log := PerformanceLog{
		Timestamp:  time.Now(),
		Level:      level,
		Operation:  operation,
		Duration:   duration,
		DurationMS: float64(duration.Nanoseconds()) / 1e6,
		Success:    err == nil,
		Labels:     labels,
		Metadata:   metadata,
	}

	if err != nil {
		log.Error = err.Error()
	}

	if pl.includeCaller {
		_, file, line, ok := runtime.Caller(1)
		if ok {
			log.Caller = fmt.Sprintf("%s:%d", file, line)
		}
	}

	if pl.includeRuntime {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		log.GoRoutines = runtime.NumGoroutine()
		log.HeapAlloc = m.HeapAlloc
		log.StackAlloc = m.StackSys
		log.Allocations = m.Mallocs
	}

	// Record to metrics
	if pl.operationCounter != nil {
		pl.operationCounter.RecordTiming(operation, duration)
		if err != nil {
			pl.operationCounter.IncrementCounter(operation + "_errors")
		}
	}

	pl.writeLog(log)
}

// LogOperationStart creates a new operation tracker
func (pl *PerformanceLogger) LogOperationStart(operation string, labels map[string]string) *OperationTracker {
	return &OperationTracker{
		operation: operation,
		start:     time.Now(),
		labels:    labels,
		metadata:  make(map[string]interface{}),
		logger:    pl,
	}
}

// writeLog writes the log entry to the output
func (pl *PerformanceLogger) writeLog(log PerformanceLog) {
	data, err := json.Marshal(log)
	if err != nil {
		return
	}
	fmt.Fprintln(pl.output, string(data))
}

// Debug logs a debug-level performance event
func (pl *PerformanceLogger) Debug(operation string, duration time.Duration, metadata map[string]interface{}) {
	if !pl.shouldLog(LogLevelDebug) {
		return
	}

	log := PerformanceLog{
		Timestamp:  time.Now(),
		Level:      LogLevelDebug,
		Operation:  operation,
		Duration:   duration,
		DurationMS: float64(duration.Nanoseconds()) / 1e6,
		Success:    true,
		Metadata:   metadata,
	}

	pl.writeLog(log)
}

// Info logs an info-level performance event
func (pl *PerformanceLogger) Info(operation string, duration time.Duration, metadata map[string]interface{}) {
	if !pl.shouldLog(LogLevelInfo) {
		return
	}

	log := PerformanceLog{
		Timestamp:  time.Now(),
		Level:      LogLevelInfo,
		Operation:  operation,
		Duration:   duration,
		DurationMS: float64(duration.Nanoseconds()) / 1e6,
		Success:    true,
		Metadata:   metadata,
	}

	pl.writeLog(log)
}

// Warn logs a warning-level performance event
func (pl *PerformanceLogger) Warn(operation string, duration time.Duration, metadata map[string]interface{}) {
	if !pl.shouldLog(LogLevelWarn) {
		return
	}

	log := PerformanceLog{
		Timestamp:  time.Now(),
		Level:      LogLevelWarn,
		Operation:  operation,
		Duration:   duration,
		DurationMS: float64(duration.Nanoseconds()) / 1e6,
		Success:    true,
		Metadata:   metadata,
	}

	pl.writeLog(log)
}

// OperationTracker tracks an in-progress operation
type OperationTracker struct {
	operation string
	start     time.Time
	labels    map[string]string
	metadata  map[string]interface{}
	logger    *PerformanceLogger
}

// AddMetadata adds metadata to the operation
func (ot *OperationTracker) AddMetadata(key string, value interface{}) {
	ot.metadata[key] = value
}

// Finish completes the operation tracking
func (ot *OperationTracker) Finish(err error) time.Duration {
	duration := time.Since(ot.start)
	ot.logger.LogOperation(ot.operation, duration, err, ot.labels, ot.metadata)
	return duration
}

// Duration returns the elapsed time without finishing
func (ot *OperationTracker) Duration() time.Duration {
	return time.Since(ot.start)
}

// Convenience functions using DefaultPerformanceLogger

// LogOp logs an operation using the default logger
func LogOp(operation string, duration time.Duration, err error, metadata map[string]interface{}) {
	DefaultPerformanceLogger.LogOperation(operation, duration, err, nil, metadata)
}

// StartOp starts tracking an operation using the default logger
func StartOp(operation string) *OperationTracker {
	return DefaultPerformanceLogger.LogOperationStart(operation, nil)
}

// StartOpWithLabels starts tracking an operation with labels
func StartOpWithLabels(operation string, labels map[string]string) *OperationTracker {
	return DefaultPerformanceLogger.LogOperationStart(operation, labels)
}
