# Feature Specification: Configuration Management System

**Feature Branch**: `002-config-management`
**Created**: 2025-11-02
**Status**: Draft
**Input**: User description: "Implement the configuration management system that loads, merges, and validates configuration from multiple sources. Support hierarchical configuration with defaults, user config files (YAML/TOML), environment variables, and CLI flags. Configuration must include UI settings, keybindings, color schemes, performance tuning, and dotnet CLI paths. Must support hot-reload for user config changes."

## Clarifications

### Session 2025-11-02

- Q: How should sensitive values (API keys, tokens, credentials) be handled in configuration files? → A: Support encrypted values with external key management (encrypt in config, decrypt at runtime using key from env var or system keychain)
- Q: What should be the default behavior for hot-reload? → A: Disabled by default (opt-in)
- Q: Should the system enforce a maximum config file size? → A: Yes, enforce reasonable limit (10 MB)
- Q: What should be the exact behavior when both YAML and TOML config files exist in the config directory? → A: Error and require user to remove one
- Q: Should semantic validation errors (values out of range) block application startup or fall back to defaults? → A: Fall back to defaults with warning

## User Scenarios & Testing *(mandatory)*

### User Story 1 - First-Time User Gets Working Defaults (Priority: P1)

A new user downloads LazyNuGet and runs it immediately without any configuration. The application starts with sensible defaults for all settings, allowing them to begin managing NuGet packages immediately.

**Why this priority**: This is the foundation of the entire configuration system. Without working defaults, the application cannot function. This represents the absolute minimum viable product - users can use LazyNuGet without any configuration effort.

**Independent Test**: Install LazyNuGet on a clean system with no config files, run it, and verify all features work with default settings. Delivers immediate value by allowing package management without setup.

**Acceptance Scenarios**:

1. **Given** LazyNuGet is installed on a system with no config files, **When** user runs `lazynuget`, **Then** application starts successfully with default UI theme, default keybindings, and default performance settings
2. **Given** no custom dotnet CLI path is configured, **When** application initializes, **Then** it automatically detects and uses the system dotnet CLI from PATH
3. **Given** default configuration is active, **When** user interacts with the UI, **Then** all keyboard shortcuts work as documented in the default keybinding scheme

---

### User Story 2 - Experienced User Customizes Settings via Config File (Priority: P2)

An experienced user wants to customize LazyNuGet to match their preferences (dark theme, vim-style keybindings, increased concurrency). They create a config file in the standard location with their preferences, and LazyNuGet respects these settings on next launch.

**Why this priority**: This is the primary way power users will customize the application. Once defaults work (P1), enabling persistent customization is the next most valuable feature. This can be developed and tested independently by verifying config file parsing and application.

**Independent Test**: Create a config file with custom settings, launch LazyNuGet, and verify all custom settings are applied. Delivers value by enabling personalization without requiring code changes.

**Acceptance Scenarios**:

1. **Given** user creates a YAML config file at the platform-specific config location, **When** user launches LazyNuGet, **Then** all settings from the config file override defaults
2. **Given** config file contains invalid YAML syntax, **When** application loads configuration, **Then** user sees a clear error message indicating the syntax error location and the application uses defaults
3. **Given** config file specifies a custom color scheme, **When** UI renders, **Then** all UI elements use the custom colors
4. **Given** config file sets `maxConcurrentOps: 8`, **When** application performs operations, **Then** up to 8 operations run concurrently
5. **Given** both YAML and TOML config files exist in config directory, **When** application loads configuration, **Then** application fails with clear error message requiring user to remove one file

---

### User Story 3 - DevOps Engineer Overrides Settings via Environment Variables (Priority: P3)

A DevOps engineer deploys LazyNuGet in CI/CD pipelines where config files are inconvenient. They use environment variables to override specific settings (log level, dotnet CLI path, non-interactive mode) without modifying any files.

**Why this priority**: This enables automation and CI/CD use cases, which are important but not required for basic usage. Can be developed after file-based config (P2) by adding environment variable parsing layer.

**Independent Test**: Set environment variables for key settings, run LazyNuGet without config file, and verify environment variables override defaults. Delivers value by enabling scriptable/automated deployments.

**Acceptance Scenarios**:

1. **Given** `LAZYNUGET_LOG_LEVEL=debug` is set, **When** application initializes, **Then** logging is set to debug level regardless of config file or default
2. **Given** `LAZYNUGET_DOTNET_PATH=/custom/dotnet` is set, **When** application needs dotnet CLI, **Then** it uses the custom path
3. **Given** both config file and environment variable specify the same setting, **When** application merges configuration, **Then** environment variable takes precedence

---

### User Story 4 - User Temporarily Overrides Settings via CLI Flags (Priority: P4)

A user wants to temporarily try a different setting (e.g., increase log verbosity for debugging, use different config file, disable colors for piping output) without modifying their config file. They use CLI flags to override specific settings for just this execution.

**Why this priority**: This provides maximum flexibility for troubleshooting and experimentation. Can be developed after environment variable support (P3) since it uses similar override logic.

**Independent Test**: Launch LazyNuGet with CLI flags overriding existing config, verify flags take precedence. Delivers value by enabling temporary changes without config modification.

**Acceptance Scenarios**:

1. **Given** user runs `lazynuget --log-level=trace`, **When** application initializes, **Then** log level is trace regardless of env vars, config file, or defaults
2. **Given** user runs `lazynuget --config=/custom/config.yml`, **When** application loads configuration, **Then** it uses the specified config file instead of default location
3. **Given** user runs `lazynuget --no-color`, **When** UI renders, **Then** all output is monochrome regardless of color scheme settings
4. **Given** multiple override sources are active (default < config file < env var < CLI flag), **When** application merges configuration, **Then** CLI flags have highest precedence

---

### User Story 5 - Power User Modifies Config and Sees Changes Immediately (Priority: P5)

A power user is experimenting with different color schemes and keybindings. They edit their config file while LazyNuGet is running, and the changes take effect within seconds without restarting the application.

**Why this priority**: This is a quality-of-life improvement for advanced users. It's valuable but not essential - users can restart the application. Should be developed last after all other config features are stable.

**Independent Test**: Start LazyNuGet, modify config file, and verify changes are applied automatically within a few seconds. Delivers value by eliminating restart friction during customization.

**Acceptance Scenarios**:

1. **Given** LazyNuGet is running with default theme, **When** user changes theme in config file and saves, **Then** UI updates to new theme within 3 seconds without restart
2. **Given** LazyNuGet is running, **When** user changes keybindings in config file and saves, **Then** new keybindings become active immediately
3. **Given** config file watcher detects a change, **When** the new config has validation errors, **Then** user sees an error notification and application continues with previous valid configuration
4. **Given** user saves config file multiple times in quick succession, **When** file watcher processes changes, **Then** only the final valid configuration is applied (no flickering between states)
5. **Given** hot-reload is triggered, **When** certain settings cannot be changed without restart (e.g., initial window size), **Then** user sees a notification indicating which settings require restart

---

### Edge Cases

- What happens when config file exists but is empty? (Falls back to defaults, startup succeeds)
- What happens when config file exceeds 10 MB size limit? (Rejects file, fails to start with clear error)
- What happens when config file contains unknown/unsupported keys? (Warns about unknown keys, ignores them, uses known settings)
- How does system handle config file that is not readable due to permissions? (Logs error, falls back to defaults, continues startup)
- What happens when multiple config file formats (both YAML and TOML) exist in config directory? (Fails to start with error requiring user to remove one)
- How does system handle environment variables with invalid values (e.g., `LAZYNUGET_MAX_CONCURRENT_OPS=abc`)? (Warns, uses default for that setting, continues startup)
- What happens when CLI flag has invalid syntax (e.g., `--max-concurrent-ops=999999`)? (Warns, uses default for that setting, continues startup)
- How does system handle partial config files (only some sections present)? (Uses provided settings, defaults for missing sections, continues startup)
- What happens when config file is modified while application is reading it? (Completes read with whatever data was accessible, may warn about inconsistency)
- What happens when user deletes config file while application is running with hot-reload enabled? (Falls back to defaults, notifies user of config file removal)
- How does system handle dotnet CLI path that points to non-existent or non-executable file? (Validation warning during startup, falls back to auto-detection from PATH)
- What happens when keybinding config creates conflicts (two actions bound to same key)? (Warns about conflict, uses first binding encountered, ignores conflicting bindings)
- How does system handle color values in invalid formats (e.g., `#GGGGGG`, `rgb(300,300,300)`)? (Warns about invalid color, uses default color for that UI element)
- What happens when encrypted config value cannot be decrypted (missing key, wrong key, corrupted ciphertext)? (Warns about decryption failure, uses default value for that setting)
- How does system handle encrypted values when secure storage (keychain/credential manager) is unavailable? (Falls back to environment variable key source, warns if neither available)

## Requirements *(mandatory)*

### Functional Requirements

#### Configuration Loading & Merging

- **FR-001**: System MUST load configuration from four sources in order: hardcoded defaults, user config file, environment variables, CLI flags
- **FR-002**: System MUST merge configuration using hierarchical precedence: CLI flags override environment variables, which override config file, which overrides defaults
- **FR-003**: System MUST support YAML format for configuration files
- **FR-004**: System MUST support TOML format for configuration files
- **FR-005**: System MUST fail with clear error message if both YAML and TOML config files exist in the same config directory, requiring user to remove one
- **FR-006**: System MUST use platform-specific default config file locations (macOS: `~/Library/Application Support/lazynuget/`, Linux: `~/.config/lazynuget/`, Windows: `%APPDATA%\lazynuget/`)
- **FR-007**: System MUST allow users to specify custom config file location via `--config` CLI flag
- **FR-008**: System MUST allow users to specify custom config file location via `LAZYNUGET_CONFIG` environment variable

#### Configuration Validation

- **FR-009**: System MUST reject configuration files larger than 10 MB with clear error message to prevent resource exhaustion and fail to start
- **FR-010**: System MUST reject config files with syntax errors (invalid YAML/TOML) and display parse error with line number, failing to start
- **FR-011**: System MUST validate all configuration values against defined constraints (e.g., positive integers for concurrency, valid hex codes for colors, range limits)
- **FR-012**: System MUST use defaults for any settings with semantic validation errors (out of range, wrong type, invalid format) rather than failing to start
- **FR-013**: System MUST log all semantic validation failures with warning level, clearly indicating which config key has invalid value, what valid values are, and what default is being used
- **FR-014**: System MUST provide clear error messages for all validation failures, whether they block startup (syntax errors, file size) or fall back to defaults (semantic errors)

#### Security & Sensitive Data

- **FR-015**: System MUST support encrypted values in configuration files for sensitive data (API keys, tokens, credentials)
- **FR-016**: System MUST decrypt encrypted configuration values at runtime using decryption keys from environment variables or system keychain
- **FR-017**: System MUST support platform-specific secure storage integration (macOS Keychain, Windows Credential Manager, Linux Secret Service API/pass)
- **FR-018**: System MUST never log decrypted sensitive values in plain text
- **FR-019**: System MUST provide a utility command to encrypt sensitive values for storage in config files (e.g., `lazynuget encrypt-value`)

#### UI Settings

- **FR-020**: Configuration MUST include `theme` setting with predefined theme names (e.g., "default", "dark", "light", "solarized")
- **FR-021**: Configuration MUST include `colorScheme` section allowing customization of individual UI element colors (borders, highlights, text, backgrounds)
- **FR-022**: Configuration MUST include `compactMode` boolean to toggle between normal and compact UI layouts
- **FR-023**: Configuration MUST include `showHints` boolean to toggle keyboard hint overlays
- **FR-024**: Configuration MUST include `showLineNumbers` boolean for list views
- **FR-025**: Configuration MUST include `dateFormat` string for timestamp display (e.g., "2006-01-02", "01/02/2006")

#### Keybindings

- **FR-026**: Configuration MUST include `keybindings` section mapping actions to key combinations
- **FR-027**: System MUST support standard key representations (e.g., `Ctrl+C`, `Alt+Enter`, `F5`, `Shift+Tab`)
- **FR-028**: System MUST detect and warn about conflicting keybindings (same key mapped to multiple actions in same context)
- **FR-029**: Configuration MUST allow users to define keybinding profiles (e.g., "default", "vim", "emacs") with quick switching
- **FR-030**: System MUST provide a way to disable individual keybindings by setting them to empty or null value

#### Performance Settings

- **FR-031**: Configuration MUST include `maxConcurrentOps` integer setting for maximum parallel operations (default: 4, range: 1-16)
- **FR-032**: Configuration MUST include `cacheSize` setting for in-memory cache limits (in megabytes)
- **FR-033**: Configuration MUST include `refreshInterval` duration for automatic package list refresh (e.g., "5m", "1h", or "0" to disable)
- **FR-034**: Configuration MUST include `timeouts` section with separate timeouts for different operation types (network requests, dotnet CLI calls)

#### Dotnet CLI Integration

- **FR-035**: Configuration MUST include `dotnetPath` setting to specify custom dotnet CLI executable path
- **FR-036**: System MUST auto-detect dotnet CLI from PATH if `dotnetPath` is not specified
- **FR-037**: Configuration MUST include `dotnetVerbosity` setting to control dotnet CLI output verbosity (e.g., "quiet", "minimal", "normal", "detailed")
- **FR-038**: System MUST validate that `dotnetPath` points to valid, executable dotnet CLI on application startup

#### Logging Configuration

- **FR-039**: Configuration MUST include `logLevel` setting (debug, info, warn, error)
- **FR-040**: Configuration MUST include `logDir` setting for log file location
- **FR-041**: Configuration MUST include `logFormat` setting (text, json)
- **FR-042**: Configuration MUST include `logRotation` settings (max size, max age, max backups)

#### Hot-Reload

- **FR-043**: Hot-reload MUST be disabled by default and require explicit opt-in via `hotReload: true` config setting or `LAZYNUGET_HOT_RELOAD=true` environment variable
- **FR-044**: System MUST watch user config file for changes when hot-reload is explicitly enabled
- **FR-045**: System MUST reload and reapply configuration within 3 seconds of config file modification when hot-reload is enabled
- **FR-046**: System MUST validate reloaded configuration before applying changes
- **FR-047**: System MUST continue using previous valid configuration if reloaded config has validation errors
- **FR-048**: System MUST notify user when config hot-reload succeeds or fails
- **FR-049**: System MUST identify which settings cannot be hot-reloaded (require restart) and notify user when these are changed

#### Environment Variable Support

- **FR-050**: System MUST support environment variable overrides using `LAZYNUGET_` prefix (e.g., `LAZYNUGET_LOG_LEVEL=debug`)
- **FR-051**: System MUST support nested config keys via underscore notation (e.g., `LAZYNUGET_COLOR_SCHEME_BORDER=#FF0000`)
- **FR-052**: System MUST parse environment variable values according to expected type (boolean, integer, duration, string)

#### CLI Flag Support

- **FR-053**: System MUST support `--config` flag to specify config file path
- **FR-054**: System MUST support common setting overrides via CLI flags (`--log-level`, `--non-interactive`, `--no-color`)
- **FR-055**: System MUST support `--print-config` flag to output merged configuration and exit (useful for debugging)
- **FR-056**: System MUST support `--validate-config` flag to validate configuration without starting application

### Key Entities

- **ConfigSource**: Represents one of the four configuration sources (defaults, file, environment variables, CLI flags) with its precedence level and raw values
- **MergedConfig**: The final configuration after merging all sources according to precedence rules, containing validated and type-converted values for all settings
- **ConfigSchema**: The definition of all valid configuration keys, their types, constraints, default values, and whether they support hot-reload
- **ValidationError**: Details about a configuration validation failure, including the config key, invalid value, constraint that failed, and suggested fix
- **ThemeDefinition**: A complete color scheme with colors for all UI elements (borders, text, backgrounds, highlights, status indicators)
- **KeybindingProfile**: A named collection of keybindings mapping actions to key combinations, supporting multiple profiles per user
- **EncryptedValue**: An encrypted configuration value stored in config file with metadata (encryption algorithm, key identifier) that can be decrypted at runtime using keys from environment variables or system secure storage

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can launch LazyNuGet without any configuration and complete all core workflows using default settings
- **SC-002**: Users can customize any setting via config file and see changes applied on next launch within 500 milliseconds
- **SC-003**: Users can override any config file setting via environment variable or CLI flag and see override take effect
- **SC-004**: When hot-reload is enabled, config file changes are applied within 3 seconds without restarting the application
- **SC-005**: 95% of configuration validation errors provide actionable error messages that allow users to fix the issue without consulting documentation
- **SC-006**: Configuration system supports at least 50 distinct settings across UI, keybindings, performance, and dotnet integration categories
- **SC-007**: Users can validate their configuration without starting the application by running `lazynuget --validate-config`
- **SC-008**: Configuration precedence (CLI > env > file > default) is consistent and predictable for all 50+ settings
- **SC-009**: System handles invalid config files gracefully without crashing, logging the error and falling back to defaults
- **SC-010**: Configuration file parsing and merging completes within 100 milliseconds for typical config files (<100 KB)

## Assumptions

- Users are familiar with YAML or TOML syntax (common configuration formats in developer tools)
- Default keybinding scheme will be similar to standard terminal application conventions (not vim/emacs style)
- Color scheme customization uses standard hex color codes (#RRGGBB format)
- Hot-reload is disabled by default for safety and predictability; users must explicitly enable it to get automatic config reloading during runtime
- Duration strings follow Go duration format (e.g., "5s", "10m", "1h")
- Configuration schema will be documented in user manual and/or example config files
- TOML support is provided for users who prefer explicit typing and less indentation
- Users will choose either YAML or TOML format for their config file, not both; system will error if both formats exist simultaneously to prevent ambiguity
- Platform-specific config locations follow each OS's standard conventions (XDG on Linux, Application Support on macOS, AppData on Windows)
- Config file watcher uses efficient OS-level file system notifications (inotify, FSEvents, ReadDirectoryChangesW) rather than polling
- Encrypted configuration values use industry-standard encryption algorithms (AES-256-GCM) with keys stored separately from config files
