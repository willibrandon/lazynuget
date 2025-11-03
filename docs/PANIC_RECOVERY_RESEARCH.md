# Go Panic Recovery Strategies for LazyNuGet

## Executive Summary

This research defines panic recovery approaches for LazyNuGet's graceful failure requirement (FR-016). The strategy focuses on recovering from panics during startup and shutdown phases, logging error details with stack traces, and exiting with appropriate exit codes rather than crashing ungracefully.

**Key Finding**: Use layered recovery with `main()` as the outermost safety net and task-specific recovery in critical goroutines. This approach mirrors bubbletea's battle-tested implementation.

---

## 1. Research Overview

### Context & Requirement
**FR-016**: "System MUST recover from panics during startup and shutdown, logging the error and exiting gracefully rather than crashing."

Key concerns:
- Panics during bootstrap should not leave the application in a broken state
- Panics during shutdown should not prevent cleanup sequences from completing
- All panic details must be logged with stack traces for debugging
- Exit codes must reflect the failure (non-zero for panic, typically exit code 2)

### Go Panic Fundamentals
- **Panic**: Runtime exception that unwinds the call stack
- **Defer**: Guaranteed execution before return/panic (LIFO order)
- **Recover**: Catches panic value, only works in deferred functions, returns nil if no panic
- **Stack Trace**: Can be captured via `runtime/debug.Stack()` or `debug.PrintStack()`

---

## 2. Panic Recovery Patterns Analysis

### Pattern 1: Basic Defer-Recover in main()
**Location**: `main()` function
**Scope**: Catches panics from entire application initialization and execution

```go
func main() {
    defer func() {
        if r := recover(); r != nil {
            logger.Fatalf("Application panic: %v\n%s", r, debug.Stack())
            os.Exit(2)
        }
    }()

    // Application code
}
```

**Pros**:
- Catches any unhandled panic
- Minimal overhead
- Works before/after any other initialization
- Simplest to implement

**Cons**:
- Single catch-all may mask bugs
- Doesn't prevent partial initialization damage
- Stack trace shows the main() wrapper, not original location

---

### Pattern 2: Layered Recovery (Recommended for LazyNuGet)
**Locations**: `main()` + component initialization + critical goroutines
**Scope**: Catches panics at multiple levels for better recovery granularity

```go
func main() {
    // Outermost safety net: catches everything
    defer func() {
        if r := recover(); r != nil {
            logger.Fatalf("Critical panic in main: %v\n%s", r, debug.Stack())
            os.Exit(2)
        }
    }()

    app, err := bootstrapApplication()
    if err != nil {
        handleBootstrapError(err)
        os.Exit(1)
    }

    err = app.Run()
    if err != nil {
        handleRunError(err)
        os.Exit(1)
    }
}

func bootstrapApplication() (*App, error) {
    // Component-level recovery: prevents partial initialization
    defer func() {
        if r := recover(); r != nil {
            logger.Error("Panic during bootstrap: %v\n%s", r, debug.Stack())
            // Re-panic to let main() handle graceful exit
            panic(r)
        }
    }()

    // Initialize components
    return nil, nil
}

func (app *App) runGoroutine() {
    // Goroutine-level recovery: prevents goroutine from crashing process
    defer func() {
        if r := recover(); r != nil {
            logger.Error("Panic in background task: %v\n%s", r, debug.Stack())
            // Log but don't re-panic; allow graceful degradation
        }
    }()

    // Long-running task
}
```

**Pros**:
- Catches panics at appropriate levels
- Allows component-level logging and context
- Prevents goroutine panics from killing entire process
- Clear exit points and controlled shutdown
- Used successfully in bubbletea (Charmbracelet)

**Cons**:
- More verbose
- Requires discipline in placement

---

### Pattern 3: Signal-Based Recovery
**Location**: Signal handler + recover
**Scope**: Graceful handling of SIGINT/SIGTERM even during panic

```go
func handleSignals(ctx context.Context, cancel context.CancelFunc) {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        defer func() {
            if r := recover(); r != nil {
                logger.Error("Panic in signal handler: %v\n%s", r, debug.Stack())
            }
        }()

        sig := <-sigChan
        logger.Info("Received signal: %v", sig)
        cancel()
    }()
}
```

**Pros**:
- Ensures shutdown sequence runs even if panic occurs
- Prevents deadlock in signal handling

**Cons**:
- Only handles signal-related panics
- Limited scope

---

## 3. Logging Panics: Best Practices

### Stack Trace Capture
Go provides two main approaches:

**Option A: `runtime/debug.Stack()`** (Recommended)
```go
import "runtime/debug"

defer func() {
    if r := recover(); r != nil {
        stackTrace := string(debug.Stack())
        logger.Error("Panic: %v\nStack:\n%s", r, stackTrace)
    }
}()
```

**Advantages**:
- Returns byte slice, easily convertible to string
- Includes goroutine ID and function information
- Zero overhead in normal path (only captures on panic)
- Used by bubbletea

**Option B: `fmt.Sprintf()` with `%+v`**
```go
import "runtime"

defer func() {
    if r := recover(); r != nil {
        var buf strings.Builder
        pcs := make([]uintptr, 32)
        n := runtime.Callers(1, pcs)

        for _, pc := range pcs[:n] {
            fn := runtime.FuncForPC(pc)
            file, line := fn.FileLine(pc)
            fmt.Fprintf(&buf, "%s:%d %s\n", file, line, fn.Name())
        }

        logger.Error("Panic: %v\nStack:\n%s", r, buf.String())
    }
}()
```

**Advantages**:
- More control over formatting
- Can filter out internal frames

### Logging Best Practices

1. **Log at ERROR level** for startup panics (component unavailable)
2. **Log at WARN level** for non-critical goroutine panics (service degraded)
3. **Always include**:
   - Panic value (e.g., `%v`)
   - Full stack trace
   - Context information (which phase, component name)
   - Timestamp (handled by logger)

```go
logger.WithFields(log.Fields{
    "phase": "startup",
    "component": "config-loader",
    "panic": fmt.Sprintf("%v", r),
}).Error("Fatal panic during initialization")

logger.WithFields(log.Fields{
    "goroutine": "background-worker",
    "panic": fmt.Sprintf("%v", r),
}).Warn("Non-fatal panic in background task")
```

---

## 4. Exit Codes: Convention & Strategy

### Standard POSIX Exit Codes

| Code | Meaning | Use Case |
|------|---------|----------|
| **0** | Success | Normal exit after completion |
| **1** | General error | User/configuration error (recoverable) |
| **2** | Misuse of command | Invalid CLI arguments |
| **126** | Command cannot execute | Permission denied |
| **127** | Command not found | Binary not in PATH |
| **128+N** | Fatal error signal | Killed by signal N |
| **130** | Interrupted (SIGINT) | Ctrl+C by user |
| **143** | Terminated (SIGTERM) | kill -15 |

### LazyNuGet Exit Code Strategy

```go
const (
    ExitSuccess           = 0  // Normal completion
    ExitUserError         = 1  // User error: invalid config, missing dep
    ExitPanic            = 2  // Unrecovered panic
    ExitSignalInterrupt  = 130 // SIGINT (Ctrl+C)
    ExitSignalTerminate  = 143 // SIGTERM
)
```

### Implementation

```go
func main() {
    defer func() {
        if r := recover(); r != nil {
            logger.Fatalf("Panic: %v\n%s", r, debug.Stack())
            os.Exit(ExitPanic) // Exit 2
        }
    }()

    app, err := NewApp()
    if err != nil {
        logger.Error("Startup failed: %v", err)
        os.Exit(ExitUserError) // Exit 1
    }

    err = app.Run()
    if err != nil {
        if errors.Is(err, context.Canceled) {
            os.Exit(ExitSuccess) // Normal shutdown
        }
        logger.Error("Runtime error: %v", err)
        os.Exit(ExitUserError)
    }
}
```

---

## 5. Testing Panic Recovery

### Unit Tests for Panic Recovery

```go
package myapp

import (
    "testing"
    "log"
    "os"
)

func TestRecoveryFromPanicInInit(t *testing.T) {
    // Create a test function that would panic
    testFunc := func() {
        defer func() {
            if r := recover(); r != nil {
                // Panic was caught
                t.Log("Successfully recovered from panic:", r)
                return
            }
            t.Fatal("Expected panic was not caught")
        }()

        initializeBrokenComponent() // This will panic
    }

    testFunc()
}

func TestPanicLogging(t *testing.T) {
    // Capture log output
    var logOutput strings.Builder
    logger := log.New(&logOutput, "", log.Lshortfile)

    defer func() {
        if r := recover(); r != nil {
            logger.Printf("Caught panic: %v", r)

            // Verify log contains panic info
            output := logOutput.String()
            if !strings.Contains(output, "Caught panic") {
                t.Fatal("Panic not logged correctly")
            }
        }
    }()

    panic("test panic")
}
```

### Integration Tests for Exit Codes

```go
func TestMainExitCodeOnPanic(t *testing.T) {
    // This test should be run as a subprocess
    if os.Getenv("TEST_SUBPROCESS") == "1" {
        main() // Will panic and exit
        return
    }

    cmd := exec.Command(os.Args[0], "-test.run=TestMainExitCodeOnPanic")
    cmd.Env = append(os.Environ(), "TEST_SUBPROCESS=1")

    err := cmd.Run()
    exitCode := err.(*exec.ExitError).ExitCode()

    expectedCode := ExitPanic // 2
    if exitCode != expectedCode {
        t.Fatalf("Expected exit code %d, got %d", expectedCode, exitCode)
    }
}
```

### Goroutine Panic Testing

```go
func TestGoroutinePanicRecovery(t *testing.T) {
    done := make(chan struct{})

    go func() {
        defer func() {
            if r := recover(); r != nil {
                t.Log("Goroutine panic recovered:", r)
            }
            close(done)
        }()

        panic("goroutine panic")
    }()

    select {
    case <-done:
        // Panic was handled
    case <-time.After(2 * time.Second):
        t.Fatal("Goroutine panic not recovered in time")
    }
}
```

---

## 6. Real-World Example: Bubbletea Implementation

Bubbletea (Charmbracelet) implements sophisticated panic recovery:

### Goroutine-Level Recovery
```go
// execBatchMsg executes commands concurrently with panic recovery
func (p *Program) execBatchMsg(msg BatchMsg) {
    if !p.startupOptions.has(withoutCatchPanics) {
        defer func() {
            if r := recover(); r != nil {
                p.recoverFromGoPanic(r)
            }
        }()
    }

    // Execute batch commands
}

// execSequenceMsg executes commands sequentially with panic recovery
func (p *Program) execSequenceMsg(msg sequenceMsg) {
    if !p.startupOptions.has(withoutCatchPanics) {
        defer func() {
            if r := recover(); r != nil {
                p.recoverFromGoPanic(r)
            }
        }()
    }

    // Execute sequence commands
}
```

### Panic Handler
```go
func (p *Program) recoverFromGoPanic(r interface{}) {
    // Send panic error on error channel (non-blocking)
    select {
    case p.errs <- ErrProgramPanic:
    default:
    }

    // Cancel execution context
    p.cancel()

    // Restore terminal state
    fmt.Printf("Caught goroutine panic:\n\n%s\n\nRestoring terminal...\n\n", r)
    debug.PrintStack()
}
```

### Key Insights
1. **Configurable**: `WithoutCatchPanics` option allows disabling for debugging
2. **Non-blocking channels**: Prevents deadlock if error channel is full
3. **Terminal restoration**: Ensures terminal state is restored even after panic
4. **Stack printing**: Uses `debug.PrintStack()` for immediate visibility
5. **Graceful degradation**: Panics don't kill the entire program

---

## 7. Decision: Recommended Approach for LazyNuGet

### Strategy: Layered Panic Recovery

**Where to Place**:
1. **Main level** (outermost safety net)
2. **Bootstrap/Initialization** (phase-specific logging)
3. **Critical goroutines** (background tasks)
4. **Signal handler** (graceful shutdown)

**Why This Works**:
- Catches panics at every critical boundary
- Provides phase-specific context in logs
- Prevents goroutines from crashing main process
- Allows graceful cleanup even after panic
- Mirrors production-tested bubbletea approach

---

## 8. Implementation Code Examples

### Main Level Recovery

```go
package main

import (
    "context"
    "fmt"
    "os"
    "runtime/debug"

    "lazynuget/internal/app"
    "lazynuget/internal/log"
)

const (
    ExitSuccess  = 0
    ExitUserErr  = 1
    ExitPanic    = 2
)

func main() {
    // Outermost safety net: catches everything
    defer func() {
        if r := recover(); r != nil {
            fmt.Fprintf(os.Stderr, "FATAL: Panic in main: %v\n", r)
            fmt.Fprintf(os.Stderr, "\nStack trace:\n%s\n", debug.Stack())
            os.Exit(ExitPanic)
        }
    }()

    // Initialize logger early
    logger := log.NewLogger(os.Stderr)

    // Create app context
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Bootstrap application
    application, err := app.New(ctx, logger)
    if err != nil {
        logger.Error("Failed to initialize application: %v", err)
        os.Exit(ExitUserErr)
    }

    // Run application
    err = application.Run()
    if err != nil {
        logger.Error("Application error: %v", err)
        os.Exit(ExitUserErr)
    }

    os.Exit(ExitSuccess)
}
```

### Component-Level Recovery

```go
package app

import (
    "runtime/debug"
    "lazynuget/internal/log"
)

func (a *App) initialize() error {
    // Component-level recovery: catches panics during init
    defer func() {
        if r := recover(); r != nil {
            a.logger.Error("Panic during app initialization: %v", r)
            a.logger.Error("Stack trace:\n%s", debug.Stack())
            // Re-panic to let main() handle exit
            panic(r)
        }
    }()

    // Initialize config
    if err := a.loadConfig(); err != nil {
        return fmt.Errorf("config initialization failed: %w", err)
    }

    // Initialize logging
    if err := a.setupLogging(); err != nil {
        return fmt.Errorf("logging setup failed: %w", err)
    }

    // Initialize platform detection
    if err := a.detectPlatform(); err != nil {
        return fmt.Errorf("platform detection failed: %w", err)
    }

    // Initialize services
    if err := a.setupServices(); err != nil {
        return fmt.Errorf("service initialization failed: %w", err)
    }

    return nil
}
```

### Goroutine-Level Recovery

```go
package app

import (
    "context"
    "runtime/debug"
)

func (a *App) runBackgroundWorker(ctx context.Context, name string) {
    go func() {
        // Goroutine-level recovery: prevents worker crash from killing app
        defer func() {
            if r := recover(); r != nil {
                a.logger.Warn(
                    "Panic in background worker %s: %v",
                    name, r,
                )
                a.logger.Debug("Stack trace:\n%s", debug.Stack())
                // Don't re-panic; allow graceful degradation
            }
        }()

        // Long-running worker logic
        a.workerLoop(ctx)
    }()
}

func (a *App) setupSignalHandlers(ctx context.Context, cancel context.CancelFunc) {
    // Signal-level recovery: ensures cleanup even with panic
    go func() {
        defer func() {
            if r := recover(); r != nil {
                a.logger.Error("Panic in signal handler: %v", r)
                a.logger.Error("Stack trace:\n%s", debug.Stack())
            }
        }()

        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

        sig := <-sigChan
        a.logger.Info("Received signal: %v", sig)
        cancel()
    }()
}
```

### Shutdown-Phase Recovery

```go
package app

import (
    "context"
    "runtime/debug"
    "time"
)

func (a *App) Shutdown(timeout time.Duration) error {
    // Shutdown-phase recovery: ensures cleanup completes
    defer func() {
        if r := recover(); r != nil {
            a.logger.Error("Panic during shutdown: %v", r)
            a.logger.Error("Stack trace:\n%s", debug.Stack())
            // Continue shutdown anyway - don't panic
        }
    }()

    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    // Gracefully shut down services
    return a.shutdownServices(ctx)
}
```

---

## 9. Testing Strategy

### Unit Test: Panic Recovery in Bootstrap

```go
func TestBootstrapPanicRecovery(t *testing.T) {
    done := make(chan struct{})
    var capturedPanic interface{}

    go func() {
        defer func() {
            if r := recover(); r != nil {
                capturedPanic = r
            }
            close(done)
        }()

        // This will panic
        app, err := NewApp(context.Background(), nil)
        assert.Error(t, err)
    }()

    select {
    case <-done:
        if capturedPanic != nil {
            t.Logf("Panic recovered: %v", capturedPanic)
        }
    case <-time.After(2 * time.Second):
        t.Fatal("Panic recovery timeout")
    }
}
```

### Integration Test: Exit Code on Panic

```go
func TestMainExitCodeOnPanic(t *testing.T) {
    cmd := exec.Command(os.Args[0])
    // Set environment variable to trigger panic in test mode
    cmd.Env = append(os.Environ(), "TEST_TRIGGER_PANIC=1")

    err := cmd.Run()
    if err == nil {
        t.Fatal("Expected non-zero exit code")
    }

    exitCode := cmd.ProcessState.ExitCode()
    expectedCode := 2 // ExitPanic

    if exitCode != expectedCode {
        t.Fatalf("Expected exit code %d, got %d", expectedCode, exitCode)
    }
}
```

---

## 10. Checklist for Implementation

- [ ] Add `ExitCode` constants to main package
- [ ] Implement main-level defer/recover block
- [ ] Add component-level defer/recover in `App.Initialize()`
- [ ] Add goroutine-level defer/recover in background workers
- [ ] Add signal handler-level defer/recover
- [ ] Add shutdown-phase defer/recover
- [ ] Verify all panics log stack traces with `runtime/debug.Stack()`
- [ ] Test exit codes match expected values (0, 1, 2)
- [ ] Test goroutine panics don't crash main process
- [ ] Test shutdown completes even with panic during shutdown
- [ ] Add panic recovery tests to CI/CD
- [ ] Document panic recovery strategy in code comments

---

## References

1. **Bubbletea Panic Recovery**
   - Source: `/Users/brandon/src/bubbletea/tea.go`
   - Implementation: `recoverFromGoPanic()` method
   - Pattern: Configurable recovery with terminal restoration

2. **Go Standard Library**
   - `runtime/debug.Stack()` - Capture stack trace
   - `os.Exit(code)` - Exit with code
   - `defer` - Guaranteed execution
   - `recover()` - Catch panics

3. **Industry Standards**
   - POSIX exit codes: 0=success, 1=general error, 2=misuse
   - Signal codes: 128+N format
   - Stack traces: Essential for debugging

---

## Summary

LazyNuGet should implement **layered panic recovery** at:
1. **Main level** - Ultimate safety net
2. **Bootstrap phase** - Initialization-specific
3. **Critical goroutines** - Background tasks
4. **Signal handling** - Graceful shutdown
5. **Shutdown phase** - Cleanup guarantee

This approach ensures:
- Panics are logged with full stack traces
- Exit codes correctly indicate failure type
- Goroutines don't crash the application
- Shutdown completes gracefully even after panic
- Clear debugging information is available in logs

Exit code convention: **0** (success), **1** (user error), **2** (panic)
