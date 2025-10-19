# BitTorrent Client Makefile

# Variables
BINARY_NAME=torrent-client
BUILD_DIR=build
CMD_DIR=./cmd
SRC_DIRS=./bencode ./peer ./torrent ./client ./cmd
GO_FILES=$(shell find . -name "*.go" -not -path "./vendor/*")

# Build info
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

# Default target
.PHONY: all
all: clean test build

# Build the binary
.PHONY: build
build: $(BUILD_DIR)/$(BINARY_NAME)

$(BUILD_DIR)/$(BINARY_NAME): $(GO_FILES)
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Install binary to $GOPATH/bin
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME) to $$GOPATH/bin..."
	go install $(LDFLAGS) $(CMD_DIR)
	@echo "✅ Installed successfully"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...
	@echo "✅ All tests passed"

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# Run benchmarks
.PHONY: bench
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "✅ Code formatted"

# Lint code
.PHONY: lint
lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, using go vet..."; \
		go vet ./...; \
	fi
	@echo "✅ Linting complete"

# Check for security issues
.PHONY: security
security:
	@echo "Checking for security issues..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "⚠️  gosec not installed. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Tidy dependencies
.PHONY: tidy
tidy:
	@echo "Tidying dependencies..."
	go mod tidy
	@echo "✅ Dependencies tidied"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	go clean
	@echo "✅ Clean complete"

# Build for multiple platforms
.PHONY: build-all
build-all: clean
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)

	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)

	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_DIR)

	# macOS AMD64
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)

	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)

	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)

	@echo "✅ Multi-platform build complete"
	@ls -la $(BUILD_DIR)/

# Create release packages
.PHONY: package
package: build-all
	@echo "Creating release packages..."
	@mkdir -p $(BUILD_DIR)/packages

	# Create tar.gz for Unix systems
	cd $(BUILD_DIR) && tar -czf packages/$(BINARY_NAME)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64 && echo "📦 Linux AMD64 package created"
	cd $(BUILD_DIR) && tar -czf packages/$(BINARY_NAME)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64 && echo "📦 Linux ARM64 package created"
	cd $(BUILD_DIR) && tar -czf packages/$(BINARY_NAME)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64 && echo "📦 macOS AMD64 package created"
	cd $(BUILD_DIR) && tar -czf packages/$(BINARY_NAME)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64 && echo "📦 macOS ARM64 package created"

	# Create zip for Windows
	cd $(BUILD_DIR) && zip packages/$(BINARY_NAME)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe && echo "📦 Windows AMD64 package created"

	@echo "✅ All packages created in $(BUILD_DIR)/packages/"
	@ls -la $(BUILD_DIR)/packages/

# Run the application (for development)
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME) (use 'make run ARGS=\"file.torrent\"' to pass arguments)..."
	$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

# Run comprehensive integration tests
.PHONY: test-integration
test-integration: build
	@echo "Running integration tests..."
	./test.sh

# Development setup
.PHONY: dev-setup
dev-setup:
	@echo "Setting up development environment..."
	@echo "Installing development tools..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci-lint/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "Installing gosec..."; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
	fi
	go mod download
	@echo "✅ Development environment ready"

# Check everything (comprehensive check)
.PHONY: check
check: fmt lint test security
	@echo "✅ All checks passed"

# Quick development cycle
.PHONY: dev
dev: fmt test build
	@echo "✅ Development cycle complete"

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Build targets:"
	@echo "  build      - Build the binary"
	@echo "  build-all  - Build for multiple platforms"
	@echo "  install    - Install binary to \$$GOPATH/bin"
	@echo "  package    - Create release packages"
	@echo ""
	@echo "Development targets:"
	@echo "  dev        - Quick development cycle (fmt + test + build)"
	@echo "  dev-setup  - Setup development environment"
	@echo "  run        - Build and run the application (use ARGS=\"...\" for arguments)"
	@echo ""
	@echo "Quality targets:"
	@echo "  test       - Run unit tests"
	@echo "  test-integration - Run comprehensive integration tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  bench      - Run benchmarks"
	@echo "  fmt        - Format code"
	@echo "  lint       - Lint code"
	@echo "  security   - Check for security issues"
	@echo "  check      - Run all quality checks"
	@echo ""
	@echo "Utility targets:"
	@echo "  clean      - Clean build artifacts"
	@echo "  tidy       - Tidy dependencies"
	@echo "  help       - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make dev                          # Quick development cycle"
	@echo "  make run ARGS=\"ubuntu.torrent\"    # Run with torrent file"
	@echo "  make test-coverage               # Generate coverage report"
	@echo "  make build-all                   # Build for all platforms"