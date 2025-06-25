# XenForo to GitHub Discussions Migration Tool

[![Go Report Card](https://goreportcard.com/badge/github.com/exileum/xenforo-to-gh-discussions)](https://goreportcard.com/report/github.com/exileum/xenforo-to-gh-discussions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org)

A robust, well-architected Go CLI tool to migrate forum threads, posts, and attachments from XenForo 2 to GitHub Discussions using their respective APIs.

## Features

- **Clean Architecture**: Modular design with separate packages for different concerns
- **GraphQL-based GitHub Integration**: Uses GitHub's GraphQL API for creating discussions and comments
- **Comprehensive BB-Code Support**: Converts XenForo BB codes to Markdown, including:
    - Text formatting (bold, italic, underline, strikethrough)
    - Empty tag detection and removal (prevents empty markdown formatting)
    - URLs and images with markdown link preservation
    - Quotes (with attribution)
    - Code blocks
    - Spoilers (both block and inline)
    - Lists
    - Media embeds
- **Smart Attachment Handling**:
    - Secure filename sanitization to prevent path traversal attacks
    - Downloads attachments and organizes by file type
    - Embeds images directly in Markdown
    - Links to other file types
- **Robust Error Handling**:
    - Proper error handling for progress saving operations
    - Retry logic with exponential backoff for rate limits
    - Progress tracking for resumable migrations
    - Detailed error logging
    - Thread-level failure handling to prevent partial migrations
- **Migration Features**:
    - Dry-run mode for testing
    - Pre-flight checks to validate configuration
    - Progress tracking with JSON persistence
    - Rate limiting compliance
    - Environment variable support for configuration

## Architecture

📋 **[View detailed architecture and technical documentation →](ARCHITECTURE.md)**

The migration tool follows a clean architecture pattern with well-separated concerns:

```
main.go                # Application entry point (30 lines, minimal)
internal/
├── config/            # Configuration management with env var support
├── xenforo/           # XenForo API client and models
├── github/            # GitHub GraphQL client and operations
├── bbcode/            # BB-code to Markdown conversion
├── attachments/       # File download and processing
├── progress/          # Migration progress tracking
└── migration/         # Migration orchestration and execution
test/
├── unit/              # Unit tests for individual packages
├── integration/       # Integration tests
├── mocks/             # Mock implementations for testing
└── testdata/          # Test data and fixtures
```

## Installation

### Install from source

```bash
go install github.com/exileum/xenforo-to-gh-discussions@latest
```

### Clone and build

```bash
git clone https://github.com/exileum/xenforo-to-gh-discussions.git
cd xenforo-to-gh-discussions
make build
```

### Install dependencies

```bash
make deps
```

### Alternative: manual build

```bash
go build -o xenforo-to-gh-discussions .
```

## Prerequisites

- Go 1.24 or higher
- XenForo 2 with REST API enabled
- GitHub repository with Discussions enabled
- API credentials for both platforms

## Configuration

The tool supports both environment variables and default configuration. Environment variables take precedence.

### Environment Variables

```bash
# XenForo Configuration
export XENFORO_API_URL="https://your-forum.com/api"
export XENFORO_API_KEY="your_xenforo_api_key"
export XENFORO_API_USER="1"
export XENFORO_NODE_ID="1"

# GitHub Configuration
export GITHUB_TOKEN="your_github_token"
export GITHUB_REPO="owner/repository"

# Migration Settings
export MAX_RETRIES="3"
export PROGRESS_FILE="migration_progress.json"
export ATTACHMENTS_DIR="./attachments"
```

### Code-based Configuration

If not using environment variables, you can modify the defaults in `internal/config/config.go`:

```go
// Default values used when environment variables are not set
XenForo: XenForoConfig{
    APIURL:  "https://your-forum.com/api",
    APIKey:  "your_xenforo_api_key",
    APIUser: "1",
    NodeID:  1,
},
GitHub: GitHubConfig{
    Token:      "your_github_token",
    Repository: "owner/repository",
    Categories: map[int]string{
        1: "DIC_kwDOxxxxxxxx", // Map node IDs to category IDs
    },
},
```

### Setting up XenForo API

1. Enable REST API in XenForo Admin Panel
2. Create a superuser API key with read access to:
    - `/threads`
    - `/posts`
    - `/attachments`
3. Note the user ID of the admin account

### Setting up GitHub

1. Create a personal access token with scopes:
    - `repo`
    - `write:discussion`
2. Enable GitHub Discussions in your target repository
3. Create discussion categories as needed

### Getting GitHub Category IDs

You need to map your XenForo forum node IDs to GitHub Discussion category IDs in the configuration. Here are several ways to find the category IDs:

#### Method 1: GitHub CLI (Recommended)

```bash
# Replace OWNER and REPO with your repository details
gh api graphql -f query='
  query {
    repository(owner: "OWNER", name: "REPO") {
      discussionCategories(first: 100) {
        nodes {
          id
          name
          description
        }
      }
    }
  }'
```

#### Method 2: GitHub GraphQL Explorer

1. Go to [GitHub GraphQL Explorer](https://docs.github.com/en/graphql/overview/explorer)
2. Run this query (replace `OWNER` and `REPO`):

```graphql
query {
  repository(owner: "OWNER", name: "REPO") {
    discussionCategories(first: 100) {
      nodes {
        id
        name
        description
      }
    }
  }
}
```

#### Method 3: Check Pre-flight Output

Run the migration tool in dry-run mode - it will list valid category IDs during pre-flight checks:

```bash
./xenforo-to-gh-discussions --dry-run
```

#### Updating the Category Mapping

Once you have the category IDs, update the mapping in your configuration:

```go
Categories: map[int]string{
    1: "DIC_kwDOxxxxxxxx",  // General Discussion
    2: "DIC_kwDOyyyyyyyy",  // Q&A
    3: "DIC_kwDOzzzzzzzz",  // Announcements
}
```

**Note**: Threads from XenForo nodes not mapped in this configuration will be skipped during migration.

## Usage

### Basic Migration

```bash
make run
# or directly:
./build/xenforo-to-gh-discussions
```

### Dry Run Mode

Test the migration without making actual changes:

```bash
./build/xenforo-to-gh-discussions --dry-run
```

### Verbose Mode

See detailed output including converted content:

```bash
./build/xenforo-to-gh-discussions --dry-run --verbose
```

### Resume from Specific Thread

Resume migration from a specific thread ID:

```bash
./build/xenforo-to-gh-discussions --resume-from=123
```

### Command Line Options

- `--dry-run`: Run without making actual API calls
- `--verbose`: Enable detailed logging
- `--resume-from=ID`: Resume from specific thread ID

## Output Structure

### Attachments

Downloaded attachments are organized by file type:

```
./attachments/
├── png/
│   ├── attachment_123_screenshot.png
│   └── attachment_456_diagram.png
├── jpg/
│   └── attachment_789_photo.jpg
├── zip/
│   └── attachment_012_archive.zip
└── pdf/
    └── attachment_345_document.pdf
```

### Discussion Format

Each discussion includes metadata in the following format:

```markdown
---
Author: username
Posted: 2025-01-15 14:30:00 UTC
Original Thread ID: 12345
---

[Message content with converted Markdown]
```

## Progress Tracking

The tool automatically saves progress to `migration_progress.json`:

```json
{
  "last_thread_id": 456,
  "completed_threads": [123, 456],
  "failed_threads": [],
  "last_updated": 1642353000
}
```

Migration automatically resumes from the last successful thread if interrupted.

## Development

### Quick Start with Makefile

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

### Project Structure

The project follows Go best practices with a clean architecture:

- `main.go`: Application entry point
- `internal/`: Private application code organized by domain
- `test/`: All tests organized by type (unit, integration, mocks)

### Running Tests

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

### Build Commands

```bash
# Development build (with race detector)
make dev

# Production build
make build

# Install to $GOPATH/bin
make install

# Build for all platforms
make build-all

# Create release packages
make package
```

### Code Quality

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

### Development Workflow

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

### Code Quality

The codebase maintains high quality standards:

- **Cyclomatic Complexity**: All functions kept below 15 complexity
- **Package Organization**: Clear separation of concerns
- **Test Coverage**: Comprehensive unit and integration tests
- **Documentation**: Detailed README and architecture docs

## Troubleshooting

### Pre-flight Check Failures

- **XenForo API authentication failed**: Verify API key and user ID
- **GitHub Discussions not enabled**: Enable Discussions in repository settings
- **Invalid category ID**: Use the GraphQL query above to get valid category IDs

### Common Issues

1. **Rate Limiting**: The tool automatically handles rate limits with exponential backoff
2. **Large Attachments**: Consider increasing timeout values for large file downloads
3. **Memory Usage**: For forums with many threads, consider migrating in batches by updating `XENFORO_NODE_ID`

### Configuration Validation

The tool validates all configuration before starting:

```bash
# Check configuration without running migration
make run -- --dry-run
```

## Security Notes

- Never commit the script with actual API credentials
- Use environment variables for production deployments
- Ensure the attachment repository is public if you want images to display
- The tool includes path traversal protection for downloaded files

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [XenForo REST API Documentation](https://xenforo.com/docs/dev/rest-api/)
- [GitHub Discussions API](https://docs.github.com/en/graphql/guides/using-the-graphql-api-for-discussions)
- [Go Resty Library](https://github.com/go-resty/resty)
- [GitHub GraphQL Client](https://github.com/shurcooL/githubv4)
