# Data Model: Configuration Management System

**Feature**: 002-config-management
**Date**: 2025-11-02
**Phase**: 1 (Design & Contracts)

## Overview

This document defines the complete data model for the configuration management system, including all entities, their fields, validation rules, relationships, and state transitions. The model supports 56 functional requirements across 9 configuration categories.

## Core Entities

### 1. Config (Main Configuration Container)

**Description**: The root configuration object containing all settings across all categories. This is the primary entity loaded, merged, validated, and passed to the application.

**Fields**:
```go
type Config struct {
    // Meta
    Version      string    `yaml:"version" toml:"version"`           // Config schema version (e.g., "1.0")
    LoadedFrom   string    `yaml:"-" toml:"-"`                       // Path to file that provided config (for debugging)
    LoadedAt     time.Time `yaml:"-" toml:"-"`                       // Timestamp of load (for hot-reload tracking)

    // UI Settings (FR-020 through FR-025)
    Theme         string      `yaml:"theme" toml:"theme" validate:"oneof=default dark light solarized" default:"default"`
    ColorScheme   ColorScheme `yaml:"colorScheme" toml:"color_scheme"`
    CompactMode   bool        `yaml:"compactMode" toml:"compact_mode" default:"false"`
    ShowHints     bool        `yaml:"showHints" toml:"show_hints" default:"true"`
    ShowLineNumbers bool      `yaml:"showLineNumbers" toml:"show_line_numbers" default:"false"`
    DateFormat    string      `yaml:"dateFormat" toml:"date_format" validate:"dateformat" default:"2006-01-02"`

    // Keybindings (FR-026 through FR-030)
    Keybindings        map[string]KeyBinding `yaml:"keybindings" toml:"keybindings"`
    KeybindingProfile  string                `yaml:"keybindingProfile" toml:"keybinding_profile" validate:"oneof=default vim emacs" default:"default"`

    // Performance (FR-031 through FR-034)
    MaxConcurrentOps int           `yaml:"maxConcurrentOps" toml:"max_concurrent_ops" validate:"min=1,max=16" default:"4"`
    CacheSize        int           `yaml:"cacheSize" toml:"cache_size" validate:"min=0" default:"50"` // MB
    RefreshInterval  time.Duration `yaml:"refreshInterval" toml:"refresh_interval" validate:"min=0" default:"0"` // 0 = disabled
    Timeouts         Timeouts      `yaml:"timeouts" toml:"timeouts"`

    // Dotnet CLI Integration (FR-035 through FR-038)
    DotnetPath      string `yaml:"dotnetPath" toml:"dotnet_path" default:""` // Empty = auto-detect from PATH
    DotnetVerbosity string `yaml:"dotnetVerbosity" toml:"dotnet_verbosity" validate:"oneof=quiet minimal normal detailed diagnostic" default:"minimal"`

    // Logging (FR-039 through FR-042)
    LogLevel    string       `yaml:"logLevel" toml:"log_level" validate:"oneof=debug info warn error" default:"info"`
    LogDir      string       `yaml:"logDir" toml:"log_dir" default:""` // Empty = platform default
    LogFormat   string       `yaml:"logFormat" toml:"log_format" validate:"oneof=text json" default:"text"`
    LogRotation LogRotation  `yaml:"logRotation" toml:"log_rotation"`

    // Hot-Reload (FR-043 through FR-049)
    HotReload bool `yaml:"hotReload" toml:"hot_reload" default:"false"` // Disabled by default for safety
}
```

**Validation Rules**:
- `Theme`: Must be one of predefined themes (default, dark, light, solarized)
- `DateFormat`: Must be valid Go time layout string (validated via `time.Parse()`)
- `MaxConcurrentOps`: Range [1, 16] inclusive
- `CacheSize`: Non-negative integer (megabytes)
- `RefreshInterval`: Non-negative duration (0 = disabled)
- `DotnetVerbosity`: Must match dotnet CLI verbosity levels
- `LogLevel`: Must be valid log level
- `LogFormat`: Must be text or json

**Default Values**:
- Applied when setting is missing or invalid (semantic validation failure)
- Hardcoded in `internal/config/defaults.go`
- See struct tags `default:` for specific values

**Relationships**:
- Contains `ColorScheme` (1:1)
- Contains `Timeouts` (1:1)
- Contains `LogRotation` (1:1)
- Contains many `KeyBinding` entries (1:many via map)

---

### 2. ColorScheme

**Description**: Customizable colors for all UI elements. Supports hex color codes (#RRGGBB format).

**Fields**:
```go
type ColorScheme struct {
    Border      string `yaml:"border" toml:"border" validate:"hexcolor" default:"#FFFFFF"`
    BorderFocus string `yaml:"borderFocus" toml:"border_focus" validate:"hexcolor" default:"#00FF00"`
    Text        string `yaml:"text" toml:"text" validate:"hexcolor" default:"#FFFFFF"`
    TextDim     string `yaml:"textDim" toml:"text_dim" validate:"hexcolor" default:"#808080"`
    Background  string `yaml:"background" toml:"background" validate:"hexcolor" default:"#000000"`
    Highlight   string `yaml:"highlight" toml:"highlight" validate:"hexcolor" default:"#FFFF00"`
    Error       string `yaml:"error" toml:"error" validate:"hexcolor" default:"#FF0000"`
    Warning     string `yaml:"warning" toml:"warning" validate:"hexcolor" default:"#FFA500"`
    Success     string `yaml:"success" toml:"success" validate:"hexcolor" default:"#00FF00"`
    Info        string `yaml:"info" toml:"info" validate:"hexcolor" default:"#00FFFF"`
}
```

**Validation Rules**:
- All fields: Must be valid hex color code matching regex `^#[0-9A-Fa-f]{6}$`
- Invalid colors fall back to corresponding default color (not entire ColorScheme)

**Default Values**: Standard terminal colors (see struct tags)

---

### 3. KeyBinding

**Description**: Maps an action name to a key combination. Supports standard key representations.

**Fields**:
```go
type KeyBinding struct {
    Action      string `yaml:"action" toml:"action"`           // Action name (e.g., "quit", "refresh", "search")
    Key         string `yaml:"key" toml:"key"`                 // Key combination (e.g., "Ctrl+C", "F5", "q")
    Description string `yaml:"description" toml:"description"` // Human-readable description for help screen
    Context     string `yaml:"context" toml:"context"`         // UI context where binding is active (e.g., "global", "package-list")
}
```

**Validation Rules**:
- `Action`: Non-empty string, must match known action in application
- `Key`: Must match key representation format (e.g., `Ctrl+C`, `Alt+Enter`, `F1-F12`, single char)
- `Context`: Must be valid UI context or "global"
- Conflict detection: Same key cannot map to multiple actions in same context (FR-028)

**Examples**:
```yaml
keybindings:
  quit:
    action: "quit"
    key: "q"
    description: "Quit application"
    context: "global"
  refresh:
    action: "refresh"
    key: "F5"
    description: "Refresh package list"
    context: "package-list"
```

---

### 4. Timeouts

**Description**: Timeout durations for different operation types.

**Fields**:
```go
type Timeouts struct {
    NetworkRequest time.Duration `yaml:"networkRequest" toml:"network_request" validate:"min=1s" default:"30s"`
    DotnetCLI      time.Duration `yaml:"dotnetCLI" toml:"dotnet_cli" validate:"min=1s" default:"60s"`
    FileOperation  time.Duration `yaml:"fileOperation" toml:"file_operation" validate:"min=100ms" default:"5s"`
}
```

**Validation Rules**:
- All fields: Positive duration (minimum varies by field)
- Invalid durations fall back to defaults

**Default Values**: Conservative timeouts suitable for typical operations

---

### 5. LogRotation

**Description**: Log file rotation settings.

**Fields**:
```go
type LogRotation struct {
    MaxSize    int `yaml:"maxSize" toml:"max_size" validate:"min=1" default:"10"`       // Megabytes
    MaxAge     int `yaml:"maxAge" toml:"max_age" validate:"min=1" default:"30"`          // Days
    MaxBackups int `yaml:"maxBackups" toml:"max_backups" validate:"min=0" default:"3"`   // Number of old logs to keep (0 = keep all)
    Compress   bool `yaml:"compress" toml:"compress" default:"true"`                      // Compress rotated logs
}
```

**Validation Rules**:
- `MaxSize`: Minimum 1 MB
- `MaxAge`: Minimum 1 day
- `MaxBackups`: Non-negative (0 means keep all backups)

**Default Values**: Reasonable defaults for typical usage (10MB files, 30 day retention, 3 backups)

---

### 6. ConfigSource

**Description**: Represents one of the four configuration sources with its precedence level and raw values.

**Fields**:
```go
type ConfigSource struct {
    Name       string                 // Source name: "defaults", "file", "env", "cli"
    Precedence int                    // Precedence level (1=lowest/defaults, 4=highest/cli)
    Data       map[string]interface{} // Raw key-value pairs from this source
    FilePath   string                 // Path to file (if source is "file"), empty otherwise
    LoadedAt   time.Time              // Timestamp when source was loaded
}
```

**Validation Rules**: N/A (internal tracking structure, not user-facing)

**Usage**:
- Track provenance for `--print-config` debugging
- Support merge algorithm with explicit precedence
- Useful for error messages ("Setting X from environment variable overrode file value Y")

---

### 7. MergedConfig

**Description**: Internal representation of configuration after merging all sources. Tracks which source provided each setting.

**Fields**:
```go
type MergedConfig struct {
    Final      Config                  // Final merged configuration
    Provenance map[string]ConfigSource // Maps config key path to source that provided it
    MergedAt   time.Time               // Timestamp of merge operation
}
```

**Validation Rules**: N/A (internal structure)

**Usage**:
- Debugging with `--print-config` shows provenance
- Hot-reload can compare timestamps to detect staleness

---

### 8. ValidationError

**Description**: Details about a configuration validation failure.

**Fields**:
```go
type ValidationError struct {
    Key         string      // Config key path (e.g., "maxConcurrentOps", "colorScheme.border")
    Value       interface{} // Invalid value that was provided
    Constraint  string      // Constraint that failed (e.g., "must be in range [1, 16]", "must be hex color")
    SuggestedFix string      // Suggested correction (e.g., "use value between 1 and 16")
    Severity     string      // "error" (blocks startup) or "warning" (falls back to default)
    DefaultUsed interface{} // Default value being used (if severity is "warning")
}
```

**Validation Rules**: N/A (error reporting structure)

**Usage**:
- Collect all validation errors during semantic validation phase
- Log warnings for non-blocking errors (FR-013)
- Return blocking errors to prevent startup (FR-009, FR-010)

**Example**:
```go
ValidationError{
    Key:         "maxConcurrentOps",
    Value:       999,
    Constraint:  "must be in range [1, 16]",
    SuggestedFix: "use value between 1 and 16",
    Severity:     "warning",
    DefaultUsed:  4,
}
```

---

### 9. EncryptedValue

**Description**: An encrypted configuration value stored in config file with metadata.

**Fields**:
```go
type EncryptedValue struct {
    Ciphertext   []byte // AES-256-GCM encrypted data (base64-encoded in config file)
    Nonce        []byte // 12-byte nonce for GCM (embedded in ciphertext or stored separately)
    KeyID        string // Identifier for encryption key (e.g., "prod", "dev")
    Algorithm    string // Encryption algorithm used (e.g., "AES-256-GCM")
    EncryptedAt  time.Time // Timestamp of encryption (for key rotation tracking)
}
```

**Validation Rules**:
- `Ciphertext`: Non-empty, valid base64
- `Nonce`: Exactly 12 bytes for GCM
- `KeyID`: Non-empty, must match key available in keychain or environment
- `Algorithm`: Must be "AES-256-GCM" (only supported algorithm)

**YAML/TOML Representation**:
```yaml
apiKey: !encrypted "base64-encoded-ciphertext-with-embedded-nonce"
```

**Decryption Process**:
1. Detect `!encrypted` tag during parsing
2. Decode base64 ciphertext
3. Extract nonce (first 12 bytes of ciphertext)
4. Look up decryption key in keychain by `KeyID` (or fall back to env var `LAZYNUGET_ENCRYPTION_KEY`)
5. Decrypt using AES-256-GCM
6. Replace EncryptedValue with decrypted string in Config struct
7. On decryption failure: Log warning, use default value for setting (FR-014)

---

### 10. ConfigSchema

**Description**: Schema definition for all valid configuration keys, their types, constraints, default values, and hot-reload support.

**Fields**:
```go
type ConfigSchema struct {
    Settings map[string]SettingSchema // Map of setting path to schema
}

type SettingSchema struct {
    Path          string      // Dot-separated path (e.g., "colorScheme.border", "maxConcurrentOps")
    Type          reflect.Type // Go type of setting
    Constraints   []Constraint // Validation constraints (range, enum, regex, etc.)
    Default       interface{}  // Default value
    HotReloadable bool         // Can this setting be changed via hot-reload without restart?
    Description   string       // Human-readable description for documentation
}

type Constraint struct {
    Type    string      // "range", "enum", "regex", "min", "max", "hexcolor", etc.
    Params  interface{} // Constraint parameters (e.g., {min: 1, max: 16} for range)
    Message string      // Error message if constraint fails
}
```

**Validation Rules**: N/A (defines validation rules for other entities)

**Usage**:
- Single source of truth for all setting schemas
- Validation engine walks Config struct, checks each field against schema
- Hot-reload checks `HotReloadable` flag before applying changes (FR-049)
- Auto-generate documentation from schema descriptions

**Example Schema Entry**:
```go
SettingSchema{
    Path:          "maxConcurrentOps",
    Type:          reflect.TypeOf(0), // int
    Constraints:   []Constraint{
        {Type: "range", Params: map[string]int{"min": 1, "max": 16}, Message: "must be between 1 and 16"},
    },
    Default:       4,
    HotReloadable: true,
    Description:   "Maximum number of concurrent operations (default: 4, range: 1-16)",
}
```

---

## Relationships

```
Config (1) ──contains──> (1) ColorScheme
Config (1) ──contains──> (1) Timeouts
Config (1) ──contains──> (1) LogRotation
Config (1) ──contains──> (many) KeyBinding

ConfigSource (4 instances) ──merges into──> (1) MergedConfig
MergedConfig ──validates with──> ConfigSchema ──produces──> (many) ValidationError
MergedConfig ──final result──> Config

Config ──may contain──> (many) EncryptedValue
EncryptedValue ──decrypts to──> (1) string
```

## State Transitions

### Configuration Lifecycle States

```
┌─────────────────┐
│   Unloaded      │  Initial state, no config loaded
└────────┬────────┘
         │ Load()
         v
┌─────────────────┐
│   Loading       │  Reading files, parsing, merging sources
└────────┬────────┘
         │
         ├─ Syntax Error ──> [ERROR: Exit with code 1]
         │
         ├─ File Too Large ──> [ERROR: Exit with code 1]
         │
         ├─ Both Formats Present ──> [ERROR: Exit with code 1]
         │
         v
┌─────────────────┐
│   Validating    │  Semantic validation against schema
└────────┬────────┘
         │
         ├─ All Valid ──────────────┐
         │                          v
         ├─ Some Invalid ─────> [WARN: Log errors, use defaults]
         │                          │
         v                          v
┌─────────────────┐          ┌─────────────────┐
│     Loaded      │          │     Loaded      │
│  (All Valid)    │          │  (Partial)      │
└────────┬────────┘          └────────┬────────┘
         │                            │
         └────────────┬───────────────┘
                      │
                      │ Hot-Reload Enabled?
                      │
         ┌────────────┴────────────┐
         │ No                      │ Yes
         v                         v
┌─────────────────┐      ┌─────────────────┐
│     Active      │      │     Watching    │  File watcher active
│   (Static)      │      │   (Hot-Reload)  │
└─────────────────┘      └────────┬────────┘
                                  │
                                  │ File Changed
                                  v
                         ┌─────────────────┐
                         │   Reloading     │  Re-parse, re-validate
                         └────────┬────────┘
                                  │
                                  ├─ Valid ──> Update Config, Notify User
                                  │
                                  ├─ Invalid ──> Keep Old Config, Warn User
                                  │
                                  └──> Back to Watching
```

### Hot-Reload State Machine

```
[Enabled] ──────────────────────> [Watching File]
   │                                      │
   │                                      │ File Event Detected
   │                                      v
   │                              [Debounce Window (100ms)]
   │                                      │
   │                                      │ Timer Expired
   │                                      v
   │                              [Load & Validate New Config]
   │                                      │
   │                                      ├─ Valid ──> [Apply & Notify]
   │                                      │                 │
   │                                      │                 v
   │                                      │           [Back to Watching]
   │                                      │
   │                                      ├─ Invalid ──> [Log Error, Keep Old]
   │                                      │                 │
   │                                      │                 v
   │                                      │           [Back to Watching]
   │                                      │
   │                                      └─ File Deleted ──> [Fallback to Defaults, Notify]
   │                                                          │
   │                                                          v
   │                                                    [Stop Watching]
   │
   └─────────────────────────────────────────────────> [Disabled (Default)]
```

## Validation Rules Summary

### Blocking Errors (Prevent Startup)
- File size > 10 MB (FR-009)
- YAML/TOML syntax errors (FR-010)
- Both YAML and TOML files present (FR-005)
- Config file unreadable (permissions, not found if explicitly specified via `--config`)

### Non-Blocking Errors (Warn + Fallback to Default)
- Setting out of range (e.g., `maxConcurrentOps: 999`)
- Invalid type (e.g., `maxConcurrentOps: "abc"`)
- Invalid format (e.g., `colorScheme.border: "#GGGGGG"`)
- Unknown config keys (warn, ignore)
- Encrypted value decryption failure
- Keybinding conflicts (use first encountered, warn about duplicates)

### Graceful Fallbacks
- Empty config file → Use all defaults
- Missing config file → Use all defaults (no error)
- Partial config file → Use provided settings, defaults for rest
- Environment variable not set → Ignore, use next precedence level
- CLI flag not provided → Ignore, use next precedence level
- Keychain unavailable → Fall back to env var for encryption key, warn if neither available

## Thread Safety

- **Config struct**: Immutable after load (except hot-reload creates new instance)
- **Hot-reload**: New `Config` instance created, atomic pointer swap in application
- **File watcher**: Runs in separate goroutine, sends events to main thread via channel
- **Validation**: Pure functions, no shared mutable state

## Memory Considerations

- Config struct size: ~2-5 KB (50+ settings with reasonable sizes)
- Keybindings map: ~1-2 KB for typical 20-30 bindings, ~10-20 KB for thousands of custom bindings
- Provenance tracking: ~1 KB (stores source name per setting)
- Validation errors: Transient (logged and discarded after validation)
- Total memory overhead: ~5-10 KB typical, ~30-50 KB maximum (with extensive keybinding customization)

## Performance Characteristics

- **Parsing**: O(n) where n = config file size (YAML/TOML parsers are linear)
- **Merging**: O(m) where m = number of settings (~50) - reflection-based field walking
- **Validation**: O(m × c) where c = average constraints per setting (~2-3) - walk all fields, check all constraints
- **Hot-reload**: Same as initial load (re-parse, re-merge, re-validate) - ~100ms total

## Cross-Platform Considerations

### Path Handling
- Use `filepath.Join()` for all path operations
- Platform-specific defaults via `runtime.GOOS` switch
- Test file operations on all three platforms (Windows paths with backslashes, Unix with forward slashes)

### Keychain Integration
- macOS: Keychain Services API via `go-keyring`
- Windows: Credential Manager via `go-keyring`
- Linux: Secret Service D-Bus API via `go-keyring`
- Fallback: Environment variable if keychain unavailable (headless/CI environments)

### File Watching
- Linux: inotify via `fsnotify`
- macOS: FSEvents via `fsnotify`
- Windows: ReadDirectoryChangesW via `fsnotify`
- Edge case: Network file systems (NFS, SMB) may have delayed notifications or not support watching - document limitation

## Example: Complete Config File (YAML)

```yaml
version: "1.0"

# UI Settings
theme: dark
colorScheme:
  border: "#333333"
  borderFocus: "#00FF00"
  text: "#FFFFFF"
  textDim: "#808080"
  background: "#1E1E1E"
  highlight: "#FFFF00"
  error: "#FF0000"
  warning: "#FFA500"
  success: "#00FF00"
  info: "#00FFFF"

compactMode: false
showHints: true
showLineNumbers: true
dateFormat: "2006-01-02 15:04:05"

# Keybindings
keybindingProfile: vim
keybindings:
  quit:
    action: "quit"
    key: "q"
    description: "Quit application"
    context: "global"
  refresh:
    action: "refresh"
    key: "Ctrl+R"
    description: "Refresh package list"
    context: "package-list"

# Performance
maxConcurrentOps: 8
cacheSize: 100
refreshInterval: 5m
timeouts:
  networkRequest: 30s
  dotnetCLI: 60s
  fileOperation: 5s

# Dotnet CLI
dotnetPath: "/usr/local/share/dotnet/dotnet"
dotnetVerbosity: normal

# Logging
logLevel: debug
logDir: "/var/log/lazynuget"
logFormat: json
logRotation:
  maxSize: 10
  maxAge: 30
  maxBackups: 5
  compress: true

# Hot-Reload
hotReload: true
```

## Example: Encrypted Value in Config

```yaml
# Regular setting
dotnetPath: "/usr/local/share/dotnet/dotnet"

# Encrypted API key
nugetApiKey: !encrypted "AES256GCM:dev:AbCdEf1234567890+nonce_and_ciphertext_base64=="
```

When parsed:
- Detect `!encrypted` YAML tag
- Parse format: `<algorithm>:<key_id>:<base64_data>`
- Create `EncryptedValue` struct
- During validation, decrypt using key from keychain (identifier "dev")
- Replace with decrypted string in final `Config` struct
