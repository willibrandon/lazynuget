// Package bootstrap provides application initialization and lifecycle management.
package bootstrap

import "fmt"

// VersionInfo contains build and release information.
type VersionInfo struct {
	// Version is the semantic version string (e.g., "1.0.0")
	Version string

	// Commit is the Git commit SHA (short or full)
	Commit string

	// Date is the build timestamp in RFC3339 format
	Date string
}

// String formats version info for display to users.
func (v VersionInfo) String() string {
	return fmt.Sprintf("LazyNuGet version %s (%s) built on %s", v.Version, v.Commit, v.Date)
}
