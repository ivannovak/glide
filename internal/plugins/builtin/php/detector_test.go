package php

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPHPDetector(t *testing.T) {
	t.Run("detects basic PHP project", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create composer.json
		composerJSON := `{
			"name": "test/project",
			"require": {
				"php": "^8.0"
			}
		}`
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "composer.json"), []byte(composerJSON), 0644))

		detector := NewPHPDetector()
		result, err := detector.Detect(tmpDir)

		require.NoError(t, err)
		require.NotNil(t, result, "result should not be nil")
		t.Logf("Detection result - Detected: %v, Confidence: %d, Framework: %+v",
			result.Detected, result.Confidence, result.Framework)
		assert.True(t, result.Detected)
		assert.Equal(t, "php", result.Framework.Name)
		assert.Equal(t, "8.0", result.Framework.Version)
		assert.Equal(t, "test/project", result.Metadata["project_name"])
	})

	t.Run("detects Laravel project", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create composer.json with Laravel
		composerJSON := `{
			"name": "laravel/laravel",
			"type": "project",
			"require": {
				"php": "^8.1",
				"laravel/framework": "^10.0"
			}
		}`
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "composer.json"), []byte(composerJSON), 0644))

		// Create artisan file
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "artisan"), []byte("#!/usr/bin/env php"), 0644))

		detector := NewPHPDetector()
		result, err := detector.Detect(tmpDir)

		require.NoError(t, err)
		assert.True(t, result.Detected)
		assert.Equal(t, "laravel", result.Metadata["frameworks"])
		assert.Contains(t, result.Commands, "serve")
		assert.Equal(t, "php artisan serve", result.Commands["serve"])
		assert.Contains(t, result.Commands, "migrate")
		assert.Contains(t, result.Commands, "tinker")
	})

	t.Run("detects Symfony project", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create composer.json with Symfony
		composerJSON := `{
			"name": "symfony/skeleton",
			"type": "project",
			"require": {
				"php": ">=8.1",
				"symfony/console": "6.3.*",
				"symfony/framework-bundle": "6.3.*"
			}
		}`
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "composer.json"), []byte(composerJSON), 0644))

		// Create symfony.lock
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "symfony.lock"), []byte("{}"), 0644))

		detector := NewPHPDetector()
		result, err := detector.Detect(tmpDir)

		require.NoError(t, err)
		assert.True(t, result.Detected)
		assert.Equal(t, "symfony", result.Metadata["frameworks"])
		assert.Contains(t, result.Commands, "console")
		assert.Equal(t, "php bin/console", result.Commands["console"])
	})

	t.Run("detects WordPress project", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create composer.json
		composerJSON := `{
			"name": "wordpress/project",
			"require": {
				"php": ">=7.4"
			}
		}`
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "composer.json"), []byte(composerJSON), 0644))

		// Create wp-config.php
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "wp-config.php"), []byte("<?php"), 0644))

		detector := NewPHPDetector()
		result, err := detector.Detect(tmpDir)

		require.NoError(t, err)
		assert.True(t, result.Detected)
		assert.Equal(t, "wordpress", result.Metadata["frameworks"])
		assert.Contains(t, result.Commands, "wp")
		assert.Contains(t, result.Commands, "wp:update")
	})

	t.Run("detects testing tools", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create composer.json with testing tools
		composerJSON := `{
			"name": "test/project",
			"require": {
				"php": "^8.0"
			},
			"require-dev": {
				"phpunit/phpunit": "^10.0",
				"phpstan/phpstan": "^1.0",
				"friendsofphp/php-cs-fixer": "^3.0"
			}
		}`
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "composer.json"), []byte(composerJSON), 0644))

		// Create phpstan.neon
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "phpstan.neon"), []byte("parameters:"), 0644))

		detector := NewPHPDetector()
		result, err := detector.Detect(tmpDir)

		require.NoError(t, err)
		assert.True(t, result.Detected)
		assert.Equal(t, "phpunit", result.Metadata["testing_framework"])
		assert.Contains(t, result.Metadata["static_analysis"], "phpstan")
		assert.Equal(t, "php-cs-fixer", result.Metadata["code_formatter"])
		assert.Contains(t, result.Commands, "test:unit")
		assert.Contains(t, result.Commands, "analyze:phpstan")
		assert.Contains(t, result.Commands, "format:fix")
	})

	t.Run("detects multiple frameworks", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create composer.json with multiple indicators
		composerJSON := `{
			"name": "multi/project",
			"require": {
				"php": "^8.0",
				"slim/slim": "^4.0"
			},
			"require-dev": {
				"pestphp/pest": "^2.0",
				"vimeo/psalm": "^5.0"
			}
		}`
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "composer.json"), []byte(composerJSON), 0644))

		detector := NewPHPDetector()
		result, err := detector.Detect(tmpDir)

		require.NoError(t, err)
		assert.True(t, result.Detected)
		assert.Equal(t, "slim", result.Metadata["frameworks"])
		assert.Equal(t, "pest", result.Metadata["testing_framework"])
		assert.Contains(t, result.Metadata["static_analysis"], "psalm")
	})

	t.Run("adds composer scripts as commands", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create composer.json with custom scripts
		composerJSON := `{
			"name": "test/project",
			"require": {
				"php": "^8.0"
			},
			"scripts": {
				"build": "echo 'Building...'",
				"deploy": "echo 'Deploying...'",
				"pre-install-cmd": "echo 'Pre-install'",
				"post-update-cmd": "echo 'Post-update'"
			}
		}`
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "composer.json"), []byte(composerJSON), 0644))

		detector := NewPHPDetector()
		result, err := detector.Detect(tmpDir)

		require.NoError(t, err)
		assert.True(t, result.Detected)
		assert.Contains(t, result.Commands, "build")
		assert.Equal(t, "composer build", result.Commands["build"])
		assert.Contains(t, result.Commands, "deploy")
		assert.Equal(t, "composer deploy", result.Commands["deploy"])
		// Should not include lifecycle scripts
		assert.NotContains(t, result.Commands, "pre-install-cmd")
		assert.NotContains(t, result.Commands, "post-update-cmd")
	})

	t.Run("fails without composer.json", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create PHP files but no composer.json
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "index.php"), []byte("<?php"), 0644))

		detector := NewPHPDetector()
		result, err := detector.Detect(tmpDir)

		require.NoError(t, err)
		assert.False(t, result.Detected)
	})

	t.Run("handles malformed composer.json gracefully", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create malformed composer.json
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "composer.json"), []byte("not json"), 0644))

		detector := NewPHPDetector()
		result, err := detector.Detect(tmpDir)

		require.NoError(t, err)
		// Should still detect as PHP project even if can't parse composer.json
		assert.True(t, result.Detected)
		assert.Equal(t, "php", result.Framework.Name)
	})

	t.Run("parses PHP version constraints", func(t *testing.T) {
		detector := NewPHPDetector()

		testCases := []struct {
			constraint string
			expected   string
		}{
			{"^8.0", "8.0"},
			{"~7.4", "7.4"},
			{">=8.1", "8.1"},
			{">7.0", "7.0"},
			{"8.0.0", "8.0.0"},
			{"^8.0 || ^9.0", "8.0"},
			{">=7.4 <8.0", "7.4"},
		}

		for _, tc := range testCases {
			result := detector.parseVersionConstraint(tc.constraint)
			assert.Equal(t, tc.expected, result, "Failed for constraint: %s", tc.constraint)
		}
	})
}