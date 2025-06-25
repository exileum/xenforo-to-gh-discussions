package bbcode

import (
	"errors"
	"fmt"
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
Author: %s
Posted: %s
Original Thread ID: %d
---

%s`, strings.TrimSpace(username), timestamp, threadID, strings.TrimSpace(content))

	return formatted, nil
}

func (p *MessageProcessor) ProcessContent(content string) string {
	return p.converter.ToMarkdown(content)
}
