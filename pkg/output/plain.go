package output

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

// PlainFormatter formats output as plain text without colors or decorations
type PlainFormatter struct {
	*BaseFormatter
}

// NewPlainFormatter creates a new plain text formatter
func NewPlainFormatter(w io.Writer, noColor, quiet bool) *PlainFormatter {
	if w == nil {
		w = os.Stdout
	}
	
	// Plain formatter always has colors disabled (ignores noColor parameter)
	return &PlainFormatter{
		BaseFormatter: NewBaseFormatter(w, true, quiet),
	}
}

// Display formats and outputs data as plain text
func (f *PlainFormatter) Display(data interface{}) error {
	if f.quiet {
		return nil
	}

	// Handle different data types
	switch v := data.(type) {
	case string:
		return f.write(v + "\n")
	case []string:
		return f.write(strings.Join(v, "\n") + "\n")
	case map[string]interface{}:
		return f.displayMap(v)
	case []map[string]interface{}:
		return f.displaySliceOfMaps(v)
	default:
		// Use reflection for struct types
		return f.displayReflect(v)
	}
}

// displayMap displays a map as plain key-value pairs
func (f *PlainFormatter) displayMap(m map[string]interface{}) error {
	for key, value := range m {
		line := fmt.Sprintf("%s: %v\n", key, value)
		if err := f.write(line); err != nil {
			return err
		}
	}
	return nil
}

// displaySliceOfMaps displays a slice of maps
func (f *PlainFormatter) displaySliceOfMaps(data []map[string]interface{}) error {
	for i, item := range data {
		if i > 0 {
			if err := f.write("\n"); err != nil {
				return err
			}
		}
		if err := f.displayMap(item); err != nil {
			return err
		}
	}
	return nil
}

// displayReflect uses reflection to display structs
func (f *PlainFormatter) displayReflect(data interface{}) error {
	v := reflect.ValueOf(data)
	
	// Handle pointers
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		return f.displayStruct(v)
	case reflect.Slice:
		return f.displaySlice(v)
	default:
		// Fallback to simple string representation
		return f.write(fmt.Sprintf("%v\n", data))
	}
}

// displayStruct displays a struct as plain key-value pairs
func (f *PlainFormatter) displayStruct(v reflect.Value) error {
	t := v.Type()
	
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}
		
		// Get field name or json tag
		name := field.Name
		if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
			parts := strings.Split(tag, ",")
			if parts[0] != "" {
				name = parts[0]
			}
		}
		
		line := fmt.Sprintf("%s: %v\n", name, value.Interface())
		if err := f.write(line); err != nil {
			return err
		}
	}
	
	return nil
}

// displaySlice displays a slice
func (f *PlainFormatter) displaySlice(v reflect.Value) error {
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		
		// Add separator between items
		if i > 0 {
			if err := f.write("\n"); err != nil {
				return err
			}
		}
		
		// Display each element
		if err := f.displayReflect(elem.Interface()); err != nil {
			return err
		}
	}
	return nil
}

// Info outputs informational messages
func (f *PlainFormatter) Info(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	return f.write(fmt.Sprintf("[INFO] %s\n", msg))
}

// Success outputs success messages
func (f *PlainFormatter) Success(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	return f.write(fmt.Sprintf("[OK] %s\n", msg))
}

// Error outputs error messages
func (f *PlainFormatter) Error(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	return f.writeError(fmt.Sprintf("[ERROR] %s\n", msg))
}

// Warning outputs warning messages
func (f *PlainFormatter) Warning(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	return f.write(fmt.Sprintf("[WARN] %s\n", msg))
}

// Raw outputs raw text without any formatting
func (f *PlainFormatter) Raw(text string) error {
	return f.write(text)
}