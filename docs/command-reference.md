# Glide Command Reference

## Overview

Glide is a context-aware development CLI that adapts its behavior based on your project structure and current location. It supports both single-repository and multi-worktree development modes.

## Command Structure

```
glide [global-flags] <command> [subcommand] [flags] [arguments]
```

### Command Aliases

Many commands support short aliases for faster typing:

| Command | Aliases | Example |
|---------|---------|---------|
| `artisan` | `a` | `glide a migrate` |
| `composer` | `c` | `glide c install` |
| `test` | `t` | `glide t --filter UserTest` |
| `project` | `g` | `glide p status` |

### Global Flags

- `--help, -h` - Show help for any command
- `--version, -v` - Display version information
- `--verbose` - Enable verbose output
- `--quiet, -q` - Suppress non-essential output
- `--config <path>` - Use alternate configuration file
- `--no-color` - Disable colored output

## Core Commands

### `glide help`

Display contextual help based on your current location.

```bash
glide help                    # Context-aware help
glide help getting-started    # Getting started guide
glide help workflows          # Common workflow examples
glide help troubleshooting    # Troubleshooting guide
glide help modes              # Development modes explained
```

**Context Detection:**
- Multi-worktree root: Shows project commands
- Main repo (`vcs/`): Shows standard commands
- Worktree: Shows worktree-specific commands
- No project: Shows setup instructions

### `glide setup`

Interactive project setup and configuration.

```bash
glide setup                   # Interactive setup wizard
glide setup --mode single     # Configure single-repo mode
glide setup --mode multi      # Configure multi-worktree mode
glide setup --minimal         # Skip optional features
```

**Options:**
- `--mode <single|multi>` - Development mode (default: auto-detect)
- `--path <dir>` - Project directory (default: current)
- `--minimal` - Minimal setup without Docker/testing
- `--force` - Overwrite existing configuration

### `glide config`

Manage Glide configuration.

```bash
glide config                  # Show current configuration
glide config get <key>        # Get specific config value
glide config set <key> <val>  # Set configuration value
glide config edit             # Open config in editor
glide config validate         # Validate configuration
```

**Subcommands:**
- `get` - Retrieve configuration value
- `set` - Update configuration value
- `edit` - Open configuration in `$EDITOR`
- `validate` - Check configuration validity

### `glide context`

Display current project context information.

```bash
glide context                 # Show all context info
glide context --json          # Output as JSON
```

**Output includes:**
- Project root location
- Development mode (single/multi)
- Current location type
- Active worktree (if applicable)
- Docker status
- Configuration status

## Development Commands

### `glide test` (alias: `t`)

Run project tests with intelligent detection.

```bash
glide test                    # Run all tests
glide t unit                  # Run unit tests only (using alias)
glide test feature            # Run feature tests
glide t <pattern>             # Run tests matching pattern (using alias)
glide test --coverage         # Generate coverage report
```

**Options:**
- `--coverage` - Generate coverage report
- `--filter <pattern>` - Filter tests by pattern
- `--parallel` - Run tests in parallel
- `--bail` - Stop on first failure
- `--verbose` - Show detailed output

**Framework Detection:**
- PHP: PHPUnit or Pest
- JavaScript: Jest or Mocha
- Go: `go test`
- Python: pytest
- Ruby: RSpec

### `glide lint`

Run code quality checks.

```bash
glide lint                    # Run all linters
glide lint --fix              # Auto-fix issues
glide lint <file>             # Lint specific file
```

**Options:**
- `--fix` - Automatically fix issues
- `--format <fmt>` - Output format (text/json/xml)
- `--severity <level>` - Minimum severity to report

**Tool Detection:**
- PHP: PHP-CS-Fixer, PHPStan
- JavaScript: ESLint, Prettier
- Go: golangci-lint
- Python: pylint, black
- General: EditorConfig

### `glide format`

Format code according to project standards.

```bash
glide format                  # Format all files
glide format <file>           # Format specific file
glide format --check          # Check without modifying
```

**Options:**
- `--check` - Check formatting without changes
- `--diff` - Show formatting differences

## Docker Commands

### `glide up`

Start Docker development environment.

```bash
glide up                      # Start all services
glide up <service>            # Start specific service
glide up --detach             # Run in background
glide up --build              # Rebuild containers
```

**Options:**
- `--detach, -d` - Run in background
- `--build` - Force rebuild containers
- `--no-cache` - Build without cache
- `--pull` - Pull latest images

### `glide down`

Stop Docker environment.

```bash
glide down                    # Stop containers
glide down --volumes          # Also remove volumes
glide down --rmi all          # Also remove images
```

**Options:**
- `--volumes, -v` - Remove volumes
- `--rmi <all|local>` - Remove images
- `--remove-orphans` - Remove orphan containers

### `glide restart`

Restart Docker services.

```bash
glide restart                 # Restart all services
glide restart <service>       # Restart specific service
```

### `glide logs`

View Docker container logs.

```bash
glide logs                    # Show all logs
glide logs <service>          # Show service logs
glide logs -f                 # Follow log output
glide logs --tail 100         # Show last 100 lines
```

**Options:**
- `--follow, -f` - Follow log output
- `--tail <n>` - Number of lines to show
- `--since <time>` - Show logs since timestamp
- `--timestamps` - Show timestamps

### `glide exec`

Execute commands in Docker containers.

```bash
glide exec <service> <cmd>    # Run command in service
glide exec php composer inst  # Example: Composer install
glide exec -T mysql mysqldump # Non-TTY execution
```

**Options:**
- `-T` - Disable pseudo-TTY allocation
- `--user <user>` - Run as specific user
- `--workdir <dir>` - Working directory
- `--env <key=val>` - Set environment variable

### `glide ps`

Show Docker container status.

```bash
glide ps                      # List running containers
glide ps -a                   # List all containers
```

**Options:**
- `--all, -a` - Show all containers
- `--quiet, -q` - Only display IDs

## Multi-Worktree Commands

These commands are available when in a multi-worktree project.

### `glide project`

Execute commands from any location in the project.

```bash
glide project status           # Project-wide status
glide project list             # List all worktrees
glide project down             # Stop all Docker containers
glide project clean            # Clean all worktrees
```

**Subcommands:**
- `status` - Show status of all worktrees
- `list` - List active worktrees
- `worktree` - Create new worktree
- `down` - Stop all Docker containers
- `clean` - Clean build artifacts
- `update` - Update all worktrees
- `test` - Run tests across worktrees

### `glide project worktree`

Manage Git worktrees.

```bash
glide project worktree <name>              # Create worktree
glide project worktree feature/api         # Create feature branch
glide project worktree fix-123 origin/fix  # From remote branch
```

**Options:**
- `--from <branch>` - Base branch (default: main)
- `--auto-setup` - Auto-configure worktree
- `--no-checkout` - Don't checkout files

### `glide project status`

Show comprehensive project status.

```bash
glide project status           # All worktree statuses
glide project status --docker  # Include Docker status
```

**Output includes:**
- Worktree locations and branches
- Git status for each worktree
- Docker container status
- Active development sessions

## Database Commands

### `glide db`

Database management commands.

```bash
glide db migrate              # Run migrations
glide db seed                 # Seed database
glide db reset                # Reset database
glide db backup               # Create backup
glide db restore <file>       # Restore from backup
```

**Subcommands:**
- `migrate` - Run pending migrations
- `seed` - Seed with test data
- `reset` - Drop and recreate
- `backup` - Create database backup
- `restore` - Restore from backup
- `shell` - Open database shell

## Plugin Commands

### `glide plugins`

Manage Glide plugins.

```bash
glide plugins list            # List installed plugins
glide plugins info <plugin>   # Show plugin details
glide plugins install <path>  # Install plugin
glide plugins uninstall <name> # Remove plugin
```

**Subcommands:**
- `list` - Show all installed plugins
- `info` - Display plugin information
- `install` - Install new plugin
- `uninstall` - Remove plugin
- `update` - Update plugin

### Plugin-Specific Commands

Installed plugins register their own commands:

```bash
glide <plugin> <command>      # Execute plugin command
```

## Shell Utilities

### `glide shell`

Open interactive shell in container.

```bash
glide shell                   # Default shell
glide shell <service>         # Specific service shell
glide shell php               # PHP container shell
glide shell mysql             # MySQL shell
```

### `glide run`

Run one-off commands.

```bash
glide run <command>           # Run in appropriate context
glide run composer install    # Runs in PHP container
glide run npm install         # Runs in Node container
```

## Project Commands

### `glide build`

Build project artifacts.

```bash
glide build                   # Full build
glide build --production      # Production build
glide build --watch           # Watch mode
```

### `glide clean`

Clean generated files and caches.

```bash
glide clean                   # Clean build artifacts
glide clean --all             # Clean everything
glide clean cache             # Clear caches only
```

### `glide update`

Update project dependencies.

```bash
glide update                  # Update all dependencies
glide update composer         # Update PHP dependencies
glide update npm              # Update Node dependencies
```

## Utility Commands

### `glide version`

Display version information.

```bash
glide version                 # Show version
glide version --full          # Detailed version info
```

### `glide doctor`

Diagnose common issues.

```bash
glide doctor                  # Run diagnostics
glide doctor --fix            # Attempt auto-fixes
```

**Checks:**
- Docker installation and daemon
- Required tools and dependencies
- Configuration validity
- File permissions
- Network connectivity

### `glide upgrade`

Self-update Glide to latest version.

```bash
glide upgrade                 # Update to latest
glide upgrade --check         # Check for updates
glide upgrade --force         # Force reinstall
```

## Environment Detection

Glide automatically detects and adapts to your environment:

### Project Structure Detection
- **Multi-worktree**: Presence of `worktrees/` directory
- **Single-repo**: Standard Git repository
- **Docker**: Presence of `docker-compose.yml`

### Location Context
- **Project Root**: Management directory in multi-worktree
- **Main Repo**: `vcs/` directory
- **Worktree**: `worktrees/<name>/`
- **No Project**: Outside any Glide project

### Framework Detection
- **Laravel**: `artisan` file
- **Symfony**: `bin/console`
- **Node.js**: `package.json`
- **Go**: `go.mod`
- **Python**: `requirements.txt` or `pyproject.toml`

## Branding Customization

Glide supports complete branding customization for organizations that want to create their own branded CLI tools. This allows you to:

- Change the CLI name (e.g., from `glide` to `acme`)
- Customize help text and descriptions
- Modify command outputs and messages
- Create organization-specific distributions

### Build-Time Branding

Create a custom-branded build:

```bash
# Build with custom branding
make build BRAND=acme

# Or use build tags
go build -tags brand_acme -o acme cmd/glid/main.go
```

### Branding Configuration

Define branding in `internal/branding/brands/`:

```go
// internal/branding/brands/acme.go
//go:build brand_acme

package brands

func init() {
    Current = Brand{
        Name:        "acme",
        DisplayName: "ACME CLI",
        Description: "ACME's development toolkit",
        CompanyName: "ACME Corporation",
        Website:     "https://acme.example.com",
        // Custom configuration
    }
}
```

### Available Brands

- **glide** (default): Standard Glide branding
- **Custom**: Create your own brand definition

## Configuration

### Configuration File

Default location: `~/.glide.yml`

```yaml
# Development mode preference
mode: multi  # or 'single'

# Project configurations
projects:
  acme:
    path: /Users/username/Code/acme
    mode: multi
    docker: true
    
# Global settings
editor: vim
color: auto
verbose: false

# Plugin settings
plugins:
  acme:
    region: us-west-2
    profile: default
```

### Environment Variables

- `GLIDE_CONFIG` - Configuration file path
- `GLIDE_PROJECT` - Active project name
- `GLIDE_MODE` - Force development mode
- `GLIDE_NO_COLOR` - Disable colored output
- `GLIDE_VERBOSE` - Enable verbose output
- `GLIDE_BRAND` - Override CLI branding (for custom builds)

## Exit Codes

- `0` - Success
- `1` - General error
- `2` - Misuse of command
- `3` - Configuration error
- `4` - Docker not available
- `5` - Project not found
- `126` - Command found but not executable
- `127` - Command not found
- `130` - Interrupted (Ctrl+C)

## Examples

### Daily Workflow

```bash
# Start your day
glide project status          # Check project status
glide up                     # Start Docker
glide test                   # Run tests

# Create new feature
glide project worktree feature/api
cd worktrees/feature-api
glide up && glide test

# End of day
glide project down            # Stop all containers
```

### Quick Commands

```bash
# Run tests
glide test

# Format and lint
glide format && glide lint

# Database operations
glide db migrate && glide db seed

# View logs
glide logs -f php

# Execute in container
glide exec php composer update
```

## See Also

- `glide help getting-started` - Getting started guide
- `glide help workflows` - Common workflow examples
- `glide help troubleshooting` - Troubleshooting guide
- [Runtime Plugin Architecture](runtime-plugin-architecture.md)
- [Plugin Development Guide](runtime-plugin-sdk.md)