package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/dlclark/regexp2"
	"github.com/go-resty/resty/v2"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// Configuration
var (
	XenForoAPIURL  = "https://your-forum.com/api" // XenForo API URL
	XenForoAPIKey  = "your_xenforo_api_key"       // XenForo super user API key
	XenForoAPIUser = "1"                          // Super user ID (e.g., 1 for admin)
	GitHubToken    = "your_github_token"          // GitHub personal access token
	GitHubRepo     = "your_username/your_repo"    // Repository for Discussions (owner/repo)
	AttachmentsDir = "./attachments"              // Local directory for downloaded attachments
	TargetNodeID   = 1                            // XenForo forum ID to migrate
	MaxRetries     = 3                            // Maximum retry attempts for failed requests
	ProgressFile   = "migration_progress.json"    // File to track migration progress
)

// NodeToCategory maps XenForo node_id to GitHub discussion category_id
var NodeToCategory = map[int]string{
	1: "DIC_kwDOxxxxxxxx", // Replace with actual category ID from GitHub
	// Add more mappings as needed
}

// XenForoThread XenForo API structures
type XenForoThread struct {
	ThreadID    int    `json:"thread_id"`
	Title       string `json:"title"`
	NodeID      int    `json:"node_id"`
	Username    string `json:"username"`
	PostDate    int64  `json:"post_date"`
	FirstPostID int    `json:"first_post_id"`
}

type XenForoPost struct {
	PostID   int    `json:"post_id"`
	ThreadID int    `json:"thread_id"`
	Username string `json:"username"`
	PostDate int64  `json:"post_date"`
	Message  string `json:"message"`
}

type XenForoAttachment struct {
	AttachmentID int    `json:"attachment_id"`
	Filename     string `json:"filename"`
	ViewURL      string `json:"view_url"`
}

// Progress tracking
type MigrationProgress struct {
	LastThreadID     int   `json:"last_thread_id"`
	CompletedThreads []int `json:"completed_threads"`
	FailedThreads    []int `json:"failed_threads"`
	LastUpdated      int64 `json:"last_updated"`
}

// GitHub GraphQL mutations
const createDiscussionMutation = `
mutation($input: CreateDiscussionInput!) {
	createDiscussion(input: $input) {
		discussion {
			id
			number
		}
	}
}
`

const addDiscussionCommentMutation = `
mutation($input: AddDiscussionCommentInput!) {
	addDiscussionComment(input: $input) {
		comment {
			id
		}
	}
}
`

var (
	dryRun       bool
	resumeFrom   int
	verbose      bool
	client       = resty.New()
	githubClient *githubv4.Client
	repositoryID string
	progress     *MigrationProgress
)

func init() {
	flag.BoolVar(&dryRun, "dry-run", false, "Run in dry-run mode (no actual API calls)")
	flag.IntVar(&resumeFrom, "resume-from", 0, "Resume from specific thread ID")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
}

func main() {
	flag.Parse()

	// Load or initialize progress
	progress = loadProgress()
	if resumeFrom > 0 {
		progress.LastThreadID = resumeFrom
	}

	// Initialize GitHub GraphQL client
	if !dryRun {
		src := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: GitHubToken},
		)
		httpClient := oauth2.NewClient(context.Background(), src)
		githubClient = githubv4.NewClient(httpClient)
	}

	// Run pre-flight checks
	log.Println("Running pre-flight checks...")
	if err := runPreflightChecks(); err != nil {
		log.Fatalf("Pre-flight checks failed: %v", err)
	}
	log.Println("✓ All pre-flight checks passed")

	// Get threads from XenForo
	log.Printf("Fetching threads from forum node %d...", TargetNodeID)
	threads, err := getXenForoThreads(TargetNodeID)
	if err != nil {
		log.Fatalf("Failed to fetch threads: %v", err)
	}
	log.Printf("✓ Found %d threads to migrate", len(threads))

	// Filter out already completed threads
	threads = filterCompletedThreads(threads)
	log.Printf("✓ %d threads remaining after filtering completed ones", len(threads))

	// Process each thread
	for i, thread := range threads {
		log.Printf("\nProcessing thread %d/%d: %s", i+1, len(threads), thread.Title)

		// Check if category mapping exists
		categoryID, ok := NodeToCategory[thread.NodeID]
		if !ok {
			log.Printf("✗ Skipped: no category mapping for node_id %d", thread.NodeID)
			continue
		}

		// Get posts for thread
		posts, err := getXenForoPosts(thread.ThreadID)
		if err != nil {
			log.Printf("✗ Failed to fetch posts: %v", err)
			progress.FailedThreads = append(progress.FailedThreads, thread.ThreadID)
			saveProgress()
			continue
		}

		// Get attachments for thread
		attachments, err := getXenForoAttachments(thread.ThreadID)
		if err != nil {
			log.Printf("⚠ Warning: Failed to fetch attachments: %v", err)
			// Continue anyway, just without attachments
		}

		// Download attachments
		if len(attachments) > 0 && !dryRun {
			log.Printf("  Downloading %d attachments...", len(attachments))
			downloadAttachments(attachments)
		}

		// Create discussion
		discussionID := ""
		discussionNumber := 0

		for j, post := range posts {
			// Convert BB-codes to Markdown
			markdown := convertBBCodeToMarkdown(post.Message)

			// Replace attachment links
			markdown = replaceAttachmentLinks(markdown, attachments)

			// Format message with metadata
			body := formatMessage(post.Username, post.PostDate, thread.ThreadID, markdown)

			if j == 0 {
				// Create discussion from first post
				if dryRun {
					log.Printf("  [DRY-RUN] Would create discussion: %s", thread.Title)
					if verbose {
						fmt.Printf("\n--- Discussion Body Preview ---\n%s\n--- End Preview ---\n", body)
					}
				} else {
					id, num, err := createGitHubDiscussion(thread.Title, body, categoryID)
					if err != nil {
						log.Printf("✗ Failed to create discussion: %v", err)
						progress.FailedThreads = append(progress.FailedThreads, thread.ThreadID)
						saveProgress()
						continue
					}
					discussionID = id
					discussionNumber = num
					log.Printf("✓ Created discussion #%d", discussionNumber)
				}
			} else {
				// Add comment to discussion
				if dryRun {
					log.Printf("  [DRY-RUN] Would add comment by %s", post.Username)
					if verbose {
						fmt.Printf("\n--- Comment Preview ---\n%s\n--- End Preview ---\n", body)
					}
				} else if discussionID != "" {
					err := addGitHubComment(discussionID, body)
					if err != nil {
						log.Printf("✗ Failed to add comment: %v", err)
					} else {
						log.Printf("  ✓ Added comment by %s", post.Username)
					}
				}
			}

			// Rate limiting
			if !dryRun {
				time.Sleep(1 * time.Second)
			}
		}

		// Mark thread as completed
		progress.CompletedThreads = append(progress.CompletedThreads, thread.ThreadID)
		progress.LastThreadID = thread.ThreadID
		saveProgress()
	}

	// Print summary
	printSummary()
}

func runPreflightChecks() error {
	// Check if dry-run mode
	if dryRun {
		log.Println("  Running in DRY-RUN mode - no actual changes will be made")
		return nil
	}

	// Test XenForo API access
	resp, err := client.R().
		SetHeader("XF-Api-Key", XenForoAPIKey).
		SetHeader("XF-Api-User", XenForoAPIUser).
		Get(XenForoAPIURL + "/")
	if err != nil {
		return fmt.Errorf("XenForo API connection failed: %v", err)
	}
	if resp.StatusCode() == 401 {
		return fmt.Errorf("XenForo API authentication failed - check API key and user ID")
	}
	log.Println("  ✓ XenForo API access verified")

	// Get repository ID for GitHub
	if !dryRun && githubClient != nil {
		parts := strings.Split(GitHubRepo, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid GitHub repository format - expected 'owner/repo'")
		}

		var query struct {
			Repository struct {
				ID                   string
				DiscussionsEnabled   bool
				DiscussionCategories struct {
					Nodes []struct {
						ID   string
						Name string
					}
				} `graphql:"discussionCategories(first: 100)"`
			} `graphql:"repository(owner: $owner, name: $name)"`
		}

		variables := map[string]interface{}{
			"owner": githubv4.String(parts[0]),
			"name":  githubv4.String(parts[1]),
		}

		err := githubClient.Query(context.Background(), &query, variables)
		if err != nil {
			return fmt.Errorf("GitHub API access failed: %v", err)
		}

		if !query.Repository.DiscussionsEnabled {
			return fmt.Errorf("GitHub Discussions is not enabled for repository %s", GitHubRepo)
		}

		repositoryID = query.Repository.ID
		log.Println("  ✓ GitHub API access verified")
		log.Println("  ✓ GitHub Discussions is enabled")

		// Verify category mappings
		validCategories := make(map[string]bool)
		for _, cat := range query.Repository.DiscussionCategories.Nodes {
			validCategories[cat.ID] = true
		}

		for nodeID, categoryID := range NodeToCategory {
			if !validCategories[categoryID] {
				return fmt.Errorf("invalid category ID '%s' for node %d", categoryID, nodeID)
			}
		}
		log.Println("  ✓ All category mappings are valid")
	}

	// Create attachments directory
	if err := os.MkdirAll(AttachmentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create attachments directory: %v", err)
	}
	log.Println("  ✓ Attachments directory ready")

	return nil
}

func convertBBCodeToMarkdown(bbcode string) string {
	// Handle empty or whitespace-only input
	if strings.TrimSpace(bbcode) == "" {
		return ""
	}
	
	// First, handle multi-line code blocks
	bbcode = regexp.MustCompile(`(?s)\[code\](.*?)\[/code\]`).ReplaceAllStringFunc(bbcode, func(match string) string {
		parts := regexp.MustCompile(`(?s)\[code\](.*?)\[/code\]`).FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		content := parts[1]
		// Check if content starts with newline (multi-line)
		if strings.HasPrefix(content, "\n") {
			return "\n```\n" + strings.TrimSpace(content) + "\n```\n"
		}
		// Single line code
		return "\n```\n" + strings.TrimSpace(content) + "\n```\n"
	})

	// Handle quotes with attribution (extract just the username, ignore extra params)
	bbcode = regexp.MustCompile(`(?s)\[quote="([^,"]+)(?:,[^\]]+)?"\](.*?)\[/quote\]`).ReplaceAllStringFunc(bbcode, func(match string) string {
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
	bbcode = regexp.MustCompile(`(?s)\[quote\](.*?)\[/quote\]`).ReplaceAllStringFunc(bbcode, func(match string) string {
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

	// URLs with quotes first
	bbcode = regexp.MustCompile(`\[url="([^"]+)"\](.*?)\[/url\]`).ReplaceAllString(bbcode, "[$2]($1)")

	// Simple replacements
	replacements := []struct {
		pattern     *regexp.Regexp
		replacement string
	}{
		// Text formatting
		{regexp.MustCompile(`\[b\](.*?)\[/b\]`), "**$1**"},
		{regexp.MustCompile(`\[i\](.*?)\[/i\]`), "*$1*"},
		{regexp.MustCompile(`\[u\](.*?)\[/u\]`), "<u>$1</u>"},
		{regexp.MustCompile(`\[s\](.*?)\[/s\]`), "~~$1~~"},
		{regexp.MustCompile(`\[strike\](.*?)\[/strike\]`), "~~$1~~"},

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

	result := bbcode
	for _, r := range replacements {
		result = r.pattern.ReplaceAllString(result, r.replacement)
	}

	// Clean up any remaining unhandled BB codes (but not markdown links or ATTACH tags)
	// Use regexp2 with negative lookahead to exclude markdown links [text](url)
	cleanupPattern := regexp2.MustCompile(`\[/?[a-zA-Z][a-zA-Z0-9=_-]*\](?!\()`, 0)
	result, _ = cleanupPattern.ReplaceFunc(result, func(m regexp2.Match) string {
		match := m.String()
		// Preserve ATTACH tags for later processing (both opening and closing)
		if strings.HasPrefix(match, "[ATTACH") || match == "[/ATTACH]" {
			return match
		}
		return "" // Remove other BB-code tags
	}, -1, -1)

	// Clean up excessive newlines (but preserve intentional trailing newlines)
	result = regexp.MustCompile(`\n{3,}`).ReplaceAllString(result, "\n\n")

	// Trim leading/trailing spaces but preserve newlines for proper formatting
	result = strings.Trim(result, " \t")

	return result
}

func formatMessage(username string, postDate int64, threadID int, content string) string {
	timestamp := time.Unix(postDate, 0).UTC().Format("2006-01-02 15:04:05 UTC")
	return fmt.Sprintf(`---
Author: %s
Posted: %s
Original Thread ID: %d
---

%s`, username, timestamp, threadID, content)
}

func getXenForoThreads(nodeID int) ([]XenForoThread, error) {
	var threads []XenForoThread
	page := 1

	for {
		resp, err := retryableRequest(func() (*resty.Response, error) {
			return client.R().
				SetHeader("XF-Api-Key", XenForoAPIKey).
				SetHeader("XF-Api-User", XenForoAPIUser).
				SetQueryParams(map[string]string{
					"page":    fmt.Sprintf("%d", page),
					"node_id": fmt.Sprintf("%d", nodeID),
				}).
				Get(XenForoAPIURL + "/threads")
		})

		if err != nil {
			return nil, err
		}

		if resp.StatusCode() != 200 {
			return nil, fmt.Errorf("XenForo API error: %s", resp.String())
		}

		var result struct {
			Threads    []XenForoThread `json:"threads"`
			Pagination struct {
				CurrentPage int `json:"current_page"`
				TotalPages  int `json:"total_pages"`
			} `json:"pagination"`
		}

		err = json.Unmarshal(resp.Body(), &result)
		if err != nil {
			return nil, err
		}

		threads = append(threads, result.Threads...)

		if result.Pagination.CurrentPage >= result.Pagination.TotalPages {
			break
		}

		page++
		time.Sleep(1 * time.Second)
	}

	return threads, nil
}

func getXenForoPosts(threadID int) ([]XenForoPost, error) {
	var posts []XenForoPost
	page := 1

	for {
		resp, err := retryableRequest(func() (*resty.Response, error) {
			return client.R().
				SetHeader("XF-Api-Key", XenForoAPIKey).
				SetHeader("XF-Api-User", XenForoAPIUser).
				SetQueryParam("page", fmt.Sprintf("%d", page)).
				Get(fmt.Sprintf("%s/threads/%d/posts", XenForoAPIURL, threadID))
		})

		if err != nil {
			return nil, err
		}

		if resp.StatusCode() != 200 {
			return nil, fmt.Errorf("XenForo API error: %s", resp.String())
		}

		var result struct {
			Posts      []XenForoPost `json:"posts"`
			Pagination struct {
				CurrentPage int `json:"current_page"`
				TotalPages  int `json:"total_pages"`
			} `json:"pagination"`
		}

		err = json.Unmarshal(resp.Body(), &result)
		if err != nil {
			return nil, err
		}

		posts = append(posts, result.Posts...)

		if result.Pagination.CurrentPage >= result.Pagination.TotalPages {
			break
		}

		page++
		time.Sleep(1 * time.Second)
	}

	return posts, nil
}

func getXenForoAttachments(threadID int) ([]XenForoAttachment, error) {
	resp, err := retryableRequest(func() (*resty.Response, error) {
		return client.R().
			SetHeader("XF-Api-Key", XenForoAPIKey).
			SetHeader("XF-Api-User", XenForoAPIUser).
			Get(fmt.Sprintf("%s/threads/%d/attachments", XenForoAPIURL, threadID))
	})

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("XenForo API error (attachments): %s", resp.String())
	}

	var result struct {
		Attachments []XenForoAttachment `json:"attachments"`
	}

	err = json.Unmarshal(resp.Body(), &result)
	return result.Attachments, err
}

func downloadAttachments(attachments []XenForoAttachment) {
	for _, attachment := range attachments {
		if dryRun {
			log.Printf("    [DRY-RUN] Would download: %s", attachment.Filename)
			continue
		}

		// Determine file extension and create directory
		ext := strings.ToLower(filepath.Ext(attachment.Filename))
		if ext == "" {
			ext = ".unknown"
		}
		ext = strings.TrimPrefix(ext, ".")

		dir := filepath.Join(AttachmentsDir, ext)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("    ✗ Failed to create directory %s: %v", dir, err)
			continue
		}

		// Download file
		filename := fmt.Sprintf("attachment_%d_%s", attachment.AttachmentID, attachment.Filename)
		filePath := filepath.Join(dir, filename)

		if _, err := os.Stat(filePath); err == nil {
			log.Printf("    ⏭ Skipped (already exists): %s", filename)
			continue
		}

		resp, err := client.R().
			SetHeader("XF-Api-Key", XenForoAPIKey).
			SetHeader("XF-Api-User", XenForoAPIUser).
			SetOutput(filePath).
			Get(attachment.ViewURL)

		if err != nil || resp.StatusCode() != 200 {
			log.Printf("    ✗ Failed to download %s: %v", filename, err)
			continue
		}

		log.Printf("    ✓ Downloaded: %s", filename)
		time.Sleep(500 * time.Millisecond) // Be nice to the server
	}
}

func replaceAttachmentLinks(message string, attachments []XenForoAttachment) string {
	for _, attachment := range attachments {
		ext := strings.ToLower(filepath.Ext(attachment.Filename))
		if ext == "" {
			ext = ".unknown"
		}
		ext = strings.TrimPrefix(ext, ".")

		filename := fmt.Sprintf("attachment_%d_%s", attachment.AttachmentID, attachment.Filename)
		relativePath := fmt.Sprintf("./%s/%s", ext, filename)

		// Determine if it's an image
		isImage := ext == "png" || ext == "jpg" || ext == "jpeg" || ext == "gif" || ext == "webp"

		// Replace BB-code with appropriate markdown
		bbCode := fmt.Sprintf("[ATTACH=%d]", attachment.AttachmentID)
		bbCodeFull := fmt.Sprintf("[ATTACH=full]%d[/ATTACH]", attachment.AttachmentID)

		var markdownLink string
		if isImage {
			markdownLink = fmt.Sprintf("![%s](%s)", attachment.Filename, relativePath)
		} else {
			markdownLink = fmt.Sprintf("[%s](%s)", attachment.Filename, relativePath)
		}

		message = strings.ReplaceAll(message, bbCode, markdownLink)
		message = strings.ReplaceAll(message, bbCodeFull, markdownLink)
	}

	// Log any remaining unhandled attach codes
	remaining := regexp.MustCompile(`\[ATTACH[^]]*\]`).FindAllString(message, -1)
	for _, code := range remaining {
		log.Printf("    ⚠ Unhandled attachment code: %s", code)
	}

	return message
}

func createGitHubDiscussion(title, body, categoryID string) (string, int, error) {
	var mutation struct {
		CreateDiscussion struct {
			Discussion struct {
				ID     string
				Number int
			}
		} `graphql:"createDiscussion(input: $input)"`
	}

	input := githubv4.CreateDiscussionInput{
		RepositoryID: githubv4.ID(repositoryID),
		Title:        githubv4.String(title),
		Body:         githubv4.String(body),
		CategoryID:   githubv4.ID(categoryID),
	}

	err := githubClient.Mutate(context.Background(), &mutation, input, nil)
	if err != nil {
		return "", 0, err
	}

	return mutation.CreateDiscussion.Discussion.ID, mutation.CreateDiscussion.Discussion.Number, nil
}

func addGitHubComment(discussionID, body string) error {
	var mutation struct {
		AddDiscussionComment struct {
			Comment struct {
				ID githubv4.ID
			}
		} `graphql:"addDiscussionComment(input: $input)"`
	}

	input := githubv4.AddDiscussionCommentInput{
		DiscussionID: githubv4.ID(discussionID),
		Body:         githubv4.String(body),
	}

	return githubClient.Mutate(context.Background(), &mutation, input, nil)
}

func retryableRequest(req func() (*resty.Response, error)) (*resty.Response, error) {
	for i := 0; i < MaxRetries; i++ {
		resp, err := req()

		// If successful or not a rate limit error, return
		if err != nil {
			return nil, err
		}

		if resp.StatusCode() != 429 {
			return resp, nil
		}

		// Rate limited - exponential backoff
		if i < MaxRetries-1 {
			delay := time.Duration(math.Pow(2, float64(i))) * time.Second
			log.Printf("    ⏳ Rate limited, retrying in %v...", delay)
			time.Sleep(delay)
		}
	}

	return nil, fmt.Errorf("max retries (%d) exceeded", MaxRetries)
}

func loadProgress() *MigrationProgress {
	progress := &MigrationProgress{
		CompletedThreads: []int{},
		FailedThreads:    []int{},
	}

	data, err := os.ReadFile(ProgressFile)
	if err != nil {
		return progress
	}

	json.Unmarshal(data, progress)
	return progress
}

func saveProgress() {
	progress.LastUpdated = time.Now().Unix()
	data, _ := json.MarshalIndent(progress, "", "  ")
	os.WriteFile(ProgressFile, data, 0644)
}

func filterCompletedThreads(threads []XenForoThread) []XenForoThread {
	completed := make(map[int]bool)
	for _, id := range progress.CompletedThreads {
		completed[id] = true
	}

	var filtered []XenForoThread
	for _, thread := range threads {
		// Skip if already completed
		if completed[thread.ThreadID] {
			continue
		}
		// Include threads that haven't been processed yet
		// (either newer than LastThreadID or were skipped before)
		filtered = append(filtered, thread)
	}

	return filtered
}

func printSummary() {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("Migration Summary")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Completed threads: %d\n", len(progress.CompletedThreads))
	fmt.Printf("Failed threads: %d\n", len(progress.FailedThreads))

	if len(progress.FailedThreads) > 0 {
		fmt.Println("\nFailed thread IDs:")
		for _, id := range progress.FailedThreads {
			fmt.Printf("  - %d\n", id)
		}
	}

	if dryRun {
		fmt.Println("\n[DRY-RUN MODE] No actual changes were made")
	}
}
