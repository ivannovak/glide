// Package config provides configuration loading and management for Glide.
//
// This package handles loading, merging, and accessing configuration from
// multiple sources including the global config file, project configs, and
// environment variables.
//
// # Configuration Loading
//
// Load configuration with default paths:
//
//	loader := config.NewLoader()
//	cfg, err := loader.Load()
//	if err != nil {
//	    return err
//	}
//
//	fmt.Printf("Debug mode: %v\n", cfg.Debug)
//
// # Configuration Sources
//
// Configuration is loaded from multiple sources in order:
//  1. Built-in defaults
//  2. Global config (~/.glide/config.yml)
//  3. Project config (.glide.yml)
//  4. Environment variables (GLIDE_*)
//
// # Configuration Merging
//
// Project configs extend global config:
//
//	// Global config
//	docker:
//	  image: "myapp:latest"
//
//	// Project config
//	docker:
//	  compose: "docker-compose.dev.yml"
//
//	// Result: Both settings merged
//
// # Config Structure
//
// The Config struct contains all settings:
//
//	type Config struct {
//	    Debug        bool
//	    Format       string
//	    Docker       DockerConfig
//	    Plugins      map[string]interface{}
//	    Commands     []CommandConfig
//	}
//
// # Recursive Discovery
//
// Project config is discovered up the directory tree:
//
//	loader.LoadWithContext(projectCtx)
//	// Searches: ./glide.yml, ../.glide.yml, etc.
//
// # Security
//
// Path validation prevents directory traversal attacks:
//
//	// Config paths are validated before access
//	loader.LoadFrom("/path/to/config.yml")
package config
