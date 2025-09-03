package output

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// Spinner represents a loading spinner
type Spinner struct {
	frames  []string
	message string
	stop    chan bool
	stopped bool
	mu      sync.Mutex
	writer  io.Writer
}

// Common spinner styles
var (
	SpinnerDots = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	SpinnerLine = []string{"-", "\\", "|", "/"}
	SpinnerArrow = []string{"←", "↖", "↑", "↗", "→", "↘", "↓", "↙"}
	SpinnerSimple = []string{".", "..", "...", "....", ".....", "......"}
	
	// ASCII fallback
	SpinnerASCII = []string{"-", "\\", "|", "/"}
)

// NewSpinner creates a new spinner
func NewSpinner(message string) *Spinner {
	// Choose spinner based on terminal capabilities
	frames := SpinnerDots
	if os.Getenv("GLIDE_ASCII_ICONS") != "" || os.Getenv("TERM") == "dumb" {
		frames = SpinnerASCII
	}
	
	return &Spinner{
		frames:  frames,
		message: message,
		stop:    make(chan bool),
		writer:  os.Stderr, // Use stderr to not interfere with stdout
	}
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	s.mu.Lock()
	if s.stopped {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		
		i := 0
		for {
			select {
			case <-s.stop:
				// Clear the spinner line
				fmt.Fprintf(s.writer, "\r%s\r", strings.Repeat(" ", len(s.frames[i])+len(s.message)+2))
				return
			case <-ticker.C:
				frame := s.frames[i%len(s.frames)]
				fmt.Fprintf(s.writer, "\r%s %s", InfoText("%s", frame), s.message)
				i++
			}
		}
	}()
}

// Stop halts the spinner
func (s *Spinner) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if !s.stopped {
		close(s.stop)
		s.stopped = true
	}
}

// UpdateMessage changes the spinner message
func (s *Spinner) UpdateMessage(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.message = message
}

// ProgressBar represents a progress bar
type ProgressBar struct {
	total    int
	current  int
	width    int
	message  string
	mu       sync.Mutex
	writer   io.Writer
	lastDraw time.Time
}

// NewProgressBar creates a new progress bar
func NewProgressBar(total int, message string) *ProgressBar {
	return &ProgressBar{
		total:   total,
		current: 0,
		width:   40,
		message: message,
		writer:  os.Stderr,
	}
}

// Update updates the progress bar
func (p *ProgressBar) Update(current int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	p.current = current
	
	// Throttle updates to max 10 per second
	if time.Since(p.lastDraw) < 100*time.Millisecond {
		return
	}
	p.lastDraw = time.Now()
	
	p.draw()
}

// Increment increases the progress by 1
func (p *ProgressBar) Increment() {
	p.Update(p.current + 1)
}

// draw renders the progress bar
func (p *ProgressBar) draw() {
	if p.total <= 0 {
		return
	}
	
	percent := float64(p.current) / float64(p.total)
	filled := int(percent * float64(p.width))
	
	// Build the bar
	bar := strings.Builder{}
	bar.WriteString("[")
	
	for i := 0; i < p.width; i++ {
		if i < filled {
			bar.WriteString("=")
		} else if i == filled {
			bar.WriteString(">")
		} else {
			bar.WriteString(" ")
		}
	}
	
	bar.WriteString("]")
	
	// Format the output
	output := fmt.Sprintf("\r%s %s %3d%% (%d/%d)",
		p.message,
		bar.String(),
		int(percent*100),
		p.current,
		p.total,
	)
	
	fmt.Fprint(p.writer, output)
	
	// Add newline when complete
	if p.current >= p.total {
		fmt.Fprintln(p.writer)
	}
}

// Clear clears the progress bar from the terminal
func (p *ProgressBar) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// Clear the line
	fmt.Fprintf(p.writer, "\r%s\r", strings.Repeat(" ", p.width+len(p.message)+20))
}

// Finish completes the progress bar
func (p *ProgressBar) Finish() {
	p.Update(p.total)
}

// IndeterminateProgress shows progress when total is unknown
type IndeterminateProgress struct {
	message string
	stop    chan bool
	stopped bool
	mu      sync.Mutex
	writer  io.Writer
}

// NewIndeterminateProgress creates a new indeterminate progress indicator
func NewIndeterminateProgress(message string) *IndeterminateProgress {
	return &IndeterminateProgress{
		message: message,
		stop:    make(chan bool),
		writer:  os.Stderr,
	}
}

// Start begins the indeterminate progress animation
func (p *IndeterminateProgress) Start() {
	p.mu.Lock()
	if p.stopped {
		p.mu.Unlock()
		return
	}
	p.mu.Unlock()

	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		
		width := 40
		position := 0
		direction := 1
		
		for {
			select {
			case <-p.stop:
				// Clear the progress line
				fmt.Fprintf(p.writer, "\r%s\r", strings.Repeat(" ", width+len(p.message)+10))
				return
			case <-ticker.C:
				// Build the progress indicator
				bar := strings.Builder{}
				bar.WriteString("[")
				
				for i := 0; i < width; i++ {
					if i >= position && i < position+5 {
						bar.WriteString("=")
					} else {
						bar.WriteString(" ")
					}
				}
				
				bar.WriteString("]")
				
				fmt.Fprintf(p.writer, "\r%s %s", p.message, bar.String())
				
				// Update position
				position += direction
				if position >= width-5 || position <= 0 {
					direction = -direction
				}
			}
		}
	}()
}

// Stop halts the progress indicator
func (p *IndeterminateProgress) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if !p.stopped {
		close(p.stop)
		p.stopped = true
	}
}

// Helper functions for common progress scenarios