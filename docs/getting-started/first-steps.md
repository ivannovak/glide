# First Steps with Glide

Welcome! This guide will walk you through your first 5 minutes with Glide.

## Understanding Context

Glide's superpower is understanding your project context. Let's see what it detects:

```bash
cd your-project
glid context
```

Example output:
```
Project Context:
  Type: Docker Compose Project
  Language: Go
  Git Branch: main
  Working Directory: /Users/you/project
  Available Plugins: docker, git-tools
```

Glide detected your project type and loaded relevant plugins automatically.

## Essential Commands

### Getting Help

See all available commands:

```bash
glid help
```

Get help for a specific command:

```bash
glid help worktree
```

### Working with Plugins

Glide's power comes from plugins that provide project-specific commands:

```bash
# See available plugins
glid plugins list

# Install a plugin from a local file
glid plugins install /path/to/plugin

# Get info about a plugin
glid plugins info [plugin-name]

# Once plugins are installed, their commands become available
# For example, a Docker plugin might provide:
# glid up, glid down, glid status, glid logs
```

## Development Modes

Glide supports two development modes:

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
glid help  # Shows mode in the header

# Switch to multi-worktree mode
glid setup

# Once in multi-worktree mode
glid project worktree feature/new    # Create a worktree
glid project list                     # List all worktrees
glid project status                   # Status across all worktrees
```

## Working with Plugins

### See What's Available

```bash
# List installed plugins
glid plugins list

# See commands from a specific plugin
glid plugins info docker
```

### Plugin Commands Feel Native

Once a plugin is installed, its commands work like built-in ones:

```bash
# These might come from different plugins
glid db migrate        # Database plugin
glid test --watch      # Testing plugin
glid deploy staging    # Deployment plugin
```

## Multi-Worktree Development

For advanced workflows, Glide supports multi-worktree development:

```bash
# First, set up multi-worktree mode
glid setup

# Once enabled, use project commands:
glid project worktree feature/awesome-feature
# Or use the short alias:
glid p worktree feature/awesome-feature

# List all worktrees
glid p list

# Check status across all worktrees
glid p status
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
      Usage: glid deploy [staging|production]
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
glid build

# Pass arguments to commands
glid deploy staging

# Use command aliases
glid d production

# Commands appear in help
glid help
```

### Plugin Configuration

Example `.glide.yml` with both commands and plugin configuration:
```yaml
# Custom commands
commands:
  lint: golangci-lint run ./...
  fmt: go fmt ./...

# Plugin-specific configuration
plugins:
  docker:
    compose_file: docker-compose.dev.yml
```

## Tips for Success

### 1. Let Context Guide You
Don't memorize commands. Use `glid help` in different projects to see what's available.

### 2. Use Tab Completion
If you set up shell completion, double-tab shows available options:
```bash
glid <TAB><TAB>
```

### 3. Explore Gradually
Start with basic commands, then explore plugins as you need them.

### 4. Check Plugin Ecosystem
Many common workflows already have plugins:
```bash
glid plugins search database
```

## Common Workflows

### Morning Routine
```bash
# Update Glide
glid self-update

# Pull latest code
git pull

# Start your environment
glid up

# Check everything's running
glid status
```

### Switching Features
```bash
# Save current work
git commit -am "WIP"

# Switch to another feature
cd ~/project/worktrees/feature-other
glid up
```

### Debugging
```bash
# Check logs
glid logs --tail 50

# Jump into container
glid shell

# Run tests
glid test
```

## Next Steps

- Learn about [Core Concepts](../core-concepts/README.md)
- Explore [Common Workflows](../guides/README.md)
- Review [Troubleshooting](../troubleshooting.md) if you encounter issues

## Getting Help

- Run `glid help` for command reference
- Check the [guides](../guides/) for specific scenarios
- Visit [GitHub Issues](https://github.com/ivannovak/glide/issues) for support