<p align="center">
  <img src="docs/assets/glide-logotype.png" alt="Glide" width="400">
</p>

<h3 align="center">Streamline your development workflow with context-aware command orchestration</h3>

<p align="center">
  <a href="https://github.com/ivannovak/glide/releases"><img src="https://img.shields.io/github/v/release/ivannovak/glide?style=flat-square" alt="Release"></a>
  <a href="https://github.com/ivannovak/glide/actions"><img src="https://img.shields.io/github/actions/workflow/status/ivannovak/glide/ci.yml?branch=main&style=flat-square" alt="Build Status"></a>
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
glid context

# List all available commands
glid help

# Manage plugins
glid plugins list

# Update Glide itself
glid self-update
```

## Core Concepts

### 🎭 Two Development Modes

Glide supports two development modes to match your workflow:

1. **Single-Repo Mode** - Standard development on one branch at a time
2. **Multi-Worktree Mode** - Work on multiple features simultaneously with isolated environments

```bash
# Check your current mode
glid help  # Shows mode in the header

# Switch between modes
glid setup

# In multi-worktree mode, additional commands become available
glid project status     # Check all worktrees
glid project worktree   # Create new worktrees
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
```

These commands become available immediately:
```bash
glid build              # Run your custom build command
glid deploy staging     # Pass arguments with $1, $2, etc.
glid d production       # Use aliases for frequently used commands
```

### 🔌 Plugin System

Extend Glide with custom commands specific to your team or project:

```bash
# List installed plugins
glid plugins list

# Install a plugin from a local file
glid plugins install /path/to/docker-plugin

# Get info about an installed plugin
glid plugins info docker-plugin

# Remove an installed plugin
glid plugins remove docker-plugin
```

### 🌳 Multi-Worktree Development

Glide supports advanced multi-worktree development for working on multiple features simultaneously:

```bash
# Set up multi-worktree mode
glid setup

# After setup, use project commands to manage worktrees
glid project worktree feature/new-feature  # Create a new worktree
glid p worktree feature/new-feature       # Short alias

# List all worktrees
glid project list

# Check status across all worktrees
glid project status
```

## Documentation

### 🚀 Getting Started
- [**Installation Guide**](docs/getting-started/installation.md) - Get Glide running in 2 minutes
- [**First Steps**](docs/getting-started/first-steps.md) - Essential commands and concepts
- [**Project Setup**](docs/getting-started/project-setup.md) - Configure Glide for your project

### 📚 Learn More
- [**Core Concepts**](docs/core-concepts/README.md) - Understand how Glide works
- [**Common Workflows**](docs/guides/README.md) - Real-world usage patterns
- [**Plugin Development**](docs/plugin-development/README.md) - Create your own plugins

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

MIT License - see [LICENSE](LICENSE) for details.

---

<p align="center">
  <sub>Built with ❤️ by developers who were tired of typing the same commands over and over.</sub>
</p>