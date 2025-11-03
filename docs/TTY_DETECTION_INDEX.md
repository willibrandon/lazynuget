# TTY Detection for LazyNuGet - Documentation Index

## Overview

This directory contains comprehensive research and implementation guidance for reliable TTY detection in LazyNuGet to auto-enable non-interactive mode in CI/test environments.

---

## Documents

### 1. **TTY_DETECTION_QUICK_REFERENCE.md** (Start Here)
**Length:** ~260 lines | **Time to read:** 5 minutes

Quick one-page reference with decision summary, code patterns, and testing checklist.

**Contains:**
- Decision summary with rationale
- Environment variables reference
- Behavior matrix (all scenarios)
- Code patterns
- Testing checklist
- What NOT to do

**Best for:** Developers who need quick answers

---

### 2. **TTY_DETECTION_IMPLEMENTATION.md** (Code Templates)
**Length:** ~510 lines | **Time to read:** 15 minutes

Production-ready code with complete test suite and integration examples.

**Contains:**
- Quick start (minimal function)
- Complete detector implementation with interface
- Full test suite
- Integration examples (3 use cases)
- Testing scenarios
- Debugging tips

**Best for:** Implementers ready to write code

---

### 3. **TTY_DETECTION_RESEARCH.md** (Deep Dive)
**Length:** ~530 lines | **Time to read:** 30 minutes

Comprehensive analysis of all approaches, platform quirks, and real-world implementations.

**Contains:**
- Primary approach (golang.org/x/term)
- Supporting environment variables
- Alternative approaches (considered and rejected)
- Real-world code examples from gonuget, bubbletea
- Platform-specific behavior (Windows, Linux, macOS, Docker, SSH)
- Testing strategy
- Migration path
- References and standards

**Best for:** Architects and people making platform decisions

---

## Quick Navigation

### "I just need the code"
→ See **TTY_DETECTION_IMPLEMENTATION.md**, section "Quick Start" or "Complete Production-Ready Implementation"

### "How do I test this?"
→ See **TTY_DETECTION_QUICK_REFERENCE.md**, section "Testing Checklist"
→ Or **TTY_DETECTION_IMPLEMENTATION.md**, section "Integration Test Script"

### "Why is this the right approach?"
→ See **TTY_DETECTION_QUICK_REFERENCE.md**, section "Why This Approach"
→ Or **TTY_DETECTION_RESEARCH.md**, section "Alternatives Considered"

### "Will this work on Windows/Docker/SSH?"
→ See **TTY_DETECTION_RESEARCH.md**, section "Platform-Specific Behavior"
→ Or **TTY_DETECTION_QUICK_REFERENCE.md**, section "Platform-Specific Notes"

### "How do I integrate this into LazyNuGet?"
→ See **TTY_DETECTION_IMPLEMENTATION.md**, section "Integration Examples"

### "What if CI detection fails?"
→ See **TTY_DETECTION_RESEARCH.md**, section "Supporting Environment Variables"
→ And **TTY_DETECTION_QUICK_REFERENCE.md**, section "Environment Variables Reference"

---

## Executive Summary

### Decision
**Use `golang.org/x/term.IsTerminal()` with environment variable fallbacks**

### Rationale
This stdlib-based approach is most reliable across platforms (Windows, Unix/Linux, macOS, SSH, Docker, containers) and doesn't require external dependencies.

### Implementation
~15 lines of code for minimal version, ~100 lines with full testability interface

### Key Features
- Detects TTY using native platform APIs (IoctlGetTermios on Unix, GetConsoleMode on Windows)
- Respects CI environment variables (CI=true)
- Respects user preferences (NO_COLOR, TERM=dumb)
- No external dependencies
- Works correctly in: terminals, SSH sessions, Docker containers, CI/CD pipelines, pipes/redirects
- Cross-platform: Linux, macOS, Windows (including WSL), Plan 9, Solaris, AIX, z/OS

---

## Implementation Timeline

### Phase 1: Add Basic Detection (1-2 hours)
1. Copy code from **TTY_DETECTION_IMPLEMENTATION.md** → `internal/interactive/detector.go`
2. Copy tests from **TTY_DETECTION_IMPLEMENTATION.md** → `internal/interactive/detector_test.go`
3. Run: `go test ./internal/interactive/`
4. Update `go.mod` if needed: `go get golang.org/x/term`

### Phase 2: Integrate into Main (2-4 hours)
1. Call `interactive.IsInteractive()` in main function
2. Pass result to restore options
3. Update output formatting to respect interactive mode
4. Test with different scenarios (terminal, pipe, CI, Docker)

### Phase 3: Full Feature (4-8 hours)
1. Add `--interactive` / `--non-interactive` CLI flags
2. Connect to progress display component
3. Connect to verbosity settings
4. Disable ANSI codes in non-interactive mode
5. Use structured logging for CI environments
6. Document in README/help text

---

## Testing Guide

### Unit Tests
```bash
go test -v ./internal/interactive
go test -cover ./internal/interactive
```

### Manual Testing
```bash
# Terminal (interactive)
./lazynuget restore project.csproj

# Piped (non-interactive)
echo "" | ./lazynuget restore project.csproj

# CI environment (non-interactive)
CI=true ./lazynuget restore project.csproj

# Docker without TTY (non-interactive)
docker run lazynuget:latest restore project.csproj

# Docker with TTY (interactive)
docker run -it lazynuget:latest restore project.csproj

# SSH with pseudo-terminal (interactive)
ssh -t user@host './lazynuget restore project.csproj'

# SSH without pseudo-terminal (non-interactive)
ssh user@host './lazynuget restore project.csproj'
```

---

## Comparison with Alternative Approaches

### Alternative: os.Stdin.Stat()
```go
func isTerminal(f *os.File) bool {
	stat, _ := f.Stat()
	return (stat.Mode() & os.ModeCharDevice) != 0
}
```
**Problems:** Unreliable on Windows, fails in some TTY multiplexers, doesn't validate terminal capabilities

### Alternative: External Library (mattn/go-isatty)
**Problems:** Unnecessary dependency when stdlib is sufficient, adds bloat

### Alternative: Check TERM Only
```go
func isTerminal() bool {
	return os.Getenv("TERM") != "dumb" && os.Getenv("TERM") != ""
}
```
**Problems:** Misses Windows entirely, Docker environments, piped input

---

## References

### Standards
- NO_COLOR: https://no-color.org/
- POSIX Terminal: https://pubs.opengroup.org/onlinepubs/9699919799/
- ANSI Escape Codes: https://en.wikipedia.org/wiki/ANSI_escape_code

### Go Documentation
- golang.org/x/term: https://pkg.go.dev/golang.org/x/term
- os.Stdin.Fd(): https://pkg.go.dev/os#File.Fd

### Real-World Examples
- Go standard tools (gofmt, etc.) - use similar approach
- bubbletea - terminal UI library, uses golang.org/x/term
- gonuget - .NET restore tool, uses os.Stat() but validates with env vars

---

## Environment Variables Cheat Sheet

| Variable | Set By | Means | LazyNuGet Action |
|---|---|---|---|
| `CI=true` | GitHub Actions, GitLab CI, etc. | Running in CI pipeline | Force non-interactive |
| `CI=false` | (explicit disable) | Not in CI | Auto-detect normally |
| `NO_COLOR=1` | User or accessibility tool | Disable colors/interactive | Go non-interactive |
| `TERM=dumb` | Dumb terminal emulator | Very limited capabilities | Go non-interactive |
| `TERM=xterm-256color` | Normal terminal | Full capabilities | Go interactive (if TTY) |
| (not set) | Normal setup | Standard terminal | Auto-detect via TTY |

---

## FAQ

### Q: Will this detect CI/CD correctly?
**A:** Yes. Most CI systems (GitHub Actions, GitLab CI, CircleCI, Travis, AppVeyor, etc.) set `CI=true` and don't allocate TTY. golang.org/x/term.IsTerminal() will correctly return false.

### Q: Does this work on Windows?
**A:** Yes. It uses Windows-specific `GetConsoleMode()` API. Works with Windows Terminal, ConEmu, Hyper, PowerShell, and WSL.

### Q: What about SSH sessions?
**A:** Works correctly:
- `ssh -t user@host 'lazynuget ...'` → Interactive (pseudo-TTY allocated)
- `ssh user@host 'lazynuget ...'` → Non-interactive (no TTY)

### Q: What about Docker?
**A:** Works correctly:
- `docker run -it lazynuget:latest ...` → Interactive
- `docker run lazynuget:latest ...` → Non-interactive

### Q: Do I need external dependencies?
**A:** No. golang.org/x/term is part of official Go infrastructure and likely already a transitive dependency.

### Q: What about accessibility?
**A:** Enhanced. Respecting `NO_COLOR` and non-interactive mode improves compatibility with screen readers and accessibility tools.

### Q: Is there a performance impact?
**A:** Negligible. Single syscall per check, microsecond latency. Can be called freely.

### Q: Can I force interactive/non-interactive mode?
**A:** Yes. Recommended to add `--interactive` / `--non-interactive` CLI flags for explicit control.

---

## Implementation Checklist

- [ ] Read TTY_DETECTION_QUICK_REFERENCE.md
- [ ] Read TTY_DETECTION_IMPLEMENTATION.md
- [ ] Ensure golang.org/x/term dependency available
- [ ] Create internal/interactive/detector.go
- [ ] Create internal/interactive/detector_test.go
- [ ] Run unit tests: `go test ./internal/interactive/`
- [ ] Integrate IsInteractive() into main
- [ ] Update output formatting
- [ ] Test with terminal
- [ ] Test with pipe
- [ ] Test with CI environment (CI=true)
- [ ] Test with Docker
- [ ] Document in README
- [ ] Update help/usage text

---

## Support

For questions:
1. Check TTY_DETECTION_QUICK_REFERENCE.md FAQ section
2. Review TTY_DETECTION_RESEARCH.md for platform-specific details
3. Look at TTY_DETECTION_IMPLEMENTATION.md code examples
4. Test manually with scenarios in "Testing Guide" section

---

## Document Versions

| Document | Date | Size | Focus |
|---|---|---|---|
| TTY_DETECTION_QUICK_REFERENCE.md | 2025-11-02 | 7.1K | Quick lookup, decision summary |
| TTY_DETECTION_IMPLEMENTATION.md | 2025-11-02 | 11K | Code templates, integration |
| TTY_DETECTION_RESEARCH.md | 2025-11-02 | 14K | Deep analysis, alternatives |

---

## Related Files

- `/go.mod` - Ensure `golang.org/x/term` is listed
- `cmd/lazynuget/main.go` - Where IsInteractive() will be called
- `internal/restore/options.go` - Where NonInteractive option is used
- `internal/output/` - Where output formatting respects interactive mode

