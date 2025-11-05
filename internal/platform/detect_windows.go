//go:build windows

package platform

import (
	"fmt"

	"golang.org/x/sys/windows"
)

// detectOSVersion returns Windows version string
func detectOSVersion() string {
	// Get Windows version
	v := windows.RtlGetVersion()
	if v != nil {
		return fmt.Sprintf("Windows %d.%d.%d", v.MajorVersion, v.MinorVersion, v.BuildNumber)
	}

	// Fallback if version detection fails
	return "Windows (version unknown)"
}
