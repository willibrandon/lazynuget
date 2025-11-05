package platform

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
