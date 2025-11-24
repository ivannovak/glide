package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ivannovak/glide/v2/internal/config"
	"github.com/ivannovak/glide/v2/internal/context"
	"github.com/ivannovak/glide/v2/internal/docker"
	"github.com/ivannovak/glide/v2/pkg/branding"
	glideErrors "github.com/ivannovak/glide/v2/pkg/errors"
	"github.com/ivannovak/glide/v2/pkg/output"
	"github.com/spf13/cobra"
)

// CompletionType represents the type of shell completion
type CompletionType string

const (
	CompletionBash CompletionType = "bash"
	CompletionZsh  CompletionType = "zsh"
	CompletionFish CompletionType = "fish"
)

// CompletionManager handles shell completion generation and installation
type CompletionManager struct {
	ctx *context.ProjectContext
	cfg *config.Config
}

// NewCompletionManager creates a new completion manager
func NewCompletionManager(ctx *context.ProjectContext, cfg *config.Config) *CompletionManager {
	return &CompletionManager{
		ctx: ctx,
		cfg: cfg,
	}
}

// NewCompletionCommand creates the completion command for manual installation
func NewCompletionCommand(ctx *context.ProjectContext, cfg *config.Config) *cobra.Command {
	manager := NewCompletionManager(ctx, cfg)

	cmd := &cobra.Command{
		Use:   "completion [shell]",
		Short: "Generate shell completion scripts",
		Long: fmt.Sprintf(`Generate shell completion scripts for bash, zsh, or fish.

To install completions:

Bash:
  %s completion bash > /etc/bash_completion.d/%s
  # or
  %s completion bash > /usr/local/etc/bash_completion.d/%s (macOS with Homebrew)

Zsh:
  %s completion zsh > "${fpath[1]}/_%s"
  # or add to your ~/.zshrc:
  source <(%s completion zsh)

Fish:
  %s completion fish > ~/.config/fish/completions/%s.fish`,
			branding.CommandName, branding.CommandName,
			branding.CommandName, branding.CommandName,
			branding.CommandName, branding.CommandName,
			branding.CommandName,
			branding.CommandName, branding.CommandName),
		ValidArgs:    []string{"bash", "zsh", "fish"},
		Args:         cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return manager.GenerateCompletion(cmd, CompletionType(args[0]))
		},
	}

	return cmd
}

// GenerateCompletion generates completion script for the specified shell
func (cm *CompletionManager) GenerateCompletion(cmd *cobra.Command, shell CompletionType) error {
	switch shell {
	case CompletionBash:
		return cmd.Root().GenBashCompletion(os.Stdout)
	case CompletionZsh:
		return cmd.Root().GenZshCompletion(os.Stdout)
	case CompletionFish:
		return cmd.Root().GenFishCompletion(os.Stdout, true)
	default:
		return glideErrors.NewConfigError(
			fmt.Sprintf("unsupported shell: %s", shell),
			glideErrors.WithSuggestions("Use 'bash', 'zsh', or 'fish'"),
		)
	}
}

// InstallCompletion installs completion scripts automatically during setup
func (cm *CompletionManager) InstallCompletion() error {
	shell := cm.detectShell()
	if shell == "" {
		output.Warning("Could not detect shell, skipping completion installation")
		return nil
	}

	output.Info("Installing %s completion...", shell)

	switch CompletionType(shell) {
	case CompletionBash:
		return cm.installBashCompletion()
	case CompletionZsh:
		return cm.installZshCompletion()
	case CompletionFish:
		return cm.installFishCompletion()
	default:
		output.Warning("Unsupported shell %s, skipping completion installation", shell)
		return nil
	}
}

// detectShell attempts to detect the user's current shell
func (cm *CompletionManager) detectShell() string {
	// Check SHELL environment variable
	if shell := os.Getenv("SHELL"); shell != "" {
		return filepath.Base(shell)
	}

	// Fallback detection methods could be added here
	return ""
}

// installBashCompletion installs bash completion
func (cm *CompletionManager) installBashCompletion() error {
	// Try common bash completion directories
	completionDirs := []string{
		"/usr/local/etc/bash_completion.d",                     // macOS with Homebrew
		"/etc/bash_completion.d",                               // Linux
		filepath.Join(os.Getenv("HOME"), ".bash_completion.d"), // User directory
	}

	for _, dir := range completionDirs {
		if cm.dirExists(dir) || cm.canCreateDir(dir) {
			completionFile := filepath.Join(dir, branding.CommandName)
			return cm.writeCompletionFile(completionFile, CompletionBash)
		}
	}

	// Fallback: Add to ~/.bash_completion
	bashCompletion := filepath.Join(os.Getenv("HOME"), ".bash_completion")
	return cm.appendCompletionToFile(bashCompletion, CompletionBash)
}

// installZshCompletion installs zsh completion
func (cm *CompletionManager) installZshCompletion() error {
	// Get zsh fpath
	fpath := os.Getenv("FPATH")
	if fpath == "" {
		// Default zsh completion directory
		fpath = "/usr/local/share/zsh/site-functions"
	}

	// Use first directory in fpath
	fpathDirs := strings.Split(fpath, ":")
	if len(fpathDirs) > 0 {
		completionFile := filepath.Join(fpathDirs[0], fmt.Sprintf("_%s", branding.CommandName))
		if cm.dirExists(fpathDirs[0]) || cm.canCreateDir(fpathDirs[0]) {
			return cm.writeCompletionFile(completionFile, CompletionZsh)
		}
	}

	// Fallback: User zsh directory
	userZshDir := filepath.Join(os.Getenv("HOME"), ".zsh", "completions")
	if err := os.MkdirAll(userZshDir, 0755); err == nil {
		completionFile := filepath.Join(userZshDir, fmt.Sprintf("_%s", branding.CommandName))
		if err := cm.writeCompletionFile(completionFile, CompletionZsh); err == nil {
			output.Info("Add to your ~/.zshrc: fpath=(~/.zsh/completions $fpath)")
			return nil
		}
	}

	return glideErrors.NewPermissionError(
		"~/.zsh/completions",
		"could not install zsh completion",
		glideErrors.WithSuggestions(
			fmt.Sprintf("Run: %s completion zsh > ~/.zsh/completions/_%s", branding.CommandName, branding.CommandName),
			"Add to ~/.zshrc: fpath=(~/.zsh/completions $fpath)",
		),
	)
}

// installFishCompletion installs fish completion
func (cm *CompletionManager) installFishCompletion() error {
	fishDir := filepath.Join(os.Getenv("HOME"), ".config", "fish", "completions")
	if err := os.MkdirAll(fishDir, 0755); err != nil {
		return glideErrors.NewPermissionError(
			fishDir,
			"could not create fish completions directory",
			glideErrors.WithSuggestions("Run: mkdir -p ~/.config/fish/completions"),
		)
	}

	completionFile := filepath.Join(fishDir, fmt.Sprintf("%s.fish", branding.CommandName))
	return cm.writeCompletionFile(completionFile, CompletionFish)
}

// writeCompletionFile writes completion script to a file
func (cm *CompletionManager) writeCompletionFile(filename string, shell CompletionType) error {
	file, err := os.Create(filename)
	if err != nil {
		return glideErrors.NewPermissionError(
			filename,
			fmt.Sprintf("could not create completion file: %s", filename),
			glideErrors.WithSuggestions(
				fmt.Sprintf("Try: sudo %s completion %s > %s", branding.CommandName, shell, filename),
				fmt.Sprintf("Or install manually with: %s completion %s", branding.CommandName, string(shell)),
			),
		)
	}
	defer file.Close()

	// Create a temporary cobra command that matches the real structure
	rootCmd := cm.createMockRootCommand()

	switch shell {
	case CompletionBash:
		err = rootCmd.GenBashCompletion(file)
	case CompletionZsh:
		err = rootCmd.GenZshCompletion(file)
	case CompletionFish:
		err = rootCmd.GenFishCompletion(file, true)
	}

	if err != nil {
		os.Remove(filename) // Clean up on error
		return fmt.Errorf("failed to generate completion: %w", err)
	}

	output.Success("Installed %s completion: %s", shell, filename)
	return nil
}

// appendCompletionToFile appends completion source to existing file
func (cm *CompletionManager) appendCompletionToFile(filename string, shell CompletionType) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return glideErrors.NewPermissionError(
			filename,
			fmt.Sprintf("could not open completion file: %s", filename),
			glideErrors.WithSuggestions("Check file permissions"),
		)
	}
	defer file.Close()

	// Add source line for completion
	sourceLine := fmt.Sprintf("\n# %s completion\nsource <(%s completion %s)\n", branding.ProjectName, branding.CommandName, shell)
	if _, err := file.WriteString(sourceLine); err != nil {
		return fmt.Errorf("failed to write completion source: %w", err)
	}

	output.Success("Added %s completion source to: %s", shell, filename)
	output.Info("Restart your shell or run: source %s", filename)
	return nil
}

// dirExists checks if a directory exists
func (cm *CompletionManager) dirExists(dir string) bool {
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}

// canCreateDir checks if we can create a directory
func (cm *CompletionManager) canCreateDir(dir string) bool {
	parent := filepath.Dir(dir)
	info, err := os.Stat(parent)
	if err != nil {
		return false
	}

	// Check if parent is writable (simplified check)
	testFile := filepath.Join(parent, fmt.Sprintf(".%s_test", branding.CommandName))
	if file, err := os.Create(testFile); err == nil {
		file.Close()
		os.Remove(testFile)
		return true
	}

	return info.Mode().Perm()&0200 != 0 // Check write permission
}

// setupCompletions configures completion functions for cobra commands
// func (cm *CompletionManager) setupCompletions(rootCmd *cobra.Command) {
// 	// Add completion for format flag
// 	rootCmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
// 		return []string{"table", "json", "yaml", "plain"}, cobra.ShellCompDirectiveNoFileComp
// 	})
//
// 	// Add completion for container services (for logs, shell commands)
// 	containerCompletion := func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
// 		return cm.getContainerCompletions(), cobra.ShellCompDirectiveNoFileComp
// 	}
//
// 	// Add completion for git branches (for worktree command)
// 	branchCompletion := func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
// 		return cm.getBranchCompletions(), cobra.ShellCompDirectiveNoFileComp
// 	}
//
// 	// Add completion for configuration keys
// 	configKeyCompletion := func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
// 		return cm.getConfigKeyCompletions(), cobra.ShellCompDirectiveNoFileComp
// 	}
//
// 	// We'll register these with specific commands when they're created
// 	_ = containerCompletion
// 	_ = branchCompletion
// 	_ = configKeyCompletion
// }

// getContainerCompletions returns available Docker containers/services
func (cm *CompletionManager) getContainerCompletions() []string {
	if cm.ctx == nil || len(cm.ctx.ComposeFiles) == 0 {
		return []string{}
	}

	manager := docker.NewContainerManager(cm.ctx)

	// Try to get running containers first
	containers, err := manager.GetStatus()
	if err == nil && len(containers) > 0 {
		var running []string
		for _, container := range containers {
			if container.State == "running" {
				running = append(running, container.Service)
			}
		}
		if len(running) > 0 {
			return running
		}
	}

	// Fallback to compose services
	services, err := manager.GetComposeServices()
	if err != nil {
		// Default services for project
		return []string{"php", "mysql", "nginx", "redis"}
	}

	return services
}

// getBranchCompletions returns available git branches
func (cm *CompletionManager) getBranchCompletions() []string {
	// This could be enhanced to actually query git branches
	// For now, return common patterns
	return []string{
		"feature/",
		"bugfix/",
		"hotfix/",
		"issue-",
	}
}

// getConfigKeyCompletions returns available configuration keys
func (cm *CompletionManager) getConfigKeyCompletions() []string {
	return []string{
		"defaults.test.parallel",
		"defaults.test.processes",
		"defaults.test.coverage",
		"defaults.test.verbose",
		"defaults.docker.compose_timeout",
		"defaults.docker.auto_start",
		"defaults.docker.remove_orphans",
		"defaults.colors.enabled",
		"defaults.worktree.auto_setup",
		"defaults.worktree.copy_env",
		"defaults.worktree.run_migrations",
		"default_project",
	}
}

// RegisterCommandCompletions registers completion functions for specific commands
func (cm *CompletionManager) RegisterCommandCompletions(rootCmd *cobra.Command) {
	// This will be called after all commands are added to register completions
	cm.walkCommands(rootCmd, func(cmd *cobra.Command) {
		switch cmd.Name() {
		case "logs", "shell":
			// Container name completion
			cmd.RegisterFlagCompletionFunc("service", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
				return cm.getContainerCompletions(), cobra.ShellCompDirectiveNoFileComp
			})

			// For positional arguments (service name)
			cmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
				if len(args) == 0 {
					return cm.getContainerCompletions(), cobra.ShellCompDirectiveNoFileComp
				}
				return []string{}, cobra.ShellCompDirectiveNoFileComp
			}

		case "worktree":
			// Branch name completion for worktree command
			cmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
				if len(args) == 0 {
					return cm.getBranchCompletions(), cobra.ShellCompDirectiveNoSpace
				}
				return []string{}, cobra.ShellCompDirectiveNoFileComp
			}

			// Migration options completion
			cmd.RegisterFlagCompletionFunc("migrate", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
				return []string{"fresh-seed", "fresh", "pending", "skip"}, cobra.ShellCompDirectiveNoFileComp
			})

		case "config":
			// Config key completion
			cmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
				if len(args) == 0 {
					return []string{"get", "set", "list"}, cobra.ShellCompDirectiveNoFileComp
				}
				if len(args) == 1 && (args[0] == "get" || args[0] == "set") {
					return cm.getConfigKeyCompletions(), cobra.ShellCompDirectiveNoFileComp
				}
				return []string{}, cobra.ShellCompDirectiveNoFileComp
			}
		}
	})
}

// createMockRootCommand creates a minimal root command for completion generation
func (cm *CompletionManager) createMockRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   branding.CommandName,
		Short: branding.GetShortDescription(),
	}

	// Add global flags with completion
	rootCmd.PersistentFlags().String("config", "", fmt.Sprintf("config file (default is $HOME/%s)", branding.ConfigFileName))
	rootCmd.PersistentFlags().String("format", "table", "Output format (table, json, yaml, plain)")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress non-error output")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable colored output")

	// Register format flag completion
	rootCmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"table", "json", "yaml", "plain"}, cobra.ShellCompDirectiveNoFileComp
	})

	// Add mock commands for completion structure
	cm.addMockCommands(rootCmd)

	return rootCmd
}

// addMockCommands adds minimal command structure for completion
func (cm *CompletionManager) addMockCommands(rootCmd *cobra.Command) {
	// Core commands
	rootCmd.AddCommand(&cobra.Command{Use: "setup", Short: "Interactive setup for Glide CLI"})
	rootCmd.AddCommand(&cobra.Command{Use: "config", Short: "Manage Glide configuration"})
	rootCmd.AddCommand(&cobra.Command{Use: "completion", Short: "Generate shell completion scripts"})

	// Docker commands
	rootCmd.AddCommand(&cobra.Command{Use: "up", Short: "Start Docker containers"})
	rootCmd.AddCommand(&cobra.Command{Use: "down", Short: "Stop Docker containers"})
	rootCmd.AddCommand(&cobra.Command{Use: "docker", Short: "Pass-through to docker-compose"})

	// Development commands
	rootCmd.AddCommand(&cobra.Command{Use: "test", Short: "Run tests"})
	rootCmd.AddCommand(&cobra.Command{Use: "artisan", Short: "Run Artisan commands"})
	rootCmd.AddCommand(&cobra.Command{Use: "composer", Short: "Run Composer commands"})

	// Interactive commands
	logsCmd := &cobra.Command{Use: "logs [service]", Short: "View container logs"}
	logsCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return cm.getContainerCompletions(), cobra.ShellCompDirectiveNoFileComp
		}
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	}
	rootCmd.AddCommand(logsCmd)

	shellCmd := &cobra.Command{Use: "shell [service]", Short: "Attach to container shell"}
	shellCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return cm.getContainerCompletions(), cobra.ShellCompDirectiveNoFileComp
		}
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	}
	rootCmd.AddCommand(shellCmd)

	rootCmd.AddCommand(&cobra.Command{Use: "mysql", Short: "Access MySQL CLI"})
	rootCmd.AddCommand(&cobra.Command{Use: "status", Short: "Show container status"})

	// Utility commands
	rootCmd.AddCommand(&cobra.Command{Use: "lint", Short: "Run PHP CS Fixer"})
	rootCmd.AddCommand(&cobra.Command{Use: "ecr-login", Short: "Authenticate with AWS ECR"})
	rootCmd.AddCommand(&cobra.Command{Use: "db-tunnel", Short: "Create SSH tunnel to database"})
	rootCmd.AddCommand(&cobra.Command{Use: "ssl-certs", Short: "Generate SSL certificates"})
	rootCmd.AddCommand(&cobra.Command{Use: "mysql-fix-permissions", Short: "Fix MySQL permissions"})

	// Global commands (multi-worktree mode)
	if cm.ctx != nil && cm.ctx.DevelopmentMode == context.ModeMultiWorktree {
		globalCmd := &cobra.Command{Use: "global", Aliases: []string{"g"}, Short: "Global commands for multi-worktree mode"}

		worktreeCmd := &cobra.Command{Use: "worktree [branch]", Short: "Create new worktree"}
		worktreeCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return cm.getBranchCompletions(), cobra.ShellCompDirectiveNoSpace
			}
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}
		worktreeCmd.Flags().String("migrate", "", "Migration option")
		worktreeCmd.RegisterFlagCompletionFunc("migrate", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{"fresh-seed", "fresh", "pending", "skip"}, cobra.ShellCompDirectiveNoFileComp
		})

		globalCmd.AddCommand(worktreeCmd)
		globalCmd.AddCommand(&cobra.Command{Use: "status", Short: "Show status of all worktrees"})
		globalCmd.AddCommand(&cobra.Command{Use: "down", Short: "Stop containers in all worktrees"})
		globalCmd.AddCommand(&cobra.Command{Use: "list", Short: "List all worktrees"})
		globalCmd.AddCommand(&cobra.Command{Use: "clean", Short: "Clean up resources"})

		rootCmd.AddCommand(globalCmd)
	}
}

// walkCommands recursively walks through all commands to apply completions
func (cm *CompletionManager) walkCommands(cmd *cobra.Command, fn func(*cobra.Command)) {
	fn(cmd)
	for _, child := range cmd.Commands() {
		cm.walkCommands(child, fn)
	}
}
