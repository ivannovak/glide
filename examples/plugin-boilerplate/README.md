# Glide Plugin Boilerplate

A starter template for creating Glide runtime plugins.

## 🚀 Quick Start

### 1. Copy this Template

```bash
# Copy to a new directory
cp -r examples/plugin-boilerplate ~/my-glide-plugin
cd ~/my-glide-plugin
```

### 2. Customize Your Plugin

Edit `main.go` and update:
- Plugin name and metadata in `GetMetadata()`
- Commands in `ListCommands()`
- Command implementations in `ExecuteCommand()`

### 3. Update Module Name

Edit `go.mod`:
```go
module github.com/yourname/glide-plugin-yourplugin
```

### 4. Build Your Plugin

```bash
# Install dependencies
go mod tidy

# Build the plugin (name must start with glide-plugin-)
go build -o glide-plugin-yourname

# Or use make
make build
```

### 5. Install Plugin

```bash
# Create plugins directory if it doesn't exist
mkdir -p ~/.glide/plugins

# Copy plugin to installation directory
cp glide-plugin-yourname ~/.glide/plugins/
chmod +x ~/.glide/plugins/glide-plugin-yourname

# Or use make
make install
```

### 6. Test Your Plugin

```bash
# List all plugins (verify yours appears)
glide plugins list

# Get plugin info
glide plugins info yourname

# Test your commands
glide yourname hello
glide yourname config
glide yourname interactive

# Test with aliases (if configured)
glide mp h            # Plugin alias 'mp' + command alias 'h' for hello
glide myp c           # Plugin alias 'myp' + command alias 'c' for config
glide mp i            # Plugin alias 'mp' + command alias 'i' for interactive
```

## 📁 Project Structure

```
your-plugin/
├── main.go           # Plugin implementation
├── go.mod            # Go module definition
├── go.sum            # Dependency lock file (generated)
├── README.md         # This file
├── Makefile          # Build automation
└── .gitignore        # Git ignore rules
```

## 🛠️ Plugin Development

### Basic Command

```go
func (p *MyPlugin) executeHello(ctx context.Context, req *sdk.ExecuteRequest) (*sdk.ExecuteResponse, error) {
    return &sdk.ExecuteResponse{
        Success:  true,
        Stdout:   []byte("Hello, World!\n"),
        ExitCode: 0,
    }, nil
}
```

### Command with Arguments

```go
func (p *MyPlugin) executeGreet(ctx context.Context, req *sdk.ExecuteRequest) (*sdk.ExecuteResponse, error) {
    name := "World"
    if len(req.Args) > 0 {
        name = req.Args[0]
    }
    
    return &sdk.ExecuteResponse{
        Success:  true,
        Stdout:   []byte(fmt.Sprintf("Hello, %s!\n", name)),
        ExitCode: 0,
    }, nil
}
```

### Interactive Command

```go
func (p *MyPlugin) StartInteractive(req *sdk.InteractiveRequest, stream sdk.GlidePlugin_StartInteractiveServer) error {
    // Handle interactive I/O through stream
    // See main.go for a complete example
}
```

### Using Configuration

Add to `.glide.yml`:
```yaml
plugins:
  yourplugin:
    api_key: "abc123"
    endpoint: "https://api.example.com"
    timeout: 30
```

Access in plugin:
```go
func (p *MyPlugin) executeWithConfig(ctx context.Context, req *sdk.ExecuteRequest) (*sdk.ExecuteResponse, error) {
    apiKey := p.config["api_key"].(string)
    endpoint := p.config["endpoint"].(string)
    // Use configuration values...
}
```

## 🔤 Using Aliases

Aliases provide shortcuts for both plugin names and commands, allowing users to type less while maintaining clarity.

### Plugin-Level Aliases

Define plugin aliases in your `GetMetadata()` function:

```go
func (p *MyPlugin) GetMetadata(ctx context.Context, _ *sdk.Empty) (*sdk.PluginMetadata, error) {
    return &sdk.PluginMetadata{
        Name:        "myplugin",
        Tags:        []string{"mp", "myp"},  // Plugin aliases
        // ... other metadata
    }, nil
}
```

Users can now use:
- `glide myplugin hello` (full name)
- `glide mp hello` (using alias)
- `glide myp hello` (using another alias)

### Command-Level Aliases

Define command aliases in your `ListCommands()` function:

```go
Commands: []*sdk.CommandInfo{
    {
        Name:        "hello",
        Aliases:     []string{"h", "hi"},  // Command aliases
        Description: "Print a greeting message",
        // ... other properties
    },
}
```

### Handling Aliases in ExecuteCommand

Your `ExecuteCommand` function should handle both full names and aliases:

```go
func (p *MyPlugin) ExecuteCommand(ctx context.Context, req *sdk.ExecuteRequest) (*sdk.ExecuteResponse, error) {
    // Map aliases to their full command names
    commandMap := map[string]string{
        "hello": "hello",
        "h":     "hello",
        "hi":    "hello",
        // ... more mappings
    }
    
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
    // ... handle other commands
    }
}
```

### Combined Usage

With both plugin and command aliases, users can use very short invocations:
- `glide mp h World` instead of `glide myplugin hello World`
- `glide myp cfg` instead of `glide myplugin config`

## 📋 Available Interfaces

### Plugin Metadata
- `Name`: Plugin identifier (lowercase, no spaces)
- `Version`: Semantic version (e.g., "1.0.0")
- `Author`: Your name or organization
- `Description`: Brief description
- `Tags`: Plugin-level aliases (array of strings)
- `Homepage`: Optional project URL
- `License`: Optional license identifier
- `MinSdk`: Minimum SDK version required

### Command Info
- `Name`: Command name (lowercase, no spaces)
- `Description`: Brief description
- `Category`: Command category for grouping
- `Interactive`: Whether command needs TTY
- `Hidden`: Hide from help output
- `Aliases`: Alternative command names

### Execute Request
- `Command`: The command being executed
- `Args`: Command line arguments
- `Flags`: Command flags (if implemented)
- `WorkingDir`: Current working directory
- `Environment`: Environment variables

### Execute Response
- `Success`: Whether command succeeded
- `Stdout`: Standard output bytes
- `Stderr`: Standard error bytes
- `Error`: Error message if failed
- `ExitCode`: Process exit code
- `RequiresInteractive`: Request interactive mode

## 🐛 Debugging

### Enable Debug Output

```bash
export GLIDE_PLUGIN_DEBUG=1
glide yourname hello
```

### Test Plugin Directly

```bash
# This will show the plugin handshake (should fail without host)
./glide-plugin-yourname
```

### Check Plugin Logs

Debug output appears in stderr when `GLIDE_PLUGIN_DEBUG=1` is set.

## 📚 Examples

### Docker Integration

```go
func (p *MyPlugin) executeDocker(ctx context.Context, req *sdk.ExecuteRequest) (*sdk.ExecuteResponse, error) {
    // Example: Run a Docker command
    cmd := exec.Command("docker", "ps", "-a")
    output, err := cmd.CombinedOutput()
    
    if err != nil {
        return &sdk.ExecuteResponse{
            Success:  false,
            Stderr:   output,
            Error:    err.Error(),
            ExitCode: 1,
        }, nil
    }
    
    return &sdk.ExecuteResponse{
        Success:  true,
        Stdout:   output,
        ExitCode: 0,
    }, nil
}
```

### File Operations

```go
func (p *MyPlugin) executeReadFile(ctx context.Context, req *sdk.ExecuteRequest) (*sdk.ExecuteResponse, error) {
    if len(req.Args) < 1 {
        return &sdk.ExecuteResponse{
            Success: false,
            Error:   "filename required",
        }, nil
    }
    
    content, err := os.ReadFile(req.Args[0])
    if err != nil {
        return &sdk.ExecuteResponse{
            Success: false,
            Error:   err.Error(),
        }, nil
    }
    
    return &sdk.ExecuteResponse{
        Success:  true,
        Stdout:   content,
        ExitCode: 0,
    }, nil
}
```

### HTTP Requests

```go
func (p *MyPlugin) executeHTTP(ctx context.Context, req *sdk.ExecuteRequest) (*sdk.ExecuteResponse, error) {
    resp, err := http.Get("https://api.example.com/data")
    if err != nil {
        return &sdk.ExecuteResponse{
            Success: false,
            Error:   err.Error(),
        }, nil
    }
    defer resp.Body.Close()
    
    body, _ := io.ReadAll(resp.Body)
    
    return &sdk.ExecuteResponse{
        Success:  true,
        Stdout:   body,
        ExitCode: 0,
    }, nil
}
```

## 🔧 Makefile Commands

- `make build` - Build the plugin binary
- `make install` - Install to ~/.glide/plugins/
- `make clean` - Remove built binary
- `make test` - Test the plugin
- `make dev` - Build and install for development
- `make release` - Build optimized release binary

## 📦 Distribution

### Binary Distribution

```bash
# Build for multiple platforms
GOOS=darwin GOARCH=amd64 go build -o glide-plugin-yourname-darwin-amd64
GOOS=darwin GOARCH=arm64 go build -o glide-plugin-yourname-darwin-arm64
GOOS=linux GOARCH=amd64 go build -o glide-plugin-yourname-linux-amd64
```

### GitHub Release

1. Tag your release:
```bash
git tag v1.0.0
git push origin v1.0.0
```

2. Create GitHub release with binaries

3. Users can install with:
```bash
# Download and install
curl -L https://github.com/you/plugin/releases/download/v1.0.0/glide-plugin-yourname-darwin-arm64 \
  -o ~/.glide/plugins/glide-plugin-yourname
chmod +x ~/.glide/plugins/glide-plugin-yourname
```

## 🤝 Contributing

1. Fork this repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## 📄 License

This boilerplate is provided as-is for use in creating Glide plugins.

## 🔗 Resources

- [Glide Documentation](https://github.com/ivannovak/glide)
- [Hashicorp go-plugin](https://github.com/hashicorp/go-plugin)
- [Protocol Buffers](https://protobuf.dev/)
- [gRPC Documentation](https://grpc.io/docs/)

## 💡 Tips

1. **Keep plugins focused** - Do one thing well
2. **Handle errors gracefully** - Return clear error messages
3. **Document commands** - Good descriptions help users
4. **Version properly** - Use semantic versioning
5. **Test thoroughly** - Plugins run in production environments
6. **Minimize dependencies** - Smaller binaries load faster
7. **Use configuration** - Make plugins flexible
8. **Provide examples** - Help users get started quickly

## 🆘 Troubleshooting

### Plugin Not Found
- Verify plugin is in `~/.glide/plugins/`
- Check filename starts with `glide-plugin-`
- Ensure file has execute permissions (`chmod +x`)

### Plugin Won't Load
- Check plugin was built with compatible Go version
- Verify SDK version compatibility
- Enable debug mode to see errors

### Commands Not Working
- Ensure command is listed in `ListCommands()`
- Check command name matches in `ExecuteCommand()`
- Verify response structure is correct

### Interactive Mode Issues
- Confirm command has `Interactive: true`
- Check `StartInteractive` implementation
- Verify stream handling is correct

---

Happy plugin development! 🚀