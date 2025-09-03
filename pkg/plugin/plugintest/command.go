package plugintest

import (
	"bytes"
	"strings"

	"github.com/spf13/cobra"
)

// CommandHelper provides utilities for testing Cobra commands
type CommandHelper struct {
	stdout *bytes.Buffer
	stderr *bytes.Buffer
	stdin  *bytes.Buffer
}

// NewCommandHelper creates a new command helper
func NewCommandHelper() *CommandHelper {
	return &CommandHelper{
		stdout: &bytes.Buffer{},
		stderr: &bytes.Buffer{},
		stdin:  &bytes.Buffer{},
	}
}

// NewTestCommand creates a test command with common setup
func NewTestCommand(use, short string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

// ExecuteCommand executes a command with arguments and captures output
func (h *CommandHelper) ExecuteCommand(cmd *cobra.Command, args ...string) (string, error) {
	h.stdout.Reset()
	h.stderr.Reset()

	cmd.SetOut(h.stdout)
	cmd.SetErr(h.stderr)
	cmd.SetIn(h.stdin)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return h.stdout.String(), err
}

// ExecuteCommandWithInput executes a command with stdin input
func (h *CommandHelper) ExecuteCommandWithInput(cmd *cobra.Command, input string, args ...string) (string, error) {
	h.stdin.Reset()
	h.stdin.WriteString(input)

	return h.ExecuteCommand(cmd, args...)
}

// GetStdout returns captured stdout
func (h *CommandHelper) GetStdout() string {
	return h.stdout.String()
}

// GetStderr returns captured stderr
func (h *CommandHelper) GetStderr() string {
	return h.stderr.String()
}

// Reset clears all buffers
func (h *CommandHelper) Reset() {
	h.stdout.Reset()
	h.stderr.Reset()
	h.stdin.Reset()
}

// CreateCommandTree creates a command tree for testing
func CreateCommandTree(commands ...*cobra.Command) *cobra.Command {
	root := &cobra.Command{Use: "root"}
	for _, cmd := range commands {
		root.AddCommand(cmd)
	}
	return root
}

// FindCommand finds a command in the tree by path
func FindCommand(root *cobra.Command, path ...string) *cobra.Command {
	cmd, _, _ := root.Find(path)
	return cmd
}

// CommandExists checks if a command exists in the tree
func CommandExists(root *cobra.Command, path ...string) bool {
	cmd := FindCommand(root, path...)
	return cmd != nil
}

// GetCommandNames returns all command names at the current level
func GetCommandNames(cmd *cobra.Command) []string {
	var names []string
	for _, c := range cmd.Commands() {
		names = append(names, c.Name())
	}
	return names
}

// GetCommandAliases returns all aliases for a command
func GetCommandAliases(cmd *cobra.Command) []string {
	return cmd.Aliases
}

// HasFlag checks if a command has a specific flag
func HasFlag(cmd *cobra.Command, flagName string) bool {
	return cmd.Flags().Lookup(flagName) != nil
}

// GetFlagValue gets the value of a flag
func GetFlagValue(cmd *cobra.Command, flagName string) (string, error) {
	flag := cmd.Flags().Lookup(flagName)
	if flag == nil {
		return "", nil
	}
	return flag.Value.String(), nil
}

// CommandBuilder provides a fluent interface for building test commands
type CommandBuilder struct {
	cmd *cobra.Command
}

// NewCommandBuilder creates a new command builder
func NewCommandBuilder(use string) *CommandBuilder {
	return &CommandBuilder{
		cmd: &cobra.Command{Use: use},
	}
}

// WithShort sets the short description
func (b *CommandBuilder) WithShort(short string) *CommandBuilder {
	b.cmd.Short = short
	return b
}

// WithLong sets the long description
func (b *CommandBuilder) WithLong(long string) *CommandBuilder {
	b.cmd.Long = long
	return b
}

// WithRunE sets the RunE function
func (b *CommandBuilder) WithRunE(runE func(cmd *cobra.Command, args []string) error) *CommandBuilder {
	b.cmd.RunE = runE
	return b
}

// WithRun sets the Run function
func (b *CommandBuilder) WithRun(run func(cmd *cobra.Command, args []string)) *CommandBuilder {
	b.cmd.Run = run
	return b
}

// WithAliases sets command aliases
func (b *CommandBuilder) WithAliases(aliases ...string) *CommandBuilder {
	b.cmd.Aliases = aliases
	return b
}

// WithFlag adds a string flag
func (b *CommandBuilder) WithFlag(name, shorthand, value, usage string) *CommandBuilder {
	b.cmd.Flags().StringP(name, shorthand, value, usage)
	return b
}

// WithBoolFlag adds a boolean flag
func (b *CommandBuilder) WithBoolFlag(name, shorthand string, value bool, usage string) *CommandBuilder {
	b.cmd.Flags().BoolP(name, shorthand, value, usage)
	return b
}

// WithSubCommand adds a subcommand
func (b *CommandBuilder) WithSubCommand(cmd *cobra.Command) *CommandBuilder {
	b.cmd.AddCommand(cmd)
	return b
}

// Build returns the built command
func (b *CommandBuilder) Build() *cobra.Command {
	return b.cmd
}

// AssertOutputContains checks if output contains expected string
func AssertOutputContains(output, expected string) bool {
	return strings.Contains(output, expected)
}

// AssertOutputNotContains checks if output does not contain string
func AssertOutputNotContains(output, unexpected string) bool {
	return !strings.Contains(output, unexpected)
}

// AssertOutputLines checks if output has expected number of lines
func AssertOutputLines(output string, expectedLines int) bool {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	return len(lines) == expectedLines
}
