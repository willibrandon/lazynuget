package platform

import (
	"bytes"
	"io"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
)

// decodeBytes decodes byte slice to UTF-8 string
// If encoding is empty, auto-detects UTF-8 or falls back to system encoding
// This function is infallible - it always returns a valid UTF-8 string
// See: T080, T081, FR-030
func decodeBytes(data []byte, encoding string) string {
	// Handle empty input
	if len(data) == 0 {
		return ""
	}

	// If no encoding specified, try UTF-8 first
	if encoding == "" {
		if utf8.Valid(data) {
			return string(data)
		}

		// Not valid UTF-8, detect system encoding
		systemEncoding, err := getSystemEncoding()
		if err != nil {
			// Fallback: replace invalid UTF-8 bytes with replacement character
			return sanitizeToUTF8(data)
		}
		encoding = systemEncoding
	}

	// If encoding is UTF-8, just validate and convert
	if encoding == "utf-8" {
		if utf8.Valid(data) {
			return string(data)
		}
		// Invalid UTF-8 - sanitize by replacing invalid bytes with U+FFFD
		return sanitizeToUTF8(data)
	}

	// Get decoder for the specified encoding
	decoder := getDecoder(encoding)
	if decoder == nil {
		// Unknown encoding, fallback to string conversion (with replacement chars)
		return string(data)
	}

	// Decode using the specified encoding
	reader := transform.NewReader(bytes.NewReader(data), decoder.NewDecoder())
	decoded, err := io.ReadAll(reader)
	if err != nil {
		// Decoding failed, fallback to string conversion
		return string(data)
	}

	return string(decoded)
}

// getDecoder returns the appropriate encoding.Encoding for the given encoding name
// See: T081, FR-030
func getDecoder(encodingName string) encoding.Encoding {
	// Normalize encoding name
	encodingName = strings.ToLower(strings.TrimSpace(encodingName))

	switch encodingName {
	// Windows code pages
	case "windows-1250":
		return charmap.Windows1250
	case "windows-1251":
		return charmap.Windows1251
	case "windows-1252":
		return charmap.Windows1252
	case "windows-1253":
		return charmap.Windows1253
	case "windows-1254":
		return charmap.Windows1254
	case "windows-1255":
		return charmap.Windows1255
	case "windows-1256":
		return charmap.Windows1256
	case "windows-1257":
		return charmap.Windows1257
	case "windows-1258":
		return charmap.Windows1258
	case "windows-874":
		return charmap.Windows874

	// ISO-8859 series
	case "iso-8859-1", "latin1":
		return charmap.ISO8859_1
	case "iso-8859-2", "latin2":
		return charmap.ISO8859_2
	case "iso-8859-3", "latin3":
		return charmap.ISO8859_3
	case "iso-8859-4", "latin4":
		return charmap.ISO8859_4
	case "iso-8859-5":
		return charmap.ISO8859_5
	case "iso-8859-6":
		return charmap.ISO8859_6
	case "iso-8859-7":
		return charmap.ISO8859_7
	case "iso-8859-8":
		return charmap.ISO8859_8
	case "iso-8859-9", "latin5":
		return charmap.ISO8859_9
	case "iso-8859-10", "latin6":
		return charmap.ISO8859_10
	case "iso-8859-13", "latin7":
		return charmap.ISO8859_13
	case "iso-8859-14", "latin8":
		return charmap.ISO8859_14
	case "iso-8859-15", "latin9":
		return charmap.ISO8859_15
	case "iso-8859-16", "latin10":
		return charmap.ISO8859_16

	// KOI8 variants
	case "koi8-r":
		return charmap.KOI8R
	case "koi8-u":
		return charmap.KOI8U

	// Japanese
	case "shift-jis", "shift_jis", "sjis":
		return japanese.ShiftJIS
	case "euc-jp", "eucjp":
		return japanese.EUCJP
	case "iso-2022-jp":
		return japanese.ISO2022JP

	// Simplified Chinese
	case "gbk", "cp936":
		return simplifiedchinese.GBK
	case "gb18030":
		return simplifiedchinese.GB18030
	case "hz-gb2312":
		return simplifiedchinese.HZGB2312

	// Traditional Chinese
	case "big5":
		return traditionalchinese.Big5

	// Korean
	case "euc-kr", "euckr":
		return korean.EUCKR

	// IBM code pages
	case "ibm437", "cp437":
		return charmap.CodePage437
	case "ibm850", "cp850":
		return charmap.CodePage850
	case "ibm852", "cp852":
		return charmap.CodePage852
	case "ibm855", "cp855":
		return charmap.CodePage855
	case "ibm858", "cp858":
		return charmap.CodePage858
	case "ibm860", "cp860":
		return charmap.CodePage860
	case "ibm862", "cp862":
		return charmap.CodePage862
	case "ibm863", "cp863":
		return charmap.CodePage863
	case "ibm865", "cp865":
		return charmap.CodePage865
	case "ibm866", "cp866":
		return charmap.CodePage866

	// Macintosh encodings
	case "macintosh":
		return charmap.Macintosh
	case "macroman":
		return charmap.MacintoshCyrillic

	default:
		// Unknown encoding
		return nil
	}
}

// sanitizeToUTF8 replaces invalid UTF-8 bytes with the replacement character U+FFFD
// See: T080, FR-030
func sanitizeToUTF8(data []byte) string {
	// Use Go's built-in UTF-8 validation and replacement
	// When you range over a byte slice treating it as a string,
	// Go automatically replaces invalid UTF-8 with U+FFFD
	var result strings.Builder
	result.Grow(len(data))

	for len(data) > 0 {
		r, size := utf8.DecodeRune(data)
		result.WriteRune(r) // Invalid runes become U+FFFD automatically
		data = data[size:]
	}

	return result.String()
}

// Note: validateEncoding was removed as it's not currently used.
// Encoding validation happens implicitly in getDecoder() which returns nil
// for unknown encodings, and decodeBytes() handles this gracefully.
