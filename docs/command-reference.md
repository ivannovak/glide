# Glide Command Reference

## Overview

Glide is a context-aware development CLI that adapts its behavior based on your project structure and current location. This reference covers the core commands built into Glide. Additional commands can be added through YAML configuration or runtime plugins.

## Command Structure

```
glide [command] [subcommand] [flags] [arguments]
```

## Core Commands

These commands are always available, regardless of context or configuration.

### `glide help`

Display context-aware help based on your current location and development mode.

```bash
glide help                     # Show available commands
glide help [command]           # Get help for a specific command
```

**Context Awareness:**
- Shows different commands based on development mode (single-repo, multi-worktree, standalone)
- Hides irrelevant commands (e.g., project commands in single-repo mode)
- Displays YAML-defined commands from your `.glide.yml`

### `glide version`

Display version information for Glide.

```bash
glide version                  # Show version, build date, and commit
glide version --json           # Output as JSON
```

**Aliases:** `v`

### `glide self-update`

Update Glide to the latest version.

```bash
glide self-update              # Download and install latest version
glide self-update --check      # Check for updates without installing
glide self-update --force      # Force reinstall even if up-to-date
```

**Aliases:** `update`, `upgrade`

### `glide plugins`

Manage runtime plugins that extend Glide's functionality.

```bash
glide plugins list             # List installed plugins
glide plugins install <path>   # Install a plugin from binary
glide plugins info <name>      # Get detailed plugin information
glide plugins uninstall <name> # Remove an installed plugin
```

**Subcommands:**
- `list` - Show all installed plugins with their commands
- `install` - Install a plugin binary (requires path to compiled plugin)
- `info` - Display detailed information about a plugin
- `uninstall` - Remove a plugin

**Note:** There is currently no plugin marketplace. Plugins must be built or obtained as binaries.

## Setup & Configuration Commands

### `glide setup`

Configure Glide for your project, including development mode selection.

```bash
glide setup                    # Interactive setup wizard
glide setup --mode single      # Configure single-repo mode
glide setup --mode multi       # Configure multi-worktree mode
```

**Options:**
- `--mode <single|multi>` - Set development mode
- `--force` - Overwrite existing configuration

**What it does:**
- In single-repo mode: Creates `.glide.yml` configuration
- In multi-worktree mode: Restructures project with `vcs/` and `worktrees/` directories

### `glide completion`

Generate shell completion scripts for your shell.

```bash
glide completion bash          # Generate bash completion
glide completion zsh           # Generate zsh completion
glide completion fish          # Generate fish completion
glide completion powershell    # Generate PowerShell completion
```

**Note:** This command is automatically hidden once completions are installed.

### `glide config`

View current configuration (debug command).

```bash
glide config                   # Display all configuration
glide config --json            # Output as JSON
```

## Multi-Worktree Commands

These commands are only available when in multi-worktree mode.

### `glide project`

Manage multiple worktrees and project-wide operations.

```bash
glide project status           # Status of all worktrees
glide project list             # List all worktrees
glide project worktree <name>  # Create new worktree
```

**Aliases:** `p`

**Subcommands:**
- `status` - Show git status across all worktrees
- `list` - List all worktrees with their branches
- `worktree` - Create a new worktree for a branch

**Example:**
```bash
glide project worktree feature/new-feature
cd worktrees/feature-new-feature
# Work in isolated environment
```

## Debug Commands

These commands are available for debugging and troubleshooting.

### `glide context`

Display detailed information about the detected project context.

```bash
glide context                  # Show context information
glide context --json           # Output as JSON
```

**Shows:**
- Project root location
- Development mode (single-repo, multi-worktree, standalone)
- Current location type
- Working directory
- Docker status (if applicable)

## YAML-Defined Commands

You can extend Glide by defining custom commands in configuration files:

### Project Commands (`.glide.yml`)

Define commands specific to your project:

```yaml
commands:
  # Simple command
  build: docker build --no-cache .

  # Command with parameters
  test: go test $@ ./...

  # Multi-line command
  setup: |
    go mod download
    docker-compose build
    echo "Setup complete!"

  # Structured format with metadata
  deploy:
    cmd: ./scripts/deploy.sh $1
    alias: d
    description: Deploy to environment
    help: |
      Deploy the application.
      Usage: glide deploy [staging|production]
    category: deployment
```

### Global Commands (`~/.glide/config.yml`)

Define commands available in all projects:

```yaml
commands:
  morning: git pull && make
  review: gh pr create --draft
```

### Command Priority

When you run a command, Glide resolves it in this order:

1. **Core commands** - Built-in Glide commands (this document)
2. **Local YAML commands** - From `.glide.yml` in current/parent directories
3. **Plugin commands** - From installed runtime plugins
4. **Global YAML commands** - From `~/.glide/config.yml`

## Development Modes

Glide adapts its behavior based on three development modes:

### Single-Repo Mode (Default)
- Standard Git repository workflow
- Work on one branch at a time
- All commands operate on current branch

### Multi-Worktree Mode
- Enabled via `glide setup`
- Work on multiple features simultaneously
- Each worktree has isolated environment
- Adds `project` commands for worktree management

### Standalone Mode
- Activated by `.glide.yml` in non-Git directory
- No Git repository required
- Only YAML-defined commands available
- Perfect for automation scripts

## Environment Variables

Glide respects the following environment variables:

- `GLIDE_CONFIG` - Alternative config file location
- `GLIDE_HOME` - Override `~/.glide` directory
- `NO_COLOR` - Disable colored output
- `EDITOR` - Editor for `glide config edit`

## Exit Codes

- `0` - Success
- `1` - General error
- `2` - Misuse of command
- `127` - Command not found

## Examples

### Getting Started
```bash
# Check what commands are available
glide help

# Set up a project for multi-worktree development
glide setup --mode multi

# Create a worktree for a new feature
glide project worktree feature/awesome
```

### Defining Custom Commands
```yaml
# .glide.yml
commands:
  # Docker operations
  up: docker-compose up -d
  down: docker-compose down
  logs: docker-compose logs -f

  # Testing
  test: npm test
  test:watch: npm test -- --watch

  # Database
  db: docker-compose exec db psql -U postgres
  migrate: npm run migrate
```

### Using Plugins
```bash
# Install a plugin (requires binary)
glide plugins install ~/Downloads/docker-plugin

# List plugin commands
glide plugins list

# Use plugin commands
glide docker:status  # If docker plugin provides this
```

## See Also

- [Getting Started Guide](getting-started/first-steps.md)
- [Core Concepts](core-concepts/README.md)
- [Common Workflows](guides/README.md)
- [Troubleshooting](troubleshooting.md)