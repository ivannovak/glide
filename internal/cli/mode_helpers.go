package cli

import (
	"strings"

	"github.com/ivannovak/glide/v3/internal/config"
	"github.com/ivannovak/glide/v3/internal/context"
	glideErrors "github.com/ivannovak/glide/v3/pkg/errors"
	"github.com/ivannovak/glide/v3/pkg/output"
)

// ShowModeError displays an error message when a command is used in wrong mode
func ShowModeError(currentMode context.DevelopmentMode, requiredMode context.DevelopmentMode, commandName string) error {
	output.Error("❌ Command '%s' is only available in %s mode", commandName, string(requiredMode))
	output.Println()
	output.Printf("Current mode: %s\n", string(currentMode))
	output.Println()

	switch requiredMode {
	case context.ModeMultiWorktree:
		output.Println("To enable multi-worktree mode, run:")
		output.Info("  glidesetup")
		output.Println()
		output.Println("Or manually set mode in ~/.glide.yml:")
		output.Println("  projects:")
		output.Println("    myproject:")
		output.Println("      mode: multi-worktree")
		output.Println()
		output.Println("Multi-worktree mode enables:")
		output.Println("  - Parallel development on multiple branches")
		output.Println("  - Global commands (glideg status, glideg down-all)")
		output.Println("  - Worktree management")
		output.Println("  - Cross-worktree operations")

	case context.ModeSingleRepo:
		output.Println("To use single-repo mode, run:")
		output.Info("  glidesetup")
		output.Println()
		output.Println("Or manually set mode in ~/.glide.yml:")
		output.Println("  projects:")
		output.Println("    myproject:")
		output.Println("      mode: single-repo")
		output.Println()
		output.Println("Single-repo mode provides:")
		output.Println("  - Simple single-branch development")
		output.Println("  - Direct repository access")
		output.Println("  - Simplified workflow")
	}

	return glideErrors.NewModeError(string(currentMode), string(requiredMode), commandName,
		glideErrors.WithSuggestions(
			"Run 'glidesetup' to change development mode",
			"Check the available commands for your current mode",
		),
	)
}

// ValidateMultiWorktreeMode checks if we're in multi-worktree mode
func ValidateMultiWorktreeMode(ctx *context.ProjectContext, commandName string) error {
	if ctx.DevelopmentMode != context.ModeMultiWorktree {
		return ShowModeError(ctx.DevelopmentMode, context.ModeMultiWorktree, commandName)
	}
	return nil
}

// ValidateSingleRepoMode checks if we're in single-repo mode
func ValidateSingleRepoMode(ctx *context.ProjectContext, commandName string) error {
	if ctx.DevelopmentMode != context.ModeSingleRepo {
		return ShowModeError(ctx.DevelopmentMode, context.ModeSingleRepo, commandName)
	}
	return nil
}

// ShowAvailableCommands shows commands available for current mode
func ShowAvailableCommands(mode context.DevelopmentMode) {
	output.Println()
	output.Info("Available commands in %s mode:", string(mode))

	switch mode {
	case context.ModeMultiWorktree:
		output.Println("\n🌐 Global Commands (glideg/global):")
		output.Println("  status         Show Docker status for all worktrees")
		output.Println("  down-all       Stop all containers across worktrees")
		output.Println("  worktree       Create and manage worktrees")
		output.Println("  list-worktrees List all active worktrees")
		output.Println("  clean          Clean up orphaned resources")

		output.Println("\n📁 Local Commands:")
		output.Println("  up, down       Container lifecycle")
		output.Println("  shell, mysql   Interactive access")
		output.Println("  status, logs   Monitoring")
		output.Println("  test, lint     Development tools")

	case context.ModeSingleRepo:
		output.Println("\n📁 Available Commands:")
		output.Println("  up, down       Container lifecycle")
		output.Println("  shell, mysql   Interactive access")
		output.Println("  status, logs   Monitoring")
		output.Println("  test, lint     Development tools")
		output.Println("  docker         Pass-through to docker-compose")
		output.Println("  composer       Package management")
		output.Println("  artisan        Laravel commands")
	}

	output.Println("\n⚙️ Always Available:")
	output.Println("  setup          Configure development mode")
	output.Println("  config         Manage configuration")
	output.Println("  ecr-login      AWS ECR authentication")
	output.Println("  db-tunnel      Database tunnel")
	output.Println("  ssl-certs      SSL certificate generation")
}

// ShowCommandSuggestion shows a helpful suggestion for a mistyped command
func ShowCommandSuggestion(attemptedCommand string, suggestions []string, ctx *context.ProjectContext) error {
	output.Error("❌ Unknown command: '%s'", attemptedCommand)
	output.Raw("\n")

	if len(suggestions) == 1 {
		output.Info("💡 Did you mean: %s", suggestions[0])
		output.Printf("  glide%s\n", suggestions[0])
	} else {
		output.Info("💡 Did you mean one of these?")
		for _, suggestion := range suggestions {
			output.Printf("  glide%s\n", suggestion)
		}
	}

	output.Raw("\n")
	output.Raw("For all available commands: glide help\n")

	return glideErrors.NewConfigError("unknown command",
		glideErrors.WithSuggestions(append(suggestions, "help")...),
	)
}

// ShowUnknownCommandError shows context-aware error for unknown commands
func ShowUnknownCommandError(command string, ctx *context.ProjectContext, cfg *config.Config) error {
	output.Error("❌ Unknown command: '%s'", command)
	output.Raw("\n")

	// Provide context-aware suggestions
	switch ctx.DevelopmentMode {
	case context.ModeMultiWorktree:
		output.Info("💡 Available commands in multi-worktree mode:")
		if ctx.Location == context.LocationRoot {
			output.Raw("  glide global status       # Check all worktrees\n")
			output.Raw("  glide global list         # List worktrees\n")
			output.Raw("  glide global worktree     # Create new worktree\n")
		} else {
			output.Raw("  glide up                  # Start containers\n")
			output.Raw("  glide test                # Run tests\n")
			output.Raw("  glide shell               # Access container\n")
		}

	case context.ModeSingleRepo:
		output.Info("💡 Available commands in single-repo mode:")
		output.Raw("  glide up                  # Start containers\n")
		output.Raw("  glide test                # Run tests\n")
		output.Raw("  glide docker              # Docker commands\n")
		output.Raw("  glide composer            # Package management\n")

	default:
		output.Info("💡 To get started:")
		output.Raw("  glide setup               # Configure project\n")
		output.Raw("  glide help getting-started # Complete guide\n")
	}

	output.Raw("\n")
	output.Raw("📚 Get help:\n")
	output.Raw("  glide help                # Context-aware help\n")
	output.Raw("  glide help workflows      # Common patterns\n")

	// Check for common typos
	commonCommands := []string{"up", "down", "test", "shell", "mysql", "status", "logs"}
	for _, cmd := range commonCommands {
		if strings.Contains(cmd, command) || strings.Contains(command, cmd) {
			output.Raw("\n")
			output.Info("💡 Did you mean: glide%s", cmd)
			break
		}
	}

	return glideErrors.NewConfigError("unknown command: "+command,
		glideErrors.WithSuggestions("help", "setup"),
	)
}

// ShowContextAwareHelp displays the enhanced help system
func ShowContextAwareHelp(ctx *context.ProjectContext, cfg *config.Config) error {
	helpCmd := NewHelpCommand(ctx, cfg)
	return helpCmd.RunE(helpCmd, []string{})
}
