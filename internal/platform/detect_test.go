package platform

import (
	"os"
	"os/exec"
	"runtime"
	"slices"
	"testing"
)

// TestNew tests the New constructor
func TestNew(t *testing.T) {
	platform := New()

	if platform == nil {
		t.Fatal("New returned nil")
	}

	if platform.OS() != runtime.GOOS {
		t.Errorf("Expected OS=%s, got %s", runtime.GOOS, platform.OS())
	}

	if platform.Arch() != runtime.GOARCH {
		t.Errorf("Expected Arch=%s, got %s", runtime.GOARCH, platform.Arch())
	}
}

// TestPlatformOS tests OS method
func TestPlatformOS(t *testing.T) {
	platform := New()
	os := platform.OS()

	if os == "" {
		t.Error("OS returned empty string")
	}

	// OS should be one of the known values
	validOS := []string{"linux", "darwin", "windows", "freebsd", "openbsd", "netbsd"}
	found := slices.Contains(validOS, os)
	if !found {
		t.Logf("Warning: OS %q not in known list (this may be OK)", os)
	}
}

// TestPlatformArch tests Arch method
func TestPlatformArch(t *testing.T) {
	platform := New()
	arch := platform.Arch()

	if arch == "" {
		t.Error("Arch returned empty string")
	}

	// Arch should be one of the common values
	validArch := []string{"amd64", "arm64", "386", "arm", "ppc64le", "s390x"}
	found := slices.Contains(validArch, arch)
	if !found {
		t.Logf("Warning: Arch %q not in known list (this may be OK)", arch)
	}
}

// TestDetect tests platform detection
func TestDetect(t *testing.T) {
	info := Detect()

	if info == nil {
		t.Fatal("Detect returned nil")
	}

	if info.OS == "" {
		t.Error("Detect returned empty OS")
	}

	if info.Arch == "" {
		t.Error("Detect returned empty Arch")
	}

	// IsCI, IsDumbTerminal, NoColor are booleans - they have valid values
	// Just verify the struct is populated
	t.Logf("Detected: OS=%s Arch=%s IsCI=%v IsDumbTerminal=%v NoColor=%v",
		info.OS, info.Arch, info.IsCI, info.IsDumbTerminal, info.NoColor)
}

// TestDetectCI tests CI detection
func TestDetectCI(t *testing.T) {
	tests := []struct {
		envVars map[string]string
		name    string
		want    bool
	}{
		{
			name:    "no CI env vars",
			envVars: map[string]string{},
			want:    false,
		},
		{
			name: "CI=true",
			envVars: map[string]string{
				"CI": "true",
			},
			want: true,
		},
		{
			name: "CI=1",
			envVars: map[string]string{
				"CI": "1",
			},
			want: true,
		},
		{
			name: "CI=yes",
			envVars: map[string]string{
				"CI": "yes",
			},
			want: true,
		},
		{
			name: "CI=false",
			envVars: map[string]string{
				"CI": "false",
			},
			want: false,
		},
		{
			name: "GITHUB_ACTIONS=true",
			envVars: map[string]string{
				"GITHUB_ACTIONS": "true",
			},
			want: true,
		},
		{
			name: "GITLAB_CI=1",
			envVars: map[string]string{
				"GITLAB_CI": "1",
			},
			want: true,
		},
		{
			name: "TRAVIS=yes",
			envVars: map[string]string{
				"TRAVIS": "yes",
			},
			want: true,
		},
		{
			name: "CIRCLECI=true",
			envVars: map[string]string{
				"CIRCLECI": "true",
			},
			want: true,
		},
		{
			name: "BUILD_NUMBER=123 (Jenkins)",
			envVars: map[string]string{
				"BUILD_NUMBER": "123",
			},
			want: false, // BUILD_NUMBER needs to be "true", "1", or "yes"
		},
		{
			name: "TF_BUILD=true (Azure)",
			envVars: map[string]string{
				"TF_BUILD": "true",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all CI env vars
			ciVars := []string{"CI", "CONTINUOUS_INTEGRATION", "BUILD_NUMBER", "GITLAB_CI", "TRAVIS", "CIRCLECI", "GITHUB_ACTIONS", "TF_BUILD"}
			for _, v := range ciVars {
				os.Unsetenv(v)
			}

			// Set test env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer func() {
				for k := range tt.envVars {
					os.Unsetenv(k)
				}
			}()

			got := detectCI()
			if got != tt.want {
				t.Errorf("detectCI() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsDumbTerminal tests dumb terminal detection
func TestIsDumbTerminal(t *testing.T) {
	tests := []struct {
		name string
		term string
		want bool
	}{
		{
			name: "empty TERM",
			term: "",
			want: false,
		},
		{
			name: "TERM=dumb",
			term: "dumb",
			want: true,
		},
		{
			name: "TERM=DUMB (uppercase)",
			term: "DUMB",
			want: true,
		},
		{
			name: "TERM=xterm",
			term: "xterm",
			want: false,
		},
		{
			name: "TERM=xterm-256color",
			term: "xterm-256color",
			want: false,
		},
	}

	origTerm := os.Getenv("TERM")
	defer os.Setenv("TERM", origTerm)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("TERM", tt.term)
			got := isDumbTerminal()
			if got != tt.want {
				t.Errorf("isDumbTerminal() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsNoColor tests NO_COLOR detection
func TestIsNoColor(t *testing.T) {
	tests := []struct {
		name   string
		value  string
		setVar bool
		want   bool
	}{
		{
			name:   "NO_COLOR not set",
			setVar: false,
			want:   false,
		},
		{
			name:   "NO_COLOR set to empty",
			setVar: true,
			value:  "",
			want:   true,
		},
		{
			name:   "NO_COLOR set to 1",
			setVar: true,
			value:  "1",
			want:   true,
		},
		{
			name:   "NO_COLOR set to any value",
			setVar: true,
			value:  "anything",
			want:   true,
		},
	}

	origNoColor := os.Getenv("NO_COLOR")
	defer func() {
		if origNoColor != "" {
			os.Setenv("NO_COLOR", origNoColor)
		} else {
			os.Unsetenv("NO_COLOR")
		}
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv("NO_COLOR")
			if tt.setVar {
				os.Setenv("NO_COLOR", tt.value)
			}

			got := isNoColor()
			if got != tt.want {
				t.Errorf("isNoColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDetermineRunMode tests run mode determination
func TestDetermineRunMode(t *testing.T) {
	tests := []struct {
		setupEnv           func()
		cleanupEnv         func()
		name               string
		want               RunMode
		nonInteractiveFlag bool
	}{
		{
			name:               "explicit non-interactive flag",
			nonInteractiveFlag: true,
			setupEnv:           func() {},
			cleanupEnv:         func() {},
			want:               RunModeNonInteractive,
		},
		{
			name:               "CI environment",
			nonInteractiveFlag: false,
			setupEnv: func() {
				os.Setenv("CI", "true")
			},
			cleanupEnv: func() {
				os.Unsetenv("CI")
			},
			want: RunModeNonInteractive,
		},
		{
			name:               "dumb terminal",
			nonInteractiveFlag: false,
			setupEnv: func() {
				os.Setenv("TERM", "dumb")
			},
			cleanupEnv: func() {
				os.Unsetenv("TERM")
			},
			want: RunModeNonInteractive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			defer tt.cleanupEnv()

			got := DetermineRunMode(tt.nonInteractiveFlag)
			if got != tt.want {
				t.Errorf("DetermineRunMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestValidateDotnetCLI tests dotnet CLI validation
func TestValidateDotnetCLI(t *testing.T) {
	// Check if dotnet is available
	_, err := exec.LookPath("dotnet")
	if err != nil {
		t.Skip("dotnet not available in PATH, skipping validation tests")
	}

	// Test successful validation
	err = ValidateDotnetCLI()
	if err != nil {
		t.Errorf("ValidateDotnetCLI() failed when dotnet is available: %v", err)
	}
}

// TestValidateDotnetCLIMissing tests error when dotnet is missing
func TestValidateDotnetCLIMissing(t *testing.T) {
	// Save and clear PATH to simulate missing dotnet
	origPath := os.Getenv("PATH")
	defer os.Setenv("PATH", origPath)

	// Set PATH to empty to ensure dotnet is not found
	os.Setenv("PATH", "")

	err := ValidateDotnetCLI()
	if err == nil {
		t.Error("Expected error when dotnet is not in PATH, got nil")
	}

	// Error should mention installation instructions
	errMsg := err.Error()
	if !contains(errMsg, "not found") {
		t.Errorf("Error should mention 'not found', got: %s", errMsg)
	}
	if !contains(errMsg, "Installation instructions") {
		t.Errorf("Error should include installation instructions, got: %s", errMsg)
	}
}

// TestPlatformInfoFields tests PlatformInfo struct fields
func TestPlatformInfoFields(t *testing.T) {
	info := &PlatformInfo{
		OS:             "linux",
		Arch:           "amd64",
		IsCI:           true,
		IsDumbTerminal: false,
		NoColor:        true,
	}

	if info.OS != "linux" {
		t.Errorf("Expected OS=linux, got %s", info.OS)
	}
	if info.Arch != "amd64" {
		t.Errorf("Expected Arch=amd64, got %s", info.Arch)
	}
	if !info.IsCI {
		t.Error("Expected IsCI=true")
	}
	if info.IsDumbTerminal {
		t.Error("Expected IsDumbTerminal=false")
	}
	if !info.NoColor {
		t.Error("Expected NoColor=true")
	}
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && (s[:len(substr)] == substr || contains(s[1:], substr))))
}
