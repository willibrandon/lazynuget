// LAZY_INIT_EXAMPLES.go
// Ready-to-use code patterns for startup optimization
// Copy these patterns into your actual implementation (Phase 2)

package bootstrap_examples

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// ============================================================================
// PATTERN 1: Custom Startup Stopwatch for Instrumentation
// ============================================================================

// Stopwatch tracks startup timing at key phases
type Stopwatch struct {
	Start   time.Time
	Markers []TimingMarker
}

// TimingMarker records a named point in time
type TimingMarker struct {
	Name string
	Time time.Time
}

// NewStopwatch creates a new stopwatch starting now
func NewStopwatch() *Stopwatch {
	return &Stopwatch{Start: time.Now()}
}

// Mark records a timing marker with current timestamp
func (s *Stopwatch) Mark(name string) {
	s.Markers = append(s.Markers, TimingMarker{
		Name: name,
		Time: time.Now(),
	})
}

// Report returns a formatted report of startup timing
func (s *Stopwatch) Report() string {
	var buf strings.Builder
	buf.WriteString("Startup Timeline:\n")

	prev := s.Start
	for i, m := range s.Markers {
		elapsed := m.Time.Sub(prev).Milliseconds()
		total := m.Time.Sub(s.Start).Milliseconds()

		if i == 0 {
			buf.WriteString(fmt.Sprintf("  %s: %dms\n", m.Name, total))
		} else {
			buf.WriteString(fmt.Sprintf("  %s: +%dms (total: %dms)\n", m.Name, elapsed, total))
		}
		prev = m.Time
	}

	return buf.String()
}

// Usage in main.go:
/*
func main() {
    sw := NewStopwatch()
    defer func() {
        if os.Getenv("DEBUG_STARTUP") == "1" {
            fmt.Fprintf(os.Stderr, "\n%s", sw.Report())
        }
    }()

    // ... startup code with sw.Mark() calls ...
}
*/

// ============================================================================
// PATTERN 2: Lazy Initialization with sync.Once
// ============================================================================

// Config represents application configuration
type Config struct {
	Version string
	// ... other fields ...
}

// Logger represents application logger
type Logger struct {
	// ... fields ...
}

// GUI represents the TUI application
type GUI struct {
	// ... fields ...
}

// AppServices holds application services with lazy initialization
type AppServices struct {
	// Eagerly initialized (fast path)
	config Config
	logger Logger

	// Lazily initialized (deferred until first access)
	guiOnce sync.Once
	gui     *GUI
	guiErr  error
}

// GetOrInitGUI returns the GUI, initializing on first call
// This is thread-safe and zero-cost after first call
func (a *AppServices) GetOrInitGUI() (*GUI, error) {
	a.guiOnce.Do(func() {
		// This block runs exactly once, on first call
		// Subsequent calls return immediately
		a.gui, a.guiErr = newGUI(a.config, a.logger)
	})
	return a.gui, a.guiErr
}

// Helper to create GUI (would be in actual gui package)
func newGUI(config Config, logger Logger) (*GUI, error) {
	// Simulate expensive GUI initialization (80-120ms)
	// time.Sleep(100 * time.Millisecond)
	return &GUI{}, nil
}

// Usage example:
/*
app := &AppServices{
    config: cfg,
    logger: log,
}

// First call initializes (80-120ms spent here)
gui, err := app.GetOrInitGUI()

// Subsequent calls return immediately (zero cost)
gui2, _ := app.GetOrInitGUI()  // Returns same gui, no initialization
*/

// ============================================================================
// PATTERN 3: Async Validation with Background Goroutine
// ============================================================================

// DotnetValidator manages async dotnet CLI validation
type DotnetValidator struct {
	// State
	mu              sync.Mutex
	validated       bool
	err             error
	validationTime  time.Time
}

// StartAsync begins validation in a background goroutine
// Returns immediately without blocking
func (v *DotnetValidator) StartAsync(ctx context.Context) {
	go v.validateInBackground(ctx)
}

// validateInBackground runs the actual validation
func (v *DotnetValidator) validateInBackground(ctx context.Context) {
	// Simulate dotnet validation (30-100ms)
	// result := checkDotnetCLI()
	result := true // Success

	v.mu.Lock()
	defer v.mu.Unlock()

	if !result {
		v.err = fmt.Errorf("dotnet CLI not found")
	}
	v.validated = true
	v.validationTime = time.Now()
}

// GetStatus returns validation result, waiting if needed
// Returns nil if validated, error if validation failed or timed out
func (v *DotnetValidator) GetStatus(ctx context.Context) error {
	// Fast path: already validated, return cached result
	v.mu.Lock()
	if v.validated {
		err := v.err
		v.mu.Unlock()
		return err
	}
	v.mu.Unlock()

	// Slow path: still validating, wait with timeout
	start := time.Now()
	timeout := 3 * time.Second

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		v.mu.Lock()
		if v.validated {
			err := v.err
			v.mu.Unlock()
			return err
		}
		v.mu.Unlock()

		if time.Since(start) > timeout {
			// Validation taking too long, don't block
			return nil
		}

		time.Sleep(10 * time.Millisecond)
	}
}

// Usage example:
/*
validator := &DotnetValidator{}

// Start validation, return immediately
validator.StartAsync(ctx)

// ... continue startup, show UI spinner ...

// Later: check if dotnet is available
if err := validator.GetStatus(ctx); err != nil {
    // dotnet validation failed, handle gracefully
}
*/

// ============================================================================
// PATTERN 4: Parallel Initialization with errgroup
// ============================================================================

// InitResult holds results from parallel initialization
type InitResult struct {
	Config   Config
	Logger   Logger
	Platform PlatformInfo
}

// PlatformInfo holds platform detection results
type PlatformInfo struct {
	OS   string
	Arch string
}

// InitializeServicesParallel initializes independent services in parallel
// Saves 10-20ms by parallelizing independent tasks
func InitializeServicesParallel(ctx context.Context) (*InitResult, error) {
	g, ctx := errgroup.WithContext(ctx)

	var config Config
	var logger Logger
	var platform PlatformInfo

	// Task 1: Load configuration
	g.Go(func() error {
		var err error
		config, err = loadConfig() // 10ms
		return err
	})

	// Task 2: Initialize logger (independent of others)
	g.Go(func() error {
		var err error
		logger, err = initLogger() // 5ms
		return err
	})

	// Task 3: Detect platform (independent of others)
	g.Go(func() error {
		var err error
		platform, err = detectPlatform() // 5ms
		return err
	})

	// Wait for all tasks to complete
	// First error causes remaining tasks to be cancelled
	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("parallel initialization failed: %w", err)
	}

	return &InitResult{
		Config:   config,
		Logger:   logger,
		Platform: platform,
	}, nil
}

// Helper functions (would be in actual packages)
func loadConfig() (Config, error) {
	// time.Sleep(10 * time.Millisecond)
	return Config{Version: "1.0.0"}, nil
}

func initLogger() (Logger, error) {
	// time.Sleep(5 * time.Millisecond)
	return Logger{}, nil
}

func detectPlatform() (PlatformInfo, error) {
	// time.Sleep(5 * time.Millisecond)
	return PlatformInfo{OS: "linux", Arch: "amd64"}, nil
}

// Usage example:
/*
result, err := InitializeServicesParallel(ctx)
if err != nil {
    return err
}

app := &AppServices{
    config: result.Config,
    logger: result.Logger,
}
*/

// ============================================================================
// PATTERN 5: Complete Bootstrap Example (Putting it Together)
// ============================================================================

// CompleteBootstrapExample shows all patterns together
type BootstrapApp struct {
	// Fast path services (initialized immediately)
	config   Config
	logger   Logger
	platform PlatformInfo

	// Slow path services (deferred)
	guiOnce sync.Once
	gui     *GUI
	guiErr  error

	// Async operations
	dotnetValidator *DotnetValidator

	// Instrumentation
	stopwatch *Stopwatch
}

// NewBootstrapApp creates a new application
func NewBootstrapApp() *BootstrapApp {
	return &BootstrapApp{
		dotnetValidator: &DotnetValidator{},
		stopwatch:       NewStopwatch(),
	}
}

// Initialize performs all startup initialization
func (b *BootstrapApp) Initialize(ctx context.Context) error {
	// Phase 1: Parse CLI flags (not shown, but must happen first)
	b.stopwatch.Mark("flags_parsed")

	// Phase 2: Parallel initialization of independent services
	// This takes ~10-15ms instead of 20ms sequential
	result, err := InitializeServicesParallel(ctx)
	if err != nil {
		return err
	}
	b.config = result.Config
	b.logger = result.Logger
	b.platform = result.Platform
	b.stopwatch.Mark("services_initialized")

	// Phase 3: Start async validation (doesn't block)
	b.dotnetValidator.StartAsync(ctx)
	b.stopwatch.Mark("validation_started")

	// Phase 4: Return immediately (GUI init deferred)
	b.stopwatch.Mark("bootstrap_complete")

	return nil
}

// Run starts the application (blocking)
func (b *BootstrapApp) Run(ctx context.Context) error {
	// Check for early-exit flags (--version, --help)
	if b.config.Version == "show" {
		fmt.Println("LazyNuGet v1.0.0")
		return nil
	}

	// Now initialize GUI (this is where 80-120ms is spent)
	_, err := b.GetOrInitGUI()
	if err != nil {
		return err
	}
	b.stopwatch.Mark("gui_initialized")

	// Run the GUI event loop
	// In actual implementation: gui.Run()

	return nil
}

// GetOrInitGUI returns GUI, initializing lazily
func (b *BootstrapApp) GetOrInitGUI() (*GUI, error) {
	b.guiOnce.Do(func() {
		b.gui, b.guiErr = newGUI(b.config, b.logger)
	})
	return b.gui, b.guiErr
}

// WaitForValidation waits for async dotnet validation
func (b *BootstrapApp) WaitForValidation(ctx context.Context) error {
	return b.dotnetValidator.GetStatus(ctx)
}

// ReportTiming outputs startup timing if DEBUG_STARTUP is set
func (b *BootstrapApp) ReportTiming() {
	if os.Getenv("DEBUG_STARTUP") == "1" {
		fmt.Fprintf(os.Stderr, "\n%s", b.stopwatch.Report())
	}
}

// Usage in main.go:
/*
func main() {
    app := NewBootstrapApp()
    defer app.ReportTiming()

    ctx := context.Background()

    if err := app.Initialize(ctx); err != nil {
        fmt.Fprintf(os.Stderr, "initialization failed: %v\n", err)
        os.Exit(1)
    }

    if err := app.Run(ctx); err != nil {
        fmt.Fprintf(os.Stderr, "runtime error: %v\n", err)
        os.Exit(1)
    }
}

// Run with: DEBUG_STARTUP=1 ./lazynuget
*/

// ============================================================================
// PATTERN 6: Build Optimization Script
// ============================================================================

// Build script content (save as scripts/build.sh):
/*
#!/bin/bash

VERSION=${VERSION:-dev}
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%SZ')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

go build \
  -o lazynuget \
  -ldflags="-w -s -X 'main.Version=${VERSION}' -X 'main.BuildTime=${BUILD_TIME}' -X 'main.GitCommit=${GIT_COMMIT}'" \
  -trimpath \
  ./cmd/lazynuget

echo "Built lazynuget (${VERSION})"
ls -lh lazynuget
*/

// ============================================================================
// PATTERN 7: Benchmarking Script
// ============================================================================

// Benchmark script content (save as scripts/benchmark.sh):
/*
#!/bin/bash

set -e

echo "Building optimized binary..."
./scripts/build.sh

echo ""
echo "Warming cache (5 runs)..."
for i in {1..5}; do
    ./lazynuget --version > /dev/null
done

echo ""
echo "Benchmarking with hyperfine..."
if command -v hyperfine &> /dev/null; then
    hyperfine \
        --warmup 3 \
        --runs 100 \
        --show-output \
        './lazynuget --version'
else
    echo "hyperfine not found. Install from: https://github.com/sharkdp/hyperfine"
    exit 1
fi
*/

// ============================================================================
// PATTERN 8: Verification Checklist
// ============================================================================

/*
IMPLEMENTATION CHECKLIST:

Fast Path (must complete <50ms):
  ☐ Flag parsing
  ☐ Config loading
  ☐ Logger initialization
  ☐ Platform detection
  ☐ Parallel initialization with errgroup
  ☐ Return for --version/--help without GUI init

Slow Path (deferred, doesn't block startup):
  ☐ GUI initialization (Bubbletea Program)
  ☐ Dotnet CLI validation (async)
  ☐ Project discovery (lazy load)
  ☐ Package metadata caching (lazy load)

Instrumentation:
  ☐ Stopwatch struct with Mark() method
  ☐ Timing report function
  ☐ DEBUG_STARTUP environment variable check
  ☐ Output to stderr

Build Optimization:
  ☐ -w flag (strip DWARF symbols)
  ☐ -s flag (strip symbol table)
  ☐ -ldflags with version injection
  ☐ -trimpath flag

Measurement:
  ☐ Install hyperfine for benchmarking
  ☐ Baseline measurement before optimization
  ☐ Per-optimization measurement
  ☐ Final p95 validation (<200ms)

Testing:
  ☐ Test --version completes in <50ms
  ☐ Test --help completes in <50ms
  ☐ Test TUI startup in <200ms (p95)
  ☐ Test memory usage < 10MB idle
  ☐ Test shutdown < 1 second
  ☐ Cross-platform testing (Windows, macOS, Linux)

CI/CD Integration:
  ☐ Add build script to build process
  ☐ Add performance benchmarking to pipeline
  ☐ Set regression alert threshold (>180ms warning)
  ☐ Track metrics over releases
*/

// ============================================================================
// NOTES FOR IMPLEMENTATION
// ============================================================================

/*
KEY DECISIONS:

1. Lazy GUI Initialization
   - Most impactful optimization (80-120ms saved)
   - Use sync.Once for thread-safe, zero-cost deferral
   - Clear separation: config→validation→GUI init→run

2. Async Validation
   - Start dotnet check in background
   - Don't block startup waiting for result
   - Show status spinner while checking
   - Graceful timeout at 3 seconds

3. Parallel Services
   - Use errgroup for clean error handling
   - Only parallelize independent tasks
   - Maintain clear dependency ordering
   - Simple to understand and maintain

4. Instrumentation
   - Conditional on DEBUG_STARTUP env var
   - Zero cost when disabled
   - Output to stderr to avoid stdout pollution
   - Use for iteration and validation

5. Build Optimization
   - Strip symbols for 30-35% size reduction
   - Inject version at build time (no init() needed)
   - Use -trimpath for reproducible builds
   - Binary size directly impacts startup time

MEASUREMENT APPROACH:

1. Establish baseline before optimizations
2. Measure each optimization independently
3. Combine optimizations and re-measure
4. Validate p95 < 200ms with 100+ runs
5. Check regressions in CI/CD pipeline

Expected Improvement:
  Baseline: ~250-300ms
  With optimizations: ~100-150ms
  TUI startup (with deferred GUI init): ~150-180ms (well under 200ms target)
*/
