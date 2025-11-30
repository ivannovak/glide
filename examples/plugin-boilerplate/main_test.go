package main

import (
	"context"
	"strings"
	"testing"

	sdk "github.com/ivannovak/glide/v3/pkg/plugin/sdk/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// MockStream implements the sdk.GlidePlugin_StartInteractiveServer interface for testing
type MockStream struct {
	grpc.ServerStream
	sentMessages []*sdk.StreamMessage
	recvMessages []*sdk.StreamMessage
	recvIndex    int
}

func NewMockStream() *MockStream {
	return &MockStream{
		sentMessages: make([]*sdk.StreamMessage, 0),
		recvMessages: make([]*sdk.StreamMessage, 0),
		recvIndex:    0,
	}
}

func (s *MockStream) Send(msg *sdk.StreamMessage) error {
	s.sentMessages = append(s.sentMessages, msg)
	return nil
}

func (s *MockStream) Recv() (*sdk.StreamMessage, error) {
	if s.recvIndex >= len(s.recvMessages) {
		return nil, nil
	}
	msg := s.recvMessages[s.recvIndex]
	s.recvIndex++
	return msg, nil
}

func (s *MockStream) AddRecvMessage(msg *sdk.StreamMessage) {
	s.recvMessages = append(s.recvMessages, msg)
}

func TestMyPlugin_GetMetadata(t *testing.T) {
	plugin := &MyPlugin{
		config: make(map[string]interface{}),
	}

	ctx := context.Background()
	metadata, err := plugin.GetMetadata(ctx, &sdk.Empty{})

	require.NoError(t, err, "GetMetadata should not return an error")
	assert.NotNil(t, metadata, "Metadata should not be nil")

	// Verify metadata fields
	assert.Equal(t, "myplugin", metadata.Name, "Plugin name should be myplugin")
	assert.Equal(t, "1.0.0", metadata.Version, "Version should be 1.0.0")
	assert.NotEmpty(t, metadata.Author, "Author should not be empty")
	assert.NotEmpty(t, metadata.Description, "Description should not be empty")
	assert.Equal(t, "v1.0.0", metadata.MinSdk, "MinSdk should be v1.0.0")
}

func TestMyPlugin_Configure(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]string
		wantErr bool
	}{
		{
			name:    "empty configuration",
			config:  map[string]string{},
			wantErr: false,
		},
		{
			name: "with configuration values",
			config: map[string]string{
				"api_key":  "test-key-123",
				"endpoint": "https://api.example.com",
				"timeout":  "30",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &MyPlugin{
				config: make(map[string]interface{}),
			}

			ctx := context.Background()
			req := &sdk.ConfigureRequest{
				Config: tt.config,
			}

			resp, err := plugin.Configure(ctx, req)

			if tt.wantErr {
				assert.Error(t, err, "Configure should return an error")
			} else {
				require.NoError(t, err, "Configure should not return an error")
				assert.NotNil(t, resp, "Response should not be nil")
				assert.True(t, resp.Success, "Configuration should be successful")
				assert.Contains(t, resp.Message, "configured successfully", "Message should indicate success")

				// Verify configuration was stored
				for k, v := range tt.config {
					assert.Equal(t, v, plugin.config[k], "Config value for %s should match", k)
				}
			}
		})
	}
}

func TestMyPlugin_ListCommands(t *testing.T) {
	plugin := &MyPlugin{
		config: make(map[string]interface{}),
	}

	ctx := context.Background()
	commandList, err := plugin.ListCommands(ctx, &sdk.Empty{})

	require.NoError(t, err, "ListCommands should not return an error")
	assert.NotNil(t, commandList, "CommandList should not be nil")
	assert.NotEmpty(t, commandList.Commands, "Commands list should not be empty")

	// Expected commands
	expectedCommands := map[string]struct {
		category    string
		interactive bool
		hidden      bool
	}{
		"hello":       {category: "example", interactive: false, hidden: false},
		"config":      {category: "debug", interactive: false, hidden: false},
		"interactive": {category: "example", interactive: true, hidden: false},
	}

	// Verify all expected commands exist
	foundCommands := make(map[string]bool)
	for _, cmd := range commandList.Commands {
		foundCommands[cmd.Name] = true

		expected, ok := expectedCommands[cmd.Name]
		if ok {
			assert.Equal(t, expected.category, cmd.Category, "Category for %s should match", cmd.Name)
			assert.Equal(t, expected.interactive, cmd.Interactive, "Interactive flag for %s should match", cmd.Name)
			assert.Equal(t, expected.hidden, cmd.Hidden, "Hidden flag for %s should match", cmd.Name)
			assert.NotEmpty(t, cmd.Description, "Description for %s should not be empty", cmd.Name)
		}
	}

	// Ensure all expected commands were found
	for cmdName := range expectedCommands {
		assert.True(t, foundCommands[cmdName], "Command %s should be in the list", cmdName)
	}
}

func TestMyPlugin_ExecuteCommand(t *testing.T) {
	tests := []struct {
		name              string
		command           string
		args              []string
		config            map[string]interface{}
		expectSuccess     bool
		expectInteractive bool
		outputContains    string
		errorContains     string
	}{
		{
			name:           "hello without args",
			command:        "hello",
			expectSuccess:  true,
			outputContains: "Hello, World!",
		},
		{
			name:           "hello with name",
			command:        "hello",
			args:           []string{"Alice"},
			expectSuccess:  true,
			outputContains: "Hello, Alice!",
		},
		{
			name:           "config without configuration",
			command:        "config",
			config:         map[string]interface{}{},
			expectSuccess:  true,
			outputContains: "No configuration provided",
		},
		{
			name:    "config with configuration",
			command: "config",
			config: map[string]interface{}{
				"api_key": "test-123",
				"timeout": "30",
			},
			expectSuccess:  true,
			outputContains: "api_key",
		},
		{
			name:              "interactive requires interactive mode",
			command:           "interactive",
			expectSuccess:     false,
			expectInteractive: true,
			errorContains:     "requires interactive mode",
		},
		{
			name:          "unknown command",
			command:       "unknown",
			expectSuccess: false,
			errorContains: "unknown command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &MyPlugin{
				config: tt.config,
			}
			if plugin.config == nil {
				plugin.config = make(map[string]interface{})
			}

			ctx := context.Background()
			req := &sdk.ExecuteRequest{
				Command: tt.command,
				Args:    tt.args,
			}

			resp, err := plugin.ExecuteCommand(ctx, req)

			require.NoError(t, err, "ExecuteCommand should not return a gRPC error")
			assert.NotNil(t, resp, "Response should not be nil")

			assert.Equal(t, tt.expectSuccess, resp.Success, "Success flag should match expected")
			assert.Equal(t, tt.expectInteractive, resp.RequiresInteractive, "RequiresInteractive flag should match")

			if tt.outputContains != "" {
				assert.Contains(t, string(resp.Stdout), tt.outputContains, "Output should contain expected text")
			}

			if tt.errorContains != "" {
				assert.Contains(t, resp.Error, tt.errorContains, "Error should contain expected text")
			}
		})
	}
}

func TestMyPlugin_StartInteractive(t *testing.T) {
	plugin := &MyPlugin{
		config: make(map[string]interface{}),
	}

	stream := NewMockStream()

	// Add mock input messages
	stream.AddRecvMessage(&sdk.StreamMessage{
		Type: sdk.StreamMessage_STDIN,
		Data: []byte("test input\n"),
	})
	stream.AddRecvMessage(&sdk.StreamMessage{
		Type: sdk.StreamMessage_STDIN,
		Data: []byte("exit\n"),
	})

	err := plugin.StartInteractive(stream)
	require.NoError(t, err, "StartInteractive should not return an error")

	// Verify sent messages
	assert.NotEmpty(t, stream.sentMessages, "Should have sent messages")

	// Check for initial message
	foundStartMessage := false
	foundPrompt := false
	foundEcho := false
	foundGoodbye := false
	foundExit := false

	for _, msg := range stream.sentMessages {
		output := string(msg.Data)
		switch msg.Type {
		case sdk.StreamMessage_STDOUT:
			if strings.Contains(output, "Starting interactive") {
				foundStartMessage = true
			}
			if strings.Contains(output, "> ") {
				foundPrompt = true
			}
			if strings.Contains(output, "You typed: test input") {
				foundEcho = true
			}
			if strings.Contains(output, "Goodbye!") {
				foundGoodbye = true
			}
		case sdk.StreamMessage_EXIT:
			foundExit = true
			assert.Equal(t, int32(0), msg.ExitCode, "Exit code should be 0")
		}
	}

	assert.True(t, foundStartMessage, "Should have sent start message")
	assert.True(t, foundPrompt, "Should have sent prompt")
	assert.True(t, foundEcho, "Should have echoed input")
	assert.True(t, foundGoodbye, "Should have sent goodbye message")
	assert.True(t, foundExit, "Should have sent exit message")
}

func TestMyPlugin_GetCapabilities(t *testing.T) {
	plugin := &MyPlugin{
		config: make(map[string]interface{}),
	}

	ctx := context.Background()
	capabilities, err := plugin.GetCapabilities(ctx, &sdk.Empty{})

	require.NoError(t, err, "GetCapabilities should not return an error")
	assert.NotNil(t, capabilities, "Capabilities should not be nil")

	// Default boilerplate doesn't require special capabilities
	assert.False(t, capabilities.RequiresDocker, "Should not require Docker by default")
	assert.False(t, capabilities.RequiresNetwork, "Should not require Network by default")
}

// Integration test to verify plugin implements the correct interface
func TestMyPlugin_ImplementsInterface(t *testing.T) {
	plugin := &MyPlugin{
		config: make(map[string]interface{}),
	}

	// This test verifies that MyPlugin implements sdk.GlidePluginServer
	var _ sdk.GlidePluginServer = plugin
}

// Benchmark tests
func BenchmarkMyPlugin_ExecuteHello(b *testing.B) {
	plugin := &MyPlugin{
		config: make(map[string]interface{}),
	}

	ctx := context.Background()
	req := &sdk.ExecuteRequest{
		Command: "hello",
		Args:    []string{"Benchmark"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = plugin.ExecuteCommand(ctx, req)
	}
}

func BenchmarkMyPlugin_ListCommands(b *testing.B) {
	plugin := &MyPlugin{
		config: make(map[string]interface{}),
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = plugin.ListCommands(ctx, &sdk.Empty{})
	}
}
