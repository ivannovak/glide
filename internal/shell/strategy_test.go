package shell

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStrategySelector_Select(t *testing.T) {
	selector := NewStrategySelector()

	tests := []struct {
		name         string
		command      *Command
		expectedType string
	}{
		{
			name: "basic strategy for simple command",
			command: &Command{
				Name: "echo",
				Args: []string{"test"},
				Options: CommandOptions{
					CaptureOutput: true,
				},
			},
			expectedType: "basic",
		},
		{
			name: "timeout strategy for command with timeout",
			command: &Command{
				Name: "sleep",
				Args: []string{"1"},
				Options: CommandOptions{
					Timeout: 2 * time.Second,
				},
			},
			expectedType: "timeout",
		},
		{
			name: "streaming strategy for stream output",
			command: &Command{
				Name: "ls",
				Options: CommandOptions{
					StreamOutput: true,
					OutputWriter: &bytes.Buffer{},
				},
			},
			expectedType: "streaming",
		},
		{
			name: "pipe strategy for stdin input",
			command: &Command{
				Name:  "cat",
				Stdin: bytes.NewBufferString("test input"),
			},
			expectedType: "pipe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := selector.Select(tt.command)
			assert.NotNil(t, strategy)
			assert.Equal(t, tt.expectedType, strategy.Name())
		})
	}
}

func TestBasicStrategy_Execute(t *testing.T) {
	// Skip if in CI environment
	if os.Getenv("CI") != "" {
		t.Skip("Skipping executor tests in CI")
	}

	strategy := NewBasicStrategy()

	t.Run("successful command", func(t *testing.T) {
		cmd := &Command{
			Name: "echo",
			Args: []string{"hello"},
			Options: CommandOptions{
				CaptureOutput: true,
			},
		}

		result, err := strategy.Execute(context.Background(), cmd)
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, string(result.Stdout), "hello")
	})

	t.Run("command with error", func(t *testing.T) {
		cmd := &Command{
			Name: "ls",
			Args: []string{"/nonexistent"},
			Options: CommandOptions{
				CaptureOutput: true,
			},
		}

		result, err := strategy.Execute(context.Background(), cmd)
		assert.NoError(t, err) // Command execution succeeds even if command returns error
		assert.NotEqual(t, 0, result.ExitCode)
	})
}

func TestTimeoutStrategy_Execute(t *testing.T) {
	// Skip if in CI environment
	if os.Getenv("CI") != "" {
		t.Skip("Skipping executor tests in CI")
	}

	strategy := NewTimeoutStrategy(5 * time.Second)

	t.Run("command completes within timeout", func(t *testing.T) {
		cmd := &Command{
			Name: "echo",
			Args: []string{"quick"},
			Options: CommandOptions{
				Timeout:       1 * time.Second,
				CaptureOutput: true,
			},
		}

		result, err := strategy.Execute(context.Background(), cmd)
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.False(t, result.Timeout)
	})

	t.Run("command exceeds timeout", func(t *testing.T) {
		cmd := &Command{
			Name: "sleep",
			Args: []string{"2"},
			Options: CommandOptions{
				Timeout:       100 * time.Millisecond,
				CaptureOutput: true,
			},
		}

		result, err := strategy.Execute(context.Background(), cmd)
		assert.NoError(t, err) // Strategy doesn't return error for timeout
		assert.True(t, result.Timeout)
		assert.NotNil(t, result.Error) // But error is in result
	})
}

func TestStreamingStrategy_Execute(t *testing.T) {
	// Skip if in CI environment
	if os.Getenv("CI") != "" {
		t.Skip("Skipping executor tests in CI")
	}

	strategy := NewStreamingStrategy(os.Stdout, os.Stderr)

	t.Run("stream to buffer", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}

		cmd := &Command{
			Name: "echo",
			Args: []string{"streamed output"},
			Options: CommandOptions{
				StreamOutput: true,
				OutputWriter: stdout,
				ErrorWriter:  stderr,
			},
		}

		result, err := strategy.Execute(context.Background(), cmd)
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, stdout.String(), "streamed output")
	})
}

func TestPipeStrategy_Execute(t *testing.T) {
	// Skip if in CI environment
	if os.Getenv("CI") != "" {
		t.Skip("Skipping executor tests in CI")
	}

	strategy := NewPipeStrategy(nil)

	t.Run("pipe input to command", func(t *testing.T) {
		input := "piped input\n"
		cmd := &Command{
			Name:  "cat",
			Stdin: bytes.NewBufferString(input),
			Options: CommandOptions{
				CaptureOutput: true,
			},
		}

		result, err := strategy.Execute(context.Background(), cmd)
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Equal(t, input, string(result.Stdout))
	})
}

func TestStrategySelector_Create(t *testing.T) {
	selector := NewStrategySelector()
	assert.NotNil(t, selector)
}

func TestExecutionStrategy_Name(t *testing.T) {
	tests := []struct {
		strategy ExecutionStrategy
		expected string
	}{
		{NewBasicStrategy(), "basic"},
		{NewTimeoutStrategy(5 * time.Second), "timeout"},
		{NewStreamingStrategy(nil, nil), "streaming"},
		{NewPipeStrategy(nil), "pipe"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.strategy.Name())
		})
	}
}
