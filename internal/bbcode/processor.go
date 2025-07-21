package bbcode

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// MessageProcessor handles complete message formatting for forum migration.
// Combines BB-code conversion with metadata formatting including author,
// timestamps, and thread information.
type MessageProcessor struct {
	converter *Converter
}

// NewMessageProcessor creates a new message processor with an integrated
// BB-code converter for complete forum post processing.
func NewMessageProcessor() *MessageProcessor {
	return &MessageProcessor{
		converter: NewConverter(),
	}
}

// FormatMessage formats a complete forum post with metadata and content conversion.
// Combines author information, timestamps, thread ID, and BB-code converted content
// into a formatted GitHub Discussion post with YAML frontmatter.
//
// Returns an error if any required parameters are invalid or timestamp conversion fails.
func (p *MessageProcessor) FormatMessage(username string, postDate int64, threadID int, content string) (string, error) {
	if strings.TrimSpace(username) == "" {
		return "", errors.New("username cannot be empty")
	}

	if threadID <= 0 {
		return "", errors.New("threadID must be positive")
	}

	if strings.TrimSpace(content) == "" {
		return "", errors.New("content cannot be empty")
	}

	if postDate < 0 {
		return "", errors.New("postDate cannot be negative")
	}

	// Handle potential time conversion issues
	var timestamp string
	func() {
		defer func() {
			if r := recover(); r != nil {
				timestamp = "Invalid Date"
			}
		}()

		t := time.Unix(postDate, 0).UTC()
		now := time.Now().UTC()
		minDate := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
		maxDate := now.AddDate(10, 0, 0)

		if t.Before(minDate) || t.After(maxDate) {
			timestamp = fmt.Sprintf("Invalid Date (timestamp: %d)", postDate)
		} else {
			timestamp = t.Format("2006-01-02 15:04:05 UTC")
		}
	}()

	if strings.Contains(timestamp, "Invalid Date") {
		return "", fmt.Errorf("invalid timestamp: %d", postDate)
	}

	formatted := fmt.Sprintf(`---
Author: **%s**
Posted: %s
Original Thread ID: %d
---

%s`, strings.TrimSpace(username), timestamp, threadID, strings.TrimSpace(content))

	return formatted, nil
}

func (p *MessageProcessor) ProcessContent(ctx context.Context, content string) (string, error) {
	result, err := p.converter.ToMarkdown(ctx, content)
	if err != nil {
		return "", err
	}

	result = p.convertAtMentions(result)

	return result, nil
}

// convertAtMentions converts @username patterns to **username** bold format
func (p *MessageProcessor) convertAtMentions(content string) string {
	mentionRe := regexp.MustCompile(`@([a-zA-Z0-9_-]*[a-zA-Z]+[a-zA-Z0-9_-]*)\b`)

	emailRe := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)

	emailMatches := emailRe.FindAllStringIndex(content, -1)

	mentionMatches := mentionRe.FindAllStringIndex(content, -1)
	if len(mentionMatches) == 0 {
		return content
	}

	result := content
	offset := 0

	for _, matchIndices := range mentionMatches {
		matchStart, matchEnd := matchIndices[0], matchIndices[1]
		match := content[matchStart:matchEnd]

		isInEmail := false
		for _, emailIndex := range emailMatches {
			emailStart, emailEnd := emailIndex[0], emailIndex[1]
			if matchStart >= emailStart && matchEnd <= emailEnd {
				isInEmail = true
				break
			}
		}

		if isInEmail {
			continue
		}

		parts := mentionRe.FindStringSubmatch(match)
		if len(parts) < 2 {
			continue
		}
		username := parts[1]
		replacement := "**" + username + "**"

		adjustedStart := matchStart + offset
		adjustedEnd := matchEnd + offset
		result = result[:adjustedStart] + replacement + result[adjustedEnd:]
		offset += len(replacement) - (matchEnd - matchStart)
	}

	return result
}
