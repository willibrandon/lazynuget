# Tasks: Configuration Management System

**Input**: Design documents from `/specs/002-config-management/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Tests are included for all user stories as specified in the feature requirements.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

LazyNuGet is a single project with the following structure:
- `internal/` for implementation packages
- `cmd/` for CLI entry points
- `tests/integration/` for integration tests
- `tests/fixtures/configs/` for test configuration files

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and configuration package structure

- [X] T001 Create internal/config/ package directory structure per plan.md
- [X] T002 [P] Add gopkg.in/yaml.v3 dependency to go.mod
- [X] T003 [P] Add github.com/BurntSushi/toml dependency to go.mod
- [X] T004 [P] Add github.com/fsnotify/fsnotify dependency to go.mod
- [X] T005 [P] Add github.com/zalando/go-keyring dependency to go.mod
- [X] T006 Create tests/fixtures/configs/ directory for test configuration files
- [X] T007 Run go mod tidy to verify all dependencies resolve correctly

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core types, defaults, and platform utilities that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T008 [P] Define Config struct with all 50+ settings in internal/config/types.go per data-model.md entity #1
- [X] T009 [P] Define ColorScheme struct in internal/config/types.go per data-model.md entity #2
- [X] T010 [P] Define KeyBinding struct in internal/config/types.go per data-model.md entity #3
- [X] T011 [P] Define Timeouts struct in internal/config/types.go per data-model.md entity #4
- [X] T012 [P] Define LogRotation struct in internal/config/types.go per data-model.md entity #5
- [X] T013 [P] Define ValidationError struct with Error() method in internal/config/types.go per data-model.md entity #8
- [X] T014 [P] Define ConfigSource struct in internal/config/types.go per data-model.md entity #6
- [X] T015 [P] Define MergedConfig struct in internal/config/types.go per data-model.md entity #7
- [X] T016 [P] Define EncryptedValue struct in internal/config/types.go per data-model.md entity #9
- [X] T017 [P] Define SettingSchema and ConfigSchema structs in internal/config/types.go per data-model.md entity #10
- [X] T018 Implement GetDefaultConfig() function in internal/config/defaults.go returning Config with all default values per FR-001
- [X] T019 Implement getPlatformConfigPath() function in internal/config/paths.go for platform-specific paths per FR-006
- [X] T020 [P] Create ConfigLoader interface in internal/config/config.go per contracts/config_loader.md
- [X] T021 [P] Create Logger interface in internal/config/config.go per contracts/config_loader.md
- [X] T022 Implement schema definition for all 50+ settings in internal/config/schema.go per data-model.md entity #10

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - First-Time User Gets Working Defaults (Priority: P1) 🎯 MVP

**Goal**: Application starts successfully with sensible defaults when no config file exists, enabling immediate use without any setup

**Independent Test**: Install LazyNuGet on clean system with no config files, run it, verify all features work with default settings

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T023 [P] [US1] Create test for default config loading in tests/integration/config_defaults_test.go verifying GetDefaultConfig returns valid Config
- [X] T024 [P] [US1] Create test for missing config file scenario in tests/integration/config_defaults_test.go verifying app uses defaults when file doesn't exist
- [X] T025 [P] [US1] Create test fixture tests/fixtures/configs/empty.yml with empty content verifying fallback to defaults

### Implementation for User Story 1

- [X] T026 [US1] Implement configLoader struct implementing ConfigLoader interface in internal/config/config.go
- [X] T027 [US1] Implement Load() method handling missing config file case in internal/config/config.go per FR-001
- [X] T028 [US1] Implement GetDefaults() method returning default Config in internal/config/config.go per contract
- [X] T029 [US1] Implement basic validation for Config struct in internal/config/validator.go checking required field types
- [X] T030 [US1] Implement Validate() method with empty validation rules in internal/config/config.go per contract
- [X] T031 [US1] Add logging for "using default configuration" in internal/config/config.go when no file present
- [X] T032 [US1] Verify T023-T025 tests now pass with default config implementation

**Checkpoint**: At this point, LazyNuGet can start with defaults - minimal viable config system

---

## Phase 4: User Story 2 - Experienced User Customizes Settings via Config File (Priority: P2)

**Goal**: Users can create YAML or TOML config files with custom settings that override defaults on launch

**Independent Test**: Create config file with custom settings, launch LazyNuGet, verify custom settings are applied

### Tests for User Story 2

- [X] T033 [P] [US2] Create tests/fixtures/configs/valid.yml with sample YAML config covering all setting categories
- [X] T034 [P] [US2] Create tests/fixtures/configs/valid.toml with sample TOML config covering all setting categories
- [X] T035 [P] [US2] Create tests/fixtures/configs/invalid_syntax.yml with intentional YAML syntax error
- [X] T036 [P] [US2] Create tests/fixtures/configs/out_of_range.yml with semantic validation errors (maxConcurrentOps: 999)
- [X] T037 [P] [US2] Create tests/fixtures/configs/unknown_keys.yml with unsupported config keys
- [X] T038 [P] [US2] Create tests/fixtures/configs/both_formats/ directory with both .yml and .toml files
- [X] T039 [US2] Create integration test in tests/integration/config_file_test.go verifying YAML config file loads and overrides defaults
- [X] T040 [US2] Create integration test in tests/integration/config_file_test.go verifying TOML config file loads and overrides defaults
- [X] T041 [US2] Create integration test in tests/integration/config_file_test.go verifying syntax error blocks startup per FR-010
- [X] T042 [US2] Create integration test in tests/integration/config_file_test.go verifying semantic errors fallback to defaults per FR-012
- [X] T043 [US2] Create integration test in tests/integration/config_file_test.go verifying both formats present triggers error per FR-005

### Implementation for User Story 2

- [X] T044 [P] [US2] Implement YAML parser in internal/config/parser_yaml.go using gopkg.in/yaml.v3 per FR-003
- [X] T045 [P] [US2] Implement TOML parser in internal/config/parser_toml.go using github.com/BurntSushi/toml per FR-004
- [X] T046 [US2] Implement format detection in internal/config/parser.go checking for .yml/.yaml/.toml extensions
- [X] T047 [US2] Implement file size validation in internal/config/parser.go rejecting files >10MB per FR-009
- [X] T048 [US2] Implement check for multiple formats in internal/config/parser.go per FR-005
- [X] T049 [US2] Implement parseConfigFile() in internal/config/parser.go handling syntax errors with line numbers per FR-010
- [X] T050 [US2] Update Load() method in internal/config/config.go to detect and load config file if present
- [X] T051 [US2] Implement mergeConfigs() in internal/config/merger.go merging file config with defaults per FR-002
- [X] T052 [US2] Implement validation rules for all setting types in internal/config/validator.go per data-model.md validation rules
- [X] T053 [US2] Add constraint validators (range, enum, hexcolor, dateformat, regex) in internal/config/validator.go
- [X] T054 [US2] Update Validate() to collect semantic validation errors in internal/config/validator.go per FR-011
- [X] T055 [US2] Implement warning logs for semantic validation errors in internal/config/validator.go per FR-013
- [X] T056 [US2] Implement fallback to defaults for invalid settings in internal/config/validator.go per FR-012
- [X] T057 [US2] Implement keybinding conflict detection in internal/config/validator.go per FR-028
- [X] T058 [US2] Implement PrintConfig() method showing merged config with provenance in internal/config/config.go per contract
- [X] T059 [US2] Verify T033-T043 tests now pass with file-based config implementation

**Checkpoint**: At this point, User Story 1 AND 2 both work independently - users can customize via files

---

## Phase 5: User Story 3 - DevOps Engineer Overrides Settings via Environment Variables (Priority: P3)

**Goal**: Environment variables with LAZYNUGET_ prefix override config file settings for CI/CD and containerized deployments

**Independent Test**: Set environment variables for key settings, run LazyNuGet without config file, verify environment variables override defaults

### Tests for User Story 3

- [X] T060 [P] [US3] Create integration test in tests/integration/config_env_test.go verifying simple env var (LAZYNUGET_LOG_LEVEL=debug) overrides default
- [X] T061 [P] [US3] Create integration test in tests/integration/config_env_test.go verifying nested env var (LAZYNUGET_COLOR_SCHEME_BORDER=#FF0000) works per FR-051
- [X] T062 [P] [US3] Create integration test in tests/integration/config_env_test.go verifying env var overrides config file value per FR-002
- [X] T063 [P] [US3] Create integration test in tests/integration/config_env_test.go verifying invalid env var value triggers fallback to default per FR-012
- [X] T064 [P] [US3] Create integration test in tests/integration/config_env_test.go verifying type conversion for bool/int/duration env vars per FR-052

### Implementation for User Story 3

- [X] T065 [P] [US3] Implement parseEnvVars() in internal/config/env.go parsing all LAZYNUGET_ environment variables per FR-050
- [X] T066 [US3] Implement splitEnvVarPath() in internal/config/env.go converting LAZYNUGET_COLOR_SCHEME_BORDER to ["colorScheme", "border"] per FR-051
- [X] T067 [US3] Implement type conversion in internal/config/env.go for bool/int/duration/string per FR-052
- [X] T068 [US3] Implement case-insensitive matching in internal/config/env.go for env var names
- [X] T069 [US3] Update Load() method in internal/config/config.go to parse environment variables
- [X] T070 [US3] Update mergeConfigs() in internal/config/merger.go to merge env vars with precedence: env > file > defaults per FR-002
- [X] T071 [US3] Add logging for env var overrides in internal/config/config.go showing which settings were overridden
- [X] T072 [US3] Verify T060-T064 tests now pass with environment variable support

**Checkpoint**: All three user stories (defaults, file, env vars) work independently and in combination

---

## Phase 6: User Story 4 - User Temporarily Overrides Settings via CLI Flags (Priority: P4)

**Goal**: CLI flags provide highest precedence overrides for temporary settings changes without modifying config

**Independent Test**: Launch LazyNuGet with CLI flags overriding existing config, verify flags take precedence

### Tests for User Story 4

- [ ] T073 [P] [US4] Create integration test in tests/integration/config_cli_test.go verifying --log-level flag overrides all other sources per FR-054
- [ ] T074 [P] [US4] Create integration test in tests/integration/config_cli_test.go verifying --config flag specifies custom config file path per FR-053
- [ ] T075 [P] [US4] Create integration test in tests/integration/config_cli_test.go verifying --non-interactive flag works per FR-054
- [ ] T076 [P] [US4] Create integration test in tests/integration/config_cli_test.go verifying --no-color flag works per FR-054
- [ ] T077 [P] [US4] Create integration test in tests/integration/config_cli_test.go verifying --print-config outputs merged config per FR-055
- [ ] T078 [P] [US4] Create integration test in tests/integration/config_cli_test.go verifying --validate-config validates without starting app per FR-056

### Implementation for User Story 4

- [ ] T079 [US4] Update LoadOptions struct in internal/config/types.go to include CLIFlags field per contracts/config_loader.md
- [ ] T080 [US4] Update CLIFlags struct in internal/config/types.go with fields for --log-level, --non-interactive, --no-color per FR-054
- [ ] T081 [US4] Update Load() method in internal/config/config.go to accept LoadOptions with CLI flags
- [ ] T082 [US4] Update mergeConfigs() in internal/config/merger.go to merge CLI flags with highest precedence: cli > env > file > defaults per FR-002
- [ ] T083 [US4] Implement --config flag handling in internal/config/config.go per FR-007
- [ ] T084 [US4] Implement LAZYNUGET_CONFIG env var handling in internal/config/config.go per FR-008
- [ ] T085 [US4] Implement --print-config functionality in internal/config/config.go calling PrintConfig() and exiting per FR-055
- [ ] T086 [US4] Implement --validate-config functionality in internal/config/config.go calling Validate() and exiting per FR-056
- [ ] T087 [US4] Update bootstrap integration in internal/bootstrap/app.go to pass CLI flags to config.Load()
- [ ] T088 [US4] Verify T073-T078 tests now pass with CLI flag support

**Checkpoint**: Full four-layer precedence system working (defaults < file < env < cli)

---

## Phase 7: User Story 5 - Power User Modifies Config and Sees Changes Immediately (Priority: P5)

**Goal**: When hot-reload is explicitly enabled, config file changes are detected and applied within 3 seconds without restart

**Independent Test**: Start LazyNuGet with hotReload: true, modify config file, verify changes apply automatically within 3 seconds

### Tests for User Story 5

- [ ] T089 [P] [US5] Create integration test in tests/integration/config_hot_reload_test.go verifying hot-reload disabled by default per FR-043
- [ ] T090 [P] [US5] Create integration test in tests/integration/config_hot_reload_test.go verifying config change detected within 3 seconds when hot-reload enabled per FR-045
- [ ] T091 [P] [US5] Create integration test in tests/integration/config_hot_reload_test.go verifying invalid reload keeps previous config per FR-047
- [ ] T092 [P] [US5] Create integration test in tests/integration/config_hot_reload_test.go verifying hot-reload success notification per FR-048
- [ ] T093 [P] [US5] Create integration test in tests/integration/config_hot_reload_test.go verifying hot-reload failure notification per FR-048
- [ ] T094 [P] [US5] Create integration test in tests/integration/config_hot_reload_test.go verifying config file deletion triggers fallback to defaults per edge cases
- [ ] T095 [P] [US5] Create integration test in tests/integration/config_hot_reload_test.go verifying rapid successive writes debounced properly

### Implementation for User Story 5

- [ ] T096 [P] [US5] Define ConfigWatcher interface in internal/config/watcher.go per contracts/watcher.md
- [ ] T097 [P] [US5] Define WatchOptions struct in internal/config/watcher.go per contracts/watcher.md
- [ ] T098 [P] [US5] Define ConfigChangeEvent and ConfigChangeType in internal/config/watcher.go per contracts/watcher.md
- [ ] T099 [US5] Implement configWatcher struct in internal/config/watcher.go using fsnotify
- [ ] T100 [US5] Implement Watch() method in internal/config/watcher.go starting background goroutine per contract
- [ ] T101 [US5] Implement file event handling in internal/config/watcher.go for Write/Modify/Delete/Rename per contract
- [ ] T102 [US5] Implement 100ms debounce logic in internal/config/watcher.go per FR-044 and contract
- [ ] T103 [US5] Implement reload validation in internal/config/watcher.go calling Load() and Validate() per FR-046
- [ ] T104 [US5] Implement callback invocation in internal/config/watcher.go for OnReload/OnError/OnFileDeleted per contract
- [ ] T105 [US5] Implement Stop() method in internal/config/watcher.go releasing resources per contract
- [ ] T106 [US5] Implement hot-reloadable flag checking in internal/config/schema.go per FR-049
- [ ] T107 [US5] Update schema definitions in internal/config/schema.go marking which settings are hot-reloadable
- [ ] T108 [US5] Add integration with application in internal/bootstrap/app.go to start watcher when hotReload: true
- [ ] T109 [US5] Verify T089-T095 tests now pass with hot-reload implementation

**Checkpoint**: All five user stories fully functional - complete configuration system

---

## Phase 8: Encryption Support (Cross-Cutting for US2-US5)

**Purpose**: Add encrypted value support that works across all config sources (file, env, cli)

### Tests for Encryption

- [ ] T110 [P] Create tests/fixtures/configs/encrypted.yml with sample encrypted value
- [ ] T111 [P] Create integration test in tests/integration/config_encryption_test.go verifying encrypted value decrypts correctly
- [ ] T112 [P] Create integration test in tests/integration/config_encryption_test.go verifying decryption failure falls back to default per FR-012
- [ ] T113 [P] Create integration test in tests/integration/config_encryption_test.go verifying keychain unavailable falls back to env var
- [ ] T114 [P] Create integration test in tests/integration/config_encryption_test.go verifying encrypted values never logged in plain text per FR-018

### Implementation for Encryption

- [ ] T115 [P] Define Encryptor interface in internal/config/encryption.go per contracts/encryption.md
- [ ] T116 [P] Define KeychainManager interface in internal/config/keychain.go per contracts/encryption.md
- [ ] T117 [P] Define KeyDerivation interface in internal/config/encryption.go per contracts/encryption.md
- [ ] T118 Implement encryptor struct in internal/config/encryption.go using crypto/aes and crypto/cipher (AES-256-GCM)
- [ ] T119 Implement Encrypt() method in internal/config/encryption.go per FR-016 and contract
- [ ] T120 Implement Decrypt() method in internal/config/encryption.go per FR-016 and contract
- [ ] T121 Implement EncryptToString() method in internal/config/encryption.go per FR-019 and contract
- [ ] T122 Implement DecryptFromString() method in internal/config/encryption.go per contract
- [ ] T123 Implement keychainManager struct in internal/config/keychain.go using github.com/zalando/go-keyring per FR-017
- [ ] T124 Implement Store() method in internal/config/keychain.go per contract
- [ ] T125 Implement Retrieve() method in internal/config/keychain.go with env var fallback per contract
- [ ] T126 Implement Delete() method in internal/config/keychain.go per contract
- [ ] T127 Implement List() method in internal/config/keychain.go per contract
- [ ] T128 Implement IsAvailable() method in internal/config/keychain.go per contract
- [ ] T129 Implement keyDerivation using PBKDF2 in internal/config/encryption.go per contract
- [ ] T130 Implement custom YAML unmarshaling for !encrypted tag in internal/config/parser_yaml.go
- [ ] T131 Update Load() to decrypt EncryptedValue instances in internal/config/config.go
- [ ] T132 Add warning logging for decryption failures in internal/config/config.go per FR-018
- [ ] T133 Create cmd/lazynuget/encrypt.go implementing `lazynuget encrypt-value` subcommand per FR-019
- [ ] T134 Verify T110-T114 tests now pass with encryption implementation

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T135 [P] Add comprehensive godoc comments to all public types in internal/config/types.go
- [ ] T136 [P] Add godoc comments to all public functions in internal/config/config.go
- [ ] T137 [P] Create example config files in tests/fixtures/configs/example.yml and example.toml
- [ ] T138 [P] Run go vet on internal/config/ package and fix any issues
- [ ] T139 [P] Run go fmt on internal/config/ package
- [ ] T140 [P] Run golangci-lint on internal/config/ package and address warnings
- [ ] T141 Add unit tests for mergeConfigs() in internal/config/merger_test.go with table-driven tests
- [ ] T142 Add unit tests for validation rules in internal/config/validator_test.go with table-driven tests
- [ ] T143 Add unit tests for env var parsing in internal/config/env_test.go with table-driven tests
- [ ] T144 Add unit tests for YAML parser in internal/config/parser_yaml_test.go
- [ ] T145 Add unit tests for TOML parser in internal/config/parser_toml_test.go
- [ ] T146 Run quickstart.md examples manually to verify all code samples work
- [ ] T147 Update quickstart.md with any corrections from manual validation
- [ ] T148 Run full integration test suite and verify >80% code coverage target
- [ ] T149 Test on Windows, macOS, and Linux platforms verifying cross-platform behavior per Constitution Principle IV
- [ ] T150 Verify startup config load completes in <500ms per performance goal
- [ ] T151 Verify hot-reload latency <3s per FR-045
- [ ] T152 Verify config file parsing <100ms for typical files per SC-010

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies - can start immediately
- **Phase 2 (Foundational)**: Depends on Phase 1 completion - BLOCKS all user stories
- **Phase 3 (US1)**: Depends on Phase 2 completion
- **Phase 4 (US2)**: Depends on Phase 2 completion, can run in parallel with US1 after Phase 2 done
- **Phase 5 (US3)**: Depends on Phase 2 completion, can run in parallel with US1/US2 after Phase 2 done
- **Phase 6 (US4)**: Depends on Phase 2 completion, can run in parallel with US1/US2/US3 after Phase 2 done
- **Phase 7 (US5)**: Depends on Phase 4 completion (needs file loading), can run in parallel with US3/US4
- **Phase 8 (Encryption)**: Depends on Phase 4 completion (needs parsers), can run in parallel with US3/US4/US5
- **Phase 9 (Polish)**: Depends on all phases above

### User Story Dependencies

- **User Story 1 (P1)**: Foundation only - no dependencies on other stories
- **User Story 2 (P2)**: Foundation only - no dependencies on other stories (can run parallel with US1)
- **User Story 3 (P3)**: Foundation only - no dependencies on other stories (can run parallel with US1/US2)
- **User Story 4 (P4)**: Foundation only - no dependencies on other stories (can run parallel with US1/US2/US3)
- **User Story 5 (P5)**: Depends on US2 (file loading) - but independently testable

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Types before implementation
- Core implementation before validation
- Validation before integration
- Story complete before moving to next priority

### Parallel Opportunities

**Phase 1 (Setup)**: Tasks T002-T005 can run in parallel (all independent dependency additions)

**Phase 2 (Foundational)**: Tasks T008-T017 can run in parallel (all independent type definitions), T020-T021 can run in parallel

**After Phase 2 Completes**:
- All user stories (US1-US4) can start in parallel if team capacity allows
- US5 should wait for US2 completion but can then run parallel with US3/US4

**Within User Stories**:
- All test fixture creation can run in parallel within a story
- All test file creation can run in parallel within a story
- All parser implementations can run in parallel (T044-T045)
- All interface definitions can run in parallel

---

## Parallel Example: Phase 2 Foundational Types

```bash
# Launch all type definitions together (T008-T017):
Task: "Define Config struct in internal/config/types.go"
Task: "Define ColorScheme struct in internal/config/types.go"
Task: "Define KeyBinding struct in internal/config/types.go"
Task: "Define Timeouts struct in internal/config/types.go"
Task: "Define LogRotation struct in internal/config/types.go"
Task: "Define ValidationError struct in internal/config/types.go"
Task: "Define ConfigSource struct in internal/config/types.go"
Task: "Define MergedConfig struct in internal/config/types.go"
Task: "Define EncryptedValue struct in internal/config/types.go"
Task: "Define SettingSchema struct in internal/config/types.go"
```

---

## Parallel Example: User Story 2 Tests

```bash
# Launch all test fixtures together (T033-T038):
Task: "Create tests/fixtures/configs/valid.yml"
Task: "Create tests/fixtures/configs/valid.toml"
Task: "Create tests/fixtures/configs/invalid_syntax.yml"
Task: "Create tests/fixtures/configs/out_of_range.yml"
Task: "Create tests/fixtures/configs/unknown_keys.yml"
Task: "Create tests/fixtures/configs/both_formats/ directory"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T007)
2. Complete Phase 2: Foundational (T008-T022) - CRITICAL
3. Complete Phase 3: User Story 1 (T023-T032)
4. **STOP and VALIDATE**: Test defaults-only scenario independently
5. You now have a minimal config system that works!

### Incremental Delivery

1. Setup + Foundational (Phases 1-2) → Foundation ready
2. Add User Story 1 (Phase 3) → Test independently → **MVP deployed!**
3. Add User Story 2 (Phase 4) → Test independently → File config support
4. Add User Story 3 (Phase 5) → Test independently → Env var support
5. Add User Story 4 (Phase 6) → Test independently → Full precedence system
6. Add User Story 5 (Phase 7) → Test independently → Hot-reload for power users
7. Add Encryption (Phase 8) → Test independently → Secure credential storage
8. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers after Phase 2 completes:

- **Developer A**: User Story 1 (T023-T032) - Defaults
- **Developer B**: User Story 2 (T033-T059) - File config
- **Developer C**: User Story 3 (T060-T072) - Env vars
- **Developer D**: User Story 4 (T073-T088) - CLI flags

Once US2 complete:
- **Developer E**: User Story 5 (T089-T109) - Hot-reload
- **Developer F**: Encryption (T110-T134) - Encrypted values

Stories complete and integrate independently with minimal merge conflicts.

---

## Notes

- **[P] tasks** = different files, no dependencies, safe to run in parallel
- **[Story] label** maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail (Red) before implementing (Green)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- All 152 tasks follow strict checklist format: `- [ ] [ID] [P?] [Story?] Description with file path`
- Performance targets: <500ms startup load, <100ms parsing, <3s hot-reload
- Cross-platform testing required per Constitution Principle IV
- No deferrals allowed per Constitution Principle VIII
