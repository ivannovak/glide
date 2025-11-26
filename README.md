<p align="center">
  <img src="docs/assets/glide-logotype.png" alt="Glide" width="400">
</p>

<h3 align="center">Streamline your development workflow with context-aware command orchestration</h3>

<p align="center">
  <a href="https://github.com/ivannovak/glide/releases"><img src="https://img.shields.io/github/v/release/ivannovak/glide?style=flat-square" alt="Release"></a>
  <a href="https://github.com/ivannovak/glide/actions"><img src="https://img.shields.io/github/actions/workflow/status/ivannovak/glide/ci.yml?branch=main&style=flat-square" alt="Build Status"></a>
  <a href="https://codecov.io/gh/ivannovak/glide"><img src="https://img.shields.io/codecov/c/github/ivannovak/glide?style=flat-square" alt="Code Coverage"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue?style=flat-square" alt="License"></a>
</p>

---

## What is Glide?

Glide is a context-aware command orchestrator that adapts to your project environment, streamlining complex development workflows through an extensible plugin system. It detects what you're working on and provides the right tools at the right time.

### Why Glide?

- **🎯 Context-Aware**: Automatically detects your project type and provides relevant commands
- **🔌 Extensible**: Add custom commands through a powerful plugin system
- **🌳 Worktree-Optimized**: First-class support for Git worktrees to work on multiple features simultaneously
- **⚡ Fast**: Written in Go for instant command execution
- **🛠️ Developer-First**: Built by developers, for developers who value efficient workflows

## Quick Start

### Install Glide

```bash
# macOS/Linux
curl -sSL https://raw.githubusercontent.com/ivannovak/glide/main/install.sh | bash

# Or download directly from releases
# https://github.com/ivannovak/glide/releases
```

### Your First Commands

```bash
# See what Glide detected about your project
glide context

# List all available commands
glide help

# Manage plugins
glide plugins list

# Update Glide itself
glide self-update
```

## Core Concepts

### 🎭 Three Development Modes

Glide adapts to your project structure automatically:

1. **Single-Repo Mode** - Standard Git repository development
2. **Multi-Worktree Mode** - Work on multiple features simultaneously with isolated environments
3. **Standalone Mode** - Use Glide in any directory with just a `.glide.yml` file (no Git required)

```bash
# Check your current mode
glide help  # Shows mode in the header

# Switch between modes
glide setup

# In multi-worktree mode, additional commands become available
glide project status     # Check all worktrees
glide project worktree   # Create new worktrees
```

### 📝 YAML-Defined Commands

Define custom commands directly in your `.glide.yml` configuration:

```yaml
commands:
  # Simple format
  build: docker build --no-cache .
  test: go test ./...

  # Structured format with metadata
  deploy:
    cmd: ./scripts/deploy.sh $1
    alias: d
    description: Deploy to environment
    help: Deploy the application to staging or production
    category: deployment

  # Multi-line scripts with shell features
  reset:
    cmd: |
      echo "Resetting project..."
      if [ -f .env ]; then
        echo "Keeping .env file"
      fi
      rm -rf build/ dist/ node_modules/
      echo "Reset complete!"
    description: Clean project build artifacts
```

These commands become available immediately:
```bash
glide build              # Run your custom build command
glide test               # Run your test suite
glide deploy staging     # Pass arguments with $1, $2, etc.
glide d production       # Use aliases for frequently used commands
glide reset              # Run multi-line shell scripts
```

#### 📄 Standalone Mode

Use Glide in any directory without a Git repository. Just create a `.glide.yml` file:

```bash
# Create a .glide.yml in any directory
cat > .glide.yml << 'EOF'
commands:
  hello:
    cmd: echo "Hello from Glide!"
    description: Say hello
EOF

# Commands are immediately available
glide hello              # Works without any Git repository!
glide help               # Shows "📄 Standalone mode" at the top
```

This is perfect for:
- Personal automation scripts
- Temporary project directories
- Build environments without Git
- Quick command organization

### 🔌 Plugin System

Extend Glide with custom commands specific to your team or project:

```bash
# List installed plugins
glide plugins list

# Install a plugin from a local file
glide plugins install /path/to/docker-plugin

# Get info about an installed plugin
glide plugins info docker-plugin

# Remove an installed plugin
glide plugins remove docker-plugin
```

### 🌳 Multi-Worktree Development

Glide supports advanced multi-worktree development for working on multiple features simultaneously:

```bash
# Set up multi-worktree mode
glide setup

# After setup, use project commands to manage worktrees
glide project worktree feature/new-feature  # Create a new worktree
glide p worktree feature/new-feature       # Short alias

# List all worktrees
glide project list

# Check status across all worktrees
glide project status
```

## Documentation

### 🚀 Getting Started
- [**Installation Guide**](docs/getting-started/installation.md) - Get Glide running in 2 minutes
- [**First Steps**](docs/getting-started/first-steps.md) - Essential commands and concepts

### 📚 Learn More
- [**Core Concepts**](docs/core-concepts/README.md) - Understand how Glide works
- [**Common Workflows**](docs/guides/README.md) - Real-world usage patterns
- [**Plugin Development**](docs/plugin-development.md) - Create your own plugins

## Built-in Commands

Glide includes essential commands out of the box:

| Command | Description |
|---------|------------|
| `setup` | Interactive setup and configuration |
| `help` | Context-aware help and guidance |
| `version` | Display version information |
| `plugins` | Manage runtime plugins |
| `completion` | Generate shell completion scripts |
| `self-update` | Update Glide to the latest version |
| `project` | Multi-worktree commands (when enabled) |

*Additional commands are provided by plugins based on your project context.*

## Philosophy

Glide follows these principles:

1. **Context is King** - Understand the environment and provide relevant tools
2. **Progressive Disclosure** - Show simple options first, reveal complexity as needed
3. **Extensible by Default** - Teams know their needs best
4. **Speed Matters** - Every millisecond counts in development workflows
5. **Respect Existing Tools** - Enhance, don't replace

## Contributing

We welcome contributions! See our [Contributing Guide](CONTRIBUTING.md) for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/ivannovak/glide/issues)
- **Discussions**: [GitHub Discussions](https://github.com/ivannovak/glide/discussions)

## License

MIT License

---

<p align="center">
  <sub>Built with ❤️ by developers who were tired of typing the same commands over and over.</sub>
</p>

<p align="center">
  <img src="docs/assets/glide-symbol.png" alt="Glide" width="200">
</p>
