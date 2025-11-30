package sdk

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// SecurityValidator validates plugin security and integrity
type SecurityValidator struct {
	trustedSources []string
	checksums      map[string]string // plugin path -> checksum
	signatures     map[string]string // plugin path -> signature
}

// NewSecurityValidator creates a new security validator
func NewSecurityValidator(trustedSources []string) *SecurityValidator {
	return &SecurityValidator{
		trustedSources: trustedSources,
		checksums:      make(map[string]string),
		signatures:     make(map[string]string),
	}
}

// AddTrustedSource adds a trusted source to the validator
func (sv *SecurityValidator) AddTrustedSource(source string) {
	sv.trustedSources = append(sv.trustedSources, source)
}

// SetTrustedSources replaces all trusted sources
func (sv *SecurityValidator) SetTrustedSources(sources []string) {
	sv.trustedSources = sources
}

// ValidatePlugin performs comprehensive plugin validation
func (sv *SecurityValidator) ValidatePlugin(pluginPath string, manifest *PluginManifest) error {
	// 1. File system security checks
	if err := sv.validateFileSystem(pluginPath); err != nil {
		return fmt.Errorf("filesystem validation failed: %w", err)
	}

	// 2. Checksum verification
	if err := sv.validateChecksum(pluginPath, manifest); err != nil {
		return fmt.Errorf("checksum validation failed: %w", err)
	}

	// 3. Source validation (if manifest available)
	if manifest != nil {
		if err := sv.validateSource(manifest); err != nil {
			return fmt.Errorf("source validation failed: %w", err)
		}
	}

	// 4. Binary analysis (basic)
	if err := sv.validateBinary(pluginPath); err != nil {
		return fmt.Errorf("binary validation failed: %w", err)
	}

	return nil
}

// validateFileSystem checks file system security
func (sv *SecurityValidator) validateFileSystem(pluginPath string) error {
	info, err := os.Stat(pluginPath)
	if err != nil {
		return fmt.Errorf("cannot stat plugin file: %w", err)
	}

	// Plugin must be executable
	if info.Mode()&0111 == 0 {
		return fmt.Errorf("plugin file is not executable")
	}

	// Plugin should not be writable by group or others (security risk)
	if info.Mode()&0022 != 0 {
		return fmt.Errorf("plugin file is writable by group or others (security risk)")
	}

	// Plugin should be owned by current user or root (Unix-specific)
	if err := sv.validateOwnership(info); err != nil {
		return err
	}

	// Plugin should not be in world-writable directories
	dir := filepath.Dir(pluginPath)
	dirInfo, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("cannot stat plugin directory: %w", err)
	}

	if dirInfo.Mode()&0002 != 0 {
		return fmt.Errorf("plugin is in world-writable directory (security risk)")
	}

	return nil
}

// validateOwnership checks file ownership (platform-specific implementation in security_*.go)

// validateChecksum verifies plugin integrity via checksum
func (sv *SecurityValidator) validateChecksum(pluginPath string, manifest *PluginManifest) error {
	// Calculate actual checksum
	actualChecksum, err := sv.calculateChecksum(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	// Check against manifest
	if manifest != nil && manifest.Spec.Executable.Checksum != "" {
		expectedChecksum := strings.TrimPrefix(manifest.Spec.Executable.Checksum, "sha256:")
		if actualChecksum != expectedChecksum {
			return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
		}
	}

	// Store checksum for future reference
	sv.checksums[pluginPath] = actualChecksum

	return nil
}

// validateSource checks if plugin comes from trusted source
func (sv *SecurityValidator) validateSource(manifest *PluginManifest) error {
	homepage := manifest.Metadata.Homepage
	if homepage == "" {
		return fmt.Errorf("plugin manifest missing homepage")
	}

	// If no trusted sources are configured, skip validation
	if len(sv.trustedSources) == 0 {
		return nil
	}

	// Check against trusted sources
	for _, trusted := range sv.trustedSources {
		if strings.Contains(homepage, trusted) {
			return nil // Found trusted source
		}
	}

	// Not from trusted source - this might be OK but should be flagged
	fmt.Printf("Warning: plugin from untrusted source: %s\n", homepage)
	return nil
}

// validateBinary performs basic binary analysis
func (sv *SecurityValidator) validateBinary(pluginPath string) error {
	file, err := os.Open(pluginPath)
	if err != nil {
		return fmt.Errorf("cannot open plugin file: %w", err)
	}
	defer file.Close()

	// Read first few bytes to check file type
	header := make([]byte, 16)
	_, err = file.Read(header)
	if err != nil {
		return fmt.Errorf("cannot read plugin header: %w", err)
	}

	// Check for ELF magic number (Linux) or Mach-O (macOS)
	if len(header) >= 4 {
		if header[0] == 0x7f && header[1] == 'E' && header[2] == 'L' && header[3] == 'F' {
			// ELF binary (Linux)
			return nil
		}
		if header[0] == 0xcf && header[1] == 0xfa && header[2] == 0xed && header[3] == 0xfe {
			// Mach-O binary (macOS, 64-bit)
			return nil
		}
		if header[0] == 0xce && header[1] == 0xfa && header[2] == 0xed && header[3] == 0xfe {
			// Mach-O binary (macOS, 32-bit)
			return nil
		}
		if header[0] == 'M' && header[1] == 'Z' {
			// PE binary (Windows)
			return nil
		}
	}

	return fmt.Errorf("unrecognized binary format")
}

// calculateChecksum calculates SHA256 checksum of a file
func (sv *SecurityValidator) calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// CapabilityValidator validates plugin capability requests
type CapabilityValidator struct {
	allowedCapabilities map[string]bool
	restrictedPaths     []string
}

// NewCapabilityValidator creates a new capability validator
func NewCapabilityValidator() *CapabilityValidator {
	return &CapabilityValidator{
		allowedCapabilities: map[string]bool{
			"docker":     true,
			"network":    false, // Restricted by default
			"filesystem": false, // Restricted by default
		},
		restrictedPaths: []string{
			"/etc/",
			"/bin/",
			"/sbin/",
			"/usr/bin/",
			"/usr/sbin/",
			"/home/",
		},
	}
}

// ValidateCapabilities checks if plugin capability requests are allowed
func (cv *CapabilityValidator) ValidateCapabilities(capabilities *Capabilities) error {
	// Check Docker capability
	if capabilities.RequiresDocker && !cv.allowedCapabilities["docker"] {
		return fmt.Errorf("docker capability not allowed")
	}

	// Check Network capability
	if capabilities.RequiresNetwork && !cv.allowedCapabilities["network"] {
		return fmt.Errorf("network capability not allowed")
	}

	// Check Filesystem capability
	if capabilities.RequiresFilesystem && !cv.allowedCapabilities["filesystem"] {
		return fmt.Errorf("filesystem capability not allowed")
	}

	// Check required paths
	for _, path := range capabilities.RequiredPaths {
		if err := cv.validatePath(path); err != nil {
			return fmt.Errorf("path %s not allowed: %w", path, err)
		}
	}

	// Check required commands
	for _, cmd := range capabilities.RequiredCommands {
		if err := cv.validateCommand(cmd); err != nil {
			return fmt.Errorf("command %s not allowed: %w", cmd, err)
		}
	}

	return nil
}

// validatePath checks if a path is allowed
func (cv *CapabilityValidator) validatePath(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Check against restricted paths
	for _, restricted := range cv.restrictedPaths {
		if strings.HasPrefix(absPath, restricted) {
			return fmt.Errorf("access to %s is restricted", restricted)
		}
	}

	return nil
}

// validateCommand checks if a command is allowed
func (cv *CapabilityValidator) validateCommand(command string) error {
	allowedCommands := []string{
		"docker",
		"git",
		"aws",
		"kubectl",
		"curl",
		"wget",
	}

	for _, allowed := range allowedCommands {
		if command == allowed {
			return nil
		}
	}

	return fmt.Errorf("command %s is not in allowed list", command)
}

// Capabilities represents plugin capability requirements
type Capabilities struct {
	RequiresDocker     bool     `json:"requires_docker"`
	RequiresNetwork    bool     `json:"requires_network"`
	RequiresFilesystem bool     `json:"requires_filesystem"`
	RequiredPaths      []string `json:"required_paths"`
	RequiredCommands   []string `json:"required_commands"`
	RequiredEnvVars    []string `json:"required_env_vars"`
	RequiredConfig     []string `json:"required_config"`
}

// PluginManifest represents a plugin manifest file
type PluginManifest struct {
	APIVersion string       `yaml:"apiVersion"`
	Kind       string       `yaml:"kind"`
	Metadata   ManifestMeta `yaml:"metadata"`
	Spec       ManifestSpec `yaml:"spec"`
}

// ManifestMeta contains plugin metadata
type ManifestMeta struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Author      string `yaml:"author"`
	Description string `yaml:"description"`
	Homepage    string `yaml:"homepage"`
	License     string `yaml:"license"`
}

// ManifestSpec contains plugin specification
type ManifestSpec struct {
	Executable   ExecutableSpec `yaml:"executable"`
	Commands     []CommandSpec  `yaml:"commands"`
	Capabilities Capabilities   `yaml:"capabilities"`
	Config       ConfigSpec     `yaml:"config"`
}

// ExecutableSpec contains executable information
type ExecutableSpec struct {
	Name     string `yaml:"name"`
	Checksum string `yaml:"checksum"`
}

// CommandSpec contains command specification
type CommandSpec struct {
	Name        string `yaml:"name"`
	Category    string `yaml:"category"`
	Description string `yaml:"description"`
	Interactive bool   `yaml:"interactive"`
}

// ConfigSpec contains configuration specification
type ConfigSpec map[string]interface{}

// SandboxValidator validates plugin sandboxing
type SandboxValidator struct {
	allowedSyscalls []string
	blockedSyscalls []string
}

// NewSandboxValidator creates a new sandbox validator
func NewSandboxValidator() *SandboxValidator {
	return &SandboxValidator{
		allowedSyscalls: []string{
			"read", "write", "open", "close",
			"execve", "fork", "clone",
		},
		blockedSyscalls: []string{
			"ptrace", "mount", "umount",
			"reboot", "kexec_load",
		},
	}
}

// ValidateSandbox checks plugin sandboxing (placeholder for advanced implementation)
func (sv *SandboxValidator) ValidateSandbox(pluginPath string) error {
	// This would implement advanced sandboxing validation
	// For now, it's a placeholder that always succeeds
	return nil
}
