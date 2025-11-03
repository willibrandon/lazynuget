package integration

import (
	"os/exec"
	"testing"
)

func TestBasicStartup(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../lazynuget-test", "../../cmd/lazynuget/main.go")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer exec.Command("rm", "../../lazynuget-test").Run()

	// Run with no arguments (should exit immediately for now since no TUI yet)
	cmd := exec.Command("../../lazynuget-test")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	// Verify exit code is 0 (successful initialization)
	if cmd.ProcessState.ExitCode() != 0 {
		t.Errorf("Expected exit code 0, got: %d\nOutput: %s", cmd.ProcessState.ExitCode(), output)
	}
}
