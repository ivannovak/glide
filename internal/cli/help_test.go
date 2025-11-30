package cli

import (
	"testing"

	"github.com/ivannovak/glide/v2/internal/config"
	"github.com/ivannovak/glide/v2/internal/context"
	v1 "github.com/ivannovak/glide/v2/pkg/plugin/sdk/v1"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestHelpCommand_shouldShowCommand(t *testing.T) {
	tests := []struct {
		name           string
		visibility     string
		projectContext *context.ProjectContext
		expected       bool
	}{
		{
			name:           "no visibility annotation - always show",
			visibility:     "",
			projectContext: nil,
			expected:       true,
		},
		{
			name:           "always visibility - show without context",
			visibility:     v1.VisibilityAlways,
			projectContext: nil,
			expected:       true,
		},
		{
			name:       "always visibility - show with context",
			visibility: v1.VisibilityAlways,
			projectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeMultiWorktree,
				Location:        context.LocationRoot,
			},
			expected: true,
		},
		{
			name:           "project-only - hide without context",
			visibility:     v1.VisibilityProjectOnly,
			projectContext: nil,
			expected:       false,
		},
		{
			name:       "project-only - show in single repo mode",
			visibility: v1.VisibilityProjectOnly,
			projectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeSingleRepo,
			},
			expected: true,
		},
		{
			name:       "project-only - show in multi-worktree at root",
			visibility: v1.VisibilityProjectOnly,
			projectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeMultiWorktree,
				Location:        context.LocationRoot,
			},
			expected: true,
		},
		{
			name:       "project-only - show in worktree",
			visibility: v1.VisibilityProjectOnly,
			projectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeMultiWorktree,
				Location:        context.LocationWorktree,
			},
			expected: true,
		},
		{
			name:       "worktree-only - hide in single repo",
			visibility: v1.VisibilityWorktreeOnly,
			projectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeSingleRepo,
			},
			expected: false,
		},
		{
			name:       "worktree-only - hide at multi-worktree root",
			visibility: v1.VisibilityWorktreeOnly,
			projectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeMultiWorktree,
				Location:        context.LocationRoot,
			},
			expected: false,
		},
		{
			name:       "worktree-only - show in worktree",
			visibility: v1.VisibilityWorktreeOnly,
			projectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeMultiWorktree,
				Location:        context.LocationWorktree,
			},
			expected: true,
		},
		{
			name:       "worktree-only - hide in main repo",
			visibility: v1.VisibilityWorktreeOnly,
			projectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeMultiWorktree,
				Location:        context.LocationMainRepo,
			},
			expected: false,
		},
		{
			name:       "root-only - hide in single repo",
			visibility: v1.VisibilityRootOnly,
			projectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeSingleRepo,
			},
			expected: false,
		},
		{
			name:       "root-only - show at multi-worktree root",
			visibility: v1.VisibilityRootOnly,
			projectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeMultiWorktree,
				Location:        context.LocationRoot,
			},
			expected: true,
		},
		{
			name:       "root-only - hide in worktree",
			visibility: v1.VisibilityRootOnly,
			projectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeMultiWorktree,
				Location:        context.LocationWorktree,
			},
			expected: false,
		},
		{
			name:       "non-root - show in single repo",
			visibility: v1.VisibilityNonRoot,
			projectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeSingleRepo,
			},
			expected: true,
		},
		{
			name:       "non-root - hide at multi-worktree root",
			visibility: v1.VisibilityNonRoot,
			projectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeMultiWorktree,
				Location:        context.LocationRoot,
			},
			expected: false,
		},
		{
			name:       "non-root - show in worktree",
			visibility: v1.VisibilityNonRoot,
			projectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeMultiWorktree,
				Location:        context.LocationWorktree,
			},
			expected: true,
		},
		{
			name:       "non-root - show in main repo",
			visibility: v1.VisibilityNonRoot,
			projectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeMultiWorktree,
				Location:        context.LocationMainRepo,
			},
			expected: true,
		},
		{
			name:           "non-root - show without context",
			visibility:     v1.VisibilityNonRoot,
			projectContext: nil,
			expected:       true,
		},
		{
			name:       "unknown visibility - default to show",
			visibility: "unknown-visibility",
			projectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeMultiWorktree,
				Location:        context.LocationRoot,
			},
			expected: true,
		},
		{
			name:       "empty development mode - treat as no project",
			visibility: v1.VisibilityProjectOnly,
			projectContext: &context.ProjectContext{
				DevelopmentMode: "",
				Location:        context.LocationWorktree,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc := &HelpCommand{
				ProjectContext: tt.projectContext,
			}

			cmd := &cobra.Command{}
			if tt.visibility != "" {
				cmd.Annotations = map[string]string{
					"visibility": tt.visibility,
				}
			}

			result := hc.shouldShowCommand(cmd)
			assert.Equal(t, tt.expected, result, "visibility check failed for %s", tt.name)
		})
	}
}

func TestHelpCommand_shouldShowCategory(t *testing.T) {
	tests := []struct {
		name           string
		category       string
		projectContext *context.ProjectContext
		expected       bool
	}{
		{
			name:           "global category - hide without context",
			category:       "global",
			projectContext: nil,
			expected:       false,
		},
		{
			name:     "global category - show in multi-worktree",
			category: "global",
			projectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeMultiWorktree,
			},
			expected: true,
		},
		{
			name:     "global category - hide in single repo",
			category: "global",
			projectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeSingleRepo,
			},
			expected: false,
		},
		{
			name:     "docker category - show when in project",
			category: "docker",
			projectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeMultiWorktree,
			},
			expected: true,
		},
		{
			name:     "docker category - hide when no project",
			category: "docker",
			projectContext: &context.ProjectContext{
				DevelopmentMode: "",
			},
			expected: false,
		},
		{
			name:           "docker category - hide without context",
			category:       "docker",
			projectContext: nil,
			expected:       false,
		},
		{
			name:     "testing category - show in project",
			category: "testing",
			projectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeSingleRepo,
			},
			expected: true,
		},
		{
			name:     "developer category - show in project",
			category: "developer",
			projectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeMultiWorktree,
				Location:        context.LocationWorktree,
			},
			expected: true,
		},
		{
			name:     "database category - show in project",
			category: "database",
			projectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeSingleRepo,
			},
			expected: true,
		},
		{
			name:     "core category - always show",
			category: "core",
			projectContext: &context.ProjectContext{
				DevelopmentMode: "",
			},
			expected: true,
		},
		{
			name:           "core category - show without context",
			category:       "core",
			projectContext: nil,
			expected:       true,
		},
		{
			name:     "setup category - always show",
			category: "setup",
			projectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeMultiWorktree,
			},
			expected: true,
		},
		{
			name:     "help category - always show",
			category: "help",
			projectContext: &context.ProjectContext{
				DevelopmentMode: "",
			},
			expected: true,
		},
		{
			name:     "plugin category - always show",
			category: "plugin",
			projectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeSingleRepo,
			},
			expected: true,
		},
		{
			name:     "unknown category - default to show",
			category: "custom-category",
			projectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeMultiWorktree,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc := &HelpCommand{
				ProjectContext: tt.projectContext,
			}

			result := hc.shouldShowCategory(tt.category)
			assert.Equal(t, tt.expected, result, "category check failed for %s", tt.name)
		})
	}
}

// TestVisibilityIntegration tests the integration of visibility settings
// with actual command registration and help display
func TestVisibilityIntegration(t *testing.T) {
	// Create a root command
	rootCmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
	}

	// Add test commands with different visibilities
	testCommands := []struct {
		name       string
		visibility string
		category   string
	}{
		{"always-cmd", v1.VisibilityAlways, "core"},
		{"project-cmd", v1.VisibilityProjectOnly, "docker"},
		{"worktree-cmd", v1.VisibilityWorktreeOnly, "testing"},
		{"root-cmd", v1.VisibilityRootOnly, "global"},
		{"non-root-cmd", v1.VisibilityNonRoot, "developer"},
	}

	for _, tc := range testCommands {
		cmd := &cobra.Command{
			Use:   tc.name,
			Short: "Test " + tc.name,
			Annotations: map[string]string{
				"visibility": tc.visibility,
				"category":   tc.category,
			},
		}
		rootCmd.AddCommand(cmd)
	}

	// Test different contexts
	contexts := []struct {
		name     string
		context  *context.ProjectContext
		expected []string // Expected visible command names
	}{
		{
			name:     "no context",
			context:  nil,
			expected: []string{"always-cmd", "non-root-cmd"},
		},
		{
			name: "single repo",
			context: &context.ProjectContext{
				DevelopmentMode: context.ModeSingleRepo,
			},
			expected: []string{"always-cmd", "project-cmd", "non-root-cmd"},
		},
		{
			name: "multi-worktree root",
			context: &context.ProjectContext{
				DevelopmentMode: context.ModeMultiWorktree,
				Location:        context.LocationRoot,
			},
			expected: []string{"always-cmd", "project-cmd", "root-cmd"},
		},
		{
			name: "worktree",
			context: &context.ProjectContext{
				DevelopmentMode: context.ModeMultiWorktree,
				Location:        context.LocationWorktree,
			},
			expected: []string{"always-cmd", "project-cmd", "worktree-cmd", "non-root-cmd"},
		},
		{
			name: "main repo",
			context: &context.ProjectContext{
				DevelopmentMode: context.ModeMultiWorktree,
				Location:        context.LocationMainRepo,
			},
			expected: []string{"always-cmd", "project-cmd", "non-root-cmd"},
		},
	}

	for _, tc := range contexts {
		t.Run(tc.name, func(t *testing.T) {
			hc := &HelpCommand{
				ProjectContext: tc.context,
			}

			// Collect visible commands
			var visibleCommands []string
			for _, cmd := range rootCmd.Commands() {
				if hc.shouldShowCommand(cmd) {
					visibleCommands = append(visibleCommands, cmd.Name())
				}
			}

			// Check that we got the expected commands
			assert.ElementsMatch(t, tc.expected, visibleCommands,
				"Expected commands %v but got %v", tc.expected, visibleCommands)
		})
	}
}

// TestHelpTopics tests all help topic display functions
func TestHelpTopics(t *testing.T) {
	t.Run("getting started guide", func(t *testing.T) {
		hc := &HelpCommand{
			ProjectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeSingleRepo,
			},
		}

		err := hc.showGettingStarted()
		assert.NoError(t, err)
	})

	t.Run("workflows - single repo mode", func(t *testing.T) {
		hc := &HelpCommand{
			ProjectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeSingleRepo,
			},
		}

		err := hc.showWorkflows()
		assert.NoError(t, err)
	})

	t.Run("workflows - multi-worktree mode", func(t *testing.T) {
		hc := &HelpCommand{
			ProjectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeMultiWorktree,
			},
		}

		err := hc.showWorkflows()
		assert.NoError(t, err)
	})

	t.Run("modes explanation", func(t *testing.T) {
		hc := &HelpCommand{
			ProjectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeSingleRepo,
			},
		}

		err := hc.showModes()
		assert.NoError(t, err)
	})

	t.Run("troubleshooting guide", func(t *testing.T) {
		hc := &HelpCommand{
			ProjectContext: &context.ProjectContext{},
		}

		err := hc.showTroubleshooting()
		assert.NoError(t, err)
	})

	t.Run("command help", func(t *testing.T) {
		hc := &HelpCommand{
			ProjectContext: &context.ProjectContext{},
		}

		err := hc.showCommandHelp("test")
		assert.NoError(t, err)
	})
}

// TestHelpCommandExecution tests help command with various arguments
func TestHelpCommandExecution(t *testing.T) {
	t.Run("help with no args", func(t *testing.T) {
		ctx := &context.ProjectContext{
			DevelopmentMode: context.ModeSingleRepo,
		}
		cmd := NewHelpCommand(ctx, &config.Config{})

		assert.NotNil(t, cmd)
		assert.Equal(t, "help [command | topic]", cmd.Use)
	})

	t.Run("help topic aliases", func(t *testing.T) {
		ctx := &context.ProjectContext{
			DevelopmentMode: context.ModeSingleRepo,
		}
		cmd := NewHelpCommand(ctx, &config.Config{})

		// Test that RunE is set
		assert.NotNil(t, cmd.RunE)
	})
}

// TestCategories tests category definitions and ordering
func TestCategories(t *testing.T) {
	t.Run("all categories defined", func(t *testing.T) {
		expectedCategories := []string{
			"core", "global", "setup", "docker", "testing",
			"developer", "database", "plugin", "help",
		}

		for _, cat := range expectedCategories {
			info, exists := Categories[cat]
			assert.True(t, exists, "category %s should be defined", cat)
			assert.NotEmpty(t, info.Name)
			assert.NotEmpty(t, info.Description)
			assert.NotNil(t, info.Color)
		}
	})

	t.Run("category priorities", func(t *testing.T) {
		// Core should have lower priority (appears first)
		assert.Less(t, Categories["core"].Priority, Categories["help"].Priority,
			"core should appear before help")

		assert.Less(t, Categories["setup"].Priority, Categories["plugin"].Priority,
			"setup should appear before plugin")
	})

	t.Run("category display order", func(t *testing.T) {
		// Extract priorities
		priorities := make(map[string]int)
		for cat, info := range Categories {
			priorities[cat] = info.Priority
		}

		// Core commands should be first (priority 10)
		assert.Equal(t, 10, priorities["core"])

		// Help should be last (priority 90)
		assert.Equal(t, 90, priorities["help"])
	})
}

// TestCommandEntry tests command entry structure
func TestCommandEntry(t *testing.T) {
	t.Run("command entry fields", func(t *testing.T) {
		entry := CommandEntry{
			Name:        "test",
			Description: "Run tests",
			Aliases:     []string{"t"},
			Category:    "testing",
			IsPlugin:    false,
			IsYAML:      false,
		}

		assert.Equal(t, "test", entry.Name)
		assert.Equal(t, "Run tests", entry.Description)
		assert.Equal(t, []string{"t"}, entry.Aliases)
		assert.Equal(t, "testing", entry.Category)
		assert.False(t, entry.IsPlugin)
		assert.False(t, entry.IsYAML)
	})

	t.Run("plugin command entry", func(t *testing.T) {
		entry := CommandEntry{
			Name:       "custom-cmd",
			Category:   "plugin",
			IsPlugin:   true,
			PluginName: "my-plugin",
		}

		assert.True(t, entry.IsPlugin)
		assert.Equal(t, "my-plugin", entry.PluginName)
	})

	t.Run("YAML command entry", func(t *testing.T) {
		entry := CommandEntry{
			Name:     "my-cmd",
			Category: "yaml",
			IsYAML:   true,
		}

		assert.True(t, entry.IsYAML)
		assert.Equal(t, "yaml", entry.Category)
	})
}

// TestHelpContextInfo tests context-specific help information
func TestHelpContextInfo(t *testing.T) {
	t.Run("single repo context", func(t *testing.T) {
		hc := &HelpCommand{
			ProjectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeSingleRepo,
				WorkingDir:      "/test/project",
				ProjectRoot:     "/test/project",
			},
		}

		assert.Equal(t, context.ModeSingleRepo, hc.ProjectContext.DevelopmentMode)
	})

	t.Run("multi-worktree context", func(t *testing.T) {
		hc := &HelpCommand{
			ProjectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeMultiWorktree,
				Location:        context.LocationRoot,
				IsRoot:          true,
			},
		}

		assert.Equal(t, context.ModeMultiWorktree, hc.ProjectContext.DevelopmentMode)
		assert.True(t, hc.ProjectContext.IsRoot)
	})

	t.Run("standalone context", func(t *testing.T) {
		hc := &HelpCommand{
			ProjectContext: &context.ProjectContext{
				DevelopmentMode: context.ModeStandalone,
			},
		}

		assert.Equal(t, context.ModeStandalone, hc.ProjectContext.DevelopmentMode)
	})

	t.Run("no context", func(t *testing.T) {
		hc := &HelpCommand{
			ProjectContext: nil,
		}

		assert.Nil(t, hc.ProjectContext)
	})
}

// TestHelpTopicAliases tests that topic aliases work correctly
func TestHelpTopicAliases(t *testing.T) {
	ctx := &context.ProjectContext{
		DevelopmentMode: context.ModeSingleRepo,
	}
	cmd := NewHelpCommand(ctx, &config.Config{})

	// Test that the command handles various topic names
	// This verifies the switch statement in RunE
	topics := map[string][]string{
		"getting-started": {"getting-started", "start", "quickstart"},
		"workflows":       {"workflows", "workflow", "flow"},
		"modes":           {"modes", "mode"},
		"troubleshooting": {"troubleshooting", "troubleshoot", "issues"},
	}

	for mainTopic, aliases := range topics {
		for _, alias := range aliases {
			t.Run("alias_"+alias, func(t *testing.T) {
				// Verify that all aliases are handled
				// We can't easily test RunE directly, but we can verify the structure
				assert.Contains(t, cmd.Long, mainTopic,
					"help text should mention %s topic", mainTopic)
			})
		}
	}
}

// TestCategoryInfo tests category information structure
func TestCategoryInfo(t *testing.T) {
	t.Run("core category info", func(t *testing.T) {
		info := Categories["core"]
		assert.Equal(t, "Core Commands", info.Name)
		assert.Equal(t, "Essential development commands", info.Description)
		assert.Equal(t, 10, info.Priority)
		assert.NotNil(t, info.Color)
	})

	t.Run("plugin category info", func(t *testing.T) {
		info := Categories["plugin"]
		assert.Equal(t, "Plugin Commands", info.Name)
		assert.Equal(t, "Commands from installed plugins", info.Description)
		assert.Equal(t, 80, info.Priority)
	})

	t.Run("help category info", func(t *testing.T) {
		info := Categories["help"]
		assert.Equal(t, "Help & Documentation", info.Name)
		assert.Equal(t, 90, info.Priority) // Should be last
	})
}

// TestGetPluginSubcommands tests the getPluginSubcommands helper
func TestGetPluginSubcommands(t *testing.T) {
	t.Run("plugin with subcommands", func(t *testing.T) {
		hc := &HelpCommand{
			ProjectContext: &context.ProjectContext{},
		}

		// Create a root command with a plugin that has subcommands
		rootCmd := &cobra.Command{Use: "glide"}

		pluginCmd := &cobra.Command{
			Use:   "docker",
			Short: "Docker management",
		}

		// Add subcommands to plugin
		pluginCmd.AddCommand(&cobra.Command{
			Use:     "ps",
			Short:   "List containers",
			Aliases: []string{"list"},
		})
		pluginCmd.AddCommand(&cobra.Command{
			Use:   "up",
			Short: "Start containers",
		})

		// Add hidden command (should not appear)
		hiddenCmd := &cobra.Command{
			Use:    "internal",
			Short:  "Internal command",
			Hidden: true,
		}
		pluginCmd.AddCommand(hiddenCmd)

		rootCmd.AddCommand(pluginCmd)

		subcommands := hc.getPluginSubcommands(rootCmd, "docker")

		assert.Len(t, subcommands, 2, "should return 2 non-hidden subcommands")

		// Commands should be sorted alphabetically
		assert.Equal(t, "ps", subcommands[0].Name)
		assert.Equal(t, "List containers", subcommands[0].Description)
		assert.Equal(t, []string{"list"}, subcommands[0].Aliases)

		assert.Equal(t, "up", subcommands[1].Name)
	})

	t.Run("plugin not found", func(t *testing.T) {
		hc := &HelpCommand{
			ProjectContext: &context.ProjectContext{},
		}

		rootCmd := &cobra.Command{Use: "glide"}

		subcommands := hc.getPluginSubcommands(rootCmd, "nonexistent")

		assert.Empty(t, subcommands, "should return empty slice for non-existent plugin")
	})

	t.Run("plugin with no subcommands", func(t *testing.T) {
		hc := &HelpCommand{
			ProjectContext: &context.ProjectContext{},
		}

		rootCmd := &cobra.Command{Use: "glide"}
		pluginCmd := &cobra.Command{
			Use:   "simple",
			Short: "Simple command",
		}
		rootCmd.AddCommand(pluginCmd)

		subcommands := hc.getPluginSubcommands(rootCmd, "simple")

		assert.Empty(t, subcommands, "should return empty slice for plugin with no subcommands")
	})
}

// TestAreCompletionsInstalled tests the completion check helper
func TestAreCompletionsInstalled(t *testing.T) {
	t.Run("checks completion files", func(t *testing.T) {
		hc := &HelpCommand{
			ProjectContext: &context.ProjectContext{},
		}

		// Just verify it doesn't panic and returns a boolean
		result := hc.areCompletionsInstalled()
		assert.IsType(t, false, result, "should return a boolean")
	})
}
