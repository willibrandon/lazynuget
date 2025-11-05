# Feature Specification: Cross-Platform Infrastructure

**Feature Branch**: `003-platform-abstraction`
**Created**: 2025-11-04
**Status**: Draft
**Input**: User description: "Build cross-platform infrastructure that abstracts platform-specific behaviors for file paths, config locations, terminal detection, and OS conventions. Must handle Windows drive letters and UNC paths, Unix-style paths, XDG Base Directory Specification on Linux, and macOS conventions. Include terminal capability detection and graceful degradation for limited terminals."

## Clarifications

### Session 2025-11-04

- Q: When XDG environment variables point to non-existent directories on Linux, how should the system respond? → A: Auto-create the directory if parent exists, log warning, continue execution
- Q: When the cache directory is on a read-only filesystem, how should the system behave? → A: Continue execution (cache operations gracefully degrade), fail only if config directory is also read-only
- Q: When terminal dimensions change during application runtime (window resize), how should the system respond? → A: Detect resize events (SIGWINCH signal), re-query dimensions, redraw UI to fit new size
- Q: When a user has both `APPDATA` (Windows) and `XDG_CONFIG_HOME` set on Windows with WSL, which takes precedence? → A: APPDATA takes precedence (Windows-native environment variables have priority on Windows)
- Q: When a spawned process produces output in non-UTF8 encoding, how should the system handle it? → A: Attempt UTF-8 first, fall back to system-detected encoding (code page on Windows, locale on Unix)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Developer Runs LazyNuGet on Any Platform (Priority: P1)

A developer installs LazyNuGet on Windows, macOS, or Linux and expects it to work identically without any platform-specific configuration or workarounds. The application automatically detects the platform, uses correct file paths for configuration and cache, and displays correctly in their terminal.

**Why this priority**: This is the foundational capability that enables all other features. Without cross-platform support, the "identical experience on all platforms" differentiator fails. This is the minimum viable platform abstraction.

**Independent Test**: Can be fully tested by installing LazyNuGet on each supported platform (Windows, macOS, Linux), running it without any configuration, and verifying it starts successfully with platform-appropriate defaults.

**Acceptance Scenarios**:

1. **Given** a fresh Windows installation, **When** user runs LazyNuGet, **Then** config is stored in `%APPDATA%\lazynuget\config.yml` and cache in `%LOCALAPPDATA%\lazynuget\cache`
2. **Given** a fresh macOS installation, **When** user runs LazyNuGet, **Then** config is stored in `~/Library/Application Support/lazynuget/config.yml` and cache in `~/Library/Caches/lazynuget`
3. **Given** a fresh Linux installation with XDG variables set, **When** user runs LazyNuGet, **Then** config is stored in `$XDG_CONFIG_HOME/lazynuget/config.yml` and cache in `$XDG_CACHE_HOME/lazynuget`
4. **Given** a fresh Linux installation without XDG variables, **When** user runs LazyNuGet, **Then** config is stored in `~/.config/lazynuget/config.yml` and cache in `~/.cache/lazynuget`
5. **Given** any platform, **When** user views the application, **Then** the UI adapts to terminal capabilities (color, unicode, dimensions)

---

### User Story 2 - User Works with Platform-Specific Paths (Priority: P2)

A user configures LazyNuGet with paths specific to their platform (Windows drive letters, UNC paths, Unix absolute/relative paths) and expects the application to handle them correctly without errors or path corruption.

**Why this priority**: Users need to specify custom locations for config files, cache directories, and NuGet sources. Supporting platform-specific path formats is essential for real-world usage but depends on the foundational platform detection from P1.

**Independent Test**: Can be fully tested by configuring custom paths on each platform (e.g., `C:\Custom\Path` on Windows, `/opt/nuget` on Linux, UNC paths like `\\server\share` on Windows) and verifying LazyNuGet correctly resolves, validates, and uses these paths.

**Acceptance Scenarios**:

1. **Given** Windows system, **When** user specifies config path `C:\Users\Dev\lazynuget.yml`, **Then** application loads config from that exact path
2. **Given** Windows system, **When** user specifies UNC path `\\fileserver\shared\nuget\config.yml`, **Then** application loads config from network share
3. **Given** Unix system, **When** user specifies absolute path `/opt/lazynuget/config.yml`, **Then** application loads config from that path
4. **Given** Unix system, **When** user specifies relative path `./configs/lazynuget.yml`, **Then** application resolves it relative to current directory
5. **Given** any platform with mixed path separators, **When** application processes paths internally, **Then** all paths are normalized to platform-native format
6. **Given** Windows path with forward slashes `C:/Users/Dev/config.yml`, **When** application normalizes path, **Then** it becomes `C:\Users\Dev\config.yml`

---

### User Story 3 - User Runs in Limited Terminal Environment (Priority: P3)

A developer runs LazyNuGet in a CI/CD environment, SSH session with limited terminal capabilities, or Windows Command Prompt without Unicode support. The application gracefully degrades its UI while maintaining full functionality.

**Why this priority**: Ensures LazyNuGet works in automation and constrained environments. This enhances usability but isn't required for core functionality. Depends on terminal detection from P1.

**Independent Test**: Can be fully tested by running LazyNuGet with `TERM=dumb`, in CI environments (GitHub Actions, Jenkins), or terminals without color/unicode support, and verifying it displays ASCII-only fallback UI without crashes.

**Acceptance Scenarios**:

1. **Given** terminal without color support, **When** user runs LazyNuGet, **Then** application displays monochrome UI with clear visual hierarchy using ASCII characters
2. **Given** terminal without Unicode support, **When** user runs LazyNuGet, **Then** application replaces Unicode symbols (✓, ✗, ▶) with ASCII equivalents (+, -, >)
3. **Given** narrow terminal (80x24), **When** user runs LazyNuGet, **Then** application adapts layout to fit available space
4. **Given** CI environment (TERM=dumb, CI=true), **When** user runs LazyNuGet with `--non-interactive` flag, **Then** application operates in text-only mode without TUI
5. **Given** terminal that supports 256 colors, **When** user runs LazyNuGet, **Then** application uses rich color palette for syntax highlighting and status indicators
6. **Given** terminal that supports only 16 colors, **When** user runs LazyNuGet, **Then** application falls back to basic 16-color palette

---

### User Story 4 - Developer Spawns Platform-Specific Processes (Priority: P2)

LazyNuGet needs to execute `dotnet` CLI commands on all platforms. The application correctly spawns processes with platform-appropriate arguments, environment variables, and handles output with correct line endings.

**Why this priority**: Core functionality depends on invoking `dotnet` commands. This is essential for NuGet operations but builds on platform detection from P1.

**Independent Test**: Can be fully tested by triggering a NuGet operation that requires process spawning (e.g., `dotnet restore`) on each platform and verifying the command executes successfully with correct output parsing.

**Acceptance Scenarios**:

1. **Given** any platform, **When** LazyNuGet spawns `dotnet restore` command, **Then** process executes with correct working directory and environment
2. **Given** Windows system, **When** LazyNuGet reads process output, **Then** CRLF line endings are correctly handled
3. **Given** Unix system, **When** LazyNuGet reads process output, **Then** LF line endings are correctly handled
4. **Given** Windows system with spaces in paths, **When** LazyNuGet spawns process with path argument, **Then** paths are properly quoted or escaped
5. **Given** any platform, **When** spawned process fails, **Then** LazyNuGet captures both stdout and stderr with platform-correct encoding
6. **Given** Unix system, **When** LazyNuGet needs to make script executable, **Then** application sets execute permissions appropriately

---

### Edge Cases

- What happens when user specifies invalid UNC path format on non-Windows system?
- **XDG non-existent directories**: System auto-creates directory if parent exists, logs warning, continues execution. If parent doesn't exist, falls back to ~/.config or ~/.cache defaults.
- **Terminal resize**: System detects resize events via SIGWINCH signal (Unix) or console events (Windows), re-queries dimensions, redraws UI to fit new size.
- What happens when terminal reports capabilities incorrectly (claims color support but doesn't render)?
- What happens when config directory path contains Unicode characters on Windows with non-Unicode locale?
- **Read-only cache directory**: System gracefully degrades cache operations, logs warning, continues execution. If config directory is also read-only, system fails startup with clear error message.
- What happens when symlinks are involved in path resolution (especially on Windows with limited symlink support)?
- **Windows with both APPDATA and XDG_CONFIG_HOME**: APPDATA takes precedence. Windows-native environment variables have priority on Windows platform.
- **Non-UTF8 process output**: System attempts UTF-8 decoding first, falls back to system-detected encoding (Windows code page via GetACP, Unix locale via LC_ALL/LANG environment variables).
- What happens when terminal is extremely small (e.g., 20x10)?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST detect operating system (Windows, macOS, Linux) and architecture (amd64, arm64) at startup
- **FR-002**: System MUST resolve platform-specific configuration directory using standard conventions (APPDATA on Windows, Application Support on macOS, XDG Base Directory on Linux)
- **FR-003**: System MUST resolve platform-specific cache directory using standard conventions (LOCALAPPDATA on Windows, Caches on macOS, XDG cache on Linux)
- **FR-029**: On Windows, system MUST prioritize Windows-native environment variables (APPDATA, LOCALAPPDATA) over XDG variables if both are set
- **FR-004**: System MUST normalize all file paths to platform-native format (backslashes on Windows, forward slashes on Unix)
- **FR-005**: System MUST handle Windows drive letters (C:, D:, etc.) in absolute paths
- **FR-006**: System MUST handle Windows UNC paths (\\server\share) for network locations
- **FR-007**: System MUST handle Unix absolute paths (/opt/..., /home/...) and relative paths (./config, ../parent)
- **FR-008**: System MUST respect XDG Base Directory Specification on Linux when XDG environment variables are set
- **FR-009**: System MUST fall back to standard defaults (~/.config, ~/.cache) when XDG variables are unset on Linux
- **FR-025**: System MUST auto-create XDG directories if they don't exist (when parent directory exists), log warning message, and continue execution
- **FR-010**: System MUST detect terminal capabilities: color support (none/16/256/truecolor), Unicode support, terminal dimensions (width/height)
- **FR-028**: System MUST detect terminal resize events (SIGWINCH signal on Unix, console events on Windows), re-query dimensions, and redraw UI to fit new size
- **FR-011**: System MUST gracefully degrade UI when terminal lacks color support (use ASCII characters and spacing for hierarchy)
- **FR-012**: System MUST gracefully degrade UI when terminal lacks Unicode support (replace Unicode symbols with ASCII equivalents)
- **FR-013**: System MUST adapt layout when terminal dimensions are below minimum thresholds (80x24 recommended, 40x10 minimum)
- **FR-014**: System MUST handle CRLF line endings on Windows when reading/writing text files
- **FR-015**: System MUST handle LF line endings on Unix when reading/writing text files
- **FR-016**: System MUST spawn processes with platform-appropriate executable discovery (PATH lookup on Unix, .exe/.cmd/.bat on Windows)
- **FR-017**: System MUST properly quote/escape command arguments when spawning processes, especially paths with spaces or special characters
- **FR-018**: System MUST capture process output (stdout/stderr) with platform-correct text encoding (UTF-8 preferred, system encoding fallback)
- **FR-030**: System MUST attempt UTF-8 decoding of process output first, then fall back to system-detected encoding (Windows code page via GetACP, Unix locale via LC_ALL/LANG)
- **FR-019**: System MUST handle process exit codes consistently across platforms (0 = success, non-zero = failure)
- **FR-020**: System MUST validate paths before use (check format validity, not just existence)
- **FR-021**: System MUST provide clear error messages when platform-specific operations fail (e.g., "UNC paths not supported on Unix")
- **FR-026**: System MUST gracefully degrade cache operations when cache directory is read-only (log warning, continue execution)
- **FR-027**: System MUST fail startup with clear error message when config directory is read-only
- **FR-022**: System MUST handle symlinks transparently on Unix systems
- **FR-023**: System MUST detect when running in CI environment (CI=true, GITHUB_ACTIONS, etc.) for automation-friendly behavior
- **FR-024**: System MUST expose platform information for logging and diagnostics (OS, arch, terminal capabilities)

### Key Entities *(include if feature involves data)*

- **Platform**: Represents the detected operating system and architecture. Attributes include OS type (Windows/macOS/Linux), architecture (amd64/arm64), version information for diagnostics.

- **PathResolver**: Encapsulates platform-specific path resolution logic. Attributes include config directory path, cache directory path, temp directory path, user home directory. Handles normalization and validation.

- **TerminalCapabilities**: Represents detected terminal features. Attributes include color depth (none/16/256/truecolor), Unicode support (boolean), dimensions (width/height in characters), TTY status (interactive vs non-interactive).

- **ProcessSpawner**: Abstracts process execution across platforms. Attributes include executable path, arguments list, working directory, environment variables. Handles output capturing and encoding.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Application starts successfully on Windows, macOS, and Linux without platform-specific configuration within 200ms (p95)
- **SC-002**: Configuration and cache directories are correctly resolved on all three platforms in 100% of test runs
- **SC-003**: All file path operations (open, read, write) succeed with Windows paths (drive letters, UNC), Unix paths (absolute, relative), and mixed separators
- **SC-004**: Application displays correctly in terminals with varying capabilities (color/no-color, Unicode/ASCII, 80x24 to 200x60 dimensions) without visual corruption
- **SC-005**: Process spawning succeeds for `dotnet` commands on all platforms with 100% success rate in test suite
- **SC-006**: Line ending handling is transparent to users - config files created on one platform are readable on another without errors
- **SC-007**: Terminal capability detection has <1% false positive/negative rate across common terminal emulators (iTerm2, Windows Terminal, GNOME Terminal, Alacritty, tmux)
- **SC-008**: Users can specify custom config/cache paths using platform-native formats without errors or warnings
- **SC-009**: Application operates correctly in CI environments (GitHub Actions, GitLab CI, Jenkins) with non-interactive mode
- **SC-010**: Path normalization performance is <1ms for typical paths (<500 characters)
- **SC-011**: Memory footprint for platform abstraction components is <2MB (measured at idle after initialization)
- **SC-012**: 100% of platform-specific code has corresponding tests for all supported platforms

## Assumptions

- **A-001**: The `dotnet` CLI is available in PATH on all platforms (validated during bootstrap as per spec 001)
- **A-002**: Users have read/write permissions to their standard config and cache directories
- **A-003**: Terminal emulators correctly report their capabilities via standard environment variables (TERM, COLORTERM) and terminfo database
- **A-004**: UTF-8 encoding is supported by all modern terminals (with graceful fallback to system encoding if needed)
- **A-005**: Minimum supported terminal size is 40x10 characters (below this, application may display warning)
- **A-006**: Windows users are on Windows 10 or later (for improved path handling and symlink support)
- **A-007**: macOS users are on macOS 10.15 (Catalina) or later
- **A-008**: Linux users have XDG-compliant desktop environments (or manually set XDG variables if needed)
- **A-009**: Process spawning requirements are limited to `dotnet` CLI and standard system commands (no exotic shell requirements)
- **A-010**: Users running in CI environments will use `--non-interactive` flag (auto-detected but explicit flag is safest)

## Dependencies

- **D-001**: Go 1.24+ standard library (`os`, `path/filepath`, `runtime`, `syscall`)
- **D-002**: `golang.org/x/term` for terminal capability detection (already used in spec 001)
- **D-003**: Existing platform detection from `internal/platform/` (spec 001)
- **D-004**: Existing config system from `internal/config/` (spec 002) - will be enhanced with path resolution
- **D-005**: XDG Base Directory Specification (https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html)
- **D-006**: Windows environment variables documentation (APPDATA, LOCALAPPDATA, USERPROFILE)
- **D-007**: macOS File System Programming Guide (Application Support, Caches directories)

## Out of Scope

- **OS-001**: Support for operating systems other than Windows, macOS, and Linux (e.g., FreeBSD, OpenBSD)
- **OS-002**: Support for architectures other than amd64 and arm64 (e.g., 32-bit x86, RISC-V)
- **OS-003**: Custom terminal emulation or terminal capability database implementation (rely on existing terminfo)
- **OS-004**: Advanced shell integration (command completion, shell-specific hooks) - this is a future feature
- **OS-005**: GUI fallback when terminal is unavailable (LazyNuGet is terminal-only)
- **OS-006**: Automatic platform-specific installer/package creation (brew, apt, chocolatey) - distribution is separate
- **OS-007**: Dynamic theme switching based on terminal color scheme - use sensible defaults
- **OS-008**: Accessibility features beyond basic terminal compatibility (screen readers, high contrast) - future work
- **OS-009**: Path conversion utilities (e.g., converting Windows paths to WSL paths) - users handle this
- **OS-010**: Advanced process management (process groups, daemonization, background services)
