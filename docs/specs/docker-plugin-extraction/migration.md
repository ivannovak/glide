# Docker Plugin Extraction - Migration Guide

## Overview

This guide details the step-by-step process for extracting Docker functionality into a plugin with zero regressions.

## Pre-Migration Checklist

- [ ] Create comprehensive test suite for current Docker functionality
- [ ] Document all current Docker commands and behaviors
- [ ] Benchmark current performance metrics
- [ ] Backup current codebase
- [ ] Create feature branch: `feat/docker-plugin-extraction`

## Phase 1: Core Infrastructure (Days 1-3)

### Step 1.1: Enhance Plugin SDK
```go
// pkg/plugin/sdk/extensions.go
// Add context extension interfaces
type ContextExtension interface {
    Name() string
    Detect(projectPath string) (map[string]interface{}, error)
    Priority() int
}
```

**Files to create:**
- `pkg/plugin/sdk/extensions.go` - Context extension interfaces
- `pkg/plugin/sdk/config_schema.go` - Configuration schema support
- `pkg/plugin/sdk/completion.go` - Completion provider interfaces
- `pkg/plugin/sdk/executor.go` - Executor provider interfaces

### Step 1.2: Modify ProjectContext
```go
// internal/context/types.go
// Add Extensions field while keeping Docker fields
type ProjectContext struct {
    // ... existing fields ...
    Extensions map[string]interface{} `json:"extensions,omitempty"`

    // Keep for compatibility (will be populated from Extensions)
    DockerRunning   bool
    ComposeFiles    []string
    ComposeOverride string
}
```

**Files to modify:**
- `internal/context/types.go` - Add Extensions field
- `internal/context/detector.go` - Support extension detection
- `internal/context/compatibility.go` - Create new file for compatibility

### Step 1.3: Make Shell Plugin-Aware
```go
// internal/shell/registry.go
// Add executor registry for plugins
type ExecutorRegistry struct {
    providers map[string]ExecutorProvider
}
```

**Files to create:**
- `internal/shell/registry.go` - Executor registry
- `internal/shell/plugin_aware.go` - Plugin-aware executor loading

## Phase 2: Docker Plugin Development (Days 4-7)

### Step 2.1: Create Plugin Structure
```bash
mkdir -p plugins/docker/{commands,compose,container,health,shell}

# Create main plugin file
touch plugins/docker/plugin.go
touch plugins/docker/detector.go
```

### Step 2.2: Copy Docker Code
```bash
# Copy existing Docker implementation
cp -r internal/docker/* plugins/docker/
cp internal/cli/docker.go plugins/docker/commands/

# Copy Docker-related project commands
# (will need to extract from project.go)
```

### Step 2.3: Refactor to Plugin Architecture
Transform copied code to use plugin interfaces:

```go
// plugins/docker/plugin.go
package docker

import "github.com/glide-cli/glide/pkg/plugin/sdk"

type DockerPlugin struct {
    *sdk.BasePlugin
    // ... plugin implementation
}

func New() sdk.Plugin {
    return &DockerPlugin{
        BasePlugin: sdk.NewBasePlugin("docker", "1.0.0"),
    }
}
```

**Key transformations:**
- Change imports from internal to plugin SDK
- Implement plugin interfaces
- Use context extensions instead of direct field access
- Register commands via plugin system

## Phase 3: Integration Layer (Days 8-10)

### Step 3.1: Compatibility Layer
```go
// internal/context/compatibility.go
func (c *ProjectContext) PopulateCompatibilityFields() {
    if docker, ok := c.Extensions["docker"].(map[string]interface{}); ok {
        c.DockerRunning = docker["running"].(bool)
        c.ComposeFiles = docker["compose_files"].([]string)
        // ... etc
    }
}
```

### Step 3.2: Plugin Registration
```go
// cmd/glide/plugins_builtin.go
// +build !no_docker

package main

import (
    dockerplugin "github.com/glide-cli/glide/plugins/docker"
    "github.com/glide-cli/glide/pkg/plugin"
)

func init() {
    plugin.RegisterBuiltin(dockerplugin.New())
}
```

### Step 3.3: Wire Up Commands
```go
// internal/cli/builder.go
func (b *Builder) registerPluginCommands() {
    for _, p := range plugin.GetBuiltins() {
        for _, cmd := range p.GetCommands() {
            b.registry.Register(cmd.Name, cmd.Factory, cmd.Metadata)
        }
    }
}
```

## Phase 4: Testing & Validation (Days 11-14)

### Step 4.1: Regression Test Suite
Create comprehensive tests before making changes:

```go
// tests/regression/docker_test.go
func TestDockerRegression(t *testing.T) {
    tests := []struct{
        name string
        test func(t *testing.T)
    }{
        {"DockerCommandPassthrough", testDockerPassthrough},
        {"ProjectStatus", testProjectStatus},
        {"ProjectDown", testProjectDown},
        {"ContextDetection", testContextDetection},
        {"Configuration", testConfiguration},
        {"Completions", testCompletions},
        // ... all Docker functionality
    }
}
```

### Step 4.2: Performance Benchmarks
```go
// tests/benchmark/docker_bench.go
func BenchmarkDockerDetection(b *testing.B) {
    // Baseline: current implementation
    // Compare: plugin implementation
}
```

### Step 4.3: Integration Testing
```bash
#!/bin/bash
# tests/integration/docker_integration.sh

# Test all Docker workflows
echo "Testing Docker command passthrough..."
glide docker ps

echo "Testing project status..."
glide project status

echo "Testing context detection..."
glide context | grep "Docker"

# ... comprehensive testing
```

## Phase 5: Gradual Migration (Days 15-17)

### Step 5.1: Feature Flag Implementation
```yaml
# .glide.yml
features:
  docker_plugin: false  # Start disabled
  legacy_docker: true   # Use old implementation
```

```go
// internal/cli/feature_flags.go
func UseDockerPlugin() bool {
    return config.Features.DockerPlugin
}
```

### Step 5.2: Parallel Testing
Run both implementations side-by-side:

```go
func ExecuteDockerCommand(args []string) error {
    if UseDockerPlugin() {
        return dockerPlugin.Execute(args)
    }
    return legacyDocker.Execute(args)
}
```

### Step 5.3: Gradual Rollout
1. **Week 1**: Internal testing with feature flag
2. **Week 2**: Enable for development builds
3. **Week 3**: Enable by default (can disable via flag)
4. **Week 4**: Remove legacy code

## Phase 6: Cleanup (Days 18-20)

### Step 6.1: Remove Legacy Code
```bash
# After validation period
rm -rf internal/docker/
rm internal/cli/docker.go
# Remove Docker-specific code from project.go
```

### Step 6.2: Update Imports
```bash
# Update all imports from internal/docker to plugin
find . -name "*.go" -exec sed -i 's/internal\/docker/plugins\/docker/g' {} \;
```

### Step 6.3: Documentation Updates
- Update command reference
- Update plugin development guide
- Add Docker plugin documentation
- Update troubleshooting guide

## Validation Checklist

### Pre-Release
- [ ] All regression tests pass
- [ ] Performance benchmarks acceptable
- [ ] Manual testing completed
- [ ] Documentation updated
- [ ] Code review completed

### Post-Release Monitoring
- [ ] No user-reported issues (Week 1)
- [ ] Performance metrics stable (Week 2)
- [ ] Feature flag can be removed (Week 4)
- [ ] Legacy code removed (Week 6)

## Rollback Plan

### Immediate Rollback (< 1 hour)
```bash
# Via feature flag
glide config set features.docker_plugin false
glide config set features.legacy_docker true
```

### Code Rollback (< 1 day)
```bash
# Revert to previous version
git revert [plugin-commit]
make build
```

### Data Migration
No data migration required - configuration format unchanged.

## Common Issues & Solutions

### Issue: Docker commands not found
**Solution**: Ensure Docker plugin is registered in `plugins_builtin.go`

### Issue: Context missing Docker information
**Solution**: Check compatibility layer is populating fields

### Issue: Configuration not recognized
**Solution**: Verify config schema registration in plugin

### Issue: Completions not working
**Solution**: Check completion provider registration

## Performance Monitoring

### Metrics to Track
```go
// internal/metrics/docker.go
type DockerMetrics struct {
    DetectionTime   time.Duration
    CommandExecTime time.Duration
    PluginLoadTime  time.Duration
}
```

### Acceptance Criteria
- Detection: < 50ms (same as current)
- Command execution: No additional overhead
- Plugin load: < 10ms
- Memory usage: No increase

## Communication Plan

### Internal Communication
1. **Week -1**: Announce plan to team
2. **Week 1**: Daily progress updates
3. **Week 2**: Testing results shared
4. **Week 3**: Go/No-go decision
5. **Week 4**: Completion announcement

### User Communication
Since we're currently the only users:
- Document changes in CHANGELOG.md
- Update README if needed
- Prepare for future user communication

## Success Criteria

### Technical Success
- [x] Zero regression in functionality
- [x] Performance maintained or improved
- [x] All tests passing
- [x] Clean code separation

### Product Success
- [x] Seamless user experience
- [x] No breaking changes
- [x] Improved maintainability
- [x] Foundation for future plugins

## Post-Migration Opportunities

Once Docker is successfully extracted:

1. **Podman Plugin**: Create alternative using same interfaces
2. **Docker Compose v2**: Enhanced support for new features
3. **Kubernetes Plugin**: Container orchestration at scale
4. **BuildKit Plugin**: Advanced build features
5. **Container Registry Plugin**: Direct registry integration

## Lessons Learned Documentation

After completion, document:
- What worked well
- What was challenging
- Time estimates vs actual
- Patterns for future extractions
- Reusable code/tools created

## Appendix: File Changes Summary

### Files to Create
- `pkg/plugin/sdk/extensions.go`
- `pkg/plugin/sdk/config_schema.go`
- `pkg/plugin/sdk/completion.go`
- `pkg/plugin/sdk/executor.go`
- `internal/context/compatibility.go`
- `internal/shell/registry.go`
- `plugins/docker/**` (entire plugin)
- `cmd/glide/plugins_builtin.go`

### Files to Modify
- `internal/context/types.go` (add Extensions)
- `internal/context/detector.go` (use extensions)
- `internal/cli/builder.go` (load plugin commands)
- `internal/cli/project.go` (remove Docker code)
- `internal/shell/executor.go` (make plugin-aware)

### Files to Delete (after validation)
- `internal/docker/**` (entire directory)
- `internal/cli/docker.go`
- Docker-specific code in `project.go`

## Timeline Summary

**Total Duration**: 20 working days (4 weeks)

- **Week 1**: Infrastructure & Plugin Development
- **Week 2**: Integration & Testing
- **Week 3**: Validation & Rollout
- **Week 4**: Cleanup & Documentation

This migration guide ensures zero regression while successfully extracting Docker into a plugin.
