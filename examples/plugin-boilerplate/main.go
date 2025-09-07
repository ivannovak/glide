// Glide Plugin Boilerplate
// This is a template for creating your own Glide runtime plugin.
//
// To use this boilerplate:
// 1. Copy this directory to a new location
// 2. Rename the plugin and update metadata
// 3. Implement your custom commands
// 4. Build with: go build -o glide-plugin-yourname
// 5. Install to: ~/.glide/plugins/
//
// Features demonstrated:
// - Plugin metadata with version, author, and description
// - Plugin-level aliases (e.g., 'mp' for 'myplugin')
// - Command-level aliases (e.g., 'h' for 'hello')
// - Non-interactive and interactive command examples
// - Configuration handling from .glide.yml
//
// With aliases, users can use shortcuts like:
// - 'glid mp h' instead of 'glid myplugin hello'
// - 'glid myp c' instead of 'glid myplugin config'

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/go-plugin"
	sdk "github.com/ivannovak/glide/pkg/plugin/sdk/v1"
)

// MyPlugin implements the Glide runtime plugin interface
// Rename this struct to match your plugin's purpose
type MyPlugin struct {
	sdk.UnimplementedGlidePluginServer
	config map[string]interface{}
}

// GetMetadata returns plugin information
// Update these fields with your plugin's information
func (p *MyPlugin) GetMetadata(ctx context.Context, _ *sdk.Empty) (*sdk.PluginMetadata, error) {
	return &sdk.PluginMetadata{
		Name:        "myplugin",                     // Change to your plugin name
		Version:     "1.0.0",                        // Your plugin version
		Author:      "Your Name",                    // Your name or organization
		Description: "Brief description of plugin",  // What your plugin does
		Homepage:    "https://github.com/user/repo", // Optional: plugin homepage
		License:     "MIT",                          // Optional: plugin license
		MinSdk:      "v1.0.0",                       // Minimum SDK version required
		// Plugin-level aliases allow users to access your plugin with shorter names
		// For example: 'glid mp' instead of 'glid myplugin'
		Aliases:     []string{"mp", "myp"},          // Plugin-level aliases
		// Tags can be used for categorization (e.g., "database", "testing", "deployment")
		Tags:        []string{"example", "boilerplate"}, // Categorization tags
	}, nil
}

// Configure sets up the plugin with configuration from Glide
func (p *MyPlugin) Configure(ctx context.Context, req *sdk.ConfigureRequest) (*sdk.ConfigureResponse, error) {
	// Store configuration for later use
	p.config = make(map[string]interface{})
	for k, v := range req.Config {
		p.config[k] = v
	}

	return &sdk.ConfigureResponse{
		Success: true,
		Message: "Plugin configured successfully",
	}, nil
}

// ListCommands returns all available commands
// Add your custom commands here
func (p *MyPlugin) ListCommands(ctx context.Context, _ *sdk.Empty) (*sdk.CommandList, error) {
	return &sdk.CommandList{
		Commands: []*sdk.CommandInfo{
			{
				Name:        "hello",
				Description: "Print a greeting message",
				Category:    "example",
				// Command aliases allow shorter invocations
				// For example: 'glid myplugin h' instead of 'glid myplugin hello'
				Aliases:     []string{"h", "hi"},
				Interactive: false,
				Hidden:      false,
			},
			{
				Name:        "config",
				Description: "Show plugin configuration",
				Category:    "debug",
				// Single character alias for quick access
				Aliases:     []string{"c", "cfg"},
				Interactive: false,
				Hidden:      false,
			},
			{
				Name:        "interactive",
				Description: "Example interactive command",
				Category:    "example",
				// Aliases for interactive mode
				Aliases:     []string{"i", "int"},
				Interactive: true,
				Hidden:      false,
			},
		},
	}, nil
}

// ExecuteCommand handles non-interactive command execution
// The command in req.Command may be either the full name or an alias
func (p *MyPlugin) ExecuteCommand(ctx context.Context, req *sdk.ExecuteRequest) (*sdk.ExecuteResponse, error) {
	// Map aliases to their full command names
	commandMap := map[string]string{
		"hello":       "hello",
		"h":           "hello",
		"hi":          "hello",
		"config":      "config",
		"c":           "config",
		"cfg":         "config",
		"interactive": "interactive",
		"i":           "interactive",
		"int":         "interactive",
	}

	// Resolve the command name (handle both full names and aliases)
	actualCommand, exists := commandMap[req.Command]
	if !exists {
		return &sdk.ExecuteResponse{
			Success: false,
			Error:   fmt.Sprintf("unknown command: %s", req.Command),
		}, nil
	}

	switch actualCommand {
	case "hello":
		return p.executeHello(ctx, req)
	case "config":
		return p.executeConfig(ctx, req)
	case "interactive":
		// Interactive commands should return RequiresInteractive
		return &sdk.ExecuteResponse{
			RequiresInteractive: true,
			Success:             false,
			Error:               "This command requires interactive mode",
		}, nil
	default:
		return &sdk.ExecuteResponse{
			Success: false,
			Error:   fmt.Sprintf("unknown command: %s", req.Command),
		}, nil
	}
}

// executeHello is an example simple command
func (p *MyPlugin) executeHello(ctx context.Context, req *sdk.ExecuteRequest) (*sdk.ExecuteResponse, error) {
	name := "World"
	if len(req.Args) > 0 {
		name = req.Args[0]
	}

	message := fmt.Sprintf("Hello, %s! 👋\n", name)
	message += "This is an example command from your Glide plugin.\n"

	return &sdk.ExecuteResponse{
		Success:  true,
		Stdout:   []byte(message),
		ExitCode: 0,
	}, nil
}

// executeConfig shows the plugin's configuration
func (p *MyPlugin) executeConfig(ctx context.Context, req *sdk.ExecuteRequest) (*sdk.ExecuteResponse, error) {
	output := "Plugin Configuration:\n"
	output += "====================\n"

	if len(p.config) == 0 {
		output += "No configuration provided.\n"
		output += "\nTo configure this plugin, add to your .glide.yml:\n"
		output += "```yaml\n"
		output += "plugins:\n"
		output += "  myplugin:\n"
		output += "    key: value\n"
		output += "```\n"
	} else {
		for key, value := range p.config {
			output += fmt.Sprintf("%s: %v\n", key, value)
		}
	}

	output += "\nAvailable Commands and Aliases:\n"
	output += "-------------------------------\n"
	output += "Plugin aliases: mp, myp\n"
	output += "\nCommands:\n"
	output += "  hello (aliases: h, hi)       - Print a greeting message\n"
	output += "  config (aliases: c, cfg)     - Show plugin configuration\n"
	output += "  interactive (aliases: i, int) - Example interactive command\n"
	output += "\nUsage examples:\n"
	output += "  glid myplugin hello World\n"
	output += "  glid mp h World              # Using plugin and command aliases\n"
	output += "  glid myp cfg                 # Show config using aliases\n"

	return &sdk.ExecuteResponse{
		Success:  true,
		Stdout:   []byte(output),
		ExitCode: 0,
	}, nil
}

// StartInteractive handles interactive command execution
// This is called for commands marked as Interactive: true
func (p *MyPlugin) StartInteractive(stream sdk.GlidePlugin_StartInteractiveServer) error {
	// For now, we just handle a simple interactive example
	// In a real implementation, you would check the command from the stream

	// Send initial output
	err := stream.Send(&sdk.StreamMessage{
		Type: sdk.StreamMessage_STDOUT,
		Data: []byte("Starting interactive session...\n"),
	})
	if err != nil {
		return err
	}

	// Example: Echo user input back
	// In a real plugin, you might:
	// - Start a shell or REPL
	// - Connect to a database CLI
	// - Launch an interactive debugger

	err = stream.Send(&sdk.StreamMessage{
		Type: sdk.StreamMessage_STDOUT,
		Data: []byte("Type 'exit' to quit.\n> "),
	})
	if err != nil {
		return err
	}

	// Read input from user
	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}

		if msg.Type == sdk.StreamMessage_STDIN {
			input := string(msg.Data)

			// Check for exit command
			if input == "exit\n" || input == "quit\n" {
				stream.Send(&sdk.StreamMessage{
					Type: sdk.StreamMessage_STDOUT,
					Data: []byte("Goodbye!\n"),
				})
				break
			}

			// Echo back the input
			response := fmt.Sprintf("You typed: %s> ", input)
			stream.Send(&sdk.StreamMessage{
				Type: sdk.StreamMessage_STDOUT,
				Data: []byte(response),
			})
		}
	}

	// Send exit message
	return stream.Send(&sdk.StreamMessage{
		Type:     sdk.StreamMessage_EXIT,
		ExitCode: 0,
	})
}

// GetCapabilities returns the plugin's required capabilities
// This helps the host understand what permissions your plugin needs
func (p *MyPlugin) GetCapabilities(ctx context.Context, _ *sdk.Empty) (*sdk.Capabilities, error) {
	return &sdk.Capabilities{
		RequiresDocker:  false, // Set to true if you need Docker access
		RequiresNetwork: false, // Set to true if you need network access
	}, nil
}

func main() {
	// Check if running in debug mode
	if os.Getenv("GLIDE_PLUGIN_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[DEBUG] Starting plugin...\n")
	}

	// Create plugin instance
	myPlugin := &MyPlugin{
		config: make(map[string]interface{}),
	}

	// Configure plugin map for Hashicorp go-plugin
	pluginMap := map[string]plugin.Plugin{
		"glide": &sdk.GlidePluginImpl{
			Impl: myPlugin,
		},
	}

	// Start the plugin server
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: sdk.HandshakeConfig,
		Plugins:         pluginMap,
		GRPCServer:      plugin.DefaultGRPCServer,
	})
}
