package integration

import (
	"os/exec"
	"syscall"
	"testing"
	"time"
)

func TestBasicStartup(t *testing.T) {
	// Build the binary
	binaryPath := buildTestBinary(t)
	defer cleanupBinary(binaryPath)

	// Start the application (it will now block waiting for signals)
	cmd := exec.Command(binaryPath)
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start application: %v", err)
	}

	// Give app time to start
	time.Sleep(500 * time.Millisecond)

	// Send SIGTERM to trigger graceful shutdown
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("Failed to send SIGTERM: %v", err)
	}

	// Wait for process to exit
	err := cmd.Wait()
	if err != nil {
		t.Errorf("Process exited with error: %v", err)
	}

	// Verify exit code is 0 (successful initialization and shutdown)
	if cmd.ProcessState.ExitCode() != 0 {
		t.Errorf("Expected exit code 0, got: %d", cmd.ProcessState.ExitCode())
	}
}
