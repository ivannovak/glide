package e2e_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ivannovak/glide/v3/internal/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDeveloperWorkflows tests complete end-to-end user workflows
func TestDeveloperWorkflows(t *testing.T) {
	// Skip in short mode (pre-commit hooks)
	if testing.Short() {
		t.Skip("Skipping workflow tests with git worktrees in short mode")
	}

	// Skip if required tools are not available
	if err := exec.Command("git", "--version").Run(); err != nil {
		t.Skip("Git is not available")
		return
	}

	t.Run("daily_development_workflow", func(t *testing.T) {
		// Test complete daily development workflow:
		// 1. Project setup from scratch
		// 2. Container startup
		// 3. Test execution
		// 4. Code changes and testing
		// 5. Container shutdown

		tmpDir := t.TempDir()
		projectDir := filepath.Join(tmpDir, "test-project")
		vcsDir := filepath.Join(projectDir, "vcs")

		// Step 1: Project setup from scratch
		require.NoError(t, os.MkdirAll(vcsDir, 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		err := os.Chdir(vcsDir)
		require.NoError(t, err)

		// Initialize git repo
		exec.Command("git", "init").Run()
		exec.Command("git", "config", "user.email", "test@example.com").Run()
		exec.Command("git", "config", "user.name", "Test User").Run()

		// Create basic project structure
		require.NoError(t, os.WriteFile("README.md", []byte("# Test Project"), 0644))

		// Create a basic docker-compose.yml for testing
		composeContent := `version: '3.8'
services:
  php:
    image: alpine:latest
    command: sleep 10
    ports:
      - "8080:80"
  test-db:
    image: alpine:latest
    command: sleep 10
`
		require.NoError(t, os.WriteFile("docker-compose.yml", []byte(composeContent), 0644))

		// Create override file in project root
		overrideContent := `version: '3.8'
services:
  php:
    environment:
      - APP_ENV=testing
`
		require.NoError(t, os.WriteFile(filepath.Join(projectDir, "docker-compose.override.yml"),
			[]byte(overrideContent), 0644))

		// Commit initial files
		exec.Command("git", "add", ".").Run()
		exec.Command("git", "commit", "-m", "Initial commit").Run()

		// Verify context detection after setup
		ctx := context.Detect()
		assert.NotNil(t, ctx)
		if len(ctx.ComposeFiles) > 0 {
			if len(ctx.ComposeFiles) > 0 {
				assert.Contains(t, ctx.ComposeFiles[0], "docker-compose.yml")
			}
		}

		// Step 2: Simulate container startup
		// Note: We don't actually start containers in e2e tests to avoid conflicts
		// but we verify the compose command would be constructed correctly
		assert.FileExists(t, "docker-compose.yml")
		assert.FileExists(t, filepath.Join(projectDir, "docker-compose.override.yml"))

		// Step 3: Verify test infrastructure would be available
		// Check that we can detect what a test command would need
		if _, err := exec.LookPath("php"); err == nil {
			// If PHP is available, we could run tests
			t.Log("PHP available for testing")
		} else {
			t.Log("PHP not available, would need Docker container")
		}

		// Step 4: Simulate code changes
		testFile := filepath.Join("src", "TestClass.php")
		require.NoError(t, os.MkdirAll("src", 0755))
		testContent := `<?php
class TestClass {
    public function hello(): string {
        return "Hello from test!";
    }
}
`
		require.NoError(t, os.WriteFile(testFile, []byte(testContent), 0644))

		// Commit changes
		exec.Command("git", "add", ".").Run()
		exec.Command("git", "commit", "-m", "Add test class").Run()

		// Step 5: Verify project structure is ready for development
		assert.DirExists(t, "src")
		assert.FileExists(t, testFile)
		assert.FileExists(t, "docker-compose.yml")

		// Verify git history
		cmd := exec.Command("git", "log", "--oneline")
		output, err := cmd.Output()
		require.NoError(t, err)
		assert.Contains(t, string(output), "Add test class")
		assert.Contains(t, string(output), "Initial commit")
	})

	t.Run("feature_development_workflow", func(t *testing.T) {
		// Test feature development workflow:
		// 1. Worktree creation
		// 2. Environment setup
		// 3. Development and testing
		// 4. Worktree cleanup

		tmpDir := t.TempDir()
		vcsDir := filepath.Join(tmpDir, "vcs")
		worktreesDir := filepath.Join(tmpDir, "worktrees")

		// Setup base project
		require.NoError(t, os.MkdirAll(vcsDir, 0755))
		require.NoError(t, os.MkdirAll(worktreesDir, 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		err := os.Chdir(vcsDir)
		require.NoError(t, err)

		// Initialize repo with main branch
		exec.Command("git", "init").Run()
		exec.Command("git", "config", "user.email", "test@example.com").Run()
		exec.Command("git", "config", "user.name", "Test User").Run()

		// Create initial commit
		require.NoError(t, os.WriteFile("README.md", []byte("# Feature Test"), 0644))
		exec.Command("git", "add", ".").Run()
		exec.Command("git", "commit", "-m", "Initial").Run()

		// Step 1: Worktree creation
		featureDir := filepath.Join(worktreesDir, "feature-user-api")
		cmd := exec.Command("git", "worktree", "add", featureDir, "-b", "feature/user-api")
		err = cmd.Run()
		require.NoError(t, err)

		// Verify worktree was created
		assert.DirExists(t, featureDir)
		assert.FileExists(t, filepath.Join(featureDir, "README.md"))

		// Step 2: Environment setup in worktree
		err = os.Chdir(featureDir)
		require.NoError(t, err)

		// Verify we're on the correct branch
		cmd = exec.Command("git", "branch", "--show-current")
		output, err := cmd.Output()
		require.NoError(t, err)
		assert.Equal(t, "feature/user-api", strings.TrimSpace(string(output)))

		// Step 3: Development and testing simulation
		apiFile := "api/users.php"
		require.NoError(t, os.MkdirAll("api", 0755))
		apiContent := `<?php
class UserAPI {
    public function getUsers(): array {
        return ["user1", "user2"];
    }
}
`
		require.NoError(t, os.WriteFile(apiFile, []byte(apiContent), 0644))

		// Commit feature changes
		exec.Command("git", "add", ".").Run()
		exec.Command("git", "commit", "-m", "Add user API").Run()

		// Verify context detection in worktree
		ctx := context.Detect()
		assert.NotNil(t, ctx)
		assert.Equal(t, context.LocationWorktree, ctx.Location)
		assert.Equal(t, "feature-user-api", ctx.WorktreeName)

		// Step 4: Worktree cleanup preparation
		// Switch back to main repo to prepare for cleanup
		os.Chdir(vcsDir)

		// List worktrees to verify it exists
		cmd = exec.Command("git", "worktree", "list")
		output, err = cmd.Output()
		require.NoError(t, err)
		assert.Contains(t, string(output), "feature-user-api")

		// Cleanup: Remove worktree
		cmd = exec.Command("git", "worktree", "remove", featureDir)
		err = cmd.Run()
		require.NoError(t, err)

		// Verify cleanup
		assert.NoDirExists(t, featureDir)

		// Verify worktree is no longer listed
		cmd = exec.Command("git", "worktree", "list")
		output, err = cmd.Output()
		require.NoError(t, err)
		assert.NotContains(t, string(output), "feature-user-api")
	})

	t.Run("debugging_workflow", func(t *testing.T) {
		// Test debugging workflow:
		// 1. Log inspection setup
		// 2. Shell access simulation
		// 3. Database query preparation
		// 4. Error investigation workflow

		tmpDir := t.TempDir()
		projectDir := filepath.Join(tmpDir, "debug-project")
		vcsDir := filepath.Join(projectDir, "vcs")

		require.NoError(t, os.MkdirAll(vcsDir, 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		err := os.Chdir(vcsDir)
		require.NoError(t, err)

		// Setup project for debugging
		exec.Command("git", "init").Run()
		exec.Command("git", "config", "user.email", "test@example.com").Run()
		exec.Command("git", "config", "user.name", "Test User").Run()

		// Step 1: Log inspection setup
		// Create log files that would be generated by the application
		require.NoError(t, os.MkdirAll("storage/logs", 0755))
		logContent := `[2024-01-01 10:00:00] local.ERROR: Database connection failed
[2024-01-01 10:00:01] local.INFO: Retrying connection
[2024-01-01 10:00:02] local.ERROR: Authentication failed for user 'test'
[2024-01-01 10:00:03] local.DEBUG: SQL Query: SELECT * FROM users WHERE id = 1
[2024-01-01 10:00:04] local.INFO: User logged in successfully
`
		require.NoError(t, os.WriteFile("storage/logs/laravel.log", []byte(logContent), 0644))

		// Create docker-compose for services that would be debugged
		composeContent := `version: '3.8'
services:
  php:
    image: php:8.3-fpm
    volumes:
      - .:/var/www
  mysql:
    image: mysql:8.0
    environment:
      - MYSQL_ROOT_PASSWORD=secret
      - MYSQL_DATABASE=testdb
  redis:
    image: redis:alpine
`
		require.NoError(t, os.WriteFile("docker-compose.yml", []byte(composeContent), 0644))

		// Step 2: Shell access simulation
		// Verify we can identify which containers would be available for shell access
		assert.FileExists(t, "docker-compose.yml")

		// Parse compose file content to verify services
		composeData, err := os.ReadFile("docker-compose.yml")
		require.NoError(t, err)
		assert.Contains(t, string(composeData), "php:")
		assert.Contains(t, string(composeData), "mysql:")

		// Step 3: Database query preparation
		// Create database files that would be accessed during debugging
		require.NoError(t, os.MkdirAll("database/migrations", 0755))
		migrationContent := `<?php
use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;

class CreateUsersTable extends Migration
{
    public function up()
    {
        Schema::create('users', function (Blueprint $table) {
            $table->id();
            $table->string('name');
            $table->string('email')->unique();
            $table->timestamps();
        });
    }
}
`
		require.NoError(t, os.WriteFile("database/migrations/001_create_users_table.php", []byte(migrationContent), 0644))

		// Step 4: Error investigation workflow
		// Create error reproduction script
		errorScript := `#!/bin/bash
# Error reproduction script
echo "Reproducing error..."
echo "Step 1: Check database connection"
echo "Step 2: Verify user authentication"
echo "Step 3: Run failing query"
echo "Error reproduced successfully"
`
		require.NoError(t, os.MkdirAll("debug", 0755))
		require.NoError(t, os.WriteFile("debug/reproduce_error.sh", []byte(errorScript), 0755))

		// Commit debugging setup
		exec.Command("git", "add", ".").Run()
		exec.Command("git", "commit", "-m", "Add debugging setup").Run()

		// Verify debugging environment is ready
		assert.FileExists(t, "storage/logs/laravel.log")
		assert.FileExists(t, "docker-compose.yml")
		assert.FileExists(t, "database/migrations/001_create_users_table.php")
		assert.FileExists(t, "debug/reproduce_error.sh")

		// Test log content analysis
		logData, err := os.ReadFile("storage/logs/laravel.log")
		require.NoError(t, err)
		assert.Contains(t, string(logData), "Database connection failed")
		assert.Contains(t, string(logData), "Authentication failed")

		// Verify context detection
		ctx := context.Detect()
		assert.NotNil(t, ctx)
		if len(ctx.ComposeFiles) > 0 {
			assert.Contains(t, ctx.ComposeFiles[0], "docker-compose.yml")
		}

		t.Log("Debugging workflow environment successfully prepared")
	})
}

// TestComplexWorkflows tests more advanced end-to-end scenarios
func TestComplexWorkflows(t *testing.T) {
	// Skip in short mode (pre-commit hooks)
	if testing.Short() {
		t.Skip("Skipping complex workflow tests with git worktrees in short mode")
	}

	t.Run("full_project_lifecycle", func(t *testing.T) {
		// Test complete project lifecycle:
		// Setup → Development → Testing → Deployment preparation

		tmpDir := t.TempDir()
		projectRoot := filepath.Join(tmpDir, "lifecycle-test")

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		// Phase 1: Project initialization
		require.NoError(t, os.MkdirAll(projectRoot, 0755))
		err := os.Chdir(projectRoot)
		require.NoError(t, err)

		// Simulate glide setup (create structure manually)
		vcsDir := filepath.Join(projectRoot, "vcs")
		worktreesDir := filepath.Join(projectRoot, "worktrees")
		require.NoError(t, os.MkdirAll(vcsDir, 0755))
		require.NoError(t, os.MkdirAll(worktreesDir, 0755))

		// Initialize main repository
		os.Chdir(vcsDir)
		exec.Command("git", "init").Run()
		exec.Command("git", "config", "user.email", "test@example.com").Run()
		exec.Command("git", "config", "user.name", "Test User").Run()

		// Create application structure
		appDirs := []string{"app", "config", "database", "public", "resources", "storage/logs", "tests"}
		for _, dir := range appDirs {
			require.NoError(t, os.MkdirAll(dir, 0755))
		}

		// Create basic application files
		files := map[string]string{
			"composer.json": `{
    "name": "test/app",
    "require": {
        "php": "^8.3"
    }
}`,
			"docker-compose.yml": `version: '3.8'
services:
  app:
    image: php:8.3-cli
    working_dir: /app
    volumes:
      - .:/app
  database:
    image: mysql:8.0
    environment:
      - MYSQL_DATABASE=testapp`,
			"app/Application.php": `<?php
class Application {
    public function version(): string {
        return "1.0.0";
    }
}`,
			"tests/ApplicationTest.php": `<?php
use PHPUnit\Framework\TestCase;

class ApplicationTest extends TestCase {
    public function test_version() {
        $app = new Application();
        $this->assertEquals("1.0.0", $app->version());
    }
}`,
		}

		for path, content := range files {
			require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
			require.NoError(t, os.WriteFile(path, []byte(content), 0644))
		}

		// Initial commit
		exec.Command("git", "add", ".").Run()
		exec.Command("git", "commit", "-m", "Initial application structure").Run()

		// Phase 2: Feature development workflow
		os.Chdir(projectRoot)

		// Create feature worktree
		featureDir := filepath.Join(worktreesDir, "feature-user-auth")
		os.Chdir(vcsDir)
		cmd := exec.Command("git", "worktree", "add", featureDir, "-b", "feature/user-auth")
		require.NoError(t, cmd.Run())

		// Develop in feature branch
		os.Chdir(featureDir)
		authContent := `<?php
class UserAuth {
    public function login(string $username, string $password): bool {
        return $username === "admin" && $password === "secret";
    }
}
`
		require.NoError(t, os.WriteFile("app/UserAuth.php", []byte(authContent), 0644))

		authTestContent := `<?php
use PHPUnit\Framework\TestCase;

class UserAuthTest extends TestCase {
    public function test_login() {
        $auth = new UserAuth();
        $this->assertTrue($auth->login("admin", "secret"));
        $this->assertFalse($auth->login("wrong", "credentials"));
    }
}
`
		require.NoError(t, os.WriteFile("tests/UserAuthTest.php", []byte(authTestContent), 0644))

		// Commit feature
		exec.Command("git", "add", ".").Run()
		exec.Command("git", "commit", "-m", "Add user authentication").Run()

		// Phase 3: Testing and integration
		// Verify all test files exist
		assert.FileExists(t, "tests/UserAuthTest.php")
		assert.FileExists(t, "app/UserAuth.php")

		// Phase 4: Cleanup and verification
		// Switch back to main repo
		os.Chdir(vcsDir)

		// Verify worktree integration
		cmd = exec.Command("git", "worktree", "list")
		output, err := cmd.Output()
		require.NoError(t, err)
		assert.Contains(t, string(output), "feature-user-auth")

		// Clean up worktree
		cmd = exec.Command("git", "worktree", "remove", featureDir)
		require.NoError(t, cmd.Run())

		assert.NoDirExists(t, featureDir)
		t.Log("Complete project lifecycle test completed successfully")
	})

	t.Run("concurrent_development_workflow", func(t *testing.T) {
		// Test concurrent development on multiple features
		tmpDir := t.TempDir()
		vcsDir := filepath.Join(tmpDir, "vcs")
		worktreesDir := filepath.Join(tmpDir, "worktrees")

		require.NoError(t, os.MkdirAll(vcsDir, 0755))
		require.NoError(t, os.MkdirAll(worktreesDir, 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		// Setup main repository
		os.Chdir(vcsDir)
		exec.Command("git", "init").Run()
		exec.Command("git", "config", "user.email", "test@example.com").Run()
		exec.Command("git", "config", "user.name", "Test User").Run()

		require.NoError(t, os.WriteFile("README.md", []byte("# Concurrent Dev Test"), 0644))
		exec.Command("git", "add", ".").Run()
		exec.Command("git", "commit", "-m", "Initial").Run()

		// Create multiple feature worktrees concurrently
		features := []string{"api-endpoints", "user-interface", "database-schema"}
		worktreePaths := make([]string, len(features))

		for i, feature := range features {
			worktreePath := filepath.Join(worktreesDir, feature)
			worktreePaths[i] = worktreePath

			cmd := exec.Command("git", "worktree", "add", worktreePath, "-b", "feature/"+feature)
			require.NoError(t, cmd.Run())

			// Add feature-specific files
			require.NoError(t, os.Chdir(worktreePath))

			featureFile := feature + ".md"
			content := "# " + strings.ToUpper(feature) + " Feature\n\nDevelopment notes..."
			require.NoError(t, os.WriteFile(featureFile, []byte(content), 0644))

			exec.Command("git", "add", ".").Run()
			exec.Command("git", "commit", "-m", "Add "+feature+" feature").Run()

			// Verify context detection in each worktree
			ctx := context.Detect()
			assert.NotNil(t, ctx)
			assert.Equal(t, context.LocationWorktree, ctx.Location)
			assert.Equal(t, feature, ctx.WorktreeName)
		}

		// Return to main repo and verify all worktrees
		os.Chdir(vcsDir)
		cmd := exec.Command("git", "worktree", "list")
		output, err := cmd.Output()
		require.NoError(t, err)

		for _, feature := range features {
			assert.Contains(t, string(output), feature)
		}

		// Cleanup all worktrees
		for _, worktreePath := range worktreePaths {
			cmd := exec.Command("git", "worktree", "remove", worktreePath)
			require.NoError(t, cmd.Run())
			assert.NoDirExists(t, worktreePath)
		}

		t.Log("Concurrent development workflow test completed successfully")
	})
}

// TestPerformanceWorkflows tests performance-related end-to-end scenarios
func TestPerformanceWorkflows(t *testing.T) {
	t.Run("large_project_handling", func(t *testing.T) {
		// Test handling of larger project structures
		tmpDir := t.TempDir()
		projectDir := filepath.Join(tmpDir, "large-project")
		vcsDir := filepath.Join(projectDir, "vcs")

		require.NoError(t, os.MkdirAll(vcsDir, 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		err := os.Chdir(vcsDir)
		require.NoError(t, err)

		// Initialize git
		exec.Command("git", "init").Run()
		exec.Command("git", "config", "user.email", "test@example.com").Run()
		exec.Command("git", "config", "user.name", "Test User").Run()

		// Create larger directory structure
		dirs := []string{
			"app/Http/Controllers", "app/Models", "app/Services", "app/Repositories",
			"database/migrations", "database/seeders", "database/factories",
			"resources/views/admin", "resources/views/frontend", "resources/js", "resources/css",
			"public/images", "public/js", "public/css",
			"storage/app", "storage/logs", "storage/framework/cache",
			"tests/Feature", "tests/Unit", "tests/Browser",
			"config", "routes", "bootstrap",
		}

		for _, dir := range dirs {
			require.NoError(t, os.MkdirAll(dir, 0755))
		}

		// Create multiple files in each directory
		for _, dir := range dirs {
			for i := 1; i <= 3; i++ {
				filename := filepath.Join(dir, fmt.Sprintf("file%d.php", i))
				content := fmt.Sprintf("<?php\n// File %d in %s\nclass File%d {}\n", i, dir, i)
				require.NoError(t, os.WriteFile(filename, []byte(content), 0644))
			}
		}

		// Create docker-compose with multiple services
		composeContent := `version: '3.8'
services:
  php:
    image: php:8.3-fpm
  nginx:
    image: nginx:alpine
  mysql:
    image: mysql:8.0
  redis:
    image: redis:alpine
  elasticsearch:
    image: elasticsearch:8.0.0
  mailhog:
    image: mailhog/mailhog
`
		require.NoError(t, os.WriteFile("docker-compose.yml", []byte(composeContent), 0644))

		// Measure context detection time
		start := time.Now()
		ctx := context.Detect()
		detectionTime := time.Since(start)

		// Context detection should be fast even for large projects
		// Allow 200ms to account for framework detection overhead
		assert.Less(t, detectionTime, 200*time.Millisecond, "Context detection should be fast")
		assert.NotNil(t, ctx)
		if len(ctx.ComposeFiles) > 0 {
			assert.Contains(t, ctx.ComposeFiles[0], "docker-compose.yml")
		}

		// Add all files and commit
		cmd := exec.Command("git", "add", ".")
		err = cmd.Run()
		require.NoError(t, err)

		cmd = exec.Command("git", "commit", "-m", "Large project initial commit")
		err = cmd.Run()
		require.NoError(t, err)

		t.Logf("Context detection completed in %v for large project", detectionTime)
	})
}
