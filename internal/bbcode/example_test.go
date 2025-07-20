package bbcode

import (
	"fmt"
	"time"
)

func ExampleConverter_ToMarkdown() {
	converter := NewConverter()
	bbcode := "[b]Bold text[/b] with [url=https://example.com]a link[/url]"
	markdown := converter.ToMarkdown(bbcode)
	fmt.Println(markdown)
	// Output: **Bold text** with [a link](https://example.com)
}

func ExampleConverter_ToMarkdown_quotes() {
	converter := NewConverter()
	bbcode := `[quote="username"]This is a quoted message[/quote]`
	markdown := converter.ToMarkdown(bbcode)
	fmt.Println(markdown)
	// Output: > **username said:**
	// > This is a quoted message
}

func ExampleConverter_ToMarkdown_formatting() {
	converter := NewConverter()
	bbcode := "[b]Bold[/b], [i]italic[/i], [u]underlined[/u], and [s]strikethrough[/s]"
	markdown := converter.ToMarkdown(bbcode)
	fmt.Println(markdown)
	// Output: **Bold**, *italic*, <u>underlined</u>, and ~~strikethrough~~
}

func ExampleConverter_ToMarkdown_links() {
	converter := NewConverter()
	bbcode := "[url=https://github.com]GitHub[/url] and [url]https://example.com[/url]"
	markdown := converter.ToMarkdown(bbcode)
	fmt.Println(markdown)
	// Output: [GitHub](https://github.com) and [https://example.com](https://example.com)
}

func ExampleConverter_ToMarkdown_images() {
	converter := NewConverter()
	bbcode := "[img]https://example.com/image.jpg[/img]"
	markdown := converter.ToMarkdown(bbcode)
	fmt.Println(markdown)
	// Output: ![](https://example.com/image.jpg)
}

func ExampleConverter_ToMarkdown_code() {
	converter := NewConverter()
	bbcode := "[code]function hello() { console.log('Hello'); }[/code]"
	markdown := converter.ToMarkdown(bbcode)
	fmt.Println(markdown)
	// Output: ```
	// function hello() { console.log('Hello'); }
	// ```
}

func ExampleConverter_ToMarkdown_spoiler() {
	converter := NewConverter()
	bbcode := "[spoiler=\"Click to reveal\"]Hidden content here[/spoiler]"
	markdown := converter.ToMarkdown(bbcode)
	fmt.Println(markdown)
	// Output: <details><summary>Spoiler</summary>
	//
	// Hidden content here
	//
	// </details>
}

func ExampleNewConverter() {
	// Create a new BBCode to Markdown converter
	converter := NewConverter()

	// Convert some BBCode
	result := converter.ToMarkdown("[b]Hello World![/b]")
	fmt.Println(result)
	// Output: **Hello World!**
}

func ExampleMessageProcessor_FormatMessage() {
	processor := NewMessageProcessor()

	username := "john_doe"
	postDate := time.Date(2023, 1, 15, 14, 30, 0, 0, time.UTC).Unix()
	threadID := 12345
	content := "This is a [b]forum post[/b] with BBCode formatting."

	formatted, err := processor.FormatMessage(username, postDate, threadID, content)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(formatted)
	// Output: ---
	// Author: **john_doe**
	// Posted: 2023-01-15 14:30:00 UTC
	// Original Thread ID: 12345
	// ---
	//
	// This is a [b]forum post[/b] with BBCode formatting.
}

func ExampleMessageProcessor_ProcessContent() {
	processor := NewMessageProcessor()

	content := "Hello @username, check this [url=https://example.com]link[/url]!"
	processed := processor.ProcessContent(content)

	fmt.Println(processed)
	// Output: Hello **username**, check this [link](https://example.com)!
}

func ExampleNewMessageProcessor() {
	// Create a new message processor
	processor := NewMessageProcessor()

	// Process BBCode content with @mentions
	content := "Hey @alice, this [b]works great[/b]!"
	result := processor.ProcessContent(content)

	fmt.Println(result)
	// Output: Hey **alice**, this **works great**!
}
