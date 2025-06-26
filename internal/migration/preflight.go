package migration

import (
	"fmt"
	"log"
	"os"

	"github.com/exileum/xenforo-to-gh-discussions/internal/config"
	"github.com/exileum/xenforo-to-gh-discussions/internal/github"
	"github.com/exileum/xenforo-to-gh-discussions/internal/xenforo"
)

// runtimeCategoryValidator implements CategoryValidator for runtime GitHub API validation
type runtimeCategoryValidator struct {
	validCategories map[string]bool
}

func (v *runtimeCategoryValidator) ValidateSingleCategory(nodeID int, categoryID string) error {
	if !v.validCategories[categoryID] {
		return fmt.Errorf("invalid GitHub category ID '%s'", categoryID)
	}
	log.Printf("  ✓ Single category mapping validated: node %d -> %s", nodeID, categoryID)
	return nil
}

func (v *runtimeCategoryValidator) ValidateMultiCategory(categories map[int]string) error {
	for nodeID, categoryID := range categories {
		if !v.validCategories[categoryID] {
			return fmt.Errorf("invalid category ID '%s' for node %d", categoryID, nodeID)
		}
	}
	log.Println("  ✓ All legacy category mappings are valid")
	return nil
}

func (v *runtimeCategoryValidator) ValidateNoConfiguration() error {
	// For runtime validation, no configuration is allowed (handled by preflight logic)
	return nil
}

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
	if p.githubClient == nil {
		return nil
	}

	info, err := p.githubClient.GetRepositoryInfo(p.config.GitHub.Repository)
	if err != nil {
		return fmt.Errorf("GitHub API check failed: %w", err)
	}

	if !info.HasDiscussionsEnabled {
		return fmt.Errorf("GitHub Discussions is not enabled for repository %s", p.config.GitHub.Repository)
	}

	// Validate category configuration
	validCategories := make(map[string]bool)
	for _, cat := range info.DiscussionCategories {
		validCategories[cat.ID] = true
	}

	// Validate category configuration using shared logic
	validator := &runtimeCategoryValidator{validCategories: validCategories}
	if err := config.ValidateCategoryConfiguration(p.config, validator); err != nil {
		return err
	}

	log.Println("  ✓ GitHub API access verified")
	log.Println("  ✓ GitHub Discussions is enabled")

	return nil
}

func (p *PreflightChecker) checkFileSystem() error {
	if p.config.Migration.DryRun {
		// In dry-run mode, just check if the path is valid without creating the directory
		if p.config.Filesystem.AttachmentsDir == "" {
			return fmt.Errorf("attachments directory path is empty")
		}
		log.Println("  ✓ Attachments directory path validated (dry-run)")
		return nil
	}

	if err := os.MkdirAll(p.config.Filesystem.AttachmentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create attachments directory: %w", err)
	}
	log.Println("  ✓ Attachments directory ready")
	return nil
}
