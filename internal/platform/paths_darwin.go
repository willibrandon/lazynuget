//go:build darwin

package platform

import (
	"os"
	"path/filepath"
)

// getConfigDir returns the macOS config directory: ~/Library/Application Support/lazynuget
func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", &PathError{
			Op:   "ConfigDir",
			Path: "~",
			Err:  "failed to get home directory: " + err.Error(),
		}
	}

	return filepath.Join(homeDir, "Library", "Application Support", "lazynuget"), nil
}

// getCacheDir returns the macOS cache directory: ~/Library/Caches/lazynuget
func getCacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", &PathError{
			Op:   "CacheDir",
			Path: "~",
			Err:  "failed to get home directory: " + err.Error(),
		}
	}

	return filepath.Join(homeDir, "Library", "Caches", "lazynuget"), nil
}
