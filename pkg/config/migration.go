package config

import (
	"fmt"
)

// Migrator handles configuration versioning and migration between schema versions.
//
// This allows plugins to evolve their configuration schemas over time while
// maintaining backward compatibility with older config files.
//
// Example usage:
//
//	migrator := NewMigrator()
//	migrator.AddMigration(1, 2, func(old map[string]interface{}) (map[string]interface{}, error) {
//	    // Migrate from v1 to v2
//	    new := make(map[string]interface{})
//	    new["api_endpoint"] = old["endpoint"]  // Renamed field
//	    new["timeout"] = 30                    // New field with default
//	    return new, nil
//	})
type Migrator struct {
	migrations map[string]Migration // Key: "fromVersion->toVersion"
}

// Migration defines a function that migrates config from one version to another.
// It takes the old configuration and returns the new configuration or an error.
type Migration func(old map[string]interface{}) (map[string]interface{}, error)

// NewMigrator creates a new configuration migrator.
func NewMigrator() *Migrator {
	return &Migrator{
		migrations: make(map[string]Migration),
	}
}

// AddMigration registers a migration function for moving from one version to another.
//
// Parameters:
//   - fromVersion: The source schema version
//   - toVersion: The target schema version
//   - migrationFn: Function that performs the migration
//
// Example:
//
//	migrator.AddMigration(1, 2, func(old map[string]interface{}) (map[string]interface{}, error) {
//	    new := make(map[string]interface{})
//	    // Rename field
//	    if endpoint, ok := old["endpoint"].(string); ok {
//	        new["api_endpoint"] = endpoint
//	    }
//	    // Add new field with default
//	    new["timeout"] = 30
//	    // Copy unchanged fields
//	    new["enabled"] = old["enabled"]
//	    return new, nil
//	})
func (m *Migrator) AddMigration(fromVersion, toVersion int, migrationFn Migration) {
	key := fmt.Sprintf("%d->%d", fromVersion, toVersion)
	m.migrations[key] = migrationFn
}

// Migrate migrates configuration from one version to another.
// If multiple hops are needed, it applies migrations sequentially.
//
// Parameters:
//   - config: The configuration to migrate
//   - fromVersion: The current schema version
//   - toVersion: The target schema version
//
// Returns the migrated configuration or an error if migration fails.
//
// Example:
//
//	// Migrate config from v1 to v3 (applies v1->v2, then v2->v3)
//	newConfig, err := migrator.Migrate(oldConfig, 1, 3)
func (m *Migrator) Migrate(config map[string]interface{}, fromVersion, toVersion int) (map[string]interface{}, error) {
	if fromVersion == toVersion {
		return config, nil
	}

	if fromVersion > toVersion {
		return nil, fmt.Errorf("cannot migrate backwards from v%d to v%d", fromVersion, toVersion)
	}

	current := config
	currentVersion := fromVersion

	// Apply migrations sequentially
	for currentVersion < toVersion {
		nextVersion := currentVersion + 1
		key := fmt.Sprintf("%d->%d", currentVersion, nextVersion)

		migration, exists := m.migrations[key]
		if !exists {
			return nil, fmt.Errorf("no migration path from v%d to v%d", currentVersion, nextVersion)
		}

		var err error
		current, err = migration(current)
		if err != nil {
			return nil, fmt.Errorf("migration from v%d to v%d failed: %w", currentVersion, nextVersion, err)
		}

		currentVersion = nextVersion
	}

	return current, nil
}

// CanMigrate checks if a migration path exists from one version to another.
//
// Parameters:
//   - fromVersion: The source schema version
//   - toVersion: The target schema version
//
// Returns true if a migration path exists (possibly through multiple hops).
func (m *Migrator) CanMigrate(fromVersion, toVersion int) bool {
	if fromVersion == toVersion {
		return true
	}

	if fromVersion > toVersion {
		return false // Cannot migrate backwards
	}

	// Check if we have a complete path
	currentVersion := fromVersion
	for currentVersion < toVersion {
		nextVersion := currentVersion + 1
		key := fmt.Sprintf("%d->%d", currentVersion, nextVersion)

		if _, exists := m.migrations[key]; !exists {
			return false
		}

		currentVersion = nextVersion
	}

	return true
}

// VersionedConfig wraps a configuration with version information.
// This is useful for storing configs that include their schema version.
type VersionedConfig struct {
	Version int                    `json:"version" yaml:"version"`
	Data    map[string]interface{} `json:"data" yaml:"data"`
}

// AutoMigrate automatically migrates a versioned config to the target version.
//
// Parameters:
//   - vc: The versioned configuration
//   - targetVersion: The desired schema version
//   - migrator: The migrator with registered migration functions
//
// Returns an error if migration fails or is not possible.
//
// Example:
//
//	vc := &VersionedConfig{
//	    Version: 1,
//	    Data: map[string]interface{}{"endpoint": "api.example.com"},
//	}
//
//	err := AutoMigrate(vc, 2, migrator)
//	// vc.Version is now 2, vc.Data has been migrated
func AutoMigrate(vc *VersionedConfig, targetVersion int, migrator *Migrator) error {
	if vc.Version == targetVersion {
		return nil // Already at target version
	}

	if !migrator.CanMigrate(vc.Version, targetVersion) {
		return fmt.Errorf("no migration path from v%d to v%d", vc.Version, targetVersion)
	}

	migratedData, err := migrator.Migrate(vc.Data, vc.Version, targetVersion)
	if err != nil {
		return err
	}

	vc.Data = migratedData
	vc.Version = targetVersion
	return nil
}

// BackwardCompatibilityLayer provides automatic fallback to legacy config formats.
//
// This is useful during migration periods where you want to support both old
// and new config formats transparently.
type BackwardCompatibilityLayer struct {
	// LegacyKeys maps old field names to new field names
	LegacyKeys map[string]string

	// DefaultTransforms applies default transformations to legacy values
	DefaultTransforms map[string]func(interface{}) interface{}
}

// NewBackwardCompatibilityLayer creates a new compatibility layer.
func NewBackwardCompatibilityLayer() *BackwardCompatibilityLayer {
	return &BackwardCompatibilityLayer{
		LegacyKeys:        make(map[string]string),
		DefaultTransforms: make(map[string]func(interface{}) interface{}),
	}
}

// AddLegacyKey registers a mapping from an old field name to a new field name.
//
// Parameters:
//   - oldKey: The deprecated field name
//   - newKey: The current field name
//
// Example:
//
//	compat.AddLegacyKey("endpoint", "api_endpoint")
//	// Old configs using "endpoint" will be automatically mapped to "api_endpoint"
func (bcl *BackwardCompatibilityLayer) AddLegacyKey(oldKey, newKey string) {
	bcl.LegacyKeys[oldKey] = newKey
}

// AddTransform registers a value transformation for a specific field.
//
// Parameters:
//   - key: The field name
//   - transform: Function to transform the old value to the new format
//
// Example:
//
//	compat.AddTransform("timeout", func(v interface{}) interface{} {
//	    // Old format was in seconds, new format is in milliseconds
//	    if seconds, ok := v.(int); ok {
//	        return seconds * 1000
//	    }
//	    return v
//	})
func (bcl *BackwardCompatibilityLayer) AddTransform(key string, transform func(interface{}) interface{}) {
	bcl.DefaultTransforms[key] = transform
}

// Transform applies the compatibility layer to a configuration map.
// This updates the config in-place to use new field names and formats.
//
// Parameters:
//   - config: The configuration to transform
//
// Example:
//
//	oldConfig := map[string]interface{}{
//	    "endpoint": "api.example.com",  // Old field name
//	    "timeout": 30,                   // Old format (seconds)
//	}
//
//	compat.AddLegacyKey("endpoint", "api_endpoint")
//	compat.AddTransform("timeout", func(v interface{}) interface{} {
//	    return v.(int) * 1000  // Convert to milliseconds
//	})
//
//	compat.Transform(oldConfig)
//	// oldConfig is now:
//	// {
//	//     "api_endpoint": "api.example.com",
//	//     "timeout": 30000
//	// }
func (bcl *BackwardCompatibilityLayer) Transform(config map[string]interface{}) {
	// Apply legacy key mappings
	for oldKey, newKey := range bcl.LegacyKeys {
		if value, exists := config[oldKey]; exists {
			// If new key doesn't exist, copy from old key
			if _, newExists := config[newKey]; !newExists {
				config[newKey] = value
			}
			// Remove old key to avoid confusion
			delete(config, oldKey)
		}
	}

	// Apply value transformations
	for key, transform := range bcl.DefaultTransforms {
		if value, exists := config[key]; exists {
			config[key] = transform(value)
		}
	}
}

// DetectVersion attempts to detect the schema version from a configuration map.
// If no version field is found, returns 1 (assuming v1 config).
//
// Parameters:
//   - config: The configuration map
//
// Returns the detected version or 1 if no version field exists.
func DetectVersion(config map[string]interface{}) int {
	if version, exists := config["version"]; exists {
		switch v := version.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			// Try to parse string as int
			var intVersion int
			if _, err := fmt.Sscanf(v, "%d", &intVersion); err == nil {
				return intVersion
			}
		}
	}

	// No version field or unparseable - assume v1
	return 1
}
