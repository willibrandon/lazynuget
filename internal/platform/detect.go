// Package platform provides platform detection and system utilities.
package platform

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/willibrandon/lazynuget/internal/config"
	"github.com/willibrandon/lazynuget/internal/logging"
)

// Platform provides platform-specific information and utilities.
type Platform interface {
	// OS returns the operating system name
	OS() string

	// Arch returns the architecture
	Arch() string
}

// platformImpl implements the Platform interface.
type platformImpl struct {
	os   string
	arch string
}

func (p *platformImpl) OS() string {
	return p.os
}

func (p *platformImpl) Arch() string {
	return p.arch
}

// New creates a new Platform instance.
// For now, this is a stub that returns basic runtime information.
func New(cfg *config.AppConfig, log logging.Logger) Platform {
	return &platformImpl{
		os:   runtime.GOOS,
		arch: runtime.GOARCH,
	}
}

// PlatformInfo contains detected platform and environment information.
type PlatformInfo struct {
	// OS is the operating system (linux, darwin, windows, etc.)
	OS string

	// Arch is the architecture (amd64, arm64, etc.)
	Arch string

	// IsCI indicates if running in a CI environment
	IsCI bool

	// IsDumbTerminal indicates if TERM=dumb is set
	IsDumbTerminal bool

	// NoColor indicates if NO_COLOR environment variable is set
	NoColor bool
}

// Detect gathers platform and environment information.
func Detect() *PlatformInfo {
	return &PlatformInfo{
		OS:             runtime.GOOS,
		Arch:           runtime.GOARCH,
		IsCI:           detectCI(),
		IsDumbTerminal: isDumbTerminal(),
		NoColor:        isNoColor(),
	}
}

// detectCI checks if we're running in a CI environment.
// Checks common CI environment variables used by GitHub Actions, GitLab CI,
// Travis CI, CircleCI, Jenkins, and others.
func detectCI() bool {
	ciEnvVars := []string{
		"CI",           // Generic CI indicator
		"CONTINUOUS_INTEGRATION",
		"BUILD_NUMBER", // Jenkins
		"GITLAB_CI",
		"TRAVIS",
		"CIRCLECI",
		"GITHUB_ACTIONS",
		"TF_BUILD", // Azure Pipelines
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
	return strings.ToLower(term) == "dumb"
}

// isNoColor checks if the NO_COLOR environment variable is set.
// See https://no-color.org/ for the standard.
func isNoColor() bool {
	_, exists := os.LookupEnv("NO_COLOR")
	return exists
}

// DetermineRunMode determines if the application should run in interactive or non-interactive mode.
// It checks the following in priority order:
//  1. Explicit --non-interactive flag
//  2. CI environment detection
//  3. TTY detection (stdin and stdout must both be terminals)
//  4. Dumb terminal detection (TERM=dumb)
//
// Returns RunModeInteractive only if all conditions allow it, otherwise RunModeNonInteractive.
func DetermineRunMode(nonInteractiveFlag bool) RunMode {
	// Explicit flag takes highest priority
	if nonInteractiveFlag {
		return RunModeNonInteractive
	}

	platform := Detect()

	// CI environment implies non-interactive
	if platform.IsCI {
		return RunModeNonInteractive
	}

	// Dumb terminals can't support TUI
	if platform.IsDumbTerminal {
		return RunModeNonInteractive
	}

	// Check if we have a real TTY (both stdin and stdout must be terminals)
	if !IsTTY() {
		return RunModeNonInteractive
	}

	// All checks passed - we can run interactively
	return RunModeInteractive
}

// ValidateDotnetCLI checks if the dotnet CLI is available and functional.
// Returns nil if dotnet is found and working, otherwise returns an error with
// installation instructions.
func ValidateDotnetCLI() error {
	// Try to find dotnet in PATH
	dotnetPath, err := exec.LookPath("dotnet")
	if err != nil {
		return fmt.Errorf("dotnet CLI not found in PATH\n\n" +
			"LazyNuGet requires the .NET SDK to manage NuGet packages.\n\n" +
			"Installation instructions:\n" +
			"  • Windows: https://dotnet.microsoft.com/download\n" +
			"  • macOS: brew install dotnet-sdk\n" +
			"  • Linux: https://docs.microsoft.com/dotnet/core/install/linux")
	}

	// Verify dotnet works by running --version
	cmd := exec.Command(dotnetPath, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("dotnet CLI found at %s but failed to execute: %w\n"+
			"Output: %s\n\n"+
			"Try reinstalling the .NET SDK", dotnetPath, err, string(output))
	}

	version := strings.TrimSpace(string(output))
	if version == "" {
		return fmt.Errorf("dotnet CLI found but returned empty version\n" +
			"Try reinstalling the .NET SDK")
	}

	// Successfully validated
	return nil
}
