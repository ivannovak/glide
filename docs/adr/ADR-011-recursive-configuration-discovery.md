# ADR 011: Recursive Configuration Discovery

## Status
Accepted

## Context
Glide needs to support configuration files at multiple levels of a project hierarchy. Users may want to define:
- Project-wide settings at the repository root
- Team or department-specific settings in parent directories
- Directory-specific overrides for specialized workflows
- Personal configurations in their working directory

The plugin system already implements recursive discovery, walking up the directory tree to find plugin directories. This pattern has proven effective for allowing flexible, hierarchical organization of functionality.

## Decision
We will implement recursive configuration discovery that walks up the directory tree from the current working directory, looking for configuration files (`.glide.yml` by default, configurable via branding).

The discovery mechanism will:
1. Start from the current working directory
2. Walk up the directory tree looking for configuration files
3. Stop at the project root (detected by `.git`), home directory, or filesystem root
4. Load and merge all discovered configurations with proper precedence

### Priority Order (Highest to Lowest)
1. Current directory (`./.glide.yml`)
2. Parent directories (walking up the tree)
3. Project root (`[project-root]/.glide.yml`)
4. Global user configuration (`~/.glide/config.yml`)

### Configuration Merging
When multiple configuration files are found:
- Commands are merged with higher-priority files overriding lower-priority ones
- Settings are merged with first non-empty values taking precedence
- Lists and maps are replaced entirely (not merged) by higher-priority configs

## Consequences

### Positive
- **Consistency**: Aligns with the existing plugin discovery mechanism
- **Flexibility**: Allows configuration at any level of the directory tree
- **Team Collaboration**: Teams can share configurations at the repository level
- **Personal Customization**: Developers can override settings locally without affecting others
- **Gradual Adoption**: Projects can start with a single config and add more as needed
- **Principle of Least Surprise**: Behaves similarly to other tools like `.gitignore` or `.eslintrc`

### Negative
- **Complexity**: Multiple configuration sources can be confusing to debug
- **Performance**: Requires filesystem traversal on every invocation
- **Precedence Confusion**: Users might not understand which config is taking effect
- **Security**: Configurations from parent directories could affect subdirectories unexpectedly

### Mitigation Strategies
- Cache discovered configurations during a session
- Provide a `glideconfig --show-sources` command to display all loaded configs
- Document the precedence order clearly
- Stop at project boundaries (`.git`) to prevent external configs from affecting projects

## Implementation Details

### Discovery Algorithm
```go
func DiscoverConfigs(startDir string) ([]string, error) {
    var configs []string
    home, _ := os.UserHomeDir()
    current := startDir

    for {
        // Check for config file
        configPath := filepath.Join(current, branding.ConfigFileName)
        if _, err := os.Stat(configPath); err == nil {
            configs = append(configs, configPath)
        }

        // Stop at project root
        if _, err := os.Stat(filepath.Join(current, ".git")); err == nil {
            break
        }

        // Stop at boundaries
        if current == "/" || current == home {
            break
        }

        current = filepath.Dir(current)
    }

    // Reverse for correct precedence
    reverse(configs)
    return configs, nil
}
```

### Configuration File Naming
The configuration filename is determined by the branding package:
- Default: `.glide.yml`
- Configurable at build time via ldflags
- Ensures consistency across the tool

## Examples

### Directory Structure
```
~/projects/myapp/
├── .git/
├── .glide.yml                 # Project-wide commands
├── backend/
│   ├── .glide.yml             # Backend-specific commands
│   └── services/
│       └── .glide.yml         # Service-specific overrides
└── frontend/
    └── .glide.yml             # Frontend-specific commands
```

### Command Resolution
When running `glidebuild` from `~/projects/myapp/backend/services/`:
1. Check `./services/.glide.yml` - highest priority
2. Check `./backend/.glide.yml`
3. Check `./.glide.yml` (project root) - stops here due to `.git`
4. Check `~/.glide/config.yml` - lowest priority

## Alternatives Considered

### Single Configuration File
Only load from project root or user home.
- **Pros**: Simple, predictable
- **Cons**: No flexibility for directory-specific workflows

### Explicit Include Directives
Require configs to explicitly include parent configs.
- **Pros**: Explicit control over inheritance
- **Cons**: More complex, requires manual maintenance

### Environment Variable for Config Paths
Use `GLIDE_CONFIG_PATH` with colon-separated paths.
- **Pros**: Explicit, no filesystem traversal
- **Cons**: Requires manual setup, less discoverable

## References
- Plugin Discovery Implementation: `pkg/plugin/sdk/manager.go`
- Configuration Types: `internal/config/types.go`
- Discovery Implementation: `internal/config/discovery.go`
- Similar Tools: git (`.gitignore`), ESLint (`.eslintrc`), npm (`.npmrc`)