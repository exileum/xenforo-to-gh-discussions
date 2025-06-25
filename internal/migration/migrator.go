package migration

import (
	"fmt"

	"github.com/exileum/xenforo-to-gh-discussions/internal/attachments"
	"github.com/exileum/xenforo-to-gh-discussions/internal/config"
	"github.com/exileum/xenforo-to-gh-discussions/internal/github"
	"github.com/exileum/xenforo-to-gh-discussions/internal/progress"
	"github.com/exileum/xenforo-to-gh-discussions/internal/xenforo"
)

type Migrator struct {
	config *config.Config
}

func NewMigrator(cfg *config.Config) *Migrator {
	return &Migrator{
		config: cfg,
	}
}

func (m *Migrator) Run() error {
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
		githubClient, err = github.NewClient(m.config.GitHub.Token)
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
	if err := checker.RunChecks(); err != nil {
		return fmt.Errorf("pre-flight checks failed: %w", err)
	}

	// Run migration
	runner := NewRunner(m.config, xenforoClient, githubClient, tracker, downloader)
	return runner.RunMigration()
}
