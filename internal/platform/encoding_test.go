package platform

import (
	"testing"
	"unicode/utf8"
)

// TestIsValidUTF8 tests UTF-8 validation
// See: T075, FR-030
func TestIsValidUTF8(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  bool
	}{
		{
			name:  "Valid UTF-8 ASCII",
			input: []byte("Hello, World!"),
			want:  true,
		},
		{
			name:  "Valid UTF-8 with Unicode",
			input: []byte("Hello, 世界!"),
			want:  true,
		},
		{
			name:  "Valid UTF-8 with emoji",
			input: []byte("Hello 👋 World 🌍"),
			want:  true,
		},
		{
			name:  "Invalid UTF-8 (invalid start byte)",
			input: []byte{0xFF, 0xFE},
			want:  false,
		},
		{
			name:  "Invalid UTF-8 (incomplete sequence)",
			input: []byte{0xC3}, // Incomplete 2-byte sequence
			want:  false,
		},
		{
			name:  "Invalid UTF-8 (invalid continuation)",
			input: []byte{0xC3, 0x28}, // Invalid continuation byte
			want:  false,
		},
		{
			name:  "Empty input (valid UTF-8)",
			input: []byte{},
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utf8.Valid(tt.input)

			if got != tt.want {
				t.Errorf("utf8.Valid(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestDecodeBytes_AutoDetection tests automatic UTF-8 detection
// See: T075, FR-030
func TestDecodeBytes_AutoDetection(t *testing.T) {
	tests := []struct {
		name  string
		want  string
		input []byte
	}{
		{
			name:  "Valid UTF-8 (no conversion needed)",
			input: []byte("Hello, World!"),
			want:  "Hello, World!",
		},
		{
			name:  "Valid UTF-8 with Unicode",
			input: []byte("Hello, 世界!"),
			want:  "Hello, 世界!",
		},
		{
			name:  "Valid UTF-8 with emoji",
			input: []byte("Hello 👋"),
			want:  "Hello 👋",
		},
		{
			name:  "Empty input",
			input: []byte{},
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use empty encoding to trigger auto-detection
			got := decodeBytes(tt.input, "")

			if got != tt.want {
				t.Errorf("decodeBytes() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestEncodeToUTF8_Replacement tests replacement character handling
// See: T075, FR-030
func TestEncodeToUTF8_Replacement(t *testing.T) {
	tests := []struct {
		name  string
		want  string
		input []byte
	}{
		{
			name:  "Invalid UTF-8 replaced with U+FFFD",
			input: []byte{0xFF, 0xFE},
			want:  "��",
		},
		{
			name:  "Incomplete UTF-8 sequence replaced",
			input: []byte{0xC3},
			want:  "�",
		},
		{
			name:  "Mixed valid and invalid UTF-8",
			input: []byte("Hello" + string([]byte{0xFF}) + "World"),
			want:  "Hello�World",
		},
		{
			name:  "Multiple invalid bytes",
			input: []byte{0xFF, 0xFE, 0xFD},
			want:  "���",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use decodeBytes which handles replacement properly
			// When encoding is "utf-8" but bytes are invalid, it returns with replacement chars
			got := decodeBytes(tt.input, "utf-8")

			// Verify the output is valid UTF-8
			if !utf8.ValidString(got) {
				t.Errorf("Output is not valid UTF-8: %q", got)
			}

			if got != tt.want {
				t.Errorf("Conversion = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestSetEncoding tests manual encoding override
// See: T075, FR-030
func TestSetEncoding(t *testing.T) {
	spawner := NewProcessSpawner()

	// Initially, encoding should be empty (auto-detect)
	if spawner.(*processSpawner).encoding != "" {
		t.Errorf("Initial encoding should be empty, got %q", spawner.(*processSpawner).encoding)
	}

	// Set encoding
	spawner.SetEncoding("windows-1252")

	if spawner.(*processSpawner).encoding != "windows-1252" {
		t.Errorf("SetEncoding(\"windows-1252\") failed, got %q", spawner.(*processSpawner).encoding)
	}

	// Reset encoding
	spawner.SetEncoding("")

	if spawner.(*processSpawner).encoding != "" {
		t.Errorf("SetEncoding(\"\") should reset to empty, got %q", spawner.(*processSpawner).encoding)
	}
}
