package context

import (
	"errors"
	"time"
)

// Common errors
var (
	ErrProjectRootNotFound = errors.New("could not find project root")
)

// DevelopmentMode represents the project's development mode
type DevelopmentMode string

const (
	ModeMultiWorktree DevelopmentMode = "multi-worktree"
	ModeSingleRepo    DevelopmentMode = "single-repo"
	ModeStandalone    DevelopmentMode = "standalone" // Non-Git project with .glide.yml
	ModeUnknown       DevelopmentMode = ""
)

// LocationType represents where the command is being executed from
type LocationType string

const (
	LocationRoot     LocationType = "root"      // Project root (in multi-worktree mode)
	LocationMainRepo LocationType = "main-repo" // vcs/ directory (in multi-worktree mode)
	LocationWorktree LocationType = "worktree"  // worktrees/*/ directory
	LocationProject  LocationType = "project"   // Single repo mode
	LocationUnknown  LocationType = ""
)

// ContainerState represents the state of a Docker container
type ContainerState string

const (
	ContainerRunning ContainerState = "running"
	ContainerStopped ContainerState = "stopped"
	ContainerExited  ContainerState = "exited"
	ContainerUnknown ContainerState = "unknown"
)

// ContainerStatus represents the status of a Docker container
type ContainerStatus struct {
	Name      string
	Status    string // running, stopped, exited, etc.
	Health    string // healthy, unhealthy, starting, none
	StartedAt time.Time
	Ports     []string
}

// ProjectContext contains all context information about the current project
type ProjectContext struct {
	// Core paths
	WorkingDir  string // Current working directory
	ProjectRoot string // Project root directory
	ProjectName string // Name of the project from config

	// Development mode and location
	DevelopmentMode DevelopmentMode // multi-worktree or single-repo
	Location        LocationType    // Where command is being run from

	// Multi-worktree specific
	IsRoot       bool   // True if in project root (multi-worktree only)
	IsMainRepo   bool   // True if in vcs/ (multi-worktree only)
	IsWorktree   bool   // True if in worktrees/*/ (multi-worktree only)
	WorktreeName string // Name of current worktree if applicable

	// Docker configuration
	ComposeFiles     []string                   // Resolved docker-compose files
	ComposeOverride  string                     // Path to override file
	DockerRunning    bool                       // Is Docker daemon running
	ContainersStatus map[string]ContainerStatus // Status of all containers

	// Framework detection
	DetectedFrameworks []string                     // List of detected framework names
	FrameworkVersions  map[string]string            // Framework name -> version mapping
	FrameworkCommands  map[string]string            // Commands provided by frameworks
	FrameworkMetadata  map[string]map[string]string // Framework -> metadata mapping

	// Command context
	CommandScope string // "global" or "local"

	// Error if context detection failed
	Error error
}

// IsValid returns true if the context was successfully detected
func (c *ProjectContext) IsValid() bool {
	return c.Error == nil && c.ProjectRoot != ""
}

// IsGlobalScope returns true if the current command should run in global scope
func (c *ProjectContext) IsGlobalScope() bool {
	return c.CommandScope == "global"
}

// CanUseProjectCommands returns true if project-wide commands are available
func (c *ProjectContext) CanUseProjectCommands() bool {
	return c.DevelopmentMode == ModeMultiWorktree
}

// GetComposeCommand builds the docker-compose command with proper file flags
func (c *ProjectContext) GetComposeCommand() []string {
	args := []string{"docker", "compose"}
	for _, file := range c.ComposeFiles {
		args = append(args, "-f", file)
	}
	return args
}
