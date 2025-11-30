// Package branding provides customizable branding for the Glide CLI.
//
// All branding elements can be overridden at build time using ldflags,
// allowing the creation of white-label versions of the CLI with custom
// names, descriptions, and configuration paths.
//
// # Build-time Customization
//
// Override branding at build time:
//
//	go build -ldflags "\
//	    -X github.com/ivannovak/glide/v2/pkg/branding.CommandName=mycli \
//	    -X github.com/ivannovak/glide/v2/pkg/branding.ProjectName=MyProject \
//	    -X github.com/ivannovak/glide/v2/pkg/branding.ConfigFileName=.mycli.yml"
//
// # Available Variables
//
// The following variables can be customized:
//   - CommandName: The CLI command name (default: "glide")
//   - ConfigFileName: The configuration file name (default: ".glide.yml")
//   - ProjectName: The project display name (default: "Glide")
//   - Description: Short description shown in help (default: "context-aware development CLI")
//   - RepositoryURL: URL for updates and documentation
//
// # Directory Structure
//
// Plugin directories, global config paths, and completion directories are
// automatically derived from the branding configuration:
//
//	~/.glide/           # Global config directory (derived from ConfigFileName)
//	~/.glide/plugins/   # Global plugin directory
//	.glide/plugins/     # Local plugin directory (in project)
package branding
