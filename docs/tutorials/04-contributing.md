# Tutorial 4: Contributing to Glide

Welcome to the Glide contributor community! This tutorial guides you through setting up a development environment, understanding the codebase, making changes, and submitting pull requests.

## What You'll Learn

- Setting up the development environment
- Understanding the codebase structure
- Running tests and benchmarks
- Making and testing changes
- Submitting pull requests

## Prerequisites

- Go 1.21 or later
- Git
- Make (optional, but recommended)
- Familiarity with Go programming

## Step 1: Fork and Clone

### Fork the Repository

1. Go to https://github.com/glide-cli/glide
2. Click "Fork" to create your copy
3. Clone your fork:

```bash
git clone https://github.com/YOUR_USERNAME/glide.git
cd glide
```

### Set Up Remotes

```bash
# Add upstream remote
git remote add upstream https://github.com/glide-cli/glide.git

# Verify remotes
git remote -v
# origin    https://github.com/YOUR_USERNAME/glide.git (fetch)
# origin    https://github.com/YOUR_USERNAME/glide.git (push)
# upstream  https://github.com/glide-cli/glide.git (fetch)
# upstream  https://github.com/glide-cli/glide.git (push)
```

## Step 2: Development Environment

### Install Dependencies

```bash
# Install Go dependencies
go mod download

# Install development tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Build Glide

```bash
# Build the binary
go build -o glide ./cmd/glide

# Or use make
make build

# Verify the build
./glide version
```

### Run Tests

```bash
# Run all tests
go test ./...

# Run with race detection
go test -race ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test ./pkg/errors/...
```

## Step 3: Understand the Codebase

### Directory Structure

```
glide/
├── cmd/glide/          # Main entry point
├── pkg/                # Public packages (SDK, APIs)
│   ├── config/         # Type-safe configuration
│   ├── container/      # Dependency injection
│   ├── errors/         # Structured errors
│   ├── logging/        # Structured logging
│   ├── output/         # Output formatting
│   ├── plugin/         # Plugin system
│   │   └── sdk/        # Plugin SDK
│   │       ├── v1/     # Legacy SDK
│   │       └── v2/     # Current SDK
│   ├── registry/       # Generic registry
│   └── validation/     # Input validation
├── internal/           # Private packages
│   ├── cli/            # CLI commands
│   ├── config/         # Config loading
│   ├── context/        # Context detection
│   ├── docker/         # Docker integration
│   └── shell/          # Shell execution
├── tests/              # Integration tests
│   ├── integration/    # Integration tests
│   └── benchmarks/     # Performance tests
└── docs/               # Documentation
```

### Key Packages

| Package | Purpose |
|---------|---------|
| `pkg/container` | Dependency injection with uber-fx |
| `pkg/plugin/sdk/v2` | Plugin SDK for external plugins |
| `internal/context` | Project context detection |
| `internal/cli` | Command implementations |

### Read the Architecture

```bash
# Key documentation
cat docs/architecture/README.md
cat docs/adr/README.md
```

## Step 4: Making Changes

### Create a Feature Branch

```bash
# Sync with upstream
git fetch upstream
git checkout main
git merge upstream/main

# Create feature branch
git checkout -b feature/my-improvement
```

### Code Style

Follow these conventions:

```go
// Package documentation
// Package mypackage does something useful.
package mypackage

// Type documentation
// MyType represents something.
type MyType struct {
    // Field documentation
    Field string
}

// Function documentation
// DoSomething does something useful.
// It returns an error if something goes wrong.
func DoSomething(input string) error {
    // Implementation
}
```

### Error Handling

Use the structured error package:

```go
import "github.com/glide-cli/glide/v3/pkg/errors"

func LoadConfig(path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return errors.NewFileNotFoundError(path)
        }
        return errors.New(errors.TypeConfig, "failed to read config",
            errors.WithCause(err),
            errors.WithContext("path", path),
        )
    }
    return nil
}
```

### Testing

Write tests for your changes:

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:  "valid input",
            input: "hello",
            want:  "HELLO",
        },
        {
            name:    "empty input",
            input:   "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := MyFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("MyFunction() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("MyFunction() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Step 5: Validate Changes

### Run Linter

```bash
# Run golangci-lint
golangci-lint run

# Or use make
make lint
```

### Run Tests

```bash
# All tests
go test ./...

# With race detection
go test -race ./...

# With coverage
go test -coverprofile=coverage.out ./...
```

### Run Benchmarks

```bash
# Run benchmarks
go test -bench=. ./...

# Run specific benchmark
go test -bench=BenchmarkContextDetection ./internal/context/...
```

### Build and Test Locally

```bash
# Build
go build -o glide ./cmd/glide

# Test manually
./glide version
./glide context
./glide help
```

## Step 6: Commit Changes

### Commit Message Format

Follow conventional commits:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Formatting
- `refactor`: Code restructuring
- `perf`: Performance improvement
- `test`: Tests
- `chore`: Maintenance

Examples:

```bash
git commit -m "feat(plugin): add lifecycle hooks to SDK v2"
git commit -m "fix(context): handle symlink in project root detection"
git commit -m "docs(guides): add error handling guide"
```

### Stage and Commit

```bash
# Stage changes
git add .

# Check what's staged
git status
git diff --staged

# Commit
git commit -m "feat(errors): add network error type"
```

## Step 7: Submit Pull Request

### Push to Your Fork

```bash
git push origin feature/my-improvement
```

### Create Pull Request

1. Go to your fork on GitHub
2. Click "Compare & pull request"
3. Fill in the PR template:

```markdown
## Summary
Brief description of changes.

## Changes
- Added X
- Fixed Y
- Updated Z

## Testing
- [ ] Added unit tests
- [ ] Ran full test suite
- [ ] Tested manually

## Related Issues
Closes #123
```

### Respond to Reviews

- Address all feedback
- Push additional commits
- Re-request review when ready

## Common Tasks

### Add a New Command

1. Create command file in `internal/cli/`:

```go
// internal/cli/mycommand.go
package cli

import "github.com/spf13/cobra"

func newMyCommand() *cobra.Command {
    return &cobra.Command{
        Use:   "mycommand",
        Short: "Does something",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Implementation
            return nil
        },
    }
}
```

2. Register in root command
3. Add tests
4. Add documentation

### Add to Plugin SDK

1. Add interface/types to `pkg/plugin/sdk/v2/`
2. Implement in adapter layer
3. Add tests
4. Update SDK documentation

### Fix a Bug

1. Write a failing test that reproduces the bug
2. Fix the code
3. Verify test passes
4. Check for regressions

## Tips for Success

### 1. Start Small

Begin with documentation fixes or small bugs to learn the codebase.

### 2. Ask Questions

- Open an issue to discuss large changes
- Ask in PR comments if unsure

### 3. Keep PRs Focused

- One feature/fix per PR
- Makes review easier

### 4. Write Good Tests

- Test edge cases
- Test error paths
- Use table-driven tests

### 5. Document Changes

- Update relevant documentation
- Add code comments for complex logic

## Getting Help

- **Issues**: https://github.com/glide-cli/glide/issues
- **Discussions**: https://github.com/glide-cli/glide/discussions
- **Documentation**: https://github.com/glide-cli/glide/docs

## Summary

In this tutorial, you learned:
- How to set up the development environment
- The codebase structure
- How to make and test changes
- How to submit pull requests
- Best practices for contributing

Thank you for contributing to Glide!

## What's Next?

- Browse [open issues](https://github.com/glide-cli/glide/issues)
- Read the [Architecture Overview](../architecture/README.md)
- Check [ADR Index](../adr/README.md) for design decisions
