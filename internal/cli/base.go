package cli

import (
	"github.com/ivannovak/glide/v2/internal/config"
	"github.com/ivannovak/glide/v2/internal/context"
	"github.com/ivannovak/glide/v2/pkg/app"
	"github.com/ivannovak/glide/v2/pkg/output"
)

// BaseCommand provides common functionality for all commands
type BaseCommand struct {
	app *app.Application
}

// NewBaseCommand creates a new base command
func NewBaseCommand(application *app.Application) BaseCommand {
	return BaseCommand{
		app: application,
	}
}

// Output returns the output manager
func (b *BaseCommand) Output() *output.Manager {
	return b.app.OutputManager
}

// Context returns the project context
func (b *BaseCommand) Context() *context.ProjectContext {
	return b.app.ProjectContext
}

// Config returns the configuration
func (b *BaseCommand) Config() *config.Config {
	return b.app.Config
}

// Application returns the full application
func (b *BaseCommand) Application() *app.Application {
	return b.app
}
