package registry

import (
	"fmt"
	"sort"
	"sync"
)

// Registry is a thread-safe generic registry for managing items with alias support
type Registry[T any] struct {
	mu      sync.RWMutex
	items   map[string]T
	aliases map[string]string // maps alias to canonical name
}

// New creates a new generic registry
func New[T any]() *Registry[T] {
	return &Registry[T]{
		items:   make(map[string]T),
		aliases: make(map[string]string),
	}
}

// Register adds an item to the registry with optional aliases
func (r *Registry[T]) Register(name string, item T, aliases ...string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if name == "" {
		return fmt.Errorf("item name cannot be empty")
	}

	// Check if name already exists
	if _, exists := r.items[name]; exists {
		return fmt.Errorf("item %s already registered", name)
	}

	// Check if name conflicts with existing alias
	if _, exists := r.aliases[name]; exists {
		return fmt.Errorf("item name %s conflicts with existing alias", name)
	}

	// Check for alias conflicts
	for _, alias := range aliases {
		if _, exists := r.items[alias]; exists {
			return fmt.Errorf("alias %s conflicts with existing item", alias)
		}
		if _, exists := r.aliases[alias]; exists {
			return fmt.Errorf("alias %s already registered", alias)
		}
	}

	// Register the item
	r.items[name] = item

	// Register all aliases
	for _, alias := range aliases {
		r.aliases[alias] = name
	}

	return nil
}

// Get retrieves an item by name or alias
func (r *Registry[T]) Get(name string) (T, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check if it's a direct item name
	item, exists := r.items[name]
	if exists {
		return item, true
	}

	// Check if it's an alias
	if canonicalName, isAlias := r.aliases[name]; isAlias {
		return r.items[canonicalName], true
	}

	var zero T
	return zero, false
}

// MustGet retrieves an item by name or alias, panics if not found
func (r *Registry[T]) MustGet(name string) T {
	item, exists := r.Get(name)
	if !exists {
		panic(fmt.Sprintf("item %s not found in registry", name))
	}
	return item
}

// Has checks if an item exists by name or alias
func (r *Registry[T]) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check direct name
	if _, exists := r.items[name]; exists {
		return true
	}

	// Check alias
	_, isAlias := r.aliases[name]
	return isAlias
}

// List returns all registered items
func (r *Registry[T]) List() []T {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]T, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, item)
	}
	return items
}

// ListNames returns all registered item names (not including aliases)
func (r *Registry[T]) ListNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.items))
	for name := range r.items {
		names = append(names, name)
	}

	// Sort for consistent ordering
	sort.Strings(names)
	return names
}

// Map returns a copy of the items map
func (r *Registry[T]) Map() map[string]T {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make(map[string]T, len(r.items))
	for name, item := range r.items {
		items[name] = item
	}
	return items
}

// Count returns the number of registered items (not including aliases)
func (r *Registry[T]) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.items)
}

// Clear removes all items and aliases from the registry
func (r *Registry[T]) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.items = make(map[string]T)
	r.aliases = make(map[string]string)
}

// Remove removes an item and its aliases from the registry
func (r *Registry[T]) Remove(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if it's a direct name
	if _, exists := r.items[name]; exists {
		delete(r.items, name)

		// Remove all aliases pointing to this item
		for alias, target := range r.aliases {
			if target == name {
				delete(r.aliases, alias)
			}
		}
		return true
	}

	// Check if it's an alias
	if canonicalName, isAlias := r.aliases[name]; isAlias {
		delete(r.items, canonicalName)

		// Remove all aliases pointing to this item
		for alias, target := range r.aliases {
			if target == canonicalName {
				delete(r.aliases, alias)
			}
		}
		return true
	}

	return false
}

// ResolveAlias resolves an alias to its canonical name
func (r *Registry[T]) ResolveAlias(alias string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	canonical, ok := r.aliases[alias]
	return canonical, ok
}

// GetAliases returns all aliases for a given item name
func (r *Registry[T]) GetAliases(name string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Make sure the item exists
	if _, exists := r.items[name]; !exists {
		return nil
	}

	var itemAliases []string
	for alias, target := range r.aliases {
		if target == name {
			itemAliases = append(itemAliases, alias)
		}
	}

	// Sort for consistent ordering
	sort.Strings(itemAliases)
	return itemAliases
}

// IsAlias checks if a given name is an alias
func (r *Registry[T]) IsAlias(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.aliases[name]
	return ok
}

// ForEach applies a function to each item in the registry
func (r *Registry[T]) ForEach(fn func(name string, item T)) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for name, item := range r.items {
		fn(name, item)
	}
}

// Filter returns items that match the given predicate
func (r *Registry[T]) Filter(predicate func(name string, item T) bool) []T {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []T
	for name, item := range r.items {
		if predicate(name, item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}
