package cli

import (
	"testing"

	"github.com/ivannovak/glide/pkg/app"
	"github.com/stretchr/testify/assert"
)

func TestNewBuilder(t *testing.T) {
	application := &app.Application{}

	builder := NewBuilder(application)

	assert.NotNil(t, builder)
	assert.NotNil(t, builder.GetRegistry())
}

func TestBuilder_Build(t *testing.T) {
	application := &app.Application{}

	builder := NewBuilder(application)
	rootCmd := builder.Build()

	assert.NotNil(t, rootCmd)
	assert.Equal(t, "glid", rootCmd.Use)
	assert.Contains(t, rootCmd.Short, "CLI")
}

func TestBuilder_RegisterCommands(t *testing.T) {
	application := &app.Application{}

	builder := NewBuilder(application)
	registry := builder.GetRegistry()

	// Check that commands are registered
	_, exists := registry.Get("setup")
	assert.True(t, exists)
	_, exists = registry.Get("config")
	assert.True(t, exists)
	_, exists = registry.Get("completion")
	assert.True(t, exists)
	_, exists = registry.Get("version")
	assert.True(t, exists)
	_, exists = registry.Get("help")
	assert.True(t, exists)
	_, exists = registry.Get("up")
	assert.True(t, exists)
	_, exists = registry.Get("down")
	assert.True(t, exists)
}

func TestBuilder_GetRegistry(t *testing.T) {
	application := &app.Application{}

	builder := NewBuilder(application)
	registry := builder.GetRegistry()

	assert.NotNil(t, registry)
}

func TestBuilder_CommandCategories(t *testing.T) {
	application := &app.Application{}

	builder := NewBuilder(application)
	registry := builder.GetRegistry()

	// Check core commands
	coreCommands := registry.GetByCategory(CategoryCore)
	assert.Contains(t, coreCommands, "setup")
	assert.Contains(t, coreCommands, "completion")
	assert.Contains(t, coreCommands, "version")
	assert.Contains(t, coreCommands, "global")
	assert.Contains(t, coreCommands, "self-update")

	// Check docker commands
	dockerCommands := registry.GetByCategory(CategoryDocker)
	assert.Contains(t, dockerCommands, "up")
	assert.Contains(t, dockerCommands, "down")
	assert.Contains(t, dockerCommands, "status")
	assert.Contains(t, dockerCommands, "logs")
	assert.Contains(t, dockerCommands, "shell")

	// Check developer commands
	developerCommands := registry.GetByCategory(CategoryDeveloper)
	assert.Contains(t, developerCommands, "test")
	assert.Contains(t, developerCommands, "artisan")
	assert.Contains(t, developerCommands, "composer")
	assert.Contains(t, developerCommands, "lint")
}

func TestBuilder_CommandMetadata(t *testing.T) {
	application := &app.Application{}

	builder := NewBuilder(application)
	registry := builder.GetRegistry()

	// Check setup command metadata
	metadata, exists := registry.GetMetadata("setup")
	assert.True(t, exists)
	assert.Equal(t, "setup", metadata.Name)
	assert.Equal(t, CategoryCore, metadata.Category)
	assert.Equal(t, "Initial setup and configuration", metadata.Description)

	// Check config command metadata (hidden debug command)
	metadata, exists = registry.GetMetadata("config")
	assert.True(t, exists)
	assert.Equal(t, "config", metadata.Name)
	assert.Equal(t, CategoryDebug, metadata.Category)
	assert.True(t, metadata.Hidden)
}

func TestBuilder_CreateCommands(t *testing.T) {
	application := &app.Application{}

	builder := NewBuilder(application)
	registry := builder.GetRegistry()

	// Test creating all commands
	commands := registry.CreateAll()
	assert.True(t, len(commands) > 0)

	// Test creating commands by category
	coreCommands := registry.CreateByCategory(CategoryCore)
	assert.True(t, len(coreCommands) > 0)

	dockerCommands := registry.CreateByCategory(CategoryDocker)
	assert.True(t, len(dockerCommands) > 0)
}
