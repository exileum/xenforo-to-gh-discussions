package progress

import (
	"path/filepath"
	"testing"

	"github.com/exileum/xenforo-to-gh-discussions/internal/xenforo"
)

// newTestTracker creates a new tracker for testing with a temp file
func newTestTracker(t *testing.T) (*Tracker, string) {
	t.Helper()
	tempDir := t.TempDir()
	progressFile := filepath.Join(tempDir, "test_progress.json")

	tracker, err := NewTracker(progressFile, false)
	if err != nil {
		t.Fatalf("Failed to create tracker: %v", err)
	}

	return tracker, progressFile
}

func TestProgressTracker(t *testing.T) {
	tracker, progressFile := newTestTracker(t)

	// Test initial state
	prog := tracker.GetProgress()
	if len(prog.CompletedThreads) != 0 {
		t.Error("New tracker should have no completed threads")
	}

	// Test marking completed
	err := tracker.MarkCompleted(123)
	if err != nil {
		t.Errorf("Failed to mark thread as completed: %v", err)
	}

	prog = tracker.GetProgress()
	if len(prog.CompletedThreads) != 1 || prog.CompletedThreads[0] != 123 {
		t.Error("Thread 123 should be marked as completed")
	}

	// Test marking failed
	err = tracker.MarkFailed(456)
	if err != nil {
		t.Errorf("Failed to mark thread as failed: %v", err)
	}

	prog = tracker.GetProgress()
	if len(prog.FailedThreads) != 1 || prog.FailedThreads[0] != 456 {
		t.Error("Thread 456 should be marked as failed")
	}

	// Test persistence by creating a new tracker
	tracker2, err := NewTracker(progressFile, false)
	if err != nil {
		t.Fatalf("Failed to create second tracker: %v", err)
	}

	prog2 := tracker2.GetProgress()
	if len(prog2.CompletedThreads) != 1 || prog2.CompletedThreads[0] != 123 {
		t.Error("Progress should persist across tracker instances")
	}
}

func TestFilterCompletedThreads(t *testing.T) {
	tracker, _ := newTestTracker(t)

	// Mark some threads as completed
	if err := tracker.MarkCompleted(1); err != nil {
		t.Fatalf("Failed to mark thread 1 as completed: %v", err)
	}
	if err := tracker.MarkCompleted(3); err != nil {
		t.Fatalf("Failed to mark thread 3 as completed: %v", err)
	}

	threads := []xenforo.Thread{
		{ThreadID: 1, Title: "Thread 1"},
		{ThreadID: 2, Title: "Thread 2"},
		{ThreadID: 3, Title: "Thread 3"},
		{ThreadID: 4, Title: "Thread 4"},
	}

	filtered := tracker.FilterCompletedThreads(threads)

	if len(filtered) != 2 {
		t.Errorf("Expected 2 threads after filtering, got %d", len(filtered))
	}

	// Check that only threads 2 and 4 remain
	expectedIDs := map[int]bool{2: true, 4: true}
	for _, thread := range filtered {
		if !expectedIDs[thread.ThreadID] {
			t.Errorf("Unexpected thread ID %d in filtered results", thread.ThreadID)
		}
	}
}

func TestMarkCompletedDuplicatePrevention(t *testing.T) {
	tracker, _ := newTestTracker(t)

	// Mark thread 1 as completed multiple times
	if err := tracker.MarkCompleted(1); err != nil {
		t.Fatalf("Failed to mark thread 1 as completed: %v", err)
	}
	if err := tracker.MarkCompleted(1); err != nil {
		t.Fatalf("Failed to mark thread 1 as completed (duplicate): %v", err)
	}
	if err := tracker.MarkCompleted(1); err != nil {
		t.Fatalf("Failed to mark thread 1 as completed (duplicate 2): %v", err)
	}

	// Check that thread 1 appears only once in CompletedThreads
	progress := tracker.GetProgress()
	count := 0
	for _, id := range progress.CompletedThreads {
		if id == 1 {
			count++
		}
	}
	if count != 1 {
		t.Errorf("Expected thread 1 to appear once in CompletedThreads, but found %d occurrences", count)
	}
}

func TestMarkFailedDuplicatePrevention(t *testing.T) {
	tracker, _ := newTestTracker(t)

	// Mark thread 2 as failed multiple times
	if err := tracker.MarkFailed(2); err != nil {
		t.Fatalf("Failed to mark thread 2 as failed: %v", err)
	}
	if err := tracker.MarkFailed(2); err != nil {
		t.Fatalf("Failed to mark thread 2 as failed (duplicate): %v", err)
	}
	if err := tracker.MarkFailed(2); err != nil {
		t.Fatalf("Failed to mark thread 2 as failed (duplicate 2): %v", err)
	}

	// Check that thread 2 appears only once in FailedThreads
	progress := tracker.GetProgress()
	count := 0
	for _, id := range progress.FailedThreads {
		if id == 2 {
			count++
		}
	}
	if count != 1 {
		t.Errorf("Expected thread 2 to appear once in FailedThreads, but found %d occurrences", count)
	}
}
