package cli

import (
	"github.com/ivannovak/glide/internal/config"
	"github.com/ivannovak/glide/internal/context"
	"github.com/ivannovak/glide/pkg/output"
	"github.com/spf13/cobra"
)

// HelpCommand handles the enhanced help system
type HelpCommand struct {
	ctx *context.ProjectContext
	cfg *config.Config
}

// NewHelpCommand creates a new enhanced help command
func NewHelpCommand(ctx *context.ProjectContext, cfg *config.Config) *cobra.Command {
	hc := &HelpCommand{
		ctx: ctx,
		cfg: cfg,
	}

	cmd := &cobra.Command{
		Use:   "help [command | topic]",
		Short: "Context-aware help and guidance",
		Long: `Get intelligent, context-aware help for Glide CLI commands and workflows.

The help system adapts to your current development mode and project context,
providing relevant guidance, examples, and troubleshooting information.

Available help topics:
  getting-started    Complete guide for new users
  workflows          Common development workflows  
  modes              Understanding single-repo vs multi-worktree
  troubleshooting    Solutions for common issues

Examples:
  glid help                    # Smart help for current context
  glid help getting-started    # New user onboarding guide
  glid help workflows          # Common workflow examples
  glid help test               # Detailed help for test command
  glid help modes              # Mode differences explained`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return hc.showContextAwareHelp()
			}

			topic := args[0]
			switch topic {
			case "getting-started", "start", "quickstart":
				return hc.showGettingStarted()
			case "workflows", "workflow", "flow":
				return hc.showWorkflows()
			case "modes", "mode":
				return hc.showModes()
			case "troubleshooting", "troubleshoot", "issues":
				return hc.showTroubleshooting()
			default:
				// Check if it's a specific command help request
				return hc.showCommandHelp(topic)
			}
		},
	}

	return cmd
}

// showContextAwareHelp displays help adapted to current context
func (hc *HelpCommand) showContextAwareHelp() error {
	// Detect context and show appropriate help
	switch hc.ctx.DevelopmentMode {
	case context.ModeMultiWorktree:
		return hc.showMultiWorktreeHelp()
	case context.ModeSingleRepo:
		return hc.showSingleRepoHelp()
	default:
		return hc.showNoProjectHelp()
	}
}

// showMultiWorktreeHelp shows help for multi-worktree mode
func (hc *HelpCommand) showMultiWorktreeHelp() error {
	output.Success("🏠 You're in a multi-worktree project")
	output.Raw("\n")
	
	// Show location-specific guidance
	switch hc.ctx.Location {
	case context.LocationRoot:
		output.Info("📍 Current location: Project root (management directory)")
		output.Raw("\n")
		output.Raw("Quick Start - Global Operations:\n")
		output.Raw("  glid global status       # Check all worktree statuses\n")
		output.Raw("  glid global list         # List active worktrees\n")
		output.Raw("  glid global worktree     # Create new feature branch\n")
		output.Raw("  glid global down         # Stop all containers\n")
		
	case context.LocationMainRepo:
		output.Info("📍 Current location: Main repository (vcs/)")
		output.Raw("\n")
		output.Raw("Quick Start - Development Commands:\n")
		output.Raw("  glid up                  # Start containers for main repo\n")
		output.Raw("  glid test                # Run tests\n")
		output.Raw("  glid shell               # Access PHP container\n")
		output.Raw("  glid status              # Check container status\n")
		
	case context.LocationWorktree:
		worktreeName := hc.ctx.WorktreeName
		if worktreeName == "" {
			worktreeName = "current worktree"
		}
		output.Info("📍 Current location: Worktree (%s)", worktreeName)
		output.Raw("\n")
		output.Raw("Quick Start - Feature Development:\n")
		output.Raw("  glid up                  # Start containers for this worktree\n")
		output.Raw("  glid test                # Run tests for your changes\n")
		output.Raw("  glid shell               # Access development environment\n")
		output.Raw("  glid logs                # Monitor container logs\n")
	}
	
	output.Raw("\n")
	output.Raw("💡 Common Workflows:\n")
	output.Raw("  • New feature: glid global worktree feature/name\n")
	output.Raw("  • Daily standup: glid global status\n")
	output.Raw("  • End of day: glid global down\n")
	output.Raw("  • Troubleshooting: glid help troubleshooting\n")
	
	output.Raw("\n")
	output.Raw("📚 Learn More:\n")
	output.Raw("  glid help workflows      # Detailed workflow examples\n")
	output.Raw("  glid help modes          # Multi-worktree vs single-repo\n")
	
	return nil
}

// showSingleRepoHelp shows help for single-repo mode
func (hc *HelpCommand) showSingleRepoHelp() error {
	output.Success("📁 You're in a single-repository project")
	output.Raw("\n")
	
	output.Raw("Quick Start - Development Commands:\n")
	output.Raw("  glid up                  # Start Docker containers\n")
	output.Raw("  glid test                # Run your test suite\n")
	output.Raw("  glid shell               # Access PHP container shell\n")
	output.Raw("  glid mysql               # Access database shell\n")
	output.Raw("  glid logs                # View container logs\n")
	output.Raw("  glid down                # Stop containers\n")
	
	output.Raw("\n")
	output.Raw("💡 Common Workflows:\n")
	output.Raw("  • Start work: glid up && glid test\n")
	output.Raw("  • Run tests: glid test -- --filter MyTest\n")
	output.Raw("  • Debug: glid shell → investigate\n")
	output.Raw("  • Database: glid mysql → run queries\n")
	output.Raw("  • Clean up: glid down\n")
	
	output.Raw("\n")
	output.Raw("📚 Learn More:\n")
	output.Raw("  glid help workflows      # Detailed workflow examples\n")
	output.Raw("  glid help modes          # Consider multi-worktree for larger teams\n")
	
	return nil
}

// showNoProjectHelp shows help when not in a project
func (hc *HelpCommand) showNoProjectHelp() error {
	output.Warning("⚠️  You're not in a Glide project directory")
	output.Raw("\n")
	
	output.Raw("🚀 Getting Started:\n")
	output.Raw("  glid setup               # Interactive project setup\n")
	output.Raw("  glid help getting-started # Complete setup guide\n")
	
	output.Raw("\n")
	output.Raw("Or navigate to an existing project:\n")
	output.Raw("  cd /path/to/your/project\n")
	output.Raw("  glid help                # Context-aware help\n")
	
	output.Raw("\n")
	output.Raw("⚙️  Available Commands (without project):\n")
	output.Raw("  glid setup               # Configure a new project\n")
	output.Raw("  glid config              # Manage global configuration\n")
	output.Raw("  glid version             # Show version information\n")
	
	return nil
}

// showGettingStarted shows the complete getting started guide
func (hc *HelpCommand) showGettingStarted() error {
	output.Success("🚀 Glide CLI - Getting Started Guide")
	output.Raw("\n")
	
	output.Info("Step 1: Project Setup")
	output.Raw("Run this command in your project directory:\n")
	output.Raw("  glid setup\n")
	output.Raw("\n")
	output.Raw("Choose your development mode:\n")
	output.Raw("  • Single-repo: Simple, one-branch-at-a-time development\n")
	output.Raw("  • Multi-worktree: Advanced, parallel branch development\n")
	
	output.Raw("\n")
	output.Info("Step 2: Start Development")
	output.Raw("Basic workflow commands:\n")
	output.Raw("  glid up                  # Start your development environment\n")
	output.Raw("  glid test                # Run your test suite\n")
	output.Raw("  glid shell               # Access your application container\n")
	
	output.Raw("\n")
	output.Info("Step 3: Learn Your Mode")
	
	if hc.ctx.DevelopmentMode == context.ModeMultiWorktree {
		output.Raw("You're in multi-worktree mode! Try:\n")
		output.Raw("  glid global status       # See all your worktrees\n")
		output.Raw("  glid global worktree feature/awesome-feature\n")
	} else {
		output.Raw("Single-repo mode commands:\n")
		output.Raw("  glid docker ps           # See running containers\n")
		output.Raw("  glid composer install    # Install dependencies\n")
		output.Raw("  glid artisan migrate     # Run migrations\n")
	}
	
	output.Raw("\n")
	output.Info("Step 4: Get Help When Needed")
	output.Raw("  glid help workflows      # Common development patterns\n")
	output.Raw("  glid help troubleshooting # Fix common issues\n")
	output.Raw("  glid [command] --help    # Detailed command help\n")
	
	output.Raw("\n")
	output.Success("You're ready to go! Run 'glid help workflows' for common patterns.")
	
	return nil
}

// showWorkflows shows common development workflows
func (hc *HelpCommand) showWorkflows() error {
	output.Success("🔄 Common Development Workflows")
	output.Raw("\n")
	
	if hc.ctx.DevelopmentMode == context.ModeMultiWorktree {
		output.Info("Multi-Worktree Workflows")
		output.Raw("\n")
		
		output.Raw("🌟 Starting a New Feature:\n")
		output.Raw("  glid global worktree feature/user-dashboard\n")
		output.Raw("  cd worktrees/feature-user-dashboard\n")
		output.Raw("  glid up\n")
		output.Raw("  glid test\n")
		
		output.Raw("\n")
		output.Raw("📊 Daily Status Check:\n")
		output.Raw("  glid global status       # All worktree statuses\n")
		output.Raw("  glid global list         # Active worktrees\n")
		
		output.Raw("\n")
		output.Raw("🧹 End of Day Cleanup:\n")
		output.Raw("  glid global down         # Stop all containers\n")
		output.Raw("  glid global clean        # Clean orphaned resources\n")
		
	} else {
		output.Info("Single-Repository Workflows")
		output.Raw("\n")
		
		output.Raw("🌟 Daily Development:\n")
		output.Raw("  glid up                  # Start environment\n")
		output.Raw("  glid test                # Verify everything works\n")
		output.Raw("  # ... make your changes ...\n")
		output.Raw("  glid test -- --filter MyTest  # Test your changes\n")
		output.Raw("  glid down                # Clean shutdown\n")
		
		output.Raw("\n")
		output.Raw("🐛 Debugging Issues:\n")
		output.Raw("  glid logs                # Check container logs\n")
		output.Raw("  glid shell               # Interactive debugging\n")
		output.Raw("  glid mysql               # Database inspection\n")
	}
	
	output.Raw("\n")
	output.Info("Universal Workflows (Any Mode)")
	output.Raw("\n")
	
	output.Raw("🧪 Testing Workflows:\n")
	output.Raw("  glid test                       # All tests\n")
	output.Raw("  glid test -- --parallel         # Parallel execution\n")
	output.Raw("  glid test -- --filter UserTest  # Specific tests\n")
	output.Raw("  glid test -- --coverage         # With coverage\n")
	
	output.Raw("\n")
	output.Raw("🔧 Development Tools:\n")
	output.Raw("  glid lint                # Fix code style\n")
	output.Raw("  glid composer install    # Install dependencies\n")
	output.Raw("  glid artisan migrate     # Database migrations\n")
	
	return nil
}

// showModes explains different development modes
func (hc *HelpCommand) showModes() error {
	output.Success("🎯 Development Modes Explained")
	output.Raw("\n")
	
	output.Info("Single-Repository Mode")
	output.Raw("Perfect for:\n")
	output.Raw("  • Solo development or small teams\n")
	output.Raw("  • Simple projects with linear development\n")
	output.Raw("  • Learning or prototyping\n")
	output.Raw("\n")
	output.Raw("How it works:\n")
	output.Raw("  • Work directly in one repository\n")
	output.Raw("  • Switch branches to change features\n")
	output.Raw("  • All commands operate on current branch\n")
	
	output.Raw("\n")
	output.Info("Multi-Worktree Mode")
	output.Raw("Perfect for:\n")
	output.Raw("  • Teams with parallel development\n")
	output.Raw("  • Multiple features in progress\n")
	output.Raw("  • Code review workflows\n")
	output.Raw("  • Production support + feature development\n")
	output.Raw("\n")
	output.Raw("How it works:\n")
	output.Raw("  • Each feature gets its own directory\n")
	output.Raw("  • Work on multiple branches simultaneously\n")
	output.Raw("  • Global commands manage all worktrees\n")
	output.Raw("  • Isolated Docker environments per feature\n")
	
	output.Raw("\n")
	output.Info("Current Mode: %s", string(hc.ctx.DevelopmentMode))
	
	if hc.ctx.DevelopmentMode == context.ModeMultiWorktree {
		output.Raw("You're using the advanced multi-worktree setup!\n")
		output.Raw("Try: glid global list\n")
	} else if hc.ctx.DevelopmentMode == context.ModeSingleRepo {
		output.Raw("You're using single-repository mode.\n")
		output.Raw("To upgrade: glid setup\n")
	} else {
		output.Raw("No mode detected. Run: glid setup\n")
	}
	
	return nil
}

// showTroubleshooting shows common issue solutions  
func (hc *HelpCommand) showTroubleshooting() error {
	output.Success("🔧 Troubleshooting Guide")
	output.Raw("\n")
	
	output.Info("Common Issues & Solutions")
	output.Raw("\n")
	
	output.Raw("❌ \"unknown command 'global'\"\n")
	output.Raw("  Problem: Trying to use global commands outside multi-worktree mode\n")
	output.Raw("  Solution: glid setup → Choose multi-worktree mode\n")
	
	output.Raw("\n")
	output.Raw("❌ \"Docker not running\"\n")
	output.Raw("  Problem: Docker daemon is not started\n")
	output.Raw("  Solution: Start Docker Desktop application\n")
	
	output.Raw("\n")
	output.Raw("❌ \"Container not found\"\n")
	output.Raw("  Problem: Containers haven't been started\n")
	output.Raw("  Solution: glid up\n")
	
	output.Raw("\n")
	output.Raw("❌ \"Permission denied\"\n")
	output.Raw("  Problem: File permission issues\n")
	output.Raw("  Solution: glid mysql-fix-permissions\n")
	
	output.Raw("\n")
	output.Raw("❌ \"Tests failing\"\n")
	output.Raw("  Problem: Environment or database issues\n")
	output.Raw("  Solution: glid down && glid up && glid test\n")
	
	output.Raw("\n")
	output.Info("Diagnostic Commands")
	output.Raw("  glid context             # Show detected project context\n")
	output.Raw("  glid config              # Show current configuration\n")
	output.Raw("  glid status              # Check container health\n")
	output.Raw("  glid logs                # View container logs\n")
	
	output.Raw("\n")
	output.Success("Still stuck? Check individual command help: glid [command] --help")
	
	return nil
}

// showCommandHelp shows help for a specific command (fallback to cobra help)
func (hc *HelpCommand) showCommandHelp(commandName string) error {
	// This would ideally integrate with cobra's help system
	// For now, suggest the standard help approach
	output.Warning("For detailed help on '%s', use:", commandName)
	output.Info("  glid %s --help", commandName)
	output.Raw("\n")
	output.Raw("Or try these help topics:\n")
	output.Raw("  glid help workflows      # Common workflow examples\n")
	output.Raw("  glid help getting-started # Complete setup guide\n")
	
	return nil
}