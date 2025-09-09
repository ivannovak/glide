package plugintest

import (
	"bytes"
	"io"
	"testing"

	"github.com/ivannovak/glide/pkg/plugin"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHarness provides a test environment for plugin testing
type TestHarness struct {
	t        *testing.T
	Registry *plugin.Registry
	RootCmd  *cobra.Command
	Config   map[string]interface{}
	stdout   *bytes.Buffer
	stderr   *bytes.Buffer
}

// NewTestHarness creates a new test harness
func NewTestHarness(t *testing.T) *TestHarness {
	return &TestHarness{
		t:        t,
		Registry: plugin.NewRegistry(),
		RootCmd:  &cobra.Command{Use: "test"},
		Config:   make(map[string]interface{}),
		stdout:   &bytes.Buffer{},
		stderr:   &bytes.Buffer{},
	}
}

// RegisterPlugin registers a plugin in the test harness
func (h *TestHarness) RegisterPlugin(p plugin.Plugin) error {
	err := h.Registry.RegisterPlugin(p)
	if err != nil {
		return err
	}

	// Configure with test config
	if err := p.Configure(h.Config); err != nil {
		return err
	}

	// Register commands
	return p.Register(h.RootCmd)
}

// WithConfig sets the configuration for the harness
func (h *TestHarness) WithConfig(config map[string]interface{}) *TestHarness {
	h.Config = config
	h.Registry.SetConfig(config)
	return h
}

// ExecuteCommand executes a command and captures output
func (h *TestHarness) ExecuteCommand(args ...string) (string, error) {
	h.stdout.Reset()
	h.stderr.Reset()

	h.RootCmd.SetOut(h.stdout)
	h.RootCmd.SetErr(h.stderr)
	h.RootCmd.SetArgs(args)

	err := h.RootCmd.Execute()
	return h.stdout.String(), err
}

// ExecuteCommandWithInput executes a command with stdin input
func (h *TestHarness) ExecuteCommandWithInput(input string, args ...string) (string, error) {
	h.stdout.Reset()
	h.stderr.Reset()

	h.RootCmd.SetIn(bytes.NewBufferString(input))
	h.RootCmd.SetOut(h.stdout)
	h.RootCmd.SetErr(h.stderr)
	h.RootCmd.SetArgs(args)

	err := h.RootCmd.Execute()
	return h.stdout.String(), err
}

// GetStdout returns the captured stdout
func (h *TestHarness) GetStdout() string {
	return h.stdout.String()
}

// GetStderr returns the captured stderr
func (h *TestHarness) GetStderr() string {
	return h.stderr.String()
}

// AssertPluginRegistered verifies a plugin is registered
func (h *TestHarness) AssertPluginRegistered(name string) {
	p, exists := h.Registry.Get(name)
	assert.True(h.t, exists, "Plugin %s should be registered", name)
	assert.NotNil(h.t, p, "Plugin %s should not be nil", name)
}

// AssertPluginNotRegistered verifies a plugin is not registered
func (h *TestHarness) AssertPluginNotRegistered(name string) {
	_, exists := h.Registry.Get(name)
	assert.False(h.t, exists, "Plugin %s should not be registered", name)
}

// AssertCommandExists verifies a command exists in the command tree
func (h *TestHarness) AssertCommandExists(cmdPath ...string) {
	cmd, _, err := h.RootCmd.Find(cmdPath)
	require.NoError(h.t, err, "Command %v should be found", cmdPath)
	assert.NotNil(h.t, cmd, "Command %v should not be nil", cmdPath)
}

// AssertCommandNotExists verifies a command does not exist
func (h *TestHarness) AssertCommandNotExists(cmdPath ...string) {
	cmd, _, _ := h.RootCmd.Find(cmdPath)
	assert.Nil(h.t, cmd, "Command %v should not exist", cmdPath)
}

// LoadAllPlugins loads all registered plugins
func (h *TestHarness) LoadAllPlugins() error {
	return h.Registry.LoadAll(h.RootCmd)
}

// ListPlugins returns all registered plugins
func (h *TestHarness) ListPlugins() []plugin.Plugin {
	return h.Registry.List()
}

// Reset resets the harness to initial state
func (h *TestHarness) Reset() {
	h.Registry = plugin.NewRegistry()
	h.RootCmd = &cobra.Command{Use: "test"}
	h.Config = make(map[string]interface{})
	h.stdout.Reset()
	h.stderr.Reset()
}

// CaptureOutput captures stdout and stderr during function execution
func (h *TestHarness) CaptureOutput(fn func()) (stdout, stderr string) {
	oldStdout := h.RootCmd.OutOrStdout()
	oldStderr := h.RootCmd.ErrOrStderr()

	stdoutBuf := &bytes.Buffer{}
	stderrBuf := &bytes.Buffer{}

	h.RootCmd.SetOut(stdoutBuf)
	h.RootCmd.SetErr(stderrBuf)

	fn()

	h.RootCmd.SetOut(oldStdout)
	h.RootCmd.SetErr(oldStderr)

	return stdoutBuf.String(), stderrBuf.String()
}

// RunWithConfig runs a function with a specific configuration
func (h *TestHarness) RunWithConfig(config map[string]interface{}, fn func()) {
	oldConfig := h.Config
	h.WithConfig(config)
	fn()
	h.Config = oldConfig
}

// GetOutputWriter returns an io.Writer for capturing output
func (h *TestHarness) GetOutputWriter() io.Writer {
	return h.stdout
}
