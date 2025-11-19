package docker

import (
	"errors"
	"fmt"
	"strings"
)

// Common Docker errors
var (
	ErrDockerNotInstalled = errors.New("docker is not installed or not in PATH")
	ErrDockerNotRunning   = errors.New("docker daemon is not running")
	ErrComposeNotFound    = errors.New("no docker-compose files found")
	ErrServiceNotFound    = errors.New("service not found")
	ErrContainerNotFound  = errors.New("container not found")
	ErrHealthCheckFailed  = errors.New("health check failed")
	ErrTimeout            = errors.New("operation timed out")
)

// DockerError represents a Docker-specific error
type DockerError struct {
	Op      string // Operation that failed
	Service string // Service name (if applicable)
	Err     error  // Underlying error
	Output  string // Command output (if available)
}

// Error implements the error interface
func (e *DockerError) Error() string {
	if e.Service != "" {
		return fmt.Sprintf("docker %s failed for service '%s': %v", e.Op, e.Service, e.Err)
	}
	return fmt.Sprintf("docker %s failed: %v", e.Op, e.Err)
}

// Unwrap returns the underlying error
func (e *DockerError) Unwrap() error {
	return e.Err
}

// IsRecoverable returns true if the error is recoverable
func (e *DockerError) IsRecoverable() bool {
	if e.Err == nil {
		return false
	}

	// Check for recoverable error patterns
	errStr := e.Err.Error()
	recoverable := []string{
		"container is already running",
		"container is already stopped",
		"no such container",
		"conflict",
	}

	for _, pattern := range recoverable {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return true
		}
	}

	return false
}

// ParseDockerError parses Docker command output to create a structured error
func ParseDockerError(op string, output string, err error) error {
	if err == nil {
		return nil
	}

	dockerErr := &DockerError{
		Op:     op,
		Err:    err,
		Output: output,
	}

	// Check for common error patterns
	outputLower := strings.ToLower(output)

	if strings.Contains(outputLower, "cannot connect to the docker daemon") ||
		strings.Contains(outputLower, "docker daemon is not running") {
		dockerErr.Err = ErrDockerNotRunning
		return dockerErr
	}

	if strings.Contains(outputLower, "no such service") {
		// Try to extract service name
		parts := strings.Split(output, "'")
		if len(parts) >= 2 {
			dockerErr.Service = parts[1]
		}
		dockerErr.Err = ErrServiceNotFound
		return dockerErr
	}

	if strings.Contains(outputLower, "no such container") {
		dockerErr.Err = ErrContainerNotFound
		return dockerErr
	}

	if strings.Contains(outputLower, "timeout") {
		dockerErr.Err = ErrTimeout
		return dockerErr
	}

	return dockerErr
}

// ErrorHandler provides methods for handling Docker errors
type ErrorHandler struct {
	verbose bool
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(verbose bool) *ErrorHandler {
	return &ErrorHandler{
		verbose: verbose,
	}
}

// Handle processes a Docker error and returns a user-friendly message
func (eh *ErrorHandler) Handle(err error) string {
	if err == nil {
		return ""
	}

	// Check if it's a DockerError
	var dockerErr *DockerError
	if errors.As(err, &dockerErr) {
		return eh.handleDockerError(dockerErr)
	}

	// Handle known errors
	switch {
	case errors.Is(err, ErrDockerNotInstalled):
		return "Docker is not installed. Please install Docker Desktop from https://www.docker.com/products/docker-desktop"
	case errors.Is(err, ErrDockerNotRunning):
		return "Docker is not running. Please start Docker Desktop and try again."
	case errors.Is(err, ErrComposeNotFound):
		return "No docker-compose.yml file found. This project may not use Docker."
	case errors.Is(err, ErrServiceNotFound):
		return "The requested service was not found. Run 'glide docker ps' to see available services."
	case errors.Is(err, ErrContainerNotFound):
		return "Container not found. The service may not be running."
	case errors.Is(err, ErrHealthCheckFailed):
		return "Container health check failed. Check logs with 'glide logs' for details."
	case errors.Is(err, ErrTimeout):
		return "Operation timed out. This may indicate a problem with Docker or the service."
	default:
		if eh.verbose {
			return fmt.Sprintf("Error: %v", err)
		}
		return "An error occurred. Use --verbose for more details."
	}
}

// handleDockerError handles DockerError specifically
func (eh *ErrorHandler) handleDockerError(err *DockerError) string {
	// Build base message
	msg := eh.Handle(err.Err)

	// Add operation context
	if err.Op != "" {
		msg = fmt.Sprintf("%s (during '%s')", msg, err.Op)
	}

	// Add service context
	if err.Service != "" {
		msg = fmt.Sprintf("%s for service '%s'", msg, err.Service)
	}

	// Add output if verbose
	if eh.verbose && err.Output != "" {
		msg = fmt.Sprintf("%s\n\nCommand output:\n%s", msg, err.Output)
	}

	// Add recovery suggestions
	if err.IsRecoverable() {
		msg += "\n\nThis error may be recoverable. Try running the command again."
	}

	return msg
}

// SuggestFix provides suggested fixes for common errors
func (eh *ErrorHandler) SuggestFix(err error) []string {
	suggestions := []string{}

	switch {
	case errors.Is(err, ErrDockerNotRunning):
		suggestions = append(suggestions,
			"Start Docker Desktop",
			"On macOS: Open Docker from Applications",
			"On Linux: Run 'sudo systemctl start docker'",
		)
	case errors.Is(err, ErrComposeNotFound):
		suggestions = append(suggestions,
			"Ensure you're in the correct directory",
			"Check if docker-compose.yml exists",
			"Run 'glide setup' if this is a new project",
		)
	case errors.Is(err, ErrServiceNotFound):
		suggestions = append(suggestions,
			"List available services: glide docker ps",
			"Check your docker-compose.yml file",
			"Ensure the service name is spelled correctly",
		)
	case errors.Is(err, ErrHealthCheckFailed):
		suggestions = append(suggestions,
			"Check service logs: glide logs [service]",
			"Restart the service: glide docker restart [service]",
			"Rebuild if needed: glide docker build",
		)
	}

	// Check for DockerError
	var dockerErr *DockerError
	if errors.As(err, &dockerErr) {
		if dockerErr.IsRecoverable() {
			suggestions = append(suggestions, "Try running the command again")
		}

		// Add operation-specific suggestions
		switch dockerErr.Op {
		case "up", "start":
			suggestions = append(suggestions,
				"Check if ports are already in use",
				"Remove old containers: glide docker down",
			)
		case "exec":
			suggestions = append(suggestions,
				"Ensure the container is running: glide docker ps",
				"Start containers: glide up",
			)
		}
	}

	return suggestions
}

// IsRetryable returns true if the error is worth retrying
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check for DockerError
	var dockerErr *DockerError
	if errors.As(err, &dockerErr) {
		return dockerErr.IsRecoverable()
	}

	// Check error message for retryable patterns
	errStr := strings.ToLower(err.Error())
	retryable := []string{
		"timeout",
		"connection refused",
		"temporary failure",
		"resource temporarily unavailable",
	}

	for _, pattern := range retryable {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}
