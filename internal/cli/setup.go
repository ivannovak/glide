package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ivannovak/glide/internal/config"
	"github.com/ivannovak/glide/internal/context"
	"github.com/ivannovak/glide/internal/docker"
	glideErrors "github.com/ivannovak/glide/pkg/errors"
	"github.com/ivannovak/glide/pkg/output"
	"github.com/ivannovak/glide/pkg/prompt"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// SetupCommand handles the interactive setup process
type SetupCommand struct {
	ctx            *context.ProjectContext
	cfg            *config.Config
	verbose        bool
	nonInteractive bool
	mode           string
	location       string
}

// NewSetupCommand creates a new setup command
func NewSetupCommand(ctx *context.ProjectContext, cfg *config.Config) *cobra.Command {
	setup := &SetupCommand{
		ctx: ctx,
		cfg: cfg,
	}

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Interactive setup for Glide CLI",
		Long: `Initialize Glide for your project with interactive configuration.
This command will:
- Detect or configure your project structure
- Set up development mode (multi-worktree or single-repo)
- Create necessary directories
- Initialize or update ~/.glide.yml configuration`,
		RunE:          setup.run,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.Flags().BoolVar(&setup.verbose, "verbose", false, "Enable verbose output")
	cmd.Flags().BoolVar(&setup.nonInteractive, "non-interactive", false, "Run in non-interactive mode")
	cmd.Flags().StringVar(&setup.mode, "mode", "", "Development mode: multi-worktree or single-repo")
	cmd.Flags().StringVar(&setup.location, "path", "", "Project path (defaults to current directory)")

	return cmd
}

func (s *SetupCommand) run(cmd *cobra.Command, args []string) error {
	output.Info("🚀 Welcome to Glide Setup!")
	output.Println()

	// Check prerequisites
	if err := s.checkPrerequisites(); err != nil {
		return err
	}

	// Check for existing installation
	existingMode, existingProject := s.checkExistingInstallation()
	if existingProject != nil {
		return s.handleExistingInstallation(existingProject, existingMode)
	}

	// Get project location
	projectPath, err := s.getProjectLocation()
	if err != nil {
		return glideErrors.Wrap(err, "failed to get project location",
			glideErrors.WithSuggestions(
				"Ensure you have proper permissions",
				"Try specifying the path explicitly: glidesetup --path /path/to/project",
			),
		)
	}

	// Select development mode
	mode, err := s.selectDevelopmentMode()
	if err != nil {
		return glideErrors.Wrap(err, "failed to select development mode",
			glideErrors.WithSuggestions(
				"Use --mode flag to specify: glidesetup --mode multi-worktree",
				"Valid modes: multi-worktree, single-repo",
			),
		)
	}

	// Create project structure
	if err := s.createProjectStructure(projectPath, mode); err != nil {
		return glideErrors.Wrap(err, "failed to create project structure",
			glideErrors.WithSuggestions(
				"Check directory permissions",
				"Ensure the path exists and is writable",
				"Try running with elevated permissions: sudo glidesetup",
			),
		)
	}

	// Update configuration
	if err := s.updateConfiguration(projectPath, mode); err != nil {
		return glideErrors.Wrap(err, "failed to update configuration",
			glideErrors.WithSuggestions(
				"Check write permissions for ~/.glide.yml",
				"Ensure HOME environment variable is set",
				"Try creating the config manually: touch ~/.glide.yml",
			),
		)
	}

	// Success message
	s.printSuccessMessage(projectPath, mode)

	return nil
}

func (s *SetupCommand) checkPrerequisites() error {
	output.Info("📋 Checking prerequisites...")

	prerequisites := []struct {
		name    string
		command string
		message string
	}{
		{
			name:    "Git",
			command: "git",
			message: "Git is required. Install from: https://git-scm.com",
		},
		{
			name:    "Docker",
			command: "docker",
			message: "Docker is required. Install from: https://www.docker.com",
		},
		{
			name:    "Docker Compose",
			command: "docker compose",
			message: "Docker Compose is required. Install Docker Desktop or docker-compose-plugin",
		},
	}

	allGood := true
	for _, prereq := range prerequisites {
		if err := s.checkCommand(prereq.command); err != nil {
			output.Error("  ✗ %s not found", prereq.name)
			output.Printf("    %s\n", prereq.message)
			allGood = false
		} else {
			output.Success("  ✓ %s found", prereq.name)
		}
	}

	if !allGood {
		return glideErrors.NewDependencyError("prerequisites", "missing prerequisites",
			glideErrors.WithSuggestions(
				"Install Git from: https://git-scm.com",
				"Install Docker from: https://www.docker.com",
				"Ensure all required tools are in your PATH",
			),
		)
	}

	output.Println()
	return nil
}

func (s *SetupCommand) checkCommand(command string) error {
	parts := strings.Split(command, " ")
	cmd := parts[0]
	args := []string{"--version"}
	if len(parts) > 1 {
		_ = append(parts[1:], args...)
	}

	if _, err := exec.LookPath(cmd); err != nil {
		return err
	}
	return nil
}

func (s *SetupCommand) checkExistingInstallation() (context.DevelopmentMode, *config.ProjectConfig) {
	// Check current directory structure
	detector, err := context.NewDetector()
	if err != nil {
		return context.ModeUnknown, nil
	}
	ctx, _ := detector.Detect()

	if ctx != nil && ctx.DevelopmentMode != context.ModeUnknown {
		// Found existing project structure
		return ctx.DevelopmentMode, nil
	}

	// Check if this project is in the config
	if s.cfg != nil {
		for _, project := range s.cfg.Projects {
			absPath, _ := filepath.Abs(project.Path)
			currentPath, _ := filepath.Abs(s.ctx.WorkingDir)

			if strings.HasPrefix(currentPath, absPath) {
				return context.DevelopmentMode(project.Mode), &project
			}
		}
	}

	return context.ModeUnknown, nil
}

func (s *SetupCommand) handleExistingInstallation(project *config.ProjectConfig, mode context.DevelopmentMode) error {
	output.Warning("⚠️  Existing Glide installation detected!")

	if project != nil {
		output.Printf("Path: %s\n", project.Path)
		output.Printf("Mode: %s\n", project.Mode)
	} else {
		output.Printf("Mode: %s\n", mode)
		output.Printf("Location: %s\n", s.ctx.WorkingDir)
	}

	output.Println()

	if s.nonInteractive {
		return glideErrors.NewConfigError("existing installation found",
			glideErrors.WithSuggestions(
				"Use --force flag to override existing setup",
				"Run interactively to choose conversion options",
				"Use 'glideconfig' to modify existing configuration",
			),
		)
	}

	// Ask what to do
	options := []string{
		"Convert to different mode (single-repo ↔ multi-worktree)",
		"Reconfigure current setup",
		"Exit",
	}

	idx, _, err := prompt.Select("What would you like to do?", options, 2)
	if err != nil {
		return glideErrors.Wrap(err, "failed to get user choice",
			glideErrors.WithSuggestions(
				"Try running setup in interactive mode",
				"Use arrow keys to select an option",
			),
		)
	}

	switch idx {
	case 0:
		return s.convertMode(project, mode)
	case 1:
		return s.reconfigure(project, mode)
	case 2:
		return nil
	default:
		return nil
	}
}

func (s *SetupCommand) convertMode(project *config.ProjectConfig, currentMode context.DevelopmentMode) error {
	newMode := context.ModeSingleRepo
	if currentMode == context.ModeSingleRepo {
		newMode = context.ModeMultiWorktree
	}

	output.Printf("\n🔄 Converting from %s to %s mode...\n", currentMode, newMode)

	projectPath := s.ctx.WorkingDir
	if project != nil {
		projectPath = project.Path
	}

	// Create new structure
	if err := s.createProjectStructure(projectPath, newMode); err != nil {
		return glideErrors.Wrap(err, "failed to create new structure during mode conversion",
			glideErrors.WithSuggestions(
				"Check directory permissions",
				"Ensure sufficient disk space",
				"Manually create required directories",
			),
		)
	}

	// Update configuration
	if err := s.updateConfiguration(projectPath, newMode); err != nil {
		return glideErrors.Wrap(err, "failed to update configuration during mode conversion",
			glideErrors.WithSuggestions(
				"Check ~/.glide.yml permissions",
				"Manually edit ~/.glide.yml to update the mode",
			),
		)
	}

	if newMode == context.ModeMultiWorktree {
		output.Info("\n⚠️  Mode conversion complete!")
		output.Info("Next steps:")
		output.Info("1. Move your existing repository to vcs/")
		output.Info("2. Update your git remote URLs")
		output.Info("3. Create worktrees with: glideworktree <branch>")
	}

	return nil
}

func (s *SetupCommand) reconfigure(project *config.ProjectConfig, mode context.DevelopmentMode) error {
	output.Info("\n♻️  Reconfiguring existing setup...")

	projectPath := s.ctx.WorkingDir
	if project != nil {
		projectPath = project.Path
	}

	// Just update the configuration
	return s.updateConfiguration(projectPath, mode)
}

func (s *SetupCommand) getProjectLocation() (string, error) {
	if s.location != "" {
		return filepath.Abs(s.location)
	}

	if s.nonInteractive {
		return os.Getwd()
	}

	cwd, _ := os.Getwd()

	input, err := prompt.InputPath("📁 Project location", cwd)
	if err != nil {
		return "", err
	}

	if input == "" {
		return cwd, nil
	}

	// Expand tilde
	if strings.HasPrefix(input, "~/") {
		home, _ := os.UserHomeDir()
		input = filepath.Join(home, input[2:])
	}

	return filepath.Abs(input)
}

func (s *SetupCommand) selectDevelopmentMode() (context.DevelopmentMode, error) {
	if s.mode != "" {
		switch s.mode {
		case "multi-worktree", "multi":
			return context.ModeMultiWorktree, nil
		case "single-repo", "single":
			return context.ModeSingleRepo, nil
		default:
			return context.ModeUnknown, glideErrors.NewConfigError(fmt.Sprintf("invalid mode: %s", s.mode),
				glideErrors.WithSuggestions(
					"Valid modes: multi-worktree, multi, single-repo, single",
					"Use 'multi-worktree' for team development",
					"Use 'single-repo' for simple setups",
				),
			)
		}
	}

	if s.nonInteractive {
		// Default to single-repo for non-interactive
		return context.ModeSingleRepo, nil
	}

	output.Info("\n🎯 Select development mode:")
	output.Println()
	output.Info("1. Multi-worktree (recommended for team development)")
	output.Info("   - Work on multiple branches simultaneously")
	output.Info("   - Isolated Docker environments per branch")
	output.Info("   - Main repo in vcs/, worktrees in worktrees/")
	output.Println()
	output.Info("2. Single-repository (simpler setup)")
	output.Info("   - Traditional single checkout")
	output.Info("   - Switch branches manually")
	output.Info("   - Single Docker environment")
	output.Println()

	options := []string{
		"Multi-worktree (recommended)",
		"Single-repository",
	}

	idx, _, err := prompt.Select("Select development mode", options, 0)
	if err != nil {
		return context.ModeUnknown, glideErrors.Wrap(err, "failed to select development mode",
			glideErrors.WithSuggestions(
				"Try running setup in interactive mode",
				"Use --mode flag to specify: glidesetup --mode multi-worktree",
			),
		)
	}

	switch idx {
	case 0:
		return context.ModeMultiWorktree, nil
	case 1:
		return context.ModeSingleRepo, nil
	default:
		return context.ModeMultiWorktree, nil
	}
}

func (s *SetupCommand) createProjectStructure(projectPath string, mode context.DevelopmentMode) error {
	output.Printf("\n📂 Creating project structure in %s...\n", projectPath)

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return glideErrors.NewPermissionError(projectPath, fmt.Sprintf("failed to create project directory: %s", projectPath),
			glideErrors.WithError(err),
			glideErrors.WithSuggestions(
				"Check parent directory permissions",
				"Ensure the path is valid",
				"Try creating the directory manually first",
			),
		)
	}

	if mode == context.ModeMultiWorktree {
		// Create multi-worktree structure
		dirs := []string{
			"vcs",
			"worktrees",
			"scripts",
		}

		for _, dir := range dirs {
			fullPath := filepath.Join(projectPath, dir)
			if err := os.MkdirAll(fullPath, 0755); err != nil {
				return glideErrors.NewPermissionError(fullPath, fmt.Sprintf("failed to create directory: %s", dir),
					glideErrors.WithError(err),
					glideErrors.WithSuggestions(
						"Check directory permissions",
						fmt.Sprintf("Manually create: mkdir -p %s", fullPath),
					),
				)
			}
			output.Success("  ✓ Created %s/", dir)
		}

		// Create .gitignore if it doesn't exist
		gitignorePath := filepath.Join(projectPath, ".gitignore")
		if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
			gitignoreContent := `# Glide multi-worktree structure
vcs/
worktrees/
docker-compose.override.yml
.env.local
*.log
`
			if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
				return glideErrors.NewPermissionError(gitignorePath, "failed to create .gitignore",
					glideErrors.WithError(err),
					glideErrors.WithSuggestions(
						"Check write permissions in the project directory",
						"Create the file manually if needed",
					),
				)
			}
			output.Success("  ✓ Created .gitignore")
		}

		// Create README if it doesn't exist
		readmePath := filepath.Join(projectPath, "README.md")
		if _, err := os.Stat(readmePath); os.IsNotExist(err) {
			readmeContent := fmt.Sprintf(`# %s

This is a multi-worktree development environment managed by Glide.

## Structure

- **vcs/**: Main repository (kept on main branch as reference)
- **worktrees/**: Feature branch worktrees (do all work here)
- **scripts/**: Management scripts
- **docker-compose.override.yml**: Shared Docker configuration

## Quick Start

1. Clone your repository into vcs/:
   `+"```bash"+`
   git clone <your-repo-url> vcs
   `+"```"+`

2. Create a new worktree:
   `+"```bash"+`
   glideworktree feature/my-feature
   `+"```"+`

3. Start Docker:
   `+"```bash"+`
   cd worktrees/feature-my-feature
   glideup
   `+"```"+`

## Commands

- `+"`glideworktree <branch>`"+`: Create a new worktree
- `+"`glidestatus`"+`: Show Docker status across all worktrees
- `+"`glidedown-all`"+`: Stop all Docker containers
`, filepath.Base(projectPath))

			if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
				return glideErrors.NewPermissionError(readmePath, "failed to create README.md",
					glideErrors.WithError(err),
					glideErrors.WithSuggestions(
						"Check write permissions in the project directory",
						"Create the file manually if needed",
					),
				)
			}
			output.Success("  ✓ Created README.md")
		}
	} else {
		// Single-repo mode - just ensure the directory exists
		output.Success("  ✓ Using existing directory structure")
	}

	return nil
}

func (s *SetupCommand) updateConfiguration(projectPath string, mode context.DevelopmentMode) error {
	output.Info("\n⚙️  Updating configuration...")

	// Get or create config
	cfg := s.cfg
	if cfg == nil {
		cfg = &config.Config{
			Projects: make(map[string]config.ProjectConfig),
			Defaults: config.DefaultsConfig{
				Test: config.TestDefaults{
					Parallel:  true,
					Processes: 8,
				},
				Docker: config.DockerDefaults{
					ComposeTimeout: 30,
					AutoStart:      true,
					RemoveOrphans:  true,
				},
				Colors: config.ColorDefaults{
					Enabled: "auto",
				},
				Worktree: config.WorktreeDefaults{
					AutoSetup:     true,
					CopyEnv:       true,
					RunMigrations: false,
				},
			},
		}
	}

	// Get project name
	projectName := filepath.Base(projectPath)
	if !s.nonInteractive {
		projectName, _ = prompt.Input("Project name", projectName, nil)
	}

	// Add or update project
	cfg.Projects[projectName] = config.ProjectConfig{
		Path: projectPath,
		Mode: string(mode),
	}

	// Set as default if it's the only project
	if len(cfg.Projects) == 1 {
		cfg.DefaultProject = projectName
	} else if !s.nonInteractive {
		setDefault, _ := prompt.Confirm("Set as default project?", false)
		if setDefault {
			cfg.DefaultProject = projectName
		}
	}

	// Save configuration
	configPath := filepath.Join(os.Getenv("HOME"), ".glide.yml")

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return glideErrors.NewConfigError("failed to marshal configuration",
			glideErrors.WithError(err),
			glideErrors.WithSuggestions(
				"Check configuration structure",
				"Report this as a bug if it persists",
			),
		)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return glideErrors.NewPermissionError(configPath, "failed to write configuration file",
			glideErrors.WithError(err),
			glideErrors.WithSuggestions(
				"Check write permissions for ~/.glide.yml",
				"Ensure HOME directory exists and is writable",
				fmt.Sprintf("Try creating manually: touch %s", configPath),
			),
		)
	}

	output.Success("  ✓ Updated ~/.glide.yml")

	// Validate Docker setup if compose files exist
	tempCtx := &context.ProjectContext{
		ProjectRoot:     projectPath,
		WorkingDir:      projectPath,
		DevelopmentMode: mode,
	}

	resolver := docker.NewResolver(tempCtx)
	if err := resolver.Resolve(); err == nil && len(tempCtx.ComposeFiles) > 0 {
		if err := resolver.ValidateSetup(); err != nil {
			output.Warning("  ⚠ Docker validation warning: %v", err)
		} else {
			output.Success("  ✓ Docker setup validated")
		}
	}

	return nil
}

func (s *SetupCommand) printSuccessMessage(projectPath string, mode context.DevelopmentMode) {
	// Install shell completions
	output.Println()
	output.Info("Installing shell completions...")
	completionManager := NewCompletionManager(s.ctx, s.cfg)
	if err := completionManager.InstallCompletion(); err != nil {
		output.Warning("Could not install shell completions: %v", err)
		output.Info("You can install manually with: glidecompletion [bash|zsh|fish]")
	}

	output.Println()
	output.Success("✅ Setup complete!")
	output.Println()

	output.Info("Your project is configured at:", projectPath)
	output.Printf("Development mode: %s\n", mode)
	output.Println()

	if mode == context.ModeMultiWorktree {
		output.Info("Next steps:")
		output.Info("1. Clone your repository into vcs/:")
		output.Printf("   cd %s\n", projectPath)
		output.Info("   git clone <your-repo-url> vcs")
		output.Println()
		output.Info("2. Create your first worktree:")
		output.Info("   glideworktree feature/my-feature")
		output.Println()
		output.Info("3. Start developing!")
		output.Info("   cd worktrees/feature-my-feature")
		output.Info("   glideup")
	} else {
		output.Info("Next steps:")
		output.Info("1. Navigate to your project:")
		output.Printf("   cd %s\n", projectPath)
		output.Println()
		output.Info("2. Start Docker:")
		output.Info("   glideup")
		output.Println()
		output.Info("3. Run tests:")
		output.Info("   glidetest")
	}

	output.Println()
	output.Info("💡 Pro tip: Tab completion is now available! Restart your shell to enable it.")
	output.Info("Run 'glide--help' to see available commands")
}
