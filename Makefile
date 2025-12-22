# Claude Agent SDK Go - Makefile
# Common development tasks

.PHONY: all build test test-race test-cover lint fmt clean help

# Default target
all: fmt lint test build

# Build all packages
build:
	go build ./...

# Run tests
test:
	go test ./...

# Run tests with race detector
test-race:
	go test -race ./...

# Run tests with coverage
test-cover:
	go test -cover ./...

# Run tests with detailed coverage report
test-cover-html:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Run integration tests (requires Claude CLI)
test-integration:
	CLAUDE_TEST_INTEGRATION=1 go test -tags=integration -v ./sdk/...

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...
	goimports -w .

# Tidy dependencies
tidy:
	go mod tidy

# Vet code
vet:
	go vet ./...

# Clean build artifacts
clean:
	rm -f coverage.out coverage.html
	rm -f agents agents-filesystem budget error-handling hooks interrupt
	rm -f mcp-server partial-streaming permissions prompt-variations
	rm -f session settings-sources simple streaming system-prompt
	rm -f tools-advanced tools-config

# Build examples
build-examples:
	@for dir in examples/*/; do \
		echo "Building $$dir..."; \
		go build -o $$(basename $$dir) ./$$dir; \
	done

# Run a specific example (usage: make run-example NAME=simple)
run-example:
	go run ./examples/$(NAME)/

# Show help
help:
	@echo "Claude Agent SDK Go - Development Tasks"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all            - Format, lint, test, and build (default)"
	@echo "  build          - Build all packages"
	@echo "  test           - Run tests"
	@echo "  test-race      - Run tests with race detector"
	@echo "  test-cover     - Run tests with coverage"
	@echo "  test-cover-html - Generate HTML coverage report"
	@echo "  test-integration - Run integration tests (requires Claude CLI)"
	@echo "  lint           - Run golangci-lint"
	@echo "  fmt            - Format code (go fmt + goimports)"
	@echo "  tidy           - Tidy go.mod dependencies"
	@echo "  vet            - Run go vet"
	@echo "  clean          - Remove build artifacts"
	@echo "  build-examples - Build all examples"
	@echo "  run-example    - Run example (NAME=simple)"
	@echo "  help           - Show this help"
