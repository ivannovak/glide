# Framework Detection Plugin System - Product Specification

## Overview

Extend Glide's plugin system to enable community-driven framework detection and automatic command provisioning. Instead of hard-coding detection for every possible framework, allow plugins to contribute detection patterns and default commands.

## Problem Statement

Currently, Glide cannot:
- Detect all possible frameworks and build systems
- Keep up with new frameworks as they emerge
- Provide specialized commands for niche or proprietary frameworks
- Allow teams to define custom framework patterns

Attempting to build all framework detection into core is unsustainable and will always leave gaps.

## Solution

Enable plugins to:
1. Register framework detection patterns
2. Provide default YAML commands when their framework is detected
3. Contribute to context detection and help output
4. Override or enhance built-in detection

## User Stories

### As a Framework Author
- I want to create a Glide plugin for my framework
- So that users get appropriate commands automatically
- Without waiting for Glide core to add support

### As a Team Lead
- I want to define detection for our proprietary build system
- So that team members get consistent commands
- Without modifying Glide core

### As a Developer
- I want Glide to detect my project's framework
- And provide relevant commands automatically
- Without manual configuration

## Key Features

### 1. Detection Pattern Registration

Plugins can register patterns to detect frameworks:

```yaml
# Plugin declares detection patterns
detection:
  files:
    - package.json      # Node.js
    - tsconfig.json     # TypeScript
  directories:
    - node_modules      # Confirms Node.js
  file_contents:
    - file: package.json
      contains: ["react", "vue", "angular"]  # Frontend frameworks
```

### 2. Automatic Command Injection

When a framework is detected, inject default commands:

```go
// Plugin provides commands for detected framework
func (p *NodePlugin) GetDefaultCommands() map[string]string {
    return map[string]string{
        "test": "npm test",
        "build": "npm run build",
        "dev": "npm run dev",
        "lint": "npm run lint",
    }
}
```

### 3. Context Enhancement

Plugins contribute to context detection:

```go
// Plugin adds to project context
func (p *NodePlugin) EnhanceContext(ctx *ProjectContext) {
    ctx.Frameworks = append(ctx.Frameworks, "node")
    ctx.PackageManager = p.detectPackageManager() // npm, yarn, pnpm
}
```

### 4. Priority and Conflict Resolution

Multiple plugins might detect overlapping patterns:

```yaml
# Core Glide config
framework_detection:
  priority:
    - custom-team-plugin    # Highest priority
    - rails-plugin
    - ruby-plugin          # Lower priority

  conflict_resolution: first  # first|all|merge
```

## User Experience

### Installation

```bash
# Install a framework plugin
glideplugins install golang-plugin

# Plugin automatically activates on Go projects
cd my-go-project
glide help
# Shows Go-specific commands
```

### Framework Detection Display

```bash
glide context
# Output:
Project: my-app
Mode: single-repo
Detected Frameworks:
  - Node.js (v18.0.0) [via node-plugin]
  - React (v18.2.0) [via react-plugin]
  - TypeScript (v5.0.0) [via typescript-plugin]
Available Framework Commands:
  - test, build, dev, lint [from node-plugin]
  - component:create [from react-plugin]
  - typecheck [from typescript-plugin]
```

### Override Detection

Users can override automatic detection:

```yaml
# .glide.yml
framework_detection:
  disable:
    - node-plugin    # Don't use Node detection
  force:
    - deno-plugin    # Use Deno even if not detected
```

## Success Criteria

1. **Extensibility**: New frameworks can be supported without modifying core
2. **Discovery**: Users can find framework plugins easily
3. **Performance**: Detection remains fast (<50ms overhead)
4. **Compatibility**: Works with existing plugin system
5. **Override Control**: Users can disable/override detection

## Non-Goals

- Building a plugin marketplace (separate effort)
- Automatic plugin installation (requires marketplace)
- Complex dependency resolution between plugins
- Runtime plugin downloading

## Migration Path

1. Extract existing framework detection into plugins
2. Ship core framework plugins with Glide
3. Document plugin creation for framework authors
4. Gradually move detection logic to community plugins

## Risks and Mitigations

**Risk**: Plugin detection conflicts
- **Mitigation**: Clear priority system and user overrides

**Risk**: Performance impact from many plugins
- **Mitigation**: Lazy loading, caching, parallel detection

**Risk**: Malicious detection patterns
- **Mitigation**: Sandboxed execution, pattern validation

## Future Enhancements

1. **Framework Templates**: Plugins provide project templates
2. **Version Detection**: Detect and handle framework versions
3. **Migration Assistance**: Help migrate between framework versions
4. **Composite Frameworks**: Handle multiple frameworks in one project
