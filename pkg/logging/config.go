package logging

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// Format represents the output format for logs
type Format string

const (
	// FormatText outputs logs in human-readable text format
	FormatText Format = "text"
	// FormatJSON outputs logs in JSON format
	FormatJSON Format = "json"
)

// Config holds the configuration for a Logger
type Config struct {
	// Level is the minimum log level to output
	Level slog.Level
	// Format is the output format (text or json)
	Format Format
	// Output is the destination for log output
	Output io.Writer
	// AddSource adds source file and line number to log entries
	AddSource bool
}

// DefaultConfig returns a Config with sensible defaults
// Default level is Warn to keep CLI output clean - use GLIDE_LOG_LEVEL=info or debug for more verbosity
func DefaultConfig() *Config {
	return &Config{
		Level:     slog.LevelWarn,
		Format:    FormatText,
		Output:    os.Stderr,
		AddSource: false,
	}
}

// FromEnv creates a Config from environment variables
// GLIDE_LOG_LEVEL: debug, info, warn, error (default: warn)
// GLIDE_LOG_FORMAT: text, json (default: text)
// GLIDE_LOG_SOURCE: true, false (default: false)
// GLIDE_DEBUG: true, false (default: false) - shorthand for GLIDE_LOG_LEVEL=debug
func FromEnv() *Config {
	config := DefaultConfig()

	// Check for GLIDE_DEBUG first (shorthand)
	if debugStr := os.Getenv("GLIDE_DEBUG"); debugStr != "" && parseSource(debugStr) {
		config.Level = slog.LevelDebug
	}

	// Parse log level (overrides GLIDE_DEBUG if both are set)
	if levelStr := os.Getenv("GLIDE_LOG_LEVEL"); levelStr != "" {
		config.Level = parseLevel(levelStr)
	}

	// Parse log format
	if formatStr := os.Getenv("GLIDE_LOG_FORMAT"); formatStr != "" {
		config.Format = parseFormat(formatStr)
	}

	// Parse add source
	if sourceStr := os.Getenv("GLIDE_LOG_SOURCE"); sourceStr != "" {
		config.AddSource = parseSource(sourceStr)
	}

	return config
}

// parseLevel converts a string to a slog.Level
func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// parseFormat converts a string to a Format
func parseFormat(s string) Format {
	switch strings.ToLower(s) {
	case "json":
		return FormatJSON
	case "text":
		return FormatText
	default:
		return FormatText
	}
}

// parseSource converts a string to a bool
func parseSource(s string) bool {
	switch strings.ToLower(s) {
	case "true", "1", "yes":
		return true
	default:
		return false
	}
}
