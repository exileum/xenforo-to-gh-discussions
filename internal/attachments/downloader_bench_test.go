package attachments

import (
	"path/filepath"
	"testing"

	"github.com/exileum/xenforo-to-gh-discussions/internal/xenforo"
)

func BenchmarkSanitizer_SanitizeFilename(b *testing.B) {
	sanitizer := NewFileSanitizer()
	filename := "test-file with spaces & special chars!@#$%^&*()_+.jpg"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sanitizer.SanitizeFilename(filename)
	}
}

func BenchmarkSanitizer_SanitizeFilename_Complex(b *testing.B) {
	sanitizer := NewFileSanitizer()
	// Complex filename with various problematic characters
	filename := "../../etc/passwd/../../../dangerous file with unicode 文件名.exe"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sanitizer.SanitizeFilename(filename)
	}
}

func BenchmarkSanitizer_SanitizeFilename_PathTraversal(b *testing.B) {
	sanitizer := NewFileSanitizer()
	// Multiple path traversal attempts
	filename := "../../../../../../../etc/passwd"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sanitizer.SanitizeFilename(filename)
	}
}

func BenchmarkSanitizer_SanitizeFilename_LongFilename(b *testing.B) {
	sanitizer := NewFileSanitizer()
	// Very long filename that exceeds typical filesystem limits
	longName := ""
	for i := 0; i < 50; i++ {
		longName += "very_long_filename_segment_"
	}
	longName += ".jpg"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sanitizer.SanitizeFilename(longName)
	}
}

func BenchmarkSanitizer_ValidatePath_Safe(b *testing.B) {
	sanitizer := NewFileSanitizer()
	safePath := filepath.Join("/tmp/test_attachments", "subfolder", "safe_file.jpg")
	baseDir := "/tmp/test_attachments"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sanitizer.ValidatePath(safePath, baseDir)
	}
}

func BenchmarkSanitizer_ValidatePath_Dangerous(b *testing.B) {
	sanitizer := NewFileSanitizer()
	dangerousPath := "/tmp/test_attachments/../../../etc/passwd"
	baseDir := "/tmp/test_attachments"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sanitizer.ValidatePath(dangerousPath, baseDir)
	}
}

func BenchmarkSanitizer_ValidatePath_Complex(b *testing.B) {
	sanitizer := NewFileSanitizer()
	// Complex path with multiple traversal attempts and normalization needs
	complexPath := "/tmp/test_attachments/./folder/../other_folder/./../../etc/passwd"
	baseDir := "/tmp/test_attachments"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sanitizer.ValidatePath(complexPath, baseDir)
	}
}

// Benchmark attachment processing workflow
func BenchmarkAttachmentWorkflow(b *testing.B) {
	sanitizer := NewFileSanitizer()

	attachments := []xenforo.Attachment{
		{AttachmentID: 1, Filename: "test file (1).jpg", DirectURL: "https://example.com/1.jpg"},
		{AttachmentID: 2, Filename: "../dangerous.exe", DirectURL: "https://example.com/2.exe"},
		{AttachmentID: 3, Filename: "normal_file.pdf", DirectURL: "https://example.com/3.pdf"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate the filename sanitization workflow
		for _, attachment := range attachments {
			_ = sanitizer.SanitizeFilename(attachment.Filename)
		}
	}
}
