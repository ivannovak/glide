# Contributing to Glide

Thank you for your interest in contributing to Glide! This guide will help you get started with contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Contributions](#making-contributions)
- [Code Style](#code-style)
- [Testing](#testing)
- [Documentation](#documentation)
- [Plugin Development](#plugin-development)
- [Submitting Changes](#submitting-changes)
- [Review Process](#review-process)

## Code of Conduct

### Our Pledge

We are committed to providing a friendly, safe, and welcoming environment for all contributors, regardless of experience level, gender identity and expression, sexual orientation, disability, personal appearance, body size, race, ethnicity, age, religion, nationality, or any other characteristic.

### Expected Behavior

- Be respectful and inclusive
- Welcome newcomers and help them get started
- Focus on constructive criticism
- Accept feedback gracefully
- Prioritize the community's best interests

### Unacceptable Behavior

- Harassment, discrimination, or offensive comments
- Personal attacks or trolling
- Publishing others' private information
- Any conduct that could be considered inappropriate in a professional setting

## Getting Started

### Prerequisites

Before contributing, ensure you have:

1. **Go 1.21+** installed
2. **Git** for version control
3. **Docker** for testing (optional but recommended)
4. **Make** for build automation
5. A GitHub account

### Understanding Glide

Familiarize yourself with:

1. [Product Specification](product-spec.md) - What Glide does
2. [Architecture Documentation](architecture.md) - How Glide works
3. [Command Reference](command-reference.md) - Available commands
4. [Plugin Architecture](runtime-plugin-architecture.md) - Plugin system

### Finding Issues to Work On

Look for issues labeled:

- `good first issue` - Perfect for newcomers
- `help wanted` - Community help needed
- `documentation` - Documentation improvements
- `bug` - Bug fixes
- `enhancement` - New features

## Development Setup

### 1. Fork and Clone

```bash
# Fork the repository on GitHub, then:
git clone https://github.com/YOUR_USERNAME/glide.git
cd glide

# Add upstream remote
git remote add upstream https://github.com/ivannovak/glide.git
```

### 2. Install Dependencies

```bash
# Install Go dependencies
go mod download

# Install development tools
make install-tools

# Verify setup
make test
```

### 3. Build Glide

```bash
# Build binary
make build

# Install locally
make install

# Verify installation
glide version
```

### 4. Configure Development Environment

```bash
# Copy example configuration
cp .env.example .env

# Set up pre-commit hooks
make setup-hooks
```

### 5. Building with Custom Branding (Optional)

If you're contributing to a branded version (e.g., ACME):

```bash
# Build with specific branding
make build BRAND=acme

# Or use build tags directly
go build -tags brand_acme -o acme cmd/glid/main.go

# Test branded version
./acme version
```

## Making Contributions

### Workflow

1. **Create a branch:**
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/issue-description
   ```

2. **Make your changes:**
   - Write clean, readable code
   - Follow existing patterns
   - Add tests for new functionality
   - Update documentation

3. **Test your changes:**
   ```bash
   make test          # Run all tests
   make test-unit     # Unit tests only
   make test-integration  # Integration tests
   make lint          # Check code style
   ```

4. **Commit your changes:**
   ```bash
   git add .
   git commit -m "feat: add new feature"
   # See commit message guidelines below
   ```

### Commit Message Guidelines

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code changes that neither fix bugs nor add features
- `perf`: Performance improvements
- `test`: Test additions or corrections
- `build`: Build system or dependency changes
- `ci`: CI/CD configuration changes
- `chore`: Other changes that don't modify src or test files

**Examples:**
```bash
feat(docker): add support for Podman
fix(context): resolve worktree detection on macOS
docs(plugin): update plugin development guide
test(shell): add timeout handling tests
```

## Code Style

### Go Code Style

We follow standard Go conventions:

1. **Format code** with `gofmt`:
   ```bash
   make format
   ```

2. **Lint code** with `golangci-lint`:
   ```bash
   make lint
   ```

3. **Key principles:**
   - Keep functions small and focused
   - Use meaningful variable names
   - Add comments for exported functions
   - Handle errors explicitly
   - Use interfaces for flexibility

### Code Organization

```
glide/
├── cmd/glid/          # Main application entry
├── internal/          # Private application code
│   ├── cli/          # CLI command implementations
│   ├── context/      # Context detection
│   ├── docker/       # Docker integration
│   ├── shell/        # Shell execution
│   └── config/       # Configuration management
├── pkg/              # Public libraries
│   └── plugin/       # Plugin system
├── docs/             # Documentation
├── examples/         # Example code and plugins
└── tests/            # Test files
```

### Best Practices

1. **Error Handling:**
   ```go
   // Good
   if err != nil {
       return fmt.Errorf("failed to execute command: %w", err)
   }
   
   // Bad
   if err != nil {
       return err
   }
   ```

2. **Interfaces:**
   ```go
   // Define interfaces in the package that uses them
   type CommandExecutor interface {
       Execute(ctx context.Context, cmd string) error
   }
   ```

3. **Context Usage:**
   ```go
   // Always accept context as first parameter
   func DoSomething(ctx context.Context, param string) error {
       // Use context for cancellation
       select {
       case <-ctx.Done():
           return ctx.Err()
       default:
           // Do work
       }
   }
   ```

## Testing

### Writing Tests

1. **Unit Tests:**
   ```go
   func TestFunction(t *testing.T) {
       // Arrange
       input := "test"
       expected := "TEST"
       
       // Act
       result := strings.ToUpper(input)
       
       // Assert
       if result != expected {
           t.Errorf("got %s, want %s", result, expected)
       }
   }
   ```

2. **Table-Driven Tests:**
   ```go
   func TestMultipleCases(t *testing.T) {
       tests := []struct {
           name     string
           input    string
           expected string
       }{
           {"empty", "", ""},
           {"lowercase", "test", "TEST"},
       }
       
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               result := strings.ToUpper(tt.input)
               if result != tt.expected {
                   t.Errorf("got %s, want %s", result, tt.expected)
               }
           })
       }
   }
   ```

3. **Integration Tests:**
   - Test interactions between components
   - Use test fixtures for consistency
   - Clean up after tests

### Test Coverage

Aim for >80% test coverage:

```bash
# Generate coverage report
make coverage

# View HTML report
make coverage-html
open coverage.html
```

### Testing Plugins

Use the plugintest package:

```go
import "github.com/ivannovak/glide/pkg/plugin/plugintest"

func TestPlugin(t *testing.T) {
    harness := plugintest.NewTestHarness(t)
    // Test plugin functionality
}
```

## Documentation

### Types of Documentation

1. **Code Comments:**
   - Document all exported functions
   - Explain complex logic
   - Add examples for public APIs

2. **README Files:**
   - Keep README.md up-to-date
   - Include examples and quick start

3. **Architecture Docs:**
   - Document design decisions
   - Explain system components
   - Include diagrams when helpful

### Documentation Standards

1. **Clear and Concise:**
   - Use simple language
   - Avoid jargon
   - Provide examples

2. **Keep Updated:**
   - Update docs with code changes
   - Remove outdated information
   - Version documentation

3. **Format:**
   - Use Markdown for documentation
   - Include code examples
   - Add table of contents for long documents

## Branding Customization

### Creating a New Brand

If you're creating a branded version for your organization:

1. **Create brand definition:**
   ```go
   // internal/branding/brands/yourbrand.go
   //go:build brand_yourbrand
   
   package brands
   
   func init() {
       Current = Brand{
           Name:        "yourbrand",
           DisplayName: "Your Brand CLI",
           Description: "Your organization's development toolkit",
           CompanyName: "Your Company",
           Website:     "https://yourcompany.com",
           ConfigFile:  ".yourbrand.yml",
           EnvPrefix:   "YOURBRAND",
       }
   }
   ```

2. **Update build configuration:**
   - Add build tag to Makefile
   - Update GitHub Actions if needed
   - Document brand-specific features

3. **Test branded build:**
   ```bash
   make build BRAND=yourbrand
   ./yourbrand version
   ./yourbrand help
   ```

### Branding Guidelines

- Keep branding consistent across all text
- Update documentation for branded versions
- Test all commands with new branding
- Ensure plugins work with branded CLI
- Maintain separate config namespaces

## Plugin Development

### Creating a Plugin

1. **Use the boilerplate:**
   ```bash
   cp -r examples/plugin-boilerplate my-plugin
   cd my-plugin
   ```

2. **Implement the interface:**
   ```go
   type MyPlugin struct {
       sdk.UnimplementedGlidePluginServer
   }
   
   func (p *MyPlugin) GetMetadata(...) {...}
   func (p *MyPlugin) ListCommands(...) {...}
   func (p *MyPlugin) ExecuteCommand(...) {...}
   ```

3. **Test your plugin:**
   ```bash
   make test
   make build
   make install
   glide plugins list
   ```

### Plugin Guidelines

- Follow semantic versioning
- Document all commands
- Handle errors gracefully
- Minimize dependencies
- Provide clear error messages

## Submitting Changes

### Before Submitting

1. **Update from upstream:**
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run all checks:**
   ```bash
   make check  # Runs tests, lint, format
   ```

3. **Update documentation:**
   - Add/update relevant docs
   - Update CHANGELOG.md
   - Add examples if applicable

### Creating a Pull Request

1. **Push your branch:**
   ```bash
   git push origin feature/your-feature
   ```

2. **Open PR on GitHub:**
   - Use a clear, descriptive title
   - Reference any related issues
   - Describe what changes were made and why
   - Include screenshots for UI changes
   - List any breaking changes

3. **PR Template:**
   ```markdown
   ## Description
   Brief description of changes
   
   ## Type of Change
   - [ ] Bug fix
   - [ ] New feature
   - [ ] Breaking change
   - [ ] Documentation update
   
   ## Testing
   - [ ] Unit tests pass
   - [ ] Integration tests pass
   - [ ] Manual testing completed
   
   ## Checklist
   - [ ] Code follows style guidelines
   - [ ] Self-review completed
   - [ ] Documentation updated
   - [ ] Tests added/updated
   - [ ] Changelog updated
   ```

## Review Process

### What to Expect

1. **Automated Checks:**
   - CI/CD runs tests
   - Code coverage verification
   - Linting and formatting checks

2. **Code Review:**
   - Maintainers review code
   - Feedback provided
   - Suggestions for improvements

3. **Iterations:**
   - Address feedback
   - Update PR as needed
   - Re-review process

### Review Timeline

- Initial review: 2-3 business days
- Follow-up reviews: 1-2 business days
- Merge decision: After approval from maintainer

### After Merge

- Delete your feature branch
- Pull latest changes
- Celebrate your contribution! 🎉

## Getting Help

### Resources

- **Documentation:** Read the docs in `/docs`
- **Examples:** Check `/examples` for code samples
- **Issues:** Search existing issues for solutions
- **Discussions:** Join GitHub Discussions

### Communication Channels

- **GitHub Issues:** Bug reports and feature requests
- **GitHub Discussions:** Questions and ideas
- **Pull Requests:** Code contributions

### Asking Questions

When asking for help:

1. Search existing issues first
2. Provide context and background
3. Include error messages and logs
4. Describe what you've tried
5. Be patient and respectful

## Recognition

Contributors are recognized in:

- CONTRIBUTORS.md file
- Release notes
- GitHub contributor graph

## License

By contributing, you agree that your contributions will be licensed under the same license as the project (MIT License).

## Thank You!

Your contributions make Glide better for everyone. We appreciate your time and effort in improving the project!