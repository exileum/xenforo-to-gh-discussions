package bbcode

import (
	"fmt"
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

func (p *MessageProcessor) FormatMessage(username string, postDate int64, threadID int, content string) string {
	timestamp := time.Unix(postDate, 0).UTC().Format("2006-01-02 15:04:05 UTC")
	return fmt.Sprintf(`---
Author: %s
Posted: %s
Original Thread ID: %d
---

%s`, username, timestamp, threadID, content)
}

func (p *MessageProcessor) ProcessContent(content string) string {
	return p.converter.ToMarkdown(content)
}
