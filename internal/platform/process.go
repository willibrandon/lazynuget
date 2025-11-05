package platform

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ProcessResult contains the output and exit status of a process
// See: T083, contracts/process.md
type ProcessResult struct {
	Stdout   string // Standard output (decoded to UTF-8)
	Stderr   string // Standard error (decoded to UTF-8)
	ExitCode int    // Process exit code (0 = success)
}

// ProcessSpawner handles platform-specific process execution
// See: T082, contracts/process.md
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

// processSpawner is the concrete implementation of ProcessSpawner
// See: T082
type processSpawner struct {
	encoding string // Manual encoding override (empty = auto-detect)
}

// NewProcessSpawner creates a new ProcessSpawner instance
// See: T082, FR-030, FR-031
func NewProcessSpawner() ProcessSpawner {
	return &processSpawner{
		encoding: "", // Auto-detect by default
	}
}

// SetEncoding overrides automatic encoding detection
// See: T085, FR-030
func (p *processSpawner) SetEncoding(encoding string) {
	p.encoding = encoding
}

// Run executes a process and waits for completion
// See: T084, T086, FR-030, FR-031
func (p *processSpawner) Run(executable string, args []string, workingDir string, env map[string]string) (ProcessResult, error) {
	// Validate inputs
	if executable == "" {
		return ProcessResult{}, fmt.Errorf("executable cannot be empty")
	}

	// Resolve executable path
	execPath, err := resolveExecutable(executable)
	if err != nil {
		return ProcessResult{}, fmt.Errorf("failed to resolve executable %q: %w", executable, err)
	}

	// Validate working directory if specified
	if workingDir != "" {
		if _, statErr := os.Stat(workingDir); statErr != nil {
			return ProcessResult{}, fmt.Errorf("working directory does not exist: %s", workingDir)
		}
	}

	// Create command
	// G204: This is safe - execPath comes from resolveExecutable which validates the path
	cmd := exec.Command(execPath, args...) // #nosec G204

	// Set working directory if specified
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	// Merge environment variables with parent environment
	if env != nil {
		// Start with parent environment
		cmdEnv := os.Environ()

		// Add/override with custom env vars
		for key, value := range env {
			// Validate key doesn't contain = or null bytes
			if strings.Contains(key, "=") || strings.Contains(key, "\x00") {
				return ProcessResult{}, fmt.Errorf("invalid environment variable key: %q", key)
			}

			// Find and replace existing var, or append new one
			found := false
			for i, e := range cmdEnv {
				if strings.HasPrefix(e, key+"=") {
					cmdEnv[i] = key + "=" + value
					found = true
					break
				}
			}
			if !found {
				cmdEnv = append(cmdEnv, key+"="+value)
			}
		}

		cmd.Env = cmdEnv
	}

	// Capture stdout and stderr separately
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	// Execute command
	execErr := cmd.Run()

	// Determine encoding to use
	encoding := p.encoding
	if encoding == "" {
		// Auto-detect: try UTF-8 first, fallback to system encoding
		encoding = "" // Empty triggers auto-detection in decodeBytes
	}

	// Decode stdout and stderr (decodeBytes is infallible)
	stdout := decodeBytes(stdoutBuf.Bytes(), encoding)
	stderr := decodeBytes(stderrBuf.Bytes(), encoding)

	// Extract exit code
	exitCode := 0
	if execErr != nil {
		// Check if it's an ExitError (command ran but returned non-zero)
		var exitErr *exec.ExitError
		if !errors.As(execErr, &exitErr) {
			// Command failed to run (not found, permission denied, etc.)
			return ProcessResult{}, fmt.Errorf("failed to execute command: %w", execErr)
		}
		exitCode = exitErr.ExitCode()
	}

	return ProcessResult{
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: exitCode,
	}, nil
}

// resolveExecutable resolves an executable name to its full path
// Handles both absolute paths and PATH lookups
// Platform-specific implementations in process_windows.go and process_unix.go
// See: T084, T087, T088, FR-031
func resolveExecutable(executable string) (string, error) {
	// Use platform-specific resolution
	// On Windows: tries .exe, .cmd, .bat extensions (process_windows.go)
	// On Unix: validates execute permissions (process_unix.go)
	return resolveExecutablePlatform(executable)
}

// resolveExecutablePlatform is implemented in platform-specific files:
// - process_windows.go: tries .exe, .cmd, .bat extensions
// - process_unix.go: validates execute permissions
// This declaration satisfies the compiler; actual implementation is in platform files
