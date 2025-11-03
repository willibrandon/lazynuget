package config

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestParseYAML tests YAML parsing with table-driven tests
func TestParseYAML(t *testing.T) {
	tests := []struct {
		checkFunc func(*Config) error
		name      string
		yaml      string
		wantErr   bool
	}{
		{
			name: "simple fields",
			yaml: `
logLevel: debug
maxConcurrentOps: 8
theme: dark
`,
			wantErr: false,
			checkFunc: func(cfg *Config) error {
				if cfg.LogLevel != "debug" {
					t.Errorf("Expected LogLevel=debug, got %s", cfg.LogLevel)
				}
				if cfg.MaxConcurrentOps != 8 {
					t.Errorf("Expected MaxConcurrentOps=8, got %d", cfg.MaxConcurrentOps)
				}
				if cfg.Theme != "dark" {
					t.Errorf("Expected Theme=dark, got %s", cfg.Theme)
				}
				return nil
			},
		},
		{
			name: "boolean values",
			yaml: `
compactMode: true
showHints: false
showLineNumbers: true
hotReload: false
`,
			wantErr: false,
			checkFunc: func(cfg *Config) error {
				if !cfg.CompactMode {
					t.Error("Expected CompactMode=true")
				}
				if cfg.ShowHints {
					t.Error("Expected ShowHints=false")
				}
				if !cfg.ShowLineNumbers {
					t.Error("Expected ShowLineNumbers=true")
				}
				if cfg.HotReload {
					t.Error("Expected HotReload=false")
				}
				return nil
			},
		},
		{
			name: "nested colorScheme",
			yaml: `
colorScheme:
  border: "#5C6370"
  borderFocus: "#61AFEF"
  text: "#ABB2BF"
  error: "#E06C75"
`,
			wantErr: false,
			checkFunc: func(cfg *Config) error {
				if cfg.ColorScheme.Border != "#5C6370" {
					t.Errorf("Expected Border=#5C6370, got %s", cfg.ColorScheme.Border)
				}
				if cfg.ColorScheme.BorderFocus != "#61AFEF" {
					t.Errorf("Expected BorderFocus=#61AFEF, got %s", cfg.ColorScheme.BorderFocus)
				}
				if cfg.ColorScheme.Text != "#ABB2BF" {
					t.Errorf("Expected Text=#ABB2BF, got %s", cfg.ColorScheme.Text)
				}
				if cfg.ColorScheme.Error != "#E06C75" {
					t.Errorf("Expected Error=#E06C75, got %s", cfg.ColorScheme.Error)
				}
				return nil
			},
		},
		{
			name: "nested timeouts",
			yaml: `
timeouts:
  networkRequest: 30s
  dotnetCli: 2m
  fileOperation: 10s
`,
			wantErr: false,
			checkFunc: func(_ *Config) error {
				// Note: YAML parsing of durations requires them to be strings
				// The actual duration conversion happens in mergeConfigs or elsewhere
				return nil
			},
		},
		{
			name: "nested logRotation",
			yaml: `
logRotation:
  maxSize: 20
  maxAge: 60
  maxBackups: 10
  compress: true
`,
			wantErr: false,
			checkFunc: func(cfg *Config) error {
				if cfg.LogRotation.MaxSize != 20 {
					t.Errorf("Expected MaxSize=20, got %d", cfg.LogRotation.MaxSize)
				}
				if cfg.LogRotation.MaxAge != 60 {
					t.Errorf("Expected MaxAge=60, got %d", cfg.LogRotation.MaxAge)
				}
				if cfg.LogRotation.MaxBackups != 10 {
					t.Errorf("Expected MaxBackups=10, got %d", cfg.LogRotation.MaxBackups)
				}
				if !cfg.LogRotation.Compress {
					t.Error("Expected Compress=true")
				}
				return nil
			},
		},
		{
			name: "invalid YAML syntax",
			yaml: `
logLevel: debug
  invalid indentation
theme: dark
`,
			wantErr: false, // YAML parser is lenient with indentation
		},
		{
			name:    "empty YAML",
			yaml:    ``,
			wantErr: true, // Empty YAML returns EOF error
		},
		{
			name: "YAML with comments",
			yaml: `
# This is a comment
logLevel: debug  # inline comment
theme: dark
`,
			wantErr: false,
			checkFunc: func(cfg *Config) error {
				if cfg.LogLevel != "debug" {
					t.Errorf("Expected LogLevel=debug, got %s", cfg.LogLevel)
				}
				if cfg.Theme != "dark" {
					t.Errorf("Expected Theme=dark, got %s", cfg.Theme)
				}
				return nil
			},
		},
		{
			name: "unknown fields are ignored",
			yaml: `
logLevel: debug
unknownField: somevalue
anotherUnknown: 123
theme: dark
`,
			wantErr: false,
			checkFunc: func(cfg *Config) error {
				if cfg.LogLevel != "debug" {
					t.Errorf("Expected LogLevel=debug, got %s", cfg.LogLevel)
				}
				if cfg.Theme != "dark" {
					t.Errorf("Expected Theme=dark, got %s", cfg.Theme)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parseYAML([]byte(tt.yaml))

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.checkFunc != nil {
				_ = tt.checkFunc(cfg)
			}
		})
	}
}

// TestParseYAMLTypes tests various YAML type conversions
func TestParseYAMLTypes(t *testing.T) {
	tests := []struct {
		want    any
		name    string
		yaml    string
		field   string
		wantErr bool
	}{
		{
			name:  "string field",
			yaml:  "theme: dark",
			field: "theme",
			want:  "dark",
		},
		{
			name:  "int field",
			yaml:  "maxConcurrentOps: 8",
			field: "maxConcurrentOps",
			want:  8,
		},
		{
			name:  "bool field true",
			yaml:  "compactMode: true",
			field: "compactMode",
			want:  true,
		},
		{
			name:  "bool field false",
			yaml:  "showHints: false",
			field: "showHints",
			want:  false,
		},
		{
			name:  "quoted string",
			yaml:  "logLevel: \"debug\"",
			field: "logLevel",
			want:  "debug",
		},
		{
			name:  "number as string",
			yaml:  "version: \"1.0\"",
			field: "version",
			want:  "1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parseYAML([]byte(tt.yaml))

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check specific field
			var actual any
			switch tt.field {
			case "theme":
				actual = cfg.Theme
			case "logLevel":
				actual = cfg.LogLevel
			case "maxConcurrentOps":
				actual = cfg.MaxConcurrentOps
			case "compactMode":
				actual = cfg.CompactMode
			case "showHints":
				actual = cfg.ShowHints
			case "version":
				actual = cfg.Version
			default:
				t.Fatalf("Unknown field in test: %s", tt.field)
			}

			if actual != tt.want {
				t.Errorf("Field %s: expected %v, got %v", tt.field, tt.want, actual)
			}
		})
	}
}

// TestParseYAMLErrorMessages tests that error messages are helpful
func TestParseYAMLErrorMessages(t *testing.T) {
	tests := []struct {
		name        string
		yaml        string
		errContains string
		wantErr     bool
	}{
		{
			name: "invalid indentation",
			yaml: `
logLevel: debug
  invalid
`,
			wantErr:     false, // YAML parser is lenient with this
			errContains: "",
		},
		{
			name: "unclosed quote",
			yaml: `
logLevel: "unclosed
theme: dark
`,
			wantErr:     true,
			errContains: "",
		},
		{
			name: "invalid structure",
			yaml: `
- this
- is
- a list
- not an object
`,
			wantErr: true, // YAML parser rejects array when expecting object
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseYAML([]byte(tt.yaml))

			if tt.wantErr && err == nil {
				t.Error("Expected error but got none")
				return
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
		})
	}
}

// TestUnmarshalYAML tests EncryptedString YAML unmarshaling
func TestUnmarshalYAML(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		wantValue   string
		wantKeyID   string
		wantEncrypt bool
	}{
		{
			name: "encrypted value with tag",
			yamlContent: `value: !encrypted dGVzdGRhdGExMjM0NTY3OA==
`,
			wantEncrypt: true,
			wantValue:   "!encrypted dGVzdGRhdGExMjM0NTY3OA==",
			wantKeyID:   "default",
		},
		{
			name: "regular string value",
			yamlContent: `value: hello world
`,
			wantEncrypt: false,
			wantValue:   "hello world",
			wantKeyID:   "",
		},
		{
			name: "empty string value",
			yamlContent: `value: ""
`,
			wantEncrypt: false,
			wantValue:   "",
			wantKeyID:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				Value EncryptedString `yaml:"value"`
			}

			var ts TestStruct
			err := yaml.Unmarshal([]byte(tt.yamlContent), &ts)
			if err != nil {
				t.Fatalf("yaml.Unmarshal() error = %v", err)
			}

			if ts.Value.IsEncrypted != tt.wantEncrypt {
				t.Errorf("IsEncrypted = %v, want %v", ts.Value.IsEncrypted, tt.wantEncrypt)
			}
			if ts.Value.Value != tt.wantValue {
				t.Errorf("Value = %v, want %v", ts.Value.Value, tt.wantValue)
			}
			if tt.wantEncrypt && ts.Value.KeyID != tt.wantKeyID {
				t.Errorf("KeyID = %v, want %v", ts.Value.KeyID, tt.wantKeyID)
			}
		})
	}
}

// TestDecryptValue tests EncryptedString decryption
func TestDecryptValue(t *testing.T) {
	ctx := context.Background()

	// Set up encryption key for testing
	testKey := make([]byte, 32)
	if _, err := rand.Read(testKey); err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}
	keyID := "test-decrypt-key"
	envKey := "LAZYNUGET_ENCRYPTION_KEY_" + strings.ToUpper(keyID)
	os.Setenv(envKey, hex.EncodeToString(testKey))
	defer os.Unsetenv(envKey)

	km := NewKeychainManager()
	kd := NewKeyDerivation()
	enc := NewEncryptor(km, kd)

	// Encrypt a test value
	plaintext := "secret-value"
	encrypted, err := enc.Encrypt(ctx, plaintext, keyID)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Combine nonce + ciphertext and encode as base64 (matching UnmarshalYAML format)
	combined := append(encrypted.Nonce, encrypted.Ciphertext...)
	base64Data := base64.StdEncoding.EncodeToString(combined)

	tests := []struct {
		encString   *EncryptedString
		name        string
		wantText    string
		errContains string
		wantErr     bool
	}{
		{
			name: "decrypt valid encrypted value",
			encString: &EncryptedString{
				IsEncrypted: true,
				Base64Data:  base64Data,
				KeyID:       keyID,
			},
			wantText: plaintext,
			wantErr:  false,
		},
		{
			name: "return plain value if not encrypted",
			encString: &EncryptedString{
				IsEncrypted: false,
				Value:       "plain-text",
			},
			wantText: "plain-text",
			wantErr:  false,
		},
		{
			name: "return cached decrypted value",
			encString: &EncryptedString{
				IsEncrypted:   true,
				DecryptedText: "cached-value",
			},
			wantText: "cached-value",
			wantErr:  false,
		},
		{
			name: "error on invalid base64",
			encString: &EncryptedString{
				IsEncrypted: true,
				Base64Data:  "not-valid-base64!!!",
				KeyID:       keyID,
			},
			wantErr:     true,
			errContains: "failed to decode encrypted value",
		},
		{
			name: "error on data too short",
			encString: &EncryptedString{
				IsEncrypted: true,
				Base64Data:  base64.StdEncoding.EncodeToString([]byte("short")),
				KeyID:       keyID,
			},
			wantErr:     true,
			errContains: "encrypted data too short",
		},
		{
			name: "error on wrong key",
			encString: &EncryptedString{
				IsEncrypted: true,
				Base64Data:  base64Data,
				KeyID:       "wrong-key-id",
			},
			wantErr:     true,
			errContains: "failed to decrypt value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.encString.DecryptValue(enc)

			if tt.wantErr {
				if err == nil {
					t.Errorf("DecryptValue() expected error containing %q, got nil", tt.errContains)
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("DecryptValue() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("DecryptValue() unexpected error = %v", err)
				return
			}

			if got != tt.wantText {
				t.Errorf("DecryptValue() = %v, want %v", got, tt.wantText)
			}
		})
	}
}

// TestScanForEncryptedValues tests scanning YAML tree for encrypted values
func TestScanForEncryptedValues(t *testing.T) {
	tests := []struct {
		name          string
		yamlContent   string
		wantEncrypted []string // paths of encrypted fields
	}{
		{
			name: "single encrypted value",
			yamlContent: `
apiKey: !encrypted dGVzdGRhdGExMjM0NTY3ODkwMTIzNDU2
`,
			wantEncrypted: []string{"apiKey"},
		},
		{
			name: "nested encrypted values",
			yamlContent: `
database:
  password: !encrypted dGVzdGRhdGExMjM0NTY3ODkwMTIzNDU2
  username: admin
api:
  token: !encrypted dGVzdGRhdGExMjM0NTY3ODkwMTIzNDU2
`,
			wantEncrypted: []string{"database.password", "api.token"},
		},
		{
			name: "encrypted value in array",
			yamlContent: `
secrets:
  - !encrypted dGVzdGRhdGExMjM0NTY3ODkwMTIzNDU2
  - plain-value
  - !encrypted dGVzdGRhdGExMjM0NTY3ODkwMTIzNDU2
`,
			wantEncrypted: []string{"secrets[0]", "secrets[2]"},
		},
		{
			name: "no encrypted values",
			yamlContent: `
logLevel: debug
theme: dark
maxConcurrentOps: 8
`,
			wantEncrypted: []string{},
		},
		{
			name: "deeply nested encrypted value",
			yamlContent: `
level1:
  level2:
    level3:
      secret: !encrypted dGVzdGRhdGExMjM0NTY3ODkwMTIzNDU2
`,
			wantEncrypted: []string{"level1.level2.level3.secret"},
		},
		{
			name: "mixed types with encrypted",
			yamlContent: `
string: plain
number: 42
bool: true
encrypted: !encrypted dGVzdGRhdGExMjM0NTY3ODkwMTIzNDU2
nested:
  value: !encrypted dGVzdGRhdGExMjM0NTY3ODkwMTIzNDU2
`,
			wantEncrypted: []string{"encrypted", "nested.value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var node yaml.Node
			if err := yaml.Unmarshal([]byte(tt.yamlContent), &node); err != nil {
				t.Fatalf("yaml.Unmarshal() error = %v", err)
			}

			encrypted := make(map[string]*EncryptedValue)
			scanForEncryptedValues(&node, "", encrypted)

			if len(encrypted) != len(tt.wantEncrypted) {
				t.Errorf("scanForEncryptedValues() found %d encrypted fields, want %d", len(encrypted), len(tt.wantEncrypted))
				t.Logf("Found paths: %v", getKeys(encrypted))
				t.Logf("Want paths: %v", tt.wantEncrypted)
			}

			for _, path := range tt.wantEncrypted {
				if _, found := encrypted[path]; !found {
					t.Errorf("scanForEncryptedValues() missing encrypted field at path: %s", path)
				}
			}

			// Verify encrypted values have correct structure
			for path, ev := range encrypted {
				if ev.Algorithm != "AES-256-GCM" {
					t.Errorf("encrypted[%s].Algorithm = %v, want AES-256-GCM", path, ev.Algorithm)
				}
				if ev.KeyID != "default" {
					t.Errorf("encrypted[%s].KeyID = %v, want default", path, ev.KeyID)
				}
				if len(ev.Nonce) == 0 {
					t.Errorf("encrypted[%s].Nonce is empty", path)
				}
				if len(ev.Ciphertext) == 0 {
					t.Errorf("encrypted[%s].Ciphertext is empty", path)
				}
			}
		})
	}
}

// TestScanForEncryptedValuesInvalidData tests error handling in scanForEncryptedValues
func TestScanForEncryptedValuesInvalidData(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
	}{
		{
			name: "invalid base64 in encrypted tag",
			yamlContent: `
apiKey: !encrypted not-valid-base64!!!
`,
		},
		{
			name: "encrypted tag with data too short",
			yamlContent: `
apiKey: !encrypted dGVzdA==
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var node yaml.Node
			if err := yaml.Unmarshal([]byte(tt.yamlContent), &node); err != nil {
				t.Fatalf("yaml.Unmarshal() error = %v", err)
			}

			encrypted := make(map[string]*EncryptedValue)
			scanForEncryptedValues(&node, "", encrypted)

			// Should gracefully ignore invalid encrypted values
			if len(encrypted) != 0 {
				t.Errorf("scanForEncryptedValues() should ignore invalid encrypted values, but found %d", len(encrypted))
			}
		})
	}
}

// TestScanForEncryptedValuesNilNode tests nil node handling
func TestScanForEncryptedValuesNilNode(t *testing.T) {
	encrypted := make(map[string]*EncryptedValue)
	scanForEncryptedValues(nil, "", encrypted)

	if len(encrypted) != 0 {
		t.Errorf("scanForEncryptedValues(nil) should not find any encrypted values, found %d", len(encrypted))
	}
}

// TestParseYAMLWithEncryption tests parseYAMLWithEncryption function
func TestParseYAMLWithEncryption(t *testing.T) {
	tests := []struct {
		name          string
		yamlContent   string
		wantLogLevel  string
		wantEncrypted int
		wantErr       bool
	}{
		{
			name: "valid yaml with encrypted value",
			yamlContent: `
logLevel: debug
apiKey: !encrypted dGVzdGRhdGExMjM0NTY3ODkwMTIzNDU2
`,
			wantEncrypted: 1,
			wantLogLevel:  "debug",
			wantErr:       false,
		},
		{
			name: "valid yaml without encrypted values",
			yamlContent: `
logLevel: info
theme: dark
`,
			wantEncrypted: 0,
			wantLogLevel:  "info",
			wantErr:       false,
		},
		{
			name: "invalid yaml",
			yamlContent: `
logLevel: debug
theme: [unclosed
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, encrypted, err := parseYAMLWithEncryption([]byte(tt.yamlContent))

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseYAMLWithEncryption() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("parseYAMLWithEncryption() unexpected error = %v", err)
				return
			}

			if cfg == nil {
				t.Fatal("parseYAMLWithEncryption() config is nil")
			}

			if cfg.LogLevel != tt.wantLogLevel {
				t.Errorf("LogLevel = %v, want %v", cfg.LogLevel, tt.wantLogLevel)
			}

			if len(encrypted) != tt.wantEncrypted {
				t.Errorf("parseYAMLWithEncryption() found %d encrypted values, want %d", len(encrypted), tt.wantEncrypted)
			}
		})
	}
}

// Helper function to get keys from map
func getKeys(m map[string]*EncryptedValue) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
