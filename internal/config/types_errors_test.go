package config

import (
	"testing"
)

// TestValidationErrorError tests ValidationError.Error() method
func TestValidationErrorError(t *testing.T) {
	tests := []struct {
		name    string
		err     *ValidationError
		wantStr string
	}{
		{
			name: "error severity validation error",
			err: &ValidationError{
				Key:        "logLevel",
				Value:      "invalid",
				Constraint: "must be one of: debug, info, warn, error",
				Severity:   "error",
			},
			wantStr: `logLevel: must be one of: debug, info, warn, error`,
		},
		{
			name: "warning severity with default",
			err: &ValidationError{
				Key:         "maxConcurrentOps",
				Value:       100,
				Constraint:  "must be between 1 and 16",
				Severity:    "warning",
				DefaultUsed: 4,
			},
			wantStr: `maxConcurrentOps: must be between 1 and 16 (using default: 4)`,
		},
		{
			name: "validation error without severity (defaults to warning)",
			err: &ValidationError{
				Key:         "theme",
				Value:       nil,
				Constraint:  "must not be nil",
				DefaultUsed: "default",
			},
			wantStr: `theme: must not be nil (using default: default)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.wantStr {
				t.Errorf("Error() = %q, want %q", got, tt.wantStr)
			}
		})
	}
}

// TestEncryptedValueString tests EncryptedValue.String() method
func TestEncryptedValueString(t *testing.T) {
	tests := []struct {
		name      string
		encrypted *EncryptedValue
		wantStr   string
	}{
		{
			name:      "encrypted value with keyID and algorithm",
			encrypted: &EncryptedValue{KeyID: "prod", Algorithm: "AES-256-GCM", Ciphertext: []byte("abc123"), Nonce: []byte("nonce")},
			wantStr:   "EncryptedValue{KeyID=prod, Algorithm=AES-256-GCM, CiphertextLen=6, NonceLen=5}",
		},
		{
			name:      "encrypted value without keyID",
			encrypted: &EncryptedValue{KeyID: "", Algorithm: "AES-256-GCM", Ciphertext: []byte("def456"), Nonce: []byte("abc")},
			wantStr:   "EncryptedValue{KeyID=, Algorithm=AES-256-GCM, CiphertextLen=6, NonceLen=3}",
		},
		{
			name:      "encrypted value with empty ciphertext",
			encrypted: &EncryptedValue{KeyID: "test", Algorithm: "AES-256-GCM", Ciphertext: []byte{}, Nonce: []byte{}},
			wantStr:   "EncryptedValue{KeyID=test, Algorithm=AES-256-GCM, CiphertextLen=0, NonceLen=0}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.encrypted.String()
			if got != tt.wantStr {
				t.Errorf("String() = %q, want %q", got, tt.wantStr)
			}
		})
	}
}
