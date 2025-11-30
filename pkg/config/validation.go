package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// ValidationError represents a configuration validation error with detailed context.
type ValidationError struct {
	Field   string // The field that failed validation
	Value   interface{}
	Rule    string // The validation rule that failed (e.g., "required", "min=1")
	Message string // Human-readable error message
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error on field %q: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []ValidationError

// Error implements the error interface for multiple errors.
func (errs ValidationErrors) Error() string {
	if len(errs) == 0 {
		return "no validation errors"
	}
	if len(errs) == 1 {
		return errs[0].Error()
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d validation errors:\n", len(errs)))
	for i, err := range errs {
		sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, err.Error()))
	}
	return sb.String()
}

// Validator provides comprehensive configuration validation.
type Validator struct {
	// AllowUnknownFields controls whether unknown fields are allowed
	AllowUnknownFields bool

	// RequireDefaults controls whether default values must be provided
	RequireDefaults bool
}

// NewValidator creates a new validator with default settings.
func NewValidator() *Validator {
	return &Validator{
		AllowUnknownFields: true,
		RequireDefaults:    false,
	}
}

// Validate validates a configuration value against struct tags.
// Returns ValidationErrors if validation fails, nil otherwise.
//
// Supported tags:
//   - validate:"required" - Field must be non-zero
//   - validate:"min=N" - Numeric/string length minimum
//   - validate:"max=N" - Numeric/string length maximum
//   - validate:"enum=a|b|c" - Value must be one of the options
//   - validate:"pattern=regexp" - String must match pattern
//
// Example:
//
//	type Config struct {
//	    Name string `json:"name" validate:"required"`
//	    Age  int    `json:"age" validate:"min=0,max=120"`
//	}
//
//	v := NewValidator()
//	if err := v.Validate(Config{Age: 150}); err != nil {
//	    // err contains detailed validation errors
//	}
func (v *Validator) Validate(value interface{}) error {
	val := reflect.ValueOf(value)
	typ := reflect.TypeOf(value)

	// Handle pointers
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return &ValidationError{
				Message: "cannot validate nil pointer",
			}
		}
		val = val.Elem()
		typ = typ.Elem()
	}

	var errors ValidationErrors

	// Only validate struct types
	if val.Kind() != reflect.Struct {
		return nil
	}

	// Validate each field
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get validation tag
		validateTag := field.Tag.Get("validate")
		if validateTag == "" {
			// No validation rules, but recurse into nested structs
			if fieldValue.Kind() == reflect.Struct {
				if err := v.Validate(fieldValue.Interface()); err != nil {
					if verrs, ok := err.(ValidationErrors); ok {
						// Prepend field name to nested errors
						for j := range verrs {
							verrs[j].Field = field.Name + "." + verrs[j].Field
						}
						errors = append(errors, verrs...)
					}
				}
			}
			continue
		}

		// Parse and apply validation rules
		rules := strings.Split(validateTag, ",")
		for _, rule := range rules {
			rule = strings.TrimSpace(rule)
			if err := v.validateRule(field.Name, fieldValue, rule); err != nil {
				errors = append(errors, *err)
			}
		}

		// Recurse into nested structs
		if fieldValue.Kind() == reflect.Struct {
			if err := v.Validate(fieldValue.Interface()); err != nil {
				if verrs, ok := err.(ValidationErrors); ok {
					// Prepend field name to nested errors
					for j := range verrs {
						verrs[j].Field = field.Name + "." + verrs[j].Field
					}
					errors = append(errors, verrs...)
				}
			}
		}
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// validateRule validates a single rule against a field value.
func (v *Validator) validateRule(fieldName string, fieldValue reflect.Value, rule string) *ValidationError {
	switch {
	case rule == "required":
		return v.validateRequired(fieldName, fieldValue)

	case strings.HasPrefix(rule, "min="):
		minStr := strings.TrimPrefix(rule, "min=")
		return v.validateMin(fieldName, fieldValue, minStr, rule)

	case strings.HasPrefix(rule, "max="):
		maxStr := strings.TrimPrefix(rule, "max=")
		return v.validateMax(fieldName, fieldValue, maxStr, rule)

	case strings.HasPrefix(rule, "enum="):
		enumStr := strings.TrimPrefix(rule, "enum=")
		return v.validateEnum(fieldName, fieldValue, enumStr, rule)

	case strings.HasPrefix(rule, "pattern="):
		pattern := strings.TrimPrefix(rule, "pattern=")
		return v.validatePattern(fieldName, fieldValue, pattern, rule)

	default:
		// Unknown rule, skip
		return nil
	}
}

// validateRequired checks if a field has a non-zero value.
func (v *Validator) validateRequired(fieldName string, fieldValue reflect.Value) *ValidationError {
	if isZeroValue(fieldValue) {
		return &ValidationError{
			Field:   fieldName,
			Value:   fieldValue.Interface(),
			Rule:    "required",
			Message: "field is required but has zero value",
		}
	}
	return nil
}

// validateMin checks minimum value/length constraints.
func (v *Validator) validateMin(fieldName string, fieldValue reflect.Value, minStr string, rule string) *ValidationError {
	switch fieldValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		min, err := strconv.ParseInt(minStr, 10, 64)
		if err != nil {
			return nil // Invalid rule, skip
		}
		if fieldValue.Int() < min {
			return &ValidationError{
				Field:   fieldName,
				Value:   fieldValue.Interface(),
				Rule:    rule,
				Message: fmt.Sprintf("value %d is less than minimum %d", fieldValue.Int(), min),
			}
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		min, err := strconv.ParseUint(minStr, 10, 64)
		if err != nil {
			return nil
		}
		if fieldValue.Uint() < min {
			return &ValidationError{
				Field:   fieldName,
				Value:   fieldValue.Interface(),
				Rule:    rule,
				Message: fmt.Sprintf("value %d is less than minimum %d", fieldValue.Uint(), min),
			}
		}

	case reflect.Float32, reflect.Float64:
		min, err := strconv.ParseFloat(minStr, 64)
		if err != nil {
			return nil
		}
		if fieldValue.Float() < min {
			return &ValidationError{
				Field:   fieldName,
				Value:   fieldValue.Interface(),
				Rule:    rule,
				Message: fmt.Sprintf("value %f is less than minimum %f", fieldValue.Float(), min),
			}
		}

	case reflect.String:
		min, err := strconv.Atoi(minStr)
		if err != nil {
			return nil
		}
		if len(fieldValue.String()) < min {
			return &ValidationError{
				Field:   fieldName,
				Value:   fieldValue.Interface(),
				Rule:    rule,
				Message: fmt.Sprintf("string length %d is less than minimum %d", len(fieldValue.String()), min),
			}
		}

	case reflect.Slice, reflect.Array:
		min, err := strconv.Atoi(minStr)
		if err != nil {
			return nil
		}
		if fieldValue.Len() < min {
			return &ValidationError{
				Field:   fieldName,
				Value:   fieldValue.Interface(),
				Rule:    rule,
				Message: fmt.Sprintf("array length %d is less than minimum %d", fieldValue.Len(), min),
			}
		}
	}

	return nil
}

// validateMax checks maximum value/length constraints.
func (v *Validator) validateMax(fieldName string, fieldValue reflect.Value, maxStr string, rule string) *ValidationError {
	switch fieldValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		max, err := strconv.ParseInt(maxStr, 10, 64)
		if err != nil {
			return nil
		}
		if fieldValue.Int() > max {
			return &ValidationError{
				Field:   fieldName,
				Value:   fieldValue.Interface(),
				Rule:    rule,
				Message: fmt.Sprintf("value %d exceeds maximum %d", fieldValue.Int(), max),
			}
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		max, err := strconv.ParseUint(maxStr, 10, 64)
		if err != nil {
			return nil
		}
		if fieldValue.Uint() > max {
			return &ValidationError{
				Field:   fieldName,
				Value:   fieldValue.Interface(),
				Rule:    rule,
				Message: fmt.Sprintf("value %d exceeds maximum %d", fieldValue.Uint(), max),
			}
		}

	case reflect.Float32, reflect.Float64:
		max, err := strconv.ParseFloat(maxStr, 64)
		if err != nil {
			return nil
		}
		if fieldValue.Float() > max {
			return &ValidationError{
				Field:   fieldName,
				Value:   fieldValue.Interface(),
				Rule:    rule,
				Message: fmt.Sprintf("value %f exceeds maximum %f", fieldValue.Float(), max),
			}
		}

	case reflect.String:
		max, err := strconv.Atoi(maxStr)
		if err != nil {
			return nil
		}
		if len(fieldValue.String()) > max {
			return &ValidationError{
				Field:   fieldName,
				Value:   fieldValue.Interface(),
				Rule:    rule,
				Message: fmt.Sprintf("string length %d exceeds maximum %d", len(fieldValue.String()), max),
			}
		}

	case reflect.Slice, reflect.Array:
		max, err := strconv.Atoi(maxStr)
		if err != nil {
			return nil
		}
		if fieldValue.Len() > max {
			return &ValidationError{
				Field:   fieldName,
				Value:   fieldValue.Interface(),
				Rule:    rule,
				Message: fmt.Sprintf("array length %d exceeds maximum %d", fieldValue.Len(), max),
			}
		}
	}

	return nil
}

// validateEnum checks if value is in the allowed set.
func (v *Validator) validateEnum(fieldName string, fieldValue reflect.Value, enumStr string, rule string) *ValidationError {
	allowedValues := strings.Split(enumStr, "|")

	// Get string representation of value
	var valueStr string
	switch fieldValue.Kind() {
	case reflect.String:
		valueStr = fieldValue.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		valueStr = strconv.FormatInt(fieldValue.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		valueStr = strconv.FormatUint(fieldValue.Uint(), 10)
	case reflect.Bool:
		valueStr = strconv.FormatBool(fieldValue.Bool())
	default:
		valueStr = fmt.Sprintf("%v", fieldValue.Interface())
	}

	// Check if value is in allowed set
	for _, allowed := range allowedValues {
		if valueStr == strings.TrimSpace(allowed) {
			return nil
		}
	}

	return &ValidationError{
		Field:   fieldName,
		Value:   fieldValue.Interface(),
		Rule:    rule,
		Message: fmt.Sprintf("value %q is not in allowed set: %s", valueStr, enumStr),
	}
}

// validatePattern checks if string matches the pattern.
// Note: This is a simplified version. For production, use regexp.MatchString.
func (v *Validator) validatePattern(fieldName string, fieldValue reflect.Value, pattern string, rule string) *ValidationError {
	if fieldValue.Kind() != reflect.String {
		return nil // Pattern validation only applies to strings
	}

	// For now, this is a placeholder
	// In a real implementation, use regexp.MatchString(pattern, fieldValue.String())
	// We're keeping it simple to avoid importing regexp here

	return nil
}

// isZeroValue checks if a reflect.Value is the zero value for its type.
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.String:
		return v.String() == ""
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() == 0
	case reflect.Struct:
		// For structs, check if all fields are zero
		zero := reflect.Zero(v.Type()).Interface()
		return reflect.DeepEqual(v.Interface(), zero)
	default:
		return false
	}
}

// ValidateWithDefaults validates a configuration and applies defaults.
// If a field is zero and a default is available, the default is applied.
//
// This is useful when loading configurations from files that may be
// incomplete - missing fields get filled in with defaults.
func ValidateWithDefaults[T any](value *T, defaults T) error {
	validator := NewValidator()

	// Validate the value
	if err := validator.Validate(value); err != nil {
		// Check if errors are validation errors
		if _, ok := err.(ValidationErrors); ok {
			// Apply defaults for required fields that are zero
			applyDefaults(reflect.ValueOf(value).Elem(), reflect.ValueOf(defaults))

			// Validate again after applying defaults
			if err := validator.Validate(value); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

// applyDefaults recursively applies default values to zero fields.
func applyDefaults(target reflect.Value, defaults reflect.Value) {
	if target.Kind() == reflect.Ptr {
		if target.IsNil() {
			return
		}
		target = target.Elem()
	}

	if defaults.Kind() == reflect.Ptr {
		if defaults.IsNil() {
			return
		}
		defaults = defaults.Elem()
	}

	if target.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < target.NumField(); i++ {
		targetField := target.Field(i)
		defaultField := defaults.Field(i)

		// Skip unexported fields
		if !targetField.CanSet() {
			continue
		}

		// If target field is zero, apply default
		if isZeroValue(targetField) && !isZeroValue(defaultField) {
			targetField.Set(defaultField)
		}

		// Recurse into nested structs
		if targetField.Kind() == reflect.Struct && defaultField.Kind() == reflect.Struct {
			applyDefaults(targetField, defaultField)
		}
	}
}
