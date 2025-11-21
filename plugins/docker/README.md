# Docker Plugin for Glide

> **Official Repository**: https://github.com/ivannovak/glide-plugin-docker

The Docker plugin provides comprehensive Docker and Docker Compose integration for Glide, offering container management, service orchestration, and development environment commands.

## Installation

The Docker plugin is distributed separately from Glide core as an independent plugin.

### From GitHub Releases (Recommended)

```bash
# Download the latest release for your platform
# Visit: https://github.com/ivannovak/glide-plugin-docker/releases

# For macOS/Linux (example for version 1.0.0, arm64):
curl -L https://github.com/ivannovak/glide-plugin-docker/releases/download/v1.0.0/glide-plugin-docker-darwin-arm64 \
  -o ~/.glide/plugins/glide-plugin-docker
chmod +x ~/.glide/plugins/glide-plugin-docker

# Verify installation
glide plugins list
```

### From Source

```bash
# Clone the repository
git clone https://github.com/ivannovak/glide-plugin-docker.git
cd glide-plugin-docker

# Build and install
make install
# Or manually:
go build -o ~/.glide/plugins/glide-plugin-docker

# Verify installation
glide plugins list
```

### Project-Specific Installation

```bash
# Install for a specific project only
mkdir -p .glide/plugins
cd .glide/plugins

# Download the binary
curl -L https://github.com/ivannovak/glide-plugin-docker/releases/latest/download/glide-plugin-docker-$(uname -s)-$(uname -m) \
  -o glide-plugin-docker
chmod +x glide-plugin-docker

# The plugin will now be available when working in this project
cd ../..
glide plugins list
```

## Prerequisites

- Docker Engine (version 20.10 or later)
- Docker Compose V2 (included with Docker Desktop or `docker compose` command)

The plugin will automatically detect Docker availability and adjust functionality accordingly.

## Features

### 🐳 Container Management
- Start, stop, and restart containers
- View container status and logs
- Execute commands in running containers
- Clean up stopped containers

### 🏗️ Docker Compose Integration
- Start and stop multi-container applications
- View service logs in real-time
- Execute commands in service containers
- Manage compose project lifecycle

### 🔍 Health Monitoring
- Check Docker daemon health
- Monitor container status
- View system resource usage
- Inspect container configurations

### 🧹 Resource Cleanup
- Remove stopped containers
- Clean up dangling images
- Remove unused volumes and networks
- System-wide cleanup operations

## Commands

All commands are registered globally (not namespaced) and categorized under **Docker Management**.

### Container Operations

#### `docker:ps`
List running containers.

```bash
glide docker:ps              # Show running containers
glide docker:ps --all        # Show all containers (including stopped)
```

**Aliases:** `dps`

#### `docker:logs`
View container logs.

```bash
glide docker:logs <container>        # View logs for a container
glide docker:logs <container> -f     # Follow logs in real-time
glide docker:logs <container> --tail 100  # Show last 100 lines
```

**Aliases:** `dlogs`

#### `docker:exec`
Execute a command in a running container.

```bash
glide docker:exec <container> <command>
glide docker:exec web bash           # Open interactive bash shell
glide docker:exec db psql            # Connect to database CLI
```

**Aliases:** `dexec`

#### `docker:stop`
Stop running containers.

```bash
glide docker:stop <container>        # Stop specific container
glide docker:stop web api worker     # Stop multiple containers
```

**Aliases:** `dstop`

#### `docker:restart`
Restart containers.

```bash
glide docker:restart <container>     # Restart specific container
```

**Aliases:** `drestart`

### Docker Compose Operations

#### `compose:up`
Start services defined in docker-compose.yml.

```bash
glide compose:up                     # Start all services
glide compose:up web db              # Start specific services
glide compose:up -d                  # Start in detached mode
```

**Aliases:** `up`, `cup`

#### `compose:down`
Stop and remove containers, networks, and volumes.

```bash
glide compose:down                   # Stop all services
glide compose:down -v                # Also remove volumes
```

**Aliases:** `down`, `cdown`

#### `compose:logs`
View service logs.

```bash
glide compose:logs                   # Show logs for all services
glide compose:logs web               # Show logs for specific service
glide compose:logs -f                # Follow logs in real-time
```

**Aliases:** `clogs`

#### `compose:exec`
Execute a command in a running service container.

```bash
glide compose:exec <service> <command>
glide compose:exec web bash          # Open shell in web service
glide compose:exec db psql -U postgres  # Run database command
```

**Aliases:** `cexec`

#### `compose:restart`
Restart services.

```bash
glide compose:restart                # Restart all services
glide compose:restart web            # Restart specific service
```

**Aliases:** `crestart`

### Health & Status

#### `docker:health`
Check Docker daemon health and system status.

```bash
glide docker:health                  # Show Docker system health
```

**Aliases:** `dhealth`

#### `docker:stats`
Display live resource usage statistics.

```bash
glide docker:stats                   # Show stats for all containers
glide docker:stats web db            # Show stats for specific containers
```

**Aliases:** `dstats`

### Cleanup Operations

#### `docker:clean`
Remove stopped containers and dangling resources.

```bash
glide docker:clean                   # Interactive cleanup
glide docker:clean --all             # Remove all unused resources
glide docker:clean --containers      # Remove only stopped containers
glide docker:clean --images          # Remove only dangling images
```

**Aliases:** `dclean`

## Configuration

The Docker plugin can be configured through your project's `.glide.yml`:

```yaml
plugins:
  docker:
    # Docker Compose file locations (optional)
    compose_files:
      - docker-compose.yml
      - docker-compose.override.yml

    # Default compose project name (optional)
    # If not set, derived from directory name
    project_name: myapp

    # Environment file (optional)
    env_file: .env.docker
```

### Auto-Detection

The plugin automatically detects:
- Docker Compose files in standard locations
- Project name from directory structure
- Available Docker Compose version (V1 vs V2)
- Environment files (`.env`, `.env.local`, etc.)

## Usage Examples

### Development Workflow

```bash
# Start your development environment
glide up

# Check service status
glide docker:ps

# View logs for a specific service
glide compose:logs -f web

# Execute commands in containers
glide compose:exec web npm test

# Stop everything when done
glide down
```

### Debugging

```bash
# Check Docker health
glide docker:health

# View resource usage
glide docker:stats

# Inspect container logs
glide docker:logs web --tail 50

# Open shell in container
glide docker:exec web sh
```

### Cleanup

```bash
# Remove stopped containers
glide docker:clean --containers

# Full cleanup of unused resources
glide docker:clean --all
```

## Project Context Integration

The Docker plugin is context-aware and adapts to your project structure:

### Single-Repo Mode
- Searches for compose files in project root
- Uses project root directory name for compose project name

### Multi-Worktree Mode
- Searches for compose files in VCS directory
- Each worktree can have isolated environments
- Project name derived from VCS directory

### Standalone Mode
- Searches for compose files in current directory
- Works without Git repository

## Troubleshooting

### Docker Not Found

If you see "Docker is not available":
1. Ensure Docker is installed and running
2. Verify `docker` command is in your PATH
3. Check Docker daemon status: `docker info`

### Compose Files Not Detected

The plugin searches for compose files in this order:
1. Files specified in `.glide.yml` config
2. `docker-compose.yml` in project root
3. `docker-compose.yaml` in project root
4. `compose.yml` and `compose.yaml` (newer format)

### Permission Errors

If you encounter permission errors:
- Ensure your user is in the `docker` group (Linux)
- On macOS/Windows, ensure Docker Desktop is running
- Try running with appropriate permissions

### Plugin Not Loading

1. Verify plugin binary is executable:
   ```bash
   chmod +x ~/.glide/plugins/glide-plugin-docker
   ```

2. Check plugin list:
   ```bash
   glide plugins list
   ```

3. Verify plugin compatibility:
   ```bash
   glide plugins info docker
   ```

## Development

The Docker plugin is developed in its own repository for independent versioning and releases.

### Contributing

```bash
# Clone the repository
git clone https://github.com/ivannovak/glide-plugin-docker.git
cd glide-plugin-docker

# Build
make build

# Run tests
make test

# Install locally for testing
make install
```

See the [Contributing Guide](https://github.com/ivannovak/glide-plugin-docker/blob/main/CONTRIBUTING.md) for more details.

## Version History

See [Releases](https://github.com/ivannovak/glide-plugin-docker/releases) for the full version history.

**Latest:**
- **1.0.0** - Initial release with core Docker and Compose functionality
  - Migrated from Glide core to standalone plugin architecture
  - Full feature parity with previous built-in Docker functionality

## License

MIT License - Same as Glide core

## Support

For issues, questions, or feature requests:
- **Plugin Issues**: [glide-plugin-docker Issues](https://github.com/ivannovak/glide-plugin-docker/issues)
- **General Glide**: [Glide Core Issues](https://github.com/ivannovak/glide/issues)
- **Discussions**: [GitHub Discussions](https://github.com/ivannovak/glide-plugin-docker/discussions)

## See Also

- **Plugin Repository**: https://github.com/ivannovak/glide-plugin-docker
- [Glide Plugin Development Guide](../../PLUGIN_DEVELOPMENT.md)
- [Glide Core Documentation](../../docs/)
- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
