// Package context provides project context detection for Glide.
//
// This package detects the current project's context including project root,
// development mode (single-repo vs multi-worktree), Docker compose files,
// and custom extensions provided by plugins.
//
// # Context Detection
//
// Create a detector and detect the current context:
//
//	detector, err := context.NewDetector()
//	if err != nil {
//	    return err
//	}
//
//	ctx, err := detector.Detect()
//	if err != nil {
//	    return err
//	}
//
//	fmt.Printf("Project root: %s\n", ctx.ProjectRoot)
//	fmt.Printf("Mode: %s\n", ctx.DevelopmentMode)
//
// # Fast Detection
//
// For startup-critical paths, use fast detection that skips expensive checks:
//
//	detector, err := context.NewDetectorFast()
//	// Skips Docker daemon status check
//
// # Project Context
//
// The ProjectContext contains detected information:
//
//	type ProjectContext struct {
//	    ProjectRoot      string                 // Absolute path to project root
//	    DevelopmentMode  DevelopmentMode        // single-repo or multi-worktree
//	    WorkingDirectory string                 // Current working directory
//	    LocationType     LocationType           // project, worktree, or external
//	    ComposeFiles     []string               // Docker compose files
//	    DockerAvailable  bool                   // Docker daemon status
//	    Extensions       map[string]interface{} // Plugin-provided extensions
//	}
//
// # Development Modes
//
// Two development modes are supported:
//
//	DevelopmentModeSingleRepo    // Standard single repository
//	DevelopmentModeMultiWorktree // Git worktree with main repository
//
// # Extension Registry
//
// Plugins can register context extensions:
//
//	registry := context.NewExtensionRegistry()
//	registry.Register("my-extension", myDetector)
//
//	detector := context.NewDetector()
//	detector.SetExtensionRegistry(registry)
//
// # Lazy Docker Checking
//
// By default, Docker status is checked lazily to improve startup time:
//
//	// Docker status is deferred
//	ctx, _ := detector.Detect()
//
//	// Check status when needed
//	detector.EnsureDockerStatus(ctx)
package context
