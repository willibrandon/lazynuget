package integration

import (
	"os/exec"
	"strings"
	"testing"
)

func TestHelpFlag(t *testing.T) {
	// Build the binary
	binaryPath := buildTestBinary(t)
	defer cleanupBinary(binaryPath)

	// Run with --help flag
	cmd := exec.Command(binaryPath, "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// Validate help text contains required sections
	requiredStrings := []string{
		"LazyNuGet",
		"Usage:",
		"Options:",
		"--version",
		"--help",
		"--config",
		"--log-level",
		"--non-interactive",
	}

	for _, required := range requiredStrings {
		if !strings.Contains(outputStr, required) {
			t.Errorf("Expected help output to contain '%s', got: %s", required, outputStr)
		}
	}

	// Verify exit code is 0
	if cmd.ProcessState.ExitCode() != 0 {
		t.Errorf("Expected exit code 0, got: %d", cmd.ProcessState.ExitCode())
	}
}
