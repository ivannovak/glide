# Framework Detection Plugin System - Technical Specification

## Architecture Overview

Extend the existing plugin RPC interface to support framework detection capabilities. Plugins can register detection patterns, provide default commands, and enhance context detection.

## Plugin Interface Extensions

### 1. Detection Interface

```go
// pkg/plugin/sdk/detection.go
package sdk

// FrameworkDetector interface for plugins that detect frameworks
type FrameworkDetector interface {
    // GetDetectionPatterns returns patterns this plugin uses for detection
    GetDetectionPatterns() DetectionPatterns

    // Detect performs detection and returns confidence score (0-100)
    Detect(projectPath string) (*DetectionResult, error)

    // GetDefaultCommands returns commands to inject when detected
    GetDefaultCommands() map[string]CommandDefinition

    // EnhanceContext adds framework-specific context
    EnhanceContext(ctx map[string]interface{}) error
}

// DetectionPatterns defines what to look for
type DetectionPatterns struct {
    // Files that must exist
    RequiredFiles []string `json:"required_files"`

    // Files that might exist
    OptionalFiles []string `json:"optional_files"`

    // Directory patterns
    Directories []string `json:"directories"`

    // File content patterns
    FileContents []ContentPattern `json:"file_contents"`

    // File extension patterns
    Extensions []string `json:"extensions"`
}

type ContentPattern struct {
    Filepath string   `json:"filepath"`
    Contains []string `json:"contains"`  // Any of these
    Regex    string   `json:"regex"`     // Or match this regex
}

type DetectionResult struct {
    Detected   bool              `json:"detected"`
    Confidence int               `json:"confidence"` // 0-100
    Framework  FrameworkInfo     `json:"framework"`
    Commands   map[string]string `json:"commands"`
    Metadata   map[string]string `json:"metadata"`
}

type FrameworkInfo struct {
    Name    string `json:"name"`
    Version string `json:"version"`
    Type    string `json:"type"` // language|framework|tool
}
```

### 2. RPC Protocol Extension

```protobuf
// proto/plugin.proto additions

service Plugin {
    // Existing methods...

    // Framework detection methods
    rpc GetDetectionPatterns(Empty) returns (DetectionPatternsResponse);
    rpc DetectFramework(DetectRequest) returns (DetectionResult);
    rpc GetFrameworkCommands(FrameworkRequest) returns (CommandsResponse);
}

message DetectionPatternsResponse {
    repeated string required_files = 1;
    repeated string optional_files = 2;
    repeated string directories = 3;
    repeated ContentPattern file_contents = 4;
}

message DetectRequest {
    string project_path = 1;
    map<string, string> context = 2;
}

message DetectionResult {
    bool detected = 1;
    int32 confidence = 2;
    FrameworkInfo framework = 3;
    map<string, string> commands = 4;
    map<string, string> metadata = 5;
}
```

### 3. Context Integration

```go
// internal/context/framework_detector.go
package context

import (
    "github.com/ivannovak/glide/pkg/plugin"
)

// FrameworkDetector aggregates all plugin detections
type FrameworkDetector struct {
    plugins []plugin.Plugin
    cache   map[string]*DetectionCache
}

// DetectFrameworks runs all plugin detections in parallel
func (fd *FrameworkDetector) DetectFrameworks(projectPath string) ([]FrameworkResult, error) {
    results := make(chan FrameworkResult, len(fd.plugins))

    // Parallel detection
    for _, p := range fd.plugins {
        go func(plugin plugin.Plugin) {
            result, _ := plugin.DetectFramework(projectPath)
            results <- result
        }(p)
    }

    // Collect results
    var frameworks []FrameworkResult
    for i := 0; i < len(fd.plugins); i++ {
        if result := <-results; result.Detected {
            frameworks = append(frameworks, result)
        }
    }

    return fd.resolveConflicts(frameworks), nil
}

// resolveConflicts handles multiple plugins detecting same framework
func (fd *FrameworkDetector) resolveConflicts(results []FrameworkResult) []FrameworkResult {
    // Sort by confidence score
    sort.Slice(results, func(i, j int) bool {
        return results[i].Confidence > results[j].Confidence
    })

    // Remove duplicates, keeping highest confidence
    seen := make(map[string]bool)
    filtered := []FrameworkResult{}

    for _, r := range results {
        if !seen[r.Framework.Name] {
            seen[r.Framework.Name] = true
            filtered = append(filtered, r)
        }
    }

    return filtered
}
```

### 4. Command Injection

```go
// internal/cli/framework_commands.go
package cli

// InjectFrameworkCommands adds detected framework commands
func InjectFrameworkCommands(
    registry *Registry,
    frameworks []FrameworkResult,
    config *config.Config,
) {
    // Collect all framework commands
    commands := make(map[string]CommandDefinition)

    for _, fw := range frameworks {
        for name, def := range fw.Commands {
            // Check for conflicts
            if existing, exists := commands[name]; exists {
                // Resolve based on priority or confidence
                if shouldOverride(def, existing, fw.Confidence) {
                    commands[name] = def
                }
            } else {
                commands[name] = def
            }
        }
    }

    // Register commands with appropriate metadata
    for name, def := range commands {
        registry.Register(name, func() *cobra.Command {
            return &cobra.Command{
                Use:   name,
                Short: def.Description,
                Run: func(cmd *cobra.Command, args []string) {
                    ExecuteFrameworkCommand(def.Cmd, args)
                },
            }
        }, Metadata{
            Name:        name,
            Category:    CategoryFramework,
            Description: def.Description,
            Source:      "framework-detection",
        })
    }
}
```

## Implementation Plan

### Phase 1: Core Infrastructure
1. Define plugin interfaces and protocols
2. Extend RPC communication
3. Add detection orchestration
4. Implement command injection

### Phase 2: Migration
1. Extract existing detection logic
2. Create built-in framework plugins:
   - `go-plugin` (Go detection)
   - `node-plugin` (Node.js/npm/yarn)
   - `docker-plugin` (Docker/Compose)
3. Maintain backward compatibility

### Phase 3: Enhancement
1. Add caching layer for detection results
2. Implement parallel detection
3. Add performance monitoring
4. Create plugin development guide

## Plugin Development Example

```go
// example-plugin/main.go
package main

import (
    "github.com/ivannovak/glide/pkg/plugin/sdk"
)

type RustPlugin struct {
    sdk.BasePlugin
}

func (p *RustPlugin) GetDetectionPatterns() sdk.DetectionPatterns {
    return sdk.DetectionPatterns{
        RequiredFiles: []string{"Cargo.toml"},
        OptionalFiles: []string{"Cargo.lock", "rust-toolchain.toml"},
        Directories:   []string{"src", "target"},
    }
}

func (p *RustPlugin) Detect(projectPath string) (*sdk.DetectionResult, error) {
    // Check for Cargo.toml
    cargoPath := filepath.Join(projectPath, "Cargo.toml")
    if _, err := os.Stat(cargoPath); err != nil {
        return &sdk.DetectionResult{Detected: false}, nil
    }

    // Parse version from Cargo.toml
    version := p.parseRustVersion(cargoPath)

    return &sdk.DetectionResult{
        Detected:   true,
        Confidence: 95,
        Framework: sdk.FrameworkInfo{
            Name:    "rust",
            Version: version,
            Type:    "language",
        },
        Commands: p.GetDefaultCommands(),
    }, nil
}

func (p *RustPlugin) GetDefaultCommands() map[string]sdk.CommandDefinition {
    return map[string]sdk.CommandDefinition{
        "build": {
            Cmd:         "cargo build",
            Description: "Build Rust project",
        },
        "test": {
            Cmd:         "cargo test",
            Description: "Run Rust tests",
        },
        "run": {
            Cmd:         "cargo run",
            Description: "Run Rust application",
        },
        "fmt": {
            Cmd:         "cargo fmt",
            Description: "Format Rust code",
        },
        "lint": {
            Cmd:         "cargo clippy",
            Description: "Lint Rust code",
        },
    }
}

func main() {
    sdk.ServePlugin(&RustPlugin{})
}
```

## Performance Considerations

### Caching Strategy
```go
type DetectionCache struct {
    ProjectPath string
    Timestamp   time.Time
    Results     []FrameworkResult
    TTL         time.Duration // Default: 5 minutes
}
```

### Parallel Execution
- Run all plugin detections concurrently
- Set timeout per plugin (default: 100ms)
- Cache results for subsequent calls

### Lazy Loading
- Only load framework plugins when entering a project
- Unload when switching projects
- Keep core plugins always loaded

## Security Considerations

1. **Pattern Validation**: Validate detection patterns to prevent path traversal
2. **Sandboxing**: Run detection in restricted environment
3. **Resource Limits**: Cap CPU/memory usage during detection
4. **Command Validation**: Validate injected commands before execution

## Testing Strategy

### Unit Tests
- Test individual detection patterns
- Test conflict resolution
- Test command injection

### Integration Tests
- Test with multiple plugins
- Test performance with many patterns
- Test caching behavior

### Example Test
```go
func TestFrameworkDetection(t *testing.T) {
    // Create test project structure
    tmpDir := t.TempDir()
    createFile(t, tmpDir, "package.json", `{"name": "test"}`)
    createFile(t, tmpDir, "tsconfig.json", `{}`)

    // Load plugins
    detector := NewFrameworkDetector()
    detector.LoadPlugin(NewNodePlugin())
    detector.LoadPlugin(NewTypeScriptPlugin())

    // Detect frameworks
    results, err := detector.DetectFrameworks(tmpDir)
    require.NoError(t, err)

    // Verify detections
    assert.Len(t, results, 2)
    assert.Contains(t, getFrameworkNames(results), "node")
    assert.Contains(t, getFrameworkNames(results), "typescript")
}
```

## Migration Guide

### For Plugin Authors

1. Implement `FrameworkDetector` interface
2. Register detection patterns
3. Provide default commands
4. Test with various project structures

### For Users

1. Install framework plugins: `glideplugins install [plugin]`
2. Verify detection: `glide context`
3. Override if needed in `.glide.yml`

## Future Enhancements

1. **Smart Detection**: ML-based pattern recognition
2. **Version Negotiation**: Handle multiple framework versions
3. **Dependency Resolution**: Detect transitive frameworks
4. **Hot Reload**: Reload detection when files change
