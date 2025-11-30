package sdk

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

// PluginMetadata represents plugin metadata for dependency resolution.
// This is a simplified version used internally by the dependency resolver.
type PluginMetadata struct {
	Name         string
	Version      string
	Author       string
	Description  string
	Dependencies []PluginDependency
}

// PluginDependency represents a dependency on another plugin.
//
// Dependencies are declared by plugins to ensure proper load order
// and to detect missing or incompatible plugin versions.
//
// Version constraints follow semantic versioning (semver) format:
//   - Exact: "1.2.3"
//   - Range: ">=1.0.0 <2.0.0"
//   - Wildcard: "1.x", "1.2.x"
//   - Caret: "^1.2.3" (>=1.2.3 <2.0.0)
//   - Tilde: "~1.2.3" (>=1.2.3 <1.3.0)
//
// Example:
//
//	deps := []PluginDependency{
//	    {Name: "docker", Version: "^1.0.0", Optional: false},
//	    {Name: "node", Version: ">=14.0.0", Optional: true},
//	}
type PluginDependency struct {
	// Name is the plugin name (must match plugin's Name() method)
	Name string `json:"name" yaml:"name"`

	// Version is a semver constraint (e.g., "^1.0.0", ">=1.2.3 <2.0.0")
	Version string `json:"version" yaml:"version"`

	// Optional indicates whether this dependency is required
	// Required dependencies cause load failure if missing/incompatible
	// Optional dependencies log warnings but allow loading to continue
	Optional bool `json:"optional" yaml:"optional"`
}

// String returns a human-readable representation of the dependency.
func (d PluginDependency) String() string {
	opt := ""
	if d.Optional {
		opt = " (optional)"
	}
	return fmt.Sprintf("%s@%s%s", d.Name, d.Version, opt)
}

// Validate checks if the dependency declaration is valid.
//
// Returns an error if:
//   - Name is empty
//   - Version is not a valid semver constraint
func (d PluginDependency) Validate() error {
	if d.Name == "" {
		return fmt.Errorf("dependency name cannot be empty")
	}

	if d.Version == "" {
		return fmt.Errorf("dependency %q: version constraint cannot be empty", d.Name)
	}

	// Validate semver constraint
	if _, err := semver.NewConstraint(d.Version); err != nil {
		return fmt.Errorf("dependency %q: invalid version constraint %q: %w", d.Name, d.Version, err)
	}

	return nil
}

// SatisfiedBy checks if a plugin version satisfies this dependency.
//
// Returns true if the provided version satisfies the version constraint.
// Returns false if the version doesn't satisfy the constraint or is invalid.
func (d PluginDependency) SatisfiedBy(version string) bool {
	constraint, err := semver.NewConstraint(d.Version)
	if err != nil {
		return false
	}

	v, err := semver.NewVersion(version)
	if err != nil {
		return false
	}

	return constraint.Check(v)
}

// DependencyGraph represents the dependency relationships between plugins.
//
// This is used by the dependency resolver to compute load order and
// detect cycles.
type DependencyGraph struct {
	// nodes maps plugin name to its dependencies
	nodes map[string][]PluginDependency
}

// NewDependencyGraph creates an empty dependency graph.
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		nodes: make(map[string][]PluginDependency),
	}
}

// AddPlugin adds a plugin and its dependencies to the graph.
func (g *DependencyGraph) AddPlugin(name string, dependencies []PluginDependency) {
	g.nodes[name] = dependencies
}

// GetDependencies returns the dependencies for a plugin.
func (g *DependencyGraph) GetDependencies(name string) []PluginDependency {
	return g.nodes[name]
}

// HasPlugin checks if a plugin exists in the graph.
func (g *DependencyGraph) HasPlugin(name string) bool {
	_, exists := g.nodes[name]
	return exists
}

// AllPlugins returns all plugin names in the graph.
func (g *DependencyGraph) AllPlugins() []string {
	plugins := make([]string, 0, len(g.nodes))
	for name := range g.nodes {
		plugins = append(plugins, name)
	}
	return plugins
}

// DependencyError represents an error related to plugin dependencies.
type DependencyError struct {
	Plugin  string
	Message string
	Cause   error
}

// Error implements the error interface.
func (e *DependencyError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("dependency error for plugin %q: %s: %v", e.Plugin, e.Message, e.Cause)
	}
	return fmt.Sprintf("dependency error for plugin %q: %s", e.Plugin, e.Message)
}

// Unwrap returns the underlying error.
func (e *DependencyError) Unwrap() error {
	return e.Cause
}

// NewDependencyError creates a new dependency error.
func NewDependencyError(plugin, message string, cause error) *DependencyError {
	return &DependencyError{
		Plugin:  plugin,
		Message: message,
		Cause:   cause,
	}
}

// CyclicDependencyError represents a cycle in the dependency graph.
type CyclicDependencyError struct {
	Cycle []string
}

// Error implements the error interface.
func (e *CyclicDependencyError) Error() string {
	return fmt.Sprintf("cyclic dependency detected: %v", e.Cycle)
}

// MissingDependencyError represents a missing required dependency.
type MissingDependencyError struct {
	Plugin     string
	Dependency PluginDependency
}

// Error implements the error interface.
func (e *MissingDependencyError) Error() string {
	return fmt.Sprintf("plugin %q requires missing dependency %s", e.Plugin, e.Dependency)
}

// VersionMismatchError represents an incompatible plugin version.
type VersionMismatchError struct {
	Plugin          string
	Dependency      PluginDependency
	ActualVersion   string
	RequiredVersion string
}

// Error implements the error interface.
func (e *VersionMismatchError) Error() string {
	return fmt.Sprintf(
		"plugin %q requires %s@%s but found version %s",
		e.Plugin,
		e.Dependency.Name,
		e.RequiredVersion,
		e.ActualVersion,
	)
}
