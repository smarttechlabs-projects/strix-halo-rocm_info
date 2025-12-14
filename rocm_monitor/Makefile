# Makefile for ROCm Monitor

# Variables
BINARY_NAME=rocm-monitor
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags "-s -w"

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_NAME) .

# Run the application
.PHONY: run
run: build
	@echo "Starting $(BINARY_NAME)..."
	./$(BINARY_NAME)

# Run with custom flags
.PHONY: run-dev
run-dev: build
	@echo "Starting $(BINARY_NAME) in development mode..."
	./$(BINARY_NAME) -interval 1s -metrics

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@$(GO) clean

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GO) test -v ./...

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Run linters
.PHONY: lint
lint:
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	$(GO) mod download
	$(GO) mod tidy

# Build for different platforms
.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 .

# Docker build (optional)
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t rocm-monitor:latest .

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make build      - Build the binary"
	@echo "  make run        - Build and run the application"
	@echo "  make run-dev    - Run in development mode with metrics"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make test       - Run tests"
	@echo "  make fmt        - Format code"
	@echo "  make lint       - Run linters"
	@echo "  make deps       - Install dependencies"
	@echo "  make build-linux - Build for Linux AMD64"
	@echo "  make help       - Show this help message"