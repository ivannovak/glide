# Branding Customization - Technical Specification

## Architecture Overview

The branding system uses Go build tags to select brand configurations at compile time, resulting in zero runtime overhead and complete brand isolation.

```
internal/branding/
├── brand.go          # Brand structure and interface
├── manager.go        # Brand management
└── brands/           # Brand definitions
    ├── glide.go      # Default brand (no build tag)
    └── acme.go       # //go:build brand_acme
```

## Technical Implementation

### 1. Brand Structure

```go
type Brand struct {
    // Core Identity
    Name         string  // Binary name (e.g., "acme")
    DisplayName  string  // Display name (e.g., "ACME CLI")
    Description  string  // Tool description
    CompanyName  string  // Organization name
    Website      string  // Organization website
    
    // Configuration
    ConfigFile   string  // Config filename (e.g., ".acme.yml")
    EnvPrefix    string  // Environment prefix (e.g., "ACME")
    PluginPrefix string  // Plugin prefix (e.g., "acme-plugin")
    
    // Customization
    PrimaryColor string  // Terminal color for branding
    LogoASCII    string  // ASCII art logo for splash
    
    // Features
    DefaultCommands []string  // Brand-specific default commands
    HiddenCommands  []string  // Commands to hide in this brand
}
```

### 2. Build-Time Selection

**Build Tags**:
```go
// internal/branding/brands/acme.go
//go:build brand_acme

package brands

func init() {
    Current = Brand{
        Name:         "acme",
        DisplayName:  "ACME Developer CLI",
        Description:  "ACME's official development toolkit",
        CompanyName:  "ACME Corporation",
        Website:      "https://acme.example.com",
        ConfigFile:   ".acme.yml",
        EnvPrefix:    "ACME",
        PluginPrefix: "acme-plugin",
    }
}
```

**Default Brand**:
```go
// internal/branding/brands/glide.go
//go:build !brand_acme

package brands

func init() {
    Current = Brand{
        Name:        "glide",
        DisplayName: "Glide CLI",
        // ... default configuration
    }
}
```

### 3. Brand Integration Points

**Command Registration**:
```go
func buildRootCommand() *cobra.Command {
    brand := branding.Current()
    
    cmd := &cobra.Command{
        Use:   brand.Name,
        Short: brand.Description,
        Long:  fmt.Sprintf("%s - %s", brand.DisplayName, brand.Description),
    }
    
    // Add brand-specific commands
    for _, cmdName := range brand.DefaultCommands {
        cmd.AddCommand(createBrandCommand(cmdName))
    }
    
    return cmd
}
```

**Configuration Loading**:
```go
func loadConfig() (*Config, error) {
    brand := branding.Current()
    
    // Look for brand-specific config file
    configPath := filepath.Join(homeDir, brand.ConfigFile)
    
    // Use brand-specific environment variables
    envPrefix := brand.EnvPrefix
    
    return loadFromFile(configPath, envPrefix)
}
```

**Plugin Discovery**:
```go
func discoverPlugins() ([]*Plugin, error) {
    brand := branding.Current()
    
    // Look for brand-specific plugins
    pattern := fmt.Sprintf("%s-*", brand.PluginPrefix)
    
    return scanDirectory(pluginDir, pattern)
}
```

### 4. Build Process

**Makefile Integration**:
```makefile
BRAND ?= glide
BRAND_TAG := $(if $(filter-out glide,$(BRAND)),brand_$(BRAND),)

build:
    go build \
        $(if $(BRAND_TAG),-tags $(BRAND_TAG)) \
        -o $(BRAND) \
        cmd/glid/main.go
```

**Build Commands**:
```bash
# Build default Glide
make build

# Build branded version
make build BRAND=acme

# Direct Go build
go build -tags brand_acme -o acme cmd/glid/main.go
```

### 5. Configuration Isolation

**File Paths**:
```go
func getConfigPaths(brand Brand) ConfigPaths {
    return ConfigPaths{
        GlobalConfig: fmt.Sprintf("~/%s", brand.ConfigFile),
        ProjectConfig: fmt.Sprintf(".%s.yml", brand.Name),
        CacheDir: fmt.Sprintf("~/.cache/%s", brand.Name),
        DataDir: fmt.Sprintf("~/.local/share/%s", brand.Name),
        PluginDir: fmt.Sprintf("~/.%s/plugins", brand.Name),
    }
}
```

**Environment Variables**:
```go
func getEnvVar(key string) string {
    brand := branding.Current()
    envKey := fmt.Sprintf("%s_%s", brand.EnvPrefix, key)
    return os.Getenv(envKey)
}
```

### 6. User Interface Customization

**Help Text**:
```go
func formatHelp(cmd *cobra.Command) string {
    brand := branding.Current()
    
    help := fmt.Sprintf(`%s

%s
%s

Usage:
  %s [command]

`, brand.LogoASCII, brand.DisplayName, brand.Description, brand.Name)
    
    return help
}
```

**Error Messages**:
```go
func formatError(err error) string {
    brand := branding.Current()
    
    return fmt.Sprintf(`%s Error: %v

For help, visit: %s/docs
Or run: %s help
`, brand.DisplayName, err, brand.Website, brand.Name)
}
```

## Implementation Strategy

### 1. Zero Runtime Overhead

All branding decisions made at compile time:
- No runtime brand detection
- No configuration lookups
- No performance impact
- Smaller binary size

### 2. Complete Isolation

Different brands never interact:
- Separate configuration files
- Distinct environment variables
- Isolated plugin directories
- Independent cache locations

### 3. Single Codebase

All brands built from same source:
- Shared core functionality
- Common bug fixes
- Unified testing
- Consistent behavior

## Testing Strategy

### 1. Brand-Specific Tests

```go
//go:build brand_acme

func TestAcmeBranding(t *testing.T) {
    brand := branding.Current()
    assert.Equal(t, "acme", brand.Name)
    assert.Equal(t, "ACME", brand.EnvPrefix)
}
```

### 2. Multi-Brand CI/CD

```yaml
# .github/workflows/build.yml
strategy:
  matrix:
    brand: [glide, acme]

steps:
  - run: make build BRAND=${{ matrix.brand }}
  - run: make test BRAND=${{ matrix.brand }}
```

### 3. Brand Compatibility

```go
func TestBrandCompatibility(t *testing.T) {
    brands := []string{"glide", "acme"}
    
    for _, brand := range brands {
        // Test core functionality works with each brand
        testCoreCommands(t, brand)
        testConfiguration(t, brand)
        testPluginSystem(t, brand)
    }
}
```

## Distribution

### 1. Binary Naming

```bash
# Output binaries match brand name
glide         # Default brand
acme          # ACME brand
```

### 2. Release Automation

```bash
# Build all brands for release
for brand in glide acme; do
    GOOS=darwin GOARCH=amd64 make build BRAND=$brand
    GOOS=linux GOARCH=amd64 make build BRAND=$brand
done
```

### 3. Package Management

```bash
# Homebrew formula for branded version
class Acme < Formula
  desc "ACME Developer CLI"
  homepage "https://acme.example.com"
  url "https://github.com/acme/cli/releases/download/v1.0.0/acme-darwin-amd64"
  
  def install
    bin.install "acme"
  end
end
```

## Security Considerations

### 1. Brand Verification

```go
// Embed brand signature at build time
var (
    BrandSignature = "ACME-OFFICIAL-BUILD"
    BuildTime      = "2025-01-01T00:00:00Z"
    BuildCommit    = "abc123"
)

func VerifyBrand() bool {
    return BrandSignature == expectedSignature()
}
```

### 2. Configuration Protection

```go
// Ensure brand isolation
func validateConfigPath(path string) error {
    brand := branding.Current()
    expected := brand.ConfigFile
    
    if !strings.HasSuffix(path, expected) {
        return ErrInvalidConfigPath
    }
    
    return nil
}
```

### 3. Plugin Compatibility

```go
// Ensure plugins match brand
func validatePlugin(plugin *Plugin) error {
    brand := branding.Current()
    
    if !strings.HasPrefix(plugin.Name, brand.PluginPrefix) {
        return ErrIncompatiblePlugin
    }
    
    return nil
}
```

## Performance Analysis

### Build Time
- Single brand: ~10 seconds
- All brands: ~30 seconds
- Parallel builds: ~15 seconds

### Binary Size
- Base size: ~15MB
- Brand overhead: ~0KB (compile-time)
- Total size: ~15MB per brand

### Runtime Performance
- Brand lookup: 0ms (compile-time)
- Config resolution: <1ms
- Plugin discovery: <5ms

## Migration Path

### From Generic to Branded

1. Install branded version:
```bash
# Install ACME CLI
curl -L https://acme.example.com/install | bash
```

2. Migrate configuration:
```bash
# Copy existing config
cp ~/.glide.yml ~/.acme.yml

# Update environment variables
export ACME_PROJECT=$GLIDE_PROJECT
```

3. Update scripts:
```bash
# Replace glide with acme
sed -i 's/glide/acme/g' scripts/*.sh
```

### Coexistence

Both generic and branded versions can coexist:
```bash
glide version    # Generic Glide
acme version     # ACME branded

# Different configs
ls ~/.*.yml
.glide.yml
.acme.yml
```

## Future Enhancements

### 1. Brand Templates
```go
// Generate brand from template
type BrandTemplate struct {
    Industry  string
    Features  []string
    ColorScheme ColorScheme
}
```

### 2. Dynamic Features
```go
// Brand-specific feature flags
type BrandFeatures struct {
    EnableCloud     bool
    EnableAnalytics bool
    CustomCommands  []Command
}
```

### 3. Brand Plugins
```go
// Plugins specific to brand
type BrandPlugin interface {
    ValidateForBrand(brand Brand) error
    GetBrandCommands(brand Brand) []Command
}
```