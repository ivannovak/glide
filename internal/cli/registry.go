package cli

import (
	"sync"

	"github.com/ivannovak/glide/pkg/registry"
	"github.com/spf13/cobra"
)

// Factory is a function that creates a cobra command
type Factory func() *cobra.Command

// Registry manages command registration and creation
type Registry struct {
	*registry.Registry[Factory]
	metaMu   sync.RWMutex
	metadata map[string]Metadata
}

// Metadata holds metadata about a command
type Metadata struct {
	Name        string
	Category    Category
	Description string
	Aliases     []string // Alternative names for the command
	Hidden      bool     // For debug commands
}

// Category represents the category of a command
type Category string

const (
	CategoryCore      Category = "core"
	CategoryDocker    Category = "docker"
	CategoryTesting   Category = "testing"
	CategoryDatabase  Category = "database"
	CategoryDeveloper Category = "developer"
	CategorySetup     Category = "setup"
	CategoryPlugin    Category = "plugin"
	CategoryGlobal    Category = "global"
	CategoryDebug     Category = "debug"
	CategoryHelp      Category = "help"
)

// NewRegistry creates a new command registry
func NewRegistry() *Registry {
	return &Registry{
		Registry: registry.New[Factory](),
		metadata: make(map[string]Metadata),
	}
}

// Register adds a command factory to the registry
func (r *Registry) Register(name string, factory Factory, metadata Metadata) error {
	r.metaMu.Lock()
	r.metadata[name] = metadata
	r.metaMu.Unlock()
	return r.Registry.Register(name, factory, metadata.Aliases...)
}

// GetMetadata retrieves command metadata by name or alias
func (r *Registry) GetMetadata(name string) (Metadata, bool) {
	// Resolve alias if necessary
	if canonicalName, isAlias := r.ResolveAlias(name); isAlias {
		name = canonicalName
	}

	r.metaMu.RLock()
	meta, ok := r.metadata[name]
	r.metaMu.RUnlock()
	return meta, ok
}

// GetByCategory returns all commands in a specific category
func (r *Registry) GetByCategory(category Category) []string {
	var commands []string
	r.metaMu.RLock()
	for name, meta := range r.metadata {
		if meta.Category == category {
			commands = append(commands, name)
		}
	}
	r.metaMu.RUnlock()
	return commands
}

// CreateAll creates all registered commands
func (r *Registry) CreateAll() []*cobra.Command {
	var commands []*cobra.Command

	r.ForEach(func(name string, factory Factory) {
		cmd := factory()

		// Set metadata on the cobra command
		r.metaMu.RLock()
		if meta, ok := r.metadata[name]; ok {
			// Set aliases if they exist
			if len(meta.Aliases) > 0 {
				cmd.Aliases = meta.Aliases
			}

			// Store category in annotations for help system
			if cmd.Annotations == nil {
				cmd.Annotations = make(map[string]string)
			}
			cmd.Annotations["category"] = string(meta.Category)

			// Mark hidden commands
			if meta.Hidden {
				cmd.Hidden = true
			}
		}
		r.metaMu.RUnlock()

		commands = append(commands, cmd)
	})

	return commands
}

// CreateByCategory creates all commands in a specific category
func (r *Registry) CreateByCategory(category Category) []*cobra.Command {
	var commands []*cobra.Command

	for _, name := range r.GetByCategory(category) {
		if factory, ok := r.Get(name); ok {
			cmd := factory()

			// Set metadata on the cobra command
			r.metaMu.RLock()
			if meta, ok := r.metadata[name]; ok {
				if len(meta.Aliases) > 0 {
					cmd.Aliases = meta.Aliases
				}

				if meta.Hidden {
					cmd.Hidden = true
				}
			}
			r.metaMu.RUnlock()

			commands = append(commands, cmd)
		}
	}

	return commands
}
