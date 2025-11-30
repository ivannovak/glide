package performance

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetBudget(t *testing.T) {
	tests := []struct {
		name       string
		budgetName string
		wantOK     bool
	}{
		{
			name:       "existing budget",
			budgetName: "context_detection",
			wantOK:     true,
		},
		{
			name:       "non-existing budget",
			budgetName: "nonexistent_operation",
			wantOK:     false,
		},
		{
			name:       "plugin discovery budget",
			budgetName: "plugin_discovery",
			wantOK:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			budget, ok := GetBudget(tt.budgetName)
			assert.Equal(t, tt.wantOK, ok)
			if tt.wantOK {
				assert.Equal(t, tt.budgetName, budget.Name)
				assert.NotEmpty(t, budget.Description)
				assert.NotEmpty(t, budget.Priority)
			}
		})
	}
}

func TestBudgetValues(t *testing.T) {
	tests := []struct {
		name        string
		budgetName  string
		maxDuration time.Duration
		priority    string
	}{
		{
			name:        "context detection",
			budgetName:  "context_detection",
			maxDuration: 100 * time.Millisecond,
			priority:    "P0",
		},
		{
			name:        "config load",
			budgetName:  "config_load",
			maxDuration: 50 * time.Millisecond,
			priority:    "P1",
		},
		{
			name:        "config merge multiple",
			budgetName:  "config_merge_multiple",
			maxDuration: 100 * time.Millisecond,
			priority:    "P1",
		},
		{
			name:        "plugin discovery",
			budgetName:  "plugin_discovery",
			maxDuration: 500 * time.Millisecond,
			priority:    "P0",
		},
		{
			name:        "startup total",
			budgetName:  "startup_total",
			maxDuration: 300 * time.Millisecond,
			priority:    "P0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			budget, ok := GetBudget(tt.budgetName)
			require.True(t, ok)
			assert.Equal(t, tt.maxDuration, budget.MaxDuration)
			assert.Equal(t, tt.priority, budget.Priority)
		})
	}
}

func TestListBudgets(t *testing.T) {
	budgets := ListBudgets()
	assert.NotEmpty(t, budgets)

	// Should have at least the critical budgets
	names := make(map[string]bool)
	for _, b := range budgets {
		names[b.Name] = true
	}

	assert.True(t, names["context_detection"])
	assert.True(t, names["plugin_discovery"])
	assert.True(t, names["startup_total"])
}

func TestListByPriority(t *testing.T) {
	t.Run("P0 priorities", func(t *testing.T) {
		p0Budgets := ListByPriority("P0")
		assert.NotEmpty(t, p0Budgets)

		for _, b := range p0Budgets {
			assert.Equal(t, "P0", b.Priority)
		}

		// Should include critical operations
		names := make(map[string]bool)
		for _, b := range p0Budgets {
			names[b.Name] = true
		}
		assert.True(t, names["context_detection"])
		assert.True(t, names["plugin_discovery"])
		assert.True(t, names["startup_total"])
	})

	t.Run("P1 priorities", func(t *testing.T) {
		p1Budgets := ListByPriority("P1")
		assert.NotEmpty(t, p1Budgets)

		for _, b := range p1Budgets {
			assert.Equal(t, "P1", b.Priority)
		}
	})

	t.Run("P2 priorities", func(t *testing.T) {
		p2Budgets := ListByPriority("P2")
		assert.NotEmpty(t, p2Budgets)

		for _, b := range p2Budgets {
			assert.Equal(t, "P2", b.Priority)
		}
	})

	t.Run("non-existent priority", func(t *testing.T) {
		budgets := ListByPriority("P99")
		assert.Empty(t, budgets)
	})
}

func TestMeasure(t *testing.T) {
	t.Run("passes all criteria", func(t *testing.T) {
		result := Measure("context_detection", 50*time.Millisecond, 100, 25*1024)

		assert.True(t, result.PassesDuration)
		assert.True(t, result.PassesAllocations)
		assert.True(t, result.PassesBytes)
		assert.True(t, result.Passes)
	})

	t.Run("fails duration", func(t *testing.T) {
		result := Measure("context_detection", 200*time.Millisecond, 100, 25*1024)

		assert.False(t, result.PassesDuration)
		assert.True(t, result.PassesAllocations)
		assert.True(t, result.PassesBytes)
		assert.False(t, result.Passes)
	})

	t.Run("fails allocations", func(t *testing.T) {
		result := Measure("context_detection", 50*time.Millisecond, 500, 25*1024)

		assert.True(t, result.PassesDuration)
		assert.False(t, result.PassesAllocations)
		assert.True(t, result.PassesBytes)
		assert.False(t, result.Passes)
	})

	t.Run("fails bytes", func(t *testing.T) {
		result := Measure("context_detection", 50*time.Millisecond, 100, 100*1024)

		assert.True(t, result.PassesDuration)
		assert.True(t, result.PassesAllocations)
		assert.False(t, result.PassesBytes)
		assert.False(t, result.Passes)
	})

	t.Run("unknown operation passes", func(t *testing.T) {
		result := Measure("unknown_operation", 1*time.Hour, 1000000, 1*1024*1024*1024)

		// Unknown operations pass by default
		assert.True(t, result.Passes)
	})

	t.Run("zero limits mean no limit", func(t *testing.T) {
		// registry_get has no allocation/byte limits
		result := Measure("registry_get", 50*time.Nanosecond, 1000, 1*1024*1024)

		assert.True(t, result.PassesDuration)
		assert.True(t, result.PassesAllocations) // No limit
		assert.True(t, result.PassesBytes)       // No limit
		assert.True(t, result.Passes)
	})
}

func TestMeasurementResult(t *testing.T) {
	result := Measure("config_load", 30*time.Millisecond, 100, 15*1024)

	assert.Equal(t, "config_load", result.Operation)
	assert.Equal(t, 30*time.Millisecond, result.Duration)
	assert.Equal(t, int64(100), result.Allocations)
	assert.Equal(t, int64(15*1024), result.Bytes)
}
