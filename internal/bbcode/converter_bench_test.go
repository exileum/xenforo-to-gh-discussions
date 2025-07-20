package bbcode

import (
	"strings"
	"testing"
)

// Sample BBCode content for benchmarking
var sampleBBCode = `[quote="username"]This is a quoted message with [b]bold[/b] text inside[/quote]

Here is some [b]bold text[/b] and [i]italic text[/i] and [u]underlined[/u].

[url=https://example.com]This is a link[/url] and [url]https://direct-link.com[/url].

[img]https://example.com/image.jpg[/img]

[spoiler="Hidden Content"]
This is hidden content that can be revealed.
[b]Bold text inside spoiler[/b]
[/spoiler]

[code]
function test() {
    console.log("Hello, World!");
    return true;
}
[/code]

[list]
[*]First item
[*]Second item with [b]bold[/b] text
[*]Third item
[/list]

[media=youtube]dQw4w9WgXcQ[/media]

Multiple nested quotes:
[quote="user1"]
Original message
[quote="user2"]
Nested quote with [url=https://test.com]link[/url]
[/quote]
Response to nested quote
[/quote]`

func BenchmarkConverter_ToMarkdown(b *testing.B) {
	converter := NewConverter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = converter.ToMarkdown(sampleBBCode)
	}
}

func BenchmarkConverter_ToMarkdown_Small(b *testing.B) {
	converter := NewConverter()
	smallContent := "[b]Bold text[/b] with [url=https://example.com]a link[/url]"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = converter.ToMarkdown(smallContent)
	}
}

func BenchmarkConverter_ToMarkdown_Large(b *testing.B) {
	converter := NewConverter()
	// Create large content by repeating the sample
	largeContent := strings.Repeat(sampleBBCode+"\n\n", 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = converter.ToMarkdown(largeContent)
	}
}

func BenchmarkConverter_ToMarkdown_QuotesOnly(b *testing.B) {
	converter := NewConverter()
	quotesContent := `[quote="user1"]Simple quote[/quote]
[quote="user2"]Another quote with [b]formatting[/b][/quote]
[quote="user3"]
Nested content:
[quote="user4"]Inner quote[/quote]
Response to inner quote
[/quote]`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = converter.ToMarkdown(quotesContent)
	}
}

func BenchmarkConverter_ToMarkdown_FormattingOnly(b *testing.B) {
	converter := NewConverter()
	formattingContent := `[b]Bold[/b] [i]Italic[/i] [u]Underline[/u] [s]Strike[/s]
[color=red]Red text[/color] [size=4]Large text[/size]
[center]Centered text[/center] [right]Right aligned[/right]`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = converter.ToMarkdown(formattingContent)
	}
}

func BenchmarkConverter_ToMarkdown_LinksOnly(b *testing.B) {
	converter := NewConverter()
	linksContent := `[url=https://example1.com]Link 1[/url]
[url=https://example2.com]Link 2[/url]
[url]https://direct-link.com[/url]
[email]test@example.com[/email]
[img]https://example.com/image1.jpg[/img]
[img]https://example.com/image2.png[/img]`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = converter.ToMarkdown(linksContent)
	}
}

func BenchmarkConverter_ToMarkdown_CodeBlocks(b *testing.B) {
	converter := NewConverter()
	codeContent := `[code]
function example() {
    const message = "Hello, World!";
    console.log(message);
    return message.length;
}
[/code]

[php]
<?php
function test() {
    echo "PHP code example";
    return true;
}
?>
[/php]

[code=javascript]
const arr = [1, 2, 3, 4, 5];
const doubled = arr.map(x => x * 2);
console.log(doubled);
[/code]`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = converter.ToMarkdown(codeContent)
	}
}

func BenchmarkConverter_ToMarkdown_ComplexNesting(b *testing.B) {
	converter := NewConverter()
	complexContent := `[quote="author"]
This is a complex example with [b]bold text[/b] and [url=https://example.com]links[/url].

[spoiler="Hidden section"]
Inside spoiler: [i]italic[/i] and [code]inline code[/code]
[list]
[*]Spoiler item 1
[*]Spoiler item 2 with [color=blue]colored text[/color]
[/list]
[/spoiler]

[quote="nested_author"]
Nested quote with [img]https://example.com/nested.jpg[/img]
[/quote]

Final response with [media=youtube]dQw4w9WgXcQ[/media]
[/quote]`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = converter.ToMarkdown(complexContent)
	}
}

func BenchmarkMessageProcessor_FormatMessage(b *testing.B) {
	processor := NewMessageProcessor()
	username := "testuser"
	postDate := int64(1640995200) // 2022-01-01 00:00:00 UTC
	threadID := 12345
	content := sampleBBCode

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = processor.FormatMessage(username, postDate, threadID, content)
	}
}

func BenchmarkMessageProcessor_ProcessContent(b *testing.B) {
	processor := NewMessageProcessor()
	content := sampleBBCode

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = processor.ProcessContent(content)
	}
}

func BenchmarkMessageProcessor_convertAtMentions(b *testing.B) {
	processor := NewMessageProcessor()
	content := `Hello @username1, please check @username2's response. 
Contact admin@example.com for support, but @moderator can help too.
@user_with_underscores and @user-with-dashes should be mentioned.
Email notifications go to user@domain.com and test@subdomain.example.org.`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = processor.convertAtMentions(content)
	}
}
