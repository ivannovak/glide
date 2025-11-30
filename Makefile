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

.PHONY: all build test test-unit test-integration test-coverage test-coverage-html test-coverage-package test-coverage-diff clean lint lint-fix install help

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

## test-coverage: Run tests with coverage gates (enforces 80% minimum)
test-coverage:
	@echo "$(GREEN)Running tests with coverage gates...$(NC)"
	@$(GOTEST) -race -coverprofile=coverage.out -covermode=atomic ./...
	@$(GOCMD) tool cover -func=coverage.out > coverage.txt
	@echo ""
	@echo "$(YELLOW)===== Coverage Report =====$(NC)"
	@cat coverage.txt | grep -E "^github.com/ivannovak/glide" | awk '{printf "%-60s %s\n", $$1, $$3}'
	@echo ""
	@TOTAL_COVERAGE=$$(cat coverage.txt | grep total: | awk '{print $$3}' | sed 's/%//'); \
	echo "$(YELLOW)Total Coverage: $$TOTAL_COVERAGE%$(NC)"; \
	echo ""; \
	if [ "$$(echo "$$TOTAL_COVERAGE < 80" | bc)" -eq 1 ]; then \
		echo "$(RED)✗ Coverage $$TOTAL_COVERAGE% is below target 80%$(NC)"; \
		echo "$(YELLOW)  Continue improving test coverage to meet gold standard.$(NC)"; \
		if [ "$$(echo "$$TOTAL_COVERAGE < 25" | bc)" -eq 1 ]; then \
			echo "$(RED)✗ Coverage $$TOTAL_COVERAGE% is below minimum 25%$(NC)"; \
			exit 1; \
		fi; \
	else \
		echo "$(GREEN)✓ Coverage meets gold standard target!$(NC)"; \
	fi

## test-coverage-html: Generate HTML coverage report
test-coverage-html:
	@echo "$(GREEN)Generating HTML coverage report...$(NC)"
	@$(GOTEST) -coverprofile=coverage.out -covermode=atomic ./...
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report generated: coverage.html$(NC)"
	@echo "$(YELLOW)  Open with: open coverage.html$(NC)"

## test-coverage-package: Show per-package coverage
test-coverage-package:
	@echo "$(GREEN)Analyzing per-package coverage...$(NC)"
	@$(GOTEST) -coverprofile=coverage.out -covermode=atomic ./...
	@echo ""
	@echo "$(YELLOW)===== Per-Package Coverage =====$(NC)"
	@$(GOCMD) tool cover -func=coverage.out | \
		grep -E "^github.com/ivannovak/glide" | \
		awk -F'[:\t]' '{pkg=$$1; sub(/\/[^\/]+\.go$$/, "", pkg); cov[pkg]+=$$3; count[pkg]++} \
		END {for (p in cov) printf "%-60s %5.1f%%\n", p, cov[p]/count[p]}' | \
		sort -k2 -n
	@echo ""
	@echo "$(YELLOW)===== Critical Packages (<80% coverage) =====$(NC)"
	@$(GOCMD) tool cover -func=coverage.out | \
		grep -E "pkg/plugin/sdk|internal/cli|pkg/prompt|internal/config|pkg/errors|pkg/output" | \
		awk '{print $$1, $$3}' | \
		awk '{if ($$2 < "80.0%") printf "%-60s %s $(RED)(BELOW TARGET)$(NC)\n", $$1, $$2; else printf "%-60s %s $(GREEN)(OK)$(NC)\n", $$1, $$2}'

## test-coverage-diff: Show coverage diff vs main branch
test-coverage-diff:
	@echo "$(GREEN)Calculating coverage diff vs main branch...$(NC)"
	@$(GOTEST) -coverprofile=coverage.out -covermode=atomic ./... 2>&1 | tail -1
	@CURRENT_COVERAGE=$$($(GOCMD) tool cover -func=coverage.out | grep total: | awk '{print $$3}' | sed 's/%//'); \
	git stash -q; \
	git checkout main -q 2>/dev/null || git checkout master -q 2>/dev/null; \
	$(GOTEST) -coverprofile=coverage-main.out -covermode=atomic ./... 2>&1 | tail -1; \
	MAIN_COVERAGE=$$($(GOCMD) tool cover -func=coverage-main.out | grep total: | awk '{print $$3}' | sed 's/%//'); \
	git checkout - -q; \
	git stash pop -q 2>/dev/null || true; \
	DIFF=$$(echo "$$CURRENT_COVERAGE - $$MAIN_COVERAGE" | bc); \
	echo ""; \
	echo "$(YELLOW)Coverage Comparison:$(NC)"; \
	echo "  Main branch:    $$MAIN_COVERAGE%"; \
	echo "  Current branch: $$CURRENT_COVERAGE%"; \
	if [ "$$(echo "$$DIFF > 0" | bc)" -eq 1 ]; then \
		echo "  Difference:     $(GREEN)+$$DIFF%$(NC) ✓ Improved"; \
	elif [ "$$(echo "$$DIFF < 0" | bc)" -eq 1 ]; then \
		echo "  Difference:     $(RED)$$DIFF%$(NC) ✗ Regression"; \
		exit 1; \
	else \
		echo "  Difference:     $$DIFF% (no change)"; \
	fi; \
	rm -f coverage-main.out

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
	@rm -f coverage.out coverage.txt coverage.html coverage-main.out
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
