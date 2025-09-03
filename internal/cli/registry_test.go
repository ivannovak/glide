package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	
	assert.NotNil(t, registry)
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()
	
	factory := func() *cobra.Command {
		return &cobra.Command{
			Use: "test",
		}
	}
	
	metadata := Metadata{
		Name:        "test",
		Category:    CategoryCore,
		Description: "Test command",
	}
	
	err := registry.Register("test", factory, metadata)
	assert.NoError(t, err)
	
	// Test duplicate registration fails
	err = registry.Register("test", factory, metadata)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()
	
	testFactory := func() *cobra.Command {
		return &cobra.Command{Use: "test"}
	}
	
	factory, exists := registry.Get("nonexistent")
	assert.Nil(t, factory)
	assert.False(t, exists)
	
	registry.Register("test", testFactory, Metadata{Name: "test"})
	
	factory, exists = registry.Get("test")
	assert.NotNil(t, factory)
	assert.True(t, exists)
}

func TestRegistry_GetMetadata(t *testing.T) {
	registry := NewRegistry()
	
	testFactory := func() *cobra.Command {
		return &cobra.Command{Use: "test"}
	}
	
	metadata, exists := registry.GetMetadata("nonexistent")
	assert.Equal(t, Metadata{}, metadata)
	assert.False(t, exists)
	
	registry.Register("test", testFactory, Metadata{
		Name:        "test",
		Category:    CategoryCore,
		Description: "Test command",
	})
	
	metadata, exists = registry.GetMetadata("test")
	assert.Equal(t, "test", metadata.Name)
	assert.Equal(t, CategoryCore, metadata.Category)
	assert.Equal(t, "Test command", metadata.Description)
	assert.True(t, exists)
}

func TestRegistry_GetByCategory(t *testing.T) {
	registry := NewRegistry()
	
	factory := func() *cobra.Command {
		return &cobra.Command{Use: "test"}
	}
	
	registry.Register("core1", factory, Metadata{Category: CategoryCore})
	registry.Register("core2", factory, Metadata{Category: CategoryCore})
	registry.Register("docker1", factory, Metadata{Category: CategoryDocker})
	
	coreCommands := registry.GetByCategory(CategoryCore)
	assert.Len(t, coreCommands, 2)
	assert.Contains(t, coreCommands, "core1")
	assert.Contains(t, coreCommands, "core2")
	
	dockerCommands := registry.GetByCategory(CategoryDocker)
	assert.Len(t, dockerCommands, 1)
	assert.Contains(t, dockerCommands, "docker1")
}

func TestRegistry_CreateAll(t *testing.T) {
	registry := NewRegistry()
	
	factory1 := func() *cobra.Command {
		return &cobra.Command{Use: "test1"}
	}
	
	factory2 := func() *cobra.Command {
		return &cobra.Command{Use: "test2"}
	}
	
	registry.Register("test1", factory1, Metadata{Name: "test1"})
	registry.Register("test2", factory2, Metadata{Name: "test2"})
	
	commands := registry.CreateAll()
	assert.Len(t, commands, 2)
	
	uses := []string{commands[0].Use, commands[1].Use}
	assert.Contains(t, uses, "test1")
	assert.Contains(t, uses, "test2")
}

func TestRegistry_CreateByCategory(t *testing.T) {
	registry := NewRegistry()
	
	coreFactory := func() *cobra.Command {
		return &cobra.Command{Use: "core"}
	}
	
	dockerFactory := func() *cobra.Command {
		return &cobra.Command{Use: "docker"}
	}
	
	registry.Register("core", coreFactory, Metadata{Category: CategoryCore})
	registry.Register("docker", dockerFactory, Metadata{Category: CategoryDocker})
	
	coreCommands := registry.CreateByCategory(CategoryCore)
	assert.Len(t, coreCommands, 1)
	assert.Equal(t, "core", coreCommands[0].Use)
	
	dockerCommands := registry.CreateByCategory(CategoryDocker)
	assert.Len(t, dockerCommands, 1)
	assert.Equal(t, "docker", dockerCommands[0].Use)
}

func TestRegistry_Categories(t *testing.T) {
	// Test that all categories are defined correctly
	assert.Equal(t, Category("core"), CategoryCore)
	assert.Equal(t, Category("docker"), CategoryDocker)
	assert.Equal(t, Category("database"), CategoryDatabase)
	assert.Equal(t, Category("developer"), CategoryDeveloper)
	assert.Equal(t, Category("debug"), CategoryDebug)
	assert.Equal(t, Category("help"), CategoryHelp)
}