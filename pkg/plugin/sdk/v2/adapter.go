// Package v2 provides backward compatibility with v1 plugins.
package v2

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ivannovak/glide/v2/pkg/plugin/sdk"
	v1 "github.com/ivannovak/glide/v2/pkg/plugin/sdk/v1"
)

// V1Adapter wraps a v1 plugin to implement the v2 Plugin interface.
// This allows v1 plugins to work seamlessly in a v2 environment.
//
// The adapter handles:
//   - Converting v1 metadata to v2 format
//   - Mapping v1 command handlers to v2 command system
//   - Bridging v1 lifecycle to v2 lifecycle
//   - Converting v1 configuration to v2 type-safe config
//
// Usage:
//
//	v1Plugin := &myV1Plugin{}
//	v2Plugin := v2.AdaptV1Plugin(v1Plugin)
type V1Adapter struct {
	v1Plugin interface{} // Can be v1.GlidePlugin or pkg/plugin.Plugin
	metadata Metadata
	commands []Command
	state    *sdk.StateTracker
}

// AdaptV1GRPCPlugin wraps a v1 gRPC plugin (v1.GlidePlugin) for v2 compatibility.
func AdaptV1GRPCPlugin(v1Plugin v1.GlidePluginClient) Plugin[map[string]interface{}] {
	adapter := &V1Adapter{
		v1Plugin: v1Plugin,
		state:    sdk.NewStateTracker("v1-adapter"),
	}

	// Fetch metadata from v1 plugin
	ctx := context.Background()
	if v1Meta, err := v1Plugin.GetMetadata(ctx, &v1.Empty{}); err == nil {
		adapter.metadata = convertV1Metadata(v1Meta)
	}

	// Fetch commands from v1 plugin
	if v1Commands, err := v1Plugin.ListCommands(ctx, &v1.Empty{}); err == nil {
		adapter.commands = convertV1Commands(v1Commands.Commands)
	}

	return adapter
}

// AdaptV1InProcessPlugin wraps a v1 in-process plugin for v2 compatibility.
// The v1 in-process plugin interface is defined in pkg/plugin/interface.go.
func AdaptV1InProcessPlugin(v1Plugin interface{}) Plugin[map[string]interface{}] {
	adapter := &V1Adapter{
		v1Plugin: v1Plugin,
		state:    sdk.NewStateTracker("v1-adapter"),
	}

	// Extract metadata using reflection or type assertion
	// The v1 in-process plugin has: Name(), Version(), Description() methods
	if metadataProvider, ok := v1Plugin.(interface {
		Name() string
		Version() string
		Description() string
	}); ok {
		adapter.metadata = Metadata{
			Name:        metadataProvider.Name(),
			Version:     metadataProvider.Version(),
			Description: metadataProvider.Description(),
		}
	}

	// v1 in-process plugins register commands directly with Cobra
	// We can't easily extract them, so commands will be empty
	// The plugin will use the old Register(root *cobra.Command) path
	adapter.commands = []Command{}

	return adapter
}

// Metadata returns the adapted v1 metadata.
func (a *V1Adapter) Metadata() Metadata {
	return a.metadata
}

// ConfigSchema returns nil for v1 plugins (no schema available).
func (a *V1Adapter) ConfigSchema() map[string]interface{} {
	return nil
}

// Configure passes configuration to the v1 plugin.
func (a *V1Adapter) Configure(ctx context.Context, config map[string]interface{}) error {
	switch p := a.v1Plugin.(type) {
	case v1.GlidePluginClient:
		// Convert map to v1 ConfigureRequest
		configMap := make(map[string]string)
		for k, v := range config {
			configMap[k] = fmt.Sprintf("%v", v)
		}
		req := &v1.ConfigureRequest{Config: configMap}
		resp, err := p.Configure(ctx, req)
		if err != nil {
			return err
		}
		if !resp.Success {
			return fmt.Errorf("v1 plugin configuration failed: %s", resp.Message)
		}
		return nil

	case interface{ Configure() error }:
		// v1 in-process plugin with Configure method
		// The config is already loaded by pkg/config, so just call Configure()
		return p.Configure()

	default:
		// Plugin doesn't support configuration
		return nil
	}
}

// Init delegates to v1 lifecycle if available.
func (a *V1Adapter) Init(ctx context.Context) error {
	// v1 gRPC plugins don't have Init in the protocol
	// v1 in-process plugins might have it via sdk.Lifecycle
	if lifecycle, ok := a.v1Plugin.(sdk.Lifecycle); ok {
		if err := lifecycle.Init(ctx); err != nil {
			a.state.ForceSet(sdk.StateErrored)
			return err
		}
		return a.state.Set(sdk.StateInitialized)
	}

	// No init method, just mark as initialized
	return a.state.Set(sdk.StateInitialized)
}

// Start delegates to v1 lifecycle if available.
func (a *V1Adapter) Start(ctx context.Context) error {
	if lifecycle, ok := a.v1Plugin.(sdk.Lifecycle); ok {
		if err := lifecycle.Start(ctx); err != nil {
			a.state.ForceSet(sdk.StateErrored)
			return err
		}
		return a.state.Set(sdk.StateStarted)
	}

	// No start method, just mark as started
	return a.state.Set(sdk.StateStarted)
}

// Stop delegates to v1 lifecycle if available.
func (a *V1Adapter) Stop(ctx context.Context) error {
	if lifecycle, ok := a.v1Plugin.(sdk.Lifecycle); ok {
		err := lifecycle.Stop(ctx)
		a.state.ForceSet(sdk.StateStopped)
		return err
	}

	// No stop method, just mark as stopped
	a.state.ForceSet(sdk.StateStopped)
	return nil
}

// HealthCheck delegates to v1 lifecycle if available.
func (a *V1Adapter) HealthCheck(_ context.Context) error {
	if lifecycle, ok := a.v1Plugin.(sdk.Lifecycle); ok {
		return lifecycle.HealthCheck()
	}

	// No health check, assume healthy
	return nil
}

// Commands returns the converted v1 commands.
func (a *V1Adapter) Commands() []Command {
	return a.commands
}

// GetV1Plugin returns the underlying v1 plugin for special handling.
func (a *V1Adapter) GetV1Plugin() interface{} {
	return a.v1Plugin
}

// convertV1Metadata converts v1 protobuf metadata to v2 Metadata.
func convertV1Metadata(v1Meta *v1.PluginMetadata) Metadata {
	meta := Metadata{
		Name:        v1Meta.Name,
		Version:     v1Meta.Version,
		Description: v1Meta.Description,
		Author:      v1Meta.Author,
		Homepage:    v1Meta.Homepage,
		License:     v1Meta.License,
		Tags:        v1Meta.Tags,
	}

	// Convert dependencies
	if len(v1Meta.Dependencies) > 0 {
		meta.Dependencies = make([]Dependency, len(v1Meta.Dependencies))
		for i, dep := range v1Meta.Dependencies {
			meta.Dependencies[i] = Dependency{
				Name:     dep.Name,
				Version:  dep.Version,
				Optional: dep.Optional,
			}
		}
	}

	// Note: Capabilities are fetched via separate RPC call in v1 (GetCapabilities)
	// We don't set them here since they're not part of PluginMetadata

	return meta
}

// convertV1Commands converts v1 protobuf commands to v2 Commands.
func convertV1Commands(v1Commands []*v1.CommandInfo) []Command {
	commands := make([]Command, len(v1Commands))

	for i, v1Cmd := range v1Commands {
		commands[i] = Command{
			Name:         v1Cmd.Name,
			Description:  v1Cmd.Description,
			Category:     v1Cmd.Category,
			Aliases:      v1Cmd.Aliases,
			Hidden:       v1Cmd.Hidden,
			Interactive:  v1Cmd.Interactive,
			RequiresTTY:  v1Cmd.RequiresTty,
			RequiresAuth: v1Cmd.RequiresAuth,
			Visibility:   v1Cmd.Visibility,
			// Handler will be set up separately by the CLI
			// since it needs to dispatch to the v1 plugin
		}
	}

	return commands
}

// V1CommandAdapter wraps a v1 command handler to implement v2 CommandHandler.
type V1CommandAdapter struct {
	v1Plugin v1.GlidePluginClient
	command  string
}

// NewV1CommandAdapter creates an adapter for a v1 command.
func NewV1CommandAdapter(v1Plugin v1.GlidePluginClient, command string) CommandHandler {
	return &V1CommandAdapter{
		v1Plugin: v1Plugin,
		command:  command,
	}
}

// Execute adapts v2 ExecuteRequest to v1 and back.
func (a *V1CommandAdapter) Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error) {
	// Convert v2 request to v1
	v1Req := &v1.ExecuteRequest{
		Command: req.Command,
		Args:    req.Args,
		Env:     req.Env,
		WorkDir: req.WorkingDir,
	}

	// Execute via v1 plugin
	v1Resp, err := a.v1Plugin.ExecuteCommand(ctx, v1Req)
	if err != nil {
		return nil, err
	}

	// Convert v1 response to v2
	// Combine stdout and stderr into single output
	output := string(v1Resp.Stdout)
	if len(v1Resp.Stderr) > 0 {
		if len(output) > 0 {
			output += "\n"
		}
		output += string(v1Resp.Stderr)
	}

	v2Resp := &ExecuteResponse{
		ExitCode: int(v1Resp.ExitCode),
		Output:   output,
		Error:    v1Resp.Error,
	}

	return v2Resp, nil
}

// V1InteractiveCommandAdapter wraps a v1 interactive command handler.
// Note: Interactive command adaptation between v1 (bidirectional streaming) and v2 (session-based)
// is not supported. V1 plugins with interactive commands should remain as v1 plugins or be
// rewritten natively in v2.
type V1InteractiveCommandAdapter struct {
	v1Plugin v1.GlidePluginClient
	command  string
}

// NewV1InteractiveCommandAdapter creates an adapter for a v1 interactive command.
// Note: This adapter does not support interactive execution. Use V1CommandAdapter for
// non-interactive commands, or keep interactive plugins as v1.
func NewV1InteractiveCommandAdapter(v1Plugin v1.GlidePluginClient, command string) InteractiveCommandHandler {
	return &V1InteractiveCommandAdapter{
		v1Plugin: v1Plugin,
		command:  command,
	}
}

// ExecuteInteractive is not supported for v1 adapted plugins.
// V1 plugins use bidirectional gRPC streaming for interactive commands, which is architecturally
// incompatible with the v2 session-based InteractiveSession interface. Interactive v1 plugins
// should either remain as v1 or be rewritten natively in v2.
func (a *V1InteractiveCommandAdapter) ExecuteInteractive(ctx context.Context, session *InteractiveSession) error {
	return fmt.Errorf("interactive command adaptation from v1 to v2 is not supported: v1 plugin %q command %q uses bidirectional streaming which cannot be adapted to v2 session interface; keep this plugin as v1 or rewrite it natively in v2", a.v1Plugin, a.command)
}

// V2ToV1Adapter wraps a v2 plugin to implement v1 interfaces.
// This allows v2 plugins to be used in v1 contexts during migration.
type V2ToV1Adapter[C any] struct {
	v2Plugin Plugin[C]
}

// AdaptV2ToV1 creates a v1-compatible adapter for a v2 plugin.
func AdaptV2ToV1[C any](v2Plugin Plugin[C]) *V2ToV1Adapter[C] {
	return &V2ToV1Adapter[C]{v2Plugin: v2Plugin}
}

// Name implements the v1 in-process PluginIdentifier interface.
func (a *V2ToV1Adapter[C]) Name() string {
	return a.v2Plugin.Metadata().Name
}

// Version implements the v1 in-process PluginIdentifier interface.
func (a *V2ToV1Adapter[C]) Version() string {
	return a.v2Plugin.Metadata().Version
}

// Description implements the v1 in-process PluginIdentifier interface.
func (a *V2ToV1Adapter[C]) Description() string {
	return a.v2Plugin.Metadata().Description
}

// Configure implements the v1 in-process PluginConfigurable interface.
func (a *V2ToV1Adapter[C]) Configure() error {
	// v2 plugins expect typed config in Configure(ctx, config)
	// For v1 compatibility, we assume config is already loaded
	// This is a simplified adapter - full implementation would need config loading
	ctx := context.Background()
	var emptyConfig C
	return a.v2Plugin.Configure(ctx, emptyConfig)
}

// Register implements the v1 in-process PluginRegistrar interface.
func (a *V2ToV1Adapter[C]) Register(root *cobra.Command) error {
	// Build v2 commands and add them to the root
	adapter := NewCobraAdapter(a.v2Plugin)
	commands := adapter.BuildCommands()
	for _, cmd := range commands {
		root.AddCommand(cmd)
	}
	return nil
}

// Init implements the v1 sdk.Lifecycle interface.
func (a *V2ToV1Adapter[C]) Init(ctx context.Context) error {
	return a.v2Plugin.Init(ctx)
}

// Start implements the v1 sdk.Lifecycle interface.
func (a *V2ToV1Adapter[C]) Start(ctx context.Context) error {
	return a.v2Plugin.Start(ctx)
}

// Stop implements the v1 sdk.Lifecycle interface.
func (a *V2ToV1Adapter[C]) Stop(ctx context.Context) error {
	return a.v2Plugin.Stop(ctx)
}

// HealthCheck implements the v1 sdk.Lifecycle interface.
func (a *V2ToV1Adapter[C]) HealthCheck() error {
	ctx := context.Background()
	return a.v2Plugin.HealthCheck(ctx)
}

// V2GRPCServer wraps a v2 plugin to implement the v1 GlidePluginServer interface.
// This allows v2 plugins to run as standalone gRPC plugin processes.
type V2GRPCServer[C any] struct {
	v1.UnimplementedGlidePluginServer
	v2Plugin Plugin[C]
}

// NewV2GRPCServer creates a gRPC server wrapper for a v2 plugin.
func NewV2GRPCServer[C any](plugin Plugin[C]) *V2GRPCServer[C] {
	return &V2GRPCServer[C]{v2Plugin: plugin}
}

// GetMetadata implements v1.GlidePluginServer.
func (s *V2GRPCServer[C]) GetMetadata(ctx context.Context, _ *v1.Empty) (*v1.PluginMetadata, error) {
	meta := s.v2Plugin.Metadata()
	return &v1.PluginMetadata{
		Name:        meta.Name,
		Version:     meta.Version,
		Description: meta.Description,
		Author:      meta.Author,
		Homepage:    meta.Homepage,
		License:     meta.License,
		Tags:        meta.Tags,
	}, nil
}

// Configure implements v1.GlidePluginServer.
// Note: v2 plugins use typed configuration loaded from .glide.yml by the host,
// not the string map passed via gRPC. The zero value is passed here because
// the actual typed config is set by the host calling Configure directly on the
// v2 plugin before starting the gRPC server.
func (s *V2GRPCServer[C]) Configure(ctx context.Context, req *v1.ConfigureRequest) (*v1.ConfigureResponse, error) {
	var config C
	if err := s.v2Plugin.Configure(ctx, config); err != nil {
		return &v1.ConfigureResponse{Success: false, Message: err.Error()}, nil
	}
	return &v1.ConfigureResponse{Success: true}, nil
}

// ListCommands implements v1.GlidePluginServer.
func (s *V2GRPCServer[C]) ListCommands(ctx context.Context, _ *v1.Empty) (*v1.CommandList, error) {
	v2Commands := s.v2Plugin.Commands()
	v1Commands := make([]*v1.CommandInfo, len(v2Commands))

	for i, cmd := range v2Commands {
		v1Commands[i] = &v1.CommandInfo{
			Name:         cmd.Name,
			Description:  cmd.Description,
			Category:     cmd.Category,
			Aliases:      cmd.Aliases,
			Hidden:       cmd.Hidden,
			Interactive:  cmd.Interactive,
			RequiresTty:  cmd.RequiresTTY,
			RequiresAuth: cmd.RequiresAuth,
			Visibility:   cmd.Visibility,
		}
	}

	return &v1.CommandList{Commands: v1Commands}, nil
}

// ExecuteCommand implements v1.GlidePluginServer.
func (s *V2GRPCServer[C]) ExecuteCommand(ctx context.Context, req *v1.ExecuteRequest) (*v1.ExecuteResponse, error) {
	// Find the command
	var handler CommandHandler
	for _, cmd := range s.v2Plugin.Commands() {
		if cmd.Name == req.Command {
			handler = cmd.Handler
			break
		}
	}

	if handler == nil {
		return &v1.ExecuteResponse{
			ExitCode: 1,
			Error:    fmt.Sprintf("unknown command: %s", req.Command),
		}, nil
	}

	// Convert v1 request to v2
	v2Req := &ExecuteRequest{
		Command:    req.Command,
		Args:       req.Args,
		Flags:      make(map[string]interface{}),
		Env:        req.Env,
		WorkingDir: req.WorkDir,
	}

	// Execute via v2 handler
	v2Resp, err := handler.Execute(ctx, v2Req)
	if err != nil {
		return &v1.ExecuteResponse{
			ExitCode: 1,
			Error:    err.Error(),
		}, nil
	}

	// Convert v2 response to v1
	// Safely convert exit code to int32 (exit codes are typically 0-255)
	exitCode := v2Resp.ExitCode
	if exitCode > 127 {
		exitCode = 127 // Cap at max positive int8 for shell compatibility
	} else if exitCode < -128 {
		exitCode = -128 // Cap at min int8
	}

	return &v1.ExecuteResponse{
		ExitCode: int32(exitCode), //nolint:gosec // exit codes are bounded above
		Stdout:   []byte(v2Resp.Output),
		Error:    v2Resp.Error,
	}, nil
}

// GetCapabilities implements v1.GlidePluginServer.
func (s *V2GRPCServer[C]) GetCapabilities(ctx context.Context, _ *v1.Empty) (*v1.Capabilities, error) {
	meta := s.v2Plugin.Metadata()
	return &v1.Capabilities{
		RequiresDocker:     meta.Capabilities.RequiresDocker,
		RequiresNetwork:    meta.Capabilities.RequiresNetwork,
		RequiresFilesystem: meta.Capabilities.RequiresFilesystem,
	}, nil
}

// GetCustomCategories implements v1.GlidePluginServer.
func (s *V2GRPCServer[C]) GetCustomCategories(ctx context.Context, _ *v1.Empty) (*v1.CategoryList, error) {
	// v2 plugins don't have custom categories in the same way
	return &v1.CategoryList{Categories: nil}, nil
}

// Serve starts a v2 plugin as a gRPC server using the v1 infrastructure.
// This is the main entry point for running a v2 plugin binary.
//
// Usage:
//
//	func main() {
//	    plugin := &MyPlugin{}
//	    if err := v2.Serve(plugin); err != nil {
//	        os.Exit(1)
//	    }
//	}
func Serve[C any](plugin Plugin[C]) error {
	server := NewV2GRPCServer(plugin)
	return v1.RunPlugin(server)
}
