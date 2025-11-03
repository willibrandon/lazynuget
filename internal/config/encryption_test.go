package config

import (
	"context"
	"encoding/hex"
	"os"
	"strings"
	"testing"
)

// TestEncryptDecrypt tests encryption and decryption round-trip
func TestEncryptDecrypt(t *testing.T) {
	ctx := context.Background()
	km := NewKeychainManager()
	kd := NewKeyDerivation()
	enc := NewEncryptor(km, kd)

	// Generate a test key
	testKey := make([]byte, 32) // 32 bytes for AES-256
	for i := range testKey {
		testKey[i] = byte(i)
	}

	keyID := "test-encrypt-key"
	envKey := "LAZYNUGET_ENCRYPTION_KEY_" + strings.ToUpper(keyID)

	// Set key in environment variable
	originalValue := os.Getenv(envKey)
	defer func() {
		if originalValue != "" {
			os.Setenv(envKey, originalValue)
		} else {
			os.Unsetenv(envKey)
		}
	}()
	os.Setenv(envKey, hex.EncodeToString(testKey))

	tests := []struct {
		name      string
		plaintext string
	}{
		{
			name:      "simple string",
			plaintext: "hello world",
		},
		{
			name:      "api key",
			plaintext: "sk-1234567890abcdef",
		},
		{
			name:      "long password",
			plaintext: "this_is_a_very_long_password_with_special_chars!@#$%^&*()",
		},
		{
			name:      "empty string",
			plaintext: "",
		},
		{
			name:      "unicode characters",
			plaintext: "Hello 世界 🌍",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			encrypted, err := enc.Encrypt(ctx, tt.plaintext, keyID)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			// Verify encrypted value has correct structure
			if encrypted.KeyID != keyID {
				t.Errorf("KeyID = %q, want %q", encrypted.KeyID, keyID)
			}
			if encrypted.Algorithm != "AES-256-GCM" {
				t.Errorf("Algorithm = %q, want AES-256-GCM", encrypted.Algorithm)
			}
			if len(encrypted.Nonce) != 12 {
				t.Errorf("Nonce length = %d, want 12", len(encrypted.Nonce))
			}
			if len(encrypted.Ciphertext) == 0 {
				t.Error("Ciphertext is empty")
			}

			// Decrypt
			decrypted, err := enc.Decrypt(ctx, encrypted)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			// Verify decrypted matches original
			if decrypted != tt.plaintext {
				t.Errorf("Decrypted = %q, want %q", decrypted, tt.plaintext)
			}
		})
	}
}

// TestEncryptToString tests EncryptToString and DecryptFromString
func TestEncryptToStringDecryptFromString(t *testing.T) {
	ctx := context.Background()
	km := NewKeychainManager()
	kd := NewKeyDerivation()
	enc := NewEncryptor(km, kd)

	// Generate a test key
	testKey := make([]byte, 32)
	for i := range testKey {
		testKey[i] = byte(i)
	}

	keyID := "default"
	envKey := "LAZYNUGET_ENCRYPTION_KEY_" + strings.ToUpper(keyID)

	// Set key in environment variable
	originalValue := os.Getenv(envKey)
	defer func() {
		if originalValue != "" {
			os.Setenv(envKey, originalValue)
		} else {
			os.Unsetenv(envKey)
		}
	}()
	os.Setenv(envKey, hex.EncodeToString(testKey))

	tests := []struct {
		name      string
		plaintext string
	}{
		{
			name:      "api token",
			plaintext: "ghp_1234567890abcdefghijklmnop",
		},
		{
			name:      "database password",
			plaintext: "P@ssw0rd!2024",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt to string
			encryptedStr, err := enc.EncryptToString(ctx, tt.plaintext, keyID)
			if err != nil {
				t.Fatalf("EncryptToString failed: %v", err)
			}

			// Verify format
			if !strings.HasPrefix(encryptedStr, "!encrypted ") {
				t.Errorf("Encrypted string should start with '!encrypted ', got: %s", encryptedStr)
			}

			// Decrypt from string
			decrypted, err := enc.DecryptFromString(ctx, encryptedStr)
			if err != nil {
				t.Fatalf("DecryptFromString failed: %v", err)
			}

			// Verify decrypted matches original
			if decrypted != tt.plaintext {
				t.Errorf("Decrypted = %q, want %q", decrypted, tt.plaintext)
			}
		})
	}
}

// TestDecryptFromStringFormats tests various encrypted string formats
func TestDecryptFromStringFormats(t *testing.T) {
	ctx := context.Background()
	km := NewKeychainManager()
	kd := NewKeyDerivation()
	enc := NewEncryptor(km, kd)

	// Generate a test key
	testKey := make([]byte, 32)
	for i := range testKey {
		testKey[i] = byte(i)
	}

	keyID := "default"
	envKey := "LAZYNUGET_ENCRYPTION_KEY_" + strings.ToUpper(keyID)

	// Set key in environment variable
	originalValue := os.Getenv(envKey)
	defer func() {
		if originalValue != "" {
			os.Setenv(envKey, originalValue)
		} else {
			os.Unsetenv(envKey)
		}
	}()
	os.Setenv(envKey, hex.EncodeToString(testKey))

	plaintext := "test-secret"

	// Encrypt first to get a valid encrypted string
	encryptedStr, err := enc.EncryptToString(ctx, plaintext, keyID)
	if err != nil {
		t.Fatalf("EncryptToString failed: %v", err)
	}

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "standard format with !encrypted prefix",
			input:   encryptedStr,
			wantErr: false,
		},
		{
			name:    "format without !encrypted prefix",
			input:   strings.TrimPrefix(encryptedStr, "!encrypted "),
			wantErr: false,
		},
		{
			name:    "format with extra whitespace",
			input:   "  " + encryptedStr + "  ",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decrypted, err := enc.DecryptFromString(ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecryptFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && decrypted != plaintext {
				t.Errorf("Decrypted = %q, want %q", decrypted, plaintext)
			}
		})
	}
}

// TestEncryptionErrors tests error conditions
func TestEncryptionErrors(t *testing.T) {
	ctx := context.Background()
	km := NewKeychainManager()
	kd := NewKeyDerivation()
	enc := NewEncryptor(km, kd)

	t.Run("encrypt with missing key", func(t *testing.T) {
		keyID := "nonexistent-key-12345"
		envKey := "LAZYNUGET_ENCRYPTION_KEY_" + strings.ToUpper(keyID)
		os.Unsetenv(envKey)

		_, err := enc.Encrypt(ctx, "test", keyID)
		if err == nil {
			t.Error("Encrypt should fail with missing key")
		}
	})

	t.Run("encrypt with wrong key length", func(t *testing.T) {
		keyID := "short-key"
		envKey := "LAZYNUGET_ENCRYPTION_KEY_" + strings.ToUpper(keyID)

		// Set a key that's too short (16 bytes instead of 32)
		shortKey := make([]byte, 16)
		originalValue := os.Getenv(envKey)
		defer func() {
			if originalValue != "" {
				os.Setenv(envKey, originalValue)
			} else {
				os.Unsetenv(envKey)
			}
		}()
		os.Setenv(envKey, hex.EncodeToString(shortKey))

		_, err := enc.Encrypt(ctx, "test", keyID)
		if err == nil {
			t.Error("Encrypt should fail with wrong key length")
		}
		if err != nil && !strings.Contains(err.Error(), "invalid key length") {
			t.Errorf("Expected 'invalid key length' error, got: %v", err)
		}
	})

	t.Run("decrypt with wrong key", func(t *testing.T) {
		// Set up two different keys
		keyID1 := "key1"
		keyID2 := "key2"

		key1 := make([]byte, 32)
		key2 := make([]byte, 32)
		for i := range key1 {
			key1[i] = byte(i)
			key2[i] = byte(i + 1) // Different key
		}

		envKey1 := "LAZYNUGET_ENCRYPTION_KEY_" + strings.ToUpper(keyID1)
		envKey2 := "LAZYNUGET_ENCRYPTION_KEY_" + strings.ToUpper(keyID2)

		orig1 := os.Getenv(envKey1)
		orig2 := os.Getenv(envKey2)
		defer func() {
			if orig1 != "" {
				os.Setenv(envKey1, orig1)
			} else {
				os.Unsetenv(envKey1)
			}
			if orig2 != "" {
				os.Setenv(envKey2, orig2)
			} else {
				os.Unsetenv(envKey2)
			}
		}()

		os.Setenv(envKey1, hex.EncodeToString(key1))
		os.Setenv(envKey2, hex.EncodeToString(key2))

		// Encrypt with key1
		encrypted, err := enc.Encrypt(ctx, "secret", keyID1)
		if err != nil {
			t.Fatalf("Encrypt failed: %v", err)
		}

		// Try to decrypt with key2
		encrypted.KeyID = keyID2
		_, err = enc.Decrypt(ctx, encrypted)
		if err == nil {
			t.Error("Decrypt should fail with wrong key")
		}
	})

	t.Run("decrypt from invalid base64", func(t *testing.T) {
		_, err := enc.DecryptFromString(ctx, "!encrypted not-valid-base64!!!")
		if err == nil {
			t.Error("DecryptFromString should fail with invalid base64")
		}
	})

	t.Run("decrypt from too short data", func(t *testing.T) {
		_, err := enc.DecryptFromString(ctx, "!encrypted AAAA")
		if err == nil {
			t.Error("DecryptFromString should fail with too short data")
		}
	})
}

// TestDeriveKey tests key derivation
func TestDeriveKey(t *testing.T) {
	kd := NewKeyDerivation()

	t.Run("derive key from password", func(t *testing.T) {
		password := "my-secure-password"
		salt := []byte("fixed-salt-16byt")
		iterations := 100000

		key := kd.DeriveKey(password, salt, iterations)

		// Verify key length
		if len(key) != 32 {
			t.Errorf("Derived key length = %d, want 32", len(key))
		}

		// Verify key is deterministic (same inputs produce same key)
		key2 := kd.DeriveKey(password, salt, iterations)
		if string(key) != string(key2) {
			t.Error("DeriveKey should be deterministic")
		}

		// Verify different password produces different key
		key3 := kd.DeriveKey("different-password", salt, iterations)
		if string(key) == string(key3) {
			t.Error("Different passwords should produce different keys")
		}

		// Verify different salt produces different key
		differentSalt := []byte("different-salt16")
		key4 := kd.DeriveKey(password, differentSalt, iterations)
		if string(key) == string(key4) {
			t.Error("Different salts should produce different keys")
		}
	})
}

// TestGenerateSalt tests salt generation
func TestGenerateSalt(t *testing.T) {
	kd := NewKeyDerivation()

	t.Run("generate salt", func(t *testing.T) {
		salt, err := kd.GenerateSalt()
		if err != nil {
			t.Fatalf("GenerateSalt failed: %v", err)
		}

		// Verify salt length
		if len(salt) != 16 {
			t.Errorf("Salt length = %d, want 16", len(salt))
		}

		// Verify salt is random (generate another and compare)
		salt2, err := kd.GenerateSalt()
		if err != nil {
			t.Fatalf("GenerateSalt failed: %v", err)
		}

		if string(salt) == string(salt2) {
			t.Error("GenerateSalt should produce different salts")
		}
	})
}

// TestNewEncryptor tests encryptor creation
func TestNewEncryptor(t *testing.T) {
	km := NewKeychainManager()
	kd := NewKeyDerivation()
	enc := NewEncryptor(km, kd)

	if enc == nil {
		t.Fatal("NewEncryptor returned nil")
	}
}

// TestNewKeyDerivation tests key derivation creation
func TestNewKeyDerivation(t *testing.T) {
	kd := NewKeyDerivation()

	if kd == nil {
		t.Fatal("NewKeyDerivation returned nil")
	}
}
