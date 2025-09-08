package cli

import (
	"fmt"
	"github.com/spf13/cobra"
)

// Factory is a function that creates a cobra command
type Factory func() *cobra.Command

// Registry manages command registration and creation
type Registry struct {
	factories map[string]Factory
	metadata  map[string]Metadata
	aliases   map[string]string // maps alias to canonical command name
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
		factories: make(map[string]Factory),
		metadata:  make(map[string]Metadata),
		aliases:   make(map[string]string),
	}
}

// Register adds a command factory to the registry
func (r *Registry) Register(name string, factory Factory, metadata Metadata) error {
	// Check if command name already exists
	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("command %s already registered", name)
	}

	// Check if command name conflicts with existing alias
	if _, exists := r.aliases[name]; exists {
		return fmt.Errorf("command name %s conflicts with existing alias", name)
	}

	// Check if any aliases conflict with existing commands or aliases
	for _, alias := range metadata.Aliases {
		if _, exists := r.factories[alias]; exists {
			return fmt.Errorf("alias %s conflicts with existing command", alias)
		}
		if _, exists := r.aliases[alias]; exists {
			return fmt.Errorf("alias %s already registered", alias)
		}
	}

	r.factories[name] = factory
	r.metadata[name] = metadata

	// Register all aliases
	for _, alias := range metadata.Aliases {
		r.aliases[alias] = name
	}

	return nil
}

// Get retrieves a command factory by name or alias
func (r *Registry) Get(name string) (Factory, bool) {
	// Check if it's a direct command name
	factory, ok := r.factories[name]
	if ok {
		return factory, true
	}

	// Check if it's an alias
	if canonicalName, isAlias := r.aliases[name]; isAlias {
		return r.factories[canonicalName], true
	}

	return nil, false
}

// GetMetadata retrieves command metadata by name or alias
func (r *Registry) GetMetadata(name string) (Metadata, bool) {
	// Check if it's a direct command name
	meta, ok := r.metadata[name]
	if ok {
		return meta, true
	}

	// Check if it's an alias
	if canonicalName, isAlias := r.aliases[name]; isAlias {
		return r.metadata[canonicalName], true
	}

	return Metadata{}, false
}

// GetByCategory returns all commands in a specific category
func (r *Registry) GetByCategory(category Category) []string {
	var commands []string
	for name, meta := range r.metadata {
		if meta.Category == category {
			commands = append(commands, name)
		}
	}
	return commands
}

// CreateAll creates all registered commands
func (r *Registry) CreateAll() []*cobra.Command {
	var commands []*cobra.Command
	for name, factory := range r.factories {
		cmd := factory()
		// Set metadata on the cobra command
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
		commands = append(commands, cmd)
	}
	return commands
}

// CreateByCategory creates all commands in a specific category
func (r *Registry) CreateByCategory(category Category) []*cobra.Command {
	var commands []*cobra.Command
	for name, meta := range r.metadata {
		if meta.Category == category {
			if factory, ok := r.factories[name]; ok {
				cmd := factory()
				// Set aliases on the cobra command if they exist
				if len(meta.Aliases) > 0 {
					cmd.Aliases = meta.Aliases
				}
				commands = append(commands, cmd)
			}
		}
	}
	return commands
}

// ResolveAlias resolves an alias to its canonical command name
func (r *Registry) ResolveAlias(alias string) (string, bool) {
	canonical, ok := r.aliases[alias]
	return canonical, ok
}

// GetAliases returns all aliases for a given command
func (r *Registry) GetAliases(commandName string) []string {
	meta, ok := r.metadata[commandName]
	if !ok {
		return nil
	}
	return meta.Aliases
}

// IsAlias checks if a given name is an alias
func (r *Registry) IsAlias(name string) bool {
	_, ok := r.aliases[name]
	return ok
}
