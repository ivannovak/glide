package errors

import (
	"strings"
)

// SuggestionEngine provides smart error suggestions based on patterns
type SuggestionEngine struct {
	patterns []ErrorPattern
}

// ErrorPattern matches error messages and provides suggestions
type ErrorPattern struct {
	Contains    []string  // Any of these strings trigger the pattern
	Type        ErrorType // Error type to assign
	Suggestions []string  // Suggestions to provide
}

// NewSuggestionEngine creates a new suggestion engine with default patterns
func NewSuggestionEngine() *SuggestionEngine {
	return &SuggestionEngine{
		patterns: defaultPatterns(),
	}
}

// GetSuggestions analyzes an error and returns relevant suggestions
func (se *SuggestionEngine) GetSuggestions(err error, context map[string]string) []string {
	if err == nil {
		return nil
	}

	errMsg := strings.ToLower(err.Error())
	suggestions := []string{}

	// Check each pattern
	for _, pattern := range se.patterns {
		if pattern.Matches(errMsg) {
			suggestions = append(suggestions, pattern.Suggestions...)
		}
	}

	// Add context-specific suggestions
	if context != nil {
		suggestions = append(suggestions, se.getContextSuggestions(context)...)
	}

	// Remove duplicates
	return uniqueStrings(suggestions)
}

// getContextSuggestions provides suggestions based on context
func (se *SuggestionEngine) getContextSuggestions(context map[string]string) []string {
	var suggestions []string

	// Container-specific suggestions
	if container, ok := context["container"]; ok {
		switch container {
		case "php":
			suggestions = append(suggestions,
				"Ensure PHP container is running: glid status",
				"Start PHP container: glid up",
			)
		case "mysql":
			suggestions = append(suggestions,
				"Check MySQL container: glid docker ps",
				"Start MySQL: glid docker up -d mysql",
				"Check MySQL logs: glid logs mysql",
			)
		case "nginx":
			suggestions = append(suggestions,
				"Check nginx container: glid docker ps",
				"Start nginx: glid docker up -d nginx",
				"Check nginx config: glid docker exec nginx nginx -t",
			)
		}
	}

	// Mode-specific suggestions
	if mode, ok := context["current_mode"]; ok {
		if mode == "single-repo" {
			suggestions = append(suggestions,
				"Some commands require multi-worktree mode",
				"Run: glid setup to change mode",
			)
		}
	}

	// Path-specific suggestions
	if path, ok := context["path"]; ok {
		if strings.Contains(path, ".env") {
			suggestions = append(suggestions,
				"Check if .env file exists",
				"Copy from .env.example if available",
				"In worktrees, .env is copied from vcs/",
			)
		}
		if strings.Contains(path, "vendor") {
			suggestions = append(suggestions,
				"Run: glid composer install",
				"Check if composer.json exists",
			)
		}
	}

	return suggestions
}

// Matches checks if a pattern matches an error message
func (p *ErrorPattern) Matches(errMsg string) bool {
	for _, substr := range p.Contains {
		if strings.Contains(errMsg, strings.ToLower(substr)) {
			return true
		}
	}
	return false
}

// defaultPatterns returns the default error patterns
func defaultPatterns() []ErrorPattern {
	return []ErrorPattern{
		// Docker daemon errors
		{
			Contains: []string{"cannot connect to the docker daemon", "docker daemon", "docker.sock"},
			Type:     TypeDocker,
			Suggestions: []string{
				"Start Docker Desktop application",
				"Check: docker ps",
				"On Linux: sudo systemctl start docker",
				"On Mac: open -a Docker",
			},
		},
		// Container not running
		{
			Contains: []string{"service", "is not running", "container", "not found"},
			Type:     TypeContainer,
			Suggestions: []string{
				"Start all containers: glid up",
				"Check container status: glid status",
				"Start specific container: glid docker up -d [container]",
			},
		},
		// Permission denied
		{
			Contains: []string{"permission denied", "access denied", "operation not permitted"},
			Type:     TypePermission,
			Suggestions: []string{
				"Check file permissions: ls -la",
				"Fix Laravel permissions: chmod -R 775 storage bootstrap/cache",
				"Fix ownership: chown -R $(whoami) .",
				"On Linux, you may need sudo",
			},
		},
		// Database connection
		{
			Contains: []string{"sqlstate", "connection refused", "access denied for user", "unknown database"},
			Type:     TypeDatabase,
			Suggestions: []string{
				"Check MySQL container: glid status",
				"Verify .env database settings",
				"Ensure DB_HOST=mysql for Docker",
				"Start MySQL: glid docker up -d mysql",
				"Check credentials: glid mysql",
			},
		},
		// Composer/dependency errors
		{
			Contains: []string{"vendor/autoload.php", "class", "not found", "composer"},
			Type:     TypeDependency,
			Suggestions: []string{
				"Install dependencies: glid composer install",
				"Update autoloader: glid composer dump-autoload",
				"Clear cache: glid artisan cache:clear",
				"Check composer.json exists",
			},
		},
		// File not found
		{
			Contains: []string{"no such file", "file not found", "cannot find", "does not exist"},
			Type:     TypeFileNotFound,
			Suggestions: []string{
				"Check if file exists: ls -la",
				"Verify you're in the correct directory",
				"For .env: copy from .env.example",
				"For vendor: run glid composer install",
			},
		},
		// Network/connection errors
		{
			Contains: []string{"connection timeout", "network unreachable", "could not resolve", "connection reset"},
			Type:     TypeNetwork,
			Suggestions: []string{
				"Check internet connection",
				"Check Docker network: docker network ls",
				"Restart Docker: glid down && glid up",
				"Check firewall settings",
			},
		},
		// Port conflicts
		{
			Contains: []string{"address already in use", "port is already allocated", "bind: address"},
			Type:     TypeNetwork,
			Suggestions: []string{
				"Stop conflicting containers: glid down-all",
				"Check what's using the port: lsof -i :PORT",
				"Kill process using port: kill -9 PID",
				"Change port in docker-compose.yml",
			},
		},
		// Git errors
		{
			Contains: []string{"not a git repository", "git", "fatal:", "worktree"},
			Type:     TypeCommand,
			Suggestions: []string{
				"Initialize git: git init",
				"Check you're in the project root",
				"For worktrees: use glid g worktree",
			},
		},
		// AWS/ECR errors
		{
			Contains: []string{"ecr", "aws", "401 unauthorized", "no basic auth credentials"},
			Type:     TypeNetwork,
			Suggestions: []string{
				"Authenticate with ECR: glid ecr-login",
				"Check AWS credentials: aws configure list",
				"Verify AWS_PROFILE is set correctly",
			},
		},
		// Timeout errors
		{
			Contains: []string{"timeout", "timed out", "deadline exceeded"},
			Type:     TypeTimeout,
			Suggestions: []string{
				"Try running the command again",
				"Check if containers are healthy: glid status",
				"Increase timeout if configurable",
				"Check system resources: docker stats",
			},
		},
		// Migration errors
		{
			Contains: []string{"migration", "migrate", "nothing to migrate"},
			Type:     TypeDatabase,
			Suggestions: []string{
				"Run migrations: glid artisan migrate",
				"Fresh migration: glid artisan migrate:fresh",
				"Check migration status: glid artisan migrate:status",
				"For tests: glid artisan migrate --env=testing",
			},
		},
	}
}

// uniqueStrings removes duplicate strings from a slice
func uniqueStrings(strings []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, str := range strings {
		if !seen[str] {
			seen[str] = true
			result = append(result, str)
		}
	}

	return result
}

// AnalyzeError provides intelligent error analysis and suggestions
func AnalyzeError(err error) *GlideError {
	if err == nil {
		return nil
	}

	// If it's already a GlideError with suggestions, return it
	if glideErr, ok := err.(*GlideError); ok && glideErr.HasSuggestions() {
		return glideErr
	}

	// Get suggestions from the engine
	engine := NewSuggestionEngine()
	suggestions := engine.GetSuggestions(err, nil)

	// Determine error type from patterns
	errType := TypeUnknown
	errMsg := strings.ToLower(err.Error())
	for _, pattern := range engine.patterns {
		if pattern.Matches(errMsg) {
			errType = pattern.Type
			break
		}
	}

	// Create or enhance the error
	if glideErr, ok := err.(*GlideError); ok {
		// Enhance existing GlideError
		glideErr.Suggestions = append(glideErr.Suggestions, suggestions...)
		if glideErr.Type == TypeUnknown {
			glideErr.Type = errType
		}
		return glideErr
	}

	// Create new GlideError
	return New(errType, err.Error(),
		WithError(err),
		WithSuggestions(suggestions...),
	)
}

// EnhanceError adds contextual suggestions to an error
func EnhanceError(err error, context map[string]string) *GlideError {
	if err == nil {
		return nil
	}

	// Get base analysis
	glideErr := AnalyzeError(err)

	// Add context
	for k, v := range context {
		glideErr.AddContext(k, v)
	}

	// Get additional context-based suggestions
	engine := NewSuggestionEngine()
	contextSuggestions := engine.getContextSuggestions(context)

	// Merge suggestions
	glideErr.Suggestions = uniqueStrings(append(glideErr.Suggestions, contextSuggestions...))

	return glideErr
}
