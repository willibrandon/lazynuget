package integration

import (
	"os/exec"
	"strings"
	"testing"
)

func TestVersionFlag(t *testing.T) {
	// Build the binary
	binaryPath := buildTestBinary(t)
	defer cleanupBinary(binaryPath)

	// Run with --version flag
	cmd := exec.Command(binaryPath, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// Validate output format
	if !strings.Contains(outputStr, "LazyNuGet version") {
		t.Errorf("Expected version output to contain 'LazyNuGet version', got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "built on") {
		t.Errorf("Expected version output to contain 'built on', got: %s", outputStr)
	}

	// Verify exit code is 0
	if cmd.ProcessState.ExitCode() != 0 {
		t.Errorf("Expected exit code 0, got: %d", cmd.ProcessState.ExitCode())
	}
}
