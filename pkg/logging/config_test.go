package logging

import (
	"log/slog"
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Level != slog.LevelWarn {
		t.Errorf("Level = %v, want %v", config.Level, slog.LevelWarn)
	}
	if config.Format != FormatText {
		t.Errorf("Format = %v, want %v", config.Format, FormatText)
	}
	if config.Output != os.Stderr {
		t.Errorf("Output = %v, want os.Stderr", config.Output)
	}
	if config.AddSource {
		t.Error("AddSource = true, want false")
	}
}

func TestFromEnv(t *testing.T) {
	tests := []struct {
		name       string
		envVars    map[string]string
		wantLevel  slog.Level
		wantFormat Format
		wantSource bool
	}{
		{
			name:       "default values",
			envVars:    map[string]string{},
			wantLevel:  slog.LevelWarn,
			wantFormat: FormatText,
			wantSource: false,
		},
		{
			name: "debug level",
			envVars: map[string]string{
				"GLIDE_LOG_LEVEL": "debug",
			},
			wantLevel:  slog.LevelDebug,
			wantFormat: FormatText,
			wantSource: false,
		},
		{
			name: "warn level",
			envVars: map[string]string{
				"GLIDE_LOG_LEVEL": "warn",
			},
			wantLevel:  slog.LevelWarn,
			wantFormat: FormatText,
			wantSource: false,
		},
		{
			name: "error level",
			envVars: map[string]string{
				"GLIDE_LOG_LEVEL": "error",
			},
			wantLevel:  slog.LevelError,
			wantFormat: FormatText,
			wantSource: false,
		},
		{
			name: "json format",
			envVars: map[string]string{
				"GLIDE_LOG_FORMAT": "json",
			},
			wantLevel:  slog.LevelWarn,
			wantFormat: FormatJSON,
			wantSource: false,
		},
		{
			name: "add source",
			envVars: map[string]string{
				"GLIDE_LOG_SOURCE": "true",
			},
			wantLevel:  slog.LevelWarn,
			wantFormat: FormatText,
			wantSource: true,
		},
		{
			name: "all options",
			envVars: map[string]string{
				"GLIDE_LOG_LEVEL":  "debug",
				"GLIDE_LOG_FORMAT": "json",
				"GLIDE_LOG_SOURCE": "1",
			},
			wantLevel:  slog.LevelDebug,
			wantFormat: FormatJSON,
			wantSource: true,
		},
		{
			name: "case insensitive",
			envVars: map[string]string{
				"GLIDE_LOG_LEVEL":  "DEBUG",
				"GLIDE_LOG_FORMAT": "JSON",
			},
			wantLevel:  slog.LevelDebug,
			wantFormat: FormatJSON,
			wantSource: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear existing env vars
			os.Unsetenv("GLIDE_LOG_LEVEL")
			os.Unsetenv("GLIDE_LOG_FORMAT")
			os.Unsetenv("GLIDE_LOG_SOURCE")

			// Set test env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer func() {
				for k := range tt.envVars {
					os.Unsetenv(k)
				}
			}()

			config := FromEnv()

			if config.Level != tt.wantLevel {
				t.Errorf("Level = %v, want %v", config.Level, tt.wantLevel)
			}
			if config.Format != tt.wantFormat {
				t.Errorf("Format = %v, want %v", config.Format, tt.wantFormat)
			}
			if config.AddSource != tt.wantSource {
				t.Errorf("AddSource = %v, want %v", config.AddSource, tt.wantSource)
			}
		})
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input string
		want  slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"INFO", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"WARN", slog.LevelWarn},
		{"error", slog.LevelError},
		{"ERROR", slog.LevelError},
		{"invalid", slog.LevelInfo}, // default
		{"", slog.LevelInfo},        // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseLevel(tt.input)
			if got != tt.want {
				t.Errorf("parseLevel(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input string
		want  Format
	}{
		{"text", FormatText},
		{"TEXT", FormatText},
		{"json", FormatJSON},
		{"JSON", FormatJSON},
		{"invalid", FormatText}, // default
		{"", FormatText},        // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseFormat(tt.input)
			if got != tt.want {
				t.Errorf("parseFormat(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseSource(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"true", true},
		{"TRUE", true},
		{"1", true},
		{"yes", true},
		{"YES", true},
		{"false", false},
		{"0", false},
		{"no", false},
		{"", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseSource(tt.input)
			if got != tt.want {
				t.Errorf("parseSource(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
