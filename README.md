# XenForo to GitHub Discussions Migration Tool

[![Go Report Card](https://goreportcard.com/badge/github.com/exileum/xenforo-to-gh-discussions)](https://goreportcard.com/report/github.com/exileum/xenforo-to-gh-discussions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org)

A robust Go CLI tool to migrate forum threads, posts, and attachments from XenForo 2 to GitHub Discussions using their respective APIs.

## Features

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

## Installation

### Install from source

```bash
go install github.com/exileum/xenforo-to-gh-discussions@latest
```

### Clone and build

```bash
git clone https://github.com/exileum/xenforo-to-gh-discussions.git
cd xenforo-to-gh-discussions
go build -o xenforo-to-gh-discussions .
```

### Install dependencies manually

```bash
go mod download
```

## Prerequisites

- Go 1.24 or higher
- XenForo 2 with REST API enabled
- GitHub repository with Discussions enabled
- API credentials for both platforms

## Configuration

Edit the constants in `main.go`:

```go
var (
    XenForoAPIURL  = "https://your-forum.com/api"
    XenForoAPIKey  = "your_xenforo_api_key"
    XenForoAPIUser = "1"  // Admin user ID
    GitHubToken    = "your_github_token"
    GitHubRepo     = "owner/repository"
    TargetNodeID   = 1    // XenForo forum ID to migrate
)
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

You need to map your XenForo forum node IDs to GitHub Discussion category IDs in the `NodeToCategory` map. Here are several ways to find the category IDs:

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

1. Go to https://docs.github.com/en/graphql/overview/explorer
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

#### Method 4: Browser Developer Tools

1. Go to your repository's Discussions tab
2. Open browser developer tools (F12)
3. Look at network requests when clicking on a category
4. Category IDs appear in the URL or API responses

#### Understanding Category IDs

GitHub Discussion category IDs have the format `DIC_kwDO` followed by encoded characters (e.g., `DIC_kwDOxxxxxxxx`).

#### Updating the Mapping

Once you have the category IDs, update the `NodeToCategory` map in `main.go`:

```go
var NodeToCategory = map[int]string{
    1: "DIC_kwDOxxxxxxxx",  // General Discussion
    2: "DIC_kwDOyyyyyyyy",  // Q&A
    3: "DIC_kwDOzzzzzzzz",  // Announcements
    // Add more mappings as needed
}
```

**Note**: Threads from XenForo nodes not mapped in this configuration will be skipped during migration.

## Usage

### Basic Migration

```bash
./xenforo-to-gh-discussions
```

### Dry Run Mode

Test the migration without making actual changes:

```bash
./xenforo-to-gh-discussions --dry-run
```

### Verbose Mode

See detailed output including converted content:

```bash
./xenforo-to-gh-discussions --dry-run --verbose
```

### Resume from Specific Thread

Resume migration from a specific thread ID:

```bash
./xenforo-to-gh-discussions --resume-from=123
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

### Running Tests

```bash
go test -v ./...
```

### Running with Coverage

```bash
go test -cover ./...
```

### Build for Multiple Platforms

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o xenforo-to-gh-discussions-linux .

# Windows
GOOS=windows GOARCH=amd64 go build -o xenforo-to-gh-discussions.exe .

# macOS
GOOS=darwin GOARCH=amd64 go build -o xenforo-to-gh-discussions-macos .
```

## Troubleshooting

### Pre-flight Check Failures

- **XenForo API authentication failed**: Verify API key and user ID
- **GitHub Discussions not enabled**: Enable Discussions in repository settings
- **Invalid category ID**: Use the GraphQL query above to get valid category IDs

### Common Issues

1. **Rate Limiting**: The tool automatically handles rate limits with exponential backoff
2. **Large Attachments**: Consider increasing timeout values for large file downloads
3. **Memory Usage**: For forums with many threads, consider migrating in batches by updating `TargetNodeID`

## Security Notes

- Never commit the script with actual API credentials
- Use environment variables for production deployments
- Ensure the attachment repository is public if you want images to display

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
- [GitHub Discussions API](https://docs.github.com/en/rest/discussions)
- [Go Resty Library](https://github.com/go-resty/resty)
- [GitHub GraphQL Client](https://github.com/shurcooL/githubv4)
