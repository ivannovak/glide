package version

import (
	"fmt"
	"runtime"
)

// Build information variables
var (
	Version   = "0.8.1"
	BuildDate = "unknown"
	GitCommit = "unknown"
)

// BuildInfo contains build-time information
type BuildInfo struct {
	Version      string
	BuildDate    string
	GitCommit    string
	GoVersion    string
	OS           string
	Architecture string
	Compiler     string
}

// SetBuildInfo sets all build information
func SetBuildInfo(version, buildDate, gitCommit string) {
	Version = version
	BuildDate = buildDate
	GitCommit = gitCommit
}

// Set sets the version (legacy compatibility)
func Set(v string) {
	Version = v
}

// Get returns the current version
func Get() string {
	return Version
}

// GetBuildInfo returns comprehensive build information
func GetBuildInfo() BuildInfo {
	return BuildInfo{
		Version:      Version,
		BuildDate:    BuildDate,
		GitCommit:    GitCommit,
		GoVersion:    runtime.Version(),
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		Compiler:     runtime.Compiler,
	}
}

// GetVersionString returns a formatted version string
func GetVersionString() string {
	if Version == "dev" {
		return fmt.Sprintf("glid version %s (development build)", Version)
	}
	return fmt.Sprintf("glid version %s", Version)
}

// GetSystemInfo returns formatted system information
func GetSystemInfo() string {
	info := GetBuildInfo()
	return fmt.Sprintf("OS: %s, Architecture: %s, Go: %s",
		info.OS, info.Architecture, info.GoVersion)
}
