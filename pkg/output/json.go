package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// JSONFormatter formats output as JSON
type JSONFormatter struct {
	*BaseFormatter
	buffer []interface{}
}

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter(w io.Writer, noColor, quiet bool) *JSONFormatter {
	if w == nil {
		w = os.Stdout
	}

	return &JSONFormatter{
		BaseFormatter: NewBaseFormatter(w, noColor, quiet),
		buffer:        make([]interface{}, 0),
	}
}

// JSONMessage represents a message in JSON output
type JSONMessage struct {
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// Display formats and outputs data as JSON
func (f *JSONFormatter) Display(data interface{}) error {
	if f.quiet {
		return nil
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return f.write(string(jsonData) + "\n")
}

// Info outputs informational messages as JSON
func (f *JSONFormatter) Info(format string, args ...interface{}) error {
	msg := &JSONMessage{
		Type:      "info",
		Message:   fmt.Sprintf(format, args...),
		Timestamp: time.Now(),
	}
	return f.Display(msg)
}

// Success outputs success messages as JSON
func (f *JSONFormatter) Success(format string, args ...interface{}) error {
	msg := &JSONMessage{
		Type:      "success",
		Message:   fmt.Sprintf(format, args...),
		Timestamp: time.Now(),
	}
	return f.Display(msg)
}

// Error outputs error messages as JSON
func (f *JSONFormatter) Error(format string, args ...interface{}) error {
	msg := &JSONMessage{
		Type:      "error",
		Message:   fmt.Sprintf(format, args...),
		Timestamp: time.Now(),
	}

	// Errors are never suppressed
	jsonData, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		return err
	}
	return f.writeError(string(jsonData) + "\n")
}

// Warning outputs warning messages as JSON
func (f *JSONFormatter) Warning(format string, args ...interface{}) error {
	msg := &JSONMessage{
		Type:      "warning",
		Message:   fmt.Sprintf(format, args...),
		Timestamp: time.Now(),
	}
	return f.Display(msg)
}

// Raw outputs raw text (not recommended for JSON formatter)
func (f *JSONFormatter) Raw(text string) error {
	// For JSON formatter, wrap raw text in a message
	msg := &JSONMessage{
		Type:      "raw",
		Message:   text,
		Timestamp: time.Now(),
	}
	return f.Display(msg)
}
