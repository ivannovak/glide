package php

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/ivannovak/glide/pkg/plugin/sdk"
)

// PHPDetector detects PHP projects
type PHPDetector struct {
	*sdk.BaseFrameworkDetector
}

// ComposerJSON represents composer.json structure
type ComposerJSON struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Require     map[string]string      `json:"require"`
	RequireDev  map[string]string      `json:"require-dev"`
	Scripts     map[string]interface{} `json:"scripts"`
	Autoload    map[string]interface{} `json:"autoload"`
	Config      map[string]interface{} `json:"config"`
	Extra       map[string]interface{} `json:"extra"`
}

// NewPHPDetector creates a new PHP detector
func NewPHPDetector() *PHPDetector {
	detector := &PHPDetector{
		BaseFrameworkDetector: sdk.NewBaseFrameworkDetector(sdk.FrameworkInfo{
			Name: "php",
			Type: "language",
		}),
	}

	// Set detection patterns - composer.json is the primary indicator
	// Keep optional items minimal to ensure detection with just composer.json
	detector.SetPatterns(sdk.DetectionPatterns{
		RequiredFiles: []string{"composer.json"},
		OptionalFiles: []string{
			"composer.lock",
		},
		Directories: []string{"vendor"},
		Extensions:  []string{".php"},
	})

	// Set default commands
	detector.SetCommands(map[string]sdk.CommandDefinition{
		"install": {
			Cmd:         "composer install",
			Description: "Install dependencies",
			Category:    "dependencies",
		},
		"update": {
			Cmd:         "composer update",
			Description: "Update dependencies",
			Category:    "dependencies",
		},
		"autoload": {
			Cmd:         "composer dump-autoload",
			Description: "Regenerate autoloader",
			Category:    "dependencies",
		},
		"test": {
			Cmd:         "composer test",
			Description: "Run tests",
			Category:    "test",
		},
		"test:unit": {
			Cmd:         "vendor/bin/phpunit",
			Description: "Run PHPUnit tests",
			Category:    "test",
		},
		"lint": {
			Cmd:         "composer lint",
			Description: "Lint PHP code",
			Category:    "lint",
		},
		"analyze": {
			Cmd:         "composer analyze",
			Description: "Run static analysis",
			Category:    "lint",
		},
		"format": {
			Cmd:         "composer format",
			Description: "Format PHP code",
			Category:    "format",
		},
		"require": {
			Cmd:         "composer require $@",
			Description: "Add a dependency",
			Category:    "dependencies",
		},
		"require:dev": {
			Cmd:         "composer require --dev $@",
			Description: "Add a dev dependency",
			Category:    "dependencies",
		},
		"remove": {
			Cmd:         "composer remove $@",
			Description: "Remove a dependency",
			Category:    "dependencies",
		},
		"outdated": {
			Cmd:         "composer outdated",
			Description: "Show outdated packages",
			Category:    "dependencies",
		},
		"show": {
			Cmd:         "composer show",
			Description: "Show installed packages",
			Category:    "dependencies",
		},
	})

	return detector
}

// Detect performs PHP-specific detection
func (d *PHPDetector) Detect(projectPath string) (*sdk.DetectionResult, error) {
	// Check for composer.json - the primary indicator of a PHP project
	composerPath := filepath.Join(projectPath, "composer.json")
	if _, err := os.Stat(composerPath); os.IsNotExist(err) {
		// No composer.json, not a PHP project
		return &sdk.DetectionResult{Detected: false}, nil
	}

	// Initialize result with basic detection
	result := &sdk.DetectionResult{
		Detected:   true,
		Confidence: 70, // Base confidence for having composer.json
		Framework: sdk.FrameworkInfo{
			Name: "php",
			Type: "language",
		},
		Commands: make(map[string]string),
		Metadata: make(map[string]string),
	}

	// Add default commands
	for name, def := range d.GetDefaultCommands() {
		result.Commands[name] = def.Cmd
	}

	// Increase confidence for additional indicators
	if _, err := os.Stat(filepath.Join(projectPath, "composer.lock")); err == nil {
		result.Confidence += 10
	}
	if _, err := os.Stat(filepath.Join(projectPath, "vendor")); err == nil {
		result.Confidence += 10
	}
	// Check for PHP files
	if d.hasFileWithExtension(projectPath, []string{".php"}) {
		result.Confidence += 10
	}
	result.Confidence = min(100, result.Confidence)

	// Read composer.json for enhanced detection
	composer, err := d.readComposerJSON(composerPath)
	if err != nil {
		return result, nil // Still detected, just can't read composer.json
	}

	// Detect PHP version requirement
	if phpReq, exists := composer.Require["php"]; exists {
		result.Framework.Version = d.parseVersionConstraint(phpReq)
		result.Metadata["php_version"] = phpReq
	}

	// Add project metadata
	if composer.Name != "" {
		result.Metadata["project_name"] = composer.Name
	}
	if composer.Type != "" {
		result.Metadata["project_type"] = composer.Type
	}

	// Detect frameworks
	frameworks := d.detectFrameworks(projectPath, composer)
	if len(frameworks) > 0 {
		result.Metadata["frameworks"] = strings.Join(frameworks, ",")
		// Add framework-specific commands
		d.addFrameworkCommands(result, frameworks)
		// Increase confidence for known frameworks
		result.Confidence = min(100, result.Confidence+len(frameworks)*10)
	}

	// Detect testing tools
	d.detectTestingTools(result, composer)

	// Detect code quality tools
	d.detectQualityTools(result, projectPath, composer)

	// Add composer script commands
	d.addComposerScripts(result, composer)

	return result, nil
}

// readComposerJSON reads and parses composer.json
func (d *PHPDetector) readComposerJSON(path string) (*ComposerJSON, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var composer ComposerJSON
	if err := json.Unmarshal(data, &composer); err != nil {
		return nil, err
	}

	return &composer, nil
}

// detectFrameworks detects specific PHP frameworks
func (d *PHPDetector) detectFrameworks(projectPath string, composer *ComposerJSON) []string {
	var frameworks []string

	// Check dependencies for frameworks
	allDeps := make(map[string]bool)
	for dep := range composer.Require {
		allDeps[dep] = true
	}
	for dep := range composer.RequireDev {
		allDeps[dep] = true
	}

	// Laravel
	if allDeps["laravel/framework"] || d.fileExists(projectPath, "artisan") {
		frameworks = append(frameworks, "laravel")
	}

	// Symfony
	if allDeps["symfony/framework-bundle"] || allDeps["symfony/console"] || d.fileExists(projectPath, "symfony.lock") {
		frameworks = append(frameworks, "symfony")
	}

	// Slim
	if allDeps["slim/slim"] {
		frameworks = append(frameworks, "slim")
	}

	// Lumen
	if allDeps["laravel/lumen-framework"] {
		frameworks = append(frameworks, "lumen")
	}

	// CodeIgniter
	if allDeps["codeigniter4/framework"] || d.fileExists(projectPath, "spark") {
		frameworks = append(frameworks, "codeigniter")
	}

	// Laminas/Zend
	if allDeps["laminas/laminas-mvc"] || allDeps["zendframework/zendframework"] {
		frameworks = append(frameworks, "laminas")
	}

	// Yii
	if allDeps["yiisoft/yii2"] || allDeps["yiisoft/yii-core"] {
		frameworks = append(frameworks, "yii")
	}

	// CakePHP
	if allDeps["cakephp/cakephp"] {
		frameworks = append(frameworks, "cakephp")
	}

	// WordPress
	if d.fileExists(projectPath, "wp-config.php") || d.fileExists(projectPath, "wp-load.php") {
		frameworks = append(frameworks, "wordpress")
	}

	// Drupal
	if allDeps["drupal/core"] || d.fileExists(projectPath, "core/lib/Drupal.php") {
		frameworks = append(frameworks, "drupal")
	}

	// Magento
	if allDeps["magento/product-community-edition"] || d.fileExists(projectPath, "app/Mage.php") {
		frameworks = append(frameworks, "magento")
	}

	return frameworks
}

// addFrameworkCommands adds framework-specific commands
func (d *PHPDetector) addFrameworkCommands(result *sdk.DetectionResult, frameworks []string) {
	for _, fw := range frameworks {
		switch fw {
		case "laravel":
			result.Commands["serve"] = "php artisan serve"
			result.Commands["migrate"] = "php artisan migrate"
			result.Commands["migrate:fresh"] = "php artisan migrate:fresh"
			result.Commands["db:seed"] = "php artisan db:seed"
			result.Commands["cache:clear"] = "php artisan cache:clear"
			result.Commands["route:list"] = "php artisan route:list"
			result.Commands["tinker"] = "php artisan tinker"
			result.Commands["queue:work"] = "php artisan queue:work"
			result.Commands["make:model"] = "php artisan make:model $1"
			result.Commands["make:controller"] = "php artisan make:controller $1"
			result.Commands["make:migration"] = "php artisan make:migration $1"

		case "symfony":
			result.Commands["serve"] = "symfony serve"
			result.Commands["console"] = "php bin/console"
			result.Commands["cache:clear"] = "php bin/console cache:clear"
			result.Commands["debug:router"] = "php bin/console debug:router"
			result.Commands["make:controller"] = "php bin/console make:controller $1"
			result.Commands["make:entity"] = "php bin/console make:entity $1"
			result.Commands["doctrine:migrate"] = "php bin/console doctrine:migrations:migrate"

		case "wordpress":
			result.Commands["wp"] = "wp $@"
			result.Commands["wp:update"] = "wp core update"
			result.Commands["wp:plugin:list"] = "wp plugin list"
			result.Commands["wp:theme:list"] = "wp theme list"
			result.Commands["wp:cache:flush"] = "wp cache flush"

		case "drupal":
			result.Commands["drush"] = "vendor/bin/drush $@"
			result.Commands["cache:rebuild"] = "vendor/bin/drush cache:rebuild"
			result.Commands["update:db"] = "vendor/bin/drush updatedb"
			result.Commands["config:export"] = "vendor/bin/drush config:export"

		case "codeigniter":
			result.Commands["serve"] = "php spark serve"
			result.Commands["migrate"] = "php spark migrate"
			result.Commands["db:seed"] = "php spark db:seed"
		}
	}
}

// detectTestingTools detects PHP testing tools
func (d *PHPDetector) detectTestingTools(result *sdk.DetectionResult, composer *ComposerJSON) {
	allDeps := make(map[string]bool)
	for dep := range composer.Require {
		allDeps[dep] = true
	}
	for dep := range composer.RequireDev {
		allDeps[dep] = true
	}

	// PHPUnit
	if allDeps["phpunit/phpunit"] {
		result.Commands["test:unit"] = "vendor/bin/phpunit"
		result.Commands["test:coverage"] = "vendor/bin/phpunit --coverage-html coverage"
		result.Metadata["testing_framework"] = "phpunit"
	}

	// Pest
	if allDeps["pestphp/pest"] {
		result.Commands["test"] = "vendor/bin/pest"
		result.Commands["test:coverage"] = "vendor/bin/pest --coverage"
		result.Metadata["testing_framework"] = "pest"
	}

	// Codeception
	if allDeps["codeception/codeception"] {
		result.Commands["test:run"] = "vendor/bin/codecept run"
		result.Commands["test:unit"] = "vendor/bin/codecept run unit"
		result.Commands["test:functional"] = "vendor/bin/codecept run functional"
		result.Metadata["testing_framework"] = "codeception"
	}

	// Behat
	if allDeps["behat/behat"] {
		result.Commands["test:behat"] = "vendor/bin/behat"
		result.Metadata["bdd_framework"] = "behat"
	}

	// PHPSpec
	if allDeps["phpspec/phpspec"] {
		result.Commands["test:spec"] = "vendor/bin/phpspec run"
		result.Metadata["spec_framework"] = "phpspec"
	}
}

// detectQualityTools detects code quality tools
func (d *PHPDetector) detectQualityTools(result *sdk.DetectionResult, projectPath string, composer *ComposerJSON) {
	allDeps := make(map[string]bool)
	for dep := range composer.Require {
		allDeps[dep] = true
	}
	for dep := range composer.RequireDev {
		allDeps[dep] = true
	}

	// PHPStan
	if allDeps["phpstan/phpstan"] || d.fileExists(projectPath, "phpstan.neon") || d.fileExists(projectPath, "phpstan.neon.dist") {
		result.Commands["analyze:phpstan"] = "vendor/bin/phpstan analyze"
		result.Metadata["static_analysis"] = appendToList(result.Metadata["static_analysis"], "phpstan")
	}

	// Psalm
	if allDeps["vimeo/psalm"] || d.fileExists(projectPath, "psalm.xml") {
		result.Commands["analyze:psalm"] = "vendor/bin/psalm"
		result.Metadata["static_analysis"] = appendToList(result.Metadata["static_analysis"], "psalm")
	}

	// PHP CS Fixer
	if allDeps["friendsofphp/php-cs-fixer"] || d.fileExists(projectPath, ".php-cs-fixer.php") || d.fileExists(projectPath, ".php-cs-fixer.dist.php") {
		result.Commands["format:fix"] = "vendor/bin/php-cs-fixer fix"
		result.Commands["format:check"] = "vendor/bin/php-cs-fixer fix --dry-run --diff"
		result.Metadata["code_formatter"] = "php-cs-fixer"
	}

	// PHP CodeSniffer
	if allDeps["squizlabs/php_codesniffer"] {
		result.Commands["lint:phpcs"] = "vendor/bin/phpcs"
		result.Commands["lint:fix"] = "vendor/bin/phpcbf"
		result.Metadata["code_sniffer"] = "phpcs"
	}

	// PHPMD (PHP Mess Detector)
	if allDeps["phpmd/phpmd"] {
		result.Commands["analyze:phpmd"] = "vendor/bin/phpmd . text cleancode,codesize,controversial,design,naming,unusedcode"
		result.Metadata["mess_detector"] = "phpmd"
	}

	// Rector
	if allDeps["rector/rector"] {
		result.Commands["refactor"] = "vendor/bin/rector process"
		result.Commands["refactor:dry"] = "vendor/bin/rector process --dry-run"
		result.Metadata["refactoring_tool"] = "rector"
	}
}

// addComposerScripts adds custom composer scripts as commands
func (d *PHPDetector) addComposerScripts(result *sdk.DetectionResult, composer *ComposerJSON) {
	for name := range composer.Scripts {
		// Skip composer lifecycle scripts
		if strings.HasPrefix(name, "pre-") || strings.HasPrefix(name, "post-") {
			continue
		}

		// Add script as command if not already defined
		if _, exists := result.Commands[name]; !exists {
			result.Commands[name] = "composer " + name
		}
	}
}

// parseVersionConstraint extracts a simple version from composer constraint
func (d *PHPDetector) parseVersionConstraint(constraint string) string {
	// Remove common constraint prefixes
	constraint = strings.TrimPrefix(constraint, "^")
	constraint = strings.TrimPrefix(constraint, "~")
	constraint = strings.TrimPrefix(constraint, ">=")
	constraint = strings.TrimPrefix(constraint, ">")
	constraint = strings.TrimSpace(constraint)

	// Take the first part if there are multiple constraints
	parts := strings.Split(constraint, " ")
	if len(parts) > 0 {
		return parts[0]
	}

	return constraint
}

// fileExists checks if a file exists in the project
func (d *PHPDetector) fileExists(projectPath, filename string) bool {
	path := filepath.Join(projectPath, filename)
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// appendToList appends a value to a comma-separated list
func appendToList(existing, value string) string {
	if existing == "" {
		return value
	}
	return existing + "," + value
}

// hasFileWithExtension checks if any files with given extensions exist
func (d *PHPDetector) hasFileWithExtension(projectPath string, extensions []string) bool {
	for _, ext := range extensions {
		pattern := filepath.Join(projectPath, "*"+ext)
		matches, err := filepath.Glob(pattern)
		if err == nil && len(matches) > 0 {
			return true
		}
		// Also check in common directories
		for _, dir := range []string{"src", "app", "public", "tests"} {
			pattern = filepath.Join(projectPath, dir, "*"+ext)
			matches, err = filepath.Glob(pattern)
			if err == nil && len(matches) > 0 {
				return true
			}
		}
	}
	return false
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
