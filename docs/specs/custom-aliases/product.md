# YAML-Defined Commands - Product Specification

## Overview

YAML-defined commands allow users to define their own first-class commands in Glide configuration files, enabling teams to standardize common workflows and create project-specific commands without writing Go plugins. These commands have the same status as plugin-provided commands.

## Problem Statement

Currently, teams need to:
- Write full Go plugins for simple command shortcuts
- Remember complex command sequences
- Maintain consistency across team members
- Document common workflows separately

## Solution

Enable users to define custom commands in `.glide.yml` that:
- Are first-class citizens alongside plugin commands
- Support both simple and complex command sequences
- Can be shared with the team via version control
- Can have their own aliases for quick access

## User Stories

### As a Developer
- I want to define my own commands without writing Go
- I want to chain multiple commands together
- I want to share these commands with my team
- I want to create aliases for my commands

### As a Team Lead
- I want to standardize workflows across the team
- I want to onboard developers faster with predefined commands
- I want to reduce errors from mistyped commands

## Features

### 1. Simple Command Format
Define commands using the shorthand syntax:
```yaml
commands:
  build: docker build --no-cache
  clean: docker system prune -af
  fresh: git pull && glideup
```

Usage:
```bash
glidebuild
glideclean
glidefresh
```

### 2. Structured Command Format
Define commands with additional properties:
```yaml
commands:
  build:
    cmd: docker build --no-cache
    alias: b
    description: "Rebuild containers without cache"
    help: |
      Rebuilds all Docker containers from scratch without using cache.
      This ensures a completely fresh build but takes longer.

      Usage: glidebuild [options]

      Options:
        --platform  Target platform (linux/amd64, linux/arm64)
    category: development

  fresh:
    cmd: git pull && glideup
    alias: f
    description: "Update code and restart services"
    category: development
```

Usage:
```bash
glidebuild  # or: glideb
glidefresh  # or: glidef
```

### 3. Multi-line Commands (Pipe Syntax)
Support complex command sequences:
```yaml
commands:
  reset: |
    glide down
    docker system prune -f
    glideup --build

  deploy-all:
    cmd: |
      echo "Deploying to all environments..."
      glidedeploy staging
      glidedeploy production
      echo "Deployment complete!"
    alias: da
```

### 4. Parameterized Commands
Pass arguments to commands:
```yaml
commands:
  deploy: glidedeploy $1 --confirm
  backup:
    cmd: pg_dump $DB_NAME > backup-$1.sql
    alias: bk
```

Usage:
```bash
glidedeploy staging
glidebackup 2024-01-15  # or: glidebk 2024-01-15
```

### 5. Environment Variable Support
Use environment variables in commands:
```yaml
commands:
  connect: psql -h $DB_HOST -U $DB_USER $DB_NAME
  test-env: NODE_ENV=test npm test
```

### 6. Plugin Integration
Plugins can bundle YAML command definitions within their plugin structure. When a plugin is loaded, Glide will check if it contains a `commands.yml` file and load those command definitions.

```yaml
# commands.yml (embedded in plugin bundle)
commands:
  up: docker-compose up -d
  rebuild:
    cmd: down && build && up
    alias: rb
    description: "Rebuild and restart containers"
```

This allows plugin authors to:
- Ship useful command combinations without Go code
- Provide sensible defaults for common workflows
- Let users customize or override these commands in their own config

## Configuration Discovery

Similar to how plugins are discovered recursively up the directory tree, configuration files will be discovered using the same mechanism:

1. **Configurable Filename**: The configuration filename is determined by the branding package (default: `.glide.yml`)
2. **Recursive Discovery**: Starting from the current directory, walk up the directory tree looking for configuration files
3. **Stop Points**: Stop at the project root (detected by `.git`), home directory, or filesystem root
4. **Precedence**: Commands defined in deeper directories override those in parent directories
5. **Merging**: Configuration from all discovered files is merged, with deeper configs taking precedence

### Discovery Order (Highest to Lowest Priority)
1. `./[config-filename]` - Current directory
2. `../[config-filename]` - Parent directory (and so on, up the tree)
3. `[project-root]/[config-filename]` - Project root configuration
4. `~/.glide/config.yml` - Global user configuration

This ensures consistency with the plugin discovery mechanism and allows for:
- Project-wide commands at the repository root
- Team/department commands in parent directories
- Personal overrides in subdirectories
- Brand-customizable configuration filenames

## Scope

### In Scope
- Simple command definitions
- Structured command definitions with properties
- Command aliases (alternative names)
- Multi-line command sequences
- Environment variable expansion
- Basic parameter substitution ($1, $2, etc.)
- Recursive configuration discovery (`.glide.yml` files up the tree)
- Project-local commands (.glide.yml)
- Global commands (~/.glide/config.yml)
- Plugin-bundled commands (embedded commands.yml)

### Out of Scope (Future)
- Conditional logic (if/then/else)
- Loops and iteration
- Complex scripting features
- Command dependencies
- Recursive command definitions

## Success Criteria

1. Users can define commands in configuration
2. Commands execute as defined
3. Multi-line commands work correctly
4. Parameters can be passed to commands
5. Command aliases work as alternative names
6. Commands appear in help output
7. Conflicts are handled gracefully

## Priority Order

When multiple commands exist with the same name:
1. Built-in core commands - highest priority (cannot be overridden)
2. Current directory commands (`./.glide.yml`)
3. Parent directory commands (walking up the tree)
4. Project root commands (`[project-root]/.glide.yml`)
5. Plugin-bundled YAML commands
6. Plugin Go commands (compiled)
7. Global commands (~/.glide/config.yml) - lowest priority

This allows users to:
- Override commands at any level of the directory tree
- Have project-wide defaults at the repository root
- Create directory-specific command variations
- Override plugin-provided commands locally

## Examples

### Plugin-Provided Commands
```yaml
# commands.yml (embedded within docker plugin)
commands:
  # Basic Docker operations with sensible defaults
  up: docker-compose up -d
  down: docker-compose down
  logs: docker-compose logs --tail 100

  # Common workflows
  fresh:
    cmd: |
      docker-compose down
      docker-compose pull
      docker-compose build --no-cache
      docker-compose up -d
    alias: f
    description: "Full refresh of containers"

  shell:
    cmd: docker-compose exec $1 /bin/bash
    description: "Open shell in container"
```

### Team Standardization
```yaml
# .glide.yml
commands:
  # Development workflow
  start: up && logs --follow
  stop: down --remove-orphans
  restart: stop && start

  # Testing workflow (structured format)
  test-all:
    cmd: test unit && test integration
    alias: ta
    description: "Run all test suites"

  test-watch:
    cmd: test --watch
    alias: tw

  coverage:
    cmd: test --coverage && open coverage/index.html
    alias: cov

  # Database workflow
  db-reset:
    cmd: |
      glidedb drop
      glidedb create
      glidedb migrate
      glidedb seed
    alias: dbr
    description: "Reset database to fresh state"

  # Deployment
  ship:
    cmd: |
      git pull origin main
      glidetest all
      glidebuild --production
      glidedeploy production --confirm
    alias: s
    description: "Full deployment pipeline"
```

### Personal Productivity
```yaml
# ~/.glide/config.yml
commands:
  # Quick access using structured format
  status:
    alias: s

  logs:
    cmd: logs --tail 50
    alias: l

  restart:
    alias: r

  # Morning routine
  morning:
    cmd: |
      git pull
      glideup
      glidedb migrate
      glide status
    alias: m
    description: "Start the day"
```

## Acceptance Criteria

- [ ] Commands can be defined in config files
- [ ] Both simple and structured formats work
- [ ] Commands execute their defined cmd
- [ ] Multi-line commands execute in sequence
- [ ] Command aliases work as alternative names
- [ ] Parameters are substituted correctly
- [ ] Environment variables are expanded
- [ ] Commands show in help output
- [ ] Priority order is respected
- [ ] Error handling works correctly
