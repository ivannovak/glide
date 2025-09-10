# ADR-010: Semantic Release Automation

## Status
Proposed

## Context
Currently, the Glide CLI release process requires manual version management and tag creation. The workflow triggers on manually created tags or through workflow_dispatch with explicit version input. This approach has several limitations:

1. **Manual Version Management**: Developers must manually determine version numbers based on changes
2. **Inconsistent Release Notes**: Release notes are generated from git commits but lack standardized formatting
3. **Error-Prone Process**: Manual tagging can lead to version conflicts or forgotten releases
4. **No Automated Changelog**: No standardized CHANGELOG.md maintained across releases
5. **Lack of Semantic Versioning Enforcement**: No automatic enforcement of semver based on change types

The team follows conventional commits, providing a foundation for automated semantic versioning.

## Decision
We will integrate [semantic-release](https://semantic-release.gitbook.io/) to automate the version management and release process. This tool will:

1. Analyze commit messages using conventional commit format
2. Automatically determine the next semantic version (major, minor, patch)
3. Update version in source code (`pkg/version/version.go`)
4. Generate and maintain a CHANGELOG.md file
5. Create git tags and GitHub releases automatically
6. Trigger existing build workflows through tag creation

## Implementation Details

### Workflow Architecture
```
Push to main → Semantic Release Workflow → Analyze Commits → Determine Version
                                          ↓
                    Create Tag & Release ← Update Version & Changelog
                                          ↓
                    Trigger Release.yml → Build Binaries → Publish Artifacts
```

### Configuration Components
1. **`.releaserc.json`**: Semantic-release configuration
   - Commit analyzer for version determination
   - Changelog generation
   - Version file updates via replace plugin
   - Git commits for changes
   - GitHub release creation

2. **`.github/workflows/semantic-release.yml`**: Automation workflow
   - Triggers on push to main
   - Runs semantic-release with Node.js
   - Uses GitHub token for authentication

3. **`package.json`**: Node.js dependencies
   - semantic-release and plugins
   - Locked versions via package-lock.json

### Version Bump Rules
- `fix:` commits → Patch release (1.0.0 → 1.0.1)
- `feat:` commits → Minor release (1.0.0 → 1.1.0)  
- `BREAKING CHANGE:` or `feat!:` → Major release (1.0.0 → 2.0.0)
- Other types (`docs:`, `style:`, `refactor:`, `test:`, `chore:`) → No release

## Consequences

### Positive
- **Automated Versioning**: Removes manual version decision-making
- **Consistent Releases**: Every merge to main evaluates release necessity
- **Standardized Changelog**: Automatically maintained CHANGELOG.md with categorized changes
- **Reduced Human Error**: No forgotten releases or version conflicts
- **Clear Release Triggers**: Commit messages directly control versioning
- **Improved Traceability**: Direct link between commits and releases
- **Faster Release Cycle**: No manual intervention required for releases

### Negative
- **Additional Dependency**: Requires Node.js/npm in CI environment
- **Commit Message Discipline**: Requires strict adherence to conventional commits
- **Learning Curve**: Team must understand conventional commit impact on versioning
- **CI Complexity**: Additional workflow and configuration to maintain
- **Recovery Complexity**: Fixing incorrect releases requires manual intervention

### Neutral
- **Existing Workflow Preservation**: Current release.yml continues to handle builds
- **Token Requirements**: Needs GitHub token with appropriate permissions
- **Dry-Run Capability**: Can test locally without creating releases

## Alternatives Considered

1. **GoReleaser with Manual Triggers**
   - Pro: Go-native tooling
   - Con: Still requires manual version determination

2. **Custom Go Script**
   - Pro: No external dependencies
   - Con: Significant development effort, reinventing the wheel

3. **GitHub Release Drafter**
   - Pro: Native GitHub integration
   - Con: Only handles release notes, not versioning

4. **Keep Current Manual Process**
   - Pro: Full control, no new dependencies
   - Con: Continues current pain points

## Migration Path

1. Merge semantic-release configuration to main
2. Ensure all developers understand conventional commits
3. First release will analyze all commits since last manual tag
4. Subsequent releases will be fully automated
5. Manual releases remain possible via workflow_dispatch

## References
- [Semantic Release Documentation](https://semantic-release.gitbook.io/)
- [Conventional Commits Specification](https://www.conventionalcommits.org/)
- [Semantic Versioning Specification](https://semver.org/)
- [ADR-005: Testing Strategy](./ADR-005-testing-strategy.md) - Commit standards align with testing practices