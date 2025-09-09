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

// Test alias registration and resolution
func TestRegistry_RegisterWithAliases(t *testing.T) {
	registry := NewRegistry()

	factory := func() *cobra.Command {
		return &cobra.Command{
			Use:   "artisan",
			Short: "Run Artisan commands",
		}
	}

	metadata := Metadata{
		Name:        "artisan",
		Category:    CategoryDeveloper,
		Description: "Run Artisan commands via Docker",
		Aliases:     []string{"a", "art"},
	}

	// Register command with aliases
	err := registry.Register("artisan", factory, metadata)
	assert.NoError(t, err)

	// Verify command can be retrieved by name
	f, exists := registry.Get("artisan")
	assert.True(t, exists)
	assert.NotNil(t, f)

	// Verify command can be retrieved by aliases
	f, exists = registry.Get("a")
	assert.True(t, exists)
	assert.NotNil(t, f)

	f, exists = registry.Get("art")
	assert.True(t, exists)
	assert.NotNil(t, f)

	// Verify metadata can be retrieved by aliases
	meta, exists := registry.GetMetadata("a")
	assert.True(t, exists)
	assert.Equal(t, "artisan", meta.Name)
	assert.Equal(t, []string{"a", "art"}, meta.Aliases)
}

func TestRegistry_AliasConflicts(t *testing.T) {
	registry := NewRegistry()

	factory1 := func() *cobra.Command {
		return &cobra.Command{Use: "artisan"}
	}

	factory2 := func() *cobra.Command {
		return &cobra.Command{Use: "another"}
	}

	// Register first command with alias
	err := registry.Register("artisan", factory1, Metadata{
		Name:    "artisan",
		Aliases: []string{"a"},
	})
	assert.NoError(t, err)

	// Try to register another command with same alias - should fail
	err = registry.Register("another", factory2, Metadata{
		Name:    "another",
		Aliases: []string{"a"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "alias a already registered")

	// Try to register a command with name that conflicts with existing alias
	err = registry.Register("a", factory2, Metadata{
		Name: "a",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "item name a conflicts with existing alias")
}

func TestRegistry_ResolveAlias(t *testing.T) {
	registry := NewRegistry()

	factory := func() *cobra.Command {
		return &cobra.Command{Use: "test"}
	}

	registry.Register("test", factory, Metadata{
		Name:    "test",
		Aliases: []string{"t", "tst"},
	})

	// Test resolving aliases
	canonical, ok := registry.ResolveAlias("t")
	assert.True(t, ok)
	assert.Equal(t, "test", canonical)

	canonical, ok = registry.ResolveAlias("tst")
	assert.True(t, ok)
	assert.Equal(t, "test", canonical)

	// Test resolving non-existent alias
	canonical, ok = registry.ResolveAlias("nonexistent")
	assert.False(t, ok)
	assert.Empty(t, canonical)

	// Test that command name is not resolved as alias
	canonical, ok = registry.ResolveAlias("test")
	assert.False(t, ok)
	assert.Empty(t, canonical)
}

func TestRegistry_GetAliases(t *testing.T) {
	registry := NewRegistry()

	factory := func() *cobra.Command {
		return &cobra.Command{Use: "composer"}
	}

	registry.Register("composer", factory, Metadata{
		Name:    "composer",
		Aliases: []string{"c", "comp"},
	})

	// Get aliases for existing command
	aliases := registry.GetAliases("composer")
	assert.Equal(t, []string{"c", "comp"}, aliases)

	// Get aliases for non-existent command
	aliases = registry.GetAliases("nonexistent")
	assert.Nil(t, aliases)
}

func TestRegistry_IsAlias(t *testing.T) {
	registry := NewRegistry()

	factory := func() *cobra.Command {
		return &cobra.Command{Use: "artisan"}
	}

	registry.Register("artisan", factory, Metadata{
		Name:    "artisan",
		Aliases: []string{"a"},
	})

	// Test checking aliases
	assert.True(t, registry.IsAlias("a"))
	assert.False(t, registry.IsAlias("artisan"))
	assert.False(t, registry.IsAlias("nonexistent"))
}

func TestRegistry_CreateAllWithAliases(t *testing.T) {
	registry := NewRegistry()

	factory1 := func() *cobra.Command {
		return &cobra.Command{
			Use:   "artisan",
			Short: "Artisan command",
		}
	}

	factory2 := func() *cobra.Command {
		return &cobra.Command{
			Use:   "composer",
			Short: "Composer command",
		}
	}

	registry.Register("artisan", factory1, Metadata{
		Name:    "artisan",
		Aliases: []string{"a"},
	})

	registry.Register("composer", factory2, Metadata{
		Name:    "composer",
		Aliases: []string{"c", "comp"},
	})

	// Create all commands
	commands := registry.CreateAll()
	assert.Len(t, commands, 2)

	// Find artisan command and check aliases
	for _, cmd := range commands {
		if cmd.Use == "artisan" {
			assert.Equal(t, []string{"a"}, cmd.Aliases)
		} else if cmd.Use == "composer" {
			assert.Equal(t, []string{"c", "comp"}, cmd.Aliases)
		}
	}
}

func TestRegistry_CreateByCategoryWithAliases(t *testing.T) {
	registry := NewRegistry()

	devFactory := func() *cobra.Command {
		return &cobra.Command{Use: "test"}
	}

	dockerFactory := func() *cobra.Command {
		return &cobra.Command{Use: "up"}
	}

	registry.Register("test", devFactory, Metadata{
		Category: CategoryDeveloper,
		Aliases:  []string{"t"},
	})

	registry.Register("up", dockerFactory, Metadata{
		Category: CategoryDocker,
		Aliases:  []string{"u"},
	})

	// Create developer commands
	devCommands := registry.CreateByCategory(CategoryDeveloper)
	assert.Len(t, devCommands, 1)
	assert.Equal(t, "test", devCommands[0].Use)
	assert.Equal(t, []string{"t"}, devCommands[0].Aliases)

	// Create docker commands
	dockerCommands := registry.CreateByCategory(CategoryDocker)
	assert.Len(t, dockerCommands, 1)
	assert.Equal(t, "up", dockerCommands[0].Use)
	assert.Equal(t, []string{"u"}, dockerCommands[0].Aliases)
}
