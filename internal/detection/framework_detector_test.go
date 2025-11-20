package detection

import (
	"testing"
	"time"

	"github.com/ivannovak/glide/pkg/plugin/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockFrameworkDetector is a mock implementation for testing
type MockFrameworkDetector struct {
	patterns        sdk.DetectionPatterns
	result          *sdk.DetectionResult
	err             error
	delay           time.Duration
	detectFunc      func(string) (*sdk.DetectionResult, error)
	detectCallCount int
}

func (m *MockFrameworkDetector) GetDetectionPatterns() sdk.DetectionPatterns {
	return m.patterns
}

func (m *MockFrameworkDetector) Detect(projectPath string) (*sdk.DetectionResult, error) {
	m.detectCallCount++
	if m.detectFunc != nil {
		return m.detectFunc(projectPath)
	}
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	return m.result, m.err
}

func (m *MockFrameworkDetector) GetDefaultCommands() map[string]sdk.CommandDefinition {
	return map[string]sdk.CommandDefinition{
		"test": {
			Cmd:         "mock test",
			Description: "Mock test command",
		},
	}
}

func (m *MockFrameworkDetector) EnhanceContext(ctx map[string]interface{}) error {
	return nil
}

func TestFrameworkDetector(t *testing.T) {
	t.Run("detect frameworks from multiple plugins", func(t *testing.T) {
		detector := NewFrameworkDetector()

		// Register mock plugins
		goPlugin := &MockFrameworkDetector{
			result: &sdk.DetectionResult{
				Detected:   true,
				Confidence: 90,
				Framework: sdk.FrameworkInfo{
					Name:    "go",
					Version: "1.20",
					Type:    "language",
				},
				Commands: map[string]string{
					"build": "go build",
					"test":  "go test",
				},
			},
		}

		nodePlugin := &MockFrameworkDetector{
			result: &sdk.DetectionResult{
				Detected:   true,
				Confidence: 85,
				Framework: sdk.FrameworkInfo{
					Name:    "node",
					Version: "18.0.0",
					Type:    "language",
				},
				Commands: map[string]string{
					"install": "npm install",
					"test":    "npm test",
				},
			},
		}

		// Register detectors
		detector.RegisterDetector(goPlugin)
		detector.RegisterDetector(nodePlugin)

		// Detect frameworks
		results, err := detector.DetectFrameworks("/test/path")
		require.NoError(t, err)
		assert.Len(t, results, 2)

		// Check results
		var foundGo, foundNode bool
		for _, r := range results {
			if r.Framework.Name == "go" {
				foundGo = true
				assert.Equal(t, 90, r.Confidence)
			}
			if r.Framework.Name == "node" {
				foundNode = true
				assert.Equal(t, 85, r.Confidence)
			}
		}
		assert.True(t, foundGo)
		assert.True(t, foundNode)
	})

	t.Run("conflict resolution keeps highest confidence", func(t *testing.T) {
		detector := NewFrameworkDetector()

		// Two plugins detecting same framework
		plugin1 := &MockFrameworkDetector{
			result: &sdk.DetectionResult{
				Detected:   true,
				Confidence: 70,
				Framework: sdk.FrameworkInfo{
					Name: "python",
				},
			},
		}

		plugin2 := &MockFrameworkDetector{
			result: &sdk.DetectionResult{
				Detected:   true,
				Confidence: 90,
				Framework: sdk.FrameworkInfo{
					Name: "python",
				},
			},
		}

		detector.RegisterDetector(plugin1)
		detector.RegisterDetector(plugin2)

		results, err := detector.DetectFrameworks("/test/path")
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "python", results[0].Framework.Name)
		assert.Equal(t, 90, results[0].Confidence)
	})

	t.Run("timeout handling for slow plugins", func(t *testing.T) {
		detector := NewFrameworkDetector()

		// Fast plugin
		fastPlugin := &MockFrameworkDetector{
			result: &sdk.DetectionResult{
				Detected: true,
				Framework: sdk.FrameworkInfo{
					Name: "fast",
				},
				Confidence: 80,
			},
			delay: 10 * time.Millisecond,
		}

		// Slow plugin (will timeout)
		slowPlugin := &MockFrameworkDetector{
			result: &sdk.DetectionResult{
				Detected: true,
				Framework: sdk.FrameworkInfo{
					Name: "slow",
				},
				Confidence: 80,
			},
			delay: 200 * time.Millisecond, // Exceeds 100ms timeout
		}

		detector.RegisterDetector(fastPlugin)
		detector.RegisterDetector(slowPlugin)

		results, err := detector.DetectFrameworks("/test/path")
		require.NoError(t, err)

		// Should only get fast plugin result
		assert.Len(t, results, 1)
		assert.Equal(t, "fast", results[0].Framework.Name)
	})

	t.Run("caching detection results", func(t *testing.T) {
		detector := NewFrameworkDetector()

		plugin := &MockFrameworkDetector{
			result: &sdk.DetectionResult{
				Detected: true,
				Framework: sdk.FrameworkInfo{
					Name: "cached",
				},
				Confidence: 80,
			},
		}

		detector.RegisterDetector(plugin)

		// First call
		results1, err := detector.DetectFrameworks("/test/path")
		require.NoError(t, err)
		assert.Len(t, results1, 1)
		assert.Equal(t, 1, plugin.detectCallCount)

		// Second call should use cache
		results2, err := detector.DetectFrameworks("/test/path")
		require.NoError(t, err)
		assert.Len(t, results2, 1)
		assert.Equal(t, 1, plugin.detectCallCount) // No additional call

		// Different path should not use cache
		results3, err := detector.DetectFrameworks("/other/path")
		require.NoError(t, err)
		assert.Len(t, results3, 1)
		assert.Equal(t, 2, plugin.detectCallCount) // New call for different path
	})

	t.Run("get framework commands", func(t *testing.T) {
		detector := NewFrameworkDetector()

		plugin := &MockFrameworkDetector{
			result: &sdk.DetectionResult{
				Detected:   true,
				Confidence: 80,
				Framework: sdk.FrameworkInfo{
					Name: "test",
				},
				Commands: map[string]string{
					"build": "make build",
					"test":  "make test",
					"clean": "make clean",
				},
			},
		}

		detector.RegisterDetector(plugin)

		commands := detector.GetFrameworkCommands("/test/path")
		assert.Len(t, commands, 3)
		assert.Equal(t, "make build", commands["build"].Cmd)
		assert.Equal(t, "make test", commands["test"].Cmd)
		assert.Equal(t, "make clean", commands["clean"].Cmd)
	})

	t.Run("cache invalidation", func(t *testing.T) {
		detector := NewFrameworkDetector()

		plugin := &MockFrameworkDetector{
			result: &sdk.DetectionResult{
				Detected: true,
				Framework: sdk.FrameworkInfo{
					Name: "test",
				},
				Confidence: 80,
			},
		}

		detector.RegisterDetector(plugin)

		// First call
		_, err := detector.DetectFrameworks("/test/path")
		require.NoError(t, err)
		assert.Equal(t, 1, plugin.detectCallCount)

		// Second call uses cache
		_, err = detector.DetectFrameworks("/test/path")
		require.NoError(t, err)
		assert.Equal(t, 1, plugin.detectCallCount)

		// Invalidate cache
		detector.InvalidateCache("/test/path")

		// Third call should not use cache
		_, err = detector.DetectFrameworks("/test/path")
		require.NoError(t, err)
		assert.Equal(t, 2, plugin.detectCallCount)
	})

	t.Run("clear all cache", func(t *testing.T) {
		detector := NewFrameworkDetector()

		plugin := &MockFrameworkDetector{
			result: &sdk.DetectionResult{
				Detected: true,
				Framework: sdk.FrameworkInfo{
					Name: "test",
				},
				Confidence: 80,
			},
		}

		detector.RegisterDetector(plugin)

		// Cache multiple paths
		_, _ = detector.DetectFrameworks("/path1")
		_, _ = detector.DetectFrameworks("/path2")

		// Clear all cache
		detector.ClearCache()

		// Verify cache is empty (would need to expose cache for testing)
		// For now, just verify the method doesn't error
		assert.NotPanics(t, func() {
			detector.ClearCache()
		})
	})
}
