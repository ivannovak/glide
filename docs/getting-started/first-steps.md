# First Steps with Glide

Welcome! This guide will walk you through your first 5 minutes with Glide.

## Understanding Context

Glide adapts to your project structure automatically:

```bash
cd your-project
glide help
```

The header shows your current mode:
- 📁 **Single-repo mode** - Standard Git repository
- 🌳 **Multi-worktree mode** - Multiple features in parallel
- 📄 **Standalone mode** - Non-Git directory with `.glide.yml`

## Essential Commands

### Getting Help

See all available commands:

```bash
glide help
```

Get help for a specific command:

```bash
glide help worktree
```

### Extending with Custom Commands

You can extend Glide with custom commands in two ways:

#### YAML-Defined Commands (Recommended)
Define commands directly in `.glide.yml`:

```yaml
commands:
  build: make build
  test: go test ./...
  deploy:
    cmd: ./scripts/deploy.sh $1
    description: Deploy to environment
```

These become available immediately:
```bash
glidebuild
glide test
glidedeploy staging
```

#### Runtime Plugins (Advanced)
For complex integrations, you can install runtime plugins:

```bash
# List installed plugins
glideplugins list

# Install a plugin binary
glideplugins install /path/to/plugin

# Get plugin information
glideplugins info [plugin-name]
```

**Getting Plugins**:
- Build your own using the [Plugin Development Guide](../plugin-development.md)
- Check community repositories and examples
- Currently no centralized marketplace (coming in future releases)
- Most users should start with YAML commands and only use plugins for complex needs

## Development Modes

Glide supports three development modes:

### Single-Repo Mode
The default mode for standard development:
- Work on one branch at a time
- Simple, straightforward workflow
- All commands operate on your current branch

### Multi-Worktree Mode
Advanced mode for parallel development:
- Work on multiple features simultaneously
- Each worktree has its own isolated environment
- Additional `project` commands for managing all worktrees

```bash
# Check your current mode
glide help  # Shows mode in the header

# Switch to multi-worktree mode
glide setup

# Once in multi-worktree mode
glide project worktree feature/new    # Create a worktree
glide project list                     # List all worktrees
glide project status                   # Status across all worktrees
```

### Standalone Mode
For directories without Git repositories:
- Use Glide anywhere with just a `.glide.yml` file
- Perfect for personal scripts and automation
- No Git required

```bash
# Create a .glide.yml in any directory
echo 'commands: { hello: "echo Hello!" }' > .glide.yml

# Commands work immediately
glidehello
```


## Multi-Worktree Development

For advanced workflows, Glide supports multi-worktree development:

```bash
# First, set up multi-worktree mode
glide setup

# Once enabled, use project commands:
glide project worktree feature/awesome-feature
# Or use the short alias:
glidep worktree feature/awesome-feature

# List all worktrees
glidep list

# Check status across all worktrees
glidep status
```

This mode allows you to work on multiple features simultaneously with isolated environments.

## Configuration

Glide looks for configuration in this order:

1. Project-specific: `.glide.yml` in your project
2. Global: `~/.glide/config.yml`

### YAML-Defined Commands

You can define custom commands directly in your configuration files. These become first-class commands available through `glid`:

```yaml
# .glide.yml
commands:
  # Simple format - just the command to run
  build: docker build --no-cache .
  test: go test ./...

  # Structured format with additional metadata
  deploy:
    cmd: ./scripts/deploy.sh $1
    alias: d
    description: Deploy to environment
    help: |
      Deploy the application to the specified environment.
      Usage: glidedeploy [staging|production]
    category: deployment

  # Multi-line commands
  setup: |
    npm install
    docker-compose build
    docker-compose run --rm app migrate
```

#### Command Features

- **Parameter Substitution**: Use `$1`, `$2`, etc. for positional arguments, `$@` or `$*` for all arguments
- **Multi-line Execution**: Commands with multiple lines are executed in sequence
- **Command Priority**: Local commands override plugin commands but cannot override core Glide commands
- **Recursive Discovery**: Commands are discovered from configuration files up the directory tree

#### Using YAML Commands

Once defined, YAML commands work just like built-in commands:

```bash
# Run a simple command
glidebuild

# Pass arguments to commands
glidedeploy staging

# Use command aliases
glided production

# Commands appear in help
glide help
```


## Tips for Success

### 1. Let Context Guide You
Don't memorize commands. Use `glide help` in different projects to see what's available.

### 2. Use Tab Completion
If you set up shell completion, double-tab shows available options:
```bash
glide<TAB><TAB>
```

### 3. Explore Gradually
Start with basic commands, then add YAML commands as you need them.

### 4. Build Your Command Library
Start with simple YAML commands and expand as needed:
```yaml
# Start simple, grow over time
commands:
  clean: rm -rf build/
  fresh: git pull && make build
  review: gh pr create --draft
```

## Common Workflows

### Morning Routine
```bash
# Update Glide
glide self-update

# Pull latest code and run your custom setup
# (assuming you've defined these in .glide.yml)
glidefresh  # Runs: git pull && make build
```

### Working with Multiple Features (Multi-Worktree Mode)
```bash
# Create a new worktree for a feature
glide project worktree feature/new-thing

# Switch to it
cd ~/project/worktrees/feature-new-thing

# List all worktrees
glide project list
```

### Creating Your Own Workflows
Define commands for your specific needs in `.glide.yml`:
```yaml
commands:
  # Morning routine
  fresh: |
    git pull
    npm install
    docker-compose up -d
    echo "Ready to work!"

  # Quick test
  check: |
    npm run lint
    npm run test

  # Deploy
  ship:
    cmd: ./deploy.sh $1
    description: Deploy to environment
    help: "Usage: glideship [staging|production]"
```

## Next Steps

- Learn about [Core Concepts](../core-concepts/README.md)
- Explore [Common Workflows](../guides/README.md)
- Review [Troubleshooting](../troubleshooting.md) if you encounter issues

## Getting Help

- Run `glide help` for command reference
- Check the [guides](../guides/) for specific scenarios
- Visit [GitHub Issues](https://github.com/ivannovak/glide/issues) for support
