# Glide Plugin SDK v1 (Deprecated)

> **⚠️ DEPRECATED: SDK v1 is no longer supported.**
>
> All plugins must use **[SDK v2](../v2/)**. See the [Migration Guide](../../../../docs/guides/PLUGIN-SDK-V2-MIGRATION.md) for upgrade instructions.

---

This package contains the legacy SDK v1 implementation. It is retained only for reference purposes during migration. Do not use for new plugin development.

## Migration to SDK v2

SDK v2 provides significant improvements:

- **Type-safe configuration** using Go generics
- **Unified lifecycle management** (Init/Start/Stop/HealthCheck)
- **Declarative command definitions**
- **Simplified API** with `BasePlugin[C]`

See [SDK v2 Migration Guide](../../../../docs/guides/PLUGIN-SDK-V2-MIGRATION.md) for step-by-step migration instructions.
