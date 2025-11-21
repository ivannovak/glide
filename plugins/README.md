# Glide Plugins Directory

This directory contains documentation references for official plugins, but not their source code.

## Plugin Architecture

Plugins are designed to be hosted in **separate repositories**, not within the Glide core repository. This ensures:

1. **Clean separation** between core Glide and organization-specific functionality
2. **Independent versioning** of plugins
3. **Private plugins** can remain in private repositories
4. **No bloat** in the core Glide repository

## Official Plugins

### Docker Plugin

**Repository**: https://github.com/ivannovak/glide-plugin-docker

The official Docker plugin provides comprehensive Docker and Docker Compose integration.

- **Documentation**: See [`docker/README.md`](docker/README.md) in this directory
- **Source Code**: https://github.com/ivannovak/glide-plugin-docker
- **Installation**: Download from releases or build from source

The `docker/` directory here contains documentation only - the actual plugin code lives in its own repository.

## Creating Your Own Plugin

Each organization hosts their plugins in their own repositories:

- **Your Plugin**: `github.com/yourorg/glide-plugin-yourorg`

## How to Create a Plugin

See the [Plugin Development Guide](../PLUGIN_DEVELOPMENT.md) for detailed instructions.

Use the Docker plugin as a reference example:
- **Source**: https://github.com/ivannovak/glide-plugin-docker
- Demonstrates best practices for production plugins
- Shows proper SDK usage and command organization

## Example Plugin Structure

The ACME plugin example that was here has been moved to demonstrate proper plugin hosting:
- It should be in its own repository: `github.com/acme/glide-plugin-acme`
- The build process imports it via build tags

```go
// cmd/glide/plugins_acme.go
//go:build acme
// +build acme

package main

import (
    _ "github.com/acme/glide-plugin-acme"
)
```

This keeps the Glide repository clean and focused on core functionality.