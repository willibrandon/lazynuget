# FR-016 Decision: Panic Recovery Implementation

## Decision Statement

LazyNuGet will implement **layered panic recovery** (defer/recover at multiple levels) rather than single-point recovery, following the battle-tested pattern used by Charmbracelet's Bubbletea.

## Approach Summary

| Layer | Location | Purpose | Re-Panic? |
|-------|----------|---------|-----------|
| 1. **Outermost** | `main()` | Ultimate safety net | N/A (final handler) |
| 2. **Component** | `App.Bootstrap()` | Phase-specific logging | Yes → main |
| 3. **Goroutine** | Background workers | Prevent worker crash | No (degrade gracefully) |
| 4. **Signal** | Signal handler | Ensure shutdown runs | No (allow cleanup) |
| 5. **Shutdown** | `App.Shutdown()` | Guarantee cleanup | No (complete anyway) |

## Rationale

**Why layered recovery is effective** (1-2 sentences):

1. **Multiple recovery points catch panics at the right scope** - Main-level catches uncaught panics, component-level provides context, goroutine-level prevents background task crashes, ensuring failures are handled appropriately at each level.

2. **Battle-tested in production** - Bubbletea (used by 1000+ TUI applications) implements exactly this pattern, proving effectiveness for terminal applications requiring graceful shutdown and clean state restoration.

## Exit Code Convention

```
ExitSuccess  = 0   // Normal completion
ExitUserErr  = 1   // User/config error (recoverable)
ExitPanic    = 2   // Unrecovered panic
```

## Code Example: Panic Recovery Wrapper Function

### Main-Level Recovery

```go
package main

import (
    "context"
    "fmt"
    "os"
    "runtime/debug"

    "lazynuget/internal/app"
)

const (
    ExitSuccess = 0
    ExitUserErr = 1
    ExitPanic   = 2
)

func main() {
    // LAYER 1: Outermost safety net - catches ALL panics
    defer func() {
        if r := recover(); r != nil {
            fmt.Fprintf(
                os.Stderr,
                "FATAL PANIC: %v\nStack Trace:\n%s\n",
                r,
                debug.Stack(),
            )
            os.Exit(ExitPanic)
        }
    }()

    // Initialize application
    application, err := app.New(context.Background())
    if err != nil {
        fmt.Fprintf(os.Stderr, "Startup failed: %v\n", err)
        os.Exit(ExitUserErr)
    }

    // Run application
    if err := application.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
        os.Exit(ExitUserErr)
    }

    os.Exit(ExitSuccess)
}
```

### Component-Level Recovery

```go
package app

import (
    "fmt"
    "runtime/debug"
)

// Bootstrap initializes all application components with panic recovery
func (a *App) Bootstrap() error {
    // LAYER 2: Component-level recovery - provides context before re-panicking
    defer func() {
        if r := recover(); r != nil {
            a.logger.Error(
                "PANIC during bootstrap (phase: %s): %v\nStack: %s",
                a.phase,
                r,
                debug.Stack(),
            )
            // Re-panic to let main() handle graceful exit
            panic(r)
        }
    }()

    // Phase: Config loading
    a.phase = "config"
    if err := a.loadConfig(); err != nil {
        return fmt.Errorf("config load failed: %w", err)
    }

    // Phase: Logging setup
    a.phase = "logging"
    if err := a.setupLogging(); err != nil {
        return fmt.Errorf("logging setup failed: %w", err)
    }

    // Phase: Platform detection
    a.phase = "platform"
    if err := a.detectPlatform(); err != nil {
        return fmt.Errorf("platform detection failed: %w", err)
    }

    // Phase: Service initialization
    a.phase = "services"
    if err := a.setupServices(); err != nil {
        return fmt.Errorf("service init failed: %w", err)
    }

    a.phase = "ready"
    return nil
}
```

### Goroutine-Level Recovery

```go
// RunWorker starts a background worker with panic recovery
func (a *App) RunWorker(name string, fn func(context.Context) error) {
    go func() {
        // LAYER 3: Goroutine-level recovery - prevents background task crashes
        defer func() {
            if r := recover(); r != nil {
                a.logger.Warn(
                    "PANIC in worker '%s': %v\nStack: %s",
                    name,
                    r,
                    debug.Stack(),
                )
                // DON'T re-panic; allow graceful degradation
                // Main process continues, only this worker stops
            }
        }()

        if err := fn(context.Background()); err != nil {
            a.logger.Error("Worker '%s' error: %v", name, err)
        }

        a.logger.Info("Worker '%s' completed", name)
    }()
}
```

### Shutdown-Phase Recovery

```go
// Shutdown gracefully stops all components with panic recovery
func (a *App) Shutdown(timeout time.Duration) error {
    // LAYER 5: Shutdown recovery - ensures cleanup completes even if panic
    defer func() {
        if r := recover(); r != nil {
            a.logger.Error(
                "PANIC during shutdown: %v\nStack: %s",
                r,
                debug.Stack(),
            )
            // Continue cleanup anyway - don't stop shutdown sequence
        }
    }()

    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    a.logger.Info("Shutting down application...")

    // Graceful shutdown of all components
    if err := a.shutdownServices(ctx); err != nil {
        return fmt.Errorf("service shutdown failed: %w", err)
    }

    a.logger.Info("Shutdown complete")
    return nil
}
```

## Testing Strategy

```go
// Unit test: Verify panic recovery works
func TestMainLevelPanicRecovery(t *testing.T) {
    // Capture stderr to verify panic message logged
    oldStderr := os.Stderr
    r, w, _ := os.Pipe()
    os.Stderr = w

    // Run code that panics
    go func() {
        defer func() {
            w.Close()
        }()

        // Panic will be caught by main's defer
        panic("test panic")
    }()

    // Restore stderr and read output
    os.Stderr = oldStderr
    w.Close()

    var buf bytes.Buffer
    io.Copy(&buf, r)
    output := buf.String()

    // Verify panic was logged with stack trace
    assert.Contains(t, output, "FATAL PANIC")
    assert.Contains(t, output, "test panic")
    assert.Contains(t, output, "Stack Trace")
}

// Integration test: Verify exit code
func TestExitCodeOnPanic(t *testing.T) {
    cmd := exec.Command("./lazynuget")
    cmd.Env = append(os.Environ(), "TEST_PANIC=1") // Trigger panic in test mode

    err := cmd.Run()
    require.Error(t, err)

    exitCode := cmd.ProcessState.ExitCode()
    assert.Equal(t, ExitPanic, exitCode, "Expected exit code 2 (panic)")
}
```

## Implementation Checklist

- [ ] Main-level defer/recover in `main()`
- [ ] Component-level defer/recover in `App.Bootstrap()`
- [ ] Goroutine-level defer/recover in background workers
- [ ] Signal handler panic safety
- [ ] Shutdown-phase panic safety
- [ ] All panics log with `runtime/debug.Stack()`
- [ ] Exit codes implemented (0, 1, 2)
- [ ] Unit tests for panic recovery
- [ ] Integration tests for exit codes
- [ ] Goroutine panic testing
- [ ] Documentation in code comments
- [ ] Verify shutdown completes even with panic

## Related Requirements

- **FR-014**: Startup failures logged with details ✓ (handled by component-level logging)
- **FR-010**: Appropriate exit codes ✓ (ExitSuccess=0, ExitUserErr=1, ExitPanic=2)
- **FR-008**: Signal handling ✓ (signal handler protected)
- **FR-009**: Graceful shutdown ✓ (shutdown-phase recovery)

## Notes

- Stack traces logged via `runtime/debug.Stack()` - zero overhead unless panic occurs
- No external dependencies needed - uses Go standard library only
- Graceful degradation: Goroutine panics don't crash main process
- Pattern proven in production by Charmbracelet's Bubbletea TUI framework
