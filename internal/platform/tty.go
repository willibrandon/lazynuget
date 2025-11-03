package platform

import (
	"os"

	"golang.org/x/term"
)

// IsTerminal checks if the given file descriptor is a terminal.
// This is used to detect if stdin/stdout are connected to a real terminal
// or if they're being piped/redirected (e.g., in CI environments or scripts).
func IsTerminal(fd int) bool {
	return term.IsTerminal(fd)
}

// IsStdinTerminal checks if standard input is a terminal.
// Returns false if input is being piped or redirected.
func IsStdinTerminal() bool {
	return IsTerminal(int(os.Stdin.Fd()))
}

// IsStdoutTerminal checks if standard output is a terminal.
// Returns false if output is being piped or redirected.
func IsStdoutTerminal() bool {
	return IsTerminal(int(os.Stdout.Fd()))
}

// IsTTY checks if both stdin and stdout are terminals.
// This is the most common check for determining if a TUI can be used.
func IsTTY() bool {
	return IsStdinTerminal() && IsStdoutTerminal()
}
