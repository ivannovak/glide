package plugin

import (
	"fmt"
	"sync"

	"github.com/ivannovak/glide/pkg/registry"
	"github.com/spf13/cobra"
)

// Registry manages plugin registration and lifecycle
type Registry struct {
	*registry.Registry[Plugin]
	configMu sync.RWMutex
	config   map[string]interface{}
}

// global registry instance
var globalRegistry = NewRegistry()

// NewRegistry creates a new plugin registry
func NewRegistry() *Registry {
	return &Registry{
		Registry: registry.New[Plugin](),
		config:   make(map[string]interface{}),
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

// SetConfig sets the configuration for all plugins
func (r *Registry) SetConfig(config map[string]interface{}) {
	r.configMu.Lock()
	r.config = config
	r.configMu.Unlock()
}

// LoadAll configures and registers all plugin commands
func (r *Registry) LoadAll(root *cobra.Command) error {
	var errors []error

	// Get config once with lock
	r.configMu.RLock()
	config := r.config
	r.configMu.RUnlock()

	r.ForEach(func(name string, plugin Plugin) {
		// Configure the plugin
		if err := plugin.Configure(config); err != nil {
			errors = append(errors, fmt.Errorf("failed to configure plugin %s: %w", name, err))
			return
		}

		// Register plugin commands
		if err := plugin.Register(root); err != nil {
			errors = append(errors, fmt.Errorf("failed to register plugin %s: %w", name, err))
			return
		}
	})

	// Return aggregated errors
	if len(errors) > 0 {
		return fmt.Errorf("plugin loading errors: %v", errors)
	}
	return nil
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
func LoadAll(root *cobra.Command) error {
	return globalRegistry.LoadAll(root)
}

// SetConfig sets configuration for the global registry
func SetConfig(config map[string]interface{}) {
	globalRegistry.SetConfig(config)
}
