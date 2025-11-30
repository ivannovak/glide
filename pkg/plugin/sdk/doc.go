// Package sdk provides the Plugin SDK for building Glide plugins.
//
// This package contains the plugin manager, discovery, validation, lifecycle
// management, and utilities for plugin development. For building plugins,
// see the sdk/v2 subpackage which provides the recommended plugin interface.
//
// # Plugin Manager
//
// The Manager handles plugin discovery, loading, and lifecycle:
//
//	mgr := sdk.NewManager(sdk.DefaultConfig())
//
//	// Discover plugins without loading them (fast)
//	err := mgr.DiscoverPluginsLazy()
//
//	// Get a specific plugin (loads on-demand)
//	plugin, err := mgr.GetPlugin("my-plugin")
//
//	// Shutdown all loaded plugins
//	mgr.Shutdown()
//
// # Plugin Discovery
//
// Plugins are discovered from configured directories:
//
//	config := sdk.DefaultConfig()
//	config.PluginDirs = []string{
//	    "/global/plugins",
//	    "/project/.glide/plugins",
//	}
//
// # Security Validation
//
// Plugins are validated before loading:
//
//   - Binary permissions (not world-writable)
//
//   - Checksum verification (optional)
//
//   - Security mode enforcement
//
//     validator := sdk.NewValidator()
//     if err := validator.ValidateBinary(pluginPath); err != nil {
//     log.Error("Plugin failed validation", "error", err)
//     }
//
// # Lifecycle Management
//
// Plugins go through defined lifecycle states:
//
//   - Discovered → Loading → Ready → Running → Stopping → Stopped
//
//     tracker := sdk.NewStateTracker()
//     tracker.TransitionTo(sdk.StateLoading, "Loading plugin")
//     tracker.TransitionTo(sdk.StateReady, "Plugin loaded")
//
// # Dependency Resolution
//
// Plugins can declare dependencies on other plugins:
//
//	resolver := sdk.NewDependencyResolver()
//	resolver.AddDependency("my-plugin", "base-plugin", ">=1.0.0")
//	order, err := resolver.Resolve()
//
// See sdk/v2 for the recommended plugin development interface.
package sdk
