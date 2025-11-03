# TTY Detection - Quick Reference

## One-Liner Decision

**Use `golang.org/x/term.IsTerminal(int(os.Stdin.Fd()))` with environment variable fallbacks**

---

## Decision Summary

| Criterion | Choice | Why |
|-----------|--------|-----|
| **Primary method** | `golang.org/x/term.IsTerminal()` | Most reliable stdlib-based approach |
| **Dependency** | Only stdlib | No external dependencies needed |
| **Cross-platform** | Native Windows/Unix/Plan9 | Handles all major platforms correctly |
| **CI/CD detection** | Check `CI` environment variable | Explicit indicator from CI systems |
| **Fallback checks** | `NO_COLOR`, `TERM=dumb` | Respect user preferences |
| **Complexity** | ~15 lines of code | Simple, maintainable |

---

## Why This Method?

### Reliability
```
Platform      | Method                    | Result
============|========================|========
Linux       | unix.IoctlGetTermios()  | ✓ Accurate
macOS       | unix.IoctlGetTermios()  | ✓ Accurate
Windows     | windows.GetConsoleMode()| ✓ Accurate
SSH (-t)    | Same as local terminal  | ✓ Works
Docker (-it)| TTY allocated properly  | ✓ Works
CI/CD       | No TTY allocated        | ✓ Returns false
Pipes       | Not a TTY               | ✓ Returns false
```

### Why Not Alternatives?

| Alternative | Problem |
|---|---|
| `os.Stdin.Stat() & os.ModeCharDevice` | Windows console detection unreliable |
| `github.com/mattn/go-isatty` | Unnecessary external dependency |
| Check `$TERM` only | Misses Windows, Docker, CI systems |
| Check multiple FDs (stdin+stdout) | Breaks legitimate piped usage |
| Custom platform detection | Error-prone, hard to maintain |

---

## Implementation Checklist

### Step 1: Verify Dependency
```bash
grep "golang.org/x/term" go.mod
# If not present, add it:
go get golang.org/x/term
```

### Step 2: Create Detection Function
```go
package interactive

import (
	"os"
	"golang.org/x/term"
)

func IsInteractive() bool {
	if os.Getenv("CI") != "" {
		return false
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
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

### Step 3: Use in Main
```go
func main() {
	opts := &RestoreOptions{
		NonInteractive: !interactive.IsInteractive(),
	}
	// ... run restore
}
```

### Step 4: Test
```bash
go test ./internal/interactive/
```

---

## Environment Variables Reference

| Variable | Set By | Means | LazyNuGet Should |
|---|---|---|---|
| `CI=true` | GitHub Actions, GitLab CI, CircleCI, Travis, etc. | Running in CI pipeline | Force non-interactive |
| `NO_COLOR=1` | User/accessibility tool | Disable colors/interactive | Go non-interactive |
| `TERM=dumb` | Ancient terminal emulator | Very limited capabilities | Go non-interactive |
| `TERM=""` | Broken terminal setup | Unknown capabilities | Go non-interactive |
| (not set) | Normal terminal | Can handle interactive | Go interactive ✓ |

---

## Behavior Matrix

### Scenarios and Expected Behavior

```
Scenario                    | TTY | CI | Result
========================|===|===|=======
Terminal (bash/zsh)         | Y  | N  | INTERACTIVE
Terminal + CI=true          | Y  | Y  | NON-INTERACTIVE
Terminal + NO_COLOR=1       | Y  | N  | NON-INTERACTIVE
SSH with -t                 | Y  | N  | INTERACTIVE
SSH without -t             | N  | N  | NON-INTERACTIVE
Docker -it                  | Y  | N  | INTERACTIVE
Docker (no -it)            | N  | N  | NON-INTERACTIVE
Piped: cat | lazynuget     | N  | N  | NON-INTERACTIVE
Redirect: lazynuget < file | N  | N  | NON-INTERACTIVE
GitHub Actions             | N  | Y  | NON-INTERACTIVE
GitLab CI                  | N  | Y  | NON-INTERACTIVE
Screen/tmux                | Y  | N  | INTERACTIVE
WSL + Windows Terminal     | Y  | N  | INTERACTIVE
```

---

## Code Patterns

### Simple Check
```go
if interactive.IsInteractive() {
	// Show spinner, progress bar, live updates
} else {
	// Show structured output, no ANSI codes
}
```

### With Explicit Override
```go
type Options struct {
	Interactive *bool // nil = auto-detect
}

func getInteractiveMode(opts *Options) bool {
	if opts.Interactive != nil {
		return *opts.Interactive
	}
	return interactive.IsInteractive()
}
```

### In Help Text
```
Usage: lazynuget restore [options]

Options:
  --interactive       Force interactive mode (progress bar, spinner)
  --non-interactive   Force non-interactive mode (structured output)
  (none)              Auto-detect based on TTY and CI environment
```

---

## Testing Checklist

```bash
# Unit tests pass
go test ./internal/interactive/ -v

# Terminal test (interactive output expected)
./lazynuget restore project.csproj

# Pipe test (non-interactive output expected)
echo "" | ./lazynuget restore project.csproj

# CI test (non-interactive output expected)
CI=true ./lazynuget restore project.csproj

# Docker test (non-interactive output expected)
docker run --rm lazynuget:latest restore project.csproj

# SSH test (with -t flag = interactive)
ssh -t user@host './lazynuget restore project.csproj'

# SSH test (without -t flag = non-interactive)
ssh user@host './lazynuget restore project.csproj'
```

---

## Rationale Summary

| Aspect | Rationale |
|--------|-----------|
| **Why not os.Stat()?** | Windows console detection unreliable; IOCTL approach more robust |
| **Why golang.org/x/term?** | Official library, well-tested, used by major Go projects (Go tools, bubbletea, etc.) |
| **Why not mattn/go-isatty?** | Stdlib is sufficient; adds unnecessary dependency |
| **Why check CI=true?** | Catches edge cases where TTY somehow allocated in CI (rare but possible) |
| **Why NO_COLOR?** | Respects user/accessibility preferences; implies non-interactive intent |
| **Why TERM=dumb?** | Support for ancient/broken terminals; conservative approach |
| **Why stdin not stdout?** | Real interactive shells need stdin TTY; output can be redirected legitimately |

---

## Platform-Specific Notes

### Windows
- Uses Windows Console API (`GetConsoleMode()`)
- Works with: Windows Terminal, ConEmu, Hyper terminal
- WSL: Uses Linux detection through WSL interop
- Cygwin/MSYS2: Uses Windows console emulation

### macOS/Linux
- Uses Unix IOCTL call (`IoctlGetTermios()`)
- Works with: any TTY-compatible terminal
- SSH: Works when `-t` flag used (allocates pseudo-TTY)
- Containers: Correctly detects absence of TTY in detached mode

### Containers
- `docker run`: No TTY → `IsTerminal() = false`
- `docker run -it`: TTY allocated → `IsTerminal() = true`
- `docker run -t`: TTY allocated → `IsTerminal() = true`
- Kubernetes: No TTY in pods → `IsTerminal() = false`

---

## Performance

**Negligible:** Function call is single syscall, microsecond-level latency. Can be called freely.

---

## Security

**None:** TTY detection uses read-only syscalls, no security implications.

---

## Accessibility

**Enhanced:** Respecting NO_COLOR and non-interactive mode improves accessibility tool compatibility.

---

## Further Reading

- Research: `/docs/TTY_DETECTION_RESEARCH.md` (comprehensive analysis)
- Implementation: `/docs/TTY_DETECTION_IMPLEMENTATION.md` (code templates)
- golang.org/x/term: https://golang.org/x/term
- NO_COLOR standard: https://no-color.org/

