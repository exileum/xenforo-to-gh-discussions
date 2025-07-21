// Package attachments provides functionality for downloading and managing
// file attachments from XenForo forums. It handles secure file downloads,
// URL sanitization, and filesystem management for migration purposes.
package attachments

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/exileum/xenforo-to-gh-discussions/internal/xenforo"
)

type Downloader struct {
	sanitizer      *FileSanitizer
	attachmentsDir string
	dryRun         bool
	client         XenForoDownloader
	rateLimitDelay time.Duration
}

type XenForoDownloader interface {
	DownloadAttachment(url, filepath string) error
}

func NewDownloader(attachmentsDir string, dryRun bool, client XenForoDownloader, rateLimitDelay time.Duration) *Downloader {
	return &Downloader{
		sanitizer:      NewFileSanitizer(),
		attachmentsDir: attachmentsDir,
		dryRun:         dryRun,
		client:         client,
		rateLimitDelay: rateLimitDelay,
	}
}

func (d *Downloader) DownloadAttachments(ctx context.Context, attachments []xenforo.Attachment) error {
	for _, attachment := range attachments {
		// Check context cancellation before each attachment
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if d.dryRun {
			log.Printf("    [DRY-RUN] Would download: %s", attachment.Filename)
			continue
		}

		if err := d.downloadSingle(ctx, attachment); err != nil {
			log.Printf("    ✗ Failed to download %s: %v", attachment.Filename, err)
			continue
		}
	}
	return nil
}

func (d *Downloader) downloadSingle(ctx context.Context, attachment xenforo.Attachment) error {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Determine file extension and create directory
	ext := d.getFileExtension(attachment.Filename)
	dir := filepath.Join(d.attachmentsDir, ext)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Generate safe filename
	sanitizedFilename, err := d.sanitizer.SanitizeFilename(ctx, attachment.Filename)
	if err != nil {
		return fmt.Errorf("failed to sanitize filename: %w", err)
	}
	filename := fmt.Sprintf("attachment_%d_%s", attachment.AttachmentID, sanitizedFilename)
	filePath := filepath.Join(dir, filename)

	// Validate path security
	if err := d.sanitizer.ValidatePath(ctx, filePath, dir); err != nil {
		return fmt.Errorf("security violation: file path escapes directory")
	}

	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		log.Printf("    ⏭ Skipped (already exists): %s", filename)
		return nil
	}

	// Download file
	if err := d.client.DownloadAttachment(attachment.DirectURL, filePath); err != nil {
		return err
	}

	log.Printf("    ✓ Downloaded: %s", filename)

	// Configurable rate limiting with context awareness
	if d.rateLimitDelay > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(d.rateLimitDelay):
		}
	}

	return nil
}

func (d *Downloader) getFileExtension(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return "unknown"
	}
	return strings.TrimPrefix(ext, ".")
}

func (d *Downloader) ReplaceAttachmentLinks(ctx context.Context, message string, attachments []xenforo.Attachment) (string, error) {
	for _, attachment := range attachments {
		// Check context cancellation before each attachment
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		sanitizedFilename, err := d.sanitizer.SanitizeFilename(ctx, attachment.Filename)
		if err != nil {
			return "", fmt.Errorf("failed to sanitize filename: %w", err)
		}
		ext := d.getFileExtension(sanitizedFilename)

		filename := fmt.Sprintf("attachment_%d_%s", attachment.AttachmentID, sanitizedFilename)
		relativePath := fmt.Sprintf("./%s/%s", ext, filename)

		// Determine if it's an image
		isImage := d.isImageFile(ext)

		// Replace BB-code with appropriate markdown
		bbCode := fmt.Sprintf("[ATTACH=%d]", attachment.AttachmentID)
		bbCodeFull := fmt.Sprintf("[ATTACH=full]%d[/ATTACH]", attachment.AttachmentID)

		var markdownLink string
		if isImage {
			markdownLink = fmt.Sprintf("![%s](%s)", sanitizedFilename, relativePath)
		} else {
			markdownLink = fmt.Sprintf("[%s](%s)", sanitizedFilename, relativePath)
		}

		message = strings.ReplaceAll(message, bbCode, markdownLink)
		message = strings.ReplaceAll(message, bbCodeFull, markdownLink)
	}

	// Log any remaining unhandled attach codes
	remaining := regexp.MustCompile(`\[ATTACH[^]]*\]`).FindAllString(message, -1)
	for _, code := range remaining {
		log.Printf("    ⚠ Unhandled attachment code: %s", code)
	}

	return message, nil
}

func (d *Downloader) isImageFile(ext string) bool {
	imageExtensions := map[string]bool{
		"png":  true,
		"jpg":  true,
		"jpeg": true,
		"gif":  true,
		"webp": true,
	}
	return imageExtensions[ext]
}
