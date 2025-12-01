package cli

import (
	"fmt"
	"os"
	"path/filepath"
	// "reflect"
	"strconv"
	"strings"

	"github.com/ivannovak/glide/v3/internal/config"
	glideErrors "github.com/ivannovak/glide/v3/pkg/errors"
	"github.com/ivannovak/glide/v3/pkg/output"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// ConfigCommand handles configuration management
type ConfigCommand struct {
	cfg     *config.Config
	cfgPath string
}

// NewConfigCommand creates the config command group
func NewConfigCommand(cfg *config.Config) *cobra.Command {
	cc := &ConfigCommand{
		cfg:     cfg,
		cfgPath: filepath.Join(os.Getenv("HOME"), ".glide.yml"),
	}

	cmd := &cobra.Command{
		Use:           "config",
		Short:         "Manage Glide configuration",
		Long:          `View and modify Glide configuration settings stored in ~/.glide.yml`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Add subcommands
	cmd.AddCommand(cc.newGetCommand())
	cmd.AddCommand(cc.newSetCommand())
	cmd.AddCommand(cc.newListCommand())
	cmd.AddCommand(cc.newUseCommand())

	return cmd
}

// newGetCommand creates the config get subcommand
func (cc *ConfigCommand) newGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Long: `Get a configuration value by key.

Examples:
  glide config get default_project
  glide config get defaults.docker.auto_start
  glide config get projects.myproject.mode`,
		Args:          cobra.ExactArgs(1),
		RunE:          cc.runGet,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

// newSetCommand creates the config set subcommand
func (cc *ConfigCommand) newSetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long: `Set a configuration value by key.

Examples:
  glide config set default_project myproject
  glide config set defaults.docker.auto_start true
  glide config set defaults.test.processes 10
  glide config set projects.myproject.path /path/to/project`,
		Args:          cobra.ExactArgs(2),
		RunE:          cc.runSet,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

// newListCommand creates the config list subcommand
func (cc *ConfigCommand) newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:           "list",
		Short:         "List all configuration settings",
		Long:          `Display all configuration settings in a readable format.`,
		Args:          cobra.NoArgs,
		RunE:          cc.runList,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

// newUseCommand creates the config use subcommand for project switching
func (cc *ConfigCommand) newUseCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "use <project>",
		Short: "Switch to a different project",
		Long: `Set the default project to use when running Glide commands.

Example:
  glide config use myproject`,
		Args:          cobra.ExactArgs(1),
		RunE:          cc.runUse,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

// runGet handles the config get command
func (cc *ConfigCommand) runGet(cmd *cobra.Command, args []string) error {
	if cc.cfg == nil {
		return glideErrors.NewConfigError(fmt.Sprintf("no configuration file found at %s", cc.cfgPath),
			glideErrors.WithSuggestions(
				"Run 'glide setup' to create a configuration file",
				"Check if the config file exists at the expected path",
			))
	}

	key := args[0]
	value, err := cc.getValue(key)
	if err != nil {
		return err
	}

	output.Println(value)
	return nil
}

// runSet handles the config set command
func (cc *ConfigCommand) runSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	// Load current config or create new one
	if cc.cfg == nil {
		cc.cfg = &config.Config{
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

	// Set the value
	if err := cc.setValue(key, value); err != nil {
		return err
	}

	// Save the configuration
	if err := cc.save(); err != nil {
		return glideErrors.Wrap(err, "failed to save configuration",
			glideErrors.WithSuggestions(
				"Check file permissions on the config directory",
				"Ensure you have write access to ~/.glide.yml",
				"Try running with elevated permissions if necessary",
			))
	}

	output.Success("✓ Set %s = %s", key, value)
	return nil
}

// runList handles the config list command
func (cc *ConfigCommand) runList(cmd *cobra.Command, args []string) error {
	if cc.cfg == nil {
		output.Info("No configuration file found at %s", cc.cfgPath)
		output.Info("Run 'glide setup' to create one.")
		return nil
	}

	output.Info("=== Glide Configuration ===")
	output.Printf("Config file: %s\n\n", cc.cfgPath)

	// Display default project
	if cc.cfg.DefaultProject != "" {
		output.Success("Default Project: %s", cc.cfg.DefaultProject)
	} else {
		output.Warning("Default Project: (none)")
	}
	output.Println()

	// Display projects
	if len(cc.cfg.Projects) > 0 {
		output.Info("Projects:")
		for name, project := range cc.cfg.Projects {
			output.Printf("  %s:\n", name)
			output.Printf("    Path: %s\n", project.Path)
			output.Printf("    Mode: %s\n", project.Mode)
		}
		output.Println()
	}

	// Display defaults
	output.Info("Defaults:")

	output.Println("  Test:")
	output.Printf("    Parallel: %v\n", cc.cfg.Defaults.Test.Parallel)
	output.Printf("    Processes: %d\n", cc.cfg.Defaults.Test.Processes)
	output.Printf("    Coverage: %v\n", cc.cfg.Defaults.Test.Coverage)
	output.Printf("    Verbose: %v\n", cc.cfg.Defaults.Test.Verbose)

	output.Println("  Docker:")
	output.Printf("    Compose Timeout: %d seconds\n", cc.cfg.Defaults.Docker.ComposeTimeout)
	output.Printf("    Auto Start: %v\n", cc.cfg.Defaults.Docker.AutoStart)
	output.Printf("    Remove Orphans: %v\n", cc.cfg.Defaults.Docker.RemoveOrphans)

	output.Println("  Colors:")
	output.Printf("    Enabled: %s\n", cc.cfg.Defaults.Colors.Enabled)

	output.Println("  Worktree:")
	output.Printf("    Auto Setup: %v\n", cc.cfg.Defaults.Worktree.AutoSetup)
	output.Printf("    Copy Env: %v\n", cc.cfg.Defaults.Worktree.CopyEnv)
	output.Printf("    Run Migrations: %v\n", cc.cfg.Defaults.Worktree.RunMigrations)

	return nil
}

// runUse handles the config use command for project switching
func (cc *ConfigCommand) runUse(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	// Load config if not loaded
	if cc.cfg == nil {
		return glideErrors.NewConfigError(fmt.Sprintf("no configuration file found at %s", cc.cfgPath),
			glideErrors.WithSuggestions(
				"Run 'glide setup' to create a configuration file",
				"Check if the config file exists at the expected path",
			))
	}

	// Check if project exists
	if _, exists := cc.cfg.Projects[projectName]; !exists {
		// List available projects
		var available []string
		for name := range cc.cfg.Projects {
			available = append(available, name)
		}

		if len(available) > 0 {
			return glideErrors.NewConfigError(fmt.Sprintf("project '%s' not found\nAvailable projects: %s",
				projectName, strings.Join(available, ", ")),
				glideErrors.WithSuggestions(
					"Use one of the listed available projects",
					fmt.Sprintf("Add the project using 'glide config set projects.%s.path /path/to/project'", projectName),
					"Run 'glide config list' to see all configured projects",
				))
		}
		return glideErrors.NewConfigError(fmt.Sprintf("project '%s' not found\nNo projects configured. Run 'glide setup' first", projectName),
			glideErrors.WithSuggestions(
				"Run 'glide setup' to initialize your first project",
				fmt.Sprintf("Add a project manually using 'glide config set projects.%s.path /path/to/project'", projectName),
			))
	}

	// Set as default project
	cc.cfg.DefaultProject = projectName

	// Save configuration
	if err := cc.save(); err != nil {
		return glideErrors.Wrap(err, "failed to save configuration",
			glideErrors.WithSuggestions(
				"Check file permissions on the config directory",
				"Ensure you have write access to ~/.glide.yml",
				"Try running with elevated permissions if necessary",
			))
	}

	project := cc.cfg.Projects[projectName]
	output.Success("✓ Switched to project: %s", projectName)
	output.Printf("  Path: %s\n", project.Path)
	output.Printf("  Mode: %s\n", project.Mode)

	return nil
}

// getValue retrieves a value from the config using dot notation
func (cc *ConfigCommand) getValue(key string) (string, error) {
	parts := strings.Split(key, ".")

	// Handle special cases
	if len(parts) == 1 && parts[0] == "default_project" {
		if cc.cfg.DefaultProject == "" {
			return "(none)", nil
		}
		return cc.cfg.DefaultProject, nil
	}

	// Handle projects
	if len(parts) >= 2 && parts[0] == "projects" {
		projectName := parts[1]
		project, exists := cc.cfg.Projects[projectName]
		if !exists {
			return "", glideErrors.NewConfigError(fmt.Sprintf("project '%s' not found", projectName),
				glideErrors.WithSuggestions(
					"Run 'glide config list' to see available projects",
					fmt.Sprintf("Add the project using 'glide config set projects.%s.path /path/to/project'", projectName),
				))
		}

		if len(parts) == 2 {
			return fmt.Sprintf("path=%s mode=%s", project.Path, project.Mode), nil
		}

		switch parts[2] {
		case "path":
			return project.Path, nil
		case "mode":
			return project.Mode, nil
		default:
			return "", glideErrors.NewConfigError(fmt.Sprintf("unknown project field: %s", parts[2]),
				glideErrors.WithSuggestions(
					"Valid project fields are: 'path', 'mode'",
					fmt.Sprintf("Use 'glide config get projects.%s' to see all fields", parts[1]),
				))
		}
	}

	// Handle defaults
	if len(parts) >= 2 && parts[0] == "defaults" {
		return cc.getDefaultValue(parts[1:])
	}

	return "", glideErrors.NewConfigError(fmt.Sprintf("unknown configuration key: %s", key),
		glideErrors.WithSuggestions(
			"Run 'glide config list' to see all available configuration keys",
			"Valid key formats: 'default_project', 'projects.NAME.FIELD', 'defaults.SECTION.FIELD'",
			"Use dot notation to access nested values",
		))
}

// getDefaultValue retrieves a value from the defaults section
func (cc *ConfigCommand) getDefaultValue(path []string) (string, error) {
	if len(path) == 0 {
		return "", glideErrors.NewConfigError("incomplete path",
			glideErrors.WithSuggestions(
				"Provide a complete path like 'defaults.test.parallel'",
				"Run 'glide config list' to see available configuration paths",
			))
	}

	switch path[0] {
	case "test":
		if len(path) == 1 {
			return fmt.Sprintf("parallel=%v processes=%d coverage=%v verbose=%v",
				cc.cfg.Defaults.Test.Parallel,
				cc.cfg.Defaults.Test.Processes,
				cc.cfg.Defaults.Test.Coverage,
				cc.cfg.Defaults.Test.Verbose), nil
		}
		switch path[1] {
		case "parallel":
			return strconv.FormatBool(cc.cfg.Defaults.Test.Parallel), nil
		case "processes":
			return strconv.Itoa(cc.cfg.Defaults.Test.Processes), nil
		case "coverage":
			return strconv.FormatBool(cc.cfg.Defaults.Test.Coverage), nil
		case "verbose":
			return strconv.FormatBool(cc.cfg.Defaults.Test.Verbose), nil
		default:
			return "", glideErrors.NewConfigError(fmt.Sprintf("unknown test field: %s", path[1]),
				glideErrors.WithSuggestions(
					"Valid test fields: 'parallel', 'processes', 'coverage', 'verbose'",
					"Use 'glide config get defaults.test' to see all test settings",
				))
		}

	case "docker":
		if len(path) == 1 {
			return fmt.Sprintf("compose_timeout=%d auto_start=%v remove_orphans=%v",
				cc.cfg.Defaults.Docker.ComposeTimeout,
				cc.cfg.Defaults.Docker.AutoStart,
				cc.cfg.Defaults.Docker.RemoveOrphans), nil
		}
		switch path[1] {
		case "compose_timeout":
			return strconv.Itoa(cc.cfg.Defaults.Docker.ComposeTimeout), nil
		case "auto_start":
			return strconv.FormatBool(cc.cfg.Defaults.Docker.AutoStart), nil
		case "remove_orphans":
			return strconv.FormatBool(cc.cfg.Defaults.Docker.RemoveOrphans), nil
		default:
			return "", glideErrors.NewConfigError(fmt.Sprintf("unknown docker field: %s", path[1]),
				glideErrors.WithSuggestions(
					"Valid docker fields: 'compose_timeout', 'auto_start', 'remove_orphans'",
					"Use 'glide config get defaults.docker' to see all docker settings",
				))
		}

	case "colors":
		if len(path) == 1 {
			return cc.cfg.Defaults.Colors.Enabled, nil
		}
		if path[1] == "enabled" {
			return cc.cfg.Defaults.Colors.Enabled, nil
		}
		return "", glideErrors.NewConfigError(fmt.Sprintf("unknown colors field: %s", path[1]),
			glideErrors.WithSuggestions(
				"Valid colors field: 'enabled'",
				"Use 'glide config get defaults.colors.enabled' to see the current setting",
			))

	case "worktree":
		if len(path) == 1 {
			return fmt.Sprintf("auto_setup=%v copy_env=%v run_migrations=%v",
				cc.cfg.Defaults.Worktree.AutoSetup,
				cc.cfg.Defaults.Worktree.CopyEnv,
				cc.cfg.Defaults.Worktree.RunMigrations), nil
		}
		switch path[1] {
		case "auto_setup":
			return strconv.FormatBool(cc.cfg.Defaults.Worktree.AutoSetup), nil
		case "copy_env":
			return strconv.FormatBool(cc.cfg.Defaults.Worktree.CopyEnv), nil
		case "run_migrations":
			return strconv.FormatBool(cc.cfg.Defaults.Worktree.RunMigrations), nil
		default:
			return "", glideErrors.NewConfigError(fmt.Sprintf("unknown worktree field: %s", path[1]),
				glideErrors.WithSuggestions(
					"Valid worktree fields: 'auto_setup', 'copy_env', 'run_migrations'",
					"Use 'glide config get defaults.worktree' to see all worktree settings",
				))
		}

	default:
		return "", glideErrors.NewConfigError(fmt.Sprintf("unknown defaults section: %s", path[0]),
			glideErrors.WithSuggestions(
				"Valid defaults sections: 'test', 'docker', 'colors', 'worktree'",
				"Use 'glide config list' to see all available sections",
			))
	}
}

// setValue sets a value in the config using dot notation
func (cc *ConfigCommand) setValue(key, value string) error {
	parts := strings.Split(key, ".")

	// Handle special cases
	if len(parts) == 1 && parts[0] == "default_project" {
		// Validate project exists
		if _, exists := cc.cfg.Projects[value]; !exists && value != "" {
			return glideErrors.NewConfigError(fmt.Sprintf("project '%s' does not exist", value),
				glideErrors.WithSuggestions(
					fmt.Sprintf("Create the project first using 'glide config set projects.%s.path /path/to/project'", value),
					"Run 'glide config list' to see existing projects",
					"Use empty string to clear default project: 'glide config set default_project '",
				))
		}
		cc.cfg.DefaultProject = value
		return nil
	}

	// Handle projects
	if len(parts) >= 3 && parts[0] == "projects" {
		projectName := parts[1]

		// Get or create project
		project, exists := cc.cfg.Projects[projectName]
		if !exists {
			project = config.ProjectConfig{}
		}

		switch parts[2] {
		case "path":
			project.Path = value
		case "mode":
			if value != "multi-worktree" && value != "single-repo" {
				return glideErrors.NewConfigError(fmt.Sprintf("invalid mode: %s (must be 'multi-worktree' or 'single-repo')", value),
					glideErrors.WithSuggestions(
						"Use 'multi-worktree' for projects with git worktrees",
						"Use 'single-repo' for traditional single-directory projects",
					))
			}
			project.Mode = value
		default:
			return glideErrors.NewConfigError(fmt.Sprintf("unknown project field: %s", parts[2]),
				glideErrors.WithSuggestions(
					"Valid project fields: 'path', 'mode'",
					fmt.Sprintf("Example: 'glide config set projects.%s.path /path/to/project'", parts[1]),
					fmt.Sprintf("Example: 'glide config set projects.%s.mode multi-worktree'", parts[1]),
				))
		}

		cc.cfg.Projects[projectName] = project
		return nil
	}

	// Handle defaults
	if len(parts) >= 3 && parts[0] == "defaults" {
		return cc.setDefaultValue(parts[1:], value)
	}

	return glideErrors.NewConfigError(fmt.Sprintf("unknown configuration key: %s", key),
		glideErrors.WithSuggestions(
			"Valid key formats: 'default_project', 'projects.NAME.FIELD', 'defaults.SECTION.FIELD'",
			"Run 'glide config list' to see all available configuration keys",
			"Use dot notation to access nested values",
		))
}

// setDefaultValue sets a value in the defaults section
func (cc *ConfigCommand) setDefaultValue(path []string, value string) error {
	if len(path) < 2 {
		return glideErrors.NewConfigError("incomplete path",
			glideErrors.WithSuggestions(
				"Provide a complete path like 'defaults.test.parallel'",
				"Format: 'defaults.SECTION.FIELD' where SECTION is test/docker/colors/worktree",
			))
	}

	switch path[0] {
	case "test":
		switch path[1] {
		case "parallel":
			b, err := strconv.ParseBool(value)
			if err != nil {
				return glideErrors.NewConfigError(fmt.Sprintf("invalid boolean value: %s", value),
					glideErrors.WithSuggestions(
						"Use 'true' or 'false' for boolean values",
						"Case insensitive: 'True', 'FALSE', '1', '0' are also valid",
					))
			}
			cc.cfg.Defaults.Test.Parallel = b
		case "processes":
			n, err := strconv.Atoi(value)
			if err != nil {
				return glideErrors.NewConfigError(fmt.Sprintf("invalid integer value: %s", value),
					glideErrors.WithSuggestions(
						"Provide a valid integer number",
						"Example: 'glide config set defaults.test.processes 8'",
					))
			}
			if n < 1 {
				return glideErrors.NewConfigError("processes must be at least 1",
					glideErrors.WithSuggestions(
						"Set a positive number of test processes (recommended: 4-16)",
						"Higher values may improve test speed but use more system resources",
					))
			}
			cc.cfg.Defaults.Test.Processes = n
		case "coverage":
			b, err := strconv.ParseBool(value)
			if err != nil {
				return glideErrors.NewConfigError(fmt.Sprintf("invalid boolean value: %s", value),
					glideErrors.WithSuggestions(
						"Use 'true' or 'false' for boolean values",
						"Case insensitive: 'True', 'FALSE', '1', '0' are also valid",
					))
			}
			cc.cfg.Defaults.Test.Coverage = b
		case "verbose":
			b, err := strconv.ParseBool(value)
			if err != nil {
				return glideErrors.NewConfigError(fmt.Sprintf("invalid boolean value: %s", value),
					glideErrors.WithSuggestions(
						"Use 'true' or 'false' for boolean values",
						"Case insensitive: 'True', 'FALSE', '1', '0' are also valid",
					))
			}
			cc.cfg.Defaults.Test.Verbose = b
		default:
			return glideErrors.NewConfigError(fmt.Sprintf("unknown test field: %s", path[1]),
				glideErrors.WithSuggestions(
					"Valid test fields: 'parallel', 'processes', 'coverage', 'verbose'",
					"Example: 'glide config set defaults.test.parallel true'",
				))
		}

	case "docker":
		switch path[1] {
		case "compose_timeout":
			n, err := strconv.Atoi(value)
			if err != nil {
				return glideErrors.NewConfigError(fmt.Sprintf("invalid integer value: %s", value),
					glideErrors.WithSuggestions(
						"Provide a valid integer number for timeout in seconds",
						"Example: 'glide config set defaults.docker.compose_timeout 60'",
					))
			}
			if n < 1 {
				return glideErrors.NewConfigError("timeout must be at least 1 second",
					glideErrors.WithSuggestions(
						"Set a positive timeout value in seconds (recommended: 30-120)",
						"Higher values give more time for slower Docker operations",
					))
			}
			cc.cfg.Defaults.Docker.ComposeTimeout = n
		case "auto_start":
			b, err := strconv.ParseBool(value)
			if err != nil {
				return glideErrors.NewConfigError(fmt.Sprintf("invalid boolean value: %s", value),
					glideErrors.WithSuggestions(
						"Use 'true' or 'false' for boolean values",
						"Case insensitive: 'True', 'FALSE', '1', '0' are also valid",
					))
			}
			cc.cfg.Defaults.Docker.AutoStart = b
		case "remove_orphans":
			b, err := strconv.ParseBool(value)
			if err != nil {
				return glideErrors.NewConfigError(fmt.Sprintf("invalid boolean value: %s", value),
					glideErrors.WithSuggestions(
						"Use 'true' or 'false' for boolean values",
						"Case insensitive: 'True', 'FALSE', '1', '0' are also valid",
					))
			}
			cc.cfg.Defaults.Docker.RemoveOrphans = b
		default:
			return glideErrors.NewConfigError(fmt.Sprintf("unknown docker field: %s", path[1]),
				glideErrors.WithSuggestions(
					"Valid docker fields: 'compose_timeout', 'auto_start', 'remove_orphans'",
					"Example: 'glide config set defaults.docker.auto_start false'",
				))
		}

	case "colors":
		if path[1] != "enabled" {
			return glideErrors.NewConfigError(fmt.Sprintf("unknown colors field: %s", path[1]),
				glideErrors.WithSuggestions(
					"Valid colors field: 'enabled'",
					"Example: 'glide config set defaults.colors.enabled never'",
				))
		}
		if value != "auto" && value != "always" && value != "never" {
			return glideErrors.NewConfigError(fmt.Sprintf("invalid color mode: %s (must be 'auto', 'always', or 'never')", value),
				glideErrors.WithSuggestions(
					"Use 'auto' to enable colors in terminals that support them",
					"Use 'always' to force colors even in non-terminal output",
					"Use 'never' to disable colors completely",
				))
		}
		cc.cfg.Defaults.Colors.Enabled = value

	case "worktree":
		switch path[1] {
		case "auto_setup":
			b, err := strconv.ParseBool(value)
			if err != nil {
				return glideErrors.NewConfigError(fmt.Sprintf("invalid boolean value: %s", value),
					glideErrors.WithSuggestions(
						"Use 'true' or 'false' for boolean values",
						"Case insensitive: 'True', 'FALSE', '1', '0' are also valid",
					))
			}
			cc.cfg.Defaults.Worktree.AutoSetup = b
		case "copy_env":
			b, err := strconv.ParseBool(value)
			if err != nil {
				return glideErrors.NewConfigError(fmt.Sprintf("invalid boolean value: %s", value),
					glideErrors.WithSuggestions(
						"Use 'true' or 'false' for boolean values",
						"Case insensitive: 'True', 'FALSE', '1', '0' are also valid",
					))
			}
			cc.cfg.Defaults.Worktree.CopyEnv = b
		case "run_migrations":
			b, err := strconv.ParseBool(value)
			if err != nil {
				return glideErrors.NewConfigError(fmt.Sprintf("invalid boolean value: %s", value),
					glideErrors.WithSuggestions(
						"Use 'true' or 'false' for boolean values",
						"Case insensitive: 'True', 'FALSE', '1', '0' are also valid",
					))
			}
			cc.cfg.Defaults.Worktree.RunMigrations = b
		default:
			return glideErrors.NewConfigError(fmt.Sprintf("unknown worktree field: %s", path[1]),
				glideErrors.WithSuggestions(
					"Valid worktree fields: 'auto_setup', 'copy_env', 'run_migrations'",
					"Example: 'glide config set defaults.worktree.auto_setup true'",
				))
		}

	default:
		return glideErrors.NewConfigError(fmt.Sprintf("unknown defaults section: %s", path[0]),
			glideErrors.WithSuggestions(
				"Valid defaults sections: 'test', 'docker', 'colors', 'worktree'",
				"Example: 'glide config set defaults.test.parallel true'",
			))
	}

	return nil
}

// save writes the configuration to disk
func (cc *ConfigCommand) save() error {
	data, err := yaml.Marshal(cc.cfg)
	if err != nil {
		return glideErrors.Wrap(err, "failed to marshal config",
			glideErrors.WithSuggestions(
				"Check if the configuration data is valid",
				"Try resetting the configuration if it's corrupted",
			))
	}

	if err := os.WriteFile(cc.cfgPath, data, 0644); err != nil {
		return glideErrors.Wrap(err, "failed to write config file",
			glideErrors.WithSuggestions(
				"Check if you have write permissions to ~/.glide.yml",
				"Ensure the home directory is accessible",
				"Try running with elevated permissions if necessary",
			))
	}

	return nil
}

// validateValue validates a configuration value based on its type
// func validateValue(v reflect.Value, value string) error {
// 	switch v.Kind() {
// 	case reflect.Bool:
// 		_, err := strconv.ParseBool(value)
// 		return err
// 	case reflect.Int, reflect.Int32, reflect.Int64:
// 		_, err := strconv.Atoi(value)
// 		return err
// 	case reflect.String:
// 		return nil
// 	default:
// 		return glideErrors.NewConfigError(fmt.Sprintf("unsupported type: %v", v.Kind()),
// 			glideErrors.WithSuggestions(
// 				"This configuration field type is not supported for validation",
// 				"Report this as a bug if you encounter this error",
// 			))
// 	}
// }
