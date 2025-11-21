# Docker Plugin Examples

This directory contains example configurations for using the Docker plugin with Glide.

## Overview

> **Plugin Repository**: https://github.com/ivannovak/glide-plugin-docker

The Docker plugin provides Docker and Docker Compose integration for Glide. These examples demonstrate various ways to configure and use the plugin in your projects.

## Examples

### 1. Basic Docker Compose Project
[`basic-compose/.glide.yml`](basic-compose/.glide.yml)

A simple web application using Docker Compose with custom commands for common operations.

**Features:**
- Auto-detected docker-compose.yml
- Custom aliases for up/down/logs
- Database access commands
- Development workflow commands

### 2. Multi-Service Application
[`multi-service/.glide.yml`](multi-service/.glide.yml)

A complex microservices application with multiple compose files and environments.

**Features:**
- Multiple compose file configuration
- Environment-specific commands
- Service-specific log viewing
- Resource cleanup commands

### 3. Plugin Configuration
[`plugin-config/.glide.yml`](plugin-config/.glide.yml)

Advanced plugin configuration with custom project names and environment files.

**Features:**
- Custom Docker Compose project names
- Environment file configuration
- Multiple compose file orchestration
- Custom command categories

## Quick Start

### Install the Docker Plugin

```bash
# Download the latest release for your platform
# Visit: https://github.com/ivannovak/glide-plugin-docker/releases

# Install globally (example for macOS arm64)
curl -L https://github.com/ivannovak/glide-plugin-docker/releases/latest/download/glide-plugin-docker-darwin-arm64 \
  -o ~/.glide/plugins/glide-plugin-docker
chmod +x ~/.glide/plugins/glide-plugin-docker

# Or install for a specific project
mkdir -p .glide/plugins
curl -L https://github.com/ivannovak/glide-plugin-docker/releases/latest/download/glide-plugin-docker-darwin-arm64 \
  -o .glide/plugins/glide-plugin-docker
chmod +x .glide/plugins/glide-plugin-docker
```

### Verify Installation

```bash
glide plugins list
# Should show: docker (version x.x.x)
```

### Use Example Configurations

```bash
# Copy an example to your project
cp examples/docker-plugin/basic-compose/.glide.yml .

# The Docker plugin commands are now available
glide help  # Shows Docker commands in "Docker Management" category
```

## Common Patterns

### Pattern 1: Simple Workflow Commands

Define shorthand commands for common Docker operations:

```yaml
commands:
  start: glide compose:up -d
  stop: glide compose:down
  restart: glide compose:restart
  logs: glide compose:logs -f
```

### Pattern 2: Service-Specific Commands

Create commands for individual services:

```yaml
commands:
  web:
    cmd: glide compose:exec web sh
    description: Open shell in web container
    category: docker

  db:
    cmd: glide compose:exec db psql -U postgres
    description: Connect to database CLI
    category: database
```

### Pattern 3: Development Environment

Combine Docker plugin commands with project setup:

```yaml
commands:
  setup:
    cmd: |
      glide compose:up -d
      glide compose:exec web npm install
      glide compose:exec web npm run migrate
      echo "Development environment ready!"
    description: Set up development environment
    category: setup

  reset:
    cmd: |
      glide compose:down -v
      docker system prune -f
      glide setup
    description: Reset development environment
    category: setup
```

### Pattern 4: Testing with Docker

Integrate tests that run in Docker containers:

```yaml
commands:
  test:
    cmd: glide compose:exec web npm test
    description: Run tests in container
    category: testing

  test:integration:
    cmd: |
      glide compose:up -d
      glide compose:exec web npm run test:integration
      glide compose:down
    description: Run integration tests
    category: testing
```

## Available Plugin Commands

The Docker plugin provides these commands (run `glide help` to see all):

### Container Management
- `docker:ps` - List containers
- `docker:logs` - View container logs
- `docker:exec` - Execute command in container
- `docker:stop` - Stop containers
- `docker:restart` - Restart containers

### Compose Operations
- `compose:up` - Start services
- `compose:down` - Stop services
- `compose:logs` - View service logs
- `compose:exec` - Execute in service container
- `compose:restart` - Restart services

### Health & Monitoring
- `docker:health` - Check Docker health
- `docker:stats` - View resource usage

### Cleanup
- `docker:clean` - Remove unused resources

See the [Docker Plugin README](../../plugins/docker/README.md) for complete documentation.

## Tips

### Combine with YAML Commands

Use YAML commands to create project-specific workflows that leverage the Docker plugin:

```yaml
commands:
  dev:
    cmd: |
      glide up
      echo "Starting in dev mode..."
      glide compose:logs -f web
    description: Start development environment
```

### Use Aliases for Efficiency

The plugin provides short aliases for common commands:
- `glide up` → `glide compose:up`
- `glide down` → `glide compose:down`
- `glide dps` → `glide docker:ps`

### Context-Aware Usage

The plugin automatically adapts to your project structure:
- Single-repo: Uses compose files in project root
- Multi-worktree: Uses compose files from VCS directory
- Standalone: Uses compose files in current directory

## Troubleshooting

### Plugin Not Found

```bash
# Verify plugin is installed
glide plugins list

# Check plugin is executable
chmod +x ~/.glide/plugins/glide-plugin-docker
```

### Docker Commands Not Showing

Make sure Docker is installed and running:
```bash
docker info  # Should show Docker information
```

### Compose Files Not Detected

Check your configuration:
```yaml
plugins:
  docker:
    compose_files:
      - docker-compose.yml
      - docker-compose.override.yml
```

## See Also

- **[Docker Plugin Repository](https://github.com/ivannovak/glide-plugin-docker)** - Source code and releases
- [Docker Plugin Documentation](../../plugins/docker/README.md) - Local reference copy
- [Plugin Development Guide](../../PLUGIN_DEVELOPMENT.md)
- [Glide Core Concepts](../../docs/core-concepts/README.md)
