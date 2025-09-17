# Common Workflows

Real-world patterns and recipes for using Glide effectively.

> **Note**: Many examples in this guide reference commands provided by plugins (like `up`, `down`, `status`, `logs`, `test`, etc.). The actual commands available to you depend on which plugins you have installed. Core Glide provides the framework; plugins provide the specific development commands.

## Core vs Plugin Commands

### Core Glide Commands (Always Available)
- `glid help` - Context-aware help
- `glid setup` - Configure Glide for your project
- `glid plugins` - Manage plugins
- `glid self-update` - Update Glide
- `glid version` - Show version info
- `glid completion` - Shell completions
- `glid project` - Multi-worktree commands (when enabled)

### Plugin Commands (Examples)
The following are common commands that plugins might provide:
- Docker plugin: `up`, `down`, `status`, `logs`, `shell`
- Testing plugin: `test`, `lint`, `coverage`
- Database plugin: `db`, `migrate`, `seed`
- Deployment plugin: `deploy`, `build`

## Daily Development

### Morning Setup

Start your day efficiently:

```bash
# Update Glide itself
glid self-update

# Check your project context
glid help

# Pull latest changes
git pull origin main

# If you have a Docker plugin installed:
# glid up       # Start services
# glid status   # Check everything is running
# glid logs     # See recent logs
```

### Feature Development

Working on a new feature with isolation (requires multi-worktree mode):

```bash
# First, ensure you're in multi-worktree mode
glid setup

# Create a feature worktree
glid project worktree feature/user-authentication

# Navigate to the worktree
cd worktrees/feature-user-authentication

# Copy environment config
cp ../../.env.example .env

# Plugin commands would then be available in the worktree
# For example, with a Docker plugin:
# glid up         # Start isolated services
# glid test       # Run tests
# glid down       # Stop services
```

### Debugging Sessions

When things go wrong:

```bash
# Check service status
glid status

# View recent logs
glid logs --tail 100

# Filter logs by service
glid logs web --follow

# Jump into a container
glid shell web

# Inside container: check processes
ps aux
netstat -tlnp

# Run diagnostic commands
glid healthcheck
```

## Team Collaboration

### Sharing Configurations

Create a `.glide.yml` for your team:

```yaml
# .glide.yml - commit this to your repo

# Define custom commands for your project
commands:
  # Simple commands for common tasks
  lint: golangci-lint run ./...
  fmt: go fmt ./...
  build: docker build --no-cache .

  # Commands with parameters
  test: go test $@ ./...
  deploy: ./scripts/deploy.sh $1

  # Multi-line commands for complex workflows
  setup: |
    go mod download
    docker-compose build
    docker-compose run --rm app migrate
    echo "Setup complete!"

  # Structured commands with metadata
  release:
    cmd: |
      go test ./...
      git tag -a v$1 -m "Release version $1"
      git push origin v$1
    alias: r
    description: Create a new release
    help: |
      Create a new release with the specified version.
      This will run tests, create a git tag, and push it.
      Usage: glid release 1.2.3
    category: deployment

# Plugin configurations
plugins:
  docker:
    compose_file: docker-compose.dev.yml
```

With YAML commands defined, your team gets consistent workflows:

```bash
# Everyone uses the same commands
glid lint       # Run linter with team settings
glid fmt        # Format code consistently
glid setup      # New developers get started quickly
glid test -v    # Parameters are passed through
glid release 1.2.3  # Complex workflows simplified
```

### Onboarding New Developers

Create an onboarding script:

```bash
#!/bin/bash
# setup.sh - First time setup

# Install Glide
curl -sSL https://raw.githubusercontent.com/ivannovak/glide/main/install.sh | bash

# Clone the repository
git clone https://github.com/team/project.git
cd project

# Install required plugins
glid plugins install docker
glid plugins install node-tools

# Setup environment
cp .env.example .env
echo "Please edit .env with your values"

# Start services
glid up

echo "Setup complete! Run 'glid help' to see available commands"
```

## Testing Workflows

### Continuous Testing

Run tests automatically on changes:

```bash
# Watch mode for unit tests
glid test --watch

# Run specific test suites
glid test unit
glid test integration
glid test e2e

# Test with coverage
glid test --coverage
```

### Pre-commit Testing

Add to `.git/hooks/pre-commit`:

```bash
#!/bin/bash
# Run tests before committing

echo "Running tests..."
glid test --quick

if [ $? -ne 0 ]; then
  echo "Tests failed! Commit aborted."
  exit 1
fi

echo "Running linter..."
glid lint

if [ $? -ne 0 ]; then
  echo "Linting failed! Please fix issues."
  exit 1
fi
```

## Multi-Feature Development

### Working on Multiple Features

Leverage worktrees for parallel development:

```bash
# Terminal 1: Feature A
cd worktrees/feature-a
glid up
glid logs --follow

# Terminal 2: Feature B
cd worktrees/feature-b
glid up
glid test --watch

# Terminal 3: Main branch for comparison
cd ../..
glid up
```

### Switching Context

Quick context switches without losing state:

```bash
# Save current work
git add .
git commit -m "WIP: current progress"

# Switch to urgent bugfix
cd ~/project/worktrees/bugfix-critical
glid up
# Fix the bug...
git commit -m "Fix: critical issue"
git push

# Return to feature
cd ~/project/worktrees/feature-a
# Everything still running!
```

## Database Operations

### Database Management

Common database tasks:

```bash
# Access database console
glid db console

# Run migrations
glid db migrate

# Rollback migrations
glid db rollback

# Seed development data
glid db seed

# Backup database
glid db backup > backup.sql

# Restore database
glid db restore < backup.sql
```

### Database Debugging

```bash
# Check connection
glid db ping

# View running queries
glid db queries

# Analyze slow queries
glid db explain "SELECT * FROM users WHERE..."
```

## Deployment Workflows

### Staging Deployment

```bash
# Run tests first
glid test

# Build for staging
glid build --target staging

# Deploy to staging
glid deploy staging

# Verify deployment
glid healthcheck --env staging

# View staging logs
glid logs --env staging --follow
```

### Production Deployment

```bash
# Ensure on main branch
git checkout main
git pull

# Run full test suite
glid test all

# Build production image
glid build --target production

# Deploy with confirmation
glid deploy production --confirm

# Monitor deployment
glid monitor production
```

## Troubleshooting Patterns

### Container Issues

```bash
# Rebuild containers
glid down
glid up --build

# Clean rebuild
glid down --volumes
docker system prune -af
glid up --build

# Check resource usage
docker stats
```

### Network Issues

```bash
# Check network connectivity
glid shell web
> ping database
> nc -zv database 5432

# Inspect network
docker network ls
docker network inspect project_default
```

### Permission Issues

```bash
# Fix file permissions
glid shell web
> chown -R www-data:www-data /app
> chmod -R 755 /app/storage
```

## Performance Optimization

### Speeding Up Development

```bash
# Use cached builds
glid build --cache

# Parallel service startup
glid up --parallel

# Skip unnecessary services
glid up --only web,database

# Use minimal rebuild
glid build --only-changed
```

### Resource Management

```bash
# Limit resource usage
glid up --memory 2g --cpus 2

# Clean up unused resources
glid cleanup

# Prune system
docker system prune -af
```

## Tips and Tricks

### Aliases for Efficiency

Add to your shell profile:

```bash
# ~/.bashrc or ~/.zshrc
alias g='glid'
alias gu='glid up'
alias gd='glid down'
alias gs='glid status'
alias gl='glid logs'
alias gt='glid test'
```

### Project Templates

Create a template for new projects:

```bash
# Save current setup as template
glid template save my-stack

# Create new project from template
glid template use my-stack new-project
```


## Next Steps

- Learn about [Plugin Development](../plugin-development/README.md)
- Read [Core Concepts](../core-concepts/README.md) for deeper understanding
- Check [GitHub](https://github.com/ivannovak/glide) for updates