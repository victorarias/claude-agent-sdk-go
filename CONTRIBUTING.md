# Contributing to Claude Agent SDK Go

Thank you for your interest in contributing to the Claude Agent SDK Go! This document provides guidelines and instructions for contributing to the project.

## Getting Started

### Prerequisites

- **Go 1.25 or later** - Install from [go.dev/dl](https://go.dev/dl/)
- **Claude CLI** - Required for running tests and examples:
  ```bash
  npm install -g @anthropic-ai/claude-code
  ```
- **Git** - For version control

### Development Setup

1. **Fork the repository** on GitHub

2. **Clone your fork**:
   ```bash
   git clone https://github.com/YOUR_USERNAME/claude-agent-sdk-go.git
   cd claude-agent-sdk-go
   ```

3. **Install dependencies**:
   ```bash
   go mod download
   ```

4. **Verify your setup**:
   ```bash
   go test ./...
   ```

5. **Enable local git hooks** (runs lint + tests before each commit):
   ```bash
   git config core.hooksPath .githooks
   ```

## Development Workflow

### Making Changes

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following the code style guidelines below

3. **Run tests and linter** to ensure everything works:
   ```bash
   # Run all tests
   go test ./...

   # Run tests with race detection
   go test -race ./...

   # Run linter
   make lint

   # Run all checks (format, lint, test, build)
   make all
   ```

4. **Commit your changes** using [Conventional Commits](https://www.conventionalcommits.org/) format:
   ```bash
   # Format: <type>(<scope>): <description>

   # Features (bumps minor version)
   git commit -m "feat: add streaming support"
   git commit -m "feat(hooks): add pre-compact hook type"

   # Bug fixes (bumps patch version)
   git commit -m "fix: handle nil pointer in parser"

   # Other types: docs, test, refactor, perf, chore, ci, build
   git commit -m "docs: update README examples"
   git commit -m "test(sdk): add coverage for edge cases"

   # Breaking changes (bumps major version)
   git commit -m "feat!: change Client.Query signature"
   ```

   See [docs/RELEASING.md](docs/RELEASING.md) for full details on commit types and versioning.

**Important**: Always run the linter before committing to ensure your code passes CI checks.

## Code Style

### Go Conventions

This project follows standard Go conventions and best practices:

- **Formatting**: All code must be formatted with `gofmt` and `goimports`:
  ```bash
  make fmt
  # or
  go fmt ./...
  goimports -w .
  ```

- **Linting**: Code must pass `golangci-lint` checks:
  ```bash
  make lint
  # or
  golangci-lint run
  ```

  See [docs/LINTING.md](docs/LINTING.md) for detailed linting configuration and guidelines.

- **Go Vet**: Run `go vet` to catch common issues (included in golangci-lint):
  ```bash
  go vet ./...
  ```

### Style Guidelines

- Use clear, descriptive names for functions, variables, and types
- Write godoc-style comments for all exported functions, types, and packages
- Keep functions focused and reasonably sized
- Prefer composition over inheritance
- Handle errors explicitly - avoid ignoring errors
- Use context for cancellation and timeouts
- Follow the project's existing patterns and idioms

### Package Organization

- `sdk/` - High-level client API (public-facing)
- `types/` - Core types, messages, options (public)
- `internal/` - Internal implementation details (private)
- `examples/` - Example applications and use cases
- `docs/` - Documentation

## Testing Requirements

### Test Coverage

- All new features must include tests
- Bug fixes should include regression tests
- Aim for meaningful test coverage, not just high percentages
- Tests should be clear and maintainable

### Running Tests

```bash
# Run all tests
go test ./...

# Run with race detector (required before submitting)
go test -race ./...

# Run specific package tests
go test ./sdk
go test ./types

# Run with verbose output
go test -v ./...

# Run with coverage report
go test -cover ./...
```

### Test Organization

- Unit tests go in `*_test.go` files alongside the code
- Integration tests use the Claude CLI and should clean up resources
- Use table-driven tests for testing multiple scenarios
- Mock external dependencies when appropriate

## Pull Request Process

### Before Submitting

1. **Ensure all tests pass**:
   ```bash
   go test ./...
   go test -race ./...
   ```

2. **Format your code**:
   ```bash
   gofmt -w .
   ```

3. **Run linters**:
   ```bash
   go vet ./...
   golint ./...
   ```

4. **Update documentation** if you've changed APIs or added features

5. **Add examples** for new features when appropriate

### Submitting a PR

1. **Push your branch** to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Open a Pull Request** on GitHub with:
   - Clear title describing the change
   - Description explaining what and why
   - Reference to any related issues
   - Test results showing all tests pass

3. **Address review feedback** promptly and professionally

4. **Keep your PR focused** - one feature or fix per PR when possible

### PR Review Process

- Maintainers will review your PR and may request changes
- All tests must pass before merging
- At least one maintainer approval is required
- PRs are typically reviewed within a few days

## Reporting Issues

### Bug Reports

When reporting bugs, include:

- **Go version**: Output of `go version`
- **Claude CLI version**: Output of `claude --version`
- **Operating system** and version
- **Clear description** of the issue
- **Steps to reproduce** the problem
- **Expected behavior** vs actual behavior
- **Code sample** demonstrating the issue (if applicable)
- **Error messages** and stack traces

### Feature Requests

When requesting features, include:

- **Clear description** of the feature and its benefits
- **Use cases** explaining when and why it would be useful
- **Examples** of how the API might look
- **Alternatives** you've considered

### Security Issues

Please report security vulnerabilities privately to the maintainers rather than opening public issues.

## Internal Development Notes

### Task Tracking

External contributors should use GitHub Issues for feature requests and bug reports.

## Getting Help

- **Documentation**: Check the [README](README.md) and [examples](examples/)
- **Issues**: Search existing issues or create a new one

## License

By contributing to this project, you agree that your contributions will be licensed under the GNU General Public License v3.0.

---

Thank you for contributing to Claude Agent SDK Go!
