package context

import (
	"fmt"
	"os"
	"os/exec"
)

// Detector is a refactored context detector using composition
type Detector struct {
	workingDir         string
	rootFinder         ProjectRootFinder
	modeDetector       DevelopmentModeDetector
	locationIdentifier LocationIdentifier
	composeResolver    ComposeFileResolver
}

// NewDetector creates a new context detector with default strategies
func NewDetector() (*Detector, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	return &Detector{
		workingDir:         wd,
		rootFinder:         NewStandardProjectRootFinder(),
		modeDetector:       NewStandardDevelopmentModeDetector(),
		locationIdentifier: NewStandardLocationIdentifier(),
		composeResolver:    NewStandardComposeFileResolver(),
	}, nil
}

// NewDetectorWithStrategies creates a detector with custom strategies
func NewDetectorWithStrategies(
	rootFinder ProjectRootFinder,
	modeDetector DevelopmentModeDetector,
	locationIdentifier LocationIdentifier,
	composeResolver ComposeFileResolver,
) (*Detector, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	return &Detector{
		workingDir:         wd,
		rootFinder:         rootFinder,
		modeDetector:       modeDetector,
		locationIdentifier: locationIdentifier,
		composeResolver:    composeResolver,
	}, nil
}

// SetRootFinder sets a custom root finder
func (d *Detector) SetRootFinder(finder ProjectRootFinder) {
	d.rootFinder = finder
}

// SetModeDetector sets a custom mode detector
func (d *Detector) SetModeDetector(detector DevelopmentModeDetector) {
	d.modeDetector = detector
}

// SetLocationIdentifier sets a custom location identifier
func (d *Detector) SetLocationIdentifier(identifier LocationIdentifier) {
	d.locationIdentifier = identifier
}

// SetComposeResolver sets a custom compose resolver
func (d *Detector) SetComposeResolver(resolver ComposeFileResolver) {
	d.composeResolver = resolver
}

// Detect analyzes the current environment and returns project context
func (d *Detector) Detect() (*ProjectContext, error) {
	ctx := &ProjectContext{
		WorkingDir: d.workingDir,
	}

	// Find project root
	projectRoot, err := d.rootFinder.FindRoot(d.workingDir)
	if err != nil {
		ctx.Error = err
		return ctx, err
	}
	ctx.ProjectRoot = projectRoot

	// Detect development mode
	ctx.DevelopmentMode = d.modeDetector.DetectMode(ctx.ProjectRoot)

	// Identify current location
	ctx.Location = d.locationIdentifier.IdentifyLocation(ctx, d.workingDir)

	// Resolve docker-compose files
	ctx.ComposeFiles = d.composeResolver.ResolveFiles(ctx)

	// Check Docker daemon status
	d.checkDockerStatus(ctx)

	return ctx, nil
}

// checkDockerStatus checks if Docker daemon is running
func (d *Detector) checkDockerStatus(ctx *ProjectContext) {
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err == nil {
		ctx.DockerRunning = true

		// Get container status if compose files are available
		if len(ctx.ComposeFiles) > 0 {
			d.getContainerStatus(ctx)
		}
	} else {
		ctx.DockerRunning = false
	}
}

// getContainerStatus retrieves status of Docker containers
func (d *Detector) getContainerStatus(ctx *ProjectContext) {
	ctx.ContainersStatus = make(map[string]ContainerStatus)

	// Build docker-compose ps command
	args := []string{"compose"}
	for _, file := range ctx.ComposeFiles {
		args = append(args, "-f", file)
	}
	args = append(args, "ps", "--format", "json", "--all")

	// Execute command
	cmd := exec.Command("docker", args...)
	_, err := cmd.Output()
	if err != nil {
		return
	}

	// Container status parsing is handled by docker.ContainerManager
	// This basic check just verifies containers exist
}

// DetectCommandScope determines if a command should run in global or local scope
func (d *Detector) DetectCommandScope(ctx *ProjectContext, isGlobalFlag bool) {
	if isGlobalFlag {
		ctx.CommandScope = "global"
		return
	}

	// In multi-worktree mode at root, default to global
	if ctx.DevelopmentMode == ModeMultiWorktree && ctx.IsRoot {
		ctx.CommandScope = "global"
		return
	}

	ctx.CommandScope = "local"
}

// DetectorBuilder provides a fluent API for building detectors
type DetectorBuilder struct {
	rootFinder         ProjectRootFinder
	modeDetector       DevelopmentModeDetector
	locationIdentifier LocationIdentifier
	composeResolver    ComposeFileResolver
}

// NewDetectorBuilder creates a new detector builder
func NewDetectorBuilder() *DetectorBuilder {
	return &DetectorBuilder{}
}

// WithRootFinder sets the root finder
func (b *DetectorBuilder) WithRootFinder(finder ProjectRootFinder) *DetectorBuilder {
	b.rootFinder = finder
	return b
}

// WithModeDetector sets the mode detector
func (b *DetectorBuilder) WithModeDetector(detector DevelopmentModeDetector) *DetectorBuilder {
	b.modeDetector = detector
	return b
}

// WithLocationIdentifier sets the location identifier
func (b *DetectorBuilder) WithLocationIdentifier(identifier LocationIdentifier) *DetectorBuilder {
	b.locationIdentifier = identifier
	return b
}

// WithComposeResolver sets the compose resolver
func (b *DetectorBuilder) WithComposeResolver(resolver ComposeFileResolver) *DetectorBuilder {
	b.composeResolver = resolver
	return b
}

// Build creates the detector with the configured strategies
func (b *DetectorBuilder) Build() (*Detector, error) {
	// Use defaults for any unset strategies
	if b.rootFinder == nil {
		b.rootFinder = NewStandardProjectRootFinder()
	}
	if b.modeDetector == nil {
		b.modeDetector = NewStandardDevelopmentModeDetector()
	}
	if b.locationIdentifier == nil {
		b.locationIdentifier = NewStandardLocationIdentifier()
	}
	if b.composeResolver == nil {
		b.composeResolver = NewStandardComposeFileResolver()
	}

	return NewDetectorWithStrategies(
		b.rootFinder,
		b.modeDetector,
		b.locationIdentifier,
		b.composeResolver,
	)
}
