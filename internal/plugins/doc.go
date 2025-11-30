// Package plugins provides internal plugin management for Glide.
//
// This package contains plugin loading logic, builtin plugins, and
// internal utilities for the plugin system. For plugin development,
// see pkg/plugin/sdk.
//
// # Builtin Plugins
//
// Glide includes builtin plugins for common operations:
//
//	builtin/
//	├── docker/     # Docker and Docker Compose integration
//	├── context/    # Context information commands
//	└── config/     # Configuration management
//
// # Plugin Loading
//
// Internal plugin loading process:
//
//  1. Discover plugins from standard paths
//  2. Validate plugin binaries
//  3. Load plugins on-demand
//  4. Register plugin commands
//
// # Builtin Registration
//
// Builtin plugins are registered at startup:
//
//	registry := plugins.NewBuiltinRegistry()
//	registry.RegisterAll(root)
//
// # Plugin Paths
//
// Standard plugin locations:
//
//	~/.glide/plugins/     # Global plugins
//	.glide/plugins/       # Project plugins
//
// See pkg/plugin for the public plugin API and SDK.
package plugins
