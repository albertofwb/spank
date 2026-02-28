# spank - Cross-platform slap detector
# Auto-detects OS and builds accordingly

# Detect OS
UNAME_S := $(shell uname -s)

# Version info
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

# Binary name
BINARY := spank

# Platform-specific settings
ifeq ($(UNAME_S),Darwin)
    TARGET_OS := darwin
    TARGET_ARCH := arm64
    BUILD_TAGS :=
    CGO_ENABLED := 0
    $(info Detected: macOS - Building with Apple Silicon accelerometer support)
else ifeq ($(UNAME_S),Linux)
    TARGET_OS := linux
    TARGET_ARCH := amd64
    BUILD_TAGS :=
    CGO_ENABLED := 1
    $(info Detected: Linux - Building with microphone support)
else
    $(error Unsupported OS: $(UNAME_S). Only macOS and Linux are supported.)
endif

# Build directory
BUILD_DIR := ./build

# Default target
.PHONY: all build clean install uninstall test check-deps help

all: check-deps build

# Check platform-specific dependencies
.PHONY: check-deps
check-deps:
ifeq ($(UNAME_S),Darwin)
	@echo "Checking macOS dependencies..."
	@which go > /dev/null || (echo "Error: Go not found. Install from https://golang.org/dl/" && exit 1)
	@echo "✓ macOS dependencies OK (requires root at runtime for IOKit HID)"
else ifeq ($(UNAME_S),Linux)
	@echo "Checking Linux dependencies..."
	@which go > /dev/null || (echo "Error: Go not found. Install from https://golang.org/dl/" && exit 1)
	@echo "Checking for PortAudio..."
	@pkg-config --exists portaudio-2.0 2>/dev/null || (echo "Warning: PortAudio not found. Install: sudo apt install libportaudio2-dev portaudio19-dev" && exit 1)
	@echo "✓ Linux dependencies OK (uses PortAudio for microphone)"
endif

# Build binary
build:
	@echo "Building $(BINARY) for $(TARGET_OS)/$(TARGET_ARCH)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) .
	@echo "✓ Build complete: $(BUILD_DIR)/$(BINARY)"

# Install to system
install: build
ifeq ($(UNAME_S),Darwin)
	@echo "Installing to /usr/local/bin/$(BINARY)..."
	@sudo cp $(BUILD_DIR)/$(BINARY) /usr/local/bin/$(BINARY)
	@sudo chmod +x /usr/local/bin/$(BINARY)
	@echo "✓ Installed. Run with: sudo $(BINARY)"
else ifeq ($(UNAME_S),Linux)
	@echo "Installing to ~/.local/bin/$(BINARY)..."
	@mkdir -p ~/.local/bin
	@cp $(BUILD_DIR)/$(BINARY) ~/.local/bin/$(BINARY)
	@chmod +x ~/.local/bin/$(BINARY)
	@echo "✓ Installed. Run with: $(BINARY) (add ~/.local/bin to PATH if needed)"
endif

# Uninstall
uninstall:
ifeq ($(UNAME_S),Darwin)
	@echo "Removing /usr/local/bin/$(BINARY)..."
	@sudo rm -f /usr/local/bin/$(BINARY)
else ifeq ($(UNAME_S),Linux)
	@echo "Removing ~/.local/bin/$(BINARY)..."
	@rm -f ~/.local/bin/$(BINARY)
endif
	@echo "✓ Uninstalled"

# Cross-compile for all supported platforms
# Note: Cross-compilation has limitations:
#   - Linux binaries: Can be built on Linux with CGO
#   - macOS binaries: Must be built on macOS (requires IOKit framework)
cross-compile:
	@echo "Cross-compiling for supported platforms..."
	@mkdir -p $(BUILD_DIR)
ifeq ($(UNAME_S),Darwin)
	@echo "Building on macOS..."
	# macOS ARM64 (native)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-darwin-arm64 .
	@echo "  ✓ macOS ARM64"
	@echo ""
	@echo "Note: To build Linux binaries from macOS, use Docker:"
	@echo "  docker run --rm -v \$$PWD:/src -w /src golang:1.26 make build"
else ifeq ($(UNAME_S),Linux)
	@echo "Building on Linux..."
	# Linux AMD64 (with CGO)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-linux-amd64 .
	@echo "  ✓ Linux AMD64"
	# Linux ARM64 (with CGO, may need cross-compiler)
	@CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC=aarch64-linux-gnu-gcc go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-linux-arm64 . 2>/dev/null && echo "  ✓ Linux ARM64" || echo "  ⚠️  Linux ARM64 skipped (install gcc-aarch64-linux-gnu)"
	@echo ""
	@echo "Note: macOS binaries must be built on macOS (IOKit framework required)"
endif
	@echo ""
	@echo "Build complete:"
	@ls -la $(BUILD_DIR)/$(BINARY)-* 2>/dev/null || true

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@go clean
	@echo "✓ Clean complete"

# Run the application
run: build
ifeq ($(UNAME_S),Darwin)
	@echo "Running with sudo (required for accelerometer access)..."
	@sudo $(BUILD_DIR)/$(BINARY)
else ifeq ($(UNAME_S),Linux)
	@echo "Running..."
	@$(BUILD_DIR)/$(BINARY) 2>/dev/null
endif

# Run in sexy mode
run-sexy: build
ifeq ($(UNAME_S),Darwin)
	@sudo $(BUILD_DIR)/$(BINARY) --sexy
else
	@$(BUILD_DIR)/$(BINARY) --sexy 2>/dev/null
endif

# Run in halo mode
run-halo: build
ifeq ($(UNAME_S),Darwin)
	@sudo $(BUILD_DIR)/$(BINARY) --halo
else
	@$(BUILD_DIR)/$(BINARY) --halo 2>/dev/null
endif

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Format code
fmt:
	@echo "Formatting code..."
	@gofmt -w *.go

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run 2>/dev/null || echo "golangci-lint not installed, skipping"

# Show help
help:
	@echo "spank - Cross-platform slap detector"
	@echo ""
	@echo "Detected OS: $(UNAME_S)"
	@echo ""
	@echo "Available targets:"
	@echo "  make build         - Build the binary for current platform"
	@echo "  make install       - Install binary to system path"
	@echo "  make uninstall     - Remove binary from system path"
	@echo "  make run           - Build and run (with sudo on macOS)"
	@echo "  make run-sexy      - Run in sexy mode"
	@echo "  make run-halo      - Run in halo mode"
	@echo "  make cross-compile - Build for all supported platforms"
	@echo "  make clean         - Remove build artifacts"
	@echo "  make test          - Run tests"
	@echo "  make fmt           - Format Go code"
	@echo "  make lint          - Run linter"
	@echo "  make check-deps    - Check platform dependencies"
	@echo "  make help          - Show this help"
	@echo ""
ifeq ($(UNAME_S),Darwin)
	@echo "macOS Notes:"
	@echo "  - Requires Apple Silicon (M2+)"
	@echo "  - Requires sudo for IOKit HID accelerometer access"
	@echo "  - Run: sudo spank"
else ifeq ($(UNAME_S),Linux)
	@echo "Linux Notes:"
	@echo "  - Requires PortAudio for microphone support"
	@echo "  - Install: sudo apt install libportaudio2-dev portaudio19-dev"
	@echo "  - Run: spank (no sudo needed)"
endif
