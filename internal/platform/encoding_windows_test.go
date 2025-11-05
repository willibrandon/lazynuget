//go:build windows

package platform

import (
	"testing"
)

// TestDetectEncoding_Windows tests Windows code page detection
// See: T073, FR-030
func TestDetectEncoding_Windows(t *testing.T) {
	tests := []struct {
		name         string
		codePage     uint32
		wantEncoding string
	}{
		{
			name:         "UTF-8 code page",
			codePage:     65001,
			wantEncoding: "utf-8",
		},
		{
			name:         "Windows-1252 (Western European)",
			codePage:     1252,
			wantEncoding: "windows-1252",
		},
		{
			name:         "Shift-JIS (Japanese)",
			codePage:     932,
			wantEncoding: "shift-jis",
		},
		{
			name:         "GBK (Simplified Chinese)",
			codePage:     936,
			wantEncoding: "gbk",
		},
		{
			name:         "Unknown code page (fallback to UTF-8)",
			codePage:     99999,
			wantEncoding: "utf-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the code page to encoding name mapping
			encoding := codePageToEncoding(tt.codePage)

			if encoding != tt.wantEncoding {
				t.Errorf("codePageToEncoding(%d) = %q, want %q", tt.codePage, encoding, tt.wantEncoding)
			}
		})
	}
}

// TestGetSystemEncoding_Windows tests system encoding detection on Windows
// See: T073, FR-030
func TestGetSystemEncoding_Windows(t *testing.T) {
	// This test verifies we can detect the system encoding without error
	encoding, err := getSystemEncoding()
	if err != nil {
		t.Fatalf("getSystemEncoding() returned error: %v", err)
	}

	if encoding == "" {
		t.Error("getSystemEncoding() returned empty string")
	}

	t.Logf("Detected system encoding: %s", encoding)

	// Verify it's a known encoding
	validEncodings := map[string]bool{
		"utf-8":        true,
		"windows-1252": true,
		"shift-jis":    true,
		"gbk":          true,
		"gb18030":      true,
		"euc-kr":       true,
		"windows-1251": true,
		"windows-1250": true,
		"iso-8859-1":   true,
		"iso-8859-2":   true,
	}

	if !validEncodings[encoding] {
		t.Logf("Warning: Detected encoding %q is not in the known list (may be valid but uncommon)", encoding)
	}
}

// TestDecodeBytes_Windows tests decoding with various Windows code pages
// See: T073, T075, FR-030
func TestDecodeBytes_Windows(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		encoding string
		want     string
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
			name:     "Windows-1252 encoded bytes",
			input:    []byte{0xC9, 0x63, 0x6F, 0x6C, 0x65}, // "École" in Windows-1252
			encoding: "windows-1252",
			want:     "École",
		},
		{
			name:     "Invalid UTF-8 (fallback to replacement)",
			input:    []byte{0xFF, 0xFE, 0x00, 0x00},
			encoding: "utf-8",
			want:     "����",
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
