package integration

import (
	"os"
	"os/exec"
	"runtime"
	"testing"
)

const (
	testBinaryBase   = "../../lazynuget-test"
	testBinarySource = "../../cmd/lazynuget/main.go"
)

// buildTestBinary builds the lazynuget binary for testing with correct platform extension.
// Returns the binary path.
func buildTestBinary(t *testing.T) string {
	t.Helper()

	binaryPath := testBinaryBase + getPlatformExt()

	// Build with literal paths to satisfy gosec G204 security check
	var buildCmd *exec.Cmd
	if runtime.GOOS == "windows" {
		buildCmd = exec.Command("go", "build", "-o", testBinaryBase+".exe", testBinarySource)
	} else {
		buildCmd = exec.Command("go", "build", "-o", testBinaryBase, testBinarySource)
	}

	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	return binaryPath
}

// cleanupBinary removes the test binary. Cleanup errors are ignored as this is best-effort.
func cleanupBinary(binaryPath string) {
	err := os.Remove(binaryPath)
	// Explicitly check error for linter compliance, but cleanup is best-effort in tests
	if err == nil {
		return // Successfully cleaned up
	}
	// Error occurred (file doesn't exist, permission denied, etc.)
	// but we don't fail tests for cleanup issues - this is intentional
}

// getPlatformExt returns the executable extension for the current platform
func getPlatformExt() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}
