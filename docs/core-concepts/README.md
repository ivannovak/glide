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

Glide supports three development modes that adapt to your project structure:

### Single-Repo Mode

The default mode for standard Git repository workflows:

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

### Standalone Mode

For directories without Git repositories:

```bash
# Create a .glide.yml in any directory
echo 'commands: { hello: "echo Hello!" }' > .glide.yml
glid help  # Shows "📄 Standalone mode"
glid hello # Your commands work immediately
```

**Characteristics:**
- No Git repository required
- Commands defined in `.glide.yml` only
- Perfect for automation scripts
- Minimal project overhead
- Works in temporary directories

### Switching Between Modes

Use `glid setup` to configure your development mode:
- Converts project structure when switching
- Preserves your work and configuration
- Can switch back at any time
- Standalone mode is automatic when `.glide.yml` exists without Git

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
# From project root
cd vcs/                  # Main repo for hotfixes
cd ../worktrees/feature-a  # Full environment running
cd ../worktrees/feature-b  # Different environment running
# No context loss, complete isolation!
```

### How Worktrees Work

```
project-root/
├── vcs/                  # Main Git repository
│   ├── .git/            # Git directory
│   └── ...              # Project files (for hotfixes/exploration)
└── worktrees/
    ├── feature-a/       # Git worktree checkout
    │   ├── .git         # File pointing to main repo
    │   ├── .env         # Isolated config
    │   └── ...          # All project files
    └── feature-b/       # Another worktree checkout
        ├── .git         # File pointing to main repo
        ├── .env         # Different config
        └── ...          # All project files
```

Key architecture points:
- **vcs/**: Contains the primary Git checkout for hotfixes and exploratory work
- **worktrees/**: Contains all Git worktrees, completely separate from main repo
- **No .gitignore changes needed**: Worktrees live outside the Git context
- **Clean separation**: Main repository remains unaware of worktree existence

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

## YAML-Defined Commands

### Command Definition

Define custom commands in `.glide.yml` without writing plugins:

```yaml
commands:
  # Simple format
  build: make build

  # Full format with metadata
  deploy:
    cmd: |
      echo "Deploying to $1..."
      ./scripts/deploy.sh $1
      echo "Deploy complete!"
    alias: d
    description: Deploy to environment
    help: |
      Deploy the application to specified environment.

      Usage: glid deploy [staging|production]
    category: deployment
```

### Command Features

**Parameter Expansion**: Pass arguments to commands
```bash
glid deploy staging  # $1 = staging
glid test unit integration  # $1 = unit, $2 = integration
```

**Shell Script Support**: Full shell capabilities
- Multi-line scripts
- Control structures (if/then/else, loops)
- Pipes and redirections
- Environment variables
- Shell functions

**Command Metadata**:
- `alias`: Short alternative name
- `description`: One-line description for help
- `help`: Detailed help text
- `category`: Group in help output

## Command Resolution

### Priority Order

When you run a command, Glide checks in order:

1. **Built-in commands** - Core Glide functionality
2. **Local YAML commands** - From `.glide.yml` in current/parent directories
3. **Plugin commands** - From active plugins
4. **Global YAML commands** - From `~/.glide/config.yml`
5. **Aliases** - User-defined shortcuts
6. **Pass-through** - To system commands

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