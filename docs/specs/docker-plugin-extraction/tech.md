# Docker Plugin Extraction - Technical Specification

## Architecture Overview

### Current State
```
glide-core/
├── internal/
│   ├── docker/          # Docker implementation (TO EXTRACT)
│   ├── context/         # Has Docker fields (TO MODIFY)
│   ├── shell/           # Has DockerExecutor (TO MODIFY)
│   └── cli/
│       ├── docker.go    # Docker command (TO EXTRACT)
│       └── project.go   # Has Docker commands (TO MODIFY)
└── pkg/
    └── plugin/
        └── sdk/         # Plugin SDK (TO ENHANCE)
```

### Target State
```
glide-core/
├── internal/
│   ├── context/         # Extended via plugins
│   ├── shell/           # Plugin-aware executors
│   └── cli/             # Loads plugin commands
├── pkg/
│   └── plugin/
│       └── sdk/         # Enhanced SDK with context extensions
└── plugins/
    └── docker/          # Docker plugin (NEW)
        ├── detector.go
        ├── plugin.go
        ├── commands/
        ├── compose/
        ├── container/
        └── shell/
```

## Core Enhancements Required

### 1. Context Extension System

#### Interface Definition
```go
// pkg/plugin/sdk/context.go
package sdk

// ContextExtension allows plugins to extend ProjectContext
type ContextExtension interface {
    // Name of the extension (e.g., "docker")
    Name() string

    // Detect performs detection and returns extension data
    Detect(projectPath string) (map[string]interface{}, error)

    // Priority for ordering multiple extensions
    Priority() int
}

// ContextProvider is implemented by plugins that extend context
type ContextProvider interface {
    Plugin
    GetContextExtensions() []ContextExtension
}
```

#### Context Modification
```go
// internal/context/types.go
type ProjectContext struct {
    // ... existing fields ...

    // Extensions holds plugin-provided context data
    Extensions map[string]interface{} `json:"extensions,omitempty"`

    // DEPRECATED: Kept for compatibility, populated from Extensions["docker"]
    DockerRunning   bool     `json:"docker_running,omitempty"`
    ComposeFiles    []string `json:"compose_files,omitempty"`
    ComposeOverride string   `json:"compose_override,omitempty"`
}

// GetDockerContext returns Docker-specific context (compatibility layer)
func (c *ProjectContext) GetDockerContext() *DockerContext {
    if c.Extensions == nil {
        return nil
    }
    if docker, ok := c.Extensions["docker"].(map[string]interface{}); ok {
        return &DockerContext{
            Running:       docker["running"].(bool),
            ComposeFiles:  docker["compose_files"].([]string),
            Override:      docker["override"].(string),
        }
    }
    return nil
}
```

### 2. Plugin-Aware Shell Executors

```go
// internal/shell/plugin_executor.go
package shell

// ExecutorProvider allows plugins to provide specialized executors
type ExecutorProvider interface {
    // Name of the executor (e.g., "docker")
    Name() string

    // CreateExecutor creates a specialized executor
    CreateExecutor(ctx *context.ProjectContext) Executor

    // CanHandle returns true if this provider can handle the context
    CanHandle(ctx *context.ProjectContext) bool
}

// ExecutorRegistry manages plugin-provided executors
type ExecutorRegistry struct {
    providers map[string]ExecutorProvider
}

func (r *ExecutorRegistry) GetExecutor(name string, ctx *context.ProjectContext) Executor {
    if provider, ok := r.providers[name]; ok && provider.CanHandle(ctx) {
        return provider.CreateExecutor(ctx)
    }
    return NewExecutor(Options{})
}
```

### 3. Configuration Schema Extension

```go
// pkg/plugin/sdk/config.go
package sdk

// ConfigSchema defines plugin configuration structure
type ConfigSchema struct {
    Section     string                 // e.g., "docker"
    Fields      map[string]FieldSchema // Field definitions
    Defaults    map[string]interface{} // Default values
}

type FieldSchema struct {
    Type        string   // "string", "int", "bool", "array"
    Description string   // Help text
    Required    bool     // Is field required
    Default     interface{} // Default value
    Validate    func(interface{}) error // Custom validation
}

// ConfigProvider is implemented by plugins with configuration
type ConfigProvider interface {
    Plugin
    GetConfigSchema() ConfigSchema
    ValidateConfig(config map[string]interface{}) error
}
```

## Docker Plugin Implementation

### Plugin Structure
```go
// plugins/docker/plugin.go
package docker

import (
    "github.com/glide-cli/glide/pkg/plugin/sdk"
)

type DockerPlugin struct {
    *sdk.BasePlugin
    detector    *DockerDetector
    commands    *CommandRegistry
    config      *Config
}

func New() sdk.Plugin {
    p := &DockerPlugin{
        BasePlugin: sdk.NewBasePlugin("docker", "1.0.0"),
    }
    p.detector = NewDockerDetector()
    p.commands = NewCommandRegistry()
    return p
}

func (p *DockerPlugin) Initialize() error {
    // Register context extension
    p.RegisterContextExtension(p.detector)

    // Register shell executor
    p.RegisterExecutorProvider(NewDockerExecutorProvider())

    // Register completion providers
    p.RegisterCompletionProvider(NewContainerCompleter())
    p.RegisterCompletionProvider(NewServiceCompleter())

    return nil
}

func (p *DockerPlugin) GetCommands() []sdk.CommandDefinition {
    return p.commands.GetAll()
}
```

### Docker Detector
```go
// plugins/docker/detector.go
package docker

type DockerDetector struct {
    *sdk.BaseFrameworkDetector
}

func NewDockerDetector() *DockerDetector {
    detector := &DockerDetector{
        BaseFrameworkDetector: sdk.NewBaseFrameworkDetector(sdk.FrameworkInfo{
            Name: "docker",
            Type: "container",
        }),
    }

    detector.SetPatterns(sdk.DetectionPatterns{
        RequiredFiles: []string{},  // Docker is optional
        OptionalFiles: []string{
            "docker-compose.yml",
            "docker-compose.yaml",
            "compose.yml",
            "compose.yaml",
            "Dockerfile",
        },
    })

    return detector
}

func (d *DockerDetector) Detect(projectPath string) (*sdk.DetectionResult, error) {
    result, err := d.BaseFrameworkDetector.Detect(projectPath)
    if err != nil {
        return nil, err
    }

    // Check Docker daemon
    dockerRunning := d.checkDockerDaemon()

    // Find compose files
    composeFiles := d.findComposeFiles(projectPath)
    overrideFile := d.findOverrideFile(projectPath)

    // Build context extension
    result.ContextExtension = map[string]interface{}{
        "running":       dockerRunning,
        "compose_files": composeFiles,
        "override":      overrideFile,
    }

    // Add Docker commands if Docker is available
    if dockerRunning || len(composeFiles) > 0 {
        result.Commands = d.getDockerCommands()
        result.Detected = true
        result.Confidence = 80
    }

    return result, nil
}
```

### Compatibility Layer

```go
// internal/context/compatibility.go
package context

// PopulateCompatibilityFields fills deprecated fields from extensions
func (c *ProjectContext) PopulateCompatibilityFields() {
    if c.Extensions == nil {
        return
    }

    // Handle Docker fields
    if docker, ok := c.Extensions["docker"].(map[string]interface{}); ok {
        if running, ok := docker["running"].(bool); ok {
            c.DockerRunning = running
        }
        if files, ok := docker["compose_files"].([]string); ok {
            c.ComposeFiles = files
        }
        if override, ok := docker["override"].(string); ok {
            c.ComposeOverride = override
        }
    }
}

// UpdateExtensionsFromCompatibility updates extensions from deprecated fields
func (c *ProjectContext) UpdateExtensionsFromCompatibility() {
    if c.Extensions == nil {
        c.Extensions = make(map[string]interface{})
    }

    // Preserve any Docker fields set directly (for backward compatibility)
    if c.DockerRunning || len(c.ComposeFiles) > 0 || c.ComposeOverride != "" {
        c.Extensions["docker"] = map[string]interface{}{
            "running":       c.DockerRunning,
            "compose_files": c.ComposeFiles,
            "override":      c.ComposeOverride,
        }
    }
}
```

### Command Migration

```go
// plugins/docker/commands/docker.go
package commands

import (
    "github.com/spf13/cobra"
    "github.com/glide-cli/glide/pkg/plugin/sdk"
)

func NewDockerCommand(ctx sdk.PluginContext) *cobra.Command {
    return &cobra.Command{
        Use:   "docker [docker-compose arguments]",
        Short: "Pass-through to docker-compose with automatic file resolution",
        Long:  dockerLongHelp,
        DisableFlagParsing: true,
        RunE: func(cmd *cobra.Command, args []string) error {
            return executeDockerCommand(ctx, args)
        },
    }
}

func executeDockerCommand(ctx sdk.PluginContext, args []string) error {
    // Get Docker context from extensions
    dockerCtx := ctx.ProjectContext.Extensions["docker"].(map[string]interface{})
    composeFiles := dockerCtx["compose_files"].([]string)

    // Build docker-compose command
    resolver := compose.NewResolver(composeFiles)
    dockerArgs := resolver.BuildCommand(args)

    // Execute via shell
    executor := ctx.GetExecutor("docker")
    return executor.RunPassthrough("docker", dockerArgs...)
}
```

## Migration Strategy

### Phase 1: Core Enhancements
1. Add context extension system to SDK
2. Add plugin configuration schema support
3. Make shell package plugin-aware
4. Add completion provider interface

### Phase 2: Docker Plugin Creation
1. Copy all Docker code to plugin
2. Refactor to use plugin SDK
3. Implement context extension
4. Port all commands

### Phase 3: Integration
1. Add compatibility layer to context
2. Register Docker plugin as built-in
3. Update CLI to load plugin commands
4. Wire up configuration

### Phase 4: Testing
1. Create regression test suite
2. Test all Docker commands
3. Verify context compatibility
4. Benchmark performance

### Phase 5: Cleanup
1. Remove old Docker code from core
2. Update imports
3. Update documentation
4. Tag release

## Build System Changes

```makefile
# Makefile
DOCKER_PLUGIN = plugins/docker

# Build with Docker plugin bundled
build-with-docker:
	go build -tags "docker" \
		-ldflags "-X main.builtinPlugins=docker" \
		-o glide cmd/glide/main.go

# Build without Docker (future option)
build-minimal:
	go build -o glide-minimal cmd/glide/main.go
```

```go
// cmd/glide/plugins_builtin.go
// +build docker

package main

import (
    _ "github.com/glide-cli/glide/plugins/docker"
)

func init() {
    // Docker plugin auto-registers when imported
}
```

## Testing Strategy

### Unit Tests
```go
// plugins/docker/plugin_test.go
func TestDockerPlugin_ZeroRegression(t *testing.T) {
    // Test all Docker functionality works identically
    scenarios := []struct {
        name     string
        context  *context.ProjectContext
        command  string
        args     []string
        expected string
    }{
        {
            name: "docker-compose up",
            context: &context.ProjectContext{
                DockerRunning: true,
                ComposeFiles: []string{"docker-compose.yml"},
            },
            command: "docker",
            args: []string{"up", "-d"},
            expected: "docker compose -f docker-compose.yml up -d",
        },
        // ... comprehensive test cases
    }

    for _, sc := range scenarios {
        t.Run(sc.name, func(t *testing.T) {
            // Test via plugin
            plugin := New()
            result := plugin.ExecuteCommand(sc.command, sc.args)
            assert.Equal(t, sc.expected, result)

            // Verify compatibility layer
            assert.Equal(t, sc.context.DockerRunning,
                sc.context.Extensions["docker"].(map[string]interface{})["running"])
        })
    }
}
```

### Integration Tests
```go
func TestDockerPlugin_Integration(t *testing.T) {
    // Create test project with docker-compose.yml
    // Run all Docker commands
    // Verify behavior matches current implementation
}
```

### Regression Tests
```go
func TestDockerPlugin_BackwardCompatibility(t *testing.T) {
    // Test old config format works
    // Test deprecated fields populated
    // Test existing scripts work
}
```

## Performance Considerations

### Optimization Strategies
1. **Lazy Loading**: Docker detection only runs if needed
2. **No Caching**: Detection runs fresh each time for immediate feedback
3. **Parallel Detection**: Docker detection runs alongside other detectors
4. **Command Registration**: Commands registered once at startup

### Benchmarks
```go
func BenchmarkDockerDetection(b *testing.B) {
    // Benchmark Docker detection speed
    // Target: <50ms (fresh detection each time)
}

func BenchmarkCommandExecution(b *testing.B) {
    // Benchmark Docker command execution
    // Target: No additional overhead
}
```

## Risk Mitigation

### Branch-Based Development
```bash
# All development on feature branch
git checkout -b feat/docker-plugin-extraction

# Complete implementation before merge
# No feature flags - single cutover when ready
```

### Rollback Plan
1. Development happens entirely on feature branch
2. Comprehensive testing before merge
3. If issues found, fix on branch before merge
4. Simple git revert if critical issues post-merge

## Documentation Updates

1. Update plugin development guide
2. Add Docker plugin documentation
3. Update command reference
4. Add migration guide for future plugin authors
5. Update troubleshooting guide

## Success Validation

### Automated Validation
```bash
# Regression test script
#!/bin/bash

# Test all Docker commands
glide docker up -d
glide docker ps
glide project status
glide project down

# Test context detection
glide context | grep "Docker Running"

# Test configuration
glide config get defaults.docker.compose_timeout

# Test completions
glide docker [TAB]
```

### Manual Validation Checklist
- [ ] All Docker commands work
- [ ] Context shows Docker info
- [ ] Config validates correctly
- [ ] Completions function
- [ ] Help text preserved
- [ ] Error messages unchanged
- [ ] Performance acceptable
