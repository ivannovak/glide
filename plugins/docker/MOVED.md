# Docker Plugin Has Moved

The Docker plugin source code has been moved to its own repository:

## New Repository

**https://github.com/ivannovak/glide-plugin-docker**

## Why the Move?

The Docker plugin is now independently versioned and distributed, demonstrating Glide's plugin architecture:

- **Independent releases**: Plugin can be updated without Glide core updates
- **Separate issue tracking**: Plugin-specific issues tracked separately
- **Better example**: Shows proper plugin development practices
- **Cleaner core**: Glide core focuses on framework, not integrations

## What's Here?

This directory now contains:
- **README.md** - Documentation reference for the Docker plugin
- **MOVED.md** - This file explaining the migration

## Installation

Install the Docker plugin from the external repository:

```bash
# Download from GitHub releases
curl -L https://github.com/ivannovak/glide-plugin-docker/releases/latest/download/glide-plugin-docker-darwin-arm64 \
  -o ~/.glide/plugins/glide-plugin-docker
chmod +x ~/.glide/plugins/glide-plugin-docker

# Or build from source
git clone https://github.com/ivannovak/glide-plugin-docker.git
cd glide-plugin-docker
make install
```

## Links

- **Repository**: https://github.com/ivannovak/glide-plugin-docker
- **Issues**: https://github.com/ivannovak/glide-plugin-docker/issues
- **Releases**: https://github.com/ivannovak/glide-plugin-docker/releases
- **Documentation**: [README.md](README.md) in this directory

## Migration History

The Docker functionality was originally built into Glide core (`internal/docker`), then:

1. **Phase 1-5**: Migrated to plugin architecture while maintaining functionality
2. **Phase 6**: Moved plugin code to external repository
3. **Phase 7**: Updated all documentation to reference external repository

This demonstrates the evolution from built-in functionality to a proper plugin system.
