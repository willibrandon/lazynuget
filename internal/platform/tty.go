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

// DetermineRunMode determines if the application should run in interactive or non-interactive mode.
// It checks the following in priority order:
//  1. Explicit non-interactive flag
//  2. CI environment detection (CI=true/1, GITHUB_ACTIONS, etc.)
//  3. Dumb terminal detection (TERM=dumb)
//  4. TTY detection (stdin and stdout must both be terminals)
//
// Returns RunModeInteractive only if all conditions allow it, otherwise RunModeNonInteractive.
func DetermineRunMode(nonInteractiveFlag bool) RunMode {
	// Explicit flag takes highest priority
	if nonInteractiveFlag {
		return RunModeNonInteractive
	}

	// CI environment implies non-interactive
	if detectCI() {
		return RunModeNonInteractive
	}

	// Dumb terminals can't support TUI
	if isDumbTerminal() {
		return RunModeNonInteractive
	}

	// Check if we have a real TTY (both stdin and stdout must be terminals)
	if !IsTTY() {
		return RunModeNonInteractive
	}

	// All checks passed - we can run interactively
	return RunModeInteractive
}

// detectCI checks if we're running in a CI environment.
// Checks common CI environment variables used by GitHub Actions, GitLab CI,
// Travis CI, CircleCI, Jenkins, and others.
func detectCI() bool {
	ciEnvVars := []string{
		"CI",                     // Generic CI indicator
		"CONTINUOUS_INTEGRATION", // Generic CI indicator
		"BUILD_NUMBER",           // Jenkins
		"GITLAB_CI",              // GitLab CI
		"TRAVIS",                 // Travis CI
		"CIRCLECI",               // CircleCI
		"GITHUB_ACTIONS",         // GitHub Actions
		"TF_BUILD",               // Azure Pipelines
	}

	for _, envVar := range ciEnvVars {
		value := os.Getenv(envVar)
		if value == "true" || value == "1" || value == "yes" {
			return true
		}
	}

	return false
}

// isDumbTerminal checks if TERM is set to "dumb".
// Dumb terminals don't support TUI features like cursor movement or colors.
func isDumbTerminal() bool {
	term := os.Getenv("TERM")
	return term == "dumb"
}
