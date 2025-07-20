package migration

import (
	"context"
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

func (r *Runner) RunMigration(ctx context.Context) error {
	log.Printf("Fetching threads from forum node %d...", r.config.GitHub.XenForoNodeID)
	threads, err := r.xenforoClient.GetThreads(r.config.GitHub.XenForoNodeID)
	if err != nil {
		return err
	}
	log.Printf("✓ Found %d threads to migrate", len(threads))

	threads = r.tracker.FilterCompletedThreads(threads)
	log.Printf("✓ %d threads remaining after filtering completed ones", len(threads))

	for i, thread := range threads {
		log.Printf("\nProcessing thread %d/%d: %s", i+1, len(threads), thread.Title)

		if err := r.processThread(ctx, thread); err != nil {
			log.Printf("✗ Failed to process thread %d: %v", thread.ThreadID, err)
			if markErr := r.tracker.MarkFailed(thread.ThreadID); markErr != nil {
				log.Printf("✗ Warning: Failed to mark thread %d as failed in progress tracker: %v", thread.ThreadID, markErr)
			}
			continue
		}

		if err := r.tracker.MarkCompleted(thread.ThreadID); err != nil {
			log.Printf("✗ Warning: Failed to mark thread %d as completed in progress tracker: %v", thread.ThreadID, err)
		}
	}

	r.tracker.PrintSummary()
	return nil
}

func (r *Runner) processThread(ctx context.Context, thread xenforo.Thread) error {
	categoryID := r.config.GitHub.GitHubCategoryID

	posts, err := r.xenforoClient.GetPosts(thread)
	if err != nil {
		return err
	}
	log.Printf("  ✓ Found %d posts for thread", len(posts))

	var threadAttachments []xenforo.Attachment
	for _, post := range posts {
		threadAttachments = append(threadAttachments, post.Attachments...)
	}

	if len(threadAttachments) > 0 {
		log.Printf("  ✓ Found %d attachments across all posts", len(threadAttachments))
		log.Printf("  Downloading attachments...")
		if err := r.downloader.DownloadAttachments(threadAttachments); err != nil {
			log.Printf("✗ Warning: Failed to download attachments for thread %d: %v", thread.ThreadID, err)
		}
	}

	var discussionID string
	discussionNumber := 0

	for j, post := range posts {
		markdown := r.processor.ProcessContent(post.Message)

		markdown = r.downloader.ReplaceAttachmentLinks(markdown, threadAttachments)

		body, err := r.processor.FormatMessage(post.Username, post.PostDate, thread.ThreadID, markdown)
		if err != nil {
			log.Printf("  Error formatting message for post by %s: %v", post.Username, err)
			return fmt.Errorf("failed to format message: %w", err)
		}

		if j == 0 {
			if r.config.Migration.DryRun {
				log.Printf("  [DRY-RUN] Would create discussion: %s", thread.Title)
				if r.config.Migration.Verbose {
					log.Printf("\n--- Discussion Body Preview ---\n%s\n--- End Preview ---\n", body)
				}
			} else {
				result, err := r.githubClient.CreateDiscussion(ctx, thread.Title, body, categoryID)
				if err != nil {
					return err
				}
				discussionID = result.ID
				discussionNumber = result.Number
				log.Printf("✓ Created discussion #%d", discussionNumber)
			}
		} else {
			if r.config.Migration.DryRun {
				log.Printf("  [DRY-RUN] Would add comment by %s", post.Username)
				if r.config.Migration.Verbose {
					log.Printf("\n--- Comment Preview ---\n%s\n--- End Preview ---\n", body)
				}
			} else if discussionID != "" {
				if err := r.githubClient.AddComment(ctx, discussionID, body); err != nil {
					log.Printf("✗ Failed to add comment: %v", err)
				} else {
					log.Printf("  ✓ Added comment by %s", post.Username)
				}
			}
		}

		if !r.config.Migration.DryRun {
			time.Sleep(1 * time.Second)
		}
	}

	return nil
}
