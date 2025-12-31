package attachments

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/exileum/xenforo-to-gh-discussions/internal/xenforo"
)

// mockContextClient simulates a slow download that can be cancelled
type mockContextClient struct {
	downloadDelay time.Duration
}

func (m *mockContextClient) DownloadAttachment(url, filepath string) error {
	// Simulate a slow download
	time.Sleep(m.downloadDelay)
	return nil
}

func TestDownloadAttachments_ContextCancellation(t *testing.T) {
	t.Run("cancellation before download", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		downloader := NewDownloader("/tmp/test", false, &mockContextClient{}, 0)
		attachments := []xenforo.Attachment{
			{AttachmentID: 1, Filename: "test.jpg", DirectURL: "http://example.com/1.jpg"},
		}

		err := downloader.DownloadAttachments(ctx, attachments)
		if err == nil {
			t.Error("Expected error from cancelled context")
		}

		if !strings.Contains(err.Error(), "attachment download cancelled") {
			t.Errorf("Expected 'attachment download cancelled' in error, got: %v", err)
		}
	})

	t.Run("cancellation during download", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		// Set up multiple attachments
		attachments := []xenforo.Attachment{
			{AttachmentID: 1, Filename: "test1.jpg", DirectURL: "http://example.com/1.jpg"},
			{AttachmentID: 2, Filename: "test2.jpg", DirectURL: "http://example.com/2.jpg"},
			{AttachmentID: 3, Filename: "test3.jpg", DirectURL: "http://example.com/3.jpg"},
		}

		downloader := NewDownloader("/tmp/test", false, &mockContextClient{downloadDelay: 100 * time.Millisecond}, 0)

		// Cancel after a short delay
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		start := time.Now()
		err := downloader.DownloadAttachments(ctx, attachments)
		elapsed := time.Since(start)

		if err == nil {
			t.Error("Expected error from cancelled context")
		}

		// Should not process all attachments
		if elapsed > 300*time.Millisecond {
			t.Errorf("Download should be cancelled early, took: %v", elapsed)
		}
	})

	t.Run("timeout during download", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		defer cancel()

		attachments := []xenforo.Attachment{
			{AttachmentID: 1, Filename: "test1.jpg", DirectURL: "http://example.com/1.jpg"},
			{AttachmentID: 2, Filename: "test2.jpg", DirectURL: "http://example.com/2.jpg"},
			{AttachmentID: 3, Filename: "test3.jpg", DirectURL: "http://example.com/3.jpg"},
		}

		downloader := NewDownloader("/tmp/test", false, &mockContextClient{downloadDelay: 100 * time.Millisecond}, 0)

		err := downloader.DownloadAttachments(ctx, attachments)
		if err == nil {
			t.Error("Expected timeout error")
		}
	})
}

func TestReplaceAttachmentLinks_ContextCancellation(t *testing.T) {
	t.Run("cancellation during replacement", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		downloader := NewDownloader("/tmp/test", false, &mockContextClient{}, 0)
		attachments := []xenforo.Attachment{
			{AttachmentID: 1, Filename: "test.jpg", DirectURL: "http://example.com/1.jpg"},
		}

		message := "Check out [ATTACH]1[/ATTACH]"
		_, err := downloader.ReplaceAttachmentLinks(ctx, message, attachments)

		if err == nil {
			t.Error("Expected error from cancelled context")
		}

		if !strings.Contains(err.Error(), "attachment link replacement cancelled") {
			t.Errorf("Expected 'attachment link replacement cancelled' in error, got: %v", err)
		}
	})
}

func TestSanitizeFilename_ContextCancellation(t *testing.T) {
	t.Run("cancellation during sanitization", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		sanitizer := NewFileSanitizer()
		_, err := sanitizer.SanitizeFilename(ctx, "test.jpg")

		if err == nil {
			t.Error("Expected error from cancelled context")
		}

		if !strings.Contains(err.Error(), "filename sanitization cancelled") {
			t.Errorf("Expected 'filename sanitization cancelled' in error, got: %v", err)
		}
	})
}

func TestValidatePath_ContextCancellation(t *testing.T) {
	t.Run("cancellation during validation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		sanitizer := NewFileSanitizer()
		err := sanitizer.ValidatePath(ctx, "/tmp/test/file.jpg", "/tmp/test")

		if err == nil {
			t.Error("Expected error from cancelled context")
		}

		if !strings.Contains(err.Error(), "path validation cancelled") {
			t.Errorf("Expected 'path validation cancelled' in error, got: %v", err)
		}
	})
}
