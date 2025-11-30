// Package registry provides a thread-safe generic registry pattern.
//
// This package implements a reusable registry with support for named items,
// aliases, and thread-safe concurrent access. It's used throughout Glide
// for managing commands, plugins, formatters, and other extensible components.
//
// # Basic Usage
//
//	// Create a registry for string formatters
//	formatters := registry.New[Formatter]()
//
//	// Register items with optional aliases
//	formatters.Register("json", jsonFormatter, "j")
//	formatters.Register("yaml", yamlFormatter, "y", "yml")
//
//	// Retrieve by name or alias
//	f, err := formatters.Get("json")  // By name
//	f, err := formatters.Get("y")     // By alias
//
// # Generic Type Safety
//
// The registry uses Go generics for compile-time type safety:
//
//	type Command interface {
//	    Execute(args []string) error
//	}
//
//	commands := registry.New[Command]()
//	commands.Register("build", buildCmd)
//	commands.Register("test", testCmd)
//
//	cmd, err := commands.Get("build")
//	// cmd is typed as Command, no type assertion needed
//
// # Thread Safety
//
// All operations are safe for concurrent use:
//
//	var wg sync.WaitGroup
//	for i := 0; i < 10; i++ {
//	    wg.Add(1)
//	    go func(n int) {
//	        defer wg.Done()
//	        reg.Register(fmt.Sprintf("item-%d", n), item)
//	    }(i)
//	}
//	wg.Wait()
//
// # Listing and Iteration
//
//	// List all registered names (excluding aliases)
//	names := reg.List()
//
//	// Check if an item exists
//	if reg.Has("json") {
//	    // ...
//	}
//
// # Error Handling
//
// Registration fails for duplicate names or alias conflicts:
//
//	err := reg.Register("json", f1)       // OK
//	err = reg.Register("json", f2)        // Error: already registered
//	err = reg.Register("new", f3, "json") // Error: alias conflicts
package registry
