## [2.1.2](https://github.com/ivannovak/glide/compare/v2.1.1...v2.1.2) (2025-11-24)


### Bug Fixes

* remove CI dependency from release workflow ([3b8b007](https://github.com/ivannovak/glide/commit/3b8b007ff4f6e68ba8aaf836a14ca641d88be0c1))

## [2.1.1](https://github.com/ivannovak/glide/compare/v2.1.0...v2.1.1) (2025-11-24)


### Bug Fixes

* add actions:write permission to trigger release workflow ([ba7cfda](https://github.com/ivannovak/glide/commit/ba7cfda98e7674614898e955245d6062bf806502))

## [2.1.0](https://github.com/ivannovak/glide/compare/v2.0.0...v2.1.0) (2025-11-24)


### Features

* auto-trigger release workflow after semantic-release ([1caec4e](https://github.com/ivannovak/glide/commit/1caec4ef1a731326629c92d6e4b36827887ead3b))

## [2.0.0](https://github.com/ivannovak/glide/compare/v1.3.0...v2.0.0) (2025-11-24)


### ⚠ BREAKING CHANGES

* Plugin installation now supports downloading from GitHub releases in addition to local files. This enables users to install plugins directly from GitHub without building from source.

Changes:
- Fix release workflow binary naming (glid-* -> glide-*)
- Add GitHub API integration for downloading release binaries
- Enhance `glide plugins install` to detect and download from github.com URLs
- Auto-detect platform (OS/arch) for binary downloads
- Add comprehensive help text with usage examples

Examples:
  glide plugins install github.com/ivannovak/glide-plugin-go
  glide plugins install ./path/to/local/binary

### Features

* add GitHub release binary downloads for plugins ([cb1d919](https://github.com/ivannovak/glide/commit/cb1d919f2a9fe97d4ca444af4df2a659840d9907))


### Bug Fixes

* add URL validation for GitHub downloads ([0648d9c](https://github.com/ivannovak/glide/commit/0648d9cf5330d1e2e0e01c19db524bf8b86c7c99))
* exclude G107 from gosec security scan ([b584549](https://github.com/ivannovak/glide/commit/b584549b78b8b7bc07a42b43c6d693dc501196df))

## [1.3.0](https://github.com/ivannovak/glide/compare/v1.2.0...v1.3.0) (2025-11-21)


### Features

* Extract Docker functionality to external plugin architecture ([#10](https://github.com/ivannovak/glide/issues/10)) ([e297fd9](https://github.com/ivannovak/glide/commit/e297fd974cd4f50c12f60d19051659f46cebdbc1))

## [1.2.0](https://github.com/ivannovak/glide/compare/v1.1.0...v1.2.0) (2025-11-20)


### Features

* improve help display with ASCII header and user command visibility ([1bf4ae7](https://github.com/ivannovak/glide/commit/1bf4ae7daf27b5822ab0ca77b7c714b0e2d0b140))


### Bug Fixes

* format help.go to pass lint checks ([30649f0](https://github.com/ivannovak/glide/commit/30649f074373c6837e6f1a4ba98d02974194b5c5))

## [1.1.0](https://github.com/ivannovak/glide/compare/v1.0.0...v1.1.0) (2025-11-20)


### Features

* Framework Detection Plugin System with Go, Node.js, and PHP support ([#9](https://github.com/ivannovak/glide/issues/9)) ([0ed3615](https://github.com/ivannovak/glide/commit/0ed361591357483c6eaab3d11de006392b23dd04))

## [1.0.0](https://github.com/ivannovak/glide/compare/v0.10.1...v1.0.0) (2025-11-19)


### ⚠ BREAKING CHANGES

* The CLI command has been renamed from "glid" to "glide".
Users will need to use "glide" instead of "glid" after this update.
* The 'global' command is now 'project' to better reflect its purpose

- Rename command from 'glid global' to 'glid project'
- Update alias from 'g' to 'p'
- Rename all GlobalCommand structs/types to ProjectCommand
- Update all documentation to use new terminology
- Update method CanUseGlobalCommands() to CanUseProjectCommands()

The term 'project' more accurately describes these commands that operate
across all worktrees within a project, avoiding confusion with system-wide
operations that 'global' might imply.

Migration guide:
- Replace 'glid global' with 'glid project' in scripts
- Replace 'glid g' with 'glid p' for the short alias

### Features

* add standalone mode and context-aware help system ([a0d72ca](https://github.com/ivannovak/glide/commit/a0d72ca580c3bb334a1108f22773defa9e3971c2))
* **commands:** add YAML-defined commands with recursive config discovery ([d22e4b9](https://github.com/ivannovak/glide/commit/d22e4b9f036f084e3f099d532ef854e487d8f12d))


### Bug Fixes

* **tests:** update plugin SDK test expectations after glide rename ([2c56b76](https://github.com/ivannovak/glide/commit/2c56b76df3f5c3324c2580fd77509df8f99b119f))


### Code Refactoring

* rename 'global' commands to 'project' commands ([3a65446](https://github.com/ivannovak/glide/commit/3a65446b5250727b0cb004823bdbefe64bd73ddf))
* rename command from glid to glide ([767bd7f](https://github.com/ivannovak/glide/commit/767bd7f2b0377fd5d4e1e58b6aec61cf0b6d0068))

## [0.10.1](https://github.com/ivannovak/glide/compare/v0.10.0...v0.10.1) (2025-09-11)


### Bug Fixes

* **ci:** prevent duplicate CI runs and ensure tests before release ([#8](https://github.com/ivannovak/glide/issues/8)) ([ac112c6](https://github.com/ivannovak/glide/commit/ac112c6ff51c977b99bf7ddbe2ce0e99abd006e2))

## [0.10.0](https://github.com/ivannovak/glide/compare/v0.9.0...v0.10.0) (2025-09-11)


### Features

* **sdk:** Add BasePlugin helper for simplified plugin authorship ([#7](https://github.com/ivannovak/glide/issues/7)) ([e9adb0b](https://github.com/ivannovak/glide/commit/e9adb0b59dbc3cdcc1197ef1ee093f0f2316e7cc))

## [0.9.0](https://github.com/ivannovak/glide/compare/v0.8.1...v0.9.0) (2025-09-10)


### Features

* **plugin:** Implement interactive command support with bidirectional streaming ([#6](https://github.com/ivannovak/glide/issues/6)) ([c95d060](https://github.com/ivannovak/glide/commit/c95d060f3c8d3167635e4c52d46c10d67c110b81))

## [0.8.1](https://github.com/ivannovak/glide/compare/v0.8.0...v0.8.1) (2025-09-10)


### Bug Fixes

* **plugin:** use branding configuration for plugin discovery ([#5](https://github.com/ivannovak/glide/issues/5)) ([8b9d2f5](https://github.com/ivannovak/glide/commit/8b9d2f55e94bfdd74eeded45e54f26610c3c1ae2))

## [0.8.0](https://github.com/ivannovak/glide/compare/v0.7.1...v0.8.0) (2025-09-10)


### Features

* major architectural improvements - registry consolidation and shell builder extraction ([#1](https://github.com/ivannovak/glide/issues/1)) ([b60b15e](https://github.com/ivannovak/glide/commit/b60b15e4467b9afe70a149e0fd5b37905ebe749b))
* **release:** integrate semantic-release for automated versioning ([#2](https://github.com/ivannovak/glide/issues/2)) ([24771e9](https://github.com/ivannovak/glide/commit/24771e9ddc8dba8260075ba6e71d6aa71613858f))


### Bug Fixes

* **ci:** correct repository references in configuration files ([#4](https://github.com/ivannovak/glide/issues/4)) ([06825e1](https://github.com/ivannovak/glide/commit/06825e19472299b7a932780fe02c874640f49878))
* **ci:** remove repository condition from semantic-release workflow ([#3](https://github.com/ivannovak/glide/issues/3)) ([ba8b757](https://github.com/ivannovak/glide/commit/ba8b757e73e259d6d3eb5f138715636ae2eeb3fe))
