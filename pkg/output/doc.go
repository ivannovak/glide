// Package output provides formatted output management for the Glide CLI.
//
// This package handles all CLI output through a unified Manager interface,
// supporting multiple output formats (table, JSON, YAML, plain) with
// consistent styling and color support.
//
// # Output Manager
//
// Create a manager for formatted output:
//
//	manager := output.NewManager(output.FormatTable, false, false, os.Stdout)
//
//	// Success messages (green)
//	manager.Success("Operation completed successfully")
//
//	// Error messages (red)
//	manager.Error("Failed to load configuration")
//
//	// Warning messages (yellow)
//	manager.Warning("Deprecated feature used")
//
//	// Info messages (default color)
//	manager.Info("Processing %d items...", count)
//
// # Output Formats
//
// Multiple output formats are supported for different use cases:
//
//	output.FormatTable  // Human-readable tables (default)
//	output.FormatJSON   // Machine-readable JSON
//	output.FormatYAML   // YAML format
//	output.FormatPlain  // Plain text without formatting
//
// Change formats dynamically:
//
//	manager.SetFormat(output.FormatJSON)
//
// # Table Output
//
// Print structured data as tables:
//
//	data := output.TableData{
//	    Headers: []string{"Name", "Status", "Size"},
//	    Rows: [][]string{
//	        {"file1.txt", "Ready", "1.2 MB"},
//	        {"file2.txt", "Pending", "3.4 MB"},
//	    },
//	}
//	manager.Print(data)
//
// # Color Support
//
// Colors are enabled by default for TTY output:
//
//	manager := output.NewManager(format, quiet, noColor, writer)
//	// noColor=true disables all color output
//
// Environment variable support:
//   - NO_COLOR: Disables colors when set
//   - TERM=dumb: Disables colors
//
// # Quiet Mode
//
// Suppress non-essential output:
//
//	manager := output.NewManager(format, true, false, writer)
//	manager.Info("This is suppressed in quiet mode")
//	manager.Error("Errors are still shown")
//
// # Progress Indicators
//
// Show progress for long operations:
//
//	progress := manager.StartProgress("Downloading...")
//	for i := 0; i < 100; i++ {
//	    progress.Update(i, 100)
//	}
//	progress.Complete()
package output
