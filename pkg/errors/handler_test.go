package errors

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultHandler(t *testing.T) {
	handler := DefaultHandler()

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.Writer)
	assert.False(t, handler.Verbose)
	assert.False(t, handler.NoColor)
	assert.False(t, handler.ShowContext)
}

func TestHandler_HandleNil(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := &Handler{Writer: buf}

	exitCode := handler.Handle(nil)

	assert.Equal(t, 0, exitCode)
	assert.Empty(t, buf.String())
}

func TestHandler_HandleGenericError(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := &Handler{
		Writer:  buf,
		NoColor: true,
	}

	err := fmt.Errorf("something went wrong")
	exitCode := handler.Handle(err)

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, buf.String(), "Error")
	assert.Contains(t, buf.String(), "something went wrong")
}

func TestHandler_HandleGlideError(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := &Handler{
		Writer:  buf,
		NoColor: true,
	}

	err := NewDockerError("docker daemon not running")
	exitCode := handler.Handle(err)

	assert.Equal(t, 125, exitCode)
	assert.Contains(t, buf.String(), "Docker Error")
	assert.Contains(t, buf.String(), "docker daemon not running")
}

func TestHandler_HandleWithSuggestions(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := &Handler{
		Writer:  buf,
		NoColor: true,
	}

	err := NewPermissionError("/tmp/test", "access denied")
	exitCode := handler.Handle(err)

	assert.Equal(t, 126, exitCode)
	output := buf.String()
	assert.Contains(t, output, "Permission Error")
	assert.Contains(t, output, "access denied")
	assert.Contains(t, output, "Possible solutions:")
	assert.Contains(t, output, "chmod 755")
}

func TestHandler_HandleVerboseMode(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := &Handler{
		Writer:  buf,
		NoColor: true,
		Verbose: true,
	}

	underlying := fmt.Errorf("underlying error")
	err := New(TypeNetwork, "connection failed", WithError(underlying))

	handler.Handle(err)

	output := buf.String()
	assert.Contains(t, output, "connection failed")
	assert.Contains(t, output, "Underlying error")
	assert.Contains(t, output, "underlying error")
}

func TestHandler_HandleVerboseWithContext(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := &Handler{
		Writer:  buf,
		NoColor: true,
		Verbose: true,
	}

	err := NewContainerError("mysql", "container failed")
	handler.Handle(err)

	output := buf.String()
	assert.Contains(t, output, "Container Error")
	assert.Contains(t, output, "container failed")
	assert.Contains(t, output, "Context:")
	assert.Contains(t, output, "container:")
	assert.Contains(t, output, "mysql")
}

func TestHandler_HandleNoContextWhenNotVerbose(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := &Handler{
		Writer:  buf,
		NoColor: true,
		Verbose: false,
	}

	err := NewContainerError("mysql", "container failed")
	handler.Handle(err)

	output := buf.String()
	assert.Contains(t, output, "container failed")
	assert.NotContains(t, output, "Context:")
}

func TestHandler_GetErrorIcon(t *testing.T) {
	handler := DefaultHandler()

	tests := []struct {
		name     string
		errType  ErrorType
		expected string
	}{
		{"docker", TypeDocker, "🐳"},
		{"container", TypeContainer, "🐳"},
		{"permission", TypePermission, "🔒"},
		{"file not found", TypeFileNotFound, "📁"},
		{"dependency", TypeDependency, "📦"},
		{"missing", TypeMissing, "📦"},
		{"config", TypeConfig, "⚙️"},
		{"network", TypeNetwork, "🌐"},
		{"connection", TypeConnection, "🌐"},
		{"database", TypeDatabase, "🗄️"},
		{"mode", TypeMode, "🔄"},
		{"wrong mode", TypeWrongMode, "🔄"},
		{"timeout", TypeTimeout, "⏱️"},
		{"command", TypeCommand, "💻"},
		{"unknown", TypeUnknown, "✗"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			icon := handler.getErrorIcon(tt.errType)
			assert.Equal(t, tt.expected, icon)
		})
	}
}

func TestHandler_GetErrorTypeString(t *testing.T) {
	handler := DefaultHandler()

	tests := []struct {
		errType  ErrorType
		expected string
	}{
		{TypeDocker, "Docker Error"},
		{TypeContainer, "Container Error"},
		{TypePermission, "Permission Error"},
		{TypeFileNotFound, "File Not Found"},
		{TypeDependency, "Dependency Error"},
		{TypeMissing, "Missing Resource"},
		{TypeConfig, "Configuration Error"},
		{TypeNetwork, "Network Error"},
		{TypeConnection, "Connection Error"},
		{TypeDatabase, "Database Error"},
		{TypeMode, "Mode Error"},
		{TypeWrongMode, "Wrong Mode"},
		{TypeTimeout, "Timeout"},
		{TypeCommand, "Command Error"},
		{TypeUnknown, "Error"},
	}

	for _, tt := range tests {
		t.Run(string(tt.errType), func(t *testing.T) {
			result := handler.getErrorTypeString(tt.errType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHandler_DisplaySuggestionsEmpty(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := &Handler{
		Writer:  buf,
		NoColor: true,
	}

	handler.displaySuggestions([]string{})
	assert.Empty(t, buf.String())
}

func TestHandler_DisplaySuggestions(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := &Handler{
		Writer:  buf,
		NoColor: true,
	}

	suggestions := []string{
		"Check the logs",
		"Restart the service",
		"Run: docker ps",
	}

	handler.displaySuggestions(suggestions)

	output := buf.String()
	assert.Contains(t, output, "Possible solutions:")
	assert.Contains(t, output, "Check the logs")
	assert.Contains(t, output, "Restart the service")
	assert.Contains(t, output, "docker ps")
}

func TestHandler_DisplaySuggestionsWithCommandPrefixes(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := &Handler{
		Writer:  buf,
		NoColor: true,
	}

	suggestions := []string{
		"Run: docker ps",
		"Check: docker logs",
		"Fix: chmod 755 /path",
	}

	handler.displaySuggestions(suggestions)

	output := buf.String()
	assert.Contains(t, output, "Run:")
	assert.Contains(t, output, "Check:")
	assert.Contains(t, output, "Fix:")
}

func TestHandler_DisplayContext(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := &Handler{
		Writer:  buf,
		NoColor: true,
	}

	context := map[string]string{
		"container": "mysql",
		"command":   "docker ps",
		"path":      "/tmp/test",
	}

	handler.displayContext(context)

	output := buf.String()
	assert.Contains(t, output, "Context:")
	assert.Contains(t, output, "container:")
	assert.Contains(t, output, "mysql")
	assert.Contains(t, output, "command:")
	assert.Contains(t, output, "docker ps")
	assert.Contains(t, output, "path:")
	assert.Contains(t, output, "/tmp/test")
}

func TestHandler_WithColor(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := &Handler{
		Writer:  buf,
		NoColor: false, // Enable color
	}

	err := NewDockerError("test error")
	handler.Handle(err)

	// Output will contain ANSI color codes when color is enabled
	// We just verify it doesn't crash and produces output
	assert.NotEmpty(t, buf.String())
}

func TestPrint(t *testing.T) {
	// Print uses DefaultHandler which writes to stderr
	// We can't easily capture stderr, so just verify it doesn't crash
	exitCode := Print(nil)
	assert.Equal(t, 0, exitCode)

	err := NewDockerError("test error")
	exitCode = Print(err)
	assert.Equal(t, 125, exitCode)
}

func TestPrintVerbose(t *testing.T) {
	exitCode := PrintVerbose(nil)
	assert.Equal(t, 0, exitCode)

	underlying := fmt.Errorf("underlying")
	err := New(TypeNetwork, "test", WithError(underlying))
	exitCode = PrintVerbose(err)
	assert.Equal(t, 1, exitCode)
}

func TestHandler_ExitCodes(t *testing.T) {
	tests := []struct {
		name         string
		err          *GlideError
		expectedCode int
	}{
		{
			name:         "docker error",
			err:          NewDockerError("test"),
			expectedCode: 125,
		},
		{
			name:         "permission error",
			err:          NewPermissionError("/tmp", "test"),
			expectedCode: 126,
		},
		{
			name:         "file not found",
			err:          NewFileNotFoundError("/tmp/file"),
			expectedCode: 127,
		},
		{
			name:         "config error",
			err:          NewConfigError("test"),
			expectedCode: 78,
		},
		{
			name:         "network error",
			err:          NewNetworkError("test"),
			expectedCode: 69,
		},
		{
			name:         "timeout error",
			err:          NewTimeoutError("operation"),
			expectedCode: 124,
		},
		{
			name:         "custom exit code",
			err:          New(TypeUnknown, "test", WithExitCode(99)),
			expectedCode: 99,
		},
		{
			name:         "default exit code",
			err:          New(TypeUnknown, "test"),
			expectedCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			handler := &Handler{Writer: buf, NoColor: true}

			exitCode := handler.Handle(tt.err)
			assert.Equal(t, tt.expectedCode, exitCode)
		})
	}
}

func TestHandler_DisplayGenericError(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := &Handler{
		Writer:  buf,
		NoColor: true,
	}

	err := fmt.Errorf("generic error message")
	handler.displayGenericError(err)

	output := buf.String()
	assert.Contains(t, output, "✗")
	assert.Contains(t, output, "Error")
	assert.Contains(t, output, "generic error message")
}

func TestHandler_ComplexErrorChain(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := &Handler{
		Writer:  buf,
		NoColor: true,
		Verbose: true,
	}

	// Create a chain: underlying -> wrapped -> handled
	underlying := fmt.Errorf("root cause")
	wrapped := NewDatabaseError("database query failed", WithError(underlying))
	wrapped.AddSuggestion("Check database connection")
	wrapped.AddContext("query", "SELECT * FROM users")

	exitCode := handler.Handle(wrapped)

	output := buf.String()
	assert.Equal(t, 69, exitCode)
	assert.Contains(t, output, "Database Error")
	assert.Contains(t, output, "database query failed")
	assert.Contains(t, output, "Underlying error")
	assert.Contains(t, output, "root cause")
	assert.Contains(t, output, "Possible solutions:")
	assert.Contains(t, output, "Check database connection")
	assert.Contains(t, output, "Context:")
	assert.Contains(t, output, "query:")
	assert.Contains(t, output, "SELECT * FROM users")
}

func TestHandler_MultipleSuggestionsFormatting(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := &Handler{
		Writer:  buf,
		NoColor: true,
	}

	err := NewDockerError("test")
	err.AddSuggestion("Run: docker ps")
	err.AddSuggestion("Check: docker version")
	err.AddSuggestion("Fix: restart docker")
	err.AddSuggestion("Regular suggestion without prefix")

	handler.Handle(err)

	output := buf.String()
	lines := strings.Split(output, "\n")

	// Count bullet points
	bulletCount := 0
	for _, line := range lines {
		if strings.Contains(line, "•") {
			bulletCount++
		}
	}

	assert.Equal(t, 4, bulletCount, "Should have 4 bullet points for 4 suggestions")
}
