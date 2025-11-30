package node

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/ivannovak/glide/v3/pkg/plugin/sdk"
)

// NodeDetector detects Node.js projects
type NodeDetector struct {
	*sdk.BaseFrameworkDetector
}

// PackageJSON represents package.json structure
type PackageJSON struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Scripts         map[string]string `json:"scripts"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Engines         map[string]string `json:"engines"`
	Type            string            `json:"type"`
	Main            string            `json:"main"`
	Private         bool              `json:"private"`
	Workspaces      interface{}       `json:"workspaces"`
}

// NewNodeDetector creates a new Node.js detector
func NewNodeDetector() *NodeDetector {
	detector := &NodeDetector{
		BaseFrameworkDetector: sdk.NewBaseFrameworkDetector(sdk.FrameworkInfo{
			Name: "node",
			Type: "language",
		}),
	}

	// Set detection patterns
	detector.SetPatterns(sdk.DetectionPatterns{
		RequiredFiles: []string{"package.json"},
		OptionalFiles: []string{
			"package-lock.json",
			"yarn.lock",
			"pnpm-lock.yaml",
			"bun.lockb",
			".npmrc",
			".nvmrc",
			"tsconfig.json",
		},
		Directories: []string{"node_modules"},
		Extensions:  []string{".js", ".mjs", ".cjs", ".ts", ".jsx", ".tsx"},
	})

	// Set default commands (will be enhanced based on package.json)
	detector.SetCommands(map[string]sdk.CommandDefinition{
		"install": {
			Cmd:         "npm install",
			Description: "Install dependencies",
			Category:    "dependencies",
		},
		"test": {
			Cmd:         "npm test",
			Description: "Run tests",
			Category:    "test",
		},
		"build": {
			Cmd:         "npm run build",
			Description: "Build the project",
			Category:    "build",
		},
		"dev": {
			Cmd:         "npm run dev",
			Description: "Start development server",
			Category:    "run",
		},
		"start": {
			Cmd:         "npm start",
			Description: "Start the application",
			Category:    "run",
		},
	})

	return detector
}

// Detect performs Node.js-specific detection
func (d *NodeDetector) Detect(projectPath string) (*sdk.DetectionResult, error) {
	// First use base detection
	result, err := d.BaseFrameworkDetector.Detect(projectPath)
	if err != nil || !result.Detected {
		return result, err
	}

	// Read package.json for enhanced detection
	packagePath := filepath.Join(projectPath, "package.json")
	pkg, err := d.readPackageJSON(packagePath)
	if err != nil {
		return result, nil // Still detected, just can't read package.json
	}

	// Set Node version if available
	if engines, ok := pkg.Engines["node"]; ok {
		result.Framework.Version = engines
	}

	// Detect package manager
	packageManager := d.detectPackageManager(projectPath)
	result.Metadata["package_manager"] = packageManager

	// Update commands based on package manager
	d.updateCommandsForPackageManager(result, packageManager)

	// Add available scripts from package.json
	d.addScriptsAsCommands(result, pkg)

	// Detect frameworks
	d.detectFrameworks(result, pkg)

	// Add project metadata
	result.Metadata["project_name"] = pkg.Name
	result.Metadata["project_version"] = pkg.Version
	if pkg.Type != "" {
		result.Metadata["module_type"] = pkg.Type
	}
	if pkg.Private {
		result.Metadata["private"] = "true"
	}
	if pkg.Workspaces != nil {
		result.Metadata["workspaces"] = "true"
		result.Metadata["monorepo"] = "true"
	}

	return result, nil
}

// readPackageJSON reads and parses package.json
func (d *NodeDetector) readPackageJSON(path string) (*PackageJSON, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}

	return &pkg, nil
}

// detectPackageManager detects which package manager is being used
func (d *NodeDetector) detectPackageManager(projectPath string) string {
	// Check lock files
	if _, err := os.Stat(filepath.Join(projectPath, "yarn.lock")); err == nil {
		return "yarn"
	}
	if _, err := os.Stat(filepath.Join(projectPath, "pnpm-lock.yaml")); err == nil {
		return "pnpm"
	}
	if _, err := os.Stat(filepath.Join(projectPath, "bun.lockb")); err == nil {
		return "bun"
	}
	if _, err := os.Stat(filepath.Join(projectPath, "package-lock.json")); err == nil {
		return "npm"
	}

	// Default to npm
	return "npm"
}

// updateCommandsForPackageManager updates commands based on package manager
func (d *NodeDetector) updateCommandsForPackageManager(result *sdk.DetectionResult, pm string) {
	baseCommands := make(map[string]string)

	switch pm {
	case "yarn":
		baseCommands["install"] = "yarn install"
		baseCommands["test"] = "yarn test"
		baseCommands["build"] = "yarn build"
		baseCommands["dev"] = "yarn dev"
		baseCommands["start"] = "yarn start"
		baseCommands["add"] = "yarn add $@"
		baseCommands["remove"] = "yarn remove $@"
		baseCommands["upgrade"] = "yarn upgrade"

	case "pnpm":
		baseCommands["install"] = "pnpm install"
		baseCommands["test"] = "pnpm test"
		baseCommands["build"] = "pnpm build"
		baseCommands["dev"] = "pnpm dev"
		baseCommands["start"] = "pnpm start"
		baseCommands["add"] = "pnpm add $@"
		baseCommands["remove"] = "pnpm remove $@"
		baseCommands["update"] = "pnpm update"

	case "bun":
		baseCommands["install"] = "bun install"
		baseCommands["test"] = "bun test"
		baseCommands["build"] = "bun run build"
		baseCommands["dev"] = "bun run dev"
		baseCommands["start"] = "bun start"
		baseCommands["add"] = "bun add $@"
		baseCommands["remove"] = "bun remove $@"

	default: // npm
		baseCommands["install"] = "npm install"
		baseCommands["test"] = "npm test"
		baseCommands["build"] = "npm run build"
		baseCommands["dev"] = "npm run dev"
		baseCommands["start"] = "npm start"
		baseCommands["add"] = "npm install $@"
		baseCommands["remove"] = "npm uninstall $@"
		baseCommands["update"] = "npm update"
	}

	// Update result commands
	for name, cmd := range baseCommands {
		result.Commands[name] = cmd
	}
}

// addScriptsAsCommands adds package.json scripts as commands
func (d *NodeDetector) addScriptsAsCommands(result *sdk.DetectionResult, pkg *PackageJSON) {
	pm := result.Metadata["package_manager"]

	for name := range pkg.Scripts {
		// Skip if it's already a base command
		if _, exists := result.Commands[name]; exists {
			continue
		}

		// Create command based on package manager
		var cmd string
		switch pm {
		case "yarn":
			cmd = "yarn " + name
		case "pnpm":
			cmd = "pnpm " + name
		case "bun":
			cmd = "bun run " + name
		default:
			cmd = "npm run " + name
		}

		result.Commands[name] = cmd
	}
}

// detectFrameworks detects specific Node.js frameworks
func (d *NodeDetector) detectFrameworks(result *sdk.DetectionResult, pkg *PackageJSON) {
	frameworks := []string{}

	// Check dependencies for frameworks
	allDeps := make(map[string]bool)
	for dep := range pkg.Dependencies {
		allDeps[dep] = true
	}
	for dep := range pkg.DevDependencies {
		allDeps[dep] = true
	}

	// React
	if allDeps["react"] {
		frameworks = append(frameworks, "react")
		if allDeps["next"] {
			frameworks = append(frameworks, "nextjs")
		}
	}

	// Vue
	if allDeps["vue"] {
		frameworks = append(frameworks, "vue")
		if allDeps["nuxt"] {
			frameworks = append(frameworks, "nuxt")
		}
	}

	// Angular
	if allDeps["@angular/core"] {
		frameworks = append(frameworks, "angular")
	}

	// Express
	if allDeps["express"] {
		frameworks = append(frameworks, "express")
	}

	// NestJS
	if allDeps["@nestjs/core"] {
		frameworks = append(frameworks, "nestjs")
	}

	// TypeScript
	if allDeps["typescript"] {
		frameworks = append(frameworks, "typescript")
	}

	// Add detected frameworks to metadata
	if len(frameworks) > 0 {
		result.Metadata["frameworks"] = strings.Join(frameworks, ",")
		// Increase confidence
		result.Confidence = min(100, result.Confidence+len(frameworks)*5)
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
