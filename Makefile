# tempural Makefile

# Go parameters
BINARY_NAME=tempural
GO=go
MAIN_PACKAGE=./

# Build and package directories
BUILD_DIR=build
DIST_DIR=dist

# Version information
VERSION=0.1.0
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date +%FT%T%z)
LDFLAGS=-ldflags "-X github.com/weslien/tempural/pkg/utils.Version=$(VERSION) -X github.com/weslien/tempural/pkg/utils.GitCommit=$(GIT_COMMIT) -X github.com/weslien/tempural/pkg/utils.BuildDate=$(BUILD_DATE)"

# Default target
.PHONY: all
all: clean build test

# Build the application
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run the application
.PHONY: run
run: build
	@$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

# Install the application to $GOPATH/bin
.PHONY: install
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install $(LDFLAGS) $(MAIN_PACKAGE)
	@echo "Installation complete"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR)
	@echo "Clean complete"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GO) test -v ./...
	@echo "Tests complete"

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@mkdir -p $(BUILD_DIR)
	$(GO) test -coverprofile=$(BUILD_DIR)/coverage.out ./...
	$(GO) tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo "Coverage report generated at $(BUILD_DIR)/coverage.html"

# Run linters and static analysis
.PHONY: lint
lint:
	@echo "Running linters..."
	@which golint > /dev/null || go install golang.org/x/lint/golint@latest
	golint ./...
	$(GO) vet ./...
	@echo "Linting complete"

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@echo "Formatting complete"

# Create distribution packages
.PHONY: dist
dist: build
	@echo "Creating distribution packages..."
	@mkdir -p $(DIST_DIR)
	@tar -czvf $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)
	@echo "Distribution packages created in $(DIST_DIR)"

# Help command
.PHONY: help
help:
	@echo "tempural Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all              Build, run tests (default)"
	@echo "  build            Build the application"
	@echo "  run ARGS='args'  Run the application with arguments"
	@echo "  install          Install the application to GOPATH/bin"
	@echo "  clean            Remove build artifacts"
	@echo "  test             Run tests"
	@echo "  test-coverage    Run tests with coverage report"
	@echo "  lint             Run linters and static analysis"
	@echo "  fmt              Format code"
	@echo "  dist             Create distribution packages"
	@echo "  help             Show this help message"
