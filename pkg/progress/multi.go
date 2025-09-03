package progress

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

// Multi manages multiple progress indicators simultaneously
type Multi struct {
	items   []multiItem
	options *Options
	
	mu       sync.Mutex
	active   bool
	stopChan chan struct{}
	writer   io.Writer
	
	// Terminal management
	lines    int
	lastDraw time.Time
}

type multiItem struct {
	typ      string // "spinner" or "bar"
	spinner  *Spinner
	bar      *Bar
	line     int
}

// NewMulti creates a new multi-progress manager
func NewMulti() *Multi {
	return NewMultiWithOptions(nil)
}

// NewMultiWithOptions creates a new multi-progress manager with options
func NewMultiWithOptions(opts *Options) *Multi {
	if opts == nil {
		opts = DefaultOptions()
	}
	
	return &Multi{
		items:   make([]multiItem, 0),
		options: opts,
		writer:  opts.Writer,
	}
}

// AddSpinner adds a spinner to the multi-progress
func (m *Multi) AddSpinner(message string) *Spinner {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Create spinner with custom writer that doesn't output directly
	spinner := NewSpinner(message)
	spinner.options.Writer = io.Discard // We'll handle rendering
	
	item := multiItem{
		typ:     "spinner",
		spinner: spinner,
		line:    len(m.items),
	}
	
	m.items = append(m.items, item)
	
	if m.active {
		m.render()
	}
	
	return spinner
}

// AddBar adds a progress bar to the multi-progress
func (m *Multi) AddBar(total int, message string) *Bar {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Create bar with custom writer that doesn't output directly
	bar := NewBar(total, message)
	bar.options.Writer = io.Discard // We'll handle rendering
	
	item := multiItem{
		typ:  "bar",
		bar:  bar,
		line: len(m.items),
	}
	
	m.items = append(m.items, item)
	
	if m.active {
		m.render()
	}
	
	return bar
}

// Start begins rendering all progress indicators
func (m *Multi) Start() {
	m.mu.Lock()
	if m.active || m.options.Quiet || !m.options.IsTTY {
		m.mu.Unlock()
		return
	}
	
	m.active = true
	m.stopChan = make(chan struct{})
	m.lastDraw = time.Now()
	
	// Start all items
	for _, item := range m.items {
		if item.typ == "spinner" {
			item.spinner.Start()
		} else if item.typ == "bar" {
			item.bar.Start()
		}
	}
	
	// Initial render
	m.render()
	m.mu.Unlock()
	
	// Start render loop
	go m.renderLoop()
}

// Stop stops all progress indicators
func (m *Multi) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if !m.active {
		return
	}
	
	m.active = false
	close(m.stopChan)
	
	// Stop all items
	for _, item := range m.items {
		if item.typ == "spinner" {
			item.spinner.Stop()
		} else if item.typ == "bar" {
			item.bar.Stop()
		}
	}
	
	// Clear all lines
	if m.options.IsTTY && !m.options.Quiet {
		m.clearAll()
	}
}

// renderLoop continuously updates the display
func (m *Multi) renderLoop() {
	ticker := time.NewTicker(m.options.RefreshRate)
	defer ticker.Stop()
	
	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.mu.Lock()
			if m.active {
				m.render()
			}
			m.mu.Unlock()
		}
	}
}

// render updates the multi-progress display
func (m *Multi) render() {
	if m.options.Quiet || !m.options.IsTTY {
		return
	}
	
	// Move cursor to start position
	if m.lines > 0 {
		// Move up to the first line
		fmt.Fprintf(m.writer, "\033[%dA", m.lines)
	}
	
	// Render each item
	for i, item := range m.items {
		// Clear line
		fmt.Fprintf(m.writer, "\r\033[K")
		
		// Render item
		line := m.renderItem(item)
		fmt.Fprint(m.writer, line)
		
		// Move to next line (except for last item)
		if i < len(m.items)-1 {
			fmt.Fprintln(m.writer)
		}
	}
	
	// Update line count
	m.lines = len(m.items)
	if m.lines > 0 {
		m.lines-- // Don't count the last line since we don't add newline
	}
}

// renderItem renders a single item
func (m *Multi) renderItem(item multiItem) string {
	if item.typ == "spinner" {
		return m.renderSpinner(item.spinner)
	} else if item.typ == "bar" {
		return m.renderBar(item.bar)
	}
	return ""
}

// renderSpinner renders a spinner line
func (m *Multi) renderSpinner(s *Spinner) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if !s.active {
		return ""
	}
	
	frame := s.style.Frames[s.frame]
	s.frame = (s.frame + 1) % len(s.style.Frames)
	
	// Build line
	parts := []string{
		color.CyanString(frame),
		s.message,
	}
	
	// Add elapsed time
	if s.options.ShowElapsedTime && !s.startTime.IsZero() {
		duration := time.Since(s.startTime)
		if duration >= time.Second {
			parts = append(parts, color.HiBlackString("(%s)", formatDuration(duration)))
		}
	}
	
	return strings.Join(parts, " ")
}

// renderBar renders a progress bar line
func (m *Multi) renderBar(b *Bar) string {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	if !b.active {
		return ""
	}
	
	// Calculate percentage
	percentage := float64(b.current) / float64(b.total)
	if b.total == 0 {
		percentage = 0
	}
	
	// Build the bar (smaller for multi-progress)
	width := 20
	filled := int(percentage * float64(width))
	if filled > width {
		filled = width
	}
	
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	
	// Build the line
	parts := []string{
		b.message,
		fmt.Sprintf("[%s]", color.CyanString(bar)),
		fmt.Sprintf("%d/%d", b.current, b.total),
		fmt.Sprintf("(%.0f%%)", percentage*100),
	}
	
	// Add ETA if available
	if b.options.ShowETA && b.current > 0 && b.current < b.total {
		elapsed := time.Since(b.startTime)
		if elapsed >= time.Second {
			itemsDone := b.current - b.startValue
			if itemsDone > 0 {
				itemsRemaining := b.total - b.current
				secondsPerItem := elapsed.Seconds() / float64(itemsDone)
				secondsRemaining := secondsPerItem * float64(itemsRemaining)
				
				if secondsRemaining >= 1 {
					eta := time.Duration(secondsRemaining * float64(time.Second))
					parts = append(parts, color.HiBlackString("ETA %s", formatDuration(eta)))
				}
			}
		}
	}
	
	return strings.Join(parts, " ")
}

// clearAll clears all lines used by the multi-progress
func (m *Multi) clearAll() {
	// Move to start
	if m.lines > 0 {
		fmt.Fprintf(m.writer, "\033[%dA", m.lines)
	}
	
	// Clear each line
	for i := 0; i <= m.lines; i++ {
		fmt.Fprintf(m.writer, "\r\033[K")
		if i < m.lines {
			fmt.Fprintln(m.writer)
		}
	}
	
	// Move back to start
	if m.lines > 0 {
		fmt.Fprintf(m.writer, "\033[%dA", m.lines)
	}
}

// Complete marks all items as complete and shows summary
func (m *Multi) Complete() {
	m.Stop()
	
	if !m.options.Quiet {
		// Show completion messages for each item
		for _, item := range m.items {
			if item.typ == "spinner" {
				fmt.Fprintf(m.writer, "%s %s\n",
					color.GreenString("✓"),
					item.spinner.message)
			} else if item.typ == "bar" {
				fmt.Fprintf(m.writer, "%s %s (%d/%d)\n",
					color.GreenString("✓"),
					item.bar.message,
					item.bar.current,
					item.bar.total)
			}
		}
	}
}

// UpdateBar updates a specific bar in the multi-progress
func (m *Multi) UpdateBar(bar *Bar, current int) {
	bar.Update(current)
	
	// Trigger a render if active
	m.mu.Lock()
	if m.active {
		m.render()
	}
	m.mu.Unlock()
}

// UpdateSpinner updates a specific spinner message
func (m *Multi) UpdateSpinner(spinner *Spinner, message string) {
	spinner.Update(message)
	
	// Trigger a render if active
	m.mu.Lock()
	if m.active {
		m.render()
	}
	m.mu.Unlock()
}