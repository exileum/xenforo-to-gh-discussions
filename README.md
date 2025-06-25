# XenForo to GitHub Discussions Migration Tool

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/exileum/xenforo-to-gh-discussions)](https://goreportcard.com/report/github.com/exileum/xenforo-to-gh-discussions)
[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org)

A robust, well-architected Go CLI tool to migrate forum threads, posts, and attachments from XenForo 2 to GitHub Discussions using their respective APIs.

## Features

- **Interactive Configuration**: User-friendly setup with guided prompts — no manual configuration needed!
- **Clean Architecture**: Modular design with separate packages for different concerns
- **Full Content Migration**: Complete BB-code to Markdown conversion with smart attachment handling
    - Text formatting, quotes, code blocks, spoilers, lists, and media embeds
    - Secure file downloads organized by type with direct image embedding
    - Empty tag detection and Markdown link preservation
- **Robust Migration Process**:
    - Interactive category selection with dry-run preview
    - Progress tracking with JSON persistence for resumable migrations
    - Interactive error handling with retry/skip/abort options
    - Rate limiting compliance with exponential backoff
    - Thread-level failure handling to prevent partial migrations
- **Flexible Deployment**: Interactive mode for manual setup or environment variables for automation

## Architecture

📋 **[View detailed architecture and technical documentation →](ARCHITECTURE.md)**

The migration tool follows a clean architecture pattern with well-separated concerns:

```
cmd/
└── xenforo-to-gh-discussions/  # Application entry point
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

## Installation

### Install from source

```bash
go install github.com/exileum/xenforo-to-gh-discussions@latest
```

### Clone and build

```bash
git clone https://github.com/exileum/xenforo-to-gh-discussions.git
cd xenforo-to-gh-discussions
make deps
make build
```

### Alternative: manual build

```bash
go build -o xenforo-to-gh-discussions ./cmd/xenforo-to-gh-discussions
```

## Prerequisites

> [!IMPORTANT]
> Before you begin, ensure you have:
> - Go 1.24 or higher
> - XenForo 2 with REST API enabled
> - GitHub repository with Discussions enabled
> - API credentials for both platforms

## Usage

### Interactive Mode (Recommended)

> [!TIP]
> The tool features an interactive setup that guides you through the entire configuration process - no manual configuration needed!

```bash
make run
# or directly:
./build/xenforo-to-gh-discussions
```

The interactive mode will:
1. **Prompt for XenForo API credentials** and validate them immediately
2. **Show available forum categories** with thread counts
3. **Prompt for GitHub token and repository** and validate permissions
4. **Display available GitHub Discussion categories**
5. **Offer a dry-run preview** with migration statistics
6. **Guide you through each category migration** with retry/skip/abort options
7. **Ask if you want to migrate additional categories** when done

### Non-Interactive Mode (Automation)

For automated deployments, use environment variables with the `--non-interactive` flag:

```bash
./build/xenforo-to-gh-discussions --non-interactive
```

## Configuration

The tool supports both interactive prompts (recommended) and environment variables for automation.

### Environment Variables (for automation)

```bash
# XenForo Configuration
export XENFORO_API_URL="https://your-forum.com/api"
export XENFORO_API_KEY="your_xenforo_api_key"
export XENFORO_API_USER="1"
export XENFORO_NODE_ID="42" # XenForo category/node ID to migrate from

# GitHub Configuration
export GITHUB_TOKEN="your_github_token"
export GITHUB_REPO="owner/repository"
export GITHUB_CATEGORY_ID="DIC_kwDOxxxxxxxx" # GitHub Discussion category ID to migrate to

# Migration Settings
export MAX_RETRIES="3"
export ATTACHMENTS_DIR="./attachments"
export PROGRESS_FILE="migration_progress.json" # Optional: custom progress file path
export ATTACHMENT_RATE_LIMIT_DELAY="500ms" # Optional: delay between downloads
```

> [!NOTE]
> Environment variables support single-category migration only. For multi-category scenarios, modify the source code to set custom category mappings.
> 
> **Example**: To migrate 3 categories, modify `internal/config/config.go` line 58:
> ```go
> Categories: map[int]string{
>     42: "DIC_kwDOxxxxxxxx", // General Discussion -> General
>     43: "DIC_kwDOyyyyyyyy", // News -> Announcements  
>     56: "DIC_kwDOzzzzzzzz", // Support -> Help
> },
> ```

> [!TIP]
> Interactive mode eliminates the need for manual configuration — it walks you through setup and validates everything automatically!

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

### Command Line Options

- `--dry-run`: Run without making actual API calls (can be combined with interactive mode)
- `--verbose`: Enable detailed logging
- `--resume-from=ID`: Resume from specific thread ID  
- `--non-interactive`: Use environment variables instead of prompts (for automation)

### Examples

```bash
# Interactive mode with dry-run
./build/xenforo-to-gh-discussions --dry-run

# Verbose interactive mode
./build/xenforo-to-gh-discussions --verbose

# Resume from specific thread (interactive)
./build/xenforo-to-gh-discussions --resume-from=123

# Automated mode (requires env vars)
./build/xenforo-to-gh-discussions --non-interactive
```

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

> [!TIP]
> **For developers**: See the comprehensive [development guide](DEVELOPMENT.md) for detailed information on:
> - Project structure and architecture guidelines
> - Testing strategies and running tests
> - Build commands and development workflow
> - Code quality standards and contribution guidelines

## Troubleshooting

### Pre-flight Check Failures

- **XenForo API authentication failed**: Verify API key and user ID
- **GitHub Discussions not enabled**: Enable Discussions in repository settings
- **Invalid category ID**: Use the GraphQL query above to get valid category IDs

### Common Issues

> [!CAUTION]
> Be aware of these potential issues:

1. **Rate Limiting**: The tool automatically handles rate limits with exponential backoff
2. **Large Attachments**: Consider increasing timeout values for large file downloads
3. **Memory Usage**: For forums with many threads, consider migrating in batches

### Configuration Validation

The tool validates all configurations before starting:

```bash
# Check configuration without running migration
make run -- --dry-run
```

## Security Notes

> [!WARNING]
> Never commit the script with actual API credentials to version control!

> [!IMPORTANT]
> Security best practices:
> - Use environment variables for production deployments
> - Ensure the attachment repository is public if you want images to display
> - The tool includes path traversal protection for downloaded files


## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [XenForo REST API Documentation](https://xenforo.com/docs/dev/rest-api/)
- [GitHub Discussions API](https://docs.github.com/en/graphql/guides/using-the-graphql-api-for-discussions)
- [Go Resty Library](https://github.com/go-resty/resty)
- [GitHub GraphQL Client](https://github.com/shurcooL/githubv4)
