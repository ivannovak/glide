package output

import (
	"os"
	"sync"
)

// Temporary global instance for migration
// Will be removed once all commands use dependency injection
var (
	globalManager *Manager
	globalOnce    sync.Once
)

// InitGlobal initializes the global manager for migration
func InitGlobal(format Format, quiet, noColor bool) {
	globalOnce.Do(func() {
		globalManager = NewManager(format, quiet, noColor, os.Stdout)
	})
}

// SetGlobalManager sets the global manager (for migration)
func SetGlobalManager(m *Manager) {
	globalManager = m
}

// getGlobalManager returns the global manager, creating default if needed
func getGlobalManager() *Manager {
	if globalManager == nil {
		InitGlobal(FormatTable, false, false)
	}
	return globalManager
}

// Temporary global functions for backward compatibility
// These will be removed once all commands are migrated

// Display outputs data using the global formatter
func Display(data interface{}) error {
	return getGlobalManager().Display(data)
}

// Info outputs an informational message
func Info(format string, args ...interface{}) error {
	return getGlobalManager().Info(format, args...)
}

// Success outputs a success message
func Success(format string, args ...interface{}) error {
	return getGlobalManager().Success(format, args...)
}

// Error outputs an error message
func Error(format string, args ...interface{}) error {
	return getGlobalManager().Error(format, args...)
}

// Warning outputs a warning message
func Warning(format string, args ...interface{}) error {
	return getGlobalManager().Warning(format, args...)
}

// Raw outputs raw text
func Raw(text string) error {
	return getGlobalManager().Raw(text)
}

// Printf formats and outputs text
func Printf(format string, args ...interface{}) error {
	return getGlobalManager().Printf(format, args...)
}

// Println outputs text with a newline
func Println(args ...interface{}) error {
	return getGlobalManager().Println(args...)
}

// SetFormat changes the global output format
func SetFormat(format Format) {
	getGlobalManager().SetFormat(format)
}

// SetQuiet enables or disables quiet mode globally
func SetQuiet(quiet bool) {
	getGlobalManager().SetQuiet(quiet)
}

// SetNoColor enables or disables color output globally
func SetNoColor(noColor bool) {
	getGlobalManager().SetNoColor(noColor)
}

// IsQuiet returns whether quiet mode is enabled globally
func IsQuiet() bool {
	return getGlobalManager().IsQuiet()
}