# Common Workflows

Real-world patterns and recipes for using Glide effectively.

> **Note**: Many examples in this guide show custom commands that you can define in your `.glide.yml` file or that might be provided by plugins you've installed. Core Glide provides the framework; you extend it with YAML commands or runtime plugins.

## Core vs Extension Commands

### Core Glide Commands (Always Available)
- `glid help` - Context-aware help
- `glid setup` - Configure Glide for your project
- `glid plugins` - Manage runtime plugins
- `glid self-update` - Update Glide
- `glid version` - Show version info
- `glid completion` - Shell completions
- `glid project` - Multi-worktree commands (when enabled)

### Extension Commands
You can extend Glide with:

1. **YAML Commands** - Define in `.glide.yml`:
```yaml
commands:
  # Docker operations
  up: docker-compose up -d
  down: docker-compose down
  status: docker ps
  logs: docker-compose logs $@
  shell: docker-compose exec $1 /bin/bash

  # Testing commands
  test: npm test $@
  lint: npm run lint
  coverage: npm run test:coverage

  # Database operations
  db: docker-compose exec db psql
  migrate: npm run migrate
  seed: npm run seed

  # Deployment
  deploy: ./scripts/deploy.sh $1
  build: docker build --no-cache .
```

2. **Runtime Plugins** - Install compiled plugin binaries:
```bash
# Install a plugin (you need the binary)
glid plugins install /path/to/plugin

# List installed plugins
glid plugins list
```

Note: There's no plugin marketplace yet. You need to build or obtain plugin binaries yourself.

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

# If you have Docker commands defined in .glide.yml:
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

# Your custom commands work in the worktree
# For example, with Docker commands in .glide.yml:
# glid up         # Start isolated services
# glid test       # Run tests
# glid down       # Stop services
```

### Debugging Sessions

When things go wrong:

```bash
# Assuming these commands are defined in your .glide.yml:

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

# Environment variables for commands
environment:
  COMPOSE_FILE: docker-compose.dev.yml
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

# Install plugins if you have any plugin binaries
# glid plugins install /path/to/plugin

# Setup environment
cp .env.example .env
echo "Please edit .env with your values"

# Start services
glid up

echo "Setup complete! Run 'glid help' to see available commands"
```

## Testing Workflows

### Continuous Testing

Define test commands in your `.glide.yml`:

```yaml
commands:
  test: npm test $@
  test:watch: npm test -- --watch
  test:unit: npm test -- --testPathPattern=unit
  test:integration: npm test -- --testPathPattern=integration
  test:e2e: npm run test:e2e
  test:coverage: npm test -- --coverage
```

Then run tests:
```bash
# Watch mode for unit tests
glid test:watch

# Run specific test suites
glid test:unit
glid test:integration
glid test:e2e

# Test with coverage
glid test:coverage
```

### Pre-commit Testing

Add to `.git/hooks/pre-commit`:

```bash
#!/bin/bash
# Run tests before committing

echo "Running tests..."
glid test

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

Define database commands in `.glide.yml`:

```yaml
commands:
  # Database access
  db:console: docker-compose exec db psql -U postgres

  # Migrations
  db:migrate: npm run migrate
  db:rollback: npm run migrate:rollback
  db:seed: npm run seed

  # Backup/restore
  db:backup: docker-compose exec -T db pg_dump -U postgres
  db:restore: docker-compose exec -T db psql -U postgres

  # Debugging
  db:ping: docker-compose exec db pg_isready
  db:queries: docker-compose exec db psql -U postgres -c "SELECT * FROM pg_stat_activity"
```

Then use them:
```bash
# Access database console
glid db:console

# Run migrations
glid db:migrate

# Backup database
glid db:backup > backup.sql

# Check connection
glid db:ping
```

## Deployment Workflows

### Staging Deployment

Define deployment commands in `.glide.yml`:

```yaml
commands:
  deploy:staging:
    cmd: |
      echo "Running tests..."
      npm test
      echo "Building for staging..."
      docker build --target staging -t myapp:staging .
      echo "Deploying to staging..."
      ./scripts/deploy.sh staging
      echo "Deployment complete!"
    description: Deploy to staging environment

  deploy:production:
    cmd: |
      echo "WARNING: Deploying to production!"
      read -p "Are you sure? (y/N) " -n 1 -r
      echo
      if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
      fi
      npm test
      docker build --target production -t myapp:production .
      ./scripts/deploy.sh production
    description: Deploy to production with confirmation
```

Then deploy:
```bash
# Deploy to staging
glid deploy:staging

# Deploy to production
glid deploy:production
```

## Troubleshooting Patterns

### Container Issues

Define troubleshooting commands:

```yaml
commands:
  rebuild: |
    docker-compose down
    docker-compose build --no-cache
    docker-compose up -d

  clean: |
    docker-compose down -v
    docker system prune -af
    docker-compose build --no-cache
    docker-compose up -d

  stats: docker stats
```

Then use them:
```bash
# Rebuild containers
glid rebuild

# Clean rebuild
glid clean

# Check resource usage
glid stats
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

Optimize with custom commands:

```yaml
commands:
  # Fast builds
  build:cache: docker build .
  build:nocache: docker build --no-cache .

  # Service management
  up:fast: docker-compose up -d web database
  up:all: docker-compose up -d

  # Cleanup
  cleanup: docker system prune -f
  cleanup:all: docker system prune -af --volumes
```

Use them for speed:
```bash
# Fast cached build
glid build:cache

# Start only essential services
glid up:fast

# Clean up resources
glid cleanup
```

## Standalone Mode Workflows

### Personal Automation Scripts

Use Glide in any directory without Git for personal automation:

```bash
# Create a scripts directory
mkdir ~/my-scripts
cd ~/my-scripts

# Create .glide.yml
cat > .glide.yml << 'EOF'
commands:
  backup:
    cmd: |
      DATE=$(date +%Y%m%d)
      tar -czf ~/backups/home-$DATE.tar.gz ~/Documents ~/Projects
      echo "Backup created: home-$DATE.tar.gz"
    description: Backup important directories

  sync-notes:
    cmd: rsync -av ~/Notes/ ~/Dropbox/Notes/
    description: Sync notes to Dropbox

  cleanup:
    cmd: |
      echo "Cleaning temporary files..."
      rm -rf ~/Downloads/*.tmp
      rm -rf ~/.cache/thumbnails/*
      echo "Cleanup complete!"
    description: Clean temporary files
EOF

# Use your commands
glid backup
glid sync-notes
glid cleanup
```

### Build Environment Scripts

Create project-agnostic build commands:

```bash
# In a build environment directory
cat > .glide.yml << 'EOF'
commands:
  docker-prune:
    cmd: docker system prune -af --volumes
    description: Deep clean Docker resources

  check-ports:
    cmd: lsof -i -P -n | grep LISTEN
    description: List all listening ports

  monitor:
    cmd: |
      echo "=== CPU Usage ==="
      top -bn1 | head -5
      echo ""
      echo "=== Memory Usage ==="
      free -h
      echo ""
      echo "=== Disk Usage ==="
      df -h
    description: System resource monitoring
EOF
```

### Temporary Project Directories

Use Glide in ephemeral directories:

```bash
# Create a temporary experiment
mkdir /tmp/experiment
cd /tmp/experiment

# Add Glide commands for the experiment
cat > .glide.yml << 'EOF'
commands:
  init:
    cmd: |
      npm init -y
      npm install express
      echo "Experiment initialized!"
    description: Initialize experiment

  run: node server.js
  clean: rm -rf node_modules package*.json
EOF

# Work with your experiment
glid init
glid run
glid clean
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

Share your `.glide.yml` as a template:

```bash
# Copy your configuration to a new project
cp .glide.yml ~/templates/my-stack.yml

# Use in a new project
cp ~/templates/my-stack.yml new-project/.glide.yml
cd new-project
glid help  # Your commands are ready!
```


## Next Steps

- Read [Core Concepts](../core-concepts/README.md) for deeper understanding
- Review the [Command Reference](../command-reference.md) for all commands
- Check [GitHub](https://github.com/ivannovak/glide) for updates