package unit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/exileum/xenforo-to-gh-discussions/internal/progress"
	"github.com/exileum/xenforo-to-gh-discussions/internal/xenforo"
)

func TestProgressTracker(t *testing.T) {
	// Create temporary file for testing
	tempDir, err := os.MkdirTemp("", "progress-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	progressFile := filepath.Join(tempDir, "test_progress.json")

	// Create new tracker
	tracker, err := progress.NewTracker(progressFile, false)
	if err != nil {
		t.Fatalf("Failed to create tracker: %v", err)
	}

	// Test initial state
	prog := tracker.GetProgress()
	if len(prog.CompletedThreads) != 0 {
		t.Error("New tracker should have no completed threads")
	}

	// Test marking completed
	err = tracker.MarkCompleted(123)
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

	// Test persistence by creating new tracker
	tracker2, err := progress.NewTracker(progressFile, false)
	if err != nil {
		t.Fatalf("Failed to create second tracker: %v", err)
	}

	prog2 := tracker2.GetProgress()
	if len(prog2.CompletedThreads) != 1 || prog2.CompletedThreads[0] != 123 {
		t.Error("Progress should persist across tracker instances")
	}
}

func TestFilterCompletedThreads(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "filter-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	progressFile := filepath.Join(tempDir, "test_progress.json")

	tracker, err := progress.NewTracker(progressFile, false)
	if err != nil {
		t.Fatalf("Failed to create tracker: %v", err)
	}

	// Mark some threads as completed
	tracker.MarkCompleted(1)
	tracker.MarkCompleted(3)

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
