package sdk

// FrameworkDetector interface for plugins that detect frameworks
type FrameworkDetector interface {
	// GetDetectionPatterns returns patterns this plugin uses for detection
	GetDetectionPatterns() DetectionPatterns

	// Detect performs detection and returns confidence score (0-100)
	Detect(projectPath string) (*DetectionResult, error)

	// GetDefaultCommands returns commands to inject when detected
	GetDefaultCommands() map[string]CommandDefinition

	// EnhanceContext adds framework-specific context
	EnhanceContext(ctx map[string]interface{}) error
}

// DetectionPatterns defines what to look for
type DetectionPatterns struct {
	// Files that must exist
	RequiredFiles []string `json:"required_files"`

	// Files that might exist
	OptionalFiles []string `json:"optional_files"`

	// Directory patterns
	Directories []string `json:"directories"`

	// File content patterns
	FileContents []ContentPattern `json:"file_contents"`

	// File extension patterns
	Extensions []string `json:"extensions"`
}

// ContentPattern defines patterns to search for in file contents
type ContentPattern struct {
	Filepath string   `json:"filepath"`
	Contains []string `json:"contains"` // Any of these
	Regex    string   `json:"regex"`    // Or match this regex
}

// DetectionResult contains the result of framework detection
type DetectionResult struct {
	Detected   bool              `json:"detected"`
	Confidence int               `json:"confidence"` // 0-100
	Framework  FrameworkInfo     `json:"framework"`
	Commands   map[string]string `json:"commands"`
	Metadata   map[string]string `json:"metadata"`
}

// FrameworkInfo describes a detected framework
type FrameworkInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Type    string `json:"type"` // language|framework|tool
}

// CommandDefinition defines a command provided by a framework
type CommandDefinition struct {
	Cmd         string            `json:"cmd"`
	Description string            `json:"description"`
	Alias       string            `json:"alias,omitempty"`
	Category    string            `json:"category,omitempty"`
	Args        []string          `json:"args,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
}
