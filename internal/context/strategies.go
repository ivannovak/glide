package context

import (
	"os"
	"path/filepath"
	"strings"
)

// DetectionStrategy defines the interface for context detection strategies
type DetectionStrategy interface {
	Detect(workingDir string) (*ProjectContext, error)
	Name() string
}

// ProjectRootFinder finds the project root directory
type ProjectRootFinder interface {
	FindRoot(workingDir string) (string, error)
}

// DevelopmentModeDetector determines the development mode
type DevelopmentModeDetector interface {
	DetectMode(projectRoot string) DevelopmentMode
}

// LocationIdentifier identifies the current location in the project
type LocationIdentifier interface {
	IdentifyLocation(ctx *ProjectContext, workingDir string) LocationType
}

// ComposeFileResolver resolves docker-compose files
type ComposeFileResolver interface {
	ResolveFiles(ctx *ProjectContext) []string
}

// DockerStatusChecker checks Docker daemon status
type DockerStatusChecker interface {
	CheckStatus(ctx *ProjectContext) bool
}

// StandardProjectRootFinder implements the standard root finding logic
type StandardProjectRootFinder struct {
	maxTraversal int
}

// NewStandardProjectRootFinder creates a new standard root finder
func NewStandardProjectRootFinder() *StandardProjectRootFinder {
	return &StandardProjectRootFinder{
		maxTraversal: 5,
	}
}

// FindRoot finds the project root directory
func (f *StandardProjectRootFinder) FindRoot(workingDir string) (string, error) {
	current := workingDir
	traversed := 0

	for traversed < f.maxTraversal {
		// Check for multi-worktree structure (has vcs/ directory)
		vcsPath := filepath.Join(current, "vcs")
		if info, err := os.Stat(vcsPath); err == nil && info.IsDir() {
			// Check if vcs contains a git repo
			gitPath := filepath.Join(vcsPath, ".git")
			if _, err := os.Stat(gitPath); err == nil {
				return current, nil
			}
		}

		// Check for single-repo structure (has .git in current)
		gitPath := filepath.Join(current, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			// Make sure this isn't inside vcs/ or worktrees/
			if !strings.Contains(current, "/vcs") && !strings.Contains(current, "/worktrees/") {
				return current, nil
			}
		}

		// Check if we're inside a worktree
		if strings.Contains(current, "/worktrees/") {
			// Find the project root (should be two levels up from worktrees/*/
			parts := strings.Split(current, "/worktrees/")
			if len(parts) > 0 {
				return parts[0], nil
			}
		}

		// Check if we're inside vcs/
		if strings.Contains(current, "/vcs") {
			// Project root should be one level up
			parts := strings.Split(current, "/vcs")
			if len(parts) > 0 {
				return parts[0], nil
			}
		}

		// Move up one directory
		parent := filepath.Dir(current)
		if parent == current {
			break // Reached filesystem root
		}
		current = parent
		traversed++
	}

	return "", ErrProjectRootNotFound
}

// StandardDevelopmentModeDetector implements standard mode detection
type StandardDevelopmentModeDetector struct{}

// NewStandardDevelopmentModeDetector creates a new mode detector
func NewStandardDevelopmentModeDetector() *StandardDevelopmentModeDetector {
	return &StandardDevelopmentModeDetector{}
}

// DetectMode determines the development mode
func (d *StandardDevelopmentModeDetector) DetectMode(projectRoot string) DevelopmentMode {
	// Check for vcs/ directory in project root
	vcsPath := filepath.Join(projectRoot, "vcs")
	if info, err := os.Stat(vcsPath); err == nil && info.IsDir() {
		// Check for worktrees/ directory
		worktreesPath := filepath.Join(projectRoot, "worktrees")
		if info, err := os.Stat(worktreesPath); err == nil && info.IsDir() {
			return ModeMultiWorktree
		}
	}

	// Check if project root itself is a git repo
	gitPath := filepath.Join(projectRoot, ".git")
	if _, err := os.Stat(gitPath); err == nil {
		return ModeSingleRepo
	}

	return ModeUnknown
}

// StandardLocationIdentifier implements standard location identification
type StandardLocationIdentifier struct{}

// NewStandardLocationIdentifier creates a new location identifier
func NewStandardLocationIdentifier() *StandardLocationIdentifier {
	return &StandardLocationIdentifier{}
}

// IdentifyLocation determines where in the project structure we are
func (i *StandardLocationIdentifier) IdentifyLocation(ctx *ProjectContext, workingDir string) LocationType {
	relPath, err := filepath.Rel(ctx.ProjectRoot, workingDir)
	if err != nil {
		return LocationUnknown
	}

	// Normalize the path
	relPath = filepath.ToSlash(relPath)

	switch ctx.DevelopmentMode {
	case ModeMultiWorktree:
		if relPath == "." {
			ctx.IsRoot = true
			return LocationRoot
		} else if relPath == "vcs" || strings.HasPrefix(relPath, "vcs/") {
			ctx.IsMainRepo = true
			return LocationMainRepo
		} else if strings.HasPrefix(relPath, "worktrees/") {
			ctx.IsWorktree = true
			
			// Extract worktree name
			parts := strings.Split(relPath, "/")
			if len(parts) >= 2 {
				ctx.WorktreeName = parts[1]
			}
			return LocationWorktree
		} else {
			ctx.IsRoot = true
			return LocationRoot
		}
	case ModeSingleRepo:
		return LocationProject
	default:
		return LocationUnknown
	}
}

// StandardComposeFileResolver implements standard compose file resolution
type StandardComposeFileResolver struct{}

// NewStandardComposeFileResolver creates a new compose file resolver
func NewStandardComposeFileResolver() *StandardComposeFileResolver {
	return &StandardComposeFileResolver{}
}

// ResolveFiles finds all docker-compose files based on location
func (r *StandardComposeFileResolver) ResolveFiles(ctx *ProjectContext) []string {
	files := []string{}

	switch ctx.Location {
	case LocationMainRepo:
		// From vcs/: docker-compose.yml + ../docker-compose.override.yml
		composePath := filepath.Join(ctx.ProjectRoot, "vcs", "docker-compose.yml")
		if _, err := os.Stat(composePath); err == nil {
			files = append(files, composePath)
		}
		
		overridePath := filepath.Join(ctx.ProjectRoot, "docker-compose.override.yml")
		if _, err := os.Stat(overridePath); err == nil {
			ctx.ComposeOverride = overridePath
			files = append(files, overridePath)
		}

	case LocationWorktree:
		// From worktrees/*/: docker-compose.yml + ../../docker-compose.override.yml
		worktreePath := filepath.Join(ctx.ProjectRoot, "worktrees", ctx.WorktreeName)
		composePath := filepath.Join(worktreePath, "docker-compose.yml")
		if _, err := os.Stat(composePath); err == nil {
			files = append(files, composePath)
		}
		
		overridePath := filepath.Join(ctx.ProjectRoot, "docker-compose.override.yml")
		if _, err := os.Stat(overridePath); err == nil {
			ctx.ComposeOverride = overridePath
			files = append(files, overridePath)
		}

	case LocationProject:
		// Single-repo mode: docker-compose.yml + docker-compose.override.yml
		composePath := filepath.Join(ctx.ProjectRoot, "docker-compose.yml")
		if _, err := os.Stat(composePath); err == nil {
			files = append(files, composePath)
		}
		
		overridePath := filepath.Join(ctx.ProjectRoot, "docker-compose.override.yml")
		if _, err := os.Stat(overridePath); err == nil {
			ctx.ComposeOverride = overridePath
			files = append(files, overridePath)
		}
	}

	return files
}