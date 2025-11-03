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
	// Updated for Phase 4: error message now says "YAML parsing error"
	expectedPhrases := []string{
		"YAML parsing error",
		"syntax errors",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(outputStr, phrase) {
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
	// Per Phase 4 (FR-012, FR-013): Semantic validation errors are non-blocking
	// Invalid values fall back to defaults with warnings, app continues to run
	tmpDir := t.TempDir()

	// Write config with invalid values (semantic errors, not syntax errors)
	invalidConfig := `
logLevel: invalid_level      # Invalid enum - falls back to default
maxConcurrentOps: 100        # Out of range (max 16) - falls back to default
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

	// Per FR-012: App should START successfully (non-blocking semantic errors)
	// Exit code 0 is success - app ran and shut down cleanly
	if err != nil && cmd.ProcessState.ExitCode() > 1 {
		t.Fatalf("App failed to start with semantic validation errors. Exit code: %d, Output: %s",
			cmd.ProcessState.ExitCode(), outputStr)
	}

	// Verify validation warnings were logged (per FR-013)
	// Should contain warnings about invalid values and mention defaults being used
	if !strings.Contains(outputStr, "WARN") && !strings.Contains(outputStr, "warn") {
		t.Errorf("Expected warning logs for invalid config values, got: %s", outputStr)
	}

	// Verify app bootstrapped and started (not blocked by semantic errors)
	if !strings.Contains(outputStr, "Bootstrap complete") {
		t.Errorf("Expected app to complete bootstrap despite invalid config values, got: %s", outputStr)
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
