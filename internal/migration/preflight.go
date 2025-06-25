package migration

import (
	"fmt"
	"log"
	"os"

	"github.com/exileum/xenforo-to-gh-discussions/internal/config"
	"github.com/exileum/xenforo-to-gh-discussions/internal/github"
	"github.com/exileum/xenforo-to-gh-discussions/internal/xenforo"
)

type PreflightChecker struct {
	config        *config.Config
	xenforoClient *xenforo.Client
	githubClient  *github.Client
}

func NewPreflightChecker(cfg *config.Config, xenforoClient *xenforo.Client, githubClient *github.Client) *PreflightChecker {
	return &PreflightChecker{
		config:        cfg,
		xenforoClient: xenforoClient,
		githubClient:  githubClient,
	}
}

func (p *PreflightChecker) RunChecks() error {
	log.Println("Running pre-flight checks...")

	if p.config.Migration.DryRun {
		log.Println("  Running in DRY-RUN mode - no actual changes will be made")
		return nil
	}

	if err := p.checkXenForoAPI(); err != nil {
		return err
	}

	if err := p.checkGitHubAPI(); err != nil {
		return err
	}

	if err := p.checkFileSystem(); err != nil {
		return err
	}

	log.Println("✓ All pre-flight checks passed")
	return nil
}

func (p *PreflightChecker) checkXenForoAPI() error {
	if err := p.xenforoClient.TestConnection(); err != nil {
		return fmt.Errorf("XenForo API check failed: %w", err)
	}
	log.Println("  ✓ XenForo API access verified")
	return nil
}

func (p *PreflightChecker) checkGitHubAPI() error {
	if p.config.Migration.DryRun || p.githubClient == nil {
		return nil
	}

	info, err := p.githubClient.GetRepositoryInfo(p.config.GitHub.Repository)
	if err != nil {
		return fmt.Errorf("GitHub API check failed: %w", err)
	}

	if !info.DiscussionsEnabled {
		return fmt.Errorf("GitHub Discussions is not enabled for repository %s", p.config.GitHub.Repository)
	}

	// Validate category mappings
	validCategories := make(map[string]bool)
	for _, cat := range info.DiscussionCategories {
		validCategories[cat.ID] = true
	}

	for nodeID, categoryID := range p.config.GitHub.Categories {
		if !validCategories[categoryID] {
			return fmt.Errorf("invalid category ID '%s' for node %d", categoryID, nodeID)
		}
	}

	log.Println("  ✓ GitHub API access verified")
	log.Println("  ✓ GitHub Discussions is enabled")
	log.Println("  ✓ All category mappings are valid")

	return nil
}

func (p *PreflightChecker) checkFileSystem() error {
	if err := os.MkdirAll(p.config.Filesystem.AttachmentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create attachments directory: %w", err)
	}
	log.Println("  ✓ Attachments directory ready")
	return nil
}
