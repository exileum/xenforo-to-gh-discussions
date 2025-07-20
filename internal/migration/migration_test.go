package migration

import (
	"errors"
	"testing"

	"github.com/exileum/xenforo-to-gh-discussions/internal/xenforo"
)

// Mock tracker that can simulate errors
type mockProgressTracker struct {
	markCompletedError error
	markFailedError    error
	completedThreads   []int
	failedThreads      []int
}

func (m *mockProgressTracker) MarkCompleted(threadID int) error {
	if m.markCompletedError != nil {
		return m.markCompletedError
	}
	m.completedThreads = append(m.completedThreads, threadID)
	return nil
}

func (m *mockProgressTracker) MarkFailed(threadID int) error {
	if m.markFailedError != nil {
		return m.markFailedError
	}
	m.failedThreads = append(m.failedThreads, threadID)
	return nil
}

func (m *mockProgressTracker) FilterCompletedThreads(threads []xenforo.Thread) []xenforo.Thread {
	return threads // Return all threads for simplicity
}

func (m *mockProgressTracker) PrintSummary() {
	// Do nothing for tests
}

// Mock migrator to test error handling
type testMigrator struct {
	tracker          *mockProgressTracker
	processError     error
	processedThreads []int
}

func (tm *testMigrator) processThread(thread xenforo.Thread) error {
	if tm.processError != nil {
		return tm.processError
	}
	tm.processedThreads = append(tm.processedThreads, thread.ThreadID)
	return nil
}

func (tm *testMigrator) runThreads(t *testing.T, threads []xenforo.Thread) error {
	// Simulate the main processing loop from runner.go
	for _, thread := range threads {
		if err := tm.processThread(thread); err != nil {
			if markErr := tm.tracker.MarkFailed(thread.ThreadID); markErr != nil {
				t.Logf("Failed to mark thread %d as failed: %v", thread.ThreadID, markErr)
			}
			continue
		}

		if err := tm.tracker.MarkCompleted(thread.ThreadID); err != nil {
			t.Logf("Failed to mark thread %d as completed: %v", thread.ThreadID, err)
		}
	}
	return nil
}

func TestProgressTrackingErrorHandling(t *testing.T) {
	tests := []struct {
		name               string
		markCompletedError error
		markFailedError    error
		processError       error
		expectedCompleted  int
		expectedFailed     int
		expectedProcessed  int
	}{
		{
			name:              "Success case - no errors",
			expectedCompleted: 2,
			expectedFailed:    0,
			expectedProcessed: 2,
		},
		{
			name:               "MarkCompleted error - should continue processing",
			markCompletedError: errors.New("failed to save progress"),
			expectedCompleted:  0, // Not marked due to error
			expectedFailed:     0,
			expectedProcessed:  2, // Should still process threads
		},
		{
			name:              "Process error with MarkFailed error",
			processError:      errors.New("thread processing failed"),
			markFailedError:   errors.New("failed to mark as failed"),
			expectedCompleted: 0,
			expectedFailed:    0, // Not marked due to error
			expectedProcessed: 0, // No threads processed due to process error
		},
		{
			name:              "Process error with successful MarkFailed",
			processError:      errors.New("thread processing failed"),
			expectedCompleted: 0,
			expectedFailed:    2, // Should be marked as failed
			expectedProcessed: 0, // No threads processed due to process error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock tracker with specified errors
			tracker := &mockProgressTracker{
				markCompletedError: tt.markCompletedError,
				markFailedError:    tt.markFailedError,
			}

			// Create test migrator
			migrator := &testMigrator{
				tracker:      tracker,
				processError: tt.processError,
			}

			// Test threads
			threads := []xenforo.Thread{
				{ThreadID: 1, Title: "Test Thread 1"},
				{ThreadID: 2, Title: "Test Thread 2"},
			}

			// Run the migration
			err := migrator.runThreads(t, threads)
			if err != nil {
				t.Errorf("runThreads should not return error: %v", err)
			}

			// Verify results
			if len(tracker.completedThreads) != tt.expectedCompleted {
				t.Errorf("Expected %d completed threads, got %d", tt.expectedCompleted, len(tracker.completedThreads))
			}

			if len(tracker.failedThreads) != tt.expectedFailed {
				t.Errorf("Expected %d failed threads, got %d", tt.expectedFailed, len(tracker.failedThreads))
			}

			if len(migrator.processedThreads) != tt.expectedProcessed {
				t.Errorf("Expected %d processed threads, got %d", tt.expectedProcessed, len(migrator.processedThreads))
			}
		})
	}
}

func TestProgressTrackingErrorMessages(t *testing.T) {
	// This test verifies that error handling logic works as expected
	// In practice, errors would be logged, but we can't easily test log output

	tracker := &mockProgressTracker{
		markCompletedError: errors.New("disk full"),
	}

	migrator := &testMigrator{
		tracker: tracker,
	}

	threads := []xenforo.Thread{
		{ThreadID: 1, Title: "Test Thread"},
	}

	// This should not panic or fail despite the tracking error
	err := migrator.runThreads(t, threads)
	if err != nil {
		t.Errorf("Should handle tracking errors gracefully: %v", err)
	}

	// The thread should still be processed
	if len(migrator.processedThreads) != 1 {
		t.Error("Thread should still be processed despite tracking error")
	}

	// But not marked as completed due to the error
	if len(tracker.completedThreads) != 0 {
		t.Error("Thread should not be marked as completed due to tracking error")
	}
}

// Mock downloader that can simulate download errors
type mockDownloader struct {
	downloadError error
	downloadCalls [][]xenforo.Attachment
}

func (m *mockDownloader) DownloadAttachments(attachments []xenforo.Attachment) error {
	m.downloadCalls = append(m.downloadCalls, attachments)
	return m.downloadError
}

func (m *mockDownloader) ReplaceAttachmentLinks(message string, attachments []xenforo.Attachment) string {
	return message // Return unchanged for simplicity
}

// Test migrator that includes downloader
type testMigratorWithDownloader struct {
	tracker      *mockProgressTracker
	downloader   *mockDownloader
	processError error
	attachments  []xenforo.Attachment
}

func (tm *testMigratorWithDownloader) processThreadWithDownloads(t *testing.T, thread xenforo.Thread) error {
	// Simulate the download logic from runner.go
	if len(tm.attachments) > 0 {
		if err := tm.downloader.DownloadAttachments(tm.attachments); err != nil {
			t.Logf("Failed to download attachments: %v", err)
		}
	}

	// Process error is independent of download error
	if tm.processError != nil {
		return tm.processError
	}

	return nil
}

func TestAttachmentDownloadErrorHandling(t *testing.T) {
	tests := []struct {
		name               string
		downloadError      error
		processError       error
		hasAttachments     bool
		expectDownloadCall bool
	}{
		{
			name:               "Success - no errors",
			hasAttachments:     true,
			expectDownloadCall: true,
		},
		{
			name:               "Download error - should continue processing",
			downloadError:      errors.New("network error downloading file"),
			hasAttachments:     true,
			expectDownloadCall: true,
		},
		{
			name:               "No attachments - no download call",
			hasAttachments:     false,
			expectDownloadCall: false,
		},
		{
			name:               "Process error with download error",
			downloadError:      errors.New("download failed"),
			processError:       errors.New("process failed"),
			hasAttachments:     true,
			expectDownloadCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := &mockProgressTracker{}
			downloader := &mockDownloader{
				downloadError: tt.downloadError,
			}

			migrator := &testMigratorWithDownloader{
				tracker:      tracker,
				downloader:   downloader,
				processError: tt.processError,
			}

			// Set up attachments
			if tt.hasAttachments {
				migrator.attachments = []xenforo.Attachment{
					{AttachmentID: 1, Filename: "test.jpg", DirectURL: "https://example.com/1"},
					{AttachmentID: 2, Filename: "doc.pdf", DirectURL: "https://example.com/2"},
				}
			}

			// Test thread
			thread := xenforo.Thread{ThreadID: 1, Title: "Test Thread"}

			// Process the thread
			err := migrator.processThreadWithDownloads(t, thread)

			// Verify download call behavior
			if tt.expectDownloadCall {
				if len(downloader.downloadCalls) != 1 {
					t.Errorf("Expected 1 download call, got %d", len(downloader.downloadCalls))
				} else {
					expectedAttachments := len(migrator.attachments)
					actualAttachments := len(downloader.downloadCalls[0])
					if actualAttachments != expectedAttachments {
						t.Errorf("Expected %d attachments in download call, got %d", expectedAttachments, actualAttachments)
					}
				}
			} else {
				if len(downloader.downloadCalls) != 0 {
					t.Errorf("Expected no download calls, got %d", len(downloader.downloadCalls))
				}
			}

			// Verify that download errors don't stop processing (unless there's a separate process error)
			if tt.processError != nil {
				if err == nil {
					t.Error("Expected process error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}
