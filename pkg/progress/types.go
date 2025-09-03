package progress

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Options configures progress indicators
type Options struct {
	// Output writer (default: os.Stderr)
	Writer io.Writer
	// Whether to show elapsed time
	ShowElapsedTime bool
	// Whether to show ETA
	ShowETA bool
	// Update frequency
	RefreshRate time.Duration
	// Minimum duration before showing progress
	MinDuration time.Duration
	// Whether we're in a TTY
	IsTTY bool
	// Whether quiet mode is enabled
	Quiet bool
}

// DefaultOptions returns default options
func DefaultOptions() *Options {
	return &Options{
		Writer:          os.Stderr,
		ShowElapsedTime: true,
		ShowETA:         true,
		RefreshRate:     100 * time.Millisecond,
		MinDuration:     100 * time.Millisecond,
		IsTTY:           checkTTY(),
		Quiet:           isQuietMode(),
	}
}

// globalOptions holds global progress settings
var (
	globalOptions = DefaultOptions()
	globalMu      sync.RWMutex
)

// SetQuiet sets global quiet mode
func SetQuiet(quiet bool) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalOptions.Quiet = quiet
}

// IsQuiet returns whether quiet mode is enabled
func IsQuiet() bool {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalOptions.Quiet
}

// SetWriter sets the global output writer
func SetWriter(w io.Writer) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalOptions.Writer = w
}

// checkTTY checks if we're in a TTY
func checkTTY() bool {
	if fileInfo, _ := os.Stderr.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		return true
	}
	return false
}

// isQuietMode checks for quiet mode from environment
func isQuietMode() bool {
	// Check for common quiet environment variables
	if os.Getenv("QUIET") == "1" || os.Getenv("QUIET") == "true" {
		return true
	}
	if os.Getenv("CI") == "true" || os.Getenv("CI") == "1" {
		return true
	}
	return false
}

// Indicator is the common interface for all progress indicators
type Indicator interface {
	Start()
	Stop()
	Success(message string)
	Error(message string)
	Warning(message string)
}

// formatDuration formats a duration for display
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return ""
	}
	
	if d < time.Minute {
		seconds := int(d.Seconds())
		if seconds == 1 {
			return "1s"
		}
		return fmt.Sprintf("%ds", seconds)
	}
	
	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		if seconds > 0 {
			return fmt.Sprintf("%dm %ds", minutes, seconds)
		}
		return fmt.Sprintf("%dm", minutes)
	}
	
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if minutes > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dh", hours)
}