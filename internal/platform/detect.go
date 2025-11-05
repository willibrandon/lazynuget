package platform

import (
	"fmt"
	"runtime"
)

// platformInfo implements the PlatformInfo interface
type platformInfo struct {
	os      string
	arch    string
	version string
}

// OS returns the operating system: "windows", "darwin", or "linux"
func (p *platformInfo) OS() string {
	return p.os
}

// Arch returns the architecture: "amd64" or "arm64"
func (p *platformInfo) Arch() string {
	return p.arch
}

// Version returns OS version string for diagnostics
func (p *platformInfo) Version() string {
	return p.version
}

// IsWindows returns true if running on Windows
func (p *platformInfo) IsWindows() bool {
	return p.os == "windows"
}

// IsDarwin returns true if running on macOS
func (p *platformInfo) IsDarwin() bool {
	return p.os == "darwin"
}

// IsLinux returns true if running on Linux
func (p *platformInfo) IsLinux() bool {
	return p.os == "linux"
}

// detectPlatform creates a new platformInfo with detected OS and architecture
func detectPlatform() (*platformInfo, error) {
	os := runtime.GOOS
	arch := runtime.GOARCH

	// Validate supported platforms
	switch os {
	case "windows", "darwin", "linux":
		// Supported
	default:
		return nil, fmt.Errorf("unsupported operating system: %s (supported: windows, darwin, linux)", os)
	}

	// Validate supported architectures
	switch arch {
	case "amd64", "arm64":
		// Supported
	default:
		return nil, fmt.Errorf("unsupported architecture: %s (supported: amd64, arm64)", arch)
	}

	p := &platformInfo{
		os:   os,
		arch: arch,
	}

	// Detect OS version (platform-specific)
	p.version = detectOSVersion()

	return p, nil
}
