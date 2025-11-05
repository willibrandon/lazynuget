//go:build windows

package platform

import (
	"os"
	"path/filepath"
)

// getConfigDir returns the Windows config directory: %APPDATA%\lazynuget
func getConfigDir() (string, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return "", &PathError{
			Op:   "ConfigDir",
			Path: "%APPDATA%",
			Err:  "APPDATA environment variable not set",
		}
	}

	return filepath.Join(appData, "lazynuget"), nil
}

// getCacheDir returns the Windows cache directory: %LOCALAPPDATA%\lazynuget
func getCacheDir() (string, error) {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return "", &PathError{
			Op:   "CacheDir",
			Path: "%LOCALAPPDATA%",
			Err:  "LOCALAPPDATA environment variable not set",
		}
	}

	return filepath.Join(localAppData, "lazynuget"), nil
}
