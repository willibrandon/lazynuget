# Contract: Process Spawning

## Struct: ProcessResult

Contains the output and exit status of a completed process.

```go
package contracts

// ProcessResult contains the output and exit status of a process
type ProcessResult struct {
	Stdout   string // Standard output (decoded to UTF-8)
	Stderr   string // Standard error (decoded to UTF-8)
	ExitCode int    // Process exit code (0 = success)
}
```

### Fields

#### Stdout (string)

Standard output captured from the process, decoded to UTF-8.

**Encoding handling**:
1. Attempt UTF-8 decode (validate with `utf8.Valid()`)
2. If invalid, fall back to system-detected encoding (per FR-030):
   - **Windows**: Use code page from `GetACP()` or `GetConsoleOutputCP()`
   - **Unix**: Parse locale from `LC_ALL`, `LC_CTYPE`, or `LANG` environment variables
3. If both fail, replace invalid bytes with `�` (U+FFFD REPLACEMENT CHARACTER)

**Line endings**: Preserved as-is (CRLF on Windows, LF on Unix) - caller can normalize if needed.

#### Stderr (string)

Standard error captured from the process, decoded to UTF-8 (same encoding handling as Stdout).

**Note**: Both stdout and stderr are captured separately and returned in decoded form.

#### ExitCode (int)

Process exit code as defined by the operating system.

**Standard values**:
- `0` - Success
- `1` - General error
- `2` - Misuse of shell command
- `126` - Command cannot execute (permissions)
- `127` - Command not found
- `130` - Terminated by Ctrl+C (SIGINT)

**Note**: Exit code semantics are application-specific. For `dotnet` CLI:
- `0` - Success
- Non-zero - Operation failed (check stderr for details)

## Interface: ProcessSpawner

Handles platform-specific process execution with automatic encoding detection and output capture.

```go
// ProcessSpawner handles platform-specific process execution
type ProcessSpawner interface {
	// Run executes a process and waits for completion
	// Automatically handles:
	// - PATH resolution for executable
	// - Argument quoting for paths with spaces
	// - Output encoding detection and conversion to UTF-8
	// - Exit code extraction
	Run(executable string, args []string, workingDir string, env map[string]string) (ProcessResult, error)

	// SetEncoding overrides automatic encoding detection
	// Use "utf-8", "windows-1252", "iso-8859-1", etc.
	// Pass empty string to re-enable auto-detection
	SetEncoding(encoding string)
}
```

## Methods

### Run(executable, args, workingDir, env) (ProcessResult, error)

Executes a process and waits for completion, capturing all output.

**Parameters**:

- **executable** (string): Command to execute
  - Bare command name: `"dotnet"` → searches PATH
  - Absolute path: `"/usr/bin/dotnet"` → executes directly
  - **Windows**: Automatically tries `.exe`, `.cmd`, `.bat` extensions
  - **Unix**: Must have execute permissions

- **args** ([]string): Command arguments
  - Each argument is a separate array element
  - Paths with spaces are automatically quoted (platform-specific rules)
  - **Windows**: Uses `CommandLineToArgvW` quoting rules
  - **Unix**: Shell-escapes spaces and special characters

- **workingDir** (string): Working directory for the process
  - Must exist (validated before execution)
  - Empty string uses current directory
  - **Windows**: Can be UNC path (`\\server\share`)
  - **Unix**: Can be symlink (resolved to target)

- **env** (map[string]string): Environment variables
  - Merged with parent process environment
  - Pass `nil` to inherit all parent env vars
  - Keys must not contain `=` or null bytes
  - Values can be empty strings (unsets variable on some platforms)

**Returns**:

- **ProcessResult**: Captured output and exit code (see struct above)
- **error**: Non-nil if process failed to start or encoding failed
  - Process execution errors (command not found, permissions, etc.)
  - Encoding detection/conversion errors
  - Working directory errors

**Behavior**:

1. **Executable resolution**:
   - If absolute path → validate it exists
   - If bare name → search PATH directories
   - **Windows**: Try `.exe`, `.cmd`, `.bat` extensions in order
   - **Unix**: Check execute permission (`os.Stat` + mode check)

2. **Argument quoting**:
   - Detect paths with spaces
   - Apply platform-specific quoting rules
   - Preserve special characters in non-path arguments

3. **Process execution**:
   - Set working directory
   - Merge environment variables
   - Spawn process with `os/exec`
   - Capture stdout and stderr separately

4. **Output encoding** (per FR-030):
   - Try UTF-8 decode first
   - Fall back to system encoding if invalid
   - **Windows**: Query code page via `GetACP()` or `GetConsoleOutputCP()`
   - **Unix**: Parse `LC_ALL`, `LC_CTYPE`, or `LANG` (e.g., `en_US.UTF-8` → UTF-8)
   - Replace invalid bytes with `�` if both fail

5. **Exit code extraction**:
   - Wait for process to complete
   - Extract exit code (platform-specific)
   - Return exit code in `ProcessResult`

**Example**:

```go
spawner := platform.NewProcessSpawner()

result, err := spawner.Run(
    "dotnet",
    []string{"restore", "/path/with spaces/project.csproj"},
    "/working/dir",
    map[string]string{
        "DOTNET_CLI_TELEMETRY_OPTOUT": "1",
    },
)

if err != nil {
    return fmt.Errorf("failed to spawn dotnet: %w", err)
}

if result.ExitCode != 0 {
    return fmt.Errorf("dotnet restore failed: %s", result.Stderr)
}

fmt.Println(result.Stdout)
```

**Error cases**:

```go
// Executable not found
Run("nonexistent", nil, "", nil)
// Returns: error "executable not found in PATH: nonexistent"

// Working directory doesn't exist
Run("dotnet", []string{"--version"}, "/nonexistent", nil)
// Returns: error "working directory does not exist: /nonexistent"

// Permission denied (Unix)
Run("/path/to/script.sh", nil, "", nil)
// Returns: error "permission denied: /path/to/script.sh (not executable)"
```

### SetEncoding(encoding string)

Overrides automatic encoding detection for process output.

**Parameters**:

- **encoding** (string): Encoding name or empty string
  - Standard names: `"utf-8"`, `"windows-1252"`, `"iso-8859-1"`, `"shift-jis"`, etc.
  - Empty string: Re-enables automatic detection (default)
  - Case-insensitive

**Supported encodings** (via `golang.org/x/text/encoding`):

| Encoding | Common Names | Use Case |
|----------|--------------|----------|
| UTF-8 | `utf-8`, `utf8` | Default, modern systems |
| Windows-1252 | `windows-1252`, `cp1252` | Western European (Windows) |
| ISO-8859-1 | `iso-8859-1`, `latin1` | Western European (Unix) |
| Shift-JIS | `shift-jis`, `sjis` | Japanese (Windows) |
| GB18030 | `gb18030`, `gbk` | Simplified Chinese |
| EUC-KR | `euc-kr` | Korean |

**Usage**:

```go
spawner := platform.NewProcessSpawner()

// Override encoding for Japanese Windows system
spawner.SetEncoding("shift-jis")

result, _ := spawner.Run("some-command", nil, "", nil)
// Output is decoded from Shift-JIS to UTF-8

// Re-enable auto-detection
spawner.SetEncoding("")

result2, _ := spawner.Run("another-command", nil, "", nil)
// Output encoding auto-detected (UTF-8 first, then system encoding)
```

**When to use**:

- Known non-UTF8 process output (legacy tools, specific locales)
- Encoding auto-detection fails or produces incorrect results
- Performance optimization (skip detection overhead)

**Note**: Manual encoding override persists across multiple `Run()` calls until reset with `SetEncoding("")`.

## Implementation Notes

### Platform-Specific Behaviors

#### Windows

- **Executable resolution**: Tries `.exe`, `.cmd`, `.bat` extensions
- **Batch files**: Executed via `cmd.exe /c` wrapper
- **Argument quoting**: Uses `CommandLineToArgvW` rules (double quotes, backslash escaping)
- **UNC paths**: Supported for working directory
- **Code page detection**: `GetACP()` for console, `GetConsoleOutputCP()` for output

#### Unix

- **Executable resolution**: PATH search only, no extension added
- **Permissions**: Must have execute bit set (`chmod +x`)
- **Argument quoting**: Shell-style escaping (backslash, single quotes)
- **Locale detection**: Parses `LC_ALL` → `LC_CTYPE` → `LANG` in priority order
- **Symlinks**: Followed transparently (both for executable and working directory)

### Performance

- **Process spawning**: <10ms overhead (filesystem-bound for PATH lookup)
- **Encoding detection**: <5ms (cached after first query)
- **Output capture**: Streaming (no buffering limits, handles large output)
- **Memory**: Output stored in strings (UTF-8 encoded, ~1.5x process output size)

### Error Handling

**Process execution errors** (returned as `error`, not in `ProcessResult`):
- Executable not found
- Permission denied
- Working directory missing
- Invalid environment variable keys

**Process failure** (returned in `ProcessResult.ExitCode`, no error):
- Command executes but fails (e.g., `dotnet restore` fails)
- Check `ExitCode != 0` and inspect `Stderr` for details

**Encoding errors**:
- UTF-8 validation fails → fallback to system encoding
- System encoding detection fails → replace invalid bytes with `�`
- No error returned (best-effort decoding)

### Testing Strategy

**Unit tests**:
- PATH resolution (with/without extensions)
- Argument quoting (paths with spaces, special characters)
- Environment variable merging
- Encoding detection (UTF-8, system encoding, fallback)

**Integration tests**:
- Spawn `dotnet --version` on all platforms
- Capture output with various encodings (UTF-8, Latin-1, Shift-JIS)
- Verify exit codes (success, failure, not found)
- Test working directory handling (absolute, relative, UNC)

## Example: Running dotnet restore

```go
import "github.com/yourusername/lazynuget/internal/platform"

func restoreProject(projectPath string) error {
    spawner := platform.NewProcessSpawner()

    result, err := spawner.Run(
        "dotnet",
        []string{
            "restore",
            projectPath,
            "--verbosity", "minimal",
        },
        filepath.Dir(projectPath), // Working directory = project directory
        map[string]string{
            "DOTNET_CLI_TELEMETRY_OPTOUT": "1", // Disable telemetry
        },
    )

    if err != nil {
        return fmt.Errorf("failed to spawn dotnet: %w", err)
    }

    if result.ExitCode != 0 {
        return fmt.Errorf("dotnet restore failed (exit %d):\n%s",
            result.ExitCode, result.Stderr)
    }

    // Success - log output
    log.Info("dotnet restore succeeded", "output", result.Stdout)
    return nil
}
```

## Example: Custom Encoding

```go
func runLegacyTool() (string, error) {
    spawner := platform.NewProcessSpawner()

    // Legacy tool outputs Shift-JIS on Japanese Windows
    spawner.SetEncoding("shift-jis")
    defer spawner.SetEncoding("") // Reset to auto-detect

    result, err := spawner.Run("legacy-tool.exe", nil, "", nil)
    if err != nil {
        return "", err
    }

    // Output is now decoded to UTF-8
    return result.Stdout, nil
}
```
