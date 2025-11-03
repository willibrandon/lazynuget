# Panic Recovery Quick Reference

## One-Line Summary
Use **layered defer/recover at 5 levels**: main (safety net) → bootstrap (context) → goroutines (degrade) → signals (shutdown) → shutdown (guarantee).

## Pattern Quick Copy-Paste

### Pattern 1: Main-Level Safety Net
```go
func main() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Fprintf(os.Stderr, "PANIC: %v\nStack:\n%s\n", r, debug.Stack())
            os.Exit(2)
        }
    }()
    // ... rest of main
}
```

### Pattern 2: Component-Level Recovery (Re-panic)
```go
func (a *App) Bootstrap() error {
    defer func() {
        if r := recover(); r != nil {
            a.logger.Error("Panic in bootstrap: %v\n%s", r, debug.Stack())
            panic(r) // Re-panic for main() to catch
        }
    }()
    // ... initialization code
}
```

### Pattern 3: Goroutine-Level Recovery (Don't Re-panic)
```go
func (a *App) RunWorker(name string) {
    go func() {
        defer func() {
            if r := recover(); r != nil {
                a.logger.Warn("Panic in %s: %v\n%s", name, r, debug.Stack())
                // DON'T re-panic - allow graceful degradation
            }
        }()
        // ... worker code
    }()
}
```

### Pattern 4: Shutdown Safety
```go
func (a *App) Shutdown(timeout time.Duration) error {
    defer func() {
        if r := recover(); r != nil {
            a.logger.Error("Panic during shutdown: %v\n%s", r, debug.Stack())
            // Continue cleanup anyway - don't panic
        }
    }()
    // ... shutdown code
}
```

## Exit Codes
```
0 = Success
1 = User/config error
2 = Panic
```

## Imports Needed
```go
import (
    "fmt"
    "os"
    "runtime/debug"
)
```

## Key Rules
1. **Always use `runtime/debug.Stack()`** - captures full stack trace
2. **Log at every level** - ERROR for startup, WARN for workers
3. **Re-panic in bootstrap** - let main() handle graceful exit
4. **Don't re-panic in goroutines** - prevents cascading crashes
5. **Complete shutdown even if panic** - never interrupt cleanup

## Testing Checklist
- [ ] Unit test panic recovery catches panic
- [ ] Unit test panic recovery logs stack trace
- [ ] Integration test exit code is 2 on panic
- [ ] Goroutine panic doesn't kill main process
- [ ] Shutdown completes even with panic in shutdown handler

## Why This Works
- **Main level**: Ultimate safety net catches everything
- **Component level**: Adds context about what phase failed
- **Goroutine level**: Prevents background task from crashing app
- **Signal level**: Ensures clean shutdown even during panic
- **Shutdown level**: Guarantees cleanup completes

## Common Mistakes to Avoid
❌ Single defer/recover in main only → No component context
❌ Using `fmt.Sprintf("%+v", err)` → Doesn't capture stack
❌ Re-panicking in goroutines → Cascading crashes
❌ Not logging panic at all → No debugging info
❌ Interrupting shutdown on panic → Incomplete cleanup

## File Locations in LazyNuGet
- **Main recovery**: `cmd/lazynuget/main.go`
- **Component recovery**: `internal/app/app.go` in Bootstrap()
- **Worker recovery**: `internal/app/worker.go` or similar
- **Signal recovery**: `internal/app/signals.go` or similar
- **Shutdown recovery**: `internal/app/app.go` in Shutdown()

## See Also
- Full research: `/Users/brandon/src/lazynuget/docs/PANIC_RECOVERY_RESEARCH.md`
- Decision document: `/Users/brandon/src/lazynuget/docs/FR-016-PANIC-RECOVERY-DECISION.md`
- Bubbletea implementation: `/Users/brandon/src/bubbletea/tea.go` (lines ~350-400)
