package v1

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInteractiveStream tests interactive command streaming
func TestInteractiveStream(t *testing.T) {
	t.Run("send and receive messages", func(t *testing.T) {
		stream := NewMockStream()

		// Queue messages to receive
		stream.AddToReceiveQueue(&StreamMessage{
			Type: StreamMessage_STDIN,
			Data: []byte("test input"),
		})

		stream.On("Send", &StreamMessage{
			Type: StreamMessage_STDOUT,
			Data: []byte("test output"),
		}).Return(nil)

		// Send message
		err := stream.Send(&StreamMessage{
			Type: StreamMessage_STDOUT,
			Data: []byte("test output"),
		})
		assert.NoError(t, err)

		// Receive message
		msg, err := stream.Recv()
		require.NoError(t, err)
		assert.Equal(t, StreamMessage_STDIN, msg.Type)
		assert.Equal(t, []byte("test input"), msg.Data)
	})

	t.Run("receive EOF", func(t *testing.T) {
		stream := NewMockStream()

		// Don't queue any messages - should get EOF
		msg, err := stream.Recv()
		assert.Equal(t, io.EOF, err)
		assert.Nil(t, msg)
	})

	t.Run("stream with signal", func(t *testing.T) {
		stream := NewMockStream()

		stream.AddToReceiveQueue(&StreamMessage{
			Type:   StreamMessage_SIGNAL,
			Signal: "SIGINT",
		})

		msg, err := stream.Recv()
		require.NoError(t, err)
		assert.Equal(t, StreamMessage_SIGNAL, msg.Type)
		assert.Equal(t, "SIGINT", msg.Signal)
	})

	t.Run("stream with resize", func(t *testing.T) {
		stream := NewMockStream()

		stream.AddToReceiveQueue(&StreamMessage{
			Type:   StreamMessage_RESIZE,
			Width:  120,
			Height: 40,
		})

		msg, err := stream.Recv()
		require.NoError(t, err)
		assert.Equal(t, StreamMessage_RESIZE, msg.Type)
		assert.Equal(t, int32(120), msg.Width)
		assert.Equal(t, int32(40), msg.Height)
	})

	t.Run("stream with exit code", func(t *testing.T) {
		stream := NewMockStream()

		stream.AddToReceiveQueue(&StreamMessage{
			Type:     StreamMessage_EXIT,
			ExitCode: 42,
		})

		msg, err := stream.Recv()
		require.NoError(t, err)
		assert.Equal(t, StreamMessage_EXIT, msg.Type)
		assert.Equal(t, int32(42), msg.ExitCode)
	})

	t.Run("stream with error message", func(t *testing.T) {
		stream := NewMockStream()

		stream.AddToReceiveQueue(&StreamMessage{
			Type:  StreamMessage_ERROR,
			Error: "something went wrong",
		})

		msg, err := stream.Recv()
		require.NoError(t, err)
		assert.Equal(t, StreamMessage_ERROR, msg.Type)
		assert.Equal(t, "something went wrong", msg.Error)
	})
}

// TestGRPCProtocol tests gRPC protocol integration
func TestGRPCProtocol(t *testing.T) {
	t.Run("handshake config", func(t *testing.T) {
		assert.Equal(t, uint(1), HandshakeConfig.ProtocolVersion)
		assert.Equal(t, "GLIDE_PLUGIN_MAGIC", HandshakeConfig.MagicCookieKey)
		assert.NotEmpty(t, HandshakeConfig.MagicCookieValue)
	})

	t.Run("plugin map", func(t *testing.T) {
		assert.NotNil(t, PluginMap)
		assert.Contains(t, PluginMap, "glide")
		assert.IsType(t, &GlidePluginImpl{}, PluginMap["glide"])
	})
}

// TestPluginMetadata tests metadata operations
func TestPluginMetadata(t *testing.T) {
	metadata := &PluginMetadata{
		Name:          "test-plugin",
		Version:       "1.0.0",
		Author:        "Test Author",
		Description:   "Test Description",
		Homepage:      "https://example.com",
		License:       "MIT",
		MinSdk:        "v1.0.0",
		BuildTimeUnix: 1234567890,
		Tags:          []string{"test", "example"},
		Aliases:       []string{"tp", "test"},
		Namespaced:    false,
	}

	t.Run("has all fields", func(t *testing.T) {
		assert.Equal(t, "test-plugin", metadata.Name)
		assert.Equal(t, "1.0.0", metadata.Version)
		assert.Equal(t, "Test Author", metadata.Author)
		assert.Equal(t, "Test Description", metadata.Description)
		assert.Equal(t, "https://example.com", metadata.Homepage)
		assert.Equal(t, "MIT", metadata.License)
		assert.Equal(t, "v1.0.0", metadata.MinSdk)
		assert.Equal(t, int64(1234567890), metadata.BuildTimeUnix)
		assert.Equal(t, []string{"test", "example"}, metadata.Tags)
		assert.Equal(t, []string{"tp", "test"}, metadata.Aliases)
		assert.False(t, metadata.Namespaced)
	})
}

// TestCommandInfo tests command information structure
func TestCommandInfo(t *testing.T) {
	cmdInfo := &CommandInfo{
		Name:         "test-cmd",
		Description:  "Test command description",
		Category:     CategoryDeveloper,
		Aliases:      []string{"tc", "test"},
		Interactive:  false,
		Hidden:       false,
		RequiresTty:  false,
		RequiresAuth: false,
	}

	t.Run("has correct fields", func(t *testing.T) {
		assert.Equal(t, "test-cmd", cmdInfo.Name)
		assert.Equal(t, "Test command description", cmdInfo.Description)
		assert.Equal(t, CategoryDeveloper, cmdInfo.Category)
		assert.Equal(t, []string{"tc", "test"}, cmdInfo.Aliases)
		assert.False(t, cmdInfo.Interactive)
		assert.False(t, cmdInfo.Hidden)
		assert.False(t, cmdInfo.RequiresTty)
		assert.False(t, cmdInfo.RequiresAuth)
	})
}

// TestExecuteRequest tests request structure
func TestExecuteRequest(t *testing.T) {
	req := &ExecuteRequest{
		Command: "test",
		Args:    []string{"arg1", "arg2"},
		Flags:   map[string]string{"flag1": "value1"},
		Env:     map[string]string{"ENV1": "VALUE1"},
		WorkDir: "/test/dir",
		Stdin:   []byte("input data"),
	}

	t.Run("has all fields", func(t *testing.T) {
		assert.Equal(t, "test", req.Command)
		assert.Equal(t, []string{"arg1", "arg2"}, req.Args)
		assert.Equal(t, map[string]string{"flag1": "value1"}, req.Flags)
		assert.Equal(t, map[string]string{"ENV1": "VALUE1"}, req.Env)
		assert.Equal(t, "/test/dir", req.WorkDir)
		assert.Equal(t, []byte("input data"), req.Stdin)
	})
}

// TestExecuteResponse tests response structure
func TestExecuteResponse(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		resp := &ExecuteResponse{
			Success:  true,
			ExitCode: 0,
			Stdout:   []byte("output"),
			Stderr:   []byte(""),
			Error:    "",
		}

		assert.True(t, resp.Success)
		assert.Equal(t, int32(0), resp.ExitCode)
		assert.Equal(t, []byte("output"), resp.Stdout)
		assert.Empty(t, resp.Stderr)
		assert.Empty(t, resp.Error)
	})

	t.Run("failed response", func(t *testing.T) {
		resp := &ExecuteResponse{
			Success:  false,
			ExitCode: 1,
			Stdout:   []byte(""),
			Stderr:   []byte("error output"),
			Error:    "command failed",
		}

		assert.False(t, resp.Success)
		assert.Equal(t, int32(1), resp.ExitCode)
		assert.Empty(t, resp.Stdout)
		assert.Equal(t, []byte("error output"), resp.Stderr)
		assert.Equal(t, "command failed", resp.Error)
	})

	t.Run("requires interactive", func(t *testing.T) {
		resp := &ExecuteResponse{
			RequiresInteractive: true,
		}

		assert.True(t, resp.RequiresInteractive)
	})
}

// TestCapabilities tests capabilities structure
func TestCapabilities(t *testing.T) {
	caps := &Capabilities{
		RequiresDocker:      true,
		RequiresNetwork:     false,
		RequiresFilesystem:  false,
		RequiresInteractive: true,
		RequiredCommands:    []string{"docker", "git"},
		RequiredPaths:       []string{"/usr/bin/docker"},
		RequiredEnvVars:     []string{"DOCKER_HOST"},
	}

	t.Run("has correct fields", func(t *testing.T) {
		assert.True(t, caps.RequiresDocker)
		assert.False(t, caps.RequiresNetwork)
		assert.False(t, caps.RequiresFilesystem)
		assert.True(t, caps.RequiresInteractive)
		assert.Equal(t, []string{"docker", "git"}, caps.RequiredCommands)
		assert.Equal(t, []string{"/usr/bin/docker"}, caps.RequiredPaths)
		assert.Equal(t, []string{"DOCKER_HOST"}, caps.RequiredEnvVars)
	})
}

// TestCategoryConstants tests category constants exist
func TestCategoryConstants(t *testing.T) {
	// Test that category constants are available
	assert.Equal(t, CategoryDeveloper, CategoryDeveloper)
}

// TestStreamMessage tests stream message types
func TestStreamMessage(t *testing.T) {
	t.Run("stdin message", func(t *testing.T) {
		msg := &StreamMessage{
			Type: StreamMessage_STDIN,
			Data: []byte("input"),
		}
		assert.Equal(t, StreamMessage_STDIN, msg.Type)
		assert.Equal(t, []byte("input"), msg.Data)
	})

	t.Run("stdout message", func(t *testing.T) {
		msg := &StreamMessage{
			Type: StreamMessage_STDOUT,
			Data: []byte("output"),
		}
		assert.Equal(t, StreamMessage_STDOUT, msg.Type)
		assert.Equal(t, []byte("output"), msg.Data)
	})

	t.Run("stderr message", func(t *testing.T) {
		msg := &StreamMessage{
			Type: StreamMessage_STDERR,
			Data: []byte("error"),
		}
		assert.Equal(t, StreamMessage_STDERR, msg.Type)
		assert.Equal(t, []byte("error"), msg.Data)
	})

	t.Run("exit message", func(t *testing.T) {
		msg := &StreamMessage{
			Type:     StreamMessage_EXIT,
			ExitCode: 0,
		}
		assert.Equal(t, StreamMessage_EXIT, msg.Type)
		assert.Equal(t, int32(0), msg.ExitCode)
	})

	t.Run("signal message", func(t *testing.T) {
		msg := &StreamMessage{
			Type:   StreamMessage_SIGNAL,
			Signal: "SIGINT",
		}
		assert.Equal(t, StreamMessage_SIGNAL, msg.Type)
		assert.Equal(t, "SIGINT", msg.Signal)
	})

	t.Run("resize message", func(t *testing.T) {
		msg := &StreamMessage{
			Type:   StreamMessage_RESIZE,
			Width:  80,
			Height: 24,
		}
		assert.Equal(t, StreamMessage_RESIZE, msg.Type)
		assert.Equal(t, int32(80), msg.Width)
		assert.Equal(t, int32(24), msg.Height)
	})

	t.Run("error message", func(t *testing.T) {
		msg := &StreamMessage{
			Type:  StreamMessage_ERROR,
			Error: "error occurred",
		}
		assert.Equal(t, StreamMessage_ERROR, msg.Type)
		assert.Equal(t, "error occurred", msg.Error)
	})
}

// TestConfigureRequest tests configuration request
func TestConfigureRequest(t *testing.T) {
	config := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	req := &ConfigureRequest{
		Config: config,
	}

	assert.Equal(t, config, req.Config)
	assert.Equal(t, "value1", req.Config["key1"])
	assert.Equal(t, "value2", req.Config["key2"])
}

// TestConfigureResponse tests configuration response
func TestConfigureResponse(t *testing.T) {
	t.Run("successful configuration", func(t *testing.T) {
		resp := &ConfigureResponse{
			Success: true,
			Message: "configured successfully",
		}

		assert.True(t, resp.Success)
		assert.Equal(t, "configured successfully", resp.Message)
	})

	t.Run("failed configuration", func(t *testing.T) {
		resp := &ConfigureResponse{
			Success: false,
			Message: "configuration failed: invalid value",
		}

		assert.False(t, resp.Success)
		assert.Contains(t, resp.Message, "failed")
	})
}

// TestCommandList tests command list structure
func TestCommandList(t *testing.T) {
	commands := []*CommandInfo{
		{Name: "cmd1", Description: "Command 1"},
		{Name: "cmd2", Description: "Command 2"},
		{Name: "cmd3", Description: "Command 3"},
	}

	cmdList := &CommandList{
		Commands: commands,
	}

	assert.Len(t, cmdList.Commands, 3)
	assert.Equal(t, "cmd1", cmdList.Commands[0].Name)
	assert.Equal(t, "cmd2", cmdList.Commands[1].Name)
	assert.Equal(t, "cmd3", cmdList.Commands[2].Name)
}

// TestCustomCategory tests custom category structure
func TestCustomCategory(t *testing.T) {
	category := &CustomCategory{
		Id:          "custom-1",
		Name:        "Custom Category",
		Description: "A custom category for special commands",
		Priority:    10,
	}

	assert.Equal(t, "custom-1", category.Id)
	assert.Equal(t, "Custom Category", category.Name)
	assert.Equal(t, "A custom category for special commands", category.Description)
	assert.Equal(t, int32(10), category.Priority)
}

// TestCustomCategoryList tests custom categories list
func TestCustomCategoryList(t *testing.T) {
	categories := []*CustomCategory{
		{Id: "cat1", Name: "Category 1"},
		{Id: "cat2", Name: "Category 2"},
	}

	assert.Len(t, categories, 2)
	assert.Equal(t, "cat1", categories[0].Id)
	assert.Equal(t, "cat2", categories[1].Id)
}

// TestEmpty tests empty message
func TestEmpty(t *testing.T) {
	empty := &Empty{}
	assert.NotNil(t, empty)
}
