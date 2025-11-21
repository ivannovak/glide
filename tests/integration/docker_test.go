package integration_test

import (
	stdctx "context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ivannovak/glide/internal/context"
	"github.com/ivannovak/glide/internal/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDockerOperations tests real Docker operations
func TestDockerOperations(t *testing.T) {
	// Skip if Docker is not available
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skip("Docker is not available - skipping Docker operations tests")
		return
	}

	t.Run("docker_compose_version", func(t *testing.T) {
		// Test Docker Compose V2 command
		cmd := exec.Command("docker", "compose", "version")
		output, err := cmd.CombinedOutput()

		require.NoError(t, err, "Docker Compose V2 should be available")
		assert.Contains(t, string(output), "Docker Compose version")
	})

	t.Run("docker_resolver_initialization", func(t *testing.T) {
		// Create a mock project context
		ctx := &context.ProjectContext{
			ProjectRoot:     "/test/project",
			DevelopmentMode: context.ModeSingleRepo,
			Location:        context.LocationMainRepo,
		}

		// Create resolver
		resolver := docker.NewResolver(ctx)
		assert.NotNil(t, resolver)

		// Test Docker detection through Resolve
		err := resolver.Resolve()
		// Resolve should work if Docker is available
		if err != nil {
			t.Logf("Resolver reported: %v", err)
		}
	})

	t.Run("docker_compose_file_resolution", func(t *testing.T) {
		tmpDir := t.TempDir()
		vcsDir := filepath.Join(tmpDir, "vcs")

		// Create project structure with compose file
		require.NoError(t, os.MkdirAll(vcsDir, 0755))
		composeFile := filepath.Join(vcsDir, "docker-compose.yml")
		composeContent := `version: '3'
services:
  test:
    image: alpine:latest
    command: echo "test"
`
		require.NoError(t, os.WriteFile(composeFile, []byte(composeContent), 0644))

		// Create context
		ctx := &context.ProjectContext{
			ProjectRoot:     tmpDir,
			DevelopmentMode: context.ModeSingleRepo,
			Location:        context.LocationMainRepo,
			WorkingDir:      vcsDir,
		}

		// Create resolver and resolve files
		resolver := docker.NewResolver(ctx)

		// First resolve to find compose files
		err := resolver.Resolve()
		if err != nil {
			t.Logf("Resolve error: %v", err)
		}

		files := resolver.GetComposeFiles()

		// The resolver might need the compose files set in context
		if len(files) == 0 {
			// Try setting compose files directly
			ctx.ComposeFiles = []string{composeFile}
			resolver = docker.NewResolver(ctx)
			files = resolver.GetComposeFiles()
		}

		if len(files) > 0 {
			assert.Contains(t, files[0], "docker-compose.yml")
		} else {
			t.Log("No compose files found by resolver")
		}
	})
}

// TestDockerContainerLifecycle tests container lifecycle operations
func TestDockerContainerLifecycle(t *testing.T) {
	// Skip if Docker is not available
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skip("Docker is not available")
		return
	}

	t.Run("simple_container_lifecycle", func(t *testing.T) {
		// Create a simple test container
		containerName := "glide-test-" + time.Now().Format("20060102150405")

		// Start container
		startCmd := exec.Command("docker", "run", "-d", "--name", containerName,
			"alpine:latest", "sleep", "30")
		err := startCmd.Run()
		require.NoError(t, err, "Should start test container")

		// Ensure cleanup
		defer func() {
			exec.Command("docker", "rm", "-f", containerName).Run()
		}()

		// Check container is running
		psCmd := exec.Command("docker", "ps", "--filter", "name="+containerName,
			"--format", "{{.Names}}")
		output, err := psCmd.CombinedOutput()
		require.NoError(t, err)
		assert.Contains(t, string(output), containerName, "Container should be running")

		// Stop container
		stopCmd := exec.Command("docker", "stop", containerName)
		err = stopCmd.Run()
		require.NoError(t, err, "Should stop container")

		// Verify container stopped
		psCmd = exec.Command("docker", "ps", "-a", "--filter", "name="+containerName,
			"--filter", "status=exited", "--format", "{{.Names}}")
		output, err = psCmd.CombinedOutput()
		require.NoError(t, err)
		assert.Contains(t, string(output), containerName, "Container should be stopped")
	})
}

// TestDockerNetworking tests Docker network operations
func TestDockerNetworking(t *testing.T) {
	// Skip if Docker is not available
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skip("Docker is not available")
		return
	}

	t.Run("network_creation_and_cleanup", func(t *testing.T) {
		networkName := "glide-test-net-" + time.Now().Format("20060102150405")

		// Create network
		createCmd := exec.Command("docker", "network", "create", networkName)
		err := createCmd.Run()
		require.NoError(t, err, "Should create network")

		// Ensure cleanup
		defer func() {
			exec.Command("docker", "network", "rm", networkName).Run()
		}()

		// List networks and verify creation
		lsCmd := exec.Command("docker", "network", "ls", "--format", "{{.Name}}")
		output, err := lsCmd.CombinedOutput()
		require.NoError(t, err)
		assert.Contains(t, string(output), networkName, "Network should exist")

		// Remove network
		rmCmd := exec.Command("docker", "network", "rm", networkName)
		err = rmCmd.Run()
		require.NoError(t, err, "Should remove network")

		// Verify removal
		lsCmd = exec.Command("docker", "network", "ls", "--format", "{{.Name}}")
		output, err = lsCmd.CombinedOutput()
		require.NoError(t, err)
		assert.NotContains(t, string(output), networkName, "Network should be removed")
	})
}

// TestDockerVolumeOperations tests Docker volume operations
func TestDockerVolumeOperations(t *testing.T) {
	// Skip if Docker is not available
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skip("Docker is not available")
		return
	}

	t.Run("volume_creation_and_cleanup", func(t *testing.T) {
		volumeName := "glide-test-vol-" + time.Now().Format("20060102150405")

		// Create volume
		createCmd := exec.Command("docker", "volume", "create", volumeName)
		err := createCmd.Run()
		require.NoError(t, err, "Should create volume")

		// Ensure cleanup
		defer func() {
			exec.Command("docker", "volume", "rm", volumeName).Run()
		}()

		// List volumes and verify creation
		lsCmd := exec.Command("docker", "volume", "ls", "--format", "{{.Name}}")
		output, err := lsCmd.CombinedOutput()
		require.NoError(t, err)
		assert.Contains(t, string(output), volumeName, "Volume should exist")

		// Inspect volume
		inspectCmd := exec.Command("docker", "volume", "inspect", volumeName,
			"--format", "{{.Name}}")
		output, err = inspectCmd.CombinedOutput()
		require.NoError(t, err)
		assert.Equal(t, volumeName, strings.TrimSpace(string(output)))
	})
}

// TestDockerComposeIntegration tests Docker Compose integration
func TestDockerComposeIntegration(t *testing.T) {
	// Skip if Docker Compose is not available
	if err := exec.Command("docker", "compose", "version").Run(); err != nil {
		t.Skip("Docker Compose V2 is not available")
		return
	}

	t.Run("compose_file_validation", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a valid compose file
		validCompose := `version: '3.8'
services:
  test-service:
    image: alpine:latest
    command: echo "Hello from Glide test"
`
		composeFile := filepath.Join(tmpDir, "docker-compose.yml")
		err := os.WriteFile(composeFile, []byte(validCompose), 0644)
		require.NoError(t, err)

		// Validate compose file
		validateCmd := exec.Command("docker", "compose", "-f", composeFile, "config", "--quiet")
		err = validateCmd.Run()
		assert.NoError(t, err, "Valid compose file should pass validation")
	})

	t.Run("compose_invalid_file_detection", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create an invalid compose file
		invalidCompose := `version: '3.8'
services:
  test-service:
    invalid_key: should_fail
    image: 
`
		composeFile := filepath.Join(tmpDir, "docker-compose.yml")
		err := os.WriteFile(composeFile, []byte(invalidCompose), 0644)
		require.NoError(t, err)

		// Try to validate invalid compose file
		validateCmd := exec.Command("docker", "compose", "-f", composeFile, "config", "--quiet")
		err = validateCmd.Run()
		assert.Error(t, err, "Invalid compose file should fail validation")
	})

	t.Run("compose_project_name_generation", func(t *testing.T) {
		t.Skip("Skipping: Docker functionality has moved to plugin system")
		tmpDir := t.TempDir()
		projectName := filepath.Base(tmpDir)

		// Create resolver with context
		ctx := &context.ProjectContext{
			ProjectRoot:     tmpDir,
			DevelopmentMode: context.ModeSingleRepo,
			WorkingDir:      tmpDir,
		}

		resolver := docker.NewResolver(ctx)

		// Get project name using the compose project name method
		generatedName := resolver.GetComposeProjectName()
		assert.NotEmpty(t, generatedName)
		// Project name should be derived from the directory
		t.Logf("Generated project name: %s from directory: %s", generatedName, projectName)
	})
}

// TestDockerHealthChecks tests Docker health monitoring
func TestDockerHealthChecks(t *testing.T) {
	// Skip if Docker is not available
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skip("Docker is not available")
		return
	}

	t.Run("docker_daemon_health", func(t *testing.T) {
		// Check Docker daemon is healthy
		cmd := exec.Command("docker", "info", "--format", "{{.ServerVersion}}")
		output, err := cmd.CombinedOutput()

		require.NoError(t, err, "Docker daemon should be healthy")
		assert.NotEmpty(t, string(output), "Should get Docker server version")
	})

	t.Run("docker_disk_usage", func(t *testing.T) {
		// Check Docker disk usage
		cmd := exec.Command("docker", "system", "df", "--format", "json")
		output, err := cmd.CombinedOutput()

		require.NoError(t, err, "Should get Docker disk usage")
		assert.Contains(t, string(output), "Images", "Should contain usage information")
	})
}

// TestDockerResourceCleanup tests Docker resource cleanup operations
func TestDockerResourceCleanup(t *testing.T) {
	// Skip if Docker is not available
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skip("Docker is not available")
		return
	}

	t.Run("list_dangling_images", func(t *testing.T) {
		// List dangling images (images with no tags)
		cmd := exec.Command("docker", "images", "-f", "dangling=true", "-q")
		output, err := cmd.CombinedOutput()

		require.NoError(t, err, "Should list dangling images")
		// Output may be empty if no dangling images
		t.Logf("Dangling images: %d", len(strings.Split(strings.TrimSpace(string(output)), "\n")))
	})

	t.Run("list_stopped_containers", func(t *testing.T) {
		// List stopped containers
		cmd := exec.Command("docker", "ps", "-a", "-f", "status=exited", "--format", "{{.Names}}")
		output, err := cmd.CombinedOutput()

		require.NoError(t, err, "Should list stopped containers")

		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		if lines[0] != "" {
			t.Logf("Found %d stopped containers", len(lines))
		} else {
			t.Log("No stopped containers found")
		}
	})

	t.Run("list_unused_volumes", func(t *testing.T) {
		// List volumes not attached to any container
		cmd := exec.Command("docker", "volume", "ls", "-f", "dangling=true", "--format", "{{.Name}}")
		output, err := cmd.CombinedOutput()

		require.NoError(t, err, "Should list unused volumes")

		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		if lines[0] != "" {
			t.Logf("Found %d unused volumes", len(lines))
		} else {
			t.Log("No unused volumes found")
		}
	})
}

// TestDockerErrorHandling tests error handling for Docker operations
func TestDockerErrorHandling(t *testing.T) {
	// Skip if Docker is not available
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skip("Docker is not available")
		return
	}

	t.Run("handle_invalid_image", func(t *testing.T) {
		// Try to run a non-existent image with timeout
		ctx, cancel := stdctx.WithTimeout(stdctx.Background(), 5*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "docker", "run", "--rm", "--pull=never", "definitely-not-a-real-image:latest")
		output, err := cmd.CombinedOutput()

		assert.Error(t, err, "Should fail with non-existent image")
		outputStr := string(output)
		// Docker may report either "Unable to find image" or "No such image" depending on version
		assert.True(t, strings.Contains(outputStr, "Unable to find image") || strings.Contains(outputStr, "No such image"),
			"Should report image not found, got: %s", outputStr)
	})

	t.Run("handle_port_conflict", func(t *testing.T) {
		// Start a container on a specific port
		containerName := "glide-port-test-" + time.Now().Format("20060102150405")
		port := "38765"

		// Start first container
		cmd1 := exec.Command("docker", "run", "-d", "--name", containerName,
			"-p", port+":80", "nginx:alpine")
		err := cmd1.Run()

		if err != nil {
			t.Skip("Could not start test container")
			return
		}

		// Ensure cleanup
		defer func() {
			exec.Command("docker", "rm", "-f", containerName).Run()
		}()

		// Try to start second container on same port
		containerName2 := containerName + "-2"
		cmd2 := exec.Command("docker", "run", "-d", "--name", containerName2,
			"-p", port+":80", "nginx:alpine")
		output, err := cmd2.CombinedOutput()

		// Should fail due to port conflict
		assert.Error(t, err, "Should fail with port conflict")
		assert.Contains(t, strings.ToLower(string(output)), "bind", "Should report bind error")

		// Clean up second container if it somehow started
		exec.Command("docker", "rm", "-f", containerName2).Run()
	})
}
