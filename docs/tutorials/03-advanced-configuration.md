# Tutorial 3: Advanced Configuration

This tutorial covers advanced Glide configuration patterns including multi-project setups, configuration inheritance, environment-specific settings, and custom command patterns.

## What You'll Learn

- Configuration inheritance and merging
- Multi-project workspace setup
- Environment-specific configuration
- Advanced command patterns
- Git worktree integration

## Prerequisites

- Completed [Tutorial 1: Getting Started](./01-getting-started.md)
- Basic understanding of YAML

## Configuration Hierarchy

Glide loads configuration from multiple sources, merged in this order:

```
Built-in Defaults (lowest priority)
    ↓
Global Config (~/.glide/config.yml)
    ↓
Parent Directory Configs (inherited)
    ↓
Project Config (.glide.yml)
    ↓
Environment Variables (GLIDE_*)
    ↓
Command-line Flags (highest priority)
```

## Step 1: Global Configuration

Create a global configuration for settings that apply everywhere:

```bash
mkdir -p ~/.glide
```

Create `~/.glide/config.yml`:

```yaml
# Global Glide configuration

# Default output format
output:
  format: table
  colors: true

# Default plugin settings
plugins:
  docker:
    compose_file: docker-compose.yml

# Global command aliases
commands:
  # Quick shortcuts available everywhere
  gs: git status
  gp: git pull
  gc: git commit -m "$@"
```

## Step 2: Project Configuration

### Basic Project Config

Create `.glide.yml` in your project:

```yaml
# Project-specific configuration

# Project metadata (optional)
project:
  name: my-app
  version: 1.0.0

# Commands for this project
commands:
  # Development
  dev: npm run dev
  build: npm run build
  test: npm test $@

  # Docker (overrides global if different)
  up: docker-compose -f docker-compose.dev.yml up -d
  down: docker-compose -f docker-compose.dev.yml down
```

### Understanding Merging

When configs are merged:
- **Maps** are deeply merged (nested keys combined)
- **Arrays** are replaced (not appended)
- **Scalars** are overwritten

Example:

```yaml
# Global (~/.glide/config.yml)
commands:
  gs: git status
  build: make build

# Project (.glide.yml)
commands:
  build: npm run build  # Overrides global
  test: npm test        # Added to global

# Result:
# gs: git status        (from global)
# build: npm run build  (from project - overrides)
# test: npm test        (from project - added)
```

## Step 3: Environment-Specific Configuration

### Using Environment Variables

```yaml
# .glide.yml

environment:
  # Default environment
  NODE_ENV: development
  DATABASE_URL: postgres://localhost/dev

commands:
  # Uses environment variables
  migrate: npm run migrate
  seed: npm run seed
```

Override with environment variables:

```bash
# Override for production
DATABASE_URL=postgres://prod-server/app glide migrate
```

### Environment Files

Create environment-specific configs:

```yaml
# .glide.yml
commands:
  # Load environment from file
  dev: |
    source .env.development
    npm run dev

  prod: |
    source .env.production
    npm run start
```

### Multi-Environment Pattern

```yaml
# .glide.yml

# Base commands
commands:
  # Environment-aware commands
  deploy:
    cmd: ./scripts/deploy.sh $1
    description: Deploy to environment

  db:migrate:
    cmd: DATABASE_URL=$DATABASE_URL npm run migrate
    description: Run database migrations

# Environment presets
environments:
  development:
    DATABASE_URL: postgres://localhost/dev
    API_URL: http://localhost:3000

  staging:
    DATABASE_URL: postgres://staging-db/app
    API_URL: https://staging.example.com

  production:
    DATABASE_URL: postgres://prod-db/app
    API_URL: https://api.example.com
```

## Step 4: Multi-Project Workspace

### Monorepo Structure

```
workspace/
├── .glide.yml           # Workspace-level config
├── apps/
│   ├── api/
│   │   └── .glide.yml   # API-specific config
│   └── web/
│       └── .glide.yml   # Web-specific config
└── packages/
    └── shared/
        └── .glide.yml   # Shared package config
```

### Workspace Config

```yaml
# workspace/.glide.yml

project:
  name: my-workspace
  type: monorepo

# Workspace-wide commands
commands:
  # Run command in all apps
  all:build: |
    (cd apps/api && npm run build)
    (cd apps/web && npm run build)

  all:test: |
    (cd apps/api && npm test)
    (cd apps/web && npm test)

  # Navigate to apps
  api: cd apps/api && $@
  web: cd apps/web && $@
```

### App-Specific Config

```yaml
# workspace/apps/api/.glide.yml

project:
  name: api
  parent: my-workspace

commands:
  # API-specific commands
  dev: npm run dev
  build: npm run build
  test: npm test
  migrate: npm run db:migrate
```

### Shared Commands

Define commands that work across the workspace:

```yaml
# workspace/.glide.yml

commands:
  # Works from any subdirectory
  root: cd $(git rev-parse --show-toplevel) && $@

  # Run in specific app from anywhere
  api:dev: cd $(git rev-parse --show-toplevel)/apps/api && npm run dev
  web:dev: cd $(git rev-parse --show-toplevel)/apps/web && npm run dev
```

## Step 5: Git Worktree Integration

### Enable Multi-Worktree Mode

```yaml
# .glide.yml

project:
  worktrees:
    enabled: true
    directory: worktrees  # Where worktrees are created
```

### Worktree Commands

```bash
# Create a feature worktree
glide project worktree feature/my-feature

# List worktrees
glide project worktree list

# Switch to a worktree
cd worktrees/feature-my-feature
```

### Worktree-Aware Commands

```yaml
# .glide.yml

commands:
  # Create and setup worktree
  feature: |
    BRANCH="feature/$1"
    glide project worktree "$BRANCH"
    cd worktrees/$(echo $1 | tr '/' '-')
    cp ../.env.example .env
    npm install
    glide up
    echo "Ready to work on $BRANCH"

  # Clean up worktree
  cleanup: |
    WORKTREE=$1
    glide down
    cd ../..
    git worktree remove "worktrees/$WORKTREE"
    echo "Worktree $WORKTREE removed"
```

## Step 6: Advanced Command Patterns

### Commands with Metadata

```yaml
commands:
  deploy:
    cmd: ./scripts/deploy.sh $1
    description: Deploy the application
    alias: d
    category: deployment
    help: |
      Deploy the application to the specified environment.

      Usage:
        glide deploy <environment>

      Examples:
        glide deploy staging
        glide deploy production
```

### Conditional Commands

```yaml
commands:
  # Check prerequisites
  check-docker: |
    if ! docker info > /dev/null 2>&1; then
      echo "Error: Docker is not running"
      exit 1
    fi
    echo "Docker is ready"

  # Conditional execution
  safe-deploy: |
    glide test || exit 1
    glide check-docker || exit 1
    glide deploy $1
```

### Interactive Commands

```yaml
commands:
  # Interactive menu
  menu: |
    echo "What would you like to do?"
    echo "1) Start development"
    echo "2) Run tests"
    echo "3) Deploy"
    read -p "Choice: " choice
    case $choice in
      1) glide dev ;;
      2) glide test ;;
      3) glide deploy ;;
      *) echo "Invalid choice" ;;
    esac
```

### Command Composition

```yaml
commands:
  # Base commands
  _docker-up: docker-compose up -d
  _wait-healthy: ./scripts/wait-for-healthy.sh
  _run-migrations: npm run migrate
  _seed-data: npm run seed

  # Composed command
  setup: |
    glide _docker-up
    glide _wait-healthy
    glide _run-migrations
    glide _seed-data
    echo "Setup complete!"
```

## Step 7: Plugin Configuration

### Configuring Installed Plugins

```yaml
# .glide.yml

plugins:
  # Docker plugin configuration
  docker:
    compose_file: docker-compose.dev.yml
    default_service: app
    env_file: .env

  # Custom plugin configuration
  my-plugin:
    api_key: ${MY_API_KEY}  # From environment
    timeout: 30
    verbose: true
```

### Plugin-Specific Commands

```yaml
# Some plugins add commands automatically
# You can override or extend them

commands:
  # Override plugin command
  docker:up: docker-compose -f docker-compose.dev.yml up -d --build

  # Extend with wrapper
  up: |
    echo "Starting services..."
    glide docker:up
    echo "Waiting for health checks..."
    sleep 5
    glide docker:ps
```

## Step 8: Debugging Configuration

### View Merged Configuration

```bash
# Show effective configuration
glide config show

# Show specific section
glide config show --section commands

# Show configuration sources
glide config show --sources
```

### Validate Configuration

```bash
# Validate YAML syntax
glide config validate

# Check for issues
glide config lint
```

### Debug Mode

```bash
# Enable debug output
GLIDE_DEBUG=1 glide dev

# Or use flag
glide --debug dev
```

## Best Practices

### 1. Keep Global Config Minimal

```yaml
# ~/.glide/config.yml
# Only truly global settings

output:
  format: table

commands:
  # Only universal shortcuts
  gs: git status
```

### 2. Document Commands

```yaml
commands:
  deploy:
    cmd: ./scripts/deploy.sh $1
    description: Deploy to environment
    help: |
      Detailed help text here.
      Include examples and prerequisites.
```

### 3. Use Consistent Naming

```yaml
commands:
  # Consistent prefixes
  db:migrate: npm run db:migrate
  db:seed: npm run db:seed
  db:reset: npm run db:reset

  # Or namespaces
  test:unit: npm run test:unit
  test:e2e: npm run test:e2e
  test:coverage: npm run test:coverage
```

### 4. Handle Errors

```yaml
commands:
  deploy: |
    set -e  # Exit on error
    glide test || { echo "Tests failed"; exit 1; }
    glide build || { echo "Build failed"; exit 1; }
    ./scripts/deploy.sh $1
```

## What's Next?

You've mastered advanced configuration! Continue with:

1. **[Tutorial 4: Contributing to Glide](./04-contributing.md)** - Become a contributor
2. **[Common Workflows](../guides/README.md)** - Real-world patterns
3. **[Architecture Overview](../architecture/README.md)** - Understand the internals

## Summary

In this tutorial, you learned:
- Configuration inheritance and merging
- Multi-project workspace setup
- Environment-specific configuration
- Git worktree integration
- Advanced command patterns
- Plugin configuration
- Debugging techniques

Configuration is powerful - design it to match your workflow!
