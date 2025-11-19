package config

import (
	"reflect"
	"testing"
)

func TestParseCommands(t *testing.T) {
	tests := []struct {
		name     string
		input    CommandMap
		expected map[string]*Command
		wantErr  bool
	}{
		{
			name: "simple string command",
			input: CommandMap{
				"build": "docker build .",
			},
			expected: map[string]*Command{
				"build": {Cmd: "docker build ."},
			},
			wantErr: false,
		},
		{
			name: "structured command with all fields",
			input: CommandMap{
				"deploy": map[string]interface{}{
					"cmd":         "deploy.sh",
					"alias":       "d",
					"description": "Deploy app",
					"help":        "Detailed help",
					"category":    "deployment",
				},
			},
			expected: map[string]*Command{
				"deploy": {
					Cmd:         "deploy.sh",
					Alias:       "d",
					Description: "Deploy app",
					Help:        "Detailed help",
					Category:    "deployment",
				},
			},
			wantErr: false,
		},
		{
			name: "multi-line command",
			input: CommandMap{
				"setup": "line1\nline2\nline3",
			},
			expected: map[string]*Command{
				"setup": {Cmd: "line1\nline2\nline3"},
			},
			wantErr: false,
		},
		{
			name: "structured command missing cmd field",
			input: CommandMap{
				"bad": map[string]interface{}{
					"alias": "b",
				},
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "mixed simple and structured commands",
			input: CommandMap{
				"simple": "echo hello",
				"structured": map[string]interface{}{
					"cmd":   "echo world",
					"alias": "s",
				},
			},
			expected: map[string]*Command{
				"simple":     {Cmd: "echo hello"},
				"structured": {Cmd: "echo world", Alias: "s"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCommands(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCommands() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("ParseCommands() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestExpandCommand(t *testing.T) {
	tests := []struct {
		name     string
		cmd      string
		args     []string
		expected string
	}{
		{
			name:     "no parameters",
			cmd:      "echo hello",
			args:     []string{},
			expected: "echo hello",
		},
		{
			name:     "single parameter",
			cmd:      "echo $1",
			args:     []string{"world"},
			expected: "echo world",
		},
		{
			name:     "multiple parameters",
			cmd:      "cp $1 $2",
			args:     []string{"source.txt", "dest.txt"},
			expected: "cp source.txt dest.txt",
		},
		{
			name:     "all arguments with $@",
			cmd:      "echo $@",
			args:     []string{"one", "two", "three"},
			expected: "echo one two three",
		},
		{
			name:     "all arguments with $*",
			cmd:      "echo $*",
			args:     []string{"one", "two", "three"},
			expected: "echo one two three",
		},
		{
			name:     "mixed parameters",
			cmd:      "cmd $1 --flag $2 -- $@",
			args:     []string{"first", "second", "third", "fourth"},
			expected: "cmd first --flag second -- first second third fourth",
		},
		{
			name:     "missing parameters",
			cmd:      "echo $1 $2 $3",
			args:     []string{"one"},
			expected: "echo one $2 $3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpandCommand(tt.cmd, tt.args)
			if got != tt.expected {
				t.Errorf("ExpandCommand() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestValidateCommand(t *testing.T) {
	tests := []struct {
		name    string
		cmd     *Command
		wantErr bool
	}{
		{
			name:    "valid command",
			cmd:     &Command{Cmd: "echo hello"},
			wantErr: false,
		},
		{
			name:    "empty command",
			cmd:     &Command{Cmd: ""},
			wantErr: true,
		},
		{
			name:    "circular reference with alias",
			cmd:     &Command{Cmd: "glidetest", Alias: "test"},
			wantErr: true,
		},
		{
			name:    "no circular reference",
			cmd:     &Command{Cmd: "glideother", Alias: "test"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCommand(tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}