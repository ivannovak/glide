# ADR-017: Plugin Lifecycle Management

## Status
Accepted

## Date
2025-11-28

## Context

The v1 plugin system had limited lifecycle management:

1. **No Defined States**: Plugins were either loaded or not
2. **No Graceful Shutdown**: Plugins terminated abruptly
3. **No Health Checking**: No way to verify plugin health
4. **Resource Leaks**: Connections/files not properly closed
5. **Initialization Ordering**: Dependencies not respected

This caused issues in production:
- Plugins holding resources after errors
- Intermittent failures during shutdown
- No visibility into plugin health
- Race conditions during startup

## Decision

We implemented a comprehensive lifecycle management system with:

1. **Defined State Machine**: Clear states and transitions
2. **Lifecycle Hooks**: Init, Start, Stop, HealthCheck
3. **State Tracking**: Observable state for each plugin
4. **Dependency-Aware Loading**: Respect plugin dependencies
5. **Graceful Shutdown**: Proper cleanup with timeout

### State Machine

```
                    ┌─────────────────────────────────────┐
                    │                                     │
                    ▼                                     │
┌──────────┐   ┌─────────┐   ┌────────────┐   ┌───────┐  │
│Discovered│──▶│ Loading │──▶│Initializing│──▶│ Ready │──┤
└──────────┘   └─────────┘   └────────────┘   └───────┘  │
                    │              │              │       │
                    │              │              ▼       │
                    │              │         ┌─────────┐  │
                    │              │         │ Running │──┘
                    │              │         └─────────┘
                    │              │              │
                    ▼              ▼              ▼
               ┌─────────┐   ┌─────────┐   ┌──────────┐
               │  Error  │   │  Error  │   │ Stopping │
               └─────────┘   └─────────┘   └──────────┘
                                                │
                                                ▼
                                           ┌─────────┐
                                           │ Stopped │
                                           └─────────┘
```

### States

| State | Description |
|-------|-------------|
| Discovered | Found in plugin directory |
| Loading | Binary being loaded, gRPC connection |
| Initializing | Init() called, one-time setup |
| Ready | Initialized, waiting for Start() |
| Running | Started, executing commands |
| Stopping | Stop() called, cleanup in progress |
| Stopped | Fully stopped, resources released |
| Error | Error occurred, may be recoverable |

## Consequences

### Positive

1. **Predictable Behavior**: Clear state transitions
2. **Resource Safety**: Proper cleanup guaranteed
3. **Health Visibility**: Can check plugin health
4. **Debugging**: State history aids troubleshooting
5. **Dependency Respect**: Plugins start in order

### Negative

1. **Complexity**: More states to manage
2. **Overhead**: State tracking adds small cost
3. **Migration**: v1 plugins need adaptation

## Implementation

### Lifecycle Interface

```go
// pkg/plugin/sdk/v2/lifecycle.go
type Lifecycle interface {
    // Init is called once after plugin load
    Init(ctx context.Context) error

    // Start is called when plugin should begin operation
    Start(ctx context.Context) error

    // Stop is called for graceful shutdown
    Stop(ctx context.Context) error

    // HealthCheck returns nil if healthy
    HealthCheck(ctx context.Context) error
}
```

### State Tracker

```go
// pkg/plugin/sdk/state.go
type StateTracker struct {
    mu       sync.RWMutex
    current  State
    history  []StateEntry
    created  time.Time
}

type StateEntry struct {
    State     State
    Timestamp time.Time
    Message   string
    Error     error
}

func (st *StateTracker) TransitionTo(state State, message string) error {
    st.mu.Lock()
    defer st.mu.Unlock()

    // Validate transition
    if !st.canTransition(state) {
        return fmt.Errorf("invalid transition: %s -> %s", st.current, state)
    }

    st.history = append(st.history, StateEntry{
        State:     state,
        Timestamp: time.Now(),
        Message:   message,
    })
    st.current = state
    return nil
}
```

### Lifecycle Manager

```go
// pkg/plugin/sdk/lifecycle_manager.go
type LifecycleManager struct {
    mu           sync.RWMutex
    plugins      map[string]*ManagedPlugin
    shutdownOnce sync.Once
}

func (lm *LifecycleManager) Start(ctx context.Context, name string) error {
    lm.mu.Lock()
    plugin := lm.plugins[name]
    lm.mu.Unlock()

    // Transition to starting
    plugin.State.TransitionTo(StateRunning, "Starting plugin")

    // Call Start hook
    if err := plugin.Plugin.Start(ctx); err != nil {
        plugin.State.TransitionTo(StateError, err.Error())
        return err
    }

    return nil
}

func (lm *LifecycleManager) StopAll(ctx context.Context) error {
    lm.shutdownOnce.Do(func() {
        // Stop in reverse dependency order
        for _, name := range lm.reverseOrder() {
            lm.stopPlugin(ctx, name)
        }
    })
    return nil
}
```

### Plugin Implementation

```go
type MyPlugin struct {
    v2.BasePlugin[Config]
    client *http.Client
    conn   *grpc.ClientConn
}

func (p *MyPlugin) Init(ctx context.Context) error {
    // One-time initialization
    p.client = &http.Client{Timeout: 30 * time.Second}
    return nil
}

func (p *MyPlugin) Start(ctx context.Context) error {
    // Connect to services
    var err error
    p.conn, err = grpc.Dial(p.GetConfig().ServerAddr)
    return err
}

func (p *MyPlugin) Stop(ctx context.Context) error {
    // Cleanup resources
    if p.conn != nil {
        return p.conn.Close()
    }
    return nil
}

func (p *MyPlugin) HealthCheck(ctx context.Context) error {
    // Verify connectivity
    _, err := p.client.Get(p.GetConfig().HealthURL)
    return err
}
```

### Dependency-Aware Loading

```go
// Plugin dependencies declared in metadata
func (p *MyPlugin) Metadata() v2.Metadata {
    return v2.Metadata{
        Name:         "my-plugin",
        Dependencies: []string{"base-plugin", "auth-plugin"},
    }
}

// Resolver ensures correct order
resolver := sdk.NewDependencyResolver()
resolver.AddPlugin("my-plugin", []string{"base-plugin", "auth-plugin"})
resolver.AddPlugin("base-plugin", nil)
resolver.AddPlugin("auth-plugin", []string{"base-plugin"})

order, err := resolver.Resolve()
// order = ["base-plugin", "auth-plugin", "my-plugin"]
```

## Alternatives Considered

### 1. Simple Start/Stop

Just Start() and Stop() without states.

**Rejected because**:
- No visibility into current state
- No initialization phase
- No health checking

### 2. Event-Based Lifecycle

Publish lifecycle events, let plugins subscribe.

**Rejected because**:
- More complex
- Harder to guarantee order
- Less explicit

### 3. Container-Managed Lifecycle

Let DI container manage plugin lifecycle.

**Rejected because**:
- Plugins are external processes
- Container can't manage gRPC plugins
- Different lifecycle model needed

## Graceful Shutdown

Shutdown with timeout and ordering:

```go
func (lm *LifecycleManager) Shutdown(timeout time.Duration) {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    // Stop in reverse dependency order
    order := lm.reverseOrder()

    for _, name := range order {
        plugin := lm.plugins[name]

        // Attempt graceful stop
        done := make(chan error, 1)
        go func() {
            done <- plugin.Stop(ctx)
        }()

        select {
        case err := <-done:
            if err != nil {
                log.Warn("Plugin stop error", "plugin", name, "error", err)
            }
        case <-ctx.Done():
            log.Warn("Plugin stop timeout", "plugin", name)
            // Force kill
        }
    }
}
```

## References

- [HashiCorp go-plugin](https://github.com/hashicorp/go-plugin)
- [Kubernetes Pod Lifecycle](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/)
- [Plugin Development Guide](../guides/plugin-development.md)
- [SDK v2 Migration Guide](../guides/PLUGIN-SDK-V2-MIGRATION.md)
- [pkg/plugin/sdk Documentation](../../pkg/plugin/sdk/doc.go)
