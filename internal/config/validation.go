package config

import (
	"fmt"
	"net/url"
	"strings"
)

// CategoryValidator defines the interface for validating GitHub category configurations
type CategoryValidator interface {
	ValidateSingleCategory(nodeID int, categoryID string) error
	ValidateMultiCategory(categories map[int]string) error
	ValidateNoConfiguration() error
}

// ValidateCategoryConfiguration handles the common branching logic for category validation
func ValidateCategoryConfiguration(config *Config, validator CategoryValidator) error {
	if config.GitHub.XenForoNodeID > 0 && config.GitHub.GitHubCategoryID != "" {
		return validator.ValidateSingleCategory(config.GitHub.XenForoNodeID, config.GitHub.GitHubCategoryID)
	} else if len(config.GitHub.Categories) > 0 {
		return validator.ValidateMultiCategory(config.GitHub.Categories)
	} else {
		return validator.ValidateNoConfiguration()
	}
}

// basicConfigValidator implements CategoryValidator for basic config validation
type basicConfigValidator struct{}

func (v *basicConfigValidator) ValidateSingleCategory(nodeID int, categoryID string) error {
	if categoryID == "DIC_kwDOxxxxxxxx" {
		return fmt.Errorf("GitHub category ID must be configured (not the default placeholder)")
	}
	return nil
}

func (v *basicConfigValidator) ValidateMultiCategory(categories map[int]string) error {
	for nodeID, categoryID := range categories {
		if nodeID <= 0 {
			return fmt.Errorf("node ID must be positive: %d", nodeID)
		}
		if categoryID == "" || categoryID == "DIC_kwDOxxxxxxxx" {
			return fmt.Errorf("category ID must be configured for node %d", nodeID)
		}
	}
	return nil
}

func (v *basicConfigValidator) ValidateNoConfiguration() error {
	return fmt.Errorf("either single-category configuration (XenForoNodeID + GitHubCategoryID) or legacy category mappings must be configured")
}

func (c *Config) Validate() error {
	if err := c.validateXenForo(); err != nil {
		return fmt.Errorf("XenForo config validation failed: %w", err)
	}

	if err := c.validateGitHub(); err != nil {
		return fmt.Errorf("GitHub config validation failed: %w", err)
	}

	if err := c.validateMigration(); err != nil {
		return fmt.Errorf("migration config validation failed: %w", err)
	}

	return nil
}

func (c *Config) validateXenForo() error {
	if c.XenForo.APIURL == "" || c.XenForo.APIURL == "https://your-forum.com/api" {
		return fmt.Errorf("XenForo API URL must be configured")
	}

	if _, err := url.Parse(c.XenForo.APIURL); err != nil {
		return fmt.Errorf("invalid XenForo API URL: %w", err)
	}

	if c.XenForo.APIKey == "" || c.XenForo.APIKey == "your_xenforo_api_key" {
		return fmt.Errorf("XenForo API key must be configured")
	}

	if c.XenForo.APIUser == "" {
		return fmt.Errorf("XenForo API user must be configured")
	}

	if c.XenForo.NodeID <= 0 {
		return fmt.Errorf("XenForo node ID must be positive")
	}

	return nil
}

func (c *Config) validateGitHub() error {
	if err := c.validateGitHubAuth(); err != nil {
		return err
	}

	if err := c.validateGitHubRepository(); err != nil {
		return err
	}

	if err := c.validateGitHubRateLimiting(); err != nil {
		return err
	}

	return c.validateGitHubCategories()
}

func (c *Config) validateGitHubAuth() error {
	if c.GitHub.Token == "" || c.GitHub.Token == "your_github_token" {
		return fmt.Errorf("GitHub token must be configured")
	}
	return nil
}

func (c *Config) validateGitHubRepository() error {
	if c.GitHub.Repository == "" || c.GitHub.Repository == "your_username/your_repo" {
		return fmt.Errorf("GitHub repository must be configured")
	}

	parts := strings.Split(c.GitHub.Repository, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("GitHub repository must be in format 'owner/repo'")
	}
	return nil
}

func (c *Config) validateGitHubRateLimiting() error {
	if c.GitHub.RateLimitDelay < 0 {
		return fmt.Errorf("GitHub rate limit delay cannot be negative")
	}

	if c.GitHub.MaxRetries < 0 {
		return fmt.Errorf("GitHub max retries cannot be negative")
	}

	if c.GitHub.RetryBackoffMultiple <= 0 {
		return fmt.Errorf("GitHub retry backoff multiple must be positive")
	}
	return nil
}

func (c *Config) validateGitHubCategories() error {
	validator := &basicConfigValidator{}
	return ValidateCategoryConfiguration(c, validator)
}

func (c *Config) validateMigration() error {
	if c.Migration.MaxRetries <= 0 {
		return fmt.Errorf("max retries must be positive")
	}

	if c.Migration.ProgressFile == "" {
		return fmt.Errorf("progress file path must be configured")
	}

	return nil
}
