// Package plugin provides the plugin system for extending Glide.
//
// This package contains the SDK for building plugins, the plugin manager
// for loading and running plugins, and utilities for plugin development.
// Plugins can add new commands, project detection, and custom functionality.
//
// # Plugin Architecture
//
// Glide uses HashiCorp's go-plugin for secure plugin isolation:
//   - Plugins run as separate processes
//   - Communication via gRPC
//   - Automatic plugin discovery from standard paths
//   - Security validation of plugin binaries
//
// # Package Structure
//
//	pkg/plugin/
//	├── sdk/           # Plugin SDK for developers
//	│   ├── v1/        # Legacy v1 protocol (stable)
//	│   └── v2/        # New v2 SDK with generics (recommended)
//	└── plugintest/    # Testing utilities
//
// # Plugin Discovery
//
// Plugins are discovered from these locations (in order):
//   - ~/.glide/plugins/     (global plugins)
//   - .glide/plugins/       (project plugins)
//   - Parent directory .glide/plugins/ (inherited)
//
// # Quick Start
//
// Build a plugin using SDK v2:
//
//	import "github.com/ivannovak/glide/v3/pkg/plugin/sdk/v2"
//
//	type MyPlugin struct {
//	    v2.BasePlugin[MyConfig]
//	}
//
//	func (p *MyPlugin) Metadata() v2.Metadata {
//	    return v2.Metadata{
//	        Name:        "my-plugin",
//	        Version:     "1.0.0",
//	        Description: "My awesome plugin",
//	    }
//	}
//
//	func (p *MyPlugin) Commands() []v2.Command {
//	    return []v2.Command{{
//	        Name: "greet",
//	        Run: func(ctx context.Context, args []string) error {
//	            fmt.Println("Hello from my plugin!")
//	            return nil
//	        },
//	    }}
//	}
//
//	func main() {
//	    v2.Serve(&MyPlugin{})
//	}
//
// See docs/guides/plugin-development.md for complete documentation.
package plugin
