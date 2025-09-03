# ADR-006: Branding Customization System

## Status
Accepted

## Date
2025-09-03

## Context
Glide was initially developed as a generic development CLI, but it's being used as the foundation for organization-specific tools. Different organizations need to:
- Brand the CLI with their own name and identity
- Customize help text and descriptions
- Maintain their own configuration namespaces
- Distribute as their own tool
- Avoid confusion between different branded versions

The challenge is supporting multiple brands without:
- Maintaining separate codebases
- Runtime configuration complexity
- Performance overhead
- Breaking existing installations

## Decision
We will implement a build-time branding system using Go build tags:

1. **Build-Time Selection**: Brands are selected at compile time using build tags
2. **Zero Runtime Overhead**: No performance impact since branding is compiled in
3. **Complete Customization**: All user-facing text can be branded
4. **Separate Namespaces**: Each brand has its own config files and environment variables
5. **Plugin Compatibility**: Plugins work across all branded versions

Implementation approach:
```go
// Build with specific brand
go build -tags brand_acme -o acme

// Brand definition
type Brand struct {
    Name         string  // Binary name
    DisplayName  string  // Display name
    Description  string  // CLI description
    CompanyName  string  // Organization
    ConfigFile   string  // Config filename
    EnvPrefix    string  // Env var prefix
}
```

Branding affects:
- Binary name (glide vs acme)
- Configuration file (~/.glide.yml vs ~/.acme.yml)
- Environment variables (GLIDE_* vs ACME_*)
- Help text and descriptions
- Plugin naming (glide-plugin-* vs acme-plugin-*)

## Consequences

### Positive
- **Zero Runtime Cost**: No performance impact
- **Complete Isolation**: Different brands don't interfere
- **Easy Distribution**: Single binary with embedded branding
- **Maintainability**: Single codebase for all brands
- **Professional**: Organizations get fully-branded tools
- **Flexibility**: New brands added without code changes

### Negative
- **Build Complexity**: Requires different builds for each brand
- **Testing Overhead**: Need to test each brand variant
- **Binary Size**: Can't switch brands at runtime
- **Documentation**: Must maintain brand-specific docs

## Implementation
Branding system implemented in `internal/branding/`:
- Build tags select brand at compile time
- Brand definitions in `brands/` directory
- Manager provides current brand to application
- All user-facing text references brand variables

## Usage Examples

### Creating a New Brand
1. Create brand definition file:
```go
// internal/branding/brands/acme.go
//go:build brand_acme

package brands

func init() {
    Current = Brand{
        Name:        "acme",
        DisplayName: "ACME CLI",
        Description: "ACME's development toolkit",
        CompanyName: "ACME Corporation",
        Website:     "https://acme.com",
        ConfigFile:  ".acme.yml",
        EnvPrefix:   "ACME",
        PluginPrefix: "acme-plugin",
    }
}
```

2. Build with brand:
```bash
go build -tags brand_acme -o acme cmd/glid/main.go
```

3. Result is fully-branded CLI:
```bash
$ acme help
ACME CLI - ACME's development toolkit

$ ls ~/.acme.yml  # Uses branded config
```

## Alternatives Considered

1. **Runtime Configuration**: Rejected due to confusion potential and complexity
2. **Separate Repositories**: Rejected due to maintenance overhead
3. **Git Branches**: Rejected due to merge complexity
4. **Templating**: Rejected due to build complexity
5. **Environment Variables**: Rejected as insufficient for complete branding

## Migration Path
Existing Glide installations are unaffected. Organizations can:
1. Continue using generic Glide
2. Migrate to branded version with same features
3. Run both versions simultaneously (different namespaces)