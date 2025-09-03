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

// ConfigLoader defines the interface for loading configuration
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

// ProjectContext defines the interface for project context
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

// OutputManager defines the interface for output management
type OutputManager interface {
	Display(data interface{}) error
	Info(format string, args ...interface{}) error
	Success(format string, args ...interface{}) error
	Error(format string, args ...interface{}) error
	Warning(format string, args ...interface{}) error
	Raw(text string) error
	Printf(format string, args ...interface{}) error
	Println(args ...interface{}) error
}

// Formatter defines the interface for output formatting
type Formatter interface {
	Display(data interface{}) error
	Info(format string, args ...interface{}) error
	Success(format string, args ...interface{}) error
	Error(format string, args ...interface{}) error
	Warning(format string, args ...interface{}) error
	Raw(text string) error
	SetWriter(w io.Writer)
}

// ProgressIndicator defines the interface for progress indication
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
