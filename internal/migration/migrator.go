// Package migration orchestrates the migration process from XenForo forums
// to GitHub Discussions. It coordinates XenForo data retrieval, content conversion,
// GitHub API interactions, progress tracking, and error recovery.
package migration

import (
	"context"
	"fmt"

	"github.com/exileum/xenforo-to-gh-discussions/internal/attachments"
	"github.com/exileum/xenforo-to-gh-discussions/internal/config"
	"github.com/exileum/xenforo-to-gh-discussions/internal/github"
	"github.com/exileum/xenforo-to-gh-discussions/internal/progress"
	"github.com/exileum/xenforo-to-gh-discussions/internal/xenforo"
)

// Migrator orchestrates the complete migration process from XenForo to GitHub Discussions.
// Coordinates all subsystems including data retrieval, content conversion, and progress tracking.
type Migrator struct {
	config *config.Config // Migration configuration
}

// NewMigrator creates a new migration orchestrator with the provided configuration.
// The configuration should be validated before creating the migrator.
func NewMigrator(cfg *config.Config) *Migrator {
	return &Migrator{
		config: cfg,
	}
}

// Run executes the complete migration process with the given context.
// Validates configuration, initializes all subsystems, and coordinates
// the migration of threads from XenForo to GitHub Discussions.
// Returns an error if any critical step fails.
func (m *Migrator) Run(ctx context.Context) error {
	// Validate configuration
	if err := m.config.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Initialize clients
	xenforoClient := xenforo.NewClient(
		m.config.XenForo.APIURL,
		m.config.XenForo.APIKey,
		m.config.XenForo.APIUser,
		m.config.Migration.MaxRetries,
	)

	var githubClient *github.Client
	if !m.config.Migration.DryRun {
		var err error
		githubClient, err = github.NewClient(
			m.config.GitHub.Token,
			m.config.GitHub.RateLimitDelay,
			m.config.GitHub.MaxRetries,
			m.config.GitHub.RetryBackoffMultiple,
		)
		if err != nil {
			return fmt.Errorf("failed to initialize GitHub client: %w", err)
		}
	}

	// Initialize progress tracker
	tracker, err := progress.NewTracker(m.config.Migration.ProgressFile, m.config.Migration.DryRun)
	if err != nil {
		return fmt.Errorf("failed to initialize progress tracker: %w", err)
	}

	// Set resume point if specified
	if m.config.Migration.ResumeFrom > 0 {
		tracker.SetResumeFrom(m.config.Migration.ResumeFrom)
	}

	// Initialize attachment downloader
	downloader := attachments.NewDownloader(
		m.config.Filesystem.AttachmentsDir,
		m.config.Migration.DryRun,
		xenforoClient,
		m.config.Filesystem.AttachmentRateLimitDelay,
	)

	// Run pre-flight checks
	checker := NewPreflightChecker(m.config, xenforoClient, githubClient)
	if err := checker.RunChecks(ctx); err != nil {
		return fmt.Errorf("pre-flight checks failed: %w", err)
	}

	// Run migration
	runner := NewRunner(m.config, xenforoClient, githubClient, tracker, downloader)
	return runner.RunMigration(ctx)
}
