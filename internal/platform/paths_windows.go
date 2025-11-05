//go:build windows

package platform

import (
	"os"
	"path/filepath"
	"strings"
	"unicode"
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

// normalize converts path to Windows-native format:
// - Converts forward slashes to backslashes
// - Uppercases drive letters
// - Removes redundant separators
// - Resolves . and .. segments
func normalize(path string) string {
	// First use filepath.Clean to handle . and .. and normalize separators
	cleaned := filepath.Clean(path)

	// Check if this is a drive letter path (e.g., "c:\..." or "C:\...")
	if len(cleaned) >= 2 && cleaned[1] == ':' {
		// Uppercase the drive letter
		cleaned = strings.ToUpper(cleaned[0:1]) + cleaned[1:]
	}

	// Check if this is a UNC path (e.g., "\\server\share" or "//server/share")
	if len(cleaned) >= 2 && cleaned[0] == '\\' && cleaned[1] == '\\' {
		// UNC path - ensure it starts with exactly \\ (filepath.Clean might have changed it)
		// Find the end of the UNC prefix (after \\server\share)
		parts := strings.Split(cleaned, "\\")
		// parts[0] and parts[1] will be empty strings due to leading \\
		// parts[2] is server name, parts[3] is share name
		if len(parts) >= 4 {
			// Reconstruct with normalized separators
			cleaned = "\\\\" + strings.Join(parts[2:], "\\")
		}
	}

	return cleaned
}

// isAbsolute returns true if the path is absolute on Windows:
// - Starts with drive letter (e.g., "C:\")
// - Starts with UNC path (e.g., "\\server\share")
func isAbsolute(path string) bool {
	// Use filepath.IsAbs which handles both drive letters and UNC paths
	return filepath.IsAbs(path)
}

// validate checks if path is valid on Windows
func validate(path string) error {
	// Basic validation: path cannot be empty
	if path == "" {
		return &PathError{
			Op:   "Validate",
			Path: path,
			Err:  "path cannot be empty",
		}
	}

	// Check for null bytes (invalid in Windows paths)
	if strings.ContainsRune(path, '\x00') {
		return &PathError{
			Op:   "Validate",
			Path: path,
			Err:  "path contains null byte",
		}
	}

	// Check for invalid characters in Windows paths
	// Invalid characters: < > : " | ? *
	// Note: We allow : for drive letters (C:) but not elsewhere
	invalidChars := []rune{'<', '>', '"', '|', '?', '*'}
	for _, ch := range invalidChars {
		if strings.ContainsRune(path, ch) {
			return &PathError{
				Op:   "Validate",
				Path: path,
				Err:  "path contains invalid character: " + string(ch),
			}
		}
	}

	// Check for colon outside of drive letter position
	colonIdx := strings.IndexRune(path, ':')
	if colonIdx != -1 {
		// Colon is only valid at position 1 for drive letters (e.g., "C:")
		if colonIdx != 1 {
			return &PathError{
				Op:   "Validate",
				Path: path,
				Err:  "colon only allowed in drive letter position",
			}
		}
		// Verify it's actually a drive letter (single letter before colon)
		if !unicode.IsLetter(rune(path[0])) {
			return &PathError{
				Op:   "Validate",
				Path: path,
				Err:  "invalid drive letter",
			}
		}
	}

	// Check for reserved device names (CON, PRN, AUX, NUL, COM1-9, LPT1-9)
	// These can appear as full path components or with extensions
	reservedNames := []string{
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	}

	// Split path into components
	parts := strings.Split(filepath.Clean(path), string(filepath.Separator))
	for _, part := range parts {
		if part == "" {
			continue
		}

		// Check if part is a reserved name (with or without extension)
		baseName := part
		if idx := strings.LastIndex(part, "."); idx != -1 {
			baseName = part[:idx]
		}

		baseNameUpper := strings.ToUpper(baseName)
		for _, reserved := range reservedNames {
			if baseNameUpper == reserved {
				return &PathError{
					Op:   "Validate",
					Path: path,
					Err:  "path contains reserved device name: " + reserved,
				}
			}
		}

		// Check for paths ending with space or period (invalid on Windows)
		if len(part) > 0 {
			lastChar := part[len(part)-1]
			if lastChar == ' ' || lastChar == '.' {
				return &PathError{
					Op:   "Validate",
					Path: path,
					Err:  "path component cannot end with space or period",
				}
			}
		}
	}

	return nil
}
