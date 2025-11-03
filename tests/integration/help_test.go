package integration

import (
	"os/exec"
	"strings"
	"testing"
)

func TestHelpFlag(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../lazynuget-test", "../../cmd/lazynuget/main.go")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer exec.Command("rm", "../../lazynuget-test").Run()

	// Run with --help flag
	cmd := exec.Command("../../lazynuget-test", "--help")
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
