package plugin

import (
	"context"
	"testing"

	"github.com/ivannovak/glide/v2/pkg/plugin/sdk"
	v1 "github.com/ivannovak/glide/v2/pkg/plugin/sdk/v1"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

// MockGlidePlugin is a mock implementation of v1.GlidePluginClient
type MockGlidePlugin struct {
	mock.Mock
}

func (m *MockGlidePlugin) GetMetadata(ctx context.Context, in *v1.Empty, opts ...grpc.CallOption) (*v1.PluginMetadata, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.PluginMetadata), args.Error(1)
}

func (m *MockGlidePlugin) Configure(ctx context.Context, in *v1.ConfigureRequest, opts ...grpc.CallOption) (*v1.ConfigureResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.ConfigureResponse), args.Error(1)
}

func (m *MockGlidePlugin) ListCommands(ctx context.Context, in *v1.Empty, opts ...grpc.CallOption) (*v1.CommandList, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.CommandList), args.Error(1)
}

func (m *MockGlidePlugin) ExecuteCommand(ctx context.Context, in *v1.ExecuteRequest, opts ...grpc.CallOption) (*v1.ExecuteResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.ExecuteResponse), args.Error(1)
}

func (m *MockGlidePlugin) StartInteractive(ctx context.Context, opts ...grpc.CallOption) (v1.GlidePlugin_StartInteractiveClient, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(v1.GlidePlugin_StartInteractiveClient), args.Error(1)
}

func (m *MockGlidePlugin) GetCapabilities(ctx context.Context, in *v1.Empty, opts ...grpc.CallOption) (*v1.Capabilities, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.Capabilities), args.Error(1)
}

func (m *MockGlidePlugin) GetCustomCategories(ctx context.Context, in *v1.Empty, opts ...grpc.CallOption) (*v1.CategoryList, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1.CategoryList), args.Error(1)
}

func TestCreatePluginCommand_VisibilityAnnotation(t *testing.T) {
	tests := []struct {
		name               string
		cmdInfo            *v1.CommandInfo
		expectedVisibility string
	}{
		{
			name: "command with always visibility",
			cmdInfo: &v1.CommandInfo{
				Name:        "test-always",
				Description: "Test command with always visibility",
				Visibility:  v1.VisibilityAlways,
			},
			expectedVisibility: v1.VisibilityAlways,
		},
		{
			name: "command with project-only visibility",
			cmdInfo: &v1.CommandInfo{
				Name:        "test-project",
				Description: "Test command with project-only visibility",
				Visibility:  v1.VisibilityProjectOnly,
			},
			expectedVisibility: v1.VisibilityProjectOnly,
		},
		{
			name: "command with worktree-only visibility",
			cmdInfo: &v1.CommandInfo{
				Name:        "test-worktree",
				Description: "Test command with worktree-only visibility",
				Visibility:  v1.VisibilityWorktreeOnly,
			},
			expectedVisibility: v1.VisibilityWorktreeOnly,
		},
		{
			name: "command with root-only visibility",
			cmdInfo: &v1.CommandInfo{
				Name:        "test-root",
				Description: "Test command with root-only visibility",
				Visibility:  v1.VisibilityRootOnly,
			},
			expectedVisibility: v1.VisibilityRootOnly,
		},
		{
			name: "command with non-root visibility",
			cmdInfo: &v1.CommandInfo{
				Name:        "test-non-root",
				Description: "Test command with non-root visibility",
				Visibility:  v1.VisibilityNonRoot,
			},
			expectedVisibility: v1.VisibilityNonRoot,
		},
		{
			name: "command without visibility defaults to always",
			cmdInfo: &v1.CommandInfo{
				Name:        "test-default",
				Description: "Test command without visibility",
				Visibility:  "", // Empty visibility
			},
			expectedVisibility: v1.VisibilityAlways,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create runtime integration
			r := NewRuntimePluginIntegration()

			// Create mock plugin
			mockPlugin := new(MockGlidePlugin)
			plugin := &sdk.LoadedPlugin{
				Name: "test-plugin",
				Metadata: &v1.PluginMetadata{
					Name:        "test-plugin",
					Description: "Test plugin",
				},
				Plugin: mockPlugin,
			}

			// Create command
			cmd := r.createPluginCommand(plugin, mockPlugin, tt.cmdInfo)

			// Check visibility annotation
			assert.NotNil(t, cmd.Annotations, "Command should have annotations")
			assert.Equal(t, tt.expectedVisibility, cmd.Annotations["visibility"],
				"Visibility annotation should be %s", tt.expectedVisibility)
		})
	}
}

func TestCreatePluginCommand_AllAnnotations(t *testing.T) {
	r := NewRuntimePluginIntegration()
	mockPlugin := new(MockGlidePlugin)
	plugin := &sdk.LoadedPlugin{
		Name: "test-plugin",
		Metadata: &v1.PluginMetadata{
			Name:        "test-plugin",
			Description: "Test plugin",
		},
		Plugin: mockPlugin,
	}

	// Test command with all fields set
	cmdInfo := &v1.CommandInfo{
		Name:        "test-cmd",
		Description: "Test command",
		Category:    "testing",
		Visibility:  v1.VisibilityProjectOnly,
		Aliases:     []string{"tc", "test"},
		Hidden:      true,
	}

	cmd := r.createPluginCommand(plugin, mockPlugin, cmdInfo)

	// Check all annotations
	assert.Equal(t, "test-plugin", cmd.Annotations["plugin"], "Should have plugin annotation")
	assert.Equal(t, "testing", cmd.Annotations["category"], "Should have category annotation")
	assert.Equal(t, v1.VisibilityProjectOnly, cmd.Annotations["visibility"], "Should have visibility annotation")

	// Check other properties
	assert.Equal(t, "test-cmd", cmd.Name(), "Command name should match")
	assert.Equal(t, "Test command", cmd.Short, "Command description should match")
	assert.Equal(t, []string{"tc", "test"}, cmd.Aliases, "Command aliases should match")
	assert.True(t, cmd.Hidden, "Command should be hidden")
}

func TestAddPluginCommands_GlobalRegistration(t *testing.T) {
	r := NewRuntimePluginIntegration()
	rootCmd := &cobra.Command{Use: "root"}

	// Create mock plugin
	mockPlugin := new(MockGlidePlugin)
	plugin := &sdk.LoadedPlugin{
		Name: "test-plugin",
		Metadata: &v1.PluginMetadata{
			Name:        "test-plugin",
			Description: "Test plugin",
			Namespaced:  false, // Global registration
		},
		Plugin: mockPlugin,
	}

	// Setup mock to return commands with different visibilities
	commandList := &v1.CommandList{
		Commands: []*v1.CommandInfo{
			{
				Name:        "global-cmd1",
				Description: "Global command 1",
				Visibility:  v1.VisibilityAlways,
			},
			{
				Name:        "global-cmd2",
				Description: "Global command 2",
				Visibility:  v1.VisibilityProjectOnly,
			},
		},
	}
	mockPlugin.On("ListCommands", mock.Anything, mock.Anything).Return(commandList, nil)
	mockPlugin.On("GetCustomCategories", mock.Anything, mock.Anything).Return(&v1.CategoryList{}, nil)

	// Add plugin commands
	err := r.addPluginCommands(rootCmd, plugin)
	assert.NoError(t, err)

	// Check that commands were added with correct annotations
	for _, cmd := range rootCmd.Commands() {
		assert.NotNil(t, cmd.Annotations, "Command should have annotations")
		assert.Equal(t, "test-plugin", cmd.Annotations["plugin"], "Should have plugin annotation")
		assert.Equal(t, "true", cmd.Annotations["global_plugin"], "Should be marked as global plugin")
		assert.NotEmpty(t, cmd.Annotations["visibility"], "Should have visibility annotation")
	}

	// Verify specific commands
	cmd1 := findCommand(rootCmd, "global-cmd1")
	assert.NotNil(t, cmd1, "Should find global-cmd1")
	assert.Equal(t, v1.VisibilityAlways, cmd1.Annotations["visibility"])

	cmd2 := findCommand(rootCmd, "global-cmd2")
	assert.NotNil(t, cmd2, "Should find global-cmd2")
	assert.Equal(t, v1.VisibilityProjectOnly, cmd2.Annotations["visibility"])
}

func TestAddPluginCommands_NamespacedRegistration(t *testing.T) {
	r := NewRuntimePluginIntegration()
	rootCmd := &cobra.Command{Use: "root"}

	// Create mock plugin
	mockPlugin := new(MockGlidePlugin)
	plugin := &sdk.LoadedPlugin{
		Name: "test-plugin",
		Metadata: &v1.PluginMetadata{
			Name:        "test-plugin",
			Description: "Test plugin",
			Namespaced:  true, // Namespaced registration
		},
		Plugin: mockPlugin,
	}

	// Setup mock to return commands
	commandList := &v1.CommandList{
		Commands: []*v1.CommandInfo{
			{
				Name:        "sub-cmd1",
				Description: "Sub command 1",
				Visibility:  v1.VisibilityWorktreeOnly,
			},
			{
				Name:        "sub-cmd2",
				Description: "Sub command 2",
				Visibility:  v1.VisibilityRootOnly,
			},
		},
	}
	mockPlugin.On("ListCommands", mock.Anything, mock.Anything).Return(commandList, nil)
	mockPlugin.On("GetCustomCategories", mock.Anything, mock.Anything).Return(&v1.CategoryList{}, nil)

	// Add plugin commands
	err := r.addPluginCommands(rootCmd, plugin)
	assert.NoError(t, err)

	// Should create a parent command for the plugin
	pluginCmd := findCommand(rootCmd, "test-plugin")
	assert.NotNil(t, pluginCmd, "Should create parent command for plugin")
	assert.Equal(t, "plugin", pluginCmd.Annotations["category"])

	// Check sub-commands
	subCmd1 := findCommand(pluginCmd, "sub-cmd1")
	assert.NotNil(t, subCmd1, "Should find sub-cmd1")
	assert.Equal(t, v1.VisibilityWorktreeOnly, subCmd1.Annotations["visibility"])

	subCmd2 := findCommand(pluginCmd, "sub-cmd2")
	assert.NotNil(t, subCmd2, "Should find sub-cmd2")
	assert.Equal(t, v1.VisibilityRootOnly, subCmd2.Annotations["visibility"])
}

func TestAddPluginCommands_ConflictHandling(t *testing.T) {
	r := NewRuntimePluginIntegration()
	rootCmd := &cobra.Command{Use: "root"}

	// Add an existing command
	existingCmd := &cobra.Command{
		Use:   "existing",
		Short: "Existing command",
	}
	rootCmd.AddCommand(existingCmd)

	// Create mock plugin with conflicting command
	mockPlugin := new(MockGlidePlugin)
	plugin := &sdk.LoadedPlugin{
		Name: "test-plugin",
		Metadata: &v1.PluginMetadata{
			Name:        "test-plugin",
			Description: "Test plugin",
			Namespaced:  false, // Global registration to test conflict
		},
		Plugin: mockPlugin,
	}

	// Setup mock to return conflicting command
	commandList := &v1.CommandList{
		Commands: []*v1.CommandInfo{
			{
				Name:        "existing", // Conflicts with existing command
				Description: "Conflicting command",
				Visibility:  v1.VisibilityAlways,
			},
			{
				Name:        "new-cmd", // Should be added successfully
				Description: "New command",
				Visibility:  v1.VisibilityAlways,
			},
		},
	}
	mockPlugin.On("ListCommands", mock.Anything, mock.Anything).Return(commandList, nil)
	mockPlugin.On("GetCustomCategories", mock.Anything, mock.Anything).Return(&v1.CategoryList{}, nil)

	// Add plugin commands
	err := r.addPluginCommands(rootCmd, plugin)
	assert.NoError(t, err, "Should not error even with conflicts")

	// Check that existing command was not replaced
	existingCheck := findCommand(rootCmd, "existing")
	assert.Equal(t, "Existing command", existingCheck.Short, "Existing command should not be replaced")

	// Check that new command was added
	newCmd := findCommand(rootCmd, "new-cmd")
	assert.NotNil(t, newCmd, "New command should be added")
	assert.Equal(t, "New command", newCmd.Short)
}

// Helper function to find a command by name
func findCommand(parent *cobra.Command, name string) *cobra.Command {
	for _, cmd := range parent.Commands() {
		if cmd.Name() == name {
			return cmd
		}
	}
	return nil
}

func TestRuntimePluginCustomCategories(t *testing.T) {
	r := NewRuntimePluginIntegration()
	rootCmd := &cobra.Command{Use: "root"}

	// Create mock plugin with custom categories
	mockPlugin := new(MockGlidePlugin)
	plugin := &sdk.LoadedPlugin{
		Name: "test-plugin",
		Metadata: &v1.PluginMetadata{
			Name:        "test-plugin",
			Description: "Test plugin",
		},
		Plugin: mockPlugin,
	}

	// Setup mock to return custom categories
	customCategories := &v1.CategoryList{
		Categories: []*v1.CustomCategory{
			{
				Id:          "infrastructure",
				Name:        "Infrastructure",
				Description: "Infrastructure management",
				Priority:    45,
			},
			{
				Id:          "monitoring",
				Name:        "Monitoring",
				Description: "Monitoring and observability",
				Priority:    55,
			},
		},
	}
	mockPlugin.On("GetCustomCategories", mock.Anything, mock.Anything).Return(customCategories, nil)
	mockPlugin.On("ListCommands", mock.Anything, mock.Anything).Return(&v1.CommandList{}, nil)

	// Add plugin commands
	err := r.addPluginCommands(rootCmd, plugin)
	assert.NoError(t, err)

	// Check that custom categories were registered
	assert.Len(t, r.customCategories, 2, "Should have 2 custom categories")
	assert.Equal(t, "infrastructure", r.customCategories[0].Id)
	assert.Equal(t, "monitoring", r.customCategories[1].Id)

	// Check global categories
	globalCats := GetGlobalPluginCategories()
	assert.Contains(t, globalCats, customCategories.Categories[0])
	assert.Contains(t, globalCats, customCategories.Categories[1])
}
