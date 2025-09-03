package output

import (
	"fmt"
	"io"
	"os"
	"sync"
)

// Manager manages output formatting
type Manager struct {
	formatter Formatter
	format    Format
	quiet     bool
	noColor   bool
	writer    io.Writer
	mu        sync.RWMutex
}

// NewManager creates a new output manager
func NewManager(format Format, quiet, noColor bool, writer io.Writer) *Manager {
	if writer == nil {
		writer = os.Stdout
	}

	m := &Manager{
		format:  format,
		quiet:   quiet,
		noColor: noColor,
		writer:  writer,
	}

	// Initialize colors based on settings
	if noColor {
		DisableColors()
	} else {
		InitColors()
	}

	// Create the appropriate formatter
	m.formatter = m.createFormatter()

	return m
}

// createFormatter creates a formatter based on the current format setting using the registry
func (m *Manager) createFormatter() Formatter {
	formatter, err := CreateFormatter(m.format, m.writer, m.noColor, m.quiet)
	if err != nil {
		// Fallback to table formatter if format not found
		return NewTableFormatter(m.writer, m.noColor, m.quiet)
	}
	return formatter
}

// SetFormat changes the output format
func (m *Manager) SetFormat(format Format) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.format = format
	m.formatter = m.createFormatter()
}

// SetQuiet enables or disables quiet mode
func (m *Manager) SetQuiet(quiet bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.quiet = quiet
	m.formatter = m.createFormatter()
}

// SetNoColor enables or disables color output
func (m *Manager) SetNoColor(noColor bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.noColor = noColor
	if noColor {
		DisableColors()
	} else {
		EnableColors()
	}
	m.formatter = m.createFormatter()
}

// SetWriter sets the output writer
func (m *Manager) SetWriter(w io.Writer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.writer = w
	m.formatter.SetWriter(w)
}

// Display outputs data using the current formatter
func (m *Manager) Display(data interface{}) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.formatter.Display(data)
}

// Info outputs an informational message
func (m *Manager) Info(format string, args ...interface{}) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.formatter.Info(format, args...)
}

// Success outputs a success message
func (m *Manager) Success(format string, args ...interface{}) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.formatter.Success(format, args...)
}

// Error outputs an error message
func (m *Manager) Error(format string, args ...interface{}) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.formatter.Error(format, args...)
}

// Warning outputs a warning message
func (m *Manager) Warning(format string, args ...interface{}) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.formatter.Warning(format, args...)
}

// Raw outputs raw text
func (m *Manager) Raw(text string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.formatter.Raw(text)
}

// Printf is a convenience method that formats and outputs text
func (m *Manager) Printf(format string, args ...interface{}) error {
	text := fmt.Sprintf(format, args...)
	return m.Raw(text)
}

// Println is a convenience method that outputs text with a newline
func (m *Manager) Println(args ...interface{}) error {
	text := fmt.Sprintln(args...)
	return m.Raw(text)
}

// IsQuiet returns whether quiet mode is enabled
func (m *Manager) IsQuiet() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.quiet
}

// GetFormat returns the current format
func (m *Manager) GetFormat() Format {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.format
}

