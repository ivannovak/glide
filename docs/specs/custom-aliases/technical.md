# YAML-Defined Commands - Technical Specification

## Overview

This document outlines the technical implementation of YAML-defined commands in Glide, allowing users to define custom commands in configuration files that become first-class citizens alongside plugin-provided commands.

## Architecture Changes

### 1. Configuration Structure

#### Update `internal/config/types.go`

Add new types to support command definitions:

```go
// Command represents a user-defined command
type Command struct {
    // The actual command(s) to execute
    Cmd string `yaml:"cmd"`

    // Optional fields for structured format
    Alias       string `yaml:"alias,omitempty"`
    Description string `yaml:"description,omitempty"`
    Help        string `yaml:"help,omitempty"`
    Category    string `yaml:"category,omitempty"`
}

// CommandMap handles both simple string and structured Command formats
type CommandMap map[string]interface{}

// Config structure additions
type Config struct {
    // ... existing fields ...
    Commands CommandMap `yaml:"commands,omitempty"`
}

// ProjectConfig structure additions
type ProjectConfig struct {
    // ... existing fields ...
    Commands CommandMap `yaml:"commands,omitempty"`
}
```

### 2. Command Parsing

#### Create `internal/config/commands.go`

New file to handle command parsing and normalization:

```go
package config

import (
    "fmt"
    "strings"
)

// ParseCommands converts CommandMap to normalized Command structures
func ParseCommands(rawCommands CommandMap) (map[string]*Command, error) {
    commands := make(map[string]*Command)

    for name, value := range rawCommands {
        cmd, err := parseCommand(name, value)
        if err != nil {
            return nil, fmt.Errorf("error parsing command %s: %w", name, err)
        }
        commands[name] = cmd
    }

    return commands, nil
}

// parseCommand handles both string and map formats
func parseCommand(name string, value interface{}) (*Command, error) {
    switch v := value.(type) {
    case string:
        // Simple format: commands: build: "docker build"
        return &Command{Cmd: v}, nil

    case map[string]interface{}:
        // Structured format with additional properties
        cmd := &Command{}

        // Parse required cmd field
        if cmdStr, ok := v["cmd"].(string); ok {
            cmd.Cmd = cmdStr
        } else if cmdStr, ok := v["cmd"].(string); ok {
            // Handle pipe syntax (multi-line)
            cmd.Cmd = cmdStr
        } else {
            return nil, fmt.Errorf("command must have 'cmd' field")
        }

        // Parse optional fields
        if alias, ok := v["alias"].(string); ok {
            cmd.Alias = alias
        }
        if desc, ok := v["description"].(string); ok {
            cmd.Description = desc
        }
        if help, ok := v["help"].(string); ok {
            cmd.Help = help
        }
        if cat, ok := v["category"].(string); ok {
            cmd.Category = cat
        }

        return cmd, nil

    default:
        return nil, fmt.Errorf("invalid command format for %s", name)
    }
}

// ExpandCommand prepares a command for execution
func ExpandCommand(cmd string, args []string) string {
    // Replace positional parameters
    expanded := cmd
    for i, arg := range args {
        placeholder := fmt.Sprintf("$%d", i+1)
        expanded = strings.ReplaceAll(expanded, placeholder, arg)
    }

    // Handle $@ for all arguments
    if strings.Contains(expanded, "$@") {
        expanded = strings.ReplaceAll(expanded, "$@", strings.Join(args, " "))
    }

    return expanded
}
```

### 3. Command Registry Integration

#### Update `internal/cli/registry.go`

Add support for YAML-defined commands:

```go
// Add new category for YAML-defined commands
const (
    // ... existing categories ...
    CategoryYAML Category = "yaml" // User-defined YAML commands
)

// AddYAMLCommand registers a YAML-defined command
func (r *Registry) AddYAMLCommand(name string, cmd *config.Command) error {
    // Create a factory that builds a cobra command from the YAML definition
    factory := func() *cobra.Command {
        cobraCmd := &cobra.Command{
            Use:   name,
            Short: cmd.Description,
            Long:  cmd.Help,
            RunE: func(c *cobra.Command, args []string) error {
                // Execute the YAML-defined command
                return executeYAMLCommand(cmd.Cmd, args)
            },
        }

        // Set alias if defined
        if cmd.Alias != "" {
            cobraCmd.Aliases = []string{cmd.Alias}
        }

        return cobraCmd
    }

    // Determine category
    category := CategoryYAML
    if cmd.Category != "" {
        // Map to existing category if possible
        category = Category(cmd.Category)
    }

    metadata := Metadata{
        Name:        name,
        Category:    category,
        Description: cmd.Description,
        Aliases:     []string{cmd.Alias},
    }

    return r.Register(name, factory, metadata)
}
```

### 4. Command Execution

#### Create `internal/cli/yaml_executor.go`

Handle execution of YAML-defined commands:

```go
package cli

import (
    "bufio"
    "fmt"
    "os"
    "os/exec"
    "strings"

    "github.com/glide-cli/glide/internal/config"
)

// executeYAMLCommand runs a YAML-defined command
func executeYAMLCommand(cmdStr string, args []string) error {
    // Expand parameters
    expanded := config.ExpandCommand(cmdStr, args)

    // Check for multi-line commands (contains newlines or &&)
    if strings.Contains(expanded, "\n") || strings.Contains(expanded, "&&") {
        return executeMultiCommand(expanded)
    }

    // Single command execution
    return executeSingleCommand(expanded)
}

// executeSingleCommand runs a single command line
func executeSingleCommand(cmdStr string) error {
    // Check if it's a glide command (recursive call)
    if strings.HasPrefix(cmdStr, "glide") || strings.HasPrefix(cmdStr, "glide ") {
        // Extract command and args
        parts := strings.Fields(cmdStr)
        if len(parts) > 1 {
            // Execute via Glide's own command system
            return executeGlideCommand(parts[1:])
        }
    }

    // Execute as shell command
    cmd := exec.Command("sh", "-c", cmdStr)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Stdin = os.Stdin

    return cmd.Run()
}

// executeMultiCommand handles multi-line or chained commands
func executeMultiCommand(cmdStr string) error {
    // Split by newlines and &&
    var commands []string

    // Handle newline-separated commands
    lines := strings.Split(cmdStr, "\n")
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line == "" {
            continue
        }

        // Further split by &&
        if strings.Contains(line, "&&") {
            parts := strings.Split(line, "&&")
            for _, part := range parts {
                part = strings.TrimSpace(part)
                if part != "" {
                    commands = append(commands, part)
                }
            }
        } else {
            commands = append(commands, line)
        }
    }

    // Execute commands in sequence
    for _, cmd := range commands {
        fmt.Printf("→ %s\n", cmd)
        if err := executeSingleCommand(cmd); err != nil {
            return fmt.Errorf("command failed: %s: %w", cmd, err)
        }
    }

    return nil
}

// executeGlideCommand recursively executes a Glide command
func executeGlideCommand(args []string) error {
    // This would integrate with the existing Glide command execution
    // to avoid spawning a new process
    // Implementation depends on existing command execution architecture
    return nil
}
```

### 5. Configuration Discovery

#### Create `internal/config/discovery.go`

Implement recursive configuration discovery similar to plugin discovery:

```go
package config

import (
    "os"
    "path/filepath"
    "github.com/glide-cli/glide/pkg/branding"
    "gopkg.in/yaml.v3"
)

// DiscoverConfigs finds all configuration files up the directory tree
func DiscoverConfigs(startDir string) ([]string, error) {
    var configs []string

    // Get home directory to stop searching there
    home, _ := os.UserHomeDir()

    // Walk up the directory tree
    current := startDir
    for {
        // Check if we've reached root or home directory
        if current == "/" || current == home || current == filepath.Dir(current) {
            break
        }

        // Check for configuration file in this directory
        // Use the branded config filename from branding package
        configPath := filepath.Join(current, branding.ConfigFileName)
        if _, err := os.Stat(configPath); err == nil {
            configs = append(configs, configPath)
        }

        // Check if we've reached project root (has .git)
        gitPath := filepath.Join(current, ".git")
        if _, err := os.Stat(gitPath); err == nil {
            // Stop here - we've found project root
            break
        }

        // Move up to parent directory
        current = filepath.Dir(current)
    }

    // Reverse the order so deepest configs come first (highest priority)
    for i, j := 0, len(configs)-1; i < j; i, j = i+1, j-1 {
        configs[i], configs[j] = configs[j], configs[i]
    }

    return configs, nil
}

// LoadAndMergeConfigs loads multiple config files and merges them
func LoadAndMergeConfigs(configPaths []string) (*Config, error) {
    merged := &Config{
        Commands: make(CommandMap),
    }

    // Load configs in reverse order (lowest priority first)
    // so that higher priority configs override
    for i := len(configPaths) - 1; i >= 0; i-- {
        data, err := os.ReadFile(configPaths[i])
        if err != nil {
            continue // Skip configs that can't be read
        }

        var cfg Config
        if err := yaml.Unmarshal(data, &cfg); err != nil {
            continue // Skip invalid configs
        }

        // Merge commands (later configs override earlier ones)
        if cfg.Commands != nil {
            for name, cmd := range cfg.Commands {
                merged.Commands[name] = cmd
            }
        }
    }

    return merged, nil
}
```

### 6. Command Loading Priority

#### Update `cmd/glide/main.go`

Implement the priority order for command loading with recursive discovery:

```go
func loadCommands(registry *cli.Registry, ctx *context.ProjectContext) error {
    // 1. Core commands are already registered (highest priority)

    // 2. Discover and load all .glide.yml files up the tree
    cwd, _ := os.Getwd()
    configPaths, err := config.DiscoverConfigs(cwd)
    if err == nil && len(configPaths) > 0 {
        localConfigs, err := config.LoadAndMergeConfigs(configPaths)
        if err == nil && localConfigs.Commands != nil {
            commands, err := config.ParseCommands(localConfigs.Commands)
            if err != nil {
                return fmt.Errorf("error parsing discovered commands: %w", err)
            }
            for name, cmd := range commands {
                // Check for conflicts with core commands
                if !isProtectedCommand(name) {
                    registry.AddYAMLCommand(name, cmd)
                }
            }
        }
    }

    // 3. Load plugin-bundled YAML commands
    loadPluginYAMLCommands(registry)

    // 4. Load plugin Go commands (already handled by plugin system)

    // 5. Load global commands (~/.glide/config.yml) - lowest priority
    if globalConfig, err := loadGlobalConfig(); err == nil {
        if globalConfig.Commands != nil {
            commands, err := config.ParseCommands(globalConfig.Commands)
            if err != nil {
                return fmt.Errorf("error parsing global commands: %w", err)
            }
            for name, cmd := range commands {
                // Only add if not already defined
                if _, exists := registry.Get(name); !exists {
                    registry.AddYAMLCommand(name, cmd)
                }
            }
        }
    }

    return nil
}

// isProtectedCommand checks if a command name is protected (core command)
func isProtectedCommand(name string) bool {
    protected := []string{
        "help", "setup", "plugins", "self-update",
        "version", "completion", "global",
    }
    for _, p := range protected {
        if name == p {
            return true
        }
    }
    return false
}
```

### 6. Plugin YAML Command Loading

#### Update `pkg/plugin/runtime.go`

Add support for loading embedded commands.yml from plugins:

```go
// LoadPluginYAMLCommands loads YAML commands from plugin bundles
func LoadPluginYAMLCommands(pluginPath string) (map[string]*config.Command, error) {
    // Check for embedded commands.yml
    commandsPath := filepath.Join(pluginPath, "commands.yml")

    // For binary plugins, we might need to extract or use a convention
    // For now, assume plugins can have a commands.yml alongside
    if _, err := os.Stat(commandsPath); os.IsNotExist(err) {
        return nil, nil // No YAML commands
    }

    data, err := os.ReadFile(commandsPath)
    if err != nil {
        return nil, fmt.Errorf("error reading plugin commands.yml: %w", err)
    }

    var rawCommands struct {
        Commands config.CommandMap `yaml:"commands"`
    }

    if err := yaml.Unmarshal(data, &rawCommands); err != nil {
        return nil, fmt.Errorf("error parsing plugin commands.yml: %w", err)
    }

    return config.ParseCommands(rawCommands.Commands)
}
```

## Implementation Phases

### Phase 1: Core Infrastructure
1. Update configuration types to support commands
2. Implement command parsing for simple format
3. Create basic command execution

### Phase 2: Advanced Features
1. Add structured command format support
2. Implement parameter substitution
3. Add multi-line command support

### Phase 3: Integration
1. Integrate with command registry
2. Implement priority ordering
3. Add plugin YAML command loading

### Phase 4: Polish
1. Add command validation
2. Improve error messages
3. Update help system to show YAML commands

## Testing Strategy

### Unit Tests

1. **Config Parsing Tests** (`internal/config/commands_test.go`)
   - Test simple command format parsing
   - Test structured command format parsing
   - Test invalid format handling
   - Test parameter expansion

2. **Execution Tests** (`internal/cli/yaml_executor_test.go`)
   - Test single command execution
   - Test multi-line command execution
   - Test parameter substitution
   - Test environment variable expansion

3. **Registry Tests** (`internal/cli/registry_test.go`)
   - Test YAML command registration
   - Test alias handling
   - Test priority ordering

### Integration Tests

1. **End-to-End Tests**
   - Test loading commands from project config
   - Test loading commands from global config
   - Test command priority resolution
   - Test plugin YAML commands

2. **Conflict Resolution Tests**
   - Test core command protection
   - Test override behavior
   - Test alias conflicts

## Error Handling

### Validation Errors
- Invalid YAML syntax
- Missing required fields (cmd)
- Circular command references
- Invalid command syntax

### Runtime Errors
- Command execution failures
- Missing executables
- Permission issues
- Parameter substitution errors

### User-Friendly Messages
```
Error: Command 'build' failed to execute
  Command: docker build --no-cache
  Error: docker: command not found

  Ensure Docker is installed and in your PATH
```

## Security Considerations

1. **Command Injection Prevention**
   - Properly quote and escape parameters
   - Validate command strings
   - Limit recursive command depth

2. **Environment Variable Safety**
   - Only expand known variables
   - Avoid exposing sensitive variables
   - Log command execution for audit

3. **File System Access**
   - Commands run with user permissions
   - No elevation of privileges
   - Respect project boundaries

## Performance Considerations

1. **Command Parsing**
   - Parse commands once at startup
   - Cache parsed command structures
   - Lazy load plugin commands

2. **Execution Optimization**
   - Use shell built-ins where possible
   - Avoid unnecessary process spawning
   - Batch multi-line commands efficiently

## Migration Path

No migration needed - this is a new feature. Existing users can:
1. Continue using Glide without YAML commands
2. Gradually adopt YAML commands as needed
3. Override plugin commands selectively

## Success Metrics

1. Command definition and execution works correctly
2. Multi-line commands execute in sequence
3. Parameters are substituted properly
4. Priority ordering is respected
5. Plugin YAML commands load successfully
6. Help system shows all available commands
7. Performance impact is minimal (<50ms startup overhead)
