package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// Formatter defines the interface for output formatting
type Formatter interface {
	// Display formats and outputs data based on the formatter type
	Display(data interface{}) error

	// Info outputs informational messages
	Info(format string, args ...interface{}) error

	// Success outputs success messages
	Success(format string, args ...interface{}) error

	// Error outputs error messages
	Error(format string, args ...interface{}) error

	// Warning outputs warning messages
	Warning(format string, args ...interface{}) error

	// Raw outputs raw text without formatting
	Raw(text string) error

	// SetWriter sets the output writer
	SetWriter(w io.Writer)
}

// Format represents the output format type
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
	FormatPlain Format = "plain"
)

// ParseFormat converts a string to a Format type
func ParseFormat(s string) (Format, error) {
	switch s {
	case "table", "":
		return FormatTable, nil
	case "json":
		return FormatJSON, nil
	case "yaml", "yml":
		return FormatYAML, nil
	case "plain", "text":
		return FormatPlain, nil
	default:
		return "", fmt.Errorf("unknown format: %s", s)
	}
}

// BaseFormatter provides common functionality for all formatters
type BaseFormatter struct {
	writer  io.Writer
	noColor bool
	quiet   bool
}

// NewBaseFormatter creates a new base formatter
func NewBaseFormatter(w io.Writer, noColor, quiet bool) *BaseFormatter {
	return &BaseFormatter{
		writer:  w,
		noColor: noColor,
		quiet:   quiet,
	}
}

// SetWriter sets the output writer
func (f *BaseFormatter) SetWriter(w io.Writer) {
	f.writer = w
}

// write outputs text to the writer
func (f *BaseFormatter) write(text string) error {
	if f.quiet {
		return nil
	}
	_, err := fmt.Fprint(f.writer, text)
	return err
}

// writeError always writes even in quiet mode (errors are never suppressed)
func (f *BaseFormatter) writeError(text string) error {
	_, err := fmt.Fprint(f.writer, text)
	return err
}

// formatJSON converts data to JSON string
func formatJSON(data interface{}) (string, error) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
