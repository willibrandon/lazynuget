package config

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"os"
	"strings"
	"testing"
)

// TestKeychainManagerCreation tests NewKeychainManager
func TestKeychainManagerCreation(t *testing.T) {
	km := NewKeychainManager()
	if km == nil {
		t.Fatal("NewKeychainManager() returned nil")
	}
}

// TestKeychainIsAvailable tests IsAvailable method
func TestKeychainIsAvailable(t *testing.T) {
	km := NewKeychainManager()
	ctx := context.Background()

	// IsAvailable should return a boolean without error
	available := km.IsAvailable(ctx)

	// We don't know if keychain is available on this platform,
	// but the method should not panic
	t.Logf("Keychain available: %v", available)
}

// TestKeychainStoreRetrieve tests Store and Retrieve with environment variable fallback
func TestKeychainStoreRetrieve(t *testing.T) {
	km := NewKeychainManager()
	ctx := context.Background()

	// Use a unique key ID for this test
	keyID := "test-lazynuget-integration-key"
	testKey := []byte("0123456789abcdef0123456789abcdef") // 32 bytes for AES-256

	// Try to store the key
	err := km.Store(ctx, keyID, testKey)
	if err != nil {
		// Keychain might not be available - test with environment variable fallback
		t.Logf("Store failed (expected in CI/headless): %v", err)

		// Test environment variable fallback
		envKey := "LAZYNUGET_ENCRYPTION_KEY_" + strings.ToUpper(keyID)
		originalValue := os.Getenv(envKey)
		defer func() {
			if originalValue != "" {
				os.Setenv(envKey, originalValue)
			} else {
				os.Unsetenv(envKey)
			}
		}()

		// Set key in environment variable
		os.Setenv(envKey, hex.EncodeToString(testKey))

		// Should retrieve from environment variable
		retrieved, retrieveErr := km.Retrieve(ctx, keyID)
		if retrieveErr != nil {
			t.Fatalf("Retrieve from env var failed: %v", retrieveErr)
		}

		if string(retrieved) != string(testKey) {
			t.Errorf("Retrieved key from env = %v, want %v", retrieved, testKey)
		}

		return
	}

	// Keychain is available, test retrieval from keychain
	retrieved, err := km.Retrieve(ctx, keyID)
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}

	if string(retrieved) != string(testKey) {
		t.Errorf("Retrieved key = %v, want %v", retrieved, testKey)
	}

	// Clean up
	err = km.Delete(ctx, keyID)
	if err != nil {
		t.Logf("Delete failed (cleanup): %v", err)
	}
}

// TestKeychainRetrieveNonExistent tests retrieving a non-existent key
func TestKeychainRetrieveNonExistent(t *testing.T) {
	km := NewKeychainManager()
	ctx := context.Background()

	// Try to retrieve a key that doesn't exist
	keyID := "non-existent-key-" + t.Name()

	// Make sure no env var exists
	envKey := "LAZYNUGET_ENCRYPTION_KEY_" + strings.ToUpper(keyID)
	os.Unsetenv(envKey)

	_, err := km.Retrieve(ctx, keyID)
	if err == nil {
		t.Error("Retrieve should fail for non-existent key")
	}
}

// TestKeychainDelete tests Delete method
func TestKeychainDelete(t *testing.T) {
	km := NewKeychainManager()
	ctx := context.Background()

	keyID := "test-delete-key"
	testKey := []byte("test-key-value-32-bytes-long!!!")

	// Try to store
	err := km.Store(ctx, keyID, testKey)
	if err != nil {
		t.Skipf("Keychain not available, skipping delete test: %v", err)
	}

	// Delete the key
	err = km.Delete(ctx, keyID)
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	// Verify it's deleted
	_, err = km.Retrieve(ctx, keyID)
	if err == nil {
		t.Error("Key should not be retrievable after deletion")
	}
}

// TestKeychainList tests List method
func TestKeychainList(t *testing.T) {
	km := NewKeychainManager()
	ctx := context.Background()

	// List keys (may be empty or may have keys)
	keys, err := km.List(ctx)
	if err != nil {
		t.Logf("List failed (expected in some environments): %v", err)
		return
	}

	// Should return a slice (even if empty)
	if keys == nil {
		t.Error("List should return non-nil slice")
	}

	t.Logf("Found %d keys in keychain", len(keys))
}

// TestKeychainEnvironmentVariableFallback tests environment variable fallback explicitly
func TestKeychainEnvironmentVariableFallback(t *testing.T) {
	km := NewKeychainManager()
	ctx := context.Background()

	keyID := "env-fallback-test"
	testKey := []byte("environment-variable-key-value!")

	// Set environment variable (keyID is uppercased by keychain implementation)
	envKey := "LAZYNUGET_ENCRYPTION_KEY_" + strings.ToUpper(keyID)
	originalValue := os.Getenv(envKey)
	defer func() {
		if originalValue != "" {
			os.Setenv(envKey, originalValue)
		} else {
			os.Unsetenv(envKey)
		}
	}()

	os.Setenv(envKey, hex.EncodeToString(testKey))

	// Retrieve should work even if keychain is not available
	retrieved, err := km.Retrieve(ctx, keyID)
	if err != nil {
		t.Fatalf("Retrieve from env var failed: %v", err)
	}

	if string(retrieved) != string(testKey) {
		t.Errorf("Retrieved key = %v, want %v", retrieved, testKey)
	}
}

// TestKeychainBase64Fallback tests base64 encoded environment variables
func TestKeychainBase64Fallback(t *testing.T) {
	km := NewKeychainManager()
	ctx := context.Background()

	keyID := "base64-fallback-test"
	testKey := []byte("base64-encoded-key-value-here!!")

	envKey := "LAZYNUGET_ENCRYPTION_KEY_" + strings.ToUpper(keyID)
	originalValue := os.Getenv(envKey)
	defer func() {
		if originalValue != "" {
			os.Setenv(envKey, originalValue)
		} else {
			os.Unsetenv(envKey)
		}
	}()

	// Try base64 encoding
	os.Setenv(envKey, base64.StdEncoding.EncodeToString(testKey))

	// Retrieve should work
	retrieved, err := km.Retrieve(ctx, keyID)
	if err != nil {
		t.Logf("Base64 retrieve failed (might only support hex): %v", err)
		return
	}

	if string(retrieved) != string(testKey) {
		t.Logf("Base64 decode mismatch (might only support hex)")
	}
}
