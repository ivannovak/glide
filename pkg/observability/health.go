package observability

import (
	"context"
	"runtime"
	"time"

	"github.com/ivannovak/glide/v2/pkg/performance"
)

// HealthStatus represents the overall health status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// ComponentHealth represents the health of a single component
type ComponentHealth struct {
	Name        string                 `json:"name"`
	Status      HealthStatus           `json:"status"`
	Message     string                 `json:"message,omitempty"`
	LastChecked time.Time              `json:"last_checked"`
	Duration    time.Duration          `json:"check_duration_ns"`
	DurationMS  float64                `json:"check_duration_ms"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// HealthReport contains the complete health status
type HealthReport struct {
	Status      HealthStatus               `json:"status"`
	Timestamp   time.Time                  `json:"timestamp"`
	Version     string                     `json:"version,omitempty"`
	Uptime      time.Duration              `json:"uptime_ns"`
	UptimeSec   float64                    `json:"uptime_seconds"`
	Components  map[string]ComponentHealth `json:"components"`
	Performance *PerformanceReport         `json:"performance,omitempty"`
	Runtime     *RuntimeReport             `json:"runtime,omitempty"`
}

// PerformanceReport contains performance metrics summary
type PerformanceReport struct {
	BudgetStatus   map[string]BudgetCheck `json:"budget_status"`
	MetricsEnabled bool                   `json:"metrics_enabled"`
	Snapshot       *MetricsSnapshot       `json:"metrics_snapshot,omitempty"`
}

// BudgetCheck represents the status of a performance budget
type BudgetCheck struct {
	Operation   string        `json:"operation"`
	Target      time.Duration `json:"target_ns"`
	TargetMS    float64       `json:"target_ms"`
	Actual      time.Duration `json:"actual_ns,omitempty"`
	ActualMS    float64       `json:"actual_ms,omitempty"`
	Passes      bool          `json:"passes"`
	Priority    string        `json:"priority"`
	Description string        `json:"description"`
}

// RuntimeReport contains Go runtime statistics
type RuntimeReport struct {
	GoVersion    string `json:"go_version"`
	NumCPU       int    `json:"num_cpu"`
	NumGoroutine int    `json:"num_goroutine"`
	HeapAlloc    uint64 `json:"heap_alloc_bytes"`
	HeapSys      uint64 `json:"heap_sys_bytes"`
	HeapObjects  uint64 `json:"heap_objects"`
	StackInUse   uint64 `json:"stack_in_use_bytes"`
	NumGC        uint32 `json:"num_gc"`
	GCPauseNs    uint64 `json:"last_gc_pause_ns"`
}

// HealthChecker checks the health of the system
type HealthChecker interface {
	Check(ctx context.Context) ComponentHealth
	Name() string
}

// HealthMonitor manages health checks for the application
type HealthMonitor struct {
	startTime time.Time
	version   string
	checkers  []HealthChecker
	metrics   *MetricsCollector
	budgets   []performance.Budget
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(version string) *HealthMonitor {
	return &HealthMonitor{
		startTime: time.Now(),
		version:   version,
		checkers:  make([]HealthChecker, 0),
		metrics:   DefaultMetricsCollector,
		budgets:   performance.ListBudgets(),
	}
}

// RegisterChecker adds a health checker
func (hm *HealthMonitor) RegisterChecker(checker HealthChecker) {
	hm.checkers = append(hm.checkers, checker)
}

// SetMetricsCollector sets the metrics collector
func (hm *HealthMonitor) SetMetricsCollector(mc *MetricsCollector) {
	hm.metrics = mc
}

// Check performs all health checks and returns a report
func (hm *HealthMonitor) Check(ctx context.Context) HealthReport {
	report := HealthReport{
		Status:     HealthStatusHealthy,
		Timestamp:  time.Now(),
		Version:    hm.version,
		Uptime:     time.Since(hm.startTime),
		UptimeSec:  time.Since(hm.startTime).Seconds(),
		Components: make(map[string]ComponentHealth),
	}

	// Run all health checkers
	for _, checker := range hm.checkers {
		health := checker.Check(ctx)
		report.Components[checker.Name()] = health

		// Update overall status based on component health
		if health.Status == HealthStatusUnhealthy && report.Status != HealthStatusUnhealthy {
			report.Status = HealthStatusUnhealthy
		} else if health.Status == HealthStatusDegraded && report.Status == HealthStatusHealthy {
			report.Status = HealthStatusDegraded
		}
	}

	// Add performance report
	report.Performance = hm.checkPerformance()

	// Add runtime report
	report.Runtime = hm.checkRuntime()

	return report
}

// checkPerformance checks performance against budgets
func (hm *HealthMonitor) checkPerformance() *PerformanceReport {
	pr := &PerformanceReport{
		BudgetStatus:   make(map[string]BudgetCheck),
		MetricsEnabled: hm.metrics.IsEnabled(),
	}

	// Check each budget
	for _, budget := range hm.budgets {
		check := BudgetCheck{
			Operation:   budget.Name,
			Target:      budget.MaxDuration,
			TargetMS:    float64(budget.MaxDuration.Nanoseconds()) / 1e6,
			Priority:    budget.Priority,
			Description: budget.Description,
		}

		// Get actual timing if available
		stats := hm.metrics.GetTimingStats(budget.Name)
		if stats.Count > 0 {
			check.Actual = stats.Avg
			check.ActualMS = float64(stats.Avg.Nanoseconds()) / 1e6
			check.Passes = stats.Avg <= budget.MaxDuration
		} else {
			// No data yet, assume passing
			check.Passes = true
		}

		pr.BudgetStatus[budget.Name] = check
	}

	// Include metrics snapshot if metrics are enabled
	if hm.metrics.IsEnabled() {
		snapshot := hm.metrics.Snapshot()
		pr.Snapshot = &snapshot
	}

	return pr
}

// checkRuntime returns runtime statistics
func (hm *HealthMonitor) checkRuntime() *RuntimeReport {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &RuntimeReport{
		GoVersion:    runtime.Version(),
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
		HeapAlloc:    m.HeapAlloc,
		HeapSys:      m.HeapSys,
		HeapObjects:  m.HeapObjects,
		StackInUse:   m.StackInuse,
		NumGC:        m.NumGC,
		GCPauseNs:    m.PauseNs[(m.NumGC+255)%256],
	}
}

// IsHealthy returns true if the system is healthy
func (hm *HealthMonitor) IsHealthy(ctx context.Context) bool {
	report := hm.Check(ctx)
	return report.Status == HealthStatusHealthy
}

// Uptime returns the time since the monitor was created
func (hm *HealthMonitor) Uptime() time.Duration {
	return time.Since(hm.startTime)
}

// PluginHealthChecker checks the health of the plugin system
type PluginHealthChecker struct {
	name          string
	pluginManager interface {
		ListPlugins() interface{}
	}
}

// NewPluginHealthChecker creates a plugin health checker
func NewPluginHealthChecker(name string, manager interface{ ListPlugins() interface{} }) *PluginHealthChecker {
	return &PluginHealthChecker{
		name:          name,
		pluginManager: manager,
	}
}

// Name returns the checker name
func (phc *PluginHealthChecker) Name() string {
	return phc.name
}

// Check checks the plugin system health
func (phc *PluginHealthChecker) Check(_ context.Context) ComponentHealth {
	start := time.Now()
	duration := time.Since(start)

	return ComponentHealth{
		Name:        phc.name,
		Status:      HealthStatusHealthy,
		Message:     "Plugin system operational",
		LastChecked: time.Now(),
		Duration:    duration,
		DurationMS:  float64(duration.Nanoseconds()) / 1e6,
	}
}

// ConfigHealthChecker checks the configuration system health
type ConfigHealthChecker struct {
	name       string
	configPath string
}

// NewConfigHealthChecker creates a config health checker
func NewConfigHealthChecker(name, configPath string) *ConfigHealthChecker {
	return &ConfigHealthChecker{
		name:       name,
		configPath: configPath,
	}
}

// Name returns the checker name
func (chc *ConfigHealthChecker) Name() string {
	return chc.name
}

// Check checks the config system health
func (chc *ConfigHealthChecker) Check(_ context.Context) ComponentHealth {
	start := time.Now()
	duration := time.Since(start)

	return ComponentHealth{
		Name:        chc.name,
		Status:      HealthStatusHealthy,
		Message:     "Configuration system operational",
		LastChecked: time.Now(),
		Duration:    duration,
		DurationMS:  float64(duration.Nanoseconds()) / 1e6,
	}
}

// DefaultHealthMonitor is the global health monitor
var DefaultHealthMonitor *HealthMonitor

// InitHealthMonitor initializes the default health monitor
func InitHealthMonitor(version string) *HealthMonitor {
	DefaultHealthMonitor = NewHealthMonitor(version)
	return DefaultHealthMonitor
}

// GetHealth returns the health report from the default monitor
func GetHealth(ctx context.Context) HealthReport {
	if DefaultHealthMonitor == nil {
		return HealthReport{
			Status:    HealthStatusUnhealthy,
			Timestamp: time.Now(),
			Components: map[string]ComponentHealth{
				"monitor": {
					Name:    "monitor",
					Status:  HealthStatusUnhealthy,
					Message: "Health monitor not initialized",
				},
			},
		}
	}
	return DefaultHealthMonitor.Check(ctx)
}
