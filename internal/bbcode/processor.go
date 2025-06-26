package bbcode

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type MessageProcessor struct {
	converter *Converter
}

func NewMessageProcessor() *MessageProcessor {
	return &MessageProcessor{
		converter: NewConverter(),
	}
}

func (p *MessageProcessor) FormatMessage(username string, postDate int64, threadID int, content string) (string, error) {
	// Validate username
	if strings.TrimSpace(username) == "" {
		return "", errors.New("username cannot be empty")
	}

	// Validate threadID
	if threadID <= 0 {
		return "", errors.New("threadID must be positive")
	}

	// Validate content
	if strings.TrimSpace(content) == "" {
		return "", errors.New("content cannot be empty")
	}

	// Validate and convert timestamp
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
		// Check if the converted time is reasonable (not too far in past/future)
		now := time.Now().UTC()
		minDate := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
		maxDate := now.AddDate(10, 0, 0) // Allow up to 10 years in the future

		if t.Before(minDate) || t.After(maxDate) {
			timestamp = fmt.Sprintf("Invalid Date (timestamp: %d)", postDate)
		} else {
			timestamp = t.Format("2006-01-02 15:04:05 UTC")
		}
	}()

	// If timestamp conversion failed, return error
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

func (p *MessageProcessor) ProcessContent(content string) string {
	result := p.converter.ToMarkdown(content)

	// Convert @username mentions to **username** to avoid GitHub user mentions
	result = p.convertAtMentions(result)

	return result
}

// convertAtMentions converts @username patterns to **username** bold format
func (p *MessageProcessor) convertAtMentions(content string) string {
	// Enhanced regex: require at least one alphabetic character and use word boundaries
	// This prevents purely numeric usernames and ensures proper boundaries
	mentionRe := regexp.MustCompile(`@([a-zA-Z0-9_-]*[a-zA-Z]+[a-zA-Z0-9_-]*)\b`)

	// More comprehensive email pattern to avoid false positives
	emailRe := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)

	// Find all email matches first to exclude them
	emailMatches := emailRe.FindAllStringIndex(content, -1)

	// Use ReplaceAllStringFunc with position tracking for efficiency
	result := mentionRe.ReplaceAllStringFunc(content, func(match string) string {
		// Get the match position using FindStringIndex
		matchIndices := mentionRe.FindAllStringIndex(content, -1)
		var matchStart, matchEnd int

		// Find the current match position by comparing the match string
		for _, indices := range matchIndices {
			if content[indices[0]:indices[1]] == match {
				matchStart, matchEnd = indices[0], indices[1]
				break
			}
		}

		// Check if this @ symbol is part of an email by checking overlap with email matches
		for _, emailIndex := range emailMatches {
			emailStart, emailEnd := emailIndex[0], emailIndex[1]
			// If the @ symbol overlaps with an email, skip conversion
			if matchStart >= emailStart && matchEnd <= emailEnd {
				return match
			}
		}

		// Extract username from the match
		parts := mentionRe.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		username := parts[1]

		return "**" + username + "**"
	})

	return result
}
