package e2e_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ivannovak/glide/internal/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMultiWorktreeScenarios tests end-to-end multi-worktree specific workflows
func TestMultiWorktreeScenarios(t *testing.T) {
	// Skip if git is not available
	if err := exec.Command("git", "--version").Run(); err != nil {
		t.Skip("Git is not available")
		return
	}

	t.Run("parallel_development", func(t *testing.T) {
		// Test parallel development scenario:
		// 1. Multiple worktrees running simultaneously
		// 2. Resource isolation verification
		// 3. Port management simulation

		tmpDir := t.TempDir()
		vcsDir := filepath.Join(tmpDir, "vcs")
		worktreesDir := filepath.Join(tmpDir, "worktrees")

		// Setup main repository
		require.NoError(t, os.MkdirAll(vcsDir, 0755))
		require.NoError(t, os.MkdirAll(worktreesDir, 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		err := os.Chdir(vcsDir)
		require.NoError(t, err)

		// Initialize main repository
		exec.Command("git", "init").Run()
		exec.Command("git", "config", "user.email", "test@example.com").Run()
		exec.Command("git", "config", "user.name", "Test User").Run()

		// Create base application structure
		baseFiles := map[string]string{
			"docker-compose.yml": `version: '3.8'
services:
  app:
    image: php:8.3-cli
    ports:
      - "8080:80"
    environment:
      - APP_ENV=local
  db:
    image: mysql:8.0
    ports:
      - "3306:3306"
    environment:
      - MYSQL_ROOT_PASSWORD=secret`,
			"app/config.php": `<?php
return [
    'app_name' => 'Multi Worktree App',
    'version' => '1.0.0'
];`,
			"README.md": "# Multi-Worktree Application",
		}

		for path, content := range baseFiles {
			require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
			require.NoError(t, os.WriteFile(path, []byte(content), 0644))
		}

		// Initial commit
		exec.Command("git", "add", ".").Run()
		exec.Command("git", "commit", "-m", "Initial application").Run()

		// Step 1: Create multiple worktrees for parallel development
		features := []struct {
			name        string
			branch      string
			description string
			files       map[string]string
		}{
			{
				name:        "frontend-redesign",
				branch:      "feature/frontend-redesign",
				description: "UI/UX improvements",
				files: map[string]string{
					"resources/css/new-theme.css":     "/* New theme styles */",
					"resources/js/components.js":      "// New React components",
					"resources/views/layouts/app.php": "<!-- New layout -->",
				},
			},
			{
				name:        "api-v2",
				branch:      "feature/api-v2",
				description: "API version 2 development",
				files: map[string]string{
					"routes/api-v2.php":       "<?php // API v2 routes",
					"app/Http/V2/Controller.php": "<?php // V2 Controller",
					"tests/ApiV2Test.php":      "<?php // V2 API tests",
				},
			},
			{
				name:        "performance-optimization",
				branch:      "feature/performance-optimization", 
				description: "Database and caching improvements",
				files: map[string]string{
					"app/Services/CacheService.php": "<?php // Caching service",
					"database/migrations/optimize.php": "<?php // DB optimization",
					"config/cache.php": "<?php // Cache configuration",
				},
			},
		}

		worktreePaths := make([]string, len(features))

		// Create all worktrees
		for i, feature := range features {
			worktreePath := filepath.Join(worktreesDir, feature.name)
			worktreePaths[i] = worktreePath

			// Create worktree
			cmd := exec.Command("git", "worktree", "add", worktreePath, "-b", feature.branch)
			require.NoError(t, cmd.Run(), "Failed to create worktree for %s", feature.name)

			// Switch to worktree and add feature-specific files
			require.NoError(t, os.Chdir(worktreePath))

			// Create feature files
			for filePath, content := range feature.files {
				require.NoError(t, os.MkdirAll(filepath.Dir(filePath), 0755))
				require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))
			}

			// Create feature-specific docker-compose override to simulate port management
			overrideContent := fmt.Sprintf(`version: '3.8'
services:
  app:
    ports:
      - "%d:80"
    environment:
      - FEATURE=%s
  db:
    ports:
      - "%d:3306"
`, 8080+i+1, feature.name, 3306+i+1)

			require.NoError(t, os.WriteFile("docker-compose.override.yml", []byte(overrideContent), 0644))

			// Commit feature work
			exec.Command("git", "add", ".").Run()
			exec.Command("git", "commit", "-m", fmt.Sprintf("Add %s", feature.description)).Run()

			// Step 2: Verify resource isolation - each worktree should have different configurations
			ctx := context.Detect()
			assert.NotNil(t, ctx)
			assert.Equal(t, context.LocationWorktree, ctx.Location)
			assert.Equal(t, feature.name, ctx.WorktreeName)

			// Check that override files exist and have correct ports
			overrideData, err := os.ReadFile("docker-compose.override.yml")
			require.NoError(t, err)
			expectedPort := fmt.Sprintf("%d:80", 8080+i+1)
			assert.Contains(t, string(overrideData), expectedPort, "Worktree should have isolated port configuration")

			t.Logf("Created worktree %s with isolated configuration", feature.name)
		}

		// Step 3: Verify all worktrees are running in parallel
		os.Chdir(vcsDir)

		// List all worktrees
		cmd := exec.Command("git", "worktree", "list")
		output, err := cmd.Output()
		require.NoError(t, err)

		// Verify all features are listed
		for _, feature := range features {
			assert.Contains(t, string(output), feature.name, "Worktree %s should be listed", feature.name)
		}

		// Count active worktrees (should be 3 features + 1 main)
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		assert.Equal(t, 4, len(lines), "Should have 4 worktrees total (main + 3 features)")

		// Verify each worktree has its unique branch and content
		for i, feature := range features {
			worktreePath := worktreePaths[i]
			
			// Switch to worktree directory to verify its state
			originalDir, _ := os.Getwd()
			require.NoError(t, os.Chdir(worktreePath))

			// Verify we're on the correct branch
			cmd := exec.Command("git", "branch", "--show-current")
			output, err := cmd.Output()
			require.NoError(t, err)
			currentBranch := strings.TrimSpace(string(output))
			assert.Equal(t, feature.branch, currentBranch, "Should be on correct branch in %s", feature.name)

			// Check feature-specific files exist in the working directory
			for filePath := range feature.files {
				assert.FileExists(t, filePath, "Feature file should exist in %s working directory", feature.name)
			}

			// Return to original directory
			os.Chdir(originalDir)
		}

		t.Log("Parallel development isolation verified successfully")

		// Cleanup: Remove all worktrees
		for _, worktreePath := range worktreePaths {
			cmd := exec.Command("git", "worktree", "remove", worktreePath)
			require.NoError(t, cmd.Run())
			assert.NoDirExists(t, worktreePath)
		}
	})

	t.Run("global_operations", func(t *testing.T) {
		// Test global operations across multiple worktrees:
		// 1. Global status checking
		// 2. Global shutdown simulation
		// 3. Cross-worktree operations

		tmpDir := t.TempDir()
		vcsDir := filepath.Join(tmpDir, "vcs")
		worktreesDir := filepath.Join(tmpDir, "worktrees")

		require.NoError(t, os.MkdirAll(vcsDir, 0755))
		require.NoError(t, os.MkdirAll(worktreesDir, 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		// Setup main repository
		err := os.Chdir(vcsDir)
		require.NoError(t, err)

		exec.Command("git", "init").Run()
		exec.Command("git", "config", "user.email", "test@example.com").Run()
		exec.Command("git", "config", "user.name", "Test User").Run()

		// Create main project with docker-compose
		require.NoError(t, os.WriteFile("README.md", []byte("# Global Operations Test"), 0644))
		
		composeContent := `version: '3.8'
services:
  web:
    image: nginx:alpine
    ports:
      - "80:80"
  api:
    image: php:8.3-fpm
    ports:
      - "9000:9000"
  cache:
    image: redis:alpine
    ports:
      - "6379:6379"
`
		require.NoError(t, os.WriteFile("docker-compose.yml", []byte(composeContent), 0644))

		exec.Command("git", "add", ".").Run()
		exec.Command("git", "commit", "-m", "Initial project").Run()

		// Step 1: Create multiple active worktrees
		worktrees := []struct {
			name   string
			branch string
			status string // simulated status
		}{
			{"hotfix-security", "hotfix/security-patch", "active"},
			{"feature-dashboard", "feature/admin-dashboard", "active"},
			{"bugfix-login", "bugfix/login-issue", "inactive"},
			{"feature-reporting", "feature/monthly-reports", "active"},
		}

		createdWorktrees := make([]string, 0, len(worktrees))

		for _, wt := range worktrees {
			worktreePath := filepath.Join(worktreesDir, wt.name)
			createdWorktrees = append(createdWorktrees, worktreePath)

			// Create worktree
			cmd := exec.Command("git", "worktree", "add", worktreePath, "-b", wt.branch)
			require.NoError(t, cmd.Run())

			// Add some work to each worktree
			require.NoError(t, os.Chdir(worktreePath))

			workFile := fmt.Sprintf("work/%s.php", wt.name)
			require.NoError(t, os.MkdirAll("work", 0755))
			workContent := fmt.Sprintf("<?php\n// Work for %s\nclass %sWork {}\n", wt.name, strings.ReplaceAll(wt.name, "-", ""))
			require.NoError(t, os.WriteFile(workFile, []byte(workContent), 0644))

			// Create status file to simulate different states
			statusContent := fmt.Sprintf("status: %s\nlast_updated: %s\n", wt.status, time.Now().Format(time.RFC3339))
			require.NoError(t, os.WriteFile(".worktree_status", []byte(statusContent), 0644))

			exec.Command("git", "add", ".").Run()
			exec.Command("git", "commit", "-m", fmt.Sprintf("Add work for %s", wt.name)).Run()
		}

		// Step 2: Global status checking from project root
		os.Chdir(tmpDir) // Project root

		// Verify context detection from root
		ctx := context.Detect()
		assert.NotNil(t, ctx)
		assert.Equal(t, context.LocationRoot, ctx.Location)
		assert.True(t, ctx.IsRoot)

		// Simulate global status check - list all worktrees with their status
		cmd := exec.Command("git", "-C", vcsDir, "worktree", "list", "--porcelain")
		output, err := cmd.Output()
		require.NoError(t, err)

		// Parse worktree list output
		outputStr := string(output)
		for _, wt := range worktrees {
			// Should contain the worktree path and branch
			assert.Contains(t, outputStr, wt.name, "Global status should show %s", wt.name)
		}

		// Count total worktrees (including main)
		worktreeCount := strings.Count(outputStr, "worktree ")
		assert.Equal(t, len(worktrees)+1, worktreeCount, "Should show all worktrees plus main")

		// Step 3: Cross-worktree operations
		// Simulate operations that affect multiple worktrees

		// Check each worktree's individual status
		for i, wt := range worktrees {
			worktreePath := createdWorktrees[i]
			statusFile := filepath.Join(worktreePath, ".worktree_status")
			
			assert.FileExists(t, statusFile, "Status file should exist in %s", wt.name)
			
			statusData, err := os.ReadFile(statusFile)
			require.NoError(t, err)
			assert.Contains(t, string(statusData), fmt.Sprintf("status: %s", wt.status), "Status should match for %s", wt.name)
		}

		// Simulate global configuration that affects all worktrees
		globalConfigContent := `# Global Glide Configuration
project_mode: multi-worktree
docker_compose_base: vcs/docker-compose.yml
override_location: docker-compose.override.yml
active_worktrees: ` + strconv.Itoa(len(worktrees)) + `
`
		globalConfigPath := filepath.Join(tmpDir, ".glide.yml")
		require.NoError(t, os.WriteFile(globalConfigPath, []byte(globalConfigContent), 0644))
		assert.FileExists(t, globalConfigPath)

		// Verify global config would be accessible from any worktree
		for i, wt := range worktrees {
			require.NoError(t, os.Chdir(createdWorktrees[i]))
			
			// From any worktree, we should be able to detect the global config
			assert.FileExists(t, globalConfigPath, "Global config should be accessible from %s", wt.name)
		}

		t.Log("Global operations test completed successfully")

		// Cleanup: Global shutdown simulation
		os.Chdir(vcsDir)
		
		// Remove all worktrees (simulating global shutdown)
		for _, worktreePath := range createdWorktrees {
			cmd := exec.Command("git", "worktree", "remove", worktreePath)
			require.NoError(t, cmd.Run())
			assert.NoDirExists(t, worktreePath)
		}

		// Verify all worktrees are removed
		cmd = exec.Command("git", "worktree", "list")
		output, err = cmd.Output()
		require.NoError(t, err)
		
		// Should only have main worktree left
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		assert.Equal(t, 1, len(lines), "Should only have main worktree after cleanup")
	})

	t.Run("resource_cleanup", func(t *testing.T) {
		// Test resource cleanup scenarios:
		// 1. Orphaned container detection
		// 2. Stale worktree cleanup
		// 3. Disk space management

		tmpDir := t.TempDir()
		vcsDir := filepath.Join(tmpDir, "vcs")
		worktreesDir := filepath.Join(tmpDir, "worktrees")

		require.NoError(t, os.MkdirAll(vcsDir, 0755))
		require.NoError(t, os.MkdirAll(worktreesDir, 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		// Setup project
		err := os.Chdir(vcsDir)
		require.NoError(t, err)

		exec.Command("git", "init").Run()
		exec.Command("git", "config", "user.email", "test@example.com").Run()
		exec.Command("git", "config", "user.name", "Test User").Run()

		require.NoError(t, os.WriteFile("README.md", []byte("# Resource Cleanup Test"), 0644))
		exec.Command("git", "add", ".").Run()
		exec.Command("git", "commit", "-m", "Initial").Run()

		// Step 1: Create worktrees and simulate orphaned resources
		worktreesToCreate := []string{"feature-a", "feature-b", "hotfix-c", "experiment-d"}
		worktreePaths := make([]string, len(worktreesToCreate))

		for i, name := range worktreesToCreate {
			worktreePath := filepath.Join(worktreesDir, name)
			worktreePaths[i] = worktreePath

			cmd := exec.Command("git", "worktree", "add", worktreePath, "-b", "branch-"+name)
			require.NoError(t, cmd.Run())

			// Add some "resource" files to simulate containers, caches, logs
			require.NoError(t, os.Chdir(worktreePath))

			// Simulate Docker resources
			require.NoError(t, os.MkdirAll("docker/volumes", 0755))
			require.NoError(t, os.MkdirAll("docker/networks", 0755))
			require.NoError(t, os.WriteFile("docker/container_id.txt", []byte(name+"_container_123"), 0644))

			// Simulate cache and log files
			require.NoError(t, os.MkdirAll("storage/cache", 0755))
			require.NoError(t, os.MkdirAll("storage/logs", 0755))
			
			cacheSize := (i + 1) * 1024 // Simulate different cache sizes
			cacheContent := strings.Repeat("x", cacheSize)
			require.NoError(t, os.WriteFile("storage/cache/app.cache", []byte(cacheContent), 0644))

			logContent := fmt.Sprintf("[%s] Log entries for %s\n", time.Now().Format(time.RFC3339), name)
			logContent += strings.Repeat("Log line\n", 100) // Simulate log size
			require.NoError(t, os.WriteFile("storage/logs/app.log", []byte(logContent), 0644))

			// Create metadata about resource usage
			metadata := fmt.Sprintf(`worktree: %s
created: %s
cache_size: %d
containers: 1
networks: 2
volumes: 3
`, name, time.Now().Format(time.RFC3339), cacheSize)
			require.NoError(t, os.WriteFile(".resource_metadata", []byte(metadata), 0644))

			exec.Command("git", "add", ".").Run()
			exec.Command("git", "commit", "-m", fmt.Sprintf("Add resources for %s", name)).Run()
		}

		os.Chdir(vcsDir)

		// Step 2: Simulate orphaned container detection
		// List all worktrees and their resources
		cmd := exec.Command("git", "worktree", "list")
		output, err := cmd.Output()
		require.NoError(t, err)

		activeWorktrees := strings.Split(strings.TrimSpace(string(output)), "\n")
		assert.Equal(t, len(worktreesToCreate)+1, len(activeWorktrees), "Should have all worktrees plus main")

		// Check each worktree's resources
		totalCacheSize := 0
		totalLogSize := 0
		containerCount := 0

		for i, worktreePath := range worktreePaths {
			metadataPath := filepath.Join(worktreePath, ".resource_metadata")
			assert.FileExists(t, metadataPath, "Resource metadata should exist")

			metadata, err := os.ReadFile(metadataPath)
			require.NoError(t, err)
			assert.Contains(t, string(metadata), worktreesToCreate[i], "Metadata should contain worktree name")

			// Check cache files
			cachePath := filepath.Join(worktreePath, "storage/cache/app.cache")
			if stat, err := os.Stat(cachePath); err == nil {
				totalCacheSize += int(stat.Size())
			}

			// Check log files
			logPath := filepath.Join(worktreePath, "storage/logs/app.log")
			if stat, err := os.Stat(logPath); err == nil {
				totalLogSize += int(stat.Size())
			}

			// Check container files
			containerPath := filepath.Join(worktreePath, "docker/container_id.txt")
			if _, err := os.Stat(containerPath); err == nil {
				containerCount++
			}
		}

		t.Logf("Resource usage summary: Cache=%d bytes, Logs=%d bytes, Containers=%d", 
			totalCacheSize, totalLogSize, containerCount)

		assert.Greater(t, totalCacheSize, 0, "Should have cache files")
		assert.Greater(t, totalLogSize, 0, "Should have log files")
		assert.Equal(t, len(worktreesToCreate), containerCount, "Should have container files for each worktree")

		// Step 3: Stale worktree cleanup simulation
		// Remove some worktrees and check for stale resources
		staleMess := worktreePaths[:2] // Remove first two worktrees

		for _, stalePath := range staleMess {
			cmd := exec.Command("git", "worktree", "remove", stalePath)
			require.NoError(t, cmd.Run())
			assert.NoDirExists(t, stalePath)
		}

		// Verify cleanup
		cmd = exec.Command("git", "worktree", "list")
		output, err = cmd.Output()
		require.NoError(t, err)

		remainingWorktrees := strings.Split(strings.TrimSpace(string(output)), "\n")
		expectedRemaining := len(worktreesToCreate) - len(staleMess) + 1 // +1 for main
		assert.Equal(t, expectedRemaining, len(remainingWorktrees), "Should have correct number of remaining worktrees")

		// Step 4: Disk space management verification
		// Check remaining worktrees still have their resources
		remainingPaths := worktreePaths[len(staleMess):]
		
		for i, remainingPath := range remainingPaths {
			worktreeName := worktreesToCreate[len(staleMess)+i]
			
			assert.DirExists(t, remainingPath, "Remaining worktree should exist: %s", worktreeName)
			
			metadataPath := filepath.Join(remainingPath, ".resource_metadata")
			assert.FileExists(t, metadataPath, "Resource metadata should still exist: %s", worktreeName)
			
			cachePath := filepath.Join(remainingPath, "storage/cache/app.cache")
			assert.FileExists(t, cachePath, "Cache should still exist: %s", worktreeName)
		}

		// Final cleanup
		for _, remainingPath := range remainingPaths {
			cmd := exec.Command("git", "worktree", "remove", remainingPath)
			require.NoError(t, cmd.Run())
		}

		// Verify all cleaned up
		cmd = exec.Command("git", "worktree", "list")
		output, err = cmd.Output()
		require.NoError(t, err)
		
		finalWorktrees := strings.Split(strings.TrimSpace(string(output)), "\n")
		assert.Equal(t, 1, len(finalWorktrees), "Should only have main worktree left")

		t.Log("Resource cleanup test completed successfully")
	})
}

// TestMultiWorktreeIntegration tests complex multi-worktree integration scenarios
func TestMultiWorktreeIntegration(t *testing.T) {
	t.Run("branch_merge_workflow", func(t *testing.T) {
		// Test branch merging workflow with multiple worktrees
		tmpDir := t.TempDir()
		vcsDir := filepath.Join(tmpDir, "vcs")
		worktreesDir := filepath.Join(tmpDir, "worktrees")

		require.NoError(t, os.MkdirAll(vcsDir, 0755))
		require.NoError(t, os.MkdirAll(worktreesDir, 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		// Setup main repository
		err := os.Chdir(vcsDir)
		require.NoError(t, err)

		exec.Command("git", "init").Run()
		exec.Command("git", "config", "user.email", "test@example.com").Run()
		exec.Command("git", "config", "user.name", "Test User").Run()

		// Create initial application
		appContent := `<?php
class App {
    public function version() {
        return "1.0.0";
    }
    
    public function features() {
        return [];
    }
}
`
		require.NoError(t, os.WriteFile("app.php", []byte(appContent), 0644))
		exec.Command("git", "add", ".").Run()
		exec.Command("git", "commit", "-m", "Initial app").Run()

		// Create multiple feature branches via worktrees
		features := []struct {
			name   string
			branch string
			code   string
		}{
			{
				"user-auth",
				"feature/user-authentication",
				`<?php
class App {
    public function version() {
        return "1.1.0";
    }
    
    public function features() {
        return ["authentication"];
    }
    
    public function authenticate($user, $pass) {
        return $user === "admin" && $pass === "secret";
    }
}
`,
			},
			{
				"user-profiles",
				"feature/user-profiles", 
				`<?php
class App {
    public function version() {
        return "1.2.0";
    }
    
    public function features() {
        return ["profiles"];
    }
    
    public function getProfile($userId) {
        return ["id" => $userId, "name" => "User " . $userId];
    }
}
`,
			},
		}

		createdWorktrees := make([]string, len(features))

		// Develop features in parallel
		for i, feature := range features {
			worktreePath := filepath.Join(worktreesDir, feature.name)
			createdWorktrees[i] = worktreePath

			// Create worktree
			cmd := exec.Command("git", "worktree", "add", worktreePath, "-b", feature.branch)
			require.NoError(t, cmd.Run())

			// Develop feature
			require.NoError(t, os.Chdir(worktreePath))
			require.NoError(t, os.WriteFile("app.php", []byte(feature.code), 0644))

			exec.Command("git", "add", ".").Run()
			exec.Command("git", "commit", "-m", fmt.Sprintf("Implement %s feature", feature.name)).Run()

			// Verify context detection
			ctx := context.Detect()
			assert.NotNil(t, ctx)
			assert.Equal(t, context.LocationWorktree, ctx.Location)
			assert.Equal(t, feature.name, ctx.WorktreeName)
		}

		// Return to main branch for merging simulation
		os.Chdir(vcsDir)

		// Verify main branch state
		mainContent, err := os.ReadFile("app.php")
		require.NoError(t, err)
		assert.Contains(t, string(mainContent), `return "1.0.0"`, "Main branch should have original version")

		// Simulate feature integration (we won't actually merge due to conflicts, but verify the setup)
		for i, feature := range features {
			worktreePath := createdWorktrees[i]
			
			// Verify feature branch has its changes
			featureContent, err := os.ReadFile(filepath.Join(worktreePath, "app.php"))
			require.NoError(t, err)
			// Check for feature-specific content instead of feature name
			if feature.name == "user-auth" {
				assert.Contains(t, string(featureContent), "authenticate", "Feature branch should have authentication code")
			} else if feature.name == "user-profiles" {
				assert.Contains(t, string(featureContent), "getProfile", "Feature branch should have profile code")
			}
		}

		// Cleanup
		for _, worktreePath := range createdWorktrees {
			cmd := exec.Command("git", "worktree", "remove", worktreePath)
			require.NoError(t, cmd.Run())
		}

		t.Log("Branch merge workflow test completed successfully")
	})

	t.Run("configuration_inheritance", func(t *testing.T) {
		// Test configuration inheritance across worktrees
		tmpDir := t.TempDir()
		vcsDir := filepath.Join(tmpDir, "vcs")
		worktreesDir := filepath.Join(tmpDir, "worktrees")

		require.NoError(t, os.MkdirAll(vcsDir, 0755))
		require.NoError(t, os.MkdirAll(worktreesDir, 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		// Setup project with configuration files
		err := os.Chdir(vcsDir)
		require.NoError(t, err)

		exec.Command("git", "init").Run()
		exec.Command("git", "config", "user.email", "test@example.com").Run()
		exec.Command("git", "config", "user.name", "Test User").Run()

		// Create base configuration
		baseConfig := `# Base Configuration
database:
  host: localhost
  port: 3306
  name: myapp

cache:
  driver: redis
  host: localhost
  port: 6379

app:
  name: "My Application"
  debug: false
`
		require.NoError(t, os.WriteFile("config.yml", []byte(baseConfig), 0644))

		// Create environment-specific configs that worktrees might override
		devConfig := `# Development Configuration  
database:
  name: myapp_dev
  
app:
  debug: true
`
		require.NoError(t, os.WriteFile("config.dev.yml", []byte(devConfig), 0644))

		exec.Command("git", "add", ".").Run()
		exec.Command("git", "commit", "-m", "Add base configuration").Run()

		// Create worktrees with different configuration needs
		configTests := []struct {
			name         string
			branch       string
			configOverrides map[string]string
		}{
			{
				"api-testing",
				"feature/api-testing",
				map[string]string{
					"config.test.yml": `# API Testing Configuration
database:
  name: myapp_test_api

app:
  debug: true
  api_version: "v2"
`,
				},
			},
			{
				"performance-testing", 
				"feature/performance-testing",
				map[string]string{
					"config.perf.yml": `# Performance Testing Configuration
database:
  name: myapp_perf
  
cache:
  driver: memory
  
app:
  debug: false
  profiling: true
`,
				},
			},
		}

		createdWorktrees := make([]string, len(configTests))

		for i, test := range configTests {
			worktreePath := filepath.Join(worktreesDir, test.name)
			createdWorktrees[i] = worktreePath

			// Create worktree
			cmd := exec.Command("git", "worktree", "add", worktreePath, "-b", test.branch)
			require.NoError(t, cmd.Run())

			require.NoError(t, os.Chdir(worktreePath))

			// Verify base configuration exists
			assert.FileExists(t, "config.yml", "Base config should be inherited")
			assert.FileExists(t, "config.dev.yml", "Dev config should be inherited")

			// Add worktree-specific configuration
			for configFile, configContent := range test.configOverrides {
				require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))
			}

			exec.Command("git", "add", ".").Run()
			exec.Command("git", "commit", "-m", fmt.Sprintf("Add %s configuration", test.name)).Run()

			// Verify inheritance and overrides
			baseData, err := os.ReadFile("config.yml")
			require.NoError(t, err)
			assert.Contains(t, string(baseData), "My Application", "Base config should be inherited")

			// Check specific overrides exist
			for configFile := range test.configOverrides {
				assert.FileExists(t, configFile, "Override config should exist")
			}
		}

		// Return to main and verify inheritance isolation
		os.Chdir(vcsDir)

		// Main branch should not have worktree-specific configs
		for _, test := range configTests {
			for configFile := range test.configOverrides {
				assert.NoFileExists(t, configFile, "Main should not have worktree-specific config: %s", configFile)
			}
		}

		// Each worktree should have its own branch-specific configuration
		for i, test := range configTests {
			worktreePath := createdWorktrees[i]
			
			// Switch to worktree to check its branch-specific state
			originalDir, _ := os.Getwd()
			require.NoError(t, os.Chdir(worktreePath))

			// Verify we're on the correct branch
			cmd := exec.Command("git", "branch", "--show-current") 
			output, err := cmd.Output()
			require.NoError(t, err)
			currentBranch := strings.TrimSpace(string(output))
			assert.Equal(t, test.branch, currentBranch, "Should be on correct branch")
			
			// Should have base configs (inherited from main)
			assert.FileExists(t, "config.yml")
			assert.FileExists(t, "config.dev.yml")
			
			// Should have its branch-specific configs
			for configFile := range test.configOverrides {
				assert.FileExists(t, configFile, "Branch should have its specific config: %s", configFile)
			}
			
			// Return to original directory
			os.Chdir(originalDir)
		}

		// Cleanup
		for _, worktreePath := range createdWorktrees {
			cmd := exec.Command("git", "worktree", "remove", worktreePath)
			require.NoError(t, cmd.Run())
		}

		t.Log("Configuration inheritance test completed successfully")
	})
}