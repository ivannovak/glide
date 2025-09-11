package v1

import (
	"context"
	"fmt"
)

// InteractiveCommandHandler is the interface that interactive commands must implement
type InteractiveCommandHandler interface {
	StartInteractive(stream GlidePlugin_StartInteractiveServer) error
}

// CommandHandler is a unified interface for both regular and interactive commands
type CommandHandler interface {
	Info() *CommandInfo
	Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error)
}

// BasePlugin provides a base implementation of GlidePluginServer with automatic routing
type BasePlugin struct {
	UnimplementedGlidePluginServer

	// Plugin metadata
	metadata *PluginMetadata

	// Registered commands mapped by name
	commands map[string]CommandHandler

	// Configuration storage
	config map[string]interface{}
}

// NewBasePlugin creates a new base plugin with the given metadata
func NewBasePlugin(metadata *PluginMetadata) *BasePlugin {
	return &BasePlugin{
		metadata: metadata,
		commands: make(map[string]CommandHandler),
		config:   make(map[string]interface{}),
	}
}

// RegisterCommand registers a command handler with the plugin
func (p *BasePlugin) RegisterCommand(name string, handler CommandHandler) {
	p.commands[name] = handler
}

// GetMetadata returns the plugin metadata
func (p *BasePlugin) GetMetadata(ctx context.Context, _ *Empty) (*PluginMetadata, error) {
	return p.metadata, nil
}

// Configure handles plugin configuration
func (p *BasePlugin) Configure(ctx context.Context, req *ConfigureRequest) (*ConfigureResponse, error) {
	// Convert map[string]string to map[string]interface{}
	p.config = make(map[string]interface{})
	for k, v := range req.Config {
		p.config[k] = v
	}

	return &ConfigureResponse{
		Success: true,
		Message: fmt.Sprintf("%s plugin configured successfully", p.metadata.Name),
	}, nil
}

// ListCommands returns all registered commands
func (p *BasePlugin) ListCommands(ctx context.Context, _ *Empty) (*CommandList, error) {
	var cmdList []*CommandInfo

	for _, handler := range p.commands {
		cmdList = append(cmdList, handler.Info())
	}

	return &CommandList{
		Commands: cmdList,
	}, nil
}

// ExecuteCommand executes a registered command
func (p *BasePlugin) ExecuteCommand(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error) {
	handler, ok := p.commands[req.Command]
	if !ok {
		return &ExecuteResponse{
			Success: false,
			Error:   fmt.Sprintf("unknown command: %s", req.Command),
		}, nil
	}

	return handler.Execute(ctx, req)
}

// StartInteractive automatically routes to the correct interactive command handler
func (p *BasePlugin) StartInteractive(stream GlidePlugin_StartInteractiveServer) error {
	// Receive initial request to get command name
	msg, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive initial message: %w", err)
	}

	// Extract command name from the initial message
	cmdName := string(msg.Data)
	if cmdName == "" {
		// Fallback: check if it's in the Signal field
		cmdName = msg.Signal
	}

	// If still no command name, return error
	if cmdName == "" {
		return fmt.Errorf("no command specified in initial message")
	}

	// Find the command handler
	handler, ok := p.commands[cmdName]
	if !ok {
		return fmt.Errorf("unknown command: %s", cmdName)
	}

	// Check if it supports interactive mode
	interactiveHandler, ok := handler.(InteractiveCommandHandler)
	if !ok {
		// Check if the command info says it's interactive
		if info := handler.Info(); info != nil && info.Interactive {
			return fmt.Errorf("command %s is marked as interactive but doesn't implement InteractiveCommandHandler", cmdName)
		}
		return fmt.Errorf("command %s does not support interactive mode", cmdName)
	}

	// Delegate to the command's interactive handler
	return interactiveHandler.StartInteractive(stream)
}

// GetConfig returns the current configuration
func (p *BasePlugin) GetConfig() map[string]interface{} {
	return p.config
}

// SimpleCommand is a helper struct for non-interactive commands
type SimpleCommand struct {
	info    *CommandInfo
	handler func(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error)
}

// NewSimpleCommand creates a new simple command
func NewSimpleCommand(info *CommandInfo, handler func(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error)) *SimpleCommand {
	return &SimpleCommand{
		info:    info,
		handler: handler,
	}
}

// Info returns the command information
func (c *SimpleCommand) Info() *CommandInfo {
	return c.info
}

// Execute runs the command handler
func (c *SimpleCommand) Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error) {
	return c.handler(ctx, req)
}

// BaseInteractiveCommand combines CommandHandler with InteractiveCommandHandler
type BaseInteractiveCommand struct {
	info               *CommandInfo
	executeHandler     func(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error)
	interactiveHandler func(stream GlidePlugin_StartInteractiveServer) error
}

// NewBaseInteractiveCommand creates a new interactive command
func NewBaseInteractiveCommand(
	info *CommandInfo,
	executeHandler func(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error),
	interactiveHandler func(stream GlidePlugin_StartInteractiveServer) error,
) *BaseInteractiveCommand {
	// Ensure the command is marked as interactive
	info.Interactive = true
	return &BaseInteractiveCommand{
		info:               info,
		executeHandler:     executeHandler,
		interactiveHandler: interactiveHandler,
	}
}

// Info returns the command information
func (c *BaseInteractiveCommand) Info() *CommandInfo {
	return c.info
}

// Execute runs the command handler
func (c *BaseInteractiveCommand) Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error) {
	if c.executeHandler != nil {
		return c.executeHandler(ctx, req)
	}
	// Default behavior for interactive commands
	return &ExecuteResponse{
		RequiresInteractive: true,
	}, nil
}

// StartInteractive handles the interactive session
func (c *BaseInteractiveCommand) StartInteractive(stream GlidePlugin_StartInteractiveServer) error {
	return c.interactiveHandler(stream)
}
