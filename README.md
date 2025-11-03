# LazyNuGet

A modern terminal user interface (TUI) for NuGet package management, inspired by lazygit and lazydocker.

## Features

- Cross-platform application bootstrap (Windows, macOS, Linux)
- Graceful shutdown with SIGINT/SIGTERM handling
- Configuration system with CLI > Env > File > Default precedence
- Hot-reload configuration changes without restart
- AES-256-GCM encryption for sensitive values
- YAML and TOML configuration file support
- Platform-specific configuration and keychain integration
- Non-interactive mode for CI/testing environments
- 5-layer panic recovery for stability
- TTY detection and automatic mode switching

## Requirements

- **Go**: 1.24 or higher
- **.NET SDK**: 6.0 or higher (for NuGet operations)
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

## Configuration

LazyNuGet supports YAML and TOML configuration files in platform-specific locations:

- **macOS**: `~/Library/Application Support/lazynuget/config.yml` (or `.toml`)
- **Linux**: `~/.config/lazynuget/config.yml` (or `.toml`)
- **Windows**: `%APPDATA%\lazynuget\config.yml` (or `.toml`)

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
