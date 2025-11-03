package config

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"time"

	"golang.org/x/crypto/pbkdf2"
)

// Encryptor handles encryption and decryption of sensitive configuration values.
// Uses AES-256-GCM for authenticated encryption with keys stored in platform keychain.
//
// Implementation: internal/config/encryption.go + keychain.go
// See: FR-015 through FR-019
type Encryptor interface {
	// Encrypt encrypts a plaintext value using AES-256-GCM.
	// The encryption key is retrieved from the platform keychain (or env var fallback).
	//
	// Parameters:
	//   - plaintext: The value to encrypt (e.g., API key, token, password)
	//   - keyID: Identifier for the encryption key (e.g., "prod", "dev")
	//
	// Returns:
	//   - *EncryptedValue: Contains ciphertext, nonce, key ID, algorithm, timestamp
	//   - error: If encryption fails (key not found, keychain unavailable, etc.)
	//
	// Key lookup order:
	//   1. Platform keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service)
	//   2. Environment variable: LAZYNUGET_ENCRYPTION_KEY_<KEYID> (e.g., LAZYNUGET_ENCRYPTION_KEY_PROD)
	//   3. If neither available: return error
	//
	// See: FR-016, FR-017
	Encrypt(ctx context.Context, plaintext, keyID string) (*EncryptedValue, error)

	// Decrypt decrypts an EncryptedValue back to plaintext.
	//
	// Parameters:
	//   - encrypted: The encrypted value to decrypt
	//
	// Returns:
	//   - string: Decrypted plaintext
	//   - error: If decryption fails (key not found, wrong key, corrupted ciphertext, etc.)
	//
	// On failure: Caller should log warning and use default value (FR-012, edge cases)
	//
	// See: FR-016, FR-017
	Decrypt(ctx context.Context, encrypted *EncryptedValue) (string, error)

	// EncryptToString encrypts and returns a string suitable for embedding in config files.
	// Format: "!encrypted <base64-ciphertext-with-embedded-nonce>"
	//
	// This is a convenience method that combines Encrypt() + base64 encoding.
	// Used by the `lazynuget encrypt-value` CLI command (FR-019).
	EncryptToString(ctx context.Context, plaintext, keyID string) (string, error)

	// DecryptFromString parses an encrypted config value string and decrypts it.
	// Reverse of EncryptToString().
	//
	// Handles formats:
	//   - "!encrypted <base64>" (YAML custom tag format)
	//   - "AES256GCM:<keyID>:<base64>" (explicit format with key ID)
	DecryptFromString(ctx context.Context, encrypted string) (string, error)
}

// KeyDerivation handles deriving encryption keys from passwords.
// Used when users want to encrypt values using a memorable password instead of managing raw keys.
type KeyDerivation interface {
	// DeriveKey derives a 256-bit encryption key from a password using PBKDF2.
	//
	// Parameters:
	//   - password: User-provided password
	//   - salt: Cryptographic salt (typically 16 bytes, randomly generated once and stored)
	//   - iterations: Number of PBKDF2 iterations (recommended: 100,000+ for security vs performance)
	//
	// Returns:
	//   - []byte: 32-byte derived key suitable for AES-256
	//
	// Note: Salt must be stored alongside the encrypted value (or in keychain) for decryption.
	// Do NOT hardcode salt - must be random per application instance or per key ID.
	DeriveKey(password string, salt []byte, iterations int) []byte

	// GenerateSalt creates a cryptographically random salt for key derivation.
	// Should be called once per key ID and stored in keychain or config metadata.
	GenerateSalt() ([]byte, error)
}

// encryptor implements the Encryptor interface using AES-256-GCM.
// See: T118
type encryptor struct {
	keychain      KeychainManager
	keyDerivation KeyDerivation
}

// NewEncryptor creates a new Encryptor instance.
func NewEncryptor(keychain KeychainManager, kd KeyDerivation) Encryptor {
	return &encryptor{
		keychain:      keychain,
		keyDerivation: kd,
	}
}

// Encrypt encrypts a plaintext value using AES-256-GCM.
// See: T119, FR-016, FR-017
func (e *encryptor) Encrypt(ctx context.Context, plaintext, keyID string) (*EncryptedValue, error) {
	// Retrieve encryption key from keychain or env var
	key, err := e.keychain.Retrieve(ctx, keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve encryption key %q: %w", keyID, err)
	}

	// Validate key length (must be 32 bytes for AES-256)
	if len(key) != 32 {
		return nil, fmt.Errorf("invalid key length: got %d bytes, want 32 bytes for AES-256", len(key))
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt plaintext
	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)

	return &EncryptedValue{
		Ciphertext:  ciphertext,
		Nonce:       nonce,
		KeyID:       keyID,
		Algorithm:   "AES-256-GCM",
		EncryptedAt: time.Now(),
	}, nil
}

// Decrypt decrypts an EncryptedValue back to plaintext.
// See: T120, FR-016, FR-017
func (e *encryptor) Decrypt(ctx context.Context, encrypted *EncryptedValue) (string, error) {
	// Retrieve decryption key
	key, err := e.keychain.Retrieve(ctx, encrypted.KeyID)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve decryption key %q: %w", encrypted.KeyID, err)
	}

	// Validate key length
	if len(key) != 32 {
		return "", fmt.Errorf("invalid key length: got %d bytes, want 32 bytes for AES-256", len(key))
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt ciphertext
	plaintext, err := gcm.Open(nil, encrypted.Nonce, encrypted.Ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// EncryptToString encrypts and returns a string suitable for embedding in config files.
// See: T121, FR-019
func (e *encryptor) EncryptToString(ctx context.Context, plaintext, keyID string) (string, error) {
	encrypted, err := e.Encrypt(ctx, plaintext, keyID)
	if err != nil {
		return "", err
	}

	// Combine nonce + ciphertext and encode as base64
	combined := append(encrypted.Nonce, encrypted.Ciphertext...)
	encoded := base64.StdEncoding.EncodeToString(combined)

	// Return in YAML custom tag format
	return fmt.Sprintf("!encrypted %s", encoded), nil
}

// DecryptFromString parses an encrypted config value string and decrypts it.
// See: T122
func (e *encryptor) DecryptFromString(ctx context.Context, encrypted string) (string, error) {
	// Handle "!encrypted <base64>" format
	encrypted = strings.TrimSpace(encrypted)
	if after, ok := strings.CutPrefix(encrypted, "!encrypted "); ok {
		encrypted = after
		encrypted = strings.TrimSpace(encrypted)
	}

	// Handle "AES256GCM:<keyID>:<base64>" format
	var keyID string
	var encoded string
	if strings.HasPrefix(encrypted, "AES256GCM:") {
		parts := strings.SplitN(encrypted, ":", 3)
		if len(parts) != 3 {
			return "", fmt.Errorf("invalid encrypted format: expected 'AES256GCM:<keyID>:<base64>'")
		}
		keyID = parts[1]
		encoded = parts[2]
	} else {
		// Simple base64 format - assume default key ID "default"
		keyID = "default"
		encoded = encrypted
	}

	// Decode base64
	combined, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Extract nonce and ciphertext
	// GCM nonce is typically 12 bytes
	const nonceSize = 12
	if len(combined) < nonceSize {
		return "", fmt.Errorf("encrypted data too short: got %d bytes, need at least %d", len(combined), nonceSize)
	}

	nonce := combined[:nonceSize]
	ciphertext := combined[nonceSize:]

	encryptedValue := &EncryptedValue{
		Ciphertext: ciphertext,
		Nonce:      nonce,
		KeyID:      keyID,
		Algorithm:  "AES-256-GCM",
	}

	return e.Decrypt(ctx, encryptedValue)
}

// keyDerivation implements the KeyDerivation interface using PBKDF2.
// See: T129
type keyDerivation struct{}

// NewKeyDerivation creates a new KeyDerivation instance.
func NewKeyDerivation() KeyDerivation {
	return &keyDerivation{}
}

// DeriveKey derives a 256-bit encryption key from a password using PBKDF2.
// See: T129, FR-016
func (kd *keyDerivation) DeriveKey(password string, salt []byte, iterations int) []byte {
	// Use PBKDF2 with SHA-256 to derive a 32-byte (256-bit) key
	return pbkdf2.Key([]byte(password), salt, iterations, 32, sha256.New)
}

// GenerateSalt creates a cryptographically random salt for key derivation.
// See: T129
func (kd *keyDerivation) GenerateSalt() ([]byte, error) {
	// Generate 16 bytes of random data for salt
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return salt, nil
}
