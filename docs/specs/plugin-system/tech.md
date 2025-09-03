# Plugin System - Technical Specification

## Architecture Overview

The plugin system uses Hashicorp's go-plugin library with gRPC for communication between Glide (host) and plugins (separate processes). This provides language independence, process isolation, and version compatibility.

```
┌─────────────┐  gRPC    ┌─────────────┐
│    Glide    │◄────────►│   Plugin    │
│    (Host)   │          │  (Process)  │
└─────────────┘          └─────────────┘
```

## Technical Components

### 1. Plugin SDK (`pkg/plugin/sdk/v1/`)

**Protocol Definition** (`plugin.proto`):
```protobuf
service GlidePlugin {
    rpc GetMetadata(Empty) returns (PluginMetadata);
    rpc Configure(ConfigureRequest) returns (ConfigureResponse);
    rpc ListCommands(Empty) returns (CommandList);
    rpc ExecuteCommand(ExecuteRequest) returns (ExecuteResponse);
    rpc StartInteractive(stream StreamMessage) returns (stream StreamMessage);
    rpc GetCapabilities(Empty) returns (Capabilities);
}
```

**Core Types**:
```go
type PluginMetadata struct {
    Name        string
    Version     string
    Author      string
    Description string
    MinSdk      string
}

type CommandInfo struct {
    Name         string
    Description  string
    Category     string
    Interactive  bool
    RequiresTty  bool
}

type Capabilities struct {
    RequiresDocker   bool
    RequiresNetwork  bool
}
```

### 2. Plugin Manager (`pkg/plugin/sdk/manager.go`)

**Lifecycle Management**:
- Discovery: Scan plugin directory
- Loading: Start plugin process
- Communication: Establish gRPC connection
- Configuration: Pass config to plugin
- Cleanup: Graceful shutdown

**Discovery Process**:
```go
func (m *Manager) Discover(dir string) ([]*Plugin, error) {
    // Scan for files matching "glide-plugin-*"
    // Verify executable permissions
    // Return discovered plugins
}
```

### 3. Plugin Registry (`pkg/plugin/registry.go`)

**Command Registration**:
```go
func (r *Registry) Register(plugin Plugin) error {
    // Get plugin metadata
    // List available commands
    // Register with Cobra command tree
    // Set up command routing
}
```

### 4. Plugin Integration (`pkg/plugin/runtime_integration.go`)

**Command Routing**:
```go
func createPluginCommand(plugin *Plugin, cmd CommandInfo) *cobra.Command {
    return &cobra.Command{
        Use:   cmd.Name,
        Short: cmd.Description,
        RunE: func(cobraCmd *cobra.Command, args []string) error {
            if cmd.Interactive {
                return executeInteractive(plugin, cmd, args)
            }
            return executeCommand(plugin, cmd, args)
        },
    }
}
```

## Plugin Development

### 1. Plugin Interface Implementation

```go
type MyPlugin struct {
    sdk.UnimplementedGlidePluginServer
    config map[string]interface{}
}

func (p *MyPlugin) GetMetadata(ctx context.Context, _ *sdk.Empty) (*sdk.PluginMetadata, error) {
    return &sdk.PluginMetadata{
        Name:        "myplugin",
        Version:     "1.0.0",
        Author:      "Author Name",
        Description: "My custom plugin",
        MinSdk:      "v1.0.0",
    }, nil
}

func (p *MyPlugin) ListCommands(ctx context.Context, _ *sdk.Empty) (*sdk.CommandList, error) {
    return &sdk.CommandList{
        Commands: []*sdk.CommandInfo{
            {
                Name:        "hello",
                Description: "Say hello",
                Category:    "example",
            },
        },
    }, nil
}

func (p *MyPlugin) ExecuteCommand(ctx context.Context, req *sdk.ExecuteRequest) (*sdk.ExecuteResponse, error) {
    // Command implementation
}
```

### 2. Plugin Main Function

```go
func main() {
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: sdk.HandshakeConfig,
        Plugins: map[string]plugin.Plugin{
            "glide": &sdk.GlidePluginImpl{
                Impl: &MyPlugin{},
            },
        },
        GRPCServer: plugin.DefaultGRPCServer,
    })
}
```

### 3. Interactive Commands

```go
func (p *MyPlugin) StartInteractive(stream sdk.GlidePlugin_StartInteractiveServer) error {
    // Bidirectional streaming for TTY support
    for {
        msg, err := stream.Recv()
        if err != nil {
            return err
        }
        
        switch msg.Type {
        case sdk.StreamMessage_STDIN:
            // Handle input
        case sdk.StreamMessage_SIGNAL:
            // Handle signals
        }
        
        // Send output
        stream.Send(&sdk.StreamMessage{
            Type: sdk.StreamMessage_STDOUT,
            Data: output,
        })
    }
}
```

## Communication Protocol

### 1. Handshake

Initial handshake ensures version compatibility:
```go
var HandshakeConfig = plugin.HandshakeConfig{
    ProtocolVersion:  1,
    MagicCookieKey:   "GLIDE_PLUGIN",
    MagicCookieValue: "v1",
}
```

### 2. Message Types

**Request/Response**:
- Metadata queries
- Command listings
- Command execution
- Configuration updates

**Streaming**:
- Interactive I/O
- Progress updates
- Log streaming
- Signal forwarding

### 3. Error Handling

```go
// Structured errors with context
type PluginError struct {
    Code    ErrorCode
    Message string
    Details map[string]string
}
```

## Security Implementation

### 1. Process Isolation

```go
// Plugin runs in separate process
cmd := exec.Command(pluginPath)
cmd.Env = sanitizedEnv()  // Limited environment
cmd.Dir = sandboxDir()     // Restricted directory
```

### 2. Capability Enforcement

```go
func (m *Manager) verifyCapabilities(plugin *Plugin, required Capabilities) error {
    capabilities := plugin.GetCapabilities()
    
    if required.RequiresDocker && !capabilities.RequiresDocker {
        return ErrDockerNotAllowed
    }
    
    if required.RequiresNetwork && !capabilities.RequiresNetwork {
        return ErrNetworkNotAllowed
    }
    
    return nil
}
```

### 3. Input Sanitization

```go
func sanitizeInput(input string) string {
    // Remove control characters
    // Validate UTF-8
    // Limit length
    // Escape special characters
}
```

## Performance Optimization

### 1. Lazy Loading

```go
// Plugins loaded only when needed
func (m *Manager) LoadPlugin(name string) (*Plugin, error) {
    if cached := m.cache.Get(name); cached != nil {
        return cached, nil
    }
    
    plugin := m.startPlugin(name)
    m.cache.Set(name, plugin)
    return plugin, nil
}
```

### 2. Connection Pooling

```go
// Reuse gRPC connections
type ConnectionPool struct {
    connections map[string]*grpc.ClientConn
    mu          sync.RWMutex
}
```

### 3. Caching

```go
// Cache plugin metadata and commands
type PluginCache struct {
    metadata map[string]*PluginMetadata
    commands map[string]*CommandList
    ttl      time.Duration
}
```

## Testing Strategy

### 1. Plugin Test Harness

```go
// Test harness for plugin development
harness := plugintest.NewTestHarness(t)
harness.RegisterPlugin(myPlugin)
harness.ExecuteCommand("hello", "--flag", "value")
harness.AssertOutput("Hello, World!")
```

### 2. Integration Tests

```go
func TestPluginLifecycle(t *testing.T) {
    // Start plugin
    // Execute commands
    // Verify output
    // Clean shutdown
}
```

### 3. Mock Plugin

```go
// Mock plugin for testing
type MockPlugin struct {
    commands []CommandInfo
    responses map[string]ExecuteResponse
}
```

## Distribution

### 1. Plugin Packaging

```bash
# Single binary distribution
go build -o glide-plugin-name

# With assets
tar -czf plugin.tar.gz glide-plugin-name assets/
```

### 2. Installation Methods

```bash
# Direct copy
cp glide-plugin-name ~/.glide/plugins/

# From URL
glide plugins install https://example.com/plugin

# From GitHub release
glide plugins install github:user/repo
```

### 3. Versioning

```go
// Semantic versioning with compatibility checks
type Version struct {
    Major int
    Minor int
    Patch int
}

func CheckCompatibility(plugin, sdk Version) bool {
    return plugin.Major == sdk.Major && 
           plugin.Minor >= sdk.Minor
}
```

## Performance Metrics

### Startup Performance
- Plugin discovery: < 5ms
- Plugin loading: < 50ms per plugin
- Command registration: < 10ms
- Total overhead: < 100ms

### Runtime Performance
- Command routing: < 1ms
- gRPC call: < 10ms
- Process spawn: < 50ms
- Memory per plugin: < 10MB

### Benchmarks

```go
func BenchmarkPluginExecution(b *testing.B) {
    plugin := loadPlugin("example")
    for i := 0; i < b.N; i++ {
        plugin.Execute("command", []string{"arg"})
    }
}
```

## Future Technical Enhancements

### 1. Plugin Dependencies
```yaml
dependencies:
  - plugin: database
    version: ">=1.0.0"
  - plugin: cloud
    version: "^2.0.0"
```

### 2. Event System
```go
// Plugins can subscribe to events
type EventSubscriber interface {
    OnEvent(event Event) error
}
```

### 3. Shared Libraries
```go
// Common functionality for plugins
type SharedLibrary interface {
    HTTPClient() *http.Client
    Logger() Logger
    Cache() Cache
}
```

### 4. Hot Reload
```go
// Reload plugins without restart
func (m *Manager) Reload(name string) error {
    m.Stop(name)
    return m.Start(name)
}
```