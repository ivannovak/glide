// Package v2 provides the next-generation Glide Plugin SDK.
//
// SDK v2 offers significant improvements over v1 including type-safe configuration
// using Go generics, simplified lifecycle management, and declarative command
// definitions. New plugins should use v2.
//
// # Key Features
//
//   - Type-safe configuration with Go generics
//   - Unified lifecycle (in-process and gRPC)
//   - Declarative command definition
//   - Simplified API with sensible defaults
//   - Full backward compatibility with v1
//
// # Creating a Plugin
//
// Define your configuration type:
//
//	type MyConfig struct {
//	    APIKey  string `json:"apiKey"`
//	    Timeout int    `json:"timeout"`
//	    Debug   bool   `json:"debug"`
//	}
//
// Implement the Plugin interface:
//
//	type MyPlugin struct {
//	    v2.BasePlugin[MyConfig]
//	}
//
//	func (p *MyPlugin) Metadata() v2.Metadata {
//	    return v2.Metadata{
//	        Name:        "my-plugin",
//	        Version:     "1.0.0",
//	        Description: "A sample plugin",
//	        Author:      "Your Name",
//	    }
//	}
//
//	func (p *MyPlugin) ConfigSchema() map[string]interface{} {
//	    return map[string]interface{}{
//	        "type": "object",
//	        "properties": map[string]interface{}{
//	            "apiKey":  map[string]interface{}{"type": "string"},
//	            "timeout": map[string]interface{}{"type": "integer"},
//	            "debug":   map[string]interface{}{"type": "boolean"},
//	        },
//	        "required": []string{"apiKey"},
//	    }
//	}
//
//	func (p *MyPlugin) Configure(ctx context.Context, cfg MyConfig) error {
//	    // Configuration is type-safe
//	    return p.Init(cfg)
//	}
//
//	func (p *MyPlugin) Commands() []v2.Command {
//	    return []v2.Command{{
//	        Name:        "hello",
//	        Description: "Say hello",
//	        Run: func(ctx context.Context, args []string) error {
//	            fmt.Println("Hello from my plugin!")
//	            return nil
//	        },
//	    }}
//	}
//
// # Plugin Entry Point
//
// Use Serve to run the plugin:
//
//	func main() {
//	    plugin := &MyPlugin{}
//	    if err := v2.Serve(plugin); err != nil {
//	        os.Exit(1)
//	    }
//	}
//
// # Lifecycle Hooks
//
// Override BasePlugin methods for lifecycle control:
//
//	func (p *MyPlugin) OnStart(ctx context.Context) error {
//	    // Initialize resources
//	    return nil
//	}
//
//	func (p *MyPlugin) OnStop(ctx context.Context) error {
//	    // Cleanup resources
//	    return nil
//	}
//
// # Testing Plugins
//
// Use the plugintest package for testing:
//
//	func TestPlugin(t *testing.T) {
//	    plugin := &MyPlugin{}
//	    ctx := context.Background()
//
//	    err := plugin.Configure(ctx, MyConfig{APIKey: "test"})
//	    require.NoError(t, err)
//
//	    err = plugin.OnStart(ctx)
//	    require.NoError(t, err)
//	}
//
// See docs/guides/plugin-development.md for complete documentation.
package v2
