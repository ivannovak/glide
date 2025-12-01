# Core Concepts

Understanding these concepts will help you get the most from Glide.

## Context Awareness

Glide's context system is what makes it feel magical. Instead of remembering different commands for different projects, Glide figures out what you're working on and provides the right tools.

### How Context Detection Works

When you run any Glide command, it:

1. **Scans your environment** - Looks for project markers (.git, .glide.yml, vcs/)
2. **Identifies development mode** - Single-repo, multi-worktree, or standalone
3. **Loads configuration** - From .glide.yml and global config
4. **Provides appropriate commands** - Based on your mode and configuration

### Example: Same Command, Different Contexts

Define context-specific commands in your `.glide.yml`:

```yaml
# Go project .glide.yml
commands:
  test: go test ./...

# Node.js project .glide.yml
commands:
  test: npm test

# Rails project .glide.yml
commands:
  test: rails test
```

Then use the same command across all projects:
```bash
glide test  # Runs the appropriate test command
```

## Extensibility

### Two Ways to Extend Glide

Glide supports two extension mechanisms:

1. **YAML Commands** (Recommended for most users)
   - Simple to define in configuration files
   - No compilation required
   - Perfect for project-specific workflows

2. **Runtime Plugins** (Advanced)
   - Compiled binaries that integrate via gRPC
   - Can provide complex functionality
   - Installed manually (no marketplace yet)

### YAML Commands

```
User → Glide → Config Loading → Command Resolution → Shell Execution
```

Define commands in:
- **Project level**: `.glide.yml`
- **Global level**: `~/.glide/config.yml`

```yaml
commands:
  build: docker build .
  test: go test ./...
  deploy: ./scripts/deploy.sh $1
```

### Runtime Plugins

```
User → Glide → Plugin Discovery → Command Routing → Plugin Execution
```

1. **Discovery**: Glide finds plugins in `~/.glide/plugins/`
2. **Registration**: Plugins declare their commands via gRPC
3. **Routing**: Glide routes commands to plugins
4. **Execution**: Plugin handles the command

**Installing Plugins**:
```bash
# Install from a binary file
glideplugins install /path/to/plugin-binary

# List installed plugins
glideplugins list
```

**Note**: Currently, you need to build or obtain plugin binaries yourself. There's no plugin marketplace or automatic discovery mechanism.

## Development Modes

Glide supports three development modes that adapt to your project structure:

### Single-Repo Mode

The default mode for standard Git repository workflows:

```bash
# All commands operate on your current branch
git checkout feature-branch
glide help  # Shows "Single-repo mode"
```

**Characteristics:**
- One active branch at a time
- Simple, straightforward workflow
- Traditional Git workflow
- All commands work on current branch

### Multi-Worktree Mode

Advanced mode for parallel development across multiple features:

```bash
# Enable multi-worktree mode
glide setup

# Now you have access to project commands
glide project worktree feature-a
glide project worktree feature-b
glide project status  # See all worktrees
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
glide help  # Shows "📄 Standalone mode"
glidehello # Your commands work immediately
```

**Characteristics:**
- No Git repository required
- Commands defined in `.glide.yml` only
- Perfect for automation scripts
- Minimal project overhead
- Works in temporary directories

### Switching Between Modes

Use `glide setup` to configure your development mode:
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
cd vcs/                    # Main branch stays clean as reference
cd ../worktrees/feature-a  # Full environment running
cd ../worktrees/feature-b  # Different environment running
# No context loss, complete isolation!
```

### How Worktrees Work

When you run `glide setup` and choose multi-worktree mode, Glide automatically creates this structure:

```
project-root/
├── vcs/                  # Main Git repository
│   ├── .git/            # Git directory (stays on main/master)
│   └── ...              # Project files (reference copy)
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

**Note**: You don't create this structure manually. The `glide setup` command handles the conversion from a standard Git repository to this multi-worktree layout automatically.

Key architecture points:
- **vcs/**: Contains the main Git repository, kept on the default branch (main/master) as a clean reference for creating new worktrees
- **worktrees/**: Contains all Git worktrees, each branched from vcs/
- **Best Practice**: Keep vcs/ on the latest main branch; do all development work in worktrees
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
glideworktree feature/new-thing

# List worktrees
glideworktree list

# Remove a worktree
glideworktree remove feature/old-thing
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

      Usage: glidedeploy [staging|production]
    category: deployment
```

### Command Features

**Parameter Expansion**: Pass arguments to commands
```bash
glidedeploy staging  # $1 = staging
glide test unit integration  # $1 = unit, $2 = integration
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
3. **Plugin commands** - From installed runtime plugins
4. **Global YAML commands** - From `~/.glide/config.yml`
5. **Aliases** - User-defined shortcuts
6. **Pass-through** - To system commands

### Command Organization

Commands can be organized by category:

**Common operations**:
```yaml
commands:
  test:
    cmd: npm test
    category: testing
  build:
    cmd: npm run build
    category: build
  deploy:
    cmd: ./deploy.sh $1
    category: deployment
```

**Grouped in help output**:
```bash
glide help
# Commands appear grouped by category
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
commands:
  morning: git pull && make
  review: gh pr create
```

**Project** - Specific to one project:
```yaml
# .glide.yml
commands:
  test: npm test
  deploy: ./scripts/deploy.sh $1
environment:
  NODE_ENV: development
```

## Performance Considerations

### Lazy Loading

Glide only loads what's needed:
- Configuration loads on first access
- Commands are discovered once and cached
- Context detection runs only when needed

### Speed Optimizations

- **Binary distribution**: No runtime dependencies
- **Fast startup**: Minimal initialization overhead
- **Smart caching**: Reuse context detection
- **Minimal overhead**: < 50ms startup time

## Security Model

### Command Execution

YAML commands:
- Run via shell in your user context
- Have access to your environment variables
- Execute with your filesystem permissions
- Should be reviewed before running

### Trust Levels

1. **Your commands**: Defined in your .glide.yml
2. **Team commands**: Shared via version control
3. **Global commands**: In ~/.glide/config.yml
4. **Third-party configs**: Review carefully before using

## Next Steps

Now that you understand the core concepts:

- Explore [Common Workflows](../guides/README.md)
- Review [Command Reference](../command-reference.md)
- Check [Troubleshooting](../troubleshooting.md) if you encounter issues
