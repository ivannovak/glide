// Package performance defines performance budgets and provides utilities
// for measuring and validating operation performance against targets.
package performance

import "time"

// Budget defines a performance budget for an operation
type Budget struct {
	// Name is the operation name (e.g., "context_detection")
	Name string

	// MaxDuration is the maximum acceptable duration for this operation
	MaxDuration time.Duration

	// MaxAllocations is the maximum number of heap allocations allowed (0 = no limit)
	MaxAllocations int64

	// MaxBytes is the maximum bytes allocated per operation (0 = no limit)
	MaxBytes int64

	// Description explains what this budget covers
	Description string

	// Priority indicates the importance of meeting this budget (P0 = critical)
	Priority string
}

// Budgets contains all defined performance budgets for Glide operations
var Budgets = map[string]Budget{
	// Context Detection - Critical for startup performance
	"context_detection": {
		Name:           "context_detection",
		MaxDuration:    100 * time.Millisecond,
		MaxAllocations: 200,
		MaxBytes:       50 * 1024, // 50KB
		Description:    "Time to detect project context (git root, frameworks, worktree mode)",
		Priority:       "P0",
	},

	// Configuration Loading
	"config_load": {
		Name:           "config_load",
		MaxDuration:    50 * time.Millisecond,
		MaxAllocations: 150,
		MaxBytes:       20 * 1024, // 20KB
		Description:    "Time to load a single configuration file",
		Priority:       "P1",
	},

	// Configuration Merging
	"config_merge_single": {
		Name:           "config_merge_single",
		MaxDuration:    30 * time.Millisecond,
		MaxAllocations: 150,
		MaxBytes:       15 * 1024, // 15KB
		Description:    "Time to merge a single configuration file",
		Priority:       "P1",
	},

	"config_merge_multiple": {
		Name:           "config_merge_multiple",
		MaxDuration:    100 * time.Millisecond,
		MaxAllocations: 600,
		MaxBytes:       50 * 1024, // 50KB
		Description:    "Time to merge multiple (5+) configuration files",
		Priority:       "P1",
	},

	// Plugin Operations
	"plugin_discovery": {
		Name:           "plugin_discovery",
		MaxDuration:    500 * time.Millisecond,
		MaxAllocations: 10000,
		MaxBytes:       2 * 1024 * 1024, // 2MB
		Description:    "Time to discover and enumerate all available plugins",
		Priority:       "P0",
	},

	"plugin_load": {
		Name:           "plugin_load",
		MaxDuration:    200 * time.Millisecond,
		MaxAllocations: 1000,
		MaxBytes:       512 * 1024, // 512KB
		Description:    "Time to load and initialize a single plugin",
		Priority:       "P1",
	},

	"plugin_cache_get": {
		Name:           "plugin_cache_get",
		MaxDuration:    10 * time.Microsecond,
		MaxAllocations: 0,
		MaxBytes:       0,
		Description:    "Time to retrieve a plugin from cache",
		Priority:       "P2",
	},

	// Startup
	"startup_total": {
		Name:           "startup_total",
		MaxDuration:    300 * time.Millisecond,
		MaxAllocations: 10000,
		MaxBytes:       5 * 1024 * 1024, // 5MB
		Description:    "Total time from start to ready state (excluding plugins)",
		Priority:       "P0",
	},

	// Command Operations
	"command_lookup": {
		Name:           "command_lookup",
		MaxDuration:    1 * time.Millisecond,
		MaxAllocations: 10,
		MaxBytes:       1024, // 1KB
		Description:    "Time to look up a command by name",
		Priority:       "P1",
	},

	// Error Handling
	"error_creation": {
		Name:           "error_creation",
		MaxDuration:    1 * time.Microsecond,
		MaxAllocations: 5,
		MaxBytes:       1024, // 1KB
		Description:    "Time to create a structured error",
		Priority:       "P2",
	},

	"error_wrap": {
		Name:           "error_wrap",
		MaxDuration:    500 * time.Nanosecond,
		MaxAllocations: 5,
		MaxBytes:       512, // 512B
		Description:    "Time to wrap an existing error",
		Priority:       "P2",
	},

	// Validation
	"path_validation": {
		Name:           "path_validation",
		MaxDuration:    50 * time.Microsecond,
		MaxAllocations: 100,
		MaxBytes:       10 * 1024, // 10KB
		Description:    "Time to validate a file path for security",
		Priority:       "P1",
	},

	// Registry Operations
	"registry_get": {
		Name:           "registry_get",
		MaxDuration:    100 * time.Nanosecond,
		MaxAllocations: 0,
		MaxBytes:       0,
		Description:    "Time to retrieve an item from registry",
		Priority:       "P2",
	},

	"registry_list": {
		Name:           "registry_list",
		MaxDuration:    10 * time.Microsecond,
		MaxAllocations: 5,
		MaxBytes:       4 * 1024, // 4KB
		Description:    "Time to list all items in registry (100 items)",
		Priority:       "P2",
	},
}

// GetBudget returns the budget for a given operation name
func GetBudget(name string) (Budget, bool) {
	budget, ok := Budgets[name]
	return budget, ok
}

// ListBudgets returns all defined budgets
func ListBudgets() []Budget {
	result := make([]Budget, 0, len(Budgets))
	for _, b := range Budgets {
		result = append(result, b)
	}
	return result
}

// ListByPriority returns budgets filtered by priority
func ListByPriority(priority string) []Budget {
	var result []Budget
	for _, b := range Budgets {
		if b.Priority == priority {
			result = append(result, b)
		}
	}
	return result
}

// MeasurementResult captures the result of a performance measurement
type MeasurementResult struct {
	// Operation is the name of the measured operation
	Operation string

	// Duration is the measured time
	Duration time.Duration

	// Allocations is the number of heap allocations
	Allocations int64

	// Bytes is the total bytes allocated
	Bytes int64

	// PassesDuration indicates if duration is within budget
	PassesDuration bool

	// PassesAllocations indicates if allocations are within budget
	PassesAllocations bool

	// PassesBytes indicates if bytes are within budget
	PassesBytes bool

	// Passes indicates if all criteria pass
	Passes bool
}

// Measure validates a measurement against a budget
func Measure(name string, duration time.Duration, allocations, bytes int64) MeasurementResult {
	budget, ok := GetBudget(name)
	if !ok {
		return MeasurementResult{
			Operation: name,
			Duration:  duration,
			Passes:    true, // No budget defined, pass by default
		}
	}

	result := MeasurementResult{
		Operation:   name,
		Duration:    duration,
		Allocations: allocations,
		Bytes:       bytes,
	}

	// Check duration
	result.PassesDuration = duration <= budget.MaxDuration

	// Check allocations (0 means no limit)
	result.PassesAllocations = budget.MaxAllocations == 0 || allocations <= budget.MaxAllocations

	// Check bytes (0 means no limit)
	result.PassesBytes = budget.MaxBytes == 0 || bytes <= budget.MaxBytes

	// Overall pass
	result.Passes = result.PassesDuration && result.PassesAllocations && result.PassesBytes

	return result
}
