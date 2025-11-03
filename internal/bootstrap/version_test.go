package bootstrap

import (
	"testing"
)

// TestShowVersion tests the version display function
func TestShowVersion(t *testing.T) {
	tests := []struct {
		name    string
		version VersionInfo
	}{
		{
			name: "full version info",
			version: VersionInfo{
				Version: "1.0.0",
				Commit:  "abc123def456",
				Date:    "2025-01-01",
			},
		},
		{
			name: "dev version",
			version: VersionInfo{
				Version: "dev",
				Commit:  "unknown",
				Date:    "unknown",
			},
		},
		{
			name: "empty version",
			version: VersionInfo{
				Version: "",
				Commit:  "",
				Date:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			// ShowVersion should not panic
			ShowVersion(tt.version)
		})
	}
}

// TestShowVersionWithRealApp tests ShowVersion with actual app version
func TestShowVersionWithRealApp(t *testing.T) {
	app, err := NewApp("1.2.3", "abc123", "2025-01-01")
	if err != nil {
		t.Fatalf("NewApp() failed: %v", err)
	}
	defer app.cancel()

	// Should not panic
	ShowVersion(app.version)
}
