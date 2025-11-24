//go:build ignore
// +build ignore

// Glide Plugin Boilerplate
// This is a template for creating your own Glide runtime plugin.
//
// To use this boilerplate:
// 1. Copy this directory to a new location
// 2. Rename the plugin and update metadata
// 3. Implement your custom commands
// 4. Build with: go build -o yourplugin
// 5. Install to: ~/.glide/plugins/
//
// This boilerplate uses the BasePlugin helper which automatically handles:
// - Command registration and routing
// - Interactive command support (no manual StartInteractive routing needed!)
// - Configuration management
// - Command listing
//
// Features demonstrated:
// - Plugin metadata with version, author, and description
// - Simple non-interactive commands
// - Interactive commands with automatic routing
// - Configuration handling from .glide.yml
// - Plugin and command aliases

package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-plugin"
	sdk "github.com/ivannovak/glide/v2/pkg/plugin/sdk/v1"
)

func main() {
	// Create plugin with metadata
	// Update these fields with your plugin's information
	basePlugin := sdk.NewBasePlugin(&sdk.PluginMetadata{
		Name:        "myplugin",                     // Change to your plugin name
		Version:     "1.0.0",                        // Your plugin version
		Author:      "Your Name",                    // Your name or organization
		Description: "Brief description of plugin",  // What your plugin does
		Homepage:    "https://github.com/user/repo", // Optional: plugin homepage
		License:     "MIT",                          // Optional: plugin license
		MinSdk:      "v1.0.0",                       // Minimum SDK version required
		Aliases:     []string{"mp", "myp"},          // Optional: shortcuts for plugin name
		Namespaced:  false,                          // false = commands at root level
	})

	// Register a simple non-interactive command
	basePlugin.RegisterCommand("hello", sdk.NewSimpleCommand(
		&sdk.CommandInfo{
			Name:        "hello",
			Description: "Say hello to someone",
			Category:    sdk.CategoryDeveloper,
			Aliases:     []string{"h"}, // Users can type 'glideh' or 'glidehello'
		},
		func(ctx context.Context, req *sdk.ExecuteRequest) (*sdk.ExecuteResponse, error) {
			name := "World"
			if len(req.Args) > 0 {
				name = strings.Join(req.Args, " ")
			}

			greeting := fmt.Sprintf("Hello, %s! 👋\n", name)
			return &sdk.ExecuteResponse{
				Success: true,
				Stdout:  []byte(greeting),
			}, nil
		},
	))

	// Register a command that shows configuration
	basePlugin.RegisterCommand("config", sdk.NewSimpleCommand(
		&sdk.CommandInfo{
			Name:        "config",
			Description: "Show plugin configuration",
			Category:    sdk.CategoryDeveloper,
			Aliases:     []string{"c"},
		},
		func(ctx context.Context, req *sdk.ExecuteRequest) (*sdk.ExecuteResponse, error) {
			config := basePlugin.GetConfig()

			if len(config) == 0 {
				return &sdk.ExecuteResponse{
					Success: true,
					Stdout:  []byte("No configuration found\n"),
				}, nil
			}

			var output strings.Builder
			output.WriteString("Plugin Configuration:\n")
			for key, value := range config {
				output.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
			}

			return &sdk.ExecuteResponse{
				Success: true,
				Stdout:  []byte(output.String()),
			}, nil
		},
	))

	// Register an interactive command - NO BOILERPLATE NEEDED!
	// The BasePlugin automatically handles command routing
	basePlugin.RegisterCommand("shell", sdk.NewBaseInteractiveCommand(
		&sdk.CommandInfo{
			Name:        "shell",
			Description: "Interactive shell example",
			Category:    sdk.CategoryDeveloper,
			Interactive: true, // Automatically set by NewBaseInteractiveCommand
			Aliases:     []string{"sh"},
		},
		// Execute handler (optional - returns RequiresInteractive by default)
		nil,
		// Interactive handler - just implement your logic!
		func(stream sdk.GlidePlugin_StartInteractiveServer) error {
			// Send welcome message
			if err := stream.Send(&sdk.StreamMessage{
				Type: sdk.StreamMessage_STDOUT,
				Data: []byte("Welcome to the example interactive shell!\n"),
			}); err != nil {
				return err
			}

			if err := stream.Send(&sdk.StreamMessage{
				Type: sdk.StreamMessage_STDOUT,
				Data: []byte("Type 'help' for commands, 'exit' to quit.\n> "),
			}); err != nil {
				return err
			}

			// Simple command loop
			for {
				msg, err := stream.Recv()
				if err != nil {
					break
				}

				if msg.Type == sdk.StreamMessage_STDIN {
					input := strings.TrimSpace(string(msg.Data))

					switch input {
					case "exit", "quit":
						stream.Send(&sdk.StreamMessage{
							Type: sdk.StreamMessage_STDOUT,
							Data: []byte("Goodbye!\n"),
						})
						stream.Send(&sdk.StreamMessage{
							Type:     sdk.StreamMessage_EXIT,
							ExitCode: 0,
						})
						return nil

					case "help":
						help := `Available commands:
  help  - Show this help message
  echo  - Echo your input
  exit  - Exit the shell
> `
						stream.Send(&sdk.StreamMessage{
							Type: sdk.StreamMessage_STDOUT,
							Data: []byte(help),
						})

					default:
						if strings.HasPrefix(input, "echo ") {
							echoText := strings.TrimPrefix(input, "echo ")
							response := fmt.Sprintf("Echo: %s\n> ", echoText)
							stream.Send(&sdk.StreamMessage{
								Type: sdk.StreamMessage_STDOUT,
								Data: []byte(response),
							})
						} else if input != "" {
							response := fmt.Sprintf("Unknown command: %s\nType 'help' for available commands.\n> ", input)
							stream.Send(&sdk.StreamMessage{
								Type: sdk.StreamMessage_STDOUT,
								Data: []byte(response),
							})
						} else {
							stream.Send(&sdk.StreamMessage{
								Type: sdk.StreamMessage_STDOUT,
								Data: []byte("> "),
							})
						}
					}
				} else if msg.Type == sdk.StreamMessage_SIGNAL {
					// Handle Ctrl+C
					if msg.Signal == "SIGINT" {
						stream.Send(&sdk.StreamMessage{
							Type:     sdk.StreamMessage_EXIT,
							ExitCode: 130, // Standard exit code for SIGINT
						})
						return nil
					}
				}
			}

			return nil
		},
	))

	// Example of adding a command that can work both ways
	basePlugin.RegisterCommand("process", sdk.NewBaseInteractiveCommand(
		&sdk.CommandInfo{
			Name:        "process",
			Description: "Process data (can be interactive or non-interactive)",
			Category:    sdk.CategoryDeveloper,
			Interactive: true,
		},
		// Execute handler - for non-interactive mode
		func(ctx context.Context, req *sdk.ExecuteRequest) (*sdk.ExecuteResponse, error) {
			if len(req.Args) > 0 {
				// Non-interactive mode: process arguments
				result := fmt.Sprintf("Processed: %s\n", strings.Join(req.Args, ", "))
				return &sdk.ExecuteResponse{
					Success: true,
					Stdout:  []byte(result),
				}, nil
			}

			// No arguments: switch to interactive mode
			return &sdk.ExecuteResponse{
				RequiresInteractive: true,
			}, nil
		},
		// Interactive handler - for interactive mode
		func(stream sdk.GlidePlugin_StartInteractiveServer) error {
			stream.Send(&sdk.StreamMessage{
				Type: sdk.StreamMessage_STDOUT,
				Data: []byte("Interactive processing mode. Enter data to process, 'done' when finished:\n"),
			})

			var items []string
			for {
				msg, err := stream.Recv()
				if err != nil {
					break
				}

				if msg.Type == sdk.StreamMessage_STDIN {
					input := strings.TrimSpace(string(msg.Data))
					if input == "done" {
						result := fmt.Sprintf("\nProcessed %d items:\n- %s\n",
							len(items), strings.Join(items, "\n- "))
						stream.Send(&sdk.StreamMessage{
							Type: sdk.StreamMessage_STDOUT,
							Data: []byte(result),
						})
						stream.Send(&sdk.StreamMessage{
							Type:     sdk.StreamMessage_EXIT,
							ExitCode: 0,
						})
						return nil
					}

					items = append(items, input)
					stream.Send(&sdk.StreamMessage{
						Type: sdk.StreamMessage_STDOUT,
						Data: []byte(fmt.Sprintf("Added: %s\n", input)),
					})
				}
			}
			return nil
		},
	))

	// Start the plugin server
	// That's it! No need to implement GetMetadata, Configure, ListCommands,
	// ExecuteCommand, or StartInteractive - BasePlugin handles it all!
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: sdk.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"glide": &sdk.GlidePlugin{Impl: basePlugin},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}

// Benefits of using BasePlugin:
// 1. No manual StartInteractive routing needed
// 2. Automatic command registration and discovery
// 3. Built-in configuration management
// 4. Clean separation of command logic
// 5. Less boilerplate, more focus on functionality
//
// Compare this to the old way where you had to:
// - Implement all gRPC methods manually
// - Handle StartInteractive routing yourself
// - Extract command names from stream messages
// - Maintain command maps and routing logic
// - Handle configuration storage
//
// Now you just register commands and implement their logic!
