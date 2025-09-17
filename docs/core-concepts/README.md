# Core Concepts

Understanding these concepts will help you get the most from Glide.

## Context Awareness

Glide's context system is what makes it feel magical. Instead of remembering different commands for different projects, Glide figures out what you're working on and provides the right tools.

### How Context Detection Works

When you run any Glide command, it:

1. **Scans your environment** - Looks for familiar files and patterns
2. **Identifies project type** - Docker, Node.js, Go, Rails, etc.
3. **Loads relevant plugins** - Only the tools you need
4. **Provides appropriate commands** - Contextual to your project

### Example: Same Command, Different Contexts

The command `glid test` adapts to your project:

```bash
# In a Go project
glid test  # Runs: go test ./...

# In a Node.js project  
glid test  # Runs: npm test

# In a Rails project
glid test  # Runs: rails test
```

You use the same mental model across all projects.

## Plugin Architecture

### What Are Plugins?

Plugins are standalone programs that extend Glide with new commands. They:
- Run in separate processes (secure)
- Communicate via gRPC (fast)
- Can be written in any language
- Install to `~/.glide/plugins/`

### How Plugins Work

```
User → Glide → Plugin Discovery → Command Routing → Plugin Execution
```

1. **Discovery**: Glide finds plugins at startup
2. **Registration**: Plugins declare their commands
3. **Routing**: Glide routes commands to the right plugin
4. **Execution**: Plugin handles the command

### Plugin Types

**Runtime Plugins**: Always available
```yaml
# Provide commands regardless of context
- git-tools
- deployment-utils
```

**Context Plugins**: Activate based on project
```yaml
# Only load in relevant projects
- docker (when docker-compose.yml exists)
- node (when package.json exists)
```

## Development Modes

Glide supports two development modes that fundamentally change how you work with your project:

### Single-Repo Mode

The default mode for standard development workflows:

```bash
# All commands operate on your current branch
git checkout feature-branch
glid help  # Shows "Single-repo mode"
```

**Characteristics:**
- One active branch at a time
- Simple, straightforward workflow
- Traditional Git workflow
- All plugin commands work on current branch

### Multi-Worktree Mode

Advanced mode for parallel development across multiple features:

```bash
# Enable multi-worktree mode
glid setup

# Now you have access to project commands
glid project worktree feature-a
glid project worktree feature-b
glid project status  # See all worktrees
```

**Characteristics:**
- Multiple features active simultaneously
- Each worktree has isolated environment
- Enables `project` command group
- No context switching between features

### Switching Between Modes

Use `glid setup` to configure your development mode:
- Converts project structure when switching
- Preserves your work and configuration
- Can switch back at any time

## Worktree Management

### The Problem Worktrees Solve

Traditional Git workflow:
```bash
# Stash changes, switch branches, rebuild, lose context
git stash
git checkout other-feature
docker-compose down
docker-compose up
# Where was I? What was I doing?
```

With Glide worktrees:
```bash
# Each feature has its own directory and environment
cd worktrees/feature-a  # Full environment running
cd worktrees/feature-b  # Different environment running
# No context loss!
```

### How Worktrees Work

```
main-repo/
├── .git/                 # Shared Git repository
├── worktrees/
│   ├── feature-a/       # Complete checkout
│   │   ├── .env         # Isolated config
│   │   └── ...          # All project files
│   └── feature-b/       # Another checkout
│       ├── .env         # Different config
│       └── ...          # All project files
```

Each worktree:
- Has its own working directory
- Maintains separate Docker containers
- Keeps isolated configuration
- Shares Git history

### Worktree Commands

```bash
# Create a new worktree
glid worktree feature/new-thing

# List worktrees
glid worktree list

# Remove a worktree
glid worktree remove feature/old-thing
```

## Command Resolution

### Priority Order

When you run a command, Glide checks in order:

1. **Built-in commands** - Core Glide functionality
2. **Plugin commands** - From active plugins
3. **Aliases** - User-defined shortcuts
4. **Pass-through** - To system commands

### Command Namespacing

Plugins can register commands at different levels:

**Project-wide commands** (common operations):
```bash
glid test
glid build
glid deploy
```

**Namespaced commands** (plugin-specific):
```bash
glid docker ps
glid docker logs
glid k8s status
```

## Configuration Hierarchy

### Loading Order

Glide loads configuration in this sequence:

1. **Built-in defaults**
2. **Global config** (`~/.glide/config.yml`)
3. **Project config** (`.glide.yml`)
4. **Environment variables** (`GLIDE_*`)
5. **Command-line flags**

Later sources override earlier ones.

### Configuration Scopes

**Global** - Applies everywhere:
```yaml
# ~/.glide/config.yml
editor: vim
theme: dark
plugins:
  global: true
```

**Project** - Specific to one project:
```yaml
# .glide.yml
plugins:
  docker:
    compose_file: docker-compose.dev.yml
environment:
  NODE_ENV: development
```

## Performance Considerations

### Lazy Loading

Glide only loads what's needed:
- Plugins load on-demand
- Commands are discovered once and cached
- Context detection runs only when needed

### Speed Optimizations

- **Binary distribution**: No runtime dependencies
- **Parallel plugin loading**: Concurrent initialization
- **Smart caching**: Reuse context detection
- **Minimal overhead**: < 50ms startup time

## Security Model

### Plugin Isolation

Each plugin:
- Runs in a separate process
- Has no access to Glide internals
- Communicates via defined RPC interface
- Can be sandboxed by the OS

### Trust Levels

1. **Official plugins**: Maintained by Glide team
2. **Community plugins**: Reviewed and verified
3. **Local plugins**: Your own development
4. **Third-party**: Use with caution

## Next Steps

Now that you understand the core concepts:

- Explore [Common Workflows](../guides/README.md)
- Learn about [Plugin Development](../plugin-development/README.md)
- Configure Glide for [Your Project](../getting-started/project-setup.md)