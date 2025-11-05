package platform

import (
	"fmt"
	"os"
	"path/filepath"
)

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

	// EnsureDir creates the directory if it doesn't exist, with appropriate permissions
	EnsureDir(path string) error
}

// pathResolver implements PathResolver interface
type pathResolver struct {
	platform PlatformInfo
}

// NewPathResolver creates a new PathResolver instance
func NewPathResolver(platform PlatformInfo) (PathResolver, error) {
	if platform == nil {
		return nil, fmt.Errorf("platform cannot be nil")
	}

	return &pathResolver{
		platform: platform,
	}, nil
}

// ConfigDir returns the platform-appropriate configuration directory
func (p *pathResolver) ConfigDir() (string, error) {
	return getConfigDir()
}

// CacheDir returns the platform-appropriate cache directory
func (p *pathResolver) CacheDir() (string, error) {
	return getCacheDir()
}

// Normalize converts path to platform-native format
func (p *pathResolver) Normalize(path string) string {
	// Use filepath.Clean to normalize path separators and remove redundancies
	return filepath.Clean(path)
}

// Validate checks if path format is valid for current platform
func (p *pathResolver) Validate(path string) error {
	// Basic validation: path cannot be empty
	if path == "" {
		return &PathError{
			Op:   "Validate",
			Path: path,
			Err:  "path cannot be empty",
		}
	}

	// More detailed validation will be added in Phase 4 (User Story 2)
	return nil
}

// IsAbsolute returns true if path is absolute for current platform
func (p *pathResolver) IsAbsolute(path string) bool {
	return filepath.IsAbs(path)
}

// Resolve makes relative path absolute relative to config directory
func (p *pathResolver) Resolve(path string) (string, error) {
	// If already absolute, return as-is
	if p.IsAbsolute(path) {
		return path, nil
	}

	// Get config directory
	configDir, err := p.ConfigDir()
	if err != nil {
		return "", err
	}

	// Join with config directory
	return filepath.Join(configDir, path), nil
}

// EnsureDir creates the directory if it doesn't exist
// Uses 0o700 permissions (owner-only) for security
func (p *pathResolver) EnsureDir(path string) error {
	// Check if directory exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create directory with owner-only permissions
			if mkdirErr := os.MkdirAll(path, 0o700); mkdirErr != nil {
				return &PathError{
					Op:   "EnsureDir",
					Path: path,
					Err:  "failed to create directory: " + mkdirErr.Error(),
				}
			}
			return nil
		}
		// Some other error occurred
		return &PathError{
			Op:   "EnsureDir",
			Path: path,
			Err:  "failed to stat directory: " + err.Error(),
		}
	}

	// Path exists, verify it's a directory
	if !info.IsDir() {
		return &PathError{
			Op:   "EnsureDir",
			Path: path,
			Err:  "path exists but is not a directory",
		}
	}

	return nil
}

// PathError represents a path operation error
type PathError struct {
	Op   string // Operation that failed (e.g., "ConfigDir", "Validate")
	Path string // Path that caused the error
	Err  string // Error description
}

// Error implements the error interface
func (e *PathError) Error() string {
	return fmt.Sprintf("%s %q: %s", e.Op, e.Path, e.Err)
}
