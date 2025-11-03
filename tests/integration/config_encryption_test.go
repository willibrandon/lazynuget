package integration

import (
	"context"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/willibrandon/lazynuget/internal/config"
)

// T111: Test that encrypted value decrypts correctly
// See: FR-016, FR-017
func TestEncryptedValueDecryptsCorrectly(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a test encryption key (32 bytes for AES-256)
	testKeyID := "test"
	testKey := make([]byte, 32) // 32 bytes for AES-256
	for i := range testKey {
		testKey[i] = byte(i)
	}
	plaintext := "my-secret-api-key"

	// Store test key in environment variable (keychain not available in CI)
	// Encode as hex for proper storage
	envVar := "LAZYNUGET_ENCRYPTION_KEY_" + strings.ToUpper(testKeyID)
	os.Setenv(envVar, hex.EncodeToString(testKey))
	defer os.Unsetenv(envVar)

	// Create encryptor and keychain manager
	keychain := config.NewKeychainManager()
	kd := config.NewKeyDerivation()
	encryptor := config.NewEncryptor(keychain, kd)

	// Encrypt the plaintext
	encrypted, err := encryptor.Encrypt(ctx, plaintext, testKeyID)
	if err != nil {
		t.Fatalf("Encrypt() failed: %v", err)
	}

	// Decrypt the ciphertext
	decrypted, err := encryptor.Decrypt(ctx, encrypted)
	if err != nil {
		t.Fatalf("Decrypt() failed: %v", err)
	}

	// Verify decrypted plaintext matches original
	if decrypted != plaintext {
		t.Errorf("Decrypted value mismatch: got %q, want %q", decrypted, plaintext)
	}
}

// T112: Test that decryption failure falls back to default per FR-012
// See: FR-012
func TestDecryptionFailureFallbackToDefault(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	// Create config with encrypted value but NO decryption key available
	configContent := `version: "1.0"
logLevel: !encrypted "invalid-encrypted-data-should-fail"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	loader := config.NewLoader()
	opts := config.LoadOptions{
		ConfigFilePath: configPath,
		EnvVarPrefix:   "LAZYNUGET_",
	}

	// Load config - should fall back to default logLevel ("info") when decryption fails
	cfg, err := loader.Load(ctx, opts)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify default value is used (not the encrypted/failed value)
	if cfg.LogLevel != "info" {
		t.Errorf("Expected default logLevel 'info', got %q", cfg.LogLevel)
	}
}

// T113: Test that keychain unavailable falls back to env var
// See: FR-017
func TestKeychainUnavailableFallbackToEnvVar(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	testKeyID := "test"
	testKey := make([]byte, 32) // 32 bytes for AES-256
	for i := range testKey {
		testKey[i] = byte(i)
	}

	// Set environment variable (keychain not available in CI)
	// Encode as hex for proper storage
	envVar := "LAZYNUGET_ENCRYPTION_KEY_" + strings.ToUpper(testKeyID)
	os.Setenv(envVar, hex.EncodeToString(testKey))
	defer os.Unsetenv(envVar)

	keychain := config.NewKeychainManager()

	// Retrieve should fall back to env var
	retrievedKey, err := keychain.Retrieve(ctx, testKeyID)
	if err != nil {
		t.Fatalf("Retrieve() failed with env var fallback: %v", err)
	}

	// Verify retrieved key matches
	if string(retrievedKey) != string(testKey) {
		t.Errorf("Retrieved key mismatch: got %q, want %q", retrievedKey, testKey)
	}
}

// T114: Test that encrypted values never logged in plain text per FR-018
// See: FR-018
func TestEncryptedValuesNeverLoggedInPlaintext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a test encryption setup
	testKeyID := "test"
	testKey := make([]byte, 32) // 32 bytes for AES-256
	for i := range testKey {
		testKey[i] = byte(i)
	}
	plaintext := "my-secret-password"

	// Store test key in environment
	// Encode as hex for proper storage
	envVar := "LAZYNUGET_ENCRYPTION_KEY_" + strings.ToUpper(testKeyID)
	os.Setenv(envVar, hex.EncodeToString(testKey))
	defer os.Unsetenv(envVar)

	// Create encryptor
	keychain := config.NewKeychainManager()
	kd := config.NewKeyDerivation()
	encryptor := config.NewEncryptor(keychain, kd)

	// Encrypt the plaintext
	encrypted, err := encryptor.Encrypt(ctx, plaintext, testKeyID)
	if err != nil {
		t.Fatalf("Encrypt() failed: %v", err)
	}

	// Convert to string representation (as would appear in logs)
	encryptedStr, err := encryptor.EncryptToString(ctx, plaintext, testKeyID)
	if err != nil {
		t.Fatalf("EncryptToString() failed: %v", err)
	}

	// Verify that the encrypted string does NOT contain the plaintext
	if strings.Contains(encryptedStr, plaintext) {
		t.Errorf("Encrypted string contains plaintext! This violates FR-018")
	}

	// Verify that the EncryptedValue struct String() method does not leak plaintext
	encryptedString := encrypted.String()
	if strings.Contains(encryptedString, plaintext) {
		t.Errorf("EncryptedValue.String() contains plaintext! This violates FR-018")
	}
}
