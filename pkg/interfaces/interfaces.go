package interfaces

import (
	"context"
	"io"
	"time"
)

// ShellExecutor defines the interface for executing shell commands
type ShellExecutor interface {
	Execute(ctx context.Context, cmd ShellCommand) (*ShellResult, error)
	ExecuteWithTimeout(cmd ShellCommand, timeout time.Duration) (*ShellResult, error)
	ExecuteWithProgress(cmd ShellCommand, message string) error
}

// ShellCommand represents a command to be executed
type ShellCommand interface {
	GetCommand() string
	GetArgs() []string
	GetOptions() ShellCommandOptions
	GetWorkingDir() string
	GetEnvironment() map[string]string
}

// ShellCommandOptions represents command execution options
type ShellCommandOptions interface {
	IsCaptureOutput() bool
	GetTimeout() time.Duration
	IsStreamOutput() bool
	GetOutputWriter() io.Writer
	GetErrorWriter() io.Writer
}

// ShellResult represents the result of a command execution
type ShellResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
}

// DockerResolver defines the interface for Docker compose file resolution
type DockerResolver interface {
	Resolve() error
	GetComposeFiles() []string
	BuildDockerCommand(args ...string) string
	GetBaseArgs() []string
}

// ContainerManager defines the interface for container management
type ContainerManager interface {
	Up() error
	Down() error
	Status() (string, error)
	Logs(service string, tail int) (string, error)
	Shell(service string) error
	ListContainers() ([]ContainerInfo, error)
}

// ContainerInfo represents information about a container
type ContainerInfo struct {
	ID      string
	Name    string
	Service string
	Status  string
	Ports   []string
}

// ConfigLoader defines the interface for loading configuration.
//
// Deprecated: This interface has a single implementation (internal/config.Loader)
// and is never mocked in tests. Consider using the concrete type directly or
// converting to a functional options pattern.
//
// This interface will be re-evaluated in v3.0.0. It may be removed if it
// continues to provide no abstraction value.
type ConfigLoader interface {
	Load(path string) error
	LoadDefault() error
	GetConfig() interface{}
	Save(path string) error
}

// ContextDetector defines the interface for detecting project context
type ContextDetector interface {
	Detect(workingDir string) (ProjectContext, error)
	DetectWithRoot(workingDir, projectRoot string) (ProjectContext, error)
}

// ProjectContext defines the interface for project context.
//
// Deprecated: This interface has a single implementation and is never mocked.
// It will be converted to a struct (*context.ProjectContext) in v3.0.0.
//
// This interface exists for historical reasons but adds no abstraction value.
// The single implementation is in internal/context/context.go and contains
// all the actual fields. This interface just wraps them with getters.
//
// Migration: Code should continue using this for now, but new code should
// prepare for direct struct usage. In v3.0.0, all GetX() methods will be
// removed in favor of direct field access.
type ProjectContext interface {
	GetWorkingDir() string
	GetProjectRoot() string
	GetDevelopmentMode() string
	GetLocation() string
	IsDockerRunning() bool
	GetComposeFiles() []string
	IsWorktree() bool
	GetWorktreeName() string
}

// StructuredOutput handles semantic output with different severity levels.
//
// This interface is for outputting structured messages with semantic meaning
// (info, success, error, warning). Each method indicates the message type.
//
// Example:
//
//	func processFile(out StructuredOutput, filename string) error {
//	    out.Info("Processing %s", filename)
//	    if err := process(filename); err != nil {
//	        out.Error("Failed: %v", err)
//	        return err
//	    }
//	    out.Success("Completed successfully")
//	    return nil
//	}
type StructuredOutput interface {
	// Display outputs structured data (tables, JSON, YAML, etc.)
	Display(data interface{}) error

	// Info outputs an informational message
	Info(format string, args ...interface{}) error

	// Success outputs a success message
	Success(format string, args ...interface{}) error

	// Error outputs an error message
	Error(format string, args ...interface{}) error

	// Warning outputs a warning message
	Warning(format string, args ...interface{}) error
}

// RawOutput handles unformatted text output.
//
// This interface is for outputting raw text without semantic meaning.
// Use this when you need direct control over output formatting.
//
// Example:
//
//	func printBanner(out RawOutput) error {
//	    return out.Printf("==== My Application ====\n")
//	}
type RawOutput interface {
	// Raw outputs raw text without formatting
	Raw(text string) error

	// Printf formats and outputs text
	Printf(format string, args ...interface{}) error

	// Println outputs text with a newline
	Println(args ...interface{}) error
}

// OutputManager is the composite interface for all output operations.
//
// This interface combines structured and raw output for convenience.
// Most code should depend on either StructuredOutput or RawOutput
// specifically, rather than the full OutputManager, to follow the
// Interface Segregation Principle.
//
// Thread Safety: Implementations must be safe for concurrent access.
type OutputManager interface {
	StructuredOutput
	RawOutput
}

// Formatter defines the interface for output formatting.
//
// Deprecated: This is a duplicate of pkg/output.Formatter. Use that instead.
// This interface will be removed in v3.0.0. The Formatter interface belongs
// in the output package where it is actually implemented.
//
// Migration: Change imports from:
//
//	"github.com/glide-cli/glide/v3/pkg/interfaces"
//
// To:
//
//	"github.com/glide-cli/glide/v3/pkg/output"
type Formatter interface {
	Display(data interface{}) error
	Info(format string, args ...interface{}) error
	Success(format string, args ...interface{}) error
	Error(format string, args ...interface{}) error
	Warning(format string, args ...interface{}) error
	Raw(text string) error
	SetWriter(w io.Writer)
}

// ProgressIndicator defines the interface for progress indication.
//
// Deprecated: This is similar to pkg/progress.Indicator. Use that instead.
// This interface will be removed in v3.0.0. The progress types belong
// in the progress package where they are actually implemented.
//
// Migration: Change imports from:
//
//	"github.com/glide-cli/glide/v3/pkg/interfaces"
//
// To:
//
//	"github.com/glide-cli/glide/v3/pkg/progress"
//
// Note: pkg/progress.Indicator has slightly different method signatures.
type ProgressIndicator interface {
	Start()
	Update(message string)
	Success()
	Fail()
	Stop()
}

// CommandBuilder defines the interface for building CLI commands
type CommandBuilder interface {
	Build() interface{} // Returns the root command
	RegisterCommand(name string, factory interface{}) error
	GetCommand(name string) (interface{}, error)
}

// Registry defines a generic registry interface
type Registry interface {
	Register(key string, value interface{}) error
	Get(key string) (interface{}, bool)
	List() []string
	Remove(key string) bool
}
