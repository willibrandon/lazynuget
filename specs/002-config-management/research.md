# Research: Configuration Management System

**Feature**: 002-config-management
**Date**: 2025-11-02
**Phase**: 0 (Architecture Research)

## Overview

This document consolidates research findings for implementing a hierarchical configuration system with multi-source merging, validation, encryption, and hot-reload capabilities. All research aims to resolve technical unknowns and establish best practices for Go-based configuration management.

## Research Areas

### 1. YAML vs TOML Parsing Libraries (Go)

**Decision**: Use `gopkg.in/yaml.v3` for YAML and `github.com/BurntSushi/toml` for TOML

**Rationale**:
- **yaml.v3**: Most widely adopted YAML library for Go (30k+ stars), supports YAML 1.2 spec, provides clear error messages with line numbers (FR-010 requirement), handles large files efficiently, well-maintained
- **BurntSushi/toml**: De facto standard for TOML in Go (4.5k+ stars), fully implements TOML v1.0.0, excellent error reporting, battle-tested in production systems
- Both libraries have zero-allocation streaming parsers suitable for <100ms parsing target (SC-010)
- Both provide strong typing and validation hooks for custom constraint checking

**Alternatives Considered**:
- `go-yaml/yaml` (yaml.v2): Older version, less active maintenance, doesn't support YAML 1.2 features
- `pelletier/go-toml`: Slower parsing, less adoption, though supports TOML v1.0.0
- JSON-based config: Rejected because JSON doesn't support comments (important for user-edited config files) and less human-friendly syntax

**Implementation Notes**:
- Use `yaml.Unmarshal()` and `toml.Unmarshal()` for initial parsing into Go structs
- Implement custom `UnmarshalYAML()` / `UnmarshalTOML()` methods for types requiring special validation (durations, hex colors, encrypted values)
- Parse errors from both libraries include file line numbers - surface these directly in error messages (FR-010)

### 2. File Watching for Hot-Reload

**Decision**: Use `github.com/fsnotify/fsnotify` v1.7+

**Rationale**:
- Cross-platform abstraction over OS-specific APIs (inotify on Linux, FSEvents on macOS, ReadDirectoryChangesW on Windows)
- Mature library (9k+ stars), used in production by major projects (Kubernetes, Hugo, Viper)
- Event-driven model (not polling) - minimal CPU/battery impact
- Handles edge cases: file deletion, file rename, rapid successive writes (debouncing)
- <3ms notification latency on modern systems - easily meets <3s hot-reload requirement (FR-045)

**Alternatives Considered**:
- Manual polling with `os.Stat()`: Simple but wasteful (constant I/O), poor battery life, high latency (poll interval vs instant notification)
- Platform-specific APIs directly: Not cross-platform, would require separate implementations for Windows/macOS/Linux
- `radovskyb/watcher`: Less mature, fewer features than fsnotify

**Implementation Notes**:
- Watch single config file path, not directory (reduces false positives from unrelated file changes)
- Debounce rapid writes: Wait 100ms after last event before triggering reload (handles editors saving temp files)
- Handle `fsnotify.Remove` event (file deleted) by falling back to defaults and notifying user
- Gracefully handle watcher errors (e.g., file moved outside watched directory, permissions changed)

### 3. Encryption for Sensitive Config Values

**Decision**: AES-256-GCM with keys stored in platform keychain via `github.com/zalando/go-keyring`

**Rationale**:
- **AES-256-GCM**: NIST-approved authenticated encryption, prevents tampering, widely supported, Go standard library implementation (`crypto/aes`, `crypto/cipher`)
- **go-keyring**: Unified API for all three platforms:
  - macOS: Keychain Services API
  - Windows: Windows Credential Manager (CredWrite/CredRead)
  - Linux: Secret Service API (D-Bus) - works with GNOME Keyring, KDE Wallet
- Separation of concerns: Encryption keys never stored in config file, only encrypted ciphertext
- Fallback strategy: If keychain unavailable, use environment variable for key (warn user about reduced security)

**Alternatives Considered**:
- Plain text with file permissions: Insecure - config files often committed to version control or backed up to cloud storage
- Age encryption: Requires external binary, not native Go, more complex key management
- Custom keychain integration: Would require platform-specific code for all three OSes, reinventing wheel

**Implementation Notes**:
- Encrypted value format in config file: `!encrypted <base64-ciphertext>` (YAML custom tag syntax)
- Key derivation: Use PBKDF2 with user-specified master password, derive 256-bit key, store derived key in keychain
- Key identifier stored in encrypted value metadata - allows multiple keys (e.g., dev vs prod)
- Nonce/IV embedded in ciphertext (12 bytes for GCM) - no need to store separately
- `lazynuget encrypt-value` CLI command for encrypting secrets before adding to config

### 4. Environment Variable Parsing with Nested Keys

**Decision**: Use underscore notation with custom parser (e.g., `LAZYNUGET_COLOR_SCHEME_BORDER=#FF0000`)

**Rationale**:
- Standard convention in 12-factor apps and containerized environments
- No ambiguity: Underscores map to struct field nesting (e.g., `ColorScheme.Border`)
- Type-safe: Parse each env var according to its schema type (bool, int, duration, string)
- Viper library demonstrates this pattern works reliably across thousands of projects

**Alternatives Considered**:
- JSON in env vars (e.g., `LAZYNUGET_COLOR_SCHEME='{"border":"#FF0000"}'`): Hard to escape properly in shell, error-prone
- Separate env var for each leaf setting: Would require 50+ env vars with no hierarchical organization
- Dot notation: Conflicts with shell variable naming rules (dots not allowed in bash/zsh)

**Implementation Notes**:
- Split `LAZYNUGET_` prefix, then split remaining by `_` to get path: `["COLOR", "SCHEME", "BORDER"]`
- Walk config struct using reflection, set field at path if exists, skip if unknown (with warning)
- Type conversion based on field type: `strconv.ParseBool`, `strconv.ParseInt`, `time.ParseDuration`, hex color parsing
- Case-insensitive matching (env vars typically uppercase, struct fields PascalCase)

### 5. Configuration Validation Strategy

**Decision**: Two-phase validation (syntax check → semantic check) with fail-fast for syntax, graceful fallback for semantics

**Rationale**:
- **Syntax errors** (invalid YAML/TOML, file too large, both formats present): Block startup - indicates corrupted/misconfigured system (FR-009, FR-010, FR-005)
- **Semantic errors** (out of range, wrong type, invalid format): Fall back to defaults - likely user typo, shouldn't prevent app from running (FR-012, FR-013)
- Clear separation allows targeted error messages: "Config file syntax invalid at line 42" vs "Setting X out of range (got Y, expected 1-16), using default Z"

**Alternatives Considered**:
- Fail-fast all errors: Too strict, minor config typo would prevent startup - poor UX
- Ignore all errors silently: Insecure, user wouldn't know their config isn't being respected
- Prompt user to fix errors: Not suitable for headless/CI environments, violates automation principle

**Implementation Notes**:
- Phase 1 (syntax): Parser returns error with line number → log error, exit with code 1
- Phase 2 (semantic): Validator returns `[]ValidationError` with details → log warnings (FR-013), substitute defaults, continue
- Validation rules stored in schema: Each setting has `Validate(value) error` method (range check, regex match, etc.)
- Group validation errors by category in log output for readability

### 6. Configuration Merge Precedence Algorithm

**Decision**: Four-layer merge with explicit precedence (defaults < file < env < CLI), last-write-wins within layer

**Rationale**:
- Precedence order matches common expectations: more explicit sources override less explicit
- CLI flags: Most explicit (user typed command), highest precedence
- Env vars: Deployment-specific (CI/CD, Docker), override file
- File: Persistent user preferences, override defaults
- Defaults: Hardcoded fallback, lowest precedence
- Simple to reason about, matches 12-factor app methodology

**Alternatives Considered**:
- First-write-wins: Counterintuitive (defaults would override CLI flags)
- Separate merge per setting type: Too complex, hard to predict behavior
- Conditional precedence (e.g., env vars only override file for specific keys): Increases complexity without clear benefit

**Implementation Notes**:
- Implement as sequential merge: `result = merge(merge(merge(defaults, file), env), cli)`
- Use reflection to walk struct fields, overwrite non-zero values from higher precedence source
- Track provenance: Store which source provided each setting (useful for `--print-config` debugging)
- Handle slice/map merging: Replace entire slice/map (don't merge elements) for predictability

## Best Practices

### Go Configuration Patterns

**Struct Tags for Schema Definition**:
```go
type Config struct {
    MaxConcurrentOps int `yaml:"maxConcurrentOps" toml:"max_concurrent_ops" validate:"min=1,max=16" default:"4"`
    Theme string `yaml:"theme" toml:"theme" validate:"oneof=default dark light solarized" default:"default"`
}
```

**Validation with go-playground/validator**:
- Use struct tags for common validations (range, enum, regex)
- Custom validators for complex rules (keybinding conflicts, color format)

**Immutable Config After Load**:
- Return `*Config` from `Load()` function, never mutate after creation (except hot-reload which creates new instance)
- Thread-safe reads without locks (immutable data)

**Error Wrapping with Context**:
```go
if err := parser.Parse(data); err != nil {
    return fmt.Errorf("parsing config file %s: %w", path, err)
}
```

### Cross-Platform File Paths

- Use `filepath.Join()` for all path construction (handles `/` vs `\` automatically)
- Use `os.UserHomeDir()` to get home directory (cross-platform)
- Platform detection via `runtime.GOOS`: "windows", "darwin", "linux"
- Example:
  ```go
  var configDir string
  switch runtime.GOOS {
  case "darwin":
      configDir = filepath.Join(home, "Library", "Application Support", "lazynuget")
  case "windows":
      configDir = filepath.Join(os.Getenv("APPDATA"), "lazynuget")
  default: // Linux and others
      xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
      if xdgConfigHome == "" {
          xdgConfigHome = filepath.Join(home, ".config")
      }
      configDir = filepath.Join(xdgConfigHome, "lazynuget")
  }
  ```

### Testing Strategies

**Table-Driven Tests for Validators**:
```go
tests := []struct {
    name    string
    value   int
    wantErr bool
}{
    {"valid min", 1, false},
    {"valid max", 16, false},
    {"below min", 0, true},
    {"above max", 17, true},
}
```

**Integration Tests with Temporary Config Files**:
- Use `t.TempDir()` for isolated test environments
- Create fixture config files in `tests/fixtures/configs/`
- Test full lifecycle: write config → load → validate → assert expected values

**Cross-Platform CI Testing**:
- GitHub Actions matrix: `os: [ubuntu-latest, macos-latest, windows-latest]`
- Verify path handling on all platforms
- Test keychain integration (may need mocks for CI environments without GUI)

## Implementation Checklist

Phase 0 (Research) ✅ COMPLETE:
- [x] Evaluate YAML/TOML parsing libraries
- [x] Research file watching solutions
- [x] Design encryption approach
- [x] Define environment variable parsing strategy
- [x] Design validation strategy
- [x] Define merge precedence algorithm

Phase 1 (Design & Contracts):
- [ ] Define complete data model (all 50+ settings with types and constraints)
- [ ] Design config schema with validation rules
- [ ] Define contracts for config loading, merging, validation
- [ ] Create quickstart guide with examples

Phase 2 (Implementation):
- [ ] Implement parser for YAML and TOML
- [ ] Implement validator with schema
- [ ] Implement merger with precedence
- [ ] Implement environment variable parsing
- [ ] Implement encryption/decryption with keychain
- [ ] Implement file watching with hot-reload
- [ ] Write unit tests (target >80% coverage)
- [ ] Write integration tests (full lifecycle scenarios)
- [ ] Test on Windows, macOS, Linux

## References

- [gopkg.in/yaml.v3 Documentation](https://pkg.go.dev/gopkg.in/yaml.v3)
- [BurntSushi/toml Documentation](https://pkg.go.dev/github.com/BurntSushi/toml)
- [fsnotify Documentation](https://pkg.go.dev/github.com/fsnotify/fsnotify)
- [go-keyring Documentation](https://pkg.go.dev/github.com/zalando/go-keyring)
- [AES-GCM in Go](https://pkg.go.dev/crypto/cipher#NewGCM)
- [12-Factor App: Config](https://12factor.net/config)
- [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html)
