package migration

import (
	"errors"
	"fmt"
)

// Error classification for migration operations.

// MigrationError represents errors that occur during the migration process.
type MigrationError struct {
	Phase    string // The migration phase where the error occurred (e.g., "fetch", "convert", "upload")
	ThreadID int    // The thread ID being processed when the error occurred (0 if not applicable)
	Message  string // Human-readable error message
	Cause    error  // Underlying error cause
}

func (e *MigrationError) Error() string {
	if e.ThreadID > 0 {
		return fmt.Sprintf("migration error in phase '%s' for thread %d: %s", e.Phase, e.ThreadID, e.Message)
	}
	return fmt.Sprintf("migration error in phase '%s': %s", e.Phase, e.Message)
}

func (e *MigrationError) Unwrap() error {
	return e.Cause
}

// NewMigrationError creates a new migration error.
func NewMigrationError(phase, message string, cause error) *MigrationError {
	return &MigrationError{
		Phase:   phase,
		Message: message,
		Cause:   cause,
	}
}

// NewThreadMigrationError creates a new migration error for a specific thread.
func NewThreadMigrationError(phase string, threadID int, message string, cause error) *MigrationError {
	return &MigrationError{
		Phase:    phase,
		ThreadID: threadID,
		Message:  message,
		Cause:    cause,
	}
}

// RetryableError represents an error that can be retried.
type RetryableError struct {
	Operation  string // The operation that failed
	Attempt    int    // The attempt number
	MaxRetries int    // Maximum number of retries allowed
	Cause      error  // Underlying error cause
}

func (e *RetryableError) Error() string {
	return fmt.Sprintf("retryable error in operation '%s' (attempt %d/%d): %v", e.Operation, e.Attempt, e.MaxRetries, e.Cause)
}

func (e *RetryableError) Unwrap() error {
	return e.Cause
}

// NewRetryableError creates a new retryable error.
func NewRetryableError(operation string, attempt, maxRetries int, cause error) *RetryableError {
	return &RetryableError{
		Operation:  operation,
		Attempt:    attempt,
		MaxRetries: maxRetries,
		Cause:      cause,
	}
}

// ProgressError represents errors in progress tracking.
type ProgressError struct {
	ThreadID int    // The thread ID that couldn't be tracked
	Action   string // The progress action that failed (e.g., "mark_completed", "mark_failed")
	Cause    error  // Underlying error cause
}

func (e *ProgressError) Error() string {
	return fmt.Sprintf("progress tracking error for thread %d during action '%s': %v", e.ThreadID, e.Action, e.Cause)
}

func (e *ProgressError) Unwrap() error {
	return e.Cause
}

// NewProgressError creates a new progress tracking error.
func NewProgressError(threadID int, action string, cause error) *ProgressError {
	return &ProgressError{
		ThreadID: threadID,
		Action:   action,
		Cause:    cause,
	}
}

// Sentinel errors for common migration issues
var (
	// ErrThreadNotFound indicates a thread could not be found
	ErrThreadNotFound = errors.New("thread not found")

	// ErrPostNotFound indicates a post could not be found
	ErrPostNotFound = errors.New("post not found")

	// ErrAttachmentNotFound indicates an attachment could not be found
	ErrAttachmentNotFound = errors.New("attachment not found")

	// ErrDiscussionNotCreated indicates a GitHub discussion could not be created
	ErrDiscussionNotCreated = errors.New("GitHub discussion not created")

	// ErrCommentNotAdded indicates a comment could not be added to a discussion
	ErrCommentNotAdded = errors.New("comment not added to discussion")

	// ErrAttachmentDownloadFailed indicates an attachment download failed
	ErrAttachmentDownloadFailed = errors.New("attachment download failed")

	// ErrContentConversionFailed indicates BBCode to Markdown conversion failed
	ErrContentConversionFailed = errors.New("content conversion failed")

	// ErrProgressTrackingFailed indicates progress tracking failed
	ErrProgressTrackingFailed = errors.New("progress tracking failed")

	// ErrMigrationAborted indicates the migration was aborted
	ErrMigrationAborted = errors.New("migration aborted")

	// ErrDryRunMode indicates an operation was skipped due to dry run mode
	ErrDryRunMode = errors.New("operation skipped in dry run mode")

	// ErrContextCancelled indicates the migration was cancelled via context
	ErrContextCancelled = errors.New("migration cancelled")

	// ErrMaxRetriesExceeded indicates maximum retries were exceeded
	ErrMaxRetriesExceeded = errors.New("maximum retries exceeded")
)

// Error classification helper functions

// IsMigrationError checks if an error is a migration error.
func IsMigrationError(err error) bool {
	var migrationErr *MigrationError
	return errors.As(err, &migrationErr)
}

// IsRetryableError checks if an error is retryable.
func IsRetryableError(err error) bool {
	var retryableErr *RetryableError
	return errors.As(err, &retryableErr)
}

// IsProgressError checks if an error is a progress tracking error.
func IsProgressError(err error) bool {
	var progressErr *ProgressError
	return errors.As(err, &progressErr)
}

// GetMigrationPhase extracts the migration phase from a migration error.
func GetMigrationPhase(err error) string {
	var migrationErr *MigrationError
	if errors.As(err, &migrationErr) {
		return migrationErr.Phase
	}
	return ""
}

// GetThreadID extracts the thread ID from migration or progress errors.
func GetThreadID(err error) int {
	var migrationErr *MigrationError
	if errors.As(err, &migrationErr) {
		return migrationErr.ThreadID
	}

	var progressErr *ProgressError
	if errors.As(err, &progressErr) {
		return progressErr.ThreadID
	}

	return 0
}

// IsRetryable determines if an error should be retried based on its type and characteristics.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific retryable errors
	if errors.Is(err, ErrAttachmentDownloadFailed) ||
		errors.Is(err, ErrDiscussionNotCreated) ||
		errors.Is(err, ErrCommentNotAdded) {
		return true
	}

	// Check for retryable error wrapper
	return IsRetryableError(err)
}

// IsPermanent determines if an error is permanent and should not be retried.
func IsPermanent(err error) bool {
	if err == nil {
		return false
	}

	// Check for permanent errors
	if errors.Is(err, ErrThreadNotFound) ||
		errors.Is(err, ErrPostNotFound) ||
		errors.Is(err, ErrMigrationAborted) ||
		errors.Is(err, ErrContextCancelled) ||
		errors.Is(err, ErrMaxRetriesExceeded) {
		return true
	}

	return false
}
