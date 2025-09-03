package shell

import (
	"bytes"
	"context"
	"testing"
	"time"
)

// BenchmarkCommand_Creation benchmarks command creation
func BenchmarkCommand_Creation(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		cmd := NewCommand("echo", "hello", "world")
		_ = cmd
	}
}

// BenchmarkCommand_String benchmarks command string representation
func BenchmarkCommand_String(b *testing.B) {
	cmd := NewCommand("echo", "hello", "world", "with", "many", "arguments")
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = cmd.String()
	}
}

// BenchmarkJoinArgs benchmarks argument joining
func BenchmarkJoinArgs(b *testing.B) {
	args := []string{
		"arg1", "arg2", "arg with spaces", "arg'with'quotes",
		"arg4", "arg5", "arg6", "arg7", "arg8", "arg9", "arg10",
	}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = JoinArgs(args)
	}
}

// BenchmarkBasicStrategy_Execute benchmarks basic strategy execution
func BenchmarkBasicStrategy_Execute(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	strategy := NewBasicStrategy()
	cmd := &Command{
		Name: "echo",
		Args: []string{"benchmark", "test"},
		Options: CommandOptions{
			CaptureOutput: true,
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, err := strategy.Execute(context.Background(), cmd)
		if err != nil {
			b.Fatal(err)
		}
		if result.ExitCode != 0 {
			b.Fatalf("Command failed with exit code %d", result.ExitCode)
		}
	}
}

// BenchmarkTimeoutStrategy_Execute benchmarks timeout strategy execution
func BenchmarkTimeoutStrategy_Execute(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	strategy := NewTimeoutStrategy(5 * time.Second)
	cmd := &Command{
		Name: "echo",
		Args: []string{"timeout", "benchmark"},
		Options: CommandOptions{
			CaptureOutput: true,
			Timeout:       1 * time.Second,
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, err := strategy.Execute(context.Background(), cmd)
		if err != nil {
			b.Fatal(err)
		}
		if result.ExitCode != 0 {
			b.Fatalf("Command failed with exit code %d", result.ExitCode)
		}
	}
}

// BenchmarkStreamingStrategy_Execute benchmarks streaming strategy execution
func BenchmarkStreamingStrategy_Execute(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	strategy := NewStreamingStrategy(stdout, stderr)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		stdout.Reset()
		stderr.Reset()

		cmd := &Command{
			Name: "echo",
			Args: []string{"streaming", "benchmark"},
			Options: CommandOptions{
				StreamOutput: true,
				OutputWriter: stdout,
				ErrorWriter:  stderr,
			},
		}

		result, err := strategy.Execute(context.Background(), cmd)
		if err != nil {
			b.Fatal(err)
		}
		if result.ExitCode != 0 {
			b.Fatalf("Command failed with exit code %d", result.ExitCode)
		}
	}
}

// BenchmarkPipeStrategy_Execute benchmarks pipe strategy execution
func BenchmarkPipeStrategy_Execute(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	strategy := NewPipeStrategy(nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		input := bytes.NewBufferString("benchmark input data\n")
		cmd := &Command{
			Name:  "cat",
			Stdin: input,
			Options: CommandOptions{
				CaptureOutput: true,
			},
		}

		result, err := strategy.Execute(context.Background(), cmd)
		if err != nil {
			b.Fatal(err)
		}
		if result.ExitCode != 0 {
			b.Fatalf("Command failed with exit code %d", result.ExitCode)
		}
	}
}

// BenchmarkStrategySelector_Select benchmarks strategy selection
func BenchmarkStrategySelector_Select(b *testing.B) {
	selector := NewStrategySelector()

	// Test different command scenarios
	commands := []*Command{
		{
			Name: "echo",
			Args: []string{"basic"},
			Options: CommandOptions{
				CaptureOutput: true,
			},
		},
		{
			Name: "sleep",
			Args: []string{"0.01"},
			Options: CommandOptions{
				Timeout: 1 * time.Second,
			},
		},
		{
			Name: "ls",
			Options: CommandOptions{
				StreamOutput: true,
				OutputWriter: &bytes.Buffer{},
			},
		},
		{
			Name:  "cat",
			Stdin: bytes.NewBufferString("input"),
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		cmd := commands[i%len(commands)]
		strategy := selector.Select(cmd)
		_ = strategy
	}
}

// BenchmarkExecutor_Execute benchmarks full executor execution
func BenchmarkExecutor_Execute(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	executor := NewExecutor(Options{})
	cmd := NewCommand("echo", "executor", "benchmark")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, err := executor.Execute(cmd)
		if err != nil {
			b.Fatal(err)
		}
		if result.ExitCode != 0 {
			b.Fatalf("Command failed with exit code %d", result.ExitCode)
		}
	}
}

// BenchmarkExecutor_ExecuteWithStrategy benchmarks executor with strategy pattern
func BenchmarkExecutor_ExecuteWithStrategy(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	executor := NewExecutor(Options{})
	cmd := NewCommand("echo", "strategy", "benchmark")
	cmd.UseStrategy = true
	cmd.Options = CommandOptions{
		CaptureOutput: true,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, err := executor.Execute(cmd)
		if err != nil {
			b.Fatal(err)
		}
		if result.ExitCode != 0 {
			b.Fatalf("Command failed with exit code %d", result.ExitCode)
		}
	}
}

// BenchmarkCommand_WithMethods benchmarks command builder methods
func BenchmarkCommand_WithMethods(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		cmd := NewCommand("test")
		cmd.WithTimeout(5*time.Second).
			WithWorkingDir("/tmp").
			WithEnv("FOO=bar", "BAZ=qux")
		_ = cmd
	}
}

// BenchmarkResult_Creation benchmarks result struct creation
func BenchmarkResult_Creation(b *testing.B) {
	stdout := []byte("test output")
	stderr := []byte("test error")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result := &Result{
			ExitCode: 0,
			Stdout:   stdout,
			Stderr:   stderr,
			Error:    nil,
			Duration: time.Millisecond * 100,
			Timeout:  false,
		}
		_ = result
	}
}

// BenchmarkCommandOptions_Creation benchmarks command options creation
func BenchmarkCommandOptions_Creation(b *testing.B) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		opts := CommandOptions{
			CaptureOutput: true,
			StreamOutput:  false,
			Timeout:       5 * time.Second,
			OutputWriter:  stdout,
			ErrorWriter:   stderr,
		}
		_ = opts
	}
}
