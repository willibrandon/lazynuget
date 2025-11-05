# LazyNuGet

A modern terminal user interface (TUI) for NuGet package management, inspired by lazygit and lazydocker.

## Features

### Application Infrastructure
- Cross-platform application bootstrap (Windows, macOS, Linux)
- Graceful shutdown with SIGINT/SIGTERM handling
- 5-layer panic recovery for stability
- Non-interactive mode for CI/testing environments

### Configuration Management
- Configuration system with CLI > Env > File > Default precedence
- Hot-reload configuration changes without restart
- AES-256-GCM encryption for sensitive values
- YAML and TOML configuration file support
- Platform-specific configuration and keychain integration

### Platform Abstraction
- **OS/Architecture Detection**: Automatic Windows, macOS, Linux detection
- **Path Resolution**: Platform-appropriate config/cache directories with XDG/APPDATA support
- **Terminal Capabilities**: Color depth detection (16/256/TrueColor), Unicode support, resize events
- **Process Spawning**: Multi-platform text encoding (UTF-8, Windows-1252, Shift-JIS, etc.)
- **TTY Detection**: Automatic interactive/non-interactive mode switching
- **Performance**: <1ms path operations, <10ms terminal detection

## Requirements

- **Go**: 1.24 or higher
- **.NET SDK**: 9.0 or higher (for NuGet operations)
- **Platform**: Windows, macOS, or Linux

## Building

```bash
# Quick development build
make build-dev

# Production build with optimizations
make build

# Install to GOPATH/bin
make install
```

## Usage

```bash
# Start LazyNuGet (interactive mode)
./lazynuget

# Show version
./lazynuget --version

# Show help
./lazynuget --help

# Use custom config
./lazynuget --config /path/to/config.yml

# Force non-interactive mode
./lazynuget --non-interactive

# Set log level
./lazynuget --log-level debug

# Encrypt sensitive values
./lazynuget encrypt "my-secret-value"
```

## Platform Support

LazyNuGet provides native support for Windows, macOS, and Linux with platform-specific optimizations:

### Directory Locations

**Configuration Directory:**
- **macOS**: `~/Library/Application Support/lazynuget/`
- **Linux**: `~/.config/lazynuget/` (respects `XDG_CONFIG_HOME`)
- **Windows**: `%APPDATA%\lazynuget\`

**Cache Directory:**
- **macOS**: `~/Library/Caches/lazynuget/`
- **Linux**: `~/.cache/lazynuget/` (respects `XDG_CACHE_HOME`)
- **Windows**: `%LOCALAPPDATA%\lazynuget\`

### Terminal Support

LazyNuGet automatically detects terminal capabilities:
- **Color Depth**: Supports 16-color, 256-color, and TrueColor (24-bit) terminals
- **Unicode Support**: Automatically falls back to ASCII if Unicode is not supported
- **Resize Handling**: Responds to terminal resize events in real-time

### Text Encoding

Process output is correctly decoded on all platforms:
- **Unix/macOS**: UTF-8 with locale detection (LC_ALL, LC_CTYPE, LANG)
- **Windows**: Automatic detection of Windows-1252, UTF-8, and legacy code pages
- **Other**: Support for Shift-JIS, EUC-JP, ISO-8859-1, and other encodings

## Configuration

LazyNuGet supports YAML and TOML configuration files. The default config file is `config.yml` (or `config.toml`) in the platform-specific configuration directory listed above.

### Example Configuration (YAML)

```yaml
version: "1.0"
logLevel: info              # debug, info, warn, error
logDir: <platform-default>
logFormat: text             # text or json
theme: default
compactMode: false
showHints: true
hotReload: true             # Auto-reload config changes
startupTimeout: 5s
shutdownTimeout: 30s
maxConcurrentOps: 4

# Color scheme
colorScheme:
  border: "#333333"
  borderFocus: "#00FF00"
  text: "#FFFFFF"
  background: "#1E1E1E"

# Operation timeouts
timeouts:
  networkRequest: 30s
  dotnetCLI: 60s
  fileOperation: 5s

# Log rotation
logRotation:
  maxSize: 10        # MB
  maxAge: 30         # days
  maxBackups: 5
  compress: true
```

### Encrypting Sensitive Values

Use the `encrypt` command to protect sensitive configuration values:

```bash
# Encrypt a value
lazynuget encrypt "my-api-key"
# Output: !encrypted:base64-encoded-ciphertext

# Use in config file
nugetSources:
  - name: "Private Feed"
    url: "https://nuget.example.com/v3/index.json"
    apiKey: !encrypted:AES256-GCM:base64data...
```

Encrypted values are stored using AES-256-GCM. The encryption key is derived from your system keychain or the `LAZYNUGET_ENCRYPTION_KEY` environment variable.

### Configuration Precedence

Configuration values are applied in this order (highest to lowest priority):

1. Command-line flags (`--log-level debug`)
2. Environment variables (`LAZYNUGET_LOG_LEVEL=debug`)
3. Configuration file (`config.yml` or `config.toml`)
4. Built-in defaults

### Environment Variables

All configuration options can be set via environment variables with the `LAZYNUGET_` prefix:

```bash
# Simple values
export LAZYNUGET_LOG_LEVEL=debug
export LAZYNUGET_THEME=dark
export LAZYNUGET_MAX_CONCURRENT_OPS=8

# Nested values (use underscores)
export LAZYNUGET_COLOR_SCHEME_BORDER="#00FF00"
export LAZYNUGET_TIMEOUTS_NETWORK_REQUEST=60s
export LAZYNUGET_LOG_ROTATION_MAX_SIZE=20
```

## Development

### Running Tests

```bash
# Run all tests
make test-all

# Run unit tests only
make test

# Run integration tests only
make test-int

# Generate coverage report
make coverage
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Clean build artifacts
make clean
```

## Contributing

LazyNuGet is in early development. Contributions are welcome once the foundation is complete.

## License

[MIT](LICENSE)
