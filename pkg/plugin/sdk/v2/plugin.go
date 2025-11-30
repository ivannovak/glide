// Package v2 provides the next-generation Glide Plugin SDK with improved type safety,
// simplified lifecycle management, and better developer experience.
//
// Key improvements over v1:
//   - Type-safe configuration using Go generics
//   - Unified lifecycle management across in-process and gRPC plugins
//   - Declarative command definition
//   - Simplified plugin development with sensible defaults
//   - Full backward compatibility via adapter layer
package v2

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// Plugin is the core v2 plugin interface that all plugins must implement.
// It uses Go generics to provide type-safe configuration.
//
// Example:
//
//	type MyConfig struct {
//	    APIKey string `json:"apiKey"`
//	    Timeout int   `json:"timeout"`
//	}
//
//	type MyPlugin struct {
//	    v2.BasePlugin[MyConfig]
//	}
//
//	func (p *MyPlugin) Metadata() v2.Metadata {
//	    return v2.Metadata{
//	        Name: "my-plugin",
//	        Version: "1.0.0",
//	        Description: "My awesome plugin",
//	    }
//	}
type Plugin[C any] interface {
	// Metadata returns plugin identification and metadata.
	// This is called before configuration and should not have side effects.
	Metadata() Metadata

	// ConfigSchema returns the JSON schema for the plugin's configuration.
	// This enables validation and IDE autocomplete for plugin configuration.
	// Return nil if the plugin doesn't require configuration.
	// The schema should be a JSON Schema object (map[string]interface{}).
	ConfigSchema() map[string]interface{}

	// Configure is called after the plugin is loaded with the validated configuration.
	// The config parameter is type-safe and guaranteed to match ConfigSchema.
	// Return an error if the configuration is invalid or initialization fails.
	Configure(ctx context.Context, config C) error

	// Lifecycle provides hooks for plugin initialization, startup, and shutdown.
	Lifecycle

	// Commands returns the list of commands this plugin provides.
	// Commands are automatically registered with the CLI.
	Commands() []Command
}

// Metadata contains plugin identification and descriptive information.
type Metadata struct {
	// Name is the unique identifier for the plugin (required).
	// Must be lowercase, alphanumeric with hyphens (e.g., "my-plugin").
	Name string

	// Version is the semantic version of the plugin (required).
	// Must follow semver format (e.g., "1.2.3").
	Version string

	// Description is a short description of what the plugin does.
	Description string

	// Author is the plugin author or organization.
	Author string

	// Homepage is the URL to the plugin's homepage or repository.
	Homepage string

	// License is the SPDX license identifier (e.g., "MIT", "Apache-2.0").
	License string

	// Tags are keywords for plugin discovery and categorization.
	Tags []string

	// Dependencies lists other plugins this plugin depends on.
	Dependencies []Dependency

	// Capabilities declares what system resources the plugin needs.
	Capabilities Capabilities
}

// Dependency represents a dependency on another plugin.
type Dependency struct {
	// Name is the plugin name to depend on.
	Name string

	// Version is a semver constraint (e.g., "^1.0.0", ">=2.0.0 <3.0.0").
	// Empty means any version is acceptable.
	Version string

	// Optional indicates this dependency is optional.
	// The plugin should handle the case where the dependency is not available.
	Optional bool
}

// Capabilities declares what system resources a plugin requires.
// This allows the CLI to check prerequisites before loading the plugin.
type Capabilities struct {
	// RequiresDocker indicates the plugin needs Docker.
	RequiresDocker bool

	// RequiresNetwork indicates the plugin needs network access.
	RequiresNetwork bool

	// RequiresFilesystem indicates the plugin needs filesystem access.
	RequiresFilesystem bool

	// RequiresInteractive indicates the plugin requires an interactive terminal.
	RequiresInteractive bool

	// RequiredCommands lists external commands that must be available.
	RequiredCommands []string

	// RequiredPaths lists filesystem paths that must exist.
	RequiredPaths []string

	// RequiredEnvVars lists environment variables that must be set.
	RequiredEnvVars []string
}

// Lifecycle provides hooks for plugin initialization and shutdown.
// All lifecycle methods are called with a context that may be cancelled.
type Lifecycle interface {
	// Init is called once when the plugin is first loaded.
	// Use this to perform one-time initialization that doesn't depend on configuration.
	// Return an error to prevent the plugin from loading.
	Init(ctx context.Context) error

	// Start is called after Configure and before the plugin is used.
	// Use this to establish connections, start background workers, etc.
	// Return an error to prevent the plugin from starting.
	Start(ctx context.Context) error

	// Stop is called when the plugin is being unloaded or the application is shutting down.
	// Use this to close connections, stop workers, and clean up resources.
	// This method should be idempotent and safe to call multiple times.
	Stop(ctx context.Context) error

	// HealthCheck returns nil if the plugin is healthy, or an error describing the problem.
	// This is called periodically to check plugin health.
	HealthCheck(ctx context.Context) error
}

// Command represents a CLI command provided by a plugin.
type Command struct {
	// Name is the command name (required).
	// This is what users type to invoke the command.
	Name string

	// Description is a short description shown in help text.
	Description string

	// Category groups related commands together in help output.
	// Common categories: "core", "setup", "testing", "database", etc.
	Category string

	// Aliases are alternative names for the command.
	Aliases []string

	// Hidden commands don't appear in help but can still be invoked.
	Hidden bool

	// Interactive indicates this command requires an interactive terminal.
	Interactive bool

	// Flags defines command-line flags for this command.
	Flags []Flag

	// Args defines positional arguments for this command.
	Args []Arg

	// Handler is called when the command is executed.
	// For in-process plugins, this is a direct function call.
	// For gRPC plugins, this is dispatched via RPC.
	Handler CommandHandler

	// InteractiveHandler is used for commands that need terminal control.
	// Set Interactive=true and provide this handler instead of Handler.
	InteractiveHandler InteractiveCommandHandler

	// RequiresTTY indicates the command requires a TTY (terminal).
	RequiresTTY bool

	// RequiresAuth indicates the command requires authentication.
	RequiresAuth bool

	// Visibility controls where the command is available.
	// Options: "always", "project-only", "worktree-only", "root-only", "non-root"
	Visibility string
}

// Flag represents a command-line flag.
type Flag struct {
	// Name is the flag name without dashes (e.g., "output" for --output).
	Name string

	// Shorthand is the single-letter shorthand (e.g., "o" for -o).
	Shorthand string

	// Description is shown in help text.
	Description string

	// Type is the value type: "string", "int", "bool", "stringSlice", etc.
	Type string

	// Default is the default value if the flag is not provided.
	Default interface{}

	// Required indicates this flag must be provided.
	Required bool

	// Deprecated marks a flag as deprecated with a message.
	Deprecated string
}

// Arg represents a positional command argument.
type Arg struct {
	// Name is the argument name shown in usage text.
	Name string

	// Description is shown in help text.
	Description string

	// Required indicates this argument must be provided.
	Required bool

	// Variadic indicates this argument accepts multiple values.
	Variadic bool
}

// CommandHandler handles non-interactive command execution.
type CommandHandler interface {
	// Execute runs the command with the given context and arguments.
	Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error)
}

// InteractiveCommandHandler handles interactive command execution with terminal control.
type InteractiveCommandHandler interface {
	// ExecuteInteractive runs an interactive command with full terminal control.
	ExecuteInteractive(ctx context.Context, session *InteractiveSession) error
}

// ExecuteRequest contains the command execution context.
type ExecuteRequest struct {
	// Command is the command name that was invoked.
	Command string

	// Args are the positional arguments passed to the command.
	Args []string

	// Flags are the parsed command-line flags.
	Flags map[string]interface{}

	// Env contains environment variables relevant to the command.
	Env map[string]string

	// WorkingDir is the current working directory.
	WorkingDir string
}

// ExecuteResponse contains the command execution result.
type ExecuteResponse struct {
	// ExitCode is the command exit code (0 for success).
	ExitCode int

	// Output is the command output (combined stdout/stderr).
	Output string

	// Error is a human-readable error message if the command failed.
	Error string

	// Data is optional structured data the command wants to return.
	Data interface{}
}

// InteractiveSession provides terminal control for interactive commands.
type InteractiveSession interface {
	// Context returns the command execution context.
	Context() context.Context

	// Request returns the original execute request.
	Request() *ExecuteRequest

	// ReadLine reads a line of input from the terminal.
	ReadLine() (string, error)

	// WriteLine writes a line to the terminal.
	WriteLine(line string) error

	// ReadByte reads a single byte (for character-by-character input).
	ReadByte() (byte, error)

	// WriteByte writes a single byte.
	WriteByte(b byte) error

	// SetRaw enables/disables raw terminal mode.
	SetRaw(enabled bool) error

	// GetSize returns the terminal size (width, height).
	GetSize() (int, int, error)

	// OnResize registers a callback for terminal resize events.
	OnResize(func(width, height int))

	// Close closes the interactive session.
	Close() error
}

// BasePlugin provides a default implementation of the Plugin interface.
// Plugins can embed this to get sensible defaults for all methods.
//
// Example:
//
//	type MyPlugin struct {
//	    v2.BasePlugin[MyConfig]
//	    apiClient *APIClient
//	}
//
//	func (p *MyPlugin) Configure(ctx context.Context, config MyConfig) error {
//	    p.apiClient = NewAPIClient(config.APIKey)
//	    return nil
//	}
type BasePlugin[C any] struct {
	metadata Metadata
	config   C
	commands []Command
}

// Ensure BasePlugin implements Plugin interface.
var _ Plugin[struct{}] = (*BasePlugin[struct{}])(nil)

// Metadata returns the plugin metadata set via SetMetadata.
func (p *BasePlugin[C]) Metadata() Metadata {
	return p.metadata
}

// SetMetadata sets the plugin metadata.
// Call this from your plugin's constructor.
func (p *BasePlugin[C]) SetMetadata(m Metadata) {
	p.metadata = m
}

// ConfigSchema returns nil by default.
// Override this to provide a JSON schema for validation and documentation.
func (p *BasePlugin[C]) ConfigSchema() map[string]interface{} {
	return nil
}

// Configure stores the configuration.
// Override this to perform configuration-dependent initialization.
func (p *BasePlugin[C]) Configure(_ context.Context, config C) error {
	p.config = config
	return nil
}

// Config returns the current configuration.
func (p *BasePlugin[C]) Config() C {
	return p.config
}

// Init performs no operation by default.
// Override this to perform initialization.
func (p *BasePlugin[C]) Init(_ context.Context) error {
	return nil
}

// Start performs no operation by default.
// Override this to perform startup logic.
func (p *BasePlugin[C]) Start(_ context.Context) error {
	return nil
}

// Stop performs no operation by default.
// Override this to perform cleanup.
func (p *BasePlugin[C]) Stop(_ context.Context) error {
	return nil
}

// HealthCheck returns nil by default.
// Override this to implement health checks.
func (p *BasePlugin[C]) HealthCheck(_ context.Context) error {
	return nil
}

// Commands returns the commands set via SetCommands or AddCommand.
func (p *BasePlugin[C]) Commands() []Command {
	return p.commands
}

// SetCommands sets the full list of commands.
func (p *BasePlugin[C]) SetCommands(commands []Command) {
	p.commands = commands
}

// AddCommand adds a single command.
func (p *BasePlugin[C]) AddCommand(cmd Command) {
	p.commands = append(p.commands, cmd)
}

// SimpleCommandHandler creates a CommandHandler from a simple function.
// This is a convenience helper for commands that don't need complex handling.
//
// Example:
//
//	cmd := v2.Command{
//	    Name: "hello",
//	    Description: "Say hello",
//	    Handler: v2.SimpleCommandHandler(func(ctx context.Context, req *v2.ExecuteRequest) (*v2.ExecuteResponse, error) {
//	        return &v2.ExecuteResponse{
//	            ExitCode: 0,
//	            Output: "Hello, world!",
//	        }, nil
//	    }),
//	}
type SimpleCommandHandler func(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error)

// Execute implements CommandHandler.
func (h SimpleCommandHandler) Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error) {
	return h(ctx, req)
}

// CobraAdapter adapts a v2 Plugin to work with Cobra commands.
// This is used internally by the CLI to bridge v2 plugins to the existing Cobra infrastructure.
type CobraAdapter[C any] struct {
	plugin Plugin[C]
}

// NewCobraAdapter creates a new adapter for a v2 plugin.
func NewCobraAdapter[C any](plugin Plugin[C]) *CobraAdapter[C] {
	return &CobraAdapter[C]{plugin: plugin}
}

// BuildCommands converts v2 Commands to Cobra commands.
func (a *CobraAdapter[C]) BuildCommands() []*cobra.Command {
	commands := make([]*cobra.Command, 0, len(a.plugin.Commands()))

	for _, cmd := range a.plugin.Commands() {
		cobraCmd := &cobra.Command{
			Use:     cmd.Name,
			Short:   cmd.Description,
			Aliases: cmd.Aliases,
			Hidden:  cmd.Hidden,
			RunE: func(cobraCmd *cobra.Command, args []string) error {
				return a.executeCommand(cobraCmd.Context(), cmd, args, cobraCmd)
			},
		}

		// Add flags
		for _, flag := range cmd.Flags {
			a.addFlag(cobraCmd, flag)
		}

		commands = append(commands, cobraCmd)
	}

	return commands
}

func (a *CobraAdapter[C]) executeCommand(ctx context.Context, cmd Command, args []string, cobraCmd *cobra.Command) error {
	// Get working directory
	workingDir, err := os.Getwd()
	if err != nil {
		workingDir = "."
	}

	// Build execute request
	req := &ExecuteRequest{
		Command:    cmd.Name,
		Args:       args,
		Flags:      make(map[string]interface{}),
		Env:        make(map[string]string),
		WorkingDir: workingDir,
	}

	// Extract flags based on their declared types
	for _, flag := range cmd.Flags {
		if cobraCmd.Flags().Changed(flag.Name) {
			var value interface{}
			var flagErr error
			switch flag.Type {
			case "string":
				value, flagErr = cobraCmd.Flags().GetString(flag.Name)
			case "bool":
				value, flagErr = cobraCmd.Flags().GetBool(flag.Name)
			case "int":
				value, flagErr = cobraCmd.Flags().GetInt(flag.Name)
			case "stringSlice":
				value, flagErr = cobraCmd.Flags().GetStringSlice(flag.Name)
			case "intSlice":
				value, flagErr = cobraCmd.Flags().GetIntSlice(flag.Name)
			case "float64":
				value, flagErr = cobraCmd.Flags().GetFloat64(flag.Name)
			case "duration":
				value, flagErr = cobraCmd.Flags().GetDuration(flag.Name)
			default:
				// Default to string for unknown types
				value, flagErr = cobraCmd.Flags().GetString(flag.Name)
			}
			if flagErr == nil {
				req.Flags[flag.Name] = value
			}
		}
	}

	// Interactive commands require InteractiveHandler
	if cmd.Interactive && cmd.InteractiveHandler != nil {
		return fmt.Errorf("interactive commands are not supported via CobraAdapter; use gRPC plugin mode for interactive commands")
	}

	if cmd.Handler != nil {
		resp, err := cmd.Handler.Execute(ctx, req)
		if err != nil {
			return err
		}
		if resp.ExitCode != 0 {
			return &CommandError{
				Command:  cmd.Name,
				ExitCode: resp.ExitCode,
				Message:  resp.Error,
			}
		}
	}

	return nil
}

func (a *CobraAdapter[C]) addFlag(cmd *cobra.Command, flag Flag) {
	a.addFlagByType(cmd, flag)

	if flag.Required {
		_ = cmd.MarkFlagRequired(flag.Name)
	}
}

// addFlagByType adds a flag to the command based on its type.
// This is separated from addFlag to reduce cyclomatic complexity.
func (a *CobraAdapter[C]) addFlagByType(cmd *cobra.Command, flag Flag) {
	switch flag.Type {
	case "string":
		defaultValue := getTypedDefault(flag.Default, "")
		if flag.Shorthand != "" {
			cmd.Flags().StringP(flag.Name, flag.Shorthand, defaultValue, flag.Description)
		} else {
			cmd.Flags().String(flag.Name, defaultValue, flag.Description)
		}
	case "bool":
		defaultValue := getTypedDefault(flag.Default, false)
		if flag.Shorthand != "" {
			cmd.Flags().BoolP(flag.Name, flag.Shorthand, defaultValue, flag.Description)
		} else {
			cmd.Flags().Bool(flag.Name, defaultValue, flag.Description)
		}
	case "int":
		defaultValue := getTypedDefault(flag.Default, 0)
		if flag.Shorthand != "" {
			cmd.Flags().IntP(flag.Name, flag.Shorthand, defaultValue, flag.Description)
		} else {
			cmd.Flags().Int(flag.Name, defaultValue, flag.Description)
		}
	case "stringSlice":
		defaultValue := getTypedDefault(flag.Default, []string(nil))
		if flag.Shorthand != "" {
			cmd.Flags().StringSliceP(flag.Name, flag.Shorthand, defaultValue, flag.Description)
		} else {
			cmd.Flags().StringSlice(flag.Name, defaultValue, flag.Description)
		}
	case "intSlice":
		defaultValue := getTypedDefault(flag.Default, []int(nil))
		if flag.Shorthand != "" {
			cmd.Flags().IntSliceP(flag.Name, flag.Shorthand, defaultValue, flag.Description)
		} else {
			cmd.Flags().IntSlice(flag.Name, defaultValue, flag.Description)
		}
	case "float64":
		defaultValue := getTypedDefault(flag.Default, 0.0)
		if flag.Shorthand != "" {
			cmd.Flags().Float64P(flag.Name, flag.Shorthand, defaultValue, flag.Description)
		} else {
			cmd.Flags().Float64(flag.Name, defaultValue, flag.Description)
		}
	case "duration":
		defaultValue := getTypedDefault(flag.Default, time.Duration(0))
		if flag.Shorthand != "" {
			cmd.Flags().DurationP(flag.Name, flag.Shorthand, defaultValue, flag.Description)
		} else {
			cmd.Flags().Duration(flag.Name, defaultValue, flag.Description)
		}
	default:
		// Default to string for unknown types
		defaultValue := ""
		if flag.Default != nil {
			defaultValue = fmt.Sprintf("%v", flag.Default)
		}
		if flag.Shorthand != "" {
			cmd.Flags().StringP(flag.Name, flag.Shorthand, defaultValue, flag.Description)
		} else {
			cmd.Flags().String(flag.Name, defaultValue, flag.Description)
		}
	}
}

// getTypedDefault safely extracts a typed default value from an interface{}.
func getTypedDefault[T any](value interface{}, fallback T) T {
	if value == nil {
		return fallback
	}
	if typed, ok := value.(T); ok {
		return typed
	}
	return fallback
}

// CommandError represents a command execution error.
type CommandError struct {
	Command  string
	ExitCode int
	Message  string
}

func (e *CommandError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "command failed with exit code " + string(rune(e.ExitCode))
}
