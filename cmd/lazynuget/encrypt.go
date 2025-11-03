package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/willibrandon/lazynuget/internal/config"
)

// runEncryptValue implements the `lazynuget encrypt-value` subcommand.
// Encrypts a plaintext value using the platform keychain and outputs the encrypted string
// suitable for embedding in config files.
// See: T133, FR-019
func runEncryptValue(args []string) int {
	// Parse arguments
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: lazynuget encrypt-value <plaintext> [key-id]\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Encrypts a plaintext value for use in configuration files.\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  <plaintext>  The value to encrypt (e.g., API key, token, password)\n")
		fmt.Fprintf(os.Stderr, "  [key-id]     Optional key identifier (default: 'default')\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "The encryption key must be stored in the platform keychain or\n")
		fmt.Fprintf(os.Stderr, "provided via environment variable LAZYNUGET_ENCRYPTION_KEY_<KEYID>.\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Example:\n")
		fmt.Fprintf(os.Stderr, "  lazynuget encrypt-value \"my-secret-api-key\" prod\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Output can be used in config files:\n")
		fmt.Fprintf(os.Stderr, "  apiKey: !encrypted <base64-output>\n")
		return 1
	}

	plaintext := args[0]
	keyID := "default"
	if len(args) > 1 {
		keyID = args[1]
	}

	// Create encryption components
	keychain := config.NewKeychainManager()
	kd := config.NewKeyDerivation()
	encryptor := config.NewEncryptor(keychain, kd)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if keychain is available
	if !keychain.IsAvailable(ctx) {
		fmt.Fprintf(os.Stderr, "Warning: Platform keychain is not available.\n")
		fmt.Fprintf(os.Stderr, "You must provide the encryption key via environment variable:\n")
		fmt.Fprintf(os.Stderr, "  export LAZYNUGET_ENCRYPTION_KEY_%s=<32-byte-hex-key>\n", keyID)
		// Don't exit - user might have env var set
	}

	// Attempt to encrypt
	encryptedStr, err := encryptor.EncryptToString(ctx, plaintext, keyID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to encrypt value: %v\n", err)
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Troubleshooting:\n")
		fmt.Fprintf(os.Stderr, "  1. Ensure encryption key is stored in keychain:\n")
		fmt.Fprintf(os.Stderr, "     Use a password manager or system keychain utility\n")
		fmt.Fprintf(os.Stderr, "  2. Or provide key via environment variable:\n")
		fmt.Fprintf(os.Stderr, "     export LAZYNUGET_ENCRYPTION_KEY_%s=<32-byte-hex-key>\n", keyID)
		fmt.Fprintf(os.Stderr, "  3. Generate a new key:\n")
		fmt.Fprintf(os.Stderr, "     openssl rand -hex 32\n")
		return 1
	}

	// Output encrypted string (suitable for YAML config)
	fmt.Println(encryptedStr)

	// Print usage hint to stderr (so it doesn't interfere with piping)
	fmt.Fprintf(os.Stderr, "\nEncryption successful! Use in config file:\n")
	fmt.Fprintf(os.Stderr, "  someKey: %s\n", encryptedStr)

	return 0
}
