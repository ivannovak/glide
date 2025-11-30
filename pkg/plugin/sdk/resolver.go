package sdk

import (
	"log"
	"sort"
)

// DependencyResolver resolves plugin dependencies and determines load order.
//
// The resolver performs:
//   - Topological sorting to determine correct load order
//   - Cycle detection to prevent circular dependencies
//   - Version constraint validation
//   - Missing dependency detection
//
// Example usage:
//
//	resolver := NewDependencyResolver()
//	loadOrder, err := resolver.Resolve(plugins)
//	if err != nil {
//	    // Handle dependency errors
//	}
//	// Load plugins in the determined order
//	for _, name := range loadOrder {
//	    // Load plugin by name
//	}
type DependencyResolver struct{}

// NewDependencyResolver creates a new dependency resolver.
func NewDependencyResolver() *DependencyResolver {
	return &DependencyResolver{}
}

// Resolve determines the correct load order for plugins based on their dependencies.
//
// Returns:
//   - A slice of plugin names in dependency order (dependencies before dependents)
//   - An error if there are circular dependencies, missing required dependencies,
//     or version mismatches
//
// Algorithm:
//  1. Build dependency graph
//  2. Validate all dependencies (existence, versions)
//  3. Perform topological sort with cycle detection
//  4. Return sorted order
func (r *DependencyResolver) Resolve(plugins map[string]PluginMetadata) ([]string, error) {
	// Build dependency graph
	graph := NewDependencyGraph()
	for name, metadata := range plugins {
		graph.AddPlugin(name, metadata.Dependencies)
	}

	// Validate all dependencies
	if err := r.validateDependencies(plugins, graph); err != nil {
		return nil, err
	}

	// Perform topological sort
	sorted, err := r.topologicalSort(graph)
	if err != nil {
		return nil, err
	}

	return sorted, nil
}

// validateDependencies checks all plugin dependencies for:
//   - Existence (dependency plugin is available)
//   - Version compatibility (dependency version satisfies constraint)
//
// Returns first error encountered, or nil if all dependencies are valid.
func (r *DependencyResolver) validateDependencies(plugins map[string]PluginMetadata, graph *DependencyGraph) error {
	for pluginName, metadata := range plugins {
		for _, dep := range metadata.Dependencies {
			// Validate dependency declaration
			if err := dep.Validate(); err != nil {
				return NewDependencyError(pluginName, "invalid dependency declaration", err)
			}

			// Check if dependency exists
			depMetadata, exists := plugins[dep.Name]
			if !exists {
				if dep.Optional {
					log.Printf("Warning: Plugin %q has optional dependency %q which is not available", pluginName, dep.Name)
					continue
				}
				return &MissingDependencyError{
					Plugin:     pluginName,
					Dependency: dep,
				}
			}

			// Check version compatibility
			if !dep.SatisfiedBy(depMetadata.Version) {
				if dep.Optional {
					log.Printf(
						"Warning: Plugin %q has optional dependency %s but found version %s",
						pluginName,
						dep,
						depMetadata.Version,
					)
					continue
				}
				return &VersionMismatchError{
					Plugin:          pluginName,
					Dependency:      dep,
					ActualVersion:   depMetadata.Version,
					RequiredVersion: dep.Version,
				}
			}
		}
	}
	return nil
}

// topologicalSort performs topological sorting using Kahn's algorithm.
//
// Algorithm:
//  1. Find all nodes with no incoming edges (no dependencies)
//  2. Add them to the result
//  3. Remove their outgoing edges
//  4. Repeat until all nodes are processed or a cycle is detected
//
// Returns:
//   - Sorted slice of plugin names
//   - CyclicDependencyError if a cycle is detected
func (r *DependencyResolver) topologicalSort(graph *DependencyGraph) ([]string, error) {
	// Calculate in-degree for each node (number of dependencies)
	inDegree := make(map[string]int)
	allPlugins := graph.AllPlugins()

	for _, plugin := range allPlugins {
		if _, exists := inDegree[plugin]; !exists {
			inDegree[plugin] = 0
		}

		// Count dependencies for each plugin
		for _, dep := range graph.GetDependencies(plugin) {
			// Only count non-optional or existing dependencies
			if graph.HasPlugin(dep.Name) {
				inDegree[plugin]++
			}
		}
	}

	// Find all nodes with no dependencies (in-degree = 0)
	queue := make([]string, 0)
	for plugin, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, plugin)
		}
	}

	// Sort queue for deterministic ordering
	sort.Strings(queue)

	result := make([]string, 0, len(allPlugins))
	processed := make(map[string]bool)

	for len(queue) > 0 {
		// Process next plugin with no dependencies
		current := queue[0]
		queue = queue[1:]

		result = append(result, current)
		processed[current] = true

		// Find plugins that depend on current plugin
		dependents := r.findDependents(graph, current)

		// Decrease in-degree for each dependent
		for _, dependent := range dependents {
			if processed[dependent] {
				continue
			}

			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
				// Keep queue sorted for deterministic ordering
				sort.Strings(queue)
			}
		}
	}

	// If not all plugins were processed, there's a cycle
	if len(result) != len(allPlugins) {
		cycle := r.findCycle(graph, processed)
		return nil, &CyclicDependencyError{Cycle: cycle}
	}

	return result, nil
}

// findDependents returns all plugins that depend on the given plugin.
func (r *DependencyResolver) findDependents(graph *DependencyGraph, plugin string) []string {
	dependents := make([]string, 0)

	for _, p := range graph.AllPlugins() {
		for _, dep := range graph.GetDependencies(p) {
			if dep.Name == plugin && graph.HasPlugin(dep.Name) {
				dependents = append(dependents, p)
				break
			}
		}
	}

	return dependents
}

// findCycle detects and returns a cycle in the dependency graph.
//
// Uses DFS to find a cycle among unprocessed plugins.
func (r *DependencyResolver) findCycle(graph *DependencyGraph, processed map[string]bool) []string {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	// Find an unprocessed plugin to start DFS
	var startPlugin string
	for _, plugin := range graph.AllPlugins() {
		if !processed[plugin] {
			startPlugin = plugin
			break
		}
	}

	if startPlugin == "" {
		return []string{"unknown cycle"}
	}

	// Perform DFS to find cycle
	var cycle []string
	r.dfsFindCycle(graph, startPlugin, visited, recStack, &cycle)

	if len(cycle) == 0 {
		return []string{startPlugin, "cycle detection failed"}
	}

	return cycle
}

// dfsFindCycle performs depth-first search to find a cycle.
func (r *DependencyResolver) dfsFindCycle(
	graph *DependencyGraph,
	plugin string,
	visited map[string]bool,
	recStack map[string]bool,
	cycle *[]string,
) bool {
	visited[plugin] = true
	recStack[plugin] = true

	for _, dep := range graph.GetDependencies(plugin) {
		if !graph.HasPlugin(dep.Name) {
			continue
		}

		if !visited[dep.Name] {
			if r.dfsFindCycle(graph, dep.Name, visited, recStack, cycle) {
				*cycle = append(*cycle, plugin)
				return true
			}
		} else if recStack[dep.Name] {
			// Found a cycle
			*cycle = append(*cycle, dep.Name, plugin)
			return true
		}
	}

	recStack[plugin] = false
	return false
}

// ValidatePluginDependencies validates a single plugin's dependencies.
//
// This is a convenience method for validating individual plugins
// without resolving the entire dependency graph.
func (r *DependencyResolver) ValidatePluginDependencies(
	plugin PluginMetadata,
	availablePlugins map[string]PluginMetadata,
) error {
	for _, dep := range plugin.Dependencies {
		// Validate dependency declaration
		if err := dep.Validate(); err != nil {
			return NewDependencyError(plugin.Name, "invalid dependency declaration", err)
		}

		// Check if dependency exists
		depMetadata, exists := availablePlugins[dep.Name]
		if !exists {
			if dep.Optional {
				log.Printf("Warning: Plugin %q has optional dependency %q which is not available", plugin.Name, dep.Name)
				continue
			}
			return &MissingDependencyError{
				Plugin:     plugin.Name,
				Dependency: dep,
			}
		}

		// Check version compatibility
		if !dep.SatisfiedBy(depMetadata.Version) {
			if dep.Optional {
				log.Printf(
					"Warning: Plugin %q has optional dependency %s but found version %s",
					plugin.Name,
					dep,
					depMetadata.Version,
				)
				continue
			}
			return &VersionMismatchError{
				Plugin:          plugin.Name,
				Dependency:      dep,
				ActualVersion:   depMetadata.Version,
				RequiredVersion: dep.Version,
			}
		}
	}

	return nil
}

// DependencyInfo provides information about resolved dependencies.
type DependencyInfo struct {
	// LoadOrder is the order in which plugins should be loaded
	LoadOrder []string

	// DirectDependencies maps each plugin to its direct dependencies
	DirectDependencies map[string][]PluginDependency

	// AllDependencies maps each plugin to all transitive dependencies
	AllDependencies map[string][]string
}

// GetDependencyInfo returns detailed dependency information.
func (r *DependencyResolver) GetDependencyInfo(plugins map[string]PluginMetadata) (*DependencyInfo, error) {
	loadOrder, err := r.Resolve(plugins)
	if err != nil {
		return nil, err
	}

	info := &DependencyInfo{
		LoadOrder:          loadOrder,
		DirectDependencies: make(map[string][]PluginDependency),
		AllDependencies:    make(map[string][]string),
	}

	// Build direct dependencies map
	for name, metadata := range plugins {
		info.DirectDependencies[name] = metadata.Dependencies
	}

	// Build transitive dependencies map
	for _, plugin := range loadOrder {
		info.AllDependencies[plugin] = r.getTransitiveDependencies(plugin, plugins, make(map[string]bool))
	}

	return info, nil
}

// getTransitiveDependencies recursively finds all transitive dependencies.
func (r *DependencyResolver) getTransitiveDependencies(
	plugin string,
	plugins map[string]PluginMetadata,
	visited map[string]bool,
) []string {
	if visited[plugin] {
		return []string{}
	}
	visited[plugin] = true

	deps := make([]string, 0)
	metadata, exists := plugins[plugin]
	if !exists {
		return deps
	}

	for _, dep := range metadata.Dependencies {
		if _, exists := plugins[dep.Name]; !exists {
			continue // Skip missing optional dependencies
		}

		deps = append(deps, dep.Name)
		// Add transitive dependencies
		transitive := r.getTransitiveDependencies(dep.Name, plugins, visited)
		deps = append(deps, transitive...)
	}

	return deps
}
