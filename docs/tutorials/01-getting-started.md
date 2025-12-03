# Tutorial 1: Getting Started with Glide

Welcome to Glide! This tutorial will guide you through installing Glide, understanding its core concepts, and using it for your first project.

## What You'll Learn

- How to install Glide
- Basic Glide concepts
- Creating your first configuration
- Running common commands

## Prerequisites

- macOS, Linux, or Windows with WSL
- Go 1.21+ (for building from source)
- Terminal/command line experience

## Step 1: Installation

### Using the Install Script (Recommended)

```bash
curl -sSL https://raw.githubusercontent.com/ivannovak/glide/main/install.sh | bash
```

### Using Go Install

```bash
go install github.com/glide-cli/glide/v3@latest
```

### Verify Installation

```bash
glide version
```

You should see output like:
```
Glide v2.4.0 (abc1234)
  Built: 2025-01-15T10:30:00Z
  Go: go1.24
  OS/Arch: darwin/arm64
```

## Step 2: Understanding Glide

### What is Glide?

Glide is a **context-aware development CLI** that:
- Detects your project type automatically
- Provides consistent commands across projects
- Supports custom command aliases
- Extends via plugins

### Core Concepts

1. **Context Detection**: Glide detects your project root, type, and environment
2. **Configuration**: Define custom commands in `.glide.yml`
3. **Plugins**: Extend Glide with additional functionality

## Step 3: Your First Project

### Create a Project Directory

```bash
mkdir my-project
cd my-project
git init
```

### Check Glide Context

```bash
glide context
```

Output:
```
Project Context:
  Root: /Users/you/my-project
  Mode: single-repo
  Working Dir: /Users/you/my-project
```

### Create a Configuration File

Create `.glide.yml`:

```yaml
# .glide.yml - Glide configuration

# Define custom commands
commands:
  # Simple commands
  hello: echo "Hello from Glide!"

  # Commands with arguments
  greet: echo "Hello, $1!"

  # Multi-line commands
  setup: |
    echo "Setting up project..."
    mkdir -p src tests
    echo "Done!"
```

### Run Your Commands

```bash
# Run a simple command
glide hello
# Output: Hello from Glide!

# Pass arguments
glide greet World
# Output: Hello, World!

# Run multi-line command
glide setup
# Output:
#   Setting up project...
#   Done!
```

## Step 4: Useful Commands

### Get Help

```bash
# Show all available commands
glide help

# Show help for a specific command
glide help context
```

### View Configuration

```bash
# Show current configuration
glide config show
```

### Self-Update

```bash
# Update Glide to the latest version
glide self-update
```

## Step 5: Real-World Example

Let's create a more practical configuration for a Node.js project:

```yaml
# .glide.yml

commands:
  # Development
  dev: npm run dev
  build: npm run build
  test: npm test $@
  lint: npm run lint

  # Database
  db:migrate: npx prisma migrate dev
  db:seed: npx prisma db seed
  db:studio: npx prisma studio

  # Docker
  up: docker-compose up -d
  down: docker-compose down
  logs: docker-compose logs -f $@

  # Shortcuts
  i: npm install
  clean: rm -rf node_modules dist

# Environment variables for commands
environment:
  NODE_ENV: development
```

Now you can use these shortcuts:

```bash
glide dev      # Start development server
glide test     # Run tests
glide up       # Start Docker containers
glide logs api # View logs for 'api' service
```

## Step 6: Shell Completions

Enable tab completion for your shell:

### Bash

```bash
glide completion bash > /etc/bash_completion.d/glide
```

### Zsh

```bash
glide completion zsh > "${fpath[1]}/_glide"
```

### Fish

```bash
glide completion fish > ~/.config/fish/completions/glide.fish
```

## What's Next?

Congratulations! You've learned the basics of Glide. Here's what to explore next:

1. **[Tutorial 2: Creating Your First Plugin](./02-first-plugin.md)** - Build custom functionality
2. **[Tutorial 3: Advanced Configuration](./03-advanced-configuration.md)** - Multi-project setups
3. **[Common Workflows](../guides/README.md)** - Real-world usage patterns

## Troubleshooting

### Command Not Found

If `glide` is not found, add it to your PATH:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Permission Denied

If you get permission errors during installation:

```bash
chmod +x $(which glide)
```

### Configuration Not Loading

Verify your `.glide.yml` is valid YAML:

```bash
glide config validate
```

## Summary

In this tutorial, you learned how to:
- Install Glide
- Understand core concepts
- Create a configuration file
- Define and run custom commands
- Enable shell completions

Glide adapts to your workflow - the more you customize it, the more efficient you become!
