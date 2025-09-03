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
	start := time.Now()

	var execCmd *exec.Cmd
	if ctx != nil {
		execCmd = exec.CommandContext(ctx, cmd.Name, cmd.Args...)
	} else {
		execCmd = exec.Command(cmd.Name, cmd.Args...)
	}

	// Set working directory if specified
	if cmd.WorkingDir != "" {
		execCmd.Dir = cmd.WorkingDir
	}

	// Set environment if specified
	if len(cmd.Environment) > 0 {
		execCmd.Env = os.Environ()
		for _, env := range cmd.Environment {
			execCmd.Env = append(execCmd.Env, env)
		}
	}

	// Handle output capture
	var stdout, stderr bytes.Buffer
	if cmd.Options.CaptureOutput || cmd.CaptureOutput {
		execCmd.Stdout = &stdout
		execCmd.Stderr = &stderr
	} else if cmd.Options.OutputWriter != nil {
		execCmd.Stdout = cmd.Options.OutputWriter
		execCmd.Stderr = cmd.Options.ErrorWriter
	} else if cmd.Stdout != nil {
		execCmd.Stdout = cmd.Stdout
		execCmd.Stderr = cmd.Stderr
	}

	// Handle stdin
	if cmd.Stdin != nil {
		execCmd.Stdin = cmd.Stdin
	}

	// Execute the command
	err := execCmd.Run()
	duration := time.Since(start)

	result := &Result{
		Stdout:   stdout.Bytes(),
		Stderr:   stderr.Bytes(),
		Duration: duration,
	}

	if err != nil {
		if ctx != nil && ctx.Err() == context.DeadlineExceeded {
			result.Timeout = true
			result.Error = fmt.Errorf("command timed out")
		} else if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = -1
		}
		if !result.Timeout {
			result.Error = err
		}
	}

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

	// Create a timeout context
	timeoutCtx, cancel := context.WithTimeout(context.Background(), timeout)
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
	start := time.Now()

	var execCmd *exec.Cmd
	if ctx != nil {
		execCmd = exec.CommandContext(ctx, cmd.Name, cmd.Args...)
	} else {
		execCmd = exec.Command(cmd.Name, cmd.Args...)
	}

	// Set working directory if specified
	if cmd.WorkingDir != "" {
		execCmd.Dir = cmd.WorkingDir
	}

	// Set environment if specified
	if len(cmd.Environment) > 0 {
		execCmd.Env = os.Environ()
		for _, env := range cmd.Environment {
			execCmd.Env = append(execCmd.Env, env)
		}
	}

	// Stream output - use command options if provided, otherwise use strategy defaults
	outputWriter := s.outputWriter
	errorWriter := s.errorWriter

	if cmd.Options.OutputWriter != nil {
		outputWriter = cmd.Options.OutputWriter
	}
	if cmd.Options.ErrorWriter != nil {
		errorWriter = cmd.Options.ErrorWriter
	}

	execCmd.Stdout = outputWriter
	execCmd.Stderr = errorWriter

	// Handle stdin
	if cmd.Stdin != nil {
		execCmd.Stdin = cmd.Stdin
	}

	// Execute the command
	err := execCmd.Run()
	duration := time.Since(start)

	result := &Result{
		Duration: duration,
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = -1
		}
		result.Error = err
	}

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
	start := time.Now()

	var execCmd *exec.Cmd
	if ctx != nil {
		execCmd = exec.CommandContext(ctx, cmd.Name, cmd.Args...)
	} else {
		execCmd = exec.Command(cmd.Name, cmd.Args...)
	}

	// Set working directory if specified
	if cmd.WorkingDir != "" {
		execCmd.Dir = cmd.WorkingDir
	}

	// Set environment if specified
	if len(cmd.Environment) > 0 {
		execCmd.Env = os.Environ()
		for _, env := range cmd.Environment {
			execCmd.Env = append(execCmd.Env, env)
		}
	}

	// Set input pipe - use command stdin if provided, otherwise use strategy input
	if cmd.Stdin != nil {
		execCmd.Stdin = cmd.Stdin
	} else if s.inputReader != nil {
		execCmd.Stdin = s.inputReader
	}

	// Handle output capture
	var stdout, stderr bytes.Buffer
	if cmd.Options.CaptureOutput || cmd.CaptureOutput {
		execCmd.Stdout = &stdout
		execCmd.Stderr = &stderr
	}

	// Execute the command
	err := execCmd.Run()
	duration := time.Since(start)

	result := &Result{
		Stdout:   stdout.Bytes(),
		Stderr:   stderr.Bytes(),
		Duration: duration,
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = -1
		}
		result.Error = err
	}

	return result, nil
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
