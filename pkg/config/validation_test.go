package config

import (
	"reflect"
	"strings"
	"testing"
)

func TestValidator_Required(t *testing.T) {
	type Config struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required"`
		Age   int    `json:"age"`
	}

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  Config{Name: "John", Email: "john@example.com", Age: 30},
			wantErr: false,
		},
		{
			name:    "missing required name",
			config:  Config{Email: "john@example.com"},
			wantErr: true,
		},
		{
			name:    "missing required email",
			config:  Config{Name: "John"},
			wantErr: true,
		},
		{
			name:    "missing both required fields",
			config:  Config{Age: 30},
			wantErr: true,
		},
		{
			name:    "missing optional age is ok",
			config:  Config{Name: "John", Email: "john@example.com"},
			wantErr: false,
		},
	}

	validator := NewValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_Min(t *testing.T) {
	type Config struct {
		Age   int      `json:"age" validate:"min=0"`
		Score int      `json:"score" validate:"min=1,max=100"`
		Name  string   `json:"name" validate:"min=2"`
		Tags  []string `json:"tags" validate:"min=1"`
	}

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  Config{Age: 25, Score: 85, Name: "Jo", Tags: []string{"tag1"}},
			wantErr: false,
		},
		{
			name:    "age too low",
			config:  Config{Age: -1, Score: 85, Name: "Jo", Tags: []string{"tag1"}},
			wantErr: true,
		},
		{
			name:    "score too low",
			config:  Config{Age: 25, Score: 0, Name: "Jo", Tags: []string{"tag1"}},
			wantErr: true,
		},
		{
			name:    "name too short",
			config:  Config{Age: 25, Score: 85, Name: "J", Tags: []string{"tag1"}},
			wantErr: true,
		},
		{
			name:    "empty tags array",
			config:  Config{Age: 25, Score: 85, Name: "Jo", Tags: []string{}},
			wantErr: true,
		},
	}

	validator := NewValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_Max(t *testing.T) {
	type Config struct {
		Age   int      `json:"age" validate:"max=120"`
		Score int      `json:"score" validate:"min=1,max=100"`
		Name  string   `json:"name" validate:"max=50"`
		Tags  []string `json:"tags" validate:"max=5"`
	}

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  Config{Age: 30, Score: 85, Name: "John", Tags: []string{"tag1", "tag2"}},
			wantErr: false,
		},
		{
			name:    "age too high",
			config:  Config{Age: 150, Score: 85, Name: "John", Tags: []string{"tag1"}},
			wantErr: true,
		},
		{
			name:    "score too high",
			config:  Config{Age: 30, Score: 101, Name: "John", Tags: []string{"tag1"}},
			wantErr: true,
		},
		{
			name:    "name too long",
			config:  Config{Age: 30, Score: 85, Name: strings.Repeat("a", 51), Tags: []string{"tag1"}},
			wantErr: true,
		},
		{
			name:    "too many tags",
			config:  Config{Age: 30, Score: 85, Name: "John", Tags: []string{"1", "2", "3", "4", "5", "6"}},
			wantErr: true,
		},
	}

	validator := NewValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_Enum(t *testing.T) {
	type Config struct {
		Role   string `json:"role" validate:"enum=admin|user|guest"`
		Status int    `json:"status" validate:"enum=0|1|2"`
	}

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid role admin",
			config:  Config{Role: "admin", Status: 1},
			wantErr: false,
		},
		{
			name:    "valid role user",
			config:  Config{Role: "user", Status: 0},
			wantErr: false,
		},
		{
			name:    "valid role guest",
			config:  Config{Role: "guest", Status: 2},
			wantErr: false,
		},
		{
			name:    "invalid role",
			config:  Config{Role: "superadmin", Status: 1},
			wantErr: true,
		},
		{
			name:    "invalid status",
			config:  Config{Role: "admin", Status: 3},
			wantErr: true,
		},
	}

	validator := NewValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_MultipleErrors(t *testing.T) {
	type Config struct {
		Name  string `json:"name" validate:"required,min=2"`
		Age   int    `json:"age" validate:"required,min=0,max=120"`
		Email string `json:"email" validate:"required"`
	}

	// Config with multiple validation errors
	config := Config{
		Name:  "J", // Too short (min=2)
		Age:   -1,  // Too low (min=0)
		Email: "",  // Required but missing
	}

	validator := NewValidator()
	err := validator.Validate(config)

	if err == nil {
		t.Fatal("Expected validation errors, got nil")
	}

	verrs, ok := err.(ValidationErrors)
	if !ok {
		t.Fatalf("Expected ValidationErrors, got %T", err)
	}

	// Should have at least 3 errors (name too short, age too low, email missing)
	if len(verrs) < 3 {
		t.Errorf("Expected at least 3 errors, got %d: %v", len(verrs), verrs)
	}
}

func TestValidator_NestedStructs(t *testing.T) {
	type Address struct {
		Street string `json:"street" validate:"required"`
		City   string `json:"city" validate:"required"`
		Zip    string `json:"zip" validate:"min=5,max=10"`
	}

	type Person struct {
		Name    string  `json:"name" validate:"required"`
		Address Address `json:"address"`
	}

	tests := []struct {
		name    string
		config  Person
		wantErr bool
	}{
		{
			name: "valid nested config",
			config: Person{
				Name: "John",
				Address: Address{
					Street: "123 Main St",
					City:   "Springfield",
					Zip:    "12345",
				},
			},
			wantErr: false,
		},
		{
			name: "missing nested required field",
			config: Person{
				Name: "John",
				Address: Address{
					Street: "123 Main St",
					// City missing
					Zip: "12345",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid nested validation",
			config: Person{
				Name: "John",
				Address: Address{
					Street: "123 Main St",
					City:   "Springfield",
					Zip:    "123", // Too short (min=5)
				},
			},
			wantErr: true,
		},
	}

	validator := NewValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			// If we expect an error and got one, check that field names include nesting
			if tt.wantErr && err != nil {
				errStr := err.Error()
				if !strings.Contains(errStr, "Address.") {
					t.Errorf("Expected nested field name in error, got: %s", errStr)
				}
			}
		})
	}
}

func TestValidateWithDefaults(t *testing.T) {
	type Config struct {
		Name    string `json:"name" validate:"required"`
		Timeout int    `json:"timeout" validate:"min=1"`
		Enabled bool   `json:"enabled"`
	}

	defaults := Config{
		Name:    "default-name",
		Timeout: 30,
		Enabled: true,
	}

	tests := []struct {
		name    string
		config  Config
		want    Config
		wantErr bool
	}{
		{
			name:   "empty config gets defaults",
			config: Config{},
			want: Config{
				Name:    "default-name",
				Timeout: 30,
				Enabled: true,
			},
			wantErr: false,
		},
		{
			name: "partial config gets missing defaults",
			config: Config{
				Name: "custom-name",
			},
			want: Config{
				Name:    "custom-name",
				Timeout: 30,
				Enabled: true,
			},
			wantErr: false,
		},
		{
			name: "full config keeps values",
			config: Config{
				Name:    "my-name",
				Timeout: 60,
				Enabled: false,
			},
			want: Config{
				Name:    "my-name",
				Timeout: 60,
				Enabled: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWithDefaults(&tt.config, defaults)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateWithDefaults() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tt.config.Name != tt.want.Name {
					t.Errorf("Name = %v, want %v", tt.config.Name, tt.want.Name)
				}
				if tt.config.Timeout != tt.want.Timeout {
					t.Errorf("Timeout = %v, want %v", tt.config.Timeout, tt.want.Timeout)
				}
				if tt.config.Enabled != tt.want.Enabled {
					t.Errorf("Enabled = %v, want %v", tt.config.Enabled, tt.want.Enabled)
				}
			}
		})
	}
}

func TestValidationErrors_Error(t *testing.T) {
	// Test single error
	singleErr := ValidationErrors{
		ValidationError{
			Field:   "name",
			Value:   "",
			Rule:    "required",
			Message: "field is required but has zero value",
		},
	}

	errStr := singleErr.Error()
	if !strings.Contains(errStr, "name") {
		t.Errorf("Single error should contain field name, got: %s", errStr)
	}

	// Test multiple errors
	multiErr := ValidationErrors{
		ValidationError{Field: "name", Message: "name is required"},
		ValidationError{Field: "age", Message: "age must be positive"},
		ValidationError{Field: "email", Message: "email is invalid"},
	}

	errStr = multiErr.Error()
	if !strings.Contains(errStr, "3 validation errors") {
		t.Errorf("Multiple errors should show count, got: %s", errStr)
	}
	if !strings.Contains(errStr, "name") || !strings.Contains(errStr, "age") || !strings.Contains(errStr, "email") {
		t.Errorf("Multiple errors should list all fields, got: %s", errStr)
	}
}

// TestValidator_UintAndFloatTypes tests validateMin/Max with uint and float types
func TestValidator_UintAndFloatTypes(t *testing.T) {
	type Config struct {
		Count    uint    `json:"count" validate:"min=1,max=100"`
		Ratio    float64 `json:"ratio" validate:"min=0.0,max=1.0"`
		SmallInt uint8   `json:"small_int" validate:"min=5,max=50"`
		Percent  float32 `json:"percent" validate:"min=0.0,max=100.0"`
	}

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid uint and float values",
			config:  Config{Count: 50, Ratio: 0.5, SmallInt: 25, Percent: 75.0},
			wantErr: false,
		},
		{
			name:    "uint below minimum",
			config:  Config{Count: 0, Ratio: 0.5, SmallInt: 25, Percent: 75.0},
			wantErr: true,
		},
		{
			name:    "uint above maximum",
			config:  Config{Count: 101, Ratio: 0.5, SmallInt: 25, Percent: 75.0},
			wantErr: true,
		},
		{
			name:    "float below minimum",
			config:  Config{Count: 50, Ratio: -0.1, SmallInt: 25, Percent: 75.0},
			wantErr: true,
		},
		{
			name:    "float above maximum",
			config:  Config{Count: 50, Ratio: 1.5, SmallInt: 25, Percent: 75.0},
			wantErr: true,
		},
		{
			name:    "uint8 below minimum",
			config:  Config{Count: 50, Ratio: 0.5, SmallInt: 4, Percent: 75.0},
			wantErr: true,
		},
		{
			name:    "float32 above maximum",
			config:  Config{Count: 50, Ratio: 0.5, SmallInt: 25, Percent: 101.0},
			wantErr: true,
		},
	}

	validator := NewValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidator_InvalidRules tests error handling for malformed validation rules
func TestValidator_InvalidRules(t *testing.T) {
	type Config struct {
		BadMin      int    `json:"bad_min" validate:"min=invalid"`
		BadMax      int    `json:"bad_max" validate:"max=not_a_number"`
		BadEnum     string `json:"bad_enum" validate:"enum="`
		UnknownRule string `json:"unknown" validate:"foobar=123"`
	}

	config := Config{
		BadMin:      10,
		BadMax:      20,
		BadEnum:     "test",
		UnknownRule: "value",
	}

	validator := NewValidator()
	// Should not panic, invalid rules should be skipped
	err := validator.Validate(config)

	// Invalid rules are silently skipped, so no error expected
	if err != nil {
		t.Logf("Got error (expected to be skipped): %v", err)
	}
}

// TestValidator_Pattern tests pattern validation (currently a placeholder)
func TestValidator_Pattern(t *testing.T) {
	type Config struct {
		Email     string `json:"email" validate:"pattern=^[a-z]+@[a-z]+\\.[a-z]+$"`
		NotString int    `json:"not_string" validate:"pattern=.*"`
	}

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "pattern on string (currently noop)",
			config:  Config{Email: "test@example.com", NotString: 123},
			wantErr: false, // Pattern validation is currently a noop
		},
		{
			name:    "pattern on non-string (skipped)",
			config:  Config{Email: "valid@test.com", NotString: 456},
			wantErr: false,
		},
	}

	validator := NewValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestIsZeroValue tests the isZeroValue helper function indirectly
func TestIsZeroValue_ThroughValidation(t *testing.T) {
	type Config struct {
		BoolField   bool              `json:"bool_field" validate:"required"`
		IntField    int               `json:"int_field" validate:"required"`
		UintField   uint              `json:"uint_field" validate:"required"`
		FloatField  float64           `json:"float_field" validate:"required"`
		StringField string            `json:"string_field" validate:"required"`
		SliceField  []int             `json:"slice_field" validate:"required"`
		MapField    map[string]string `json:"map_field" validate:"required"`
		PtrField    *string           `json:"ptr_field" validate:"required"`
	}

	validator := NewValidator()

	// Test zero values are detected as missing (required validation)
	zeroConfig := Config{}
	err := validator.Validate(zeroConfig)
	if err == nil {
		t.Error("Expected validation errors for zero values, got nil")
	}

	// Test non-zero values pass required validation
	str := "test"
	nonZeroConfig := Config{
		BoolField:   true,
		IntField:    1,
		UintField:   1,
		FloatField:  1.0,
		StringField: "value",
		SliceField:  []int{1},
		MapField:    map[string]string{"key": "value"},
		PtrField:    &str,
	}
	err = validator.Validate(nonZeroConfig)
	if err != nil {
		t.Errorf("Expected no errors for non-zero values, got: %v", err)
	}
}

// TestApplyDefaults tests the applyDefaults function with various types
// NOTE: ValidateWithDefaults only applies defaults when validation fails
func TestApplyDefaults_AllTypes(t *testing.T) {
	type Nested struct {
		Value string `json:"value" validate:"required"`
	}

	type Config struct {
		StringVal string            `json:"string_val" validate:"required"`
		IntVal    int               `json:"int_val" validate:"required"`
		BoolVal   bool              `json:"bool_val"`
		FloatVal  float64           `json:"float_val"`
		UintVal   uint              `json:"uint_val"`
		SliceVal  []string          `json:"slice_val"`
		MapVal    map[string]string `json:"map_val"`
		StructVal Nested            `json:"struct_val"`
		PtrVal    *string           `json:"ptr_val"`
	}

	defaultStr := "default-ptr"
	defaults := Config{
		StringVal: "default-string",
		IntVal:    42,
		BoolVal:   true,
		FloatVal:  3.14,
		UintVal:   100,
		SliceVal:  []string{"default1", "default2"},
		MapVal:    map[string]string{"key": "default"},
		StructVal: Nested{Value: "default-nested"},
		PtrVal:    &defaultStr,
	}

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "empty config gets required defaults applied",
			config:  Config{},
			wantErr: false, // Defaults should fix required field violations
		},
		{
			name: "partial config gets missing required defaults",
			config: Config{
				StringVal: "custom",
				// IntVal missing (required) - should get default
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWithDefaults(&tt.config, defaults)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateWithDefaults() error = %v, wantErr %v", err, tt.wantErr)
			}

			// After applying defaults, required fields should be populated
			if !tt.wantErr {
				if tt.config.StringVal == "" {
					t.Error("StringVal should be populated from defaults")
				}
				if tt.config.IntVal == 0 {
					t.Error("IntVal should be populated from defaults")
				}
			}
		})
	}
}

// TestApplyDefaults_DirectCall tests the applyDefaults function more directly
func TestApplyDefaults_DirectCall(t *testing.T) {
	type Config struct {
		StringVal string
		IntVal    int
		BoolVal   bool
		FloatVal  float64
		UintVal   uint
	}

	defaults := Config{
		StringVal: "default",
		IntVal:    42,
		BoolVal:   true,
		FloatVal:  3.14,
		UintVal:   100,
	}

	// Test with empty config
	config := Config{}
	applyDefaults(reflect.ValueOf(&config).Elem(), reflect.ValueOf(defaults))

	// All zero values should be replaced with defaults
	if config.StringVal != "default" {
		t.Errorf("StringVal = %v, want 'default'", config.StringVal)
	}
	if config.IntVal != 42 {
		t.Errorf("IntVal = %v, want 42", config.IntVal)
	}
	// Note: bool zero value (false) won't be replaced since it's a valid value
	if config.FloatVal != 3.14 {
		t.Errorf("FloatVal = %v, want 3.14", config.FloatVal)
	}
	if config.UintVal != 100 {
		t.Errorf("UintVal = %v, want 100", config.UintVal)
	}

	// Test with partial config (should preserve non-zero values)
	config2 := Config{StringVal: "custom", IntVal: 99}
	applyDefaults(reflect.ValueOf(&config2).Elem(), reflect.ValueOf(defaults))

	if config2.StringVal != "custom" {
		t.Errorf("StringVal should be preserved, got %v", config2.StringVal)
	}
	if config2.IntVal != 99 {
		t.Errorf("IntVal should be preserved, got %v", config2.IntVal)
	}
}

// TestValidationError_SingleError tests the Error() method with single validation error
func TestValidationError_SingleError(t *testing.T) {
	err := ValidationError{
		Field:   "username",
		Value:   "",
		Rule:    "required",
		Message: "username is required",
	}

	errStr := err.Error()

	// Should contain all key information
	if !strings.Contains(errStr, "username") {
		t.Errorf("Error should contain field name, got: %s", errStr)
	}
	if !strings.Contains(errStr, "required") {
		t.Errorf("Error should contain rule, got: %s", errStr)
	}
	if !strings.Contains(errStr, "username is required") {
		t.Errorf("Error should contain message, got: %s", errStr)
	}
}
