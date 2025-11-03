# Tasks: Application Bootstrap and Lifecycle Management

**Input**: Design documents from `/specs/001-app-bootstrap/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Integration tests are included as this is core infrastructure requiring validation. Unit tests are co-located with implementation files.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

LazyNuGet uses single project structure with Go conventions:
- `cmd/lazynuget/` - Entry point
- `internal/bootstrap/` - Bootstrap implementation
- `internal/config/`, `internal/logging/`, `internal/platform/` - Dependencies (minimal stubs for bootstrap)
- `tests/integration/` - Integration tests
- `tests/fixtures/` - Test data

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic Go structure

- [X] T001 Initialize Go module with `go mod init github.com/yourusername/lazynuget` in repository root
- [X] T002 Create directory structure: `cmd/lazynuget/`, `internal/bootstrap/`, `internal/config/`, `internal/logging/`, `internal/platform/`, `tests/integration/`, `tests/fixtures/configs/`
- [X] T003 [P] Create `.gitignore` with Go patterns (binaries, vendor/, coverage files)
- [X] T004 [P] Create `Makefile` with build targets (build, build-dev, test, test-int, clean) per quickstart.md
- [X] T005 [P] Create `go.work` file if using Go workspaces (optional)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Minimal dependency stubs that MUST exist before bootstrap implementation

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T006 Create minimal `internal/config/config.go` stub with `AppConfig` struct and `Load()` function returning defaults
- [X] T007 [P] Create minimal `internal/logging/logger.go` stub with `Logger` interface and `New()` function returning no-op logger
- [X] T008 [P] Create minimal `internal/platform/detect.go` stub with `Platform` interface and `New()` function returning platform info
- [X] T009 Create `internal/bootstrap/types.go` with `VersionInfo` struct (version, commit, date fields)

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Quick Launch and Version Check (Priority: P1) 🎯 MVP

**Goal**: Enable basic application startup with `lazynuget`, `lazynuget --version`, and `lazynuget --help` commands

**Independent Test**: Run `./lazynuget`, `./lazynuget --version`, `./lazynuget --help` and verify appropriate output and exit codes

### Implementation for User Story 1

- [X] T010 [P] [US1] Create `internal/bootstrap/flags.go` with CLI flag definitions (--version, --help, --config, --log-level, --non-interactive) using Go flag package
- [X] T011 [P] [US1] Create `internal/bootstrap/version.go` with `ShowVersion()` function that formats and prints VersionInfo
- [X] T012 [US1] Create `internal/bootstrap/app.go` with `App` struct (config, logger, platform, lifecycle, gui, ctx, cancel, version, startTime fields) and `NewApp(version, commit, date string) (*App, error)` constructor
- [X] T013 [US1] Implement `App.ParseFlags()` method in `internal/bootstrap/flags.go` that parses CLI args and returns early for --version/--help
- [X] T014 [US1] Implement `App.Bootstrap()` method in `internal/bootstrap/app.go` that initializes config, logger, platform in order (with phase tracking for error context)
- [X] T015 [US1] Create `cmd/lazynuget/main.go` with main() function: call NewApp(), ParseFlags(), Bootstrap(), handle --version/--help, exit with appropriate codes
- [X] T016 [US1] Add Layer 1 panic recovery in `cmd/lazynuget/main.go` main() function (defer/recover with exit code 2)
- [X] T017 [US1] Add Layer 2 panic recovery in `App.Bootstrap()` method (defer/recover with phase logging, re-panic)
- [X] T018 [US1] Add help text generation in `internal/bootstrap/flags.go` that documents all flags with descriptions

### Integration Tests for User Story 1

- [X] T019 [P] [US1] Create `tests/integration/version_test.go` with TestVersionFlag that runs `./lazynuget --version` and validates output format
- [X] T020 [P] [US1] Create `tests/integration/help_test.go` with TestHelpFlag that runs `./lazynuget --help` and validates all flags are documented
- [X] T021 [P] [US1] Create `tests/integration/startup_test.go` with TestBasicStartup that validates app initializes without errors (stub, will be enhanced in US4)

**Checkpoint**: At this point, User Story 1 should be fully functional - `lazynuget --version` and `lazynuget --help` work correctly

---

## Phase 4: User Story 2 - Graceful Shutdown on User Exit (Priority: P1)

**Goal**: Enable graceful application shutdown via signals (SIGINT, SIGTERM) with resource cleanup

**Independent Test**: Start application (stub mode), send SIGINT, verify exit code 0 and <1s shutdown time

### Implementation for User Story 2

- [X] T022 [P] [US2] Create `internal/lifecycle` package with `Manager` struct (state, stateMutex, shutdownTimeout, shutdownHandlers fields)
- [X] T023 [P] [US2] Create `internal/lifecycle/signals.go` with `SignalHandler` struct and `WaitForShutdownSignal()` method using signal.Notify
- [X] T024 [US2] Implement `Manager.SetState()` method that transitions states with validation (Uninitialized → Initializing → Running)
- [X] T025 [US2] Implement `Manager.Shutdown(ctx)` method that transitions Running → ShuttingDown → ShutdownComplete, runs shutdown handlers, enforces timeout
- [X] T026 [US2] Implement `Manager.GetState()` method (thread-safe read using RWMutex)
- [X] T027 [US2] Implement `Manager.RegisterShutdownHandler(handler)` method for registering cleanup functions with priority
- [X] T028 [US2] Implement `SignalHandler.WaitForShutdownSignal()` method that blocks until SIGINT/SIGTERM received
- [X] T029 [US2] Implement errgroup pattern in `internal/lifecycle/errgroup.go` for safe goroutine management
- [X] T030 [US2] Integrate lifecycle manager into `App` - create lifecycle manager in NewApp(), call SetState in Bootstrap()
- [X] T031 [US2] Update `cmd/lazynuget/main.go` to call app.Run() which waits for signals and calls Shutdown
- [X] T032 [US2] Add Layer 4 panic recovery in signal handler goroutine (defer/recover with logging, don't re-panic)
- [X] T033 [US2] Add Layer 5 panic recovery in `Manager.Shutdown()` method (defer/recover with logging, continue cleanup)

### Integration Tests for User Story 2

- [X] T034 [P] [US2] Create `tests/integration/signal_test.go` with TestSIGINTHandling that sends SIGINT and validates graceful exit
- [X] T035 [P] [US2] Create `tests/integration/signal_test.go` with TestSIGTERMHandling that sends SIGTERM and validates graceful exit
- [X] T036 [P] [US2] Create `tests/integration/signal_test.go` with TestMultipleSignals that tests handling of multiple signals
- [X] T037 [P] [US2] Create `tests/integration/shutdown_test.go` with TestShutdownWithTimeout that validates timeout enforcement

**Checkpoint**: At this point, User Stories 1 AND 2 work - application starts and shuts down gracefully on signals

---

## Phase 5: User Story 3 - Configuration Override with CLI Flags (Priority: P2)

**Goal**: Enable configuration loading from files and CLI flag overrides with proper precedence (CLI > Env > File > Default)

**Independent Test**: Run `lazynuget --config /custom/path/config.yml --log-level debug` and verify settings are applied correctly

### Implementation for User Story 3

- [X] T038 [P] [US3] Implement full `internal/config/config.go` with `AppConfig` struct (all fields from data-model.md: logLevel, configPath, nonInteractive, configDir, logDir, cacheDir, theme, compactMode, showHints, startupTimeout, shutdownTimeout, maxConcurrentOps, dotnetPath, isInteractive)
- [X] T039 [US3] Implement `config.Load(args []string) (*AppConfig, error)` that merges CLI flags, environment variables, config file (YAML), and defaults with correct precedence
- [X] T040 [US3] Implement `config.Validate() error` that applies all validation rules from data-model.md (VR-006 through VR-013)
- [X] T041 [US3] Implement `config.DefaultConfig() *AppConfig` factory function with platform-specific paths resolved
- [X] T042 [US3] Add YAML config file loading in `config.Load()` using gopkg.in/yaml.v3 (handle missing file gracefully - use defaults)
- [X] T043 [US3] Add environment variable support in `config.Load()` (LAZYNUGET_LOG_LEVEL, LAZYNUGET_CONFIG, etc.)
- [X] T044 [US3] Update `App.Bootstrap()` to use real `config.Load()` instead of stub (pass CLI args)
- [X] T045 [US3] Update `internal/logging/logger.go` to implement real logger with levels (using log/slog or similar) and file output
- [X] T046 [US3] Create `tests/fixtures/configs/valid.yml` with sample valid configuration
- [X] T047 [P] [US3] Create `tests/fixtures/configs/invalid.yml` with invalid configuration (for error testing)

### Integration Tests for User Story 3

- [X] T048 [P] [US3] Create `tests/integration/config_test.go` with TestConfigFileLoading that validates loading from custom path
- [X] T049 [P] [US3] Create `tests/integration/config_test.go` with TestCLIFlagPrecedence that validates CLI flags override config file
- [X] T050 [P] [US3] Create `tests/integration/config_test.go` with TestEnvironmentVariables that validates env vars override config file
- [X] T051 [P] [US3] Create `tests/integration/config_test.go` with TestInvalidConfigFallback that validates graceful fallback to defaults

**Checkpoint**: All configuration features work - flags, env vars, files, precedence, validation

---

## Phase 6: User Story 4 - Non-Interactive Mode for Testing and Automation (Priority: P2)

**Goal**: Enable non-interactive mode via flag or TTY detection for CI/testing environments

**Independent Test**: Run `lazynuget --non-interactive --version` and `echo | lazynuget --version` (piped) and verify no TUI initialization

### Implementation for User Story 4

- [X] T052 [P] [US4] Create `internal/platform/tty.go` with `IsTerminal() bool` function using golang.org/x/term.IsTerminal()
- [X] T053 [US4] Implement `DetermineRunMode(nonInteractiveFlag bool) RunMode` in `internal/platform/detect.go` that checks flag, TTY, and environment variables (CI, NO_COLOR, TERM=dumb)
- [X] T054 [US4] Add `RunMode` enum in `internal/platform/types.go` (RunModeInteractive, RunModeNonInteractive)
- [X] T055 [US4] Update `App.Bootstrap()` to call `DetermineRunMode()` and store result in app context
- [X] T056 [US4] Add lazy GUI initialization in `App.GetGUI()` method using sync.Once pattern (only initialize if RunModeInteractive)
- [X] T057 [US4] Update `cmd/lazynuget/main.go` to check run mode before attempting GUI operations (skip GUI in non-interactive mode)
- [X] T058 [US4] Implement platform detection in `internal/platform/detect.go` (OS, architecture, paths) for platform-specific config directories

### Integration Tests for User Story 4

- [X] T059 [P] [US4] Create `tests/integration/noninteractive_test.go` with TestNonInteractiveFlagExplicit that validates --non-interactive skips TUI
- [X] T060 [P] [US4] Create `tests/integration/noninteractive_test.go` with TestNonInteractiveTTYDetection that simulates piped input and validates auto-detection
- [X] T061 [P] [US4] Create `tests/integration/noninteractive_test.go` with TestCIEnvironmentDetection that sets CI=true and validates non-interactive mode
- [X] T062 [P] [US4] Create `tests/integration/noninteractive_test.go` with TestStartupPerformance that measures startup time and validates <200ms target (using hyperfine or custom timing)

**Checkpoint**: Non-interactive mode fully functional - TTY detection, CI mode, explicit flag all work

---

## Phase 7: User Story 5 - Startup Error Recovery (Priority: P3)

**Goal**: Provide clear, actionable error messages for startup failures with proper logging and exit codes

**Independent Test**: Intentionally break config (invalid YAML, missing dotnet), verify clear error messages and correct exit codes

### Implementation for User Story 5

- [X] T063 [P] [US5] Implement error wrapping in `App.Bootstrap()` for each initialization phase (config, logger, platform) with clear error messages
- [X] T064 [P] [US5] Implement dotnet CLI validation in `internal/platform/detect.go` with `ValidateDotnetCLI() error` function (check PATH, run `dotnet --version`)
- [X] T065 [US5] Add dotnet validation to `App.Bootstrap()` after platform initialization (make it async and non-blocking per research.md Layer 2 optimization)
- [ ] T066 [US5] Implement config validation error formatting in `config.Validate()` with field name, constraint, and current value in error message
- [X] T067 [US5] Add YAML parse error handling in `config.Load()` with line/column information from parser
- [ ] T068 [US5] Add config directory permission checking in `App.Bootstrap()` (warn if cannot create, use temp directory fallback)
- [X] T069 [US5] Update error logging in panic recovery layers to write to stderr and log file (if logger initialized)
- [X] T070 [US5] Add exit code constants in `cmd/lazynuget/main.go` (ExitSuccess=0, ExitUserError=1, ExitSystemError=2) and use consistently

### Integration Tests for User Story 5

- [X] T071 [P] [US5] Create `tests/integration/errors_test.go` with TestInvalidYAMLError that provides invalid YAML and validates error message clarity
- [X] T072 [P] [US5] Create `tests/integration/errors_test.go` with TestMissingDotnetCLI that removes dotnet from PATH and validates error message with installation instructions
- [X] T073 [P] [US5] Create `tests/integration/errors_test.go` with TestConfigValidationError that provides conflicting config and validates resolution
- [X] T074 [P] [US5] Create `tests/integration/errors_test.go` with TestGracefulErrorRecovery that validates component failure handling and cleanup

**Checkpoint**: All error scenarios handled gracefully with clear messages and correct exit codes

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Final improvements that affect multiple user stories

- [ ] T075 [P] Add unit tests for `internal/bootstrap/app.go` (TestNewApp, TestBootstrap, TestShutdown) co-located in `app_test.go`
- [ ] T076 [P] Add unit tests for `internal/bootstrap/lifecycle.go` (TestStateTransitions, TestShutdownTimeout, TestGoroutineTracking) in `lifecycle_test.go`
- [ ] T077 [P] Add unit tests for `internal/bootstrap/signals.go` (TestSignalRegistration, TestForceQuit) in `signals_test.go`
- [ ] T078 [P] Add unit tests for `internal/config/config.go` (TestPrecedence, TestValidation, TestDefaults) in `config_test.go`
- [ ] T079 [P] Create `scripts/dev/test-startup.sh` performance measurement script per quickstart.md
- [ ] T080 [P] Add `DEBUG_STARTUP=1` instrumentation in `App.Bootstrap()` to log phase timing with stopwatch pattern from research.md
- [ ] T081 [P] Add build optimization flags in `Makefile` build target (-ldflags="-w -s", -trimpath) per research.md Layer 3
- [ ] T082 [P] Add version injection in `Makefile` build target (-X main.version, -X main.commit, -X main.date)
- [ ] T083 [P] Create `.github/workflows/test.yml` CI workflow with unit tests, integration tests, and coverage reporting
- [ ] T084 [P] Create `.github/workflows/performance.yml` CI workflow with startup time benchmarking using hyperfine
- [ ] T085 Create resource leak test in `tests/integration/leak_test.go` that runs 1000 startup/shutdown cycles and validates no leaks (SC-010)
- [ ] T086 Code review and refactoring pass - ensure all constitutional principles satisfied (cross-reference with plan.md constitution check)
- [ ] T087 Run `quickstart.md` validation - manually test all developer guide procedures
- [ ] T088 Update main project README.md with build and usage instructions

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-7)**: All depend on Foundational phase completion
  - US1 and US2 (P1 priority): Should be completed first - these are MVP
  - US3 and US4 (P2 priority): Can proceed after US1/US2 or in parallel if staffed
  - US5 (P3 priority): Can proceed after US1/US2 or in parallel if staffed
- **Polish (Phase 8)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories (but integrates with US1)
- **User Story 3 (P2)**: Can start after Foundational (Phase 2) - Extends US1 (adds real config loading)
- **User Story 4 (P2)**: Can start after Foundational (Phase 2) - Extends US1 (adds run mode detection)
- **User Story 5 (P3)**: Can start after Foundational (Phase 2) - Extends US1 (adds better error handling)

**Note**: US3, US4, and US5 all extend US1 capabilities but can be implemented in any order after US1/US2 are complete.

### Within Each User Story

- Implementation tasks before integration tests (need code to test)
- Models/types before services/logic
- Core functionality before error handling
- Tests can be written in parallel with implementation (TDD approach)

### Parallel Opportunities

- **Phase 1**: All tasks marked [P] can run in parallel
- **Phase 2**: T007 and T008 can run in parallel (different files)
- **Phase 3 (US1)**: T010, T011, T019, T020, T021 can run in parallel
- **Phase 4 (US2)**: T022, T023, T034, T035, T036, T037 can run in parallel initially
- **Phase 5 (US3)**: T038, T046, T047, T048, T049, T050, T051 can run in parallel
- **Phase 6 (US4)**: T052, T059, T060, T061, T062 can run in parallel
- **Phase 7 (US5)**: T063, T064, T071, T072, T073, T074 can run in parallel
- **Phase 8**: Most tasks marked [P] can run in parallel (different files/concerns)

---

## Parallel Example: User Story 1 (MVP Core)

```bash
# Launch flag parsing and version display together (different files):
Task T010: "Create internal/bootstrap/flags.go with CLI flag definitions"
Task T011: "Create internal/bootstrap/version.go with ShowVersion() function"

# Launch integration tests together (different test files):
Task T019: "Create tests/integration/version_test.go"
Task T020: "Create tests/integration/help_test.go"
Task T021: "Create tests/integration/startup_test.go"
```

## Parallel Example: User Story 2 (Shutdown)

```bash
# Launch lifecycle and signals together (different files):
Task T022: "Create internal/bootstrap/lifecycle.go with LifecycleManager struct"
Task T023: "Create internal/bootstrap/signals.go with SignalHandler struct"

# Launch all signal tests together (different test cases in same file):
Task T034: "TestSIGINTShutdown"
Task T035: "TestSIGTERMShutdown"
Task T036: "TestForceQuit"
Task T037: "TestShutdownTimeout"
```

---

## Implementation Strategy

### MVP First (User Stories 1 & 2 Only)

1. Complete Phase 1: Setup → **~1 hour**
2. Complete Phase 2: Foundational → **~1 hour**
3. Complete Phase 3: User Story 1 (Quick Launch) → **~4 hours**
4. Complete Phase 4: User Story 2 (Graceful Shutdown) → **~4 hours**
5. **STOP and VALIDATE**: Test startup and shutdown independently
6. Build binary, test `--version`, `--help`, basic startup/shutdown
7. **MVP COMPLETE** ✅ - Deploy/demo if ready

**Total MVP Time Estimate**: ~10 hours

### Incremental Delivery

1. Complete Setup + Foundational → **Foundation ready**
2. Add User Story 1 + 2 → Test independently → **MVP (P1 stories)**
3. Add User Story 3 → Test independently → **Config flexibility added**
4. Add User Story 4 → Test independently → **Testing/CI support added**
5. Add User Story 5 → Test independently → **Error handling polished**
6. Add Polish phase → **Production ready**

Each story adds value without breaking previous stories.

### Parallel Team Strategy

With multiple developers:

1. **Team completes Setup + Foundational together** → ~2 hours
2. **Once Foundational is done, split work**:
   - Developer A: User Story 1 (Quick Launch) → ~4 hours
   - Developer B: User Story 2 (Graceful Shutdown) → ~4 hours
3. **After MVP (US1+US2) complete, split again**:
   - Developer A: User Story 3 (Config Override) → ~3 hours
   - Developer B: User Story 4 (Non-Interactive Mode) → ~3 hours
   - Developer C: User Story 5 (Error Recovery) → ~3 hours
4. **Team completes Polish phase together** → ~4 hours

**Total Team Time**: ~13-15 hours (with 3 developers, calendar time ~5-6 hours)

---

## Notes

- **[P] tasks** = different files, no dependencies on incomplete work in other tasks
- **[Story] label** = maps task to specific user story for traceability and independent testing
- Each user story should be independently completable and testable
- **MVP is US1 + US2 only** - quick launch and graceful shutdown
- US3, US4, US5 are enhancements that can be deferred or parallelized
- Verify tests (pass or fail appropriately) after each story
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- **Constitutional compliance**: All tasks align with 7 principles from plan.md
- **Performance target**: <200ms startup validated in T062 (US4)
- **Safety**: 5-layer panic recovery implemented in T016, T017, T032, T033
- **Cross-platform**: Platform detection and TTY handling in US4

---

## Task Statistics

- **Total Tasks**: 88
- **Setup Tasks**: 5 (Phase 1)
- **Foundational Tasks**: 4 (Phase 2)
- **User Story 1 Tasks**: 12 (Phase 3) - MVP Core
- **User Story 2 Tasks**: 16 (Phase 4) - MVP Core
- **User Story 3 Tasks**: 14 (Phase 5) - Config Enhancement
- **User Story 4 Tasks**: 11 (Phase 6) - Testing Support
- **User Story 5 Tasks**: 12 (Phase 7) - Error Handling
- **Polish Tasks**: 14 (Phase 8)

**Parallel Opportunities**: 45 tasks marked [P] can run in parallel (51% of tasks)

**MVP Scope**: 37 tasks (Phases 1-4: Setup + Foundational + US1 + US2)

**Independent Test Criteria**:
- US1: Run `./lazynuget --version` and `./lazynuget --help` → verify output
- US2: Start app, send SIGINT → verify exit code 0, <1s shutdown
- US3: Run with `--config custom.yml --log-level debug` → verify settings applied
- US4: Run `echo | ./lazynuget --version` → verify no TUI, correct output
- US5: Provide invalid config → verify clear error message, exit code 1

**Format Validation**: ✅ All 88 tasks follow checklist format with checkbox, ID, optional [P], Story label (where applicable), and file paths
