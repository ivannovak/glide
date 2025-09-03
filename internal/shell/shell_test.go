package shell

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommand_WithMethods(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "WithTimeout sets timeout",
			test: func(t *testing.T) {
				cmd := NewCommand("echo", "test")
				cmd.WithTimeout(5 * time.Second)
				assert.Equal(t, 5*time.Second, cmd.Timeout)
			},
		},
		{
			name: "WithWorkingDir sets working directory",
			test: func(t *testing.T) {
				cmd := NewCommand("ls")
				cmd.WithWorkingDir("/tmp")
				assert.Equal(t, "/tmp", cmd.WorkingDir)
			},
		},
		{
			name: "WithEnv adds environment variables",
			test: func(t *testing.T) {
				cmd := NewCommand("env")
				cmd.WithEnv("FOO=bar", "BAZ=qux")
				assert.Contains(t, cmd.Environment, "FOO=bar")
				assert.Contains(t, cmd.Environment, "BAZ=qux")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

func TestCommand_String(t *testing.T) {
	tests := []struct {
		name     string
		command  *Command
		expected string
	}{
		{
			name:     "command without args",
			command:  NewCommand("ls"),
			expected: "ls",
		},
		{
			name:     "command with simple args",
			command:  NewCommand("echo", "hello", "world"),
			expected: "echo hello world",
		},
		{
			name:     "command with spaces in args",
			command:  NewCommand("echo", "hello world", "test"),
			expected: `echo "hello world" test`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.command.String())
		})
	}
}

func TestJoinArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "simple args",
			args:     []string{"arg1", "arg2", "arg3"},
			expected: "arg1 arg2 arg3",
		},
		{
			name:     "args with spaces",
			args:     []string{"hello world", "test"},
			expected: "'hello world' test",
		},
		{
			name:     "args with special characters",
			args:     []string{"$HOME", "test\"quote", "back\\slash"},
			expected: "'$HOME' 'test\"quote' 'back\\slash'",
		},
		{
			name:     "args with single quotes",
			args:     []string{"it's", "test"},
			expected: "'it'\\''s' test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JoinArgs(tt.args)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewCommand(t *testing.T) {
	cmd := NewCommand("echo", "hello")

	assert.Equal(t, "echo", cmd.Name)
	assert.Equal(t, []string{"hello"}, cmd.Args)
	assert.Equal(t, ModeCapture, cmd.Mode)
	assert.True(t, cmd.InheritEnv)
	assert.True(t, cmd.SignalForward)
}

func TestNewPassthroughCommand(t *testing.T) {
	cmd := NewPassthroughCommand("docker", "ps")

	assert.Equal(t, "docker", cmd.Name)
	assert.Equal(t, []string{"ps"}, cmd.Args)
	assert.Equal(t, ModePassthrough, cmd.Mode)
}

func TestNewInteractiveCommand(t *testing.T) {
	cmd := NewInteractiveCommand("bash")

	assert.Equal(t, "bash", cmd.Name)
	assert.Equal(t, ModeInteractive, cmd.Mode)
	assert.True(t, cmd.AllocateTTY)
}

func TestExecutor_Execute(t *testing.T) {
	// Skip tests that require actual command execution in CI
	if os.Getenv("CI") != "" {
		t.Skip("Skipping executor tests in CI")
	}

	executor := NewExecutor(Options{})

	t.Run("capture mode", func(t *testing.T) {
		cmd := NewCommand("echo", "hello")
		result, err := executor.Execute(cmd)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, string(result.Stdout), "hello")
	})

	t.Run("command with timeout", func(t *testing.T) {
		cmd := NewCommand("sleep", "0.1")
		cmd.WithTimeout(1 * time.Second)

		result, err := executor.Execute(cmd)
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.False(t, result.Timeout)
	})

	t.Run("command timeout exceeded", func(t *testing.T) {
		cmd := NewCommand("sleep", "1")        // Reduce sleep time
		cmd.WithTimeout(50 * time.Millisecond) // Very short timeout to ensure it triggers

		result, err := executor.Execute(cmd)
		assert.NoError(t, err) // Timeout doesn't return error, sets result.Timeout instead

		// Debug output to understand what's happening
		t.Logf("Timeout: %v, Error: %v, ExitCode: %d, err: %v", result.Timeout, result.Error, result.ExitCode, err)

		assert.True(t, result.Timeout)
		assert.NotNil(t, result.Error) // Error should be in result.Error field
	})
}

func TestResult_ExitCode(t *testing.T) {
	tests := []struct {
		name     string
		result   *Result
		expected int
	}{
		{
			name:     "exit code 0",
			result:   &Result{ExitCode: 0},
			expected: 0,
		},
		{
			name:     "exit code 1",
			result:   &Result{ExitCode: 1},
			expected: 1,
		},
		{
			name:     "exit code -1",
			result:   &Result{ExitCode: -1},
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.result.ExitCode)
		})
	}
}

func TestCommandOptions(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	opts := CommandOptions{
		CaptureOutput: true,
		StreamOutput:  false,
		Timeout:       5 * time.Second,
		OutputWriter:  stdout,
		ErrorWriter:   stderr,
	}

	assert.True(t, opts.CaptureOutput)
	assert.False(t, opts.StreamOutput)
	assert.Equal(t, 5*time.Second, opts.Timeout)
	assert.Equal(t, stdout, opts.OutputWriter)
	assert.Equal(t, stderr, opts.ErrorWriter)
}

func TestExecutionMode(t *testing.T) {
	modes := []ExecutionMode{
		ModePassthrough,
		ModeCapture,
		ModeInteractive,
		ModeBackground,
	}

	expected := []string{
		"passthrough",
		"capture",
		"interactive",
		"background",
	}

	for i, mode := range modes {
		assert.Equal(t, ExecutionMode(expected[i]), mode)
	}
}

func TestCommand_UseStrategy(t *testing.T) {
	cmd := &Command{
		Name:          "echo",
		Args:          []string{"test"},
		UseStrategy:   true,
		CaptureOutput: true,
		Options: CommandOptions{
			CaptureOutput: true,
			Timeout:       5 * time.Second,
		},
	}

	assert.True(t, cmd.UseStrategy)
	assert.True(t, cmd.CaptureOutput)
	assert.Equal(t, 5*time.Second, cmd.Options.Timeout)
}

// Helper function for testing
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	out, _ := io.ReadAll(r)
	return string(out)
}
