package platform

// RunMode defines whether the application runs in interactive or non-interactive mode.
type RunMode int

const (
	// RunModeInteractive indicates the application should run with a full TUI.
	// This is the default mode when connected to a terminal.
	RunModeInteractive RunMode = iota

	// RunModeNonInteractive indicates the application should run without a TUI.
	// Used in CI environments, piped contexts, or when explicitly requested.
	RunModeNonInteractive
)

// String returns the string representation of the run mode.
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

// IsInteractive returns true if the run mode is interactive.
func (r RunMode) IsInteractive() bool {
	return r == RunModeInteractive
}
