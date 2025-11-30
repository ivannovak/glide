// Package logging provides structured logging for Glide applications.
//
// This package wraps the standard library's log/slog package, providing
// a consistent logging interface with configurable levels, formats, and
// output destinations.
//
// # Basic Usage
//
//	// Use the default logger
//	log := logging.Default()
//	log.Info("Application started", "version", "1.0.0")
//
//	// Create a custom logger
//	cfg := &logging.Config{
//	    Level:  slog.LevelDebug,
//	    Format: logging.FormatJSON,
//	    Output: os.Stderr,
//	}
//	log := logging.New(cfg)
//
// # Log Levels
//
// Standard slog levels are supported:
//
//   - Debug: Detailed debugging information
//
//   - Info: General operational information
//
//   - Warn: Warning conditions
//
//   - Error: Error conditions
//
//     log.Debug("Query executed", "sql", query, "duration", elapsed)
//     log.Info("User logged in", "user", username)
//     log.Warn("Disk space low", "available", available)
//     log.Error("Failed to connect", "error", err)
//
// # Output Formats
//
// Two output formats are available:
//
//	logging.FormatText  // Human-readable text format
//	logging.FormatJSON  // Structured JSON format for log aggregation
//
// # Context-Aware Logging
//
// Create loggers with bound attributes:
//
//	reqLog := log.With("request_id", reqID, "user", userID)
//	reqLog.Info("Processing request")
//	reqLog.Info("Request complete", "status", 200)
//
// # Environment Configuration
//
// Configure via environment variables:
//   - GLIDE_LOG_LEVEL: debug, info, warn, error
//   - GLIDE_LOG_FORMAT: text, json
//
// # Integration with Container
//
// The container automatically provides a configured logger:
//
//	c.Run(ctx, func(log *logging.Logger) error {
//	    log.Info("Container started")
//	    return nil
//	})
package logging
