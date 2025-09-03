package progress

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

// SpinnerStyle defines the spinner animation style
type SpinnerStyle struct {
	Frames []string
	FPS    int
}

// Predefined spinner styles
var (
	SpinnerDots = SpinnerStyle{
		Frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		FPS:    10,
	}
	SpinnerArrows = SpinnerStyle{
		Frames: []string{"←", "↖", "↑", "↗", "→", "↘", "↓", "↙"},
		FPS:    10,
	}
	SpinnerBar = SpinnerStyle{
		Frames: []string{"|", "/", "-", "\\"},
		FPS:    10,
	}
	SpinnerCircle = SpinnerStyle{
		Frames: []string{"◐", "◓", "◑", "◒"},
		FPS:    8,
	}
)

// Spinner represents an indeterminate progress indicator
type Spinner struct {
	message   string
	style     SpinnerStyle
	options   *Options
	
	mu        sync.Mutex
	active    bool
	startTime time.Time
	stopChan  chan struct{}
	frame     int
	lastLine  string
}

// NewSpinner creates a new spinner with default style
func NewSpinner(message string) *Spinner {
	return NewSpinnerWithStyle(message, SpinnerDots)
}

// NewSpinnerWithStyle creates a new spinner with a specific style
func NewSpinnerWithStyle(message string, style SpinnerStyle) *Spinner {
	opts := DefaultOptions()
	return &Spinner{
		message: message,
		style:   style,
		options: opts,
	}
}

// NewSpinnerWithOptions creates a new spinner with custom options
func NewSpinnerWithOptions(message string, opts *Options) *Spinner {
	if opts == nil {
		opts = DefaultOptions()
	}
	return &Spinner{
		message: message,
		style:   SpinnerDots,
		options: opts,
	}
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	s.mu.Lock()
	if s.active || s.options.Quiet || !s.options.IsTTY {
		s.mu.Unlock()
		return
	}
	
	s.active = true
	s.startTime = time.Now()
	s.stopChan = make(chan struct{})
	s.frame = 0
	s.mu.Unlock()
	
	go s.animate()
}

// Stop stops the spinner and clears the line
func (s *Spinner) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if !s.active {
		return
	}
	
	s.active = false
	close(s.stopChan)
	
	// Clear the line
	if s.options.IsTTY && !s.options.Quiet {
		s.clearLine()
	}
}

// Success stops the spinner with a success message
func (s *Spinner) Success(message string) {
	s.Stop()
	if !s.options.Quiet {
		duration := s.getElapsedTime()
		if s.options.ShowElapsedTime && duration != "" {
			fmt.Fprintf(s.options.Writer, "%s %s %s\n", 
				color.GreenString("✓"),
				message,
				color.HiBlackString(duration))
		} else {
			fmt.Fprintf(s.options.Writer, "%s %s\n", 
				color.GreenString("✓"),
				message)
		}
	}
}

// Error stops the spinner with an error message
func (s *Spinner) Error(message string) {
	s.Stop()
	if !s.options.Quiet {
		duration := s.getElapsedTime()
		if s.options.ShowElapsedTime && duration != "" {
			fmt.Fprintf(s.options.Writer, "%s %s %s\n",
				color.RedString("✗"),
				message,
				color.HiBlackString(duration))
		} else {
			fmt.Fprintf(s.options.Writer, "%s %s\n",
				color.RedString("✗"),
				message)
		}
	}
}

// Warning stops the spinner with a warning message
func (s *Spinner) Warning(message string) {
	s.Stop()
	if !s.options.Quiet {
		duration := s.getElapsedTime()
		if s.options.ShowElapsedTime && duration != "" {
			fmt.Fprintf(s.options.Writer, "%s %s %s\n",
				color.YellowString("⚠"),
				message,
				color.HiBlackString(duration))
		} else {
			fmt.Fprintf(s.options.Writer, "%s %s\n",
				color.YellowString("⚠"),
				message)
		}
	}
}

// Update changes the spinner message while running
func (s *Spinner) Update(message string) {
	s.mu.Lock()
	s.message = message
	s.mu.Unlock()
}

// animate runs the spinner animation
func (s *Spinner) animate() {
	ticker := time.NewTicker(time.Second / time.Duration(s.style.FPS))
	defer ticker.Stop()
	
	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.mu.Lock()
			if s.active {
				s.render()
				s.frame = (s.frame + 1) % len(s.style.Frames)
			}
			s.mu.Unlock()
		}
	}
}

// render draws the current spinner frame
func (s *Spinner) render() {
	if s.options.Quiet || !s.options.IsTTY {
		return
	}
	
	// Clear previous line
	s.clearLine()
	
	// Build the new line
	frame := color.CyanString(s.style.Frames[s.frame])
	message := s.message
	
	// Add elapsed time if enabled
	elapsed := ""
	if s.options.ShowElapsedTime {
		duration := time.Since(s.startTime)
		if duration >= time.Second {
			elapsed = color.HiBlackString(" (%s)", formatDuration(duration))
		}
	}
	
	line := fmt.Sprintf("\r%s %s%s", frame, message, elapsed)
	s.lastLine = line
	
	fmt.Fprint(s.options.Writer, line)
}

// clearLine clears the current line
func (s *Spinner) clearLine() {
	if s.lastLine != "" {
		// Clear the entire line
		fmt.Fprintf(s.options.Writer, "\r%s\r", strings.Repeat(" ", len(s.lastLine)))
	}
}

// getElapsedTime returns formatted elapsed time
func (s *Spinner) getElapsedTime() string {
	if !s.options.ShowElapsedTime || s.startTime.IsZero() {
		return ""
	}
	
	duration := time.Since(s.startTime)
	if duration < s.options.MinDuration {
		return ""
	}
	
	return fmt.Sprintf("(%s)", formatDuration(duration))
}

// SpinnerGroup manages multiple spinners
type SpinnerGroup struct {
	spinners []*Spinner
	mu       sync.Mutex
}

// NewSpinnerGroup creates a new spinner group
func NewSpinnerGroup() *SpinnerGroup {
	return &SpinnerGroup{
		spinners: make([]*Spinner, 0),
	}
}

// Add adds a spinner to the group
func (sg *SpinnerGroup) Add(message string) *Spinner {
	sg.mu.Lock()
	defer sg.mu.Unlock()
	
	spinner := NewSpinner(message)
	sg.spinners = append(sg.spinners, spinner)
	return spinner
}

// StartAll starts all spinners in the group
func (sg *SpinnerGroup) StartAll() {
	sg.mu.Lock()
	defer sg.mu.Unlock()
	
	for _, spinner := range sg.spinners {
		spinner.Start()
	}
}

// StopAll stops all spinners in the group
func (sg *SpinnerGroup) StopAll() {
	sg.mu.Lock()
	defer sg.mu.Unlock()
	
	for _, spinner := range sg.spinners {
		spinner.Stop()
	}
}