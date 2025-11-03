# Developer Quickstart: Application Bootstrap

**Feature**: Application Bootstrap and Lifecycle Management
**Branch**: 001-app-bootstrap
**Date**: 2025-11-02

This guide helps developers build, test, and debug the LazyNuGet bootstrap system. It covers local development setup, testing approaches, performance measurement, and troubleshooting.

---

## Prerequisites

### Required Tools

- **Go 1.24+**: [Download](https://go.dev/dl/)
- **.NET SDK 6.0+**: [Download](https://dotnet.microsoft.com/download) (for runtime validation)
- **Git**: For version control
- **Make**: (Optional) For build automation

### Verify Installation

```bash
# Check Go version
go version  # Should show 1.24 or higher

# Check .NET SDK
dotnet --version  # Should show 6.0 or higher

# Check Git
git --version
```

---

## Project Setup

### Clone Repository

```bash
git clone https://github.com/yourusername/lazynuget.git
cd lazynuget
```

### Switch to Feature Branch

```bash
git checkout 001-app-bootstrap
```

### Install Dependencies

```bash
# Download all Go modules
go mod download

# Verify dependencies
go mod verify
```

---

## Building LazyNuGet

### Development Build (Fast)

```bash
# Build without optimizations for quick iteration
go build -o lazynuget cmd/lazynuget/main.go
```

### Production Build (Optimized)

```bash
# Build with size optimizations and version info
VERSION="1.0.0"
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

go build \
  -ldflags="-w -s -X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE" \
  -trimpath \
  -o lazynuget \
  cmd/lazynuget/main.go
```

**Build Flags Explained**:
- `-w`: Omit DWARF symbol table (reduces binary size)
- `-s`: Omit symbol table and debug info (reduces binary size)
- `-X main.version=$VERSION`: Inject version string
- `-trimpath`: Remove build path from binary (reproducible builds)

**Expected Binary Size**: ~8-12 MB (depending on platform)

### Using Makefile (Optional)

Create `Makefile` in repository root:

```makefile
.PHONY: build build-dev clean test test-int

# Development build
build-dev:
	go build -o lazynuget cmd/lazynuget/main.go

# Production build with version injection
build:
	@VERSION=$$(git describe --tags --always --dirty) && \
	COMMIT=$$(git rev-parse --short HEAD) && \
	DATE=$$(date -u +"%Y-%m-%dT%H:%M:%SZ") && \
	go build \
		-ldflags="-w -s -X main.version=$$VERSION -X main.commit=$$COMMIT -X main.date=$$DATE" \
		-trimpath \
		-o lazynuget \
		cmd/lazynuget/main.go

# Clean build artifacts
clean:
	rm -f lazynuget
	go clean -cache -testcache

# Run unit tests
test:
	go test -v -race ./internal/bootstrap/...

# Run integration tests
test-int:
	go test -v -race ./tests/integration/...
```

Usage:
```bash
make build-dev    # Quick dev build
make build        # Production build
make test         # Unit tests
make test-int     # Integration tests
```

---

## Running LazyNuGet

### Basic Usage

```bash
# Start interactive TUI
./lazynuget

# Show version information
./lazynuget --version

# Show help
./lazynuget --help
```

### With Custom Configuration

```bash
# Use custom config file
./lazynuget --config /path/to/config.yml

# Set log level
./lazynuget --log-level debug

# Non-interactive mode (for testing)
./lazynuget --non-interactive
```

### Environment Variables

```bash
# Set log level via environment
LAZYNUGET_LOG_LEVEL=debug ./lazynuget

# Force non-interactive mode
NO_TTY=1 ./lazynuget

# Custom config path
LAZYNUGET_CONFIG=/custom/path/config.yml ./lazynuget
```

---

## Testing

### Unit Tests

Unit tests are co-located with source files in `internal/bootstrap/`.

```bash
# Run all bootstrap unit tests
go test ./internal/bootstrap/...

# Run with verbose output
go test -v ./internal/bootstrap/...

# Run with race detector
go test -race ./internal/bootstrap/...

# Run specific test
go test -v ./internal/bootstrap/ -run TestNewApp

# Run with coverage
go test -cover ./internal/bootstrap/...

# Generate coverage report
go test -coverprofile=coverage.out ./internal/bootstrap/...
go tool cover -html=coverage.out
```

### Integration Tests

Integration tests are in `tests/integration/` and test full startup/shutdown cycles.

```bash
# Run all integration tests
go test ./tests/integration/...

# Run specific integration test
go test -v ./tests/integration/ -run TestBootstrapCycle

# Run with timeout (important for signal tests)
go test -timeout 30s ./tests/integration/...
```

### Test Categories

**Unit Tests** (`internal/bootstrap/*_test.go`):
- `TestNewApp`: Application context construction
- `TestBootstrap`: Subsystem initialization
- `TestShutdown`: Cleanup handlers
- `TestLazyGUI`: GUI lazy initialization
- `TestConfigPrecedence`: CLI > Env > File > Default
- `TestValidation`: Config validation rules

**Integration Tests** (`tests/integration/*_test.go`):
- `TestBootstrapCycle`: Full startup → shutdown
- `TestSignalHandling`: SIGINT/SIGTERM processing
- `TestNonInteractiveMode`: TTY detection and behavior
- `TestPerformance`: Startup time validation (<200ms)
- `TestResourceLeaks`: 1000 iteration leak check
- `TestPanicRecovery`: Panic handling at all layers

### Test Fixtures

Configuration fixtures are in `tests/fixtures/configs/`:

```yaml
# tests/fixtures/configs/valid.yml
logLevel: debug
theme: default
shutdownTimeout: 3s
```

```yaml
# tests/fixtures/configs/invalid.yml
logLevel: invalid_level  # Triggers validation error
shutdownTimeout: 100s    # Exceeds maximum
```

### Table-Driven Test Example

```go
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Configuration
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Configuration{
				LogLevel: "info",
				ShutdownTimeout: 3 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "invalid log level",
			config: Configuration{
				LogLevel: "invalid",
				ShutdownTimeout: 3 * time.Second,
			},
			wantErr: true,
			errMsg:  "logLevel must be one of",
		},
		// More test cases...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error message %q does not contain %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
```

---

## Performance Measurement

### Startup Time Measurement

The bootstrap system includes instrumentation for measuring startup performance.

#### Enable Debug Mode

```bash
# Enable startup timing instrumentation
DEBUG_STARTUP=1 ./lazynuget --version
```

**Sample Output**:
```
init: +5ms
config: +25ms (cumulative: 30ms)
logging: +8ms (cumulative: 38ms)
platform: +18ms (cumulative: 56ms)
dependencies: +12ms (cumulative: 68ms)
signals: +2ms (cumulative: 70ms)
ready: +0ms (cumulative: 70ms)

Total startup: 70ms ✅ (target: <200ms)
```

#### Using Hyperfine (Recommended)

[Hyperfine](https://github.com/sharkdp/hyperfine) is a command-line benchmarking tool.

```bash
# Install hyperfine
brew install hyperfine  # macOS
# or
cargo install hyperfine  # Any platform with Rust

# Benchmark startup time
hyperfine './lazynuget --version' --runs 100

# Compare optimized vs unoptimized builds
hyperfine './lazynuget-dev --version' './lazynuget-prod --version'
```

**Sample Output**:
```
Benchmark 1: ./lazynuget --version
  Time (mean ± σ):      67.3 ms ±   4.2 ms    [User: 45.1 ms, System: 18.2 ms]
  Range (min … max):    62.1 ms …  78.4 ms    100 runs
```

**Target**: p95 < 200ms (from SC-001)

#### Manual Measurement Script

Create `scripts/dev/test-startup.sh`:

```bash
#!/bin/bash
set -e

BINARY="${1:-./lazynuget}"
RUNS="${2:-100}"

echo "Testing startup performance: $RUNS runs"
echo "Binary: $BINARY"
echo ""

total=0
min=999999
max=0

for i in $(seq 1 $RUNS); do
    start=$(date +%s%N)
    $BINARY --version > /dev/null 2>&1
    end=$(date +%s%N)

    elapsed=$(( (end - start) / 1000000 ))  # Convert to ms
    total=$((total + elapsed))

    if [ $elapsed -lt $min ]; then min=$elapsed; fi
    if [ $elapsed -gt $max ]; then max=$elapsed; fi

    if [ $((i % 10)) -eq 0 ]; then
        echo "Progress: $i/$RUNS runs"
    fi
done

mean=$((total / RUNS))
echo ""
echo "Results:"
echo "  Mean: ${mean}ms"
echo "  Min:  ${min}ms"
echo "  Max:  ${max}ms"

if [ $mean -lt 200 ]; then
    echo "✅ PASS: Mean startup time under 200ms target"
else
    echo "❌ FAIL: Mean startup time exceeds 200ms target"
    exit 1
fi
```

Usage:
```bash
chmod +x scripts/dev/test-startup.sh
./scripts/dev/test-startup.sh ./lazynuget 100
```

### Memory Profiling

```bash
# Run with memory profiling
go test -memprofile=mem.prof ./internal/bootstrap/

# View memory profile
go tool pprof mem.prof
# Commands in pprof: top, list, web
```

### CPU Profiling

```bash
# Run with CPU profiling
go test -cpuprofile=cpu.prof ./internal/bootstrap/

# View CPU profile
go tool pprof cpu.prof
```

### Continuous Performance Testing

Add to CI pipeline (`.github/workflows/performance.yml`):

```yaml
name: Performance Tests

on: [push, pull_request]

jobs:
  performance:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Build optimized binary
        run: make build

      - name: Install hyperfine
        run: |
          wget https://github.com/sharkdp/hyperfine/releases/download/v1.16.1/hyperfine_1.16.1_amd64.deb
          sudo dpkg -i hyperfine_1.16.1_amd64.deb

      - name: Benchmark startup time
        run: |
          hyperfine './lazynuget --version' --runs 100 --export-json results.json

      - name: Validate performance
        run: |
          # Parse JSON and check if mean < 200ms
          MEAN=$(jq '.results[0].mean * 1000' results.json)
          if (( $(echo "$MEAN < 200" | bc -l) )); then
            echo "✅ Performance target met: ${MEAN}ms"
          else
            echo "❌ Performance target missed: ${MEAN}ms"
            exit 1
          fi
```

---

## Manual Testing Procedures

### Test Startup

```bash
# Test basic startup
./lazynuget

# Verify:
# - Application starts within 200ms
# - TUI displays correctly
# - No error messages in logs
```

### Test Version Display

```bash
# Test version flag
./lazynuget --version

# Expected output:
# LazyNuGet version 1.0.0 (abc123) built on 2025-11-02T10:30:00Z

# Verify:
# - Version displayed immediately
# - No TUI launched
# - Process exits with code 0
```

### Test Help Display

```bash
# Test help flag
./lazynuget --help

# Expected output:
# Usage: lazynuget [options]
# Options:
#   --version           Show version information
#   --help              Show this help message
#   --config PATH       Custom config file path
#   --log-level LEVEL   Set log level (debug|info|warn|error)
#   --non-interactive   Run without TUI (for testing/CI)

# Verify:
# - Help text displayed
# - All flags documented
# - Process exits with code 0
```

### Test Graceful Shutdown

```bash
# Start application
./lazynuget

# Press 'q' to quit
# OR press Ctrl+C

# Verify:
# - Application exits within 1 second
# - No error messages
# - Terminal restored properly
# - Exit code 0
```

### Test Force Quit

```bash
# Start application
./lazynuget

# Press Ctrl+C once
# Wait 100ms
# Press Ctrl+C again (force quit)

# Verify:
# - First Ctrl+C: "Shutting down gracefully..."
# - Second Ctrl+C: Immediate exit
# - Exit code 130 (interrupted)
```

### Test Non-Interactive Mode

```bash
# Test explicit flag
./lazynuget --non-interactive --version

# Test TTY detection
echo "" | ./lazynuget --version  # Piped input, no TTY

# Test CI environment
CI=true ./lazynuget --version

# Verify:
# - No TUI launched
# - Output to stdout
# - Exit code 0
```

### Test Configuration Loading

```bash
# Test with custom config
./lazynuget --config tests/fixtures/configs/valid.yml --log-level debug

# Verify:
# - Config loaded from custom path
# - CLI flag (--log-level) overrides file setting
# - Debug messages appear in logs
```

### Test Error Handling

```bash
# Test invalid config
./lazynuget --config tests/fixtures/configs/invalid.yml

# Expected output:
# Error: Configuration validation failed: logLevel must be one of: debug, info, warn, error
# Exit code: 1

# Test missing config
./lazynuget --config /nonexistent/config.yml

# Expected output:
# Error: Config file not found: /nonexistent/config.yml
# Using default configuration...
# (Application continues with defaults)
```

---

## Debugging

### Enable Debug Logging

```bash
# Set log level to debug
./lazynuget --log-level debug

# Or via environment
LAZYNUGET_LOG_LEVEL=debug ./lazynuget
```

**Log Location**: `~/.config/lazynuget/logs/lazynuget.log` (Linux/macOS) or `%APPDATA%\lazynuget\logs\lazynuget.log` (Windows)

### View Logs

```bash
# Tail logs in real-time
tail -f ~/.config/lazynuget/logs/lazynuget.log

# Search for errors
grep ERROR ~/.config/lazynuget/logs/lazynuget.log

# View last 100 lines
tail -n 100 ~/.config/lazynuget/logs/lazynuget.log
```

### Debug Startup Issues

```bash
# Enable startup instrumentation
DEBUG_STARTUP=1 ./lazynuget

# Enable both startup and lifecycle debugging
DEBUG_STARTUP=1 DEBUG_LIFECYCLE=1 ./lazynuget
```

### Using Delve Debugger

```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Start debugging session
dlv debug cmd/lazynuget/main.go

# Set breakpoint
(dlv) break main.main
(dlv) break internal/bootstrap.NewApp

# Continue execution
(dlv) continue

# Inspect variables
(dlv) print app
(dlv) print config

# Step through code
(dlv) next
(dlv) step
```

### Common Issues

**Issue**: "dotnet CLI not found"
```bash
# Verify dotnet is in PATH
which dotnet  # Should show path to dotnet

# If missing, install .NET SDK
# https://dotnet.microsoft.com/download
```

**Issue**: "Permission denied creating config directory"
```bash
# Check directory permissions
ls -ld ~/.config

# Fix permissions
chmod 755 ~/.config
```

**Issue**: "Startup time exceeds 200ms"
```bash
# Profile startup with DEBUG_STARTUP
DEBUG_STARTUP=1 ./lazynuget --version

# Identify slow phase (should be output in ms per phase)
# Focus optimization on slowest phase
```

---

## Code Organization

### Directory Structure

```
internal/bootstrap/
├── app.go              # Application context, DI container
├── app_test.go         # Unit tests for app.go
├── lifecycle.go        # Lifecycle manager implementation
├── lifecycle_test.go   # Unit tests for lifecycle.go
├── flags.go            # CLI argument parsing
├── flags_test.go       # Unit tests for flags.go
├── signals.go          # Signal handler implementation
├── signals_test.go     # Unit tests for signals.go
├── version.go          # Version display
└── version_test.go     # Unit tests for version.go

tests/integration/
├── bootstrap_test.go      # Full startup/shutdown tests
├── signals_test.go        # Signal handling tests
└── noninteractive_test.go # TTY detection tests
```

### Key Files

**`cmd/lazynuget/main.go`**: Entry point
- Minimal - delegates to bootstrap
- Handles panic recovery (Layer 1)
- Sets exit codes

**`internal/bootstrap/app.go`**: Application context
- Dependency injection container
- Owns all subsystems (config, logger, platform, GUI)
- Lazy GUI initialization

**`internal/bootstrap/lifecycle.go`**: Lifecycle manager
- State machine (Uninitialized → Running → Stopped)
- Goroutine tracking
- Shutdown coordination

**`internal/bootstrap/signals.go`**: Signal handler
- Cross-platform signal handling
- Force-quit support
- Integration with context cancellation

---

## Performance Targets

From `spec.md` and `plan.md`:

| Metric | Target | Validation |
|--------|--------|------------|
| Cold start (TUI) | <200ms p95 | SC-001, hyperfine |
| Cold start (--version) | <50ms | Optimization Layer 1 |
| Normal shutdown | <1s | SC-002 |
| Forced shutdown | <3s | SC-003 |
| Memory (idle) | <10MB | SC-008 |
| Resource leaks | 0 over 1000 runs | SC-010 |

---

## Next Steps

After completing bootstrap implementation:

1. **Run full test suite**: `make test && make test-int`
2. **Validate performance**: `hyperfine './lazynuget --version' --runs 100`
3. **Check coverage**: `go test -cover ./internal/bootstrap/...` (target: >80%)
4. **Run static analysis**: `go vet ./...` and `staticcheck ./...`
5. **Proceed to Track 1, Spec 002**: Configuration Management

---

## Resources

- **LazyNuGet Constitution**: `.specify/memory/constitution.md`
- **Spec 001**: `specs/001-app-bootstrap/spec.md`
- **Research Findings**: `specs/001-app-bootstrap/research.md`
- **Data Model**: `specs/001-app-bootstrap/data-model.md`
- **API Contracts**: `specs/001-app-bootstrap/contracts/`
- **Go Documentation**: https://pkg.go.dev/
- **Bubbletea Docs**: https://github.com/charmbracelet/bubbletea

---

**Status**: Phase 1 Quickstart COMPLETE
