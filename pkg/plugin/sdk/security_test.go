package sdk

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNewSecurityValidator tests security validator creation
func TestNewSecurityValidator(t *testing.T) {
	trustedSources := []string{"github.com/ivannovak", "gitlab.com/myorg"}

	sv := NewSecurityValidator(trustedSources)

	if sv == nil {
		t.Fatal("NewSecurityValidator returned nil")
	}

	if len(sv.trustedSources) != len(trustedSources) {
		t.Errorf("trustedSources length = %d, want %d", len(sv.trustedSources), len(trustedSources))
	}

	if sv.checksums == nil {
		t.Error("checksums map not initialized")
	}

	if sv.signatures == nil {
		t.Error("signatures map not initialized")
	}
}

// TestAddTrustedSource tests adding trusted sources
func TestAddTrustedSource(t *testing.T) {
	sv := NewSecurityValidator(nil)

	sv.AddTrustedSource("github.com/trusted")

	if len(sv.trustedSources) != 1 {
		t.Errorf("expected 1 trusted source, got %d", len(sv.trustedSources))
	}

	if sv.trustedSources[0] != "github.com/trusted" {
		t.Errorf("trustedSources[0] = %s, want github.com/trusted", sv.trustedSources[0])
	}
}

// TestSetTrustedSources tests replacing trusted sources
func TestSetTrustedSources(t *testing.T) {
	sv := NewSecurityValidator([]string{"old-source"})

	newSources := []string{"source1", "source2", "source3"}
	sv.SetTrustedSources(newSources)

	if len(sv.trustedSources) != len(newSources) {
		t.Errorf("expected %d trusted sources, got %d", len(newSources), len(sv.trustedSources))
	}

	for i, source := range newSources {
		if sv.trustedSources[i] != source {
			t.Errorf("trustedSources[%d] = %s, want %s", i, sv.trustedSources[i], source)
		}
	}
}

// TestValidatePlugin_FileNotFound tests validation with missing plugin
func TestValidatePlugin_FileNotFound(t *testing.T) {
	sv := NewSecurityValidator(nil)

	err := sv.ValidatePlugin("/nonexistent/plugin", nil)

	if err == nil {
		t.Fatal("expected error for non-existent plugin, got nil")
	}
}

// TestValidatePlugin_ValidPlugin tests successful validation
func TestValidatePlugin_ValidPlugin(t *testing.T) {
	tmpDir := t.TempDir()
	pluginPath := filepath.Join(tmpDir, "test-plugin")

	// Create valid ELF binary with proper permissions
	elfHeader := []byte{0x7f, 'E', 'L', 'F', 0x02, 0x01, 0x01, 0x00}
	if err := os.WriteFile(pluginPath, elfHeader, 0755); err != nil {
		t.Fatalf("failed to create test plugin: %v", err)
	}

	// Make directory not world-writable
	if err := os.Chmod(tmpDir, 0755); err != nil {
		t.Fatalf("failed to set directory permissions: %v", err)
	}

	sv := NewSecurityValidator(nil)

	err := sv.ValidatePlugin(pluginPath, nil)
	if err != nil {
		t.Errorf("expected no error for valid plugin, got: %v", err)
	}
}

// TestValidateFileSystem_NotExecutable tests file system validation for non-executable
func TestValidateFileSystem_NotExecutable(t *testing.T) {
	tmpDir := t.TempDir()
	pluginPath := filepath.Join(tmpDir, "not-executable")

	// Create non-executable file
	if err := os.WriteFile(pluginPath, []byte{0x7f, 'E', 'L', 'F'}, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	sv := NewSecurityValidator(nil)

	err := sv.validateFileSystem(pluginPath)
	if err == nil {
		t.Fatal("expected error for non-executable file, got nil")
	}
}

// TestValidateFileSystem_WorldWritable tests world-writable file detection
func TestValidateFileSystem_WorldWritable(t *testing.T) {
	tmpDir := t.TempDir()
	pluginPath := filepath.Join(tmpDir, "world-writable")

	// Create world-writable file
	if err := os.WriteFile(pluginPath, []byte{0x7f, 'E', 'L', 'F'}, 0777); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Verify permissions were actually set (may not work on all filesystems)
	info, _ := os.Stat(pluginPath)
	if info.Mode()&0022 == 0 {
		t.Skip("filesystem doesn't support world-writable permissions")
	}

	sv := NewSecurityValidator(nil)

	err := sv.validateFileSystem(pluginPath)
	if err == nil {
		t.Fatal("expected error for world-writable file, got nil")
	}

	if err.Error() != "plugin file is writable by group or others (security risk)" {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestValidateFileSystem_GroupWritable tests group-writable file detection
func TestValidateFileSystem_GroupWritable(t *testing.T) {
	tmpDir := t.TempDir()
	pluginPath := filepath.Join(tmpDir, "group-writable")

	// Create group-writable file
	if err := os.WriteFile(pluginPath, []byte{0x7f, 'E', 'L', 'F'}, 0775); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Verify permissions were actually set (may not work on all filesystems)
	info, _ := os.Stat(pluginPath)
	if info.Mode()&0020 == 0 {
		t.Skip("filesystem doesn't support group-writable permissions")
	}

	sv := NewSecurityValidator(nil)

	err := sv.validateFileSystem(pluginPath)
	if err == nil {
		t.Fatal("expected error for group-writable file, got nil")
	}
}

// TestValidateFileSystem_WorldWritableDirectory tests world-writable directory detection
func TestValidateFileSystem_WorldWritableDirectory(t *testing.T) {
	// Create world-writable directory
	tmpDir := filepath.Join(os.TempDir(), "world-writable-dir")
	if err := os.MkdirAll(tmpDir, 0777); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Verify directory permissions were actually set (may not work on all filesystems)
	dirInfo, _ := os.Stat(tmpDir)
	if dirInfo.Mode()&0002 == 0 {
		t.Skip("filesystem doesn't support world-writable directory permissions")
	}

	pluginPath := filepath.Join(tmpDir, "plugin")
	if err := os.WriteFile(pluginPath, []byte{0x7f, 'E', 'L', 'F'}, 0755); err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	sv := NewSecurityValidator(nil)

	err := sv.validateFileSystem(pluginPath)
	if err == nil {
		t.Fatal("expected error for plugin in world-writable directory, got nil")
	}
}

// TestValidateChecksum_Success tests successful checksum validation
func TestValidateChecksum_Success(t *testing.T) {
	tmpDir := t.TempDir()
	pluginPath := filepath.Join(tmpDir, "test-plugin")

	content := []byte{0x7f, 'E', 'L', 'F', 0x01, 0x02, 0x03}
	if err := os.WriteFile(pluginPath, content, 0755); err != nil {
		t.Fatalf("failed to create test plugin: %v", err)
	}

	sv := NewSecurityValidator(nil)

	// Calculate expected checksum
	expectedChecksum, err := sv.calculateChecksum(pluginPath)
	if err != nil {
		t.Fatalf("failed to calculate checksum: %v", err)
	}

	manifest := &PluginManifest{
		Spec: ManifestSpec{
			Executable: ExecutableSpec{
				Checksum: "sha256:" + expectedChecksum,
			},
		},
	}

	err = sv.validateChecksum(pluginPath, manifest)
	if err != nil {
		t.Errorf("expected no error with correct checksum, got: %v", err)
	}

	// Verify checksum was stored
	if sv.checksums[pluginPath] != expectedChecksum {
		t.Error("checksum not stored correctly")
	}
}

// TestValidateChecksum_Mismatch tests checksum mismatch detection
func TestValidateChecksum_Mismatch(t *testing.T) {
	tmpDir := t.TempDir()
	pluginPath := filepath.Join(tmpDir, "test-plugin")

	content := []byte{0x7f, 'E', 'L', 'F'}
	if err := os.WriteFile(pluginPath, content, 0755); err != nil {
		t.Fatalf("failed to create test plugin: %v", err)
	}

	sv := NewSecurityValidator(nil)

	manifest := &PluginManifest{
		Spec: ManifestSpec{
			Executable: ExecutableSpec{
				Checksum: "sha256:deadbeef",
			},
		},
	}

	err := sv.validateChecksum(pluginPath, manifest)
	if err == nil {
		t.Fatal("expected error for checksum mismatch, got nil")
	}
}

// TestValidateChecksum_NoManifest tests checksum validation without manifest
func TestValidateChecksum_NoManifest(t *testing.T) {
	tmpDir := t.TempDir()
	pluginPath := filepath.Join(tmpDir, "test-plugin")

	content := []byte{0x7f, 'E', 'L', 'F'}
	if err := os.WriteFile(pluginPath, content, 0755); err != nil {
		t.Fatalf("failed to create test plugin: %v", err)
	}

	sv := NewSecurityValidator(nil)

	err := sv.validateChecksum(pluginPath, nil)
	if err != nil {
		t.Errorf("expected no error without manifest, got: %v", err)
	}
}

// TestValidateSource_TrustedSource tests trusted source validation
func TestValidateSource_TrustedSource(t *testing.T) {
	trustedSources := []string{"github.com/ivannovak", "gitlab.com/myorg"}
	sv := NewSecurityValidator(trustedSources)

	manifest := &PluginManifest{
		Metadata: ManifestMeta{
			Homepage: "https://github.com/ivannovak/my-plugin",
		},
	}

	err := sv.validateSource(manifest)
	if err != nil {
		t.Errorf("expected no error for trusted source, got: %v", err)
	}
}

// TestValidateSource_UntrustedSource tests untrusted source handling
func TestValidateSource_UntrustedSource(t *testing.T) {
	trustedSources := []string{"github.com/ivannovak"}
	sv := NewSecurityValidator(trustedSources)

	manifest := &PluginManifest{
		Metadata: ManifestMeta{
			Homepage: "https://example.com/untrusted-plugin",
		},
	}

	// Should not error, but should warn (captured in logs)
	err := sv.validateSource(manifest)
	if err != nil {
		t.Errorf("expected no error for untrusted source (warning only), got: %v", err)
	}
}

// TestValidateSource_NoHomepage tests missing homepage handling
func TestValidateSource_NoHomepage(t *testing.T) {
	sv := NewSecurityValidator([]string{"github.com/ivannovak"})

	manifest := &PluginManifest{
		Metadata: ManifestMeta{
			Homepage: "",
		},
	}

	err := sv.validateSource(manifest)
	if err == nil {
		t.Fatal("expected error for missing homepage, got nil")
	}
}

// TestValidateSource_NoTrustedSources tests validation with no trusted sources configured
func TestValidateSource_NoTrustedSources(t *testing.T) {
	sv := NewSecurityValidator(nil)

	manifest := &PluginManifest{
		Metadata: ManifestMeta{
			Homepage: "https://example.com/plugin",
		},
	}

	err := sv.validateSource(manifest)
	if err != nil {
		t.Errorf("expected no error when no trusted sources configured, got: %v", err)
	}
}

// TestValidateBinary_ValidFormats tests various valid binary formats
func TestValidateBinary_ValidFormats(t *testing.T) {
	tests := []struct {
		name   string
		header []byte
	}{
		{"ELF binary", []byte{0x7f, 'E', 'L', 'F', 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"Mach-O 64-bit (little endian)", []byte{0xcf, 0xfa, 0xed, 0xfe, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"Mach-O 32-bit (little endian)", []byte{0xce, 0xfa, 0xed, 0xfe, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"PE binary", []byte{'M', 'Z', 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	}

	sv := NewSecurityValidator(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := filepath.Join(t.TempDir(), "test-binary")

			if err := os.WriteFile(tmpFile, tt.header, 0755); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			err := sv.validateBinary(tmpFile)
			if err != nil {
				t.Errorf("expected no error for %s, got: %v", tt.name, err)
			}
		})
	}
}

// TestValidateBinary_InvalidFormat tests invalid binary format detection
func TestValidateBinary_InvalidFormat(t *testing.T) {
	sv := NewSecurityValidator(nil)

	tmpFile := filepath.Join(t.TempDir(), "invalid-binary")
	content := []byte("this is not a valid binary format at all")

	if err := os.WriteFile(tmpFile, content, 0755); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	err := sv.validateBinary(tmpFile)
	if err == nil {
		t.Fatal("expected error for invalid binary format, got nil")
	}

	if err.Error() != "unrecognized binary format" {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestCalculateChecksum tests checksum calculation
func TestCalculateChecksum(t *testing.T) {
	sv := NewSecurityValidator(nil)

	tmpFile := filepath.Join(t.TempDir(), "test-file")
	content := []byte("test content for checksum calculation")

	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	checksum1, err := sv.calculateChecksum(tmpFile)
	if err != nil {
		t.Fatalf("calculateChecksum failed: %v", err)
	}

	if checksum1 == "" {
		t.Error("expected non-empty checksum")
	}

	// Verify consistency
	checksum2, err := sv.calculateChecksum(tmpFile)
	if err != nil {
		t.Fatalf("calculateChecksum failed on second call: %v", err)
	}

	if checksum1 != checksum2 {
		t.Errorf("checksums don't match: %s != %s", checksum1, checksum2)
	}
}

// TestCalculateChecksum_FileNotFound tests checksum calculation with missing file
func TestCalculateChecksum_FileNotFound(t *testing.T) {
	sv := NewSecurityValidator(nil)

	_, err := sv.calculateChecksum("/nonexistent/file")
	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}
}

// TestNewCapabilityValidator tests capability validator creation
func TestNewCapabilityValidator(t *testing.T) {
	cv := NewCapabilityValidator()

	if cv == nil {
		t.Fatal("NewCapabilityValidator returned nil")
	}

	if cv.allowedCapabilities == nil {
		t.Error("allowedCapabilities map not initialized")
	}

	if cv.restrictedPaths == nil {
		t.Error("restrictedPaths slice not initialized")
	}

	// Check default capabilities
	if !cv.allowedCapabilities["docker"] {
		t.Error("docker capability should be allowed by default")
	}

	if cv.allowedCapabilities["network"] {
		t.Error("network capability should be restricted by default")
	}

	if cv.allowedCapabilities["filesystem"] {
		t.Error("filesystem capability should be restricted by default")
	}
}

// TestValidateCapabilities_AllowedCapabilities tests allowed capability validation
func TestValidateCapabilities_AllowedCapabilities(t *testing.T) {
	cv := NewCapabilityValidator()

	caps := &Capabilities{
		RequiresDocker: true,
	}

	err := cv.ValidateCapabilities(caps)
	if err != nil {
		t.Errorf("expected no error for allowed docker capability, got: %v", err)
	}
}

// TestValidateCapabilities_RestrictedCapabilities tests restricted capability detection
func TestValidateCapabilities_RestrictedCapabilities(t *testing.T) {
	cv := NewCapabilityValidator()

	tests := []struct {
		name string
		caps *Capabilities
	}{
		{"network", &Capabilities{RequiresNetwork: true}},
		{"filesystem", &Capabilities{RequiresFilesystem: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cv.ValidateCapabilities(tt.caps)
			if err == nil {
				t.Fatalf("expected error for restricted %s capability, got nil", tt.name)
			}
		})
	}
}

// TestValidateCapabilities_RequiredPaths tests required path validation
func TestValidateCapabilities_RequiredPaths(t *testing.T) {
	cv := NewCapabilityValidator()

	// Allowed path (e.g., user's home directory or temp)
	caps := &Capabilities{
		RequiredPaths: []string{"/tmp/my-plugin-data"},
	}

	err := cv.ValidateCapabilities(caps)
	if err != nil {
		t.Errorf("expected no error for allowed path, got: %v", err)
	}
}

// TestValidateCapabilities_RestrictedPaths tests restricted path detection
func TestValidateCapabilities_RestrictedPaths(t *testing.T) {
	cv := NewCapabilityValidator()

	restrictedPaths := []string{
		"/etc/passwd",
		"/bin/bash",
		"/sbin/init",
		"/usr/bin/sudo",
	}

	for _, path := range restrictedPaths {
		t.Run(path, func(t *testing.T) {
			caps := &Capabilities{
				RequiredPaths: []string{path},
			}

			err := cv.ValidateCapabilities(caps)
			if err == nil {
				t.Fatalf("expected error for restricted path %s, got nil", path)
			}
		})
	}
}

// TestValidateCapabilities_RequiredCommands tests required command validation
func TestValidateCapabilities_RequiredCommands(t *testing.T) {
	cv := NewCapabilityValidator()

	allowedCommands := []string{"docker", "git", "aws", "kubectl", "curl", "wget"}

	for _, cmd := range allowedCommands {
		t.Run(cmd, func(t *testing.T) {
			caps := &Capabilities{
				RequiredCommands: []string{cmd},
			}

			err := cv.ValidateCapabilities(caps)
			if err != nil {
				t.Errorf("expected no error for allowed command %s, got: %v", cmd, err)
			}
		})
	}
}

// TestValidateCapabilities_DisallowedCommands tests disallowed command detection
func TestValidateCapabilities_DisallowedCommands(t *testing.T) {
	cv := NewCapabilityValidator()

	disallowedCommands := []string{"rm", "mkfs", "dd", "fdisk"}

	for _, cmd := range disallowedCommands {
		t.Run(cmd, func(t *testing.T) {
			caps := &Capabilities{
				RequiredCommands: []string{cmd},
			}

			err := cv.ValidateCapabilities(caps)
			if err == nil {
				t.Fatalf("expected error for disallowed command %s, got nil", cmd)
			}
		})
	}
}

// TestValidatePath tests path validation
func TestValidatePath(t *testing.T) {
	cv := NewCapabilityValidator()

	// Test allowed path
	err := cv.validatePath("/tmp/plugin-data")
	if err != nil {
		t.Errorf("expected no error for allowed path, got: %v", err)
	}

	// Test restricted path
	err = cv.validatePath("/etc/shadow")
	if err == nil {
		t.Fatal("expected error for restricted path, got nil")
	}
}

// TestValidateCommand tests command validation
func TestValidateCommand(t *testing.T) {
	cv := NewCapabilityValidator()

	// Test allowed command
	err := cv.validateCommand("docker")
	if err != nil {
		t.Errorf("expected no error for allowed command, got: %v", err)
	}

	// Test disallowed command
	err = cv.validateCommand("rm")
	if err == nil {
		t.Fatal("expected error for disallowed command, got nil")
	}
}

// TestNewSandboxValidator tests sandbox validator creation
func TestNewSandboxValidator(t *testing.T) {
	sv := NewSandboxValidator()

	if sv == nil {
		t.Fatal("NewSandboxValidator returned nil")
	}

	if sv.allowedSyscalls == nil {
		t.Error("allowedSyscalls not initialized")
	}

	if sv.blockedSyscalls == nil {
		t.Error("blockedSyscalls not initialized")
	}
}

// TestValidateSandbox tests sandbox validation (placeholder)
func TestValidateSandbox(t *testing.T) {
	sv := NewSandboxValidator()

	tmpFile := filepath.Join(t.TempDir(), "test-plugin")
	if err := os.WriteFile(tmpFile, []byte{0x7f, 'E', 'L', 'F'}, 0755); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Current implementation is a placeholder that always succeeds
	err := sv.ValidateSandbox(tmpFile)
	if err != nil {
		t.Errorf("expected no error (placeholder implementation), got: %v", err)
	}
}

// TestValidatePlugin_Complete tests complete validation flow
func TestValidatePlugin_Complete(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.Chmod(tmpDir, 0755); err != nil {
		t.Fatalf("failed to set directory permissions: %v", err)
	}

	pluginPath := filepath.Join(tmpDir, "complete-plugin")

	// Create valid ELF binary
	elfContent := []byte{0x7f, 'E', 'L', 'F', 0x02, 0x01, 0x01, 0x00}
	if err := os.WriteFile(pluginPath, elfContent, 0755); err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	sv := NewSecurityValidator([]string{"github.com/ivannovak"})

	// Calculate checksum
	checksum, err := sv.calculateChecksum(pluginPath)
	if err != nil {
		t.Fatalf("failed to calculate checksum: %v", err)
	}

	manifest := &PluginManifest{
		Metadata: ManifestMeta{
			Name:     "test-plugin",
			Version:  "1.0.0",
			Homepage: "https://github.com/ivannovak/test-plugin",
		},
		Spec: ManifestSpec{
			Executable: ExecutableSpec{
				Name:     "complete-plugin",
				Checksum: "sha256:" + checksum,
			},
		},
	}

	err = sv.ValidatePlugin(pluginPath, manifest)
	if err != nil {
		t.Errorf("expected no error for complete valid plugin, got: %v", err)
	}
}
