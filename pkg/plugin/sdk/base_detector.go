package sdk

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// BaseFrameworkDetector provides base implementation for framework detection
type BaseFrameworkDetector struct {
	patterns DetectionPatterns
	commands map[string]CommandDefinition
	info     FrameworkInfo
}

// NewBaseFrameworkDetector creates a new base detector
func NewBaseFrameworkDetector(info FrameworkInfo) *BaseFrameworkDetector {
	return &BaseFrameworkDetector{
		info:     info,
		commands: make(map[string]CommandDefinition),
	}
}

// SetPatterns sets the detection patterns
func (d *BaseFrameworkDetector) SetPatterns(patterns DetectionPatterns) {
	d.patterns = patterns
}

// SetCommands sets the default commands
func (d *BaseFrameworkDetector) SetCommands(commands map[string]CommandDefinition) {
	d.commands = commands
}

// GetDetectionPatterns returns the detection patterns
func (d *BaseFrameworkDetector) GetDetectionPatterns() DetectionPatterns {
	return d.patterns
}

// GetDefaultCommands returns the default commands
func (d *BaseFrameworkDetector) GetDefaultCommands() map[string]CommandDefinition {
	return d.commands
}

// Detect performs basic framework detection
func (d *BaseFrameworkDetector) Detect(projectPath string) (*DetectionResult, error) {
	confidence := 0
	maxConfidence := 0

	// Check required files (each worth 20 points)
	for _, file := range d.patterns.RequiredFiles {
		maxConfidence += 20
		if d.fileExists(projectPath, file) {
			confidence += 20
		} else {
			// If required file doesn't exist, not detected
			return &DetectionResult{Detected: false}, nil
		}
	}

	// Check optional files (each worth 10 points)
	for _, file := range d.patterns.OptionalFiles {
		maxConfidence += 10
		if d.fileExists(projectPath, file) {
			confidence += 10
		}
	}

	// Check directories (each worth 10 points)
	for _, dir := range d.patterns.Directories {
		maxConfidence += 10
		if d.dirExists(projectPath, dir) {
			confidence += 10
		}
	}

	// Check file contents (each worth 15 points)
	for _, pattern := range d.patterns.FileContents {
		maxConfidence += 15
		if d.checkFileContent(projectPath, pattern) {
			confidence += 15
		}
	}

	// Check extensions (worth 5 points if any found)
	if len(d.patterns.Extensions) > 0 {
		maxConfidence += 5
		if d.hasFileWithExtension(projectPath, d.patterns.Extensions) {
			confidence += 5
		}
	}

	// Calculate percentage confidence
	if maxConfidence > 0 {
		confidence = (confidence * 100) / maxConfidence
	}

	// Need at least 50% confidence
	if confidence < 50 {
		return &DetectionResult{Detected: false}, nil
	}

	// Build commands map
	commands := make(map[string]string)
	for name, def := range d.commands {
		commands[name] = def.Cmd
	}

	return &DetectionResult{
		Detected:   true,
		Confidence: confidence,
		Framework:  d.info,
		Commands:   commands,
		Metadata:   make(map[string]string),
	}, nil
}

// EnhanceContext adds framework info to context
func (d *BaseFrameworkDetector) EnhanceContext(ctx map[string]interface{}) error {
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}

	// Add framework to detected frameworks list
	frameworks, ok := ctx["frameworks"].([]string)
	if !ok {
		frameworks = []string{}
	}
	frameworks = append(frameworks, d.info.Name)
	ctx["frameworks"] = frameworks

	// Add framework version if available
	if d.info.Version != "" {
		versions, ok := ctx["framework_versions"].(map[string]string)
		if !ok {
			versions = make(map[string]string)
		}
		versions[d.info.Name] = d.info.Version
		ctx["framework_versions"] = versions
	}

	return nil
}

// Helper methods

func (d *BaseFrameworkDetector) fileExists(projectPath, filename string) bool {
	path := filepath.Join(projectPath, filename)
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func (d *BaseFrameworkDetector) dirExists(projectPath, dirname string) bool {
	path := filepath.Join(projectPath, dirname)
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func (d *BaseFrameworkDetector) checkFileContent(projectPath string, pattern ContentPattern) bool {
	path := filepath.Join(projectPath, pattern.Filepath)
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	contentStr := string(content)

	// Check for any of the contains strings
	for _, substr := range pattern.Contains {
		if strings.Contains(contentStr, substr) {
			return true
		}
	}

	// Check regex if provided
	if pattern.Regex != "" {
		matched, err := regexp.MatchString(pattern.Regex, contentStr)
		if err == nil && matched {
			return true
		}
	}

	return false
}

func (d *BaseFrameworkDetector) hasFileWithExtension(projectPath string, extensions []string) bool {
	for _, ext := range extensions {
		pattern := filepath.Join(projectPath, "*"+ext)
		matches, err := filepath.Glob(pattern)
		if err == nil && len(matches) > 0 {
			return true
		}
	}
	return false
}
