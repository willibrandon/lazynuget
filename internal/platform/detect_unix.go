//go:build !windows

package platform

import (
	"bytes"
	"os/exec"
	"runtime"
	"strings"
)

// detectOSVersion returns OS version string for Unix-like systems
func detectOSVersion() string {
	switch runtime.GOOS {
	case "darwin":
		return detectDarwinVersion()
	case "linux":
		return detectLinuxVersion()
	default:
		return runtime.GOOS + " (version unknown)"
	}
}

// detectDarwinVersion returns macOS version
func detectDarwinVersion() string {
	cmd := exec.Command("sw_vers", "-productVersion")
	out, err := cmd.Output()
	if err != nil {
		return "macOS (version unknown)"
	}

	version := strings.TrimSpace(string(out))
	return "macOS " + version
}

// detectLinuxVersion returns Linux kernel version
func detectLinuxVersion() string {
	cmd := exec.Command("uname", "-r")
	out, err := cmd.Output()
	if err != nil {
		return "Linux (version unknown)"
	}

	version := strings.TrimSpace(string(bytes.TrimSpace(out)))
	return "Linux " + version
}
