# Implementation Plan: Configuration Management System

**Branch**: `002-config-management` | **Date**: 2025-11-02 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/002-config-management/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Implement a hierarchical configuration system that loads and merges settings from four sources (defaults, file, environment variables, CLI flags) with proper precedence. The system supports YAML and TOML formats, validates all settings with graceful fallback to defaults, includes encrypted value support for sensitive data, and provides optional hot-reload capability. Configuration covers UI settings, keybindings, color schemes, performance tuning, dotnet CLI paths, and logging options. The design prioritizes safety (opt-in hot-reload), cross-platform compatibility (platform-specific config locations), and user experience (clear validation errors, fast parsing <100ms).

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**:
- `gopkg.in/yaml.v3` - YAML parsing and serialization
- `github.com/BurntSushi/toml` - TOML parsing
- `github.com/fsnotify/fsnotify` - Cross-platform file watching (inotify/FSEvents/ReadDirectoryChangesW)
- Platform keychain integration: `github.com/zalando/go-keyring` (macOS Keychain, Windows Credential Manager, Linux Secret Service)
- Existing project dependencies (from spec 001): logging, platform detection

**Storage**: File-based configuration (YAML/TOML), system keychain for encryption keys
**Testing**: Go testing framework (`go test`), table-driven tests for validation, integration tests for full config lifecycle
**Target Platform**: Cross-platform (Windows, macOS, Linux) - identical behavior required
**Project Type**: Single project (LazyNuGet TUI application)
**Performance Goals**:
- Config file parsing: <100ms for typical files (<100 KB)
- Hot-reload latency: <3 seconds from file modification to application
- Startup config load: <500ms (part of <200ms total startup budget)
- Memory overhead: Minimal (config held in memory, ~1-2MB for typical 50+ settings)

**Constraints**:
- File size limit: 10 MB maximum (prevents resource exhaustion)
- Cross-platform path handling: Use `filepath` package for all operations
- No external config servers: All sources local (file, env vars, CLI flags)
- Encryption: AES-256-GCM for encrypted values
- Hot-reload: Opt-in only (disabled by default for safety)

**Scale/Scope**:
- 50+ distinct configuration settings across 9 categories
- Support thousands of keybinding customizations (user-defined profiles)
- Config file watching for single file (not recursive directory watching)
- Graceful degradation: Continue startup even with invalid config (fall back to defaults)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Discoverability ✅ PASS
- **Assessment**: Configuration is not user-facing UI but internal system. However, validation errors provide clear, actionable messages (FR-014). The `--print-config` and `--validate-config` flags allow users to inspect and verify configuration without reading source code.
- **Status**: No concerns. Clear error messages and debugging tools meet discoverability for config system.

### Principle II: Simplicity & 80/20 Rule ✅ PASS
- **Assessment**: 80% use case is "use defaults, maybe customize theme/keybindings via file." This requires ZERO configuration for default case (P1), and simple file editing for customization (P2). Advanced features (encryption, hot-reload) are opt-in and don't complicate common workflows.
- **Status**: No concerns. Complexity is progressive - defaults work immediately, advanced features hidden until needed.

### Principle III: Safety & Confirmation ✅ PASS
- **Assessment**:
  - Syntax errors (YAML/TOML invalid, file too large, both formats present) block startup with clear errors - prevents running with corrupted config
  - Semantic errors (out of range, invalid values) fall back to defaults with warnings - prevents config typos from blocking startup
  - Hot-reload is opt-in (disabled by default) - prevents unexpected behavior changes during operation
  - Encrypted values never logged in plain text (FR-018)
- **Status**: No concerns. Design prioritizes safety through sensible defaults and explicit opt-in for potentially disruptive features.

### Principle IV: Cross-Platform Excellence ✅ PASS
- **Assessment**:
  - Platform-specific default paths (FR-006): macOS `~/Library/Application Support`, Linux XDG `~/.config`, Windows `%APPDATA%`
  - Cross-platform file watching via `fsnotify` (abstracts inotify/FSEvents/ReadDirectoryChangesW)
  - Cross-platform keychain integration via `go-keyring` (macOS Keychain, Windows Credential Manager, Linux Secret Service)
  - All file operations use `filepath` package for correct path separators
  - Test suite must validate behavior on Windows, macOS, Linux (constitution requirement)
- **Status**: No concerns. Design explicitly addresses platform differences and uses cross-platform libraries.

### Principle V: Performance & Responsiveness ✅ PASS
- **Assessment**:
  - Parsing <100ms for typical files (SC-010) - well within budget
  - Startup config load target <500ms - fits within <200ms total startup budget (p95)
  - Hot-reload 3s maximum latency - acceptable for opt-in feature
  - Memory minimal: ~1-2MB for 50+ settings - negligible overhead
  - File watching uses OS-level notifications (not polling) - efficient
- **Status**: No concerns. Performance targets are realistic and monitored via success criteria.

### Principle VI: Conformity with dotnet CLI ✅ PASS
- **Assessment**: Configuration system supports `dotnetPath` and `dotnetVerbosity` settings (FR-035, FR-037) that integrate with existing dotnet CLI detection and execution. Does not reinvent dotnet CLI, merely provides configuration for it. Respects .NET conventions via configurable paths.
- **Status**: No concerns. Config system augments existing dotnet CLI integration, doesn't replace it.

### Principle VII: Clean, Testable, Maintainable Code ✅ PASS
- **Assessment**:
  - Clear separation: `config` package handles loading/merging/validation, distinct from application logic
  - Testability: All parsing, validation, merging logic is pure functions suitable for unit tests
  - Test coverage target >80% via table-driven tests for validators, integration tests for full lifecycle
  - Design patterns: Strategy pattern for format parsers (YAML/TOML), Observer pattern for file watching
  - Zero premature optimization: Simple merge logic, only optimize if profiling shows bottleneck
- **Status**: No concerns. Architecture supports testing and follows established Go patterns.

### Principle VIII: Complete Implementation - No Deferrals ✅ PASS
- **Assessment**: All 56 functional requirements are fully specified and implementable. No "TODO" or "Future work" items in spec. Clarifications resolved all ambiguities (sensitive data handling, hot-reload default, file size limits, format conflict handling, validation behavior). Every requirement has clear acceptance criteria in user stories.
- **Status**: No concerns. Specification is complete and ready for full implementation without deferrals.

### **GATE STATUS: ✅ ALL PASS** - Proceed to Phase 0 Research

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
internal/
├── config/                  # Configuration system (this feature)
│   ├── config.go           # Main Config struct, Load/Validate/Merge
│   ├── defaults.go         # Hardcoded default values
│   ├── parser.go           # Format detection, delegates to yaml/toml
│   ├── parser_yaml.go      # YAML parsing implementation
│   ├── parser_toml.go      # TOML parsing implementation
│   ├── validator.go        # Validation rules and constraint checking
│   ├── merger.go           # Multi-source merge with precedence
│   ├── env.go              # Environment variable parsing (LAZYNUGET_ prefix)
│   ├── encryption.go       # Encrypted value support (AES-256-GCM)
│   ├── keychain.go         # Platform keychain integration (via go-keyring)
│   ├── watcher.go          # File watching for hot-reload (via fsnotify)
│   ├── types.go            # ConfigSchema, ValidationError, EncryptedValue
│   ├── schema.go           # Schema definitions for all 50+ settings
│   └── paths.go            # Platform-specific config file path resolution
│
├── bootstrap/               # (Existing from spec 001)
│   └── app.go              # Integrates with config.Load() during Bootstrap()
│
├── platform/                # (Existing from spec 001 - used by config for paths)
│   ├── detect.go
│   └── types.go
│
└── logging/                 # (Existing from spec 001 - configured by config)
    └── logger.go

cmd/lazynuget/
├── main.go                  # (Existing) Calls config.Load() via bootstrap
└── encrypt.go               # NEW: `lazynuget encrypt-value` subcommand

tests/
├── integration/
│   ├── config_test.go      # Full lifecycle: load → merge → validate → apply
│   ├── config_hot_reload_test.go  # Hot-reload with file watcher
│   ├── config_encryption_test.go  # Encrypted values with keychain
│   └── config_platforms_test.go   # Cross-platform path handling
│
└── fixtures/
    └── configs/             # Sample YAML/TOML files for testing
        ├── valid.yml
        ├── invalid_syntax.yml
        ├── out_of_range.yml
        ├── encrypted.yml
        └── valid.toml
```

**Structure Decision**: Single project structure (Option 1). LazyNuGet is a monolithic TUI application with clear package separation. The `internal/config` package is self-contained and integrates with existing `bootstrap`, `platform`, and `logging` packages from spec 001-app-bootstrap. The configuration system follows Go conventions: internal packages for implementation, cmd for CLI entry points, tests for validation.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No constitutional violations detected. All principles pass.
