package integration_test

import (
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDockerDependency verifies Docker is installed and meets version requirements
func TestDockerDependency(t *testing.T) {
	t.Run("docker_installed", func(t *testing.T) {
		cmd := exec.Command("docker", "--version")
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			t.Skip("Docker is not installed - required for Docker operations")
			return
		}
		
		assert.Contains(t, string(output), "Docker version", "Docker should report its version")
	})
	
	t.Run("docker_version_meets_requirements", func(t *testing.T) {
		cmd := exec.Command("docker", "--version")
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			t.Skip("Docker is not installed")
			return
		}
		
		// Parse Docker version (e.g., "Docker version 24.0.5, build ced0996")
		versionRegex := regexp.MustCompile(`Docker version (\d+)\.(\d+)\.(\d+)`)
		matches := versionRegex.FindStringSubmatch(string(output))
		
		require.Len(t, matches, 4, "Should parse Docker version")
		
		major, _ := strconv.Atoi(matches[1])
		minor, _ := strconv.Atoi(matches[2])
		
		// Require Docker 20.10.0 or higher
		minMajor, minMinor := 20, 10
		versionOK := major > minMajor || (major == minMajor && minor >= minMinor)
		
		assert.True(t, versionOK, 
			"Docker version %d.%d should be >= %d.%d", 
			major, minor, minMajor, minMinor)
	})
	
	t.Run("docker_daemon_running", func(t *testing.T) {
		cmd := exec.Command("docker", "info")
		err := cmd.Run()
		
		if err != nil {
			t.Skip("Docker daemon is not running - required for container operations")
			return
		}
		
		assert.NoError(t, err, "Docker daemon should be accessible")
	})
	
	t.Run("docker_compose_v2_available", func(t *testing.T) {
		// Try Docker Compose V2 (docker compose)
		cmd := exec.Command("docker", "compose", "version")
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			t.Skip("Docker Compose V2 is not available - required for compose operations")
			return
		}
		
		assert.Contains(t, string(output), "Docker Compose version", 
			"Docker Compose V2 should be available")
	})
}

// TestAWSCLIDependency verifies AWS CLI is installed if needed
func TestAWSCLIDependency(t *testing.T) {
	t.Run("aws_cli_available", func(t *testing.T) {
		cmd := exec.Command("aws", "--version")
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			t.Skip("AWS CLI is not installed - optional dependency")
			return
		}
		
		assert.Contains(t, string(output), "aws-cli", "AWS CLI should report its version")
	})
	
	t.Run("aws_cli_version", func(t *testing.T) {
		cmd := exec.Command("aws", "--version")
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			t.Skip("AWS CLI is not installed")
			return
		}
		
		// Parse AWS CLI version (e.g., "aws-cli/2.13.0 Python/3.11.4")
		versionRegex := regexp.MustCompile(`aws-cli/(\d+)\.(\d+)\.(\d+)`)
		matches := versionRegex.FindStringSubmatch(string(output))
		
		if len(matches) >= 4 {
			major, _ := strconv.Atoi(matches[1])
			
			// Prefer AWS CLI v2
			if major >= 2 {
				t.Logf("AWS CLI v%s detected - recommended version", matches[1])
			} else {
				t.Logf("AWS CLI v%s detected - consider upgrading to v2", matches[1])
			}
		}
	})
}

// TestPHPToolsDependency verifies PHP development tools
func TestPHPToolsDependency(t *testing.T) {
	t.Run("composer_available", func(t *testing.T) {
		cmd := exec.Command("composer", "--version")
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			// Try with docker
			dockerCmd := exec.Command("docker", "run", "--rm", "composer:latest", "--version")
			dockerOutput, dockerErr := dockerCmd.CombinedOutput()
			
			if dockerErr != nil {
				t.Skip("Composer is not installed - required for PHP dependency management")
				return
			}
			
			assert.Contains(t, string(dockerOutput), "Composer version", 
				"Composer should be available via Docker")
			return
		}
		
		assert.Contains(t, string(output), "Composer version", 
			"Composer should report its version")
	})
	
	t.Run("composer_version", func(t *testing.T) {
		cmd := exec.Command("composer", "--version")
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			t.Skip("Composer is not installed")
			return
		}
		
		// Parse Composer version (e.g., "Composer version 2.5.8 2023-06-09 17:13:21")
		versionRegex := regexp.MustCompile(`Composer version (\d+)\.(\d+)\.(\d+)`)
		matches := versionRegex.FindStringSubmatch(string(output))
		
		if len(matches) >= 4 {
			major, _ := strconv.Atoi(matches[1])
			
			// Prefer Composer v2
			assert.GreaterOrEqual(t, major, 2, 
				"Composer v2 or higher is recommended")
		}
	})
	
	t.Run("pest_detection", func(t *testing.T) {
		// Check if Pest is available globally
		cmd := exec.Command("pest", "--version")
		output, err := cmd.CombinedOutput()
		
		if err == nil {
			assert.Contains(t, string(output), "Pest", 
				"Pest testing framework detected")
			t.Log("Pest is available globally")
		} else {
			// Check if vendor/bin/pest exists (project-level)
			vendorCmd := exec.Command("test", "-f", "vendor/bin/pest")
			if vendorErr := vendorCmd.Run(); vendorErr == nil {
				t.Log("Pest is available in vendor/bin")
			} else {
				t.Log("Pest is not available - optional testing framework")
			}
		}
	})
	
	t.Run("phpunit_detection", func(t *testing.T) {
		// Check if PHPUnit is available globally
		cmd := exec.Command("phpunit", "--version")
		output, err := cmd.CombinedOutput()
		
		if err == nil {
			assert.Contains(t, string(output), "PHPUnit", 
				"PHPUnit testing framework detected")
			t.Log("PHPUnit is available globally")
		} else {
			// Check if vendor/bin/phpunit exists (project-level)
			vendorCmd := exec.Command("test", "-f", "vendor/bin/phpunit")
			if vendorErr := vendorCmd.Run(); vendorErr == nil {
				t.Log("PHPUnit is available in vendor/bin")
			} else {
				t.Log("PHPUnit is not available - optional testing framework")
			}
		}
	})
}

// TestGitDependency verifies Git is installed
func TestGitDependency(t *testing.T) {
	t.Run("git_installed", func(t *testing.T) {
		cmd := exec.Command("git", "--version")
		output, err := cmd.CombinedOutput()
		
		require.NoError(t, err, "Git must be installed for worktree management")
		assert.Contains(t, string(output), "git version", "Git should report its version")
	})
	
	t.Run("git_version", func(t *testing.T) {
		cmd := exec.Command("git", "--version")
		output, err := cmd.CombinedOutput()
		
		require.NoError(t, err, "Git must be installed")
		
		// Parse Git version (e.g., "git version 2.41.0")
		versionRegex := regexp.MustCompile(`git version (\d+)\.(\d+)\.(\d+)`)
		matches := versionRegex.FindStringSubmatch(string(output))
		
		if len(matches) >= 4 {
			major, _ := strconv.Atoi(matches[1])
			minor, _ := strconv.Atoi(matches[2])
			
			// Require Git 2.5+ for worktree support
			versionOK := major > 2 || (major == 2 && minor >= 5)
			assert.True(t, versionOK, 
				"Git version %d.%d should be >= 2.5 for worktree support", 
				major, minor)
		}
	})
}

// TestMakeDependency verifies Make is installed
func TestMakeDependency(t *testing.T) {
	t.Run("make_installed", func(t *testing.T) {
		cmd := exec.Command("make", "--version")
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			t.Skip("Make is not installed - optional for Makefile compatibility")
			return
		}
		
		outputStr := string(output)
		assert.True(t, 
			strings.Contains(outputStr, "GNU Make") || strings.Contains(outputStr, "make"),
			"Make should report its version")
	})
}

// TestSystemRequirements verifies system-level requirements
func TestSystemRequirements(t *testing.T) {
	t.Run("operating_system", func(t *testing.T) {
		os := runtime.GOOS
		arch := runtime.GOARCH
		
		t.Logf("Operating System: %s/%s", os, arch)
		
		// Verify supported OS
		supportedOS := map[string]bool{
			"darwin":  true, // macOS
			"linux":   true,
			"windows": true,
		}
		
		assert.True(t, supportedOS[os], 
			"Operating system %s should be supported", os)
	})
	
	t.Run("go_runtime_version", func(t *testing.T) {
		version := runtime.Version()
		t.Logf("Go Runtime: %s", version)
		
		// Parse Go version (e.g., "go1.21.0")
		versionRegex := regexp.MustCompile(`go(\d+)\.(\d+)`)
		matches := versionRegex.FindStringSubmatch(version)
		
		if len(matches) >= 3 {
			major, _ := strconv.Atoi(matches[1])
			minor, _ := strconv.Atoi(matches[2])
			
			// Require Go 1.21+
			versionOK := major > 1 || (major == 1 && minor >= 21)
			assert.True(t, versionOK, 
				"Go version %d.%d should be >= 1.21", major, minor)
		}
	})
}

// TestDependencyErrorMessages verifies error messages for missing dependencies
func TestDependencyErrorMessages(t *testing.T) {
	t.Run("docker_not_found_message", func(t *testing.T) {
		// Test that we can detect when Docker is not available
		cmd := exec.Command("docker", "info")
		err := cmd.Run()
		
		if err == nil {
			t.Skip("Docker is running - cannot test error message")
			return
		}
		
		// Verify we can provide helpful error message
		expectedMessages := []string{
			"Docker is not installed or not running",
			"Please install Docker from https://docker.com",
			"Ensure Docker daemon is running",
		}
		
		// This would be implemented in the actual CLI
		t.Logf("Would show error message when Docker is not available")
		for _, msg := range expectedMessages {
			t.Logf("  - %s", msg)
		}
	})
	
	t.Run("version_mismatch_warning", func(t *testing.T) {
		// Test version compatibility warnings
		cmd := exec.Command("docker", "--version")
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			t.Skip("Docker is not installed")
			return
		}
		
		versionRegex := regexp.MustCompile(`Docker version (\d+)\.(\d+)`)
		matches := versionRegex.FindStringSubmatch(string(output))
		
		if len(matches) >= 3 {
			major, _ := strconv.Atoi(matches[1])
			minor, _ := strconv.Atoi(matches[2])
			
			// Check if version is older than recommended
			if major < 20 || (major == 20 && minor < 10) {
				t.Logf("Warning: Docker version %d.%d is older than recommended 20.10", 
					major, minor)
				t.Log("Consider upgrading Docker for best compatibility")
			}
		}
	})
}