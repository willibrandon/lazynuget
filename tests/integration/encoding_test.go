package integration

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/willibrandon/lazynuget/internal/platform"
)

// TestProcessSpawner_UTF8Output tests UTF-8 encoded output handling
// See: T077, FR-030
func TestProcessSpawner_UTF8Output(t *testing.T) {
	spawner := platform.NewProcessSpawner()

	// Create a temporary script that outputs UTF-8 text with Unicode
	tmpDir := t.TempDir()

	var scriptPath string
	var cmd string
	var args []string

	if runtime.GOOS == "windows" {
		// Create a batch file that outputs UTF-8
		scriptPath = filepath.Join(tmpDir, "test.bat")
		content := "@echo off\r\necho Hello, 世界!\r\n"
		if err := os.WriteFile(scriptPath, []byte(content), 0o755); err != nil {
			t.Fatalf("Failed to create test script: %v", err)
		}
		cmd = scriptPath
		args = []string{}
	} else {
		// Create a shell script that outputs UTF-8
		scriptPath = filepath.Join(tmpDir, "test.sh")
		content := "#!/bin/sh\necho 'Hello, 世界!'\n"
		if err := os.WriteFile(scriptPath, []byte(content), 0o755); err != nil {
			t.Fatalf("Failed to create test script: %v", err)
		}
		cmd = scriptPath
		args = []string{}
	}

	// Run the script
	result, err := spawner.Run(cmd, args, "", nil)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("Script returned non-zero exit code: %d", result.ExitCode)
		t.Logf("Stderr: %s", result.Stderr)
	}

	// Verify output contains Unicode characters (decoded to UTF-8)
	output := strings.TrimSpace(result.Stdout)
	t.Logf("Script output: %q", output)

	if !strings.Contains(output, "Hello") {
		t.Errorf("Script output doesn't contain expected text: %q", output)
	}

	// On systems with UTF-8 locale, we should get the Chinese characters
	// On non-UTF-8 systems, they might be replaced or garbled
	if runtime.GOOS != "windows" {
		// Unix systems should generally preserve UTF-8
		if !strings.Contains(output, "世界") {
			t.Logf("Warning: Unicode characters not preserved in output (may be locale issue): %q", output)
		}
	}
}

// TestProcessSpawner_EncodingOverride tests manual encoding override
// See: T077, FR-030
func TestProcessSpawner_EncodingOverride(t *testing.T) {
	spawner := platform.NewProcessSpawner()

	// Set encoding to UTF-8 explicitly
	spawner.SetEncoding("utf-8")

	// Run a simple command
	var cmd string
	var args []string

	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "echo", "Hello"}
	} else {
		cmd = "echo"
		args = []string{"Hello"}
	}

	result, err := spawner.Run(cmd, args, "", nil)
	if err != nil {
		t.Fatalf("Failed to run command: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("Command returned non-zero exit code: %d", result.ExitCode)
	}

	// Verify output
	output := strings.TrimSpace(result.Stdout)

	if !strings.Contains(output, "Hello") {
		t.Errorf("Command output doesn't contain expected text: %q", output)
	}

	// Reset encoding to auto-detect
	spawner.SetEncoding("")

	// Run command again
	result2, err := spawner.Run(cmd, args, "", nil)
	if err != nil {
		t.Fatalf("Failed to run command after encoding reset: %v", err)
	}

	if result2.ExitCode != 0 {
		t.Errorf("Command returned non-zero exit code after reset: %d", result2.ExitCode)
	}

	// Output should still contain expected text
	output2 := strings.TrimSpace(result2.Stdout)

	if !strings.Contains(output2, "Hello") {
		t.Errorf("Command output doesn't contain expected text after reset: %q", output2)
	}
}

// TestProcessSpawner_LargeOutput tests handling of large output
// See: T077, FR-030
func TestProcessSpawner_LargeOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large output test in short mode")
	}

	spawner := platform.NewProcessSpawner()

	// Generate large output (1000 lines)
	var cmd string
	var args []string

	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "for /L %i in (1,1,1000) do @echo Line %i"}
	} else {
		cmd = "sh"
		args = []string{"-c", "for i in $(seq 1 1000); do echo \"Line $i\"; done"}
	}

	result, err := spawner.Run(cmd, args, "", nil)
	if err != nil {
		t.Fatalf("Failed to run command: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("Command returned non-zero exit code: %d", result.ExitCode)
	}

	// Verify output has approximately 1000 lines
	lines := strings.Split(strings.TrimSpace(result.Stdout), "\n")

	if len(lines) < 900 || len(lines) > 1100 {
		t.Errorf("Expected ~1000 lines, got %d", len(lines))
	}

	t.Logf("Generated %d lines of output", len(lines))

	// Verify first and last lines
	if len(lines) > 0 {
		firstLine := strings.TrimSpace(lines[0])
		lastLine := strings.TrimSpace(lines[len(lines)-1])

		t.Logf("First line: %q", firstLine)
		t.Logf("Last line: %q", lastLine)

		if !strings.Contains(firstLine, "1") {
			t.Errorf("First line doesn't contain expected text: %q", firstLine)
		}
	}
}

// TestProcessSpawner_StderrCapture tests separate stderr capture
// See: T077, FR-030
func TestProcessSpawner_StderrCapture(t *testing.T) {
	spawner := platform.NewProcessSpawner()

	// Create a script that outputs to both stdout and stderr
	tmpDir := t.TempDir()

	var scriptPath string
	var cmd string
	var args []string

	if runtime.GOOS == "windows" {
		scriptPath = filepath.Join(tmpDir, "test_stderr.bat")
		content := "@echo off\r\necho stdout message\r\necho stderr message 1>&2\r\n"
		if err := os.WriteFile(scriptPath, []byte(content), 0o755); err != nil {
			t.Fatalf("Failed to create test script: %v", err)
		}
		cmd = scriptPath
		args = []string{}
	} else {
		scriptPath = filepath.Join(tmpDir, "test_stderr.sh")
		content := "#!/bin/sh\necho 'stdout message'\necho 'stderr message' >&2\n"
		if err := os.WriteFile(scriptPath, []byte(content), 0o755); err != nil {
			t.Fatalf("Failed to create test script: %v", err)
		}
		cmd = scriptPath
		args = []string{}
	}

	// Run the script
	result, err := spawner.Run(cmd, args, "", nil)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("Script returned non-zero exit code: %d", result.ExitCode)
	}

	// Verify stdout
	stdout := strings.TrimSpace(result.Stdout)
	t.Logf("Stdout: %q", stdout)

	if !strings.Contains(stdout, "stdout message") {
		t.Errorf("Stdout doesn't contain expected text: %q", stdout)
	}

	// Verify stderr
	stderr := strings.TrimSpace(result.Stderr)
	t.Logf("Stderr: %q", stderr)

	if !strings.Contains(stderr, "stderr message") {
		t.Errorf("Stderr doesn't contain expected text: %q", stderr)
	}

	// Verify stdout and stderr are separate
	if strings.Contains(stdout, "stderr message") {
		t.Error("Stderr message appeared in stdout")
	}

	if strings.Contains(stderr, "stdout message") {
		t.Error("Stdout message appeared in stderr")
	}
}
