package platform

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// TerminalCapabilities provides terminal feature detection
type TerminalCapabilities interface {
	// GetColorDepth returns detected color support level
	GetColorDepth() ColorDepth

	// SupportsUnicode returns true if terminal can display Unicode
	SupportsUnicode() bool

	// GetSize returns terminal dimensions (width, height in characters)
	GetSize() (width, height int, err error)

	// IsTTY returns true if stdout is connected to an interactive terminal
	IsTTY() bool

	// WatchResize registers a callback for terminal resize events
	// Returns a stop function to unregister the watcher
	WatchResize(callback func(width, height int)) (stop func())
}

// terminalCapabilities implements TerminalCapabilities interface
type terminalCapabilities struct {
	colorDepth      ColorDepth
	supportsUnicode bool
}

// NewTerminalCapabilities creates a new TerminalCapabilities instance
func NewTerminalCapabilities() TerminalCapabilities {
	return &terminalCapabilities{
		colorDepth:      detectColorDepth(),
		supportsUnicode: detectUnicodeSupport(),
	}
}

// GetColorDepth returns the detected color support level
func (t *terminalCapabilities) GetColorDepth() ColorDepth {
	return t.colorDepth
}

// SupportsUnicode returns true if terminal can display Unicode
func (t *terminalCapabilities) SupportsUnicode() bool {
	return t.supportsUnicode
}

// GetSize returns terminal dimensions (width, height in characters)
func (t *terminalCapabilities) GetSize() (width, height int, err error) {
	// Try to get size from stdout
	width, height, err = term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		// Fall back to default size if detection fails
		return 80, 24, fmt.Errorf("failed to get terminal size: %w (using defaults)", err)
	}

	return width, height, nil
}

// IsTTY returns true if stdout is connected to an interactive terminal
func (t *terminalCapabilities) IsTTY() bool {
	return IsTerminal(int(os.Stdout.Fd()))
}

// WatchResize registers a callback for terminal resize events
// Note: This is a stub implementation. Full resize watching requires platform-specific code
// and signal handling, which will be implemented in a future phase if needed.
func (t *terminalCapabilities) WatchResize(_ func(width, height int)) (stop func()) {
	// For now, return a no-op stop function
	// Full implementation would involve SIGWINCH signal handling on Unix
	// and similar mechanisms on Windows
	return func() {}
}

// detectColorDepth detects terminal color support level
func detectColorDepth() ColorDepth {
	// Check NO_COLOR environment variable (https://no-color.org/)
	if _, exists := os.LookupEnv("NO_COLOR"); exists {
		return ColorNone
	}

	// Check if not a TTY
	if !IsTerminal(int(os.Stdout.Fd())) {
		return ColorNone
	}

	// Get TERM environment variable
	term := os.Getenv("TERM")

	// Check for true color support
	colorTerm := os.Getenv("COLORTERM")
	if colorTerm == "truecolor" || colorTerm == "24bit" {
		return ColorTrueColor
	}

	// Check TERM for color depth indicators
	if strings.Contains(term, "256color") {
		return ColorExtended256
	}

	if strings.Contains(term, "color") {
		return ColorBasic16
	}

	// Check for dumb terminal
	if term == "dumb" || term == "" {
		return ColorNone
	}

	// Default to basic 16-color support for most terminals
	return ColorBasic16
}

// detectUnicodeSupport detects if terminal can display Unicode characters
func detectUnicodeSupport() bool {
	// Check LANG environment variable for UTF-8
	lang := os.Getenv("LANG")
	if strings.Contains(strings.ToUpper(lang), "UTF-8") || strings.Contains(strings.ToUpper(lang), "UTF8") {
		return true
	}

	// Check LC_ALL
	lcAll := os.Getenv("LC_ALL")
	if strings.Contains(strings.ToUpper(lcAll), "UTF-8") || strings.Contains(strings.ToUpper(lcAll), "UTF8") {
		return true
	}

	// Check LC_CTYPE
	lcCtype := os.Getenv("LC_CTYPE")
	if strings.Contains(strings.ToUpper(lcCtype), "UTF-8") || strings.Contains(strings.ToUpper(lcCtype), "UTF8") {
		return true
	}

	// Default to false for safety
	return false
}
