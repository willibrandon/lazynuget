.PHONY: build build-dev clean test test-int test-all coverage fmt vet lint lint-fix tidy install run help

# Variables
BINARY_NAME=lazynuget
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-w -s -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

## help: Display this help message
help:
	@echo "LazyNuGet - Makefile targets:"
	@echo ""
	@grep -E '^##' $(MAKEFILE_LIST) | sed 's/##//'

## build-dev: Quick development build without optimizations
build-dev:
	@echo "Building $(BINARY_NAME) (dev)..."
	go build -o $(BINARY_NAME) cmd/lazynuget/main.go

## build: Production build with optimizations and version injection
build:
	@echo "Building $(BINARY_NAME) v$(VERSION) ($(COMMIT))..."
	go build $(LDFLAGS) -trimpath -o $(BINARY_NAME) cmd/lazynuget/main.go

## test: Run unit tests with race detector
test:
	@echo "Running unit tests..."
	go test -v -race ./internal/...

## test-int: Run integration tests
test-int:
	@echo "Running integration tests..."
	go test -v -race ./tests/integration/...

## test-all: Run all tests (unit + integration)
test-all: test test-int

## coverage: Generate test coverage report
coverage:
	@echo "Generating coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## clean: Remove build artifacts and caches
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	go clean -cache -testcache

## tidy: Clean up go.mod and go.sum
tidy:
	@echo "Tidying go modules..."
	go mod tidy

## fmt: Format Go code with gofumpt (stricter than gofmt)
fmt:
	@echo "Formatting code with gofumpt..."
	@go run mvdan.cc/gofumpt@latest -l -w .

## vet: Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

## lint: Run all modern code quality checks (read-only)
lint:
	@echo "Running modern linters..."
	@echo "→ go vet"
	@go vet ./...
	@echo "→ gofumpt (check only)"
	@go run mvdan.cc/gofumpt@latest -l . | tee /dev/stderr | test -z "$$(cat)"
	@echo "→ goimports (check only)"
	@go run golang.org/x/tools/cmd/goimports@latest -l . | tee /dev/stderr | test -z "$$(cat)"
	@echo "→ staticcheck"
	@go run honnef.co/go/tools/cmd/staticcheck@latest ./...
	@echo "→ modernize (check only)"
	@go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest ./...
	@echo "→ golangci-lint"
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run
	@echo "✓ All lint checks passed"

## lint-fix: Run all automatic fixers
lint-fix:
	@echo "Running automatic fixers..."
	@echo "→ gofumpt (format)"
	@go run mvdan.cc/gofumpt@latest -l -w .
	@echo "→ goimports (fix imports)"
	@go run golang.org/x/tools/cmd/goimports@latest -w .
	@echo "→ modernize (apply fixes)"
	@go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest -fix ./...
	@echo "→ golangci-lint (auto-fix)"
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run --fix
	@echo "✓ All automatic fixes applied"

## install: Install binary to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) -trimpath cmd/lazynuget/main.go

## run: Build and run the application
run: build-dev
	./$(BINARY_NAME)

.DEFAULT_GOAL := help
