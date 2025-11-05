# Contract: Platform Detection

## Interface: PlatformInfo

Provides operating system and architecture detection.

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

## Methods

### OS() string

Returns the operating system identifier.

**Possible values**:
- `"windows"` - Microsoft Windows
- `"darwin"` - Apple macOS
- `"linux"` - Linux distributions

**Usage**:
```go
p := platform.New()
if p.OS() == "windows" {
    // Windows-specific logic
}
```

### Arch() string

Returns the processor architecture.

**Possible values**:
- `"amd64"` - 64-bit x86 (Intel/AMD)
- `"arm64"` - 64-bit ARM (Apple Silicon, ARM servers)

**Usage**:
```go
p := platform.New()
fmt.Printf("Running on %s/%s\n", p.OS(), p.Arch())
// Output: Running on darwin/arm64
```

### Version() string

Returns the OS version string for diagnostics.

**Format**:
- Windows: `"Windows 10.0.19045"`
- macOS: `"macOS 14.1"`
- Linux: `"Linux 5.15.0"` (kernel version)

**Note**: May be empty if version detection fails. This is informational only.

### IsWindows() bool

Convenience method returning `true` if running on Windows.

Equivalent to: `p.OS() == "windows"`

### IsDarwin() bool

Convenience method returning `true` if running on macOS.

Equivalent to: `p.OS() == "darwin"`

### IsLinux() bool

Convenience method returning `true` if running on Linux.

Equivalent to: `p.OS() == "linux"`

## Implementation Notes

- Platform detection is performed once at startup and cached
- The platform instance is immutable after creation (singleton pattern)
- Unsupported OS/arch combinations will cause initialization to fail with clear error
- Detection uses Go's `runtime.GOOS` and `runtime.GOARCH` constants

## Example Usage

```go
import "github.com/yourusername/lazynuget/internal/platform"

func main() {
    p := platform.New()

    // Check OS
    if p.IsWindows() {
        fmt.Println("Using Windows-specific paths")
    } else if p.IsDarwin() {
        fmt.Println("Using macOS ~/Library conventions")
    } else if p.IsLinux() {
        fmt.Println("Using XDG Base Directory Specification")
    }

    // Get detailed info
    fmt.Printf("Platform: %s/%s (version: %s)\n",
        p.OS(), p.Arch(), p.Version())
}
```
