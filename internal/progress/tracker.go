// Package progress provides migration progress tracking and persistence.
// It maintains state of completed and failed thread migrations, with JSON
// persistence for recovery and resumption of interrupted migrations.
package progress

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/exileum/xenforo-to-gh-discussions/internal/xenforo"
)

type MigrationProgress struct {
	LastThreadID     int   `json:"last_thread_id"`
	CompletedThreads []int `json:"completed_threads"`
	FailedThreads    []int `json:"failed_threads"`
	LastUpdated      int64 `json:"last_updated"`
}

type Tracker struct {
	progress *MigrationProgress
	persist  *Persistence
	dryRun   bool
}

func NewTracker(ctx context.Context, progressFile string, dryRun bool) (*Tracker, error) {
	persist := NewPersistence(progressFile)
	progress, err := persist.Load(ctx)
	if err != nil {
		// Return default progress on load error
		progress = &MigrationProgress{
			CompletedThreads: []int{},
			FailedThreads:    []int{},
		}
	}

	return &Tracker{
		progress: progress,
		persist:  persist,
		dryRun:   dryRun,
	}, nil
}

func (t *Tracker) GetProgress() *MigrationProgress {
	return t.progress
}

func (t *Tracker) SetResumeFrom(threadID int) {
	t.progress.LastThreadID = threadID
}

func (t *Tracker) MarkCompleted(ctx context.Context, threadID int) error {
	// Check if threadID already exists in CompletedThreads
	for _, id := range t.progress.CompletedThreads {
		if id == threadID {
			return nil // Already marked as completed, no need to add again
		}
	}

	t.progress.CompletedThreads = append(t.progress.CompletedThreads, threadID)
	t.progress.LastThreadID = threadID
	return t.save(ctx)
}

func (t *Tracker) MarkFailed(ctx context.Context, threadID int) error {
	// Check if threadID already exists in FailedThreads
	for _, id := range t.progress.FailedThreads {
		if id == threadID {
			return nil // Already marked as failed, no need to add again
		}
	}

	t.progress.FailedThreads = append(t.progress.FailedThreads, threadID)
	return t.save(ctx)
}

func (t *Tracker) FilterCompletedThreads(threads []xenforo.Thread) []xenforo.Thread {
	completed := make(map[int]bool)
	for _, id := range t.progress.CompletedThreads {
		completed[id] = true
	}

	var filtered []xenforo.Thread
	for _, thread := range threads {
		if !completed[thread.ThreadID] {
			filtered = append(filtered, thread)
		}
	}

	return filtered
}

func (t *Tracker) PrintSummary() {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("Migration Summary")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Completed threads: %d\n", len(t.progress.CompletedThreads))
	fmt.Printf("Failed threads: %d\n", len(t.progress.FailedThreads))

	if len(t.progress.FailedThreads) > 0 {
		fmt.Println("\nFailed thread IDs:")
		for _, id := range t.progress.FailedThreads {
			fmt.Printf("  - %d\n", id)
		}
	}

	if t.dryRun {
		fmt.Println("\n[DRY-RUN MODE] No actual changes were made")
	}
}

func (t *Tracker) save(ctx context.Context) error {
	t.progress.LastUpdated = time.Now().Unix()
	return t.persist.Save(ctx, t.progress)
}
