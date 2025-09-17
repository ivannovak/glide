package context

import (
	"os"
	"path/filepath"
	"testing"
)

// BenchmarkNewDetector benchmarks detector creation
func BenchmarkNewDetector(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		detector, err := NewDetector()
		if err != nil {
			b.Fatal(err)
		}
		_ = detector
	}
}

// BenchmarkDetector_Detect benchmarks context detection
func BenchmarkDetector_Detect(b *testing.B) {
	detector, err := NewDetector()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx, _ := detector.Detect()
		_ = ctx
	}
}

// BenchmarkProjectRootFinder_FindRoot benchmarks project root finding
func BenchmarkProjectRootFinder_FindRoot(b *testing.B) {
	finder := NewStandardProjectRootFinder()
	wd, err := os.Getwd()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = finder.FindRoot(wd)
	}
}

// BenchmarkDevelopmentModeDetector_DetectMode benchmarks development mode detection
func BenchmarkDevelopmentModeDetector_DetectMode(b *testing.B) {
	detector := NewStandardDevelopmentModeDetector()
	projectRoot := "/home/user/project"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		mode := detector.DetectMode(projectRoot)
		_ = mode
	}
}

// BenchmarkLocationIdentifier_IdentifyLocation benchmarks location identification
func BenchmarkLocationIdentifier_IdentifyLocation(b *testing.B) {
	identifier := NewStandardLocationIdentifier()
	workingDir := "/home/user/project/vcs"
	projectRoot := "/home/user/project"
	mode := ModeMultiWorktree

	ctx := &ProjectContext{
		ProjectRoot:     projectRoot,
		DevelopmentMode: mode,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		locType := identifier.IdentifyLocation(ctx, workingDir)
		_ = locType
	}
}

// BenchmarkComposeFileResolver_ResolveFiles benchmarks compose file resolution
func BenchmarkComposeFileResolver_ResolveFiles(b *testing.B) {
	resolver := NewStandardComposeFileResolver()

	// Create temporary test environment
	tempDir := b.TempDir()
	composeFile := filepath.Join(tempDir, "docker-compose.yml")
	err := os.WriteFile(composeFile, []byte("version: '3'"), 0644)
	if err != nil {
		b.Fatal(err)
	}

	ctx := &ProjectContext{
		ProjectRoot: tempDir,
		Location:    LocationProject,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		files := resolver.ResolveFiles(ctx)
		_ = files
	}
}

// BenchmarkProjectContext_Creation benchmarks project context creation
func BenchmarkProjectContext_Creation(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx := &ProjectContext{
			WorkingDir:      "/home/user/project",
			ProjectRoot:     "/home/user/project",
			ProjectName:     "myproject",
			DevelopmentMode: ModeMultiWorktree,
			Location:        LocationMainRepo,
			IsRoot:          false,
			IsMainRepo:      true,
			IsWorktree:      false,
			WorktreeName:    "",
			ComposeFiles:    []string{"docker-compose.yml"},
			DockerRunning:   true,
			CommandScope:    "local",
		}
		_ = ctx
	}
}

// BenchmarkProjectContext_IsValid benchmarks context validation
func BenchmarkProjectContext_IsValid(b *testing.B) {
	ctx := &ProjectContext{
		ProjectRoot: "/home/user/project",
		Error:       nil,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		valid := ctx.IsValid()
		_ = valid
	}
}

// BenchmarkProjectContext_GetComposeCommand benchmarks compose command generation
func BenchmarkProjectContext_GetComposeCommand(b *testing.B) {
	ctx := &ProjectContext{
		ComposeFiles: []string{
			"docker-compose.yml",
			"docker-compose.override.yml",
			"docker-compose.local.yml",
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		cmd := ctx.GetComposeCommand()
		_ = cmd
	}
}

// BenchmarkProjectContext_CanUseProjectCommands benchmarks project command check
func BenchmarkProjectContext_CanUseProjectCommands(b *testing.B) {
	ctx := &ProjectContext{
		DevelopmentMode: ModeMultiWorktree,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		canUse := ctx.CanUseProjectCommands()
		_ = canUse
	}
}

// BenchmarkDetector_DetectWithCaching benchmarks detection with different scenarios
func BenchmarkDetector_DetectWithCaching(b *testing.B) {
	detector, err := NewDetector()
	if err != nil {
		b.Fatal(err)
	}

	// Create temporary directory structure for consistent benchmarking
	tempDir := b.TempDir()
	gitDir := filepath.Join(tempDir, ".git")
	err = os.Mkdir(gitDir, 0755)
	if err != nil {
		b.Fatal(err)
	}

	vcsDir := filepath.Join(tempDir, "vcs")
	err = os.Mkdir(vcsDir, 0755)
	if err != nil {
		b.Fatal(err)
	}

	// Change to test directory for consistent results
	originalDir, err := os.Getwd()
	if err != nil {
		b.Fatal(err)
	}
	defer os.Chdir(originalDir)

	err = os.Chdir(vcsDir)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx, _ := detector.Detect()
		_ = ctx
	}
}

// BenchmarkContainerStatus_Creation benchmarks container status creation
func BenchmarkContainerStatus_Creation(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		status := ContainerStatus{
			Name:   "test-container",
			Status: "running",
			Health: "healthy",
			Ports:  []string{"8080:80", "443:443"},
		}
		_ = status
	}
}

// BenchmarkDevelopmentMode_String benchmarks development mode string conversion
func BenchmarkDevelopmentMode_String(b *testing.B) {
	mode := ModeMultiWorktree

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		str := string(mode)
		_ = str
	}
}

// BenchmarkLocationType_String benchmarks location type string conversion
func BenchmarkLocationType_String(b *testing.B) {
	locType := LocationWorktree

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		str := string(locType)
		_ = str
	}
}
