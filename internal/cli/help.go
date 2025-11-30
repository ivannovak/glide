package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/ivannovak/glide/v3/internal/config"
	"github.com/ivannovak/glide/v3/internal/context"
	"github.com/ivannovak/glide/v3/pkg/branding"
	"github.com/ivannovak/glide/v3/pkg/output"
	"github.com/ivannovak/glide/v3/pkg/plugin"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// HelpCommand handles the enhanced help system
type HelpCommand struct {
	ProjectContext *context.ProjectContext
	Config         *config.Config
}

// NewHelpCommand creates a new enhanced help command
func NewHelpCommand(ctx *context.ProjectContext, cfg *config.Config) *cobra.Command {
	hc := &HelpCommand{
		ProjectContext: ctx,
		Config:         cfg,
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
  glide help                    # Smart help for current context
  glide help getting-started    # New user onboarding guide
  glide help workflows          # Common workflow examples
  glide help test               # Detailed help for test command
  glide help modes              # Mode differences explained`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				// Show the default help
				return hc.ShowHelp(cmd.Root())
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

// showGettingStarted shows the complete getting started guide
func (hc *HelpCommand) showGettingStarted() error {
	output.Success("🚀 Glide CLI - Getting Started Guide")
	output.Raw("\n")

	output.Info("Step 1: Project Setup")
	output.Raw("Run this command in your project directory:\n")
	output.Raw("  glide setup\n")
	output.Raw("\n")
	output.Raw("Choose your development mode:\n")
	output.Raw("  • Single-repo: Simple, one-branch-at-a-time development\n")
	output.Raw("  • Multi-worktree: Advanced, parallel branch development\n")

	output.Raw("\n")
	output.Info("Step 2: Start Development")
	output.Raw("Basic workflow commands:\n")
	output.Raw("  glide up                  # Start your development environment\n")
	output.Raw("  glide test                # Run your test suite\n")
	output.Raw("  glide shell               # Access your application container\n")

	output.Raw("\n")
	output.Info("Step 3: Learn Your Mode")

	if hc.ProjectContext.DevelopmentMode == context.ModeMultiWorktree {
		output.Raw("You're in multi-worktree mode! Try:\n")
		output.Raw("  glide global status       # See all your worktrees\n")
		output.Raw("  glide global worktree feature/awesome-feature\n")
	} else {
		output.Raw("Single-repo mode commands:\n")
		output.Raw("  glide docker ps           # See running containers\n")
		output.Raw("  glide composer install    # Install dependencies\n")
		output.Raw("  glide artisan migrate     # Run migrations\n")
	}

	output.Raw("\n")
	output.Info("Step 4: Get Help When Needed")
	output.Raw("  glide help workflows      # Common development patterns\n")
	output.Raw("  glide help troubleshooting # Fix common issues\n")
	output.Raw("  glide [command] --help    # Detailed command help\n")

	output.Raw("\n")
	output.Success("You're ready to go! Run 'glide help workflows' for common patterns.")

	return nil
}

// showWorkflows shows common development workflows
func (hc *HelpCommand) showWorkflows() error {
	output.Success("🔄 Common Development Workflows")
	output.Raw("\n")

	if hc.ProjectContext.DevelopmentMode == context.ModeMultiWorktree {
		output.Info("Multi-Worktree Workflows")
		output.Raw("\n")

		output.Raw("🌟 Starting a New Feature:\n")
		output.Raw("  glide global worktree feature/user-dashboard\n")
		output.Raw("  cd worktrees/feature-user-dashboard\n")
		output.Raw("  glide up\n")
		output.Raw("  glide test\n")

		output.Raw("\n")
		output.Raw("📊 Daily Status Check:\n")
		output.Raw("  glide global status       # All worktree statuses\n")
		output.Raw("  glide global list         # Active worktrees\n")

		output.Raw("\n")
		output.Raw("🧹 End of Day Cleanup:\n")
		output.Raw("  glide global down         # Stop all containers\n")
		output.Raw("  glide global clean        # Clean orphaned resources\n")

	} else {
		output.Info("Single-Repository Workflows")
		output.Raw("\n")

		output.Raw("🌟 Daily Development:\n")
		output.Raw("  glide up                  # Start environment\n")
		output.Raw("  glide test                # Verify everything works\n")
		output.Raw("  # ... make your changes ...\n")
		output.Raw("  glide test -- --filter MyTest  # Test your changes\n")
		output.Raw("  glide down                # Clean shutdown\n")

		output.Raw("\n")
		output.Raw("🐛 Debugging Issues:\n")
		output.Raw("  glide logs                # Check container logs\n")
		output.Raw("  glide shell               # Interactive debugging\n")
		output.Raw("  glide mysql               # Database inspection\n")
	}

	output.Raw("\n")
	output.Info("Universal Workflows (Any Mode)")
	output.Raw("\n")

	output.Raw("🧪 Testing Workflows:\n")
	output.Raw("  glide test                       # All tests\n")
	output.Raw("  glide test -- --parallel         # Parallel execution\n")
	output.Raw("  glide test -- --filter UserTest  # Specific tests\n")
	output.Raw("  glide test -- --coverage         # With coverage\n")

	output.Raw("\n")
	output.Raw("🔧 Development Tools:\n")
	output.Raw("  glide lint                # Fix code style\n")
	output.Raw("  glide composer install    # Install dependencies\n")
	output.Raw("  glide artisan migrate     # Database migrations\n")

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
	output.Info("Current Mode: %s", string(hc.ProjectContext.DevelopmentMode))

	if hc.ProjectContext.DevelopmentMode == context.ModeMultiWorktree {
		output.Raw("You're using the advanced multi-worktree setup!\n")
		output.Raw("Try: glide global list\n")
	} else if hc.ProjectContext.DevelopmentMode == context.ModeSingleRepo {
		output.Raw("You're using single-repository mode.\n")
		output.Raw("To upgrade: glide setup\n")
	} else {
		output.Raw("No mode detected. Run: glide setup\n")
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
	output.Raw("  Solution: glide setup → Choose multi-worktree mode\n")

	output.Raw("\n")
	output.Raw("❌ \"Docker not running\"\n")
	output.Raw("  Problem: Docker daemon is not started\n")
	output.Raw("  Solution: Start Docker Desktop application\n")

	output.Raw("\n")
	output.Raw("❌ \"Container not found\"\n")
	output.Raw("  Problem: Containers haven't been started\n")
	output.Raw("  Solution: glide up\n")

	output.Raw("\n")
	output.Raw("❌ \"Permission denied\"\n")
	output.Raw("  Problem: File permission issues\n")
	output.Raw("  Solution: glide mysql-fix-permissions\n")

	output.Raw("\n")
	output.Raw("❌ \"Tests failing\"\n")
	output.Raw("  Problem: Environment or database issues\n")
	output.Raw("  Solution: glide down && glide up && glide test\n")

	output.Raw("\n")
	output.Info("Diagnostic Commands")
	output.Raw("  glide context             # Show detected project context\n")
	output.Raw("  glide config              # Show current configuration\n")
	output.Raw("  glide status              # Check container health\n")
	output.Raw("  glide logs                # View container logs\n")

	output.Raw("\n")
	output.Success("Still stuck? Check individual command help: glide [command] --help")

	return nil
}

// showCommandHelp shows help for a specific command (fallback to cobra help)
func (hc *HelpCommand) showCommandHelp(commandName string) error {
	// This would ideally integrate with cobra's help system
	// For now, suggest the standard help approach
	output.Warning("For detailed help on '%s', use:", commandName)
	output.Info("  glide %s --help", commandName)
	output.Raw("\n")
	output.Raw("Or try these help topics:\n")
	output.Raw("  glide help workflows      # Common workflow examples\n")
	output.Raw("  glide help getting-started # Complete setup guide\n")

	return nil
}

// CategoryInfo holds information about a command category
type CategoryInfo struct {
	Name        string
	Description string
	Priority    int // Lower numbers appear first
	Color       *color.Color
}

// Categories defines all command categories with their display properties
var Categories = map[string]CategoryInfo{
	"core": {
		Name:        "Core Commands",
		Description: "Essential development commands",
		Priority:    10,
		Color:       color.New(color.FgYellow, color.Bold),
	},
	"global": {
		Name:        "Global Commands",
		Description: "Multi-worktree management",
		Priority:    20,
		Color:       color.New(color.FgYellow, color.Bold),
	},
	"setup": {
		Name:        "Setup & Configuration",
		Description: "Project setup and configuration",
		Priority:    30,
		Color:       color.New(color.FgYellow, color.Bold),
	},
	// Project-specific categories (40-60) - will be moved to plugins
	"docker": {
		Name:        "Docker Management",
		Description: "Container and service control",
		Priority:    40,
		Color:       color.New(color.FgYellow, color.Bold),
	},
	"testing": {
		Name:        "Testing",
		Description: "Test execution and coverage",
		Priority:    50,
		Color:       color.New(color.FgYellow, color.Bold),
	},
	"developer": {
		Name:        "Development Tools",
		Description: "Code quality and utilities",
		Priority:    60,
		Color:       color.New(color.FgYellow, color.Bold),
	},
	"database": {
		Name:        "Database",
		Description: "Database management and access",
		Priority:    70,
		Color:       color.New(color.FgYellow, color.Bold),
	},
	// Plugin commands get their own section
	"plugin": {
		Name:        "Plugin Commands",
		Description: "Commands from installed plugins",
		Priority:    80,
		Color:       color.New(color.FgYellow, color.Bold),
	},
	// Help is always last
	"help": {
		Name:        "Help & Documentation",
		Description: "Help topics and guides",
		Priority:    90,
		Color:       color.New(color.FgYellow, color.Bold),
	},
}

// CommandEntry represents a command for display
type CommandEntry struct {
	Name        string
	Description string
	Aliases     []string
	Category    string
	IsPlugin    bool
	IsYAML      bool // User-defined YAML command
	PluginName  string
}

// ShowHelp displays the categorized help output
func (hc *HelpCommand) ShowHelp(rootCmd *cobra.Command) error {
	// Load custom categories from plugins
	hc.loadPluginCategories()

	// ASCII Art Header
	asciiHeader := `
   ___ _ _    _
  / __| (_)__| |___
 | (_ | | / _` + "`" + ` / -_)
  \___|_|_\__,_\___|

`
	headerColor := color.New(color.FgBlue, color.Bold)
	headerColor.Print(asciiHeader)

	// Subtitle
	subtitleColor := color.New(color.FgWhite)
	subtitleColor.Printf("    %s\n", rootCmd.Short)

	// Show context-specific information if we have project context
	if hc.ProjectContext != nil {
		hc.showContextInfo()

		// In standalone mode, ensure YAML command categories are visible
		if hc.ProjectContext.DevelopmentMode == context.ModeStandalone {
			// Enable all categories that may be used in .glide.yml
			// This is handled in shouldShowCategory but we set a flag here
		}
	}

	// Usage
	fmt.Println("\nUsage:")
	fmt.Printf("  %s [flags]\n", rootCmd.Use)
	fmt.Printf("  %s [command]\n\n", rootCmd.Use)

	// Collect all commands and organize by category
	commandsByCategory := make(map[string][]CommandEntry)

	// Process built-in commands
	for _, cmd := range rootCmd.Commands() {
		if os.Getenv("GLIDE_HELP_DEBUG") != "" {
			fmt.Fprintf(os.Stderr, "DEBUG: Processing command '%s'\n", cmd.Name())
		}

		if cmd.Hidden {
			if os.Getenv("GLIDE_HELP_DEBUG") != "" {
				fmt.Fprintf(os.Stderr, "DEBUG: Skipping hidden command '%s'\n", cmd.Name())
			}
			continue
		}

		// Check visibility for plugin commands
		if !hc.shouldShowCommand(cmd) {
			if os.Getenv("GLIDE_HELP_DEBUG") != "" {
				fmt.Fprintf(os.Stderr, "DEBUG: Command '%s' failed shouldShowCommand\n", cmd.Name())
			}
			continue
		}

		entry := CommandEntry{
			Name:        cmd.Name(),
			Description: cmd.Short,
			Aliases:     cmd.Aliases,
		}

		// Check if this is a user-defined YAML command
		if cmd.Annotations != nil {
			if _, isYAML := cmd.Annotations["yaml_command"]; isYAML {
				entry.IsYAML = true
			}
		}

		// Get category from command annotations (set by registry)
		category := "core" // default
		if cmd.Annotations != nil {
			if cat, ok := cmd.Annotations["category"]; ok {
				category = cat
			}
			// Debug: Check if this is a plugin command
			if _, isPlugin := cmd.Annotations["plugin"]; isPlugin && category == "core" {
				// Plugin commands without explicit category should go to plugin category
				category = "plugin"
			}
		}
		entry.Category = category

		if os.Getenv("GLIDE_HELP_DEBUG") != "" {
			fmt.Fprintf(os.Stderr, "DEBUG: Command '%s' has category '%s'\n", cmd.Name(), category)
		}

		// Special handling for plugin commands
		if strings.HasPrefix(cmd.Use, "plugin:") || cmd.Annotations != nil && cmd.Annotations["plugin"] != "" {
			entry.IsPlugin = true
			// Plugin category might be overridden by the plugin itself
			if cmd.Annotations != nil && cmd.Annotations["plugin_category"] != "" {
				entry.Category = cmd.Annotations["plugin_category"]
			}
		}

		commandsByCategory[category] = append(commandsByCategory[category], entry)
	}

	// Process plugin commands
	pluginCommands := hc.getPluginCommands()
	commandsByCategory["plugin"] = append(commandsByCategory["plugin"], pluginCommands...)

	// Sort categories by priority
	var sortedCategories []string
	for cat := range commandsByCategory {
		sortedCategories = append(sortedCategories, cat)
	}
	sort.Slice(sortedCategories, func(i, j int) bool {
		catI, okI := Categories[sortedCategories[i]]
		catJ, okJ := Categories[sortedCategories[j]]
		if !okI {
			return false
		}
		if !okJ {
			return true
		}
		return catI.Priority < catJ.Priority
	})

	// Display commands by category
	for _, category := range sortedCategories {
		commands := commandsByCategory[category]
		if len(commands) == 0 {
			continue
		}

		// Check if any commands in this category are user-defined YAML commands
		hasYAMLCommands := false
		for _, cmd := range commands {
			if cmd.IsYAML {
				hasYAMLCommands = true
				break
			}
		}

		// Context-aware filtering - but ALWAYS show categories with user-defined commands
		if !hasYAMLCommands && !hc.shouldShowCategory(category) {
			if os.Getenv("GLIDE_HELP_DEBUG") != "" {
				fmt.Fprintf(os.Stderr, "DEBUG HELP: Skipping category %s (shouldShowCategory=false, no YAML commands)\n", category)
			}
			continue
		}

		catInfo, ok := Categories[category]
		if !ok {
			caser := cases.Title(language.English)
			catInfo = CategoryInfo{
				Name:  caser.String(category),
				Color: color.New(color.FgWhite),
			}
		}

		// Category header
		fmt.Println()
		if catInfo.Color != nil {
			catInfo.Color.Printf("%s", catInfo.Name)
		} else {
			fmt.Printf("%s", catInfo.Name)
		}

		if catInfo.Description != "" {
			color.New(color.Faint).Printf(" - %s", catInfo.Description)
		}
		fmt.Println()

		// Sort commands alphabetically
		sort.Slice(commands, func(i, j int) bool {
			return commands[i].Name < commands[j].Name
		})

		// Find the longest command name and alias for alignment
		maxLen := 0
		maxAliasLen := 0
		for _, cmd := range commands {
			nameLen := len(cmd.Name)
			if nameLen > maxLen {
				maxLen = nameLen
			}

			if len(cmd.Aliases) > 0 {
				aliasLen := len(strings.Join(cmd.Aliases, ", "))
				if aliasLen > maxAliasLen {
					maxAliasLen = aliasLen
				}
			}
		}

		// Add some padding if there are no aliases in this category
		if maxAliasLen == 0 {
			maxAliasLen = 1
		}

		// Display commands
		commandColor := color.New(color.FgGreen)
		aliasColor := color.New(color.Faint)
		faintGray := color.New(color.Faint)

		for _, cmd := range commands {
			// Print command name in green
			fmt.Print("  ")
			commandColor.Printf("%-*s", maxLen, cmd.Name)

			// Print aliases in faint gray (if any)
			if len(cmd.Aliases) > 0 {
				aliasStr := strings.Join(cmd.Aliases, ", ")
				fmt.Print("  ")
				aliasColor.Printf("%-*s", maxAliasLen, aliasStr)
			} else {
				fmt.Print("  ")
				fmt.Printf("%-*s", maxAliasLen, "") // Empty space for alignment
			}

			// Print description
			fmt.Printf("  %s\n", cmd.Description)

			// Show plugin source if applicable
			if cmd.IsPlugin && cmd.PluginName != "" {
				pluginColor := color.New(color.Faint, color.Italic)
				fmt.Printf("  %-*s  %-*s  %s\n", maxLen, "", maxAliasLen, "", pluginColor.Sprintf("from %s plugin", cmd.PluginName))
			}

			// Show plugin subcommands if this is a plugin
			if category == "plugin" {
				subcommands := hc.getPluginSubcommands(rootCmd, cmd.Name)
				if len(subcommands) > 0 {
					for i, subcmd := range subcommands {
						fmt.Print("    ")
						// Use └─ for last item, ├─ for others (in muted gray)
						if i == len(subcommands)-1 {
							faintGray.Print("└─ ")
						} else {
							faintGray.Print("├─ ")
						}
						// Print subcommand name in normal green
						commandColor.Printf("%-*s", maxLen-3, subcmd.Name)

						if len(subcmd.Aliases) > 0 {
							fmt.Print("  ")
							aliasColor.Printf("%-*s", maxAliasLen, strings.Join(subcmd.Aliases, ", "))
						} else {
							fmt.Print("  ")
							fmt.Printf("%-*s", maxAliasLen, "")
						}
						fmt.Printf("  %s\n", subcmd.Description)
					}
				}
			}
		}
	}

	// Footer with help topics
	fmt.Println()
	color.New(color.FgWhite, color.Bold).Println("Getting Help:")
	fmt.Println("  glide help [command]         Show detailed help for a command")
	fmt.Println("  glide [command] --help       Same as above")
	fmt.Println("  glide help getting-started   New user guide")
	fmt.Println("  glide help workflows         Common development patterns")

	// Context-aware tips
	if hc.ProjectContext != nil {
		fmt.Println()
		hc.showContextTips()
	}

	// Version and more info
	fmt.Println()
	faintColor := color.New(color.Faint)
	faintColor.Printf("Use \"glide [command] --help\" for more information about a command.\n")

	return nil
}

// getPluginCommands retrieves commands from loaded plugins
func (hc *HelpCommand) getPluginCommands() []CommandEntry {
	var commands []CommandEntry

	// Get plugin info from the plugin manager
	plugins := plugin.List()
	for _, p := range plugins {
		meta := p.Metadata()
		for _, cmd := range meta.Commands {
			entry := CommandEntry{
				Name:        cmd.Name,
				Description: cmd.Description,
				Category:    "plugin",
				IsPlugin:    true,
				PluginName:  meta.Name,
			}

			// Override category if plugin specifies one
			if cmd.Category != "" {
				// Map plugin categories to our categories
				switch cmd.Category {
				case "docker":
					entry.Category = "docker"
				case "database":
					entry.Category = "database"
				case "setup":
					entry.Category = "setup"
				default:
					// Keep as plugin category
				}
			}

			commands = append(commands, entry)
		}
	}

	return commands
}

// SubcommandEntry represents a subcommand for display
type SubcommandEntry struct {
	Name        string
	Description string
	Aliases     []string
}

// getPluginSubcommands retrieves subcommands for a specific plugin
func (hc *HelpCommand) getPluginSubcommands(rootCmd *cobra.Command, pluginName string) []SubcommandEntry {
	var subcommands []SubcommandEntry

	// Find the plugin command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == pluginName {
			// Get all subcommands of this plugin
			for _, subcmd := range cmd.Commands() {
				if !subcmd.Hidden {
					subcommands = append(subcommands, SubcommandEntry{
						Name:        subcmd.Name(),
						Description: subcmd.Short,
						Aliases:     subcmd.Aliases,
					})
				}
			}
			break
		}
	}

	// Sort subcommands alphabetically
	sort.Slice(subcommands, func(i, j int) bool {
		return subcommands[i].Name < subcommands[j].Name
	})

	return subcommands
}

// areCompletionsInstalled checks if shell completions are already installed
func (hc *HelpCommand) areCompletionsInstalled() bool {
	// Check common completion locations for various shells
	homeDir := os.Getenv("HOME")

	// Locations to check for installed completions
	completionPaths := []string{
		// Bash completions
		fmt.Sprintf("/usr/local/etc/bash_completion.d/%s", branding.CommandName),
		fmt.Sprintf("/etc/bash_completion.d/%s", branding.CommandName),
		fmt.Sprintf("%s/.bash_completion.d/%s", homeDir, branding.CommandName),

		// Zsh completions
		fmt.Sprintf("/usr/local/share/zsh/site-functions/_%s", branding.CommandName),
		fmt.Sprintf("/usr/share/zsh/site-functions/_%s", branding.CommandName),
		fmt.Sprintf("%s/.zsh/completions/_%s", homeDir, branding.CommandName),

		// Fish completions
		fmt.Sprintf("%s/.config/fish/completions/%s.fish", homeDir, branding.CommandName),
	}

	// Also check if completion content exists in shell config files
	shellConfigs := []string{
		fmt.Sprintf("%s/.bashrc", homeDir),
		fmt.Sprintf("%s/.bash_profile", homeDir),
		fmt.Sprintf("%s/.bash_completion", homeDir),
		fmt.Sprintf("%s/.zshrc", homeDir),
	}

	// Check if any completion file exists
	for _, path := range completionPaths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	// Check if completion is sourced in shell configs
	for _, configPath := range shellConfigs {
		if data, err := os.ReadFile(configPath); err == nil {
			if strings.Contains(string(data), branding.CommandName) &&
				strings.Contains(string(data), "complete") {
				return true
			}
		}
	}

	return false
}

// shouldShowCategory determines if a category should be shown based on context
func (hc *HelpCommand) shouldShowCategory(category string) bool {
	// No context means show everything except global and development categories
	if hc.ProjectContext == nil {
		// Hide categories that require a project context
		switch category {
		case "global", "project", "docker", "testing", "developer", "database":
			return false
		default:
			return true
		}
	}

	// In standalone mode, hide project-oriented built-in categories
	// Note: Categories with user-defined YAML commands will still be shown
	// because the display logic bypasses this filter for categories with YAML commands
	if hc.ProjectContext.DevelopmentMode == context.ModeStandalone {
		switch category {
		case "project", "docker", "testing", "database":
			// These built-in commands don't make sense in standalone (non-Git) mode
			// But YAML commands in these categories will still be shown
			return false
		case "setup":
			// In standalone mode, only completion might be useful
			// The setup command itself doesn't apply to non-Git directories
			// This is handled at the command level in shouldShowCommand
			return true
		default:
			return true
		}
	}

	switch category {
	case "global", "project":
		// Only show project commands in multi-worktree mode
		return hc.ProjectContext.DevelopmentMode == context.ModeMultiWorktree

	case "docker", "testing", "developer", "database":
		// Don't show development commands when not in a project
		if hc.ProjectContext == nil || hc.ProjectContext.DevelopmentMode == "" {
			return false
		}
		// Show these categories in project modes
		return true

	default:
		// Show all other categories
		return true
	}
}

// shouldShowCommand checks if a command should be shown based on its visibility setting
func (hc *HelpCommand) shouldShowCommand(cmd *cobra.Command) bool {
	// YAML commands (user-defined) are ALWAYS shown, regardless of mode or category
	// Users should have full control over their custom commands
	if cmd.Annotations != nil {
		if _, isYAML := cmd.Annotations["yaml_command"]; isYAML {
			return true
		}
	}

	// Special handling for specific commands
	switch cmd.Name() {
	case "setup":
		// Setup doesn't make sense in standalone mode (no Git to configure)
		if hc.ProjectContext != nil && hc.ProjectContext.DevelopmentMode == context.ModeStandalone {
			return false
		}
	case "completion":
		// Hide if completions are already installed
		if hc.areCompletionsInstalled() {
			return false
		}
	}

	// Commands without visibility annotation are always shown
	if cmd.Annotations == nil {
		return true
	}

	visibility, hasVisibility := cmd.Annotations["visibility"]
	if !hasVisibility {
		return true
	}

	// No project context means we're not in a project
	if hc.ProjectContext == nil {
		// Show commands that are always visible or explicitly allow non-project contexts
		switch visibility {
		case "always", "non-root":
			return true
		default:
			return false
		}
	}

	// Check visibility based on project context
	switch visibility {
	case "always":
		return true

	case "project-only":
		// Show only when in a project (any mode)
		return hc.ProjectContext.DevelopmentMode != ""

	case "worktree-only":
		// Show only when in a worktree (not at multi-worktree root)
		if hc.ProjectContext.DevelopmentMode != context.ModeMultiWorktree {
			return false
		}
		return hc.ProjectContext.Location == context.LocationWorktree

	case "root-only":
		// Show only at multi-worktree root
		if hc.ProjectContext.DevelopmentMode != context.ModeMultiWorktree {
			return false
		}
		return hc.ProjectContext.Location == context.LocationRoot

	case "non-root":
		// Show everywhere except multi-worktree root
		if hc.ProjectContext != nil && hc.ProjectContext.DevelopmentMode == context.ModeMultiWorktree {
			return hc.ProjectContext.Location != context.LocationRoot
		}
		return true

	default:
		// Unknown visibility setting, default to showing
		return true
	}
}

// loadPluginCategories loads custom categories from plugins and adds them to the global Categories map
func (hc *HelpCommand) loadPluginCategories() {
	// Get custom categories from plugins
	customCategories := plugin.GetGlobalPluginCategories()
	for _, cat := range customCategories {
		// Add to the global Categories map
		Categories[cat.Id] = CategoryInfo{
			Name:        cat.Name,
			Description: cat.Description,
			Priority:    int(cat.Priority),
			Color:       color.New(color.FgYellow, color.Bold), // Yellow bold for custom categories
		}
	}
}

// showContextInfo displays context information at the top of help
func (hc *HelpCommand) showContextInfo() {
	contextColor := color.New(color.FgCyan)

	switch hc.ProjectContext.DevelopmentMode {
	case context.ModeMultiWorktree:
		contextColor.Print("📂 Multi-worktree mode")
		switch hc.ProjectContext.Location {
		case context.LocationRoot:
			fmt.Printf(" • Project root")
		case context.LocationMainRepo:
			fmt.Printf(" • Main repository (vcs/)")
		case context.LocationWorktree:
			if hc.ProjectContext.WorktreeName != "" {
				fmt.Printf(" • Worktree: %s", hc.ProjectContext.WorktreeName)
			} else {
				fmt.Printf(" • Worktree")
			}
		}
		fmt.Println()

	case context.ModeSingleRepo:
		contextColor.Println("📁 Single-repo mode")

	case context.ModeStandalone:
		contextColor.Println("📄 Standalone mode")

	default:
		color.New(color.FgYellow).Println("⚠️  No project detected")
	}
}

// showContextTips shows context-aware tips based on the current location
func (hc *HelpCommand) showContextTips() {
	tipColor := color.New(color.FgYellow)

	switch hc.ProjectContext.DevelopmentMode {
	case context.ModeMultiWorktree:
		switch hc.ProjectContext.Location {
		case context.LocationRoot:
			tipColor.Println("💡 Tip: You're in the project root. Use 'glide project' commands to manage worktrees.")
		case context.LocationMainRepo:
			tipColor.Println("💡 Tip: You're in vcs/ (main branch). Create worktrees with 'glide project worktree <branch>'.")
		case context.LocationWorktree:
			tipColor.Println("💡 Tip: You're in a worktree. All commands operate on this feature branch.")
		}
	case context.ModeSingleRepo:
		tipColor.Println("💡 Tip: Single-repo mode active. All commands operate on the current branch.")
	case context.ModeStandalone:
		tipColor.Println("💡 Tip: Standalone mode active. Commands from .glide.yml are available.")
	default:
		tipColor.Println("💡 Tip: Run 'glide setup' to configure your project.")
	}
}
