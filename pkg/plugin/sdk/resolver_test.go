package sdk

import (
	"reflect"
	"sort"
	"testing"
)

func TestNewDependencyResolver(t *testing.T) {
	resolver := NewDependencyResolver()
	if resolver == nil {
		t.Fatal("NewDependencyResolver() returned nil")
	}
}

func TestDependencyResolver_Resolve_Simple(t *testing.T) {
	resolver := NewDependencyResolver()

	t.Run("no dependencies", func(t *testing.T) {
		plugins := map[string]PluginMetadata{
			"plugin-a": {Name: "plugin-a", Version: "1.0.0"},
			"plugin-b": {Name: "plugin-b", Version: "1.0.0"},
			"plugin-c": {Name: "plugin-c", Version: "1.0.0"},
		}

		order, err := resolver.Resolve(plugins)
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}

		if len(order) != 3 {
			t.Errorf("got %d plugins, want 3", len(order))
		}

		// All plugins should be present (order doesn't matter when no dependencies)
		for _, name := range []string{"plugin-a", "plugin-b", "plugin-c"} {
			found := false
			for _, p := range order {
				if p == name {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("plugin %q not in load order", name)
			}
		}
	})

	t.Run("linear dependencies", func(t *testing.T) {
		plugins := map[string]PluginMetadata{
			"plugin-a": {
				Name:    "plugin-a",
				Version: "1.0.0",
			},
			"plugin-b": {
				Name:    "plugin-b",
				Version: "1.0.0",
				Dependencies: []PluginDependency{
					{Name: "plugin-a", Version: "^1.0.0"},
				},
			},
			"plugin-c": {
				Name:    "plugin-c",
				Version: "1.0.0",
				Dependencies: []PluginDependency{
					{Name: "plugin-b", Version: "^1.0.0"},
				},
			},
		}

		order, err := resolver.Resolve(plugins)
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}

		expected := []string{"plugin-a", "plugin-b", "plugin-c"}
		if !reflect.DeepEqual(order, expected) {
			t.Errorf("got order %v, want %v", order, expected)
		}
	})

	t.Run("diamond dependencies", func(t *testing.T) {
		plugins := map[string]PluginMetadata{
			"base": {
				Name:    "base",
				Version: "1.0.0",
			},
			"left": {
				Name:    "left",
				Version: "1.0.0",
				Dependencies: []PluginDependency{
					{Name: "base", Version: "^1.0.0"},
				},
			},
			"right": {
				Name:    "right",
				Version: "1.0.0",
				Dependencies: []PluginDependency{
					{Name: "base", Version: "^1.0.0"},
				},
			},
			"top": {
				Name:    "top",
				Version: "1.0.0",
				Dependencies: []PluginDependency{
					{Name: "left", Version: "^1.0.0"},
					{Name: "right", Version: "^1.0.0"},
				},
			},
		}

		order, err := resolver.Resolve(plugins)
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}

		if len(order) != 4 {
			t.Errorf("got %d plugins, want 4", len(order))
		}

		// Verify base comes first
		if order[0] != "base" {
			t.Errorf("base should be first, got %v", order)
		}

		// Verify top comes last
		if order[3] != "top" {
			t.Errorf("top should be last, got %v", order)
		}

		// Verify left and right come before top but after base
		leftIdx := indexOf(order, "left")
		rightIdx := indexOf(order, "right")
		baseIdx := indexOf(order, "base")
		topIdx := indexOf(order, "top")

		if leftIdx < baseIdx || leftIdx > topIdx {
			t.Errorf("left should be after base and before top, got order %v", order)
		}
		if rightIdx < baseIdx || rightIdx > topIdx {
			t.Errorf("right should be after base and before top, got order %v", order)
		}
	})
}

func TestDependencyResolver_Resolve_CyclicDependency(t *testing.T) {
	resolver := NewDependencyResolver()

	t.Run("direct cycle", func(t *testing.T) {
		plugins := map[string]PluginMetadata{
			"plugin-a": {
				Name:    "plugin-a",
				Version: "1.0.0",
				Dependencies: []PluginDependency{
					{Name: "plugin-b", Version: "^1.0.0"},
				},
			},
			"plugin-b": {
				Name:    "plugin-b",
				Version: "1.0.0",
				Dependencies: []PluginDependency{
					{Name: "plugin-a", Version: "^1.0.0"},
				},
			},
		}

		_, err := resolver.Resolve(plugins)
		if err == nil {
			t.Fatal("Resolve() should return error for cyclic dependency")
		}

		if _, ok := err.(*CyclicDependencyError); !ok {
			t.Errorf("expected CyclicDependencyError, got %T: %v", err, err)
		}
	})

	t.Run("indirect cycle", func(t *testing.T) {
		plugins := map[string]PluginMetadata{
			"plugin-a": {
				Name:    "plugin-a",
				Version: "1.0.0",
				Dependencies: []PluginDependency{
					{Name: "plugin-b", Version: "^1.0.0"},
				},
			},
			"plugin-b": {
				Name:    "plugin-b",
				Version: "1.0.0",
				Dependencies: []PluginDependency{
					{Name: "plugin-c", Version: "^1.0.0"},
				},
			},
			"plugin-c": {
				Name:    "plugin-c",
				Version: "1.0.0",
				Dependencies: []PluginDependency{
					{Name: "plugin-a", Version: "^1.0.0"},
				},
			},
		}

		_, err := resolver.Resolve(plugins)
		if err == nil {
			t.Fatal("Resolve() should return error for cyclic dependency")
		}

		if _, ok := err.(*CyclicDependencyError); !ok {
			t.Errorf("expected CyclicDependencyError, got %T: %v", err, err)
		}
	})
}

func TestDependencyResolver_Resolve_MissingDependency(t *testing.T) {
	resolver := NewDependencyResolver()

	t.Run("required dependency missing", func(t *testing.T) {
		plugins := map[string]PluginMetadata{
			"plugin-a": {
				Name:    "plugin-a",
				Version: "1.0.0",
				Dependencies: []PluginDependency{
					{Name: "missing-plugin", Version: "^1.0.0", Optional: false},
				},
			},
		}

		_, err := resolver.Resolve(plugins)
		if err == nil {
			t.Fatal("Resolve() should return error for missing required dependency")
		}

		if _, ok := err.(*MissingDependencyError); !ok {
			t.Errorf("expected MissingDependencyError, got %T: %v", err, err)
		}
	})

	t.Run("optional dependency missing", func(t *testing.T) {
		plugins := map[string]PluginMetadata{
			"plugin-a": {
				Name:    "plugin-a",
				Version: "1.0.0",
				Dependencies: []PluginDependency{
					{Name: "missing-plugin", Version: "^1.0.0", Optional: true},
				},
			},
		}

		order, err := resolver.Resolve(plugins)
		if err != nil {
			t.Fatalf("Resolve() should not error for missing optional dependency, got: %v", err)
		}

		if len(order) != 1 || order[0] != "plugin-a" {
			t.Errorf("got order %v, want [plugin-a]", order)
		}
	})
}

func TestDependencyResolver_Resolve_VersionMismatch(t *testing.T) {
	resolver := NewDependencyResolver()

	t.Run("required version mismatch", func(t *testing.T) {
		plugins := map[string]PluginMetadata{
			"plugin-a": {
				Name:    "plugin-a",
				Version: "1.0.0",
			},
			"plugin-b": {
				Name:    "plugin-b",
				Version: "1.0.0",
				Dependencies: []PluginDependency{
					{Name: "plugin-a", Version: "^2.0.0", Optional: false},
				},
			},
		}

		_, err := resolver.Resolve(plugins)
		if err == nil {
			t.Fatal("Resolve() should return error for version mismatch")
		}

		if _, ok := err.(*VersionMismatchError); !ok {
			t.Errorf("expected VersionMismatchError, got %T: %v", err, err)
		}
	})

	t.Run("optional version mismatch", func(t *testing.T) {
		plugins := map[string]PluginMetadata{
			"plugin-a": {
				Name:    "plugin-a",
				Version: "1.0.0",
			},
			"plugin-b": {
				Name:    "plugin-b",
				Version: "1.0.0",
				Dependencies: []PluginDependency{
					{Name: "plugin-a", Version: "^2.0.0", Optional: true},
				},
			},
		}

		order, err := resolver.Resolve(plugins)
		if err != nil {
			t.Fatalf("Resolve() should not error for optional version mismatch, got: %v", err)
		}

		// Both plugins should load, but order might vary
		if len(order) != 2 {
			t.Errorf("got %d plugins, want 2", len(order))
		}
	})
}

func TestDependencyResolver_Resolve_InvalidDependency(t *testing.T) {
	resolver := NewDependencyResolver()

	t.Run("empty dependency name", func(t *testing.T) {
		plugins := map[string]PluginMetadata{
			"plugin-a": {
				Name:    "plugin-a",
				Version: "1.0.0",
				Dependencies: []PluginDependency{
					{Name: "", Version: "^1.0.0"},
				},
			},
		}

		_, err := resolver.Resolve(plugins)
		if err == nil {
			t.Fatal("Resolve() should return error for invalid dependency")
		}

		if _, ok := err.(*DependencyError); !ok {
			t.Errorf("expected DependencyError, got %T: %v", err, err)
		}
	})

	t.Run("invalid version constraint", func(t *testing.T) {
		plugins := map[string]PluginMetadata{
			"plugin-a": {
				Name:    "plugin-a",
				Version: "1.0.0",
				Dependencies: []PluginDependency{
					{Name: "plugin-b", Version: "not-a-version"},
				},
			},
		}

		_, err := resolver.Resolve(plugins)
		if err == nil {
			t.Fatal("Resolve() should return error for invalid version constraint")
		}
	})
}

func TestDependencyResolver_ValidatePluginDependencies(t *testing.T) {
	resolver := NewDependencyResolver()

	t.Run("valid dependencies", func(t *testing.T) {
		plugin := PluginMetadata{
			Name:    "my-plugin",
			Version: "1.0.0",
			Dependencies: []PluginDependency{
				{Name: "docker", Version: "^1.0.0"},
			},
		}

		available := map[string]PluginMetadata{
			"docker": {Name: "docker", Version: "1.5.0"},
		}

		err := resolver.ValidatePluginDependencies(plugin, available)
		if err != nil {
			t.Errorf("ValidatePluginDependencies() error = %v", err)
		}
	})

	t.Run("missing required dependency", func(t *testing.T) {
		plugin := PluginMetadata{
			Name:    "my-plugin",
			Version: "1.0.0",
			Dependencies: []PluginDependency{
				{Name: "docker", Version: "^1.0.0", Optional: false},
			},
		}

		available := map[string]PluginMetadata{}

		err := resolver.ValidatePluginDependencies(plugin, available)
		if err == nil {
			t.Fatal("ValidatePluginDependencies() should return error for missing dependency")
		}

		if _, ok := err.(*MissingDependencyError); !ok {
			t.Errorf("expected MissingDependencyError, got %T", err)
		}
	})

	t.Run("version mismatch", func(t *testing.T) {
		plugin := PluginMetadata{
			Name:    "my-plugin",
			Version: "1.0.0",
			Dependencies: []PluginDependency{
				{Name: "docker", Version: "^2.0.0", Optional: false},
			},
		}

		available := map[string]PluginMetadata{
			"docker": {Name: "docker", Version: "1.5.0"},
		}

		err := resolver.ValidatePluginDependencies(plugin, available)
		if err == nil {
			t.Fatal("ValidatePluginDependencies() should return error for version mismatch")
		}

		if _, ok := err.(*VersionMismatchError); !ok {
			t.Errorf("expected VersionMismatchError, got %T", err)
		}
	})
}

func TestDependencyResolver_GetDependencyInfo(t *testing.T) {
	resolver := NewDependencyResolver()

	plugins := map[string]PluginMetadata{
		"base": {
			Name:    "base",
			Version: "1.0.0",
		},
		"middle": {
			Name:    "middle",
			Version: "1.0.0",
			Dependencies: []PluginDependency{
				{Name: "base", Version: "^1.0.0"},
			},
		},
		"top": {
			Name:    "top",
			Version: "1.0.0",
			Dependencies: []PluginDependency{
				{Name: "middle", Version: "^1.0.0"},
			},
		},
	}

	info, err := resolver.GetDependencyInfo(plugins)
	if err != nil {
		t.Fatalf("GetDependencyInfo() error = %v", err)
	}

	// Check load order
	expected := []string{"base", "middle", "top"}
	if !reflect.DeepEqual(info.LoadOrder, expected) {
		t.Errorf("LoadOrder = %v, want %v", info.LoadOrder, expected)
	}

	// Check direct dependencies
	if len(info.DirectDependencies["base"]) != 0 {
		t.Error("base should have no direct dependencies")
	}
	if len(info.DirectDependencies["middle"]) != 1 {
		t.Error("middle should have 1 direct dependency")
	}
	if len(info.DirectDependencies["top"]) != 1 {
		t.Error("top should have 1 direct dependency")
	}

	// Check all dependencies (transitive)
	if len(info.AllDependencies["base"]) != 0 {
		t.Error("base should have no transitive dependencies")
	}
	if len(info.AllDependencies["middle"]) != 1 {
		t.Error("middle should have 1 transitive dependency (base)")
	}
	// top should have both middle and base as transitive dependencies
	if len(info.AllDependencies["top"]) != 2 {
		t.Errorf("top should have 2 transitive dependencies, got %d: %v",
			len(info.AllDependencies["top"]), info.AllDependencies["top"])
	}
}

func TestDependencyResolver_Resolve_ComplexScenario(t *testing.T) {
	resolver := NewDependencyResolver()

	// Complex real-world scenario with multiple dependency chains
	plugins := map[string]PluginMetadata{
		"logger": {
			Name:    "logger",
			Version: "1.0.0",
		},
		"config": {
			Name:    "config",
			Version: "1.0.0",
		},
		"database": {
			Name:    "database",
			Version: "1.0.0",
			Dependencies: []PluginDependency{
				{Name: "logger", Version: "^1.0.0"},
				{Name: "config", Version: "^1.0.0"},
			},
		},
		"cache": {
			Name:    "cache",
			Version: "1.0.0",
			Dependencies: []PluginDependency{
				{Name: "logger", Version: "^1.0.0"},
			},
		},
		"api": {
			Name:    "api",
			Version: "1.0.0",
			Dependencies: []PluginDependency{
				{Name: "database", Version: "^1.0.0"},
				{Name: "cache", Version: "^1.0.0"},
				{Name: "logger", Version: "^1.0.0"},
			},
		},
		"web": {
			Name:    "web",
			Version: "1.0.0",
			Dependencies: []PluginDependency{
				{Name: "api", Version: "^1.0.0"},
			},
		},
	}

	order, err := resolver.Resolve(plugins)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if len(order) != 6 {
		t.Errorf("got %d plugins, want 6", len(order))
	}

	// Verify dependency constraints
	assertBefore(t, order, "logger", "database")
	assertBefore(t, order, "logger", "cache")
	assertBefore(t, order, "logger", "api")
	assertBefore(t, order, "config", "database")
	assertBefore(t, order, "database", "api")
	assertBefore(t, order, "cache", "api")
	assertBefore(t, order, "api", "web")

	// logger and config should be first (no dependencies)
	firstTwo := order[:2]
	sort.Strings(firstTwo)
	if !reflect.DeepEqual(firstTwo, []string{"config", "logger"}) {
		t.Errorf("first two plugins should be config and logger (in any order), got %v", firstTwo)
	}

	// web should be last
	if order[len(order)-1] != "web" {
		t.Errorf("web should be last, got %v", order)
	}
}

// Helper functions

func indexOf(slice []string, value string) int {
	for i, v := range slice {
		if v == value {
			return i
		}
	}
	return -1
}

func assertBefore(t *testing.T, order []string, before, after string) {
	t.Helper()
	beforeIdx := indexOf(order, before)
	afterIdx := indexOf(order, after)

	if beforeIdx == -1 {
		t.Errorf("%q not found in order", before)
		return
	}
	if afterIdx == -1 {
		t.Errorf("%q not found in order", after)
		return
	}
	if beforeIdx >= afterIdx {
		t.Errorf("%q should come before %q, got order: %v", before, after, order)
	}
}
