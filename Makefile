.PHONY: build run test clean install iso lint fmt help deps

# Binary names
BINARY_NAME=ctrlsrvd
BINARY_PATH=bin/$(BINARY_NAME)

# Build variables
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags "-s -w"

# Directories
BUILD_DIR=bin
ISO_DIR=debian-iso
INSTALL_PREFIX=/usr/local

# Default target
all: build

## help: Show this help message
help:
	@echo 'Usage:'
	@echo '  make <target>'
	@echo ''
	@echo 'Targets:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## deps: Download Go dependencies
deps:
	$(GO) mod download
	$(GO) mod tidy
	$(GO) mod verify

## build: Build the binary
build: deps
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_PATH) ./cmd/ctrlsrvd

## run: Run the application locally
run: build
	@echo "Running $(BINARY_NAME)..."
	@if [ -f config.yaml ]; then \
		CONFIG_PATH=config.yaml ./$(BINARY_PATH); \
	else \
		CONFIG_PATH=config.example.yaml ./$(BINARY_PATH); \
	fi

## test: Run tests
test:
	$(GO) test -v -race -coverprofile=coverage.out ./...

## coverage: Show test coverage
coverage: test
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## lint: Run linters
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin"; \
	fi
	@if command -v shellcheck >/dev/null 2>&1; then \
		shellcheck scripts/*.sh; \
	else \
		echo "shellcheck not installed. Install: apt install shellcheck"; \
	fi

## fmt: Format code
fmt:
	@echo "Formatting Go code..."
	$(GO) fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi

## clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	rm -f *.iso
	rm -rf $(ISO_DIR)/binary $(ISO_DIR)/chroot $(ISO_DIR)/.build

## install: Install binary to system (requires sudo)
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PREFIX)/bin/..."
	sudo install -m 755 $(BINARY_PATH) $(INSTALL_PREFIX)/bin/
	@echo "Installing systemd service..."
	sudo install -m 644 config/systemd/ctrlsrvd.service /etc/systemd/system/
	sudo systemctl daemon-reload
	@echo "Installation complete. Enable with: sudo systemctl enable ctrlsrvd"

## uninstall: Remove installed files (requires sudo)
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	sudo systemctl stop ctrlsrvd 2>/dev/null || true
	sudo systemctl disable ctrlsrvd 2>/dev/null || true
	sudo rm -f $(INSTALL_PREFIX)/bin/$(BINARY_NAME)
	sudo rm -f /etc/systemd/system/ctrlsrvd.service
	sudo systemctl daemon-reload

## iso: Build custom Debian ISO
iso:
	@echo "Building custom Debian ISO..."
	cd $(ISO_DIR) && sudo ./build-iso.sh

## dev-setup: Setup development environment
dev-setup:
	@echo "Setting up development environment..."
	$(GO) install golang.org/x/tools/cmd/goimports@latest
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Installing system dependencies..."
	@echo "Run: sudo apt install shellcheck"

## docker-build: Build in Docker (for testing)
docker-build:
	docker run --rm -v "$(PWD)":/workspace -w /workspace golang:1.24 make build

## release: Build release binaries for multiple platforms
release: clean
	@echo "Building release binaries..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/ctrlsrvd
	GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/ctrlsrvd
	GOOS=linux GOARCH=arm GOARM=7 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-armv7 ./cmd/ctrlsrvd
	@echo "Release binaries built in $(BUILD_DIR)/"

## verify: Run all checks before commit
verify: fmt lint test
	@echo "All checks passed!"

# Development helpers
.PHONY: watch
## watch: Auto-rebuild on file changes (requires entr)
watch:
	@if command -v entr >/dev/null 2>&1; then \
		find . -name '*.go' | entr -r make run; \
	else \
		echo "entr not installed. Install: apt install entr"; \
	fi