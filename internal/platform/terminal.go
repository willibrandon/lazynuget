package platform

// TerminalCapabilities provides terminal feature detection
type TerminalCapabilities interface {
	// GetColorDepth returns detected color support level
	GetColorDepth() ColorDepth

	// SupportsUnicode returns true if terminal can display Unicode
	SupportsUnicode() bool

	// GetSize returns terminal dimensions (width, height in characters)
	GetSize() (width int, height int, err error)

	// IsTTY returns true if stdout is connected to an interactive terminal
	IsTTY() bool

	// WatchResize registers a callback for terminal resize events
	// Returns a stop function to unregister the watcher
	WatchResize(callback func(width, height int)) (stop func())
}
