# Implementation Plan: Application Bootstrap and Lifecycle Management

**Branch**: `001-app-bootstrap` | **Date**: 2025-11-02 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-app-bootstrap/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Implement the core application bootstrap system that initializes LazyNuGet with proper dependency injection, handles graceful shutdown with signal processing, manages application lifecycle, parses command-line arguments, and supports both interactive TUI mode and non-interactive mode for testing. This is the foundation infrastructure that all other features depend on.

**Technical Approach**: Use Go's standard library for CLI parsing and signal handling, implement a lightweight dependency injection container pattern, integrate with Bubbletea's Program lifecycle for TUI management, and ensure cross-platform compatibility through platform detection and path normalization.

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**:
- Go standard library (flag, os, os/signal, context)
- github.com/charmbracelet/bubbletea (TUI framework - deferred to Track 4)
- Platform detection utilities (Track 1, Spec 003)
- Configuration system (Track 1, Spec 002)
- Logging framework (Track 1, Spec 004)

**Storage**: Configuration files (YAML) in platform-specific locations; no database required for bootstrap
**Testing**: Go's testing package with table-driven tests, integration tests for startup/shutdown cycles
**Target Platform**: Cross-platform (Windows, macOS, Linux), SSH sessions, containers, WSL
**Project Type**: Single binary CLI application with TUI
**Performance Goals**:
- <200ms p95 cold start (constitutional requirement)
- <1s normal shutdown
- <3s forced shutdown with cleanup
- <10MB memory at idle post-bootstrap

**Constraints**:
- Must work without any configuration files (embedded defaults)
- Must handle missing dotnet CLI gracefully
- Must support TTY and non-TTY environments
- Signal handling must work consistently across all platforms
- No external process dependencies at bootstrap time (except dotnet validation)

**Scale/Scope**:
- Single-instance application (no multi-instance coordination)
- Handles 0-100+ projects in solution
- Supports all .NET SDK versions with dotnet CLI
- Works in terminals from 80x24 to 200x100+

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle Alignment

**I. Discoverability**: ✅ PASS
- `--help` flag provides usage information
- `--version` flag shows version
- Clear error messages for startup failures
- No hidden behavior - all flags documented

**II. Simplicity & 80/20 Rule**: ✅ PASS
- Zero-config startup (works with defaults)
- Simple command: `lazynuget` to start
- Common operations don't require flags
- Advanced options available but not required

**III. Safety & Confirmation**: ✅ PASS
- Graceful shutdown prevents data loss
- Panic recovery prevents crashes
- Backup/rollback handled by operation layers (not bootstrap)
- Clean resource cleanup

**IV. Cross-Platform Excellence**: ✅ PASS
- Platform detection and path normalization required
- Signal handling cross-platform
- Config locations follow platform conventions
- Identical experience on Windows, macOS, Linux
- SSH/container/WSL support via non-interactive mode

**V. Performance & Responsiveness**: ✅ PASS
- <200ms startup target (FR-018, SC-001)
- <10MB memory baseline (SC-008)
- Non-blocking initialization
- Resource leak prevention (SC-010)

**VI. Conformity with dotnet CLI**: ✅ PASS
- Validates dotnet CLI availability (FR-015)
- No custom implementations at bootstrap level
- Respects .NET conventions

**VII. Clean, Testable, Maintainable Code**: ✅ PASS
- Non-interactive mode for testing (FR-011, FR-012)
- Clear separation of concerns (bootstrap → config → services → GUI)
- Integration tests for startup/shutdown (SC-010)
- Dependency injection for testability

### Gates Status

✅ **All gates PASSED** - No constitutional violations. Ready for Phase 0 research.

## Project Structure

### Documentation (this feature)

```text
specs/001-app-bootstrap/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (architectural patterns)
├── data-model.md        # Phase 1 output (entities: AppContext, Config, Lifecycle)
├── quickstart.md        # Phase 1 output (developer guide)
├── contracts/           # Phase 1 output (bootstrap API contracts)
│   ├── bootstrap.go     # Application bootstrap interface
│   ├── lifecycle.go     # Lifecycle management interface
│   └── signals.go       # Signal handling interface
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created yet)
```

### Source Code (repository root)

```text
cmd/
└── lazynuget/
    └── main.go          # Entry point, minimal - delegates to bootstrap

internal/
├── bootstrap/           # THIS FEATURE
│   ├── app.go          # Application struct, dependency container
│   ├── lifecycle.go    # Lifecycle manager (startup/shutdown)
│   ├── flags.go        # CLI argument parsing
│   ├── signals.go      # Signal handling (SIGINT, SIGTERM)
│   ├── version.go      # Version information display
│   └── bootstrap_test.go
│
├── config/              # Track 1, Spec 002 (dependency)
│   └── config.go       # Config loading (to be implemented)
│
├── platform/            # Track 1, Spec 003 (dependency)
│   └── detect.go       # Platform detection (to be implemented)
│
├── logging/             # Track 1, Spec 004 (dependency)
│   └── logger.go       # Logging framework (to be implemented)
│
└── gui/                 # Track 4, Spec 019 (integration point)
    └── app.go          # Bubbletea app (to be implemented)

tests/
├── integration/
│   ├── bootstrap_test.go      # Full startup/shutdown cycle tests
│   ├── signals_test.go        # Signal handling tests
│   └── noninteractive_test.go # Non-TTY mode tests
└── fixtures/
    └── configs/               # Test configuration files

scripts/
└── dev/
    └── test-startup.sh        # Development script for manual testing
```

**Structure Decision**: Single project structure selected. LazyNuGet is a single binary CLI tool with internal packages for organization. The `cmd/lazynuget/main.go` entry point is minimal and delegates immediately to `internal/bootstrap/app.go`. Internal packages enforce encapsulation and prevent external dependencies on internal implementation details. Testing infrastructure includes both unit tests (co-located with code) and integration tests (separate tests/ directory) per Go conventions.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations detected. All constitutional principles satisfied.

## Phase 0: Research & Architecture

**Status**: ✅ COMPLETE

### Research Tasks

1. **Dependency Injection Pattern** (Priority: High)
   - **Question**: What DI pattern best suits Go for LazyNuGet's bootstrap needs?
   - **Context**: Need lightweight DI for wiring up config, logger, platform utils, GUI
   - **Research Focus**: Compare manual DI, wire, dig, fx patterns
   - **Success Criteria**: Pattern selected with rationale documented

2. **Signal Handling Cross-Platform** (Priority: High)
   - **Question**: How to ensure consistent signal behavior across Windows, macOS, Linux?
   - **Context**: SIGINT/SIGTERM handling varies by platform, especially Windows
   - **Research Focus**: os/signal package capabilities, platform differences, best practices
   - **Success Criteria**: Cross-platform signal handling approach documented

3. **Bubbletea Lifecycle Integration** (Priority: High)
   - **Question**: How does bootstrap hand off to Bubbletea Program and coordinate shutdown?
   - **Context**: Need clean integration between bootstrap and TUI framework
   - **Research Focus**: Bubbletea Program lifecycle, tea.Quit(), context cancellation
   - **Success Criteria**: Integration pattern documented with examples

4. **Non-Interactive Mode Detection** (Priority: Medium)
   - **Question**: Best way to detect TTY vs non-TTY environments reliably?
   - **Context**: Auto-detect CI/testing environments, support SSH/containers
   - **Research Focus**: terminal.IsTerminal(), os.Stdin.Stat(), platform quirks
   - **Success Criteria**: TTY detection approach with fallback strategies

5. **Startup Performance Optimization** (Priority: Medium)
   - **Question**: What initialization strategies keep startup under 200ms?
   - **Context**: Must meet constitutional <200ms p95 target
   - **Research Focus**: Lazy initialization, parallel loading, measurement approaches
   - **Success Criteria**: Performance strategy documented with measurement plan

6. **Panic Recovery Strategy** (Priority: Low)
   - **Question**: Where and how should panic recovery be implemented?
   - **Context**: FR-016 requires graceful panic recovery during startup/shutdown
   - **Research Focus**: defer/recover patterns, logging, exit codes
   - **Success Criteria**: Panic recovery approach documented

### Research Execution

Research tasks will be dispatched to specialized agents to gather best practices, evaluate alternatives, and document decisions. Each research task will produce a decision with rationale and alternatives considered.

**Output**: `research.md` with all decisions documented

---

## Phase 1: Design & Contracts

**Status**: ✅ COMPLETE

### Data Model

**Output**: `data-model.md`

**Entities to Define**:

1. **Application Context** (from spec Key Entities)
   - Fields: config, logger, lifecycle, services, GUI, context.Context
   - Lifecycle states: Starting, Running, ShuttingDown, Stopped
   - Relationships: Owns all major subsystems

2. **Configuration** (from spec Key Entities)
   - Fields: flags (CLIFlags), paths (Paths), ui (UIPreferences), perf (Performance)
   - Validation rules: Required fields, value constraints
   - Precedence: CLI > env > file > defaults

3. **Lifecycle Manager** (from spec Key Entities)
   - Fields: state (State), shutdownTimeout (duration), signalChan (chan os.Signal)
   - State transitions: Start() → Run() → Shutdown() → Stop()
   - Operations: Initialize components, coordinate shutdown

### API Contracts

**Output**: `contracts/` directory with interface definitions

**Contracts to Generate**:

1. **bootstrap.go**: Application bootstrap interface
   ```go
   // Bootstrap interface for application initialization
   type Bootstrapper interface {
       Initialize() error
       Run() error
       Shutdown(ctx context.Context) error
   }
   ```

2. **lifecycle.go**: Lifecycle management interface
   ```go
   // Lifecycle manages application state transitions
   type Lifecycle interface {
       Start(ctx context.Context) error
       Stop(ctx context.Context) error
       State() State
   }
   ```

3. **signals.go**: Signal handling interface
   ```go
   // SignalHandler manages OS signal processing
   type SignalHandler interface {
       Register(signals ...os.Signal)
       Wait() os.Signal
       Shutdown()
   }
   ```

### Quickstart Guide

**Output**: `quickstart.md`

**Content**:
- Developer setup instructions
- How to build and run lazynuget
- How to run tests (unit + integration)
- How to test startup/shutdown manually
- How to test non-interactive mode
- Performance measurement approach

### Agent Context Update

After Phase 1 completes, run:
```bash
.specify/scripts/bash/update-agent-context.sh claude
```

This will update `.claude/context.md` with technology choices from research.md and data-model.md.

---

## Phase 1 Artifacts Generated

✅ **data-model.md**: Comprehensive entity definitions with:
- Application Context (12 fields, state diagram, validation rules)
- Configuration (15 fields, precedence matrix, validation rules)
- Lifecycle Manager (8 fields, state enumeration, transition rules)
- Entity relationships and initialization/shutdown ordering
- Performance characteristics and testing strategy

✅ **contracts/** directory with 3 interface definitions:
- `bootstrap.go`: Bootstrapper interface with StartupPhase, ShutdownPhase, ExitCode, StartupMetrics
- `lifecycle.go`: Lifecycle interface with State, ShutdownHandler, LifecycleConfig, LifecycleMetrics
- `signals.go`: SignalHandler interface with SignalNotification, PlatformSignals, integration examples

✅ **quickstart.md**: Developer guide with:
- Prerequisites and project setup
- Build instructions (dev + production with optimizations)
- Testing strategies (unit, integration, table-driven)
- Performance measurement (hyperfine, manual scripts, CI integration)
- Manual testing procedures for all scenarios
- Debugging techniques (logs, Delve, profiling)
- Code organization and performance targets

---

## Phase 2: Task Breakdown

**Status**: ⏸️ PENDING (blocked on Phase 1 design completion - NOW READY)

Task breakdown happens in the next command: `/speckit.tasks`

The tasks command will generate `tasks.md` with prioritized, dependency-ordered tasks for implementation based on this plan and the design artifacts from Phase 1.

---

## Next Steps

1. ✅ Complete Phase 0: Research architectural patterns → `research.md`
2. ✅ Complete Phase 1: Design data model and contracts → `data-model.md`, `contracts/`, `quickstart.md`
3. ⏸️ Update agent context → `.claude/context.md` (run `.specify/scripts/bash/update-agent-context.sh claude`)
4. ⏸️ Run `/speckit.tasks` to generate task breakdown → `tasks.md`
5. ⏸️ Run `/speckit.implement` to execute tasks

**Current Status**: Phase 0 and Phase 1 COMPLETE. Ready for task breakdown.
