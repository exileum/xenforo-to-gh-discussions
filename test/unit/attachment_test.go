package unit

import (
	"strings"
	"testing"

	"github.com/exileum/xenforo-to-gh-discussions/internal/attachments"
	"github.com/exileum/xenforo-to-gh-discussions/internal/xenforo"
)

func TestFileSanitizer(t *testing.T) {
	sanitizer := attachments.NewFileSanitizer()

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
	downloader := attachments.NewDownloader("./test_attachments", true, mockClient)

	attachments := []xenforo.Attachment{
		{
			AttachmentID: 1,
			Filename:     "test.png",
			ViewURL:      "https://example.com/1",
		},
	}

	// Test in dry-run mode (should not actually download)
	err := downloader.DownloadAttachments(attachments)
	if err != nil {
		t.Errorf("Dry run should not return error: %v", err)
	}
}

func TestReplaceAttachmentLinks(t *testing.T) {
	mockClient := &mockXenForoClient{}
	downloader := attachments.NewDownloader("./attachments", true, mockClient)

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

	// Should replace image with markdown image syntax
	if !strings.Contains(result, "![image.png](./png/attachment_1_image.png)") {
		t.Error("Should replace image attachment with markdown image syntax")
	}

	// Should replace document with markdown link syntax
	if !strings.Contains(result, "[document.pdf](./pdf/attachment_2_document.pdf)") {
		t.Error("Should replace document attachment with markdown link syntax")
	}
}
