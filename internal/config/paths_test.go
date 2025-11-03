package config

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestGetPlatformConfigPath tests platform-specific config path detection
func TestGetPlatformConfigPath(t *testing.T) {
	path := getPlatformConfigPath()

	// Path may be empty if environment variables are not set (HOME, XDG_CONFIG_HOME, APPDATA)
	// This is valid behavior - the application will fall back to defaults
	if path == "" {
		return
	}

	// Path should be absolute
	if !filepath.IsAbs(path) {
		t.Errorf("Expected absolute path, got: %s", path)
	}

	// Path should contain "lazynuget" directory
	if !strings.Contains(path, "lazynuget") {
		t.Errorf("Expected path to contain 'lazynuget', got: %s", path)
	}

	// Verify platform-specific behavior
	switch runtime.GOOS {
	case "darwin":
		// macOS: should be under ~/Library/Application Support
		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatal("UserHomeDir failed on macOS")
		}
		expected := filepath.Join(home, "Library", "Application Support", "lazynuget")
		if path != expected {
			t.Errorf("macOS path should be %s, got: %s", expected, path)
		}
	case "linux":
		// Linux: should be under ~/.config (or XDG_CONFIG_HOME)
		if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
			expected := filepath.Join(xdgConfig, "lazynuget")
			if path != expected {
				t.Errorf("Linux path should be %s (XDG_CONFIG_HOME), got: %s", expected, path)
			}
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				t.Fatal("UserHomeDir failed on Linux")
			}
			expected := filepath.Join(home, ".config", "lazynuget")
			if path != expected {
				t.Errorf("Linux path should be %s, got: %s", expected, path)
			}
		}
	case "windows":
		// Windows: should be under %APPDATA%
		appData := os.Getenv("APPDATA")
		if appData == "" {
			t.Fatal("APPDATA environment variable not set on Windows")
		}
		expected := filepath.Join(appData, "lazynuget")
		if path != expected {
			t.Errorf("Windows path should be %s, got: %s", expected, path)
		}
	}
}

// TestGetPlatformConfigPathReturnsDirectory tests that the path is a directory path
func TestGetPlatformConfigPathReturnsDirectory(t *testing.T) {
	path := getPlatformConfigPath()

	// Path may be empty if environment variables are not set (HOME, XDG_CONFIG_HOME, APPDATA)
	// This is valid behavior - the application will fall back to defaults
	if path == "" {
		return
	}

	// Path should not have a file extension
	ext := filepath.Ext(path)
	if ext != "" {
		t.Errorf("Expected directory path (no extension), got path with extension: %s", ext)
	}

	// Path should end with "lazynuget"
	base := filepath.Base(path)
	if base != "lazynuget" {
		t.Errorf("Expected path to end with 'lazynuget', got: %s", base)
	}
}

// TestGetPlatformConfigPathCrossPlatform tests cross-platform compatibility
func TestGetPlatformConfigPathCrossPlatform(t *testing.T) {
	// Save original environment
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	origAppData := os.Getenv("APPDATA")
	defer func() {
		if origXDG != "" {
			os.Setenv("XDG_CONFIG_HOME", origXDG)
		}
		if origAppData != "" {
			os.Setenv("APPDATA", origAppData)
		}
	}()

	// Test with custom XDG_CONFIG_HOME on Linux
	if runtime.GOOS == "linux" {
		customXDG := "/tmp/custom-config"
		os.Setenv("XDG_CONFIG_HOME", customXDG)

		path := getPlatformConfigPath()
		if !strings.Contains(path, customXDG) {
			t.Errorf("Expected path to use XDG_CONFIG_HOME (%s), got: %s", customXDG, path)
		}
	}

	// Test with custom APPDATA on Windows
	if runtime.GOOS == "windows" {
		customAppData := "C:\\CustomAppData"
		os.Setenv("APPDATA", customAppData)

		path := getPlatformConfigPath()
		if !strings.Contains(path, customAppData) {
			t.Errorf("Expected path to use APPDATA (%s), got: %s", customAppData, path)
		}
	}
}

// TestGetPlatformConfigPathConsistency tests that the function returns consistent results
func TestGetPlatformConfigPathConsistency(t *testing.T) {
	// Call multiple times
	path1 := getPlatformConfigPath()
	path2 := getPlatformConfigPath()
	path3 := getPlatformConfigPath()

	// Should return same path every time
	if path1 != path2 {
		t.Errorf("Inconsistent results: %s != %s", path1, path2)
	}
	if path1 != path3 {
		t.Errorf("Inconsistent results: %s != %s", path1, path3)
	}
}

// TestGetPlatformConfigPathEmptyHandling tests that empty path is handled gracefully
func TestGetPlatformConfigPathEmptyHandling(t *testing.T) {
	// Note: getPlatformConfigPath() returns empty string when UserHomeDir() fails
	// The config system handles this by using defaults (no config file)
	// This is tested in TestConfigLoaderLoadFromDefaultLocation in config_test.go

	path := getPlatformConfigPath()

	// On normal systems, path should not be empty
	// If it is empty, that's still valid behavior (fallback to defaults)
	if path != "" && !filepath.IsAbs(path) {
		t.Errorf("Non-empty path should be absolute, got: %s", path)
	}
}

// TestConfigPathIntegrationWithLoad tests that config loading uses platform path
func TestConfigPathIntegrationWithLoad(t *testing.T) {
	loader := NewLoader()

	// Load without specifying config path - should use platform default
	opts := LoadOptions{
		EnvVarPrefix: "LAZYNUGET_",
		StrictMode:   false,
	}

	cfg, err := loader.Load(context.Background(), opts)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg == nil {
		t.Fatal("Load returned nil config")
	}

	// Should have default values since no config file exists at platform path
	if cfg.LogLevel != "info" {
		t.Errorf("Expected default LogLevel=info, got %s", cfg.LogLevel)
	}
}
