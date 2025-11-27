package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"sync"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   Format
	}{
		{
			name:   "creates logger with text format",
			config: &Config{Level: slog.LevelInfo, Format: FormatText, Output: &bytes.Buffer{}},
			want:   FormatText,
		},
		{
			name:   "creates logger with JSON format",
			config: &Config{Level: slog.LevelInfo, Format: FormatJSON, Output: &bytes.Buffer{}},
			want:   FormatJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.config)
			if logger == nil {
				t.Fatal("New() returned nil")
			}
			if logger.handler == nil {
				t.Error("handler is nil")
			}
			if logger.level == nil {
				t.Error("level is nil")
			}
		})
	}
}

func TestLogger_SetLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	config := &Config{
		Level:  slog.LevelInfo,
		Format: FormatText,
		Output: buf,
	}
	logger := New(config)

	// Initially at Info level, Debug should not log
	logger.Debug("debug message")
	if buf.Len() > 0 {
		t.Error("Debug message logged at Info level")
	}

	// Set to Debug level
	logger.SetLevel(slog.LevelDebug)
	logger.Debug("debug message")
	if buf.Len() == 0 {
		t.Error("Debug message not logged at Debug level")
	}
}

func TestLogger_LogLevels(t *testing.T) {
	tests := []struct {
		name     string
		logLevel slog.Level
		logFunc  func(*Logger, string)
		message  string
		wantLog  bool
	}{
		{
			name:     "debug at debug level",
			logLevel: slog.LevelDebug,
			logFunc:  func(l *Logger, msg string) { l.Debug(msg) },
			message:  "debug message",
			wantLog:  true,
		},
		{
			name:     "debug at info level",
			logLevel: slog.LevelInfo,
			logFunc:  func(l *Logger, msg string) { l.Debug(msg) },
			message:  "debug message",
			wantLog:  false,
		},
		{
			name:     "info at info level",
			logLevel: slog.LevelInfo,
			logFunc:  func(l *Logger, msg string) { l.Info(msg) },
			message:  "info message",
			wantLog:  true,
		},
		{
			name:     "warn at info level",
			logLevel: slog.LevelInfo,
			logFunc:  func(l *Logger, msg string) { l.Warn(msg) },
			message:  "warn message",
			wantLog:  true,
		},
		{
			name:     "error at info level",
			logLevel: slog.LevelInfo,
			logFunc:  func(l *Logger, msg string) { l.Error(msg) },
			message:  "error message",
			wantLog:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			config := &Config{
				Level:  tt.logLevel,
				Format: FormatText,
				Output: buf,
			}
			logger := New(config)

			tt.logFunc(logger, tt.message)

			output := buf.String()
			hasLog := len(output) > 0 && strings.Contains(output, tt.message)

			if hasLog != tt.wantLog {
				t.Errorf("wantLog = %v, but hasLog = %v, output = %q", tt.wantLog, hasLog, output)
			}
		})
	}
}

func TestLogger_With(t *testing.T) {
	buf := &bytes.Buffer{}
	config := &Config{
		Level:  slog.LevelInfo,
		Format: FormatJSON,
		Output: buf,
	}
	logger := New(config)

	// Create logger with additional fields
	childLogger := logger.With("component", "test", "version", "1.0")
	childLogger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "component") {
		t.Error("output missing 'component' field")
	}
	if !strings.Contains(output, "test") {
		t.Error("output missing 'test' value")
	}
	if !strings.Contains(output, "version") {
		t.Error("output missing 'version' field")
	}
	if !strings.Contains(output, "1.0") {
		t.Error("output missing '1.0' value")
	}
}

func TestLogger_WithGroup(t *testing.T) {
	buf := &bytes.Buffer{}
	config := &Config{
		Level:  slog.LevelInfo,
		Format: FormatJSON,
		Output: buf,
	}
	logger := New(config)

	// Create logger with group
	groupLogger := logger.WithGroup("mygroup")
	groupLogger.Info("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "mygroup") {
		t.Error("output missing group name")
	}
}

func TestLogger_Context(t *testing.T) {
	buf := &bytes.Buffer{}
	config := &Config{
		Level:  slog.LevelInfo,
		Format: FormatText,
		Output: buf,
	}
	logger := New(config)

	ctx := context.Background()
	logger.InfoContext(ctx, "test message")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Error("context logging failed")
	}
}

func TestLogger_JSONFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	config := &Config{
		Level:  slog.LevelInfo,
		Format: FormatJSON,
		Output: buf,
	}
	logger := New(config)

	logger.Info("test message", "key", "value", "number", 42)

	// Verify it's valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	// Check fields
	if result["msg"] != "test message" {
		t.Errorf("msg = %v, want 'test message'", result["msg"])
	}
	if result["key"] != "value" {
		t.Errorf("key = %v, want 'value'", result["key"])
	}
	if result["number"] != float64(42) {
		t.Errorf("number = %v, want 42", result["number"])
	}
}

func TestLogger_TextFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	config := &Config{
		Level:  slog.LevelInfo,
		Format: FormatText,
		Output: buf,
	}
	logger := New(config)

	logger.Info("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Error("output missing message")
	}
	if !strings.Contains(output, "key=value") {
		t.Error("output missing key=value")
	}
}

func TestFieldHelpers(t *testing.T) {
	buf := &bytes.Buffer{}
	config := &Config{
		Level:  slog.LevelInfo,
		Format: FormatJSON,
		Output: buf,
	}
	logger := New(config)

	logger.Info("test",
		String("str", "value"),
		Int("int", 42),
		Int64("int64", int64(100)),
		Bool("bool", true),
	)

	output := buf.String()
	if !strings.Contains(output, `"str":"value"`) {
		t.Error("String field helper failed")
	}
	if !strings.Contains(output, `"int":42`) {
		t.Error("Int field helper failed")
	}
	if !strings.Contains(output, `"int64":100`) {
		t.Error("Int64 field helper failed")
	}
	if !strings.Contains(output, `"bool":true`) {
		t.Error("Bool field helper failed")
	}
}

func TestDefault(t *testing.T) {
	// Reset default logger for testing
	defaultLogger = nil
	once = sync.Once{}

	logger := Default()
	if logger == nil {
		t.Fatal("Default() returned nil")
	}

	// Calling again should return same instance
	logger2 := Default()
	if logger != logger2 {
		t.Error("Default() returned different instances")
	}
}

func TestSetDefault(t *testing.T) {
	buf := &bytes.Buffer{}
	config := &Config{
		Level:  slog.LevelDebug,
		Format: FormatText,
		Output: buf,
	}
	customLogger := New(config)

	SetDefault(customLogger)

	// Use global functions
	Debug("test debug")
	if !strings.Contains(buf.String(), "test debug") {
		t.Error("custom default logger not used")
	}
}

func TestGlobalFunctions(t *testing.T) {
	// Reset default logger
	buf := &bytes.Buffer{}
	config := &Config{
		Level:  slog.LevelDebug,
		Format: FormatText,
		Output: buf,
	}
	SetDefault(New(config))

	tests := []struct {
		name    string
		logFunc func()
		want    string
	}{
		{
			name:    "Debug",
			logFunc: func() { Debug("debug msg") },
			want:    "debug msg",
		},
		{
			name:    "Info",
			logFunc: func() { Info("info msg") },
			want:    "info msg",
		},
		{
			name:    "Warn",
			logFunc: func() { Warn("warn msg") },
			want:    "warn msg",
		},
		{
			name:    "Error",
			logFunc: func() { Error("error msg") },
			want:    "error msg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()
			if !strings.Contains(buf.String(), tt.want) {
				t.Errorf("output missing %q", tt.want)
			}
		})
	}
}

func TestGlobalSetLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	config := &Config{
		Level:  slog.LevelInfo,
		Format: FormatText,
		Output: buf,
	}
	SetDefault(New(config))

	// Debug should not log at Info level
	Debug("test")
	if buf.Len() > 0 {
		t.Error("Debug logged at Info level")
	}

	// Set to Debug
	SetLevel(slog.LevelDebug)
	Debug("test")
	if buf.Len() == 0 {
		t.Error("Debug not logged at Debug level")
	}
}
