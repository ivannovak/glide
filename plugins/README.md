# Glide Plugins Directory

This directory is intentionally empty in the main Glide repository.

## Why is this empty?

Plugins are designed to be hosted in **separate repositories**, not within the Glide core repository. This ensures:

1. **Clean separation** between core Glide and organization-specific functionality
2. **Independent versioning** of plugins
3. **Private plugins** can remain in private repositories
4. **No bloat** in the core Glide repository

## Where are plugins hosted?

Each organization hosts their plugins in their own repositories:

- **Your Plugin**: `github.com/yourorg/glide-plugin-yourorg`

## How to create a plugin

See the [Plugin Architecture Documentation](../docs/plugin-architecture.md) for detailed instructions.

## Example Plugin Structure

The ACME plugin example that was here has been moved to demonstrate proper plugin hosting:
- It should be in its own repository: `github.com/acme/glide-plugin-acme`
- The build process imports it via build tags

```go
// cmd/glid/plugins_acme.go
//go:build acme
// +build acme

package main

import (
    _ "github.com/acme/glide-plugin-acme"
)
```

This keeps the Glide repository clean and focused on core functionality.