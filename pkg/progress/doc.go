// Package progress provides progress indicators for CLI operations.
//
// This package offers spinners, progress bars, and status indicators
// for showing operation progress in terminal environments.
//
// # Spinners
//
// Show a spinner for long-running operations:
//
//	// Simple spinner with automatic cleanup
//	err := progress.ShowSpinner("Loading...", func() error {
//	    // Long operation
//	    return nil
//	})
//
//	// Spinner with timeout
//	err := progress.ShowSpinnerWithTimeout("Connecting...", 30*time.Second, func() error {
//	    return connect()
//	})
//
// # Manual Spinner Control
//
// For more control over the spinner:
//
//	spinner := progress.NewSpinner("Processing items")
//	spinner.Start()
//
//	for i, item := range items {
//	    spinner.UpdateMessage(fmt.Sprintf("Processing item %d/%d", i+1, len(items)))
//	    process(item)
//	}
//
//	if err != nil {
//	    spinner.Error("Processing failed")
//	} else {
//	    spinner.Success("Processing complete")
//	}
//
// # Progress Bars
//
// Show progress for operations with known total:
//
//	bar := progress.NewBar(totalItems)
//	for i := 0; i < totalItems; i++ {
//	    processItem(i)
//	    bar.Increment()
//	}
//	bar.Finish()
//
// # Non-TTY Handling
//
// Progress indicators gracefully degrade in non-TTY environments:
//   - Spinners show start/end messages only
//   - Progress bars show percentage updates
package progress
