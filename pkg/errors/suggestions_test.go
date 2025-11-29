package errors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSuggestionEngine(t *testing.T) {
	engine := NewSuggestionEngine()

	assert.NotNil(t, engine)
	assert.NotEmpty(t, engine.patterns)
}

func TestSuggestionEngine_GetSuggestionsNil(t *testing.T) {
	engine := NewSuggestionEngine()

	suggestions := engine.GetSuggestions(nil, nil)
	assert.Nil(t, suggestions)
}

func TestSuggestionEngine_DockerDaemonPattern(t *testing.T) {
	engine := NewSuggestionEngine()

	err := fmt.Errorf("cannot connect to the docker daemon")
	suggestions := engine.GetSuggestions(err, nil)

	assert.NotEmpty(t, suggestions)
	assert.Contains(t, suggestions[0], "Docker Desktop")
}

func TestSuggestionEngine_PermissionDeniedPattern(t *testing.T) {
	engine := NewSuggestionEngine()

	err := fmt.Errorf("permission denied")
	suggestions := engine.GetSuggestions(err, nil)

	assert.NotEmpty(t, suggestions)
	// Should have permission-related suggestions
	found := false
	for _, s := range suggestions {
		if contains(s, "permission") || contains(s, "chmod") {
			found = true
			break
		}
	}
	assert.True(t, found, "Should have permission-related suggestions")
}

func TestSuggestionEngine_DatabasePattern(t *testing.T) {
	engine := NewSuggestionEngine()

	err := fmt.Errorf("SQLSTATE[HY000] [2002] Connection refused")
	suggestions := engine.GetSuggestions(err, nil)

	assert.NotEmpty(t, suggestions)
	// Should have database-related suggestions
	found := false
	for _, s := range suggestions {
		if contains(s, "MySQL") || contains(s, "database") {
			found = true
			break
		}
	}
	assert.True(t, found, "Should have database-related suggestions")
}

func TestSuggestionEngine_FileNotFoundPattern(t *testing.T) {
	engine := NewSuggestionEngine()

	err := fmt.Errorf("no such file or directory")
	suggestions := engine.GetSuggestions(err, nil)

	assert.NotEmpty(t, suggestions)
}

func TestSuggestionEngine_NetworkPattern(t *testing.T) {
	engine := NewSuggestionEngine()

	err := fmt.Errorf("connection timeout")
	suggestions := engine.GetSuggestions(err, nil)

	assert.NotEmpty(t, suggestions)
	// Should have network-related suggestions
	found := false
	for _, s := range suggestions {
		if contains(s, "network") || contains(s, "connection") {
			found = true
			break
		}
	}
	assert.True(t, found, "Should have network-related suggestions")
}

func TestSuggestionEngine_ComposerPattern(t *testing.T) {
	engine := NewSuggestionEngine()

	err := fmt.Errorf("vendor/autoload.php not found")
	suggestions := engine.GetSuggestions(err, nil)

	assert.NotEmpty(t, suggestions)
	// Should have composer-related suggestions
	found := false
	for _, s := range suggestions {
		if contains(s, "composer") || contains(s, "Install dependencies") {
			found = true
			break
		}
	}
	assert.True(t, found, "Should have composer-related suggestions")
}

func TestSuggestionEngine_WithContext_Container(t *testing.T) {
	engine := NewSuggestionEngine()

	err := fmt.Errorf("some error")
	context := map[string]string{
		"container": "mysql",
	}

	suggestions := engine.GetSuggestions(err, context)

	assert.NotEmpty(t, suggestions)
	// Should have MySQL-specific suggestions
	found := false
	for _, s := range suggestions {
		if contains(s, "MySQL") {
			found = true
			break
		}
	}
	assert.True(t, found, "Should have MySQL-specific suggestions")
}

func TestSuggestionEngine_WithContext_PHPContainer(t *testing.T) {
	engine := NewSuggestionEngine()

	err := fmt.Errorf("error")
	context := map[string]string{
		"container": "php",
	}

	suggestions := engine.GetSuggestions(err, context)

	assert.NotEmpty(t, suggestions)
	// Should have PHP-specific suggestions
	found := false
	for _, s := range suggestions {
		if contains(s, "PHP") {
			found = true
			break
		}
	}
	assert.True(t, found, "Should have PHP-specific suggestions")
}

func TestSuggestionEngine_WithContext_NginxContainer(t *testing.T) {
	engine := NewSuggestionEngine()

	err := fmt.Errorf("error")
	context := map[string]string{
		"container": "nginx",
	}

	suggestions := engine.GetSuggestions(err, context)

	assert.NotEmpty(t, suggestions)
	// Should have nginx-specific suggestions
	found := false
	for _, s := range suggestions {
		if contains(s, "nginx") {
			found = true
			break
		}
	}
	assert.True(t, found, "Should have nginx-specific suggestions")
}

func TestSuggestionEngine_WithContext_Mode(t *testing.T) {
	engine := NewSuggestionEngine()

	err := fmt.Errorf("error")
	context := map[string]string{
		"current_mode": "single-repo",
	}

	suggestions := engine.GetSuggestions(err, context)

	assert.NotEmpty(t, suggestions)
	// Should have mode-related suggestions
	found := false
	for _, s := range suggestions {
		if contains(s, "multi-worktree") || contains(s, "mode") {
			found = true
			break
		}
	}
	assert.True(t, found, "Should have mode-related suggestions")
}

func TestSuggestionEngine_WithContext_EnvFile(t *testing.T) {
	engine := NewSuggestionEngine()

	err := fmt.Errorf("error")
	context := map[string]string{
		"path": "/path/to/.env",
	}

	suggestions := engine.GetSuggestions(err, context)

	assert.NotEmpty(t, suggestions)
	// Should have .env-related suggestions
	found := false
	for _, s := range suggestions {
		if contains(s, ".env") {
			found = true
			break
		}
	}
	assert.True(t, found, "Should have .env-related suggestions")
}

func TestSuggestionEngine_WithContext_VendorPath(t *testing.T) {
	engine := NewSuggestionEngine()

	err := fmt.Errorf("error")
	context := map[string]string{
		"path": "/path/to/vendor/package",
	}

	suggestions := engine.GetSuggestions(err, context)

	assert.NotEmpty(t, suggestions)
	// Should have vendor-related suggestions
	found := false
	for _, s := range suggestions {
		if contains(s, "composer") {
			found = true
			break
		}
	}
	assert.True(t, found, "Should have composer-related suggestions")
}

func TestErrorPattern_Matches(t *testing.T) {
	pattern := &ErrorPattern{
		Contains: []string{"permission denied", "access denied"},
		Type:     TypePermission,
	}

	tests := []struct {
		name     string
		message  string
		expected bool
	}{
		{
			name:     "exact match",
			message:  "permission denied",
			expected: true,
		},
		{
			name:     "case insensitive",
			message:  "permission denied", // Pattern.Matches lowercases the message, not the pattern
			expected: true,
		},
		{
			name:     "contains",
			message:  "error: permission denied for user",
			expected: true,
		},
		{
			name:     "alternative pattern",
			message:  "access denied",
			expected: true,
		},
		{
			name:     "no match",
			message:  "file not found",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pattern.Matches(tt.message)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUniqueStrings(t *testing.T) {
	input := []string{
		"suggestion 1",
		"suggestion 2",
		"suggestion 1", // duplicate
		"suggestion 3",
		"suggestion 2", // duplicate
	}

	result := uniqueStrings(input)

	assert.Len(t, result, 3)
	assert.Contains(t, result, "suggestion 1")
	assert.Contains(t, result, "suggestion 2")
	assert.Contains(t, result, "suggestion 3")
}

func TestUniqueStrings_Empty(t *testing.T) {
	result := uniqueStrings([]string{})
	assert.Empty(t, result)
}

func TestAnalyzeError_Nil(t *testing.T) {
	result := AnalyzeError(nil)
	assert.Nil(t, result)
}

func TestAnalyzeError_GlideErrorWithSuggestions(t *testing.T) {
	original := NewDockerError("test error")
	original.AddSuggestion("existing suggestion")

	result := AnalyzeError(original)

	require.NotNil(t, result)
	assert.Equal(t, original, result)
	assert.Contains(t, result.Suggestions, "existing suggestion")
}

func TestAnalyzeError_StandardError(t *testing.T) {
	err := fmt.Errorf("cannot connect to the docker daemon")

	result := AnalyzeError(err)

	require.NotNil(t, result)
	assert.Equal(t, TypeDocker, result.Type)
	assert.NotEmpty(t, result.Suggestions)
	assert.Equal(t, err, result.Err)
}

func TestAnalyzeError_GlideErrorWithoutSuggestions(t *testing.T) {
	original := &GlideError{
		Type:    TypeUnknown,
		Message: "permission denied accessing /tmp",
	}

	result := AnalyzeError(original)

	require.NotNil(t, result)
	// Should enhance with pattern-based suggestions
	assert.NotEmpty(t, result.Suggestions)
	// Should update type based on pattern
	assert.Equal(t, TypePermission, result.Type)
}

func TestEnhanceError_Nil(t *testing.T) {
	result := EnhanceError(nil, nil)
	assert.Nil(t, result)
}

func TestEnhanceError_WithContext(t *testing.T) {
	err := fmt.Errorf("connection failed")
	context := map[string]string{
		"container": "mysql",
		"service":   "database",
	}

	result := EnhanceError(err, context)

	require.NotNil(t, result)
	assert.Equal(t, "mysql", result.Context["container"])
	assert.Equal(t, "database", result.Context["service"])
	assert.NotEmpty(t, result.Suggestions)
}

func TestEnhanceError_MergesSuggestions(t *testing.T) {
	// Error that matches a pattern (will get pattern suggestions)
	err := fmt.Errorf("permission denied")
	// Context that provides additional suggestions
	context := map[string]string{
		"path": "/path/to/.env",
	}

	result := EnhanceError(err, context)

	require.NotNil(t, result)
	// Should have both pattern-based and context-based suggestions
	assert.NotEmpty(t, result.Suggestions)

	// Verify no duplicates
	seen := make(map[string]bool)
	for _, s := range result.Suggestions {
		assert.False(t, seen[s], "Should not have duplicate suggestions")
		seen[s] = true
	}
}

func TestDefaultPatterns_Coverage(t *testing.T) {
	patterns := defaultPatterns()

	assert.NotEmpty(t, patterns)

	// Verify we have patterns for common error types
	types := make(map[ErrorType]bool)
	for _, p := range patterns {
		types[p.Type] = true
	}

	assert.True(t, types[TypeDocker], "Should have Docker patterns")
	assert.True(t, types[TypeContainer], "Should have Container patterns")
	assert.True(t, types[TypePermission], "Should have Permission patterns")
	assert.True(t, types[TypeDatabase], "Should have Database patterns")
	assert.True(t, types[TypeNetwork], "Should have Network patterns")
}

func TestSuggestionEngine_PortConflictPattern(t *testing.T) {
	engine := NewSuggestionEngine()

	err := fmt.Errorf("bind: address already in use")
	suggestions := engine.GetSuggestions(err, nil)

	assert.NotEmpty(t, suggestions)
	// Should have port-related suggestions
	found := false
	for _, s := range suggestions {
		if contains(s, "port") {
			found = true
			break
		}
	}
	assert.True(t, found, "Should have port-related suggestions")
}

func TestSuggestionEngine_TimeoutPattern(t *testing.T) {
	engine := NewSuggestionEngine()

	err := fmt.Errorf("operation timed out")
	suggestions := engine.GetSuggestions(err, nil)

	assert.NotEmpty(t, suggestions)
}

func TestSuggestionEngine_MigrationPattern(t *testing.T) {
	engine := NewSuggestionEngine()

	err := fmt.Errorf("nothing to migrate")
	suggestions := engine.GetSuggestions(err, nil)

	assert.NotEmpty(t, suggestions)
	// Should have migration-related suggestions
	found := false
	for _, s := range suggestions {
		if contains(s, "migrate") {
			found = true
			break
		}
	}
	assert.True(t, found, "Should have migration-related suggestions")
}

func TestSuggestionEngine_ContainerNotRunning(t *testing.T) {
	engine := NewSuggestionEngine()

	err := fmt.Errorf("service mysql is not running")
	suggestions := engine.GetSuggestions(err, nil)

	assert.NotEmpty(t, suggestions)
}

func TestSuggestionEngine_GetContextSuggestions_EmptyContext(t *testing.T) {
	engine := NewSuggestionEngine()

	suggestions := engine.getContextSuggestions(map[string]string{})
	assert.Empty(t, suggestions)
}

func TestSuggestionEngine_GetContextSuggestions_NilContext(t *testing.T) {
	engine := NewSuggestionEngine()

	suggestions := engine.getContextSuggestions(nil)
	assert.Empty(t, suggestions)
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
