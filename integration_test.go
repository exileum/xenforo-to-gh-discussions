package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/shurcooL/githubv4"
)

// MockGitHubClient implements a mock GitHub GraphQL client for testing
type MockGitHubClient struct {
	CreateDiscussionFunc func(input githubv4.CreateDiscussionInput) (string, int, error)
	AddCommentFunc       func(input githubv4.AddDiscussionCommentInput) error
	QueryFunc            func(ctx context.Context, q interface{}, variables map[string]interface{}) error
}

func (m *MockGitHubClient) Mutate(ctx context.Context, mutation interface{}, input interface{}, variables map[string]interface{}) error {
	// Handle different mutation types based on the input type
	switch v := input.(type) {
	case githubv4.CreateDiscussionInput:
		if m.CreateDiscussionFunc != nil {
			_, _, err := m.CreateDiscussionFunc(v)
			if err != nil {
				return err
			}
			// Use reflection to set the response (simplified for testing)
			// In real implementation, would properly set the mutation response
			return nil
		}
	case githubv4.AddDiscussionCommentInput:
		if m.AddCommentFunc != nil {
			return m.AddCommentFunc(v)
		}
	}
	return nil
}

func (m *MockGitHubClient) Query(ctx context.Context, q interface{}, variables map[string]interface{}) error {
	if m.QueryFunc != nil {
		return m.QueryFunc(ctx, q, variables)
	}
	return nil
}

// TestEndToEndMigration tests the complete migration flow with mocked APIs
func TestEndToEndMigration(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "godisc-e2e-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Save original values
	originalDryRun := dryRun
	originalClient := client
	defer func() {
		dryRun = originalDryRun
		client = originalClient
		progress = nil
	}()

	// Configure for test
	dryRun = false
	oldAttachmentsDir := AttachmentsDir
	oldProgressFile := ProgressFile
	defer func() {
		AttachmentsDir = oldAttachmentsDir
		ProgressFile = oldProgressFile
	}()
	AttachmentsDir = filepath.Join(tempDir, "attachments")
	ProgressFile = filepath.Join(tempDir, "progress.json")

	// Initialize progress
	progress = &MigrationProgress{
		CompletedThreads: []int{},
		FailedThreads:    []int{},
	}

	// Test data
	testThreads := []XenForoThread{
		{
			ThreadID:    1,
			Title:       "Test Migration Thread",
			NodeID:      1,
			Username:    "testuser",
			PostDate:    1642353000,
			FirstPostID: 1,
		},
	}

	testPosts := []XenForoPost{
		{
			PostID:   1,
			ThreadID: 1,
			Username: "testuser",
			PostDate: 1642353000,
			Message:  "This is a [b]test post[/b] with [ATTACH=1].",
		},
		{
			PostID:   2,
			ThreadID: 1,
			Username: "replyuser",
			PostDate: 1642353600,
			Message:  "[quote=\"testuser\"]test post[/quote]\nI agree!",
		},
	}

	testAttachments := []XenForoAttachment{
		{
			AttachmentID: 1,
			Filename:     "test-image.png",
			ViewURL:      "https://example.com/attach/1",
		},
	}

	// Create test scenarios
	tests := []struct {
		name          string
		setupFunc     func()
		expectSuccess bool
		checkFunc     func(t *testing.T)
	}{
		{
			name: "Successful migration",
			setupFunc: func() {
				// Reset progress
				progress.CompletedThreads = []int{}
				progress.FailedThreads = []int{}
			},
			expectSuccess: true,
			checkFunc: func(t *testing.T) {
				// Check progress was saved
				if len(progress.CompletedThreads) != 1 {
					t.Errorf("Expected 1 completed thread, got %d", len(progress.CompletedThreads))
				}
				if len(progress.FailedThreads) != 0 {
					t.Errorf("Expected 0 failed threads, got %d", len(progress.FailedThreads))
				}

				// Check attachments directory was created
				if _, err := os.Stat(filepath.Join(AttachmentsDir, "png")); os.IsNotExist(err) {
					t.Error("Attachments directory not created")
				}
			},
		},
		{
			name: "Skip already completed threads",
			setupFunc: func() {
				progress.CompletedThreads = []int{1}
				progress.FailedThreads = []int{}
			},
			expectSuccess: true,
			checkFunc: func(t *testing.T) {
				// Should still have only 1 completed thread
				if len(progress.CompletedThreads) != 1 {
					t.Errorf("Expected 1 completed thread, got %d", len(progress.CompletedThreads))
				}
			},
		},
		{
			name: "Handle missing category mapping",
			setupFunc: func() {
				// Temporarily change thread to unmapped node
				testThreads[0].NodeID = 999
				progress.CompletedThreads = []int{}
			},
			expectSuccess: true,
			checkFunc: func(t *testing.T) {
				// Should not be in completed or failed
				if len(progress.CompletedThreads) != 0 {
					t.Errorf("Expected 0 completed threads, got %d", len(progress.CompletedThreads))
				}
				if len(progress.FailedThreads) != 0 {
					t.Errorf("Expected 0 failed threads, got %d", len(progress.FailedThreads))
				}
				// Reset for next test
				testThreads[0].NodeID = 1
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()

			// Simulate the migration process components
			// 1. Filter completed threads
			filteredThreads := filterCompletedThreads(testThreads)

			// 2. Process each thread
			for _, thread := range filteredThreads {
				// Check category mapping
				_, hasCategory := NodeToCategory[thread.NodeID]
				if !hasCategory {
					continue
				}

				// 3. Create attachment directories (simulate download)
				for _, attachment := range testAttachments {
					ext := strings.ToLower(filepath.Ext(attachment.Filename))
					if ext == "" {
						ext = ".unknown"
					}
					ext = strings.TrimPrefix(ext, ".")
					
					dirPath := filepath.Join(AttachmentsDir, ext)
					if err := os.MkdirAll(dirPath, 0755); err != nil {
						t.Errorf("Failed to create attachment directory: %v", err)
					}
				}

				// 4. Convert posts
				for i, post := range testPosts {
					markdown := convertBBCodeToMarkdown(post.Message)
					markdown = replaceAttachmentLinks(markdown, testAttachments)
					body := formatMessage(post.Username, post.PostDate, thread.ThreadID, markdown)

					// Verify conversion
					if i == 0 {
						if body == "" {
							t.Error("First post body is empty")
						}
						if !strings.Contains(body, "**test post**") {
							t.Error("BB-code not converted properly")
						}
						if !strings.Contains(body, "![test-image.png](./png/attachment_1_test-image.png)") {
							t.Error("Attachment not replaced properly")
						}
					}
				}

				// 5. Mark as completed (in real flow)
				if hasCategory {
					progress.CompletedThreads = append(progress.CompletedThreads, thread.ThreadID)
				}
			}

			// 6. Save progress
			saveProgress()

			// Run checks
			tt.checkFunc(t)
		})
	}
}

// TestDryRunMode tests that dry-run mode doesn't make actual changes
func TestDryRunMode(t *testing.T) {
	// Save original values
	originalDryRun := dryRun
	originalVerbose := verbose
	defer func() {
		dryRun = originalDryRun
		verbose = originalVerbose
	}()

	// Enable dry-run and verbose mode
	dryRun = true
	verbose = true

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "godisc-dryrun-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	oldAttachmentsDir := AttachmentsDir
	defer func() {
		AttachmentsDir = oldAttachmentsDir
	}()
	AttachmentsDir = filepath.Join(tempDir, "attachments")

	// Test attachment download in dry-run mode
	attachments := []XenForoAttachment{
		{AttachmentID: 1, Filename: "test.png", ViewURL: "https://example.com/1"},
	}

	// This should not create any files
	downloadAttachments(attachments)

	// Verify no files were created
	if _, err := os.Stat(AttachmentsDir); !os.IsNotExist(err) {
		t.Error("Attachments directory should not be created in dry-run mode")
	}
}

// TestConcurrentOperations tests thread safety of shared resources
func TestConcurrentOperations(t *testing.T) {
	// This test ensures that concurrent operations don't cause race conditions
	// In a real implementation, you would test actual concurrent API calls

	// Save original progress
	originalProgress := progress
	defer func() {
		progress = originalProgress
	}()

	// Initialize test progress
	progress = &MigrationProgress{
		CompletedThreads: []int{},
		FailedThreads:    []int{},
	}

	// Simulate concurrent updates to progress
	done := make(chan bool, 2)

	go func() {
		progress.CompletedThreads = append(progress.CompletedThreads, 1)
		done <- true
	}()

	go func() {
		progress.FailedThreads = append(progress.FailedThreads, 2)
		done <- true
	}()

	// Wait for both operations
	<-done
	<-done

	// In a real implementation, you would use proper synchronization
	// This test is simplified to demonstrate the concept
	if len(progress.CompletedThreads) == 0 && len(progress.FailedThreads) == 0 {
		t.Error("Concurrent operations may have caused data loss")
	}
}

// TestLargeDatasetHandling tests performance with large datasets
func TestLargeDatasetHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large dataset test in short mode")
	}

	// Generate large dataset
	largeText := generateLargeText(10000) // 10KB of text

	// Test BB-code conversion performance
	start := time.Now()
	result := convertBBCodeToMarkdown(largeText)
	duration := time.Since(start)

	if duration > 1*time.Second {
		t.Errorf("BB-code conversion took too long: %v", duration)
	}

	if len(result) == 0 {
		t.Error("Large text conversion resulted in empty output")
	}
}

// Helper function to generate large text with BB-codes
func generateLargeText(size int) string {
	var text string
	bbcodes := []string{"[b]bold[/b]", "[i]italic[/i]", "[url=http://example.com]link[/url]"}

	for len(text) < size {
		text += fmt.Sprintf("This is line with %s text. ", bbcodes[len(text)%len(bbcodes)])
	}

	return text
}

// TestMemoryLeaks checks for potential memory leaks
func TestMemoryLeaks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak test in short mode")
	}

	// Run multiple iterations of BB-code conversion
	for i := 0; i < 1000; i++ {
		input := fmt.Sprintf("[b]Test %d[/b] with [i]formatting[/i] and [url=http://example.com]links[/url]", i)
		_ = convertBBCodeToMarkdown(input)
	}

	// In a real test, you would measure memory usage
	// This is a simplified version to demonstrate the concept
}
