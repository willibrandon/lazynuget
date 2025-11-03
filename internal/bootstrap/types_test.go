package bootstrap

import (
	"strings"
	"testing"
)

// TestVersionInfoString tests the String method
func TestVersionInfoString(t *testing.T) {
	tests := []struct {
		name     string
		version  VersionInfo
		contains []string
	}{
		{
			name: "full version info",
			version: VersionInfo{
				Version: "1.0.0",
				Commit:  "abc123",
				Date:    "2025-01-01",
			},
			contains: []string{"1.0.0", "abc123", "2025-01-01"},
		},
		{
			name: "dev version",
			version: VersionInfo{
				Version: "dev",
				Commit:  "unknown",
				Date:    "unknown",
			},
			contains: []string{"dev", "unknown"},
		},
		{
			name: "empty version",
			version: VersionInfo{
				Version: "",
				Commit:  "",
				Date:    "",
			},
			contains: []string{},
		},
		{
			name: "partial version",
			version: VersionInfo{
				Version: "0.1.0-alpha",
				Commit:  "abcdef123456",
				Date:    "2025-11-03",
			},
			contains: []string{"0.1.0-alpha", "abcdef123456", "2025-11-03"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.version.String()

			if got == "" && len(tt.contains) > 0 {
				t.Error("String() returned empty string")
			}

			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("String() = %q, want it to contain %q", got, want)
				}
			}
		})
	}
}

// TestVersionInfoStringNotEmpty tests that String always returns something
func TestVersionInfoStringNotEmpty(t *testing.T) {
	versions := []VersionInfo{
		{Version: "1.0.0", Commit: "abc", Date: "2025-01-01"},
		{Version: "dev", Commit: "unknown", Date: "unknown"},
		{Version: "test", Commit: "test", Date: "test"},
	}

	for _, v := range versions {
		s := v.String()
		if s == "" {
			t.Errorf("String() returned empty for version %+v", v)
		}
	}
}

// TestVersionInfoStringConsistency tests that multiple calls return same result
func TestVersionInfoStringConsistency(t *testing.T) {
	v := VersionInfo{
		Version: "1.2.3",
		Commit:  "abc123",
		Date:    "2025-01-01",
	}

	s1 := v.String()
	s2 := v.String()
	s3 := v.String()

	if s1 != s2 {
		t.Errorf("String() not consistent: %q != %q", s1, s2)
	}
	if s1 != s3 {
		t.Errorf("String() not consistent: %q != %q", s1, s3)
	}
}
