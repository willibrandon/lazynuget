package config

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"

	"gopkg.in/yaml.v3"
)

// parseYAML parses YAML config file content into a Config struct.
// Unknown fields are silently ignored per FR-011 and FR-013.
// See: T044, FR-003, FR-010, FR-011
func parseYAML(data []byte) (*Config, error) {
	var cfg Config

	// Use decoder WITHOUT strict mode - unknown fields should be ignored
	// Per FR-011: Unknown config keys are ignored with warning
	// Per FR-013: Unknown keys are non-blocking semantic errors
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(false) // Allow unknown fields

	if err := decoder.Decode(&cfg); err != nil {
		// Provide helpful error message with line/column info if available
		return nil, fmt.Errorf("YAML parsing error: %w\n\n"+
			"Please check the file for syntax errors:\n"+
			"  • Ensure proper indentation (use spaces, not tabs)\n"+
			"  • Check for missing colons or quotes\n"+
			"  • Validate YAML syntax at https://www.yamllint.com/", err)
	}

	return &cfg, nil
}

// EncryptedString is a custom type that can be unmarshaled from YAML's !encrypted tag.
// It stores the encrypted value as base64 and can be decrypted later.
// See: T130, FR-015, FR-016
type EncryptedString struct {
	Value         string
	KeyID         string
	Base64Data    string
	DecryptedText string
	IsEncrypted   bool
}

// UnmarshalYAML implements custom YAML unmarshaling for encrypted values.
// Handles the !encrypted tag format: !encrypted <base64>
func (es *EncryptedString) UnmarshalYAML(node *yaml.Node) error {
	// Check if this is an encrypted value (custom tag)
	if node.Tag == "!encrypted" {
		es.IsEncrypted = true
		es.Base64Data = node.Value
		es.KeyID = "default" // Default key ID unless specified
		es.Value = "!encrypted " + node.Value
		return nil
	}

	// Regular string value
	es.IsEncrypted = false
	es.Value = node.Value
	return nil
}

// DecryptValue decrypts the encrypted string using the provided encryptor.
// Returns the decrypted plaintext or an error.
func (es *EncryptedString) DecryptValue(encryptor Encryptor) (string, error) {
	if !es.IsEncrypted {
		return es.Value, nil
	}

	// Return cached value if already decrypted
	if es.DecryptedText != "" {
		return es.DecryptedText, nil
	}

	// Decode base64
	combined, err := base64.StdEncoding.DecodeString(es.Base64Data)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted value: %w", err)
	}

	// Extract nonce and ciphertext
	const nonceSize = 12
	if len(combined) < nonceSize {
		return "", fmt.Errorf("encrypted data too short")
	}

	nonce := combined[:nonceSize]
	ciphertext := combined[nonceSize:]

	encrypted := &EncryptedValue{
		Ciphertext: ciphertext,
		Nonce:      nonce,
		KeyID:      es.KeyID,
		Algorithm:  "AES-256-GCM",
	}

	// Decrypt
	plaintext, err := encryptor.Decrypt(context.TODO(), encrypted)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt value: %w", err)
	}

	// Cache decrypted value
	es.DecryptedText = plaintext
	return plaintext, nil
}

// parseYAMLWithEncryption parses YAML config and handles encrypted values.
// This is an internal helper that will be used by Load() to decrypt values.
// See: T130, T131
func parseYAMLWithEncryption(data []byte) (*Config, map[string]*EncryptedValue, error) {
	// First parse as raw YAML to detect encrypted tags
	var rawNode yaml.Node
	if err := yaml.Unmarshal(data, &rawNode); err != nil {
		return nil, nil, fmt.Errorf("YAML parsing error: %w", err)
	}

	// Parse normally
	cfg, err := parseYAML(data)
	if err != nil {
		return nil, nil, err
	}

	// Scan for encrypted values in the YAML tree
	encryptedFields := make(map[string]*EncryptedValue)
	scanForEncryptedValues(&rawNode, "", encryptedFields)

	return cfg, encryptedFields, nil
}

// scanForEncryptedValues recursively scans a YAML node tree for !encrypted tags.
func scanForEncryptedValues(node *yaml.Node, path string, encrypted map[string]*EncryptedValue) {
	if node == nil {
		return
	}

	// Check if this node is an encrypted value
	if node.Tag == "!encrypted" && node.Kind == yaml.ScalarNode {
		// Decode the encrypted value
		combined, err := base64.StdEncoding.DecodeString(node.Value)
		if err != nil {
			return
		}

		// Extract nonce and ciphertext
		const nonceSize = 12
		if len(combined) < nonceSize {
			return
		}

		nonce := combined[:nonceSize]
		ciphertext := combined[nonceSize:]

		encrypted[path] = &EncryptedValue{
			Ciphertext: ciphertext,
			Nonce:      nonce,
			KeyID:      "default",
			Algorithm:  "AES-256-GCM",
		}
		return
	}

	// Recursively scan child nodes
	switch node.Kind {
	case yaml.DocumentNode:
		for _, child := range node.Content {
			scanForEncryptedValues(child, path, encrypted)
		}
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]
			fieldPath := path
			if fieldPath != "" {
				fieldPath += "."
			}
			fieldPath += keyNode.Value
			scanForEncryptedValues(valueNode, fieldPath, encrypted)
		}
	case yaml.SequenceNode:
		for i, child := range node.Content {
			fieldPath := fmt.Sprintf("%s[%d]", path, i)
			scanForEncryptedValues(child, fieldPath, encrypted)
		}
	}
}
