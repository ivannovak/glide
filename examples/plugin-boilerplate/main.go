//go:build ignore
// +build ignore

// Glide Plugin Boilerplate (SDK v2)
//
// This is a template for creating your own Glide runtime plugin using SDK v2.
//
// To use this boilerplate:
// 1. Copy this directory to a new location
// 2. Rename the plugin and update metadata
// 3. Implement your custom commands
// 4. Build with: go build -o glide-plugin-yourname
// 5. Install to: ~/.glide/plugins/
//
// SDK v2 Features:
// - Type-safe configuration with Go generics
// - Unified lifecycle management (Init/Start/Stop/HealthCheck)
// - Declarative command definitions
// - Simplified API with BasePlugin[C]

package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ivannovak/glide/v2/pkg/plugin/sdk/v2"
)

// Config defines the plugin's type-safe configuration.
// Users configure this in .glide.yml under plugins.myplugin
type Config struct {
	// Greeting prefix for the hello command
	Greeting string `json:"greeting" yaml:"greeting"`

	// Enable debug output
	Debug bool `json:"debug" yaml:"debug"`

	// Maximum retries for operations
	MaxRetries int `json:"maxRetries" yaml:"maxRetries" validate:"min=0,max=10"`
}

// DefaultConfig returns sensible defaults
func DefaultConfig() Config {
	return Config{
		Greeting:   "Hello",
		Debug:      false,
		MaxRetries: 3,
	}
}

// MyPlugin is the main plugin implementation.
// Embed v2.BasePlugin[Config] for automatic handling of common functionality.
type MyPlugin struct {
	v2.BasePlugin[Config]
}

// Metadata returns plugin information.
// Update these fields with your plugin's details.
func (p *MyPlugin) Metadata() v2.Metadata {
	return v2.Metadata{
		Name:        "myplugin",
		Version:     "1.0.0",
		Author:      "Your Name",
		Description: "Brief description of your plugin",
		License:     "MIT",
		Homepage:    "https://github.com/user/glide-plugin-myplugin",
		Aliases:     []string{"mp", "myp"}, // Optional shortcuts
	}
}

// Configure is called with the type-safe configuration.
func (p *MyPlugin) Configure(ctx context.Context, config Config) error {
	// Store configuration in BasePlugin
	if err := p.BasePlugin.Configure(ctx, config); err != nil {
		return err
	}

	// Additional validation or setup can go here
	if p.GetConfig().Debug {
		fmt.Fprintln(os.Stderr, "[DEBUG] Plugin configured with debug mode enabled")
	}

	return nil
}

// Init is called once after plugin load (optional).
func (p *MyPlugin) Init(ctx context.Context) error {
	// One-time initialization
	return nil
}

// Start is called when the plugin should begin operation (optional).
func (p *MyPlugin) Start(ctx context.Context) error {
	// Connect to services, start background workers, etc.
	return nil
}

// Stop is called for graceful shutdown (optional).
func (p *MyPlugin) Stop(ctx context.Context) error {
	// Cleanup resources
	return nil
}

// HealthCheck returns nil if the plugin is healthy (optional).
func (p *MyPlugin) HealthCheck(ctx context.Context) error {
	// Verify connectivity, resources, etc.
	return nil
}

// Commands returns the list of commands this plugin provides.
func (p *MyPlugin) Commands() []v2.Command {
	return []v2.Command{
		// Simple greeting command
		{
			Name:        "hello",
			Description: "Say hello to someone",
			Category:    "developer",
			Aliases:     []string{"h"},
			Handler:     v2.SimpleCommandHandler(p.helloCommand),
		},

		// Command that shows configuration
		{
			Name:        "config",
			Description: "Show plugin configuration",
			Category:    "developer",
			Aliases:     []string{"c"},
			Handler:     v2.SimpleCommandHandler(p.configCommand),
		},

		// Command with flags
		{
			Name:        "greet",
			Description: "Greet with options",
			Category:    "developer",
			Flags: []v2.Flag{
				{Name: "name", Short: "n", Description: "Name to greet", Default: "World"},
				{Name: "loud", Short: "l", Description: "Use uppercase", Type: v2.FlagBool},
			},
			Handler: v2.SimpleCommandHandler(p.greetCommand),
		},

		// Example showing error handling
		{
			Name:        "validate",
			Description: "Validate something (demonstrates error handling)",
			Category:    "developer",
			Handler:     v2.SimpleCommandHandler(p.validateCommand),
		},
	}
}

// helloCommand implements the hello command.
func (p *MyPlugin) helloCommand(ctx context.Context, req *v2.ExecuteRequest) (*v2.ExecuteResponse, error) {
	name := "World"
	if len(req.Args) > 0 {
		name = strings.Join(req.Args, " ")
	}

	// Use configured greeting prefix
	greeting := p.GetConfig().Greeting
	if greeting == "" {
		greeting = "Hello"
	}

	output := fmt.Sprintf("%s, %s!\n", greeting, name)
	return &v2.ExecuteResponse{
		ExitCode: 0,
		Output:   output,
	}, nil
}

// configCommand shows the current configuration.
func (p *MyPlugin) configCommand(ctx context.Context, req *v2.ExecuteRequest) (*v2.ExecuteResponse, error) {
	config := p.GetConfig()

	output := fmt.Sprintf(`Plugin Configuration:
  Greeting:   %s
  Debug:      %v
  MaxRetries: %d
`, config.Greeting, config.Debug, config.MaxRetries)

	return &v2.ExecuteResponse{
		ExitCode: 0,
		Output:   output,
	}, nil
}

// greetCommand demonstrates flag handling.
func (p *MyPlugin) greetCommand(ctx context.Context, req *v2.ExecuteRequest) (*v2.ExecuteResponse, error) {
	name := req.Flags["name"]
	loud := req.Flags["loud"] == "true"

	greeting := fmt.Sprintf("%s, %s!", p.GetConfig().Greeting, name)
	if loud {
		greeting = strings.ToUpper(greeting)
	}

	return &v2.ExecuteResponse{
		ExitCode: 0,
		Output:   greeting + "\n",
	}, nil
}

// validateCommand demonstrates error handling.
func (p *MyPlugin) validateCommand(ctx context.Context, req *v2.ExecuteRequest) (*v2.ExecuteResponse, error) {
	if len(req.Args) == 0 {
		return &v2.ExecuteResponse{
			ExitCode: 1,
			Error:    "validation requires at least one argument",
		}, nil
	}

	// Successful validation
	return &v2.ExecuteResponse{
		ExitCode: 0,
		Output:   fmt.Sprintf("Validated: %s\n", strings.Join(req.Args, ", ")),
	}, nil
}

func main() {
	plugin := &MyPlugin{}

	// Serve the plugin
	if err := v2.Serve(plugin); err != nil {
		fmt.Fprintf(os.Stderr, "Plugin error: %v\n", err)
		os.Exit(1)
	}
}

// User Configuration Example (.glide.yml):
//
// plugins:
//   myplugin:
//     greeting: "Hi there"
//     debug: true
//     maxRetries: 5
//
// Usage Examples:
//   glide hello              # Uses configured greeting
//   glide hello John         # Greet John
//   glide mp hello           # Using plugin alias
//   glide config             # Show configuration
//   glide greet -n Alice -l  # Greet Alice loudly
