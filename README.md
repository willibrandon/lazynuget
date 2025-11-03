# LazyNuGet

A modern terminal user interface (TUI) for NuGet package management, inspired by lazygit and lazydocker.

## Features

- Cross-platform application bootstrap (Windows, macOS, Linux)
- Graceful shutdown with SIGINT/SIGTERM handling
- Configuration system with CLI > Env > File > Default precedence
- Non-interactive mode for CI/testing environments
- 5-layer panic recovery for stability
- Platform-specific configuration directories
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
```

## Configuration

LazyNuGet looks for configuration files in platform-specific locations:

- **macOS**: `~/Library/Application Support/lazynuget/config.yml`
- **Linux**: `~/.config/lazynuget/config.yml`
- **Windows**: `%APPDATA%\lazynuget\config.yml`

### Example Configuration

```yaml
logLevel: info              # debug, info, warn, error
logDir: <platform-default>
theme: default
compactMode: false
showHints: true
startupTimeout: 5s
shutdownTimeout: 30s
maxConcurrentOps: 4
```

### Configuration Precedence

Configuration values are applied in this order (highest to lowest priority):

1. Command-line flags (`--log-level debug`)
2. Environment variables (`LAZYNUGET_LOG_LEVEL=debug`)
3. Configuration file (`config.yml`)
4. Built-in defaults

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
