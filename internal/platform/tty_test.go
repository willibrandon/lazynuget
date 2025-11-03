package platform

import (
	"os"
	"testing"
)

// TestIsTerminal tests terminal detection for file descriptors
func TestIsTerminal(t *testing.T) {
	// Test with stdin (fd 0)
	stdinIsTerm := IsTerminal(int(os.Stdin.Fd()))
	t.Logf("stdin is terminal: %v", stdinIsTerm)

	// Test with stdout (fd 1)
	stdoutIsTerm := IsTerminal(int(os.Stdout.Fd()))
	t.Logf("stdout is terminal: %v", stdoutIsTerm)

	// Test with stderr (fd 2)
	stderrIsTerm := IsTerminal(int(os.Stderr.Fd()))
	t.Logf("stderr is terminal: %v", stderrIsTerm)

	// The function should not panic with valid file descriptors
	// Actual terminal detection depends on how tests are run (TTY vs pipe)
}

// TestIsStdinTerminal tests stdin terminal detection
func TestIsStdinTerminal(t *testing.T) {
	result := IsStdinTerminal()
	t.Logf("IsStdinTerminal() = %v", result)

	// Should be consistent with IsTerminal for stdin
	expected := IsTerminal(int(os.Stdin.Fd()))
	if result != expected {
		t.Errorf("IsStdinTerminal() = %v, want %v", result, expected)
	}
}

// TestIsStdoutTerminal tests stdout terminal detection
func TestIsStdoutTerminal(t *testing.T) {
	result := IsStdoutTerminal()
	t.Logf("IsStdoutTerminal() = %v", result)

	// Should be consistent with IsTerminal for stdout
	expected := IsTerminal(int(os.Stdout.Fd()))
	if result != expected {
		t.Errorf("IsStdoutTerminal() = %v, want %v", result, expected)
	}
}

// TestIsTTY tests combined stdin/stdout terminal detection
func TestIsTTY(t *testing.T) {
	result := IsTTY()
	t.Logf("IsTTY() = %v", result)

	// Should be true only if both stdin and stdout are terminals
	expectedStdin := IsStdinTerminal()
	expectedStdout := IsStdoutTerminal()
	expected := expectedStdin && expectedStdout

	if result != expected {
		t.Errorf("IsTTY() = %v, want %v (stdin=%v, stdout=%v)",
			result, expected, expectedStdin, expectedStdout)
	}
}

// TestTerminalConsistency verifies consistency between helper functions
func TestTerminalConsistency(t *testing.T) {
	// IsTTY should match the AND of stdin and stdout checks
	isTTY := IsTTY()
	isStdin := IsStdinTerminal()
	isStdout := IsStdoutTerminal()

	if isTTY != (isStdin && isStdout) {
		t.Errorf("IsTTY() = %v, but IsStdinTerminal() = %v and IsStdoutTerminal() = %v",
			isTTY, isStdin, isStdout)
	}
}
