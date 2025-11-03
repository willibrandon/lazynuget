package config

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/zalando/go-keyring"
)

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

const (
	keychainService = "LazyNuGet"
)

// keychainManager implements KeychainManager using github.com/zalando/go-keyring.
// See: T123
type keychainManager struct{}

// NewKeychainManager creates a new KeychainManager instance.
func NewKeychainManager() KeychainManager {
	return &keychainManager{}
}

// Store saves an encryption key to the platform keychain.
// See: T124, FR-017
func (km *keychainManager) Store(_ context.Context, keyID string, key []byte) error {
	// Encode key as hex for storage
	keyHex := hex.EncodeToString(key)

	// Store in platform keychain
	if err := keyring.Set(keychainService, keyID, keyHex); err != nil {
		return fmt.Errorf("failed to store key in keychain: %w", err)
	}

	return nil
}

// Retrieve fetches an encryption key from the platform keychain.
// Falls back to environment variable if keychain unavailable.
// See: T125, FR-017
func (km *keychainManager) Retrieve(_ context.Context, keyID string) ([]byte, error) {
	// Try to retrieve from keychain first
	keyHex, err := keyring.Get(keychainService, keyID)
	if err == nil {
		// Decode hex to bytes
		key, err := hex.DecodeString(keyHex)
		if err != nil {
			return nil, fmt.Errorf("failed to decode key from keychain: %w", err)
		}
		return key, nil
	}

	// Keychain retrieval failed, try environment variable fallback
	envVar := "LAZYNUGET_ENCRYPTION_KEY_" + strings.ToUpper(keyID)
	envValue := os.Getenv(envVar)
	if envValue == "" {
		return nil, fmt.Errorf("key %q not found in keychain or environment variable %s", keyID, envVar)
	}

	// Try to decode from hex first
	if key, err := hex.DecodeString(envValue); err == nil {
		return key, nil
	}

	// Try to decode from base64
	if key, err := base64.StdEncoding.DecodeString(envValue); err == nil {
		return key, nil
	}

	// If neither encoding works, treat as raw bytes
	return []byte(envValue), nil
}

// Delete removes an encryption key from the platform keychain.
// See: T126
func (km *keychainManager) Delete(_ context.Context, keyID string) error {
	if err := keyring.Delete(keychainService, keyID); err != nil {
		return fmt.Errorf("failed to delete key from keychain: %w", err)
	}
	return nil
}

// List returns all key IDs stored in the keychain for this application.
// See: T127
func (km *keychainManager) List(_ context.Context) ([]string, error) {
	// Note: go-keyring doesn't provide a List function, so we can't enumerate keys
	// This is a limitation of the underlying platform keychains
	// Return empty list with a note in the error
	return []string{}, fmt.Errorf("listing keys is not supported by the underlying keychain implementation")
}

// IsAvailable checks if the platform keychain is accessible.
// See: T128
func (km *keychainManager) IsAvailable(_ context.Context) bool {
	// Try to perform a test operation (get a non-existent key)
	// If we get an error other than "not found", keychain is unavailable
	_, err := keyring.Get(keychainService, "test-availability-check")
	if err == nil {
		// Key exists (unlikely but possible)
		return true
	}

	// Check if error is "not found" (keychain is available)
	// vs other errors like "keychain locked" or "service unavailable"
	errStr := err.Error()
	if strings.Contains(errStr, "not found") || strings.Contains(errStr, "cannot find") {
		return true
	}

	// Keychain is unavailable
	return false
}
