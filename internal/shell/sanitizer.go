package shell

import (
	"fmt"
	"regexp"
	"strings"
)

// SanitizationMode determines how commands are sanitized
type SanitizationMode string

const (
	// ModeDisabled performs no sanitization (UNSAFE - use only for trusted commands)
	ModeDisabled SanitizationMode = "disabled"
	// ModeWarn validates and warns but allows execution
	ModeWarn SanitizationMode = "warn"
	// ModeStrict validates and blocks unsafe commands
	ModeStrict SanitizationMode = "strict"
	// ModeAllowlist only allows specifically whitelisted commands
	ModeAllowlist SanitizationMode = "allowlist"
)

// CommandSanitizer validates and sanitizes shell commands to prevent injection attacks
type CommandSanitizer interface {
	// Validate checks if a command is safe to execute
	Validate(command string, args []string) error
	// Sanitize returns a safe version of the command or an error
	Sanitize(command string, args []string) (string, error)
	// Mode returns the current sanitization mode
	Mode() SanitizationMode
}

// SanitizerConfig configures command sanitization behavior
type SanitizerConfig struct {
	// Mode determines the sanitization strategy
	Mode SanitizationMode
	// AllowedCommands is a list of allowed command names (for allowlist mode)
	AllowedCommands []string
	// AllowedPatterns is a list of regex patterns for allowed commands
	AllowedPatterns []*regexp.Regexp
	// BlockDangerousChars blocks commands containing shell metacharacters
	BlockDangerousChars bool
	// AllowPipes permits pipe operators in commands
	AllowPipes bool
	// AllowRedirects permits redirect operators in commands
	AllowRedirects bool
}

// DefaultConfig returns a secure default configuration
func DefaultConfig() *SanitizerConfig {
	return &SanitizerConfig{
		Mode:                ModeStrict,
		AllowedCommands:     []string{},
		AllowedPatterns:     []*regexp.Regexp{},
		BlockDangerousChars: true,
		AllowPipes:          false,
		AllowRedirects:      false,
	}
}

// AllowlistConfig returns a configuration for allowlist mode
func AllowlistConfig(allowedCommands ...string) *SanitizerConfig {
	return &SanitizerConfig{
		Mode:                ModeAllowlist,
		AllowedCommands:     allowedCommands,
		AllowedPatterns:     []*regexp.Regexp{},
		BlockDangerousChars: true,
		AllowPipes:          false,
		AllowRedirects:      false,
	}
}

// StrictSanitizer implements strict command validation
type StrictSanitizer struct {
	config *SanitizerConfig
}

// NewSanitizer creates a new command sanitizer with the given configuration
func NewSanitizer(config *SanitizerConfig) CommandSanitizer {
	if config == nil {
		config = DefaultConfig()
	}
	return &StrictSanitizer{config: config}
}

// NewAllowlistSanitizer creates a sanitizer that only allows specific commands
func NewAllowlistSanitizer(allowedCommands ...string) CommandSanitizer {
	return &StrictSanitizer{
		config: AllowlistConfig(allowedCommands...),
	}
}

// Mode returns the current sanitization mode
func (s *StrictSanitizer) Mode() SanitizationMode {
	return s.config.Mode
}

// Validate checks if a command and its arguments are safe to execute
func (s *StrictSanitizer) Validate(command string, args []string) error {
	if s.config.Mode == ModeDisabled {
		return nil
	}

	// Check for command injection patterns in the command itself
	if err := s.validateString(command, "command"); err != nil {
		return err
	}

	// Check each argument for injection patterns
	for i, arg := range args {
		if err := s.validateString(arg, fmt.Sprintf("argument %d", i+1)); err != nil {
			return err
		}
	}

	// Check allowlist if in allowlist mode
	if s.config.Mode == ModeAllowlist {
		if err := s.validateAllowlist(command); err != nil {
			return err
		}
	}

	return nil
}

// validateString checks a string for dangerous shell metacharacters and patterns
func (s *StrictSanitizer) validateString(str, context string) error {
	// Check for null bytes first (can be used to bypass checks)
	if strings.Contains(str, "\x00") {
		return fmt.Errorf("null byte detected in %s", context)
	}

	// Check for command chaining
	dangerousPatterns := []struct {
		pattern string
		desc    string
	}{
		{";", "command chaining (semicolon)"},
		{"&&", "command chaining (AND)"},
		{"||", "command chaining (OR)"},
		{"`", "command substitution (backtick)"},
		{"$(", "command substitution"},
		{"${", "variable expansion"},
		{"\n", "newline injection"},
		{"\r", "carriage return injection"},
		{"&", "background execution"},
	}

	// Only check pipes and redirects if they're not allowed
	if !s.config.AllowPipes {
		dangerousPatterns = append(dangerousPatterns, struct {
			pattern string
			desc    string
		}{"|", "pipe operator"})
	}

	if !s.config.AllowRedirects {
		dangerousPatterns = append(dangerousPatterns,
			struct {
				pattern string
				desc    string
			}{">", "output redirection"},
			struct {
				pattern string
				desc    string
			}{"<", "input redirection"},
		)
	}

	for _, p := range dangerousPatterns {
		if strings.Contains(str, p.pattern) {
			return fmt.Errorf("dangerous pattern detected in %s: %s (%s)", context, p.desc, p.pattern)
		}
	}

	// Check for path traversal
	if strings.Contains(str, "../") {
		return fmt.Errorf("path traversal detected in %s: ../", context)
	}

	return nil
}

// validateAllowlist checks if a command is in the allowlist
func (s *StrictSanitizer) validateAllowlist(command string) error {
	// Extract the base command (first word)
	baseCmd := strings.Fields(command)
	if len(baseCmd) == 0 {
		return fmt.Errorf("empty command")
	}

	// Check exact matches
	for _, allowed := range s.config.AllowedCommands {
		if baseCmd[0] == allowed {
			return nil
		}
	}

	// Check pattern matches
	for _, pattern := range s.config.AllowedPatterns {
		if pattern.MatchString(baseCmd[0]) {
			return nil
		}
	}

	return fmt.Errorf("command '%s' not in allowlist (allowed: %v)", baseCmd[0], s.config.AllowedCommands)
}

// Sanitize attempts to make a command safe or returns an error
func (s *StrictSanitizer) Sanitize(command string, args []string) (string, error) {
	if s.config.Mode == ModeDisabled {
		// In disabled mode, just join everything (UNSAFE)
		if len(args) == 0 {
			return command, nil
		}
		return command + " " + strings.Join(args, " "), nil
	}

	// First validate
	if err := s.Validate(command, args); err != nil {
		return "", fmt.Errorf("sanitization failed: %w", err)
	}

	// For strict mode, we've already validated, so we can construct the command
	// We don't actually modify the command, we just ensure it's safe
	if len(args) == 0 {
		return command, nil
	}

	// Escape arguments for shell safety
	escapedArgs := make([]string, len(args))
	for i, arg := range args {
		escapedArgs[i] = escapeShellArg(arg)
	}

	return command + " " + strings.Join(escapedArgs, " "), nil
}

// escapeShellArg escapes a string for safe use as a shell argument
// This provides defense-in-depth even after validation
func escapeShellArg(arg string) string {
	// If the argument is simple (alphanumeric, dash, underscore, dot, slash), don't quote
	if isSimpleArg(arg) {
		return arg
	}

	// Otherwise, single-quote it and escape any single quotes within
	// This is the safest approach for shell arguments
	escaped := strings.ReplaceAll(arg, "'", "'\"'\"'")
	return "'" + escaped + "'"
}

// isSimpleArg checks if an argument contains only safe characters
func isSimpleArg(arg string) bool {
	// Only allow alphanumeric, dash, underscore, dot, slash, colon
	for _, r := range arg {
		if !((r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '-' || r == '_' || r == '.' || r == '/' || r == ':') {
			return false
		}
	}
	return true
}

// ValidationError provides detailed information about validation failures
type ValidationError struct {
	Command string
	Args    []string
	Reason  string
	Pattern string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("command validation failed: %s (pattern: %s)", e.Reason, e.Pattern)
}

// NewValidationError creates a new validation error
func NewValidationError(command string, args []string, reason, pattern string) error {
	return &ValidationError{
		Command: command,
		Args:    args,
		Reason:  reason,
		Pattern: pattern,
	}
}
