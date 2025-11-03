# TTY Detection Implementation Guide for LazyNuGet

## Quick Start

Copy this single function into your codebase:

```go
package interactive

import (
	"os"
	"golang.org/x/term"
)

// IsInteractive detects whether LazyNuGet should run in interactive mode.
// Auto-detects TTY and respects CI/CD environment flags.
func IsInteractive() bool {
	// Force non-interactive in CI environments
	if os.Getenv("CI") != "" {
		return false
	}

	// Check if stdin is connected to a terminal
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return false
	}

	// Respect NO_COLOR convention
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Dumb terminals can't handle interactive features
	if os.Getenv("TERM") == "dumb" {
		return false
	}

	return true
}
```

---

## Complete Production-Ready Implementation

### File: `internal/interactive/detector.go`

```go
package interactive

import (
	"os"
	"golang.org/x/term"
)

// Detector provides TTY detection capability.
// This interface allows mocking in tests.
type Detector interface {
	IsTerminal(fd int) bool
}

// realDetector implements Detector using golang.org/x/term
type realDetector struct{}

func (d *realDetector) IsTerminal(fd int) bool {
	return term.IsTerminal(fd)
}

var detector Detector = &realDetector{}

// SetDetector replaces the detector with a custom implementation.
// Useful for testing. Should only be called from tests.
func SetDetector(d Detector) {
	detector = d
}

// IsInteractive determines if LazyNuGet should run in interactive mode.
//
// Returns false in these scenarios:
// - CI environment variable is set
// - Standard input is not a terminal (piped or redirected)
// - NO_COLOR environment variable is set
// - TERM environment variable is "dumb"
//
// This handles:
// - GitHub Actions, GitLab CI, CircleCI, etc. (CI=true)
// - Piped input: echo "input" | lazynuget
// - File redirect: lazynuget < input.txt
// - Docker without -it flags
// - SSH without pseudo-terminal allocation
// - Screen readers and accessibility tools
//
// Returns true for:
// - Interactive terminal (bash, zsh, PowerShell, etc.)
// - SSH with pseudo-terminal (-t flag)
// - Docker with -it flags
// - Terminal multiplexers (tmux, screen)
func IsInteractive() bool {
	// 1. CI environments always run non-interactive
	if os.Getenv("CI") != "" {
		return false
	}

	// 2. If stdin is not a terminal, don't run interactively
	if !detector.IsTerminal(int(os.Stdin.Fd())) {
		return false
	}

	// 3. Respect user's NO_COLOR preference
	// Users who set NO_COLOR typically also want non-interactive output
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// 4. Dumb terminals have no special terminal features
	if os.Getenv("TERM") == "dumb" {
		return false
	}

	return true
}

// IsColorEnabled determines if color output should be used.
// Separate from IsInteractive() as some non-interactive terminals can use colors.
func IsColorEnabled() bool {
	// NO_COLOR takes precedence
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Dumb terminals don't support colors
	if os.Getenv("TERM") == "dumb" {
		return false
	}

	// Stdout needs to be a terminal for colors to work
	if !detector.IsTerminal(int(os.Stdout.Fd())) {
		return false
	}

	return true
}
```

### File: `internal/interactive/detector_test.go`

```go
package interactive

import (
	"os"
	"testing"
)

// mockDetector simulates TTY detection for testing
type mockDetector struct {
	isTerminal bool
}

func (m *mockDetector) IsTerminal(fd int) bool {
	return m.isTerminal
}

func TestIsInteractive_WithTTY(t *testing.T) {
	oldDetector := detector
	detector = &mockDetector{isTerminal: true}
	defer func() { detector = oldDetector }()

	// Clear environment variables
	t.Setenv("CI", "")
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")

	if !IsInteractive() {
		t.Error("expected interactive=true with TTY and no env vars set")
	}
}

func TestIsInteractive_WithCI(t *testing.T) {
	oldDetector := detector
	detector = &mockDetector{isTerminal: true}
	defer func() { detector = oldDetector }()

	// Even with TTY, CI env var forces non-interactive
	t.Setenv("CI", "true")
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")

	if IsInteractive() {
		t.Error("expected interactive=false when CI=true")
	}
}

func TestIsInteractive_WithNOCOLOR(t *testing.T) {
	oldDetector := detector
	detector = &mockDetector{isTerminal: true}
	defer func() { detector = oldDetector }()

	t.Setenv("CI", "")
	t.Setenv("NO_COLOR", "1")
	t.Setenv("TERM", "xterm-256color")

	if IsInteractive() {
		t.Error("expected interactive=false when NO_COLOR is set")
	}
}

func TestIsInteractive_WithDumbTerminal(t *testing.T) {
	oldDetector := detector
	detector = &mockDetector{isTerminal: true}
	defer func() { detector = oldDetector }()

	t.Setenv("CI", "")
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "dumb")

	if IsInteractive() {
		t.Error("expected interactive=false when TERM=dumb")
	}
}

func TestIsInteractive_WithoutTTY(t *testing.T) {
	oldDetector := detector
	detector = &mockDetector{isTerminal: false}
	defer func() { detector = oldDetector }()

	t.Setenv("CI", "")
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")

	if IsInteractive() {
		t.Error("expected interactive=false when not a TTY")
	}
}

func TestIsColorEnabled_WithNOCOLOR(t *testing.T) {
	oldDetector := detector
	detector = &mockDetector{isTerminal: true}
	defer func() { detector = oldDetector }()

	t.Setenv("NO_COLOR", "1")
	t.Setenv("TERM", "xterm-256color")

	if IsColorEnabled() {
		t.Error("expected colors disabled when NO_COLOR is set")
	}
}

func TestIsColorEnabled_WithDumbTerminal(t *testing.T) {
	oldDetector := detector
	detector = &mockDetector{isTerminal: true}
	defer func() { detector = oldDetector }()

	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "dumb")

	if IsColorEnabled() {
		t.Error("expected colors disabled when TERM=dumb")
	}
}

func TestIsColorEnabled_WithoutTTY(t *testing.T) {
	oldDetector := detector
	detector = &mockDetector{isTerminal: false}
	defer func() { detector = oldDetector }()

	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")

	if IsColorEnabled() {
		t.Error("expected colors disabled when not a TTY")
	}
}

func TestIsColorEnabled_AllConditionsMet(t *testing.T) {
	oldDetector := detector
	detector = &mockDetector{isTerminal: true}
	defer func() { detector = oldDetector }()

	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")

	if !IsColorEnabled() {
		t.Error("expected colors enabled when TTY and good TERM")
	}
}
```

---

## Integration Examples

### Example 1: In Main Function

```go
package main

import (
	"context"
	"fmt"
	"os"

	"lazynuget/internal/interactive"
	"lazynuget/internal/restore"
)

func main() {
	ctx := context.Background()

	// Auto-detect interactive mode
	isInteractive := interactive.IsInteractive()

	// Log detection result (always shown in verbose mode)
	if os.Getenv("LAZYNUGET_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "Interactive mode: %v\n", isInteractive)
	}

	// Pass to restore operation
	opts := &restore.Options{
		NonInteractive: !isInteractive,
		// Other options...
	}

	if err := restore.Run(ctx, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

### Example 2: Output Formatting

```go
package output

import (
	"fmt"
	"os"

	"lazynuget/internal/interactive"
)

// Formatter provides different output based on mode
type Formatter struct {
	interactive bool
	colors      bool
}

// NewFormatter creates a formatter based on current environment
func NewFormatter() *Formatter {
	return &Formatter{
		interactive: interactive.IsInteractive(),
		colors:      interactive.IsColorEnabled(),
	}
}

// Printf prints a message, with colors if enabled
func (f *Formatter) Printf(format string, args ...interface{}) {
	if f.colors {
		// Use ANSI color codes
		fmt.Printf(format, args...)
	} else {
		// Plain text, no ANSI codes
		fmt.Printf(format, args...)
	}
}

// ShowProgress displays progress indicator (only in interactive mode)
func (f *Formatter) ShowProgress(message string) {
	if f.interactive {
		// Use spinner, live updates, etc.
		fmt.Printf("%s...\r", message)
	} else {
		// Use static line
		fmt.Printf("%s\n", message)
	}
}
```

### Example 3: Conditional Feature Behavior

```go
package restore

import (
	"lazynuget/internal/interactive"
)

// Options for restore operation
type Options struct {
	Verbosity      string
	NonInteractive bool
}

// SetVerbosityDefault sets verbosity based on interactive mode
func (o *Options) SetVerbosityDefault() {
	if o.Verbosity != "" {
		return // Already set by user
	}

	isInteractive := interactive.IsInteractive()

	if isInteractive {
		// Show progress, status updates
		o.Verbosity = "normal"
	} else {
		// Show structured output for CI parsing
		o.Verbosity = "detailed"
		o.NonInteractive = true
	}
}
```

---

## Testing Different Scenarios

### Unit Test Examples

```bash
# Run tests with verbose output
go test -v ./internal/interactive

# Test with coverage
go test -cover ./internal/interactive
```

### Integration Test Script

```bash
#!/bin/bash
set -e

PROJECT_DIR="test_project"
PROJ_FILE="$PROJECT_DIR/test.csproj"

# Setup
mkdir -p "$PROJECT_DIR"
cat > "$PROJ_FILE" << 'EOF'
<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <TargetFramework>net8.0</TargetFramework>
  </PropertyGroup>
  <ItemGroup>
    <PackageReference Include="Newtonsoft.Json" Version="13.0.1" />
  </ItemGroup>
</Project>
EOF

echo "=== Test 1: Direct Terminal (Interactive) ==="
lazynuget restore "$PROJ_FILE" 2>&1 | head -5

echo ""
echo "=== Test 2: Piped Input (Non-Interactive) ==="
echo "" | lazynuget restore "$PROJ_FILE" 2>&1 | head -5

echo ""
echo "=== Test 3: CI Environment (Non-Interactive) ==="
CI=true lazynuget restore "$PROJ_FILE" 2>&1 | head -5

echo ""
echo "=== Test 4: NO_COLOR Set (Non-Interactive) ==="
NO_COLOR=1 lazynuget restore "$PROJ_FILE" 2>&1 | head -5

echo ""
echo "=== Test 5: TERM=dumb (Non-Interactive) ==="
TERM=dumb lazynuget restore "$PROJ_FILE" 2>&1 | head -5

# Cleanup
rm -rf "$PROJECT_DIR"

echo ""
echo "All integration tests passed!"
```

---

## Debugging

### Enable Debug Output

```bash
# See detection result
LAZYNUGET_DEBUG=1 lazynuget restore project.csproj
```

### Add to Code

```go
if os.Getenv("LAZYNUGET_DEBUG") != "" {
	fmt.Fprintf(os.Stderr, "TTY Detection Debug:\n")
	fmt.Fprintf(os.Stderr, "  IsInteractive: %v\n", interactive.IsInteractive())
	fmt.Fprintf(os.Stderr, "  IsColorEnabled: %v\n", interactive.IsColorEnabled())
	fmt.Fprintf(os.Stderr, "  stdin fd: %d\n", os.Stdin.Fd())
	fmt.Fprintf(os.Stderr, "  CI: %q\n", os.Getenv("CI"))
	fmt.Fprintf(os.Stderr, "  NO_COLOR: %q\n", os.Getenv("NO_COLOR"))
	fmt.Fprintf(os.Stderr, "  TERM: %q\n", os.Getenv("TERM"))
}
```

---

## Summary Checklist

- [ ] Add `golang.org/x/term` import
- [ ] Copy `detector.go` to `internal/interactive/`
- [ ] Copy `detector_test.go` to `internal/interactive/`
- [ ] Run tests: `go test ./internal/interactive`
- [ ] Integrate `IsInteractive()` into main entry point
- [ ] Update output formatting to check `IsInteractive()`
- [ ] Test with different scenarios (terminal, pipe, CI, Docker)
- [ ] Document in `README.md` how to force interactive/non-interactive mode

