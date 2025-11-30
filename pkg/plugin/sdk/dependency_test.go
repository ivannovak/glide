package sdk

import (
	"testing"
)

func TestPluginDependency_String(t *testing.T) {
	tests := []struct {
		name string
		dep  PluginDependency
		want string
	}{
		{
			name: "required dependency",
			dep: PluginDependency{
				Name:     "docker",
				Version:  "^1.0.0",
				Optional: false,
			},
			want: "docker@^1.0.0",
		},
		{
			name: "optional dependency",
			dep: PluginDependency{
				Name:     "kubernetes",
				Version:  ">=2.0.0",
				Optional: true,
			},
			want: "kubernetes@>=2.0.0 (optional)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.dep.String(); got != tt.want {
				t.Errorf("PluginDependency.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPluginDependency_Validate(t *testing.T) {
	tests := []struct {
		name    string
		dep     PluginDependency
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid dependency",
			dep: PluginDependency{
				Name:    "docker",
				Version: "^1.0.0",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			dep: PluginDependency{
				Name:    "",
				Version: "^1.0.0",
			},
			wantErr: true,
			errMsg:  "dependency name cannot be empty",
		},
		{
			name: "empty version",
			dep: PluginDependency{
				Name:    "docker",
				Version: "",
			},
			wantErr: true,
			errMsg:  "version constraint cannot be empty",
		},
		{
			name: "invalid semver constraint",
			dep: PluginDependency{
				Name:    "docker",
				Version: "not-a-version",
			},
			wantErr: true,
			errMsg:  "invalid version constraint",
		},
		{
			name: "valid caret constraint",
			dep: PluginDependency{
				Name:    "docker",
				Version: "^1.2.3",
			},
			wantErr: false,
		},
		{
			name: "valid tilde constraint",
			dep: PluginDependency{
				Name:    "docker",
				Version: "~1.2.3",
			},
			wantErr: false,
		},
		{
			name: "valid range constraint",
			dep: PluginDependency{
				Name:    "docker",
				Version: ">=1.0.0 <2.0.0",
			},
			wantErr: false,
		},
		{
			name: "valid wildcard constraint",
			dep: PluginDependency{
				Name:    "docker",
				Version: "1.x",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dep.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("PluginDependency.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if !containsString(err.Error(), tt.errMsg) {
					t.Errorf("PluginDependency.Validate() error = %v, should contain %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestPluginDependency_SatisfiedBy(t *testing.T) {
	tests := []struct {
		name     string
		dep      PluginDependency
		version  string
		expected bool
	}{
		{
			name: "exact version match",
			dep: PluginDependency{
				Name:    "docker",
				Version: "1.2.3",
			},
			version:  "1.2.3",
			expected: true,
		},
		{
			name: "exact version mismatch",
			dep: PluginDependency{
				Name:    "docker",
				Version: "1.2.3",
			},
			version:  "1.2.4",
			expected: false,
		},
		{
			name: "caret constraint satisfied",
			dep: PluginDependency{
				Name:    "docker",
				Version: "^1.2.3",
			},
			version:  "1.5.0",
			expected: true,
		},
		{
			name: "caret constraint not satisfied (major version change)",
			dep: PluginDependency{
				Name:    "docker",
				Version: "^1.2.3",
			},
			version:  "2.0.0",
			expected: false,
		},
		{
			name: "tilde constraint satisfied",
			dep: PluginDependency{
				Name:    "docker",
				Version: "~1.2.3",
			},
			version:  "1.2.9",
			expected: true,
		},
		{
			name: "tilde constraint not satisfied (minor version change)",
			dep: PluginDependency{
				Name:    "docker",
				Version: "~1.2.3",
			},
			version:  "1.3.0",
			expected: false,
		},
		{
			name: "range constraint satisfied",
			dep: PluginDependency{
				Name:    "docker",
				Version: ">=1.0.0 <2.0.0",
			},
			version:  "1.5.0",
			expected: true,
		},
		{
			name: "range constraint not satisfied (too high)",
			dep: PluginDependency{
				Name:    "docker",
				Version: ">=1.0.0 <2.0.0",
			},
			version:  "2.0.0",
			expected: false,
		},
		{
			name: "range constraint not satisfied (too low)",
			dep: PluginDependency{
				Name:    "docker",
				Version: ">=1.0.0 <2.0.0",
			},
			version:  "0.9.0",
			expected: false,
		},
		{
			name: "invalid version string",
			dep: PluginDependency{
				Name:    "docker",
				Version: "^1.0.0",
			},
			version:  "not-a-version",
			expected: false,
		},
		{
			name: "wildcard constraint satisfied",
			dep: PluginDependency{
				Name:    "docker",
				Version: "1.x",
			},
			version:  "1.9.9",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.dep.SatisfiedBy(tt.version); got != tt.expected {
				t.Errorf("PluginDependency.SatisfiedBy(%q) = %v, want %v", tt.version, got, tt.expected)
			}
		})
	}
}

func TestDependencyGraph(t *testing.T) {
	t.Run("basic operations", func(t *testing.T) {
		graph := NewDependencyGraph()

		// Test empty graph
		if graph.HasPlugin("docker") {
			t.Error("empty graph should not have 'docker' plugin")
		}

		// Add plugin
		deps := []PluginDependency{
			{Name: "base", Version: "^1.0.0"},
		}
		graph.AddPlugin("docker", deps)

		// Test HasPlugin
		if !graph.HasPlugin("docker") {
			t.Error("graph should have 'docker' plugin after adding")
		}

		// Test GetDependencies
		gotDeps := graph.GetDependencies("docker")
		if len(gotDeps) != 1 {
			t.Errorf("got %d dependencies, want 1", len(gotDeps))
		}
		if gotDeps[0].Name != "base" {
			t.Errorf("got dependency name %q, want %q", gotDeps[0].Name, "base")
		}

		// Test AllPlugins
		allPlugins := graph.AllPlugins()
		if len(allPlugins) != 1 {
			t.Errorf("got %d plugins, want 1", len(allPlugins))
		}
		if allPlugins[0] != "docker" {
			t.Errorf("got plugin %q, want %q", allPlugins[0], "docker")
		}
	})

	t.Run("multiple plugins", func(t *testing.T) {
		graph := NewDependencyGraph()

		graph.AddPlugin("plugin-a", []PluginDependency{})
		graph.AddPlugin("plugin-b", []PluginDependency{
			{Name: "plugin-a", Version: "^1.0.0"},
		})
		graph.AddPlugin("plugin-c", []PluginDependency{
			{Name: "plugin-a", Version: "^1.0.0"},
			{Name: "plugin-b", Version: "^1.0.0"},
		})

		allPlugins := graph.AllPlugins()
		if len(allPlugins) != 3 {
			t.Errorf("got %d plugins, want 3", len(allPlugins))
		}

		// Verify dependencies
		if len(graph.GetDependencies("plugin-a")) != 0 {
			t.Error("plugin-a should have no dependencies")
		}
		if len(graph.GetDependencies("plugin-b")) != 1 {
			t.Error("plugin-b should have 1 dependency")
		}
		if len(graph.GetDependencies("plugin-c")) != 2 {
			t.Error("plugin-c should have 2 dependencies")
		}
	})
}

func TestDependencyError(t *testing.T) {
	t.Run("without cause", func(t *testing.T) {
		err := NewDependencyError("my-plugin", "something went wrong", nil)
		expected := `dependency error for plugin "my-plugin": something went wrong`
		if err.Error() != expected {
			t.Errorf("got error %q, want %q", err.Error(), expected)
		}
		if err.Unwrap() != nil {
			t.Error("Unwrap() should return nil when no cause")
		}
	})

	t.Run("with cause", func(t *testing.T) {
		cause := &MissingDependencyError{
			Plugin: "my-plugin",
			Dependency: PluginDependency{
				Name:    "docker",
				Version: "^1.0.0",
			},
		}
		err := NewDependencyError("my-plugin", "dependency missing", cause)
		if err.Unwrap() != cause {
			t.Error("Unwrap() should return the cause error")
		}
		if !containsString(err.Error(), "dependency missing") {
			t.Errorf("error should contain message: %s", err.Error())
		}
	})
}

func TestCyclicDependencyError(t *testing.T) {
	err := &CyclicDependencyError{
		Cycle: []string{"plugin-a", "plugin-b", "plugin-a"},
	}
	expectedMsg := "cyclic dependency detected: [plugin-a plugin-b plugin-a]"
	if err.Error() != expectedMsg {
		t.Errorf("got error %q, want %q", err.Error(), expectedMsg)
	}
}

func TestMissingDependencyError(t *testing.T) {
	err := &MissingDependencyError{
		Plugin: "my-plugin",
		Dependency: PluginDependency{
			Name:     "docker",
			Version:  "^1.0.0",
			Optional: false,
		},
	}
	expected := `plugin "my-plugin" requires missing dependency docker@^1.0.0`
	if err.Error() != expected {
		t.Errorf("got error %q, want %q", err.Error(), expected)
	}
}

func TestVersionMismatchError(t *testing.T) {
	err := &VersionMismatchError{
		Plugin: "my-plugin",
		Dependency: PluginDependency{
			Name:    "docker",
			Version: "^2.0.0",
		},
		ActualVersion:   "1.5.0",
		RequiredVersion: "^2.0.0",
	}
	expected := `plugin "my-plugin" requires docker@^2.0.0 but found version 1.5.0`
	if err.Error() != expected {
		t.Errorf("got error %q, want %q", err.Error(), expected)
	}
}

// Helper function for substring matching in tests
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
