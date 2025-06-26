package attachments

import (
	"strings"
	"testing"
	"time"

	"github.com/exileum/xenforo-to-gh-discussions/internal/xenforo"
)

func TestFileSanitizer(t *testing.T) {
	sanitizer := NewFileSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal filename",
			input:    "test.txt",
			expected: "test.txt",
		},
		{
			name:     "Filename with unsafe characters",
			input:    "test<>:|*.txt",
			expected: "test_____.txt",
		},
		{
			name:     "Empty filename",
			input:    "",
			expected: "unnamed_file",
		},
		{
			name:     "Whitespace only",
			input:    "   ",
			expected: "unnamed_file",
		},
		{
			name:     "Path traversal attempt",
			input:    "../../../etc/passwd",
			expected: "passwd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

type mockXenForoClient struct {
	downloadError error
}

func (m *mockXenForoClient) DownloadAttachment(url, filepath string) error {
	return m.downloadError
}

func TestDownloader(t *testing.T) {
	mockClient := &mockXenForoClient{}
	tempDir := t.TempDir()
	downloader := NewDownloader(tempDir, true, mockClient, 100*time.Millisecond)

	attachments := []xenforo.Attachment{
		{
			AttachmentID: 1,
			Filename:     "test.png",
			ViewURL:      "https://example.com/1",
		},
	}

	// Test in dry-run mode (should not download)
	err := downloader.DownloadAttachments(attachments)
	if err != nil {
		t.Errorf("Dry run should not return error: %v", err)
	}
}

func TestReplaceAttachmentLinks(t *testing.T) {
	mockClient := &mockXenForoClient{}
	tempDir := t.TempDir()
	downloader := NewDownloader(tempDir, true, mockClient, 0) // No rate limiting for test

	message := "Check out this image: [ATTACH=1] and this file: [ATTACH=full]2[/ATTACH]"
	attachments := []xenforo.Attachment{
		{
			AttachmentID: 1,
			Filename:     "image.png",
			ViewURL:      "https://example.com/1",
		},
		{
			AttachmentID: 2,
			Filename:     "document.pdf",
			ViewURL:      "https://example.com/2",
		},
	}

	result := downloader.ReplaceAttachmentLinks(message, attachments)

	// Should replace image with Markdown image syntax
	if !strings.Contains(result, "![image.png](./png/attachment_1_image.png)") {
		t.Error("Should replace image attachment with markdown image syntax")
	}

	// Should replace a document with Markdown link syntax
	if !strings.Contains(result, "[document.pdf](./pdf/attachment_2_document.pdf)") {
		t.Error("Should replace document attachment with markdown link syntax")
	}
}

func TestValidatePath(t *testing.T) {
	sanitizer := NewFileSanitizer()

	tests := []struct {
		name      string
		filePath  string
		baseDir   string
		shouldErr bool
		errMsg    string
	}{
		{
			name:      "Valid path within base directory",
			filePath:  "/tmp/safe/file.txt",
			baseDir:   "/tmp/safe",
			shouldErr: false,
		},
		{
			name:      "Valid subdirectory path",
			filePath:  "/tmp/safe/subdir/file.txt",
			baseDir:   "/tmp/safe",
			shouldErr: false,
		},
		{
			name:      "Path equals base directory",
			filePath:  "/tmp/safe",
			baseDir:   "/tmp/safe",
			shouldErr: false,
		},
		{
			name:      "Simple directory traversal with ../",
			filePath:  "/tmp/safe/../../../etc/passwd",
			baseDir:   "/tmp/safe",
			shouldErr: true,
			errMsg:    "path traversal detected",
		},
		{
			name:      "Directory traversal at start",
			filePath:  "../../../etc/passwd",
			baseDir:   "/tmp/safe",
			shouldErr: true,
			errMsg:    "path traversal detected",
		},
		{
			name:      "Directory traversal in middle",
			filePath:  "/tmp/safe/subdir/../../../etc/passwd",
			baseDir:   "/tmp/safe",
			shouldErr: true,
			errMsg:    "path traversal detected",
		},
		{
			name:      "Directory traversal at end (resolves to parent)",
			filePath:  "/tmp/safe/../",
			baseDir:   "/tmp/safe",
			shouldErr: true,
			errMsg:    "path traversal detected",
		},
		{
			name:      "Single .. path",
			filePath:  "..",
			baseDir:   "/tmp/safe",
			shouldErr: true,
			errMsg:    "path traversal detected",
		},
		{
			name:      "Multiple .. segments",
			filePath:  "../../..",
			baseDir:   "/tmp/safe",
			shouldErr: true,
			errMsg:    "path traversal detected",
		},
		{
			name:      "Encoded directory traversal (should be normalized)",
			filePath:  "/tmp/safe/subdir/%2E%2E/file.txt",
			baseDir:   "/tmp/safe",
			shouldErr: false, // This should be valid as %2E%2E won't be decoded by filepath
		},
		{
			name:      "Symlink-style traversal attempt",
			filePath:  "/tmp/safe/link/../../../etc/passwd",
			baseDir:   "/tmp/safe",
			shouldErr: true,
			errMsg:    "path traversal detected",
		},
		{
			name:      "Windows-style path separators (on Unix)",
			filePath:  "/tmp/safe/subdir\\..\\..\\etc\\passwd",
			baseDir:   "/tmp/safe",
			shouldErr: false, // On Unix, backslashes are treated as regular filename characters
		},
		{
			name:      "Valid relative path within base directory",
			filePath:  "/tmp/safe/file.txt",
			baseDir:   "/tmp",
			shouldErr: false,
		},
		{
			name:      "Complex valid nested path",
			filePath:  "/tmp/safe/deep/nested/directory/file.txt",
			baseDir:   "/tmp/safe",
			shouldErr: false,
		},
		{
			name:      "Path with .. that gets cleaned but still escapes",
			filePath:  "/tmp/safe/subdir/../../etc/passwd",
			baseDir:   "/tmp/safe",
			shouldErr: true,
			errMsg:    "path traversal detected",
		},
		{
			name:      "Valid file within subdirectory that gets cleaned",
			filePath:  "/tmp/safe/subdir/../file.txt",
			baseDir:   "/tmp/safe",
			shouldErr: false, // This resolves to /tmp/safe/file.txt which is valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := sanitizer.ValidatePath(tt.filePath, tt.baseDir)

			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error but got none for path: %s", tt.filePath)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error message to contain %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v for path: %s", err, tt.filePath)
				}
			}
		})
	}
}

func TestDownloaderRateLimiting(t *testing.T) {
	tests := []struct {
		name           string
		rateLimitDelay time.Duration
		expectMinTime  time.Duration
		expectMaxTime  time.Duration
	}{
		{
			name:           "No rate limiting (zero delay)",
			rateLimitDelay: 0,
			expectMinTime:  0,
			expectMaxTime:  50 * time.Millisecond, // Allow some processing overhead
		},
		{
			name:           "Short rate limiting",
			rateLimitDelay: 100 * time.Millisecond,
			expectMinTime:  90 * time.Millisecond,  // Allow some timing variance
			expectMaxTime:  300 * time.Millisecond, // Allow extra overhead for loaded CI machines
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockXenForoClient{}
			tempDir := t.TempDir()
			downloader := NewDownloader(tempDir, false, mockClient, tt.rateLimitDelay) // Don't use dry-run for timing test

			attachments := []xenforo.Attachment{
				{
					AttachmentID: 1,
					Filename:     "test.png",
					ViewURL:      "https://example.com/1",
				},
			}

			// Measure execution time
			start := time.Now()
			err := downloader.DownloadAttachments(attachments)
			elapsed := time.Since(start)

			if err != nil {
				t.Errorf("DownloadAttachments should not return error: %v", err)
			}

			// Verify timing expectations
			if elapsed < tt.expectMinTime {
				t.Errorf("Expected minimum time %v, but took %v", tt.expectMinTime, elapsed)
			}
			if elapsed > tt.expectMaxTime {
				t.Errorf("Expected maximum time %v, but took %v", tt.expectMaxTime, elapsed)
			}
		})
	}
}
