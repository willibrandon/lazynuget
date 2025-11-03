# TTY Detection Research for LazyNuGet

## Executive Summary

**Decision: Use `golang.org/x/term.IsTerminal()` with environment variable fallbacks**

This stdlib-based approach is the most reliable for cross-platform TTY detection without external dependencies. It provides native platform support for Windows console, Unix TTYs, SSH sessions, Docker containers, and CI/CD environments.

---

## 1. Primary Approach: golang.org/x/term.IsTerminal()

### Implementation
```go
import "golang.org/x/term"

func isInteractive() bool {
	// Check if stdin is a terminal
	return term.IsTerminal(int(os.Stdin.Fd()))
}
```

### Why It's Reliable

**Cross-Platform Foundation:**
- **Linux/macOS/Unix** (term_unix.go): Uses `unix.IoctlGetTermios()` IOCTL call
  - Checks if the file descriptor is connected to a real TTY
  - Works correctly with SSH sessions, tmux, screen

- **Windows** (term_windows.go): Uses `windows.GetConsoleMode()`
  - Detects Windows console handles directly
  - Returns `false` for pipes and file redirects
  - **Correctly handles:** Conhost.exe, Windows Terminal, PowerShell console
  - **Correctly rejects:** Git Bash (MSYS2), Cygwin pipes (unless they're actual terminals)

- **Plan 9**: Checks `/dev/cons` device
- **Unsupported platforms**: Conservatively returns `false`

### Platform-Specific Behavior

| Environment | IsTerminal Result | Reason |
|---|---|---|
| Interactive terminal | `true` | File descriptor has terminal attributes |
| SSH session | `true` | SSH pseudo-allocates TTY |
| GitHub Actions | `false` | No TTY allocated in runners |
| GitLab CI | `false` | No TTY allocated |
| Docker (interactive) | `true` | docker run -it allocates TTY |
| Docker (detached) | `false` | No TTY in detached mode |
| Piped output | `false` | IOCTL fails, not character device |
| File redirect | `false` | Pipe/file, not character device |
| Windows Terminal | `true` | GetConsoleMode succeeds |
| WSL + Windows Terminal | `true` | WSL maintains TTY through interop |
| Cygwin/MSYS2 with PTY | `true` | Windows console emulation works |

---

## 2. Supporting Environment Variables

### Recommended Checks (in order of priority)

```go
func isInteractive() bool {
	// 1. Check if stdin is a terminal (most reliable)
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return false
	}

	// 2. Respect explicit disables
	if os.Getenv("CI") != "" {
		// CI environment - force non-interactive
		return false
	}

	// 3. Respect NO_COLOR (implies non-interactive intent)
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// 4. Check TERM=dumb (ancient terminals with no capabilities)
	if os.Getenv("TERM") == "dumb" {
		return false
	}

	return true
}
```

### Environment Variables Explained

| Variable | Value | Meaning | Action |
|---|---|---|---|
| `CI` | any value | Running in CI/CD system | Force non-interactive |
| `NO_COLOR` | any value | User/system prefers no colors | Disable interactive (colors matter) |
| `TERM` | `dumb` | Dumb terminal (no capabilities) | Non-interactive mode |
| `TERM` | empty | Terminal type unknown | Could be problematic (from gonuget) |

### CI Environment Variables (detection only)

These **indicate** CI but `CI=true` is the standard:
- GitHub Actions: `GITHUB_ACTIONS=true`
- GitLab CI: `GITLAB_CI=true`
- CircleCI: `CIRCLECI=true`
- Travis CI: `TRAVIS=true`
- AppVeyor: `APPVEYOR=true`

**Recommendation:** Only check `CI` variable if `term.IsTerminal()` doesn't catch it, as a fallback.

---

## 3. Alternatives Considered

### Option A: github.com/mattn/go-isatty
**Pros:**
- Pure Syscall implementation, very reliable
- Supports Cygwin/MSYS2 detection (Windows-specific)
- No dependencies

**Cons:**
- Requires external dependency (github.com/mattn/go-isatty)
- Essentially duplicates stdlib functionality
- `fatih/color` already uses it transitively

**Verdict:** Unnecessary - stdlib is sufficient

### Option B: os.Stdin.Stat() with Mode Checking
```go
func isInteractive() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) != 0
}
```

**Pros:**
- Pure stdlib, no imports needed
- Works for basic cases

**Cons:**
- **Does NOT work reliably on all platforms**
- Windows: ModeCharDevice not set correctly for console
- May fail in some TTY multiplexers (tmux edge cases)
- Missing ioctls that validate terminal capabilities

**Verdict:** Avoid - less reliable than term.IsTerminal()

### Option C: Check Multiple File Descriptors
```go
func isInteractive() bool {
	return (term.IsTerminal(int(os.Stdin.Fd())) ||
		term.IsTerminal(int(os.Stdout.Fd())) ||
		term.IsTerminal(int(os.Stderr.Fd())))
}
```

**Pros:**
- Catches case where stdout is TTY but stdin is piped

**Cons:**
- Interactive shell usually needs both stdin AND stdout as TTY
- Different use cases (e.g., `echo "input" | lazynuget restore` should be non-interactive)
- Adds complexity

**Verdict:** Check stdin primarily, allow override via CI env vars

---

## 4. Real-World Code Examples

### From gonuget (Using os.Stat())
```go
// colors.go - isTerminal checks ModeCharDevice
func isTerminal(f *os.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// With environment fallbacks
func IsColorEnabled() bool {
	if !isTerminal(os.Stdout) {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	term := os.Getenv("TERM")
	if term == "dumb" || term == "" {
		return false
	}
	return true
}
```

### From bubbletea (Using term.IsTerminal)
```go
// tty.go - uses golang.org/x/term directly
if f, ok := p.input.(term.File); ok && term.IsTerminal(f.Fd()) {
	p.ttyInput = f
	p.previousTtyInputState, err = term.MakeRaw(p.ttyInput.Fd())
}
```

### From gonuget restore (Abstracted interface)
```go
// tty_detector.go - mockable interface for testing
type TTYDetector interface {
	IsTTY(w io.Writer) bool
	GetSize(w io.Writer) (width, height int, err error)
}

type RealTTYDetector struct{}

func (d *RealTTYDetector) IsTTY(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}
```

---

## 5. Recommended Implementation for LazyNuGet

### Basic Version (Recommended)
```go
package interactive

import (
	"os"
	"golang.org/x/term"
)

// IsInteractive detects if running in an interactive terminal.
// Returns false for:
// - CI/CD environments (CI=true)
// - Piped input/output
// - Containers without TTY allocation (docker, k8s)
// - SSH sessions without pseudo-TTY
// - Test environments
func IsInteractive() bool {
	// If running in CI, always non-interactive
	if os.Getenv("CI") != "" {
		return false
	}

	// Check if stdin is connected to a terminal
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return false
	}

	// Respect explicit user preferences
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

### Advanced Version (With Testability)
```go
package interactive

import (
	"io"
	"os"
	"golang.org/x/term"
)

// Detector provides TTY detection with mockable interface
type Detector interface {
	IsTerminal(f *os.File) bool
}

type realDetector struct{}

func (d *realDetector) IsTerminal(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}

var detector Detector = &realDetector{}

// SetDetector allows tests to inject mock detector
func SetDetector(d Detector) {
	detector = d
}

// IsInteractive detects interactive mode with testability
func IsInteractive() bool {
	if os.Getenv("CI") != "" {
		return false
	}

	if !detector.IsTerminal(os.Stdin) {
		return false
	}

	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	if os.Getenv("TERM") == "dumb" {
		return false
	}

	return true
}
```

### Usage in Main Program
```go
package main

import (
	"context"
	"lazynuget/internal/interactive"
)

func main() {
	ctx := context.Background()

	// Auto-detect interactive mode
	isInteractive := interactive.IsInteractive()

	// Pass to restore operation
	opts := &RestoreOptions{
		NonInteractive: !isInteractive,
		ProgressUI:     isInteractive, // Show progress bars only in interactive
		Verbosity:      getVerbosity(isInteractive),
	}

	if err := restore.Run(ctx, opts); err != nil {
		os.Exit(1)
	}
}

func getVerbosity(isInteractive bool) string {
	if isInteractive {
		return "normal" // Show progress and status
	}
	return "detailed" // Show structured logs for CI parsing
}
```

---

## 6. Testing Strategy

### Unit Tests
```go
func TestIsInteractive_WithTTY(t *testing.T) {
	// Mock detector that returns true
	SetDetector(&mockDetector{returnsTTY: true})
	defer SetDetector(&realDetector{})

	t.Setenv("CI", "")
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")

	if !IsInteractive() {
		t.Error("expected interactive=true with TTY")
	}
}

func TestIsInteractive_WithCI(t *testing.T) {
	SetDetector(&mockDetector{returnsTTY: true})
	defer SetDetector(&realDetector{})

	t.Setenv("CI", "true")

	if IsInteractive() {
		t.Error("expected interactive=false with CI=true")
	}
}

func TestIsInteractive_WithPipe(t *testing.T) {
	SetDetector(&mockDetector{returnsTTY: false})
	defer SetDetector(&realDetector{})

	if IsInteractive() {
		t.Error("expected interactive=false when piped")
	}
}
```

### Integration Tests (Real Environments)
```bash
#!/bin/bash

# Test 1: Direct terminal
go run ./cmd/lazynuget restore project.csproj
# Expected: Interactive output with progress bar

# Test 2: Piped input
echo "" | go run ./cmd/lazynuget restore project.csproj
# Expected: Non-interactive output, no ANSI codes

# Test 3: CI environment
CI=true go run ./cmd/lazynuget restore project.csproj
# Expected: Non-interactive output

# Test 4: Docker without TTY
docker run lazynuget restore project.csproj
# Expected: Non-interactive output

# Test 5: Docker with TTY
docker run -it lazynuget restore project.csproj
# Expected: Interactive output (if resources allow)

# Test 6: SSH without PTY allocation
ssh user@host 'lazynuget restore project.csproj'
# Expected: Non-interactive output

# Test 7: SSH with PTY allocation (-t flag)
ssh -t user@host 'lazynuget restore project.csproj'
# Expected: Interactive output
```

---

## 7. Platform-Specific Considerations

### macOS
- Uses `unix.IoctlGetTermios()` (term_unix.go)
- Works with: Terminal.app, iTerm2, Alacritty, kitty
- SSH works correctly with pseudo-terminal allocation
- **Edge case:** Vim-like subshells - term.IsTerminal still correct

### Linux
- Uses `unix.IoctlGetTermios()` (term_unix.go)
- Works with: gnome-terminal, KDE Konsole, xterm, alacritty, tmux, screen
- Container detection: IsTerminal returns false unless `docker run -it`
- SSH: Works with `-t` flag or forced PTY allocation

### Windows
- Uses `windows.GetConsoleMode()`
- Works with: Windows console (conhost.exe), Windows Terminal, PowerShell
- **WSL:** Windows Subsystem for Linux properly handles TTY detection
  - WSL runs Linux (term_unix.go), but interops correctly with Windows
- **Not a true console but works:**
  - Conemu
  - Hyper terminal
  - Git Bash (MSYS2) - uses pseudo-console emulation

### Docker/Kubernetes
- Without `-it` flags: IsTerminal returns false ✓
- With `-it` flags: IsTerminal returns true ✓
- No special handling needed - standard behavior correct

### GitHub Actions / GitLab CI / Other CI Systems
- No TTY allocated: IsTerminal returns false ✓
- `CI` env var also set: Redundant catch-all
- Structured logging format: Auto-detected by non-interactive mode ✓

---

## 8. Migration Path for LazyNuGet

1. **Phase 1 (Current):**
   - Add `golang.org/x/term` import (if not present)
   - Implement `isInteractive()` function
   - Write unit tests

2. **Phase 2 (Incremental rollout):**
   - Add feature flag: `--interactive` / `--no-interactive` explicit override
   - Default to auto-detection
   - Log detection result in verbose mode: "Running in interactive mode: true/false"

3. **Phase 3 (Full integration):**
   - Connect to progress display component
   - Connect to verbosity/output formatting
   - Disable ANSI codes in non-interactive mode
   - Use structured logging in CI

---

## 9. Dependencies

### Required
```
golang.org/x/term
```
This is already imported by many Go projects and is part of official Go tools.

### Avoid
- `github.com/mattn/go-isatty` - Stdlib suffices
- `github.com/fatih/color` - Use for color output only if needed
- Custom platform detection - Error-prone

---

## 10. References

### golang.org/x/term Implementation
- **Source:** https://github.com/golang/go/tree/master/src/golang.org/x/term
- **Unix:** Uses `unix.IoctlGetTermios()` IOCTL for TTY validation
- **Windows:** Uses `windows.GetConsoleMode()` API
- **Tested across:** Linux, macOS, Windows, FreeBSD, OpenBSD, Solaris, Plan 9, AIX, z/OS

### Related Projects
- **gonuget:** Uses `os.Stat()` + env vars (fallback approach)
- **bubbletea:** Uses `golang.org/x/term.IsTerminal()` directly (our recommendation)
- **fatih/color:** Uses `github.com/mattn/go-isatty` + env vars
- **git:** Uses custom TTY detection per platform + heuristics

### Standards
- **NO_COLOR:** https://no-color.org/
- **TERM variable:** https://man7.org/linux/man-pages/man5/terminfo.5.html
- **POSIX TTY:** https://pubs.opengroup.org/onlinepubs/9699919799/basedefs/termios.h.html

---

## Summary Table

| Aspect | Recommendation | Rationale |
|---|---|---|
| **Primary check** | `golang.org/x/term.IsTerminal()` | Most reliable stdlib-based approach |
| **Fallback env var** | `CI` environment variable | Explicit CI/CD detection |
| **Color/UI disable** | `NO_COLOR` environment variable | Respects user intent |
| **Dumb terminal** | Check `TERM=dumb` | Support ancient systems |
| **Dependencies** | Only `golang.org/x/term` (stdlib) | Minimal, well-maintained |
| **Testing** | Mock TTYDetector interface | Easy to test all scenarios |
| **Cross-platform** | Native support Windows/Unix/Plan9 | No custom platform code needed |
| **Containers** | Auto-detected correctly | Works with Docker/K8s standards |
| **SSH** | Works with `-t` PTY allocation | Standard SSH behavior |

