# Initial MVP - Product Specification

## Executive Summary

Glide is a Go-based command-line tool designed to replace Makefile-based development workflows. It addresses the fundamental limitations of Make (particularly around argument passing) while providing a modern, intuitive interface for managing Docker environments, running tests, and executing development tasks. The CLI is context-aware, automatically detecting whether it's being run from the project root, main repository (vcs/), or a worktree, and adjusts its behavior accordingly.

## Problem Statement

### Current Pain Points
1. **Make's Argument Parsing Limitations**: Cannot pass CLI flags naturally (e.g., `--parallel`, `--processes=5`) without them being intercepted by Make itself
2. **Complex Workarounds**: Requires transformation scripts and variable assignments to work around Make's design
3. **Limited Composability**: Difficult to chain commands or build complex workflows
4. **Poor Error Handling**: Make's error messages are often cryptic and unhelpful
5. **No Native Progress Tracking**: No built-in way to show progress for long-running operations
6. **Cross-Platform Issues**: Make behavior varies between macOS and Linux

### User Needs
- Direct pass-through of arguments to underlying tools (especially test frameworks)
- Intelligent context awareness (detecting worktree location, finding compose files)
- Clear, helpful error messages with suggested fixes
- Consistent behavior across platforms
- Fast execution with minimal overhead
- Tab completion and command discovery

## Solution Overview

A standalone CLI application that:
- Installs as a zero-dependency binary - no runtime requirements
- Distributes via direct download or automated install script
- Uses a global configuration file at `~/.glide.yml` for project locations
- Provides direct command execution without argument interpretation issues
- Auto-detects project context (worktree, compose files, environment)
- Offers intelligent defaults with full override capability
- Maintains backward compatibility with existing workflow patterns
- Provides superior UX with progress indicators, colored output, and helpful suggestions

## Core Features

### 1. Context-Aware Execution
The CLI detects the current working directory and adjusts its behavior:

**From project root (e.g., `/Users/ivan/Code/acme/`):**
- `glide` commands operate on the main repository
- `glide global` commands available for multi-worktree operations

**From main repository (e.g., `/Users/ivan/Code/acme/vcs/`):**
- Commands operate on the main repository
- Compose files resolved to `docker-compose.yml` and `../docker-compose.override.yml`

**From worktree (e.g., `/Users/ivan/Code/acme/worktrees/feature-branch/`):**
- Commands operate on specific worktree
- Compose files resolved to `docker-compose.yml` and `../../docker-compose.override.yml`

### 2. Essential Commands

#### Testing
- `glide test [args]` - Run tests with full argument pass-through
- Auto-detects test framework (PHPUnit/Pest)
- Supports all test framework flags directly

#### Docker Management
- `glide up` - Start Docker environment
- `glide down` - Stop Docker environment
- `glide restart [service]` - Restart services
- `glide logs [service]` - View logs
- `glide exec <service> <command>` - Execute in container
- `glide ps` - Show container status

#### Global Operations (Multi-worktree)
- `glide global status` - Status across all worktrees
- `glide global list` - List active worktrees
- `glide global worktree <name>` - Create new worktree
- `glide global down` - Stop all Docker containers
- `glide global clean` - Clean all build artifacts

#### Development Workflow
- `glide lint` - Run code quality checks
- `glide format` - Format code
- `glide build` - Build project
- `glide db <command>` - Database operations

### 3. User Experience

#### Intelligent Defaults
- Automatic project detection
- Smart compose file resolution
- Framework detection for tests
- Sensible timeout values

#### Clear Feedback
- Colored output for readability
- Progress indicators for long operations
- Clear error messages with suggestions
- Success/failure indicators

#### Command Discovery
- `glide help` - Context-aware help
- Tab completion support
- Command suggestions on typos
- Interactive mode for complex operations

## Success Criteria

### Functional Requirements
- ✅ Complete argument pass-through to underlying commands
- ✅ Context detection (worktree, compose files, etc.)
- ✅ All essential commands implemented
- ✅ Cross-platform compatibility (macOS, Linux)
- ✅ Zero runtime dependencies

### Performance Requirements
- ✅ Startup time < 50ms
- ✅ Command overhead < 10ms
- ✅ Binary size < 20MB

### User Experience Requirements
- ✅ Install in < 30 seconds
- ✅ Clear, actionable error messages
- ✅ Tab completion available
- ✅ Context-aware help system
- ✅ No breaking changes to existing workflows

## Non-Goals

### Out of Scope for MVP
- GUI interface
- Windows native support (WSL2 is sufficient)
- Remote execution
- Build system replacement (just wrapping)
- Package management
- CI/CD pipeline integration

## User Stories

### Developer Daily Workflow
"As a developer, I want to run tests with specific flags without fighting Make's argument parsing, so I can iterate quickly on my code."

### DevOps Engineer
"As a DevOps engineer, I want to manage multiple worktree Docker environments efficiently, so I can test different configurations simultaneously."

### New Team Member
"As a new team member, I want clear command discovery and help, so I can be productive without memorizing complex Make targets."

## Metrics for Success

### Adoption Metrics
- 90% of development team using within 1 month
- 50% reduction in Make-related support questions
- 80% user satisfaction in surveys

### Performance Metrics
- Zero performance regressions vs Make
- 50% faster test execution due to better parallelization
- 75% reduction in Docker-related errors

### Quality Metrics
- Zero critical bugs in first month
- < 5 second mean time to resolution for issues
- 95% uptime for all commands

## Future Considerations

### Phase 2 Features
- Plugin system for extensibility
- Cloud integration for remote development
- Advanced workflow automation
- Team configuration sharing
- Metrics and analytics

### Long-term Vision
Glide becomes the standard CLI for modern development workflows, replacing Make for interactive development while maintaining compatibility with CI/CD pipelines that still use Make.