// Package config provides configuration management for the XenForo to GitHub
// Discussions migration tool. It supports both interactive prompts and
// environment variable configuration, with comprehensive validation.
package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration settings for the migration tool.
// It aggregates XenForo source settings, GitHub destination settings,
// migration behavior controls, and filesystem configuration.
type Config struct {
	XenForo    XenForoConfig
	GitHub     GitHubConfig
	Migration  MigrationConfig
	Filesystem FilesystemConfig
}

// XenForoConfig contains XenForo forum API connection settings.
// All fields are required for successful forum data retrieval.
type XenForoConfig struct {
	APIURL  string // Base URL for XenForo API (e.g., "https://forum.example.com/api")
	APIKey  string // XenForo API key for authentication
	APIUser string // XenForo user ID for API requests
	NodeID  int    // Forum node/category ID to migrate
}

// GitHubConfig contains GitHub API connection and rate limiting settings.
// Supports both legacy multi-category mapping and single-category migration.
type GitHubConfig struct {
	Token                string         // GitHub personal access token
	Repository           string         // Target repository in "owner/repo" format
	Categories           map[int]string // Kept for backward compatibility
	XenForoNodeID        int            // Single source category
	GitHubCategoryID     string         // Single target category
	RateLimitDelay       time.Duration  // Delay between API calls
	MaxRetries           int            // Maximum retries for rate limited requests
	RetryBackoffMultiple int            // Multiplier for exponential backoff (seconds)
}

// MigrationConfig controls migration behavior and retry logic.
// Provides options for dry-run testing and verbose output.
type MigrationConfig struct {
	MaxRetries       int  // Maximum retries for failed operations
	DryRun           bool // Enable dry-run mode (no actual changes)
	Verbose          bool // Enable verbose logging
	ResumeFrom       int
	ProgressFile     string
	UserMapping      map[int]int
	OperationTimeout time.Duration // Timeout for individual operations (0 = no timeout)
	RequestTimeout   time.Duration // Timeout for HTTP requests
}

// FilesystemConfig contains settings for file attachment handling.
// Controls where attachments are stored and download rate limiting.
type FilesystemConfig struct {
	AttachmentsDir           string        // Directory for storing downloaded attachments
	AttachmentRateLimitDelay time.Duration // Delay between attachment downloads
}

// New creates a new Config with default values populated from environment variables.
// Falls back to placeholder values if environment variables are not set.
func New() *Config {
	return &Config{
		XenForo: XenForoConfig{
			APIURL:  getEnvOrDefault("XENFORO_API_URL", "https://your-forum.com/api"),
			APIKey:  getEnvOrDefault("XENFORO_API_KEY", "your_xenforo_api_key"),
			APIUser: getEnvOrDefault("XENFORO_API_USER", "1"),
			NodeID:  getEnvIntOrDefault("XENFORO_NODE_ID", 1),
		},
		GitHub: GitHubConfig{
			Token:                getEnvOrDefault("GITHUB_TOKEN", "your_github_token"),
			Repository:           getEnvOrDefault("GITHUB_REPO", "your_username/your_repo"),
			Categories:           make(map[int]string),
			XenForoNodeID:        getEnvIntOrDefault("XENFORO_NODE_ID", 1),
			GitHubCategoryID:     getEnvOrDefault("GITHUB_CATEGORY_ID", "DIC_kwDOxxxxxxxx"),
			RateLimitDelay:       getEnvDurationOrDefault("GITHUB_RATE_LIMIT_DELAY", 1*time.Second),
			MaxRetries:           getEnvIntOrDefault("GITHUB_MAX_RETRIES", 5),
			RetryBackoffMultiple: getEnvIntOrDefault("GITHUB_RETRY_BACKOFF_MULTIPLE", 2),
		},
		Migration: MigrationConfig{
			MaxRetries:       getEnvIntOrDefault("MAX_RETRIES", 3),
			ProgressFile:     getEnvOrDefault("PROGRESS_FILE", "migration_progress.json"),
			UserMapping:      make(map[int]int),
			OperationTimeout: getEnvDurationOrDefault("OPERATION_TIMEOUT", 5*time.Minute),
			RequestTimeout:   getEnvDurationOrDefault("REQUEST_TIMEOUT", 30*time.Second),
		},
		Filesystem: FilesystemConfig{
			AttachmentsDir:           getEnvOrDefault("ATTACHMENTS_DIR", "./attachments"),
			AttachmentRateLimitDelay: getEnvDurationOrDefault("ATTACHMENT_RATE_LIMIT_DELAY", 500*time.Millisecond),
		},
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
