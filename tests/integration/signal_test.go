package integration

import (
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestSIGINTHandling(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../lazynuget-test", "../../cmd/lazynuget/main.go")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer exec.Command("rm", "../../lazynuget-test").Run()

	// Start the application
	cmd := exec.Command("../../lazynuget-test")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start application: %v", err)
	}

	// Give app time to start
	time.Sleep(500 * time.Millisecond)

	// Send SIGINT (Ctrl+C)
	if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
		t.Fatalf("Failed to send SIGINT: %v", err)
	}

	// Wait for process to exit with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		// Process exited successfully
		if err != nil {
			t.Errorf("Process exited with error: %v", err)
		}

		// Verify exit code is 0 (graceful shutdown)
		if cmd.ProcessState.ExitCode() != 0 {
			t.Errorf("Expected exit code 0 after SIGINT, got: %d", cmd.ProcessState.ExitCode())
		}
	case <-time.After(5 * time.Second):
		// Force kill if still running
		cmd.Process.Kill()
		t.Fatal("Application did not shutdown within timeout after SIGINT")
	}
}

func TestSIGTERMHandling(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../lazynuget-test", "../../cmd/lazynuget/main.go")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer exec.Command("rm", "../../lazynuget-test").Run()

	// Start the application
	cmd := exec.Command("../../lazynuget-test")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start application: %v", err)
	}

	// Give app time to start
	time.Sleep(500 * time.Millisecond)

	// Send SIGTERM
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("Failed to send SIGTERM: %v", err)
	}

	// Wait for process to exit with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		// Process exited successfully
		if err != nil {
			t.Errorf("Process exited with error: %v", err)
		}

		// Verify exit code is 0 (graceful shutdown)
		if cmd.ProcessState.ExitCode() != 0 {
			t.Errorf("Expected exit code 0 after SIGTERM, got: %d", cmd.ProcessState.ExitCode())
		}
	case <-time.After(5 * time.Second):
		// Force kill if still running
		cmd.Process.Kill()
		t.Fatal("Application did not shutdown within timeout after SIGTERM")
	}
}

func TestMultipleSignals(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../lazynuget-test", "../../cmd/lazynuget/main.go")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer exec.Command("rm", "../../lazynuget-test").Run()

	// Start the application
	cmd := exec.Command("../../lazynuget-test")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start application: %v", err)
	}

	// Give app time to start
	time.Sleep(500 * time.Millisecond)

	// Send multiple signals rapidly
	for i := 0; i < 3; i++ {
		if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
			// Process might have already exited
			if !strings.Contains(err.Error(), "process already finished") {
				t.Logf("Signal %d failed: %v", i+1, err)
			}
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Wait for process to exit with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-done:
		// Process exited (error is expected for signals)
		if cmd.ProcessState.ExitCode() != 0 {
			t.Logf("Process exited with code: %d (multiple signals test)", cmd.ProcessState.ExitCode())
		}
	case <-time.After(5 * time.Second):
		// Force kill if still running
		cmd.Process.Kill()
		t.Fatal("Application did not shutdown within timeout after multiple signals")
	}
}

func TestShutdownLogsPresent(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../lazynuget-test", "../../cmd/lazynuget/main.go")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer exec.Command("rm", "../../lazynuget-test").Run()

	// Start the application and capture output
	cmd := exec.Command("../../lazynuget-test")
	outputPipe, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to get stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start application: %v", err)
	}

	// Give app time to start and log
	time.Sleep(500 * time.Millisecond)

	// Send SIGINT to trigger shutdown
	if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
		t.Fatalf("Failed to send SIGINT: %v", err)
	}

	// Read all output
	output := make([]byte, 8192)
	n, _ := outputPipe.Read(output)
	outputStr := string(output[:n])

	// Wait for process to complete
	cmd.Wait()

	// Verify startup log messages are present
	// Note: In non-interactive mode, the app exits after bootstrap
	expectedLogs := []string{
		"Bootstrap complete",
		"Run mode determined",
	}

	for _, expected := range expectedLogs {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Expected log output to contain '%s', got: %s", expected, outputStr)
		}
	}

	// Verify log format is correct (slog text format)
	if !strings.Contains(outputStr, "level=INFO") {
		t.Error("Expected slog text format with level=INFO")
	}
}
