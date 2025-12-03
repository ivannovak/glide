package shell

// This file is deprecated. All progress indicator functionality
// has been moved to the pkg/progress package.
//
// Please use:
//   import "github.com/glide-cli/glide/v3/pkg/progress"
//
// Examples:
//   spinner := progress.NewSpinner("Loading...")
//   spinner.Start()
//   // ... do work ...
//   spinner.Success("Done!")
//
// The new package provides:
//   - Spinners with elapsed time display
//   - Progress bars with ETA calculation
//   - Multi-progress support for concurrent operations
//   - Quiet mode for CI/non-TTY environments
//   - Unified API with better error handling
//
// Migration guide:
//   shell.NewProgressIndicator() -> progress.NewSpinner()
//   All method names remain the same (Start, Stop, Success, Error)
