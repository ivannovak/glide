package cli

import (
	"github.com/ivannovak/glide/v2/internal/config"
	"github.com/ivannovak/glide/v2/internal/context"
	"github.com/ivannovak/glide/v2/pkg/output"
)

// BaseCommand provides common functionality for all commands
type BaseCommand struct {
	outputManager  *output.Manager
	projectContext *context.ProjectContext
	config         *config.Config
}

// NewBaseCommand creates a new base command
func NewBaseCommand(
	outputManager *output.Manager,
	projectContext *context.ProjectContext,
	config *config.Config,
) BaseCommand {
	return BaseCommand{
		outputManager:  outputManager,
		projectContext: projectContext,
		config:         config,
	}
}

// Output returns the output manager
func (b *BaseCommand) Output() *output.Manager {
	return b.outputManager
}

// Context returns the project context
func (b *BaseCommand) Context() *context.ProjectContext {
	return b.projectContext
}

// Config returns the configuration
func (b *BaseCommand) Config() *config.Config {
	return b.config
}
