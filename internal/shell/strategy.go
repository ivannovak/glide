package shell

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"
)

// ExecutionStrategy defines the interface for command execution strategies
type ExecutionStrategy interface {
	Execute(ctx context.Context, cmd *Command) (*Result, error)
	Name() string
}

// BasicStrategy executes commands without any special handling
type BasicStrategy struct{}

// NewBasicStrategy creates a new basic execution strategy
func NewBasicStrategy() *BasicStrategy {
	return &BasicStrategy{}
}

// Execute runs the command with basic execution
func (s *BasicStrategy) Execute(ctx context.Context, cmd *Command) (*Result, error) {
	builder := NewCommandBuilder(cmd).WithContext(ctx)
	execCmd, stdout, stderr := builder.BuildWithMixedOutput()
	result := builder.ExecuteAndCollectResult(execCmd, stdout, stderr)
	return result, nil
}

// Name returns the strategy name
func (s *BasicStrategy) Name() string {
	return "basic"
}

// TimeoutStrategy executes commands with timeout enforcement
type TimeoutStrategy struct {
	timeout time.Duration
}

// NewTimeoutStrategy creates a new timeout execution strategy
func NewTimeoutStrategy(timeout time.Duration) *TimeoutStrategy {
	return &TimeoutStrategy{
		timeout: timeout,
	}
}

// Execute runs the command with timeout
func (s *TimeoutStrategy) Execute(ctx context.Context, cmd *Command) (*Result, error) {
	// Use command timeout if specified, otherwise use strategy timeout
	timeout := s.timeout
	if cmd.Options.Timeout > 0 {
		timeout = cmd.Options.Timeout
	} else if cmd.Timeout > 0 {
		timeout = cmd.Timeout
	}

	// Create a timeout context derived from the passed context
	// This ensures parent context cancellation is respected
	var timeoutCtx context.Context
	var cancel context.CancelFunc
	if ctx != nil {
		timeoutCtx, cancel = context.WithTimeout(ctx, timeout)
	} else {
		timeoutCtx, cancel = context.WithTimeout(context.Background(), timeout)
	}
	defer cancel()

	// Use basic strategy with timeout context
	basic := NewBasicStrategy()
	result, err := basic.Execute(timeoutCtx, cmd)

	// BasicStrategy should handle timeout detection, but double-check
	if !result.Timeout && timeoutCtx.Err() == context.DeadlineExceeded {
		result.Timeout = true
		result.Error = fmt.Errorf("command timed out after %s", timeout)
	}

	return result, err
}

// Name returns the strategy name
func (s *TimeoutStrategy) Name() string {
	return "timeout"
}

// StreamingStrategy executes commands with real-time output streaming
type StreamingStrategy struct {
	outputWriter io.Writer
	errorWriter  io.Writer
}

// NewStreamingStrategy creates a new streaming execution strategy
func NewStreamingStrategy(output, error io.Writer) *StreamingStrategy {
	if output == nil {
		output = os.Stdout
	}
	if error == nil {
		error = os.Stderr
	}

	return &StreamingStrategy{
		outputWriter: output,
		errorWriter:  error,
	}
}

// Execute runs the command with streaming output
func (s *StreamingStrategy) Execute(ctx context.Context, cmd *Command) (*Result, error) {
	builder := NewCommandBuilder(cmd).WithContext(ctx)
	execCmd := builder.BuildWithStreaming(s.outputWriter, s.errorWriter)
	result := builder.ExecuteAndCollectResult(execCmd, nil, nil)
	return result, nil
}

// Name returns the strategy name
func (s *StreamingStrategy) Name() string {
	return "streaming"
}

// PipeStrategy executes commands with piping support
type PipeStrategy struct {
	inputReader io.Reader
}

// NewPipeStrategy creates a new pipe execution strategy
func NewPipeStrategy(input io.Reader) *PipeStrategy {
	return &PipeStrategy{
		inputReader: input,
	}
}

// Execute runs the command with input piping
func (s *PipeStrategy) Execute(ctx context.Context, cmd *Command) (*Result, error) {
	// Create a defensive copy to avoid mutating the original command
	cmdCopy := *cmd

	// Override stdin if strategy has input
	if s.inputReader != nil && cmdCopy.Stdin == nil {
		cmdCopy.Stdin = s.inputReader
	}

	builder := NewCommandBuilder(&cmdCopy).WithContext(ctx)

	// Use capture or not based on command options
	if cmdCopy.Options.CaptureOutput || cmdCopy.CaptureOutput {
		execCmd, stdout, stderr := builder.BuildWithCapture()
		result := builder.ExecuteAndCollectResult(execCmd, stdout, stderr)
		return result, nil
	} else {
		execCmd := builder.Build()
		result := builder.ExecuteAndCollectResult(execCmd, nil, nil)
		return result, nil
	}
}

// Name returns the strategy name
func (s *PipeStrategy) Name() string {
	return "pipe"
}

// StrategySelector selects the appropriate execution strategy
type StrategySelector struct {
	strategies map[string]ExecutionStrategy
}

// NewStrategySelector creates a new strategy selector
func NewStrategySelector() *StrategySelector {
	selector := &StrategySelector{
		strategies: make(map[string]ExecutionStrategy),
	}

	// Register default strategies
	selector.Register(NewBasicStrategy())
	selector.Register(NewTimeoutStrategy(30 * time.Second))
	selector.Register(NewStreamingStrategy(os.Stdout, os.Stderr))
	selector.Register(NewPipeStrategy(os.Stdin))

	return selector
}

// Register adds a strategy to the selector
func (s *StrategySelector) Register(strategy ExecutionStrategy) {
	s.strategies[strategy.Name()] = strategy
}

// Select chooses the appropriate strategy based on command options
func (s *StrategySelector) Select(cmd *Command) ExecutionStrategy {
	// Choose strategy based on command options
	if cmd.Options.Timeout > 0 || cmd.Timeout > 0 {
		timeout := cmd.Options.Timeout
		if timeout == 0 {
			timeout = cmd.Timeout
		}
		return NewTimeoutStrategy(timeout)
	}

	if cmd.Options.StreamOutput || cmd.StreamOutput {
		outputWriter := cmd.Options.OutputWriter
		errorWriter := cmd.Options.ErrorWriter
		if outputWriter == nil {
			outputWriter = cmd.Stdout
		}
		if errorWriter == nil {
			errorWriter = cmd.Stderr
		}
		return NewStreamingStrategy(outputWriter, errorWriter)
	}

	if cmd.Stdin != nil {
		return NewPipeStrategy(cmd.Stdin)
	}

	// Default to basic strategy
	return NewBasicStrategy()
}

// Get retrieves a strategy by name
func (s *StrategySelector) Get(name string) (ExecutionStrategy, bool) {
	strategy, ok := s.strategies[name]
	return strategy, ok
}
