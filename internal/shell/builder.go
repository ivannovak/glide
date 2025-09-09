package shell

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

const (
	// MaxBufferSize is the maximum size for captured output buffers (10MB)
	MaxBufferSize = 10 * 1024 * 1024
)

// LimitedBuffer wraps a bytes.Buffer with a size limit
type LimitedBuffer struct {
	buffer bytes.Buffer // Don't embed to prevent type assertion bypass
	limit  int
	closed bool // Track if we've hit the limit
}

// Write implements io.Writer with size limit enforcement
func (b *LimitedBuffer) Write(p []byte) (n int, err error) {
	// If we've already hit the limit, reject all writes
	if b.closed {
		return 0, fmt.Errorf("output buffer full, limit of %d bytes reached", b.limit)
	}

	currentLen := b.buffer.Len()
	if currentLen+len(p) > b.limit {
		// Write as much as we can within the limit
		remaining := b.limit - currentLen
		if remaining > 0 {
			n, err = b.buffer.Write(p[:remaining])
			b.closed = true // Mark as closed after hitting limit
			if err == nil {
				err = fmt.Errorf("output exceeded maximum buffer size of %d bytes", b.limit)
			}
			return n, err
		}
		b.closed = true
		return 0, fmt.Errorf("output buffer full, limit of %d bytes reached", b.limit)
	}
	return b.buffer.Write(p)
}

// Bytes returns the buffer contents
func (b *LimitedBuffer) Bytes() []byte {
	return b.buffer.Bytes()
}

// String returns the buffer contents as a string
func (b *LimitedBuffer) String() string {
	return b.buffer.String()
}

// Len returns the current buffer length
func (b *LimitedBuffer) Len() int {
	return b.buffer.Len()
}

// CommandBuilder builds and configures exec.Cmd instances from Command structs
type CommandBuilder struct {
	cmd *Command
	ctx context.Context
}

// NewCommandBuilder creates a new command builder
func NewCommandBuilder(cmd *Command) *CommandBuilder {
	return &CommandBuilder{
		cmd: cmd,
	}
}

// WithContext sets the context for the command
func (b *CommandBuilder) WithContext(ctx context.Context) *CommandBuilder {
	b.ctx = ctx
	return b
}

// Build creates and configures an exec.Cmd
func (b *CommandBuilder) Build() *exec.Cmd {
	// Create the exec.Cmd with or without context
	var execCmd *exec.Cmd
	if b.ctx != nil {
		execCmd = exec.CommandContext(b.ctx, b.cmd.Name, b.cmd.Args...)
	} else {
		execCmd = exec.Command(b.cmd.Name, b.cmd.Args...)
	}

	// Set working directory if specified
	if b.cmd.WorkingDir != "" {
		execCmd.Dir = b.cmd.WorkingDir
	}

	// Set environment
	b.configureEnvironment(execCmd)

	// Configure I/O
	b.configureIO(execCmd)

	return execCmd
}

// configureEnvironment sets up the command environment
func (b *CommandBuilder) configureEnvironment(execCmd *exec.Cmd) {
	if len(b.cmd.Environment) > 0 {
		execCmd.Env = os.Environ()
		execCmd.Env = append(execCmd.Env, b.cmd.Environment...)
	}
}

// configureIO sets up standard I/O for the command
func (b *CommandBuilder) configureIO(execCmd *exec.Cmd) {
	// Configure stdin
	if b.cmd.Stdin != nil {
		execCmd.Stdin = b.cmd.Stdin
	}

	// Note: Stdout and Stderr configuration is left to strategies
	// since they have different requirements
}

// BuildWithCapture creates an exec.Cmd configured for output capture with size limits
func (b *CommandBuilder) BuildWithCapture() (*exec.Cmd, *bytes.Buffer, *bytes.Buffer) {
	execCmd := b.Build()

	// Use LimitedBuffer to prevent memory exhaustion
	stdout := &LimitedBuffer{limit: MaxBufferSize}
	stderr := &LimitedBuffer{limit: MaxBufferSize}
	execCmd.Stdout = stdout
	execCmd.Stderr = stderr

	// Return the internal buffers for reading the captured output
	// The LimitedBuffer will enforce the size limit during writes
	return execCmd, &stdout.buffer, &stderr.buffer
}

// BuildWithStreaming creates an exec.Cmd configured for streaming output
func (b *CommandBuilder) BuildWithStreaming(outputWriter, errorWriter io.Writer) *exec.Cmd {
	execCmd := b.Build()

	// Use consolidated writer resolution logic
	stdout, stderr := b.resolveWriters(outputWriter, errorWriter)

	execCmd.Stdout = stdout
	execCmd.Stderr = stderr

	return execCmd
}

// BuildWithMixedOutput creates an exec.Cmd with configurable output handling
func (b *CommandBuilder) BuildWithMixedOutput() (*exec.Cmd, *bytes.Buffer, *bytes.Buffer) {
	execCmd := b.Build()

	var stdoutBuf, stderrBuf *bytes.Buffer
	shouldCapture := b.cmd.Options.CaptureOutput || b.cmd.CaptureOutput

	if shouldCapture {
		// Use LimitedBuffer for capture scenarios
		stdout := &LimitedBuffer{limit: MaxBufferSize}
		stderr := &LimitedBuffer{limit: MaxBufferSize}
		execCmd.Stdout = stdout
		execCmd.Stderr = stderr
		stdoutBuf = &stdout.buffer
		stderrBuf = &stderr.buffer
	} else if b.cmd.Options.OutputWriter != nil {
		execCmd.Stdout = b.cmd.Options.OutputWriter
		execCmd.Stderr = b.cmd.Options.ErrorWriter
		// Return empty buffers for non-capture scenarios
		stdoutBuf = &bytes.Buffer{}
		stderrBuf = &bytes.Buffer{}
	} else if b.cmd.Stdout != nil {
		execCmd.Stdout = b.cmd.Stdout
		execCmd.Stderr = b.cmd.Stderr
		stdoutBuf = &bytes.Buffer{}
		stderrBuf = &bytes.Buffer{}
	} else {
		// Default case - return empty buffers
		stdoutBuf = &bytes.Buffer{}
		stderrBuf = &bytes.Buffer{}
	}

	return execCmd, stdoutBuf, stderrBuf
}

// ExecuteAndCollectResult runs the command and collects the result
func (b *CommandBuilder) ExecuteAndCollectResult(execCmd *exec.Cmd, stdout, stderr *bytes.Buffer) *Result {
	start := time.Now()
	err := execCmd.Run()
	duration := time.Since(start)

	result := &Result{
		Duration: duration,
	}

	if stdout != nil {
		result.Stdout = stdout.Bytes()
	}
	if stderr != nil {
		result.Stderr = stderr.Bytes()
	}

	if err != nil {
		b.handleError(err, result)
	}

	return result
}

// handleError processes command execution errors
func (b *CommandBuilder) handleError(err error, result *Result) {
	if b.ctx != nil && b.ctx.Err() == context.DeadlineExceeded {
		result.Timeout = true
		result.Error = err
	} else if exitError, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitError.ExitCode()
		result.Error = err
	} else {
		result.ExitCode = -1
		result.Error = err
	}
}

// DetermineTimeout calculates the effective timeout for a command
func (b *CommandBuilder) DetermineTimeout(defaultTimeout time.Duration) time.Duration {
	if b.cmd.Options.Timeout > 0 {
		return b.cmd.Options.Timeout
	}
	if b.cmd.Timeout > 0 {
		return b.cmd.Timeout
	}
	return defaultTimeout
}

// ShouldStream determines if the command should stream output
func (b *CommandBuilder) ShouldStream() bool {
	return b.cmd.Options.StreamOutput || b.cmd.StreamOutput
}

// ShouldCapture determines if the command should capture output
func (b *CommandBuilder) ShouldCapture() bool {
	return b.cmd.Options.CaptureOutput || b.cmd.CaptureOutput
}

// GetOutputWriters returns the configured output writers for streaming
func (b *CommandBuilder) GetOutputWriters() (io.Writer, io.Writer) {
	// Use consolidated writer resolution logic with nil defaults
	return b.resolveWriters(nil, nil)
}

// resolveWriters consolidates the writer resolution logic with consistent precedence:
// 1. Direct command writers (highest priority)
// 2. Command options writers
// 3. Provided writers (from parameters)
// 4. OS defaults (lowest priority)
func (b *CommandBuilder) resolveWriters(outputWriter, errorWriter io.Writer) (io.Writer, io.Writer) {
	// Start with provided writers or defaults
	if outputWriter == nil {
		outputWriter = os.Stdout
	}
	if errorWriter == nil {
		errorWriter = os.Stderr
	}

	// Apply command options if they exist (medium priority)
	if b.cmd.Options.OutputWriter != nil {
		outputWriter = b.cmd.Options.OutputWriter
	}
	if b.cmd.Options.ErrorWriter != nil {
		errorWriter = b.cmd.Options.ErrorWriter
	}

	// Apply direct command writers if they exist (highest priority)
	if b.cmd.Stdout != nil {
		outputWriter = b.cmd.Stdout
	}
	if b.cmd.Stderr != nil {
		errorWriter = b.cmd.Stderr
	}

	return outputWriter, errorWriter
}
