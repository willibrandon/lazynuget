# Implementation Plan: Cross-Platform Infrastructure

**Branch**: `003-platform-abstraction` | **Date**: 2025-11-04 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/003-platform-abstraction/spec.md`

## Summary

Build comprehensive cross-platform infrastructure to abstract platform-specific behaviors for file paths, configuration directories, terminal detection, and OS conventions. This infrastructure will enable identical user experience across Windows, macOS, and Linux by handling platform differences transparently - from Windows drive letters and UNC paths to XDG Base Directory on Linux, with full terminal capability detection and graceful degradation for limited environments.

**Technical Approach**: Extend existing `internal/platform/` module (from spec 001) with enhanced path resolution, directory management, and terminal capabilities. Leverage Go standard library for OS detection, `golang.org/x/term` for terminal capabilities, and platform-specific APIs for encoding and process management. Build on existing config system (spec 002) to provide platform-aware path resolution.

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**:
- Go standard library: `os`, `path/filepath`, `runtime`, `syscall`, `os/user`
- `golang.org/x/term` (already in use from spec 001)
- `golang.org/x/sys` for platform-specific system calls (Windows GetACP, Unix locale)

**Storage**: File-based (config YAML, cache directories per platform)
**Testing**: Go testing framework (`testing` package), table-driven tests, cross-platform integration tests
**Target Platform**: Windows 10+, macOS 10.15+, Linux (all distributions with glibc 2.17+)
**Project Type**: Single project (Go binary) with internal package structure
**Performance Goals**:
- Platform detection: <1ms (cached after first call)
- Path normalization: <1ms for typical paths (<500 chars)
- Terminal detection: <10ms (one-time at startup)
- Directory creation: <50ms (filesystem-bound)
- Total platform init overhead: <15ms (added to 200ms startup budget)

**Constraints**:
- Memory footprint: <2MB for all platform abstraction state
- Zero allocations in hot paths (path normalization called frequently)
- No external C dependencies (pure Go for cross-compilation)
- Must work in restricted environments (containers, SSH, CI)

**Scale/Scope**:
- 3 operating systems × 2 architectures = 6 platform variants
- 10+ terminal emulators tested
- 500+ path test cases (edge cases, Unicode, long paths, UNC)
- 30+ functional requirements with cross-platform validation

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### ✅ I. Discoverability
- **Status**: PASS
- **Rationale**: Platform abstraction is infrastructure - not user-facing UI. Discoverability applies to TUI layer (spec 004, not this spec). Platform detection and path handling are transparent to users.

### ✅ II. Simplicity & 80/20 Rule
- **Status**: PASS
- **Rationale**: Platform abstraction handles complexity internally so application code remains simple. Common case (standard config/cache directories) requires zero configuration. Edge cases (custom paths, read-only filesystems) handled gracefully without complicating happy path.

### ✅ III. Safety & Confirmation
- **Status**: PASS
- **Rationale**: Path validation prevents directory traversal attacks. Auto-creation of directories only when parent exists prevents unintended filesystem modifications. Read-only config directory fails fast with clear error (prevents silent failures). All destructive operations (directory creation) logged for transparency.

### ✅ IV. Cross-Platform Excellence (CRITICAL)
- **Status**: PASS - This feature IS cross-platform excellence
- **Rationale**: This is the foundation for Principle IV compliance across entire codebase. Success criteria explicitly require 100% test pass rate on Windows, macOS, Linux. All platform-specific code paths tested on native platforms. Terminal compatibility matrix covers primary and secondary terminals. Performance targets uniform across platforms.

### ✅ V. Performance & Responsiveness
- **Status**: PASS
- **Rationale**: Platform init adds <15ms to 200ms startup budget (7.5% overhead). Path operations <1ms (hot path). Memory <2MB (4% of 50MB budget). All detection cached after first use. No blocking I/O in UI rendering path.

### ✅ VI. Conformity with dotnet CLI
- **Status**: PASS
- **Rationale**: Platform abstraction provides process spawning infrastructure used by dotnet CLI integration. Process encoding detection ensures correct parsing of dotnet output. Line ending handling enables cross-platform config file portability. This feature enables conformity, doesn't bypass it.

### ✅ VII. Clean, Testable, Maintainable Code
- **Status**: PASS
- **Rationale**: Clear separation: Platform detection → Path resolution → Directory management → Terminal capabilities. Each component testable in isolation. Table-driven tests for path normalization (500+ cases). Integration tests on all platforms via CI. No platform-specific code without corresponding test coverage.

### ✅ VIII. Complete Implementation - No Deferrals
- **Status**: PASS - Commitment to full implementation
- **Rationale**: All 30 functional requirements will be fully implemented. No TODOs, no deferred work. Edge cases resolved via clarifications (XDG directories, read-only cache, encoding fallback, environment variable precedence). Any remaining ambiguities will be resolved during implementation, not deferred.

### ✅ IX. No Reduction in Scope or Complexity
- **Status**: PASS - Full scope maintained
- **Rationale**: All platform-specific behaviors in scope (Windows drive letters, UNC paths, XDG spec, macOS conventions). Terminal capability detection comprehensive (color depth, Unicode, dimensions, resize events). Process spawning with full encoding support (UTF-8 + fallback). No features cut, no corners reduced. Complexity necessary for true cross-platform excellence.

**GATE RESULT**: ✅ **PASS** - All constitutional principles satisfied. Proceed to Phase 0.

## Project Structure

### Documentation (this feature)

```text
specs/003-platform-abstraction/
├── spec.md              # Feature specification (completed)
├── plan.md              # This file
├── research.md          # Phase 0: Architecture research
├── data-model.md        # Phase 1: Entity definitions
├── quickstart.md        # Phase 1: Developer guide
├── contracts/           # Phase 1: Interface contracts
│   ├── platform.md      # Platform detection interface
│   ├── paths.md         # Path resolution interface
│   ├── terminal.md      # Terminal capabilities interface
│   └── process.md       # Process spawning interface
└── tasks.md             # Phase 2: Implementation breakdown (via /speckit.tasks)
```

### Source Code (repository root)

```text
internal/platform/       # Existing from spec 001 - will be ENHANCED
├── detect.go           # EXISTING: OS/arch detection
├── detect_test.go      # EXISTING: Detection tests
├── tty.go              # EXISTING: Basic TTY detection
├── tty_test.go         # EXISTING: TTY tests
├── types.go            # EXISTING: Platform, RunMode types
├── paths.go            # NEW: Path resolution & normalization
├── paths_test.go       # NEW: Path tests (500+ cases)
├── paths_windows.go    # NEW: Windows-specific path handling
├── paths_unix.go       # NEW: Unix-specific path handling (build tag)
├── directories.go      # NEW: Config/cache directory resolution
├── directories_test.go # NEW: Directory tests
├── terminal.go         # NEW: Terminal capability detection (extends tty.go)
├── terminal_test.go    # NEW: Terminal capability tests
├── encoding.go         # NEW: Text encoding detection & conversion
├── encoding_test.go    # NEW: Encoding tests
├── encoding_windows.go # NEW: Windows code page (GetACP)
├── encoding_unix.go    # NEW: Unix locale parsing (build tag)
└── process.go          # NEW: Process spawning with encoding
└── process_test.go     # NEW: Process spawning tests

internal/config/         # Existing from spec 002 - will be ENHANCED
├── config.go           # EXISTING: Config loading
├── types.go            # EXISTING: AppConfig struct
├── defaults.go         # MODIFY: Use platform.Directories for paths
└── defaults_test.go    # MODIFY: Test platform-aware defaults

tests/integration/       # Existing - ADD new cross-platform tests
├── platform_test.go    # NEW: Full platform integration tests
├── paths_test.go       # NEW: Real filesystem path tests
├── terminal_test.go    # NEW: Terminal capability detection
└── process_test.go     # NEW: Process spawning with encoding

cmd/lazynuget/
└── main.go             # MODIFY: Use enhanced platform abstraction (minimal changes)
```

**Structure Decision**: Enhance existing `internal/platform/` package rather than creating new package. This maintains architectural coherence from spec 001 and avoids circular dependencies. Platform detection already exists; we're extending it with path handling, terminal capabilities, and process management. Build tags (`//go:build windows` and `//go:build unix`) isolate platform-specific code for testability.

## Complexity Tracking

> No constitutional violations to justify - all gates passed.

---

## Phase 0: Outline & Research

Research tasks to resolve all technical unknowns before detailed design.

### Research Task 1: Windows Path Handling Best Practices

**Question**: How to correctly handle Windows drive letters, UNC paths, and mixed separators in Go?

**Investigation Areas**:
- `filepath.Clean()` behavior on Windows vs Unix
- `filepath.ToSlash()` and `filepath.FromSlash()` for normalization
- UNC path detection regex patterns (`\\server\share` vs `//server/share`)
- Drive letter validation (A-Z, case insensitivity)
- Relative path resolution with drive letters (`C:file.txt` vs `C:\file.txt`)
- Long path support (`\\?\` prefix on Windows 10+)
- Path length limits (260 chars legacy, 32767 with long paths enabled)

**Deliverable**: Document recommended approach for path normalization, validation, and edge case handling. Include code examples and test cases.

### Research Task 2: XDG Base Directory Specification Compliance

**Question**: How to correctly implement XDG Base Directory Specification on Linux?

**Investigation Areas**:
- XDG environment variable precedence: `XDG_CONFIG_HOME`, `XDG_CACHE_HOME`, `XDG_DATA_HOME`, `XDG_STATE_HOME`
- Fallback defaults when variables unset (`~/.config`, `~/.cache`, `~/.local/share`, `~/.local/state`)
- Directory creation semantics (permissions 0700 for config, 0755 for cache)
- Parent directory validation (clarification: only create if parent exists)
- Multi-user scenarios (system-wide config in `/etc/xdg`, user overrides)
- XDG runtime directory (`XDG_RUNTIME_DIR`) - not needed for this spec but document for completeness

**Deliverable**: Implementation guide for XDG compliance with directory creation logic flowchart. Include edge case handling (non-existent parents, symlinks, permissions).

### Research Task 3: Terminal Capability Detection Methods

**Question**: How to reliably detect terminal color support, Unicode, and dimensions across platforms?

**Investigation Areas**:
- Environment variables: `TERM`, `COLORTERM`, `TERM_PROGRAM`
- terminfo database queries (`tput colors`, `tput cols`, `tput lines`)
- Fallback when terminfo unavailable (Windows, containers)
- `golang.org/x/term` capabilities: `IsTerminal()`, `GetSize()`, `GetState()`
- Color support detection: dumb (0), 16-color, 256-color, truecolor (16M)
- Unicode detection heuristics (`LC_ALL`, `LANG` environment, Windows code page 65001)
- Window resize signal handling: `SIGWINCH` on Unix, console events on Windows
- tmux/screen detection and capability overrides

**Deliverable**: Decision matrix for terminal capability detection with priority order (env vars > terminfo > fallbacks). Include platform-specific code paths and test strategy.

### Research Task 4: Process Output Encoding Detection

**Question**: How to detect and handle non-UTF8 process output on Windows and Unix?

**Investigation Areas**:
- Windows: `GetACP()` API for active code page, code page → encoding mapping
- Unix: `LC_ALL`, `LC_CTYPE`, `LANG` environment variable parsing
- Common encodings: UTF-8, UTF-16LE (Windows), ISO-8859-1, CP1252 (Windows), Shift-JIS (Japanese)
- Go encoding packages: `golang.org/x/text/encoding`, `golang.org/x/text/encoding/charmap`
- UTF-8 validation: `utf8.Valid()` before attempting decode
- Fallback strategy: UTF-8 → system encoding → best-effort (replace invalid sequences)
- Performance: encoding conversion overhead (acceptable for process output, not hot path)

**Deliverable**: Encoding detection and conversion implementation guide. Include code examples for Windows GetACP via syscall and Unix locale parsing. Document performance characteristics and error handling.

### Research Task 5: Cross-Platform Directory Conventions

**Question**: What are the correct config/cache directory paths for each platform?

**Investigation Areas**:
- **Windows**:
  - Config: `%APPDATA%\lazynuget` (e.g., `C:\Users\Name\AppData\Roaming\lazynuget`)
  - Cache: `%LOCALAPPDATA%\lazynuget` (e.g., `C:\Users\Name\AppData\Local\lazynuget`)
  - Fallback if env vars missing: `%USERPROFILE%\AppData\...`
- **macOS**:
  - Config: `~/Library/Application Support/lazynuget`
  - Cache: `~/Library/Caches/lazynuget`
  - No Logs directory needed (use cache or system logging)
- **Linux**:
  - Config: `$XDG_CONFIG_HOME/lazynuget` (default `~/.config/lazynuget`)
  - Cache: `$XDG_CACHE_HOME/lazynuget` (default `~/.cache/lazynuget`)
  - State: `$XDG_STATE_HOME/lazynuget` (default `~/.local/state/lazynuget`) - for logs if needed

**Precedence Rules**:
- Windows: APPDATA/LOCALAPPDATA always take precedence over XDG (even in WSL running Windows binary)
- macOS: No XDG support (macOS conventions only)
- Linux: XDG variables > defaults (~/.config, ~/.cache)

**Deliverable**: Platform-specific directory resolution table with precedence rules and fallback behavior. Include environment variable lookup order and validation logic.

### Research Task 6: Build Tags for Platform-Specific Code

**Question**: How to structure platform-specific code with Go build tags for maximum testability?

**Investigation Areas**:
- Build tag syntax: `//go:build windows`, `//go:build unix`, `//go:build darwin`
- Multiple constraints: `//go:build (linux || darwin) && !wasm`
- Shared interface pattern: `paths.go` (interface) + `paths_windows.go` + `paths_unix.go` (implementations)
- Testing platform-specific code: `paths_test.go` (interface tests) + `paths_windows_test.go` (Windows-only tests)
- CI matrix: GitHub Actions with `runs-on: [ubuntu-latest, macos-latest, windows-latest]`
- Local testing: `GOOS=windows go test ./...` (compile check only, not execute)
- Mock vs real platform tests: when to use mocks (unit) vs real OS (integration)

**Deliverable**: Build tag strategy document with file naming conventions and test structure. Include CI configuration guidance for cross-platform test execution.

---

## Phase 0 Output: research.md

**File**: `/Users/brandon/src/lazynuget/specs/003-platform-abstraction/research.md`

**Contents**: Consolidated findings from all 6 research tasks above, structured as:

```markdown
# Cross-Platform Infrastructure: Architecture Research

## Windows Path Handling
- **Decision**: Use filepath.Clean() + custom UNC detection
- **Rationale**: [from research]
- **Alternatives considered**: [from research]
- **Implementation notes**: [from research]

## XDG Base Directory Compliance
- **Decision**: [from research]
- **Rationale**: [from research]
...

## Terminal Capability Detection
...

## Process Output Encoding
...

## Directory Conventions
...

## Build Tag Strategy
...
```

Each section includes decision, rationale, alternatives, code examples, and test strategy.

---

## Phase 1: Design & Contracts

**Prerequisites:** `research.md` complete (all NEEDS CLARIFICATION resolved)

### 1. Data Model: Entity Definitions

**File**: `/Users/brandon/src/lazynuget/specs/003-platform-abstraction/data-model.md`

Extract entities from spec and map to Go types with validation rules:

```markdown
# Data Model: Cross-Platform Infrastructure

## Entity: Platform

**Purpose**: Represents detected operating system and architecture

**Fields**:
- `OS` (string, enum): "windows" | "darwin" | "linux" - validated from runtime.GOOS
- `Arch` (string, enum): "amd64" | "arm64" - validated from runtime.GOARCH
- `Version` (string, optional): OS version for diagnostics (e.g., "Windows 10.0.19045", "macOS 14.1")

**Validation Rules**:
- OS must be one of supported values (reject "freebsd", "openbsd", etc.)
- Arch must be one of supported values (reject "386", "arm", etc.)
- Version is informational only (no validation beyond string type)

**Invariants**:
- Platform is immutable after detection (singleton pattern)
- Detection occurs once at startup, cached for lifetime of process

---

## Entity: PathResolver

**Purpose**: Handles path normalization and validation

**Fields**:
- `configDir` (string): Resolved config directory path (platform-specific)
- `cacheDir` (string): Resolved cache directory path (platform-specific)
- `homeDir` (string): User home directory (from os/user or $HOME)

**Methods**:
- `Normalize(path string) string`: Convert to platform-native format
- `Validate(path string) error`: Check path format validity
- `IsAbsolute(path string) bool`: Platform-aware absolute path check
- `Resolve(path string) (string, error)`: Resolve relative to config dir

**Validation Rules**:
- Windows paths: Drive letter ([A-Z]:) or UNC (\\\\server\\share)
- Unix paths: Must start with / for absolute, ./ or ../ for relative
- All paths: Max 500 chars (performance target), no null bytes
- Normalized paths: Platform-native separators (\ on Windows, / on Unix)

**Edge Cases**:
- Mixed separators: C:/Users/Name → C:\Users\Name (Windows)
- UNC paths on Unix: Return error "UNC paths not supported on Unix"
- Symlinks: Resolve to target on Unix, treat as file on Windows (limited support)

---

## Entity: TerminalCapabilities

**Purpose**: Detected terminal features for UI adaptation

**Fields**:
- `ColorDepth` (enum): None (0) | Basic16 | Extended256 | TrueColor
- `SupportsUnicode` (bool): Can display Unicode characters
- `Width` (int): Terminal width in characters
- `Height` (int): Terminal height in characters
- `IsTTY` (bool): Is interactive terminal (vs redirected/piped)

**Validation Rules**:
- Width: [40, 500] range (minimum 40 per assumption A-005, max 500 reasonable)
- Height: [10, 200] range (minimum 10 per assumption A-005, max 200 reasonable)
- ColorDepth: Must be valid enum value
- All values re-queried on SIGWINCH/console events (mutable for resize)

**Detection Logic**:
- ColorDepth: COLORTERM=truecolor → TrueColor, TERM=*-256color → Extended256, else Basic16 or None
- SupportsUnicode: LANG/LC_ALL contains UTF-8, or Windows code page 65001
- Width/Height: golang.org/x/term.GetSize() with fallback to env COLUMNS/LINES
- IsTTY: golang.org/x/term.IsTerminal(os.Stdout.Fd())

---

## Entity: ProcessSpawner

**Purpose**: Abstract process execution with encoding support

**Fields**:
- `Executable` (string): Process path (dotnet, git, etc.)
- `Args` ([]string): Command arguments
- `WorkingDir` (string): Process working directory
- `Env` (map[string]string): Environment variables (merged with parent env)

**Methods**:
- `Run() (stdout, stderr string, exitCode int, err error)`: Execute and capture output
- `Start() (Process, error)`: Start async process (for future long-running commands)

**Encoding Handling**:
- Try UTF-8 decode first (utf8.Valid check)
- Fallback to system encoding: Windows GetACP → charmap, Unix LC_ALL parsing
- Replace invalid sequences with � (U+FFFD) if both fail

**Validation Rules**:
- Executable must exist and be executable (os.Stat + permission check on Unix)
- Args must not contain null bytes
- WorkingDir must exist
- Env keys must not contain = or null bytes

**Platform-Specific**:
- Windows: Add .exe/.cmd/.bat extension search, use cmd.exe for batch files
- Unix: Use PATH lookup, check execute permissions (os.Stat + mode check)
- Quoting: Windows uses CommandLineToArgvW rules, Unix uses shell quoting for spaces

---

## State Transitions

**Directory Creation**:
1. Check if directory exists (os.Stat)
2. If not exists → check if parent exists
3. If parent exists → create with os.MkdirAll, log warning
4. If parent missing → fallback to defaults (~/.config or ~/.cache)
5. If creation fails → error only if config dir (cache degrades gracefully)

**Terminal Resize**:
1. Register SIGWINCH handler (Unix) or console event handler (Windows)
2. On signal → re-query Width/Height via term.GetSize()
3. Update TerminalCapabilities fields
4. Notify UI layer to redraw (via channel or callback)
```

This document will include ALL 4 entities from spec plus state transition diagrams for directory creation and resize handling.

### 2. API Contracts: Interface Definitions

**Directory**: `/Users/brandon/src/lazynuget/specs/003-platform-abstraction/contracts/`

Generate markdown documentation files for Go interfaces based on entities:

**`contracts/platform.md`**:
```go
package contracts

// PlatformInfo provides operating system and architecture detection
type PlatformInfo interface {
    // OS returns the operating system: "windows", "darwin", or "linux"
    OS() string

    // Arch returns the architecture: "amd64" or "arm64"
    Arch() string

    // Version returns OS version string for diagnostics (optional, may be empty)
    Version() string

    // IsWindows returns true if running on Windows
    IsWindows() bool

    // IsDarwin returns true if running on macOS
    IsDarwin() bool

    // IsLinux returns true if running on Linux
    IsLinux() bool
}
```

**`contracts/paths.md`**:
```go
package contracts

// PathResolver handles platform-specific path operations
type PathResolver interface {
    // ConfigDir returns the platform-appropriate configuration directory
    // Windows: %APPDATA%\lazynuget
    // macOS: ~/Library/Application Support/lazynuget
    // Linux: $XDG_CONFIG_HOME/lazynuget or ~/.config/lazynuget
    ConfigDir() (string, error)

    // CacheDir returns the platform-appropriate cache directory
    // Windows: %LOCALAPPDATA%\lazynuget
    // macOS: ~/Library/Caches/lazynuget
    // Linux: $XDG_CACHE_HOME/lazynuget or ~/.cache/lazynuget
    CacheDir() (string, error)

    // Normalize converts path to platform-native format
    // - Windows: backslashes, drive letters uppercase
    // - Unix: forward slashes
    // - Removes redundant separators, resolves . and ..
    Normalize(path string) string

    // Validate checks if path format is valid for current platform
    // Returns error with descriptive message if invalid
    Validate(path string) error

    // IsAbsolute returns true if path is absolute for current platform
    // - Windows: starts with drive letter or UNC
    // - Unix: starts with /
    IsAbsolute(path string) bool

    // Resolve makes relative path absolute relative to config directory
    // If path is already absolute, returns it unchanged
    Resolve(path string) (string, error)
}
```

**`contracts/terminal.md`**:
```go
package contracts

// ColorDepth represents terminal color support level
type ColorDepth int

const (
    ColorNone       ColorDepth = 0   // No color support
    ColorBasic16    ColorDepth = 16  // 16 ANSI colors
    ColorExtended256 ColorDepth = 256 // 256-color palette
    ColorTrueColor  ColorDepth = 16777216 // 24-bit true color
)

// TerminalCapabilities provides terminal feature detection
type TerminalCapabilities interface {
    // GetColorDepth returns detected color support level
    GetColorDepth() ColorDepth

    // SupportsUnicode returns true if terminal can display Unicode
    SupportsUnicode() bool

    // GetSize returns terminal dimensions (width, height in characters)
    GetSize() (width int, height int, err error)

    // IsTTY returns true if stdout is connected to an interactive terminal
    IsTTY() bool

    // WatchResize registers a callback for terminal resize events
    // Returns a stop function to unregister the watcher
    WatchResize(callback func(width, height int)) (stop func())
}
```

**`contracts/process.md`**:
```go
package contracts

// ProcessResult contains the output and exit status of a process
type ProcessResult struct {
    Stdout   string // Standard output (decoded to UTF-8)
    Stderr   string // Standard error (decoded to UTF-8)
    ExitCode int    // Process exit code (0 = success)
}

// ProcessSpawner handles platform-specific process execution
type ProcessSpawner interface {
    // Run executes a process and waits for completion
    // Automatically handles:
    // - PATH resolution for executable
    // - Argument quoting for paths with spaces
    // - Output encoding detection and conversion to UTF-8
    // - Exit code extraction
    Run(executable string, args []string, workingDir string, env map[string]string) (ProcessResult, error)

    // SetEncoding overrides automatic encoding detection
    // Use "utf-8", "windows-1252", "iso-8859-1", etc.
    // Pass empty string to re-enable auto-detection
    SetEncoding(encoding string)
}
```

### 3. Quickstart Guide

**File**: `/Users/brandon/src/lazynuget/specs/003-platform-abstraction/quickstart.md`

Developer-focused guide with code examples for common scenarios:

```markdown
# Quickstart: Cross-Platform Infrastructure

## Overview

This guide shows how to use the platform abstraction layer for common tasks. All examples assume you've imported the necessary packages.

## Platform Detection

```go
import "github.com/yourusername/lazynuget/internal/platform"

func main() {
    p := platform.New()

    if p.IsWindows() {
        // Windows-specific logic
    } else if p.IsDarwin() {
        // macOS-specific logic
    } else if p.IsLinux() {
        // Linux-specific logic
    }

    fmt.Printf("Running on %s/%s\n", p.OS(), p.Arch())
}
```

## Path Resolution

```go
import "github.com/yourusername/lazynuget/internal/platform"

func getConfigPath(filename string) (string, error) {
    resolver := platform.NewPathResolver()

    // Get platform-appropriate config directory
    configDir, err := resolver.ConfigDir()
    if err != nil {
        return "", err
    }

    // Combine with filename using platform separators
    path := filepath.Join(configDir, filename)

    // Normalize to platform format
    return resolver.Normalize(path), nil
}
```

## Terminal Capabilities

```go
import "github.com/yourusername/lazynuget/internal/platform"

func setupUI() {
    term := platform.NewTerminalCapabilities()

    // Check color support
    switch term.GetColorDepth() {
    case platform.ColorTrueColor:
        // Use 24-bit RGB colors
    case platform.ColorExtended256:
        // Use 256-color palette
    case platform.ColorBasic16:
        // Use 16 ANSI colors
    default:
        // Monochrome mode
    }

    // Check Unicode support
    checkmark := "√"  // ✓ in Unicode
    if !term.SupportsUnicode() {
        checkmark = "+" // ASCII fallback
    }

    // Get dimensions
    width, height, _ := term.GetSize()
    fmt.Printf("Terminal: %dx%d\n", width, height)

    // Watch for resize
    stop := term.WatchResize(func(w, h int) {
        fmt.Printf("Resized to %dx%d\n", w, h)
        // Redraw UI
    })
    defer stop()
}
```

## Process Spawning

```go
import "github.com/yourusername/lazynuget/internal/platform"

func runDotnetRestore(projectPath string) error {
    spawner := platform.NewProcessSpawner()

    result, err := spawner.Run(
        "dotnet",
        []string{"restore", projectPath},
        filepath.Dir(projectPath),
        nil, // inherit environment
    )

    if err != nil {
        return fmt.Errorf("failed to spawn dotnet: %w", err)
    }

    if result.ExitCode != 0 {
        return fmt.Errorf("dotnet restore failed: %s", result.Stderr)
    }

    fmt.Println(result.Stdout)
    return nil
}
```

## Testing Platform-Specific Code

```go
// paths_test.go
func TestNormalize(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        wantUnix string
        wantWin  string
    }{
        {
            name: "mixed separators",
            input: "C:/Users/Dev/config.yml",
            wantUnix: "C:/Users/Dev/config.yml",  // unchanged on Unix
            wantWin: "C:\\Users\\Dev\\config.yml", // normalized on Windows
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            resolver := platform.NewPathResolver()
            got := resolver.Normalize(tt.input)

            want := tt.wantUnix
            if runtime.GOOS == "windows" {
                want = tt.wantWin
            }

            if got != want {
                t.Errorf("Normalize() = %q, want %q", got, want)
            }
        })
    }
}
```

## Common Patterns

### Graceful Degradation (Read-Only Cache)

```go
func ensureCacheDir() error {
    resolver := platform.NewPathResolver()
    cacheDir, err := resolver.CacheDir()
    if err != nil {
        return err
    }

    // Try to create cache directory
    if err := os.MkdirAll(cacheDir, 0755); err != nil {
        // Log warning but continue (cache operations will fail gracefully)
        log.Warn("Cache directory is read-only, using in-memory cache")
        return nil
    }

    return nil
}
```

### Environment Variable Precedence (Windows vs XDG)

```go
func resolveConfigDir() (string, error) {
    p := platform.New()

    if p.IsWindows() {
        // Windows: APPDATA always wins (even if XDG vars set in WSL)
        if appdata := os.Getenv("APPDATA"); appdata != "" {
            return filepath.Join(appdata, "lazynuget"), nil
        }
        return "", errors.New("APPDATA not set")
    }

    if p.IsLinux() {
        // Linux: Check XDG first, fall back to ~/.config
        if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
            return filepath.Join(xdgConfig, "lazynuget"), nil
        }

        homeDir, _ := os.UserHomeDir()
        return filepath.Join(homeDir, ".config", "lazynuget"), nil
    }

    // macOS: Use ~/Library/Application Support
    homeDir, _ := os.UserHomeDir()
    return filepath.Join(homeDir, "Library", "Application Support", "lazynuget"), nil
}
```
```

This quickstart will include 10+ code examples covering all common scenarios.

### 4. Agent Context Update

After generating all Phase 1 artifacts, run:

```bash
.specify/scripts/bash/update-agent-context.sh claude
```

This will update `.specify/memory/agent-context.md` with new technologies:
- Platform abstraction patterns (build tags, interface pattern)
- Terminal capability detection approach
- Process encoding handling strategy

Only new information is added; manual context is preserved.

---

## Phase 1 Output Summary

**Generated Files**:
1. `research.md` - Architecture decisions with rationale (~1,800 lines)
2. `data-model.md` - 4 entities with validation rules and state transitions
3. `contracts/*.md` - 4 interface documentation files with comprehensive examples
   - `platform.md` - PlatformInfo interface (OS/arch detection)
   - `paths.md` - PathResolver interface (path handling, directories)
   - `terminal.md` - TerminalCapabilities interface (color, Unicode, resize)
   - `process.md` - ProcessSpawner interface (encoding-aware execution)
4. `quickstart.md` - Developer guide with 10+ code examples
5. `.specify/memory/agent-context.md` - Updated with platform abstraction context

**Next Command**: `/speckit.tasks` to generate implementation task breakdown

---

## Notes for Implementation Phase

**Critical Success Factors**:
1. **Test Coverage**: 500+ path test cases, all platforms in CI
2. **Performance**: <15ms total overhead, zero allocs in hot paths
3. **Error Messages**: Clear, actionable (e.g., "UNC paths not supported on Unix. Use absolute path instead: /mnt/share/...")
4. **Graceful Degradation**: Read-only cache, missing XDG parents, encoding fallback
5. **Build Tags**: Isolate platform code for testability and maintainability

**Risk Areas**:
- Windows UNC paths edge cases (\\server vs //server, long UNC \\?\UNC\)
- Terminal resize signal delivery in tmux/screen (nested multiplexers)
- Encoding detection false positives (UTF-8 vs Latin-1 ambiguity)
- Symlink handling on Windows (requires admin/dev mode permissions)

**Mitigation Strategy**:
- Extensive integration tests with real paths, terminals, processes
- Fallback to conservative defaults when detection uncertain
- Clear error messages with suggested fixes
- Document known limitations in quickstart (e.g., symlink support)
