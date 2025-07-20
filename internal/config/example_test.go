package config

import (
	"fmt"
	"time"
)

func ExampleNew() {
	// Create a new configuration with default values
	cfg := New()

	// Display some default values
	fmt.Printf("Default XenForo URL: %s\n", cfg.XenForo.APIURL)
	fmt.Printf("Default GitHub rate limit delay: %v\n", cfg.GitHub.RateLimitDelay)
	fmt.Printf("Default max retries: %d\n", cfg.Migration.MaxRetries)
	// Output: Default XenForo URL: https://your-forum.com/api
	// Default GitHub rate limit delay: 1s
	// Default max retries: 3
}

func ExampleConfig_Validate() {
	cfg := &Config{
		XenForo: XenForoConfig{
			APIURL:  "https://forum.example.com/api",
			APIKey:  "valid_key",
			APIUser: "1",
			NodeID:  1,
		},
		GitHub: GitHubConfig{
			Token:                "valid_token",
			Repository:           "owner/repo",
			XenForoNodeID:        1,
			GitHubCategoryID:     "DIC_kwDOtest123",
			RateLimitDelay:       1 * time.Second,
			MaxRetries:           3,
			RetryBackoffMultiple: 2,
		},
		Migration: MigrationConfig{
			MaxRetries:   3,
			DryRun:       false,
			Verbose:      true,
			ProgressFile: "./progress.json",
		},
		Filesystem: FilesystemConfig{
			AttachmentsDir:           "./attachments",
			AttachmentRateLimitDelay: 500 * time.Millisecond,
		},
	}

	err := cfg.Validate()
	if err != nil {
		fmt.Printf("Validation failed: %v\n", err)
	} else {
		fmt.Println("Configuration is valid")
	}
	// Output: Configuration is valid
}

func ExampleConfig_Validate_invalid() {
	cfg := &Config{
		XenForo: XenForoConfig{
			APIURL:  "https://your-forum.com/api", // Invalid placeholder URL
			APIKey:  "test_key",
			APIUser: "1",
			NodeID:  1,
		},
		GitHub: GitHubConfig{
			Token:                "test_token",
			Repository:           "owner/repo",
			XenForoNodeID:        1,
			GitHubCategoryID:     "DIC_kwDOtest123",
			RateLimitDelay:       1 * time.Second,
			MaxRetries:           3,
			RetryBackoffMultiple: 2,
		},
	}

	err := cfg.Validate()
	if err != nil {
		fmt.Printf("Validation failed: %v\n", err)
	} else {
		fmt.Println("Configuration is valid")
	}
	// Output: Validation failed: XenForo config validation failed: XenForo API URL must be configured
}

func ExampleXenForoConfig() {
	// Example XenForo configuration
	xenforoConfig := XenForoConfig{
		APIURL:  "https://myform.com/api",
		APIKey:  "your_api_key_here",
		APIUser: "1",
		NodeID:  42,
	}

	fmt.Printf("Forum API URL: %s\n", xenforoConfig.APIURL)
	fmt.Printf("Target Node ID: %d\n", xenforoConfig.NodeID)
	// Output: Forum API URL: https://myform.com/api
	// Target Node ID: 42
}

func ExampleGitHubConfig() {
	// Example GitHub configuration with rate limiting
	githubConfig := GitHubConfig{
		Token:                "ghp_your_token_here",
		Repository:           "myorg/my-discussions-repo",
		XenForoNodeID:        42,
		GitHubCategoryID:     "DIC_kwDOABCDEF1234",
		RateLimitDelay:       2 * time.Second,
		MaxRetries:           5,
		RetryBackoffMultiple: 2,
	}

	fmt.Printf("Repository: %s\n", githubConfig.Repository)
	fmt.Printf("Rate limit delay: %v\n", githubConfig.RateLimitDelay)
	fmt.Printf("Max retries: %d\n", githubConfig.MaxRetries)
	// Output: Repository: myorg/my-discussions-repo
	// Rate limit delay: 2s
	// Max retries: 5
}

func ExampleMigrationConfig() {
	// Example migration configuration
	migrationConfig := MigrationConfig{
		MaxRetries:   3,
		DryRun:       true, // Safe testing mode
		Verbose:      true, // Detailed logging
		ResumeFrom:   0,    // Start from beginning
		ProgressFile: "./migration_progress.json",
		UserMapping:  map[int]int{1: 101, 2: 102}, // Map old user IDs to new ones
	}

	fmt.Printf("Dry run mode: %t\n", migrationConfig.DryRun)
	fmt.Printf("Verbose logging: %t\n", migrationConfig.Verbose)
	fmt.Printf("Progress file: %s\n", migrationConfig.ProgressFile)
	// Output: Dry run mode: true
	// Verbose logging: true
	// Progress file: ./migration_progress.json
}
