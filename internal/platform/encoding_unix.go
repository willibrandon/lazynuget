//go:build !windows

package platform

import (
	"os"
	"strings"
)

// getSystemEncoding detects the system encoding on Unix systems
// Parses locale from LC_ALL, LC_CTYPE, or LANG environment variables
// See: T079, FR-030
func getSystemEncoding() (string, error) {
	// Check environment variables in order of precedence
	// LC_ALL > LC_CTYPE > LANG
	locale := os.Getenv("LC_ALL")
	if locale == "" {
		locale = os.Getenv("LC_CTYPE")
	}
	if locale == "" {
		locale = os.Getenv("LANG")
	}

	return localeToEncoding(locale), nil
}

// localeToEncoding extracts encoding from Unix locale string
// Examples: en_US.UTF-8 → utf-8, en_US.ISO-8859-1 → iso-8859-1
// See: T079, FR-030
func localeToEncoding(locale string) string {
	// Handle empty, C, or POSIX locales (fallback to UTF-8)
	if locale == "" || locale == "C" || locale == "POSIX" {
		return "utf-8"
	}

	// Extract encoding part after the dot
	// Format: language_COUNTRY.ENCODING[@modifier]
	parts := strings.Split(locale, ".")
	if len(parts) < 2 {
		// No encoding specified, assume UTF-8
		return "utf-8"
	}

	encoding := parts[1]

	// Remove optional modifier (e.g., @euro, @latin)
	if idx := strings.Index(encoding, "@"); idx >= 0 {
		encoding = encoding[:idx]
	}

	// Normalize encoding name to lowercase and handle variants
	encoding = strings.ToLower(strings.TrimSpace(encoding))

	// Map common locale encoding names to standard names
	switch encoding {
	case "utf-8", "utf8":
		return "utf-8"
	case "iso-8859-1", "iso8859-1", "latin1", "latin-1":
		return "iso-8859-1"
	case "iso-8859-2", "iso8859-2", "latin2", "latin-2":
		return "iso-8859-2"
	case "iso-8859-5", "iso8859-5":
		return "iso-8859-5"
	case "iso-8859-7", "iso8859-7":
		return "iso-8859-7"
	case "iso-8859-9", "iso8859-9":
		return "iso-8859-9"
	case "iso-8859-15", "iso8859-15", "latin9", "latin-9":
		return "iso-8859-15"
	case "euc-jp", "eucjp":
		return "euc-jp"
	case "shift-jis", "shift_jis", "sjis":
		return "shift-jis"
	case "gbk", "cp936":
		return "gbk"
	case "gb18030":
		return "gb18030"
	case "euc-kr", "euckr":
		return "euc-kr"
	case "big5":
		return "big5"
	case "koi8-r", "koi8r":
		return "koi8-r"
	case "koi8-u", "koi8u":
		return "koi8-u"
	default:
		// Unknown encoding, fallback to UTF-8
		return "utf-8"
	}
}
