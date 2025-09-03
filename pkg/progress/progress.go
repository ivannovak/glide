// Package progress provides consistent progress indicators for CLI operations
package progress

import (
	"fmt"
	"time"
)

// Simple API for common use cases

// ShowSpinner shows a spinner for an operation and returns the result
func ShowSpinner(message string, operation func() error) error {
	spinner := NewSpinner(message)
	spinner.Start()

	err := operation()

	if err != nil {
		spinner.Error(message + " failed")
		return err
	}

	spinner.Success(message + " completed")
	return nil
}

// ShowSpinnerWithTimeout shows a spinner with a timeout
func ShowSpinnerWithTimeout(message string, timeout time.Duration, operation func() error) error {
	spinner := NewSpinner(message)
	spinner.Start()

	done := make(chan error, 1)
	go func() {
		done <- operation()
	}()

	select {
	case err := <-done:
		if err != nil {
			spinner.Error(message + " failed")
			return err
		}
		spinner.Success(message + " completed")
		return nil

	case <-time.After(timeout):
		spinner.Error(message + " timed out")
		return fmt.Errorf("operation timed out after %v", timeout)
	}
}

// ShowProgress shows a progress bar for items processing
func ShowProgress(total int, message string, processor func(i int) error) error {
	bar := NewBar(total, message)
	bar.Start()

	for i := 0; i < total; i++ {
		if err := processor(i); err != nil {
			bar.Error(fmt.Sprintf("%s failed at item %d", message, i+1))
			return err
		}
		bar.Update(i + 1)
	}

	bar.Success(message + " completed")
	return nil
}

// RunWithProgress runs multiple operations with a progress bar
func RunWithProgress(operations []Operation) error {
	if len(operations) == 0 {
		return nil
	}

	bar := NewBar(len(operations), "Running operations")
	bar.Start()

	for i, op := range operations {
		bar.Update(i)

		if err := op.Run(); err != nil {
			bar.Error(fmt.Sprintf("Operation '%s' failed", op.Name()))
			return err
		}
	}

	bar.Finish()
	bar.Success("All operations completed")
	return nil
}

// Operation represents an operation that can be run with progress
type Operation interface {
	Name() string
	Run() error
}

// SimpleOperation is a basic implementation of Operation
type SimpleOperation struct {
	name string
	fn   func() error
}

// NewOperation creates a new simple operation
func NewOperation(name string, fn func() error) Operation {
	return &SimpleOperation{
		name: name,
		fn:   fn,
	}
}

// Name returns the operation name
func (o *SimpleOperation) Name() string {
	return o.name
}

// Run executes the operation
func (o *SimpleOperation) Run() error {
	return o.fn()
}

// Concurrent runs multiple operations concurrently with progress
func Concurrent(operations []Operation) error {
	if len(operations) == 0 {
		return nil
	}

	multi := NewMulti()

	// Add a spinner for each operation
	spinners := make([]*Spinner, len(operations))
	for i, op := range operations {
		spinners[i] = multi.AddSpinner(op.Name())
	}

	multi.Start()
	defer multi.Stop()

	// Run operations concurrently
	errors := make(chan error, len(operations))
	for i, op := range operations {
		go func(idx int, operation Operation) {
			err := operation.Run()
			if err != nil {
				spinners[idx].Error(operation.Name() + " failed")
			} else {
				spinners[idx].Success(operation.Name() + " completed")
			}
			errors <- err
		}(i, op)
	}

	// Wait for all operations
	var firstError error
	for i := 0; i < len(operations); i++ {
		if err := <-errors; err != nil && firstError == nil {
			firstError = err
		}
	}

	multi.Complete()
	return firstError
}

// WithElapsedTime runs an operation and reports elapsed time
func WithElapsedTime(message string, operation func() error) error {
	start := time.Now()

	spinner := NewSpinner(message)
	spinner.Start()

	err := operation()
	elapsed := time.Since(start)

	if err != nil {
		spinner.Error(fmt.Sprintf("%s failed (%s)", message, formatDuration(elapsed)))
		return err
	}

	spinner.Success(fmt.Sprintf("%s completed (%s)", message, formatDuration(elapsed)))
	return nil
}

// Example usage patterns for documentation
func examples() {
	// Simple spinner
	_ = ShowSpinner("Loading configuration", func() error {
		time.Sleep(2 * time.Second)
		return nil
	})

	// Progress bar for items
	items := []string{"file1.txt", "file2.txt", "file3.txt"}
	_ = ShowProgress(len(items), "Processing files", func(i int) error {
		// Process items[i]
		time.Sleep(500 * time.Millisecond)
		return nil
	})

	// Multiple operations with progress
	ops := []Operation{
		NewOperation("Download dependencies", func() error {
			time.Sleep(2 * time.Second)
			return nil
		}),
		NewOperation("Build project", func() error {
			time.Sleep(3 * time.Second)
			return nil
		}),
		NewOperation("Run tests", func() error {
			time.Sleep(1 * time.Second)
			return nil
		}),
	}
	_ = RunWithProgress(ops)

	// Concurrent operations
	_ = Concurrent(ops)
}
