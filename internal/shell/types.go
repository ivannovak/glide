package shell

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// ExecutionMode defines how a command should be executed
type ExecutionMode string

const (
	// ModePassthrough passes all I/O directly to/from the subprocess
	ModePassthrough ExecutionMode = "passthrough"
	// ModeCapture captures output for processing
	ModeCapture ExecutionMode = "capture"
	// ModeInteractive allocates a TTY for interactive commands
	ModeInteractive ExecutionMode = "interactive"
	// ModeBackground runs the command in background
	ModeBackground ExecutionMode = "background"
)

// Command represents a command to be executed
type Command struct {
	// Command and arguments
	Name string
	Args []string
	
	// Execution settings
	Mode        ExecutionMode
	WorkingDir  string
	Environment []string
	Timeout     time.Duration
	
	// I/O settings
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	
	// Options
	AllocateTTY   bool // Allocate pseudo-TTY for interactive commands
	InheritEnv    bool // Inherit parent process environment
	SignalForward bool // Forward signals to subprocess
	
	// Strategy settings
	UseStrategy   bool            // Use strategy pattern for execution
	CaptureOutput bool            // Capture stdout/stderr to Result
	StreamOutput  bool            // Stream output in real-time
	Options       CommandOptions  // Additional command options
}

// CommandOptions represents additional command execution options
type CommandOptions struct {
	CaptureOutput bool
	StreamOutput  bool
	Timeout       time.Duration
	OutputWriter  io.Writer
	ErrorWriter   io.Writer
}

// Result represents the result of command execution
type Result struct {
	ExitCode int
	Stdout   []byte
	Stderr   []byte
	Error    error
	Duration time.Duration
	Timeout  bool
}

// Options represents executor configuration
type Options struct {
	// Default timeout for all commands
	DefaultTimeout time.Duration
	
	// Buffer size for output capture
	BufferSize int
	
	// Whether to print commands before execution (debug mode)
	Verbose bool
	
	// Custom environment variables to add to all commands
	GlobalEnv []string
}

// NewCommand creates a new command with defaults
func NewCommand(name string, args ...string) *Command {
	return &Command{
		Name:          name,
		Args:          args,
		Mode:          ModeCapture,
		InheritEnv:    true,
		SignalForward: true,
	}
}

// NewPassthroughCommand creates a command that passes I/O directly
func NewPassthroughCommand(name string, args ...string) *Command {
	cmd := NewCommand(name, args...)
	cmd.Mode = ModePassthrough
	return cmd
}

// NewInteractiveCommand creates a command with TTY allocation
func NewInteractiveCommand(name string, args ...string) *Command {
	cmd := NewCommand(name, args...)
	cmd.Mode = ModeInteractive
	cmd.AllocateTTY = true
	return cmd
}

// WithTimeout sets the command timeout
func (c *Command) WithTimeout(timeout time.Duration) *Command {
	c.Timeout = timeout
	return c
}

// WithWorkingDir sets the working directory
func (c *Command) WithWorkingDir(dir string) *Command {
	c.WorkingDir = dir
	return c
}

// WithEnv adds environment variables
func (c *Command) WithEnv(env ...string) *Command {
	c.Environment = append(c.Environment, env...)
	return c
}

// String returns a string representation of the command
func (c *Command) String() string {
	if len(c.Args) > 0 {
		return c.Name + " " + joinArgs(c.Args)
	}
	return c.Name
}

// joinArgs joins arguments with proper quoting
func joinArgs(args []string) string {
	result := ""
	for i, arg := range args {
		if i > 0 {
			result += " "
		}
		// Quote if contains spaces
		if containsSpace(arg) {
			result += `"` + arg + `"`
		} else {
			result += arg
		}
	}
	return result
}

func containsSpace(s string) bool {
	for _, r := range s {
		if r == ' ' || r == '\t' {
			return true
		}
	}
	return false
}

// JoinArgs joins command arguments into a string, properly quoting if needed
// This is the public version for use by other packages
func JoinArgs(args []string) string {
	result := make([]string, len(args))
	for i, arg := range args {
		// Quote arguments that contain spaces or special characters
		if strings.ContainsAny(arg, " \t\n\"'\\$") {
			result[i] = fmt.Sprintf("'%s'", strings.ReplaceAll(arg, "'", "'\\''"))
		} else {
			result[i] = arg
		}
	}
	return strings.Join(result, " ")
}