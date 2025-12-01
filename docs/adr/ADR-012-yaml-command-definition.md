# ADR 012: YAML Command Definition

## Status
Accepted

## Context
Glide provides a plugin system for extending functionality, but creating plugins requires:
- Writing Go code
- Compiling binaries
- Understanding the plugin SDK
- Distributing plugin files

Many common use cases involve simple command aliases, shortcuts, or workflows that don't justify the overhead of creating a full plugin. Teams need a way to:
- Standardize common workflows across developers
- Create project-specific commands without writing Go
- Share command definitions via version control
- Override or extend plugin-provided commands

## Decision
We will implement YAML-defined commands that allow users to define custom commands directly in Glide configuration files. These commands will be first-class citizens alongside plugin-provided commands.

### Command Format
Support two formats for maximum flexibility:

#### Simple Format
```yaml
commands:
  build: docker build --no-cache
  test: go test ./...
```

#### Structured Format
```yaml
commands:
  deploy:
    cmd: |
      go test ./...
      go build -o app
      ./deploy.sh
    alias: d
    description: "Deploy the application"
    help: |
      Detailed help text explaining usage,
      options, and examples.
    category: deployment
```

### Features
1. **Parameter Substitution**: Support `$1`, `$2`, `$@`, `$*` for arguments
2. **Multi-line Commands**: Execute sequences with progress indicators
3. **Environment Variables**: Expand environment variables in commands
4. **Command Aliases**: Alternative names for commands
5. **Recursive Execution**: Support calling other Glide commands
6. **Plugin Integration**: Plugins can bundle `commands.yml` files

### Priority Hierarchy
When multiple sources define the same command:
1. Built-in core commands (cannot be overridden)
2. Current directory commands (`./.glide.yml`)
3. Parent directory commands (recursive discovery)
4. Project root commands
5. Plugin-bundled YAML commands
6. Plugin Go commands
7. Global user commands (`~/.glide/config.yml`)

## Consequences

### Positive
- **Low Barrier to Entry**: No programming required
- **Version Control**: Commands tracked with the project
- **Team Standardization**: Shared workflows via config files
- **Rapid Iteration**: Modify commands without recompilation
- **Progressive Enhancement**: Start simple, add structure as needed
- **Override Capability**: Users can customize plugin commands locally

### Negative
- **Limited Logic**: No conditionals, loops, or complex logic
- **Security Concerns**: Executing arbitrary shell commands
- **Debugging Difficulty**: No IDE support or syntax checking
- **Performance**: Shell invocation overhead
- **Portability**: Shell commands may not work across platforms

### Mitigation Strategies
- Validate commands before execution
- Detect circular references
- Provide clear error messages with command context
- Document platform-specific considerations
- Suggest plugin development for complex logic

## Implementation Details

### Command Parsing
```go
type Command struct {
    Cmd         string `yaml:"cmd"`
    Alias       string `yaml:"alias,omitempty"`
    Description string `yaml:"description,omitempty"`
    Help        string `yaml:"help,omitempty"`
    Category    string `yaml:"category,omitempty"`
}

type CommandMap map[string]interface{}

func ParseCommands(raw CommandMap) (map[string]*Command, error) {
    // Handle both string and structured formats
    // Validate required fields
    // Return normalized Command structures
}
```

### Command Execution
```go
func ExecuteYAMLCommand(cmdStr string, args []string) error {
    // Expand parameters
    expanded := ExpandCommand(cmdStr, args)

    // Execute as a shell script
    // This properly handles:
    // - Single commands
    // - Multi-line scripts
    // - Pipes and redirects
    // - Control structures (if/then/else, loops, etc.)
    // - Shell built-ins and functions
    return executeShellCommand(expanded)
}
```

All YAML commands are executed through `sh -c`, ensuring consistent behavior whether they are single-line commands or complex multi-line shell scripts. This approach preserves shell semantics including control flow, variable scope, and pipe behaviors.

### Registry Integration
YAML commands are registered in the same registry as plugin commands:
```go
func (r *Registry) AddYAMLCommand(name string, cmd *Command) error {
    factory := func() *cobra.Command {
        return &cobra.Command{
            Use:   name,
            Short: cmd.Description,
            Long:  cmd.Help,
            RunE: func(c *cobra.Command, args []string) error {
                return ExecuteYAMLCommand(cmd.Cmd, args)
            },
        }
    }
    return r.Register(name, factory, metadata)
}
```

## Examples

### Team Workflow Standardization
```yaml
commands:
  # Development workflow
  fresh:
    cmd: |
      git pull
      glideup
      glidedb migrate
    alias: f
    description: "Update and restart everything"

  # Testing workflow
  test-all:
    cmd: |
      glidetest unit
      glidetest integration
      glidetest e2e
    description: "Run complete test suite"

  # Deployment
  ship:
    cmd: |
      glidetest-all
      glidebuild --production
      glidedeploy $1 --confirm
    description: "Test, build, and deploy"
```

### Project-Specific Commands
```yaml
commands:
  # Database operations
  db-reset: glidedb drop && glidedb create && glidedb migrate && glidedb seed

  # Docker shortcuts
  rebuild: glide down && docker build --no-cache . && glideup

  # Git workflows
  sync: git pull --rebase && git push
```

### Plugin Command Overrides
```yaml
# Override plugin-provided 'test' command
commands:
  test:
    cmd: |
      echo "Running custom test sequence..."
      go test -v -race ./...
      golangci-lint run
    description: "Enhanced test with race detection and linting"
```

## Alternatives Considered

### Shell Aliases Only
Use shell aliases or functions instead of Glide commands.
- **Pros**: Native shell features, no overhead
- **Cons**: Not portable, not versioned with project

### Embedded Scripting Language
Embed Lua, JavaScript, or another scripting language.
- **Pros**: Full programming capabilities
- **Cons**: Complexity, dependencies, learning curve

### Makefile Integration
Use Makefiles for command definitions.
- **Pros**: Familiar to many developers
- **Cons**: Make-specific syntax, requires Make installation

### External Script Files
Reference external script files from config.
- **Pros**: Full shell/script capabilities
- **Cons**: Multiple files to manage, path resolution issues

## Security Considerations

### Command Injection
- Commands are executed via `sh -c`, which could allow injection
- Mitigation: Properly quote and escape parameters
- Document risks of untrusted configurations

### Path Traversal
- Commands could access files outside project
- Mitigation: None (by design - users need filesystem access)
- Document best practices for command isolation

### Resource Consumption
- Commands could consume excessive resources
- Mitigation: Consider timeout options in future versions
- Monitor for abuse patterns

## Future Enhancements
- Conditional execution based on environment
- Command dependencies and prerequisites
- Built-in functions for common operations
- Dry-run mode for testing commands
- Command composition and inheritance
- Platform-specific command variants

## References
- Command Types: `internal/config/types.go`
- Command Parser: `internal/config/commands.go`
- Command Executor: `internal/cli/yaml_executor.go`
- Registry Integration: `internal/cli/registry.go`
- Example Configuration: `.glide.example.yml`
