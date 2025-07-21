package attachments

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

type FileSanitizer struct{}

func NewFileSanitizer() *FileSanitizer {
	return &FileSanitizer{}
}

func (s *FileSanitizer) SanitizeFilename(ctx context.Context, filename string) (string, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	if filename == "" {
		return "unnamed_file", nil
	}

	// Check if filename is local (no path traversal)
	if !filepath.IsLocal(filename) {
		// Extract just the base filename if path traversal detected
		filename = filepath.Base(filename)
	}

	// Replace filesystem-unsafe characters with underscores
	unsafe := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range unsafe {
		filename = strings.ReplaceAll(filename, char, "_")
	}

	// Trim whitespace and ensure not empty
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return "unnamed_file", nil
	}

	return filename, nil
}

func (s *FileSanitizer) ValidatePath(ctx context.Context, filePath, baseDir string) error {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Clean and normalize both paths
	cleanFilePath := filepath.Clean(filePath)
	cleanBaseDir := filepath.Clean(baseDir)

	// Get absolute paths for security check
	absBaseDir, err := filepath.Abs(cleanBaseDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for base directory: %w", err)
	}

	absFilePath, err := filepath.Abs(cleanFilePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for file: %w", err)
	}

	// Use filepath.Rel to compute the relative path from baseDir to filePath
	relPath, err := filepath.Rel(absBaseDir, absFilePath)
	if err != nil {
		return fmt.Errorf("failed to compute relative path: %w", err)
	}

	// Check if the relative path escapes the base directory
	// The path escapes if it contains ".." segments in any form
	pathSegments := strings.Split(relPath, string(filepath.Separator))
	for _, segment := range pathSegments {
		if segment == ".." {
			return errors.New("path traversal detected: file path escapes base directory")
		}
	}

	// Additional check: ensure the relative path doesn't start with "/" (absolute path)
	if strings.HasPrefix(relPath, string(filepath.Separator)) {
		return errors.New("invalid relative path: path is absolute")
	}

	return nil
}
