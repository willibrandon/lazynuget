//go:build linux

package platform

import (
	"os"
	"path/filepath"
)

// getConfigDir returns the Linux config directory following XDG Base Directory Specification
// Returns $XDG_CONFIG_HOME/lazynuget or ~/.config/lazynuget
func getConfigDir() (string, error) {
	// Check XDG_CONFIG_HOME first
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "lazynuget"), nil
	}

	// Fall back to ~/.config
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", &PathError{
			Op:   "ConfigDir",
			Path: "~",
			Err:  "failed to get home directory: " + err.Error(),
		}
	}

	return filepath.Join(homeDir, ".config", "lazynuget"), nil
}

// getCacheDir returns the Linux cache directory following XDG Base Directory Specification
// Returns $XDG_CACHE_HOME/lazynuget or ~/.cache/lazynuget
func getCacheDir() (string, error) {
	// Check XDG_CACHE_HOME first
	if xdgCache := os.Getenv("XDG_CACHE_HOME"); xdgCache != "" {
		return filepath.Join(xdgCache, "lazynuget"), nil
	}

	// Fall back to ~/.cache
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", &PathError{
			Op:   "CacheDir",
			Path: "~",
			Err:  "failed to get home directory: " + err.Error(),
		}
	}

	return filepath.Join(homeDir, ".cache", "lazynuget"), nil
}
