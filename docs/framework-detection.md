# Framework Detection Plugin System

## Overview

Glide's Framework Detection Plugin System enables automatic detection of programming languages, frameworks, and tools in your projects. This system allows plugins to contribute detection patterns and automatically inject relevant commands when their framework is detected.

## Features

- **Automatic Detection**: Detects frameworks based on file patterns, directory structure, and file contents
- **Command Injection**: Automatically adds framework-specific commands to Glide
- **Parallel Detection**: Runs all detections concurrently for fast results
- **Caching**: Caches detection results to avoid repeated scanning
- **Extensible**: Easy to add new framework detectors via plugins

## How It Works

### Detection Flow

1. When you enter a project directory, Glide runs framework detection
2. All registered detectors scan the project in parallel
3. Each detector returns a confidence score (0-100)
4. Conflicts are resolved by keeping the highest confidence detection
5. Framework-specific commands are injected into the command registry
6. Results are cached for 5 minutes

### Detection Patterns

Detectors can look for:
- **Required Files**: Files that must exist (e.g., `go.mod` for Go)
- **Optional Files**: Files that increase confidence (e.g., `go.sum`)
- **Directories**: Directory patterns (e.g., `node_modules` for Node.js)
- **File Contents**: Patterns within files (e.g., `"react"` in package.json)
- **File Extensions**: Common file extensions (e.g., `.go`, `.js`)

## Built-in Detectors

### Go Detector
- **Detects**: Go projects
- **Required**: `go.mod`
- **Commands**: build, test, run, fmt, vet, mod:tidy, etc.

### Node.js Detector
- **Detects**: Node.js projects
- **Required**: `package.json`
- **Auto-detects**: npm, yarn, pnpm, or bun
- **Commands**: install, test, build, dev, start, etc.
- **Frameworks**: React, Vue, Angular, Express, NestJS, TypeScript

### PHP Detector
- **Detects**: PHP projects
- **Required**: `composer.json`
- **Commands**: install, update, test, lint, analyze, format, etc.
- **Frameworks**: Laravel, Symfony, WordPress, Drupal, CodeIgniter, Slim, Yii, CakePHP
- **Testing Tools**: PHPUnit, Pest, Codeception, Behat, PHPSpec
- **Quality Tools**: PHPStan, Psalm, PHP CS Fixer, PHP CodeSniffer, PHPMD, Rector
- **Framework Commands**:
  - Laravel: artisan commands (serve, migrate, tinker, etc.)
  - Symfony: console commands, doctrine migrations
  - WordPress: WP-CLI commands
  - Drupal: Drush commands

## Creating a Framework Detector

### 1. Implement the FrameworkDetector Interface

```go
package myplugin

import (
    "github.com/ivannovak/glide/pkg/plugin/sdk"
)

type MyDetector struct {
    *sdk.BaseFrameworkDetector
}

func NewMyDetector() *MyDetector {
    detector := &MyDetector{
        BaseFrameworkDetector: sdk.NewBaseFrameworkDetector(sdk.FrameworkInfo{
            Name: "myframework",
            Type: "framework", // or "language" or "tool"
        }),
    }

    // Set detection patterns
    detector.SetPatterns(sdk.DetectionPatterns{
        RequiredFiles: []string{"myconfig.yml"},
        OptionalFiles: []string{"mylock.yml"},
        Directories:   []string{".myframework"},
    })

    // Set default commands
    detector.SetCommands(map[string]sdk.CommandDefinition{
        "build": {
            Cmd:         "myframework build",
            Description: "Build the project",
            Category:    "build",
        },
        "test": {
            Cmd:         "myframework test",
            Description: "Run tests",
            Category:    "test",
        },
    })

    return detector
}
```

### 2. Custom Detection Logic (Optional)

For more complex detection, override the `Detect` method:

```go
func (d *MyDetector) Detect(projectPath string) (*sdk.DetectionResult, error) {
    // First use base detection
    result, err := d.BaseFrameworkDetector.Detect(projectPath)
    if err != nil || !result.Detected {
        return result, err
    }

    // Add custom detection logic
    version := d.detectVersion(projectPath)
    if version != "" {
        result.Framework.Version = version
        result.Metadata["version"] = version
    }

    // Adjust confidence based on additional checks
    if d.hasAdvancedFeatures(projectPath) {
        result.Confidence = min(100, result.Confidence + 20)
    }

    return result, nil
}
```

### 3. Register Your Detector

In your plugin's initialization:

```go
func (p *MyPlugin) Initialize() error {
    detector := NewMyDetector()
    // Register with the detection system
    p.RegisterDetector(detector)
    return nil
}
```

## Confidence Scoring

The base detector calculates confidence based on:
- Required files: 20 points each (must all exist)
- Optional files: 10 points each
- Directories: 10 points each
- File contents: 15 points each
- Extensions: 5 points total

A minimum of 50% confidence is required for detection.

## Context Display

View detected frameworks with:

```bash
glide context
```

Output includes:
```
Detected Frameworks:
  - go (v1.20)
  - docker

Available Framework Commands:
  - build
  - test
  - run
  - compose:up
  - compose:down
```

## Caching

Detection results are cached for 5 minutes per project path. The cache is automatically invalidated when:
- Files change in the project
- You manually clear the cache
- The TTL expires

## Performance

- Detection runs with a 100ms timeout per detector
- All detectors run in parallel
- Results are cached to avoid repeated detection
- Typical detection time: <50ms for most projects

## Best Practices

### For Plugin Authors

1. **Be Specific**: Use required files to avoid false positives
2. **Be Fast**: Keep detection logic simple and fast
3. **Provide Context**: Include version detection when possible
4. **Document Commands**: Provide clear descriptions for injected commands
5. **Handle Errors**: Gracefully handle missing files or permissions

### For Users

1. **Check Detection**: Use `glide context` to see what was detected
2. **Override When Needed**: Use `.glide.yml` to force specific frameworks
3. **Report Issues**: If detection fails, check the patterns and report bugs

## Configuration

Disable or force detection in `.glide.yml`:

```yaml
framework_detection:
  disable:
    - node    # Don't detect Node.js
  force:
    - go      # Always treat as Go project
```

## Troubleshooting

### Framework Not Detected

1. Check required files exist
2. Verify file permissions
3. Check detection patterns match your project
4. Increase log verbosity for details

### Wrong Framework Detected

1. Check confidence scores in verbose output
2. Disable incorrect detection in config
3. Force correct framework in config

### Commands Not Available

1. Verify framework was detected (`glide context`)
2. Check command injection succeeded
3. Verify no naming conflicts with existing commands

## Examples

### Multi-Framework Project

A project with both Go backend and React frontend:
- Detects both Go and Node.js
- Provides commands from both frameworks
- Higher confidence framework takes precedence for conflicts

### Monorepo Detection

In a monorepo, detection runs from current directory:
- Different directories may detect different frameworks
- Commands are contextual to current location
- Cache is per-directory for accuracy