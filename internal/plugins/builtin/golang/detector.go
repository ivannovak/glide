package golang

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ivannovak/glide/pkg/plugin/sdk"
)

// GoDetector detects Go projects
type GoDetector struct {
	*sdk.BaseFrameworkDetector
}

// NewGoDetector creates a new Go detector
func NewGoDetector() *GoDetector {
	detector := &GoDetector{
		BaseFrameworkDetector: sdk.NewBaseFrameworkDetector(sdk.FrameworkInfo{
			Name: "go",
			Type: "language",
		}),
	}

	// Set detection patterns
	detector.SetPatterns(sdk.DetectionPatterns{
		RequiredFiles: []string{"go.mod"},
		OptionalFiles: []string{"go.sum", "go.work"},
		Directories:   []string{"vendor"},
		Extensions:    []string{".go"},
	})

	// Set default commands
	detector.SetCommands(map[string]sdk.CommandDefinition{
		"build": {
			Cmd:         "go build ./...",
			Description: "Build Go project",
			Category:    "build",
		},
		"test": {
			Cmd:         "go test ./...",
			Description: "Run Go tests",
			Category:    "test",
		},
		"test:v": {
			Cmd:         "go test -v ./...",
			Description: "Run Go tests with verbose output",
			Category:    "test",
		},
		"test:race": {
			Cmd:         "go test -race ./...",
			Description: "Run Go tests with race detector",
			Category:    "test",
		},
		"test:cover": {
			Cmd:         "go test -cover ./...",
			Description: "Run Go tests with coverage",
			Category:    "test",
		},
		"run": {
			Cmd:         "go run .",
			Description: "Run Go application",
			Category:    "run",
		},
		"fmt": {
			Cmd:         "go fmt ./...",
			Description: "Format Go code",
			Category:    "format",
		},
		"vet": {
			Cmd:         "go vet ./...",
			Description: "Examine Go source code",
			Category:    "lint",
		},
		"mod:tidy": {
			Cmd:         "go mod tidy",
			Description: "Add missing and remove unused modules",
			Category:    "dependencies",
		},
		"mod:download": {
			Cmd:         "go mod download",
			Description: "Download modules to local cache",
			Category:    "dependencies",
		},
		"mod:vendor": {
			Cmd:         "go mod vendor",
			Description: "Make vendored copy of dependencies",
			Category:    "dependencies",
		},
		"generate": {
			Cmd:         "go generate ./...",
			Description: "Generate Go files",
			Category:    "build",
		},
	})

	return detector
}

// Detect performs Go-specific detection
func (d *GoDetector) Detect(projectPath string) (*sdk.DetectionResult, error) {
	// First use base detection
	result, err := d.BaseFrameworkDetector.Detect(projectPath)
	if err != nil || !result.Detected {
		return result, err
	}

	// Enhance with Go-specific detection
	goModPath := filepath.Join(projectPath, "go.mod")
	if version, err := d.detectGoVersion(goModPath); err == nil {
		result.Framework.Version = version
		result.Metadata["module"] = d.detectModuleName(goModPath)
		result.Metadata["go_version"] = version
	}

	// Check for workspace
	if _, err := os.Stat(filepath.Join(projectPath, "go.work")); err == nil {
		result.Metadata["workspace"] = "true"
	}

	// Check for common Go tools and adjust confidence
	if d.hasGoTools(projectPath) {
		result.Confidence = min(100, result.Confidence+10)
	}

	return result, nil
}

// detectGoVersion extracts Go version from go.mod
func (d *GoDetector) detectGoVersion(goModPath string) (string, error) {
	file, err := os.Open(goModPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "go ") {
			return strings.TrimPrefix(line, "go "), nil
		}
	}

	return "", fmt.Errorf("go version not found in go.mod")
}

// detectModuleName extracts module name from go.mod
func (d *GoDetector) detectModuleName(goModPath string) string {
	file, err := os.Open(goModPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimPrefix(line, "module ")
		}
	}

	return ""
}

// hasGoTools checks for common Go development tools
func (d *GoDetector) hasGoTools(projectPath string) bool {
	toolFiles := []string{
		".golangci.yml",
		".golangci.yaml",
		".goreleaser.yml",
		".goreleaser.yaml",
		"Makefile", // Often used for Go projects
	}

	for _, file := range toolFiles {
		if _, err := os.Stat(filepath.Join(projectPath, file)); err == nil {
			return true
		}
	}

	return false
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
