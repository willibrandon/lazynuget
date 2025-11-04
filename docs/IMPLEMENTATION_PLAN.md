# LazyNuGet Complete Implementation Plan

**Version**: 1.1
**Author**: willibrandon
**Date**: 2025-11-03
**Status**: Planning
**Update**: Added gonuget library integration strategy (hybrid approach)

This document provides the complete list of /speckit.specify commands required to implement a production-ready LazyNuGet from the ground up. Every feature from the proposal will be implemented - nothing is deferred to V2.

**Key Change in v1.1**: Incorporates hybrid architecture combining gonuget library (for proven components like config parsing, version comparison, solution/project parsing) with dotnet CLI (for safety-critical operations like package installation, removal, and restoration). LazyNuGet requires **gonuget >= v0.1.0** as a Go module dependency.

---

## Overview

The implementation is organized into 7 major tracks with 38 distinct specifications. Each specification represents an independently testable feature or component that contributes to the complete system.

**Total Specifications**: 38
**Estimated Timeline**: 24 weeks (full-time)
**No Deferred Features**: Complete implementation of proposal

### Prerequisites

LazyNuGet has the following dependencies:

1. **Go**: 1.24 or higher (language runtime)
2. **dotnet CLI**: .NET SDK 9.0 or higher (for NuGet operations)
3. **gonuget**: v0.1.0 or higher (Go library dependency for NuGet protocol components)
   - Repository: https://github.com/willibrandon/gonuget
   - Installation: `go get github.com/willibrandon/gonuget@v0.1.0`
   - Used for: Config parsing, version comparison, solution/project parsing

---

## gonuget Library Integration Strategy

LazyNuGet will use a **hybrid approach** combining gonuget library components with dotnet CLI execution for optimal safety and velocity.

### Dependency Management

LazyNuGet depends on **gonuget >= v0.1.0** as a standard Go module dependency:

```go
// go.mod
require github.com/willibrandon/gonuget v0.1.0
```

**gonuget v0.1.0 Scope**: Includes production-ready, battle-tested components:
- ✅ NuGet.Config parsing (hierarchical merging, source mappings, credentials)
- ✅ Version parsing & comparison (<20ns/op, SemVer 2.0 compliant)
- ✅ Solution parsing (.sln, .slnx, .slnf - refactored from CLI to library)
- ✅ Project file parsing (.csproj SDK-style and legacy)
- ✅ Data models (Package, Project, Dependency, Source types)
- ✅ Framework compatibility (TFM parsing and comparison)

**Future gonuget versions** will add more capabilities as they stabilize (search API v0.2.0+, dependency resolver v0.3.0+, package operations v0.4.0+).

### Why Hybrid Instead of All-or-Nothing

gonuget (github.com/willibrandon/gonuget) is a NuGet client library achieving protocol parity with the official .NET NuGet.Client. While certain components are battle-tested and production-ready (included in v0.1.0), other features remain in active development. The hybrid approach allows LazyNuGet to:

1. **Leverage Production-Ready Components**: Use proven parts (config parsing, version comparison) immediately
2. **Maintain Safety**: Keep critical operations (installation, restoration) with the official dotnet CLI
3. **Avoid Duplication**: Don't rebuild what's already working in gonuget
4. **Enable Dogfooding**: LazyNuGet becomes a real-world test bed for gonuget's stable components
5. **Support Incremental Migration**: Replace dotnet CLI calls with gonuget as components mature

### Integration Decision Matrix

| Component | Strategy | Rationale |
|-----------|----------|-----------|
| **NuGet.Config Parsing** | ✅ **Use gonuget** | Production-ready, handles hierarchical config merging, package source mappings, credentials |
| **Version Parsing & Comparison** | ✅ **Use gonuget** | Extremely fast (<20ns/op), handles SemVer 2.0, ranges, pre-release, wildcard, floating versions |
| **Data Models** | ✅ **Use gonuget types** | Proven structures for Package, Dependency, Source, avoiding duplication |
| **Solution/Project Parsing** | ✅ **Use gonuget** | Supports .sln, .slnx, .slnf, .csproj - already implemented |
| **Package Search (Read-Only)** | 🔄 **Hybrid with fallback** | Use gonuget if V3 search API is solid, otherwise fall back to dotnet CLI |
| **Package Metadata Fetching** | 🔄 **Hybrid with fallback** | Use gonuget for read-only metadata, fall back to dotnet CLI on errors |
| **Package Installation** | ❌ **Use dotnet CLI** | Critical operation - let official tool handle file writes, csproj mutations, lock files |
| **Package Removal** | ❌ **Use dotnet CLI** | Safety-critical - avoid breaking user builds |
| **Package Restore** | ❌ **Use dotnet CLI** | Complex dependency resolution with multi-targeting - don't risk edge cases |
| **Dependency Resolution (Display)** | ✅ **Use gonuget** | Read-only visualization of dependency trees is safe |

### Implementation Guidelines

1. **Add gonuget dependency**: `go get github.com/willibrandon/gonuget@v0.1.0`
2. **Import gonuget packages**: `import "github.com/willibrandon/gonuget/core/config"` (and other packages)
3. **Wrap gonuget APIs**: Create thin adapters in `internal/nuget/` that wrap gonuget calls
4. **Add fallback mechanisms**: For hybrid operations, catch gonuget errors and fall back to dotnet CLI
5. **Log strategy decisions**: Use debug logging to show when gonuget vs dotnet CLI is used
6. **Version upgrades**: As gonuget releases v0.2.0+, evaluate migrating more operations from dotnet CLI

### Example Integration Pattern

```go
// internal/nuget/config.go
import "github.com/willibrandon/gonuget/core/config"

func LoadNuGetConfig(path string) (*Config, error) {
    // Use gonuget's proven config parser
    cfg, err := config.Load(path)
    if err != nil {
        return nil, fmt.Errorf("loading nuget config: %w", err)
    }
    return adaptGonugetConfig(cfg), nil
}

// internal/nuget/operations.go
func InstallPackage(ctx context.Context, project, packageID, version string) error {
    // Use dotnet CLI for safety-critical operations
    cmd := exec.CommandContext(ctx, "dotnet", "add", project, "package", packageID, "-v", version)
    return cmd.Run()
}

// internal/nuget/search.go
import "github.com/willibrandon/gonuget/core/protocol"

func SearchPackages(ctx context.Context, query string) ([]Package, error) {
    // Try gonuget first (if search API is stable)
    results, err := protocol.Search(ctx, query)
    if err != nil {
        // Fallback to dotnet CLI
        logger.Debug("gonuget search failed, falling back to dotnet CLI: %v", err)
        return searchViaDotnetCLI(ctx, query)
    }
    return adaptGonugetResults(results), nil
}
```

### Benefits of This Approach

1. **Velocity**: No need to implement config parsing, version comparison, or solution parsing from scratch
2. **Correctness**: gonuget's proven components ensure NuGet protocol compliance
3. **Performance**: gonuget's optimizations (15-17x faster CLI, <20ns version parsing) improve UX
4. **Safety**: Critical operations stay with official dotnet CLI, avoiding risk of breaking user builds
5. **Evolution Path**: As gonuget components mature, replace dotnet CLI calls incrementally
6. **Reduced Maintenance**: Less code to maintain in LazyNuGet, fewer NuGet protocol bugs

---

## Track 1: Foundation Infrastructure (Specs 001-005)

### Spec 001: Application Bootstrap and Lifecycle

```bash
/speckit.specify Create the core application bootstrap system that initializes LazyNuGet, manages dependency injection, handles graceful shutdown, and provides application lifecycle management. This includes command-line argument parsing, version information display, and error handling for startup failures. The bootstrap must support running in normal mode and non-interactive mode for testing.
```

**Why**: Foundation for everything else. Nothing works without proper initialization.

**Key Requirements**:
- Parse CLI flags (--version, --help, --config, etc.)
- Initialize logging before any other operations
- Load configuration from multiple sources (defaults, files, env vars)
- Wire up dependency graph (app → gui → commands)
- Handle startup errors gracefully
- Clean shutdown with resource cleanup
- Signal handling (SIGINT, SIGTERM)

---

### Spec 002: Configuration Management System

```bash
/speckit.specify Implement the configuration management system that loads, merges, and validates configuration from multiple sources. Support hierarchical configuration with defaults, user config files (YAML/TOML), environment variables, and CLI flags. Configuration must include UI settings, keybindings, color schemes, performance tuning, and dotnet CLI paths. Must support hot-reload for user config changes.
```

**Why**: Configuration drives behavior across the entire application.

**Key Requirements**:
- Load defaults from embedded config
- Read user config from ~/.config/lazynuget/config.yml (Linux), ~/Library/Application Support/lazynuget/config.yml (macOS), %APPDATA%/lazynuget/config.yml (Windows)
- Override with environment variables (LAZYNUGET_*)
- Override with CLI flags
- Validate configuration schema
- Provide config migration for version changes
- Hot-reload on file change
- Export current config to file

---

### Spec 003: Cross-Platform Support Infrastructure

```bash
/speckit.specify Build cross-platform infrastructure that abstracts platform-specific behaviors for file paths, config locations, terminal detection, and OS conventions. Must handle Windows drive letters and UNC paths, Unix-style paths, XDG Base Directory Specification on Linux, and macOS conventions. Include terminal capability detection and graceful degradation for limited terminals.
```

**Why**: Core differentiator - identical experience on all platforms.

**Key Requirements**:
- Platform detection (Windows, macOS, Linux)
- Path normalization (forward/backward slashes, drive letters)
- Config directory resolution per platform
- Cache directory resolution per platform
- Terminal capability detection (color support, unicode, dimensions)
- Fallback modes for limited terminals
- Line ending handling (CRLF vs LF)
- Process spawning abstractions

---

### Spec 004: Error Handling and Logging Framework

```bash
/speckit.specify Create a logging and error handling framework that provides structured logging, error categorization, user-friendly error messages, and debugging support. Must support log levels (debug, info, warn, error), log to file with rotation, provide actionable error messages to users, and include context in error reports. Support opt-in error reporting and debugging modes.
```

**Why**: Essential for troubleshooting and user experience.

**Key Requirements**:
- Structured logging with levels
- Log file rotation (size and time-based)
- User-facing error messages (no stack traces)
- Detailed error messages in logs
- Error categorization (user error, config error, CLI error, bug)
- Suggestion system for common errors
- Debug mode with verbose output
- Panic recovery and graceful degradation

---

### Spec 005: Testing Infrastructure

```bash
/speckit.specify Establish comprehensive testing infrastructure including unit test framework, integration test framework, test fixtures, mocks for dotnet CLI, and cross-platform test execution. Must support table-driven tests, parallel test execution, test coverage reporting, and golden file testing for UI snapshots. Include CI/CD pipeline configuration for Linux, macOS, and Windows.
```

**Why**: Quality assurance and confidence in cross-platform behavior.

**Key Requirements**:
- Unit test framework with assertions
- Integration test framework with real dotnet CLI
- Mock dotnet CLI for fast tests
- Test fixtures for projects and solutions
- Cross-platform test execution (GitHub Actions)
- Test coverage reporting (>80% target)
- Golden file testing for UI
- Performance benchmarking tests

---

## Track 2: Domain Layer - dotnet CLI Integration (Specs 006-012)

**Integration Note**: Track 2 establishes the foundational integration layer. While originally designed around pure dotnet CLI integration, many specs now leverage gonuget library components for production-ready functionality (config parsing, version handling, solution/project parsing). The dotnet CLI command builder and executor remain essential for safety-critical operations (install, remove, restore).

### Spec 006: dotnet CLI Command Builder

```bash
/speckit.specify Implement the command builder pattern for constructing dotnet CLI commands with fluent API. Support all NuGet operations (list, add, remove, restore, search, why, locals) with conditional arguments, validation, and testability. Builder must validate arguments, provide defaults, and generate both command strings and exec.Cmd objects.
```

**Key Requirements**:
- Fluent API for command construction
- Type-safe argument building
- Validation of required arguments
- Conditional argument inclusion
- Output format selection (JSON, text)
- Verbosity control
- Timeout configuration
- Working directory specification

---

### Spec 007: dotnet CLI Executor and Output Capture

```bash
/speckit.specify Create the dotnet CLI executor that runs commands, captures output (stdout/stderr), handles errors, and provides progress feedback. Must support synchronous and asynchronous execution, streaming output, cancellation, and timeout. Include retry logic for transient failures and proper signal handling.
```

**Key Requirements**:
- Execute commands with timeout
- Capture stdout and stderr separately
- Stream output for long-running commands
- Parse exit codes and error conditions
- Cancellation support (context-based)
- Retry logic for network failures
- Progress reporting for long operations
- Environment variable injection

---

### Spec 008: JSON Output Parser

```bash
/speckit.specify Build JSON output parser for dotnet CLI commands that deserialize structured output into Go types. Support parsing list package output, search results, why dependency chains, and handle version differences in JSON schema. Include validation and error recovery for malformed JSON.
```

**Key Requirements**:
- Parse `dotnet list package --format json`
- Parse `dotnet package search --format json`
- Parse `dotnet nuget why` output
- Handle JSON schema versioning
- Validate JSON structure
- Error recovery for partial JSON
- Map to internal models

---

### Spec 009: Project and Solution File Parser

```bash
/speckit.specify Implement project and solution file parsing using gonuget's proven parsers that handle .csproj (SDK-style and legacy), .sln, .slnx (VS2022+), and .slnf (solution filters). Extract PackageReference elements, target frameworks, project properties, solution structure, and project relationships. For write operations, use dotnet CLI commands (dotnet sln add, dotnet add package) to avoid corrupting files.
```

**Integration Strategy**: ✅ **Use gonuget for parsing** (read-only), ❌ **Use dotnet CLI for modifications**

**Key Requirements**:
- Parse SDK-style projects (gonuget)
- Parse legacy project files (gonuget)
- Parse .sln solutions (gonuget)
- Parse .slnx solutions (gonuget)
- Parse .slnf solution filters (gonuget)
- Extract PackageReference elements
- Extract TargetFramework(s)
- Extract project properties
- Extract solution project hierarchy
- Modifications via dotnet CLI (safety)
- Adapter layer for gonuget types

---

### Spec 010: NuGet.Config Parser

```bash
/speckit.specify Integrate gonuget's production-ready NuGet.Config parser that handles package sources, credentials, disabled sources, fallback folders, package source mappings, and audit sources. Support hierarchical config merging (machine, user, repo-level) which gonuget already implements. For write operations (add/remove sources, enable/disable), use gonuget's config writer or dotnet CLI depending on maturity. Handle encrypted credentials appropriately.
```

**Integration Strategy**: ✅ **Use gonuget exclusively** (production-ready, battle-tested)

**Why gonuget**: This is explicitly called out as rock-solid and production-ready. gonuget's config parser handles all NuGet.Config complexity including hierarchical merging, package source mappings, and credential handling.

**Key Requirements**:
- Parse packageSources section (gonuget)
- Parse disabledPackageSources section (gonuget)
- Parse packageSourceCredentials (gonuget, warn on clear-text)
- Parse packageSourceMapping (gonuget)
- Parse fallbackPackageFolders (gonuget)
- Parse auditSources (gonuget)
- Hierarchical config resolution (gonuget)
- Write source modifications (gonuget or dotnet CLI)
- Validate config structure (gonuget)
- Adapter layer for gonuget config types

---

### Spec 011: Data Models (Package, Project, Dependency)

```bash
/speckit.specify Define core data models by wrapping or extending gonuget's proven types for Package, Project, Dependency, Source, and PackageReference. Add LazyNuGet-specific extensions for UI state (selection, display flags, transient state). Create adapter functions to convert between gonuget types and LazyNuGet UI models. Include validation rules, serialization support, and lifecycle management for UI operations.
```

**Integration Strategy**: ✅ **Leverage gonuget types as foundation**, extend with UI-specific fields

**Why gonuget**: Reusing gonuget's data models ensures NuGet protocol compatibility and avoids duplicating complex validation logic for version ranges, framework identifiers, etc.

**Key Requirements**:
- Wrap gonuget Package types (ID, version, metadata)
- Wrap gonuget Project types (path, frameworks, references)
- Wrap gonuget Dependency types (version range, transitive)
- Wrap gonuget Source types (name, URL, enabled)
- Add UI-specific fields (selected, display state, icons)
- Add Vulnerability model (CVE, severity, advisory) if not in gonuget
- Model validation (leverage gonuget validators)
- JSON serialization/deserialization
- Adapter functions (gonuget ↔ LazyNuGet)
- Equality and comparison methods

---

### Spec 012: Package and Project Loaders

```bash
/speckit.specify Build loader components that fetch and construct Package and Project models using hybrid approach: gonuget parsers for file reading (.csproj, NuGet.Config), gonuget protocol clients for NuGet.org API metadata (if stable), and dotnet CLI output as fallback. Support parallel loading, caching (leverage gonuget's multi-tier cache), incremental updates, and error handling. Loaders must enrich models with metadata from multiple sources.
```

**Integration Strategy**: 🔄 **Hybrid** - gonuget for file parsing, gonuget/dotnet CLI for API calls depending on stability

**Key Requirements**:
- Load projects from .csproj files (gonuget parser)
- Load solutions from .sln/.slnx/.slnf files (gonuget parser)
- Load packages from `dotnet list package` (parse JSON output)
- Load package metadata from NuGet.org API (try gonuget, fallback to dotnet CLI)
- Load vulnerability data (dotnet CLI or NuGet API)
- Parallel loading for performance
- Caching with TTL (leverage gonuget's cache if available)
- Incremental updates
- Error handling and partial results
- Fallback strategy logging

---

## Track 3: Domain Layer - Operations (Specs 013-018)

**Integration Note**: Track 3 implements package operations. Version comparison (Spec 013) uses gonuget exclusively for performance and correctness. All write operations (add, remove, update, restore) use dotnet CLI for safety. Cache management uses dotnet CLI since it's primarily a file system operation.

### Spec 013: Version Comparison and Resolution

```bash
/speckit.specify Integrate gonuget's high-performance version parsing and comparison library that handles NuGet-compliant semantic versioning, pre-release versions, version ranges ([1.0,2.0)), wildcard versions (1.0.*), and floating versions (1.0.0-*). Use gonuget's version utilities for determining update severity (major/minor/patch). This is a performance-critical component (target: <20ns/op for parsing, <10ns/op for comparison).
```

**Integration Strategy**: ✅ **Use gonuget exclusively** (<20ns/op performance, battle-tested)

**Why gonuget**: Version parsing and comparison is extremely performance-critical (executed for every package in every comparison operation). gonuget's implementation is <20ns/op for parsing and handles all NuGet version semantics correctly. This is faster than .NET's own implementation.

**Key Requirements**:
- Parse NuGet version strings (gonuget, <20ns/op)
- Compare versions (>, <, ==) (gonuget, <10ns/op)
- Parse version ranges (gonuget)
- Check version satisfies range (gonuget)
- Determine update severity (major/minor/patch) (gonuget or build on top)
- Handle pre-release versions (gonuget)
- Handle wildcard versions (1.0.*) (gonuget)
- Handle floating versions (1.0.0-*) (gonuget)
- Adapter layer for gonuget version types
- Performance benchmarks to verify <20ns/op target

---

### Spec 014: Cache Manager

```bash
/speckit.specify Create cache manager that interacts with NuGet's global-packages, http-cache, temp, and plugins-cache folders. Support listing cached packages, clearing specific packages or entire caches, calculating cache sizes, and detecting stale cache entries. Must work cross-platform with appropriate path resolution.
```

**Key Requirements**:
- List cache locations via `dotnet nuget locals`
- List cached packages in global-packages
- Calculate cache sizes
- Clear specific package versions
- Clear entire caches
- Detect stale entries
- Cross-platform path handling
- Provide cache statistics

---

### Spec 015: Package Add Operation

```bash
/speckit.specify Implement the add package operation using dotnet CLI (dotnet add package) for safety. Validate inputs using gonuget version parsing and project file parsing, but delegate the actual modification to dotnet CLI to ensure lock file consistency and avoid corrupting project files. Support version specification, framework-specific references, pre-release packages, and source selection. Include validation, conflict detection (using gonuget dependency resolver for display), backup/rollback, and user confirmation for large dependency trees.
```

**Integration Strategy**: ❌ **Use dotnet CLI for writes**, ✅ **Use gonuget for validation/display**

**Why dotnet CLI**: Package installation is safety-critical. The official tool handles csproj mutations, lock file updates, and transitive dependency resolution correctly. We validate and display using gonuget, but execute via dotnet CLI.

**Key Requirements**:
- Execute via `dotnet add package` (CLI)
- Validate version using gonuget version parser
- Validate package exists (gonuget API or CLI)
- Check for conflicts (gonuget dependency resolver for preview)
- Preview dependency tree (gonuget)
- Framework-specific addition (CLI)
- Pre-release package support (CLI)
- Source specification (CLI)
- Backup project file
- Optional auto-restore (CLI)
- Rollback on failure

---

### Spec 016: Package Remove Operation

```bash
/speckit.specify Implement the remove package operation using dotnet CLI (dotnet remove package) for safety. Use gonuget dependency resolver to detect reverse dependencies and warn about potential breaks. Provide preview of what will be removed using gonuget's dependency graph visualization. Include backup/rollback. Support force removal and cascade removal of unused transitive dependencies.
```

**Integration Strategy**: ❌ **Use dotnet CLI for writes**, ✅ **Use gonuget for validation/display**

**Key Requirements**:
- Execute via `dotnet remove package` (CLI)
- Check for reverse dependencies (gonuget resolver)
- Warn about potential breaks (gonuget analysis)
- Preview removal impact (gonuget dependency graph)
- Backup project file
- Optional cascade removal (CLI or gonuget-guided)
- Force removal option (CLI)
- Rollback on failure

---

### Spec 017: Package Update Operation

```bash
/speckit.specify Create package update operation using dotnet CLI (dotnet add package with --version) for execution. Use gonuget for version comparison to determine latest stable/pre-release and update severity (major/minor/patch). Use gonuget dependency resolver to detect conflicts and warn about breaking changes. Support bulk update across multiple projects with coordinated version selection.
```

**Integration Strategy**: ❌ **Use dotnet CLI for writes**, ✅ **Use gonuget for version logic and validation**

**Key Requirements**:
- Execute via `dotnet add package --version` (CLI)
- Determine latest stable (gonuget version comparison)
- Determine latest pre-release (gonuget)
- Update to specific version (CLI with gonuget validation)
- Update within constraints (gonuget version satisfies)
- Detect version conflicts (gonuget resolver)
- Warn about breaking changes (gonuget version diff: major bump)
- Multi-project bulk update (orchestrate CLI calls)
- Preview before updating (gonuget analysis)
- Backup and rollback
- Update lock files (CLI handles automatically)

---

### Spec 018: Package Restore Operation

```bash
/speckit.specify Implement package restore operation using dotnet CLI (dotnet restore) exclusively. This is a complex operation involving dependency resolution, framework targeting, and lock file generation - best left to the official tool. Focus on progress tracking, error parsing, and user-friendly error messages for common failures (network issues, missing packages, authentication errors).
```

**Integration Strategy**: ❌ **Use dotnet CLI exclusively** (too complex and critical for gonuget)

**Why dotnet CLI**: Restore involves complex multi-targeting dependency resolution, lock file generation, and NuGet cache interaction. This is where edge cases live. Let the official tool handle it.

**Key Requirements**:
- Execute `dotnet restore` (CLI)
- Force restore option (CLI --force)
- No-cache restore (CLI --no-cache)
- Source specification (CLI --source)
- Parallel restore for solutions (orchestrate CLI calls)
- Progress tracking (parse CLI output)
- Detailed error messages (parse and enhance CLI errors)
- Retry on transient failures
- Detect common error patterns and suggest fixes

---

## Track 4: GUI Framework (Specs 019-025)

### Spec 019: Bubbletea Application Architecture

```bash
/speckit.specify Build the Bubbletea application scaffold implementing the Elm architecture (Model, Update, View). Create main application model, message routing, command batching, and integration with panels. Support window resize handling, graceful error display, and startup/shutdown lifecycle integration with the bootstrap layer.
```

**Key Requirements**:
- Implement tea.Model interface
- Define message types
- Route messages to panels
- Batch commands
- Handle window resize
- Handle errors in UI
- Initialize from bootstrap
- Clean shutdown

---

### Spec 020: Layout Manager and Panel System

```bash
/speckit.specify Create layout manager that arranges panels in responsive layouts based on terminal size. Support fixed layouts (header, sidebar, main, status), adaptive layouts for different terminal sizes (80x24, 120x30, 160x40+), panel focus management, and smooth transitions. Include panel base types and composition patterns.
```

**Key Requirements**:
- Header bar layout
- Sidebar + main panel layout
- Split main panel layout
- Status bar layout
- Responsive sizing
- Minimum size handling (80x24)
- Panel focus management
- Panel composition
- Smooth transitions

---

### Spec 021: Keybinding System and Input Handling

```bash
/speckit.specify Implement keybinding system that maps keys to actions with context-awareness and conflicts resolution. Support global keybindings, panel-specific keybindings, configurable keybindings, key sequences, and disabled action hints. Include vim-style navigation (hjkl) alongside arrow keys and provide help screen generation from keybindings.
```

**Key Requirements**:
- Define keybinding types
- Global keybindings (?, q, r, /)
- Panel-specific keybindings
- Configurable keybindings
- Key conflict detection
- Disabled action reasons
- Help screen generation
- Vim-style navigation
- Key sequence support

---

### Spec 022: Status Bar and Help System

```bash
/speckit.specify Build status bar component that displays context-sensitive keybindings and mode indicators. Support dynamic keybinding display based on focused panel and active modes, truncation for narrow terminals, and color-coded hints. Include help system with searchable keybinding reference, tips, and quick start guide.
```

**Key Requirements**:
- Display active keybindings
- Context-sensitive display
- Mode indicators
- Truncate for narrow terminals
- Color-coded hints
- Help screen (?) with all keybindings
- Searchable help
- Quick start guide

---

### Spec 023: Color Scheme and Theming System

```bash
/speckit.specify Create color scheme system using lipgloss that provides semantic colors (success, warning, error, info), adapts to terminal capabilities, and supports user themes. Include default theme, high-contrast theme, monochrome theme, and custom theme support. Must respect terminal's base colors and provide graceful degradation for limited color terminals.
```

**Key Requirements**:
- Define semantic colors
- Default theme
- High-contrast theme
- Monochrome theme
- Custom theme support
- Terminal capability detection
- Graceful degradation
- User theme configuration
- Color preview

---

### Spec 024: Modal Dialog System

```bash
/speckit.specify Implement modal dialog system for user input, confirmations, selections, and text entry. Support prompt dialogs, confirmation dialogs (yes/no), selection dialogs (list), text input dialogs, and progress dialogs. Include keyboard navigation, validation, and proper focus management that returns to previous panel after dismissal.
```

**Key Requirements**:
- Prompt dialog (message + ok)
- Confirmation dialog (yes/no)
- Selection dialog (list of options)
- Text input dialog (single line)
- Multi-line text dialog
- Progress dialog (with cancel)
- Keyboard navigation
- Input validation
- Focus management

---

### Spec 025: Component Library (Lists, Tables, Trees)

```bash
/speckit.specify Build reusable UI component library including filterable lists, sortable tables, expandable trees, viewports, and progress indicators. Components must support keyboard navigation, selection, filtering, sorting, pagination, lazy loading, and custom rendering. Include components for common patterns in bubbles library.
```

**Key Requirements**:
- List component (filterable, selectable)
- Table component (sortable, multi-column)
- Tree component (expandable, collapsible)
- Viewport component (scrollable content)
- Progress bar component
- Spinner component
- Text input component
- Textarea component

---

## Track 5: Core Features (Specs 026-033)

### Spec 026: Project Explorer Panel

```bash
/speckit.specify Build project explorer panel that displays solution and project hierarchy as an expandable tree. Support navigation with keyboard (hjkl/arrows), expand/collapse, project selection, display of framework badges and package counts, and integration with file watching for real-time updates. Handle solutions, folder structures, and standalone projects.
```

**Key Requirements**:
- Display solution hierarchy
- Display standalone projects
- Expand/collapse folders
- Keyboard navigation
- Project selection
- Framework badges
- Package count display
- File watching integration
- Empty state handling
- Error state display

---

### Spec 027: Package List Panel

```bash
/speckit.specify Create package list panel that displays installed packages in a sortable, filterable table. Support columns (name, current version, latest version, status), status icons (up-to-date, outdated, vulnerable, deprecated), filtering (all, direct, transitive, outdated, vulnerable), sorting (name, version, status), and selection for detail view. Include keyboard shortcuts for quick actions.
```

**Key Requirements**:
- Table view of packages
- Columns: name, version, latest, status
- Status icons and colors
- Filter by type (direct, transitive)
- Filter by status (outdated, vulnerable)
- Sort by any column
- Keyboard navigation
- Quick actions (u=update, d=remove)
- Selection for details
- Empty state handling

---

### Spec 028: Package Details Panel with Tabs

```bash
/speckit.specify Implement package details panel with tabbed interface showing metadata, dependencies, versions, and readme. Metadata tab displays author, license, downloads, description, and URLs. Dependencies tab shows dependency tree. Versions tab lists available versions with release dates. Readme tab renders markdown from package. Support tab switching with keyboard and mouse.
```

**Key Requirements**:
- Tabbed interface (←/→ to switch)
- Metadata tab (author, license, downloads, description)
- Dependencies tab (tree view)
- Versions tab (list with dates)
- Readme tab (markdown rendering)
- Scrollable content
- Loading states
- Error states (package not found)
- Cache metadata

---

### Spec 029: Package Search Interface

```bash
/speckit.specify Build package search interface with hybrid search backend: try gonuget's V3 protocol search API first for performance (if stable), fall back to dotnet CLI (dotnet package search) on errors. Provide fuzzy search input and live results display. Support searching across configured sources, filtering by pre-release, sorting by relevance/downloads, preview of results, and installation from search. Include autocomplete (using gonuget's autocomplete API if available), search history, and quick package details view.
```

**Integration Strategy**: 🔄 **Hybrid with fallback** - gonuget search API (fast) with dotnet CLI fallback (safe)

**Why Hybrid**: Package search is read-only and performance-critical for UX. If gonuget's V3 search implementation is stable, use it for speed (15-17x faster). If it errors, fall back to dotnet CLI silently. Log which backend is used for debugging.

**Key Requirements**:
- Search backend: try gonuget, fallback to `dotnet package search` (CLI)
- Search input with autocomplete (gonuget API or none)
- Live results as you type
- Fuzzy matching (client-side)
- Search across sources (gonuget multi-source or CLI)
- Filter: include pre-release (gonuget or CLI --prerelease)
- Sort: relevance, downloads, name (gonuget or parse CLI)
- Result preview (gonuget metadata or CLI)
- Quick install (enter) - delegates to Spec 015
- Search history (local storage)
- Cancel search (esc)
- Log backend selection for debugging

---

### Spec 030: Update Manager Panel

```bash
/speckit.specify Create update manager panel that displays all outdated packages across solution with bulk update capabilities. Show current and latest versions, update severity (major/minor/patch), breaking change indicators, and allow selective or bulk updates. Include preview of update impact, filtering by severity, and multi-project coordination.
```

**Key Requirements**:
- List all outdated packages
- Show current and latest versions
- Indicate update severity
- Mark breaking changes (major)
- Select packages for update
- Bulk update all
- Update by severity (patch only, etc.)
- Preview update impact
- Multi-project updates
- Progress tracking

---

### Spec 031: Basic Operations Integration

```bash
/speckit.specify Integrate all basic package operations (add, remove, update, restore, clear cache) into the GUI with proper user flows, confirmation dialogs, progress feedback, and error handling. Ensure operations work across single projects and multiple projects, provide undo capability for file modifications, and display detailed operation logs.
```

**Key Requirements**:
- Add package flow (search → select → add)
- Remove package flow (select → confirm → remove)
- Update package flow (select → choose version → update)
- Restore flow (trigger → progress → complete)
- Clear cache flow (confirm → clear → report)
- Multi-project operations
- Progress indicators
- Error handling and display
- Undo capability
- Operation logs

---

### Spec 032: File Watching and Auto-Refresh

```bash
/speckit.specify Implement file watching system using fsnotify that monitors .csproj files, packages.lock.json, project.assets.json, and NuGet.Config for changes. Trigger auto-refresh of affected panels when files change externally. Support debouncing to avoid excessive refreshes, user notification of external changes, and option to disable auto-refresh.
```

**Key Requirements**:
- Watch .csproj files
- Watch packages.lock.json
- Watch project.assets.json
- Watch NuGet.Config
- Debounce rapid changes (300ms)
- Trigger panel refresh
- Notify user of external changes
- Option to disable
- Handle watch errors

---

### Spec 033: Background Task Management

```bash
/speckit.specify Build background task management system that executes long-running operations without blocking UI. Support cancellable tasks, progress reporting, task queuing, and concurrent task limits. Include visual feedback for running tasks, task history, and error recovery. Use Go's context for cancellation and worker pools for concurrency control.
```

**Key Requirements**:
- Execute tasks in background
- Cancellation support (context)
- Progress reporting to UI
- Task queue
- Concurrent task limits
- Visual task indicator
- Task history/logs
- Error recovery
- Worker pool management

---

## Track 6: Advanced Features (Specs 034-038)

### Spec 034: Vulnerability Dashboard and Scanning

```bash
/speckit.specify Create vulnerability dashboard that displays all vulnerable packages across solution with severity levels, CVE IDs, affected versions, and remediation advice. Support scanning on demand, automatic scanning on restore, filtering by severity, and direct update to fixed versions. Integrate with dotnet CLI vulnerability detection and NuGet advisory database.
```

**Key Requirements**:
- List vulnerable packages
- Display CVE IDs
- Show severity (critical, high, medium, low)
- Display affected versions
- Show remediation advice
- Link to advisories
- Filter by severity
- Direct update to fixed version
- Scan on demand
- Auto-scan on restore

---

### Spec 035: Dependency Graph Visualization

```bash
/speckit.specify Implement interactive dependency graph visualization using gonuget's dependency resolver for tree construction and conflict detection. Display package dependency trees with expand/collapse, highlighting of transitive dependencies, path finding (why is this package included), and conflict identification using gonuget's resolution engine. Support multiple visualization modes (tree, graph), export to graphviz, and integration with 'dotnet nuget why' command for validation.
```

**Integration Strategy**: ✅ **Use gonuget dependency resolver** (read-only, safe visualization)

**Why gonuget**: Dependency graph construction is read-only and gonuget's resolver provides structured data for visualization. Safer than parsing dotnet CLI text output. Validate results against 'dotnet nuget why' for correctness.

**Key Requirements**:
- Build dependency tree (gonuget resolver)
- Display dependency tree (UI)
- Expand/collapse nodes (UI state)
- Highlight transitive dependencies (gonuget marks direct vs transitive)
- Show version constraints (gonuget version ranges)
- Path finding (gonuget resolver: why package included)
- Identify conflicts (gonuget resolver)
- Tree view mode (UI)
- Graph view mode if feasible (UI)
- Export to graphviz (format gonuget tree)
- Cross-validate with 'dotnet nuget why' (CLI)

---

### Spec 036: Package Source Manager

```bash
/speckit.specify Build package source manager using gonuget's NuGet.Config parser (Spec 010) for reading sources, and gonuget's config writer (if stable) or dotnet CLI for modifications. Display configured NuGet sources with enable/disable toggle, add/remove sources, test connectivity (gonuget protocol test or CLI), and manage credentials. Support source priority ordering, package source mapping, and warning for insecure sources (HTTP). Include import/export of source configurations.
```

**Integration Strategy**: ✅ **Use gonuget for reading**, 🔄 **Hybrid for writing** (gonuget writer if stable, else CLI)

**Why gonuget**: Source management is built on NuGet.Config which gonuget parses perfectly (Spec 010). For modifications, use gonuget's writer if it preserves formatting, otherwise use dotnet CLI to be safe.

**Key Requirements**:
- List configured sources (gonuget config parser)
- Parse package source mappings (gonuget)
- Enable/disable sources (gonuget writer or CLI)
- Add new source (gonuget writer or CLI)
- Remove source (gonuget writer or CLI)
- Edit source (gonuget writer or CLI)
- Test source connectivity (gonuget protocol test or ping)
- Manage credentials (warn on clear-text) (gonuget)
- Source priority ordering (gonuget or manual reorder + write)
- Import/export sources (gonuget read + write)
- Warn about HTTP sources (check URL scheme)

---

### Spec 037: Cache Browser and Management

```bash
/speckit.specify Create cache browser that displays contents of global-packages cache with package list, version list, size information, and last accessed time. Support browsing packages, viewing package details, clearing individual packages, clearing by pattern, and calculating total cache size. Include cache validation and cleanup of corrupted packages.
```

**Key Requirements**:
- List cached packages
- Show package versions
- Display sizes
- Show last accessed
- Browse package contents
- Clear individual packages
- Clear by pattern (e.g., all pre-release)
- Clear all cache
- Validate cache integrity
- Repair corrupted cache

---

### Spec 038: Advanced Features (Conflict Resolver, License Auditor, Custom Commands)

```bash
/speckit.specify Implement three advanced features: (1) Version conflict resolver that detects conflicts and suggests resolutions, (2) License auditor that scans package licenses and flags incompatibilities, (3) Custom commands system that allows users to define reusable command sequences via configuration. Each must integrate seamlessly with existing panels.
```

**Key Requirements**:

**Conflict Resolver**:
- Detect version conflicts (gonuget dependency resolver)
- Explain conflict cause (gonuget constraint analysis)
- Suggest resolutions (gonuget resolution strategies)
- Apply resolution (dotnet CLI for actual package updates)
- Preview impact (gonuget dependency tree diff)

**License Auditor**:
- Scan package licenses
- Display license types
- Flag incompatible licenses
- Export license report
- License compatibility matrix

**Custom Commands**:
- Define commands in config
- Command templates (Go templates)
- Keybinding assignment
- Context-specific commands
- Command execution
- Command history

---

## Execution Strategy

### Phase Sequencing

**Phase 1: Foundation (Weeks 1-4)**
- Execute Track 1 (Specs 001-005) to establish infrastructure
- Goal: Working application that starts, loads config, handles errors

**Phase 2: Domain Layer (Weeks 5-10)**
- Execute Track 2 (Specs 006-012) for dotnet CLI integration
- Execute Track 3 (Specs 013-018) for operations
- Goal: Complete domain layer that can execute all NuGet operations

**Phase 3: GUI Framework (Weeks 11-14)**
- Execute Track 4 (Specs 019-025) for Bubbletea framework
- Goal: Working GUI shell with layout, keybindings, components

**Phase 4: Core Features (Weeks 15-19)**
- Execute Track 5 (Specs 026-033) for main features
- Goal: Full MVP with project explorer, package list, search, operations

**Phase 5: Advanced Features (Weeks 20-24)**
- Execute Track 6 (Specs 034-038) for advanced features
- Goal: Complete feature set including vulnerabilities, dependencies, custom commands

### Parallel Execution

Many specifications can be implemented in parallel:
- **Track 2 & 3** (Domain) can be developed concurrently
- **Track 4** (GUI) can start once Track 1 is stable
- **Track 5** (Features) can begin once Tracks 2, 3, 4 are complete
- **Track 6** (Advanced) depends on Track 5 completion

### Testing Integration

- Each specification includes acceptance criteria and testing requirements
- Unit tests developed alongside implementation (TDD approach)
- Integration tests added once domain layer is complete
- Cross-platform testing throughout all phases
- No specification is "complete" until tests pass on Linux, macOS, Windows

---

## Command Reference

### To Begin Implementation

Start with Track 1, Spec 001:

```bash
/speckit.specify Create the core application bootstrap system that initializes LazyNuGet, manages dependency injection, handles graceful shutdown, and provides application lifecycle management. This includes command-line argument parsing, version information display, and error handling for startup failures. The bootstrap must support running in normal mode and non-interactive mode for testing.
```

Then proceed through each specification in sequence, completing tracks before moving to dependent tracks.

### Verification Commands

After each specification, run:
```bash
/speckit.plan     # Generate implementation plan
/speckit.tasks    # Break down into actionable tasks
/speckit.analyze  # Verify cross-artifact consistency
/speckit.implement # Execute implementation
```

---

## Success Criteria

Implementation is complete when:

1. ✅ All 38 specifications implemented and tested
2. ✅ All acceptance criteria from proposal met
3. ✅ Test coverage >80% (per constitution)
4. ✅ Cross-platform tests passing on Linux, macOS, Windows
5. ✅ Performance targets met (<200ms startup, 60 FPS, <50MB memory)
6. ✅ All features from proposal functional
7. ✅ Documentation complete (README, ARCHITECTURE, KEYBINDINGS, CONFIGURATION)
8. ✅ No known critical or high-severity bugs

---

## Notes

- **Nothing Deferred**: Every feature in the proposal is included
- **Production Ready**: Each spec includes production-quality requirements
- **Testable**: Each spec is independently testable
- **Cross-Platform**: All specs include cross-platform requirements
- **Constitutional Compliance**: All specs align with the 8 core principles (including Principle VIII: Complete Implementation - No Deferrals)
- **gonuget Dependency**: Requires gonuget >= v0.1.0 as a standard Go module dependency
- **Hybrid Architecture**: Combines gonuget library (v0.1.0: config, versions, parsing) with dotnet CLI (safety-critical: install, remove, restore) for optimal velocity and safety
- **gonuget Dogfooding**: LazyNuGet serves as a real-world test bed for gonuget's stable components while maintaining safety through dotnet CLI for critical operations
- **Incremental Migration**: As gonuget releases v0.2.0+ with additional stable components, dotnet CLI calls can be replaced incrementally without architectural changes
- **Version Evolution Path**: v0.1.0 (parsing/config) → v0.2.0 (search) → v0.3.0 (dependency resolver) → v0.4.0+ (operations) → v1.0.0 (full client)

---

**End of Implementation Plan**
