package integration

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestInvalidYAMLError(t *testing.T) {
	// Create a temporary directory for config
	tmpDir := t.TempDir()

	// Write invalid YAML to config file
	invalidYAML := `
logLevel: debug
	invalid indentation here
compactMode true  # missing colon
`
	configPath := filepath.Join(tmpDir, "config.yml")
	if err := os.WriteFile(configPath, []byte(invalidYAML), 0o644); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	// Build the binary
	binaryPath := buildTestBinary(t)
	defer cleanupBinary(binaryPath)

	// Run with invalid config
	cmd := exec.Command(binaryPath, "--config", configPath, "--non-interactive")
	output, err := cmd.CombinedOutput()

	outputStr := string(output)

	// Should exit with error
	if err == nil {
		t.Fatalf("Expected error for invalid YAML but got none. Output: %s", outputStr)
	}

	// Verify error message is helpful
	expectedPhrases := []string{
		"failed to parse YAML",
		"syntax errors",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(strings.ToLower(outputStr), strings.ToLower(phrase)) {
			t.Errorf("Expected error message to contain '%s', got: %s", phrase, outputStr)
		}
	}

	// Verify exit code indicates user error
	if cmd.ProcessState.ExitCode() != 1 {
		t.Errorf("Expected exit code 1 for config error, got: %d", cmd.ProcessState.ExitCode())
	}
}

func TestMissingDotnetCLI(t *testing.T) {
	// Build the binary
	binaryPath := buildTestBinary(t)
	defer cleanupBinary(binaryPath)

	// Create a temporary directory with NO dotnet executable
	tmpDir := t.TempDir()
	// PATH with only temp dir - excludes /usr/bin and /usr/local/bin where dotnet lives
	fakePath := tmpDir

	// Run without --non-interactive flag so app stays running and we can capture async validation output
	cmd := exec.Command(binaryPath)
	cmd.Env = []string{
		"PATH=" + fakePath,
		"HOME=" + os.Getenv("HOME"),
		// Without a real PTY, the app will detect non-TTY and switch to non-interactive mode,
		// but it will still call Run() and wait for signals, keeping the process alive
	}

	// Set up pipes for stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Start the command
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start command: %v", err)
	}

	// Give the async validation goroutine time to complete and log
	// The validation runs in background, needs at least 500ms-1s
	time.Sleep(1500 * time.Millisecond)

	// Kill the process to end the test
	if cmd.Process != nil {
		cmd.Process.Kill()
	}
	cmd.Wait()

	// Collect all output
	outputStr := stdout.String() + stderr.String()

	// Should log a warning about missing dotnet
	if !strings.Contains(strings.ToLower(outputStr), "dotnet") || !strings.Contains(strings.ToLower(outputStr), "warn") {
		t.Errorf("Expected warning about missing dotnet, got: %s", outputStr)
	}

	// Verify error message includes installation instructions
	if !strings.Contains(strings.ToLower(outputStr), "installation") {
		t.Errorf("Expected installation instructions in dotnet warning, got: %s", outputStr)
	}
}

func TestConfigValidationError(t *testing.T) {
	// Create a temporary directory for config
	tmpDir := t.TempDir()

	// Write config with invalid values
	invalidConfig := `
logLevel: invalid_level
startupTimeout: 100s  # Too high (max 30s)
maxConcurrentOps: 100 # Too high (max 16)
`
	configPath := filepath.Join(tmpDir, "config.yml")
	if err := os.WriteFile(configPath, []byte(invalidConfig), 0o644); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	// Build the binary
	binaryPath := buildTestBinary(t)
	defer cleanupBinary(binaryPath)

	// Run with invalid config
	cmd := exec.Command(binaryPath, "--config", configPath, "--non-interactive")
	output, err := cmd.CombinedOutput()

	outputStr := string(output)

	// Should exit with error
	if err == nil {
		t.Fatalf("Expected error for invalid config values but got none. Output: %s", outputStr)
	}

	// Verify error message mentions validation failure
	if !strings.Contains(strings.ToLower(outputStr), "invalid") &&
		!strings.Contains(strings.ToLower(outputStr), "validation") {
		t.Errorf("Expected validation error message, got: %s", outputStr)
	}

	// Verify exit code indicates user error
	if cmd.ProcessState.ExitCode() != 1 {
		t.Errorf("Expected exit code 1 for validation error, got: %d", cmd.ProcessState.ExitCode())
	}
}

func TestGracefulErrorRecovery(t *testing.T) {
	// Build the binary
	binaryPath := buildTestBinary(t)
	defer cleanupBinary(binaryPath)

	// Test that basic errors don't cause panics
	testCases := []struct {
		name string
		args []string
	}{
		{"invalid flag", []string{"--invalid-flag"}},
		{"missing config file", []string{"--config", "/nonexistent/file.yml"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tc.args...)
			output, err := cmd.CombinedOutput()

			outputStr := string(output)

			// Should not panic
			if strings.Contains(strings.ToLower(outputStr), "panic") {
				t.Errorf("Unexpected panic in output: %s", outputStr)
			}

			// Should exit with non-zero code
			if err == nil {
				t.Error("Expected error for invalid input")
			}

			// Should provide some error message
			if len(outputStr) == 0 {
				t.Error("Expected error message in output")
			}
		})
	}
}
