package progress

import (
	"fmt"
	"io"
	"os"
	"sync"
)

// QuietLogger provides non-visual progress tracking for quiet mode
type QuietLogger struct {
	writer io.Writer
	mu     sync.Mutex
}

// NewQuietLogger creates a new quiet logger
func NewQuietLogger() *QuietLogger {
	return &QuietLogger{
		writer: os.Stdout,
	}
}

// Log logs a message in quiet mode
func (q *QuietLogger) Log(format string, args ...interface{}) {
	q.mu.Lock()
	defer q.mu.Unlock()

	fmt.Fprintf(q.writer, format+"\n", args...)
}

// QuietSpinner is a no-op spinner for quiet mode
type QuietSpinner struct {
	message string
	logger  *QuietLogger
}

// NewQuietSpinner creates a spinner that only logs in quiet mode
func NewQuietSpinner(message string) *QuietSpinner {
	return &QuietSpinner{
		message: message,
		logger:  NewQuietLogger(),
	}
}

// Start logs the start message
func (q *QuietSpinner) Start() {
	if q.logger != nil {
		q.logger.Log("Starting: %s", q.message)
	}
}

// Stop does nothing in quiet mode
func (q *QuietSpinner) Stop() {}

// Success logs success
func (q *QuietSpinner) Success(message string) {
	if q.logger != nil {
		q.logger.Log("✓ %s", message)
	}
}

// Error logs error
func (q *QuietSpinner) Error(message string) {
	if q.logger != nil {
		q.logger.Log("✗ %s", message)
	}
}

// Warning logs warning
func (q *QuietSpinner) Warning(message string) {
	if q.logger != nil {
		q.logger.Log("⚠ %s", message)
	}
}

// QuietBar is a no-op progress bar for quiet mode
type QuietBar struct {
	total   int
	current int
	message string
	logger  *QuietLogger
}

// NewQuietBar creates a progress bar that only logs in quiet mode
func NewQuietBar(total int, message string) *QuietBar {
	return &QuietBar{
		total:   total,
		current: 0,
		message: message,
		logger:  NewQuietLogger(),
	}
}

// Start logs the start
func (q *QuietBar) Start() {
	if q.logger != nil {
		q.logger.Log("Starting: %s (0/%d)", q.message, q.total)
	}
}

// Update logs progress at key milestones
func (q *QuietBar) Update(current int) {
	q.current = current

	// Log at 25%, 50%, 75%, and 100%
	percentage := float64(current) / float64(q.total) * 100

	if percentage == 25 || percentage == 50 || percentage == 75 {
		if q.logger != nil {
			q.logger.Log("Progress: %s (%.0f%%)", q.message, percentage)
		}
	}
}

// Increment increments by 1
func (q *QuietBar) Increment() {
	q.Update(q.current + 1)
}

// Finish logs completion
func (q *QuietBar) Finish() {
	if q.logger != nil {
		q.logger.Log("Completed: %s (%d/%d)", q.message, q.total, q.total)
	}
}

// Stop does nothing in quiet mode
func (q *QuietBar) Stop() {}

// Success logs success
func (q *QuietBar) Success(message string) {
	if q.logger != nil {
		q.logger.Log("✓ %s", message)
	}
}

// Error logs error
func (q *QuietBar) Error(message string) {
	if q.logger != nil {
		q.logger.Log("✗ %s", message)
	}
}

// Warning logs warning
func (q *QuietBar) Warning(message string) {
	if q.logger != nil {
		q.logger.Log("⚠ %s", message)
	}
}

// Factory functions that return appropriate implementations based on quiet mode

// CreateSpinner returns either a real spinner or quiet spinner based on global quiet mode
func CreateSpinner(message string) Indicator {
	if IsQuiet() {
		return NewQuietSpinner(message)
	}
	return NewSpinner(message)
}

// CreateBar returns either a real progress bar or quiet bar based on global quiet mode
func CreateBar(total int, message string) interface {
	Indicator
	Update(int)
	Increment()
	Finish()
} {
	if IsQuiet() {
		return NewQuietBar(total, message)
	}
	return NewBar(total, message)
}
