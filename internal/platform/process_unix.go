//go:build !windows

package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// resolveExecutablePlatform performs Unix-specific executable resolution
// Checks execute permissions
// See: T088, FR-031
func resolveExecutablePlatform(executable string) (string, error) {
	// If absolute path, validate it exists and is executable
	if filepath.IsAbs(executable) {
		return validateExecutablePermissions(executable)
	}

	// Search in PATH
	path, err := exec.LookPath(executable)
	if err != nil {
		return "", fmt.Errorf("executable not found in PATH: %s", executable)
	}

	// Validate permissions
	return validateExecutablePermissions(path)
}

// validateExecutablePermissions checks if a file exists and has execute permissions
// See: T088, FR-031
func validateExecutablePermissions(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("executable not found: %s", path)
	}

	// Check if it's a regular file (not a directory)
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("not a regular file: %s", path)
	}

	// Check execute permissions
	// On Unix, check if any execute bit is set (owner, group, or other)
	if info.Mode().Perm()&0o111 == 0 {
		return "", fmt.Errorf("permission denied: %s (not executable)", path)
	}

	return path, nil
}

// Note: Argument quoting is handled automatically by exec.Command
// Go's exec package doesn't invoke a shell, so arguments are passed directly
// to the process without needing manual quoting. The functions quoteArgument
// and needsQuoting from T090 are not implemented as they're unnecessary.
