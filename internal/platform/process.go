package platform

// ProcessSpawner handles platform-specific process execution
type ProcessSpawner interface {
	// Run executes a process and waits for completion
	// Automatically handles:
	// - PATH resolution for executable
	// - Argument quoting for paths with spaces
	// - Output encoding detection and conversion to UTF-8
	// - Exit code extraction
	Run(executable string, args []string, workingDir string, env map[string]string) (ProcessResult, error)

	// SetEncoding overrides automatic encoding detection
	// Use "utf-8", "windows-1252", "iso-8859-1", etc.
	// Pass empty string to re-enable auto-detection
	SetEncoding(encoding string)
}
