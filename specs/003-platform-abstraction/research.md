# Platform Abstraction Research

## Document Overview

This document provides comprehensive research for implementing cross-platform functionality in LazyNuGet, a Go-based TUI application for NuGet package management. The research covers six critical areas: Windows path handling, XDG Base Directory specification, terminal capability detection, process output encoding, cross-platform directory conventions, and build tags strategy.

**Target Platforms**: Windows (10+), macOS (10.15+), Linux (any modern distribution)

**Key Requirements**:
- Identical user experience across all platforms
- Graceful degradation when features are unavailable
- Proper handling of platform-specific edge cases
- Maintainable code structure with minimal duplication

---

## 1. Windows Path Handling

### Decision

Use Go's `filepath` package exclusively for all path operations, with explicit handling for:
1. **Drive letters** - Normalize to uppercase, validate format
2. **UNC paths** - Detect via `\\` or `/` prefix, preserve format
3. **Long path support** - Auto-prepend `\\?\` for paths >260 chars on Windows
4. **Mixed separators** - Always use `filepath.Clean()` to normalize

**Key function**: `filepath.Clean()` handles most normalization, but requires custom logic for UNC paths and long path prefix.

### Rationale

**Why not path package**: The `path` package is designed for URL paths and always uses forward slashes. On Windows, this breaks when interacting with the OS file system, registry, or external tools.

**Why not manual string manipulation**: Manually replacing separators (`strings.ReplaceAll(p, "/", "\\")`) is error-prone and misses edge cases like:
- `C:/Users\John` (mixed separators)
- `\\?\C:\very\long\path` (already has long path prefix)
- `\\server\share` (UNC path)
- `/c/Users` (Git Bash style path)

**Why filepath.Clean()**: It's battle-tested, handles separators correctly per platform, and removes redundant elements like `.` and `..`. However, it has quirks:
- Converts `/` to `\` on Windows
- Can corrupt UNC paths if not careful (e.g., `\\server\share` → `\server\share`)
- Doesn't handle long path prefix automatically

### Alternatives Considered

**1. Third-party libraries (e.g., pathlib for Go)**
- **Rejected**: Adds dependency for functionality we can implement with stdlib. Most libraries don't handle all Windows edge cases (long paths, UNC, mixed separators).

**2. Windows-specific path package (golang.org/x/sys/windows)**
- **Partially adopted**: Used for advanced operations (GetFullPathName, GetLongPathName) but not for general path manipulation.

**3. Always using forward slashes and converting at OS boundaries**
- **Rejected**: Breaks when passing paths to external tools (dotnet CLI, NuGet.exe) which expect native separators. Also confusing for users on Windows who see backslashes everywhere else.

### Implementation Notes

#### Drive Letter Normalization

```go
// NormalizeDriveLetter ensures drive letters are uppercase and properly formatted
func NormalizeDriveLetter(path string) string {
    if len(path) < 2 {
        return path
    }

    // Check for drive letter pattern: X: or X:\
    if (path[1] == ':') && isLetter(path[0]) {
        // Normalize to uppercase
        normalized := strings.ToUpper(path[0:1]) + path[1:]
        return filepath.Clean(normalized)
    }

    return filepath.Clean(path)
}

func isLetter(c byte) bool {
    return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}
```

**Edge cases**:
- `c:/users` → `C:\users` (lowercase drive, forward slash)
- `C:file.txt` → `C:file.txt` (relative to current directory on C:)
- `C:` → `C:.` (current directory on C:)
- `/c/Users` → `/c/Users` (Git Bash format, NOT a drive letter)

#### UNC Path Detection and Handling

```go
// IsUNCPath detects Universal Naming Convention paths
func IsUNCPath(path string) bool {
    // UNC paths start with \\ or // (after normalization)
    // Examples: \\server\share, //server/share
    if len(path) < 3 {
        return false
    }

    // Check for \\ prefix (Windows native)
    if path[0] == '\\' && path[1] == '\\' {
        return true
    }

    // Check for // prefix (some tools accept this)
    if path[0] == '/' && path[1] == '/' {
        return true
    }

    return false
}

// NormalizeUNCPath ensures UNC paths maintain correct format
func NormalizeUNCPath(path string) string {
    if !IsUNCPath(path) {
        return filepath.Clean(path)
    }

    // Preserve UNC prefix but clean the rest
    // filepath.Clean() can corrupt UNC paths by removing leading separator

    if strings.HasPrefix(path, `\\`) {
        // Already Windows format
        cleaned := filepath.Clean(path)

        // Ensure we still have UNC prefix after clean
        if !strings.HasPrefix(cleaned, `\\`) {
            return `\\` + strings.TrimPrefix(cleaned, `\`)
        }
        return cleaned
    }

    if strings.HasPrefix(path, `//`) {
        // Convert to Windows format
        normalized := strings.ReplaceAll(path, "/", "\\")
        return NormalizeUNCPath(normalized)
    }

    return filepath.Clean(path)
}
```

**Edge cases**:
- `\\server\share\path` → `\\server\share\path` (preserve UNC)
- `//server/share/path` → `\\server\share\path` (normalize separators)
- `\\server\share\.\path` → `\\server\share\path` (remove redundant elements)
- `\\?\UNC\server\share` → `\\?\UNC\server\share` (long UNC path)

#### Long Path Support (>260 characters)

Windows historically limited paths to 260 characters (MAX_PATH). Modern Windows 10+ can support longer paths if:
1. Long path support is enabled in registry (Windows 10 1607+)
2. Application manifest declares long path awareness
3. Path is prefixed with `\\?\`

```go
import (
    "golang.org/x/sys/windows"
    "syscall"
)

const MAX_PATH = 260

// EnableLongPathSupport checks if long paths are enabled
func EnableLongPathSupport() bool {
    if runtime.GOOS != "windows" {
        return true // Not applicable
    }

    // Try to create a long path to test support
    // This is more reliable than checking registry
    testPath := `C:\` + strings.Repeat("a", 300) + ".txt"

    // Use GetFullPathName which fails on unsupported systems
    buf := make([]uint16, 32768) // Max path length with \\?\
    _, err := syscall.GetFullPathName(syscall.StringToUTF16Ptr(testPath), uint32(len(buf)), &buf[0], nil)

    return err == nil
}

// ToLongPath converts a path to long path format if needed
func ToLongPath(path string) string {
    if runtime.GOOS != "windows" {
        return path
    }

    // Already has long path prefix
    if strings.HasPrefix(path, `\\?\`) || strings.HasPrefix(path, `\??\`) {
        return path
    }

    // UNC paths need special handling
    if IsUNCPath(path) {
        // \\server\share → \\?\UNC\server\share
        trimmed := strings.TrimPrefix(path, `\\`)
        return `\\?\UNC\` + trimmed
    }

    // Regular paths
    // Get absolute path first
    absPath, err := filepath.Abs(path)
    if err != nil {
        return path // Fall back to original
    }

    // Check if we need long path support (>260 chars)
    if len(absPath) <= MAX_PATH {
        return absPath
    }

    // Add long path prefix
    // Must be absolute path without forward slashes
    normalized := filepath.Clean(absPath)
    return `\\?\` + normalized
}

// FromLongPath removes long path prefix for display
func FromLongPath(path string) string {
    if !strings.HasPrefix(path, `\\?\`) {
        return path
    }

    // \\?\C:\path → C:\path
    if len(path) > 4 && path[5] == ':' {
        return path[4:]
    }

    // \\?\UNC\server\share → \\server\share
    if strings.HasPrefix(path, `\\?\UNC\`) {
        return `\\` + path[8:]
    }

    return path
}
```

**Edge cases**:
- Paths with `\\?\` already present → Don't double-prefix
- UNC long paths → Use `\\?\UNC\` prefix format
- Relative paths → Convert to absolute first
- Display paths → Remove prefix for user-facing output

#### Mixed Separator Normalization

```go
// NormalizePath handles all Windows path quirks
func NormalizePath(path string) string {
    if runtime.GOOS != "windows" {
        return filepath.Clean(path)
    }

    // Handle empty path
    if path == "" {
        return "."
    }

    // Preserve long path prefix
    hasLongPrefix := strings.HasPrefix(path, `\\?\`)

    // Check for UNC path before cleaning
    isUNC := IsUNCPath(path)

    // Replace forward slashes with backslashes
    // filepath.Clean() does this but we need to handle it explicitly for UNC detection
    normalized := strings.ReplaceAll(path, "/", "\\")

    // Handle drive letters
    normalized = NormalizeDriveLetter(normalized)

    // Handle UNC paths
    if isUNC {
        normalized = NormalizeUNCPath(normalized)
    } else {
        normalized = filepath.Clean(normalized)
    }

    // Restore long path prefix if it was present
    if hasLongPrefix && !strings.HasPrefix(normalized, `\\?\`) {
        return ToLongPath(normalized)
    }

    return normalized
}
```

**Edge cases**:
- `C:/Users\John/Documents` → `C:\Users\John\Documents`
- `\\server/share\path` → `\\server\share\path`
- `/c/Users` → `/c/Users` (Git Bash format, preserved on Windows)
- `./path/./to/./file` → `path\to\file`

#### Testing Path Functions

```go
func TestNormalizePath(t *testing.T) {
    if runtime.GOOS != "windows" {
        t.Skip("Windows-specific test")
    }

    tests := []struct {
        name string
        path string
        want string
    }{
        {
            name: "lowercase drive letter",
            path: "c:/users",
            want: `C:\users`,
        },
        {
            name: "mixed separators",
            path: `C:\Users/john\Documents`,
            want: `C:\Users\john\Documents`,
        },
        {
            name: "UNC path with forward slashes",
            path: "//server/share/path",
            want: `\\server\share\path`,
        },
        {
            name: "UNC path with backslashes",
            path: `\\server\share\path`,
            want: `\\server\share\path`,
        },
        {
            name: "relative path with dots",
            path: `.\path\.\to\..\file`,
            want: `path\file`,
        },
        {
            name: "long path prefix preserved",
            path: `\\?\C:\very\long\path`,
            want: `\\?\C:\very\long\path`,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := NormalizePath(tt.path)
            if got != tt.want {
                t.Errorf("NormalizePath(%q) = %q, want %q", tt.path, got, tt.want)
            }
        })
    }
}
```

### Testing Strategy

**Unit Tests**:
1. Test each function in isolation with table-driven tests
2. Cover all edge cases (drive letters, UNC, long paths, mixed separators)
3. Use `runtime.GOOS` to skip Windows-specific tests on other platforms

**Integration Tests**:
1. Create actual files with long paths (>260 chars)
2. Test UNC paths against network shares (if available in CI)
3. Test with real dotnet CLI invocations that receive paths

**Cross-Platform Tests**:
1. Ensure Unix path handling doesn't break (`/usr/local/bin`)
2. Verify `filepath.Clean()` behavior on each platform
3. Test path joining with `filepath.Join()` across platforms

**CI Matrix**:
- Windows Server 2019, 2022 (GitHub Actions)
- macOS 12, 13 (GitHub Actions)
- Ubuntu 20.04, 22.04 (GitHub Actions)

---

## 2. XDG Base Directory Specification

### Decision

Implement full XDG Base Directory Specification (version 0.8) on Linux systems with proper fallbacks:

1. **XDG_CONFIG_HOME** - User-specific configuration files (default: `~/.config`)
2. **XDG_DATA_HOME** - User-specific data files (default: `~/.local/share`)
3. **XDG_CACHE_HOME** - User-specific cache files (default: `~/.cache`)
4. **XDG_STATE_HOME** - User-specific state data (default: `~/.local/state`) [Added in 0.8]
5. **XDG_CONFIG_DIRS** - System-wide configuration (default: `/etc/xdg`)
6. **XDG_DATA_DIRS** - System-wide data (default: `/usr/local/share:/usr/share`)

Create directories with permissions `0700` (user-only) for security, following the specification.

### Rationale

**Why XDG**: It's the standard on Linux. Users expect applications to respect XDG variables and place files in the correct locations. Non-compliance leads to:
- Cluttered home directories (`~/.lazynuget` instead of `~/.config/lazynuget`)
- Difficult backup strategies (config vs data vs cache)
- Conflicts with system-wide configurations

**Why not just ~/.config**: While `~/.config` is the default, users may have customized XDG directories (e.g., putting config on a different partition, using tmpfs for cache). We must respect their choices.

**Why 0700 permissions**: The specification recommends user-only access for security. LazyNuGet config may contain sensitive data (NuGet feed credentials, API keys). Cache and state directories are less sensitive but should still be user-only by default.

### Alternatives Considered

**1. Ignore XDG, use ~/.lazynuget for everything**
- **Rejected**: Non-compliant with Linux standards. Makes it harder for users to manage backups, sync, and cleanup.

**2. Only implement XDG_CONFIG_HOME, ignore others**
- **Rejected**: Incomplete implementation. Cache files (downloaded packages, temp data) should not be in config directory.

**3. Use 0755 permissions (world-readable)**
- **Rejected**: Security risk. Config files may contain credentials. Better to be restrictive by default.

**4. Support XDG on all platforms (Windows, macOS)**
- **Rejected**: XDG is Linux-specific. Windows has `%APPDATA%`, macOS has `~/Library`. Forcing XDG on those platforms breaks user expectations.

### Implementation Notes

#### XDG Directory Resolution

```go
package platform

import (
    "os"
    "path/filepath"
    "strings"
)

// XDGDirs holds XDG base directories
type XDGDirs struct {
    ConfigHome string   // User config: ~/.config
    DataHome   string   // User data: ~/.local/share
    CacheHome  string   // User cache: ~/.cache
    StateHome  string   // User state: ~/.local/state (XDG 0.8+)
    ConfigDirs []string // System config: /etc/xdg
    DataDirs   []string // System data: /usr/local/share, /usr/share
}

// GetXDGDirs returns XDG base directories with proper fallbacks
func GetXDGDirs() (*XDGDirs, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return nil, fmt.Errorf("failed to get home directory: %w", err)
    }

    dirs := &XDGDirs{
        ConfigHome: getEnvOrDefault("XDG_CONFIG_HOME", filepath.Join(homeDir, ".config")),
        DataHome:   getEnvOrDefault("XDG_DATA_HOME", filepath.Join(homeDir, ".local", "share")),
        CacheHome:  getEnvOrDefault("XDG_CACHE_HOME", filepath.Join(homeDir, ".cache")),
        StateHome:  getEnvOrDefault("XDG_STATE_HOME", filepath.Join(homeDir, ".local", "state")),
    }

    // ConfigDirs: colon-separated list, default /etc/xdg
    configDirs := os.Getenv("XDG_CONFIG_DIRS")
    if configDirs == "" {
        configDirs = "/etc/xdg"
    }
    dirs.ConfigDirs = strings.Split(configDirs, ":")

    // DataDirs: colon-separated list, default /usr/local/share:/usr/share
    dataDirs := os.Getenv("XDG_DATA_DIRS")
    if dataDirs == "" {
        dataDirs = "/usr/local/share:/usr/share"
    }
    dirs.DataDirs = strings.Split(dataDirs, ":")

    return dirs, nil
}

func getEnvOrDefault(key, defaultVal string) string {
    if val := os.Getenv(key); val != "" {
        return val
    }
    return defaultVal
}
```

#### Application-Specific Paths

```go
// GetConfigPath returns the application config directory
func GetConfigPath(appName string) (string, error) {
    if runtime.GOOS != "linux" {
        return "", fmt.Errorf("XDG is Linux-specific")
    }

    xdg, err := GetXDGDirs()
    if err != nil {
        return "", err
    }

    return filepath.Join(xdg.ConfigHome, appName), nil
}

// GetDataPath returns the application data directory
func GetDataPath(appName string) (string, error) {
    if runtime.GOOS != "linux" {
        return "", fmt.Errorf("XDG is Linux-specific")
    }

    xdg, err := GetXDGDirs()
    if err != nil {
        return "", err
    }

    return filepath.Join(xdg.DataHome, appName), nil
}

// GetCachePath returns the application cache directory
func GetCachePath(appName string) (string, error) {
    if runtime.GOOS != "linux" {
        return "", fmt.Errorf("XDG is Linux-specific")
    }

    xdg, err := GetXDGDirs()
    if err != nil {
        return "", err
    }

    return filepath.Join(xdg.CacheHome, appName), nil
}

// GetStatePath returns the application state directory
func GetStatePath(appName string) (string, error) {
    if runtime.GOOS != "linux" {
        return "", fmt.Errorf("XDG is Linux-specific")
    }

    xdg, err := GetXDGDirs()
    if err != nil {
        return "", err
    }

    return filepath.Join(xdg.StateHome, appName), nil
}
```

#### Directory Creation with Permissions

```go
// EnsureXDGDir creates directory if it doesn't exist, with proper permissions
func EnsureXDGDir(path string) error {
    // Check if already exists
    info, err := os.Stat(path)
    if err == nil {
        // Exists, verify it's a directory
        if !info.IsDir() {
            return fmt.Errorf("%s exists but is not a directory", path)
        }

        // Verify permissions (warn if world-readable)
        if info.Mode().Perm()&0077 != 0 {
            // World or group readable/writable
            // Log warning but don't fail
            log.Warnf("Directory %s has permissive permissions: %o", path, info.Mode().Perm())
        }

        return nil
    }

    if !os.IsNotExist(err) {
        return fmt.Errorf("failed to stat %s: %w", path, err)
    }

    // Create with 0700 permissions (user-only)
    if err := os.MkdirAll(path, 0700); err != nil {
        return fmt.Errorf("failed to create directory %s: %w", path, err)
    }

    return nil
}
```

#### System-Wide Configuration Search

```go
// FindConfigFile searches for config file in XDG hierarchy
func FindConfigFile(appName, filename string) (string, error) {
    if runtime.GOOS != "linux" {
        return "", fmt.Errorf("XDG is Linux-specific")
    }

    xdg, err := GetXDGDirs()
    if err != nil {
        return "", err
    }

    // Check user config first (highest priority)
    userConfig := filepath.Join(xdg.ConfigHome, appName, filename)
    if _, err := os.Stat(userConfig); err == nil {
        return userConfig, nil
    }

    // Check system config dirs in order
    for _, dir := range xdg.ConfigDirs {
        systemConfig := filepath.Join(dir, appName, filename)
        if _, err := os.Stat(systemConfig); err == nil {
            return systemConfig, nil
        }
    }

    // Not found, return user config path (for creation)
    return userConfig, nil
}
```

#### Edge Cases

**Empty XDG Variables**:
```bash
# If XDG_CONFIG_HOME="" (empty, not unset), should we use default?
# Specification says: "If empty, use default"
export XDG_CONFIG_HOME=""
```

```go
func getEnvOrDefault(key, defaultVal string) string {
    val := os.Getenv(key)
    // Empty string counts as "not set"
    if val == "" {
        return defaultVal
    }
    return val
}
```

**Relative Paths in XDG Variables**:
```bash
# What if user sets XDG_CONFIG_HOME to relative path?
export XDG_CONFIG_HOME="./config"
```

```go
func GetXDGDirs() (*XDGDirs, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return nil, fmt.Errorf("failed to get home directory: %w", err)
    }

    configHome := getEnvOrDefault("XDG_CONFIG_HOME", filepath.Join(homeDir, ".config"))

    // Ensure absolute path
    if !filepath.IsAbs(configHome) {
        // Make it absolute relative to home
        configHome = filepath.Join(homeDir, configHome)
    }

    dirs := &XDGDirs{
        ConfigHome: configHome,
        // ... rest
    }

    return dirs, nil
}
```

**Non-Existent Home Directory**:
```go
// What if $HOME doesn't exist? (rare but possible)
func GetXDGDirs() (*XDGDirs, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return nil, fmt.Errorf("failed to get home directory: %w", err)
    }

    // Verify home actually exists
    if _, err := os.Stat(homeDir); os.IsNotExist(err) {
        return nil, fmt.Errorf("home directory %s does not exist", homeDir)
    }

    // ... rest
}
```

**Colon in Directory Names**:
```bash
# XDG_CONFIG_DIRS uses colon as separator
# What if a directory name contains a colon? (very rare on Linux)
export XDG_CONFIG_DIRS="/etc/xdg:/opt/app:1.0/config"
# Split would break this
```

This is extremely rare and not handled by the specification. We follow the spec and split on colons. Users with colons in paths must escape them (not standard).

### Testing Strategy

**Unit Tests**:
```go
func TestGetXDGDirs(t *testing.T) {
    if runtime.GOOS != "linux" {
        t.Skip("Linux-specific test")
    }

    // Save original env
    origConfigHome := os.Getenv("XDG_CONFIG_HOME")
    defer os.Setenv("XDG_CONFIG_HOME", origConfigHome)

    tests := []struct {
        name       string
        setupEnv   func()
        wantConfig string // Pattern or exact match
    }{
        {
            name: "default when unset",
            setupEnv: func() {
                os.Unsetenv("XDG_CONFIG_HOME")
            },
            wantConfig: filepath.Join(os.Getenv("HOME"), ".config"),
        },
        {
            name: "custom XDG_CONFIG_HOME",
            setupEnv: func() {
                os.Setenv("XDG_CONFIG_HOME", "/custom/config")
            },
            wantConfig: "/custom/config",
        },
        {
            name: "empty XDG_CONFIG_HOME uses default",
            setupEnv: func() {
                os.Setenv("XDG_CONFIG_HOME", "")
            },
            wantConfig: filepath.Join(os.Getenv("HOME"), ".config"),
        },
        {
            name: "relative path made absolute",
            setupEnv: func() {
                os.Setenv("XDG_CONFIG_HOME", "my/config")
            },
            wantConfig: filepath.Join(os.Getenv("HOME"), "my/config"),
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.setupEnv()
            dirs, err := GetXDGDirs()
            if err != nil {
                t.Fatalf("GetXDGDirs() error = %v", err)
            }
            if dirs.ConfigHome != tt.wantConfig {
                t.Errorf("ConfigHome = %q, want %q", dirs.ConfigHome, tt.wantConfig)
            }
        })
    }
}
```

**Integration Tests**:
1. Create temporary home directory
2. Set XDG variables to point to temp locations
3. Create directories and files
4. Verify permissions (0700)
5. Test config file search hierarchy (user > system)

**CI Testing**:
- Run on multiple Linux distributions (Ubuntu, Fedora, Alpine)
- Test with and without XDG variables set
- Verify behavior in containerized environments (Docker)

---

## 3. Terminal Capability Detection

### Decision

Implement multi-level terminal capability detection with graceful degradation:

1. **Color Support**: Detect 256-color, true color, or monochrome
2. **Unicode Support**: Detect UTF-8 vs ASCII-only terminals
3. **Dimensions**: Get initial size and handle SIGWINCH for resize
4. **Interactive Check**: Detect if stdin/stdout are TTY

Use a combination of:
- Environment variables (`TERM`, `COLORTERM`, `LANG`, `LC_ALL`)
- `golang.org/x/term.IsTerminal()` for TTY detection
- `golang.org/x/term.GetSize()` for dimensions
- Signal handling (`syscall.SIGWINCH`) for resize events

Create a `TerminalCapabilities` struct that encapsulates all detected features.

### Rationale

**Why not assume full color/Unicode**: Not all terminals support advanced features. Examples:
- Linux console (without fbcon): 16 colors, limited Unicode
- Windows Command Prompt (pre-Windows 10): 16 colors, CP437 encoding
- SSH sessions with `TERM=dumb`: No color, no cursor movement
- CI environments: Often no TTY, no color

**Why check TERM and COLORTERM**: These are standard Unix conventions. `TERM` describes the terminal type, `COLORTERM` indicates true color support. Checking these prevents:
- Sending ANSI codes to non-ANSI terminals
- Using 256 colors on 16-color terminals (garbled output)
- Displaying emoji on ASCII-only terminals (boxes/question marks)

**Why handle SIGWINCH**: Terminal resize is common (users resize windows, tmux splits). Without handling resize:
- TUI layout breaks (text overflows, panels misaligned)
- Bubbletea receives incorrect dimensions
- Users must restart application

### Alternatives Considered

**1. Use terminfo database (golang.org/x/term/terminfo)**
- **Partially adopted**: We use `x/term` for TTY detection and size, but not full terminfo parsing (complex, large dependency).

**2. Assume xterm-compatible always**
- **Rejected**: Breaks on Windows, dumb terminals, and exotic terminal emulators.

**3. Force color/Unicode, let user opt-out**
- **Rejected**: Bad UX. Users shouldn't need to configure "no-color" on every non-color terminal. Better to detect and gracefully degrade.

**4. Use third-party TUI libraries' detection (Bubble Tea, tcell)**
- **Adopted where possible**: Bubble Tea handles much of this, but we need detection before starting Bubble Tea (for non-interactive mode).

### Implementation Notes

#### Terminal Capabilities Struct

```go
package platform

import (
    "golang.org/x/term"
    "os"
    "strings"
    "syscall"
)

// TerminalCapabilities describes terminal features
type TerminalCapabilities struct {
    IsInteractive  bool            // Is stdin/stdout a TTY?
    ColorLevel     ColorLevel      // None, 16, 256, TrueColor
    SupportsUnicode bool           // Can display Unicode characters?
    Width          int             // Terminal width (columns)
    Height         int             // Terminal height (rows)
    ResizeChannel  <-chan struct{} // Notified on SIGWINCH
}

// ColorLevel indicates terminal color support
type ColorLevel int

const (
    ColorNone ColorLevel = iota // No color (dumb terminal)
    Color16                     // 16 colors (standard ANSI)
    Color256                    // 256 colors (xterm-256color)
    ColorTrueColor              // 24-bit true color (16 million colors)
)

func (c ColorLevel) String() string {
    switch c {
    case ColorNone:
        return "none"
    case Color16:
        return "16-color"
    case Color256:
        return "256-color"
    case ColorTrueColor:
        return "true-color"
    default:
        return "unknown"
    }
}
```

#### TTY Detection

```go
// DetectTerminalCapabilities analyzes terminal features
func DetectTerminalCapabilities() *TerminalCapabilities {
    caps := &TerminalCapabilities{
        IsInteractive:  isInteractiveTerm(),
        ColorLevel:     detectColorLevel(),
        SupportsUnicode: detectUnicodeSupport(),
        Width:          80,  // Default fallback
        Height:         24,  // Default fallback
    }

    // Get actual dimensions if interactive
    if caps.IsInteractive {
        if width, height, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
            caps.Width = width
            caps.Height = height
        }
    }

    // Set up resize notification
    caps.ResizeChannel = setupResizeNotification()

    return caps
}

// isInteractiveTerm checks if stdin and stdout are TTY
func isInteractiveTerm() bool {
    // Check stdin
    stdinIsTTY := term.IsTerminal(int(os.Stdin.Fd()))

    // Check stdout
    stdoutIsTTY := term.IsTerminal(int(os.Stdout.Fd()))

    // Must be both for interactive
    return stdinIsTTY && stdoutIsTTY
}
```

**Edge cases**:
- Piped input: `cat file.txt | lazynuget` → stdin not TTY
- Redirected output: `lazynuget > output.txt` → stdout not TTY
- SSH without TTY allocation: `ssh user@host lazynuget` → no TTY

#### Color Level Detection

```go
// detectColorLevel determines terminal color support
func detectColorLevel() ColorLevel {
    // NO_COLOR env var forces no color (https://no-color.org/)
    if os.Getenv("NO_COLOR") != "" {
        return ColorNone
    }

    // Check COLORTERM for true color
    colorterm := os.Getenv("COLORTERM")
    if colorterm == "truecolor" || colorterm == "24bit" {
        return ColorTrueColor
    }

    // Check TERM for capabilities
    term := os.Getenv("TERM")

    // Dumb terminal or unset
    if term == "" || term == "dumb" {
        return ColorNone
    }

    // True color terminals
    if strings.Contains(term, "truecolor") || strings.Contains(term, "24bit") {
        return ColorTrueColor
    }

    // 256 color terminals
    if strings.Contains(term, "256color") || strings.Contains(term, "256colour") {
        return Color256
    }

    // Check for specific terminal types
    switch {
    case strings.HasPrefix(term, "xterm"),
         strings.HasPrefix(term, "screen"),
         strings.HasPrefix(term, "tmux"),
         strings.HasPrefix(term, "rxvt"):
        // Most modern xterm-like terminals support at least 256 colors
        // But we play it safe and default to 16 unless explicitly stated
        return Color16

    case term == "linux",
         term == "console":
        // Linux console typically supports 16 colors
        return Color16

    default:
        // Conservative default
        return Color16
    }
}
```

**Special cases**:
- **Windows Terminal**: Sets `COLORTERM=truecolor`
- **iTerm2**: Sets `TERM=xterm-256color`, supports true color
- **VS Code terminal**: Sets `COLORTERM=truecolor`
- **SSH with no TERM**: Falls back to Color16

#### Unicode Support Detection

```go
// detectUnicodeSupport checks if terminal can display Unicode
func detectUnicodeSupport() bool {
    // Check LANG and LC_ALL for UTF-8
    lang := os.Getenv("LANG")
    lcAll := os.Getenv("LC_ALL")

    // LC_ALL overrides LANG
    locale := lcAll
    if locale == "" {
        locale = lang
    }

    // Empty locale or C/POSIX means ASCII-only
    if locale == "" || locale == "C" || locale == "POSIX" {
        return false
    }

    // Check for UTF-8 suffix
    locale = strings.ToLower(locale)
    if strings.Contains(locale, "utf-8") || strings.Contains(locale, "utf8") {
        return true
    }

    // Windows check
    if runtime.GOOS == "windows" {
        // Windows 10+ with UTF-8 support
        // Check console code page (requires syscall)
        return checkWindowsUTF8()
    }

    // Conservative default: assume no Unicode
    return false
}

// checkWindowsUTF8 verifies Windows console supports UTF-8
func checkWindowsUTF8() bool {
    if runtime.GOOS != "windows" {
        return false
    }

    // Windows code page 65001 is UTF-8
    // We'd need golang.org/x/sys/windows to check this properly
    // For now, assume modern Windows Terminal supports UTF-8

    // Check for Windows Terminal
    if os.Getenv("WT_SESSION") != "" {
        return true
    }

    // Conservative default for older cmd.exe
    return false
}
```

**Locale Examples**:
- `LANG=en_US.UTF-8` → Unicode supported
- `LANG=C` → ASCII only
- `LANG=en_US.ISO-8859-1` → Not UTF-8 (Latin-1)
- `LANG` unset, `LC_ALL=en_GB.UTF-8` → Unicode supported

#### Terminal Resize Handling

```go
// setupResizeNotification creates channel that receives resize events
func setupResizeNotification() <-chan struct{} {
    ch := make(chan struct{}, 1) // Buffered to avoid blocking signal handler

    // Only relevant for Unix systems with SIGWINCH
    if runtime.GOOS == "windows" {
        // Windows doesn't have SIGWINCH
        // Bubble Tea handles this internally on Windows
        return ch
    }

    // Set up SIGWINCH handler
    sigwinch := make(chan os.Signal, 1)
    signal.Notify(sigwinch, syscall.SIGWINCH)

    go func() {
        for range sigwinch {
            // Non-blocking send
            select {
            case ch <- struct{}{}:
            default:
                // Channel full, drop event (next resize will notify)
            }
        }
    }()

    return ch
}
```

**Usage in Bubble Tea**:
```go
// In Bubble Tea Update function
case resizeMsg := <-caps.ResizeChannel:
    // Get new dimensions
    width, height, _ := term.GetSize(int(os.Stdout.Fd()))

    // Send Bubble Tea window size message
    return m, tea.WindowSize{
        Width:  width,
        Height: height,
    }
}
```

**Edge cases**:
- Rapid resizes → Debounce by using buffered channel (size 1)
- Resize during shutdown → Signal handler must not panic
- Windows → No SIGWINCH, rely on Bubble Tea's internal detection

#### Graceful Degradation Examples

```go
// RenderBox adapts to terminal capabilities
func RenderBox(text string, caps *TerminalCapabilities) string {
    if !caps.SupportsUnicode {
        // Use ASCII box characters
        return renderASCIIBox(text)
    }

    // Use Unicode box-drawing characters
    return renderUnicodeBox(text)
}

func renderASCIIBox(text string) string {
    // +-------+
    // | text  |
    // +-------+
    width := len(text) + 2
    top := "+" + strings.Repeat("-", width) + "+"
    middle := "| " + text + " |"
    bottom := top
    return top + "\n" + middle + "\n" + bottom
}

func renderUnicodeBox(text string) string {
    // ┌───────┐
    // │ text  │
    // └───────┘
    width := len(text) + 2
    top := "┌" + strings.Repeat("─", width) + "┐"
    middle := "│ " + text + " │"
    bottom := "└" + strings.Repeat("─", width) + "┘"
    return top + "\n" + middle + "\n" + bottom
}

// ColorizeText applies color based on support level
func ColorizeText(text string, color string, caps *TerminalCapabilities) string {
    switch caps.ColorLevel {
    case ColorNone:
        return text // No color

    case Color16:
        // Use ANSI 16-color codes
        return applyANSI16Color(text, color)

    case Color256:
        // Use ANSI 256-color codes
        return applyANSI256Color(text, color)

    case ColorTrueColor:
        // Use 24-bit RGB codes
        return applyTrueColor(text, color)

    default:
        return text
    }
}
```

### Testing Strategy

**Unit Tests**:
```go
func TestDetectColorLevel(t *testing.T) {
    tests := []struct {
        name      string
        setupEnv  func()
        want      ColorLevel
    }{
        {
            name: "NO_COLOR forces no color",
            setupEnv: func() {
                os.Setenv("NO_COLOR", "1")
                os.Setenv("TERM", "xterm-256color")
            },
            want: ColorNone,
        },
        {
            name: "COLORTERM=truecolor",
            setupEnv: func() {
                os.Unsetenv("NO_COLOR")
                os.Setenv("COLORTERM", "truecolor")
                os.Setenv("TERM", "xterm")
            },
            want: ColorTrueColor,
        },
        {
            name: "TERM=xterm-256color",
            setupEnv: func() {
                os.Unsetenv("NO_COLOR")
                os.Unsetenv("COLORTERM")
                os.Setenv("TERM", "xterm-256color")
            },
            want: Color256,
        },
        {
            name: "TERM=dumb",
            setupEnv: func() {
                os.Unsetenv("NO_COLOR")
                os.Setenv("TERM", "dumb")
            },
            want: ColorNone,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Save original env
            origNoColor := os.Getenv("NO_COLOR")
            origColorterm := os.Getenv("COLORTERM")
            origTerm := os.Getenv("TERM")
            defer func() {
                os.Setenv("NO_COLOR", origNoColor)
                os.Setenv("COLORTERM", origColorterm)
                os.Setenv("TERM", origTerm)
            }()

            tt.setupEnv()
            got := detectColorLevel()
            if got != tt.want {
                t.Errorf("detectColorLevel() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

**Integration Tests**:
1. Test with real terminal emulators (xterm, iTerm2, Windows Terminal)
2. Test with piped input/output (`echo | lazynuget | cat`)
3. Test resize handling (send SIGWINCH, verify dimensions update)
4. Test in CI environment (no TTY, TERM=dumb)

**Manual Testing**:
- Run in different terminal emulators
- Resize window during operation
- Test with `TERM=dumb` and `NO_COLOR=1`
- SSH to server and test

---

## 4. Process Output Encoding

### Decision

Implement automatic detection and transcoding of process output to UTF-8:

1. **Windows**: Use `GetConsoleOutputCP()` to get code page, transcode with `golang.org/x/text/encoding`
2. **Unix**: Parse `LANG`/`LC_ALL` environment variables, default to UTF-8
3. **Fallback**: If detection fails or encoding is unknown, use `latin1` (safe for binary data)

Create an `OutputDecoder` that wraps `io.Reader` and transcodes on-the-fly.

### Rationale

**Why this matters**: The dotnet CLI and NuGet tools may output non-UTF-8 text:
- **Windows**: Default console code page is CP437 (US), CP850 (Western Europe), CP932 (Japan), etc.
- **Unix**: Locale determines encoding (`ISO-8859-1`, `Shift-JIS`, etc.)

If we don't transcode, we get:
- Garbled text (e.g., `Ã©` instead of `é`)
- Replacement characters (`�`)
- JSON parsing failures (NuGet outputs JSON in console encoding)

**Why detect instead of assuming UTF-8**: Not all systems use UTF-8:
- Older Windows installations (pre-Windows 10 1903)
- Legacy Unix systems with non-UTF-8 locales
- Corporate environments with specific code page requirements

**Why use golang.org/x/text**: It's the standard Go library for encoding conversion, supports 100+ encodings, and handles edge cases properly.

### Alternatives Considered

**1. Assume UTF-8 everywhere**
- **Rejected**: Breaks on Windows with non-UTF-8 code pages and legacy Unix systems.

**2. Set UTF-8 code page before running dotnet (`chcp 65001`)**
- **Rejected**: Requires shelling out to `chcp`, may fail without admin rights, doesn't work in all contexts (CI, SSH).

**3. Use dotnet CLI with `--output-format json` always**
- **Partially adopted**: JSON is easier to parse, but some commands don't support it, and error messages are still in console encoding.

**4. Force UTF-8 via environment variables**
- **Partially adopted**: Set `LANG=en_US.UTF-8` for child processes on Unix, but doesn't work on Windows.

### Implementation Notes

#### Encoding Detection

```go
package platform

import (
    "fmt"
    "golang.org/x/text/encoding"
    "golang.org/x/text/encoding/charmap"
    "golang.org/x/text/encoding/japanese"
    "golang.org/x/text/encoding/korean"
    "golang.org/x/text/encoding/simplifiedchinese"
    "golang.org/x/text/encoding/traditionalchinese"
    "golang.org/x/text/encoding/unicode"
    "os"
    "runtime"
    "strings"
)

// DetectSystemEncoding returns the system's default text encoding
func DetectSystemEncoding() (encoding.Encoding, error) {
    if runtime.GOOS == "windows" {
        return detectWindowsEncoding()
    }
    return detectUnixEncoding()
}

// detectWindowsEncoding gets Windows console code page
func detectWindowsEncoding() (encoding.Encoding, error) {
    // Use golang.org/x/sys/windows to call GetConsoleOutputCP()
    cp, err := getWindowsCodePage()
    if err != nil {
        // Fallback to CP437 (US default)
        return charmap.CodePage437, nil
    }

    return codePageToEncoding(cp)
}

// detectUnixEncoding parses LANG/LC_ALL for encoding
func detectUnixEncoding() (encoding.Encoding, error) {
    // Check LC_ALL first (highest priority)
    locale := os.Getenv("LC_ALL")
    if locale == "" {
        // Fall back to LANG
        locale = os.Getenv("LANG")
    }

    if locale == "" {
        // Default to UTF-8 if no locale set
        return unicode.UTF8, nil
    }

    // Parse locale format: language_COUNTRY.ENCODING
    // Examples: en_US.UTF-8, ja_JP.EUC-JP, fr_FR.ISO-8859-1

    parts := strings.Split(locale, ".")
    if len(parts) < 2 {
        // No encoding specified, default to UTF-8
        return unicode.UTF8, nil
    }

    encodingName := strings.ToLower(parts[1])

    return nameToEncoding(encodingName)
}
```

#### Code Page Mapping

```go
// codePageToEncoding maps Windows code page to Go encoding
func codePageToEncoding(cp int) (encoding.Encoding, error) {
    switch cp {
    case 437: // US
        return charmap.CodePage437, nil
    case 850: // Western Europe
        return charmap.CodePage850, nil
    case 852: // Central Europe
        return charmap.CodePage852, nil
    case 855: // Cyrillic
        return charmap.CodePage855, nil
    case 866: // Russian
        return charmap.CodePage866, nil
    case 932: // Japanese (Shift-JIS)
        return japanese.ShiftJIS, nil
    case 936: // Simplified Chinese (GBK)
        return simplifiedchinese.GBK, nil
    case 949: // Korean
        return korean.EUCKR, nil
    case 950: // Traditional Chinese (Big5)
        return traditionalchinese.Big5, nil
    case 1250: // Windows Central Europe
        return charmap.Windows1250, nil
    case 1251: // Windows Cyrillic
        return charmap.Windows1251, nil
    case 1252: // Windows Western Europe
        return charmap.Windows1252, nil
    case 1253: // Windows Greek
        return charmap.Windows1253, nil
    case 1254: // Windows Turkish
        return charmap.Windows1254, nil
    case 1255: // Windows Hebrew
        return charmap.Windows1255, nil
    case 1256: // Windows Arabic
        return charmap.Windows1256, nil
    case 1257: // Windows Baltic
        return charmap.Windows1257, nil
    case 1258: // Windows Vietnamese
        return charmap.Windows1258, nil
    case 65001: // UTF-8
        return unicode.UTF8, nil
    default:
        return nil, fmt.Errorf("unsupported code page: %d", cp)
    }
}

// nameToEncoding maps encoding name to Go encoding
func nameToEncoding(name string) (encoding.Encoding, error) {
    name = strings.ToLower(name)
    name = strings.ReplaceAll(name, "-", "")
    name = strings.ReplaceAll(name, "_", "")

    switch name {
    case "utf8", "utf-8":
        return unicode.UTF8, nil
    case "iso88591", "iso-8859-1", "latin1":
        return charmap.ISO8859_1, nil
    case "iso88592", "iso-8859-2":
        return charmap.ISO8859_2, nil
    case "iso88595", "iso-8859-5":
        return charmap.ISO8859_5, nil
    case "iso885915", "iso-8859-15":
        return charmap.ISO8859_15, nil
    case "eucjp", "euc-jp":
        return japanese.EUCJP, nil
    case "shiftjis", "shift_jis", "sjis":
        return japanese.ShiftJIS, nil
    case "euckr", "euc-kr":
        return korean.EUCKR, nil
    case "gbk", "gb2312":
        return simplifiedchinese.GBK, nil
    case "big5":
        return traditionalchinese.Big5, nil
    default:
        return nil, fmt.Errorf("unsupported encoding: %s", name)
    }
}
```

#### Output Decoder

```go
import (
    "io"
    "golang.org/x/text/encoding"
    "golang.org/x/text/transform"
)

// OutputDecoder transcodes process output to UTF-8
type OutputDecoder struct {
    reader   io.Reader
    encoding encoding.Encoding
}

// NewOutputDecoder creates decoder for process output
func NewOutputDecoder(reader io.Reader) (*OutputDecoder, error) {
    enc, err := DetectSystemEncoding()
    if err != nil {
        // Fallback to ISO-8859-1 (Latin-1)
        // Latin-1 is safe for binary data (all bytes are valid)
        enc = charmap.ISO8859_1
    }

    return &OutputDecoder{
        reader:   reader,
        encoding: enc,
    }, nil
}

// Read transcodes data to UTF-8
func (d *OutputDecoder) Read(p []byte) (n int, err error) {
    // If already UTF-8, no transcoding needed
    if d.encoding == unicode.UTF8 {
        return d.reader.Read(p)
    }

    // Use transform.Reader for transcoding
    decoder := d.encoding.NewDecoder()
    r := transform.NewReader(d.reader, decoder)
    return r.Read(p)
}

// ReadAll reads and transcodes all output
func (d *OutputDecoder) ReadAll() ([]byte, error) {
    return io.ReadAll(d)
}
```

#### Usage with exec.Command

```go
// RunDotnetCommand executes dotnet CLI with proper encoding
func RunDotnetCommand(args ...string) (string, error) {
    cmd := exec.Command("dotnet", args...)

    // Capture stdout
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return "", err
    }

    // Start command
    if err := cmd.Start(); err != nil {
        return "", err
    }

    // Create decoder
    decoder, err := NewOutputDecoder(stdout)
    if err != nil {
        return "", err
    }

    // Read and transcode output
    output, err := decoder.ReadAll()
    if err != nil {
        return "", err
    }

    // Wait for command to finish
    if err := cmd.Wait(); err != nil {
        return "", err
    }

    return string(output), nil
}
```

#### Windows Code Page Detection (syscall)

```go
//go:build windows

package platform

import (
    "syscall"
    "golang.org/x/sys/windows"
)

// getWindowsCodePage gets the console output code page
func getWindowsCodePage() (int, error) {
    kernel32 := syscall.NewLazyDLL("kernel32.dll")
    getConsoleOutputCP := kernel32.NewProc("GetConsoleOutputCP")

    ret, _, err := getConsoleOutputCP.Call()
    if ret == 0 {
        return 0, err
    }

    return int(ret), nil
}

// SetWindowsCodePage attempts to set console to UTF-8
func SetWindowsCodePage(cp int) error {
    kernel32 := syscall.NewLazyDLL("kernel32.dll")
    setConsoleOutputCP := kernel32.NewProc("SetConsoleOutputCP")

    ret, _, err := setConsoleOutputCP.Call(uintptr(cp))
    if ret == 0 {
        return err
    }

    return nil
}

// TryEnableUTF8 attempts to enable UTF-8 mode on Windows
func TryEnableUTF8() error {
    // Try to set UTF-8 (code page 65001)
    if err := SetWindowsCodePage(65001); err != nil {
        // Not fatal, we'll transcode instead
        return err
    }
    return nil
}
```

#### Edge Cases

**Binary Output**:
If process outputs binary data (not text), transcoding will corrupt it.

```go
// IsBinaryOutput checks if data looks like binary
func IsBinaryOutput(data []byte) bool {
    // Heuristic: if >10% of bytes are control characters (except common ones),
    // it's probably binary

    controlChars := 0
    for _, b := range data {
        // Control characters (0x00-0x1F) except tab, newline, carriage return
        if b < 0x20 && b != '\t' && b != '\n' && b != '\r' {
            controlChars++
        }
    }

    // If >10% control chars, assume binary
    return float64(controlChars)/float64(len(data)) > 0.1
}

// SmartDecoder decides whether to transcode
func SmartDecoder(reader io.Reader) (io.Reader, error) {
    // Read first 512 bytes to detect
    buf := make([]byte, 512)
    n, err := reader.Read(buf)
    if err != nil && err != io.EOF {
        return nil, err
    }

    sample := buf[:n]

    if IsBinaryOutput(sample) {
        // Don't transcode binary data
        return io.MultiReader(bytes.NewReader(sample), reader), nil
    }

    // Transcode text data
    decoder, err := NewOutputDecoder(io.MultiReader(bytes.NewReader(sample), reader))
    if err != nil {
        return nil, err
    }

    return decoder, nil
}
```

**Invalid Sequences**:
Some encodings may have invalid byte sequences. The `encoding` package replaces these with `U+FFFD` (replacement character).

```go
// Example with error handling
decoder := enc.NewDecoder()
decoder = decoder.Transformer.(transform.Transformer)

// Check for replacement characters in output
output, err := decoder.ReadAll()
if bytes.Contains(output, []byte("�")) {
    // Warning: invalid characters were replaced
    log.Warn("Output contains invalid characters, encoding detection may be wrong")
}
```

### Testing Strategy

**Unit Tests**:
```go
func TestDetectSystemEncoding(t *testing.T) {
    // Save original env
    origLang := os.Getenv("LANG")
    defer os.Setenv("LANG", origLang)

    tests := []struct {
        name    string
        lang    string
        want    encoding.Encoding
        wantErr bool
    }{
        {
            name: "UTF-8 locale",
            lang: "en_US.UTF-8",
            want: unicode.UTF8,
        },
        {
            name: "ISO-8859-1 locale",
            lang: "en_US.ISO-8859-1",
            want: charmap.ISO8859_1,
        },
        {
            name: "No locale (default to UTF-8)",
            lang: "",
            want: unicode.UTF8,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if runtime.GOOS == "windows" {
                t.Skip("Unix-specific test")
            }

            os.Setenv("LANG", tt.lang)
            got, err := DetectSystemEncoding()
            if (err != nil) != tt.wantErr {
                t.Errorf("DetectSystemEncoding() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("DetectSystemEncoding() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

**Integration Tests**:
1. Create test program that outputs non-UTF-8 text
2. Run dotnet CLI on system with non-UTF-8 code page (Windows with CP437)
3. Verify transcoding produces correct UTF-8
4. Test with various languages (Japanese, Russian, Arabic)

**Manual Testing**:
- Windows: `chcp 437`, run dotnet, verify output
- Linux: `LANG=ja_JP.EUC-JP`, run dotnet, verify output
- Test with NuGet package names containing Unicode

---

## 5. Cross-Platform Directory Conventions

### Decision

Implement platform-specific directory conventions with proper precedence:

| Purpose | Windows | macOS | Linux |
|---------|---------|-------|-------|
| **Config** | `%APPDATA%\lazynuget` | `~/Library/Application Support/lazynuget` | `~/.config/lazynuget` (XDG) |
| **Data** | `%LOCALAPPDATA%\lazynuget` | `~/Library/Application Support/lazynuget` | `~/.local/share/lazynuget` (XDG) |
| **Cache** | `%LOCALAPPDATA%\lazynuget\Cache` | `~/Library/Caches/lazynuget` | `~/.cache/lazynuget` (XDG) |
| **Logs** | `%LOCALAPPDATA%\lazynuget\Logs` | `~/Library/Logs/lazynuget` | `~/.local/state/lazynuget` (XDG) |

**Precedence Rules** (highest to lowest):
1. Explicit CLI flag (`--config /path/to/config.yml`)
2. Environment variable (`LAZYNUGET_CONFIG`)
3. Platform-specific default (above table)
4. Fallback to temp directory if none writable

### Rationale

**Why platform-specific**: Users expect applications to follow OS conventions:
- **Windows**: `%APPDATA%` is for roaming data (synced across machines), `%LOCALAPPDATA%` is for machine-specific data
- **macOS**: `~/Library` is standard, subdirectories (`Application Support`, `Caches`, `Logs`) are organized by purpose
- **Linux**: XDG Base Directory Specification is the standard

**Why not cross-platform directory**: Using `.lazynuget` in home directory on all platforms is:
- Non-standard on Windows and macOS (clutters home directory)
- Makes backup/sync harder (mixing config, cache, logs)
- Breaks platform conventions (antivirus, backup software expect standard locations)

**Why APPDATA vs LOCALAPPDATA on Windows**:
- `APPDATA`: Roaming profile, synced across domain machines (config files)
- `LOCALAPPDATA`: Machine-specific (cache, logs, large data)

We use `APPDATA` for config (allows roaming), `LOCALAPPDATA` for everything else.

### Alternatives Considered

**1. Use .lazynuget in home directory everywhere**
- **Rejected**: Non-standard, clutters home directory on Windows/macOS.

**2. Use os.UserConfigDir() only**
- **Partially adopted**: `os.UserConfigDir()` gives us `APPDATA` on Windows, `~/Library/Application Support` on macOS, `~/.config` on Linux. But we need separate paths for cache/logs/data.

**3. Let user customize all paths in config**
- **Partially adopted**: We allow CLI flag and env var overrides, but not full customization in config file (creates chicken-and-egg problem: where is config file?).

### Implementation Notes

#### Platform-Specific Paths

```go
package platform

import (
    "fmt"
    "os"
    "path/filepath"
    "runtime"
)

const AppName = "lazynuget"

// AppDirs holds all application directories
type AppDirs struct {
    Config string // Configuration files
    Data   string // Persistent data
    Cache  string // Cache files (safe to delete)
    Logs   string // Log files
}

// GetAppDirs returns platform-appropriate application directories
func GetAppDirs() (*AppDirs, error) {
    switch runtime.GOOS {
    case "windows":
        return getWindowsDirs()
    case "darwin":
        return getDarwinDirs()
    case "linux", "freebsd", "openbsd", "netbsd":
        return getLinuxDirs()
    default:
        return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
    }
}

// getWindowsDirs returns Windows directory paths
func getWindowsDirs() (*AppDirs, error) {
    // Get APPDATA (roaming profile)
    appData := os.Getenv("APPDATA")
    if appData == "" {
        return nil, fmt.Errorf("APPDATA environment variable not set")
    }

    // Get LOCALAPPDATA (machine-specific)
    localAppData := os.Getenv("LOCALAPPDATA")
    if localAppData == "" {
        // Fallback: LOCALAPPDATA is usually APPDATA\..\Local
        userProfile := os.Getenv("USERPROFILE")
        if userProfile != "" {
            localAppData = filepath.Join(userProfile, "AppData", "Local")
        } else {
            localAppData = appData // Last resort
        }
    }

    return &AppDirs{
        Config: filepath.Join(appData, AppName),
        Data:   filepath.Join(localAppData, AppName),
        Cache:  filepath.Join(localAppData, AppName, "Cache"),
        Logs:   filepath.Join(localAppData, AppName, "Logs"),
    }, nil
}

// getDarwinDirs returns macOS directory paths
func getDarwinDirs() (*AppDirs, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return nil, fmt.Errorf("failed to get home directory: %w", err)
    }

    libraryDir := filepath.Join(homeDir, "Library")

    return &AppDirs{
        Config: filepath.Join(libraryDir, "Application Support", AppName),
        Data:   filepath.Join(libraryDir, "Application Support", AppName),
        Cache:  filepath.Join(libraryDir, "Caches", AppName),
        Logs:   filepath.Join(libraryDir, "Logs", AppName),
    }, nil
}

// getLinuxDirs returns Linux directory paths (XDG)
func getLinuxDirs() (*AppDirs, error) {
    xdg, err := GetXDGDirs()
    if err != nil {
        return nil, err
    }

    return &AppDirs{
        Config: filepath.Join(xdg.ConfigHome, AppName),
        Data:   filepath.Join(xdg.DataHome, AppName),
        Cache:  filepath.Join(xdg.CacheHome, AppName),
        Logs:   filepath.Join(xdg.StateHome, AppName), // XDG 0.8: state includes logs
    }, nil
}
```

#### Precedence with Overrides

```go
// GetConfigPath returns config directory with precedence
func GetConfigPath() (string, error) {
    // 1. Check CLI flag (passed separately)
    // (handled in config package)

    // 2. Check environment variable
    if envPath := os.Getenv("LAZYNUGET_CONFIG"); envPath != "" {
        // Validate it's a directory
        if info, err := os.Stat(envPath); err == nil && info.IsDir() {
            return envPath, nil
        }
        // If set but invalid, log warning and continue to default
        log.Warnf("LAZYNUGET_CONFIG is set but invalid: %s", envPath)
    }

    // 3. Use platform default
    dirs, err := GetAppDirs()
    if err != nil {
        return "", err
    }

    return dirs.Config, nil
}

// Similar functions for GetDataPath, GetCachePath, GetLogsPath
```

#### Fallback to Temp Directory

```go
// GetConfigPathWithFallback tries standard locations then falls back to temp
func GetConfigPathWithFallback() string {
    // Try standard path
    path, err := GetConfigPath()
    if err == nil {
        // Try to create it
        if err := EnsureDir(path); err == nil {
            return path
        }
    }

    // Fall back to temp directory
    tempDir := os.TempDir()
    fallbackPath := filepath.Join(tempDir, AppName, "config")

    log.Warnf("Using temporary config directory: %s", fallbackPath)

    // Try to create temp dir
    if err := EnsureDir(fallbackPath); err != nil {
        // Last resort: current directory
        log.Errorf("Failed to create temp config dir, using current directory")
        return "."
    }

    return fallbackPath
}
```

#### Directory Initialization

```go
// InitializeAppDirs creates all application directories
func InitializeAppDirs() (*AppDirs, error) {
    dirs, err := GetAppDirs()
    if err != nil {
        return nil, err
    }

    // Create all directories
    for _, dir := range []string{dirs.Config, dirs.Data, dirs.Cache, dirs.Logs} {
        if err := EnsureDir(dir); err != nil {
            return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
        }
    }

    return dirs, nil
}

// EnsureDir creates directory with appropriate permissions
func EnsureDir(path string) error {
    // Check if exists
    info, err := os.Stat(path)
    if err == nil {
        if !info.IsDir() {
            return fmt.Errorf("%s exists but is not a directory", path)
        }
        return nil
    }

    if !os.IsNotExist(err) {
        return fmt.Errorf("failed to stat %s: %w", path, err)
    }

    // Create with appropriate permissions
    // Windows: 0755 (permissions don't work the same way)
    // Unix: 0700 (user-only for security)
    perm := os.FileMode(0700)
    if runtime.GOOS == "windows" {
        perm = 0755
    }

    if err := os.MkdirAll(path, perm); err != nil {
        return fmt.Errorf("failed to create %s: %w", path, err)
    }

    return nil
}
```

#### Config File Path

```go
// GetConfigFilePath returns full path to config file
func GetConfigFilePath() (string, error) {
    configDir, err := GetConfigPath()
    if err != nil {
        return "", err
    }

    return filepath.Join(configDir, "config.yml"), nil
}

// GetConfigFilePathWithFallback includes temp fallback
func GetConfigFilePathWithFallback() string {
    configDir := GetConfigPathWithFallback()
    return filepath.Join(configDir, "config.yml")
}
```

#### Edge Cases

**APPDATA Not Set on Windows**:
```go
func getWindowsDirs() (*AppDirs, error) {
    appData := os.Getenv("APPDATA")
    if appData == "" {
        // Try to construct from USERPROFILE
        userProfile := os.Getenv("USERPROFILE")
        if userProfile != "" {
            appData = filepath.Join(userProfile, "AppData", "Roaming")
        } else {
            // Last resort: temp directory
            appData = filepath.Join(os.TempDir(), AppName)
        }
    }

    // ... rest
}
```

**Home Directory Not Found**:
```go
func getDarwinDirs() (*AppDirs, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        // Try HOME env var
        homeDir = os.Getenv("HOME")
        if homeDir == "" {
            return nil, fmt.Errorf("failed to determine home directory")
        }
    }

    // ... rest
}
```

**Read-Only File System**:
```go
func EnsureDir(path string) error {
    // ... try to create ...

    if err := os.MkdirAll(path, perm); err != nil {
        // Check if it's a permission error
        if os.IsPermission(err) {
            // Try temp directory instead
            tempPath := filepath.Join(os.TempDir(), filepath.Base(path))
            log.Warnf("Cannot create %s (permission denied), using %s", path, tempPath)
            return os.MkdirAll(tempPath, perm)
        }
        return err
    }

    return nil
}
```

### Testing Strategy

**Unit Tests**:
```go
func TestGetAppDirs(t *testing.T) {
    dirs, err := GetAppDirs()
    if err != nil {
        t.Fatalf("GetAppDirs() error = %v", err)
    }

    // Verify all paths are absolute
    for name, path := range map[string]string{
        "Config": dirs.Config,
        "Data":   dirs.Data,
        "Cache":  dirs.Cache,
        "Logs":   dirs.Logs,
    } {
        if !filepath.IsAbs(path) {
            t.Errorf("%s path is not absolute: %s", name, path)
        }
    }

    // Verify platform-specific expectations
    switch runtime.GOOS {
    case "windows":
        if !strings.Contains(dirs.Config, "AppData") {
            t.Errorf("Windows config path should contain AppData: %s", dirs.Config)
        }

    case "darwin":
        if !strings.Contains(dirs.Config, "Library") {
            t.Errorf("macOS config path should contain Library: %s", dirs.Config)
        }

    case "linux":
        if !strings.Contains(dirs.Config, ".config") && !strings.Contains(dirs.Config, "XDG") {
            t.Errorf("Linux config path should contain .config: %s", dirs.Config)
        }
    }
}
```

**Integration Tests**:
1. Create directories and write files
2. Verify permissions (0700 on Unix)
3. Test fallback when primary location unavailable
4. Test environment variable overrides

**CI Testing**:
- Test on Windows (GitHub Actions: windows-2019, windows-2022)
- Test on macOS (GitHub Actions: macos-12, macos-13)
- Test on Linux (GitHub Actions: ubuntu-20.04, ubuntu-22.04)
- Test with custom XDG variables on Linux

---

## 6. Build Tags Strategy

### Decision

Use Go build tags (constraints) to separate platform-specific code:

1. **File-based tags**: `file_windows.go`, `file_unix.go`, `file_linux.go`
2. **Explicit tag comments**: `//go:build windows` at top of file
3. **Interface pattern**: Define common interface, implement per-platform
4. **Factory function**: Single entry point returns platform-appropriate implementation

**Structure**:
```
internal/platform/
├── platform.go          // Interface definitions, factory function
├── platform_windows.go  // Windows implementation
├── platform_unix.go     // Unix implementation (macOS, Linux, BSD)
├── platform_linux.go    // Linux-specific (if needed)
├── platform_darwin.go   // macOS-specific (if needed)
├── encoding.go          // Shared encoding logic
├── encoding_windows.go  // Windows code page detection
└── encoding_unix.go     // Unix locale detection
```

### Rationale

**Why build tags**: They allow clean separation of platform-specific code without runtime checks. Benefits:
- Code is only compiled on target platform (reduces binary size)
- IDE/LSP correctly handles platform-specific types and functions
- Impossible to accidentally call Windows-specific code on Unix
- Tests only run on relevant platforms

**Why interface pattern**: Provides a unified API while allowing platform-specific implementations. Alternative (big if/else blocks) leads to:
- Code duplication
- Hard to test (can't mock platform-specific behavior)
- Messy file with mixed Windows/Unix code

**Why factory function**: Single entry point (`platform.New()`) that returns the right implementation. Caller doesn't need to know platform details.

### Alternatives Considered

**1. Runtime checks (if runtime.GOOS == "windows")**
- **Rejected**: Compiles all platform code into every binary, increases size, allows mistakes (calling Windows APIs on Linux).

**2. Separate packages per platform (internal/platform/windows, internal/platform/unix)**
- **Rejected**: Creates import complexity, harder to share common code.

**3. Build tags only, no interfaces**
- **Rejected**: Leads to duplicated function signatures across files, harder to ensure consistency.

### Implementation Notes

#### Build Tag Syntax

Go 1.17+ uses `//go:build` (not `// +build`):

```go
//go:build windows
// +build windows

package platform

// Windows-specific code
```

**Important**:
- Must be first line of file (before package)
- Blank line required between `//go:build` and `package`
- Old `// +build` is for Go 1.16 compatibility (can be omitted if Go 1.18+ only)

**Common tags**:
- `windows` - Windows
- `darwin` - macOS
- `linux` - Linux
- `freebsd`, `openbsd`, `netbsd` - BSDs
- `unix` - Any Unix-like (macOS, Linux, BSD) - custom build tag we define
- `!windows` - Not Windows (equivalent to Unix-like)

#### File Naming Convention

Go automatically applies build tags based on filename:
- `file_windows.go` → Compiled only on Windows
- `file_linux.go` → Compiled only on Linux
- `file_darwin.go` → Compiled only on macOS
- `file_unix.go` → Requires explicit tag (not automatic)

We still add explicit `//go:build` comments for clarity.

#### Interface Definition

```go
//go:build !windows
// For Unix-like systems, we need to define a build tag

package platform

// Platform provides platform-specific functionality
type Platform interface {
    // GetAppDirs returns application directories
    GetAppDirs() (*AppDirs, error)

    // DetectEncoding returns system text encoding
    DetectEncoding() (encoding.Encoding, error)

    // GetTerminalCapabilities analyzes terminal features
    GetTerminalCapabilities() (*TerminalCapabilities, error)

    // NormalizePath handles platform-specific path quirks
    NormalizePath(path string) string
}

// New creates a platform-specific implementation
func New() (Platform, error) {
    // Implementation determined at compile time by build tags
    return newPlatform()
}
```

#### Windows Implementation

```go
//go:build windows

package platform

import (
    "golang.org/x/sys/windows"
)

type windowsPlatform struct {
    // Windows-specific fields
}

// newPlatform creates Windows implementation
func newPlatform() (Platform, error) {
    return &windowsPlatform{}, nil
}

func (p *windowsPlatform) GetAppDirs() (*AppDirs, error) {
    // Windows-specific implementation
    return getWindowsDirs()
}

func (p *windowsPlatform) DetectEncoding() (encoding.Encoding, error) {
    // Windows-specific implementation
    return detectWindowsEncoding()
}

func (p *windowsPlatform) GetTerminalCapabilities() (*TerminalCapabilities, error) {
    // Windows terminal detection
    return detectWindowsTerminal()
}

func (p *windowsPlatform) NormalizePath(path string) string {
    // Windows path normalization
    return normalizeWindowsPath(path)
}
```

#### Unix Implementation

```go
//go:build !windows

package platform

type unixPlatform struct {
    // Unix-specific fields
}

// newPlatform creates Unix implementation
func newPlatform() (Platform, error) {
    return &unixPlatform{}, nil
}

func (p *unixPlatform) GetAppDirs() (*AppDirs, error) {
    // Unix-specific implementation (XDG on Linux, ~/Library on macOS)
    if runtime.GOOS == "darwin" {
        return getDarwinDirs()
    }
    return getLinuxDirs()
}

func (p *unixPlatform) DetectEncoding() (encoding.Encoding, error) {
    // Unix locale detection
    return detectUnixEncoding()
}

func (p *unixPlatform) GetTerminalCapabilities() (*TerminalCapabilities, error) {
    // Unix terminal detection
    return detectUnixTerminal()
}

func (p *unixPlatform) NormalizePath(path string) string {
    // Unix path normalization (simpler than Windows)
    return filepath.Clean(path)
}
```

#### Linux-Specific Code (if needed)

```go
//go:build linux

package platform

// LinuxSpecific provides Linux-only functionality
type LinuxSpecific interface {
    Platform

    // GetXDGDirs returns XDG base directories
    GetXDGDirs() (*XDGDirs, error)
}

type linuxPlatform struct {
    unixPlatform // Embed common Unix implementation
}

func newPlatform() (Platform, error) {
    return &linuxPlatform{}, nil
}

func (p *linuxPlatform) GetXDGDirs() (*XDGDirs, error) {
    // Linux-specific XDG implementation
    return getXDGDirs()
}
```

#### Shared Code Pattern

Code that's common across platforms goes in non-tagged files:

```go
// platform.go (no build tag - compiled everywhere)

package platform

import "path/filepath"

// AppDirs is used by all platforms
type AppDirs struct {
    Config string
    Data   string
    Cache  string
    Logs   string
}

// Common utility functions
func (d *AppDirs) ConfigFile(name string) string {
    return filepath.Join(d.Config, name)
}
```

#### Testing with Build Tags

```go
//go:build windows

package platform

import "testing"

func TestWindowsPaths(t *testing.T) {
    // Windows-specific tests
    dirs, err := GetAppDirs()
    if err != nil {
        t.Fatal(err)
    }

    if !strings.Contains(dirs.Config, "AppData") {
        t.Errorf("Expected AppData in path, got: %s", dirs.Config)
    }
}
```

```go
//go:build !windows

package platform

import "testing"

func TestUnixPaths(t *testing.T) {
    // Unix-specific tests
    dirs, err := GetAppDirs()
    if err != nil {
        t.Fatal(err)
    }

    // Test should pass on both macOS and Linux
    if !filepath.IsAbs(dirs.Config) {
        t.Errorf("Expected absolute path, got: %s", dirs.Config)
    }
}
```

#### CI Testing Strategy

**GitHub Actions Matrix**:
```yaml
name: Test

on: [push, pull_request]

jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        go: ['1.24']

    runs-on: ${{ matrix.os }}

    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: ${{ matrix.go }}

      - name: Test
        run: go test -v -race ./...

      - name: Test with coverage
        run: go test -v -race -coverprofile=coverage.out ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

This ensures:
- Windows tests run on Windows
- macOS tests run on macOS
- Linux tests run on Linux
- Build tags correctly exclude platform-specific code

#### Cross-Compilation

Build tags work with cross-compilation:

```bash
# Build for Windows from Linux
GOOS=windows GOARCH=amd64 go build -o lazynuget.exe ./cmd/lazynuget

# Build for macOS from Windows
GOOS=darwin GOARCH=amd64 go build -o lazynuget-mac ./cmd/lazynuget

# Build for Linux from macOS
GOOS=linux GOARCH=amd64 go build -o lazynuget-linux ./cmd/lazynuget
```

Only the code with matching build tags is compiled.

#### Advanced Build Tag Logic

```go
//go:build (linux || darwin) && !android
// Compile on Linux or macOS, but not Android

//go:build windows && amd64
// Compile only on 64-bit Windows

//go:build !race
// Exclude when testing with race detector

//go:build integration
// Only compile for integration tests (requires -tags=integration)
```

**Custom tags** for integration tests:
```go
//go:build integration

package integration

// Only compiled with: go test -tags=integration ./...
```

### Testing Strategy

**Unit Tests**:
1. Write platform-specific tests with matching build tags
2. Test interface compliance (all platforms implement same interface)
3. Test factory function returns correct type

**Integration Tests**:
1. Run full test suite on each platform in CI
2. Test cross-platform file handling (create on Windows, read on Linux)
3. Verify binary only contains code for target platform (size check)

**Build Verification**:
```bash
# Verify Windows code not in Linux binary
GOOS=linux go build ./...
# Check no Windows API calls in output

# Verify Linux code not in Windows binary
GOOS=windows go build ./...
# Check no Unix syscalls in output
```

**CI Matrix Testing**:
- 3 OSes × 2-3 Go versions = 6-9 test runs per commit
- Verify all tests pass on each platform
- Check code coverage per platform (should be similar)

---

## Summary and Cross-Cutting Concerns

### Integration Between Areas

The six research areas are interconnected:

1. **Windows Path Handling + Cross-Platform Directories**:
   - Directories returned by `GetAppDirs()` must be normalized with `NormalizePath()`
   - Long path support applies to config/cache/log directories
   - UNC paths must work for network shares (`\\server\share\lazynuget\config`)

2. **XDG + Cross-Platform Directories**:
   - Linux directory functions must respect XDG environment variables
   - Precedence: CLI flag > Env var > XDG > Default

3. **Terminal Capabilities + Process Output Encoding**:
   - Terminal color level affects how we display dotnet CLI output
   - If terminal doesn't support Unicode, transcode emoji/box-drawing to ASCII
   - Non-interactive mode (no TTY) requires different output encoding strategy

4. **Build Tags + All Areas**:
   - Each platform-specific feature uses build tags for clean separation
   - Shared interfaces ensure consistent API across platforms
   - Factory pattern (`platform.New()`) provides single entry point

### Recommended Implementation Order

1. **Build Tags & Directory Structure** (Foundation)
   - Set up `internal/platform/` with proper build tag structure
   - Define interfaces and factory function
   - Ensure clean separation before adding platform-specific code

2. **Cross-Platform Directories** (Core Feature)
   - Implement `GetAppDirs()` for each platform
   - Add precedence logic (CLI > Env > Default)
   - Test directory creation and permissions

3. **Windows Path Handling** (Windows Foundation)
   - Implement path normalization
   - Add long path support
   - Handle UNC paths
   - Integrate with `GetAppDirs()` on Windows

4. **XDG Base Directory** (Linux Foundation)
   - Implement XDG detection and fallbacks
   - Integrate with `GetAppDirs()` on Linux
   - Test environment variable overrides

5. **Terminal Capability Detection** (UX Foundation)
   - Implement TTY detection
   - Add color level detection
   - Handle Unicode detection
   - Set up SIGWINCH handling

6. **Process Output Encoding** (dotnet Integration)
   - Implement encoding detection (Windows code page, Unix locale)
   - Create `OutputDecoder` for transcoding
   - Integrate with dotnet CLI command execution
   - Test with non-UTF-8 locales

### Performance Considerations

**Caching**:
- Cache `GetAppDirs()` result (directories don't change at runtime)
- Cache terminal capabilities (detect once on startup)
- Cache system encoding (doesn't change during execution)

**Lazy Initialization**:
```go
type Platform struct {
    appDirs     *AppDirs
    appDirsOnce sync.Once

    termCaps     *TerminalCapabilities
    termCapsOnce sync.Once

    encoding     encoding.Encoding
    encodingOnce sync.Once
}

func (p *Platform) GetAppDirs() (*AppDirs, error) {
    var err error
    p.appDirsOnce.Do(func() {
        p.appDirs, err = detectAppDirs()
    })
    return p.appDirs, err
}
```

**Avoid Repeated Syscalls**:
- Don't call `GetConsoleOutputCP()` on every command execution
- Don't stat directories repeatedly (cache existence checks)
- Don't parse `LANG` on every encoding detection

### Error Handling Strategy

**Graceful Degradation**:
- If XDG detection fails → fall back to hardcoded defaults
- If terminal color detection fails → assume 16-color
- If encoding detection fails → assume UTF-8 (or Latin-1)
- If app directory creation fails → use temp directory

**Error Context**:
```go
if err := os.MkdirAll(path, 0700); err != nil {
    return fmt.Errorf("failed to create directory %s: %w", path, err)
}
```

Always include context (what operation, what path, original error).

**Logging**:
- Warn when using fallbacks (user should know)
- Log platform detection details (helpful for debugging)
- Don't fail on non-critical errors (e.g., can't set UTF-8 code page)

### Security Considerations

**Directory Permissions**:
- Create config/data/logs directories with `0700` (user-only) on Unix
- Windows: rely on NTFS ACLs (inherited from parent directory)

**Path Traversal**:
- Always validate paths before using them
- Use `filepath.Clean()` to remove `..` and `.` elements
- Never trust user input in paths without validation

**Sensitive Data**:
- Config files may contain NuGet API keys
- Never log file contents
- Set restrictive permissions by default

### Documentation Requirements

**User-Facing**:
- Document config file locations for each platform
- Explain environment variable overrides
- Provide troubleshooting guide for path issues

**Developer-Facing**:
- Document build tag conventions
- Explain platform abstraction interfaces
- Provide examples of adding new platform-specific features

**README.md Additions**:
```markdown
## Configuration Locations

LazyNuGet stores configuration in platform-appropriate locations:

- **Windows**: `%APPDATA%\lazynuget\config.yml`
- **macOS**: `~/Library/Application Support/lazynuget/config.yml`
- **Linux**: `~/.config/lazynuget/config.yml`

You can override with:
- `--config /path/to/config.yml` flag
- `LAZYNUGET_CONFIG=/path/to/dir` environment variable
```

### Testing Checklist

For each platform (Windows, macOS, Linux):

- [ ] Directories are created in correct locations
- [ ] Paths are properly normalized (separators, drive letters)
- [ ] Long paths work (>260 chars on Windows)
- [ ] UNC paths work (Windows)
- [ ] XDG variables are respected (Linux)
- [ ] Terminal color detection works
- [ ] Unicode detection works
- [ ] Terminal resize handling works
- [ ] Process output encoding is correct
- [ ] Non-UTF-8 locales are handled
- [ ] Environment variable overrides work
- [ ] CLI flag overrides work
- [ ] Fallback to temp directory works
- [ ] Error messages are helpful
- [ ] Build tags exclude incorrect code
- [ ] Tests pass in CI

### Maintenance Guidelines

**Adding New Platform-Specific Code**:
1. Define in interface (`Platform`)
2. Implement in `platform_windows.go`
3. Implement in `platform_unix.go` (or `platform_linux.go`, `platform_darwin.go`)
4. Add tests with matching build tags
5. Update CI matrix if needed

**Deprecating Old Platforms**:
1. Update build tags to exclude (e.g., `//go:build !oldplatform`)
2. Remove platform-specific files
3. Update documentation
4. Update CI matrix

**Handling New Platform Conventions**:
- Monitor XDG Base Directory spec for updates
- Watch Windows API changes (long path support, UTF-8 mode)
- Track macOS changes to Library structure

---

## References and Further Reading

### Official Documentation

1. **Go Build Tags**:
   - https://pkg.go.dev/cmd/go#hdr-Build_constraints
   - https://go.dev/blog/go1.17 (new //go:build syntax)

2. **XDG Base Directory Specification**:
   - https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html
   - https://wiki.archlinux.org/title/XDG_Base_Directory

3. **Windows Paths**:
   - https://learn.microsoft.com/en-us/windows/win32/fileio/maximum-file-path-limitation
   - https://learn.microsoft.com/en-us/windows/win32/fileio/naming-a-file

4. **Windows Code Pages**:
   - https://learn.microsoft.com/en-us/windows/win32/intl/code-page-identifiers
   - https://learn.microsoft.com/en-us/windows/console/console-code-pages

5. **Terminal Capabilities**:
   - https://no-color.org/ (NO_COLOR standard)
   - https://gist.github.com/XVilka/8346728 (true color detection)

### Go Packages

1. **filepath**: https://pkg.go.dev/path/filepath
2. **term**: https://pkg.go.dev/golang.org/x/term
3. **text/encoding**: https://pkg.go.dev/golang.org/x/text/encoding
4. **sys/windows**: https://pkg.go.dev/golang.org/x/sys/windows

### Related Projects

1. **lazygit**: Cross-platform TUI reference implementation
2. **lazydocker**: Similar approach to platform detection
3. **Bubble Tea**: TUI framework (handles some terminal detection internally)

### Best Practices

1. **Cross-Platform Go**:
   - https://go.dev/blog/ports
   - https://golang.org/doc/install/source#environment

2. **Security**:
   - OWASP Path Traversal: https://owasp.org/www-community/attacks/Path_Traversal
   - File Permissions Best Practices: https://www.redhat.com/sysadmin/linux-file-permissions-explained

---

## Conclusion

This research document provides comprehensive guidance for implementing cross-platform functionality in LazyNuGet. The key principles are:

1. **Follow Platform Conventions**: Use Windows `%APPDATA%`, macOS `~/Library`, Linux XDG
2. **Graceful Degradation**: Detect capabilities and fall back when unavailable
3. **Clean Architecture**: Use build tags and interfaces for maintainable platform separation
4. **Security First**: Proper permissions, path validation, encoding safety
5. **Testability**: CI matrix, platform-specific tests, integration testing

By following these decisions and implementations, LazyNuGet will provide a consistent, native experience on all supported platforms.
