package bbcode

import (
	"regexp"
	"strings"

	"github.com/dlclark/regexp2"
)

type Converter struct{}

func NewConverter() *Converter {
	return &Converter{}
}

func (c *Converter) ToMarkdown(bbcode string) string {
	if strings.TrimSpace(bbcode) == "" {
		return ""
	}

	result := bbcode

	// First, handle multi-line code blocks
	result = c.processCodeBlocks(result)

	// Handle quotes with attribution
	result = c.processQuotes(result)

	// URLs with quotes first
	result = regexp.MustCompile(`\[url="([^"]+)"\](.*?)\[/url\]`).ReplaceAllString(result, "[$2]($1)")

	// Handle text formatting with empty tag removal
	result = c.processFormattingTag(result, `\[b\](.*?)\[/b\]`, "**", "**")
	result = c.processFormattingTag(result, `\[i\](.*?)\[/i\]`, "*", "*")
	result = c.processFormattingTag(result, `\[u\](.*?)\[/u\]`, "<u>", "</u>")
	result = c.processFormattingTag(result, `\[s\](.*?)\[/s\]`, "~~", "~~")
	result = c.processFormattingTag(result, `\[strike\](.*?)\[/strike\]`, "~~", "~~")

	// Apply simple replacements
	result = c.applySimpleReplacements(result)

	// Clean up unhandled BB codes
	result = c.cleanupUnhandledTags(result)

	// Final cleanup
	result = c.finalCleanup(result)

	return result
}

func (c *Converter) processCodeBlocks(input string) string {
	return regexp.MustCompile(`(?s)\[code\](.*?)\[/code\]`).ReplaceAllStringFunc(input, func(match string) string {
		parts := regexp.MustCompile(`(?s)\[code\](.*?)\[/code\]`).FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		content := parts[1]
		return "\n```\n" + strings.TrimSpace(content) + "\n```\n"
	})
}

func (c *Converter) processQuotes(input string) string {
	// Process quotes iteratively to handle nested quotes
	result := input
	maxIterations := 10 // Prevent infinite loops

	for i := 0; i < maxIterations; i++ {
		oldResult := result

		// Handle quotes with attribution first
		result = regexp.MustCompile(`(?s)\[quote="([^,"]+)(?:,[^\]]+)?"\](.*?)\[/quote\]`).ReplaceAllStringFunc(result, func(match string) string {
			parts := regexp.MustCompile(`(?s)\[quote="([^,"]+)(?:,[^\]]+)?"\](.*?)\[/quote\]`).FindStringSubmatch(match)
			if len(parts) < 3 {
				return match
			}
			author := parts[1]
			content := parts[2]
			lines := strings.Split(strings.TrimSpace(content), "\n")
			quoted := "> **" + author + " said:**\n"
			for _, line := range lines {
				quoted += "> " + line + "\n"
			}
			return quoted
		})

		// Handle simple quotes
		result = regexp.MustCompile(`(?s)\[quote\](.*?)\[/quote\]`).ReplaceAllStringFunc(result, func(match string) string {
			parts := regexp.MustCompile(`(?s)\[quote\](.*?)\[/quote\]`).FindStringSubmatch(match)
			if len(parts) < 2 {
				return match
			}
			content := parts[1]
			lines := strings.Split(strings.TrimSpace(content), "\n")
			quoted := ""
			for _, line := range lines {
				quoted += "> " + line + "\n"
			}
			return quoted
		})

		// If no changes were made, we're done
		if result == oldResult {
			break
		}
	}

	return result
}

func (c *Converter) processFormattingTag(input, pattern, openTag, closeTag string) string {
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllStringFunc(input, func(match string) string {
		submatch := re.FindStringSubmatch(match)
		if len(submatch) < 2 {
			return match
		}
		content := submatch[1]
		if strings.TrimSpace(content) == "" {
			return ""
		}
		return openTag + content + closeTag
	})
}

func (c *Converter) applySimpleReplacements(input string) string {
	replacements := []struct {
		pattern     *regexp.Regexp
		replacement string
	}{
		// URLs (without quotes)
		{regexp.MustCompile(`\[url=([^\]]+)\](.*?)\[/url\]`), "[$2]($1)"},
		{regexp.MustCompile(`\[url\](.*?)\[/url\]`), "[$1]($1)"},

		// Images
		{regexp.MustCompile(`\[img\](.*?)\[/img\]`), "![]($1)"},

		// Spoilers
		{regexp.MustCompile(`(?s)\[spoiler(?:="[^"]*")?\](.*?)\[/spoiler\]`), "<details><summary>Spoiler</summary>\n\n$1\n\n</details>"},
		{regexp.MustCompile(`\[ispoiler\](.*?)\[/ispoiler\]`), "||$1||"},

		// Media embeds
		{regexp.MustCompile(`\[media=([^\]]+)\](.*?)\[/media\]`), "[$1]($2)"},

		// Lists
		{regexp.MustCompile(`\[\*\]`), "- "},
		{regexp.MustCompile(`\[list=1\]\n`), "\n"},
		{regexp.MustCompile(`\[list\]\n`), "\n"},
		{regexp.MustCompile(`\n\[/list\]`), "\n"},

		// Center alignment
		{regexp.MustCompile(`\[center\](.*?)\[/center\]`), "<center>$1</center>"},

		// Remove color, size, font tags
		{regexp.MustCompile(`\[color=[^\]]+\](.*?)\[/color\]`), "$1"},
		{regexp.MustCompile(`\[size=[^\]]+\](.*?)\[/size\]`), "$1"},
		{regexp.MustCompile(`\[font=[^\]]+\](.*?)\[/font\]`), "$1"},
	}

	result := input
	for _, r := range replacements {
		result = r.pattern.ReplaceAllString(result, r.replacement)
	}

	return result
}

func (c *Converter) cleanupUnhandledTags(input string) string {
	cleanupPattern := regexp2.MustCompile(`\[/?[a-zA-Z][a-zA-Z0-9=_-]*\](?!\()`, 0)
	result, _ := cleanupPattern.ReplaceFunc(input, func(m regexp2.Match) string {
		match := m.String()
		// Preserve ATTACH tags for later processing
		if strings.HasPrefix(match, "[ATTACH") || match == "[/ATTACH]" {
			return match
		}
		return ""
	}, -1, -1)

	return result
}

func (c *Converter) finalCleanup(input string) string {
	result := regexp.MustCompile(`\n{3,}`).ReplaceAllString(input, "\n\n")
	return strings.Trim(result, " \t")
}
