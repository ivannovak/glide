# Makefile for Glide CLI
# Ensures local development matches CI pipeline exactly

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet

# Binary name
BINARY_NAME=glide
BINARY_PATH=./cmd/glide

# Versions (match CI)
GOLANGCI_VERSION=v1.61.0
GO_VERSION=1.24

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
NC=\033[0m # No Color

.PHONY: all build test clean lint lint-fix install help

## help: Display this help message
help:
	@echo "Glide CLI Development Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^##' Makefile | sed 's/## /  /'

## all: Run all checks and build
all: lint test build

## build: Build the binary
build:
	@echo "$(GREEN)Building $(BINARY_NAME)...$(NC)"
	$(GOBUILD) -o $(BINARY_NAME) $(BINARY_PATH)

## install: Install the binary to $GOPATH/bin
install: build
	@echo "$(GREEN)Installing $(BINARY_NAME) to $(GOPATH)/bin...$(NC)"
	@cp $(BINARY_NAME) $(GOPATH)/bin/

## test: Run all tests
test:
	@echo "$(GREEN)Running tests...$(NC)"
	$(GOTEST) -race -coverprofile=coverage.out -covermode=atomic ./...
	@echo "$(GREEN)Test coverage:$(NC)"
	@$(GOCMD) tool cover -func=coverage.out | grep total: || true

## test-unit: Run unit tests only
test-unit:
	@echo "$(GREEN)Running unit tests...$(NC)"
	$(GOTEST) -race -short ./...

## test-integration: Run integration tests
test-integration:
	@echo "$(GREEN)Running integration tests...$(NC)"
	$(GOTEST) -tags=integration -timeout=60s ./tests/integration/...
	$(GOTEST) -v -timeout=10m ./tests/e2e/...

## lint: Run all linters (matches CI exactly)
lint: lint-fmt lint-vet
	@echo "$(GREEN)✓ All lint checks passed!$(NC)"

## lint-fmt: Check code formatting (matches CI)
lint-fmt:
	@echo "$(YELLOW)Checking code formatting...$(NC)"
	@if [ "$$($(GOFMT) -s -l . | wc -l)" -gt 0 ]; then \
		echo "$(RED)The following files need formatting:$(NC)"; \
		$(GOFMT) -s -l .; \
		echo ""; \
		echo "Run 'make lint-fix' to automatically fix formatting"; \
		exit 1; \
	else \
		echo "$(GREEN)✓ Code formatting OK$(NC)"; \
	fi

## lint-vet: Run go vet
lint-vet:
	@echo "$(YELLOW)Running go vet...$(NC)"
	@$(GOVET) ./... || (echo "$(RED)✗ go vet failed$(NC)" && exit 1)
	@echo "$(GREEN)✓ go vet OK$(NC)"

## lint-fix: Automatically fix formatting issues
lint-fix:
	@echo "$(GREEN)Fixing code formatting...$(NC)"
	@$(GOFMT) -s -w .
	@echo "$(GREEN)✓ Formatting fixed$(NC)"

## clean: Remove build artifacts
clean:
	@echo "$(GREEN)Cleaning...$(NC)"
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out
	@rm -f results.sarif

## mod-tidy: Tidy and verify go modules
mod-tidy:
	@echo "$(GREEN)Tidying modules...$(NC)"
	@$(GOMOD) tidy
	@$(GOMOD) verify

## pre-commit: Run all checks before committing
pre-commit: lint-fix lint test
	@echo "$(GREEN)✓ Ready to commit!$(NC)"

## ci: Simulate CI checks locally (what actually runs in CI)
ci: lint-fmt test
	@echo "$(GREEN)✓ CI checks passed locally!$(NC)"

# Default target
.DEFAULT_GOAL := help
