package output

import (
	"fmt"
	"io"
	"sync"
)

// Factory is a function that creates a Formatter
type Factory func(w io.Writer, noColor, quiet bool) Formatter

// Registry manages formatter registration and creation
type Registry struct {
	mu        sync.RWMutex
	factories map[Format]Factory
}

// globalRegistry is the default registry instance
var globalRegistry = NewRegistry()

// NewRegistry creates a new formatter registry
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[Format]Factory),
	}
}

// Register adds a formatter factory to the registry
func (r *Registry) Register(format Format, factory Factory) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.factories[format]; exists {
		return fmt.Errorf("formatter %s already registered", format)
	}
	
	r.factories[format] = factory
	return nil
}

// Create creates a formatter instance for the given format
func (r *Registry) Create(format Format, w io.Writer, noColor, quiet bool) (Formatter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	factory, ok := r.factories[format]
	if !ok {
		return nil, fmt.Errorf("unknown format: %s", format)
	}
	
	return factory(w, noColor, quiet), nil
}

// IsRegistered checks if a format is registered
func (r *Registry) IsRegistered(format Format) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	_, ok := r.factories[format]
	return ok
}

// GetFormats returns all registered formats
func (r *Registry) GetFormats() []Format {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	formats := make([]Format, 0, len(r.factories))
	for format := range r.factories {
		formats = append(formats, format)
	}
	return formats
}

// InitDefaultRegistry initializes the default formatter registry
func InitDefaultRegistry() {
	// Register all built-in formatters
	globalRegistry.Register(FormatJSON, func(w io.Writer, noColor, quiet bool) Formatter {
		return NewJSONFormatter(w, noColor, quiet)
	})
	
	globalRegistry.Register(FormatYAML, func(w io.Writer, noColor, quiet bool) Formatter {
		return NewYAMLFormatter(w, noColor, quiet)
	})
	
	globalRegistry.Register(FormatTable, func(w io.Writer, noColor, quiet bool) Formatter {
		return NewTableFormatter(w, noColor, quiet)
	})
	
	globalRegistry.Register(FormatPlain, func(w io.Writer, noColor, quiet bool) Formatter {
		return NewPlainFormatter(w, noColor, quiet)
	})
}

// GetGlobalRegistry returns the global formatter registry
func GetGlobalRegistry() *Registry {
	return globalRegistry
}

// RegisterFormatter registers a formatter in the global registry
func RegisterFormatter(format Format, factory Factory) error {
	return globalRegistry.Register(format, factory)
}

// CreateFormatter creates a formatter from the global registry
func CreateFormatter(format Format, w io.Writer, noColor, quiet bool) (Formatter, error) {
	return globalRegistry.Create(format, w, noColor, quiet)
}

// init initializes the default formatters
func init() {
	InitDefaultRegistry()
}