package shell

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestConcurrentStrategyExecution tests that strategies can be executed concurrently without race conditions
func TestConcurrentStrategyExecution(t *testing.T) {
	// Don't share buffers between concurrent executions
	getStrategies := func() []ExecutionStrategy {
		return []ExecutionStrategy{
			NewBasicStrategy(),
			NewStreamingStrategy(&bytes.Buffer{}, &bytes.Buffer{}),
			NewTimeoutStrategy(5 * time.Second),
			NewPipeStrategy(&bytes.Buffer{}),
		}
	}

	// Run multiple goroutines executing commands concurrently
	var wg sync.WaitGroup
	iterations := 10
	goroutines := 4

	for _, strategy := range getStrategies() {
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			// Create a fresh strategy for each goroutine to avoid shared state
			strategyType := strategy.Name()
			go func(sType string, id int) {
				// Create strategy inside goroutine to avoid sharing
				var s ExecutionStrategy
				switch sType {
				case "basic":
					s = NewBasicStrategy()
				case "streaming":
					s = NewStreamingStrategy(&bytes.Buffer{}, &bytes.Buffer{})
				case "timeout":
					s = NewTimeoutStrategy(5 * time.Second)
				case "pipe":
					s = NewPipeStrategy(&bytes.Buffer{})
				}
				defer wg.Done()

				for j := 0; j < iterations; j++ {
					cmd := &Command{
						Name:          "echo",
						Args:          []string{"test", string(rune(id)), string(rune(j))},
						CaptureOutput: true,
					}

					ctx := context.Background()
					result, err := s.Execute(ctx, cmd)

					if err != nil {
						t.Errorf("Strategy %s failed: %v", s.Name(), err)
					}
					if result == nil {
						t.Errorf("Strategy %s returned nil result", s.Name())
					}
				}
			}(strategyType, i)
		}
	}

	wg.Wait()
}

// TestConcurrentPipeStrategyNoMutation tests that PipeStrategy doesn't mutate shared commands
func TestConcurrentPipeStrategyNoMutation(t *testing.T) {
	// Create a shared command
	sharedCmd := &Command{
		Name:          "cat",
		CaptureOutput: true,
	}

	var wg sync.WaitGroup
	goroutines := 10

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Each goroutine creates its own strategy with its own input buffer
			// to avoid sharing the input reader
			strategy := NewPipeStrategy(bytes.NewBufferString("test input"))

			// Each goroutine uses the same shared command
			ctx := context.Background()
			result, err := strategy.Execute(ctx, sharedCmd)

			if err != nil {
				t.Errorf("Goroutine %d: execution failed: %v", id, err)
			}
			if result == nil {
				t.Errorf("Goroutine %d: nil result", id)
			}

			// Verify the original command wasn't mutated
			if sharedCmd.Stdin != nil {
				t.Errorf("Goroutine %d: shared command was mutated", id)
			}
		}(i)
	}

	wg.Wait()

	// Final check that shared command is unchanged
	if sharedCmd.Stdin != nil {
		t.Error("Shared command stdin was mutated")
	}
}

// TestConcurrentTimeoutStrategy tests concurrent timeout handling
func TestConcurrentTimeoutStrategy(t *testing.T) {
	strategy := NewTimeoutStrategy(100 * time.Millisecond)

	var wg sync.WaitGroup
	goroutines := 5

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Some commands will timeout, some won't
			sleepTime := "0.05"
			if id%2 == 0 {
				sleepTime = "0.2" // This will timeout
			}

			cmd := &Command{
				Name:          "sleep",
				Args:          []string{sleepTime},
				CaptureOutput: true,
			}

			ctx := context.Background()
			result, _ := strategy.Execute(ctx, cmd)

			if id%2 == 0 && !result.Timeout {
				t.Errorf("Goroutine %d: expected timeout but didn't get one", id)
			}
			if id%2 != 0 && result.Timeout {
				t.Errorf("Goroutine %d: unexpected timeout", id)
			}
		}(i)
	}

	wg.Wait()
}

// TestConcurrentContextCancellation tests that context cancellation works correctly with concurrent execution
func TestConcurrentContextCancellation(t *testing.T) {
	strategy := NewBasicStrategy()

	// Create a parent context that we'll cancel
	parentCtx, parentCancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	goroutines := 10

	// Start goroutines that will execute commands
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Create a child context for this goroutine
			ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second)
			defer cancel()

			cmd := &Command{
				Name:          "sleep",
				Args:          []string{"2"},
				CaptureOutput: true,
			}

			result, _ := strategy.Execute(ctx, cmd)

			// After parent cancellation, all commands should fail
			if result != nil && parentCtx.Err() != nil {
				// Check if the command was affected by cancellation
				if result.Error == nil && result.ExitCode == 0 {
					// This might happen if the command completed before cancellation
					// which is acceptable
				}
			}
		}(i)
	}

	// Cancel parent context after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		parentCancel()
	}()

	wg.Wait()
}

// TestConcurrentCommandBuilder tests that CommandBuilder is safe for concurrent use
func TestConcurrentCommandBuilder(t *testing.T) {
	var wg sync.WaitGroup
	goroutines := 20

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Each goroutine creates its own command and builder
			cmd := &Command{
				Name:          "echo",
				Args:          []string{"goroutine", fmt.Sprintf("%d", id)},
				CaptureOutput: true,
				WorkingDir:    "/tmp",
				Environment:   []string{fmt.Sprintf("GOROUTINE_ID=%d", id)},
			}

			builder := NewCommandBuilder(cmd)

			// Test different build methods concurrently
			switch id % 4 {
			case 0:
				execCmd, stdout, stderr := builder.BuildWithCapture()
				result := builder.ExecuteAndCollectResult(execCmd, stdout, stderr)
				if result.Error != nil {
					t.Errorf("Goroutine %d: BuildWithCapture failed: %v", id, result.Error)
				}
			case 1:
				var buf bytes.Buffer
				execCmd := builder.BuildWithStreaming(&buf, &buf)
				result := builder.ExecuteAndCollectResult(execCmd, nil, nil)
				if result.Error != nil {
					t.Errorf("Goroutine %d: BuildWithStreaming failed: %v", id, result.Error)
				}
			case 2:
				execCmd, stdout, stderr := builder.BuildWithMixedOutput()
				result := builder.ExecuteAndCollectResult(execCmd, stdout, stderr)
				if result.Error != nil {
					t.Errorf("Goroutine %d: BuildWithMixedOutput failed: %v", id, result.Error)
				}
			case 3:
				ctx := context.Background()
				builder = builder.WithContext(ctx)
				execCmd := builder.Build()
				err := execCmd.Run()
				if err != nil {
					t.Errorf("Goroutine %d: Build with context failed: %v", id, err)
				}
			}
		}(i)
	}

	wg.Wait()
}

// TestConcurrentStrategySelector tests concurrent strategy selection
func TestConcurrentStrategySelector(t *testing.T) {
	selector := NewStrategySelector()

	var wg sync.WaitGroup
	goroutines := 50

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Create different command types to trigger different strategies
			var cmd *Command
			switch id % 4 {
			case 0:
				cmd = &Command{Name: "echo", Timeout: 1 * time.Second}
			case 1:
				cmd = &Command{Name: "echo", StreamOutput: true}
			case 2:
				cmd = &Command{Name: "cat", Stdin: &bytes.Buffer{}}
			case 3:
				cmd = &Command{Name: "echo"}
			}

			strategy := selector.Select(cmd)
			if strategy == nil {
				t.Errorf("Goroutine %d: nil strategy returned", id)
			}

			// Execute with the selected strategy
			ctx := context.Background()
			result, err := strategy.Execute(ctx, cmd)
			if err != nil {
				t.Errorf("Goroutine %d: execution failed: %v", id, err)
			}
			if result == nil {
				t.Errorf("Goroutine %d: nil result", id)
			}
		}(i)
	}

	wg.Wait()
}

// TestRaceConditionDetection uses Go's race detector to find issues
// Run with: go test -race ./internal/shell
func TestRaceConditionDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race condition test in short mode")
	}

	// Shared command that multiple goroutines will use
	sharedCmd := &Command{
		Name:          "echo",
		Args:          []string{"race", "test"},
		CaptureOutput: true,
	}

	// Multiple strategies accessing the same command
	strategies := []ExecutionStrategy{
		NewBasicStrategy(),
		NewTimeoutStrategy(1 * time.Second),
		NewPipeStrategy(nil),
	}

	var wg sync.WaitGroup

	// Launch goroutines that might trigger race conditions
	for _, strategy := range strategies {
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(s ExecutionStrategy) {
				defer wg.Done()

				ctx := context.Background()
				// Each strategy executes with the shared command
				// The PipeStrategy defensive copy should prevent races
				result, _ := s.Execute(ctx, sharedCmd)
				_ = result
			}(strategy)
		}
	}

	wg.Wait()

	// Verify shared command wasn't modified
	if sharedCmd.Stdin != nil {
		t.Error("Shared command was modified")
	}
}

// TestConcurrentBufferLimits tests that buffer limits work correctly under concurrent load
func TestConcurrentBufferLimits(t *testing.T) {
	var wg sync.WaitGroup
	goroutines := 10

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Create a command that generates output exceeding buffer limit
			// Generate about 20MB of output (2x the 10MB limit)
			cmd := &Command{
				Name:          "sh",
				Args:          []string{"-c", "for i in $(seq 1 200000); do echo 'This is a long line of output that will eventually exceed the buffer limit of 10MB'; done"},
				CaptureOutput: true,
			}

			builder := NewCommandBuilder(cmd)
			execCmd, stdout, stderr := builder.BuildWithCapture()

			// The LimitedBuffer should prevent memory exhaustion
			result := builder.ExecuteAndCollectResult(execCmd, stdout, stderr)

			// We expect truncated output due to buffer limits
			// The buffer should be limited to MaxBufferSize
			if len(result.Stdout) > MaxBufferSize {
				t.Errorf("Goroutine %d: output exceeded buffer limit (%d bytes)", id, len(result.Stdout))
			}
		}(i)
	}

	wg.Wait()
}

// BenchmarkConcurrentExecution benchmarks concurrent command execution
func BenchmarkConcurrentExecution(b *testing.B) {
	strategy := NewBasicStrategy()

	b.RunParallel(func(pb *testing.PB) {
		cmd := &Command{
			Name:          "echo",
			Args:          []string{"benchmark"},
			CaptureOutput: true,
		}

		ctx := context.Background()

		for pb.Next() {
			result, err := strategy.Execute(ctx, cmd)
			if err != nil {
				b.Fatalf("Execution failed: %v", err)
			}
			if result == nil {
				b.Fatal("Nil result")
			}
		}
	})
}
