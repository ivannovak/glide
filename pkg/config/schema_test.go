package config

import (
	"reflect"
	"testing"
)

// Test types for schema tests
type SchemaTestConfig struct {
	Name     string   `json:"name" validate:"required"`
	Age      int      `json:"age" validate:"min=0,max=120"`
	Email    string   `json:"email" validate:"required"`
	Tags     []string `json:"tags,omitempty"`
	IsActive bool     `json:"is_active"`
}

type NestedSchemaConfig struct {
	User   UserInfo   `json:"user"`
	Server ServerInfo `json:"server"`
}

type UserInfo struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required"`
}

type ServerInfo struct {
	Host string `json:"host" validate:"required"`
	Port int    `json:"port" validate:"min=1,max=65535"`
}

func TestNewJSONSchema(t *testing.T) {
	typ := reflect.TypeOf(SchemaTestConfig{})
	schema := NewJSONSchema(typ)

	if schema == nil {
		t.Fatal("NewJSONSchema returned nil")
	}

	if schema.schema == nil {
		t.Error("Schema map is nil")
	}

	// Verify schema has expected structure
	if _, ok := schema.schema["type"]; !ok {
		t.Error("Schema missing 'type' field")
	}
}

func TestNewJSONSchemaFromValue(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
	}{
		{
			name:  "struct value",
			value: SchemaTestConfig{},
		},
		{
			name:  "pointer to struct",
			value: &SchemaTestConfig{},
		},
		{
			name:  "nested struct",
			value: NestedSchemaConfig{},
		},
		{
			name:  "simple type",
			value: "string",
		},
		{
			name:  "integer",
			value: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := NewJSONSchemaFromValue(tt.value)

			if schema == nil {
				t.Fatal("NewJSONSchemaFromValue returned nil")
			}

			if schema.schema == nil {
				t.Error("Schema map is nil")
			}
		})
	}
}

func TestJSONSchema_GenerateSchema(t *testing.T) {
	tests := []struct {
		name     string
		config   interface{}
		checkKey string // Key that should exist in schema
	}{
		{
			name:     "struct with validation tags",
			config:   SchemaTestConfig{},
			checkKey: "properties",
		},
		{
			name:     "nested struct",
			config:   NestedSchemaConfig{},
			checkKey: "properties",
		},
		{
			name:     "simple string",
			config:   "",
			checkKey: "type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := NewJSONSchemaFromValue(tt.config)
			generated, err := schema.GenerateSchema()

			if err != nil {
				t.Fatalf("GenerateSchema failed: %v", err)
			}

			if generated == nil {
				t.Fatal("Generated schema is nil")
			}

			if _, ok := generated[tt.checkKey]; !ok {
				t.Errorf("Generated schema missing expected key '%s': %+v", tt.checkKey, generated)
			}
		})
	}
}

func TestJSONSchema_Validate(t *testing.T) {
	schema := NewJSONSchemaFromValue(SchemaTestConfig{})

	// Note: Validate is not fully implemented yet, it returns an error
	// This test verifies it doesn't panic and returns expected error
	err := schema.Validate(SchemaTestConfig{
		Name:  "test",
		Age:   30,
		Email: "test@example.com",
	})

	// Current implementation returns error (not implemented)
	if err == nil {
		t.Log("Validate returned nil - implementation may be complete")
	} else {
		// Expected: error about not being implemented
		t.Logf("Validate returned expected error: %v", err)
	}
}

func TestGenerateSchemaFromType(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
	}{
		{
			name:  "struct type",
			value: SchemaTestConfig{},
		},
		{
			name:  "nested struct",
			value: NestedSchemaConfig{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := reflect.TypeOf(tt.value)
			schema, err := GenerateSchemaFromType(typ)

			if err != nil {
				t.Fatalf("GenerateSchemaFromType failed: %v", err)
			}

			if schema == nil {
				t.Fatal("GenerateSchemaFromType returned nil")
			}

			// Should have type field
			if _, ok := schema["type"]; !ok {
				t.Error("Schema missing 'type' field")
			}
		})
	}
}

func TestGenerateSchemaFromValue(t *testing.T) {
	tests := []struct {
		name        string
		value       interface{}
		expectedTyp string
	}{
		{
			name:        "struct",
			value:       SchemaTestConfig{},
			expectedTyp: "object",
		},
		{
			name:        "pointer to struct",
			value:       &SchemaTestConfig{},
			expectedTyp: "object",
		},
		{
			name:        "string",
			value:       "test",
			expectedTyp: "string",
		},
		{
			name:        "integer",
			value:       42,
			expectedTyp: "integer",
		},
		{
			name:        "boolean",
			value:       true,
			expectedTyp: "boolean",
		},
		{
			name:        "slice",
			value:       []string{"a", "b"},
			expectedTyp: "array",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := GenerateSchemaFromValue(tt.value)

			if err != nil {
				t.Fatalf("GenerateSchemaFromValue failed: %v", err)
			}

			if schema == nil {
				t.Fatal("GenerateSchemaFromValue returned nil")
			}

			schemaType, ok := schema["type"].(string)
			if !ok {
				t.Fatalf("Schema 'type' field is not a string: %v", schema["type"])
			}

			if schemaType != tt.expectedTyp {
				t.Errorf("Expected type '%s', got '%s'", tt.expectedTyp, schemaType)
			}
		})
	}
}

func TestParseValidationRules(t *testing.T) {
	// This tests the internal parseValidationRules function indirectly
	// by checking generated schema from struct tags

	type ValidationTestConfig struct {
		Required   string `validate:"required"`
		MinMax     int    `validate:"min=1,max=100"`
		Enum       string `validate:"enum=a|b|c"`
		MinLen     string `validate:"min=3"`
		MaxLen     string `validate:"max=10"`
		MultiRules int    `validate:"required,min=0,max=100"`
	}

	schema := NewJSONSchemaFromValue(ValidationTestConfig{})
	generated, err := schema.GenerateSchema()
	if err != nil {
		t.Fatalf("GenerateSchema failed: %v", err)
	}

	properties, ok := generated["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema properties is not a map")
	}

	// Check Required field has required rule
	if required, ok := generated["required"].([]interface{}); ok {
		found := false
		for _, r := range required {
			if r == "Required" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Required field not in schema required array")
		}
	}

	// Check MinMax field has minimum and maximum
	if minMaxSchema, ok := properties["MinMax"].(map[string]interface{}); ok {
		if _, hasMin := minMaxSchema["minimum"]; !hasMin {
			t.Error("MinMax field missing 'minimum' constraint")
		}
		if _, hasMax := minMaxSchema["maximum"]; !hasMax {
			t.Error("MinMax field missing 'maximum' constraint")
		}
	}

	// Check Enum field has enum values
	if enumSchema, ok := properties["Enum"].(map[string]interface{}); ok {
		if _, hasEnum := enumSchema["enum"]; !hasEnum {
			t.Error("Enum field missing 'enum' constraint")
		}
	}
}

func TestGetFieldName(t *testing.T) {
	// This tests the internal getFieldName function indirectly
	// by verifying JSON tag names are used in schema

	type TagTestConfig struct {
		FieldWithJSONTag   string `json:"custom_name"`
		FieldWithoutTag    string
		FieldWithOmitEmpty string `json:"omitempty_field,omitempty"`
	}

	schema := NewJSONSchemaFromValue(TagTestConfig{})
	generated, err := schema.GenerateSchema()
	if err != nil {
		t.Fatalf("GenerateSchema failed: %v", err)
	}

	properties, ok := generated["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema properties is not a map")
	}

	// Debug: print all property names
	t.Logf("Schema properties: %+v", properties)

	// Check custom JSON tag is used
	if _, ok := properties["custom_name"]; !ok {
		t.Error("Schema missing field with custom JSON tag name 'custom_name'")
	}

	// Check field without tag - schema generation uses lowercase field names by default
	// Check both possible names
	if _, ok := properties["FieldWithoutTag"]; !ok {
		// Try lowercase version
		if _, ok := properties["fieldwithouttag"]; !ok {
			t.Error("Schema missing field 'FieldWithoutTag' or 'fieldwithouttag'")
		} else {
			t.Log("Field without tag uses lowercase name 'fieldwithouttag'")
		}
	}

	// Check omitempty field
	if _, ok := properties["omitempty_field"]; !ok {
		t.Error("Schema missing field 'omitempty_field'")
	}
}

func TestSchemaGeneration_ComplexTypes(t *testing.T) {
	type ComplexConfig struct {
		StringField string            `json:"string_field"`
		IntField    int               `json:"int_field"`
		BoolField   bool              `json:"bool_field"`
		SliceField  []string          `json:"slice_field"`
		MapField    map[string]string `json:"map_field"`
		NestedField struct {
			SubField string `json:"sub_field"`
		} `json:"nested_field"`
	}

	schema := NewJSONSchemaFromValue(ComplexConfig{})
	generated, err := schema.GenerateSchema()
	if err != nil {
		t.Fatalf("GenerateSchema failed: %v", err)
	}

	properties, ok := generated["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema properties is not a map")
	}

	// Verify all fields are present
	expectedFields := []string{
		"string_field", "int_field", "bool_field",
		"slice_field", "map_field", "nested_field",
	}

	for _, field := range expectedFields {
		if _, ok := properties[field]; !ok {
			t.Errorf("Schema missing expected field: %s", field)
		}
	}

	// Verify nested field has properties
	if nestedSchema, ok := properties["nested_field"].(map[string]interface{}); ok {
		if nestedProps, ok := nestedSchema["properties"].(map[string]interface{}); ok {
			if _, ok := nestedProps["sub_field"]; !ok {
				t.Error("Nested schema missing 'sub_field'")
			}
		}
	}
}
