# Data Model: Application Bootstrap and Lifecycle Management

**Feature**: Application Bootstrap and Lifecycle Management
**Branch**: 001-app-bootstrap
**Date**: 2025-11-02
**Status**: Phase 1 Design

This document defines the core entities for LazyNuGet's bootstrap system based on research findings from `research.md` and requirements from `spec.md`.

---

## Entity Overview

The bootstrap system manages three primary entities that work together to initialize, run, and gracefully shutdown LazyNuGet:

1. **Application Context** - The root container holding all application state and subsystems
2. **Configuration** - Merged settings from all sources with validation
3. **Lifecycle Manager** - Coordinates startup/shutdown sequences and state transitions

---

## Entity 1: Application Context

### Description

The Application Context represents the running LazyNuGet instance. It serves as the dependency injection container, holding references to all major subsystems and coordinating their lifecycle. This is the single source of truth for application state.

### Fields

| Field | Type | Description | Nullable | Default |
|-------|------|-------------|----------|---------|
| `config` | `*config.AppConfig` | Merged configuration from all sources | No | - |
| `logger` | `logger.Logger` | Logging subsystem | No | - |
| `platform` | `platform.Platform` | Platform detection and utilities | No | - |
| `lifecycle` | `*LifecycleManager` | Lifecycle coordinator | No | - |
| `gui` | `*tea.Program` | Bubbletea TUI program (lazy-initialized) | Yes | `nil` |
| `ctx` | `context.Context` | Root cancellation context | No | - |
| `cancel` | `context.CancelFunc` | Cancellation function for shutdown | No | - |
| `version` | `VersionInfo` | Build information (version, commit, date) | No | - |
| `startTime` | `time.Time` | Application start timestamp | No | `time.Now()` |
| `shutdownHandlers` | `[]func(context.Context) error` | Ordered cleanup handlers | No | `[]` |
| `guiOnce` | `sync.Once` | Ensures GUI initialized only once | No | - |

### Relationships

- **Owns** `Configuration` (1:1) - Configuration is immutable after initialization
- **Owns** `LifecycleManager` (1:1) - Lifecycle manager coordinates this context
- **Owns** `Logger` (1:1) - Logger is initialized first, used by all subsystems
- **Owns** `Platform` (1:1) - Platform utilities available to all components
- **Owns** `tea.Program` (0:1) - GUI only exists if in interactive mode

### State Diagram

```
┌─────────────┐
│  Uninitialized│
└──────┬──────┘
       │ NewApp()
       ▼
┌─────────────┐
│ Initializing │ ◄─── Config load
└──────┬──────┘      Logger setup
       │             Platform detect
       │ Bootstrap()
       ▼
┌─────────────┐
│   Running   │ ◄─── Event loop active
└──────┬──────┘      Processing commands
       │
       │ Shutdown signal
       ▼
┌─────────────┐
│ ShuttingDown│ ◄─── Cancel context
└──────┬──────┘      Run cleanup handlers
       │             Wait for goroutines
       │ Cleanup complete
       ▼
┌─────────────┐
│   Stopped   │ ◄─── Exit process
└─────────────┘
```

### Validation Rules

- **VR-001**: `config`, `logger`, `platform`, `lifecycle` MUST be non-nil after initialization
- **VR-002**: `ctx` MUST NOT be `context.Background()` - must be cancellable
- **VR-003**: `shutdownHandlers` MUST be registered in reverse dependency order (last init = first cleanup)
- **VR-004**: `gui` MAY be nil if in non-interactive mode
- **VR-005**: State transitions MUST follow: Uninitialized → Initializing → Running → ShuttingDown → Stopped

### Operations

**Construction**:
```go
func NewApp(version, commit, date string) (*AppContext, error)
```
- Creates new Application Context with version information
- Initializes root cancellable context
- Does NOT initialize subsystems (deferred to Bootstrap)

**Initialization**:
```go
func (app *AppContext) Bootstrap() error
```
- Initializes subsystems in dependency order (logger → config → platform)
- Validates configuration (FR-006)
- Registers shutdown handlers
- Transitions to Running state

**GUI Access** (Lazy):
```go
func (app *AppContext) GetGUI() *tea.Program
```
- Uses `sync.Once` to initialize GUI on first access (Layer 1 optimization from research)
- Returns nil if in non-interactive mode
- Initialization time excluded from startup budget unless TUI is actually used

**Shutdown**:
```go
func (app *AppContext) Shutdown(timeout time.Duration) error
```
- Cancels root context (signals all goroutines)
- Runs shutdown handlers in order with timeout enforcement
- Waits for GUI cleanup if running
- Transitions to Stopped state

---

## Entity 2: Configuration

### Description

The Configuration entity represents merged settings from all sources with proper precedence: CLI flags (highest) → environment variables → user config file → system defaults (lowest). Configuration is immutable after loading and validation.

### Fields

| Field | Type | Description | Source | Default |
|-------|------|-------------|--------|---------|
| **Flags (CLI Arguments)** | | | | |
| `showVersion` | `bool` | Display version and exit | CLI only | `false` |
| `showHelp` | `bool` | Display help and exit | CLI only | `false` |
| `configPath` | `string` | Custom config file path | CLI only | `""` |
| `logLevel` | `string` | Log verbosity level | CLI/Env/File | `"info"` |
| `nonInteractive` | `bool` | Force non-interactive mode | CLI/Env | Auto-detect |
| **Paths** | | | | |
| `configDir` | `string` | Configuration directory | Platform-specific | `~/.config/lazynuget` (Linux/macOS)<br>`%APPDATA%\lazynuget` (Windows) |
| `logDir` | `string` | Log file directory | File/Defaults | `<configDir>/logs` |
| `cacheDir` | `string` | Cache directory | File/Defaults | `<configDir>/cache` |
| **UI Preferences** | | | | |
| `theme` | `string` | Color theme name | File | `"default"` |
| `compactMode` | `bool` | Use compact UI layout | File | `false` |
| `showHints` | `bool` | Display keyboard hints | File | `true` |
| **Performance** | | | | |
| `startupTimeout` | `time.Duration` | Max startup time | File | `5s` |
| `shutdownTimeout` | `time.Duration` | Max shutdown time | File | `3s` |
| `maxConcurrentOps` | `int` | Parallel operation limit | File | `4` |
| **Environment Detection** | | | | |
| `dotnetPath` | `string` | Path to dotnet CLI | Auto-detect | From PATH |
| `isInteractive` | `bool` | TTY detected | Auto-detect | Via `term.IsTerminal()` |

### Precedence Matrix

| Setting | CLI Flag | Environment Variable | Config File | Default |
|---------|----------|---------------------|-------------|---------|
| Log Level | `--log-level debug` | `LAZYNUGET_LOG_LEVEL=debug` | `logLevel: debug` | `"info"` |
| Config Path | `--config /path` | `LAZYNUGET_CONFIG=/path` | N/A | Platform-specific |
| Non-Interactive | `--non-interactive` | `CI=true` or `NO_TTY=1` | N/A | Auto-detect |
| Theme | N/A | N/A | `theme: monokai` | `"default"` |

**Precedence Order**: CLI > Environment > File > Default (left to right, first wins)

### Validation Rules

- **VR-006**: `logLevel` MUST be one of: `"debug"`, `"info"`, `"warn"`, `"error"` (case-insensitive)
- **VR-007**: `configPath` MUST exist and be readable if specified
- **VR-008**: `startupTimeout` MUST be >= 1 second and <= 30 seconds
- **VR-009**: `shutdownTimeout` MUST be >= 1 second and <= 10 seconds
- **VR-010**: `maxConcurrentOps` MUST be >= 1 and <= 16
- **VR-011**: All directory paths MUST be absolute after resolution
- **VR-012**: `theme` MUST be a recognized theme name or `"default"`
- **VR-013**: If `showVersion` or `showHelp` is true, all other config is optional

### Operations

**Load**:
```go
func Load(args []string) (*Configuration, error)
```
- Parses CLI flags from args
- Loads environment variables
- Locates and parses config file (if exists)
- Merges sources with correct precedence
- Validates all settings (VR-006 through VR-013)
- Returns immutable Configuration or error with validation details

**Validation**:
```go
func (cfg *Configuration) Validate() error
```
- Applies all validation rules
- Checks dotnet CLI availability (FR-015)
- Verifies directory permissions
- Returns detailed error with field name and constraint

**Default Factory**:
```go
func DefaultConfig() *Configuration
```
- Returns configuration with all defaults
- Used when no config file exists (FR-009: zero-config startup)
- Platform-specific paths pre-resolved

---

## Entity 3: Lifecycle Manager

### Description

The Lifecycle Manager coordinates the application's startup and shutdown sequences. It manages state transitions, registers signal handlers, tracks running goroutines, and ensures clean resource cleanup within timeout constraints.

### Fields

| Field | Type | Description | Nullable | Default |
|-------|------|-------------|----------|---------|
| `state` | `State` | Current lifecycle state | No | `StateUninitialized` |
| `stateMutex` | `sync.RWMutex` | Protects state transitions | No | - |
| `shutdownTimeout` | `time.Duration` | Max time for graceful shutdown | No | `3s` |
| `signalChan` | `chan os.Signal` | OS signal notifications | No | buffered (1) |
| `shutdownChan` | `chan struct{}` | Internal shutdown trigger | No | unbuffered |
| `logger` | `logger.Logger` | Logger for lifecycle events | No | - |
| `goroutines` | `sync.WaitGroup` | Tracks background workers | No | - |
| `stopSignal` | `context.CancelFunc` | Restores default signal behavior | Yes | From `signal.NotifyContext` |

### State Enumeration

```go
type State int

const (
    StateUninitialized State = iota  // 0: Not yet started
    StateInitializing                 // 1: Loading subsystems
    StateRunning                      // 2: Fully operational
    StateShuttingDown                 // 3: Cleanup in progress
    StateStopped                      // 4: Fully stopped
)
```

### State Transition Rules

| From State | To State | Trigger | Conditions |
|------------|----------|---------|------------|
| Uninitialized | Initializing | `Start()` called | None |
| Initializing | Running | All subsystems initialized | Config valid, logger ready |
| Initializing | Stopped | Initialization failure | Any subsystem fails to start |
| Running | ShuttingDown | Signal received / `Shutdown()` | SIGINT, SIGTERM, or explicit call |
| Running | ShuttingDown | Panic recovered | Layer 2 panic recovery triggered |
| ShuttingDown | Stopped | Cleanup complete | All handlers finished within timeout |
| ShuttingDown | Stopped | Timeout exceeded | `shutdownTimeout` exceeded |

**Invalid Transitions** (will return error):
- Any state → Uninitialized (can't go backwards)
- Stopped → Any state (can't restart, must create new instance)

### Validation Rules

- **VR-014**: State transitions MUST be atomic (mutex-protected)
- **VR-015**: `shutdownTimeout` MUST be enforced - force exit if exceeded
- **VR-016**: Signal handler MUST be registered before entering Running state
- **VR-017**: `goroutines` WaitGroup MUST reach zero before transition to Stopped
- **VR-018**: Second signal (force quit) MUST call `stopSignal()` to restore default behavior

### Operations

**Start**:
```go
func (lm *LifecycleManager) Start(ctx context.Context, app *AppContext) error
```
- Transitions Uninitialized → Initializing
- Registers signal handlers (SIGINT, SIGTERM) via `signal.NotifyContext`
- Initializes all subsystems via `app.Bootstrap()`
- Transitions Initializing → Running on success
- Transitions Initializing → Stopped on failure

**Run**:
```go
func (lm *LifecycleManager) Run(ctx context.Context) error
```
- Blocks until shutdown signal received
- Monitors `signalChan` and `ctx.Done()`
- First signal: Graceful shutdown
- Second signal: Force quit (calls `lm.stopSignal()`)

**Shutdown**:
```go
func (lm *LifecycleManager) Shutdown(ctx context.Context) error
```
- Transitions Running → ShuttingDown
- Creates timeout context from `shutdownTimeout`
- Closes `shutdownChan` to notify all listeners
- Waits for `goroutines` WaitGroup with timeout
- Calls all shutdown handlers in order
- Transitions ShuttingDown → Stopped when complete or timeout

**State Query**:
```go
func (lm *LifecycleManager) State() State
```
- Thread-safe read of current state
- Used by subsystems to check if shutdown in progress

**Goroutine Tracking**:
```go
func (lm *LifecycleManager) Go(fn func() error)
```
- Wraps goroutine with WaitGroup tracking
- Applies Layer 3 panic recovery
- Logs errors from goroutine

---

## Entity Relationships

```
┌─────────────────────────────────────────────────────────┐
│             Application Context (Root)                   │
│  ┌───────────────────────────────────────────────────┐  │
│  │ Fields:                                           │  │
│  │  - config: *Configuration                        │  │
│  │  - logger: Logger                                │  │
│  │  - platform: Platform                            │  │
│  │  - lifecycle: *LifecycleManager                  │  │
│  │  - gui: *tea.Program (lazy)                      │  │
│  │  - ctx: context.Context                          │  │
│  └───────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
         │              │              │
         │ owns         │ owns         │ owns
         ▼              ▼              ▼
┌────────────────┐ ┌────────────────┐ ┌────────────────┐
│ Configuration  │ │Lifecycle       │ │ Logger         │
│                │ │Manager         │ │ Platform       │
│ - Flags        │ │                │ │ GUI            │
│ - Paths        │ │ - State        │ │                │
│ - UI Prefs     │ │ - Signals      │ │                │
│ - Performance  │ │ - Goroutines   │ │                │
└────────────────┘ └────────────────┘ └────────────────┘
```

**Initialization Order** (from research.md manual DI pattern):
1. Logger (first - captures all events)
2. Configuration (needs logger for validation warnings)
3. Platform (needs config and logger)
4. Lifecycle Manager (needs logger, coordinates others)
5. GUI (lazy - only if interactive mode)

**Shutdown Order** (reverse):
1. GUI (stop event loop, restore terminal)
2. Lifecycle Manager (cancel goroutines, wait for completion)
3. Platform (release platform-specific resources)
4. Configuration (flush any cached writes)
5. Logger (flush buffers, close files)

---

## Performance Characteristics

Based on research.md optimization strategies:

| Entity | Initialization Time | Memory Footprint | Notes |
|--------|-------------------|------------------|-------|
| Configuration | <30ms | ~5KB | Includes file I/O and validation |
| Logger | <10ms | ~2KB | File handle + buffer |
| Platform | <20ms | ~3KB | Detection + path normalization |
| Lifecycle Manager | <5ms | ~1KB | Mostly just state setup |
| GUI (lazy) | 80-120ms | ~5MB | Only initialized when needed |
| **Total (non-interactive)** | **<70ms** | **<15KB** | Excludes GUI |
| **Total (interactive)** | **<165ms** | **~5MB** | Includes GUI |

**Optimization Applied**:
- Layer 1: Lazy GUI initialization (80-120ms saved when not needed)
- Layer 2: Async dotnet validation (non-blocking)
- Layer 4: Parallel init with `errgroup` (10-20ms saved)

**Target**: <200ms p95 startup ✅ Achievable

---

## Testing Strategy

### Unit Tests

**Configuration Entity**:
- Test precedence: CLI > Env > File > Default
- Test validation rules (VR-006 through VR-013)
- Test invalid config file formats
- Test missing config file (should use defaults)

**Lifecycle Manager Entity**:
- Test state transitions (valid and invalid)
- Test timeout enforcement
- Test goroutine tracking
- Test signal handling (mock signals)

**Application Context Entity**:
- Test lazy GUI initialization
- Test shutdown handler ordering
- Test context cancellation propagation

### Integration Tests

**Startup Cycle**:
- Full bootstrap → Running state
- Measure startup time (<200ms target)
- Verify all subsystems initialized

**Shutdown Cycle**:
- Running → Stopped within 1 second (normal)
- Running → Stopped within 3 seconds (with background work)
- Verify no resource leaks (run 1000 iterations per SC-010)

**Signal Handling**:
- SIGINT triggers graceful shutdown
- SIGTERM triggers graceful shutdown
- Second signal forces immediate exit

---

## Constitutional Alignment

All entities align with LazyNuGet's constitutional principles:

- **Principle II (Simplicity)**: Zero-config startup via defaults in Configuration
- **Principle III (Safety)**: Panic recovery at all lifecycle layers
- **Principle IV (Cross-Platform)**: Platform-specific paths in Configuration
- **Principle V (Performance)**: Lazy initialization, <200ms target, 10MB baseline
- **Principle VII (Clean Code)**: Clear separation of concerns, testable interfaces

---

## Next Steps

Phase 1 continues with:

1. ✅ `data-model.md` (this document)
2. ⏸️ `contracts/` directory - Interface definitions for each entity
3. ⏸️ `quickstart.md` - Developer guide for working with these entities

After Phase 1 completes:
- Run `/speckit.tasks` to generate implementation task breakdown
- Begin Phase 2 implementation

**Status**: Phase 1 Data Model COMPLETE
