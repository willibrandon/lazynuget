package platform

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
