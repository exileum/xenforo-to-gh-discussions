package config

import (
	"fmt"
	"net/url"
	"strings"
)

func (c *Config) Validate() error {
	if err := c.validateXenForo(); err != nil {
		return fmt.Errorf("XenForo config validation failed: %w", err)
	}

	if err := c.validateGitHub(); err != nil {
		return fmt.Errorf("GitHub config validation failed: %w", err)
	}

	if err := c.validateMigration(); err != nil {
		return fmt.Errorf("Migration config validation failed: %w", err)
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
	if c.GitHub.Token == "" || c.GitHub.Token == "your_github_token" {
		return fmt.Errorf("GitHub token must be configured")
	}

	if c.GitHub.Repository == "" || c.GitHub.Repository == "your_username/your_repo" {
		return fmt.Errorf("GitHub repository must be configured")
	}

	parts := strings.Split(c.GitHub.Repository, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("GitHub repository must be in format 'owner/repo'")
	}

	if len(c.GitHub.Categories) == 0 {
		return fmt.Errorf("at least one GitHub category mapping must be configured")
	}

	for nodeID, categoryID := range c.GitHub.Categories {
		if nodeID <= 0 {
			return fmt.Errorf("node ID must be positive: %d", nodeID)
		}
		if categoryID == "" || categoryID == "DIC_kwDOxxxxxxxx" {
			return fmt.Errorf("category ID must be configured for node %d", nodeID)
		}
	}

	return nil
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
