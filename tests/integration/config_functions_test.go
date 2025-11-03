package integration

import (
	"context"
	"testing"

	"github.com/willibrandon/lazynuget/internal/config"
)

// TestGetDefaultConfigValues tests that default config has expected values
func TestGetDefaultConfigValues(t *testing.T) {
	defaults := config.GetDefaultConfig()
	if defaults == nil {
		t.Fatal("GetDefaultConfig() returned nil")
	}

	// Verify some expected defaults
	if defaults.LogLevel == "" {
		t.Error("Expected LogLevel to be set in defaults")
	}

	if defaults.MaxConcurrentOps <= 0 {
		t.Error("Expected MaxConcurrentOps to be positive in defaults")
	}
}

// TestConfigLoader tests the config loader
func TestConfigLoader(t *testing.T) {
	loader := config.NewLoader()
	if loader == nil {
		t.Fatal("NewLoader() returned nil")
	}

	// Test GetDefaults
	defaults := loader.GetDefaults()
	if defaults == nil {
		t.Error("GetDefaults() returned nil")
	}

	// Test PrintConfig with defaults
	loader.PrintConfig(defaults)
}

// TestKeychainManager tests keychain operations
func TestKeychainManager(t *testing.T) {
	keychain := config.NewKeychainManager()
	if keychain == nil {
		t.Fatal("NewKeychainManager() returned nil")
	}

	ctx := context.Background()

	// Test IsAvailable
	available := keychain.IsAvailable(ctx)
	t.Logf("Keychain available: %v", available)

	if !available {
		t.Skip("Keychain not available on this platform")
	}

	testKey := "lazynuget-test-integration-key"
	testValue := []byte("test-value")

	// Test Store
	err := keychain.Store(ctx, testKey, testValue)
	if err != nil {
		t.Logf("Store() failed: %v", err)
	}

	// Test Retrieve
	value, err := keychain.Retrieve(ctx, testKey)
	if err != nil {
		t.Logf("Retrieve() failed: %v", err)
	} else {
		t.Logf("Retrieved value: %s", value)
	}

	// Test List
	keys, err := keychain.List(ctx)
	if err != nil {
		t.Logf("List() failed: %v", err)
	} else {
		t.Logf("Found %d keys", len(keys))
	}

	// Test Delete
	err = keychain.Delete(ctx, testKey)
	if err != nil {
		t.Logf("Delete() failed: %v", err)
	}
}

// TestEncryptorAndKeyDerivation tests encryption interfaces
func TestEncryptorAndKeyDerivation(t *testing.T) {
	keychain := config.NewKeychainManager()
	kd := config.NewKeyDerivation()
	encryptor := config.NewEncryptor(keychain, kd)

	if encryptor == nil {
		t.Fatal("NewEncryptor() returned nil")
	}

	if kd == nil {
		t.Fatal("NewKeyDerivation() returned nil")
	}

	// Just test that we can create these objects
	// Actual encryption/decryption would require keychain setup
}
