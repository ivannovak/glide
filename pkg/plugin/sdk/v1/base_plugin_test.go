package v1

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

// MockStream implements GlidePlugin_StartInteractiveServer for testing
type MockStream struct {
	mock.Mock
	sentMessages []*StreamMessage
	receiveQueue []*StreamMessage
	receiveIndex int
}

func NewMockStream() *MockStream {
	return &MockStream{
		sentMessages: make([]*StreamMessage, 0),
		receiveQueue: make([]*StreamMessage, 0),
	}
}

func (m *MockStream) Send(msg *StreamMessage) error {
	args := m.Called(msg)
	m.sentMessages = append(m.sentMessages, msg)
	return args.Error(0)
}

func (m *MockStream) Recv() (*StreamMessage, error) {
	// First check if we have messages in the queue
	if m.receiveIndex < len(m.receiveQueue) {
		msg := m.receiveQueue[m.receiveIndex]
		m.receiveIndex++
		return msg, nil
	}

	// If queue is empty and mock expectations are set, use them
	if len(m.ExpectedCalls) > 0 {
		args := m.Called()

		// Check if we have both return values
		if len(args) >= 2 {
			if args.Get(0) != nil {
				return args.Get(0).(*StreamMessage), args.Error(1)
			}
			return nil, args.Error(1)
		}
	}

	return nil, io.EOF
}

// Add message to the receive queue
func (m *MockStream) AddToReceiveQueue(msg *StreamMessage) {
	m.receiveQueue = append(m.receiveQueue, msg)
}

// Implement other required methods
func (m *MockStream) SetHeader(metadata.MD) error  { return nil }
func (m *MockStream) SendHeader(metadata.MD) error { return nil }
func (m *MockStream) SetTrailer(metadata.MD)       {}
func (m *MockStream) Context() context.Context     { return context.Background() }
func (m *MockStream) SendMsg(interface{}) error    { return nil }
func (m *MockStream) RecvMsg(interface{}) error    { return nil }

// TestBasePlugin_Metadata tests plugin metadata handling
func TestBasePlugin_Metadata(t *testing.T) {
	metadata := &PluginMetadata{
		Name:        "test-plugin",
		Version:     "1.0.0",
		Author:      "Test Author",
		Description: "Test Description",
		MinSdk:      "v1.0.0",
		Namespaced:  false,
	}

	plugin := NewBasePlugin(metadata)

	// Test GetMetadata
	result, err := plugin.GetMetadata(context.Background(), &Empty{})
	require.NoError(t, err)
	assert.Equal(t, metadata, result)
}

// TestBasePlugin_Configuration tests plugin configuration
func TestBasePlugin_Configuration(t *testing.T) {
	plugin := NewBasePlugin(&PluginMetadata{
		Name: "test-plugin",
	})

	// Test Configure
	config := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	resp, err := plugin.Configure(context.Background(), &ConfigureRequest{
		Config: config,
	})

	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Contains(t, resp.Message, "test-plugin")

	// Verify configuration was stored
	storedConfig := plugin.GetConfig()
	assert.Equal(t, "value1", storedConfig["key1"])
	assert.Equal(t, "value2", storedConfig["key2"])
}

// TestBasePlugin_CommandRegistration tests registering and listing commands
func TestBasePlugin_CommandRegistration(t *testing.T) {
	plugin := NewBasePlugin(&PluginMetadata{
		Name: "test-plugin",
	})

	// Register a simple command
	simpleCmd := NewSimpleCommand(
		&CommandInfo{
			Name:        "test-cmd",
			Description: "Test command",
			Category:    CategoryDeveloper,
		},
		func(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error) {
			return &ExecuteResponse{
				Success: true,
				Stdout:  []byte("test output"),
			}, nil
		},
	)

	plugin.RegisterCommand("test-cmd", simpleCmd)

	// Register an interactive command
	interactiveCmd := NewBaseInteractiveCommand(
		&CommandInfo{
			Name:        "interactive-cmd",
			Description: "Interactive test command",
			Category:    CategoryDeveloper,
		},
		nil, // Use default execute handler
		func(stream GlidePlugin_StartInteractiveServer) error {
			return nil
		},
	)

	plugin.RegisterCommand("interactive-cmd", interactiveCmd)

	// Test ListCommands
	cmdList, err := plugin.ListCommands(context.Background(), &Empty{})
	require.NoError(t, err)
	assert.Len(t, cmdList.Commands, 2)

	// Verify commands are registered
	var foundSimple, foundInteractive bool
	for _, cmd := range cmdList.Commands {
		if cmd.Name == "test-cmd" {
			foundSimple = true
			assert.False(t, cmd.Interactive)
		}
		if cmd.Name == "interactive-cmd" {
			foundInteractive = true
			assert.True(t, cmd.Interactive)
		}
	}
	assert.True(t, foundSimple)
	assert.True(t, foundInteractive)
}

// TestBasePlugin_ExecuteCommand tests non-interactive command execution
func TestBasePlugin_ExecuteCommand(t *testing.T) {
	plugin := NewBasePlugin(&PluginMetadata{
		Name: "test-plugin",
	})

	// Register a command
	executed := false
	cmd := NewSimpleCommand(
		&CommandInfo{
			Name: "test-cmd",
		},
		func(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error) {
			executed = true
			assert.Equal(t, "test-cmd", req.Command)
			assert.Equal(t, []string{"arg1", "arg2"}, req.Args)
			return &ExecuteResponse{
				Success: true,
				Stdout:  []byte("output"),
			}, nil
		},
	)

	plugin.RegisterCommand("test-cmd", cmd)

	// Execute the command
	resp, err := plugin.ExecuteCommand(context.Background(), &ExecuteRequest{
		Command: "test-cmd",
		Args:    []string{"arg1", "arg2"},
	})

	require.NoError(t, err)
	assert.True(t, executed)
	assert.True(t, resp.Success)
	assert.Equal(t, []byte("output"), resp.Stdout)
}

// TestBasePlugin_ExecuteUnknownCommand tests executing unknown command
func TestBasePlugin_ExecuteUnknownCommand(t *testing.T) {
	plugin := NewBasePlugin(&PluginMetadata{
		Name: "test-plugin",
	})

	resp, err := plugin.ExecuteCommand(context.Background(), &ExecuteRequest{
		Command: "unknown-cmd",
	})

	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Error, "unknown command")
}

// TestBasePlugin_StartInteractive tests interactive command routing
func TestBasePlugin_StartInteractive(t *testing.T) {
	plugin := NewBasePlugin(&PluginMetadata{
		Name: "test-plugin",
	})

	// Track if interactive handler was called
	interactiveCalled := false

	// Register an interactive command
	cmd := NewBaseInteractiveCommand(
		&CommandInfo{
			Name:        "shell",
			Interactive: true,
		},
		nil,
		func(stream GlidePlugin_StartInteractiveServer) error {
			interactiveCalled = true

			// Verify we can send a message
			err := stream.Send(&StreamMessage{
				Type: StreamMessage_STDOUT,
				Data: []byte("Interactive session started"),
			})
			assert.NoError(t, err)

			return nil
		},
	)

	plugin.RegisterCommand("shell", cmd)

	// Create mock stream
	stream := NewMockStream()

	// Set up the initial message with command name
	stream.On("Recv").Return(&StreamMessage{
		Type: StreamMessage_STDIN,
		Data: []byte("shell"),
	}, nil).Once()

	// Allow Send to work
	stream.On("Send", mock.Anything).Return(nil)

	// Call StartInteractive
	err := plugin.StartInteractive(stream)

	require.NoError(t, err)
	assert.True(t, interactiveCalled)

	// Verify the message was sent
	assert.Len(t, stream.sentMessages, 1)
	assert.Equal(t, StreamMessage_STDOUT, stream.sentMessages[0].Type)
	assert.Equal(t, []byte("Interactive session started"), stream.sentMessages[0].Data)
}

// TestBasePlugin_StartInteractive_NoCommandName tests error when no command name is provided
func TestBasePlugin_StartInteractive_NoCommandName(t *testing.T) {
	plugin := NewBasePlugin(&PluginMetadata{
		Name: "test-plugin",
	})

	stream := NewMockStream()

	// Send empty message
	stream.On("Recv").Return(&StreamMessage{
		Type: StreamMessage_STDIN,
		Data: []byte(""),
	}, nil).Once()

	err := plugin.StartInteractive(stream)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no command specified")
}

// TestBasePlugin_StartInteractive_UnknownCommand tests error for unknown command
func TestBasePlugin_StartInteractive_UnknownCommand(t *testing.T) {
	plugin := NewBasePlugin(&PluginMetadata{
		Name: "test-plugin",
	})

	stream := NewMockStream()

	// Send unknown command name
	stream.On("Recv").Return(&StreamMessage{
		Type: StreamMessage_STDIN,
		Data: []byte("unknown-cmd"),
	}, nil).Once()

	err := plugin.StartInteractive(stream)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown command")
}

// TestBasePlugin_StartInteractive_NonInteractiveCommand tests error when command doesn't support interactive
func TestBasePlugin_StartInteractive_NonInteractiveCommand(t *testing.T) {
	plugin := NewBasePlugin(&PluginMetadata{
		Name: "test-plugin",
	})

	// Register a non-interactive command
	cmd := NewSimpleCommand(
		&CommandInfo{
			Name:        "simple-cmd",
			Interactive: false,
		},
		func(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error) {
			return &ExecuteResponse{Success: true}, nil
		},
	)

	plugin.RegisterCommand("simple-cmd", cmd)

	stream := NewMockStream()

	// Try to use it interactively
	stream.On("Recv").Return(&StreamMessage{
		Type: StreamMessage_STDIN,
		Data: []byte("simple-cmd"),
	}, nil).Once()

	err := plugin.StartInteractive(stream)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not support interactive mode")
}

// TestBasePlugin_StartInteractive_RecvError tests handling of receive errors
func TestBasePlugin_StartInteractive_RecvError(t *testing.T) {
	plugin := NewBasePlugin(&PluginMetadata{
		Name: "test-plugin",
	})

	stream := NewMockStream()

	// Simulate receive error
	expectedErr := errors.New("connection lost")
	stream.On("Recv").Return((*StreamMessage)(nil), expectedErr).Once()

	err := plugin.StartInteractive(stream)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to receive initial message")
}

// TestInteractiveCommand_DefaultExecute tests default execute behavior for interactive commands
func TestInteractiveCommand_DefaultExecute(t *testing.T) {
	cmd := NewBaseInteractiveCommand(
		&CommandInfo{
			Name: "test-interactive",
		},
		nil, // No custom execute handler
		func(stream GlidePlugin_StartInteractiveServer) error {
			return nil
		},
	)

	// Should return RequiresInteractive by default
	resp, err := cmd.Execute(context.Background(), &ExecuteRequest{})

	require.NoError(t, err)
	assert.True(t, resp.RequiresInteractive)
}

// TestSimpleCommand tests the SimpleCommand helper
func TestSimpleCommand(t *testing.T) {
	executed := false
	cmd := NewSimpleCommand(
		&CommandInfo{
			Name:        "simple",
			Description: "Simple test command",
		},
		func(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error) {
			executed = true
			return &ExecuteResponse{
				Success: true,
				Stdout:  []byte("simple output"),
			}, nil
		},
	)

	// Test Info
	info := cmd.Info()
	assert.Equal(t, "simple", info.Name)
	assert.Equal(t, "Simple test command", info.Description)

	// Test Execute
	resp, err := cmd.Execute(context.Background(), &ExecuteRequest{})
	require.NoError(t, err)
	assert.True(t, executed)
	assert.True(t, resp.Success)
	assert.Equal(t, []byte("simple output"), resp.Stdout)
}

// TestBasePlugin_CompleteInteractiveFlow tests a complete interactive session flow
func TestBasePlugin_CompleteInteractiveFlow(t *testing.T) {
	plugin := NewBasePlugin(&PluginMetadata{
		Name: "test-plugin",
	})

	// Register an echo-like interactive command
	cmd := NewBaseInteractiveCommand(
		&CommandInfo{
			Name:        "echo",
			Interactive: true,
		},
		nil,
		func(stream GlidePlugin_StartInteractiveServer) error {
			// Send welcome
			stream.Send(&StreamMessage{
				Type: StreamMessage_STDOUT,
				Data: []byte("Echo started\n"),
			})

			// Echo loop
			for {
				msg, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					return err
				}

				if msg.Type == StreamMessage_STDIN {
					input := string(msg.Data)
					if input == "exit\n" {
						stream.Send(&StreamMessage{
							Type:     StreamMessage_EXIT,
							ExitCode: 0,
						})
						return nil
					}

					// Echo back
					stream.Send(&StreamMessage{
						Type: StreamMessage_STDOUT,
						Data: []byte("Echo: " + input),
					})
				}
			}
			return nil
		},
	)

	plugin.RegisterCommand("echo", cmd)

	// Create mock stream
	stream := NewMockStream()

	// Setup the interaction sequence by adding messages to the receive queue
	stream.AddToReceiveQueue(&StreamMessage{
		Type: StreamMessage_STDIN,
		Data: []byte("echo"),
	})
	stream.AddToReceiveQueue(&StreamMessage{
		Type: StreamMessage_STDIN,
		Data: []byte("hello\n"),
	})
	stream.AddToReceiveQueue(&StreamMessage{
		Type: StreamMessage_STDIN,
		Data: []byte("exit\n"),
	})

	stream.On("Send", mock.Anything).Return(nil)

	// Run the interactive session
	err := plugin.StartInteractive(stream)
	require.NoError(t, err)

	// Verify the sent messages
	require.Len(t, stream.sentMessages, 3)

	// Welcome message
	assert.Equal(t, StreamMessage_STDOUT, stream.sentMessages[0].Type)
	assert.Equal(t, []byte("Echo started\n"), stream.sentMessages[0].Data)

	// Echo response
	assert.Equal(t, StreamMessage_STDOUT, stream.sentMessages[1].Type)
	assert.Equal(t, []byte("Echo: hello\n"), stream.sentMessages[1].Data)

	// Exit message
	assert.Equal(t, StreamMessage_EXIT, stream.sentMessages[2].Type)
	assert.Equal(t, int32(0), stream.sentMessages[2].ExitCode)
}
