package attachments

import (
	"path/filepath"
	"strings"
)

type FileSanitizer struct{}

func NewFileSanitizer() *FileSanitizer {
	return &FileSanitizer{}
}

func (s *FileSanitizer) SanitizeFilename(filename string) string {
	if filename == "" {
		return "unnamed_file"
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
		return "unnamed_file"
	}

	return filename
}

func (s *FileSanitizer) ValidatePath(filePath, baseDir string) error {
	// Get absolute paths for security check
	absDir, err := filepath.Abs(baseDir)
	if err != nil {
		return err
	}

	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}

	// Ensure file path doesn't escape the base directory
	if !strings.HasPrefix(absFilePath, absDir+string(filepath.Separator)) && absFilePath != absDir {
		return filepath.ErrBadPattern
	}

	return nil
}
