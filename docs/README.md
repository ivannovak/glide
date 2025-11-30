# Glide Documentation

Welcome to the Glide CLI documentation. This directory contains user-facing documentation, guides, and architectural decision records.

## 📚 Documentation Structure

### Getting Started
- **[Getting Started](getting-started/)** - Installation and first steps
- **[Core Concepts](core-concepts/)** - Understanding how Glide works
- **[Tutorials](tutorials/)** - Progressive learning path from beginner to advanced

### User Documentation
Essential guides for users and contributors:

- **[Command Reference](command-reference.md)** - Complete list of all Glide commands and their usage
- **[Framework Detection](framework-detection.md)** - Automatic framework and language detection system
- **[Plugin Development](plugin-development.md)** - Guide for creating Glide plugins (SDK v2)
- **[Troubleshooting Guide](troubleshooting.md)** - Common issues and their solutions
- **[Contributing Guide](CONTRIBUTING.md)** - How to contribute to the Glide project

### Developer Guides
- **[Architecture Overview](architecture/)** - System architecture and design
- **[Error Handling](guides/error-handling.md)** - Best practices for error handling
- **[Performance](guides/performance.md)** - Performance optimization guide
- **[SDK v2 Migration](guides/PLUGIN-SDK-V2-MIGRATION.md)** - Migrate legacy v1 plugins to v2 (v1 deprecated)

### Architecture Decision Records (ADRs)
Formal documentation of architectural decisions (17 total):

- **[ADR Index](adr/README.md)** - Complete list of all architectural decisions
- **[Context-Aware Architecture](adr/ADR-001-context-aware-architecture.md)** - Core architectural principles
- **[Plugin System Design](adr/ADR-002-plugin-system-design.md)** - Plugin architecture decisions
- **[Configuration Management](adr/ADR-003-configuration-management.md)** - Configuration strategy
- **[Dependency Injection](adr/ADR-013-dependency-injection.md)** - DI architecture
- **[Performance Budgets](adr/ADR-014-performance-budgets.md)** - Performance targets
- **[Observability](adr/ADR-015-observability.md)** - Monitoring infrastructure
- **[Type-Safe Configuration](adr/ADR-016-type-safe-configuration.md)** - Generic config system
- **[Plugin Lifecycle](adr/ADR-017-plugin-lifecycle.md)** - Lifecycle management

## 📋 Specifications

Detailed product and technical specifications have been organized in the [`specs/`](specs/) directory:

- **[Initial MVP](specs/initial-mvp/)** - Original MVP requirements and implementation
- **[Plugin System](specs/plugin-system/)** - Extensibility through plugins
- **[Branding](specs/branding/)** - Custom branding capabilities
- **[Architectural Improvements](specs/architectural-improvements/)** - Refactoring and technical debt

Each specification contains:
- `product.md` - Product requirements and user stories
- `tech.md` - Technical implementation details

## 🔧 For Developers

If you're looking to:
- **Use Glide**: Start with the [Command Reference](command-reference.md)
- **Fix Issues**: Check the [Troubleshooting Guide](troubleshooting.md)
- **Contribute**: Read the [Contributing Guide](CONTRIBUTING.md)
- **Understand Architecture**: Browse the [ADRs](adr/)
- **Review Specifications**: Visit the [`specs/`](specs/) directory

## 📖 Documentation Standards

### For User Documentation
- Write in clear, concise language
- Include practical examples
- Focus on common use cases
- Keep updated with latest features

### For ADRs
- Follow ADR template structure
- Document context, decision, and consequences
- Date all records accurately
- Link related ADRs

### For Specifications
- Separate product requirements from technical details
- Include success criteria
- Document non-goals explicitly
- Track implementation status

## 🤝 Contributing to Documentation

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on contributing to documentation.
