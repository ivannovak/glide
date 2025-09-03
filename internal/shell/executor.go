package shell

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/fatih/color"
)

// Executor handles command execution
type Executor struct {
	options  Options
	verbose  bool
	selector *StrategySelector
}

// NewExecutor creates a new command executor
func NewExecutor(options Options) *Executor {
	return &Executor{
		options:  options,
		verbose:  options.Verbose,
		selector: NewStrategySelector(),
	}
}

// Execute runs a command based on its mode or strategy
func (e *Executor) Execute(cmd *Command) (*Result, error) {
	if e.verbose {
		color.Cyan("› %s", cmd.String())
	}
	
	// Use strategy pattern if enabled
	if cmd.UseStrategy {
		strategy := e.selector.Select(cmd)
		return strategy.Execute(context.Background(), cmd)
	}
	
	// Legacy mode-based execution for backward compatibility
	start := time.Now()
	switch cmd.Mode {
	case ModePassthrough:
		return e.executePassthrough(cmd, start)
	case ModeInteractive:
		return e.executeInteractive(cmd, start)
	case ModeCapture:
		return e.executeCapture(cmd, start)
	case ModeBackground:
		return e.executeBackground(cmd, start)
	default:
		return e.executeCapture(cmd, start)
	}
}

// ExecuteWithContext runs a command with a context for cancellation using strategy pattern
func (e *Executor) ExecuteWithContext(ctx context.Context, cmd *Command) (*Result, error) {
	if e.verbose {
		color.Cyan("› %s", cmd.String())
	}
	
	// Always use strategy pattern when context is provided
	cmd.UseStrategy = true
	strategy := e.selector.Select(cmd)
	return strategy.Execute(ctx, cmd)
}

// executePassthrough runs a command with direct I/O passthrough
func (e *Executor) executePassthrough(cmd *Command, start time.Time) (*Result, error) {
	ctx := context.Background()
	if cmd.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cmd.Timeout)
		defer cancel()
	}
	
	execCmd := exec.CommandContext(ctx, cmd.Name, cmd.Args...)
	
	// Set working directory
	if cmd.WorkingDir != "" {
		execCmd.Dir = cmd.WorkingDir
	}
	
	// Configure environment
	if cmd.InheritEnv {
		execCmd.Env = os.Environ()
	}
	execCmd.Env = append(execCmd.Env, e.options.GlobalEnv...)
	execCmd.Env = append(execCmd.Env, cmd.Environment...)
	
	// Direct I/O passthrough
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	
	// Signal forwarding
	var cleanupSignals func()
	if cmd.SignalForward {
		cleanupSignals = e.setupSignalForwarding(execCmd)
		defer cleanupSignals()
	}
	
	// Run the command
	err := execCmd.Run()
	
	result := &Result{
		Duration: time.Since(start),
	}
	
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			result.Timeout = true
			result.Error = fmt.Errorf("command timed out after %s", cmd.Timeout)
			return result, nil
		} else if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = -1
			result.Error = err
		}
	}
	
	return result, nil
}

// executeInteractive runs a command with TTY allocation
func (e *Executor) executeInteractive(cmd *Command, start time.Time) (*Result, error) {
	// For interactive commands, we use passthrough with TTY settings
	// This is simplified - full TTY support would require pty package
	return e.executePassthrough(cmd, start)
}

// executeCapture runs a command and captures output
func (e *Executor) executeCapture(cmd *Command, start time.Time) (*Result, error) {
	ctx := context.Background()
	if cmd.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cmd.Timeout)
		defer cancel()
	}
	
	execCmd := exec.CommandContext(ctx, cmd.Name, cmd.Args...)
	
	// Set working directory
	if cmd.WorkingDir != "" {
		execCmd.Dir = cmd.WorkingDir
	}
	
	// Configure environment
	if cmd.InheritEnv {
		execCmd.Env = os.Environ()
	}
	execCmd.Env = append(execCmd.Env, e.options.GlobalEnv...)
	execCmd.Env = append(execCmd.Env, cmd.Environment...)
	
	// Capture output
	var stdout, stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr
	
	// Custom I/O if provided
	if cmd.Stdin != nil {
		execCmd.Stdin = cmd.Stdin
	}
	if cmd.Stdout != nil {
		execCmd.Stdout = io.MultiWriter(&stdout, cmd.Stdout)
	}
	if cmd.Stderr != nil {
		execCmd.Stderr = io.MultiWriter(&stderr, cmd.Stderr)
	}
	
	// Run the command
	err := execCmd.Run()
	
	result := &Result{
		Stdout:   stdout.Bytes(),
		Stderr:   stderr.Bytes(),
		Duration: time.Since(start),
	}
	
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			result.Timeout = true
			result.Error = fmt.Errorf("command timed out after %s", cmd.Timeout)
			return result, nil
		} else if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = -1
			result.Error = err
		}
	}
	
	return result, nil
}

// executeBackground starts a command in the background
func (e *Executor) executeBackground(cmd *Command, start time.Time) (*Result, error) {
	execCmd := exec.Command(cmd.Name, cmd.Args...)
	
	// Set working directory
	if cmd.WorkingDir != "" {
		execCmd.Dir = cmd.WorkingDir
	}
	
	// Configure environment
	if cmd.InheritEnv {
		execCmd.Env = os.Environ()
	}
	execCmd.Env = append(execCmd.Env, e.options.GlobalEnv...)
	execCmd.Env = append(execCmd.Env, cmd.Environment...)
	
	// Start the command
	err := execCmd.Start()
	if err != nil {
		return &Result{
			ExitCode: -1,
			Error:    err,
			Duration: time.Since(start),
		}, err
	}
	
	// Return immediately for background commands
	return &Result{
		ExitCode: 0,
		Duration: time.Since(start),
	}, nil
}

// setupSignalForwarding sets up signal forwarding to subprocess
// It returns a cleanup function that should be called after the command completes
func (e *Executor) setupSignalForwarding(cmd *exec.Cmd) func() {
	// Create a channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		for sig := range sigChan {
			if cmd.Process != nil {
				// Forward the signal to the subprocess
				cmd.Process.Signal(sig)
			}
		}
	}()
	
	// Return cleanup function to be called after command completes
	return func() {
		signal.Stop(sigChan)
		close(sigChan)
	}
}

// Run is a convenience method for simple command execution
func (e *Executor) Run(name string, args ...string) error {
	cmd := NewPassthroughCommand(name, args...)
	result, err := e.Execute(cmd)
	if err != nil {
		return err
	}
	if result.Error != nil {
		return result.Error
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("command failed with exit code %d", result.ExitCode)
	}
	return nil
}

// RunCapture runs a command and returns captured output
func (e *Executor) RunCapture(name string, args ...string) (string, error) {
	cmd := NewCommand(name, args...)
	result, err := e.Execute(cmd)
	if err != nil {
		return "", err
	}
	if result.Error != nil {
		return "", result.Error
	}
	if result.ExitCode != 0 {
		return string(result.Stderr), fmt.Errorf("command failed with exit code %d", result.ExitCode)
	}
	return string(result.Stdout), nil
}

// RunWithTimeout runs a command with a timeout
func (e *Executor) RunWithTimeout(timeout time.Duration, name string, args ...string) error {
	cmd := NewPassthroughCommand(name, args...).WithTimeout(timeout)
	result, err := e.Execute(cmd)
	if err != nil {
		return err
	}
	if result.Timeout {
		return fmt.Errorf("command timed out after %s", timeout)
	}
	if result.Error != nil {
		return result.Error
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("command failed with exit code %d", result.ExitCode)
	}
	return nil
}