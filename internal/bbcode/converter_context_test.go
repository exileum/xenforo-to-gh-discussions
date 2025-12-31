package bbcode

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestToMarkdown_ContextCancellation(t *testing.T) {
	converter := NewConverter()

	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		bbcode := "[b]Bold text[/b] with [url=https://example.com]a link[/url]"
		_, err := converter.ToMarkdown(ctx, bbcode)

		if err == nil {
			t.Error("Expected error from cancelled context")
		}

		if !strings.Contains(err.Error(), "BBCode conversion cancelled") {
			t.Errorf("Expected 'BBCode conversion cancelled' in error, got: %v", err)
		}
	})

	t.Run("timeout context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		// Use a large BBCode string to ensure processing takes some time
		bbcode := strings.Repeat("[b]Bold[/b] [i]Italic[/i] ", 1000)
		_, err := converter.ToMarkdown(ctx, bbcode)

		if err == nil {
			t.Error("Expected timeout error")
		}

		if !strings.Contains(err.Error(), "cancelled") {
			t.Errorf("Expected cancellation error, got: %v", err)
		}
	})

	t.Run("successful conversion with active context", func(t *testing.T) {
		ctx := context.Background()
		bbcode := "[b]Bold text[/b]"

		result, err := converter.ToMarkdown(ctx, bbcode)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		expected := "**Bold text**"
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})

	t.Run("empty input with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Empty input should return immediately without checking context
		result, err := converter.ToMarkdown(ctx, "")
		if err != nil {
			t.Errorf("Empty input should not check context, got error: %v", err)
		}

		if result != "" {
			t.Errorf("Expected empty string, got %q", result)
		}
	})
}

func TestProcessContent_ContextCancellation(t *testing.T) {
	processor := NewMessageProcessor()

	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		content := "[b]Some content[/b] with @mention"
		_, err := processor.ProcessContent(ctx, content)

		if err == nil {
			t.Error("Expected error from cancelled context")
		}

		if !strings.Contains(err.Error(), "cancelled") {
			t.Errorf("Expected cancellation error, got: %v", err)
		}
	})

	t.Run("successful processing", func(t *testing.T) {
		ctx := context.Background()
		content := "@testuser mentioned here"

		result, err := processor.ProcessContent(ctx, content)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if !strings.Contains(result, "**testuser**") {
			t.Errorf("Expected mention to be bolded, got: %s", result)
		}
	})
}
