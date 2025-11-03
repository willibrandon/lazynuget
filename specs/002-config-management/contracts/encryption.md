# Encryption Contracts

Interfaces for encryption, keychain management, and key derivation.

## Encryptor Interface

```go
package contracts

import (
	"context"
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
	Encrypt(ctx context.Context, plaintext string, keyID string) (*EncryptedValue, error)

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
	EncryptToString(ctx context.Context, plaintext string, keyID string) (string, error)

	// DecryptFromString parses an encrypted config value string and decrypts it.
	// Reverse of EncryptToString().
	//
	// Handles formats:
	//   - "!encrypted <base64>" (YAML custom tag format)
	//   - "AES256GCM:<keyID>:<base64>" (explicit format with key ID)
	DecryptFromString(ctx context.Context, encrypted string) (string, error)
}
```

## KeychainManager Interface

```go
// KeychainManager manages encryption keys in the platform's secure storage.
// Abstracts differences between macOS Keychain, Windows Credential Manager, and Linux Secret Service.
//
// Implementation: internal/config/keychain.go (via github.com/zalando/go-keyring)
// See: FR-017
type KeychainManager interface {
	// Store saves an encryption key to the platform keychain.
	//
	// Parameters:
	//   - keyID: Identifier for the key (e.g., "prod", "dev")
	//   - key: The encryption key (32 bytes for AES-256)
	//
	// Returns:
	//   - error: If storage fails (keychain unavailable, permissions, etc.)
	//
	// Platform-specific behavior:
	//   - macOS: Stores in user keychain with service name "LazyNuGet" and account name = keyID
	//   - Windows: Stores in Credential Manager with target name "LazyNuGet:<keyID>"
	//   - Linux: Stores in Secret Service (GNOME Keyring/KDE Wallet) with label "LazyNuGet:<keyID>"
	Store(ctx context.Context, keyID string, key []byte) error

	// Retrieve fetches an encryption key from the platform keychain.
	//
	// Parameters:
	//   - keyID: Identifier for the key
	//
	// Returns:
	//   - []byte: The encryption key (32 bytes for AES-256)
	//   - error: If retrieval fails (key not found, keychain unavailable, etc.)
	//
	// Fallback behavior:
	//   If keychain retrieval fails, check environment variable:
	//   LAZYNUGET_ENCRYPTION_KEY_<KEYID> (e.g., LAZYNUGET_ENCRYPTION_KEY_PROD)
	//
	//   If env var found: Decode from hex/base64, return key
	//   If env var not found: Return error
	Retrieve(ctx context.Context, keyID string) ([]byte, error)

	// Delete removes an encryption key from the platform keychain.
	// Used for key rotation or cleanup.
	Delete(ctx context.Context, keyID string) error

	// List returns all key IDs stored in the keychain for this application.
	// Useful for debugging and key management.
	List(ctx context.Context) ([]string, error)

	// IsAvailable checks if the platform keychain is accessible.
	// Returns false for headless environments, CI systems, or when keychain is locked.
	//
	// If false, Encryptor will fall back to environment variable key storage.
	IsAvailable(ctx context.Context) bool
}
```

## KeyDerivation Interface

```go
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
```
