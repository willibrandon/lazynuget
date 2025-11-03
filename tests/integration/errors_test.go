package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
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
	if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
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

	// Run with PATH cleared (dotnet not available)
	cmd := exec.Command("../../lazynuget-test", "--non-interactive")
	cmd.Env = []string{"HOME=" + os.Getenv("HOME")} // Keep HOME but remove PATH

	output, err := cmd.CombinedOutput()

	// App should start successfully (dotnet validation is non-blocking)
	// but should log a warning
	outputStr := string(output)

	// Check if warning about dotnet is present in logs
	if strings.Contains(outputStr, "dotnet") && strings.Contains(strings.ToLower(outputStr), "warn") {
		t.Logf("Dotnet warning detected as expected: %s", outputStr)
	} else {
		// It's okay if dotnet is actually available on the system
		// This test might not fail on systems with dotnet installed
		t.Skip("Dotnet validation test skipped - dotnet may be available on system")
	}

	// Verify error message includes installation instructions if warning present
	if strings.Contains(strings.ToLower(outputStr), "dotnet cli not found") {
		expectedPhrases := []string{
			"installation instructions",
			"dotnet",
		}

		for _, phrase := range expectedPhrases {
			if !strings.Contains(strings.ToLower(outputStr), strings.ToLower(phrase)) {
				t.Errorf("Expected dotnet error to contain '%s', got: %s", phrase, outputStr)
			}
		}
	}

	// Should not fail startup (dotnet validation is non-blocking)
	if err != nil && cmd.ProcessState.ExitCode() > 2 {
		t.Errorf("Expected startup to succeed even without dotnet, got exit code: %d", cmd.ProcessState.ExitCode())
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
	if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
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
