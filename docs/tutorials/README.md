# Glide Tutorials

Step-by-step tutorials to learn Glide from beginner to advanced.

## Learning Path

Start from the beginning and work your way through:

### 1. [Getting Started](./01-getting-started.md)
**Time: 15 minutes** | **Level: Beginner**

Learn the basics of Glide:
- Installation
- Core concepts
- First configuration
- Basic commands

### 2. [Creating Your First Plugin](./02-first-plugin.md)
**Time: 30 minutes** | **Level: Intermediate**

Build a complete plugin using SDK v2:
- Project structure
- Plugin interface
- Type-safe configuration
- Building and installing

### 3. [Advanced Configuration](./03-advanced-configuration.md)
**Time: 25 minutes** | **Level: Intermediate**

Master complex configuration patterns:
- Configuration inheritance
- Multi-project workspaces
- Environment-specific settings
- Git worktree integration

### 4. [Contributing to Glide](./04-contributing.md)
**Time: 20 minutes** | **Level: Advanced**

Become a Glide contributor:
- Development environment
- Codebase structure
- Making changes
- Pull requests

## Quick Start

If you're in a hurry, here's the minimum path:

```bash
# Install Glide
curl -sSL https://raw.githubusercontent.com/ivannovak/glide/main/install.sh | bash

# Create config in your project
cat > .glide.yml << 'EOF'
commands:
  dev: npm run dev
  test: npm test
  build: npm run build
EOF

# Use Glide
glide dev
glide test
glide build
```

## Prerequisites

| Tutorial | Prerequisites |
|----------|---------------|
| Getting Started | Terminal basics |
| First Plugin | Go 1.21+, basic Go knowledge |
| Advanced Config | Completed Tutorial 1 |
| Contributing | Go 1.21+, Git experience |

## Additional Resources

- **[Developer Guides](../guides/README.md)** - In-depth guides on specific topics
- **[Command Reference](../command-reference.md)** - Complete command documentation
- **[Architecture](../architecture/README.md)** - System design and internals
- **[ADR Index](../adr/README.md)** - Design decisions

## Feedback

Found an issue with a tutorial? [Open an issue](https://github.com/ivannovak/glide/issues) to help us improve.
