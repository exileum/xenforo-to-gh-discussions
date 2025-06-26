package bbcode

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestBBCodeConverter(t *testing.T) {
	converter := NewConverter()

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
			input:    "This is [b]bold[/b] and [i]italic[/i] text with [url=https://example.com]a link[/url].",
			expected: "This is **bold** and *italic* text with [a link](https://example.com).",
		},
		{
			name:     "Quotes with attribution",
			input:    "[quote=\"John\"]This is a quoted message[/quote]",
			expected: "> **John said:**\n> This is a quoted message\n",
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
	processor := NewMessageProcessor()

	content := "[b]Test message[/b]"
	result := processor.ProcessContent(content)
	expected := "**Test message**"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestAtMentionConversion(t *testing.T) {
	processor := NewMessageProcessor()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple @ mention",
			input:    "Hello @john how are you?",
			expected: "Hello **john** how are you?",
		},
		{
			name:     "Multiple @ mentions",
			input:    "@alice and @bob are here",
			expected: "**alice** and **bob** are here",
		},
		{
			name:     "@ mention with underscore",
			input:    "Hey @user_name",
			expected: "Hey **user_name**",
		},
		{
			name:     "@ mention with hyphen",
			input:    "Hi @user-name",
			expected: "Hi **user-name**",
		},
		{
			name:     "Email should not be converted",
			input:    "Contact user@example.com for help",
			expected: "Contact user@example.com for help",
		},
		{
			name:     "Purely numeric username should not be converted",
			input:    "Check thread @123 for details",
			expected: "Check thread @123 for details",
		},
		{
			name:     "Complex email patterns should not be converted",
			input:    "Reach out to test.user+tag@sub.example.co.uk",
			expected: "Reach out to test.user+tag@sub.example.co.uk",
		},
		{
			name:     "Mixed content",
			input:    "Thanks @admin for [b]fixing[/b] the issue!",
			expected: "Thanks **admin** for **fixing** the issue!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.ProcessContent(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFormatMessage(t *testing.T) {
	processor := NewMessageProcessor()

	tests := []struct {
		name      string
		username  string
		postDate  int64
		threadID  int
		content   string
		shouldErr bool
		errMsg    string
	}{
		{
			name:      "Valid input",
			username:  "testuser",
			postDate:  1642353000, // Valid Unix timestamp
			threadID:  123,
			content:   "Test content",
			shouldErr: false,
		},
		{
			name:      "Empty username",
			username:  "",
			postDate:  1642353000,
			threadID:  123,
			content:   "Test content",
			shouldErr: true,
			errMsg:    "username cannot be empty",
		},
		{
			name:      "Whitespace only username",
			username:  "   \t\n   ",
			postDate:  1642353000,
			threadID:  123,
			content:   "Test content",
			shouldErr: true,
			errMsg:    "username cannot be empty",
		},
		{
			name:      "Negative threadID",
			username:  "testuser",
			postDate:  1642353000,
			threadID:  -1,
			content:   "Test content",
			shouldErr: true,
			errMsg:    "threadID must be positive",
		},
		{
			name:      "Zero threadID",
			username:  "testuser",
			postDate:  1642353000,
			threadID:  0,
			content:   "Test content",
			shouldErr: true,
			errMsg:    "threadID must be positive",
		},
		{
			name:      "Empty content",
			username:  "testuser",
			postDate:  1642353000,
			threadID:  123,
			content:   "",
			shouldErr: true,
			errMsg:    "content cannot be empty",
		},
		{
			name:      "Whitespace only content",
			username:  "testuser",
			postDate:  1642353000,
			threadID:  123,
			content:   "   \t\n   ",
			shouldErr: true,
			errMsg:    "content cannot be empty",
		},
		{
			name:      "Negative postDate",
			username:  "testuser",
			postDate:  -1,
			threadID:  123,
			content:   "Test content",
			shouldErr: true,
			errMsg:    "postDate cannot be negative",
		},
		{
			name:      "Very old timestamp (before 1970)",
			username:  "testuser",
			postDate:  -86400, // One day before Unix epoch
			threadID:  123,
			content:   "Test content",
			shouldErr: true,
			errMsg:    "postDate cannot be negative",
		},
		{
			name:      "Future timestamp (but reasonable)",
			username:  "testuser",
			postDate:  time.Now().Unix() + 86400, // Tomorrow
			threadID:  123,
			content:   "Test content",
			shouldErr: false,
		},
		{
			name:      "Far future timestamp (unreasonable)",
			username:  "testuser",
			postDate:  time.Now().AddDate(20, 0, 0).Unix(), // 20 years in future
			threadID:  123,
			content:   "Test content",
			shouldErr: true,
			errMsg:    "invalid timestamp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processor.FormatMessage(tt.username, tt.postDate, tt.threadID, tt.content)

			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error message to contain %q, got %q", tt.errMsg, err.Error())
				}
				if result != "" {
					t.Error("Expected empty result when error occurs")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
					return
				}
				if result == "" {
					t.Error("Expected non-empty result")
					return
				}

				// Verify content structure for valid cases
				if !strings.Contains(result, fmt.Sprintf("Author: **%s**", strings.TrimSpace(tt.username))) {
					t.Error("Message should contain author")
				}
				if !strings.Contains(result, fmt.Sprintf("Original Thread ID: %d", tt.threadID)) {
					t.Error("Message should contain thread ID")
				}
				if !strings.Contains(result, strings.TrimSpace(tt.content)) {
					t.Error("Message should contain content")
				}

				expectedTime := time.Unix(tt.postDate, 0).UTC().Format("2006-01-02 15:04:05 UTC")
				if !strings.Contains(result, expectedTime) {
					t.Error("Message should contain formatted timestamp")
				}
			}
		})
	}
}
