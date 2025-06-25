package progress

import (
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

func NewTracker(progressFile string, dryRun bool) (*Tracker, error) {
	persist := NewPersistence(progressFile)
	progress, err := persist.Load()
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

func (t *Tracker) MarkCompleted(threadID int) error {
	t.progress.CompletedThreads = append(t.progress.CompletedThreads, threadID)
	t.progress.LastThreadID = threadID
	return t.save()
}

func (t *Tracker) MarkFailed(threadID int) error {
	t.progress.FailedThreads = append(t.progress.FailedThreads, threadID)
	return t.save()
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

func (t *Tracker) save() error {
	t.progress.LastUpdated = time.Now().Unix()
	return t.persist.Save(t.progress)
}
