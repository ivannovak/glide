package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ivannovak/glide/v2/internal/config"
	"github.com/ivannovak/glide/v2/internal/shell"
)

var (
	// yamlCommandSanitizer is the global sanitizer for YAML commands
	// Can be configured via environment variables or config file
	yamlCommandSanitizer shell.CommandSanitizer
)

func init() {
	// Initialize sanitizer based on environment
	mode := os.Getenv("GLIDE_YAML_SANITIZE_MODE")

	switch strings.ToLower(mode) {
	case "disabled", "off":
		yamlCommandSanitizer = shell.NewSanitizer(&shell.SanitizerConfig{
			Mode: shell.ModeDisabled,
		})
	case "warn":
		yamlCommandSanitizer = shell.NewSanitizer(&shell.SanitizerConfig{
			Mode: shell.ModeWarn,
		})
	case "strict", "":
		// Default to strict mode for security
		yamlCommandSanitizer = shell.NewSanitizer(shell.DefaultConfig())
	default:
		// Unknown mode, default to strict
		fmt.Fprintf(os.Stderr, "Warning: Unknown GLIDE_YAML_SANITIZE_MODE '%s', using 'strict'\n", mode)
		yamlCommandSanitizer = shell.NewSanitizer(shell.DefaultConfig())
	}
}

// ExecuteYAMLCommand runs a YAML-defined command
func ExecuteYAMLCommand(cmdStr string, args []string) error {
	// Validate command before expansion (check command string itself)
	if err := yamlCommandSanitizer.Validate(cmdStr, []string{}); err != nil {
		return fmt.Errorf("YAML command validation failed: %w\n\nTo disable sanitization (UNSAFE): export GLIDE_YAML_SANITIZE_MODE=disabled", err)
	}

	// Validate arguments before expansion
	if err := yamlCommandSanitizer.Validate("", args); err != nil {
		return fmt.Errorf("YAML command arguments validation failed: %w\n\nTo disable sanitization (UNSAFE): export GLIDE_YAML_SANITIZE_MODE=disabled", err)
	}

	// Expand parameters
	expanded := config.ExpandCommand(cmdStr, args)

	// Validate expanded command as final check
	// This catches injection attempts that might occur during expansion
	if err := yamlCommandSanitizer.Validate(expanded, []string{}); err != nil {
		return fmt.Errorf("expanded YAML command validation failed: %w\n\nCommand after expansion: %s\n\nTo disable sanitization (UNSAFE): export GLIDE_YAML_SANITIZE_MODE=disabled", err, expanded)
	}

	// Execute as a shell script
	// This properly handles:
	// - Single commands
	// - Multi-line scripts
	// - Pipes and redirects (if allowed by sanitizer)
	// - Control structures (if allowed by sanitizer)
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

// SetYAMLCommandSanitizer allows overriding the global sanitizer (for testing)
func SetYAMLCommandSanitizer(sanitizer shell.CommandSanitizer) {
	yamlCommandSanitizer = sanitizer
}
