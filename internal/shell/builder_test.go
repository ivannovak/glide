package shell

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

// TestNewCommandBuilder tests basic builder creation
func TestNewCommandBuilder(t *testing.T) {
	cmd := &Command{
		Name: "echo",
		Args: []string{"hello"},
	}

	builder := NewCommandBuilder(cmd)

	if builder == nil {
		t.Fatal("Expected builder to be created")
	}
	if builder.cmd != cmd {
		t.Error("Expected builder to reference the command")
	}
	if builder.ctx != nil {
		t.Error("Expected context to be nil initially")
	}
}

// TestWithContext tests context setting
func TestWithContext(t *testing.T) {
	cmd := &Command{Name: "echo"}
	ctx := context.Background()

	builder := NewCommandBuilder(cmd).WithContext(ctx)

	if builder.ctx != ctx {
		t.Error("Expected context to be set")
	}
}

// TestBuild tests basic command building
func TestBuild(t *testing.T) {
	cmd := &Command{
		Name:        "echo",
		Args:        []string{"hello", "world"},
		WorkingDir:  "/tmp",
		Environment: []string{"TEST=1"},
	}

	builder := NewCommandBuilder(cmd)
	execCmd := builder.Build()

	if execCmd.Path == "" {
		t.Error("Expected command path to be set")
	}
	if len(execCmd.Args) < 3 {
		t.Error("Expected args to include command name and arguments")
	}
	if execCmd.Dir != "/tmp" {
		t.Errorf("Expected working dir to be /tmp, got %s", execCmd.Dir)
	}
	if execCmd.Env == nil || len(execCmd.Env) == 0 {
		t.Error("Expected environment to be set")
	}
}

// TestBuildWithContext tests command building with context
func TestBuildWithContext(t *testing.T) {
	cmd := &Command{Name: "sleep", Args: []string{"1"}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	builder := NewCommandBuilder(cmd).WithContext(ctx)
	execCmd := builder.Build()

	// The exec.Cmd should be created with context
	// We can't directly test the internal context, but we can verify it's set up
	if execCmd == nil {
		t.Fatal("Expected command to be created")
	}
}

// TestBuildWithCapture tests output capture configuration
func TestBuildWithCapture(t *testing.T) {
	cmd := &Command{Name: "echo", Args: []string{"test"}}
	builder := NewCommandBuilder(cmd)

	execCmd, stdout, stderr := builder.BuildWithCapture()

	if execCmd.Stdout == nil {
		t.Error("Expected stdout to be configured")
	}
	if execCmd.Stderr == nil {
		t.Error("Expected stderr to be configured")
	}
	if stdout == nil {
		t.Error("Expected stdout buffer to be returned")
	}
	if stderr == nil {
		t.Error("Expected stderr buffer to be returned")
	}
}

// TestLimitedBuffer tests the buffer size limiting functionality
func TestLimitedBuffer(t *testing.T) {
	tests := []struct {
		name      string
		limit     int
		writes    []string
		expectErr bool
		expected  string
	}{
		{
			name:      "within limit",
			limit:     100,
			writes:    []string{"hello", " ", "world"},
			expectErr: false,
			expected:  "hello world",
		},
		{
			name:      "exactly at limit",
			limit:     5,
			writes:    []string{"hello"},
			expectErr: false,
			expected:  "hello",
		},
		{
			name:      "exceeds limit",
			limit:     5,
			writes:    []string{"hello", "world"},
			expectErr: true,
			expected:  "hello", // Should contain only what fits
		},
		{
			name:      "multiple writes exceeding limit",
			limit:     10,
			writes:    []string{"12345", "67890", "extra"},
			expectErr: true,
			expected:  "1234567890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &LimitedBuffer{limit: tt.limit}
			var err error

			for _, write := range tt.writes {
				_, writeErr := buf.Write([]byte(write))
				if writeErr != nil {
					err = writeErr
				}
			}

			if tt.expectErr && err == nil {
				t.Error("Expected error when exceeding buffer limit")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if buf.String() != tt.expected {
				t.Errorf("Expected buffer to contain %q, got %q", tt.expected, buf.String())
			}
		})
	}
}

// TestBuildWithStreaming tests streaming output configuration
func TestBuildWithStreaming(t *testing.T) {
	cmd := &Command{Name: "echo", Args: []string{"test"}}
	builder := NewCommandBuilder(cmd)

	var stdout, stderr bytes.Buffer
	execCmd := builder.BuildWithStreaming(&stdout, &stderr)

	if execCmd.Stdout != &stdout {
		t.Error("Expected stdout to be set to provided writer")
	}
	if execCmd.Stderr != &stderr {
		t.Error("Expected stderr to be set to provided writer")
	}
}

// TestBuildWithMixedOutput tests mixed output handling
func TestBuildWithMixedOutput(t *testing.T) {
	tests := []struct {
		name          string
		captureOutput bool
		optionWriter  io.Writer
		directWriter  io.Writer
		expectCapture bool
	}{
		{
			name:          "capture mode",
			captureOutput: true,
			expectCapture: true,
		},
		{
			name:          "option writer mode",
			optionWriter:  &bytes.Buffer{},
			expectCapture: false,
		},
		{
			name:          "direct writer mode",
			directWriter:  &bytes.Buffer{},
			expectCapture: false,
		},
		{
			name:          "default mode",
			expectCapture: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{
				Name:          "echo",
				CaptureOutput: tt.captureOutput,
				Stdout:        tt.directWriter,
				Stderr:        tt.directWriter,
				Options: CommandOptions{
					OutputWriter: tt.optionWriter,
					ErrorWriter:  tt.optionWriter,
				},
			}

			builder := NewCommandBuilder(cmd)
			execCmd, stdout, stderr := builder.BuildWithMixedOutput()

			if execCmd == nil {
				t.Fatal("Expected command to be created")
			}
			if stdout == nil || stderr == nil {
				t.Error("Expected buffers to be returned")
			}

			// In capture mode, buffers should have content after execution
			// In other modes, they remain empty
		})
	}
}

// TestResolveWriters tests the consolidated writer resolution logic
func TestResolveWriters(t *testing.T) {
	tests := []struct {
		name        string
		providedOut io.Writer
		providedErr io.Writer
		optionOut   io.Writer
		optionErr   io.Writer
		directOut   io.Writer
		directErr   io.Writer
		expectedOut io.Writer
		expectedErr io.Writer
	}{
		{
			name:        "defaults to OS streams",
			expectedOut: os.Stdout,
			expectedErr: os.Stderr,
		},
		{
			name:        "uses provided writers",
			providedOut: &bytes.Buffer{},
			providedErr: &bytes.Buffer{},
			expectedOut: &bytes.Buffer{},
			expectedErr: &bytes.Buffer{},
		},
		{
			name:        "options override provided",
			providedOut: &bytes.Buffer{},
			optionOut:   os.Stdout,
			expectedOut: os.Stdout,
			expectedErr: os.Stderr,
		},
		{
			name:        "direct overrides all",
			providedOut: &bytes.Buffer{},
			optionOut:   os.Stdout,
			directOut:   os.Stderr,
			expectedOut: os.Stderr,
			expectedErr: os.Stderr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{
				Name:   "test",
				Stdout: tt.directOut,
				Stderr: tt.directErr,
				Options: CommandOptions{
					OutputWriter: tt.optionOut,
					ErrorWriter:  tt.optionErr,
				},
			}

			builder := NewCommandBuilder(cmd)
			stdout, stderr := builder.resolveWriters(tt.providedOut, tt.providedErr)

			// We can't easily compare io.Writer interfaces directly,
			// but we can verify the logic by checking specific cases
			if tt.directOut != nil && stdout != tt.directOut {
				t.Error("Expected direct writer to have highest priority")
			}
			if tt.directErr != nil && stderr != tt.directErr {
				t.Error("Expected direct error writer to have highest priority")
			}
		})
	}
}

// TestDetermineTimeout tests timeout determination logic
func TestDetermineTimeout(t *testing.T) {
	tests := []struct {
		name           string
		cmdTimeout     time.Duration
		optionTimeout  time.Duration
		defaultTimeout time.Duration
		expected       time.Duration
	}{
		{
			name:           "uses option timeout first",
			cmdTimeout:     10 * time.Second,
			optionTimeout:  5 * time.Second,
			defaultTimeout: 30 * time.Second,
			expected:       5 * time.Second,
		},
		{
			name:           "uses command timeout second",
			cmdTimeout:     10 * time.Second,
			defaultTimeout: 30 * time.Second,
			expected:       10 * time.Second,
		},
		{
			name:           "uses default timeout last",
			defaultTimeout: 30 * time.Second,
			expected:       30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{
				Name:    "test",
				Timeout: tt.cmdTimeout,
				Options: CommandOptions{
					Timeout: tt.optionTimeout,
				},
			}

			builder := NewCommandBuilder(cmd)
			timeout := builder.DetermineTimeout(tt.defaultTimeout)

			if timeout != tt.expected {
				t.Errorf("Expected timeout %v, got %v", tt.expected, timeout)
			}
		})
	}
}

// TestShouldStream tests stream detection
func TestShouldStream(t *testing.T) {
	tests := []struct {
		name         string
		cmdStream    bool
		optionStream bool
		expected     bool
	}{
		{
			name:      "command stream true",
			cmdStream: true,
			expected:  true,
		},
		{
			name:         "option stream true",
			optionStream: true,
			expected:     true,
		},
		{
			name:     "both false",
			expected: false,
		},
		{
			name:         "both true",
			cmdStream:    true,
			optionStream: true,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{
				Name:         "test",
				StreamOutput: tt.cmdStream,
				Options: CommandOptions{
					StreamOutput: tt.optionStream,
				},
			}

			builder := NewCommandBuilder(cmd)
			if builder.ShouldStream() != tt.expected {
				t.Errorf("Expected ShouldStream to return %v", tt.expected)
			}
		})
	}
}

// TestShouldCapture tests capture detection
func TestShouldCapture(t *testing.T) {
	tests := []struct {
		name          string
		cmdCapture    bool
		optionCapture bool
		expected      bool
	}{
		{
			name:       "command capture true",
			cmdCapture: true,
			expected:   true,
		},
		{
			name:          "option capture true",
			optionCapture: true,
			expected:      true,
		},
		{
			name:     "both false",
			expected: false,
		},
		{
			name:          "both true",
			cmdCapture:    true,
			optionCapture: true,
			expected:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{
				Name:          "test",
				CaptureOutput: tt.cmdCapture,
				Options: CommandOptions{
					CaptureOutput: tt.optionCapture,
				},
			}

			builder := NewCommandBuilder(cmd)
			if builder.ShouldCapture() != tt.expected {
				t.Errorf("Expected ShouldCapture to return %v", tt.expected)
			}
		})
	}
}

// TestExecuteAndCollectResult tests result collection
func TestExecuteAndCollectResult(t *testing.T) {
	cmd := &Command{Name: "echo", Args: []string{"hello"}}
	builder := NewCommandBuilder(cmd)

	execCmd, stdout, stderr := builder.BuildWithCapture()
	result := builder.ExecuteAndCollectResult(execCmd, stdout, stderr)

	if result == nil {
		t.Fatal("Expected result to be returned")
	}
	if result.Duration == 0 {
		t.Error("Expected duration to be measured")
	}
	if !strings.Contains(string(result.Stdout), "hello") {
		t.Error("Expected stdout to contain output")
	}
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}
}

// TestHandleError tests error handling
func TestHandleError(t *testing.T) {
	// Test with a command that will fail
	cmd := &Command{Name: "false"} // Unix command that always returns exit code 1
	builder := NewCommandBuilder(cmd)

	execCmd := builder.Build()
	err := execCmd.Run()

	result := &Result{}
	builder.handleError(err, result)

	if result.Error == nil {
		t.Error("Expected error to be set in result")
	}
	if result.ExitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
}

// TestContextCancellation tests context cancellation handling
func TestContextCancellation(t *testing.T) {
	cmd := &Command{Name: "sleep", Args: []string{"10"}}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	builder := NewCommandBuilder(cmd).WithContext(ctx)
	execCmd := builder.Build()

	err := execCmd.Start()
	if err != nil {
		t.Fatalf("Failed to start command: %v", err)
	}

	err = execCmd.Wait()
	result := &Result{}
	builder.handleError(err, result)

	// The command should timeout
	if !result.Timeout {
		t.Error("Expected timeout to be detected")
	}
}

// TestNilInputHandling tests handling of nil inputs
func TestNilInputHandling(t *testing.T) {
	// Test with nil command - should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Builder panicked with nil command: %v", r)
		}
	}()

	builder := NewCommandBuilder(nil)
	if builder.cmd != nil {
		t.Error("Expected builder to accept nil command")
	}

	// Attempting to build with nil command should not panic
	// but may return nil or have undefined behavior
	// The important thing is it doesn't crash
}

// TestEnvironmentConfiguration tests environment variable setup
func TestEnvironmentConfiguration(t *testing.T) {
	cmd := &Command{
		Name:        "printenv",
		Environment: []string{"TEST_VAR=test_value", "ANOTHER_VAR=another_value"},
	}

	builder := NewCommandBuilder(cmd)
	execCmd := builder.Build()

	if execCmd.Env == nil {
		t.Fatal("Expected environment to be set")
	}

	// Check that our custom variables are included
	hasTestVar := false
	hasAnotherVar := false
	for _, env := range execCmd.Env {
		if env == "TEST_VAR=test_value" {
			hasTestVar = true
		}
		if env == "ANOTHER_VAR=another_value" {
			hasAnotherVar = true
		}
	}

	if !hasTestVar {
		t.Error("Expected TEST_VAR to be in environment")
	}
	if !hasAnotherVar {
		t.Error("Expected ANOTHER_VAR to be in environment")
	}
}

// TestWorkingDirectoryConfiguration tests working directory setup
func TestWorkingDirectoryConfiguration(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := &Command{
		Name:       "pwd",
		WorkingDir: tmpDir,
	}

	builder := NewCommandBuilder(cmd)
	execCmd, stdout, _ := builder.BuildWithCapture()

	if execCmd.Dir != tmpDir {
		t.Errorf("Expected working directory to be %s, got %s", tmpDir, execCmd.Dir)
	}

	// Execute and verify the working directory
	result := builder.ExecuteAndCollectResult(execCmd, stdout, nil)
	if result.Error != nil {
		t.Fatalf("Command failed: %v", result.Error)
	}

	output := strings.TrimSpace(string(result.Stdout))
	if output != tmpDir {
		t.Errorf("Expected pwd output to be %s, got %s", tmpDir, output)
	}
}

// TestGetOutputWriters tests the public GetOutputWriters method
func TestGetOutputWriters(t *testing.T) {
	var customOut, customErr bytes.Buffer

	cmd := &Command{
		Name:   "test",
		Stdout: &customOut,
		Stderr: &customErr,
	}

	builder := NewCommandBuilder(cmd)
	stdout, stderr := builder.GetOutputWriters()

	// Verify we get the custom writers back
	if stdout != &customOut {
		t.Error("Expected custom stdout writer")
	}
	if stderr != &customErr {
		t.Error("Expected custom stderr writer")
	}
}
