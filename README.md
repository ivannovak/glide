# Glide CLI

A modern, context-aware development CLI that glides through complex workflows with intelligent project detection and transparent argument passing.

## Features

- 🚀 **Zero Dependencies** - Single binary, no runtime requirements
- 🎯 **Context Aware** - Automatically detects project structure and development mode  
- 🔄 **Transparent Pass-through** - Full argument support for underlying tools
- 🌳 **Multi-Worktree Support** - Manage multiple feature branches in parallel
- ⚡ **Fast** - Native Go binary with < 50ms startup time

## Installation

### Quick Install (Recommended)

```bash
curl -sSL https://raw.githubusercontent.com/ivannovak/glide/main/install.sh | bash
```

### Manual Download

Download the appropriate binary for your platform from [releases](https://github.com/ivannovak/glide/releases):

- macOS Apple Silicon: `glid-darwin-arm64`
- macOS Intel: `glid-darwin-amd64`
- Linux x64: `glid-linux-amd64`
- Linux ARM: `glid-linux-arm64`

### Build from Source

```bash
git clone https://github.com/ivannovak/glide.git
cd glide
./scripts/build.sh
sudo mv dist/glid-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m) /usr/local/bin/glid
```

### Shell Completion

Enable tab autocompletion for your shell:

#### Zsh (Oh My Zsh)
```bash
mkdir -p ~/.oh-my-zsh/completions
glid completion zsh > ~/.oh-my-zsh/completions/_glid
source ~/.zshrc
```

#### Zsh (Standard)
```bash
# Add to ~/.zshrc
source <(glid completion zsh)
```

#### Bash
```bash
# Add to ~/.bashrc or ~/.bash_profile
source <(glid completion bash)
```

#### Fish
```bash
glid completion fish > ~/.config/fish/completions/glid.fish
```

After installation, tab completion will work for all Glide commands, subcommands, flags, and plugin commands.

## Quick Start

### Initial Setup

```bash
# Run from your project directory
glid setup

# Choose development mode:
# 1. Multi-Worktree (recommended for teams)
# 2. Single-Repository (traditional)
```

### Common Commands

```bash
# Start Docker environment
glid up

# Run tests with full argument support
glid test --parallel --processes=5 --filter=MyTest

# Docker commands with auto-detected compose files
glid docker exec app sh
glid docker logs -f app

# Create a new worktree (multi-worktree mode)
glid g worktree feature/new-api

# Stop all Docker containers
glid g down-all
```

## Development Modes

### Multi-Worktree Mode
- Manage multiple feature branches in parallel
- Global commands via `glid g` or `glid global`
- Isolated Docker environments per worktree
- Ideal for teams and complex projects

### Single-Repository Mode  
- Traditional single checkout workflow
- Simpler command structure
- No worktree management overhead
- Perfect for getting started

## Project Structure

```
glide/
├── cmd/glid/         # Main entry point
├── internal/         # Core implementation
│   ├── cli/         # Command definitions
│   ├── context/     # Project detection
│   ├── docker/      # Docker operations
│   ├── shell/       # Shell execution
│   └── config/      # Configuration
├── pkg/             # Reusable packages
├── scripts/         # Build and install scripts
└── dist/            # Compiled binaries
```

## Configuration

Glide uses a global configuration file at `~/.glide.yml`:

```yaml
projects:
  myproject:
    path: /Users/ivan/Code/myproject
    mode: multi-worktree

defaults:
  test:
    parallel: true
    processes: 3
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and guidelines.

## License

MIT