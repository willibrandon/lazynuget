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
