# Glide CLI Makefile

.PHONY: test coverage coverage-html coverage-func clean

# Run all tests
test:
	go test ./... -v

# Run tests with coverage summary
coverage:
	go test ./... -cover

# Generate detailed coverage report
coverage-func:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out
	@echo ""
	@echo "Total coverage:"
	@go tool cover -func=coverage.out | tail -1

# Generate HTML coverage report and open it
coverage-html:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Opening coverage report in browser..."
	@open coverage.html 2>/dev/null || xdg-open coverage.html 2>/dev/null || echo "Please open coverage.html manually"

# Run tests for specific package with coverage
test-pkg:
	@read -p "Enter package path (e.g., ./internal/config): " pkg; \
	go test $$pkg -v -cover

# Clean coverage files
clean:
	rm -f coverage.out coverage.html

# Quick coverage check (one-liner)
quick-coverage:
	@go test ./... -cover 2>&1 | grep -E "ok|FAIL" | column -t