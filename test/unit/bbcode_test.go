package unit

import (
	"strings"
	"testing"
	"time"

	"github.com/exileum/xenforo-to-gh-discussions/internal/bbcode"
)

func TestBBCodeConverter(t *testing.T) {
	converter := bbcode.NewConverter()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Bold text",
			input:    "[b]bold text[/b]",
			expected: "**bold text**",
		},
		{
			name:     "Italic text",
			input:    "[i]italic text[/i]",
			expected: "*italic text*",
		},
		{
			name:     "Empty tags are removed",
			input:    "[b][/b]",
			expected: "",
		},
		{
			name:     "URLs with description",
			input:    "[url=\"https://example.com\"]Example[/url]",
			expected: "[Example](https://example.com)",
		},
		{
			name:     "Simple quotes",
			input:    "[quote]This is a quote[/quote]",
			expected: "> This is a quote\n",
		},
		{
			name:     "Code blocks",
			input:    "[code]console.log('hello')[/code]",
			expected: "\n```\nconsole.log('hello')\n```\n",
		},
		{
			name:     "Complex BB-code",
			input:    "This is [b]bold[/b] and [i]italic[/i] text with [url=http://example.com]a link[/url].",
			expected: "This is **bold** and *italic* text with [a link](http://example.com).",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.ToMarkdown(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestMessageProcessor(t *testing.T) {
	processor := bbcode.NewMessageProcessor()

	content := "[b]Test message[/b]"
	result := processor.ProcessContent(content)
	expected := "**Test message**"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestFormatMessage(t *testing.T) {
	processor := bbcode.NewMessageProcessor()

	username := "testuser"
	postDate := int64(1642353000)
	threadID := 123
	content := "Test content"

	result := processor.FormatMessage(username, postDate, threadID, content)

	if !strings.Contains(result, "Author: testuser") {
		t.Error("Message should contain author")
	}
	if !strings.Contains(result, "Original Thread ID: 123") {
		t.Error("Message should contain thread ID")
	}
	if !strings.Contains(result, "Test content") {
		t.Error("Message should contain content")
	}

	expectedTime := time.Unix(postDate, 0).UTC().Format("2006-01-02 15:04:05 UTC")
	if !strings.Contains(result, expectedTime) {
		t.Error("Message should contain formatted timestamp")
	}
}
