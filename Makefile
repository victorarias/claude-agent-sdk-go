.PHONY: all test lint fmt build clean install-tools help

# Tool versions
GOTESTSUM_VERSION := latest
GOLANGCI_LINT_VERSION := v2.7.2

# Default target
all: fmt lint test build

# Install development tools
install-tools:
	go install gotest.tools/gotestsum@$(GOTESTSUM_VERSION)
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

# Run tests with gotestsum for clear output
test:
	go run gotest.tools/gotestsum@$(GOTESTSUM_VERSION) --format testname -- ./... -race

# Run tests with short format (CI-friendly)
test-ci:
	go run gotest.tools/gotestsum@$(GOTESTSUM_VERSION) --format pkgname --junitfile test-results.xml -- ./... -race

# Run tests with coverage
test-coverage:
	go run gotest.tools/gotestsum@$(GOTESTSUM_VERSION) --format testname -- ./... -race -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run ./internal/... ./sdk/... ./types/...

# Format code
fmt:
	gofmt -w .
	goimports -w -local github.com/victorarias/claude-agent-sdk-go .

# Build all packages
build:
	go build ./...

# Build examples
build-examples:
	go build ./examples/...

# Clean build artifacts
clean:
	rm -f coverage.out coverage.html test-results.xml
	go clean ./...

# Show help
help:
	@echo "Available targets:"
	@echo "  all            - Format, lint, test, and build (default)"
	@echo "  install-tools  - Install gotestsum and golangci-lint"
	@echo "  test           - Run tests with gotestsum"
	@echo "  test-ci        - Run tests with JUnit output for CI"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  lint           - Run golangci-lint"
	@echo "  fmt            - Format code with gofmt and goimports"
	@echo "  build          - Build all packages"
	@echo "  build-examples - Build example applications"
	@echo "  clean          - Remove build artifacts"
