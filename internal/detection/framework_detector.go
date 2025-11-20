package detection

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/ivannovak/glide/pkg/plugin/sdk"
)

// FrameworkDetector aggregates all plugin detections
type FrameworkDetector struct {
	mu        sync.RWMutex
	detectors []sdk.FrameworkDetector
	cache     map[string]*DetectionCache
}

// DetectionCache caches detection results
type DetectionCache struct {
	ProjectPath string
	Timestamp   time.Time
	Results     []sdk.DetectionResult
	TTL         time.Duration
}

// FrameworkResult combines detection result with plugin info
type FrameworkResult struct {
	sdk.DetectionResult
	PluginName string
	Priority   int
}

// NewFrameworkDetector creates a new framework detector
func NewFrameworkDetector() *FrameworkDetector {
	return &FrameworkDetector{
		cache: make(map[string]*DetectionCache),
	}
}

// RegisterDetector registers a detector for framework detection
func (fd *FrameworkDetector) RegisterDetector(d sdk.FrameworkDetector) {
	fd.mu.Lock()
	defer fd.mu.Unlock()
	fd.detectors = append(fd.detectors, d)
}

// DetectFrameworks runs all plugin detections in parallel
func (fd *FrameworkDetector) DetectFrameworks(projectPath string) ([]FrameworkResult, error) {
	// Check cache first
	if cached := fd.getFromCache(projectPath); cached != nil {
		return fd.convertToFrameworkResults(cached), nil
	}

	results := make(chan FrameworkResult, len(fd.detectors))
	var wg sync.WaitGroup

	// Parallel detection with timeout
	timeout := 100 * time.Millisecond
	for i, d := range fd.detectors {
		wg.Add(1)
		go func(detector sdk.FrameworkDetector, priority int) {
			defer wg.Done()

			done := make(chan *sdk.DetectionResult, 1)
			go func() {
				result, err := detector.Detect(projectPath)
				if err == nil && result != nil {
					done <- result
				}
				close(done)
			}()

			select {
			case result := <-done:
				if result != nil && result.Detected {
					results <- FrameworkResult{
						DetectionResult: *result,
						PluginName:      fmt.Sprintf("detector-%d", priority),
						Priority:        priority,
					}
				}
			case <-time.After(timeout):
				// Timeout, skip this detector
			}
		}(d, i)
	}

	// Wait for all detections to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var frameworks []FrameworkResult
	for result := range results {
		frameworks = append(frameworks, result)
	}

	// Resolve conflicts and cache
	resolved := fd.resolveConflicts(frameworks)
	fd.cacheResults(projectPath, resolved)

	return resolved, nil
}

// resolveConflicts handles multiple plugins detecting same framework
func (fd *FrameworkDetector) resolveConflicts(results []FrameworkResult) []FrameworkResult {
	if len(results) == 0 {
		return results
	}

	// Sort by confidence score and priority
	sort.Slice(results, func(i, j int) bool {
		if results[i].Confidence == results[j].Confidence {
			return results[i].Priority < results[j].Priority
		}
		return results[i].Confidence > results[j].Confidence
	})

	// Remove duplicates, keeping highest confidence
	seen := make(map[string]bool)
	filtered := []FrameworkResult{}

	for _, r := range results {
		if !seen[r.Framework.Name] {
			seen[r.Framework.Name] = true
			filtered = append(filtered, r)
		}
	}

	return filtered
}

// GetFrameworkCommands returns aggregated commands from all detected frameworks
func (fd *FrameworkDetector) GetFrameworkCommands(projectPath string) map[string]sdk.CommandDefinition {
	frameworks, _ := fd.DetectFrameworks(projectPath)
	commands := make(map[string]sdk.CommandDefinition)

	// Aggregate commands from all frameworks
	for _, fw := range frameworks {
		// Get full command definitions from the plugin
		if len(fw.Commands) > 0 {
			for name, cmd := range fw.Commands {
				// Simple command to CommandDefinition conversion
				// In real implementation, would get from plugin
				commands[name] = sdk.CommandDefinition{
					Cmd:         cmd,
					Description: fmt.Sprintf("%s command from %s", name, fw.Framework.Name),
					Category:    "framework",
				}
			}
		}
	}

	return commands
}

// GetDetectedFrameworks returns framework names and versions for context enhancement
func (fd *FrameworkDetector) GetDetectedFrameworks(projectPath string) ([]string, map[string]string, error) {
	frameworks, err := fd.DetectFrameworks(projectPath)
	if err != nil {
		return nil, nil, err
	}

	// Add detected frameworks to lists
	var frameworkNames []string
	frameworkVersions := make(map[string]string)

	for _, fw := range frameworks {
		frameworkNames = append(frameworkNames, fw.Framework.Name)
		if fw.Framework.Version != "" {
			frameworkVersions[fw.Framework.Name] = fw.Framework.Version
		}
	}

	return frameworkNames, frameworkVersions, nil
}

// Cache management

func (fd *FrameworkDetector) getFromCache(projectPath string) []sdk.DetectionResult {
	fd.mu.RLock()
	defer fd.mu.RUnlock()

	cache, exists := fd.cache[projectPath]
	if !exists {
		return nil
	}

	// Check if cache is still valid (5 minute TTL)
	if time.Since(cache.Timestamp) > 5*time.Minute {
		return nil
	}

	return cache.Results
}

func (fd *FrameworkDetector) cacheResults(projectPath string, results []FrameworkResult) {
	fd.mu.Lock()
	defer fd.mu.Unlock()

	// Convert to DetectionResult for caching
	detectionResults := make([]sdk.DetectionResult, len(results))
	for i, r := range results {
		detectionResults[i] = r.DetectionResult
	}

	fd.cache[projectPath] = &DetectionCache{
		ProjectPath: projectPath,
		Timestamp:   time.Now(),
		Results:     detectionResults,
		TTL:         5 * time.Minute,
	}
}

func (fd *FrameworkDetector) convertToFrameworkResults(results []sdk.DetectionResult) []FrameworkResult {
	frameworks := make([]FrameworkResult, len(results))
	for i, r := range results {
		frameworks[i] = FrameworkResult{
			DetectionResult: r,
			PluginName:      "cached",
			Priority:        i,
		}
	}
	return frameworks
}

// ClearCache clears the detection cache
func (fd *FrameworkDetector) ClearCache() {
	fd.mu.Lock()
	defer fd.mu.Unlock()
	fd.cache = make(map[string]*DetectionCache)
}

// InvalidateCache invalidates cache for a specific project
func (fd *FrameworkDetector) InvalidateCache(projectPath string) {
	fd.mu.Lock()
	defer fd.mu.Unlock()
	delete(fd.cache, projectPath)
}
