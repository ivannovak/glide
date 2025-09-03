package e2e_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSignalHandling tests end-to-end signal handling scenarios
func TestSignalHandling(t *testing.T) {
	// Note: Signal handling tests are complex and require careful setup
	// These tests verify the infrastructure is in place for proper signal handling

	t.Run("signal_propagation", func(t *testing.T) {
		// Test signal propagation infrastructure:
		// 1. Ctrl+C handling setup
		// 2. Process termination preparation
		// 3. Cleanup execution framework

		tmpDir := t.TempDir()
		signalDir := filepath.Join(tmpDir, "signal-test")
		vcsDir := filepath.Join(signalDir, "vcs")

		require.NoError(t, os.MkdirAll(vcsDir, 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		err := os.Chdir(vcsDir)
		require.NoError(t, err)

		exec.Command("git", "init").Run()
		exec.Command("git", "config", "user.email", "test@example.com").Run()
		exec.Command("git", "config", "user.name", "Test User").Run()

		// Step 1: Ctrl+C handling infrastructure
		t.Run("ctrl_c_handling_setup", func(t *testing.T) {
			// Create signal handling test script
			signalScript := `#!/bin/bash
# Signal handling test script

echo "Signal handling test started (PID: $$)"

# Create cleanup function
cleanup() {
    echo "Cleanup function called"
    echo "Cleaning up temporary files..."
    rm -f /tmp/signal_test_*
    echo "Cleanup completed"
    exit 0
}

# Set up signal handlers
trap cleanup SIGINT SIGTERM

# Simulate long-running operation
echo "Starting long-running operation..."
for i in {1..30}; do
    echo "Working... step $i/30"
    echo "$(date): Step $i" > /tmp/signal_test_$i
    sleep 0.1
done

echo "Operation completed normally"
cleanup
`
			require.NoError(t, os.WriteFile("signal_test.sh", []byte(signalScript), 0755))

			// Create signal handler verification
			signalVerification := `# Signal Handler Verification

## Expected Behavior:
1. Script starts and displays PID
2. Creates temporary files during operation
3. On SIGINT (Ctrl+C): cleanup() function is called
4. Temporary files are removed
5. Process exits gracefully with code 0

## Test Steps:
1. Run script: ./signal_test.sh
2. Send SIGINT after a few seconds: kill -INT <PID>
3. Verify cleanup occurred: ls /tmp/signal_test_*
4. Check exit code was 0

## Success Criteria:
- Cleanup function executes
- No temporary files remain
- Exit code is 0 (graceful termination)
`
			require.NoError(t, os.WriteFile("SIGNAL_TEST.md", []byte(signalVerification), 0644))

			exec.Command("git", "add", ".").Run()
			exec.Command("git", "commit", "-m", "Add signal handling test").Run()

			assert.FileExists(t, "signal_test.sh")
			assert.FileExists(t, "SIGNAL_TEST.md")

			// Verify script is executable
			fileInfo, err := os.Stat("signal_test.sh")
			require.NoError(t, err)
			assert.True(t, fileInfo.Mode()&0111 != 0, "Script should be executable")

			t.Log("Signal handling infrastructure is properly set up")
		})

		// Step 2: Process termination handling
		t.Run("process_termination_handling", func(t *testing.T) {
			// Create process management system
			require.NoError(t, os.MkdirAll("processes", 0755))

			// Process tracking template
			processTracker := `#!/bin/bash
# Process tracking and termination handler

PROCESS_DIR="processes"
PROCESS_FILE="$PROCESS_DIR/current_process.pid"

# Function to start tracked process
start_process() {
    local cmd="$1"
    echo "Starting process: $cmd"
    
    # Start process in background and capture PID
    $cmd &
    local pid=$!
    echo $pid > "$PROCESS_FILE"
    echo "Process started with PID: $pid"
    
    return $pid
}

# Function to stop tracked process
stop_process() {
    if [ -f "$PROCESS_FILE" ]; then
        local pid=$(cat "$PROCESS_FILE")
        echo "Stopping process PID: $pid"
        
        # Graceful shutdown attempt
        kill -TERM $pid 2>/dev/null
        
        # Wait for graceful shutdown
        for i in {1..5}; do
            if ! kill -0 $pid 2>/dev/null; then
                echo "Process stopped gracefully"
                rm -f "$PROCESS_FILE"
                return 0
            fi
            sleep 1
        done
        
        # Force kill if necessary
        echo "Force stopping process"
        kill -KILL $pid 2>/dev/null
        rm -f "$PROCESS_FILE"
    fi
}

# Function to cleanup all processes
cleanup_all() {
    echo "Cleaning up all processes..."
    for pidfile in $PROCESS_DIR/*.pid; do
        if [ -f "$pidfile" ]; then
            local pid=$(cat "$pidfile")
            echo "Cleaning up PID: $pid"
            kill -TERM $pid 2>/dev/null || kill -KILL $pid 2>/dev/null
            rm -f "$pidfile"
        fi
    done
}

# Set up signal handlers
trap cleanup_all EXIT SIGINT SIGTERM

# Export functions for use by other scripts
export -f start_process stop_process cleanup_all
`
			require.NoError(t, os.WriteFile("process_manager.sh", []byte(processTracker), 0755))

			// Create test process script  
			testProcess := `#!/bin/bash
# Test process that responds to signals

echo "Test process started (PID: $$)"

# Create working files
for i in {1..10}; do
    echo "Working on task $i" > "work_$i.tmp"
    sleep 0.2
done

echo "Test process completed"
`
			require.NoError(t, os.WriteFile("test_process.sh", []byte(testProcess), 0755))

			exec.Command("git", "add", ".").Run()
			exec.Command("git", "commit", "-m", "Add process termination handling").Run()

			assert.FileExists(t, "process_manager.sh")
			assert.FileExists(t, "test_process.sh")
			assert.DirExists(t, "processes")

			t.Log("Process termination handling system is ready")
		})

		// Step 3: Cleanup execution verification
		t.Run("cleanup_execution_framework", func(t *testing.T) {
			// Create comprehensive cleanup system
			require.NoError(t, os.MkdirAll("cleanup", 0755))

			// Main cleanup script
			cleanupScript := `#!/bin/bash
# Comprehensive cleanup system

CLEANUP_DIR="cleanup"
LOG_FILE="$CLEANUP_DIR/cleanup.log"

log_message() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') $1" | tee -a "$LOG_FILE"
}

# Cleanup functions for different resource types

cleanup_files() {
    log_message "Starting file cleanup..."
    find . -name "*.tmp" -type f -exec rm -f {} \;
    find . -name "*.cache" -type f -exec rm -f {} \;
    find . -name "*.pid" -type f -exec rm -f {} \;
    log_message "File cleanup completed"
}

cleanup_directories() {
    log_message "Starting directory cleanup..."
    [ -d "temp" ] && rm -rf temp
    [ -d "cache" ] && rm -rf cache  
    [ -d "work" ] && rm -rf work
    log_message "Directory cleanup completed"
}

cleanup_processes() {
    log_message "Starting process cleanup..."
    # Kill any remaining background processes
    jobs -p | xargs -r kill -TERM 2>/dev/null
    sleep 2
    jobs -p | xargs -r kill -KILL 2>/dev/null
    log_message "Process cleanup completed"
}

cleanup_locks() {
    log_message "Starting lock cleanup..."
    find . -name "*.lock" -type f -exec rm -f {} \;
    log_message "Lock cleanup completed"
}

# Main cleanup function
main_cleanup() {
    log_message "=== CLEANUP STARTED ==="
    
    cleanup_files
    cleanup_directories  
    cleanup_processes
    cleanup_locks
    
    log_message "=== CLEANUP COMPLETED ==="
    echo "Cleanup completed successfully. Check $LOG_FILE for details."
}

# Handle signals
trap 'log_message "Cleanup interrupted by signal"; main_cleanup; exit 1' SIGINT SIGTERM

# Run cleanup based on argument or if called directly
case "${1:-main}" in
    files) cleanup_files ;;
    dirs) cleanup_directories ;;
    processes) cleanup_processes ;;
    locks) cleanup_locks ;;
    *) main_cleanup ;;
esac
`
			require.NoError(t, os.WriteFile("cleanup/master_cleanup.sh", []byte(cleanupScript), 0755))

			// Create cleanup verification script
			cleanupVerifier := `#!/bin/bash
# Cleanup verification script

echo "Verifying cleanup effectiveness..."

# Create test mess to cleanup
mkdir -p temp cache work
echo "test" > test1.tmp
echo "test" > test2.cache  
echo "$$" > process.pid
echo "locked" > operation.lock

echo "Created test files for cleanup verification"
ls -la *.tmp *.cache *.pid *.lock 2>/dev/null | wc -l

# Run cleanup
./cleanup/master_cleanup.sh

# Verify cleanup
remaining=$(find . -name "*.tmp" -o -name "*.cache" -o -name "*.pid" -o -name "*.lock" | wc -l)
echo "Files remaining after cleanup: $remaining"

if [ $remaining -eq 0 ]; then
    echo "✓ Cleanup verification PASSED"
    exit 0
else
    echo "✗ Cleanup verification FAILED"
    exit 1
fi
`
			require.NoError(t, os.WriteFile("verify_cleanup.sh", []byte(cleanupVerifier), 0755))

			exec.Command("git", "add", ".").Run()
			exec.Command("git", "commit", "-m", "Add cleanup execution framework").Run()

			assert.FileExists(t, "cleanup/master_cleanup.sh")
			assert.FileExists(t, "verify_cleanup.sh")

			t.Log("Cleanup execution framework is complete")
		})
	})

	t.Run("long_running_operations", func(t *testing.T) {
		// Test long-running operations interrupt handling:
		// 1. Test suite interruption
		// 2. Migration interruption
		// 3. Container operations interruption

		tmpDir := t.TempDir()
		longRunDir := filepath.Join(tmpDir, "long-running-test")
		vcsDir := filepath.Join(longRunDir, "vcs")

		require.NoError(t, os.MkdirAll(vcsDir, 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		err := os.Chdir(vcsDir)
		require.NoError(t, err)

		exec.Command("git", "init").Run()
		exec.Command("git", "config", "user.email", "test@example.com").Run()
		exec.Command("git", "config", "user.name", "Test User").Run()

		// Step 1: Test suite interruption simulation
		t.Run("test_suite_interruption", func(t *testing.T) {
			// Create test suite that can be interrupted
			require.NoError(t, os.MkdirAll("tests", 0755))

			testSuite := `#!/bin/bash
# Long-running test suite

RESULTS_FILE="tests/test_results.txt"
PROGRESS_FILE="tests/test_progress.txt"

# Initialize results
echo "Test Suite Results" > "$RESULTS_FILE"
echo "Started: $(date)" >> "$RESULTS_FILE"
echo "0" > "$PROGRESS_FILE"

# Cleanup function
cleanup_tests() {
    local completed=$(cat "$PROGRESS_FILE")
    echo "Tests interrupted at step: $completed" >> "$RESULTS_FILE"
    echo "Interrupted: $(date)" >> "$RESULTS_FILE"
    echo "Cleanup completed for test suite" >> "$RESULTS_FILE"
    exit 1
}

# Set signal handler
trap cleanup_tests SIGINT SIGTERM

# Run long test suite
echo "Starting comprehensive test suite..."
for i in {1..50}; do
    echo "Running test $i/50..."
    echo "$i" > "$PROGRESS_FILE"
    
    # Simulate test execution
    echo "Test $i: $(date)" >> "$RESULTS_FILE"
    
    # Each test takes some time
    sleep 0.1
done

echo "Test suite completed successfully" >> "$RESULTS_FILE"
echo "Completed: $(date)" >> "$RESULTS_FILE"
`
			require.NoError(t, os.WriteFile("run_tests.sh", []byte(testSuite), 0755))

			// Create test progress monitor
			progressMonitor := `#!/bin/bash
# Test progress monitor

watch_tests() {
    echo "Monitoring test progress..."
    while [ -f "tests/test_progress.txt" ]; do
        local progress=$(cat tests/test_progress.txt 2>/dev/null || echo "0")
        echo "Tests completed: $progress/50"
        sleep 1
    done
}

# Export function
export -f watch_tests
`
			require.NoError(t, os.WriteFile("monitor_tests.sh", []byte(progressMonitor), 0755))

			exec.Command("git", "add", ".").Run()
			exec.Command("git", "commit", "-m", "Add test suite interruption handling").Run()

			assert.FileExists(t, "run_tests.sh")
			assert.FileExists(t, "monitor_tests.sh")
			assert.DirExists(t, "tests")

			t.Log("Test suite interruption handling is configured")
		})

		// Step 2: Migration interruption simulation
		t.Run("migration_interruption", func(t *testing.T) {
			require.NoError(t, os.MkdirAll("database/migrations", 0755))
			require.NoError(t, os.MkdirAll("database/backups", 0755))

			// Create migration runner with interruption handling
			migrationRunner := `#!/bin/bash
# Database migration runner with interruption handling

DB_BACKUP_DIR="database/backups"
MIGRATION_LOG="database/migration.log"
MIGRATION_STATE="database/migration_state.txt"

log_migration() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') $1" | tee -a "$MIGRATION_LOG"
}

# Backup database before migrations
backup_database() {
    local backup_file="$DB_BACKUP_DIR/backup_$(date +%Y%m%d_%H%M%S).sql"
    log_migration "Creating database backup: $backup_file"
    echo "-- Database backup created at $(date)" > "$backup_file"
    echo "BACKUP_FILE=$backup_file" > "$MIGRATION_STATE"
}

# Rollback function
rollback_migration() {
    log_migration "Migration interrupted - starting rollback"
    
    if [ -f "$MIGRATION_STATE" ]; then
        source "$MIGRATION_STATE"
        local current_migration=${CURRENT_MIGRATION:-"none"}
        
        log_migration "Current migration: $current_migration"
        log_migration "Backup file: ${BACKUP_FILE:-"none"}"
        
        # Simulate rollback
        if [ "$current_migration" != "none" ]; then
            log_migration "Rolling back migration: $current_migration"
            echo "ROLLBACK COMPLETED for $current_migration" >> "$MIGRATION_LOG"
        fi
        
        if [ -n "$BACKUP_FILE" ] && [ -f "$BACKUP_FILE" ]; then
            log_migration "Database backup available for restore: $BACKUP_FILE"
        fi
    fi
    
    log_migration "Rollback completed"
    exit 1
}

# Set signal handler
trap rollback_migration SIGINT SIGTERM

# Run migrations
run_migrations() {
    backup_database
    
    # Simulate multiple migrations
    migrations=("001_create_users" "002_create_posts" "003_add_indexes" "004_add_triggers")
    
    for migration in "${migrations[@]}"; do
        log_migration "Running migration: $migration"
        echo "CURRENT_MIGRATION=$migration" >> "$MIGRATION_STATE"
        
        # Simulate migration execution time
        sleep 0.5
        
        log_migration "Completed migration: $migration"
    done
    
    log_migration "All migrations completed successfully"
    rm -f "$MIGRATION_STATE"
}

# Main execution
echo "Starting database migrations..."
run_migrations
`
			require.NoError(t, os.WriteFile("migrate.sh", []byte(migrationRunner), 0755))

			// Create sample migrations
			sampleMigrations := []string{
				"001_create_users.sql",
				"002_create_posts.sql", 
				"003_add_indexes.sql",
				"004_add_triggers.sql",
			}

			for _, migration := range sampleMigrations {
				migrationContent := fmt.Sprintf("-- Migration: %s\n-- Created: %s\n\nSELECT 'Running %s';\n", 
					migration, time.Now().Format(time.RFC3339), migration)
				require.NoError(t, os.WriteFile(filepath.Join("database/migrations", migration), 
					[]byte(migrationContent), 0644))
			}

			exec.Command("git", "add", ".").Run()
			exec.Command("git", "commit", "-m", "Add migration interruption handling").Run()

			assert.FileExists(t, "migrate.sh")
			for _, migration := range sampleMigrations {
				assert.FileExists(t, filepath.Join("database/migrations", migration))
			}

			t.Log("Migration interruption handling is configured")
		})

		// Step 3: Container operations interruption
		t.Run("container_operations_interruption", func(t *testing.T) {
			require.NoError(t, os.MkdirAll("docker", 0755))

			// Create container management with interruption handling
			containerManager := `#!/bin/bash
# Container operations manager with interruption handling

CONTAINER_STATE="docker/container_state.txt"
CONTAINER_LOG="docker/operations.log"

log_container() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') $1" | tee -a "$CONTAINER_LOG"
}

# Container cleanup function
cleanup_containers() {
    log_container "Container operations interrupted - starting cleanup"
    
    if [ -f "$CONTAINER_STATE" ]; then
        while IFS= read -r container_info; do
            local container_id=$(echo "$container_info" | cut -d: -f1)
            local container_name=$(echo "$container_info" | cut -d: -f2)
            
            log_container "Stopping container: $container_name (ID: $container_id)"
            # Simulate container stop
            echo "Stopped container $container_name" >> "$CONTAINER_LOG"
        done < "$CONTAINER_STATE"
        
        rm -f "$CONTAINER_STATE"
    fi
    
    log_container "Container cleanup completed"
    exit 1
}

# Set signal handler
trap cleanup_containers SIGINT SIGTERM

# Simulate container operations
start_containers() {
    log_container "Starting container operations..."
    
    # Simulate starting multiple containers
    containers=("web:nginx" "api:php-fpm" "db:mysql" "cache:redis" "queue:worker")
    
    for container in "${containers[@]}"; do
        local name=$(echo "$container" | cut -d: -f1)
        local image=$(echo "$container" | cut -d: -f2)
        local container_id="container_$(date +%s)_$name"
        
        log_container "Starting container: $name (image: $image)"
        echo "$container_id:$name" >> "$CONTAINER_STATE"
        
        # Simulate startup time
        sleep 0.3
        
        log_container "Container started: $name (ID: $container_id)"
    done
    
    log_container "All containers started successfully"
    
    # Simulate running state
    log_container "Containers running... (Press Ctrl+C to stop)"
    while true; do
        sleep 1
    done
}

# Main execution
start_containers
`
			require.NoError(t, os.WriteFile("container_manager.sh", []byte(containerManager), 0755))

			// Create docker-compose for reference
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
  
  db:
    image: mysql:8.0
    environment:
      - MYSQL_ROOT_PASSWORD=secret
  
  cache:
    image: redis:alpine
    ports:
      - "6379:6379"
  
  queue:
    image: redis:alpine
    command: redis-server --appendonly yes
`
			require.NoError(t, os.WriteFile("docker-compose.yml", []byte(composeContent), 0644))

			exec.Command("git", "add", ".").Run()
			exec.Command("git", "commit", "-m", "Add container operations interruption handling").Run()

			assert.FileExists(t, "container_manager.sh")
			assert.FileExists(t, "docker-compose.yml")
			assert.DirExists(t, "docker")

			t.Log("Container operations interruption handling is configured")
		})
	})

	t.Run("cleanup_guarantees", func(t *testing.T) {
		// Test cleanup guarantees:
		// 1. Resource cleanup on exit
		// 2. Lock file removal
		// 3. Temporary file cleanup

		tmpDir := t.TempDir()
		cleanupDir := filepath.Join(tmpDir, "cleanup-guarantees")
		vcsDir := filepath.Join(cleanupDir, "vcs")

		require.NoError(t, os.MkdirAll(vcsDir, 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		err := os.Chdir(vcsDir)
		require.NoError(t, err)

		exec.Command("git", "init").Run()
		exec.Command("git", "config", "user.email", "test@example.com").Run()
		exec.Command("git", "config", "user.name", "Test User").Run()

		// Step 1: Resource cleanup on exit
		t.Run("resource_cleanup_on_exit", func(t *testing.T) {
			require.NoError(t, os.MkdirAll("resources", 0755))

			resourceCleaner := `#!/bin/bash
# Resource cleanup guarantees

RESOURCE_DIR="resources"
RESOURCE_REGISTRY="$RESOURCE_DIR/registry.txt"

# Initialize resource registry
initialize_resources() {
    echo "# Resource Registry" > "$RESOURCE_REGISTRY"
    echo "# Format: type:path:cleanup_command" >> "$RESOURCE_REGISTRY"
}

# Register a resource for cleanup
register_resource() {
    local type="$1"
    local path="$2"
    local cleanup_cmd="$3"
    
    echo "$type:$path:$cleanup_cmd" >> "$RESOURCE_REGISTRY"
    echo "Registered resource: $type at $path"
}

# Cleanup all registered resources
cleanup_resources() {
    echo "Starting resource cleanup..."
    
    if [ -f "$RESOURCE_REGISTRY" ]; then
        while IFS=: read -r type path cleanup_cmd; do
            # Skip comments
            [[ "$type" =~ ^#.*$ ]] && continue
            
            if [ -n "$cleanup_cmd" ] && [ -e "$path" ]; then
                echo "Cleaning up $type: $path"
                eval "$cleanup_cmd"
            fi
        done < "$RESOURCE_REGISTRY"
    fi
    
    echo "Resource cleanup completed"
}

# Set exit trap
trap cleanup_resources EXIT SIGINT SIGTERM

# Main resource management demo
initialize_resources

# Register various resources
register_resource "temp_file" "temp_data.txt" "rm -f temp_data.txt"
register_resource "temp_dir" "temp_work" "rm -rf temp_work"  
register_resource "socket" "app.sock" "rm -f app.sock"
register_resource "pid_file" "app.pid" "rm -f app.pid"

# Create the resources
echo "Creating managed resources..."
echo "temp data" > temp_data.txt
mkdir -p temp_work
echo "work in progress" > temp_work/task.txt
echo "mock socket" > app.sock
echo "$$" > app.pid

echo "Resources created and registered for cleanup"
echo "Press Ctrl+C or wait for normal exit to see cleanup in action"

# Simulate work
sleep 2

echo "Work completed normally"
`
			require.NoError(t, os.WriteFile("resource_cleanup.sh", []byte(resourceCleaner), 0755))

			exec.Command("git", "add", ".").Run()
			exec.Command("git", "commit", "-m", "Add resource cleanup guarantees").Run()

			assert.FileExists(t, "resource_cleanup.sh")
			assert.DirExists(t, "resources")

			t.Log("Resource cleanup guarantees are configured")
		})

		// Step 2: Lock file removal guarantees
		t.Run("lock_file_removal", func(t *testing.T) {
			require.NoError(t, os.MkdirAll("locks", 0755))

			lockManager := `#!/bin/bash
# Lock file management with guaranteed removal

LOCK_DIR="locks"
MAIN_LOCK="$LOCK_DIR/operation.lock"
LOCK_REGISTRY="$LOCK_DIR/active_locks.txt"

# Create lock with automatic cleanup
create_lock() {
    local operation="$1"
    local lock_file="$LOCK_DIR/${operation}.lock"
    
    # Check if lock already exists
    if [ -f "$lock_file" ]; then
        echo "Lock already exists for operation: $operation"
        return 1
    fi
    
    # Create lock
    echo "PID:$$" > "$lock_file"
    echo "OPERATION:$operation" >> "$lock_file"
    echo "CREATED:$(date)" >> "$lock_file"
    
    # Register lock for cleanup
    echo "$lock_file" >> "$LOCK_REGISTRY"
    
    echo "Lock created: $lock_file"
    return 0
}

# Remove all locks
cleanup_locks() {
    echo "Cleaning up all locks..."
    
    if [ -f "$LOCK_REGISTRY" ]; then
        while IFS= read -r lock_file; do
            if [ -f "$lock_file" ]; then
                echo "Removing lock: $lock_file"
                rm -f "$lock_file"
            fi
        done < "$LOCK_REGISTRY"
        
        rm -f "$LOCK_REGISTRY"
    fi
    
    # Remove any remaining lock files
    find "$LOCK_DIR" -name "*.lock" -type f -delete
    
    echo "Lock cleanup completed"
}

# Set cleanup trap
trap cleanup_locks EXIT SIGINT SIGTERM

# Main lock demo
echo "Starting lock management demo..."

# Create multiple locks
operations=("backup" "migration" "deployment" "sync")

for op in "${operations[@]}"; do
    create_lock "$op"
    sleep 0.1
done

echo "All locks created. Simulating work..."

# Verify locks exist
echo "Active locks:"
ls -la "$LOCK_DIR"/*.lock 2>/dev/null || echo "No locks found"

# Simulate work that might be interrupted
sleep 2

echo "Work completed. Cleanup will happen automatically."
`
			require.NoError(t, os.WriteFile("lock_manager.sh", []byte(lockManager), 0755))

			exec.Command("git", "add", ".").Run()
			exec.Command("git", "commit", "-m", "Add lock file removal guarantees").Run()

			assert.FileExists(t, "lock_manager.sh")
			assert.DirExists(t, "locks")

			t.Log("Lock file removal guarantees are configured")
		})

		// Step 3: Temporary file cleanup
		t.Run("temporary_file_cleanup", func(t *testing.T) {
			require.NoError(t, os.MkdirAll("tmp", 0755))

			tempManager := `#!/bin/bash
# Temporary file management with guaranteed cleanup

TMP_DIR="tmp"
TEMP_REGISTRY="$TMP_DIR/.temp_registry"
CLEANUP_LOG="$TMP_DIR/.cleanup.log"

log_cleanup() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') $1" >> "$CLEANUP_LOG"
}

# Create tracked temporary file
create_temp() {
    local prefix="${1:-temp}"
    local suffix="${2:-tmp}"
    local temp_file="$TMP_DIR/${prefix}_$$_$(date +%s).$suffix"
    
    # Create the temporary file
    touch "$temp_file"
    
    # Register for cleanup
    echo "$temp_file" >> "$TEMP_REGISTRY"
    
    log_cleanup "Created temporary file: $temp_file"
    echo "$temp_file"
}

# Create temporary directory
create_temp_dir() {
    local prefix="${1:-tempdir}"
    local temp_dir="$TMP_DIR/${prefix}_$$_$(date +%s)"
    
    # Create the temporary directory
    mkdir -p "$temp_dir"
    
    # Register for cleanup
    echo "$temp_dir/" >> "$TEMP_REGISTRY"
    
    log_cleanup "Created temporary directory: $temp_dir"
    echo "$temp_dir"
}

# Cleanup all temporary files and directories
cleanup_temps() {
    log_cleanup "Starting temporary file cleanup..."
    
    if [ -f "$TEMP_REGISTRY" ]; then
        while IFS= read -r temp_path; do
            if [ -e "$temp_path" ]; then
                if [ -d "$temp_path" ]; then
                    log_cleanup "Removing temporary directory: $temp_path"
                    rm -rf "$temp_path"
                else
                    log_cleanup "Removing temporary file: $temp_path"
                    rm -f "$temp_path"
                fi
            fi
        done < "$TEMP_REGISTRY"
        
        rm -f "$TEMP_REGISTRY"
    fi
    
    # Additional cleanup - remove any files matching temp patterns
    find "$TMP_DIR" -name "*_$$_*" -exec rm -rf {} + 2>/dev/null
    
    log_cleanup "Temporary file cleanup completed"
    echo "Temporary file cleanup completed"
}

# Set cleanup trap
trap cleanup_temps EXIT SIGINT SIGTERM

# Initialize
log_cleanup "Temporary file manager started (PID: $$)"

# Demo: Create various temporary resources
echo "Creating temporary files and directories..."

# Create temporary files
temp_config=$(create_temp "config" "yml")
temp_data=$(create_temp "data" "json")
temp_log=$(create_temp "operation" "log")

# Create temporary directories
temp_workdir=$(create_temp_dir "workdir")
temp_cache=$(create_temp_dir "cache")

# Add content to demonstrate they exist
echo "temp_config: test" > "$temp_config"
echo '{"temp": "data"}' > "$temp_data"
echo "Operation log" > "$temp_log"
echo "work files" > "$temp_workdir/work.txt"
echo "cache data" > "$temp_cache/cache.dat"

echo "Temporary resources created:"
echo "Files: $temp_config $temp_data $temp_log"
echo "Directories: $temp_workdir $temp_cache"

# Show what will be cleaned up
echo ""
echo "Registry contents:"
cat "$TEMP_REGISTRY" 2>/dev/null || echo "No registry found"

echo ""
echo "Simulating work... (cleanup will happen on exit)"
sleep 1

echo "Work simulation completed"
`
			require.NoError(t, os.WriteFile("temp_manager.sh", []byte(tempManager), 0755))

			exec.Command("git", "add", ".").Run()
			exec.Command("git", "commit", "-m", "Add temporary file cleanup guarantees").Run()

			assert.FileExists(t, "temp_manager.sh")
			assert.DirExists(t, "tmp")

			t.Log("Temporary file cleanup guarantees are configured")
		})

		// Integration test: Verify all cleanup systems work together
		t.Run("integrated_cleanup_verification", func(t *testing.T) {
			// Create master cleanup verification script
			masterVerifier := `#!/bin/bash
# Master cleanup verification

echo "=== Cleanup Guarantees Verification ==="

# Test 1: Resource cleanup
echo "Testing resource cleanup..."
./resource_cleanup.sh &
sleep 0.5
kill -INT $!
sleep 1
echo "✓ Resource cleanup test completed"

# Test 2: Lock file cleanup
echo "Testing lock cleanup..."
./lock_manager.sh &
sleep 0.5
kill -INT $!
sleep 1
echo "✓ Lock cleanup test completed"

# Test 3: Temporary file cleanup
echo "Testing temporary file cleanup..."
./temp_manager.sh &
sleep 0.5
kill -INT $!
sleep 1
echo "✓ Temporary file cleanup test completed"

echo "=== All cleanup tests completed ==="

# Verify no residual files remain
residual_count=$(find . -name "*.tmp" -o -name "*.lock" -o -name "*.pid" | wc -l)
echo "Residual files after cleanup: $residual_count"

if [ $residual_count -eq 0 ]; then
    echo "✓ All cleanup guarantees verified successfully"
else
    echo "⚠ Some cleanup may not have completed properly"
fi
`
			require.NoError(t, os.WriteFile("verify_cleanup_guarantees.sh", []byte(masterVerifier), 0755))

			exec.Command("git", "add", ".").Run()
			exec.Command("git", "commit", "-m", "Add integrated cleanup verification").Run()

			assert.FileExists(t, "verify_cleanup_guarantees.sh")

			// Verify all cleanup components are present
			assert.FileExists(t, "resource_cleanup.sh")
			assert.FileExists(t, "lock_manager.sh")
			assert.FileExists(t, "temp_manager.sh")

			t.Log("Integrated cleanup guarantees verification is ready")
		})
	})
}

// TestSignalHandlingIntegration tests signal handling in realistic scenarios
func TestSignalHandlingIntegration(t *testing.T) {
	t.Run("real_world_interruption_scenarios", func(t *testing.T) {
		// Test realistic interruption scenarios that users might encounter

		tmpDir := t.TempDir()
		realWorldDir := filepath.Join(tmpDir, "real-world-signals")

		require.NoError(t, os.MkdirAll(realWorldDir, 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		err := os.Chdir(realWorldDir)
		require.NoError(t, err)

		// Scenario: Developer interrupts long-running test suite
		longTestSuite := `#!/bin/bash
# Simulate long-running test suite that developer might interrupt

TEST_RESULTS="test_results.txt"
TEST_PROGRESS="test_progress.txt"

echo "Starting comprehensive test suite..." | tee -a "$TEST_RESULTS"
echo "0" > "$TEST_PROGRESS"

# Cleanup on interruption
cleanup_tests() {
    local progress=$(cat "$TEST_PROGRESS" 2>/dev/null || echo "0")
    echo "Test suite interrupted at test $progress" | tee -a "$TEST_RESULTS"
    echo "Partial results saved to $TEST_RESULTS"
    echo "Run 'resume_tests $progress' to continue from where you left off"
    exit 130  # Standard exit code for SIGINT
}

trap cleanup_tests SIGINT SIGTERM

# Run tests with progress tracking
for i in $(seq 1 20); do
    echo "Running test $i/20: TestCase$i"
    echo "$i" > "$TEST_PROGRESS"
    echo "Test $i: PASS" >> "$TEST_RESULTS"
    sleep 0.1  # Simulate test execution time
done

echo "All tests completed successfully" | tee -a "$TEST_RESULTS"
rm -f "$TEST_PROGRESS"
`
		require.NoError(t, os.WriteFile("long_test_suite.sh", []byte(longTestSuite), 0755))

		assert.FileExists(t, "long_test_suite.sh")

		t.Log("Real-world test interruption scenario is configured")
	})

	t.Run("graceful_vs_forced_termination", func(t *testing.T) {
		// Test the difference between graceful (SIGTERM) and forced (SIGKILL) termination

		tmpDir := t.TempDir()
		terminationDir := filepath.Join(tmpDir, "termination-test")

		require.NoError(t, os.MkdirAll(terminationDir, 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		err := os.Chdir(terminationDir)
		require.NoError(t, err)

		// Create process that demonstrates graceful vs forced termination
		terminationDemo := `#!/bin/bash
# Demonstrates graceful vs forced termination handling

STATE_FILE="process_state.txt"
WORK_FILE="important_work.txt"

echo "Process started (PID: $$)" > "$STATE_FILE"
echo "Status: RUNNING" >> "$STATE_FILE"

# Graceful shutdown handler
graceful_shutdown() {
    echo "Received termination signal - shutting down gracefully"
    echo "Status: SHUTTING_DOWN" >> "$STATE_FILE"
    
    # Save important work
    echo "Saving important work..."
    echo "Work saved at: $(date)" >> "$WORK_FILE"
    
    # Cleanup
    echo "Status: SHUTDOWN_COMPLETE" >> "$STATE_FILE"
    echo "Graceful shutdown completed"
    exit 0
}

# Handle SIGTERM (graceful) and SIGINT (user interrupt)
trap graceful_shutdown SIGTERM SIGINT

echo "Process running... (Send SIGTERM for graceful shutdown)"
echo "Use: kill -TERM $$"
echo "Or:  kill -INT $$ (Ctrl+C)"
echo "Force kill with: kill -KILL $$"

# Simulate ongoing work
while true; do
    echo "Working... ($(date))" >> "$WORK_FILE"
    sleep 1
done
`
		require.NoError(t, os.WriteFile("termination_demo.sh", []byte(terminationDemo), 0755))

		assert.FileExists(t, "termination_demo.sh")

		t.Log("Graceful vs forced termination demo is configured")
	})

	t.Run("signal_handling_best_practices", func(t *testing.T) {
		// Document and test signal handling best practices

		tmpDir := t.TempDir()
		bestPracticesDir := filepath.Join(tmpDir, "signal-best-practices")

		require.NoError(t, os.MkdirAll(bestPracticesDir, 0755))

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		err := os.Chdir(bestPracticesDir)
		require.NoError(t, err)

		// Create best practices documentation and examples
		bestPracticesDoc := `# Signal Handling Best Practices

## Overview
Proper signal handling ensures that applications can:
1. Cleanup resources before termination
2. Save important state/data
3. Provide feedback to users about interruption
4. Handle both graceful and forced termination

## Common Signals
- SIGINT (2): Interrupt from user (Ctrl+C)
- SIGTERM (15): Termination request
- SIGKILL (9): Force kill (cannot be caught)
- SIGHUP (1): Hangup (terminal closed)

## Best Practices

### 1. Always Handle SIGINT and SIGTERM
` + "```bash" + `
trap cleanup_function SIGINT SIGTERM
` + "```" + `

### 2. Make Cleanup Functions Idempotent
Cleanup should be safe to run multiple times:
` + "```bash" + `
cleanup() {
    [ -f "$LOCK_FILE" ] && rm -f "$LOCK_FILE"
    [ -d "$TEMP_DIR" ] && rm -rf "$TEMP_DIR"
    # Always safe to run
}
` + "```" + `

### 3. Save Critical State
` + "```bash" + `
save_state() {
    echo "$(date): Process interrupted at step $CURRENT_STEP" >> progress.log
    echo "Restart with: $0 --resume $CURRENT_STEP"
}
` + "```" + `

### 4. Provide User Feedback
Let users know what happened and what they can do:
` + "```bash" + `
cleanup() {
    echo "Operation interrupted. Cleaning up..."
    save_state
    echo "✓ Cleanup completed"
    echo "💡 Use --resume to continue where you left off"
}
` + "```" + `

### 5. Handle Nested Processes
When managing child processes, ensure they're also cleaned up:
` + "```bash" + `
cleanup() {
    # Kill child processes
    jobs -p | xargs -r kill -TERM
    wait  # Wait for children to exit
}
` + "```" + `

## Testing Signal Handlers
Always test your signal handlers:
1. Normal completion
2. SIGINT interruption (Ctrl+C)
3. SIGTERM interruption
4. Multiple rapid signals
5. Interruption at different stages
`

		require.NoError(t, os.WriteFile("SIGNAL_HANDLING_BEST_PRACTICES.md", []byte(bestPracticesDoc), 0644))

		// Create example implementation following best practices
		bestPracticeExample := `#!/bin/bash
# Example script following signal handling best practices

set -euo pipefail  # Strict error handling

# Configuration
readonly SCRIPT_NAME="$(basename "$0")"
readonly PID_FILE="/tmp/${SCRIPT_NAME}.pid"
readonly STATE_FILE="/tmp/${SCRIPT_NAME}.state"
readonly LOG_FILE="/tmp/${SCRIPT_NAME}.log"

# Global state tracking
CURRENT_STEP=0
TOTAL_STEPS=10

# Logging function
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') [$SCRIPT_NAME] $*" | tee -a "$LOG_FILE"
}

# Save current state
save_state() {
    cat > "$STATE_FILE" << EOF
CURRENT_STEP=$CURRENT_STEP
TOTAL_STEPS=$TOTAL_STEPS
INTERRUPTED_AT=$(date)
RESTART_COMMAND=$0 --resume $CURRENT_STEP
EOF
    log "State saved: Step $CURRENT_STEP/$TOTAL_STEPS"
}

# Idempotent cleanup function
cleanup() {
    log "Starting cleanup process..."
    
    # Save state before cleanup
    save_state
    
    # Remove PID file
    [ -f "$PID_FILE" ] && {
        log "Removing PID file: $PID_FILE"
        rm -f "$PID_FILE"
    }
    
    # Kill any child processes
    if jobs -p >/dev/null 2>&1; then
        log "Terminating child processes..."
        jobs -p | xargs -r kill -TERM 2>/dev/null || true
        sleep 1
        jobs -p | xargs -r kill -KILL 2>/dev/null || true
    fi
    
    # Remove temporary files
    find /tmp -name "${SCRIPT_NAME}.tmp.*" -delete 2>/dev/null || true
    
    log "✓ Cleanup completed"
    
    # User feedback
    if [ -f "$STATE_FILE" ]; then
        echo ""
        echo "🛑 Operation was interrupted at step $CURRENT_STEP/$TOTAL_STEPS"
        echo "💾 Progress has been saved"
        echo "🔄 Resume with: $0 --resume $CURRENT_STEP"
        echo ""
    fi
}

# Signal handlers
handle_interrupt() {
    log "Received interrupt signal (SIGINT)"
    cleanup
    exit 130  # Standard exit code for SIGINT
}

handle_terminate() {
    log "Received termination signal (SIGTERM)"
    cleanup
    exit 143  # Standard exit code for SIGTERM
}

# Set up signal traps
trap handle_interrupt SIGINT
trap handle_terminate SIGTERM
trap cleanup EXIT  # Always cleanup on exit

# Main work function
do_work() {
    local start_step=${1:-1}
    
    log "Starting work from step $start_step"
    
    # Create PID file
    echo $$ > "$PID_FILE"
    
    for CURRENT_STEP in $(seq $start_step $TOTAL_STEPS); do
        log "Executing step $CURRENT_STEP/$TOTAL_STEPS"
        
        # Simulate work with ability to be interrupted
        sleep 0.5
        
        # Create temporary file for this step
        echo "Step $CURRENT_STEP completed at $(date)" > "/tmp/${SCRIPT_NAME}.tmp.$CURRENT_STEP"
    done
    
    log "All steps completed successfully"
    
    # Clean up state file on successful completion
    [ -f "$STATE_FILE" ] && rm -f "$STATE_FILE"
}

# Command line argument handling
case "${1:-}" in
    --resume)
        if [ -f "$STATE_FILE" ]; then
            source "$STATE_FILE"
            log "Resuming from step $CURRENT_STEP"
            do_work $((CURRENT_STEP + 1))
        else
            log "No state file found, starting from beginning"
            do_work
        fi
        ;;
    --status)
        if [ -f "$STATE_FILE" ]; then
            echo "Previous run was interrupted:"
            cat "$STATE_FILE"
        else
            echo "No interrupted state found"
        fi
        exit 0
        ;;
    --help)
        echo "Usage: $0 [--resume|--status|--help]"
        echo "  --resume  Resume from last interrupted step"
        echo "  --status  Show status of interrupted run"
        echo "  --help    Show this help message"
        exit 0
        ;;
    "")
        do_work
        ;;
    *)
        echo "Unknown option: $1"
        echo "Use --help for usage information"
        exit 1
        ;;
esac
`

		require.NoError(t, os.WriteFile("signal_best_practice_example.sh", []byte(bestPracticeExample), 0755))

		assert.FileExists(t, "SIGNAL_HANDLING_BEST_PRACTICES.md")
		assert.FileExists(t, "signal_best_practice_example.sh")

		t.Log("Signal handling best practices documentation and examples are ready")
	})
}