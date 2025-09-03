package output

import (
	"fmt"
	"io"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// YAMLFormatter formats output as YAML
type YAMLFormatter struct {
	*BaseFormatter
}

// NewYAMLFormatter creates a new YAML formatter
func NewYAMLFormatter(w io.Writer, noColor, quiet bool) *YAMLFormatter {
	if w == nil {
		w = os.Stdout
	}

	return &YAMLFormatter{
		BaseFormatter: NewBaseFormatter(w, noColor, quiet),
	}
}

// YAMLMessage represents a message in YAML output
type YAMLMessage struct {
	Type      string    `yaml:"type"`
	Message   string    `yaml:"message"`
	Timestamp time.Time `yaml:"timestamp"`
}

// Display formats and outputs data as YAML
func (f *YAMLFormatter) Display(data interface{}) error {
	if f.quiet {
		return nil
	}

	// Marshal to YAML
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	return f.write(string(yamlData))
}

// Info outputs informational messages as YAML
func (f *YAMLFormatter) Info(format string, args ...interface{}) error {
	msg := &YAMLMessage{
		Type:      "info",
		Message:   fmt.Sprintf(format, args...),
		Timestamp: time.Now(),
	}
	return f.Display(msg)
}

// Success outputs success messages as YAML
func (f *YAMLFormatter) Success(format string, args ...interface{}) error {
	msg := &YAMLMessage{
		Type:      "success",
		Message:   fmt.Sprintf(format, args...),
		Timestamp: time.Now(),
	}
	return f.Display(msg)
}

// Error outputs error messages as YAML
func (f *YAMLFormatter) Error(format string, args ...interface{}) error {
	msg := &YAMLMessage{
		Type:      "error",
		Message:   fmt.Sprintf(format, args...),
		Timestamp: time.Now(),
	}

	// Errors are never suppressed
	yamlData, err := yaml.Marshal(msg)
	if err != nil {
		return err
	}
	return f.writeError(string(yamlData))
}

// Warning outputs warning messages as YAML
func (f *YAMLFormatter) Warning(format string, args ...interface{}) error {
	msg := &YAMLMessage{
		Type:      "warning",
		Message:   fmt.Sprintf(format, args...),
		Timestamp: time.Now(),
	}
	return f.Display(msg)
}

// Raw outputs raw text (not recommended for YAML formatter)
func (f *YAMLFormatter) Raw(text string) error {
	// For YAML formatter, wrap raw text in a message
	msg := &YAMLMessage{
		Type:      "raw",
		Message:   text,
		Timestamp: time.Now(),
	}
	return f.Display(msg)
}
