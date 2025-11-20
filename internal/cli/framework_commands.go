package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ivannovak/glide/internal/detection"
	"github.com/ivannovak/glide/internal/shell"
	"github.com/ivannovak/glide/pkg/plugin/sdk"
	"github.com/spf13/cobra"
)

// FrameworkCommandInjector handles injection of framework-detected commands
type FrameworkCommandInjector struct {
	detector *detection.FrameworkDetector
	registry *Registry
	shell    *shell.Executor
}

// NewFrameworkCommandInjector creates a new command injector
func NewFrameworkCommandInjector(detector *detection.FrameworkDetector, reg *Registry) *FrameworkCommandInjector {
	return &FrameworkCommandInjector{
		detector: detector,
		registry: reg,
		shell:    shell.NewExecutor(shell.Options{}),
	}
}

// InjectCommands injects framework commands into the command registry
func (fci *FrameworkCommandInjector) InjectCommands(projectPath string) error {
	// Get framework commands
	commands := fci.detector.GetFrameworkCommands(projectPath)
	if len(commands) == 0 {
		return nil
	}

	// Register each command
	for name, def := range commands {
		// Skip if command already exists in registry
		if fci.registry.Has(name) {
			continue
		}

		// Capture variables for closure
		cmdName := name
		cmdDef := def

		// Create cobra command for the framework command
		cmdFactory := func() *cobra.Command {
			return fci.createFrameworkCommand(cmdName, cmdDef)
		}

		// Register with the registry
		metadata := Metadata{
			Name:        name,
			Category:    CategoryFramework,
			Description: def.Description,
		}

		fci.registry.Register(name, cmdFactory, metadata)
	}

	return nil
}

// createFrameworkCommand creates a cobra command for a framework command
func (fci *FrameworkCommandInjector) createFrameworkCommand(name string, def sdk.CommandDefinition) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: def.Description,
		Long:  def.Description,
	}

	// Add alias if provided
	if def.Alias != "" {
		cmd.Aliases = []string{def.Alias}
	}

	// Set the run function
	cmd.RunE = func(cobraCmd *cobra.Command, args []string) error {
		return fci.executeFrameworkCommand(def, args)
	}

	return cmd
}

// executeFrameworkCommand executes a framework command
func (fci *FrameworkCommandInjector) executeFrameworkCommand(def sdk.CommandDefinition, args []string) error {
	// Prepare the command
	cmdStr := def.Cmd

	// Expand any argument placeholders
	if len(args) > 0 {
		cmdStr = fci.expandArguments(cmdStr, args)
	}

	// Set up environment variables
	env := os.Environ()
	for key, value := range def.Env {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	// Parse command and arguments
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	// Create shell command
	cmd := &shell.Command{
		Name:        parts[0],
		Args:        parts[1:],
		Environment: env,
		Mode:        shell.ModeCapture,
		InheritEnv:  true,
	}

	// Execute the command
	result, err := fci.shell.Execute(cmd)
	if err != nil {
		return fmt.Errorf("failed to execute framework command: %w", err)
	}

	// Print output
	if len(result.Stdout) > 0 {
		fmt.Print(string(result.Stdout))
	}
	if len(result.Stderr) > 0 {
		fmt.Fprint(os.Stderr, string(result.Stderr))
	}

	// Return error if command failed
	if result.ExitCode != 0 {
		return fmt.Errorf("command exited with code %d", result.ExitCode)
	}

	return nil
}

// expandArguments expands argument placeholders in command string
func (fci *FrameworkCommandInjector) expandArguments(cmd string, args []string) string {
	// Replace positional arguments
	for i, arg := range args {
		placeholder := fmt.Sprintf("$%d", i+1)
		cmd = strings.ReplaceAll(cmd, placeholder, arg)
	}

	// Replace $@ with all arguments
	if strings.Contains(cmd, "$@") {
		cmd = strings.ReplaceAll(cmd, "$@", strings.Join(args, " "))
	}

	// Replace $* with all arguments
	if strings.Contains(cmd, "$*") {
		cmd = strings.ReplaceAll(cmd, "$*", strings.Join(args, " "))
	}

	return cmd
}

// GetFrameworkInfo returns information about detected frameworks
func (fci *FrameworkCommandInjector) GetFrameworkInfo(projectPath string) ([]detection.FrameworkResult, error) {
	return fci.detector.DetectFrameworks(projectPath)
}

// shouldOverride determines if a command should override an existing one
func shouldOverride(newDef, existingDef sdk.CommandDefinition, confidence int) bool {
	// For now, don't override existing commands
	// In future, could use confidence score or priority
	return false
}

// ExecuteFrameworkCommand is a helper to execute a framework command directly
func ExecuteFrameworkCommand(cmd string, args []string) error {
	// Parse the command
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	// Create command
	command := exec.Command(parts[0], parts[1:]...)

	// Add arguments
	command.Args = append(command.Args, args...)

	// Set up stdio
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	// Run the command
	return command.Run()
}

