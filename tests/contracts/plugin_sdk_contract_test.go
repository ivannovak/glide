// Package contracts contains contract tests for the plugin SDK
//
// Contract tests ensure backward compatibility and protocol compliance:
// - Plugins built with old SDK versions work with new host
// - SDK changes don't break existing plugin interfaces
// - gRPC protocol remains compatible
package contracts

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	sdk "github.com/ivannovak/glide/v3/pkg/plugin/sdk"
	v1 "github.com/ivannovak/glide/v3/pkg/plugin/sdk/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSDKv1Protocol tests that the SDK correctly implements the v1 protocol
func TestSDKv1Protocol(t *testing.T) {
	t.Run("protocol version is 1", func(t *testing.T) {
		assert.Equal(t, uint(1), v1.HandshakeConfig.ProtocolVersion)
	})

	t.Run("magic cookie is defined", func(t *testing.T) {
		assert.NotEmpty(t, v1.HandshakeConfig.MagicCookieKey)
		assert.NotEmpty(t, v1.HandshakeConfig.MagicCookieValue)
	})

	t.Run("plugin map contains glide plugin", func(t *testing.T) {
		assert.Contains(t, v1.PluginMap, "glide")
	})
}

// TestPluginInterface tests that plugins implement all required methods
func TestPluginInterface(t *testing.T) {
	plugin := v1.NewBasePlugin(&v1.PluginMetadata{
		Name:    "contract-test",
		Version: "1.0.0",
	})

	t.Run("implements GetMetadata", func(t *testing.T) {
		meta, err := plugin.GetMetadata(context.Background(), &v1.Empty{})
		require.NoError(t, err)
		assert.Equal(t, "contract-test", meta.Name)
	})

	t.Run("implements Configure", func(t *testing.T) {
		resp, err := plugin.Configure(context.Background(), &v1.ConfigureRequest{
			Config: map[string]string{"key": "value"},
		})
		require.NoError(t, err)
		assert.True(t, resp.Success)
	})

	t.Run("implements ListCommands", func(t *testing.T) {
		list, err := plugin.ListCommands(context.Background(), &v1.Empty{})
		require.NoError(t, err)
		assert.NotNil(t, list)
	})

	t.Run("implements ExecuteCommand", func(t *testing.T) {
		// Register a test command first
		plugin.RegisterCommand("test", v1.NewSimpleCommand(
			&v1.CommandInfo{Name: "test"},
			func(ctx context.Context, req *v1.ExecuteRequest) (*v1.ExecuteResponse, error) {
				return &v1.ExecuteResponse{Success: true}, nil
			},
		))

		resp, err := plugin.ExecuteCommand(context.Background(), &v1.ExecuteRequest{
			Command: "test",
		})
		require.NoError(t, err)
		assert.True(t, resp.Success)
	})
}

// TestBackwardCompatibility tests that old plugin contracts still work
func TestBackwardCompatibility(t *testing.T) {
	t.Run("v1 metadata fields are preserved", func(t *testing.T) {
		// Test that all original v1 metadata fields exist and work
		meta := &v1.PluginMetadata{
			Name:        "old-plugin",
			Version:     "1.0.0",
			Author:      "Author",
			Description: "Description",
			MinSdk:      "v1.0.0",
		}

		assert.Equal(t, "old-plugin", meta.Name)
		assert.Equal(t, "1.0.0", meta.Version)
		assert.Equal(t, "Author", meta.Author)
		assert.Equal(t, "Description", meta.Description)
		assert.Equal(t, "v1.0.0", meta.MinSdk)
	})

	t.Run("v1 command categories still work", func(t *testing.T) {
		info := &v1.CommandInfo{
			Category: v1.CategoryDeveloper,
		}
		assert.Equal(t, v1.CategoryDeveloper, info.Category)
	})

	t.Run("v1 execute request format unchanged", func(t *testing.T) {
		req := &v1.ExecuteRequest{
			Command: "cmd",
			Args:    []string{"arg1"},
			Flags:   map[string]string{"flag": "value"},
		}

		assert.Equal(t, "cmd", req.Command)
		assert.Equal(t, []string{"arg1"}, req.Args)
		assert.Equal(t, map[string]string{"flag": "value"}, req.Flags)
	})

	t.Run("v1 execute response format unchanged", func(t *testing.T) {
		resp := &v1.ExecuteResponse{
			Success:  true,
			ExitCode: 0,
			Stdout:   []byte("output"),
		}

		assert.True(t, resp.Success)
		assert.Equal(t, int32(0), resp.ExitCode)
		assert.Equal(t, []byte("output"), resp.Stdout)
	})
}

// TestManagerContract tests the manager's contract with plugins
func TestManagerContract(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping contract test in short mode")
	}

	tmpDir := t.TempDir()
	pluginPath := buildContractTestPlugin(t, tmpDir)

	config := &sdk.ManagerConfig{
		PluginDirs:     []string{tmpDir},
		SecurityStrict: false,
	}
	manager := sdk.NewManager(config)

	t.Run("manager can load v1 plugins", func(t *testing.T) {
		err := manager.LoadPlugin(pluginPath)
		assert.NoError(t, err)
	})

	t.Run("manager can get plugin metadata", func(t *testing.T) {
		plugin, err := manager.GetPlugin("contract-plugin")
		require.NoError(t, err)
		assert.NotNil(t, plugin.Metadata)
		assert.Equal(t, "contract-plugin", plugin.Metadata.Name)
	})

	t.Run("manager can list plugin commands", func(t *testing.T) {
		plugin, err := manager.GetPlugin("contract-plugin")
		require.NoError(t, err)

		commands, err := plugin.Plugin.ListCommands(context.Background(), &v1.Empty{})
		require.NoError(t, err)
		assert.NotNil(t, commands)
	})

	t.Run("manager can execute plugin commands", func(t *testing.T) {
		err := manager.ExecuteCommand("contract-plugin", "test", []string{})
		assert.NoError(t, err)
	})

	t.Run("manager cleanup kills plugins", func(t *testing.T) {
		plugin, err := manager.GetPlugin("contract-plugin")
		require.NoError(t, err)
		assert.False(t, plugin.Client.Exited())

		manager.Cleanup()

		// Give it a moment to shutdown
		time.Sleep(100 * time.Millisecond)
		assert.True(t, plugin.Client.Exited())
	})
}

// TestSecurityContract tests security validation contracts
func TestSecurityContract(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("validator rejects world-writable plugins", func(t *testing.T) {
		pluginPath := filepath.Join(tmpDir, "bad-plugin")
		err := os.WriteFile(pluginPath, []byte("#!/bin/sh\n"), 0755)
		require.NoError(t, err)
		// Explicitly set world-writable permissions (bypasses umask)
		err = os.Chmod(pluginPath, 0777)
		require.NoError(t, err)

		// Verify the file is actually world-writable
		info, err := os.Stat(pluginPath)
		require.NoError(t, err)
		if info.Mode()&0022 == 0 {
			t.Skip("filesystem does not support group/other write permissions")
		}

		validator := sdk.NewValidator(true)
		validator.AddTrustedPath(tmpDir)
		err = validator.Validate(pluginPath)
		assert.Error(t, err)
		// Error message varies based on OS and security checks
	})

	t.Run("validator accepts properly secured plugins", func(t *testing.T) {
		pluginPath := filepath.Join(tmpDir, "good-plugin")
		err := os.WriteFile(pluginPath, []byte{0x7f, 'E', 'L', 'F'}, 0755)
		require.NoError(t, err)

		validator := sdk.NewValidator(false)
		validator.AddTrustedPath(tmpDir)
		err = validator.Validate(pluginPath)
		assert.NoError(t, err)
	})

	t.Run("validator enforces trusted paths in strict mode", func(t *testing.T) {
		validator := sdk.NewValidator(true)
		validator.AddTrustedPath("/trusted/path")

		err := validator.Validate("/untrusted/plugin")
		assert.Error(t, err)
		// Error message varies - file doesn't exist or not in trusted path
	})
}

// buildContractTestPlugin creates a plugin for contract testing
func buildContractTestPlugin(t *testing.T, dir string) string {
	t.Helper()

	pluginSrc := `package main

import (
	"context"
	"github.com/hashicorp/go-plugin"
	sdk "github.com/ivannovak/glide/v3/pkg/plugin/sdk/v1"
)

func main() {
	basePlugin := sdk.NewBasePlugin(&sdk.PluginMetadata{
		Name:        "contract-plugin",
		Version:     "1.0.0",
		Author:      "Contract Test",
		Description: "Plugin for contract testing",
		MinSdk:      "v1.0.0",
	})

	basePlugin.RegisterCommand("test", sdk.NewSimpleCommand(
		&sdk.CommandInfo{
			Name:        "test",
			Description: "Test command",
			Category:    sdk.CategoryDeveloper,
		},
		func(ctx context.Context, req *sdk.ExecuteRequest) (*sdk.ExecuteResponse, error) {
			return &sdk.ExecuteResponse{
				Success: true,
				Stdout:  []byte("contract test output\n"),
			}, nil
		},
	))

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: sdk.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"glide": &sdk.GlidePluginImpl{Impl: basePlugin},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
`

	// Get glide module root
	cwd, err := os.Getwd()
	require.NoError(t, err)
	glideRoot := filepath.Join(cwd, "../..")
	glideRoot, err = filepath.Abs(glideRoot)
	require.NoError(t, err)

	goMod := `module contracttest

go 1.23

replace github.com/ivannovak/glide/v3 => ` + glideRoot + `

require github.com/ivannovak/glide/v3 v3.0.0
`

	// Write files
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "main.go"), []byte(pluginSrc), 0644))

	// Build
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = dir
	tidyOutput, err := tidyCmd.CombinedOutput()
	require.NoError(t, err, "go mod tidy failed: %s", tidyOutput)

	binPath := filepath.Join(dir, "contract-plugin")
	buildCmd := exec.Command("go", "build", "-o", binPath, "main.go")
	buildCmd.Dir = dir
	buildOutput, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "go build failed: %s", buildOutput)

	require.NoError(t, os.Chmod(binPath, 0755))
	return binPath
}

// TestPluginCommunicationContract tests the gRPC communication contract
func TestPluginCommunicationContract(t *testing.T) {
	plugin := v1.NewBasePlugin(&v1.PluginMetadata{
		Name:    "comm-test",
		Version: "1.0.0",
	})

	plugin.RegisterCommand("echo", v1.NewSimpleCommand(
		&v1.CommandInfo{Name: "echo"},
		func(ctx context.Context, req *v1.ExecuteRequest) (*v1.ExecuteResponse, error) {
			return &v1.ExecuteResponse{
				Success: true,
				Stdout:  []byte(req.Args[0]),
			}, nil
		},
	))

	t.Run("command execution preserves stdout", func(t *testing.T) {
		resp, err := plugin.ExecuteCommand(context.Background(), &v1.ExecuteRequest{
			Command: "echo",
			Args:    []string{"test message"},
		})

		require.NoError(t, err)
		assert.Equal(t, []byte("test message"), resp.Stdout)
	})

	t.Run("command execution handles errors", func(t *testing.T) {
		resp, err := plugin.ExecuteCommand(context.Background(), &v1.ExecuteRequest{
			Command: "nonexistent",
		})

		require.NoError(t, err)
		assert.False(t, resp.Success)
		assert.NotEmpty(t, resp.Error)
	})
}
