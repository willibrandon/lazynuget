package integration

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestNonInteractiveFlagExplicit(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../lazynuget-test", "../../cmd/lazynuget/main.go")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer exec.Command("rm", "../../lazynuget-test").Run()

	// Start the application with --non-interactive flag
	cmd := exec.Command("../../lazynuget-test", "--non-interactive")
	outputPipe, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to get stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start application: %v", err)
	}

	// Read output
	output := make([]byte, 8192)
	n, _ := outputPipe.Read(output)
	outputStr := string(output[:n])

	// Wait for process to complete (should exit immediately)
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Process exited with error: %v", err)
		}

		// Verify exit code is 0
		if cmd.ProcessState.ExitCode() != 0 {
			t.Errorf("Expected exit code 0 in non-interactive mode, got: %d", cmd.ProcessState.ExitCode())
		}

		// Verify log output indicates non-interactive mode
		if !strings.Contains(outputStr, "non-interactive") {
			t.Errorf("Expected output to contain 'non-interactive', got: %s", outputStr)
		}

	case <-time.After(3 * time.Second):
		// Force kill if still running
		cmd.Process.Kill()
		t.Fatal("Application did not exit promptly in non-interactive mode")
	}
}

func TestNonInteractiveTTYDetection(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../lazynuget-test", "../../cmd/lazynuget/main.go")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer exec.Command("rm", "../../lazynuget-test").Run()

	// Use echo to pipe stdin (simulates piped input)
	cmd := exec.Command("sh", "-c", "echo | ../../lazynuget-test 2>&1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Non-zero exit is acceptable if it's a controlled shutdown
		if cmd.ProcessState.ExitCode() > 2 {
			t.Fatalf("Process exited with unexpected error: %v\nOutput: %s", err, string(output))
		}
	}

	outputStr := string(output)

	// Verify TTY detection triggered non-interactive mode
	if !strings.Contains(outputStr, "non-interactive") {
		t.Errorf("Expected TTY detection to trigger non-interactive mode. Output: %s", outputStr)
	}

	// Verify no TUI was attempted
	if strings.Contains(strings.ToLower(outputStr), "gui") || strings.Contains(strings.ToLower(outputStr), "tui") {
		t.Error("Expected no GUI/TUI initialization in piped context")
	}
}

func TestCIEnvironmentDetection(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../lazynuget-test", "../../cmd/lazynuget/main.go")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer exec.Command("rm", "../../lazynuget-test").Run()

	// Start the application with CI environment variable set
	cmd := exec.Command("../../lazynuget-test")
	cmd.Env = append(os.Environ(), "CI=true")

	outputPipe, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to get stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start application: %v", err)
	}

	// Read output
	output := make([]byte, 8192)
	n, _ := outputPipe.Read(output)
	outputStr := string(output[:n])

	// Wait for process to complete
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-done:
		// Verify non-interactive mode was detected via CI environment
		if !strings.Contains(outputStr, "non-interactive") {
			t.Errorf("Expected CI environment to trigger non-interactive mode. Output: %s", outputStr)
		}

	case <-time.After(3 * time.Second):
		cmd.Process.Kill()
		t.Fatal("Application did not exit promptly with CI environment set")
	}
}

func TestStartupPerformance(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../lazynuget-test", "../../cmd/lazynuget/main.go")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer exec.Command("rm", "../../lazynuget-test").Run()

	// Measure startup time in non-interactive mode
	start := time.Now()

	cmd := exec.Command("../../lazynuget-test", "--non-interactive")
	if err := cmd.Run(); err != nil {
		// Accept exit codes 0-2 (success or controlled errors)
		if cmd.ProcessState.ExitCode() > 2 {
			t.Fatalf("Application failed to start: %v", err)
		}
	}

	elapsed := time.Since(start)

	// Verify startup time is under 200ms (as per requirements)
	if elapsed > 200*time.Millisecond {
		t.Logf("WARNING: Startup time %v exceeds 200ms target", elapsed)
		// Don't fail the test - this is a performance guideline, not a hard requirement
		// Hardware differences can affect timing
	} else {
		t.Logf("Startup time: %v (within 200ms target)", elapsed)
	}

	// Verify startup time is reasonable (under 1 second - hard limit)
	if elapsed > 1*time.Second {
		t.Errorf("Startup time %v exceeds 1 second hard limit", elapsed)
	}
}

func TestDumbTerminalDetection(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../lazynuget-test", "../../cmd/lazynuget/main.go")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer exec.Command("rm", "../../lazynuget-test").Run()

	// Start the application with TERM=dumb
	cmd := exec.Command("../../lazynuget-test")
	cmd.Env = append(os.Environ(), "TERM=dumb")

	outputPipe, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to get stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start application: %v", err)
	}

	// Read output
	output := make([]byte, 8192)
	n, _ := outputPipe.Read(output)
	outputStr := string(output[:n])

	// Wait for process to complete
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-done:
		// Verify non-interactive mode was detected via TERM=dumb
		if !strings.Contains(outputStr, "non-interactive") {
			t.Errorf("Expected TERM=dumb to trigger non-interactive mode. Output: %s", outputStr)
		}

	case <-time.After(3 * time.Second):
		cmd.Process.Kill()
		t.Fatal("Application did not exit promptly with TERM=dumb")
	}
}
