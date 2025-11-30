package context

import (
	"context"

	"github.com/ivannovak/glide/v3/pkg/plugin/sdk"
)

// pluginExtensionAdapter adapts the plugin system to the context ExtensionRegistry interface
type pluginExtensionAdapter struct {
	providers []interface{}
}

// DetectAll runs detection for all registered plugins that provide context extensions
func (a *pluginExtensionAdapter) DetectAll(projectRoot string) (map[string]interface{}, error) {
	results := make(map[string]interface{})
	ctx := context.Background()

	for _, p := range a.providers {
		// Check if plugin provides context extension
		provider, ok := p.(sdk.ContextProvider)
		if !ok {
			continue
		}

		ext := provider.ProvideContext()
		if ext == nil {
			continue
		}

		// Detect extension data
		data, err := ext.Detect(ctx, projectRoot)
		if err != nil {
			// Continue with other extensions if one fails
			// Don't break the entire detection process
			continue
		}

		if data != nil {
			results[ext.Name()] = data
		}
	}

	return results, nil
}
