package plugin

import (
	"fmt"
	"strings"

	"github.com/ivannovak/glide/v2/pkg/logging"
	"github.com/ivannovak/glide/v2/pkg/registry"
	"github.com/spf13/cobra"
)

// PluginError represents an error that occurred during plugin loading
type PluginError struct {
	Name    string
	Error   error
	IsFatal bool
}

// PluginLoadResult contains the results of loading all plugins
type PluginLoadResult struct {
	Loaded   []string
	Failed   []PluginError
	Warnings []string
}

// HasErrors returns true if any plugins failed to load
func (r *PluginLoadResult) HasErrors() bool {
	return len(r.Failed) > 0
}

// HasFatalErrors returns true if any fatal errors occurred
func (r *PluginLoadResult) HasFatalErrors() bool {
	for _, err := range r.Failed {
		if err.IsFatal {
			return true
		}
	}
	return false
}

// ErrorMessage returns a formatted error message with all failures
func (r *PluginLoadResult) ErrorMessage() string {
	if !r.HasErrors() {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("Plugin loading issues:\n")

	for _, err := range r.Failed {
		severity := "warning"
		if err.IsFatal {
			severity = "FATAL"
		}
		builder.WriteString(fmt.Sprintf("  [%s] %s: %v\n", severity, err.Name, err.Error))
	}

	if len(r.Loaded) > 0 {
		builder.WriteString(fmt.Sprintf("\nSuccessfully loaded %d plugins: %s", len(r.Loaded), strings.Join(r.Loaded, ", ")))
	}

	return builder.String()
}

// Registry manages plugin registration and lifecycle.
//
// The Registry provides a centralized location for plugin management, including
// registration and command loading.
//
// NOTE: Plugin configuration is now handled by the pkg/config type-safe system.
// Plugins should register their typed configs using config.Register() in init().
type Registry struct {
	*registry.Registry[Plugin]
}

// global registry instance
var globalRegistry = NewRegistry()

// NewRegistry creates a new plugin registry
func NewRegistry() *Registry {
	return &Registry{
		Registry: registry.New[Plugin](),
	}
}

// Register adds a plugin to the global registry
func Register(p Plugin) error {
	return globalRegistry.RegisterPlugin(p)
}

// RegisterPlugin adds a plugin to the registry
func (r *Registry) RegisterPlugin(p Plugin) error {
	if p == nil {
		return fmt.Errorf("cannot register nil plugin")
	}

	name := p.Name()
	if name == "" {
		return fmt.Errorf("plugin must have a name")
	}

	// Get plugin metadata to register aliases
	meta := p.Metadata()

	// Use the generic registry's Register method with aliases
	return r.Registry.Register(name, p, meta.Aliases...)
}

// LoadAll registers all plugin commands
func (r *Registry) LoadAll(root *cobra.Command) (*PluginLoadResult, error) {
	logging.Debug("Loading all plugins")

	result := &PluginLoadResult{
		Loaded:   make([]string, 0),
		Failed:   make([]PluginError, 0),
		Warnings: make([]string, 0),
	}

	// Track if we encountered any fatal errors
	var fatalError error

	r.ForEach(func(name string, plugin Plugin) {
		logging.Debug("Loading plugin", "name", name)
		// If we already have a fatal error, skip remaining plugins
		if fatalError != nil {
			return
		}

		// NOTE: Plugin configuration is now handled via pkg/config type-safe registry.
		// Plugins access their typed config in Configure() using config.Get[T](name).
		if err := plugin.Configure(); err != nil {
			// Configuration errors are typically non-fatal
			// Log and continue with other plugins
			logging.Warn("Plugin configuration failed", "name", name, "error", err)
			result.Failed = append(result.Failed, PluginError{
				Name:    name,
				Error:   fmt.Errorf("failed to configure: %w", err),
				IsFatal: false,
			})
			return
		}

		// Register plugin commands
		if err := plugin.Register(root); err != nil {
			// Command registration errors are typically non-fatal
			// Log and continue with other plugins
			logging.Warn("Plugin command registration failed", "name", name, "error", err)
			result.Failed = append(result.Failed, PluginError{
				Name:    name,
				Error:   fmt.Errorf("failed to register commands: %w", err),
				IsFatal: false,
			})
			return
		}

		// Successfully loaded
		logging.Info("Plugin loaded successfully", "name", name)
		result.Loaded = append(result.Loaded, name)
	})

	// Return fatal error if encountered
	if fatalError != nil {
		logging.Error("Fatal error during plugin loading", "error", fatalError)
		return result, fatalError
	}

	logging.Info("All plugins loaded", "loaded", len(result.Loaded), "failed", len(result.Failed))
	return result, nil
}

// Global registry functions

// GetGlobalRegistry returns the global plugin registry
func GetGlobalRegistry() *Registry {
	return globalRegistry
}

// List returns all plugins from the global registry
func List() []Plugin {
	return globalRegistry.List()
}

// Get returns a plugin from the global registry
func Get(name string) (Plugin, bool) {
	return globalRegistry.Get(name)
}

// LoadAll loads all plugins from the global registry
func LoadAll(root *cobra.Command) (*PluginLoadResult, error) {
	return globalRegistry.LoadAll(root)
}
