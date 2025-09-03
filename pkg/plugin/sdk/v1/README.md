# Glide Plugin SDK v1

This is the SDK for creating runtime plugins for Glide using Hashicorp's go-plugin library.

## Quick Start

### Install Dependencies

```bash
go get github.com/hashicorp/go-plugin
go get google.golang.org/grpc
go get google.golang.org/protobuf
```

### Generate Protocol Buffers (if needed)

```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       plugin.proto
```

### Creating a Plugin

1. Implement the `GlidePlugin` interface:

```go
package main

import (
    "context"
    sdk "github.com/ivannovak/glide/pkg/plugin/sdk/v1"
)

type MyPlugin struct {
    config map[string]interface{}
}

func (p *MyPlugin) GetMetadata(ctx context.Context, _ *sdk.Empty) (*sdk.PluginMetadata, error) {
    return &sdk.PluginMetadata{
        Name:        "myplugin",
        Version:     "1.0.0",
        Description: "My custom plugin",
    }, nil
}

func (p *MyPlugin) ExecuteCommand(ctx context.Context, req *sdk.ExecuteRequest) (*sdk.ExecuteResponse, error) {
    // Implement command logic
    return &sdk.ExecuteResponse{
        Success: true,
        Stdout:  []byte("Command executed successfully"),
    }, nil
}
```

2. Create the plugin main:

```go
func main() {
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: sdk.HandshakeConfig,
        Plugins: map[string]plugin.Plugin{
            "glide": &sdk.GlidePluginGRPC{
                Impl: &MyPlugin{},
            },
        },
        GRPCServer: plugin.DefaultGRPCServer,
    })
}
```

3. Build your plugin:

```bash
go build -o glide-plugin-myplugin
```

4. Install the plugin:

```bash
mkdir -p ~/.glide/plugins
cp glide-plugin-myplugin ~/.glide/plugins/
```

## Interactive Commands

For interactive commands like MySQL CLI, use the bidirectional streaming:

```go
func (p *MyPlugin) StartInteractive(stream sdk.GlidePlugin_StartInteractiveServer) error {
    // Read initial request
    msg, err := stream.Recv()
    if err != nil {
        return err
    }
    
    // Handle stdin/stdout/stderr streaming
    for {
        msg, err := stream.Recv()
        if err != nil {
            return err
        }
        
        switch msg.Type {
        case sdk.StreamType_STDIN:
            // Process input
        case sdk.StreamType_SIGNAL:
            // Handle signals
        case sdk.StreamType_RESIZE:
            // Handle terminal resize
        }
        
        // Send output
        if output := getOutput(); output != nil {
            stream.Send(&sdk.StreamMessage{
                Type: sdk.StreamType_STDOUT,
                Data: output,
            })
        }
    }
}
```

## Testing

Use the provided test helpers to test your plugin:

```go
func TestPlugin(t *testing.T) {
    plugin := &MyPlugin{}
    
    // Test metadata
    metadata, err := plugin.GetMetadata(context.Background(), &sdk.Empty{})
    assert.NoError(t, err)
    assert.Equal(t, "myplugin", metadata.Name)
    
    // Test command execution
    resp, err := plugin.ExecuteCommand(context.Background(), &sdk.ExecuteRequest{
        Command: "mycommand",
        Args:    []string{"arg1"},
    })
    assert.NoError(t, err)
    assert.True(t, resp.Success)
}
```

## Best Practices

1. **Error Handling**: Always return meaningful errors
2. **Resource Cleanup**: Clean up resources in defer statements
3. **Logging**: Use structured logging through the host interface
4. **Configuration**: Support environment variable overrides
5. **Security**: Validate all inputs and paths

## API Reference

See the [types.go](types.go) file for complete API documentation.