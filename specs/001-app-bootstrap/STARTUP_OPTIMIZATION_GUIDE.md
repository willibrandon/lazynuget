# LazyNuGet Startup Optimization: Implementation Guide

**Scope**: Actionable decisions and code patterns for achieving <200ms p95 cold start
**Target Audience**: Developers implementing app bootstrap (Phase 2)
**Related Documents**: `research.md` (detailed background), `spec.md` (requirements)

---

## Quick Reference: Optimization Stack

| Priority | Technique | Time Savings | Complexity | Status |
|----------|-----------|--------------|-----------|--------|
| **P1** | Lazy GUI init (defer Bubbletea) | 80-120ms | Low | Design in Phase 1 |
| **P1** | Async dotnet validation | 30-50ms | Medium | Design in Phase 1 |
| **P1** | Build flags (-w -s -ldflags) | 5-10ms | Trivial | Use in build.sh |
| **P2** | Parallel init (errgroup) | 10-20ms | Low | Implement in bootstrap.go |
| **P2** | Startup instrumentation | 0ms cost | Low | Add timing markers |
| **P3** | Avoid init() functions | 5-30ms | Design | Code review gate |

**Estimated Total Savings**: 130-210ms → Target <200ms achieved ✓

---

## Section 1: Lazy GUI Initialization (P1)

### What Happens Normally (SLOW)

```go
// Typical startup sequence (BLOCKS startup)
func main() {
    // ... early startup ...

    // PROBLEM: Bubbletea init takes 80-120ms
    program := tea.NewProgram(model)

    // PROBLEM: This runs immediately and blocks
    if _, err := program.Run(); err != nil {
        log.Fatal(err)
    }
}
```

**Cost**: 80-120ms spent loading Bubbletea, initializing TTY, starting event loop

### The Lazy Pattern (FAST)

Instead of initializing GUI immediately, defer until first frame needs to render:

```go
// internal/bootstrap/app.go
package bootstrap

import (
    "sync"
    tea "github.com/charmbracelet/bubbletea"
)

type App struct {
    config Config
    logger Logger

    // GUI is initialized lazily on first access
    guiOnce sync.Once
    gui     *GUI
    guiErr  error
}

// GetOrInitGUI returns GUI, initializing on first call
func (a *App) GetOrInitGUI() (*GUI, error) {
    a.guiOnce.Do(func() {
        a.gui, a.guiErr = NewGUI(a.config, a.logger)
    })
    return a.gui, a.guiErr
}

// Run the application (called from main)
func (a *App) Run() error {
    // Fast path: validate CLI args, load config, check TTY
    // This completes in <50ms

    // If --version or --help, return immediately
    if a.config.ShowVersion {
        fmt.Println(a.config.Version)
        return nil
    }
    if a.config.ShowHelp {
        fmt.Println(usage)
        return nil
    }

    // Only NOW initialize GUI (deferred until needed)
    gui, err := a.GetOrInitGUI()  // 80-120ms HERE
    if err != nil {
        return fmt.Errorf("gui init failed: %w", err)
    }

    // Run Bubbletea event loop
    if _, err := gui.Run(); err != nil {
        return err
    }

    return nil
}
```

### Architectural Flow with Lazy Init

```
main.go (1ms)
  ↓
Parse flags (5ms)
  ↓
Load config (10ms)
  ↓
Initialize logging (5ms)
  ↓
Detect platform (5ms)
  ↓
Check for --version/--help (1ms)
  ↓
Return early if flags requested (29ms total for --version)
  ↓
OTHERWISE: Initialize GUI LAZILY (80-120ms deferred)
  ↓
Run Bubbletea event loop
  ↓
GUI shows at ~100-140ms total (still <200ms)
```

### Implementation Checklist

- [ ] Define `App` struct with `guiOnce` and `gui` fields
- [ ] Create `GetOrInitGUI()` method using sync.Once pattern
- [ ] Move `tea.NewProgram()` into GUI init function, not main
- [ ] Update `main.go` to call `app.GetOrInitGUI()` only when needed
- [ ] Test `--version` completes in <50ms
- [ ] Test `--help` completes in <50ms
- [ ] Test normal startup (TUI mode) completes in <200ms

---

## Section 2: Async External Validation (P1)

### Problem: Dotnet CLI Validation Blocks Startup

```go
// SLOW: Validation blocks startup
func main() {
    // User waits for dotnet to respond...
    if err := validateDotnetCLI(); err != nil {
        log.Fatal(err)
    }
    // 30-100ms delay while we check if dotnet exists

    // ONLY THEN start GUI
    gui.Run()
}
```

**Cost**: 30-100ms wasted before GUI even starts

### The Async Pattern (FAST)

Validate dotnet in background while showing UI:

```go
// internal/bootstrap/app.go
package bootstrap

import "time"

type App struct {
    config Config
    logger Logger

    // Async validation state
    dotnetValidated bool
    dotnetErr       error
    dotnetCheckTime time.Time
}

// BootstrapAsync starts fast initialization, validates dotnet async
func (a *App) BootstrapAsync(ctx context.Context) error {
    // Fast path: flags, config, logging (~30ms)
    if err := a.parseFlagsAndConfig(); err != nil {
        return err
    }

    // Start async validation (doesn't block startup)
    go a.validateDotnetAsync(ctx)

    // Fast path: return immediately
    return nil
}

// validateDotnetAsync checks dotnet availability in background
func (a *App) validateDotnetAsync(ctx context.Context) {
    // Run this without blocking main startup
    dotnetFound := checkDotnetCLI()

    a.dotnetValidated = true
    if !dotnetFound {
        a.dotnetErr = ErrDotnetNotFound
        a.logger.Warn("dotnet CLI not found - some features unavailable")
    }

    a.dotnetCheckTime = time.Now()
}

// GetDotnetStatus returns validation result, waiting if needed
func (a *App) GetDotnetStatus() error {
    // If already checked, return cached result
    if a.dotnetValidated {
        return a.dotnetErr
    }

    // If still validating, wait up to 3 seconds
    start := time.Now()
    for time.Since(start) < 3*time.Second {
        if a.dotnetValidated {
            return a.dotnetErr
        }
        time.Sleep(10 * time.Millisecond)
    }

    // Timeout - assume validation taking long
    // Show warning in UI but don't block
    return nil
}
```

### GUI Integration: Show Validation Status

In your TUI model, show validation spinner:

```go
// internal/gui/model.go
type Model struct {
    dotnetValidated bool
    dotnetErr       error
    app             *bootstrap.App
}

// View shows UI with validation status
func (m Model) View() string {
    if !m.dotnetValidated {
        status := m.app.GetDotnetStatus()
        if status == nil {
            // Not done validating yet, show spinner
            return "Validating dotnet CLI... "
        }
        m.dotnetErr = status
        m.dotnetValidated = true
    }

    // ... rest of UI ...
}
```

### Implementation Checklist

- [ ] Move `validateDotnetCLI()` to run in background goroutine
- [ ] Store validation result in `App` struct
- [ ] Return from bootstrap before validation completes
- [ ] Show validation spinner in UI while checking
- [ ] Cache result so second access returns immediately
- [ ] Timeout validation at 3-5 seconds (don't block forever)
- [ ] Test startup completes in <200ms even if dotnet check takes 100ms

---

## Section 3: Build Optimization Flags (P1)

### Current Build

```bash
go build ./cmd/lazynuget
# Output: lazynuget binary (~12-15MB)
```

**Startup cost**: Larger binary = longer to load from disk

### Optimized Build

Update your build script or Makefile:

```bash
#!/bin/bash
# scripts/build.sh

VERSION=${VERSION:-dev}
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%SZ')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

go build \
  -o lazynuget \
  -ldflags="-w -s -X 'main.Version=${VERSION}' -X 'main.BuildTime=${BUILD_TIME}' -X 'main.GitCommit=${GIT_COMMIT}'" \
  -trimpath \
  ./cmd/lazynuget
```

**Flags Explained**:
- `-w`: Strip DWARF debug symbols
- `-s`: Strip symbol table
- `-X 'main.Version=...'`: Inject version at build time (no init() needed)
- `-trimpath`: Remove absolute paths (reproducible builds)

**Result**: Binary shrinks from ~15MB → ~8-10MB (30-35% reduction)

### Makefile Integration

```makefile
# Makefile
VERSION ?= dev
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%SZ')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

LDFLAGS := -ldflags="-w -s -X 'main.Version=$(VERSION)' -X 'main.BuildTime=$(BUILD_TIME)' -X 'main.GitCommit=$(GIT_COMMIT)'"

.PHONY: build
build:
	@go build $(LDFLAGS) -trimpath -o lazynuget ./cmd/lazynuget
	@ls -lh lazynuget
	@echo "Built lazynuget ($(VERSION))"

.PHONY: release
release:
	@VERSION=$(shell git describe --tags) make build
```

**Usage**:
```bash
make build          # Build dev version
make VERSION=1.0.0 build  # Build release
```

### Verification

```bash
# Check binary size
ls -lh lazynuget

# Check if debug symbols stripped
file lazynuget  # Should show "not stripped" only if -w/-s not applied

# Verify version injection
./lazynuget --version  # Should show injected version
```

### Implementation Checklist

- [ ] Update build script with -w -s flags
- [ ] Add -ldflags with version injection
- [ ] Use -trimpath for reproducible builds
- [ ] Test binary size < 10MB
- [ ] Verify --version shows correct value
- [ ] Add build command to Makefile
- [ ] Update CI/CD to use optimized build

---

## Section 4: Parallel Initialization (P2)

### Normal Sequential Init (50ms)

```go
// internal/bootstrap/app.go
func (a *App) initializeServices() error {
    // Sequential: each waits for previous
    config, err := loadConfig()      // 10ms
    if err != nil {
        return err
    }

    logger, err := initLogger()      // 5ms (depends on nothing, could be parallel)
    if err != nil {
        return err
    }

    platform, err := detectPlatform()  // 5ms (depends on nothing, could be parallel)
    if err != nil {
        return err
    }

    // Total: 20ms sequential
    // But logger and platform are independent!
    return nil
}
```

### Parallel Init with errgroup (15ms)

```go
// internal/bootstrap/app.go
package bootstrap

import (
    "context"
    "golang.org/x/sync/errgroup"
)

func (a *App) initializeServices(ctx context.Context) error {
    g, ctx := errgroup.WithContext(ctx)

    // These three operations are independent
    // Paralelization potential: 2-3x speedup

    var config Config
    var logger Logger
    var platform PlatformInfo

    g.Go(func() error {
        var err error
        config, err = loadConfig()  // 10ms
        return err
    })

    g.Go(func() error {
        var err error
        logger, err = initLogger()  // 5ms (now parallel with others)
        return err
    })

    g.Go(func() error {
        var err error
        platform, err = detectPlatform()  // 5ms (now parallel with others)
        return err
    })

    if err := g.Wait(); err != nil {
        return fmt.Errorf("initialization failed: %w", err)
    }

    a.config = config
    a.logger = logger
    a.platform = platform
    return nil
}
```

**Speedup**: 20ms → 10-15ms (depends on hardware cores)

### Important Constraints

```go
// ❌ DON'T parallelize these - they have dependencies!

// Config must load first (determines logger level)
func (a *App) init() error {
    g, ctx := errgroup.WithContext(context.Background())

    var config Config
    var logger Logger

    g.Go(func() error {
        // WRONG: logger init depends on config
        logger, _ = initLogger(config)
        return nil
    })

    g.Go(func() error {
        // WRONG: config not ready yet
        config, _ = loadConfig()
        return nil
    })

    // ❌ Config and logger can't parallelize - logger depends on config!
    return nil
}

// ✅ CORRECT: Load config first, then logger
func (a *App) init() error {
    // Sequential: config determines logger
    config, err := loadConfig()
    if err != nil {
        return err
    }

    g, ctx := errgroup.WithContext(context.Background())

    var logger Logger
    var platform PlatformInfo

    g.Go(func() error {
        // Logger depends on config (now ready)
        var err error
        logger, err = initLogger(config)
        return err
    })

    g.Go(func() error {
        // Platform independent
        var err error
        platform, err = detectPlatform()
        return err
    })

    // ✅ logger and platform can parallelize
    return g.Wait()
}
```

### Real Dependency Graph

```
Flags parsed (must be first)
    ↓
Config loaded (depends on flags)
    ↓
    ├─→ Logger init (depends on config, run parallel)
    ├─→ Platform detect (independent, run parallel)
    └─→ Validation (independent, run parallel)
    ↓
GUI init (depends on config + logger)
```

### Implementation Checklist

- [ ] Identify independent initialization tasks
- [ ] Map out dependency graph
- [ ] Import `golang.org/x/sync/errgroup`
- [ ] Group independent tasks into `g.Go()` calls
- [ ] Ensure config is loaded before dependent ops
- [ ] Test error handling (first error cancels others)
- [ ] Measure timing improvement (~10ms savings)
- [ ] Verify no race conditions with shared state

---

## Section 5: Startup Instrumentation (P2)

### Add Timing Markers

Insert minimal timing code to identify bottlenecks:

```go
// cmd/lazynuget/main.go
package main

import (
    "fmt"
    "os"
    "time"
    "lazynuget/internal/bootstrap"
)

func main() {
    sw := bootstrap.NewStopwatch()

    // Mark each major phase
    defer func() {
        if os.Getenv("DEBUG_STARTUP") == "1" {
            fmt.Fprintf(os.Stderr, "\n%s", sw.Report())
        }
    }()

    // Phase 1: Initialize app container
    app := bootstrap.NewApp()
    sw.Mark("app_created")

    // Phase 2: Parse CLI flags
    if err := app.ParseFlags(); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
    sw.Mark("flags_parsed")

    // Phase 3: Load configuration
    if err := app.LoadConfig(); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
    sw.Mark("config_loaded")

    // Phase 4: Initialize logging
    if err := app.InitializeLogging(); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
    sw.Mark("logging_ready")

    // Phase 5: Run application
    if err := app.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
}
```

### Sample Output

```bash
$ DEBUG_STARTUP=1 ./lazynuget

Startup Timeline:
  app_created: 1ms
  flags_parsed: +4ms (total: 5ms)
  config_loaded: +9ms (total: 14ms)
  logging_ready: +5ms (total: 19ms)
  [GUI init deferred: 80-120ms happens here on Run()]
```

### Instrumentation Checklist

- [ ] Create `Stopwatch` struct with `Mark()` method
- [ ] Add marks at key startup phases
- [ ] Make report output readable
- [ ] Make it conditional on `DEBUG_STARTUP` env var
- [ ] Document how to use in README/docs
- [ ] Output to stderr (doesn't interfere with stdout)

---

## Section 6: Build & Test Workflow

### Build & Benchmark Script

```bash
#!/bin/bash
# scripts/benchmark-startup.sh

set -e

echo "Building optimized binary..."
./scripts/build.sh

echo ""
echo "Warming cache (5 runs)..."
for i in {1..5}; do
    ./lazynuget --version > /dev/null
done

echo ""
echo "Benchmarking with hyperfine (100 runs)..."

if command -v hyperfine &> /dev/null; then
    hyperfine \
        --warmup 3 \
        --runs 100 \
        --show-output \
        './lazynuget --version'
else
    echo "hyperfine not found. Installing..."
    # Installation varies by OS
    echo "Please install hyperfine: https://github.com/sharkdp/hyperfine"
    exit 1
fi
```

### CI/CD Integration

```yaml
# .github/workflows/performance.yml
name: Performance Check

on: [push, pull_request]

jobs:
  startup-performance:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Install hyperfine
        run: cargo install hyperfine || true

      - name: Benchmark startup
        run: |
          go build ./cmd/lazynuget
          for i in {1..5}; do
            ./lazynuget --version > /dev/null
          done
          hyperfine --runs 50 './lazynuget --version' \
            --export-json /tmp/bench.json

      - name: Check performance regression
        run: |
          # Parse JSON, check if p95 < 200ms
          python3 -c "
          import json
          with open('/tmp/bench.json') as f:
              data = json.load(f)
              mean_ms = data['results'][0]['mean'] * 1000
              if mean_ms > 180:
                  print(f'WARNING: Startup {mean_ms:.0f}ms (target: <200ms)')
              else:
                  print(f'✓ Startup {mean_ms:.0f}ms (target: <200ms)')
          "
```

### Pre-Commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Quick startup check before commit
echo "Running startup performance check..."

go build ./cmd/lazynuget
time ./lazynuget --version > /dev/null

echo "✓ Startup check passed"
```

### Checklist for Build/Test Setup

- [ ] Create `scripts/build.sh` with optimization flags
- [ ] Create `scripts/benchmark-startup.sh` for local testing
- [ ] Add performance checks to CI/CD pipeline
- [ ] Document baseline <200ms target
- [ ] Set up regression alerts (warn if >180ms)
- [ ] Create pre-commit hook for local verification

---

## Section 7: Measurement Validation Strategy

### Baseline Establishment

**Before any optimizations**:
1. Build with standard flags: `go build ./cmd/lazynuget`
2. Warm cache: Run binary 5 times
3. Measure: `hyperfine './lazynuget --version' --runs 100`
4. Record baseline time (mean)

### Iterative Optimization

For each optimization:
1. Implement change
2. Rebuild: `go build ./cmd/lazynuget`
3. Re-measure: same hyperfine command
4. Calculate: `(baseline - new) / baseline * 100` = % improvement
5. Document in commit message

### Final Validation

When all optimizations complete:
1. Verify p95 < 200ms (run 100+ iterations)
2. Verify memory < 10MB idle
3. Run on different hardware if possible
4. Test in CI/CD pipeline
5. Record final metrics in documentation

### Success Criteria Checklist

- [ ] `lazynuget --version` < 200ms p95
- [ ] `lazynuget --help` < 200ms p95
- [ ] TUI startup (full init) < 300ms p95
- [ ] Memory at idle < 10MB
- [ ] No startup errors in any environment
- [ ] Benchmarks pass in CI/CD

---

## Decision Record: Optimization Strategy

**Feature**: Application Bootstrap
**Requirement**: <200ms p95 cold start
**Approach**: Multi-strategy (lazy init + async validation + build optimization + parallelization)

### Technique Selection Rationale

| Technique | Why Chosen | Why Not Alternative |
|-----------|-----------|-------------------|
| Lazy GUI init | 80-120ms gain, architectural benefit | Always-init wastes time for --version |
| Async validation | Responsive startup, graceful fallback | Blocking validation adds 30-100ms delay |
| Build flags | 5-10ms gain, trivial to implement | Larger binary penalizes every startup |
| Parallel init | 10-20ms gain, idiomatic with errgroup | Sequential adds unnecessary latency |
| Custom instrumentation | Identifies bottlenecks, visibility | pprof overkill for monitoring startup |

### Total Expected Impact

```
Baseline (estimated):       ~250-300ms
├─ Lazy GUI init:           -80-120ms
├─ Async validation:        -30-50ms
├─ Build optimization:      -5-10ms
└─ Parallel init:           -10-20ms
─────────────────────────────────
Target:                      <200ms ✓
```

### Risk Assessment

| Technique | Risk | Mitigation |
|-----------|------|-----------|
| Lazy GUI init | Missing initialization step | Clear separation via sync.Once |
| Async validation | Undetected missing dotnet | Show status in UI, timeout handling |
| Build flags | Loss of debuggability | Keep symbols in debug builds |
| Parallel init | Race conditions | Clear dependency ordering |

**Overall Risk**: Very Low (patterns are standard, proven in production CLIs)

---

## Appendix: Code Template - Bootstrapper Interface

Ready-to-use interface definition:

```go
// internal/bootstrap/interfaces.go
package bootstrap

import "context"

// Bootstrapper manages application initialization and lifecycle
type Bootstrapper interface {
    // Initialize performs all startup initialization
    Initialize(ctx context.Context) error

    // Run starts the application (blocking)
    Run(ctx context.Context) error

    // Shutdown performs graceful shutdown
    Shutdown(ctx context.Context) error
}

// App implements Bootstrapper
type App struct {
    config Config
    logger Logger
    // ... other fields ...
}

var _ Bootstrapper = (*App)(nil)  // Compile-time check
```

---

## Next Steps: Phase 2 Implementation

Use this guide to implement:
1. Lazy GUI initialization (main architect decision)
2. Async dotnet validation (improves responsiveness)
3. Build optimization (trivial win)
4. Parallel init with errgroup (measurable improvement)
5. Startup instrumentation (enables validation)

**Expected Phase 2 Output**: Working bootstrap with <200ms verified startup time

---

**Document Version**: 1.0
**Created**: 2025-11-02
**Status**: Ready for Implementation
**Next Review**: After Phase 2 completion
