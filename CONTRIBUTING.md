# Contributing to Glide

## Development Setup

### Prerequisites
- Go 1.24 or higher
- Make
- Git

### Getting Started
```bash
# Clone the repository
git clone https://github.com/ivannovak/glide.git
cd glide

# Install dependencies
go mod download

# Run tests
make test

# Build the binary
make build
```

## Development Workflow

### Before Committing

Always run the pre-commit checks to ensure your code matches CI requirements:

```bash
make pre-commit
```

This will:
1. Fix any formatting issues automatically
2. Run all linters
3. Run all tests

### Available Make Commands

| Command | Description |
|---------|-------------|
| `make help` | Show all available commands |
| `make build` | Build the glid binary |
| `make install` | Install glid to $GOPATH/bin |
| `make test` | Run all tests with coverage |
| `make lint` | Run all linters (matches CI exactly) |
| `make lint-fix` | Automatically fix formatting issues |
| `make ci` | Simulate CI checks locally |
| `make pre-commit` | Run all checks before committing |

### Linting

The project uses several linters to maintain code quality:

#### Code Formatting (`gofmt -s`)
The CI pipeline enforces code formatting with the `-s` flag for simplification:
```bash
# Check formatting
make lint-fmt

# Fix formatting automatically
make lint-fix
```

#### Go Vet
Checks for suspicious constructs:
```bash
make lint-vet
```

#### Running All Linters
To run the exact same checks as CI:
```bash
make lint
```

### Testing

#### Run All Tests
```bash
make test
```

#### Run Unit Tests Only
```bash
make test-unit
```

#### Run Integration Tests
```bash
make test-integration
```

### CI Pipeline

The CI pipeline runs on every push and pull request. It includes:
1. Module validation
2. Code formatting check (`gofmt -s -l .`)
3. Go vet
4. Unit tests with coverage
5. Integration tests
6. Multi-platform build verification
7. Security scanning with gosec

To simulate the CI pipeline locally:
```bash
make ci
```

## Code Style

### Go Formatting
- Use `gofmt -s` for all Go files (the `-s` flag simplifies code)
- Run `make lint-fix` to automatically format your code

### Imports
- Group imports in the following order:
  1. Standard library
  2. Third-party packages
  3. Internal packages
- Use goimports or manually organize

### Comments
- Export all public functions, types, and constants
- Use complete sentences starting with the name of the element

## Commit Messages

Follow conventional commits format:
```
<type>(<scope>): <subject>

<body>

<footer>
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Test changes
- `chore`: Build process or auxiliary tool changes

## Pull Request Process

1. Create a feature branch from `main`
2. Make your changes
3. Run `make pre-commit` to ensure all checks pass
4. Commit your changes with a descriptive message
5. Push to your fork and create a pull request
6. Ensure all CI checks pass
7. Request review from maintainers

## Troubleshooting

### Formatting Issues
If CI fails with formatting errors:
```bash
# See which files need formatting
make lint-fmt

# Fix all formatting issues
make lint-fix
```

### Go Version Issues
The project requires Go 1.24. If you have issues with staticcheck or other tools:
- The CI uses the latest versions of tools
- Some tools may lag behind the latest Go version
- Core checks (`gofmt -s` and `go vet`) are most important

### Quick CI Check
Before pushing, always run:
```bash
make ci
```

This runs the essential CI checks locally and will catch most issues before they reach GitHub.