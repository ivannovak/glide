package config

import (
	"reflect"
	"testing"
)

func TestMigrator_SimpleMigration(t *testing.T) {
	migrator := NewMigrator()

	// Define a migration from v1 to v2 that renames a field
	migrator.AddMigration(1, 2, func(old map[string]interface{}) (map[string]interface{}, error) {
		new := make(map[string]interface{})
		// Rename "endpoint" to "api_endpoint"
		if endpoint, ok := old["endpoint"].(string); ok {
			new["api_endpoint"] = endpoint
		}
		// Add new field with default
		new["timeout"] = 30
		// Copy unchanged fields
		if enabled, ok := old["enabled"]; ok {
			new["enabled"] = enabled
		}
		return new, nil
	})

	oldConfig := map[string]interface{}{
		"endpoint": "api.example.com",
		"enabled":  true,
	}

	newConfig, err := migrator.Migrate(oldConfig, 1, 2)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Check that old field was renamed
	if _, exists := newConfig["endpoint"]; exists {
		t.Error("Old field 'endpoint' should not exist after migration")
	}

	if apiEndpoint, ok := newConfig["api_endpoint"].(string); !ok || apiEndpoint != "api.example.com" {
		t.Errorf("Expected api_endpoint='api.example.com', got %v", newConfig["api_endpoint"])
	}

	// Check that new field was added
	if timeout, ok := newConfig["timeout"].(int); !ok || timeout != 30 {
		t.Errorf("Expected timeout=30, got %v", newConfig["timeout"])
	}

	// Check that unchanged field was copied
	if enabled, ok := newConfig["enabled"].(bool); !ok || !enabled {
		t.Errorf("Expected enabled=true, got %v", newConfig["enabled"])
	}
}

func TestMigrator_MultiHopMigration(t *testing.T) {
	migrator := NewMigrator()

	// Migration v1 -> v2: rename field
	migrator.AddMigration(1, 2, func(old map[string]interface{}) (map[string]interface{}, error) {
		new := make(map[string]interface{})
		new["name"] = old["username"] // Rename username to name
		return new, nil
	})

	// Migration v2 -> v3: add new field
	migrator.AddMigration(2, 3, func(old map[string]interface{}) (map[string]interface{}, error) {
		new := make(map[string]interface{})
		new["name"] = old["name"]
		new["version"] = "3.0" // Add version field
		return new, nil
	})

	oldConfig := map[string]interface{}{
		"username": "john_doe",
	}

	// Migrate from v1 to v3 (should apply both migrations)
	newConfig, err := migrator.Migrate(oldConfig, 1, 3)
	if err != nil {
		t.Fatalf("Multi-hop migration failed: %v", err)
	}

	if name, ok := newConfig["name"].(string); !ok || name != "john_doe" {
		t.Errorf("Expected name='john_doe', got %v", newConfig["name"])
	}

	if version, ok := newConfig["version"].(string); !ok || version != "3.0" {
		t.Errorf("Expected version='3.0', got %v", newConfig["version"])
	}
}

func TestMigrator_CanMigrate(t *testing.T) {
	migrator := NewMigrator()

	migrator.AddMigration(1, 2, func(old map[string]interface{}) (map[string]interface{}, error) {
		return old, nil
	})
	migrator.AddMigration(2, 3, func(old map[string]interface{}) (map[string]interface{}, error) {
		return old, nil
	})

	tests := []struct {
		name  string
		from  int
		to    int
		canDo bool
	}{
		{"same version", 2, 2, true},
		{"one hop forward", 1, 2, true},
		{"two hops forward", 1, 3, true},
		{"backward not allowed", 3, 1, false},
		{"missing path", 1, 4, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			can := migrator.CanMigrate(tt.from, tt.to)
			if can != tt.canDo {
				t.Errorf("CanMigrate(%d, %d) = %v, want %v", tt.from, tt.to, can, tt.canDo)
			}
		})
	}
}

func TestAutoMigrate(t *testing.T) {
	migrator := NewMigrator()

	migrator.AddMigration(1, 2, func(old map[string]interface{}) (map[string]interface{}, error) {
		new := make(map[string]interface{})
		new["name_v2"] = old["name"]
		return new, nil
	})

	vc := &VersionedConfig{
		Version: 1,
		Data: map[string]interface{}{
			"name": "test",
		},
	}

	err := AutoMigrate(vc, 2, migrator)
	if err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	if vc.Version != 2 {
		t.Errorf("Expected version 2, got %d", vc.Version)
	}

	if nameV2, ok := vc.Data["name_v2"].(string); !ok || nameV2 != "test" {
		t.Errorf("Expected name_v2='test', got %v", vc.Data["name_v2"])
	}
}

func TestBackwardCompatibilityLayer_LegacyKeys(t *testing.T) {
	compat := NewBackwardCompatibilityLayer()
	compat.AddLegacyKey("endpoint", "api_endpoint")
	compat.AddLegacyKey("ttl", "timeout")

	config := map[string]interface{}{
		"endpoint": "api.example.com",
		"ttl":      60,
		"enabled":  true,
	}

	compat.Transform(config)

	// Old keys should be removed
	if _, exists := config["endpoint"]; exists {
		t.Error("Legacy key 'endpoint' should be removed")
	}
	if _, exists := config["ttl"]; exists {
		t.Error("Legacy key 'ttl' should be removed")
	}

	// New keys should exist
	if apiEndpoint, ok := config["api_endpoint"].(string); !ok || apiEndpoint != "api.example.com" {
		t.Errorf("Expected api_endpoint='api.example.com', got %v", config["api_endpoint"])
	}

	if timeout, ok := config["timeout"].(int); !ok || timeout != 60 {
		t.Errorf("Expected timeout=60, got %v", config["timeout"])
	}

	// Unchanged fields should remain
	if enabled, ok := config["enabled"].(bool); !ok || !enabled {
		t.Errorf("Expected enabled=true, got %v", config["enabled"])
	}
}

func TestBackwardCompatibilityLayer_Transforms(t *testing.T) {
	compat := NewBackwardCompatibilityLayer()

	// Transform seconds to milliseconds
	compat.AddTransform("timeout", func(v interface{}) interface{} {
		if seconds, ok := v.(int); ok {
			return seconds * 1000
		}
		return v
	})

	config := map[string]interface{}{
		"timeout": 30, // 30 seconds
	}

	compat.Transform(config)

	if timeout, ok := config["timeout"].(int); !ok || timeout != 30000 {
		t.Errorf("Expected timeout=30000 (milliseconds), got %v", config["timeout"])
	}
}

func TestBackwardCompatibilityLayer_Combined(t *testing.T) {
	compat := NewBackwardCompatibilityLayer()

	// Rename field
	compat.AddLegacyKey("connect_timeout", "timeout")

	// Transform old value (seconds) to new format (milliseconds)
	compat.AddTransform("timeout", func(v interface{}) interface{} {
		if seconds, ok := v.(int); ok {
			return seconds * 1000
		}
		return v
	})

	config := map[string]interface{}{
		"connect_timeout": 30, // Old field name, old format (seconds)
	}

	compat.Transform(config)

	// Old field should be removed
	if _, exists := config["connect_timeout"]; exists {
		t.Error("Legacy key 'connect_timeout' should be removed")
	}

	// New field should exist with transformed value
	if timeout, ok := config["timeout"].(int); !ok || timeout != 30000 {
		t.Errorf("Expected timeout=30000 (renamed and converted to ms), got %v", config["timeout"])
	}
}

func TestDetectVersion(t *testing.T) {
	tests := []struct {
		name   string
		config map[string]interface{}
		want   int
	}{
		{
			name:   "no version field defaults to 1",
			config: map[string]interface{}{"name": "test"},
			want:   1,
		},
		{
			name:   "explicit int version",
			config: map[string]interface{}{"version": 3, "name": "test"},
			want:   3,
		},
		{
			name:   "float64 version (from JSON)",
			config: map[string]interface{}{"version": float64(2), "name": "test"},
			want:   2,
		},
		{
			name:   "string version",
			config: map[string]interface{}{"version": "4", "name": "test"},
			want:   4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectVersion(tt.config)
			if got != tt.want {
				t.Errorf("DetectVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMigrator_ErrorHandling(t *testing.T) {
	migrator := NewMigrator()

	tests := []struct {
		name    string
		from    int
		to      int
		wantErr bool
	}{
		{
			name:    "backward migration not allowed",
			from:    2,
			to:      1,
			wantErr: true,
		},
		{
			name:    "missing migration path",
			from:    1,
			to:      5,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := map[string]interface{}{"test": "data"}
			_, err := migrator.Migrate(config, tt.from, tt.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("Migrate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBackwardCompatibilityLayer_NoOverwrite(t *testing.T) {
	compat := NewBackwardCompatibilityLayer()
	compat.AddLegacyKey("old_name", "new_name")

	// Config has both old and new keys
	config := map[string]interface{}{
		"old_name": "old_value",
		"new_name": "new_value",
	}

	compat.Transform(config)

	// New value should be preserved (not overwritten by old value)
	if newName, ok := config["new_name"].(string); !ok || newName != "new_value" {
		t.Errorf("New value should not be overwritten, got %v", config["new_name"])
	}

	// Old key should still be removed
	if _, exists := config["old_name"]; exists {
		t.Error("Old key should be removed even when new key exists")
	}
}

func TestIsZeroValue(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  bool
	}{
		{"zero int", 0, true},
		{"non-zero int", 5, false},
		{"empty string", "", true},
		{"non-empty string", "test", false},
		{"false bool", false, true},
		{"true bool", true, false},
		{"nil pointer", (*int)(nil), true},
		{"empty slice", []string{}, true},
		{"non-empty slice", []string{"a"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := reflect.ValueOf(tt.value)
			got := isZeroValue(v)
			if got != tt.want {
				t.Errorf("isZeroValue(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}
