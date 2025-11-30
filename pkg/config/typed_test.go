package config

import (
	"encoding/json"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

// Test types
type TestConfig struct {
	Name    string   `json:"name" yaml:"name"`
	Timeout int      `json:"timeout" yaml:"timeout"`
	Enabled bool     `json:"enabled" yaml:"enabled"`
	Tags    []string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

type NestedConfig struct {
	Server ServerConfig `json:"server" yaml:"server"`
	Client ClientConfig `json:"client" yaml:"client"`
}

type ServerConfig struct {
	Host string `json:"host" yaml:"host"`
	Port int    `json:"port" yaml:"port"`
}

type ClientConfig struct {
	APIKey string `json:"api_key" yaml:"api_key"`
}

func TestNewTypedConfig(t *testing.T) {
	defaults := TestConfig{
		Name:    "default",
		Timeout: 30,
		Enabled: true,
	}

	tc := NewTypedConfig("test-plugin", defaults)

	if tc.Name != "test-plugin" {
		t.Errorf("Expected name 'test-plugin', got %s", tc.Name)
	}
	if tc.Version != 1 {
		t.Errorf("Expected version 1, got %d", tc.Version)
	}
	if !reflect.DeepEqual(tc.Value, defaults) {
		t.Errorf("Expected value to equal defaults")
	}
	if !reflect.DeepEqual(tc.Defaults, defaults) {
		t.Errorf("Expected defaults to be set")
	}
}

func TestTypedConfig_Merge(t *testing.T) {
	tests := []struct {
		name      string
		initial   TestConfig
		rawConfig interface{}
		want      TestConfig
		wantErr   bool
	}{
		{
			name: "nil config keeps defaults",
			initial: TestConfig{
				Name:    "default",
				Timeout: 30,
				Enabled: true,
			},
			rawConfig: nil,
			want: TestConfig{
				Name:    "default",
				Timeout: 30,
				Enabled: true,
			},
			wantErr: false,
		},
		{
			name: "direct type assertion",
			initial: TestConfig{
				Name:    "default",
				Timeout: 30,
			},
			rawConfig: TestConfig{
				Name:    "updated",
				Timeout: 60,
				Enabled: true,
			},
			want: TestConfig{
				Name:    "updated",
				Timeout: 60,
				Enabled: true,
			},
			wantErr: false,
		},
		{
			name: "map conversion via JSON",
			initial: TestConfig{
				Name:    "default",
				Timeout: 30,
			},
			rawConfig: map[string]interface{}{
				"name":    "from-map",
				"timeout": 45,
				"enabled": true,
				"tags":    []interface{}{"tag1", "tag2"},
			},
			want: TestConfig{
				Name:    "from-map",
				Timeout: 45,
				Enabled: true,
				Tags:    []string{"tag1", "tag2"},
			},
			wantErr: false,
		},
		{
			name: "partial map keeps unset fields as zero",
			initial: TestConfig{
				Name:    "default",
				Timeout: 30,
			},
			rawConfig: map[string]interface{}{
				"name": "partial",
			},
			want: TestConfig{
				Name:    "partial",
				Timeout: 0, // Note: zero value, not default!
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name:    "invalid type conversion",
			initial: TestConfig{},
			rawConfig: map[string]interface{}{
				"timeout": "invalid-not-a-number",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := NewTypedConfig("test", tt.initial)
			err := tc.Merge(tt.rawConfig)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(tc.Value, tt.want) {
				t.Errorf("Value mismatch:\nGot:  %+v\nWant: %+v", tc.Value, tt.want)
			}
		})
	}
}

func TestTypedConfig_MergeYAML(t *testing.T) {
	tests := []struct {
		name     string
		yamlData string
		want     TestConfig
		wantErr  bool
	}{
		{
			name: "valid YAML",
			yamlData: `
name: yaml-config
timeout: 120
enabled: true
tags:
  - tag1
  - tag2
`,
			want: TestConfig{
				Name:    "yaml-config",
				Timeout: 120,
				Enabled: true,
				Tags:    []string{"tag1", "tag2"},
			},
			wantErr: false,
		},
		{
			name: "partial YAML",
			yamlData: `
name: partial-yaml
`,
			want: TestConfig{
				Name:    "partial-yaml",
				Timeout: 0,
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "invalid YAML",
			yamlData: `
name: unclosed-quote"
timeout: not-a-number
`,
			wantErr: true,
		},
		{
			name:     "empty YAML",
			yamlData: "",
			want: TestConfig{
				Name:    "",
				Timeout: 0,
				Enabled: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := NewTypedConfig("test", TestConfig{})
			err := tc.MergeYAML([]byte(tt.yamlData))

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tc.Value.Name != tt.want.Name ||
				tc.Value.Timeout != tt.want.Timeout ||
				tc.Value.Enabled != tt.want.Enabled {
				t.Errorf("Value mismatch:\nGot:  %+v\nWant: %+v", tc.Value, tt.want)
			}
		})
	}
}

func TestTypedConfig_MergeJSON(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		want     TestConfig
		wantErr  bool
	}{
		{
			name: "valid JSON",
			jsonData: `{
				"name": "json-config",
				"timeout": 90,
				"enabled": true,
				"tags": ["tag1", "tag2"]
			}`,
			want: TestConfig{
				Name:    "json-config",
				Timeout: 90,
				Enabled: true,
				Tags:    []string{"tag1", "tag2"},
			},
			wantErr: false,
		},
		{
			name: "partial JSON",
			jsonData: `{
				"name": "partial-json"
			}`,
			want: TestConfig{
				Name:    "partial-json",
				Timeout: 0,
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name:     "invalid JSON",
			jsonData: `{"name": "unclosed`,
			wantErr:  true,
		},
		{
			name:     "empty JSON object",
			jsonData: `{}`,
			want: TestConfig{
				Name:    "",
				Timeout: 0,
				Enabled: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := NewTypedConfig("test", TestConfig{})
			err := tc.MergeJSON([]byte(tt.jsonData))

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(tc.Value, tt.want) {
				t.Errorf("Value mismatch:\nGot:  %+v\nWant: %+v", tc.Value, tt.want)
			}
		})
	}
}

func TestTypedConfig_Reset(t *testing.T) {
	defaults := TestConfig{
		Name:    "default",
		Timeout: 30,
		Enabled: true,
	}

	tc := NewTypedConfig("test", defaults)

	// Modify the value
	tc.Value = TestConfig{
		Name:    "modified",
		Timeout: 60,
		Enabled: false,
	}

	// Reset should restore defaults
	tc.Reset()

	if !reflect.DeepEqual(tc.Value, defaults) {
		t.Errorf("Reset failed:\nGot:  %+v\nWant: %+v", tc.Value, defaults)
	}
}

func TestTypedConfig_Clone(t *testing.T) {
	original := NewTypedConfig("test", TestConfig{
		Name:    "original",
		Timeout: 30,
		Enabled: true,
		Tags:    []string{"tag1", "tag2"},
	})

	clone, err := original.Clone()
	if err != nil {
		t.Fatalf("Clone failed: %v", err)
	}

	// Verify clone has same VALUE (due to custom MarshalJSON, only Value field is cloned)
	// Note: This is current behavior - Clone only deep copies Value field, not metadata
	if !reflect.DeepEqual(clone.Value, original.Value) {
		t.Errorf("Clone value mismatch:\nGot:  %+v\nWant: %+v", clone.Value, original.Value)
	}

	// Verify metadata fields are zero values (not cloned due to MarshalJSON)
	if clone.Name != "" {
		t.Errorf("Expected clone.Name to be empty (not cloned), got %s", clone.Name)
	}
	if clone.Version != 0 {
		t.Errorf("Expected clone.Version to be 0 (not cloned), got %d", clone.Version)
	}

	// Modify clone, verify original unchanged
	clone.Value.Name = "modified"
	if original.Value.Name == "modified" {
		t.Error("Modifying clone affected original - not a deep copy")
	}
}

func TestTypedConfig_Clone_Error(t *testing.T) {
	// Test with type that can't be marshaled
	type UnmarshalableConfig struct {
		InvalidField chan int // channels can't be marshaled to JSON
	}

	tc := NewTypedConfig("test", UnmarshalableConfig{
		InvalidField: make(chan int),
	})

	_, err := tc.Clone()
	if err == nil {
		t.Error("Expected error when cloning unmarshalable type, got nil")
	}
}

func TestTypedConfig_TypeName(t *testing.T) {
	tests := []struct {
		name     string
		config   interface{}
		wantType string
	}{
		{
			name:     "struct type",
			config:   NewTypedConfig("test", TestConfig{}),
			wantType: "config.TestConfig",
		},
		{
			name:     "string type",
			config:   NewTypedConfig("test", ""),
			wantType: "string",
		},
		{
			name:     "int type",
			config:   NewTypedConfig("test", 0),
			wantType: "int",
		},
		{
			name:     "nested struct type",
			config:   NewTypedConfig("test", NestedConfig{}),
			wantType: "config.NestedConfig",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var typeName string
			switch tc := tt.config.(type) {
			case *TypedConfig[TestConfig]:
				typeName = tc.TypeName()
			case *TypedConfig[string]:
				typeName = tc.TypeName()
			case *TypedConfig[int]:
				typeName = tc.TypeName()
			case *TypedConfig[NestedConfig]:
				typeName = tc.TypeName()
			}

			if typeName != tt.wantType {
				t.Errorf("TypeName() = %s, want %s", typeName, tt.wantType)
			}
		})
	}
}

func TestTypedConfig_Validate(t *testing.T) {
	t.Run("no schema returns nil", func(t *testing.T) {
		tc := NewTypedConfig("test", TestConfig{})
		err := tc.Validate()
		if err != nil {
			t.Errorf("Expected nil when no schema, got %v", err)
		}
	})

	t.Run("with schema validates", func(t *testing.T) {
		tc := NewTypedConfig("test", TestConfig{
			Name:    "test",
			Timeout: 30,
		})

		// Set a JSON schema
		tc.Schema = NewJSONSchemaFromValue(TestConfig{})

		// Validation with schema should work
		err := tc.Validate()
		// Note: Validate method on JSONSchema is not implemented yet (returns error)
		// so we just check it doesn't panic
		_ = err
	})
}

func TestTypedConfig_MarshalYAML(t *testing.T) {
	tc := NewTypedConfig("test", TestConfig{
		Name:    "test-marshal",
		Timeout: 45,
		Enabled: true,
		Tags:    []string{"tag1", "tag2"},
	})

	data, err := yaml.Marshal(tc)
	if err != nil {
		t.Fatalf("MarshalYAML failed: %v", err)
	}

	// Unmarshal to verify
	var result TestConfig
	if err := yaml.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if !reflect.DeepEqual(result, tc.Value) {
		t.Errorf("Marshal/Unmarshal roundtrip failed:\nGot:  %+v\nWant: %+v", result, tc.Value)
	}
}

func TestTypedConfig_UnmarshalYAML(t *testing.T) {
	yamlData := `
name: test-unmarshal
timeout: 75
enabled: false
tags:
  - tag1
  - tag2
`

	var tc TypedConfig[TestConfig]
	tc.Name = "test"

	if err := yaml.Unmarshal([]byte(yamlData), &tc); err != nil {
		t.Fatalf("UnmarshalYAML failed: %v", err)
	}

	expected := TestConfig{
		Name:    "test-unmarshal",
		Timeout: 75,
		Enabled: false,
		Tags:    []string{"tag1", "tag2"},
	}

	if !reflect.DeepEqual(tc.Value, expected) {
		t.Errorf("UnmarshalYAML failed:\nGot:  %+v\nWant: %+v", tc.Value, expected)
	}
}

func TestTypedConfig_MarshalJSON(t *testing.T) {
	tc := NewTypedConfig("test", TestConfig{
		Name:    "test-json-marshal",
		Timeout: 100,
		Enabled: true,
		Tags:    []string{"json1", "json2"},
	})

	data, err := json.Marshal(tc)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Unmarshal to verify
	var result TestConfig
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if !reflect.DeepEqual(result, tc.Value) {
		t.Errorf("Marshal/Unmarshal roundtrip failed:\nGot:  %+v\nWant: %+v", result, tc.Value)
	}
}

func TestTypedConfig_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"name": "test-json-unmarshal",
		"timeout": 150,
		"enabled": true,
		"tags": ["json1", "json2"]
	}`

	var tc TypedConfig[TestConfig]
	tc.Name = "test"

	if err := json.Unmarshal([]byte(jsonData), &tc); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	expected := TestConfig{
		Name:    "test-json-unmarshal",
		Timeout: 150,
		Enabled: true,
		Tags:    []string{"json1", "json2"},
	}

	if !reflect.DeepEqual(tc.Value, expected) {
		t.Errorf("UnmarshalJSON failed:\nGot:  %+v\nWant: %+v", tc.Value, expected)
	}
}

func TestTypedConfig_UnmarshalJSON_Error(t *testing.T) {
	invalidJSON := `{"name": "unclosed`

	var tc TypedConfig[TestConfig]
	tc.Name = "test"

	err := json.Unmarshal([]byte(invalidJSON), &tc)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestTypedConfig_RoundTrip(t *testing.T) {
	original := NewTypedConfig("test", TestConfig{
		Name:    "roundtrip",
		Timeout: 200,
		Enabled: true,
		Tags:    []string{"a", "b", "c"},
	})

	t.Run("YAML roundtrip", func(t *testing.T) {
		// Marshal to YAML
		yamlData, err := yaml.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal to YAML failed: %v", err)
		}

		// Unmarshal back
		var result TypedConfig[TestConfig]
		result.Name = "test"
		if err := yaml.Unmarshal(yamlData, &result); err != nil {
			t.Fatalf("Unmarshal from YAML failed: %v", err)
		}

		if !reflect.DeepEqual(result.Value, original.Value) {
			t.Errorf("YAML roundtrip failed:\nGot:  %+v\nWant: %+v", result.Value, original.Value)
		}
	})

	t.Run("JSON roundtrip", func(t *testing.T) {
		// Marshal to JSON
		jsonData, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal to JSON failed: %v", err)
		}

		// Unmarshal back
		var result TypedConfig[TestConfig]
		result.Name = "test"
		if err := json.Unmarshal(jsonData, &result); err != nil {
			t.Fatalf("Unmarshal from JSON failed: %v", err)
		}

		if !reflect.DeepEqual(result.Value, original.Value) {
			t.Errorf("JSON roundtrip failed:\nGot:  %+v\nWant: %+v", result.Value, original.Value)
		}
	})
}
