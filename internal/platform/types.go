package platform

// ColorDepth represents terminal color support level
type ColorDepth int

const (
	ColorNone        ColorDepth = 0        // No color support
	ColorBasic16     ColorDepth = 16       // 16 ANSI colors
	ColorExtended256 ColorDepth = 256      // 256-color palette
	ColorTrueColor   ColorDepth = 16777216 // 24-bit true color
)

// String returns human-readable color depth description
func (c ColorDepth) String() string {
	switch c {
	case ColorNone:
		return "none"
	case ColorBasic16:
		return "16-color"
	case ColorExtended256:
		return "256-color"
	case ColorTrueColor:
		return "true-color"
	default:
		return "unknown"
	}
}

// RunMode represents whether the application is running interactively or not
type RunMode int

const (
	RunModeInteractive    RunMode = iota // Interactive mode (TTY available)
	RunModeNonInteractive                // Non-interactive mode (no TTY, CI, etc.)
)

// String returns human-readable run mode description
func (r RunMode) String() string {
	switch r {
	case RunModeInteractive:
		return "interactive"
	case RunModeNonInteractive:
		return "non-interactive"
	default:
		return "unknown"
	}
}

// IsInteractive returns true if running in interactive mode
func (r RunMode) IsInteractive() bool {
	return r == RunModeInteractive
}
