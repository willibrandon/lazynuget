//go:build windows

package platform

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// windowsExecutableExtensions are the extensions to try when resolving executables on Windows
// See: T087, FR-031
var windowsExecutableExtensions = []string{".exe", ".cmd", ".bat", ".com"}

// resolveExecutablePlatform performs Windows-specific executable resolution
// Tries .exe, .cmd, .bat extensions automatically
// See: T087, FR-031
func resolveExecutablePlatform(executable string) (string, error) {
	// If it already has an extension, try it as-is first
	if filepath.Ext(executable) != "" {
		if path, err := exec.LookPath(executable); err == nil {
			return path, nil
		}
	}

	// If absolute path, try with extensions
	if filepath.IsAbs(executable) {
		return tryWindowsExtensions(executable)
	}

	// Search in PATH with extensions
	for _, ext := range windowsExecutableExtensions {
		testName := executable
		if !strings.HasSuffix(strings.ToLower(executable), ext) {
			testName = executable + ext
		}

		if path, err := exec.LookPath(testName); err == nil {
			return path, nil
		}
	}

	// Last resort: try the original name
	return exec.LookPath(executable)
}

// tryWindowsExtensions tries adding Windows executable extensions
// See: T087
func tryWindowsExtensions(basePath string) (string, error) {
	// Try the path as-is first
	if _, err := os.Stat(basePath); err == nil {
		return basePath, nil
	}

	// Try with each extension
	for _, ext := range windowsExecutableExtensions {
		testPath := basePath
		if !strings.HasSuffix(strings.ToLower(basePath), ext) {
			testPath = basePath + ext
		}

		if _, err := os.Stat(testPath); err == nil {
			return testPath, nil
		}
	}

	// None found
	return "", os.ErrNotExist
}

// Note: Argument quoting is handled automatically by exec.Command
// Go's exec package doesn't invoke a shell, so arguments are passed directly
// to the process without needing manual quoting. The functions quoteArgument
// and needsQuoting from T089 are not implemented as they're unnecessary.
