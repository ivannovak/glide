// Package version provides version information and build metadata.
//
// This package contains the current version of Glide and build-time
// information. Values can be set at build time using ldflags.
//
// # Build-time Configuration
//
// Set version information at build time:
//
//	go build -ldflags "\
//	    -X github.com/glide-cli/glide/v3/pkg/version.Version=1.2.3 \
//	    -X github.com/glide-cli/glide/v3/pkg/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
//	    -X github.com/glide-cli/glide/v3/pkg/version.GitCommit=$(git rev-parse HEAD)"
//
// # Accessing Version Information
//
// Get the current version:
//
//	v := version.Get()
//	fmt.Println("Glide version:", v)
//
// Get comprehensive build info:
//
//	info := version.GetBuildInfo()
//	fmt.Printf("Version: %s\n", info.Version)
//	fmt.Printf("Built: %s\n", info.BuildDate)
//	fmt.Printf("Commit: %s\n", info.GitCommit)
//	fmt.Printf("Go: %s\n", info.GoVersion)
//	fmt.Printf("OS/Arch: %s/%s\n", info.OS, info.Architecture)
//
// # Format Strings
//
// Get formatted version strings:
//
//	// Simple version
//	simple := version.GetSimple() // "1.2.3"
//
//	// Detailed version
//	detailed := version.GetDetailed() // "glide v1.2.3 (abc123)"
//
//	// Full build info
//	full := version.GetFull() // Multi-line with all build details
//
// # Semantic Versioning
//
// Versions follow semantic versioning (semver.org):
//   - MAJOR: Breaking changes
//   - MINOR: New features, backward compatible
//   - PATCH: Bug fixes, backward compatible
package version
