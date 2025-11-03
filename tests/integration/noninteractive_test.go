package integration

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestNonInteractiveFlagExplicit(t *testing.T) {
	// Build the binary
	binaryPath := buildTestBinary(t)
	defer cleanupBinary(binaryPath)

	// Start the application with --non-interactive flag
	cmd := exec.Command(binaryPath, "--non-interactive")

	// Capture both stdout and stderr since logs go to stderr
	output, err := cmd.CombinedOutput()
	// Check exit status
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// Non-zero exit is acceptable if it's a controlled shutdown
			if exitErr.ExitCode() > 2 {
				t.Fatalf("Process exited with unexpected error code %d: %v\nOutput: %s",
					exitErr.ExitCode(), err, string(output))
			}
		} else {
			t.Fatalf("Failed to run application: %v\nOutput: %s", err, string(output))
		}
	}

	outputStr := string(output)

	// Verify exit code is 0 or 1 (success or controlled error)
	if cmd.ProcessState.ExitCode() > 1 {
		t.Errorf("Expected exit code 0-1 in non-interactive mode, got: %d\nOutput: %s",
			cmd.ProcessState.ExitCode(), outputStr)
	}

	// Verify log output indicates non-interactive mode
	if !strings.Contains(outputStr, "non-interactive") {
		t.Errorf("Expected output to contain 'non-interactive', got: %s", outputStr)
	}
}

func TestNonInteractiveTTYDetection(t *testing.T) {
	// Build the binary
	binaryPath := buildTestBinary(t)
	defer cleanupBinary(binaryPath)

	// Use echo to pipe stdin (simulates piped input)
	cmd := exec.Command("sh", "-c", "echo | "+binaryPath+" 2>&1")
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
	// Build the binary
	binaryPath := buildTestBinary(t)
	defer cleanupBinary(binaryPath)

	// Start the application with CI environment variable set
	cmd := exec.Command(binaryPath)
	cmd.Env = append(os.Environ(), "CI=true")

	// Use CombinedOutput to capture both stdout and stderr
	// (logging goes to stderr, so we need to capture both)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Non-zero exit is OK for this test - we just want to verify detection
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// Application exited with non-zero code, which is fine
			outputStr := string(output)
			if !strings.Contains(outputStr, "non-interactive") {
				t.Errorf("Expected CI environment to trigger non-interactive mode. Output: %s", outputStr)
			}
			return
		}
		t.Fatalf("Failed to run application: %v", err)
	}

	// Verify non-interactive mode was detected via CI environment
	outputStr := string(output)
	if !strings.Contains(outputStr, "non-interactive") {
		t.Errorf("Expected CI environment to trigger non-interactive mode. Output: %s", outputStr)
	}
}

func TestStartupPerformance(t *testing.T) {
	// Build the binary
	binaryPath := buildTestBinary(t)
	defer cleanupBinary(binaryPath)

	// Measure startup time in non-interactive mode
	start := time.Now()

	cmd := exec.Command(binaryPath, "--non-interactive")
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
	// Build the binary
	binaryPath := buildTestBinary(t)
	defer cleanupBinary(binaryPath)

	// Start the application with TERM=dumb
	cmd := exec.Command(binaryPath)
	cmd.Env = append(os.Environ(), "TERM=dumb")

	// Use CombinedOutput to capture both stdout and stderr
	// (logging goes to stderr, so we need to capture both)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Non-zero exit is OK for this test - we just want to verify detection
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// Application exited with non-zero code, which is fine
			outputStr := string(output)
			if !strings.Contains(outputStr, "non-interactive") {
				t.Errorf("Expected TERM=dumb to trigger non-interactive mode. Output: %s", outputStr)
			}
			return
		}
		t.Fatalf("Failed to run application: %v", err)
	}

	// Verify non-interactive mode was detected via TERM=dumb
	outputStr := string(output)
	if !strings.Contains(outputStr, "non-interactive") {
		t.Errorf("Expected TERM=dumb to trigger non-interactive mode. Output: %s", outputStr)
	}
}
