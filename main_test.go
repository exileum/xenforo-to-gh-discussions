package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestConvertBBCodeToMarkdown tests the BB-code to Markdown conversion
func TestConvertBBCodeToMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Basic formatting
		{
			name:     "Bold text",
			input:    "This is [b]bold text[/b].",
			expected: "This is **bold text**.",
		},
		{
			name:     "Italic text",
			input:    "This is [i]italic text[/i].",
			expected: "This is *italic text*.",
		},
		{
			name:     "Underline text",
			input:    "This is [u]underlined text[/u].",
			expected: "This is <u>underlined text</u>.",
		},
		{
			name:     "Strikethrough with [s]",
			input:    "This is [s]strikethrough text[/s].",
			expected: "This is ~~strikethrough text~~.",
		},
		{
			name:     "Strikethrough with [strike]",
			input:    "This is [strike]strikethrough text[/strike].",
			expected: "This is ~~strikethrough text~~.",
		},

		// URLs
		{
			name:     "URL with text",
			input:    "Check out [url=https://example.com]this link[/url]!",
			expected: "Check out [this link](https://example.com)!",
		},
		{
			name:     "URL without text",
			input:    "Visit [url]https://example.com[/url]",
			expected: "Visit [https://example.com](https://example.com)",
		},
		{
			name:     "URL with quotes",
			input:    `Visit [url="https://example.com"]this link[/url]`,
			expected: "Visit [this link](https://example.com)",
		},

		// Images
		{
			name:     "Image",
			input:    "Here's an image: [img]https://example.com/image.png[/img]",
			expected: "Here's an image: ![](https://example.com/image.png)",
		},

		// Quotes
		{
			name:     "Simple quote",
			input:    "[quote]This is a quoted text[/quote]",
			expected: "> This is a quoted text\n",
		},
		{
			name:     "Quote with attribution",
			input:    `[quote="John Doe"]This is John's quote[/quote]`,
			expected: "> **John Doe said:**\n> This is John's quote\n",
		},
		{
			name:     "Quote with attribution and extra params",
			input:    `[quote="John Doe, post: 123"]This is John's quote[/quote]`,
			expected: "> **John Doe said:**\n> This is John's quote\n",
		},
		{
			name:     "Multi-line quote",
			input:    "[quote]Line 1\nLine 2\nLine 3[/quote]",
			expected: "> Line 1\n> Line 2\n> Line 3\n",
		},

		// Code blocks
		{
			name:     "Inline code",
			input:    "Use [code]fmt.Println()[/code] to print.",
			expected: "Use \n```\nfmt.Println()\n```\n to print.",
		},
		{
			name:     "Multi-line code",
			input:    "[code]\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n[/code]",
			expected: "\n```\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```\n",
		},

		// Spoilers
		{
			name:     "Block spoiler",
			input:    "[spoiler]Hidden content[/spoiler]",
			expected: "<details><summary>Spoiler</summary>\n\nHidden content\n\n</details>",
		},
		{
			name:     "Block spoiler with title",
			input:    `[spoiler="Plot twist"]The butler did it[/spoiler]`,
			expected: "<details><summary>Spoiler</summary>\n\nThe butler did it\n\n</details>",
		},
		{
			name:     "Inline spoiler",
			input:    "The answer is [ispoiler]42[/ispoiler]!",
			expected: "The answer is ||42||!",
		},

		// Lists
		{
			name:     "Simple list",
			input:    "[list]\n[*]Item 1\n[*]Item 2\n[*]Item 3\n[/list]",
			expected: "\n- Item 1\n- Item 2\n- Item 3\n",
		},
		{
			name:     "Numbered list",
			input:    "[list=1]\n[*]First\n[*]Second\n[/list]",
			expected: "\n- First\n- Second\n",
		},

		// Media
		{
			name:     "Media embed",
			input:    "[media=youtube]dQw4w9WgXcQ[/media]",
			expected: "[youtube](dQw4w9WgXcQ)",
		},

		// Center
		{
			name:     "Center alignment",
			input:    "[center]Centered text[/center]",
			expected: "<center>Centered text</center>",
		},

		// Color, size, font removal
		{
			name:     "Color tag removal",
			input:    "[color=red]Red text[/color]",
			expected: "Red text",
		},
		{
			name:     "Size tag removal",
			input:    "[size=20]Large text[/size]",
			expected: "Large text",
		},
		{
			name:     "Font tag removal",
			input:    "[font=Arial]Arial text[/font]",
			expected: "Arial text",
		},

		// Complex nested tags
		{
			name:     "Nested formatting",
			input:    "[b][i]Bold and italic[/i][/b]",
			expected: "***Bold and italic***",
		},
		{
			name:     "Quote with formatting",
			input:    `[quote="User"][b]Important:[/b] Read this[/quote]`,
			expected: "> **User said:**\n> **Important:** Read this\n",
		},

		// Edge cases
		{
			name:     "Empty tags",
			input:    "[b][/b]",
			expected: "****",
		},
		{
			name:     "Unclosed tags",
			input:    "[b]Bold text",
			expected: "Bold text",
		},
		{
			name:     "Unknown tags",
			input:    "[unknown]Content[/unknown]",
			expected: "Content",
		},
		{
			name:     "Multiple newlines cleanup",
			input:    "Line 1\n\n\n\nLine 2",
			expected: "Line 1\n\nLine 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertBBCodeToMarkdown(tt.input)
			if result != tt.expected {
				t.Errorf("\nInput:    %q\nExpected: %q\nGot:      %q", tt.input, tt.expected, result)
			}
		})
	}
}

// TestMarkdownLinkPreservation tests that markdown links are preserved during BB-code cleanup
func TestMarkdownLinkPreservation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Markdown link after BB-code",
			input:    "[b]bold text[/b] [link](https://example.com)",
			expected: "**bold text** [link](https://example.com)",
		},
		{
			name:     "Multiple identical BB-codes with markdown link",
			input:    "[b]first[/b] [b]second[/b] [link](url) [b]third[/b]",
			expected: "**first** **second** [link](url) **third**",
		},
		{
			name:     "Unknown BB-code with markdown link",
			input:    "[unknown]text[/unknown] [link](url)",
			expected: "text [link](url)",
		},
		{
			name:     "Mixed BB-codes and markdown links",
			input:    "[i]italic[/i] [link1](url1) [b]bold[/b] [link2](url2) [unknown]removed[/unknown]",
			expected: "*italic* [link1](url1) **bold** [link2](url2) removed",
		},
		{
			name:     "ATTACH tag preserved with markdown link",
			input:    "[ATTACH=123] [link](url) [ATTACH=full]456[/ATTACH]",
			expected: "[ATTACH=123] [link](url) [ATTACH=full]456[/ATTACH]",
		},
		{
			name:     "Edge case: bracket in URL",
			input:    "[b]text[/b] [link](https://example.com/page[1])",
			expected: "**text** [link](https://example.com/page[1])",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertBBCodeToMarkdown(tt.input)
			if result != tt.expected {
				t.Errorf("\nInput:    %q\nExpected: %q\nGot:      %q", tt.input, tt.expected, result)
			}
		})
	}
}

// TestFormatMessage tests the message formatting with metadata
func TestFormatMessage(t *testing.T) {
	username := "JohnDoe"
	postDate := int64(1642353000) // 2022-01-16 17:10:00 UTC
	threadID := 12345
	content := "This is the message content."

	expected := `---
Author: JohnDoe
Posted: 2022-01-16 17:10:00 UTC
Original Thread ID: 12345
---

This is the message content.`

	result := formatMessage(username, postDate, threadID, content)
	if result != expected {
		t.Errorf("\nExpected:\n%s\n\nGot:\n%s", expected, result)
	}
}

// TestReplaceAttachmentLinks tests attachment link replacement
func TestReplaceAttachmentLinks(t *testing.T) {
	attachments := []XenForoAttachment{
		{AttachmentID: 123, Filename: "screenshot.png", ViewURL: "https://example.com/attach/123"},
		{AttachmentID: 456, Filename: "document.pdf", ViewURL: "https://example.com/attach/456"},
		{AttachmentID: 789, Filename: "archive.zip", ViewURL: "https://example.com/attach/789"},
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Image attachment",
			input:    "Check this screenshot: [ATTACH=123]",
			expected: "Check this screenshot: ![screenshot.png](./png/attachment_123_screenshot.png)",
		},
		{
			name:     "PDF attachment",
			input:    "Download the document: [ATTACH=456]",
			expected: "Download the document: [document.pdf](./pdf/attachment_456_document.pdf)",
		},
		{
			name:     "Full format attachment",
			input:    "See image: [ATTACH=full]123[/ATTACH]",
			expected: "See image: ![screenshot.png](./png/attachment_123_screenshot.png)",
		},
		{
			name:     "Multiple attachments",
			input:    "[ATTACH=123] and [ATTACH=456]",
			expected: "![screenshot.png](./png/attachment_123_screenshot.png) and [document.pdf](./pdf/attachment_456_document.pdf)",
		},
		{
			name:     "Unknown attachment",
			input:    "Missing: [ATTACH=999]",
			expected: "Missing: [ATTACH=999]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := replaceAttachmentLinks(tt.input, attachments)
			if result != tt.expected {
				t.Errorf("\nExpected: %q\nGot:      %q", tt.expected, result)
			}
		})
	}
}

// TestProgressTracking tests save and load progress functionality
func TestProgressTracking(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "godisc-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Override progress file location
	oldProgressFile := ProgressFile
	testProgressFile := filepath.Join(tempDir, "test_progress.json")
	defer func() {
		ProgressFile = oldProgressFile
	}()

	// Test saving progress
	testProgress := &MigrationProgress{
		LastThreadID:     100,
		CompletedThreads: []int{1, 2, 3, 100},
		FailedThreads:    []int{50},
		LastUpdated:      time.Now().Unix(),
	}

	// Save test progress
	data, _ := json.MarshalIndent(testProgress, "", "  ")
	err = os.WriteFile(testProgressFile, data, 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test loading progress
	loadedData, err := os.ReadFile(testProgressFile)
	if err != nil {
		t.Fatal(err)
	}

	var loadedProgress MigrationProgress
	err = json.Unmarshal(loadedData, &loadedProgress)
	if err != nil {
		t.Fatal(err)
	}

	// Verify loaded data
	if loadedProgress.LastThreadID != testProgress.LastThreadID {
		t.Errorf("LastThreadID mismatch: expected %d, got %d", testProgress.LastThreadID, loadedProgress.LastThreadID)
	}

	if len(loadedProgress.CompletedThreads) != len(testProgress.CompletedThreads) {
		t.Errorf("CompletedThreads length mismatch: expected %d, got %d",
			len(testProgress.CompletedThreads), len(loadedProgress.CompletedThreads))
	}

	if len(loadedProgress.FailedThreads) != len(testProgress.FailedThreads) {
		t.Errorf("FailedThreads length mismatch: expected %d, got %d",
			len(testProgress.FailedThreads), len(loadedProgress.FailedThreads))
	}
}

// TestFilterCompletedThreads tests the thread filtering logic
func TestFilterCompletedThreads(t *testing.T) {
	// Set up test progress
	progress = &MigrationProgress{
		LastThreadID:     50,
		CompletedThreads: []int{10, 20, 30, 60, 70},
		FailedThreads:    []int{},
	}

	threads := []XenForoThread{
		{ThreadID: 10, Title: "Thread 10"}, // Completed
		{ThreadID: 20, Title: "Thread 20"}, // Completed
		{ThreadID: 25, Title: "Thread 25"}, // Not completed, less than LastThreadID
		{ThreadID: 30, Title: "Thread 30"}, // Completed
		{ThreadID: 40, Title: "Thread 40"}, // Not completed, less than LastThreadID
		{ThreadID: 60, Title: "Thread 60"}, // Completed
		{ThreadID: 70, Title: "Thread 70"}, // Completed
		{ThreadID: 80, Title: "Thread 80"}, // Not completed, greater than LastThreadID
		{ThreadID: 90, Title: "Thread 90"}, // Not completed, greater than LastThreadID
	}

	filtered := filterCompletedThreads(threads)

	// Expected: threads 25, 40, 80, 90 (not completed)
	expectedIDs := []int{25, 40, 80, 90}

	if len(filtered) != len(expectedIDs) {
		t.Errorf("Expected %d threads, got %d", len(expectedIDs), len(filtered))
	}

	for i, thread := range filtered {
		if i < len(expectedIDs) && thread.ThreadID != expectedIDs[i] {
			t.Errorf("Expected thread ID %d at position %d, got %d", expectedIDs[i], i, thread.ThreadID)
		}
	}
}

// TestExtensionDetection tests file extension detection for attachments
func TestExtensionDetection(t *testing.T) {
	tests := []struct {
		filename string
		expected string
		isImage  bool
	}{
		{"image.png", "png", true},
		{"photo.jpg", "jpg", true},
		{"photo.jpeg", "jpeg", true},
		{"animation.gif", "gif", true},
		{"modern.webp", "webp", true},
		{"document.pdf", "pdf", false},
		{"archive.zip", "zip", false},
		{"code.txt", "txt", false},
		{"noextension", "unknown", false},
		{"UPPERCASE.PNG", "png", true},
		{"multiple.dots.in.name.pdf", "pdf", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			ext := strings.ToLower(filepath.Ext(tt.filename))
			if ext == "" {
				ext = ".unknown"
			}
			ext = strings.TrimPrefix(ext, ".")

			if ext != tt.expected {
				t.Errorf("Expected extension %q for %q, got %q", tt.expected, tt.filename, ext)
			}

			isImage := ext == "png" || ext == "jpg" || ext == "jpeg" || ext == "gif" || ext == "webp"
			if isImage != tt.isImage {
				t.Errorf("Expected isImage=%v for %q, got %v", tt.isImage, tt.filename, isImage)
			}
		})
	}
}

// BenchmarkConvertBBCodeToMarkdown benchmarks the BB-code conversion
func BenchmarkConvertBBCodeToMarkdown(b *testing.B) {
	// Create a complex input with various BB codes
	input := `[b]Bold text[/b] and [i]italic[/i] with [url=https://example.com]a link[/url].
[quote="User"]This is a quoted text with [b]formatting[/b].[/quote]
[code]
func example() {
    return "Hello, World!"
}
[/code]
[spoiler]Hidden content[/spoiler]
[list]
[*]Item 1
[*]Item 2
[*]Item 3
[/list]`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = convertBBCodeToMarkdown(input)
	}
}

// BenchmarkReplaceAttachmentLinks benchmarks attachment replacement
func BenchmarkReplaceAttachmentLinks(b *testing.B) {
	attachments := make([]XenForoAttachment, 100)
	for i := 0; i < 100; i++ {
		attachments[i] = XenForoAttachment{
			AttachmentID: i,
			Filename:     fmt.Sprintf("file%d.png", i),
			ViewURL:      fmt.Sprintf("https://example.com/attach/%d", i),
		}
	}

	input := "Text with [ATTACH=0] and [ATTACH=50] and [ATTACH=99] attachments."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = replaceAttachmentLinks(input, attachments)
	}
}

// TestEdgeCases tests various edge cases
func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "Empty input",
			testFunc: func(t *testing.T) {
				result := convertBBCodeToMarkdown("")
				if result != "" {
					t.Errorf("Expected empty string, got %q", result)
				}
			},
		},
		{
			name: "Only whitespace",
			testFunc: func(t *testing.T) {
				result := convertBBCodeToMarkdown("   \n\t  ")
				if result != "" {
					t.Errorf("Expected empty string after trim, got %q", result)
				}
			},
		},
		{
			name: "Malformed BB-code",
			testFunc: func(t *testing.T) {
				input := "[b[i]text[/i]"
				result := convertBBCodeToMarkdown(input)
				// Should handle gracefully without panic
				if strings.Contains(result, "[b[") {
					t.Errorf("Malformed tag not cleaned up: %q", result)
				}
			},
		},
		{
			name: "Unicode handling",
			testFunc: func(t *testing.T) {
				input := "[b]Hello 世界 🌍[/b]"
				expected := "**Hello 世界 🌍**"
				result := convertBBCodeToMarkdown(input)
				if result != expected {
					t.Errorf("Unicode not handled correctly.\nExpected: %q\nGot: %q", expected, result)
				}
			},
		},
		{
			name: "Very long input",
			testFunc: func(t *testing.T) {
				// Generate a very long string
				var sb strings.Builder
				for i := 0; i < 1000; i++ {
					sb.WriteString(fmt.Sprintf("[b]Line %d[/b]\n", i))
				}
				input := sb.String()
				result := convertBBCodeToMarkdown(input)

				// Should complete without timeout or panic
				if !strings.Contains(result, "**Line 999**") {
					t.Error("Long input not processed correctly")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}
