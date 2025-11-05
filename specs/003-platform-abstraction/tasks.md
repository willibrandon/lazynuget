# Tasks: Cross-Platform Infrastructure

**Input**: Design documents from `/specs/003-platform-abstraction/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Tests are included as this is foundational infrastructure requiring comprehensive validation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and platform abstraction structure

- [X] T001 Create internal/platform/ directory structure per plan.md
- [X] T002 Define PlatformInfo interface in internal/platform/platform.go
- [X] T003 Define PathResolver interface in internal/platform/paths.go
- [X] T004 Define TerminalCapabilities interface in internal/platform/terminal.go
- [X] T005 Define ProcessSpawner interface in internal/platform/process.go
- [X] T006 Define shared types (ColorDepth, ProcessResult) in internal/platform/types.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core platform detection and factory that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T007 Implement PlatformInfo detection (OS, Arch, Version) in internal/platform/detect.go
- [X] T008 Add build tag for Windows-specific code: internal/platform/detect_windows.go
- [X] T009 [P] Add build tag for Unix-specific code: internal/platform/detect_unix.go
- [X] T010 Create factory function platform.New() in internal/platform/factory.go
- [X] T011 [P] Add unit tests for platform detection in internal/platform/detect_test.go
- [X] T012 Update internal/bootstrap/app.go to use platform.New() instead of existing detection

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Developer Runs LazyNuGet on Any Platform (Priority: P1) 🎯 MVP

**Goal**: LazyNuGet auto-detects platform, uses correct config/cache paths, and displays correctly in terminal

**Independent Test**: Install on Windows/macOS/Linux, run without config, verify correct default paths and startup

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T013 [P] [US1] Integration test for Windows config directory in tests/integration/platform_windows_test.go
- [ ] T014 [P] [US1] Integration test for macOS config directory in tests/integration/platform_darwin_test.go
- [ ] T015 [P] [US1] Integration test for Linux XDG config directory in tests/integration/platform_linux_test.go
- [ ] T016 [P] [US1] Unit test for ConfigDir() in internal/platform/paths_test.go
- [ ] T017 [P] [US1] Unit test for CacheDir() in internal/platform/paths_test.go

### Implementation for User Story 1

**Path Resolution**:

- [ ] T018 [P] [US1] Implement Windows directory resolution in internal/platform/paths_windows.go
- [ ] T019 [P] [US1] Implement macOS directory resolution in internal/platform/paths_darwin.go
- [ ] T020 [P] [US1] Implement Linux XDG directory resolution in internal/platform/paths_linux.go
- [ ] T021 [US1] Implement PathResolver factory in internal/platform/paths.go
- [ ] T022 [US1] Add directory creation with permissions in internal/platform/paths.go (EnsureDir)
- [ ] T023 [US1] Add parent validation and fallback logic per FR-025

**Terminal Capabilities**:

- [ ] T024 [P] [US1] Implement color depth detection in internal/platform/terminal.go (detectColorLevel)
- [ ] T025 [P] [US1] Implement Unicode support detection in internal/platform/terminal.go (detectUnicodeSupport)
- [ ] T026 [P] [US1] Implement terminal size detection in internal/platform/terminal.go (GetSize)
- [ ] T027 [P] [US1] Implement TTY detection in internal/platform/terminal.go (IsTTY)
- [ ] T028 [US1] Create TerminalCapabilities factory in internal/platform/terminal.go
- [ ] T029 [P] [US1] Add unit tests for color detection in internal/platform/terminal_test.go
- [ ] T030 [P] [US1] Add unit tests for Unicode detection in internal/platform/terminal_test.go

**Integration**:

- [ ] T031 [US1] Update internal/config/config.go to use PathResolver.ConfigDir()
- [ ] T032 [US1] Update internal/bootstrap/app.go to cache PathResolver instance
- [ ] T033 [US1] Update internal/platform/detect.go to use TerminalCapabilities.IsTTY()
- [ ] T034 [US1] Add logging for detected capabilities at startup
- [ ] T035 [US1] Verify integration test passes on all platforms in CI

**Checkpoint**: User Story 1 fully functional - LazyNuGet starts correctly with platform-appropriate defaults on Windows/macOS/Linux

---

## Phase 4: User Story 2 - User Works with Platform-Specific Paths (Priority: P2)

**Goal**: Users can configure custom paths in platform-native formats (drive letters, UNC, Unix paths)

**Independent Test**: Set custom config/cache paths on each platform, verify LazyNuGet correctly loads and validates them

### Tests for User Story 2

- [ ] T036 [P] [US2] Unit test for Windows drive letter normalization in internal/platform/paths_windows_test.go
- [ ] T037 [P] [US2] Unit test for Windows UNC path handling in internal/platform/paths_windows_test.go
- [ ] T038 [P] [US2] Unit test for Unix absolute/relative path handling in internal/platform/paths_unix_test.go
- [ ] T039 [P] [US2] Unit test for path validation in internal/platform/paths_test.go
- [ ] T040 [P] [US2] Integration test for mixed separators in tests/integration/paths_test.go

### Implementation for User Story 2

**Windows Path Handling**:

- [ ] T041 [P] [US2] Implement drive letter normalization in internal/platform/paths_windows.go
- [ ] T042 [P] [US2] Implement UNC path detection and handling in internal/platform/paths_windows.go
- [ ] T043 [P] [US2] Implement long path support (>260 chars) in internal/platform/paths_windows.go
- [ ] T044 [US2] Implement Normalize() for Windows in internal/platform/paths_windows.go

**Unix Path Handling**:

- [ ] T045 [P] [US2] Implement absolute path detection in internal/platform/paths_unix.go
- [ ] T046 [P] [US2] Implement relative path resolution in internal/platform/paths_unix.go
- [ ] T047 [US2] Implement Normalize() for Unix in internal/platform/paths_unix.go

**Path Validation**:

- [ ] T048 [US2] Implement Validate() with platform-specific rules in internal/platform/paths.go
- [ ] T049 [US2] Implement IsAbsolute() with platform detection in internal/platform/paths.go
- [ ] T050 [US2] Implement Resolve() for relative-to-config paths in internal/platform/paths.go
- [ ] T051 [US2] Add descriptive error messages for invalid paths per FR-021

**Integration**:

- [ ] T052 [US2] Update internal/config/config.go to call Validate() on all user-provided paths
- [ ] T053 [US2] Update internal/config/config.go to call Normalize() before storing paths
- [ ] T054 [US2] Add integration test for custom config path via CLI flag in tests/integration/config_test.go
- [ ] T055 [US2] Add integration test for custom config path via env var in tests/integration/config_test.go

**Checkpoint**: User Story 2 complete - Users can specify platform-native paths with full validation

---

## Phase 5: User Story 3 - User Runs in Limited Terminal Environment (Priority: P3)

**Goal**: LazyNuGet gracefully degrades UI in CI, SSH, or limited terminals (no color, no Unicode, narrow width)

**Independent Test**: Run with TERM=dumb, verify ASCII-only output without crashes; run in GitHub Actions, verify non-interactive mode

### Tests for User Story 3

- [ ] T056 [P] [US3] Unit test for ColorNone detection with NO_COLOR in internal/platform/terminal_test.go
- [ ] T057 [P] [US3] Unit test for ColorNone detection with TERM=dumb in internal/platform/terminal_test.go
- [ ] T058 [P] [US3] Unit test for Unicode fallback with LANG=C in internal/platform/terminal_test.go
- [ ] T059 [P] [US3] Integration test for non-interactive mode in CI in tests/integration/noninteractive_test.go
- [ ] T060 [P] [US3] Integration test for terminal resize handling in tests/integration/terminal_test.go

### Implementation for User Story 3

**Graceful Degradation**:

- [ ] T061 [P] [US3] Add NO_COLOR environment variable check in internal/platform/terminal.go
- [ ] T062 [P] [US3] Add TERM=dumb handling in internal/platform/terminal.go
- [ ] T063 [P] [US3] Add minimum dimension validation in internal/platform/terminal.go (40x10 minimum)
- [ ] T064 [US3] Add dimension clamping (40-500 width, 10-200 height) per data-model.md

**Terminal Resize**:

- [ ] T065 [US3] Implement SIGWINCH signal handler in internal/platform/terminal_unix.go
- [ ] T066 [US3] Implement console event handler for Windows in internal/platform/terminal_windows.go
- [ ] T067 [US3] Implement WatchResize() with callback pattern in internal/platform/terminal.go
- [ ] T068 [US3] Add resize channel with buffering to prevent blocking per FR-028

**Integration**:

- [ ] T069 [US3] Update internal/bootstrap/app.go to log terminal capabilities at startup
- [ ] T070 [US3] Add warning log when terminal dimensions are below minimum
- [ ] T071 [US3] Verify graceful degradation with TERM=dumb in CI tests
- [ ] T072 [US3] Add manual testing checklist for iTerm2, Windows Terminal, Alacritty per research.md

**Checkpoint**: User Story 3 complete - LazyNuGet works in constrained environments with graceful degradation

---

## Phase 6: User Story 4 - Developer Spawns Platform-Specific Processes (Priority: P2)

**Goal**: LazyNuGet executes `dotnet` commands correctly on all platforms with proper output encoding

**Independent Test**: Trigger `dotnet restore` on each platform, verify command executes and output is correctly decoded to UTF-8

### Tests for User Story 4

- [ ] T073 [P] [US4] Unit test for Windows code page detection in internal/platform/process_windows_test.go
- [ ] T074 [P] [US4] Unit test for Unix locale detection in internal/platform/process_unix_test.go
- [ ] T075 [P] [US4] Unit test for UTF-8 encoding in internal/platform/process_test.go
- [ ] T076 [P] [US4] Integration test for dotnet --version on all platforms in tests/integration/process_test.go
- [ ] T077 [P] [US4] Integration test for process with non-UTF-8 output in tests/integration/encoding_test.go

### Implementation for User Story 4

**Encoding Detection**:

- [ ] T078 [P] [US4] Implement Windows code page detection (GetACP/GetConsoleOutputCP) in internal/platform/encoding_windows.go
- [ ] T079 [P] [US4] Implement Unix locale parsing (LC_ALL/LANG) in internal/platform/encoding_unix.go
- [ ] T080 [US4] Implement encoding detection with UTF-8 fallback in internal/platform/encoding.go
- [ ] T081 [US4] Add mapping from code page to golang.org/x/text/encoding per research.md

**Process Spawning**:

- [ ] T082 [P] [US4] Implement ProcessSpawner.Run() with output capture in internal/platform/process.go
- [ ] T083 [P] [US4] Implement ProcessSpawner.SetEncoding() for manual override in internal/platform/process.go
- [ ] T084 [US4] Add UTF-8 decode attempt with fallback to system encoding per FR-030
- [ ] T085 [US4] Add replacement character (�) for invalid sequences
- [ ] T086 [US4] Implement executable PATH resolution in internal/platform/process.go

**Platform-Specific Process Handling**:

- [ ] T087 [P] [US4] Implement Windows executable extension search (.exe/.cmd/.bat) in internal/platform/process_windows.go
- [ ] T088 [P] [US4] Implement Windows argument quoting (CommandLineToArgvW rules) in internal/platform/process_windows.go
- [ ] T089 [P] [US4] Implement Unix execute permission check in internal/platform/process_unix.go
- [ ] T090 [P] [US4] Implement Unix shell-style argument quoting in internal/platform/process_unix.go

**Integration**:

- [ ] T091 [US4] Create ProcessSpawner factory in internal/platform/process.go
- [ ] T092 [US4] Update internal/bootstrap/app.go to use ProcessSpawner for dotnet validation
- [ ] T093 [US4] Add error context for process failures (executable not found, permission denied)
- [ ] T094 [US4] Add integration test for dotnet restore with actual project file
- [ ] T095 [US4] Verify encoding detection works with CP437, Shift-JIS locales in manual testing

**Checkpoint**: User Story 4 complete - LazyNuGet correctly spawns processes with proper encoding on all platforms

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T096 [P] Add comprehensive godoc comments to all public interfaces
- [ ] T097 [P] Update specs/003-platform-abstraction/quickstart.md with real code examples
- [ ] T098 [P] Add performance benchmarks for path normalization (target <1ms per FR-010)
- [ ] T099 [P] Add performance benchmarks for terminal detection (target <10ms)
- [ ] T100 Code review: Verify all build tags are correct (//go:build syntax)
- [ ] T101 Code review: Verify all errors include context per research.md
- [ ] T102 [P] Add integration test for environment variable precedence (APPDATA > XDG on Windows)
- [ ] T103 [P] Add integration test for read-only cache directory graceful degradation per FR-026
- [ ] T104 [P] Add integration test for read-only config directory failure per FR-027
- [ ] T105 Update CLAUDE.md with platform abstraction architecture notes
- [ ] T106 Run go mod tidy to ensure dependencies are minimal
- [ ] T107 Verify code coverage >80% per constitutional principle 7
- [ ] T108 Run quickstart.md validation (compile and run all code examples)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion
  - User Story 1 (P1): No dependencies on other stories
  - User Story 2 (P2): No dependencies on other stories (independent path handling)
  - User Story 3 (P3): No dependencies on other stories (terminal degradation)
  - User Story 4 (P2): No dependencies on other stories (process spawning)
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Independence

Each user story can be implemented and tested independently after Foundation is complete:

- **US1 (Platform Detection)**: Standalone - config/cache paths + terminal detection
- **US2 (Path Handling)**: Standalone - path normalization and validation
- **US3 (Terminal Graceful Degradation)**: Standalone - limited terminal support
- **US4 (Process Spawning)**: Standalone - dotnet command execution

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Interfaces before implementations
- Platform-specific code before integration
- Unit tests before integration tests
- Story complete before moving to next priority

### Parallel Opportunities

**Phase 1 (Setup)**: All tasks T002-T006 can run in parallel (different files)

**Phase 2 (Foundational)**: T008-T009, T011 can run in parallel after T007

**User Story 1**:
- Tests T013-T017 can run in parallel
- Implementations T018-T020 can run in parallel
- Implementations T024-T027 can run in parallel
- Implementations T029-T030 can run in parallel

**User Story 2**:
- Tests T036-T040 can run in parallel
- Implementations T041-T043 can run in parallel
- Implementations T045-T046 can run in parallel

**User Story 3**:
- Tests T056-T060 can run in parallel
- Implementations T061-T063 can run in parallel

**User Story 4**:
- Tests T073-T077 can run in parallel
- Implementations T078-T079 can run in parallel
- Implementations T082-T083 can run in parallel
- Implementations T087-T090 can run in parallel

**Phase 7 (Polish)**: Most tasks T096-T104 can run in parallel

**Cross-Story Parallelization**: After Foundational (Phase 2), all four user stories can be worked on in parallel by different developers

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together:
Task: "Integration test for Windows config directory in tests/integration/platform_windows_test.go"
Task: "Integration test for macOS config directory in tests/integration/platform_darwin_test.go"
Task: "Integration test for Linux XDG config directory in tests/integration/platform_linux_test.go"
Task: "Unit test for ConfigDir() in internal/platform/paths_test.go"
Task: "Unit test for CacheDir() in internal/platform/paths_test.go"

# Launch all Windows path implementations together:
Task: "Implement Windows directory resolution in internal/platform/paths_windows.go"

# Launch all terminal detection implementations together:
Task: "Implement color depth detection in internal/platform/terminal.go"
Task: "Implement Unicode support detection in internal/platform/terminal.go"
Task: "Implement terminal size detection in internal/platform/terminal.go"
Task: "Implement TTY detection in internal/platform/terminal.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test User Story 1 independently on all platforms
5. Deploy/demo - LazyNuGet now works cross-platform with correct defaults

### Incremental Delivery (Recommended)

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 (P1) → Test independently → Deploy/Demo (MVP!)
3. Add User Story 4 (P2) → Test independently → Deploy/Demo (process spawning)
4. Add User Story 2 (P2) → Test independently → Deploy/Demo (custom paths)
5. Add User Story 3 (P3) → Test independently → Deploy/Demo (graceful degradation)
6. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (P1 - highest priority)
   - Developer B: User Story 4 (P2 - process spawning)
   - Developer C: User Story 2 (P2 - path handling)
   - Developer D: User Story 3 (P3 - terminal degradation)
3. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Build tags ensure platform-specific code only compiles on target platform
- All paths must use filepath package, not path package
- Terminal detection cached at startup for performance
- Process encoding detection fallback chain: UTF-8 → system encoding → replacement character
