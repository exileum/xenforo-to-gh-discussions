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
	threads, err := r.xenforoClient.GetThreads(ctx, r.config.GitHub.XenForoNodeID)
	if err != nil {
		return err
	}
	log.Printf("✓ Found %d threads to migrate", len(threads))

	threads = r.tracker.FilterCompletedThreads(threads)
	log.Printf("✓ %d threads remaining after filtering completed ones", len(threads))

	for i, thread := range threads {
		// Check context cancellation before each thread
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		log.Printf("\nProcessing thread %d/%d: %s", i+1, len(threads), thread.Title)

		if err := r.processThread(ctx, thread); err != nil {
			log.Printf("✗ Failed to process thread %d: %v", thread.ThreadID, err)
			if markErr := r.tracker.MarkFailed(ctx, thread.ThreadID); markErr != nil {
				log.Printf("✗ Warning: Failed to mark thread %d as failed in progress tracker: %v", thread.ThreadID, markErr)
			}
			continue
		}

		if err := r.tracker.MarkCompleted(ctx, thread.ThreadID); err != nil {
			log.Printf("✗ Warning: Failed to mark thread %d as completed in progress tracker: %v", thread.ThreadID, err)
		}
	}

	r.tracker.PrintSummary()
	return nil
}

func (r *Runner) processThread(ctx context.Context, thread xenforo.Thread) error {
	posts, err := r.fetchPosts(ctx, thread)
	if err != nil {
		return err
	}

	threadAttachments := r.collectAttachments(posts)
	if err := r.downloadAttachments(ctx, thread.ThreadID, threadAttachments); err != nil {
		// Log warning but continue processing
		log.Printf("✗ Warning: Failed to download attachments for thread %d: %v", thread.ThreadID, err)
	}

	return r.processPosts(ctx, thread, posts, threadAttachments)
}

func (r *Runner) fetchPosts(ctx context.Context, thread xenforo.Thread) ([]xenforo.Post, error) {
	posts, err := r.xenforoClient.GetPosts(ctx, thread)
	if err != nil {
		return nil, err
	}
	log.Printf("  ✓ Found %d posts for thread", len(posts))
	return posts, nil
}

func (r *Runner) collectAttachments(posts []xenforo.Post) []xenforo.Attachment {
	var threadAttachments []xenforo.Attachment
	for _, post := range posts {
		threadAttachments = append(threadAttachments, post.Attachments...)
	}
	return threadAttachments
}

func (r *Runner) downloadAttachments(ctx context.Context, threadID int, attachments []xenforo.Attachment) error {
	if len(attachments) == 0 {
		return nil
	}

	log.Printf("  ✓ Found %d attachments across all posts", len(attachments))
	log.Printf("  Downloading attachments...")
	return r.downloader.DownloadAttachments(ctx, attachments)
}

func (r *Runner) processPosts(ctx context.Context, thread xenforo.Thread, posts []xenforo.Post, threadAttachments []xenforo.Attachment) error {
	var discussionID string

	for j, post := range posts {
		// Check context cancellation before each post
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		body, err := r.formatPost(ctx, post, thread.ThreadID, threadAttachments)
		if err != nil {
			return err
		}

		if j == 0 {
			discussionID, _, err = r.createDiscussion(ctx, thread, body)
			if err != nil {
				return err
			}
		} else {
			if err := r.addComment(ctx, post, discussionID, body); err != nil {
				log.Printf("✗ Failed to add comment: %v", err)
			}
		}

		if !r.config.Migration.DryRun {
			// Context-aware sleep
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(1 * time.Second):
			}
		}
	}

	return nil
}

func (r *Runner) formatPost(ctx context.Context, post xenforo.Post, threadID int, threadAttachments []xenforo.Attachment) (string, error) {
	markdown, err := r.processor.ProcessContent(ctx, post.Message)
	if err != nil {
		return "", fmt.Errorf("failed to process BB code content: %w", err)
	}
	
	markdown, err = r.downloader.ReplaceAttachmentLinks(ctx, markdown, threadAttachments)
	if err != nil {
		return "", fmt.Errorf("failed to replace attachment links: %w", err)
	}

	body, err := r.processor.FormatMessage(post.Username, post.PostDate, threadID, markdown)
	if err != nil {
		log.Printf("  Error formatting message for post by %s: %v", post.Username, err)
		return "", fmt.Errorf("failed to format message: %w", err)
	}
	return body, nil
}

func (r *Runner) createDiscussion(ctx context.Context, thread xenforo.Thread, body string) (string, int, error) {
	categoryID := r.config.GitHub.GitHubCategoryID

	if r.config.Migration.DryRun {
		log.Printf("  [DRY-RUN] Would create discussion: %s", thread.Title)
		if r.config.Migration.Verbose {
			log.Printf("\n--- Discussion Body Preview ---\n%s\n--- End Preview ---\n", body)
		}
		return "", 0, nil
	}

	result, err := r.githubClient.CreateDiscussion(ctx, thread.Title, body, categoryID)
	if err != nil {
		return "", 0, err
	}
	log.Printf("✓ Created discussion #%d", result.Number)
	return result.ID, result.Number, nil
}

func (r *Runner) addComment(ctx context.Context, post xenforo.Post, discussionID, body string) error {
	if r.config.Migration.DryRun {
		log.Printf("  [DRY-RUN] Would add comment by %s", post.Username)
		if r.config.Migration.Verbose {
			log.Printf("\n--- Comment Preview ---\n%s\n--- End Preview ---\n", body)
		}
		return nil
	}

	if discussionID == "" {
		return nil
	}

	if err := r.githubClient.AddComment(ctx, discussionID, body); err != nil {
		return err
	}
	log.Printf("  ✓ Added comment by %s", post.Username)
	return nil
}
