package output

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"text/tabwriter"
)

// TableFormatter formats output as human-readable tables
type TableFormatter struct {
	*BaseFormatter
	writer *tabwriter.Writer
}

// NewTableFormatter creates a new table formatter
func NewTableFormatter(w io.Writer, noColor, quiet bool) *TableFormatter {
	if w == nil {
		w = os.Stdout
	}
	
	return &TableFormatter{
		BaseFormatter: NewBaseFormatter(w, noColor, quiet),
		writer:        tabwriter.NewWriter(w, 0, 0, 2, ' ', 0),
	}
}

// SetWriter updates the output writer
func (f *TableFormatter) SetWriter(w io.Writer) {
	f.BaseFormatter.SetWriter(w)
	f.writer = tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
}

// Display formats and outputs data as a table
func (f *TableFormatter) Display(data interface{}) error {
	if f.quiet {
		return nil
	}

	// Handle different data types
	switch v := data.(type) {
	case string:
		return f.write(v + "\n")
	case []string:
		for _, line := range v {
			if err := f.write(line + "\n"); err != nil {
				return err
			}
		}
		return nil
	case map[string]interface{}:
		return f.displayMap(v)
	case []map[string]interface{}:
		return f.displayTable(v)
	default:
		// Use reflection for struct types
		return f.displayReflect(v)
	}
}

// displayMap displays a single map as key-value pairs
func (f *TableFormatter) displayMap(m map[string]interface{}) error {
	for key, value := range m {
		line := fmt.Sprintf("%s:\t%v\n", Bold("%s", key), value)
		if err := f.write(line); err != nil {
			return err
		}
	}
	return f.writer.Flush()
}

// displayTable displays a slice of maps as a table
func (f *TableFormatter) displayTable(data []map[string]interface{}) error {
	if len(data) == 0 {
		return f.write("No data to display\n")
	}

	// Extract headers from first item
	var headers []string
	for key := range data[0] {
		headers = append(headers, key)
	}

	// Write headers
	headerLine := strings.Join(headers, "\t") + "\n"
	if err := f.write(Bold("%s", headerLine)); err != nil {
		return err
	}

	// Write separator
	var separators []string
	for range headers {
		separators = append(separators, strings.Repeat("-", 10))
	}
	if err := f.write(strings.Join(separators, "\t") + "\n"); err != nil {
		return err
	}

	// Write data rows
	for _, row := range data {
		var values []string
		for _, header := range headers {
			values = append(values, fmt.Sprintf("%v", row[header]))
		}
		if err := f.write(strings.Join(values, "\t") + "\n"); err != nil {
			return err
		}
	}

	return f.writer.Flush()
}

// displayReflect uses reflection to display structs
func (f *TableFormatter) displayReflect(data interface{}) error {
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

// displayStruct displays a struct as key-value pairs
func (f *TableFormatter) displayStruct(v reflect.Value) error {
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
		
		line := fmt.Sprintf("%s:\t%v\n", Bold("%s", name), value.Interface())
		if err := f.write(line); err != nil {
			return err
		}
	}
	
	return f.writer.Flush()
}

// displaySlice displays a slice of structs as a table
func (f *TableFormatter) displaySlice(v reflect.Value) error {
	if v.Len() == 0 {
		return f.write("No data to display\n")
	}

	// Get the type of elements
	elemType := v.Type().Elem()
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}

	if elemType.Kind() != reflect.Struct {
		// Not a struct slice, display as list
		for i := 0; i < v.Len(); i++ {
			if err := f.write(fmt.Sprintf("%v\n", v.Index(i).Interface())); err != nil {
				return err
			}
		}
		return nil
	}

	// Extract headers from struct fields
	var headers []string
	var fieldIndices []int
	for i := 0; i < elemType.NumField(); i++ {
		field := elemType.Field(i)
		if !field.IsExported() {
			continue
		}
		
		name := field.Name
		if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
			parts := strings.Split(tag, ",")
			if parts[0] != "" {
				name = parts[0]
			}
		}
		headers = append(headers, name)
		fieldIndices = append(fieldIndices, i)
	}

	// Write headers
	if err := f.write(Bold("%s", strings.Join(headers, "\t") + "\n")); err != nil {
		return err
	}

	// Write separator
	var separators []string
	for range headers {
		separators = append(separators, strings.Repeat("-", 10))
	}
	if err := f.write(strings.Join(separators, "\t") + "\n"); err != nil {
		return err
	}

	// Write data rows
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
		
		var values []string
		for _, idx := range fieldIndices {
			values = append(values, fmt.Sprintf("%v", elem.Field(idx).Interface()))
		}
		if err := f.write(strings.Join(values, "\t") + "\n"); err != nil {
			return err
		}
	}

	return f.writer.Flush()
}

// Info outputs informational messages
func (f *TableFormatter) Info(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	icon := GetIcon(IconInfo)
	return f.write(fmt.Sprintf("%s %s\n", InfoText("%s", icon), msg))
}

// Success outputs success messages
func (f *TableFormatter) Success(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	icon := GetIcon(IconSuccess)
	return f.write(fmt.Sprintf("%s %s\n", SuccessText("%s", icon), msg))
}

// Error outputs error messages
func (f *TableFormatter) Error(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	icon := GetIcon(IconError)
	return f.writeError(fmt.Sprintf("%s %s\n", ErrorText("%s", icon), msg))
}

// Warning outputs warning messages
func (f *TableFormatter) Warning(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	icon := GetIcon(IconWarning)
	return f.write(fmt.Sprintf("%s %s\n", WarningText("%s", icon), msg))
}

// Raw outputs raw text without formatting
func (f *TableFormatter) Raw(text string) error {
	return f.write(text)
}