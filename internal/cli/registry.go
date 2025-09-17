package cli

import (
	"sync"

	"github.com/ivannovak/glide/internal/config"
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
	CategoryProject   Category = "project"
	CategoryDebug     Category = "debug"
	CategoryHelp      Category = "help"
	CategoryYAML      Category = "yaml" // User-defined YAML commands
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

// AddYAMLCommand registers a YAML-defined command
func (r *Registry) AddYAMLCommand(name string, cmd *config.Command) error {
	// Create a factory that builds a cobra command from the YAML definition
	factory := func() *cobra.Command {
		cobraCmd := &cobra.Command{
			Use:   name,
			Short: cmd.Description,
			Long:  cmd.Help,
			RunE: func(c *cobra.Command, args []string) error {
				// Execute the YAML-defined command
				return ExecuteYAMLCommand(cmd.Cmd, args)
			},
		}

		// Set alias if defined
		if cmd.Alias != "" {
			cobraCmd.Aliases = []string{cmd.Alias}
		}

		// Allow arbitrary args to be passed through
		cobraCmd.DisableFlagParsing = true

		return cobraCmd
	}

	// Determine category
	category := CategoryYAML
	if cmd.Category != "" {
		// Try to map to existing category
		switch cmd.Category {
		case "core":
			category = CategoryCore
		case "docker":
			category = CategoryDocker
		case "testing":
			category = CategoryTesting
		case "database":
			category = CategoryDatabase
		case "developer":
			category = CategoryDeveloper
		case "setup":
			category = CategorySetup
		case "plugin":
			category = CategoryPlugin
		case "global":
			category = CategoryGlobal
		case "debug":
			category = CategoryDebug
		case "help":
			category = CategoryHelp
		default:
			category = CategoryYAML
		}
	}

	metadata := Metadata{
		Name:        name,
		Category:    category,
		Description: cmd.Description,
	}

	// Add alias to metadata if defined
	if cmd.Alias != "" {
		metadata.Aliases = []string{cmd.Alias}
	}

	return r.Register(name, factory, metadata)
}
