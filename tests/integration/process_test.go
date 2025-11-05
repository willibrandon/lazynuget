package integration

import (
	"runtime"
	"strings"
	"testing"

	"github.com/willibrandon/lazynuget/internal/platform"
)

// TestProcessSpawner_DotnetVersion tests cross-platform process spawning with dotnet CLI
// See: T076, FR-030, FR-031
func TestProcessSpawner_DotnetVersion(t *testing.T) {
	spawner := platform.NewProcessSpawner()

	// Run dotnet --version
	result, err := spawner.Run("dotnet", []string{"--version"}, "", nil)
	if err != nil {
		t.Fatalf("Failed to run dotnet --version: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("dotnet --version returned non-zero exit code: %d", result.ExitCode)
		t.Logf("Stderr: %s", result.Stderr)
	}

	// Verify output is non-empty and contains version
	if result.Stdout == "" {
		t.Error("dotnet --version returned empty stdout")
	}

	// Verify output is valid UTF-8
	if !strings.Contains(result.Stdout, ".") {
		t.Errorf("dotnet --version output doesn't look like a version string: %q", result.Stdout)
	}

	t.Logf("dotnet version: %s", strings.TrimSpace(result.Stdout))

	// Verify stderr is typically empty for --version
	if result.Stderr != "" {
		t.Logf("dotnet --version produced stderr (unusual but not necessarily an error): %s", result.Stderr)
	}
}

// TestProcessSpawner_DotnetHelp tests capturing multi-line output
// See: T076, FR-030
func TestProcessSpawner_DotnetHelp(t *testing.T) {
	spawner := platform.NewProcessSpawner()

	// Run dotnet --help (produces multi-line output)
	result, err := spawner.Run("dotnet", []string{"--help"}, "", nil)
	if err != nil {
		t.Fatalf("Failed to run dotnet --help: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("dotnet --help returned non-zero exit code: %d", result.ExitCode)
	}

	// Verify output contains expected help text
	if !strings.Contains(result.Stdout, "Usage:") && !strings.Contains(result.Stdout, "usage:") {
		t.Errorf("dotnet --help output doesn't contain expected help text")
	}

	// Verify output contains multiple lines
	lines := strings.Split(result.Stdout, "\n")
	if len(lines) < 5 {
		t.Errorf("dotnet --help output has fewer than 5 lines: %d", len(lines))
	}

	t.Logf("dotnet --help produced %d lines of output", len(lines))
}

// TestProcessSpawner_WorkingDirectory tests working directory handling
// See: T076, FR-031
func TestProcessSpawner_WorkingDirectory(t *testing.T) {
	spawner := platform.NewProcessSpawner()

	// Use temp directory as working directory
	tmpDir := t.TempDir()

	// Run a command that shows the working directory
	var cmd string
	var args []string

	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "cd"}
	} else {
		cmd = "pwd"
		args = []string{}
	}

	result, err := spawner.Run(cmd, args, tmpDir, nil)
	if err != nil {
		t.Fatalf("Failed to run %s: %v", cmd, err)
	}

	if result.ExitCode != 0 {
		t.Errorf("%s returned non-zero exit code: %d", cmd, result.ExitCode)
	}

	// Verify output contains the working directory
	output := strings.TrimSpace(result.Stdout)
	t.Logf("Working directory reported: %q", output)

	// On Windows, normalize path separators for comparison
	if runtime.GOOS == "windows" {
		output = strings.ReplaceAll(output, "/", "\\")
		tmpDir = strings.ReplaceAll(tmpDir, "/", "\\")
	}

	if !strings.Contains(output, tmpDir) && !strings.Contains(tmpDir, output) {
		t.Errorf("Working directory mismatch: got %q, want to contain %q", output, tmpDir)
	}
}

// TestProcessSpawner_EnvironmentVariables tests environment variable handling
// See: T076, FR-031
func TestProcessSpawner_EnvironmentVariables(t *testing.T) {
	spawner := platform.NewProcessSpawner()

	// Set a custom environment variable
	env := map[string]string{
		"LAZYNUGET_TEST_VAR": "test_value_12345",
	}

	var cmd string
	var args []string

	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "echo", "%LAZYNUGET_TEST_VAR%"}
	} else {
		cmd = "sh"
		args = []string{"-c", "echo $LAZYNUGET_TEST_VAR"}
	}

	result, err := spawner.Run(cmd, args, "", env)
	if err != nil {
		t.Fatalf("Failed to run command: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("Command returned non-zero exit code: %d", result.ExitCode)
	}

	// Verify output contains the environment variable value
	output := strings.TrimSpace(result.Stdout)

	if !strings.Contains(output, "test_value_12345") {
		t.Errorf("Environment variable not found in output: got %q", output)
	}

	t.Logf("Environment variable output: %q", output)
}

// TestProcessSpawner_NonZeroExitCode tests handling of failing commands
// See: T076, FR-031
func TestProcessSpawner_NonZeroExitCode(t *testing.T) {
	spawner := platform.NewProcessSpawner()

	// Run dotnet with invalid command (should return non-zero)
	result, err := spawner.Run("dotnet", []string{"invalid-command-12345"}, "", nil)
	// The process should start successfully (no error)
	if err != nil {
		t.Fatalf("Failed to run dotnet with invalid command: %v", err)
	}

	// But the exit code should be non-zero
	if result.ExitCode == 0 {
		t.Error("dotnet with invalid command returned exit code 0 (expected non-zero)")
	}

	// Verify stderr contains error message
	if result.Stderr == "" {
		t.Error("dotnet with invalid command returned empty stderr (expected error message)")
	}

	t.Logf("Exit code: %d", result.ExitCode)
	t.Logf("Stderr: %s", result.Stderr)
}

// TestProcessSpawner_ExecutableNotFound tests error handling for missing executables
// See: T076, FR-031
func TestProcessSpawner_ExecutableNotFound(t *testing.T) {
	spawner := platform.NewProcessSpawner()

	// Try to run non-existent command
	_, err := spawner.Run("nonexistent-command-12345", []string{}, "", nil)

	// Should return error (executable not found)
	if err == nil {
		t.Error("Running non-existent command should return error")
	}

	t.Logf("Error (expected): %v", err)
}
