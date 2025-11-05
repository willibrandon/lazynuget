//go:build windows

package platform

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	kernel32          = syscall.NewLazyDLL("kernel32.dll")
	getACPProc        = kernel32.NewProc("GetACP")
	getConsoleOCPProc = kernel32.NewProc("GetConsoleOutputCP")
)

// getSystemEncoding detects the system encoding on Windows
// Uses GetACP() for console applications
// See: T078, FR-030
func getSystemEncoding() (string, error) {
	// Try GetConsoleOutputCP first (more specific for console apps)
	ret, _, err := getConsoleOCPProc.Call()
	if ret != 0 {
		codePage := uint32(ret)
		encoding := codePageToEncoding(codePage)
		return encoding, nil
	}

	// Fallback to GetACP (ANSI code page)
	ret, _, err = getACPProc.Call()
	if ret == 0 {
		return "", fmt.Errorf("failed to get code page: %w", err)
	}

	codePage := uint32(ret)
	encoding := codePageToEncoding(codePage)
	return encoding, nil
}

// codePageToEncoding maps Windows code page IDs to encoding names
// See: T078, FR-030
func codePageToEncoding(codePage uint32) string {
	// Map common Windows code pages to encoding names
	// Reference: https://docs.microsoft.com/en-us/windows/win32/intl/code-page-identifiers
	switch codePage {
	case 65001:
		return "utf-8"
	case 1252:
		return "windows-1252" // Western European
	case 1251:
		return "windows-1251" // Cyrillic
	case 1250:
		return "windows-1250" // Central European
	case 1253:
		return "windows-1253" // Greek
	case 1254:
		return "windows-1254" // Turkish
	case 1255:
		return "windows-1255" // Hebrew
	case 1256:
		return "windows-1256" // Arabic
	case 1257:
		return "windows-1257" // Baltic
	case 1258:
		return "windows-1258" // Vietnamese
	case 932:
		return "shift-jis" // Japanese
	case 936:
		return "gbk" // Simplified Chinese
	case 949:
		return "euc-kr" // Korean
	case 950:
		return "big5" // Traditional Chinese
	case 874:
		return "windows-874" // Thai
	case 28591:
		return "iso-8859-1" // Latin-1
	case 28592:
		return "iso-8859-2" // Latin-2
	case 28595:
		return "iso-8859-5" // Cyrillic
	case 28597:
		return "iso-8859-7" // Greek
	case 28599:
		return "iso-8859-9" // Turkish
	default:
		// Unknown code page, fallback to UTF-8
		return "utf-8"
	}
}

// getActiveCodePage returns the active console code page
// Used for diagnostics and testing
func getActiveCodePage() uint32 {
	ret, _, _ := getConsoleOCPProc.Call()
	if ret != 0 {
		return uint32(ret)
	}

	ret, _, _ = getACPProc.Call()
	return uint32(ret)
}

// Ensure these functions are properly linked
var _ = unsafe.Pointer(nil)
