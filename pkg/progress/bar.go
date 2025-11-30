package progress

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

// Bar represents a determinate progress bar
type Bar struct {
	total   int
	current int
	message string
	width   int
	options *Options

	mu         sync.Mutex
	active     bool
	startTime  time.Time
	lastUpdate time.Time
	lastLine   string

	// For throughput calculation
	startValue int
	samples    []throughputSample
}

type throughputSample struct {
	time  time.Time
	value int
}

// NewBar creates a new progress bar
func NewBar(total int, message string) *Bar {
	return NewBarWithOptions(total, message, nil)
}

// NewBarWithOptions creates a new progress bar with custom options
func NewBarWithOptions(total int, message string, opts *Options) *Bar {
	if opts == nil {
		opts = DefaultOptions()
	}

	// Determine bar width based on terminal size
	width := 40
	if opts.IsTTY {
		// Could use terminal size detection here
		width = 30
	}

	return &Bar{
		total:   total,
		current: 0,
		message: message,
		width:   width,
		options: opts,
		samples: make([]throughputSample, 0, 10),
	}
}

// Start begins rendering the progress bar
func (b *Bar) Start() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.active || b.options.Quiet || !b.options.IsTTY {
		return
	}

	b.active = true
	b.startTime = time.Now()
	b.lastUpdate = time.Now()
	b.startValue = b.current

	// Add initial sample
	b.samples = append(b.samples, throughputSample{
		time:  b.startTime,
		value: b.current,
	})

	b.render()
}

// Update updates the progress bar's current value
func (b *Bar) Update(current int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if current > b.total {
		current = b.total
	}
	if current < 0 {
		current = 0
	}

	b.current = current

	// Add throughput sample (max 10 samples for smoothing)
	now := time.Now()
	b.samples = append(b.samples, throughputSample{
		time:  now,
		value: current,
	})
	if len(b.samples) > 10 {
		b.samples = b.samples[1:]
	}

	// Only render if enough time has passed
	if b.active && now.Sub(b.lastUpdate) >= b.options.RefreshRate {
		b.render()
		b.lastUpdate = now
	}
}

// Increment increments the progress by 1
func (b *Bar) Increment() {
	b.Update(b.current + 1)
}

// IncrementBy increments the progress by a specific amount
func (b *Bar) IncrementBy(amount int) {
	b.Update(b.current + amount)
}

// SetTotal updates the total value
func (b *Bar) SetTotal(total int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.total = total
	if b.active {
		b.render()
	}
}

// Finish completes the progress bar
func (b *Bar) Finish() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.active {
		return
	}

	b.current = b.total
	b.render()
	b.active = false

	if b.options.IsTTY && !b.options.Quiet {
		// Safe to ignore: Newline after progress bar completion (cosmetic only)
		_, _ = fmt.Fprintln(b.options.Writer)
	}
}

// Success finishes with a success message
func (b *Bar) Success(message string) {
	b.Finish()
	if !b.options.Quiet {
		duration := b.getElapsedTime()
		if b.options.ShowElapsedTime && duration != "" {
			// Safe to ignore: Success message formatting (informational only)
			_, _ = fmt.Fprintf(b.options.Writer, "%s %s %s\n",
				color.GreenString("✓"),
				message,
				color.HiBlackString(duration))
		} else {
			// Safe to ignore: Success message formatting (informational only)
			_, _ = fmt.Fprintf(b.options.Writer, "%s %s\n",
				color.GreenString("✓"),
				message)
		}
	}
}

// Error finishes with an error message
func (b *Bar) Error(message string) {
	b.Stop()
	if !b.options.Quiet {
		// Safe to ignore: Error message formatting (informational only)
		_, _ = fmt.Fprintf(b.options.Writer, "%s %s\n",
			color.RedString("✗"),
			message)
	}
}

// Warning finishes with a warning message
func (b *Bar) Warning(message string) {
	b.Stop()
	if !b.options.Quiet {
		// Safe to ignore: Warning message formatting (informational only)
		_, _ = fmt.Fprintf(b.options.Writer, "%s %s\n",
			color.YellowString("⚠"),
			message)
	}
}

// Stop stops the progress bar without completing it
func (b *Bar) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.active {
		return
	}

	b.active = false
	if b.options.IsTTY && !b.options.Quiet {
		b.clearLine()
		// Safe to ignore: Newline after stopping progress bar (cosmetic only)
		_, _ = fmt.Fprintln(b.options.Writer)
	}
}

// render draws the progress bar
func (b *Bar) render() {
	if b.options.Quiet || !b.options.IsTTY {
		return
	}

	// Clear previous line
	b.clearLine()

	// Calculate percentage
	percentage := float64(b.current) / float64(b.total)
	if b.total == 0 {
		percentage = 0
	}

	// Build the bar
	filled := int(percentage * float64(b.width))
	if filled > b.width {
		filled = b.width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", b.width-filled)

	// Build the line components
	components := []string{
		b.message,
		fmt.Sprintf("[%s]", color.CyanString(bar)),
		fmt.Sprintf("%d/%d", b.current, b.total),
		fmt.Sprintf("(%.0f%%)", percentage*100),
	}

	// Add throughput if available
	if throughput := b.getThroughput(); throughput != "" {
		components = append(components, throughput)
	}

	// Add ETA if enabled
	if b.options.ShowETA {
		if eta := b.getETA(); eta != "" {
			components = append(components, eta)
		}
	}

	// Add elapsed time if enabled
	if b.options.ShowElapsedTime {
		if elapsed := b.getElapsedTimeFormatted(); elapsed != "" {
			components = append(components, elapsed)
		}
	}

	line := "\r" + strings.Join(components, " ")
	b.lastLine = line

	// Safe to ignore: Progress bar rendering (cosmetic display, doesn't affect operation)
	_, _ = fmt.Fprint(b.options.Writer, line)
}

// clearLine clears the current line
func (b *Bar) clearLine() {
	if b.lastLine != "" {
		fmt.Fprintf(b.options.Writer, "\r%s\r", strings.Repeat(" ", len(b.lastLine)))
	}
}

// getThroughput calculates current throughput
func (b *Bar) getThroughput() string {
	if len(b.samples) < 2 {
		return ""
	}

	// Use the last few samples for smoothing
	first := b.samples[0]
	last := b.samples[len(b.samples)-1]

	duration := last.time.Sub(first.time)
	if duration < time.Second {
		return ""
	}

	itemsDone := last.value - first.value
	itemsPerSecond := float64(itemsDone) / duration.Seconds()

	if itemsPerSecond >= 1 {
		return color.HiBlackString("%.1f/s", itemsPerSecond)
	} else if itemsPerSecond > 0 {
		return color.HiBlackString("%.2f/s", itemsPerSecond)
	}

	return ""
}

// getETA calculates estimated time to completion
func (b *Bar) getETA() string {
	if b.current == 0 || b.current >= b.total {
		return ""
	}

	// Calculate based on average speed
	elapsed := time.Since(b.startTime)
	if elapsed < time.Second {
		return ""
	}

	itemsDone := b.current - b.startValue
	if itemsDone <= 0 {
		return ""
	}

	itemsRemaining := b.total - b.current
	secondsPerItem := elapsed.Seconds() / float64(itemsDone)
	secondsRemaining := secondsPerItem * float64(itemsRemaining)

	if secondsRemaining < 1 {
		return ""
	}

	eta := time.Duration(secondsRemaining * float64(time.Second))
	return color.HiBlackString("ETA %s", formatDuration(eta))
}

// getElapsedTime returns the elapsed time since start
func (b *Bar) getElapsedTime() string {
	if b.startTime.IsZero() {
		return ""
	}

	duration := time.Since(b.startTime)
	if duration < b.options.MinDuration {
		return ""
	}

	return fmt.Sprintf("(%s)", formatDuration(duration))
}

// getElapsedTimeFormatted returns formatted elapsed time for display
func (b *Bar) getElapsedTimeFormatted() string {
	if b.startTime.IsZero() {
		return ""
	}

	duration := time.Since(b.startTime)
	if duration < time.Second {
		return ""
	}

	return color.HiBlackString("[%s]", formatDuration(duration))
}

// BarGroup manages multiple progress bars
type BarGroup struct {
	bars []*Bar
	mu   sync.Mutex
}

// NewBarGroup creates a new bar group
func NewBarGroup() *BarGroup {
	return &BarGroup{
		bars: make([]*Bar, 0),
	}
}

// Add adds a progress bar to the group
func (bg *BarGroup) Add(total int, message string) *Bar {
	bg.mu.Lock()
	defer bg.mu.Unlock()

	bar := NewBar(total, message)
	bg.bars = append(bg.bars, bar)
	return bar
}

// StartAll starts all bars in the group
func (bg *BarGroup) StartAll() {
	bg.mu.Lock()
	defer bg.mu.Unlock()

	for _, bar := range bg.bars {
		bar.Start()
	}
}

// FinishAll finishes all bars in the group
func (bg *BarGroup) FinishAll() {
	bg.mu.Lock()
	defer bg.mu.Unlock()

	for _, bar := range bg.bars {
		bar.Finish()
	}
}
