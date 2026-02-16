.PHONY: help test test-verbose test-coverage build run clean run-dev fmt install demo lint

# Variables
BINARY_NAME=clippy
BINARY_PATH=./$(BINARY_NAME)
CMD_PATH=./cmd/$(BINARY_NAME)
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_FLAGS=-ldflags "-X main.version=$(VERSION)"

# Default target
help:
	@echo "Clippy - Terminal Clipboard Manager"
	@echo ""
	@echo "Available targets:"
	@echo "  make test              Run all tests"
	@echo "  make test-verbose      Run tests with verbose output"
	@echo "  make test-coverage     Run tests with coverage report"
	@echo "  make build             Build the binary"
	@echo "  make run               Build and run the app"
	@echo "  make run-dev           Run the app without building (if already built)"
	@echo "  make install           Build and install to /usr/local/bin"
	@echo "  make clean             Remove build artifacts"
	@echo "  make fmt               Format code with gofmt"
	@echo "  make lint              Run golangci-lint"
	@echo "  make demo              Run the demo application"
	@echo "  make help              Show this help message"

# Test targets
test:
	@echo "Running tests..."
	@go test ./...

test-verbose:
	@echo "Running tests (verbose)..."
	@go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Build targets
build:
	@echo "Building $(BINARY_NAME)..."
	@go build $(BUILD_FLAGS) -o $(BINARY_PATH) $(CMD_PATH)
	@echo "Build complete: $(BINARY_PATH)"

run: build
	@echo "Running $(BINARY_NAME)..."
	@$(BINARY_PATH)

run-dev:
	@echo "Running $(BINARY_NAME)..."
	@$(BINARY_PATH)

# Installation
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo cp $(BINARY_PATH) /usr/local/bin/$(BINARY_NAME)
	@echo "Installation complete: /usr/local/bin/$(BINARY_NAME)"

# Code quality
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Format complete"

lint:
	@echo "Running linter..."
	@golangci-lint run ./... 2>/dev/null || echo "golangci-lint not installed. Install with: brew install golangci-lint"

# Cleaning
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_PATH)
	@rm -f coverage.out coverage.html
	@go clean
	@echo "Clean complete"

# Demo
demo:
	@echo "Running demo application..."
	@go run ./demo/main.go
