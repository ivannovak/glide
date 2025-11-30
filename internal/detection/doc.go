// Package detection provides file and environment detection utilities.
//
// This package contains helpers for detecting project characteristics
// based on file presence, content patterns, and environment signals.
//
// # File Detection
//
// Check for specific files:
//
//	if detection.FileExists("package.json") {
//	    // Node.js project
//	}
//
//	if detection.FileExists("go.mod") {
//	    // Go module
//	}
//
// # Pattern Detection
//
// Detect project type by patterns:
//
//	projectType := detection.DetectProjectType(".")
//	switch projectType {
//	case detection.ProjectTypeGo:
//	    // Go project
//	case detection.ProjectTypeNode:
//	    // Node.js project
//	case detection.ProjectTypePython:
//	    // Python project
//	}
//
// # Marker Files
//
// Common marker files for detection:
//
//	var MarkerFiles = map[string]ProjectType{
//	    "go.mod":        ProjectTypeGo,
//	    "package.json":  ProjectTypeNode,
//	    "Cargo.toml":    ProjectTypeRust,
//	    "pyproject.toml": ProjectTypePython,
//	}
//
// # Environment Detection
//
// Detect environment characteristics:
//
//	if detection.IsCI() {
//	    // Running in CI environment
//	}
//
//	if detection.IsDocker() {
//	    // Running inside Docker
//	}
//
// # Integration with Context
//
// Detection is used by the context package:
//
//	detector := context.NewDetector()
//	// Uses detection utilities internally
package detection
