package migration

import (
	"fmt"
	"log"
	"time"

	"github.com/exileum/xenforo-to-gh-discussions/internal/attachments"
	"github.com/exileum/xenforo-to-gh-discussions/internal/bbcode"
	"github.com/exileum/xenforo-to-gh-discussions/internal/config"
	"github.com/exileum/xenforo-to-gh-discussions/internal/github"
	"github.com/exileum/xenforo-to-gh-discussions/internal/progress"
	"github.com/exileum/xenforo-to-gh-discussions/internal/xenforo"
)

type Runner struct {
	config        *config.Config
	xenforoClient *xenforo.Client
	githubClient  *github.Client
	tracker       *progress.Tracker
	downloader    *attachments.Downloader
	processor     *bbcode.MessageProcessor
}

func NewRunner(cfg *config.Config, xenforoClient *xenforo.Client, githubClient *github.Client, tracker *progress.Tracker, downloader *attachments.Downloader) *Runner {
	return &Runner{
		config:        cfg,
		xenforoClient: xenforoClient,
		githubClient:  githubClient,
		tracker:       tracker,
		downloader:    downloader,
		processor:     bbcode.NewMessageProcessor(),
	}
}

func (r *Runner) RunMigration() error {
	// Get threads from XenForo
	log.Printf("Fetching threads from forum node %d...", r.config.XenForo.NodeID)
	threads, err := r.xenforoClient.GetThreads(r.config.XenForo.NodeID)
	if err != nil {
		return err
	}
	log.Printf("✓ Found %d threads to migrate", len(threads))

	// Filter out already completed threads
	threads = r.tracker.FilterCompletedThreads(threads)
	log.Printf("✓ %d threads remaining after filtering completed ones", len(threads))

	// Process each thread
	for i, thread := range threads {
		log.Printf("\nProcessing thread %d/%d: %s", i+1, len(threads), thread.Title)

		if err := r.processThread(thread); err != nil {
			log.Printf("✗ Failed to process thread %d: %v", thread.ThreadID, err)
			if markErr := r.tracker.MarkFailed(thread.ThreadID); markErr != nil {
				log.Printf("✗ Warning: Failed to mark thread %d as failed in progress tracker: %v", thread.ThreadID, markErr)
			}
			continue
		}

		if err := r.tracker.MarkCompleted(thread.ThreadID); err != nil {
			log.Printf("✗ Warning: Failed to mark thread %d as completed in progress tracker: %v", thread.ThreadID, err)
			// Continue processing despite tracking error - the thread was actually processed successfully
		}
	}

	r.tracker.PrintSummary()
	return nil
}

func (r *Runner) processThread(thread xenforo.Thread) error {
	// Check if category mapping exists
	categoryID, ok := r.config.GitHub.Categories[thread.NodeID]
	if !ok {
		log.Printf("✗ Skipped: no category mapping for node_id %d", thread.NodeID)
		return nil
	}

	// Get posts for thread
	posts, err := r.xenforoClient.GetPosts(thread.ThreadID)
	if err != nil {
		return err
	}

	// Get attachments for thread
	attachments, err := r.xenforoClient.GetAttachments(thread.ThreadID)
	if err != nil {
		log.Printf("⚠ Warning: Failed to fetch attachments: %v", err)
		// Continue anyway, just without attachments
	}

	// Download attachments
	if len(attachments) > 0 {
		log.Printf("  Downloading %d attachments...", len(attachments))
		if err := r.downloader.DownloadAttachments(attachments); err != nil {
			log.Printf("✗ Warning: Failed to download attachments for thread %d: %v", thread.ThreadID, err)
			// Continue processing without attachments - the thread content can still be migrated
		}
	}

	// Process posts
	var discussionID string
	discussionNumber := 0

	for j, post := range posts {
		// Convert BB-codes to Markdown
		markdown := r.processor.ProcessContent(post.Message)

		// Replace attachment links
		markdown = r.downloader.ReplaceAttachmentLinks(markdown, attachments)

		// Format message with metadata
		body, err := r.processor.FormatMessage(post.Username, post.PostDate, thread.ThreadID, markdown)
		if err != nil {
			log.Printf("  Error formatting message for post by %s: %v", post.Username, err)
			return fmt.Errorf("failed to format message: %w", err)
		}

		if j == 0 {
			// Create discussion from first post
			if r.config.Migration.DryRun {
				log.Printf("  [DRY-RUN] Would create discussion: %s", thread.Title)
				if r.config.Migration.Verbose {
					log.Printf("\n--- Discussion Body Preview ---\n%s\n--- End Preview ---\n", body)
				}
			} else {
				result, err := r.githubClient.CreateDiscussion(thread.Title, body, categoryID)
				if err != nil {
					return err
				}
				discussionID = result.ID
				discussionNumber = result.Number
				log.Printf("✓ Created discussion #%d", discussionNumber)
			}
		} else {
			// Add comment to discussion
			if r.config.Migration.DryRun {
				log.Printf("  [DRY-RUN] Would add comment by %s", post.Username)
				if r.config.Migration.Verbose {
					log.Printf("\n--- Comment Preview ---\n%s\n--- End Preview ---\n", body)
				}
			} else if discussionID != "" {
				if err := r.githubClient.AddComment(discussionID, body); err != nil {
					log.Printf("✗ Failed to add comment: %v", err)
				} else {
					log.Printf("  ✓ Added comment by %s", post.Username)
				}
			}
		}

		// Rate limiting
		if !r.config.Migration.DryRun {
			time.Sleep(1 * time.Second)
		}
	}

	return nil
}
