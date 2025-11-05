//go:build !windows

package platform

import (
	"os"
	"testing"
)

// TestDetectEncoding_Unix tests Unix locale-based encoding detection
// See: T074, FR-030
func TestDetectEncoding_Unix(t *testing.T) {
	tests := []struct {
		name         string
		locale       string
		wantEncoding string
	}{
		{
			name:         "UTF-8 locale (en_US.UTF-8)",
			locale:       "en_US.UTF-8",
			wantEncoding: "utf-8",
		},
		{
			name:         "UTF-8 locale (C.UTF-8)",
			locale:       "C.UTF-8",
			wantEncoding: "utf-8",
		},
		{
			name:         "ISO-8859-1 locale",
			locale:       "en_US.ISO-8859-1",
			wantEncoding: "iso-8859-1",
		},
		{
			name:         "ISO-8859-1 (latin1 variant)",
			locale:       "en_US.latin1",
			wantEncoding: "iso-8859-1",
		},
		{
			name:         "EUC-JP (Japanese)",
			locale:       "ja_JP.eucJP",
			wantEncoding: "euc-jp",
		},
		{
			name:         "C locale (fallback to UTF-8)",
			locale:       "C",
			wantEncoding: "utf-8",
		},
		{
			name:         "POSIX locale (fallback to UTF-8)",
			locale:       "POSIX",
			wantEncoding: "utf-8",
		},
		{
			name:         "Empty locale (fallback to UTF-8)",
			locale:       "",
			wantEncoding: "utf-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoding := localeToEncoding(tt.locale)

			if encoding != tt.wantEncoding {
				t.Errorf("localeToEncoding(%q) = %q, want %q", tt.locale, encoding, tt.wantEncoding)
			}
		})
	}
}

// TestGetSystemEncoding_Unix tests system encoding detection on Unix
// See: T074, FR-030
func TestGetSystemEncoding_Unix(t *testing.T) {
	// Save original environment variables
	origLCALL := os.Getenv("LC_ALL")
	origLCCTYPE := os.Getenv("LC_CTYPE")
	origLANG := os.Getenv("LANG")

	// Restore after test
	defer func() {
		os.Setenv("LC_ALL", origLCALL)
		os.Setenv("LC_CTYPE", origLCCTYPE)
		os.Setenv("LANG", origLANG)
	}()

	tests := []struct {
		name         string
		lcAll        string
		lcCType      string
		lang         string
		wantEncoding string
	}{
		{
			name:         "LC_ALL takes precedence",
			lcAll:        "en_US.UTF-8",
			lcCType:      "en_US.ISO-8859-1",
			lang:         "C",
			wantEncoding: "utf-8",
		},
		{
			name:         "LC_CTYPE used if LC_ALL not set",
			lcAll:        "",
			lcCType:      "en_US.ISO-8859-1",
			lang:         "en_US.UTF-8",
			wantEncoding: "iso-8859-1",
		},
		{
			name:         "LANG used if neither LC_ALL nor LC_CTYPE set",
			lcAll:        "",
			lcCType:      "",
			lang:         "en_US.UTF-8",
			wantEncoding: "utf-8",
		},
		{
			name:         "Fallback to UTF-8 if all empty",
			lcAll:        "",
			lcCType:      "",
			lang:         "",
			wantEncoding: "utf-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set test environment
			os.Setenv("LC_ALL", tt.lcAll)
			os.Setenv("LC_CTYPE", tt.lcCType)
			os.Setenv("LANG", tt.lang)

			encoding, err := getSystemEncoding()
			if err != nil {
				t.Fatalf("getSystemEncoding() returned error: %v", err)
			}

			if encoding != tt.wantEncoding {
				t.Errorf("getSystemEncoding() = %q, want %q", encoding, tt.wantEncoding)
			}
		})
	}
}

// TestDecodeBytes_Unix tests decoding with various Unix encodings
// See: T074, T075, FR-030
func TestDecodeBytes_Unix(t *testing.T) {
	tests := []struct {
		name     string
		encoding string
		want     string
		input    []byte
	}{
		{
			name:     "Valid UTF-8",
			input:    []byte("Hello, World!"),
			encoding: "utf-8",
			want:     "Hello, World!",
		},
		{
			name:     "Valid UTF-8 with Unicode",
			input:    []byte("Hello, 世界!"),
			encoding: "utf-8",
			want:     "Hello, 世界!",
		},
		{
			name:     "ISO-8859-1 encoded bytes",
			input:    []byte{0xC9, 0x63, 0x6F, 0x6C, 0x65}, // "École" in ISO-8859-1
			encoding: "iso-8859-1",
			want:     "École",
		},
		{
			name:     "Invalid UTF-8 (fallback to replacement)",
			input:    []byte{0xFF, 0xFE, 0x00, 0x00},
			encoding: "utf-8",
			want:     "��\x00\x00", // 0xFF and 0xFE are invalid, but 0x00 (null) is valid UTF-8
		},
		{
			name:     "Empty input",
			input:    []byte{},
			encoding: "utf-8",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := decodeBytes(tt.input, tt.encoding)

			if got != tt.want {
				t.Errorf("decodeBytes() = %q, want %q", got, tt.want)
			}
		})
	}
}
