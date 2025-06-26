# Development Guide

This document provides comprehensive information for developers working on the XenForo to GitHub Discussions migration tool.

> [!NOTE]
> This project follows Go best practices with a clean architecture and comprehensive testing strategy.

## Quick Start with Makefile

> [!TIP]
> Use the Makefile for all development tasks - it provides consistent commands across environments:

```bash
# Get help with available commands
make help

# Set up development environment
make deps

# Build the project
make build

# Run tests
make test

# Run pre-commit checks
make check

# Format code
make fmt
```

## Project Structure

The project follows Go best practices with a clean architecture:

```
cmd/
└── xenforo-to-gh-discussions/  # Application entry point (30 lines, minimal)
    └── main.go
internal/
├── config/            # Configuration and interactive prompts
├── xenforo/           # XenForo API client and models
├── github/            # GitHub GraphQL client and operations
├── bbcode/            # BB-code to Markdown conversion
├── attachments/       # File download and processing
├── progress/          # Migration progress tracking
├── migration/         # Migration orchestration and interactive flow
└── testutil/          # Shared test utilities and mocks
test/
├── integration/       # Integration tests
└── testdata/          # Test data and fixtures
```

> [!IMPORTANT]
> **Unit tests** are located alongside their respective source code following Go conventions (e.g., `internal/config/config_test.go`).

## Running Tests

### All Test Types

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests only
make test-integration

# Run with coverage report
make test-coverage

# Run with race detector
make test-race

# Run benchmarks
make bench
```

### Test Organization

> [!TIP]
> **Unit Tests**: Located with source code (`internal/*/`*`_test.go`)
> - Direct access to package internals
> - Better IDE integration and discovery
> - Simpler import paths

> [!NOTE]
> **Integration Tests**: Centralized in `test/integration/`
> - End-to-end migration workflows
> - Mock-based full pipeline testing
> - Real API interaction patterns

## Build Commands

### Development Builds

```bash
# Development build (with race detector)
make dev

# Production build
make build

# Install to $GOPATH/bin
make install
```

### Release Builds

```bash
# Build for all platforms
make build-all

# Create release packages
make package

# Full release process (clean, lint, test, build-all)
make release
```

### Alternative Manual Build

```bash
# Manual build without Makefile
go build -o xenforo-to-gh-discussions ./cmd/xenforo-to-gh-discussions
```

## Code Quality

### Linting and Formatting

```bash
# Format code
make fmt

# Run linter checks
make lint

# Run golangci-lint (if installed)
make golangci-lint

# Run all pre-commit checks
make check
```

> [!IMPORTANT]
> Always run `make check` before committing - it ensures code quality and test coverage.

### Quality Standards

The codebase maintains high-quality standards:

- **Cyclomatic Complexity**: All functions kept below complexity 15
- **Package Organization**: Clear separation of concerns
- **Test Coverage**: Comprehensive unit and integration tests
- **Documentation**: Detailed README and architecture docs

## Development Workflow

### Daily Development

```bash
# Watch for changes and auto-rebuild
make watch

# Clean build artifacts
make clean

# Update dependencies
make deps-update

# Tidy dependencies
make tidy
```

### Version Information

```bash
# Show version and build information
make version
```

### Docker Development

```bash
# Build Docker image
make docker-build

# Build and run Docker container
make docker-run
```

## Testing Strategy

### Unit Testing
- **Location**: Alongside source code in each package
- **Scope**: Individual function behavior and edge cases
- **Dependencies**: Minimal external dependencies, use mocks

### Integration Testing
- **Location**: `test/integration/`
- **Scope**: Complete migration workflows
- **Dependencies**: Uses `internal/testutil/` mocks

### Performance Testing
- **Benchmarks**: Critical migration path performance
- **Memory profiling**: Large dataset handling
- **Rate limiting**: API compliance testing

## Architecture Guidelines

> [!TIP]
> Follow these principles when contributing:

### Package Responsibilities
- **config**: Configuration management with interactive prompts
- **xenforo**: XenForo API client with retry logic
- **github**: GitHub GraphQL operations
- **bbcode**: BB-code to Markdown conversion
- **attachments**: Secure file handling
- **progress**: Migration progress tracking
- **migration**: High-level orchestration and interactive workflow
- **testutil**: Shared test utilities and mocks

### Design Patterns Used
- **Strategy Pattern**: BB-code conversion, file handling
- **Repository Pattern**: Progress persistence, configuration
- **Adapter Pattern**: API clients, mock implementations
- **Command Pattern**: Migration operations with retry logic

## Contributing Guidelines

### Code Style
- Follow standard Go formatting (`gofmt`)
- Use meaningful variable and function names
- Keep functions focused and under 15 cyclomatic complexity
- Add comprehensive tests for new functionality

### Commit Process
1. Run `make check` to ensure quality
2. Write clear, descriptive commit messages
3. Include tests for new features
4. Update documentation as needed

### Pull Request Process
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Make your changes following the guidelines above
4. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
5. Push to the branch (`git push origin feature/AmazingFeature`)
6. Open a Pull Request

## Debugging and Troubleshooting

### Common Development Issues

> [!CAUTION]
> Watch out for these common issues:

1. **Import cycles**: Keep package dependencies clean
2. **Race conditions**: Use `make test-race` regularly
3. **Memory leaks**: Profile with large datasets
4. **API rate limits**: Test with realistic delays

### Debug Build

```bash
# Build with debug information
go build -gcflags="all=-N -l" -o debug-binary ./cmd/xenforo-to-gh-discussions
```

### Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=.

# Memory profiling  
go test -memprofile=mem.prof -bench=.
```

## IDE Integration

### VS Code
- Install Go extension
- Use workspace settings for consistent formatting
- Enable automatic test discovery

### GoLand
- Import project settings
- Configure Go modules properly
- Use built-in test runner

## Environment Setup

### Prerequisites
- Go 1.24 or higher
- Make (for using Makefile commands)
- Git (for version control)
- Optional: Docker (for containerized development)

### Getting Started
1. Clone the repository
2. Run `make deps` to install dependencies
3. Run `make test` to verify setup
4. Run `make build` to create the binary
5. Start developing!

---

> [!TIP]
> For additional information, check the [main README](README.md) or [architecture documentation](ARCHITECTURE.md) first.
> 
