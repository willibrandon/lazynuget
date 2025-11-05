//go:build !windows

package platform

import (
	"path/filepath"
	"strings"
)

// normalize converts path to Unix-native format:
// - Removes redundant separators
// - Resolves . and .. segments
// - Uses forward slashes
func normalize(path string) string {
	// Use filepath.Clean to handle . and .. and normalize separators
	return filepath.Clean(path)
}

// isAbsolute returns true if the path is absolute on Unix:
// - Starts with / (root)
func isAbsolute(path string) bool {
	return filepath.IsAbs(path)
}

// validate checks if path is valid on Unix
func validate(path string) error {
	// Basic validation: path cannot be empty
	if path == "" {
		return &PathError{
			Op:   "Validate",
			Path: path,
			Err:  "path cannot be empty",
		}
	}

	// Check for null bytes (invalid in Unix paths)
	if strings.ContainsRune(path, '\x00') {
		return &PathError{
			Op:   "Validate",
			Path: path,
			Err:  "path contains null byte",
		}
	}

	// Unix paths are very permissive - almost any character is valid
	// The main restriction is the null byte, which we already checked
	return nil
}
