.PHONY: build build-dev clean test test-int help

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

## fmt: Format Go code
fmt:
	@echo "Formatting code..."
	go fmt ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

## lint: Run all code quality checks
lint: fmt vet
	@echo "Linting complete"

## install: Install binary to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) -trimpath cmd/lazynuget/main.go

## run: Build and run the application
run: build-dev
	./$(BINARY_NAME)

.DEFAULT_GOAL := help
