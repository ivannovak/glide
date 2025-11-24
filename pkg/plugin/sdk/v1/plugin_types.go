package v1

import (
	"context"
	"fmt"
	"os"

	plugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// HandshakeConfig is the handshake configuration for plugins
var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "GLIDE_PLUGIN_MAGIC",
	MagicCookieValue: "d3b07384-d9a7-4e0b-9c0a-7c9e9b9c9e9e",
}

// PluginMap is the plugin map for Glide
var PluginMap = map[string]plugin.Plugin{
	"glide": &GlidePluginImpl{},
}

// GlidePluginImpl is the gRPC implementation of the plugin
type GlidePluginImpl struct {
	plugin.Plugin
	Impl GlidePluginServer
}

func (p *GlidePluginImpl) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	RegisterGlidePluginServer(s, p.Impl)
	return nil
}

func (p *GlidePluginImpl) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return NewGlidePluginClient(c), nil
}

// InteractiveSession represents an interactive command session
type InteractiveSession interface {
	Send(*StreamMessage) error
	Recv() (*StreamMessage, error)
	Close() error
}

// ProjectContext provides project information
type ProjectContext struct {
	Root        string
	WorkingDir  string
	Development string
}

// DockerRequest represents a Docker operation request
type DockerRequest struct {
	Operation   string
	Service     string
	Command     []string
	Args        []string
	WorkDir     string
	Interactive bool
	TTY         bool
}

// DockerResponse represents a Docker operation response
type DockerResponse struct {
	Success  bool
	Output   []byte
	Error    string
	ExitCode int
}

// ExecResult contains command execution results
type ExecResult struct {
	ExitCode int
	Stdout   []byte
	Stderr   []byte
	Error    error
}

// Host provides access to host capabilities for plugins
type Host struct {
	// These would be implemented by the host side
}

// ExecuteDockerInteractive executes an interactive Docker command
func (h *Host) ExecuteDockerInteractive(ctx context.Context, req *DockerRequest) (InteractiveSession, error) {
	// This would be implemented by the host
	return nil, fmt.Errorf("not implemented")
}

// RunPlugin starts a plugin server with the given implementation.
// This is a convenience function that handles the boilerplate of setting up
// the hashicorp/go-plugin server with the correct configuration.
func RunPlugin(impl GlidePluginServer) error {
	// Verify we're being run as a plugin
	if os.Getenv("GLIDE_PLUGIN_MAGIC") == "" {
		return fmt.Errorf("this binary must be run as a Glide plugin")
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"glide": &GlidePluginImpl{Impl: impl},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})

	return nil
}
