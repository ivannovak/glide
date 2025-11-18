package cli

import (
	"os"
	"os/exec"

	"github.com/ivannovak/glide/internal/config"
)

// ExecuteYAMLCommand runs a YAML-defined command
func ExecuteYAMLCommand(cmdStr string, args []string) error {
	// Expand parameters
	expanded := config.ExpandCommand(cmdStr, args)

	// Execute as a shell script
	// This properly handles:
	// - Single commands
	// - Multi-line scripts
	// - Pipes and redirects
	// - Control structures (if/then/else, loops, etc.)
	// - Shell built-ins and functions
	return executeShellCommand(expanded)
}

// executeShellCommand runs a command through the shell
func executeShellCommand(cmdStr string) error {
	// Use sh -c to handle pipes, redirects, and other shell features
	cmd := exec.Command("sh", "-c", cmdStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Set environment to include current environment
	cmd.Env = os.Environ()

	return cmd.Run()
}