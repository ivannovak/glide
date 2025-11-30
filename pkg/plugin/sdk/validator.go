package sdk

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ivannovak/glide/v3/pkg/validation"
)

// Validator validates plugin binaries for security
type Validator struct {
	strict           bool
	trustedPaths     []string
	allowedChecksums map[string]string
}

// NewValidator creates a new plugin validator
func NewValidator(strict bool) *Validator {
	home, _ := os.UserHomeDir()
	return &Validator{
		strict: strict,
		trustedPaths: []string{
			filepath.Join(home, ".glide", "plugins"),
			"/usr/local/lib/glide/plugins",
		},
		allowedChecksums: make(map[string]string),
	}
}

// Validate checks if a plugin is safe to load
func (v *Validator) Validate(path string) error {
	// 0. Validate path to prevent directory traversal
	// Try to validate against each trusted path
	var validatedPath string
	var validationErr error

	for _, trustedPath := range v.trustedPaths {
		validated, err := validation.ValidatePath(path, validation.PathValidationOptions{
			BaseDir:        trustedPath,
			AllowAbsolute:  true, // Plugin paths can be absolute
			FollowSymlinks: true, // Follow symlinks but validate they stay within bounds
			RequireExists:  true, // Plugin must exist
		})
		if err == nil {
			validatedPath = validated
			validationErr = nil
			break
		}
		validationErr = err
	}

	// If validation failed against all trusted paths, return the last error
	if validationErr != nil {
		return fmt.Errorf("invalid plugin path: %w", validationErr)
	}

	// Use validated path for all subsequent operations
	path = validatedPath

	// 1. Check file exists (already done in path validation, but get info)
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("plugin not found: %w", err)
	}

	// 2. Check file is not a directory
	if info.IsDir() {
		return fmt.Errorf("plugin path is a directory")
	}

	// 3. Check file permissions (must be executable)
	if info.Mode()&0111 == 0 {
		return fmt.Errorf("plugin is not executable")
	}

	// 4. In strict mode, check if file is not world-writable
	if v.strict {
		if info.Mode()&0022 != 0 {
			return fmt.Errorf("plugin must not be world-writable in strict mode")
		}
	}

	// 5. Check if path is in trusted location
	if !v.isInTrustedPath(path) && v.strict {
		return fmt.Errorf("plugin is not in a trusted location")
	}

	// 6. Verify checksum if available
	if expectedChecksum, exists := v.allowedChecksums[path]; exists {
		actualChecksum, err := v.calculateChecksum(path)
		if err != nil {
			return fmt.Errorf("failed to calculate checksum: %w", err)
		}
		if actualChecksum != expectedChecksum {
			return fmt.Errorf("checksum verification failed")
		}
	}

	// 7. Basic binary validation
	if !v.isValidBinary(path) {
		return fmt.Errorf("invalid plugin binary format")
	}

	return nil
}

// AddTrustedPath adds a path to the trusted paths list
func (v *Validator) AddTrustedPath(path string) {
	v.trustedPaths = append(v.trustedPaths, path)
}

// SetChecksum sets the expected checksum for a plugin
func (v *Validator) SetChecksum(pluginPath, checksum string) {
	v.allowedChecksums[pluginPath] = checksum
}

// isInTrustedPath checks if a plugin is in a trusted directory
func (v *Validator) isInTrustedPath(pluginPath string) bool {
	absPath, err := filepath.Abs(pluginPath)
	if err != nil {
		return false
	}

	for _, trustedPath := range v.trustedPaths {
		trustedAbs, err := filepath.Abs(trustedPath)
		if err != nil {
			continue
		}

		// Check if plugin is within trusted path
		if strings.HasPrefix(absPath, trustedAbs) {
			return true
		}
	}

	return false
}

// calculateChecksum calculates SHA256 checksum of a file
func (v *Validator) calculateChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// isValidBinary performs basic binary validation
func (v *Validator) isValidBinary(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	// Read first few bytes to check for valid executable
	header := make([]byte, 4)
	if _, err := file.Read(header); err != nil {
		return false
	}

	// Check for common executable formats
	// ELF (Linux/Unix)
	if header[0] == 0x7f && header[1] == 'E' && header[2] == 'L' && header[3] == 'F' {
		return true
	}

	// Mach-O (macOS) - both 32-bit and 64-bit
	if (header[0] == 0xfe && header[1] == 0xed && header[2] == 0xfa && header[3] == 0xce) ||
		(header[0] == 0xfe && header[1] == 0xed && header[2] == 0xfa && header[3] == 0xcf) ||
		(header[0] == 0xce && header[1] == 0xfa && header[2] == 0xed && header[3] == 0xfe) ||
		(header[0] == 0xcf && header[1] == 0xfa && header[2] == 0xed && header[3] == 0xfe) {
		return true
	}

	// PE (Windows)
	if header[0] == 'M' && header[1] == 'Z' {
		return true
	}

	// Shebang scripts (#!/...)
	if header[0] == '#' && header[1] == '!' {
		return true
	}

	return false
}

// ValidateManifest validates a plugin manifest file
func (v *Validator) ValidateManifest(manifestPath string) error {
	// Check if manifest exists
	if _, err := os.Stat(manifestPath); err != nil {
		return fmt.Errorf("manifest not found: %w", err)
	}

	// In strict mode, ensure manifest is not world-writable
	if v.strict {
		info, err := os.Stat(manifestPath)
		if err != nil {
			return err
		}
		if info.Mode()&0022 != 0 {
			return fmt.Errorf("manifest must not be world-writable in strict mode")
		}
	}

	return nil
}

// SetStrict enables or disables strict mode
func (v *Validator) SetStrict(strict bool) {
	v.strict = strict
}
