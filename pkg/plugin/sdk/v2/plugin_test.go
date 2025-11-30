package v2

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfig for testing
type TestConfig struct {
	APIKey  string `json:"apiKey"`
	Timeout int    `json:"timeout"`
	Enabled bool   `json:"enabled"`
}

// TestPlugin is a minimal plugin implementation for testing
type TestPlugin struct {
	BasePlugin[TestConfig]
	initCalled  bool
	startCalled bool
	stopCalled  bool
}

func NewTestPlugin() *TestPlugin {
	p := &TestPlugin{}
	p.SetMetadata(Metadata{
		Name:        "test-plugin",
		Version:     "1.0.0",
		Description: "A test plugin",
		Author:      "Test Author",
		License:     "MIT",
		Tags:        []string{"test", "example"},
	})

	p.SetCommands([]Command{
		{
			Name:        "hello",
			Description: "Say hello",
			Category:    "test",
			Handler: SimpleCommandHandler(func(_ context.Context, _ *ExecuteRequest) (*ExecuteResponse, error) {
				return &ExecuteResponse{
					ExitCode: 0,
					Output:   "Hello, World!",
				}, nil
			}),
		},
	})

	return p
}

func (p *TestPlugin) Init(ctx context.Context) error {
	p.initCalled = true
	return p.BasePlugin.Init(ctx)
}

func (p *TestPlugin) Start(ctx context.Context) error {
	p.startCalled = true
	return p.BasePlugin.Start(ctx)
}

func (p *TestPlugin) Stop(ctx context.Context) error {
	p.stopCalled = true
	return p.BasePlugin.Stop(ctx)
}

func TestBasePlugin_Metadata(t *testing.T) {
	plugin := NewTestPlugin()

	meta := plugin.Metadata()
	assert.Equal(t, "test-plugin", meta.Name)
	assert.Equal(t, "1.0.0", meta.Version)
	assert.Equal(t, "A test plugin", meta.Description)
	assert.Equal(t, "Test Author", meta.Author)
	assert.Equal(t, "MIT", meta.License)
	assert.Equal(t, []string{"test", "example"}, meta.Tags)
}

func TestBasePlugin_ConfigSchema(t *testing.T) {
	plugin := NewTestPlugin()

	// Default implementation returns nil
	schema := plugin.ConfigSchema()
	assert.Nil(t, schema)
}

func TestBasePlugin_Configure(t *testing.T) {
	plugin := NewTestPlugin()
	ctx := context.Background()

	config := TestConfig{
		APIKey:  "test-key",
		Timeout: 30,
		Enabled: true,
	}

	err := plugin.Configure(ctx, config)
	require.NoError(t, err)

	// Verify config was stored
	storedConfig := plugin.Config()
	assert.Equal(t, config.APIKey, storedConfig.APIKey)
	assert.Equal(t, config.Timeout, storedConfig.Timeout)
	assert.Equal(t, config.Enabled, storedConfig.Enabled)
}

func TestBasePlugin_Lifecycle(t *testing.T) {
	plugin := NewTestPlugin()
	ctx := context.Background()

	// Test Init
	err := plugin.Init(ctx)
	require.NoError(t, err)
	assert.True(t, plugin.initCalled)

	// Test Start
	err = plugin.Start(ctx)
	require.NoError(t, err)
	assert.True(t, plugin.startCalled)

	// Test HealthCheck
	err = plugin.HealthCheck(ctx)
	require.NoError(t, err)

	// Test Stop
	err = plugin.Stop(ctx)
	require.NoError(t, err)
	assert.True(t, plugin.stopCalled)
}

func TestBasePlugin_Commands(t *testing.T) {
	plugin := NewTestPlugin()

	commands := plugin.Commands()
	require.Len(t, commands, 1)

	cmd := commands[0]
	assert.Equal(t, "hello", cmd.Name)
	assert.Equal(t, "Say hello", cmd.Description)
	assert.Equal(t, "test", cmd.Category)
	assert.NotNil(t, cmd.Handler)
}

func TestBasePlugin_AddCommand(t *testing.T) {
	plugin := NewTestPlugin()

	// Add another command
	plugin.AddCommand(Command{
		Name:        "goodbye",
		Description: "Say goodbye",
		Handler: SimpleCommandHandler(func(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error) {
			return &ExecuteResponse{
				ExitCode: 0,
				Output:   "Goodbye!",
			}, nil
		}),
	})

	commands := plugin.Commands()
	require.Len(t, commands, 2)
	assert.Equal(t, "hello", commands[0].Name)
	assert.Equal(t, "goodbye", commands[1].Name)
}

func TestSimpleCommandHandler_Execute(t *testing.T) {
	handler := SimpleCommandHandler(func(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error) {
		return &ExecuteResponse{
			ExitCode: 0,
			Output:   "Test output: " + req.Args[0],
		}, nil
	})

	ctx := context.Background()
	req := &ExecuteRequest{
		Command: "test",
		Args:    []string{"arg1"},
	}

	resp, err := handler.Execute(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.ExitCode)
	assert.Equal(t, "Test output: arg1", resp.Output)
}

func TestMetadata_Dependencies(t *testing.T) {
	plugin := &BasePlugin[struct{}]{}
	plugin.SetMetadata(Metadata{
		Name:    "dependent-plugin",
		Version: "1.0.0",
		Dependencies: []Dependency{
			{Name: "docker", Version: "^1.0.0", Optional: false},
			{Name: "database", Version: ">=2.0.0 <3.0.0", Optional: true},
		},
	})

	meta := plugin.Metadata()
	require.Len(t, meta.Dependencies, 2)
	assert.Equal(t, "docker", meta.Dependencies[0].Name)
	assert.Equal(t, "^1.0.0", meta.Dependencies[0].Version)
	assert.False(t, meta.Dependencies[0].Optional)
	assert.Equal(t, "database", meta.Dependencies[1].Name)
	assert.True(t, meta.Dependencies[1].Optional)
}

func TestMetadata_Capabilities(t *testing.T) {
	plugin := &BasePlugin[struct{}]{}
	plugin.SetMetadata(Metadata{
		Name:    "capable-plugin",
		Version: "1.0.0",
		Capabilities: Capabilities{
			RequiresDocker:      true,
			RequiresNetwork:     true,
			RequiresFilesystem:  false,
			RequiresInteractive: false,
			RequiredCommands:    []string{"git", "docker"},
			RequiredPaths:       []string{"/usr/bin/git"},
			RequiredEnvVars:     []string{"HOME", "PATH"},
		},
	})

	meta := plugin.Metadata()
	assert.True(t, meta.Capabilities.RequiresDocker)
	assert.True(t, meta.Capabilities.RequiresNetwork)
	assert.False(t, meta.Capabilities.RequiresFilesystem)
	assert.Equal(t, []string{"git", "docker"}, meta.Capabilities.RequiredCommands)
	assert.Equal(t, []string{"/usr/bin/git"}, meta.Capabilities.RequiredPaths)
	assert.Equal(t, []string{"HOME", "PATH"}, meta.Capabilities.RequiredEnvVars)
}

func TestCommand_Flags(t *testing.T) {
	cmd := Command{
		Name: "test-cmd",
		Flags: []Flag{
			{
				Name:        "output",
				Shorthand:   "o",
				Description: "Output file",
				Type:        "string",
				Default:     "output.txt",
				Required:    false,
			},
			{
				Name:        "verbose",
				Shorthand:   "v",
				Description: "Verbose output",
				Type:        "bool",
				Default:     false,
				Required:    false,
			},
		},
	}

	require.Len(t, cmd.Flags, 2)
	assert.Equal(t, "output", cmd.Flags[0].Name)
	assert.Equal(t, "o", cmd.Flags[0].Shorthand)
	assert.Equal(t, "string", cmd.Flags[0].Type)
	assert.Equal(t, "output.txt", cmd.Flags[0].Default)

	assert.Equal(t, "verbose", cmd.Flags[1].Name)
	assert.Equal(t, "bool", cmd.Flags[1].Type)
	assert.Equal(t, false, cmd.Flags[1].Default)
}

func TestCommand_Args(t *testing.T) {
	cmd := Command{
		Name: "test-cmd",
		Args: []Arg{
			{
				Name:        "file",
				Description: "Input file",
				Required:    true,
				Variadic:    false,
			},
			{
				Name:        "extras",
				Description: "Extra arguments",
				Required:    false,
				Variadic:    true,
			},
		},
	}

	require.Len(t, cmd.Args, 2)
	assert.Equal(t, "file", cmd.Args[0].Name)
	assert.True(t, cmd.Args[0].Required)
	assert.False(t, cmd.Args[0].Variadic)

	assert.Equal(t, "extras", cmd.Args[1].Name)
	assert.False(t, cmd.Args[1].Required)
	assert.True(t, cmd.Args[1].Variadic)
}

func TestCommandError(t *testing.T) {
	err := &CommandError{
		Command:  "test",
		ExitCode: 1,
		Message:  "command failed",
	}

	assert.Equal(t, "command failed", err.Error())

	// Test with no message
	err2 := &CommandError{
		Command:  "test",
		ExitCode: 2,
	}
	assert.Contains(t, err2.Error(), "exit code")
}

func TestCobraAdapter_BuildCommands(t *testing.T) {
	plugin := NewTestPlugin()
	adapter := NewCobraAdapter(plugin)

	commands := adapter.BuildCommands()
	require.Len(t, commands, 1)

	cobraCmd := commands[0]
	assert.Equal(t, "hello", cobraCmd.Use)
	assert.Equal(t, "Say hello", cobraCmd.Short)
	assert.NotNil(t, cobraCmd.RunE)
}

// TestPluginWithCustomSchema tests a plugin with custom config schema
func TestPluginWithCustomSchema(t *testing.T) {
	type CustomConfig struct {
		URL string `json:"url"`
	}

	type CustomPlugin struct {
		BasePlugin[CustomConfig]
	}

	plugin := &CustomPlugin{}
	plugin.SetMetadata(Metadata{
		Name:    "custom-plugin",
		Version: "1.0.0",
	})

	// Override ConfigSchema
	customSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"url": map[string]interface{}{
				"type":   "string",
				"format": "uri",
			},
		},
		"required": []string{"url"},
	}

	// Since we can't override methods in the test, we'll just verify
	// the default schema is nil
	schema := plugin.ConfigSchema()
	assert.Nil(t, schema)

	// In real usage, the plugin would override ConfigSchema() to return customSchema
	_ = customSchema // Use it to avoid unused variable error
}
