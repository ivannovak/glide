package plugin

import (
	"testing"

	v1 "github.com/ivannovak/glide/v3/pkg/plugin/sdk/v1"
	"github.com/stretchr/testify/assert"
)

func TestCustomCategories(t *testing.T) {
	// Reset global state
	globalPluginCategories = nil

	// Create integration
	integration := NewRuntimePluginIntegration()

	// Test registering custom categories
	categories := []*v1.CustomCategory{
		{
			Id:          "infrastructure",
			Name:        "Infrastructure Management",
			Description: "AWS and cloud resources",
			Priority:    110,
		},
		{
			Id:          "monitoring",
			Name:        "Monitoring",
			Description: "Metrics and logs",
			Priority:    120,
		},
	}

	integration.registerCustomCategories(categories)

	// Verify they were registered
	assert.Equal(t, 2, len(integration.GetCustomCategories()))
	assert.Equal(t, 2, len(GetGlobalPluginCategories()))

	// Verify content
	global := GetGlobalPluginCategories()
	assert.Equal(t, "infrastructure", global[0].Id)
	assert.Equal(t, "Infrastructure Management", global[0].Name)
	assert.Equal(t, int32(110), global[0].Priority)

	assert.Equal(t, "monitoring", global[1].Id)
	assert.Equal(t, "Monitoring", global[1].Name)
	assert.Equal(t, int32(120), global[1].Priority)
}

func TestCustomCategoriesMultiplePlugins(t *testing.T) {
	// Reset global state
	globalPluginCategories = nil

	// Simulate multiple plugins registering categories
	integration1 := NewRuntimePluginIntegration()
	integration2 := NewRuntimePluginIntegration()

	categories1 := []*v1.CustomCategory{
		{
			Id:       "cloud",
			Name:     "Cloud Operations",
			Priority: 100,
		},
	}

	categories2 := []*v1.CustomCategory{
		{
			Id:       "security",
			Name:     "Security Tools",
			Priority: 105,
		},
	}

	integration1.registerCustomCategories(categories1)
	integration2.registerCustomCategories(categories2)

	// Global should have both
	global := GetGlobalPluginCategories()
	assert.Equal(t, 2, len(global))
	assert.Equal(t, "cloud", global[0].Id)
	assert.Equal(t, "security", global[1].Id)
}
