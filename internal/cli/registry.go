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
}

// Metadata holds metadata about a command
type Metadata struct {
	Name        string
	Category    Category
	Description string
	Hidden      bool // For debug commands
}

// Category represents the category of a command
type Category string

const (
	CategoryCore      Category = "core"
	CategoryDocker    Category = "docker"
	CategoryDatabase  Category = "database"
	CategoryDeveloper Category = "developer"
	CategoryDebug     Category = "debug"
	CategoryHelp      Category = "help"
)

// NewRegistry creates a new command registry
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]Factory),
		metadata:  make(map[string]Metadata),
	}
}

// Register adds a command factory to the registry
func (r *Registry) Register(name string, factory Factory, metadata Metadata) error {
	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("command %s already registered", name)
	}
	
	r.factories[name] = factory
	r.metadata[name] = metadata
	return nil
}

// Get retrieves a command factory by name
func (r *Registry) Get(name string) (Factory, bool) {
	factory, ok := r.factories[name]
	return factory, ok
}

// GetMetadata retrieves command metadata by name
func (r *Registry) GetMetadata(name string) (Metadata, bool) {
	meta, ok := r.metadata[name]
	return meta, ok
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
	for _, factory := range r.factories {
		commands = append(commands, factory())
	}
	return commands
}

// CreateByCategory creates all commands in a specific category
func (r *Registry) CreateByCategory(category Category) []*cobra.Command {
	var commands []*cobra.Command
	for name, meta := range r.metadata {
		if meta.Category == category {
			if factory, ok := r.factories[name]; ok {
				commands = append(commands, factory())
			}
		}
	}
	return commands
}