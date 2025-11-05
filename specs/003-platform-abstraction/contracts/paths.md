# Contract: Path Resolution

## Interface: PathResolver

Handles platform-specific path operations including normalization, validation, and directory resolution.

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

## Methods

### ConfigDir() (string, error)

Returns the platform-appropriate configuration directory.

**Platform-specific paths**:
- **Windows**: `%APPDATA%\lazynuget` (e.g., `C:\Users\Name\AppData\Roaming\lazynuget`)
- **macOS**: `~/Library/Application Support/lazynuget`
- **Linux**: `$XDG_CONFIG_HOME/lazynuget` or `~/.config/lazynuget` (if XDG var unset)

**Behavior**:
- Creates directory if it doesn't exist (with parent validation per FR-025)
- Logs warning if directory created
- Returns error if config directory is read-only (per FR-027)

**Error cases**:
- Environment variables not set and fallback unavailable
- Directory creation fails (permissions, read-only filesystem)
- Parent directory doesn't exist and can't be created

### CacheDir() (string, error)

Returns the platform-appropriate cache directory.

**Platform-specific paths**:
- **Windows**: `%LOCALAPPDATA%\lazynuget` (e.g., `C:\Users\Name\AppData\Local\lazynuget`)
- **macOS**: `~/Library/Caches/lazynuget`
- **Linux**: `$XDG_CACHE_HOME/lazynuget` or `~/.cache/lazynuget` (if XDG var unset)

**Behavior**:
- Creates directory if it doesn't exist (with parent validation)
- Gracefully degrades if read-only (per FR-026) - logs warning but doesn't error
- Cache operations will fail later but application continues

**Note**: Unlike ConfigDir(), cache directory failures are non-fatal.

### Normalize(path string) string

Converts path to platform-native format.

**Windows normalization**:
- Convert `/` to `\`
- Uppercase drive letters: `c:\path` → `C:\path`
- Preserve UNC paths: `\\server\share` unchanged
- Remove redundant separators: `C:\\\\path` → `C:\path`
- Resolve `.` and `..`: `C:\foo\.\bar\..\baz` → `C:\foo\baz`

**Unix normalization**:
- Forward slashes only (already standard)
- Remove redundant separators: `//home//user` → `/home/user`
- Resolve `.` and `..`: `/foo/./bar/../baz` → `/foo/baz`

**Example**:
```go
// Windows
resolver.Normalize("C:/Users/Dev/config.yml")
// Returns: "C:\Users\Dev\config.yml"

// Unix
resolver.Normalize("/home//user/./config.yml")
// Returns: "/home/user/config.yml"
```

### Validate(path string) error

Checks if path format is valid for the current platform.

**Validation rules**:
- **Windows**: Must have drive letter (`C:\`) or UNC prefix (`\\server\share`)
- **Unix**: Absolute paths must start with `/`, relative with `./` or `../`
- **All platforms**: No null bytes, max 500 characters (performance target)

**Error messages**:
```go
// Windows
Validate("\\server\share") // OK (UNC)
Validate("/unix/path")     // Error: "UNC paths not supported on Unix. Use absolute path instead: /mnt/share"

// Unix
Validate("/home/user")     // OK (absolute)
Validate("./config.yml")   // OK (relative)
Validate("C:\path")        // Error: "Windows drive letters not supported on Unix"
```

### IsAbsolute(path string) bool

Returns true if path is absolute for the current platform.

**Platform-specific rules**:
- **Windows**: Starts with drive letter (`C:\`) or UNC (`\\server\share`)
- **Unix**: Starts with `/`

**Examples**:
```go
// Windows
IsAbsolute("C:\Users\Dev")    // true
IsAbsolute("\\server\share")  // true
IsAbsolute("relative\path")   // false

// Unix
IsAbsolute("/home/user")      // true
IsAbsolute("./relative")      // false
```

### Resolve(path string) (string, error)

Makes relative paths absolute by resolving them relative to the config directory. Absolute paths are returned unchanged.

**Behavior**:
```go
// If ConfigDir() returns "/home/user/.config/lazynuget"

Resolve("config.yml")         // "/home/user/.config/lazynuget/config.yml"
Resolve("./settings/theme")   // "/home/user/.config/lazynuget/settings/theme"
Resolve("/absolute/path")     // "/absolute/path" (unchanged)
```

**Error cases**:
- ConfigDir() fails
- Path normalization fails
- Invalid path format

## Implementation Notes

### Environment Variable Precedence (per FR-029)

On Windows, Windows-native environment variables ALWAYS take precedence over XDG variables:

```
Windows: APPDATA/LOCALAPPDATA > XDG_CONFIG_HOME/XDG_CACHE_HOME
macOS:   ~/Library only (no XDG support)
Linux:   XDG vars > ~/.config/~/.cache defaults
```

### Directory Creation Strategy

Per clarifications from spec:
1. Check if directory exists
2. If not, check if parent exists
3. If parent exists → create with `os.MkdirAll`, log warning
4. If parent missing → fall back to platform defaults
5. Cache failures are warnings; config failures are errors

### Performance Considerations

- `Normalize()` is in hot path (<1ms target, zero allocations)
- `ConfigDir()` and `CacheDir()` are cached after first call
- Path validation uses fast string prefix checks, not regex

## Example Usage

```go
import (
    "path/filepath"
    "github.com/yourusername/lazynuget/internal/platform"
)

func loadConfig() error {
    resolver := platform.NewPathResolver()

    // Get platform-appropriate config directory
    configDir, err := resolver.ConfigDir()
    if err != nil {
        return fmt.Errorf("failed to get config dir: %w", err)
    }

    // Combine with filename
    configPath := filepath.Join(configDir, "config.yml")

    // Normalize to platform format
    configPath = resolver.Normalize(configPath)

    // Validate before use
    if err := resolver.Validate(configPath); err != nil {
        return fmt.Errorf("invalid config path: %w", err)
    }

    // Load config file
    return loadConfigFile(configPath)
}
```
