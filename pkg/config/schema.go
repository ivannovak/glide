package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// ConfigSchema defines the interface for configuration validation.
// Implementations can use JSON Schema, custom validators, or other validation mechanisms.
type ConfigSchema interface {
	// Validate checks if the given value conforms to this schema
	Validate(value interface{}) error

	// GenerateSchema generates a JSON Schema representation
	// Returns the schema as a map suitable for JSON serialization
	GenerateSchema() (map[string]interface{}, error)
}

// JSONSchema implements ConfigSchema using JSON Schema validation.
// This is a simple implementation that generates schemas from Go struct tags.
type JSONSchema struct {
	schema map[string]interface{}
}

// NewJSONSchema creates a JSON Schema from a Go type using reflection.
// It examines struct tags (json, yaml) to build the schema.
//
// Supported tags:
//   - json/yaml: field name and omitempty
//   - validate: validation rules (required, min, max, etc.)
//
// Example:
//
//	type Config struct {
//	    Name    string `json:"name" validate:"required"`
//	    Timeout int    `json:"timeout" validate:"min=1,max=3600"`
//	    Enabled bool   `json:"enabled"`
//	}
func NewJSONSchema(typ reflect.Type) *JSONSchema {
	schema := generateJSONSchema(typ)
	return &JSONSchema{
		schema: schema,
	}
}

// NewJSONSchemaFromValue creates a JSON Schema from a value using reflection.
func NewJSONSchemaFromValue(value interface{}) *JSONSchema {
	typ := reflect.TypeOf(value)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return NewJSONSchema(typ)
}

// Validate validates a value against this JSON Schema.
// This is a basic implementation - for production use, consider using a full JSON Schema validator library.
func (js *JSONSchema) Validate(value interface{}) error {
	// For now, do basic type checking
	// TODO: Implement full JSON Schema validation using a library like github.com/xeipuuv/gojsonschema

	// Convert value to JSON and back to validate it's serializable
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("value is not JSON-serializable: %w", err)
	}

	// Verify it can be unmarshaled
	var temp interface{}
	if err := json.Unmarshal(jsonData, &temp); err != nil {
		return fmt.Errorf("value failed JSON round-trip: %w", err)
	}

	// Basic validation: check required fields
	if required, ok := js.schema["required"].([]string); ok {
		valueMap, ok := temp.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected object type for validation")
		}

		for _, fieldName := range required {
			if _, exists := valueMap[fieldName]; !exists {
				return fmt.Errorf("required field missing: %s", fieldName)
			}
		}
	}

	return nil
}

// GenerateSchema returns the JSON Schema as a map.
func (js *JSONSchema) GenerateSchema() (map[string]interface{}, error) {
	return js.schema, nil
}

// generateJSONSchema recursively generates a JSON Schema from a Go type.
func generateJSONSchema(typ reflect.Type) map[string]interface{} {
	// Handle pointer types
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	schema := make(map[string]interface{})
	schema["$schema"] = "http://json-schema.org/draft-07/schema#"

	switch typ.Kind() {
	case reflect.Struct:
		schema["type"] = "object"
		properties := make(map[string]interface{})
		var required []string

		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)

			// Skip unexported fields
			if !field.IsExported() {
				continue
			}

			// Get field name from json or yaml tag
			fieldName := getFieldName(field)
			if fieldName == "-" {
				continue // Skip fields marked with "-"
			}

			// Generate schema for this field
			fieldSchema := generateFieldSchema(field)
			properties[fieldName] = fieldSchema

			// Check if field is required
			if isRequired(field) {
				required = append(required, fieldName)
			}
		}

		schema["properties"] = properties
		if len(required) > 0 {
			schema["required"] = required
		}

	case reflect.Map:
		schema["type"] = "object"
		// For maps, we allow additional properties
		schema["additionalProperties"] = generateJSONSchema(typ.Elem())

	case reflect.Slice, reflect.Array:
		schema["type"] = "array"
		schema["items"] = generateJSONSchema(typ.Elem())

	case reflect.String:
		schema["type"] = "string"

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema["type"] = "integer"

	case reflect.Float32, reflect.Float64:
		schema["type"] = "number"

	case reflect.Bool:
		schema["type"] = "boolean"

	default:
		// Unknown type, allow anything
		schema["type"] = "object"
	}

	return schema
}

// generateFieldSchema generates a JSON Schema for a struct field.
func generateFieldSchema(field reflect.StructField) map[string]interface{} {
	schema := generateJSONSchema(field.Type)

	// Parse validate tag for additional constraints
	if validateTag := field.Tag.Get("validate"); validateTag != "" {
		parseValidationRules(validateTag, schema)
	}

	// Add description from comment if available
	// Note: This would require parsing Go doc comments, which is complex
	// For now, we can add a description tag if needed
	if desc := field.Tag.Get("description"); desc != "" {
		schema["description"] = desc
	}

	return schema
}

// parseValidationRules parses validation rules from struct tags and adds them to the schema.
func parseValidationRules(validateTag string, schema map[string]interface{}) {
	rules := strings.Split(validateTag, ",")
	for _, rule := range rules {
		rule = strings.TrimSpace(rule)

		switch {
		case rule == "required":
			// Required is handled at the parent level
			continue

		case strings.HasPrefix(rule, "min="):
			val := strings.TrimPrefix(rule, "min=")
			if schema["type"] == "integer" || schema["type"] == "number" {
				schema["minimum"] = parseNumber(val)
			} else if schema["type"] == "string" {
				schema["minLength"] = parseNumber(val)
			}

		case strings.HasPrefix(rule, "max="):
			val := strings.TrimPrefix(rule, "max=")
			if schema["type"] == "integer" || schema["type"] == "number" {
				schema["maximum"] = parseNumber(val)
			} else if schema["type"] == "string" {
				schema["maxLength"] = parseNumber(val)
			}

		case strings.HasPrefix(rule, "enum="):
			val := strings.TrimPrefix(rule, "enum=")
			enumValues := strings.Split(val, "|")
			schema["enum"] = enumValues

		case strings.HasPrefix(rule, "pattern="):
			val := strings.TrimPrefix(rule, "pattern=")
			schema["pattern"] = val
		}
	}
}

// parseNumber attempts to parse a string as an int or float.
func parseNumber(s string) interface{} {
	// Try int first
	var i int
	if _, err := fmt.Sscanf(s, "%d", &i); err == nil {
		return i
	}

	// Try float
	var f float64
	if _, err := fmt.Sscanf(s, "%f", &f); err == nil {
		return f
	}

	return 0
}

// getFieldName extracts the field name from json or yaml tags.
func getFieldName(field reflect.StructField) string {
	// Try json tag first
	if jsonTag := field.Tag.Get("json"); jsonTag != "" {
		parts := strings.Split(jsonTag, ",")
		if parts[0] != "" {
			return parts[0]
		}
	}

	// Try yaml tag
	if yamlTag := field.Tag.Get("yaml"); yamlTag != "" {
		parts := strings.Split(yamlTag, ",")
		if parts[0] != "" {
			return parts[0]
		}
	}

	// Fall back to field name in lowercase
	return strings.ToLower(field.Name)
}

// isRequired checks if a field is marked as required in its validate tag.
func isRequired(field reflect.StructField) bool {
	validateTag := field.Tag.Get("validate")
	if validateTag == "" {
		return false
	}

	rules := strings.Split(validateTag, ",")
	for _, rule := range rules {
		if strings.TrimSpace(rule) == "required" {
			return true
		}
	}

	return false
}

// GenerateSchemaFromType is a convenience function to generate a JSON Schema from a Go type.
//
// Example:
//
//	type MyConfig struct {
//	    Name string `json:"name" validate:"required"`
//	}
//	schema, err := config.GenerateSchemaFromType(reflect.TypeOf(MyConfig{}))
func GenerateSchemaFromType(typ reflect.Type) (map[string]interface{}, error) {
	schema := generateJSONSchema(typ)
	return schema, nil
}

// GenerateSchemaFromValue is a convenience function to generate a JSON Schema from a value.
func GenerateSchemaFromValue(value interface{}) (map[string]interface{}, error) {
	typ := reflect.TypeOf(value)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return GenerateSchemaFromType(typ)
}
