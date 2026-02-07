# Linting Guide

This document describes the linting setup for the Claude Agent SDK Go project.

## Overview

The project uses [golangci-lint](https://golangci-lint.run/), a fast and comprehensive Go linters aggregator. The configuration is defined in `.golangci.yml` at the project root.

## Installation

### macOS (Homebrew)
```bash
brew install golangci-lint
```

### Linux
```bash
# Binary will be installed to $(go env GOPATH)/bin
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

### Other Methods
See the [official installation guide](https://golangci-lint.run/usage/install/).

## Usage

### Basic Commands

```bash
# Verify configuration is valid
golangci-lint config verify

# Run all linters
golangci-lint run

# Run with auto-fix where possible
golangci-lint run --fix

# Run on specific directories
golangci-lint run ./sdk/...
golangci-lint run ./types/...

# Show all linter names
golangci-lint linters
```

### Using Make

The project includes convenient Make targets:

```bash
# Run linter (same as golangci-lint run)
make lint

# Full check: format, lint, test, and build
make all
```

## Enabled Linters

The project uses a carefully selected set of linters:

### Essential
- **govet** - Official Go static analyzer
- **errcheck** - Checks for unchecked errors
- **staticcheck** - Advanced static analysis (SA checks)
- **unused** - Finds unused code
- **gosimple** - Suggests code simplifications
- **ineffassign** - Detects ineffectual assignments
- **typecheck** - Type checking errors

### Formatting
- **gofmt** - Checks code formatting
- **goimports** - Checks import ordering

### Quality
- **misspell** - Finds spelling mistakes in strings and comments
- **unconvert** - Removes unnecessary type conversions
- **goconst** - Finds repeated strings that could be constants
- **gosec** - Security-focused checks
- **revive** - General linting rules

### Complexity
- **gocyclo** - Detects high cyclomatic complexity
- **cyclop** - Additional complexity checks
- **funlen** - Limits function length

### Style/Best Practices
- **gocritic** - Comprehensive style checks
- **exportloopref** - Loop variable capture issues
- **noctx** - HTTP requests without context
- **nilerr** - Return nil instead of error
- **predeclared** - Shadowing predeclared identifiers
- **thelper** - Test helper detection
- **tparallel** - Parallel test detection
- **whitespace** - Whitespace issues

## Configuration Highlights

### Excluded Directories
The following directories are excluded from linting:
- `.github/` - GitHub workflows
- `.claude/` - Claude configuration

### Test File Relaxations
Test files (`*_test.go`) have relaxed rules for:
- **funlen** - Tests can be longer
- **gocyclo** - Tests can be more complex
- **cyclop** - Tests can be more complex
- **goconst** - Repeated strings are okay in tests
- **lll** - Longer lines are acceptable in tests

### Security Settings
The `gosec` linter has specific exclusions needed for SDK functionality:
- **G304** - File paths from variables (needed for SDK)
- **G204** - Subprocess with variables (needed for CLI management)

### Complexity Thresholds
- **Cyclomatic complexity**: 15 (moderate)
- **Function length**: 100 lines / 50 statements
- **Line length**: 140 characters

## CI/CD Integration

The linter runs automatically on:
- Push to `main` branch
- Pull requests to `main` branch

See `.github/workflows/lint.yml` for CI configuration.

The CI uses the same `.golangci.yml` configuration, ensuring local and CI results match.

## Common Issues and Solutions

### Import Ordering
If you get `goimports` errors, run:
```bash
goimports -w .
# or
make fmt
```

### Unused Code
Remove unused variables, functions, or imports. If the code is intentionally unused (e.g., future use), consider:
1. Removing it and adding it back when needed
2. Adding a clear TODO comment explaining why it's there

### Security Warnings
Review all `gosec` warnings carefully. Only ignore warnings if:
1. You understand the security implication
2. The usage is intentional and safe in context
3. You've documented why it's safe

### Long Functions
If a function exceeds length limits:
1. Extract helper functions for logical sub-operations
2. Consider if the function has too many responsibilities
3. Test functions can be longer (automatically excluded)

## Best Practices

1. **Run linter before committing**
   ```bash
   make lint
   ```

2. **Fix issues incrementally**
   - Don't accumulate linter issues
   - Fix warnings as you write code

3. **Use auto-fix when appropriate**
   ```bash
   golangci-lint run --fix
   ```

4. **Understand warnings, don't just suppress them**
   - Avoid using `//nolint` directives
   - If you must suppress, add a clear comment explaining why

5. **Keep configuration up to date**
   - Review enabled linters periodically
   - Update thresholds as project matures

## Troubleshooting

### Linter is too slow
The timeout is set to 5 minutes. If it's still too slow:
```bash
# Run on specific packages
golangci-lint run ./sdk/...
```

### Version mismatch
Ensure you're using a recent version:
```bash
golangci-lint --version
# Should be v2.7.0 or newer
```

### Config not found
Ensure you're running from the project root:
```bash
cd /path/to/claude-agent-sdk-go
golangci-lint run
```

## Resources

- [golangci-lint Documentation](https://golangci-lint.run/)
- [Linters List](https://golangci-lint.run/usage/linters/)
- [Configuration Reference](https://golangci-lint.run/usage/configuration/)
- [False Positives](https://golangci-lint.run/usage/false-positives/)
