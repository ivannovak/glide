# Glide Plugin Makefile
# Customize PLUGIN_NAME to match your plugin

# Configuration
PLUGIN_NAME := myplugin
BINARY_NAME := glide-plugin-$(PLUGIN_NAME)
INSTALL_PATH := $(HOME)/.glide/plugins
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Go build flags
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"
BUILD_FLAGS := -trimpath

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
NC := \033[0m # No Color

.PHONY: all build install clean test dev release help

## help: Show this help message
help:
	@echo "$(GREEN)Glide Plugin Makefile$(NC)"
	@echo ""
	@echo "$(YELLOW)Usage:$(NC)"
	@echo "  make [target]"
	@echo ""
	@echo "$(YELLOW)Targets:$(NC)"
	@grep -E '^## ' Makefile | sed 's/## /  /'

## build: Build the plugin binary
build:
	@echo "$(GREEN)Building plugin...$(NC)"
	go build $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "$(GREEN)✓ Built $(BINARY_NAME)$(NC)"

## install: Install plugin to ~/.glide/plugins/
install: build
	@echo "$(GREEN)Installing plugin...$(NC)"
	@mkdir -p $(INSTALL_PATH)
	@cp $(BINARY_NAME) $(INSTALL_PATH)/
	@chmod +x $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "$(GREEN)✓ Installed to $(INSTALL_PATH)/$(BINARY_NAME)$(NC)"

## clean: Remove built binary
clean:
	@echo "$(YELLOW)Cleaning...$(NC)"
	@rm -f $(BINARY_NAME)
	@rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "$(GREEN)✓ Cleaned$(NC)"

## test: Test the plugin (build and run basic tests)
test: build
	@echo "$(GREEN)Testing plugin...$(NC)"
	@# Test that binary was created
	@if [ ! -f $(BINARY_NAME) ]; then \
		echo "$(RED)✗ Binary not found$(NC)"; \
		exit 1; \
	fi
	@# Test that it's executable
	@if [ ! -x $(BINARY_NAME) ]; then \
		echo "$(RED)✗ Binary not executable$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)✓ Basic tests passed$(NC)"
	@# Run Go tests if any exist
	@if [ -f "*_test.go" ]; then \
		go test -v ./...; \
	fi

## dev: Build and install for development (quick iteration)
dev:
	@echo "$(GREEN)Development build...$(NC)"
	@go build -o $(BINARY_NAME) .
	@mkdir -p $(INSTALL_PATH)
	@cp $(BINARY_NAME) $(INSTALL_PATH)/
	@chmod +x $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "$(GREEN)✓ Ready for testing$(NC)"
	@echo "$(YELLOW)Test with: glide $(PLUGIN_NAME) hello$(NC)"

## release: Build optimized release binary
release:
	@echo "$(GREEN)Building release binary...$(NC)"
	CGO_ENABLED=0 go build -a $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "$(GREEN)✓ Release binary built$(NC)"
	@echo "  Size: $$(du -h $(BINARY_NAME) | cut -f1)"
	@echo "  Version: $(VERSION)"

## deps: Download and tidy dependencies
deps:
	@echo "$(GREEN)Managing dependencies...$(NC)"
	go mod download
	go mod tidy
	@echo "$(GREEN)✓ Dependencies updated$(NC)"

## verify: Verify plugin loads correctly
verify: install
	@echo "$(GREEN)Verifying plugin...$(NC)"
	@if glide plugins list | grep -q "$(PLUGIN_NAME)"; then \
		echo "$(GREEN)✓ Plugin loads correctly$(NC)"; \
		glide plugins info $(PLUGIN_NAME); \
	else \
		echo "$(RED)✗ Plugin not found$(NC)"; \
		echo "$(YELLOW)Check that:$(NC)"; \
		echo "  1. Plugin is installed to $(INSTALL_PATH)/$(BINARY_NAME)"; \
		echo "  2. Plugin has execute permissions"; \
		echo "  3. Glide can access the plugins directory"; \
		exit 1; \
	fi

## run: Run a plugin command (use with CMD=commandname)
run: install
	@if [ -z "$(CMD)" ]; then \
		echo "$(YELLOW)Usage: make run CMD=hello$(NC)"; \
		echo "$(YELLOW)Available commands:$(NC)"; \
		glide $(PLUGIN_NAME) --help 2>/dev/null || true; \
	else \
		echo "$(GREEN)Running: glide $(PLUGIN_NAME) $(CMD)$(NC)"; \
		glide $(PLUGIN_NAME) $(CMD); \
	fi

## cross-compile: Build for multiple platforms
cross-compile:
	@echo "$(GREEN)Cross-compiling...$(NC)"
	@mkdir -p dist
	
	@echo "Building for Darwin AMD64..."
	@GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) \
		-o dist/$(BINARY_NAME)-darwin-amd64 .
	
	@echo "Building for Darwin ARM64..."
	@GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS) $(LDFLAGS) \
		-o dist/$(BINARY_NAME)-darwin-arm64 .
	
	@echo "Building for Linux AMD64..."
	@GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) \
		-o dist/$(BINARY_NAME)-linux-amd64 .
	
	@echo "Building for Linux ARM64..."
	@GOOS=linux GOARCH=arm64 go build $(BUILD_FLAGS) $(LDFLAGS) \
		-o dist/$(BINARY_NAME)-linux-arm64 .
	
	@echo "$(GREEN)✓ Cross-compilation complete$(NC)"
	@echo "Binaries in: ./dist/"
	@ls -lh dist/

## uninstall: Remove plugin from ~/.glide/plugins/
uninstall:
	@echo "$(YELLOW)Uninstalling plugin...$(NC)"
	@rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "$(GREEN)✓ Plugin uninstalled$(NC)"

## fmt: Format Go code
fmt:
	@echo "$(GREEN)Formatting code...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)✓ Code formatted$(NC)"

## lint: Run linter (requires golangci-lint)
lint:
	@echo "$(GREEN)Running linter...$(NC)"
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)golangci-lint not installed, skipping...$(NC)"; \
		echo "Install with: brew install golangci-lint"; \
	fi

# Default target
all: build