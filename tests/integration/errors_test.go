package integration

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/creack/pty"
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
	buildCmd := exec.Command("go", "build", "-o", "../../lazynuget-test", "../../cmd/lazynuget/main.go")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer exec.Command("rm", "../../lazynuget-test").Run()

	// Run with invalid config
	cmd := exec.Command("../../lazynuget-test", "--config", configPath, "--non-interactive")
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
	buildCmd := exec.Command("go", "build", "-o", "../../lazynuget-test", "../../cmd/lazynuget/main.go")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer exec.Command("rm", "../../lazynuget-test").Run()

	// Create a temporary directory with NO dotnet executable
	tmpDir := t.TempDir()
	fakePath := tmpDir + ":/usr/bin:/bin" // Add minimal PATH without dotnet locations

	// Run with PTY (real terminal) so app stays in interactive mode
	cmd := exec.Command("../../lazynuget-test")
	cmd.Env = []string{
		"PATH=" + fakePath,
		"HOME=" + os.Getenv("HOME"),
		"TERM=xterm",
	}

	// Start command with PTY (creates real terminal)
	ptmx, err := pty.Start(cmd)
	if err != nil {
		t.Fatalf("Failed to start command with pty: %v", err)
	}
	defer ptmx.Close()

	// Collect output in background
	outputChan := make(chan string)
	go func() {
		output, _ := io.ReadAll(ptmx)
		outputChan <- string(output)
	}()

	// Give async goroutine time to execute (dotnet validation runs in background)
	// Wait longer to ensure the async validation completes
	time.Sleep(2 * time.Second)

	// Kill the process
	cmd.Process.Kill()
	cmd.Wait()

	// Get collected output
	outputStr := <-outputChan

	// Should log a warning about missing dotnet (async validation)
	// Note: This is a warning, not a fatal error - app continues to run
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
	buildCmd := exec.Command("go", "build", "-o", "../../lazynuget-test", "../../cmd/lazynuget/main.go")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer exec.Command("rm", "../../lazynuget-test").Run()

	// Run with invalid config
	cmd := exec.Command("../../lazynuget-test", "--config", configPath, "--non-interactive")
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
	buildCmd := exec.Command("go", "build", "-o", "../../lazynuget-test", "../../cmd/lazynuget/main.go")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer exec.Command("rm", "../../lazynuget-test").Run()

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
			cmd := exec.Command("../../lazynuget-test", tc.args...)
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
